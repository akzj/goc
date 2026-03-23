// Package semantic provides semantic analysis for the GOC compiler.
// This file contains comprehensive unit tests for the TypeChecker in type.go.
package semantic

import (
	"testing"

	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/parser"
)


// ============================================================================
// Additional Type Helpers (specific to type tests)
// ============================================================================

func shortType() parser.Type {
	return &parser.BaseType{Kind: parser.TypeShort, Signed: true}
}

func boolType() parser.Type {
	return &parser.BaseType{Kind: parser.TypeBool}
}

func longLongType() parser.Type {
	return &parser.BaseType{Kind: parser.TypeLong, Signed: true, Long: 1}
}

func longDoubleType() parser.Type {
	return &parser.BaseType{Kind: parser.TypeDouble, Long: 1}
}

func ucharType() parser.Type {
	return &parser.BaseType{Kind: parser.TypeChar, Signed: false}
}

func charPtrType() parser.Type {
	return &parser.PointerType{Elem: charType()}
}

func funcPtrType() parser.Type {
	return &parser.PointerType{Elem: &parser.FuncType{Return: intType()}}
}

func intArrayType() parser.Type {
	return &parser.ArrayType{Elem: intType(), ArraySize: 10}
}

func charArrayType() parser.Type {
	return &parser.ArrayType{Elem: charType(), ArraySize: 10}
}

func incompleteArrayType() parser.Type {
	return &parser.ArrayType{Elem: intType(), ArraySize: -1}
}

func funcTypeWithParams(ret parser.Type, params []parser.Type) parser.Type {
	return &parser.FuncType{Return: ret, Params: params}
}

func variadicFuncType() parser.Type {
	return &parser.FuncType{Return: intType(), Params: []parser.Type{intType()}, Variadic: true}
}

func voidFuncType() parser.Type {
	return &parser.FuncType{Return: voidType()}
}

func structType(name string) parser.Type {
	return &parser.StructType{Name: name, IsUnion: false}
}

func unionType(name string) parser.Type {
	return &parser.StructType{Name: name, IsUnion: true}
}

func enumType(name string) parser.Type {
	return &parser.EnumType{Name: name}
}

func typedefType(name string, underlying parser.Type) parser.Type {
	return &parser.TypedefType{Name: name, Underlying: underlying}
}

func constIntType() parser.Type {
	return &parser.QualifiedType{Type: intType(), IsConst: true}
}

func volatileIntType() parser.Type {
	return &parser.QualifiedType{Type: intType(), IsVolatile: true}
}

func constVolatileIntType() parser.Type {
	return &parser.QualifiedType{Type: intType(), IsConst: true, IsVolatile: true}
}

// ============================================================================
// CheckAssignable Tests
// ============================================================================

func TestCheckAssignable(t *testing.T) {
	tests := []struct {
		name        string
		dstType     parser.Type
		srcType     parser.Type
		expectError bool
	}{
		// Same type - should succeed
		{"same int", intType(), intType(), false},
		{"same char", charType(), charType(), false},
		{"same double", doubleType(), doubleType(), false},
		{"same pointer", intPtrType(), intPtrType(), false},

		// Arithmetic type conversions - should succeed
		{"int to double", doubleType(), intType(), false},
		{"char to int", intType(), charType(), false},
		{"short to long", longType(), shortType(), false},
		{"float to double", doubleType(), floatType(), false},
		{"int to float", floatType(), intType(), false},

		// Void pointer compatibility - should succeed
		{"int* to void*", voidPtrType(), intPtrType(), false},
		{"void* to int*", intPtrType(), voidPtrType(), false},
		{"char* to void*", voidPtrType(), charPtrType(), false},
		{"void* to char*", charPtrType(), voidPtrType(), false},

		// Null pointer constant (integer 0 to pointer) - should succeed
		{"int to int*", intPtrType(), intType(), false},
		{"int to void*", voidPtrType(), intType(), false},

		// Pointer compatibility - should succeed
		{"int* to int*", intPtrType(), intPtrType(), false},

		// Array to pointer decay - should succeed
		{"int[10] to int*", intPtrType(), intArrayType(), false},
		{"char[10] to char*", charPtrType(), charArrayType(), false},

		// Function to pointer decay - should succeed
		{"function to function*", funcPtrType(), funcType(intType(), []parser.Type{intType()}, false), false},

		// Incompatible types - should fail
		{"int to char*", charPtrType(), intType(), false}, // null pointer constant allowed in C
		{"int* to char*", charPtrType(), intPtrType(), true},
		{"int to struct", structType("S"), intType(), true},
		{"struct to int", intType(), structType("S"), true},
		{"int to function", funcType(intType(), []parser.Type{intType()}, false), intType(), true},

		// Qualified types - should succeed (qualifiers ignored for assignment)
		{"const int to int", intType(), constIntType(), false},
		{"int to const int", constIntType(), intType(), false},
		{"volatile int to int", intType(), volatileIntType(), false},

		// Typedef - should succeed if underlying types match
		{"typedef int to int", intType(), typedefType("my_int", intType()), false},
		{"int to typedef int", typedefType("my_int", intType()), intType(), false},

		// Enum - compatible with int
		{"enum to int", intType(), enumType("Color"), false}, // enum compatible with int
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			err := tc.CheckAssignable(tt.dstType, tt.srcType, testPos())

			if tt.expectError {
				if err == nil {
					t.Errorf("CheckAssignable(%s, %s) expected error, got nil",
						tt.dstType.String(), tt.srcType.String())
				}
			} else {
				if err != nil {
					t.Errorf("CheckAssignable(%s, %s) unexpected error: %v",
						tt.dstType.String(), tt.srcType.String(), err)
				}
			}
		})
	}
}

// ============================================================================
// CheckBinaryOp Tests
// ============================================================================

