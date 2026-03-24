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

// CommaExpr represents a comma expression (a, b, c).
// Comma expressions evaluate left-to-right and return the rightmost value.
type CommaExpr struct {
	// Exprs is the list of expressions in order.
	Exprs []Expr
	// pos is the starting position.
	pos lexer.Position
	// end is the ending position.
	end lexer.Position
}

// exprNode implements Expr.
func (c *CommaExpr) exprNode() {}

// Pos returns the starting position.
func (c *CommaExpr) Pos() lexer.Position {
	return c.pos
}

// End returns the ending position.
func (c *CommaExpr) End() lexer.Position {
	return c.end
}

// String returns a string representation.
func (c *CommaExpr) String() string {
	exprs := make([]string, len(c.Exprs))
	for i, e := range c.Exprs {
		exprs[i] = e.String()
	}
	return fmt.Sprintf("CommaExpr{exprs=[%s]}", strings.Join(exprs, ", "))
}

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

// ParseExpression parses an expression (entry point).
// Expression = CommaExpression .
func (p *Parser) ParseExpression() Expr {
	return p.parseComma()
}

func (p *Parser) parseAssignment() Expr {
	left := p.parseConditional()
	if left == nil {
		return nil
	}
	
	for p.match(lexer.ASSIGN, lexer.ADD_ASSIGN, lexer.SUB_ASSIGN, lexer.MUL_ASSIGN,
		lexer.QUO_ASSIGN, lexer.REM_ASSIGN, lexer.SHL_ASSIGN, lexer.SHR_ASSIGN,
		lexer.AND_ASSIGN, lexer.OR_ASSIGN, lexer.XOR_ASSIGN) {
		op := p.advance().Type
		right := p.parseConditional()
		if right == nil {
			// Assignment operator without right-hand side
			// Return what we have so far
			return left
		}
		left = &AssignExpr{
			Op:    op,
			Left:  left,
			Right: right,
			pos:   left.Pos(),
			end:   right.End(),
		}
	}
	
	return left
}

// parseComma handles comma expressions (a, b, c).
// Comma expressions evaluate left-to-right and return the rightmost value.
func (p *Parser) parseComma() Expr {
	left := p.parseAssignment()
	if left == nil {
		return nil
	}
	exprs := []Expr{left}
	
	for p.match(lexer.COMMA) {
		p.advance()
		right := p.parseAssignment()
		if right == nil {
			// Comma followed by nothing valid (e.g., comma before closing paren)
			// Return just the expressions we have so far
			break
		}
		exprs = append(exprs, right)
	}
	
	if len(exprs) == 1 {
		return exprs[0]
	}
	
	return &CommaExpr{
		Exprs: exprs,
		pos:   exprs[0].Pos(),
		end:   exprs[len(exprs)-1].End(),
	}
}

func (p *Parser) parseConditional() Expr {
	cond := p.parseLogicalOr()
	
	if p.match(lexer.QUESTION) {
		p.advance()
		trueExpr := p.ParseExpression()
		p.expect(lexer.COLON)
		falseExpr := p.parseConditional()
		
		return &CondExpr{
			Cond:  cond,
			True:  trueExpr,
			False: falseExpr,
			pos:   cond.Pos(),
			end:   falseExpr.End(),
		}
	}
	
	return cond
}

func (p *Parser) parseLogicalOr() Expr {
	left := p.parseLogicalAnd()
	if left == nil {
		return nil
	}
	
	for p.match(lexer.LOR) {
		op := p.advance().Type
		right := p.parseLogicalAnd()
		if right == nil {
			return left
		}
		left = &BinaryExpr{
			Op:    op,
			Left:  left,
			Right: right,
			pos:   left.Pos(),
			end:   right.End(),
		}
	}
	
	return left
}

func (p *Parser) parseLogicalAnd() Expr {
	left := p.parseBitwiseOr()
	if left == nil {
		return nil
	}
	
	for p.match(lexer.LAND) {
		op := p.advance().Type
		right := p.parseBitwiseOr()
		if right == nil {
			return left
		}
		left = &BinaryExpr{
			Op:    op,
			Left:  left,
			Right: right,
			pos:   left.Pos(),
			end:   right.End(),
		}
	}
	
	return left
}

func (p *Parser) parseBitwiseOr() Expr {
	left := p.parseBitwiseXor()
	if left == nil {
		return nil
	}
	
	for p.match(lexer.OR) {
		op := p.advance().Type
		right := p.parseBitwiseXor()
		if right == nil {
			return left
		}
		left = &BinaryExpr{
			Op:    op,
			Left:  left,
			Right: right,
			pos:   left.Pos(),
			end:   right.End(),
		}
	}
	
	return left
}

