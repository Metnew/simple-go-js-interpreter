package ast

import "github.com/example/jsgo/token"

// Node is the interface all AST nodes implement.
type Node interface {
	TokenLiteral() string
	nodeType() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of every AST.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}
func (p *Program) nodeType() string { return "Program" }

// ---------- Statements ----------

type VariableDeclaration struct {
	Token        token.Token // var, let, or const
	Kind         string      // "var", "let", "const"
	Declarations []*VariableDeclarator
}

type VariableDeclarator struct {
	Token token.Token
	Name  Expression // Identifier or destructuring pattern
	Value Expression // may be nil
}

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

type ReturnStatement struct {
	Token token.Token
	Value Expression // may be nil
}

type IfStatement struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative Statement // may be nil; can be *IfStatement or *BlockStatement
}

type WhileStatement struct {
	Token     token.Token
	Condition Expression
	Body      Statement
}

type DoWhileStatement struct {
	Token     token.Token
	Body      Statement
	Condition Expression
}

type ForStatement struct {
	Token  token.Token
	Init   Node       // Statement or Expression, may be nil
	Test   Expression // may be nil
	Update Expression // may be nil
	Body   Statement
}

type ForInStatement struct {
	Token token.Token
	Left  Node // VariableDeclaration or Expression
	Right Expression
	Body  Statement
}

type ForOfStatement struct {
	Token token.Token
	Left  Node
	Right Expression
	Body  Statement
}

type BreakStatement struct {
	Token token.Token
	Label *Identifier // may be nil
}

type ContinueStatement struct {
	Token token.Token
	Label *Identifier // may be nil
}

type SwitchStatement struct {
	Token        token.Token
	Discriminant Expression
	Cases        []*SwitchCase
}

type SwitchCase struct {
	Token      token.Token
	Test       Expression // nil for default
	Consequent []Statement
}

type ThrowStatement struct {
	Token    token.Token
	Argument Expression
}

type TryStatement struct {
	Token   token.Token
	Block   *BlockStatement
	Handler *CatchClause // may be nil
	Finalizer *BlockStatement // may be nil
}

type CatchClause struct {
	Token token.Token
	Param Expression // may be nil (ES2019 optional catch binding)
	Body  *BlockStatement
}

type FunctionDeclaration struct {
	Token      token.Token
	Name       *Identifier
	Params     []Expression // Identifiers or patterns
	Body       *BlockStatement
	Generator  bool
	Async      bool
	Defaults   []Expression // default param values, may contain nils
	Rest       Expression   // rest parameter, may be nil
}

type ClassDeclaration struct {
	Token      token.Token
	Name       *Identifier
	SuperClass Expression // may be nil
	Body       *ClassBody
}

type ClassBody struct {
	Token   token.Token
	Methods []*MethodDefinition
}

type MethodDefinition struct {
	Token    token.Token
	Key      Expression
	Value    *FunctionExpression
	Kind     string // "constructor", "method", "get", "set"
	Static   bool
	Computed bool
}

type LabeledStatement struct {
	Token token.Token
	Label *Identifier
	Body  Statement
}

type DebuggerStatement struct {
	Token token.Token
}

type EmptyStatement struct {
	Token token.Token
}

type WithStatement struct {
	Token  token.Token
	Object Expression
	Body   Statement
}

// ---------- Expressions ----------

type Identifier struct {
	Token token.Token
	Value string
}

type NumberLiteral struct {
	Token token.Token
	Value float64
}

type StringLiteral struct {
	Token token.Token
	Value string
}

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

type NullLiteral struct {
	Token token.Token
}

type UndefinedLiteral struct {
	Token token.Token
}

type RegExpLiteral struct {
	Token   token.Token
	Pattern string
	Flags   string
}

type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression // may contain nils for elisions [1,,3]
}

type ObjectLiteral struct {
	Token      token.Token
	Properties []*Property
}

type Property struct {
	Token     token.Token
	Key       Expression
	Value     Expression
	Kind      string // "init", "get", "set"
	Shorthand bool
	Computed  bool
	Method    bool
}

