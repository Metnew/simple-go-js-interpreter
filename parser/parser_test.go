package parser

import (
	"testing"

	"github.com/example/jsgo/ast"
)

func parse(t *testing.T, input string) *ast.Program {
	t.Helper()
	p := New(input)
	prog, errs := p.ParseProgram()
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parser error: %s", e)
		}
		t.FailNow()
	}
	return prog
}

func parseWithErrors(input string) (*ast.Program, []error) {
	p := New(input)
	return p.ParseProgram()
}

func expectStmtCount(t *testing.T, prog *ast.Program, n int) {
	t.Helper()
	if len(prog.Statements) != n {
		t.Fatalf("expected %d statements, got %d", n, len(prog.Statements))
	}
}

// ---------- Variable Declarations ----------

func TestVarDeclaration(t *testing.T) {
	prog := parse(t, `var x = 1;`)
	expectStmtCount(t, prog, 1)
	decl, ok := prog.Statements[0].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("expected VariableDeclaration, got %T", prog.Statements[0])
	}
	if decl.Kind != "var" {
		t.Errorf("expected kind var, got %s", decl.Kind)
	}
	if len(decl.Declarations) != 1 {
		t.Fatalf("expected 1 declarator, got %d", len(decl.Declarations))
	}
	ident, ok := decl.Declarations[0].Name.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier, got %T", decl.Declarations[0].Name)
	}
	if ident.Value != "x" {
		t.Errorf("expected x, got %s", ident.Value)
	}
}

func TestLetConstDeclaration(t *testing.T) {
	prog := parse(t, `let a = 1; const b = 2;`)
	expectStmtCount(t, prog, 2)

	decl1 := prog.Statements[0].(*ast.VariableDeclaration)
	if decl1.Kind != "let" {
		t.Errorf("expected let, got %s", decl1.Kind)
	}

	decl2 := prog.Statements[1].(*ast.VariableDeclaration)
	if decl2.Kind != "const" {
		t.Errorf("expected const, got %s", decl2.Kind)
	}
}

func TestMultipleDeclarators(t *testing.T) {
	prog := parse(t, `var a = 1, b = 2, c;`)
	expectStmtCount(t, prog, 1)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	if len(decl.Declarations) != 3 {
		t.Fatalf("expected 3 declarators, got %d", len(decl.Declarations))
	}
	if decl.Declarations[2].Value != nil {
		t.Error("expected nil value for c")
	}
}

func TestDestructuringObject(t *testing.T) {
	prog := parse(t, `const { a, b: c } = obj;`)
	expectStmtCount(t, prog, 1)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	pat, ok := decl.Declarations[0].Name.(*ast.ObjectPattern)
	if !ok {
		t.Fatalf("expected ObjectPattern, got %T", decl.Declarations[0].Name)
	}
	if len(pat.Properties) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(pat.Properties))
	}
}

func TestDestructuringArray(t *testing.T) {
	prog := parse(t, `const [a, , b] = arr;`)
	expectStmtCount(t, prog, 1)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	pat, ok := decl.Declarations[0].Name.(*ast.ArrayPattern)
	if !ok {
		t.Fatalf("expected ArrayPattern, got %T", decl.Declarations[0].Name)
	}
	if len(pat.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(pat.Elements))
	}
	if pat.Elements[1] != nil {
		t.Error("expected nil for elision")
	}
}

func TestDestructuringWithDefaults(t *testing.T) {
	prog := parse(t, `const { a = 1, b = 2 } = obj;`)
	expectStmtCount(t, prog, 1)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	pat := decl.Declarations[0].Name.(*ast.ObjectPattern)
	if len(pat.Properties) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(pat.Properties))
	}
	// First property should have AssignmentPattern as value
	_, ok := pat.Properties[0].Value.(*ast.AssignmentPattern)
	if !ok {
		t.Errorf("expected AssignmentPattern for default, got %T", pat.Properties[0].Value)
	}
}

func TestDestructuringRest(t *testing.T) {
	prog := parse(t, `const [a, ...rest] = arr;`)
	expectStmtCount(t, prog, 1)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	pat := decl.Declarations[0].Name.(*ast.ArrayPattern)
	if len(pat.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(pat.Elements))
	}
	_, ok := pat.Elements[1].(*ast.RestElement)
	if !ok {
		t.Errorf("expected RestElement, got %T", pat.Elements[1])
	}
}

// ---------- Expression Statements ----------

func TestNumberLiteral(t *testing.T) {
	prog := parse(t, `42;`)
	expectStmtCount(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	num, ok := stmt.Expression.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral, got %T", stmt.Expression)
	}
	if num.Value != 42 {
		t.Errorf("expected 42, got %f", num.Value)
	}
}

func TestStringLiteral(t *testing.T) {
	prog := parse(t, `"hello";`)
	expectStmtCount(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	str, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expected StringLiteral, got %T", stmt.Expression)
	}
	if str.Value != "hello" {
		t.Errorf("expected hello, got %s", str.Value)
	}
}

func TestBooleanLiteral(t *testing.T) {
	prog := parse(t, `true; false;`)
	expectStmtCount(t, prog, 2)
	stmt1 := prog.Statements[0].(*ast.ExpressionStatement)
	b1 := stmt1.Expression.(*ast.BooleanLiteral)
	if !b1.Value {
		t.Error("expected true")
	}
	stmt2 := prog.Statements[1].(*ast.ExpressionStatement)
	b2 := stmt2.Expression.(*ast.BooleanLiteral)
	if b2.Value {
		t.Error("expected false")
	}
}

