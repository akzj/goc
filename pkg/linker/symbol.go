// Package linker links object files into an ELF64 executable.
// This file defines linker symbols and symbol table operations.
package linker

import (
	"fmt"
)

// Symbol represents a symbol in the linker.
// Reference: docs/architecture-design-phases-2-7.md Section 7.4
type Symbol struct {
	// Name is the symbol name.
	Name string
	// Value is the symbol value (address).
	Value uint64
	// Size is the symbol size.
	Size uint64
	// Section is the section containing the symbol.
	Section *Section
	// Binding is the symbol binding (STB_LOCAL, STB_GLOBAL, STB_WEAK).
	Binding SymbolBinding
	// Type is the symbol type (STT_NOTYPE, STT_OBJECT, STT_FUNC, etc.).
	Type SymbolType
	// Defined is true if the symbol is defined.
	Defined bool
}

// SymbolBinding represents symbol binding.
// ELF64 symbol binding specifies the visibility and linkage of symbols.
// See STB_LOCAL, STB_GLOBAL, STB_WEAK constants in elf.go.
type SymbolBinding int

// String returns the string representation of the symbol binding.
func (b SymbolBinding) String() string {
	switch b {
	case STB_LOCAL:
		return "LOCAL"
	case STB_GLOBAL:
		return "GLOBAL"
	case STB_WEAK:
		return "WEAK"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", b)
	}
}

// SymbolType represents symbol type.
// ELF64 symbol type specifies the kind of entity the symbol represents.
// See STT_NOTYPE, STT_OBJECT, STT_FUNC, etc. constants in elf.go.
type SymbolType int

