// Package linker tests for the linker implementation.
package linker

import (
	"bytes"
	"encoding/binary"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/akzj/goc/internal/errhand"
)

// ============================================================================
// StringTable Tests
// ============================================================================

func TestStringTableNew(t *testing.T) {
	st := NewStringTable()

	if st == nil {
		t.Fatal("NewStringTable returned nil")
	}

	if st.strings[""] != 0 {
		t.Errorf("Empty string index = %d, want 0", st.strings[""])
	}

	if len(st.data) != 1 || st.data[0] != 0 {
		t.Error("Initial string table should contain only null byte")
	}
}

func TestStringTableAdd(t *testing.T) {
	st := NewStringTable()

	idx1 := st.Add("hello")
	if idx1 != 1 {
		t.Errorf("First string index = %d, want 1", idx1)
	}

	idx2 := st.Add("world")
	if idx2 != 7 {
		t.Errorf("Second string index = %d, want 7", idx2)
	}

	idx3 := st.Add("hello")
	if idx3 != idx1 {
		t.Errorf("Duplicate string index = %d, want %d", idx3, idx1)
	}
}

func TestStringTableData(t *testing.T) {
	st := NewStringTable()
	st.Add("test")

	data := st.Data()
	expected := []byte{0, 't', 'e', 's', 't', 0}
	if !bytes.Equal(data, expected) {
		t.Errorf("Data = %v, want %v", data, expected)
	}
}

func TestStringTableSize(t *testing.T) {
	st := NewStringTable()

	if st.Size() != 1 {
		t.Errorf("Initial size = %d, want 1", st.Size())
	}

	st.Add("hello")
	if st.Size() != 7 {
		t.Errorf("Size after adding 'hello' = %d, want 7", st.Size())
	}
}

// ============================================================================
// Linker Basic Tests
// ============================================================================

func TestNewLinker(t *testing.T) {
	errHandler := errhand.NewErrorHandler()
	l := NewLinker(errHandler)

	if l == nil {
		t.Fatal("NewLinker returned nil")
	}

	if l.symbols == nil {
		t.Error("Linker symbols should not be nil")
	}

	if l.entry != "_start" {
		t.Errorf("Default entry = %s, want _start", l.entry)
	}

	if l.baseAddr != 0x400000 {
		t.Errorf("Default baseAddr = 0x%x, want 0x400000", l.baseAddr)
	}
}

func TestLinkerSetEntryPoint(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	l.SetEntryPoint("main")
	if l.GetEntryPoint() != "main" {
		t.Errorf("EntryPoint = %s, want main", l.GetEntryPoint())
	}
}

// ============================================================================
// Object File Parsing Tests
// ============================================================================

func TestLinkerParseObjectFile(t *testing.T) {
	asmCode := `
	.text
	.globl main
	.type main, @function
main:
	pushq %rbp
	movq %rsp, %rbp
	movl $42, %eax
	popq %rbp
	ret
	.size main, .-main
`

	tmpAsm, err := os.CreateTemp("", "goc-test-*.s")
	if err != nil {
		t.Skipf("Cannot create temp file: %v", err)
	}
	tmpAsmName := tmpAsm.Name()
	defer os.Remove(tmpAsmName)

	if _, err := tmpAsm.WriteString(asmCode); err != nil {
		tmpAsm.Close()
		t.Skipf("Cannot write temp file: %v", err)
	}
	tmpAsm.Close()

	objFile := tmpAsmName[:len(tmpAsmName)-2] + ".o"
	defer os.Remove(objFile)

	cmd := exec.Command("as", "-o", objFile, tmpAsmName)
	if err := cmd.Run(); err != nil {
		t.Skipf("System assembler not available: %v", err)
	}

	l := NewLinker(errhand.NewErrorHandler())
	obj, err := l.parseObjectFile(objFile)
	if err != nil {
		t.Fatalf("parseObjectFile failed: %v", err)
	}

	if !strings.HasSuffix(obj.Name, ".o") {
		t.Errorf("Object file name = %s, should end with .o", obj.Name)
	}

	if _, ok := obj.Sections[".text"]; !ok {
		t.Error("Object file should have .text section")
	}

	mainSym, found := obj.Symbols.Lookup("main")
	if !found {
		t.Error("Object file should have main symbol")
	} else {
		if !mainSym.IsFunction() {
			t.Error("main symbol should be a function")
		}
		if !mainSym.IsDefined() {
			t.Error("main symbol should be defined")
		}
	}
}

// ============================================================================
// Symbol Resolution Tests
// ============================================================================

