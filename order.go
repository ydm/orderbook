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
	Side     int             //  4 bytes
	Quantity decimal.Decimal //  8 bytes
	ID       string          // 16 bytes
} //            Total: at least 28 bytes