type FunctionExpression struct {
	Token     token.Token
	Name      *Identifier // may be nil for anonymous
	Params    []Expression
	Body      *BlockStatement
	Generator bool
	Async     bool
	Defaults  []Expression
	Rest      Expression
}

type ArrowFunctionExpression struct {
	Token  token.Token
	Params []Expression
	Body   Node // BlockStatement or Expression
	Async  bool
	Defaults []Expression
	Rest     Expression
}

type UnaryExpression struct {
	Token    token.Token
	Operator string
	Operand  Expression
	Prefix   bool
}

type UpdateExpression struct {
	Token    token.Token
	Operator string // ++ or --
	Operand  Expression
	Prefix   bool
}

type BinaryExpression struct {
	Token    token.Token
	Operator string
	Left     Expression
	Right    Expression
}

type LogicalExpression struct {
	Token    token.Token
	Operator string // && or ||
	Left     Expression
	Right    Expression
}

type AssignmentExpression struct {
	Token    token.Token
	Operator string
	Left     Expression
	Right    Expression
}

type ConditionalExpression struct {
	Token       token.Token
	Test        Expression
	Consequent  Expression
	Alternate   Expression
}

type CallExpression struct {
	Token     token.Token
	Callee    Expression
	Arguments []Expression
}

type MemberExpression struct {
	Token    token.Token
	Object   Expression
	Property Expression
	Computed bool
}

type NewExpression struct {
	Token     token.Token
	Callee    Expression
	Arguments []Expression
}

type SequenceExpression struct {
	Token       token.Token
	Expressions []Expression
}

type TemplateLiteralExpr struct {
	Token       token.Token
	Quasis      []*TemplateElement
	Expressions []Expression
}

type TemplateElement struct {
	Token  token.Token
	Value  string
	Tail   bool
}

type TaggedTemplateExpression struct {
	Token    token.Token
	Tag      Expression
	Quasi    *TemplateLiteralExpr
}

type SpreadElement struct {
	Token    token.Token
	Argument Expression
}

type YieldExpression struct {
	Token    token.Token
	Argument Expression // may be nil
	Delegate bool       // yield*
}

type AwaitExpression struct {
	Token    token.Token
	Argument Expression
}

type ClassExpression struct {
	Token      token.Token
	Name       *Identifier // may be nil
	SuperClass Expression
	Body       *ClassBody
}

type ThisExpression struct {
	Token token.Token
}

type SuperExpression struct {
	Token token.Token
}

// Destructuring patterns
type ObjectPattern struct {
	Token      token.Token
	Properties []*Property
}

type ArrayPattern struct {
	Token    token.Token
	Elements []Expression // may contain nils for holes
}

type AssignmentPattern struct {
	Token token.Token
	Left  Expression
	Right Expression
}

type RestElement struct {
	Token    token.Token
	Argument Expression
}

type ComputedPropertyName struct {
	Token      token.Token
	Expression Expression
}

// --- Node interface implementations ---
// Statement markers
func (s *VariableDeclaration) statementNode()  {}
func (s *ExpressionStatement) statementNode()   {}
func (s *BlockStatement) statementNode()        {}
func (s *ReturnStatement) statementNode()       {}
func (s *IfStatement) statementNode()           {}
func (s *WhileStatement) statementNode()        {}
func (s *DoWhileStatement) statementNode()      {}
func (s *ForStatement) statementNode()          {}
func (s *ForInStatement) statementNode()        {}
func (s *ForOfStatement) statementNode()        {}
func (s *BreakStatement) statementNode()        {}
func (s *ContinueStatement) statementNode()     {}
func (s *SwitchStatement) statementNode()       {}
func (s *ThrowStatement) statementNode()        {}
func (s *TryStatement) statementNode()          {}
func (s *FunctionDeclaration) statementNode()   {}
func (s *ClassDeclaration) statementNode()      {}
func (s *LabeledStatement) statementNode()      {}
func (s *DebuggerStatement) statementNode()     {}
func (s *EmptyStatement) statementNode()        {}
func (s *WithStatement) statementNode()         {}

