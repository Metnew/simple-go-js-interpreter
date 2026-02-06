package builtins

import (
	"testing"

	"github.com/example/jsgo/runtime"
)

func setupArray() {
	createObjectConstructor()
	createArrayConstructor(ObjectPrototype)
}

func makeTestArray(vals ...float64) *runtime.Value {
	data := make([]*runtime.Value, len(vals))
	for i, v := range vals {
		data[i] = runtime.NewNumber(v)
	}
	return runtime.NewObject(newArray(data))
}

func TestArrayPushPop(t *testing.T) {
	setupArray()
	arr := makeTestArray(1, 2, 3)

	length, _ := arrayPush(arr, []*runtime.Value{runtime.NewNumber(4)})
	if length.Number != 4 {
		t.Errorf("push: expected length 4, got %v", length.Number)
	}

	popped, _ := arrayPop(arr, nil)
	if popped.Number != 4 {
		t.Errorf("pop: expected 4, got %v", popped.Number)
	}
}

func TestArrayShiftUnshift(t *testing.T) {
	setupArray()
	arr := makeTestArray(1, 2, 3)

	shifted, _ := arrayShift(arr, nil)
	if shifted.Number != 1 {
		t.Errorf("shift: expected 1, got %v", shifted.Number)
	}

	length, _ := arrayUnshift(arr, []*runtime.Value{runtime.NewNumber(0)})
	if length.Number != 3 {
		t.Errorf("unshift: expected length 3, got %v", length.Number)
	}
}

func TestArraySlice(t *testing.T) {
	setupArray()
	arr := makeTestArray(1, 2, 3, 4, 5)

	result, _ := arraySlice(arr, []*runtime.Value{runtime.NewNumber(1), runtime.NewNumber(3)})
	data := getArrayData(result)
	if len(data) != 2 || data[0].Number != 2 || data[1].Number != 3 {
		t.Errorf("slice(1,3): expected [2,3], got %v", data)
	}
}

func TestArraySplice(t *testing.T) {
	setupArray()
	arr := makeTestArray(1, 2, 3, 4, 5)

	removed, _ := arraySplice(arr, []*runtime.Value{runtime.NewNumber(1), runtime.NewNumber(2), runtime.NewNumber(10), runtime.NewNumber(20)})
	removedData := getArrayData(removed)
	if len(removedData) != 2 || removedData[0].Number != 2 || removedData[1].Number != 3 {
		t.Errorf("splice removed: expected [2,3], got %v", removedData)
	}
	arrData := getArrayData(arr)
	if len(arrData) != 5 || arrData[1].Number != 10 || arrData[2].Number != 20 {
		t.Error("splice: array not modified correctly")
	}
}

func TestArrayIndexOf(t *testing.T) {
	setupArray()
	arr := makeTestArray(1, 2, 3, 2, 1)

	result, _ := arrayIndexOf(arr, []*runtime.Value{runtime.NewNumber(2)})
	if result.Number != 1 {
		t.Errorf("indexOf(2): expected 1, got %v", result.Number)
	}

	result, _ = arrayLastIndexOf(arr, []*runtime.Value{runtime.NewNumber(2)})
	if result.Number != 3 {
		t.Errorf("lastIndexOf(2): expected 3, got %v", result.Number)
	}

	result, _ = arrayIndexOf(arr, []*runtime.Value{runtime.NewNumber(99)})
	if result.Number != -1 {
		t.Errorf("indexOf(99): expected -1, got %v", result.Number)
	}
}

func TestArrayIncludes(t *testing.T) {
	setupArray()
	arr := makeTestArray(1, 2, 3)

	result, _ := arrayIncludes(arr, []*runtime.Value{runtime.NewNumber(2)})
	if !result.Bool {
		t.Error("includes(2) should be true")
	}
	result, _ = arrayIncludes(arr, []*runtime.Value{runtime.NewNumber(5)})
	if result.Bool {
		t.Error("includes(5) should be false")
	}
}

func TestArrayMap(t *testing.T) {
	setupArray()
	arr := makeTestArray(1, 2, 3)

	double := newFuncObject("double", 1, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		return runtime.NewNumber(args[0].Number * 2), nil
	})

	result, err := arrayMap(arr, []*runtime.Value{runtime.NewObject(double)})
	if err != nil {
		t.Fatal(err)
	}
	data := getArrayData(result)
	if len(data) != 3 || data[0].Number != 2 || data[1].Number != 4 || data[2].Number != 6 {
		t.Error("map: expected [2,4,6]")
	}
}

func TestArrayFilter(t *testing.T) {
	setupArray()
	arr := makeTestArray(1, 2, 3, 4, 5)

	isEven := newFuncObject("isEven", 1, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		return runtime.NewBool(int(args[0].Number)%2 == 0), nil
	})

	result, err := arrayFilter(arr, []*runtime.Value{runtime.NewObject(isEven)})
	if err != nil {
		t.Fatal(err)
	}
	data := getArrayData(result)
	if len(data) != 2 || data[0].Number != 2 || data[1].Number != 4 {
		t.Error("filter: expected [2,4]")
	}
}

