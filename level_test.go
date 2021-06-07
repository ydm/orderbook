package orderbook

import (
	"orderbook/pkg/decimal"
	"testing"
)

func TestOrderQueue(t *testing.T) {
	q := NewOrderQueue(2)
	if q.Len() != 0 {
		t.Fail()
	}

	inp := Order{
		Side:     SideBuy,
		Quantity: decimal.NewFromStringPanic("1"),
		ID:       "7bfa0e20",
	}
	q.Add(inp)
	if q.Len() != 1 {
		t.Fail()
	}

	out := q.Remove()
	if q.Len() != 0 {
		t.Fail()
	}

	if inp.Side != out.Side || !inp.Quantity.Equal(out.Quantity) || inp.ID != out.ID {
		t.Fail()
	}

	// Make sure also the queue enlarges.
	for i := 1; i <= 16; i++ {
		q.Add(inp)
		if q.Len() != i {
			t.Fail()
		}
	}
}

func TestLevelMap(t *testing.T) {

}
