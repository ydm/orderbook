package orderbook_test

import (
	"strconv"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/ydm/orderbook"
)

func TestBinarySearch(t *testing.T) {
	t.Parallel()

	assertEq := func(have, want int) {
		t.Helper()

		if have != want {
			t.Errorf("have %d, want %d", have, want)
		}
	}

	xs := []int{1, 3, 6, 10, 15, 21, 28, 36, 45, 55}
	ys := make([]*orderbook.Order, len(xs))

	for i, x := range xs {
		order := orderbook.NewOrder(strconv.Itoa(i), decimal.NewFromInt(1))
		ys[i] = &order
		ys[i].InsertionIndex = x
	}

	assertEq(orderbook.BinarySearch(ys, 1), 0)
	assertEq(orderbook.BinarySearch(ys, 3), 1)
	assertEq(orderbook.BinarySearch(ys, 10), 3)
	assertEq(orderbook.BinarySearch(ys, 45), 8)
	assertEq(orderbook.BinarySearch(ys, 55), 9)
	assertEq(orderbook.BinarySearch(ys, -1), -1)
	assertEq(orderbook.BinarySearch(ys, 2), -2)
	assertEq(orderbook.BinarySearch(ys, 50), -10)
	assertEq(orderbook.BinarySearch(ys, 60), -11)
}

func TestOrderQueue(t *testing.T) {
	t.Parallel()

	q := orderbook.NewOrderQueue(2)
	if q.Len() != 0 {
		t.Fail()
	}

	inp := orderbook.Order{
		ID:             "7bfa0e20",
		Quantity:       decimal.NewFromInt(1),
		InsertionIndex: 0,
	}

	q.Add(inp)

	if q.Len() != 1 {
		t.Errorf("have %d, want 1", q.Len())
	}

	out := q.Remove() //nolint:ifshort

	if q.Len() != 0 {
		t.Errorf("have %d, want 0", q.Len())
	}

	if inp.ID != out.ID || !inp.Quantity.Equal(out.Quantity) {
		t.Errorf("have %v, want %v", out, inp)
	}

	// Make sure an order with the same ID cannot be submitted more than once.
	for i := 1; i <= 16; i++ {
		q.Add(inp)

		if q.Len() != 1 {
			t.Errorf("have %d, want 1", q.Len())
		}
	}
}

func TestOrderQueue_Remove(t *testing.T) {
	t.Parallel()

	const N = 1000

	q := orderbook.NewOrderQueue(8)

	for i := 0; i < N; i++ {
		s := strconv.Itoa(i)
		o := orderbook.Order{
			ID:             s,
			Quantity:       decimal.NewFromInt(int64(i)),
			InsertionIndex: 0,
		}
		q.Add(o)
	}

	for i := 0; i < N; i++ {
		if q.Len() != (N - i) {
			t.Errorf("have %d, want %d", q.Len(), N-i)
		}

		popped := q.Remove()
		wantedID := strconv.Itoa(i)
		wantedQuantity := decimal.NewFromInt(int64(i))

		if popped.ID != wantedID || !popped.Quantity.Equal(wantedQuantity) {
			t.Errorf("have={%s %v}, want={%s, %s}", popped.ID, popped.Quantity, wantedID, wantedID)
		}

		if q.Len() != (N - i - 1) {
			t.Errorf("have %d, want %d", q.Len(), N-i-1)
		}
	}
}

func TestOrderQueue_RemoveByID(t *testing.T) {
	t.Parallel()

	q := orderbook.NewOrderQueue(2)

	for i := 0; i < 1000; i++ {
		s := strconv.Itoa(i)
		o := orderbook.Order{
			ID:             s,
			Quantity:       decimal.NewFromInt(int64(i)),
			InsertionIndex: 0,
		}
		q.Add(o)
	}

	if q.Len() != 1000 {
		t.Fail()
	}

	if q.RemoveByID("nonexistent") {
		t.Fail()
	}

	if !q.RemoveByID("681") {
		t.Fail()
	}

	if q.RemoveByID("681") {
		t.Fail()
	}

	if q.Len() != 999 {
		t.Fail()
	}

	for q.Len() > 0 {
		order := q.Remove()
		if order.ID == "681" {
			t.Error()
		}
	}
}
