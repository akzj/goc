// Package errhand provides error handling and diagnostic reporting for the GOC compiler.
// This file contains unit tests for the error handler.
package errhand

import (
	"bytes"
	"strings"
	"testing"
)

// TestNewErrorHandler tests the default constructor.
func TestNewErrorHandler(t *testing.T) {
	h := NewErrorHandler()

	if h == nil {
		t.Fatal("NewErrorHandler() returned nil")
	}

	if h.errors == nil {
		t.Error("NewErrorHandler() did not initialize errors slice")
	}

	if h.source == nil {
		t.Error("NewErrorHandler() did not initialize source map")
	}

	if h.maxErrors != 0 {
		t.Errorf("NewErrorHandler() maxErrors = %d, want 0 (unlimited)", h.maxErrors)
	}

	if h.outputFormat != HumanReadable {
		t.Errorf("NewErrorHandler() outputFormat = %v, want HumanReadable", h.outputFormat)
	}
}

// TestNewErrorHandlerWithConfig tests the configurable constructor.
func TestNewErrorHandlerWithConfig(t *testing.T) {
	config := ErrorHandlerConfig{
		MaxErrors:    10,
		OutputFormat: MachineReadable,
	}
	h := NewErrorHandlerWithConfig(config)

	if h == nil {
		t.Fatal("NewErrorHandlerWithConfig() returned nil")
	}

	if h.maxErrors != 10 {
		t.Errorf("NewErrorHandlerWithConfig() maxErrors = %d, want 10", h.maxErrors)
	}

	if h.outputFormat != MachineReadable {
		t.Errorf("NewErrorHandlerWithConfig() outputFormat = %v, want MachineReadable", h.outputFormat)
	}
}

// TestErrorHandler_Error tests adding errors.
func TestErrorHandler_Error(t *testing.T) {
	h := NewErrorHandler()

	added := h.Error(ErrSyntaxError, "syntax error", Position{File: "test.c", Line: 10, Column: 5})

	if !added {
		t.Error("Error() returned false, expected true")
	}

	if h.ErrorCount() != 1 {
		t.Errorf("ErrorCount() = %d, want 1", h.ErrorCount())
	}

	if len(h.errors) != 1 {
		t.Errorf("len(errors) = %d, want 1", len(h.errors))
	}

	err := h.errors[0]
	if err.Level != ERROR {
		t.Errorf("Error level = %v, want ERROR", err.Level)
	}
	if err.Code != ErrSyntaxError {
		t.Errorf("Error code = %q, want %q", err.Code, ErrSyntaxError)
	}
	if err.Message != "syntax error" {
		t.Errorf("Error message = %q, want %q", err.Message, "syntax error")
	}
	if err.Position.Line != 10 {
		t.Errorf("Error line = %d, want 10", err.Position.Line)
	}
}

// TestErrorHandler_Warning tests adding warnings.
func TestErrorHandler_Warning(t *testing.T) {
	h := NewErrorHandler()

	added := h.Warning(ErrorCode("W0001"), "unused variable", Position{File: "test.c", Line: 5, Column: 1})

	if !added {
		t.Error("Warning() returned false, expected true")
	}

	if h.WarningCount() != 1 {
		t.Errorf("WarningCount() = %d, want 1", h.WarningCount())
	}

	if h.ErrorCount() != 0 {
		t.Errorf("ErrorCount() = %d, want 0 (warnings shouldn't count as errors)", h.ErrorCount())
	}

	err := h.errors[0]
	if err.Level != WARNING {
		t.Errorf("Warning level = %v, want WARNING", err.Level)
	}
}

// TestErrorHandler_Note tests adding notes.
func TestErrorHandler_Note(t *testing.T) {
	h := NewErrorHandler()

	added := h.Note("previous declaration here", Position{File: "test.c", Line: 3, Column: 1})

	if !added {
		t.Error("Note() returned false, expected true")
	}

	if h.NoteCount() != 1 {
		t.Errorf("NoteCount() = %d, want 1", h.NoteCount())
	}

	if h.ErrorCount() != 0 {
		t.Errorf("ErrorCount() = %d, want 0 (notes shouldn't count as errors)", h.ErrorCount())
	}

	err := h.errors[0]
	if err.Level != NOTE {
		t.Errorf("Note level = %v, want NOTE", err.Level)
	}
	if err.Code != "" {
		t.Errorf("Note code = %q, want empty string", err.Code)
	}
}

