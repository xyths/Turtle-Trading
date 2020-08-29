package turtle

import (
	"context"
	"errors"
	"fmt"
	"github.com/markcheno/go-talib"
	"github.com/shopspring/decimal"
	"github.com/xyths/Turtle-Trading/cmd/utils"
	"github.com/xyths/Turtle-Trading/exchange"
	"github.com/xyths/Turtle-Trading/executor"
	"github.com/xyths/Turtle-Trading/portfolio"
	"github.com/xyths/Turtle-Trading/types"
	"github.com/xyths/hs"
	"github.com/xyths/hs/logger"
	"time"
)

type Turtle struct {
	Config Config

	tickerCh chan exchange.Ticker
	candleCh chan exchange.Candle
	ex       exchange.Exchange

	portfolio portfolio.Portfolio
	executor  executor.Executor

	candle    exchange.Candle
	position  int
	buyTimes  int
	sellTimes int
}

type Config struct {
	CandleCapacity int `json:"candleCapacity"`
	MaxBuyTimes    int `json:"maxBuyTimes"`
}

const (
	DefaultChannelBuffer  = 3
	DefaultCandleCapacity = 1000
)

func New(conf Config, ex exchange.Exchange, executor executor.Executor, portfolio portfolio.Portfolio) *Turtle {
	t := &Turtle{
		Config:    conf,
		ex:        ex,
		executor:  executor,
		portfolio: portfolio,
		candle:    hs.NewCandle(conf.CandleCapacity),
	}
	if t.Config.MaxBuyTimes <= 0 {
		t.Config.MaxBuyTimes = defaultMaxBuyTimes
	}
	t.tickerCh = make(chan exchange.Ticker, DefaultChannelBuffer)
	t.candleCh = make(chan exchange.Candle, DefaultChannelBuffer)
	ex.SetTickerChannel(t.tickerCh)
	t.ex.SetCandleChannel(t.candleCh)
	return t
}

const (
	periodATR   = 14
	periodUpper = 20
	periodLower = 10

	defaultMaxBuyTimes = 3

	positionLongFull  = 2
	positionLongOpen  = 1
	positionEmpty     = 0
	positionShortOpen = -1
	postionShortFull  = -2
)

func (t *Turtle) OnTicker(ticker exchange.Ticker) (*types.Signal, error) {
	//t.rateLimit <- 1
	//defer func() { <-t.rateLimit }()

	t.candle.Append(ticker)

	return t.signal()
}

func (t *Turtle) OnCandle(candle exchange.Candle) (*types.Signal, error) {
	//t.rateLimit <- 1
	//defer func() { <-t.rateLimit }()

	t.candle.Add(candle)

	return t.signal()
}

var hundred decimal.Decimal = decimal.NewFromInt(100)

