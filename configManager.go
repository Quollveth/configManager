package configManager

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// Returned by Set when an option's value fails to parse
var ErrParse = errors.New("parse error")

// Returned by Parse when format is set to CUSTOM and no marshaller or unmarshaller is provided
var ErrNoParser = errors.New("no parser provided for custom format")

// Returned by Parse when value is not within the allowed range
var ErrRange = errors.New("value outside allowed range")

// Used to dynamically store the value of an option
// Since all options are read from a file the default value is a string
// Methods may be called with a zero value receiver
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

	// Location of configuration file
	Location string
	// Format of configuration file, must be set to constants JSON, XML or CUSTOM
	Format fileFormat

	// Unmarshaller to be used for CUSTOM fileFormat
	// If Format is set to CUSTOM and no unmarshaller is provided a call to Parse will return ErrNoParser
	// If Format is not set to CUSTOM this can remain unset or nil
	Unmarshaller func(data []byte, v any) error

	// Marshaller to be used for CUSTOM fileFormat
	// If Format is set to CUSTOM and no marshaller is provided a call to Save will return ErrNoParser
	// If Format is not set to CUSTOM this can remain unset or nil
	Marshaller func(v any) ([]byte, error)
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

// Defines an option with the specified name and default value.
// The type is defined by the first argument, which is a Value interface
// It's methods determine how the value is interacted with
func (c *ConfigSet) Var(value Value, name string) error {
	opt := &Option{name, value.String(), value}

	_, exists := c.formal[name]
	if exists {
		return fmt.Errorf("%s option redefined", name)
	}

	if c.formal == nil {
		c.formal = make(map[string]*Option)
	}

	c.formal[name] = opt
	return nil
}

// Parse the configuration from the given data and sets all options
func (c *ConfigSet) ParseFromData(data []byte) error {
	switch c.Format {
	case JSON: c.Unmarshaller = json.Unmarshal
	case XML: c.Unmarshaller = xml.Unmarshal
	case CUSTOM:
		if c.Unmarshaller == nil {
			return ErrNoParser
		}
	}

	var d = make(map[string]interface{})

	err := c.Unmarshaller(data, &d)
	if err != nil {
		return err
	}

	c.VisitAll(func(o *Option) {
		if _, present := c.actual[o.Name]; present {
			// do not set repeat options
			return
		}

		if v, ok := d[o.Name]; ok {
			vs := fmt.Sprint(v)

			e := o.Value.Set(vs)
			if e != nil {
				err = e
				return
			}

			if c.actual == nil {
				c.actual = make(map[string]*Option)
			}
			c.actual[o.Name] = o
		}
	})

	return err
}

// Parse the configuration file and sets all options
func (c *ConfigSet) Parse() error {
	if c.Location == "" {
		return fmt.Errorf("No file location provided")
	}

	fdat, err := os.ReadFile(c.Location)
	if err != nil {
		return err
	}

	return c.ParseFromData(fdat)
}

// Save the configuration file with set options to provided location
// Set may be called to provide values to options, otherwise default values will be used
func (c *ConfigSet) Save() error {
	if c.Location == "" {
		return fmt.Errorf("No file location provided")
	}

	err := os.MkdirAll(path.Dir(c.Location), 0755)
	if err != nil {
		return fmt.Errorf("Could not save configuration: %v", err)
	}

	data, err := c.SaveTo()
	if err != nil {
		return fmt.Errorf("Could not save configuration: %v", err)
	}

	err = os.WriteFile(c.Location, data, 0644)
	return err
}