// String returns the string representation of the symbol type.
func (t SymbolType) String() string {
	switch t {
	case STT_NOTYPE:
		return "NOTYPE"
	case STT_OBJECT:
		return "OBJECT"
	case STT_FUNC:
		return "FUNC"
	case STT_SECTION:
		return "SECTION"
	case STT_FILE:
		return "FILE"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}

// Section represents a section in the linker.
// Sections contain the actual code and data of the program.
type Section struct {
	// Name is the section name.
	Name string
	// Type is the section type.
	Type SectionType
	// Flags is the section flags.
	Flags SectionFlags
	// Data is the section data.
	Data []byte
	// Addr is the section address.
	Addr uint64
}

// NewSymbol creates a new symbol with the given parameters.
// This is a helper function for creating symbols with common defaults.
func NewSymbol(name string, binding SymbolBinding, typ SymbolType) *Symbol {
	return &Symbol{
		Name:    name,
		Binding: binding,
		Type:    typ,
		Defined: false,
		Value:   0,
		Size:    0,
		Section: nil,
	}
}

// NewDefinedSymbol creates a new defined symbol with the given section and value.
func NewDefinedSymbol(name string, binding SymbolBinding, typ SymbolType, section *Section, value uint64, size uint64) *Symbol {
	return &Symbol{
		Name:    name,
		Binding: binding,
		Type:    typ,
		Defined: true,
		Value:   value,
		Size:    size,
		Section: section,
	}
}

// NewFunctionSymbol creates a new function symbol.
func NewFunctionSymbol(name string, binding SymbolBinding, section *Section, value uint64, size uint64) *Symbol {
	return NewDefinedSymbol(name, binding, STT_FUNC, section, value, size)
}

// NewObjectSymbol creates a new object (data) symbol.
func NewObjectSymbol(name string, binding SymbolBinding, section *Section, value uint64, size uint64) *Symbol {
	return NewDefinedSymbol(name, binding, STT_OBJECT, section, value, size)
}

// NewSectionSymbol creates a new section symbol.
func NewSectionSymbol(name string, section *Section) *Symbol {
	return NewDefinedSymbol(name, STB_LOCAL, STT_SECTION, section, section.Addr, 0)
}

// NewFileSymbol creates a new file symbol (source file name).
func NewFileSymbol(name string) *Symbol {
	return &Symbol{
		Name:    name,
		Binding: STB_LOCAL,
		Type:    STT_FILE,
		Defined: true,
	}
}

// IsLocal returns true if the symbol has local binding.
func (s *Symbol) IsLocal() bool {
	return s.Binding == STB_LOCAL
}

// IsGlobal returns true if the symbol has global binding.
func (s *Symbol) IsGlobal() bool {
	return s.Binding == STB_GLOBAL
}

// IsWeak returns true if the symbol has weak binding.
func (s *Symbol) IsWeak() bool {
	return s.Binding == STB_WEAK
}

// IsFunction returns true if the symbol is a function.
func (s *Symbol) IsFunction() bool {
	return s.Type == STT_FUNC
}

// IsObject returns true if the symbol is an object (data).
func (s *Symbol) IsObject() bool {
	return s.Type == STT_OBJECT
}

// IsDefined returns true if the symbol is defined.
func (s *Symbol) IsDefined() bool {
	return s.Defined
}

// IsUndefined returns true if the symbol is undefined.
func (s *Symbol) IsUndefined() bool {
	return !s.Defined
}

// IsExported returns true if the symbol should be exported (visible to other object files).
func (s *Symbol) IsExported() bool {
	return s.IsGlobal() || s.IsWeak()
}

// SetDefined marks the symbol as defined with the given section and value.
func (s *Symbol) SetDefined(section *Section, value uint64, size uint64) {
	s.Defined = true
	s.Section = section
	s.Value = value
	s.Size = size
}

// SetSection sets the section for the symbol.
func (s *Symbol) SetSection(section *Section) {
	s.Section = section
}

// SetValue sets the value (address) of the symbol.
func (s *Symbol) SetValue(value uint64) {
	s.Value = value
}

// SetSize sets the size of the symbol.
func (s *Symbol) SetSize(size uint64) {
	s.Size = size
}

// String returns a string representation of the symbol.
func (s *Symbol) String() string {
	defined := "undefined"
	if s.Defined {
		defined = "defined"
	}
	sectionName := "none"
	if s.Section != nil {
		sectionName = s.Section.Name
	}
	return fmt.Sprintf("Symbol{%s, %s, %s, %s, section=%s, addr=0x%x, size=%d}",
		s.Name, s.Binding.String(), s.Type.String(), defined, sectionName, s.Value, s.Size)
}

// NewSection creates a new section with the given name and type.
func NewSection(name string, typ SectionType, flags SectionFlags) *Section {
	return &Section{
		Name:  name,
		Type:  typ,
		Flags: flags,
		Data:  make([]byte, 0),
		Addr:  0,
	}
}

// NewCodeSection creates a new code section (.text).
func NewCodeSection() *Section {
	return NewSection(".text", SHT_PROGBITS, SHF_ALLOC|SHF_EXECINSTR)
}

// NewDataSection creates a new data section (.data).
func NewDataSection() *Section {
	return NewSection(".data", SHT_PROGBITS, SHF_ALLOC|SHF_WRITE)
}

// NewReadOnlySection creates a new read-only data section (.rodata).
func NewReadOnlySection() *Section {
	return NewSection(".rodata", SHT_PROGBITS, SHF_ALLOC)
}

// NewBSSSection creates a new BSS section (.bss).
func NewBSSSection() *Section {
	return NewSection(".bss", SHT_NOBITS, SHF_ALLOC|SHF_WRITE)
}

// NewSymTabSection creates a new symbol table section (.symtab).
func NewSymTabSection() *Section {
	return NewSection(".symtab", SHT_SYMTAB, 0)
}

// NewStrTabSection creates a new string table section (.strtab).
func NewStrTabSection() *Section {
	return NewSection(".strtab", SHT_STRTAB, 0)
}

// NewShStrTabSection creates a new section header string table section (.shstrtab).
func NewShStrTabSection() *Section {
	return NewSection(".shstrtab", SHT_STRTAB, 0)
}

// NewRelaSection creates a new relocation with addends section (.rela.*).
func NewRelaSection(name string) *Section {
	return NewSection(name, SHT_RELA, 0)
}

// NewRelSection creates a new relocation section (.rel.*).
func NewRelSection(name string) *Section {
	return NewSection(name, SHT_REL, 0)
}

// AddData appends data to the section.
func (s *Section) AddData(data []byte) {
	s.Data = append(s.Data, data...)
}

// AddByte appends a single byte to the section.
func (s *Section) AddByte(b byte) {
	s.Data = append(s.Data, b)
}

// AddBytes appends multiple bytes to the section.
func (s *Section) AddBytes(bytes ...byte) {
	s.Data = append(s.Data, bytes...)
}

// AddUint16 appends a uint16 in little-endian format.
func (s *Section) AddUint16(v uint16) {
	s.Data = append(s.Data, byte(v), byte(v>>8))
}

// AddUint32 appends a uint32 in little-endian format.
func (s *Section) AddUint32(v uint32) {
	s.Data = append(s.Data, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

// AddUint64 appends a uint64 in little-endian format.
func (s *Section) AddUint64(v uint64) {
	s.Data = append(s.Data,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}

// AddInt32 appends an int32 in little-endian format.
func (s *Section) AddInt32(v int32) {
	s.AddUint32(uint32(v))
}

// AddInt64 appends an int64 in little-endian format.
func (s *Section) AddInt64(v int64) {
	s.AddUint64(uint64(v))
}

// Size returns the size of the section data.
func (s *Section) Size() uint64 {
	return uint64(len(s.Data))
}

// IsAllocatable returns true if the section should be loaded into memory.
func (s *Section) IsAllocatable() bool {
	return s.Flags&SHF_ALLOC != 0
}

// IsWritable returns true if the section is writable.
func (s *Section) IsWritable() bool {
	return s.Flags&SHF_WRITE != 0
}

// IsExecutable returns true if the section is executable.
func (s *Section) IsExecutable() bool {
	return s.Flags&SHF_EXECINSTR != 0
}

// SetAddr sets the address of the section.
func (s *Section) SetAddr(addr uint64) {
	s.Addr = addr
}

// String returns a string representation of the section.
func (s *Section) String() string {
	return fmt.Sprintf("Section{%s, type=%d, flags=0x%x, addr=0x%x, size=%d}",
		s.Name, s.Type, s.Flags, s.Addr, s.Size())
}

// SymbolTable represents a table of symbols for the linker.
type SymbolTable struct {
	// symbols is the map of symbol name to symbol.
	symbols map[string]*Symbol
	// localSymbols is the list of local symbols.
	localSymbols []*Symbol
	// globalSymbols is the list of global symbols.
	globalSymbols []*Symbol
	// undefinedSymbols is the list of undefined symbols.
	undefinedSymbols []*Symbol
}

// NewSymbolTable creates a new symbol table.
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		symbols:          make(map[string]*Symbol),
		localSymbols:     make([]*Symbol, 0),
		globalSymbols:    make([]*Symbol, 0),
		undefinedSymbols: make([]*Symbol, 0),
	}
}

// Add adds a symbol to the symbol table.
// Returns an error if a symbol with the same name already exists.
func (st *SymbolTable) Add(symbol *Symbol) error {
	if _, exists := st.symbols[symbol.Name]; exists {
		return fmt.Errorf("duplicate symbol: %s", symbol.Name)
	}

	st.symbols[symbol.Name] = symbol

	if symbol.IsLocal() {
		st.localSymbols = append(st.localSymbols, symbol)
	} else {
		st.globalSymbols = append(st.globalSymbols, symbol)
	}

	if symbol.IsUndefined() {
		st.undefinedSymbols = append(st.undefinedSymbols, symbol)
	}

	return nil
}

// Get retrieves a symbol by name.
// Returns nil if the symbol is not found.
func (st *SymbolTable) Get(name string) *Symbol {
	return st.symbols[name]
}

// Lookup looks up a symbol by name in the symbol table.
// Returns the symbol and true if found, nil and false otherwise.
func (st *SymbolTable) Lookup(name string) (*Symbol, bool) {
	symbol, ok := st.symbols[name]
	return symbol, ok
}

// Remove removes a symbol from the symbol table.
// Returns true if the symbol was found and removed.
func (st *SymbolTable) Remove(name string) bool {
	symbol, ok := st.symbols[name]
	if !ok {
		return false
	}

	delete(st.symbols, name)

	// Remove from appropriate list
	if symbol.IsLocal() {
		st.localSymbols = removeSymbolFromList(st.localSymbols, symbol)
	} else {
		st.globalSymbols = removeSymbolFromList(st.globalSymbols, symbol)
	}

	if symbol.IsUndefined() {
		st.undefinedSymbols = removeSymbolFromList(st.undefinedSymbols, symbol)
	}

	return true
}

// removeSymbolFromList removes a symbol from a list of symbols.
func removeSymbolFromList(list []*Symbol, target *Symbol) []*Symbol {
	result := make([]*Symbol, 0, len(list))
	for _, s := range list {
		if s != target {
			result = append(result, s)
		}
	}
	return result
}

// GetAll returns all symbols in the table.
func (st *SymbolTable) GetAll() []*Symbol {
	result := make([]*Symbol, 0, len(st.symbols))
	for _, symbol := range st.symbols {
		result = append(result, symbol)
	}
	return result
}

// GetLocal returns all local symbols.
func (st *SymbolTable) GetLocal() []*Symbol {
	return st.localSymbols
}

// GetGlobal returns all global symbols.
func (st *SymbolTable) GetGlobal() []*Symbol {
	return st.globalSymbols
}

// GetUndefined returns all undefined symbols.
func (st *SymbolTable) GetUndefined() []*Symbol {
	return st.undefinedSymbols
}

// GetDefined returns all defined symbols.
func (st *SymbolTable) GetDefined() []*Symbol {
	var result []*Symbol
	for _, symbol := range st.symbols {
		if symbol.IsDefined() {
			result = append(result, symbol)
		}
	}
	return result
}

// GetFunctions returns all function symbols.
func (st *SymbolTable) GetFunctions() []*Symbol {
	var result []*Symbol
	for _, symbol := range st.symbols {
		if symbol.IsFunction() {
			result = append(result, symbol)
		}
	}
	return result
}

// GetObjects returns all object (data) symbols.
func (st *SymbolTable) GetObjects() []*Symbol {
	var result []*Symbol
	for _, symbol := range st.symbols {
		if symbol.IsObject() {
			result = append(result, symbol)
		}
	}
	return result
}

// Count returns the number of symbols in the table.
func (st *SymbolTable) Count() int {
	return len(st.symbols)
}

// HasUndefined returns true if there are any undefined symbols.
func (st *SymbolTable) HasUndefined() bool {
	return len(st.undefinedSymbols) > 0
}

// MarkDefined marks a symbol as defined with the given section and value.
// Returns an error if the symbol is not found.
func (st *SymbolTable) MarkDefined(name string, section *Section, value uint64, size uint64) error {
	symbol, ok := st.symbols[name]
	if !ok {
		return fmt.Errorf("symbol not found: %s", name)
	}

	// Remove from undefined list if present
	if symbol.IsUndefined() {
		st.undefinedSymbols = removeSymbolFromList(st.undefinedSymbols, symbol)
	}

	symbol.SetDefined(section, value, size)
	return nil
}

// String returns a string representation of the symbol table.
func (st *SymbolTable) String() string {
	return fmt.Sprintf("SymbolTable{count=%d, local=%d, global=%d, undefined=%d}",
		st.Count(), len(st.localSymbols), len(st.globalSymbols), len(st.undefinedSymbols))
}