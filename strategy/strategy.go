package strategy

import (
	"context"
	"github.com/xyths/Turtle-Trading/exchange"
	"github.com/xyths/Turtle-Trading/types"
)

type Strategy interface {
	OnTicker(ticker exchange.Ticker) (*types.Signal, error)
	OnCandle(candle exchange.Candle) (*types.Signal, error)
	Start(ctx context.Context)
	//Stop(ctx context.Context)
}

type Config struct {
	Total      float64
	CandleType string `json:"candle"`
}

const (
	CandleType1Min  = "1m"
	CandleType5Min  = "5m"
	CandleType15Min = "15m"
	CandleType30Min = "30m"
	CandleType1H    = "1h"
	CandleType4H    = "4h"
	CandleType1D    = "1d"
)
