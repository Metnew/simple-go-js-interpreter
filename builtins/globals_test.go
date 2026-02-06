package builtins

import (
	"math"
	"testing"

	"github.com/example/jsgo/runtime"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		input string
		radix int
		want  float64
	}{
		{"42", 10, 42},
		{"0xFF", 16, 255},
		{"10", 2, 2},
		{"10", 8, 8},
		{"  42  ", 10, 42},
		{"-5", 10, -5},
	}
	for _, tt := range tests {
		args := []*runtime.Value{runtime.NewString(tt.input)}
		if tt.radix != 10 {
			args = append(args, runtime.NewNumber(float64(tt.radix)))
		}
		result, _ := globalParseInt(runtime.Undefined, args)
		if result.Number != tt.want {
			t.Errorf("parseInt(%q, %d): expected %v, got %v", tt.input, tt.radix, tt.want, result.Number)
		}
	}
}

func TestParseIntNaN(t *testing.T) {
	result, _ := globalParseInt(runtime.Undefined, []*runtime.Value{runtime.NewString("abc")})
	if !math.IsNaN(result.Number) {
		t.Errorf("parseInt('abc'): expected NaN, got %v", result.Number)
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"3.14", 3.14},
		{"  42  ", 42},
		{"Infinity", math.Inf(1)},
	}
	for _, tt := range tests {
		result, _ := globalParseFloat(runtime.Undefined, []*runtime.Value{runtime.NewString(tt.input)})
		if result.Number != tt.want {
			t.Errorf("parseFloat(%q): expected %v, got %v", tt.input, tt.want, result.Number)
		}
	}
}

func TestGlobalIsNaN(t *testing.T) {
	result, _ := globalIsNaN(runtime.Undefined, []*runtime.Value{runtime.NaN})
	if !result.Bool {
		t.Error("isNaN(NaN) should be true")
	}
	result, _ = globalIsNaN(runtime.Undefined, []*runtime.Value{runtime.NewNumber(42)})
	if result.Bool {
		t.Error("isNaN(42) should be false")
	}
	result, _ = globalIsNaN(runtime.Undefined, []*runtime.Value{runtime.NewString("hello")})
	if !result.Bool {
		t.Error("isNaN('hello') should be true (with coercion)")
	}
}

func TestGlobalIsFinite(t *testing.T) {
	result, _ := globalIsFinite(runtime.Undefined, []*runtime.Value{runtime.NewNumber(42)})
	if !result.Bool {
		t.Error("isFinite(42) should be true")
	}
	result, _ = globalIsFinite(runtime.Undefined, []*runtime.Value{runtime.PosInf})
	if result.Bool {
		t.Error("isFinite(Infinity) should be false")
	}
}

func TestEncodeDecodeURI(t *testing.T) {
	input := "hello world"
	encoded, _ := globalEncodeURIComponent(runtime.Undefined, []*runtime.Value{runtime.NewString(input)})
	if encoded.Str != "hello%20world" {
		t.Errorf("encodeURIComponent(%q): expected 'hello%%20world', got %q", input, encoded.Str)
	}

	decoded, _ := globalDecodeURIComponent(runtime.Undefined, []*runtime.Value{encoded})
	if decoded.Str != input {
		t.Errorf("decodeURIComponent(%q): expected %q, got %q", encoded.Str, input, decoded.Str)
	}
}

func TestEvalThrows(t *testing.T) {
	_, err := globalEval(runtime.Undefined, []*runtime.Value{runtime.NewString("1+1")})
	if err == nil {
		t.Error("eval should throw EvalError")
	}
}
