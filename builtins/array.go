package builtins

import (
	"fmt"
	"sort"
	"strings"

	"github.com/example/jsgo/runtime"
)

var ArrayPrototype *runtime.Object

func createArrayConstructor(objProto *runtime.Object) (*runtime.Object, *runtime.Object) {
	proto := runtime.NewOrdinaryObject(objProto)
	proto.OType = runtime.ObjTypeArray
	proto.ArrayData = []*runtime.Value{}
	ArrayPrototype = proto

	setMethod(proto, "push", 1, arrayPush)
	setMethod(proto, "pop", 0, arrayPop)
	setMethod(proto, "shift", 0, arrayShift)
	setMethod(proto, "unshift", 1, arrayUnshift)
	setMethod(proto, "splice", 2, arraySplice)
	setMethod(proto, "slice", 2, arraySlice)
	setMethod(proto, "concat", 1, arrayConcat)
	setMethod(proto, "indexOf", 1, arrayIndexOf)
	setMethod(proto, "lastIndexOf", 1, arrayLastIndexOf)
	setMethod(proto, "includes", 1, arrayIncludes)
	setMethod(proto, "find", 1, arrayFind)
	setMethod(proto, "findIndex", 1, arrayFindIndex)
	setMethod(proto, "forEach", 1, arrayForEach)
	setMethod(proto, "map", 1, arrayMap)
	setMethod(proto, "filter", 1, arrayFilter)
	setMethod(proto, "reduce", 1, arrayReduce)
	setMethod(proto, "reduceRight", 1, arrayReduceRight)
	setMethod(proto, "every", 1, arrayEvery)
	setMethod(proto, "some", 1, arraySome)
	setMethod(proto, "sort", 1, arraySort)
	setMethod(proto, "reverse", 0, arrayReverse)
	setMethod(proto, "fill", 1, arrayFill)
	setMethod(proto, "copyWithin", 2, arrayCopyWithin)
	setMethod(proto, "join", 1, arrayJoin)
	setMethod(proto, "toString", 0, arrayToString)
	setMethod(proto, "keys", 0, arrayKeys)
	setMethod(proto, "values", 0, arrayValues)
	setMethod(proto, "entries", 0, arrayEntries)
	setMethod(proto, "flat", 0, arrayFlat)
	setMethod(proto, "flatMap", 1, arrayFlatMap)

	ctor := newFuncObject("Array", 1, arrayConstructorCall)
	ctor.Constructor = arrayConstructorCall
	ctor.Prototype = proto

	setMethod(ctor, "isArray", 1, arrayIsArray)
	setMethod(ctor, "from", 1, arrayFrom)
	setMethod(ctor, "of", 0, arrayOf)

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor, proto
}

func newArray(data []*runtime.Value) *runtime.Object {
	arr := &runtime.Object{
		OType:      runtime.ObjTypeArray,
		Properties: make(map[string]*runtime.Property),
		Prototype:  ArrayPrototype,
		ArrayData:  data,
	}
	arr.Set("length", runtime.NewNumber(float64(len(data))))
	return arr
}

func getArrayData(v *runtime.Value) []*runtime.Value {
	obj := toObject(v)
	if obj == nil {
		return nil
	}
	return obj.ArrayData
}

func arrayConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	if len(args) == 1 && args[0].Type == runtime.TypeNumber {
		n := int(args[0].Number)
		if n < 0 {
			return nil, fmt.Errorf("RangeError: Invalid array length")
		}
		data := make([]*runtime.Value, n)
		for i := range data {
			data[i] = runtime.Undefined
		}
		return runtime.NewObject(newArray(data)), nil
	}
	data := make([]*runtime.Value, len(args))
	copy(data, args)
	return runtime.NewObject(newArray(data)), nil
}

func arrayIsArray(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	a := argAt(args, 0)
	if a.Type == runtime.TypeObject && a.Object != nil && a.Object.OType == runtime.ObjTypeArray {
		return runtime.True, nil
	}
	return runtime.False, nil
}

