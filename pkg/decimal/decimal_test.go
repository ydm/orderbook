package decimal

import (
	"errors"
	"testing"
)

func TestNewFromString(t *testing.T) {
	f := func(s string, want int64) {
		t.Helper()

		d, err := NewFromString(s)
		if err != nil {
			t.Error(err)
		}

		have := int64(d)
		if have != want {
			t.Errorf("have %v, want %d", have, want)
		}
	}
	e := func(s string, target error) {
		t.Helper()

		_, err := NewFromString(s)
		if !errors.Is(err, target) {
			t.Errorf("error %v is not a subtype of %v", err, target)
		}
	}

	f("0", 0)
	f("0.", 0)
	f(".0", 0)

	f(".2", 2000_0000)
	f("2.", 2_0000_0000)
	f("12345678.", 1234_5678_0000_0000)
	f(".87654321", 8765_4321)

	f("1.2", 1_2000_0000)
	f("9.12345678", 9_1234_5678)
	f("12345678.87654321", 1234_5678_8765_4321)

	// Test trimming.
	f("9.123456789", 9_1234_5678)

	e(".", ErrInputStringError)
	e("", ErrInputStringError)

	e("1.2.3", ErrMultipleSeparators)
	e("1..3", ErrMultipleSeparators)
	e("..3", ErrMultipleSeparators)
	e("1..", ErrMultipleSeparators)
}

func TestDecimal_String(t *testing.T) {
	f := func(inp, want string) {
		t.Helper()

		x, err := NewFromString(inp)
		if err != nil {
			panic(err)
		}
		if have := x.String(); have != want {
			t.Errorf("have %s, want %s", have, want)
		}
	}
	f("0", "0.0")
	f("1.2", "1.2")
	f("9.12345678", "9.12345678")
	f("9.123456789", "9.12345678")
	f("12345678.87654321", "12345678.87654321")
	f("12345678.", "12345678.0")
	f(".87654321", "0.87654321")
	f("92233720368.54775807", "92233720368.54775807")
	// f("92233720368.54775808", "92233720368.54775808")
}

func TestDecimal_Add(t *testing.T) {
	f := func(a, b, want string) {
		t.Helper()

		x, _ := NewFromString(a)
		y, _ := NewFromString(b)
		x = x.Add(y)

		have := x.String()
		if have != want {
			t.Errorf("have %s, want %s", have, want)
		}
	}
	f("0.5", "0.5", "1.0")
	f("123.123", "321.321", "444.444")
	f("0.001", "0.999", "1.0")
}

func TestDecimal_LessThan(t *testing.T) {
	f := func(a, b string, want bool) {
		t.Helper()

		x, _ := NewFromString(a)
		y, _ := NewFromString(b)

		have := x.LessThan(y)
		if have != want {
			t.Errorf("have %t, want %t", have, want)
		}
	}
	f("1", "1", false)
	f("1", "1.00000001", true)

	f("10000000", "10000000", false)
	f("10000000", "10000000.00000001", true)

	f("0.99999999", "0.99999999", false)
	f("0.99999998", "0.99999999", true)

	f("0.999999998", "0.999999999", false) // Out of precision.

	f("92233720368.54775807", "92233720368.54775806", false)
	f("92233720368.54775807", "92233720368.54775807", false)
	f("92233720368.54775806", "92233720368.54775807", true)
}

func TestDecimal_Equal(t *testing.T) {
	f := func(a, b string, want bool) {
		t.Helper()

		x, _ := NewFromString(a)
		y, _ := NewFromString(b)

		have := x.Equal(y)
		if have != want {
			t.Errorf("have %t, want %t", have, want)
		}
	}
	f("1", "1", true)
	f("1", "1.00000001", false)

	f("10000000", "10000000", true)
	f("10000000", "10000000.00000001", false)

	f("0.99999999", "0.99999999", true)
	f("0.99999998", "0.99999999", false)

	f("0.999999998", "0.999999999", true) // Out of precision.

	f("92233720368.54775807", "92233720368.54775806", false)
	f("92233720368.54775807", "92233720368.54775807", true)
	f("92233720368.54775806", "92233720368.54775807", false)
}
