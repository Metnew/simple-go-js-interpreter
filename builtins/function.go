package builtins

import (
	"fmt"

	"github.com/example/jsgo/runtime"
)

var FunctionPrototype *runtime.Object

func createFunctionConstructor(objProto *runtime.Object) (*runtime.Object, *runtime.Object) {
	proto := runtime.NewOrdinaryObject(objProto)
	proto.OType = runtime.ObjTypeFunction
	FunctionPrototype = proto

	setMethod(proto, "call", 1, functionCall)
	setMethod(proto, "apply", 2, functionApply)
	setMethod(proto, "bind", 1, functionBind)
	setMethod(proto, "toString", 0, functionToString)

	ctor := newFuncObject("Function", 1, functionConstructorCall)
	ctor.Constructor = functionConstructorCall
	ctor.Prototype = proto

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor, proto
}

func functionConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return nil, fmt.Errorf("TypeError: Function constructor is not supported")
}

func functionCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	fn := getCallable(this)
	if fn == nil {
		return nil, fmt.Errorf("TypeError: not a function")
	}
	thisArg := argAt(args, 0)
	callArgs := args[1:]
	return fn(thisArg, callArgs)
}

func functionApply(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	fn := getCallable(this)
	if fn == nil {
		return nil, fmt.Errorf("TypeError: not a function")
	}
	thisArg := argAt(args, 0)
	var callArgs []*runtime.Value
	if len(args) > 1 && args[1].Type == runtime.TypeObject && args[1].Object != nil {
		if args[1].Object.OType == runtime.ObjTypeArray {
			callArgs = args[1].Object.ArrayData
		}
	}
	return fn(thisArg, callArgs)
}

func functionBind(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	fn := getCallable(this)
	if fn == nil {
		return nil, fmt.Errorf("TypeError: not a function")
	}
	thisArg := argAt(args, 0)
	boundArgs := make([]*runtime.Value, 0)
	if len(args) > 1 {
		boundArgs = append(boundArgs, args[1:]...)
	}
	boundFn := func(callThis *runtime.Value, callArgs []*runtime.Value) (*runtime.Value, error) {
		allArgs := make([]*runtime.Value, 0, len(boundArgs)+len(callArgs))
		allArgs = append(allArgs, boundArgs...)
		allArgs = append(allArgs, callArgs...)
		return fn(thisArg, allArgs)
	}
	obj := newFuncObject("bound ", 0, boundFn)
	return runtime.NewObject(obj), nil
}

func functionToString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	if this == nil || this.Type != runtime.TypeObject || this.Object == nil {
		return runtime.NewString("function () { [native code] }"), nil
	}
	name := this.Object.Get("name")
	if name != runtime.Undefined {
		return runtime.NewString(fmt.Sprintf("function %s() { [native code] }", name.ToString())), nil
	}
	return runtime.NewString("function () { [native code] }"), nil
}