func (t *Turtle) signal() (signal *types.Signal, err error) {
	//logger.Sugar.Debugw(
	//	"candle",
	//	"timestamp", t.candle.Timestamp[0],
	//	"high", t.candle.High[0],
	//	"low", t.candle.Low[0],
	//	"close", t.candle.Close[0],
	//)
	l := t.candle.Length()
	if l < periodATR || l < periodLower || l < periodUpper {
		return
	}
	atrs := talib.Atr(t.candle.High, t.candle.Low, t.candle.Close, periodATR)
	atr := atrs[len(atrs)-2]
	//logger.Sugar.Debugf("atr: %v", atr)
	N := decimal.NewFromFloat(atrs[len(atrs)-1])
	if N.IsZero() {
		return signal, errors.New("atr is zero")
	}
	uppers := talib.Max(t.candle.High, periodUpper)
	upper := uppers[len(uppers)-2]
	lowers := talib.Min(t.candle.Low, periodLower)
	lower := lowers[len(lowers)-2]
	logger.Sugar.Debugf("Ticker[%d] %d, %f, %f, %f, %f, Upper: %f, Lower: %f, ATR: %f",
		len(atrs)-1, t.candle.Timestamp[l-1],
		t.candle.Open[l-1], t.candle.High[l-1], t.candle.Low[l-1], t.candle.Close[l-1],
		upper, lower, atr,
	)

	secondsEastOfUTC := int((8 * time.Hour).Seconds())
	beijing := time.FixedZone("Beijing Time", secondsEastOfUTC)
	timestamp := time.Unix(t.candle.Timestamp[l-1], 0).In(beijing).Format(utils.TimeLayout)

	unit := t.portfolio.Value(decimal.NewFromFloat(t.candle.Close[l-1])).Div(hundred).Div(N)

	if t.candle.High[l-1] >= upper {
		logger.Sugar.Infof("突破上轨, Timestamp: %s, Upper: %f, High: %f", timestamp, uppers[l-2], t.candle.High[l-1])
		if t.position == positionEmpty {
			t.buyTimes = 1
			//unit := t.portfolio.Value(decimal.NewFromFloat(t.candle.Close[l-1])).Div(hundred).Div(N)
			logger.Sugar.Debugf("N: %s, unit: %s", N, unit)
			// open position
			price := decimal.NewFromFloat(upper).Round(t.executor.PricePrecision())
			amount := unit.Round(t.executor.AmountPrecision())
			total := price.Mul(amount)
			if total.GreaterThan(t.portfolio.Cash()) {
				amount = t.portfolio.Cash().Div(price).Round(t.executor.AmountPrecision())
				total = price.Mul(amount)
			}
			if amount.GreaterThanOrEqual(t.executor.MinAmount()) && total.GreaterThanOrEqual(t.executor.MinTotal()) {
				signal = &types.Signal{
					Direction:
					types.Buy,
					Price:  price,
					Amount: amount,
				}
				t.position = positionLongOpen
			} else {
				logger.Sugar.Infow("下单数量太小", "price", price, "amount", amount, "total", total)
			}
		} else {
			logger.Sugar.Info("目前不是空仓，无需操作")
		}
	}
	if signal == nil && t.candle.Low[l-1] <= lowers[l-2] {
		logger.Sugar.Infof("突破下轨, Timestamp: %s, Lower: %f, Low: %f", timestamp, lowers[l-2], t.candle.Low[l-1])
		if t.position >= positionLongOpen {
			t.buyTimes = 0
			t.sellTimes++
			// try to close position
			price := decimal.NewFromFloat(lower).Round(t.executor.PricePrecision())
			amount := t.portfolio.Currency().Round(t.executor.AmountPrecision())
			total := price.Mul(amount)
			if amount.GreaterThanOrEqual(t.executor.MinAmount()) && total.GreaterThanOrEqual(t.executor.MinTotal()) {
				signal = &types.Signal{
					Direction: types.Sell,
					Price:     price,
					Amount:    amount,
				}
				t.position = positionEmpty
			} else {
				logger.Sugar.Infow("下单数量太小", "price", price, "amount", amount, "total", total)
			}
		} else {
			logger.Sugar.Info("目前是空仓，无需操作")
		}
	}
	if signal == nil && decimal.NewFromFloat(t.candle.Low[l-1]).LessThanOrEqual(t.portfolio.LastBuyPrice().Sub(N.Mul(decimal.NewFromFloat(2)))) {
		logger.Sugar.Infof("下跌超过2N, Timestamp: %s, N: %s, LastBuy: %s, Low: %f", timestamp, N, t.portfolio.LastBuyPrice(), t.candle.Low[l-1])
		if t.position >= positionLongOpen {
			t.buyTimes = 0
			t.sellTimes++
			// try to close position
			price := t.portfolio.LastBuyPrice().Sub(N.Mul(decimal.NewFromFloat(2))).Round(t.executor.PricePrecision())
			amount := t.portfolio.Currency().Round(t.executor.AmountPrecision())
			total := price.Mul(amount)
			if amount.GreaterThanOrEqual(t.executor.MinAmount()) && total.GreaterThanOrEqual(t.executor.MinTotal()) {
				signal = &types.Signal{
					Direction: types.Sell,
					Price:     price,
					Amount:    amount,
				}
				t.position = positionEmpty
			} else {
				logger.Sugar.Infow("下单数量太小", "price", price, "amount", amount, "total", total)
			}
		} else {
			logger.Sugar.Info("目前是空仓，无需操作")
		}
	}
	if signal == nil && decimal.NewFromFloat(t.candle.High[l-1]).GreaterThanOrEqual(t.portfolio.LastBuyPrice().Add(N.Mul(decimal.NewFromFloat(0.5)))) {
		logger.Sugar.Infof("上涨超过0.5N, Timestamp: %s, N: %s, LastBuy: %s, High: %f", timestamp, N, t.portfolio.LastBuyPrice(), t.candle.High[l-1])
		// add position
		if t.position == positionLongOpen && t.buyTimes < t.Config.MaxBuyTimes {
			// add position
			price := t.portfolio.LastBuyPrice().Add(N.Mul(decimal.NewFromFloat(0.5))).Round(t.executor.PricePrecision())
			amount := unit.Round(t.executor.AmountPrecision())
			total := price.Mul(amount)
			if total.GreaterThan(t.portfolio.Cash()) {
				amount = t.portfolio.Cash().Div(price).Round(t.executor.AmountPrecision())
				total = price.Mul(amount)
			}
			if amount.GreaterThanOrEqual(t.executor.MinAmount()) && total.GreaterThanOrEqual(t.executor.MinTotal()) {
				signal = &types.Signal{
					Direction: types.Buy,
					Price:     price,
					Amount:    amount,
				}
				t.buyTimes++
				if t.buyTimes == t.Config.MaxBuyTimes {
					t.position = positionLongFull
				}
			} else {
				logger.Sugar.Infow("下单数量太小", "price", price, "amount", amount, "total", total)
			}
		} else if t.position == positionLongFull {
			logger.Sugar.Info("目前是满仓，无需操作")
		} else if t.position == positionEmpty {
			logger.Sugar.Info("目前是空仓，无需操作")
		}
	}

	return
}

