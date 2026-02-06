package interpreter

import (
	"github.com/example/jsgo/ast"
	"github.com/example/jsgo/runtime"
)

// hoistComprehensive performs comprehensive var and function hoisting.
// It replaces the simple hoist() by:
// 1. Recursively walking ALL nested structures for var declarations
// 2. Hoisting function declarations at the current level with their values
// 3. Annex B: hoisting function declarations inside blocks to function scope
func (interp *Interpreter) hoistComprehensive(stmts []ast.Statement, env *runtime.Environment) {
	funcScope := env.GetFunctionScope()

	// First pass: recursively hoist all var declarations to function scope.
	interp.collectVarDecls(stmts, funcScope)

	// Second pass: hoist function declarations at this level with their values.
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.FunctionDeclaration:
			fnVal := interp.createFunction(s.Name, s.Params, s.Defaults, s.Rest, s.Body, env, false)
			env.Declare(s.Name.Value, "function", fnVal)
		case *ast.LabeledStatement:
			if fd, ok := s.Body.(*ast.FunctionDeclaration); ok {
				fnVal := interp.createFunction(fd.Name, fd.Params, fd.Defaults, fd.Rest, fd.Body, env, false)
				env.Declare(fd.Name.Value, "function", fnVal)
			}
		}
	}

	// Annex B: if this is a function/program scope (not a block scope),
	// also hoist function declarations found inside blocks to the function scope.
	// Per spec, skip names that would conflict with lexical (let/const) declarations
	// or parameter names (including "arguments").
	if funcScope == env {
		lexicalNames := interp.collectTopLevelLexicalNames(stmts)
		// Per spec B.3.3.1: skip names that are in parameterNames.
		// Parameter names are already bound as "let" in the function scope,
		// and "arguments" is bound as "var". Collect all existing non-var
		// bindings plus "arguments" (which is a special parameter name).
		env.ForEachBinding(func(name string, kind string) {
			if kind == "let" || kind == "const" {
				lexicalNames[name] = true
			}
		})
		// "arguments" is in parameterNames per spec step 22f, skip it
		if env.HasBinding("arguments") {
			lexicalNames["arguments"] = true
		}
		interp.collectBlockFuncDecls(stmts, env, lexicalNames)
	}
}

// collectTopLevelLexicalNames collects all let/const declared names at the top level
// of a statement list. Used to prevent Annex B block-function hoisting from conflicting
// with lexical declarations.
func (interp *Interpreter) collectTopLevelLexicalNames(stmts []ast.Statement) map[string]bool {
	names := make(map[string]bool)
	for _, stmt := range stmts {
		if vd, ok := stmt.(*ast.VariableDeclaration); ok {
			if vd.Kind == "let" || vd.Kind == "const" {
				for _, decl := range vd.Declarations {
					for _, name := range interp.extractBindingNames(decl.Name) {
						names[name] = true
					}
				}
			}
		}
	}
	return names
}

// collectVarDecls recursively walks all nested structures to find var declarations
// and hoists them to the given function scope with undefined.
func (interp *Interpreter) collectVarDecls(stmts []ast.Statement, funcScope *runtime.Environment) {
	for _, stmt := range stmts {
		interp.collectVarDeclsFromStmt(stmt, funcScope)
	}
}