func arrayFrom(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	src := argAt(args, 0)
	var mapFn runtime.CallableFunc
	if len(args) > 1 && args[1].Type == runtime.TypeObject && args[1].Object != nil && args[1].Object.Callable != nil {
		mapFn = args[1].Object.Callable
	}
	if src.Type == runtime.TypeString {
		runes := []rune(src.Str)
		data := make([]*runtime.Value, len(runes))
		for i, r := range runes {
			val := runtime.NewString(string(r))
			if mapFn != nil {
				v, err := mapFn(runtime.Undefined, []*runtime.Value{val, runtime.NewNumber(float64(i))})
				if err != nil {
					return nil, err
				}
				data[i] = v
			} else {
				data[i] = val
			}
		}
		return runtime.NewObject(newArray(data)), nil
	}
	if src.Type == runtime.TypeObject && src.Object != nil && src.Object.OType == runtime.ObjTypeArray {
		srcData := src.Object.ArrayData
		data := make([]*runtime.Value, len(srcData))
		for i, v := range srcData {
			if mapFn != nil {
				val, err := mapFn(runtime.Undefined, []*runtime.Value{v, runtime.NewNumber(float64(i))})
				if err != nil {
					return nil, err
				}
				data[i] = val
			} else {
				data[i] = v
			}
		}
		return runtime.NewObject(newArray(data)), nil
	}
	return runtime.NewObject(newArray([]*runtime.Value{})), nil
}

func arrayOf(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	data := make([]*runtime.Value, len(args))
	copy(data, args)
	return runtime.NewObject(newArray(data)), nil
}

func arrayPush(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.Undefined, nil
	}
	obj.ArrayData = append(obj.ArrayData, args...)
	length := float64(len(obj.ArrayData))
	obj.Set("length", runtime.NewNumber(length))
	return runtime.NewNumber(length), nil
}

func arrayPop(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil || len(obj.ArrayData) == 0 {
		return runtime.Undefined, nil
	}
	last := obj.ArrayData[len(obj.ArrayData)-1]
	obj.ArrayData = obj.ArrayData[:len(obj.ArrayData)-1]
	obj.Set("length", runtime.NewNumber(float64(len(obj.ArrayData))))
	return last, nil
}

func arrayShift(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil || len(obj.ArrayData) == 0 {
		return runtime.Undefined, nil
	}
	first := obj.ArrayData[0]
	obj.ArrayData = obj.ArrayData[1:]
	obj.Set("length", runtime.NewNumber(float64(len(obj.ArrayData))))
	return first, nil
}

func arrayUnshift(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.Undefined, nil
	}
	obj.ArrayData = append(args, obj.ArrayData...)
	length := float64(len(obj.ArrayData))
	obj.Set("length", runtime.NewNumber(length))
	return runtime.NewNumber(length), nil
}

func arraySplice(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.NewObject(newArray(nil)), nil
	}
	length := len(obj.ArrayData)
	start := 0
	if len(args) > 0 {
		start = int(toInteger(args[0]))
		if start < 0 {
			start = length + start
			if start < 0 {
				start = 0
			}
		}
		if start > length {
			start = length
		}
	}
	deleteCount := length - start
	if len(args) > 1 {
		deleteCount = int(toInteger(args[1]))
		if deleteCount < 0 {
			deleteCount = 0
		}
		if deleteCount > length-start {
			deleteCount = length - start
		}
	}
	removed := make([]*runtime.Value, deleteCount)
	copy(removed, obj.ArrayData[start:start+deleteCount])
	items := args[2:]
	newData := make([]*runtime.Value, 0, length-deleteCount+len(items))
	newData = append(newData, obj.ArrayData[:start]...)
	newData = append(newData, items...)
	newData = append(newData, obj.ArrayData[start+deleteCount:]...)
	obj.ArrayData = newData
	obj.Set("length", runtime.NewNumber(float64(len(newData))))
	return runtime.NewObject(newArray(removed)), nil
}

