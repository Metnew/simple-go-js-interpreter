package builtins

import (
	"testing"

	"github.com/example/jsgo/runtime"
)

func TestSymbolConstructor(t *testing.T) {
	result, err := symbolConstructorCall(runtime.Undefined, []*runtime.Value{runtime.NewString("test")})
	if err != nil {
		t.Fatal(err)
	}
	if result.Type != runtime.TypeSymbol {
		t.Error("expected symbol type")
	}
	if result.Symbol.Description != "test" {
		t.Errorf("description: expected 'test', got %q", result.Symbol.Description)
	}
}

func TestSymbolFor(t *testing.T) {
	// Reset registry
	symbolRegistry = make(map[string]*runtime.Symbol)

	s1, _ := symbolFor(runtime.Undefined, []*runtime.Value{runtime.NewString("shared")})
	s2, _ := symbolFor(runtime.Undefined, []*runtime.Value{runtime.NewString("shared")})
	if s1.Symbol != s2.Symbol {
		t.Error("Symbol.for should return same symbol for same key")
	}

	s3, _ := symbolFor(runtime.Undefined, []*runtime.Value{runtime.NewString("other")})
	if s1.Symbol == s3.Symbol {
		t.Error("Symbol.for should return different symbols for different keys")
	}
}

func TestSymbolKeyFor(t *testing.T) {
	symbolRegistry = make(map[string]*runtime.Symbol)

	s1, _ := symbolFor(runtime.Undefined, []*runtime.Value{runtime.NewString("test")})
	key, _ := symbolKeyFor(runtime.Undefined, []*runtime.Value{s1})
	if key.Str != "test" {
		t.Errorf("Symbol.keyFor: expected 'test', got %q", key.Str)
	}

	// Non-registered symbol
	s2, _ := symbolConstructorCall(runtime.Undefined, []*runtime.Value{runtime.NewString("local")})
	key, _ = symbolKeyFor(runtime.Undefined, []*runtime.Value{s2})
	if key.Type != runtime.TypeUndefined {
		t.Error("Symbol.keyFor for non-registered symbol should return undefined")
	}
}

func TestWellKnownSymbols(t *testing.T) {
	createSymbolConstructor(nil)

	if SymIterator == nil {
		t.Error("Symbol.iterator should be defined")
	}
	if SymToPrimitive == nil {
		t.Error("Symbol.toPrimitive should be defined")
	}
	if SymHasInstance == nil {
		t.Error("Symbol.hasInstance should be defined")
	}
	if SymToStringTag == nil {
		t.Error("Symbol.toStringTag should be defined")
	}
}
