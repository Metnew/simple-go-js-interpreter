package builtins

import (
	"testing"

	"github.com/example/jsgo/runtime"
)

func TestReflectGet(t *testing.T) {
	obj := runtime.NewOrdinaryObject(nil)
	obj.Set("x", runtime.NewNumber(42))

	result, err := reflectGet(runtime.Undefined, []*runtime.Value{runtime.NewObject(obj), runtime.NewString("x")})
	if err != nil {
		t.Fatal(err)
	}
	if result.Number != 42 {
		t.Errorf("Reflect.get: expected 42, got %v", result.Number)
	}
}

func TestReflectSet(t *testing.T) {
	obj := runtime.NewOrdinaryObject(nil)
	result, err := reflectSet(runtime.Undefined, []*runtime.Value{runtime.NewObject(obj), runtime.NewString("x"), runtime.NewNumber(10)})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Bool {
		t.Error("Reflect.set should return true")
	}
	if obj.Get("x").Number != 10 {
		t.Error("property not set")
	}
}

func TestReflectHas(t *testing.T) {
	obj := runtime.NewOrdinaryObject(nil)
	obj.Set("a", runtime.NewNumber(1))

	result, _ := reflectHas(runtime.Undefined, []*runtime.Value{runtime.NewObject(obj), runtime.NewString("a")})
	if !result.Bool {
		t.Error("Reflect.has('a') should be true")
	}
	result, _ = reflectHas(runtime.Undefined, []*runtime.Value{runtime.NewObject(obj), runtime.NewString("b")})
	if result.Bool {
		t.Error("Reflect.has('b') should be false")
	}
}

func TestReflectDeleteProperty(t *testing.T) {
	obj := runtime.NewOrdinaryObject(nil)
	obj.Set("x", runtime.NewNumber(1))

	result, _ := reflectDeleteProperty(runtime.Undefined, []*runtime.Value{runtime.NewObject(obj), runtime.NewString("x")})
	if !result.Bool {
		t.Error("Reflect.deleteProperty should return true")
	}
	if obj.HasOwnProperty("x") {
		t.Error("property should be deleted")
	}
}

func TestReflectOwnKeys(t *testing.T) {
	obj := runtime.NewOrdinaryObject(nil)
	obj.Set("a", runtime.NewNumber(1))
	obj.Set("b", runtime.NewNumber(2))

	result, err := reflectOwnKeys(runtime.Undefined, []*runtime.Value{runtime.NewObject(obj)})
	if err != nil {
		t.Fatal(err)
	}
	arr := toObject(result)
	if arr == nil || len(arr.ArrayData) != 2 {
		t.Errorf("Reflect.ownKeys: expected 2 keys, got %v", arr)
	}
}

func TestReflectApply(t *testing.T) {
	fn := newFuncObject("add", 2, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		a := args[0].Number
		b := args[1].Number
		return runtime.NewNumber(a + b), nil
	})
	argsArr := newArray([]*runtime.Value{runtime.NewNumber(3), runtime.NewNumber(4)})

	result, err := reflectApply(runtime.Undefined, []*runtime.Value{runtime.NewObject(fn), runtime.Undefined, runtime.NewObject(argsArr)})
	if err != nil {
		t.Fatal(err)
	}
	if result.Number != 7 {
		t.Errorf("Reflect.apply: expected 7, got %v", result.Number)
	}
}

func TestProxyConstructor(t *testing.T) {
	target := runtime.NewOrdinaryObject(nil)
	target.Set("x", runtime.NewNumber(1))
	handler := runtime.NewOrdinaryObject(nil)

	result, err := proxyConstructorCall(runtime.Undefined, []*runtime.Value{runtime.NewObject(target), runtime.NewObject(handler)})
	if err != nil {
		t.Fatal(err)
	}
	obj := toObject(result)
	if obj == nil || obj.OType != runtime.ObjTypeProxy {
		t.Error("expected proxy object")
	}
}
