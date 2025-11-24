package configManager

import (
	"math/rand"
	"strconv"
	"testing"
)

func Test_boolOpt(t *testing.T){
	var c ConfigSet

	rd := rand.Float32() > 0.5
	if e := optionTester(&c, "foo", rd); e != nil { t.Fatal(e) }

	if dv := c.Lookup("foo").DefValue; dv != strconv.FormatBool(rd) {
		t.Fatalf("Option default value mismatch, expected: [%v] received: [%v]", rd, dv)
	}

	rd = rand.Float32() > 0.5
	c.Set("foo", strconv.FormatBool(rd))
	if ov := c.Lookup("foo").Value.Get(); ov != rd {
		t.Fatalf("Option value mismatch, expected: [%v] received: [%v]", rd, ov)
	}
}

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

func Test_int32Opt(t *testing.T) {
	var c ConfigSet

	rd := rand.Int31()
	if e := optionTester(&c, "foo", rd); e != nil { t.Fatal(e) }

	if dv := c.Lookup("foo").DefValue; dv != strconv.FormatInt(int64(rd), 10) {
		t.Fatalf("Option default value mismatch, expected: [%v] received: [%v]", rd, dv)
	}

	rd = int32(rand.Int31())
	c.Set("foo", strconv.FormatInt(int64(rd), 10))
	if ov := c.Lookup("foo").Value.Get(); ov != rd {
		t.Fatalf("Option value mismatch, expected: [%v] received: [%v]", rd, ov)
	}
}

func Test_int64Opt(t *testing.T) {
	var c ConfigSet

	rd := rand.Int63()
	if e := optionTester(&c, "foo", rd); e != nil { t.Fatal(e) }

	if dv := c.Lookup("foo").DefValue; dv != strconv.FormatInt(rd, 10) {
		t.Fatalf("Option default value mismatch, expected: [%v] received: [%v]", rd, dv)
	}

	rd = rand.Int63()
	c.Set("foo", strconv.FormatInt(rd, 10))
	if ov := c.Lookup("foo").Value.Get(); ov != rd {
		t.Fatalf("Option value mismatch, expected: [%v] received: [%v]", rd, ov)
	}
}

func Test_float32Opt(t *testing.T) {
	var c ConfigSet

	rd := rand.Float32()
	if e := optionTester(&c, "foo", rd); e != nil { t.Fatal(e) }

	if dv := c.Lookup("foo").DefValue; dv != strconv.FormatFloat(float64(rd), 'f', -1, 32) {
		t.Fatalf("Option default value mismatch, expected: [%v] received: [%v]", rd, dv)
	}

	rd = rand.Float32()
	c.Set("foo", strconv.FormatFloat(float64(rd), 'f', -1, 32))
	if ov := c.Lookup("foo").Value.Get(); ov != rd {
		t.Fatalf("Option value mismatch, expected: [%v] received: [%v]", rd, ov)
	}
}

func Test_float64Opt(t *testing.T) {
	var c ConfigSet

	rd := rand.Float64()
	if e := optionTester(&c, "foo", rd); e != nil { t.Fatal(e) }

	if dv := c.Lookup("foo").DefValue; dv != strconv.FormatFloat(rd, 'f', -1, 64) {
		t.Fatalf("Option default value mismatch, expected: [%v] received: [%v]", rd, dv)
	}

	rd = rand.Float64()
	c.Set("foo", strconv.FormatFloat(rd, 'f', -1, 64))
	if ov := c.Lookup("foo").Value.Get(); ov != rd {
		t.Fatalf("Option value mismatch, expected: [%v] received: [%v]", rd, ov)
	}
}

