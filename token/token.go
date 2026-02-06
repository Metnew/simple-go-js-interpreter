package token

type TokenType int

const (
	// Literals
	Illegal TokenType = iota
	EOF
	Identifier
	Number
	String
	TemplateLiteral
	RegExp

	// Operators
	Plus
	Minus
	Asterisk
	Slash
	Percent
	Exponent // **
	Assign
	PlusAssign
	MinusAssign
	AsteriskAssign
	SlashAssign
	PercentAssign
	ExponentAssign
	AmpersandAssign
	PipeAssign
	CaretAssign
	LeftShiftAssign
	RightShiftAssign
	UnsignedRightShiftAssign
	NullishAssign    // ??=
	AndAssign        // &&=
	OrAssign         // ||=
	Equal
	NotEqual
	StrictEqual
	StrictNotEqual
	LessThan
	GreaterThan
	LessThanOrEqual
	GreaterThanOrEqual
	And
	Or
	Not
	BitwiseAnd
	BitwiseOr
	BitwiseXor
	BitwiseNot
	LeftShift
	RightShift
	UnsignedRightShift
	Increment
	Decrement

	// Delimiters
	LeftParen
	RightParen
	LeftBrace
	RightBrace
	LeftBracket
	RightBracket
	Semicolon
	Colon
	Comma
	Dot
	Spread // ...
	Arrow  // =>
	QuestionMark
	OptionalChain   // ?.
	NullishCoalesce // ??

	// Keywords
	Var
	Let
	Const
	Function
	Return
	If
	Else
	While
	For
	Do
	Break
	Continue
	Switch
	Case
	Default
	Throw
	Try
	Catch
	Finally
	New
	Delete
	Typeof
	Void
	In
	Instanceof
	This
	Class
	Extends
	Super
	Import
	Export
	From
	As
	Of
	Yield
	Async
	Await
	True
	False
	Null
	Undefined
	Debugger
	With

	// Template literal parts
	TemplateHead
	TemplateMiddle
	TemplateTail
	NoSubstitutionTemplate
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

var Keywords = map[string]TokenType{
	"var":        Var,
	"let":        Let,
	"const":      Const,
	"function":   Function,
	"return":     Return,
	"if":         If,
	"else":       Else,
	"while":      While,
	"for":        For,
	"do":         Do,
	"break":      Break,
	"continue":   Continue,
	"switch":     Switch,
	"case":       Case,
	"default":    Default,
	"throw":      Throw,
	"try":        Try,
	"catch":      Catch,
	"finally":    Finally,
	"new":        New,
	"delete":     Delete,
	"typeof":     Typeof,
	"void":       Void,
	"in":         In,
	"instanceof": Instanceof,
	"this":       This,
	"class":      Class,
	"extends":    Extends,
	"super":      Super,
	"import":     Import,
	"export":     Export,
	"from":       From,
	"as":         As,
	"of":         Of,
	"yield":      Yield,
	"async":      Async,
	"await":      Await,
	"true":       True,
	"false":      False,
	"null":       Null,
	"undefined":  Undefined,
	"debugger":   Debugger,
	"with":       With,
}

func LookupIdentifier(ident string) TokenType {
	if tok, ok := Keywords[ident]; ok {
		return tok
	}
	return Identifier
}
