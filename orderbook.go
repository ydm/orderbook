// Package orderbook represents a simplified order book.  It supports
// submitting, canceling, querying and matching orders.
package orderbook

import (
	"errors"
	"sync"
)

var (
	ErrOrderExists             = errors.New("order with this ID already exists")
	ErrOrderDoesNotExist       = errors.New("order with this ID does not exist")
	ErrCannotCancelMarketOrder = errors.New("cannot cancel market order")
	ErrInvalidSide             = errors.New("invalid order side")
	ErrInvalidType             = errors.New("invalid order type")
	ErrMarketNotFullyExecuted  = errors.New("market order is not (fully) executed")
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
	// Check if order with this ID already exists.
	_, ok := b.database[order.ID]
	if ok {
		return ErrOrderExists
	}

	// We'll be matching this order against the opposite ladder, i.e. if this is a
	// buying order, we'll try to match it first against the asks.
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

	switch order.Type {
	case TypeMarket:
		// Market orders get executed immediately against the orders we have in
		// the order book.  If the market order is not fully executed, we return
		// an error.
		b.mu.Lock()
		left, _ := op.MatchOrderMarket(x)
		b.mu.Unlock()

		order.ExecutedQuantity = order.OriginalQuantity.Sub(left)
		b.database[order.ID] = order

		if !left.IsZero() {
			return ErrMarketNotFullyExecuted
		}
	case TypeLimit:
		// Limit orders may first be matched against the opposite side of the
		// order book.  If the order remains unexecuted, it's placed in the order
		// book.
		b.mu.Lock()
		left, _ := op.MatchOrderLimit(order.Price, x)
		if left.IsPositive() {
			my.AddOrder(order.Price, NewOrder(order.ID, left))
		}
		b.mu.Unlock()

		order.ExecutedQuantity = order.OriginalQuantity.Sub(left)
		b.database[order.ID] = order
	default:
		return ErrInvalidType
	}

	return nil
}

func (b *Book) CancelOrder(id string) error {
	order, ok := b.database[id]
	if !ok {
		return ErrOrderDoesNotExist
	}
	if order.Type == TypeMarket {
		return ErrCannotCancelMarketOrder
	}
	switch order.Side {
	case SideBuy:
	case SideSell:
	}
	return nil
}

func (b *Book) GetOrder(id string) (ClientOrder, error) {
	order, ok := b.database[id]
	if !ok {
		return order, ErrOrderDoesNotExist
	}
	return order, nil
}
