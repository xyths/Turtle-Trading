package executor

import (
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/xyths/Turtle-Trading/exchange"
	"github.com/xyths/Turtle-Trading/types"
	"github.com/xyths/hs"
	"github.com/xyths/hs/logger"
)

// for test
type TurtleExecutor struct {
	config Config
	ex     exchange.Exchange

	// for log and database
	quoteCurrency string
	baseCurrency  string
	feeCurrency   string
	// for exchange request
	symbol    string
	feeSymbol string

	minAmount       decimal.Decimal
	minTotal        decimal.Decimal
	pricePrecision  int
	amountPrecision int
}

func NewTurtleExecutor(config Config, ex exchange.Exchange) *TurtleExecutor {
	t := &TurtleExecutor{config: config, ex: ex}
	//t.parseSymbol()
	return t
}

//
//func (t *TurtleExecutor) parseSymbol() {
//	symbols := t.ex.symbol()
//	if len(symbols) < 1 {
//		logger.Sugar.Fatalf("no symbols set for exchange")
//	}
//	tokens := strings.Split(symbols[0], "/")
//	t.baseCurrency = tokens[0]
//	t.quoteCurrency = tokens[1]
//	if t.config.FeeCurrency == "" {
//		logger.Sugar.Fatalf("no fee currency")
//	}
//	t.feeCurrency = t.config.FeeCurrency
//}

// BTC/USDT -> USDT
func (t *TurtleExecutor) QuoteCurrency() string {
	return t.quoteCurrency
}

// BTC/USDT -> BTC
func (t *TurtleExecutor) BaseCurrency() string {
	return t.baseCurrency
}

func (t *TurtleExecutor) FeeCurrency() string {
	return t.feeCurrency
}

func (t *TurtleExecutor) Balance() (cash, currency, fee decimal.Decimal) {
	return t.ex.Balance()
}

// price of the base currency
func (t *TurtleExecutor) Price() (decimal.Decimal, error) {
	return t.ex.Price()
}

func (t *TurtleExecutor) FeePrice() (decimal.Decimal, error) {
	return t.ex.FeePrice()
}

func (t *TurtleExecutor) MinAmount() decimal.Decimal {
	return decimal.Zero
}

func (t *TurtleExecutor) MinTotal() decimal.Decimal {
	return decimal.Zero
}

func (t *TurtleExecutor) PricePrecision() int32 {
	return 2
}

func (t *TurtleExecutor) AmountPrecision() int32 {
	return 5
}

func (t *TurtleExecutor) Buy(price, amount decimal.Decimal, clientId string) (hs.Order, error) {
	return t.ex.Buy(price, amount, clientId)
}

func (t *TurtleExecutor) Sell(price, amount decimal.Decimal, clientId string) (hs.Order, error) {
	return t.ex.Sell(price, amount, clientId)
}

func (t *TurtleExecutor) PlaceOrder(signal *types.Signal, clientId string) (hs.Order, error) {
	logger.Sugar.Debugf("place order, clientId: %s, direction: %d, price: %s, amount: %s", clientId, signal.Direction, signal.Price, signal.Amount)
	if signal.Direction == types.Buy {
		return t.Buy(signal.Price, signal.Amount, clientId)
	} else if signal.Direction == types.Sell {
		return t.Sell(signal.Price, signal.Amount, clientId)
	}
	return hs.Order{}, errors.New("unknown order type")
}
