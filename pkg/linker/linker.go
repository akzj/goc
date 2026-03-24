// Package linker links object files into an ELF64 executable.
// This file implements the main linker functionality.
package linker

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/akzj/goc/internal/errhand"
)

// Linker links object files into an executable.
type Linker struct {
	// errs is the error handler.
	errs *errhand.ErrorHandler
	// symbols is the global symbol table.
	symbols *SymbolTable
	// sections is the list of sections.
	sections []*Section
	// objectFiles is the list of loaded object files.
	objectFiles []*ObjectFile
	// unifiedSections maps section names to unified section objects.
	// This ensures all references to .text point to the same section object.
	unifiedSections map[string]*Section
	// entry is the entry point symbol name.
	entry string
	// baseAddr is the base address for the executable.
	baseAddr uint64
	// nextAddr is the next available address.
	nextAddr uint64
	// stringTable is the string table for symbol names.
	stringTable *StringTable
	// sectionStringTable is the string table for section names.
	sectionStringTable *StringTable
	// libraryPaths is the list of library search paths (-L flags).
	libraryPaths []string
	// loadedLibraries tracks which libraries have been loaded.
	loadedLibraries map[string]bool
}

// ObjectFile represents a parsed object file.
type ObjectFile struct {
	// Name is the file name.
	Name string
	// Sections is the map of section name to section.
	Sections map[string]*Section
	// Symbols is the symbol table.
	Symbols *SymbolTable
	// Relocations is the list of relocations.
	Relocations []*Relocation
	// Data is the raw section data.
	Data map[string][]byte
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
	// Section is the section containing the relocation.
	Section *Section
}

// StringTable manages a string table for ELF.
type StringTable struct {
	// strings is the map of string to index.
	strings map[string]uint32
	// data is the raw string table data.
	data []byte
}

// NewLinker creates a new linker.
func NewLinker(errorHandler *errhand.ErrorHandler) *Linker {
	return &Linker{
		errs:               errorHandler,
		symbols:            NewSymbolTable(),
		sections:           make([]*Section, 0),
		objectFiles:        make([]*ObjectFile, 0),
		unifiedSections:    make(map[string]*Section),
		entry:              "_start",
		baseAddr:           0x400000,
		nextAddr:           0x400000,
		stringTable:        NewStringTable(),
		sectionStringTable: NewStringTable(),
		libraryPaths:       make([]string, 0),
		loadedLibraries:    make(map[string]bool),
	}
}

// NewStringTable creates a new string table.
func NewStringTable() *StringTable {
	st := &StringTable{
		strings: make(map[string]uint32),
		data:    make([]byte, 1), // First byte is always null
	}
	st.strings[""] = 0
	return st
}

// Add adds a string to the string table and returns its index.
func (st *StringTable) Add(s string) uint32 {
	if idx, ok := st.strings[s]; ok {
		return idx
	}
	idx := uint32(len(st.data))
	st.data = append(st.data, []byte(s)...)
	st.data = append(st.data, 0) // Null terminator
	st.strings[s] = idx
	return idx
}

// Data returns the raw string table data.
func (st *StringTable) Data() []byte {
	return st.data
}

// Size returns the size of the string table.
func (st *StringTable) Size() uint64 {
	return uint64(len(st.data))
}

// Link links the given object files and libraries.
func (l *Linker) Link(objects []ObjectFile, libs []string) ([]byte, error) {
	// Load all object files
	for _, obj := range objects {
		if err := l.loadObjectFile(obj); err != nil {
			return nil, fmt.Errorf("loading object file %s: %w", obj.Name, err)
		}
	}

	// Load all libraries
	for _, lib := range libs {
		if err := l.loadLibrary(lib); err != nil {
			return nil, fmt.Errorf("loading library %s: %w", lib, err)
		}
	}

	// Resolve symbols
	if err := l.ResolveSymbols(); err != nil {
		return nil, fmt.Errorf("resolving symbols: %w", err)
	}

	// Assign addresses to sections
	l.assignAddresses()

	// Perform relocations
	if err := l.Relocate(); err != nil {
		return nil, fmt.Errorf("performing relocations: %w", err)
	}

	// Emit the final binary
	return l.Emit()
}

