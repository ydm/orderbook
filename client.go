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

const (
	StateInitial = iota
	StatePlaced
	StateFilled
	StatePartiallyFilled
	StateCanceled
)

type ClientOrder struct {
	// Symbol           string
	Side             int
	OriginalQuantity decimal.Decimal
	ExecutedQuantity decimal.Decimal
	Price            decimal.Decimal
	ID               string
	Type             int
	// State            int
}
