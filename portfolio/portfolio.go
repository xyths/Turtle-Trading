package portfolio

import (
	"context"
	"github.com/shopspring/decimal"
)

type Portfolio interface {
	Init(cash, currency, fee, price decimal.Decimal)
	Value(price decimal.Decimal) decimal.Decimal
	Update(cash, currency, price decimal.Decimal, fee map[string]decimal.Decimal)
	LastBuyPrice() decimal.Decimal
	Currency() decimal.Decimal
	Profit(price decimal.Decimal) decimal.Decimal
	Start(ctx context.Context)
	//Stop(ctx context.Context)
}
