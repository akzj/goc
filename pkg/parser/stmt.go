// Package parser provides parsing of C11 source code into an Abstract Syntax Tree (AST).
// This file defines statement node types.
package parser

import (
	"fmt"
	"strings"

	"github.com/akzj/goc/pkg/lexer"
)

// Statement is the interface implemented by all statement nodes.
type Statement interface {
	Node
	stmtNode()
}

// CompoundStmt represents a block statement { ... }.
type CompoundStmt struct {
	// pos is the starting position of the block (opening brace).
	pos lexer.Position
	// end is the ending position of the block (closing brace).
	end lexer.Position
	// Statements is the list of statements in the block.
	Statements []Statement
	// Declarations is the list of declarations in the block.
	Declarations []Declaration
}

// stmtNode implements Statement.
func (c *CompoundStmt) stmtNode() {}

// Pos returns the starting position.
func (c *CompoundStmt) Pos() lexer.Position {
	return c.pos
}

// End returns the ending position.
func (c *CompoundStmt) End() lexer.Position {
	return c.end
}

// String returns a string representation.
func (c *CompoundStmt) String() string {
	var sb strings.Builder
	sb.WriteString("CompoundStmt {")
	if len(c.Declarations) > 0 {
		sb.WriteString(fmt.Sprintf(" decls:%d", len(c.Declarations)))
	}
	if len(c.Statements) > 0 {
		sb.WriteString(fmt.Sprintf(" stmts:%d", len(c.Statements)))
	}
	sb.WriteString(" }")
	return sb.String()
}

// ExprStmt represents an expression statement (expr;).
type ExprStmt struct {
	// pos is the starting position of the expression.
	pos lexer.Position
	// end is the ending position (semicolon).
	end lexer.Position
	// Expr is the expression (may be nil for empty statement).
	Expr Expr
}

// stmtNode implements Statement.
func (e *ExprStmt) stmtNode() {}

// Pos returns the starting position.
func (e *ExprStmt) Pos() lexer.Position {
	return e.pos
}

// End returns the ending position.
func (e *ExprStmt) End() lexer.Position {
	return e.end
}

// String returns a string representation.
func (e *ExprStmt) String() string {
	if e.Expr != nil {
		return fmt.Sprintf("ExprStmt(%s)", e.Expr.String())
	}
	return "ExprStmt(<empty>)"
}

// ReturnStmt represents a return statement.
type ReturnStmt struct {
	// pos is the starting position of the 'return' keyword.
	pos lexer.Position
	// end is the ending position (semicolon).
	end lexer.Position
	// Value is the return value (nil for void return).
	Value Expr
}

// stmtNode implements Statement.
func (r *ReturnStmt) stmtNode() {}

// Pos returns the starting position.
func (r *ReturnStmt) Pos() lexer.Position {
	return r.pos
}

// End returns the ending position.
func (r *ReturnStmt) End() lexer.Position {
	return r.end
}

// String returns a string representation.
func (r *ReturnStmt) String() string {
	if r.Value != nil {
		return fmt.Sprintf("ReturnStmt(%s)", r.Value.String())
	}
	return "ReturnStmt(void)"
}

// IfStmt represents an if statement.
type IfStmt struct {
	// pos is the starting position of the 'if' keyword.
	pos lexer.Position
	// end is the ending position of the entire if statement.
	end lexer.Position
	// Cond is the condition expression.
	Cond Expr
	// Then is the then-branch statement.
	Then Statement
	// Else is the else-branch statement (nil if no else).
	Else Statement
}

// stmtNode implements Statement.
func (i *IfStmt) stmtNode() {}

// Pos returns the starting position.
func (i *IfStmt) Pos() lexer.Position {
	return i.pos
}

// End returns the ending position.
func (i *IfStmt) End() lexer.Position {
	return i.end
}

// String returns a string representation.
func (i *IfStmt) String() string {
	var sb strings.Builder
	sb.WriteString("IfStmt(")
	if i.Cond != nil {
		sb.WriteString(i.Cond.String())
	}
	sb.WriteString(")")
	if i.Else != nil {
		sb.WriteString(" [has-else]")
	}
	return sb.String()
}

