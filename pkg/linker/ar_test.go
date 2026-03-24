// Package linker tests for AR archive parsing.
package linker

import (
	"bytes"
	"fmt"
	"testing"
)

// ============================================================================
// Helper Functions
// ============================================================================

// createArArchive creates a minimal valid AR archive with the given members.
func createArArchive(members []struct {
	name string
	data []byte
}) []byte {
	var buf bytes.Buffer

	// Write magic header
	buf.WriteString(ARMagic)

	// Write each member
	for _, member := range members {
		// Pad name to 16 bytes with spaces, add trailing /
		name := member.name + "/"
		for len(name) < 16 {
			name += " "
		}
		buf.WriteString(name[:16])

		// Timestamp: 12 bytes (padded with spaces)
		buf.WriteString("0           ")

		// Owner ID: 6 bytes
		buf.WriteString("0     ")

		// Group ID: 6 bytes
		buf.WriteString("0     ")

		// Mode: 8 bytes (octal, padded)
		buf.WriteString("100644  ")

		// Size: 10 bytes (decimal, padded)
		sizeBytes := make([]byte, 10)
		for i := 0; i < 9; i++ {
			sizeBytes[i] = ' '
		}
		sizeBytes[9] = '0' + byte(len(member.data)%10)
		buf.Write(sizeBytes[:9])
		buf.WriteByte('0' + byte(len(member.data)%10))

		// End marker: 0x60, 0x0A
		buf.WriteByte(0x60)
		buf.WriteByte(0x0A)

		// Write data
		buf.Write(member.data)

		// Pad to even boundary
		if len(member.data)%2 != 0 {
			buf.WriteByte(0x0A)
		}
	}

	return buf.Bytes()
}

// createArMemberHeader creates a proper AR member header.
func createArMemberHeader(name string, size int64) []byte {
	header := make([]byte, ARHeaderSize)

	// Name: 16 bytes
	nameBytes := []byte(name + "/")
	for len(nameBytes) < 16 {
		nameBytes = append(nameBytes, ' ')
	}
	copy(header[0:16], nameBytes[:16])

	// Timestamp: 12 bytes
	copy(header[16:28], "0           ")

	// Owner ID: 6 bytes
	copy(header[28:34], "0     ")

	// Group ID: 6 bytes
	copy(header[34:40], "0     ")

	// Mode: 8 bytes
	copy(header[40:48], "100644  ")

	// Size: 10 bytes
	sizeStr := make([]byte, 10)
	for i := 0; i < 10; i++ {
		sizeStr[i] = ' '
	}
	// Write size as decimal string
	sizeBytes := []byte(string(rune('0')))
	copy(sizeStr[10-len(sizeBytes):], sizeBytes)

	// End marker
	header[58] = 0x60
	header[59] = 0x0A

	return header
}

// ============================================================================
// ParseArArchive Tests
// ============================================================================

func TestParseArArchiveValidMagic(t *testing.T) {
	// Create minimal valid archive
	data := []byte(ARMagic)

	archive, err := ParseArArchive(data)
	if err != nil {
		t.Fatalf("ParseArArchive() error = %v, want nil", err)
	}

	if archive == nil {
		t.Fatal("ParseArArchive() returned nil archive")
	}

	if len(archive.Members) != 0 {
		t.Errorf("Members count = %d, want 0", len(archive.Members))
	}
}

