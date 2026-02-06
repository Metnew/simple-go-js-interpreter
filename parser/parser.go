package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/example/jsgo/ast"
	"github.com/example/jsgo/lexer"
	"github.com/example/jsgo/token"
)

// Precedence levels for Pratt parsing
const (
	_ int = iota
	precComma
	precAssignment
	precConditional
	precNullishCoalesce
	precLogicalOr
	precLogicalAnd
	precBitwiseOr
	precBitwiseXor
	precBitwiseAnd
	precEquality
	precRelational
	precShift
	precAdditive
	precMultiplicative
	precExponent
	precUnary
	precPostfix
	precCall
	precMember
)

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	prevType  token.TokenType
	prevLine  int
	errors    []error
	noIn      bool // suppress 'in' as binary operator (for-in disambiguation)
}

func New(source string) *Parser {
	p := &Parser{
		l:        lexer.New(source),
		prevType: token.EOF,
	}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) ParseProgram() (*ast.Program, []error) {
	program := &ast.Program{}
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}
	return program, p.errors
}

func (p *Parser) nextToken() {
	p.prevType = p.curToken.Type
	p.prevLine = p.curToken.Line
	p.curToken = p.peekToken
	p.peekToken = p.l.NextTokenWithRegex(p.curToken.Type)
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expect(t token.TokenType) bool {
	if p.curTokenIs(t) {
		p.nextToken()
		return true
	}
	p.addError("expected %s, got %s (%q)", tokenName(t), tokenName(p.curToken.Type), p.curToken.Literal)
	return false
}

func (p *Parser) addError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	err := fmt.Errorf("parse error at %d:%d: %s", p.curToken.Line, p.curToken.Column, msg)
	p.errors = append(p.errors, err)
}

// parseStatement dispatches to the appropriate statement parser.
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.Var, token.Let, token.Const:
		return p.parseVariableDeclaration()
	case token.LeftBrace:
		return p.parseBlockStatement()
	case token.Return:
		return p.parseReturnStatement()
	case token.If:
		return p.parseIfStatement()
	case token.While:
		return p.parseWhileStatement()
	case token.Do:
		return p.parseDoWhileStatement()
	case token.For:
		return p.parseForStatement()
	case token.Break:
		return p.parseBreakStatement()
	case token.Continue:
		return p.parseContinueStatement()
	case token.Switch:
		return p.parseSwitchStatement()
	case token.Throw:
		return p.parseThrowStatement()
	case token.Try:
		return p.parseTryStatement()
	case token.Function:
		return p.parseFunctionDeclaration()
	case token.Class:
		return p.parseClassDeclaration()
	case token.Debugger:
		return p.parseDebuggerStatement()
	case token.Semicolon:
		return p.parseEmptyStatement()
	case token.With:
		return p.parseWithStatement()
	case token.Async:
		if p.peekTokenIs(token.Function) {
			return p.parseAsyncFunctionDeclaration()
		}
		return p.parseExpressionOrLabeledStatement()
	default:
		return p.parseExpressionOrLabeledStatement()
	}
}

func (p *Parser) parseExpressionOrLabeledStatement() ast.Statement {
	if p.curTokenIs(token.Identifier) && p.peekTokenIs(token.Colon) {
		return p.parseLabeledStatement()
	}
	return p.parseExpressionStatement()
}

// ---------- Statement Parsers ----------

func (p *Parser) parseVariableDeclaration() *ast.VariableDeclaration {
	stmt := &ast.VariableDeclaration{Token: p.curToken, Kind: p.curToken.Literal}
	p.nextToken() // consume var/let/const

	for {
		decl := p.parseVariableDeclarator()
		stmt.Declarations = append(stmt.Declarations, decl)
		if !p.curTokenIs(token.Comma) {
			break
		}
		p.nextToken() // consume comma
	}

	p.consumeSemicolon()
	return stmt
}

func (p *Parser) parseVariableDeclarator() *ast.VariableDeclarator {
	decl := &ast.VariableDeclarator{Token: p.curToken}
	decl.Name = p.parseBindingPattern()

	if p.curTokenIs(token.Assign) {
		p.nextToken() // consume =
		decl.Value = p.parseAssignmentExpression()
	}
	return decl
}

func (p *Parser) parseBindingPattern() ast.Expression {
	switch p.curToken.Type {
	case token.LeftBrace:
		return p.parseObjectPattern()
	case token.LeftBracket:
		return p.parseArrayPattern()
	default:
		return p.parseIdentifier()
	}
}

func (p *Parser) parseObjectPattern() *ast.ObjectPattern {
	pat := &ast.ObjectPattern{Token: p.curToken}
	p.nextToken() // consume {

	for !p.curTokenIs(token.RightBrace) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.Spread) {
			rest := &ast.RestElement{Token: p.curToken}
			p.nextToken() // consume ...
			rest.Argument = p.parseBindingPattern()
			prop := &ast.Property{
				Token: rest.Token,
				Key:   rest,
				Value: rest,
			}
			pat.Properties = append(pat.Properties, prop)
		} else {
			prop := p.parseBindingProperty()
			pat.Properties = append(pat.Properties, prop)
		}
		if !p.curTokenIs(token.Comma) {
			break
		}
		p.nextToken() // consume comma
	}
	p.expect(token.RightBrace)
	return pat
}

func (p *Parser) parseBindingProperty() *ast.Property {
	prop := &ast.Property{Token: p.curToken, Kind: "init"}

	if p.curTokenIs(token.LeftBracket) {
		prop.Computed = true
		p.nextToken() // consume [
		prop.Key = p.parseAssignmentExpression()
		p.expect(token.RightBracket)
		p.expect(token.Colon)
		prop.Value = p.parseBindingElement()
		return prop
	}

	prop.Key = p.parsePropertyName()

	if p.curTokenIs(token.Colon) {
		prop.Shorthand = false
		p.nextToken() // consume :
		prop.Value = p.parseBindingElement()
	} else {
		prop.Shorthand = true
		prop.Value = prop.Key
		if p.curTokenIs(token.Assign) {
			p.nextToken() // consume =
			prop.Value = &ast.AssignmentPattern{
				Token: prop.Token,
				Left:  prop.Key,
				Right: p.parseAssignmentExpression(),
			}
		}
	}
	return prop
}

func (p *Parser) parseBindingElement() ast.Expression {
	elem := p.parseBindingPattern()
	if p.curTokenIs(token.Assign) {
		tok := p.curToken
		p.nextToken()
		return &ast.AssignmentPattern{Token: tok, Left: elem, Right: p.parseAssignmentExpression()}
	}
	return elem
}

