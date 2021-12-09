package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type TradingClient struct {
	hostname string
	login string
	password string
	client *http.Client
	priceUpdate chan MarketAsset

	marketAssets map[string]*MarketAsset
}

func NewClient(hostname, login, pw string) *TradingClient {
	return &TradingClient{
		hostname: hostname,
		login:       login,
		password:    pw,
		client:      &http.Client{},
		priceUpdate: make(chan MarketAsset),
		marketAssets: make(map[string]*MarketAsset),
	}
}

func (c *TradingClient) Observe(until time.Time) {
	log.Printf("observing price structure until %v\n", until)
	priceRec := c.newPriceReceiver()

	lastStatusPrint := time.Now()

	outer:
	for {
		select {
		case maUpdate, ok := <- c.priceUpdate:
			if !ok {
				break outer
			}
			if ma, exists := c.marketAssets[maUpdate.Name]; exists {
				if maUpdate.Price > ma.maxSeen {
					ma.maxSeen = maUpdate.Price
				}
				if maUpdate.Price < ma.minSeen {
					ma.minSeen = maUpdate.Price
				}
			} else {
				maUpdate.maxSeen = maUpdate.Price
				maUpdate.minSeen = maUpdate.Price
				c.marketAssets[maUpdate.Name] = &maUpdate
			}
		}

		if time.Now().After(until) {
			fmt.Println("observation completed")
			break
		}

		if time.Now().Sub(lastStatusPrint) > 5 * time.Second {
			lastStatusPrint = time.Now()
			log.Println("Running observation")
			for _, ma := range c.marketAssets {
				fmt.Printf("%v\n", ma)
			}
			fmt.Println()
		}
	}

	priceRec.Stop()

	log.Println("Completed observation")
	for _, ma := range c.marketAssets {
		fmt.Printf("%v\n", ma)
	}
}

func (c *TradingClient) getAccount() *Account {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%v:8002/account", c.hostname), nil)
	if err != nil {
		log.Fatalf("getAccount NewRequest failed: %v", err)
	}

	req.SetBasicAuth(c.login, c.password)

	resp, err := c.client.Do(req)
	if err != nil {
		log.Fatalf("failed to query account: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("account query HTTP error %v: %v", resp.StatusCode, resp.Status)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("getAccount() read body failed: %v", err)
	}
	defer resp.Body.Close()

	acc := &Account{}
	err = json.Unmarshal(buf, acc)
	if err != nil {
		log.Fatalf("failed to decode account: %v", err)
	}

	return acc
}

