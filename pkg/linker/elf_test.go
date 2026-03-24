// Package linker tests for ELF64 format handling.
package linker

import (
	"bytes"
	"testing"
)

// ============================================================================
// ELF Header Tests
// ============================================================================

func TestNewELFHeader(t *testing.T) {
	h := NewELFHeader()
	
	if h == nil {
		t.Fatal("NewELFHeader returned nil")
	}
	
	// Check magic number
	if h.Ident[0] != ELFMAG0 {
		t.Errorf("Ident[0] = 0x%02x, want 0x%02x", h.Ident[0], ELFMAG0)
	}
	if h.Ident[1] != ELFMAG1 {
		t.Errorf("Ident[1] = %c, want %c", h.Ident[1], ELFMAG1)
	}
	if h.Ident[2] != ELFMAG2 {
		t.Errorf("Ident[2] = %c, want %c", h.Ident[2], ELFMAG2)
	}
	if h.Ident[3] != ELFMAG3 {
		t.Errorf("Ident[3] = %c, want %c", h.Ident[3], ELFMAG3)
	}
	
	// Check class (64-bit)
	if h.Ident[4] != ELFCLASS64 {
		t.Errorf("Ident[4] = %d, want %d", h.Ident[4], ELFCLASS64)
	}
	
	// Check data encoding (little-endian)
	if h.Ident[5] != ELFDATA2LSB {
		t.Errorf("Ident[5] = %d, want %d", h.Ident[5], ELFDATA2LSB)
	}
	
	// Check version
	if h.Ident[6] != EV_CURRENT {
		t.Errorf("Ident[6] = %d, want %d", h.Ident[6], EV_CURRENT)
	}
	
	// Check type (executable)
	if h.Type != ET_EXEC {
		t.Errorf("Type = %d, want %d", h.Type, ET_EXEC)
	}
	
	// Check machine (x86-64)
	if h.Machine != EM_X86_64 {
		t.Errorf("Machine = %d, want %d", h.Machine, EM_X86_64)
	}
	
	// Check sizes
	if h.Ehsize != 64 {
		t.Errorf("Ehsize = %d, want 64", h.Ehsize)
	}
	if h.Phentsize != 56 {
		t.Errorf("Phentsize = %d, want 56", h.Phentsize)
	}
	if h.Shentsize != 64 {
		t.Errorf("Shentsize = %d, want 64", h.Shentsize)
	}
}

func TestNewELFHeaderShared(t *testing.T) {
	h := NewELFHeaderShared()
	
	if h.Type != ET_DYN {
		t.Errorf("Type = %d, want %d (ET_DYN)", h.Type, ET_DYN)
	}
}

func TestNewELFHeaderRelocatable(t *testing.T) {
	h := NewELFHeaderRelocatable()
	
	if h.Type != ET_REL {
		t.Errorf("Type = %d, want %d (ET_REL)", h.Type, ET_REL)
	}
}