func TestNullLiteral(t *testing.T) {
	prog := parse(t, `null;`)
	expectStmtCount(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	_, ok := stmt.Expression.(*ast.NullLiteral)
	if !ok {
		t.Fatalf("expected NullLiteral, got %T", stmt.Expression)
	}
}

func TestUndefinedLiteral(t *testing.T) {
	prog := parse(t, `undefined;`)
	expectStmtCount(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	_, ok := stmt.Expression.(*ast.UndefinedLiteral)
	if !ok {
		t.Fatalf("expected UndefinedLiteral, got %T", stmt.Expression)
	}
}

// ---------- Binary Expressions ----------

func TestBinaryExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{"1 + 2;", "+"},
		{"1 - 2;", "-"},
		{"1 * 2;", "*"},
		{"1 / 2;", "/"},
		{"1 % 2;", "%"},
		{"1 ** 2;", "**"},
		{"1 == 2;", "=="},
		{"1 != 2;", "!="},
		{"1 === 2;", "==="},
		{"1 !== 2;", "!=="},
		{"1 < 2;", "<"},
		{"1 > 2;", ">"},
		{"1 <= 2;", "<="},
		{"1 >= 2;", ">="},
		{"1 & 2;", "&"},
		{"1 | 2;", "|"},
		{"1 ^ 2;", "^"},
		{"1 << 2;", "<<"},
		{"1 >> 2;", ">>"},
		{"1 >>> 2;", ">>>"},
		{"a instanceof b;", "instanceof"},
	}

	for _, tt := range tests {
		prog := parse(t, tt.input)
		stmt := prog.Statements[0].(*ast.ExpressionStatement)
		bin, ok := stmt.Expression.(*ast.BinaryExpression)
		if !ok {
			t.Errorf("for %q: expected BinaryExpression, got %T", tt.input, stmt.Expression)
			continue
		}
		if bin.Operator != tt.operator {
			t.Errorf("for %q: expected operator %s, got %s", tt.input, tt.operator, bin.Operator)
		}
	}
}

func TestLogicalExpressions(t *testing.T) {
	prog := parse(t, `a && b || c;`)
	expectStmtCount(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	// Should be (a && b) || c due to precedence
	log, ok := stmt.Expression.(*ast.LogicalExpression)
	if !ok {
		t.Fatalf("expected LogicalExpression, got %T", stmt.Expression)
	}
	if log.Operator != "||" {
		t.Errorf("expected ||, got %s", log.Operator)
	}
	left, ok := log.Left.(*ast.LogicalExpression)
	if !ok {
		t.Fatalf("expected LogicalExpression on left, got %T", log.Left)
	}
	if left.Operator != "&&" {
		t.Errorf("expected &&, got %s", left.Operator)
	}
}

func TestOperatorPrecedence(t *testing.T) {
	// 1 + 2 * 3 should be 1 + (2 * 3)
	prog := parse(t, `1 + 2 * 3;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	add, ok := stmt.Expression.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected BinaryExpression, got %T", stmt.Expression)
	}
	if add.Operator != "+" {
		t.Errorf("expected +, got %s", add.Operator)
	}
	mul, ok := add.Right.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected BinaryExpression on right, got %T", add.Right)
	}
	if mul.Operator != "*" {
		t.Errorf("expected *, got %s", mul.Operator)
	}
}

func TestExponentRightAssociativity(t *testing.T) {
	// 2 ** 3 ** 2 should be 2 ** (3 ** 2)
	prog := parse(t, `2 ** 3 ** 2;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	exp1 := stmt.Expression.(*ast.BinaryExpression)
	if exp1.Operator != "**" {
		t.Errorf("expected **, got %s", exp1.Operator)
	}
	exp2, ok := exp1.Right.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected BinaryExpression on right, got %T", exp1.Right)
	}
	if exp2.Operator != "**" {
		t.Errorf("expected **, got %s", exp2.Operator)
	}
}

// ---------- Unary Expressions ----------

func TestUnaryExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{"!a;", "!"},
		{"~a;", "~"},
		{"-a;", "-"},
		{"+a;", "+"},
		{"typeof a;", "typeof"},
		{"void 0;", "void"},
		{"delete obj.x;", "delete"},
	}

	for _, tt := range tests {
		prog := parse(t, tt.input)
		stmt := prog.Statements[0].(*ast.ExpressionStatement)
		unary, ok := stmt.Expression.(*ast.UnaryExpression)
		if !ok {
			t.Errorf("for %q: expected UnaryExpression, got %T", tt.input, stmt.Expression)
			continue
		}
		if unary.Operator != tt.operator {
			t.Errorf("for %q: expected operator %s, got %s", tt.input, tt.operator, unary.Operator)
		}
		if !unary.Prefix {
			t.Errorf("for %q: expected prefix", tt.input)
		}
	}
}

// ---------- Update Expressions ----------

func TestPrefixUpdate(t *testing.T) {
	prog := parse(t, `++x;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	upd, ok := stmt.Expression.(*ast.UpdateExpression)
	if !ok {
		t.Fatalf("expected UpdateExpression, got %T", stmt.Expression)
	}
	if upd.Operator != "++" {
		t.Errorf("expected ++, got %s", upd.Operator)
	}
	if !upd.Prefix {
		t.Error("expected prefix")
	}
}

func TestPostfixUpdate(t *testing.T) {
	prog := parse(t, `x++;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	upd, ok := stmt.Expression.(*ast.UpdateExpression)
	if !ok {
		t.Fatalf("expected UpdateExpression, got %T", stmt.Expression)
	}
	if upd.Operator != "++" {
		t.Errorf("expected ++, got %s", upd.Operator)
	}
	if upd.Prefix {
		t.Error("expected postfix")
	}
}

// ---------- Assignment ----------

func TestAssignmentExpression(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{"x = 1;", "="},
		{"x += 1;", "+="},
		{"x -= 1;", "-="},
		{"x *= 1;", "*="},
		{"x /= 1;", "/="},
		{"x %= 1;", "%="},
		{"x **= 1;", "**="},
		{"x &= 1;", "&="},
		{"x |= 1;", "|="},
		{"x ^= 1;", "^="},
		{"x <<= 1;", "<<="},
		{"x >>= 1;", ">>="},
		{"x >>>= 1;", ">>>="},
	}

	for _, tt := range tests {
		prog := parse(t, tt.input)
		stmt := prog.Statements[0].(*ast.ExpressionStatement)
		assign, ok := stmt.Expression.(*ast.AssignmentExpression)
		if !ok {
			t.Errorf("for %q: expected AssignmentExpression, got %T", tt.input, stmt.Expression)
			continue
		}
		if assign.Operator != tt.operator {
			t.Errorf("for %q: expected operator %s, got %s", tt.input, tt.operator, assign.Operator)
		}
	}
}

