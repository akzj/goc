// Package linker provides tests for linker symbols.
package linker

import (
	"testing"
)

// TestSymbolBindingString tests the String method of SymbolBinding.
func TestSymbolBindingString(t *testing.T) {
	tests := []struct {
		binding  SymbolBinding
		expected string
	}{
		{STB_LOCAL, "LOCAL"},
		{STB_GLOBAL, "GLOBAL"},
		{STB_WEAK, "WEAK"},
		{SymbolBinding(99), "UNKNOWN(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.binding.String()
			if result != tt.expected {
				t.Errorf("SymbolBinding(%d).String() = %q, want %q", tt.binding, result, tt.expected)
			}
		})
	}
}

// TestSymbolTypeString tests the String method of SymbolType.
func TestSymbolTypeString(t *testing.T) {
	tests := []struct {
		typ      SymbolType
		expected string
	}{
		{STT_NOTYPE, "NOTYPE"},
		{STT_OBJECT, "OBJECT"},
		{STT_FUNC, "FUNC"},
		{STT_SECTION, "SECTION"},
		{STT_FILE, "FILE"},
		{SymbolType(99), "UNKNOWN(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.typ.String()
			if result != tt.expected {
				t.Errorf("SymbolType(%d).String() = %q, want %q", tt.typ, result, tt.expected)
			}
		})
	}
}

// TestNewSymbol tests the NewSymbol function.
func TestNewSymbol(t *testing.T) {
	symbol := NewSymbol("test_func", STB_GLOBAL, STT_FUNC)

	if symbol.Name != "test_func" {
		t.Errorf("NewSymbol() Name = %q, want %q", symbol.Name, "test_func")
	}
	if symbol.Binding != STB_GLOBAL {
		t.Errorf("NewSymbol() Binding = %v, want %v", symbol.Binding, STB_GLOBAL)
	}
	if symbol.Type != STT_FUNC {
		t.Errorf("NewSymbol() Type = %v, want %v", symbol.Type, STT_FUNC)
	}
	if symbol.Defined {
		t.Error("NewSymbol() Defined = true, want false")
	}
	if symbol.Value != 0 {
		t.Errorf("NewSymbol() Value = %d, want 0", symbol.Value)
	}
	if symbol.Size != 0 {
		t.Errorf("NewSymbol() Size = %d, want 0", symbol.Size)
	}
	if symbol.Section != nil {
		t.Error("NewSymbol() Section should be nil")
	}
}

// TestNewDefinedSymbol tests the NewDefinedSymbol function.
func TestNewDefinedSymbol(t *testing.T) {
	section := NewSection(".text", SHT_PROGBITS, SHF_ALLOC|SHF_EXECINSTR)
	symbol := NewDefinedSymbol("test_func", STB_GLOBAL, STT_FUNC, section, 0x1000, 64)

	if !symbol.Defined {
		t.Error("NewDefinedSymbol() Defined = false, want true")
	}
	if symbol.Value != 0x1000 {
		t.Errorf("NewDefinedSymbol() Value = 0x%x, want 0x1000", symbol.Value)
	}
	if symbol.Size != 64 {
		t.Errorf("NewDefinedSymbol() Size = %d, want 64", symbol.Size)
	}
	if symbol.Section != section {
		t.Error("NewDefinedSymbol() Section mismatch")
	}
}

// TestNewFunctionSymbol tests the NewFunctionSymbol function.
func TestNewFunctionSymbol(t *testing.T) {
	section := NewCodeSection()
	symbol := NewFunctionSymbol("main", STB_GLOBAL, section, 0x1000, 128)

	if !symbol.IsFunction() {
		t.Error("NewFunctionSymbol() should create a function symbol")
	}
	if !symbol.IsGlobal() {
		t.Error("NewFunctionSymbol() should create a global symbol")
	}
	if !symbol.IsDefined() {
		t.Error("NewFunctionSymbol() should create a defined symbol")
	}
}

