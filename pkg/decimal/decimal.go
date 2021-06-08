// Package decimal represents a POSITIVE decimal numeral with an
// integer part and fractional part encoded together into an int64
// (aliased as Decimal).  Precision is hard-coded to 8.
package decimal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Decimal int64

const (
	Precision = 8

	// TODO: Can I express this a function of Precision?
	FractionalDenominator = 1_0000_0000
)

var ErrMultipleSeparators = errors.New("multiple decimal separators")
var ErrInputStringError = errors.New("could not convert input string")
var ErrOutOfRange = errors.New("decimal out of range")

// prep pads a numeric string with zeros until it reaches a length of
// `Precision` characters.  If the given string is longer, it gets
// truncated.
func prep(s string) string {
	if len(s) > Precision {
		return s[:Precision]
	}
	xs := make([]byte, Precision)
	var (
		r rune
		i int
	)
	xs[0] = '0' // In case s is empty.
	for i, r = range s {
		xs[i] = byte(r)
	}
	for i += 1; i < Precision; i++ {
		xs[i] = '0'
	}
	return string(xs)
}

func NewFromString(s string) (Decimal, error) {
	if s == "" {
		return Decimal(0), ErrInputStringError
	}
	xs := strings.Split(s, ".")
	if len(xs) > 2 {
		return Decimal(0), ErrMultipleSeparators
	}
	var (
		integer    int64
		fractional int64
		err        error
	)
	if len(xs[0]) > 0 {
		integer, err = strconv.ParseInt(xs[0], 10, 64)
		if err != nil {
			return Decimal(0), err
		}
	} else if len(xs[1]) <= 0 {
		return Decimal(0), ErrInputStringError
	}
	if len(xs) >= 2 {
		var err error
		fractional, err = strconv.ParseInt(prep(xs[1]), 10, 64)
		if err != nil {
			return Decimal(0), err
		}
	}
	x := integer*FractionalDenominator + fractional
	return Decimal(x), nil
}

func NewFromStringPanic(s string) Decimal {
	x, err := NewFromString(s)
	if err != nil {
		panic(err)
	}
	return x
}

func (d Decimal) Integer() int64 {
	return d.Raw() / FractionalDenominator
}

func (d Decimal) Fractional() int64 {
	return d.Raw() % FractionalDenominator
}

func (d Decimal) Raw() int64 {
	return int64(d)
}

func (d Decimal) Add(rhs Decimal) Decimal {
	return Decimal(d.Raw() + rhs.Raw())
}

func (d Decimal) LessThan(rhs Decimal) bool {
	return d.Raw() < rhs.Raw()
}

func (d Decimal) Equal(rhs Decimal) bool {
	return d.Raw() == rhs.Raw()
}

func (d Decimal) LessThanEqual(rhs Decimal) bool {
	return d.LessThan(rhs) || d.Equal(rhs)
}

func (d Decimal) GreaterThan(rhs Decimal) bool {
	return !d.LessThanEqual(rhs)
}

func (d Decimal) GreaterThanEqual(rhs Decimal) bool {
	return !d.LessThan(rhs)
}

func (d Decimal) String() string {
	x := d.Fractional()
	for x != 0 && (x%10) == 0 {
		x /= 10
	}
	return fmt.Sprintf("%d.%d", d.Integer(), x)
}