// loadObjectFile loads an object file into the linker.
func (l *Linker) loadObjectFile(obj ObjectFile) error {
	l.objectFiles = append(l.objectFiles, &obj)

	// First pass: Create or merge unified sections for all sections in this object
	// Track the offset where each section's data was appended
	sectionOffsets := make(map[string]uint64)
	for sectionName, objSection := range obj.Sections {
		// Check if unified section already exists
		if _, ok := l.unifiedSections[sectionName]; ok {
			// Merge this section's data into the unified section
			offset := l.mergeSectionData(sectionName, objSection)
			sectionOffsets[sectionName] = offset
		} else {
			// Create new unified section
			l.getObjectSection(obj.Name, sectionName, objSection)
			sectionOffsets[sectionName] = 0
		}
	}

	// Second pass: Add symbols to global symbol table with adjusted values
	for _, sym := range obj.Symbols.GetAll() {
		// Skip file symbols
		if sym.Type == STT_FILE {
			continue
		}

		// Update symbol's section reference to point to unified section and adjust value
		if sym.Section != nil {
			if unifiedSection, ok := l.unifiedSections[sym.Section.Name]; ok {
				sym.Section = unifiedSection
				// Adjust symbol value by the offset where this section's data was merged
				sym.Value += sectionOffsets[sym.Section.Name]
			}
		}

		// Check for duplicate definitions
		if existing := l.symbols.Get(sym.Name); existing != nil {
			if existing.IsDefined() && sym.IsDefined() {
				return fmt.Errorf("multiple definition of symbol %s", sym.Name)
			}
			// Use the defined symbol
			if sym.IsDefined() {
				l.symbols.Remove(sym.Name)
				l.symbols.Add(sym)
			}
		} else {
			l.symbols.Add(sym)
		}
	}

	// Third pass: Update relocation section references to point to unified sections
	for _, rel := range obj.Relocations {
		if rel.Section != nil {
			if unifiedSection, ok := l.unifiedSections[rel.Section.Name]; ok {
				rel.Section = unifiedSection
			}
		}
	}

	return nil
}

// getObjectSection finds or creates a unified section by name.
// When multiple object files have sections with the same name, their data is merged.
func (l *Linker) getObjectSection(objName, sectionName string, objSection *Section) (*Section, bool) {
	// Check if we already have a unified section for this name
	if unified, ok := l.unifiedSections[sectionName]; ok {
		return unified, true
	}

	// First occurrence: use the object's section as the canonical unified section
	// so callers/tests holding the same pointer see address updates from assignAddresses.
	l.unifiedSections[sectionName] = objSection
	return objSection, true
}

// mergeSectionData appends data from objSection to the unified section.
// Returns the offset where the new data was appended (for symbol adjustment).
func (l *Linker) mergeSectionData(sectionName string, objSection *Section) uint64 {
	unified, ok := l.unifiedSections[sectionName]
	if !ok {
		return 0
	}
	
	// Calculate offset where new data will be appended
	offset := uint64(len(unified.Data))
	
	// Append the new section data
	if len(objSection.Data) > 0 {
		unified.Data = append(unified.Data, objSection.Data...)
	}
	
	// Update section size
	// Note: Size() returns the size based on symbols, so we need to track it separately
	// For now, we'll rely on the data length
	
	return offset
}

// ResolveSymbols resolves undefined symbols.
func (l *Linker) ResolveSymbols() error {
	// Check for undefined symbols
	undefined := l.symbols.GetUndefined()
	
	// Check if entry point is defined (prefer configured entry, else main for C programs)
	entrySym := l.symbols.Get(l.entry)
	if entrySym == nil || !entrySym.IsDefined() {
		if mainSym := l.symbols.Get("main"); mainSym != nil && mainSym.IsDefined() {
			l.entry = "main"
			entrySym = mainSym
		}
	}

	if entrySym == nil || !entrySym.IsDefined() {
		return fmt.Errorf("undefined entry point: %s", l.entry)
	}

	// For now, report undefined symbols as error
	// In a full linker, we would search libraries
	if len(undefined) > 0 {
		var undefNames []string
		for _, sym := range undefined {
			if sym.Name != l.entry {
				undefNames = append(undefNames, sym.Name)
			}
		}

		if len(undefNames) > 0 {
			return fmt.Errorf("undefined symbols: %s", strings.Join(undefNames, ", "))
		}
	}

	return nil
}

