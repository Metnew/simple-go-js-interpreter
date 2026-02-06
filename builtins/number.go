package builtins

import (
	"fmt"
	"math"
	"strconv"

	"github.com/example/jsgo/runtime"
)

var NumberPrototype *runtime.Object

func createNumberConstructor(objProto *runtime.Object) (*runtime.Object, *runtime.Object) {
	proto := runtime.NewOrdinaryObject(objProto)
	NumberPrototype = proto

	setMethod(proto, "toFixed", 1, numberToFixed)
	setMethod(proto, "toPrecision", 1, numberToPrecision)
	setMethod(proto, "toExponential", 1, numberToExponential)
	setMethod(proto, "toString", 1, numberToString)
	setMethod(proto, "valueOf", 0, numberValueOf)

	ctor := newFuncObject("Number", 1, numberConstructorCall)
	ctor.Constructor = numberConstructorCall
	ctor.Prototype = proto

	setMethod(ctor, "isInteger", 1, numberIsInteger)
	setMethod(ctor, "isFinite", 1, numberIsFinite)
	setMethod(ctor, "isNaN", 1, numberIsNaN)
	setMethod(ctor, "isSafeInteger", 1, numberIsSafeInteger)
	setMethod(ctor, "parseInt", 2, globalParseInt)
	setMethod(ctor, "parseFloat", 1, globalParseFloat)

	setConstant(ctor, "EPSILON", runtime.NewNumber(math.SmallestNonzeroFloat64*math.Pow(2, 1022)))
	setConstant(ctor, "MAX_SAFE_INTEGER", runtime.NewNumber(9007199254740991))
	setConstant(ctor, "MIN_SAFE_INTEGER", runtime.NewNumber(-9007199254740991))
	setConstant(ctor, "MAX_VALUE", runtime.NewNumber(math.MaxFloat64))
	setConstant(ctor, "MIN_VALUE", runtime.NewNumber(math.SmallestNonzeroFloat64))
	setConstant(ctor, "NaN", runtime.NaN)
	setConstant(ctor, "POSITIVE_INFINITY", runtime.PosInf)
	setConstant(ctor, "NEGATIVE_INFINITY", runtime.NegInf)

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor, proto
}

func getNumberValue(this *runtime.Value) float64 {
	if this == nil {
		return 0
	}
	if this.Type == runtime.TypeNumber {
		return this.Number
	}
	if this.Type == runtime.TypeObject && this.Object != nil {
		if iv, ok := this.Object.Internal["NumberData"]; ok {
			return iv.(float64)
		}
	}
	return toNumber(this)
}

func numberConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	if len(args) == 0 {
		return runtime.Zero, nil
	}
	return runtime.NewNumber(toNumber(args[0])), nil
}

func numberToFixed(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	n := getNumberValue(this)
	digits := 0
	if len(args) > 0 {
		digits = int(toInteger(args[0]))
	}
	if digits < 0 || digits > 100 {
		return nil, fmt.Errorf("RangeError: toFixed() digits argument must be between 0 and 100")
	}
	return runtime.NewString(strconv.FormatFloat(n, 'f', digits, 64)), nil
}

func numberToPrecision(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	n := getNumberValue(this)
	if len(args) == 0 || args[0].Type == runtime.TypeUndefined {
		return runtime.NewString(fmt.Sprintf("%g", n)), nil
	}
	prec := int(toInteger(args[0]))
	if prec < 1 || prec > 100 {
		return nil, fmt.Errorf("RangeError: toPrecision() argument must be between 1 and 100")
	}
	return runtime.NewString(strconv.FormatFloat(n, 'g', prec, 64)), nil
}

func numberToExponential(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	n := getNumberValue(this)
	digits := -1
	if len(args) > 0 && args[0].Type != runtime.TypeUndefined {
		digits = int(toInteger(args[0]))
		if digits < 0 || digits > 100 {
			return nil, fmt.Errorf("RangeError: toExponential() argument must be between 0 and 100")
		}
	}
	if digits < 0 {
		return runtime.NewString(fmt.Sprintf("%e", n)), nil
	}
	return runtime.NewString(strconv.FormatFloat(n, 'e', digits, 64)), nil
}

func numberToString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	n := getNumberValue(this)
	radix := 10
	if len(args) > 0 && args[0].Type != runtime.TypeUndefined {
		radix = int(toInteger(args[0]))
	}
	if radix < 2 || radix > 36 {
		return nil, fmt.Errorf("RangeError: toString() radix must be between 2 and 36")
	}
	if isNaN(n) {
		return runtime.NewString("NaN"), nil
	}
	if isInf(n, 1) {
		return runtime.NewString("Infinity"), nil
	}
	if isInf(n, -1) {
		return runtime.NewString("-Infinity"), nil
	}
	if radix == 10 {
		return runtime.NewString(fmt.Sprintf("%g", n)), nil
	}
	intVal := int64(n)
	return runtime.NewString(strconv.FormatInt(intVal, radix)), nil
}

func numberValueOf(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return runtime.NewNumber(getNumberValue(this)), nil
}

func numberIsInteger(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	a := argAt(args, 0)
	if a.Type != runtime.TypeNumber {
		return runtime.False, nil
	}
	if isNaN(a.Number) || isInf(a.Number, 0) {
		return runtime.False, nil
	}
	return runtime.NewBool(math.Floor(a.Number) == a.Number), nil
}

func numberIsFinite(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	a := argAt(args, 0)
	if a.Type != runtime.TypeNumber {
		return runtime.False, nil
	}
	return runtime.NewBool(!isNaN(a.Number) && !isInf(a.Number, 0)), nil
}

func numberIsNaN(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	a := argAt(args, 0)
	if a.Type != runtime.TypeNumber {
		return runtime.False, nil
	}
	return runtime.NewBool(isNaN(a.Number)), nil
}

func numberIsSafeInteger(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	a := argAt(args, 0)
	if a.Type != runtime.TypeNumber {
		return runtime.False, nil
	}
	if isNaN(a.Number) || isInf(a.Number, 0) {
		return runtime.False, nil
	}
	if math.Floor(a.Number) != a.Number {
		return runtime.False, nil
	}
	return runtime.NewBool(math.Abs(a.Number) <= 9007199254740991), nil
}
