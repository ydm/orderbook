package orderbook

import (
	"container/heap"
	"testing"

	"github.com/shopspring/decimal"
)

func TestLevelHeap_Walk_1(t *testing.T) {
	xs := NewLevelHeap(16)
	heap.Push(&xs, NewLevel(decimal.NewFromInt(4), Ask))
	heap.Push(&xs, NewLevel(decimal.NewFromInt(2), Ask))
	heap.Push(&xs, NewLevel(decimal.NewFromInt(5), Ask))
	heap.Push(&xs, NewLevel(decimal.NewFromInt(1), Ask))
	heap.Push(&xs, NewLevel(decimal.NewFromInt(3), Ask))

	expected := []string{"1", "2", "3", "4", "5"}
	index := 0
	xs.Walk(func(level *Level) bool {
		t.Helper()
		if level.Price.String() != expected[index] {
			t.Errorf("have %v, want %s", level.Price, expected[index])
		}
		index++
		return true
	})
}

func TestLevelHeap_Walk_2(t *testing.T) {
	xs := NewLevelHeap(16)
	heap.Push(&xs, NewLevel(decimal.NewFromInt(4), Bid))
	heap.Push(&xs, NewLevel(decimal.NewFromInt(2), Bid))
	heap.Push(&xs, NewLevel(decimal.NewFromInt(5), Bid))
	heap.Push(&xs, NewLevel(decimal.NewFromInt(1), Bid))
	heap.Push(&xs, NewLevel(decimal.NewFromInt(3), Bid))

	expected := []string{"5", "4", "3", "2", "1"}
	index := 0
	xs.Walk(func(level *Level) bool {
		t.Helper()
		if level.Price.String() != expected[index] {
			t.Errorf("have %v, want %s", level.Price, expected[index])
		}
		index++
		return true
	})
}
