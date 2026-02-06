package interpreter

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/example/jsgo/ast"
	"github.com/example/jsgo/parser"
	"github.com/example/jsgo/runtime"
)

// Signal types for control flow
type signalType int

const (
	sigNone signalType = iota
	sigReturn
	sigBreak
	sigContinue
	sigThrow
)

type signal struct {
	typ   signalType
	value *runtime.Value
	label string // for labeled break/continue
}

// jsError wraps a JS value as a Go error for try/catch.
type jsError struct {
	value *runtime.Value
}

func (e *jsError) Error() string {
	if e.value != nil {
		return e.value.ToString()
	}
	return "undefined"
}

// makeErrorObject creates a proper JS Error object (TypeError, ReferenceError, etc.)
// that works with instanceof. It looks up the constructor from the environment to get
// the right prototype chain. Falls back to a simple object if the constructor isn't available.
func makeErrorObject(errorType string, message string, env *runtime.Environment) *runtime.Value {
	ctorVal, err := env.Get(errorType)
	if err == nil && ctorVal.Type == runtime.TypeObject && ctorVal.Object != nil {
		protoProp := ctorVal.Object.Get("prototype")
		if protoProp.Type == runtime.TypeObject && protoProp.Object != nil {
			obj := &runtime.Object{
				OType:      runtime.ObjTypeError,
				Properties: make(map[string]*runtime.Property),
				Prototype:  protoProp.Object,
			}
			obj.Set("name", runtime.NewString(errorType))
			obj.Set("message", runtime.NewString(message))
			obj.Set("stack", runtime.NewString(fmt.Sprintf("%s: %s", errorType, message)))
			return runtime.NewObject(obj)
		}
	}
	// Fallback: create a simple error object without prototype chain
	obj := runtime.NewErrorObject(nil, message)
	obj.Set("name", runtime.NewString(errorType))
	return runtime.NewObject(obj)
}

// errorFromGoError converts a Go error (from environment.Get/Set) into a proper JS Error object.
// It parses the error type prefix (e.g. "ReferenceError: ...") and creates the right error type.
func errorFromGoError(goErr error, env *runtime.Environment) *runtime.Value {
	msg := goErr.Error()
	errorTypes := []string{"TypeError", "ReferenceError", "SyntaxError", "RangeError", "URIError", "EvalError"}
	for _, et := range errorTypes {
		prefix := et + ": "
		if strings.HasPrefix(msg, prefix) {
			return makeErrorObject(et, strings.TrimPrefix(msg, prefix), env)
		}
	}
	return makeErrorObject("Error", msg, env)
}

// Interpreter evaluates an AST using tree-walking.
type Interpreter struct {
	global  *runtime.Environment
	natives map[string]runtime.CallableFunc
}

func New() *Interpreter {
	interp := &Interpreter{
		global:  runtime.NewEnvironment(nil, false),
		natives: make(map[string]runtime.CallableFunc),
	}
	return interp
}

// RegisterNative registers a native Go function as a global JS function.
func (interp *Interpreter) RegisterNative(name string, fn runtime.CallableFunc) {
	interp.natives[name] = fn
}

// GlobalEnv returns the interpreter's global environment for builtin registration.
func (interp *Interpreter) GlobalEnv() *runtime.Environment {
	return interp.global
}

// Eval parses and evaluates a JS source string.
func (interp *Interpreter) Eval(source string) (*runtime.Value, error) {
	p := parser.New(source)
	program, errs := p.ParseProgram()
	if len(errs) > 0 {
		return nil, fmt.Errorf("parse errors: %v", errs)
	}

	env := runtime.NewEnvironment(interp.global, false)

	// register natives
	for name, fn := range interp.natives {
		fnObj := runtime.NewFunctionObject(nil, fn)
		env.Declare(name, "var", runtime.NewObject(fnObj))
	}

	// register eval — tagged via Internal so evalCall can detect direct eval
	evalFnObj := runtime.NewFunctionObject(nil, func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		// indirect eval: evaluate in global scope
		if len(args) == 0 {
			return runtime.Undefined, nil
		}
		if args[0].Type != runtime.TypeString {
			return args[0], nil
		}
		result, sig := interp.evalCodeInEnv(args[0].Str, env)
		if sig.typ == sigThrow {
			return nil, &jsError{value: sig.value}
		}
		return result, nil
	})
	evalFnObj.Internal = map[string]interface{}{"isBuiltinEval": true}
	evalFnObj.Set("length", runtime.NewNumber(1))
	env.Declare("eval", "var", runtime.NewObject(evalFnObj))

	// register Function constructor
	funcCtor := interp.makeFunctionConstructor(env)
	env.Declare("Function", "var", funcCtor)

	// hoist var declarations and function declarations
	interp.hoist(program.Statements, env)

	var result *runtime.Value
	for _, stmt := range program.Statements {
		val, sig := interp.execStatement(stmt, env)
		if sig.typ == sigThrow {
			return nil, &jsError{value: sig.value}
		}
		if sig.typ == sigReturn {
			return sig.value, nil
		}
		if val != nil {
			result = val
		}
	}

	if result == nil {
		return runtime.Undefined, nil
	}
	return result, nil
}

// hoist performs var and function hoisting (delegates to hoistComprehensive in hoisting.go).
func (interp *Interpreter) hoist(stmts []ast.Statement, env *runtime.Environment) {
	interp.hoistComprehensive(stmts, env)
}

func (interp *Interpreter) extractBindingNames(node ast.Expression) []string {
	switch n := node.(type) {
	case *ast.Identifier:
		return []string{n.Value}
	case *ast.ObjectPattern:
		var names []string
		for _, prop := range n.Properties {
			if rest, ok := prop.Value.(*ast.RestElement); ok {
				names = append(names, interp.extractBindingNames(rest.Argument)...)
			} else if prop.Value != nil {
				names = append(names, interp.extractBindingNames(prop.Value)...)
			} else {
				names = append(names, interp.extractBindingNames(prop.Key)...)
			}
		}
		return names
	case *ast.ArrayPattern:
		var names []string
		for _, elem := range n.Elements {
			if elem != nil {
				names = append(names, interp.extractBindingNames(elem)...)
			}
		}
		return names
	case *ast.AssignmentPattern:
		return interp.extractBindingNames(n.Left)
	case *ast.RestElement:
		return interp.extractBindingNames(n.Argument)
	}
	return nil
}

// execStatement executes a statement, returning a value and a control flow signal.
func (interp *Interpreter) execStatement(stmt ast.Statement, env *runtime.Environment) (*runtime.Value, signal) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		val, sig := interp.evalExpression(s.Expression, env)
		return val, sig
	case *ast.VariableDeclaration:
		return interp.execVarDecl(s, env)
	case *ast.BlockStatement:
		return interp.execBlock(s, env)
	case *ast.ReturnStatement:
		return interp.execReturn(s, env)
	case *ast.IfStatement:
		return interp.execIf(s, env)
	case *ast.WhileStatement:
		return interp.execWhile(s, env)
	case *ast.DoWhileStatement:
		return interp.execDoWhile(s, env)
	case *ast.ForStatement:
		return interp.execFor(s, env)
	case *ast.ForInStatement:
		return interp.execForIn(s, env)
	case *ast.ForOfStatement:
		return interp.execForOf(s, env)
	case *ast.BreakStatement:
		label := ""
		if s.Label != nil {
			label = s.Label.Value
		}
		return nil, signal{typ: sigBreak, label: label}
	case *ast.ContinueStatement:
		label := ""
		if s.Label != nil {
			label = s.Label.Value
		}
		return nil, signal{typ: sigContinue, label: label}
	case *ast.SwitchStatement:
		return interp.execSwitch(s, env)
	case *ast.ThrowStatement:
		return interp.execThrow(s, env)
	case *ast.TryStatement:
		return interp.execTry(s, env)
	case *ast.FunctionDeclaration:
		// already hoisted
		return nil, signal{}
	case *ast.ClassDeclaration:
		return interp.execClassDecl(s, env)
	case *ast.LabeledStatement:
		return interp.execLabeled(s, env)
	case *ast.EmptyStatement:
		return nil, signal{}
	case *ast.DebuggerStatement:
		return nil, signal{}
	default:
		return nil, signal{typ: sigThrow, value: runtime.NewString(fmt.Sprintf("unsupported statement: %T", stmt))}
	}
}

func (interp *Interpreter) execVarDecl(s *ast.VariableDeclaration, env *runtime.Environment) (*runtime.Value, signal) {
	for _, decl := range s.Declarations {
		if decl.Value == nil && s.Kind == "var" {
			// var with no initializer: already hoisted, don't overwrite
			continue
		}
		var val *runtime.Value
		if decl.Value != nil {
			var sig signal
			val, sig = interp.evalExpression(decl.Value, env)
			if sig.typ != sigNone {
				return nil, sig
			}
		} else {
			val = runtime.Undefined
		}

		sig := interp.bindPattern(decl.Name, val, s.Kind, env)
		if sig.typ != sigNone {
			return nil, sig
		}
	}
	return nil, signal{}
}