func TestELFHeaderValidate(t *testing.T) {
	tests := []struct {
		name    string
		header  *ELFHeader
		wantErr bool
	}{
		{
			name:    "valid header",
			header:  NewELFHeader(),
			wantErr: false,
		},
		{
			name: "invalid magic",
			header: func() *ELFHeader {
				h := NewELFHeader()
				h.Ident[0] = 0x00
				return h
			}(),
			wantErr: true,
		},
		{
			name: "32-bit class",
			header: func() *ELFHeader {
				h := NewELFHeader()
				h.Ident[4] = ELFCLASS32
				return h
			}(),
			wantErr: true,
		},
		{
			name: "big-endian",
			header: func() *ELFHeader {
				h := NewELFHeader()
				h.Ident[5] = ELFDATA2MSB
				return h
			}(),
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.header.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestELFHeaderWriteTo(t *testing.T) {
	h := NewELFHeader()
	
	var buf bytes.Buffer
	_, err := h.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	
	if buf.Len() != 64 {
		t.Errorf("WriteTo() wrote %d bytes, want 64", buf.Len())
	}
}

func TestELFHeaderReadFrom(t *testing.T) {
	h1 := NewELFHeader()
	h1.Entry = 0x400000
	h1.Phoff = 64
	h1.Shoff = 0x1000
	
	var buf bytes.Buffer
	if _, err := h1.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	
	h2 := &ELFHeader{}
	if _, err := h2.ReadFrom(&buf); err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}
	
	if h2.Entry != h1.Entry {
		t.Errorf("Entry = 0x%x, want 0x%x", h2.Entry, h1.Entry)
	}
	if h2.Phoff != h1.Phoff {
		t.Errorf("Phoff = 0x%x, want 0x%x", h2.Phoff, h1.Phoff)
	}
	if h2.Shoff != h1.Shoff {
		t.Errorf("Shoff = 0x%x, want 0x%x", h2.Shoff, h1.Shoff)
	}
}

// ============================================================================
// Program Header Tests
// ============================================================================

func TestNewProgramHeader(t *testing.T) {
	ph := NewProgramHeader(PT_LOAD, PF_R|PF_X)
	
	if ph.Type != PT_LOAD {
		t.Errorf("Type = %d, want %d", ph.Type, PT_LOAD)
	}
	if ph.Flags != PF_R|PF_X {
		t.Errorf("Flags = 0x%x, want 0x%x", ph.Flags, PF_R|PF_X)
	}
	if ph.Align != 0x1000 {
		t.Errorf("Align = 0x%x, want 0x1000", ph.Align)
	}
}

func TestNewLoadProgramHeader(t *testing.T) {
	ph := NewLoadProgramHeader(PF_R | PF_W)
	
	if ph.Type != PT_LOAD {
		t.Errorf("Type = %d, want %d", ph.Type, PT_LOAD)
	}
	if ph.Flags != PF_R|PF_W {
		t.Errorf("Flags = 0x%x, want 0x%x", ph.Flags, PF_R|PF_W)
	}
}

func TestProgramHeaderWriteTo(t *testing.T) {
	ph := NewProgramHeader(PT_LOAD, PF_R|PF_X)
	ph.Offset = 0x1000
	ph.Vaddr = 0x401000
	ph.Filesz = 0x200
	ph.Memsz = 0x200
	
	var buf bytes.Buffer
	_, err := ph.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	
	if buf.Len() != 56 {
		t.Errorf("WriteTo() wrote %d bytes, want 56", buf.Len())
	}
}

func TestProgramHeaderReadFrom(t *testing.T) {
	ph1 := NewProgramHeader(PT_LOAD, PF_R|PF_X)
	ph1.Offset = 0x1000
	ph1.Vaddr = 0x401000
	ph1.Paddr = 0x401000
	ph1.Filesz = 0x200
	ph1.Memsz = 0x300
	
	var buf bytes.Buffer
	if _, err := ph1.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	
	ph2 := &ProgramHeader{}
	if _, err := ph2.ReadFrom(&buf); err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}
	
	if ph2.Offset != ph1.Offset {
		t.Errorf("Offset = 0x%x, want 0x%x", ph2.Offset, ph1.Offset)
	}
	if ph2.Vaddr != ph1.Vaddr {
		t.Errorf("Vaddr = 0x%x, want 0x%x", ph2.Vaddr, ph1.Vaddr)
	}
	if ph2.Filesz != ph1.Filesz {
		t.Errorf("Filesz = 0x%x, want 0x%x", ph2.Filesz, ph1.Filesz)
	}
}

// ============================================================================
// Section Header Tests
// ============================================================================

func TestNewSectionHeader(t *testing.T) {
	sh := NewSectionHeader(1, SHT_PROGBITS, SHF_ALLOC|SHF_EXECINSTR)
	
	if sh.Name != 1 {
		t.Errorf("Name = %d, want 1", sh.Name)
	}
	if sh.Type != SHT_PROGBITS {
		t.Errorf("Type = %d, want %d", sh.Type, SHT_PROGBITS)
	}
	if sh.Flags != SHF_ALLOC|SHF_EXECINSTR {
		t.Errorf("Flags = 0x%x, want 0x%x", sh.Flags, SHF_ALLOC|SHF_EXECINSTR)
	}
}

func TestNewSectionHeaderNull(t *testing.T) {
	sh := NewSectionHeaderNull()
	
	if sh.Type != SHT_NULL {
		t.Errorf("Type = %d, want %d", sh.Type, SHT_NULL)
	}
}

func TestNewSectionHeaderProgBits(t *testing.T) {
	sh := NewSectionHeaderProgBits(1, SHF_ALLOC|SHF_EXECINSTR)
	
	if sh.Type != SHT_PROGBITS {
		t.Errorf("Type = %d, want %d", sh.Type, SHT_PROGBITS)
	}
	if sh.Flags != SHF_ALLOC|SHF_EXECINSTR {
		t.Errorf("Flags = 0x%x, want 0x%x", sh.Flags, SHF_ALLOC|SHF_EXECINSTR)
	}
}

func TestNewSectionHeaderStrTab(t *testing.T) {
	sh := NewSectionHeaderStrTab(1)
	
	if sh.Type != SHT_STRTAB {
		t.Errorf("Type = %d, want %d", sh.Type, SHT_STRTAB)
	}
}

func TestNewSectionHeaderSymTab(t *testing.T) {
	sh := NewSectionHeaderSymTab(1, 2)
	
	if sh.Type != SHT_SYMTAB {
		t.Errorf("Type = %d, want %d", sh.Type, SHT_SYMTAB)
	}
	if sh.Link != 2 {
		t.Errorf("Link = %d, want 2", sh.Link)
	}
	if sh.Entsize != 24 {
		t.Errorf("Entsize = %d, want 24", sh.Entsize)
	}
}

func TestNewSectionHeaderRela(t *testing.T) {
	sh := NewSectionHeaderRela(1, 2, 3)
	
	if sh.Type != SHT_RELA {
		t.Errorf("Type = %d, want %d", sh.Type, SHT_RELA)
	}
	if sh.Link != 2 {
		t.Errorf("Link = %d, want 2", sh.Link)
	}
	if sh.Info != 3 {
		t.Errorf("Info = %d, want 3", sh.Info)
	}
	if sh.Entsize != 24 {
		t.Errorf("Entsize = %d, want 24", sh.Entsize)
	}
}

func TestNewSectionHeaderNoBits(t *testing.T) {
	sh := NewSectionHeaderNoBits(1, SHF_ALLOC|SHF_WRITE)
	
	if sh.Type != SHT_NOBITS {
		t.Errorf("Type = %d, want %d", sh.Type, SHT_NOBITS)
	}
	if sh.Flags != SHF_ALLOC|SHF_WRITE {
		t.Errorf("Flags = 0x%x, want 0x%x", sh.Flags, SHF_ALLOC|SHF_WRITE)
	}
}

func TestSectionHeaderWriteTo(t *testing.T) {
	sh := NewSectionHeader(1, SHT_PROGBITS, SHF_ALLOC)
	sh.Addr = 0x401000
	sh.Offset = 0x1000
	sh.Size = 0x200
	
	var buf bytes.Buffer
	_, err := sh.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	
	if buf.Len() != 64 {
		t.Errorf("WriteTo() wrote %d bytes, want 64", buf.Len())
	}
}

func TestSectionHeaderReadFrom(t *testing.T) {
	sh1 := NewSectionHeader(1, SHT_PROGBITS, SHF_ALLOC)
	sh1.Addr = 0x401000
	sh1.Offset = 0x1000
	sh1.Size = 0x200
	sh1.Link = 2
	sh1.Info = 0
	
	var buf bytes.Buffer
	if _, err := sh1.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	
	sh2 := &SectionHeader{}
	if _, err := sh2.ReadFrom(&buf); err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}
	
	if sh2.Addr != sh1.Addr {
		t.Errorf("Addr = 0x%x, want 0x%x", sh2.Addr, sh1.Addr)
	}
	if sh2.Offset != sh1.Offset {
		t.Errorf("Offset = 0x%x, want 0x%x", sh2.Offset, sh1.Offset)
	}
	if sh2.Size != sh1.Size {
		t.Errorf("Size = 0x%x, want 0x%x", sh2.Size, sh1.Size)
	}
}

