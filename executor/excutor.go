package executor

import "github.com/shopspring/decimal"

type Executor interface {
	Buy(price, amount decimal.Decimal)
	Sell(price, amount decimal.Decimal)
}
