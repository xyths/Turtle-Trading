package node

import (
	"context"
	"github.com/shopspring/decimal"
	"github.com/xyths/Turtle-Trading/exchange"
	"github.com/xyths/Turtle-Trading/executor"
	"github.com/xyths/Turtle-Trading/portfolio"
	"github.com/xyths/Turtle-Trading/strategy"
	"github.com/xyths/Turtle-Trading/turtle"
	"github.com/xyths/hs"
	"github.com/xyths/hs/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

type Config struct {
	Mongo    hs.MongoConf
	Exchange exchange.Config
	Executor executor.Config
	Strategy strategy.Config
}

type Node struct {
	config Config

	db        *mongo.Database
	executor  executor.Executor
	exchange  exchange.Exchange
	portfolio portfolio.Portfolio
	strategy  strategy.Strategy
}

func New(conf Config) *Node {
	return &Node{config: conf}
}

func (n *Node) Init(ctx context.Context) {
	if db, err := hs.ConnectMongo(ctx, n.config.Mongo); err != nil {
		logger.Sugar.Fatal(err)
	} else {
		n.db = db
	}
	n.initEx(ctx)
	n.initPortfolio(ctx)
	n.initStrategy(ctx)
}

func (n *Node) initEx(ctx context.Context) {
	n.exchange = exchange.New(n.config.Exchange)
	n.executor = executor.New(n.config.Executor, n.exchange)
}

func (n *Node) initPortfolio(ctx context.Context) {
	n.portfolio = portfolio.New(n.db)
	if !n.portfolio.Load(ctx) {
		// get balance
		cash, currency, fee := n.executor.Balance()
		price, err := n.executor.Price()
		if err != nil {
			logger.Sugar.Fatalf("get price error: %s", err)
		}
		//feePrice, err := n.executor.FeePrice()
		//if err != nil {
		//	logger.Sugar.Fatalf("get fee price error: %s", err)
		//}
		feeMap := make(map[string]decimal.Decimal)
		feeMap[n.executor.FeeCurrency()] = fee
		n.portfolio.Init(cash, currency, price, feeMap)
		//_ = feePrice
		if err := n.portfolio.Save(ctx); err != nil {
			logger.Sugar.Fatalf("portfolio save error: %s", err)
		}
	}
}

func (n *Node) initStrategy(ctx context.Context) {
	n.strategy = turtle.New(turtle.Config{2000, 3}, n.exchange, n.executor, n.portfolio)
}

func (n *Node) Close(ctx context.Context) {
	if n.db != nil {
		if err := n.db.Client().Disconnect(ctx); err != nil {
			logger.Sugar.Error(err)
		}
	}
	logger.Sugar.Info("turtle node stopped")
}

func (n *Node) Trade(ctx context.Context) error {
	n.strategy.Start(ctx)
	n.exchange.Start(ctx)

	// serve
	logger.Sugar.Info("turtle node started")
	<-ctx.Done()

	return nil
}

func (n *Node) Run(ctx context.Context) error {
	n.strategy.Start(ctx)
	n.exchange.Start(ctx)

	// serve
	logger.Sugar.Info("turtle node(local) started")
	<-ctx.Done()

	return nil
}