func (interp *Interpreter) bindPattern(pattern ast.Expression, val *runtime.Value, kind string, env *runtime.Environment) signal {
	switch p := pattern.(type) {
	case *ast.Identifier:
		if kind == "var" {
			funcScope := env.GetFunctionScope()
			funcScope.SetInCurrentScope(p.Value, val)
		} else {
			if err := env.Declare(p.Value, kind, val); err != nil {
				return signal{typ: sigThrow, value: errorFromGoError(err, env)}
			}
		}
	case *ast.ObjectPattern:
		if val == nil || val.Type == runtime.TypeUndefined || val.Type == runtime.TypeNull {
			return signal{typ: sigThrow, value: makeErrorObject("TypeError", "Cannot destructure "+val.ToString(), env)}
		}
		if val.Type != runtime.TypeObject || val.Object == nil {
			return signal{}
		}
		used := make(map[string]bool)
		for _, prop := range p.Properties {
			if rest, ok := prop.Value.(*ast.RestElement); ok {
				restObj := runtime.NewOrdinaryObject(nil)
				for k, v := range val.Object.Properties {
					if !used[k] {
						restObj.Set(k, v.Value)
					}
				}
				return interp.bindPattern(rest.Argument, runtime.NewObject(restObj), kind, env)
			}
			var key string
			if ident, ok := prop.Key.(*ast.Identifier); ok {
				key = ident.Value
			} else if sl, ok := prop.Key.(*ast.StringLiteral); ok {
				key = sl.Value
			}
			used[key] = true
			propVal := val.Object.Get(key)
			target := prop.Value
			if target == nil {
				target = prop.Key
			}
			if ap, ok := target.(*ast.AssignmentPattern); ok {
				if propVal == nil || propVal.Type == runtime.TypeUndefined {
					var sig signal
					propVal, sig = interp.evalExpression(ap.Right, env)
					if sig.typ != sigNone {
						return sig
					}
				}
				target = ap.Left
			}
			sig := interp.bindPattern(target, propVal, kind, env)
			if sig.typ != sigNone {
				return sig
			}
		}
	case *ast.ArrayPattern:
		var elements []*runtime.Value
		if val != nil && val.Type == runtime.TypeObject && val.Object != nil && val.Object.OType == runtime.ObjTypeArray {
			elements = val.Object.ArrayData
		}
		for i, elem := range p.Elements {
			if elem == nil {
				continue
			}
			if rest, ok := elem.(*ast.RestElement); ok {
				var restElems []*runtime.Value
				if i < len(elements) {
					restElems = elements[i:]
				}
				restArr := runtime.NewArrayObject(nil, restElems)
				return interp.bindPattern(rest.Argument, runtime.NewObject(restArr), kind, env)
			}
			var elemVal *runtime.Value
			if i < len(elements) {
				elemVal = elements[i]
			} else {
				elemVal = runtime.Undefined
			}
			if ap, ok := elem.(*ast.AssignmentPattern); ok {
				if elemVal == nil || elemVal.Type == runtime.TypeUndefined {
					var sig signal
					elemVal, sig = interp.evalExpression(ap.Right, env)
					if sig.typ != sigNone {
						return sig
					}
				}
				sig := interp.bindPattern(ap.Left, elemVal, kind, env)
				if sig.typ != sigNone {
					return sig
				}
				continue
			}
			sig := interp.bindPattern(elem, elemVal, kind, env)
			if sig.typ != sigNone {
				return sig
			}
		}
	case *ast.AssignmentPattern:
		if val == nil || val.Type == runtime.TypeUndefined {
			var sig signal
			val, sig = interp.evalExpression(p.Right, env)
			if sig.typ != sigNone {
				return sig
			}
		}
		return interp.bindPattern(p.Left, val, kind, env)
	}
	return signal{}
}

func (interp *Interpreter) execBlock(s *ast.BlockStatement, env *runtime.Environment) (*runtime.Value, signal) {
	blockEnv := runtime.NewEnvironment(env, true)
	interp.hoist(s.Statements, blockEnv)
	var result *runtime.Value
	for _, stmt := range s.Statements {
		val, sig := interp.execStatement(stmt, blockEnv)
		if sig.typ != sigNone {
			return val, sig
		}
		if val != nil {
			result = val
		}
	}
	return result, signal{}
}

func (interp *Interpreter) execReturn(s *ast.ReturnStatement, env *runtime.Environment) (*runtime.Value, signal) {
	if s.Value == nil {
		return nil, signal{typ: sigReturn, value: runtime.Undefined}
	}
	val, sig := interp.evalExpression(s.Value, env)
	if sig.typ != sigNone {
		return nil, sig
	}
	return nil, signal{typ: sigReturn, value: val}
}

func (interp *Interpreter) execIf(s *ast.IfStatement, env *runtime.Environment) (*runtime.Value, signal) {
	cond, sig := interp.evalExpression(s.Condition, env)
	if sig.typ != sigNone {
		return nil, sig
	}
	if cond.ToBoolean() {
		return interp.execStatement(s.Consequence, env)
	}
	if s.Alternative != nil {
		return interp.execStatement(s.Alternative, env)
	}
	return nil, signal{}
}

func (interp *Interpreter) execWhile(s *ast.WhileStatement, env *runtime.Environment) (*runtime.Value, signal) {
	var result *runtime.Value
	for {
		cond, sig := interp.evalExpression(s.Condition, env)
		if sig.typ != sigNone {
			return nil, sig
		}
		if !cond.ToBoolean() {
			break
		}
		val, sig := interp.execStatement(s.Body, env)
		if sig.typ == sigBreak {
			if sig.label != "" {
				return val, sig // propagate labeled break
			}
			break
		}
		if sig.typ == sigContinue {
			if sig.label != "" {
				return val, sig // propagate labeled continue
			}
			continue
		}
		if sig.typ != sigNone {
			return val, sig
		}
		if val != nil {
			result = val
		}
	}
	return result, signal{}
}

func (interp *Interpreter) execDoWhile(s *ast.DoWhileStatement, env *runtime.Environment) (*runtime.Value, signal) {
	var result *runtime.Value
	for {
		val, sig := interp.execStatement(s.Body, env)
		if sig.typ == sigBreak {
			if sig.label != "" {
				return val, sig
			}
			break
		}
		if sig.typ == sigContinue {
			if sig.label != "" {
				return val, sig
			}
			// continue goes to condition check
		} else if sig.typ != sigNone {
			return val, sig
		}
		if val != nil {
			result = val
		}
		cond, sig := interp.evalExpression(s.Condition, env)
		if sig.typ != sigNone {
			return nil, sig
		}
		if !cond.ToBoolean() {
			break
		}
	}
	return result, signal{}
}

func (interp *Interpreter) execFor(s *ast.ForStatement, env *runtime.Environment) (*runtime.Value, signal) {
	forEnv := runtime.NewEnvironment(env, true)

	if s.Init != nil {
		switch init := s.Init.(type) {
		case ast.Statement:
			_, sig := interp.execStatement(init, forEnv)
			if sig.typ != sigNone {
				return nil, sig
			}
		case ast.Expression:
			_, sig := interp.evalExpression(init, forEnv)
			if sig.typ != sigNone {
				return nil, sig
			}
		}
	}

	var result *runtime.Value
	for {
		if s.Test != nil {
			cond, sig := interp.evalExpression(s.Test, forEnv)
			if sig.typ != sigNone {
				return nil, sig
			}
			if !cond.ToBoolean() {
				break
			}
		}

		val, sig := interp.execStatement(s.Body, forEnv)
		if sig.typ == sigBreak {
			if sig.label != "" {
				return val, sig
			}
			break
		}
		if sig.typ == sigContinue {
			if sig.label != "" {
				return val, sig
			}
			// fall through to update
		} else if sig.typ != sigNone {
			return val, sig
		}
		if val != nil {
			result = val
		}

		if s.Update != nil {
			_, sig := interp.evalExpression(s.Update, forEnv)
			if sig.typ != sigNone {
				return nil, sig
			}
		}
	}
	return result, signal{}
}

func (interp *Interpreter) execForIn(s *ast.ForInStatement, env *runtime.Environment) (*runtime.Value, signal) {
	rightVal, sig := interp.evalExpression(s.Right, env)
	if sig.typ != sigNone {
		return nil, sig
	}

	if rightVal.Type != runtime.TypeObject || rightVal.Object == nil {
		return nil, signal{}
	}

	keys := interp.getEnumerableKeys(rightVal.Object)

	var result *runtime.Value
	for _, key := range keys {
		loopEnv := runtime.NewEnvironment(env, true)
		interp.assignLoopVar(s.Left, runtime.NewString(key), loopEnv)

		val, sig := interp.execStatement(s.Body, loopEnv)
		if sig.typ == sigBreak {
			if sig.label != "" {
				return val, sig
			}
			break
		}
		if sig.typ == sigContinue {
			if sig.label != "" {
				return val, sig
			}
			continue
		}
		if sig.typ != sigNone {
			return val, sig
		}
		if val != nil {
			result = val
		}
	}
	return result, signal{}
}

