// Package lexer provides lexical analysis for C11 source code.
// This file contains tests for the Lexer implementation.
package lexer

import (
	"testing"
)

func TestLexerBasicTokens(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []Token
	}{
		{
			name:   "empty source",
			source: "",
			expected: []Token{
				{Type: EOF, Value: "", Pos: Position{Line: 1, Column: 1}},
			},
		},
		{
			name:   "whitespace only",
			source: "   \t\n  ",
			expected: []Token{
				{Type: EOF, Value: "", Pos: Position{Line: 2, Column: 3}},
			},
		},
		{
			name:   "single keyword",
			source: "int",
			expected: []Token{
				{Type: INT, Value: "int", Pos: Position{Line: 1, Column: 1}},
				{Type: EOF, Value: "", Pos: Position{Line: 1, Column: 4}},
			},
		},
		{
			name:   "identifier",
			source: "myVariable",
			expected: []Token{
				{Type: IDENT, Value: "myVariable", Pos: Position{Line: 1, Column: 1}},
				{Type: EOF, Value: "", Pos: Position{Line: 1, Column: 11}},
			},
		},
		{
			name:   "integer literal",
			source: "42",
			expected: []Token{
				{Type: INT_LIT, Value: "42", Pos: Position{Line: 1, Column: 1}},
				{Type: EOF, Value: "", Pos: Position{Line: 1, Column: 3}},
			},
		},
		{
			name:   "float literal",
			source: "3.14",
			expected: []Token{
				{Type: FLOAT_LIT, Value: "3.14", Pos: Position{Line: 1, Column: 1}},
				{Type: EOF, Value: "", Pos: Position{Line: 1, Column: 5}},
			},
		},
		{
			name:   "character literal",
			source: "'a'",
			expected: []Token{
				{Type: CHAR_LIT, Value: "'a'", Pos: Position{Line: 1, Column: 1}},
				{Type: EOF, Value: "", Pos: Position{Line: 1, Column: 4}},
			},
		},
		{
			name:   "string literal",
			source: `"hello"`,
			expected: []Token{
				{Type: STRING_LIT, Value: `"hello"`, Pos: Position{Line: 1, Column: 1}},
				{Type: EOF, Value: "", Pos: Position{Line: 1, Column: 8}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.source, "")
			tokens := lexer.Tokenize()
			
			if len(tokens) != len(tt.expected) {
				t.Errorf("expected %d tokens, got %d", len(tt.expected), len(tokens))
				return
			}
			
			for i, expected := range tt.expected {
				actual := tokens[i]
				if actual.Type != expected.Type {
					t.Errorf("token %d: expected type %s, got %s", i, expected.Type, actual.Type)
				}
				if actual.Value != expected.Value {
					t.Errorf("token %d: expected value %q, got %q", i, expected.Value, actual.Value)
				}
				if actual.Pos.Line != expected.Pos.Line || actual.Pos.Column != expected.Pos.Column {
					t.Errorf("token %d: expected position %d:%d, got %d:%d", 
						i, expected.Pos.Line, expected.Pos.Column, actual.Pos.Line, actual.Pos.Column)
				}
			}
		})
	}
}

func TestLexerKeywords(t *testing.T) {
	keywords := []string{
		"auto", "break", "case", "char", "const", "continue", "default", "do",
		"double", "else", "enum", "extern", "float", "for", "goto", "if",
		"inline", "int", "long", "register", "restrict", "return", "short",
		"signed", "sizeof", "static", "struct", "switch", "typedef", "union",
		"unsigned", "void", "volatile", "while", "_Alignas", "_Alignof",
		"_Atomic", "_Bool", "_Complex", "_Generic", "_Imaginary", "_Noreturn",
		"_Static_assert", "_Thread_local",
	}
	
	for _, kw := range keywords {
		t.Run(kw, func(t *testing.T) {
			lexer := NewLexer(kw, "")
			tokens := lexer.Tokenize()
			
			if len(tokens) < 1 {
				t.Errorf("no tokens produced for keyword %s", kw)
				return
			}
			
			if tokens[0].Type == IDENT {
				t.Errorf("keyword %s tokenized as IDENT", kw)
			}
			
			if tokens[0].Value != kw {
				t.Errorf("keyword %s has value %q", kw, tokens[0].Value)
			}
		})
	}
}

