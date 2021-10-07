package orderbook

import (
	"errors"
	"fmt"
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

	// Make sure the same order cannot be submitted twice.
	if err := b.AddOrder(limit); !errors.Is(err, ErrOrderExists) {
		t.Error()
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
	assertLevels(t, &b.asks, pq{"10000", "1"})

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
	assertLevels(t, &b.bids, pq{"10000", "2"})

	assertExecutedQuantities(t, b, iq{"one", "1"}, iq{"two", "1"})
}

func TestBook_CancelOrder_1(t *testing.T) {
	b := NewBook()
	if err := b.CancelOrder(""); !errors.Is(err, ErrInvalidID) {
		t.Error()
	}
	if err := b.CancelOrder("market"); !errors.Is(err, ErrOrderDoesNotExist) {
		t.Error()
	}
	market := ClientOrder{
		Side:             SideBuy,
		OriginalQuantity: decimal.NewFromInt(3),
		ExecutedQuantity: decimal.Zero,
		ID:               "market",
		Type:             TypeMarket,
	}
	if err := b.AddOrder(market); !errors.Is(err, ErrMarketNotFullyExecuted) {
		t.Error()
	}
	if err := b.CancelOrder("market"); !errors.Is(err, ErrCannotCancelMarketOrder) {
		t.Error()
	}
}

// Cancel an unexecuted limit order.
func TestBook_CancelOrder_2(t *testing.T) {
	b := NewBook()
	limit := ClientOrder{
		Side:             SideBuy,
		OriginalQuantity: decimal.NewFromInt(3),
		ExecutedQuantity: decimal.Zero,
		Price:            decimal.NewFromInt(10_000),
		ID:               "limit",
		Type:             TypeLimit,
	}
	b.AddOrder(limit)
	assertCountLevels(t, b, 0, 1)
	if err := b.CancelOrder("limit"); err != nil {
		t.Error(err)
	}
	assertCountLevels(t, b, 0, 0)
}

// Cancel an executed order.
func TestBook_CancelOrder_3(t *testing.T) {
	b := NewBook()
	limit := ClientOrder{
		Side:             SideSell,
		OriginalQuantity: decimal.NewFromInt(1),
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
	b.AddOrder(limit)
	assertCountLevels(t, b, 1, 0)
	b.AddOrder(market)
	assertCountLevels(t, b, 0, 0)
	if err := b.CancelOrder("limit"); !errors.Is(err, ErrCannotCancelOrder) {
		t.Error(err)
	}
}

func TestBook_GetSnapshot(t *testing.T) {
	b := NewBook()
	for price := 11; price <= 30; price++ {
		for i := 0; i < price; i++ {
			order := ClientOrder{
				Side:             SideBuy,
				OriginalQuantity: decimal.NewFromInt(int64(2 * price)),
				ExecutedQuantity: decimal.Zero,
				Price:            decimal.NewFromInt(int64(price)),
				ID:               fmt.Sprintf("%d_%d", price, i),
				Type:             TypeLimit,
			}
			if price >= 21 {
				order.Side = SideSell
			}
			b.AddOrder(order)
		}
	}
	var snapshot Snapshot

	snapshot = b.GetSnapshot(0)
	if len(snapshot.Asks) != 0 || len(snapshot.Bids) != 0 {
		t.Error()
	}

	snapshot = b.GetSnapshot(5)
	if len(snapshot.Asks) != 5 || len(snapshot.Bids) != 5 {
		t.Errorf("have %d and %d, want 5", len(snapshot.Asks), len(snapshot.Bids))
	}

	snapshot = b.GetSnapshot(20)
	if len(snapshot.Asks) != 10 || len(snapshot.Bids) != 10 {
		t.Error()
	}

	assertEq := func(level ClientLevel, price int64) {
		t.Helper()
		if !level.Price.Equal(decimal.NewFromInt(price)) {
			t.Errorf("have price %v, want price %d", level.Price, price)
		}
		if !level.Quantity.Equal(decimal.NewFromInt(2 * price * price)) {
			t.Error()
		}
	}

	assertEq(snapshot.Asks[9], 30)
	assertEq(snapshot.Asks[8], 29)
	assertEq(snapshot.Asks[7], 28)
	assertEq(snapshot.Asks[6], 27)
	assertEq(snapshot.Asks[5], 26)
	assertEq(snapshot.Asks[4], 25)
	assertEq(snapshot.Asks[3], 24)
	assertEq(snapshot.Asks[2], 23)
	assertEq(snapshot.Asks[1], 22)
	assertEq(snapshot.Asks[0], 21)

	assertEq(snapshot.Bids[0], 20)
	assertEq(snapshot.Bids[1], 19)
	assertEq(snapshot.Bids[2], 18)
	assertEq(snapshot.Bids[3], 17)
	assertEq(snapshot.Bids[4], 16)
	assertEq(snapshot.Bids[5], 15)
	assertEq(snapshot.Bids[6], 14)
	assertEq(snapshot.Bids[7], 13)
	assertEq(snapshot.Bids[8], 12)
	assertEq(snapshot.Bids[9], 11)
}