func (p *Parser) parseArrayPattern() *ast.ArrayPattern {
	pat := &ast.ArrayPattern{Token: p.curToken}
	p.nextToken() // consume [

	for !p.curTokenIs(token.RightBracket) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.Comma) {
			pat.Elements = append(pat.Elements, nil)
			p.nextToken()
			continue
		}
		if p.curTokenIs(token.Spread) {
			rest := &ast.RestElement{Token: p.curToken}
			p.nextToken()
			rest.Argument = p.parseBindingPattern()
			pat.Elements = append(pat.Elements, rest)
			break
		}
		elem := p.parseBindingElement()
		pat.Elements = append(pat.Elements, elem)
		if !p.curTokenIs(token.Comma) {
			break
		}
		p.nextToken() // consume comma
	}
	p.expect(token.RightBracket)
	return pat
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	p.nextToken() // consume {

	for !p.curTokenIs(token.RightBrace) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
	}
	p.expect(token.RightBrace)
	return block
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken() // consume return

	if !p.curTokenIs(token.Semicolon) && !p.curTokenIs(token.RightBrace) && !p.curTokenIs(token.EOF) {
		stmt.Value = p.parseExpression(precComma)
	}
	p.consumeSemicolon()
	return stmt
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}
	p.nextToken() // consume if
	p.expect(token.LeftParen)
	stmt.Condition = p.parseExpression(precComma)
	p.expect(token.RightParen)

	if p.curTokenIs(token.LeftBrace) {
		stmt.Consequence = p.parseBlockStatement()
	} else {
		inner := p.parseStatement()
		stmt.Consequence = &ast.BlockStatement{
			Token:      p.curToken,
			Statements: []ast.Statement{inner},
		}
	}

	if p.curTokenIs(token.Else) {
		p.nextToken()
		if p.curTokenIs(token.If) {
			stmt.Alternative = p.parseIfStatement()
		} else if p.curTokenIs(token.LeftBrace) {
			stmt.Alternative = p.parseBlockStatement()
		} else {
			inner := p.parseStatement()
			stmt.Alternative = &ast.BlockStatement{
				Token:      p.curToken,
				Statements: []ast.Statement{inner},
			}
		}
	}
	return stmt
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken}
	p.nextToken() // consume while
	p.expect(token.LeftParen)
	stmt.Condition = p.parseExpression(precComma)
	p.expect(token.RightParen)
	stmt.Body = p.parseStatement()
	return stmt
}

func (p *Parser) parseDoWhileStatement() *ast.DoWhileStatement {
	stmt := &ast.DoWhileStatement{Token: p.curToken}
	p.nextToken() // consume do
	stmt.Body = p.parseStatement()
	p.expect(token.While)
	p.expect(token.LeftParen)
	stmt.Condition = p.parseExpression(precComma)
	p.expect(token.RightParen)
	p.consumeSemicolon()
	return stmt
}

func (p *Parser) parseForStatement() ast.Statement {
	tok := p.curToken
	p.nextToken() // consume for
	p.expect(token.LeftParen)

	// for (var/let/const ...
	if p.curTokenIs(token.Var) || p.curTokenIs(token.Let) || p.curTokenIs(token.Const) {
		return p.parseForWithDeclaration(tok)
	}

	// for (; ...
	if p.curTokenIs(token.Semicolon) {
		p.nextToken()
		return p.parseForStandard(tok, nil)
	}

	// for (expr ...
	// Parse with noIn to prevent 'in' from being consumed as binary operator
	p.noIn = true
	expr := p.parseExpression(0)
	p.noIn = false

	if p.curTokenIs(token.In) {
		p.nextToken()
		right := p.parseExpression(precComma)
		p.expect(token.RightParen)
		body := p.parseStatement()
		return &ast.ForInStatement{Token: tok, Left: expr, Right: right, Body: body}
	}
	if p.curTokenIs(token.Of) {
		p.nextToken()
		right := p.parseAssignmentExpression()
		p.expect(token.RightParen)
		body := p.parseStatement()
		return &ast.ForOfStatement{Token: tok, Left: expr, Right: right, Body: body}
	}

	p.expect(token.Semicolon)
	return p.parseForStandard(tok, &ast.ExpressionStatement{Token: tok, Expression: expr})
}

func (p *Parser) parseForWithDeclaration(tok token.Token) ast.Statement {
	declToken := p.curToken
	kind := p.curToken.Literal
	p.nextToken() // consume var/let/const

	decl := &ast.VariableDeclaration{Token: declToken, Kind: kind}
	d := &ast.VariableDeclarator{Token: p.curToken}
	d.Name = p.parseBindingPattern()

	// for-in / for-of
	if p.curTokenIs(token.In) {
		decl.Declarations = append(decl.Declarations, d)
		p.nextToken()
		right := p.parseExpression(precComma)
		p.expect(token.RightParen)
		body := p.parseStatement()
		return &ast.ForInStatement{Token: tok, Left: decl, Right: right, Body: body}
	}
	if p.curTokenIs(token.Of) {
		decl.Declarations = append(decl.Declarations, d)
		p.nextToken()
		right := p.parseAssignmentExpression()
		p.expect(token.RightParen)
		body := p.parseStatement()
		return &ast.ForOfStatement{Token: tok, Left: decl, Right: right, Body: body}
	}

	// standard for with initializer
	if p.curTokenIs(token.Assign) {
		p.nextToken()
		d.Value = p.parseAssignmentExpression()
	}
	decl.Declarations = append(decl.Declarations, d)

	for p.curTokenIs(token.Comma) {
		p.nextToken()
		d2 := p.parseVariableDeclarator()
		decl.Declarations = append(decl.Declarations, d2)
	}

	p.expect(token.Semicolon)
	return p.parseForStandard(tok, decl)
}

func (p *Parser) parseForStandard(tok token.Token, init ast.Node) *ast.ForStatement {
	stmt := &ast.ForStatement{Token: tok, Init: init}

	if !p.curTokenIs(token.Semicolon) {
		stmt.Test = p.parseExpression(precComma)
	}
	p.expect(token.Semicolon)

	if !p.curTokenIs(token.RightParen) {
		stmt.Update = p.parseExpression(precComma)
	}
	p.expect(token.RightParen)
	stmt.Body = p.parseStatement()
	return stmt
}

func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	stmt := &ast.BreakStatement{Token: p.curToken}
	p.nextToken() // consume break
	if p.curTokenIs(token.Identifier) && !p.prevTokenWasNewline() {
		stmt.Label = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
	}
	p.consumeSemicolon()
	return stmt
}

func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	stmt := &ast.ContinueStatement{Token: p.curToken}
	p.nextToken() // consume continue
	if p.curTokenIs(token.Identifier) && !p.prevTokenWasNewline() {
		stmt.Label = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
	}
	p.consumeSemicolon()
	return stmt
}

func (p *Parser) parseSwitchStatement() *ast.SwitchStatement {
	stmt := &ast.SwitchStatement{Token: p.curToken}
	p.nextToken() // consume switch
	p.expect(token.LeftParen)
	stmt.Discriminant = p.parseExpression(precComma)
	p.expect(token.RightParen)
	p.expect(token.LeftBrace)

	for !p.curTokenIs(token.RightBrace) && !p.curTokenIs(token.EOF) {
		sc := &ast.SwitchCase{Token: p.curToken}
		if p.curTokenIs(token.Case) {
			p.nextToken()
			sc.Test = p.parseExpression(precComma)
		} else if p.curTokenIs(token.Default) {
			p.nextToken()
		} else {
			p.addError("expected case or default")
			p.nextToken()
			continue
		}
		p.expect(token.Colon)
		for !p.curTokenIs(token.Case) && !p.curTokenIs(token.Default) && !p.curTokenIs(token.RightBrace) && !p.curTokenIs(token.EOF) {
			s := p.parseStatement()
			if s != nil {
				sc.Consequent = append(sc.Consequent, s)
			}
		}
		stmt.Cases = append(stmt.Cases, sc)
	}
	p.expect(token.RightBrace)
	return stmt
}

func (p *Parser) parseThrowStatement() *ast.ThrowStatement {
	stmt := &ast.ThrowStatement{Token: p.curToken}
	p.nextToken() // consume throw
	stmt.Argument = p.parseExpression(precComma)
	p.consumeSemicolon()
	return stmt
}

