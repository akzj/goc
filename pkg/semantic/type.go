// Package semantic performs semantic analysis on the AST.
// This file defines type checking utilities.
package semantic

import (
	"fmt"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/parser"
)

// TypeChecker performs type checking for semantic analysis.
// It implements C11 type compatibility rules including integer promotion,
// usual arithmetic conversions, and implicit casts.
type TypeChecker struct {
	// analyzer is the parent semantic analyzer.
	analyzer *SemanticAnalyzer
	// errors is the error handler for reporting type errors.
	errors *errhand.ErrorHandler
}

// NewTypeChecker creates a new type checker.
// The type checker uses the provided analyzer for symbol lookups and
// the error handler for reporting type errors.
func NewTypeChecker(analyzer *SemanticAnalyzer) *TypeChecker {
	return &TypeChecker{
		analyzer: analyzer,
		errors:   analyzer.errors,
	}
}

// toErrhandPos converts lexer.Position to errhand.Position.
func toErrhandPos(pos lexer.Position) errhand.Position {
	return errhand.Position{
		File:   pos.File,
		Line:   pos.Line,
		Column: pos.Column,
	}
}

// InferType infers the type of an expression by walking the AST.
// It handles all expression node types and returns the inferred type.
// For identifiers, it performs symbol table lookup.
// For complex expressions, it uses type checking methods.
// Returns nil if the type cannot be inferred (error is reported).
func (tc *TypeChecker) InferType(expr parser.Expr) parser.Type {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	// Literals
	case *parser.IntLiteral:
		return tc.inferIntLiteralType(e)
	case *parser.FloatLiteral:
		return tc.inferFloatLiteralType(e)
	case *parser.CharLiteral:
		return tc.inferCharLiteralType(e)
	case *parser.StringLiteral:
		return tc.inferStringLiteralType(e)

	// Identifiers
	case *parser.IdentExpr:
		return tc.inferIdentType(e)

	// Binary operations
	case *parser.BinaryExpr:
		return tc.inferBinaryExprType(e)

	// Unary operations
	case *parser.UnaryExpr:
		return tc.inferUnaryExprType(e)

	// Function calls
	case *parser.CallExpr:
		return tc.inferCallExprType(e)

	// Member access
	case *parser.MemberExpr:
		return tc.inferMemberExprType(e)

	// Array indexing
	case *parser.IndexExpr:
		return tc.inferIndexExprType(e)

	// Ternary conditional
	case *parser.CondExpr:
		return tc.inferCondExprType(e)

	// Type cast
	case *parser.CastExpr:
		return e.Type

	// Sizeof expression
	case *parser.SizeofExpr:
		return tc.inferSizeofExprType(e)

	// Assignment
	case *parser.AssignExpr:
		return tc.inferAssignExprType(e)

	// Initializer list
	case *parser.InitListExpr:
		return tc.inferInitListExprType(e)

	default:
		tc.errors.Error(errhand.ErrInvalidType,
			fmt.Sprintf("cannot infer type for expression type %T", e),
			toErrhandPos(expr.Pos()))
		return nil
	}
}

// inferIntLiteralType infers the type of an integer literal.
func (tc *TypeChecker) inferIntLiteralType(lit *parser.IntLiteral) parser.Type {
	// Determine type based on suffix
	suffix := lit.Suffix
	if suffix != "" {
		// Handle suffixes: u, l, ll, ul, ull, etc.
		isUnsigned := false
		isLong := 0
		for _, c := range suffix {
			switch c {
			case 'u', 'U':
				isUnsigned = true
			case 'l', 'L':
				isLong++
			}
		}
		if isLong >= 2 {
			if isUnsigned {
				return &parser.BaseType{Kind: parser.TypeLong, Signed: false, Long: 1}
			}
			return &parser.BaseType{Kind: parser.TypeLong, Signed: true, Long: 1}
		}
		if isLong == 1 {
			if isUnsigned {
				return &parser.BaseType{Kind: parser.TypeLong, Signed: false, Long: 0}
			}
			return &parser.BaseType{Kind: parser.TypeLong, Signed: true, Long: 0}
		}
		if isUnsigned {
			return &parser.BaseType{Kind: parser.TypeInt, Signed: false, Long: 0}
		}
	}
	// Default: signed int
	return &parser.BaseType{Kind: parser.TypeInt, Signed: true, Long: 0}
}

// inferFloatLiteralType infers the type of a floating-point literal.
func (tc *TypeChecker) inferFloatLiteralType(lit *parser.FloatLiteral) parser.Type {
	// Determine type based on suffix
	suffix := lit.Suffix
	if suffix != "" {
		for _, c := range suffix {
			switch c {
			case 'f', 'F':
				return &parser.BaseType{Kind: parser.TypeFloat, Signed: true}
			case 'l', 'L':
				return &parser.BaseType{Kind: parser.TypeDouble, Signed: true} // long double
			}
		}
	}
	// Default: double
	return &parser.BaseType{Kind: parser.TypeDouble, Signed: true}
}

