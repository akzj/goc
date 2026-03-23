// Package lexer provides lexical analysis for C11 source code.
// This file contains tests for token definitions.
package lexer

import (
	"testing"
)

func TestPosition(t *testing.T) {
	tests := []struct {
		name     string
		pos      Position
		expected string
		isValid  bool
	}{
		{
			name:     "valid position with file",
			pos:      Position{File: "test.c", Line: 10, Column: 5},
			expected: "test.c:10:5",
			isValid:  true,
		},
		{
			name:     "valid position without file",
			pos:      Position{File: "", Line: 1, Column: 1},
			expected: "1:1",
			isValid:  true,
		},
		{
			name:     "invalid position",
			pos:      Position{File: "test.c", Line: 0, Column: 0},
			expected: "test.c:0:0",
			isValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pos.String(); got != tt.expected {
				t.Errorf("Position.String() = %q, want %q", got, tt.expected)
			}
			if got := tt.pos.IsValid(); got != tt.isValid {
				t.Errorf("Position.IsValid() = %v, want %v", got, tt.isValid)
			}
		})
	}
}

func TestTokenString(t *testing.T) {
	tests := []struct {
		name     string
		token    Token
		expected string
	}{
		{
			name:     "simple keyword",
			token:    Token{Type: INT, Value: "int", Pos: Position{Line: 1, Column: 1}},
			expected: "int @ 1:1",
		},
		{
			name:     "identifier with value",
			token:    Token{Type: IDENT, Value: "myVar", Pos: Position{File: "test.c", Line: 5, Column: 10}},
			expected: `IDENT("myVar") @ test.c:5:10`,
		},
		{
			name:     "integer literal",
			token:    Token{Type: INT_LIT, Value: "42", Pos: Position{Line: 1, Column: 1}},
			expected: `INT_LIT("42") @ 1:1`,
		},
		{
			name:     "operator",
			token:    Token{Type: ADD, Value: "+", Pos: Position{Line: 1, Column: 5}},
			expected: "+ @ 1:5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.String(); got != tt.expected {
				t.Errorf("Token.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTokenClassification(t *testing.T) {
	tests := []struct {
		name      string
		token     Token
		isKeyword bool
		isIdent   bool
		isLiteral bool
		isOp      bool
		isPunct   bool
	}{
		{
			name:      "keyword int",
			token:     Token{Type: INT, Value: "int"},
			isKeyword: true,
		},
		{
			name:    "identifier",
			token:   Token{Type: IDENT, Value: "myVar"},
			isIdent: true,
		},
		{
			name:      "integer literal",
			token:     Token{Type: INT_LIT, Value: "42"},
			isLiteral: true,
		},
		{
			name:      "float literal",
			token:     Token{Type: FLOAT_LIT, Value: "3.14"},
			isLiteral: true,
		},
		{
			name:      "char literal",
			token:     Token{Type: CHAR_LIT, Value: "a"},
			isLiteral: true,
		},
		{
			name:      "string literal",
			token:     Token{Type: STRING_LIT, Value: "hello"},
			isLiteral: true,
		},
		{
			name:  "addition operator",
			token: Token{Type: ADD, Value: "+"},
			isOp:  true,
		},
		{
			name:  "logical and operator",
			token: Token{Type: LAND, Value: "&&"},
			isOp:  true,
		},
		{
			name:    "left paren",
			token:   Token{Type: LPAREN, Value: "("},
			isPunct: true,
		},
		{
			name:    "semicolon",
			token:   Token{Type: SEMICOLON, Value: ";"},
			isPunct: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsKeyword(); got != tt.isKeyword {
				t.Errorf("Token.IsKeyword() = %v, want %v", got, tt.isKeyword)
			}
			if got := tt.token.IsIdentifier(); got != tt.isIdent {
				t.Errorf("Token.IsIdentifier() = %v, want %v", got, tt.isIdent)
			}
			if got := tt.token.IsLiteral(); got != tt.isLiteral {
				t.Errorf("Token.IsLiteral() = %v, want %v", got, tt.isLiteral)
			}
			if got := tt.token.IsOperator(); got != tt.isOp {
				t.Errorf("Token.IsOperator() = %v, want %v", got, tt.isOp)
			}
			if got := tt.token.IsPunctuation(); got != tt.isPunct {
				t.Errorf("Token.IsPunctuation() = %v, want %v", got, tt.isPunct)
			}
		})
	}
}

func TestIsKeyword(t *testing.T) {
	// Test all C11 keywords
	keywords := []TokenType{
		AUTO, REGISTER, STATIC, EXTERN, TYPEDEF,
		CONST, VOLATILE, RESTRICT,
		VOID, CHAR, SHORT, INT, LONG, FLOAT, DOUBLE,
		SIGNED, UNSIGNED, COMPLEX, IMAGINARY, BOOL, ATOMIC,
		STRUCT, UNION, ENUM,
		IF, ELSE, SWITCH, CASE, DEFAULT,
		WHILE, DO, FOR, GOTO, CONTINUE, BREAK, RETURN,
		INLINE, NORETURN, THREAD_LOCAL,
		ALIGNAS, ALIGNOF, GENERIC, STATIC_ASSERT,
		SIZEOF,
	}

	for _, kw := range keywords {
		if !IsKeyword(kw) {
			t.Errorf("IsKeyword(%q) = false, want true", kw)
		}
	}

	// Test non-keywords
	nonKeywords := []TokenType{IDENT, INT_LIT, ADD, LPAREN, ILLEGAL}
	for _, nonKw := range nonKeywords {
		if IsKeyword(nonKw) {
			t.Errorf("IsKeyword(%q) = true, want false", nonKw)
		}
	}
}

func TestIsKeywordString(t *testing.T) {
	tests := []struct {
		s       string
		isKw    bool
	}{
		{"int", true},
		{"float", true},
		{"_Bool", true},
		{"_Generic", true},
		{"_Static_assert", true},
		{"myVar", false},
		{"INT", false}, // Keywords are case-sensitive
		{"", false},
	}

	for _, tt := range tests {
		if got := IsKeywordString(tt.s); got != tt.isKw {
			t.Errorf("IsKeywordString(%q) = %v, want %v", tt.s, got, tt.isKw)
		}
	}
}

func TestLookupKeyword(t *testing.T) {
	tests := []struct {
		s        string
		expected TokenType
	}{
		{"int", INT},
		{"float", FLOAT},
		{"_Bool", BOOL},
		{"_Generic", GENERIC},
		{"myVar", IDENT},
		{"INT", IDENT}, // Case-sensitive
		{"", IDENT},
	}

	for _, tt := range tests {
		if got := LookupKeyword(tt.s); got != tt.expected {
			t.Errorf("LookupKeyword(%q) = %q, want %q", tt.s, got, tt.expected)
		}
	}
}

func TestIsLiteral(t *testing.T) {
	literals := []TokenType{INT_LIT, FLOAT_LIT, CHAR_LIT, STRING_LIT}
	for _, lit := range literals {
		if !IsLiteral(lit) {
			t.Errorf("IsLiteral(%q) = false, want true", lit)
		}
	}

	nonLiterals := []TokenType{IDENT, INT, ADD, LPAREN, ILLEGAL}
	for _, nonLit := range nonLiterals {
		if IsLiteral(nonLit) {
			t.Errorf("IsLiteral(%q) = true, want false", nonLit)
		}
	}
}

func TestIsOperator(t *testing.T) {
	operators := []TokenType{
		ADD, SUB, MUL, QUO, REM,
		AND, OR, XOR, SHL, SHR, BITNOT,
		LAND, LOR, NOT,
		EQL, NEQ, LSS, LEQ, GTR, GEQ,
		ASSIGN, ADD_ASSIGN, SUB_ASSIGN, MUL_ASSIGN, QUO_ASSIGN, REM_ASSIGN,
		AND_ASSIGN, OR_ASSIGN, XOR_ASSIGN, SHL_ASSIGN, SHR_ASSIGN,
		INC, DEC, ARROW, DOT,
		QUESTION, COLON,
	}

	for _, op := range operators {
		if !IsOperator(op) {
			t.Errorf("IsOperator(%q) = false, want true", op)
		}
	}

	nonOperators := []TokenType{IDENT, INT_LIT, INT, LPAREN, SEMICOLON}
	for _, nonOp := range nonOperators {
		if IsOperator(nonOp) {
			t.Errorf("IsOperator(%q) = true, want false", nonOp)
		}
	}
}

func TestIsPunctuation(t *testing.T) {
	punctuation := []TokenType{
		LPAREN, RPAREN, LBRACK, RBRACK, LBRACE, RBRACE,
		COMMA, SEMICOLON, ELLIPSIS,
	}

	for _, punct := range punctuation {
		if !IsPunctuation(punct) {
			t.Errorf("IsPunctuation(%q) = false, want true", punct)
		}
	}

	nonPunctuation := []TokenType{IDENT, INT_LIT, INT, ADD, ASSIGN}
	for _, nonPunct := range nonPunctuation {
		if IsPunctuation(nonPunct) {
			t.Errorf("IsPunctuation(%q) = true, want false", nonPunct)
		}
	}
}

func TestIsAssignmentOperator(t *testing.T) {
	assignOps := []TokenType{
		ASSIGN, ADD_ASSIGN, SUB_ASSIGN, MUL_ASSIGN, QUO_ASSIGN, REM_ASSIGN,
		AND_ASSIGN, OR_ASSIGN, XOR_ASSIGN, SHL_ASSIGN, SHR_ASSIGN,
	}

	for _, op := range assignOps {
		if !IsAssignmentOperator(op) {
			t.Errorf("IsAssignmentOperator(%q) = false, want true", op)
		}
	}

	nonAssignOps := []TokenType{ADD, SUB, EQL, INC, DEC}
	for _, op := range nonAssignOps {
		if IsAssignmentOperator(op) {
			t.Errorf("IsAssignmentOperator(%q) = true, want false", op)
		}
	}
}

func TestIsBinaryOperator(t *testing.T) {
	binaryOps := []TokenType{
		ADD, SUB, MUL, QUO, REM,
		AND, OR, XOR, SHL, SHR,
		LAND, LOR,
		EQL, NEQ, LSS, LEQ, GTR, GEQ,
		COMMA,
	}

	for _, op := range binaryOps {
		if !IsBinaryOperator(op) {
			t.Errorf("IsBinaryOperator(%q) = false, want true", op)
		}
	}

	nonBinaryOps := []TokenType{ASSIGN, INC, DEC, NOT, BITNOT}
	for _, op := range nonBinaryOps {
		if IsBinaryOperator(op) {
			t.Errorf("IsBinaryOperator(%q) = true, want false", op)
		}
	}
}

func TestIsUnaryOperator(t *testing.T) {
	unaryOps := []TokenType{ADD, SUB, MUL, AND, NOT, BITNOT, INC, DEC, SIZEOF}

	for _, op := range unaryOps {
		if !IsUnaryOperator(op) {
			t.Errorf("IsUnaryOperator(%q) = false, want true", op)
		}
	}

	nonUnaryOps := []TokenType{EQL, LAND, LOR, QUO, REM}
	for _, op := range nonUnaryOps {
		if IsUnaryOperator(op) {
			t.Errorf("IsUnaryOperator(%q) = true, want false", op)
		}
	}
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		op          TokenType
		precedence  int
	}{
		{COMMA, 1},
		{ASSIGN, 2},
		{QUESTION, 3},
		{LOR, 4},
		{LAND, 5},
		{OR, 6},
		{XOR, 7},
		{AND, 8},
		{EQL, 9},
		{LSS, 10},
		{SHL, 11},
		{ADD, 12},
		{MUL, 13},
		{INC, 14},
		{IDENT, -1},
		{INT_LIT, -1},
	}

	for _, tt := range tests {
		if got := OperatorPrecedence(tt.op); got != tt.precedence {
			t.Errorf("OperatorPrecedence(%q) = %d, want %d", tt.op, got, tt.precedence)
		}
	}
}

func TestOperatorAssociativity(t *testing.T) {
	tests := []struct {
		op            TokenType
		associativity string
	}{
		{ASSIGN, "right"},
		{ADD_ASSIGN, "right"},
		{QUESTION, "right"},
		{ADD, "left"},
		{MUL, "left"},
		{LAND, "left"},
		{INC, "none"},
		{DEC, "none"},
		{IDENT, "none"},
	}

	for _, tt := range tests {
		if got := OperatorAssociativity(tt.op); got != tt.associativity {
			t.Errorf("OperatorAssociativity(%q) = %q, want %q", tt.op, got, tt.associativity)
		}
	}
}

func TestCategorize(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		category  TokenCategory
	}{
		{ILLEGAL, CategorySpecial},
		{EOF, CategorySpecial},
		{COMMENT, CategorySpecial},
		{NEWLINE, CategorySpecial},
		{IDENT, CategoryIdentifier},
		{INT_LIT, CategoryLiteral},
		{FLOAT_LIT, CategoryLiteral},
		{CHAR_LIT, CategoryLiteral},
		{STRING_LIT, CategoryLiteral},
		{INT, CategoryKeyword},
		{FLOAT, CategoryKeyword},
		{IF, CategoryKeyword},
		{ADD, CategoryOperator},
		{LAND, CategoryOperator},
		{ASSIGN, CategoryOperator},
		{LPAREN, CategoryPunctuation},
		{SEMICOLON, CategoryPunctuation},
		{LBRACE, CategoryPunctuation},
	}

	for _, tt := range tests {
		if got := Categorize(tt.tokenType); got != tt.category {
			t.Errorf("Categorize(%q) = %v, want %v", tt.tokenType, got, tt.category)
		}
	}
}

