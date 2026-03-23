// Package parser provides parsing of C11 source code into an Abstract Syntax Tree (AST).
// This file contains tests for expression node definitions in expr.go.
package parser

import (
	"testing"

	"github.com/akzj/goc/pkg/lexer"
)

// TestBinaryExpr tests the BinaryExpr node.
func TestBinaryExpr(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 5}

	t.Run("Pos", func(t *testing.T) {
		be := &BinaryExpr{pos: pos, end: end}
		if be.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", be.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		be := &BinaryExpr{pos: pos, end: end}
		if be.End() != end {
			t.Errorf("End() = %v, want %v", be.End(), end)
		}
	})

	t.Run("String", func(t *testing.T) {
		be := &BinaryExpr{
			Op:    lexer.ADD,
			Left:  &IntLiteral{Value: 1, pos: pos, end: end},
			Right: &IntLiteral{Value: 2, pos: pos, end: end},
			pos:   pos,
			end:   end,
		}
		s := be.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "BinaryExpr") {
			t.Errorf("String() = %q, should contain 'BinaryExpr'", s)
		}
	})
}

// TestUnaryExpr tests the UnaryExpr node.
func TestUnaryExpr(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 3}

	t.Run("Pos", func(t *testing.T) {
		ue := &UnaryExpr{pos: pos, end: end}
		if ue.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", ue.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		ue := &UnaryExpr{pos: pos, end: end}
		if ue.End() != end {
			t.Errorf("End() = %v, want %v", ue.End(), end)
		}
	})

	t.Run("String prefix", func(t *testing.T) {
		ue := &UnaryExpr{
			Op:        lexer.SUB,
			Operand:   &IntLiteral{Value: 5, pos: pos, end: end},
			IsPostfix: false,
			pos:       pos,
			end:       end,
		}
		s := ue.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "UnaryExpr") {
			t.Errorf("String() = %q, should contain 'UnaryExpr'", s)
		}
		if contains(s, "postfix") {
			t.Errorf("String() = %q, should not contain 'postfix'", s)
		}
	})

	t.Run("String postfix", func(t *testing.T) {
		ue := &UnaryExpr{
			Op:        lexer.INC,
			Operand:   &IdentExpr{Name: "x", pos: pos, end: end},
			IsPostfix: true,
			pos:       pos,
			end:       end,
		}
		s := ue.String()
		if !contains(s, "postfix=true") {
			t.Errorf("String() = %q, should contain 'postfix=true'", s)
		}
	})
}

// TestCallExpr tests the CallExpr node.
func TestCallExpr(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 10}

	t.Run("Pos", func(t *testing.T) {
		ce := &CallExpr{pos: pos, end: end}
		if ce.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", ce.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		ce := &CallExpr{pos: pos, end: end}
		if ce.End() != end {
			t.Errorf("End() = %v, want %v", ce.End(), end)
		}
	})

	t.Run("String no args", func(t *testing.T) {
		ce := &CallExpr{
			Func: &IdentExpr{Name: "foo", pos: pos, end: end},
			Args: []Expr{},
			pos:  pos,
			end:  end,
		}
		s := ce.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "CallExpr") {
			t.Errorf("String() = %q, should contain 'CallExpr'", s)
		}
	})

	t.Run("String with args", func(t *testing.T) {
		ce := &CallExpr{
			Func: &IdentExpr{Name: "printf", pos: pos, end: end},
			Args: []Expr{
				&StringLiteral{Value: "hello", pos: pos, end: end},
				&IntLiteral{Value: 42, pos: pos, end: end},
			},
			pos: pos,
			end: end,
		}
		s := ce.String()
		if !contains(s, "printf") {
			t.Errorf("String() = %q, should contain 'printf'", s)
		}
	})
}