func TestLinkerResolveSymbols(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	mainSym := NewFunctionSymbol("main", STB_GLOBAL, NewCodeSection(), 0x1000, 32)
	obj1 := ObjectFile{
		Name: "main.o",
		Sections: map[string]*Section{
			".text": NewCodeSection(),
		},
		Symbols: NewSymbolTable(),
	}
	obj1.Symbols.Add(mainSym)

	printfSym := NewSymbol("printf", STB_GLOBAL, STT_FUNC)
	obj2 := ObjectFile{
		Name: "io.o",
		Sections: map[string]*Section{
			".text": NewCodeSection(),
		},
		Symbols: NewSymbolTable(),
	}
	obj2.Symbols.Add(printfSym)

	l.loadObjectFile(obj1)
	l.loadObjectFile(obj2)

	err := l.ResolveSymbols()
	if err != nil {
		if !strings.Contains(err.Error(), "undefined symbols") {
			t.Errorf("ResolveSymbols failed unexpectedly: %v", err)
		}
	}
}

func TestLinkerResolveSymbolsNoEntry(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	sym := NewFunctionSymbol("other_func", STB_GLOBAL, NewCodeSection(), 0x1000, 32)
	obj := ObjectFile{
		Name: "other.o",
		Sections: map[string]*Section{
			".text": NewCodeSection(),
		},
		Symbols: NewSymbolTable(),
	}
	obj.Symbols.Add(sym)

	l.loadObjectFile(obj)

	err := l.ResolveSymbols()
	if err == nil {
		if l.symbols.HasUndefined() {
			return
		}
		t.Error("ResolveSymbols should fail without entry point")
	}
}

func TestLinkerMultipleDefinitions(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	sym1 := NewFunctionSymbol("duplicate", STB_GLOBAL, NewCodeSection(), 0x1000, 32)
	obj1 := ObjectFile{
		Name: "first.o",
		Sections: map[string]*Section{
			".text": NewCodeSection(),
		},
		Symbols: NewSymbolTable(),
	}
	obj1.Symbols.Add(sym1)

	sym2 := NewFunctionSymbol("duplicate", STB_GLOBAL, NewCodeSection(), 0x2000, 32)
	obj2 := ObjectFile{
		Name: "second.o",
		Sections: map[string]*Section{
			".text": NewCodeSection(),
		},
		Symbols: NewSymbolTable(),
	}
	obj2.Symbols.Add(sym2)

	l.loadObjectFile(obj1)
	err := l.loadObjectFile(obj2)

	if err == nil {
		t.Error("loadObjectFile should fail with multiple definitions")
	}
}

// ============================================================================
// Address Assignment Tests
// ============================================================================

func TestLinkerAssignAddresses(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	textSec := NewCodeSection()
	for i := 0; i < 100; i++ {
		textSec.AddByte(0x90)
	}

	dataSec := NewDataSection()
	for i := 0; i < 50; i++ {
		dataSec.AddByte(0x00)
	}

	obj := ObjectFile{
		Name: "test.o",
		Sections: map[string]*Section{
			".text": textSec,
			".data": dataSec,
		},
		Symbols: NewSymbolTable(),
	}

	l.loadObjectFile(obj)
	l.assignAddresses()

	if textSec.Addr != 0x400000 {
		t.Errorf("Text section addr = 0x%x, want 0x400000", textSec.Addr)
	}

	expectedDataAddr := uint64(0x401000)
	if dataSec.Addr != expectedDataAddr {
		t.Errorf("Data section addr = 0x%x, want 0x%x", dataSec.Addr, expectedDataAddr)
	}
}

// ============================================================================
// Relocation Tests
// ============================================================================

func TestLinkerRelocate(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	section := NewCodeSection()
	section.AddUint64(0)
	section.SetAddr(0)

	symbol := NewFunctionSymbol("target", STB_GLOBAL, section, 0x1000, 32)

	rel := &Relocation{
		Offset:  0,
		Type:    R_X86_64_64,
		Symbol:  symbol,
		Addend:  0,
		Section: section,
	}

	obj := ObjectFile{
		Name: "test.o",
		Sections: map[string]*Section{
			".text": section,
		},
		Symbols:     NewSymbolTable(),
		Relocations: []*Relocation{rel},
	}
	obj.Symbols.Add(symbol)

	l.loadObjectFile(obj)
	l.assignAddresses()

	err := l.Relocate()
	if err != nil {
		t.Fatalf("Relocate failed: %v", err)
	}

	value := binary.LittleEndian.Uint64(section.Data)
	if value != 0x401000 {
		t.Errorf("Relocated value = 0x%x, want 0x401000", value)
	}
}

func TestLinkerRelocatePCRelative(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	section := NewCodeSection()
	section.AddUint32(0)
	section.SetAddr(0x401000)

	symbol := NewFunctionSymbol("target", STB_GLOBAL, section, 0x401050, 32)

	rel := &Relocation{
		Offset:  0,
		Type:    R_X86_64_PLT32,
		Symbol:  symbol,
		Addend:  0,
		Section: section,
	}

	obj := ObjectFile{
		Name: "test.o",
		Sections: map[string]*Section{
			".text": section,
		},
		Symbols:     NewSymbolTable(),
		Relocations: []*Relocation{rel},
	}
	obj.Symbols.Add(symbol)

	l.loadObjectFile(obj)
	l.assignAddresses()

	err := l.Relocate()
	if err != nil {
		t.Fatalf("Relocate failed: %v", err)
	}

	value := binary.LittleEndian.Uint32(section.Data)
	// ELF PC32: S + A - P with P = sectionAddr + r_offset (start of 32-bit field).
	// After assignAddresses: section 0x400000, symbol 0x400050, offset 0 → 0x50.
	if value != 0x50 {
		t.Errorf("Relocated value = 0x%x, want 0x50", value)
	}
}

