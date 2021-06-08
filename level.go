package orderbook

import (
	"container/heap"
	"fmt"
	"orderbook/pkg/decimal"
)

// +------------+
// | OrderQueue |
// +------------+

// binarySearch returns index of the search key, if it is contained in
// the array, otherwise (-(insertion point) – 1).
func binarySearch(xs []Order, x int) int {
	i, j := 0, len(xs)
	for i < j {
		h := i + (j-i)/2
		if xs[h].InsertionIndex < x {
			i = h + 1
		} else {
			j = h
		}
	}
	if i < len(xs) && xs[i].InsertionIndex == x {
		return i
	} else {
		return -i - 1
	}
}

// OrderQueue holds all the orders at a particular level of the order book.  It keeps them
// in a queue (FIFO), so orders of the same price level get executed in the order they
// were submitted.  OrderQueue also allows querying using the order ID.
type OrderQueue struct {
	queue []Order

	// indices maps an order ID to its insertion order index.
	// This way removing by ID takes O(2logN) + copy().
	indices map[string]int

	next int
}

func NewOrderQueue(n int) OrderQueue {
	return OrderQueue{
		queue:   make([]Order, 0, n),
		indices: make(map[string]int),
	}
}

func (q *OrderQueue) Add(order Order) bool {
	_, ok := q.indices[order.ID]
	if ok {
		// There is already an order with this ID.
		return false
	}
	order.InsertionIndex = q.next
	q.queue = append(q.queue, order)
	q.indices[order.ID] = q.next
	q.next++
	return true
}

func (q *OrderQueue) Remove() Order {
	order := q.queue[0]
	q.queue = q.queue[1:]
	delete(q.indices, order.ID)
	return order
}

func (q *OrderQueue) RemoveByID(orderID string) bool {
	insertionIndex, ok := q.indices[orderID]
	if ok {
		i := binarySearch(q.queue, insertionIndex)
		if i >= 0 {
			order := q.queue[i]

			// Delete order from queue.
			copy(q.queue[i:], q.queue[i+1:])
			q.queue = q.queue[:len(q.queue)-1]

			// Delete order from map.
			delete(q.indices, order.ID)
			return true
		}
	}
	return false
}

func (q *OrderQueue) Len() int {
	if len(q.queue) != len(q.indices) {
		fmt.Printf("%d %d\n", len(q.queue), len(q.indices))
		panic("invariant")
	}
	return len(q.queue)
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
	index  int             // Heap index.
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
	h[i].index = i
	h[j].index = j
}

func (h *LevelHeap) Push(p interface{}) {
	level := p.(*Level)
	level.index = len(*h)
	*h = append(*h, level)
}

func (h *LevelHeap) Pop() interface{} {
	n := len(*h)
	level := (*h)[n-1]
	level.index = -1
	*h = (*h)[:n-1]
	return level
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
// iteration and querying by price or order ID.  Ladder is either of
// type Ask or Bid.
type Ladder struct {
	mapping LevelMap  // Maps price to level.
	heap    LevelHeap // Holds all levels in a convenient container.
	Type    int       // Ask or Bid.
}

func NewLadder(ladderType int) Ladder {
	return Ladder{
		mapping: make(LevelMap),
		heap:    make(LevelHeap, 0, 256),
		Type:    ladderType,
	}
}

func (d *Ladder) AddOrder(price decimal.Decimal, o Order) bool {
	// First check if level exists.
	level, ok := d.mapping[price]
	if ok {
		return level.Orders.Add(o)
	}

	// Level does not exist, create it and add the order.
	level = NewLevel(price, d.Type)
	if !level.Orders.Add(o) {
		panic("illegal state")
	}

	// Save the newly made level into our data structures.
	d.mapping[price] = level
	heap.Push(&d.heap, level)
	return true
}

func (d *Ladder) RemoveOrder(price decimal.Decimal, ID string) bool {
	level, ok := d.mapping[price]
	if ok {
		ans := level.Orders.RemoveByID(ID)
		if level.Orders.Len() <= 0 {
			heap.Remove(&d.heap, level.index)
			delete(d.mapping, price)
		}
		return ans
	}
	return false
}

func (d *Ladder) Walk(f func(level *Level) bool) {
	d.heap.Walk(f)
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
