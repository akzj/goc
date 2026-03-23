// Package lexer provides lexical analysis for C11 source code.
// This file implements the Lexer struct and tokenization methods.
package lexer

import (
	"unicode/utf8"
)

// Lexer tokenizes C11 source code into a stream of tokens.
type Lexer struct {
	source   string    // Source code to tokenize
	pos      int       // Current position in source (byte offset)
	line     int       // Current line number (1-based)
	column   int       // Current column number (1-based, in runes)
	fileName string    // Source file name (for error reporting)
	tokens   []Token   // Tokenized tokens
	hasSpace bool      // Whether next token is preceded by whitespace
}

// NewLexer creates a new lexer for the given source code.
func NewLexer(source string, fileName string) *Lexer {
	return &Lexer{
		source:   source,
		pos:      0,
		line:     1,
		column:   1,
		fileName: fileName,
		tokens:   make([]Token, 0),
		hasSpace: false,
	}
}

// Tokenize tokenizes the entire source code and returns a slice of tokens.
func (l *Lexer) Tokenize() []Token {
	for {
		token := l.nextToken()
		l.tokens = append(l.tokens, token)
		
		if token.Type == EOF {
			break
		}
	}
	return l.tokens
}

// nextToken returns the next token from the source.
func (l *Lexer) nextToken() Token {
	// Skip whitespace and comments
	l.skipWhitespaceAndComments()
	
	// Check for EOF
	if l.atEnd() {
		return l.makeToken(EOF, "")
	}
	
	// Record position before reading
	startPos := l.pos
	startLine := l.line
	startColumn := l.column
	hasSpace := l.hasSpace
	l.hasSpace = false
	
	// Read the next character
	ch := l.advance()
	
	// Handle different token types
	switch {
	// Identifiers and keywords
	case isAlpha(ch) || ch == '_':
		return l.scanIdentifier(startPos, startLine, startColumn, hasSpace)
	
	// Numeric literals
	case isDigit(ch):
		return l.scanNumber(startPos, startLine, startColumn, hasSpace, ch)
	
	// Character literals
	case ch == '\'':
		return l.scanCharLiteral(startPos, startLine, startColumn, hasSpace)
	
	// String literals
	case ch == '"':
		return l.scanStringLiteral(startPos, startLine, startColumn, hasSpace)
	
	// Operators and punctuation
	default:
		return l.scanOperator(startPos, startLine, startColumn, hasSpace, ch)
	}
}

// skipWhitespaceAndComments skips whitespace and comments.
func (l *Lexer) skipWhitespaceAndComments() {
	for !l.atEnd() {
		ch := l.peek()
		
		switch ch {
		// Whitespace
		case ' ', '\t', '\r', '\v', '\f':
			l.advance()
			l.hasSpace = true
		
		// Newline
		case '\n':
			l.advance()
			l.line++
			l.column = 1
			l.hasSpace = true
		
		// Comments
		case '/':
			next := l.peekNext()
			if next == '/' {
				// Single-line comment
				l.skipSingleLineComment()
			} else if next == '*' {
				// Multi-line comment
				l.skipMultiLineComment()
			} else {
				return
			}
		
		default:
			return
		}
	}
}

// skipSingleLineComment skips a single-line comment (// ...).
func (l *Lexer) skipSingleLineComment() {
	// Skip //
	l.advance()
	l.advance()
	
	// Skip until end of line or EOF
	for !l.atEnd() && l.peek() != '\n' {
		l.advance()
	}
}

// skipMultiLineComment skips a multi-line comment (/* ... */).
func (l *Lexer) skipMultiLineComment() {
	// Skip /*
	l.advance()
	l.advance()
	
	// Skip until */
	for !l.atEnd() {
		ch := l.peek()
		if ch == '*' && l.peekNext() == '/' {
			l.advance() // *
			l.advance() // /
			return
		}
		if ch == '\n' {
			l.line++
			l.column = 1
		}
		l.advance()
	}
}

