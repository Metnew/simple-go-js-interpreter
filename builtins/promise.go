package builtins

import (
	"fmt"

	"github.com/example/jsgo/runtime"
)

var PromisePrototype *runtime.Object

const (
	promisePending   = 0
	promiseFulfilled = 1
	promiseRejected  = 2
)

func createPromiseConstructor(objProto *runtime.Object) (*runtime.Object, *runtime.Object) {
	proto := runtime.NewOrdinaryObject(objProto)
	proto.OType = runtime.ObjTypePromise
	PromisePrototype = proto

	setMethod(proto, "then", 2, promiseThen)
	setMethod(proto, "catch", 1, promiseCatch)
	setMethod(proto, "finally", 1, promiseFinally)

	ctor := newFuncObject("Promise", 1, promiseConstructorCall)
	ctor.Constructor = promiseConstructorCall

	setMethod(ctor, "resolve", 1, promiseResolve)
	setMethod(ctor, "reject", 1, promiseReject)
	setMethod(ctor, "all", 1, promiseAll)
	setMethod(ctor, "race", 1, promiseRace)
	setMethod(ctor, "allSettled", 1, promiseAllSettled)

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor, proto
}

type promiseData struct {
	state    int
	result   *runtime.Value
	onFulfill []*runtime.Value
	onReject  []*runtime.Value
}

func getPromiseData(obj *runtime.Object) *promiseData {
	if obj == nil || obj.Internal == nil {
		return nil
	}
	pd, _ := obj.Internal["promise"].(*promiseData)
	return pd
}

func newPromiseObject() (*runtime.Object, *promiseData) {
	pd := &promiseData{
		state:  promisePending,
		result: runtime.Undefined,
	}
	obj := &runtime.Object{
		OType:      runtime.ObjTypePromise,
		Properties: make(map[string]*runtime.Property),
		Prototype:  PromisePrototype,
		Internal:   map[string]interface{}{"promise": pd},
	}
	return obj, pd
}

func resolvePromise(pd *promiseData, val *runtime.Value) {
	if pd.state != promisePending {
		return
	}
	pd.state = promiseFulfilled
	pd.result = val
	for _, handler := range pd.onFulfill {
		fn := getCallable(handler)
		if fn != nil {
			fn(runtime.Undefined, []*runtime.Value{val})
		}
	}
	pd.onFulfill = nil
	pd.onReject = nil
}

func rejectPromise(pd *promiseData, val *runtime.Value) {
	if pd.state != promisePending {
		return
	}
	pd.state = promiseRejected
	pd.result = val
	for _, handler := range pd.onReject {
		fn := getCallable(handler)
		if fn != nil {
			fn(runtime.Undefined, []*runtime.Value{val})
		}
	}
	pd.onFulfill = nil
	pd.onReject = nil
}

func promiseConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	executor := getCallable(argAt(args, 0))
	if executor == nil {
		return nil, fmt.Errorf("TypeError: Promise resolver is not a function")
	}
	obj, pd := newPromiseObject()
	resolveFn := newFuncObject("resolve", 1, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		resolvePromise(pd, argAt(args, 0))
		return runtime.Undefined, nil
	})
	rejectFn := newFuncObject("reject", 1, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		rejectPromise(pd, argAt(args, 0))
		return runtime.Undefined, nil
	})
	_, err := executor(runtime.Undefined, []*runtime.Value{runtime.NewObject(resolveFn), runtime.NewObject(rejectFn)})
	if err != nil {
		rejectPromise(pd, runtime.NewString(err.Error()))
	}
	return runtime.NewObject(obj), nil
}

func promiseThen(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	pd := getPromiseData(obj)
	if pd == nil {
		return runtime.Undefined, nil
	}
	newObj, newPd := newPromiseObject()
	onFulfilled := argAt(args, 0)
	onRejected := argAt(args, 1)
	handleFulfill := func(val *runtime.Value) {
		fn := getCallable(onFulfilled)
		if fn != nil {
			result, err := fn(runtime.Undefined, []*runtime.Value{val})
			if err != nil {
				rejectPromise(newPd, runtime.NewString(err.Error()))
			} else {
				resolvePromise(newPd, result)
			}
		} else {
			resolvePromise(newPd, val)
		}
	}
	handleReject := func(val *runtime.Value) {
		fn := getCallable(onRejected)
		if fn != nil {
			result, err := fn(runtime.Undefined, []*runtime.Value{val})
			if err != nil {
				rejectPromise(newPd, runtime.NewString(err.Error()))
			} else {
				resolvePromise(newPd, result)
			}
		} else {
			rejectPromise(newPd, val)
		}
	}
	switch pd.state {
	case promiseFulfilled:
		handleFulfill(pd.result)
	case promiseRejected:
		handleReject(pd.result)
	case promisePending:
		fulfillWrapper := newFuncObject("", 1, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			handleFulfill(argAt(args, 0))
			return runtime.Undefined, nil
		})
		rejectWrapper := newFuncObject("", 1, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			handleReject(argAt(args, 0))
			return runtime.Undefined, nil
		})
		pd.onFulfill = append(pd.onFulfill, runtime.NewObject(fulfillWrapper))
		pd.onReject = append(pd.onReject, runtime.NewObject(rejectWrapper))
	}
	return runtime.NewObject(newObj), nil
}

