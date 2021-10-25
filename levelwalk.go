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
func Walk(h LevelHeap, f func(level *Level) bool) {
	indices := NewPriorityQueue(h.Len()/2 + 1)
	push := func(i int) {
		if i < h.Len() {
			heap.Push(&indices, Item{
				index:    i,
				priority: h[i].Key(),
			})
		}
	}

	push(0)

	for indices.Len() > 0 {
		index := heap.Pop(&indices).(Item).index

		if index >= len(h) {
			continue
		}

		push(2*index + 1) // Left child.
		push(2*index + 2) // Right child.

		if !f(h[index]) {
			break
		}
	}
}