// ============================================================================
// ELF Emission Tests
// ============================================================================

func TestLinkerEmit(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	textSec := NewCodeSection()
	textSec.AddBytes(0x90, 0x90, 0x90)
	textSec.SetAddr(0x401000)

	mainSym := NewFunctionSymbol("main", STB_GLOBAL, textSec, 0x401000, 3)

	obj := ObjectFile{
		Name: "test.o",
		Sections: map[string]*Section{
			".text": textSec,
		},
		Symbols: NewSymbolTable(),
	}
	obj.Symbols.Add(mainSym)

	l.loadObjectFile(obj)
	l.assignAddresses()
	l.Relocate()

	elfData, err := l.Emit()
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	if len(elfData) < 64 {
		t.Fatal("ELF binary too small")
	}

	if string(elfData[0:4]) != "\x7fELF" {
		t.Error("Invalid ELF magic number")
	}

	if elfData[4] != ELFCLASS64 {
		t.Error("Not a 64-bit ELF")
	}

	if elfData[5] != ELFDATA2LSB {
		t.Error("Not little-endian")
	}

	typ := binary.LittleEndian.Uint16(elfData[16:18])
	if typ != ET_EXEC {
		t.Errorf("ELF type = %d, want %d (ET_EXEC)", typ, ET_EXEC)
	}

	machine := binary.LittleEndian.Uint16(elfData[18:20])
	if machine != EM_X86_64 {
		t.Errorf("Machine = %d, want %d (EM_X86_64)", machine, EM_X86_64)
	}
}

