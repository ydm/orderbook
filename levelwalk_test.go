package orderbook_test

import (
	"container/heap"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/ydm/orderbook"
)

func TestLevelHeap_Walk(t *testing.T) {
	t.Parallel()

	f := func(levelType int, expected []string) {
		xs := orderbook.NewLevelHeap(16)
		heap.Push(&xs, orderbook.NewLevel(decimal.NewFromInt(4), levelType))
		heap.Push(&xs, orderbook.NewLevel(decimal.NewFromInt(2), levelType))
		heap.Push(&xs, orderbook.NewLevel(decimal.NewFromInt(5), levelType))
		heap.Push(&xs, orderbook.NewLevel(decimal.NewFromInt(1), levelType))
		heap.Push(&xs, orderbook.NewLevel(decimal.NewFromInt(3), levelType))

		index := 0
		xs.Walk(func(level *orderbook.Level) bool {
			t.Helper()
			if level.Price.String() != expected[index] {
				t.Errorf("have %v, want %s", level.Price, expected[index])
			}
			index++
			return true
		})
	}

	f(orderbook.Ask, []string{"1", "2", "3", "4", "5"})
	f(orderbook.Bid, []string{"5", "4", "3", "2", "1"})
}