// WhileStmt represents a while statement.
type WhileStmt struct {
	// pos is the starting position of the 'while' keyword.
	pos lexer.Position
	// end is the ending position of the entire while statement.
	end lexer.Position
	// Cond is the condition expression.
	Cond Expr
	// Body is the loop body statement.
	Body Statement
}

// stmtNode implements Statement.
func (w *WhileStmt) stmtNode() {}

// Pos returns the starting position.
func (w *WhileStmt) Pos() lexer.Position {
	return w.pos
}

// End returns the ending position.
func (w *WhileStmt) End() lexer.Position {
	return w.end
}

// String returns a string representation.
func (w *WhileStmt) String() string {
	var sb strings.Builder
	sb.WriteString("WhileStmt(")
	if w.Cond != nil {
		sb.WriteString(w.Cond.String())
	}
	sb.WriteString(")")
	return sb.String()
}

// DoWhileStmt represents a do-while statement.
type DoWhileStmt struct {
	// pos is the starting position of the 'do' keyword.
	pos lexer.Position
	// end is the ending position of the entire do-while statement.
	end lexer.Position
	// Body is the loop body statement.
	Body Statement
	// Cond is the condition expression.
	Cond Expr
}

// stmtNode implements Statement.
func (d *DoWhileStmt) stmtNode() {}

// Pos returns the starting position.
func (d *DoWhileStmt) Pos() lexer.Position {
	return d.pos
}

// End returns the ending position.
func (d *DoWhileStmt) End() lexer.Position {
	return d.end
}

// String returns a string representation.
func (d *DoWhileStmt) String() string {
	var sb strings.Builder
	sb.WriteString("DoWhileStmt(")
	if d.Cond != nil {
		sb.WriteString(d.Cond.String())
	}
	sb.WriteString(")")
	return sb.String()
}

// ForStmt represents a for statement.
type ForStmt struct {
	// pos is the starting position of the 'for' keyword.
	pos lexer.Position
	// end is the ending position of the entire for statement.
	end lexer.Position
	// Init is the initialization expression/statement (may be nil).
	Init Node
	// Cond is the condition expression (may be nil).
	Cond Expr
	// Update is the update expression (may be nil).
	Update Expr
	// Body is the loop body statement.
	Body Statement
}

// stmtNode implements Statement.
func (f *ForStmt) stmtNode() {}

// Pos returns the starting position.
func (f *ForStmt) Pos() lexer.Position {
	return f.pos
}

// End returns the ending position.
func (f *ForStmt) End() lexer.Position {
	return f.end
}

// String returns a string representation.
func (f *ForStmt) String() string {
	var sb strings.Builder
	sb.WriteString("ForStmt(")
	parts := []string{}
	if f.Init != nil {
		if initExpr, ok := f.Init.(Expr); ok {
			parts = append(parts, initExpr.String())
		} else {
			parts = append(parts, "init")
		}
	}
	if f.Cond != nil {
		parts = append(parts, f.Cond.String())
	}
	if f.Update != nil {
		parts = append(parts, f.Update.String())
	}
	sb.WriteString(strings.Join(parts, "; "))
	sb.WriteString(")")
	return sb.String()
}

// BreakStmt represents a break statement.
type BreakStmt struct {
	// pos is the starting position of the 'break' keyword.
	pos lexer.Position
	// end is the ending position (semicolon).
	end lexer.Position
}

// stmtNode implements Statement.
func (b *BreakStmt) stmtNode() {}

// Pos returns the starting position.
func (b *BreakStmt) Pos() lexer.Position {
	return b.pos
}

// End returns the ending position.
func (b *BreakStmt) End() lexer.Position {
	return b.end
}

// String returns a string representation.
func (b *BreakStmt) String() string {
	return "BreakStmt"
}

// ContinueStmt represents a continue statement.
type ContinueStmt struct {
	// pos is the starting position of the 'continue' keyword.
	pos lexer.Position
	// end is the ending position (semicolon).
	end lexer.Position
}

// stmtNode implements Statement.
func (c *ContinueStmt) stmtNode() {}

// Pos returns the starting position.
func (c *ContinueStmt) Pos() lexer.Position {
	return c.pos
}

