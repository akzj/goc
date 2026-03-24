package errhand

import (
	"strings"
	"testing"
)

// TestSourceContext_String tests the String() method for SourceContext.
func TestSourceContext_String(t *testing.T) {
	tests := []struct {
		name        string
		context     *SourceContext
		expected    string
		expectEmpty bool
	}{
		{
			name: "simple error with single caret",
			context: &SourceContext{
				Position:    Position{File: "test.c", Line: 10, Column: 5},
				LineContent: "int x = 5;",
				StartCol:    5,
				EndCol:      5,
			},
			expected: "10 | int x = 5;\n         ^",
		},
		{
			name: "error at beginning of line",
			context: &SourceContext{
				Position:    Position{File: "test.c", Line: 1, Column: 1},
				LineContent: "return 0;",
				StartCol:    1,
				EndCol:      1,
			},
			expected: "1 | return 0;\n    ^",
		},
		{
			name: "nil context returns empty",
			context:     nil,
			expected:    "",
			expectEmpty: true,
		},
		{
			name: "empty line content returns empty",
			context: &SourceContext{
				Position:    Position{File: "test.c", Line: 10, Column: 5},
				LineContent: "",
				StartCol:    5,
				EndCol:      5,
			},
			expected:    "",
			expectEmpty: true,
		},
		{
			name: "uses Position.Column when StartCol is 0",
			context: &SourceContext{
				Position:    Position{File: "test.c", Line: 3, Column: 7},
				LineContent: "x = y + z;",
				StartCol:    0,
				EndCol:      0,
			},
			expected: "3 | x = y + z;\n          ^",
		},
		{
			name: "large line number",
			context: &SourceContext{
				Position:    Position{File: "test.c", Line: 100, Column: 10},
				LineContent: "printf(\"hello\");",
				StartCol:    1,
				EndCol:      1,
			},
			expected: "100 | printf(\"hello\");\n      ^",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.context.String()
			if tt.expectEmpty {
				if result != "" {
					t.Errorf("SourceContext.String() = %q, want empty string", result)
				}
			} else {
				if result != tt.expected {
					t.Errorf("SourceContext.String() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

// TestSourceContext_String_MultiCharHighlight tests highlighting multiple characters.
func TestSourceContext_String_MultiCharHighlight(t *testing.T) {
	ctx := &SourceContext{
		Position:    Position{File: "test.c", Line: 15, Column: 10},
		LineContent: "int result = calculate(a, b);",
		StartCol:    16,
		EndCol:      25,
	}

	result := ctx.String()
	expected := "15 | int result = calculate(a, b);\n                    ^^^^^^^^^^"

	if result != expected {
		t.Errorf("SourceContext.String() = %q, want %q", result, expected)
	}
}

// TestSourceContext_String_EdgeCases tests edge cases.
func TestSourceContext_String_EdgeCases(t *testing.T) {
	// Test when EndCol < StartCol (should default to single caret)
	ctx := &SourceContext{
		Position:    Position{File: "test.c", Line: 1, Column: 5},
		LineContent: "test",
		StartCol:    3,
		EndCol:      1,
	}

	result := ctx.String()
	// Should still show at least one caret
	if len(result) == 0 {
		t.Error("SourceContext.String() returned empty string for edge case")
	}
}

// TestSourceContext_WithSurroundings tests multi-line context display.
func TestSourceContext_WithSurroundings(t *testing.T) {
	ctx := &SourceContext{
		Position:    Position{File: "test.c", Line: 3, Column: 5},
		LineContent: "    return 0;",
		StartCol:    5,
		EndCol:      5,
		BeforeLines: []string{"#include <stdio.h>", "int main() {"},
		AfterLines:  []string{"}", ""},
		UseColors:   false,
	}

	result := ctx.String()

	// Check that all lines are present
	if !strings.Contains(result, "#include <stdio.h>") {
		t.Errorf("Result missing before line 1: %q", result)
	}
	if !strings.Contains(result, "int main() {") {
		t.Errorf("Result missing before line 2: %q", result)
	}
	if !strings.Contains(result, "    return 0;") {
		t.Errorf("Result missing error line: %q", result)
	}
	if !strings.Contains(result, "}") {
		t.Errorf("Result missing after line 1: %q", result)
	}

	// Check line numbers
	if !strings.Contains(result, "1 | ") {
		t.Errorf("Result missing line number 1: %q", result)
	}
	if !strings.Contains(result, "2 | ") {
		t.Errorf("Result missing line number 2: %q", result)
	}
	if !strings.Contains(result, "3 | ") {
		t.Errorf("Result missing line number 3: %q", result)
	}
	if !strings.Contains(result, "4 | ") {
		t.Errorf("Result missing line number 4: %q", result)
	}

	// Check caret
	if !strings.Contains(result, "^") {
		t.Errorf("Result missing caret: %q", result)
	}
}

// TestSourceContext_WithColors tests ANSI color codes.
func TestSourceContext_WithColors(t *testing.T) {
	ctx := &SourceContext{
		Position:    Position{File: "test.c", Line: 1, Column: 5},
		LineContent: "int x = 5;",
		StartCol:    5,
		EndCol:      5,
		UseColors:   true,
	}

	result := ctx.String()

	// Check for ANSI color codes
	if !strings.Contains(result, "\033[") {
		t.Errorf("Result missing ANSI color codes: %q", result)
	}
	if !strings.Contains(result, ColorReset) {
		t.Errorf("Result missing ColorReset: %q", result)
	}
	if !strings.Contains(result, ColorRed) {
		t.Errorf("Result missing ColorRed for caret: %q", result)
	}
}

// TestSourceContext_WithoutColors tests output without colors.
func TestSourceContext_WithoutColors(t *testing.T) {
	ctx := &SourceContext{
		Position:    Position{File: "test.c", Line: 1, Column: 5},
		LineContent: "int x = 5;",
		StartCol:    5,
		EndCol:      5,
		UseColors:   false,
	}

	result := ctx.String()

	// Should not contain ANSI codes
	if strings.Contains(result, "\033[") {
		t.Errorf("Result should not contain ANSI codes when UseColors=false: %q", result)
	}
}

// TestSourceContext_BeginningOfFile tests context at start of file.
func TestSourceContext_BeginningOfFile(t *testing.T) {
	ctx := &SourceContext{
		Position:    Position{File: "test.c", Line: 1, Column: 1},
		LineContent: "int main() {",
		StartCol:    1,
		EndCol:      1,
		BeforeLines: []string{}, // No lines before first line
		AfterLines:  []string{"    return 0;", "}"},
		UseColors:   false,
	}

	result := ctx.String()

	// Should have error line and after lines
	if !strings.Contains(result, "1 | int main() {") {
		t.Errorf("Result missing error line: %q", result)
	}
	if !strings.Contains(result, "2 |     return 0;") {
		t.Errorf("Result missing after line: %q", result)
	}
}

// TestSourceContext_EndOfFile tests context at end of file.
func TestSourceContext_EndOfFile(t *testing.T) {
	ctx := &SourceContext{
		Position:    Position{File: "test.c", Line: 3, Column: 1},
		LineContent: "}",
		StartCol:    1,
		EndCol:      1,
		BeforeLines: []string{"int main() {", "    return 0;"},
		AfterLines:  []string{}, // No lines after last line
		UseColors:   false,
	}

	result := ctx.String()

	// Should have before lines and error line
	if !strings.Contains(result, "1 | int main() {") {
		t.Errorf("Result missing before line: %q", result)
	}
	if !strings.Contains(result, "3 | }") {
		t.Errorf("Result missing error line: %q", result)
	}
}

// TestSourceContext_LineNumberAlignment tests proper alignment of line numbers.
func TestSourceContext_LineNumberAlignment(t *testing.T) {
	ctx := &SourceContext{
		Position:    Position{File: "test.c", Line: 100, Column: 5},
		LineContent: "printf(\"hello\");",
		StartCol:    5,
		EndCol:      5,
		BeforeLines: []string{"line 98", "line 99"},
		AfterLines:  []string{"line 101", "line 102"},
		UseColors:   false,
	}

	result := ctx.String()

	// All line numbers should be right-aligned with the same width
	// Max line number is 102 (3 digits), so all should be 3 chars wide
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.Contains(line, " | ") {
			parts := strings.SplitN(line, " | ", 2)
			if len(parts) > 0 {
				numPart := parts[0]
				// Should be 3 chars wide (including leading spaces for alignment)
				if len(numPart) != 3 && numPart != "" {
					t.Errorf("Line number not properly aligned: %q (len=%d, want 3) in line %q", numPart, len(numPart), line)
				}
			}
		}
	}
}

// TestSourceContext_EmptySurroundings tests with no surrounding lines.
func TestSourceContext_EmptySurroundings(t *testing.T) {
	ctx := &SourceContext{
		Position:    Position{File: "test.c", Line: 1, Column: 5},
		LineContent: "int x = 5;",
		StartCol:    5,
		EndCol:      5,
		BeforeLines: []string{},
		AfterLines:  []string{},
		UseColors:   false,
	}

	result := ctx.String()

	// Should only have error line and caret
	if !strings.Contains(result, "1 | int x = 5;") {
		t.Errorf("Result missing error line: %q", result)
	}
	if !strings.Contains(result, "^") {
		t.Errorf("Result missing caret: %q", result)
	}

	// Should not have extra blank lines
	lineCount := strings.Count(result, "\n")
	if lineCount > 1 {
		t.Errorf("Result has too many lines: %d, expected 1 newline", lineCount)
	}
}