// TestErrorHandler_ErrorWithHint tests adding errors with hints.
func TestErrorHandler_ErrorWithHint(t *testing.T) {
	h := NewErrorHandler()

	added := h.ErrorWithHint(ErrUndefinedSymbol, "undefined symbol", "did you mean to declare the variable?", Position{File: "test.c", Line: 15, Column: 10})

	if !added {
		t.Error("ErrorWithHint() returned false, expected true")
	}

	err := h.errors[0]
	if err.Hint != "did you mean to declare the variable?" {
		t.Errorf("Error hint = %q, want %q", err.Hint, "did you mean to declare the variable?")
	}
}

// TestErrorHandler_ErrorWithRelated tests adding errors with related info.
func TestErrorHandler_ErrorWithRelated(t *testing.T) {
	h := NewErrorHandler()

	related := []RelatedInfo{
		{Position: Position{File: "test.c", Line: 5, Column: 5}, Message: "first declaration here"},
	}

	added := h.ErrorWithRelated(ErrDuplicateSymbol, "duplicate symbol", Position{File: "test.c", Line: 10, Column: 5}, related)

	if !added {
		t.Error("ErrorWithRelated() returned false, expected true")
	}

	err := h.errors[0]
	if len(err.Related) != 1 {
		t.Errorf("len(Related) = %d, want 1", len(err.Related))
	}
	if err.Related[0].Message != "first declaration here" {
		t.Errorf("Related message = %q, want %q", err.Related[0].Message, "first declaration here")
	}
}

// TestErrorHandler_MaxErrors tests the maxErrors limit.
func TestErrorHandler_MaxErrors(t *testing.T) {
	h := NewErrorHandlerWithConfig(ErrorHandlerConfig{
		MaxErrors: 2,
	})

	// Add first error - should succeed
	if !h.Error(ErrSyntaxError, "error 1", Position{File: "test.c", Line: 1, Column: 1}) {
		t.Error("First error should be added")
	}

	// Add second error - should succeed
	if !h.Error(ErrSyntaxError, "error 2", Position{File: "test.c", Line: 2, Column: 1}) {
		t.Error("Second error should be added")
	}

	// Add third error - should fail (max reached)
	if h.Error(ErrSyntaxError, "error 3", Position{File: "test.c", Line: 3, Column: 1}) {
		t.Error("Third error should not be added (max reached)")
	}

	// Warnings should still be added even after max errors
	if !h.Warning(ErrorCode("W0001"), "warning", Position{File: "test.c", Line: 4, Column: 1}) {
		t.Error("Warning should be added even after max errors")
	}

	if h.ErrorCount() != 2 {
		t.Errorf("ErrorCount() = %d, want 2", h.ErrorCount())
	}

	if h.WarningCount() != 1 {
		t.Errorf("WarningCount() = %d, want 1", h.WarningCount())
	}
}

// TestErrorHandler_HasErrors tests the HasErrors method.
func TestErrorHandler_HasErrors(t *testing.T) {
	h := NewErrorHandler()

	if h.HasErrors() {
		t.Error("HasErrors() = true, want false (no errors yet)")
	}

	h.Error(ErrSyntaxError, "error", Position{File: "test.c", Line: 1, Column: 1})

	if !h.HasErrors() {
		t.Error("HasErrors() = false, want true (error added)")
	}

	// Reset and test with only warnings
	h.Reset()
	h.Warning(ErrorCode("W0001"), "warning", Position{File: "test.c", Line: 1, Column: 1})

	if h.HasErrors() {
		t.Error("HasErrors() = true, want false (only warnings)")
	}
}

