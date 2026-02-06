package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/example/jsgo/token"
)

type Lexer struct {
	input   string
	pos     int // current position in input (points to current char)
	readPos int // current reading position (after current char)
	ch      rune
	line    int
	col     int

	// For template literal interpolation tracking
	braceDepth    int
	templateStack []int // stack of brace depths where template interpolations started
}

func New(input string) *Lexer {
	l := &Lexer{
		input: input,
		line:  1,
		col:   0,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
		l.pos = l.readPos
		l.readPos++
		l.col++
		return
	}
	r, size := utf8.DecodeRuneInString(l.input[l.readPos:])
	l.ch = r
	l.pos = l.readPos
	l.readPos += size
	l.col++
}

func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPos:])
	return r
}

func (l *Lexer) peekCharAt(offset int) rune {
	pos := l.readPos + offset
	if pos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[pos:])
	return r
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' || l.ch == '\n' {
		if l.ch == '\n' {
			l.line++
			l.col = 0
		}
		l.readChar()
	}
}

func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func (l *Lexer) skipBlockComment() {
	// skip past /*
	l.readChar()
	l.readChar()
	for {
		if l.ch == 0 {
			return
		}
		if l.ch == '\n' {
			l.line++
			l.col = 0
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar()
			l.readChar()
			return
		}
		l.readChar()
	}
}

func (l *Lexer) skipWhitespaceAndComments() {
	sawNewline := l.col <= 1 // treat start of input as start of line
	for {
		prevLine := l.line
		l.skipWhitespace()
		if l.line > prevLine {
			sawNewline = true
		}
		if l.ch == '/' && l.peekChar() == '/' {
			l.skipLineComment()
			sawNewline = true // line comment ends at newline
			continue
		}
		if l.ch == '/' && l.peekChar() == '*' {
			prevLine = l.line
			l.skipBlockComment()
			if l.line > prevLine {
				sawNewline = true
			}
			continue
		}
		// Annex B: <!-- is a single-line comment (anywhere)
		if l.ch == '<' && l.peekChar() == '!' && l.peekCharAt(1) == '-' && l.peekCharAt(2) == '-' {
			l.skipLineComment()
			sawNewline = true
			continue
		}
		// Annex B: --> is a single-line comment ONLY after a line terminator
		if sawNewline && l.ch == '-' && l.peekChar() == '-' && l.peekCharAt(1) == '>' {
			l.skipLineComment()
			sawNewline = true
			continue
		}
		break
	}
}

// prevTokenType tracks what the last meaningful token was, for regex detection
var regexPrecedingTokens = map[token.TokenType]bool{
	token.Identifier:             false,
	token.Number:                 false,
	token.String:                 false,
	token.True:                   false,
	token.False:                  false,
	token.Null:                   false,
	token.This:                   false,
	token.RightParen:             false,
	token.RightBracket:           false,
	token.Increment:              false,
	token.Decrement:              false,
	token.NoSubstitutionTemplate: false,
	token.TemplateTail:           false,
}