// inferCharLiteralType infers the type of a character literal.
// In C, char literals have type int (after integer promotion).
func (tc *TypeChecker) inferCharLiteralType(lit *parser.CharLiteral) parser.Type {
	return &parser.BaseType{Kind: parser.TypeInt, Signed: true, Long: 0}
}

// inferStringLiteralType infers the type of a string literal.
// String literals have type "array of char" which decays to "pointer to char".
func (tc *TypeChecker) inferStringLiteralType(lit *parser.StringLiteral) parser.Type {
	// String literal type is array of char, which decays to char*
	charType := &parser.BaseType{Kind: parser.TypeChar, Signed: true, Long: 0}
	return &parser.PointerType{Elem: charType}
}

// inferIdentType infers the type of an identifier by symbol table lookup.
func (tc *TypeChecker) inferIdentType(ident *parser.IdentExpr) parser.Type {
	if tc.analyzer == nil {
		tc.errors.Error(errhand.ErrUndefinedSymbol,
			fmt.Sprintf("undefined identifier '%s'", ident.Name),
			toErrhandPos(ident.Pos()))
		return nil
	}

	symbol := tc.analyzer.Lookup(ident.Name)
	if symbol == nil {
		tc.errors.Error(errhand.ErrUndefinedSymbol,
			fmt.Sprintf("undefined identifier '%s'", ident.Name),
			toErrhandPos(ident.Pos()))
		return nil
	}

	return symbol.Type
}

// inferBinaryExprType infers the result type of a binary expression.
func (tc *TypeChecker) inferBinaryExprType(expr *parser.BinaryExpr) parser.Type {
	leftType := tc.InferType(expr.Left)
	rightType := tc.InferType(expr.Right)

	if leftType == nil || rightType == nil {
		return nil
	}

	resultType, err := tc.CheckBinaryOp(expr.Op, leftType, rightType, expr.Pos())
	if err != nil {
		return nil
	}
	return resultType
}

// inferUnaryExprType infers the result type of a unary expression.
func (tc *TypeChecker) inferUnaryExprType(expr *parser.UnaryExpr) parser.Type {
	operandType := tc.InferType(expr.Operand)
	if operandType == nil {
		return nil
	}

	resultType, err := tc.CheckUnaryOp(expr.Op, operandType, expr.Pos())
	if err != nil {
		return nil
	}
	return resultType
}

// inferCallExprType infers the return type of a function call.
func (tc *TypeChecker) inferCallExprType(expr *parser.CallExpr) parser.Type {
	// Infer the type of the function expression
	funcType := tc.InferType(expr.Func)
	if funcType == nil {
		return nil
	}

	// Unwrap to get function type
	funcBase := tc.unwrapType(funcType)
	funcTyped, ok := funcBase.(*parser.FuncType)
	if !ok {
		tc.errors.Error(errhand.ErrInvalidType,
			fmt.Sprintf("calling non-function type '%s'", funcType.String()),
			toErrhandPos(expr.Pos()))
		return nil
	}

	// Infer argument types
	argTypes := make([]parser.Type, len(expr.Args))
	for i, arg := range expr.Args {
		argTypes[i] = tc.InferType(arg)
		if argTypes[i] == nil {
			return nil
		}
	}

	// Check the call and get return type
	resultType, err := tc.CheckCall(funcTyped, argTypes, expr.Pos())
	if err != nil {
		return nil
	}
	return resultType
}

// inferMemberExprType infers the type of a member access expression.
func (tc *TypeChecker) inferMemberExprType(expr *parser.MemberExpr) parser.Type {
	// Infer the type of the object
	objType := tc.InferType(expr.Object)
	if objType == nil {
		return nil
	}

	objBase := tc.unwrapType(objType)

	// Handle pointer access (->)
	if expr.IsPointer {
		if ptrType, ok := objBase.(*parser.PointerType); ok {
			objBase = tc.unwrapType(ptrType.Elem)
		} else {
			tc.errors.Error(errhand.ErrInvalidType,
				fmt.Sprintf("cannot use '->' on non-pointer type '%s'", objType.String()),
				toErrhandPos(expr.Pos()))
			return nil
		}
	}

	// Get struct/union type
	var structType *parser.StructType
	switch st := objBase.(type) {
	case *parser.StructType:
		structType = st
	case *parser.QualifiedType:
		if s, ok := st.Type.(*parser.StructType); ok {
			structType = s
		}
	}

	if structType == nil {
		tc.errors.Error(errhand.ErrInvalidType,
			fmt.Sprintf("cannot access member '%s' on non-struct type '%s'", expr.Field, objType.String()),
			toErrhandPos(expr.Pos()))
		return nil
	}

	// Find the field
	for _, field := range structType.Fields {
		if field.Name == expr.Field {
			return field.Type
		}
	}

	tc.errors.Error(errhand.ErrUndefinedSymbol,
		fmt.Sprintf("unknown field '%s' in struct/union '%s'", expr.Field, structType.Name),
		toErrhandPos(expr.Pos()))
	return nil
}

