package orderbook

import (
	"container/heap"
	"testing"

	"orderbook/pkg/decimal"
)

func TestLevelHeap_Walk_1(t *testing.T) {
	xs := NewLevelHeap(16)
	heap.Push(&xs, NewLevel(decimal.NewFromStringPanic("4"), Ask))
	heap.Push(&xs, NewLevel(decimal.NewFromStringPanic("2"), Ask))
	heap.Push(&xs, NewLevel(decimal.NewFromStringPanic("5"), Ask))
	heap.Push(&xs, NewLevel(decimal.NewFromStringPanic("1"), Ask))
	heap.Push(&xs, NewLevel(decimal.NewFromStringPanic("3"), Ask))

	expected := []string{"1.0", "2.0", "3.0", "4.0", "5.0"}
	index := 0
	xs.Walk(func(level *Level) bool {
		t.Helper()
		if level.Price.String() != expected[index] {
			t.Errorf("have %v, want %v", level.Price.String(), expected[index])
		}
		index++
		return true
	})
}

func TestLevelHeap_Walk_2(t *testing.T) {
	xs := NewLevelHeap(16)
	heap.Push(&xs, NewLevel(decimal.NewFromStringPanic("4"), Bid))
	heap.Push(&xs, NewLevel(decimal.NewFromStringPanic("2"), Bid))
	heap.Push(&xs, NewLevel(decimal.NewFromStringPanic("5"), Bid))
	heap.Push(&xs, NewLevel(decimal.NewFromStringPanic("1"), Bid))
	heap.Push(&xs, NewLevel(decimal.NewFromStringPanic("3"), Bid))

	expected := []string{"5.0", "4.0", "3.0", "2.0", "1.0"}
	index := 0
	xs.Walk(func(level *Level) bool {
		t.Helper()
		if level.Price.String() != expected[index] {
			t.Errorf("have %v, want %v", level.Price.String(), expected[index])
		}
		index++
		return true
	})
}
