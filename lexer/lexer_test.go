package lexer

import (
	"testing"

	"github.com/example/jsgo/token"
)

func TestSingleCharTokens(t *testing.T) {
	input := `( ) { } [ ] ; : , ~`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.LeftParen, "("},
		{token.RightParen, ")"},
		{token.LeftBrace, "{"},
		{token.RightBrace, "}"},
		{token.LeftBracket, "["},
		{token.RightBracket, "]"},
		{token.Semicolon, ";"},
		{token.Colon, ":"},
		{token.Comma, ","},
		{token.BitwiseNot, "~"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestArithmeticOperators(t *testing.T) {
	input := `+ - * / % **`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Plus, "+"},
		{token.Minus, "-"},
		{token.Asterisk, "*"},
		{token.Slash, "/"},
		{token.Percent, "%"},
		{token.Exponent, "**"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestComparisonOperators(t *testing.T) {
	input := `== != === !== < > <= >=`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Equal, "=="},
		{token.NotEqual, "!="},
		{token.StrictEqual, "==="},
		{token.StrictNotEqual, "!=="},
		{token.LessThan, "<"},
		{token.GreaterThan, ">"},
		{token.LessThanOrEqual, "<="},
		{token.GreaterThanOrEqual, ">="},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestLogicalOperators(t *testing.T) {
	input := `&& || ! ?? ??=`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.And, "&&"},
		{token.Or, "||"},
		{token.Not, "!"},
		{token.NullishCoalesce, "??"},
		{token.NullishAssign, "??="},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestBitwiseOperators(t *testing.T) {
	input := `& | ^ ~ << >> >>>`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.BitwiseAnd, "&"},
		{token.BitwiseOr, "|"},
		{token.BitwiseXor, "^"},
		{token.BitwiseNot, "~"},
		{token.LeftShift, "<<"},
		{token.RightShift, ">>"},
		{token.UnsignedRightShift, ">>>"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestAssignmentOperators(t *testing.T) {
	input := `= += -= *= /= %= **= &= |= ^= <<= >>= >>>= &&= ||= ??=`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Assign, "="},
		{token.PlusAssign, "+="},
		{token.MinusAssign, "-="},
		{token.AsteriskAssign, "*="},
		{token.SlashAssign, "/="},
		{token.PercentAssign, "%="},
		{token.ExponentAssign, "**="},
		{token.AmpersandAssign, "&="},
		{token.PipeAssign, "|="},
		{token.CaretAssign, "^="},
		{token.LeftShiftAssign, "<<="},
		{token.RightShiftAssign, ">>="},
		{token.UnsignedRightShiftAssign, ">>>="},
		{token.AndAssign, "&&="},
		{token.OrAssign, "||="},
		{token.NullishAssign, "??="},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestIncrementDecrement(t *testing.T) {
	input := `++ --`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Increment, "++"},
		{token.Decrement, "--"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestDotAndSpread(t *testing.T) {
	input := `a.b ...c`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Identifier, "a"},
		{token.Dot, "."},
		{token.Identifier, "b"},
		{token.Spread, "..."},
		{token.Identifier, "c"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestArrow(t *testing.T) {
	input := `=>`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.Arrow {
		t.Errorf("expected Arrow, got %d (lit=%q)", tok.Type, tok.Literal)
	}
	if tok.Literal != "=>" {
		t.Errorf("expected '=>', got %q", tok.Literal)
	}
}

func TestOptionalChainAndQuestion(t *testing.T) {
	input := `a?.b ? c`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Identifier, "a"},
		{token.OptionalChain, "?."},
		{token.Identifier, "b"},
		{token.QuestionMark, "?"},
		{token.Identifier, "c"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestKeywords(t *testing.T) {
	input := `var let const function return if else while for do break continue switch case default throw try catch finally new delete typeof void in instanceof this class extends super import export from as of yield async await true false null undefined debugger with`

	keywords := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Var, "var"},
		{token.Let, "let"},
		{token.Const, "const"},
		{token.Function, "function"},
		{token.Return, "return"},
		{token.If, "if"},
		{token.Else, "else"},
		{token.While, "while"},
		{token.For, "for"},
		{token.Do, "do"},
		{token.Break, "break"},
		{token.Continue, "continue"},
		{token.Switch, "switch"},
		{token.Case, "case"},
		{token.Default, "default"},
		{token.Throw, "throw"},
		{token.Try, "try"},
		{token.Catch, "catch"},
		{token.Finally, "finally"},
		{token.New, "new"},
		{token.Delete, "delete"},
		{token.Typeof, "typeof"},
		{token.Void, "void"},
		{token.In, "in"},
		{token.Instanceof, "instanceof"},
		{token.This, "this"},
		{token.Class, "class"},
		{token.Extends, "extends"},
		{token.Super, "super"},
		{token.Import, "import"},
		{token.Export, "export"},
		{token.From, "from"},
		{token.As, "as"},
		{token.Of, "of"},
		{token.Yield, "yield"},
		{token.Async, "async"},
		{token.Await, "await"},
		{token.True, "true"},
		{token.False, "false"},
		{token.Null, "null"},
		{token.Undefined, "undefined"},
		{token.Debugger, "debugger"},
		{token.With, "with"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range keywords {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong for %q. expected=%d, got=%d", i, exp.lit, exp.typ, tok.Type)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestIdentifiers(t *testing.T) {
	tests := []struct {
		input string
		lit   string
	}{
		{"foo", "foo"},
		{"_bar", "_bar"},
		{"$baz", "$baz"},
		{"camelCase", "camelCase"},
		{"PascalCase", "PascalCase"},
		{"snake_case", "snake_case"},
		{"_$mixed123", "_$mixed123"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != token.Identifier {
			t.Errorf("input=%q: type wrong. expected=Identifier, got=%d", tt.input, tok.Type)
		}
		if tok.Literal != tt.lit {
			t.Errorf("input=%q: literal wrong. expected=%q, got=%q", tt.input, tt.lit, tok.Literal)
		}
	}
}

func TestNumberLiterals(t *testing.T) {
	tests := []struct {
		input string
		lit   string
	}{
		{"0", "0"},
		{"42", "42"},
		{"3.14", "3.14"},
		{".5", ".5"},
		{"1e10", "1e10"},
		{"1E10", "1E10"},
		{"1.5e+3", "1.5e+3"},
		{"1.5e-3", "1.5e-3"},
		{"0xFF", "0xFF"},
		{"0XAB", "0XAB"},
		{"0o77", "0o77"},
		{"0O77", "0O77"},
		{"0b1010", "0b1010"},
		{"0B1010", "0B1010"},
		{"1_000_000", "1_000_000"},
		{"0xFF_FF", "0xFF_FF"},
		{"0b1010_0101", "0b1010_0101"},
		{"100n", "100n"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != token.Number {
			t.Errorf("input=%q: type wrong. expected=Number, got=%d (lit=%q)", tt.input, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.lit {
			t.Errorf("input=%q: literal wrong. expected=%q, got=%q", tt.input, tt.lit, tok.Literal)
		}
	}
}

func TestStringLiterals(t *testing.T) {
	tests := []struct {
		input string
		lit   string
	}{
		{`"hello"`, "hello"},
		{`'hello'`, "hello"},
		{`""`, ""},
		{`''`, ""},
		{`"hello world"`, "hello world"},
		{`"escape\nnewline"`, "escape\nnewline"},
		{`"tab\there"`, "tab\there"},
		{`"back\\slash"`, "back\\slash"},
		{`"quote\""`, `quote"`},
		{`'quote\''`, `quote'`},
		{`"null\0char"`, "null\x00char"},
		{`"\x41"`, "A"},
		{`"\u0041"`, "A"},
		{`"\u{41}"`, "A"},
		{`"\u{1F600}"`, "\U0001F600"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != token.String {
			t.Errorf("input=%q: type wrong. expected=String, got=%d (lit=%q)", tt.input, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.lit {
			t.Errorf("input=%q: literal wrong. expected=%q, got=%q", tt.input, tt.lit, tok.Literal)
		}
	}
}

func TestUnterminatedString(t *testing.T) {
	l := New(`"hello`)
	tok := l.NextToken()
	if tok.Type != token.Illegal {
		t.Errorf("expected Illegal for unterminated string, got %d", tok.Type)
	}
}

func TestNoSubstitutionTemplate(t *testing.T) {
	input := "`hello world`"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.NoSubstitutionTemplate {
		t.Errorf("expected NoSubstitutionTemplate, got %d (lit=%q)", tok.Type, tok.Literal)
	}
	if tok.Literal != "hello world" {
		t.Errorf("expected 'hello world', got %q", tok.Literal)
	}
}

func TestTemplateLiteralWithInterpolation(t *testing.T) {
	input := "`hello ${name}!`"
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.TemplateHead, "hello "},
		{token.Identifier, "name"},
		{token.TemplateTail, "!"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestTemplateLiteralMultipleInterpolations(t *testing.T) {
	input := "`${a} and ${b}`"
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.TemplateHead, ""},
		{token.Identifier, "a"},
		{token.TemplateMiddle, " and "},
		{token.Identifier, "b"},
		{token.TemplateTail, ""},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestTemplateLiteralNestedBraces(t *testing.T) {
	input := "`${a + {b: 1}.b}`"
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.TemplateHead, ""},
		{token.Identifier, "a"},
		{token.Plus, "+"},
		{token.LeftBrace, "{"},
		{token.Identifier, "b"},
		{token.Colon, ":"},
		{token.Number, "1"},
		{token.RightBrace, "}"},
		{token.Dot, "."},
		{token.Identifier, "b"},
		{token.TemplateTail, ""},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestTemplateEscapes(t *testing.T) {
	input := "`\\n\\t\\\\\\``"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.NoSubstitutionTemplate {
		t.Errorf("expected NoSubstitutionTemplate, got %d", tok.Type)
	}
	if tok.Literal != "\n\t\\`" {
		t.Errorf("expected '\\n\\t\\\\`', got %q", tok.Literal)
	}
}

func TestRegExpLiteral(t *testing.T) {
	tests := []struct {
		input string
		lit   string
	}{
		{"/abc/", "/abc/"},
		{"/abc/gi", "/abc/gi"},
		{"/[a-z]+/", "/[a-z]+/"},
		{`/foo\/bar/`, `/foo\/bar/`},
		{"/[/]/", "/[/]/"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextTokenWithRegex(token.EOF) // EOF = start of input, regex is valid
		if tok.Type != token.RegExp {
			t.Errorf("input=%q: type wrong. expected=RegExp, got=%d (lit=%q)", tt.input, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.lit {
			t.Errorf("input=%q: literal wrong. expected=%q, got=%q", tt.input, tt.lit, tok.Literal)
		}
	}
}

func TestRegExpVsDivision(t *testing.T) {
	// After an identifier, '/' is division
	input := "a / b"
	tokens := Tokenize(input)
	if tokens[1].Type != token.Slash {
		t.Errorf("expected Slash after identifier, got %d", tokens[1].Type)
	}

	// After a number, '/' is division
	input = "1 / 2"
	tokens = Tokenize(input)
	if tokens[1].Type != token.Slash {
		t.Errorf("expected Slash after number, got %d", tokens[1].Type)
	}

	// After '=', '/' starts a regex
	input = "x = /foo/g"
	tokens = Tokenize(input)
	if tokens[2].Type != token.RegExp {
		t.Errorf("expected RegExp after '=', got %d (lit=%q)", tokens[2].Type, tokens[2].Literal)
	}

	// After '(', '/' starts a regex
	input = "(/abc/)"
	tokens = Tokenize(input)
	if tokens[1].Type != token.RegExp {
		t.Errorf("expected RegExp after '(', got %d (lit=%q)", tokens[1].Type, tokens[1].Literal)
	}
}

func TestLineComments(t *testing.T) {
	input := `a // this is a comment
b`
	l := New(input)
	tok1 := l.NextToken()
	if tok1.Type != token.Identifier || tok1.Literal != "a" {
		t.Errorf("expected identifier 'a', got %d %q", tok1.Type, tok1.Literal)
	}
	tok2 := l.NextToken()
	if tok2.Type != token.Identifier || tok2.Literal != "b" {
		t.Errorf("expected identifier 'b', got %d %q", tok2.Type, tok2.Literal)
	}
}

func TestBlockComments(t *testing.T) {
	input := `a /* block
comment */ b`
	l := New(input)
	tok1 := l.NextToken()
	if tok1.Type != token.Identifier || tok1.Literal != "a" {
		t.Errorf("expected identifier 'a', got %d %q", tok1.Type, tok1.Literal)
	}
	tok2 := l.NextToken()
	if tok2.Type != token.Identifier || tok2.Literal != "b" {
		t.Errorf("expected identifier 'b', got %d %q", tok2.Type, tok2.Literal)
	}
}

func TestLineTracking(t *testing.T) {
	input := "a\nb\nc"
	l := New(input)

	tok := l.NextToken()
	if tok.Line != 1 {
		t.Errorf("token 'a': expected line 1, got %d", tok.Line)
	}

	tok = l.NextToken()
	if tok.Line != 2 {
		t.Errorf("token 'b': expected line 2, got %d", tok.Line)
	}

	tok = l.NextToken()
	if tok.Line != 3 {
		t.Errorf("token 'c': expected line 3, got %d", tok.Line)
	}
}

func TestColumnTracking(t *testing.T) {
	input := "ab cd"
	l := New(input)

	tok := l.NextToken()
	if tok.Column != 1 {
		t.Errorf("token 'ab': expected col 1, got %d", tok.Column)
	}

	tok = l.NextToken()
	if tok.Column != 4 {
		t.Errorf("token 'cd': expected col 4, got %d", tok.Column)
	}
}

func TestLetStatement(t *testing.T) {
	input := `let x = 5;`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Let, "let"},
		{token.Identifier, "x"},
		{token.Assign, "="},
		{token.Number, "5"},
		{token.Semicolon, ";"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestArrowFunction(t *testing.T) {
	input := `(x) => x + 1`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.LeftParen, "("},
		{token.Identifier, "x"},
		{token.RightParen, ")"},
		{token.Arrow, "=>"},
		{token.Identifier, "x"},
		{token.Plus, "+"},
		{token.Number, "1"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestClassDeclaration(t *testing.T) {
	input := `class Foo extends Bar { constructor() { super(); } }`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Class, "class"},
		{token.Identifier, "Foo"},
		{token.Extends, "extends"},
		{token.Identifier, "Bar"},
		{token.LeftBrace, "{"},
		{token.Identifier, "constructor"},
		{token.LeftParen, "("},
		{token.RightParen, ")"},
		{token.LeftBrace, "{"},
		{token.Super, "super"},
		{token.LeftParen, "("},
		{token.RightParen, ")"},
		{token.Semicolon, ";"},
		{token.RightBrace, "}"},
		{token.RightBrace, "}"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestAsyncAwait(t *testing.T) {
	input := `async function fetchData() { const data = await fetch(); }`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Async, "async"},
		{token.Function, "function"},
		{token.Identifier, "fetchData"},
		{token.LeftParen, "("},
		{token.RightParen, ")"},
		{token.LeftBrace, "{"},
		{token.Const, "const"},
		{token.Identifier, "data"},
		{token.Assign, "="},
		{token.Await, "await"},
		{token.Identifier, "fetch"},
		{token.LeftParen, "("},
		{token.RightParen, ")"},
		{token.Semicolon, ";"},
		{token.RightBrace, "}"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestDestructuring(t *testing.T) {
	input := `const { a, b: c, ...rest } = obj;`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Const, "const"},
		{token.LeftBrace, "{"},
		{token.Identifier, "a"},
		{token.Comma, ","},
		{token.Identifier, "b"},
		{token.Colon, ":"},
		{token.Identifier, "c"},
		{token.Comma, ","},
		{token.Spread, "..."},
		{token.Identifier, "rest"},
		{token.RightBrace, "}"},
		{token.Assign, "="},
		{token.Identifier, "obj"},
		{token.Semicolon, ";"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestImportExport(t *testing.T) {
	input := `import { foo as bar } from "module"; export default class {};`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Import, "import"},
		{token.LeftBrace, "{"},
		{token.Identifier, "foo"},
		{token.As, "as"},
		{token.Identifier, "bar"},
		{token.RightBrace, "}"},
		{token.From, "from"},
		{token.String, "module"},
		{token.Semicolon, ";"},
		{token.Export, "export"},
		{token.Default, "default"},
		{token.Class, "class"},
		{token.LeftBrace, "{"},
		{token.RightBrace, "}"},
		{token.Semicolon, ";"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestForOfLoop(t *testing.T) {
	input := `for (const x of items) {}`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.For, "for"},
		{token.LeftParen, "("},
		{token.Const, "const"},
		{token.Identifier, "x"},
		{token.Of, "of"},
		{token.Identifier, "items"},
		{token.RightParen, ")"},
		{token.LeftBrace, "{"},
		{token.RightBrace, "}"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestTernary(t *testing.T) {
	input := `a ? b : c`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Identifier, "a"},
		{token.QuestionMark, "?"},
		{token.Identifier, "b"},
		{token.Colon, ":"},
		{token.Identifier, "c"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestNullishCoalescing(t *testing.T) {
	input := `a ?? b`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Identifier, "a"},
		{token.NullishCoalesce, "??"},
		{token.Identifier, "b"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestTokenize(t *testing.T) {
	input := `let x = 42;`
	tokens := Tokenize(input)
	if len(tokens) != 6 {
		t.Fatalf("expected 6 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != token.Let {
		t.Errorf("token 0: expected Let, got %d", tokens[0].Type)
	}
	if tokens[4].Type != token.Semicolon {
		t.Errorf("token 4: expected Semicolon, got %d", tokens[4].Type)
	}
	if tokens[5].Type != token.EOF {
		t.Errorf("token 5: expected EOF, got %d", tokens[5].Type)
	}
}

func TestEmptyInput(t *testing.T) {
	l := New("")
	tok := l.NextToken()
	if tok.Type != token.EOF {
		t.Errorf("expected EOF for empty input, got %d", tok.Type)
	}
}

func TestWhitespaceOnly(t *testing.T) {
	l := New("   \t\n\r\n  ")
	tok := l.NextToken()
	if tok.Type != token.EOF {
		t.Errorf("expected EOF for whitespace-only input, got %d", tok.Type)
	}
}

func TestDotNumber(t *testing.T) {
	// .5 should be a number
	l := New(".5")
	tok := l.NextToken()
	if tok.Type != token.Number {
		t.Errorf("expected Number for '.5', got %d (lit=%q)", tok.Type, tok.Literal)
	}
	if tok.Literal != ".5" {
		t.Errorf("expected '.5', got %q", tok.Literal)
	}
}

func TestOptionalChainVsQuestionDot(t *testing.T) {
	// ?. followed by a digit should be ? and .5 (ternary + number)
	input := `x?.5`
	l := New(input)
	tok1 := l.NextToken()
	if tok1.Type != token.Identifier {
		t.Errorf("expected Identifier, got %d", tok1.Type)
	}
	tok2 := l.NextToken()
	if tok2.Type != token.QuestionMark {
		t.Errorf("expected QuestionMark (not OptionalChain before digit), got %d", tok2.Type)
	}
	tok3 := l.NextToken()
	if tok3.Type != token.Number || tok3.Literal != ".5" {
		t.Errorf("expected Number '.5', got %d %q", tok3.Type, tok3.Literal)
	}
}

func TestComplexExpression(t *testing.T) {
	input := `const result = arr.filter(x => x > 0).map(x => x ** 2);`
	tokens := Tokenize(input)

	// Just check it doesn't crash and produces reasonable number of tokens
	if len(tokens) < 15 {
		t.Errorf("expected many tokens for complex expression, got %d", len(tokens))
	}
	if tokens[len(tokens)-1].Type != token.EOF {
		t.Errorf("last token should be EOF")
	}
}

func TestUnicodeIdentifier(t *testing.T) {
	tests := []struct {
		input string
		lit   string
	}{
		{`\u0061`, "a"},           // \u0061 = 'a'
		{`\u{62}`, "b"},          // \u{62} = 'b'
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != token.Identifier {
			t.Errorf("input=%q: expected Identifier, got %d (lit=%q)", tt.input, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.lit {
			t.Errorf("input=%q: expected literal %q, got %q", tt.input, tt.lit, tok.Literal)
		}
	}
}

func TestMultiLineBlockComment(t *testing.T) {
	input := `a /*
line 2
line 3
*/ b`
	l := New(input)
	l.NextToken() // a
	tok := l.NextToken() // b
	if tok.Type != token.Identifier || tok.Literal != "b" {
		t.Errorf("expected 'b', got %d %q", tok.Type, tok.Literal)
	}
	if tok.Line != 4 {
		t.Errorf("expected line 4, got %d", tok.Line)
	}
}

func TestMultiLineTemplate(t *testing.T) {
	input := "`line1\nline2`"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.NoSubstitutionTemplate {
		t.Errorf("expected NoSubstitutionTemplate, got %d", tok.Type)
	}
	if tok.Literal != "line1\nline2" {
		t.Errorf("expected 'line1\\nline2', got %q", tok.Literal)
	}
}

func TestIllegalCharacter(t *testing.T) {
	l := New("\x01")
	tok := l.NextToken()
	if tok.Type != token.Illegal {
		t.Errorf("expected Illegal for control char, got %d", tok.Type)
	}
}

func TestYieldExpression(t *testing.T) {
	input := `function* gen() { yield 1; }`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Function, "function"},
		{token.Asterisk, "*"},
		{token.Identifier, "gen"},
		{token.LeftParen, "("},
		{token.RightParen, ")"},
		{token.LeftBrace, "{"},
		{token.Yield, "yield"},
		{token.Number, "1"},
		{token.Semicolon, ";"},
		{token.RightBrace, "}"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestSwitchStatement(t *testing.T) {
	input := `switch(x) { case 1: break; default: return; }`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Switch, "switch"},
		{token.LeftParen, "("},
		{token.Identifier, "x"},
		{token.RightParen, ")"},
		{token.LeftBrace, "{"},
		{token.Case, "case"},
		{token.Number, "1"},
		{token.Colon, ":"},
		{token.Break, "break"},
		{token.Semicolon, ";"},
		{token.Default, "default"},
		{token.Colon, ":"},
		{token.Return, "return"},
		{token.Semicolon, ";"},
		{token.RightBrace, "}"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestTryCatch(t *testing.T) {
	input := `try { throw new Error(); } catch(e) { } finally { }`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Try, "try"},
		{token.LeftBrace, "{"},
		{token.Throw, "throw"},
		{token.New, "new"},
		{token.Identifier, "Error"},
		{token.LeftParen, "("},
		{token.RightParen, ")"},
		{token.Semicolon, ";"},
		{token.RightBrace, "}"},
		{token.Catch, "catch"},
		{token.LeftParen, "("},
		{token.Identifier, "e"},
		{token.RightParen, ")"},
		{token.LeftBrace, "{"},
		{token.RightBrace, "}"},
		{token.Finally, "finally"},
		{token.LeftBrace, "{"},
		{token.RightBrace, "}"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestDeleteTypeof(t *testing.T) {
	input := `delete obj.x; typeof y;`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Delete, "delete"},
		{token.Identifier, "obj"},
		{token.Dot, "."},
		{token.Identifier, "x"},
		{token.Semicolon, ";"},
		{token.Typeof, "typeof"},
		{token.Identifier, "y"},
		{token.Semicolon, ";"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}

func TestVoidInstanceof(t *testing.T) {
	input := `void 0; x instanceof Array`
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.Void, "void"},
		{token.Number, "0"},
		{token.Semicolon, ";"},
		{token.Identifier, "x"},
		{token.Instanceof, "instanceof"},
		{token.Identifier, "Array"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. expected=%d, got=%d (lit=%q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Errorf("test[%d]: literal wrong. expected=%q, got=%q", i, exp.lit, tok.Literal)
		}
	}
}
