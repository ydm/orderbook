package orderbook

import (
	"errors"
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
)

// Submit a market order against an empty order book.
func TestBook_AddOrder_1(t *testing.T) {
	b := NewBook()
	order := ClientOrder{
		Side:             SideBuy,
		OriginalQuantity: decimal.NewFromInt(1),
		ExecutedQuantity: decimal.Zero,
		ID:               "id1",
		Type:             TypeMarket,
	}
	err := b.AddOrder(order)
	if !errors.Is(err, ErrMarketNotFullyExecuted) {
		t.Error()
	}
	x, err := b.GetOrder("id1")
	if err != nil {
		t.Error(err)
	}
	if !x.ExecutedQuantity.IsZero() {
		t.Error()
	}
}

// Submit a market order against a single limit order.
func TestBook_AddOrder_2(t *testing.T) {
	b := NewBook()
	limit := ClientOrder{
		Side:             SideSell,
		OriginalQuantity: decimal.NewFromInt(2),
		ExecutedQuantity: decimal.Zero,
		Price:            decimal.NewFromInt(10000),
		ID:               "limit",
		Type:             TypeLimit,
	}
	market := ClientOrder{
		Side:             SideBuy,
		OriginalQuantity: decimal.NewFromInt(1),
		ExecutedQuantity: decimal.Zero,
		ID:               "market",
		Type:             TypeMarket,
	}
	if err := b.AddOrder(limit); err != nil {
		t.Error(err)
	}
	if err := b.AddOrder(market); err != nil {
		t.Error(err)
	}

	fmt.Printf("%v\n", b.asks)
	fmt.Printf("%v\n", b.bids)
}
