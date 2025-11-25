package configManager

/*
import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

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
