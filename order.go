package orderbook

import (
	"orderbook/pkg/decimal"
)

// const (
// 	TypeLimit = iota
// 	TypeMarket
// )

const (
	SideBuy = iota
	SideSell
)

type Order struct {
	Quantity decimal.Decimal //  8 bytes
	ID       string          // 16 bytes
} //                Total: at least 24 bytes
