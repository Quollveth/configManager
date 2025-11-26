package configManager

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
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
		if ret := v.String(); strings.ToLower(ret) != strings.ToLower(val) {
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

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(l int) string {
	b := make([]rune, l)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func optionTester[T any](c *ConfigSet, key string, val T) error {
	if p, i := didPanic(func() { AddOptionToSet(c, key, val) }); p {
		return fmt.Errorf("Panicked during add operation: %v", i)
	}

	if c.Lookup(key) == nil {
		return fmt.Errorf("Option added as nil")
	}

	if p, i := didPanic(func() { c.Lookup(key) }); p {
		return fmt.Errorf("Panicked during lookup operation: %v", i)
	}

	if p, i := didPanic(func() { c.Set(key, "") }); p {
		return fmt.Errorf("Panicked during set operation: %v", i)
	}
	return nil
}
