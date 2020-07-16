package main

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"github.com/xyths/Turtle-Trading/cmd/utils"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var app *cli.App

func init() {
	log.SetFlags(log.Ldate | log.Ltime)

	app = &cli.App{
		Name:    filepath.Base(os.Args[0]),
		Usage:   "the turtle trading robot",
		Version: "0.1.0",
		Action:  trade,
	}

	app.Commands = []*cli.Command{
		tradeCommand,
		testCommand,
		downloadCommand,
	}
	app.Flags = []cli.Flag{
		utils.ConfigFlag,
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		cancel()
	}()

	if err := app.RunContext(ctx, os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