func arraySlice(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.NewObject(newArray(nil)), nil
	}
	length := len(obj.ArrayData)
	start := 0
	end := length
	if len(args) > 0 {
		start = int(toInteger(args[0]))
		if start < 0 {
			start = length + start
			if start < 0 {
				start = 0
			}
		}
	}
	if len(args) > 1 && args[1].Type != runtime.TypeUndefined {
		end = int(toInteger(args[1]))
		if end < 0 {
			end = length + end
			if end < 0 {
				end = 0
			}
		}
	}
	if start > length {
		start = length
	}
	if end > length {
		end = length
	}
	if start >= end {
		return runtime.NewObject(newArray([]*runtime.Value{})), nil
	}
	data := make([]*runtime.Value, end-start)
	copy(data, obj.ArrayData[start:end])
	return runtime.NewObject(newArray(data)), nil
}

func arrayConcat(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	var result []*runtime.Value
	if obj != nil {
		result = make([]*runtime.Value, len(obj.ArrayData))
		copy(result, obj.ArrayData)
	}
	for _, a := range args {
		if a.Type == runtime.TypeObject && a.Object != nil && a.Object.OType == runtime.ObjTypeArray {
			result = append(result, a.Object.ArrayData...)
		} else {
			result = append(result, a)
		}
	}
	return runtime.NewObject(newArray(result)), nil
}

func arrayIndexOf(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.NewNumber(-1), nil
	}
	search := argAt(args, 0)
	from := 0
	if len(args) > 1 {
		from = int(toInteger(args[1]))
		if from < 0 {
			from = len(obj.ArrayData) + from
			if from < 0 {
				from = 0
			}
		}
	}
	for i := from; i < len(obj.ArrayData); i++ {
		if strictEquals(obj.ArrayData[i], search) {
			return runtime.NewNumber(float64(i)), nil
		}
	}
	return runtime.NewNumber(-1), nil
}

func arrayLastIndexOf(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.NewNumber(-1), nil
	}
	search := argAt(args, 0)
	from := len(obj.ArrayData) - 1
	if len(args) > 1 {
		from = int(toInteger(args[1]))
		if from < 0 {
			from = len(obj.ArrayData) + from
		}
	}
	if from >= len(obj.ArrayData) {
		from = len(obj.ArrayData) - 1
	}
	for i := from; i >= 0; i-- {
		if strictEquals(obj.ArrayData[i], search) {
			return runtime.NewNumber(float64(i)), nil
		}
	}
	return runtime.NewNumber(-1), nil
}

func arrayIncludes(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.False, nil
	}
	search := argAt(args, 0)
	from := 0
	if len(args) > 1 {
		from = int(toInteger(args[1]))
		if from < 0 {
			from = len(obj.ArrayData) + from
			if from < 0 {
				from = 0
			}
		}
	}
	for i := from; i < len(obj.ArrayData); i++ {
		if sameValueZero(obj.ArrayData[i], search) {
			return runtime.True, nil
		}
	}
	return runtime.False, nil
}

func arrayFind(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.Undefined, nil
	}
	cb := getCallable(argAt(args, 0))
	if cb == nil {
		return nil, fmt.Errorf("TypeError: callback is not a function")
	}
	for i, v := range obj.ArrayData {
		result, err := cb(this, []*runtime.Value{v, runtime.NewNumber(float64(i)), this})
		if err != nil {
			return nil, err
		}
		if result.ToBoolean() {
			return v, nil
		}
	}
	return runtime.Undefined, nil
}

func arrayFindIndex(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.NewNumber(-1), nil
	}
	cb := getCallable(argAt(args, 0))
	if cb == nil {
		return nil, fmt.Errorf("TypeError: callback is not a function")
	}
	for i, v := range obj.ArrayData {
		result, err := cb(this, []*runtime.Value{v, runtime.NewNumber(float64(i)), this})
		if err != nil {
			return nil, err
		}
		if result.ToBoolean() {
			return runtime.NewNumber(float64(i)), nil
		}
	}
	return runtime.NewNumber(-1), nil
}