// ============================================================================
// Symbol Tests
// ============================================================================

func TestNewSymbol64(t *testing.T) {
	sym := NewSymbol64(1, STB_GLOBAL, STT_FUNC, 1, 0x401000, 32)
	
	if sym.Name != 1 {
		t.Errorf("Name = %d, want 1", sym.Name)
	}
	if sym.Info != ELF64_ST_INFO(STB_GLOBAL, STT_FUNC) {
		t.Errorf("Info = 0x%02x, want 0x%02x", sym.Info, ELF64_ST_INFO(STB_GLOBAL, STT_FUNC))
	}
	if sym.Shndx != 1 {
		t.Errorf("Shndx = %d, want 1", sym.Shndx)
	}
	if sym.Value != 0x401000 {
		t.Errorf("Value = 0x%x, want 0x401000", sym.Value)
	}
	if sym.Size != 32 {
		t.Errorf("Size = %d, want 32", sym.Size)
	}
}

func TestELF64_ST_INFO(t *testing.T) {
	tests := []struct {
		bind int
		typ  int
		want byte
	}{
		{STB_LOCAL, STT_NOTYPE, 0x00},
		{STB_GLOBAL, STT_FUNC, 0x12},
		{STB_WEAK, STT_OBJECT, 0x21},
		{STB_GLOBAL, STT_SECTION, 0x13},
	}
	
	for _, tt := range tests {
		got := ELF64_ST_INFO(tt.bind, tt.typ)
		if got != tt.want {
			t.Errorf("ELF64_ST_INFO(%d, %d) = 0x%02x, want 0x%02x",
				tt.bind, tt.typ, got, tt.want)
		}
	}
}

