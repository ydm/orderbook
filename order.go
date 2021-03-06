package orderbook

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type Order struct {
	//nolint:godox
	//TODO: Turn ID into int64!
	ID             string          // 16 bytes
	Quantity       decimal.Decimal // 16 bytes
	InsertionIndex int             //  8 bytes
} //                      Total: at least 40 bytes

func NewOrder(id string, quantity decimal.Decimal) Order {
	return Order{
		ID:             id,
		Quantity:       quantity,
		InsertionIndex: 0,
	}
}

func (o Order) String() string {
	return fmt.Sprintf("[Order ID=%s Quantity=%v]", o.ID, o.Quantity)
}