func arrayForEach(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.Undefined, nil
	}
	cb := getCallable(argAt(args, 0))
	if cb == nil {
		return nil, fmt.Errorf("TypeError: callback is not a function")
	}
	for i, v := range obj.ArrayData {
		_, err := cb(this, []*runtime.Value{v, runtime.NewNumber(float64(i)), this})
		if err != nil {
			return nil, err
		}
	}
	return runtime.Undefined, nil
}

func arrayMap(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.NewObject(newArray(nil)), nil
	}
	cb := getCallable(argAt(args, 0))
	if cb == nil {
		return nil, fmt.Errorf("TypeError: callback is not a function")
	}
	result := make([]*runtime.Value, len(obj.ArrayData))
	for i, v := range obj.ArrayData {
		r, err := cb(this, []*runtime.Value{v, runtime.NewNumber(float64(i)), this})
		if err != nil {
			return nil, err
		}
		result[i] = r
	}
	return runtime.NewObject(newArray(result)), nil
}

func arrayFilter(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.NewObject(newArray(nil)), nil
	}
	cb := getCallable(argAt(args, 0))
	if cb == nil {
		return nil, fmt.Errorf("TypeError: callback is not a function")
	}
	var result []*runtime.Value
	for i, v := range obj.ArrayData {
		r, err := cb(this, []*runtime.Value{v, runtime.NewNumber(float64(i)), this})
		if err != nil {
			return nil, err
		}
		if r.ToBoolean() {
			result = append(result, v)
		}
	}
	return runtime.NewObject(newArray(result)), nil
}

func arrayReduce(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil || len(obj.ArrayData) == 0 {
		if len(args) < 2 {
			return nil, fmt.Errorf("TypeError: Reduce of empty array with no initial value")
		}
		return args[1], nil
	}
	cb := getCallable(argAt(args, 0))
	if cb == nil {
		return nil, fmt.Errorf("TypeError: callback is not a function")
	}
	startIdx := 0
	var acc *runtime.Value
	if len(args) > 1 {
		acc = args[1]
	} else {
		if len(obj.ArrayData) == 0 {
			return nil, fmt.Errorf("TypeError: Reduce of empty array with no initial value")
		}
		acc = obj.ArrayData[0]
		startIdx = 1
	}
	for i := startIdx; i < len(obj.ArrayData); i++ {
		r, err := cb(this, []*runtime.Value{acc, obj.ArrayData[i], runtime.NewNumber(float64(i)), this})
		if err != nil {
			return nil, err
		}
		acc = r
	}
	return acc, nil
}

func arrayReduceRight(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil || len(obj.ArrayData) == 0 {
		if len(args) < 2 {
			return nil, fmt.Errorf("TypeError: Reduce of empty array with no initial value")
		}
		return args[1], nil
	}
	cb := getCallable(argAt(args, 0))
	if cb == nil {
		return nil, fmt.Errorf("TypeError: callback is not a function")
	}
	startIdx := len(obj.ArrayData) - 1
	var acc *runtime.Value
	if len(args) > 1 {
		acc = args[1]
	} else {
		acc = obj.ArrayData[startIdx]
		startIdx--
	}
	for i := startIdx; i >= 0; i-- {
		r, err := cb(this, []*runtime.Value{acc, obj.ArrayData[i], runtime.NewNumber(float64(i)), this})
		if err != nil {
			return nil, err
		}
		acc = r
	}
	return acc, nil
}

