// Package parser tests for char** (pointer-to-pointer) support
package parser

import (
	"testing"
)

// TestParseCharPointerToPointer tests parsing of char** and char*** declarations
func TestParseCharPointerToPointer(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"char_double_pointer", "char **argv;"},
		{"char_triple_pointer", "char ***ptr;"},
		{"main_signature", "int main(int argc, char **argv) {}"},
		{"function_with_charpp", "void func(char **args) {}"},
		{"const_char_double_pointer", "const char **argv;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Errorf("ParseTranslationUnit() recorded %d errors for %q: %v", 
					errHandler.ErrorCount(), tt.source, errHandler.Errors())
			}
		})
	}
}

// TestParseCharPointerToPointer_VerifyAST verifies the nested PointerType structure
func TestParseCharPointerToPointer_VerifyAST(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectDepth    int // Expected pointer depth (1 for *, 2 for **, etc.)
		expectBaseType string // Expected base type name
	}{
		{"char_star", "char *p;", 1, "char"},
		{"char_star_star", "char **p;", 2, "char"},
		{"char_star_star_star", "char ***p;", 3, "char"},
		{"int_star_star", "int **p;", 2, "int"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			tu := p.ParseTranslationUnit()

			if errHandler.HasErrors() {
				t.Fatalf("ParseTranslationUnit() recorded errors: %v", errHandler.Errors())
			}

			if len(tu.Declarations) == 0 {
				t.Fatal("Expected at least one declaration")
			}

			// Get the variable declaration
			varDecl, ok := tu.Declarations[0].(*VarDecl)
			if !ok {
				t.Fatalf("Expected VarDecl, got %T", tu.Declarations[0])
			}

			// Verify pointer depth
			depth := getPointerDepth(varDecl.Type)
			if depth != tt.expectDepth {
				t.Errorf("Pointer depth = %d, want %d for %q", depth, tt.expectDepth, tt.source)
			}

			// Verify base type
			baseType := getBaseTypeName(varDecl.Type)
			if baseType != tt.expectBaseType {
				t.Errorf("Base type = %q, want %q for %q", baseType, tt.expectBaseType, tt.source)
			}
		})
	}
}

// getPointerDepth returns the depth of pointer nesting
func getPointerDepth(typ Type) int {
	depth := 0
	for typ != nil {
		if pt, ok := typ.(*PointerType); ok {
			depth++
			typ = pt.Elem
		} else {
			break
		}
	}
	return depth
}

// getBaseTypeName returns the name of the base type
func getBaseTypeName(typ Type) string {
	for typ != nil {
		if pt, ok := typ.(*PointerType); ok {
			typ = pt.Elem
		} else if qt, ok := typ.(*QualifiedType); ok {
			typ = qt.Type
		} else if bt, ok := typ.(*BaseType); ok {
			return bt.String()
		} else {
			return typ.String()
		}
	}
	return ""
}