func TestELF64_ST_BIND(t *testing.T) {
	tests := []struct {
		info byte
		want byte
	}{
		{0x00, STB_LOCAL},
		{0x12, STB_GLOBAL},
		{0x21, STB_WEAK},
		{0x13, STB_GLOBAL},
	}
	
	for _, tt := range tests {
		got := ELF64_ST_BIND(tt.info)
		if got != tt.want {
			t.Errorf("ELF64_ST_BIND(0x%02x) = %d, want %d", tt.info, got, tt.want)
		}
	}
}

func TestELF64_ST_TYPE(t *testing.T) {
	tests := []struct {
		info byte
		want byte
	}{
		{0x00, STT_NOTYPE},
		{0x12, STT_FUNC},
		{0x21, STT_OBJECT},
		{0x13, STT_SECTION},
	}
	
	for _, tt := range tests {
		got := ELF64_ST_TYPE(tt.info)
		if got != tt.want {
			t.Errorf("ELF64_ST_TYPE(0x%02x) = %d, want %d", tt.info, got, tt.want)
		}
	}
}

func TestSymbol64WriteTo(t *testing.T) {
	sym := NewSymbol64(1, STB_GLOBAL, STT_FUNC, 1, 0x401000, 32)
	
	var buf bytes.Buffer
	_, err := sym.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	
	if buf.Len() != 24 {
		t.Errorf("WriteTo() wrote %d bytes, want 24", buf.Len())
	}
}

