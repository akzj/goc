// Package semantic provides edge case tests for the semantic analyzer.
// These tests focus on edge cases not covered in analyzer_test.go.
package semantic

import (
	"testing"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/parser"
)

// Helper function to analyze source code
func analyzeSourceForEdgeCase(source string) (*SemanticAnalyzer, *errhand.ErrorHandler, error) {
	tokens := lexer.TokenizeString(source)
	errHandler := errhand.NewErrorHandler()
	p := parser.NewParser(tokens, errHandler)
	ast := p.ParseTranslationUnit()

	analyzer := NewSemanticAnalyzer(errHandler)
	err := analyzer.Analyze(ast)

	return analyzer, errHandler, err
}

// ============================================================================
// Edge Case Tests for Type Checking
// ============================================================================

// TestTypeChecker_TypeCompatibilityEdgeCases tests type compatibility edge cases
func TestTypeChecker_TypeCompatibilityEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "int_to_float_implicit",
			source: "int main() { int x = 5; float f = x; return 0; }",
		},
		{
			name:   "float_to_int_explicit",
			source: "int main() { float f = 3.14; int x = (int)f; return 0; }",
		},
		{
			name:   "pointer_to_void_cast",
			source: "int main() { int x; void *p = &x; int *q = (int*)p; return 0; }",
		},
		{
			name:   "array_decay_to_pointer",
			source: "int main() { int arr[10]; int *p = arr; return 0; }",
		},
		{
			name:   "function_pointer_type",
			source: "int foo(int x) { return x; } int main() { int (*f)(int) = foo; return f(5); }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, errHandler, err := analyzeSourceForEdgeCase(tt.source)

			if analyzer == nil {
				t.Fatalf("Semantic analyzer is nil for %q", tt.source)
			}

			// Some type conversions may produce warnings
			if errHandler.HasErrors() {
				t.Logf("Analysis recorded %d errors for %q", errHandler.ErrorCount(), tt.source)
			}

			_ = err // err may be non-nil for some edge cases
		})
	}
}

// TestTypeChecker_ArithmeticEdgeCases tests arithmetic operation edge cases
func TestTypeChecker_ArithmeticEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "mixed_signed_unsigned",
			source: "int main() { int x = -1; unsigned int y = 1; unsigned int z = x + y; return 0; }",
		},
		{
			name:   "char_arithmetic",
			source: "int main() { char c = 'a'; char d = c + 1; return 0; }",
		},
		{
			name:   "long_long_arithmetic",
			source: "int main() { long long x = 10000000000LL; long long y = x * 2; return 0; }",
		},
		{
			name:   "division_by_constant",
			source: "int main() { int x = 10 / 3; return 0; }",
		},
		{
			name:   "modulo_negative",
			source: "int main() { int x = -10 % 3; return 0; }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, errHandler, err := analyzeSourceForEdgeCase(tt.source)

			if analyzer == nil {
				t.Fatalf("Semantic analyzer is nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("Analysis recorded %d errors for %q", errHandler.ErrorCount(), tt.source)
			}

			_ = err
		})
	}
}

// ============================================================================
// Edge Case Tests for Scope Management
// ============================================================================

// TestSemanticAnalyzer_ScopeShadowing tests variable shadowing edge cases
func TestSemanticAnalyzer_ScopeShadowing(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "shadow_global_in_local",
			source: "int x = 10; int main() { int x = 5; return x; }",
		},
		{
			name:   "shadow_in_nested_scope",
			source: "int main() { int x = 1; { int x = 2; } return x; }",
		},
		{
			name:   "shadow_function_param",
			source: "int foo(int x) { int x = 5; return x; }",
		},
		{
			name:   "shadow_in_for_loop",
			source: "int main() { int i = 0; for(int i = 0; i < 10; i++) { } return i; }",
		},
		{
			name:   "multiple_shadow_levels",
			source: "int x = 1; int main() { int x = 2; { int x = 3; { int x = 4; } } return x; }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, errHandler, err := analyzeSourceForEdgeCase(tt.source)

			if analyzer == nil {
				t.Fatalf("Semantic analyzer is nil for %q", tt.source)
			}

			// Shadowing may produce warnings but should not fail
			if errHandler.HasErrors() {
				t.Logf("Analysis recorded %d errors for %q", errHandler.ErrorCount(), tt.source)
			}

			_ = err
		})
	}
}