// ---------- Conditional (Ternary) ----------

func TestConditionalExpression(t *testing.T) {
	prog := parse(t, `a ? b : c;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	cond, ok := stmt.Expression.(*ast.ConditionalExpression)
	if !ok {
		t.Fatalf("expected ConditionalExpression, got %T", stmt.Expression)
	}
	test := cond.Test.(*ast.Identifier)
	if test.Value != "a" {
		t.Errorf("expected a, got %s", test.Value)
	}
}

// ---------- Call Expressions ----------

func TestCallExpression(t *testing.T) {
	prog := parse(t, `foo(1, 2, 3);`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", stmt.Expression)
	}
	callee := call.Callee.(*ast.Identifier)
	if callee.Value != "foo" {
		t.Errorf("expected foo, got %s", callee.Value)
	}
	if len(call.Arguments) != 3 {
		t.Errorf("expected 3 args, got %d", len(call.Arguments))
	}
}

func TestCallExpressionWithSpread(t *testing.T) {
	prog := parse(t, `foo(...args);`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	call := stmt.Expression.(*ast.CallExpression)
	if len(call.Arguments) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Arguments))
	}
	_, ok := call.Arguments[0].(*ast.SpreadElement)
	if !ok {
		t.Errorf("expected SpreadElement, got %T", call.Arguments[0])
	}
}

// ---------- Member Expressions ----------

func TestDotMemberExpression(t *testing.T) {
	prog := parse(t, `obj.prop;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	mem, ok := stmt.Expression.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("expected MemberExpression, got %T", stmt.Expression)
	}
	if mem.Computed {
		t.Error("expected non-computed")
	}
}

func TestBracketMemberExpression(t *testing.T) {
	prog := parse(t, `obj["prop"];`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	mem, ok := stmt.Expression.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("expected MemberExpression, got %T", stmt.Expression)
	}
	if !mem.Computed {
		t.Error("expected computed")
	}
}

func TestChainedMemberExpressions(t *testing.T) {
	prog := parse(t, `a.b.c;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	mem, ok := stmt.Expression.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("expected MemberExpression, got %T", stmt.Expression)
	}
	inner, ok := mem.Object.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("expected MemberExpression, got %T", mem.Object)
	}
	obj := inner.Object.(*ast.Identifier)
	if obj.Value != "a" {
		t.Errorf("expected a, got %s", obj.Value)
	}
}

// ---------- New Expressions ----------

func TestNewExpression(t *testing.T) {
	prog := parse(t, `new Foo(1, 2);`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	ne, ok := stmt.Expression.(*ast.NewExpression)
	if !ok {
		t.Fatalf("expected NewExpression, got %T", stmt.Expression)
	}
	callee := ne.Callee.(*ast.Identifier)
	if callee.Value != "Foo" {
		t.Errorf("expected Foo, got %s", callee.Value)
	}
	if len(ne.Arguments) != 2 {
		t.Errorf("expected 2 args, got %d", len(ne.Arguments))
	}
}

func TestNewWithoutArgs(t *testing.T) {
	prog := parse(t, `new Foo;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	ne, ok := stmt.Expression.(*ast.NewExpression)
	if !ok {
		t.Fatalf("expected NewExpression, got %T", stmt.Expression)
	}
	if len(ne.Arguments) != 0 {
		t.Errorf("expected 0 args, got %d", len(ne.Arguments))
	}
}

// ---------- If Statement ----------

func TestIfStatement(t *testing.T) {
	prog := parse(t, `if (x) { y; }`)
	expectStmtCount(t, prog, 1)
	stmt, ok := prog.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", prog.Statements[0])
	}
	if stmt.Alternative != nil {
		t.Error("expected no alternative")
	}
}

func TestIfElseStatement(t *testing.T) {
	prog := parse(t, `if (x) { y; } else { z; }`)
	stmt := prog.Statements[0].(*ast.IfStatement)
	if stmt.Alternative == nil {
		t.Error("expected alternative")
	}
}

func TestIfElseIfStatement(t *testing.T) {
	prog := parse(t, `if (a) { b; } else if (c) { d; } else { e; }`)
	stmt := prog.Statements[0].(*ast.IfStatement)
	alt, ok := stmt.Alternative.(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement as alternative, got %T", stmt.Alternative)
	}
	if alt.Alternative == nil {
		t.Error("expected final else")
	}
}

// ---------- While ----------

func TestWhileStatement(t *testing.T) {
	prog := parse(t, `while (true) { x; }`)
	expectStmtCount(t, prog, 1)
	_, ok := prog.Statements[0].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("expected WhileStatement, got %T", prog.Statements[0])
	}
}

// ---------- Do-While ----------

func TestDoWhileStatement(t *testing.T) {
	prog := parse(t, `do { x; } while (y);`)
	expectStmtCount(t, prog, 1)
	stmt, ok := prog.Statements[0].(*ast.DoWhileStatement)
	if !ok {
		t.Fatalf("expected DoWhileStatement, got %T", prog.Statements[0])
	}
	_ = stmt
}

// ---------- For ----------

func TestForStatement(t *testing.T) {
	prog := parse(t, `for (var i = 0; i < 10; i++) { x; }`)
	expectStmtCount(t, prog, 1)
	stmt, ok := prog.Statements[0].(*ast.ForStatement)
	if !ok {
		t.Fatalf("expected ForStatement, got %T", prog.Statements[0])
	}
	if stmt.Init == nil {
		t.Error("expected init")
	}
	if stmt.Test == nil {
		t.Error("expected test")
	}
	if stmt.Update == nil {
		t.Error("expected update")
	}
}

func TestForInStatement(t *testing.T) {
	prog := parse(t, `for (var k in obj) { x; }`)
	expectStmtCount(t, prog, 1)
	_, ok := prog.Statements[0].(*ast.ForInStatement)
	if !ok {
		t.Fatalf("expected ForInStatement, got %T", prog.Statements[0])
	}
}

func TestForOfStatement(t *testing.T) {
	prog := parse(t, `for (const item of arr) { x; }`)
	expectStmtCount(t, prog, 1)
	_, ok := prog.Statements[0].(*ast.ForOfStatement)
	if !ok {
		t.Fatalf("expected ForOfStatement, got %T", prog.Statements[0])
	}
}

