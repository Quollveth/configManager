package configManager

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

type point struct {
	x, y float32
}

func (p point) String() string {
	return fmt.Sprintf("(%g,%g)", p.x, p.y)
}

func (p *point) Set(val string) error {
	pp, err := parsePoint(val)
	if err != nil {
		return ErrParse
	}

	p.x = pp.x
	p.y = pp.y

	return nil
}

func (p point) Get() any {
	return point(p)
}

func parsePoint(s string) (point, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "(")
	s = strings.TrimSuffix(s, ")")

	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return point{}, fmt.Errorf("Invalid format")
	}

	x64, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 32)
	if err != nil {
		return point{}, err
	}
	y64, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 32)
	if err != nil {
		return point{}, err
	}

	return point{float32(x64), float32(y64)}, nil
}

func pointEquals(p1, p2 point) bool {
	// shhhh floating point comparisons are not real
	return p1.String() == p2.String()	
}

func Test_customTypeOpt(t *testing.T) {
	var _ Value = &point{}

	var c ConfigSet
	RegisterType(func(t *point) Value { return (*point)(t) })

	rd := point{rand.Float32(), rand.Float32()}

	if e := optionTester(&c, "foo", rd); e != nil {
		t.Fatal(e)
	}

	if dv := c.Lookup("foo").DefValue; rd.String() != dv {
		t.Fatalf("Option default value mismatch, expected: [%v] received: [%v]", rd, dv)
	}

	rd = point{rand.Float32(), rand.Float32()}
	c.Set("foo", rd.String())

	if ov := c.Lookup("foo").Value.Get(); !pointEquals(ov.(point), rd) {
		t.Fatalf("Option value mismatch, expected: [%v] received: [%v]", rd, ov)
	}
}
