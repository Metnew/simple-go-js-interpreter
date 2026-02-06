package builtins

import (
	"fmt"
	"sort"

	"github.com/example/jsgo/runtime"
)

var ObjectPrototype *runtime.Object

func createObjectConstructor() (*runtime.Object, *runtime.Object) {
	proto := runtime.NewOrdinaryObject(nil)
	ObjectPrototype = proto

	// Object.prototype methods
	setMethod(proto, "hasOwnProperty", 1, objectProtoHasOwnProperty)
	setMethod(proto, "toString", 0, objectProtoToString)
	setMethod(proto, "valueOf", 0, objectProtoValueOf)
	setMethod(proto, "isPrototypeOf", 1, objectProtoIsPrototypeOf)
	setMethod(proto, "propertyIsEnumerable", 1, objectProtoPropertyIsEnumerable)

	// Object constructor
	ctor := newFuncObject("Object", 1, objectConstructorCall)
	ctor.Constructor = objectConstructorCall
	ctor.Prototype = proto

	setMethod(ctor, "keys", 1, objectKeys)
	setMethod(ctor, "values", 1, objectValues)
	setMethod(ctor, "entries", 1, objectEntries)
	setMethod(ctor, "assign", 2, objectAssign)
	setMethod(ctor, "create", 2, objectCreate)
	setMethod(ctor, "defineProperty", 3, objectDefineProperty)
	setMethod(ctor, "defineProperties", 2, objectDefineProperties)
	setMethod(ctor, "getOwnPropertyDescriptor", 2, objectGetOwnPropertyDescriptor)
	setMethod(ctor, "getOwnPropertyNames", 1, objectGetOwnPropertyNames)
	setMethod(ctor, "getPrototypeOf", 1, objectGetPrototypeOf)
	setMethod(ctor, "setPrototypeOf", 2, objectSetPrototypeOf)
	setMethod(ctor, "freeze", 1, objectFreeze)
	setMethod(ctor, "seal", 1, objectSeal)
	setMethod(ctor, "isFrozen", 1, objectIsFrozen)
	setMethod(ctor, "isSealed", 1, objectIsSealed)
	setMethod(ctor, "is", 2, objectIs)

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor, proto
}

func objectConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	arg := argAt(args, 0)
	if arg.Type == runtime.TypeUndefined || arg.Type == runtime.TypeNull {
		return runtime.NewObject(runtime.NewOrdinaryObject(ObjectPrototype)), nil
	}
	if arg.Type == runtime.TypeObject {
		return arg, nil
	}
	return runtime.NewObject(runtime.NewOrdinaryObject(ObjectPrototype)), nil
}

func objectProtoHasOwnProperty(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.False, nil
	}
	name := argAt(args, 0).ToString()
	return runtime.NewBool(obj.HasOwnProperty(name)), nil
}

func objectProtoToString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	if this == nil || this.Type == runtime.TypeUndefined {
		return runtime.NewString("[object Undefined]"), nil
	}
	if this.Type == runtime.TypeNull {
		return runtime.NewString("[object Null]"), nil
	}
	tag := "Object"
	if this.Type == runtime.TypeObject && this.Object != nil {
		switch this.Object.OType {
		case runtime.ObjTypeArray:
			tag = "Array"
		case runtime.ObjTypeFunction:
			tag = "Function"
		case runtime.ObjTypeRegExp:
			tag = "RegExp"
		case runtime.ObjTypeError:
			tag = "Error"
		case runtime.ObjTypeMap:
			tag = "Map"
		case runtime.ObjTypeSet:
			tag = "Set"
		}
		if ts := this.Object.Get("@@toStringTag"); ts != runtime.Undefined {
			tag = ts.ToString()
		}
	}
	return runtime.NewString("[object " + tag + "]"), nil
}

func objectProtoValueOf(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	if this == nil {
		return runtime.Undefined, nil
	}
	return this, nil
}

func objectProtoIsPrototypeOf(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.False, nil
	}
	target := toObject(argAt(args, 0))
	if target == nil {
		return runtime.False, nil
	}
	p := target.Prototype
	for p != nil {
		if p == obj {
			return runtime.True, nil
		}
		p = p.Prototype
	}
	return runtime.False, nil
}

