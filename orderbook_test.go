package orderbook

import "github.com/shopspring/decimal"

func newDecimalPanic(s string) decimal.Decimal {
	x, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return x
}
