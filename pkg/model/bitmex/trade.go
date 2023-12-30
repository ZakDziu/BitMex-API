package bitmex

import "time"

type TradeMessage struct {
	Table  string            `json:"table"`
	Action string            `json:"action"`
	Data   []TradeDataRecord `json:"data"`
}

type TradeDataRecord struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}
