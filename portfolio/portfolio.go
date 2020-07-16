package portfolio

import (
	"context"
	"github.com/shopspring/decimal"
)

type Portfolio interface {
	//Setup(db *mongo.Database, ex exchange.Exchange)
	Value() decimal.Decimal
	Profit() decimal.Decimal
	Start(ctx context.Context)
	//Stop(ctx context.Context)
}
