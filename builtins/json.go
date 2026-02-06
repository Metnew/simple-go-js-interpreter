package builtins

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/example/jsgo/runtime"
)

func createJSONObject(objProto *runtime.Object) *runtime.Object {
	j := runtime.NewOrdinaryObject(objProto)

	setMethod(j, "parse", 2, jsonParse)
	setMethod(j, "stringify", 3, jsonStringify)

	j.Set("@@toStringTag", runtime.NewString("JSON"))
	return j
}

func jsonParse(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	text := argAt(args, 0).ToString()
	var raw interface{}
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return nil, fmt.Errorf("SyntaxError: JSON.parse: %v", err)
	}
	result := goToJSValue(raw)
	if len(args) > 1 {
		reviver := getCallable(args[1])
		if reviver != nil {
			result = reviveValue(reviver, runtime.NewString(""), result)
		}
	}
	return result, nil
}

func goToJSValue(v interface{}) *runtime.Value {
	if v == nil {
		return runtime.Null
	}
	switch val := v.(type) {
	case bool:
		return runtime.NewBool(val)
	case float64:
		return runtime.NewNumber(val)
	case string:
		return runtime.NewString(val)
	case []interface{}:
		data := make([]*runtime.Value, len(val))
		for i, item := range val {
			data[i] = goToJSValue(item)
		}
		return runtime.NewObject(newArray(data))
	case map[string]interface{}:
		obj := runtime.NewOrdinaryObject(ObjectPrototype)
		for k, item := range val {
			obj.Set(k, goToJSValue(item))
		}
		return runtime.NewObject(obj)
	}
	return runtime.Undefined
}

func reviveValue(reviver runtime.CallableFunc, key *runtime.Value, val *runtime.Value) *runtime.Value {
	if val.Type == runtime.TypeObject && val.Object != nil {
		if val.Object.OType == runtime.ObjTypeArray {
			for i := range val.Object.ArrayData {
				k := runtime.NewNumber(float64(i))
				newVal := reviveValue(reviver, k, val.Object.ArrayData[i])
				if newVal.Type == runtime.TypeUndefined {
					val.Object.ArrayData[i] = runtime.Undefined
				} else {
					val.Object.ArrayData[i] = newVal
				}
			}
		} else {
			for k := range val.Object.Properties {
				propVal := val.Object.Get(k)
				newVal := reviveValue(reviver, runtime.NewString(k), propVal)
				if newVal.Type == runtime.TypeUndefined {
					delete(val.Object.Properties, k)
				} else {
					val.Object.Set(k, newVal)
				}
			}
		}
	}
	result, err := reviver(runtime.Undefined, []*runtime.Value{key, val})
	if err != nil {
		return val
	}
	return result
}

func jsonStringify(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	val := argAt(args, 0)
	var replacer runtime.CallableFunc
	var replacerArray []string
	if len(args) > 1 && args[1].Type == runtime.TypeObject && args[1].Object != nil {
		if args[1].Object.Callable != nil {
			replacer = args[1].Object.Callable
		} else if args[1].Object.OType == runtime.ObjTypeArray {
			for _, v := range args[1].Object.ArrayData {
				replacerArray = append(replacerArray, v.ToString())
			}
		}
	}
	indent := ""
	if len(args) > 2 {
		sp := args[2]
		if sp.Type == runtime.TypeNumber {
			n := int(sp.Number)
			if n > 10 {
				n = 10
			}
			if n > 0 {
				indent = strings.Repeat(" ", n)
			}
		} else if sp.Type == runtime.TypeString {
			indent = sp.Str
			if len(indent) > 10 {
				indent = indent[:10]
			}
		}
	}
	result := stringifyValue(val, replacer, replacerArray, indent, "")
	if result == "" {
		return runtime.Undefined, nil
	}
	return runtime.NewString(result), nil
}

func stringifyValue(val *runtime.Value, replacer runtime.CallableFunc, replacerArray []string, indent, currentIndent string) string {
	if val == nil || val.Type == runtime.TypeUndefined {
		return ""
	}
	if val.Type == runtime.TypeObject && val.Object != nil {
		toJSON := val.Object.Get("toJSON")
		if toJSON != runtime.Undefined {
			fn := getCallable(toJSON)
			if fn != nil {
				result, err := fn(val, nil)
				if err == nil {
					val = result
				}
			}
		}
	}
	switch val.Type {
	case runtime.TypeNull:
		return "null"
	case runtime.TypeBoolean:
		if val.Bool {
			return "true"
		}
		return "false"
	case runtime.TypeNumber:
		if isNaN(val.Number) || isInf(val.Number, 0) {
			return "null"
		}
		return val.ToString()
	case runtime.TypeString:
		b, _ := json.Marshal(val.Str)
		return string(b)
	case runtime.TypeObject:
		if val.Object == nil {
			return "null"
		}
		if val.Object.OType == runtime.ObjTypeArray {
			return stringifyArray(val.Object, replacer, replacerArray, indent, currentIndent)
		}
		return stringifyObject(val.Object, replacer, replacerArray, indent, currentIndent)
	}
	return ""
}

func stringifyArray(obj *runtime.Object, replacer runtime.CallableFunc, replacerArray []string, indent, currentIndent string) string {
	if len(obj.ArrayData) == 0 {
		return "[]"
	}
	newIndent := currentIndent + indent
	parts := make([]string, 0, len(obj.ArrayData))
	for i, v := range obj.ArrayData {
		if replacer != nil {
			result, err := replacer(runtime.NewObject(obj), []*runtime.Value{runtime.NewNumber(float64(i)), v})
			if err == nil {
				v = result
			}
		}
		s := stringifyValue(v, replacer, replacerArray, indent, newIndent)
		if s == "" {
			s = "null"
		}
		parts = append(parts, s)
	}
	if indent == "" {
		return "[" + strings.Join(parts, ",") + "]"
	}
	inner := strings.Join(parts, ",\n"+newIndent)
	return "[\n" + newIndent + inner + "\n" + currentIndent + "]"
}

func stringifyObject(obj *runtime.Object, replacer runtime.CallableFunc, replacerArray []string, indent, currentIndent string) string {
	keys := getEnumerableOwnKeys(obj)
	if replacerArray != nil {
		filtered := make([]string, 0)
		for _, k := range replacerArray {
			if obj.HasOwnProperty(k) {
				filtered = append(filtered, k)
			}
		}
		keys = filtered
	}
	newIndent := currentIndent + indent
	parts := make([]string, 0)
	for _, k := range keys {
		v := obj.Get(k)
		if replacer != nil {
			result, err := replacer(runtime.NewObject(obj), []*runtime.Value{runtime.NewString(k), v})
			if err == nil {
				v = result
			}
		}
		s := stringifyValue(v, replacer, replacerArray, indent, newIndent)
		if s == "" {
			continue
		}
		keyStr, _ := json.Marshal(k)
		if indent == "" {
			parts = append(parts, string(keyStr)+":"+s)
		} else {
			parts = append(parts, string(keyStr)+": "+s)
		}
	}
	if len(parts) == 0 {
		return "{}"
	}
	if indent == "" {
		return "{" + strings.Join(parts, ",") + "}"
	}
	inner := strings.Join(parts, ",\n"+newIndent)
	return "{\n" + newIndent + inner + "\n" + currentIndent + "}"
}
