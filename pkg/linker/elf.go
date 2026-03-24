// Package linker links object files into an ELF64 executable.
// This file defines ELF64 structures and helper methods.
package linker

import (
	"encoding/binary"
	"io"
)

// ============================================================================
// ELF64 Header Constants
// ============================================================================

// ELF magic number
const (
	ELFMAG0    = 0x7f
	ELFMAG1    = 'E'
	ELFMAG2    = 'L'
	ELFMAG3    = 'F'
	ELFMAG     = "\x7fELF"
	SELFMAG    = 16
)

// ELF class (32-bit or 64-bit)
const (
	ELFCLASSNONE = 0
	ELFCLASS32   = 1
	ELFCLASS64   = 2
)

// ELF data encoding (little-endian or big-endian)
const (
	ELFDATANONE = 0
	ELFDATA2LSB = 1 // Little-endian
	ELFDATA2MSB = 2 // Big-endian
)

// ELF version
const (
	EV_NONE    = 0
	EV_CURRENT = 1
)

// ELF file types
const (
	ET_NONE = 0 // No file type
	ET_REL  = 1 // Relocatable file
	ET_EXEC = 2 // Executable file
	ET_DYN  = 3 // Shared object file
	ET_CORE = 4 // Core file
)

// ELF machine types
const (
	EM_NONE        = 0  // No machine
	EM_386         = 3  // Intel 80386
	EM_X86_64      = 62 // AMD x86-64
	EM_AARCH64     = 183// ARM 64-bit
	EM_RISCV       = 243// RISC-V
)

// ELF version (program header)
const (
	PV_NONE    = 0
	PV_CURRENT = 1
)

// ============================================================================
// Program Header Constants
// ============================================================================

// Program header types
const (
	PT_NULL    ProgramType = 0  // Unused entry
	PT_LOAD    ProgramType = 1  // Loadable segment
	PT_DYNAMIC ProgramType = 2  // Dynamic linking info
	PT_INTERP  ProgramType = 3  // Program interpreter
	PT_NOTE    ProgramType = 4  // Auxiliary info
	PT_SHLIB   ProgramType = 5  // Reserved
	PT_PHDR    ProgramType = 6  // Program header table
	PT_TLS     ProgramType = 7  // Thread-local storage
	PT_GNU_EH_FRAME ProgramType = 0x6474e550 // GCC exception handling
	PT_GNU_STACK    ProgramType = 0x6474e551 // Stack permissions
	PT_GNU_RELRO    ProgramType = 0x6474e552 // Read-only after relocation
)

// Program header flags
const (
	PF_X ProgramFlags = 0x1 // Execute
	PF_W ProgramFlags = 0x2 // Write
	PF_R ProgramFlags = 0x4 // Read
)

// ============================================================================
// Section Header Constants
// ============================================================================

// Section types
const (
	SHT_NULL     SectionType = 0  // Inactive
	SHT_PROGBITS SectionType = 1  // Program data
	SHT_SYMTAB   SectionType = 2  // Symbol table
	SHT_STRTAB   SectionType = 3  // String table
	SHT_RELA     SectionType = 4  // Relocation with addend
	SHT_HASH     SectionType = 5  // Symbol hash table
	SHT_DYNAMIC  SectionType = 6  // Dynamic linking info
	SHT_NOTE     SectionType = 7  // Auxiliary info
	SHT_NOBITS   SectionType = 8  // Uninitialized data (BSS)
	SHT_REL      SectionType = 9  // Relocation without addend
	SHT_SHLIB    SectionType = 10 // Reserved
	SHT_DYNSYM   SectionType = 11 // Dynamic symbol table
	SHT_INIT_ARRAY SectionType = 14 // Initialization function pointers
	SHT_FINI_ARRAY SectionType = 15 // Termination function pointers
	SHT_GNU_HASH SectionType = 0x6ffffff6 // GNU hash table
)

// Section flags
const (
	SHF_WRITE     SectionFlags = 0x1        // Writable
	SHF_ALLOC     SectionFlags = 0x2        // Occupies memory
	SHF_EXECINSTR SectionFlags = 0x4        // Executable
	SHF_MERGE     SectionFlags = 0x10       // Mergeable
	SHF_STRINGS   SectionFlags = 0x20       // Strings
	SHF_INFO_LINK SectionFlags = 0x40       // Info field is section index
	SHF_LINK_ORDER SectionFlags = 0x80      // Preserve link order
	SHF_OS_NONCONFORMING SectionFlags = 0x100 // OS-specific
	SHF_GROUP     SectionFlags = 0x200      // Section is a group
	SHF_TLS       SectionFlags = 0x400      // Thread-local storage
	SHF_COMPRESSED SectionFlags = 0x800     // Compressed section
)