// TestNewObjectSymbol tests the NewObjectSymbol function.
func TestNewObjectSymbol(t *testing.T) {
	section := NewDataSection()
	symbol := NewObjectSymbol("global_var", STB_GLOBAL, section, 0x2000, 8)

	if !symbol.IsObject() {
		t.Error("NewObjectSymbol() should create an object symbol")
	}
	if !symbol.IsDefined() {
		t.Error("NewObjectSymbol() should create a defined symbol")
	}
}

// TestNewSectionSymbol tests the NewSectionSymbol function.
func TestNewSectionSymbol(t *testing.T) {
	section := NewCodeSection()
	section.SetAddr(0x1000)
	symbol := NewSectionSymbol(".text", section)

	if symbol.Type != STT_SECTION {
		t.Errorf("NewSectionSymbol() Type = %v, want STT_SECTION", symbol.Type)
	}
	if symbol.Binding != STB_LOCAL {
		t.Errorf("NewSectionSymbol() Binding = %v, want STB_LOCAL", symbol.Binding)
	}
	if symbol.Value != 0x1000 {
		t.Errorf("NewSectionSymbol() Value = 0x%x, want 0x1000", symbol.Value)
	}
}

// TestNewFileSymbol tests the NewFileSymbol function.
func TestNewFileSymbol(t *testing.T) {
	symbol := NewFileSymbol("source.c")

	if symbol.Type != STT_FILE {
		t.Errorf("NewFileSymbol() Type = %v, want STT_FILE", symbol.Type)
	}
	if symbol.Binding != STB_LOCAL {
		t.Errorf("NewFileSymbol() Binding = %v, want STB_LOCAL", symbol.Binding)
	}
	if !symbol.IsDefined() {
		t.Error("NewFileSymbol() should create a defined symbol")
	}
}

// TestSymbolPredicates tests the symbol predicate methods.
func TestSymbolPredicates(t *testing.T) {
	localSym := NewSymbol("local", STB_LOCAL, STT_FUNC)
	globalSym := NewSymbol("global", STB_GLOBAL, STT_FUNC)
	weakSym := NewSymbol("weak", STB_WEAK, STT_FUNC)
	objectSym := NewSymbol("data", STB_GLOBAL, STT_OBJECT)

	tests := []struct {
		symbol   *Symbol
		method   string
		expected bool
	}{
		{localSym, "IsLocal", true},
		{localSym, "IsGlobal", false},
		{localSym, "IsWeak", false},
		{globalSym, "IsLocal", false},
		{globalSym, "IsGlobal", true},
		{globalSym, "IsWeak", false},
		{weakSym, "IsLocal", false},
		{weakSym, "IsGlobal", false},
		{weakSym, "IsWeak", true},
		{localSym, "IsFunction", true},
		{objectSym, "IsFunction", false},
		{objectSym, "IsObject", true},
		{localSym, "IsObject", false},
	}

	for _, tt := range tests {
		t.Run(tt.symbol.Name+"_"+tt.method, func(t *testing.T) {
			var result bool
			switch tt.method {
			case "IsLocal":
				result = tt.symbol.IsLocal()
			case "IsGlobal":
				result = tt.symbol.IsGlobal()
			case "IsWeak":
				result = tt.symbol.IsWeak()
			case "IsFunction":
				result = tt.symbol.IsFunction()
			case "IsObject":
				result = tt.symbol.IsObject()
			default:
				t.Fatalf("Unknown method: %s", tt.method)
			}

			if result != tt.expected {
				t.Errorf("%s.%s() = %v, want %v", tt.symbol.Name, tt.method, result, tt.expected)
			}
		})
	}
}

// TestSymbolDefined tests the symbol defined/undefined methods.
func TestSymbolDefined(t *testing.T) {
	symbol := NewSymbol("test", STB_GLOBAL, STT_FUNC)

	if symbol.IsDefined() {
		t.Error("New symbol should be undefined")
	}
	if !symbol.IsUndefined() {
		t.Error("New symbol should be undefined")
	}

	section := NewCodeSection()
	symbol.SetDefined(section, 0x1000, 64)

	if !symbol.IsDefined() {
		t.Error("Symbol should be defined after SetDefined")
	}
	if symbol.IsUndefined() {
		t.Error("Symbol should not be undefined after SetDefined")
	}
}