func TestArrayReduce(t *testing.T) {
	setupArray()
	arr := makeTestArray(1, 2, 3, 4)

	sum := newFuncObject("sum", 2, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		return runtime.NewNumber(args[0].Number + args[1].Number), nil
	})

	result, err := arrayReduce(arr, []*runtime.Value{runtime.NewObject(sum), runtime.NewNumber(0)})
	if err != nil {
		t.Fatal(err)
	}
	if result.Number != 10 {
		t.Errorf("reduce: expected 10, got %v", result.Number)
	}
}

func TestArrayEvery(t *testing.T) {
	setupArray()
	arr := makeTestArray(2, 4, 6)

	isEven := newFuncObject("isEven", 1, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		return runtime.NewBool(int(args[0].Number)%2 == 0), nil
	})

	result, _ := arrayEvery(arr, []*runtime.Value{runtime.NewObject(isEven)})
	if !result.Bool {
		t.Error("every(isEven) on [2,4,6] should be true")
	}
}

func TestArraySome(t *testing.T) {
	setupArray()
	arr := makeTestArray(1, 3, 4)

	isEven := newFuncObject("isEven", 1, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		return runtime.NewBool(int(args[0].Number)%2 == 0), nil
	})

	result, _ := arraySome(arr, []*runtime.Value{runtime.NewObject(isEven)})
	if !result.Bool {
		t.Error("some(isEven) on [1,3,4] should be true")
	}
}

func TestArraySort(t *testing.T) {
	setupArray()
	arr := makeTestArray(3, 1, 2)

	numSort := newFuncObject("cmp", 2, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		return runtime.NewNumber(args[0].Number - args[1].Number), nil
	})

	arraySort(arr, []*runtime.Value{runtime.NewObject(numSort)})
	data := getArrayData(arr)
	if data[0].Number != 1 || data[1].Number != 2 || data[2].Number != 3 {
		t.Error("sort: expected [1,2,3]")
	}
}

func TestArrayReverse(t *testing.T) {
	setupArray()
	arr := makeTestArray(1, 2, 3)

	arrayReverse(arr, nil)
	data := getArrayData(arr)
	if data[0].Number != 3 || data[1].Number != 2 || data[2].Number != 1 {
		t.Error("reverse: expected [3,2,1]")
	}
}

func TestArrayJoin(t *testing.T) {
	setupArray()
	arr := makeTestArray(1, 2, 3)

	result, _ := arrayJoin(arr, []*runtime.Value{runtime.NewString("-")})
	if result.Str != "1-2-3" {
		t.Errorf("join('-'): expected '1-2-3', got %q", result.Str)
	}
}

func TestArrayConcat(t *testing.T) {
	setupArray()
	arr1 := makeTestArray(1, 2)
	arr2 := makeTestArray(3, 4)

	result, _ := arrayConcat(arr1, []*runtime.Value{arr2})
	data := getArrayData(result)
	if len(data) != 4 {
		t.Errorf("concat: expected 4 elements, got %d", len(data))
	}
}

func TestArrayFind(t *testing.T) {
	setupArray()
	arr := makeTestArray(1, 2, 3, 4)

	findThree := newFuncObject("findThree", 1, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		return runtime.NewBool(args[0].Number == 3), nil
	})

	result, _ := arrayFind(arr, []*runtime.Value{runtime.NewObject(findThree)})
	if result.Number != 3 {
		t.Errorf("find: expected 3, got %v", result.Number)
	}
}

func TestArrayFlat(t *testing.T) {
	setupArray()
	inner := makeTestArray(3, 4)
	arr := runtime.NewObject(newArray([]*runtime.Value{runtime.NewNumber(1), runtime.NewNumber(2), inner}))

	result, _ := arrayFlat(arr, nil)
	data := getArrayData(result)
	if len(data) != 4 {
		t.Errorf("flat: expected 4 elements, got %d", len(data))
	}
}

func TestArrayIsArray(t *testing.T) {
	setupArray()
	arr := makeTestArray(1, 2, 3)
	obj := runtime.NewObject(runtime.NewOrdinaryObject(nil))

	result, _ := arrayIsArray(runtime.Undefined, []*runtime.Value{arr})
	if !result.Bool {
		t.Error("Array.isArray(arr) should be true")
	}
	result, _ = arrayIsArray(runtime.Undefined, []*runtime.Value{obj})
	if result.Bool {
		t.Error("Array.isArray(obj) should be false")
	}
}

func TestArrayFill(t *testing.T) {
	setupArray()
	arr := makeTestArray(1, 2, 3, 4)
	arrayFill(arr, []*runtime.Value{runtime.NewNumber(0), runtime.NewNumber(1), runtime.NewNumber(3)})
	data := getArrayData(arr)
	if data[0].Number != 1 || data[1].Number != 0 || data[2].Number != 0 || data[3].Number != 4 {
		t.Error("fill: expected [1,0,0,4]")
	}
}

func TestArrayIterators(t *testing.T) {
	setupArray()
	arr := makeTestArray(10, 20, 30)

	// keys
	iter, _ := arrayKeys(arr, nil)
	iterObj := toObject(iter)
	nextFn := getCallable(iterObj.Get("next"))

	r1, _ := nextFn(iter, nil)
	if toObject(r1).Get("value").Number != 0 || toObject(r1).Get("done").Bool != false {
		t.Error("keys iterator: expected {value: 0, done: false}")
	}
}