func canPrecedeRegex(tt token.TokenType) bool {
	if _, found := regexPrecedingTokens[tt]; found {
		return false
	}
	return true
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespaceAndComments()

	line := l.line
	col := l.col

	tok := func(tt token.TokenType, lit string) token.Token {
		return token.Token{Type: tt, Literal: lit, Line: line, Column: col}
	}

	// Check for template middle/tail when closing a template interpolation
	if l.ch == '}' && len(l.templateStack) > 0 && l.braceDepth-1 == l.templateStack[len(l.templateStack)-1] {
		l.templateStack = l.templateStack[:len(l.templateStack)-1]
		return l.readTemplateContinuation(line, col)
	}

	switch {
	case l.ch == 0:
		return tok(token.EOF, "")

	case l.ch == '(':
		l.readChar()
		return tok(token.LeftParen, "(")
	case l.ch == ')':
		l.readChar()
		return tok(token.RightParen, ")")
	case l.ch == '{':
		l.braceDepth++
		l.readChar()
		return tok(token.LeftBrace, "{")
	case l.ch == '}':
		l.braceDepth--
		l.readChar()
		return tok(token.RightBrace, "}")
	case l.ch == '[':
		l.readChar()
		return tok(token.LeftBracket, "[")
	case l.ch == ']':
		l.readChar()
		return tok(token.RightBracket, "]")
	case l.ch == ';':
		l.readChar()
		return tok(token.Semicolon, ";")
	case l.ch == ':':
		l.readChar()
		return tok(token.Colon, ":")
	case l.ch == ',':
		l.readChar()
		return tok(token.Comma, ",")
	case l.ch == '~':
		l.readChar()
		return tok(token.BitwiseNot, "~")

	case l.ch == '.':
		if l.peekChar() == '.' && l.peekCharAt(1) == '.' {
			l.readChar()
			l.readChar()
			l.readChar()
			return tok(token.Spread, "...")
		}
		if isDigit(l.peekChar()) {
			return l.readNumber(line, col)
		}
		l.readChar()
		return tok(token.Dot, ".")

	case l.ch == '+':
		l.readChar()
		if l.ch == '+' {
			l.readChar()
			return tok(token.Increment, "++")
		}
		if l.ch == '=' {
			l.readChar()
			return tok(token.PlusAssign, "+=")
		}
		return tok(token.Plus, "+")

	case l.ch == '-':
		l.readChar()
		if l.ch == '-' {
			l.readChar()
			return tok(token.Decrement, "--")
		}
		if l.ch == '=' {
			l.readChar()
			return tok(token.MinusAssign, "-=")
		}
		return tok(token.Minus, "-")

	case l.ch == '*':
		l.readChar()
		if l.ch == '*' {
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return tok(token.ExponentAssign, "**=")
			}
			return tok(token.Exponent, "**")
		}
		if l.ch == '=' {
			l.readChar()
			return tok(token.AsteriskAssign, "*=")
		}
		return tok(token.Asterisk, "*")

	case l.ch == '/':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return tok(token.SlashAssign, "/=")
		}
		return tok(token.Slash, "/")

	case l.ch == '%':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return tok(token.PercentAssign, "%=")
		}
		return tok(token.Percent, "%")

	case l.ch == '=':
		l.readChar()
		if l.ch == '>' {
			l.readChar()
			return tok(token.Arrow, "=>")
		}
		if l.ch == '=' {
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return tok(token.StrictEqual, "===")
			}
			return tok(token.Equal, "==")
		}
		return tok(token.Assign, "=")

	case l.ch == '!':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return tok(token.StrictNotEqual, "!==")
			}
			return tok(token.NotEqual, "!=")
		}
		return tok(token.Not, "!")

	case l.ch == '<':
		l.readChar()
		if l.ch == '<' {
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return tok(token.LeftShiftAssign, "<<=")
			}
			return tok(token.LeftShift, "<<")
		}
		if l.ch == '=' {
			l.readChar()
			return tok(token.LessThanOrEqual, "<=")
		}
		return tok(token.LessThan, "<")

	case l.ch == '>':
		l.readChar()
		if l.ch == '>' {
			l.readChar()
			if l.ch == '>' {
				l.readChar()
				if l.ch == '=' {
					l.readChar()
					return tok(token.UnsignedRightShiftAssign, ">>>=")
				}
				return tok(token.UnsignedRightShift, ">>>")
			}
			if l.ch == '=' {
				l.readChar()
				return tok(token.RightShiftAssign, ">>=")
			}
			return tok(token.RightShift, ">>")
		}
		if l.ch == '=' {
			l.readChar()
			return tok(token.GreaterThanOrEqual, ">=")
		}
		return tok(token.GreaterThan, ">")

	case l.ch == '&':
		l.readChar()
		if l.ch == '&' {
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return tok(token.AndAssign, "&&=")
			}
			return tok(token.And, "&&")
		}
		if l.ch == '=' {
			l.readChar()
			return tok(token.AmpersandAssign, "&=")
		}
		return tok(token.BitwiseAnd, "&")

	case l.ch == '|':
		l.readChar()
		if l.ch == '|' {
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return tok(token.OrAssign, "||=")
			}
			return tok(token.Or, "||")
		}
		if l.ch == '=' {
			l.readChar()
			return tok(token.PipeAssign, "|=")
		}
		return tok(token.BitwiseOr, "|")

	case l.ch == '^':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return tok(token.CaretAssign, "^=")
		}
		return tok(token.BitwiseXor, "^")

	case l.ch == '?':
		l.readChar()
		if l.ch == '.' && !isDigit(l.peekChar()) {
			l.readChar()
			return tok(token.OptionalChain, "?.")
		}
		if l.ch == '?' {
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return tok(token.NullishAssign, "??=")
			}
			return tok(token.NullishCoalesce, "??")
		}
		return tok(token.QuestionMark, "?")

	case l.ch == '`':
		return l.readTemplateLiteral(line, col)

	case l.ch == '"' || l.ch == '\'':
		return l.readString(line, col)

	case isDigit(l.ch):
		return l.readNumber(line, col)

	case isIdentStart(l.ch):
		return l.readIdentifier(line, col)

	case l.ch == '\\' && l.peekChar() == 'u':
		return l.readIdentifier(line, col)

	default:
		ch := l.ch
		l.readChar()
		return tok(token.Illegal, string(ch))
	}
}

