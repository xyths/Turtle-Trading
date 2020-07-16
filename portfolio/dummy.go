package portfolio

import (
	"context"
	"github.com/shopspring/decimal"
	"github.com/xyths/hs/logger"
)

type DummyPortfolio struct {
}

func New() *DummyPortfolio {
	return &DummyPortfolio{}
}

func (d *DummyPortfolio) Value() decimal.Decimal {
	logger.Sugar.Info("dummy portfolio value: 0")
	return decimal.Zero
}

func (d *DummyPortfolio) Profit() decimal.Decimal {
	logger.Sugar.Info("dummy portfolio profit: 0")
	return decimal.Zero
}

func (d *DummyPortfolio) Start(ctx context.Context) {
	logger.Sugar.Info("dummy started")
}

//func (d *DummyPortfolio) Stop(ctx context.Context) {
//	logger.Info("dummy stopped")
//}