func (interp *Interpreter) collectVarDeclsFromStmt(stmt ast.Statement, funcScope *runtime.Environment) {
	switch s := stmt.(type) {
	case *ast.VariableDeclaration:
		if s.Kind == "var" {
			for _, decl := range s.Declarations {
				names := interp.extractBindingNames(decl.Name)
				for _, name := range names {
					funcScope.SetInCurrentScope(name, runtime.Undefined)
				}
			}
		}
	case *ast.BlockStatement:
		interp.collectVarDecls(s.Statements, funcScope)
	case *ast.IfStatement:
		if s.Consequence != nil {
			interp.collectVarDecls(s.Consequence.Statements, funcScope)
		}
		if s.Alternative != nil {
			interp.collectVarDeclsFromStmt(s.Alternative, funcScope)
		}
	case *ast.ForStatement:
		if s.Init != nil {
			if initStmt, ok := s.Init.(ast.Statement); ok {
				interp.collectVarDeclsFromStmt(initStmt, funcScope)
			}
		}
		if s.Body != nil {
			interp.collectVarDeclsFromStmt(s.Body, funcScope)
		}
	case *ast.ForInStatement:
		if left, ok := s.Left.(*ast.VariableDeclaration); ok && left.Kind == "var" {
			for _, decl := range left.Declarations {
				names := interp.extractBindingNames(decl.Name)
				for _, name := range names {
					funcScope.SetInCurrentScope(name, runtime.Undefined)
				}
			}
		}
		if s.Body != nil {
			interp.collectVarDeclsFromStmt(s.Body, funcScope)
		}
	case *ast.ForOfStatement:
		if left, ok := s.Left.(*ast.VariableDeclaration); ok && left.Kind == "var" {
			for _, decl := range left.Declarations {
				names := interp.extractBindingNames(decl.Name)
				for _, name := range names {
					funcScope.SetInCurrentScope(name, runtime.Undefined)
				}
			}
		}
		if s.Body != nil {
			interp.collectVarDeclsFromStmt(s.Body, funcScope)
		}
	case *ast.WhileStatement:
		if s.Body != nil {
			interp.collectVarDeclsFromStmt(s.Body, funcScope)
		}
	case *ast.DoWhileStatement:
		if s.Body != nil {
			interp.collectVarDeclsFromStmt(s.Body, funcScope)
		}
	case *ast.SwitchStatement:
		for _, c := range s.Cases {
			interp.collectVarDecls(c.Consequent, funcScope)
		}
	case *ast.TryStatement:
		if s.Block != nil {
			interp.collectVarDecls(s.Block.Statements, funcScope)
		}
		if s.Handler != nil && s.Handler.Body != nil {
			interp.collectVarDecls(s.Handler.Body.Statements, funcScope)
		}
		if s.Finalizer != nil {
			interp.collectVarDecls(s.Finalizer.Statements, funcScope)
		}
	case *ast.LabeledStatement:
		if s.Body != nil {
			interp.collectVarDeclsFromStmt(s.Body, funcScope)
		}
	}
}

// collectBlockFuncDecls walks into blocks/if/switch/try to find function declarations
// and hoists their NAMES to the function scope as var (initialized to undefined).
// Per Annex B semantics, the actual function value is NOT assigned here;
// it gets assigned when the declaration is reached during execution.
// Names in lexicalNames are skipped (they conflict with let/const at this or any
// enclosing block level between the function scope and the declaration).
func (interp *Interpreter) collectBlockFuncDecls(stmts []ast.Statement, env *runtime.Environment, lexicalNames map[string]bool) {
	for _, stmt := range stmts {
		interp.collectBlockFuncDeclsFromStmt(stmt, env, lexicalNames)
	}
}

// collectLexicalNamesFromStmts collects let/const/function declared names from a
// list of statements. In block scope, function declarations are block-scoped (like let),
// so they block Annex B hoisting of the same name from deeper blocks.
func (interp *Interpreter) collectLexicalNamesFromStmts(stmts []ast.Statement) map[string]bool {
	names := make(map[string]bool)
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.VariableDeclaration:
			if s.Kind == "let" || s.Kind == "const" {
				for _, decl := range s.Declarations {
					for _, name := range interp.extractBindingNames(decl.Name) {
						names[name] = true
					}
				}
			}
		case *ast.FunctionDeclaration:
			names[s.Name.Value] = true
		}
	}
	return names
}

// mergeLexicalNames returns a new map containing all names from both maps.
func mergeLexicalNames(a, b map[string]bool) map[string]bool {
	if len(b) == 0 {
		return a
	}
	merged := make(map[string]bool, len(a)+len(b))
	for k := range a {
		merged[k] = true
	}
	for k := range b {
		merged[k] = true
	}
	return merged
}

// collectBlockFuncDeclsInBlock processes a block's statements: hoists direct function
// declarations using only the parent lexical names (so they don't block themselves),
// then recurses deeper with merged lexical names (including this block's let/const/function
// names) so deeper function declarations are blocked appropriately.
func (interp *Interpreter) collectBlockFuncDeclsInBlock(stmts []ast.Statement, env *runtime.Environment, lexicalNames map[string]bool) {
	// Direct function declarations at this level are checked against parent lexicalNames only
	for _, inner := range stmts {
		if fd, ok := inner.(*ast.FunctionDeclaration); ok {
			if !lexicalNames[fd.Name.Value] {
				env.DeclareVar(fd.Name.Value)
			}
		}
	}

	// For deeper recursion, merge this block's lexical names (let/const/function)
	blockLexNames := interp.collectLexicalNamesFromStmts(stmts)
	merged := mergeLexicalNames(lexicalNames, blockLexNames)

	for _, inner := range stmts {
		interp.collectBlockFuncDeclsFromStmt(inner, env, merged)
	}
}

