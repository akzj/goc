// Package linker - Library loading support for static (.a) and shared (.so) libraries.

package linker

import (
	"bytes"
	"debug/elf"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// loadLibrary loads a library by name, searching library paths.
// The lib parameter should be in the format "name" (for -lname) or "path/to/lib.a" (full path).
func (l *Linker) loadLibrary(lib string) error {
	// Check if already loaded
	if l.loadedLibraries[lib] {
		return nil
	}

	var libPath string
	var err error

	// If lib contains a slash, it's a full path
	if strings.Contains(lib, "/") {
		libPath = lib
	} else {
		// Search for the library
		libPath, err = l.findLibrary(lib)
		if err != nil {
			return err
		}
	}

	// Load based on file extension
	if strings.HasSuffix(libPath, ".a") {
		return l.loadStaticLibrary(libPath)
	} else if strings.HasSuffix(libPath, ".so") {
		return l.loadSharedLibrary(libPath)
	} else {
		return fmt.Errorf("unsupported library format: %s", libPath)
	}
}

// findLibrary searches for a library in library paths.
// Converts library name (e.g., "lua") to filename (e.g., "liblua.a").
func (l *Linker) findLibrary(name string) (string, error) {
	// Try static library first
	libName := fmt.Sprintf("lib%s.a", name)
	
	// Search in -L paths first
	for _, path := range l.libraryPaths {
		libPath := filepath.Join(path, libName)
		if _, err := os.Stat(libPath); err == nil {
			return libPath, nil
		}
	}
	
	// Search in system paths
	systemPaths := []string{"/usr/lib", "/lib", "/usr/local/lib"}
	for _, path := range systemPaths {
		libPath := filepath.Join(path, libName)
		if _, err := os.Stat(libPath); err == nil {
			return libPath, nil
		}
	}
	
	// Try shared library if static not found
	libName = fmt.Sprintf("lib%s.so", name)
	for _, path := range l.libraryPaths {
		libPath := filepath.Join(path, libName)
		if _, err := os.Stat(libPath); err == nil {
			return libPath, nil
		}
	}
	
	for _, path := range systemPaths {
		libPath := filepath.Join(path, libName)
		if _, err := os.Stat(libPath); err == nil {
			return libPath, nil
		}
	}
	
	return "", fmt.Errorf("library not found: %s", name)
}

// loadStaticLibrary loads a static library (.a file).
func (l *Linker) loadStaticLibrary(path string) error {
	// Read the archive file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading archive %s: %w", path, err)
	}
	
	// Parse the AR archive
	archive, err := ParseArArchive(data)
	if err != nil {
		return fmt.Errorf("parsing archive %s: %w", path, err)
	}
	
	// Extract and load object files from the archive
	// For now, load all members (simple approach)
	// A more sophisticated linker would only load members that satisfy undefined symbols
	for _, member := range archive.Members {
		// Skip symbol table and string table members
		if member.Name == "__SYMDEF" || member.Name == "/" || member.Name == "//" {
			continue
		}
		
		// Parse the object file from member data
		obj, err := l.parseObjectData(member.Name, member.Data)
		if err != nil {
			// Skip members that aren't valid object files
			continue
		}
		
		// Load the object file
		if err := l.loadObjectFile(*obj); err != nil {
			return fmt.Errorf("loading object from archive %s: %w", member.Name, err)
		}
	}
	
	// Mark library as loaded
	l.loadedLibraries[path] = true
	return nil
}

// loadSharedLibrary loads a shared library (.so file).
// This is a stub for now - full implementation in Phase 2.
func (l *Linker) loadSharedLibrary(path string) error {
	// Read the ELF file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading shared library %s: %w", path, err)
	}
	
	// Parse as ELF
	elfFile, err := elf.NewFile(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("parsing ELF %s: %w", path, err)
	}
	defer elfFile.Close()
	
	// Verify it's a shared object
	if elfFile.Type != elf.ET_DYN {
		return fmt.Errorf("not a shared object: %s (type: %v)", path, elfFile.Type)
	}
	
	// For now, just mark as loaded
	// Full implementation will load segments and process dynamic symbols
	l.loadedLibraries[path] = true
	return nil
}

// parseObjectData parses object file data into an ObjectFile.
func (l *Linker) parseObjectData(name string, data []byte) (*ObjectFile, error) {
	// Write data to temp file and parse
	tmpFile, err := os.CreateTemp("", "goc-obj-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("writing temp file: %w", err)
	}
	tmpFile.Close()
	
	return l.parseObjectFile(tmpFile.Name())
}

// AddLibraryPath adds a library search path (-L flag).
func (l *Linker) AddLibraryPath(path string) {
	l.libraryPaths = append(l.libraryPaths, path)
}

// GetLibraryPaths returns the current library search paths.
func (l *Linker) GetLibraryPaths() []string {
	return append([]string(nil), l.libraryPaths...)
}
