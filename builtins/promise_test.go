package builtins

import (
	"testing"

	"github.com/example/jsgo/runtime"
)

func setupPromise() {
	createObjectConstructor()
	createArrayConstructor(ObjectPrototype)
	createPromiseConstructor(ObjectPrototype)
}

func TestPromiseResolve(t *testing.T) {
	setupPromise()
	result, err := promiseResolve(runtime.Undefined, []*runtime.Value{runtime.NewNumber(42)})
	if err != nil {
		t.Fatal(err)
	}
	obj := toObject(result)
	pd := getPromiseData(obj)
	if pd == nil {
		t.Fatal("expected promise data")
	}
	if pd.state != promiseFulfilled {
		t.Error("expected fulfilled state")
	}
	if pd.result.Number != 42 {
		t.Errorf("expected result 42, got %v", pd.result.Number)
	}
}

func TestPromiseReject(t *testing.T) {
	setupPromise()
	result, err := promiseReject(runtime.Undefined, []*runtime.Value{runtime.NewString("error!")})
	if err != nil {
		t.Fatal(err)
	}
	obj := toObject(result)
	pd := getPromiseData(obj)
	if pd.state != promiseRejected {
		t.Error("expected rejected state")
	}
	if pd.result.Str != "error!" {
		t.Errorf("expected 'error!', got %q", pd.result.Str)
	}
}

func TestPromiseConstructorSync(t *testing.T) {
	setupPromise()
	executor := newFuncObject("executor", 2, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		resolve := getCallable(args[0])
		resolve(runtime.Undefined, []*runtime.Value{runtime.NewNumber(100)})
		return runtime.Undefined, nil
	})

	result, err := promiseConstructorCall(runtime.Undefined, []*runtime.Value{runtime.NewObject(executor)})
	if err != nil {
		t.Fatal(err)
	}
	obj := toObject(result)
	pd := getPromiseData(obj)
	if pd.state != promiseFulfilled || pd.result.Number != 100 {
		t.Error("promise should be fulfilled with 100")
	}
}

func TestPromiseThenSync(t *testing.T) {
	setupPromise()
	executor := newFuncObject("executor", 2, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		resolve := getCallable(args[0])
		resolve(runtime.Undefined, []*runtime.Value{runtime.NewNumber(5)})
		return runtime.Undefined, nil
	})
	promise, _ := promiseConstructorCall(runtime.Undefined, []*runtime.Value{runtime.NewObject(executor)})

	onFulfilled := newFuncObject("onFulfilled", 1, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		return runtime.NewNumber(args[0].Number * 2), nil
	})

	thenResult, err := promiseThen(promise, []*runtime.Value{runtime.NewObject(onFulfilled)})
	if err != nil {
		t.Fatal(err)
	}
	thenObj := toObject(thenResult)
	thenPd := getPromiseData(thenObj)
	if thenPd.state != promiseFulfilled || thenPd.result.Number != 10 {
		t.Errorf("then: expected fulfilled with 10, got state=%d result=%v", thenPd.state, thenPd.result)
	}
}

func TestPromiseAll(t *testing.T) {
	setupPromise()
	p1, _ := promiseResolve(runtime.Undefined, []*runtime.Value{runtime.NewNumber(1)})
	p2, _ := promiseResolve(runtime.Undefined, []*runtime.Value{runtime.NewNumber(2)})
	p3, _ := promiseResolve(runtime.Undefined, []*runtime.Value{runtime.NewNumber(3)})

	arr := runtime.NewObject(newArray([]*runtime.Value{p1, p2, p3}))
	result, err := promiseAll(runtime.Undefined, []*runtime.Value{arr})
	if err != nil {
		t.Fatal(err)
	}
	obj := toObject(result)
	pd := getPromiseData(obj)
	if pd.state != promiseFulfilled {
		t.Error("Promise.all should be fulfilled")
	}
	resArr := toObject(pd.result)
	if resArr == nil || len(resArr.ArrayData) != 3 {
		t.Error("Promise.all result should be array of 3 items")
	}
}

func TestPromiseRace(t *testing.T) {
	setupPromise()
	p1, _ := promiseResolve(runtime.Undefined, []*runtime.Value{runtime.NewString("first")})
	p2, _ := promiseResolve(runtime.Undefined, []*runtime.Value{runtime.NewString("second")})

	arr := runtime.NewObject(newArray([]*runtime.Value{p1, p2}))
	result, err := promiseRace(runtime.Undefined, []*runtime.Value{arr})
	if err != nil {
		t.Fatal(err)
	}
	obj := toObject(result)
	pd := getPromiseData(obj)
	if pd.state != promiseFulfilled || pd.result.Str != "first" {
		t.Error("Promise.race should resolve with first value")
	}
}
