package configManager

import (
	"testing"
)

func Test_parseFrom(t *testing.T) {
	toParse := `{
		"value":69,
		"name":"john golang"
	}`

	var c ConfigSet

	valueOpt, _ := AddOptionToSet(&c, "value", 0)
	nameOpt, _ := AddOptionToSet(&c, "name", "")

	c.ParseFromData([]byte(toParse))

	if *valueOpt != 69 {
		t.Fatalf("Option value mismatch, expected: [69] received: %v", *valueOpt)
	}

	if *nameOpt != "john golang" {
		t.Fatalf("Option value mismatch, expected: [john golang] received: %v", *nameOpt)
	}
}

func Test_parseFile(t *testing.T) {
	fileLoc := "./test_config.json"
	var c ConfigSet
	greeting, _ := AddOptionToSet(&c, "greeting", "")
	c.Location = fileLoc
	c.Parse()
	t.Log(*greeting)
	iz, _ := c.IsZeroValue("greeting")
	if iz {
		t.Fatal("Option set to zero value")
	}
}

func Test_saveTo(t *testing.T){
	fileLoc := "./test_save.json"
	var c ConfigSet
	c.Location = fileLoc

	AddOptionToSet(&c, "greeting", "hello world")
	AddOptionToSet(&c, "repeats", 9)
	AddOptionToSet(&c, "do the thing", false)

	c.Set("greeting", "how ya doin")

	c.Save()
}
