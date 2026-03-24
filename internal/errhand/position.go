// Package errhand provides error handling and diagnostic reporting for the GOC compiler.
// This file defines source position tracking and enhanced source context display.
package errhand

import (
	"fmt"
	"strings"
)

// ANSI color codes for terminal output.
const (
	// ColorReset resets all attributes.
	ColorReset = "\033[0m"
	// ColorBold makes text bold.
	ColorBold = "\033[1m"
	// ColorRed makes text red (for errors).
	ColorRed = "\033[31m"
	// ColorYellow makes text yellow (for warnings).
	ColorYellow = "\033[33m"
	// ColorBlue makes text blue (for notes/info).
	ColorBlue = "\033[34m"
	// ColorGreen makes text green (for success).
	ColorGreen = "\033[32m"
	// ColorCyan makes text cyan (for line numbers).
	ColorCyan = "\033[36m"
	// ColorGray makes text gray (for context lines).
	ColorGray = "\033[90m"
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
// Enhanced to support multi-line context with surrounding lines.
type SourceContext struct {
	// Position is the error position.
	Position Position
	// LineContent is the full line content for the error line.
	LineContent string
	// StartCol is the start column for highlighting.
	StartCol int
	// EndCol is the end column for highlighting.
	EndCol int
	// BeforeLines contains lines before the error line (optional).
	BeforeLines []string
	// AfterLines contains lines after the error line (optional).
	AfterLines []string
	// UseColors enables ANSI color codes in output.
	UseColors bool
}

// String returns a formatted context string with source snippet.
// Shows the source line with a caret (^) pointing to the error position.
// Enhanced version includes surrounding context lines with ANSI colors.
func (c *SourceContext) String() string {
	if c == nil || c.LineContent == "" {
		return ""
	}

	var sb strings.Builder

	// Calculate the maximum line number width for alignment
	maxLineNum := c.Position.Line
	if len(c.AfterLines) > 0 {
		maxLineNum = c.Position.Line + len(c.AfterLines)
	}
	lineNumWidth := len(fmt.Sprintf("%d", maxLineNum))

	// Display lines before the error (context)
	for i, line := range c.BeforeLines {
		lineNum := c.Position.Line - len(c.BeforeLines) + i
		if c.UseColors {
			sb.WriteString(fmt.Sprintf("%s%*d%s | %s%s\n", ColorGray, lineNumWidth, lineNum, ColorReset, line, ColorReset))
		} else {
			sb.WriteString(fmt.Sprintf("%*d | %s\n", lineNumWidth, lineNum, line))
		}
	}

	// Display the error line
	if c.UseColors {
		sb.WriteString(fmt.Sprintf("%s%*d%s | %s%s\n", ColorBold, lineNumWidth, c.Position.Line, ColorReset, c.LineContent, ColorReset))
	} else {
		sb.WriteString(fmt.Sprintf("%*d | %s\n", lineNumWidth, c.Position.Line, c.LineContent))
	}

	// Display the caret line
	prefixWidth := lineNumWidth + len(" | ")
	startCol := c.StartCol
	if startCol <= 0 {
		startCol = c.Position.Column
	}

	// Add leading spaces
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

	if c.UseColors {
		sb.WriteString(fmt.Sprintf("%s%s%s", ColorRed, strings.Repeat("^", caretCount), ColorReset))
	} else {
		sb.WriteString(strings.Repeat("^", caretCount))
	}

	// Display lines after the error (context)
	for i, line := range c.AfterLines {
		lineNum := c.Position.Line + 1 + i
		if c.UseColors {
			sb.WriteString(fmt.Sprintf("\n%s%*d%s | %s%s", ColorGray, lineNumWidth, lineNum, ColorReset, line, ColorReset))
		} else {
			sb.WriteString(fmt.Sprintf("\n%*d | %s", lineNumWidth, lineNum, line))
		}
	}

	return sb.String()
}