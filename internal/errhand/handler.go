// Package errhand provides error handling and diagnostic reporting for the GOC compiler.
// This file implements the error handler that collects and reports errors.
package errhand

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// ErrorHandler collects and reports compilation errors.
// It supports error collection, source context display, and both
// human-readable and machine-readable output formats.
type ErrorHandler struct {
	// errors is the list of collected errors.
	errors []*Error
	// maxErrors is the maximum number of errors before stopping.
	maxErrors int
	// source is a cache of source file contents for context.
	source map[string]string
	// outputFormat controls the output format (human or machine).
	outputFormat OutputFormat
}

// OutputFormat controls how errors are formatted.
type OutputFormat int

const (
	// HumanReadable format includes source context and formatting.
	HumanReadable OutputFormat = iota
	// MachineReadable format is simple and parsable.
	MachineReadable
)

// ErrorHandlerConfig configures a new ErrorHandler.
type ErrorHandlerConfig struct {
	// MaxErrors is the maximum number of errors before stopping (0 = unlimited).
	MaxErrors int
	// OutputFormat controls the output format.
	OutputFormat OutputFormat
}

// NewErrorHandler creates a new error handler with default configuration.
// Default: unlimited errors, human-readable output.
func NewErrorHandler() *ErrorHandler {
	return NewErrorHandlerWithConfig(ErrorHandlerConfig{
		MaxErrors:    0,
		OutputFormat: HumanReadable,
	})
}

// NewErrorHandlerWithConfig creates a new error handler with custom configuration.
func NewErrorHandlerWithConfig(config ErrorHandlerConfig) *ErrorHandler {
	return &ErrorHandler{
		errors:       make([]*Error, 0),
		maxErrors:    config.MaxErrors,
		source:       make(map[string]string),
		outputFormat: config.OutputFormat,
	}
}

// Error adds an error to the handler.
// Returns true if the error was added, false if maxErrors limit reached.
func (h *ErrorHandler) Error(code ErrorCode, message string, pos Position) bool {
	return h.addError(&Error{
		Level:    ERROR,
		Code:     code,
		Message:  message,
		Position: pos,
		Hint:     "",
		Related:  nil,
	})
}

// Warning adds a warning to the handler.
// Warnings do not count toward maxErrors limit.
func (h *ErrorHandler) Warning(code ErrorCode, message string, pos Position) bool {
	return h.addError(&Error{
		Level:    WARNING,
		Code:     code,
		Message:  message,
		Position: pos,
		Hint:     "",
		Related:  nil,
	})
}

// Note adds a note to the handler.
// Notes are informational and don't count toward maxErrors limit.
func (h *ErrorHandler) Note(message string, pos Position) bool {
	return h.addError(&Error{
		Level:    NOTE,
		Code:     "",
		Message:  message,
		Position: pos,
		Hint:     "",
		Related:  nil,
	})
}

// ErrorWithHint adds an error with a hint to the handler.
// Returns true if the error was added, false if maxErrors limit reached.
func (h *ErrorHandler) ErrorWithHint(code ErrorCode, message, hint string, pos Position) bool {
	return h.addError(&Error{
		Level:    ERROR,
		Code:     code,
		Message:  message,
		Position: pos,
		Hint:     hint,
		Related:  nil,
	})
}

// ErrorWithRelated adds an error with related information.
// Useful for multi-point errors (e.g., duplicate declarations).
// Returns true if the error was added, false if maxErrors limit reached.
func (h *ErrorHandler) ErrorWithRelated(code ErrorCode, message string, pos Position, related []RelatedInfo) bool {
	return h.addError(&Error{
		Level:    ERROR,
		Code:     code,
		Message:  message,
		Position: pos,
		Hint:     "",
		Related:  related,
	})
}