// Expression markers
func (e *Identifier) expressionNode()                {}
func (e *NumberLiteral) expressionNode()              {}
func (e *StringLiteral) expressionNode()              {}
func (e *BooleanLiteral) expressionNode()             {}
func (e *NullLiteral) expressionNode()                {}
func (e *UndefinedLiteral) expressionNode()           {}
func (e *RegExpLiteral) expressionNode()               {}
func (e *ArrayLiteral) expressionNode()               {}
func (e *ObjectLiteral) expressionNode()              {}
func (e *FunctionExpression) expressionNode()         {}
func (e *ArrowFunctionExpression) expressionNode()    {}
func (e *UnaryExpression) expressionNode()            {}
func (e *UpdateExpression) expressionNode()           {}
func (e *BinaryExpression) expressionNode()           {}
func (e *LogicalExpression) expressionNode()          {}
func (e *AssignmentExpression) expressionNode()       {}
func (e *ConditionalExpression) expressionNode()      {}
func (e *CallExpression) expressionNode()             {}
func (e *MemberExpression) expressionNode()           {}
func (e *NewExpression) expressionNode()              {}
func (e *SequenceExpression) expressionNode()         {}
func (e *TemplateLiteralExpr) expressionNode()        {}
func (e *TaggedTemplateExpression) expressionNode()   {}
func (e *SpreadElement) expressionNode()              {}
func (e *YieldExpression) expressionNode()            {}
func (e *AwaitExpression) expressionNode()            {}
func (e *ClassExpression) expressionNode()            {}
func (e *ThisExpression) expressionNode()             {}
func (e *SuperExpression) expressionNode()            {}
func (e *ObjectPattern) expressionNode()              {}
func (e *ArrayPattern) expressionNode()               {}
func (e *AssignmentPattern) expressionNode()          {}
func (e *RestElement) expressionNode()                {}
func (e *ComputedPropertyName) expressionNode()       {}
func (e *VariableDeclarator) expressionNode()         {}

// TokenLiteral implementations
func (s *VariableDeclaration) TokenLiteral() string  { return s.Token.Literal }
func (s *VariableDeclarator) TokenLiteral() string   { return s.Token.Literal }
func (s *ExpressionStatement) TokenLiteral() string   { return s.Token.Literal }
func (s *BlockStatement) TokenLiteral() string        { return s.Token.Literal }
func (s *ReturnStatement) TokenLiteral() string       { return s.Token.Literal }
func (s *IfStatement) TokenLiteral() string           { return s.Token.Literal }
func (s *WhileStatement) TokenLiteral() string        { return s.Token.Literal }
func (s *DoWhileStatement) TokenLiteral() string      { return s.Token.Literal }
func (s *ForStatement) TokenLiteral() string          { return s.Token.Literal }
func (s *ForInStatement) TokenLiteral() string        { return s.Token.Literal }
func (s *ForOfStatement) TokenLiteral() string        { return s.Token.Literal }
func (s *BreakStatement) TokenLiteral() string        { return s.Token.Literal }
func (s *ContinueStatement) TokenLiteral() string     { return s.Token.Literal }
func (s *SwitchStatement) TokenLiteral() string       { return s.Token.Literal }
func (s *ThrowStatement) TokenLiteral() string        { return s.Token.Literal }
func (s *TryStatement) TokenLiteral() string          { return s.Token.Literal }
func (s *CatchClause) TokenLiteral() string           { return s.Token.Literal }
func (s *FunctionDeclaration) TokenLiteral() string   { return s.Token.Literal }
func (s *ClassDeclaration) TokenLiteral() string      { return s.Token.Literal }
func (s *ClassBody) TokenLiteral() string             { return s.Token.Literal }
func (s *MethodDefinition) TokenLiteral() string      { return s.Token.Literal }
func (s *LabeledStatement) TokenLiteral() string      { return s.Token.Literal }
func (s *DebuggerStatement) TokenLiteral() string     { return s.Token.Literal }
func (s *EmptyStatement) TokenLiteral() string        { return s.Token.Literal }
func (s *WithStatement) TokenLiteral() string         { return s.Token.Literal }

