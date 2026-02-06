package builtins

import (
	"fmt"

	"github.com/example/jsgo/runtime"
)

var (
	MapPrototype     *runtime.Object
	SetPrototype     *runtime.Object
	WeakMapPrototype *runtime.Object
	WeakSetPrototype *runtime.Object
)

// mapEntry stores key-value pairs preserving insertion order
type mapEntry struct {
	key   *runtime.Value
	value *runtime.Value
}

func createMapConstructor(objProto *runtime.Object) (*runtime.Object, *runtime.Object) {
	proto := runtime.NewOrdinaryObject(objProto)
	proto.OType = runtime.ObjTypeMap
	MapPrototype = proto

	setMethod(proto, "get", 1, mapGet)
	setMethod(proto, "set", 2, mapSet)
	setMethod(proto, "has", 1, mapHas)
	setMethod(proto, "delete", 1, mapDelete)
	setMethod(proto, "clear", 0, mapClear)
	setMethod(proto, "forEach", 1, mapForEach)
	setMethod(proto, "keys", 0, mapKeys)
	setMethod(proto, "values", 0, mapValues)
	setMethod(proto, "entries", 0, mapEntries)

	ctor := newFuncObject("Map", 0, mapConstructorCall)
	ctor.Constructor = mapConstructorCall
	ctor.Prototype = proto

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor, proto
}

func getMapEntries(obj *runtime.Object) []*mapEntry {
	if obj == nil || obj.Internal == nil {
		return nil
	}
	entries, _ := obj.Internal["entries"].([]*mapEntry)
	return entries
}

func setMapEntries(obj *runtime.Object, entries []*mapEntry) {
	if obj.Internal == nil {
		obj.Internal = make(map[string]interface{})
	}
	obj.Internal["entries"] = entries
}

func findMapEntry(entries []*mapEntry, key *runtime.Value) int {
	for i, e := range entries {
		if sameValueZero(e.key, key) {
			return i
		}
	}
	return -1
}

func mapConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := &runtime.Object{
		OType:      runtime.ObjTypeMap,
		Properties: make(map[string]*runtime.Property),
		Prototype:  MapPrototype,
		Internal:   map[string]interface{}{"entries": []*mapEntry{}},
	}
	obj.Set("size", runtime.NewNumber(0))
	result := runtime.NewObject(obj)
	if len(args) > 0 && args[0].Type == runtime.TypeObject && args[0].Object != nil && args[0].Object.OType == runtime.ObjTypeArray {
		for _, item := range args[0].Object.ArrayData {
			if item.Type == runtime.TypeObject && item.Object != nil && item.Object.OType == runtime.ObjTypeArray && len(item.Object.ArrayData) >= 2 {
				_, _ = mapSet(result, []*runtime.Value{item.Object.ArrayData[0], item.Object.ArrayData[1]})
			}
		}
	}
	return result, nil
}

func mapGet(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	entries := getMapEntries(obj)
	key := argAt(args, 0)
	idx := findMapEntry(entries, key)
	if idx == -1 {
		return runtime.Undefined, nil
	}
	return entries[idx].value, nil
}

func mapSet(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return this, nil
	}
	key := argAt(args, 0)
	value := argAt(args, 1)
	entries := getMapEntries(obj)
	idx := findMapEntry(entries, key)
	if idx >= 0 {
		entries[idx].value = value
	} else {
		entries = append(entries, &mapEntry{key: key, value: value})
		setMapEntries(obj, entries)
	}
	obj.Set("size", runtime.NewNumber(float64(len(entries))))
	return this, nil
}

func mapHas(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	entries := getMapEntries(obj)
	key := argAt(args, 0)
	return runtime.NewBool(findMapEntry(entries, key) >= 0), nil
}

func mapDelete(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.False, nil
	}
	entries := getMapEntries(obj)
	key := argAt(args, 0)
	idx := findMapEntry(entries, key)
	if idx == -1 {
		return runtime.False, nil
	}
	entries = append(entries[:idx], entries[idx+1:]...)
	setMapEntries(obj, entries)
	obj.Set("size", runtime.NewNumber(float64(len(entries))))
	return runtime.True, nil
}

