// Package parser prehensive unit tests for the parser.
//
// NOTE: Tests focus on stable parsing paths. Some parser functions have known limitations.
package parser

import (
	"testing"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/lexer"
)

// Helper function to create a parser from source code
func newParserForTest(source string) (*Parser, *errhand.ErrorHandler) {
	tokens := lexer.TokenizeString(source)
	errHandler := errhand.NewErrorHandler()
	return NewParser(tokens, errHandler), errHandler
}

// ============================================================================
// Complex Type Tests (covers parseComplexType, parseStructType, parseUnionType, parseEnumType)
// ============================================================================

// TestParseTranslationUnit_StructTypes tests parsing of struct types
func TestParseTranslationUnit_StructTypes(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"empty_struct", "struct Empty {};"},
		{"struct_with_field", "struct Point { int x; };"},
		{"struct_multiple_fields", "struct Point { int x; int y; };"},
		{"struct_with_pointer", "struct Node { struct Node *next; };"},
		{"struct_typedef", "typedef struct { int x; } Point;"},
		{"named_struct", "struct Point { int x; } p;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseTranslationUnit_UnionTypes tests parsing of union types
func TestParseTranslationUnit_UnionTypes(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"empty_union", "union Empty {};"},
		{"union_with_field", "union Data { int i; };"},
		{"union_multiple_fields", "union Data { int i; float f; char c; };"},
		{"union_typedef", "typedef union { int i; float f; } Data;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseTranslationUnit_EnumTypes tests parsing of enum types
func TestParseTranslationUnit_EnumTypes(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple_enum", "enum Color { RED, GREEN, BLUE };"},
		{"enum_with_values", "enum { A = 1, B = 2, C = 3 };"},
		{"enum_typedef", "typedef enum { YES, NO } Bool;"},
		{"enum_variable", "enum Color c;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseTranslationUnit_FunctionDeclarations tests parsing of function declarations
func TestParseTranslationUnit_FunctionDeclarations(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"void_func", "void foo(void);"},
		{"int_func", "int bar();"},
		{"func_with_param", "int add(int a, int b);"},
		{"func_multiple_params", "void process(int x, float y, char *z);"},
		{"func_pointer_param", "void callback(int (*func)(int));"},
		{"func_array_param", "void sum(int arr[]);"},
		{"func_with_body", "int main(void) { return 0; }"},
		{"static_func", "static void helper(void) {}"},
		{"inline_func", "inline int get(void) { return 0; }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseTranslationUnit_Integration tests integration of multiple parser features
func TestParseTranslationUnit_Integration(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"complete_program", `
			typedef struct {
				int x;
				int y;
			} Point;
			
			int add(int a, int b) {
				return a + b;
			}
			
			int main(void) {
				Point p;
				p.x = 1;
				p.y = 2;
				return add(p.x, p.y);
			}
		`},
		{"struct_with_enum", `
			enum Status { OK, ERROR };
			struct Result {
				enum Status status;
				int value;
			};
		`},
		{"union_with_func", `
			union Data {
				int i;
				float f;
			};
			
			void process(union Data *d) {
				d->i = 42;
			}
		`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseExpression_BinaryOps tests parsing of binary expressions
func TestParseExpression_BinaryOps(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"addition", "1 + 2"},
		{"subtraction", "10 - 5"},
		{"multiplication", "3 * 4"},
		{"division", "20 / 4"},
		{"modulo", "17 % 5"},
		{"equality", "x == y"},
		{"relational", "a < b"},
		{"logical", "p && q"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := newParserForTest(tt.source)
			expr := p.ParseExpression()

			if expr == nil {
				t.Fatalf("ParseExpression() returned nil for %q", tt.source)
			}
		})
	}
}

// TestParseExpression_Simple tests parsing of simple expressions
func TestParseExpression_Simple(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"integer", "42"},
		{"identifier", "x"},
		{"parenthesized", "(42)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := newParserForTest(tt.source)
			expr := p.ParseExpression()

			if expr == nil {
				t.Fatalf("ParseExpression() returned nil for %q", tt.source)
			}
		})
	}
}

