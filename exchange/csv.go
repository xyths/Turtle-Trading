package exchange

import (
	"context"
	"encoding/csv"
	"github.com/shopspring/decimal"
	"github.com/xyths/Turtle-Trading/cmd/utils"
	"github.com/xyths/hs"
	"github.com/xyths/hs/logger"
	"log"
	"os"
	"strconv"
	"time"
)

type CsvExchange struct {
	symbol    string
	candleCsv string
	startTime time.Time

	feeCurrency string
	FeeRatio    decimal.Decimal
	feePrice    decimal.Decimal

	tickerCh chan Ticker
	candleCh chan Candle

	globalOrderId uint64
	globalTradeId uint64
	balance       map[string]decimal.Decimal
}

type CsvExchangeConfig struct {
	File      string `json:"file"`
	StartTime string `json:"startTime"`

	Balance     map[string]float64
	FeeCurrency string  `json:"feeCurrency"`
	FeeRatio    float64 `json:"feeRatio"`
	FeePrice    float64 `json:"feePrice"`
}

const DefaultFeePrice = 0.5

func NewCsvExchange(config CsvExchangeConfig, symbols []string) *CsvExchange {
	startTime, err := utils.ParseTime(config.StartTime)
	if err != nil {
		logger.Sugar.Fatalf("start time format error: %s", err)
	}

	c := &CsvExchange{
		symbol:      symbols[0],
		candleCsv:   config.File,
		startTime:   startTime,
		balance:     make(map[string]decimal.Decimal),
		feeCurrency: config.FeeCurrency,
		FeeRatio:    decimal.NewFromFloat(config.FeeRatio),
	}
	for k, v := range config.Balance {
		c.balance[k] = decimal.NewFromFloat(v)
	}
	if config.FeePrice == 0 {
		c.feePrice = decimal.NewFromFloat(DefaultFeePrice)
	} else {
		c.feePrice = decimal.NewFromFloat(config.FeePrice)
	}
	return c
}

func (c *CsvExchange) QuoteCurrency() string {
	return c.symbol
}

func (c *CsvExchange) BaseCurrency() string {
	return c.symbol
}

func (c *CsvExchange) Symbol() string {
	return c.symbol
}

func (c *CsvExchange) FeeCurrency() string {
	return c.feeCurrency
}

func (c *CsvExchange) Balance() (cash, currency, fee decimal.Decimal) {
	return
}

func (c *CsvExchange) Price() (decimal.Decimal, error) {
	return decimal.Decimal{}, nil
}

func (c *CsvExchange) FeePrice() (decimal.Decimal, error) {
	return decimal.Decimal{}, nil
}

func (c *CsvExchange) LastPrice() decimal.Decimal {
	file, err := os.Open(c.candleCsv)
	if err != nil {
		log.Fatalf("can not open the file, err is %+v", err)
	}
	defer func() { _ = file.Close() }()

	r := csv.NewReader(file)
	record, err := r.Read()
	if err != nil {
		log.Fatal(err)
	}
	open := decimal.RequireFromString(record[1])
	high := decimal.RequireFromString(record[2])
	low := decimal.RequireFromString(record[3])
	close1 := decimal.RequireFromString(record[4])
	price := open.Add(high).Add(low).Add(close1).Div(decimal.NewFromInt(4))
	return price
}

func (c *CsvExchange) SetTickerChannel(tickerCh chan Ticker) {
	c.tickerCh = tickerCh
}
func (c *CsvExchange) SetCandleChannel(candleCh chan Candle) {
	c.candleCh = candleCh
}

func (c *CsvExchange) MinAmount() decimal.Decimal {
	return decimal.Zero
}
func (c *CsvExchange) MinTotal() decimal.Decimal {
	return decimal.Zero
}
func (c *CsvExchange) PricePrecision() int32 {
	return 0
}
func (c *CsvExchange) AmountPrecision() int32 {
	return 0
}

