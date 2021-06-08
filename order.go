package orderbook

import "github.com/shopspring/decimal"

const (
	TypeLimit = iota
	TypeMarket
)

const (
	SideBuy = iota
	SideSell
)

type Order struct {
	ID             string          // 16 bytes
	Quantity       decimal.Decimal //  8 bytes
	insertionIndex int             //  8 bytes
} //                      Total: at least 32 bytes

func NewOrder(id string, quantity decimal.Decimal) Order {
	return Order{
		ID:       id,
		Quantity: quantity,
	}
}