// TestSymbolExported tests the IsExported method.
func TestSymbolExported(t *testing.T) {
	localSym := NewSymbol("local", STB_LOCAL, STT_FUNC)
	globalSym := NewSymbol("global", STB_GLOBAL, STT_FUNC)
	weakSym := NewSymbol("weak", STB_WEAK, STT_FUNC)

	if localSym.IsExported() {
		t.Error("Local symbol should not be exported")
	}
	if !globalSym.IsExported() {
		t.Error("Global symbol should be exported")
	}
	if !weakSym.IsExported() {
		t.Error("Weak symbol should be exported")
	}
}

// TestSymbolSetters tests the symbol setter methods.
func TestSymbolSetters(t *testing.T) {
	symbol := NewSymbol("test", STB_GLOBAL, STT_FUNC)
	section := NewCodeSection()

	symbol.SetSection(section)
	if symbol.Section != section {
		t.Error("SetSection failed")
	}

	symbol.SetValue(0x2000)
	if symbol.Value != 0x2000 {
		t.Errorf("SetValue failed: got 0x%x, want 0x2000", symbol.Value)
	}

	symbol.SetSize(128)
	if symbol.Size != 128 {
		t.Errorf("SetSize failed: got %d, want 128", symbol.Size)
	}
}

// TestSymbolString tests the String method of Symbol.
func TestSymbolString(t *testing.T) {
	symbol := NewDefinedSymbol("test_func", STB_GLOBAL, STT_FUNC, NewCodeSection(), 0x1000, 64)
	str := symbol.String()

	expectedParts := []string{"test_func", "GLOBAL", "FUNC", "defined"}
	for _, part := range expectedParts {
		if !contains(str, part) {
			t.Errorf("Symbol.String() = %q, should contain %q", str, part)
		}
	}
}

// TestNewSection tests the NewSection function.
func TestNewSection(t *testing.T) {
	section := NewSection(".text", SHT_PROGBITS, SHF_ALLOC|SHF_EXECINSTR)

	if section.Name != ".text" {
		t.Errorf("NewSection() Name = %q, want %q", section.Name, ".text")
	}
	if section.Type != SHT_PROGBITS {
		t.Errorf("NewSection() Type = %v, want %v", section.Type, SHT_PROGBITS)
	}
	if section.Flags != (SHF_ALLOC|SHF_EXECINSTR) {
		t.Errorf("NewSection() Flags = 0x%x, want 0x%x", section.Flags, SHF_ALLOC|SHF_EXECINSTR)
	}
	if section.Data == nil {
		t.Error("NewSection() Data should not be nil")
	}
	if len(section.Data) != 0 {
		t.Error("NewSection() Data should be empty")
	}
}

// TestSectionHelpers tests the section helper functions.
func TestSectionHelpers(t *testing.T) {
	tests := []struct {
		name     string
		create   func() *Section
		expected string
		flags    SectionFlags
	}{
		{"CodeSection", NewCodeSection, ".text", SHF_ALLOC | SHF_EXECINSTR},
		{"DataSection", NewDataSection, ".data", SHF_ALLOC | SHF_WRITE},
		{"ReadOnlySection", NewReadOnlySection, ".rodata", SHF_ALLOC},
		{"BSSSection", NewBSSSection, ".bss", SHF_ALLOC | SHF_WRITE},
		{"SymTabSection", NewSymTabSection, ".symtab", 0},
		{"StrTabSection", NewStrTabSection, ".strtab", 0},
		{"ShStrTabSection", NewShStrTabSection, ".shstrtab", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			section := tt.create()
			if section.Name != tt.expected {
				t.Errorf("%s Name = %q, want %q", tt.name, section.Name, tt.expected)
			}
			if section.Flags != tt.flags {
				t.Errorf("%s Flags = 0x%x, want 0x%x", tt.name, section.Flags, tt.flags)
			}
		})
	}
}