// addError adds an error to the internal list.
// Returns true if the error was added, false if maxErrors limit reached.
func (h *ErrorHandler) addError(err *Error) bool {
	// Check if we've reached the max error limit (only for ERROR level)
	if err.Level == ERROR && h.maxErrors > 0 && h.ErrorCount() >= h.maxErrors {
		return false
	}

	h.errors = append(h.errors, err)
	return true
}

// HasErrors returns true if any errors (ERROR level) have been reported.
func (h *ErrorHandler) HasErrors() bool {
	for _, err := range h.errors {
		if err.Level == ERROR {
			return true
		}
	}
	return false
}

// ErrorCount returns the number of errors (not including warnings/notes).
func (h *ErrorHandler) ErrorCount() int {
	count := 0
	for _, err := range h.errors {
		if err.Level == ERROR {
			count++
		}
	}
	return count
}

// WarningCount returns the number of warnings.
func (h *ErrorHandler) WarningCount() int {
	count := 0
	for _, err := range h.errors {
		if err.Level == WARNING {
			count++
		}
	}
	return count
}

// NoteCount returns the number of notes.
func (h *ErrorHandler) NoteCount() int {
	count := 0
	for _, err := range h.errors {
		if err.Level == NOTE {
			count++
		}
	}
	return count
}

// TotalCount returns the total number of diagnostics (errors + warnings + notes).
func (h *ErrorHandler) TotalCount() int {
	return len(h.errors)
}

// Report prints all errors to stderr in the configured format.
func (h *ErrorHandler) Report() {
	h.ReportTo(os.Stderr)
}

// ReportTo prints all errors to the specified writer.
func (h *ErrorHandler) ReportTo(w io.Writer) {
	if h.outputFormat == MachineReadable {
		h.reportMachineReadable(w)
	} else {
		h.reportHumanReadable(w)
	}
}

// reportHumanReadable prints errors with source context and formatting.
func (h *ErrorHandler) reportHumanReadable(w io.Writer) {
	for _, err := range h.errors {
		fmt.Fprintln(w, err.String())

		// Add source context if available
		if err.Position.IsValid() && err.Position.File != "" {
			// Use enhanced context with surrounding lines
			if ctx := h.GetContextWithSurroundings(err.Position, 2, 2); ctx != nil && ctx.LineContent != "" {
				fmt.Fprintln(w, ctx.String())
			}
		}

		// Add blank line between errors for readability
		fmt.Fprintln(w)
	}

	// Print summary
	h.printSummary(w)
}

// reportMachineReadable prints errors in a simple, parsable format.
// Format: file:line:col:LEVEL:CODE:message
func (h *ErrorHandler) reportMachineReadable(w io.Writer) {
	for _, err := range h.errors {
		fmt.Fprintf(w, "%s:%s:%s: %s\n",
			err.Position.String(),
			err.Level.String(),
			err.Code,
			err.Message)
	}
}

// printSummary prints a summary of errors, warnings, and notes.
func (h *ErrorHandler) printSummary(w io.Writer) {
	errorCount := h.ErrorCount()
	warningCount := h.WarningCount()
	noteCount := h.NoteCount()

	if errorCount > 0 || warningCount > 0 || noteCount > 0 {
		fmt.Fprintln(w, "Summary:")
		if errorCount > 0 {
			fmt.Fprintf(w, "  %d error(s)\n", errorCount)
		}
		if warningCount > 0 {
			fmt.Fprintf(w, "  %d warning(s)\n", warningCount)
		}
		if noteCount > 0 {
			fmt.Fprintf(w, "  %d note(s)\n", noteCount)
		}
	}
}

// Reset clears all errors.
func (h *ErrorHandler) Reset() {
	h.errors = make([]*Error, 0)
}

// Errors returns a copy of the error list.
// The returned slice is a copy; modifications won't affect the handler.
func (h *ErrorHandler) Errors() []*Error {
	result := make([]*Error, len(h.errors))
	copy(result, h.errors)
	return result
}

