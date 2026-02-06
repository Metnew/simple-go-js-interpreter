package builtins

import (
	"testing"

	"github.com/example/jsgo/runtime"
)

func TestStringCharAt(t *testing.T) {
	this := runtime.NewString("hello")
	result, _ := stringCharAt(this, []*runtime.Value{runtime.NewNumber(1)})
	if result.Str != "e" {
		t.Errorf("charAt(1): expected 'e', got %q", result.Str)
	}

	result, _ = stringCharAt(this, []*runtime.Value{runtime.NewNumber(10)})
	if result.Str != "" {
		t.Errorf("charAt(10): expected '', got %q", result.Str)
	}
}

func TestStringCharCodeAt(t *testing.T) {
	this := runtime.NewString("A")
	result, _ := stringCharCodeAt(this, []*runtime.Value{runtime.NewNumber(0)})
	if result.Number != 65 {
		t.Errorf("charCodeAt(0): expected 65, got %v", result.Number)
	}
}

func TestStringIndexOf(t *testing.T) {
	this := runtime.NewString("hello world")
	result, _ := stringIndexOf(this, []*runtime.Value{runtime.NewString("world")})
	if result.Number != 6 {
		t.Errorf("indexOf('world'): expected 6, got %v", result.Number)
	}

	result, _ = stringIndexOf(this, []*runtime.Value{runtime.NewString("xyz")})
	if result.Number != -1 {
		t.Errorf("indexOf('xyz'): expected -1, got %v", result.Number)
	}
}

func TestStringIncludes(t *testing.T) {
	this := runtime.NewString("hello world")
	result, _ := stringIncludes(this, []*runtime.Value{runtime.NewString("world")})
	if !result.Bool {
		t.Error("includes('world') should be true")
	}
}

func TestStringStartsEndsWith(t *testing.T) {
	this := runtime.NewString("hello world")
	result, _ := stringStartsWith(this, []*runtime.Value{runtime.NewString("hello")})
	if !result.Bool {
		t.Error("startsWith('hello') should be true")
	}
	result, _ = stringEndsWith(this, []*runtime.Value{runtime.NewString("world")})
	if !result.Bool {
		t.Error("endsWith('world') should be true")
	}
}

func TestStringSlice(t *testing.T) {
	this := runtime.NewString("hello world")
	result, _ := stringSlice(this, []*runtime.Value{runtime.NewNumber(0), runtime.NewNumber(5)})
	if result.Str != "hello" {
		t.Errorf("slice(0,5): expected 'hello', got %q", result.Str)
	}

	result, _ = stringSlice(this, []*runtime.Value{runtime.NewNumber(-5)})
	if result.Str != "world" {
		t.Errorf("slice(-5): expected 'world', got %q", result.Str)
	}
}

func TestStringUpperLower(t *testing.T) {
	this := runtime.NewString("Hello")
	result, _ := stringToUpperCase(this, nil)
	if result.Str != "HELLO" {
		t.Errorf("toUpperCase: expected 'HELLO', got %q", result.Str)
	}
	result, _ = stringToLowerCase(this, nil)
	if result.Str != "hello" {
		t.Errorf("toLowerCase: expected 'hello', got %q", result.Str)
	}
}

func TestStringTrim(t *testing.T) {
	this := runtime.NewString("  hello  ")
	result, _ := stringTrim(this, nil)
	if result.Str != "hello" {
		t.Errorf("trim: expected 'hello', got %q", result.Str)
	}
	result, _ = stringTrimStart(this, nil)
	if result.Str != "hello  " {
		t.Errorf("trimStart: expected 'hello  ', got %q", result.Str)
	}
	result, _ = stringTrimEnd(this, nil)
	if result.Str != "  hello" {
		t.Errorf("trimEnd: expected '  hello', got %q", result.Str)
	}
}

func TestStringRepeat(t *testing.T) {
	this := runtime.NewString("ab")
	result, _ := stringRepeat(this, []*runtime.Value{runtime.NewNumber(3)})
	if result.Str != "ababab" {
		t.Errorf("repeat(3): expected 'ababab', got %q", result.Str)
	}
}

func TestStringPadStartEnd(t *testing.T) {
	this := runtime.NewString("5")
	result, _ := stringPadStart(this, []*runtime.Value{runtime.NewNumber(3), runtime.NewString("0")})
	if result.Str != "005" {
		t.Errorf("padStart(3,'0'): expected '005', got %q", result.Str)
	}

	result, _ = stringPadEnd(this, []*runtime.Value{runtime.NewNumber(3), runtime.NewString("0")})
	if result.Str != "500" {
		t.Errorf("padEnd(3,'0'): expected '500', got %q", result.Str)
	}
}

func TestStringSplit(t *testing.T) {
	this := runtime.NewString("a,b,c")
	result, _ := stringSplit(this, []*runtime.Value{runtime.NewString(",")})
	data := getArrayData(result)
	if len(data) != 3 || data[0].Str != "a" || data[1].Str != "b" || data[2].Str != "c" {
		t.Error("split(','): expected ['a','b','c']")
	}
}

func TestStringReplace(t *testing.T) {
	this := runtime.NewString("hello world")
	result, _ := stringReplace(this, []*runtime.Value{runtime.NewString("world"), runtime.NewString("go")})
	if result.Str != "hello go" {
		t.Errorf("replace: expected 'hello go', got %q", result.Str)
	}
}

func TestStringConcat(t *testing.T) {
	this := runtime.NewString("hello")
	result, _ := stringConcat(this, []*runtime.Value{runtime.NewString(" "), runtime.NewString("world")})
	if result.Str != "hello world" {
		t.Errorf("concat: expected 'hello world', got %q", result.Str)
	}
}

func TestStringFromCharCode(t *testing.T) {
	result, _ := stringFromCharCode(runtime.Undefined, []*runtime.Value{runtime.NewNumber(72), runtime.NewNumber(105)})
	if result.Str != "Hi" {
		t.Errorf("fromCharCode(72,105): expected 'Hi', got %q", result.Str)
	}
}

func TestStringSubstring(t *testing.T) {
	this := runtime.NewString("hello")
	result, _ := stringSubstring(this, []*runtime.Value{runtime.NewNumber(1), runtime.NewNumber(4)})
	if result.Str != "ell" {
		t.Errorf("substring(1,4): expected 'ell', got %q", result.Str)
	}
}

func TestStringAt(t *testing.T) {
	this := runtime.NewString("hello")
	result, _ := stringAt(this, []*runtime.Value{runtime.NewNumber(-1)})
	if result.Str != "o" {
		t.Errorf("at(-1): expected 'o', got %q", result.Str)
	}
}
