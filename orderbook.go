// Package orderbook represents a simplified order book.  It supports
// submitting, canceling, querying and matching orders.
package orderbook

import (
	"errors"
	"sync"

	"github.com/shopspring/decimal"
)

var (
	ErrCannotCancelMarketOrder     = errors.New("cannot cancel market order")
	ErrCannotCancelOrder           = errors.New("given order is not eligible for cancelation")
	ErrInvalidID                   = errors.New("invalid order ID")
	ErrInvalidPrice                = errors.New("invalid order price")
	ErrInvalidQuantity             = errors.New("invalid order quantity")
	ErrInvalidSide                 = errors.New("invalid order side")
	ErrInvalidType                 = errors.New("invalid order type")
	ErrMarketOrderNotFullyExecuted = errors.New("market order not (fully) executed")
	ErrMarketOrderHasPrice         = errors.New("given market order has price set")
	ErrOrderDoesNotExist           = errors.New("order with this ID does not exist")
	ErrOrderExists                 = errors.New("order with this ID already exists")
)

type Book struct {
	Asks Ladder
	Bids Ladder
	mu   sync.Mutex

	// Ser, please imagine this is a database.
	database      map[string]ClientOrder
	databaseMutex sync.Mutex
}

func NewBook() *Book {
	return &Book{
		Asks:          NewLadder(Ask),
		Bids:          NewLadder(Bid),
		mu:            sync.Mutex{},
		database:      make(map[string]ClientOrder),
		databaseMutex: sync.Mutex{},
	}
}

func (b *Book) checkOrder(order ClientOrder) error {
	// Check order properties.
	if order.OriginalQuantity.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidQuantity
	}

	if !order.ExecutedQuantity.IsZero() {
		return ErrInvalidQuantity
	}

	if order.ID == "" {
		return ErrInvalidID
	}

	// Check if order with this ID already exists.
	b.databaseMutex.Lock()
	_, ok := b.database[order.ID] //nolint:ifshort
	b.databaseMutex.Unlock()

	if ok {
		return ErrOrderExists
	}

	return nil
}

func (b *Book) matchSides(side int) (*Ladder, *Ladder, error) {
	switch side {
	case SideBuy:
		return &b.Bids, &b.Asks, nil
	case SideSell:
		return &b.Asks, &b.Bids, nil
	default:
		return nil, nil, ErrInvalidSide
	}
}

func (b *Book) store(order ClientOrder, matches Matches) {
	// Store new order.
	b.databaseMutex.Lock()
	b.database[order.ID] = order

	// Update matched orders.
	for id, x := range matches {
		maker, ok := b.database[id]
		if !ok {
			panic("illegal state")
		}

		maker.ExecutedQuantity = maker.ExecutedQuantity.Add(x)
		b.database[maker.ID] = maker
	}

	b.databaseMutex.Unlock()
}

//nolint:cyclop
func (b *Book) AddOrder(order ClientOrder) error {
	if err := b.checkOrder(order); err != nil {
		return err
	}

	// We'll be matching this order against the opposite ladder, i.e. if
	// this is a buy order, we'll try to match it first against the asks.
	// If it's also a limit order and left unmatched, it will be added.
	my, op, err := b.matchSides(order.Side)
	if err != nil {
		return err
	}

	x := NewOrder(order.ID, order.OriginalQuantity)

	var (
		left    decimal.Decimal
		matches Matches
	)

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
		// order book.  If the order remains not fully executed, it's placed in
		// the order book.
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
	b.store(order, matches)

	if order.Type == TypeMarket && order.ExecutedQuantity.LessThan(order.OriginalQuantity) {
		return ErrMarketOrderNotFullyExecuted
	}

	return nil
}

func (b *Book) CancelOrder(id string) error {
	if id == "" {
		return ErrInvalidID
	}

	// Check if order exists.
	b.databaseMutex.Lock()
	order, ok := b.database[id]
	b.databaseMutex.Unlock()

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
		b.mu.Lock()
		defer b.mu.Unlock()

		if b.Bids.RemoveOrder(order.Price, order.ID) {
			return nil
		}
	case SideSell:
		b.mu.Lock()
		defer b.mu.Unlock()

		if b.Asks.RemoveOrder(order.Price, order.ID) {
			return nil
		}
	default:
		return ErrInvalidSide
	}

	// At this point this order was not eligible for cancellation.
	return ErrCannotCancelOrder
}

func (b *Book) GetOrder(id string) (ClientOrder, error) {
	b.databaseMutex.Lock()
	order, ok := b.database[id]
	b.databaseMutex.Unlock()

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
	b.Asks.Walk(ask)
	b.Bids.Walk(bid)
	b.mu.Unlock()

	return ans
}