// Special section indices
const (
	SHN_UNDEF     = 0      // Undefined section
	SHN_LORESERVE = 0xff00 // Low reserved
	SHN_LOPROC    = 0xff00 // Low processor-specific
	SHN_HIPROC    = 0xff1f // High processor-specific
	SHN_LOOS      = 0xff20 // Low OS-specific
	SHN_HIOS      = 0xff3f // High OS-specific
	SHN_ABS       = 0xfff1 // Absolute section
	SHN_COMMON    = 0xfff2 // Common section
	SHN_XINDEX    = 0xffff // Extended section indices
	SHN_HIRESERVE = 0xffff // High reserved
)

// ============================================================================
// Symbol Binding and Type Constants (ELF Standard Values)
// These are used by symbol.go for SymbolBinding and SymbolType
// ============================================================================

// Symbol binding constants (ELF standard values)
const (
	STB_LOCAL  = 0  // Local symbol
	STB_GLOBAL = 1  // Global symbol
	STB_WEAK   = 2  // Weak symbol
	STB_LOOS   = 10 // Low OS-specific
	STB_HIOS   = 12 // High OS-specific
	STB_LOPROC = 13 // Low processor-specific
	STB_HIPROC = 15 // High processor-specific
)

// Symbol type constants (ELF standard values)
const (
	STT_NOTYPE  = 0  // No type
	STT_OBJECT  = 1  // Data object
	STT_FUNC    = 2  // Function
	STT_SECTION = 3  // Section
	STT_FILE    = 4  // File
	STT_COMMON  = 5  // Common object
	STT_TLS     = 6  // Thread-local storage
	STT_LOOS    = 10 // Low OS-specific
	STT_HIOS    = 12 // High OS-specific
	STT_LOPROC  = 13 // Low processor-specific
	STT_HIPROC  = 15 // High processor-specific
)

// Symbol visibility
const (
	STV_DEFAULT   = 0 // Default visibility
	STV_INTERNAL  = 1 // Internal visibility
	STV_HIDDEN    = 2 // Hidden visibility
	STV_PROTECTED = 3 // Protected visibility
)

// ELF64_ST_INFO creates st_info from binding and type.
func ELF64_ST_INFO(bind, typ int) byte {
	return byte((bind << 4) | (typ & 0xf))
}

// ELF64_ST_BIND extracts binding from st_info.
func ELF64_ST_BIND(info byte) byte {
	return info >> 4
}

// ELF64_ST_TYPE extracts type from st_info.
func ELF64_ST_TYPE(info byte) byte {
	return info & 0xf
}

// ============================================================================
// Relocation Constants (x86-64)
// ============================================================================

// x86-64 relocation types (from System V AMD64 ABI)
const (
	R_X86_64_NONE        = 0  // No relocation
	R_X86_64_64          = 1  // Direct 64-bit
	R_X86_64_PC32        = 2  // PC-relative 32-bit
	R_X86_64_32          = 3  // Direct 32-bit
	R_X86_64_32S         = 4  // Direct 32-bit sign-extended
	R_X86_64_16          = 5  // Direct 16-bit
	R_X86_64_PC16        = 6  // PC-relative 16-bit
	R_X86_64_8           = 7  // Direct 8-bit
	R_X86_64_PC8         = 8  // PC-relative 8-bit
	R_X86_64_PLT32       = 42 // PC-relative 32-bit (PLT)
	R_X86_64_GOT32       = 9  // 32-bit GOT entry
	R_X86_64_GOT64       = 24 // 64-bit GOT entry
	R_X86_64_GOTPCREL    = 9  // PC-relative GOT entry
	R_X86_64_REX_GOTPCRELX = 42 // GOTPCRELX with REX prefix
	R_X86_64_GOTTPOFF    = 29 // GOT-relative thread offset
	R_X86_64_TPOFF32     = 26 // Thread offset 32-bit
	R_X86_64_TLSGD       = 25 // TLS general dynamic
	R_X86_64_TLSLD       = 26 // TLS local dynamic
	R_X86_64_DTPOFF32    = 27 // DTP-relative offset 32-bit
	R_X86_64_DTPOFF64    = 28 // DTP-relative offset 64-bit
	R_X86_64_GOTPC32     = 32 // PC-relative GOT base
	R_X86_64_GOTPCRELX   = 42 // GOTPCRELX
)