func TestSymbol64ReadFrom(t *testing.T) {
	sym1 := NewSymbol64(1, STB_GLOBAL, STT_FUNC, 1, 0x401000, 32)
	
	var buf bytes.Buffer
	if _, err := sym1.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	
	sym2 := &Symbol64{}
	if _, err := sym2.ReadFrom(&buf); err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}
	
	if sym2.Name != sym1.Name {
		t.Errorf("Name = %d, want %d", sym2.Name, sym1.Name)
	}
	if sym2.Info != sym1.Info {
		t.Errorf("Info = 0x%02x, want 0x%02x", sym2.Info, sym1.Info)
	}
	if sym2.Value != sym1.Value {
		t.Errorf("Value = 0x%x, want 0x%x", sym2.Value, sym1.Value)
	}
}

// ============================================================================
// Relocation Tests
// ============================================================================

func TestNewRela64(t *testing.T) {
	rel := NewRela64(0x100, 1, R_X86_64_64, 0)
	
	if rel.Offset != 0x100 {
		t.Errorf("Offset = 0x%x, want 0x100", rel.Offset)
	}
	if ELF64_R_SYM(rel.Info) != 1 {
		t.Errorf("Symbol = %d, want 1", ELF64_R_SYM(rel.Info))
	}
	if ELF64_R_TYPE(rel.Info) != R_X86_64_64 {
		t.Errorf("Type = %d, want %d", ELF64_R_TYPE(rel.Info), R_X86_64_64)
	}
	if rel.Addend != 0 {
		t.Errorf("Addend = %d, want 0", rel.Addend)
	}
}

func TestELF64_R_INFO(t *testing.T) {
	tests := []struct {
		sym  uint32
		typ  uint32
		want uint64
	}{
		{0, 0, 0},
		{1, R_X86_64_64, (1 << 32) | R_X86_64_64},
		{2, R_X86_64_PC32, (2 << 32) | R_X86_64_PC32},
		{0xffffffff, 0xffffffff, 0xffffffffffffffff},
	}
	
	for _, tt := range tests {
		got := ELF64_R_INFO(tt.sym, tt.typ)
		if got != tt.want {
			t.Errorf("ELF64_R_INFO(%d, %d) = 0x%x, want 0x%x",
				tt.sym, tt.typ, got, tt.want)
		}
	}
}

func TestELF64_R_SYM(t *testing.T) {
	tests := []struct {
		info uint64
		want uint32
	}{
		{0, 0},
		{(1 << 32) | R_X86_64_64, 1},
		{(2 << 32) | R_X86_64_PC32, 2},
		{0xffffffffffffffff, 0xffffffff},
	}
	
	for _, tt := range tests {
		got := ELF64_R_SYM(tt.info)
		if got != tt.want {
			t.Errorf("ELF64_R_SYM(0x%x) = %d, want %d", tt.info, got, tt.want)
		}
	}
}

func TestELF64_R_TYPE(t *testing.T) {
	tests := []struct {
		info uint64
		want uint32
	}{
		{0, 0},
		{(1 << 32) | R_X86_64_64, R_X86_64_64},
		{(2 << 32) | R_X86_64_PC32, R_X86_64_PC32},
		{0xffffffffffffffff, 0xffffffff},
	}
	
	for _, tt := range tests {
		got := ELF64_R_TYPE(tt.info)
		if got != tt.want {
			t.Errorf("ELF64_R_TYPE(0x%x) = %d, want %d", tt.info, got, tt.want)
		}
	}
}

func TestRela64WriteTo(t *testing.T) {
	rel := NewRela64(0x100, 1, R_X86_64_64, 42)
	
	var buf bytes.Buffer
	_, err := rel.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	
	if buf.Len() != 24 {
		t.Errorf("WriteTo() wrote %d bytes, want 24", buf.Len())
	}
}

func TestRela64ReadFrom(t *testing.T) {
	rel1 := NewRela64(0x100, 1, R_X86_64_64, 42)
	
	var buf bytes.Buffer
	if _, err := rel1.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	
	rel2 := &Rela64{}
	if _, err := rel2.ReadFrom(&buf); err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}
	
	if rel2.Offset != rel1.Offset {
		t.Errorf("Offset = 0x%x, want 0x%x", rel2.Offset, rel1.Offset)
	}
	if rel2.Info != rel1.Info {
		t.Errorf("Info = 0x%x, want 0x%x", rel2.Info, rel1.Info)
	}
	if rel2.Addend != rel1.Addend {
		t.Errorf("Addend = %d, want %d", rel2.Addend, rel1.Addend)
	}
}