func TestTokenList(t *testing.T) {
	tokens := TokenList{
		{Type: INT, Value: "int", Pos: Position{Line: 1, Column: 1}},
		{Type: IDENT, Value: "main", Pos: Position{Line: 1, Column: 5}},
		{Type: LPAREN, Value: "(", Pos: Position{Line: 1, Column: 9}},
		{Type: RPAREN, Value: ")", Pos: Position{Line: 1, Column: 10}},
		{Type: LBRACE, Value: "{", Pos: Position{Line: 1, Column: 12}},
		{Type: RETURN, Value: "return", Pos: Position{Line: 2, Column: 5}},
		{Type: INT_LIT, Value: "0", Pos: Position{Line: 2, Column: 12}},
		{Type: SEMICOLON, Value: ";", Pos: Position{Line: 2, Column: 13}},
		{Type: RBRACE, Value: "}", Pos: Position{Line: 3, Column: 1}},
	}

	t.Run("Filter", func(t *testing.T) {
		keywords := tokens.Filter(func(tok Token) bool {
			return tok.IsKeyword()
		})
		if len(keywords) != 2 { // int, return
			t.Errorf("Filter keywords: got %d, want 2", len(keywords))
		}
	})

	t.Run("FilterByType", func(t *testing.T) {
		parens := tokens.FilterByType(LPAREN, RPAREN)
		if len(parens) != 2 {
			t.Errorf("FilterByType LPAREN, RPAREN: got %d, want 2", len(parens))
		}
	})

	t.Run("FilterByCategory", func(t *testing.T) {
		ops := tokens.FilterByCategory(CategoryOperator)
		if len(ops) != 0 {
			t.Errorf("FilterByCategory Operator: got %d, want 0", len(ops))
		}
		punct := tokens.FilterByCategory(CategoryPunctuation)
		if len(punct) != 5 { // (, ), {, ;, }
			t.Errorf("FilterByCategory Punctuation: got %d, want 5", len(punct))
		}
	})

	t.Run("Positions", func(t *testing.T) {
		positions := tokens.Positions()
		if len(positions) != len(tokens) {
			t.Errorf("Positions: got %d, want %d", len(positions), len(tokens))
		}
	})

	t.Run("Values", func(t *testing.T) {
		values := tokens.Values()
		expected := []string{"int", "main", "(", ")", "{", "return", "0", ";", "}"}
		for i, v := range values {
			if v != expected[i] {
				t.Errorf("Values[%d]: got %q, want %q", i, v, expected[i])
			}
		}
	})

	t.Run("Types", func(t *testing.T) {
		types := tokens.Types()
		if len(types) != len(tokens) {
			t.Errorf("Types: got %d, want %d", len(types), len(tokens))
		}
	})

	t.Run("FindByPosition", func(t *testing.T) {
		idx := tokens.FindByPosition(Position{Line: 1, Column: 5})
		if idx != 1 {
			t.Errorf("FindByPosition (1,5): got %d, want 1", idx)
		}
		idx = tokens.FindByPosition(Position{Line: 100, Column: 100})
		if idx != -1 {
			t.Errorf("FindByPosition (100,100): got %d, want -1", idx)
		}
	})

	t.Run("FindByLine", func(t *testing.T) {
		line1 := tokens.FindByLine(1)
		if len(line1) != 5 { // int, main, (, ), {
			t.Errorf("FindByLine(1): got %d tokens, want 5", len(line1))
		}
		line2 := tokens.FindByLine(2)
		if len(line2) != 3 { // return, 0, ;
			t.Errorf("FindByLine(2): got %d tokens, want 3", len(line2))
		}
	})
}

