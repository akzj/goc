// Package linker tests for ELF shared object (.so) parsing and dynamic symbol resolution.
package linker

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// ============================================================================
// Test Helper Functions
// ============================================================================

// createTestSOFile creates a real .so file from C source code for testing.
// Returns the file path and a cleanup function.
func createTestSOFile(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "so_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cSource := `
int test_add(int a, int b) {
    return a + b;
}

int test_multiply(int a, int b) {
    return a * b;
}

int global_counter = 0;

void increment_counter(void) {
    global_counter++;
}
`

	srcFile := filepath.Join(tmpDir, "testlib.c")
	soFile := filepath.Join(tmpDir, "libtest.so")

	if err := os.WriteFile(srcFile, []byte(cSource), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to write C source: %v", err)
	}

	cmd := exec.Command("gcc", "-shared", "-fPIC", "-o", soFile, srcFile)
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to compile .so: %v", err)
	}

	return soFile, func() {
		os.RemoveAll(tmpDir)
	}
}

// readSOFile reads a .so file and returns its contents.
func readSOFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read SO file: %v", err)
	}

	return data
}

// parseELFHeader parses ELF header from data.
func parseELFHeader(t *testing.T, data []byte) *ELFHeader {
	t.Helper()

	header := &ELFHeader{}
	if _, err := header.ReadFrom(bytes.NewReader(data)); err != nil {
		t.Fatalf("Failed to parse ELF header: %v", err)
	}

	return header
}

// parseSectionHeaders parses all section headers from ELF data.
func parseSectionHeaders(t *testing.T, data []byte, header *ELFHeader) []*SectionHeader {
	t.Helper()

	sections := make([]*SectionHeader, header.Shnum)
	for i := uint16(0); i < header.Shnum; i++ {
		offset := header.Shoff + uint64(i)*uint64(header.Shentsize)
		if offset+uint64(header.Shentsize) > uint64(len(data)) {
			t.Fatalf("Section header %d out of bounds", i)
		}

		sh := &SectionHeader{}
		if _, err := sh.ReadFrom(bytes.NewReader(data[offset : offset+uint64(header.Shentsize)])); err != nil {
			t.Fatalf("Failed to parse section header %d: %v", i, err)
		}
		sections[i] = sh
	}

	return sections
}

// getSectionData returns the data for a given section.
func getSectionData(t *testing.T, data []byte, sh *SectionHeader) []byte {
	t.Helper()

	if sh.Type == SHT_NOBITS {
		return make([]byte, sh.Size)
	}

	if sh.Offset+sh.Size > uint64(len(data)) {
		t.Fatalf("Section data out of bounds")
	}

	return data[sh.Offset : sh.Offset+sh.Size]
}

// getStringFromTable retrieves a string from a string table at the given offset.
func getStringFromTable(t *testing.T, strtab []byte, offset uint32) string {
	t.Helper()

	if offset >= uint32(len(strtab)) {
		t.Fatalf("String offset %d out of bounds", offset)
	}

	end := offset
	for end < uint32(len(strtab)) && strtab[end] != 0 {
		end++
	}

	return string(strtab[offset:end])
}

// ============================================================================
// ELF Shared Object Type Detection Tests
// ============================================================================

func TestET_DYNConstant(t *testing.T) {
	if ET_DYN != 3 {
		t.Errorf("ET_DYN = %d, want 3", ET_DYN)
	}
}

func TestNewELFHeaderSharedType(t *testing.T) {
	header := NewELFHeaderShared()

	if header.Type != ET_DYN {
		t.Errorf("NewELFHeaderShared() Type = %d, want %d (ET_DYN)", header.Type, ET_DYN)
	}
}

func TestDetectSharedObjectType(t *testing.T) {
	soFile, cleanup := createTestSOFile(t)
	defer cleanup()

	data := readSOFile(t, soFile)
	header := parseELFHeader(t, data)

	if header.Type != ET_DYN {
		t.Errorf("ELF type = %d, want %d (ET_DYN for shared object)", header.Type, ET_DYN)
	}
}

// ============================================================================
// Dynamic Symbol Table Parsing Tests
// ============================================================================

func TestFindDynSymSection(t *testing.T) {
	soFile, cleanup := createTestSOFile(t)
	defer cleanup()

	data := readSOFile(t, soFile)
	header := parseELFHeader(t, data)
	sections := parseSectionHeaders(t, data, header)

	var dynSymSection *SectionHeader
	for _, sh := range sections {
		if sh.Type == SHT_DYNSYM {
			dynSymSection = sh
			break
		}
	}

	if dynSymSection == nil {
		t.Fatal("Could not find .dynsym section")
	}

	if dynSymSection.Entsize != uint64(Symbol64Size()) {
		t.Errorf("dynsym entry size = %d, want %d", dynSymSection.Entsize, Symbol64Size())
	}
}

func TestParseDynamicSymbols(t *testing.T) {
	soFile, cleanup := createTestSOFile(t)
	defer cleanup()

	data := readSOFile(t, soFile)
	header := parseELFHeader(t, data)
	sections := parseSectionHeaders(t, data, header)

	var dynSymSection, dynStrSection *SectionHeader
	for _, sh := range sections {
		switch sh.Type {
		case SHT_DYNSYM:
			dynSymSection = sh
		case SHT_STRTAB:
			if sh.Flags&SHF_ALLOC != 0 {
				dynStrSection = sh
			}
		}
	}

	if dynSymSection == nil {
		t.Fatal("Could not find .dynsym section")
	}
	if dynStrSection == nil {
		t.Fatal("Could not find .dynstr section")
	}

	dynSymData := getSectionData(t, data, dynSymSection)
	dynStrData := getSectionData(t, data, dynStrSection)

	numSymbols := len(dynSymData) / Symbol64Size()
	if numSymbols < 2 {
		t.Errorf("Expected at least 2 symbols, got %d", numSymbols)
	}

	for i := 1; i < numSymbols && i < 5; i++ {
		offset := i * Symbol64Size()
		sym := &Symbol64{}
		if _, err := sym.ReadFrom(bytes.NewReader(dynSymData[offset : offset+Symbol64Size()])); err != nil {
			t.Fatalf("Failed to parse symbol %d: %v", i, err)
		}

		name := getStringFromTable(t, dynStrData, sym.Name)
		if name == "" {
			continue
		}

		binding := ELF64_ST_BIND(sym.Info)
		symType := ELF64_ST_TYPE(sym.Info)

		t.Logf("Symbol %d: name=%q, binding=%d, type=%d, value=0x%x, size=%d",
			i, name, binding, symType, sym.Value, sym.Size)
	}
}

// ============================================================================
// Dynamic String Table Tests
// ============================================================================

func TestDynStrSectionExists(t *testing.T) {
	soFile, cleanup := createTestSOFile(t)
	defer cleanup()

	data := readSOFile(t, soFile)
	header := parseELFHeader(t, data)
	sections := parseSectionHeaders(t, data, header)

	var found bool
	for _, sh := range sections {
		if sh.Type == SHT_STRTAB && sh.Flags&SHF_ALLOC != 0 {
			found = true
			if sh.Size == 0 {
				t.Error(".dynstr section has zero size")
			}
			break
		}
	}

	if !found {
		t.Error("Could not find .dynstr section")
	}
}

func TestStringTableLookup(t *testing.T) {
	strtab := []byte("\x00hello\x00world\x00test\x00")

	tests := []struct {
		offset uint32
		want   string
	}{
		{1, "hello"},
		{7, "world"},
		{13, "test"},
		{0, ""},
	}

	for _, tt := range tests {
		got := getStringFromTable(t, strtab, tt.offset)
		if got != tt.want {
			t.Errorf("getStringFromTable(offset=%d) = %q, want %q", tt.offset, got, tt.want)
		}
	}
}

// ============================================================================
// Dynamic Section Tests
// ============================================================================

func TestDynamicSectionExists(t *testing.T) {
	soFile, cleanup := createTestSOFile(t)
	defer cleanup()

	data := readSOFile(t, soFile)
	header := parseELFHeader(t, data)
	sections := parseSectionHeaders(t, data, header)

	var found bool
	for _, sh := range sections {
		if sh.Type == SHT_DYNAMIC {
			found = true
			if sh.Entsize != uint64(Dyn64Size()) {
				t.Errorf("dynamic entry size = %d, want %d", sh.Entsize, Dyn64Size())
			}
			break
		}
	}

	if !found {
		t.Error("Could not find .dynamic section")
	}
}

func TestParseDynamicEntries(t *testing.T) {
	soFile, cleanup := createTestSOFile(t)
	defer cleanup()

	data := readSOFile(t, soFile)
	header := parseELFHeader(t, data)
	sections := parseSectionHeaders(t, data, header)

	var dynSection *SectionHeader
	for _, sh := range sections {
		if sh.Type == SHT_DYNAMIC {
			dynSection = sh
			break
		}
	}

	if dynSection == nil {
		t.Skip("No .dynamic section found")
	}

	dynData := getSectionData(t, data, dynSection)
	numEntries := len(dynData) / Dyn64Size()

	t.Logf("Found %d dynamic entries", numEntries)

	var foundNull bool
	for i := 0; i < numEntries; i++ {
		offset := i * Dyn64Size()
		entry := &Dyn64{}
		if _, err := entry.ReadFrom(bytes.NewReader(dynData[offset : offset+Dyn64Size()])); err != nil {
			t.Fatalf("Failed to parse dynamic entry %d: %v", i, err)
		}

		if entry.Tag == DT_NULL {
			foundNull = true
			break
		}

		t.Logf("Dynamic entry %d: tag=%d, val=%d", i, entry.Tag, entry.Val)
	}

	if !foundNull {
		t.Error("Dynamic section does not have DT_NULL terminator")
	}
}

// ============================================================================
// dlsym-like Symbol Lookup Tests
// ============================================================================

// DlsymLookup simulates dlsym by looking up a symbol in the dynamic symbol table.
func DlsymLookup(data []byte, header *ELFHeader, symbolName string) (*Symbol64, error) {
	sections := make([]*SectionHeader, header.Shnum)
	for i := uint16(0); i < header.Shnum; i++ {
		offset := header.Shoff + uint64(i)*uint64(header.Shentsize)
		if offset+uint64(header.Shentsize) > uint64(len(data)) {
			return nil, &ELFError{Message: "section header out of bounds"}
		}

		sh := &SectionHeader{}
		if _, err := sh.ReadFrom(bytes.NewReader(data[offset : offset+uint64(header.Shentsize)])); err != nil {
			return nil, err
		}
		sections[i] = sh
	}

	var dynSymSection, dynStrSection *SectionHeader
	for _, sh := range sections {
		if sh.Type == SHT_DYNSYM {
			dynSymSection = sh
		}
		if sh.Type == SHT_STRTAB && sh.Flags&SHF_ALLOC != 0 {
			dynStrSection = sh
		}
	}

	if dynSymSection == nil {
		return nil, &ELFError{Message: "no .dynsym section"}
	}
	if dynStrSection == nil {
		return nil, &ELFError{Message: "no .dynstr section"}
	}

	dynSymData := data[dynSymSection.Offset : dynSymSection.Offset+dynSymSection.Size]
	dynStrData := data[dynStrSection.Offset : dynStrSection.Offset+dynStrSection.Size]

	numSymbols := len(dynSymData) / Symbol64Size()
	for i := 0; i < numSymbols; i++ {
		offset := i * Symbol64Size()
		sym := &Symbol64{}
		if _, err := sym.ReadFrom(bytes.NewReader(dynSymData[offset : offset+Symbol64Size()])); err != nil {
			continue
		}

		if sym.Name >= uint32(len(dynStrData)) {
			continue
		}
		end := sym.Name
		for end < uint32(len(dynStrData)) && dynStrData[end] != 0 {
			end++
		}
		name := string(dynStrData[sym.Name:end])

		if name == symbolName {
			return sym, nil
		}
	}

	return nil, &ELFError{Message: "symbol not found"}
}

func TestDlsymLookup(t *testing.T) {
	soFile, cleanup := createTestSOFile(t)
	defer cleanup()

	data := readSOFile(t, soFile)
	header := parseELFHeader(t, data)

	sym, err := DlsymLookup(data, header, "test_add")
	if err != nil {
		t.Logf("Lookup 'test_add': %v", err)
	} else {
		t.Logf("Found test_add: value=0x%x, size=%d", sym.Value, sym.Size)
		binding := ELF64_ST_BIND(sym.Info)
		symType := ELF64_ST_TYPE(sym.Info)
		if binding != STB_GLOBAL {
			t.Errorf("test_add binding = %d, want %d (STB_GLOBAL)", binding, STB_GLOBAL)
		}
		if symType != STT_FUNC {
			t.Errorf("test_add type = %d, want %d (STT_FUNC)", symType, STT_FUNC)
		}
	}

	_, err = DlsymLookup(data, header, "nonexistent_symbol")
	if err == nil {
		t.Error("Expected error for non-existent symbol")
	}
}

// ============================================================================
// Segment Loading Tests
// ============================================================================

func TestLoadSegmentParsing(t *testing.T) {
	soFile, cleanup := createTestSOFile(t)
	defer cleanup()

	data := readSOFile(t, soFile)
	header := parseELFHeader(t, data)

	if header.Phnum == 0 {
		t.Fatal("No program headers found")
	}

	var loadSegments int
	for i := uint16(0); i < header.Phnum; i++ {
		offset := header.Phoff + uint64(i)*uint64(header.Phentsize)
		if offset+uint64(header.Phentsize) > uint64(len(data)) {
			t.Fatalf("Program header %d out of bounds", i)
		}

		ph := &ProgramHeader{}
		if _, err := ph.ReadFrom(bytes.NewReader(data[offset : offset+uint64(header.Phentsize)])); err != nil {
			t.Fatalf("Failed to parse program header %d: %v", i, err)
		}

		if ph.Type == PT_LOAD {
			loadSegments++
			t.Logf("LOAD segment: offset=0x%x, vaddr=0x%x, filesz=%d, memsz=%d, flags=0x%x",
				ph.Offset, ph.Vaddr, ph.Filesz, ph.Memsz, ph.Flags)

			if ph.Offset+ph.Filesz > uint64(len(data)) && ph.Filesz > 0 {
				t.Errorf("LOAD segment extends beyond file: offset=%d, filesz=%d, file_size=%d",
					ph.Offset, ph.Filesz, len(data))
			}
		}
	}

	if loadSegments == 0 {
		t.Error("No LOAD segments found in shared object")
	}
}

func TestSegmentFlags(t *testing.T) {
	tests := []struct {
		flags    ProgramFlags
		wantRead bool
		wantWrite bool
		wantExec bool
	}{
		{PF_R, true, false, false},
		{PF_W, false, true, false},
		{PF_X, false, false, true},
		{PF_R | PF_W, true, true, false},
		{PF_R | PF_X, true, false, true},
		{PF_R | PF_W | PF_X, true, true, true},
	}

	for _, tt := range tests {
		gotRead := tt.flags&PF_R != 0
		gotWrite := tt.flags&PF_W != 0
		gotExec := tt.flags&PF_X != 0

		if gotRead != tt.wantRead {
			t.Errorf("Flags 0x%x: read=%v, want %v", tt.flags, gotRead, tt.wantRead)
		}
		if gotWrite != tt.wantWrite {
			t.Errorf("Flags 0x%x: write=%v, want %v", tt.flags, gotWrite, tt.wantWrite)
		}
		if gotExec != tt.wantExec {
			t.Errorf("Flags 0x%x: exec=%v, want %v", tt.flags, gotExec, tt.wantExec)
		}
	}
}

// ============================================================================
// Relocation Processing Tests
// ============================================================================

func TestFindRelocationSections(t *testing.T) {
	soFile, cleanup := createTestSOFile(t)
	defer cleanup()

	data := readSOFile(t, soFile)
	header := parseELFHeader(t, data)
	sections := parseSectionHeaders(t, data, header)

	var relaCount, relCount int
	for _, sh := range sections {
		switch sh.Type {
		case SHT_RELA:
			relaCount++
			t.Logf("Found .rela.* section: offset=0x%x, size=%d, entsize=%d",
				sh.Offset, sh.Size, sh.Entsize)
			if sh.Entsize != uint64(Rela64Size()) {
				t.Errorf("RELA entry size = %d, want %d", sh.Entsize, Rela64Size())
			}
		case SHT_REL:
			relCount++
			t.Logf("Found .rel.* section: offset=0x%x, size=%d", sh.Offset, sh.Size)
			if sh.Entsize != uint64(Rel64Size()) {
				t.Errorf("REL entry size = %d, want %d", sh.Entsize, Rel64Size())
			}
		}
	}

	t.Logf("Found %d RELA sections and %d REL sections", relaCount, relCount)
}

func TestParseRelocationEntries(t *testing.T) {
	soFile, cleanup := createTestSOFile(t)
	defer cleanup()

	data := readSOFile(t, soFile)
	header := parseELFHeader(t, data)
	sections := parseSectionHeaders(t, data, header)

	for i, sh := range sections {
		if sh.Type != SHT_RELA && sh.Type != SHT_REL {
			continue
		}

		if sh.Size == 0 {
			continue
		}

		entrySize := sh.Entsize
		if sh.Type == SHT_REL {
			entrySize = uint64(Rel64Size())
		}

		numRels := sh.Size / entrySize
		t.Logf("Section %d: %d relocation entries", i, numRels)

		if numRels == 0 {
			continue
		}

		for j := uint64(0); j < numRels && j < 3; j++ {
			offset := sh.Offset + j*entrySize
			if sh.Type == SHT_RELA {
				rel := &Rela64{}
				if _, err := rel.ReadFrom(bytes.NewReader(data[offset : offset+entrySize])); err != nil {
					continue
				}
				symIdx := ELF64_R_SYM(rel.Info)
				relType := ELF64_R_TYPE(rel.Info)
				t.Logf("  RELA[%d]: offset=0x%x, sym=%d, type=%d, addend=%d",
					j, rel.Offset, symIdx, relType, rel.Addend)
			} else {
				rel := &Rel64{}
				if _, err := rel.ReadFrom(bytes.NewReader(data[offset : offset+entrySize])); err != nil {
					continue
				}
				symIdx := ELF64_R_SYM(rel.Info)
				relType := ELF64_R_TYPE(rel.Info)
				t.Logf("  REL[%d]: offset=0x%x, sym=%d, type=%d",
					j, rel.Offset, symIdx, relType)
			}
		}
	}
}

func TestRelocationTypeConstants(t *testing.T) {
	if R_X86_64_NONE != 0 {
		t.Errorf("R_X86_64_NONE = %d, want 0", R_X86_64_NONE)
	}
	if R_X86_64_64 != 1 {
		t.Errorf("R_X86_64_64 = %d, want 1", R_X86_64_64)
	}
	if R_X86_64_PC32 != 2 {
		t.Errorf("R_X86_64_PC32 = %d, want 2", R_X86_64_PC32)
	}
	if R_X86_64_PLT32 != 4 {
		t.Errorf("R_X86_64_PLT32 = %d, want 4", R_X86_64_PLT32)
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestIntegrationParseSystemLibrary(t *testing.T) {
	systemLibs := []string{
		"/lib/x86_64-linux-gnu/libc.so.6",
		"/lib/x86_64-linux-gnu/libm.so.6",
		"/lib/x86_64-linux-gnu/libpthread.so.0",
	}

	var tested bool
	for _, libPath := range systemLibs {
		if _, err := os.Stat(libPath); os.IsNotExist(err) {
			continue
		}

		tested = true
		data, err := os.ReadFile(libPath)
		if err != nil {
			t.Logf("Could not read %s: %v", libPath, err)
			continue
		}

		header := &ELFHeader{}
		if _, err := header.ReadFrom(bytes.NewReader(data)); err != nil {
			t.Logf("Could not parse %s: %v", libPath, err)
			continue
		}

		if header.Type != ET_DYN {
			t.Logf("%s: not a shared object (type=%d)", libPath, header.Type)
			continue
		}

		t.Logf("Successfully parsed %s: machine=%d, sections=%d, symbols=%d",
			libPath, header.Machine, header.Shnum, countDynamicSymbols(data, header))
	}

	if !tested {
		t.Skip("No system libraries available for testing")
	}
}

func TestIntegrationFullSOParsing(t *testing.T) {
	soFile, cleanup := createTestSOFile(t)
	defer cleanup()

	data := readSOFile(t, soFile)

	header := parseELFHeader(t, data)
	if header.Type != ET_DYN {
		t.Fatalf("Not a shared object: type=%d", header.Type)
	}

	var loadCount int
	for i := uint16(0); i < header.Phnum; i++ {
		offset := header.Phoff + uint64(i)*uint64(header.Phentsize)
		ph := &ProgramHeader{}
		if _, err := ph.ReadFrom(bytes.NewReader(data[offset : offset+uint64(header.Phentsize)])); err != nil {
			continue
		}
		if ph.Type == PT_LOAD {
			loadCount++
		}
	}

	if loadCount == 0 {
		t.Error("No LOAD segments found")
	}

	sections := parseSectionHeaders(t, data, header)

	var dynSymFound, dynStrFound, dynamicFound bool
	for _, sh := range sections {
		switch sh.Type {
		case SHT_DYNSYM:
			dynSymFound = true
		case SHT_STRTAB:
			if sh.Flags&SHF_ALLOC != 0 {
				dynStrFound = true
			}
		case SHT_DYNAMIC:
			dynamicFound = true
		}
	}

	if !dynSymFound {
		t.Error(".dynsym section not found")
	}
	if !dynStrFound {
		t.Error(".dynstr section not found")
	}
	if !dynamicFound {
		t.Error(".dynamic section not found")
	}

	sym, err := DlsymLookup(data, header, "test_add")
	if err != nil {
		t.Logf("Symbol lookup for 'test_add': %v", err)
	} else {
		t.Logf("Found symbol 'test_add' at 0x%x", sym.Value)
	}

	t.Log("Full SO parsing integration test completed successfully")
}

func countDynamicSymbols(data []byte, header *ELFHeader) int {
	sections := make([]*SectionHeader, header.Shnum)
	for i := uint16(0); i < header.Shnum; i++ {
		offset := header.Shoff + uint64(i)*uint64(header.Shentsize)
		if offset+uint64(header.Shentsize) > uint64(len(data)) {
			return 0
		}
		sh := &SectionHeader{}
		if _, err := sh.ReadFrom(bytes.NewReader(data[offset : offset+uint64(header.Shentsize)])); err != nil {
			return 0
		}
		sections[i] = sh
	}

	for _, sh := range sections {
		if sh.Type == SHT_DYNSYM {
			return int(sh.Size / sh.Entsize)
		}
	}

	return 0
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestParseInvalidSO(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "empty file",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "too short for header",
			data:    []byte{0x7f, 'E', 'L', 'F'},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := &ELFHeader{}
			_, err := header.ReadFrom(bytes.NewReader(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFrom() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDlsymLookupErrors(t *testing.T) {
	data := []byte("invalid")
	header := &ELFHeader{}

	_, err := DlsymLookup(data, header, "test")
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

// ============================================================================
// Shared Object Specific Constant Tests
// ============================================================================

func TestSOSectionTypeConstants(t *testing.T) {
	if SHT_DYNSYM != 11 {
		t.Errorf("SHT_DYNSYM = %d, want 11", SHT_DYNSYM)
	}
	if SHT_DYNAMIC != 6 {
		t.Errorf("SHT_DYNAMIC = %d, want 6", SHT_DYNAMIC)
	}
	if SHT_HASH != 5 {
		t.Errorf("SHT_HASH = %d, want 5", SHT_HASH)
	}
}

func TestSODynamicTagConstants(t *testing.T) {
	if DT_NULL != 0 {
		t.Errorf("DT_NULL = %d, want 0", DT_NULL)
	}
	if DT_NEEDED != 1 {
		t.Errorf("DT_NEEDED = %d, want 1", DT_NEEDED)
	}
	if DT_SYMTAB != 6 {
		t.Errorf("DT_SYMTAB = %d, want 6", DT_SYMTAB)
	}
	if DT_STRTAB != 5 {
		t.Errorf("DT_STRTAB = %d, want 5", DT_STRTAB)
	}
	if DT_STRSZ != 10 {
		t.Errorf("DT_STRSZ = %d, want 10", DT_STRSZ)
	}
	if DT_SYMENT != 11 {
		t.Errorf("DT_SYMENT = %d, want 11", DT_SYMENT)
	}
}