// inferIndexExprType infers the type of an array indexing expression.
func (tc *TypeChecker) inferIndexExprType(expr *parser.IndexExpr) parser.Type {
	arrayType := tc.InferType(expr.Array)
	if arrayType == nil {
		return nil
	}

	arrayBase := tc.unwrapType(arrayType)

	// Handle pointer to array (array decays to pointer)
	if ptrType, ok := arrayBase.(*parser.PointerType); ok {
		arrayBase = tc.unwrapType(ptrType.Elem)
	}

	// Get array element type
	var elemType parser.Type
	switch arr := arrayBase.(type) {
	case *parser.ArrayType:
		elemType = arr.Elem
	case *parser.PointerType:
		elemType = arr.Elem
	default:
		tc.errors.Error(errhand.ErrInvalidType,
			fmt.Sprintf("cannot index non-array type '%s'", arrayType.String()),
			toErrhandPos(expr.Pos()))
		return nil
	}

	// Also check index expression type (should be integer)
	indexType := tc.InferType(expr.Index)
	if indexType != nil {
		indexBase := tc.unwrapType(indexType)
		if !tc.isIntegerType(indexBase) {
			tc.errors.Error(errhand.ErrInvalidType,
				fmt.Sprintf("array index must be integer, got '%s'", indexType.String()),
				toErrhandPos(expr.Pos()))
		}
	}

	return elemType
}

// inferCondExprType infers the type of a ternary conditional expression.
func (tc *TypeChecker) inferCondExprType(expr *parser.CondExpr) parser.Type {
	// Condition must be scalar (already checked during type checking)
	trueType := tc.InferType(expr.True)
	falseType := tc.InferType(expr.False)

	if trueType == nil || falseType == nil {
		return nil
	}

	// Both branches must have compatible types
	trueBase := tc.unwrapType(trueType)
	falseBase := tc.unwrapType(falseType)

	// If types are the same, use that type
	if tc.typesEqual(trueBase, falseBase) {
		return trueType
	}

	// If both are arithmetic types, apply usual arithmetic conversions
	if tc.isArithmeticType(trueBase) && tc.isArithmeticType(falseBase) {
		return tc.usualArithmeticConversions(trueType, falseType)
	}

	// If both are pointers, check compatibility
	if tc.isPointerType(trueBase) && tc.isPointerType(falseBase) {
		truePtr := trueBase.(*parser.PointerType)
		falsePtr := falseBase.(*parser.PointerType)
		if tc.typesEqual(truePtr.Elem, falsePtr.Elem) {
			return trueType
		}
		// Void pointer compatibility
		if tc.isVoidType(truePtr.Elem) {
			return falseType
		}
		if tc.isVoidType(falsePtr.Elem) {
			return trueType
		}
	}

	// Types are incompatible
	tc.errors.Error(errhand.ErrTypeMismatch,
		fmt.Sprintf("incompatible types in conditional: '%s' and '%s'", trueType.String(), falseType.String()),
		toErrhandPos(expr.Pos()))
	return nil
}

// inferSizeofExprType infers the type of a sizeof expression.
// sizeof always returns size_t (represented as unsigned int).
func (tc *TypeChecker) inferSizeofExprType(expr *parser.SizeofExpr) parser.Type {
	// sizeof returns size_t, which we represent as unsigned int
	return &parser.BaseType{Kind: parser.TypeInt, Signed: false, Long: 0}
}

// inferAssignExprType infers the type of an assignment expression.
// The result type is the type of the left operand.
func (tc *TypeChecker) inferAssignExprType(expr *parser.AssignExpr) parser.Type {
	leftType := tc.InferType(expr.Left)
	if leftType == nil {
		return nil
	}

	// Check assignment compatibility
	rightType := tc.InferType(expr.Right)
	if rightType != nil {
		tc.CheckAssignable(leftType, rightType, expr.Pos())
	}

	return leftType
}

// inferInitListExprType infers the type of an initializer list.
// This is context-dependent; we return a placeholder type.
func (tc *TypeChecker) inferInitListExprType(expr *parser.InitListExpr) parser.Type {
	// Initializer lists don't have a fixed type - they depend on context
	// For now, return nil (type will be determined by context)
	// In a full implementation, this would need context from the initialization target
	return nil
}