// TestMemberExpr tests the MemberExpr node.
func TestMemberExpr(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 10}

	t.Run("Pos", func(t *testing.T) {
		me := &MemberExpr{pos: pos, end: end}
		if me.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", me.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		me := &MemberExpr{pos: pos, end: end}
		if me.End() != end {
			t.Errorf("End() = %v, want %v", me.End(), end)
		}
	})

	t.Run("String dot access", func(t *testing.T) {
		me := &MemberExpr{
			Object:    &IdentExpr{Name: "obj", pos: pos, end: end},
			Field:     "field",
			IsPointer: false,
			pos:       pos,
			end:       end,
		}
		s := me.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "MemberExpr") {
			t.Errorf("String() = %q, should contain 'MemberExpr'", s)
		}
		if !contains(s, ".") {
			t.Errorf("String() = %q, should contain '.'", s)
		}
	})

	t.Run("String arrow access", func(t *testing.T) {
		me := &MemberExpr{
			Object:    &IdentExpr{Name: "ptr", pos: pos, end: end},
			Field:     "field",
			IsPointer: true,
			pos:       pos,
			end:       end,
		}
		s := me.String()
		if !contains(s, "->") {
			t.Errorf("String() = %q, should contain '->'", s)
		}
	})
}

// TestIndexExpr tests the IndexExpr node.
func TestIndexExpr(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 10}

	t.Run("Pos", func(t *testing.T) {
		ie := &IndexExpr{pos: pos, end: end}
		if ie.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", ie.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		ie := &IndexExpr{pos: pos, end: end}
		if ie.End() != end {
			t.Errorf("End() = %v, want %v", ie.End(), end)
		}
	})

	t.Run("String", func(t *testing.T) {
		ie := &IndexExpr{
			Array: &IdentExpr{Name: "arr", pos: pos, end: end},
			Index: &IntLiteral{Value: 0, pos: pos, end: end},
			pos:   pos,
			end:   end,
		}
		s := ie.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "IndexExpr") {
			t.Errorf("String() = %q, should contain 'IndexExpr'", s)
		}
	})
}

// TestCondExpr tests the CondExpr node.
func TestCondExpr(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 20}

	t.Run("Pos", func(t *testing.T) {
		ce := &CondExpr{pos: pos, end: end}
		if ce.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", ce.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		ce := &CondExpr{pos: pos, end: end}
		if ce.End() != end {
			t.Errorf("End() = %v, want %v", ce.End(), end)
		}
	})

	t.Run("String", func(t *testing.T) {
		ce := &CondExpr{
			Cond:  &IdentExpr{Name: "x", pos: pos, end: end},
			True:  &IntLiteral{Value: 1, pos: pos, end: end},
			False: &IntLiteral{Value: 0, pos: pos, end: end},
			pos:   pos,
			end:   end,
		}
		s := ce.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "CondExpr") {
			t.Errorf("String() = %q, should contain 'CondExpr'", s)
		}
	})
}

// TestCastExpr tests the CastExpr node.
func TestCastExpr(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 15}

	t.Run("Pos", func(t *testing.T) {
		ce := &CastExpr{pos: pos, end: end}
		if ce.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", ce.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		ce := &CastExpr{pos: pos, end: end}
		if ce.End() != end {
			t.Errorf("End() = %v, want %v", ce.End(), end)
		}
	})

	t.Run("String", func(t *testing.T) {
		ce := &CastExpr{
			Type: &BaseType{Kind: TypeInt},
			Expr: &IdentExpr{Name: "x", pos: pos, end: end},
			pos:  pos,
			end:  end,
		}
		s := ce.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "CastExpr") {
			t.Errorf("String() = %q, should contain 'CastExpr'", s)
		}
	})
}