func TestForOfWithDestructuring(t *testing.T) {
	prog := parse(t, `for (const [a, b] of arr) { x; }`)
	expectStmtCount(t, prog, 1)
	stmt, ok := prog.Statements[0].(*ast.ForOfStatement)
	if !ok {
		t.Fatalf("expected ForOfStatement, got %T", prog.Statements[0])
	}
	decl := stmt.Left.(*ast.VariableDeclaration)
	_, ok = decl.Declarations[0].Name.(*ast.ArrayPattern)
	if !ok {
		t.Errorf("expected ArrayPattern, got %T", decl.Declarations[0].Name)
	}
}

// ---------- Break / Continue ----------

func TestBreakStatement(t *testing.T) {
	prog := parse(t, `break;`)
	expectStmtCount(t, prog, 1)
	stmt := prog.Statements[0].(*ast.BreakStatement)
	if stmt.Label != nil {
		t.Error("expected no label")
	}
}

func TestBreakWithLabel(t *testing.T) {
	prog := parse(t, `break foo;`)
	stmt := prog.Statements[0].(*ast.BreakStatement)
	if stmt.Label == nil || stmt.Label.Value != "foo" {
		t.Error("expected label foo")
	}
}

func TestContinueStatement(t *testing.T) {
	prog := parse(t, `continue;`)
	stmt := prog.Statements[0].(*ast.ContinueStatement)
	if stmt.Label != nil {
		t.Error("expected no label")
	}
}

// ---------- Switch ----------

func TestSwitchStatement(t *testing.T) {
	prog := parse(t, `switch (x) { case 1: a; break; case 2: b; break; default: c; }`)
	expectStmtCount(t, prog, 1)
	stmt := prog.Statements[0].(*ast.SwitchStatement)
	if len(stmt.Cases) != 3 {
		t.Fatalf("expected 3 cases, got %d", len(stmt.Cases))
	}
	if stmt.Cases[2].Test != nil {
		t.Error("expected default case (nil test)")
	}
}

// ---------- Try/Catch/Finally ----------

func TestTryCatch(t *testing.T) {
	prog := parse(t, `try { a; } catch (e) { b; }`)
	expectStmtCount(t, prog, 1)
	stmt := prog.Statements[0].(*ast.TryStatement)
	if stmt.Handler == nil {
		t.Error("expected handler")
	}
	if stmt.Finalizer != nil {
		t.Error("expected no finalizer")
	}
}

func TestTryCatchFinally(t *testing.T) {
	prog := parse(t, `try { a; } catch (e) { b; } finally { c; }`)
	stmt := prog.Statements[0].(*ast.TryStatement)
	if stmt.Handler == nil {
		t.Error("expected handler")
	}
	if stmt.Finalizer == nil {
		t.Error("expected finalizer")
	}
}

func TestTryFinally(t *testing.T) {
	prog := parse(t, `try { a; } finally { b; }`)
	stmt := prog.Statements[0].(*ast.TryStatement)
	if stmt.Handler != nil {
		t.Error("expected no handler")
	}
	if stmt.Finalizer == nil {
		t.Error("expected finalizer")
	}
}

func TestOptionalCatchBinding(t *testing.T) {
	prog := parse(t, `try { a; } catch { b; }`)
	stmt := prog.Statements[0].(*ast.TryStatement)
	if stmt.Handler.Param != nil {
		t.Error("expected no catch param")
	}
}

// ---------- Throw ----------

func TestThrowStatement(t *testing.T) {
	prog := parse(t, `throw new Error("oops");`)
	expectStmtCount(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ThrowStatement)
	_, ok := stmt.Argument.(*ast.NewExpression)
	if !ok {
		t.Errorf("expected NewExpression, got %T", stmt.Argument)
	}
}

// ---------- Return ----------

func TestReturnStatement(t *testing.T) {
	prog := parse(t, `return 42;`)
	expectStmtCount(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ReturnStatement)
	num := stmt.Value.(*ast.NumberLiteral)
	if num.Value != 42 {
		t.Errorf("expected 42, got %f", num.Value)
	}
}

func TestReturnVoid(t *testing.T) {
	prog := parse(t, `return;`)
	stmt := prog.Statements[0].(*ast.ReturnStatement)
	if stmt.Value != nil {
		t.Error("expected nil return value")
	}
}

// ---------- Function Declaration ----------

func TestFunctionDeclaration(t *testing.T) {
	prog := parse(t, `function foo(a, b) { return a + b; }`)
	expectStmtCount(t, prog, 1)
	fn := prog.Statements[0].(*ast.FunctionDeclaration)
	if fn.Name.Value != "foo" {
		t.Errorf("expected foo, got %s", fn.Name.Value)
	}
	if len(fn.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(fn.Params))
	}
	if fn.Generator {
		t.Error("should not be generator")
	}
}

func TestGeneratorFunction(t *testing.T) {
	prog := parse(t, `function* gen() { yield 1; }`)
	fn := prog.Statements[0].(*ast.FunctionDeclaration)
	if !fn.Generator {
		t.Error("expected generator")
	}
}

func TestAsyncFunction(t *testing.T) {
	prog := parse(t, `async function foo() { await bar(); }`)
	fn := prog.Statements[0].(*ast.FunctionDeclaration)
	if !fn.Async {
		t.Error("expected async")
	}
}

func TestAsyncGeneratorFunction(t *testing.T) {
	prog := parse(t, `async function* foo() { yield 1; }`)
	fn := prog.Statements[0].(*ast.FunctionDeclaration)
	if !fn.Async {
		t.Error("expected async")
	}
	if !fn.Generator {
		t.Error("expected generator")
	}
}

func TestFunctionDefaultParams(t *testing.T) {
	prog := parse(t, `function foo(a, b = 1, c = 2) {}`)
	fn := prog.Statements[0].(*ast.FunctionDeclaration)
	if len(fn.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(fn.Params))
	}
	if len(fn.Defaults) != 3 {
		t.Fatalf("expected 3 defaults, got %d", len(fn.Defaults))
	}
	if fn.Defaults[0] != nil {
		t.Error("expected nil default for a")
	}
	if fn.Defaults[1] == nil {
		t.Error("expected non-nil default for b")
	}
}

