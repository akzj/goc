// Package linker links object files into an ELF64 executable.
// This file defines the linker interface.
package linker

import (
	"github.com/akzj/goc/internal/errhand"
)

// TODO: Implement linker
// Reference: docs/architecture-design-phases-2-7.md Section 7.3

// Linker links object files into an executable.
type Linker struct {
	// errors is the error handler.
	errs *errhand.ErrorHandler
	// symbols is the symbol table.
	symbols map[string]*Symbol
	// sections is the list of sections.
	sections []*Section
	// entry is the entry point symbol.
	entry string
}

// NewLinker creates a new linker.
func NewLinker(errorHandler *errhand.ErrorHandler) *Linker {
	// TODO: Implement
	return nil
}

// Link links the given object files and libraries.
func (l *Linker) Link(objects []ObjectFile, libs []string) ([]byte, error) {
	// TODO: Implement
	return nil, nil
}

// ResolveSymbols resolves undefined symbols.
func (l *Linker) ResolveSymbols() error {
	// TODO: Implement
	return nil
}

// Relocate performs relocations.
func (l *Linker) Relocate() error {
	// TODO: Implement
	return nil
}

// Emit emits the final ELF binary.
func (l *Linker) Emit() ([]byte, error) {
	// TODO: Implement
	return nil, nil
}

// ObjectFile represents an object file.
type ObjectFile struct {
	// Name is the file name.
	Name string
	// Sections is the list of sections.
	Sections []*Section
	// Symbols is the list of symbols.
	Symbols []*Symbol
	// Relocations is the list of relocations.
	Relocations []*Relocation
}

// Relocation represents a relocation entry.
type Relocation struct {
	// Offset is the offset within the section.
	Offset uint64
	// Type is the relocation type.
	Type uint32
	// Symbol is the symbol to relocate against.
	Symbol *Symbol
	// Addend is the addend.
	Addend int64
}