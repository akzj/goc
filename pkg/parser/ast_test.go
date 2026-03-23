// Package parser provides parsing of C11 source code into an Abstract Syntax Tree (AST).
// This file contains tests for AST node definitions in ast.go.
package parser

import (
	"testing"

	"github.com/akzj/goc/pkg/lexer"
)

// TestTranslationUnit tests the TranslationUnit node.
func TestTranslationUnit(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 10, Column: 1}

	t.Run("Pos", func(t *testing.T) {
		tu := &TranslationUnit{pos: pos, end: end}
		got := tu.Pos()
		if got != pos {
			t.Errorf("Pos() = %v, want %v", got, pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		tu := &TranslationUnit{pos: pos, end: end}
		got := tu.End()
		if got != end {
			t.Errorf("End() = %v, want %v", got, end)
		}
	})

	t.Run("String with declarations", func(t *testing.T) {
		tu := &TranslationUnit{
			pos: pos,
			end: end,
			Declarations: []Declaration{
				&FunctionDecl{Name: "main", pos: pos, end: end},
			},
		}
		s := tu.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "TranslationUnit") {
			t.Errorf("String() = %q, should contain 'TranslationUnit'", s)
		}
	})

	t.Run("String empty", func(t *testing.T) {
		tu := &TranslationUnit{pos: pos, end: end, Declarations: []Declaration{}}
		s := tu.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
	})
}

// TestFunctionDecl tests the FunctionDecl node.
func TestFunctionDecl(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 5, Column: 1}

	t.Run("Pos", func(t *testing.T) {
		fd := &FunctionDecl{pos: pos, end: end}
		if fd.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", fd.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		fd := &FunctionDecl{pos: pos, end: end}
		if fd.End() != end {
			t.Errorf("End() = %v, want %v", fd.End(), end)
		}
	})

	t.Run("String basic", func(t *testing.T) {
		fd := &FunctionDecl{
			Name: "main",
			pos:  pos,
			end:  end,
		}
		s := fd.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "FunctionDecl") || !contains(s, "main") {
			t.Errorf("String() = %q, should contain 'FunctionDecl' and 'main'", s)
		}
	})

	t.Run("String with static", func(t *testing.T) {
		fd := &FunctionDecl{Name: "func", IsStatic: true, pos: pos, end: end}
		s := fd.String()
		if !contains(s, "static") {
			t.Errorf("String() = %q, should contain 'static'", s)
		}
	})

	t.Run("String with inline", func(t *testing.T) {
		fd := &FunctionDecl{Name: "func", IsInline: true, pos: pos, end: end}
		s := fd.String()
		if !contains(s, "inline") {
			t.Errorf("String() = %q, should contain 'inline'", s)
		}
	})

	t.Run("String with extern", func(t *testing.T) {
		fd := &FunctionDecl{Name: "func", IsExtern: true, pos: pos, end: end}
		s := fd.String()
		if !contains(s, "extern") {
			t.Errorf("String() = %q, should contain 'extern'", s)
		}
	})

	t.Run("String with body", func(t *testing.T) {
		fd := &FunctionDecl{
			Name: "main",
			Body: &CompoundStmt{pos: pos, end: end},
			pos:  pos,
			end:  end,
		}
		s := fd.String()
		if !contains(s, "{ ... }") {
			t.Errorf("String() = %q, should contain '{ ... }'", s)
		}
	})

	t.Run("String without body", func(t *testing.T) {
		fd := &FunctionDecl{Name: "decl", pos: pos, end: end}
		s := fd.String()
		if !contains(s, ";") {
			t.Errorf("String() = %q, should contain ';'", s)
		}
	})

	t.Run("String with params", func(t *testing.T) {
		fd := &FunctionDecl{
			Name: "func",
			Params: []*ParamDecl{
				{Name: "a", Type: &BaseType{Kind: TypeInt}, pos: pos, end: end},
				{Name: "b", Type: &BaseType{Kind: TypeInt}, pos: pos, end: end},
			},
			pos: pos,
			end: end,
		}
		s := fd.String()
		if !contains(s, "a") || !contains(s, "b") {
			t.Errorf("String() = %q, should contain param names", s)
		}
	})
}