func TestNewToken(t *testing.T) {
	pos := Position{File: "test.c", Line: 10, Column: 5}
	tok := NewToken(INT, "int", pos)

	if tok.Type != INT {
		t.Errorf("NewToken Type: got %q, want %q", tok.Type, INT)
	}
	if tok.Value != "int" {
		t.Errorf("NewToken Value: got %q, want %q", tok.Value, "int")
	}
	if tok.Pos != pos {
		t.Errorf("NewToken Pos: got %v, want %v", tok.Pos, pos)
	}
	if tok.Raw != "int" {
		t.Errorf("NewToken Raw: got %q, want %q", tok.Raw, "int")
	}
}

func TestNewTokenWithRaw(t *testing.T) {
	pos := Position{Line: 1, Column: 1}
	tok := NewTokenWithRaw(INT_LIT, "42", "0x2A", pos)

	if tok.Type != INT_LIT {
		t.Errorf("NewTokenWithRaw Type: got %q, want %q", tok.Type, INT_LIT)
	}
	if tok.Value != "42" {
		t.Errorf("NewTokenWithRaw Value: got %q, want %q", tok.Value, "42")
	}
	if tok.Raw != "0x2A" {
		t.Errorf("NewTokenWithRaw Raw: got %q, want %q", tok.Raw, "0x2A")
	}
}

