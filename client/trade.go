package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func (c *TradingClient) Trade(until time.Time) {
	asset := c.chooseBestAsset()
	acc := c.getAccount()

	log.Printf("trade %v until %v\n", asset.Name, until)
	initialBalance := acc.Balance
	log.Printf("initial balance: %.3f\n", initialBalance)

	time.Sleep(250 * time.Millisecond)
	priceRec := c.newPriceReceiver()
	defer priceRec.Stop()

	priceFlow := make(chan float64)

	// fill priceFlow
	go func(assetName string) {
		for {
			select {
			case ma, ok := <- c.priceUpdate:
				if !ok {
					return
				}

				if ma.Name == assetName {
					priceFlow <- ma.Price
				}
			}
		}
	}(asset.Name)

	sell := false
	ownAsset := acc.GetAsset(asset.Name)

	if ownAsset != nil {
		sell = true
	}

	var lastBuyPrice float64
	priceHist := [3]float64{}
	var priceHistIdx int

	lastMinMaxPrintout := time.Now()

	var turnDown, turnUp bool

	for price := range priceFlow {
		// store next price sample
		priceHistIdx = (priceHistIdx + 1) % len(priceHist)
		priceHist[priceHistIdx] = price

		prev := priceHist[(priceHistIdx-1+len(priceHist))%len(priceHist)]
		prev2 := priceHist[(priceHistIdx-2+len(priceHist))%len(priceHist)]

		if prev == 0 || prev2 == 0 {
			// fill the price tracking array before jumping to conclusions
			continue
		}

		if prev2 < prev && price < prev {
			turnDown = true
			log.Printf("price turns down (%.3f,%.3f,%.3f)\n", prev2, prev, price)
		} else {
			turnDown = false
		}

		if prev2 > prev && price > prev {
			turnUp = true
			log.Printf("price turns up (%.3f,%.3f,%.3f)\n", prev2, prev, price)
		} else {
			turnUp = false
		}

		if sell && turnDown && (price > lastBuyPrice || lastBuyPrice == 0) {
			acc2, err := c.Sell(*ownAsset, acc)
			if err != nil {
				log.Fatalf("failed to sell asset %v: %v", ownAsset, err)
			}

			log.Printf("--> win %.3f\n", (price-lastBuyPrice)*ownAsset.Amount)

			sell = false
			ownAsset = nil
			acc = acc2

			if time.Now().After(until) {
				break
			}
		} else if !sell && turnUp {
			acc2, err := c.Buy(*asset, acc)
			if err != nil {
				log.Fatalf("failed to buy asset %v: %v", asset, err)
			}

			lastBuyPrice = price

			sell = true
			ownAsset = acc2.GetAsset(asset.Name)
			acc = acc2
		}
	}

	priceRec.Stop()
	log.Printf("final balance: %.3f  win: %3f\n", acc.Balance, acc.Balance-initialBalance)
}

func (c *TradingClient) chooseBestAsset() *MarketAsset {
	var max *MarketAsset
	for _, ma := range c.marketAssets {
		if max == nil || max.maxSeen < ma.maxSeen {
			max = ma
		}
	}

	return max
}

func (c *TradingClient) Buy(asset MarketAsset, account *Account)  (*Account, error) {
	amount := account.Balance / asset.Price * 0.95 // small margin allowing for price skew

	trans := &Transaction{
		Asset:  asset.Name,
		Amount: amount,
	}

	log.Printf("buy %.3f units of %v for %.3f\n", amount, asset.Name, account.Balance * 0.95)

	bytebuf, err := json.Marshal(trans)
	if err != nil {
		return nil, fmt.Errorf("Buy Marshal transaction failed: %v", err)
	}

	input := bytes.NewBuffer(bytebuf)

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%v:8002/buy", c.hostname), input)
	if err != nil {
		return nil, fmt.Errorf("Buy NewRequest failed: %v", err)
	}

	req.SetBasicAuth(c.login, c.password)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to POST /buy: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("buy query HTTP error %v: %v", resp.StatusCode, resp.Status)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Buy() read body failed: %v", err)
	}
	defer resp.Body.Close()

	acc := &Account{}
	err = json.Unmarshal(buf, acc)
	if err != nil {
		return nil, fmt.Errorf("Buy() failed to decode account: %v", err)
	}

	return acc, nil
}

func (c *TradingClient) Sell(asset UserAsset, account *Account) (*Account, error) {
	trans := &Transaction{
		Asset:  asset.Name,
		Amount: asset.Amount,
	}

	bytebuf, err := json.Marshal(trans)
	if err != nil {
		return nil, fmt.Errorf("Sell Marshal transaction failed: %v", err)
	}

	input := bytes.NewBuffer(bytebuf)

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%v:8002/sell", c.hostname), input)
	if err != nil {
		return nil, fmt.Errorf("Sell NewRequest failed: %v", err)
	}

	req.SetBasicAuth(c.login, c.password)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to POST /sell: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sell query HTTP error %v: %v", resp.StatusCode, resp.Status)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Sell() read body failed: %v", err)
	}
	defer resp.Body.Close()

	acc := &Account{}
	err = json.Unmarshal(buf, acc)
	if err != nil {
		return nil, fmt.Errorf("Sell() failed to decode account: %v", err)
	}

	log.Printf("sell %.3f units of %v for %.3f\n", trans.Amount, asset.Name, acc.Balance-account.Balance)

	return acc, nil
}

