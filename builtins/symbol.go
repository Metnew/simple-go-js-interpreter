package builtins

import (
	"fmt"
	"sync/atomic"

	"github.com/example/jsgo/runtime"
)

var (
	symbolCounter uint64
	symbolRegistry = make(map[string]*runtime.Symbol)

	SymIterator    *runtime.Symbol
	SymToPrimitive *runtime.Symbol
	SymHasInstance *runtime.Symbol
	SymToStringTag *runtime.Symbol
)

func nextSymbolID() uint64 {
	return atomic.AddUint64(&symbolCounter, 1)
}

func createSymbolConstructor(objProto *runtime.Object) *runtime.Object {
	ctor := newFuncObject("Symbol", 0, symbolConstructorCall)

	setMethod(ctor, "for", 1, symbolFor)
	setMethod(ctor, "keyFor", 1, symbolKeyFor)

	// Well-known symbols
	SymIterator = &runtime.Symbol{Description: "Symbol.iterator"}
	SymToPrimitive = &runtime.Symbol{Description: "Symbol.toPrimitive"}
	SymHasInstance = &runtime.Symbol{Description: "Symbol.hasInstance"}
	SymToStringTag = &runtime.Symbol{Description: "Symbol.toStringTag"}

	setConstant(ctor, "iterator", &runtime.Value{Type: runtime.TypeSymbol, Symbol: SymIterator})
	setConstant(ctor, "toPrimitive", &runtime.Value{Type: runtime.TypeSymbol, Symbol: SymToPrimitive})
	setConstant(ctor, "hasInstance", &runtime.Value{Type: runtime.TypeSymbol, Symbol: SymHasInstance})
	setConstant(ctor, "toStringTag", &runtime.Value{Type: runtime.TypeSymbol, Symbol: SymToStringTag})

	return ctor
}

func symbolConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	desc := ""
	if len(args) > 0 && args[0].Type != runtime.TypeUndefined {
		desc = args[0].ToString()
	}
	sym := &runtime.Symbol{Description: desc}
	return &runtime.Value{Type: runtime.TypeSymbol, Symbol: sym}, nil
}

func symbolFor(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	key := argAt(args, 0).ToString()
	if sym, ok := symbolRegistry[key]; ok {
		return &runtime.Value{Type: runtime.TypeSymbol, Symbol: sym}, nil
	}
	sym := &runtime.Symbol{Description: key}
	symbolRegistry[key] = sym
	return &runtime.Value{Type: runtime.TypeSymbol, Symbol: sym}, nil
}

func symbolKeyFor(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	a := argAt(args, 0)
	if a.Type != runtime.TypeSymbol || a.Symbol == nil {
		return nil, fmt.Errorf("TypeError: Symbol.keyFor requires a symbol")
	}
	for k, v := range symbolRegistry {
		if v == a.Symbol {
			return runtime.NewString(k), nil
		}
	}
	return runtime.Undefined, nil
}