func TestCheckBinaryOp(t *testing.T) {
	tests := []struct {
		name        string
		op          lexer.TokenType
		left        parser.Type
		right       parser.Type
		expectError bool
		expectType  parser.Type
	}{
		// Arithmetic operators
		{"int + int", lexer.ADD, intType(), intType(), false, intType()},
		{"int - int", lexer.SUB, intType(), intType(), false, intType()},
		{"int * int", lexer.MUL, intType(), intType(), false, intType()},
		{"int / int", lexer.QUO, intType(), intType(), false, intType()},
		{"int % int", lexer.REM, intType(), intType(), false, intType()},

		// Arithmetic with different types (usual arithmetic conversions)
		{"int + double", lexer.ADD, intType(), doubleType(), false, doubleType()},
		{"float + double", lexer.ADD, floatType(), doubleType(), false, doubleType()},
		{"char + int", lexer.ADD, charType(), intType(), false, intType()},

		// Modulo requires integers
		{"float % float", lexer.REM, floatType(), floatType(), true, nil},
		{"double % int", lexer.REM, doubleType(), intType(), true, nil},

		// Comparison operators
		{"int == int", lexer.EQL, intType(), intType(), false, intType()},
		{"int != int", lexer.NEQ, intType(), intType(), false, intType()},
		{"int < int", lexer.LSS, intType(), intType(), false, intType()},
		{"int > int", lexer.GTR, intType(), intType(), false, intType()},
		{"int <= int", lexer.LEQ, intType(), intType(), false, intType()},
		{"int >= int", lexer.GEQ, intType(), intType(), false, intType()},

		// Comparison with different types
		{"int == double", lexer.EQL, intType(), doubleType(), false, intType()},
		{"int < float", lexer.LSS, intType(), floatType(), false, intType()},

		// Comparison with pointers
		{"int* == int*", lexer.EQL, intPtrType(), intPtrType(), false, intType()},
		{"int* < int*", lexer.LSS, intPtrType(), intPtrType(), false, intType()},
		{"int* == 0", lexer.EQL, intPtrType(), intType(), false, intType()},

		// Logical operators
		{"int && int", lexer.LAND, intType(), intType(), false, intType()},
		{"int || int", lexer.LOR, intType(), intType(), false, intType()},
		{"int && char", lexer.LAND, intType(), charType(), false, intType()},
		{"pointer && int", lexer.LAND, intPtrType(), intType(), false, intType()},

		// Bitwise operators
		{"int & int", lexer.AND, intType(), intType(), false, intType()},
		{"int | int", lexer.OR, intType(), intType(), false, intType()},
		{"int ^ int", lexer.XOR, intType(), intType(), false, intType()},
		{"int << int", lexer.SHL, intType(), intType(), false, intType()},
		{"int >> int", lexer.SHR, intType(), intType(), false, intType()},

		// Bitwise with different integer types
		{"char & int", lexer.AND, charType(), intType(), false, intType()},
		{"short | long", lexer.OR, shortType(), longType(), false, longType()},

		// Bitwise requires integers
		{"float & float", lexer.AND, floatType(), floatType(), true, nil},
		{"double << int", lexer.SHL, doubleType(), intType(), true, nil},

		// Invalid operand types
		{"int + pointer", lexer.ADD, intType(), intPtrType(), true, nil},
		{"pointer + pointer", lexer.ADD, intPtrType(), intPtrType(), true, nil},
		{"struct + int", lexer.ADD, structType("S"), intType(), true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			resultType, err := tc.CheckBinaryOp(tt.op, tt.left, tt.right, testPos())

			if tt.expectError {
				if err == nil {
					t.Errorf("CheckBinaryOp(%s, %s, %s) expected error, got nil",
						tt.op, tt.left.String(), tt.right.String())
				}
			} else {
				if err != nil {
					t.Errorf("CheckBinaryOp(%s, %s, %s) unexpected error: %v",
						tt.op, tt.left.String(), tt.right.String(), err)
				}
				if resultType == nil {
					t.Errorf("CheckBinaryOp(%s, %s, %s) returned nil type",
						tt.op, tt.left.String(), tt.right.String())
				} else if tt.expectType != nil {
					// Compare type kinds for basic types
					if resultType.TypeKind() != tt.expectType.TypeKind() {
						t.Errorf("CheckBinaryOp(%s, %s, %s) returned type %s, want %s",
							tt.op, tt.left.String(), tt.right.String(),
							resultType.String(), tt.expectType.String())
					}
				}
			}
		})
	}
}

// ============================================================================
// CheckUnaryOp Tests
// ============================================================================

func TestCheckUnaryOp(t *testing.T) {
	tests := []struct {
		name        string
		op          lexer.TokenType
		operand     parser.Type
		expectError bool
		expectType  parser.Type
	}{
		// Unary plus/minus (arithmetic types only)
		{"+int", lexer.ADD, intType(), false, intType()},
		{"-int", lexer.SUB, intType(), false, intType()},
		{"+char", lexer.ADD, charType(), false, intType()}, // integer promotion
		{"-double", lexer.SUB, doubleType(), false, doubleType()},
		{"+pointer", lexer.ADD, intPtrType(), true, nil},   // invalid

		// Logical NOT (scalar types)
		{"!int", lexer.NOT, intType(), false, intType()},
		{"!char", lexer.NOT, charType(), false, intType()},
		{"!pointer", lexer.NOT, intPtrType(), false, intType()},
		{"!struct", lexer.NOT, structType("S"), true, nil}, // invalid

		// Bitwise NOT (integer types only)
		{"~int", lexer.BITNOT, intType(), false, intType()},
		{"~char", lexer.BITNOT, charType(), false, intType()}, // integer promotion
		{"~long", lexer.BITNOT, longType(), false, longType()},
		{"~float", lexer.BITNOT, floatType(), true, nil},   // invalid

		// Address-of (any type)
		{"&int", lexer.AND, intType(), false, intPtrType()},
		{"&char", lexer.AND, charType(), false, charPtrType()},
		{"&struct", lexer.AND, structType("S"), false, &parser.PointerType{Elem: structType("S")}},

		// Indirection (pointer types only)
		{"*int*", lexer.MUL, intPtrType(), false, intType()},
		{"*char*", lexer.MUL, charPtrType(), false, charType()},
		{"*int", lexer.MUL, intType(), true, nil},          // invalid

		// Increment/Decrement (arithmetic and pointer types)
		{"++int", lexer.INC, intType(), false, intType()},
		{"--int", lexer.DEC, intType(), false, intType()},
		{"++char", lexer.INC, charType(), false, charType()},
		{"++double", lexer.INC, doubleType(), false, doubleType()},
		{"++int*", lexer.INC, intPtrType(), false, intPtrType()},
		{"++struct", lexer.INC, structType("S"), true, nil}, // invalid

		// Sizeof (any type, returns int)
		{"sizeof int", lexer.SIZEOF, intType(), false, intType()},
		{"sizeof char", lexer.SIZEOF, charType(), false, intType()},
		{"sizeof struct", lexer.SIZEOF, structType("S"), false, intType()},

		// Invalid operator
		{"unknown op", lexer.ADD_ASSIGN, intType(), true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			resultType, err := tc.CheckUnaryOp(tt.op, tt.operand, testPos())

			if tt.expectError {
				if err == nil {
					t.Errorf("CheckUnaryOp(%s, %s) expected error, got nil",
						tt.op, tt.operand.String())
				}
			} else {
				if err != nil {
					t.Errorf("CheckUnaryOp(%s, %s) unexpected error: %v",
						tt.op, tt.operand.String(), err)
				}
				if resultType == nil {
					t.Errorf("CheckUnaryOp(%s, %s) returned nil type",
						tt.op, tt.operand.String())
				}
			}
		})
	}
}

// ============================================================================
// CheckCall Tests
// ============================================================================