// NextTokenWithRegex is like NextToken but allows the caller to hint that
// a '/' should be interpreted as the start of a regex literal rather than division.
func (l *Lexer) NextTokenWithRegex(prevType token.TokenType) token.Token {
	l.skipWhitespaceAndComments()

	if l.ch == '/' && l.peekChar() != '/' && l.peekChar() != '*' && canPrecedeRegex(prevType) {
		line := l.line
		col := l.col
		return l.readRegExp(line, col)
	}
	return l.NextToken()
}

func (l *Lexer) readIdentifier(line, col int) token.Token {
	start := l.pos
	var buf strings.Builder
	hasEscape := false

	for isIdentPart(l.ch) || l.ch == '\\' {
		if l.ch == '\\' {
			hasEscape = true
			l.readChar() // consume backslash
			if l.ch != 'u' {
				return token.Token{Type: token.Illegal, Literal: "invalid escape in identifier", Line: line, Column: col}
			}
			l.readChar() // consume 'u'
			r := l.readUnicodeEscape()
			if r < 0 {
				return token.Token{Type: token.Illegal, Literal: "invalid unicode escape", Line: line, Column: col}
			}
			buf.WriteRune(rune(r))
		} else {
			buf.WriteRune(l.ch)
			l.readChar()
		}
	}

	var literal string
	if hasEscape {
		literal = buf.String()
	} else {
		literal = l.input[start:l.pos]
	}

	tt := token.LookupIdentifier(literal)
	return token.Token{Type: tt, Literal: literal, Line: line, Column: col}
}

// writeUTF16CodeUnit writes a UTF-16 code unit (including surrogates) to a string builder.
// For surrogates, it uses WTF-8 encoding (3-byte sequences like regular code points)
// rather than the replacement character that Go's WriteRune would produce.
func writeUTF16CodeUnit(buf *strings.Builder, cu uint16) {
	if cu < 0x80 {
		buf.WriteByte(byte(cu))
	} else if cu < 0x800 {
		buf.WriteByte(byte(0xC0 | (cu >> 6)))
		buf.WriteByte(byte(0x80 | (cu & 0x3F)))
	} else {
		// This works for surrogates too (0xD800-0xDFFF) using WTF-8 encoding
		buf.WriteByte(byte(0xE0 | (cu >> 12)))
		buf.WriteByte(byte(0x80 | ((cu >> 6) & 0x3F)))
		buf.WriteByte(byte(0x80 | (cu & 0x3F)))
	}
}