func objectProtoPropertyIsEnumerable(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.False, nil
	}
	name := argAt(args, 0).ToString()
	prop, ok := obj.Properties[name]
	if !ok {
		return runtime.False, nil
	}
	return runtime.NewBool(prop.Enumerable), nil
}

func objectKeys(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(argAt(args, 0))
	if obj == nil {
		return runtime.Undefined, fmt.Errorf("TypeError: Object.keys called on non-object")
	}
	keys := getEnumerableOwnKeys(obj)
	return createStringArray(keys), nil
}

func objectValues(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(argAt(args, 0))
	if obj == nil {
		return runtime.Undefined, fmt.Errorf("TypeError: Object.values called on non-object")
	}
	keys := getEnumerableOwnKeys(obj)
	vals := make([]*runtime.Value, len(keys))
	for i, k := range keys {
		vals[i] = obj.Get(k)
	}
	return createValueArray(vals), nil
}

func objectEntries(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(argAt(args, 0))
	if obj == nil {
		return runtime.Undefined, fmt.Errorf("TypeError: Object.entries called on non-object")
	}
	keys := getEnumerableOwnKeys(obj)
	entries := make([]*runtime.Value, len(keys))
	for i, k := range keys {
		pair := createValueArray([]*runtime.Value{runtime.NewString(k), obj.Get(k)})
		entries[i] = pair
	}
	return createValueArray(entries), nil
}

func objectAssign(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	target := toObject(argAt(args, 0))
	if target == nil {
		return runtime.Undefined, fmt.Errorf("TypeError: Object.assign called on non-object")
	}
	for i := 1; i < len(args); i++ {
		src := toObject(args[i])
		if src == nil {
			continue
		}
		for k, p := range src.Properties {
			if p.Enumerable {
				target.Set(k, p.Value)
			}
		}
	}
	return runtime.NewObject(target), nil
}

func objectCreate(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	arg := argAt(args, 0)
	var proto *runtime.Object
	if arg.Type == runtime.TypeObject && arg.Object != nil {
		proto = arg.Object
	} else if arg.Type != runtime.TypeNull {
		return runtime.Undefined, fmt.Errorf("TypeError: Object prototype may only be an Object or null")
	}
	obj := runtime.NewOrdinaryObject(proto)
	if len(args) > 1 && args[1].Type == runtime.TypeObject {
		if err := definePropertiesFromDescriptors(obj, args[1].Object); err != nil {
			return runtime.Undefined, err
		}
	}
	return runtime.NewObject(obj), nil
}

func objectDefineProperty(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	arg0 := argAt(args, 0)
	if arg0.Type != runtime.TypeObject || arg0.Object == nil {
		return runtime.Undefined, fmt.Errorf("TypeError: Object.defineProperty called on non-object")
	}
	obj := arg0.Object
	name := argAt(args, 1).ToString()
	descArg := argAt(args, 2)
	if descArg.Type != runtime.TypeObject || descArg.Object == nil {
		return runtime.Undefined, fmt.Errorf("TypeError: Property description must be an object")
	}
	desc, err := descriptorToProperty(descArg.Object)
	if err != nil {
		return runtime.Undefined, err
	}
	if err := validateDefineOwnProperty(obj, name, desc); err != nil {
		return runtime.Undefined, err
	}
	mergeAndDefineProperty(obj, name, desc)
	return args[0], nil
}

func objectDefineProperties(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	arg0 := argAt(args, 0)
	if arg0.Type != runtime.TypeObject || arg0.Object == nil {
		return runtime.Undefined, fmt.Errorf("TypeError: Object.defineProperties called on non-object")
	}
	obj := arg0.Object
	props := toObject(argAt(args, 1))
	if props == nil {
		return args[0], nil
	}
	if err := definePropertiesFromDescriptors(obj, props); err != nil {
		return runtime.Undefined, err
	}
	return args[0], nil
}

func objectGetOwnPropertyDescriptor(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(argAt(args, 0))
	if obj == nil {
		return runtime.Undefined, nil
	}
	name := argAt(args, 1).ToString()
	prop, ok := obj.Properties[name]
	if !ok {
		return runtime.Undefined, nil
	}
	return propertyToDescriptor(prop), nil
}

