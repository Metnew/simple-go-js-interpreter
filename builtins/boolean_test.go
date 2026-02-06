package builtins

import (
	"testing"

	"github.com/example/jsgo/runtime"
)

func TestBooleanConstructor(t *testing.T) {
	result, _ := booleanConstructorCall(runtime.Undefined, []*runtime.Value{runtime.NewNumber(0)})
	if result.Bool {
		t.Error("Boolean(0) should be false")
	}
	result, _ = booleanConstructorCall(runtime.Undefined, []*runtime.Value{runtime.NewNumber(1)})
	if !result.Bool {
		t.Error("Boolean(1) should be true")
	}
	result, _ = booleanConstructorCall(runtime.Undefined, []*runtime.Value{runtime.NewString("")})
	if result.Bool {
		t.Error("Boolean('') should be false")
	}
	result, _ = booleanConstructorCall(runtime.Undefined, []*runtime.Value{runtime.NewString("x")})
	if !result.Bool {
		t.Error("Boolean('x') should be true")
	}
	result, _ = booleanConstructorCall(runtime.Undefined, nil)
	if result.Bool {
		t.Error("Boolean() should be false")
	}
}

func TestBooleanToString(t *testing.T) {
	result, _ := booleanToString(runtime.True, nil)
	if result.Str != "true" {
		t.Errorf("true.toString(): expected 'true', got %q", result.Str)
	}
	result, _ = booleanToString(runtime.False, nil)
	if result.Str != "false" {
		t.Errorf("false.toString(): expected 'false', got %q", result.Str)
	}
}