func (l *Lexer) readUnicodeEscape() int {
	if l.ch == '{' {
		// \u{XXXX} form
		l.readChar()
		val := 0
		digits := 0
		for l.ch != '}' && l.ch != 0 {
			d := hexVal(l.ch)
			if d < 0 {
				return -1
			}
			val = val*16 + d
			digits++
			l.readChar()
		}
		if l.ch != '}' || digits == 0 || val > 0x10FFFF {
			return -1
		}
		l.readChar() // consume '}'
		return val
	}
	// \uXXXX form (exactly 4 hex digits)
	val := 0
	for i := 0; i < 4; i++ {
		d := hexVal(l.ch)
		if d < 0 {
			return -1
		}
		val = val*16 + d
		l.readChar()
	}
	return val
}

func (l *Lexer) readString(line, col int) token.Token {
	quote := l.ch
	l.readChar() // skip opening quote
	var buf strings.Builder

	for l.ch != quote && l.ch != 0 && l.ch != '\n' {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case '\\':
				buf.WriteByte('\\')
			case '\'':
				buf.WriteByte('\'')
			case '"':
				buf.WriteByte('"')
			case '0', '1', '2', '3', '4', '5', '6', '7':
				// Octal escape sequence (Annex B, non-strict mode)
				val := int(l.ch - '0')
				l.readChar()
				if l.ch >= '0' && l.ch <= '7' {
					val = val*8 + int(l.ch-'0')
					l.readChar()
					if val <= 037 && l.ch >= '0' && l.ch <= '7' {
						val = val*8 + int(l.ch-'0')
						l.readChar()
					}
				}
				buf.WriteRune(rune(val))
				continue
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case 'v':
				buf.WriteByte('\v')
			case 'x':
				l.readChar()
				d1 := hexVal(l.ch)
				l.readChar()
				d2 := hexVal(l.ch)
				if d1 < 0 || d2 < 0 {
					return token.Token{Type: token.Illegal, Literal: "invalid hex escape", Line: line, Column: col}
				}
				buf.WriteRune(rune(d1*16 + d2))
			case 'u':
				l.readChar()
				r := l.readUnicodeEscape()
				if r < 0 {
					return token.Token{Type: token.Illegal, Literal: "invalid unicode escape", Line: line, Column: col}
				}
				// Handle surrogate pairs: if high surrogate followed by \uDCxx
				if r >= 0xD800 && r <= 0xDBFF && l.ch == '\\' && l.peekChar() == 'u' {
					// Save position to backtrack if not a low surrogate
					savedPos := l.pos
					savedReadPos := l.readPos
					savedCh := l.ch
					l.readChar() // skip '\'
					l.readChar() // skip 'u'
					r2 := l.readUnicodeEscape()
					if r2 >= 0xDC00 && r2 <= 0xDFFF {
						// Valid surrogate pair - combine
						combined := 0x10000 + (r-0xD800)*0x400 + (r2 - 0xDC00)
						buf.WriteRune(rune(combined))
					} else {
						// Not a low surrogate - write high surrogate as-is and backtrack
						writeUTF16CodeUnit(&buf, uint16(r))
						// Restore position
						l.pos = savedPos
						l.readPos = savedReadPos
						l.ch = savedCh
					}
				} else if r >= 0xD800 && r <= 0xDFFF {
					// Lone surrogate - write as raw bytes
					writeUTF16CodeUnit(&buf, uint16(r))
				} else {
					buf.WriteRune(rune(r))
				}
				continue // readUnicodeEscape already advanced past the escape
			case '\n':
				l.line++
				l.col = 0
				// line continuation - don't add to string
			case '\r':
				if l.peekChar() == '\n' {
					l.readChar()
				}
				l.line++
				l.col = 0
			default:
				buf.WriteRune(l.ch)
			}
			l.readChar()
			continue
		}
		buf.WriteRune(l.ch)
		l.readChar()
	}

	if l.ch != quote {
		return token.Token{Type: token.Illegal, Literal: "unterminated string", Line: line, Column: col}
	}
	l.readChar() // skip closing quote
	return token.Token{Type: token.String, Literal: buf.String(), Line: line, Column: col}
}

