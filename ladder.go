package orderbook

import (
	"container/heap"

	"github.com/shopspring/decimal"
)

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
	level, ok := d.mapping[levelMapKey(price)]
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
	d.mapping[levelMapKey(price)] = level
	heap.Push(&d.heap, level)

	return true
}

func (d *Ladder) RemoveOrder(price decimal.Decimal, ID string) bool {
	// Check if this level exists.
	level, ok := d.mapping[levelMapKey(price)]
	if ok {
		// Remove the order by its ID.
		ans := level.Orders.RemoveByID(ID)

		// If at this point the level is empty, remove it from
		// this Ladder.
		if level.Orders.Len() <= 0 {
			delete(d.mapping, levelMapKey(price))
			if heap.Remove(&d.heap, level.index) == nil {
				panic("illegal state")
			}
		}

		return ans
	}
	return false
}

// func (d *Ladder) MatchOrder(price decimal.Decimal, quantity decimal.Decimal) (Order, bool) {
// 	level, ok := d.mapping[price]
// 	if ok {
// 		for _, order := range level.Orders {
// 		}
// 	}
// }

func (d *Ladder) Walk(f func(level *Level) bool) {
	d.heap.Walk(f)
}