func (interp *Interpreter) getEnumerableKeys(obj *runtime.Object) []string {
	var keys []string
	if obj.OType == runtime.ObjTypeArray {
		for i := range obj.ArrayData {
			keys = append(keys, strconv.Itoa(i))
		}
	}
	for k, prop := range obj.Properties {
		if prop.Enumerable && k != "length" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

func (interp *Interpreter) execForOf(s *ast.ForOfStatement, env *runtime.Environment) (*runtime.Value, signal) {
	rightVal, sig := interp.evalExpression(s.Right, env)
	if sig.typ != sigNone {
		return nil, sig
	}

	var elements []*runtime.Value
	if rightVal.Type == runtime.TypeObject && rightVal.Object != nil {
		if rightVal.Object.OType == runtime.ObjTypeArray {
			elements = rightVal.Object.ArrayData
		} else if rightVal.Object.IteratorNext != nil {
			for {
				val, done := rightVal.Object.IteratorNext()
				if done {
					break
				}
				elements = append(elements, val)
			}
		}
	} else if rightVal.Type == runtime.TypeString {
		for _, ch := range rightVal.Str {
			elements = append(elements, runtime.NewString(string(ch)))
		}
	}

	var result *runtime.Value
	for _, elem := range elements {
		loopEnv := runtime.NewEnvironment(env, true)
		interp.assignLoopVar(s.Left, elem, loopEnv)

		val, sig := interp.execStatement(s.Body, loopEnv)
		if sig.typ == sigBreak {
			if sig.label != "" {
				return val, sig
			}
			break
		}
		if sig.typ == sigContinue {
			if sig.label != "" {
				return val, sig
			}
			continue
		}
		if sig.typ != sigNone {
			return val, sig
		}
		if val != nil {
			result = val
		}
	}
	return result, signal{}
}

func (interp *Interpreter) assignLoopVar(left ast.Node, val *runtime.Value, env *runtime.Environment) {
	switch l := left.(type) {
	case *ast.VariableDeclaration:
		if len(l.Declarations) > 0 {
			interp.bindPattern(l.Declarations[0].Name, val, l.Kind, env)
		}
	case ast.Expression:
		interp.assignToExpression(l, val, env)
	}
}

func (interp *Interpreter) execSwitch(s *ast.SwitchStatement, env *runtime.Environment) (*runtime.Value, signal) {
	disc, sig := interp.evalExpression(s.Discriminant, env)
	if sig.typ != sigNone {
		return nil, sig
	}

	switchEnv := runtime.NewEnvironment(env, true)
	matched := false
	defaultIdx := -1

	for i, c := range s.Cases {
		if c.Test == nil {
			defaultIdx = i
			continue
		}
		testVal, sig := interp.evalExpression(c.Test, switchEnv)
		if sig.typ != sigNone {
			return nil, sig
		}
		if runtime.StrictEquals(disc, testVal) {
			matched = true
		}
		if matched {
			for _, stmt := range c.Consequent {
				val, sig := interp.execStatement(stmt, switchEnv)
				if sig.typ == sigBreak {
					return val, signal{}
				}
				if sig.typ != sigNone {
					return val, sig
				}
			}
		}
	}

	if !matched && defaultIdx >= 0 {
		for i := defaultIdx; i < len(s.Cases); i++ {
			for _, stmt := range s.Cases[i].Consequent {
				val, sig := interp.execStatement(stmt, switchEnv)
				if sig.typ == sigBreak {
					return val, signal{}
				}
				if sig.typ != sigNone {
					return val, sig
				}
			}
		}
	}

	return nil, signal{}
}

func (interp *Interpreter) execThrow(s *ast.ThrowStatement, env *runtime.Environment) (*runtime.Value, signal) {
	val, sig := interp.evalExpression(s.Argument, env)
	if sig.typ != sigNone {
		return nil, sig
	}
	return nil, signal{typ: sigThrow, value: val}
}

func (interp *Interpreter) execTry(s *ast.TryStatement, env *runtime.Environment) (*runtime.Value, signal) {
	val, sig := interp.execBlock(s.Block, env)

	if sig.typ == sigThrow && s.Handler != nil {
		catchEnv := runtime.NewEnvironment(env, true)
		if s.Handler.Param != nil {
			interp.bindPattern(s.Handler.Param, sig.value, "let", catchEnv)
		}
		interp.hoist(s.Handler.Body.Statements, catchEnv)
		var catchResult *runtime.Value
		for _, stmt := range s.Handler.Body.Statements {
			r, csig := interp.execStatement(stmt, catchEnv)
			if csig.typ != sigNone {
				val = r
				sig = csig
				goto finalizer
			}
			if r != nil {
				catchResult = r
			}
		}
		val = catchResult
		sig = signal{}
	}

finalizer:
	if s.Finalizer != nil {
		fval, fsig := interp.execBlock(s.Finalizer, env)
		_ = fval
		if fsig.typ != sigNone {
			return nil, fsig
		}
	}

	return val, sig
}

func (interp *Interpreter) execLabeled(s *ast.LabeledStatement, env *runtime.Environment) (*runtime.Value, signal) {
	val, sig := interp.execStatement(s.Body, env)
	if (sig.typ == sigBreak || sig.typ == sigContinue) && sig.label == s.Label.Value {
		return val, signal{}
	}
	return val, sig
}

func (interp *Interpreter) execClassDecl(s *ast.ClassDeclaration, env *runtime.Environment) (*runtime.Value, signal) {
	classVal, sig := interp.buildClass(s.Name, s.SuperClass, s.Body, env)
	if sig.typ != sigNone {
		return nil, sig
	}
	if err := env.Declare(s.Name.Value, "let", classVal); err != nil {
		return nil, signal{typ: sigThrow, value: errorFromGoError(err, env)}
	}
	return nil, signal{}
}

func (interp *Interpreter) buildClass(name *ast.Identifier, superExpr ast.Expression, body *ast.ClassBody, env *runtime.Environment) (*runtime.Value, signal) {
	var superProto *runtime.Object
	var superConstructor runtime.CallableFunc
	if superExpr != nil {
		superVal, sig := interp.evalExpression(superExpr, env)
		if sig.typ != sigNone {
			return nil, sig
		}
		if superVal.Type == runtime.TypeObject && superVal.Object != nil {
			superConstructor = superVal.Object.Callable
			protoProp := superVal.Object.Get("prototype")
			if protoProp.Type == runtime.TypeObject && protoProp.Object != nil {
				superProto = protoProp.Object
			}
		}
	}

	proto := runtime.NewOrdinaryObject(superProto)
	var constructorFn runtime.CallableFunc

	classObj := runtime.NewFunctionObject(nil, nil)
	classObj.Set("prototype", runtime.NewObject(proto))

	for _, method := range body.Methods {
		methodName := interp.getPropertyKey(method.Key, method.Computed, env)
		fn := interp.createFunctionFromExpr(method.Value, env)

		if method.Kind == "constructor" {
			constructorFn = interp.makeConstructor(method.Value, env, proto, superConstructor)
			continue
		}

		fnVal := fn
		target := proto
		if method.Static {
			target = classObj
		}

		if method.Kind == "get" {
			target.DefineProperty(methodName, &runtime.Property{
				Getter:       fnVal,
				IsAccessor:   true,
				Enumerable:   true,
				Configurable: true,
			})
		} else if method.Kind == "set" {
			if existing, ok := target.Properties[methodName]; ok && existing.IsAccessor {
				existing.Setter = fnVal
			} else {
				target.DefineProperty(methodName, &runtime.Property{
					Setter:       fnVal,
					IsAccessor:   true,
					Enumerable:   true,
					Configurable: true,
				})
			}
		} else {
			target.Set(methodName, fnVal)
		}
	}

	if constructorFn == nil {
		if superConstructor != nil {
			sc := superConstructor
			constructorFn = func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
				return sc(this, args)
			}
		} else {
			constructorFn = func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
				return this, nil
			}
		}
	}

	classObj.Callable = constructorFn
	classObj.Constructor = constructorFn

	proto.Set("constructor", runtime.NewObject(classObj))

	return runtime.NewObject(classObj), signal{}
}

func (interp *Interpreter) makeConstructor(fe *ast.FunctionExpression, env *runtime.Environment, proto *runtime.Object, superCtor runtime.CallableFunc) runtime.CallableFunc {
	return func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		fnEnv := runtime.NewEnvironment(env, false)

		// bind this
		fnEnv.Declare("this", "const", this)

		// super function
		if superCtor != nil {
			superFn := runtime.NewFunctionObject(nil, func(thisVal *runtime.Value, superArgs []*runtime.Value) (*runtime.Value, error) {
				return superCtor(this, superArgs)
			})
			fnEnv.Declare("super", "const", runtime.NewObject(superFn))
		}

		interp.bindFunctionParams(fe.Params, fe.Defaults, fe.Rest, args, fnEnv)
		interp.hoist(fe.Body.Statements, fnEnv)

		for _, stmt := range fe.Body.Statements {
			_, sig := interp.execStatement(stmt, fnEnv)
			if sig.typ == sigReturn {
				if sig.value != nil && sig.value.Type == runtime.TypeObject {
					return sig.value, nil
				}
				return this, nil
			}
			if sig.typ == sigThrow {
				return nil, &jsError{value: sig.value}
			}
		}
		return this, nil
	}
}

func (interp *Interpreter) getPropertyKey(key ast.Expression, computed bool, env *runtime.Environment) string {
	if computed {
		val, _ := interp.evalExpression(key, env)
		return val.ToString()
	}
	switch k := key.(type) {
	case *ast.Identifier:
		return k.Value
	case *ast.StringLiteral:
		return k.Value
	case *ast.NumberLiteral:
		return fmt.Sprintf("%g", k.Value)
	}
	return ""
}

// ---------- Expression evaluation ----------

func (interp *Interpreter) evalExpression(expr ast.Expression, env *runtime.Environment) (*runtime.Value, signal) {
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		return runtime.NewNumber(e.Value), signal{}
	case *ast.StringLiteral:
		return runtime.NewString(e.Value), signal{}
	case *ast.BooleanLiteral:
		return runtime.NewBool(e.Value), signal{}
	case *ast.NullLiteral:
		return runtime.Null, signal{}
	case *ast.UndefinedLiteral:
		return runtime.Undefined, signal{}
	case *ast.Identifier:
		return interp.evalIdentifier(e, env)
	case *ast.ThisExpression:
		val, err := env.Get("this")
		if err != nil {
			return runtime.Undefined, signal{}
		}
		return val, signal{}
	case *ast.ArrayLiteral:
		return interp.evalArrayLiteral(e, env)
	case *ast.ObjectLiteral:
		return interp.evalObjectLiteral(e, env)
	case *ast.FunctionExpression:
		return interp.createFunctionFromExpr(e, env), signal{}
	case *ast.ArrowFunctionExpression:
		return interp.createArrowFunction(e, env), signal{}
	case *ast.UnaryExpression:
		return interp.evalUnary(e, env)
	case *ast.UpdateExpression:
		return interp.evalUpdate(e, env)
	case *ast.BinaryExpression:
		return interp.evalBinary(e, env)
	case *ast.LogicalExpression:
		return interp.evalLogical(e, env)
	case *ast.AssignmentExpression:
		return interp.evalAssignment(e, env)
	case *ast.ConditionalExpression:
		return interp.evalConditional(e, env)
	case *ast.CallExpression:
		return interp.evalCall(e, env)
	case *ast.MemberExpression:
		return interp.evalMember(e, env)
	case *ast.NewExpression:
		return interp.evalNew(e, env)
	case *ast.SequenceExpression:
		return interp.evalSequence(e, env)
	case *ast.TemplateLiteralExpr:
		return interp.evalTemplateLiteral(e, env)
	case *ast.SpreadElement:
		return interp.evalExpression(e.Argument, env)
	case *ast.ClassExpression:
		return interp.buildClass(e.Name, e.SuperClass, e.Body, env)
	case *ast.SuperExpression:
		val, err := env.Get("super")
		if err != nil {
			return runtime.Undefined, signal{}
		}
		return val, signal{}
	case *ast.ComputedPropertyName:
		return interp.evalExpression(e.Expression, env)
	default:
		return runtime.Undefined, signal{typ: sigThrow, value: runtime.NewString(fmt.Sprintf("unsupported expression: %T", expr))}
	}
}

func (interp *Interpreter) evalIdentifier(e *ast.Identifier, env *runtime.Environment) (*runtime.Value, signal) {
	// Handle special globals
	switch e.Value {
	case "NaN":
		return runtime.NaN, signal{}
	case "Infinity":
		return runtime.PosInf, signal{}
	}
	val, err := env.Get(e.Value)
	if err != nil {
		return nil, signal{typ: sigThrow, value: errorFromGoError(err, env)}
	}
	return val, signal{}
}

