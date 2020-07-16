package main

import (
	"fmt"
	"github.com/huobirdcenter/huobi_golang/pkg/getrequest"
	"github.com/urfave/cli/v2"
	"github.com/xyths/Turtle-Trading/cmd/utils"
	"github.com/xyths/Turtle-Trading/node"
	"github.com/xyths/Turtle-Trading/strategy"
	"github.com/xyths/hs"
	"github.com/xyths/hs/exchange/huobi"
	"github.com/xyths/hs/logger"
	"log"
	"os"
)

var (
	tradeCommand = &cli.Command{
		Action: trade,
		Name:   "trade",
		Usage:  "Trading with turtle strategy",
	}
	downloadCommand = &cli.Command{
		Action: download,
		Name:   "download",
		Usage:  "Download candles to csv",
		Flags: []cli.Flag{
			utils.OutFileFlag,
			utils.StartTimeFlag,
			utils.EndTimeFlag,
		},
	}
	testCommand = &cli.Command{
		Action: localRun,
		Name:   "test",
		Usage:  "Test the turtle strategy parameters",
	}
)

func trade(ctx *cli.Context) error {
	var conf node.Config
	if err := hs.ParseJsonConfig(ctx.String(utils.ConfigFlag.Name), &conf); err != nil {
		logger.Sugar.Fatal(err)
	}
	n := node.New(conf)
	n.Init(ctx.Context)
	defer n.Close(ctx.Context)
	return n.Trade(ctx.Context)
}

func localRun(ctx *cli.Context) error {
	var conf node.Config
	if err := hs.ParseJsonConfig(ctx.String(utils.ConfigFlag.Name), &conf); err != nil {
		logger.Sugar.Fatal(err)
	}
	n := node.New(conf)
	n.Init(ctx.Context)
	defer n.Close(ctx.Context)
	return n.Run(ctx.Context)
}

func download(ctx *cli.Context) error {
	var conf node.Config
	if err := hs.ParseJsonConfig(ctx.String(utils.ConfigFlag.Name), &conf); err != nil {
		logger.Sugar.Fatal(err)
	}

	start := ctx.String(utils.StartTimeFlag.Name)
	end := ctx.String(utils.EndTimeFlag.Name)
	from, to := utils.ParseStartEndTime(start, end)

	exchangeConf := conf.Exchange
	var ex hs.Exchange
	switch exchangeConf.Name {
	case hs.Huobi:
		ex = huobi.New(exchangeConf.Label, exchangeConf.Key, exchangeConf.Secret, exchangeConf.Host)
	default:
		log.Fatalf("exchange %s is not supported", exchangeConf.Name)
	}
	period := getrequest.MIN1
	switch conf.Strategy.CandleType {
	case strategy.CandleType5Min:
		period = getrequest.MIN5
	case strategy.CandleType1H:
		period = getrequest.MIN60
	case strategy.CandleType1D:
		period = getrequest.DAY1
	}
	candles, err := ex.GetCandle(exchangeConf.Symbols[0], "1101", period, from, to)
	if err != nil {
		log.Fatal(err)
	}
	outFile := ctx.String(utils.OutFileFlag.Name)
	f, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	for i := 0; i < candles.Length(); i++ {
		if _, err := fmt.Fprintf(f, "%d,%f,%f,%f,%f,%f\n", candles.Timestamp[i], candles.Open[i], candles.High[i], candles.Low[i], candles.Close[i], candles.Volume[i]); err != nil {
			log.Println(err)
		}
	}

	return nil
}