func (p *Parser) parseBitwiseXor() Expr {
	left := p.parseBitwiseAnd()
	if left == nil {
		return nil
	}
	
	for p.match(lexer.XOR) {
		op := p.advance().Type
		right := p.parseBitwiseAnd()
		if right == nil {
			return left
		}
		left = &BinaryExpr{
			Op:    op,
			Left:  left,
			Right: right,
			pos:   left.Pos(),
			end:   right.End(),
		}
	}
	
	return left
}

func (p *Parser) parseBitwiseAnd() Expr {
	left := p.parseEquality()
	if left == nil {
		return nil
	}
	
	for p.match(lexer.AND) {
		op := p.advance().Type
		right := p.parseEquality()
		if right == nil {
			return left
		}
		left = &BinaryExpr{
			Op:    op,
			Left:  left,
			Right: right,
			pos:   left.Pos(),
			end:   right.End(),
		}
	}
	
	return left
}

func (p *Parser) parseEquality() Expr {
	left := p.parseRelational()
	if left == nil {
		return nil
	}
	
	for p.match(lexer.EQL, lexer.NEQ) {
		op := p.advance().Type
		right := p.parseRelational()
		if right == nil {
			return left
		}
		left = &BinaryExpr{
			Op:    op,
			Left:  left,
			Right: right,
			pos:   left.Pos(),
			end:   right.End(),
		}
	}
	
	return left
}

func (p *Parser) parseRelational() Expr {
	left := p.parseShift()
	if left == nil {
		return nil
	}
	
	for p.match(lexer.LEQ, lexer.GEQ, lexer.GTR, lexer.LSS) {
		op := p.advance().Type
		right := p.parseShift()
		if right == nil {
			return left
		}
		left = &BinaryExpr{
			Op:    op,
			Left:  left,
			Right: right,
			pos:   left.Pos(),
			end:   right.End(),
		}
	}
	
	return left
}

func (p *Parser) parseShift() Expr {
	left := p.parseAdditive()
	if left == nil {
		return nil
	}
	
	for p.match(lexer.SHL, lexer.SHR) {
		op := p.advance().Type
		right := p.parseAdditive()
		if right == nil {
			return left
		}
		left = &BinaryExpr{
			Op:    op,
			Left:  left,
			Right: right,
			pos:   left.Pos(),
			end:   right.End(),
		}
	}
	
	return left
}

func (p *Parser) parseAdditive() Expr {
	left := p.parseMultiplicative()
	if left == nil {
		return nil
	}
	
	for p.match(lexer.ADD, lexer.SUB) {
		op := p.advance().Type
		right := p.parseMultiplicative()
		if right == nil {
			return left
		}
		left = &BinaryExpr{
			Op:    op,
			Left:  left,
			Right: right,
			pos:   left.Pos(),
			end:   right.End(),
		}
	}
	
	return left
}

func (p *Parser) parseMultiplicative() Expr {
	left := p.parseCast()
	if left == nil {
		return nil
	}
	
	for p.match(lexer.MUL, lexer.QUO, lexer.REM) {
		op := p.advance().Type
		right := p.parseCast()
		if right == nil {
			return left
		}
		left = &BinaryExpr{
			Op:    op,
			Left:  left,
			Right: right,
			pos:   left.Pos(),
			end:   right.End(),
		}
	}
	
	return left
}

func (p *Parser) parseCast() Expr {
	if p.current().Type == lexer.LPAREN {
		savePos := p.pos
		p.advance()
		
		if p.isTypeName() {
			typ := p.parseTypeName()
			typ = p.parseDeclarator(typ)
			p.expect(lexer.RPAREN)
			expr := p.parseCast()
			if expr == nil {
				return nil
			}
			return &CastExpr{
				Type: typ,
				Expr: expr,
				pos:  p.tokens[savePos].Pos,
				end:  expr.End(),
			}
		}
		
		p.pos = savePos
	}
	
	return p.parseUnary()
}

func (p *Parser) isTypeSpecifier() bool {
	t := p.current().Type
	return t == lexer.VOID || t == lexer.CHAR || t == lexer.SHORT || t == lexer.INT || t == lexer.LONG ||
		t == lexer.FLOAT || t == lexer.DOUBLE || t == lexer.SIGNED || t == lexer.UNSIGNED || t == lexer.BOOL ||
		t == lexer.STRUCT || t == lexer.UNION || t == lexer.ENUM
}

