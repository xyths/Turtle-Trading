package exchange

import (
	"context"
	"github.com/shopspring/decimal"
	"github.com/xyths/hs"
	"github.com/xyths/hs/logger"
)

type Ticker = hs.Ticker
type Candle = hs.Candle

// exchange interface for turtle trade only. with only one symbol, only one fee currency, and almost no error
type Exchange interface {
	Symbol() string

	QuoteCurrency() string // BTC/USDT -> USDT
	BaseCurrency() string  // BTC/USDT -> BTC
	FeeCurrency() string

	MinAmount() decimal.Decimal
	MinTotal() decimal.Decimal
	PricePrecision() int32
	AmountPrecision() int32

	Balance() (cash, currency, fee decimal.Decimal)

	Price() (decimal.Decimal, error)
	FeePrice() (decimal.Decimal, error)

	Buy(price, amount decimal.Decimal, clientId string) (hs.Order, error)
	Sell(price, amount decimal.Decimal, clientId string) (hs.Order, error)

	SetTickerChannel(tickerCh chan Ticker)
	SetCandleChannel(candleCh chan Candle)
	Start(ctx context.Context)
}

type Config struct {
	hs.ExchangeConf
	FeeCurrency string `json:"feeCurrency"` // eg. HT
	ClientId    string `json:"clientId"`
	Period      string // candle type, 1m, 15m, 1h

	Csv *CsvExchangeConfig `json:"csv,omitempty"`
}

func New(config Config) (ex Exchange) {
	switch config.Name {
	case "csv":
		if config.Csv == nil {
			logger.Sugar.Fatal("no csv config")
		}
		ex = NewCsvExchange(*config.Csv, config.Symbols)
	default:
		ex = NewTurtle(config)
	}
	return
}