func (t *Turtle) Start(ctx context.Context) {
	// init portfolio first
	//cash, currency, fee := t.executor.Balance()
	//price,err := t.ex.Price()
	//t.portfolio.Init(cash, currency, fee, price)

	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Sugar.Info("turtle strategy stopped")
				return
			case ticker := <-t.tickerCh:
				//logger.Sugar.Debugf("ticker received: %v", ticker)
				logger.Sugar.Debugf("profit: %s", t.portfolio.Profit(decimal.NewFromFloat(ticker.Close)))
				signal, err := t.OnTicker(ticker)
				if err != nil {
					logger.Sugar.Errorf("onTicker error: %s", err)
				}
				if signal != nil {
					logger.Sugar.Infow("got signal", "timestamp", ticker.Timestamp, "signal", signal)
					// place order
					t.order(ctx, signal)
				}
			case candle := <-t.candleCh:
				//logger.Sugar.Debugf("candle received, len: %v", candle.Length())
				signal, err := t.OnCandle(candle)
				if err != nil {
					logger.Sugar.Errorf("OnCandle error: %s", err)
				}
				if signal != nil {
					logger.Sugar.Infow("got signal", "timestamp", candle.Timestamp[0], "signal", signal)
					// place order
					t.order(ctx, signal)
				}
			}
		}
	}()

	logger.Sugar.Info("turtle strategy started")
}

func (t *Turtle) order(ctx context.Context, signal *types.Signal) {
	if signal == nil {
		return
	}
	var clientId string
	switch signal.Direction {
	case types.Buy:
		clientId = fmt.Sprintf("b-%d-%d", t.sellTimes, t.buyTimes)
	case types.Sell:
		clientId = fmt.Sprintf("s-%d-%d", t.sellTimes, t.buyTimes)
	}
	o, err := t.executor.PlaceOrder(signal, clientId)
	if err != nil {
		logger.Sugar.Errorf("place order error: %s", err)
		return
	}
	switch o.Type {
	case hs.Buy:
		t.portfolio.Update(o.FilledPrice.Mul(o.FilledAmount).Neg(), o.FilledAmount, o.FilledPrice, o.Fee)
	case hs.Sell:
		t.portfolio.Update(o.FilledPrice.Mul(o.FilledAmount), o.FilledAmount.Neg(), o.FilledPrice, o.Fee)
	}

}

//func (t *Turtle) Stop(ctx context.Context) {
//}

func (t *Turtle) open(price, amount decimal.Decimal) {
}

func (t *Turtle) add(price, amount decimal.Decimal) {
}

func (t *Turtle) clear(price, amount decimal.Decimal) {
}

func reverse(in []float64) (out []float64) {
	for i := len(in) - 1; i >= 0; i-- {
		out = append(out, in[i])
	}
	return out
}

func average(open, high, low, close float64) float64 {
	return (open + high + low + close) / 4
}