func (p *Parser) parseTryStatement() *ast.TryStatement {
	stmt := &ast.TryStatement{Token: p.curToken}
	p.nextToken() // consume try
	stmt.Block = p.parseBlockStatement()

	if p.curTokenIs(token.Catch) {
		stmt.Handler = &ast.CatchClause{Token: p.curToken}
		p.nextToken() // consume catch
		if p.curTokenIs(token.LeftParen) {
			p.nextToken()
			stmt.Handler.Param = p.parseBindingPattern()
			p.expect(token.RightParen)
		}
		stmt.Handler.Body = p.parseBlockStatement()
	}
	if p.curTokenIs(token.Finally) {
		p.nextToken()
		stmt.Finalizer = p.parseBlockStatement()
	}
	return stmt
}

func (p *Parser) parseFunctionDeclaration() *ast.FunctionDeclaration {
	decl := &ast.FunctionDeclaration{Token: p.curToken}
	p.nextToken() // consume function

	if p.curTokenIs(token.Asterisk) {
		decl.Generator = true
		p.nextToken()
	}

	decl.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	p.parseFunctionParams(decl)
	decl.Body = p.parseBlockStatement()
	return decl
}

func (p *Parser) parseAsyncFunctionDeclaration() *ast.FunctionDeclaration {
	decl := &ast.FunctionDeclaration{Token: p.curToken, Async: true}
	p.nextToken() // consume async
	p.nextToken() // consume function

	if p.curTokenIs(token.Asterisk) {
		decl.Generator = true
		p.nextToken()
	}

	decl.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	p.parseFunctionParams(decl)
	decl.Body = p.parseBlockStatement()
	return decl
}

type funcParamTarget interface {
	setParams([]ast.Expression)
	setDefaults([]ast.Expression)
	setRest(ast.Expression)
}

type funcDeclTarget struct{ d *ast.FunctionDeclaration }

func (t funcDeclTarget) setParams(p []ast.Expression)   { t.d.Params = p }
func (t funcDeclTarget) setDefaults(d []ast.Expression)  { t.d.Defaults = d }
func (t funcDeclTarget) setRest(r ast.Expression)        { t.d.Rest = r }

type funcExprTarget struct{ e *ast.FunctionExpression }

func (t funcExprTarget) setParams(p []ast.Expression)   { t.e.Params = p }
func (t funcExprTarget) setDefaults(d []ast.Expression)  { t.e.Defaults = d }
func (t funcExprTarget) setRest(r ast.Expression)        { t.e.Rest = r }

func (p *Parser) parseFunctionParams(decl *ast.FunctionDeclaration) {
	target := funcDeclTarget{decl}
	p.parseFunctionParamsGeneric(target)
}

func (p *Parser) parseFunctionParamsGeneric(target funcParamTarget) {
	p.expect(token.LeftParen)
	var params []ast.Expression
	var defaults []ast.Expression
	hasDefaults := false

	for !p.curTokenIs(token.RightParen) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.Spread) {
			restTok := p.curToken
			p.nextToken()
			rest := p.parseBindingPattern()
			target.setRest(&ast.RestElement{Token: restTok, Argument: rest})
			if p.curTokenIs(token.Comma) {
				p.nextToken()
			}
			break
		}

		param := p.parseBindingPattern()
		params = append(params, param)

		if p.curTokenIs(token.Assign) {
			hasDefaults = true
			p.nextToken()
			defaults = append(defaults, p.parseAssignmentExpression())
		} else {
			defaults = append(defaults, nil)
		}

		if !p.curTokenIs(token.Comma) {
			break
		}
		p.nextToken()
	}
	p.expect(token.RightParen)

	target.setParams(params)
	if hasDefaults {
		target.setDefaults(defaults)
	}
}

func (p *Parser) parseClassDeclaration() *ast.ClassDeclaration {
	decl := &ast.ClassDeclaration{Token: p.curToken}
	p.nextToken() // consume class

	if p.curTokenIs(token.Identifier) {
		decl.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
	}

	if p.curTokenIs(token.Extends) {
		p.nextToken()
		decl.SuperClass = p.parseLeftHandSideExpression()
	}

	decl.Body = p.parseClassBody()
	return decl
}

func (p *Parser) parseClassBody() *ast.ClassBody {
	body := &ast.ClassBody{Token: p.curToken}
	p.expect(token.LeftBrace)

	for !p.curTokenIs(token.RightBrace) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.Semicolon) {
			p.nextToken()
			continue
		}
		method := p.parseMethodDefinition()
		body.Methods = append(body.Methods, method)
	}
	p.expect(token.RightBrace)
	return body
}

func (p *Parser) parseMethodDefinition() *ast.MethodDefinition {
	md := &ast.MethodDefinition{Token: p.curToken, Kind: "method"}

	if p.curTokenIs(token.Identifier) && p.curToken.Literal == "static" {
		md.Static = true
		p.nextToken()
	}

	if p.curTokenIs(token.Identifier) && (p.curToken.Literal == "get" || p.curToken.Literal == "set") {
		if !p.peekTokenIs(token.LeftParen) {
			md.Kind = p.curToken.Literal
			p.nextToken()
		}
	}

	if p.curTokenIs(token.Asterisk) {
		p.nextToken()
		md.Key = p.parseMethodKey(md)
		fe := p.parseMethodFunctionExpression()
		fe.Generator = true
		md.Value = fe
		return md
	}

	if (p.curTokenIs(token.Async) || (p.curTokenIs(token.Identifier) && p.curToken.Literal == "async")) && !p.peekTokenIs(token.LeftParen) {
		p.nextToken()
		isGen := false
		if p.curTokenIs(token.Asterisk) {
			isGen = true
			p.nextToken()
		}
		md.Key = p.parseMethodKey(md)
		fe := p.parseMethodFunctionExpression()
		fe.Async = true
		fe.Generator = isGen
		md.Value = fe
		return md
	}

	md.Key = p.parseMethodKey(md)

	if ident, ok := md.Key.(*ast.Identifier); ok && ident.Value == "constructor" && md.Kind == "method" {
		md.Kind = "constructor"
	}

	md.Value = p.parseMethodFunctionExpression()
	return md
}

func (p *Parser) parseMethodKey(md *ast.MethodDefinition) ast.Expression {
	if p.curTokenIs(token.LeftBracket) {
		md.Computed = true
		p.nextToken()
		key := p.parseAssignmentExpression()
		p.expect(token.RightBracket)
		return key
	}
	return p.parsePropertyName()
}

func (p *Parser) parseMethodFunctionExpression() *ast.FunctionExpression {
	fe := &ast.FunctionExpression{Token: p.curToken}
	target := funcExprTarget{fe}
	p.parseFunctionParamsGeneric(target)
	fe.Body = p.parseBlockStatement()
	return fe
}

func (p *Parser) parseLabeledStatement() *ast.LabeledStatement {
	stmt := &ast.LabeledStatement{Token: p.curToken}
	stmt.Label = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken() // consume identifier
	p.nextToken() // consume colon
	stmt.Body = p.parseStatement()
	return stmt
}

func (p *Parser) parseDebuggerStatement() *ast.DebuggerStatement {
	stmt := &ast.DebuggerStatement{Token: p.curToken}
	p.nextToken()
	p.consumeSemicolon()
	return stmt
}

func (p *Parser) parseEmptyStatement() *ast.EmptyStatement {
	stmt := &ast.EmptyStatement{Token: p.curToken}
	p.nextToken()
	return stmt
}

func (p *Parser) parseWithStatement() *ast.WithStatement {
	stmt := &ast.WithStatement{Token: p.curToken}
	p.nextToken() // consume with
	p.expect(token.LeftParen)
	stmt.Object = p.parseExpression(precComma)
	p.expect(token.RightParen)
	stmt.Body = p.parseStatement()
	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(0)
	p.consumeSemicolon()
	return stmt
}