func (p *Parser) isTypeName() bool {
	savePos := p.pos
	
	for p.isTypeSpecifier() {
		p.advance()
		if p.current().Type == lexer.STRUCT || p.current().Type == lexer.UNION || p.current().Type == lexer.ENUM {
			p.advance()
			if p.current().Type == lexer.IDENT {
				p.advance()
			}
		}
		for p.current().Type == lexer.MUL {
			p.advance()
		}
	}
	
	isType := p.pos > savePos
	p.pos = savePos
	return isType
}

func (p *Parser) parseTypeName() Type {
	specs := p.parseDeclarationSpecifiers()
	return p.specifiersToType(specs)
}

func (p *Parser) parseUnary() Expr {
	switch p.current().Type {
	case lexer.ADD, lexer.SUB, lexer.NOT, lexer.BITNOT, lexer.INC, lexer.DEC, lexer.MUL, lexer.AND:
		op := p.current().Type
		p.advance()
		operand := p.parseUnary()
		if operand == nil {
			return nil
		}
		return &UnaryExpr{
			Op:        op,
			Operand:   operand,
			IsPostfix: false,
			pos:       operand.Pos(),
			end:       operand.End(),
		}
	case lexer.SIZEOF:
		tok := p.advance()
		return p.parseSizeof(tok.Pos)
	}
	return p.parsePostfix()
}

func (p *Parser) parseSizeof(startPos lexer.Position) Expr {
	if p.current().Type == lexer.LPAREN {
		p.advance()
		
		if p.isTypeName() {
			typ := p.parseTypeName()
			typ = p.parseDeclarator(typ)
			p.expect(lexer.RPAREN)
			
			return &SizeofExpr{
				Type: typ,
				pos:  startPos,
				end:  p.current().Pos,
			}
		}
		
		expr := p.ParseExpression()
		p.expect(lexer.RPAREN)
		
		return &SizeofExpr{
			Expr: expr,
			pos:  startPos,
			end:  p.current().Pos,
		}
	}
	
	expr := p.parseUnary()
	if expr == nil {
		return nil
	}
	return &SizeofExpr{
		Expr: expr,
		pos:  startPos,
		end:  expr.End(),
	}
}

func (p *Parser) parsePostfix() Expr {
	expr := p.parsePrimary()
	if expr == nil {
		return nil
	}
	
	for {
		if p.current().Type == lexer.LPAREN {
			p.advance()
			args := p.parseArgumentList()
			p.expect(lexer.RPAREN)
			
			expr = &CallExpr{
				Func: expr,
				Args: args,
				pos:  expr.Pos(),
				end:  p.current().Pos,
			}
		} else if p.match(lexer.LBRACK) {
			p.advance()
			index := p.ParseExpression()
			p.expect(lexer.RBRACK)
			
			expr = &IndexExpr{
				Array: expr,
				Index: index,
				pos:   expr.Pos(),
				end:   p.current().Pos,
			}
		} else if p.match(lexer.DOT) {
			if !p.match(lexer.IDENT) {
				p.errs.Error("E1001", "expected identifier after '.'", toErrhandPos(p.current().Pos))
				return expr
			}
			field := p.current().Value
			endPos := p.current().Pos
			p.advance()
			
			expr = &MemberExpr{
				Object:    expr,
				Field:     field,
				IsPointer: false,
				pos:       expr.Pos(),
				end:       endPos,
			}
		} else if p.match(lexer.ARROW) {
			if !p.match(lexer.IDENT) {
				p.errs.Error("E1001", "expected identifier after '->'", toErrhandPos(p.current().Pos))
				return expr
			}
			field := p.current().Value
			endPos := p.current().Pos
			p.advance()
			
			expr = &MemberExpr{
				Object:    expr,
				Field:     field,
				IsPointer: true,
				pos:       expr.Pos(),
				end:       endPos,
			}
		} else if p.match(lexer.INC, lexer.DEC) {
			op := p.advance().Type
			endPos := p.current().Pos
			
			expr = &UnaryExpr{
				Op:        op,
				Operand:   expr,
				IsPostfix: true,
				pos:       expr.Pos(),
				end:       endPos,
			}
		} else {
			break
		}
	}
	
	return expr
}

func (p *Parser) parseArgumentList() []Expr {
	args := []Expr{}

	if p.current().Type == lexer.RPAREN {
		return args
	}

	// C: each argument is an assignment-expression, not a full comma-expression.
	// Using ParseExpression would merge "f(1, 2)" into one comma expr and one arg.
	args = append(args, p.parseAssignment())

	for p.current().Type == lexer.COMMA {
		p.advance()
		args = append(args, p.parseAssignment())
	}

	return args
}

