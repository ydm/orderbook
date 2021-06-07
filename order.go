package orderbook

import (
	decimal "orderbook/pkg/decimal"
)

const (
	TypeLimit = iota
	TypeMarket
)

type Order struct {
	Type   int
	Symbol string
	ID     string
	Price  decimal.Decimal
}