// TestSectionAddData tests the section data addition methods.
func TestSectionAddData(t *testing.T) {
	section := NewSection(".test", SHT_PROGBITS, 0)

	// Test AddByte
	section.AddByte(0x42)
	if len(section.Data) != 1 || section.Data[0] != 0x42 {
		t.Error("AddByte failed")
	}

	// Test AddBytes
	section.AddBytes(0x01, 0x02, 0x03)
	if len(section.Data) != 4 {
		t.Errorf("AddBytes failed: length = %d, want 4", len(section.Data))
	}

	// Test AddUint16
	section2 := NewSection(".test2", SHT_PROGBITS, 0)
	section2.AddUint16(0x1234)
	if len(section2.Data) != 2 || section2.Data[0] != 0x34 || section2.Data[1] != 0x12 {
		t.Error("AddUint16 failed (little-endian)")
	}

	// Test AddUint32
	section3 := NewSection(".test3", SHT_PROGBITS, 0)
	section3.AddUint32(0x12345678)
	if len(section3.Data) != 4 ||
		section3.Data[0] != 0x78 || section3.Data[1] != 0x56 ||
		section3.Data[2] != 0x34 || section3.Data[3] != 0x12 {
		t.Error("AddUint32 failed (little-endian)")
	}

	// Test AddUint64
	section4 := NewSection(".test4", SHT_PROGBITS, 0)
	section4.AddUint64(0x123456789ABCDEF0)
	if len(section4.Data) != 8 || section4.Data[0] != 0xF0 {
		t.Error("AddUint64 failed (little-endian)")
	}
}

// TestSectionSize tests the Size method of Section.
func TestSectionSize(t *testing.T) {
	section := NewSection(".test", SHT_PROGBITS, 0)

	if section.Size() != 0 {
		t.Errorf("Empty section size = %d, want 0", section.Size())
	}

	section.AddBytes(0x01, 0x02, 0x03, 0x04)
	if section.Size() != 4 {
		t.Errorf("Section size = %d, want 4", section.Size())
	}
}

// TestSectionFlags tests the section flag methods.
func TestSectionFlags(t *testing.T) {
	codeSection := NewCodeSection()
	dataSection := NewDataSection()

	if !codeSection.IsAllocatable() {
		t.Error("Code section should be allocatable")
	}
	if codeSection.IsWritable() {
		t.Error("Code section should not be writable")
	}
	if !codeSection.IsExecutable() {
		t.Error("Code section should be executable")
	}

	if !dataSection.IsAllocatable() {
		t.Error("Data section should be allocatable")
	}
	if !dataSection.IsWritable() {
		t.Error("Data section should be writable")
	}
	if dataSection.IsExecutable() {
		t.Error("Data section should not be executable")
	}
}

// TestSectionSetAddr tests the SetAddr method of Section.
func TestSectionSetAddr(t *testing.T) {
	section := NewSection(".test", SHT_PROGBITS, 0)

	if section.Addr != 0 {
		t.Errorf("Initial address = 0x%x, want 0", section.Addr)
	}

	section.SetAddr(0x1000)
	if section.Addr != 0x1000 {
		t.Errorf("SetAddr failed: got 0x%x, want 0x1000", section.Addr)
	}
}

// TestSectionString tests the String method of Section.
func TestSectionString(t *testing.T) {
	section := NewDefinedSymbol("test", STB_GLOBAL, STT_FUNC, NewCodeSection(), 0x1000, 64).Section
	section.AddBytes(0x01, 0x02, 0x03)
	str := section.String()

	expectedParts := []string{".text", "addr=0x", "size=3"}
	for _, part := range expectedParts {
		if !contains(str, part) {
			t.Errorf("Section.String() = %q, should contain %q", str, part)
		}
	}
}