// scanIdentifier scans an identifier or keyword.
func (l *Lexer) scanIdentifier(startPos, startLine, startColumn int, hasSpace bool) Token {
	// Continue reading while we have alphanumeric or underscore
	for !l.atEnd() && (isAlphaNumeric(l.peek()) || l.peek() == '_') {
		l.advance()
	}
	
	// Extract the identifier
	value := l.source[startPos:l.pos]
	
	// Check if it's a keyword
	tokenType := IDENT
	if kw, ok := Keywords[value]; ok {
		tokenType = kw
	}
	
	return Token{
		Type:     tokenType,
		Value:    value,
		Pos:      Position{File: l.fileName, Line: startLine, Column: startColumn},
		Raw:      value,
		HasSpace: hasSpace,
	}
}

// scanNumber scans a numeric literal (integer or float).
func (l *Lexer) scanNumber(startPos, startLine, startColumn int, hasSpace bool, firstChar rune) Token {
	value := string(firstChar)
	isFloat := false
	isHex := false
	
	// Check for hex prefix
	if firstChar == '0' && (l.peek() == 'x' || l.peek() == 'X') {
		value += string(l.advance()) // x or X
		isHex = true
		
		// Read hex digits (integer part)
		for !l.atEnd() && isHexDigit(l.peek()) {
			value += string(l.advance())
		}
		
		// Check for decimal point in hex float
		if l.peek() == '.' {
			isFloat = true
			value += string(l.advance())
			
			// Read fractional part (hex digits)
			for !l.atEnd() && isHexDigit(l.peek()) {
				value += string(l.advance())
			}
		}
		
		// Check for hex float exponent
		if l.peek() == 'p' || l.peek() == 'P' {
			isFloat = true
			value += string(l.advance())
			
			// Optional sign
			if l.peek() == '+' || l.peek() == '-' {
				value += string(l.advance())
			}
			
			// Read exponent digits
			for !l.atEnd() && isDigit(l.peek()) {
				value += string(l.advance())
			}
		}
	} else if firstChar == '0' && (l.peek() == 'b' || l.peek() == 'B') {
		// Binary literal (C23, but commonly supported)
		value += string(l.advance()) // b or B
		
		// Read binary digits
		for !l.atEnd() && (l.peek() == '0' || l.peek() == '1') {
			value += string(l.advance())
		}
	} else {
		// Decimal or octal
		// Read integer part
		for !l.atEnd() && isDigit(l.peek()) {
			value += string(l.advance())
		}
		
		// Check for decimal point
		if l.peek() == '.' {
			isFloat = true
			value += string(l.advance())
			
			// Read fractional part
			for !l.atEnd() && isDigit(l.peek()) {
				value += string(l.advance())
			}
		}
		
		// Check for exponent
		if !isHex && (l.peek() == 'e' || l.peek() == 'E') {
			isFloat = true
			value += string(l.advance())
			
			// Optional sign
			if l.peek() == '+' || l.peek() == '-' {
				value += string(l.advance())
			}
			
			// Read exponent digits
			for !l.atEnd() && isDigit(l.peek()) {
				value += string(l.advance())
			}
		}
	}
	
	// Read integer or float suffix
	for !l.atEnd() {
		ch := l.peek()
		if ch == 'u' || ch == 'U' || ch == 'l' || ch == 'L' || 
		   ch == 'f' || ch == 'F' {
			value += string(l.advance())
		} else {
			break
		}
	}
	
	tokenType := INT_LIT
	if isFloat {
		tokenType = FLOAT_LIT
	}
	
	return Token{
		Type:     tokenType,
		Value:    value,
		Pos:      Position{File: l.fileName, Line: startLine, Column: startColumn},
		Raw:      value,
		HasSpace: hasSpace,
	}
}

// scanCharLiteral scans a character literal.
func (l *Lexer) scanCharLiteral(startPos, startLine, startColumn int, hasSpace bool) Token {
	value := "'"
	
	// Scan the character(s)
	for !l.atEnd() && l.peek() != '\'' {
		ch := l.peek()
		if ch == '\\' {
			// Escape sequence
			value += string(l.advance())
			if !l.atEnd() {
				value += string(l.advance())
			}
		} else if ch == '\n' {
			// Newline in character literal (error, but handle gracefully)
			break
		} else {
			value += string(l.advance())
		}
	}
	
	// Closing quote
	if l.peek() == '\'' {
		value += string(l.advance())
	}
	
	return Token{
		Type:     CHAR_LIT,
		Value:    value,
		Pos:      Position{File: l.fileName, Line: startLine, Column: startColumn},
		Raw:      value,
		HasSpace: hasSpace,
	}
}

