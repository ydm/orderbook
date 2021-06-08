package orderbook

import (
	"orderbook/pkg/decimal"
)

const (
	TypeLimit = iota
	TypeMarket
)

const (
	SideBuy = iota
	SideSell
)

type Order struct {
	ID       string          // 16 bytes
	Quantity decimal.Decimal //  8 bytes
} //                Total: at least 24 bytes
