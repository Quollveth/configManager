package configManager

/*
import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func Test_boolVal(t *testing.T) {
	var b bool
	v := newBoolValue(&b)

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
	v := newStringValue(&s)

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

func Test_floatVal(t *testing.T) {
	var f float64
	v := newFloat64Value(&f)

	if err := valueTester(
		v,
		[]string{
			"69",
			"420.69",
			"NaN",
			"-42",
			"-6.7",
			fmt.Sprint(math.MaxFloat64),
		},

		[]string{
			"",
		},
		&f,
		func(a string, b float64) bool { return a == strconv.FormatFloat(b, 'g', -1, 64) },
	); err != nil {
		t.Fatal(err)
	}
}

func Test_intVal(t *testing.T) {
	var f int
	v := newIntValue(&f)

	if err := valueTester(
		v,
		[]string{
			"69",
			"-42",
			fmt.Sprint(math.MaxInt32),
			fmt.Sprint(math.MinInt32),
		},

		[]string{
			"",
			"6.9",
			"NaN",
		},
		&f,
		func(a string, b int) bool { return a == strconv.Itoa(b) },
	); err != nil {
		t.Fatal(err)
	}
}

func Test_int64Val(t *testing.T) {
	var f int64
	v := newInt64Value(&f)

	if err := valueTester(
		v,
		[]string{
			"69",
			"-42",
			fmt.Sprint(math.MaxInt32 + 1),
			fmt.Sprint(math.MinInt32 - 1),
			fmt.Sprint(math.MaxInt64),
			fmt.Sprint(math.MinInt64),
		},

		[]string{
			"",
			"6.9",
			"NaN",
		},
		&f,
		func(a string, b int64) bool { return a == strconv.FormatInt(b, 10) },
	); err != nil {
		t.Fatal(err)
	}
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


func Test_parseFrom(t *testing.T) {
	var c ConfigSet

	cfg := make(map[string]interface{})
	var opts [20]*string
	var vals [20]string

	for i := range 20 {
		name := fmt.Sprintf("option%v", i)

		rd := randString(14)
		cfg[name] = rd

		vals[i] = rd
		opts[i] = c.String(name, "")
	}

	jason, _ := json.Marshal(cfg)
	c.ParseFromData([]byte(jason))

	for i, p := range opts {
		if *p != vals[i] {
			t.Fatalf("Option set to wrong value, expected %v | got %v", vals[i], *p)
		}
	}
}

func Test_parseFile(t *testing.T) {
	var c ConfigSet

	cfg := make(map[string]interface{})
	var opts [20]*string
	var vals [20]string

	for i := range 20 {
		name := fmt.Sprintf("option%v", i)

		rd := randString(14)
		cfg[name] = rd

		vals[i] = rd
		opts[i] = c.String(name, "")
	}

	jason, _ := json.Marshal(cfg)

	f, err := os.CreateTemp("", "test-file")
	if err != nil {
		t.Fatal(err)
	}

	_, err = f.Write(jason)
	if err != nil {
		t.Fatal(err)
	}

	c.Location = f.Name()
	c.Parse()

	for i, p := range opts {
		if *p != vals[i] {
			t.Fatalf("Option set to wrong value, expected %v | got %v", vals[i], *p)
		}
	}
}
*/
