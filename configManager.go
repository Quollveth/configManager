package configManager

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// Returned by Set when an option's value fails to parse
var ErrParse = errors.New("parse error")

// Returned by Set when an option's value is outside the defined range
var ErrRange = errors.New("value out of range")

// Returned by Parse when format is set to CUSTOM and no unmarshaller is provided
var ErrNoParser = errors.New("no parser provided for custom format")

// Used to dynamically store the value of an option
// Since all options are read from a file the default value is a string
// String may be called with a zero value receiver
type Value interface {
	String() string   // Returns this value as a string
	Set(string) error // Set the option to this value
	Get() any         // Get the value of this option
}

type Option struct {
	Name     string // name as it appears on the file
	DefValue string // Default value as string
	Value    Value
}

// Check wether this option is set to it's zero value
func (o *Option) IsZeroValue() (ok bool, err error) {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := reflect.TypeOf(o.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Pointer {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	// Catch panics calling the String method, which shouldn't prevent the
	// usage message from being printed, but that we should report to the
	// user so that they know to fix their code.
	defer func() {
		if e := recover(); e != nil {
			if typ.Kind() == reflect.Pointer {
				typ = typ.Elem()
			}
			ok, err = false, fmt.Errorf("panic calling String method on zero %v for flag %s: %v", typ, o.Name, e)
		}
	}()

	return o.Value.String() == z.Interface().(Value).String(), nil
}

type fileFormat int

const (
	JSON fileFormat = iota
	XML
	CUSTOM
)

type ConfigSet struct {
	formal map[string]*Option // All options
	actual map[string]*Option // Set options

	// Handler for errors
	// If left as nil errors are written to Output (stderr by default)
	OnError func(error)
	// Output of error messages, if nill stderr is used
	Output io.Writer

	// Location of configuration file
	Location string
	// Format of configuration file, must be set to constants JSON, XML or CUSTOM
	Format fileFormat

	// Unmarshaller to be used for CUSTOM fileFormat
	// If Format is set to CUSTOM and no unmarshaller is provided a call to Parse will return ErrNoParser
	// If Format is not set to CUSTOM this can remain unset or nil
	Unmarshaller func(data []byte, v any) error
}

// Returns a lexicographically sorted slice of all options
func (c *ConfigSet) sortOptions(opts map[string]*Option) []*Option {
	result := make([]*Option, len(opts))
	i := 0
	for _, o := range opts {
		result[i] = o
		i++
	}
	slices.SortFunc(result, func(a, b *Option) int {
		return strings.Compare(a.Name, b.Name)
	})

	return result
}

// Visits all options in lexicographical order, calling fn for each
// Visits unset options
func (c *ConfigSet) VisitAll(fn func(*Option)) {
	for _, o := range c.sortOptions(c.formal) {
		fn(o)
	}
}

// Visits all options in lexicographical order, calling fn for each
// Only visits set options
func (c *ConfigSet) Visit(fn func(*Option)) {
	for _, o := range c.sortOptions(c.actual) {
		fn(o)
	}
}

// Sets the value of the named option
func (c *ConfigSet) Set(name, value string) error {
	opt, ok := c.formal[name]
	if !ok {
		return fmt.Errorf("No such option: %v", name)
	}

	err := opt.Value.Set(value)
	if err != nil {
		return err
	}

	if c.actual == nil {
		c.actual = make(map[string]*Option)
	}

	c.actual[name] = opt
	return nil
}

// Lookups [Option] struct of the named option
func (c *ConfigSet) Lookup(name string) *Option { return c.formal[name] }

// Checks wether named option is set to it's zero value
func (c *ConfigSet) IsZeroValue(name string) (bool, error) {
	opt, ok := c.actual[name]
	if !ok {
		return false, fmt.Errorf("No such option %v", name)
	}

	return opt.IsZeroValue()
}

// Called by Parse when an error happens
func (c *ConfigSet) error(err error) {
	if c.OnError != nil {
		c.OnError(err)
		return
	}
	if c.Output != nil {
		c.Output.Write([]byte(err.Error()))
		return
	}

	os.Stderr.WriteString(err.Error())
}

// Defines an option with the specified name and default value.
// The type is defined by the first argument, which is a Value interface
// It's methods determine how the value is interacted with
func (c *ConfigSet) Var(value Value, name string) {
	opt := &Option{name, value.String(), value}

	_, exists := c.formal[name]
	if exists {
		panic(fmt.Sprintf("%s option redefined", name))
	}

	if c.formal == nil {
		c.formal = make(map[string]*Option)
	}

	c.formal[name] = opt
}

func (c *ConfigSet) ParseFromData(b []byte) {
	switch c.Format {
	case JSON: c.Unmarshaller = json.Unmarshal
	case XML: c.Unmarshaller = xml.Unmarshal
	case CUSTOM:
		if c.Unmarshaller == nil {
			c.OnError(ErrNoParser)
			return
		}
	}

	var data = make(map[string]string)

	err := c.Unmarshaller(b, &data)
	if err != nil {
		c.OnError(err)
	}

	c.VisitAll(func(o *Option) {
		if v, ok := data[o.Name]; ok {
			o.Value.Set(v)
		}
	})
}

// Parse the configuration file and sets all options
func (c *ConfigSet) Parse() {
	fdat, err := os.ReadFile(c.Location)
	if err != nil {
		c.OnError(err)
		return
	}

	c.ParseFromData(fdat)
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Default Values
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// =-=-= boolValue
type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		err = ErrParse
	}
	*b = boolValue(v)
	return err
}

func (b *boolValue) Get() any { return bool(*b) }

func (b *boolValue) String() string { return strconv.FormatBool(bool(*b)) }

// =-=-= stringValue
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(str string) error {
	*s = (stringValue)(str)
	return nil
}

func (s *stringValue) Get() any { return string(*s) }

func (s *stringValue) String() string { return string(*s) }

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Option Binds
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

func (c *ConfigSet) BoolVar(p *bool, key string, defaultValue bool) {
	c.Var(newBoolValue(defaultValue, p), key)
}

func (c *ConfigSet) Bool(key string, defaultValue bool) *bool {
	p := new(bool)
	c.BoolVar(p, key, defaultValue)
	return p
}

func (c *ConfigSet) StringVar(p *string, key string, defaultValue string) {
	c.Var(newStringValue(defaultValue, p), key)
}

func (c *ConfigSet) String(key string, defaultValue string) *string {
	p := new(string)
	c.StringVar(p, key, defaultValue)
	return p
}
