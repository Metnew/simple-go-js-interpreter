package builtins

import (
	"testing"

	"github.com/example/jsgo/runtime"
)

func setupError() {
	createObjectConstructor()
	createErrorConstructor(ObjectPrototype)
}

func TestErrorConstructor(t *testing.T) {
	setupError()
	result, err := errorConstructorCall(runtime.Undefined, []*runtime.Value{runtime.NewString("something failed")})
	if err != nil {
		t.Fatal(err)
	}
	obj := toObject(result)
	if obj == nil {
		t.Fatal("expected object")
	}
	if obj.Get("message").Str != "something failed" {
		t.Errorf("message: expected 'something failed', got %q", obj.Get("message").Str)
	}
	if obj.Get("name").Str != "Error" {
		t.Errorf("name: expected 'Error', got %q", obj.Get("name").Str)
	}
}

func TestErrorToString(t *testing.T) {
	setupError()
	obj := &runtime.Object{
		OType:      runtime.ObjTypeError,
		Properties: make(map[string]*runtime.Property),
		Prototype:  ErrorPrototype,
	}
	obj.Set("name", runtime.NewString("TypeError"))
	obj.Set("message", runtime.NewString("not a function"))

	result, _ := errorToString(runtime.NewObject(obj), nil)
	if result.Str != "TypeError: not a function" {
		t.Errorf("toString: expected 'TypeError: not a function', got %q", result.Str)
	}
}

func TestErrorSubtypes(t *testing.T) {
	setupError()
	typeErr := createErrorSubtype("TypeError", ObjectPrototype, ErrorPrototype)
	result, err := typeErr.Callable(runtime.Undefined, []*runtime.Value{runtime.NewString("bad type")})
	if err != nil {
		t.Fatal(err)
	}
	obj := toObject(result)
	if obj.Get("name").Str != "TypeError" {
		t.Errorf("name: expected 'TypeError', got %q", obj.Get("name").Str)
	}
	if obj.Get("message").Str != "bad type" {
		t.Errorf("message: expected 'bad type', got %q", obj.Get("message").Str)
	}
}
