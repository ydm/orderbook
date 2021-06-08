// Package orderbook represents a simplified order book.  It supports
// submitting, canceling, querying and matching orders.
package orderbook

type Book struct {
	Asks Ladder
	Bids Ladder
}

func NewBook() *Book {
	return &Book{
		Asks: NewLadder(Ask),
		Bids: NewLadder(Bid),
	}
}

// func (b *Book) AddOrder(orderType int) {
// }
