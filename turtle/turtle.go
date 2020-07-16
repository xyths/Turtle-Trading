package turtle

import (
	"context"
	"github.com/markcheno/go-talib"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/xyths/Turtle-Trading/exchange"
	"github.com/xyths/Turtle-Trading/executor"
	"github.com/xyths/Turtle-Trading/portfolio"
	"github.com/xyths/Turtle-Trading/types"
	"github.com/xyths/hs/logger"
)

type Turtle struct {
	tickerCh chan exchange.Ticker
	candleCh chan exchange.Candle
	ex       exchange.Exchange

	portfolio portfolio.Portfolio
	executor  executor.Executor
}

const ChannelBuffer = 3

func New(ex exchange.Exchange, executor executor.Executor, portfolio portfolio.Portfolio) *Turtle {
	t := &Turtle{ex: ex, executor: executor, portfolio: portfolio}
	t.tickerCh = make(chan exchange.Ticker, ChannelBuffer)
	t.candleCh = make(chan exchange.Candle, ChannelBuffer)
	ex.SetTickerChannel(t.tickerCh)
	t.ex.SetCandleChannel(t.candleCh)
	return t
}

const (
	periodATR   = 14
	periodUpper = 20
	periodLower = 10
)

func (t *Turtle) OnTicker(quote types.Quote) (types.Signal, error) {
	atr := talib.Atr(quote.High, quote.Low, quote.Close, periodATR)
	N := decimal.NewFromFloat(atr[len(atr)-1])
	if N.IsZero() {
		return types.Signal{}, errors.New("atr is zero")
	}
	hundred := decimal.NewFromInt(100)
	unit := t.portfolio.Value().Div(hundred).Div(N)
	logger.Sugar.Debugf("atr: %f, unit: %s", atr[len(atr)-1], unit.String())
	upper := talib.Max(quote.High, periodUpper)
	lower := talib.Min(quote.Low, periodLower)
	if talib.Crossover(quote.Close, upper) {
		// try to open position or add
	} else if talib.Crossunder(quote.Close, lower) {
		// try to close position
	}
	return types.Signal{}, nil
}

func (t *Turtle) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Sugar.Info("turtle strategy stopped")
				return
			case ticker := <-t.tickerCh:
				logger.Sugar.Debugf("ticker received: %v", ticker)
				//go t.OnTicker()
			case quote := <-t.candleCh:
				logger.Sugar.Debugf("quote received: %v", quote)
			}
		}
	}()

	logger.Sugar.Info("turtle strategy started")
}

//func (t *Turtle) Stop(ctx context.Context) {
//}