func promiseCatch(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return promiseThen(this, []*runtime.Value{runtime.Undefined, argAt(args, 0)})
}

func promiseFinally(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	onFinally := argAt(args, 0)
	fn := getCallable(onFinally)
	if fn == nil {
		return promiseThen(this, []*runtime.Value{onFinally, onFinally})
	}
	thenFulfill := newFuncObject("", 1, func(_ *runtime.Value, callArgs []*runtime.Value) (*runtime.Value, error) {
		_, _ = fn(runtime.Undefined, nil)
		return argAt(callArgs, 0), nil
	})
	thenReject := newFuncObject("", 1, func(_ *runtime.Value, callArgs []*runtime.Value) (*runtime.Value, error) {
		_, _ = fn(runtime.Undefined, nil)
		return nil, fmt.Errorf("%s", argAt(callArgs, 0).ToString())
	})
	return promiseThen(this, []*runtime.Value{runtime.NewObject(thenFulfill), runtime.NewObject(thenReject)})
}

func promiseResolve(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	val := argAt(args, 0)
	if val.Type == runtime.TypeObject && val.Object != nil && val.Object.OType == runtime.ObjTypePromise {
		return val, nil
	}
	obj, pd := newPromiseObject()
	resolvePromise(pd, val)
	return runtime.NewObject(obj), nil
}

func promiseReject(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj, pd := newPromiseObject()
	rejectPromise(pd, argAt(args, 0))
	return runtime.NewObject(obj), nil
}

func promiseAll(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	iterable := argAt(args, 0)
	if iterable.Type != runtime.TypeObject || iterable.Object == nil || iterable.Object.OType != runtime.ObjTypeArray {
		return nil, fmt.Errorf("TypeError: Promise.all requires an iterable")
	}
	promises := iterable.Object.ArrayData
	if len(promises) == 0 {
		obj, pd := newPromiseObject()
		resolvePromise(pd, runtime.NewObject(newArray([]*runtime.Value{})))
		return runtime.NewObject(obj), nil
	}
	obj, pd := newPromiseObject()
	results := make([]*runtime.Value, len(promises))
	remaining := len(promises)
	for i, p := range promises {
		idx := i
		pObj := toObject(p)
		if pObj != nil && pObj.OType == runtime.ObjTypePromise {
			ppd := getPromiseData(pObj)
			if ppd != nil {
				switch ppd.state {
				case promiseFulfilled:
					results[idx] = ppd.result
					remaining--
				case promiseRejected:
					rejectPromise(pd, ppd.result)
					return runtime.NewObject(obj), nil
				case promisePending:
					fulfillHandler := newFuncObject("", 1, func(_ *runtime.Value, a []*runtime.Value) (*runtime.Value, error) {
						results[idx] = argAt(a, 0)
						remaining--
						if remaining == 0 {
							resolvePromise(pd, runtime.NewObject(newArray(results)))
						}
						return runtime.Undefined, nil
					})
					rejectHandler := newFuncObject("", 1, func(_ *runtime.Value, a []*runtime.Value) (*runtime.Value, error) {
						rejectPromise(pd, argAt(a, 0))
						return runtime.Undefined, nil
					})
					ppd.onFulfill = append(ppd.onFulfill, runtime.NewObject(fulfillHandler))
					ppd.onReject = append(ppd.onReject, runtime.NewObject(rejectHandler))
				}
				continue
			}
		}
		results[idx] = p
		remaining--
	}
	if remaining == 0 {
		resolvePromise(pd, runtime.NewObject(newArray(results)))
	}
	return runtime.NewObject(obj), nil
}

