// Package linker links object files into an ELF64 executable.
// This file implements AR archive (static library .a) parsing.
package linker

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// AR archive magic header
const (
	ARMagic = "!<arch>\n"
)

// AR file header is 60 bytes
const (
	ARHeaderSize = 60
)

// ArMember represents a member file within an AR archive.
type ArMember struct {
	// Name is the member name (may include trailing / and spaces).
	Name string
	// Timestamp is the file modification timestamp (ASCII decimal).
	Timestamp string
	// OwnerID is the owner ID (ASCII decimal).
	OwnerID string
	// GroupID is the group ID (ASCII decimal).
	GroupID string
	// Mode is the file mode (octal ASCII).
	Mode string
	// Size is the file size in bytes (ASCII decimal).
	Size int64
	// Data is the member file data.
	Data []byte
}

// ArArchive represents a parsed AR archive.
type ArArchive struct {
	// Members is the list of archive members.
	Members []*ArMember
	// SymbolTable is the GNU symbol table (if present).
	SymbolTable []byte
	// LongNames is the long name table (if present).
	LongNames []byte
}

// ParseArArchive parses an AR archive from the given data.
// It validates the magic header and extracts all members.
// Returns an error if the archive is invalid.
func ParseArArchive(data []byte) (*ArArchive, error) {
	if len(data) < len(ARMagic) {
		return nil, fmt.Errorf("invalid AR archive: too short for magic header")
	}

	// Validate magic header
	magic := string(data[:len(ARMagic)])
	if magic != ARMagic {
		return nil, fmt.Errorf("invalid AR archive: magic header mismatch, got %q, want %q", magic, ARMagic)
	}

	archive := &ArArchive{
		Members: make([]*ArMember, 0),
	}

	// Start parsing after magic header
	offset := len(ARMagic)

	for offset < len(data) {
		// Need at least 60 bytes for header
		if offset+ARHeaderSize > len(data) {
			break
		}

		// Parse file header
		header := data[offset : offset+ARHeaderSize]
		member, err := parseArMemberHeader(header, data, offset+ARHeaderSize, archive)
		if err != nil {
			return nil, fmt.Errorf("parsing AR member header at offset %d: %w", offset, err)
		}

		// Handle special members
		if member.Name == "/" || member.Name == "__SYMDEF" {
			// GNU symbol table
			archive.SymbolTable = member.Data
		} else if member.Name == "//" {
			// Long name table
			archive.LongNames = member.Data
		} else {
			// Regular member
			archive.Members = append(archive.Members, member)
		}

		// Move to next member (size rounded up to even boundary)
		dataSize := member.Size
		if dataSize%2 != 0 {
			dataSize++
		}
		offset += ARHeaderSize + int(dataSize)
	}

	return archive, nil
}