func (interp *Interpreter) evalArrayLiteral(e *ast.ArrayLiteral, env *runtime.Environment) (*runtime.Value, signal) {
	var elements []*runtime.Value
	for _, elem := range e.Elements {
		if elem == nil {
			elements = append(elements, runtime.Undefined)
			continue
		}
		if spread, ok := elem.(*ast.SpreadElement); ok {
			arrVal, sig := interp.evalExpression(spread.Argument, env)
			if sig.typ != sigNone {
				return nil, sig
			}
			if arrVal.Type == runtime.TypeObject && arrVal.Object != nil && arrVal.Object.OType == runtime.ObjTypeArray {
				elements = append(elements, arrVal.Object.ArrayData...)
			}
			continue
		}
		val, sig := interp.evalExpression(elem, env)
		if sig.typ != sigNone {
			return nil, sig
		}
		elements = append(elements, val)
	}
	arr := runtime.NewArrayObject(nil, elements)
	return runtime.NewObject(arr), signal{}
}

func (interp *Interpreter) evalObjectLiteral(e *ast.ObjectLiteral, env *runtime.Environment) (*runtime.Value, signal) {
	obj := runtime.NewOrdinaryObject(nil)
	for _, prop := range e.Properties {
		if spread, ok := prop.Key.(*ast.SpreadElement); ok {
			srcVal, sig := interp.evalExpression(spread.Argument, env)
			if sig.typ != sigNone {
				return nil, sig
			}
			if srcVal.Type == runtime.TypeObject && srcVal.Object != nil {
				for k, v := range srcVal.Object.Properties {
					if v.Enumerable {
						obj.Set(k, v.Value)
					}
				}
			}
			continue
		}

		key := interp.getPropertyKey(prop.Key, prop.Computed, env)

		if prop.Kind == "get" || prop.Kind == "set" {
			fnVal, sig := interp.evalExpression(prop.Value, env)
			if sig.typ != sigNone {
				return nil, sig
			}
			if prop.Kind == "get" {
				existing := obj.Properties[key]
				if existing != nil && existing.IsAccessor {
					existing.Getter = fnVal
				} else {
					obj.DefineProperty(key, &runtime.Property{
						Getter:       fnVal,
						IsAccessor:   true,
						Enumerable:   true,
						Configurable: true,
					})
				}
			} else {
				existing := obj.Properties[key]
				if existing != nil && existing.IsAccessor {
					existing.Setter = fnVal
				} else {
					obj.DefineProperty(key, &runtime.Property{
						Setter:       fnVal,
						IsAccessor:   true,
						Enumerable:   true,
						Configurable: true,
					})
				}
			}
			continue
		}

		if prop.Shorthand {
			val, sig := interp.evalExpression(prop.Key, env)
			if sig.typ != sigNone {
				return nil, sig
			}
			obj.Set(key, val)
			continue
		}

		val, sig := interp.evalExpression(prop.Value, env)
		if sig.typ != sigNone {
			return nil, sig
		}
		obj.Set(key, val)
	}
	return runtime.NewObject(obj), signal{}
}

func (interp *Interpreter) createFunction(name *ast.Identifier, params []ast.Expression, defaults []ast.Expression, rest ast.Expression, body *ast.BlockStatement, env *runtime.Environment, isArrow bool) *runtime.Value {
	closureEnv := env
	var fnName string
	if name != nil {
		fnName = name.Value
	}

	var callable runtime.CallableFunc
	callable = func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		fnEnv := runtime.NewEnvironment(closureEnv, false)

		if !isArrow {
			fnEnv.Declare("this", "const", this)
			// arguments object
			argsArr := runtime.NewArrayObject(nil, args)
			fnEnv.Declare("arguments", "var", runtime.NewObject(argsArr))
		}

		if fnName != "" {
			fnObj := runtime.NewFunctionObject(nil, callable)
			fnEnv.Declare(fnName, "const", runtime.NewObject(fnObj))
		}

		interp.bindFunctionParams(params, defaults, rest, args, fnEnv)
		interp.hoist(body.Statements, fnEnv)

		for _, stmt := range body.Statements {
			_, sig := interp.execStatement(stmt, fnEnv)
			if sig.typ == sigReturn {
				return sig.value, nil
			}
			if sig.typ == sigThrow {
				return nil, &jsError{value: sig.value}
			}
		}
		return runtime.Undefined, nil
	}

	fnObj := runtime.NewFunctionObject(nil, callable)
	fnObj.Set("prototype", runtime.NewObject(runtime.NewOrdinaryObject(runtime.DefaultObjectPrototype)))
	if fnName != "" {
		fnObj.Set("name", runtime.NewString(fnName))
	}
	fnObj.Set("length", runtime.NewNumber(float64(len(params))))

	return runtime.NewObject(fnObj)
}

func (interp *Interpreter) createFunctionFromExpr(e *ast.FunctionExpression, env *runtime.Environment) *runtime.Value {
	return interp.createFunction(e.Name, e.Params, e.Defaults, e.Rest, e.Body, env, false)
}

func (interp *Interpreter) createArrowFunction(e *ast.ArrowFunctionExpression, env *runtime.Environment) *runtime.Value {
	closureEnv := env

	var callable runtime.CallableFunc
	callable = func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		fnEnv := runtime.NewEnvironment(closureEnv, false)

		interp.bindFunctionParams(e.Params, e.Defaults, e.Rest, args, fnEnv)

		switch body := e.Body.(type) {
		case *ast.BlockStatement:
			interp.hoist(body.Statements, fnEnv)
			for _, stmt := range body.Statements {
				_, sig := interp.execStatement(stmt, fnEnv)
				if sig.typ == sigReturn {
					return sig.value, nil
				}
				if sig.typ == sigThrow {
					return nil, &jsError{value: sig.value}
				}
			}
			return runtime.Undefined, nil
		case ast.Expression:
			val, sig := interp.evalExpression(body, fnEnv)
			if sig.typ == sigThrow {
				return nil, &jsError{value: sig.value}
			}
			return val, nil
		}
		return runtime.Undefined, nil
	}

	fnObj := runtime.NewFunctionObject(nil, callable)
	fnObj.Set("length", runtime.NewNumber(float64(len(e.Params))))
	return runtime.NewObject(fnObj)
}

func (interp *Interpreter) bindFunctionParams(params []ast.Expression, defaults []ast.Expression, rest ast.Expression, args []*runtime.Value, env *runtime.Environment) {
	for i, param := range params {
		var val *runtime.Value
		if i < len(args) {
			val = args[i]
		} else {
			val = runtime.Undefined
		}
		if val.Type == runtime.TypeUndefined && i < len(defaults) && defaults[i] != nil {
			defVal, sig := interp.evalExpression(defaults[i], env)
			if sig.typ == sigNone {
				val = defVal
			}
		}
		interp.bindPattern(param, val, "let", env)
	}
	if rest != nil {
		var restArgs []*runtime.Value
		if len(params) < len(args) {
			restArgs = args[len(params):]
		}
		restArr := runtime.NewArrayObject(nil, restArgs)
		restVal := runtime.NewObject(restArr)
		// rest may be *ast.RestElement wrapping an Identifier
		if re, ok := rest.(*ast.RestElement); ok {
			interp.bindPattern(re.Argument, restVal, "let", env)
		} else {
			interp.bindPattern(rest, restVal, "let", env)
		}
	}
}

func (interp *Interpreter) evalUnary(e *ast.UnaryExpression, env *runtime.Environment) (*runtime.Value, signal) {
	if e.Operator == "typeof" {
		if ident, ok := e.Operand.(*ast.Identifier); ok {
			val, err := env.Get(ident.Value)
			if err != nil {
				return runtime.NewString("undefined"), signal{}
			}
			return interp.typeofValue(val), signal{}
		}
		val, sig := interp.evalExpression(e.Operand, env)
		if sig.typ != sigNone {
			return nil, sig
		}
		return interp.typeofValue(val), signal{}
	}

	if e.Operator == "delete" {
		if member, ok := e.Operand.(*ast.MemberExpression); ok {
			objVal, sig := interp.evalExpression(member.Object, env)
			if sig.typ != sigNone {
				return nil, sig
			}
			if objVal.Type == runtime.TypeObject && objVal.Object != nil {
				key := interp.resolveMemberKey(member, env)
				delete(objVal.Object.Properties, key)
				return runtime.True, signal{}
			}
		}
		return runtime.True, signal{}
	}

	operand, sig := interp.evalExpression(e.Operand, env)
	if sig.typ != sigNone {
		return nil, sig
	}

	switch e.Operator {
	case "-":
		return runtime.NewNumber(-operand.ToNumber()), signal{}
	case "+":
		return runtime.NewNumber(operand.ToNumber()), signal{}
	case "!":
		return runtime.NewBool(!operand.ToBoolean()), signal{}
	case "~":
		n := int32(operand.ToNumber())
		return runtime.NewNumber(float64(^n)), signal{}
	case "void":
		return runtime.Undefined, signal{}
	}
	return runtime.Undefined, signal{}
}

func (interp *Interpreter) typeofValue(val *runtime.Value) *runtime.Value {
	if val == nil {
		return runtime.NewString("undefined")
	}
	switch val.Type {
	case runtime.TypeUndefined:
		return runtime.NewString("undefined")
	case runtime.TypeNull:
		return runtime.NewString("object")
	case runtime.TypeBoolean:
		return runtime.NewString("boolean")
	case runtime.TypeNumber:
		return runtime.NewString("number")
	case runtime.TypeString:
		return runtime.NewString("string")
	case runtime.TypeObject:
		if val.Object != nil && val.Object.Callable != nil {
			return runtime.NewString("function")
		}
		return runtime.NewString("object")
	}
	return runtime.NewString("undefined")
}

func (interp *Interpreter) evalUpdate(e *ast.UpdateExpression, env *runtime.Environment) (*runtime.Value, signal) {
	old, sig := interp.evalExpression(e.Operand, env)
	if sig.typ != sigNone {
		return nil, sig
	}
	oldNum := old.ToNumber()
	var newNum float64
	if e.Operator == "++" {
		newNum = oldNum + 1
	} else {
		newNum = oldNum - 1
	}
	newVal := runtime.NewNumber(newNum)
	asig := interp.assignToExpression(e.Operand, newVal, env)
	if asig.typ != sigNone {
		return nil, asig
	}

	if e.Prefix {
		return newVal, signal{}
	}
	return runtime.NewNumber(oldNum), signal{}
}