// ---------- Expression Parsing (Pratt) ----------

func (p *Parser) parseExpression(minPrec int) ast.Expression {
	left := p.parsePrefixExpression()
	for {
		prec := p.infixPrecedence()
		if prec <= minPrec {
			break
		}
		left = p.parseInfixExpression(left, prec)
	}
	return left
}

func (p *Parser) parseAssignmentExpression() ast.Expression {
	return p.parseExpression(precComma)
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	switch p.curToken.Type {
	case token.Identifier:
		return p.parseIdentifierOrArrow()
	case token.Number:
		return p.parseNumberLiteral()
	case token.String:
		return p.parseStringLiteral()
	case token.True, token.False:
		return p.parseBooleanLiteral()
	case token.Null:
		return p.parseNullLiteral()
	case token.Undefined:
		return p.parseUndefinedLiteral()
	case token.LeftParen:
		return p.parseParenthesizedOrArrow()
	case token.LeftBracket:
		return p.parseArrayLiteral()
	case token.LeftBrace:
		return p.parseObjectLiteral()
	case token.Function:
		return p.parseFunctionExpression()
	case token.Class:
		return p.parseClassExpression()
	case token.This:
		return p.parseThisExpression()
	case token.Super:
		return p.parseSuperExpression()
	case token.New:
		return p.parseNewExpression()
	case token.Not, token.BitwiseNot, token.Typeof, token.Void, token.Delete:
		return p.parseUnaryExpression()
	case token.Plus, token.Minus:
		return p.parseUnaryExpression()
	case token.Increment, token.Decrement:
		return p.parsePrefixUpdateExpression()
	case token.Spread:
		return p.parseSpreadElement()
	case token.Yield:
		return p.parseYieldExpression()
	case token.Await:
		return p.parseAwaitExpression()
	case token.Async:
		return p.parseAsyncExpressionPrefix()
	case token.NoSubstitutionTemplate:
		return p.parseNoSubstitutionTemplate()
	case token.TemplateHead:
		return p.parseTemplateLiteral()
	case token.RegExp:
		return p.parseRegExpLiteral()
	default:
		p.addError("unexpected token %s (%q)", tokenName(p.curToken.Type), p.curToken.Literal)
		tok := p.curToken
		p.nextToken()
		return &ast.Identifier{Token: tok, Value: tok.Literal}
	}
}

func (p *Parser) parseIdentifierOrArrow() ast.Expression {
	if p.curTokenIs(token.Identifier) && p.peekTokenIs(token.Arrow) {
		return p.parseSingleParamArrow()
	}
	return p.parseIdentifier()
}

func (p *Parser) parseSingleParamArrow() ast.Expression {
	paramTok := p.curToken
	param := &ast.Identifier{Token: paramTok, Value: paramTok.Literal}
	p.nextToken() // consume identifier
	arrowTok := p.curToken
	p.nextToken() // consume =>
	arrow := &ast.ArrowFunctionExpression{Token: arrowTok, Params: []ast.Expression{param}}
	if p.curTokenIs(token.LeftBrace) {
		arrow.Body = p.parseBlockStatement()
	} else {
		arrow.Body = p.parseAssignmentExpression()
	}
	return arrow
}

func (p *Parser) parseAsyncExpressionPrefix() ast.Expression {
	asyncTok := p.curToken

	if p.peekTokenIs(token.Function) {
		return p.parseAsyncFunctionExpression()
	}

	if p.peekTokenIs(token.LeftParen) {
		return p.parseAsyncArrowOrCall()
	}

	if p.peekTokenIs(token.Identifier) {
		p.nextToken()
		if p.peekTokenIs(token.Arrow) {
			paramTok := p.curToken
			param := &ast.Identifier{Token: paramTok, Value: paramTok.Literal}
			p.nextToken() // consume identifier
			arrowTok := p.curToken
			p.nextToken() // consume =>
			arrow := &ast.ArrowFunctionExpression{
				Token:  arrowTok,
				Params: []ast.Expression{param},
				Async:  true,
			}
			if p.curTokenIs(token.LeftBrace) {
				arrow.Body = p.parseBlockStatement()
			} else {
				arrow.Body = p.parseAssignmentExpression()
			}
			return arrow
		}
		return &ast.Identifier{Token: asyncTok, Value: asyncTok.Literal}
	}

	return p.parseIdentifier()
}

func (p *Parser) parseAsyncArrowOrCall() ast.Expression {
	asyncTok := p.curToken
	p.nextToken() // consume async, now on (

	// Parse the parenthesized content
	result := p.parseParenthesizedOrArrow()

	// If it was parsed as an arrow, mark it async
	if arrow, ok := result.(*ast.ArrowFunctionExpression); ok {
		arrow.Async = true
		return arrow
	}

	// Otherwise it was a group expression, treat "async" as an identifier being called
	callee := &ast.Identifier{Token: asyncTok, Value: "async"}
	// The result was already parsed as a group, but we need to convert it to a call
	// Since parseParenthesizedOrArrow consumed the parens, we wrap it as a call
	var args []ast.Expression
	if seq, ok := result.(*ast.SequenceExpression); ok {
		args = seq.Expressions
	} else {
		args = []ast.Expression{result}
	}
	call := &ast.CallExpression{Token: asyncTok, Callee: callee, Arguments: args}
	return p.parsePostfixOps(call)
}

func (p *Parser) parseAsyncFunctionExpression() *ast.FunctionExpression {
	p.nextToken() // consume async
	fe := p.parseFunctionExpression()
	fe.Async = true
	return fe
}

func (p *Parser) parseIdentifier() *ast.Identifier {
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()
	return ident
}

func (p *Parser) parseNumberLiteral() *ast.NumberLiteral {
	lit := &ast.NumberLiteral{Token: p.curToken}
	val, err := parseJSNumber(p.curToken.Literal)
	if err != nil {
		p.addError("invalid number: %s", p.curToken.Literal)
	}
	lit.Value = val
	p.nextToken()
	return lit
}

func parseJSNumber(s string) (float64, error) {
	if len(s) > 0 && s[len(s)-1] == 'n' {
		s = s[:len(s)-1]
	}
	cleaned := ""
	for _, c := range s {
		if c != '_' {
			cleaned += string(c)
		}
	}
	if len(cleaned) > 2 {
		switch cleaned[:2] {
		case "0x", "0X":
			val, err := strconv.ParseInt(cleaned[2:], 16, 64)
			return float64(val), err
		case "0o", "0O":
			val, err := strconv.ParseInt(cleaned[2:], 8, 64)
			return float64(val), err
		case "0b", "0B":
			val, err := strconv.ParseInt(cleaned[2:], 2, 64)
			return float64(val), err
		}
	}
	return strconv.ParseFloat(cleaned, 64)
}

func (p *Parser) parseStringLiteral() *ast.StringLiteral {
	lit := &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()
	return lit
}

func (p *Parser) parseBooleanLiteral() *ast.BooleanLiteral {
	lit := &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.True)}
	p.nextToken()
	return lit
}

func (p *Parser) parseNullLiteral() *ast.NullLiteral {
	lit := &ast.NullLiteral{Token: p.curToken}
	p.nextToken()
	return lit
}

func (p *Parser) parseUndefinedLiteral() *ast.UndefinedLiteral {
	lit := &ast.UndefinedLiteral{Token: p.curToken}
	p.nextToken()
	return lit
}

func (p *Parser) parseThisExpression() *ast.ThisExpression {
	expr := &ast.ThisExpression{Token: p.curToken}
	p.nextToken()
	return expr
}

