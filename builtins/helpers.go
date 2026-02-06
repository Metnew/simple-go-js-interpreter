package builtins

import (
	"github.com/example/jsgo/runtime"
)

func newFuncObject(name string, length int, fn runtime.CallableFunc) *runtime.Object {
	obj := &runtime.Object{
		OType:      runtime.ObjTypeFunction,
		Properties: make(map[string]*runtime.Property),
		Callable:   fn,
	}
	obj.Set("name", runtime.NewString(name))
	obj.DefineProperty("length", &runtime.Property{
		Value:        runtime.NewNumber(float64(length)),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return obj
}

func setMethod(obj *runtime.Object, name string, length int, fn runtime.CallableFunc) {
	funcObj := newFuncObject(name, length, fn)
	obj.DefineProperty(name, &runtime.Property{
		Value:        runtime.NewObject(funcObj),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})
}

func setDataProp(obj *runtime.Object, name string, val *runtime.Value, writable, enumerable, configurable bool) {
	obj.DefineProperty(name, &runtime.Property{
		Value:        val,
		Writable:     writable,
		Enumerable:   enumerable,
		Configurable: configurable,
	})
}

func setConstant(obj *runtime.Object, name string, val *runtime.Value) {
	setDataProp(obj, name, val, false, false, false)
}

func toObject(v *runtime.Value) *runtime.Object {
	if v != nil && v.Type == runtime.TypeObject && v.Object != nil {
		return v.Object
	}
	return nil
}

func argAt(args []*runtime.Value, i int) *runtime.Value {
	if i < len(args) {
		return args[i]
	}
	return runtime.Undefined
}

func toNumber(v *runtime.Value) float64 {
	if v == nil {
		return 0
	}
	switch v.Type {
	case runtime.TypeUndefined:
		return math_NaN()
	case runtime.TypeNull:
		return 0
	case runtime.TypeBoolean:
		if v.Bool {
			return 1
		}
		return 0
	case runtime.TypeNumber:
		return v.Number
	case runtime.TypeString:
		return parseStringToNumber(v.Str)
	case runtime.TypeObject:
		return math_NaN()
	}
	return math_NaN()
}

func toInteger(v *runtime.Value) float64 {
	n := toNumber(v)
	if isNaN(n) {
		return 0
	}
	if n == 0 || isInf(n, 0) {
		return n
	}
	if n < 0 {
		return -math_Floor(-n)
	}
	return math_Floor(n)
}

func toInt32(v *runtime.Value) int32 {
	n := toNumber(v)
	if isNaN(n) || isInf(n, 0) || n == 0 {
		return 0
	}
	return int32(int64(n))
}

func toUint32(v *runtime.Value) uint32 {
	n := toNumber(v)
	if isNaN(n) || isInf(n, 0) || n == 0 {
		return 0
	}
	return uint32(int64(n))
}