func (interp *Interpreter) evalBinary(e *ast.BinaryExpression, env *runtime.Environment) (*runtime.Value, signal) {
	left, sig := interp.evalExpression(e.Left, env)
	if sig.typ != sigNone {
		return nil, sig
	}
	right, sig := interp.evalExpression(e.Right, env)
	if sig.typ != sigNone {
		return nil, sig
	}

	switch e.Operator {
	case "+":
		if left.Type == runtime.TypeString || right.Type == runtime.TypeString {
			return runtime.NewString(left.ToString() + right.ToString()), signal{}
		}
		return runtime.NewNumber(left.ToNumber() + right.ToNumber()), signal{}
	case "-":
		return runtime.NewNumber(left.ToNumber() - right.ToNumber()), signal{}
	case "*":
		return runtime.NewNumber(left.ToNumber() * right.ToNumber()), signal{}
	case "/":
		rn := right.ToNumber()
		if rn == 0 {
			ln := left.ToNumber()
			if math.IsNaN(ln) {
				return runtime.NaN, signal{}
			}
			if ln == 0 {
				return runtime.NaN, signal{}
			}
			if ln > 0 {
				return runtime.PosInf, signal{}
			}
			return runtime.NegInf, signal{}
		}
		return runtime.NewNumber(left.ToNumber() / rn), signal{}
	case "%":
		rn := right.ToNumber()
		if rn == 0 {
			return runtime.NaN, signal{}
		}
		return runtime.NewNumber(math.Mod(left.ToNumber(), rn)), signal{}
	case "**":
		return runtime.NewNumber(math.Pow(left.ToNumber(), right.ToNumber())), signal{}
	case "==":
		return runtime.NewBool(runtime.AbstractEquals(left, right)), signal{}
	case "!=":
		return runtime.NewBool(!runtime.AbstractEquals(left, right)), signal{}
	case "===":
		return runtime.NewBool(runtime.StrictEquals(left, right)), signal{}
	case "!==":
		return runtime.NewBool(!runtime.StrictEquals(left, right)), signal{}
	case "<":
		return interp.compareValues(left, right, false, false), signal{}
	case ">":
		return interp.compareValues(right, left, false, false), signal{}
	case "<=":
		return interp.compareValues(right, left, true, true), signal{}
	case ">=":
		return interp.compareValues(left, right, true, true), signal{}
	case "&":
		return runtime.NewNumber(float64(int32(left.ToNumber()) & int32(right.ToNumber()))), signal{}
	case "|":
		return runtime.NewNumber(float64(int32(left.ToNumber()) | int32(right.ToNumber()))), signal{}
	case "^":
		return runtime.NewNumber(float64(int32(left.ToNumber()) ^ int32(right.ToNumber()))), signal{}
	case "<<":
		return runtime.NewNumber(float64(int32(left.ToNumber()) << (uint32(right.ToNumber()) & 0x1f))), signal{}
	case ">>":
		return runtime.NewNumber(float64(int32(left.ToNumber()) >> (uint32(right.ToNumber()) & 0x1f))), signal{}
	case ">>>":
		return runtime.NewNumber(float64(uint32(left.ToNumber()) >> (uint32(right.ToNumber()) & 0x1f))), signal{}
	case "instanceof":
		return interp.evalInstanceof(left, right), signal{}
	case "in":
		return interp.evalIn(left, right), signal{}
	case "??":
		if left.Type == runtime.TypeNull || left.Type == runtime.TypeUndefined {
			return right, signal{}
		}
		return left, signal{}
	}
	return runtime.Undefined, signal{}
}

func (interp *Interpreter) compareValues(left, right *runtime.Value, invert, negate bool) *runtime.Value {
	if left.Type == runtime.TypeString && right.Type == runtime.TypeString {
		if invert {
			return runtime.NewBool(!(left.Str < right.Str))
		}
		return runtime.NewBool(left.Str < right.Str)
	}
	ln := left.ToNumber()
	rn := right.ToNumber()
	if math.IsNaN(ln) || math.IsNaN(rn) {
		return runtime.False
	}
	result := ln < rn
	if invert {
		if negate {
			return runtime.NewBool(!result)
		}
	}
	return runtime.NewBool(result)
}

func (interp *Interpreter) evalInstanceof(left, right *runtime.Value) *runtime.Value {
	if right.Type != runtime.TypeObject || right.Object == nil || right.Object.Callable == nil {
		return runtime.False
	}
	if left.Type != runtime.TypeObject || left.Object == nil {
		return runtime.False
	}

	protoProp := right.Object.Get("prototype")
	if protoProp.Type != runtime.TypeObject || protoProp.Object == nil {
		return runtime.False
	}

	proto := left.Object.Prototype
	for proto != nil {
		if proto == protoProp.Object {
			return runtime.True
		}
		proto = proto.Prototype
	}
	return runtime.False
}

func (interp *Interpreter) evalIn(left, right *runtime.Value) *runtime.Value {
	if right.Type != runtime.TypeObject || right.Object == nil {
		return runtime.False
	}
	key := left.ToString()
	if right.Object.OType == runtime.ObjTypeArray {
		idx, err := strconv.Atoi(key)
		if err == nil && idx >= 0 && idx < len(right.Object.ArrayData) {
			return runtime.True
		}
	}
	return runtime.NewBool(right.Object.HasProperty(key))
}

func (interp *Interpreter) evalLogical(e *ast.LogicalExpression, env *runtime.Environment) (*runtime.Value, signal) {
	left, sig := interp.evalExpression(e.Left, env)
	if sig.typ != sigNone {
		return nil, sig
	}

	switch e.Operator {
	case "&&":
		if !left.ToBoolean() {
			return left, signal{}
		}
		return interp.evalExpression(e.Right, env)
	case "||":
		if left.ToBoolean() {
			return left, signal{}
		}
		return interp.evalExpression(e.Right, env)
	case "??":
		if left.Type != runtime.TypeNull && left.Type != runtime.TypeUndefined {
			return left, signal{}
		}
		return interp.evalExpression(e.Right, env)
	}
	return left, signal{}
}

func (interp *Interpreter) evalAssignment(e *ast.AssignmentExpression, env *runtime.Environment) (*runtime.Value, signal) {
	right, sig := interp.evalExpression(e.Right, env)
	if sig.typ != sigNone {
		return nil, sig
	}

	if e.Operator != "=" {
		old, sig := interp.evalExpression(e.Left, env)
		if sig.typ != sigNone {
			return nil, sig
		}
		right = interp.applyCompoundOp(e.Operator, old, right)
	}

	// handle destructuring on the left
	switch left := e.Left.(type) {
	case *ast.ObjectPattern:
		interp.destructureAssign(left, right, env)
		return right, signal{}
	case *ast.ArrayPattern:
		interp.destructureAssignArray(left, right, env)
		return right, signal{}
	}

	asig := interp.assignToExpression(e.Left, right, env)
	if asig.typ != sigNone {
		return nil, asig
	}
	return right, signal{}
}

func (interp *Interpreter) destructureAssign(pattern *ast.ObjectPattern, val *runtime.Value, env *runtime.Environment) {
	if val.Type != runtime.TypeObject || val.Object == nil {
		return
	}
	for _, prop := range pattern.Properties {
		var key string
		if ident, ok := prop.Key.(*ast.Identifier); ok {
			key = ident.Value
		}
		propVal := val.Object.Get(key)
		target := prop.Value
		if target == nil {
			target = prop.Key
		}
		if ap, ok := target.(*ast.AssignmentPattern); ok {
			if propVal == nil || propVal.Type == runtime.TypeUndefined {
				defVal, sig := interp.evalExpression(ap.Right, env)
				if sig.typ == sigNone {
					propVal = defVal
				}
			}
			target = ap.Left
		}
		interp.assignToExpression(target, propVal, env)
	}
}

func (interp *Interpreter) destructureAssignArray(pattern *ast.ArrayPattern, val *runtime.Value, env *runtime.Environment) {
	var elements []*runtime.Value
	if val.Type == runtime.TypeObject && val.Object != nil && val.Object.OType == runtime.ObjTypeArray {
		elements = val.Object.ArrayData
	}
	for i, elem := range pattern.Elements {
		if elem == nil {
			continue
		}
		var elemVal *runtime.Value
		if i < len(elements) {
			elemVal = elements[i]
		} else {
			elemVal = runtime.Undefined
		}
		if rest, ok := elem.(*ast.RestElement); ok {
			var restElems []*runtime.Value
			if i < len(elements) {
				restElems = elements[i:]
			}
			restArr := runtime.NewArrayObject(nil, restElems)
			interp.assignToExpression(rest.Argument, runtime.NewObject(restArr), env)
			continue
		}
		interp.assignToExpression(elem, elemVal, env)
	}
}

func (interp *Interpreter) applyCompoundOp(op string, left, right *runtime.Value) *runtime.Value {
	switch op {
	case "+=":
		if left.Type == runtime.TypeString || right.Type == runtime.TypeString {
			return runtime.NewString(left.ToString() + right.ToString())
		}
		return runtime.NewNumber(left.ToNumber() + right.ToNumber())
	case "-=":
		return runtime.NewNumber(left.ToNumber() - right.ToNumber())
	case "*=":
		return runtime.NewNumber(left.ToNumber() * right.ToNumber())
	case "/=":
		return runtime.NewNumber(left.ToNumber() / right.ToNumber())
	case "%=":
		return runtime.NewNumber(math.Mod(left.ToNumber(), right.ToNumber()))
	case "**=":
		return runtime.NewNumber(math.Pow(left.ToNumber(), right.ToNumber()))
	case "&=":
		return runtime.NewNumber(float64(int32(left.ToNumber()) & int32(right.ToNumber())))
	case "|=":
		return runtime.NewNumber(float64(int32(left.ToNumber()) | int32(right.ToNumber())))
	case "^=":
		return runtime.NewNumber(float64(int32(left.ToNumber()) ^ int32(right.ToNumber())))
	case "<<=":
		return runtime.NewNumber(float64(int32(left.ToNumber()) << (uint32(right.ToNumber()) & 0x1f)))
	case ">>=":
		return runtime.NewNumber(float64(int32(left.ToNumber()) >> (uint32(right.ToNumber()) & 0x1f)))
	case ">>>=":
		return runtime.NewNumber(float64(uint32(left.ToNumber()) >> (uint32(right.ToNumber()) & 0x1f)))
	case "??=":
		if left.Type == runtime.TypeNull || left.Type == runtime.TypeUndefined {
			return right
		}
		return left
	case "&&=":
		if left.ToBoolean() {
			return right
		}
		return left
	case "||=":
		if left.ToBoolean() {
			return left
		}
		return right
	}
	return right
}