func TestParseArArchiveInvalidMagic(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "too short",
			data:    []byte("!<ar"),
			wantErr: true,
		},
		{
			name:    "wrong magic",
			data:    []byte("!<wrong>\n"),
			wantErr: true,
		},
		{
			name:    "empty",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "similar but wrong",
			data:    []byte("!<arch>"), // missing newline
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseArArchive(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseArArchive() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseArArchiveEmptyArchive(t *testing.T) {
	data := []byte(ARMagic)

	archive, err := ParseArArchive(data)
	if err != nil {
		t.Fatalf("ParseArArchive() error = %v, want nil", err)
	}

	if len(archive.Members) != 0 {
		t.Errorf("Members = %d, want 0", len(archive.Members))
	}

	if archive.SymbolTable != nil {
		t.Error("SymbolTable should be nil for empty archive")
	}

	if archive.LongNames != nil {
		t.Error("LongNames should be nil for empty archive")
	}
}

func TestParseArArchiveWithMembers(t *testing.T) {
	// Create archive with members using proper header format
	var buf bytes.Buffer
	buf.WriteString(ARMagic)

	// Add a member "test.o" with data "hello"
	memberName := "test.o"
	memberData := []byte("hello")

	// Create proper header
	header := make([]byte, ARHeaderSize)
	nameField := memberName + "/"
	copy(header[0:], nameField)
	for i := len(nameField); i < 16; i++ {
		header[i] = ' '
	}
	copy(header[16:28], "0           ") // timestamp
	copy(header[28:34], "0     ")       // owner
	copy(header[34:40], "0     ")       // group
	copy(header[40:48], "100644  ")     // mode
	copy(header[48:58], "         5")   // size (5 bytes)
	header[58] = 0x60
	header[59] = 0x0A

	buf.Write(header)
	buf.Write(memberData)
	// Pad to even boundary (5 is odd, so add 1 byte)
	buf.WriteByte(0x0A)

	data := buf.Bytes()

	archive, err := ParseArArchive(data)
	if err != nil {
		t.Fatalf("ParseArArchive() error = %v, want nil", err)
	}

	if len(archive.Members) != 1 {
		t.Fatalf("Members count = %d, want 1", len(archive.Members))
	}

	member := archive.Members[0]
	if member.Name != "test.o" {
		t.Errorf("Member.Name = %q, want %q", member.Name, "test.o")
	}

	if member.Size != 5 {
		t.Errorf("Member.Size = %d, want 5", member.Size)
	}

	if string(member.Data) != "hello" {
		t.Errorf("Member.Data = %q, want %q", string(member.Data), "hello")
	}
}

func TestParseArArchiveGnuSymbolTable(t *testing.T) {
	// Create archive with GNU symbol table member "/"
	var buf bytes.Buffer
	buf.WriteString(ARMagic)

	// Symbol table member "/"
	symData := []byte{0, 0, 0, 0} // 0 symbols (big-endian)

	header := make([]byte, ARHeaderSize)
	header[0] = '/'
	header[1] = ' '
	for i := 2; i < 16; i++ {
		header[i] = ' '
	}
	copy(header[16:28], "0           ")
	copy(header[28:34], "0     ")
	copy(header[34:40], "0     ")
	copy(header[40:48], "100644  ")
	copy(header[48:58], "         4") // size 4
	header[58] = 0x60
	header[59] = 0x0A

	buf.Write(header)
	buf.Write(symData)

	data := buf.Bytes()

	archive, err := ParseArArchive(data)
	if err != nil {
		t.Fatalf("ParseArArchive() error = %v, want nil", err)
	}

	if archive.SymbolTable == nil {
		t.Fatal("SymbolTable should not be nil")
	}

	if len(archive.SymbolTable) != 4 {
		t.Errorf("SymbolTable length = %d, want 4", len(archive.SymbolTable))
	}

	// HasSymbolTable returns true because SymbolTable is not nil and has data
	if !archive.HasSymbolTable() {
		t.Error("HasSymbolTable should return true when symbol table exists")
	}
}

func TestParseArArchiveSymDefMember(t *testing.T) {
	// Create archive with __SYMDEF member (BSD style)
	var buf bytes.Buffer
	buf.WriteString(ARMagic)

	// __SYMDEF member
	symData := []byte("symbol data")

	header := make([]byte, ARHeaderSize)
	nameField := "__SYMDEF/"
	copy(header[0:], nameField)
	for i := len(nameField); i < 16; i++ {
		header[i] = ' '
	}
	copy(header[16:28], "0           ")
	copy(header[28:34], "0     ")
	copy(header[34:40], "0     ")
	copy(header[40:48], "100644  ")
	copy(header[48:58], "        11") // size 11
	header[58] = 0x60
	header[59] = 0x0A

	buf.Write(header)
	buf.Write(symData)

	data := buf.Bytes()

	archive, err := ParseArArchive(data)
	if err != nil {
		t.Fatalf("ParseArArchive() error = %v, want nil", err)
	}

	if archive.SymbolTable == nil {
		t.Fatal("SymbolTable should not be nil for __SYMDEF member")
	}
}

func TestParseArArchiveLongNames(t *testing.T) {
	// Create archive with long name table "//"
	var buf bytes.Buffer
	buf.WriteString(ARMagic)

	// Long names member "//"
	longNameData := []byte("very_long_object_file.o/\x00")

	header := make([]byte, ARHeaderSize)
	header[0] = '/'
	header[1] = '/'
	for i := 2; i < 16; i++ {
		header[i] = ' '
	}
	copy(header[16:28], "0           ")
	copy(header[28:34], "0     ")
	copy(header[34:40], "0     ")
	copy(header[40:48], "100644  ")
	copy(header[48:58], "        26") // size 26
	header[58] = 0x60
	header[59] = 0x0A

	buf.Write(header)
	buf.Write(longNameData)
	// Pad to even boundary
	buf.WriteByte(0x0A)

	data := buf.Bytes()

	archive, err := ParseArArchive(data)
	if err != nil {
		t.Fatalf("ParseArArchive() error = %v, want nil", err)
	}

	if archive.LongNames == nil {
		t.Fatal("LongNames should not be nil")
	}
}

// ============================================================================
// ArMember Tests
// ============================================================================

func TestArMemberFields(t *testing.T) {
	member := &ArMember{
		Name:      "test.o",
		Timestamp: "1234567890",
		OwnerID:   "1000",
		GroupID:   "1000",
		Mode:      "100644",
		Size:      1024,
		Data:      []byte("test data"),
	}

	if member.Name != "test.o" {
		t.Errorf("Name = %q, want %q", member.Name, "test.o")
	}

	if member.Size != 1024 {
		t.Errorf("Size = %d, want 1024", member.Size)
	}

	if len(member.Data) != 9 {
		t.Errorf("Data length = %d, want 9", len(member.Data))
	}
}

// ============================================================================
// ArArchive Methods Tests
// ============================================================================

func TestArArchiveGetMember(t *testing.T) {
	archive := &ArArchive{
		Members: []*ArMember{
			{Name: "file1.o", Data: []byte("data1")},
			{Name: "file2.o", Data: []byte("data2")},
			{Name: "file3.o", Data: []byte("data3")},
		},
	}

	tests := []struct {
		name     string
		query    string
		wantName string
		wantNil  bool
	}{
		{"found first", "file1.o", "file1.o", false},
		{"found middle", "file2.o", "file2.o", false},
		{"found last", "file3.o", "file3.o", false},
		{"not found", "file4.o", "", true},
		{"empty query", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member := archive.GetMember(tt.query)
			if tt.wantNil {
				if member != nil {
					t.Errorf("GetMember(%q) = %v, want nil", tt.query, member)
				}
			} else {
				if member == nil {
					t.Fatalf("GetMember(%q) returned nil", tt.query)
				}
				if member.Name != tt.wantName {
					t.Errorf("GetMember(%q).Name = %q, want %q", tt.query, member.Name, tt.wantName)
				}
			}
		})
	}
}

func TestArArchiveGetSymbolTable(t *testing.T) {
	symData := []byte{0, 0, 0, 1, 0, 0, 0, 0}

	archive := &ArArchive{
		Members:     []*ArMember{},
		SymbolTable: symData,
	}

	result := archive.GetSymbolTable()
	if result == nil {
		t.Fatal("GetSymbolTable() returned nil")
	}

	if !bytes.Equal(result, symData) {
		t.Error("GetSymbolTable() returned different data")
	}

	// Test nil symbol table
	archive2 := &ArArchive{
		Members:     []*ArMember{},
		SymbolTable: nil,
	}

	result2 := archive2.GetSymbolTable()
	if result2 != nil {
		t.Error("GetSymbolTable() should return nil when no symbol table")
	}
}

func TestArArchiveHasSymbolTable(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"nil", nil, false},
		{"empty", []byte{}, false},
		{"valid", []byte{0, 0, 0, 1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive := &ArArchive{
				SymbolTable: tt.data,
			}
			if got := archive.HasSymbolTable(); got != tt.want {
				t.Errorf("HasSymbolTable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// ParseArSymbolTable Tests
// ============================================================================

func TestParseArSymbolTableEmpty(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{"empty", []byte{}, true},
		{"too short", []byte{0, 0, 0}, true},
		{"zero symbols", []byte{0, 0, 0, 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseArSymbolTable(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseArSymbolTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseArSymbolTableWithSymbols(t *testing.T) {
	// Create a symbol table with 2 symbols
	// Format: [count:4][offset1:4][offset2:4][name1\0][name2\0]
	data := []byte{
		0, 0, 0, 2, // 2 symbols (big-endian)
		0, 0, 0, 10, // offset 10
		0, 0, 0, 20, // offset 20
		's', 'y', 'm', '1', 0, // "sym1"
		's', 'y', 'm', '2', 0, // "sym2"
	}

	symbols, err := ParseArSymbolTable(data)
	if err != nil {
		t.Fatalf("ParseArSymbolTable() error = %v, want nil", err)
	}

	if len(symbols) != 2 {
		t.Fatalf("Got %d symbols, want 2", len(symbols))
	}

	if symbols[0].Offset != 10 {
		t.Errorf("Symbol[0].Offset = %d, want 10", symbols[0].Offset)
	}
	if symbols[0].Name != "sym1" {
		t.Errorf("Symbol[0].Name = %q, want %q", symbols[0].Name, "sym1")
	}

	if symbols[1].Offset != 20 {
		t.Errorf("Symbol[1].Offset = %d, want 20", symbols[1].Offset)
	}
	if symbols[1].Name != "sym2" {
		t.Errorf("Symbol[1].Name = %q, want %q", symbols[1].Name, "sym2")
	}
}

func TestParseArSymbolTableInvalidData(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "count but no offsets",
			data:    []byte{0, 0, 0, 1}, // 1 symbol but no offset
			wantErr: true,
		},
		{
			name:    "offsets but no names",
			data:    []byte{0, 0, 0, 1, 0, 0, 0, 10}, // 1 symbol, offset, but no name data
			wantErr: true,
		},
		{
			name:    "empty name",
			data:    []byte{0, 0, 0, 1, 0, 0, 0, 8, 0}, // 1 symbol, offset 8, but name is empty
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseArSymbolTable(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseArSymbolTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// ReadArArchiveFromReader Tests
// ============================================================================

func TestReadArArchiveFromReader(t *testing.T) {
	data := []byte(ARMagic)
	reader := bytes.NewReader(data)

	archive, err := ReadArArchiveFromReader(reader)
	if err != nil {
		t.Fatalf("ReadArArchiveFromReader() error = %v, want nil", err)
	}

	if archive == nil {
		t.Fatal("ReadArArchiveFromReader() returned nil")
	}
}

func TestReadArArchiveFromReaderError(t *testing.T) {
	// Create a reader that will fail
	reader := &errorReader{}

	_, err := ReadArArchiveFromReader(reader)
	if err == nil {
		t.Error("ReadArArchiveFromReader() should return error for failing reader")
	}
}

// errorReader is a test helper that always returns an error.
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, bytes.ErrTooLarge
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestParseArArchiveMultipleMembers(t *testing.T) {
	// Create archive with multiple members
	var buf bytes.Buffer
	buf.WriteString(ARMagic)

	members := []struct {
		name string
		data string
	}{
		{"file1.o", "content1"},
		{"file2.o", "content2"},
		{"file3.o", "content3"},
	}

	for _, m := range members {
		header := make([]byte, ARHeaderSize)
		nameField := m.name + "/"
		copy(header[0:], nameField)
		for i := len(nameField); i < 16; i++ {
			header[i] = ' '
		}
		copy(header[16:28], "0           ")
		copy(header[28:34], "0     ")
		copy(header[34:40], "0     ")
		copy(header[40:48], "100644  ")
		sizeStr := make([]byte, 10)
		for i := 0; i < 10; i++ {
			sizeStr[i] = ' '
		}
		copy(header[48:58], sizeStr)
		// Write size properly
		sizeBytes := []byte(fmt.Sprintf("%10d", len(m.data)))
		copy(header[48:58], sizeBytes)
		header[58] = 0x60
		header[59] = 0x0A

		buf.Write(header)
		buf.Write([]byte(m.data))
		if len(m.data)%2 != 0 {
			buf.WriteByte(0x0A)
		}
	}

	data := buf.Bytes()

	archive, err := ParseArArchive(data)
	if err != nil {
		t.Fatalf("ParseArArchive() error = %v, want nil", err)
	}

	if len(archive.Members) != len(members) {
		t.Fatalf("Members count = %d, want %d", len(archive.Members), len(members))
	}

	for i, m := range members {
		if archive.Members[i].Name != m.name {
			t.Errorf("Member[%d].Name = %q, want %q", i, archive.Members[i].Name, m.name)
		}
		if string(archive.Members[i].Data) != m.data {
			t.Errorf("Member[%d].Data = %q, want %q", i, archive.Members[i].Data, m.data)
		}
	}
}

func TestParseArArchiveInvalidHeaderEndMarker(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString(ARMagic)

	// Create header with invalid end marker
	header := make([]byte, ARHeaderSize)
	copy(header[0:16], "test.o/         ")
	copy(header[16:28], "0           ")
	copy(header[28:34], "0     ")
	copy(header[34:40], "0     ")
	copy(header[40:48], "100644  ")
	copy(header[48:58], "         5")
	// Invalid end marker
	header[58] = 0x00
	header[59] = 0x00

	buf.Write(header)
	buf.Write([]byte("hello"))

	data := buf.Bytes()

	_, err := ParseArArchive(data)
	if err == nil {
		t.Error("ParseArArchive() should return error for invalid end marker")
	}
}

func TestParseArArchiveInvalidSize(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString(ARMagic)

	// Create header with invalid size field
	header := make([]byte, ARHeaderSize)
	copy(header[0:16], "test.o/         ")
	copy(header[16:28], "0           ")
	copy(header[28:34], "0     ")
	copy(header[34:40], "0     ")
	copy(header[40:48], "100644  ")
	// Invalid size (non-numeric)
	copy(header[48:58], "invalid!!!")
	header[58] = 0x60
	header[59] = 0x0A

	buf.Write(header)
	buf.Write([]byte("hello"))

	data := buf.Bytes()

	_, err := ParseArArchive(data)
	if err == nil {
		t.Error("ParseArArchive() should return error for invalid size field")
	}
}
