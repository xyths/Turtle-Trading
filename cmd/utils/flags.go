package utils

import "github.com/urfave/cli/v2"

// common flags for cmd
var (
	ConfigFlag = &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Value:   "config.json",
		Usage:   "load configuration from `file`",
	}
	OutFileFlag = &cli.StringFlag{
		Name:    "out",
		Aliases: []string{"o"},
		Value:   "candles.csv",
		Usage:   "save candle to `file`",
	}
	StartTimeFlag = &cli.StringFlag{
		Name:  "start",
		Value: "",
		Usage: "start `time` (e.g. \"2020-04-01 00:00:00\")",
	}
	EndTimeFlag = &cli.StringFlag{
		Name:  "end",
		Value: "",
		Usage: "end `time` (e.g. \"2020-04-01 23:59:59\")",
	}
)