func arrayEvery(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.True, nil
	}
	cb := getCallable(argAt(args, 0))
	if cb == nil {
		return nil, fmt.Errorf("TypeError: callback is not a function")
	}
	for i, v := range obj.ArrayData {
		r, err := cb(this, []*runtime.Value{v, runtime.NewNumber(float64(i)), this})
		if err != nil {
			return nil, err
		}
		if !r.ToBoolean() {
			return runtime.False, nil
		}
	}
	return runtime.True, nil
}

func arraySome(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.False, nil
	}
	cb := getCallable(argAt(args, 0))
	if cb == nil {
		return nil, fmt.Errorf("TypeError: callback is not a function")
	}
	for i, v := range obj.ArrayData {
		r, err := cb(this, []*runtime.Value{v, runtime.NewNumber(float64(i)), this})
		if err != nil {
			return nil, err
		}
		if r.ToBoolean() {
			return runtime.True, nil
		}
	}
	return runtime.False, nil
}

func arraySort(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return this, nil
	}
	compareFn := getCallable(argAt(args, 0))
	sort.SliceStable(obj.ArrayData, func(i, j int) bool {
		a := obj.ArrayData[i]
		b := obj.ArrayData[j]
		if a.Type == runtime.TypeUndefined && b.Type == runtime.TypeUndefined {
			return false
		}
		if a.Type == runtime.TypeUndefined {
			return false
		}
		if b.Type == runtime.TypeUndefined {
			return true
		}
		if compareFn != nil {
			r, _ := compareFn(runtime.Undefined, []*runtime.Value{a, b})
			if r != nil {
				return r.Number < 0
			}
		}
		return a.ToString() < b.ToString()
	})
	return this, nil
}

func arrayReverse(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return this, nil
	}
	for i, j := 0, len(obj.ArrayData)-1; i < j; i, j = i+1, j-1 {
		obj.ArrayData[i], obj.ArrayData[j] = obj.ArrayData[j], obj.ArrayData[i]
	}
	return this, nil
}

func arrayFill(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return this, nil
	}
	val := argAt(args, 0)
	length := len(obj.ArrayData)
	start := 0
	end := length
	if len(args) > 1 {
		start = int(toInteger(args[1]))
		if start < 0 {
			start = length + start
			if start < 0 {
				start = 0
			}
		}
	}
	if len(args) > 2 {
		end = int(toInteger(args[2]))
		if end < 0 {
			end = length + end
			if end < 0 {
				end = 0
			}
		}
	}
	if start > length {
		start = length
	}
	if end > length {
		end = length
	}
	for i := start; i < end; i++ {
		obj.ArrayData[i] = val
	}
	return this, nil
}

func arrayCopyWithin(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return this, nil
	}
	length := len(obj.ArrayData)
	target := int(toInteger(argAt(args, 0)))
	start := int(toInteger(argAt(args, 1)))
	end := length
	if len(args) > 2 && args[2].Type != runtime.TypeUndefined {
		end = int(toInteger(args[2]))
	}
	if target < 0 {
		target = length + target
	}
	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}
	count := end - start
	if count <= 0 {
		return this, nil
	}
	temp := make([]*runtime.Value, count)
	copy(temp, obj.ArrayData[start:start+count])
	for i := 0; i < count && target+i < length; i++ {
		obj.ArrayData[target+i] = temp[i]
	}
	return this, nil
}

func arrayJoin(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.NewString(""), nil
	}
	sep := ","
	if len(args) > 0 && args[0].Type != runtime.TypeUndefined {
		sep = args[0].ToString()
	}
	parts := make([]string, len(obj.ArrayData))
	for i, v := range obj.ArrayData {
		if v == nil || v.Type == runtime.TypeUndefined || v.Type == runtime.TypeNull {
			parts[i] = ""
		} else {
			parts[i] = v.ToString()
		}
	}
	return runtime.NewString(strings.Join(parts, sep)), nil
}

func arrayToString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return arrayJoin(this, nil)
}

