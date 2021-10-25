package orderbook

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// BinarySearch returns index of the search key, if it is contained in
// the array, otherwise (-(insertion point) – 1).
func BinarySearch(xs []*Order, x int) int {
	low, high := 0, len(xs)

	for low < high {
		mid := low + (high-low)/2

		if xs[mid].InsertionIndex < x {
			low = mid + 1
		} else {
			high = mid
		}
	}

	if low < len(xs) && xs[low].InsertionIndex == x {
		return low
	}

	return -low - 1
}

// OrderQueue holds all the orders at a particular level of the order book.  It keeps them
// in a queue (FIFO), so orders of the same price level get executed in the order they
// were submitted.  OrderQueue also allows querying using the order ID.
type OrderQueue struct {
	queue []*Order

	// indices maps an order ID to its insertion order index.
	// This way removing by ID may use binarySearch() and thus
	// take O(logN + copy).
	indices map[string]int

	next int
}

func NewOrderQueue(n int) OrderQueue {
	return OrderQueue{
		queue:   make([]*Order, 0, n),
		indices: make(map[string]int),
		next:    0,
	}
}

func (q *OrderQueue) Add(order Order) bool {
	if _, ok := q.indices[order.ID]; ok {
		// There is already an order with this ID.
		return false
	}

	// Set insertionIndex to order and also save it to our
	// ID -> insertionIndex mapping.
	order.InsertionIndex = q.next
	q.indices[order.ID] = q.next
	q.next++

	// Append order to queue.
	q.queue = append(q.queue, &order)

	return true
}

func (q *OrderQueue) Remove() *Order {
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
		i := BinarySearch(q.queue, insertionIndex)
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

func (q *OrderQueue) GetByID(orderID string) (Order, bool) {
	insertionIndex, ok := q.indices[orderID]

	if ok {
		i := BinarySearch(q.queue, insertionIndex)

		if i >= 0 {
			order := q.queue[i]

			return *order, true
		}
	}

	return Order{
		ID:             orderID,
		Quantity:       decimal.Zero,
		InsertionIndex: insertionIndex,
	}, false
}

func (q *OrderQueue) Iter() []*Order {
	return q.queue
}

func (q *OrderQueue) Len() int {
	// Len() also makes sure both of our data structures have the
	// same length.
	if len(q.queue) != len(q.indices) {
		msg := fmt.Sprintf("invariant: len(queue)=%d len(indices)=%d\n", len(q.queue), len(q.indices))
		panic(msg)
	}

	return len(q.queue)
}

func (q *OrderQueue) String() string {
	var b strings.Builder

	for i, order := range q.queue {
		if i > 0 {
			fmt.Fprintf(&b, "\n")
		}

		fmt.Fprintf(&b, "    %v", order)
	}

	return b.String()
}