// TestErrorHandler_Counts tests all counting methods.
func TestErrorHandler_Counts(t *testing.T) {
	h := NewErrorHandler()

	h.Error(ErrSyntaxError, "error 1", Position{File: "test.c", Line: 1, Column: 1})
	h.Error(ErrSyntaxError, "error 2", Position{File: "test.c", Line: 2, Column: 1})
	h.Warning(ErrorCode("W0001"), "warning 1", Position{File: "test.c", Line: 3, Column: 1})
	h.Warning(ErrorCode("W0001"), "warning 2", Position{File: "test.c", Line: 4, Column: 1})
	h.Warning(ErrorCode("W0001"), "warning 3", Position{File: "test.c", Line: 5, Column: 1})
	h.Note("note 1", Position{File: "test.c", Line: 6, Column: 1})

	if h.ErrorCount() != 2 {
		t.Errorf("ErrorCount() = %d, want 2", h.ErrorCount())
	}
	if h.WarningCount() != 3 {
		t.Errorf("WarningCount() = %d, want 3", h.WarningCount())
	}
	if h.NoteCount() != 1 {
		t.Errorf("NoteCount() = %d, want 1", h.NoteCount())
	}
	if h.TotalCount() != 6 {
		t.Errorf("TotalCount() = %d, want 6", h.TotalCount())
	}
}

// TestErrorHandler_Reset tests the Reset method.
func TestErrorHandler_Reset(t *testing.T) {
	h := NewErrorHandler()

	h.Error(ErrSyntaxError, "error", Position{File: "test.c", Line: 1, Column: 1})
	h.Warning(ErrorCode("W0001"), "warning", Position{File: "test.c", Line: 2, Column: 1})

	if h.ErrorCount() != 1 {
		t.Errorf("Before reset: ErrorCount() = %d, want 1", h.ErrorCount())
	}

	h.Reset()

	if h.ErrorCount() != 0 {
		t.Errorf("After reset: ErrorCount() = %d, want 0", h.ErrorCount())
	}
	if h.WarningCount() != 0 {
		t.Errorf("After reset: WarningCount() = %d, want 0", h.WarningCount())
	}
	if h.TotalCount() != 0 {
		t.Errorf("After reset: TotalCount() = %d, want 0", h.TotalCount())
	}
}

// TestErrorHandler_Errors tests the Errors method returns a copy.
func TestErrorHandler_Errors(t *testing.T) {
	h := NewErrorHandler()

	h.Error(ErrSyntaxError, "error 1", Position{File: "test.c", Line: 1, Column: 1})
	h.Error(ErrSyntaxError, "error 2", Position{File: "test.c", Line: 2, Column: 1})

	errors := h.Errors()

	if len(errors) != 2 {
		t.Errorf("len(Errors()) = %d, want 2", len(errors))
	}

	// Test that modifying the slice length doesn't affect the handler
	originalLen := len(errors)
	errors = append(errors, &Error{Message: "added"})

	if len(h.errors) != originalLen {
		t.Error("Appending to returned slice affected handler (not a copy)")
	}

	// Test that we can modify the slice without affecting handler
	errors[0] = nil
	if h.errors[0] == nil {
		t.Error("Setting slice element to nil affected handler (not a copy)")
	}
}

// TestErrorHandler_CacheSource tests source caching.
func TestErrorHandler_CacheSource(t *testing.T) {
	h := NewErrorHandler()

	source := "int main() {\n    return 0;\n}\n"
	h.CacheSource("test.c", source)

	cached, ok := h.source["test.c"]
	if !ok {
		t.Fatal("Source not cached")
	}
	if cached != source {
		t.Errorf("Cached source = %q, want %q", cached, source)
	}
}

