package configManager

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

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
