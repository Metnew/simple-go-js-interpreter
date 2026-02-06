package builtins

import (
	"testing"

	"github.com/example/jsgo/runtime"
)

func setupJSON() {
	createObjectConstructor()
	createArrayConstructor(ObjectPrototype)
}

func TestJSONParseSimple(t *testing.T) {
	setupJSON()
	tests := []struct {
		input string
		check func(*runtime.Value) bool
	}{
		{`42`, func(v *runtime.Value) bool { return v.Number == 42 }},
		{`"hello"`, func(v *runtime.Value) bool { return v.Str == "hello" }},
		{`true`, func(v *runtime.Value) bool { return v.Bool == true }},
		{`null`, func(v *runtime.Value) bool { return v.Type == runtime.TypeNull }},
		{`[1,2,3]`, func(v *runtime.Value) bool {
			arr := toObject(v)
			return arr != nil && len(arr.ArrayData) == 3
		}},
		{`{"a":1}`, func(v *runtime.Value) bool {
			obj := toObject(v)
			return obj != nil && obj.Get("a").Number == 1
		}},
	}
	for _, tt := range tests {
		result, err := jsonParse(runtime.Undefined, []*runtime.Value{runtime.NewString(tt.input)})
		if err != nil {
			t.Errorf("JSON.parse(%q): %v", tt.input, err)
			continue
		}
		if !tt.check(result) {
			t.Errorf("JSON.parse(%q): check failed, got %v", tt.input, result)
		}
	}
}

func TestJSONStringifySimple(t *testing.T) {
	setupJSON()
	tests := []struct {
		val  *runtime.Value
		want string
	}{
		{runtime.NewNumber(42), "42"},
		{runtime.NewString("hello"), `"hello"`},
		{runtime.True, "true"},
		{runtime.Null, "null"},
	}
	for _, tt := range tests {
		result, err := jsonStringify(runtime.Undefined, []*runtime.Value{tt.val})
		if err != nil {
			t.Errorf("JSON.stringify(%v): %v", tt.val, err)
			continue
		}
		if result.Str != tt.want {
			t.Errorf("JSON.stringify(%v): got %q, want %q", tt.val, result.Str, tt.want)
		}
	}
}

func TestJSONStringifyObject(t *testing.T) {
	setupJSON()
	obj := runtime.NewOrdinaryObject(nil)
	obj.Set("a", runtime.NewNumber(1))
	obj.Set("b", runtime.NewString("hello"))

	result, err := jsonStringify(runtime.Undefined, []*runtime.Value{runtime.NewObject(obj)})
	if err != nil {
		t.Fatal(err)
	}
	// Keys are sorted, so we get a deterministic output
	expected1 := `{"a":1,"b":"hello"}`
	expected2 := `{"b":"hello","a":1}`
	if result.Str != expected1 && result.Str != expected2 {
		t.Errorf("JSON.stringify object: got %q", result.Str)
	}
}

func TestJSONStringifyArray(t *testing.T) {
	setupJSON()
	arr := newArray([]*runtime.Value{runtime.NewNumber(1), runtime.NewString("two"), runtime.True})

	result, err := jsonStringify(runtime.Undefined, []*runtime.Value{runtime.NewObject(arr)})
	if err != nil {
		t.Fatal(err)
	}
	if result.Str != `[1,"two",true]` {
		t.Errorf("JSON.stringify array: got %q, want %q", result.Str, `[1,"two",true]`)
	}
}

func TestJSONStringifyWithIndent(t *testing.T) {
	setupJSON()
	obj := runtime.NewOrdinaryObject(nil)
	obj.Set("x", runtime.NewNumber(1))

	result, err := jsonStringify(runtime.Undefined, []*runtime.Value{runtime.NewObject(obj), runtime.Undefined, runtime.NewNumber(2)})
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\n  \"x\": 1\n}"
	if result.Str != expected {
		t.Errorf("JSON.stringify with indent: got %q, want %q", result.Str, expected)
	}
}

func TestJSONParseSyntaxError(t *testing.T) {
	_, err := jsonParse(runtime.Undefined, []*runtime.Value{runtime.NewString("{invalid}")})
	if err == nil {
		t.Error("expected SyntaxError for invalid JSON")
	}
}

func TestJSONStringifyNaN(t *testing.T) {
	result, err := jsonStringify(runtime.Undefined, []*runtime.Value{runtime.NaN})
	if err != nil {
		t.Fatal(err)
	}
	if result.Str != "null" {
		t.Errorf("JSON.stringify(NaN): expected 'null', got %q", result.Str)
	}
}