func TestLexerOperators(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		tokenType  TokenType
	}{
		{"add", "+", ADD},
		{"sub", "-", SUB},
		{"mul", "*", MUL},
		{"div", "/", QUO},
		{"mod", "%", REM},
		{"inc", "++", INC},
		{"dec", "--", DEC},
		{"eq", "==", EQL},
		{"ne", "!=", NEQ},
		{"lt", "<", LSS},
		{"gt", ">", GTR},
		{"le", "<=", LEQ},
		{"ge", ">=", GEQ},
		{"land", "&&", LAND},
		{"lor", "||", LOR},
		{"not", "!", NOT},
		{"bnot", "~", BITNOT},
		{"and", "&", AND},
		{"or", "|", OR},
		{"xor", "^", XOR},
		{"shl", "<<", SHL},
		{"shr", ">>", SHR},
		{"assign", "=", ASSIGN},
		{"add_assign", "+=", ADD_ASSIGN},
		{"sub_assign", "-=", SUB_ASSIGN},
		{"mul_assign", "*=", MUL_ASSIGN},
		{"div_assign", "/=", QUO_ASSIGN},
		{"mod_assign", "%=", REM_ASSIGN},
		{"and_assign", "&=", AND_ASSIGN},
		{"or_assign", "|=", OR_ASSIGN},
		{"xor_assign", "^=", XOR_ASSIGN},
		{"shl_assign", "<<=", SHL_ASSIGN},
		{"shr_assign", ">>=", SHR_ASSIGN},
		{"arrow", "->", ARROW},
		{"dot", ".", DOT},
		{"ellipsis", "...", ELLIPSIS},
		{"question", "?", QUESTION},
		{"colon", ":", COLON},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.source, "")
			tokens := lexer.Tokenize()
			
			if len(tokens) < 1 {
				t.Errorf("no tokens produced for operator %s", tt.source)
				return
			}
			
			if tokens[0].Type != tt.tokenType {
				t.Errorf("operator %s tokenized as %s, expected %s", 
					tt.source, tokens[0].Type, tt.tokenType)
			}
		})
	}
}

func TestLexerPunctuation(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		tokenType TokenType
	}{
		{"lparen", "(", LPAREN},
		{"rparen", ")", RPAREN},
		{"lbrack", "[", LBRACK},
		{"rbrack", "]", RBRACK},
		{"lbrace", "{", LBRACE},
		{"rbrace", "}", RBRACE},
		{"comma", ",", COMMA},
		{"semicolon", ";", SEMICOLON},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.source, "")
			tokens := lexer.Tokenize()
			
			if len(tokens) < 1 {
				t.Errorf("no tokens produced for punctuation %s", tt.source)
				return
			}
			
			if tokens[0].Type != tt.tokenType {
				t.Errorf("punctuation %s tokenized as %s, expected %s", 
					tt.source, tokens[0].Type, tt.tokenType)
			}
		})
	}
}

func TestLexerIntegerLiterals(t *testing.T) {
	tests := []struct {
		name   string
		source string
		value  string
	}{
		{"decimal", "42", "42"},
		{"zero", "0", "0"},
		{"large decimal", "1234567890", "1234567890"},
		{"hex lowercase", "0x1a2b", "0x1a2b"},
		{"hex uppercase", "0X1A2B", "0X1A2B"},
		{"octal", "0755", "0755"},
		{"binary", "0b1010", "0b1010"},
		{"unsigned suffix", "42u", "42u"},
		{"long suffix", "42l", "42l"},
		{"ulong suffix", "42ul", "42ul"},
		{"ULL suffix", "42ULL", "42ULL"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.source, "")
			tokens := lexer.Tokenize()
			
			if len(tokens) < 1 {
				t.Errorf("no tokens produced for integer literal %s", tt.source)
				return
			}
			
			if tokens[0].Type != INT_LIT {
				t.Errorf("integer literal %s tokenized as %s", tt.source, tokens[0].Type)
			}
			
			if tokens[0].Value != tt.value {
				t.Errorf("integer literal %s has value %q, expected %q", 
					tt.source, tokens[0].Value, tt.value)
			}
		})
	}
}

