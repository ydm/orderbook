package orderbook

import (
	decimal "orderbook/pkg/decimal"
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
	Type     int
	Symbol   string
	Side     int
	Quantity decimal.Decimal
	Price    decimal.Decimal
	ID       string
}
