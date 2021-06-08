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
	low, high := 0, len(xs)
	for low < high {
		mid := low + (high-low)/2
		if xs[mid].insertionIndex < x {
			low = mid + 1
		} else {
			high = mid
		}
	}
	if low < len(xs) && xs[low].insertionIndex == x {
		return low
	} else {
		return -low - 1
	}
}

// OrderQueue holds all the orders at a particular level of the order book.  It keeps them
// in a queue (FIFO), so orders of the same price level get executed in the order they
// were submitted.  OrderQueue also allows querying using the order ID.
type OrderQueue struct {
	queue []Order

	// indices maps an order ID to its insertion order index.
	// This way removing by ID may use binarySearch() and thus
	// take O(2logN + copy).
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

	// Set insertionIndex to order and also save it to our
	// ID -> insertionIndex mapping.
	order.insertionIndex = q.next
	q.indices[order.ID] = q.next
	q.next++

	// Append order to queue.
	q.queue = append(q.queue, order)

	return true
}

func (q *OrderQueue) Remove() Order {
	// Take order.
	order := q.queue[0]

	// Pop queue.
	q.queue = q.queue[1:]

	// Delete from the ID -> index mapping.
	delete(q.indices, order.ID)

	return order
}

func (q *OrderQueue) RemoveByID(orderID string) bool {
	// Check if we have an order with this ID.
	insertionIndex, ok := q.indices[orderID]
	if ok {
		// If yes, locate its index in the queue.
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
	// Len() also makes sure both of our data structures have the
	// same length.
	if len(q.queue) != len(q.indices) {
		fmt.Printf("len(queue)=%d len(indices)=%d\n",
			len(q.queue), len(q.indices))
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
	Price  decimal.Decimal // Also serves as Key() in the heap.
	Orders OrderQueue      // All of the orders on this level.
	Type   int             // Ask or Bid, controls behavior of Key().
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

// LevelHeap lets us dynamically insert and remove orders in O(logN)
// and also iterate Levels ordered by price with a just bit of extra
// work [see Walk()].
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
// inspections and modifications.  It is either of type Ask or Bid.
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
	// First check if this level exists.
	level, ok := d.mapping[price]
	if ok {
		// Add the order to this existing level.
		return level.Orders.Add(o)
	}

	// Level does not exist.  Create it and add the order.
	level = NewLevel(price, d.Type)
	if !level.Orders.Add(o) {
		panic("illegal state")
	}

	// Save the newly made level into our heap and mapping.
	d.mapping[price] = level
	heap.Push(&d.heap, level)

	return true
}

func (d *Ladder) RemoveOrder(price decimal.Decimal, ID string) bool {
	// Check if this level exists.
	level, ok := d.mapping[price]
	if ok {
		// Remove the order by its ID.
		ans := level.Orders.RemoveByID(ID)

		// If at this point the level is empty, remove it from
		// this Ladder.
		if level.Orders.Len() <= 0 {
			delete(d.mapping, price)
			if heap.Remove(&d.heap, level.index) == nil {
				panic("illegal state")
			}
		}

		return ans
	}
	return false
}

func (d *Ladder) Walk(f func(level *Level) bool) {
	d.heap.Walk(f)
}
