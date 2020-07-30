package executor

import (
	"github.com/shopspring/decimal"
	"github.com/xyths/Turtle-Trading/exchange"
	"github.com/xyths/Turtle-Trading/types"
	"github.com/xyths/hs"
)

// A exchange agent for turtle use.
// It saves some useful variables, eg. the only symbol, base-currency, quote-currency, precisions.
type Executor interface {
	QuoteCurrency() string // BTC/USDT -> USDT
	BaseCurrency() string  // BTC/USDT -> BTC
	FeeCurrency() string
	Balance() (cash, currency, fee decimal.Decimal)
	MinAmount() decimal.Decimal
	MinTotal() decimal.Decimal
	PricePrecision() int
	AmountPrecision() int
	Buy(price, amount decimal.Decimal, clientId string) (hs.Order, error)
	Sell(price, amount decimal.Decimal, clientId string) (hs.Order, error)
	PlaceOrder(signal *types.Signal, clientId string) (hs.Order, error)
}

func New(ex exchange.Exchange) Executor {
	return &TurtleExecutor{
		ex: ex,
	}
}
