package portfolio

import (
	"github.com/shopspring/decimal"
	"testing"
)

func TestTurtlePortfolio_Init(t *testing.T) {
	p := New()
	cash := decimal.NewFromFloat(1000)
	currency := decimal.NewFromFloat(300)
	fee := decimal.NewFromFloat(100)
	price := decimal.NewFromFloat(9999)
	p.Init(cash, currency, fee, price)
	if !p.Currency().Equal(currency) {
		t.Errorf("currency not equal, expect: %s, actual: %s", currency, p.Currency())
	}
	expect := cash.Add(currency.Mul(price))
	actual := p.Value(price)
	if !actual.Equal(expect) {
		t.Errorf("value expect: %s, actual: %s", expect, actual)
	}
}

func TestTurtlePortfolio_Value(t *testing.T) {
	var tests = []struct {
		cash     decimal.Decimal
		currency decimal.Decimal
		fee      decimal.Decimal
		price    decimal.Decimal
		value    decimal.Decimal
	}{
		{decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero},
		{decimal.NewFromFloat(1), decimal.Zero, decimal.Zero, decimal.Zero, decimal.NewFromFloat(1)},
		{decimal.NewFromFloat(1), decimal.NewFromFloat(3), decimal.NewFromFloat(3), decimal.NewFromFloat(3), decimal.NewFromFloat(10)},
	}
	p := New()

	for i, tt := range tests {
		p.Init(tt.cash, tt.currency, tt.fee, tt.price)
		if !p.Value(tt.price).Equal(tt.value) {
			t.Errorf("[%d] expect: %s, actual: %s", i, tt.value, p.Value(tt.price))
		}
	}
}

func TestTurtlePortfolio_LastBuyPrice(t *testing.T) {
	p := New()
	cash := decimal.NewFromFloat(1000)
	currency := decimal.NewFromFloat(0)
	fee := decimal.NewFromFloat(100)
	price := decimal.NewFromFloat(10)
	p.Init(cash, currency, fee, price)

	t.Run("buy at 10", func(t *testing.T) {
		p.Update(decimal.NewFromFloat(-10), decimal.NewFromFloat(1), price, map[string]decimal.Decimal{"HT": decimal.NewFromFloat(1)})
		if !p.LastBuyPrice().Equal(price) {
			t.Errorf("LastBuyPrice expect: %s, actual: %s", price, p.LastBuyPrice())
		}
		price1 := decimal.NewFromFloat(5)
		profit := decimal.NewFromFloat(-5)
		if !p.Profit(price1).Equal(profit) {
			t.Errorf("profit wrong, expect: %s, actual: %s", profit, p.Profit(price1))
		}
	})
	t.Run("sell at 11", func(t *testing.T) {
		price2 := decimal.NewFromFloat(11)
		p.Update(decimal.NewFromFloat(11), decimal.NewFromFloat(-1), price2, map[string]decimal.Decimal{"HT": decimal.NewFromFloat(1)})
		if !p.LastBuyPrice().Equal(price) {
			t.Errorf("LastBuyPrice expect: %s, actual: %s", price, p.LastBuyPrice())
		}
		profit := decimal.NewFromFloat(1)
		if !p.Profit(price2).Equal(profit) {
			t.Errorf("profit wrong, expect: %s, actual: %s", profit, p.Profit(price2))
		}
	})
	t.Run("buy second time at 5", func(t *testing.T) {
		price3 := decimal.NewFromFloat(5)
		p.Update(decimal.NewFromFloat(-5), decimal.NewFromFloat(1), price3, map[string]decimal.Decimal{"HT": decimal.NewFromFloat(1)})
		if !p.LastBuyPrice().Equal(price3) {
			t.Errorf("LastBuyPrice expect: %s, actual: %s", price3, p.LastBuyPrice())
		}
	})
}
