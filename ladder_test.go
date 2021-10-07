package orderbook_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/ydm/orderbook"
)

func assertMatches(t *testing.T, have orderbook.Matches, want map[string]string) {
	t.Helper()

	if len(have) != len(want) {
		t.Errorf("have %d, want %d", len(have), len(want))
	}

	for k, v := range want {
		x, ok := have[k]
		if !ok {
			t.Error()
		}

		y, err := decimal.NewFromString(v)
		if err != nil {
			panic(err)
		}

		if !x.Equal(y) {
			t.Errorf("have %v, want %v", x, y)
		}
	}
}

func TestLadder_Walk_1(t *testing.T) {
	t.Parallel()

	assertEq := func(have, want bool) {
		t.Helper()

		if have != want {
			t.Errorf("have %v, want %v", have, want)
		}
	}

	d := orderbook.NewLadder(orderbook.Ask)
	assertEq(d.AddOrder(decimal.NewFromInt(4), orderbook.NewOrder("id1", decimal.NewFromFloat(0.1))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(4), orderbook.NewOrder("id1", decimal.NewFromFloat(0.1))), false)
	assertEq(d.AddOrder(decimal.NewFromInt(4), orderbook.NewOrder("id1", decimal.NewFromFloat(0.1))), false)
	assertEq(d.AddOrder(decimal.NewFromInt(2), orderbook.NewOrder("id2", decimal.NewFromFloat(0.2))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(5), orderbook.NewOrder("id3", decimal.NewFromFloat(0.3))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(1), orderbook.NewOrder("id4", decimal.NewFromFloat(0.4))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(3), orderbook.NewOrder("id5", decimal.NewFromFloat(0.5))), true)

	expected := []string{"1", "2", "3", "4", "5"}
	index := 0

	d.Walk(func(level *orderbook.Level) bool {
		t.Helper()
		if level.Price.String() != expected[index] {
			t.Errorf("have %v, want %s", level.Price, expected[index])
		}
		index++
		return true
	})
}

func TestLadder_RemoveOrder(t *testing.T) {
	t.Parallel()

	assertEq := func(have, want bool) {
		t.Helper()
		if have != want {
			t.Error()
		}
	}

	d := orderbook.NewLadder(orderbook.Ask)
	assertEq(d.AddOrder(decimal.NewFromInt(4), orderbook.NewOrder("id1", decimal.NewFromFloat(0.1))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(2), orderbook.NewOrder("id2", decimal.NewFromFloat(0.2))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(5), orderbook.NewOrder("id3", decimal.NewFromFloat(0.3))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(1), orderbook.NewOrder("id4", decimal.NewFromFloat(0.4))), true)
	assertEq(d.AddOrder(decimal.NewFromInt(3), orderbook.NewOrder("id5", decimal.NewFromFloat(0.5))), true)
	assertEq(d.RemoveOrder(decimal.NewFromInt(4), "id1"), true)
	assertEq(d.RemoveOrder(decimal.NewFromInt(4), "id1"), false)

	expected := []string{"1", "2", "3", "5"}
	index := 0

	d.Walk(func(level *orderbook.Level) bool {
		t.Helper()

		if level.Price.String() != expected[index] {
			t.Errorf("have %v, want %s", level.Price, expected[index])
		}
		index++
		return true
	})
}

func TestLadder_MatchOrderLimit_1(t *testing.T) {
	t.Parallel()

	d := orderbook.NewLadder(orderbook.Ask)
	f := func(present bool, price int64) int {
		t.Helper()

		dec := decimal.NewFromInt(price)
		key := orderbook.LevelMapKey(dec)
		level, ok := d.Mapping[key]
		if ok != present {
			t.Errorf("have %t, want %t", ok, present)
		}
		if ok {
			return level.Orders.Len()
		}
		return 0
	}

	d.AddOrder(decimal.NewFromInt(9), orderbook.NewOrder("id1", decimal.NewFromInt(10)))
	d.AddOrder(decimal.NewFromInt(10), orderbook.NewOrder("id2", decimal.NewFromInt(1)))
	d.AddOrder(decimal.NewFromInt(10), orderbook.NewOrder("id3", decimal.NewFromInt(2)))
	d.AddOrder(decimal.NewFromInt(10), orderbook.NewOrder("id4", decimal.NewFromInt(3)))
	d.AddOrder(decimal.NewFromInt(11), orderbook.NewOrder("id5", decimal.NewFromInt(10)))

	if have := f(true, 10); have != 3 {
		t.Errorf("have %d, want 3", have)
	}

	left, matches := d.MatchOrderLimit(decimal.NewFromInt(10), orderbook.NewOrder("id6", decimal.NewFromInt(3)))
	if !left.IsZero() {
		t.Errorf("have %v, want 0", left)
	}
	assertMatches(t, matches, map[string]string{"id2": "1", "id3": "2"})

	if have := f(true, 10); have != 1 {
		t.Errorf("have %d, want 1", have)
	}
}

