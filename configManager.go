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

// Parse the configuration from the given data and sets all options
func (c *ConfigSet) ParseFromData(data []byte) {
	switch c.Format {
	case JSON:
		c.Unmarshaller = json.Unmarshal
	case XML:
		c.Unmarshaller = xml.Unmarshal
	case CUSTOM:
		if c.Unmarshaller == nil {
			c.OnError(ErrNoParser)
			return
		}
	}

	var d = make(map[string]interface{})

	err := c.Unmarshaller(data, &d)
	if err != nil {
		c.error(err)
	}

	c.VisitAll(func(o *Option) {
		if _, present := c.actual[o.Name]; present {
			// do not set repeat options
			return
		}

		if v, ok := d[o.Name]; ok {
			vs := fmt.Sprint(v)
			o.Value.Set(vs)

			if c.actual == nil {
				c.actual = make(map[string]*Option)
			}
			c.actual[o.Name] = o
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

type myType int

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Generics
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

type valueFactory func(p any) Value

var valueFactories = map[reflect.Type]valueFactory{
	reflect.TypeOf((*bool)(nil)):    func(p any) Value { return newBoolValue(p.(*bool)) },
	reflect.TypeOf((*string)(nil)):  func(p any) Value { return newStringValue(p.(*string)) },
	reflect.TypeOf((*int)(nil)):     func(p any) Value { return newIntValue(p.(*int)) },
	reflect.TypeOf((*int32)(nil)):   func(p any) Value { return newInt32Value(p.(*int32)) },
	reflect.TypeOf((*int64)(nil)):   func(p any) Value { return newInt64Value(p.(*int64)) },
	reflect.TypeOf((*float64)(nil)): func(p any) Value { return newFloat64Value(p.(*float64)) },
	reflect.TypeOf((*float32)(nil)): func(p any) Value { return newFloat32Value(p.(*float32)) },
}

/* Register a new type of option in the configuration

This is a monadic interface and expects a factory function that wraps a type implementing Value interface
The factory function must receive the type as a pointer and return it wrapped in a Value interface

Usage examples:
if myType is implementing Value interface
type myType struct {...}

RegisterType(func(t *myType) Value {return (myType)(t)})

if myType is an alias to a basic type
type myType int64
RegisterType(func(t *myType) Value {return newInt64Value(int64(*t))})
*/
func RegisterType[T any](factory func(*T) Value) {
	var ptr *T
	t := reflect.TypeOf(ptr)

	valueFactories[t] = func(p any) Value {
		return factory(p.(*T))
	}
}

// whoever made methods not allowed to be generic: yo moms a hoe

// Add a new option to the configuration set c
// key is the name it has on the file and defaultValue is used when the option is not present
// p is the pointer the value will be set to after parsing the configuration
func AddOptionToSetVar[T any](c *ConfigSet, p *T, key string, defaultValue T) {
    *p = defaultValue
    t := reflect.TypeOf(p)

    factory, ok := valueFactories[t]
	//TODO: theres probable something better to do instead of panic but im too lazy to do it now
    if !ok {
        panic(fmt.Sprintf("no ValueFactory registered for type %v", t))
    }
    c.Var(factory(p), key)
}

// Add a new option to the configuration set c
// key is the name it has on the file and defaultValue is used when the option is not present
func AddOptionToSet[T any](c *ConfigSet, key string, defaultValue T) *T {
	p := new(T)
	AddOptionToSetVar(c, p, key, defaultValue)
	return p
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Global Binds
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
var globalConfig ConfigSet

// Add a new configuration option
// key is the name it has on the file and defaultValue is used when the option is not present
// p is the pointer the value will be set to after parsing the configuration
func AddOptionVar[T any](p *T, key string, defaultValue T) {
	AddOptionToSetVar(&globalConfig, p, key, defaultValue)
}

// Add a new configuration option
// key is the name it has on the file and defaultValue is used when the option is not present
func AddOption[T any](key string, defaultValue T) *T { return AddOptionToSet(&globalConfig, key, defaultValue) }

// Parse the configuration from the given data and sets all options
func ParseFromData(data []byte) { globalConfig.ParseFromData(data) }

// Parse the configuration file and sets all options
func Parse() { globalConfig.Parse() }

// Sets the location for the configuration file
func SetFileLocation(filename string) { globalConfig.Location = filename }

// Sets the format of the configuration file
// Expects constants JSON, XML or CUSTOM
// If set to CUSTOM a unmarshaller must be provided via SetFileUnmarshaller
func SetFileFormat(format fileFormat) { globalConfig.Format = format }

// Sets the unmarshaller to be used by a custom file format
// Function must abide by interface used by json.Unmarshal and xml.Unmarshal
func SetFileUnmarshaller(unmarshaller func(data []byte, v any) error) { globalConfig.Unmarshaller = unmarshaller }

// Sets output of error messages, by default stderr is used
// This behavior is only used if SetOnError was not given custom behavior
func SetErrorOutput(output io.Writer) { globalConfig.Output = output }

// Sets a function to be called when an error happens during parsing
// By default the behavior is to write error to Output which is stderr by default or set by SetErrorOutput
func SetOnError(onError func(error)) { globalConfig.OnError = onError }

// Visits all options in lexicographical order, calling fn for each
// Visits unset options
func VisitAll(fn func(*Option)) { globalConfig.VisitAll(fn) }

// Visits all options in lexicographical order, calling fn for each
// Visits only set options
func Visit(fn func(*Option)) { globalConfig.Visit(fn) }

// Sets the value of the named option
func Set(name, value string) error { return globalConfig.Set(name, value) }

// Lookups [Option] struct of the named option
func Lookup(name string) *Option { return globalConfig.Lookup(name) }

// Checks wether named option is set to it's zero value
func IsZeroValue(name string) (bool, error) { return globalConfig.IsZeroValue(name) }

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Basic Values
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// =-=-= boolValue
type boolValue bool

func newBoolValue(p *bool) *boolValue { return (*boolValue)(p) }

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

func newStringValue(p *string) *stringValue { return (*stringValue)(p) }

func (s *stringValue) Set(str string) error {
	*s = (stringValue)(str)
	return nil
}

func (s *stringValue) Get() any { return string(*s) }

func (s *stringValue) String() string { return string(*s) }

// =-=-= float64Value
type float64Value float64

func newFloat64Value(p *float64) *float64Value { return (*float64Value)(p) }

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		err = ErrParse
	}
	*f = float64Value(v)
	return err
}

func (f *float64Value) Get() any { return float64(*f) }

func (f *float64Value) String() string { return strconv.FormatFloat(float64(*f), 'g', -1, 64) }

// =-=-= float32Value
type float32Value float32

func newFloat32Value(p *float32) *float32Value { return (*float32Value)(p) }

func (f *float32Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 32)
	if err != nil {
		err = ErrParse
	}
	*f = float32Value(v)
	return err
}

func (f *float32Value) Get() any { return float32(*f) }

func (f *float32Value) String() string { return strconv.FormatFloat(float64(*f), 'g', -1, 32) }

// =-=-= intValue
type intValue int

func newIntValue(p *int) *intValue { return (*intValue)(p) }

func (i *intValue) Set(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		err = ErrParse
	}
	*i = intValue(v)
	return err
}

func (i *intValue) Get() any { return int(*i) }

func (i *intValue) String() string { return strconv.Itoa(int(*i)) }

// =-=-= int32Value
type int32Value int32

func newInt32Value(p *int32) *int32Value { return (*int32Value)(p) }

func (i *int32Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		err = ErrParse
	}
	*i = int32Value(v)
	return err
}

