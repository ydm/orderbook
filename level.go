package orderbook

import "github.com/shopspring/decimal"

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

func (v *Level) Key() decimal.Decimal {
	switch v.Type {
	case Ask:
		return v.Price
	case Bid:
		return v.Price.Neg()
	default:
		panic("illegal type")
	}
}

func (v *Level) Less(rhs *Level) bool {
	return v.Key().LessThan(rhs.Key())
}

// +-----------+
// | LevelHeap |
// +-----------+

// LevelHeap is a Level collection that lets us dynamically insert and
// remove orders in O(logN) and iterate levels ordered by price with
// just bit of extra work [see Walk()].
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
type LevelMap map[int64]*Level

func levelMapKey(d decimal.Decimal) int64 {
	return d.Mul(decimal.NewFromInt(1_0000_0000)).IntPart()
}