func (p *Parser) parseSuperExpression() *ast.SuperExpression {
	expr := &ast.SuperExpression{Token: p.curToken}
	p.nextToken()
	return expr
}

func (p *Parser) parseParenthesizedOrArrow() ast.Expression {
	// Use a fresh lexer-based parser to speculatively test for arrow params.
	// We create an entirely new parser from the same source position if needed.
	// Simpler approach: parse as group expression, then if we see => after ),
	// convert the parsed expressions into arrow params.
	openTok := p.curToken
	p.nextToken() // consume (

	if p.curTokenIs(token.RightParen) {
		p.nextToken()
		if p.curTokenIs(token.Arrow) {
			arrowTok := p.curToken
			p.nextToken()
			arrow := &ast.ArrowFunctionExpression{Token: arrowTok}
			if p.curTokenIs(token.LeftBrace) {
				arrow.Body = p.parseBlockStatement()
			} else {
				arrow.Body = p.parseAssignmentExpression()
			}
			return arrow
		}
		p.addError("unexpected token after ()")
		return &ast.Identifier{Token: p.curToken}
	}

	// Try to parse content - if it looks like params and we see =>, it's an arrow
	var items []ast.Expression
	var defaults []ast.Expression
	var rest ast.Expression
	hasDefaults := false
	canBeArrow := true

	for !p.curTokenIs(token.RightParen) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.Spread) {
			restTok := p.curToken
			p.nextToken()
			arg := p.parseBindingPattern()
			rest = &ast.RestElement{Token: restTok, Argument: arg}
			if p.curTokenIs(token.Comma) {
				p.nextToken()
			}
			break
		}

		item := p.parseAssignmentExpression()
		items = append(items, item)
		defaults = append(defaults, nil)

		// Check if this was param = default
		if assign, ok := item.(*ast.AssignmentExpression); ok && assign.Operator == "=" {
			hasDefaults = true
			items[len(items)-1] = assign.Left
			defaults[len(defaults)-1] = assign.Right
		}

		if !p.curTokenIs(token.Comma) {
			break
		}
		p.nextToken()
	}

	if !p.curTokenIs(token.RightParen) {
		canBeArrow = false
	}

	p.expect(token.RightParen)

	if canBeArrow && p.curTokenIs(token.Arrow) {
		arrowTok := p.curToken
		p.nextToken() // consume =>
		arrow := &ast.ArrowFunctionExpression{Token: arrowTok, Params: items}
		if hasDefaults {
			arrow.Defaults = defaults
		}
		if rest != nil {
			arrow.Rest = rest
		}
		if p.curTokenIs(token.LeftBrace) {
			arrow.Body = p.parseBlockStatement()
		} else {
			arrow.Body = p.parseAssignmentExpression()
		}
		return arrow
	}

	// Not an arrow - return the parsed expression
	_ = openTok
	if len(items) == 0 {
		p.addError("empty parenthesized expression")
		return &ast.Identifier{Token: openTok}
	}
	if len(items) == 1 && rest == nil {
		return items[0]
	}
	// Multiple items = sequence expression
	seq := &ast.SequenceExpression{Token: openTok, Expressions: items}
	return seq
}


func (p *Parser) parseGroupExpression() ast.Expression {
	p.nextToken() // consume (

	if p.curTokenIs(token.RightParen) {
		p.nextToken()
		if p.curTokenIs(token.Arrow) {
			arrowTok := p.curToken
			p.nextToken()
			arrow := &ast.ArrowFunctionExpression{Token: arrowTok}
			if p.curTokenIs(token.LeftBrace) {
				arrow.Body = p.parseBlockStatement()
			} else {
				arrow.Body = p.parseAssignmentExpression()
			}
			return arrow
		}
		p.addError("unexpected token after ()")
		return &ast.Identifier{Token: p.curToken}
	}

	expr := p.parseExpression(0)

	p.expect(token.RightParen)
	return expr
}

func (p *Parser) parseArrayLiteral() *ast.ArrayLiteral {
	arr := &ast.ArrayLiteral{Token: p.curToken}
	p.nextToken() // consume [

	for !p.curTokenIs(token.RightBracket) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.Comma) {
			arr.Elements = append(arr.Elements, nil)
			p.nextToken()
			continue
		}
		if p.curTokenIs(token.Spread) {
			spread := &ast.SpreadElement{Token: p.curToken}
			p.nextToken()
			spread.Argument = p.parseAssignmentExpression()
			arr.Elements = append(arr.Elements, spread)
		} else {
			arr.Elements = append(arr.Elements, p.parseAssignmentExpression())
		}
		if !p.curTokenIs(token.Comma) {
			break
		}
		p.nextToken()
	}
	p.expect(token.RightBracket)
	return arr
}

func (p *Parser) parseObjectLiteral() *ast.ObjectLiteral {
	obj := &ast.ObjectLiteral{Token: p.curToken}
	p.nextToken() // consume {

	for !p.curTokenIs(token.RightBrace) && !p.curTokenIs(token.EOF) {
		prop := p.parseObjectProperty()
		obj.Properties = append(obj.Properties, prop)
		if !p.curTokenIs(token.Comma) {
			break
		}
		p.nextToken()
	}
	p.expect(token.RightBrace)
	return obj
}

func (p *Parser) parseObjectProperty() *ast.Property {
	prop := &ast.Property{Token: p.curToken, Kind: "init"}

	// Spread property
	if p.curTokenIs(token.Spread) {
		spread := &ast.SpreadElement{Token: p.curToken}
		p.nextToken()
		spread.Argument = p.parseAssignmentExpression()
		return &ast.Property{Token: spread.Token, Key: spread, Value: spread, Kind: "init"}
	}

	// getter/setter
	if p.curTokenIs(token.Identifier) && (p.curToken.Literal == "get" || p.curToken.Literal == "set") {
		kindTok := p.curToken
		kind := kindTok.Literal
		if p.peekTokenIs(token.LeftParen) || p.peekTokenIs(token.Colon) || p.peekTokenIs(token.Comma) || p.peekTokenIs(token.RightBrace) || p.peekTokenIs(token.Assign) {
			goto normalProperty
		}
		p.nextToken()
		prop.Kind = kind
		prop.Key = p.parseObjectPropertyKey(prop)
		fe := &ast.FunctionExpression{Token: p.curToken}
		target := funcExprTarget{fe}
		p.parseFunctionParamsGeneric(target)
		fe.Body = p.parseBlockStatement()
		prop.Value = fe
		prop.Method = true
		return prop
	}

normalProperty:
	// async method
	if (p.curTokenIs(token.Async) || (p.curTokenIs(token.Identifier) && p.curToken.Literal == "async")) && !p.peekTokenIs(token.Colon) && !p.peekTokenIs(token.Comma) && !p.peekTokenIs(token.RightBrace) && !p.peekTokenIs(token.LeftParen) && !p.peekTokenIs(token.Assign) {
		p.nextToken()
		isGen := false
		if p.curTokenIs(token.Asterisk) {
			isGen = true
			p.nextToken()
		}
		prop.Key = p.parseObjectPropertyKey(prop)
		fe := &ast.FunctionExpression{Token: p.curToken}
		target := funcExprTarget{fe}
		p.parseFunctionParamsGeneric(target)
		fe.Body = p.parseBlockStatement()
		fe.Async = true
		fe.Generator = isGen
		prop.Value = fe
		prop.Method = true
		return prop
	}

	// Generator method
	if p.curTokenIs(token.Asterisk) {
		p.nextToken()
		prop.Key = p.parseObjectPropertyKey(prop)
		fe := &ast.FunctionExpression{Token: p.curToken}
		target := funcExprTarget{fe}
		p.parseFunctionParamsGeneric(target)
		fe.Body = p.parseBlockStatement()
		fe.Generator = true
		prop.Value = fe
		prop.Method = true
		return prop
	}

	prop.Key = p.parseObjectPropertyKey(prop)

	// Method shorthand: key(...)
	if p.curTokenIs(token.LeftParen) {
		fe := &ast.FunctionExpression{Token: p.curToken}
		target := funcExprTarget{fe}
		p.parseFunctionParamsGeneric(target)
		fe.Body = p.parseBlockStatement()
		prop.Value = fe
		prop.Method = true
		return prop
	}

	// key: value
	if p.curTokenIs(token.Colon) {
		p.nextToken()
		prop.Value = p.parseAssignmentExpression()
		return prop
	}

	// Shorthand property {x} or {x = default}
	prop.Shorthand = true
	prop.Value = prop.Key
	if p.curTokenIs(token.Assign) {
		p.nextToken()
		prop.Value = &ast.AssignmentPattern{
			Token: prop.Token,
			Left:  prop.Key,
			Right: p.parseAssignmentExpression(),
		}
	}
	return prop
}