// TestSymbolTableNew tests the NewSymbolTable function.
func TestSymbolTableNew(t *testing.T) {
	st := NewSymbolTable()

	if st.Count() != 0 {
		t.Errorf("New symbol table count = %d, want 0", st.Count())
	}
	if len(st.GetLocal()) != 0 {
		t.Error("New symbol table should have no local symbols")
	}
	if len(st.GetGlobal()) != 0 {
		t.Error("New symbol table should have no global symbols")
	}
	if len(st.GetUndefined()) != 0 {
		t.Error("New symbol table should have no undefined symbols")
	}
}

// TestSymbolTableAdd tests the Add method of SymbolTable.
func TestSymbolTableAdd(t *testing.T) {
	st := NewSymbolTable()

	// Add local symbol
	localSym := NewSymbol("local_func", STB_LOCAL, STT_FUNC)
	err := st.Add(localSym)
	if err != nil {
		t.Errorf("Add local symbol failed: %v", err)
	}
	if st.Count() != 1 {
		t.Errorf("Symbol table count = %d, want 1", st.Count())
	}
	if len(st.GetLocal()) != 1 {
		t.Errorf("Local symbols count = %d, want 1", len(st.GetLocal()))
	}

	// Add global symbol
	globalSym := NewSymbol("global_func", STB_GLOBAL, STT_FUNC)
	err = st.Add(globalSym)
	if err != nil {
		t.Errorf("Add global symbol failed: %v", err)
	}
	if st.Count() != 2 {
		t.Errorf("Symbol table count = %d, want 2", st.Count())
	}
	if len(st.GetGlobal()) != 1 {
		t.Errorf("Global symbols count = %d, want 1", len(st.GetGlobal()))
	}

	// Add duplicate symbol
	dupSym := NewSymbol("local_func", STB_LOCAL, STT_FUNC)
	err = st.Add(dupSym)
	if err == nil {
		t.Error("Adding duplicate symbol should fail")
	}
}

// TestSymbolTableGet tests the Get and Lookup methods of SymbolTable.
func TestSymbolTableGet(t *testing.T) {
	st := NewSymbolTable()
	symbol := NewSymbol("test_func", STB_GLOBAL, STT_FUNC)
	st.Add(symbol)

	// Test Get
	result := st.Get("test_func")
	if result != symbol {
		t.Error("Get returned wrong symbol")
	}

	result = st.Get("nonexistent")
	if result != nil {
		t.Error("Get should return nil for nonexistent symbol")
	}

	// Test Lookup
	result, found := st.Lookup("test_func")
	if !found || result != symbol {
		t.Error("Lookup failed for existing symbol")
	}

	_, found = st.Lookup("nonexistent")
	if found {
		t.Error("Lookup should return false for nonexistent symbol")
	}
}

// TestSymbolTableRemove tests the Remove method of SymbolTable.
func TestSymbolTableRemove(t *testing.T) {
	st := NewSymbolTable()
	symbol := NewSymbol("test_func", STB_GLOBAL, STT_FUNC)
	st.Add(symbol)

	// Remove existing symbol
	removed := st.Remove("test_func")
	if !removed {
		t.Error("Remove should return true for existing symbol")
	}
	if st.Count() != 0 {
		t.Errorf("Symbol table count after remove = %d, want 0", st.Count())
	}

	// Remove nonexistent symbol
	removed = st.Remove("nonexistent")
	if removed {
		t.Error("Remove should return false for nonexistent symbol")
	}
}

// TestSymbolTableGetAll tests the GetAll method of SymbolTable.
func TestSymbolTableGetAll(t *testing.T) {
	st := NewSymbolTable()
	st.Add(NewSymbol("sym1", STB_LOCAL, STT_FUNC))
	st.Add(NewSymbol("sym2", STB_GLOBAL, STT_FUNC))
	st.Add(NewSymbol("sym3", STB_WEAK, STT_FUNC))

	all := st.GetAll()
	if len(all) != 3 {
		t.Errorf("GetAll returned %d symbols, want 3", len(all))
	}
}