func TestFunctionRestParams(t *testing.T) {
	prog := parse(t, `function foo(a, ...rest) {}`)
	fn := prog.Statements[0].(*ast.FunctionDeclaration)
	if len(fn.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(fn.Params))
	}
	if fn.Rest == nil {
		t.Fatal("expected rest parameter")
	}
	rest, ok := fn.Rest.(*ast.RestElement)
	if !ok {
		t.Fatalf("expected RestElement, got %T", fn.Rest)
	}
	ident := rest.Argument.(*ast.Identifier)
	if ident.Value != "rest" {
		t.Errorf("expected rest, got %s", ident.Value)
	}
}

// ---------- Class Declaration ----------

func TestClassDeclaration(t *testing.T) {
	input := `class Foo {
		constructor(x) { this.x = x; }
		method() { return this.x; }
		static create() { return new Foo(1); }
	}`
	prog := parse(t, input)
	expectStmtCount(t, prog, 1)
	cls := prog.Statements[0].(*ast.ClassDeclaration)
	if cls.Name.Value != "Foo" {
		t.Errorf("expected Foo, got %s", cls.Name.Value)
	}
	if len(cls.Body.Methods) != 3 {
		t.Fatalf("expected 3 methods, got %d", len(cls.Body.Methods))
	}
	if cls.Body.Methods[0].Kind != "constructor" {
		t.Errorf("expected constructor, got %s", cls.Body.Methods[0].Kind)
	}
	if cls.Body.Methods[1].Kind != "method" {
		t.Errorf("expected method, got %s", cls.Body.Methods[1].Kind)
	}
	if !cls.Body.Methods[2].Static {
		t.Error("expected static")
	}
}

func TestClassWithExtends(t *testing.T) {
	input := `class Bar extends Foo { constructor() { super(); } }`
	prog := parse(t, input)
	cls := prog.Statements[0].(*ast.ClassDeclaration)
	if cls.SuperClass == nil {
		t.Error("expected superclass")
	}
	super := cls.SuperClass.(*ast.Identifier)
	if super.Value != "Foo" {
		t.Errorf("expected Foo, got %s", super.Value)
	}
}

func TestClassGetterSetter(t *testing.T) {
	input := `class Foo {
		get x() { return this._x; }
		set x(v) { this._x = v; }
	}`
	prog := parse(t, input)
	cls := prog.Statements[0].(*ast.ClassDeclaration)
	if len(cls.Body.Methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(cls.Body.Methods))
	}
	if cls.Body.Methods[0].Kind != "get" {
		t.Errorf("expected get, got %s", cls.Body.Methods[0].Kind)
	}
	if cls.Body.Methods[1].Kind != "set" {
		t.Errorf("expected set, got %s", cls.Body.Methods[1].Kind)
	}
}

func TestClassComputedMethod(t *testing.T) {
	input := `class Foo { [Symbol.iterator]() {} }`
	prog := parse(t, input)
	cls := prog.Statements[0].(*ast.ClassDeclaration)
	if !cls.Body.Methods[0].Computed {
		t.Error("expected computed method")
	}
}

// ---------- Arrow Functions ----------

func TestArrowFunctionExpression(t *testing.T) {
	prog := parse(t, `const f = (a, b) => a + b;`)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	arrow, ok := decl.Declarations[0].Value.(*ast.ArrowFunctionExpression)
	if !ok {
		t.Fatalf("expected ArrowFunctionExpression, got %T", decl.Declarations[0].Value)
	}
	if len(arrow.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(arrow.Params))
	}
	if _, ok := arrow.Body.(*ast.BinaryExpression); !ok {
		t.Errorf("expected expression body, got %T", arrow.Body)
	}
}

func TestArrowFunctionBlockBody(t *testing.T) {
	prog := parse(t, `const f = (x) => { return x; };`)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	arrow := decl.Declarations[0].Value.(*ast.ArrowFunctionExpression)
	if _, ok := arrow.Body.(*ast.BlockStatement); !ok {
		t.Errorf("expected BlockStatement body, got %T", arrow.Body)
	}
}

func TestArrowFunctionSingleParam(t *testing.T) {
	prog := parse(t, `const f = x => x * 2;`)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	arrow := decl.Declarations[0].Value.(*ast.ArrowFunctionExpression)
	if len(arrow.Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(arrow.Params))
	}
}

func TestArrowFunctionNoParams(t *testing.T) {
	prog := parse(t, `const f = () => 42;`)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	arrow := decl.Declarations[0].Value.(*ast.ArrowFunctionExpression)
	if len(arrow.Params) != 0 {
		t.Errorf("expected 0 params, got %d", len(arrow.Params))
	}
}

func TestAsyncArrowFunction(t *testing.T) {
	prog := parse(t, `const f = async (x) => await x;`)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	arrow, ok := decl.Declarations[0].Value.(*ast.ArrowFunctionExpression)
	if !ok {
		t.Fatalf("expected ArrowFunctionExpression, got %T", decl.Declarations[0].Value)
	}
	if !arrow.Async {
		t.Error("expected async")
	}
}

func TestAsyncArrowSingleParam(t *testing.T) {
	prog := parse(t, `const f = async x => x;`)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	arrow, ok := decl.Declarations[0].Value.(*ast.ArrowFunctionExpression)
	if !ok {
		t.Fatalf("expected ArrowFunctionExpression, got %T", decl.Declarations[0].Value)
	}
	if !arrow.Async {
		t.Error("expected async")
	}
	if len(arrow.Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(arrow.Params))
	}
}

// ---------- Array Literals ----------

func TestArrayLiteral(t *testing.T) {
	prog := parse(t, `[1, 2, 3];`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	arr, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("expected ArrayLiteral, got %T", stmt.Expression)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr.Elements))
	}
}

func TestArrayWithElision(t *testing.T) {
	prog := parse(t, `[1, , 3];`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	arr := stmt.Expression.(*ast.ArrayLiteral)
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}
	if arr.Elements[1] != nil {
		t.Error("expected nil for elision")
	}
}