func (p *Parser) parseObjectPropertyKey(prop *ast.Property) ast.Expression {
	if p.curTokenIs(token.LeftBracket) {
		prop.Computed = true
		p.nextToken()
		key := p.parseAssignmentExpression()
		p.expect(token.RightBracket)
		return key
	}
	return p.parsePropertyName()
}

func (p *Parser) parsePropertyName() ast.Expression {
	switch p.curToken.Type {
	case token.Identifier, token.Var, token.Let, token.Const, token.Function, token.Return,
		token.If, token.Else, token.While, token.For, token.Do, token.Break, token.Continue,
		token.Switch, token.Case, token.Default, token.Throw, token.Try, token.Catch,
		token.Finally, token.New, token.Delete, token.Typeof, token.Void, token.In,
		token.Instanceof, token.This, token.Class, token.Extends, token.Super, token.Import,
		token.Export, token.From, token.As, token.Of, token.Yield, token.Async, token.Await,
		token.True, token.False, token.Null, token.Undefined, token.Debugger, token.With:
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
		return ident
	case token.Number:
		return p.parseNumberLiteral()
	case token.String:
		return p.parseStringLiteral()
	default:
		p.addError("unexpected token in property name: %s", tokenName(p.curToken.Type))
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
		return ident
	}
}

func (p *Parser) parseFunctionExpression() *ast.FunctionExpression {
	fe := &ast.FunctionExpression{Token: p.curToken}
	p.nextToken() // consume function

	if p.curTokenIs(token.Asterisk) {
		fe.Generator = true
		p.nextToken()
	}

	if p.curTokenIs(token.Identifier) {
		fe.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
	}

	target := funcExprTarget{fe}
	p.parseFunctionParamsGeneric(target)
	fe.Body = p.parseBlockStatement()
	return fe
}

func (p *Parser) parseClassExpression() *ast.ClassExpression {
	expr := &ast.ClassExpression{Token: p.curToken}
	p.nextToken() // consume class

	if p.curTokenIs(token.Identifier) && !p.curTokenIs(token.Extends) {
		expr.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
	}

	if p.curTokenIs(token.Extends) {
		p.nextToken()
		expr.SuperClass = p.parseLeftHandSideExpression()
	}

	expr.Body = p.parseClassBody()
	return expr
}

func (p *Parser) parseNewExpression() ast.Expression {
	tok := p.curToken
	p.nextToken() // consume new

	callee := p.parseLeftHandSideExpression()

	if p.curTokenIs(token.LeftParen) {
		args := p.parseArguments()
		expr := &ast.NewExpression{Token: tok, Callee: callee, Arguments: args}
		return p.parsePostfixOps(expr)
	}
	return &ast.NewExpression{Token: tok, Callee: callee}
}

func (p *Parser) parseLeftHandSideExpression() ast.Expression {
	var left ast.Expression
	switch p.curToken.Type {
	case token.Identifier:
		left = p.parseIdentifier()
	case token.This:
		left = p.parseThisExpression()
	case token.Super:
		left = p.parseSuperExpression()
	case token.LeftParen:
		left = p.parseGroupExpression()
	default:
		left = p.parsePrefixExpression()
	}

	for {
		if p.curTokenIs(token.Dot) {
			tok := p.curToken
			p.nextToken()
			prop := p.parsePropertyName()
			left = &ast.MemberExpression{Token: tok, Object: left, Property: prop}
		} else if p.curTokenIs(token.LeftBracket) {
			tok := p.curToken
			p.nextToken()
			prop := p.parseExpression(precComma)
			p.expect(token.RightBracket)
			left = &ast.MemberExpression{Token: tok, Object: left, Property: prop, Computed: true}
		} else {
			break
		}
	}
	return left
}

func (p *Parser) parseUnaryExpression() ast.Expression {
	tok := p.curToken
	op := tok.Literal
	p.nextToken()
	operand := p.parseExpression(precUnary)
	return &ast.UnaryExpression{Token: tok, Operator: op, Operand: operand, Prefix: true}
}

func (p *Parser) parsePrefixUpdateExpression() ast.Expression {
	tok := p.curToken
	op := tok.Literal
	p.nextToken()
	operand := p.parseExpression(precUnary)
	return &ast.UpdateExpression{Token: tok, Operator: op, Operand: operand, Prefix: true}
}

func (p *Parser) parseSpreadElement() *ast.SpreadElement {
	spread := &ast.SpreadElement{Token: p.curToken}
	p.nextToken() // consume ...
	spread.Argument = p.parseAssignmentExpression()
	return spread
}

func (p *Parser) parseYieldExpression() *ast.YieldExpression {
	expr := &ast.YieldExpression{Token: p.curToken}
	p.nextToken() // consume yield

	if p.curTokenIs(token.Asterisk) {
		expr.Delegate = true
		p.nextToken()
	}

	if !p.curTokenIs(token.Semicolon) && !p.curTokenIs(token.RightBrace) &&
		!p.curTokenIs(token.RightParen) && !p.curTokenIs(token.RightBracket) &&
		!p.curTokenIs(token.Comma) && !p.curTokenIs(token.Colon) && !p.curTokenIs(token.EOF) {
		expr.Argument = p.parseAssignmentExpression()
	}
	return expr
}

func (p *Parser) parseAwaitExpression() *ast.AwaitExpression {
	expr := &ast.AwaitExpression{Token: p.curToken}
	p.nextToken()
	expr.Argument = p.parseExpression(precUnary)
	return expr
}

func (p *Parser) parseNoSubstitutionTemplate() *ast.TemplateLiteralExpr {
	tmpl := &ast.TemplateLiteralExpr{Token: p.curToken}
	tmpl.Quasis = append(tmpl.Quasis, &ast.TemplateElement{
		Token: p.curToken,
		Value: p.curToken.Literal,
		Tail:  true,
	})
	p.nextToken()
	return tmpl
}

