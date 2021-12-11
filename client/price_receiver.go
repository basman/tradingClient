package client

import (
	"encoding/base64"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

type priceReceiver struct {
	conn    *websocket.Conn
	updates chan MarketAsset
	stop    chan bool
}

func (c *TradingClient) newPriceReceiver() *priceReceiver {
	reqHeader := http.Header{
		"Authorization" :
		{"Basic " + base64.StdEncoding.EncodeToString([]byte(c.login+":"+c.password))},
	}
	conn, res, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%v:8002/rates/stream", c.hostname), reqHeader)
	if err != nil {
		log.Fatalf("newPriceReceiver: could not connect to websocket; HTTP status=%v '%v'; %v", res.StatusCode, res.Status, err)
	}

	if res.StatusCode != http.StatusSwitchingProtocols && res.StatusCode != http.StatusOK {
		log.Fatalf("newPriceReceiver: HTTP error %v %v", res.StatusCode, res.Status)
	}

	priceRec := &priceReceiver{
		conn:    conn,
		updates: c.priceUpdate,
		stop:    make(chan bool, 10),
	}

	go priceRec.feed()

	return priceRec
}

func (priceRec *priceReceiver) Stop() {
	// called by user
	priceRec.stop <- true
}

func (priceRec *priceReceiver) feed() {
	ch := make(chan MarketAsset)

	go func() {
		for {
			ma := MarketAsset{}
			priceRec.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			err := priceRec.conn.ReadJSON(&ma)
			if err != nil {
				log.Printf("read websocket stream failed: %v", err)
				break
			}
			priceRec.conn.SetReadDeadline(time.Time{})

			ch <- ma
		}
		priceRec.conn.Close()
	}()

	for {
		select {
		case <- priceRec.stop:
			priceRec.conn.Close() // FIXME leads to goroutine above to show an error, despite shutdown is intended
			return
		case ma := <- ch:
			priceRec.updates <- ma
		}
	}
}
