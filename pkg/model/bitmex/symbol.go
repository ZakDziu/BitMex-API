package bitmex

import "time"

type SymbolInfo struct {
	Symbol         string    `json:"symbol"`
	RootSymbol     string    `json:"rootSymbol"`
	State          string    `json:"state"`
	Typ            string    `json:"typ"`
	Listing        time.Time `json:"listing"`
	Front          time.Time `json:"front"`
	Expiry         time.Time `json:"expiry"`
	Settle         time.Time `json:"settle"`
	RelistInterval time.Time `json:"relistInterval"`
}
