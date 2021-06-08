package orderbook

import (
	"strconv"
	"testing"

	"orderbook/pkg/decimal"
)

func TestLadder_Walk_1(t *testing.T) {
	assertEq := func(have, want bool) {
		t.Helper()
		if have != want {
			t.Error()
		}
	}
	f := func(x float64) decimal.Decimal {
		return decimal.NewFromStringPanic(strconv.FormatFloat(x, 'f', -1, 64))
	}

	d := NewLadder(Ask)
	assertEq(d.AddOrder(f(4), NewOrder("id1", f(0.1))), true)
	assertEq(d.AddOrder(f(4), NewOrder("id1", f(0.1))), false)
	assertEq(d.AddOrder(f(4), NewOrder("id1", f(0.1))), false)
	assertEq(d.AddOrder(f(2), NewOrder("id2", f(0.2))), true)
	assertEq(d.AddOrder(f(5), NewOrder("id3", f(0.3))), true)
	assertEq(d.AddOrder(f(1), NewOrder("id4", f(0.4))), true)
	assertEq(d.AddOrder(f(3), NewOrder("id5", f(0.5))), true)

	expected := []string{"1.0", "2.0", "3.0", "4.0", "5.0"}
	index := 0
	d.Walk(func(level *Level) bool {
		t.Helper()
		if level.Price.String() != expected[index] {
			t.Errorf("have %v, want %v", level.Price.String(), expected[index])
		}
		index++
		return true
	})
}

func TestLadder_RemoveOrder(t *testing.T) {
	assertEq := func(have, want bool) {
		t.Helper()
		if have != want {
			t.Error()
		}
	}
	f := func(x float64) decimal.Decimal {
		return decimal.NewFromStringPanic(strconv.FormatFloat(x, 'f', -1, 64))
	}

	d := NewLadder(Ask)
	assertEq(d.AddOrder(f(4), NewOrder("id1", f(0.1))), true)
	assertEq(d.AddOrder(f(2), NewOrder("id2", f(0.2))), true)
	assertEq(d.AddOrder(f(5), NewOrder("id3", f(0.3))), true)
	assertEq(d.AddOrder(f(1), NewOrder("id4", f(0.4))), true)
	assertEq(d.AddOrder(f(3), NewOrder("id5", f(0.5))), true)
	assertEq(d.RemoveOrder(f(4), "id1"), true)
	assertEq(d.RemoveOrder(f(4), "id1"), false)

	expected := []string{"1.0", "2.0", "3.0", "5.0"}
	index := 0
	d.Walk(func(level *Level) bool {
		t.Helper()

		if level.Price.String() != expected[index] {
			t.Errorf("have %v, want %v", level.Price.String(), expected[index])
		}
		index++
		return true
	})
}
