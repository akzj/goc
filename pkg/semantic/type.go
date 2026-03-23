// Package semantic performs semantic analysis on the AST.
// This file defines type checking utilities.
package semantic

import (
	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/parser"
)

// TODO: Implement type checking
// Reference: docs/architecture-design-phases-2-7.md Section 4.4

// TypeChecker performs type checking.
type TypeChecker struct {
	// analyzer is the parent semantic analyzer.
	analyzer *SemanticAnalyzer
	// errors is the error handler.
	errors *errhand.ErrorHandler
}

// NewTypeChecker creates a new type checker.
func NewTypeChecker(analyzer *SemanticAnalyzer) *TypeChecker {
	// TODO: Implement
	return nil
}

// CheckAssignable checks if srcType can be assigned to dstType.
func (tc *TypeChecker) CheckAssignable(dstType, srcType parser.Type, pos lexer.Position) error {
	// TODO: Implement
	return nil
}

// CheckBinaryOp checks if binary operation is valid.
func (tc *TypeChecker) CheckBinaryOp(op lexer.TokenType, left, right parser.Type, pos lexer.Position) (parser.Type, error) {
	// TODO: Implement
	return nil, nil
}

// CheckUnaryOp checks if unary operation is valid.
func (tc *TypeChecker) CheckUnaryOp(op lexer.TokenType, operand parser.Type, pos lexer.Position) (parser.Type, error) {
	// TODO: Implement
	return nil, nil
}

// CheckCall checks if function call is valid.
func (tc *TypeChecker) CheckCall(funcType *parser.FuncType, args []parser.Type, pos lexer.Position) (parser.Type, error) {
	// TODO: Implement
	return nil, nil
}

// ImplicitCast performs implicit type conversion.
func (tc *TypeChecker) ImplicitCast(expr parser.Expr, from, to parser.Type) parser.Expr {
	// TODO: Implement
	return nil
}