func TestNewIllegalToken(t *testing.T) {
	pos := Position{Line: 1, Column: 1}
	tok := NewIllegalToken("@", pos)

	if tok.Type != ILLEGAL {
		t.Errorf("NewIllegalToken Type: got %q, want %q", tok.Type, ILLEGAL)
	}
	if tok.Value != "@" {
		t.Errorf("NewIllegalToken Value: got %q, want %q", tok.Value, "@")
	}
}

func TestNewEOFToken(t *testing.T) {
	pos := Position{Line: 10, Column: 1}
	tok := NewEOFToken(pos)

	if tok.Type != EOF {
		t.Errorf("NewEOFToken Type: got %q, want %q", tok.Type, EOF)
	}
	if tok.Value != "" {
		t.Errorf("NewEOFToken Value: got %q, want empty", tok.Value)
	}
}

func TestNewCommentToken(t *testing.T) {
	pos := Position{Line: 1, Column: 1}
	tok := NewCommentToken("// This is a comment", pos)

	if tok.Type != COMMENT {
		t.Errorf("NewCommentToken Type: got %q, want %q", tok.Type, COMMENT)
	}
}

func TestNewIntLiteralToken(t *testing.T) {
	pos := Position{Line: 1, Column: 1}
	tok := NewIntLiteralToken("255", "0xFF", pos)

	if tok.Type != INT_LIT {
		t.Errorf("NewIntLiteralToken Type: got %q, want %q", tok.Type, INT_LIT)
	}
	if tok.Value != "255" {
		t.Errorf("NewIntLiteralToken Value: got %q, want %q", tok.Value, "255")
	}
	if tok.Raw != "0xFF" {
		t.Errorf("NewIntLiteralToken Raw: got %q, want %q", tok.Raw, "0xFF")
	}
}