func TestRel64WriteTo(t *testing.T) {
	rel := &Rel64{
		Offset: 0x100,
		Info:   ELF64_R_INFO(1, R_X86_64_64),
	}
	
	var buf bytes.Buffer
	_, err := rel.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	
	if buf.Len() != 16 {
		t.Errorf("WriteTo() wrote %d bytes, want 16", buf.Len())
	}
}

// ============================================================================
// Dynamic Section Tests
// ============================================================================

func TestNewDyn64(t *testing.T) {
	dyn := NewDyn64(DT_NULL, 0)
	
	if dyn.Tag != DT_NULL {
		t.Errorf("Tag = %d, want %d", dyn.Tag, DT_NULL)
	}
	if dyn.Val != 0 {
		t.Errorf("Val = %d, want 0", dyn.Val)
	}
}

func TestDyn64WriteTo(t *testing.T) {
	dyn := NewDyn64(DT_NEEDED, 1)
	
	var buf bytes.Buffer
	_, err := dyn.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	
	if buf.Len() != 16 {
		t.Errorf("WriteTo() wrote %d bytes, want 16", buf.Len())
	}
}

func TestDyn64ReadFrom(t *testing.T) {
	dyn1 := NewDyn64(DT_NEEDED, 1)
	
	var buf bytes.Buffer
	if _, err := dyn1.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	
	dyn2 := &Dyn64{}
	if _, err := dyn2.ReadFrom(&buf); err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}
	
	if dyn2.Tag != dyn1.Tag {
		t.Errorf("Tag = %d, want %d", dyn2.Tag, dyn1.Tag)
	}
	if dyn2.Val != dyn1.Val {
		t.Errorf("Val = %d, want %d", dyn2.Val, dyn1.Val)
	}
}

// ============================================================================
// Size Function Tests
// ============================================================================

func TestELFHeaderSize(t *testing.T) {
	if ELFHeaderSize() != 64 {
		t.Errorf("ELFHeaderSize() = %d, want 64", ELFHeaderSize())
	}
}

func TestProgramHeaderSize(t *testing.T) {
	if ProgramHeaderSize() != 56 {
		t.Errorf("ProgramHeaderSize() = %d, want 56", ProgramHeaderSize())
	}
}

func TestSectionHeaderSize(t *testing.T) {
	if SectionHeaderSize() != 64 {
		t.Errorf("SectionHeaderSize() = %d, want 64", SectionHeaderSize())
	}
}

func TestSymbol64Size(t *testing.T) {
	if Symbol64Size() != 24 {
		t.Errorf("Symbol64Size() = %d, want 24", Symbol64Size())
	}
}

func TestRela64Size(t *testing.T) {
	if Rela64Size() != 24 {
		t.Errorf("Rela64Size() = %d, want 24", Rela64Size())
	}
}

func TestRel64Size(t *testing.T) {
	if Rel64Size() != 16 {
		t.Errorf("Rel64Size() = %d, want 16", Rel64Size())
	}
}

func TestDyn64Size(t *testing.T) {
	if Dyn64Size() != 16 {
		t.Errorf("Dyn64Size() = %d, want 16", Dyn64Size())
	}
}

// ============================================================================
// ELF Error Tests
// ============================================================================

func TestELFError(t *testing.T) {
	err := &ELFError{Message: "test error"}
	
	if err.Error() != "ELF error: test error" {
		t.Errorf("Error() = %q, want %q", err.Error(), "ELF error: test error")
	}
}

// ============================================================================
// Integration Test
// ============================================================================

