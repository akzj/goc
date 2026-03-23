// Package semantic performs semantic analysis on the AST.
// This file defines symbol table and symbol representations.
package semantic

import (
	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/parser"
)

// TODO: Implement symbol table
// Reference: docs/architecture-design-phases-2-7.md Section 4.3

// SymbolTable manages symbols across all scopes.
type SymbolTable struct {
	// globalScope is the global scope.
	globalScope *Scope
	// currentScope is the current scope.
	currentScope *Scope
	// scopes is the scope stack.
	scopes []*Scope
}

// NewSymbolTable creates a new symbol table.
func NewSymbolTable() *SymbolTable {
	// TODO: Implement
	return nil
}

// Scope represents a lexical scope.
type Scope struct {
	// name is the scope name (for debugging).
	name string
	// symbols maps names to symbols.
	symbols map[string]*Symbol
	// parent is the enclosing scope.
	parent *Scope
	// children are nested scopes.
	children []*Scope
	// level is the nesting level.
	level int
}

// Symbol represents a declared symbol.
type Symbol struct {
	// Name is the symbol name.
	Name string
	// Kind is the symbol kind.
	Kind SymbolKind
	// Type is the symbol type.
	Type parser.Type
	// Position is the declaration position.
	Position lexer.Position
	// Flags contains symbol flags (const, static, etc.).
	Flags SymbolFlags
}

// SymbolKind represents the kind of symbol.
type SymbolKind int

const (
	// SymbolFunction represents a function.
	SymbolFunction SymbolKind = iota
	// SymbolVariable represents a variable.
	SymbolVariable
	// SymbolParameter represents a function parameter.
	SymbolParameter
	// SymbolTypedef represents a typedef.
	SymbolTypedef
	// SymbolStruct represents a struct type.
	SymbolStruct
	// SymbolUnion represents a union type.
	SymbolUnion
	// SymbolEnum represents an enum type.
	SymbolEnum
	// SymbolEnumConstant represents an enum constant.
	SymbolEnumConstant
	// SymbolLabel represents a goto label.
	SymbolLabel
)

// SymbolFlags represents symbol modifiers.
type SymbolFlags int

const (
	// FlagNone indicates no flags.
	FlagNone SymbolFlags = 0
	// FlagConst indicates const.
	FlagConst SymbolFlags = 1 << iota
	// FlagVolatile indicates volatile.
	FlagVolatile
	// FlagStatic indicates static.
	FlagStatic
	// FlagExtern indicates extern.
	FlagExtern
	// FlagInline indicates inline.
	FlagInline
	// FlagThreadLocal indicates _Thread_local.
	FlagThreadLocal
)