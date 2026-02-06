package builtins

import (
	"testing"

	"github.com/example/jsgo/runtime"
)

func setupObject() {
	createObjectConstructor()
}

func TestObjectKeys(t *testing.T) {
	setupObject()
	obj := runtime.NewOrdinaryObject(nil)
	obj.Set("a", runtime.NewNumber(1))
	obj.Set("b", runtime.NewNumber(2))

	result, err := objectKeys(runtime.Undefined, []*runtime.Value{runtime.NewObject(obj)})
	if err != nil {
		t.Fatal(err)
	}
	arr := toObject(result)
	if arr == nil || len(arr.ArrayData) != 2 {
		t.Fatalf("expected 2 keys, got %v", arr)
	}
}

func TestObjectValues(t *testing.T) {
	setupObject()
	obj := runtime.NewOrdinaryObject(nil)
	obj.Set("x", runtime.NewNumber(10))
	obj.Set("y", runtime.NewNumber(20))

	result, err := objectValues(runtime.Undefined, []*runtime.Value{runtime.NewObject(obj)})
	if err != nil {
		t.Fatal(err)
	}
	arr := toObject(result)
	if arr == nil || len(arr.ArrayData) != 2 {
		t.Fatalf("expected 2 values, got %v", arr)
	}
}

func TestObjectAssign(t *testing.T) {
	setupObject()
	target := runtime.NewOrdinaryObject(nil)
	target.Set("a", runtime.NewNumber(1))
	source := runtime.NewOrdinaryObject(nil)
	source.Set("b", runtime.NewNumber(2))
	source.Set("c", runtime.NewNumber(3))

	result, err := objectAssign(runtime.Undefined, []*runtime.Value{
		runtime.NewObject(target),
		runtime.NewObject(source),
	})
	if err != nil {
		t.Fatal(err)
	}
	obj := toObject(result)
	if obj.Get("a").Number != 1 || obj.Get("b").Number != 2 || obj.Get("c").Number != 3 {
		t.Error("Object.assign did not merge properties correctly")
	}
}

func TestObjectFreezeSeal(t *testing.T) {
	setupObject()
	obj := runtime.NewOrdinaryObject(nil)
	obj.Set("x", runtime.NewNumber(42))
	val := runtime.NewObject(obj)

	objectFreeze(runtime.Undefined, []*runtime.Value{val})
	frozen, _ := objectIsFrozen(runtime.Undefined, []*runtime.Value{val})
	if !frozen.Bool {
		t.Error("expected frozen")
	}

	obj2 := runtime.NewOrdinaryObject(nil)
	obj2.Set("y", runtime.NewNumber(1))
	val2 := runtime.NewObject(obj2)
	objectSeal(runtime.Undefined, []*runtime.Value{val2})
	sealed, _ := objectIsSealed(runtime.Undefined, []*runtime.Value{val2})
	if !sealed.Bool {
		t.Error("expected sealed")
	}
}

func TestObjectIs(t *testing.T) {
	tests := []struct {
		a, b *runtime.Value
		want bool
	}{
		{runtime.NaN, runtime.NaN, true},
		{runtime.Zero, runtime.NewNumber(-0.0), true}, // both are 0, but +0/-0 check
		{runtime.NewNumber(1), runtime.NewNumber(1), true},
		{runtime.NewString("a"), runtime.NewString("a"), true},
		{runtime.NewString("a"), runtime.NewString("b"), false},
		{runtime.Null, runtime.Null, true},
		{runtime.Undefined, runtime.Undefined, true},
		{runtime.Null, runtime.Undefined, false},
	}
	for i, tt := range tests {
		result, _ := objectIs(runtime.Undefined, []*runtime.Value{tt.a, tt.b})
		if result.Bool != tt.want {
			t.Errorf("test %d: Object.is(%v, %v) = %v, want %v", i, tt.a, tt.b, result.Bool, tt.want)
		}
	}
}

func TestObjectCreate(t *testing.T) {
	setupObject()
	proto := runtime.NewOrdinaryObject(nil)
	proto.Set("hello", runtime.NewString("world"))
	result, err := objectCreate(runtime.Undefined, []*runtime.Value{runtime.NewObject(proto)})
	if err != nil {
		t.Fatal(err)
	}
	obj := toObject(result)
	if obj.Prototype != proto {
		t.Error("prototype not set correctly")
	}
	if obj.Get("hello").Str != "world" {
		t.Error("prototype lookup failed")
	}
}

func TestObjectHasOwnProperty(t *testing.T) {
	setupObject()
	obj := runtime.NewOrdinaryObject(nil)
	obj.Set("x", runtime.NewNumber(1))
	thisVal := runtime.NewObject(obj)

	result, _ := objectProtoHasOwnProperty(thisVal, []*runtime.Value{runtime.NewString("x")})
	if !result.Bool {
		t.Error("expected true for own property")
	}
	result, _ = objectProtoHasOwnProperty(thisVal, []*runtime.Value{runtime.NewString("y")})
	if result.Bool {
		t.Error("expected false for non-existent property")
	}
}