func (i *int32Value) Get() any { return int32(*i) }

func (i *int32Value) String() string { return strconv.FormatInt(int64(*i), 10) }

// =-=-= int64Value
type int64Value int64

func newInt64Value(p *int64) *int64Value { return (*int64Value)(p) }

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		err = ErrParse
	}
	*i = int64Value(v)
	return err
}

func (i *int64Value) Get() any { return int64(*i) }

func (i *int64Value) String() string { return strconv.FormatInt(int64(*i), 10) }

/*
still here if generics break again i dont have to rewrite it

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Option Binds
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

func (c *ConfigSet) BoolVar(p *bool, key string, defaultValue bool) {
	c.Var(newBoolValue(p), key)
}

func (c *ConfigSet) Bool(key string, defaultValue bool) *bool {
	p := new(bool)
	c.BoolVar(p, key, defaultValue)
	return p
}

func (c *ConfigSet) StringVar(p *string, key string, defaultValue string) {
	c.Var(newStringValue(p), key)
}

func (c *ConfigSet) String(key string, defaultValue string) *string {
	p := new(string)
	c.StringVar(p, key, defaultValue)
	return p
}

func (c *ConfigSet) IntVar(p *int, key string, defaultValue int) {
	c.Var(newIntValue(p), key)
}

func (c *ConfigSet) Int(key string, defaultValue int) *int {
	p := new(int)
	c.IntVar(p, key, defaultValue)
	return p
}

func (c *ConfigSet) FloatVar(p *float64, key string, defaultValue float64) {
	c.Var(newFloat64Value(p), key)
}

func (c *ConfigSet) Float(key string, defaultValue float64) *float64 {
	p := new(float64)
	c.FloatVar(p, key, defaultValue)
	return p
}

*/