// ELF64_R_INFO creates relocation info from symbol and type.
func ELF64_R_INFO(sym, typ uint32) uint64 {
	return (uint64(sym) << 32) | uint64(typ&0xffffffff)
}

// ELF64_R_SYM extracts symbol index from relocation info.
func ELF64_R_SYM(info uint64) uint32 {
	return uint32(info >> 32)
}

// ELF64_R_TYPE extracts type from relocation info.
func ELF64_R_TYPE(info uint64) uint32 {
	return uint32(info & 0xffffffff)
}

// ============================================================================
// Dynamic Section Constants
// ============================================================================

// Dynamic tags
const (
	DT_NULL     = 0  // End of dynamic section
	DT_NEEDED   = 1  // Shared library needed
	DT_PLTRELSZ = 2  // Size of PLT relocation entries
	DT_PLTGOT   = 3  // PLT/GOT address
	DT_HASH     = 4  // Symbol hash table address
	DT_STRTAB   = 5  // String table address
	DT_SYMTAB   = 6  // Symbol table address
	DT_RELA     = 7  // Relocation table address
	DT_RELASZ   = 8  // Relocation table size
	DT_RELAENT  = 9  // Relocation entry size
	DT_STRSZ    = 10 // String table size
	DT_SYMENT   = 11 // Symbol table entry size
	DT_INIT     = 12 // Initialization function
	DT_FINI     = 13 // Termination function
	DT_SONAME   = 14 // Shared object name
	DT_RPATH    = 15 // Runtime library search path
	DT_SYMBOLIC = 16 // Symbolic linking
	DT_REL      = 17 // Relocation table (without addend)
	DT_RELSZ    = 18 // Relocation table size
	DT_RELENT   = 19 // Relocation entry size
	DT_PLTREL   = 20 // PLT relocation type
	DT_DEBUG    = 21 // Debug entry
	DT_TEXTREL  = 22 // Text relocations exist
	DT_JMPREL   = 23 // PLT relocation table
	DT_BIND_NOW = 24 // Bind now
	DT_INIT_ARRAY = 25 // Initialization function array
	DT_FINI_ARRAY = 26 // Termination function array
	DT_INIT_ARRAYSZ = 27 // Init array size
	DT_FINI_ARRAYSZ = 28 // Fini array size
)

// ============================================================================
// ELF64 Structures
// ============================================================================

// ELFHeader represents the ELF64 header (64 bytes).
type ELFHeader struct {
	Ident      [16]byte // ELF identification
	Type       uint16   // Object file type
	Machine    uint16   // Machine type
	Version    uint32   // Object file version
	Entry      uint64   // Entry point address
	Phoff      uint64   // Program header table offset
	Shoff      uint64   // Section header table offset
	Flags      uint32   // Processor-specific flags
	Ehsize     uint16   // ELF header size
	Phentsize  uint16   // Program header entry size
	Phnum      uint16   // Number of program header entries
	Shentsize  uint16   // Section header entry size
	Shnum      uint16   // Number of section header entries
	Shstrndx   uint16   // Section name string table index
}

// ProgramHeader represents an ELF64 program header (56 bytes).
type ProgramHeader struct {
	Type   ProgramType // Segment type
	Flags  ProgramFlags // Segment flags
	Offset uint64      // Offset in file
	Vaddr  uint64      // Virtual address in memory
	Paddr  uint64      // Physical address (unused)
	Filesz uint64      // Size in file
	Memsz  uint64      // Size in memory
	Align  uint64      // Alignment
}

// ProgramFlags represents program header flags.
type ProgramFlags uint32

// SectionHeader represents an ELF64 section header (64 bytes).
type SectionHeader struct {
	Name      uint32       // Section name (string table index)
	Type      SectionType  // Section type
	Flags     SectionFlags // Section flags
	Addr      uint64       // Virtual address
	Offset    uint64       // Offset in file
	Size      uint64       // Section size
	Link      uint32       // Link to another section
	Info      uint32       // Additional section info
	Addralign uint64       // Address alignment
	Entsize   uint64       // Entry size (for tables)
}

// SectionType represents a section type.
type SectionType uint32

// SectionFlags represents section flags.
type SectionFlags uint64

// ProgramType represents a program header type.
type ProgramType uint32

// ============================================================================
// ELF64 Symbol Table Entry
// ============================================================================

// Symbol64 represents an ELF64 symbol table entry (24 bytes).
type Symbol64 struct {
	Name  uint32 // Symbol name (string table index)
	Info  byte   // Symbol binding and type
	Other byte   // Symbol visibility
	Shndx uint16 // Section index
	Value uint64 // Symbol value
	Size  uint64 // Symbol size
}