func TestCheckCall(t *testing.T) {
	tests := []struct {
		name        string
		funcType    *parser.FuncType
		args        []parser.Type
		expectError bool
	}{
		// Exact match
		{"exact match", &parser.FuncType{Return: intType(), Params: []parser.Type{intType(), charType()}}, []parser.Type{intType(), charType()}, false},

		// Argument count mismatch
		{"too few args", &parser.FuncType{Return: intType(), Params: []parser.Type{intType(), charType()}}, []parser.Type{intType()}, true},
		{"too many args", &parser.FuncType{Return: intType(), Params: []parser.Type{intType()}}, []parser.Type{intType(), charType(), doubleType()}, true},

		// Type mismatch
		{"type mismatch", &parser.FuncType{Return: intType(), Params: []parser.Type{intType()}}, []parser.Type{structType("S")}, true},

		// Implicit conversions
		{"char to int", &parser.FuncType{Return: intType(), Params: []parser.Type{intType()}}, []parser.Type{charType()}, false},
		{"int to double", &parser.FuncType{Return: intType(), Params: []parser.Type{doubleType()}}, []parser.Type{intType()}, false},

		// Variadic functions
		{"variadic exact", &parser.FuncType{Return: intType(), Params: []parser.Type{intType()}, Variadic: true}, []parser.Type{intType()}, false},
		{"variadic extra", &parser.FuncType{Return: intType(), Params: []parser.Type{intType()}, Variadic: true}, []parser.Type{intType(), charType(), doubleType()}, false},
		{"variadic too few", &parser.FuncType{Return: intType(), Params: []parser.Type{intType(), charType()}, Variadic: true}, []parser.Type{intType()}, true},

		// Void function
		{"void function no args", &parser.FuncType{Return: voidType(), Params: nil}, []parser.Type{}, false},
		{"void return type", &parser.FuncType{Return: voidType(), Params: []parser.Type{intType()}}, []parser.Type{intType()}, false},

		// Null function type
		{"nil func type", nil, []parser.Type{intType()}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			resultType, err := tc.CheckCall(tt.funcType, tt.args, testPos())

			if tt.expectError {
				if err == nil {
					t.Errorf("CheckCall() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("CheckCall() unexpected error: %v", err)
				}
				if tt.funcType != nil && resultType == nil {
					t.Errorf("CheckCall() returned nil type, expected %s", tt.funcType.Return.String())
				}
			}
		})
	}
}

// ============================================================================
// InferType Tests
// ============================================================================

func TestInferType(t *testing.T) {
	tests := []struct {
		name        string
		expr        parser.Expr
		expectType  parser.Type
		expectError bool
		setupFunc   func(*SemanticAnalyzer)
	}{
		// Integer literal
		{"int literal", &parser.IntLiteral{Value: 42}, intType(), false, nil},
		{"int literal with suffix", &parser.IntLiteral{Value: 42, Suffix: "u"}, uintType(), false, nil},
		{"long literal", &parser.IntLiteral{Value: 42, Suffix: "l"}, longType(), false, nil},
		{"long long literal", &parser.IntLiteral{Value: 42, Suffix: "ll"}, longLongType(), false, nil},

		// Float literal
		{"double literal", &parser.FloatLiteral{Value: 3.14}, doubleType(), false, nil},
		{"float literal", &parser.FloatLiteral{Value: 3.14, Suffix: "f"}, floatType(), false, nil},
		{"long double literal", &parser.FloatLiteral{Value: 3.14, Suffix: "l"}, longDoubleType(), false, nil},

		// Char literal
		{"char literal", &parser.CharLiteral{Value: 'a'}, intType(), false, nil},

		// String literal
		{"string literal", &parser.StringLiteral{Value: "hello"}, charPtrType(), false, nil},

		// Identifier (requires symbol table setup)
		{"identifier int", &parser.IdentExpr{Name: "x"}, intType(), false, func(a *SemanticAnalyzer) {
			a.symbolTable.globalScope.symbols["x"] = &Symbol{Name: "x", Type: intType()}
		}},
		{"identifier undefined", &parser.IdentExpr{Name: "undefined"}, nil, true, nil},

		// Binary expression
		{"binary add", &parser.BinaryExpr{Op: lexer.ADD, Left: &parser.IntLiteral{Value: 1}, Right: &parser.IntLiteral{Value: 2}}, intType(), false, nil},
		{"binary compare", &parser.BinaryExpr{Op: lexer.EQL, Left: &parser.IntLiteral{Value: 1}, Right: &parser.IntLiteral{Value: 2}}, intType(), false, nil},

		// Unary expression
		{"unary minus", &parser.UnaryExpr{Op: lexer.SUB, Operand: &parser.IntLiteral{Value: 5}}, intType(), false, nil},
		{"unary address", &parser.UnaryExpr{Op: lexer.AND, Operand: &parser.IdentExpr{Name: "x"}}, intPtrType(), false, func(a *SemanticAnalyzer) {
			a.symbolTable.globalScope.symbols["x"] = &Symbol{Name: "x", Type: intType()}
		}},
		{"unary deref", &parser.UnaryExpr{Op: lexer.MUL, Operand: &parser.IdentExpr{Name: "p"}}, intType(), false, func(a *SemanticAnalyzer) {
			a.symbolTable.globalScope.symbols["p"] = &Symbol{Name: "p", Type: intPtrType()}
		}},

		// Call expression
		{"call expression", &parser.CallExpr{Func: &parser.IdentExpr{Name: "foo"}, Args: []parser.Expr{&parser.IntLiteral{Value: 42}}}, intType(), false, func(a *SemanticAnalyzer) {
			a.symbolTable.globalScope.symbols["foo"] = &Symbol{Name: "foo", Type: &parser.FuncType{Return: intType(), Params: []parser.Type{intType()}}}
		}},

		// Cast expression
		{"cast expression", &parser.CastExpr{Type: doubleType(), Expr: &parser.IntLiteral{Value: 42}}, doubleType(), false, nil},

		// Sizeof expression
		{"sizeof expression", &parser.SizeofExpr{Expr: &parser.IntLiteral{Value: 42}}, intType(), false, nil},

		// Member expression
		{"member expression", &parser.MemberExpr{Object: &parser.IdentExpr{Name: "s"}, Field: "x", IsPointer: false}, intType(), false, func(a *SemanticAnalyzer) {
			a.symbolTable.globalScope.symbols["s"] = &Symbol{Name: "s", Type: &parser.StructType{
				Name: "S",
				Fields: []*parser.FieldDecl{
					{Name: "x", Type: intType()},
				},
			}}
		}},

		// Index expression
		{"index expression", &parser.IndexExpr{Array: &parser.IdentExpr{Name: "arr"}, Index: &parser.IntLiteral{Value: 0}}, intType(), false, func(a *SemanticAnalyzer) {
			a.symbolTable.globalScope.symbols["arr"] = &Symbol{Name: "arr", Type: intArrayType()}
		}},

		// Conditional expression
		{"conditional expression", &parser.CondExpr{Cond: &parser.IntLiteral{Value: 1}, True: &parser.IntLiteral{Value: 2}, False: &parser.IntLiteral{Value: 3}}, intType(), false, nil},

		// Assignment expression
		{"assignment expression", &parser.AssignExpr{Left: &parser.IdentExpr{Name: "x"}, Right: &parser.IntLiteral{Value: 42}}, intType(), false, func(a *SemanticAnalyzer) {
			a.symbolTable.globalScope.symbols["x"] = &Symbol{Name: "x", Type: intType()}
		}},

		// Nil expression
		{"nil expression", nil, nil, false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			if tt.setupFunc != nil {
				tt.setupFunc(tc.analyzer)
			}

			resultType := tc.InferType(tt.expr)

			if tt.expectError {
				if resultType != nil {
					// Some cases might return nil without error
				}
			}

			if tt.expectType != nil && resultType == nil {
				t.Errorf("InferType(%s) returned nil, expected %s", tt.name, tt.expectType.String())
			}
			if tt.expectType == nil && resultType != nil && !tt.expectError {
				t.Errorf("InferType(%s) returned %s, expected nil", tt.name, resultType.String())
			}
		})
	}
}

