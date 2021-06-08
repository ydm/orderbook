package orderbook

import (
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func assertCountLevels(t *testing.T, b *Book, asks, bids int) {
	t.Helper()

	if have := b.asks.heap.CountLevels(); have != asks {
		t.Errorf("have %d, want %d", have, asks)
	}

	if have := b.bids.heap.CountLevels(); have != bids {
		t.Errorf("have %d, want %d", have, bids)
	}
}

// Submit a market order against an empty order book.
func TestBook_AddOrder_1(t *testing.T) {
	b := NewBook()
	order := ClientOrder{
		Side:             SideBuy,
		OriginalQuantity: decimal.NewFromInt(1),
		ExecutedQuantity: decimal.Zero,
		ID:               "id1",
		Type:             TypeMarket,
	}
	assertCountLevels(t, b, 0, 0)
	err := b.AddOrder(order)
	assertCountLevels(t, b, 0, 0)

	if !errors.Is(err, ErrMarketNotFullyExecuted) {
		t.Error()
	}
	x, err := b.GetOrder("id1")
	if err != nil {
		t.Error(err)
	}
	if !x.ExecutedQuantity.IsZero() {
		t.Error()
	}
}

// Submit a market order against a single limit order.
func TestBook_AddOrder_2(t *testing.T) {
	b := NewBook()
	limit := ClientOrder{
		Side:             SideSell,
		OriginalQuantity: decimal.NewFromInt(2),
		ExecutedQuantity: decimal.Zero,
		Price:            decimal.NewFromInt(10_000),
		ID:               "limit",
		Type:             TypeLimit,
	}
	market := ClientOrder{
		Side:             SideBuy,
		OriginalQuantity: decimal.NewFromInt(1),
		ExecutedQuantity: decimal.Zero,
		ID:               "market",
		Type:             TypeMarket,
	}
	assertCountLevels(t, b, 0, 0)
	if err := b.AddOrder(limit); err != nil {
		t.Error(err)
	}
	if err := b.AddOrder(market); err != nil {
		t.Error(err)
	}

	if have := b.asks.TotalQuantity(decimal.NewFromInt(10_000)); !have.Equal(decimal.NewFromInt(1)) {
		t.Errorf("have %v, want 1", have)
	}
}
