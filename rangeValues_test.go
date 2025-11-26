package configManager

import (
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