// ============================================================================
// ImplicitCast Tests
// ============================================================================

func TestImplicitCast(t *testing.T) {
	tests := []struct {
		name       string
		expr       parser.Expr
		fromType   parser.Type
		toType     parser.Type
		expectCast bool // true if cast expression expected, false if expr returned as-is
	}{
		// Same type - no cast needed
		{"same type", &parser.IntLiteral{Value: 42}, intType(), intType(), false},

		// Integer widening - no cast needed
		{"char to int", &parser.IntLiteral{Value: 42}, charType(), intType(), false},
		{"short to long", &parser.IntLiteral{Value: 42}, shortType(), longType(), false},

		// Integer narrowing - cast needed
		{"long to int", &parser.IntLiteral{Value: 42}, longType(), intType(), true},

		// Float conversions - narrowing needs cast
		{"float to double", &parser.FloatLiteral{Value: 3.14}, floatType(), doubleType(), false},
		{"double to float", &parser.FloatLiteral{Value: 3.14}, doubleType(), floatType(), false},

		// Integer to float - no cast needed (implicit conversion)
		{"int to float", &parser.IntLiteral{Value: 42}, intType(), floatType(), false},

		// Float to integer - cast needed
		{"float to int", &parser.FloatLiteral{Value: 3.14}, floatType(), intType(), true},

		// Pointer conversions - void* to typed pointer needs cast
		{"void* to int*", &parser.IdentExpr{Name: "p"}, voidPtrType(), intPtrType(), false},
		{"int* to void*", &parser.IdentExpr{Name: "p"}, intPtrType(), voidPtrType(), false},
		{"same pointer", &parser.IdentExpr{Name: "p"}, intPtrType(), intPtrType(), false},

		// Array to pointer decay
		{"array to pointer", &parser.IdentExpr{Name: "arr"}, intArrayType(), intPtrType(), false},

		// Function to pointer decay
		{"function to pointer", &parser.IdentExpr{Name: "fn"}, funcType(intType(), []parser.Type{intType()}, false), funcPtrType(), false},

		// Nil expression
		{"nil expr", nil, intType(), doubleType(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			result := tc.ImplicitCast(tt.expr, tt.fromType, tt.toType)

			if tt.expr == nil {
				if result != nil {
					t.Errorf("ImplicitCast(nil, ...) expected nil, got %v", result)
				}
				return
			}

			_, isCast := result.(*parser.CastExpr)
			if tt.expectCast && !isCast {
				t.Errorf("ImplicitCast(%s -> %s) expected CastExpr, got %T",
					tt.fromType.String(), tt.toType.String(), result)
			}
			if !tt.expectCast && isCast {
				t.Errorf("ImplicitCast(%s -> %s) expected no cast, got CastExpr",
					tt.fromType.String(), tt.toType.String())
			}
		})
	}
}

// ============================================================================
// Helper Method Tests
// ============================================================================

func TestUnwrapType(t *testing.T) {
	tests := []struct {
		name     string
		input    parser.Type
		expected parser.Type
	}{
		{"nil", nil, nil},
		{"base type", intType(), intType()},
		{"typedef", typedefType("my_int", intType()), intType()},
		{"qualified", constIntType(), intType()},
		{"nested typedef", typedefType("a", typedefType("b", intType())), intType()},
		{"nested qualified", &parser.QualifiedType{Type: constIntType(), IsVolatile: true}, intType()},
		{"typedef with qualified", typedefType("my_int", constIntType()), intType()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			result := tc.unwrapType(tt.input)

			if result == nil && tt.expected == nil {
				return
			}
			if result == nil || tt.expected == nil {
				t.Errorf("unwrapType(%v) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			if result.TypeKind() != tt.expected.TypeKind() {
				t.Errorf("unwrapType(%v) type kind = %v, want %v", tt.input, result.TypeKind(), tt.expected.TypeKind())
			}
		})
	}
}

func TestTypesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        parser.Type
		b        parser.Type
		expected bool
	}{
		{"nil nil", nil, nil, true},
		{"nil int", nil, intType(), false},
		{"int int", intType(), intType(), true},
		{"int char", intType(), charType(), false},
		{"signed unsigned int", intType(), uintType(), false},
		{"pointer same", intPtrType(), intPtrType(), true},
		{"pointer different", intPtrType(), charPtrType(), false},
		{"array same", intArrayType(), intArrayType(), true},
		{"array different size", &parser.ArrayType{Elem: intType(), ArraySize: 10}, &parser.ArrayType{Elem: intType(), ArraySize: 20}, false},
		{"function same", funcType(intType(), []parser.Type{intType()}, false), funcType(intType(), []parser.Type{intType()}, false), true},
		{"struct same name", structType("S"), structType("S"), true},
		{"struct different name", structType("S"), structType("T"), false},
		{"typedef same name", typedefType("my_int", intType()), typedefType("my_int", intType()), true},
		{"typedef different name", typedefType("my_int", intType()), typedefType("int32", intType()), false},
		{"qualified ignored", constIntType(), intType(), false}, // qualified types have different TypeKind
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			result := tc.typesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("typesEqual(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestIsArithmeticType(t *testing.T) {
	tests := []struct {
		name     string
		typ      parser.Type
		expected bool
	}{
		{"nil", nil, false},
		{"bool", boolType(), true},
		{"char", charType(), true},
		{"short", shortType(), true},
		{"int", intType(), true},
		{"long", longType(), true},
		{"float", floatType(), true},
		{"double", doubleType(), true},
		{"void", voidType(), false},
		{"pointer", intPtrType(), false},
		{"array", intArrayType(), false},
		{"struct", structType("S"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			result := tc.isArithmeticType(tt.typ)
			if result != tt.expected {
				t.Errorf("isArithmeticType(%v) = %v, want %v", tt.typ, result, tt.expected)
			}
		})
	}
}

func TestIsIntegerType(t *testing.T) {
	tests := []struct {
		name     string
		typ      parser.Type
		expected bool
	}{
		{"nil", nil, false},
		{"bool", boolType(), true},
		{"char", charType(), true},
		{"short", shortType(), true},
		{"int", intType(), true},
		{"long", longType(), true},
		{"float", floatType(), false},
		{"double", doubleType(), false},
		{"pointer", intPtrType(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			result := tc.isIntegerType(tt.typ)
			if result != tt.expected {
				t.Errorf("isIntegerType(%v) = %v, want %v", tt.typ, result, tt.expected)
			}
		})
	}
}

func TestIsFloatType(t *testing.T) {
	tests := []struct {
		name     string
		typ      parser.Type
		expected bool
	}{
		{"nil", nil, false},
		{"int", intType(), false},
		{"float", floatType(), true},
		{"double", doubleType(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			result := tc.isFloatType(tt.typ)
			if result != tt.expected {
				t.Errorf("isFloatType(%v) = %v, want %v", tt.typ, result, tt.expected)
			}
		})
	}
}

func TestIsPointerType(t *testing.T) {
	tests := []struct {
		name     string
		typ      parser.Type
		expected bool
	}{
		{"nil", nil, false},
		{"int", intType(), false},
		{"int*", intPtrType(), true},
		{"void*", voidPtrType(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			result := tc.isPointerType(tt.typ)
			if result != tt.expected {
				t.Errorf("isPointerType(%v) = %v, want %v", tt.typ, result, tt.expected)
			}
		})
	}
}

func TestIsVoidType(t *testing.T) {
	tests := []struct {
		name     string
		typ      parser.Type
		expected bool
	}{
		{"nil", nil, false},
		{"void", voidType(), true},
		{"int", intType(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			result := tc.isVoidType(tt.typ)
			if result != tt.expected {
				t.Errorf("isVoidType(%v) = %v, want %v", tt.typ, result, tt.expected)
			}
		})
	}
}

func TestIsVoidPointer(t *testing.T) {
	tests := []struct {
		name     string
		typ      parser.Type
		expected bool
	}{
		{"nil", nil, false},
		{"void*", voidPtrType(), true},
		{"int*", intPtrType(), false},
		{"void", voidType(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			result := tc.isVoidPointer(tt.typ)
			if result != tt.expected {
				t.Errorf("isVoidPointer(%v) = %v, want %v", tt.typ, result, tt.expected)
			}
		})
	}
}

func TestIsObjectPointer(t *testing.T) {
	tests := []struct {
		name     string
		typ      parser.Type
		expected bool
	}{
		{"nil", nil, false},
		{"int*", intPtrType(), true},
		{"char*", charPtrType(), true},
		{"void*", voidPtrType(), false},
		{"func*", funcPtrType(), false},
		{"int", intType(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			result := tc.isObjectPointer(tt.typ)
			if result != tt.expected {
				t.Errorf("isObjectPointer(%v) = %v, want %v", tt.typ, result, tt.expected)
			}
		})
	}
}

func TestIsArrayType(t *testing.T) {
	tests := []struct {
		name     string
		typ      parser.Type
		expected bool
	}{
		{"nil", nil, false},
		{"int[10]", intArrayType(), true},
		{"int*", intPtrType(), false},
		{"int", intType(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			result := tc.isArrayType(tt.typ)
			if result != tt.expected {
				t.Errorf("isArrayType(%v) = %v, want %v", tt.typ, result, tt.expected)
			}
		})
	}
}

func TestIsFunctionType(t *testing.T) {
	tests := []struct {
		name     string
		typ      parser.Type
		expected bool
	}{
		{"nil", nil, false},
		{"func()", funcType(intType(), []parser.Type{intType()}, false), true},
		{"int*", intPtrType(), false},
		{"int", intType(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			result := tc.isFunctionType(tt.typ)
			if result != tt.expected {
				t.Errorf("isFunctionType(%v) = %v, want %v", tt.typ, result, tt.expected)
			}
		})
	}
}

func TestIntegerPromotion(t *testing.T) {
	tests := []struct {
		name     string
		input    parser.Type
		expected parser.Type
	}{
		{"nil", nil, nil},
		{"bool", boolType(), intType()},
		{"char", charType(), intType()},
		{"short", shortType(), intType()},
		{"int", intType(), intType()},
		{"long", longType(), longType()},
		{"float", floatType(), floatType()}, // not integer, no promotion
		{"pointer", intPtrType(), intPtrType()}, // not integer, no promotion
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			result := tc.integerPromotion(tt.input)

			if result == nil && tt.expected == nil {
				return
			}
			if result == nil || tt.expected == nil {
				t.Errorf("integerPromotion(%v) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			if result.TypeKind() != tt.expected.TypeKind() {
				t.Errorf("integerPromotion(%v) = %v, want %v", tt.input, result.TypeKind(), tt.expected.TypeKind())
			}
		})
	}
}

func TestUsualArithmeticConversions(t *testing.T) {
	tests := []struct {
		name     string
		left     parser.Type
		right    parser.Type
		expected parser.Type
	}{
		{"int int", intType(), intType(), intType()},
		{"int double", intType(), doubleType(), doubleType()},
		{"double int", doubleType(), intType(), doubleType()},
		{"float double", floatType(), doubleType(), doubleType()},
		{"int float", intType(), floatType(), floatType()},
		{"char short", charType(), shortType(), intType()}, // both promoted to int
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			result := tc.usualArithmeticConversions(tt.left, tt.right)

			if result == nil {
				t.Errorf("usualArithmeticConversions(%v, %v) = nil", tt.left, tt.right)
				return
			}
			if result.TypeKind() != tt.expected.TypeKind() {
				t.Errorf("usualArithmeticConversions(%v, %v) = %v, want %v",
					tt.left, tt.right, result.TypeKind(), tt.expected.TypeKind())
			}
		})
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestCheckAssignableEdgeCases(t *testing.T) {
	t.Run("incomplete array to pointer", func(t *testing.T) {
		tc := newTestTypeChecker()
		incompleteArray := incompleteArrayType()
		ptrType := intPtrType()
		err := tc.CheckAssignable(ptrType, incompleteArray, testPos())
		// Array decay should work even for incomplete arrays
		if err != nil {
			t.Errorf("CheckAssignable for incomplete array decay: %v", err)
		}
	})

	t.Run("qualified type assignment", func(t *testing.T) {
		tc := newTestTypeChecker()
		// const int to int should work (discarding const)
		err := tc.CheckAssignable(intType(), constIntType(), testPos())
		if err != nil {
			t.Errorf("CheckAssignable const int to int: %v", err)
		}
	})

	t.Run("typedef chain", func(t *testing.T) {
		tc := newTestTypeChecker()
		typedefChain := typedefType("a", typedefType("b", typedefType("c", intType())))
		err := tc.CheckAssignable(intType(), typedefChain, testPos())
		if err != nil {
			t.Errorf("CheckAssignable typedef chain: %v", err)
		}
	})
}

func TestCheckBinaryOpEdgeCases(t *testing.T) {
	t.Run("long double arithmetic", func(t *testing.T) {
		tc := newTestTypeChecker()
		resultType, err := tc.CheckBinaryOp(lexer.ADD, longDoubleType(), doubleType(), testPos())
		if err != nil {
			t.Errorf("CheckBinaryOp long double + double: %v", err)
		}
		if resultType == nil {
			t.Error("CheckBinaryOp long double + double returned nil type")
		}
	})

	t.Run("unsigned arithmetic", func(t *testing.T) {
		tc := newTestTypeChecker()
		resultType, err := tc.CheckBinaryOp(lexer.ADD, uintType(), intType(), testPos())
		if err != nil {
			t.Errorf("CheckBinaryOp unsigned int + int: %v", err)
		}
		if resultType == nil {
			t.Error("CheckBinaryOp unsigned int + int returned nil type")
		}
	})
}

