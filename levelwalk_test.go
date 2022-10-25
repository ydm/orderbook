package orderbook_test

import (
	"container/heap"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/ydm/orderbook"
)

func TestLevelHeap_Walk(t *testing.T) {
	t.Parallel()

	check := func(levelType int, expected []string) {
		levels := orderbook.NewLevelHeap(16)
		heap.Push(&levels, orderbook.NewLevel(decimal.NewFromInt(4), levelType))
		heap.Push(&levels, orderbook.NewLevel(decimal.NewFromInt(2), levelType))
		heap.Push(&levels, orderbook.NewLevel(decimal.NewFromInt(5), levelType))
		heap.Push(&levels, orderbook.NewLevel(decimal.NewFromInt(1), levelType))
		heap.Push(&levels, orderbook.NewLevel(decimal.NewFromInt(3), levelType))

		index := 0

		levels.Walk(func(level *orderbook.Level) bool {
			t.Helper()

			if level.Price.String() != expected[index] {
				t.Errorf("have %v, want %s", level.Price, expected[index])
			}
			index++

			return true
		})
	}

	check(orderbook.Ask, []string{"1", "2", "3", "4", "5"})
	check(orderbook.Bid, []string{"5", "4", "3", "2", "1"})
}
