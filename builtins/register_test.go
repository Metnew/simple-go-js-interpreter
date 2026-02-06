package builtins

import (
	"testing"

	"github.com/example/jsgo/runtime"
)

func TestRegisterAll(t *testing.T) {
	env := runtime.NewEnvironment(nil, false)
	globalObj := runtime.NewOrdinaryObject(nil)

	RegisterAll(env, globalObj)

	// Check that core built-ins are registered
	names := []string{
		"Object", "Function", "Array", "String", "Number", "Boolean",
		"Symbol", "Error", "TypeError", "ReferenceError", "SyntaxError",
		"RangeError", "URIError", "EvalError",
		"RegExp", "Map", "Set", "WeakMap", "WeakSet",
		"Promise", "Proxy", "Reflect",
		"Math", "JSON", "console",
		"parseInt", "parseFloat", "isNaN", "isFinite",
		"encodeURI", "decodeURI", "encodeURIComponent", "decodeURIComponent",
		"eval", "undefined", "NaN", "Infinity",
	}
	for _, name := range names {
		val, err := env.Get(name)
		if err != nil {
			t.Errorf("missing built-in: %s (error: %v)", name, err)
			continue
		}
		if val == nil {
			t.Errorf("built-in %s is nil", name)
		}
	}
}