// TestSizeofExpr tests the SizeofExpr node.
func TestSizeofExpr(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 15}

	t.Run("Pos", func(t *testing.T) {
		se := &SizeofExpr{pos: pos, end: end}
		if se.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", se.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		se := &SizeofExpr{pos: pos, end: end}
		if se.End() != end {
			t.Errorf("End() = %v, want %v", se.End(), end)
		}
	})

	t.Run("String with type", func(t *testing.T) {
		se := &SizeofExpr{
			Type: &BaseType{Kind: TypeInt},
			pos:  pos,
			end:  end,
		}
		s := se.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "SizeofExpr") {
			t.Errorf("String() = %q, should contain 'SizeofExpr'", s)
		}
		if !contains(s, "type") {
			t.Errorf("String() = %q, should contain 'type'", s)
		}
	})

	t.Run("String with expr", func(t *testing.T) {
		se := &SizeofExpr{
			Expr: &IdentExpr{Name: "x", pos: pos, end: end},
			pos:  pos,
			end:  end,
		}
		s := se.String()
		if !contains(s, "expr") {
			t.Errorf("String() = %q, should contain 'expr'", s)
		}
	})
}

// TestAssignExpr tests the AssignExpr node.
func TestAssignExpr(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 10}

	t.Run("Pos", func(t *testing.T) {
		ae := &AssignExpr{pos: pos, end: end}
		if ae.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", ae.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		ae := &AssignExpr{pos: pos, end: end}
		if ae.End() != end {
			t.Errorf("End() = %v, want %v", ae.End(), end)
		}
	})

	t.Run("String", func(t *testing.T) {
		ae := &AssignExpr{
			Op:    lexer.ASSIGN,
			Left:  &IdentExpr{Name: "x", pos: pos, end: end},
			Right: &IntLiteral{Value: 42, pos: pos, end: end},
			pos:   pos,
			end:   end,
		}
		s := ae.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "AssignExpr") {
			t.Errorf("String() = %q, should contain 'AssignExpr'", s)
		}
	})
}

// TestIdentExpr tests the IdentExpr node.
func TestIdentExpr(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 5}

	t.Run("Pos", func(t *testing.T) {
		ie := &IdentExpr{pos: pos, end: end}
		if ie.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", ie.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		ie := &IdentExpr{pos: pos, end: end}
		if ie.End() != end {
			t.Errorf("End() = %v, want %v", ie.End(), end)
		}
	})

	t.Run("String", func(t *testing.T) {
		ie := &IdentExpr{Name: "myVar", pos: pos, end: end}
		s := ie.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "IdentExpr") || !contains(s, "myVar") {
			t.Errorf("String() = %q, should contain 'IdentExpr' and 'myVar'", s)
		}
	})

	t.Run("String empty name", func(t *testing.T) {
		ie := &IdentExpr{Name: "", pos: pos, end: end}
		s := ie.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
	})
}

// TestIntLiteral tests the IntLiteral node.
func TestIntLiteral(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 3}

	t.Run("Pos", func(t *testing.T) {
		il := &IntLiteral{pos: pos, end: end}
		if il.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", il.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		il := &IntLiteral{pos: pos, end: end}
		if il.End() != end {
			t.Errorf("End() = %v, want %v", il.End(), end)
		}
	})

	t.Run("String", func(t *testing.T) {
		il := &IntLiteral{Value: 42, Raw: "42", Suffix: "", pos: pos, end: end}
		s := il.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "IntLiteral") {
			t.Errorf("String() = %q, should contain 'IntLiteral'", s)
		}
	})

	t.Run("String with suffix", func(t *testing.T) {
		il := &IntLiteral{Value: 42, Raw: "42U", Suffix: "U", pos: pos, end: end}
		s := il.String()
		if !contains(s, "suffix=U") {
			t.Errorf("String() = %q, should contain 'suffix=U'", s)
		}
	})

	t.Run("String hex", func(t *testing.T) {
		il := &IntLiteral{Value: 255, Raw: "0xFF", pos: pos, end: end}
		s := il.String()
		if !contains(s, "0xFF") {
			t.Errorf("String() = %q, should contain '0xFF'", s)
		}
	})
}