func TestNewFloatLiteralToken(t *testing.T) {
	pos := Position{Line: 1, Column: 1}
	tok := NewFloatLiteralToken("3.14", "3.14f", pos)

	if tok.Type != FLOAT_LIT {
		t.Errorf("NewFloatLiteralToken Type: got %q, want %q", tok.Type, FLOAT_LIT)
	}
}

func TestNewCharLiteralToken(t *testing.T) {
	pos := Position{Line: 1, Column: 1}
	tok := NewCharLiteralToken("a", "'a'", pos)

	if tok.Type != CHAR_LIT {
		t.Errorf("NewCharLiteralToken Type: got %q, want %q", tok.Type, CHAR_LIT)
	}
}

func TestNewStringLiteralToken(t *testing.T) {
	pos := Position{Line: 1, Column: 1}
	tok := NewStringLiteralToken("hello", "\"hello\"", pos)

	if tok.Type != STRING_LIT {
		t.Errorf("NewStringLiteralToken Type: got %q, want %q", tok.Type, STRING_LIT)
	}
}

func TestNewIdentifierToken(t *testing.T) {
	pos := Position{Line: 1, Column: 1}
	tok := NewIdentifierToken("myVariable", pos)

	if tok.Type != IDENT {
		t.Errorf("NewIdentifierToken Type: got %q, want %q", tok.Type, IDENT)
	}
	if tok.Value != "myVariable" {
		t.Errorf("NewIdentifierToken Value: got %q, want %q", tok.Value, "myVariable")
	}
}

