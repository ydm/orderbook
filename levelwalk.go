package orderbook

import (
	"container/heap"
	"fmt"

	"github.com/shopspring/decimal"
)

type Item struct {
	index    int
	priority decimal.Decimal
}

func (i *Item) String() string {
	return fmt.Sprintf("{index=%d priority=%d}", i.index, i.priority)
}

type PriorityQueue []Item

func NewPriorityQueue(n int) PriorityQueue {
	pq := make(PriorityQueue, 0, n)

	return pq
}

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].priority.LessThan(pq[j].priority)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item, ok := x.(Item)
	if !ok {
		panic("")
	}

	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]

	return item
}

// Walk in order over Levels in a LevelHeap.
func Walk(levels LevelHeap, each func(level *Level) bool) {
	const (
		left  = 1
		right = 2
	)

	indices := NewPriorityQueue(levels.Len()/2 + 1)
	push := func(i int) {
		if i < levels.Len() {
			heap.Push(&indices, Item{
				index:    i,
				priority: levels[i].Key(),
			})
		}
	}

	push(0)

	for indices.Len() > 0 {
		item, ok := heap.Pop(&indices).(Item)

		if !ok {
			continue
		}

		index := item.index

		if index >= len(levels) {
			continue
		}

		push(2*index + left)  // Left child.
		push(2*index + right) // Right child.

		if !each(levels[index]) {
			break
		}
	}
}
