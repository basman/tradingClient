package persist

import (
	"encoding/gob"
	"log"
	"os"
	"time"
)

const filename = "deal.bin"

type Deal struct {
	Time time.Time
	Asset string
	Amount float64
	Price float64
}

func (deal Deal) Store() {
	f, err := os.OpenFile(filename, os.O_CREATE | os.O_TRUNC | os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("failed to open/create deal file: %v\n", err)
		return
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	err = enc.Encode(&deal)
	if err != nil {
		log.Printf("failed to store deal: %v\n", err)
	}
}

func Load() *Deal {
	f, err := os.Open(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("failed to open deal file: %v\n", err)
		}
		return nil
	}
	defer f.Close()

	deal := Deal{}

	dec := gob.NewDecoder(f)
	err = dec.Decode(&deal)
	if err != nil {
		log.Printf("failed to load deal: %v\n", err)
		return nil
	}

	log.Printf("loaded deal %v", deal)
	return &deal
}
