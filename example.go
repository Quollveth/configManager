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

func test() {
	c := ConfigSet{}

	stringOption := AddOptionToSet(
		c,             // config set
		"greeting",    // key name in file
		"hello world", // default value for option
	)
	// since default value (third parameter) is the generic, return value gets inferred to *string
	// so stringOption is of type *string

	floatOption := AddOptionToSet(c, "maximum", 0.0)
	// in this one floatOption is of type *float64 which got inferred from 0.0

	RegisterType(myStrut{}, func(p any) Value { return newMyStrutValue(p.(*myStrut)) })

	strutOption := AddOptionToSet(c, "my option", myStrut{})
	// in this one strutOption is inferred to *myStrut


	unused(floatOption)
	unused(stringOption)
	unused(strutOption)
}
