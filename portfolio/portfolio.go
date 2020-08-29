package portfolio

import (
	"context"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/mongo"
)

type Portfolio interface {
	Load(ctx context.Context) bool
	Save(ctx context.Context) error
	Init(cash, currency, price decimal.Decimal, fee map[string]decimal.Decimal)
	Cash() decimal.Decimal
	Currency() decimal.Decimal
	Value(price decimal.Decimal) decimal.Decimal
	Update(cash, currency, price decimal.Decimal, fee map[string]decimal.Decimal)
	LastBuyPrice() decimal.Decimal
	Profit(price decimal.Decimal) decimal.Decimal
	Start(ctx context.Context)
	//Stop(ctx context.Context)
}

func New(db *mongo.Database) Portfolio {
	return &turtlePortfolio{
		db:  db,
		fee: make(map[string]decimal.Decimal),
	}
}
