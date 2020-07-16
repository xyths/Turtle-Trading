package exchange

import (
	"context"
	"github.com/huobirdcenter/huobi_golang/logging/applogger"
	"github.com/huobirdcenter/huobi_golang/pkg/response/market"
	"github.com/shopspring/decimal"
	"github.com/xyths/hs"
	"github.com/xyths/hs/exchange/huobi"
	"github.com/xyths/hs/logger"
	"log"
)

const clientId = "turtle"

type TurtleExchange struct {
	Symbol   string
	clientId string
	ex       hs.Exchange
	tickerCh chan Ticker
	candleCh chan Candle
}

func NewTurtle(config hs.ExchangeConf) *TurtleExchange {
	var ex hs.Exchange
	switch config.Name {
	case hs.Huobi:
		ex = huobi.New(config.Label, config.Key, config.Secret, config.Host)
	default:
		log.Fatalf("exchange %s is not supported", config.Name)
	}
	t := &TurtleExchange{
		Symbol:   config.Symbols[0],
		clientId: clientId,
		ex:       ex,
	}
	return t
}

func (t *TurtleExchange) SetTickerChannel(tickerCh chan Ticker) {
	t.tickerCh = tickerCh
}
func (t *TurtleExchange) SetCandleChannel(candleCh chan Candle) {
	t.candleCh = candleCh
}

func (t *TurtleExchange) MinAmount() decimal.Decimal {
	return decimal.Zero
}
func (t *TurtleExchange) MinTotal() decimal.Decimal {
	return decimal.Zero
}
func (t *TurtleExchange) PricePrecision() int {
	return 0
}
func (t *TurtleExchange) AmountPrecision() int {
	return 0
}

func (t *TurtleExchange) Start(ctx context.Context) {
	go t.ex.SubscribeCandlestick(ctx, t.Symbol, t.clientId, t.tickerHandler)
	logger.Sugar.Info("exchange started")

	<-ctx.Done()

	logger.Sugar.Info("exchange stopped")

}

func (t *TurtleExchange) tickerHandler(resp interface{}) {
	candlestickResponse, ok := resp.(market.SubscribeCandlestickResponse)
	if ok {
		if &candlestickResponse != nil {
			if candlestickResponse.Tick != nil {
				tick := candlestickResponse.Tick
				logger.Sugar.Info("Candlestick update, id: %d, count: %v, volume: %v, OHLC[%v, %v, %v, %v]",
					tick.Id, tick.Count, tick.Vol, tick.Open, tick.High, tick.Low, tick.Close)
				ticker := hs.Ticker{
					Timestamp: tick.Id,
				}
				ticker.Open, _ = tick.Open.Float64()
				ticker.High, _ = tick.High.Float64()
				ticker.Low, _ = tick.Low.Float64()
				ticker.Close, _ = tick.Close.Float64()
				ticker.Volume, _ = tick.Vol.Float64()

				select {
				case t.tickerCh <- ticker:
					logger.Sugar.Debugf("ticker sent: %v", ticker)
				default:
					logger.Sugar.Error("no buffer in ticker channel")
				}
			}

			if candlestickResponse.Data != nil {
				var candle hs.Candle
				for i, tick := range candlestickResponse.Data {
					logger.Sugar.Info("Candlestick data[%d], id: %d, count: %v, volume: %v, OHLC[%v, %v, %v, %v]",
						i, tick.Id, tick.Count, tick.Vol, tick.Open, tick.High, tick.Low, tick.Close)
					ticker := hs.Ticker{
						Timestamp: tick.Id,
					}
					ticker.Open, _ = tick.Open.Float64()
					ticker.High, _ = tick.High.Float64()
					ticker.Low, _ = tick.Low.Float64()
					ticker.Close, _ = tick.Close.Float64()
					ticker.Volume, _ = tick.Vol.Float64()
					candle.Append(ticker)
				}
				select {
				case t.candleCh <- candle:
					logger.Sugar.Debugf("candle sent, length: %d", candle.Length())
				default:
					logger.Sugar.Error("no buffer in ticker channel")
				}
			}
		}
	} else {
		applogger.Warn("Unknown response: %v", resp)
	}
}
