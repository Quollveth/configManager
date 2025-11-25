package configManager_test

import (
	"fmt"
	config "github.com/quollveth/configManager"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

type vec3d struct {
	x, y, z float64
}

func parseVec(s string) (vec3d, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 3 {
		return vec3d{}, fmt.Errorf("invalid format: expected 3 comma-separated values")
	}

	floats := make([]float64, 3)
	for i, p := range parts {
		f, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return vec3d{}, fmt.Errorf("invalid float at position %d: %w", i, err)
		}
		floats[i] = f
	}

	return vec3d{floats[0], floats[1], floats[2]}, nil
}

func randomUnitVec() vec3d {
	return vec3d{
		x: rand.Float64(),
		y: rand.Float64(),
		z: rand.Float64(),
	}
}

// vec3d implements Value interface

func (v vec3d) String() string {
	return fmt.Sprintf("%g,%g,%g", v.x, v.y, v.z)
}
func (v vec3d) StringComp() string {
	return fmt.Sprintf("x:%g,y:%g,z:%g", v.x, v.y, v.z)
}

func (v *vec3d) Set(val string) error {
	pv, e := parseVec(val)
	if e != nil {
		return e
	}
	v.x = pv.x
	v.y = pv.y
	v.z = pv.z
	return nil
}

func (v vec3d) Get() any {
	return v
}

func Test_usageExample(t *testing.T) {
	fileloc := "./test_config.json"

	greeting, _ := config.AddOption("greeting", "")
	repeats, _ := config.AddOption("repeats", 0)
	pi, _ := config.AddOption("pi", 0.0)
	doIt, _ := config.AddOption("do the thing", false)

	// since vec3d already implements Value no wrapping is needed
	config.RegisterType(func(t *vec3d) config.Value { return (*vec3d)(t) })
	origin, _ := config.AddOption("local origin", vec3d{})

	config.SetFileLocation(fileloc)
	config.Parse()

	t.Logf("greeting set to %v", *greeting)
	t.Logf("repeats set to %v", *repeats)
	t.Logf("pi set to %v", *pi)
	t.Logf("doIt set to %v", *doIt)
	t.Logf("origin set to %v", origin.StringComp())
}