func (p *Parser) parseTemplateLiteral() *ast.TemplateLiteralExpr {
	tmpl := &ast.TemplateLiteralExpr{Token: p.curToken}
	tmpl.Quasis = append(tmpl.Quasis, &ast.TemplateElement{
		Token: p.curToken,
		Value: p.curToken.Literal,
	})
	p.nextToken() // move past TemplateHead

	for {
		expr := p.parseExpression(precComma)
		tmpl.Expressions = append(tmpl.Expressions, expr)

		if p.curTokenIs(token.TemplateTail) {
			tmpl.Quasis = append(tmpl.Quasis, &ast.TemplateElement{
				Token: p.curToken,
				Value: p.curToken.Literal,
				Tail:  true,
			})
			p.nextToken()
			break
		}
		if p.curTokenIs(token.TemplateMiddle) {
			tmpl.Quasis = append(tmpl.Quasis, &ast.TemplateElement{
				Token: p.curToken,
				Value: p.curToken.Literal,
			})
			p.nextToken()
			continue
		}
		p.addError("expected template middle or tail, got %s", tokenName(p.curToken.Type))
		break
	}
	return tmpl
}

func (p *Parser) parseRegExpLiteral() ast.Expression {
	raw := p.curToken.Literal // e.g. "/pattern/flags"
	// Find last '/' which separates pattern from flags
	lastSlash := strings.LastIndex(raw, "/")
	pattern := raw[1:lastSlash]
	flags := raw[lastSlash+1:]
	lit := &ast.RegExpLiteral{Token: p.curToken, Pattern: pattern, Flags: flags}
	p.nextToken()
	return lit
}

// ---------- Infix Parsing ----------

func (p *Parser) infixPrecedence() int {
	switch p.curToken.Type {
	case token.Comma:
		return precComma
	case token.Assign, token.PlusAssign, token.MinusAssign, token.AsteriskAssign,
		token.SlashAssign, token.PercentAssign, token.ExponentAssign,
		token.AmpersandAssign, token.PipeAssign, token.CaretAssign,
		token.LeftShiftAssign, token.RightShiftAssign, token.UnsignedRightShiftAssign,
		token.NullishAssign, token.AndAssign, token.OrAssign:
		return precAssignment
	case token.QuestionMark:
		return precConditional
	case token.NullishCoalesce:
		return precNullishCoalesce
	case token.Or:
		return precLogicalOr
	case token.And:
		return precLogicalAnd
	case token.BitwiseOr:
		return precBitwiseOr
	case token.BitwiseXor:
		return precBitwiseXor
	case token.BitwiseAnd:
		return precBitwiseAnd
	case token.Equal, token.NotEqual, token.StrictEqual, token.StrictNotEqual:
		return precEquality
	case token.LessThan, token.GreaterThan, token.LessThanOrEqual, token.GreaterThanOrEqual,
		token.Instanceof:
		return precRelational
	case token.In:
		if p.noIn {
			return 0
		}
		return precRelational
	case token.LeftShift, token.RightShift, token.UnsignedRightShift:
		return precShift
	case token.Plus, token.Minus:
		return precAdditive
	case token.Asterisk, token.Slash, token.Percent:
		return precMultiplicative
	case token.Exponent:
		return precExponent
	case token.Increment, token.Decrement:
		return precPostfix
	case token.LeftParen:
		return precCall
	case token.Dot, token.LeftBracket, token.OptionalChain:
		return precMember
	case token.TemplateHead, token.NoSubstitutionTemplate:
		return precMember
	default:
		return 0
	}
}

func (p *Parser) parseInfixExpression(left ast.Expression, prec int) ast.Expression {
	switch p.curToken.Type {
	case token.Comma:
		return p.parseSequenceExpression(left)
	case token.Assign, token.PlusAssign, token.MinusAssign, token.AsteriskAssign,
		token.SlashAssign, token.PercentAssign, token.ExponentAssign,
		token.AmpersandAssign, token.PipeAssign, token.CaretAssign,
		token.LeftShiftAssign, token.RightShiftAssign, token.UnsignedRightShiftAssign,
		token.NullishAssign, token.AndAssign, token.OrAssign:
		return p.parseAssignmentInfix(left)
	case token.QuestionMark:
		return p.parseConditionalExpression(left)
	case token.Or, token.And:
		return p.parseLogicalInfix(left)
	case token.Exponent:
		return p.parseExponentInfix(left)
	case token.LeftParen:
		return p.parseCallExpression(left)
	case token.Dot:
		return p.parseDotMember(left)
	case token.LeftBracket:
		return p.parseBracketMember(left)
	case token.OptionalChain:
		return p.parseOptionalChain(left)
	case token.Increment, token.Decrement:
		return p.parsePostfixUpdate(left)
	case token.TemplateHead, token.NoSubstitutionTemplate:
		return p.parseTaggedTemplate(left)
	default:
		return p.parseBinaryInfix(left)
	}
}

func (p *Parser) parseSequenceExpression(left ast.Expression) ast.Expression {
	seq := &ast.SequenceExpression{Token: p.curToken, Expressions: []ast.Expression{left}}
	for p.curTokenIs(token.Comma) {
		p.nextToken()
		seq.Expressions = append(seq.Expressions, p.parseAssignmentExpression())
	}
	return seq
}

func (p *Parser) parseAssignmentInfix(left ast.Expression) ast.Expression {
	tok := p.curToken
	p.nextToken()
	right := p.parseAssignmentExpression()
	return &ast.AssignmentExpression{Token: tok, Operator: tok.Literal, Left: left, Right: right}
}

func (p *Parser) parseConditionalExpression(left ast.Expression) ast.Expression {
	tok := p.curToken
	p.nextToken() // consume ?
	consequent := p.parseAssignmentExpression()
	p.expect(token.Colon)
	alternate := p.parseAssignmentExpression()
	return &ast.ConditionalExpression{Token: tok, Test: left, Consequent: consequent, Alternate: alternate}
}

func (p *Parser) parseLogicalInfix(left ast.Expression) ast.Expression {
	tok := p.curToken
	prec := p.infixPrecedence()
	p.nextToken()
	right := p.parseExpression(prec)
	return &ast.LogicalExpression{Token: tok, Operator: tok.Literal, Left: left, Right: right}
}

func (p *Parser) parseBinaryInfix(left ast.Expression) ast.Expression {
	tok := p.curToken
	prec := p.infixPrecedence()
	p.nextToken()
	right := p.parseExpression(prec)
	return &ast.BinaryExpression{Token: tok, Operator: tok.Literal, Left: left, Right: right}
}

func (p *Parser) parseExponentInfix(left ast.Expression) ast.Expression {
	tok := p.curToken
	p.nextToken()
	// Right-associative: use prec - 1
	right := p.parseExpression(precExponent - 1)
	return &ast.BinaryExpression{Token: tok, Operator: tok.Literal, Left: left, Right: right}
}

func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	tok := p.curToken
	args := p.parseArguments()
	return p.parsePostfixOps(&ast.CallExpression{Token: tok, Callee: left, Arguments: args})
}

func (p *Parser) parseArguments() []ast.Expression {
	p.nextToken() // consume (
	var args []ast.Expression

	for !p.curTokenIs(token.RightParen) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.Spread) {
			args = append(args, p.parseSpreadElement())
		} else {
			args = append(args, p.parseAssignmentExpression())
		}
		if !p.curTokenIs(token.Comma) {
			break
		}
		p.nextToken()
	}
	p.expect(token.RightParen)
	return args
}

func (p *Parser) parseDotMember(left ast.Expression) ast.Expression {
	tok := p.curToken
	p.nextToken() // consume .
	prop := p.parsePropertyName()
	result := &ast.MemberExpression{Token: tok, Object: left, Property: prop}
	return p.parsePostfixOps(result)
}

func (p *Parser) parseBracketMember(left ast.Expression) ast.Expression {
	tok := p.curToken
	p.nextToken() // consume [
	prop := p.parseExpression(precComma)
	p.expect(token.RightBracket)
	result := &ast.MemberExpression{Token: tok, Object: left, Property: prop, Computed: true}
	return p.parsePostfixOps(result)
}