func (p *Parser) parsePrimary() Expr {
	if p.current().Type == lexer.IDENT {
		tok := p.advance()
		return &IdentExpr{
			Name: tok.Value,
			pos:  tok.Pos,
			end:  tok.Pos,
		}
	}
	
	if p.current().Type == lexer.INT_LIT {
		tok := p.advance()
		value := p.parseIntLiteral(tok.Value)
		suffix := ""
		if len(tok.Value) > 0 {
			lastChar := tok.Value[len(tok.Value)-1]
			if lastChar == 'u' || lastChar == 'U' || lastChar == 'l' || lastChar == 'L' {
				suffix = string(lastChar)
			}
		}
		
		return &IntLiteral{
			Value:  value,
			Raw:    tok.Value,
			Suffix: suffix,
			pos:    tok.Pos,
			end:    tok.Pos,
		}
	}
	
	if p.current().Type == lexer.FLOAT_LIT {
		tok := p.advance()
		value := p.parseFloatLiteral(tok.Value)
		suffix := ""
		if len(tok.Value) > 0 {
			lastChar := tok.Value[len(tok.Value)-1]
			if lastChar == 'f' || lastChar == 'F' || lastChar == 'l' || lastChar == 'L' {
				suffix = string(lastChar)
			}
		}
		
		return &FloatLiteral{
			Value:  value,
			Raw:    tok.Value,
			Suffix: suffix,
			pos:    tok.Pos,
			end:    tok.Pos,
		}
	}
	
	if p.current().Type == lexer.CHAR_LIT {
		tok := p.advance()
		value := p.parseCharLiteral(tok.Value)
		
		return &CharLiteral{
			Value: value,
			Raw:   tok.Value,
			pos:   tok.Pos,
			end:   tok.Pos,
		}
	}
	
	if p.current().Type == lexer.STRING_LIT {
		tok := p.advance()
		value := p.parseStringLiteral(tok.Value)
		
		return &StringLiteral{
			Value: value,
			Raw:   tok.Value,
			pos:   tok.Pos,
			end:   tok.Pos,
		}
	}
	
	if p.current().Type == lexer.LPAREN {
		p.advance()
		expr := p.ParseExpression()
		p.expect(lexer.RPAREN)
		return expr
	}
	
	// Return nil for tokens that indicate end of expression (not errors)
	// This allows the caller to handle the situation appropriately
	if p.current().Type == lexer.SEMICOLON || p.current().Type == lexer.RBRACE ||
		p.current().Type == lexer.RPAREN || p.current().Type == lexer.RBRACK ||
		p.current().Type == lexer.COLON || p.current().Type == lexer.COMMA {
		return nil
	}
	
	p.errs.Error("E1001", fmt.Sprintf("unexpected token %q in expression", p.current().Value), toErrhandPos(p.current().Pos))
	p.advance()
	return nil
}

func (p *Parser) parseIntLiteral(s string) int64 {
	var value int64
	
	if len(s) > 2 {
		if s[0] == '0' {
			if s[1] == 'x' || s[1] == 'X' {
				s = s[2:]
			} else if s[1] == 'b' || s[1] == 'B' {
				s = s[2:]
			} else {
				s = s[1:]
			}
		}
	}
	
	for i, c := range s {
		if c == 'u' || c == 'U' || c == 'l' || c == 'L' {
			s = s[:i]
			break
		}
	}
	
	if s == "" {
		return 0
	}
	
	fmt.Sscanf(s, "%d", &value)
	return value
}

func (p *Parser) parseFloatLiteral(s string) float64 {
	var value float64
	fmt.Sscanf(s, "%f", &value)
	return value
}

func (p *Parser) parseCharLiteral(s string) rune {
	if len(s) >= 2 {
		s = s[1 : len(s)-1]
		if len(s) == 1 {
			return rune(s[0])
		}
		if len(s) == 2 && s[0] == '\\' {
			switch s[1] {
			case 'n':
				return '\n'
			case 't':
				return '\t'
			case 'r':
				return '\r'
			case '\\':
				return '\\'
			case '\'':
				return '\''
			case '"':
				return '"'
			case '0':
				return '\x00'
			}
		}
	}
	return 0
}

func (p *Parser) parseStringLiteral(s string) string {
	if len(s) >= 2 {
		return s[1 : len(s)-1]
	}
	return ""
}
