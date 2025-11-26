package configManager

import (
	"fmt"
	"math"
	"strconv"
	"testing"
)

func Test_boolVal(t *testing.T) {
	var b bool
	v := newBoolValue(&b)

	if err := valueTester(
		v,
		[]string{
			"true",
			"false",
		},

		[]string{
			"tru",
			"",
			"null",
			"ok",
		},
		&b,
		func(s string, b bool) bool { return strconv.FormatBool(b) == s },
	); err != nil {
		t.Fatal(err)
	}
}

func Test_stringVal(t *testing.T) {
	var s string
	v := newStringValue(&s)

	if err := valueTester(
		v,
		[]string{
			"true",
			"TRue",
			"67",
			"",
			"𐐘",
		},

		[]string{}, // any string is a valid string value
		&s,
		func(a string, b string) bool { return a == b },
	); err != nil {
		t.Fatal(err)
	}
}

func Test_floatVal(t *testing.T) {
	var f float64
	v := newFloat64Value(&f)

	if err := valueTester(
		v,
		[]string{
			"69",
			"420.69",
			"NaN",
			"-42",
			"-6.7",
			fmt.Sprint(math.MaxFloat64),
		},

		[]string{
			"",
		},
		&f,
		func(a string, b float64) bool { return a == strconv.FormatFloat(b, 'g', -1, 64) },
	); err != nil {
		t.Fatal(err)
	}
}

func Test_int64Val(t *testing.T) {
	var f int64
	v := newInt64Value(&f)

	if err := valueTester(
		v,
		[]string{
			"69",
			"-42",
			fmt.Sprint(math.MaxInt32 + 1),
			fmt.Sprint(math.MinInt32 - 1),
			fmt.Sprint(math.MaxInt64),
			fmt.Sprint(math.MinInt64),
		},

		[]string{
			"",
			"6.9",
			"NaN",
		},
		&f,
		func(a string, b int64) bool { return a == strconv.FormatInt(b, 10) },
	); err != nil {
		t.Fatal(err)
	}
}

