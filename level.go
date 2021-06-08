package orderbook

import (
	"container/heap"
	"fmt"
	"orderbook/pkg/decimal"
	"sync"
)

// +------------+
// | OrderQueue |
// +------------+

// OrderQueue holds all the orders at a particular level of the order book.  It keeps them
// in a queue (FIFO) and also allows quick access using an ID.
type OrderQueue struct {
	queue   []Order
	indices map[string]int
	mu      sync.Mutex
}

func NewOrderQueue(n int) OrderQueue {
	return OrderQueue{
		queue:   make([]Order, 0, n),
		indices: make(map[string]int),
	}
}

func (q *OrderQueue) Add(o Order) {
	q.mu.Lock()
	defer q.mu.Unlock()

	_, ok := q.indices[o.ID]
	if ok {
		// There is already an order with this ID.
		return
	}
	q.indices[o.ID] = len(q.queue)
	q.queue = append(q.queue, o)
}

func (q *OrderQueue) Remove() Order {
	q.mu.Lock()
	defer q.mu.Unlock()

	item := q.queue[0]
	q.queue = q.queue[1:]
	delete(q.indices, item.ID)
	return item
}

func (q *OrderQueue) RemoveByID(orderID string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	index, ok := q.indices[orderID]
	if ok {
		copy(q.queue[index:], q.queue[index+1:])
		q.queue = q.queue[:len(q.queue)-1]
		delete(q.indices, orderID)
	}
	return ok
}

func (q *OrderQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) != len(q.indices) {
		fmt.Printf("%d %d\n", len(q.queue), len(q.indices))
		panic("invariant")
	}
	return len(q.queue)
}

// +-------+
// | Level |
// +-------+

// Level represents a level in the order book (either ask or bid).  It
// has a price and a queue of limit orders waiting to get executed.
type Level struct {
	Price  decimal.Decimal
	Orders OrderQueue
	key    int64
}

func NewLevelAsk(price decimal.Decimal) Level {
	return Level{
		Price:  price,
		Orders: NewOrderQueue(0),
		key:    price.Raw(),
	}
}

func NewLevelBid(price decimal.Decimal) Level {
	return Level{
		Price:  price,
		Orders: NewOrderQueue(0),
		key:    -price.Raw(),
	}
}

// +-----------+
// | LevelHeap |
// +-----------+

// LevelHeap keeps Levels ordered.
type LevelHeap []*Level

func NewLevelHeap(n int) LevelHeap {
	xs := make(LevelHeap, 0, n)
	heap.Init(&xs)
	return xs
}

func (h LevelHeap) Len() int { return len(h) }

func (h LevelHeap) Less(i, j int) bool {
	return h[i].Price.LessThan(h[j].Price)
}

func (h LevelHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *LevelHeap) Push(p interface{}) {
	level := p.(*Level)
	*h = append(*h, level)
}

func (h *LevelHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

func (h *LevelHeap) AddOrder(price decimal.Decimal, order Order) {
}

func (h *LevelHeap) Iterate(f func(level *Level) bool) {
}
