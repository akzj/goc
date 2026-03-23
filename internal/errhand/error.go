// Package errhand provides error handling and diagnostic reporting for the GOC compiler.
// This package defines error types, position tracking, and error collection.
package errhand

import (
	"strings"
)

// ErrorLevel indicates the severity of a diagnostic.
type ErrorLevel int

const (
	// ERROR indicates a compilation error that prevents further compilation.
	ERROR ErrorLevel = iota
	// WARNING indicates a potential issue that doesn't prevent compilation.
	WARNING
	// NOTE indicates informational messages.
	NOTE
)

// String returns the string representation of the error level.
func (l ErrorLevel) String() string {
	switch l {
	case ERROR:
		return "ERROR"
	case WARNING:
		return "WARNING"
	case NOTE:
		return "NOTE"
	default:
		return "UNKNOWN"
	}
}

// ErrorCode uniquely identifies error types.
// Error codes are in ranges:
//   - E0001-E0999: Lexer errors
//   - E1001-E1999: Parser errors
//   - E2001-E2999: Semantic errors
//   - E3001-E3999: IR errors
//   - E4001-E4999: CodeGen errors
//   - E5001-E5999: Linker errors
type ErrorCode string

// Error represents a compilation error or warning.
type Error struct {
	// Level is the severity of the error.
	Level ErrorLevel
	// Code is the unique error code.
	Code ErrorCode
	// Message is the human-readable error message.
	Message string
	// Position is the source location of the error.
	Position Position
	// Hint is an optional hint for fixing the error.
	Hint string
	// Related contains related locations for multi-point errors.
	Related []RelatedInfo
}

// RelatedInfo provides additional context for an error.
type RelatedInfo struct {
	// Position is the related source location.
	Position Position
	// Message describes the relationship.
	Message string
}

// String returns a formatted error string.
// Format: "position: LEVEL: CODE: message [hint]"
// Example: "main.c:10:5: ERROR: E1001: syntax error"
func (e *Error) String() string {
	var sb strings.Builder
	sb.WriteString(e.Position.String())
	sb.WriteString(": ")
	sb.WriteString(e.Level.String())
	sb.WriteString(": ")
	sb.WriteString(string(e.Code))
	sb.WriteString(": ")
	sb.WriteString(e.Message)
	if e.Hint != "" {
		sb.WriteString("\n  hint: ")
		sb.WriteString(e.Hint)
	}
	// Add related info if present
	for _, rel := range e.Related {
		sb.WriteString("\n  note: ")
		sb.WriteString(rel.Position.String())
		sb.WriteString(": ")
		sb.WriteString(rel.Message)
	}
	return sb.String()
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.String()
}