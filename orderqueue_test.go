package orderbook

import (
	"strconv"
	"testing"
)

func TestBinarySearch(t *testing.T) {
	assertEq := func(have, want int) {
		t.Helper()
		if have != want {
			t.Errorf("have %d, want %d", have, want)
		}
	}
	xs := []int{1, 3, 6, 10, 15, 21, 28, 36, 45, 55}
	ys := make([]*Order, len(xs))
	for i, x := range xs {
		order := NewOrder(strconv.Itoa(i), newDecimalPanic("1"))
		ys[i] = &order
		ys[i].insertionIndex = x
	}
	assertEq(binarySearch(ys, 1), 0)
	assertEq(binarySearch(ys, 3), 1)
	assertEq(binarySearch(ys, 10), 3)
	assertEq(binarySearch(ys, 45), 8)
	assertEq(binarySearch(ys, 55), 9)
	assertEq(binarySearch(ys, -1), -1)
	assertEq(binarySearch(ys, 2), -2)
	assertEq(binarySearch(ys, 50), -10)
	assertEq(binarySearch(ys, 60), -11)
}

func TestOrderQueue(t *testing.T) {
	q := NewOrderQueue(2)
	if q.Len() != 0 {
		t.Fail()
	}

	inp := Order{
		Quantity: newDecimalPanic("1"),
		ID:       "7bfa0e20",
	}
	q.Add(inp)
	if q.Len() != 1 {
		t.Errorf("have %d, want 1", q.Len())
	}

	out := q.Remove()
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
	const N = 1000

	q := NewOrderQueue(8)
	for i := 0; i < N; i++ {
		s := strconv.Itoa(i)
		o := Order{
			Quantity: newDecimalPanic(s),
			ID:       s,
		}
		q.Add(o)
	}

	for i := 0; i < N; i++ {
		if q.Len() != (N - i) {
			t.Errorf("have %d, want %d", q.Len(), N-i)
		}
		s := strconv.Itoa(i)
		popped := q.Remove()
		if popped.ID != s || !popped.Quantity.Equal(newDecimalPanic(s)) {
			t.Errorf("have={%s %v}, want={%s, %s}", popped.ID, popped.Quantity, s, s)
		}
		if q.Len() != (N - i - 1) {
			t.Errorf("have %d, want %d", q.Len(), N-i-1)
		}
	}
}

func TestOrderQueue_RemoveByID(t *testing.T) {
	q := NewOrderQueue(2)

	for i := 0; i < 1000; i++ {
		s := strconv.Itoa(i)
		o := Order{
			Quantity: newDecimalPanic(s),
			ID:       s,
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
