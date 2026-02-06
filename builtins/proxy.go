package builtins

import (
	"fmt"

	"github.com/example/jsgo/runtime"
)

func createProxyConstructor(objProto *runtime.Object) *runtime.Object {
	ctor := newFuncObject("Proxy", 2, proxyConstructorCall)
	ctor.Constructor = proxyConstructorCall
	return ctor
}

func proxyConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	target := toObject(argAt(args, 0))
	handler := toObject(argAt(args, 1))
	if target == nil || handler == nil {
		return nil, fmt.Errorf("TypeError: Cannot create proxy with a non-object as target or handler")
	}
	proxy := &runtime.Object{
		OType:      runtime.ObjTypeProxy,
		Properties: make(map[string]*runtime.Property),
		Prototype:  target.Prototype,
		Internal: map[string]interface{}{
			"target":  target,
			"handler": handler,
		},
	}
	return runtime.NewObject(proxy), nil
}

func createReflectObject(objProto *runtime.Object) *runtime.Object {
	reflect := runtime.NewOrdinaryObject(objProto)

	setMethod(reflect, "get", 2, reflectGet)
	setMethod(reflect, "set", 3, reflectSet)
	setMethod(reflect, "has", 2, reflectHas)
	setMethod(reflect, "deleteProperty", 2, reflectDeleteProperty)
	setMethod(reflect, "apply", 3, reflectApply)
	setMethod(reflect, "construct", 2, reflectConstruct)
	setMethod(reflect, "ownKeys", 1, reflectOwnKeys)

	reflect.Set("@@toStringTag", runtime.NewString("Reflect"))
	return reflect
}

func reflectGet(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	target := toObject(argAt(args, 0))
	if target == nil {
		return nil, fmt.Errorf("TypeError: Reflect.get requires object target")
	}
	key := argAt(args, 1).ToString()
	return target.Get(key), nil
}

func reflectSet(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	target := toObject(argAt(args, 0))
	if target == nil {
		return nil, fmt.Errorf("TypeError: Reflect.set requires object target")
	}
	key := argAt(args, 1).ToString()
	val := argAt(args, 2)
	target.Set(key, val)
	return runtime.True, nil
}

func reflectHas(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	target := toObject(argAt(args, 0))
	if target == nil {
		return nil, fmt.Errorf("TypeError: Reflect.has requires object target")
	}
	key := argAt(args, 1).ToString()
	return runtime.NewBool(target.HasProperty(key)), nil
}

func reflectDeleteProperty(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	target := toObject(argAt(args, 0))
	if target == nil {
		return nil, fmt.Errorf("TypeError: Reflect.deleteProperty requires object target")
	}
	key := argAt(args, 1).ToString()
	_, ok := target.Properties[key]
	if ok {
		delete(target.Properties, key)
	}
	return runtime.NewBool(ok), nil
}

func reflectApply(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	target := argAt(args, 0)
	fn := getCallable(target)
	if fn == nil {
		return nil, fmt.Errorf("TypeError: Reflect.apply requires callable target")
	}
	thisArg := argAt(args, 1)
	var callArgs []*runtime.Value
	argsArray := toObject(argAt(args, 2))
	if argsArray != nil && argsArray.OType == runtime.ObjTypeArray {
		callArgs = argsArray.ArrayData
	}
	return fn(thisArg, callArgs)
}

func reflectConstruct(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	target := argAt(args, 0)
	targetObj := toObject(target)
	if targetObj == nil || targetObj.Constructor == nil {
		return nil, fmt.Errorf("TypeError: Reflect.construct requires constructor target")
	}
	var ctorArgs []*runtime.Value
	argsArray := toObject(argAt(args, 1))
	if argsArray != nil && argsArray.OType == runtime.ObjTypeArray {
		ctorArgs = argsArray.ArrayData
	}
	return targetObj.Constructor(runtime.Undefined, ctorArgs)
}

func reflectOwnKeys(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	target := toObject(argAt(args, 0))
	if target == nil {
		return nil, fmt.Errorf("TypeError: Reflect.ownKeys requires object target")
	}
	keys := getAllOwnKeys(target)
	return createStringArray(keys), nil
}