func TestELFStructureSerialization(t *testing.T) {
	// Create a minimal ELF structure
	h := NewELFHeader()
	h.Entry = 0x400000
	h.Phoff = 64
	h.Phnum = 2
	h.Shoff = 64 + 2*56
	h.Shnum = 4
	h.Shstrndx = 3
	
	ph1 := NewLoadProgramHeader(PF_R | PF_X)
	ph1.Offset = 64 + 2*56
	ph1.Vaddr = 0x400000
	ph1.Filesz = 0x200
	ph1.Memsz = 0x200
	
	ph2 := NewLoadProgramHeader(PF_R | PF_W)
	ph2.Offset = 64 + 2*56 + 0x200
	ph2.Vaddr = 0x400200
	ph2.Filesz = 0x100
	ph2.Memsz = 0x200
	
	sh1 := NewSectionHeaderNull()
	sh2 := NewSectionHeaderProgBits(7, SHF_ALLOC|SHF_EXECINSTR)
	sh2.Offset = ph1.Offset
	sh2.Addr = ph1.Vaddr
	sh2.Size = ph1.Filesz
	sh3 := NewSectionHeaderProgBits(14, SHF_ALLOC|SHF_WRITE)
	sh3.Offset = ph2.Offset
	sh3.Addr = ph2.Vaddr
	sh3.Size = ph2.Filesz
	sh4 := NewSectionHeaderStrTab(20)
	
	var buf bytes.Buffer
	
	// Write ELF header
	if _, err := h.WriteTo(&buf); err != nil {
		t.Fatalf("Write ELF header: %v", err)
	}
	
	// Write program headers
	if _, err := ph1.WriteTo(&buf); err != nil {
		t.Fatalf("Write program header 1: %v", err)
	}
	if _, err := ph2.WriteTo(&buf); err != nil {
		t.Fatalf("Write program header 2: %v", err)
	}
	
	// Write section headers
	if _, err := sh1.WriteTo(&buf); err != nil {
		t.Fatalf("Write section header 1: %v", err)
	}
	if _, err := sh2.WriteTo(&buf); err != nil {
		t.Fatalf("Write section header 2: %v", err)
	}
	if _, err := sh3.WriteTo(&buf); err != nil {
		t.Fatalf("Write section header 3: %v", err)
	}
	if _, err := sh4.WriteTo(&buf); err != nil {
		t.Fatalf("Write section header 4: %v", err)
	}
	
	// Verify total size
	expectedSize := 64 + 2*56 + 4*64
	if buf.Len() != expectedSize {
		t.Errorf("Total size = %d, want %d", buf.Len(), expectedSize)
	}
}

// ============================================================================
// Constant Tests
// ============================================================================

func TestELFConstants(t *testing.T) {
	// Magic number
	if ELFMAG0 != 0x7f {
		t.Errorf("ELFMAG0 = 0x%x, want 0x7f", ELFMAG0)
	}
	if string([]byte{ELFMAG0, ELFMAG1, ELFMAG2, ELFMAG3}) != "\x7fELF" {
		t.Error("ELFMAG string mismatch")
	}
	
	// Class
	if ELFCLASS64 != 2 {
		t.Errorf("ELFCLASS64 = %d, want 2", ELFCLASS64)
	}
	
	// Data encoding
	if ELFDATA2LSB != 1 {
		t.Errorf("ELFDATA2LSB = %d, want 1", ELFDATA2LSB)
	}
	
	// File types
	if ET_EXEC != 2 {
		t.Errorf("ET_EXEC = %d, want 2", ET_EXEC)
	}
	if ET_DYN != 3 {
		t.Errorf("ET_DYN = %d, want 3", ET_DYN)
	}
	
	// Machine types
	if EM_X86_64 != 62 {
		t.Errorf("EM_X86_64 = %d, want 62", EM_X86_64)
	}
}