func (l *Lexer) readNumber(line, col int) token.Token {
	start := l.pos

	if l.ch == '0' {
		next := l.peekChar()
		switch {
		case next == 'x' || next == 'X':
			l.readChar() // 0
			l.readChar() // x
			if !isHexDigit(l.ch) {
				return token.Token{Type: token.Illegal, Literal: "invalid hex literal", Line: line, Column: col}
			}
			for isHexDigit(l.ch) || l.ch == '_' {
				l.readChar()
			}
			return token.Token{Type: token.Number, Literal: l.input[start:l.pos], Line: line, Column: col}

		case next == 'o' || next == 'O':
			l.readChar() // 0
			l.readChar() // o
			if !isOctalDigit(l.ch) {
				return token.Token{Type: token.Illegal, Literal: "invalid octal literal", Line: line, Column: col}
			}
			for isOctalDigit(l.ch) || l.ch == '_' {
				l.readChar()
			}
			return token.Token{Type: token.Number, Literal: l.input[start:l.pos], Line: line, Column: col}

		case next == 'b' || next == 'B':
			l.readChar() // 0
			l.readChar() // b
			if l.ch != '0' && l.ch != '1' {
				return token.Token{Type: token.Illegal, Literal: "invalid binary literal", Line: line, Column: col}
			}
			for l.ch == '0' || l.ch == '1' || l.ch == '_' {
				l.readChar()
			}
			return token.Token{Type: token.Number, Literal: l.input[start:l.pos], Line: line, Column: col}
		}
	}

	// Decimal: integer part
	l.readDecimalDigits()

	// Fractional part
	if l.ch == '.' {
		l.readChar()
		l.readDecimalDigits()
	}

	// Exponent
	if l.ch == 'e' || l.ch == 'E' {
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		l.readDecimalDigits()
	}

	// BigInt suffix
	if l.ch == 'n' {
		l.readChar()
	}

	return token.Token{Type: token.Number, Literal: l.input[start:l.pos], Line: line, Column: col}
}

func (l *Lexer) readDecimalDigits() {
	for isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
}

func (l *Lexer) readTemplateLiteral(line, col int) token.Token {
	l.readChar() // skip opening backtick
	var buf strings.Builder

	for {
		if l.ch == 0 {
			return token.Token{Type: token.Illegal, Literal: "unterminated template literal", Line: line, Column: col}
		}
		if l.ch == '`' {
			l.readChar()
			return token.Token{Type: token.NoSubstitutionTemplate, Literal: buf.String(), Line: line, Column: col}
		}
		if l.ch == '$' && l.peekChar() == '{' {
			l.readChar() // skip $
			l.readChar() // skip {
			l.templateStack = append(l.templateStack, l.braceDepth)
			l.braceDepth++
			return token.Token{Type: token.TemplateHead, Literal: buf.String(), Line: line, Column: col}
		}
		if l.ch == '\\' {
			l.readChar()
			l.readTemplateEscape(&buf)
			continue
		}
		if l.ch == '\n' {
			l.line++
			l.col = 0
		}
		buf.WriteRune(l.ch)
		l.readChar()
	}
}

func (l *Lexer) readTemplateContinuation(line, col int) token.Token {
	l.readChar() // skip closing }
	l.braceDepth--
	var buf strings.Builder

	for {
		if l.ch == 0 {
			return token.Token{Type: token.Illegal, Literal: "unterminated template literal", Line: line, Column: col}
		}
		if l.ch == '`' {
			l.readChar()
			return token.Token{Type: token.TemplateTail, Literal: buf.String(), Line: line, Column: col}
		}
		if l.ch == '$' && l.peekChar() == '{' {
			l.readChar() // skip $
			l.readChar() // skip {
			l.templateStack = append(l.templateStack, l.braceDepth)
			l.braceDepth++
			return token.Token{Type: token.TemplateMiddle, Literal: buf.String(), Line: line, Column: col}
		}
		if l.ch == '\\' {
			l.readChar()
			l.readTemplateEscape(&buf)
			continue
		}
		if l.ch == '\n' {
			l.line++
			l.col = 0
		}
		buf.WriteRune(l.ch)
		l.readChar()
	}
}

