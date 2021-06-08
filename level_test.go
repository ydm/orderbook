package orderbook

import (
	"strconv"
	"testing"

	"orderbook/pkg/decimal"
)

func TestOrderQueue(t *testing.T) {
	q := NewOrderQueue(2)
	if q.Len() != 0 {
		t.Fail()
	}

	inp := Order{
		Quantity: decimal.NewFromStringPanic("1"),
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

	// Make sure size grows.
	for i := 1; i <= 16; i++ {
		q.Add(inp)
		if q.Len() != i {
			t.Errorf("have %d, want %d", q.Len(), i)
		}
	}
}

func TestOrderQueue_Remove(t *testing.T) {
	const N = 1000

	q := NewOrderQueue(8)
	for i := 0; i < N; i++ {
		s := strconv.Itoa(i)
		o := Order{
			Quantity: decimal.NewFromStringPanic(s),
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
		if popped.ID != s || popped.Quantity != decimal.NewFromStringPanic(s) {
			t.Errorf("have={%s %v}, want={%s, %s}", popped.ID, popped.Quantity, s, s)
		}
		if q.Len() != (N - i - 1) {
			t.Errorf("have %d, want %d", q.Len(), N-i-1)
		}
	}
}

// func TestOrderQueue_RemoveByID(t *testing.T) {
// 	q := NewOrderQueue(2)

// 	for i := 0; i < 1000; i++ {
// 		s := strconv.Itoa(i)
// 		o := Order{
// 			Quantity: decimal.NewFromStringPanic(s),
// 			ID:       s,
// 		}
// 		q.Add(o)
// 	}

// 	if q.Len() != 1000 {
// 		t.Fail()
// 	}
// 	if q.RemoveByID("nonexistent") {
// 		t.Fail()
// 	}
// 	if !q.RemoveByID("681") {
// 		t.Fail()
// 	}
// 	if q.RemoveByID("681") {
// 		t.Fail()
// 	}
// 	if q.Len() != 999 {
// 		t.Fail()
// 	}

// 	for q.Len() > 0 {
// 		order := q.Remove()
// 		if order.ID == "681" {
// 			t.Error()
// 		}
// 	}
// }

func TestLevelHeap(t *testing.T) {

}
