package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/akzj/goc/pkg/lexer"
)

// TestOutputTokensJSON tests the outputTokensJSON function
func TestOutputTokensJSON(t *testing.T) {
	// Test case 1: Normal tokens with EOF filtering
	t.Run("NormalTokensWithEOF", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		tokens := []lexer.Token{
			{Type: lexer.IDENT, Value: "test", Pos: lexer.Position{File: "test.go", Line: 1, Column: 1}},
			{Type: lexer.EOF, Value: "", Pos: lexer.Position{File: "test.go", Line: 1, Column: 5}},
		}

		outputTokensJSON(tokens)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Verify EOF token is not in output
		if strings.Contains(output, "EOF") {
			t.Error("EOF token should be filtered out")
		}
		if !strings.Contains(output, "IDENT") {
			t.Error("IDENT token should be in output")
		}
	})

	// Test case 2: Empty tokens
	t.Run("EmptyTokens", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		tokens := []lexer.Token{}

		outputTokensJSON(tokens)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Should output empty array
		if !strings.Contains(output, "[]") {
			t.Error("Empty tokens should output empty array")
		}
	})

	// Test case 3: Only EOF token
	t.Run("OnlyEOFToken", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		tokens := []lexer.Token{
			{Type: lexer.EOF, Value: "", Pos: lexer.Position{File: "test.go", Line: 1, Column: 1}},
		}

		outputTokensJSON(tokens)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Should output empty array since EOF is filtered
		if !strings.Contains(output, "[]") {
			t.Error("Only EOF token should output empty array")
		}
	})
}

// TestOutputTokensCompact tests the outputTokensCompact function
func TestOutputTokensCompact(t *testing.T) {
	// Test case 1: Normal tokens with EOF filtering
	t.Run("NormalTokensWithEOF", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		tokens := []lexer.Token{
			{Type: lexer.IDENT, Value: "test", Pos: lexer.Position{File: "test.go", Line: 1, Column: 1}, HasSpace: true},
			{Type: lexer.EOF, Value: "", Pos: lexer.Position{File: "test.go", Line: 1, Column: 5}},
		}

		outputTokensCompact(tokens)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Verify EOF token is not in output
		if strings.Contains(output, "EOF") {
			t.Error("EOF token should be filtered out")
		}
		if !strings.Contains(output, "IDENT") {
			t.Error("IDENT token should be in output")
		}
	})

	// Test case 2: Empty tokens
	t.Run("EmptyTokens", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		tokens := []lexer.Token{}

		outputTokensCompact(tokens)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Should output nothing or minimal output for empty tokens
		if strings.TrimSpace(output) != "" {
			t.Error("Empty tokens should produce no output")
		}
	})

	// Test case 3: Token with empty value
	t.Run("TokenWithEmptyValue", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		tokens := []lexer.Token{
			{Type: lexer.IDENT, Value: "", Pos: lexer.Position{File: "test.go", Line: 1, Column: 1}, HasSpace: false},
		}

		outputTokensCompact(tokens)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Should still output the token even with empty value
		if !strings.Contains(output, "IDENT") {
			t.Error("Token with empty value should still be output")
		}
	})
}

// TestTokenJSONSerialization tests the TokenJSON struct serialization
func TestTokenJSONSerialization(t *testing.T) {
	token := TokenJSON{
		Type:     "IDENT",
		Value:    "test",
		File:     "test.go",
		Line:     1,
		Column:   1,
		HasSpace: true,
	}

	data, err := json.Marshal(token)
	if err != nil {
		t.Fatalf("Failed to marshal TokenJSON: %v", err)
	}

	var result TokenJSON
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal TokenJSON: %v", err)
	}

	if result.Type != token.Type {
		t.Errorf("Type mismatch: expected %s, got %s", token.Type, result.Type)
	}
	if result.Value != token.Value {
		t.Errorf("Value mismatch: expected %s, got %s", token.Value, result.Value)
	}
	if result.HasSpace != token.HasSpace {
		t.Errorf("HasSpace mismatch: expected %v, got %v", token.HasSpace, result.HasSpace)
	}
}

