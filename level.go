package orderbook

import (
	"container/heap"
	"orderbook/pkg/decimal"
)

// +------------+
// | OrderQueue |
// +------------+

// OrderQueue holds all the orders at a particular level of the order book.  It keeps them
// in a queue (FIFO) and also allows quick access using an ID.
type OrderQueue []Order

func NewOrderQueue(n int) OrderQueue {
	return make([]Order, 0, n)
}

func (xs *OrderQueue) Add(o Order) {
	*xs = append(*xs, o)
}

func (xs *OrderQueue) Remove() Order {
	item := (*xs)[0]
	*xs = (*xs)[1:]
	return item
}

func (xs *OrderQueue) Len() int {
	return len(*xs)
}

// +-------+
// | Level |
// +-------+

const (
	Ask = iota
	Bid
)

// Level represents a level in the order book (either ask or bid).  It
// has a price and a queue of limit orders waiting to get executed.
type Level struct {
	Price  decimal.Decimal // Also serves as key in the heap.
	Orders OrderQueue      // All of the orders on this level.
	Type   int             // Ask or Bid.
}

func NewLevelAsk(price decimal.Decimal) Level {
	return Level{
		Price:  price,
		Orders: NewOrderQueue(0),
		Type:   Ask,
	}
}

func NewLevelBid(price decimal.Decimal) Level {
	return Level{
		Price:  price,
		Orders: NewOrderQueue(0),
		Type:   Bid,
	}
}

func (v *Level) Less(rhs *Level) bool {
	switch v.Type {
	case Ask:
		return v.Price.LessThan(rhs.Price)
	case Bid:
		return v.Price.GreaterThan(rhs.Price)
	default:
		panic("illegal type")
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
	return h[i].Less(h[j])
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