func (interp *Interpreter) assignToExpression(expr ast.Expression, val *runtime.Value, env *runtime.Environment) signal {
	switch e := expr.(type) {
	case *ast.Identifier:
		err := env.Set(e.Value, val)
		if err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "TypeError") {
				return signal{typ: sigThrow, value: errorFromGoError(err, env)}
			}
			// might be undeclared in global scope; set in function scope
			funcScope := env.GetFunctionScope()
			funcScope.SetInCurrentScope(e.Value, val)
		}
	case *ast.MemberExpression:
		obj, sig := interp.evalExpression(e.Object, env)
		if sig.typ != sigNone {
			return sig
		}
		if obj.Type == runtime.TypeObject && obj.Object != nil {
			key := interp.resolveMemberKey(e, env)
			if obj.Object.OType == runtime.ObjTypeArray {
				idx, err := strconv.Atoi(key)
				if err == nil && idx >= 0 {
					for len(obj.Object.ArrayData) <= idx {
						obj.Object.ArrayData = append(obj.Object.ArrayData, runtime.Undefined)
					}
					obj.Object.ArrayData[idx] = val
					obj.Object.Set("length", runtime.NewNumber(float64(len(obj.Object.ArrayData))))
					return signal{}
				}
			}
			obj.Object.Set(key, val)
		}
	}
	return signal{}
}

func (interp *Interpreter) resolveMemberKey(e *ast.MemberExpression, env *runtime.Environment) string {
	if e.Computed {
		keyVal, _ := interp.evalExpression(e.Property, env)
		return keyVal.ToString()
	}
	if ident, ok := e.Property.(*ast.Identifier); ok {
		return ident.Value
	}
	return ""
}

func (interp *Interpreter) evalConditional(e *ast.ConditionalExpression, env *runtime.Environment) (*runtime.Value, signal) {
	test, sig := interp.evalExpression(e.Test, env)
	if sig.typ != sigNone {
		return nil, sig
	}
	if test.ToBoolean() {
		return interp.evalExpression(e.Consequent, env)
	}
	return interp.evalExpression(e.Alternate, env)
}

func (interp *Interpreter) evalCall(e *ast.CallExpression, env *runtime.Environment) (*runtime.Value, signal) {
	// Handle super() calls
	if _, ok := e.Callee.(*ast.SuperExpression); ok {
		return interp.evalSuperCall(e, env)
	}

	// Handle direct eval() calls — eval needs access to the calling scope
	if ident, ok := e.Callee.(*ast.Identifier); ok && ident.Value == "eval" {
		val, err := env.Get("eval")
		if err == nil && val.Type == runtime.TypeObject && val.Object != nil &&
			val.Object.Internal != nil && val.Object.Internal["isBuiltinEval"] != nil {
			args, argSig := interp.evalArguments(e.Arguments, env)
			if argSig.typ != sigNone {
				return nil, argSig
			}
			return interp.evalDirectEval(args, env)
		}
	}

	var thisVal *runtime.Value
	var callee *runtime.Value
	var sig signal

	// determine this binding
	if member, ok := e.Callee.(*ast.MemberExpression); ok {
		thisVal, sig = interp.evalExpression(member.Object, env)
		if sig.typ != sigNone {
			return nil, sig
		}
		if thisVal.Type == runtime.TypeObject && thisVal.Object != nil {
			key := interp.resolveMemberKey(member, env)
			// Check array methods first
			if thisVal.Object.OType == runtime.ObjTypeArray {
				method := interp.getArrayMethod(thisVal, key)
				if method != nil {
					callee = method
				} else {
					callee = thisVal.Object.Get(key)
				}
			} else {
				callee = thisVal.Object.Get(key)
			}
		} else if thisVal.Type == runtime.TypeString {
			callee = interp.getStringMethod(thisVal, member, env)
		} else {
			callee = runtime.Undefined
		}
	} else {
		callee, sig = interp.evalExpression(e.Callee, env)
		if sig.typ != sigNone {
			return nil, sig
		}
		thisVal = runtime.Undefined
	}

	if callee == nil || callee.Type != runtime.TypeObject || callee.Object == nil || callee.Object.Callable == nil {
		name := ""
		if ident, ok := e.Callee.(*ast.Identifier); ok {
			name = ident.Value
		}
		return nil, signal{typ: sigThrow, value: makeErrorObject("TypeError", fmt.Sprintf("%s is not a function", name), env)}
	}

	// evaluate arguments
	args, argSig := interp.evalArguments(e.Arguments, env)
	if argSig.typ != sigNone {
		return nil, argSig
	}

	result, err := callee.Object.Callable(thisVal, args)
	if err != nil {
		if jsErr, ok := err.(*jsError); ok {
			return nil, signal{typ: sigThrow, value: jsErr.value}
		}
		return nil, signal{typ: sigThrow, value: errorFromGoError(err, env)}
	}
	if result == nil {
		result = runtime.Undefined
	}
	return result, signal{}
}

func (interp *Interpreter) evalSuperCall(e *ast.CallExpression, env *runtime.Environment) (*runtime.Value, signal) {
	superVal, err := env.Get("super")
	if err != nil {
		return nil, signal{typ: sigThrow, value: makeErrorObject("ReferenceError", "super is not defined", env)}
	}
	if superVal.Type != runtime.TypeObject || superVal.Object == nil || superVal.Object.Callable == nil {
		return nil, signal{typ: sigThrow, value: makeErrorObject("TypeError", "super is not a function", env)}
	}
	thisVal, _ := env.Get("this")

	args, argSig := interp.evalArguments(e.Arguments, env)
	if argSig.typ != sigNone {
		return nil, argSig
	}

	result, callErr := superVal.Object.Callable(thisVal, args)
	if callErr != nil {
		if jsErr, ok := callErr.(*jsError); ok {
			return nil, signal{typ: sigThrow, value: jsErr.value}
		}
		return nil, signal{typ: sigThrow, value: errorFromGoError(callErr, env)}
	}
	if result == nil {
		result = runtime.Undefined
	}
	return result, signal{}
}

func (interp *Interpreter) getStringMethod(strVal *runtime.Value, member *ast.MemberExpression, env *runtime.Environment) *runtime.Value {
	key := interp.resolveMemberKey(member, env)
	s := strVal.Str

	switch key {
	case "length":
		return runtime.NewNumber(float64(len(s)))
	case "charAt":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			idx := 0
			if len(args) > 0 {
				idx = int(args[0].ToNumber())
			}
			if idx < 0 || idx >= len(s) {
				return runtime.NewString(""), nil
			}
			return runtime.NewString(string(s[idx])), nil
		})
	case "indexOf":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 {
				return runtime.NewNumber(-1), nil
			}
			search := args[0].ToString()
			return runtime.NewNumber(float64(strings.Index(s, search))), nil
		})
	case "slice":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			start := 0
			end := len(s)
			if len(args) > 0 {
				start = int(args[0].ToNumber())
				if start < 0 {
					start = len(s) + start
				}
			}
			if len(args) > 1 {
				end = int(args[1].ToNumber())
				if end < 0 {
					end = len(s) + end
				}
			}
			if start < 0 {
				start = 0
			}
			if end > len(s) {
				end = len(s)
			}
			if start >= end {
				return runtime.NewString(""), nil
			}
			return runtime.NewString(s[start:end]), nil
		})
	case "toUpperCase":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			return runtime.NewString(strings.ToUpper(s)), nil
		})
	case "toLowerCase":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			return runtime.NewString(strings.ToLower(s)), nil
		})
	case "trim":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			return runtime.NewString(strings.TrimSpace(s)), nil
		})
	case "split":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			sep := ""
			if len(args) > 0 {
				sep = args[0].ToString()
			}
			parts := strings.Split(s, sep)
			var elems []*runtime.Value
			for _, p := range parts {
				elems = append(elems, runtime.NewString(p))
			}
			arr := runtime.NewArrayObject(nil, elems)
			return runtime.NewObject(arr), nil
		})
	case "includes":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 {
				return runtime.False, nil
			}
			search := args[0].ToString()
			return runtime.NewBool(strings.Contains(s, search)), nil
		})
	case "startsWith":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 {
				return runtime.False, nil
			}
			return runtime.NewBool(strings.HasPrefix(s, args[0].ToString())), nil
		})
	case "endsWith":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 {
				return runtime.False, nil
			}
			return runtime.NewBool(strings.HasSuffix(s, args[0].ToString())), nil
		})
	case "repeat":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			count := 0
			if len(args) > 0 {
				count = int(args[0].ToNumber())
			}
			if count < 0 {
				return nil, fmt.Errorf("RangeError: Invalid count value")
			}
			return runtime.NewString(strings.Repeat(s, count)), nil
		})
	case "replace":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) < 2 {
				return runtime.NewString(s), nil
			}
			search := args[0].ToString()
			replacement := args[1].ToString()
			return runtime.NewString(strings.Replace(s, search, replacement, 1)), nil
		})
	case "substring":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			start := 0
			end := len(s)
			if len(args) > 0 {
				start = int(args[0].ToNumber())
			}
			if len(args) > 1 {
				end = int(args[1].ToNumber())
			}
			if start < 0 {
				start = 0
			}
			if end > len(s) {
				end = len(s)
			}
			if start > end {
				start, end = end, start
			}
			return runtime.NewString(s[start:end]), nil
		})
	case "padStart":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			targetLen := 0
			padStr := " "
			if len(args) > 0 {
				targetLen = int(args[0].ToNumber())
			}
			if len(args) > 1 {
				padStr = args[1].ToString()
			}
			result := s
			for len(result) < targetLen {
				result = padStr + result
			}
			if len(result) > targetLen {
				result = result[len(result)-targetLen:]
			}
			return runtime.NewString(result), nil
		})
	case "padEnd":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			targetLen := 0
			padStr := " "
			if len(args) > 0 {
				targetLen = int(args[0].ToNumber())
			}
			if len(args) > 1 {
				padStr = args[1].ToString()
			}
			result := s
			for len(result) < targetLen {
				result = result + padStr
			}
			if len(result) > targetLen {
				result = result[:targetLen]
			}
			return runtime.NewString(result), nil
		})
	}

	// handle bracket access for string chars
	if member.Computed {
		keyVal, _ := interp.evalExpression(member.Property, env)
		idx := int(keyVal.ToNumber())
		if idx >= 0 && idx < len(s) {
			return runtime.NewString(string(s[idx]))
		}
		return runtime.Undefined
	}

	return runtime.Undefined
}