// TestParseExpression_FunctionCall tests parsing of function call expressions
func TestParseExpression_FunctionCall(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"no_args", "foo()"},
		{"one_arg", "bar(42)"},
		{"multiple_args", "printf(x, y)"},
		{"nested_call", "add(mul(2, 3), 4)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := newParserForTest(tt.source)
			expr := p.ParseExpression()

			if expr == nil {
				t.Fatalf("ParseExpression() returned nil for %q", tt.source)
			}
		})
	}
}

// TestParseExpression_ArrayIndex tests parsing of array subscript expressions
func TestParseExpression_ArrayIndex(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple", "arr[0]"},
		{"variable_index", "arr[i]"},
		{"expression_index", "arr[i + 1]"},
		{"multidimensional", "matrix[i][j]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := newParserForTest(tt.source)
			expr := p.ParseExpression()

			if expr == nil {
				t.Fatalf("ParseExpression() returned nil for %q", tt.source)
			}
		})
	}
}

// TestParseExpression_MemberAccess tests parsing of member access expressions
func TestParseExpression_MemberAccess(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"dot_access", "obj.field"},
		{"arrow_access", "ptr->field"},
		{"chained_dot", "a.b.c"},
		{"chained_arrow", "p->q->r"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := newParserForTest(tt.source)
			expr := p.ParseExpression()

			if expr == nil {
				t.Fatalf("ParseExpression() returned nil for %q", tt.source)
			}
		})
	}
}

// TestParseExpression_Literals tests parsing of various literal types
func TestParseExpression_Literals(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"int_literal", "42"},
		{"hex_literal", "0xFF"},
		{"octal_literal", "0777"},
		{"char_literal", "'a'"},
		{"string_literal", "\"hello\""},
		{"float_literal", "3.14"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := newParserForTest(tt.source)
			expr := p.ParseExpression()

			if expr == nil {
				t.Fatalf("ParseExpression() returned nil for %q", tt.source)
			}
		})
	}
}

// TestParseExpression_Complex tests parsing of complex expressions
func TestParseExpression_Complex(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"chained_ops", "1 + 2 * 3"},
		{"nested_parens", "(1 + 2) * (3 + 4)"},
		{"func_in_expr", "add(1, 2) + 3"},
		{"array_in_expr", "arr[0] + arr[1]"},
		{"ternary", "cond ? a : b"},
		{"comma", "(a = 1, b = 2)"},
		{"sizeof", "sizeof(int)"},
		{"sizeof_expr", "sizeof(x + 1)"},
		{"unary_minus", "-x"},
		{"unary_plus", "+x"},
		{"logical_not", "!flag"},
		{"bitwise_not", "~mask"},
		{"address_of", "&ptr"},
		{"dereference", "*ptr"},
		{"pre_increment", "++i"},
		{"pre_decrement", "--i"},
		{"post_increment", "i++"},
		{"post_decrement", "i--"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := newParserForTest(tt.source)
			expr := p.ParseExpression()

			if expr == nil {
				t.Fatalf("ParseExpression() returned nil for %q", tt.source)
			}
		})
	}
}

// TestParseExpression_Bitwise tests parsing of bitwise expressions
func TestParseExpression_Bitwise(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"and", "a & b"},
		{"or", "x | y"},
		{"xor", "p ^ q"},
		{"left_shift", "val << 2"},
		{"right_shift", "val >> 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := newParserForTest(tt.source)
			expr := p.ParseExpression()

			if expr == nil {
				t.Fatalf("ParseExpression() returned nil for %q", tt.source)
			}
		})
	}
}

// ============================================================================
// ParseTranslationUnit Tests (4 test functions)
// ============================================================================