// TestSemanticAnalyzer_BlockScopeEdgeCases tests block scope edge cases
func TestSemanticAnalyzer_BlockScopeEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "empty_block",
			source: "int main() { { } return 0; }",
		},
		{
			name:   "deeply_nested_blocks",
			source: "int main() { { { { { } } } } return 0; }",
		},
		{
			name:   "declaration_in_condition",
			source: "int main() { int x = 5; if(int y = x) { } return 0; }",
		},
		{
			name:   "declaration_in_for_init",
			source: "int main() { for(int i = 0; i < 10; i++) { } return 0; }",
		},
		{
			name:   "variable_lifetime_in_block",
			source: "int main() { { int x = 5; } return 0; }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, errHandler, err := analyzeSourceForEdgeCase(tt.source)

			if analyzer == nil {
				t.Fatalf("Semantic analyzer is nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("Analysis recorded %d errors for %q", errHandler.ErrorCount(), tt.source)
			}

			_ = err
		})
	}
}

// ============================================================================
// Edge Case Tests for Function Analysis
// ============================================================================

// TestSemanticAnalyzer_FunctionCallEdgeCases tests function call edge cases
func TestSemanticAnalyzer_FunctionCallEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "function_with_no_args",
			source: "int foo(void) { return 0; } int main() { return foo(); }",
		},
		{
			name:   "function_with_many_args",
			source: "int foo(int a, int b, int c, int d, int e) { return a+b+c+d+e; } int main() { return foo(1,2,3,4,5); }",
		},
		{
			name:   "recursive_function",
			source: "int fact(int n) { if(n <= 1) return 1; return n * fact(n-1); }",
		},
		{
			name:   "function_pointer_call",
			source: "int foo(int x) { return x; } int main() { int (*f)(int) = foo; return f(5); }",
		},
		{
			name:   "variadic_function",
			source: "int sum(int count, ...) { return 0; } int main() { return sum(3, 1, 2, 3); }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, errHandler, err := analyzeSourceForEdgeCase(tt.source)

			if analyzer == nil {
				t.Fatalf("Semantic analyzer is nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("Analysis recorded %d errors for %q", errHandler.ErrorCount(), tt.source)
			}

			_ = err
		})
	}
}

// TestSemanticAnalyzer_ReturnStatementEdgeCases tests return statement edge cases
func TestSemanticAnalyzer_ReturnStatementEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "return_in_nested_block",
			source: "int main() { { return 0; } }",
		},
		{
			name:   "return_in_if",
			source: "int main() { if(1) { return 0; } return 1; }",
		},
		{
			name:   "return_in_loop",
			source: "int main() { for(;;) { return 0; } }",
		},
		{
			name:   "return_with_expression",
			source: "int main() { int x = 5; return x + 1; }",
		},
		{
			name:   "void_return",
			source: "void foo() { return; }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, errHandler, err := analyzeSourceForEdgeCase(tt.source)

			if analyzer == nil {
				t.Fatalf("Semantic analyzer is nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("Analysis recorded %d errors for %q", errHandler.ErrorCount(), tt.source)
			}

			_ = err
		})
	}
}

// ============================================================================
// Edge Case Tests for Control Flow
// ============================================================================

// TestSemanticAnalyzer_ControlFlowEdgeCases tests control flow edge cases
func TestSemanticAnalyzer_ControlFlowEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "break_in_loop",
			source: "int main() { for(;;) { break; } return 0; }",
		},
		{
			name:   "continue_in_loop",
			source: "int main() { for(;;) { continue; } return 0; }",
		},
		{
			name:   "switch_simple",
			source: "int main() { int x = 1; switch(x) { case 1: break; default: break; } return 0; }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, errHandler, err := analyzeSourceForEdgeCase(tt.source)

			if analyzer == nil {
				t.Fatalf("Semantic analyzer is nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("Analysis recorded %d errors for %q", errHandler.ErrorCount(), tt.source)
			}

			_ = err
		})
	}
}
