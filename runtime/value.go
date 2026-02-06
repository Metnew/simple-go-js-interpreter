package runtime

import (
	"fmt"
	"math"
)

// ValueType represents the type of a JavaScript value.
type ValueType int

const (
	TypeUndefined ValueType = iota
	TypeNull
	TypeBoolean
	TypeNumber
	TypeString
	TypeObject
	TypeSymbol
)

func (t ValueType) String() string {
	switch t {
	case TypeUndefined:
		return "undefined"
	case TypeNull:
		return "object" // typeof null === "object" in JS
	case TypeBoolean:
		return "boolean"
	case TypeNumber:
		return "number"
	case TypeString:
		return "string"
	case TypeObject:
		return "object"
	case TypeSymbol:
		return "symbol"
	default:
		return "unknown"
	}
}

// Value represents a JavaScript value.
type Value struct {
	Type     ValueType
	Bool     bool
	Number   float64
	Str      string
	Object   *Object
	Symbol   *Symbol
}

var (
	Undefined = &Value{Type: TypeUndefined}
	Null      = &Value{Type: TypeNull}
	True      = &Value{Type: TypeBoolean, Bool: true}
	False     = &Value{Type: TypeBoolean, Bool: false}
	NaN       = &Value{Type: TypeNumber, Number: math_NaN()}
	PosInf    = &Value{Type: TypeNumber, Number: math_Inf(1)}
	NegInf    = &Value{Type: TypeNumber, Number: math_Inf(-1)}
	Zero      = &Value{Type: TypeNumber, Number: 0}
)

func NewNumber(n float64) *Value {
	return &Value{Type: TypeNumber, Number: n}
}

func NewString(s string) *Value {
	return &Value{Type: TypeString, Str: s}
}

func NewBool(b bool) *Value {
	if b {
		return True
	}
	return False
}

func NewObject(obj *Object) *Value {
	return &Value{Type: TypeObject, Object: obj}
}

// ToBoolean implements the ECMAScript ToBoolean abstract operation.
func (v *Value) ToBoolean() bool {
	switch v.Type {
	case TypeUndefined, TypeNull:
		return false
	case TypeBoolean:
		return v.Bool
	case TypeNumber:
		return v.Number != 0 && !isNaN(v.Number)
	case TypeString:
		return len(v.Str) > 0
	case TypeObject:
		return true
	default:
		return false
	}
}

// ToString implements the ECMAScript ToString abstract operation.
func (v *Value) ToString() string {
	switch v.Type {
	case TypeUndefined:
		return "undefined"
	case TypeNull:
		return "null"
	case TypeBoolean:
		if v.Bool {
			return "true"
		}
		return "false"
	case TypeNumber:
		if isNaN(v.Number) {
			return "NaN"
		}
		if isInf(v.Number, 1) {
			return "Infinity"
		}
		if isInf(v.Number, -1) {
			return "-Infinity"
		}
		if v.Number == 0 {
			return "0"
		}
		return fmt.Sprintf("%g", v.Number)
	case TypeString:
		return v.Str
	case TypeObject:
		if v.Object != nil && v.Object.OType == ObjTypeError {
			name := v.Object.Get("name")
			msg := v.Object.Get("message")
			nameStr := "Error"
			if name != nil && name.Type == TypeString && name.Str != "" {
				nameStr = name.Str
			}
			msgStr := ""
			if msg != nil && msg.Type == TypeString {
				msgStr = msg.Str
			}
			if msgStr == "" {
				return nameStr
			}
			return nameStr + ": " + msgStr
		}
		return "[object Object]"
	default:
		return "undefined"
	}
}

// ObjectType describes the kind of object.
type ObjectType int

const (
	ObjTypeOrdinary ObjectType = iota
	ObjTypeArray
	ObjTypeFunction
	ObjTypeRegExp
	ObjTypeDate
	ObjTypeError
	ObjTypeMap
	ObjTypeSet
	ObjTypeWeakMap
	ObjTypeWeakSet
	ObjTypePromise
	ObjTypeIterator
	ObjTypeGenerator
	ObjTypeProxy
)

// Object represents a JavaScript object.
type Object struct {
	OType      ObjectType
	Properties map[string]*Property
	Prototype  *Object
	Callable   CallableFunc
	Constructor CallableFunc
	Internal   map[string]interface{} // internal slots

	// Array-specific
	ArrayData []*Value

	// For iterables
	IteratorNext func() (*Value, bool)
}

// Property represents a property descriptor.
type Property struct {
	Value        *Value
	Getter       *Value // for accessor properties
	Setter       *Value // for accessor properties
	Writable     bool
	Enumerable   bool
	Configurable bool
	IsAccessor   bool
}

// CallableFunc is the Go function signature for JS callable objects.
type CallableFunc func(this *Value, args []*Value) (*Value, error)

// Symbol represents an ES6 Symbol.
type Symbol struct {
	Description string
	id          uint64
}

// NewOrdinaryObject creates a plain object.
func NewOrdinaryObject(proto *Object) *Object {
	return &Object{
		OType:      ObjTypeOrdinary,
		Properties: make(map[string]*Property),
		Prototype:  proto,
	}
}

// Get retrieves a property, walking the prototype chain.
func (o *Object) Get(name string) *Value {
	if prop, ok := o.Properties[name]; ok {
		if prop.IsAccessor && prop.Getter != nil {
			val, _ := prop.Getter.Object.Callable(NewObject(o), nil)
			return val
		}
		return prop.Value
	}
	if o.Prototype != nil {
		return o.Prototype.Get(name)
	}
	return Undefined
}

// Set sets a property value.
func (o *Object) Set(name string, val *Value) {
	if prop, ok := o.Properties[name]; ok {
		if prop.IsAccessor && prop.Setter != nil {
			prop.Setter.Object.Callable(NewObject(o), []*Value{val})
			return
		}
		if prop.Writable {
			prop.Value = val
		}
		return
	}
	o.Properties[name] = &Property{
		Value:        val,
		Writable:     true,
		Enumerable:   true,
		Configurable: true,
	}
}

// DefineProperty defines a property with full descriptor control.
func (o *Object) DefineProperty(name string, prop *Property) {
	o.Properties[name] = prop
}

// HasProperty checks own and prototype chain.
func (o *Object) HasProperty(name string) bool {
	if _, ok := o.Properties[name]; ok {
		return true
	}
	if o.Prototype != nil {
		return o.Prototype.HasProperty(name)
	}
	return false
}

// HasOwnProperty checks only own properties.
func (o *Object) HasOwnProperty(name string) bool {
	_, ok := o.Properties[name]
	return ok
}

func math_NaN() float64              { return math.NaN() }
func math_Inf(sign int) float64      { return math.Inf(sign) }
func isNaN(f float64) bool           { return math.IsNaN(f) }
func isInf(f float64, sign int) bool { return math.IsInf(f, sign) }
