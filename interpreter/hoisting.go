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
	// Per spec, skip names that would conflict with lexical (let/const) declarations.
	if funcScope == env {
		lexicalNames := interp.collectTopLevelLexicalNames(stmts)
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
// Names in lexicalNames are skipped (they conflict with let/const in the function scope).
func (interp *Interpreter) collectBlockFuncDecls(stmts []ast.Statement, env *runtime.Environment, lexicalNames map[string]bool) {
	for _, stmt := range stmts {
		interp.collectBlockFuncDeclsFromStmt(stmt, env, lexicalNames)
	}
}

func (interp *Interpreter) collectBlockFuncDeclsFromStmt(stmt ast.Statement, env *runtime.Environment, lexicalNames map[string]bool) {
	switch s := stmt.(type) {
	case *ast.BlockStatement:
		for _, inner := range s.Statements {
			if fd, ok := inner.(*ast.FunctionDeclaration); ok {
				if !lexicalNames[fd.Name.Value] {
					env.DeclareVar(fd.Name.Value)
				}
			}
			interp.collectBlockFuncDeclsFromStmt(inner, env, lexicalNames)
		}
	case *ast.IfStatement:
		if s.Consequence != nil {
			for _, inner := range s.Consequence.Statements {
				if fd, ok := inner.(*ast.FunctionDeclaration); ok {
					if !lexicalNames[fd.Name.Value] {
						env.DeclareVar(fd.Name.Value)
					}
				}
				interp.collectBlockFuncDeclsFromStmt(inner, env, lexicalNames)
			}
		}
		if s.Alternative != nil {
			interp.collectBlockFuncDeclsFromStmt(s.Alternative, env, lexicalNames)
		}
	case *ast.SwitchStatement:
		for _, c := range s.Cases {
			for _, inner := range c.Consequent {
				if fd, ok := inner.(*ast.FunctionDeclaration); ok {
					if !lexicalNames[fd.Name.Value] {
						env.DeclareVar(fd.Name.Value)
					}
				}
				interp.collectBlockFuncDeclsFromStmt(inner, env, lexicalNames)
			}
		}
	case *ast.TryStatement:
		if s.Block != nil {
			for _, inner := range s.Block.Statements {
				if fd, ok := inner.(*ast.FunctionDeclaration); ok {
					if !lexicalNames[fd.Name.Value] {
						env.DeclareVar(fd.Name.Value)
					}
				}
				interp.collectBlockFuncDeclsFromStmt(inner, env, lexicalNames)
			}
		}
		if s.Handler != nil && s.Handler.Body != nil {
			for _, inner := range s.Handler.Body.Statements {
				if fd, ok := inner.(*ast.FunctionDeclaration); ok {
					if !lexicalNames[fd.Name.Value] {
						env.DeclareVar(fd.Name.Value)
					}
				}
				interp.collectBlockFuncDeclsFromStmt(inner, env, lexicalNames)
			}
		}
		if s.Finalizer != nil {
			for _, inner := range s.Finalizer.Statements {
				if fd, ok := inner.(*ast.FunctionDeclaration); ok {
					if !lexicalNames[fd.Name.Value] {
						env.DeclareVar(fd.Name.Value)
					}
				}
				interp.collectBlockFuncDeclsFromStmt(inner, env, lexicalNames)
			}
		}
	case *ast.ForStatement:
		if s.Body != nil {
			interp.collectBlockFuncDeclsFromStmt(s.Body, env, lexicalNames)
		}
	case *ast.ForInStatement:
		if s.Body != nil {
			interp.collectBlockFuncDeclsFromStmt(s.Body, env, lexicalNames)
		}
	case *ast.ForOfStatement:
		if s.Body != nil {
			interp.collectBlockFuncDeclsFromStmt(s.Body, env, lexicalNames)
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
