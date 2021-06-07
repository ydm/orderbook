package orderbook

import (
	"sync"

	"orderbook/pkg/decimal"
)

// +------------+
// | OrderQueue |
// +------------+

// OrderQueue holds all the orders at a particular level of the order book.
type OrderQueue []Order

func NewOrderQueue(n int) OrderQueue {
	return make([]Order, 0, n)
}

func (q *OrderQueue) Add(o Order) {
	*q = append(*q, o)
}

func (q *OrderQueue) Remove() Order {
	old := *q
	o := old[0]
	*q = old[1:]
	return o
}

func (q OrderQueue) Len() int {
	return len(q)
}

// +----------+
// | LevelMap |
// +----------+

// LevelMap maps Price keys to OrderQueue values.
type LevelMap struct {
	inner sync.Map
}

func (m *LevelMap) Delete(price decimal.Decimal) {
	m.inner.Delete(price)
}

func (m *LevelMap) Load(price decimal.Decimal) (queue OrderQueue, ok bool) {
	var value interface{}
	value, ok = m.inner.Load(price)
	queue = value.(OrderQueue)
	return
}

func (m *LevelMap) LoadAndDelete(price decimal.Decimal) (queue OrderQueue, loaded bool) {
	var value interface{}
	value, loaded = m.inner.LoadAndDelete(price)
	queue = value.(OrderQueue)
	return
}

func (m *LevelMap) LoadOrStore(key, queue OrderQueue) (actual OrderQueue, loaded bool) {
	var value interface{}
	value, loaded = m.inner.LoadOrStore(key, queue)
	actual = value.(OrderQueue)
	return
}

func (m *LevelMap) Range(f func(price decimal.Decimal, queue OrderQueue) bool) {
	g := func(key, queue interface{}) bool {
		a := key.(decimal.Decimal)
		b := queue.(OrderQueue)
		return f(a, b)
	}
	m.inner.Range(g)
}

func (m *LevelMap) Store(price decimal.Decimal, queue OrderQueue) {
	m.inner.Store(price, queue)
}
