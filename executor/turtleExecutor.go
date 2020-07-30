package executor

import (
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/xyths/Turtle-Trading/exchange"
	"github.com/xyths/Turtle-Trading/types"
	"github.com/xyths/hs"
	"github.com/xyths/hs/logger"
	"strings"
)

// for test
type TurtleExecutor struct {
	ex exchange.Exchange

	quoteCurrency string
	baseCurrency  string
	feeCurrency   string

	minAmount       decimal.Decimal
	minTotal        decimal.Decimal
	pricePrecision  int
	amountPrecision int
}

func (t *TurtleExecutor) parseSymbol() {
	symbols := t.ex.Symbols()
	if len(symbols) < 1 {
		logger.Sugar.Fatalf("no symbols set for exchange")
	}
	tokens := strings.Split(symbols[0], "/")
	t.baseCurrency = tokens[0]
	t.quoteCurrency = tokens[1]
}

// BTC/USDT -> USDT
func (t *TurtleExecutor) QuoteCurrency() string {
	if t.quoteCurrency == "" {
		t.parseSymbol()
	}
	return t.quoteCurrency
}

// BTC/USDT -> BTC
func (t *TurtleExecutor) BaseCurrency() string {
	if t.quoteCurrency == "" {
		t.parseSymbol()
	}
	return t.baseCurrency
}

func (t *TurtleExecutor) FeeCurrency() string {
	if t.quoteCurrency == "" {
		t.feeCurrency = t.ex.FeeCurrency()
	}
	return t.feeCurrency
}

func (t *TurtleExecutor) Balance() (cash, currency, fee decimal.Decimal) {
	balance := t.ex.Balance()
	cash = balance[t.QuoteCurrency()]
	currency = balance[t.BaseCurrency()]
	fee = balance[t.FeeCurrency()]
	return
}

func (t *TurtleExecutor) MinAmount() decimal.Decimal {
	return decimal.Zero
}

func (t *TurtleExecutor) MinTotal() decimal.Decimal {
	return decimal.Zero
}

func (t *TurtleExecutor) PricePrecision() int {
	return 0
}

func (t *TurtleExecutor) AmountPrecision() int {
	return 0
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
