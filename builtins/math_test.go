package builtins

import (
	"math"
	"testing"

	"github.com/example/jsgo/runtime"
)

func TestMathConstants(t *testing.T) {
	objProto := runtime.NewOrdinaryObject(nil)
	m := createMathObject(objProto)

	pi := m.Get("PI")
	if pi.Number != math.Pi {
		t.Errorf("Math.PI: expected %v, got %v", math.Pi, pi.Number)
	}
	e := m.Get("E")
	if e.Number != math.E {
		t.Errorf("Math.E: expected %v, got %v", math.E, e.Number)
	}
}

func TestMathAbs(t *testing.T) {
	result, _ := mathAbs(nil, []*runtime.Value{runtime.NewNumber(-5)})
	if result.Number != 5 {
		t.Errorf("Math.abs(-5): expected 5, got %v", result.Number)
	}
}

func TestMathFloorCeilRound(t *testing.T) {
	result, _ := mathFloor(nil, []*runtime.Value{runtime.NewNumber(4.7)})
	if result.Number != 4 {
		t.Errorf("Math.floor(4.7): expected 4, got %v", result.Number)
	}
	result, _ = mathCeil(nil, []*runtime.Value{runtime.NewNumber(4.1)})
	if result.Number != 5 {
		t.Errorf("Math.ceil(4.1): expected 5, got %v", result.Number)
	}
	result, _ = mathRound(nil, []*runtime.Value{runtime.NewNumber(4.5)})
	if result.Number != 5 {
		t.Errorf("Math.round(4.5): expected 5, got %v", result.Number)
	}
}

func TestMathMinMax(t *testing.T) {
	result, _ := mathMax(nil, []*runtime.Value{runtime.NewNumber(1), runtime.NewNumber(5), runtime.NewNumber(3)})
	if result.Number != 5 {
		t.Errorf("Math.max(1,5,3): expected 5, got %v", result.Number)
	}
	result, _ = mathMin(nil, []*runtime.Value{runtime.NewNumber(1), runtime.NewNumber(5), runtime.NewNumber(3)})
	if result.Number != 1 {
		t.Errorf("Math.min(1,5,3): expected 1, got %v", result.Number)
	}
}

func TestMathPow(t *testing.T) {
	result, _ := mathPow(nil, []*runtime.Value{runtime.NewNumber(2), runtime.NewNumber(10)})
	if result.Number != 1024 {
		t.Errorf("Math.pow(2,10): expected 1024, got %v", result.Number)
	}
}

func TestMathSqrt(t *testing.T) {
	result, _ := mathSqrt(nil, []*runtime.Value{runtime.NewNumber(144)})
	if result.Number != 12 {
		t.Errorf("Math.sqrt(144): expected 12, got %v", result.Number)
	}
}

func TestMathSign(t *testing.T) {
	tests := []struct {
		in   float64
		want float64
	}{
		{5, 1},
		{-3, -1},
		{0, 0},
	}
	for _, tt := range tests {
		result, _ := mathSign(nil, []*runtime.Value{runtime.NewNumber(tt.in)})
		if result.Number != tt.want {
			t.Errorf("Math.sign(%v): expected %v, got %v", tt.in, tt.want, result.Number)
		}
	}
}

func TestMathTrunc(t *testing.T) {
	result, _ := mathTrunc(nil, []*runtime.Value{runtime.NewNumber(4.9)})
	if result.Number != 4 {
		t.Errorf("Math.trunc(4.9): expected 4, got %v", result.Number)
	}
	result, _ = mathTrunc(nil, []*runtime.Value{runtime.NewNumber(-4.9)})
	if result.Number != -4 {
		t.Errorf("Math.trunc(-4.9): expected -4, got %v", result.Number)
	}
}

func TestMathRandom(t *testing.T) {
	result, _ := mathRandom(nil, nil)
	if result.Number < 0 || result.Number >= 1 {
		t.Errorf("Math.random(): expected [0,1), got %v", result.Number)
	}
}

func TestMathClz32(t *testing.T) {
	result, _ := mathClz32(nil, []*runtime.Value{runtime.NewNumber(1)})
	if result.Number != 31 {
		t.Errorf("Math.clz32(1): expected 31, got %v", result.Number)
	}
	result, _ = mathClz32(nil, []*runtime.Value{runtime.NewNumber(0)})
	if result.Number != 32 {
		t.Errorf("Math.clz32(0): expected 32, got %v", result.Number)
	}
}

func TestMathImul(t *testing.T) {
	result, _ := mathImul(nil, []*runtime.Value{runtime.NewNumber(3), runtime.NewNumber(4)})
	if result.Number != 12 {
		t.Errorf("Math.imul(3,4): expected 12, got %v", result.Number)
	}
}

func TestMathHypot(t *testing.T) {
	result, _ := mathHypot(nil, []*runtime.Value{runtime.NewNumber(3), runtime.NewNumber(4)})
	if result.Number != 5 {
		t.Errorf("Math.hypot(3,4): expected 5, got %v", result.Number)
	}
}