func (interp *Interpreter) collectBlockFuncDeclsFromStmt(stmt ast.Statement, env *runtime.Environment, lexicalNames map[string]bool) {
	switch s := stmt.(type) {
	case *ast.BlockStatement:
		interp.collectBlockFuncDeclsInBlock(s.Statements, env, lexicalNames)
	case *ast.IfStatement:
		if s.Consequence != nil {
			interp.collectBlockFuncDeclsInBlock(s.Consequence.Statements, env, lexicalNames)
		}
		if s.Alternative != nil {
			interp.collectBlockFuncDeclsFromStmt(s.Alternative, env, lexicalNames)
		}
	case *ast.SwitchStatement:
		// Switch cases share a single block scope; collect lexical names across all cases.
		var allStmts []ast.Statement
		for _, c := range s.Cases {
			allStmts = append(allStmts, c.Consequent...)
		}
		interp.collectBlockFuncDeclsInBlock(allStmts, env, lexicalNames)
	case *ast.TryStatement:
		if s.Block != nil {
			interp.collectBlockFuncDeclsInBlock(s.Block.Statements, env, lexicalNames)
		}
		if s.Handler != nil && s.Handler.Body != nil {
			// Per B.3.5: catch parameter blocks Annex B hoisting ONLY when it is
			// a destructuring pattern (not a simple BindingIdentifier).
			// A simple catch(f) allows var f inside the block, so Annex B hoisting
			// of function f is also allowed. But catch({f}) or catch([f]) blocks it.
			catchLex := make(map[string]bool)
			if s.Handler.Param != nil {
				if _, isIdent := s.Handler.Param.(*ast.Identifier); !isIdent {
					for _, name := range interp.extractBindingNames(s.Handler.Param) {
						catchLex[name] = true
					}
				}
			}
			merged := mergeLexicalNames(lexicalNames, catchLex)
			interp.collectBlockFuncDeclsInBlock(s.Handler.Body.Statements, env, merged)
		}
		if s.Finalizer != nil {
			interp.collectBlockFuncDeclsInBlock(s.Finalizer.Statements, env, lexicalNames)
		}
	case *ast.ForStatement:
		// for (let f; ...) introduces lexical names that block hoisting in body
		forLex := make(map[string]bool)
		if s.Init != nil {
			if vd, ok := s.Init.(*ast.VariableDeclaration); ok {
				if vd.Kind == "let" || vd.Kind == "const" {
					for _, decl := range vd.Declarations {
						for _, name := range interp.extractBindingNames(decl.Name) {
							forLex[name] = true
						}
					}
				}
			}
		}
		merged := mergeLexicalNames(lexicalNames, forLex)
		if s.Body != nil {
			interp.collectBlockFuncDeclsFromStmt(s.Body, env, merged)
		}
	case *ast.ForInStatement:
		forLex := make(map[string]bool)
		if vd, ok := s.Left.(*ast.VariableDeclaration); ok {
			if vd.Kind == "let" || vd.Kind == "const" {
				for _, decl := range vd.Declarations {
					for _, name := range interp.extractBindingNames(decl.Name) {
						forLex[name] = true
					}
				}
			}
		}
		merged := mergeLexicalNames(lexicalNames, forLex)
		if s.Body != nil {
			interp.collectBlockFuncDeclsFromStmt(s.Body, env, merged)
		}
	case *ast.ForOfStatement:
		forLex := make(map[string]bool)
		if vd, ok := s.Left.(*ast.VariableDeclaration); ok {
			if vd.Kind == "let" || vd.Kind == "const" {
				for _, decl := range vd.Declarations {
					for _, name := range interp.extractBindingNames(decl.Name) {
						forLex[name] = true
					}
				}
			}
		}
		merged := mergeLexicalNames(lexicalNames, forLex)
		if s.Body != nil {
			interp.collectBlockFuncDeclsFromStmt(s.Body, env, merged)
		}
	case *ast.WhileStatement:
		if s.Body != nil {
			interp.collectBlockFuncDeclsFromStmt(s.Body, env, lexicalNames)
		}
	case *ast.DoWhileStatement:
		if s.Body != nil {
			interp.collectBlockFuncDeclsFromStmt(s.Body, env, lexicalNames)
		}
	case *ast.LabeledStatement:
		interp.collectBlockFuncDeclsFromStmt(s.Body, env, lexicalNames)
	}
}