func promiseRace(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	iterable := argAt(args, 0)
	if iterable.Type != runtime.TypeObject || iterable.Object == nil || iterable.Object.OType != runtime.ObjTypeArray {
		return nil, fmt.Errorf("TypeError: Promise.race requires an iterable")
	}
	obj, pd := newPromiseObject()
	for _, p := range iterable.Object.ArrayData {
		pObj := toObject(p)
		if pObj != nil && pObj.OType == runtime.ObjTypePromise {
			ppd := getPromiseData(pObj)
			if ppd != nil {
				switch ppd.state {
				case promiseFulfilled:
					resolvePromise(pd, ppd.result)
					return runtime.NewObject(obj), nil
				case promiseRejected:
					rejectPromise(pd, ppd.result)
					return runtime.NewObject(obj), nil
				case promisePending:
					fulfillHandler := newFuncObject("", 1, func(_ *runtime.Value, a []*runtime.Value) (*runtime.Value, error) {
						resolvePromise(pd, argAt(a, 0))
						return runtime.Undefined, nil
					})
					rejectHandler := newFuncObject("", 1, func(_ *runtime.Value, a []*runtime.Value) (*runtime.Value, error) {
						rejectPromise(pd, argAt(a, 0))
						return runtime.Undefined, nil
					})
					ppd.onFulfill = append(ppd.onFulfill, runtime.NewObject(fulfillHandler))
					ppd.onReject = append(ppd.onReject, runtime.NewObject(rejectHandler))
				}
				continue
			}
		}
		resolvePromise(pd, p)
		return runtime.NewObject(obj), nil
	}
	return runtime.NewObject(obj), nil
}

func promiseAllSettled(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	iterable := argAt(args, 0)
	if iterable.Type != runtime.TypeObject || iterable.Object == nil || iterable.Object.OType != runtime.ObjTypeArray {
		return nil, fmt.Errorf("TypeError: Promise.allSettled requires an iterable")
	}
	promises := iterable.Object.ArrayData
	if len(promises) == 0 {
		obj, pd := newPromiseObject()
		resolvePromise(pd, runtime.NewObject(newArray([]*runtime.Value{})))
		return runtime.NewObject(obj), nil
	}
	obj, pd := newPromiseObject()
	results := make([]*runtime.Value, len(promises))
	remaining := len(promises)
	for i, p := range promises {
		idx := i
		pObj := toObject(p)
		if pObj != nil && pObj.OType == runtime.ObjTypePromise {
			ppd := getPromiseData(pObj)
			if ppd != nil {
				switch ppd.state {
				case promiseFulfilled:
					result := runtime.NewOrdinaryObject(nil)
					result.Set("status", runtime.NewString("fulfilled"))
					result.Set("value", ppd.result)
					results[idx] = runtime.NewObject(result)
					remaining--
				case promiseRejected:
					result := runtime.NewOrdinaryObject(nil)
					result.Set("status", runtime.NewString("rejected"))
					result.Set("reason", ppd.result)
					results[idx] = runtime.NewObject(result)
					remaining--
				case promisePending:
					fulfillHandler := newFuncObject("", 1, func(_ *runtime.Value, a []*runtime.Value) (*runtime.Value, error) {
						r := runtime.NewOrdinaryObject(nil)
						r.Set("status", runtime.NewString("fulfilled"))
						r.Set("value", argAt(a, 0))
						results[idx] = runtime.NewObject(r)
						remaining--
						if remaining == 0 {
							resolvePromise(pd, runtime.NewObject(newArray(results)))
						}
						return runtime.Undefined, nil
					})
					rejectHandler := newFuncObject("", 1, func(_ *runtime.Value, a []*runtime.Value) (*runtime.Value, error) {
						r := runtime.NewOrdinaryObject(nil)
						r.Set("status", runtime.NewString("rejected"))
						r.Set("reason", argAt(a, 0))
						results[idx] = runtime.NewObject(r)
						remaining--
						if remaining == 0 {
							resolvePromise(pd, runtime.NewObject(newArray(results)))
						}
						return runtime.Undefined, nil
					})
					ppd.onFulfill = append(ppd.onFulfill, runtime.NewObject(fulfillHandler))
					ppd.onReject = append(ppd.onReject, runtime.NewObject(rejectHandler))
				}
				continue
			}
		}
		result := runtime.NewOrdinaryObject(nil)
		result.Set("status", runtime.NewString("fulfilled"))
		result.Set("value", p)
		results[idx] = runtime.NewObject(result)
		remaining--
	}
	if remaining == 0 {
		resolvePromise(pd, runtime.NewObject(newArray(results)))
	}
	return runtime.NewObject(obj), nil
}