func (p *Parser) parseOptionalChain(left ast.Expression) ast.Expression {
	tok := p.curToken
	p.nextToken() // consume ?.

	if p.curTokenIs(token.LeftParen) {
		args := p.parseArguments()
		return p.parsePostfixOps(&ast.CallExpression{Token: tok, Callee: left, Arguments: args})
	}

	if p.curTokenIs(token.LeftBracket) {
		p.nextToken()
		prop := p.parseExpression(precComma)
		p.expect(token.RightBracket)
		return p.parsePostfixOps(&ast.MemberExpression{Token: tok, Object: left, Property: prop, Computed: true})
	}

	prop := p.parsePropertyName()
	return p.parsePostfixOps(&ast.MemberExpression{Token: tok, Object: left, Property: prop})
}

func (p *Parser) parsePostfixUpdate(left ast.Expression) ast.Expression {
	tok := p.curToken
	p.nextToken()
	return &ast.UpdateExpression{Token: tok, Operator: tok.Literal, Operand: left, Prefix: false}
}

func (p *Parser) parseTaggedTemplate(tag ast.Expression) ast.Expression {
	tok := p.curToken
	var quasi *ast.TemplateLiteralExpr
	if p.curTokenIs(token.NoSubstitutionTemplate) {
		quasi = p.parseNoSubstitutionTemplate()
	} else {
		quasi = p.parseTemplateLiteral()
	}
	return &ast.TaggedTemplateExpression{Token: tok, Tag: tag, Quasi: quasi}
}

func (p *Parser) parsePostfixOps(expr ast.Expression) ast.Expression {
	for {
		switch p.curToken.Type {
		case token.Dot:
			tok := p.curToken
			p.nextToken()
			prop := p.parsePropertyName()
			expr = &ast.MemberExpression{Token: tok, Object: expr, Property: prop}
		case token.LeftBracket:
			tok := p.curToken
			p.nextToken()
			prop := p.parseExpression(precComma)
			p.expect(token.RightBracket)
			expr = &ast.MemberExpression{Token: tok, Object: expr, Property: prop, Computed: true}
		case token.LeftParen:
			tok := p.curToken
			args := p.parseArguments()
			expr = &ast.CallExpression{Token: tok, Callee: expr, Arguments: args}
		case token.OptionalChain:
			tok := p.curToken
			p.nextToken()
			if p.curTokenIs(token.LeftParen) {
				args := p.parseArguments()
				expr = &ast.CallExpression{Token: tok, Callee: expr, Arguments: args}
			} else if p.curTokenIs(token.LeftBracket) {
				p.nextToken()
				prop := p.parseExpression(precComma)
				p.expect(token.RightBracket)
				expr = &ast.MemberExpression{Token: tok, Object: expr, Property: prop, Computed: true}
			} else {
				prop := p.parsePropertyName()
				expr = &ast.MemberExpression{Token: tok, Object: expr, Property: prop}
			}
		case token.TemplateHead, token.NoSubstitutionTemplate:
			expr = p.parseTaggedTemplate(expr)
		default:
			return expr
		}
	}
}

// ---------- Helpers ----------

func (p *Parser) consumeSemicolon() {
	if p.curTokenIs(token.Semicolon) {
		p.nextToken()
	}
}

func (p *Parser) prevTokenWasNewline() bool {
	return p.curToken.Line > p.prevLine
}

func tokenName(t token.TokenType) string {
	names := map[token.TokenType]string{
		token.EOF:                      "EOF",
		token.Illegal:                  "ILLEGAL",
		token.Identifier:              "IDENTIFIER",
		token.Number:                  "NUMBER",
		token.String:                  "STRING",
		token.Plus:                    "+",
		token.Minus:                   "-",
		token.Asterisk:                "*",
		token.Slash:                   "/",
		token.Percent:                 "%",
		token.Exponent:                "**",
		token.Assign:                  "=",
		token.PlusAssign:              "+=",
		token.MinusAssign:             "-=",
		token.AsteriskAssign:          "*=",
		token.SlashAssign:             "/=",
		token.PercentAssign:           "%=",
		token.ExponentAssign:          "**=",
		token.AmpersandAssign:         "&=",
		token.PipeAssign:              "|=",
		token.CaretAssign:             "^=",
		token.LeftShiftAssign:         "<<=",
		token.RightShiftAssign:        ">>=",
		token.UnsignedRightShiftAssign: ">>>=",
		token.NullishAssign:           "??=",
		token.AndAssign:               "&&=",
		token.OrAssign:                "||=",
		token.Equal:                   "==",
		token.NotEqual:                "!=",
		token.StrictEqual:             "===",
		token.StrictNotEqual:          "!==",
		token.LessThan:                "<",
		token.GreaterThan:             ">",
		token.LessThanOrEqual:         "<=",
		token.GreaterThanOrEqual:      ">=",
		token.And:                     "&&",
		token.Or:                      "||",
		token.Not:                     "!",
		token.BitwiseAnd:              "&",
		token.BitwiseOr:               "|",
		token.BitwiseXor:              "^",
		token.BitwiseNot:              "~",
		token.LeftShift:               "<<",
		token.RightShift:              ">>",
		token.UnsignedRightShift:      ">>>",
		token.Increment:               "++",
		token.Decrement:               "--",
		token.LeftParen:               "(",
		token.RightParen:              ")",
		token.LeftBrace:               "{",
		token.RightBrace:              "}",
		token.LeftBracket:             "[",
		token.RightBracket:            "]",
		token.Semicolon:               ";",
		token.Colon:                   ":",
		token.Comma:                   ",",
		token.Dot:                     ".",
		token.Spread:                  "...",
		token.Arrow:                   "=>",
		token.QuestionMark:            "?",
		token.OptionalChain:           "?.",
		token.NullishCoalesce:         "??",
		token.Var:                     "var",
		token.Let:                     "let",
		token.Const:                   "const",
		token.Function:                "function",
		token.Return:                  "return",
		token.If:                      "if",
		token.Else:                    "else",
		token.While:                   "while",
		token.For:                     "for",
		token.Do:                      "do",
		token.Break:                   "break",
		token.Continue:                "continue",
		token.Switch:                  "switch",
		token.Case:                    "case",
		token.Default:                 "default",
		token.Throw:                   "throw",
		token.Try:                     "try",
		token.Catch:                   "catch",
		token.Finally:                 "finally",
		token.New:                     "new",
		token.Delete:                  "delete",
		token.Typeof:                  "typeof",
		token.Void:                    "void",
		token.In:                      "in",
		token.Instanceof:              "instanceof",
		token.This:                    "this",
		token.Class:                   "class",
		token.Extends:                 "extends",
		token.Super:                   "super",
		token.Import:                  "import",
		token.Export:                   "export",
		token.From:                    "from",
		token.As:                      "as",
		token.Of:                      "of",
		token.Yield:                   "yield",
		token.Async:                   "async",
		token.Await:                   "await",
		token.True:                    "true",
		token.False:                   "false",
		token.Null:                    "null",
		token.Undefined:               "undefined",
		token.Debugger:                "debugger",
		token.With:                    "with",
		token.TemplateHead:            "TEMPLATE_HEAD",
		token.TemplateMiddle:          "TEMPLATE_MIDDLE",
		token.TemplateTail:            "TEMPLATE_TAIL",
		token.NoSubstitutionTemplate:  "TEMPLATE",
		token.RegExp:                  "REGEXP",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return fmt.Sprintf("TOKEN(%d)", t)
}