func TestProgramHeaderConstants(t *testing.T) {
	if PT_NULL != 0 {
		t.Errorf("PT_NULL = %d, want 0", PT_NULL)
	}
	if PT_LOAD != 1 {
		t.Errorf("PT_LOAD = %d, want 1", PT_LOAD)
	}
	if PT_DYNAMIC != 2 {
		t.Errorf("PT_DYNAMIC = %d, want 2", PT_DYNAMIC)
	}
	
	if PF_X != 0x1 {
		t.Errorf("PF_X = 0x%x, want 0x1", PF_X)
	}
	if PF_W != 0x2 {
		t.Errorf("PF_W = 0x%x, want 0x2", PF_W)
	}
	if PF_R != 0x4 {
		t.Errorf("PF_R = 0x%x, want 0x4", PF_R)
	}
}

func TestSectionHeaderConstants(t *testing.T) {
	if SHT_NULL != 0 {
		t.Errorf("SHT_NULL = %d, want 0", SHT_NULL)
	}
	if SHT_PROGBITS != 1 {
		t.Errorf("SHT_PROGBITS = %d, want 1", SHT_PROGBITS)
	}
	if SHT_SYMTAB != 2 {
		t.Errorf("SHT_SYMTAB = %d, want 2", SHT_SYMTAB)
	}
	if SHT_STRTAB != 3 {
		t.Errorf("SHT_STRTAB = %d, want 3", SHT_STRTAB)
	}
	if SHT_NOBITS != 8 {
		t.Errorf("SHT_NOBITS = %d, want 8", SHT_NOBITS)
	}
	
	if SHF_WRITE != 0x1 {
		t.Errorf("SHF_WRITE = 0x%x, want 0x1", SHF_WRITE)
	}
	if SHF_ALLOC != 0x2 {
		t.Errorf("SHF_ALLOC = 0x%x, want 0x2", SHF_ALLOC)
	}
	if SHF_EXECINSTR != 0x4 {
		t.Errorf("SHF_EXECINSTR = 0x%x, want 0x4", SHF_EXECINSTR)
	}
}

func TestSymbolConstants(t *testing.T) {
	if STB_LOCAL != 0 {
		t.Errorf("STB_LOCAL = %d, want 0", STB_LOCAL)
	}
	if STB_GLOBAL != 1 {
		t.Errorf("STB_GLOBAL = %d, want 1", STB_GLOBAL)
	}
	if STB_WEAK != 2 {
		t.Errorf("STB_WEAK = %d, want 2", STB_WEAK)
	}
	
	if STT_NOTYPE != 0 {
		t.Errorf("STT_NOTYPE = %d, want 0", STT_NOTYPE)
	}
	if STT_OBJECT != 1 {
		t.Errorf("STT_OBJECT = %d, want 1", STT_OBJECT)
	}
	if STT_FUNC != 2 {
		t.Errorf("STT_FUNC = %d, want 2", STT_FUNC)
	}
}

func TestRelocationConstants(t *testing.T) {
	if R_X86_64_NONE != 0 {
		t.Errorf("R_X86_64_NONE = %d, want 0", R_X86_64_NONE)
	}
	if R_X86_64_64 != 1 {
		t.Errorf("R_X86_64_64 = %d, want 1", R_X86_64_64)
	}
	if R_X86_64_PC32 != 2 {
		t.Errorf("R_X86_64_PC32 = %d, want 2", R_X86_64_PC32)
	}
	if R_X86_64_32 != 10 {
		t.Errorf("R_X86_64_32 = %d, want 10", R_X86_64_32)
	}
}

func TestDynamicConstants(t *testing.T) {
	if DT_NULL != 0 {
		t.Errorf("DT_NULL = %d, want 0", DT_NULL)
	}
	if DT_NEEDED != 1 {
		t.Errorf("DT_NEEDED = %d, want 1", DT_NEEDED)
	}
	if DT_STRTAB != 5 {
		t.Errorf("DT_STRTAB = %d, want 5", DT_STRTAB)
	}
	if DT_SYMTAB != 6 {
		t.Errorf("DT_SYMTAB = %d, want 6", DT_SYMTAB)
	}
}