package builtins

import (
	"testing"

	"github.com/example/jsgo/runtime"
)

func setupRegExp() {
	createObjectConstructor()
	createArrayConstructor(ObjectPrototype)
	createRegExpConstructor(ObjectPrototype)
}

func TestRegExpTest(t *testing.T) {
	setupRegExp()
	re, err := createRegExpObject("[0-9]+", "")
	if err != nil {
		t.Fatal(err)
	}

	result, _ := regexpTest(re, []*runtime.Value{runtime.NewString("hello123")})
	if !result.Bool {
		t.Error("test('hello123') should match")
	}

	result, _ = regexpTest(re, []*runtime.Value{runtime.NewString("hello")})
	if result.Bool {
		t.Error("test('hello') should not match")
	}
}

func TestRegExpExec(t *testing.T) {
	setupRegExp()
	re, err := createRegExpObject("(\\w+)@(\\w+)", "")
	if err != nil {
		t.Fatal(err)
	}

	result, _ := regexpExec(re, []*runtime.Value{runtime.NewString("user@host")})
	if result.Type == runtime.TypeNull {
		t.Fatal("exec should return match")
	}
	obj := toObject(result)
	if obj == nil || len(obj.ArrayData) != 3 {
		t.Fatalf("expected 3 groups, got %v", obj)
	}
	if obj.ArrayData[0].Str != "user@host" {
		t.Errorf("full match: expected 'user@host', got %q", obj.ArrayData[0].Str)
	}
	if obj.ArrayData[1].Str != "user" {
		t.Errorf("group 1: expected 'user', got %q", obj.ArrayData[1].Str)
	}
}

func TestRegExpExecNoMatch(t *testing.T) {
	setupRegExp()
	re, _ := createRegExpObject("xyz", "")
	result, _ := regexpExec(re, []*runtime.Value{runtime.NewString("abc")})
	if result.Type != runtime.TypeNull {
		t.Error("exec should return null for no match")
	}
}

func TestRegExpToString(t *testing.T) {
	setupRegExp()
	re, _ := createRegExpObject("abc", "gi")
	result, _ := regexpToString(re, nil)
	if result.Str != "/abc/gi" {
		t.Errorf("toString: expected '/abc/gi', got %q", result.Str)
	}
}

func TestRegExpCaseInsensitive(t *testing.T) {
	setupRegExp()
	re, _ := createRegExpObject("hello", "i")
	result, _ := regexpTest(re, []*runtime.Value{runtime.NewString("HELLO")})
	if !result.Bool {
		t.Error("case insensitive test should match")
	}
}