// scanStringLiteral scans a string literal.
func (l *Lexer) scanStringLiteral(startPos, startLine, startColumn int, hasSpace bool) Token {
	value := "\""
	
	// Scan the string content
	for !l.atEnd() && l.peek() != '"' {
		ch := l.peek()
		if ch == '\\' {
			// Escape sequence
			value += string(l.advance())
			if !l.atEnd() {
				value += string(l.advance())
			}
		} else if ch == '\n' {
			// Newline in string literal (error, but handle gracefully)
			break
		} else {
			value += string(l.advance())
		}
	}
	
	// Closing quote
	if l.peek() == '"' {
		value += string(l.advance())
	}
	
	return Token{
		Type:     STRING_LIT,
		Value:    value,
		Pos:      Position{File: l.fileName, Line: startLine, Column: startColumn},
		Raw:      value,
		HasSpace: hasSpace,
	}
}

// scanOperator scans an operator or punctuation.
func (l *Lexer) scanOperator(startPos, startLine, startColumn int, hasSpace bool, firstChar rune) Token {
	value := string(firstChar)
	tokenType := ILLEGAL
	
	// Try to match multi-character operators first
	switch firstChar {
	case '+':
		if l.peek() == '+' {
			value += string(l.advance())
			tokenType = INC
		} else if l.peek() == '=' {
			value += string(l.advance())
			tokenType = ADD_ASSIGN
		} else {
			tokenType = ADD
		}
	
	case '-':
		if l.peek() == '-' {
			value += string(l.advance())
			tokenType = DEC
		} else if l.peek() == '=' {
			value += string(l.advance())
			tokenType = SUB_ASSIGN
		} else if l.peek() == '>' {
			value += string(l.advance())
			tokenType = ARROW
		} else {
			tokenType = SUB
		}
	
	case '*':
		if l.peek() == '=' {
			value += string(l.advance())
			tokenType = MUL_ASSIGN
		} else {
			tokenType = MUL
		}
	
	case '/':
		if l.peek() == '=' {
			value += string(l.advance())
			tokenType = QUO_ASSIGN
		} else {
			tokenType = QUO
		}
	
	case '%':
		if l.peek() == '=' {
			value += string(l.advance())
			tokenType = REM_ASSIGN
		} else if l.peek() == '>' {
			// %>} (alternative token for })
			value += string(l.advance())
			tokenType = RBRACE
		} else if l.peek() == ':' {
			// %: (alternative token for #)
			value += string(l.advance())
			if l.peek() == '%' && l.peekNext() == ':' {
				// %:%: (alternative token for ##)
				value += string(l.advance())
				value += string(l.advance())
				tokenType = CONCAT
			} else {
				tokenType = PREPROCESSOR
			}
		} else {
			tokenType = REM
		}
	
	case '&':
		if l.peek() == '&' {
			value += string(l.advance())
			tokenType = LAND
		} else if l.peek() == '=' {
			value += string(l.advance())
			tokenType = AND_ASSIGN
		} else {
			tokenType = AND
		}
	
	case '|':
		if l.peek() == '|' {
			value += string(l.advance())
			tokenType = LOR
		} else if l.peek() == '=' {
			value += string(l.advance())
			tokenType = OR_ASSIGN
		} else {
			tokenType = OR
		}
	
	case '^':
		if l.peek() == '=' {
			value += string(l.advance())
			tokenType = XOR_ASSIGN
		} else {
			tokenType = XOR
		}
	
	case '=':
		if l.peek() == '=' {
			value += string(l.advance())
			tokenType = EQL
		} else {
			tokenType = ASSIGN
		}
	
	case '!':
		if l.peek() == '=' {
			value += string(l.advance())
			tokenType = NEQ
		} else {
			tokenType = NOT
		}
	
	case '<':
		if l.peek() == '<' {
			value += string(l.advance())
			if l.peek() == '=' {
				value += string(l.advance())
				tokenType = SHL_ASSIGN
			} else {
				tokenType = SHL
			}
		} else if l.peek() == '=' {
			value += string(l.advance())
			tokenType = LEQ
		} else if l.peek() == ':' {
			// <: (alternative token for [)
			value += string(l.advance())
			tokenType = LBRACK
		} else if l.peek() == '%' {
			// <% (alternative token for {)
			value += string(l.advance())
			tokenType = LBRACE
		} else {
			tokenType = LSS
		}
	
	case '>':
		if l.peek() == '>' {
			value += string(l.advance())
			if l.peek() == '=' {
				value += string(l.advance())
				tokenType = SHR_ASSIGN
			} else {
				tokenType = SHR
			}
		} else if l.peek() == '=' {
			value += string(l.advance())
			tokenType = GEQ
		} else {
			tokenType = GTR
		}
	
	case '.':
		if l.peek() == '.' && l.peekNext() == '.' {
			value += string(l.advance())
			value += string(l.advance())
			tokenType = ELLIPSIS
		} else {
			tokenType = DOT
		}
	
	case '#':
		if l.peek() == '#' {
			value += string(l.advance())
			tokenType = CONCAT
		} else {
			tokenType = PREPROCESSOR
		}
	
	case ':':
		if l.peek() == '>' {
			// :> (alternative token for ])
			value += string(l.advance())
			tokenType = RBRACK
		} else {
			tokenType = COLON
		}
	
	// Simple single-character operators and punctuation
	case '(':
		tokenType = LPAREN
	case ')':
		tokenType = RPAREN
	case '[':
		tokenType = LBRACK
	case ']':
		tokenType = RBRACK
	case '{':
		tokenType = LBRACE
	case '}':
		tokenType = RBRACE
	case ',':
		tokenType = COMMA
	case ';':
		tokenType = SEMICOLON
	case '?':
		tokenType = QUESTION
	case '~':
		tokenType = BITNOT
	}
	
	return Token{
		Type:     tokenType,
		Value:    value,
		Pos:      Position{File: l.fileName, Line: startLine, Column: startColumn},
		Raw:      value,
		HasSpace: hasSpace,
	}
}