// CheckAssignable checks if srcType can be assigned to dstType according to C11 rules.
// It handles assignment compatibility, conversions, and promotions.
// Returns an error if the assignment is not valid.
func (tc *TypeChecker) CheckAssignable(dstType, srcType parser.Type, pos lexer.Position) error {
	// Unwrap typedefs and qualified types for comparison
	dstBase := tc.unwrapType(dstType)
	srcBase := tc.unwrapType(srcType)

	// Case 1: Same type (after unwrapping)
	if tc.typesEqual(dstBase, srcBase) {
		return nil
	}

	// Case 2: Both arithmetic types - implicit conversion is allowed
	if tc.isArithmeticType(dstBase) && tc.isArithmeticType(srcBase) {
		return nil
	}

	// Case 3: Void pointer compatibility
	if tc.isVoidPointer(dstBase) && tc.isObjectPointer(srcBase) {
		return nil
	}
	if tc.isObjectPointer(dstBase) && tc.isVoidPointer(srcBase) {
		return nil
	}

	// Case 4: Null pointer constant (integer constant expression with value 0)
	if tc.isPointerType(dstBase) && tc.isIntegerType(srcBase) {
		return nil
	}

	// Case 4b: Enum and int compatibility (enums are compatible with int)
	if tc.isEnumType(dstBase) && tc.isIntegerType(srcBase) {
		return nil
	}
	if tc.isIntegerType(dstBase) && tc.isEnumType(srcBase) {
		return nil
	}

	// Case 5: Pointer compatibility (same pointed-to type)
	if dstPtr, ok := dstBase.(*parser.PointerType); ok {
		if srcPtr, ok := srcBase.(*parser.PointerType); ok {
			if tc.typesEqual(dstPtr.Elem, srcPtr.Elem) {
				return nil
			}
		}
	}

	// Case 6: Array to pointer decay
	if tc.isArrayType(srcBase) && tc.isPointerType(dstBase) {
		srcArray := srcBase.(*parser.ArrayType)
		dstPtr := dstBase.(*parser.PointerType)
		if tc.typesEqual(srcArray.Elem, dstPtr.Elem) {
			return nil
		}
	}

	// Case 7: Function to pointer decay
	if tc.isFunctionType(srcBase) && tc.isPointerType(dstBase) {
		return nil
	}

	// Assignment not allowed
	tc.errors.Error(errhand.ErrTypeMismatch,
		fmt.Sprintf("cannot assign '%s' to '%s'", srcType.String(), dstType.String()),
		toErrhandPos(pos))
	return fmt.Errorf("type mismatch: cannot assign %s to %s", srcType.String(), dstType.String())
}

// CheckBinaryOp checks if a binary operation is valid and returns the result type.
// It handles operator type checking and C11 usual arithmetic conversions.
func (tc *TypeChecker) CheckBinaryOp(op lexer.TokenType, left, right parser.Type, pos lexer.Position) (parser.Type, error) {
	leftBase := tc.unwrapType(left)
	rightBase := tc.unwrapType(right)

	switch op {
	// Arithmetic operators
	case lexer.ADD, lexer.SUB, lexer.MUL, lexer.QUO, lexer.REM:
		return tc.checkArithmeticBinaryOp(op, leftBase, rightBase, pos)

	// Comparison operators
	case lexer.EQL, lexer.NEQ, lexer.LSS, lexer.GTR, lexer.LEQ, lexer.GEQ:
		return tc.checkComparisonOp(leftBase, rightBase, pos)

	// Logical operators
	case lexer.LAND, lexer.LOR:
		return tc.checkLogicalOp(leftBase, rightBase, pos)

	// Bitwise operators
	case lexer.AND, lexer.OR, lexer.XOR, lexer.SHL, lexer.SHR:
		return tc.checkBitwiseOp(op, leftBase, rightBase, pos)

	default:
		tc.errors.Error(errhand.ErrInvalidType,
			fmt.Sprintf("unknown binary operator '%s'", op),
			toErrhandPos(pos))
		return nil, fmt.Errorf("unknown binary operator: %s", op)
	}
}