// assignAddresses assigns memory addresses to sections.
func (l *Linker) assignAddresses() {
	// Group sections by type - use unifiedSections, not obj.Sections
	// Symbols point to unifiedSections, so we must use those for address assignment
	textSections := make([]*Section, 0)
	dataSections := make([]*Section, 0)
	rodataSections := make([]*Section, 0)
	bssSections := make([]*Section, 0)

	for _, section := range l.unifiedSections {
		switch section.Name {
		case ".text":
			textSections = append(textSections, section)
		case ".data":
			dataSections = append(dataSections, section)
		case ".rodata":
			rodataSections = append(rodataSections, section)
		case ".bss":
			bssSections = append(bssSections, section)
		}
	}

	// Assign addresses
	// Text segment starts at base address
	addr := l.baseAddr

	// Align to page boundary (0x1000)
	addr = (addr + 0xFFF) &^ 0xFFF

	// Text sections
	for _, section := range textSections {
		oldAddr := section.Addr
		section.SetAddr(addr)
		addrDelta := int64(addr) - int64(oldAddr)

		// Update symbol values for symbols in this section
		for _, sym := range l.symbols.GetAll() {
			if sym.Section != nil && sym.Section == section {
				sym.Value = uint64(int64(sym.Value) + addrDelta)
			}
		}

		addr += section.Size()
		l.sections = append(l.sections, section)
	}

	// Align to page boundary
	addr = (addr + 0xFFF) &^ 0xFFF

	// Read-only data sections
	for _, section := range rodataSections {
		oldAddr := section.Addr
		section.SetAddr(addr)
		addrDelta := int64(addr) - int64(oldAddr)

		// Update symbol values for symbols in this section
		for _, sym := range l.symbols.GetAll() {
			if sym.Section != nil && sym.Section == section {
				sym.Value = uint64(int64(sym.Value) + addrDelta)
			}
		}

		addr += section.Size()
		l.sections = append(l.sections, section)
	}

	// Align to page boundary
	addr = (addr + 0xFFF) &^ 0xFFF

	// Data sections
	for _, section := range dataSections {
		oldAddr := section.Addr
		section.SetAddr(addr)
		addrDelta := int64(addr) - int64(oldAddr)

		// Update symbol values for symbols in this section
		for _, sym := range l.symbols.GetAll() {
			if sym.Section != nil && sym.Section == section {
				sym.Value = uint64(int64(sym.Value) + addrDelta)
			}
		}

		addr += section.Size()
		l.sections = append(l.sections, section)
	}

	// BSS sections (don't occupy file space)
	for _, section := range bssSections {
		oldAddr := section.Addr
		section.SetAddr(addr)
		addrDelta := int64(addr) - int64(oldAddr)

		// Update symbol values for symbols in this section
		for _, sym := range l.symbols.GetAll() {
			if sym.Section != nil && sym.Section == section {
				sym.Value = uint64(int64(sym.Value) + addrDelta)
			}
		}

		addr += section.Size()
		l.sections = append(l.sections, section)
	}

	l.nextAddr = addr
}

// Relocate performs relocations.
func (l *Linker) Relocate() error {
	// Fix: Update relocation symbols to point to global symbol table entries
	// This is necessary because address assignment updates global symbols,
	// but relocations still reference local object file symbols.
	for _, obj := range l.objectFiles {
		for _, rel := range obj.Relocations {
			if rel.Symbol != nil && rel.Symbol.Name != "" {
				// Look up the symbol in the global symbol table
				globalSym := l.symbols.Get(rel.Symbol.Name)
				if globalSym != nil {
					// Update the relocation to use the global symbol
					rel.Symbol = globalSym
				}
			}
		}
	}
	
	// Now apply relocations with correct symbol values
	for _, obj := range l.objectFiles {
		for _, rel := range obj.Relocations {
			if err := l.applyRelocation(rel); err != nil {
				return fmt.Errorf("applying relocation: %w", err)
			}
		}
	}
	return nil
}