func objectGetOwnPropertyNames(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(argAt(args, 0))
	if obj == nil {
		return runtime.Undefined, fmt.Errorf("TypeError: Object.getOwnPropertyNames called on non-object")
	}
	keys := getAllOwnKeys(obj)
	return createStringArray(keys), nil
}

func objectGetPrototypeOf(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(argAt(args, 0))
	if obj == nil {
		return runtime.Null, nil
	}
	if obj.Prototype == nil {
		return runtime.Null, nil
	}
	return runtime.NewObject(obj.Prototype), nil
}

func objectSetPrototypeOf(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(argAt(args, 0))
	if obj == nil {
		return runtime.Undefined, fmt.Errorf("TypeError: Object.setPrototypeOf called on non-object")
	}
	proto := argAt(args, 1)
	if proto.Type == runtime.TypeNull {
		obj.Prototype = nil
	} else if proto.Type == runtime.TypeObject {
		obj.Prototype = proto.Object
	} else {
		return runtime.Undefined, fmt.Errorf("TypeError: Object prototype may only be an Object or null")
	}
	return args[0], nil
}

func objectFreeze(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(argAt(args, 0))
	if obj == nil {
		return argAt(args, 0), nil
	}
	for _, p := range obj.Properties {
		p.Configurable = false
		if !p.IsAccessor {
			p.Writable = false
		}
	}
	if obj.Internal == nil {
		obj.Internal = make(map[string]interface{})
	}
	obj.Internal["frozen"] = true
	obj.Internal["sealed"] = true
	return args[0], nil
}

func objectSeal(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(argAt(args, 0))
	if obj == nil {
		return argAt(args, 0), nil
	}
	for _, p := range obj.Properties {
		p.Configurable = false
	}
	if obj.Internal == nil {
		obj.Internal = make(map[string]interface{})
	}
	obj.Internal["sealed"] = true
	return args[0], nil
}

func objectIsFrozen(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(argAt(args, 0))
	if obj == nil {
		return runtime.True, nil
	}
	if obj.Internal != nil {
		if v, ok := obj.Internal["frozen"]; ok && v.(bool) {
			return runtime.True, nil
		}
	}
	return runtime.False, nil
}

func objectIsSealed(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(argAt(args, 0))
	if obj == nil {
		return runtime.True, nil
	}
	if obj.Internal != nil {
		if v, ok := obj.Internal["sealed"]; ok && v.(bool) {
			return runtime.True, nil
		}
	}
	return runtime.False, nil
}

func objectIs(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	a := argAt(args, 0)
	b := argAt(args, 1)
	return runtime.NewBool(sameValue(a, b)), nil
}

func sameValue(a, b *runtime.Value) bool {
	if a.Type != b.Type {
		return false
	}
	switch a.Type {
	case runtime.TypeUndefined, runtime.TypeNull:
		return true
	case runtime.TypeNumber:
		if isNaN(a.Number) && isNaN(b.Number) {
			return true
		}
		if a.Number == 0 && b.Number == 0 {
			return (1/a.Number > 0) == (1/b.Number > 0)
		}
		return a.Number == b.Number
	case runtime.TypeString:
		return a.Str == b.Str
	case runtime.TypeBoolean:
		return a.Bool == b.Bool
	case runtime.TypeObject:
		return a.Object == b.Object
	}
	return false
}

// helpers

