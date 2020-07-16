package strategy

import (
	"context"
	"github.com/xyths/Turtle-Trading/types"
)

type Strategy interface {
	OnTicker(quote types.Quote) (types.Signal, error)
	Start(ctx context.Context)
	//Stop(ctx context.Context)
}

type Config struct {
	Total      float64
	CandleType string`json:"candle"`
}

const (
	CandleType1Min = "1m"
	CandleType5Min = "5m"
	CandleType1H   = "1h"
	CandleType1D   = "1d"
)