// TestVarDecl tests the VarDecl node.
func TestVarDecl(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 10}

	t.Run("Pos", func(t *testing.T) {
		vd := &VarDecl{pos: pos, end: end}
		if vd.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", vd.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		vd := &VarDecl{pos: pos, end: end}
		if vd.End() != end {
			t.Errorf("End() = %v, want %v", vd.End(), end)
		}
	})

	t.Run("String basic", func(t *testing.T) {
		vd := &VarDecl{
			Name: "x",
			Type: &BaseType{Kind: TypeInt},
			pos:  pos,
			end:  end,
		}
		s := vd.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "VarDecl") || !contains(s, "x") {
			t.Errorf("String() = %q, should contain 'VarDecl' and 'x'", s)
		}
	})

	t.Run("String with static", func(t *testing.T) {
		vd := &VarDecl{Name: "x", IsStatic: true, Type: &BaseType{Kind: TypeInt}, pos: pos, end: end}
		s := vd.String()
		if !contains(s, "static") {
			t.Errorf("String() = %q, should contain 'static'", s)
		}
	})

	t.Run("String with extern", func(t *testing.T) {
		vd := &VarDecl{Name: "x", IsExtern: true, Type: &BaseType{Kind: TypeInt}, pos: pos, end: end}
		s := vd.String()
		if !contains(s, "extern") {
			t.Errorf("String() = %q, should contain 'extern'", s)
		}
	})

	t.Run("String with const", func(t *testing.T) {
		vd := &VarDecl{Name: "x", IsConst: true, Type: &BaseType{Kind: TypeInt}, pos: pos, end: end}
		s := vd.String()
		if !contains(s, "const") {
			t.Errorf("String() = %q, should contain 'const'", s)
		}
	})

	t.Run("String with init", func(t *testing.T) {
		vd := &VarDecl{
			Name: "x",
			Type: &BaseType{Kind: TypeInt},
			Init: &IntLiteral{Value: 42, pos: pos, end: end},
			pos:  pos,
			end:  end,
		}
		s := vd.String()
		if !contains(s, "=") {
			t.Errorf("String() = %q, should contain '='", s)
		}
	})
}

// TestParamDecl tests the ParamDecl node.
func TestParamDecl(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 5}

	t.Run("Pos", func(t *testing.T) {
		pd := &ParamDecl{pos: pos, end: end}
		if pd.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", pd.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		pd := &ParamDecl{pos: pos, end: end}
		if pd.End() != end {
			t.Errorf("End() = %v, want %v", pd.End(), end)
		}
	})

	t.Run("String with name", func(t *testing.T) {
		pd := &ParamDecl{
			Name: "arg",
			Type: &BaseType{Kind: TypeInt},
			pos:  pos,
			end:  end,
		}
		s := pd.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "arg") {
			t.Errorf("String() = %q, should contain 'arg'", s)
		}
	})

	t.Run("String without name", func(t *testing.T) {
		pd := &ParamDecl{
			Name: "",
			Type: &BaseType{Kind: TypeInt},
			pos:  pos,
			end:  end,
		}
		s := pd.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
	})
}

// TestStructDecl tests the StructDecl node.
func TestStructDecl(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 5, Column: 1}

	t.Run("Pos", func(t *testing.T) {
		sd := &StructDecl{pos: pos, end: end}
		if sd.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", sd.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		sd := &StructDecl{pos: pos, end: end}
		if sd.End() != end {
			t.Errorf("End() = %v, want %v", sd.End(), end)
		}
	})

	t.Run("String struct with name", func(t *testing.T) {
		sd := &StructDecl{Name: "MyStruct", pos: pos, end: end}
		s := sd.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "struct") || !contains(s, "MyStruct") {
			t.Errorf("String() = %q, should contain 'struct' and 'MyStruct'", s)
		}
	})

	t.Run("String union", func(t *testing.T) {
		sd := &StructDecl{Name: "MyUnion", IsUnion: true, pos: pos, end: end}
		s := sd.String()
		if !contains(s, "union") {
			t.Errorf("String() = %q, should contain 'union'", s)
		}
	})

	t.Run("String anonymous struct", func(t *testing.T) {
		sd := &StructDecl{Name: "", pos: pos, end: end}
		s := sd.String()
		if !contains(s, "struct") {
			t.Errorf("String() = %q, should contain 'struct'", s)
		}
	})

	t.Run("String with fields", func(t *testing.T) {
		sd := &StructDecl{
			Name: "S",
			Fields: []*FieldDecl{
				{Name: "a", Type: &BaseType{Kind: TypeInt}, pos: pos, end: end},
			},
			pos: pos,
			end: end,
		}
		s := sd.String()
		if !contains(s, "{") || !contains(s, "}") {
			t.Errorf("String() = %q, should contain braces", s)
		}
	})

	t.Run("String without fields", func(t *testing.T) {
		sd := &StructDecl{Name: "S", pos: pos, end: end}
		s := sd.String()
		if !contains(s, ";") {
			t.Errorf("String() = %q, should contain ';'", s)
		}
	})
}