func getEnumerableOwnKeys(obj *runtime.Object) []string {
	keys := make([]string, 0, len(obj.Properties))
	for k, p := range obj.Properties {
		if p.Enumerable {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

func getAllOwnKeys(obj *runtime.Object) []string {
	keys := make([]string, 0, len(obj.Properties))
	for k := range obj.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func createStringArray(strs []string) *runtime.Value {
	arr := &runtime.Object{
		OType:      runtime.ObjTypeArray,
		Properties: make(map[string]*runtime.Property),
		ArrayData:  make([]*runtime.Value, len(strs)),
	}
	for i, s := range strs {
		arr.ArrayData[i] = runtime.NewString(s)
	}
	arr.Set("length", runtime.NewNumber(float64(len(strs))))
	return runtime.NewObject(arr)
}

func createValueArray(vals []*runtime.Value) *runtime.Value {
	arr := &runtime.Object{
		OType:      runtime.ObjTypeArray,
		Properties: make(map[string]*runtime.Property),
		ArrayData:  vals,
	}
	arr.Set("length", runtime.NewNumber(float64(len(vals))))
	return runtime.NewObject(arr)
}

func descriptorToProperty(desc *runtime.Object) (*runtime.Property, error) {
	prop := &runtime.Property{}

	hasValue := desc.HasProperty("value")
	hasWritable := desc.HasProperty("writable")
	hasGet := desc.HasProperty("get")
	hasSet := desc.HasProperty("set")

	// Check for mixed accessor + data descriptor (spec 8.10.5 step 9)
	isAccessorDesc := hasGet || hasSet
	isDataDesc := hasValue || hasWritable
	if isAccessorDesc && isDataDesc {
		return nil, fmt.Errorf("TypeError: Invalid property descriptor. Cannot both specify accessors and a value or writable attribute")
	}

	if hasValue {
		v := desc.Get("value")
		if v == nil {
			v = runtime.Undefined
		}
		prop.Value = v
		prop.HasValue = true
	}
	if hasWritable {
		v := desc.Get("writable")
		if v != nil {
			prop.Writable = v.ToBoolean()
		}
		prop.HasWritable = true
	}
	if desc.HasProperty("enumerable") {
		v := desc.Get("enumerable")
		if v != nil {
			prop.Enumerable = v.ToBoolean()
		}
		prop.HasEnumerable = true
	}
	if desc.HasProperty("configurable") {
		v := desc.Get("configurable")
		if v != nil {
			prop.Configurable = v.ToBoolean()
		}
		prop.HasConfigurable = true
	}
	if hasGet {
		v := desc.Get("get")
		if v == nil {
			v = runtime.Undefined
		}
		if v != runtime.Undefined {
			if v.Type != runtime.TypeObject || v.Object == nil || v.Object.Callable == nil {
				return nil, fmt.Errorf("TypeError: Getter must be a function")
			}
			prop.Getter = v
		}
		prop.IsAccessor = true
		prop.HasGet = true
	}
	if hasSet {
		v := desc.Get("set")
		if v == nil {
			v = runtime.Undefined
		}
		if v != runtime.Undefined {
			if v.Type != runtime.TypeObject || v.Object == nil || v.Object.Callable == nil {
				return nil, fmt.Errorf("TypeError: Setter must be a function")
			}
			prop.Setter = v
		}
		prop.IsAccessor = true
		prop.HasSet = true
	}
	return prop, nil
}

// mergeAndDefineProperty merges a new property descriptor with an existing one
// (if any) and sets the result on the object. Implements ES5 8.12.9 steps 4-12.
func mergeAndDefineProperty(obj *runtime.Object, name string, desc *runtime.Property) {
	current, exists := obj.Properties[name]
	if !exists {
		// New property: fill in defaults for unspecified attributes
		prop := &runtime.Property{}
		if desc.IsAccessor {
			prop.IsAccessor = true
			prop.Getter = desc.Getter
			prop.Setter = desc.Setter
		} else {
			if desc.HasValue {
				prop.Value = desc.Value
			} else {
				prop.Value = runtime.Undefined
			}
			prop.Writable = desc.Writable
		}
		prop.Enumerable = desc.Enumerable
		prop.Configurable = desc.Configurable
		obj.DefineProperty(name, prop)
		return
	}

	// Existing property: merge only the specified attributes from desc into current.
	if desc.IsAccessor && !current.IsAccessor {
		// Converting data to accessor
		current.IsAccessor = true
		current.Value = nil
		current.Writable = false
		if desc.HasGet {
			current.Getter = desc.Getter
		}
		if desc.HasSet {
			current.Setter = desc.Setter
		}
	} else if !desc.IsAccessor && current.IsAccessor && (desc.HasValue || desc.HasWritable) {
		// Converting accessor to data
		current.IsAccessor = false
		current.Getter = nil
		current.Setter = nil
		if desc.HasValue {
			current.Value = desc.Value
		} else {
			current.Value = runtime.Undefined
		}
		if desc.HasWritable {
			current.Writable = desc.Writable
		} else {
			current.Writable = false
		}
	} else if desc.IsAccessor {
		// Both accessor: update only specified get/set
		if desc.HasGet {
			current.Getter = desc.Getter
		}
		if desc.HasSet {
			current.Setter = desc.Setter
		}
	} else {
		// Both data: update only specified value/writable
		if desc.HasValue {
			current.Value = desc.Value
		}
		if desc.HasWritable {
			current.Writable = desc.Writable
		}
	}
	if desc.HasEnumerable {
		current.Enumerable = desc.Enumerable
	}
	if desc.HasConfigurable {
		current.Configurable = desc.Configurable
	}
}

// validateDefineOwnProperty implements the [[DefineOwnProperty]] validation
// per ES5 8.12.9. It checks if redefining a property is allowed.
func validateDefineOwnProperty(obj *runtime.Object, name string, desc *runtime.Property) error {
	current, exists := obj.Properties[name]
	if !exists {
		return nil
	}

	// If current property is configurable, any change is allowed
	if current.Configurable {
		return nil
	}

	// Current is non-configurable.

	// Can't make it configurable (step 7a)
	if desc.HasConfigurable && desc.Configurable {
		return fmt.Errorf("TypeError: Cannot redefine property: %s", name)
	}

	// Check enumerable change on non-configurable (step 7b)
	if desc.HasEnumerable && desc.Enumerable != current.Enumerable {
		return fmt.Errorf("TypeError: Cannot redefine property: %s", name)
	}

	// Generic descriptor (no value/writable/get/set) - allowed on non-configurable
	isGenericDesc := !desc.IsAccessor && !desc.HasValue && !desc.HasWritable

	if !isGenericDesc {
		// If changing between accessor and data descriptor types (step 9)
		if desc.IsAccessor != current.IsAccessor {
			// Changing data to accessor or accessor to data on non-configurable
			if desc.IsAccessor || (desc.HasValue || desc.HasWritable) {
				return fmt.Errorf("TypeError: Cannot redefine property: %s", name)
			}
		}

		if current.IsAccessor && desc.IsAccessor {
			// For accessor properties: can't change get or set on non-configurable (step 11)
			if desc.HasGet && desc.Getter != current.Getter {
				return fmt.Errorf("TypeError: Cannot redefine property: %s", name)
			}
			if desc.HasSet && desc.Setter != current.Setter {
				return fmt.Errorf("TypeError: Cannot redefine property: %s", name)
			}
		} else if !current.IsAccessor && !desc.IsAccessor {
			// For data properties (step 10)
			if !current.Writable {
				// Can't change writable from false to true on non-configurable
				if desc.HasWritable && desc.Writable {
					return fmt.Errorf("TypeError: Cannot redefine property: %s", name)
				}
				// Can't change value on non-writable non-configurable
				if desc.HasValue && !sameValue(desc.Value, current.Value) {
					return fmt.Errorf("TypeError: Cannot redefine property: %s", name)
				}
			}
		}
	}

	return nil
}

func propertyToDescriptor(prop *runtime.Property) *runtime.Value {
	desc := runtime.NewOrdinaryObject(nil)
	if prop.IsAccessor {
		if prop.Getter != nil {
			desc.Set("get", prop.Getter)
		} else {
			desc.Set("get", runtime.Undefined)
		}
		if prop.Setter != nil {
			desc.Set("set", prop.Setter)
		} else {
			desc.Set("set", runtime.Undefined)
		}
	} else {
		desc.Set("value", prop.Value)
		desc.Set("writable", runtime.NewBool(prop.Writable))
	}
	desc.Set("enumerable", runtime.NewBool(prop.Enumerable))
	desc.Set("configurable", runtime.NewBool(prop.Configurable))
	return runtime.NewObject(desc)
}

func definePropertiesFromDescriptors(obj *runtime.Object, descs *runtime.Object) error {
	for k, p := range descs.Properties {
		if p.Enumerable && p.Value != nil && p.Value.Type == runtime.TypeObject {
			prop, err := descriptorToProperty(p.Value.Object)
			if err != nil {
				return err
			}
			if err := validateDefineOwnProperty(obj, k, prop); err != nil {
				return err
			}
			mergeAndDefineProperty(obj, k, prop)
		}
	}
	return nil
}
