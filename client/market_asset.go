package client

import (
	"fmt"
)

type MarketAsset struct {
	Name string
	Price float64 `json:",string"`

	maxSeen float64
	minSeen float64
}

func (ma MarketAsset) String() string {
	return fmt.Sprintf("%v %.3f (min %.3f, max %.3f)", ma.Name, ma.Price, ma.minSeen, ma.maxSeen)
}