func (interp *Interpreter) makeNativeMethod(fn runtime.CallableFunc) *runtime.Value {
	fnObj := runtime.NewFunctionObject(nil, fn)
	return runtime.NewObject(fnObj)
}

func (interp *Interpreter) evalArguments(arguments []ast.Expression, env *runtime.Environment) ([]*runtime.Value, signal) {
	var args []*runtime.Value
	for _, arg := range arguments {
		if spread, ok := arg.(*ast.SpreadElement); ok {
			arrVal, sig := interp.evalExpression(spread.Argument, env)
			if sig.typ != sigNone {
				return nil, sig
			}
			if arrVal.Type == runtime.TypeObject && arrVal.Object != nil && arrVal.Object.OType == runtime.ObjTypeArray {
				args = append(args, arrVal.Object.ArrayData...)
			}
			continue
		}
		val, sig := interp.evalExpression(arg, env)
		if sig.typ != sigNone {
			return nil, sig
		}
		args = append(args, val)
	}
	return args, signal{}
}

func (interp *Interpreter) evalMember(e *ast.MemberExpression, env *runtime.Environment) (*runtime.Value, signal) {
	obj, sig := interp.evalExpression(e.Object, env)
	if sig.typ != sigNone {
		return nil, sig
	}

	if obj == nil || obj.Type == runtime.TypeUndefined || obj.Type == runtime.TypeNull {
		name := ""
		if ident, ok := e.Object.(*ast.Identifier); ok {
			name = ident.Value
		}
		return nil, signal{typ: sigThrow, value: makeErrorObject("TypeError", fmt.Sprintf("Cannot read properties of %s (reading '%s')", obj.ToString(), name), env)}
	}

	if obj.Type == runtime.TypeString {
		key := interp.resolveMemberKey(e, env)
		if key == "length" {
			return runtime.NewNumber(float64(len(obj.Str))), signal{}
		}
		methodVal := interp.getStringMethod(obj, e, env)
		return methodVal, signal{}
	}

	if obj.Type == runtime.TypeObject && obj.Object != nil {
		key := interp.resolveMemberKey(e, env)

		// array length and index access
		if obj.Object.OType == runtime.ObjTypeArray {
			if key == "length" {
				return runtime.NewNumber(float64(len(obj.Object.ArrayData))), signal{}
			}
			idx, err := strconv.Atoi(key)
			if err == nil && idx >= 0 && idx < len(obj.Object.ArrayData) {
				return obj.Object.ArrayData[idx], signal{}
			}
			// array methods
			method := interp.getArrayMethod(obj, key)
			if method != nil {
				return method, signal{}
			}
		}

		return obj.Object.Get(key), signal{}
	}

	return runtime.Undefined, signal{}
}

func (interp *Interpreter) getArrayMethod(arrVal *runtime.Value, method string) *runtime.Value {
	arr := arrVal.Object
	switch method {
	case "push":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			arr.ArrayData = append(arr.ArrayData, args...)
			arr.Set("length", runtime.NewNumber(float64(len(arr.ArrayData))))
			return runtime.NewNumber(float64(len(arr.ArrayData))), nil
		})
	case "pop":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(arr.ArrayData) == 0 {
				return runtime.Undefined, nil
			}
			last := arr.ArrayData[len(arr.ArrayData)-1]
			arr.ArrayData = arr.ArrayData[:len(arr.ArrayData)-1]
			arr.Set("length", runtime.NewNumber(float64(len(arr.ArrayData))))
			return last, nil
		})
	case "shift":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(arr.ArrayData) == 0 {
				return runtime.Undefined, nil
			}
			first := arr.ArrayData[0]
			arr.ArrayData = arr.ArrayData[1:]
			arr.Set("length", runtime.NewNumber(float64(len(arr.ArrayData))))
			return first, nil
		})
	case "unshift":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			arr.ArrayData = append(args, arr.ArrayData...)
			arr.Set("length", runtime.NewNumber(float64(len(arr.ArrayData))))
			return runtime.NewNumber(float64(len(arr.ArrayData))), nil
		})
	case "indexOf":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 {
				return runtime.NewNumber(-1), nil
			}
			for i, v := range arr.ArrayData {
				if runtime.StrictEquals(v, args[0]) {
					return runtime.NewNumber(float64(i)), nil
				}
			}
			return runtime.NewNumber(-1), nil
		})
	case "includes":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 {
				return runtime.False, nil
			}
			for _, v := range arr.ArrayData {
				if runtime.StrictEquals(v, args[0]) {
					return runtime.True, nil
				}
			}
			return runtime.False, nil
		})
	case "join":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			sep := ","
			if len(args) > 0 {
				sep = args[0].ToString()
			}
			var parts []string
			for _, v := range arr.ArrayData {
				parts = append(parts, v.ToString())
			}
			return runtime.NewString(strings.Join(parts, sep)), nil
		})
	case "reverse":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			for i, j := 0, len(arr.ArrayData)-1; i < j; i, j = i+1, j-1 {
				arr.ArrayData[i], arr.ArrayData[j] = arr.ArrayData[j], arr.ArrayData[i]
			}
			return arrVal, nil
		})
	case "slice":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			start := 0
			end := len(arr.ArrayData)
			if len(args) > 0 {
				start = int(args[0].ToNumber())
				if start < 0 {
					start = len(arr.ArrayData) + start
				}
			}
			if len(args) > 1 {
				end = int(args[1].ToNumber())
				if end < 0 {
					end = len(arr.ArrayData) + end
				}
			}
			if start < 0 {
				start = 0
			}
			if end > len(arr.ArrayData) {
				end = len(arr.ArrayData)
			}
			if start >= end {
				return runtime.NewObject(runtime.NewArrayObject(nil, nil)), nil
			}
			newData := make([]*runtime.Value, end-start)
			copy(newData, arr.ArrayData[start:end])
			return runtime.NewObject(runtime.NewArrayObject(nil, newData)), nil
		})
	case "concat":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			var result []*runtime.Value
			result = append(result, arr.ArrayData...)
			for _, arg := range args {
				if arg.Type == runtime.TypeObject && arg.Object != nil && arg.Object.OType == runtime.ObjTypeArray {
					result = append(result, arg.Object.ArrayData...)
				} else {
					result = append(result, arg)
				}
			}
			return runtime.NewObject(runtime.NewArrayObject(nil, result)), nil
		})
	case "map":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 || args[0].Type != runtime.TypeObject || args[0].Object.Callable == nil {
				return nil, fmt.Errorf("TypeError: callback is not a function")
			}
			cb := args[0].Object.Callable
			var result []*runtime.Value
			for i, v := range arr.ArrayData {
				r, err := cb(runtime.Undefined, []*runtime.Value{v, runtime.NewNumber(float64(i)), arrVal})
				if err != nil {
					return nil, err
				}
				result = append(result, r)
			}
			return runtime.NewObject(runtime.NewArrayObject(nil, result)), nil
		})
	case "filter":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 || args[0].Type != runtime.TypeObject || args[0].Object.Callable == nil {
				return nil, fmt.Errorf("TypeError: callback is not a function")
			}
			cb := args[0].Object.Callable
			var result []*runtime.Value
			for i, v := range arr.ArrayData {
				r, err := cb(runtime.Undefined, []*runtime.Value{v, runtime.NewNumber(float64(i)), arrVal})
				if err != nil {
					return nil, err
				}
				if r.ToBoolean() {
					result = append(result, v)
				}
			}
			return runtime.NewObject(runtime.NewArrayObject(nil, result)), nil
		})
	case "reduce":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 || args[0].Type != runtime.TypeObject || args[0].Object.Callable == nil {
				return nil, fmt.Errorf("TypeError: callback is not a function")
			}
			cb := args[0].Object.Callable
			var acc *runtime.Value
			startIdx := 0
			if len(args) > 1 {
				acc = args[1]
			} else {
				if len(arr.ArrayData) == 0 {
					return nil, fmt.Errorf("TypeError: Reduce of empty array with no initial value")
				}
				acc = arr.ArrayData[0]
				startIdx = 1
			}
			for i := startIdx; i < len(arr.ArrayData); i++ {
				var err error
				acc, err = cb(runtime.Undefined, []*runtime.Value{acc, arr.ArrayData[i], runtime.NewNumber(float64(i)), arrVal})
				if err != nil {
					return nil, err
				}
			}
			return acc, nil
		})
	case "forEach":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 || args[0].Type != runtime.TypeObject || args[0].Object.Callable == nil {
				return nil, fmt.Errorf("TypeError: callback is not a function")
			}
			cb := args[0].Object.Callable
			for i, v := range arr.ArrayData {
				_, err := cb(runtime.Undefined, []*runtime.Value{v, runtime.NewNumber(float64(i)), arrVal})
				if err != nil {
					return nil, err
				}
			}
			return runtime.Undefined, nil
		})
	case "find":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 || args[0].Type != runtime.TypeObject || args[0].Object.Callable == nil {
				return nil, fmt.Errorf("TypeError: callback is not a function")
			}
			cb := args[0].Object.Callable
			for i, v := range arr.ArrayData {
				r, err := cb(runtime.Undefined, []*runtime.Value{v, runtime.NewNumber(float64(i)), arrVal})
				if err != nil {
					return nil, err
				}
				if r.ToBoolean() {
					return v, nil
				}
			}
			return runtime.Undefined, nil
		})
	case "findIndex":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 || args[0].Type != runtime.TypeObject || args[0].Object.Callable == nil {
				return nil, fmt.Errorf("TypeError: callback is not a function")
			}
			cb := args[0].Object.Callable
			for i, v := range arr.ArrayData {
				r, err := cb(runtime.Undefined, []*runtime.Value{v, runtime.NewNumber(float64(i)), arrVal})
				if err != nil {
					return nil, err
				}
				if r.ToBoolean() {
					return runtime.NewNumber(float64(i)), nil
				}
			}
			return runtime.NewNumber(-1), nil
		})
	case "every":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 || args[0].Type != runtime.TypeObject || args[0].Object.Callable == nil {
				return nil, fmt.Errorf("TypeError: callback is not a function")
			}
			cb := args[0].Object.Callable
			for i, v := range arr.ArrayData {
				r, err := cb(runtime.Undefined, []*runtime.Value{v, runtime.NewNumber(float64(i)), arrVal})
				if err != nil {
					return nil, err
				}
				if !r.ToBoolean() {
					return runtime.False, nil
				}
			}
			return runtime.True, nil
		})
	case "some":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 || args[0].Type != runtime.TypeObject || args[0].Object.Callable == nil {
				return nil, fmt.Errorf("TypeError: callback is not a function")
			}
			cb := args[0].Object.Callable
			for i, v := range arr.ArrayData {
				r, err := cb(runtime.Undefined, []*runtime.Value{v, runtime.NewNumber(float64(i)), arrVal})
				if err != nil {
					return nil, err
				}
				if r.ToBoolean() {
					return runtime.True, nil
				}
			}
			return runtime.False, nil
		})
	case "flat":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			depth := 1
			if len(args) > 0 {
				depth = int(args[0].ToNumber())
			}
			result := flattenArray(arr.ArrayData, depth)
			return runtime.NewObject(runtime.NewArrayObject(nil, result)), nil
		})
	case "fill":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 {
				return arrVal, nil
			}
			fillVal := args[0]
			start := 0
			end := len(arr.ArrayData)
			if len(args) > 1 {
				start = int(args[1].ToNumber())
			}
			if len(args) > 2 {
				end = int(args[2].ToNumber())
			}
			if start < 0 {
				start = 0
			}
			if end > len(arr.ArrayData) {
				end = len(arr.ArrayData)
			}
			for i := start; i < end; i++ {
				arr.ArrayData[i] = fillVal
			}
			return arrVal, nil
		})
	case "splice":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) == 0 {
				return runtime.NewObject(runtime.NewArrayObject(nil, nil)), nil
			}
			start := int(args[0].ToNumber())
			if start < 0 {
				start = len(arr.ArrayData) + start
			}
			if start < 0 {
				start = 0
			}
			if start > len(arr.ArrayData) {
				start = len(arr.ArrayData)
			}
			deleteCount := len(arr.ArrayData) - start
			if len(args) > 1 {
				deleteCount = int(args[1].ToNumber())
			}
			if deleteCount < 0 {
				deleteCount = 0
			}
			if start+deleteCount > len(arr.ArrayData) {
				deleteCount = len(arr.ArrayData) - start
			}
			removed := make([]*runtime.Value, deleteCount)
			copy(removed, arr.ArrayData[start:start+deleteCount])
			var newItems []*runtime.Value
			if len(args) > 2 {
				newItems = args[2:]
			}
			newData := make([]*runtime.Value, 0, len(arr.ArrayData)-deleteCount+len(newItems))
			newData = append(newData, arr.ArrayData[:start]...)
			newData = append(newData, newItems...)
			newData = append(newData, arr.ArrayData[start+deleteCount:]...)
			arr.ArrayData = newData
			arr.Set("length", runtime.NewNumber(float64(len(arr.ArrayData))))
			return runtime.NewObject(runtime.NewArrayObject(nil, removed)), nil
		})
	case "sort":
		return interp.makeNativeMethod(func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			if len(args) > 0 && args[0].Type == runtime.TypeObject && args[0].Object != nil && args[0].Object.Callable != nil {
				cb := args[0].Object.Callable
				sort.SliceStable(arr.ArrayData, func(i, j int) bool {
					r, err := cb(runtime.Undefined, []*runtime.Value{arr.ArrayData[i], arr.ArrayData[j]})
					if err != nil {
						return false
					}
					return r.ToNumber() < 0
				})
			} else {
				sort.SliceStable(arr.ArrayData, func(i, j int) bool {
					return arr.ArrayData[i].ToString() < arr.ArrayData[j].ToString()
				})
			}
			return arrVal, nil
		})
	}
	return nil
}

