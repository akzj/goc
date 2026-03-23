// Package parser provides parsing of C11 source code into an Abstract Syntax Tree (AST).
// This file contains tests for statement node definitions in stmt.go.
package parser

import (
	"testing"

	"github.com/akzj/goc/pkg/lexer"
)

// TestCompoundStmt tests the CompoundStmt node.
func TestCompoundStmt(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 10, Column: 1}

	t.Run("Pos", func(t *testing.T) {
		cs := &CompoundStmt{pos: pos, end: end}
		if cs.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", cs.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		cs := &CompoundStmt{pos: pos, end: end}
		if cs.End() != end {
			t.Errorf("End() = %v, want %v", cs.End(), end)
		}
	})

	t.Run("String empty", func(t *testing.T) {
		cs := &CompoundStmt{pos: pos, end: end}
		s := cs.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "CompoundStmt") {
			t.Errorf("String() = %q, should contain 'CompoundStmt'", s)
		}
	})

	t.Run("String with statements", func(t *testing.T) {
		cs := &CompoundStmt{
			Statements: []Statement{
				&ReturnStmt{pos: pos, end: end},
			},
			pos: pos,
			end: end,
		}
		s := cs.String()
		if !contains(s, "stmts:1") {
			t.Errorf("String() = %q, should contain 'stmts:1'", s)
		}
	})

	t.Run("String with declarations", func(t *testing.T) {
		cs := &CompoundStmt{
			Declarations: []Declaration{
				&VarDecl{Name: "x", Type: &BaseType{Kind: TypeInt}, pos: pos, end: end},
			},
			pos: pos,
			end: end,
		}
		s := cs.String()
		if !contains(s, "decls:1") {
			t.Errorf("String() = %q, should contain 'decls:1'", s)
		}
	})
}

// TestExprStmt tests the ExprStmt node.
func TestExprStmt(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 5}

	t.Run("Pos", func(t *testing.T) {
		es := &ExprStmt{pos: pos, end: end}
		if es.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", es.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		es := &ExprStmt{pos: pos, end: end}
		if es.End() != end {
			t.Errorf("End() = %v, want %v", es.End(), end)
		}
	})

	t.Run("String with expression", func(t *testing.T) {
		es := &ExprStmt{
			Expr: &IntLiteral{Value: 42, pos: pos, end: end},
			pos:  pos,
			end:  end,
		}
		s := es.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "ExprStmt") {
			t.Errorf("String() = %q, should contain 'ExprStmt'", s)
		}
	})

	t.Run("String empty expression", func(t *testing.T) {
		es := &ExprStmt{Expr: nil, pos: pos, end: end}
		s := es.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "<empty>") {
			t.Errorf("String() = %q, should contain '<empty>'", s)
		}
	})
}

// TestReturnStmt tests the ReturnStmt node.
func TestReturnStmt(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 10}

	t.Run("Pos", func(t *testing.T) {
		rs := &ReturnStmt{pos: pos, end: end}
		if rs.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", rs.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		rs := &ReturnStmt{pos: pos, end: end}
		if rs.End() != end {
			t.Errorf("End() = %v, want %v", rs.End(), end)
		}
	})

	t.Run("String with value", func(t *testing.T) {
		rs := &ReturnStmt{
			Value: &IntLiteral{Value: 0, pos: pos, end: end},
			pos:   pos,
			end:   end,
		}
		s := rs.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "ReturnStmt") {
			t.Errorf("String() = %q, should contain 'ReturnStmt'", s)
		}
	})

	t.Run("String void", func(t *testing.T) {
		rs := &ReturnStmt{Value: nil, pos: pos, end: end}
		s := rs.String()
		if !contains(s, "void") {
			t.Errorf("String() = %q, should contain 'void'", s)
		}
	})
}

// TestIfStmt tests the IfStmt node.
func TestIfStmt(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 5, Column: 1}

	t.Run("Pos", func(t *testing.T) {
		is := &IfStmt{pos: pos, end: end}
		if is.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", is.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		is := &IfStmt{pos: pos, end: end}
		if is.End() != end {
			t.Errorf("End() = %v, want %v", is.End(), end)
		}
	})

	t.Run("String with condition", func(t *testing.T) {
		is := &IfStmt{
			Cond: &IntLiteral{Value: 1, pos: pos, end: end},
			pos:  pos,
			end:  end,
		}
		s := is.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "IfStmt") {
			t.Errorf("String() = %q, should contain 'IfStmt'", s)
		}
	})

	t.Run("String with else", func(t *testing.T) {
		is := &IfStmt{
			Cond: &IntLiteral{Value: 1, pos: pos, end: end},
			Else: &ReturnStmt{pos: pos, end: end},
			pos:  pos,
			end:  end,
		}
		s := is.String()
		if !contains(s, "has-else") {
			t.Errorf("String() = %q, should contain 'has-else'", s)
		}
	})

	t.Run("String nil condition", func(t *testing.T) {
		is := &IfStmt{Cond: nil, pos: pos, end: end}
		s := is.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
	})
}

