package orderbook

import (
	"container/heap"

	"github.com/shopspring/decimal"
)

// Matches maps order ID to executed quantity.
type Matches map[string]decimal.Decimal

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

// MatchOrderLimit tries to match the given quantity at the given
// price.  Returns the order quantity left unmatched.
func (d *Ladder) MatchOrderLimit(price decimal.Decimal, taker Order) (decimal.Decimal, Matches) {
	level, ok := d.mapping[levelMapKey(price)]
	matches := make(Matches)
	if ok {
		remove := make([]*Order, 0, 2)
		for _, maker := range level.Orders.Iter() {
			if taker.Quantity.LessThanOrEqual(maker.Quantity) {
				// Given order (taker) is fully executed against an order
				// from the order book (maker), which gets partially
				// executed.
				matches[maker.ID] = taker.Quantity
				maker.Quantity = maker.Quantity.Sub(taker.Quantity)
				taker.Quantity = decimal.Zero
				if maker.Quantity.LessThanOrEqual(decimal.Zero) {
					remove = append(remove, maker)
				}
				// fmt.Printf("[1] taker=%v maker=%v\n", taker, maker)
				break
			} else {
				// Given order (taker) gets partially executed against an
				// order from the order book (maker), which gets fully
				// executed.
				matches[maker.ID] = maker.Quantity
				taker.Quantity = taker.Quantity.Sub(maker.Quantity)
				maker.Quantity = decimal.Zero
				remove = append(remove, maker)
				// fmt.Printf("[2] taker=%v maker=%v\n", taker, maker)
			}
		}

		for _, order := range remove {
			d.RemoveOrder(price, order.ID)
		}
	}
	return taker.Quantity, matches
}

func (d *Ladder) MatchOrderMarket(taker Order) (decimal.Decimal, Matches) {
	matches := make(Matches)
	// While there is still quantity to be matched and the ladder is not empty.
	for taker.Quantity.IsPositive() && d.heap.Len() > 0 {
		price := d.heap[0].Price
		q, xs := d.MatchOrderLimit(price, taker)
		taker.Quantity = q
		for k, v := range xs {
			matches[k] = v
		}
	}
	return taker.Quantity, matches
}

func (d *Ladder) GetOrder(price decimal.Decimal, id string) (Order, bool) {
	level, ok := d.mapping[levelMapKey(price)]
	if ok {
		return level.Orders.GetByID(id)
	}
	return Order{}, false
}

func (d *Ladder) TotalQuantity(price decimal.Decimal) decimal.Decimal {
	level, ok := d.mapping[levelMapKey(price)]
	if ok {
		return level.TotalQuantity()
	}
	return decimal.Zero
}

func (d *Ladder) Walk(f func(level *Level) bool) {
	d.heap.Walk(f)
}

func (d *Ladder) String() string {
	return d.heap.String()
}