func (e *Identifier) TokenLiteral() string                { return e.Token.Literal }
func (e *NumberLiteral) TokenLiteral() string              { return e.Token.Literal }
func (e *StringLiteral) TokenLiteral() string              { return e.Token.Literal }
func (e *BooleanLiteral) TokenLiteral() string             { return e.Token.Literal }
func (e *NullLiteral) TokenLiteral() string                { return e.Token.Literal }
func (e *UndefinedLiteral) TokenLiteral() string           { return e.Token.Literal }
func (e *RegExpLiteral) TokenLiteral() string              { return e.Token.Literal }
func (e *ArrayLiteral) TokenLiteral() string               { return e.Token.Literal }
func (e *ObjectLiteral) TokenLiteral() string              { return e.Token.Literal }
func (e *Property) TokenLiteral() string                   { return e.Token.Literal }
func (e *FunctionExpression) TokenLiteral() string         { return e.Token.Literal }
func (e *ArrowFunctionExpression) TokenLiteral() string    { return e.Token.Literal }
func (e *UnaryExpression) TokenLiteral() string            { return e.Token.Literal }
func (e *UpdateExpression) TokenLiteral() string           { return e.Token.Literal }
func (e *BinaryExpression) TokenLiteral() string           { return e.Token.Literal }
func (e *LogicalExpression) TokenLiteral() string          { return e.Token.Literal }
func (e *AssignmentExpression) TokenLiteral() string       { return e.Token.Literal }
func (e *ConditionalExpression) TokenLiteral() string      { return e.Token.Literal }
func (e *CallExpression) TokenLiteral() string             { return e.Token.Literal }
func (e *MemberExpression) TokenLiteral() string           { return e.Token.Literal }
func (e *NewExpression) TokenLiteral() string              { return e.Token.Literal }
func (e *SequenceExpression) TokenLiteral() string         { return e.Token.Literal }
func (e *TemplateLiteralExpr) TokenLiteral() string        { return e.Token.Literal }
func (e *TemplateElement) TokenLiteral() string            { return e.Token.Literal }
func (e *TaggedTemplateExpression) TokenLiteral() string   { return e.Token.Literal }
func (e *SpreadElement) TokenLiteral() string              { return e.Token.Literal }
func (e *YieldExpression) TokenLiteral() string            { return e.Token.Literal }
func (e *AwaitExpression) TokenLiteral() string            { return e.Token.Literal }
func (e *ClassExpression) TokenLiteral() string            { return e.Token.Literal }
func (e *ThisExpression) TokenLiteral() string             { return e.Token.Literal }
func (e *SuperExpression) TokenLiteral() string            { return e.Token.Literal }
func (e *ObjectPattern) TokenLiteral() string              { return e.Token.Literal }
func (e *ArrayPattern) TokenLiteral() string               { return e.Token.Literal }
func (e *AssignmentPattern) TokenLiteral() string          { return e.Token.Literal }
func (e *RestElement) TokenLiteral() string                { return e.Token.Literal }
func (e *ComputedPropertyName) TokenLiteral() string       { return e.Token.Literal }
func (e *SwitchCase) TokenLiteral() string                 { return e.Token.Literal }

