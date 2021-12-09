package main

import (
	"time"
	client2 "tradingClient/client"
)

const hostname = "localhost"
const login = "test"
const password = "test"

func main() {
	client := client2.NewClient(hostname, login, password)

	client.Observe(time.Now().Add(30 * time.Second))

	client.Trade(time.Now().Add(1 * time.Minute))
}