// NewSymbol64 creates a new ELF64 symbol entry.
func NewSymbol64(name uint32, bind, typ int, shndx uint16, value, size uint64) Symbol64 {
	return Symbol64{
		Name:  name,
		Info:  ELF64_ST_INFO(bind, typ),
		Other: 0,
		Shndx: shndx,
		Value: value,
		Size:  size,
	}
}

// ============================================================================
// ELF64 Relocation Entries
// ============================================================================

// Rel64 represents an ELF64 relocation entry without addend (16 bytes).
type Rel64 struct {
	Offset uint64 // Offset in section
	Info   uint64 // Symbol index and type
}

// Rela64 represents an ELF64 relocation entry with addend (24 bytes).
type Rela64 struct {
	Offset uint64 // Offset in section
	Info   uint64 // Symbol index and type
	Addend int64  // Addend
}

// NewRela64 creates a new ELF64 relocation entry with addend.
func NewRela64(offset uint64, sym, typ uint32, addend int64) Rela64 {
	return Rela64{
		Offset: offset,
		Info:   ELF64_R_INFO(sym, typ),
		Addend: addend,
	}
}

// ============================================================================
// ELF64 Dynamic Section Entry
// ============================================================================

// Dyn64 represents an ELF64 dynamic section entry (16 bytes).
type Dyn64 struct {
	Tag int64  // Dynamic tag
	Val uint64 // Value (interpretation depends on tag)
}

// NewDyn64 creates a new ELF64 dynamic entry.
func NewDyn64(tag int64, val uint64) Dyn64 {
	return Dyn64{
		Tag: tag,
		Val: val,
	}
}

// ============================================================================
// Helper Methods
// ============================================================================

// NewELFHeader creates a new ELF64 header with default values for an executable.
func NewELFHeader() *ELFHeader {
	h := &ELFHeader{}
	// Set magic number
	h.Ident[0] = ELFMAG0
	h.Ident[1] = ELFMAG1
	h.Ident[2] = ELFMAG2
	h.Ident[3] = ELFMAG3
	// Set class (64-bit)
	h.Ident[4] = ELFCLASS64
	// Set data encoding (little-endian)
	h.Ident[5] = ELFDATA2LSB
	// Set version
	h.Ident[6] = EV_CURRENT
	// OS/ABI (System V)
	h.Ident[7] = 0
	// Padding (bytes 8-15 are zero)

	h.Type = ET_EXEC
	h.Machine = EM_X86_64
	h.Version = EV_CURRENT
	h.Ehsize = 64
	h.Phentsize = 56
	h.Shentsize = 64

	return h
}

// NewELFHeaderShared creates a new ELF64 header for a shared object (PIE).
func NewELFHeaderShared() *ELFHeader {
	h := NewELFHeader()
	h.Type = ET_DYN
	return h
}

// NewELFHeaderRelocatable creates a new ELF64 header for a relocatable object.
func NewELFHeaderRelocatable() *ELFHeader {
	h := NewELFHeader()
	h.Type = ET_REL
	return h
}

// WriteTo writes the ELF header to the given writer in binary format.
func (h *ELFHeader) WriteTo(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, h)
}

// ReadFrom reads the ELF header from the given reader in binary format.
func (h *ELFHeader) ReadFrom(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, h)
}

// Validate validates the ELF header.
func (h *ELFHeader) Validate() error {
	// Check magic number
	if h.Ident[0] != ELFMAG0 || h.Ident[1] != ELFMAG1 ||
		h.Ident[2] != ELFMAG2 || h.Ident[3] != ELFMAG3 {
		return &ELFError{"invalid magic number"}
	}
	// Check class
	if h.Ident[4] != ELFCLASS64 {
		return &ELFError{"not a 64-bit ELF"}
	}
	// Check data encoding
	if h.Ident[5] != ELFDATA2LSB {
		return &ELFError{"not little-endian"}
	}
	return nil
}

// NewProgramHeader creates a new program header.
func NewProgramHeader(typ ProgramType, flags ProgramFlags) *ProgramHeader {
	return &ProgramHeader{
		Type:  typ,
		Flags: flags,
		Align: 0x1000, // Default page alignment
	}
}

// NewLoadProgramHeader creates a new LOAD program header.
func NewLoadProgramHeader(flags ProgramFlags) *ProgramHeader {
	return NewProgramHeader(PT_LOAD, flags)
}

