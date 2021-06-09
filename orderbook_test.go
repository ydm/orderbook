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

type pq struct{ price, quantity string }

func assertLevels(t *testing.T, ladder *Ladder, xs ...pq) {
	t.Helper()

	ladder.Walk(func(level *Level) bool {
		t.Helper()

		if len(xs) <= 0 {
			t.Errorf("unexpected level at price %v", level.Price)
		}
		x := xs[0]
		xs = xs[1:]

		price, err := decimal.NewFromString(x.price)
		if err != nil {
			panic(err)
		}
		quantity, err := decimal.NewFromString(x.quantity)
		if err != nil {
			panic(err)
		}

		if !level.Price.Equal(price) {
			t.Errorf("have %v, want %v", level.Price, price)
		}
		if have := level.TotalQuantity(); !have.Equal(quantity) {
			t.Errorf("have %v, want %v", have, quantity)
		}

		return true
	})
}

type iq struct{ id, quantity string }

func assertExecutedQuantities(t *testing.T, b *Book, xs ...iq) {
	t.Helper()

	for _, x := range xs {
		order, err := b.GetOrder(x.id)
		if err != nil {
			t.Error(err)
		}

		quantity, err := decimal.NewFromString(x.quantity)
		if err != nil {
			panic(err)
		}
		if !order.ExecutedQuantity.Equal(quantity) {
			t.Errorf("order=%s: have executed quantity %v, want executed quantity %v",
				order.ID, order.ExecutedQuantity, quantity)
		}
	}
}

func assertQuantities(t *testing.T, ladder *Ladder, xs ...int64) {
	t.Helper()

	ladder.Walk(func(level *Level) bool {
		t.Helper()

		if len(xs) <= 0 {
			t.Errorf("unexpected level at price %v", level.Price)
		}
		x := xs[0]
		xs = xs[1:]

		if !level.Price.Equal(decimal.NewFromInt(x)) {
			t.Errorf("have %v, want %d", level.Price, x)
		}

		return true
	})
}

// Submit a market order against an empty order book.
func TestBook_AddOrder_1(t *testing.T) {
	b := NewBook()
	assertCountLevels(t, b, 0, 0)

	order := ClientOrder{
		Side:             SideBuy,
		OriginalQuantity: decimal.NewFromInt(1),
		ExecutedQuantity: decimal.Zero,
		ID:               "id1",
		Type:             TypeMarket,
	}

	// Make sure market orders do not end up in the order book, but rather get matched
	// against what's in the book.
	err := b.AddOrder(order)
	assertCountLevels(t, b, 0, 0)

	// Since the book is empty, the error returned should notify of incomplete
	// execution.
	if !errors.Is(err, ErrMarketNotFullyExecuted) {
		t.Error()
	}

	// All orders are kept by their ID in the so called database.  For this particular
	// order the executed quantity should be 0.
	assertExecutedQuantities(t, b, iq{"id1", "0"})
}

// Submit a market order and match it against a limit order from the order book.  The
// limit order covers the market order and thus the later gets executed immediately.  The
// limit order from the book is partially executed.
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

	// Make sure limit orders get added to the order book.
	if err := b.AddOrder(limit); err != nil {
		t.Error(err)
	}
	assertCountLevels(t, b, 1, 0)
	assertLevels(t, &b.asks, pq{"10000", "2"})

	// Make sure this market gets matched and what's left in the order book is the
	// partially executed limit order.
	if err := b.AddOrder(market); err != nil {
		t.Error(err)
	}
	assertCountLevels(t, b, 1, 0)
	assertLevels(t, &b.asks, pq{"10000", "1"})
}

// Submit a market order and match it against a limit order from the order book.  The
// market order covers the limit order.
func TestBook_AddOrder_3(t *testing.T) {
	b := NewBook()
	limit := ClientOrder{
		Side:             SideSell,
		OriginalQuantity: decimal.NewFromInt(1),
		ExecutedQuantity: decimal.Zero,
		Price:            decimal.NewFromInt(10_000),
		ID:               "limit",
		Type:             TypeLimit,
	}
	b.AddOrder(limit)

	market := ClientOrder{
		Side:             SideBuy,
		OriginalQuantity: decimal.NewFromInt(3),
		ExecutedQuantity: decimal.Zero,
		ID:               "market",
		Type:             TypeMarket,
	}
	err := b.AddOrder(market)

	// Make sure the order book is now empty.
	assertCountLevels(t, b, 0, 0)

	// Make sure the market order didn't execute fully.
	if !errors.Is(err, ErrMarketNotFullyExecuted) {
		t.Error()
	}

	// Check the database record for this order exists and the executed quantity is
	// properly set to 1.
	assertExecutedQuantities(t, b, iq{"market", "1"})
}

// Add two limit orders that do not touch each other's prices.
func TestBook_AddOrder_4(t *testing.T) {
	b := NewBook()
	sell := ClientOrder{
		Side:             SideSell,
		OriginalQuantity: decimal.NewFromInt(1),
		ExecutedQuantity: decimal.Zero,
		Price:            decimal.NewFromInt(10_001),
		ID:               "one",
		Type:             TypeLimit,
	}
	if err := b.AddOrder(sell); err != nil {
		t.Error(err)
	}
	assertCountLevels(t, b, 1, 0)

	buy := ClientOrder{
		Side:             SideBuy,
		OriginalQuantity: decimal.NewFromInt(3),
		ExecutedQuantity: decimal.Zero,
		Price:            decimal.NewFromInt(10_000),
		ID:               "two",
		Type:             TypeLimit,
	}
	if err := b.AddOrder(buy); err != nil {
		t.Error(err)
	}
	assertCountLevels(t, b, 1, 1)
	assertExecutedQuantities(t, b, iq{"one", "0"}, iq{"two", "0"})
}

// Match a limit order with another limit order.
func TestBook_AddOrder_5(t *testing.T) {
	b := NewBook()
	sell := ClientOrder{
		Side:             SideSell,
		OriginalQuantity: decimal.NewFromInt(1),
		ExecutedQuantity: decimal.Zero,
		Price:            decimal.NewFromInt(10_000),
		ID:               "one",
		Type:             TypeLimit,
	}
	if err := b.AddOrder(sell); err != nil {
		t.Error(err)
	}
	assertCountLevels(t, b, 1, 0)

	buy := ClientOrder{
		Side:             SideBuy,
		OriginalQuantity: decimal.NewFromInt(3),
		ExecutedQuantity: decimal.Zero,
		Price:            decimal.NewFromInt(10_000),
		ID:               "two",
		Type:             TypeLimit,
	}
	if err := b.AddOrder(buy); err != nil {
		t.Error(err)
	}
	assertCountLevels(t, b, 0, 1)
	assertExecutedQuantities(t, b, iq{"one", "1"}, iq{"two", "1"})
}