// applyRelocation applies a single relocation.
func (l *Linker) applyRelocation(rel *Relocation) error {
	if rel.Symbol == nil {
		return fmt.Errorf("relocation with nil symbol")
	}

	if !rel.Symbol.IsDefined() {
		return fmt.Errorf("relocation against undefined symbol: %s", rel.Symbol.Name)
	}

	symbolValue := rel.Symbol.Value
	sectionAddr := rel.Section.Addr

	switch rel.Type {
	case R_X86_64_64:
		// Direct 64-bit relocation
		value := symbolValue + uint64(rel.Addend)
		l.writeUint64(rel.Section.Data, rel.Offset, value)

	case R_X86_64_PC32, R_X86_64_PLT32:
		// PC-relative 32-bit: (S + A) - P where P is the address of the first byte
		// *after* the 32-bit field (RIP points past the displacement on x86-64).
		p := sectionAddr + rel.Offset + 4
		value := int64(symbolValue) + rel.Addend - int64(p)
		l.writeInt32(rel.Section.Data, rel.Offset, int32(value))

	case R_X86_64_32, R_X86_64_32S:
		// Direct 32-bit relocation
		value := uint32(symbolValue + uint64(rel.Addend))
		l.writeUint32(rel.Section.Data, rel.Offset, value)

	default:
		return fmt.Errorf("unsupported relocation type: %d", rel.Type)
	}

	return nil
}

// writeUint64 writes a uint64 in little-endian format at the given offset.
func (l *Linker) writeUint64(data []byte, offset uint64, value uint64) {
	if int(offset)+8 > len(data) {
		return
	}
	binary.LittleEndian.PutUint64(data[offset:], value)
}

// writeUint32 writes a uint32 in little-endian format at the given offset.
func (l *Linker) writeUint32(data []byte, offset uint64, value uint32) {
	if int(offset)+4 > len(data) {
		return
	}
	binary.LittleEndian.PutUint32(data[offset:], value)
}

// writeInt32 writes an int32 in little-endian format at the given offset.
func (l *Linker) writeInt32(data []byte, offset uint64, value int32) {
	l.writeUint32(data, offset, uint32(value))
}

// Emit emits the final ELF binary.
func (l *Linker) Emit() ([]byte, error) {
	var buf bytes.Buffer

	// Create program headers first to know the count
	programHeaders := l.createProgramHeaders()

	// Create ELF header
	elfHeader := NewELFHeader()
	elfHeader.Phoff = uint64(ELFHeaderSize())
	elfHeader.Shoff = 0 // Will be set later
	elfHeader.Phnum = uint16(len(programHeaders)) // Dynamic based on actual segments
	elfHeader.Shnum = 0 // Will be set later
	elfHeader.Shstrndx = 0 // Will be set later

	// Calculate offsets
	headerSize := uint64(ELFHeaderSize())
	phSize := uint64(len(programHeaders) * ProgramHeaderSize())

	// Section data starts after headers
	sectionDataOffset := headerSize + phSize

	// Create section headers
	sectionHeaders, sectionData := l.createSectionHeaders(sectionDataOffset)

	// Update ELF header
	elfHeader.Shoff = sectionDataOffset + uint64(len(sectionData))
	elfHeader.Shnum = uint16(len(sectionHeaders))
	// Shstrndx is the section header index of .shstrtab (last section)
	elfHeader.Shstrndx = uint16(len(sectionHeaders) - 1)

	// Set entry point address
	entrySym := l.symbols.Get(l.entry)
	if entrySym != nil && entrySym.IsDefined() {
		elfHeader.Entry = entrySym.Value
	}

	// Write ELF header
	if _, err := elfHeader.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("writing ELF header: %w", err)
	}

	// Write program headers
	for _, ph := range programHeaders {
		if _, err := ph.WriteTo(&buf); err != nil {
			return nil, fmt.Errorf("writing program header: %w", err)
		}
	}

	// Write section data
	buf.Write(sectionData)

	// Write section headers
	for _, sh := range sectionHeaders {
		if _, err := sh.WriteTo(&buf); err != nil {
			return nil, fmt.Errorf("writing section header: %w", err)
		}
	}

	return buf.Bytes(), nil
}