func (c *CsvExchange) Buy(price, amount decimal.Decimal, clientId string) (hs.Order, error) {
	c.globalOrderId++
	c.globalTradeId++
	feeAmount := price.Mul(amount).Mul(c.FeeRatio).Div(c.feePrice)
	fee := make(map[string]decimal.Decimal)
	fee[c.feeCurrency] = feeAmount
	trade := hs.Trade{
		Id:          c.globalTradeId,
		Price:       price,
		Amount:      amount,
		FeeCurrency: c.feeCurrency,
		FeeAmount:   feeAmount,
	}
	return hs.Order{
		Id:            c.globalOrderId,
		ClientId:      clientId,
		Type:          hs.Buy,
		Symbol:        c.symbol,
		InitialPrice:  price,
		InitialAmount: amount,
		Timestamp:     time.Now().Unix(),
		// always full filled by initial price
		Status:       hs.Closed,
		FilledPrice:  price,
		FilledAmount: amount,
		Trades:       []hs.Trade{trade},
		Fee:          fee,
	}, nil
}

func (c *CsvExchange) Sell(price, amount decimal.Decimal, clientId string) (hs.Order, error) {
	c.globalOrderId++
	c.globalTradeId++
	feeAmount := price.Mul(amount).Mul(c.FeeRatio).Div(c.feePrice)
	fee := make(map[string]decimal.Decimal)
	fee[c.feeCurrency] = feeAmount
	trade := hs.Trade{
		Id:          c.globalTradeId,
		Price:       price,
		Amount:      amount,
		FeeCurrency: c.feeCurrency,
		FeeAmount:   feeAmount,
	}
	return hs.Order{
		Id:            c.globalOrderId,
		ClientId:      clientId,
		Type:          hs.Sell,
		Symbol:        c.symbol,
		InitialPrice:  price,
		InitialAmount: amount,
		Timestamp:     time.Now().Unix(),
		// always full filled by initial price
		Status:       hs.Closed,
		FilledPrice:  price,
		FilledAmount: amount,
		Trades:       []hs.Trade{trade},
		Fee:          fee,
	}, nil
}

func (c *CsvExchange) Start(ctx context.Context) {
	file, err := os.Open(c.candleCsv)
	if err != nil {
		log.Fatalf("can not open the file, err is %+v", err)
	}
	defer func() { _ = file.Close() }()

	r := csv.NewReader(file)
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	go c.send(ctx, records)
}

func (c *CsvExchange) send(ctx context.Context, records [][]string) {
	candle := hs.NewCandle(len(records))
	i := 0
	for ; i < len(records); i++ {
		r := records[i]
		t, _ := strconv.ParseInt(r[0], 10, 64)
		now := time.Unix(t, 0)
		if now.After(c.startTime) {
			break
		}
		ticker := Ticker{
			Timestamp: t,
		}
		ticker.Open, _ = strconv.ParseFloat(r[1], 64)
		ticker.High, _ = strconv.ParseFloat(r[2], 64)
		ticker.Low, _ = strconv.ParseFloat(r[3], 64)
		ticker.Close, _ = strconv.ParseFloat(r[4], 64)
		ticker.Volume, _ = strconv.ParseFloat(r[5], 64)
		candle.Append(ticker)
	}
	c.candleCh <- candle

	for ; i < len(records); i++ {
		r := records[i]
		t, _ := strconv.ParseInt(r[0], 10, 64)
		ticker := Ticker{
			Timestamp: t,
		}
		ticker.Open, _ = strconv.ParseFloat(r[1], 64)
		ticker.High, _ = strconv.ParseFloat(r[2], 64)
		ticker.Low, _ = strconv.ParseFloat(r[3], 64)
		ticker.Close, _ = strconv.ParseFloat(r[4], 64)
		ticker.Volume, _ = strconv.ParseFloat(r[5], 64)
		c.tickerCh <- ticker
	}
}
