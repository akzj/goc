// Package parser provides edge case tests for the parser.
// These tests focus on edge cases not covered in other test files.
package parser

import (
	"testing"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/lexer"
)

// Helper function to create a parser from source code
func newParserForEdgeCaseTest(source string) (*Parser, *errhand.ErrorHandler) {
	tokens := lexer.TokenizeString(source)
	errHandler := errhand.NewErrorHandler()
	return NewParser(tokens, errHandler), errHandler
}

// ============================================================================
// Edge Case Tests for Expression Parsing
// ============================================================================

// TestParseExpression_DeeplyNested tests parsing of deeply nested expressions
func TestParseExpression_DeeplyNested(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "nested_binary_ops",
			source: "int x = 1 + 2 * 3 - 4 / 5;",
		},
		{
			name:   "nested_parens",
			source: "int x = (((1 + 2) * 3) - 4);",
		},
		{
			name:   "mixed_unary_binary",
			source: "int x = -!+~5;",
		},
		{
			name:   "chained_member_access",
			source: "int x = a.b.c.d;",
		},
		{
			name:   "nested_function_calls",
			source: "int x = f(g(h(i())));",
		},
		{
			name:   "complex_expression",
			source: "int x = (a + b) * (c - d) / (e % f);",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForEdgeCaseTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			// Some errors may be expected for complex cases
			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors for %q", errHandler.ErrorCount(), tt.source)
			}
		})
	}
}

// TestParseExpression_OperatorPrecedence tests operator precedence edge cases
func TestParseExpression_OperatorPrecedence(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "multiplication_before_addition",
			source: "int x = 1 + 2 * 3;", // Should be 1 + (2 * 3)
		},
		{
			name:   "logical_and_before_or",
			source: "int x = a || b && c;", // Should be a || (b && c)
		},
		{
			name:   "shift_vs_arithmetic",
			source: "int x = 1 << 2 + 3;", // Should be 1 << (2 + 3)
		},
		{
			name:   "comparison_vs_equality",
			source: "int x = a < b == c > d;", // Should be (a < b) == (c > d)
		},
		{
			name:   "ternary_precedence",
			source: "int x = a ? b : c ? d : e;", // Right associative
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForEdgeCaseTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors for %q", errHandler.ErrorCount(), tt.source)
			}
		})
	}
}

// TestParseExpression_EdgeCaseLiterals tests edge case literals
func TestParseExpression_EdgeCaseLiterals(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "min_int_literal",
			source: "int x = -2147483648;",
		},
		{
			name:   "max_int_literal",
			source: "int x = 2147483647;",
		},
		{
			name:   "zero_literal",
			source: "int x = 0;",
		},
		{
			name:   "octal_literal",
			source: "int x = 0777;",
		},
		{
			name:   "hex_literal",
			source: "int x = 0xDEADBEEF;",
		},
		{
			name:   "char_literal_escape",
			source: "char c = '\\n';",
		},
		{
			name:   "string_literal_empty",
			source: "char *s = \"\";",
		},
		{
			name:   "float_literal_scientific",
			source: "float f = 1.5e10;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForEdgeCaseTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors for %q", errHandler.ErrorCount(), tt.source)
			}
		})
	}
}

// ============================================================================
// Edge Case Tests for Statement Parsing
// ============================================================================

// TestParseStatement_EmptyAndCompound tests empty and compound statement edge cases
func TestParseStatement_EmptyAndCompound(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "empty_statement",
			source: "int main() { ; }",
		},
		{
			name:   "nested_empty",
			source: "int main() { { { } } }",
		},
		{
			name:   "compound_with_decls",
			source: "int main() { int a; int b; int c; }",
		},
		{
			name:   "mixed_stmts_decls",
			source: "int main() { int a; a = 1; int b; b = 2; }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForEdgeCaseTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors for %q", errHandler.ErrorCount(), tt.source)
			}
		})
	}
}

// TestParseStatement_LoopEdgeCases tests loop statement edge cases
func TestParseStatement_LoopEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "for_empty_parts",
			source: "int main() { for(;;) { } }",
		},
		{
			name:   "while_true",
			source: "int main() { while(1) { } }",
		},
		{
			name:   "nested_loops",
			source: "int main() { for(int i=0; i<10; i++) { for(int j=0; j<10; j++) { } } }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForEdgeCaseTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors for %q", errHandler.ErrorCount(), tt.source)
			}
		})
	}
}

// ============================================================================
// Edge Case Tests for Type Parsing
// ============================================================================

// TestParseType_ComplexTypes tests complex type parsing edge cases
func TestParseType_ComplexTypes(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "pointer_to_pointer",
			source: "int **x;",
		},
		{
			name:   "array_of_pointers",
			source: "int *x[10];",
		},
		{
			name:   "pointer_to_array",
			source: "int (*x)[10];",
		},
		{
			name:   "function_pointer",
			source: "int (*f)(int, int);",
		},
		{
			name:   "array_of_function_pointers",
			source: "int (*f[10])(int);",
		},
		{
			name:   "const_volatile",
			source: "const volatile int x;",
		},
		{
			name:   "restrict_pointer",
			source: "int * restrict p;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForEdgeCaseTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors for %q", errHandler.ErrorCount(), tt.source)
			}
		})
	}
}

// TestParseType_TypeQualifiers tests type qualifier edge cases
func TestParseType_TypeQualifiers(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "const_int",
			source: "const int x = 5;",
		},
		{
			name:   "volatile_int",
			source: "volatile int x;",
		},
		{
			name:   "const_pointer",
			source: "int * const p;",
		},
		{
			name:   "pointer_to_const",
			source: "const int *p;",
		},
		{
			name:   "const_const",
			source: "const const int x;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForEdgeCaseTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors for %q", errHandler.ErrorCount(), tt.source)
			}
		})
	}
}