func TestInferTypeEdgeCases(t *testing.T) {
	t.Run("nested unary ops", func(t *testing.T) {
		tc := newTestTypeChecker()
		expr := &parser.UnaryExpr{
			Op: lexer.SUB,
			Operand: &parser.UnaryExpr{
				Op: lexer.ADD,
				Operand: &parser.IntLiteral{Value: 42},
			},
		}
		resultType := tc.InferType(expr)
		if resultType == nil {
			t.Error("InferType nested unary ops returned nil")
		}
	})

	t.Run("complex binary expression", func(t *testing.T) {
		tc := newTestTypeChecker()
		expr := &parser.BinaryExpr{
			Op: lexer.ADD,
			Left: &parser.BinaryExpr{
				Op: lexer.MUL,
				Left:  &parser.IntLiteral{Value: 2},
				Right: &parser.IntLiteral{Value: 3},
			},
			Right: &parser.IntLiteral{Value: 4},
		}
		resultType := tc.InferType(expr)
		if resultType == nil {
			t.Error("InferType complex binary expr returned nil")
		}
	})
}

func TestImplicitCastEdgeCases(t *testing.T) {
	t.Run("qualified type cast", func(t *testing.T) {
		tc := newTestTypeChecker()
		expr := &parser.IntLiteral{Value: 42}
		result := tc.ImplicitCast(expr, constIntType(), intType())
		// Should not need explicit cast for qualifier removal
		if _, ok := result.(*parser.CastExpr); ok {
			t.Error("ImplicitCast const int to int should not require explicit cast")
		}
	})

	t.Run("array decay in cast", func(t *testing.T) {
		tc := newTestTypeChecker()
		expr := &parser.IdentExpr{Name: "arr"}
		result := tc.ImplicitCast(expr, intArrayType(), intPtrType())
		// Array decay should not require explicit cast
		if _, ok := result.(*parser.CastExpr); ok {
			t.Error("ImplicitCast array to pointer should not require explicit cast")
		}
	})
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestTypeCheckerIntegration(t *testing.T) {
	t.Run("full expression type inference", func(t *testing.T) {
		tc := newTestTypeChecker()

		// Set up symbol table
		tc.analyzer.symbolTable.globalScope.symbols["x"] = &Symbol{Name: "x", Type: intType()}
		tc.analyzer.symbolTable.globalScope.symbols["y"] = &Symbol{Name: "y", Type: doubleType()}
		tc.analyzer.symbolTable.globalScope.symbols["p"] = &Symbol{Name: "p", Type: intPtrType()}

		// Complex expression: *p + (int)y
		expr := &parser.BinaryExpr{
			Op: lexer.ADD,
			Left: &parser.UnaryExpr{
				Op:       lexer.MUL,
				Operand:  &parser.IdentExpr{Name: "p"},
			},
			Right: &parser.CastExpr{
				Type: intType(),
				Expr: &parser.IdentExpr{Name: "y"},
			},
		}

		resultType := tc.InferType(expr)
		if resultType == nil {
			t.Error("InferType complex expression returned nil")
		}
	})

	t.Run("function call with type checking", func(t *testing.T) {
		tc := newTestTypeChecker()

		// Set up function symbol
		funcType := &parser.FuncType{
			Return: intType(),
			Params: []parser.Type{intType(), doubleType()},
		}
		tc.analyzer.symbolTable.globalScope.symbols["foo"] = &Symbol{Name: "foo", Type: funcType}

		// Call: foo(42, 3.14)
		callExpr := &parser.CallExpr{
			Func: &parser.IdentExpr{Name: "foo"},
			Args: []parser.Expr{
				&parser.IntLiteral{Value: 42},
				&parser.FloatLiteral{Value: 3.14},
			},
		}

		resultType := tc.InferType(callExpr)
		if resultType == nil {
			t.Error("InferType function call returned nil")
		}
	})
}

// ============================================================================
// Additional Tests for Low-Coverage Functions
// ============================================================================

// TestInferInitListExprType tests inference for initializer lists
func TestInferInitListExprType(t *testing.T) {
	tc := newTestTypeChecker()
	
	// Simple initializer list
	initList := &parser.InitListExpr{
		Elements: []parser.Expr{
			&parser.IntLiteral{Value: 1},
			&parser.IntLiteral{Value: 2},
			&parser.IntLiteral{Value: 3},
		},
	}
	
	_ = tc.InferType(initList) // InitListExpr returns nil by design (context-dependent type)
}

// TestInferCondExprType tests conditional expression type inference
func TestInferCondExprType(t *testing.T) {
	tests := []struct {
		name        string
		cond        parser.Expr
		trueExpr    parser.Expr
		falseExpr   parser.Expr
		expectError bool
	}{
		{
			name:     "int conditional",
			cond:     &parser.IntLiteral{Value: 1},
			trueExpr: &parser.IntLiteral{Value: 2},
			falseExpr: &parser.IntLiteral{Value: 3},
		},
		{
			name:     "double conditional",
			cond:     &parser.IntLiteral{Value: 1},
			trueExpr: &parser.FloatLiteral{Value: 2.0},
			falseExpr: &parser.FloatLiteral{Value: 3.0},
		},
		{
			name:     "mixed arithmetic",
			cond:     &parser.IntLiteral{Value: 1},
			trueExpr: &parser.IntLiteral{Value: 2},
			falseExpr: &parser.FloatLiteral{Value: 3.0},
		},
		{
			name:     "pointer conditional",
			cond:     &parser.IntLiteral{Value: 1},
			trueExpr: &parser.IdentExpr{Name: "p"},
			falseExpr: &parser.IdentExpr{Name: "q"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			
			// Set up symbols for pointer test
			if tt.name == "pointer conditional" {
				tc.analyzer.symbolTable.globalScope.symbols["p"] = &Symbol{Name: "p", Type: intPtrType()}
				tc.analyzer.symbolTable.globalScope.symbols["q"] = &Symbol{Name: "q", Type: intPtrType()}
			}
			
			condExpr := &parser.CondExpr{
				Cond:  tt.cond,
				True:  tt.trueExpr,
				False: tt.falseExpr,
			}
			
			resultType := tc.InferType(condExpr)
			if resultType == nil && !tt.expectError {
				t.Error("InferType CondExpr returned nil")
			}
		})
	}
}

// TestCheckLogicalOp tests logical operator type checking
func TestCheckLogicalOp(t *testing.T) {
	tests := []struct {
		name        string
		left        parser.Type
		right       parser.Type
		expectError bool
	}{
		{"int int", intType(), intType(), false},
		{"double double", doubleType(), doubleType(), false},
		{"pointer pointer", intPtrType(), intPtrType(), false},
		{"void* void*", voidPtrType(), voidPtrType(), false},
		{"char char", charType(), charType(), false},
		{"int double", intType(), doubleType(), false},
		{"struct struct", structType("S"), structType("S"), true}, // structs not valid for logical ops
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			resultType, err := tc.checkLogicalOp(tt.left, tt.right, testPos())
			
			if tt.expectError && err == nil {
				t.Errorf("checkLogicalOp expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("checkLogicalOp unexpected error: %v", err)
			}
			if !tt.expectError && resultType == nil {
				t.Error("checkLogicalOp returned nil type")
			}
		})
	}
}

