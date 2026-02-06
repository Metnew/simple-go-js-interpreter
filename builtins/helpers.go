package builtins

import (
	"fmt"

	"github.com/example/jsgo/runtime"
)

func newFuncObject(name string, length int, fn runtime.CallableFunc) *runtime.Object {
	obj := &runtime.Object{
		OType:      runtime.ObjTypeFunction,
		Properties: make(map[string]*runtime.Property),
		Callable:   fn,
		Prototype:  FunctionPrototype, // may be nil during early init, fixed by SetFunctionPrototype
	}
	obj.DefineProperty("name", &runtime.Property{
		Value:        runtime.NewString(name),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	obj.DefineProperty("length", &runtime.Property{
		Value:        runtime.NewNumber(float64(length)),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return obj
}

// setFuncPrototypeRecursive walks an object's own properties and sets Prototype
// on any function objects that have nil Prototype. Called after FunctionPrototype is created.
func setFuncPrototypeRecursive(obj *runtime.Object) {
	if obj == nil {
		return
	}
	if obj.OType == runtime.ObjTypeFunction && obj.Prototype == nil {
		obj.Prototype = FunctionPrototype
	}
	for _, p := range obj.Properties {
		if p.Value != nil && p.Value.Type == runtime.TypeObject && p.Value.Object != nil {
			inner := p.Value.Object
			if inner.OType == runtime.ObjTypeFunction && inner.Prototype == nil {
				inner.Prototype = FunctionPrototype
			}
		}
	}
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

// jsToString implements the JavaScript ToString abstract operation for objects,
// calling Symbol.toPrimitive, toString(), and valueOf() methods.
// Returns the string result and an error if any method throws.
func jsToString(v *runtime.Value) (string, error) {
	if v == nil {
		return "undefined", nil
	}
	if v.Type == runtime.TypeSymbol {
		return "", fmt.Errorf("TypeError: Cannot convert a Symbol value to a string")
	}
	if v.Type != runtime.TypeObject || v.Object == nil {
		return v.ToString(), nil
	}
	obj := v.Object
	// Try Symbol.toPrimitive first (if available)
	var toPrim *runtime.Value
	if SymToPrimitive != nil {
		toPrim = obj.GetSymbol(SymToPrimitive)
	}
	if toPrim != nil && toPrim.Type == runtime.TypeObject && toPrim.Object != nil && toPrim.Object.Callable != nil {
		hint := runtime.NewString("string")
		result, err := toPrim.Object.Callable(v, []*runtime.Value{hint})
		if err != nil {
			return "", err
		}
		if result != nil && result.Type != runtime.TypeObject {
			return result.ToString(), nil
		}
		// toPrimitive returned an object - throw TypeError
		return "", fmt.Errorf("TypeError: Cannot convert object to primitive value")
	}
	// Try toString first (hint "string")
	toStr := obj.Get("toString")
	if toStr != nil && toStr.Type == runtime.TypeObject && toStr.Object != nil && toStr.Object.Callable != nil {
		result, err := toStr.Object.Callable(v, nil)
		if err != nil {
			return "", err
		}
		if result != nil && result.Type != runtime.TypeObject {
			return result.ToString(), nil
		}
	}
	// Fall back to valueOf
	valueOf := obj.Get("valueOf")
	if valueOf != nil && valueOf.Type == runtime.TypeObject && valueOf.Object != nil && valueOf.Object.Callable != nil {
		result, err := valueOf.Object.Callable(v, nil)
		if err != nil {
			return "", err
		}
		if result != nil && result.Type != runtime.TypeObject {
			return result.ToString(), nil
		}
	}
	return "", fmt.Errorf("TypeError: Cannot convert object to primitive value")
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
	n, _ := toNumberErr(v)
	return n
}

// toNumberErr implements the JS ToNumber abstract operation with error propagation.
// It calls valueOf()/toString() on objects and throws TypeError for Symbols.
func toNumberErr(v *runtime.Value) (float64, error) {
	if v == nil {
		return 0, nil
	}
	switch v.Type {
	case runtime.TypeUndefined:
		return math_NaN(), nil
	case runtime.TypeNull:
		return 0, nil
	case runtime.TypeBoolean:
		if v.Bool {
			return 1, nil
		}
		return 0, nil
	case runtime.TypeNumber:
		return v.Number, nil
	case runtime.TypeString:
		return parseStringToNumber(v.Str), nil
	case runtime.TypeSymbol:
		return 0, fmt.Errorf("TypeError: Cannot convert a Symbol value to a number")
	case runtime.TypeObject:
		if v.Object != nil {
			// Try valueOf first
			valueOf := v.Object.Get("valueOf")
			if valueOf != nil && valueOf.Type == runtime.TypeObject && valueOf.Object != nil && valueOf.Object.Callable != nil {
				result, err := valueOf.Object.Callable(v, nil)
				if err != nil {
					return 0, err
				}
				if result != nil && result.Type != runtime.TypeObject {
					return toNumberErr(result)
				}
			}
			// Try toString
			toStr := v.Object.Get("toString")
			if toStr != nil && toStr.Type == runtime.TypeObject && toStr.Object != nil && toStr.Object.Callable != nil {
				result, err := toStr.Object.Callable(v, nil)
				if err != nil {
					return 0, err
				}
				if result != nil && result.Type != runtime.TypeObject {
					return toNumberErr(result)
				}
			}
		}
		return math_NaN(), nil
	}
	return math_NaN(), nil
}

// toIntegerErr is like toInteger but propagates errors from ToNumber
func toIntegerErr(v *runtime.Value) (float64, error) {
	n, err := toNumberErr(v)
	if err != nil {
		return 0, err
	}
	if isNaN(n) {
		return 0, nil
	}
	if n == 0 || isInf(n, 0) {
		return n, nil
	}
	if n < 0 {
		return -math_Floor(-n), nil
	}
	return math_Floor(n), nil
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
