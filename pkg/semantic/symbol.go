// Package semantic performs semantic analysis on the AST.
// This file defines symbol table and symbol representations.
package semantic

import (
	"fmt"

	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/parser"
)

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
	globalScope := &Scope{
		name:     "global",
		symbols:  make(map[string]*Symbol),
		parent:   nil,
		children: nil,
		level:    0,
	}
	return &SymbolTable{
		globalScope:  globalScope,
		currentScope: globalScope,
		scopes:       []*Scope{globalScope},
	}
}

// PushScope creates and enters a new scope.
func (st *SymbolTable) PushScope(name string) *Scope {
	parentLevel := 0
	if st.currentScope != nil {
		parentLevel = st.currentScope.level
	}
	newScope := &Scope{
		name:     name,
		symbols:  make(map[string]*Symbol),
		parent:   st.currentScope,
		children: nil,
		level:    parentLevel + 1,
	}
	if st.currentScope != nil {
		st.currentScope.children = append(st.currentScope.children, newScope)
	}
	st.currentScope = newScope
	st.scopes = append(st.scopes, newScope)
	return newScope
}

// PopScope exits the current scope and returns to the parent.
func (st *SymbolTable) PopScope() *Scope {
	if st.currentScope == nil || st.currentScope.parent == nil {
		// Can't pop if already at global or nil
		if st.currentScope == nil && len(st.scopes) > 0 {
			// Restore from scopes stack
			st.currentScope = st.scopes[len(st.scopes)-1]
		}
		if len(st.scopes) > 1 {
			st.scopes = st.scopes[:len(st.scopes)-1]
			st.currentScope = st.scopes[len(st.scopes)-1]
			return st.currentScope
		}
		return st.currentScope
	}
	popped := st.currentScope
	st.currentScope = st.currentScope.parent
	if len(st.scopes) > 1 {
		st.scopes = st.scopes[:len(st.scopes)-1]
	}
	return popped
}

// Lookup looks up a symbol in the current scope chain.
func (st *SymbolTable) Lookup(name string) *Symbol {
	scope := st.currentScope
	for scope != nil {
		if symbol, ok := scope.symbols[name]; ok {
			return symbol
		}
		scope = scope.parent
	}
	return nil
}

// Declare declares a symbol in the current scope.
func (st *SymbolTable) Declare(symbol *Symbol) error {
	if st.currentScope == nil {
		return fmt.Errorf("no current scope")
	}
	if _, exists := st.currentScope.symbols[symbol.Name]; exists {
		return fmt.Errorf("symbol '%s' already declared", symbol.Name)
	}
	st.currentScope.symbols[symbol.Name] = symbol
	return nil
}

// GetCurrentScope returns the current scope.
func (st *SymbolTable) GetCurrentScope() *Scope {
	return st.currentScope
}

// GetGlobalScope returns the global scope.
func (st *SymbolTable) GetGlobalScope() *Scope {
	return st.globalScope
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

// GetName returns the scope name.
func (s *Scope) GetName() string {
	return s.name
}

// GetLevel returns the scope nesting level.
func (s *Scope) GetLevel() int {
	return s.level
}

// GetParent returns the parent scope.
func (s *Scope) GetParent() *Scope {
	return s.parent
}

// GetChildren returns the child scopes.
func (s *Scope) GetChildren() []*Scope {
	return s.children
}

// GetSymbols returns all symbols in this scope.
func (s *Scope) GetSymbols() map[string]*Symbol {
	return s.symbols
}

// Lookup looks up a symbol in this scope only (not parent scopes).
func (s *Scope) Lookup(name string) *Symbol {
	return s.symbols[name]
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

// String returns a string representation of the symbol.
func (s *Symbol) String() string {
	return fmt.Sprintf("Symbol{name=%s, kind=%v, type=%s}", s.Name, s.Kind, s.Type)
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

// String returns a string representation of the symbol kind.
func (k SymbolKind) String() string {
	switch k {
	case SymbolFunction:
		return "function"
	case SymbolVariable:
		return "variable"
	case SymbolParameter:
		return "parameter"
	case SymbolTypedef:
		return "typedef"
	case SymbolStruct:
		return "struct"
	case SymbolUnion:
		return "union"
	case SymbolEnum:
		return "enum"
	case SymbolEnumConstant:
		return "enum constant"
	case SymbolLabel:
		return "label"
	default:
		return "unknown"
	}
}

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

// HasFlag checks if a flag is set.
func (f SymbolFlags) HasFlag(flag SymbolFlags) bool {
	return (f & flag) != 0
}

// AddFlag adds a flag.
func (f *SymbolFlags) AddFlag(flag SymbolFlags) {
	*f |= flag
}

// RemoveFlag removes a flag.
func (f *SymbolFlags) RemoveFlag(flag SymbolFlags) {
	*f &= ^flag
}