// TestErrorHandler_GetContext tests source context retrieval.
func TestErrorHandler_GetContext(t *testing.T) {
	h := NewErrorHandler()

	source := "int main() {\n    return 0;\n}\n"
	h.CacheSource("test.c", source)

	ctx := h.GetContext(Position{File: "test.c", Line: 1, Column: 5})

	if ctx == nil {
		t.Fatal("GetContext() returned nil")
	}
	if ctx.LineContent != "int main() {" {
		t.Errorf("LineContent = %q, want %q", ctx.LineContent, "int main() {")
	}
	if ctx.Position.Line != 1 {
		t.Errorf("Position.Line = %d, want 1", ctx.Position.Line)
	}
	if ctx.StartCol != 5 {
		t.Errorf("StartCol = %d, want 5", ctx.StartCol)
	}
}

// TestErrorHandler_GetContext_InvalidPosition tests context retrieval with invalid positions.
func TestErrorHandler_GetContext_InvalidPosition(t *testing.T) {
	h := NewErrorHandler()

	source := "int main() {\n    return 0;\n}\n"
	h.CacheSource("test.c", source)

	// Test with invalid line
	ctx := h.GetContext(Position{File: "test.c", Line: 0, Column: 1})
	if ctx != nil {
		t.Error("GetContext() should return nil for line 0")
	}

	// Test with line beyond file
	ctx = h.GetContext(Position{File: "test.c", Line: 100, Column: 1})
	if ctx != nil {
		t.Error("GetContext() should return nil for line beyond file")
	}

	// Test with no file
	ctx = h.GetContext(Position{File: "", Line: 1, Column: 1})
	if ctx != nil {
		t.Error("GetContext() should return nil for empty file")
	}

	// Test with uncached file
	ctx = h.GetContext(Position{File: "unknown.c", Line: 1, Column: 1})
	if ctx != nil {
		t.Error("GetContext() should return nil for uncached file")
	}
}

// TestErrorHandler_GetContextWithRange tests context with custom range.
func TestErrorHandler_GetContextWithRange(t *testing.T) {
	h := NewErrorHandler()

	source := "int result = calculate(a, b);"
	h.CacheSource("test.c", source)

	ctx := h.GetContextWithRange(Position{File: "test.c", Line: 1, Column: 16}, 16, 25)

	if ctx == nil {
		t.Fatal("GetContextWithRange() returned nil")
	}
	if ctx.StartCol != 16 {
		t.Errorf("StartCol = %d, want 16", ctx.StartCol)
	}
	if ctx.EndCol != 25 {
		t.Errorf("EndCol = %d, want 25", ctx.EndCol)
	}
}

// TestErrorHandler_SetOutputFormat tests output format configuration.
func TestErrorHandler_SetOutputFormat(t *testing.T) {
	h := NewErrorHandler()

	if h.outputFormat != HumanReadable {
		t.Errorf("Default outputFormat = %v, want HumanReadable", h.outputFormat)
	}

	h.SetOutputFormat(MachineReadable)

	if h.outputFormat != MachineReadable {
		t.Errorf("After SetOutputFormat: outputFormat = %v, want MachineReadable", h.outputFormat)
	}
}

// TestErrorHandler_SetMaxErrors tests maxErrors configuration.
func TestErrorHandler_SetMaxErrors(t *testing.T) {
	h := NewErrorHandler()

	if h.maxErrors != 0 {
		t.Errorf("Default maxErrors = %d, want 0", h.maxErrors)
	}

	h.SetMaxErrors(5)

	if h.maxErrors != 5 {
		t.Errorf("After SetMaxErrors: maxErrors = %d, want 5", h.maxErrors)
	}
}

// TestErrorHandler_ShouldStop tests the ShouldStop method.
func TestErrorHandler_ShouldStop(t *testing.T) {
	h := NewErrorHandler()

	// Unlimited errors - should never stop
	h.Error(ErrSyntaxError, "error 1", Position{File: "test.c", Line: 1, Column: 1})
	h.Error(ErrSyntaxError, "error 2", Position{File: "test.c", Line: 2, Column: 1})

	if h.ShouldStop() {
		t.Error("ShouldStop() = true with unlimited errors, want false")
	}

	// Limited errors - should stop after max
	h = NewErrorHandlerWithConfig(ErrorHandlerConfig{MaxErrors: 2})
	h.Error(ErrSyntaxError, "error 1", Position{File: "test.c", Line: 1, Column: 1})

	if h.ShouldStop() {
		t.Error("ShouldStop() = true before reaching max, want false")
	}

	h.Error(ErrSyntaxError, "error 2", Position{File: "test.c", Line: 2, Column: 1})

	if !h.ShouldStop() {
		t.Error("ShouldStop() = false after reaching max, want true")
	}
}

