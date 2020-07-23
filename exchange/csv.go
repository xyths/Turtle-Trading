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
	Symbol    string
	candleCsv string
	startTime time.Time

	tickerCh chan Ticker
	candleCh chan Candle
}

type CsvExchangeConfig struct {
	File      string `json:"file"`
	StartTime string `json:"startTime"`
}

func NewCsvExchange(config CsvExchangeConfig) *CsvExchange {
	startTime, err := utils.ParseTime(config.StartTime)
	if err != nil {
		logger.Sugar.Fatalf("start time format error: %s", err)
	}
	return &CsvExchange{
		candleCsv: config.File,
		startTime: startTime,
	}
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
func (c *CsvExchange) PricePrecision() int {
	return 0
}
func (c *CsvExchange) AmountPrecision() int {
	return 0
}

func (c *CsvExchange) Start(ctx context.Context) {
	file, err := os.Open(c.candleCsv)
	if err != nil {
		log.Fatalf("can not open the file, err is %+v", err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
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
