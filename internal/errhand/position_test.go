package errhand

import (
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