func mapClear(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj != nil {
		setMapEntries(obj, []*mapEntry{})
		obj.Set("size", runtime.NewNumber(0))
	}
	return runtime.Undefined, nil
}

func mapForEach(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	cb := getCallable(argAt(args, 0))
	if cb == nil {
		return nil, fmt.Errorf("TypeError: callback is not a function")
	}
	entries := getMapEntries(obj)
	for _, e := range entries {
		_, err := cb(this, []*runtime.Value{e.value, e.key, this})
		if err != nil {
			return nil, err
		}
	}
	return runtime.Undefined, nil
}

func mapKeys(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	entries := getMapEntries(obj)
	idx := 0
	iter := &runtime.Object{
		OType:      runtime.ObjTypeIterator,
		Properties: make(map[string]*runtime.Property),
		IteratorNext: func() (*runtime.Value, bool) {
			if idx >= len(entries) {
				return runtime.Undefined, true
			}
			v := entries[idx].key
			idx++
			return v, false
		},
	}
	setMethod(iter, "next", 0, makeIteratorNext(iter))
	return runtime.NewObject(iter), nil
}

func mapValues(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	entries := getMapEntries(obj)
	idx := 0
	iter := &runtime.Object{
		OType:      runtime.ObjTypeIterator,
		Properties: make(map[string]*runtime.Property),
		IteratorNext: func() (*runtime.Value, bool) {
			if idx >= len(entries) {
				return runtime.Undefined, true
			}
			v := entries[idx].value
			idx++
			return v, false
		},
	}
	setMethod(iter, "next", 0, makeIteratorNext(iter))
	return runtime.NewObject(iter), nil
}

func mapEntries(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	entries := getMapEntries(obj)
	idx := 0
	iter := &runtime.Object{
		OType:      runtime.ObjTypeIterator,
		Properties: make(map[string]*runtime.Property),
		IteratorNext: func() (*runtime.Value, bool) {
			if idx >= len(entries) {
				return runtime.Undefined, true
			}
			pair := createValueArray([]*runtime.Value{entries[idx].key, entries[idx].value})
			idx++
			return pair, false
		},
	}
	setMethod(iter, "next", 0, makeIteratorNext(iter))
	return runtime.NewObject(iter), nil
}

// --- Set ---

func createSetConstructor(objProto *runtime.Object) (*runtime.Object, *runtime.Object) {
	proto := runtime.NewOrdinaryObject(objProto)
	proto.OType = runtime.ObjTypeSet
	SetPrototype = proto

	setMethod(proto, "add", 1, setAdd)
	setMethod(proto, "has", 1, setHas)
	setMethod(proto, "delete", 1, setDelete)
	setMethod(proto, "clear", 0, setClear)
	setMethod(proto, "forEach", 1, setForEach)
	setMethod(proto, "keys", 0, setValues) // Set.keys === Set.values
	setMethod(proto, "values", 0, setValues)
	setMethod(proto, "entries", 0, setEntries)

	ctor := newFuncObject("Set", 0, setConstructorCall)
	ctor.Constructor = setConstructorCall
	ctor.Prototype = proto

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor, proto
}

func getSetItems(obj *runtime.Object) []*runtime.Value {
	if obj == nil || obj.Internal == nil {
		return nil
	}
	items, _ := obj.Internal["items"].([]*runtime.Value)
	return items
}

func setSetItems(obj *runtime.Object, items []*runtime.Value) {
	if obj.Internal == nil {
		obj.Internal = make(map[string]interface{})
	}
	obj.Internal["items"] = items
}

func findSetItem(items []*runtime.Value, val *runtime.Value) int {
	for i, item := range items {
		if sameValueZero(item, val) {
			return i
		}
	}
	return -1
}

func setConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := &runtime.Object{
		OType:      runtime.ObjTypeSet,
		Properties: make(map[string]*runtime.Property),
		Prototype:  SetPrototype,
		Internal:   map[string]interface{}{"items": []*runtime.Value{}},
	}
	obj.Set("size", runtime.NewNumber(0))
	result := runtime.NewObject(obj)
	if len(args) > 0 && args[0].Type == runtime.TypeObject && args[0].Object != nil && args[0].Object.OType == runtime.ObjTypeArray {
		for _, item := range args[0].Object.ArrayData {
			_, _ = setAdd(result, []*runtime.Value{item})
		}
	}
	return result, nil
}

func setAdd(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return this, nil
	}
	val := argAt(args, 0)
	items := getSetItems(obj)
	if findSetItem(items, val) < 0 {
		items = append(items, val)
		setSetItems(obj, items)
		obj.Set("size", runtime.NewNumber(float64(len(items))))
	}
	return this, nil
}

func setHas(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	items := getSetItems(obj)
	val := argAt(args, 0)
	return runtime.NewBool(findSetItem(items, val) >= 0), nil
}

func setDelete(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.False, nil
	}
	items := getSetItems(obj)
	val := argAt(args, 0)
	idx := findSetItem(items, val)
	if idx < 0 {
		return runtime.False, nil
	}
	items = append(items[:idx], items[idx+1:]...)
	setSetItems(obj, items)
	obj.Set("size", runtime.NewNumber(float64(len(items))))
	return runtime.True, nil
}

func setClear(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj != nil {
		setSetItems(obj, []*runtime.Value{})
		obj.Set("size", runtime.NewNumber(0))
	}
	return runtime.Undefined, nil
}

func setForEach(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	cb := getCallable(argAt(args, 0))
	if cb == nil {
		return nil, fmt.Errorf("TypeError: callback is not a function")
	}
	items := getSetItems(obj)
	for _, item := range items {
		_, err := cb(this, []*runtime.Value{item, item, this})
		if err != nil {
			return nil, err
		}
	}
	return runtime.Undefined, nil
}

func setValues(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	items := getSetItems(obj)
	idx := 0
	iter := &runtime.Object{
		OType:      runtime.ObjTypeIterator,
		Properties: make(map[string]*runtime.Property),
		IteratorNext: func() (*runtime.Value, bool) {
			if idx >= len(items) {
				return runtime.Undefined, true
			}
			v := items[idx]
			idx++
			return v, false
		},
	}
	setMethod(iter, "next", 0, makeIteratorNext(iter))
	return runtime.NewObject(iter), nil
}

func setEntries(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	items := getSetItems(obj)
	idx := 0
	iter := &runtime.Object{
		OType:      runtime.ObjTypeIterator,
		Properties: make(map[string]*runtime.Property),
		IteratorNext: func() (*runtime.Value, bool) {
			if idx >= len(items) {
				return runtime.Undefined, true
			}
			pair := createValueArray([]*runtime.Value{items[idx], items[idx]})
			idx++
			return pair, false
		},
	}
	setMethod(iter, "next", 0, makeIteratorNext(iter))
	return runtime.NewObject(iter), nil
}

// --- WeakMap ---

func createWeakMapConstructor(objProto *runtime.Object) *runtime.Object {
	proto := runtime.NewOrdinaryObject(objProto)
	proto.OType = runtime.ObjTypeWeakMap
	WeakMapPrototype = proto

	setMethod(proto, "get", 1, weakMapGet)
	setMethod(proto, "set", 2, weakMapSet)
	setMethod(proto, "has", 1, weakMapHas)
	setMethod(proto, "delete", 1, weakMapDelete)

	ctor := newFuncObject("WeakMap", 0, weakMapConstructorCall)
	ctor.Constructor = weakMapConstructorCall
	ctor.Prototype = proto

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor
}

