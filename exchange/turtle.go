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
	"time"
)

const clientId = "turtle"

type TurtleExchange struct {
	config Config

	quoteCurrency string
	baseCurrency  string
	symbol        string
	feeCurrency   string
	feeSymbol     string

	clientId string
	ex       hs.Exchange
	tickerCh chan Ticker
	candleCh chan Candle
}

func NewTurtle(config Config) *TurtleExchange {
	var ex hs.Exchange
	switch config.Name {
	case hs.Huobi:
		ex = huobi.New(config.Label, config.Key, config.Secret, config.Host)
		t := &TurtleExchange{
			config:      config,
			symbol:      config.Symbols[0],
			feeCurrency: config.FeeCurrency,
			clientId:    config.ClientId,
			ex:          ex,
		}
		t.quoteCurrency = "usdt"
		t.baseCurrency = "btc"
		t.feeSymbol = t.feeCurrency + t.quoteCurrency
		return t
	default:
		log.Fatalf("exchange %s is not supported", config.Name)
	}
	return nil
}

func (t *TurtleExchange) QuoteCurrency() string {
	return t.quoteCurrency
}

func (t *TurtleExchange) BaseCurrency() string {
	return t.baseCurrency
}

func (t *TurtleExchange) Symbol() string {
	return t.symbol
}

func (t *TurtleExchange) FeeCurrency() string {
	return t.config.FeeCurrency
}

func (t *TurtleExchange) Balance() (cash, currency, fee decimal.Decimal) {
	b, err := t.ex.GetSpotBalance()
	if err != nil {
		logger.Sugar.Errorf("get balance error: %s", err)
		return
	}
	cash = b[t.quoteCurrency]
	currency = b[t.baseCurrency]
	fee = b[t.feeCurrency]
	return
}

func (t *TurtleExchange) Price() (decimal.Decimal, error) {
	return t.ex.GetPrice(t.symbol)
}
func (t *TurtleExchange) FeePrice() (decimal.Decimal, error) {
	return t.ex.GetPrice(t.feeSymbol)
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
func (t *TurtleExchange) PricePrecision() int32 {
	return t.ex.PricePrecision(t.symbol)
}
func (t *TurtleExchange) AmountPrecision() int32 {
	return t.ex.AmountPrecision(t.symbol)
}

func (t *TurtleExchange) Buy(price, amount decimal.Decimal, clientId string) (hs.Order, error) {
	return hs.Order{}, nil
}

func (t *TurtleExchange) Sell(price, amount decimal.Decimal, clientId string) (hs.Order, error) {
	return hs.Order{}, nil
}

func (t *TurtleExchange) Start(ctx context.Context) {
	period, err := time.ParseDuration(t.config.Period)
	if err != nil {
		logger.Sugar.Fatal("bad period: %s", err)
	}
	go t.ex.SubscribeCandlestickWithReq(ctx, t.symbol, t.clientId, period, t.tickerHandler)
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
				logger.Sugar.Infof("Candlestick update, id: %d, count: %v, volume: %v, OHLC[%v, %v, %v, %v]",
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
					logger.Sugar.Infof("Candlestick data[%d], id: %d, count: %v, volume: %v, OHLC[%v, %v, %v, %v]",
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
