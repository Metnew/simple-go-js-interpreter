package builtins

import (
	"testing"

	"github.com/example/jsgo/runtime"
)

func TestFunctionCall(t *testing.T) {
	fn := newFuncObject("add", 2, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		return runtime.NewNumber(args[0].Number + args[1].Number), nil
	})
	fnVal := runtime.NewObject(fn)

	result, err := functionCall(fnVal, []*runtime.Value{runtime.Undefined, runtime.NewNumber(3), runtime.NewNumber(4)})
	if err != nil {
		t.Fatal(err)
	}
	if result.Number != 7 {
		t.Errorf("call: expected 7, got %v", result.Number)
	}
}

func TestFunctionApply(t *testing.T) {
	setupArray()
	fn := newFuncObject("sum", 2, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		total := 0.0
		for _, a := range args {
			total += a.Number
		}
		return runtime.NewNumber(total), nil
	})
	fnVal := runtime.NewObject(fn)
	argsArr := newArray([]*runtime.Value{runtime.NewNumber(1), runtime.NewNumber(2), runtime.NewNumber(3)})

	result, err := functionApply(fnVal, []*runtime.Value{runtime.Undefined, runtime.NewObject(argsArr)})
	if err != nil {
		t.Fatal(err)
	}
	if result.Number != 6 {
		t.Errorf("apply: expected 6, got %v", result.Number)
	}
}

func TestFunctionBind(t *testing.T) {
	fn := newFuncObject("multiply", 2, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		return runtime.NewNumber(args[0].Number * args[1].Number), nil
	})
	fnVal := runtime.NewObject(fn)

	bound, err := functionBind(fnVal, []*runtime.Value{runtime.Undefined, runtime.NewNumber(2)})
	if err != nil {
		t.Fatal(err)
	}
	boundFn := getCallable(bound)
	result, err := boundFn(runtime.Undefined, []*runtime.Value{runtime.NewNumber(5)})
	if err != nil {
		t.Fatal(err)
	}
	if result.Number != 10 {
		t.Errorf("bind: expected 10, got %v", result.Number)
	}
}

func TestFunctionToString(t *testing.T) {
	fn := newFuncObject("myFunc", 0, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		return runtime.Undefined, nil
	})
	result, _ := functionToString(runtime.NewObject(fn), nil)
	expected := "function myFunc() { [native code] }"
	if result.Str != expected {
		t.Errorf("toString: expected %q, got %q", expected, result.Str)
	}
}