// TestFieldDecl tests the FieldDecl node.
func TestFieldDecl(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 10}

	t.Run("Pos", func(t *testing.T) {
		fd := &FieldDecl{pos: pos, end: end}
		if fd.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", fd.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		fd := &FieldDecl{pos: pos, end: end}
		if fd.End() != end {
			t.Errorf("End() = %v, want %v", fd.End(), end)
		}
	})

	t.Run("String basic", func(t *testing.T) {
		fd := &FieldDecl{
			Name: "field",
			Type: &BaseType{Kind: TypeInt},
			pos:  pos,
			end:  end,
		}
		s := fd.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "FieldDecl") || !contains(s, "field") {
			t.Errorf("String() = %q, should contain 'FieldDecl' and 'field'", s)
		}
	})

	t.Run("String with bitfield", func(t *testing.T) {
		fd := &FieldDecl{
			Name:     "bits",
			Type:     &BaseType{Kind: TypeInt},
			BitWidth: &IntLiteral{Value: 4, pos: pos, end: end},
			pos:      pos,
			end:      end,
		}
		s := fd.String()
		if !contains(s, ":") {
			t.Errorf("String() = %q, should contain ':'", s)
		}
	})

	t.Run("String anonymous field", func(t *testing.T) {
		fd := &FieldDecl{
			Name: "",
			Type: &BaseType{Kind: TypeInt},
			pos:  pos,
			end:  end,
		}
		s := fd.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
	})
}

// TestEnumDecl tests the EnumDecl node.
func TestEnumDecl(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 5, Column: 1}

	t.Run("Pos", func(t *testing.T) {
		ed := &EnumDecl{pos: pos, end: end}
		if ed.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", ed.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		ed := &EnumDecl{pos: pos, end: end}
		if ed.End() != end {
			t.Errorf("End() = %v, want %v", ed.End(), end)
		}
	})

	t.Run("String with name", func(t *testing.T) {
		ed := &EnumDecl{Name: "Color", pos: pos, end: end}
		s := ed.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "enum") || !contains(s, "Color") {
			t.Errorf("String() = %q, should contain 'enum' and 'Color'", s)
		}
	})

	t.Run("String anonymous", func(t *testing.T) {
		ed := &EnumDecl{Name: "", pos: pos, end: end}
		s := ed.String()
		if !contains(s, "enum") {
			t.Errorf("String() = %q, should contain 'enum'", s)
		}
	})

	t.Run("String with values", func(t *testing.T) {
		ed := &EnumDecl{
			Name: "Color",
			Values: []*EnumValue{
				{Name: "RED", pos: pos, end: end},
				{Name: "GREEN", pos: pos, end: end},
			},
			pos: pos,
			end: end,
		}
		s := ed.String()
		if !contains(s, "{") || !contains(s, "}") {
			t.Errorf("String() = %q, should contain braces", s)
		}
	})

	t.Run("String without values", func(t *testing.T) {
		ed := &EnumDecl{Name: "Color", pos: pos, end: end}
		s := ed.String()
		if !contains(s, ";") {
			t.Errorf("String() = %q, should contain ';'", s)
		}
	})
}

// TestEnumValue tests the EnumValue node.
func TestEnumValue(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 1, Column: 1}
	end := lexer.Position{File: "test.c", Line: 1, Column: 5}

	t.Run("Pos", func(t *testing.T) {
		ev := &EnumValue{pos: pos, end: end}
		if ev.Pos() != pos {
			t.Errorf("Pos() = %v, want %v", ev.Pos(), pos)
		}
	})

	t.Run("End", func(t *testing.T) {
		ev := &EnumValue{pos: pos, end: end}
		if ev.End() != end {
			t.Errorf("End() = %v, want %v", ev.End(), end)
		}
	})

	t.Run("String without value", func(t *testing.T) {
		ev := &EnumValue{Name: "RED", pos: pos, end: end}
		s := ev.String()
		if s == "" {
			t.Error("String() returned empty string")
		}
		if !contains(s, "EnumValue") || !contains(s, "RED") {
			t.Errorf("String() = %q, should contain 'EnumValue' and 'RED'", s)
		}
	})

	t.Run("String with value", func(t *testing.T) {
		ev := &EnumValue{
			Name:  "RED",
			Value: &IntLiteral{Value: 0, pos: pos, end: end},
			pos:   pos,
			end:   end,
		}
		s := ev.String()
		if !contains(s, "=") {
			t.Errorf("String() = %q, should contain '='", s)
		}
	})
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