func TestArrayWithSpread(t *testing.T) {
	prog := parse(t, `[1, ...rest, 3];`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	arr := stmt.Expression.(*ast.ArrayLiteral)
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}
	_, ok := arr.Elements[1].(*ast.SpreadElement)
	if !ok {
		t.Errorf("expected SpreadElement, got %T", arr.Elements[1])
	}
}

// ---------- Object Literals ----------

func TestObjectLiteral(t *testing.T) {
	prog := parse(t, `({ a: 1, b: 2 });`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	obj, ok := stmt.Expression.(*ast.ObjectLiteral)
	if !ok {
		t.Fatalf("expected ObjectLiteral, got %T", stmt.Expression)
	}
	if len(obj.Properties) != 2 {
		t.Errorf("expected 2 properties, got %d", len(obj.Properties))
	}
}

func TestObjectShorthand(t *testing.T) {
	prog := parse(t, `({ x, y });`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	obj := stmt.Expression.(*ast.ObjectLiteral)
	if len(obj.Properties) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(obj.Properties))
	}
	if !obj.Properties[0].Shorthand {
		t.Error("expected shorthand property")
	}
}

func TestObjectMethod(t *testing.T) {
	prog := parse(t, `({ foo() { return 1; } });`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	obj := stmt.Expression.(*ast.ObjectLiteral)
	if !obj.Properties[0].Method {
		t.Error("expected method property")
	}
}

func TestObjectComputedKey(t *testing.T) {
	prog := parse(t, `({ [key]: value });`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	obj := stmt.Expression.(*ast.ObjectLiteral)
	if !obj.Properties[0].Computed {
		t.Error("expected computed property")
	}
}

func TestObjectGetterSetter(t *testing.T) {
	prog := parse(t, `({ get x() { return 1; }, set x(v) {} });`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	obj := stmt.Expression.(*ast.ObjectLiteral)
	if obj.Properties[0].Kind != "get" {
		t.Errorf("expected get, got %s", obj.Properties[0].Kind)
	}
	if obj.Properties[1].Kind != "set" {
		t.Errorf("expected set, got %s", obj.Properties[1].Kind)
	}
}

func TestObjectGeneratorMethod(t *testing.T) {
	prog := parse(t, `({ *gen() { yield 1; } });`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	obj := stmt.Expression.(*ast.ObjectLiteral)
	if !obj.Properties[0].Method {
		t.Error("expected method")
	}
	fe := obj.Properties[0].Value.(*ast.FunctionExpression)
	if !fe.Generator {
		t.Error("expected generator")
	}
}

func TestObjectSpread(t *testing.T) {
	prog := parse(t, `({ ...other, a: 1 });`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	obj := stmt.Expression.(*ast.ObjectLiteral)
	if len(obj.Properties) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(obj.Properties))
	}
}

// ---------- Template Literals ----------

func TestTemplateLiteralSimple(t *testing.T) {
	prog := parse(t, "`hello`;")
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	tmpl, ok := stmt.Expression.(*ast.TemplateLiteralExpr)
	if !ok {
		t.Fatalf("expected TemplateLiteralExpr, got %T", stmt.Expression)
	}
	if len(tmpl.Quasis) != 1 {
		t.Fatalf("expected 1 quasi, got %d", len(tmpl.Quasis))
	}
	if tmpl.Quasis[0].Value != "hello" {
		t.Errorf("expected hello, got %s", tmpl.Quasis[0].Value)
	}
}

func TestTemplateLiteralWithExpressions(t *testing.T) {
	prog := parse(t, "`hello ${name} world`;")
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	tmpl := stmt.Expression.(*ast.TemplateLiteralExpr)
	if len(tmpl.Quasis) != 2 {
		t.Fatalf("expected 2 quasis, got %d", len(tmpl.Quasis))
	}
	if len(tmpl.Expressions) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(tmpl.Expressions))
	}
}

func TestTaggedTemplate(t *testing.T) {
	prog := parse(t, "tag`hello`;")
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	tagged, ok := stmt.Expression.(*ast.TaggedTemplateExpression)
	if !ok {
		t.Fatalf("expected TaggedTemplateExpression, got %T", stmt.Expression)
	}
	tagIdent := tagged.Tag.(*ast.Identifier)
	if tagIdent.Value != "tag" {
		t.Errorf("expected tag, got %s", tagIdent.Value)
	}
}

// ---------- This and Super ----------

func TestThisExpression(t *testing.T) {
	prog := parse(t, `this;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	_, ok := stmt.Expression.(*ast.ThisExpression)
	if !ok {
		t.Fatalf("expected ThisExpression, got %T", stmt.Expression)
	}
}

func TestSuperExpression(t *testing.T) {
	prog := parse(t, `super.method();`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", stmt.Expression)
	}
	mem := call.Callee.(*ast.MemberExpression)
	_, ok = mem.Object.(*ast.SuperExpression)
	if !ok {
		t.Errorf("expected SuperExpression, got %T", mem.Object)
	}
}

// ---------- Yield ----------

func TestYieldExpression(t *testing.T) {
	prog := parse(t, `yield 1;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	yld, ok := stmt.Expression.(*ast.YieldExpression)
	if !ok {
		t.Fatalf("expected YieldExpression, got %T", stmt.Expression)
	}
	if yld.Delegate {
		t.Error("should not be delegate")
	}
}

func TestYieldDelegateExpression(t *testing.T) {
	prog := parse(t, `yield* gen();`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	yld := stmt.Expression.(*ast.YieldExpression)
	if !yld.Delegate {
		t.Error("expected delegate")
	}
}

// ---------- Await ----------

func TestAwaitExpression(t *testing.T) {
	prog := parse(t, `await promise;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	aw, ok := stmt.Expression.(*ast.AwaitExpression)
	if !ok {
		t.Fatalf("expected AwaitExpression, got %T", stmt.Expression)
	}
	arg := aw.Argument.(*ast.Identifier)
	if arg.Value != "promise" {
		t.Errorf("expected promise, got %s", arg.Value)
	}
}

// ---------- Labeled Statement ----------

func TestLabeledStatement(t *testing.T) {
	prog := parse(t, `outer: for (;;) { break outer; }`)
	expectStmtCount(t, prog, 1)
	stmt := prog.Statements[0].(*ast.LabeledStatement)
	if stmt.Label.Value != "outer" {
		t.Errorf("expected outer, got %s", stmt.Label.Value)
	}
}

// ---------- Debugger ----------

func TestDebuggerStatement(t *testing.T) {
	prog := parse(t, `debugger;`)
	expectStmtCount(t, prog, 1)
	_, ok := prog.Statements[0].(*ast.DebuggerStatement)
	if !ok {
		t.Fatalf("expected DebuggerStatement, got %T", prog.Statements[0])
	}
}

// ---------- Empty Statement ----------

func TestEmptyStatement(t *testing.T) {
	prog := parse(t, `;`)
	expectStmtCount(t, prog, 1)
	_, ok := prog.Statements[0].(*ast.EmptyStatement)
	if !ok {
		t.Fatalf("expected EmptyStatement, got %T", prog.Statements[0])
	}
}

// ---------- With Statement ----------

func TestWithStatement(t *testing.T) {
	prog := parse(t, `with (obj) { x; }`)
	expectStmtCount(t, prog, 1)
	_, ok := prog.Statements[0].(*ast.WithStatement)
	if !ok {
		t.Fatalf("expected WithStatement, got %T", prog.Statements[0])
	}
}

// ---------- Block Statement ----------

func TestBlockStatement(t *testing.T) {
	prog := parse(t, `{ let x = 1; let y = 2; }`)
	expectStmtCount(t, prog, 1)
	block := prog.Statements[0].(*ast.BlockStatement)
	if len(block.Statements) != 2 {
		t.Errorf("expected 2 statements in block, got %d", len(block.Statements))
	}
}

// ---------- Function Expression ----------

func TestFunctionExpression(t *testing.T) {
	prog := parse(t, `const f = function(x) { return x; };`)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	fe, ok := decl.Declarations[0].Value.(*ast.FunctionExpression)
	if !ok {
		t.Fatalf("expected FunctionExpression, got %T", decl.Declarations[0].Value)
	}
	if fe.Name != nil {
		t.Error("expected anonymous function")
	}
}

func TestNamedFunctionExpression(t *testing.T) {
	prog := parse(t, `const f = function foo(x) { return x; };`)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	fe := decl.Declarations[0].Value.(*ast.FunctionExpression)
	if fe.Name == nil || fe.Name.Value != "foo" {
		t.Error("expected named function foo")
	}
}

func TestGeneratorExpression(t *testing.T) {
	prog := parse(t, `const g = function*() { yield 1; };`)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	fe := decl.Declarations[0].Value.(*ast.FunctionExpression)
	if !fe.Generator {
		t.Error("expected generator")
	}
}

// ---------- Class Expression ----------

func TestClassExpression(t *testing.T) {
	prog := parse(t, `const C = class { constructor() {} };`)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	cls, ok := decl.Declarations[0].Value.(*ast.ClassExpression)
	if !ok {
		t.Fatalf("expected ClassExpression, got %T", decl.Declarations[0].Value)
	}
	if cls.Name != nil {
		t.Error("expected anonymous class")
	}
}

func TestNamedClassExpression(t *testing.T) {
	prog := parse(t, `const C = class Foo { constructor() {} };`)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	cls := decl.Declarations[0].Value.(*ast.ClassExpression)
	if cls.Name == nil || cls.Name.Value != "Foo" {
		t.Error("expected named class Foo")
	}
}

// ---------- Sequence Expression ----------

func TestSequenceExpression(t *testing.T) {
	prog := parse(t, `a, b, c;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	seq, ok := stmt.Expression.(*ast.SequenceExpression)
	if !ok {
		t.Fatalf("expected SequenceExpression, got %T", stmt.Expression)
	}
	if len(seq.Expressions) != 3 {
		t.Errorf("expected 3 expressions, got %d", len(seq.Expressions))
	}
}

// ---------- Complex Expressions ----------

func TestChainedCallsAndMembers(t *testing.T) {
	prog := parse(t, `a.b().c[0].d();`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	// Should end with CallExpression
	_, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", stmt.Expression)
	}
}

func TestNestedTernary(t *testing.T) {
	prog := parse(t, `a ? b ? c : d : e;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	cond := stmt.Expression.(*ast.ConditionalExpression)
	_, ok := cond.Consequent.(*ast.ConditionalExpression)
	if !ok {
		t.Errorf("expected nested conditional, got %T", cond.Consequent)
	}
}

func TestGroupedExpression(t *testing.T) {
	prog := parse(t, `(1 + 2) * 3;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	mul := stmt.Expression.(*ast.BinaryExpression)
	if mul.Operator != "*" {
		t.Errorf("expected *, got %s", mul.Operator)
	}
	add, ok := mul.Left.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected BinaryExpression on left, got %T", mul.Left)
	}
	if add.Operator != "+" {
		t.Errorf("expected +, got %s", add.Operator)
	}
}

// ---------- In operator context ----------

func TestInOperator(t *testing.T) {
	prog := parse(t, `"x" in obj;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	bin := stmt.Expression.(*ast.BinaryExpression)
	if bin.Operator != "in" {
		t.Errorf("expected in, got %s", bin.Operator)
	}
}

// ---------- Number formats ----------

func TestHexNumber(t *testing.T) {
	prog := parse(t, `0xFF;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	num := stmt.Expression.(*ast.NumberLiteral)
	if num.Value != 255 {
		t.Errorf("expected 255, got %f", num.Value)
	}
}

func TestOctalNumber(t *testing.T) {
	prog := parse(t, `0o77;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	num := stmt.Expression.(*ast.NumberLiteral)
	if num.Value != 63 {
		t.Errorf("expected 63, got %f", num.Value)
	}
}

func TestBinaryNumber(t *testing.T) {
	prog := parse(t, `0b1010;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	num := stmt.Expression.(*ast.NumberLiteral)
	if num.Value != 10 {
		t.Errorf("expected 10, got %f", num.Value)
	}
}

func TestFloatNumber(t *testing.T) {
	prog := parse(t, `3.14;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	num := stmt.Expression.(*ast.NumberLiteral)
	if num.Value != 3.14 {
		t.Errorf("expected 3.14, got %f", num.Value)
	}
}

// ---------- Error Reporting ----------

func TestParseErrors(t *testing.T) {
	_, errs := parseWithErrors(`if (`)
	if len(errs) == 0 {
		t.Error("expected parse errors")
	}
}

// ---------- Complex Programs ----------

func TestComplexProgram(t *testing.T) {
	input := `
		class Animal {
			constructor(name) {
				this.name = name;
			}
			speak() {
				return this.name;
			}
		}

		class Dog extends Animal {
			constructor(name) {
				super(name);
			}
			speak() {
				return super.speak() + " barks";
			}
		}

		const dog = new Dog("Rex");
		const result = dog.speak();
	`
	prog := parse(t, input)
	if len(prog.Statements) != 4 {
		t.Errorf("expected 4 statements, got %d", len(prog.Statements))
	}
}

func TestIteratorPattern(t *testing.T) {
	input := `
		function* range(start, end) {
			for (let i = start; i < end; i++) {
				yield i;
			}
		}
		for (const n of range(0, 10)) {
			console.log(n);
		}
	`
	prog := parse(t, input)
	if len(prog.Statements) != 2 {
		t.Errorf("expected 2 statements, got %d", len(prog.Statements))
	}
}

func TestAsyncAwaitPattern(t *testing.T) {
	input := `
		async function fetchData(url) {
			const response = await fetch(url);
			const data = await response.json();
			return data;
		}
	`
	prog := parse(t, input)
	if len(prog.Statements) != 1 {
		t.Errorf("expected 1 statement, got %d", len(prog.Statements))
	}
	fn := prog.Statements[0].(*ast.FunctionDeclaration)
	if !fn.Async {
		t.Error("expected async")
	}
}

func TestDestructuringComplex(t *testing.T) {
	input := `
		const { a: { b, c }, d: [e, f] } = obj;
	`
	prog := parse(t, input)
	expectStmtCount(t, prog, 1)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	pat := decl.Declarations[0].Name.(*ast.ObjectPattern)
	if len(pat.Properties) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(pat.Properties))
	}
}

func TestArrowFunctionWithDefaults(t *testing.T) {
	prog := parse(t, `const f = (a = 1, b = 2) => a + b;`)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	arrow := decl.Declarations[0].Value.(*ast.ArrowFunctionExpression)
	if len(arrow.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(arrow.Params))
	}
	if len(arrow.Defaults) != 2 {
		t.Errorf("expected 2 defaults, got %d", len(arrow.Defaults))
	}
}

func TestArrowFunctionWithRest(t *testing.T) {
	prog := parse(t, `const f = (a, ...rest) => rest;`)
	decl := prog.Statements[0].(*ast.VariableDeclaration)
	arrow := decl.Declarations[0].Value.(*ast.ArrowFunctionExpression)
	if len(arrow.Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(arrow.Params))
	}
	if arrow.Rest == nil {
		t.Error("expected rest param")
	}
}

func TestForEmptyParts(t *testing.T) {
	prog := parse(t, `for (;;) { break; }`)
	stmt := prog.Statements[0].(*ast.ForStatement)
	if stmt.Init != nil {
		t.Error("expected nil init")
	}
	if stmt.Test != nil {
		t.Error("expected nil test")
	}
	if stmt.Update != nil {
		t.Error("expected nil update")
	}
}

func TestNullishCoalescing(t *testing.T) {
	prog := parse(t, `a ?? b;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	bin := stmt.Expression.(*ast.BinaryExpression)
	if bin.Operator != "??" {
		t.Errorf("expected ??, got %s", bin.Operator)
	}
}

func TestOptionalChaining(t *testing.T) {
	prog := parse(t, `a?.b;`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	mem, ok := stmt.Expression.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("expected MemberExpression, got %T", stmt.Expression)
	}
	obj := mem.Object.(*ast.Identifier)
	if obj.Value != "a" {
		t.Errorf("expected a, got %s", obj.Value)
	}
}

func TestMultipleTemplateLiteralExpressions(t *testing.T) {
	prog := parse(t, "`${a} ${b}`;")
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	tmpl := stmt.Expression.(*ast.TemplateLiteralExpr)
	if len(tmpl.Expressions) != 2 {
		t.Errorf("expected 2 expressions, got %d", len(tmpl.Expressions))
	}
	if len(tmpl.Quasis) != 3 {
		t.Errorf("expected 3 quasis, got %d", len(tmpl.Quasis))
	}
}

func TestForInExpressionLeft(t *testing.T) {
	prog := parse(t, `for (x in obj) {}`)
	stmt, ok := prog.Statements[0].(*ast.ForInStatement)
	if !ok {
		t.Fatalf("expected ForInStatement, got %T", prog.Statements[0])
	}
	ident, ok := stmt.Left.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier, got %T", stmt.Left)
	}
	if ident.Value != "x" {
		t.Errorf("expected x, got %s", ident.Value)
	}
}

func TestObjectAsyncMethod(t *testing.T) {
	prog := parse(t, `({ async foo() {} });`)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	obj := stmt.Expression.(*ast.ObjectLiteral)
	if len(obj.Properties) != 1 {
		t.Fatalf("expected 1 property, got %d", len(obj.Properties))
	}
	fe := obj.Properties[0].Value.(*ast.FunctionExpression)
	if !fe.Async {
		t.Error("expected async method")
	}
}

func TestClassAsyncMethod(t *testing.T) {
	input := `class Foo { async bar() {} }`
	prog := parse(t, input)
	cls := prog.Statements[0].(*ast.ClassDeclaration)
	if len(cls.Body.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(cls.Body.Methods))
	}
	if !cls.Body.Methods[0].Value.Async {
		t.Error("expected async method")
	}
}

func TestClassGeneratorMethod(t *testing.T) {
	input := `class Foo { *gen() { yield 1; } }`
	prog := parse(t, input)
	cls := prog.Statements[0].(*ast.ClassDeclaration)
	if !cls.Body.Methods[0].Value.Generator {
		t.Error("expected generator method")
	}
}