func TestLexerFloatLiterals(t *testing.T) {
	tests := []struct {
		name   string
		source string
		value  string
	}{
		{"simple float", "3.14", "3.14"},
		{"float with exponent", "1.5e10", "1.5e10"},
		{"float with negative exponent", "1.5e-10", "1.5e-10"},
		{"float with positive exponent", "1.5e+10", "1.5e+10"},
		{"uppercase exponent", "1.5E10", "1.5E10"},
		{"float suffix", "3.14f", "3.14f"},
		{"double suffix", "3.14l", "3.14l"},
		{"hex float", "0x1.8p1", "0x1.8p1"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.source, "")
			tokens := lexer.Tokenize()
			
			if len(tokens) < 1 {
				t.Errorf("no tokens produced for float literal %s", tt.source)
				return
			}
			
			if tokens[0].Type != FLOAT_LIT {
				t.Errorf("float literal %s tokenized as %s", tt.source, tokens[0].Type)
			}
			
			if tokens[0].Value != tt.value {
				t.Errorf("float literal %s has value %q, expected %q", 
					tt.source, tokens[0].Value, tt.value)
			}
		})
	}
}

func TestLexerCharacterLiterals(t *testing.T) {
	tests := []struct {
		name   string
		source string
		value  string
	}{
		{"simple char", "'a'", "'a'"},
		{"digit char", "'5'", "'5'"},
		{"newline escape", "'\\n'", "'\\n'"},
		{"tab escape", "'\\t'", "'\\t'"},
		{"backslash escape", "'\\\\'", "'\\\\'"},
		{"single quote escape", "'\\''", "'\\''"},
		{"hex escape", "'\\x41'", "'\\x41'"},
		{"octal escape", "'\\101'", "'\\101'"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.source, "")
			tokens := lexer.Tokenize()
			
			if len(tokens) < 1 {
				t.Errorf("no tokens produced for char literal %s", tt.source)
				return
			}
			
			if tokens[0].Type != CHAR_LIT {
				t.Errorf("char literal %s tokenized as %s", tt.source, tokens[0].Type)
			}
			
			if tokens[0].Value != tt.value {
				t.Errorf("char literal %s has value %q, expected %q", 
					tt.source, tokens[0].Value, tt.value)
			}
		})
	}
}

func TestLexerStringLiterals(t *testing.T) {
	tests := []struct {
		name   string
		source string
		value  string
	}{
		{"simple string", `"hello"`, `"hello"`},
		{"empty string", `""`, `""`},
		{"string with spaces", `"hello world"`, `"hello world"`},
		{"string with escape", `"hello\\nworld"`, `"hello\\nworld"`},
		{"string with quote", "\"say \\\"hi\\\"\"", "\"say \\\"hi\\\"\""},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.source, "")
			tokens := lexer.Tokenize()
			
			if len(tokens) < 1 {
				t.Errorf("no tokens produced for string literal %s", tt.source)
				return
			}
			
			if tokens[0].Type != STRING_LIT {
				t.Errorf("string literal %s tokenized as %s", tt.source, tokens[0].Type)
			}
			
			if tokens[0].Value != tt.value {
				t.Errorf("string literal %s has value %q, expected %q", 
					tt.source, tokens[0].Value, tt.value)
			}
		})
	}
}

func TestLexerComments(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		expectedType TokenType
		expectedValue string
	}{
		{
			name:         "single line comment",
			source:       "int x; // comment\nint y;",
			expectedType: INT,
			expectedValue: "int",
		},
		{
			name:         "multi line comment",
			source:       "int x; /* comment */ int y;",
			expectedType: INT,
			expectedValue: "int",
		},
		{
			name:         "comment at start",
			source:       "// comment\nint x;",
			expectedType: INT,
			expectedValue: "int",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.source, "")
			tokens := lexer.Tokenize()
			
			// Find the first non-EOF token
			for _, token := range tokens {
				if token.Type == tt.expectedType {
					if token.Value != tt.expectedValue {
						t.Errorf("expected value %q, got %q", tt.expectedValue, token.Value)
					}
					return
				}
			}
			
			t.Errorf("expected token type %s not found", tt.expectedType)
		})
	}
}

