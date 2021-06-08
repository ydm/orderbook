package orderbook

import (
	"container/heap"
	"orderbook/pkg/decimal"
)

// +------------+
// | OrderQueue |
// +------------+

// OrderQueue holds all the orders at a particular level of the order book.  It keeps them
// in a queue (FIFO), so orders of the same price level get executed in the order they
// were submitted.
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
	Type   int             // Ask or Bid.  TODO: Remove!
}

func NewLevel(price decimal.Decimal, levelType int) *Level {
	return &Level{
		Price:  price,
		Orders: NewOrderQueue(16),
		Type:   levelType,
	}
}

func (v *Level) Key() int64 {
	switch v.Type {
	case Ask:
		return v.Price.Raw()
	case Bid:
		return -v.Price.Raw()
	default:
		panic("illegal type")
	}
}

func (v *Level) Less(rhs *Level) bool {
	return v.Key() < rhs.Key()
}

// +-----------+
// | LevelHeap |
// +-----------+

// LevelHeap lets us iterate Levels ordered by price.
type LevelHeap []*Level

func NewLevelHeap(n int) LevelHeap {
	xs := make([]*Level, 0, n)
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
	n := len(*h)
	item := (*h)[n-1]
	*h = (*h)[:n-1]
	return item
}

func (h *LevelHeap) Walk(f func(level *Level) bool) {
	Walk(*h, f)
}

// +----------+
// | LevelMap |
// +----------+

// LevelMap maps Price to Level.
type LevelMap map[decimal.Decimal]*Level

// +--------+
// | Ladder |
// +--------+

// Ladder keeps all price levels and their respective orders, allows
// iteration and querying by price.
type Ladder struct {
	Map  LevelMap
	Heap LevelHeap
	Type int // Ask or Bid.
}

func NewLadder() Ladder {
	return Ladder{
		Map:  make(LevelMap),
		Heap: make(LevelHeap, 0, 256),
	}
}

func (d *Ladder) AddOrder(price decimal.Decimal, o Order) {
	level, ok := d.Map[price]
	if ok {
		level.Orders.Add(o)
	}
	level = NewLevel(price, d.Type)
	d.Map[price] = level
	heap.Push(&d.Heap, level)
}

func (d *Ladder) Walk(f func(level *Level) bool) {
	d.Heap.Walk(f)
}

// func (h *LevelHeap) AddOrder(price decimal.Decimal, o Order) {
// 	level, ok := h.Find(price)
// 	if ok {
// 		level.Orders.Add(o)
// 	} else {
// 		Level{
// 		}
// 	}

// 	if h.Contains()
// 	h.Walk(func(level Level) bool {

// 	})
// }

// func (h *LevelHeap) Find(price decimal.Decimal) (level *Level, ok bool) {
// 	h.Walk(func(x Level) bool {
// 		if x.Price.Equal(price) {
// 			level = x
// 			ok = true
// 			return false
// 		}
// 		return true
// 	})
// 	return
// }