// TestOutputTokensErrorHandling tests error handling in token output
func TestOutputTokensErrorHandling(t *testing.T) {
	// Test with nil writer to trigger error
	t.Run("NilWriterError", func(t *testing.T) {
		oldStdout := os.Stdout
		oldExitFunc := exitFunc
		exitCalled := false
		
		exitFunc = func(code int) {
			exitCalled = true
			panic("exit called")
		}
		
		os.Stdout = nil // This will cause an error

		// Should call exitFunc but not panic
		defer func() {
			os.Stdout = oldStdout
			exitFunc = oldExitFunc
			if r := recover(); r != "exit called" {
				t.Errorf("Function should call exit on error: %v", r)
			}
			if !exitCalled {
				t.Error("exitFunc should have been called")
			}
		}()

		tokens := []lexer.Token{
			{Type: lexer.IDENT, Value: "test", Pos: lexer.Position{File: "test.go", Line: 1, Column: 1}},
		}

		// This should handle the error gracefully by calling exitFunc
		outputTokensJSON(tokens)
	})
}

// TestOutputTokensCompactFormat tests the compact output format
func TestOutputTokensCompactFormat(t *testing.T) {
	t.Run("CompactFormatStructure", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		tokens := []lexer.Token{
			{Type: lexer.IDENT, Value: "func", Pos: lexer.Position{File: "test.go", Line: 1, Column: 1}, HasSpace: true},
			{Type: lexer.IDENT, Value: "main", Pos: lexer.Position{File: "test.go", Line: 1, Column: 6}, HasSpace: false},
		}

		outputTokensCompact(tokens)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Compact format should include token info in a condensed format
		if !strings.Contains(output, "func") && !strings.Contains(output, "main") {
			t.Error("Compact output should contain token values")
		}
	})
}

// TestOutputTokensJSONFormat tests the JSON output format structure
func TestOutputTokensJSONFormat(t *testing.T) {
	t.Run("JSONFormatValid", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		tokens := []lexer.Token{
			{Type: lexer.IDENT, Value: "test", Pos: lexer.Position{File: "test.go", Line: 1, Column: 1}, HasSpace: true},
		}

		outputTokensJSON(tokens)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Verify it's valid JSON
		var result []TokenJSON
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Errorf("Output is not valid JSON: %v", err)
		}

		if len(result) != 1 {
			t.Errorf("Expected 1 token, got %d", len(result))
		}
	})
}

// BenchmarkOutputTokensJSON benchmarks the JSON output function
func BenchmarkOutputTokensJSON(b *testing.B) {
	tokens := make([]lexer.Token, 100)
	for i := range tokens {
		tokens[i] = lexer.Token{
			Type:     lexer.IDENT,
			Value:    "token",
			Pos:      lexer.Position{File: "test.go", Line: 1, Column: 1},
			HasSpace: true,
		}
	}

	oldStdout := os.Stdout
	for i := 0; i < b.N; i++ {
		r, w, _ := os.Pipe()
		os.Stdout = w
		outputTokensJSON(tokens)
		w.Close()
		io.Copy(io.Discard, r)
		os.Stdout = oldStdout
	}
}

// BenchmarkOutputTokensCompact benchmarks the compact output function
func BenchmarkOutputTokensCompact(b *testing.B) {
	tokens := make([]lexer.Token, 100)
	for i := range tokens {
		tokens[i] = lexer.Token{
			Type:     lexer.IDENT,
			Value:    "token",
			Pos:      lexer.Position{File: "test.go", Line: 1, Column: 1},
			HasSpace: true,
		}
	}

	oldStdout := os.Stdout
	for i := 0; i < b.N; i++ {
		r, w, _ := os.Pipe()
		os.Stdout = w
		outputTokensCompact(tokens)
		w.Close()
		io.Copy(io.Discard, r)
		os.Stdout = oldStdout
	}
}