// TestWhileStmt tests the WhileStmt node.
func TestWhileStmt(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 5, Column: 1}

	t.Run("Pos", func(t *testing.T) {
		ws := &WhileStmt{pos: pos, end: end}
		if ws.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", ws.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		ws := &WhileStmt{pos: pos, end: end}
		if ws.End() != end {
			t.Errorf("End() = %v, want %v", ws.End(), end)
		}
	})

	t.Run("String with condition", func(t *testing.T) {
		ws := &WhileStmt{
			Cond: &IntLiteral{Value: 1, pos: pos, end: end},
			pos:  pos,
			end:  end,
		}
		s := ws.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "WhileStmt") {
			t.Errorf("String() = %q, should contain 'WhileStmt'", s)
		}
	})

	t.Run("String nil condition", func(t *testing.T) {
		ws := &WhileStmt{Cond: nil, pos: pos, end: end}
		s := ws.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
	})
}

// TestDoWhileStmt tests the DoWhileStmt node.
func TestDoWhileStmt(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 5, Column: 1}

	t.Run("Pos", func(t *testing.T) {
		dws := &DoWhileStmt{pos: pos, end: end}
		if dws.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", dws.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		dws := &DoWhileStmt{pos: pos, end: end}
		if dws.End() != end {
			t.Errorf("End() = %v, want %v", dws.End(), end)
		}
	})

	t.Run("String with condition", func(t *testing.T) {
		dws := &DoWhileStmt{
			Cond: &IntLiteral{Value: 1, pos: pos, end: end},
			pos:  pos,
			end:  end,
		}
		s := dws.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "DoWhileStmt") {
			t.Errorf("String() = %q, should contain 'DoWhileStmt'", s)
		}
	})
}

// TestForStmt tests the ForStmt node.
func TestForStmt(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 5, Column: 1}

	t.Run("Pos", func(t *testing.T) {
		fs := &ForStmt{pos: pos, end: end}
		if fs.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", fs.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		fs := &ForStmt{pos: pos, end: end}
		if fs.End() != end {
			t.Errorf("End() = %v, want %v", fs.End(), end)
		}
	})

	t.Run("String empty", func(t *testing.T) {
		fs := &ForStmt{pos: pos, end: end}
		s := fs.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "ForStmt") {
			t.Errorf("String() = %q, should contain 'ForStmt'", s)
		}
	})

	t.Run("String with init", func(t *testing.T) {
		fs := &ForStmt{
			Init: &IdentExpr{Name: "i", pos: pos, end: end},
			pos:  pos,
			end:  end,
		}
		s := fs.String()
		if !contains(s, "IdentExpr") {
			t.Errorf("String() = %q, should contain init expression", s)
		}
	})

	t.Run("String with cond", func(t *testing.T) {
		fs := &ForStmt{
			Cond: &BinaryExpr{Op: lexer.LSS, pos: pos, end: end},
			pos:  pos,
			end:  end,
		}
		s := fs.String()
		if !contains(s, "BinaryExpr") {
			t.Errorf("String() = %q, should contain condition", s)
		}
	})

	t.Run("String with update", func(t *testing.T) {
		fs := &ForStmt{
			Update: &UnaryExpr{Op: lexer.INC, pos: pos, end: end},
			pos:    pos,
			end:    end,
		}
		s := fs.String()
		if !contains(s, "UnaryExpr") {
			t.Errorf("String() = %q, should contain update", s)
		}
	})

	t.Run("String with all parts", func(t *testing.T) {
		fs := &ForStmt{
			Init:   &IdentExpr{Name: "i", pos: pos, end: end},
			Cond:   &BinaryExpr{Op: lexer.LSS, pos: pos, end: end},
			Update: &UnaryExpr{Op: lexer.INC, pos: pos, end: end},
			pos:    pos,
			end:    end,
		}
		s := fs.String()
		if !contains(s, "ForStmt") {
			t.Errorf("String() = %q, should contain 'ForStmt'", s)
		}
	})
}

// TestBreakStmt tests the BreakStmt node.
func TestBreakStmt(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 6}

	t.Run("Pos", func(t *testing.T) {
		bs := &BreakStmt{pos: pos, end: end}
		if bs.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", bs.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		bs := &BreakStmt{pos: pos, end: end}
		if bs.End() != end {
			t.Errorf("End() = %v, want %v", bs.End(), end)
		}
	})

	t.Run("String", func(t *testing.T) {
		bs := &BreakStmt{pos: pos, end: end}
		s := bs.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if s != "BreakStmt" {
			t.Errorf("String() = %q, want 'BreakStmt'", s)
		}
	})
}