// CacheSource caches source code for a file for error context.
// This enables source context display in error messages.
func (h *ErrorHandler) CacheSource(file, source string) {
	if h.source == nil {
		h.source = make(map[string]string)
	}
	h.source[file] = source
}

// GetContext returns source context for a position.
// Returns nil if source is not cached or position is invalid.
func (h *ErrorHandler) GetContext(pos Position) *SourceContext {
	if !pos.IsValid() || pos.File == "" {
		return nil
	}

	source, ok := h.source[pos.File]
	if !ok {
		return nil
	}

	lines := strings.Split(source, "\n")
	if pos.Line < 1 || pos.Line > len(lines) {
		return nil
	}

	lineContent := lines[pos.Line-1]

	return &SourceContext{
		Position:    pos,
		LineContent: lineContent,
		StartCol:    pos.Column,
		EndCol:      pos.Column,
	}
}

// GetContextWithRange returns source context with a custom highlight range.
// This is useful for highlighting specific tokens or expressions.
func (h *ErrorHandler) GetContextWithRange(pos Position, startCol, endCol int) *SourceContext {
	if !pos.IsValid() || pos.File == "" {
		return nil
	}

	source, ok := h.source[pos.File]
	if !ok {
		return nil
	}

	lines := strings.Split(source, "\n")
	if pos.Line < 1 || pos.Line > len(lines) {
		return nil
	}

	lineContent := lines[pos.Line-1]

	return &SourceContext{
		Position:    pos,
		LineContent: lineContent,
		StartCol:    startCol,
		EndCol:      endCol,
	}
}

// GetContextWithSurroundings returns source context with surrounding lines.
// Shows contextLinesBefore lines before and contextLinesAfter lines after the error.
// Default: 2 lines before and 2 lines after if contextLinesBefore/After <= 0.
func (h *ErrorHandler) GetContextWithSurroundings(pos Position, contextLinesBefore, contextLinesAfter int) *SourceContext {
	if !pos.IsValid() || pos.File == "" {
		return nil
	}

	source, ok := h.source[pos.File]
	if !ok {
		return nil
	}

	lines := strings.Split(source, "\n")
	// Remove trailing empty line if source ends with newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if pos.Line < 1 || pos.Line > len(lines) {
		return nil
	}

	// Set defaults if not specified
	if contextLinesBefore <= 0 {
		contextLinesBefore = 2
	}
	if contextLinesAfter <= 0 {
		contextLinesAfter = 2
	}

	lineContent := lines[pos.Line-1]

	// Calculate surrounding lines
	beforeStart := pos.Line - 1 - contextLinesBefore
	if beforeStart < 0 {
		beforeStart = 0
	}
	beforeLines := lines[beforeStart : pos.Line-1]

	afterEnd := pos.Line + contextLinesAfter
	if afterEnd > len(lines) {
		afterEnd = len(lines)
	}
	afterLines := lines[pos.Line:afterEnd]

	return &SourceContext{
		Position:    pos,
		LineContent: lineContent,
		StartCol:    pos.Column,
		EndCol:      pos.Column,
		BeforeLines: beforeLines,
		AfterLines:  afterLines,
		UseColors:   true,
	}
}

// SetOutputFormat sets the output format for error reporting.
func (h *ErrorHandler) SetOutputFormat(format OutputFormat) {
	h.outputFormat = format
}

// SetMaxErrors sets the maximum number of errors before stopping.
// Set to 0 for unlimited errors.
func (h *ErrorHandler) SetMaxErrors(max int) {
	h.maxErrors = max
}

// ShouldStop returns true if compilation should stop due to errors.
func (h *ErrorHandler) ShouldStop() bool {
	return h.maxErrors > 0 && h.ErrorCount() >= h.maxErrors
}

// String returns a string representation of all errors.
func (h *ErrorHandler) String() string {
	var sb strings.Builder
	for _, err := range h.errors {
		sb.WriteString(err.String())
		sb.WriteString("\n")
	}
	return sb.String()
}