// CheckUnaryOp checks if a unary operation is valid and returns the result type.
func (tc *TypeChecker) CheckUnaryOp(op lexer.TokenType, operand parser.Type, pos lexer.Position) (parser.Type, error) {
	operandBase := tc.unwrapType(operand)

	switch op {
	// Unary plus/minus
	case lexer.ADD, lexer.SUB:
		if !tc.isArithmeticType(operandBase) {
			tc.errors.Error(errhand.ErrInvalidType,
				fmt.Sprintf("invalid operand type '%s' for unary '%s'", operand.String(), op),
				toErrhandPos(pos))
			return nil, fmt.Errorf("invalid operand for unary %s: %s", op, operand.String())
		}
		return tc.integerPromotion(operandBase), nil

	// Logical NOT
	case lexer.NOT:
		if !tc.isScalarType(operandBase) {
			tc.errors.Error(errhand.ErrInvalidType,
				fmt.Sprintf("invalid operand type '%s' for logical NOT", operand.String()),
				toErrhandPos(pos))
			return nil, fmt.Errorf("invalid operand for logical NOT: %s", operand.String())
		}
		return &parser.BaseType{Kind: parser.TypeInt, Signed: true}, nil

	// Bitwise NOT
	case lexer.BITNOT:
		if !tc.isIntegerType(operandBase) {
			tc.errors.Error(errhand.ErrInvalidType,
				fmt.Sprintf("invalid operand type '%s' for bitwise NOT", operand.String()),
				toErrhandPos(pos))
			return nil, fmt.Errorf("invalid operand for bitwise NOT: %s", operand.String())
		}
		return tc.integerPromotion(operandBase), nil

	// Address-of
	case lexer.AND:
		return &parser.PointerType{Elem: operand}, nil

	// Indirection
	case lexer.MUL:
		if !tc.isPointerType(operandBase) {
			tc.errors.Error(errhand.ErrInvalidType,
				fmt.Sprintf("invalid operand type '%s' for indirection", operand.String()),
				toErrhandPos(pos))
			return nil, fmt.Errorf("invalid operand for indirection: %s", operand.String())
		}
		ptrType := operandBase.(*parser.PointerType)
		return ptrType.Elem, nil

	// Increment/Decrement
	case lexer.INC, lexer.DEC:
		if !tc.isArithmeticType(operandBase) && !tc.isPointerType(operandBase) {
			tc.errors.Error(errhand.ErrInvalidType,
				fmt.Sprintf("invalid operand type '%s' for '%s'", operand.String(), op),
				toErrhandPos(pos))
			return nil, fmt.Errorf("invalid operand for %s: %s", op, operand.String())
		}
		return operand, nil

	// Sizeof
	case lexer.SIZEOF:
		return &parser.BaseType{Kind: parser.TypeInt, Signed: true, Long: 0}, nil

	default:
		tc.errors.Error(errhand.ErrInvalidType,
			fmt.Sprintf("unknown unary operator '%s'", op),
			toErrhandPos(pos))
		return nil, fmt.Errorf("unknown unary operator: %s", op)
	}
}

// CheckCall checks if a function call is valid and returns the return type.
func (tc *TypeChecker) CheckCall(funcType *parser.FuncType, args []parser.Type, pos lexer.Position) (parser.Type, error) {
	if funcType == nil {
		tc.errors.Error(errhand.ErrInvalidType,
			"calling non-function type",
			toErrhandPos(pos))
		return nil, fmt.Errorf("calling non-function type")
	}

	params := funcType.Params
	variadic := funcType.Variadic

	// Check parameter count
	if !variadic && len(args) != len(params) {
		tc.errors.Error(errhand.ErrTypeMismatch,
			fmt.Sprintf("expected %d arguments, got %d", len(params), len(args)),
			toErrhandPos(pos))
		return nil, fmt.Errorf("argument count mismatch: expected %d, got %d", len(params), len(args))
	}

	// Variadic functions require at least as many arguments as fixed parameters
	if variadic && len(args) < len(params) {
		tc.errors.Error(errhand.ErrTypeMismatch,
			fmt.Sprintf("expected at least %d arguments, got %d", len(params), len(args)),
			toErrhandPos(pos))
		return nil, fmt.Errorf("argument count mismatch: expected at least %d, got %d", len(params), len(args))
	}

	// Check each fixed parameter
	for i := 0; i < len(params); i++ {
		paramType := params[i]
		argType := args[i]

		if err := tc.CheckAssignable(paramType, argType, pos); err != nil {
			tc.errors.Error(errhand.ErrTypeMismatch,
				fmt.Sprintf("argument %d: cannot convert '%s' to '%s'", i+1, argType.String(), paramType.String()),
				toErrhandPos(pos))
			return nil, err
		}
	}

	return funcType.Return, nil
}