// TestFloatLiteral tests the FloatLiteral node.
func TestFloatLiteral(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 5}

	t.Run("Pos", func(t *testing.T) {
		fl := &FloatLiteral{pos: pos, end: end}
		if fl.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", fl.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		fl := &FloatLiteral{pos: pos, end: end}
		if fl.End() != end {
			t.Errorf("End() = %v, want %v", fl.End(), end)
		}
	})

	t.Run("String", func(t *testing.T) {
		fl := &FloatLiteral{Value: 3.14, Raw: "3.14", Suffix: "", pos: pos, end: end}
		s := fl.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "FloatLiteral") {
			t.Errorf("String() = %q, should contain 'FloatLiteral'", s)
		}
	})

	t.Run("String with suffix", func(t *testing.T) {
		fl := &FloatLiteral{Value: 3.14, Raw: "3.14f", Suffix: "f", pos: pos, end: end}
		s := fl.String()
		if !contains(s, "suffix=f") {
			t.Errorf("String() = %q, should contain 'suffix=f'", s)
		}
	})
}

// TestCharLiteral tests the CharLiteral node.
func TestCharLiteral(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 4}

	t.Run("Pos", func(t *testing.T) {
		cl := &CharLiteral{pos: pos, end: end}
		if cl.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", cl.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		cl := &CharLiteral{pos: pos, end: end}
		if cl.End() != end {
			t.Errorf("End() = %v, want %v", cl.End(), end)
		}
	})

	t.Run("String", func(t *testing.T) {
		cl := &CharLiteral{Value: 'a', Raw: "'a'", pos: pos, end: end}
		s := cl.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "CharLiteral") {
			t.Errorf("String() = %q, should contain 'CharLiteral'", s)
		}
	})

	t.Run("String newline", func(t *testing.T) {
		cl := &CharLiteral{Value: '\n', Raw: "'\\n'", pos: pos, end: end}
		s := cl.String()
		if !contains(s, "\\n") {
			t.Errorf("String() = %q, should contain escaped newline", s)
		}
	})
}

// TestStringLiteral tests the StringLiteral node.
func TestStringLiteral(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 7}

	t.Run("Pos", func(t *testing.T) {
		sl := &StringLiteral{pos: pos, end: end}
		if sl.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", sl.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		sl := &StringLiteral{pos: pos, end: end}
		if sl.End() != end {
			t.Errorf("End() = %v, want %v", sl.End(), end)
		}
	})

	t.Run("String", func(t *testing.T) {
		sl := &StringLiteral{Value: "hello", Raw: "\"hello\"", pos: pos, end: end}
		s := sl.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "StringLiteral") {
			t.Errorf("String() = %q, should contain 'StringLiteral'", s)
		}
	})

	t.Run("String empty", func(t *testing.T) {
		sl := &StringLiteral{Value: "", Raw: "\"\"", pos: pos, end: end}
		s := sl.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
	})
}

// TestInitListExpr tests the InitListExpr node.
func TestInitListExpr(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 20}

	t.Run("Pos", func(t *testing.T) {
		ile := &InitListExpr{pos: pos, end: end}
		if ile.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", ile.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		ile := &InitListExpr{pos: pos, end: end}
		if ile.End() != end {
			t.Errorf("End() = %v, want %v", ile.End(), end)
		}
	})

	t.Run("String empty", func(t *testing.T) {
		ile := &InitListExpr{Elements: []Expr{}, pos: pos, end: end}
		s := ile.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "InitListExpr") {
			t.Errorf("String() = %q, should contain 'InitListExpr'", s)
		}
	})

	t.Run("String with elements", func(t *testing.T) {
		ile := &InitListExpr{
			Elements: []Expr{
				&IntLiteral{Value: 1, pos: pos, end: end},
				&IntLiteral{Value: 2, pos: pos, end: end},
			},
			pos: pos,
			end: end,
		}
		s := ile.String()
		if !contains(s, "elements") {
			t.Errorf("String() = %q, should contain 'elements'", s)
		}
	})
}
