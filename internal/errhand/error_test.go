// Package errhand provides error handling and diagnostic reporting for the GOC compiler.
// This file contains unit tests for error types.
package errhand

import (
	"strings"
	"testing"
)

// TestErrorLevel_String tests the String() method for ErrorLevel.
func TestErrorLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    ErrorLevel
		expected string
	}{
		{"ERROR", ERROR, "ERROR"},
		{"WARNING", WARNING, "WARNING"},
		{"NOTE", NOTE, "NOTE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("ErrorLevel.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestErrorLevel_String_Unknown tests unknown error levels.
func TestErrorLevel_String_Unknown(t *testing.T) {
	// Test an unknown level (out of defined range)
	unknown := ErrorLevel(999)
	result := unknown.String()
	if result != "UNKNOWN" {
		t.Errorf("ErrorLevel(999).String() = %q, want %q", result, "UNKNOWN")
	}
}

// TestError_String tests the String() method for Error.
func TestError_String(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "basic error",
			err: &Error{
				Level:    ERROR,
				Code:     "E1001",
				Message:  "syntax error",
				Position: Position{File: "main.c", Line: 10, Column: 5},
				Hint:     "",
				Related:  nil,
			},
			expected: "main.c:10:5: ERROR: E1001: syntax error",
		},
		{
			name: "error with hint",
			err: &Error{
				Level:    ERROR,
				Code:     "E2001",
				Message:  "undefined symbol",
				Position: Position{File: "test.c", Line: 20, Column: 15},
				Hint:     "did you mean to declare the variable?",
				Related:  nil,
			},
			expected: "test.c:20:15: ERROR: E2001: undefined symbol\n  hint: did you mean to declare the variable?",
		},
		{
			name: "warning",
			err: &Error{
				Level:    WARNING,
				Code:     "E0001",
				Message:  "unused variable",
				Position: Position{File: "main.c", Line: 5, Column: 1},
				Hint:     "",
				Related:  nil,
			},
			expected: "main.c:5:1: WARNING: E0001: unused variable",
		},
		{
			name: "note",
			err: &Error{
				Level:    NOTE,
				Code:     "",
				Message:  "previous declaration here",
				Position: Position{File: "main.c", Line: 3, Column: 1},
				Hint:     "",
				Related:  nil,
			},
			expected: "main.c:3:1: NOTE: : previous declaration here",
		},
		{
			name: "error with related info",
			err: &Error{
				Level:    ERROR,
				Code:     "E2002",
				Message:  "duplicate symbol",
				Position: Position{File: "main.c", Line: 10, Column: 5},
				Hint:     "",
				Related: []RelatedInfo{
					{
						Position: Position{File: "main.c", Line: 5, Column: 5},
						Message:  "first declaration here",
					},
				},
			},
			expected: "main.c:10:5: ERROR: E2002: duplicate symbol\n  note: main.c:5:5: first declaration here",
		},
		{
			name: "error with multiple related info",
			err: &Error{
				Level:    ERROR,
				Code:     "E2003",
				Message:  "type mismatch",
				Position: Position{File: "main.c", Line: 15, Column: 10},
				Hint:     "check variable types",
				Related: []RelatedInfo{
					{
						Position: Position{File: "main.c", Line: 5, Column: 5},
						Message:  "expected type int",
					},
					{
						Position: Position{File: "main.c", Line: 15, Column: 10},
						Message:  "got type float",
					},
				},
			},
			expected: "main.c:15:10: ERROR: E2003: type mismatch\n  hint: check variable types\n  note: main.c:5:5: expected type int\n  note: main.c:15:10: got type float",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.String()
			if result != tt.expected {
				t.Errorf("Error.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestError_Error tests the Error() method implements error interface.
func TestError_Error(t *testing.T) {
	err := &Error{
		Level:    ERROR,
		Code:     "E1001",
		Message:  "syntax error",
		Position: Position{File: "main.c", Line: 10, Column: 5},
		Hint:     "",
		Related:  nil,
	}

	// Error() should return the same as String()
	errorResult := err.Error()
	stringResult := err.String()

	if errorResult != stringResult {
		t.Errorf("Error.Error() = %q, Error.String() = %q, they should be equal", errorResult, stringResult)
	}

	// Verify it contains expected components
	if !strings.Contains(errorResult, "main.c:10:5") {
		t.Errorf("Error.Error() should contain position, got %q", errorResult)
	}
	if !strings.Contains(errorResult, "ERROR") {
		t.Errorf("Error.Error() should contain level, got %q", errorResult)
	}
	if !strings.Contains(errorResult, "E1001") {
		t.Errorf("Error.Error() should contain error code, got %q", errorResult)
	}
	if !strings.Contains(errorResult, "syntax error") {
		t.Errorf("Error.Error() should contain message, got %q", errorResult)
	}
}

// TestError_Error_Interface verifies Error implements the error interface.
func TestError_Error_Interface(t *testing.T) {
	err := &Error{
		Level:    ERROR,
		Code:     "E1001",
		Message:  "test error",
		Position: Position{File: "test.c", Line: 1, Column: 1},
	}

	// Verify Error implements the error interface
	var _ error = err
}

// TestErrorCode tests ErrorCode type.
func TestErrorCode(t *testing.T) {
	// Test that ErrorCode is a string type
	var code ErrorCode = "E1001"
	if string(code) != "E1001" {
		t.Errorf("ErrorCode conversion failed, got %q", string(code))
	}

	// Test predefined error codes
	if ErrInvalidChar != "E0001" {
		t.Errorf("ErrInvalidChar = %q, want %q", ErrInvalidChar, "E0001")
	}
	if ErrSyntaxError != "E1001" {
		t.Errorf("ErrSyntaxError = %q, want %q", ErrSyntaxError, "E1001")
	}
	if ErrUndefinedSymbol != "E2001" {
		t.Errorf("ErrUndefinedSymbol = %q, want %q", ErrUndefinedSymbol, "E2001")
	}
}

// TestErrorLevel_Constants tests ErrorLevel constants.
func TestErrorLevel_Constants(t *testing.T) {
	if ERROR != 0 {
		t.Errorf("ERROR = %d, want 0", ERROR)
	}
	if WARNING != 1 {
		t.Errorf("WARNING = %d, want 1", WARNING)
	}
	if NOTE != 2 {
		t.Errorf("NOTE = %d, want 2", NOTE)
	}
}