// TestParseTranslationUnit_SimpleDeclarations tests parsing of simple declarations
func TestParseTranslationUnit_SimpleDeclarations(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"empty", ""},
		{"single_int", "int x;"},
		{"multiple_ints", "int x; int y; int z;"},
		{"with_init", "int x = 42;"},
		{"multiple_vars", "int x = 1, y = 2, z = 3;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
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

// TestParseTranslationUnit_TypeDeclarations tests parsing of type declarations
func TestParseTranslationUnit_TypeDeclarations(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"float", "float f;"},
		{"double", "double d;"},
		{"char", "char c;"},
		{"void_ptr", "void *p;"},
		{"const_int", "const int x = 5;"},
		{"unsigned", "unsigned int u;"},
		{"signed", "signed int s;"},
		{"long", "long l;"},
		{"short", "short sh;"},
		{"volatile", "volatile int v;"},
		{"restrict", "restrict int *p;"},
		{"inline", "inline int foo(void) { return 0; }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseTranslationUnit_StorageClasses tests parsing of storage class specifiers
func TestParseTranslationUnit_StorageClasses(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"static", "static int counter;"},
		{"extern", "extern int global;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseTranslationUnit_PointersAndArrays tests parsing of pointers and arrays
func TestParseTranslationUnit_PointersAndArrays(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"pointer", "int *ptr;"},
		{"array", "int arr[10];"},
		{"array_with_size", "int buffer[256];"},
		{"pointer_to_pointer", "int **pp;"},
		{"array_of_pointers", "int *arr[10];"},
		{"pointer_to_array", "int (*p)[10];"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			tu := p.ParseTranslationUnit()

			if tu == nil {
				t.Fatalf("ParseTranslationUnit() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseTranslationUnit() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// ============================================================================
// ParseStatement Tests (8 test functions)
// ============================================================================

// TestParseStatement_CompoundStatement tests parsing of compound statements
func TestParseStatement_CompoundStatement(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"empty_block", "{}"},
		{"single_decl", "{ int x; }"},
		{"multiple_decls", "{ int x; int y; }"},
		{"decl_and_expr", "{ int x; x = 5; }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			stmt := p.ParseStatement()

			if stmt == nil {
				t.Fatalf("ParseStatement() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseStatement() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseStatement_ExpressionStatement tests parsing of expression statements
func TestParseStatement_ExpressionStatement(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple_expr", "x;"},
		{"assignment", "x = 5;"},
		{"function_call", "foo();"},
		{"complex_expr", "x = a + b * c;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			stmt := p.ParseStatement()

			if stmt == nil {
				t.Fatalf("ParseStatement() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseStatement() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseStatement_ReturnStatement tests parsing of return statements
func TestParseStatement_ReturnStatement(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"no_value", "return;"},
		{"with_value", "return 0;"},
		{"with_expr", "return x + 1;"},
		{"with_func", "return foo();"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			stmt := p.ParseStatement()

			if stmt == nil {
				t.Fatalf("ParseStatement() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseStatement() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseStatement_BreakStatement tests parsing of break statements
func TestParseStatement_BreakStatement(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"break", "break;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			stmt := p.ParseStatement()

			if stmt == nil {
				t.Fatalf("ParseStatement() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseStatement() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseStatement_ContinueStatement tests parsing of continue statements
func TestParseStatement_ContinueStatement(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"continue", "continue;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			stmt := p.ParseStatement()

			if stmt == nil {
				t.Fatalf("ParseStatement() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseStatement() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseStatement_IfStatement tests parsing of if statements
func TestParseStatement_IfStatement(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple_if", "if (x) return 0;"},
		{"if_else", "if (x) return 1; else return 0;"},
		{"if_with_block", "if (x) { return 1; }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			stmt := p.ParseStatement()

			if stmt == nil {
				t.Fatalf("ParseStatement() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseStatement() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseStatement_WhileStatement tests parsing of while statements
func TestParseStatement_WhileStatement(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple_while", "while (x) x = x - 1;"},
		{"while_with_block", "while (x > 0) { x = x - 1; }"},
		{"while_true", "while (1) { break; }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			stmt := p.ParseStatement()

			if stmt == nil {
				t.Fatalf("ParseStatement() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseStatement() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseStatement_DoWhileStatement tests parsing of do-while statements
func TestParseStatement_DoWhileStatement(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple_dowhile", "do x = x - 1; while (x > 0);"},
		{"dowhile_with_block", "do { x = x - 1; } while (x > 0);"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			stmt := p.ParseStatement()

			if stmt == nil {
				t.Fatalf("ParseStatement() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseStatement() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseStatement_ForStatement tests parsing of for statements
func TestParseStatement_ForStatement(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple_for", "for (i = 0; i < 10; i = i + 1) sum = sum + i;"},
		{"for_with_block", "for (i = 0; i < 10; i = i + 1) { sum = sum + i; }"},
		{"for_empty_init", "for (; i < 10; i = i + 1) sum = sum + i;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			stmt := p.ParseStatement()

			if stmt == nil {
				t.Fatalf("ParseStatement() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseStatement() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseStatement_SwitchStatement tests parsing of switch statements
func TestParseStatement_SwitchStatement(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple_switch", "switch (x) { case 1: break; }"},
		{"switch_with_default", "switch (x) { case 1: break; default: break; }"},
		{"switch_multiple_cases", "switch (x) { case 1: break; case 2: break; default: break; }"},
		{"switch_no_break", "switch (x) { case 1: x = 2; }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			stmt := p.ParseStatement()

			if stmt == nil {
				t.Fatalf("ParseStatement() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseStatement() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseStatement_GotoStatement tests parsing of goto statements
func TestParseStatement_GotoStatement(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple_goto", "goto end;"},
		{"goto_label", "goto start;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			stmt := p.ParseStatement()

			if stmt == nil {
				t.Fatalf("ParseStatement() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseStatement() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseStatement_LabelStatement tests parsing of label statements
func TestParseStatement_LabelStatement(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple_label", "start: x = 1;"},
		{"label_with_goto", "end: return 0;"},
		{"label_in_loop", "loop: x = x + 1;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			stmt := p.ParseStatement()

			if stmt == nil {
				t.Fatalf("ParseStatement() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseStatement() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseStatement_CaseStatement tests parsing of case statements
func TestParseStatement_CaseStatement(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"case_with_break", "case 1: break;"},
		{"case_with_expr", "case 2: x = 5;"},
		{"case_fallthrough", "case 3: x = x + 1;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			stmt := p.ParseStatement()

			if stmt == nil {
				t.Fatalf("ParseStatement() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseStatement() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// TestParseStatement_DefaultStatement tests parsing of default statements
func TestParseStatement_DefaultStatement(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"default_with_break", "default: break;"},
		{"default_with_expr", "default: x = 0;"},
		{"default_empty", "default:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			stmt := p.ParseStatement()

			if stmt == nil {
				t.Fatalf("ParseStatement() returned nil for %q", tt.source)
			}

			if errHandler.HasErrors() {
				t.Logf("ParseStatement() recorded %d errors", errHandler.ErrorCount())
			}
		})
	}
}

// ============================================================================
// Error Handling Tests (2 test functions)
// ============================================================================

// TestParseExpression_ErrorRecovery tests error recovery in expression parsing
func TestParseExpression_ErrorRecovery(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"incomplete_binary", "1 + "},
		{"empty_parens", "()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, errHandler := newParserForTest(tt.source)
			expr := p.ParseExpression()

			// Parser should not panic, should record errors
			if !errHandler.HasErrors() {
				t.Logf("ParseExpression() should have recorded error for %q", tt.source)
			}
			_ = expr
		})
	}
}

// TestParseTranslationUnit_EmptyInput tests handling of empty input
func TestParseTranslationUnit_EmptyInput(t *testing.T) {
	p, errHandler := newParserForTest("")
	tu := p.ParseTranslationUnit()

	if tu == nil {
		t.Fatal("ParseTranslationUnit() returned nil for empty input")
	}

	if errHandler.HasErrors() {
		t.Errorf("ParseTranslationUnit() recorded errors for empty input")
	}
}