// createProgramHeaders creates program headers for the executable.
func (l *Linker) createProgramHeaders() []*ProgramHeader {
	var phs []*ProgramHeader

	// Find text and data segments
	var textStart, textEnd uint64 = 0, 0
	var dataStart, dataEnd uint64 = 0, 0
	var textSize, dataSize uint64 = 0, 0

	for _, section := range l.sections {
		if section.IsExecutable() {
			if textStart == 0 || section.Addr < textStart {
				textStart = section.Addr
			}
			if section.Addr+section.Size() > textEnd {
				textEnd = section.Addr + section.Size()
			}
			textSize += section.Size()
		}
		if section.IsWritable() && !section.IsExecutable() {
			if dataStart == 0 || section.Addr < dataStart {
				dataStart = section.Addr
			}
			if section.Addr+section.Size() > dataEnd {
				dataEnd = section.Addr + section.Size()
			}
			dataSize += section.Size()
		}
	}

	// Build list of non-NULL program headers first
	var loadPHs []*ProgramHeader
	
	// TEXT segment
	var textPH *ProgramHeader
	if textStart > 0 {
		textPH = NewLoadProgramHeader(PF_R | PF_X)
		textPH.Vaddr = textStart
		textPH.Paddr = textStart
		textPH.Filesz = textSize
		textPH.Memsz = textSize
		textPH.Align = 0x1000
		loadPHs = append(loadPHs, textPH)
	}

	// DATA segment - only create if there's actual data
	if dataStart > 0 && dataSize > 0 {
		dataPH := NewLoadProgramHeader(PF_R | PF_W)
		dataPH.Vaddr = dataStart
		dataPH.Paddr = dataStart
		dataPH.Filesz = dataSize
		dataPH.Memsz = dataSize
		dataPH.Align = 0x1000
		loadPHs = append(loadPHs, dataPH)
	}

	// Now calculate offsets based on actual number of headers
	// Offset for first LOAD segment = ELF header + all program headers
	// Note: We add 1 for the NULL program header that's always first
	baseOffset := uint64(ELFHeaderSize() + (len(loadPHs)+1)*ProgramHeaderSize())

	for i, ph := range loadPHs {
		if i == 0 {
			ph.Offset = baseOffset
		} else {
			// Subsequent segments follow the previous one
			prevPH := loadPHs[i-1]
			ph.Offset = prevPH.Offset + prevPH.Filesz
		}
		phs = append(phs, ph)
	}

	// Leading NULL program header (conventional ELF layout)
	nullPH := &ProgramHeader{Type: PT_NULL}
	return append([]*ProgramHeader{nullPH}, phs...)
}

// createSectionHeaders creates section headers and combines section data.
func (l *Linker) createSectionHeaders(sectionDataOffset uint64) ([]*SectionHeader, []byte) {
	var headers []*SectionHeader
	var sectionData bytes.Buffer

	// Add NULL section header
	headers = append(headers, NewSectionHeaderNull())

	// Add section headers for each section
	for _, section := range l.sections {
		nameIdx := l.sectionStringTable.Add(section.Name)

		var sh *SectionHeader
		switch section.Type {
		case SHT_NOBITS:
			sh = NewSectionHeaderNoBits(nameIdx, section.Flags)
		default:
			sh = NewSectionHeaderProgBits(nameIdx, section.Flags)
		}

		sh.Addr = section.Addr
		sh.Offset = sectionDataOffset + uint64(sectionData.Len())
		sh.Size = section.Size()
		sh.Addralign = 0x10 // Default alignment

		// Add section data (skip for NOBITS)
		if section.Type != SHT_NOBITS {
			sectionData.Write(section.Data)
		}

		headers = append(headers, sh)
	}

	// Add symbol table section (offset will be set after writing data)
	symTabIdx := uint32(len(headers))
	symTabSH := NewSectionHeaderSymTab(l.sectionStringTable.Add(".symtab"), 0)
	symTabSH.Link = symTabIdx + 1 // Link to string table
	headers = append(headers, symTabSH)

	// Add string table section (offset will be set after writing data)
	strTabSH := NewSectionHeaderStrTab(l.sectionStringTable.Add(".strtab"))
	headers = append(headers, strTabSH)

	// Add section header string table (offset will be set after writing data)
	shStrTabSH := NewSectionHeaderStrTab(l.sectionStringTable.Add(".shstrtab"))
	headers = append(headers, shStrTabSH)

	// Add symbol and string data to section data
	// First symbol is always NULL
	nullSym := Symbol64{}
	var symBuf bytes.Buffer
	_, _ = nullSym.WriteTo(&symBuf)

	for _, sym := range l.symbols.GetAll() {
		nameIdx := l.stringTable.Add(sym.Name)
		shndx := uint16(SHN_UNDEF)
		if sym.Section != nil {
			// Find section index
			for i, s := range l.sections {
				if s == sym.Section {
					shndx = uint16(i + 1) // +1 for NULL section
					break
				}
			}
		}

		elfSym := NewSymbol64(nameIdx, int(sym.Binding), int(sym.Type), shndx, sym.Value, sym.Size)
		_, _ = elfSym.WriteTo(&symBuf)
	}

	// Write symbol data and record offset for .symtab
	symTabSH.Offset = sectionDataOffset + uint64(sectionData.Len())
	symTabSH.Size = uint64(symBuf.Len())
	sectionData.Write(symBuf.Bytes())

	// Write string table data and record offset for .strtab
	strTabSH.Offset = sectionDataOffset + uint64(sectionData.Len())
	strTabSH.Size = l.stringTable.Size()
	sectionData.Write(l.stringTable.Data())

	// Write section header string table data and record offset for .shstrtab
	shStrTabSH.Offset = sectionDataOffset + uint64(sectionData.Len())
	shStrTabSH.Size = l.sectionStringTable.Size()
	sectionData.Write(l.sectionStringTable.Data())

	return headers, sectionData.Bytes()
}

