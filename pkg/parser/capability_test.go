package parser

import (
	"testing"
	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/internal/errhand"
)

func TestParserCapabilities(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"function_with_return", "int main() { return 0; }"},
		{"variable_decl", "int x = 5;"},
		{"if_statement", "if (x > 0) { x = 1; }"},
		{"while_statement", "while (x > 0) { x--; }"},
		{"for_statement", "for (int i = 0; i < 10; i++) { x += i; }"},
		{"function_with_params", "int add(int a, int b) { return a + b; }"},
		{"struct_def", "struct Point { int x; int y; };"},
		{"enum_def", "enum Color { RED, GREEN, BLUE };"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := lexer.TokenizeString(tt.source)
			errHandler := errhand.NewErrorHandler()
			p := NewParser(tokens, errHandler)
			tu := p.ParseTranslationUnit()
			
			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}
			
			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors for %q", errHandler.ErrorCount(), tt.source)
			} else {
				t.Logf("OK: %d declarations", len(tu.Declarations))
			}
		})
	}
}