func flattenArray(data []*runtime.Value, depth int) []*runtime.Value {
	var result []*runtime.Value
	for _, v := range data {
		if depth > 0 && v.Type == runtime.TypeObject && v.Object != nil && v.Object.OType == runtime.ObjTypeArray {
			result = append(result, flattenArray(v.Object.ArrayData, depth-1)...)
		} else {
			result = append(result, v)
		}
	}
	return result
}

func (interp *Interpreter) evalNew(e *ast.NewExpression, env *runtime.Environment) (*runtime.Value, signal) {
	callee, sig := interp.evalExpression(e.Callee, env)
	if sig.typ != sigNone {
		return nil, sig
	}

	if callee.Type != runtime.TypeObject || callee.Object == nil {
		return nil, signal{typ: sigThrow, value: makeErrorObject("TypeError", "is not a constructor", env)}
	}

	args, argSig := interp.evalArguments(e.Arguments, env)
	if argSig.typ != sigNone {
		return nil, argSig
	}

	// Get prototype
	var proto *runtime.Object
	protoProp := callee.Object.Get("prototype")
	if protoProp.Type == runtime.TypeObject && protoProp.Object != nil {
		proto = protoProp.Object
	}

	// Create new instance
	instance := runtime.NewOrdinaryObject(proto)
	thisVal := runtime.NewObject(instance)

	// Call constructor
	constructor := callee.Object.Callable
	if callee.Object.Constructor != nil {
		constructor = callee.Object.Constructor
	}
	if constructor == nil {
		return thisVal, signal{}
	}

	result, err := constructor(thisVal, args)
	if err != nil {
		if jsErr, ok := err.(*jsError); ok {
			return nil, signal{typ: sigThrow, value: jsErr.value}
		}
		return nil, signal{typ: sigThrow, value: errorFromGoError(err, env)}
	}

	// If constructor returns an object, use that; otherwise use this
	if result != nil && result.Type == runtime.TypeObject {
		return result, signal{}
	}
	return thisVal, signal{}
}

func (interp *Interpreter) evalSequence(e *ast.SequenceExpression, env *runtime.Environment) (*runtime.Value, signal) {
	var result *runtime.Value
	for _, expr := range e.Expressions {
		var sig signal
		result, sig = interp.evalExpression(expr, env)
		if sig.typ != sigNone {
			return nil, sig
		}
	}
	return result, signal{}
}

func (interp *Interpreter) evalTemplateLiteral(e *ast.TemplateLiteralExpr, env *runtime.Environment) (*runtime.Value, signal) {
	var sb strings.Builder
	for i, quasi := range e.Quasis {
		sb.WriteString(quasi.Value)
		if i < len(e.Expressions) {
			val, sig := interp.evalExpression(e.Expressions[i], env)
			if sig.typ != sigNone {
				return nil, sig
			}
			sb.WriteString(val.ToString())
		}
	}
	return runtime.NewString(sb.String()), signal{}
}

// evalCodeInEnv parses JS source code and evaluates it in the given environment.
// Used by both direct eval and indirect eval.
func (interp *Interpreter) evalCodeInEnv(code string, env *runtime.Environment) (*runtime.Value, signal) {
	p := parser.New(code)
	program, errs := p.ParseProgram()
	if len(errs) > 0 {
		errMsg := fmt.Sprintf("%v", errs[0])
		errObj := makeErrorObject("SyntaxError", errMsg, env)
		return nil, signal{typ: sigThrow, value: errObj}
	}

	interp.hoist(program.Statements, env)

	var result *runtime.Value
	for _, stmt := range program.Statements {
		val, sig := interp.execStatement(stmt, env)
		if sig.typ != sigNone {
			return nil, sig
		}
		if val != nil {
			result = val
		}
	}
	if result == nil {
		return runtime.Undefined, signal{}
	}
	return result, signal{}
}

// evalDirectEval handles direct eval(code) calls with access to the calling scope.
func (interp *Interpreter) evalDirectEval(args []*runtime.Value, env *runtime.Environment) (*runtime.Value, signal) {
	if len(args) == 0 {
		return runtime.Undefined, signal{}
	}
	if args[0].Type != runtime.TypeString {
		return args[0], signal{}
	}
	return interp.evalCodeInEnv(args[0].Str, env)
}

// makeFunctionConstructor creates the global Function constructor.
func (interp *Interpreter) makeFunctionConstructor(env *runtime.Environment) *runtime.Value {
	ctor := func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		var paramStr, bodyStr string
		if len(args) == 0 {
			bodyStr = ""
		} else if len(args) == 1 {
			bodyStr = args[0].ToString()
		} else {
			params := make([]string, len(args)-1)
			for i := 0; i < len(args)-1; i++ {
				params[i] = args[i].ToString()
			}
			paramStr = strings.Join(params, ",")
			bodyStr = args[len(args)-1].ToString()
		}

		source := "function anonymous(" + paramStr + ") { " + bodyStr + " }"
		p := parser.New(source)
		program, errs := p.ParseProgram()
		if len(errs) > 0 {
			errMsg := fmt.Sprintf("%v", errs[0])
			errObj := makeErrorObject("SyntaxError", errMsg, env)
			return nil, &jsError{value: errObj}
		}

		if len(program.Statements) == 0 {
			return runtime.Undefined, nil
		}

		funcDecl, ok := program.Statements[0].(*ast.FunctionDeclaration)
		if !ok {
			return runtime.Undefined, nil
		}

		// Function constructor creates functions that execute in the global scope
		fnVal := interp.createFunction(funcDecl.Name, funcDecl.Params, funcDecl.Defaults, funcDecl.Rest, funcDecl.Body, env, false)
		return fnVal, nil
	}

	fnObj := runtime.NewFunctionObject(nil, ctor)
	fnObj.Constructor = ctor
	fnObj.Set("prototype", runtime.NewObject(runtime.NewOrdinaryObject(runtime.DefaultObjectPrototype)))
	fnObj.Set("length", runtime.NewNumber(1))
	fnObj.Set("name", runtime.NewString("Function"))
	return runtime.NewObject(fnObj)
}
