// Package parser provides parsing of C11 source code into an Abstract Syntax Tree (AST).
// This file defines expression node types.
package parser

import (
	"fmt"
	"strings"

	"github.com/akzj/goc/pkg/lexer"
)

// Expr is the interface implemented by all expression nodes.
type Expr interface {
	Node
	exprNode()
}

// BinaryExpr represents a binary operation (a + b, a * b, etc.).
type BinaryExpr struct {
	// Op is the operator token.
	Op lexer.TokenType
	// Left is the left operand.
	Left Expr
	// Right is the right operand.
	Right Expr
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// exprNode implements Expr.
func (b *BinaryExpr) exprNode() {}

// Pos returns the starting position.
func (b *BinaryExpr) Pos() lexer.Position {
	return b.pos
}

// End returns the ending position.
func (b *BinaryExpr) End() lexer.Position {
	return b.end
}

// String returns a string representation.
func (b *BinaryExpr) String() string {
	return fmt.Sprintf("BinaryExpr{op=%s, left=%s, right=%s}", b.Op, b.Left, b.Right)
}

// UnaryExpr represents a unary operation (-x, *x, &x, !x, etc.).
type UnaryExpr struct {
	// Op is the operator token.
	Op lexer.TokenType
	// Operand is the operand expression.
	Operand Expr
	// IsPostfix is true if this is a postfix operator (x++, x--).
	IsPostfix bool
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// exprNode implements Expr.
func (u *UnaryExpr) exprNode() {}

// Pos returns the starting position.
func (u *UnaryExpr) Pos() lexer.Position {
	return u.pos
}

// End returns the ending position.
func (u *UnaryExpr) End() lexer.Position {
	return u.end
}

// String returns a string representation.
func (u *UnaryExpr) String() string {
	if u.IsPostfix {
		return fmt.Sprintf("UnaryExpr{op=%s, operand=%s, postfix=true}", u.Op, u.Operand)
	}
	return fmt.Sprintf("UnaryExpr{op=%s, operand=%s}", u.Op, u.Operand)
}

// CallExpr represents a function call (func(args)).
type CallExpr struct {
	// Func is the function expression.
	Func Expr
	// Args is the list of argument expressions.
	Args []Expr
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// exprNode implements Expr.
func (c *CallExpr) exprNode() {}

// Pos returns the starting position.
func (c *CallExpr) Pos() lexer.Position {
	return c.pos
}

// End returns the ending position.
func (c *CallExpr) End() lexer.Position {
	return c.end
}

// String returns a string representation.
func (c *CallExpr) String() string {
	args := make([]string, len(c.Args))
	for i, arg := range c.Args {
		args[i] = arg.String()
	}
	return fmt.Sprintf("CallExpr{func=%s, args=[%s]}", c.Func, strings.Join(args, ", "))
}

// MemberExpr represents a member access (obj.field or ptr->field).
type MemberExpr struct {
	// Object is the object expression.
	Object Expr
	// Field is the field name.
	Field string
	// IsPointer is true if this is pointer access (->).
	IsPointer bool
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// exprNode implements Expr.
func (m *MemberExpr) exprNode() {}

// Pos returns the starting position.
func (m *MemberExpr) Pos() lexer.Position {
	return m.pos
}

// End returns the ending position.
func (m *MemberExpr) End() lexer.Position {
	return m.end
}

// String returns a string representation.
func (m *MemberExpr) String() string {
	op := "."
	if m.IsPointer {
		op = "->"
	}
	return fmt.Sprintf("MemberExpr{object=%s, field=%s, op=%s}", m.Object, m.Field, op)
}

// IndexExpr represents an array subscript (arr[index]).
type IndexExpr struct {
	// Array is the array expression.
	Array Expr
	// Index is the index expression.
	Index Expr
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// exprNode implements Expr.
func (i *IndexExpr) exprNode() {}

// Pos returns the starting position.
func (i *IndexExpr) Pos() lexer.Position {
	return i.pos
}

// End returns the ending position.
func (i *IndexExpr) End() lexer.Position {
	return i.end
}

// String returns a string representation.
func (i *IndexExpr) String() string {
	return fmt.Sprintf("IndexExpr{array=%s, index=%s}", i.Array, i.Index)
}

// CondExpr represents a ternary conditional (cond ? true : false).
type CondExpr struct {
	// Cond is the condition expression.
	Cond Expr
	// True is the true-branch expression.
	True Expr
	// False is the false-branch expression.
	False Expr
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// exprNode implements Expr.
func (c *CondExpr) exprNode() {}

// Pos returns the starting position.
func (c *CondExpr) Pos() lexer.Position {
	return c.pos
}

// End returns the ending position.
func (c *CondExpr) End() lexer.Position {
	return c.end
}

// String returns a string representation.
func (c *CondExpr) String() string {
	return fmt.Sprintf("CondExpr{cond=%s, true=%s, false=%s}", c.Cond, c.True, c.False)
}

// CastExpr represents a type cast ((type)expr).
type CastExpr struct {
	// Type is the target type.
	Type Type
	// Expr is the expression being cast.
	Expr Expr
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// exprNode implements Expr.
func (c *CastExpr) exprNode() {}

// Pos returns the starting position.
func (c *CastExpr) Pos() lexer.Position {
	return c.pos
}

// End returns the ending position.
func (c *CastExpr) End() lexer.Position {
	return c.end
}

// String returns a string representation.
func (c *CastExpr) String() string {
	return fmt.Sprintf("CastExpr{type=%s, expr=%s}", c.Type, c.Expr)
}

// SizeofExpr represents a sizeof expression.
type SizeofExpr struct {
	// Type is the type operand (if sizeof(type)).
	Type Type
	// Expr is the expression operand (if sizeof(expr)).
	Expr Expr
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// exprNode implements Expr.
func (s *SizeofExpr) exprNode() {}

// Pos returns the starting position.
func (s *SizeofExpr) Pos() lexer.Position {
	return s.pos
}

// End returns the ending position.
func (s *SizeofExpr) End() lexer.Position {
	return s.end
}

// String returns a string representation.
func (s *SizeofExpr) String() string {
	if s.Type != nil {
		return fmt.Sprintf("SizeofExpr{type=%s}", s.Type)
	}
	return fmt.Sprintf("SizeofExpr{expr=%s}", s.Expr)
}

// AssignExpr represents an assignment (a = b, a += b, etc.).
type AssignExpr struct {
	// Op is the assignment operator (=, +=, -=, etc.).
	Op lexer.TokenType
	// Left is the left-hand side (must be an lvalue).
	Left Expr
	// Right is the right-hand side expression.
	Right Expr
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// exprNode implements Expr.
func (a *AssignExpr) exprNode() {}

// Pos returns the starting position.
func (a *AssignExpr) Pos() lexer.Position {
	return a.pos
}

// End returns the ending position.
func (a *AssignExpr) End() lexer.Position {
	return a.end
}

// String returns a string representation.
func (a *AssignExpr) String() string {
	return fmt.Sprintf("AssignExpr{op=%s, left=%s, right=%s}", a.Op, a.Left, a.Right)
}

// IdentExpr represents an identifier (variable or function name).
type IdentExpr struct {
	// Name is the identifier name.
	Name string
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// exprNode implements Expr.
func (i *IdentExpr) exprNode() {}

// Pos returns the starting position.
func (i *IdentExpr) Pos() lexer.Position {
	return i.pos
}

// End returns the ending position.
func (i *IdentExpr) End() lexer.Position {
	return i.end
}

// String returns a string representation.
func (i *IdentExpr) String() string {
	return fmt.Sprintf("IdentExpr{name=%s}", i.Name)
}

// IntLiteral represents an integer literal.
type IntLiteral struct {
	// Value is the integer value.
	Value int64
	// Raw is the raw source text.
	Raw string
	// Suffix indicates the type suffix (u, l, ll, etc.).
	Suffix string
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// exprNode implements Expr.
func (i *IntLiteral) exprNode() {}

// Pos returns the starting position.
func (i *IntLiteral) Pos() lexer.Position {
	return i.pos
}

// End returns the ending position.
func (i *IntLiteral) End() lexer.Position {
	return i.end
}

// String returns a string representation.
func (i *IntLiteral) String() string {
	return fmt.Sprintf("IntLiteral{value=%d, raw=%s, suffix=%s}", i.Value, i.Raw, i.Suffix)
}

// FloatLiteral represents a floating-point literal.
type FloatLiteral struct {
	// Value is the float value.
	Value float64
	// Raw is the raw source text.
	Raw string
	// Suffix indicates the type suffix (f, l).
	Suffix string
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// exprNode implements Expr.
func (f *FloatLiteral) exprNode() {}

// Pos returns the starting position.
func (f *FloatLiteral) Pos() lexer.Position {
	return f.pos
}

// End returns the ending position.
func (f *FloatLiteral) End() lexer.Position {
	return f.end
}

// String returns a string representation.
func (f *FloatLiteral) String() string {
	return fmt.Sprintf("FloatLiteral{value=%v, raw=%s, suffix=%s}", f.Value, f.Raw, f.Suffix)
}

// CharLiteral represents a character literal.
type CharLiteral struct {
	// Value is the character value.
	Value rune
	// Raw is the raw source text (including quotes).
	Raw string
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// exprNode implements Expr.
func (c *CharLiteral) exprNode() {}

// Pos returns the starting position.
func (c *CharLiteral) Pos() lexer.Position {
	return c.pos
}

// End returns the ending position.
func (c *CharLiteral) End() lexer.Position {
	return c.end
}

// String returns a string representation.
func (c *CharLiteral) String() string {
	return fmt.Sprintf("CharLiteral{value=%q, raw=%s}", c.Value, c.Raw)
}

// StringLiteral represents a string literal.
type StringLiteral struct {
	// Value is the string value (without quotes).
	Value string
	// Raw is the raw source text (including quotes).
	Raw string
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// exprNode implements Expr.
func (s *StringLiteral) exprNode() {}

// Pos returns the starting position.
func (s *StringLiteral) Pos() lexer.Position {
	return s.pos
}

// End returns the ending position.
func (s *StringLiteral) End() lexer.Position {
	return s.end
}

// String returns a string representation.
func (s *StringLiteral) String() string {
	return fmt.Sprintf("StringLiteral{value=%q, raw=%s}", s.Value, s.Raw)
}

// InitListExpr represents an initializer list ({a, b, c}).
type InitListExpr struct {
	// Elements is the list of initializer elements.
	Elements []Expr
	// Designators is the list of designators for designated initializers.
	Designators []Designator
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// Designator represents a designator in a designated initializer.
type Designator struct {
	// Index is the array index (for [index] = value).
	Index Expr
	// Field is the field name (for .field = value).
	Field string
}

// exprNode implements Expr.
func (i *InitListExpr) exprNode() {}

// Pos returns the starting position.
func (i *InitListExpr) Pos() lexer.Position {
	return i.pos
}

// End returns the ending position.
func (i *InitListExpr) End() lexer.Position {
	return i.end
}

// String returns a string representation.
func (i *InitListExpr) String() string {
	elements := make([]string, len(i.Elements))
	for j, elem := range i.Elements {
		elements[j] = elem.String()
	}
	return fmt.Sprintf("InitListExpr{elements=[%s]}", strings.Join(elements, ", "))
}