// AssembleAndLink assembles assembly files and links them into an executable.
func (l *Linker) AssembleAndLink(asmFiles []string, outputFile string, compileToObject bool) error {
	objectFiles := make([]string, 0)

	// Assemble each assembly file to object file
	for _, asmFile := range asmFiles {
		objFile, err := l.assemble(asmFile)
		if err != nil {
			return fmt.Errorf("assembling %s: %w", asmFile, err)
		}
		objectFiles = append(objectFiles, objFile)
		defer os.Remove(objFile) // Clean up temporary object files
	}

	if compileToObject {
		// Just copy the first object file as output
		if len(objectFiles) == 0 {
			return fmt.Errorf("no object files to output")
		}
		data, err := os.ReadFile(objectFiles[0])
		if err != nil {
			return fmt.Errorf("reading object file: %w", err)
		}
		return os.WriteFile(outputFile, data, 0755)
	}

	// Link object files
	return l.linkObjectFiles(objectFiles, outputFile)
}

// assemble assembles an assembly file to an object file using system assembler.
func (l *Linker) assemble(asmFile string) (string, error) {
	objFile := strings.TrimSuffix(asmFile, ".s") + ".o"

	// Use system assembler (as)
	cmd := exec.Command("as", "-o", objFile, asmFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("assembler failed: %w\n%s", err, string(output))
	}

	return objFile, nil
}

// linkObjectFiles links object files into an executable.
func (l *Linker) linkObjectFiles(objFiles []string, outputFile string) error {
	// Parse object files
	objects := make([]ObjectFile, 0)
	for _, objFile := range objFiles {
		obj, err := l.parseObjectFile(objFile)
		if err != nil {
			return fmt.Errorf("parsing object file %s: %w", objFile, err)
		}
		objects = append(objects, *obj)
	}

	// Link objects
	binary, err := l.Link(objects, nil)
	if err != nil {
		return fmt.Errorf("linking: %w", err)
	}

	// Write executable
	if err := os.WriteFile(outputFile, binary, 0755); err != nil {
		return fmt.Errorf("writing executable: %w", err)
	}

	return nil
}