// ImplicitCast performs implicit type conversion from one type to another.
func (tc *TypeChecker) ImplicitCast(expr parser.Expr, from, to parser.Type) parser.Expr {
	if expr == nil {
		return nil
	}

	fromBase := tc.unwrapType(from)
	toBase := tc.unwrapType(to)

	// No cast needed if types are the same
	if tc.typesEqual(fromBase, toBase) {
		return expr
	}

	// Arithmetic conversions
	if tc.isArithmeticType(fromBase) && tc.isArithmeticType(toBase) {
		if tc.isIntegerType(fromBase) && tc.isIntegerType(toBase) {
			if fromBase.Size() > toBase.Size() {
				return &parser.CastExpr{Type: to, Expr: expr}
			}
			return expr
		}
		if tc.isFloatType(fromBase) && tc.isFloatType(toBase) {
			// Float conversions don't need explicit cast
			return expr
		}
		if tc.isIntegerType(fromBase) && tc.isFloatType(toBase) {
			return expr
		}
		if tc.isFloatType(fromBase) && tc.isIntegerType(toBase) {
			return &parser.CastExpr{Type: to, Expr: expr}
		}
	}

	// Pointer conversions
	if tc.isPointerType(fromBase) && tc.isPointerType(toBase) {
		fromPtr := fromBase.(*parser.PointerType)
		toPtr := toBase.(*parser.PointerType)

		// void* to typed pointer is allowed without cast
		if tc.isVoidType(fromPtr.Elem) && !tc.isVoidType(toPtr.Elem) {
			return expr
		}

		// typed pointer to void* is safe (no cast needed)
		if !tc.isVoidType(fromPtr.Elem) && tc.isVoidType(toPtr.Elem) {
			return expr
		}

		if tc.typesEqual(fromPtr.Elem, toPtr.Elem) {
			return expr
		}

		return &parser.CastExpr{Type: to, Expr: expr}
	}

	// Array to pointer decay
	if tc.isArrayType(fromBase) && tc.isPointerType(toBase) {
		return expr
	}

	// Function to pointer decay
	if tc.isFunctionType(fromBase) && tc.isPointerType(toBase) {
		return expr
	}

	return &parser.CastExpr{Type: to, Expr: expr}
}

// ============================================================================
// Helper Methods
// ============================================================================

// unwrapType removes typedef and qualified type wrappers to get the base type.
func (tc *TypeChecker) unwrapType(t parser.Type) parser.Type {
	if t == nil {
		return nil
	}

	for {
		switch typ := t.(type) {
		case *parser.TypedefType:
			if typ.Underlying != nil {
				t = typ.Underlying
			} else {
				return t
			}
		case *parser.QualifiedType:
			if typ.Type != nil {
				t = typ.Type
			} else {
				return t
			}
		default:
			return t
		}
	}
}

// typesEqual checks if two types are equal (after unwrapping).
func (tc *TypeChecker) typesEqual(a, b parser.Type) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if a.TypeKind() != b.TypeKind() {
		return false
	}

	switch a.TypeKind() {
	case parser.TypeVoid, parser.TypeBool:
		return true

	case parser.TypeChar, parser.TypeShort, parser.TypeInt, parser.TypeLong,
		parser.TypeFloat, parser.TypeDouble:
		aBase := a.(*parser.BaseType)
		bBase := b.(*parser.BaseType)
		return aBase.Kind == bBase.Kind && aBase.Signed == bBase.Signed && aBase.Long == bBase.Long

	case parser.TypePointer:
		aPtr := a.(*parser.PointerType)
		bPtr := b.(*parser.PointerType)
		return tc.typesEqual(aPtr.Elem, bPtr.Elem)

	case parser.TypeArray:
		aArr := a.(*parser.ArrayType)
		bArr := b.(*parser.ArrayType)
		return tc.typesEqual(aArr.Elem, bArr.Elem) && aArr.ArraySize == bArr.ArraySize

	case parser.TypeFunction:
		aFunc := a.(*parser.FuncType)
		bFunc := b.(*parser.FuncType)
		if !tc.typesEqual(aFunc.Return, bFunc.Return) {
			return false
		}
		if len(aFunc.Params) != len(bFunc.Params) {
			return false
		}
		for i := range aFunc.Params {
			if !tc.typesEqual(aFunc.Params[i], bFunc.Params[i]) {
				return false
			}
		}
		return aFunc.Variadic == bFunc.Variadic

	case parser.TypeStruct, parser.TypeUnion:
		aStruct := a.(*parser.StructType)
		bStruct := b.(*parser.StructType)
		return aStruct.Name == bStruct.Name && aStruct.IsUnion == bStruct.IsUnion

	case parser.TypeEnum:
		aEnum := a.(*parser.EnumType)
		bEnum := b.(*parser.EnumType)
		return aEnum.Name == bEnum.Name

	case parser.TypeTypedef:
		aTypedef := a.(*parser.TypedefType)
		bTypedef := b.(*parser.TypedefType)
		return aTypedef.Name == bTypedef.Name

	default:
		return a == b
	}
}

// isArithmeticType checks if a type is an arithmetic type (integer or floating-point).
func (tc *TypeChecker) isArithmeticType(t parser.Type) bool {
	if t == nil {
		return false
	}
	switch t.TypeKind() {
	case parser.TypeBool, parser.TypeChar, parser.TypeShort, parser.TypeInt,
		parser.TypeLong, parser.TypeFloat, parser.TypeDouble:
		return true
	default:
		return false
	}
}

// isIntegerType checks if a type is an integer type.
func (tc *TypeChecker) isIntegerType(t parser.Type) bool {
	if t == nil {
		return false
	}
	switch t.TypeKind() {
	case parser.TypeBool, parser.TypeChar, parser.TypeShort, parser.TypeInt, parser.TypeLong:
		return true
	default:
		return false
	}
}

