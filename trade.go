package main

import (
	"log"
	"os"
	"strconv"
	"time"
	client2 "tradingClient/client"
)

const hostname = "localhost"
const login = "test"
const password = "test"

func main() {
	tradeDuration := parseArgTradeDuration()

	client := client2.NewClient(hostname, login, password)

	client.Observe(time.Now().Add(20 * time.Second))

	client.Trade(time.Now().Add(tradeDuration))
}

func parseArgTradeDuration() time.Duration {
	tradeDuration := 1 * time.Minute
	if len(os.Args) > 1 {
		minutes, err := strconv.Atoi(os.Args[1])
		if err != nil {
			log.Fatalf("failed to parse trading duration (minutes) '%v': %v\n", os.Args[1], err)
		}
		tradeDuration = time.Duration(minutes * int(time.Minute))
	}
	return tradeDuration
}