// parseObjectFile parses an ELF object file.
func (l *Linker) parseObjectFile(path string) (*ObjectFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	elfFile, err := elf.NewFile(f)
	if err != nil {
		return nil, fmt.Errorf("parsing ELF: %w", err)
	}
	defer elfFile.Close()

	obj := &ObjectFile{
		Name:     filepath.Base(path),
		Sections: make(map[string]*Section),
		Symbols:  NewSymbolTable(),
		Data:     make(map[string][]byte),
	}

	// Parse sections
	for _, section := range elfFile.Sections {
		if section.Type == elf.SHT_NULL {
			continue
		}

		data, err := section.Data()
		if err != nil {
			continue
		}

		sec := &Section{
			Name:  section.Name,
			Type:  SectionType(section.Type),
			Flags: SectionFlags(section.Flags),
			Data:  data,
		}
		obj.Sections[section.Name] = sec
		obj.Data[section.Name] = data
	}

	// Parse symbols using elfFile.Symbols()
	symbols, err := elfFile.Symbols()
	if err != nil {
		// Try dynamic symbols
		symbols, err = elfFile.DynamicSymbols()
		if err != nil {
			symbols = nil
		}
	}

	if symbols != nil {
		for _, sym := range symbols {
			binding := SymbolBinding(ELF64_ST_BIND(byte(sym.Info)))
			typ := SymbolType(ELF64_ST_TYPE(byte(sym.Info)))

			var section *Section
			if sym.Section < elf.SHN_LORESERVE && int(sym.Section) < len(elfFile.Sections) {
				elfSec := elfFile.Sections[sym.Section]
				section = obj.Sections[elfSec.Name]
			}

			symbol := &Symbol{
				Name:    sym.Name,
				Value:   sym.Value,
				Size:    sym.Size,
				Section: section,
				Binding: binding,
				Type:    typ,
				Defined: sym.Section != elf.SHN_UNDEF,
			}
			obj.Symbols.Add(symbol)
		}
	}

	// Parse relocations
	for _, section := range elfFile.Sections {
		if section.Type != elf.SHT_REL && section.Type != elf.SHT_RELA {
			continue
		}

		// Get target section name from relocation section
		targetSecName := strings.TrimPrefix(section.Name, ".rel")
		targetSecName = strings.TrimPrefix(targetSecName, "a")
		targetSec := obj.Sections[targetSecName]

		// Read raw section data for relocations
		data, err := section.Data()
		if err != nil || len(data) == 0 {
			continue
		}

		isRela := section.Type == elf.SHT_RELA
		entrySize := int(section.Entsize)
		if entrySize == 0 {
			entrySize = 16 // Default for REL
			if isRela {
				entrySize = 24 // RELA entries are 24 bytes
			}
		}

		// Parse relocation entries
		for offset := 0; offset+entrySize <= len(data); offset += entrySize {
			entry := data[offset : offset+entrySize]
			
			var relOffset uint64
			var relInfo uint64
			var addend int64

			if isRela {
				// RELA: offset(8) + info(8) + addend(8)
				relOffset = binary.LittleEndian.Uint64(entry[0:8])
				relInfo = binary.LittleEndian.Uint64(entry[8:16])
				addend = int64(binary.LittleEndian.Uint64(entry[16:24]))
			} else {
				// REL: offset(8) + info(8)
				relOffset = binary.LittleEndian.Uint64(entry[0:8])
				relInfo = binary.LittleEndian.Uint64(entry[8:16])
			}

			symIdx := ELF64_R_SYM(relInfo)
			relType := ELF64_R_TYPE(relInfo)

			// Find symbol by index
			var sym *Symbol
			if symIdx > 0 && int(symIdx) <= len(symbols) {
				symName := symbols[symIdx-1].Name // symbols[0] is NULL
				sym, _ = obj.Symbols.Lookup(symName)
			}

			obj.Relocations = append(obj.Relocations, &Relocation{
				Offset:  relOffset,
				Type:    relType,
				Symbol:  sym,
				Addend:  addend,
				Section: targetSec,
			})
		}
	}

	return obj, nil
}

// getSymbolByIndex retrieves a symbol by its index in the symbol table.
func getSymbolByIndex(st *SymbolTable, idx uint32) *Symbol {
	// This is a simplified implementation
	// In a real linker, we'd maintain symbol indices
	symbols := st.GetAll()
	if int(idx) < len(symbols) {
		return symbols[idx]
	}
	return nil
}

// readString reads a null-terminated string from data starting at offset.
func readString(data []byte, offset int) string {
	if offset < 0 || offset >= len(data) {
		return ""
	}
	end := offset
	for end < len(data) && data[end] != 0 {
		end++
	}
	return string(data[offset:end])
}

// LinkAssembly links assembly code directly to an executable.
func (l *Linker) LinkAssembly(assembly string, outputFile string) error {
	// Write assembly to temporary file
	tmpAsm, err := os.CreateTemp("", "goc-*.s")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpAsmName := tmpAsm.Name()
	defer os.Remove(tmpAsmName)

	if _, err := tmpAsm.WriteString(assembly); err != nil {
		tmpAsm.Close()
		return fmt.Errorf("writing assembly: %w", err)
	}
	tmpAsm.Close()

	// Assemble and link
	return l.AssembleAndLink([]string{tmpAsmName}, outputFile, false)
}

// CompileToObject compiles assembly to an object file.
func (l *Linker) CompileToObject(assembly string, outputFile string) error {
	// Write assembly to temporary file
	tmpAsm, err := os.CreateTemp("", "goc-*.s")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpAsmName := tmpAsm.Name()
	defer os.Remove(tmpAsmName)

	if _, err := tmpAsm.WriteString(assembly); err != nil {
		tmpAsm.Close()
		return fmt.Errorf("writing assembly: %w", err)
	}
	tmpAsm.Close()

	// Assemble to object file
	_, err = l.assemble(tmpAsmName)
	if err != nil {
		return err
	}

	// Move object file to output
	objFile := strings.TrimSuffix(tmpAsmName, ".s") + ".o"
	return os.Rename(objFile, outputFile)
}

