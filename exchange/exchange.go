package exchange

import (
	"context"
	"github.com/shopspring/decimal"
	"github.com/xyths/hs"
	"github.com/xyths/hs/logger"
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

type Config struct {
	hs.ExchangeConf

	Csv *CsvExchangeConfig `json:"csv,omitempty"`
}

func New(config Config) (ex Exchange) {
	switch config.Name {
	case "csv":
		if config.Csv == nil {
			logger.Sugar.Fatal("no csv config")
		}
		ex = NewCsvExchange(*config.Csv)
	default:
		ex = NewTurtle(config)
	}
	return
}
