package configManager

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

func didPanic(f func()) (p bool, info any) {
	defer func() {
		if r := recover(); r != nil {
			p = true
			info = r
		}
	}()
	f()
	return
}

func valueTester[T any](v Value, valid []string, invalid []string, base *T, equals func(string, T) bool) error {
	// linter disable no unused variables
	if p, _ := didPanic(func() { _ = v.Get() }); p {
		return fmt.Errorf("Panicked when calling Get in empty value")
	}

	// whoever wrote default go linter with no options you moms a hoe
	if p, _ := didPanic(func() { _ = v.String() }); p {
		return fmt.Errorf("Panicked when calling String in empty value")
	}

	ret := v.Get()
	retType := reflect.TypeOf(ret)
	baseType := reflect.TypeOf(*base)

	if retType != baseType {
		return fmt.Errorf("Get returned incorrect type: got %v, want %v", retType, baseType)
	}

	for _, val := range valid {
		if err := v.Set(val); err != nil {
			return fmt.Errorf("Set(%q) rejected valid value: %v", val, err)
		}
		if ret := v.String(); ret != val {
			return fmt.Errorf("String() = %q, want %q", ret, val)
		}

		// since we don't know what stupid implementation is inside newValue and Get() returns any
		// the test author has to say what equality even means
		got := v.Get()
		if !equals(val, got.(T)) {
			return fmt.Errorf("Set(%q) produced unexpected value: %v", val, got)
		}
	}

	for _, val := range invalid {
		if err := v.Set(val); err == nil {
			return fmt.Errorf("Set(%q) accepted invalid value", val)
		}
	}

	return nil
}

// Example test
func Test_boolVal(t *testing.T) {
	var b bool
	v := newBoolValue(b, &b)

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
	v := newStringValue(s, &s)

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