// parseArMemberHeader parses a single AR member header.
func parseArMemberHeader(header []byte, data []byte, dataOffset int, archive *ArArchive) (*ArMember, error) {
	if len(header) != ARHeaderSize {
		return nil, fmt.Errorf("invalid header size: got %d, want %d", len(header), ARHeaderSize)
	}

	// Parse header fields
	// Name: 16 bytes (padded with spaces, terminated with /)
	nameBytes := header[0:16]
	name := strings.TrimRight(string(nameBytes), " ")
	// Only remove trailing "/" for regular members, not for GNU special members
	if name != "/" && name != "//" && strings.HasSuffix(name, "/") {
		name = strings.TrimSuffix(name, "/")
	}

	// Handle long names (name starts with / followed by offset)
	if strings.HasPrefix(name, "/") && len(name) > 1 {
		// Long name reference - parse offset
		var nameOffset int
		fmt.Sscanf(name[1:], "%d", &nameOffset)
		if archive.LongNames != nil && nameOffset < len(archive.LongNames) {
			// Find null-terminated name
			end := nameOffset
			for end < len(archive.LongNames) && archive.LongNames[end] != 0 {
				end++
			}
			if end > nameOffset {
				name = string(archive.LongNames[nameOffset:end])
			}
		}
	}

	// Timestamp: 12 bytes
	timestamp := strings.TrimSpace(string(header[16:28]))

	// Owner ID: 6 bytes
	ownerID := strings.TrimSpace(string(header[28:34]))

	// Group ID: 6 bytes
	groupID := strings.TrimSpace(string(header[34:40]))

	// Mode: 8 bytes (octal)
	mode := strings.TrimSpace(string(header[40:48]))

	// Size: 10 bytes (decimal)
	sizeStr := strings.TrimSpace(string(header[48:58]))
	var size int64
	if _, err := fmt.Sscanf(sizeStr, "%d", &size); err != nil {
		return nil, fmt.Errorf("invalid size field %q: %w", sizeStr, err)
	}

	// End marker: 2 bytes (0x60, 0x0A)
	if header[58] != 0x60 || header[59] != 0x0A {
		return nil, fmt.Errorf("invalid header end marker: got 0x%02x 0x%02x, want 0x60 0x0a", header[58], header[59])
	}

	// Read member data
	dataEnd := dataOffset + int(size)
	if dataEnd > len(data) {
		return nil, fmt.Errorf("member data extends beyond archive: need %d bytes, have %d", dataEnd, len(data))
	}

	memberData := make([]byte, size)
	copy(memberData, data[dataOffset:dataOffset+int(size)])

	return &ArMember{
		Name:      name,
		Timestamp: timestamp,
		OwnerID:   ownerID,
		GroupID:   groupID,
		Mode:      mode,
		Size:      size,
		Data:      memberData,
	}, nil
}

// GetMember returns the member with the given name.
// Returns nil if the member is not found.
func (a *ArArchive) GetMember(name string) *ArMember {
	for _, member := range a.Members {
		if member.Name == name {
			return member
		}
	}
	return nil
}

// GetSymbolTable returns the GNU symbol table if present.
// Returns nil if no symbol table exists.
func (a *ArArchive) GetSymbolTable() []byte {
	return a.SymbolTable
}

// HasSymbolTable returns true if the archive contains a GNU symbol table.
func (a *ArArchive) HasSymbolTable() bool {
	return a.SymbolTable != nil && len(a.SymbolTable) > 0
}

// ReadArArchiveFromReader parses an AR archive from an io.Reader.
// It reads all data into memory before parsing.
func ReadArArchiveFromReader(r io.Reader) (*ArArchive, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading AR archive: %w", err)
	}
	return ParseArArchive(data)
}

// ParseArSymbolTable parses the GNU symbol table format.
// The symbol table format is:
// - 4 bytes: number of symbols (big-endian)
// - N * 4 bytes: symbol offsets (big-endian)
// - N null-terminated strings: symbol names
// Returns a list of (offset, name) pairs.
func ParseArSymbolTable(data []byte) ([]struct {
	Offset uint32
	Name   string
}, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("symbol table too short")
	}

	// Read number of symbols (big-endian)
	numSymbols := binary.BigEndian.Uint32(data[0:4])

	if len(data) < 4+int(numSymbols)*4 {
		return nil, fmt.Errorf("symbol table too short for %d symbols", numSymbols)
	}

	// Read symbol offsets
	offsets := make([]uint32, numSymbols)
	for i := uint32(0); i < numSymbols; i++ {
		offsets[i] = binary.BigEndian.Uint32(data[4+i*4 : 4+(i+1)*4])
	}

	// Read symbol names (null-terminated strings after offsets)
	nameStart := 4 + int(numSymbols)*4
	if nameStart > len(data) {
		return nil, fmt.Errorf("symbol table too short for names")
	}

	result := make([]struct {
		Offset uint32
		Name   string
	}, numSymbols)

	nameOffset := nameStart
	for i := uint32(0); i < numSymbols; i++ {
		// Find null terminator
		end := nameOffset
		for end < len(data) && data[end] != 0 {
			end++
		}
		if end <= nameOffset {
			return nil, fmt.Errorf("invalid symbol name at index %d", i)
		}
		result[i].Offset = offsets[i]
		result[i].Name = string(data[nameOffset:end])
		nameOffset = end + 1
	}

	return result, nil
}
