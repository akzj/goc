// Package errhand provides error handling and diagnostic reporting for the GOC compiler.
// This file defines source position tracking.
package errhand

import (
	"fmt"
	"strings"
)

// Position tracks a location in source code for error reporting.
type Position struct {
	// File is the source file path (may be empty for stdin).
	File string
	// Line is the 1-based line number.
	Line int
	// Column is the 1-based column number (in runes).
	Column int
}

// String returns a formatted position string.
// Format: "file:line:column" or "line:column" if no file.
func (p Position) String() string {
	if p.File != "" {
		return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
	}
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// IsValid returns true if the position is valid (line > 0).
func (p Position) IsValid() bool {
	return p.Line > 0
}

// SourceContext provides source code context around a position.
type SourceContext struct {
	// Position is the error position.
	Position Position
	// LineContent is the full line content.
	LineContent string
	// StartCol is the start column for highlighting.
	StartCol int
	// EndCol is the end column for highlighting.
	EndCol int
}

// String returns a formatted context string with source snippet.
// Shows the source line with a caret (^) pointing to the error position.
func (c *SourceContext) String() string {
	if c == nil || c.LineContent == "" {
		return ""
	}

	var sb strings.Builder

	// Add the source line with line number
	sb.WriteString(fmt.Sprintf("%d | %s\n", c.Position.Line, c.LineContent))

	// Add the caret line
	// Calculate the offset for the caret (accounting for line number width and " | ")
	prefixWidth := len(fmt.Sprintf("%d | ", c.Position.Line))

	// Add spaces to position the caret
	startCol := c.StartCol
	if startCol <= 0 {
		startCol = c.Position.Column
	}

	// Add leading spaces (column is 1-based, so subtract 1)
	sb.WriteString(strings.Repeat(" ", prefixWidth+startCol-1))

	// Add caret(s)
	endCol := c.EndCol
	if endCol <= 0 {
		endCol = startCol
	}
	caretCount := endCol - startCol + 1
	if caretCount < 1 {
		caretCount = 1
	}
	sb.WriteString(strings.Repeat("^", caretCount))

	return sb.String()
}