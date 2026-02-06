package runtime

import (
	"math"
	"strconv"
	"strings"
)

// ToNumber implements the ECMAScript ToNumber abstract operation.
func (v *Value) ToNumber() float64 {
	switch v.Type {
	case TypeUndefined:
		return math.NaN()
	case TypeNull:
		return 0
	case TypeBoolean:
		if v.Bool {
			return 1
		}
		return 0
	case TypeNumber:
		return v.Number
	case TypeString:
		s := strings.TrimSpace(v.Str)
		if s == "" {
			return 0
		}
		if s == "Infinity" || s == "+Infinity" {
			return math.Inf(1)
		}
		if s == "-Infinity" {
			return math.Inf(-1)
		}
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return math.NaN()
		}
		return n
	case TypeObject:
		return math.NaN()
	default:
		return math.NaN()
	}
}

// StrictEquals implements === comparison.
func StrictEquals(a, b *Value) bool {
	if a.Type != b.Type {
		return false
	}
	switch a.Type {
	case TypeUndefined, TypeNull:
		return true
	case TypeBoolean:
		return a.Bool == b.Bool
	case TypeNumber:
		if math.IsNaN(a.Number) || math.IsNaN(b.Number) {
			return false
		}
		return a.Number == b.Number
	case TypeString:
		return a.Str == b.Str
	case TypeObject:
		return a.Object == b.Object
	default:
		return false
	}
}

// AbstractEquals implements == comparison.
func AbstractEquals(a, b *Value) bool {
	if a.Type == b.Type {
		return StrictEquals(a, b)
	}
	if (a.Type == TypeNull && b.Type == TypeUndefined) ||
		(a.Type == TypeUndefined && b.Type == TypeNull) {
		return true
	}
	if a.Type == TypeNumber && b.Type == TypeString {
		return AbstractEquals(a, NewNumber(b.ToNumber()))
	}
	if a.Type == TypeString && b.Type == TypeNumber {
		return AbstractEquals(NewNumber(a.ToNumber()), b)
	}
	if a.Type == TypeBoolean {
		return AbstractEquals(NewNumber(a.ToNumber()), b)
	}
	if b.Type == TypeBoolean {
		return AbstractEquals(a, NewNumber(b.ToNumber()))
	}
	return false
}

// NewArrayObject creates an array object from values.
func NewArrayObject(proto *Object, elements []*Value) *Object {
	obj := &Object{
		OType:      ObjTypeArray,
		Properties: make(map[string]*Property),
		Prototype:  proto,
		ArrayData:  elements,
	}
	obj.Set("length", NewNumber(float64(len(elements))))
	return obj
}

// NewFunctionObject creates a function object.
func NewFunctionObject(proto *Object, callable CallableFunc) *Object {
	return &Object{
		OType:      ObjTypeFunction,
		Properties: make(map[string]*Property),
		Prototype:  proto,
		Callable:   callable,
	}
}

// NewErrorObject creates an error object with a message.
func NewErrorObject(proto *Object, message string) *Object {
	obj := &Object{
		OType:      ObjTypeError,
		Properties: make(map[string]*Property),
		Prototype:  proto,
	}
	obj.Set("message", NewString(message))
	obj.Set("name", NewString("Error"))
	return obj
}