// WriteTo writes the program header to the given writer in binary format.
func (ph *ProgramHeader) WriteTo(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, ph)
}

// ReadFrom reads the program header from the given reader in binary format.
func (ph *ProgramHeader) ReadFrom(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, ph)
}

// NewSectionHeader creates a new section header.
func NewSectionHeader(name uint32, typ SectionType, flags SectionFlags) *SectionHeader {
	return &SectionHeader{
		Name:  name,
		Type:  typ,
		Flags: flags,
	}
}

// NewSectionHeaderNull creates a null section header (first section is always null).
func NewSectionHeaderNull() *SectionHeader {
	return &SectionHeader{
		Type: SHT_NULL,
	}
}

// NewSectionHeaderProgBits creates a PROGBITS section header.
func NewSectionHeaderProgBits(name uint32, flags SectionFlags) *SectionHeader {
	return NewSectionHeader(name, SHT_PROGBITS, flags)
}

// NewSectionHeaderStrTab creates a STRTAB section header.
func NewSectionHeaderStrTab(name uint32) *SectionHeader {
	return NewSectionHeader(name, SHT_STRTAB, 0)
}

// NewSectionHeaderSymTab creates a SYMTAB section header.
func NewSectionHeaderSymTab(name uint32, link uint32) *SectionHeader {
	return &SectionHeader{
		Name:    name,
		Type:    SHT_SYMTAB,
		Flags:   0,
		Link:    link,
		Entsize: 24, // Size of Symbol64
	}
}

// NewSectionHeaderRela creates a RELA section header.
func NewSectionHeaderRela(name uint32, link, info uint32) *SectionHeader {
	return &SectionHeader{
		Name:    name,
		Type:    SHT_RELA,
		Flags:   0,
		Link:    link,
		Info:    info,
		Entsize: 24, // Size of Rela64
	}
}

// NewSectionHeaderNoBits creates a NOBITS section header (for BSS).
func NewSectionHeaderNoBits(name uint32, flags SectionFlags) *SectionHeader {
	return NewSectionHeader(name, SHT_NOBITS, flags)
}

// WriteTo writes the section header to the given writer in binary format.
func (sh *SectionHeader) WriteTo(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, sh)
}

// ReadFrom reads the section header from the given reader in binary format.
func (sh *SectionHeader) ReadFrom(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, sh)
}

// WriteTo writes the symbol to the given writer in binary format.
func (sym *Symbol64) WriteTo(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, sym)
}

// ReadFrom reads the symbol from the given reader in binary format.
func (sym *Symbol64) ReadFrom(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, sym)
}

// WriteTo writes the relocation entry to the given writer in binary format.
func (rel *Rel64) WriteTo(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, rel)
}

// ReadFrom reads the relocation entry from the given reader in binary format.
func (rel *Rel64) ReadFrom(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, rel)
}

// WriteTo writes the relocation entry with addend to the given writer.
func (rel *Rela64) WriteTo(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, rel)
}

// ReadFrom reads the relocation entry with addend from the given reader.
func (rel *Rela64) ReadFrom(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, rel)
}

// WriteTo writes the dynamic entry to the given writer in binary format.
func (dyn *Dyn64) WriteTo(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, dyn)
}

// ReadFrom reads the dynamic entry from the given reader in binary format.
func (dyn *Dyn64) ReadFrom(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, dyn)
}

// ============================================================================
// ELF Error
// ============================================================================

// ELFError represents an ELF parsing or validation error.
type ELFError struct {
	Message string
}

func (e *ELFError) Error() string {
	return "ELF error: " + e.Message
}

// ============================================================================
// Utility Functions
// ============================================================================

// ELFHeaderSize returns the size of an ELF64 header in bytes.
func ELFHeaderSize() int {
	return 64
}

// ProgramHeaderSize returns the size of an ELF64 program header in bytes.
func ProgramHeaderSize() int {
	return 56
}

// SectionHeaderSize returns the size of an ELF64 section header in bytes.
func SectionHeaderSize() int {
	return 64
}

// Symbol64Size returns the size of an ELF64 symbol entry in bytes.
func Symbol64Size() int {
	return 24
}

// Rela64Size returns the size of an ELF64 relocation entry with addend in bytes.
func Rela64Size() int {
	return 24
}

// Rel64Size returns the size of an ELF64 relocation entry without addend in bytes.
func Rel64Size() int {
	return 16
}

// Dyn64Size returns the size of an ELF64 dynamic entry in bytes.
func Dyn64Size() int {
	return 16
}