func (tc *TypeChecker) isFloatType(t parser.Type) bool {
	if t == nil {
		return false
	}
	switch t.TypeKind() {
	case parser.TypeFloat, parser.TypeDouble:
		return true
	default:
		return false
	}
}

// isScalarType checks if a type is a scalar type (arithmetic or pointer).
func (tc *TypeChecker) isScalarType(t parser.Type) bool {
	return tc.isArithmeticType(t) || tc.isPointerType(t)
}

// isPointerType checks if a type is a pointer type.
func (tc *TypeChecker) isPointerType(t parser.Type) bool {
	if t == nil {
		return false
	}
	return t.TypeKind() == parser.TypePointer
}

// isEnumType checks if a type is an enum type.
func (tc *TypeChecker) isEnumType(t parser.Type) bool {
	if t == nil {
		return false
	}
	_, ok := t.(*parser.EnumType)
	return ok
}

// isVoidType checks if a type is void.
func (tc *TypeChecker) isVoidType(t parser.Type) bool {
	if t == nil {
		return false
	}
	return t.TypeKind() == parser.TypeVoid
}

// isVoidPointer checks if a type is a pointer to void.
func (tc *TypeChecker) isVoidPointer(t parser.Type) bool {
	if t == nil {
		return false
	}
	if ptr, ok := t.(*parser.PointerType); ok {
		return tc.isVoidType(ptr.Elem)
	}
	return false
}

// isObjectPointer checks if a type is a pointer to an object type.
func (tc *TypeChecker) isObjectPointer(t parser.Type) bool {
	if t == nil {
		return false
	}
	if ptr, ok := t.(*parser.PointerType); ok {
		if tc.isVoidType(ptr.Elem) {
			return false
		}
		if tc.isFunctionType(ptr.Elem) {
			return false
		}
		return true
	}
	return false
}

// isArrayType checks if a type is an array type.
func (tc *TypeChecker) isArrayType(t parser.Type) bool {
	if t == nil {
		return false
	}
	return t.TypeKind() == parser.TypeArray
}

// isFunctionType checks if a type is a function type.
func (tc *TypeChecker) isFunctionType(t parser.Type) bool {
	if t == nil {
		return false
	}
	return t.TypeKind() == parser.TypeFunction
}

// integerPromotion applies C11 integer promotion rules.
func (tc *TypeChecker) integerPromotion(t parser.Type) parser.Type {
	if t == nil {
		return t
	}

	if !tc.isIntegerType(t) {
		return t
	}

	base := tc.unwrapType(t)
	if base.TypeKind() == parser.TypeBool || base.TypeKind() == parser.TypeChar || base.TypeKind() == parser.TypeShort {
		return &parser.BaseType{Kind: parser.TypeInt, Signed: true, Long: 0}
	}

	return t
}

// usualArithmeticConversions applies C11 usual arithmetic conversions.
func (tc *TypeChecker) usualArithmeticConversions(left, right parser.Type) parser.Type {
	leftBase := tc.unwrapType(left)
	rightBase := tc.unwrapType(right)

	leftPromoted := tc.integerPromotion(leftBase)
	rightPromoted := tc.integerPromotion(rightBase)

	if leftPromoted.TypeKind() == parser.TypeDouble || rightPromoted.TypeKind() == parser.TypeDouble {
		return &parser.BaseType{Kind: parser.TypeDouble, Signed: true}
	}

	if leftPromoted.TypeKind() == parser.TypeFloat || rightPromoted.TypeKind() == parser.TypeFloat {
		return &parser.BaseType{Kind: parser.TypeFloat, Signed: true}
	}

	return tc.integerConversions(leftPromoted, rightPromoted)
}

// integerConversions applies C11 integer conversion rules.
func (tc *TypeChecker) integerConversions(left, right parser.Type) parser.Type {
	leftBase := tc.unwrapType(left)
	rightBase := tc.unwrapType(right)

	leftSize := leftBase.Size()
	rightSize := rightBase.Size()

	if leftBase.TypeKind() == rightBase.TypeKind() {
		return leftBase
	}

	if leftSize == rightSize {
		if !leftBase.(*parser.BaseType).Signed {
			return leftBase
		}
		return rightBase
	}

	if leftSize > rightSize {
		return leftBase
	}
	return rightBase
}

