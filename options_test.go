package configManager

import (
	"fmt"
	"math/rand"
	"testing"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(l int) string {
	b := make([]rune, l)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func optionPanicTester(c *ConfigSet, opt string) error {
	if p, _ := didPanic(func() { c.Lookup(opt) }); p {
		return fmt.Errorf("Panicked during lookup operation")
	}
	if p, _ := didPanic(func() { c.Set(opt, "") }); p {
		return fmt.Errorf("Panicked during set operation")
	}
	return nil
}

func Test_stringOption(t *testing.T) {
	var c ConfigSet
	rd := randString(32)
	var o = c.String("foo", rd)

	if p := optionPanicTester(&c, "foo"); p != nil {
		t.Fatalf("%v", p)
	}

	odv := c.Lookup("foo").DefValue
	if odv != rd {
		t.Fatalf("Option default value not set to expected value [%v], received [%v]\n", rd, odv)
	}

	rd = randString(24)

	c.Set("foo", rd)

	if *o != rd {
		t.Fatalf("Option not set to expected value [%v], received [%v]\n", rd, *o)
	}
}

func Test_boolOption(t *testing.T) {
	var c ConfigSet
	var o = c.Bool("foo", false)

	if p := optionPanicTester(&c, "foo"); p != nil {
		t.Fatalf("%v", p)
	}

	c.Set("foo", "true")

	if !*o {
		t.Fatalf("Option not set to expected value [true]")
	}
}

