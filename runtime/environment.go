package runtime

import "fmt"

// Environment represents a lexical scope.
type Environment struct {
	store      map[string]*Binding
	outer      *Environment
	isBlock    bool // true for block scopes (let/const), false for function scopes
	annexBNames map[string]bool // names hoisted by Annex B (block-level function decls)
}

type Binding struct {
	Value    *Value
	Mutable  bool // false for const
	Kind     string // "var", "let", "const", "function"
	Declared bool   // false until initialized (for TDZ)
}

func NewEnvironment(outer *Environment, isBlock bool) *Environment {
	return &Environment{
		store:   make(map[string]*Binding),
		outer:   outer,
		isBlock: isBlock,
	}
}

// Declare declares a variable in the current scope.
func (e *Environment) Declare(name string, kind string, value *Value) error {
	if kind == "let" || kind == "const" {
		if _, exists := e.store[name]; exists {
			return fmt.Errorf("SyntaxError: Identifier '%s' has already been declared", name)
		}
	}
	e.store[name] = &Binding{
		Value:    value,
		Mutable:  kind != "const",
		Kind:     kind,
		Declared: true,
	}
	return nil
}

// Get retrieves a variable value, walking up the scope chain.
func (e *Environment) Get(name string) (*Value, error) {
	if binding, ok := e.store[name]; ok {
		if !binding.Declared {
			return nil, fmt.Errorf("ReferenceError: Cannot access '%s' before initialization", name)
		}
		return binding.Value, nil
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return nil, fmt.Errorf("ReferenceError: %s is not defined", name)
}

// Set updates a variable value in the scope where it was declared.
func (e *Environment) Set(name string, value *Value) error {
	if binding, ok := e.store[name]; ok {
		if !binding.Mutable {
			return fmt.Errorf("TypeError: Assignment to constant variable '%s'", name)
		}
		binding.Value = value
		return nil
	}
	if e.outer != nil {
		return e.outer.Set(name, value)
	}
	return fmt.Errorf("ReferenceError: %s is not defined", name)
}

// SetInCurrentScope sets/creates a variable in the current scope (for var hoisting).
func (e *Environment) SetInCurrentScope(name string, value *Value) {
	if binding, ok := e.store[name]; ok {
		binding.Value = value
		return
	}
	e.store[name] = &Binding{
		Value:    value,
		Mutable:  true,
		Kind:     "var",
		Declared: true,
	}
}

// DeclareVar declares a var binding only if the name doesn't already exist in this scope.
// Used for Annex B block-level function hoisting: the name is hoisted as undefined
// but must not overwrite existing bindings (var, let, const, or function).
// Tracks the name as an Annex B hoisted name for runtime propagation checks.
func (e *Environment) DeclareVar(name string) {
	if e.annexBNames == nil {
		e.annexBNames = make(map[string]bool)
	}
	e.annexBNames[name] = true
	if _, exists := e.store[name]; exists {
		return
	}
	e.store[name] = &Binding{
		Value:    Undefined,
		Mutable:  true,
		Kind:     "var",
		Declared: true,
	}
}

// IsAnnexBHoisted returns true if the given name was Annex B hoisted in this scope.
func (e *Environment) IsAnnexBHoisted(name string) bool {
	return e.annexBNames != nil && e.annexBNames[name]
}

// GetFunctionScope walks up to find the nearest function scope (or global).
func (e *Environment) GetFunctionScope() *Environment {
	if !e.isBlock {
		return e
	}
	if e.outer != nil {
		return e.outer.GetFunctionScope()
	}
	return e
}

// HasVarBinding returns true if the given name has a var or function binding in this scope.
// Used by Annex B to check whether propagating a block function to function scope is safe.
func (e *Environment) HasVarBinding(name string) bool {
	if binding, ok := e.store[name]; ok {
		return binding.Kind == "var" || binding.Kind == "function"
	}
	return false
}

// IsBlock returns true if this is a block scope (not a function/program scope).
func (e *Environment) IsBlock() bool {
	return e.isBlock
}

// HasLexicalInEnclosingBlocks checks whether any block scope between this scope's
// parent and the target scope (exclusive) has a lexical binding (let/const/function)
// for the given name. Used to determine if Annex B propagation should be suppressed.
func (e *Environment) HasLexicalInEnclosingBlocks(name string, target *Environment) bool {
	for cur := e.outer; cur != nil && cur != target; cur = cur.outer {
		if !cur.isBlock {
			break
		}
		if binding, ok := cur.store[name]; ok {
			if binding.Kind == "let" || binding.Kind == "const" || binding.Kind == "function" {
				return true
			}
		}
	}
	return false
}

// ForEachBinding calls fn for each binding in the current scope.
func (e *Environment) ForEachBinding(fn func(name string, kind string)) {
	for name, binding := range e.store {
		fn(name, binding.Kind)
	}
}

// HasBinding returns true if this scope has a binding for the given name.
func (e *Environment) HasBinding(name string) bool {
	_, ok := e.store[name]
	return ok
}

// Outer returns the parent environment.
func (e *Environment) Outer() *Environment {
	return e.outer
}