// WeakMap uses the same internal map approach but keyed by object pointer
func getWeakMapStore(obj *runtime.Object) map[*runtime.Object]*runtime.Value {
	if obj == nil || obj.Internal == nil {
		return nil
	}
	store, _ := obj.Internal["store"].(map[*runtime.Object]*runtime.Value)
	return store
}

func weakMapConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := &runtime.Object{
		OType:      runtime.ObjTypeWeakMap,
		Properties: make(map[string]*runtime.Property),
		Prototype:  WeakMapPrototype,
		Internal:   map[string]interface{}{"store": make(map[*runtime.Object]*runtime.Value)},
	}
	return runtime.NewObject(obj), nil
}

func weakMapGet(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	store := getWeakMapStore(obj)
	key := argAt(args, 0)
	if key.Type != runtime.TypeObject || key.Object == nil {
		return runtime.Undefined, nil
	}
	if v, ok := store[key.Object]; ok {
		return v, nil
	}
	return runtime.Undefined, nil
}

func weakMapSet(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	store := getWeakMapStore(obj)
	key := argAt(args, 0)
	val := argAt(args, 1)
	if key.Type != runtime.TypeObject || key.Object == nil {
		return nil, fmt.Errorf("TypeError: Invalid value used as weak map key")
	}
	store[key.Object] = val
	return this, nil
}

func weakMapHas(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	store := getWeakMapStore(obj)
	key := argAt(args, 0)
	if key.Type != runtime.TypeObject || key.Object == nil {
		return runtime.False, nil
	}
	_, ok := store[key.Object]
	return runtime.NewBool(ok), nil
}

func weakMapDelete(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	store := getWeakMapStore(obj)
	key := argAt(args, 0)
	if key.Type != runtime.TypeObject || key.Object == nil {
		return runtime.False, nil
	}
	if _, ok := store[key.Object]; ok {
		delete(store, key.Object)
		return runtime.True, nil
	}
	return runtime.False, nil
}

// --- WeakSet ---

func createWeakSetConstructor(objProto *runtime.Object) *runtime.Object {
	proto := runtime.NewOrdinaryObject(objProto)
	proto.OType = runtime.ObjTypeWeakSet
	WeakSetPrototype = proto

	setMethod(proto, "add", 1, weakSetAdd)
	setMethod(proto, "has", 1, weakSetHas)
	setMethod(proto, "delete", 1, weakSetDelete)

	ctor := newFuncObject("WeakSet", 0, weakSetConstructorCall)
	ctor.Constructor = weakSetConstructorCall
	ctor.Prototype = proto

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor
}

func getWeakSetStore(obj *runtime.Object) map[*runtime.Object]struct{} {
	if obj == nil || obj.Internal == nil {
		return nil
	}
	store, _ := obj.Internal["store"].(map[*runtime.Object]struct{})
	return store
}

func weakSetConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := &runtime.Object{
		OType:      runtime.ObjTypeWeakSet,
		Properties: make(map[string]*runtime.Property),
		Prototype:  WeakSetPrototype,
		Internal:   map[string]interface{}{"store": make(map[*runtime.Object]struct{})},
	}
	return runtime.NewObject(obj), nil
}

func weakSetAdd(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	store := getWeakSetStore(obj)
	val := argAt(args, 0)
	if val.Type != runtime.TypeObject || val.Object == nil {
		return nil, fmt.Errorf("TypeError: Invalid value used in weak set")
	}
	store[val.Object] = struct{}{}
	return this, nil
}

func weakSetHas(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	store := getWeakSetStore(obj)
	val := argAt(args, 0)
	if val.Type != runtime.TypeObject || val.Object == nil {
		return runtime.False, nil
	}
	_, ok := store[val.Object]
	return runtime.NewBool(ok), nil
}

func weakSetDelete(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	store := getWeakSetStore(obj)
	val := argAt(args, 0)
	if val.Type != runtime.TypeObject || val.Object == nil {
		return runtime.False, nil
	}
	if _, ok := store[val.Object]; ok {
		delete(store, val.Object)
		return runtime.True, nil
	}
	return runtime.False, nil
}