// TestContinueStmt tests the ContinueStmt node.
func TestContinueStmt(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 9}

	t.Run("Pos", func(t *testing.T) {
		cs := &ContinueStmt{pos: pos, end: end}
		if cs.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", cs.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		cs := &ContinueStmt{pos: pos, end: end}
		if cs.End() != end {
			t.Errorf("End() = %v, want %v", cs.End(), end)
		}
	})

	t.Run("String", func(t *testing.T) {
		cs := &ContinueStmt{pos: pos, end: end}
		s := cs.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if s != "ContinueStmt" {
			t.Errorf("String() = %q, want 'ContinueStmt'", s)
		}
	})
}

// TestGotoStmt tests the GotoStmt node.
func TestGotoStmt(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 10}

	t.Run("Pos", func(t *testing.T) {
		gs := &GotoStmt{pos: pos, end: end}
		if gs.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", gs.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		gs := &GotoStmt{pos: pos, end: end}
		if gs.End() != end {
			t.Errorf("End() = %v, want %v", gs.End(), end)
		}
	})

	t.Run("String with label", func(t *testing.T) {
		gs := &GotoStmt{Label: "end", pos: pos, end: end}
		s := gs.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "GotoStmt") || !contains(s, "end") {
			t.Errorf("String() = %q, should contain 'GotoStmt' and 'end'", s)
		}
	})

	t.Run("String empty label", func(t *testing.T) {
		gs := &GotoStmt{Label: "", pos: pos, end: end}
		s := gs.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
	})
}

// TestLabelStmt tests the LabelStmt node.
func TestLabelStmt(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 2, Column: 1}

	t.Run("Pos", func(t *testing.T) {
		ls := &LabelStmt{pos: pos, end: end}
		if ls.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", ls.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		ls := &LabelStmt{pos: pos, end: end}
		if ls.End() != end {
			t.Errorf("End() = %v, want %v", ls.End(), end)
		}
	})

	t.Run("String with label", func(t *testing.T) {
		ls := &LabelStmt{Label: "start", pos: pos, end: end}
		s := ls.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "LabelStmt") || !contains(s, "start") {
			t.Errorf("String() = %q, should contain 'LabelStmt' and 'start'", s)
		}
	})
}

// TestSwitchStmt tests the SwitchStmt node.
func TestSwitchStmt(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 10, Column: 1}

	t.Run("Pos", func(t *testing.T) {
		ss := &SwitchStmt{pos: pos, end: end}
		if ss.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", ss.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		ss := &SwitchStmt{pos: pos, end: end}
		if ss.End() != end {
			t.Errorf("End() = %v, want %v", ss.End(), end)
		}
	})

	t.Run("String with condition", func(t *testing.T) {
		ss := &SwitchStmt{
			Cond: &IdentExpr{Name: "x", pos: pos, end: end},
			pos:  pos,
			end:  end,
		}
		s := ss.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "SwitchStmt") {
			t.Errorf("String() = %q, should contain 'SwitchStmt'", s)
		}
	})

	t.Run("String with body", func(t *testing.T) {
		ss := &SwitchStmt{
			Cond: &IdentExpr{Name: "x", pos: pos, end: end},
			Body: &CompoundStmt{
				Statements: []Statement{
					&CaseStmt{pos: pos, end: end},
				},
				pos: pos,
				end: end,
			},
			pos: pos,
			end: end,
		}
		s := ss.String()
		if !contains(s, "cases:") {
			t.Errorf("String() = %q, should contain 'cases:'", s)
		}
	})

	t.Run("String nil condition", func(t *testing.T) {
		ss := &SwitchStmt{Cond: nil, pos: pos, end: end}
		s := ss.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
	})
}

// TestCaseStmt tests the CaseStmt node.
func TestCaseStmt(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 10}

	t.Run("Pos", func(t *testing.T) {
		cs := &CaseStmt{pos: pos, end: end}
		if cs.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", cs.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		cs := &CaseStmt{pos: pos, end: end}
		if cs.End() != end {
			t.Errorf("End() = %v, want %v", cs.End(), end)
		}
	})

	t.Run("String with value", func(t *testing.T) {
		cs := &CaseStmt{
			Value: &IntLiteral{Value: 1, pos: pos, end: end},
			pos:   pos,
			end:   end,
		}
		s := cs.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "CaseStmt") {
			t.Errorf("String() = %q, should contain 'CaseStmt'", s)
		}
	})

	t.Run("String default", func(t *testing.T) {
		cs := &CaseStmt{Value: nil, pos: pos, end: end}
		s := cs.String()
		if !contains(s, "default") {
			t.Errorf("String() = %q, should contain 'default'", s)
		}
	})
}
