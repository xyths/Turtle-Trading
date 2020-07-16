package executor

import (
	"github.com/shopspring/decimal"
	"github.com/xyths/Turtle-Trading/portfolio"
)

// for test
type DummyExecutor struct {
}

func New(portfolio portfolio.Portfolio) *DummyExecutor {
	return &DummyExecutor{}
}

func (d *DummyExecutor) Buy(price, amount decimal.Decimal) {

}

func (d *DummyExecutor) Sell(price, amount decimal.Decimal) {

}
