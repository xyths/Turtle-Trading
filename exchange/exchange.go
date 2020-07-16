package exchange

import (
	"context"
	"github.com/shopspring/decimal"
	"github.com/xyths/hs"
)

type Ticker = hs.Ticker
type Candle = hs.Candle

type Exchange interface {
	MinAmount() decimal.Decimal
	MinTotal() decimal.Decimal
	PricePrecision() int
	AmountPrecision() int

	SetTickerChannel(tickerCh chan Ticker)
	SetCandleChannel(candleCh chan Candle)
	Start(ctx context.Context)
}
