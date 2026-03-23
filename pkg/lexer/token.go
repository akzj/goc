// Package lexer provides lexical analysis for C11 source code.
// This file defines all token types and their structures for the C11 lexer.
package lexer

import (
	"fmt"
	"strings"
)

// Position represents a position in source code for error reporting and debugging.
type Position struct {
	File   string // Source file name (may be empty for stdin)
	Line   int    // Line number (1-based)
	Column int    // Column number (1-based, counts Unicode code points)
}

// String returns a formatted position string.
func (p Position) String() string {
	if p.File != "" {
		return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
	}
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// IsValid returns true if the position is valid (line > 0).
func (p Position) IsValid() bool {
	return p.Line > 0
}

// Token represents a lexical token in C source code.
type Token struct {
	Type     TokenType // The type of the token
	Value    string    // The string value of the token
	Pos      Position  // Position in source code
	Raw      string    // Raw source text (includes whitespace/comments for some tokens)
	HasSpace bool      // Whether the token was preceded by whitespace
}

// String returns a formatted token string for debugging.
func (t Token) String() string {
	if t.Value != "" && string(t.Type) != t.Value {
		return fmt.Sprintf("%s(%q) @ %s", t.Type, t.Value, t.Pos)
	}
	return fmt.Sprintf("%s @ %s", t.Type, t.Pos)
}

// IsKeyword returns true if the token is a C11 keyword.
func (t Token) IsKeyword() bool {
	return IsKeyword(t.Type)
}

// IsIdentifier returns true if the token is an identifier.
func (t Token) IsIdentifier() bool {
	return t.Type == IDENT
}

// IsLiteral returns true if the token is a literal (int, float, char, string).
func (t Token) IsLiteral() bool {
	return IsLiteral(t.Type)
}

// IsOperator returns true if the token is an operator.
func (t Token) IsOperator() bool {
	return IsOperator(t.Type)
}

// IsPunctuation returns true if the token is punctuation.
func (t Token) IsPunctuation() bool {
	return IsPunctuation(t.Type)
}

// TokenType represents the type of a lexical token.
type TokenType string

// Token type constants for C11 lexer.
// Organized by category: keywords, operators, literals, identifiers, special.
const (
	// Special tokens
	ILLEGAL TokenType = "ILLEGAL" // Illegal character or token
	EOF     TokenType = "EOF"     // End of file
	COMMENT TokenType = "COMMENT" // Comment (preserved for preprocessing)
	NEWLINE TokenType = "NEWLINE" // Newline (for preprocessor)

	// Identifiers
	IDENT TokenType = "IDENT" // Identifier (e.g., variable names, function names)

	// Literals
	INT_LIT    TokenType = "INT_LIT"    // Integer literal (decimal, octal, hex, binary)
	FLOAT_LIT  TokenType = "FLOAT_LIT"   // Floating-point literal
	CHAR_LIT   TokenType = "CHAR_LIT"    // Character literal
	STRING_LIT TokenType = "STRING_LIT" // String literal

	// C11 Keywords (37 keywords in C11)
	// Storage class specifiers
	AUTO     TokenType = "auto"
	REGISTER TokenType = "register"
	STATIC   TokenType = "static"
	EXTERN   TokenType = "extern"
	TYPEDEF  TokenType = "typedef"

	// Type qualifiers
	CONST    TokenType = "const"
	VOLATILE TokenType = "volatile"
	RESTRICT TokenType = "restrict"

	// Type specifiers
	VOID      TokenType = "void"
	CHAR      TokenType = "char"
	SHORT     TokenType = "short"
	INT       TokenType = "int"
	LONG      TokenType = "long"
	FLOAT     TokenType = "float"
	DOUBLE    TokenType = "double"
	SIGNED    TokenType = "signed"
	UNSIGNED  TokenType = "unsigned"
	COMPLEX  TokenType = "_Complex"  // C99/C11
	IMAGINARY TokenType = "_Imaginary" // C99/C11
	BOOL     TokenType = "_Bool"     // C99/C11
	ATOMIC   TokenType = "_Atomic"   // C11

	// Structure/union/enum specifiers
	STRUCT TokenType = "struct"
	UNION  TokenType = "union"
	ENUM   TokenType = "enum"

	// Control flow keywords
	IF       TokenType = "if"
	ELSE     TokenType = "else"
	SWITCH   TokenType = "switch"
	CASE     TokenType = "case"
	DEFAULT  TokenType = "default"
	WHILE    TokenType = "while"
	DO       TokenType = "do"
	FOR      TokenType = "for"
	GOTO     TokenType = "goto"
	CONTINUE TokenType = "continue"
	BREAK    TokenType = "break"
	RETURN   TokenType = "return"

	// Function specifiers
	INLINE     TokenType = "inline"     // C99/C11
	NORETURN  TokenType = "_Noreturn"  // C11
	THREAD_LOCAL TokenType = "_Thread_local" // C11

	// Alignment specifiers (C11)
	ALIGNAS TokenType = "_Alignas" // C11
	ALIGNOF TokenType = "_Alignof" // C11

	// Generic selection (C11)
	GENERIC TokenType = "_Generic" // C11

	// Static assertion (C11)
	STATIC_ASSERT TokenType = "_Static_assert" // C11

	// Other keywords
	SIZEOF TokenType = "sizeof"

	// Operators
	// Arithmetic operators
	ADD TokenType = "+"  // Addition
	SUB TokenType = "-"  // Subtraction
	MUL TokenType = "*"  // Multiplication/Dereference
	QUO TokenType = "/"  // Division
	REM TokenType = "%"  // Modulus

	// Bitwise operators
	AND     TokenType = "&"  // Bitwise AND/Address-of
	OR      TokenType = "|"  // Bitwise OR
	XOR     TokenType = "^"  // Bitwise XOR
	SHL     TokenType = "<<" // Left shift
	SHR     TokenType = ">>" // Right shift
	BITNOT  TokenType = "~"  // Bitwise NOT

	// Logical operators
	LAND TokenType = "&&" // Logical AND
	LOR  TokenType = "||" // Logical OR
	NOT  TokenType = "!"  // Logical NOT

	// Comparison operators
	EQL TokenType = "==" // Equal
	NEQ TokenType = "!=" // Not equal
	LSS TokenType = "<"  // Less than
	LEQ TokenType = "<=" // Less than or equal
	GTR TokenType = ">"  // Greater than
	GEQ TokenType = ">=" // Greater than or equal

	// Assignment operators
	ASSIGN    TokenType = "="  // Simple assignment
	ADD_ASSIGN TokenType = "+=" // Addition assignment
	SUB_ASSIGN TokenType = "-=" // Subtraction assignment
	MUL_ASSIGN TokenType = "*=" // Multiplication assignment
	QUO_ASSIGN TokenType = "/=" // Division assignment
	REM_ASSIGN TokenType = "%=" // Modulus assignment
	AND_ASSIGN TokenType = "&=" // Bitwise AND assignment
	OR_ASSIGN  TokenType = "|=" // Bitwise OR assignment
	XOR_ASSIGN TokenType = "^=" // Bitwise XOR assignment
	SHL_ASSIGN TokenType = "<<=" // Left shift assignment
	SHR_ASSIGN TokenType = ">>=" // Right shift assignment

	// Increment/Decrement
	INC TokenType = "++" // Increment
	DEC TokenType = "--" // Decrement

	// Member access
	ARROW TokenType = "->" // Pointer member access
	DOT   TokenType = "."  // Structure member access

	// Conditional operator
	QUESTION TokenType = "?" // Ternary conditional
	COLON    TokenType = ":" // Ternary conditional/label

	// Punctuation
	LPAREN   TokenType = "("  // Left parenthesis
	RPAREN   TokenType = ")"  // Right parenthesis
	LBRACK   TokenType = "["  // Left bracket
	RBRACK   TokenType = "]"  // Right bracket
	LBRACE   TokenType = "{"  // Left brace
	RBRACE   TokenType = "}"  // Right brace
	COMMA    TokenType = ","  // Comma
	SEMICOLON TokenType = ";" // Semicolon
	ELLIPSIS TokenType = "..." // Variadic function parameter

	// Preprocessor directives (treated as tokens for preprocessor)
	PREPROCESSOR TokenType = "#" // Preprocessor directive start
	STRINGIFY    TokenType = "#" // Stringify operator (in macros)
	CONCAT       TokenType = "##" // Token concatenation operator (in macros)

	// Special C11 tokens
	HEADER_NAME TokenType = "HEADER_NAME" // <header> or "header" in #include
	PRAGMA_OP   TokenType = "PRAGMA_OP"     // PRAGMA_OP operator (C99/C11)
)

// Keywords maps keyword strings to their token types.
var Keywords = map[string]TokenType{
	// Storage class specifiers
	"auto":     AUTO,
	"register": REGISTER,
	"static":   STATIC,
	"extern":   EXTERN,
	"typedef":  TYPEDEF,

	// Type qualifiers
	"const":    CONST,
	"volatile": VOLATILE,
	"restrict": RESTRICT,

	// Type specifiers
	"void":      VOID,
	"char":      CHAR,
	"short":     SHORT,
	"int":       INT,
	"long":      LONG,
	"float":     FLOAT,
	"double":    DOUBLE,
	"signed":    SIGNED,
	"unsigned":  UNSIGNED,
	"_Complex":  COMPLEX,
	"_Imaginary": IMAGINARY,
	"_Bool":     BOOL,
	"_Atomic":   ATOMIC,

	// Structure/union/enum specifiers
	"struct": STRUCT,
	"union":  UNION,
	"enum":   ENUM,

	// Control flow keywords
	"if":       IF,
	"else":     ELSE,
	"switch":   SWITCH,
	"case":     CASE,
	"default":  DEFAULT,
	"while":    WHILE,
	"do":       DO,
	"for":      FOR,
	"goto":     GOTO,
	"continue": CONTINUE,
	"break":    BREAK,
	"return":   RETURN,

	// Function specifiers
	"inline":       INLINE,
	"_Noreturn":    NORETURN,
	"_Thread_local": THREAD_LOCAL,

	// Alignment specifiers (C11)
	"_Alignas": ALIGNAS,
	"_Alignof": ALIGNOF,

	// Generic selection (C11)
	"_Generic": GENERIC,

	// Static assertion (C11)
	"_Static_assert": STATIC_ASSERT,

	// Other keywords
	"sizeof": SIZEOF,
	"PRAGMA_OP": PRAGMA_OP,
}

// IsKeyword returns true if the token type is a C11 keyword.
func IsKeyword(t TokenType) bool {
	_, ok := Keywords[string(t)]
	return ok
}

// IsKeywordString returns true if the string is a C11 keyword.
func IsKeywordString(s string) bool {
	_, ok := Keywords[s]
	return ok
}

// LookupKeyword returns the token type for a keyword string, or IDENT if not a keyword.
func LookupKeyword(s string) TokenType {
	if tok, ok := Keywords[s]; ok {
		return tok
	}
	return IDENT
}

// IsLiteral returns true if the token type is a literal.
func IsLiteral(t TokenType) bool {
	switch t {
	case INT_LIT, FLOAT_LIT, CHAR_LIT, STRING_LIT:
		return true
	default:
		return false
	}
}

// IsOperator returns true if the token type is an operator.
func IsOperator(t TokenType) bool {
	switch t {
	// Arithmetic
	case ADD, SUB, MUL, QUO, REM:
		return true
	// Bitwise
	case AND, OR, XOR, SHL, SHR, BITNOT:
		return true
	// Logical
	case LAND, LOR, NOT:
		return true
	// Comparison
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		return true
	// Assignment
	case ASSIGN, ADD_ASSIGN, SUB_ASSIGN, MUL_ASSIGN, QUO_ASSIGN, REM_ASSIGN,
		AND_ASSIGN, OR_ASSIGN, XOR_ASSIGN, SHL_ASSIGN, SHR_ASSIGN:
		return true
	// Increment/Decrement
	case INC, DEC:
		return true
	// Member access
	case ARROW, DOT:
		return true
	// Conditional
	case QUESTION, COLON:
		return true
	default:
		return false
	}
}

// IsPunctuation returns true if the token type is punctuation.
func IsPunctuation(t TokenType) bool {
	switch t {
	case LPAREN, RPAREN, LBRACK, RBRACK, LBRACE, RBRACE,
		COMMA, SEMICOLON, ELLIPSIS:
		return true
	default:
		return false
	}
}

// IsAssignmentOperator returns true if the token type is an assignment operator.
func IsAssignmentOperator(t TokenType) bool {
	switch t {
	case ASSIGN, ADD_ASSIGN, SUB_ASSIGN, MUL_ASSIGN, QUO_ASSIGN, REM_ASSIGN,
		AND_ASSIGN, OR_ASSIGN, XOR_ASSIGN, SHL_ASSIGN, SHR_ASSIGN:
		return true
	default:
		return false
	}
}

// IsBinaryOperator returns true if the token type is a binary operator.
func IsBinaryOperator(t TokenType) bool {
	switch t {
	// Arithmetic
	case ADD, SUB, MUL, QUO, REM:
		return true
	// Bitwise
	case AND, OR, XOR, SHL, SHR:
		return true
	// Logical
	case LAND, LOR:
		return true
	// Comparison
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		return true
	// Comma operator
	case COMMA:
		return true
	default:
		return false
	}
}

// IsUnaryOperator returns true if the token type can be a unary operator.
func IsUnaryOperator(t TokenType) bool {
	switch t {
	case ADD, SUB, MUL, AND, NOT, BITNOT, INC, DEC, SIZEOF:
		return true
	default:
		return false
	}
}

// OperatorPrecedence returns the precedence of an operator (higher = tighter binding).
// Returns -1 if not an operator.
func OperatorPrecedence(t TokenType) int {
	switch t {
	case COMMA:
		return 1
	case ASSIGN, ADD_ASSIGN, SUB_ASSIGN, MUL_ASSIGN, QUO_ASSIGN, REM_ASSIGN,
		AND_ASSIGN, OR_ASSIGN, XOR_ASSIGN, SHL_ASSIGN, SHR_ASSIGN:
		return 2
	case QUESTION, COLON:
		return 3
	case LOR:
		return 4
	case LAND:
		return 5
	case OR:
		return 6
	case XOR:
		return 7
	case AND:
		return 8
	case EQL, NEQ:
		return 9
	case LSS, LEQ, GTR, GEQ:
		return 10
	case SHL, SHR:
		return 11
	case ADD, SUB:
		return 12
	case MUL, QUO, REM:
		return 13
	case INC, DEC, ARROW, DOT:
		return 14 // Postfix
	default:
		return -1
	}
}

// OperatorAssociativity returns the associativity of an operator.
// Returns "left", "right", or "none".
func OperatorAssociativity(t TokenType) string {
	switch t {
	case ASSIGN, ADD_ASSIGN, SUB_ASSIGN, MUL_ASSIGN, QUO_ASSIGN, REM_ASSIGN,
		AND_ASSIGN, OR_ASSIGN, XOR_ASSIGN, SHL_ASSIGN, SHR_ASSIGN:
		return "right"
	case QUESTION, COLON:
		return "right"
	case INC, DEC:
		return "none" // Can be prefix or postfix
	default:
		if IsBinaryOperator(t) {
			return "left"
		}
		return "none"
	}
}

// TokenCategory represents a category of tokens.
type TokenCategory int

const (
	CategorySpecial TokenCategory = iota
	CategoryKeyword
	CategoryLiteral
	CategoryOperator
	CategoryPunctuation
	CategoryIdentifier
)

// Categorize returns the category of a token type.
func Categorize(t TokenType) TokenCategory {
	switch t {
	case ILLEGAL, EOF, COMMENT, NEWLINE:
		return CategorySpecial
	case IDENT:
		return CategoryIdentifier
	case INT_LIT, FLOAT_LIT, CHAR_LIT, STRING_LIT:
		return CategoryLiteral
	}
	if IsKeyword(t) {
		return CategoryKeyword
	}
	if IsOperator(t) {
		return CategoryOperator
	}
	if IsPunctuation(t) {
		return CategoryPunctuation
	}
	return CategorySpecial
}

// TokenList represents a list of tokens with utility methods.
type TokenList []Token

// String returns a formatted string representation of all tokens.
func (tl TokenList) String() string {
	var sb strings.Builder
	for i, tok := range tl {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(tok.String())
	}
	return sb.String()
}

// Filter returns a new token list with tokens matching the predicate.
func (tl TokenList) Filter(pred func(Token) bool) TokenList {
	var result TokenList
	for _, tok := range tl {
		if pred(tok) {
			result = append(result, tok)
		}
	}
	return result
}

// FilterByType returns a new token list with tokens of the specified types.
func (tl TokenList) FilterByType(types ...TokenType) TokenList {
	typeSet := make(map[TokenType]bool)
	for _, t := range types {
		typeSet[t] = true
	}
	return tl.Filter(func(tok Token) bool {
		return typeSet[tok.Type]
	})
}

// FilterByCategory returns a new token list with tokens of the specified category.
func (tl TokenList) FilterByCategory(cat TokenCategory) TokenList {
	return tl.Filter(func(tok Token) bool {
		return Categorize(tok.Type) == cat
	})
}

// Positions returns all positions in the token list.
func (tl TokenList) Positions() []Position {
	positions := make([]Position, len(tl))
	for i, tok := range tl {
		positions[i] = tok.Pos
	}
	return positions
}

// Values returns all values in the token list.
func (tl TokenList) Values() []string {
	values := make([]string, len(tl))
	for i, tok := range tl {
		values[i] = tok.Value
	}
	return values
}

// Types returns all types in the token list.
func (tl TokenList) Types() []TokenType {
	types := make([]TokenType, len(tl))
	for i, tok := range tl {
		types[i] = tok.Type
	}
	return types
}

// FindByPosition returns the token at the given position, or -1 if not found.
func (tl TokenList) FindByPosition(pos Position) int {
	for i, tok := range tl {
		if tok.Pos.Line == pos.Line && tok.Pos.Column == pos.Column {
			return i
		}
	}
	return -1
}

// FindByLine returns all tokens on the given line.
func (tl TokenList) FindByLine(line int) TokenList {
	return tl.Filter(func(tok Token) bool {
		return tok.Pos.Line == line
	})
}

// Helper functions for creating tokens

// NewToken creates a new token with the given parameters.
func NewToken(t TokenType, value string, pos Position) Token {
	return Token{
		Type:  t,
		Value: value,
		Pos:   pos,
		Raw:   value,
	}
}

// NewTokenWithRaw creates a new token with a separate raw value.
func NewTokenWithRaw(t TokenType, value, raw string, pos Position) Token {
	return Token{
		Type:  t,
		Value: value,
		Pos:   pos,
		Raw:   raw,
	}
}

// NewIllegalToken creates an illegal token with an error message.
func NewIllegalToken(value string, pos Position) Token {
	return Token{
		Type:  ILLEGAL,
		Value: value,
		Pos:   pos,
		Raw:   value,
	}
}

// NewEOFToken creates an EOF token at the given position.
func NewEOFToken(pos Position) Token {
	return Token{
		Type:  EOF,
		Value: "",
		Pos:   pos,
		Raw:   "",
	}
}

// NewCommentToken creates a comment token.
func NewCommentToken(value string, pos Position) Token {
	return Token{
		Type:  COMMENT,
		Value: value,
		Pos:   pos,
		Raw:   value,
	}
}

// NewIntLiteralToken creates an integer literal token.
func NewIntLiteralToken(value, raw string, pos Position) Token {
	return Token{
		Type:  INT_LIT,
		Value: value,
		Pos:   pos,
		Raw:   raw,
	}
}

// NewFloatLiteralToken creates a floating-point literal token.
func NewFloatLiteralToken(value, raw string, pos Position) Token {
	return Token{
		Type:  FLOAT_LIT,
		Value: value,
		Pos:   pos,
		Raw:   raw,
	}
}

// NewCharLiteralToken creates a character literal token.
func NewCharLiteralToken(value, raw string, pos Position) Token {
	return Token{
		Type:  CHAR_LIT,
		Value: value,
		Pos:   pos,
		Raw:   raw,
	}
}

// NewStringLiteralToken creates a string literal token.
func NewStringLiteralToken(value, raw string, pos Position) Token {
	return Token{
		Type:  STRING_LIT,
		Value: value,
		Pos:   pos,
		Raw:   raw,
	}
}

// NewIdentifierToken creates an identifier token.
func NewIdentifierToken(value string, pos Position) Token {
	return Token{
		Type:  IDENT,
		Value: value,
		Pos:   pos,
		Raw:   value,
	}
}

// NewKeywordToken creates a keyword token.
func NewKeywordToken(t TokenType, value string, pos Position) Token {
	return Token{
		Type:  t,
		Value: value,
		Pos:   pos,
		Raw:   value,
	}
}

// NewOperatorToken creates an operator token.
func NewOperatorToken(t TokenType, value string, pos Position) Token {
	return Token{
		Type:  t,
		Value: value,
		Pos:   pos,
		Raw:   value,
	}
}

// NewPunctuationToken creates a punctuation token.
func NewPunctuationToken(t TokenType, value string, pos Position) Token {
	return Token{
		Type:  t,
		Value: value,
		Pos:   pos,
		Raw:   value,
	}
}

// Integer suffix information for C11
type IntegerSuffix struct {
	Unsigned bool // 'u' or 'U' suffix
	Long     int  // number of 'l' or 'L' suffixes (0, 1, or 2)
}

// FloatSuffix represents the suffix of a floating-point literal.
type FloatSuffix struct {
	Type FloatType // f, F, l, or L
}

// FloatType represents the type suffix of a floating-point literal.
type FloatType int

const (
	FloatNone FloatType = iota
	FloatFloat  // f or F suffix
	FloatLong    // l or L suffix
	FloatDecimal // decimal floating-point (df, DF, etc.) - C23
)

// EscapeSequence represents an escape sequence in a character or string literal.
type EscapeSequence struct {
	Type    EscapeType
	Value   rune
	Octal   string // For octal escapes
	Hex     string // For hex escapes
	Unicode string // For universal character names
}

// EscapeType represents the type of escape sequence.
type EscapeType int

const (
	EscapeSimple EscapeType = iota // Simple escape like \n, \t
	EscapeOctal                    // Octal escape like \123
	EscapeHex                      // Hex escape like \x1a
	EscapeUnicode                  // Universal character name like \u1234 or \U12345678
)