// Helper methods

// atEnd returns true if we've reached the end of the source.
func (l *Lexer) atEnd() bool {
	return l.pos >= len(l.source)
}

// peek returns the next character without consuming it.
func (l *Lexer) peek() rune {
	if l.atEnd() {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.source[l.pos:])
	return r
}

// peekNext returns the character after the next without consuming it.
func (l *Lexer) peekNext() rune {
	if l.pos >= len(l.source) {
		return 0
	}
	_, size := utf8.DecodeRuneInString(l.source[l.pos:])
	if l.pos+size >= len(l.source) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.source[l.pos+size:])
	return r
}

// advance consumes and returns the next character.
func (l *Lexer) advance() rune {
	if l.atEnd() {
		return 0
	}
	r, size := utf8.DecodeRuneInString(l.source[l.pos:])
	l.pos += size
	l.column++
	return r
}

// makeToken creates a token with the given type and value.
func (l *Lexer) makeToken(tokenType TokenType, value string) Token {
	return Token{
		Type:     tokenType,
		Value:    value,
		Pos:      Position{File: l.fileName, Line: l.line, Column: l.column},
		Raw:      value,
		HasSpace: l.hasSpace,
	}
}

// Character classification functions

// isAlpha returns true if the character is alphabetic.
func isAlpha(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// isDigit returns true if the character is a digit.
func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

// isAlphaNumeric returns true if the character is alphanumeric.
func isAlphaNumeric(ch rune) bool {
	return isAlpha(ch) || isDigit(ch)
}

// isHexDigit returns true if the character is a hexadecimal digit.
func isHexDigit(ch rune) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// TokenizeString is a convenience function that tokenizes a string of C source code.
func TokenizeString(source string) []Token {
	lexer := NewLexer(source, "")
	return lexer.Tokenize()
}