func TestNewKeywordToken(t *testing.T) {
	pos := Position{Line: 1, Column: 1}
	tok := NewKeywordToken(IF, "if", pos)

	if tok.Type != IF {
		t.Errorf("NewKeywordToken Type: got %q, want %q", tok.Type, IF)
	}
}

func TestNewOperatorToken(t *testing.T) {
	pos := Position{Line: 1, Column: 1}
	tok := NewOperatorToken(LAND, "&&", pos)

	if tok.Type != LAND {
		t.Errorf("NewOperatorToken Type: got %q, want %q", tok.Type, LAND)
	}
}

func TestNewPunctuationToken(t *testing.T) {
	pos := Position{Line: 1, Column: 1}
	tok := NewPunctuationToken(LPAREN, "(", pos)

	if tok.Type != LPAREN {
		t.Errorf("NewPunctuationToken Type: got %q, want %q", tok.Type, LPAREN)
	}
}

func TestAllKeywordsInMap(t *testing.T) {
	// Ensure all keyword token types are in the Keywords map
	keywordTypes := []TokenType{
		AUTO, REGISTER, STATIC, EXTERN, TYPEDEF,
		CONST, VOLATILE, RESTRICT,
		VOID, CHAR, SHORT, INT, LONG, FLOAT, DOUBLE,
		SIGNED, UNSIGNED, COMPLEX, IMAGINARY, BOOL, ATOMIC,
		STRUCT, UNION, ENUM,
		IF, ELSE, SWITCH, CASE, DEFAULT,
		WHILE, DO, FOR, GOTO, CONTINUE, BREAK, RETURN,
		INLINE, NORETURN, THREAD_LOCAL,
		ALIGNAS, ALIGNOF, GENERIC, STATIC_ASSERT,
		SIZEOF, PRAGMA_OP,
	}

	for _, kt := range keywordTypes {
		if _, ok := Keywords[string(kt)]; !ok {
			t.Errorf("Keyword %q not in Keywords map", kt)
		}
	}
}

func TestKeywordsCount(t *testing.T) {
	// C11 has 37 keywords plus some additional ones
	// We should have all of them
	if len(Keywords) < 37 {
		t.Errorf("Keywords map has %d entries, expected at least 37", len(Keywords))
	}
}

// Benchmark tests
func BenchmarkLookupKeyword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LookupKeyword("int")
		LookupKeyword("myVar")
		LookupKeyword("_Generic")
	}
}

func BenchmarkIsKeyword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsKeyword(INT)
		IsKeyword(IDENT)
	}
}

func BenchmarkTokenString(b *testing.B) {
	tok := Token{
		Type:  IDENT,
		Value: "myVariable",
		Pos:   Position{File: "test.c", Line: 100, Column: 50},
	}
	for i := 0; i < b.N; i++ {
		_ = tok.String()
	}
}

func BenchmarkOperatorPrecedence(b *testing.B) {
	for i := 0; i < b.N; i++ {
		OperatorPrecedence(ADD)
		OperatorPrecedence(LAND)
		OperatorPrecedence(ASSIGN)
	}
}