// TestErrorHandler_String tests the String method.
func TestErrorHandler_String(t *testing.T) {
	h := NewErrorHandler()

	h.Error(ErrSyntaxError, "syntax error", Position{File: "test.c", Line: 10, Column: 5})

	str := h.String()

	if !strings.Contains(str, "test.c:10:5") {
		t.Errorf("String() = %q, should contain position", str)
	}
	if !strings.Contains(str, "ERROR") {
		t.Errorf("String() = %q, should contain level", str)
	}
	if !strings.Contains(str, "E1001") {
		t.Errorf("String() = %q, should contain error code", str)
	}
	if !strings.Contains(str, "syntax error") {
		t.Errorf("String() = %q, should contain message", str)
	}
}

// TestErrorHandler_Report_HumanReadable tests human-readable report format.
func TestErrorHandler_Report_HumanReadable(t *testing.T) {
	h := NewErrorHandler()
	h.CacheSource("test.c", "int main() {\n    return 0;\n}\n")

	h.Error(ErrSyntaxError, "syntax error", Position{File: "test.c", Line: 1, Column: 5})

	var buf bytes.Buffer
	h.reportHumanReadable(&buf)

	output := buf.String()

	if !strings.Contains(output, "test.c:1:5") {
		t.Errorf("Report output = %q, should contain position", output)
	}
	if !strings.Contains(output, "syntax error") {
		t.Errorf("Report output = %q, should contain message", output)
	}
	if !strings.Contains(output, "Summary:") {
		t.Errorf("Report output = %q, should contain summary", output)
	}
}

// TestErrorHandler_Report_MachineReadable tests machine-readable report format.
func TestErrorHandler_Report_MachineReadable(t *testing.T) {
	h := NewErrorHandlerWithConfig(ErrorHandlerConfig{
		OutputFormat: MachineReadable,
	})

	h.Error(ErrSyntaxError, "syntax error", Position{File: "test.c", Line: 10, Column: 5})

	var buf bytes.Buffer
	h.reportMachineReadable(&buf)

	output := buf.String()

	// Machine format: file:line:col:LEVEL:CODE: message
	expected := "test.c:10:5:ERROR:E1001: syntax error\n"
	if output != expected {
		t.Errorf("MachineReadable output = %q, want %q", output, expected)
	}
}

// TestErrorHandler_ReportWithSourceContext tests that source context is included.
func TestErrorHandler_ReportWithSourceContext(t *testing.T) {
	h := NewErrorHandler()
	source := "int x = 5;\nint y = 10;\n"
	h.CacheSource("test.c", source)

	h.Error(ErrSyntaxError, "syntax error", Position{File: "test.c", Line: 1, Column: 5})

	var buf bytes.Buffer
	h.reportHumanReadable(&buf)

	output := buf.String()

	if !strings.Contains(output, "1 | int x = 5;") {
		t.Errorf("Report output = %q, should contain source line", output)
	}
	if !strings.Contains(output, "^") {
		t.Errorf("Report output = %q, should contain caret", output)
	}
}

// TestErrorHandler_MultipleErrors tests handling multiple errors.
func TestErrorHandler_MultipleErrors(t *testing.T) {
	h := NewErrorHandler()

	h.Error(ErrSyntaxError, "error 1", Position{File: "test.c", Line: 1, Column: 1})
	h.Warning(ErrorCode("W0001"), "warning 1", Position{File: "test.c", Line: 2, Column: 1})
	h.Error(ErrUndefinedSymbol, "error 2", Position{File: "test.c", Line: 3, Column: 1})
	h.Note("note 1", Position{File: "test.c", Line: 4, Column: 1})

	if h.TotalCount() != 4 {
		t.Errorf("TotalCount() = %d, want 4", h.TotalCount())
	}

	if h.ErrorCount() != 2 {
		t.Errorf("ErrorCount() = %d, want 2", h.ErrorCount())
	}

	if !h.HasErrors() {
		t.Error("HasErrors() = false, want true")
	}
}

