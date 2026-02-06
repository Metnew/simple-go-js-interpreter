package builtins

import (
	"math"
	"testing"

	"github.com/example/jsgo/runtime"
)

func TestNumberIsInteger(t *testing.T) {
	tests := []struct {
		val  *runtime.Value
		want bool
	}{
		{runtime.NewNumber(5), true},
		{runtime.NewNumber(5.5), false},
		{runtime.NewNumber(0), true},
		{runtime.NaN, false},
		{runtime.PosInf, false},
		{runtime.NewString("5"), false},
	}
	for i, tt := range tests {
		result, _ := numberIsInteger(runtime.Undefined, []*runtime.Value{tt.val})
		if result.Bool != tt.want {
			t.Errorf("test %d: Number.isInteger(%v) = %v, want %v", i, tt.val, result.Bool, tt.want)
		}
	}
}

func TestNumberIsFinite(t *testing.T) {
	result, _ := numberIsFinite(runtime.Undefined, []*runtime.Value{runtime.NewNumber(42)})
	if !result.Bool {
		t.Error("Number.isFinite(42) should be true")
	}
	result, _ = numberIsFinite(runtime.Undefined, []*runtime.Value{runtime.PosInf})
	if result.Bool {
		t.Error("Number.isFinite(Infinity) should be false")
	}
}

func TestNumberIsNaN(t *testing.T) {
	result, _ := numberIsNaN(runtime.Undefined, []*runtime.Value{runtime.NaN})
	if !result.Bool {
		t.Error("Number.isNaN(NaN) should be true")
	}
	result, _ = numberIsNaN(runtime.Undefined, []*runtime.Value{runtime.NewNumber(42)})
	if result.Bool {
		t.Error("Number.isNaN(42) should be false")
	}
	result, _ = numberIsNaN(runtime.Undefined, []*runtime.Value{runtime.NewString("NaN")})
	if result.Bool {
		t.Error("Number.isNaN('NaN') should be false (no coercion)")
	}
}

func TestNumberIsSafeInteger(t *testing.T) {
	result, _ := numberIsSafeInteger(runtime.Undefined, []*runtime.Value{runtime.NewNumber(9007199254740991)})
	if !result.Bool {
		t.Error("Number.isSafeInteger(MAX_SAFE_INTEGER) should be true")
	}
	result, _ = numberIsSafeInteger(runtime.Undefined, []*runtime.Value{runtime.NewNumber(9007199254740992)})
	if result.Bool {
		t.Error("Number.isSafeInteger(MAX_SAFE_INTEGER+1) should be false")
	}
}

func TestNumberToFixed(t *testing.T) {
	this := runtime.NewNumber(3.14159)
	result, _ := numberToFixed(this, []*runtime.Value{runtime.NewNumber(2)})
	if result.Str != "3.14" {
		t.Errorf("toFixed(2): expected '3.14', got %q", result.Str)
	}
}

func TestNumberToString(t *testing.T) {
	this := runtime.NewNumber(255)
	result, _ := numberToString(this, []*runtime.Value{runtime.NewNumber(16)})
	if result.Str != "ff" {
		t.Errorf("toString(16): expected 'ff', got %q", result.Str)
	}

	this = runtime.NewNumber(10)
	result, _ = numberToString(this, []*runtime.Value{runtime.NewNumber(2)})
	if result.Str != "1010" {
		t.Errorf("toString(2): expected '1010', got %q", result.Str)
	}
}

func TestNumberConstants(t *testing.T) {
	_, objProto := createObjectConstructor()
	ctor, _ := createNumberConstructor(objProto)

	epsilon := ctor.Get("EPSILON")
	if epsilon.Number <= 0 {
		t.Error("EPSILON should be positive")
	}

	maxSafe := ctor.Get("MAX_SAFE_INTEGER")
	if maxSafe.Number != 9007199254740991 {
		t.Errorf("MAX_SAFE_INTEGER: expected 9007199254740991, got %v", maxSafe.Number)
	}

	maxVal := ctor.Get("MAX_VALUE")
	if maxVal.Number != math.MaxFloat64 {
		t.Error("MAX_VALUE should be math.MaxFloat64")
	}
}
