package configManager

import (
	"strconv"
	"strings"
	"testing"
)

func Test_stringRangeVal(t *testing.T) {
	var s string

	v := newStringRangeVal(&s, false, "foo", "bar", "baz")

	if err := valueTester(
		v,
		[]string{
			"foo",
			"bar",
			"baz",
			"FOO",
		},

		[]string{
			"",
			"hello",
			"foobar",
			"literally anything",
			"69 haha",
		},
		&s,
		func(a string, b string) bool {
			if v.caseSensitive {
				return a == b
			}
			return strings.ToLower(a) == strings.ToLower(b)
		},
	); err != nil {
		t.Fatal(err)
	}
}

func Test_stringRangeOption(t *testing.T) {
	var c ConfigSet

	_, err := StringRangeSet(&c, "foo", "", false, "bar", "baz")

	if err == nil {
		t.Fatal("Option accepted invalid default value")
	}

	o, err := StringRangeSet(&c, "direction", "up", false, "up", "down", "left", "right")

	jason := `{"direction":"left"}`
	if err != nil {
		t.Fatal(err)
	}

	if e, p := didPanic(func() { err = c.ParseFromData([]byte(jason)) }); e {
		t.Fatal(p)
	}

	if *o != "left" {
		t.Fatalf("Option set to unexpected value, expected [left] got %v", *o)
	}
}

func Test_int32RangeVal(t *testing.T) {
	var n int32

	v := newInt32RangeValue(&n, -10, 10)

	if err := valueTester(
		v,
		[]string{
			"-10",
			"10",
			"0",
			"5",
		},
		[]string{
			"-15",
			"15",
		},
		&n,
		func(a string, b int32) bool { return strconv.FormatInt(int64(b), 10) == a },
	); err != nil {
		t.Fatal(err)
	}
}

func Test_int64RangeVal(t *testing.T) {
	var n int64

	v := newInt64RangeValue(&n, -10, 10)

	if err := valueTester(
		v,
		[]string{
			"-10",
			"10",
			"0",
			"5",
		},
		[]string{
			"-15",
			"15",
		},
		&n,
		func(a string, b int64) bool { return strconv.FormatInt(b, 10) == a },
	); err != nil {
		t.Fatal(err)
	}
}

func Test_float32RangeVal(t *testing.T) {
	var f float32

	v := newFloat32RangeValue(&f, -10.0, 10.0)

	if err := valueTester(
		v,
		[]string{
			"-10",
			"10",
			"0",
			"5.5",
		},
		[]string{
			"-10.1",
			"10.1",
		},
		&f,
		func(a string, b float32) bool {
			return strconv.FormatFloat(float64(b), 'f', -1, 32) == a
		},
	); err != nil {
		t.Fatal(err)
	}
}

func Test_float64RangeVal(t *testing.T) {
	var f float64

	v := newFloat64RangeValue(&f, -10.0, 10.0)

	if err := valueTester(
		v,
		[]string{
			"-10",
			"10",
			"0",
			"5.5",
		},
		[]string{
			"-10.1",
			"10.1",
		},
		&f,
		func(a string, b float64) bool {
			return strconv.FormatFloat(b, 'f', -1, 64) == a
		},
	); err != nil {
		t.Fatal(err)
	}
}


