package types

import (
	"github.com/shopspring/decimal"
	"time"
)

type Quote struct {
	Symbol string `json:"symbol"`
	//Precision int64       `json:"-"`
	Date   []time.Time `json:"date"`
	Open   []float64   `json:"open"`
	High   []float64   `json:"high"`
	Low    []float64   `json:"low"`
	Close  []float64   `json:"close"`
	Volume []float64   `json:"volume"`
}
type Direction int

const (
	Buy  Direction = 1
	Sell           = -1
)

type Signal struct {
	Direction Direction
	Price     decimal.Decimal
	Amount    decimal.Decimal
}
