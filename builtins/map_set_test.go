package builtins

import (
	"testing"

	"github.com/example/jsgo/runtime"
)

func setupMapSet() {
	createObjectConstructor()
	createArrayConstructor(ObjectPrototype)
	createMapConstructor(ObjectPrototype)
	createSetConstructor(ObjectPrototype)
}

func TestMapBasic(t *testing.T) {
	setupMapSet()
	m, err := mapConstructorCall(runtime.Undefined, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = mapSet(m, []*runtime.Value{runtime.NewString("key1"), runtime.NewNumber(100)})
	if err != nil {
		t.Fatal(err)
	}

	result, _ := mapGet(m, []*runtime.Value{runtime.NewString("key1")})
	if result.Number != 100 {
		t.Errorf("Map.get('key1'): expected 100, got %v", result.Number)
	}

	has, _ := mapHas(m, []*runtime.Value{runtime.NewString("key1")})
	if !has.Bool {
		t.Error("Map.has('key1') should be true")
	}

	has, _ = mapHas(m, []*runtime.Value{runtime.NewString("key2")})
	if has.Bool {
		t.Error("Map.has('key2') should be false")
	}

	obj := toObject(m)
	if obj.Get("size").Number != 1 {
		t.Errorf("Map.size: expected 1, got %v", obj.Get("size").Number)
	}
}

func TestMapDelete(t *testing.T) {
	setupMapSet()
	m, _ := mapConstructorCall(runtime.Undefined, nil)
	mapSet(m, []*runtime.Value{runtime.NewString("a"), runtime.NewNumber(1)})
	mapSet(m, []*runtime.Value{runtime.NewString("b"), runtime.NewNumber(2)})

	deleted, _ := mapDelete(m, []*runtime.Value{runtime.NewString("a")})
	if !deleted.Bool {
		t.Error("Map.delete should return true")
	}

	has, _ := mapHas(m, []*runtime.Value{runtime.NewString("a")})
	if has.Bool {
		t.Error("Map.has('a') should be false after delete")
	}
	obj := toObject(m)
	if obj.Get("size").Number != 1 {
		t.Errorf("Map.size after delete: expected 1, got %v", obj.Get("size").Number)
	}
}

func TestMapClear(t *testing.T) {
	setupMapSet()
	m, _ := mapConstructorCall(runtime.Undefined, nil)
	mapSet(m, []*runtime.Value{runtime.NewString("x"), runtime.NewNumber(1)})
	mapClear(m, nil)
	obj := toObject(m)
	if obj.Get("size").Number != 0 {
		t.Errorf("Map.size after clear: expected 0, got %v", obj.Get("size").Number)
	}
}

func TestSetBasic(t *testing.T) {
	setupMapSet()
	s, err := setConstructorCall(runtime.Undefined, nil)
	if err != nil {
		t.Fatal(err)
	}

	setAdd(s, []*runtime.Value{runtime.NewNumber(1)})
	setAdd(s, []*runtime.Value{runtime.NewNumber(2)})
	setAdd(s, []*runtime.Value{runtime.NewNumber(1)}) // duplicate

	obj := toObject(s)
	if obj.Get("size").Number != 2 {
		t.Errorf("Set.size: expected 2, got %v", obj.Get("size").Number)
	}

	has, _ := setHas(s, []*runtime.Value{runtime.NewNumber(1)})
	if !has.Bool {
		t.Error("Set.has(1) should be true")
	}

	has, _ = setHas(s, []*runtime.Value{runtime.NewNumber(3)})
	if has.Bool {
		t.Error("Set.has(3) should be false")
	}
}

func TestSetDelete(t *testing.T) {
	setupMapSet()
	s, _ := setConstructorCall(runtime.Undefined, nil)
	setAdd(s, []*runtime.Value{runtime.NewNumber(1)})
	setAdd(s, []*runtime.Value{runtime.NewNumber(2)})

	deleted, _ := setDelete(s, []*runtime.Value{runtime.NewNumber(1)})
	if !deleted.Bool {
		t.Error("Set.delete should return true")
	}

	has, _ := setHas(s, []*runtime.Value{runtime.NewNumber(1)})
	if has.Bool {
		t.Error("Set.has(1) should be false after delete")
	}
}

func TestMapForEach(t *testing.T) {
	setupMapSet()
	m, _ := mapConstructorCall(runtime.Undefined, nil)
	mapSet(m, []*runtime.Value{runtime.NewString("a"), runtime.NewNumber(1)})
	mapSet(m, []*runtime.Value{runtime.NewString("b"), runtime.NewNumber(2)})

	count := 0
	cb := newFuncObject("cb", 3, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		count++
		return runtime.Undefined, nil
	})
	mapForEach(m, []*runtime.Value{runtime.NewObject(cb)})
	if count != 2 {
		t.Errorf("Map.forEach: expected 2 calls, got %d", count)
	}
}

func TestSetForEach(t *testing.T) {
	setupMapSet()
	s, _ := setConstructorCall(runtime.Undefined, nil)
	setAdd(s, []*runtime.Value{runtime.NewNumber(10)})
	setAdd(s, []*runtime.Value{runtime.NewNumber(20)})

	count := 0
	cb := newFuncObject("cb", 3, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		count++
		return runtime.Undefined, nil
	})
	setForEach(s, []*runtime.Value{runtime.NewObject(cb)})
	if count != 2 {
		t.Errorf("Set.forEach: expected 2 calls, got %d", count)
	}
}
