package builtins

import (
	"fmt"

	"github.com/example/jsgo/runtime"
)

var ErrorPrototype *runtime.Object

func createErrorConstructor(objProto *runtime.Object) *runtime.Object {
	proto := runtime.NewOrdinaryObject(objProto)
	proto.OType = runtime.ObjTypeError
	ErrorPrototype = proto

	proto.Set("name", runtime.NewString("Error"))
	proto.Set("message", runtime.NewString(""))
	setMethod(proto, "toString", 0, errorToString)

	ctor := newFuncObject("Error", 1, errorConstructorCall)
	ctor.Constructor = errorConstructorCall
	ctor.Prototype = proto

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor
}

func createErrorSubtype(name string, objProto *runtime.Object, errProto *runtime.Object) *runtime.Object {
	proto := runtime.NewOrdinaryObject(errProto)
	proto.OType = runtime.ObjTypeError
	proto.Set("name", runtime.NewString(name))
	proto.Set("message", runtime.NewString(""))

	ctor := newFuncObject(name, 1, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		return makeErrorValue(name, args, proto), nil
	})
	ctor.Constructor = ctor.Callable
	ctor.Prototype = proto

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor
}

func errorConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return makeErrorValue("Error", args, ErrorPrototype), nil
}

func makeErrorValue(name string, args []*runtime.Value, proto *runtime.Object) *runtime.Value {
	obj := &runtime.Object{
		OType:      runtime.ObjTypeError,
		Properties: make(map[string]*runtime.Property),
		Prototype:  proto,
	}
	msg := ""
	if len(args) > 0 && args[0].Type != runtime.TypeUndefined {
		msg = args[0].ToString()
	}
	obj.Set("name", runtime.NewString(name))
	obj.Set("message", runtime.NewString(msg))
	obj.Set("stack", runtime.NewString(fmt.Sprintf("%s: %s", name, msg)))
	return runtime.NewObject(obj)
}

func errorToString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.NewString("Error"), nil
	}
	name := obj.Get("name")
	nameStr := "Error"
	if name != runtime.Undefined {
		nameStr = name.ToString()
	}
	msg := obj.Get("message")
	msgStr := ""
	if msg != runtime.Undefined {
		msgStr = msg.ToString()
	}
	if nameStr == "" {
		return runtime.NewString(msgStr), nil
	}
	if msgStr == "" {
		return runtime.NewString(nameStr), nil
	}
	return runtime.NewString(nameStr + ": " + msgStr), nil
}
