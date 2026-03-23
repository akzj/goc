// Package parser provides parsing of C11 source code into an Abstract Syntax Tree (AST).
// This file defines all AST node types.
package parser

import (
	"fmt"
	"strings"

	"github.com/akzj/goc/pkg/lexer"
)

// Node is the interface implemented by all AST nodes.
type Node interface {
	// Pos returns the starting position of the node.
	Pos() lexer.Position
	// End returns the ending position of the node.
	End() lexer.Position
	// String returns a string representation for debugging.
	String() string
}

// TranslationUnit represents the root of the AST (a complete C source file).
type TranslationUnit struct {
	// pos is the starting position in source code.
	pos lexer.Position
	// end is the ending position in source code.
	end lexer.Position
	// Declarations is the list of top-level declarations.
	Declarations []Declaration
}

// Pos returns the starting position.
func (t *TranslationUnit) Pos() lexer.Position {
	return t.pos
}

// End returns the ending position.
func (t *TranslationUnit) End() lexer.Position {
	return t.end
}

// String returns a string representation.
func (t *TranslationUnit) String() string {
	var sb strings.Builder
	sb.WriteString("TranslationUnit {\n")
	for i, decl := range t.Declarations {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("  %s", decl.String()))
	}
	if len(t.Declarations) > 0 {
		sb.WriteString("\n")
	}
	sb.WriteString("}")
	return sb.String()
}

// Declaration is the interface implemented by all declaration nodes.
type Declaration interface {
	Node
	declNode()
}

// FunctionDecl represents a function declaration or definition.
type FunctionDecl struct {
	// pos is the starting position in source code.
	pos lexer.Position
	// end is the ending position in source code.
	end lexer.Position
	// Type is the function type (return type).
	Type Type
	// Name is the function name.
	Name string
	// Params is the list of parameters.
	Params []*ParamDecl
	// Body is the function body (nil for declarations).
	Body *CompoundStmt
	// IsInline is true if the function is inline.
	IsInline bool
	// IsStatic is true if the function has static storage.
	IsStatic bool
	// IsExtern is true if the function has external linkage.
	IsExtern bool
}

// declNode implements Declaration.
func (f *FunctionDecl) declNode() {}

// Pos returns the starting position.
func (f *FunctionDecl) Pos() lexer.Position {
	return f.pos
}

// End returns the ending position.
func (f *FunctionDecl) End() lexer.Position {
	return f.end
}

// String returns a string representation.
func (f *FunctionDecl) String() string {
	var sb strings.Builder
	sb.WriteString("FunctionDecl ")
	if f.IsStatic {
		sb.WriteString("static ")
	}
	if f.IsInline {
		sb.WriteString("inline ")
	}
	if f.IsExtern {
		sb.WriteString("extern ")
	}
	sb.WriteString(fmt.Sprintf("%s(", f.Name))
	for i, param := range f.Params {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(param.String())
	}
	sb.WriteString(")")
	if f.Body != nil {
		sb.WriteString(" { ... }")
	} else {
		sb.WriteString(";")
	}
	return sb.String()
}

// VarDecl represents a variable declaration.
type VarDecl struct {
	// pos is the starting position in source code.
	pos lexer.Position
	// end is the ending position in source code.
	end lexer.Position
	// Type is the variable type.
	Type Type
	// Name is the variable name.
	Name string
	// Init is the initializer expression (nil if uninitialized).
	Init Expr
	// IsStatic is true if the variable has static storage.
	IsStatic bool
	// IsExtern is true if the variable has external linkage.
	IsExtern bool
	// IsConst is true if the variable is const.
	IsConst bool
}

// declNode implements Declaration.
func (v *VarDecl) declNode() {}

// Pos returns the starting position.
func (v *VarDecl) Pos() lexer.Position {
	return v.pos
}

// End returns the ending position.
func (v *VarDecl) End() lexer.Position {
	return v.end
}

// String returns a string representation.
func (v *VarDecl) String() string {
	var sb strings.Builder
	if v.IsStatic {
		sb.WriteString("static ")
	}
	if v.IsExtern {
		sb.WriteString("extern ")
	}
	if v.IsConst {
		sb.WriteString("const ")
	}
	sb.WriteString(fmt.Sprintf("VarDecl %s: %s", v.Name, v.Type.String()))
	if v.Init != nil {
		sb.WriteString(fmt.Sprintf(" = %s", v.Init.String()))
	}
	return sb.String()
}