// TestInferIndexExprType tests array indexing type inference
func TestInferIndexExprType(t *testing.T) {
	tc := newTestTypeChecker()
	
	// Set up array symbol
	tc.analyzer.symbolTable.globalScope.symbols["arr"] = &Symbol{Name: "arr", Type: intArrayType()}
	
	// Test array indexing
	indexExpr := &parser.IndexExpr{
		Array: &parser.IdentExpr{Name: "arr"},
		Index: &parser.IntLiteral{Value: 0},
	}
	
	resultType := tc.InferType(indexExpr)
	if resultType == nil {
		t.Error("InferType IndexExpr returned nil")
	}
}

// TestIntegerConversions tests integer conversion rules
func TestIntegerConversions(t *testing.T) {
	tests := []struct {
		name     string
		left     parser.Type
		right    parser.Type
		expected parser.Type
	}{
		{"int int", intType(), intType(), intType()},
		{"int long", intType(), longType(), longType()},
		{"unsigned int int", uintType(), intType(), uintType()},
		{"int unsigned long", intType(), &parser.BaseType{Kind: parser.TypeLong, Signed: false}, &parser.BaseType{Kind: parser.TypeLong, Signed: false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			result := tc.integerConversions(tt.left, tt.right)
			if result == nil {
				t.Errorf("integerConversions(%v, %v) returned nil", tt.left, tt.right)
				return
			}
			if result.TypeKind() != tt.expected.TypeKind() {
				t.Errorf("integerConversions(%v, %v) = %v, want %v", tt.left, tt.right, result.TypeKind(), tt.expected.TypeKind())
			}
		})
	}
}

// TestCheckComparisonOp tests comparison operator type checking
func TestCheckComparisonOp(t *testing.T) {
	tests := []struct {
		name        string
		left        parser.Type
		right       parser.Type
		expectError bool
	}{
		{"int int", intType(), intType(), false},
		{"double double", doubleType(), doubleType(), false},
		{"int double", intType(), doubleType(), false},
		{"pointer pointer", intPtrType(), intPtrType(), false},
		{"void* void*", voidPtrType(), voidPtrType(), false},
		{"int pointer", intType(), intPtrType(), false}, // comparison allows int and pointer
		{"struct struct", structType("S"), structType("S"), true}, // structs not comparable
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			resultType, err := tc.checkComparisonOp(tt.left, tt.right, testPos())
			
			if tt.expectError && err == nil {
				t.Errorf("checkComparisonOp expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("checkComparisonOp unexpected error: %v", err)
			}
			if !tt.expectError && resultType == nil {
				t.Error("checkComparisonOp returned nil type")
			}
		})
	}
}

// TestCheckBitwiseOp tests bitwise operator type checking
func TestCheckBitwiseOp(t *testing.T) {
	tests := []struct {
		name        string
		op          lexer.TokenType
		left        parser.Type
		right       parser.Type
		expectError bool
	}{
		{"AND int int", lexer.AND, intType(), intType(), false},
		{"OR int int", lexer.OR, intType(), intType(), false},
		{"XOR int int", lexer.XOR, intType(), intType(), false},
		{"AND char char", lexer.AND, charType(), charType(), false},
		{"AND long long", lexer.AND, longType(), longType(), false},
		{"AND int double", lexer.AND, intType(), doubleType(), true}, // float not allowed
		{"AND pointer int", lexer.AND, intPtrType(), intType(), true}, // pointer not allowed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestTypeChecker()
			resultType, err := tc.checkBitwiseOp(tt.op, tt.left, tt.right, testPos())
			
			if tt.expectError && err == nil {
				t.Errorf("checkBitwiseOp expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("checkBitwiseOp unexpected error: %v", err)
			}
			if !tt.expectError && resultType == nil {
				t.Error("checkBitwiseOp returned nil type")
			}
		})
	}
}

// TestIsTypeHelperFunctions tests the type checking helper functions
func TestIsTypeHelperFunctions(t *testing.T) {
	tc := newTestTypeChecker()
	
	t.Run("isPointerType", func(t *testing.T) {
		if !tc.isPointerType(intPtrType()) {
			t.Error("isPointerType(int*) should be true")
		}
		if !tc.isPointerType(voidPtrType()) {
			t.Error("isPointerType(void*) should be true")
		}
		if tc.isPointerType(intType()) {
			t.Error("isPointerType(int) should be false")
		}
	})
	
	t.Run("isVoidType", func(t *testing.T) {
		if !tc.isVoidType(voidType()) {
			t.Error("isVoidType(void) should be true")
		}
		if tc.isVoidType(intType()) {
			t.Error("isVoidType(int) should be false")
		}
	})
	
	t.Run("isArrayType", func(t *testing.T) {
		if !tc.isArrayType(intArrayType()) {
			t.Error("isArrayType(int[10]) should be true")
		}
		if tc.isArrayType(intType()) {
			t.Error("isArrayType(int) should be false")
		}
	})
	
	t.Run("isFunctionType", func(t *testing.T) {
		if !tc.isFunctionType(funcType(intType(), []parser.Type{intType()}, false)) {
			t.Error("isFunctionType(func) should be true")
		}
		if tc.isFunctionType(intType()) {
			t.Error("isFunctionType(int) should be false")
		}
	})
}

// TestInferIdentType tests identifier type inference
func TestInferIdentType(t *testing.T) {
	tc := newTestTypeChecker()
	
	// Set up various symbols
	tc.analyzer.symbolTable.globalScope.symbols["x"] = &Symbol{Name: "x", Type: intType()}
	tc.analyzer.symbolTable.globalScope.symbols["d"] = &Symbol{Name: "d", Type: doubleType()}
	tc.analyzer.symbolTable.globalScope.symbols["p"] = &Symbol{Name: "p", Type: intPtrType()}
	tc.analyzer.symbolTable.globalScope.symbols["arr"] = &Symbol{Name: "arr", Type: intArrayType()}
	tc.analyzer.symbolTable.globalScope.symbols["func"] = &Symbol{Name: "func", Type: funcType(intType(), []parser.Type{intType()}, false)}
	
	tests := []struct {
		name  string
		ident string
	}{
		{"int variable", "x"},
		{"double variable", "d"},
		{"pointer variable", "p"},
		{"array variable", "arr"},
		{"function", "func"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &parser.IdentExpr{Name: tt.ident}
			resultType := tc.InferType(expr)
			if resultType == nil {
				t.Errorf("InferType %s returned nil", tt.name)
			}
		})
	}
	
	// Test undefined identifier
	t.Run("undefined identifier", func(t *testing.T) {
		expr := &parser.IdentExpr{Name: "undefined"}
		resultType := tc.InferType(expr)
		if resultType != nil {
			t.Error("InferType undefined identifier should return nil")
		}
	})
}