// End returns the ending position.
func (c *ContinueStmt) End() lexer.Position {
	return c.end
}

// String returns a string representation.
func (c *ContinueStmt) String() string {
	return "ContinueStmt"
}

// GotoStmt represents a goto statement.
type GotoStmt struct {
	// pos is the starting position of the 'goto' keyword.
	pos lexer.Position
	// end is the ending position (semicolon).
	end lexer.Position
	// Label is the target label name.
	Label string
}

// stmtNode implements Statement.
func (g *GotoStmt) stmtNode() {}

// Pos returns the starting position.
func (g *GotoStmt) Pos() lexer.Position {
	return g.pos
}

// End returns the ending position.
func (g *GotoStmt) End() lexer.Position {
	return g.end
}

// String returns a string representation.
func (g *GotoStmt) String() string {
	return fmt.Sprintf("GotoStmt(%s)", g.Label)
}

// LabelStmt represents a labeled statement.
type LabelStmt struct {
	// pos is the starting position of the label.
	pos lexer.Position
	// end is the ending position of the labeled statement.
	end lexer.Position
	// Label is the label name.
	Label string
	// Stmt is the labeled statement.
	Stmt Statement
}

// stmtNode implements Statement.
func (l *LabelStmt) stmtNode() {}

// Pos returns the starting position.
func (l *LabelStmt) Pos() lexer.Position {
	return l.pos
}

// End returns the ending position.
func (l *LabelStmt) End() lexer.Position {
	return l.end
}

// String returns a string representation.
func (l *LabelStmt) String() string {
	return fmt.Sprintf("LabelStmt(%s:)", l.Label)
}

// SwitchStmt represents a switch statement.
type SwitchStmt struct {
	// pos is the starting position of the 'switch' keyword.
	pos lexer.Position
	// end is the ending position of the entire switch statement.
	end lexer.Position
	// Cond is the switch expression.
	Cond Expr
	// Body is the switch body (compound statement with cases).
	Body *CompoundStmt
}

// stmtNode implements Statement.
func (s *SwitchStmt) stmtNode() {}

// Pos returns the starting position.
func (s *SwitchStmt) Pos() lexer.Position {
	return s.pos
}

// End returns the ending position.
func (s *SwitchStmt) End() lexer.Position {
	return s.end
}

// String returns a string representation.
func (s *SwitchStmt) String() string {
	var sb strings.Builder
	sb.WriteString("SwitchStmt(")
	if s.Cond != nil {
		sb.WriteString(s.Cond.String())
	}
	sb.WriteString(")")
	if s.Body != nil {
		sb.WriteString(fmt.Sprintf(" [cases:%d]", len(s.Body.Statements)+len(s.Body.Declarations)))
	}
	return sb.String()
}

// CaseStmt represents a case label in a switch statement.
type CaseStmt struct {
	// pos is the starting position of the 'case' or 'default' keyword.
	pos lexer.Position
	// end is the ending position of the case statement.
	end lexer.Position
	// Value is the case value expression (nil for default).
	Value Expr
	// Stmt is the statement following the case.
	Stmt Statement
}

// stmtNode implements Statement.
func (c *CaseStmt) stmtNode() {}

// Pos returns the starting position.
func (c *CaseStmt) Pos() lexer.Position {
	return c.pos
}

// End returns the ending position.
func (c *CaseStmt) End() lexer.Position {
	return c.end
}

// String returns a string representation.
func (c *CaseStmt) String() string {
	if c.Value != nil {
		return fmt.Sprintf("CaseStmt(%s:)", c.Value.String())
	}
	return "CaseStmt(default:)"
}