// nodeType implementations
func (s *VariableDeclaration) nodeType() string  { return "VariableDeclaration" }
func (s *VariableDeclarator) nodeType() string   { return "VariableDeclarator" }
func (s *ExpressionStatement) nodeType() string   { return "ExpressionStatement" }
func (s *BlockStatement) nodeType() string        { return "BlockStatement" }
func (s *ReturnStatement) nodeType() string       { return "ReturnStatement" }
func (s *IfStatement) nodeType() string           { return "IfStatement" }
func (s *WhileStatement) nodeType() string        { return "WhileStatement" }
func (s *DoWhileStatement) nodeType() string      { return "DoWhileStatement" }
func (s *ForStatement) nodeType() string          { return "ForStatement" }
func (s *ForInStatement) nodeType() string        { return "ForInStatement" }
func (s *ForOfStatement) nodeType() string        { return "ForOfStatement" }
func (s *BreakStatement) nodeType() string        { return "BreakStatement" }
func (s *ContinueStatement) nodeType() string     { return "ContinueStatement" }
func (s *SwitchStatement) nodeType() string       { return "SwitchStatement" }
func (s *ThrowStatement) nodeType() string        { return "ThrowStatement" }
func (s *TryStatement) nodeType() string          { return "TryStatement" }
func (s *CatchClause) nodeType() string           { return "CatchClause" }
func (s *FunctionDeclaration) nodeType() string   { return "FunctionDeclaration" }
func (s *ClassDeclaration) nodeType() string      { return "ClassDeclaration" }
func (s *ClassBody) nodeType() string             { return "ClassBody" }
func (s *MethodDefinition) nodeType() string      { return "MethodDefinition" }
func (s *LabeledStatement) nodeType() string      { return "LabeledStatement" }
func (s *DebuggerStatement) nodeType() string     { return "DebuggerStatement" }
func (s *EmptyStatement) nodeType() string        { return "EmptyStatement" }
func (s *WithStatement) nodeType() string         { return "WithStatement" }
func (s *SwitchCase) nodeType() string            { return "SwitchCase" }

func (e *Identifier) nodeType() string                { return "Identifier" }
func (e *NumberLiteral) nodeType() string              { return "NumberLiteral" }
func (e *StringLiteral) nodeType() string              { return "StringLiteral" }
func (e *BooleanLiteral) nodeType() string             { return "BooleanLiteral" }
func (e *NullLiteral) nodeType() string                { return "NullLiteral" }
func (e *UndefinedLiteral) nodeType() string           { return "UndefinedLiteral" }
func (e *RegExpLiteral) nodeType() string              { return "RegExpLiteral" }
func (e *ArrayLiteral) nodeType() string               { return "ArrayLiteral" }
func (e *ObjectLiteral) nodeType() string              { return "ObjectLiteral" }
func (e *Property) nodeType() string                   { return "Property" }
func (e *FunctionExpression) nodeType() string         { return "FunctionExpression" }
func (e *ArrowFunctionExpression) nodeType() string    { return "ArrowFunctionExpression" }
func (e *UnaryExpression) nodeType() string            { return "UnaryExpression" }
func (e *UpdateExpression) nodeType() string           { return "UpdateExpression" }
func (e *BinaryExpression) nodeType() string           { return "BinaryExpression" }
func (e *LogicalExpression) nodeType() string          { return "LogicalExpression" }
func (e *AssignmentExpression) nodeType() string       { return "AssignmentExpression" }
func (e *ConditionalExpression) nodeType() string      { return "ConditionalExpression" }
func (e *CallExpression) nodeType() string             { return "CallExpression" }
func (e *MemberExpression) nodeType() string           { return "MemberExpression" }
func (e *NewExpression) nodeType() string              { return "NewExpression" }
func (e *SequenceExpression) nodeType() string         { return "SequenceExpression" }
func (e *TemplateLiteralExpr) nodeType() string        { return "TemplateLiteralExpr" }
func (e *TemplateElement) nodeType() string            { return "TemplateElement" }
func (e *TaggedTemplateExpression) nodeType() string   { return "TaggedTemplateExpression" }
func (e *SpreadElement) nodeType() string              { return "SpreadElement" }
func (e *YieldExpression) nodeType() string            { return "YieldExpression" }
func (e *AwaitExpression) nodeType() string            { return "AwaitExpression" }
func (e *ClassExpression) nodeType() string            { return "ClassExpression" }
func (e *ThisExpression) nodeType() string             { return "ThisExpression" }
func (e *SuperExpression) nodeType() string            { return "SuperExpression" }
func (e *ObjectPattern) nodeType() string              { return "ObjectPattern" }
func (e *ArrayPattern) nodeType() string               { return "ArrayPattern" }
func (e *AssignmentPattern) nodeType() string          { return "AssignmentPattern" }
func (e *RestElement) nodeType() string                { return "RestElement" }
func (e *ComputedPropertyName) nodeType() string       { return "ComputedPropertyName" }
