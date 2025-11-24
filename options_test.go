package configManager

import "testing"

func Test_stringOpt(t *testing.T) {
	var c ConfigSet

	rd := randString(32)
	if e := optionTester(&c, "foo", rd); e != nil { t.Fatal(e) }

	if dv := c.Lookup("foo").DefValue; dv != rd {
		t.Fatalf("Option default value mismatch, expected: [%v] received: [%v]", rd, dv)
	}

	rd = randString(32)
	c.Set("foo", rd)
	if ov := c.Lookup("foo").Value.Get(); ov != rd {
		t.Fatalf("Option value mismatch, expected: [%v] received: [%v]", rd, ov)
	}
}