func TestLadder_MatchOrderLimit_2(t *testing.T) {
	t.Parallel()

	d := orderbook.NewLadder(orderbook.Ask)
	f := func(present bool, price int64) int {
		t.Helper()

		dec := decimal.NewFromInt(price)
		key := orderbook.LevelMapKey(dec)
		level, ok := d.Mapping[key]
		if ok != present {
			t.Errorf("have %t, want %t", ok, present)
		}
		if ok {
			return level.Orders.Len()
		}
		return 0
	}

	d.AddOrder(decimal.NewFromInt(9), orderbook.NewOrder("id1", decimal.NewFromInt(10)))
	d.AddOrder(decimal.NewFromInt(10), orderbook.NewOrder("id2", decimal.NewFromInt(1)))
	d.AddOrder(decimal.NewFromInt(10), orderbook.NewOrder("id3", decimal.NewFromInt(2)))
	d.AddOrder(decimal.NewFromInt(10), orderbook.NewOrder("id4", decimal.NewFromInt(3)))
	d.AddOrder(decimal.NewFromInt(11), orderbook.NewOrder("id5", decimal.NewFromInt(10)))

	if have := f(true, 10); have != 3 {
		t.Errorf("have %d, want 3", have)
	}

	left, matches := d.MatchOrderLimit(decimal.NewFromInt(10), orderbook.NewOrder("id6", decimal.NewFromInt(10)))
	if !left.Equal(decimal.NewFromInt(4)) {
		t.Errorf("have %v, want 4", left)
	}
	assertMatches(t, matches, map[string]string{"id2": "1", "id3": "2", "id4": "3"})

	if have := f(false, 10); have != 0 {
		t.Errorf("have %d, want 0", have)
	}
}

func TestLadder_MatchOrderLimit_3(t *testing.T) {
	t.Parallel()

	d := orderbook.NewLadder(orderbook.Ask)
	f := func(present bool, price int64) int {
		t.Helper()

		dec := decimal.NewFromInt(price)
		key := orderbook.LevelMapKey(dec)
		level, ok := d.Mapping[key]
		if ok != present {
			t.Errorf("have %t, want %t", ok, present)
		}
		if ok {
			return level.Orders.Len()
		}
		return 0
	}

	d.AddOrder(decimal.NewFromInt(9), orderbook.NewOrder("id1", decimal.NewFromInt(10)))
	d.AddOrder(decimal.NewFromInt(10), orderbook.NewOrder("id2", decimal.NewFromInt(1)))
	d.AddOrder(decimal.NewFromInt(10), orderbook.NewOrder("id3", decimal.NewFromInt(2)))
	d.AddOrder(decimal.NewFromInt(10), orderbook.NewOrder("id4", decimal.NewFromInt(3)))
	d.AddOrder(decimal.NewFromInt(11), orderbook.NewOrder("id5", decimal.NewFromInt(10)))

	if have := f(true, 10); have != 3 {
		t.Errorf("have %d, want 3", have)
	}

	left, matches := d.MatchOrderLimit(decimal.NewFromInt(10), orderbook.NewOrder("id6", decimal.NewFromInt(2)))
	if !left.Equal(decimal.Zero) {
		t.Errorf("have %v, want 0", left)
	}
	assertMatches(t, matches, map[string]string{"id2": "1", "id3": "1"})
	// fmt.Printf("%v\n", d.heap)

	if have := f(true, 10); have != 2 {
		t.Errorf("have %d, want 2", have)
	}
}

func TestLadder_MatchOrderMarket_1(t *testing.T) {
	t.Parallel()

	d := orderbook.NewLadder(orderbook.Ask)
	f := func(present bool, price int64) int {
		t.Helper()

		dec := decimal.NewFromInt(price)
		key := orderbook.LevelMapKey(dec)
		level, ok := d.Mapping[key]
		if ok != present {
			t.Errorf("have %t, want %t", ok, present)
		}
		if ok {
			return level.Orders.Len()
		}
		return 0
	}

	d.AddOrder(decimal.NewFromInt(9), orderbook.NewOrder("id1", decimal.NewFromInt(10)))
	d.AddOrder(decimal.NewFromInt(10), orderbook.NewOrder("id2", decimal.NewFromInt(1)))
	d.AddOrder(decimal.NewFromInt(10), orderbook.NewOrder("id3", decimal.NewFromInt(2)))
	d.AddOrder(decimal.NewFromInt(10), orderbook.NewOrder("id4", decimal.NewFromInt(3)))
	d.AddOrder(decimal.NewFromInt(11), orderbook.NewOrder("id5", decimal.NewFromInt(10)))

	left, matches := d.MatchOrderMarket(orderbook.NewOrder("id6", decimal.NewFromInt(20)))
	if !left.Equal(decimal.Zero) {
		t.Errorf("have %v, want 0", left)
	}
	assertMatches(t, matches, map[string]string{"id1": "10", "id2": "1", "id3": "2", "id4": "3", "id5": "4"})

	if f(false, 9) != 0 || f(false, 10) != 0 || f(true, 11) != 1 {
		t.Error()
	}
}

func TestLadder_GetOrder(t *testing.T) {
	t.Parallel()

	d := orderbook.NewLadder(orderbook.Ask)

	_, ok := d.GetOrder(decimal.NewFromInt(9), "id1")
	if ok {
		t.Error()
	}
	d.AddOrder(decimal.NewFromInt(9), orderbook.NewOrder("id1", decimal.NewFromInt(10)))

	order, ok := d.GetOrder(decimal.NewFromInt(9), "id1")
	if !ok {
		t.Error()
	}
	if order.ID != "id1" {
		t.Error()
	}
	if !order.Quantity.Equal(decimal.NewFromInt(10)) {
		t.Error()
	}
}
