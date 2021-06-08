package orderbook

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestLadder_Walk_1(t *testing.T) {
	assertEq := func(have, want bool) {
		t.Helper()
		if have != want {
			t.Errorf("have %v, want %v", have, want)
		}
	}

	d := NewLadder(Ask)
	assertEq(d.AddOrder(decimal.NewFromInt(4), NewOrder("id1", decimal.NewFromFloat(0.1))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(4), NewOrder("id1", decimal.NewFromFloat(0.1))), false)
	assertEq(d.AddOrder(decimal.NewFromInt(4), NewOrder("id1", decimal.NewFromFloat(0.1))), false)
	assertEq(d.AddOrder(decimal.NewFromInt(2), NewOrder("id2", decimal.NewFromFloat(0.2))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(5), NewOrder("id3", decimal.NewFromFloat(0.3))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(1), NewOrder("id4", decimal.NewFromFloat(0.4))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(3), NewOrder("id5", decimal.NewFromFloat(0.5))), true)

	expected := []string{"1", "2", "3", "4", "5"}
	index := 0
	d.Walk(func(level *Level) bool {
		t.Helper()
		if level.Price.String() != expected[index] {
			t.Errorf("have %v, want %s", level.Price, expected[index])
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

	d := NewLadder(Ask)
	assertEq(d.AddOrder(decimal.NewFromInt(4), NewOrder("id1", decimal.NewFromFloat(0.1))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(2), NewOrder("id2", decimal.NewFromFloat(0.2))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(5), NewOrder("id3", decimal.NewFromFloat(0.3))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(1), NewOrder("id4", decimal.NewFromFloat(0.4))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(3), NewOrder("id5", decimal.NewFromFloat(0.5))), true)
	assertEq(d.RemoveOrder(decimal.NewFromInt(4), "id1"), true)
	assertEq(d.RemoveOrder(decimal.NewFromInt(4), "id1"), false)

	expected := []string{"1", "2", "3", "5"}
	index := 0
	d.Walk(func(level *Level) bool {
		t.Helper()

		if level.Price.String() != expected[index] {
			t.Errorf("have %v, want %s", level.Price, expected[index])
		}
		index++
		return true
	})
}

func TestLadder_MatchOrderLimit_1(t *testing.T) {
	d := NewLadder(Ask)
	f := func(present bool, price int64) int {
		t.Helper()

		dec := decimal.NewFromInt(price)
		key := levelMapKey(dec)
		level, ok := d.mapping[key]
		if ok != present {
			t.Errorf("have %t, want %t", ok, present)
		}
		if ok {
			return level.Orders.Len()
		}
		return 0
	}

	d.AddOrder(decimal.NewFromInt(9), NewOrder("id1", decimal.NewFromInt(10)))
	d.AddOrder(decimal.NewFromInt(10), NewOrder("id2", decimal.NewFromInt(1)))
	d.AddOrder(decimal.NewFromInt(10), NewOrder("id3", decimal.NewFromInt(2)))
	d.AddOrder(decimal.NewFromInt(10), NewOrder("id4", decimal.NewFromInt(3)))
	d.AddOrder(decimal.NewFromInt(11), NewOrder("id5", decimal.NewFromInt(10)))

	if have := f(true, 10); have != 3 {
		t.Errorf("have %d, want 3", have)
	}

	left := d.MatchOrderLimit(decimal.NewFromInt(10), NewOrder("id6", decimal.NewFromInt(3)))
	if !left.IsZero() {
		t.Errorf("have %v, want 0", left)
	}

	if have := f(true, 10); have != 1 {
		t.Errorf("have %d, want 1", have)
	}
}

func TestLadder_MatchOrderLimit_2(t *testing.T) {
	d := NewLadder(Ask)
	f := func(present bool, price int64) int {
		t.Helper()

		dec := decimal.NewFromInt(price)
		key := levelMapKey(dec)
		level, ok := d.mapping[key]
		if ok != present {
			t.Errorf("have %t, want %t", ok, present)
		}
		if ok {
			return level.Orders.Len()
		}
		return 0
	}

	d.AddOrder(decimal.NewFromInt(9), NewOrder("id1", decimal.NewFromInt(10)))
	d.AddOrder(decimal.NewFromInt(10), NewOrder("id2", decimal.NewFromInt(1)))
	d.AddOrder(decimal.NewFromInt(10), NewOrder("id3", decimal.NewFromInt(2)))
	d.AddOrder(decimal.NewFromInt(10), NewOrder("id4", decimal.NewFromInt(3)))
	d.AddOrder(decimal.NewFromInt(11), NewOrder("id5", decimal.NewFromInt(10)))

	if have := f(true, 10); have != 3 {
		t.Errorf("have %d, want 3", have)
	}

	left := d.MatchOrderLimit(decimal.NewFromInt(10), NewOrder("id6", decimal.NewFromInt(10)))
	if !left.Equal(decimal.NewFromInt(4)) {
		t.Errorf("have %v, want 4", left)
	}

	if have := f(false, 10); have != 0 {
		t.Errorf("have %d, want 0", have)
	}
}

func TestLadder_MatchOrderLimit_3(t *testing.T) {
	d := NewLadder(Ask)
	f := func(present bool, price int64) int {
		t.Helper()

		dec := decimal.NewFromInt(price)
		key := levelMapKey(dec)
		level, ok := d.mapping[key]
		if ok != present {
			t.Errorf("have %t, want %t", ok, present)
		}
		if ok {
			return level.Orders.Len()
		}
		return 0
	}

	d.AddOrder(decimal.NewFromInt(9), NewOrder("id1", decimal.NewFromInt(10)))
	d.AddOrder(decimal.NewFromInt(10), NewOrder("id2", decimal.NewFromInt(1)))
	d.AddOrder(decimal.NewFromInt(10), NewOrder("id3", decimal.NewFromInt(2)))
	d.AddOrder(decimal.NewFromInt(10), NewOrder("id4", decimal.NewFromInt(3)))
	d.AddOrder(decimal.NewFromInt(11), NewOrder("id5", decimal.NewFromInt(10)))

	if have := f(true, 10); have != 3 {
		t.Errorf("have %d, want 3", have)
	}

	left := d.MatchOrderLimit(decimal.NewFromInt(10), NewOrder("id6", decimal.NewFromInt(2)))
	if !left.Equal(decimal.Zero) {
		t.Errorf("have %v, want 0", left)
	}
	// fmt.Printf("%v\n", d.heap)

	if have := f(true, 10); have != 2 {
		t.Errorf("have %d, want 2", have)
	}
}

func TestLadder_MatchOrderMarket_1(t *testing.T) {
	d := NewLadder(Ask)
	f := func(present bool, price int64) int {
		t.Helper()

		dec := decimal.NewFromInt(price)
		key := levelMapKey(dec)
		level, ok := d.mapping[key]
		if ok != present {
			t.Errorf("have %t, want %t", ok, present)
		}
		if ok {
			return level.Orders.Len()
		}
		return 0
	}

	d.AddOrder(decimal.NewFromInt(9), NewOrder("id1", decimal.NewFromInt(10)))
	d.AddOrder(decimal.NewFromInt(10), NewOrder("id2", decimal.NewFromInt(1)))
	d.AddOrder(decimal.NewFromInt(10), NewOrder("id3", decimal.NewFromInt(2)))
	d.AddOrder(decimal.NewFromInt(10), NewOrder("id4", decimal.NewFromInt(3)))
	d.AddOrder(decimal.NewFromInt(11), NewOrder("id5", decimal.NewFromInt(10)))

	left := d.MatchOrderMarket(NewOrder("id6", decimal.NewFromInt(20)))
	if !left.Equal(decimal.Zero) {
		t.Errorf("have %v, want 0", left)
	}

	if f(false, 9) != 0 || f(false, 10) != 0 || f(true, 11) != 1 {
		t.Error()
	}
}
