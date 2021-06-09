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
	Side             int             `json:"side"`
	OriginalQuantity decimal.Decimal `json:"quantity"`
	ExecutedQuantity decimal.Decimal `json:"executedQuantity"`
	Price            decimal.Decimal `json:"price"`
	ID               string          `json:"id"`
	Type             int             `json:"type"`
	// State            int
}

type ClientLevel struct {
	Price    decimal.Decimal `json:"price"`
	Quantity decimal.Decimal `json:"quantity"`
}

type Snapshot struct {
	Asks []ClientLevel
	Bids []ClientLevel
}