func arrayKeys(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.Undefined, nil
	}
	idx := 0
	iter := &runtime.Object{
		OType:      runtime.ObjTypeIterator,
		Properties: make(map[string]*runtime.Property),
		IteratorNext: func() (*runtime.Value, bool) {
			if idx >= len(obj.ArrayData) {
				return runtime.Undefined, true
			}
			v := runtime.NewNumber(float64(idx))
			idx++
			return v, false
		},
	}
	setMethod(iter, "next", 0, makeIteratorNext(iter))
	return runtime.NewObject(iter), nil
}

func arrayValues(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.Undefined, nil
	}
	idx := 0
	iter := &runtime.Object{
		OType:      runtime.ObjTypeIterator,
		Properties: make(map[string]*runtime.Property),
		IteratorNext: func() (*runtime.Value, bool) {
			if idx >= len(obj.ArrayData) {
				return runtime.Undefined, true
			}
			v := obj.ArrayData[idx]
			idx++
			return v, false
		},
	}
	setMethod(iter, "next", 0, makeIteratorNext(iter))
	return runtime.NewObject(iter), nil
}

func arrayEntries(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.Undefined, nil
	}
	idx := 0
	iter := &runtime.Object{
		OType:      runtime.ObjTypeIterator,
		Properties: make(map[string]*runtime.Property),
		IteratorNext: func() (*runtime.Value, bool) {
			if idx >= len(obj.ArrayData) {
				return runtime.Undefined, true
			}
			pair := createValueArray([]*runtime.Value{runtime.NewNumber(float64(idx)), obj.ArrayData[idx]})
			idx++
			return pair, false
		},
	}
	setMethod(iter, "next", 0, makeIteratorNext(iter))
	return runtime.NewObject(iter), nil
}

func arrayFlat(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.NewObject(newArray(nil)), nil
	}
	depth := 1
	if len(args) > 0 && args[0].Type == runtime.TypeNumber {
		depth = int(args[0].Number)
	}
	result := flattenArray(obj.ArrayData, depth)
	return runtime.NewObject(newArray(result)), nil
}

func arrayFlatMap(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.NewObject(newArray(nil)), nil
	}
	cb := getCallable(argAt(args, 0))
	if cb == nil {
		return nil, fmt.Errorf("TypeError: callback is not a function")
	}
	var result []*runtime.Value
	for i, v := range obj.ArrayData {
		r, err := cb(this, []*runtime.Value{v, runtime.NewNumber(float64(i)), this})
		if err != nil {
			return nil, err
		}
		if r.Type == runtime.TypeObject && r.Object != nil && r.Object.OType == runtime.ObjTypeArray {
			result = append(result, r.Object.ArrayData...)
		} else {
			result = append(result, r)
		}
	}
	return runtime.NewObject(newArray(result)), nil
}

// helpers

func makeIteratorNext(iter *runtime.Object) runtime.CallableFunc {
	return func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		val, done := iter.IteratorNext()
		result := runtime.NewOrdinaryObject(nil)
		result.Set("value", val)
		result.Set("done", runtime.NewBool(done))
		return runtime.NewObject(result), nil
	}
}

func flattenArray(data []*runtime.Value, depth int) []*runtime.Value {
	var result []*runtime.Value
	for _, v := range data {
		if depth > 0 && v.Type == runtime.TypeObject && v.Object != nil && v.Object.OType == runtime.ObjTypeArray {
			result = append(result, flattenArray(v.Object.ArrayData, depth-1)...)
		} else {
			result = append(result, v)
		}
	}
	return result
}

func strictEquals(a, b *runtime.Value) bool {
	if a.Type != b.Type {
		return false
	}
	switch a.Type {
	case runtime.TypeUndefined, runtime.TypeNull:
		return true
	case runtime.TypeNumber:
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

func sameValueZero(a, b *runtime.Value) bool {
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

func getCallable(v *runtime.Value) runtime.CallableFunc {
	if v != nil && v.Type == runtime.TypeObject && v.Object != nil && v.Object.Callable != nil {
		return v.Object.Callable
	}
	return nil
}