// TestErrorHandler_EmptyReport tests reporting with no errors.
func TestErrorHandler_EmptyReport(t *testing.T) {
	h := NewErrorHandler()

	var buf bytes.Buffer
	h.reportHumanReadable(&buf)

	output := buf.String()

	// Should be empty or just summary with zeros
	if strings.Contains(output, "error(s)") || strings.Contains(output, "warning(s)") {
		// Check if counts are 0
		if !strings.Contains(output, "0 error(s)") && !strings.Contains(output, "Summary:") {
			t.Errorf("Empty report should have empty summary, got: %q", output)
		}
	}
}

// TestErrorHandler_AddError_ReturnValue tests the return value of addError.
func TestErrorHandler_AddError_ReturnValue(t *testing.T) {
	h := NewErrorHandlerWithConfig(ErrorHandlerConfig{MaxErrors: 1})

	// First error should return true
	if !h.Error(ErrSyntaxError, "error 1", Position{File: "test.c", Line: 1, Column: 1}) {
		t.Error("First error should return true")
	}

	// Second error should return false (max reached)
	if h.Error(ErrSyntaxError, "error 2", Position{File: "test.c", Line: 2, Column: 1}) {
		t.Error("Second error should return false (max reached)")
	}

	// Warning should still return true (doesn't count toward max)
	if !h.Warning(ErrorCode("W0001"), "warning", Position{File: "test.c", Line: 3, Column: 1}) {
		t.Error("Warning should return true even after max errors")
	}
}

// TestErrorHandler_PositionValidation tests position validation in context retrieval.
func TestErrorHandler_PositionValidation(t *testing.T) {
	h := NewErrorHandler()
	h.CacheSource("test.c", "line1\nline2\nline3\n")

	tests := []struct {
		name     string
		pos      Position
		wantNil  bool
		wantLine string
	}{
		{"valid position", Position{File: "test.c", Line: 2, Column: 1}, false, "line2"},
		{"line 0", Position{File: "test.c", Line: 0, Column: 1}, true, ""},
		{"negative line", Position{File: "test.c", Line: -1, Column: 1}, true, ""},
		{"line beyond file", Position{File: "test.c", Line: 100, Column: 1}, true, ""},
		{"empty file", Position{File: "", Line: 1, Column: 1}, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := h.GetContext(tt.pos)
			if tt.wantNil {
				if ctx != nil {
					t.Errorf("GetContext(%v) = %v, want nil", tt.pos, ctx)
				}
			} else {
				if ctx == nil {
					t.Fatalf("GetContext(%v) = nil, want non-nil", tt.pos)
				}
				if ctx.LineContent != tt.wantLine {
					t.Errorf("LineContent = %q, want %q", ctx.LineContent, tt.wantLine)
				}
			}
		})
	}
}

// TestErrorHandler_OutputFormat_Constants tests OutputFormat constants.
func TestErrorHandler_OutputFormat_Constants(t *testing.T) {
	if HumanReadable != 0 {
		t.Errorf("HumanReadable = %d, want 0", HumanReadable)
	}
	if MachineReadable != 1 {
		t.Errorf("MachineReadable = %d, want 1", MachineReadable)
	}
}

// TestErrorHandler_ConfigStruct tests ErrorHandlerConfig struct.
func TestErrorHandler_ConfigStruct(t *testing.T) {
	config := ErrorHandlerConfig{
		MaxErrors:    5,
		OutputFormat: MachineReadable,
	}

	if config.MaxErrors != 5 {
		t.Errorf("Config.MaxErrors = %d, want 5", config.MaxErrors)
	}
	if config.OutputFormat != MachineReadable {
		t.Errorf("Config.OutputFormat = %v, want MachineReadable", config.OutputFormat)
	}
}