func TestLexerComplexCode(t *testing.T) {
	source := `
int main() {
    int x = 42;
    float y = 3.14;
    char *str = "hello";
    if (x > 0) {
        return x;
    }
    return 0;
}
`
	
	lexer := NewLexer(source, "test.c")
	tokens := lexer.Tokenize()
	
	// Verify we get a reasonable number of tokens
	if len(tokens) < 20 {
		t.Errorf("expected at least 20 tokens, got %d", len(tokens))
	}
	
	// Verify first token is 'int'
	if tokens[0].Type != INT {
		t.Errorf("expected first token to be INT, got %s", tokens[0].Type)
	}
	
	// Verify last token is EOF
	if tokens[len(tokens)-1].Type != EOF {
		t.Errorf("expected last token to be EOF, got %s", tokens[len(tokens)-1].Type)
	}
	
	// Check for specific tokens
	foundMain := false
	foundX := false
	found42 := false
	
	for _, token := range tokens {
		if token.Type == IDENT && token.Value == "main" {
			foundMain = true
		}
		if token.Type == IDENT && token.Value == "x" {
			foundX = true
		}
		if token.Type == INT_LIT && token.Value == "42" {
			found42 = true
		}
	}
	
	if !foundMain {
		t.Error("expected to find 'main' identifier")
	}
	if !foundX {
		t.Error("expected to find 'x' identifier")
	}
	if !found42 {
		t.Error("expected to find '42' integer literal")
	}
}

func TestLexerPositionTracking(t *testing.T) {
	source := "int x;\nfloat y;"
	
	lexer := NewLexer(source, "test.c")
	tokens := lexer.Tokenize()
	
	// First token 'int' should be at line 1, column 1
	if tokens[0].Pos.Line != 1 || tokens[0].Pos.Column != 1 {
		t.Errorf("expected 'int' at 1:1, got %d:%d", tokens[0].Pos.Line, tokens[0].Pos.Column)
	}
	
	// Find 'float' token (should be on line 2)
	for _, token := range tokens {
		if token.Type == FLOAT {
			if token.Pos.Line != 2 {
				t.Errorf("expected 'float' on line 2, got line %d", token.Pos.Line)
			}
			return
		}
	}
	
	t.Error("expected to find 'float' token")
}

func TestLexerPreprocessor(t *testing.T) {
	tests := []struct {
		name   string
		source string
		tokenType TokenType
	}{
		{"hash", "#", PREPROCESSOR},
		{"concat", "##", CONCAT},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.source, "")
			tokens := lexer.Tokenize()
			
			if len(tokens) < 1 {
				t.Errorf("no tokens produced for %s", tt.source)
				return
			}
			
			if tokens[0].Type != tt.tokenType {
				t.Errorf("expected %s, got %s", tt.tokenType, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeString(t *testing.T) {
	source := "int x = 42;"
	tokens := TokenizeString(source)
	
	if len(tokens) < 4 {
		t.Errorf("expected at least 4 tokens, got %d", len(tokens))
		return
	}
	
	if tokens[0].Type != INT {
		t.Errorf("expected first token to be INT, got %s", tokens[0].Type)
	}
	
	if tokens[1].Type != IDENT || tokens[1].Value != "x" {
		t.Errorf("expected second token to be IDENT 'x', got %s %q", tokens[1].Type, tokens[1].Value)
	}
	
	if tokens[2].Type != ASSIGN {
		t.Errorf("expected third token to be ASSIGN, got %s", tokens[2].Type)
	}
	
	if tokens[3].Type != INT_LIT || tokens[3].Value != "42" {
		t.Errorf("expected fourth token to be INT_LIT '42', got %s %q", tokens[3].Type, tokens[3].Value)
	}
}

func TestLexerEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
		desc   string
	}{
		{"unicode identifier", "int 变量 = 42;", "should handle unicode identifiers"},
		{"multiple spaces", "int    x;", "should handle multiple spaces"},
		{"tabs", "int\tx;", "should handle tabs"},
		{"mixed whitespace", "int  \t  x;", "should handle mixed whitespace"},
		{"empty string literal", `""`, "should handle empty string"},
		{"escaped quotes", `"hello\"world"`, "should handle escaped quotes in strings"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.source, "")
			tokens := lexer.Tokenize()
			
			// Just verify it doesn't crash and produces some tokens
			if len(tokens) == 0 {
				t.Errorf("%s: no tokens produced", tt.desc)
			}
		})
	}
}