// TestSymbolTableGetByType tests the type-specific getter methods.
func TestSymbolTableGetByType(t *testing.T) {
	st := NewSymbolTable()
	st.Add(NewSymbol("func1", STB_GLOBAL, STT_FUNC))
	st.Add(NewSymbol("func2", STB_LOCAL, STT_FUNC))
	st.Add(NewSymbol("var1", STB_GLOBAL, STT_OBJECT))
	st.Add(NewSymbol("var2", STB_LOCAL, STT_OBJECT))

	funcs := st.GetFunctions()
	if len(funcs) != 2 {
		t.Errorf("GetFunctions returned %d symbols, want 2", len(funcs))
	}

	objects := st.GetObjects()
	if len(objects) != 2 {
		t.Errorf("GetObjects returned %d symbols, want 2", len(objects))
	}
}

// TestSymbolTableDefined tests the GetDefined method.
func TestSymbolTableDefined(t *testing.T) {
	st := NewSymbolTable()

	defined := NewDefinedSymbol("defined", STB_GLOBAL, STT_FUNC, NewCodeSection(), 0x1000, 64)
	undefined := NewSymbol("undefined", STB_GLOBAL, STT_FUNC)

	st.Add(defined)
	st.Add(undefined)

	definedList := st.GetDefined()
	if len(definedList) != 1 {
		t.Errorf("GetDefined returned %d symbols, want 1", len(definedList))
	}
	if definedList[0] != defined {
		t.Error("GetDefined returned wrong symbol")
	}
}

// TestSymbolTableHasUndefined tests the HasUndefined method.
func TestSymbolTableHasUndefined(t *testing.T) {
	st := NewSymbolTable()

	if st.HasUndefined() {
		t.Error("Empty symbol table should not have undefined symbols")
	}

	undefined := NewSymbol("undefined", STB_GLOBAL, STT_FUNC)
	st.Add(undefined)

	if !st.HasUndefined() {
		t.Error("Symbol table with undefined symbol should report HasUndefined = true")
	}

	section := NewCodeSection()
	st.MarkDefined("undefined", section, 0x1000, 64)

	if st.HasUndefined() {
		t.Error("Symbol table after MarkDefined should not have undefined symbols")
	}
}

// TestSymbolTableMarkDefined tests the MarkDefined method.
func TestSymbolTableMarkDefined(t *testing.T) {
	st := NewSymbolTable()
	symbol := NewSymbol("test_func", STB_GLOBAL, STT_FUNC)
	st.Add(symbol)

	section := NewCodeSection()
	err := st.MarkDefined("test_func", section, 0x1000, 64)
	if err != nil {
		t.Errorf("MarkDefined failed: %v", err)
	}

	if !symbol.IsDefined() {
		t.Error("Symbol should be defined after MarkDefined")
	}
	if symbol.Value != 0x1000 {
		t.Errorf("Symbol value = 0x%x, want 0x1000", symbol.Value)
	}
	if symbol.Section != section {
		t.Error("Symbol section mismatch after MarkDefined")
	}

	// Test MarkDefined for nonexistent symbol
	err = st.MarkDefined("nonexistent", section, 0x1000, 64)
	if err == nil {
		t.Error("MarkDefined should fail for nonexistent symbol")
	}
}

// TestSymbolTableString tests the String method of SymbolTable.
func TestSymbolTableString(t *testing.T) {
	st := NewSymbolTable()
	st.Add(NewSymbol("local", STB_LOCAL, STT_FUNC))
	st.Add(NewSymbol("global", STB_GLOBAL, STT_FUNC))

	str := st.String()

	expectedParts := []string{"count=2", "local=1", "global=1"}
	for _, part := range expectedParts {
		if !contains(str, part) {
			t.Errorf("SymbolTable.String() = %q, should contain %q", str, part)
		}
	}
}

// Helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}