// SetEntryPoint sets the entry point symbol.
func (l *Linker) SetEntryPoint(entry string) {
	l.entry = entry
}

// GetEntryPoint returns the entry point symbol.
func (l *Linker) GetEntryPoint() string {
	return l.entry
}

// parseAssembly parses assembly code and creates an ObjectFile.
// This is a simple parser for basic assembly output from our codegen.
func (l *Linker) parseAssembly(assembly string) (*ObjectFile, error) {
	obj := &ObjectFile{
		Name:     "generated.o",
		Sections: make(map[string]*Section),
		Symbols:  NewSymbolTable(),
		Data:     make(map[string][]byte),
	}

	// Create default sections
	textSection := NewCodeSection()
	dataSection := NewDataSection()
	rodataSection := NewReadOnlySection()

	obj.Sections[".text"] = textSection
	obj.Sections[".data"] = dataSection
	obj.Sections[".rodata"] = rodataSection

	// Parse assembly lines
	lines := strings.Split(assembly, "\n")
	currentSection := textSection

	// Regex patterns
	globalRe := regexp.MustCompile(`^\s*\.globl\s+(\w+)`)
	typeRe := regexp.MustCompile(`^\s*\.type\s+(\w+),\s*@(\w+)`)
	dataRe := regexp.MustCompile(`^\s*\.data\s*$`)
	textRe := regexp.MustCompile(`^\s*\.text\s*$`)
	rodataRe := regexp.MustCompile(`^\s*\.section\s+\.rodata\s*$`)
	labelRe := regexp.MustCompile(`^(\w+):`)
	quadRe := regexp.MustCompile(`^\s*\.quad\s+(.+)`)
	stringRe := regexp.MustCompile(`^\s*\.string\s+"(.*)"`)
	zeroRe := regexp.MustCompile(`^\s*\.zero\s+(\d+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle directives
		if matches := globalRe.FindStringSubmatch(line); matches != nil {
			// Global symbol declaration - will be handled by .type or label
			continue
		}

		if matches := typeRe.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			typ := matches[2]

			var symbolType SymbolType
			if typ == "function" {
				symbolType = STT_FUNC
			} else if typ == "object" {
				symbolType = STT_OBJECT
			} else {
				symbolType = STT_NOTYPE
			}

			sym := NewSymbol(name, STB_GLOBAL, symbolType)
			obj.Symbols.Add(sym)
			continue
		}

		if textRe.MatchString(line) {
			currentSection = textSection
			continue
		}

		if dataRe.MatchString(line) {
			currentSection = dataSection
			continue
		}

		if rodataRe.MatchString(line) {
			currentSection = rodataSection
			continue
		}

		// Handle labels (function/data labels)
		if matches := labelRe.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			if sym := obj.Symbols.Get(name); sym != nil {
				// Mark symbol as defined
				sym.SetDefined(currentSection, uint64(currentSection.Size()), 0)
			}
			continue
		}

		// Handle data directives
		if matches := quadRe.FindStringSubmatch(line); matches != nil {
			valueStr := strings.TrimSpace(matches[1])
			value, err := strconv.ParseUint(valueStr, 0, 64)
			if err == nil {
				currentSection.AddUint64(value)
			}
			continue
		}

		if matches := stringRe.FindStringSubmatch(line); matches != nil {
			str := matches[1]
			currentSection.AddData([]byte(str))
			currentSection.AddByte(0) // Null terminator
			continue
		}

		if matches := zeroRe.FindStringSubmatch(line); matches != nil {
			size, err := strconv.ParseUint(matches[1], 10, 64)
			if err == nil {
				for i := uint64(0); i < size; i++ {
					currentSection.AddByte(0)
				}
			}
			continue
		}

		// Handle instructions (simplified - just count bytes)
		// In a real assembler, we'd parse and encode each instruction
		if strings.HasPrefix(line, "\t") {
			// This is an instruction - for now, add placeholder bytes
			// A real implementation would encode the instruction
			currentSection.AddByte(0x90) // NOP as placeholder
		}
	}

	return obj, nil
}