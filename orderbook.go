// Package orderbook represents a simplified order book.  It supports
// submitting, canceling, querying and matching orders.
package orderbook

import (
	"errors"
	"sync"

	"github.com/shopspring/decimal"
)

var (
	ErrCannotCancelMarketOrder = errors.New("cannot cancel market order")
	ErrCannotCancelOrder       = errors.New("given order is not eligible for cancelation")
	ErrInvalidID               = errors.New("invalid order ID")
	ErrInvalidPrice            = errors.New("invalid order price")
	ErrInvalidQuantity         = errors.New("invalid order quantity")
	ErrInvalidSide             = errors.New("invalid order side")
	ErrInvalidType             = errors.New("invalid order type")
	ErrMarketNotFullyExecuted  = errors.New("market order is not (fully) executed")
	ErrMarketOrderHasPrice     = errors.New("given market order has price set")
	ErrOrderDoesNotExist       = errors.New("order with this ID does not exist")
	ErrOrderExists             = errors.New("order with this ID already exists")
)

type Book struct {
	asks Ladder
	bids Ladder

	// Imagine this is a database.
	//
	// TODO: TURN THIS INTO sync.Map!
	database map[string]ClientOrder

	mu sync.Mutex
}

func NewBook() *Book {
	return &Book{
		asks:     NewLadder(Ask),
		bids:     NewLadder(Bid),
		database: make(map[string]ClientOrder),
	}
}

func (b *Book) AddOrder(order ClientOrder) error {
	// Check order properties.
	if order.OriginalQuantity.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidQuantity
	}
	if order.ID == "" {
		return ErrInvalidID
	}

	// Check if order with this ID already exists.
	_, ok := b.database[order.ID]
	if ok {
		return ErrOrderExists
	}

	// We'll be matching this order against the opposite ladder, i.e. if
	// this is a buy order, we'll try to match it first against the asks.
	// If it's also a limit order and left unmatched, it will be added.
	var my, op *Ladder
	switch order.Side {
	case SideBuy:
		my = &b.bids
		op = &b.asks
	case SideSell:
		my = &b.asks
		op = &b.bids
	default:
		return ErrInvalidSide
	}

	x := NewOrder(order.ID, order.OriginalQuantity)
	var left decimal.Decimal
	var matches Matches

	switch order.Type {
	case TypeMarket:
		if !order.Price.IsZero() {
			return ErrMarketOrderHasPrice
		}

		// Market orders get executed immediately against the orders we have in
		// the order book.  If the market order is not fully executed, we return
		// an error.
		b.mu.Lock()
		left, matches = op.MatchOrderMarket(x)
		b.mu.Unlock()

	case TypeLimit:
		if order.Price.IsNegative() {
			return ErrInvalidPrice
		}

		// Limit orders may first be matched against the opposite side of the
		// order book.  If the order remains unexecuted, it's placed in the order
		// book.
		b.mu.Lock()
		left, matches = op.MatchOrderLimit(order.Price, x)
		if left.IsPositive() {
			my.AddOrder(order.Price, NewOrder(order.ID, left))
		}
		b.mu.Unlock()

	default:
		return ErrInvalidType
	}

	order.ExecutedQuantity = order.OriginalQuantity.Sub(left)

	// Store the new order to our database.
	b.database[order.ID] = order

	// Update matched orders in our database.
	for id, x := range matches {
		maker, ok := b.database[id]
		if !ok {
			panic("illegal state")
		}
		maker.ExecutedQuantity = maker.ExecutedQuantity.Add(x)
		b.database[maker.ID] = maker
	}

	if order.Type == TypeMarket && !order.OriginalQuantity.Sub(order.ExecutedQuantity).IsZero() {
		return ErrMarketNotFullyExecuted
	}

	return nil
}

func (b *Book) CancelOrder(id string) error {
	if id == "" {
		return ErrInvalidID
	}

	// Check if order exists.
	order, ok := b.database[id]
	if !ok {
		return ErrOrderDoesNotExist
	}

	// Check the order type.
	if order.Type == TypeMarket {
		return ErrCannotCancelMarketOrder
	} else if order.Type != TypeLimit {
		return ErrInvalidType
	}

	// Actually try to remove the order.
	switch order.Side {
	case SideBuy:
		if b.bids.RemoveOrder(order.Price, order.ID) {
			return nil
		}
	case SideSell:
		if b.asks.RemoveOrder(order.Price, order.ID) {
			return nil
		}
	default:
		return ErrInvalidSide
	}

	// At this point this order was not eligible for cancellation.
	return ErrCannotCancelOrder
}

func (b *Book) GetOrder(id string) (ClientOrder, error) {
	order, ok := b.database[id]
	if !ok {
		return order, ErrOrderDoesNotExist
	}
	return order, nil
}

func (b *Book) GetSnapshot(depth int) Snapshot {
	ans := Snapshot{
		Asks: make([]ClientLevel, 0, depth),
		Bids: make([]ClientLevel, 0, depth),
	}

	askDepth := 0
	ask := func(level *Level) bool {
		if askDepth >= depth {
			return false
		}
		askDepth++

		ans.Asks = append(ans.Asks, ClientLevel{
			Price:    level.Price,
			Quantity: level.TotalQuantity(),
		})
		return true
	}

	bidDepth := 0
	bid := func(level *Level) bool {
		if bidDepth >= depth {
			return false
		}
		bidDepth++

		ans.Bids = append(ans.Bids, ClientLevel{
			Price:    level.Price,
			Quantity: level.TotalQuantity(),
		})
		return true
	}

	b.mu.Lock()
	b.asks.Walk(ask)
	b.bids.Walk(bid)
	b.mu.Unlock()
	return ans
}
