package configManager

func unused(any) {}

type myStrut struct {
	x, y float64
}

// make strut implement Value interface
func (s myStrut) String() string      { return "" }
func (s myStrut) Set(ss string) error { return nil }
func (s myStrut) Get() any            { return myStrut{} }

func newMyStrutValue(p *myStrut) Value {
	return (*myStrut)(p)
}

func example() {
	c := ConfigSet{}

	stringOption, _ := AddOptionToSet(
		&c,            // config set because methods cannot be generic for reasons
		"greeting",    // key name in file
		"hello world", // default value for option
	)

	// since default value (third parameter) is the generic, return value gets inferred to *string
	// so stringOption is of type *string

	floatOption, _ := AddOptionToSet(&c, "maximum", 0.0)
	// in this one floatOption is of type *float64 which got inferred from 0.0

	// register new type in map and give it a factory function
	RegisterType(func(t *myStrut) Value { return (myStrut{}) })

	strutOption, _ := AddOptionToSet(&c, "my option", myStrut{})
	// in this one strutOption is inferred to *myStrut

	unused(floatOption)
	unused(stringOption)
	unused(strutOption)
}