func (l *Lexer) readTemplateEscape(buf *strings.Builder) {
	switch l.ch {
	case 'n':
		buf.WriteByte('\n')
		l.readChar()
	case 'r':
		buf.WriteByte('\r')
		l.readChar()
	case 't':
		buf.WriteByte('\t')
		l.readChar()
	case '\\':
		buf.WriteByte('\\')
		l.readChar()
	case '`':
		buf.WriteByte('`')
		l.readChar()
	case '$':
		buf.WriteByte('$')
		l.readChar()
	case '0', '1', '2', '3', '4', '5', '6', '7':
		// Octal escape sequence (Annex B, non-strict mode)
		val := int(l.ch - '0')
		l.readChar()
		if l.ch >= '0' && l.ch <= '7' {
			val = val*8 + int(l.ch-'0')
			l.readChar()
			if val <= 037 && l.ch >= '0' && l.ch <= '7' {
				val = val*8 + int(l.ch-'0')
				l.readChar()
			}
		}
		buf.WriteRune(rune(val))
		return
	case '\n':
		l.line++
		l.col = 0
		l.readChar()
	default:
		buf.WriteByte('\\')
		buf.WriteRune(l.ch)
		l.readChar()
	}
}

func (l *Lexer) readRegExp(line, col int) token.Token {
	var buf strings.Builder
	buf.WriteByte('/')
	l.readChar() // skip opening /

	inCharClass := false
	for {
		if (l.ch == 0 && l.pos >= len(l.input)) || l.ch == '\n' || l.ch == '\r' {
			return token.Token{Type: token.Illegal, Literal: "unterminated regexp", Line: line, Column: col}
		}
		if l.ch == '\\' {
			buf.WriteRune(l.ch)
			l.readChar()
			if (l.ch == 0 && l.pos >= len(l.input)) || l.ch == '\n' || l.ch == '\r' {
				return token.Token{Type: token.Illegal, Literal: "unterminated regexp", Line: line, Column: col}
			}
			buf.WriteRune(l.ch)
			l.readChar()
			continue
		}
		if l.ch == '[' {
			inCharClass = true
		} else if l.ch == ']' {
			inCharClass = false
		}
		if l.ch == '/' && !inCharClass {
			buf.WriteByte('/')
			l.readChar()
			break
		}
		buf.WriteRune(l.ch)
		l.readChar()
	}

	// Read flags
	for isIdentPart(l.ch) {
		buf.WriteRune(l.ch)
		l.readChar()
	}

	return token.Token{Type: token.RegExp, Literal: buf.String(), Line: line, Column: col}
}

// Tokenize returns all tokens from the input. It uses context-aware regex detection.
func Tokenize(input string) []token.Token {
	l := New(input)
	var tokens []token.Token
	prevType := token.EOF // EOF means "start of input" - regex is valid here

	for {
		tok := l.NextTokenWithRegex(prevType)
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
		prevType = tok.Type
	}
	return tokens
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isOctalDigit(ch rune) bool {
	return ch >= '0' && ch <= '7'
}

func isIdentStart(ch rune) bool {
	return ch == '_' || ch == '$' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch > 127 && unicode.IsLetter(ch))
}

func isIdentPart(ch rune) bool {
	return isIdentStart(ch) || isDigit(ch) || ch == '\u200C' || ch == '\u200D'
}

func hexVal(ch rune) int {
	switch {
	case ch >= '0' && ch <= '9':
		return int(ch - '0')
	case ch >= 'a' && ch <= 'f':
		return int(ch-'a') + 10
	case ch >= 'A' && ch <= 'F':
		return int(ch-'A') + 10
	default:
		return -1
	}
}
