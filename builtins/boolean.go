package builtins

import (
	"github.com/example/jsgo/runtime"
)

var BooleanPrototype *runtime.Object

func createBooleanConstructor(objProto *runtime.Object) (*runtime.Object, *runtime.Object) {
	proto := runtime.NewOrdinaryObject(objProto)
	BooleanPrototype = proto

	setMethod(proto, "toString", 0, booleanToString)
	setMethod(proto, "valueOf", 0, booleanValueOf)

	ctor := newFuncObject("Boolean", 1, booleanConstructorCall)
	ctor.Constructor = booleanConstructorCall
	ctor.Prototype = proto

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor, proto
}

func getBoolValue(this *runtime.Value) bool {
	if this == nil {
		return false
	}
	if this.Type == runtime.TypeBoolean {
		return this.Bool
	}
	if this.Type == runtime.TypeObject && this.Object != nil {
		if iv, ok := this.Object.Internal["BooleanData"]; ok {
			return iv.(bool)
		}
	}
	return this.ToBoolean()
}

func booleanConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	if len(args) == 0 {
		return runtime.False, nil
	}
	return runtime.NewBool(args[0].ToBoolean()), nil
}

func booleanToString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	b := getBoolValue(this)
	if b {
		return runtime.NewString("true"), nil
	}
	return runtime.NewString("false"), nil
}

func booleanValueOf(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return runtime.NewBool(getBoolValue(this)), nil
}