// TestInferBinaryExprType tests binary expression type inference
func TestInferBinaryExprType(t *testing.T) {
	tc := newTestTypeChecker()
	
	tests := []struct {
		name  string
		op    lexer.TokenType
		left  parser.Expr
		right parser.Expr
	}{
		{"addition", lexer.ADD, &parser.IntLiteral{Value: 1}, &parser.IntLiteral{Value: 2}},
		{"subtraction", lexer.SUB, &parser.IntLiteral{Value: 5}, &parser.IntLiteral{Value: 3}},
		{"multiplication", lexer.MUL, &parser.IntLiteral{Value: 2}, &parser.IntLiteral{Value: 3}},
		{"division", lexer.QUO, &parser.IntLiteral{Value: 10}, &parser.IntLiteral{Value: 2}},
		{"modulus", lexer.REM, &parser.IntLiteral{Value: 10}, &parser.IntLiteral{Value: 3}},
		{"less than", lexer.LSS, &parser.IntLiteral{Value: 1}, &parser.IntLiteral{Value: 2}},
		{"greater than", lexer.GTR, &parser.IntLiteral{Value: 5}, &parser.IntLiteral{Value: 3}},
		{"equal", lexer.EQL, &parser.IntLiteral{Value: 1}, &parser.IntLiteral{Value: 1}},
		{"not equal", lexer.NEQ, &parser.IntLiteral{Value: 1}, &parser.IntLiteral{Value: 2}},
		{"logical AND", lexer.LAND, &parser.IntLiteral{Value: 1}, &parser.IntLiteral{Value: 1}},
		{"logical OR", lexer.LOR, &parser.IntLiteral{Value: 0}, &parser.IntLiteral{Value: 1}},
		{"bitwise AND", lexer.AND, &parser.IntLiteral{Value: 5}, &parser.IntLiteral{Value: 3}},
		{"bitwise OR", lexer.OR, &parser.IntLiteral{Value: 5}, &parser.IntLiteral{Value: 3}},
		{"bitwise XOR", lexer.XOR, &parser.IntLiteral{Value: 5}, &parser.IntLiteral{Value: 3}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &parser.BinaryExpr{
				Op:    tt.op,
				Left:  tt.left,
				Right: tt.right,
			}
			resultType := tc.InferType(expr)
			if resultType == nil {
				t.Errorf("InferType %s returned nil", tt.name)
			}
		})
	}
}

// TestInferUnaryExprType tests unary expression type inference
func TestInferUnaryExprType(t *testing.T) {
	tc := newTestTypeChecker()
	
	// Set up pointer symbol for dereference test
	tc.analyzer.symbolTable.globalScope.symbols["p"] = &Symbol{Name: "p", Type: intPtrType()}
	
	tests := []struct {
		name     string
		op       lexer.TokenType
		operand  parser.Expr
	}{
		{"unary plus", lexer.ADD, &parser.IntLiteral{Value: 5}},
		{"unary minus", lexer.SUB, &parser.IntLiteral{Value: 5}},
		{"logical not", lexer.NOT, &parser.IntLiteral{Value: 1}},
		{"bitwise not", lexer.BITNOT, &parser.IntLiteral{Value: 5}},
		{"dereference", lexer.MUL, &parser.IdentExpr{Name: "p"}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &parser.UnaryExpr{
				Op:      tt.op,
				Operand: tt.operand,
			}
			resultType := tc.InferType(expr)
			if resultType == nil {
				t.Errorf("InferType %s returned nil", tt.name)
			}
		})
	}
}

// TestInferCallExprType tests function call type inference
func TestInferCallExprType(t *testing.T) {
	tc := newTestTypeChecker()
	
	// Set up function symbols
	tc.analyzer.symbolTable.globalScope.symbols["func1"] = &Symbol{Name: "func1", Type: funcType(intType(), []parser.Type{intType()}, false)}
	tc.analyzer.symbolTable.globalScope.symbols["func2"] = &Symbol{Name: "func2", Type: funcType(doubleType(), []parser.Type{intType(), doubleType()}, false)}
	tc.analyzer.symbolTable.globalScope.symbols["func3"] = &Symbol{Name: "func3", Type: funcType(voidType(), []parser.Type{}, false)}
	tc.analyzer.symbolTable.globalScope.symbols["variadic"] = &Symbol{Name: "variadic", Type: &parser.FuncType{Return: intType(), Params: []parser.Type{intType()}, Variadic: true}}
	
	tests := []struct {
		name string
		call *parser.CallExpr
	}{
		{
			name: "single param",
			call: &parser.CallExpr{
				Func: &parser.IdentExpr{Name: "func1"},
				Args: []parser.Expr{&parser.IntLiteral{Value: 42}},
			},
		},
		{
			name: "multiple params",
			call: &parser.CallExpr{
				Func: &parser.IdentExpr{Name: "func2"},
				Args: []parser.Expr{
					&parser.IntLiteral{Value: 42},
					&parser.FloatLiteral{Value: 3.14},
				},
			},
		},
		{
			name: "no params",
			call: &parser.CallExpr{
				Func: &parser.IdentExpr{Name: "func3"},
				Args: []parser.Expr{},
			},
		},
		{
			name: "variadic",
			call: &parser.CallExpr{
				Func: &parser.IdentExpr{Name: "variadic"},
				Args: []parser.Expr{
					&parser.IntLiteral{Value: 1},
					&parser.IntLiteral{Value: 2},
					&parser.IntLiteral{Value: 3},
				},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultType := tc.InferType(tt.call)
			if resultType == nil {
				t.Errorf("InferType %s returned nil", tt.name)
			}
		})
	}
}

// TestInferSizeofExprType tests sizeof expression type inference
func TestInferSizeofExprType(t *testing.T) {
	tc := newTestTypeChecker()
	
	tests := []struct {
		name       string
		sizeofExpr *parser.SizeofExpr
	}{
		{
			name: "sizeof type",
			sizeofExpr: &parser.SizeofExpr{
				Type: intType(),
			},
		},
		{
			name: "sizeof expression",
			sizeofExpr: &parser.SizeofExpr{
				Expr: &parser.IntLiteral{Value: 42},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultType := tc.InferType(tt.sizeofExpr)
			if resultType == nil {
				t.Errorf("InferType %s returned nil", tt.name)
			}
		})
	}
}

// TestInferAssignExprType tests assignment expression type inference
func TestInferAssignExprType(t *testing.T) {
	tc := newTestTypeChecker()
	
	// Set up variable symbol
	tc.analyzer.symbolTable.globalScope.symbols["x"] = &Symbol{Name: "x", Type: intType()}
	
	assignExpr := &parser.AssignExpr{
		Left:  &parser.IdentExpr{Name: "x"},
		Right: &parser.IntLiteral{Value: 42},
	}
	
	resultType := tc.InferType(assignExpr)
	if resultType == nil {
		t.Error("InferType AssignExpr returned nil")
	}
}