func TestLinkerEmitWithSections(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	textSec := NewCodeSection()
	textSec.AddBytes(0x90, 0x90)
	textSec.SetAddr(0x401000)

	dataSec := NewDataSection()
	dataSec.AddUint64(42)
	dataSec.SetAddr(0x402000)

	obj := ObjectFile{
		Name: "test.o",
		Sections: map[string]*Section{
			".text": textSec,
			".data": dataSec,
		},
		Symbols: NewSymbolTable(),
	}

	l.loadObjectFile(obj)
	l.assignAddresses()

	binary, err := l.Emit()
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	if len(binary) < 64+56+64 {
		t.Error("ELF binary too small for headers")
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestLinkerAssembleAndLink(t *testing.T) {
	if _, err := exec.LookPath("as"); err != nil {
		t.Skip("System assembler not available")
	}

	asmCode := `
	.text
	.globl _start
	.type _start, @function
_start:
	movq $60, %rax
	movq $0, %rdi
	syscall
	.size _start, .-_start
`

	tmpAsm, err := os.CreateTemp("", "goc-integration-*.s")
	if err != nil {
		t.Skipf("Cannot create temp file: %v", err)
	}
	tmpAsmName := tmpAsm.Name()
	defer os.Remove(tmpAsmName)

	if _, err := tmpAsm.WriteString(asmCode); err != nil {
		tmpAsm.Close()
		t.Skipf("Cannot write temp file: %v", err)
	}
	tmpAsm.Close()

	tmpExe, err := os.CreateTemp("", "goc-integration-*.exe")
	if err != nil {
		t.Skipf("Cannot create temp exe: %v", err)
	}
	tmpExeName := tmpExe.Name()
	tmpExe.Close()
	defer os.Remove(tmpExeName)

	l := NewLinker(errhand.NewErrorHandler())
	err = l.AssembleAndLink([]string{tmpAsmName}, tmpExeName, false)
	if err != nil {
		t.Fatalf("AssembleAndLink failed: %v", err)
	}

	info, err := os.Stat(tmpExeName)
	if err != nil {
		t.Fatalf("Output file not created: %v", err)
	}

	if info.Size() < 64 {
		t.Error("Output file too small to be valid ELF")
	}

	data, err := os.ReadFile(tmpExeName)
	if err != nil {
		t.Fatalf("Cannot read output file: %v", err)
	}

	if string(data[0:4]) != "\x7fELF" {
		t.Error("Output is not a valid ELF file")
	}
}

func TestLinkerCompileToObject(t *testing.T) {
	if _, err := exec.LookPath("as"); err != nil {
		t.Skip("System assembler not available")
	}

	asmCode := `
	.text
	.globl main
	.type main, @function
main:
	pushq %rbp
	movq %rsp, %rbp
	movl $42, %eax
	popq %rbp
	ret
	.size main, .-main
`

	tmpAsm, err := os.CreateTemp("", "goc-object-*.s")
	if err != nil {
		t.Skipf("Cannot create temp file: %v", err)
	}
	tmpAsmName := tmpAsm.Name()
	defer os.Remove(tmpAsmName)

	if _, err := tmpAsm.WriteString(asmCode); err != nil {
		tmpAsm.Close()
		t.Skipf("Cannot write temp file: %v", err)
	}
	tmpAsm.Close()

	tmpObj, err := os.CreateTemp("", "goc-object-*.o")
	if err != nil {
		t.Skipf("Cannot create temp obj: %v", err)
	}
	tmpObjName := tmpObj.Name()
	tmpObj.Close()
	defer os.Remove(tmpObjName)

	l := NewLinker(errhand.NewErrorHandler())
	err = l.CompileToObject(asmCode, tmpObjName)
	if err != nil {
		t.Fatalf("CompileToObject failed: %v", err)
	}

	info, err := os.Stat(tmpObjName)
	if err != nil {
		t.Fatalf("Output file not created: %v", err)
	}

	if info.Size() < 64 {
		t.Error("Output file too small to be valid ELF object")
	}

	data, err := os.ReadFile(tmpObjName)
	if err != nil {
		t.Fatalf("Cannot read output file: %v", err)
	}

	if string(data[0:4]) != "\x7fELF" {
		t.Error("Output is not a valid ELF file")
	}

	typ := binary.LittleEndian.Uint16(data[16:18])
	if typ != ET_REL {
		t.Errorf("Object file type = %d, want %d (ET_REL)", typ, ET_REL)
	}
}

func TestLinkerLinkAssembly(t *testing.T) {
	if _, err := exec.LookPath("as"); err != nil {
		t.Skip("System assembler not available")
	}

	asmCode := `
	.text
	.globl _start
	.type _start, @function
_start:
	movq $60, %rax
	movq $0, %rdi
	syscall
	.size _start, .-_start
`

	tmpExe, err := os.CreateTemp("", "goc-link-*.exe")
	if err != nil {
		t.Skipf("Cannot create temp exe: %v", err)
	}
	tmpExeName := tmpExe.Name()
	tmpExe.Close()
	defer os.Remove(tmpExeName)

	l := NewLinker(errhand.NewErrorHandler())
	err = l.LinkAssembly(asmCode, tmpExeName)
	if err != nil {
		t.Fatalf("LinkAssembly failed: %v", err)
	}

	data, err := os.ReadFile(tmpExeName)
	if err != nil {
		t.Fatalf("Cannot read output file: %v", err)
	}

	if string(data[0:4]) != "\x7fELF" {
		t.Error("Output is not a valid ELF file")
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestLinkerEmptyObjectFile(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	obj := ObjectFile{
		Name:     "empty.o",
		Sections: make(map[string]*Section),
		Symbols:  NewSymbolTable(),
	}

	err := l.loadObjectFile(obj)
	if err != nil {
		t.Errorf("loadObjectFile failed for empty object: %v", err)
	}
}

func TestLinkerRelocateUndefinedSymbol(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	section := NewCodeSection()
	section.AddUint64(0)
	section.SetAddr(0x401000)

	symbol := NewSymbol("undefined", STB_GLOBAL, STT_FUNC)

	rel := &Relocation{
		Offset:  0,
		Type:    R_X86_64_64,
		Symbol:  symbol,
		Addend:  0,
		Section: section,
	}

	obj := ObjectFile{
		Name: "test.o",
		Sections: map[string]*Section{
			".text": section,
		},
		Symbols:     NewSymbolTable(),
		Relocations: []*Relocation{rel},
	}

	l.loadObjectFile(obj)
	l.assignAddresses()

	err := l.Relocate()
	if err == nil {
		t.Error("Relocate should fail with undefined symbol")
	}
}

func TestLinkerRelocateNilSymbol(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	section := NewCodeSection()
	section.AddUint64(0)
	section.SetAddr(0x401000)

	rel := &Relocation{
		Offset:  0,
		Type:    R_X86_64_64,
		Symbol:  nil,
		Addend:  0,
		Section: section,
	}

	obj := ObjectFile{
		Name: "test.o",
		Sections: map[string]*Section{
			".text": section,
		},
		Symbols:     NewSymbolTable(),
		Relocations: []*Relocation{rel},
	}

	l.loadObjectFile(obj)
	l.assignAddresses()

	err := l.Relocate()
	if err == nil {
		t.Error("Relocate should fail with nil symbol")
	}
}

func TestLinkerUnsupportedRelocationType(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	section := NewCodeSection()
	section.AddUint64(0)
	section.SetAddr(0x401000)

	symbol := NewFunctionSymbol("target", STB_GLOBAL, section, 0x402000, 32)

	rel := &Relocation{
		Offset:  0,
		Type:    999,
		Symbol:  symbol,
		Addend:  0,
		Section: section,
	}

	obj := ObjectFile{
		Name: "test.o",
		Sections: map[string]*Section{
			".text": section,
		},
		Symbols:     NewSymbolTable(),
		Relocations: []*Relocation{rel},
	}
	obj.Symbols.Add(symbol)

	l.loadObjectFile(obj)
	l.assignAddresses()

	err := l.Relocate()
	if err == nil {
		t.Error("Relocate should fail with unsupported relocation type")
	}
}

// ============================================================================
// Program Header Tests
// ============================================================================

func TestLinkerCreateProgramHeaders(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	textSec := NewCodeSection()
	for i := 0; i < 100; i++ {
		textSec.AddByte(0x90)
	}
	textSec.SetAddr(0x401000)

	dataSec := NewDataSection()
	for i := 0; i < 50; i++ {
		dataSec.AddByte(0x00)
	}
	dataSec.SetAddr(0x402000)

	l.sections = append(l.sections, textSec, dataSec)

	phs := l.createProgramHeaders()

	if len(phs) < 2 {
		t.Errorf("Expected at least 2 program headers, got %d", len(phs))
	}

	if phs[0].Type != PT_NULL {
		t.Error("First program header should be NULL")
	}

	hasTextLoad := false
	hasDataLoad := false

	for _, ph := range phs {
		if ph.Type == PT_LOAD {
			if ph.Flags&(PF_R|PF_X) == (PF_R | PF_X) {
				hasTextLoad = true
			}
			if ph.Flags&(PF_R|PF_W) == (PF_R | PF_W) {
				hasDataLoad = true
			}
		}
	}

	if !hasTextLoad {
		t.Error("Should have LOAD header for text segment")
	}
	if !hasDataLoad {
		t.Error("Should have LOAD header for data segment")
	}
}

// ============================================================================
// Section Header Tests
// ============================================================================

func TestLinkerCreateSectionHeaders(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	textSec := NewCodeSection()
	textSec.AddBytes(0x90, 0x90)
	textSec.SetAddr(0x401000)

	l.sections = append(l.sections, textSec)

	headers, data := l.createSectionHeaders(0)

	if len(headers) < 1 {
		t.Errorf("Expected at least 1 section header, got %d", len(headers))
	}

	if headers[0].Type != SHT_NULL {
		t.Error("First section header should be NULL")
	}

	if len(data) == 0 {
		t.Error("Section data should not be empty")
	}
}

// ============================================================================
// Additional Coverage Tests
// ============================================================================

func TestLinkerWriteUint64OutOfBounds(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())
	data := make([]byte, 4)
	l.writeUint64(data, 2, 0x1234567890ABCDEF)
}

func TestLinkerWriteUint32OutOfBounds(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())
	data := make([]byte, 2)
	l.writeUint32(data, 1, 0x12345678)
}

func TestLinkerLinkWithLibraries(t *testing.T) {
	// Create fresh linker for this test
	l := NewLinker(errhand.NewErrorHandler())

	textSec := NewCodeSection()
	textSec.AddBytes(0x90, 0x90)

	// Use _start as the entry point
	startSym := NewFunctionSymbol("_start", STB_GLOBAL, textSec, 0, 2)

	obj := ObjectFile{
		Name: "test.o",
		Sections: map[string]*Section{
			".text": textSec,
		},
		Symbols: NewSymbolTable(),
	}
	obj.Symbols.Add(startSym)

	// Link directly - don't call loadObjectFile separately
	binary, err := l.Link([]ObjectFile{obj}, []string{})
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	if len(binary) < 64 {
		t.Error("Binary too small")
	}
}

func TestLinkerAssignAddressesWithAllSectionTypes(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	textSec := NewCodeSection()
	for i := 0; i < 100; i++ {
		textSec.AddByte(0x90)
	}

	dataSec := NewDataSection()
	for i := 0; i < 50; i++ {
		dataSec.AddByte(0x00)
	}

	rodataSec := NewReadOnlySection()
	rodataSec.AddData([]byte("constant"))

	bssSec := NewBSSSection()
	for i := 0; i < 100; i++ {
		bssSec.AddByte(0x00)
	}

	obj := ObjectFile{
		Name: "test.o",
		Sections: map[string]*Section{
			".text":   textSec,
			".data":   dataSec,
			".rodata": rodataSec,
			".bss":    bssSec,
		},
		Symbols: NewSymbolTable(),
	}

	l.loadObjectFile(obj)
	l.assignAddresses()

	if textSec.Addr == 0 {
		t.Error(".text section should have address")
	}
	if dataSec.Addr == 0 {
		t.Error(".data section should have address")
	}
	if rodataSec.Addr == 0 {
		t.Error(".rodata section should have address")
	}
	if bssSec.Addr == 0 {
		t.Error(".bss section should have address")
	}

	if bssSec.Addr <= dataSec.Addr {
		t.Error(".bss should be after .data")
	}
}

func TestLinkerResolveSymbolsWithEntry(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	textSec := NewCodeSection()
	textSec.AddBytes(0x90, 0x90)

	startSym := NewFunctionSymbol("_start", STB_GLOBAL, textSec, 0, 2)

	obj := ObjectFile{
		Name: "test.o",
		Sections: map[string]*Section{
			".text": textSec,
		},
		Symbols: NewSymbolTable(),
	}
	obj.Symbols.Add(startSym)

	l.loadObjectFile(obj)

	err := l.ResolveSymbols()
	if err != nil {
		t.Errorf("ResolveSymbols failed: %v", err)
	}
}

func TestLinkerResolveSymbolsWithMainFallback(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	textSec := NewCodeSection()
	textSec.AddBytes(0x90, 0x90)

	mainSym := NewFunctionSymbol("main", STB_GLOBAL, textSec, 0, 2)

	obj := ObjectFile{
		Name: "test.o",
		Sections: map[string]*Section{
			".text": textSec,
		},
		Symbols: NewSymbolTable(),
	}
	obj.Symbols.Add(mainSym)

	l.loadObjectFile(obj)

	err := l.ResolveSymbols()
	if err != nil {
		t.Errorf("ResolveSymbols failed: %v", err)
	}

	if l.entry != "main" {
		t.Errorf("Entry point = %s, want main", l.entry)
	}
}

func TestLinkerReadStringOutOfBounds(t *testing.T) {
	result := readString([]byte("hello"), -1)
	if result != "" {
		t.Errorf("readString with negative offset should return empty string, got %q", result)
	}

	result = readString([]byte("hello"), 10)
	if result != "" {
		t.Errorf("readString with out-of-bounds offset should return empty string, got %q", result)
	}

	result = readString([]byte(""), 0)
	if result != "" {
		t.Errorf("readString with empty data should return empty string, got %q", result)
	}
}

func TestLinkerGetSymbolByIndexOutOfBounds(t *testing.T) {
	st := NewSymbolTable()
	sym := NewSymbol("test", STB_GLOBAL, STT_FUNC)
	st.Add(sym)

	result := getSymbolByIndex(st, 999)
	if result != nil {
		t.Error("getSymbolByIndex should return nil for out-of-bounds index")
	}
}

func TestLinkerGetSymbolByIndexValid(t *testing.T) {
	st := NewSymbolTable()
	sym := NewSymbol("test", STB_GLOBAL, STT_FUNC)
	st.Add(sym)

	result := getSymbolByIndex(st, 0)
	if result == nil {
		t.Error("getSymbolByIndex should return symbol for valid index")
	} else if result.Name != "test" {
		t.Errorf("Symbol name = %s, want test", result.Name)
	}
}

func TestLinkerParseAssemblySimple(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	asm := `
	.text
	.globl main
	.type main, @function
main:
	nop
	nop
`

	obj, err := l.parseAssembly(asm)
	if err != nil {
		t.Fatalf("parseAssembly failed: %v", err)
	}

	if obj.Name != "generated.o" {
		t.Errorf("Object name = %s, want generated.o", obj.Name)
	}

	if _, ok := obj.Sections[".text"]; !ok {
		t.Error("Should have .text section")
	}
}

func TestLinkerParseAssemblyWithData(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	asm := `
	.data
	.quad 42
	.string "hello"
	.zero 10
`

	obj, err := l.parseAssembly(asm)
	if err != nil {
		t.Fatalf("parseAssembly failed: %v", err)
	}

	if _, ok := obj.Sections[".data"]; !ok {
		t.Error("Should have .data section")
	}
}

func TestLinkerParseAssemblyWithRodata(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	asm := `
	.section .rodata
	.string "constant"
`

	obj, err := l.parseAssembly(asm)
	if err != nil {
		t.Fatalf("parseAssembly failed: %v", err)
	}

	if _, ok := obj.Sections[".rodata"]; !ok {
		t.Error("Should have .rodata section")
	}
}

func TestLinkerParseAssemblyInvalidQuad(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	asm := `
	.data
	.quad notanumber
`

	obj, err := l.parseAssembly(asm)
	if err != nil {
		t.Fatalf("parseAssembly failed: %v", err)
	}

	if obj.Sections[".data"].Size() != 0 {
		t.Error("Invalid quad should not add data")
	}
}

func TestLinkerParseAssemblyInvalidZero(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	asm := `
	.data
	.zero notanumber
`

	obj, err := l.parseAssembly(asm)
	if err != nil {
		t.Fatalf("parseAssembly failed: %v", err)
	}

	if obj.Sections[".data"].Size() != 0 {
		t.Error("Invalid zero should not add data")
	}
}

func TestLinkerAssembleAndLinkCompileToObject(t *testing.T) {
	if _, err := exec.LookPath("as"); err != nil {
		t.Skip("System assembler not available")
	}

	asmCode := `
	.text
	.globl main
	.type main, @function
main:
	pushq %rbp
	movq %rsp, %rbp
	movl $42, %eax
	popq %rbp
	ret
	.size main, .-main
`

	tmpAsm, err := os.CreateTemp("", "goc-compile-*.s")
	if err != nil {
		t.Skipf("Cannot create temp file: %v", err)
	}
	tmpAsmName := tmpAsm.Name()
	defer os.Remove(tmpAsmName)

	if _, err := tmpAsm.WriteString(asmCode); err != nil {
		tmpAsm.Close()
		t.Skipf("Cannot write temp file: %v", err)
	}
	tmpAsm.Close()

	tmpObj, err := os.CreateTemp("", "goc-compile-*.o")
	if err != nil {
		t.Skipf("Cannot create temp obj: %v", err)
	}
	tmpObjName := tmpObj.Name()
	tmpObj.Close()
	defer os.Remove(tmpObjName)

	l := NewLinker(errhand.NewErrorHandler())
	err = l.AssembleAndLink([]string{tmpAsmName}, tmpObjName, true)
	if err != nil {
		t.Fatalf("AssembleAndLink (compile to object) failed: %v", err)
	}

	data, err := os.ReadFile(tmpObjName)
	if err != nil {
		t.Fatalf("Cannot read output file: %v", err)
	}

	if string(data[0:4]) != "\x7fELF" {
		t.Error("Output is not a valid ELF file")
	}

	typ := binary.LittleEndian.Uint16(data[16:18])
	if typ != ET_REL {
		t.Errorf("Object file type = %d, want %d (ET_REL)", typ, ET_REL)
	}
}

func TestLinkerAssembleAndLinkNoObjectFiles(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	err := l.AssembleAndLink([]string{}, "/tmp/output.o", true)
	if err == nil {
		t.Error("AssembleAndLink should fail with no object files")
	}
}

func TestLinkerAssembleAndLinkInvalidAssembly(t *testing.T) {
	if _, err := exec.LookPath("as"); err != nil {
		t.Skip("System assembler not available")
	}

	asmCode := `
	.this is not valid assembly syntax !!!
`

	tmpAsm, err := os.CreateTemp("", "goc-invalid-*.s")
	if err != nil {
		t.Skipf("Cannot create temp file: %v", err)
	}
	tmpAsmName := tmpAsm.Name()
	defer os.Remove(tmpAsmName)

	if _, err := tmpAsm.WriteString(asmCode); err != nil {
		tmpAsm.Close()
		t.Skipf("Cannot write temp file: %v", err)
	}
	tmpAsm.Close()

	l := NewLinker(errhand.NewErrorHandler())
	err = l.AssembleAndLink([]string{tmpAsmName}, "/tmp/output.exe", false)
	if err == nil {
		t.Error("AssembleAndLink should fail with invalid assembly")
	}
}

func TestLinkerLinkObjectFilesParseError(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	err := l.linkObjectFiles([]string{"/nonexistent/file.o"}, "/tmp/output.exe")
	if err == nil {
		t.Error("linkObjectFiles should fail with non-existent file")
	}
}

func TestLinkerCompileToObjectWriteError(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	err := l.CompileToObject(".text\nnop", "/nonexistent/dir/output.o")
	if err == nil {
		t.Error("CompileToObject should fail with invalid path")
	}
}

func TestLinkerLinkAssemblyWriteError(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	err := l.LinkAssembly(".text\nnop", "/nonexistent/dir/output.exe")
	if err == nil {
		t.Error("LinkAssembly should fail with invalid path")
	}
}

func TestLinkerParseObjectFileNoSymtab(t *testing.T) {
	if _, err := exec.LookPath("as"); err != nil {
		t.Skip("System assembler not available")
	}

	asmCode := `
	.text
	nop
`

	tmpAsm, err := os.CreateTemp("", "goc-nosym-*.s")
	if err != nil {
		t.Skipf("Cannot create temp file: %v", err)
	}
	tmpAsmName := tmpAsm.Name()
	defer os.Remove(tmpAsmName)

	if _, err := tmpAsm.WriteString(asmCode); err != nil {
		tmpAsm.Close()
		t.Skipf("Cannot write temp file: %v", err)
	}
	tmpAsm.Close()

	objFile := tmpAsmName[:len(tmpAsmName)-2] + ".o"
	defer os.Remove(objFile)

	cmd := exec.Command("as", "-o", objFile, tmpAsmName)
	if err := cmd.Run(); err != nil {
		t.Skipf("System assembler not available: %v", err)
	}

	l := NewLinker(errhand.NewErrorHandler())
	obj, err := l.parseObjectFile(objFile)
	if err != nil {
		t.Fatalf("parseObjectFile failed: %v", err)
	}

	if obj == nil {
		t.Error("parseObjectFile should return non-nil object")
	}
}

func TestLinkerEmitWithBSS(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	bssSec := NewBSSSection()
	for i := 0; i < 100; i++ {
		bssSec.AddByte(0x00)
	}
	bssSec.SetAddr(0x403000)

	l.sections = append(l.sections, bssSec)

	binary, err := l.Emit()
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	if len(binary) < 64 {
		t.Error("Binary too small")
	}
}

func TestLinkerCreateProgramHeadersNoSections(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	phs := l.createProgramHeaders()

	if len(phs) < 1 {
		t.Errorf("Expected at least 1 program header, got %d", len(phs))
	}

	if phs[0].Type != PT_NULL {
		t.Error("First program header should be NULL")
	}
}

func TestLinkerCreateSectionHeadersWithSymbols(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	textSec := NewCodeSection()
	textSec.AddBytes(0x90, 0x90)
	textSec.SetAddr(0x401000)

	mainSym := NewFunctionSymbol("main", STB_GLOBAL, textSec, 0x401000, 2)
	l.symbols.Add(mainSym)

	l.sections = append(l.sections, textSec)

	headers, data := l.createSectionHeaders(0)

	if len(headers) < 4 {
		t.Errorf("Expected at least 4 section headers, got %d", len(headers))
	}

	if len(data) == 0 {
		t.Error("Section data should not be empty")
	}
}

func TestLinkerLoadObjectFileWithFileSymbol(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	fileSym := NewFileSymbol("test.c")

	obj := ObjectFile{
		Name: "test.o",
		Sections: map[string]*Section{
			".text": NewCodeSection(),
		},
		Symbols: NewSymbolTable(),
	}
	obj.Symbols.Add(fileSym)

	err := l.loadObjectFile(obj)
	if err != nil {
		t.Errorf("loadObjectFile failed: %v", err)
	}

	if l.symbols.Get("test.c") != nil {
		t.Error("File symbol should not be added to global symbol table")
	}
}

func TestLinkerLoadObjectFileWithUndefinedThenDefined(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	undefSym := NewSymbol("external", STB_GLOBAL, STT_FUNC)

	obj1 := ObjectFile{
		Name: "first.o",
		Sections: map[string]*Section{
			".text": NewCodeSection(),
		},
		Symbols: NewSymbolTable(),
	}
	obj1.Symbols.Add(undefSym)

	textSec := NewCodeSection()
	defSym := NewFunctionSymbol("external", STB_GLOBAL, textSec, 0, 10)

	obj2 := ObjectFile{
		Name: "second.o",
		Sections: map[string]*Section{
			".text": textSec,
		},
		Symbols: NewSymbolTable(),
	}
	obj2.Symbols.Add(defSym)

	l.loadObjectFile(obj1)
	err := l.loadObjectFile(obj2)
	if err != nil {
		t.Errorf("loadObjectFile failed: %v", err)
	}

	sym := l.symbols.Get("external")
	if sym == nil {
		t.Error("Should have external symbol")
	} else if !sym.IsDefined() {
		t.Error("External symbol should be defined")
	}
}

func TestLinkerResolveSymbolsWithOnlyUndefined(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	undefSym := NewSymbol("external", STB_GLOBAL, STT_FUNC)

	obj := ObjectFile{
		Name: "test.o",
		Sections: map[string]*Section{
			".text": NewCodeSection(),
		},
		Symbols: NewSymbolTable(),
	}
	obj.Symbols.Add(undefSym)

	l.loadObjectFile(obj)

	err := l.ResolveSymbols()
	if err == nil {
		t.Error("ResolveSymbols should fail with only undefined symbols")
	}
}

func TestLinkerStringTableAddEmptyString(t *testing.T) {
	st := NewStringTable()

	idx := st.Add("")
	if idx != 0 {
		t.Errorf("Empty string index = %d, want 0", idx)
	}
}

func TestLinkerStringTableAddLongString(t *testing.T) {
	st := NewStringTable()

	longStr := strings.Repeat("a", 1000)
	idx := st.Add(longStr)

	if idx == 0 {
		t.Error("Long string should have non-zero index")
	}

	if st.strings[longStr] != idx {
		t.Error("Long string should be in map")
	}
}

func TestLinkerStringTableDataAfterMultipleAdds(t *testing.T) {
	st := NewStringTable()
	st.Add("a")
	st.Add("b")
	st.Add("c")

	data := st.Data()

	expected := []byte{0, 'a', 0, 'b', 0, 'c', 0}
	if !bytes.Equal(data, expected) {
		t.Errorf("Data = %v, want %v", data, expected)
	}
}

func TestLinkerUnsupportedRelocationType32S(t *testing.T) {
	l := NewLinker(errhand.NewErrorHandler())

	section := NewCodeSection()
	section.AddUint32(0)
	section.SetAddr(0x401000)

	symbol := NewFunctionSymbol("target", STB_GLOBAL, section, 0x402000, 32)

	// Use an actually unsupported relocation type
	rel := &Relocation{
		Offset:  0,
		Type:    999, // Unsupported type
		Symbol:  symbol,
		Addend:  0,
		Section: section,
	}

	obj := ObjectFile{
		Name: "test.o",
		Sections: map[string]*Section{
			".text": section,
		},
		Symbols:     NewSymbolTable(),
		Relocations: []*Relocation{rel},
	}
	obj.Symbols.Add(symbol)

	l.loadObjectFile(obj)
	l.assignAddresses()

	err := l.Relocate()
	if err == nil {
		t.Error("Relocate should fail with unsupported relocation type")
	}
}