// ParamDecl represents a function parameter.
type ParamDecl struct {
	// pos is the starting position in source code.
	pos lexer.Position
	// end is the ending position in source code.
	end lexer.Position
	// Type is the parameter type.
	Type Type
	// Name is the parameter name (may be empty).
	Name string
}

// Pos returns the starting position.
func (p *ParamDecl) Pos() lexer.Position {
	return p.pos
}

// End returns the ending position.
func (p *ParamDecl) End() lexer.Position {
	return p.end
}

// String returns a string representation.
func (p *ParamDecl) String() string {
	if p.Name != "" {
		return fmt.Sprintf("%s %s", p.Type.String(), p.Name)
	}
	return p.Type.String()
}

// StructDecl represents a struct/union declaration.
type StructDecl struct {
	// pos is the starting position in source code.
	pos lexer.Position
	// end is the ending position in source code.
	end lexer.Position
	// Name is the struct/union tag name (may be empty).
	Name string
	// Fields is the list of fields (nil for forward declarations).
	Fields []*FieldDecl
	// IsUnion is true if this is a union.
	IsUnion bool
}

// declNode implements Declaration.
func (s *StructDecl) declNode() {}

// Pos returns the starting position.
func (s *StructDecl) Pos() lexer.Position {
	return s.pos
}

// End returns the ending position.
func (s *StructDecl) End() lexer.Position {
	return s.end
}

// String returns a string representation.
func (s *StructDecl) String() string {
	var sb strings.Builder
	if s.IsUnion {
		sb.WriteString("union")
	} else {
		sb.WriteString("struct")
	}
	if s.Name != "" {
		sb.WriteString(fmt.Sprintf(" %s", s.Name))
	}
	if len(s.Fields) > 0 {
		sb.WriteString(" {")
		for _, field := range s.Fields {
			sb.WriteString(fmt.Sprintf("\n  %s", field.String()))
		}
		sb.WriteString("\n}")
	} else {
		sb.WriteString(";")
	}
	return sb.String()
}

// FieldDecl represents a struct/union field.
type FieldDecl struct {
	// pos is the starting position in source code.
	pos lexer.Position
	// end is the ending position in source code.
	end lexer.Position
	// Type is the field type.
	Type Type
	// Name is the field name (may be empty for anonymous fields).
	Name string
	// BitWidth is the bitfield width (nil if not a bitfield).
	BitWidth Expr
}

// Pos returns the starting position.
func (f *FieldDecl) Pos() lexer.Position {
	return f.pos
}

// End returns the ending position.
func (f *FieldDecl) End() lexer.Position {
	return f.end
}

// String returns a string representation.
func (f *FieldDecl) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("FieldDecl %s: %s", f.Name, f.Type.String()))
	if f.BitWidth != nil {
		sb.WriteString(fmt.Sprintf(" : %s", f.BitWidth.String()))
	}
	return sb.String()
}

// EnumDecl represents an enum declaration.
type EnumDecl struct {
	// pos is the starting position in source code.
	pos lexer.Position
	// end is the ending position in source code.
	end lexer.Position
	// Name is the enum tag name (may be empty).
	Name string
	// Values is the list of enum constants.
	Values []*EnumValue
}

// declNode implements Declaration.
func (e *EnumDecl) declNode() {}

// Pos returns the starting position.
func (e *EnumDecl) Pos() lexer.Position {
	return e.pos
}

// End returns the ending position.
func (e *EnumDecl) End() lexer.Position {
	return e.end
}

// String returns a string representation.
func (e *EnumDecl) String() string {
	var sb strings.Builder
	sb.WriteString("enum")
	if e.Name != "" {
		sb.WriteString(fmt.Sprintf(" %s", e.Name))
	}
	if len(e.Values) > 0 {
		sb.WriteString(" {")
		for i, val := range e.Values {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("\n  %s", val.String()))
		}
		sb.WriteString("\n}")
	} else {
		sb.WriteString(";")
	}
	return sb.String()
}

// EnumValue represents an enum constant.
type EnumValue struct {
	// pos is the starting position in source code.
	pos lexer.Position
	// end is the ending position in source code.
	end lexer.Position
	// Name is the constant name.
	Name string
	// Value is the constant value (nil if not explicitly set).
	Value Expr
}

// Pos returns the starting position.
func (e *EnumValue) Pos() lexer.Position {
	return e.pos
}

// End returns the ending position.
func (e *EnumValue) End() lexer.Position {
	return e.end
}

// String returns a string representation.
func (e *EnumValue) String() string {
	if e.Value != nil {
		return fmt.Sprintf("EnumValue %s = %s", e.Name, e.Value.String())
	}
	return fmt.Sprintf("EnumValue %s", e.Name)
}