// Write configuration file with set options and returns data
// Set may be called to provide values to options, otherwise default values will be used
func (c *ConfigSet) SaveTo() ([]byte, error) {
	switch c.Format {
	case JSON: c.Marshaller = func(v any) ([]byte, error) { return json.MarshalIndent(v, "", "  ") }
	case XML: c.Marshaller = func(v any) ([]byte, error) { return xml.MarshalIndent(v, "", "  ") }
	case CUSTOM:
		if c.Marshaller == nil {
			return nil, ErrNoParser
		}
	}

	toSave := make(map[string]any)
	c.VisitAll(func(o *Option) {
		toSave[o.Name] = o.Value.Get()
	})

	return c.Marshaller(toSave)
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Generics
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

type valueFactory func(p any) Value

var valueFactories = map[reflect.Type]valueFactory{
	reflect.TypeOf((*bool)(nil)):    func(p any) Value { return newBoolValue(p.(*bool)) },
	reflect.TypeOf((*string)(nil)):  func(p any) Value { return newStringValue(p.(*string)) },
	reflect.TypeOf((*int32)(nil)):   func(p any) Value { return newInt32Value(p.(*int32)) },
	reflect.TypeOf((*int64)(nil)):   func(p any) Value { return newInt64Value(p.(*int64)) },
	reflect.TypeOf((*float64)(nil)): func(p any) Value { return newFloat64Value(p.(*float64)) },
	reflect.TypeOf((*float32)(nil)): func(p any) Value { return newFloat32Value(p.(*float32)) },
}

/*
	Register a new type of option in the configuration

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
func AddOptionToSetVar[T any](c *ConfigSet, p *T, key string, defaultValue T) error {
	*p = defaultValue
	t := reflect.TypeOf(p)

	factory, ok := valueFactories[t]
	if !ok {
		return fmt.Errorf("no ValueFactory registered for type %v", t)
	}
	return c.Var(factory(p), key)
}

// Add a new option to the configuration set c
// key is the name it has on the file and defaultValue is used when the option is not present
// type of option is inferred from the default value, only if a custom type is passed an error may be returned in case it lacks a Value wrapper
// to register an option with a custom type first RegisterType must be called to ensure it has a Value interface wrapper
// when called with a primitive type (bool, int, int32, int64, float32, float64 or string) this function should never return an error
func AddOptionToSet[T any](c *ConfigSet, key string, defaultValue T) (*T, error) {
	p := new(T)
	err := AddOptionToSetVar(c, p, key, defaultValue)
	return p, err
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
func AddOption[T any](key string, defaultValue T) (*T, error) {
	return AddOptionToSet(&globalConfig, key, defaultValue)
}

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
func SetFileUnmarshaller(unmarshaller func(data []byte, v any) error) {
	globalConfig.Unmarshaller = unmarshaller
}

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

// Save the configuration file with set options to provided location
// Set may be called to provide values to options, otherwise default values will be used
func Save() error { return globalConfig.Save() }

// Write configuration file with set options and returns data
// Set may be called to provide values to options, otherwise default values will be used
func SaveTo() ([]byte, error) { return globalConfig.SaveTo() }

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

func (b boolValue) Get() any { return bool(b) }

func (b boolValue) String() string { return strconv.FormatBool(bool(b)) }

// =-=-= stringValue
type stringValue string

func newStringValue(p *string) *stringValue { return (*stringValue)(p) }

func (s *stringValue) Set(str string) error {
	*s = (stringValue)(str)
	return nil
}

func (s stringValue) Get() any { return string(s) }

func (s stringValue) String() string { return string(s) }

// =-=-= float64Value
type float64Value float64

func newFloat64Value(p *float64) *float64Value { return (*float64Value)(p) }

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return ErrParse
	}
	*f = float64Value(v)
	return err
}

func (f float64Value) Get() any { return float64(f) }

func (f float64Value) String() string { return strconv.FormatFloat(float64(f), 'g', -1, 64) }

// =-=-= float32Value
type float32Value float32

func newFloat32Value(p *float32) *float32Value { return (*float32Value)(p) }

func (f *float32Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return ErrParse
	}
	*f = float32Value(v)
	return err
}

func (f float32Value) Get() any { return float32(f) }

func (f float32Value) String() string { return strconv.FormatFloat(float64(f), 'g', -1, 32) }

// =-=-= int32Value
type int32Value int32

func newInt32Value(p *int32) *int32Value { return (*int32Value)(p) }

func (i *int32Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 32)
	v32 := int32(v)
	if err != nil {
		return ErrParse
	}
	*i = int32Value(v32)
	return err
}

func (i int32Value) Get() any { return int32(i) }

func (i int32Value) String() string { return strconv.FormatInt(int64(i), 10) }

// =-=-= int64Value
type int64Value int64

func newInt64Value(p *int64) *int64Value { return (*int64Value)(p) }

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return ErrParse
	}
	*i = int64Value(v)
	return err
}

func (i int64Value) Get() any { return int64(i) }

func (i int64Value) String() string { return strconv.FormatInt(int64(i), 10) }

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Range Values
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// =-=-= stringRangeValue

type stringRangeValue struct {
	ptr           *string
	val           string
	caseSensitive bool
	allowed       []string
}

func newStringRangeVal(p *string, caseSensitive bool, allowed ...string) *stringRangeValue {
	if caseSensitive {
		return &stringRangeValue{p, *p, caseSensitive, allowed}
	}

	a := []string{}
	for _, s := range allowed {
		a = append(a, strings.ToLower(s))
	}

	return &stringRangeValue{p, *p, caseSensitive, a}
}

func (s *stringRangeValue) Set(str string) error {
	if !s.caseSensitive {
		str = strings.ToLower(str)
	}

	if !slices.Contains(s.allowed, str) {
		return ErrRange
	}

	s.val = str
	*s.ptr = str
	return nil
}

func (s stringRangeValue) Get() any { return string(s.val) }

func (s stringRangeValue) String() string { return s.val }

// Defines a new string option with a specific set of allowed values on the set c, setting option to a value outside allowed set will result in ErrRange
// Empty string is NOT an accepted value unless specified
func StringRangeVarSet(c *ConfigSet, p *string, key, defaultValue string, caseSensitive bool, allowed ...string) error {
	v := newStringRangeVal(p, caseSensitive, allowed...)
	err := v.Set(defaultValue)
	if err != nil {
		return err
	}
	*p = defaultValue
	return c.Var(v, key)
}

// Defines a new string option with a specific set of allowed values on the set c, setting option to a value outside allowed set will result in ErrRange
// Empty string is NOT an accepted value unless specified
func StringRangeSet(c *ConfigSet, key, defaultValue string, caseSensitive bool, allowed ...string) (*string, error) {
	p := new(string)
	err := StringRangeVarSet(c, p, key, defaultValue, caseSensitive, allowed...)
	return p, err
}

// Defines a new string option with a specific set of allowed values, setting option to a value outside allowed set will result in ErrRange
// Empty string is NOT an accepted value unless specified
func StringRangeVar(p *string, key, defaultValue string, caseSensitive bool, allowed ...string) error {
	return StringRangeVarSet(&globalConfig, p, key, defaultValue, caseSensitive, allowed...)
}

// Defines a new string option with a specific set of allowed values, setting option to a value outside allowed set will result in ErrRange
// Empty string is NOT an accepted value unless specified
func StringRange(key, defaultValue string, caseSensitive bool, allowed ...string) (*string, error) {
	return StringRangeSet(&globalConfig, key, defaultValue, caseSensitive, allowed...)
}

// =-=-= int32Range

type int32RangeValue struct {
	ptr           *int32
	val, min, max int32
}

func newInt32RangeValue(p *int32, min, max int32) *int32RangeValue {
	return &int32RangeValue{
		ptr: p,
		min: min,
		max: max,
	}
}

func (i *int32RangeValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 32)
	if err != nil {
		return ErrParse
	}
	v32 := int32(v)

	if v32 > i.max || v32 < i.min {
		return ErrRange
	}

	i.val = v32
	*i.ptr = v32

	return nil
}

func (i int32RangeValue) Get() any { return i.val }

func (i int32RangeValue) String() string { return strconv.FormatInt(int64(i.val), 10) }

// Defines a new int32 option with the specified range (inclusive) on the set c, setting option to a value outside allowed range result in ErrRange
// 0 is not a valid value unless within range
func Int32RangeVarSet(c *ConfigSet, p *int32, key string, defaultValue, minv, maxv int32) error {
	v := newInt32RangeValue(p, minv, maxv)
	err := v.Set(strconv.FormatInt(int64(defaultValue), 10))
	if err != nil {
		return err
	}
	*p = defaultValue
	return c.Var(v, key)
}

// Defines a new int32 option with the specified range (inclusive) on the set c, setting option to a value outside allowed range result in ErrRange
// 0 is not a valid value unless within range
func Int32RangeSet(c *ConfigSet ,key string, defaultValue, minv, maxv int32) (*int32, error) {
	p := new(int32)
	err := Int32RangeVarSet(c, p, key, defaultValue, minv, maxv)
	return p, err
}

// =-=-= int64Range

type int64RangeValue struct {
	ptr           *int64
	val, min, max int64
}

func newInt64RangeValue(p *int64, min, max int64) *int64RangeValue {
	return &int64RangeValue{
		ptr: p,
		min: min,
		max: max,
	}
}

func (i *int64RangeValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return ErrParse
	}

	if v > i.max || v < i.min {
		return ErrRange
	}

	i.val = v
	*i.ptr = v
	return nil
}

func (i int64RangeValue) Get() any { return i.val }

func (i int64RangeValue) String() string { return strconv.FormatInt(i.val, 10) }

func Int64RangeVarSet(c *ConfigSet, p *int64, key string, defaultValue, minv, maxv int64) error {
	v := newInt64RangeValue(p, minv, maxv)
	err := v.Set(strconv.FormatInt(defaultValue, 10))
	if err != nil {
		return err
	}
	*p = defaultValue
	return c.Var(v, key)
}

func Int64RangeSet(c *ConfigSet, key string, defaultValue, minv, maxv int64) (*int64, error) {
	p := new(int64)
	err := Int64RangeVarSet(c, p, key, defaultValue, minv, maxv)
	return p, err
}

// =-=-= float32Range

type float32RangeValue struct {
	ptr        *float32
	val, min, max float32
}

func newFloat32RangeValue(p *float32, min, max float32) *float32RangeValue {
	return &float32RangeValue{
		ptr: p,
		min: min,
		max: max,
	}
}

func (f *float32RangeValue) Set(s string) error {
	v, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return ErrParse
	}
	v32 := float32(v)

	if v32 > f.max || v32 < f.min {
		return ErrRange
	}

	f.val = v32
	*f.ptr = v32
	return nil
}

func (f float32RangeValue) Get() any { return f.val }

func (f float32RangeValue) String() string { return strconv.FormatFloat(float64(f.val), 'f', -1, 32) }

func Float32RangeVarSet(c *ConfigSet, p *float32, key string, defaultValue, minv, maxv float32) error {
	v := newFloat32RangeValue(p, minv, maxv)
	err := v.Set(strconv.FormatFloat(float64(defaultValue), 'f', -1, 32))
	if err != nil {
		return err
	}
	*p = defaultValue
	return c.Var(v, key)
}

func Float32RangeSet(c *ConfigSet, key string, defaultValue, minv, maxv float32) (*float32, error) {
	p := new(float32)
	err := Float32RangeVarSet(c, p, key, defaultValue, minv, maxv)
	return p, err
}

// =-=-= float64Range

type float64RangeValue struct {
	ptr        *float64
	val, min, max float64
}

func newFloat64RangeValue(p *float64, min, max float64) *float64RangeValue {
	return &float64RangeValue{
		ptr: p,
		min: min,
		max: max,
	}
}

func (f *float64RangeValue) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return ErrParse
	}

	if v > f.max || v < f.min {
		return ErrRange
	}

	f.val = v
	*f.ptr = v
	return nil
}

func (f float64RangeValue) Get() any { return f.val }

func (f float64RangeValue) String() string { return strconv.FormatFloat(f.val, 'f', -1, 64) }

func Float64RangeVarSet(c *ConfigSet, p *float64, key string, defaultValue, minv, maxv float64) error {
	v := newFloat64RangeValue(p, minv, maxv)
	err := v.Set(strconv.FormatFloat(defaultValue, 'f', -1, 64))
	if err != nil {
		return err
	}
	*p = defaultValue
	return c.Var(v, key)
}

func Float64RangeSet(c *ConfigSet, key string, defaultValue, minv, maxv float64) (*float64, error) {
	p := new(float64)
	err := Float64RangeVarSet(c, p, key, defaultValue, minv, maxv)
	return p, err
}