// ParseStatement parses a statement.
func (p *Parser) ParseStatement() Statement {
	if p.match(lexer.LBRACE) {
		return p.parseCompoundStatement()
	}
	
	if p.match(lexer.IF) {
		return p.parseIfStatement()
	}
	
	if p.match(lexer.WHILE) {
		return p.parseWhileStatement()
	}
	
	if p.match(lexer.DO) {
		return p.parseDoWhileStatement()
	}
	
	if p.match(lexer.FOR) {
		return p.parseForStatement()
	}
	
	if p.match(lexer.SWITCH) {
		return p.parseSwitchStatement()
	}
	
	if p.match(lexer.CASE) {
		return p.parseCaseStatement()
	}
	
	if p.match(lexer.DEFAULT) {
		return p.parseDefaultStatement()
	}
	
	if p.match(lexer.BREAK) {
		return p.parseBreakStatement()
	}
	
	if p.match(lexer.CONTINUE) {
		return p.parseContinueStatement()
	}
	
	if p.match(lexer.RETURN) {
		return p.parseReturnStatement()
	}
	
	if p.match(lexer.GOTO) {
		return p.parseGotoStatement()
	}
	
	if p.match(lexer.IDENT) && p.peek(1).Type == lexer.COLON {
		return p.parseLabelStatement()
	}
	
	if p.isDeclaration() {
		// ParseDeclaration consumes the full declaration including ';'.
		// Do not fall through to parseExpressionStatement (would re-parse from
		// the same line and can blow up the parser).
		p.ParseDeclaration()
		return nil
	}

	return p.parseExpressionStatement()
}

func (p *Parser) parseExpressionStatement() Statement {
	if p.match(lexer.SEMICOLON) {
		tok := p.advance()
		return &ExprStmt{
			pos: tok.Pos,
			end: tok.Pos,
		}
	}
	
	expr := p.ParseExpression()
	endPos := expr.End()
	
	if p.match(lexer.SEMICOLON) {
		endPos = p.advance().Pos
	}
	
	return &ExprStmt{
		Expr: expr,
		pos:  expr.Pos(),
		end:  endPos,
	}
}

func (p *Parser) parseIfStatement() Statement {
	startTok := p.expect(lexer.IF)
	p.expect(lexer.LPAREN)
	cond := p.ParseExpression()
	p.expect(lexer.RPAREN)
	thenStmt := p.ParseStatement()
	
	var elseStmt Statement
	if p.match(lexer.ELSE) {
		p.advance()
		elseStmt = p.ParseStatement()
	}
	
	endPos := thenStmt.End()
	if elseStmt != nil {
		endPos = elseStmt.End()
	}
	
	return &IfStmt{
		Cond: cond,
		Then: thenStmt,
		Else: elseStmt,
		pos:  startTok.Pos,
		end:  endPos,
	}
}

func (p *Parser) parseWhileStatement() Statement {
	startTok := p.expect(lexer.WHILE)
	p.expect(lexer.LPAREN)
	cond := p.ParseExpression()
	p.expect(lexer.RPAREN)
	body := p.ParseStatement()
	
	return &WhileStmt{
		Cond: cond,
		Body: body.(*CompoundStmt),
		pos:  startTok.Pos,
		end:  body.End(),
	}
}

func (p *Parser) parseDoWhileStatement() Statement {
	startTok := p.expect(lexer.DO)
	body := p.ParseStatement()
	p.expect(lexer.WHILE)
	p.expect(lexer.LPAREN)
	cond := p.ParseExpression()
	p.expect(lexer.RPAREN)
	endPos := p.expect(lexer.SEMICOLON).Pos
	
	return &DoWhileStmt{
		Body: body.(*CompoundStmt),
		Cond: cond,
		pos:  startTok.Pos,
		end:  endPos,
	}
}

func (p *Parser) parseForStatement() Statement {
	startTok := p.expect(lexer.FOR)
	p.expect(lexer.LPAREN)
	
	var init, cond, update Expr
	
	if !p.match(lexer.SEMICOLON) {
		init = p.ParseExpression()
	}
	p.expect(lexer.SEMICOLON)
	
	if !p.match(lexer.SEMICOLON) {
		cond = p.ParseExpression()
	}
	p.expect(lexer.SEMICOLON)
	
	if !p.match(lexer.RPAREN) {
		update = p.ParseExpression()
	}
	
	p.expect(lexer.RPAREN)
	body := p.ParseStatement()
	
	return &ForStmt{
		Init:   init,
		Cond:   cond,
		Update: update,
		Body:   body,
		pos:    startTok.Pos,
		end:    body.End(),
	}
}