// checkArithmeticBinaryOp checks arithmetic binary operators.
func (tc *TypeChecker) checkArithmeticBinaryOp(op lexer.TokenType, left, right parser.Type, pos lexer.Position) (parser.Type, error) {
	if !tc.isArithmeticType(left) {
		tc.errors.Error(errhand.ErrInvalidType,
			fmt.Sprintf("invalid left operand type '%s' for arithmetic operator", left.String()),
			toErrhandPos(pos))
		return nil, fmt.Errorf("invalid left operand for arithmetic operator: %s", left.String())
	}
	if !tc.isArithmeticType(right) {
		tc.errors.Error(errhand.ErrInvalidType,
			fmt.Sprintf("invalid right operand type '%s' for arithmetic operator", right.String()),
			toErrhandPos(pos))
		return nil, fmt.Errorf("invalid right operand for arithmetic operator: %s", right.String())
	}

	if op == lexer.REM {
		if !tc.isIntegerType(left) || !tc.isIntegerType(right) {
			tc.errors.Error(errhand.ErrInvalidType,
				"invalid operands to binary % (have '%s' and '%s')",
				toErrhandPos(pos))
			return nil, fmt.Errorf("invalid operands to binary %%")
		}
	}

	resultType := tc.usualArithmeticConversions(left, right)
	return resultType, nil
}

// checkComparisonOp checks comparison operators.
func (tc *TypeChecker) checkComparisonOp(left, right parser.Type, pos lexer.Position) (parser.Type, error) {
	leftBase := tc.unwrapType(left)
	rightBase := tc.unwrapType(right)

	if tc.isArithmeticType(leftBase) && tc.isArithmeticType(rightBase) {
		return &parser.BaseType{Kind: parser.TypeInt, Signed: true}, nil
	}

	if tc.isPointerType(leftBase) && tc.isPointerType(rightBase) {
		return &parser.BaseType{Kind: parser.TypeInt, Signed: true}, nil
	}

	if tc.isPointerType(leftBase) && tc.isIntegerType(rightBase) {
		return &parser.BaseType{Kind: parser.TypeInt, Signed: true}, nil
	}
	if tc.isIntegerType(leftBase) && tc.isPointerType(rightBase) {
		return &parser.BaseType{Kind: parser.TypeInt, Signed: true}, nil
	}

	tc.errors.Error(errhand.ErrTypeMismatch,
		fmt.Sprintf("invalid operands to comparison (have '%s' and '%s')", left.String(), right.String()),
		toErrhandPos(pos))
	return nil, fmt.Errorf("invalid operands to comparison")
}

// checkLogicalOp checks logical operators.
func (tc *TypeChecker) checkLogicalOp(left, right parser.Type, pos lexer.Position) (parser.Type, error) {
	if !tc.isScalarType(left) {
		tc.errors.Error(errhand.ErrInvalidType,
			fmt.Sprintf("invalid left operand type '%s' for logical operator", left.String()),
			toErrhandPos(pos))
		return nil, fmt.Errorf("invalid left operand for logical operator: %s", left.String())
	}
	if !tc.isScalarType(right) {
		tc.errors.Error(errhand.ErrInvalidType,
			fmt.Sprintf("invalid right operand type '%s' for logical operator", right.String()),
			toErrhandPos(pos))
		return nil, fmt.Errorf("invalid right operand for logical operator: %s", right.String())
	}

	return &parser.BaseType{Kind: parser.TypeInt, Signed: true}, nil
}

// checkBitwiseOp checks bitwise operators.
func (tc *TypeChecker) checkBitwiseOp(op lexer.TokenType, left, right parser.Type, pos lexer.Position) (parser.Type, error) {
	switch op {
	case lexer.SHL, lexer.SHR:
		if !tc.isIntegerType(left) {
			tc.errors.Error(errhand.ErrInvalidType,
				fmt.Sprintf("invalid left operand type '%s' for shift operator", left.String()),
				toErrhandPos(pos))
			return nil, fmt.Errorf("invalid left operand for shift operator: %s", left.String())
		}
		if !tc.isIntegerType(right) {
			tc.errors.Error(errhand.ErrInvalidType,
				fmt.Sprintf("invalid right operand type '%s' for shift operator", right.String()),
				toErrhandPos(pos))
			return nil, fmt.Errorf("invalid right operand for shift operator: %s", right.String())
		}
		return tc.integerPromotion(left), nil

	default:
		if !tc.isIntegerType(left) {
			tc.errors.Error(errhand.ErrInvalidType,
				fmt.Sprintf("invalid left operand type '%s' for bitwise operator", left.String()),
				toErrhandPos(pos))
			return nil, fmt.Errorf("invalid left operand for bitwise operator: %s", left.String())
		}
		if !tc.isIntegerType(right) {
			tc.errors.Error(errhand.ErrInvalidType,
				fmt.Sprintf("invalid right operand type '%s' for bitwise operator", right.String()),
				toErrhandPos(pos))
			return nil, fmt.Errorf("invalid right operand for bitwise operator: %s", right.String())
		}
		leftPromoted := tc.integerPromotion(left)
		rightPromoted := tc.integerPromotion(right)
		return tc.integerConversions(leftPromoted, rightPromoted), nil
	}
}