func (p *Parser) parseSwitchStatement() Statement {
	startTok := p.expect(lexer.SWITCH)
	p.expect(lexer.LPAREN)
	cond := p.ParseExpression()
	p.expect(lexer.RPAREN)
	body := p.ParseStatement()
	
	return &SwitchStmt{
		Cond: cond,
		Body: body.(*CompoundStmt),
		pos:  startTok.Pos,
		end:  body.End(),
	}
}

func (p *Parser) parseCaseStatement() Statement {
	startTok := p.expect(lexer.CASE)
	value := p.ParseExpression()
	p.expect(lexer.COLON)
	
	var stmt Statement
	if !p.isEOF() && !p.match(lexer.CASE, lexer.DEFAULT, lexer.RBRACE) {
		stmt = p.ParseStatement()
	}
	
	endPos := value.End()
	if stmt != nil {
		endPos = stmt.End()
	}
	
	return &CaseStmt{
		Value: value,
		Stmt:  stmt,
		pos:   startTok.Pos,
		end:   endPos,
	}
}

func (p *Parser) parseDefaultStatement() Statement {
	startTok := p.expect(lexer.DEFAULT)
	p.expect(lexer.COLON)
	
	var stmt Statement
	if !p.isEOF() && !p.match(lexer.CASE, lexer.DEFAULT, lexer.RBRACE) {
		stmt = p.ParseStatement()
	}
	
	endPos := startTok.Pos
	if stmt != nil {
		endPos = stmt.End()
	}
	
	return &CaseStmt{
		Stmt: stmt,
		pos:  startTok.Pos,
		end:  endPos,
	}
}

func (p *Parser) parseBreakStatement() Statement {
	startTok := p.expect(lexer.BREAK)
	endPos := startTok.Pos
	
	if p.match(lexer.SEMICOLON) {
		endPos = p.advance().Pos
	}
	
	return &BreakStmt{
		pos: startTok.Pos,
		end: endPos,
	}
}

func (p *Parser) parseContinueStatement() Statement {
	startTok := p.expect(lexer.CONTINUE)
	endPos := startTok.Pos
	
	if p.match(lexer.SEMICOLON) {
		endPos = p.advance().Pos
	}
	
	return &ContinueStmt{
		pos: startTok.Pos,
		end: endPos,
	}
}

func (p *Parser) parseReturnStatement() Statement {
	startTok := p.expect(lexer.RETURN)
	
	var expr Expr
	endPos := startTok.Pos
	
	if !p.match(lexer.SEMICOLON) {
		expr = p.ParseExpression()
		if expr != nil {
			endPos = expr.End()
		}
	}
	
	if p.match(lexer.SEMICOLON) {
		endPos = p.advance().Pos
	}
	
	return &ReturnStmt{
		Value: expr,
		pos:  startTok.Pos,
		end:  endPos,
	}
}

func (p *Parser) parseGotoStatement() Statement {
	startTok := p.expect(lexer.GOTO)
	
	if !p.match(lexer.IDENT) {
		p.errs.Error("E1001", "expected identifier after 'goto'", toErrhandPos(p.current().Pos))
		return &GotoStmt{
			pos: startTok.Pos,
			end: startTok.Pos,
		}
	}
	
	labelTok := p.advance()
	endPos := labelTok.Pos
	
	if p.match(lexer.SEMICOLON) {
		endPos = p.advance().Pos
	}
	
	return &GotoStmt{
		Label: labelTok.Value,
		pos:   startTok.Pos,
		end:   endPos,
	}
}

func (p *Parser) parseLabelStatement() Statement {
	labelTok := p.expect(lexer.IDENT)
	p.expect(lexer.COLON)
	
	var stmt Statement
	if !p.isEOF() && !p.match(lexer.CASE, lexer.DEFAULT, lexer.RBRACE) {
		stmt = p.ParseStatement()
	}
	
	endPos := labelTok.Pos
	if stmt != nil {
		endPos = stmt.End()
	}
	
	return &LabelStmt{
		Label: labelTok.Value,
		Stmt:  stmt,
		pos:   labelTok.Pos,
		end:   endPos,
	}
}
