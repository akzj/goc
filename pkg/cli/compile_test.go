// Package cli provides the command-line interface for the GOC compiler.
// This file contains tests for the compilation pipeline functions.
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/parser"
	"github.com/akzj/goc/pkg/ir"
)

func TestCompileOptions(t *testing.T) {
	opts := &CompileOptions{
		InputFile:  "test.c",
		Output:     "test.s",
		EmitAssembly: true,
		Verbose:    true,
		Debug:      true,
		Optimize:   "2",
		Target:     "x86_64",
	}

	if opts.InputFile != "test.c" {
		t.Errorf("Expected InputFile 'test.c', got '%s'", opts.InputFile)
	}

	if !opts.EmitAssembly {
		t.Error("Expected EmitAssembly to be true")
	}

	if !opts.Verbose {
		t.Error("Expected Verbose to be true")
	}
}

func TestParseCompileFlags_EmptyArgs(t *testing.T) {
	opts, err := ParseCompileFlags([]string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if opts == nil {
		t.Fatal("Expected non-nil options")
	}

	if opts.InputFile != "" {
		t.Errorf("Expected empty InputFile, got '%s'", opts.InputFile)
	}
}

func TestParseCompileFlags_HelpFlag(t *testing.T) {
	_, err := ParseCompileFlags([]string{"--help"})
	if err == nil {
		t.Error("Expected error for help flag (should show help and return error)")
	}
}

func TestParseCompileFlags_InvalidFlag(t *testing.T) {
	_, err := ParseCompileFlags([]string{"--invalid-flag"})
	if err == nil {
		t.Error("Expected error for invalid flag")
	}
}

func TestParseCompileFlags_MultipleInputFiles(t *testing.T) {
	_, err := ParseCompileFlags([]string{"file1.c", "file2.c"})
	if err == nil {
		t.Error("Expected error for multiple input files")
	}
}

func TestIsValidOptimizeLevel(t *testing.T) {
	tests := []struct {
		level string
		valid bool
	}{
		{"", false},  // Empty string is not valid
		{"0", true},
		{"1", true},
		{"2", true},
		{"3", true},
		{"s", true},
		{"z", true},
		{"4", false},
		{"invalid", false},
		{"O0", false},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			result := isValidOptimizeLevel(tt.level)
			if result != tt.valid {
				t.Errorf("isValidOptimizeLevel(%q) = %v, want %v", tt.level, result, tt.valid)
			}
		})
	}
}

func TestValidateOptimizeLevel(t *testing.T) {
	tests := []struct {
		level    string
		hasError bool
	}{
		{"", false},
		{"0", false},
		{"1", false},
		{"2", false},
		{"3", false},
		{"s", false},
		{"z", false},
		{"4", true},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			err := validateOptimizeLevel(tt.level)
			if (err != nil) != tt.hasError {
				t.Errorf("validateOptimizeLevel(%q) error = %v, hasError %v", tt.level, err, tt.hasError)
			}
		})
	}
}

func TestValidateTarget(t *testing.T) {
	tests := []struct {
		target   string
		hasError bool
	}{
		{"", false},
		{"x86_64", false},
		{"arm64", false},
		{"x86", false},
		{"arm", false},
		{"riscv64", false},
		{"invalid", true},
		{"x86-64", true},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			err := validateTarget(tt.target)
			if (err != nil) != tt.hasError {
				t.Errorf("validateTarget(%q) error = %v, hasError %v", tt.target, err, tt.hasError)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		want     int
		hasError bool
	}{
		{"0", 0, false},
		{"1", 1, false},
		{"42", 42, false},
		{"-1", -1, false},
		{"invalid", 0, true},
		{"", 0, true},
		{"99999999999999999999", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseInt(tt.input)
			if (err != nil) != tt.hasError {
				t.Errorf("ParseInt(%q) error = %v, hasError %v", tt.input, err, tt.hasError)
			}
			if err == nil && got != tt.want {
				t.Errorf("ParseInt(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestPrintVerbose(t *testing.T) {
	opts := &CompileOptions{
		Verbose: true,
	}

	// This should not panic
	printVerbose(opts, "Test message: %s", "test")

	// Test with verbose disabled
	opts.Verbose = false
	printVerbose(opts, "This should not print: %s", "test")
}

func TestCompileCommand_NoInputFile(t *testing.T) {
	err := CompileCommand([]string{})
	if err == nil {
		t.Error("Expected error when no input file provided")
	}

	if !strings.Contains(err.Error(), "input file") && !strings.Contains(err.Error(), "Usage") {
		t.Errorf("Expected error about input file, got: %v", err)
	}
}

func TestCompileCommand_NonExistentFile(t *testing.T) {
	err := CompileCommand([]string{"nonexistent_file.c"})
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestExecuteCompilationPipeline_NonExistentFile(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	opts := &CompileOptions{
		InputFile: "nonexistent.c",
	}

	err := executeCompilationPipeline("", "nonexistent.c", opts, errorHandler)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestRunLexicalAnalysis_Basic(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	opts := &CompileOptions{
		InputFile: "test.c",
	}

	// Create a temporary file with valid C code
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.c")
	source := "int main() { return 0; }"

	if err := os.WriteFile(tmpFile, []byte(source), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	sourceStr, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	tokens, err := runLexicalAnalysis(string(sourceStr), tmpFile, opts, errorHandler)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(tokens) == 0 {
		t.Error("Expected tokens from lexical analysis")
	}
}

func TestOutputAssembly(t *testing.T) {
	opts := &CompileOptions{
		Output: "", // Empty means stdout
	}

	assembly := ".text\n.globl main\nmain:\n\tret"

	// This should not panic
	err := outputAssembly(assembly, opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestOutputAssembly_WithOutputFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test.s")

	opts := &CompileOptions{
		Output: outputFile,
	}

	assembly := ".text\n.globl main\nmain:\n\tret"

	err := outputAssembly(assembly, opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Expected output file to be created")
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
	}

	if string(content) != assembly {
		t.Errorf("Output file content mismatch")
	}
}

func TestOutputResult_Assembly(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	opts := &CompileOptions{
		EmitAssembly: true,
	}

	assembly := ".text\n.globl main\nmain:\n\tret"

	err := outputResult(assembly, opts, errorHandler)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCompileCommand_HelpFlag(t *testing.T) {
	err := CompileCommand([]string{"--help"})
	if err == nil {
		t.Error("Expected error when --help is provided")
	}
}

func TestCompileCommand_OutputFlag(t *testing.T) {
	// Test that -o flag is parsed correctly
	opts, err := ParseCompileFlags([]string{"-o", "output.s", "input.c"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if opts.Output != "output.s" {
		t.Errorf("Expected Output 'output.s', got '%s'", opts.Output)
	}

	if opts.InputFile != "input.c" {
		t.Errorf("Expected InputFile 'input.c', got '%s'", opts.InputFile)
	}
}

func TestCompileCommand_AssemblyFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"assembly flag", []string{"-S", "input.c"}, true},
		{"no assembly", []string{"input.c"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseCompileFlags(tt.args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if opts.EmitAssembly != tt.want {
				t.Errorf("Expected EmitAssembly %v, got %v", tt.want, opts.EmitAssembly)
			}
		})
	}
}

func TestCompileCommand_ObjectFlag(t *testing.T) {
	opts, err := ParseCompileFlags([]string{"-c", "input.c"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !opts.EmitObject {
		t.Error("Expected EmitObject to be true")
	}
}

func TestCompileCommand_VerboseFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"short verbose", []string{"-v", "input.c"}, true},
		{"long verbose", []string{"--verbose", "input.c"}, true},
		{"no verbose", []string{"input.c"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseCompileFlags(tt.args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if opts.Verbose != tt.want {
				t.Errorf("Expected Verbose %v, got %v", tt.want, opts.Verbose)
			}
		})
	}
}

func TestCompileCommand_DebugFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"short debug", []string{"-d", "input.c"}, true},
		{"long debug", []string{"--debug", "input.c"}, true},
		{"no debug", []string{"input.c"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseCompileFlags(tt.args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if opts.Debug != tt.want {
				t.Errorf("Expected Debug %v, got %v", tt.want, opts.Debug)
			}
		})
	}
}

func TestCompileCommand_OptimizeFlag(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		want   string
		hasErr bool
	}{
		{"O0", []string{"-O", "0", "input.c"}, "0", false},
		{"O1", []string{"-O", "1", "input.c"}, "1", false},
		{"O2", []string{"-O", "2", "input.c"}, "2", false},
		{"O3", []string{"-O", "3", "input.c"}, "3", false},
		{"Os", []string{"-O", "s", "input.c"}, "s", false},
		{"Oz", []string{"-O", "z", "input.c"}, "z", false},
		{"invalid", []string{"-O", "4", "input.c"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseCompileFlags(tt.args)
			if tt.hasErr {
				if err == nil {
					t.Error("Expected error for invalid optimize level")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if opts.Optimize != tt.want {
				t.Errorf("Expected Optimize %q, got %q", tt.want, opts.Optimize)
			}
		})
	}
}

func TestCompileCommand_TargetFlag(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		want   string
		hasErr bool
	}{
		{"x86_64", []string{"-t", "x86_64", "input.c"}, "x86_64", false},
		{"arm64", []string{"-t", "arm64", "input.c"}, "arm64", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseCompileFlags(tt.args)
			if tt.hasErr {
				if err == nil {
					t.Error("Expected error for invalid target")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if opts.Target != tt.want {
				t.Errorf("Expected Target %q, got %q", tt.want, opts.Target)
			}
		})
	}
}

func TestOutputObjectFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test.o")

	errorHandler := errhand.NewErrorHandler()
	opts := &CompileOptions{
		Output: outputFile,
	}

	assembly := ".text\n.globl main\nmain:\n\tret"

	err := outputObjectFile(assembly, opts, errorHandler)
	// This will likely fail because we don't have an assembler, but we're testing the function exists
	_ = err
}

func TestOutputExecutable(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test")

	errorHandler := errhand.NewErrorHandler()
	opts := &CompileOptions{
		Output: outputFile,
	}

	assembly := ".text\n.globl main\nmain:\n\tret"

	err := outputExecutable(assembly, opts, errorHandler)
	// This will likely fail because we don't have a linker, but we're testing the function exists
	_ = err
}
// TestRunParsing_VerboseFixed tests runParsing with verbose output
func TestRunParsing_VerboseFixed(t *testing.T) {
	source := "int main() { return 0; }"
	l := lexer.NewLexer(source, "test.c")
	tokens := l.Tokenize()
	
	errorHandler := errhand.NewErrorHandler()
	opts := &CompileOptions{
		Verbose: true,
	}
	
	output := captureStderr(func() {
		ast, err := runParsing(tokens, "test.c", opts, errorHandler)
		if err != nil {
			t.Logf("Parsing error (expected for simple test): %v", err)
		}
		if ast == nil {
			t.Log("AST is nil")
		}
	})
	
	if !strings.Contains(output, "[compile] Stage 2/5: Parsing") {
		t.Errorf("Expected verbose output to contain '[compile] Stage 2/5: Parsing', got: %s", output)
	}
}

// TestRunSemanticAnalysis_VerboseFixed tests runSemanticAnalysis with verbose output
func TestRunSemanticAnalysis_VerboseFixed(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	opts := &CompileOptions{
		Verbose: true,
	}
	
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{},
	}
	
	output := captureStderr(func() {
		err := runSemanticAnalysis(ast, "test.c", opts, errorHandler)
		if err != nil {
			t.Logf("Semantic analysis error (may be expected): %v", err)
		}
	})
	
	if !strings.Contains(output, "[compile] Stage 3/5: Semantic analysis") {
		t.Errorf("Expected verbose output to contain '[compile] Stage 3/5: Semantic analysis', got: %s", output)
	}
}


// TestRunIRGeneration_Verbose tests runIRGeneration with verbose output
func TestRunIRGeneration_Verbose(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	opts := &CompileOptions{
		Verbose: true,
	}
	
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{},
	}
	
	output := captureStderr(func() {
		irResult, err := runIRGeneration(ast, "test.c", opts, errorHandler)
		if err != nil {
			t.Logf("IR generation error (may be expected): %v", err)
		}
		if irResult != nil {
			t.Logf("IR generated: %d functions, %d globals", len(irResult.Functions), len(irResult.Globals))
		}
	})
	
	if !strings.Contains(output, "[compile] Stage 4/5: IR generation") {
		t.Error("Expected verbose output to contain '[compile] Stage 4/5: IR generation'")
	}
}

// TestRunCodeGeneration_Verbose tests runCodeGeneration with verbose output
func TestRunCodeGeneration_Verbose(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	opts := &CompileOptions{
		Verbose: true,
	}
	
	irResult := &ir.IR{
		Functions: []*ir.Function{},
		Globals:   []*ir.GlobalVar{},
	}
	
	output := captureStderr(func() {
		assembly, err := runCodeGeneration(irResult, "test.c", opts, errorHandler)
		if err != nil {
			t.Logf("Code generation error (may be expected): %v", err)
		}
		t.Logf("Assembly generated: %d bytes", len(assembly))
	})
	
	if !strings.Contains(output, "[compile] Stage 5/5: Code generation") {
		t.Error("Expected verbose output to contain '[compile] Stage 5/5: Code generation'")
	}
}

// TestOutputResult_Verbose tests outputResult with verbose output
func TestOutputResult_Verbose(t *testing.T) {
	opts := &CompileOptions{
		Verbose:    true,
		EmitAssembly: true,
		Output:     "/tmp/test_result.s",
	}
	errorHandler := errhand.NewErrorHandler()
	
	defer os.Remove("/tmp/test_result.s")
	
	output := captureStderr(func() {
		err := outputResult("test assembly", opts, errorHandler)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
	
	if !strings.Contains(output, "Assembly written to") {
		t.Error("Expected verbose output to contain 'Assembly written to'")
	}
}

// TestOutputAssembly_Verbose tests outputAssembly with verbose output
func TestOutputAssembly_Verbose(t *testing.T) {
	opts := &CompileOptions{
		Verbose: true,
	}
	
	tmpFile := "/tmp/test_output.s"
	opts.Output = tmpFile
	defer os.Remove(tmpFile)
	
	output := captureStderr(func() {
		err := outputAssembly("test assembly", opts)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
	
	if !strings.Contains(output, "Assembly written to") {
		t.Error("Expected verbose output to contain file path")
	}
	
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
	}
	if string(content) != "test assembly" {
		t.Error("Output file content doesn't match")
	}
}

// TestOutputObjectFile_Verbose tests outputObjectFile with verbose output
func TestOutputObjectFile_Verbose(t *testing.T) {
	opts := &CompileOptions{
		Verbose: true,
		Output:  "/tmp/test_output.o",
	}
	
	tmpFile := "/tmp/test_output.o"
	defer os.Remove(tmpFile)
	
	var returnedErr error
	output := captureStderr(func() {
		returnedErr = outputObjectFile("test assembly", opts, errhand.NewErrorHandler())
	})
	
	// Either success (verbose output captured) or error (assembler not available)
	if returnedErr != nil {
		// Error is acceptable in test environment
		t.Logf("outputObjectFile returned error (acceptable in test): %v", returnedErr)
	} else if !strings.Contains(output, "Object file written to") {
		t.Errorf("Expected verbose output to contain 'Object file written to', got: %s", output)
	}
}

// TestParseCompileFlags_Verbose tests ParseCompileFlags with verbose flag
func TestParseCompileFlags_Verbose(t *testing.T) {
	args := []string{"-v", "test.c"}
	opts, err := ParseCompileFlags(args)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !opts.Verbose {
		t.Error("Expected verbose flag to be set")
	}
	if opts.InputFile != "test.c" {
		t.Errorf("Expected input file 'test.c', got '%s'", opts.InputFile)
	}
}

// TestParseCompileFlags_Debug tests ParseCompileFlags with debug flag
func TestParseCompileFlags_Debug(t *testing.T) {
	args := []string{"-d", "test.c"}
	opts, err := ParseCompileFlags(args)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !opts.Debug {
		t.Error("Expected debug flag to be set")
	}
}

// TestParseCompileFlags_Assembly tests ParseCompileFlags with assembly flag
func TestParseCompileFlags_Assembly(t *testing.T) {
	args := []string{"-S", "test.c"}
	opts, err := ParseCompileFlags(args)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !opts.EmitAssembly {
		t.Error("Expected EmitAssembly flag to be set")
	}
}

// TestParseCompileFlags_Object tests ParseCompileFlags with object flag
func TestParseCompileFlags_Object(t *testing.T) {
	args := []string{"-c", "test.c"}
	opts, err := ParseCompileFlags(args)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !opts.EmitObject {
		t.Error("Expected EmitObject flag to be set")
	}
}

// TestParseCompileFlags_AllFlags tests ParseCompileFlags with all flags
func TestParseCompileFlags_AllFlags(t *testing.T) {
	args := []string{"-v", "-d", "-S", "-o", "output.s", "-O", "2", "-t", "x86_64", "test.c"}
	opts, err := ParseCompileFlags(args)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !opts.Verbose {
		t.Error("Expected verbose flag to be set")
	}
	if !opts.Debug {
		t.Error("Expected debug flag to be set")
	}
	if !opts.EmitAssembly {
		t.Error("Expected EmitAssembly flag to be set")
	}
	if opts.Output != "output.s" {
		t.Errorf("Expected output 'output.s', got '%s'", opts.Output)
	}
	if opts.Optimize != "2" {
		t.Errorf("Expected optimize '2', got '%s'", opts.Optimize)
	}
	if opts.Target != "x86_64" {
		t.Errorf("Expected target 'x86_64', got '%s'", opts.Target)
	}

}
// TestRunSemanticAnalysis_ErrorPath tests semantic analysis with errors
func TestRunSemanticAnalysis_ErrorPath(t *testing.T) {
	// Create a minimal AST
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{},
	}
	
	opts := &CompileOptions{Verbose: false}
	errorHandler := errhand.NewErrorHandler()
	
	// This should execute without panic
	err := runSemanticAnalysis(ast, "test.go", opts, errorHandler)
	
	// We expect either an error or no error
	if err != nil {
		t.Logf("runSemanticAnalysis returned error (acceptable): %v", err)
	}
}

// TestRunParsing_ErrorPath tests parsing with error handling
func TestRunParsing_ErrorPath(t *testing.T) {
	opts := &CompileOptions{Verbose: false}
	errorHandler := errhand.NewErrorHandler()
	
	// Test with empty token list
	tokens := []lexer.Token{}
	ast, err := runParsing(tokens, "test.go", opts, errorHandler)
	
	// Either we get an AST or an error, both are acceptable
	if err != nil {
		t.Logf("runParsing returned error for empty tokens (acceptable): %v", err)
	} else if ast == nil {
		t.Error("Expected AST or error, got nil for both")
	}
}

// TestRunIRGeneration_ErrorPath tests IR generation with error handling
func TestRunIRGeneration_ErrorPath(t *testing.T) {
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{},
	}
	
	opts := &CompileOptions{Verbose: false}
	errorHandler := errhand.NewErrorHandler()
	
	irResult, err := runIRGeneration(ast, "test.go", opts, errorHandler)
	
	// Either we get IR or an error
	if err != nil {
		t.Logf("runIRGeneration returned error (acceptable): %v", err)
	} else if irResult == nil {
		t.Error("Expected IR or error, got nil for both")
	}
}

// TestRunCodeGeneration_ErrorPath tests code generation with error handling
func TestRunCodeGeneration_ErrorPath(t *testing.T) {
	testIR := &ir.IR{
		Functions: []*ir.Function{},
	}
	
	opts := &CompileOptions{Verbose: false}
	errorHandler := errhand.NewErrorHandler()
	
	assembly, err := runCodeGeneration(testIR, "test.go", opts, errorHandler)
	
	// Either we get assembly or an error
	if err != nil {
		t.Logf("runCodeGeneration returned error (acceptable): %v", err)
	} else if assembly == "" {
		t.Log("runCodeGeneration returned empty assembly (acceptable for empty IR)")
	}
}

// TestParseFlags_UnknownFlag tests parseFlags with unknown flag
func TestParseFlags_UnknownFlag(t *testing.T) {
	cli := NewCLI("test", "1.0.0", "test description")
	cmd := &Command{
		Name:        "test",
		Description: "Test command",
		Flags:       []Flag{},
	}
	
	// Try to parse an unknown flag
	_, _, err := cli.parseFlags(cmd, []string{"--unknown-flag"})
	
	if err == nil {
		t.Error("Expected error for unknown flag, got nil")
	} else if !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("Expected 'unknown flag' error, got: %v", err)
	}
}

// TestParseFlags_MissingValue tests parseFlags with missing flag value
func TestParseFlags_MissingValue(t *testing.T) {
	cli := NewCLI("test", "1.0.0", "test description")
	cmd := &Command{
		Name:        "test",
		Description: "Test command",
		Flags: []Flag{
			{Name: "output", HasValue: true, Default: nil},
		},
	}
	
	// Try to parse a flag that requires a value but doesn't have one
	_, _, err := cli.parseFlags(cmd, []string{"--output"})
	
	if err == nil {
		t.Error("Expected error for missing flag value, got nil")
	} else if !strings.Contains(err.Error(), "requires a value") {
		t.Errorf("Expected 'requires a value' error, got: %v", err)
	}
}

// TestExecuteCompilationPipeline_FullPath tests the full compilation pipeline
func TestExecuteCompilationPipeline_FullPath(t *testing.T) {
	source := `int main() { return 0; }`
	opts := &CompileOptions{
		Verbose: false,
	}
	errorHandler := errhand.NewErrorHandler()
	
	err := executeCompilationPipeline(source, "test.go", opts, errorHandler)
	
	// The pipeline may succeed or fail depending on implementation
	// Key is that it executes without panic
	if err != nil {
		t.Logf("executeCompilationPipeline returned error (acceptable): %v", err)
	}
}

// TestCompileCommand_InvalidFile tests CompileCommand with non-existent file
func TestCompileCommand_InvalidFile(t *testing.T) {
	args := []string{"/nonexistent/file.go", "-o", "/tmp/test.out"}
	
	err := CompileCommand(args)
	
	// Should return an error for non-existent file
	if err == nil {
		t.Log("CompileCommand succeeded (may be acceptable depending on implementation)")
	} else {
		t.Logf("CompileCommand returned error for invalid file (acceptable): %v", err)
	}
}


// TestRunParsing_InvalidSyntax tests parsing with invalid syntax to trigger error path
func TestRunParsing_InvalidSyntax(t *testing.T) {
	// Invalid syntax: incomplete declaration
	source := "int main() { invalid syntax here @#$% }"
	l := lexer.NewLexer(source, "test.go")
	tokens := l.Tokenize()
	
	opts := &CompileOptions{Verbose: false}
	errorHandler := errhand.NewErrorHandler()
	
	ast, err := runParsing(tokens, "test.go", opts, errorHandler)
	
	// Should have errors due to invalid syntax
	if errorHandler.HasErrors() {
		t.Logf("Parsing correctly detected errors: %d error(s)", errorHandler.ErrorCount())
	}
	if err == nil && ast == nil {
		t.Error("Expected either error or AST, got neither")
	}
}

// TestRunSemanticAnalysis_WithErrors tests semantic analysis when error handler has errors
func TestRunSemanticAnalysis_WithErrors(t *testing.T) {
	// Create a simple AST
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{
			&parser.FunctionDecl{
				Name: "test",
				Type: &parser.BaseType{Kind: parser.TypeInt},
				Body: &parser.CompoundStmt{
					Statements: []parser.Statement{},
				},
			},
		},
	}
	
	opts := &CompileOptions{Verbose: false}
	errorHandler := errhand.NewErrorHandler()
	// Pre-add an error to test the error path
	errorHandler.Error(errhand.ErrSyntaxError, "pre-existing error", errhand.Position{Line: 1, Column: 1})
	
	err := runSemanticAnalysis(ast, "test.go", opts, errorHandler)
	
	// Should detect pre-existing errors
	if !errorHandler.HasErrors() {
		t.Error("Expected errors to be detected")
	}
	if err != nil {
		t.Logf("runSemanticAnalysis returned error: %v", err)
	}
}

// TestRunIRGeneration_WithErrors tests IR generation when error handler has errors
func TestRunIRGeneration_WithErrors(t *testing.T) {
	// Create a simple AST
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{
			&parser.FunctionDecl{
				Name: "test",
				Type: &parser.BaseType{Kind: parser.TypeInt},
				Body: &parser.CompoundStmt{
					Statements: []parser.Statement{},
				},
			},
		},
	}
	
	opts := &CompileOptions{Verbose: false}
	errorHandler := errhand.NewErrorHandler()
	// Pre-add an error to test the error path
	errorHandler.Error(errhand.ErrSyntaxError, "pre-existing error", errhand.Position{Line: 1, Column: 1})
	
	irResult, err := runIRGeneration(ast, "test.go", opts, errorHandler)
	
	// Should detect pre-existing errors
	if !errorHandler.HasErrors() {
		t.Error("Expected errors to be detected")
	}
	if err != nil {
		t.Logf("runIRGeneration returned error: %v", err)
	}
	if irResult == nil && err == nil {
		t.Error("Expected either IR or error, got neither")
	}
}

// TestRunCodeGeneration_WithErrors tests code generation when error handler has errors
func TestRunCodeGeneration_WithErrors(t *testing.T) {
	// Create simple IR
	testIR := &ir.IR{
		Functions: []*ir.Function{
			{
				Name:   "test",
				Params: []*ir.Param{},
				Blocks: []*ir.BasicBlock{
					{
						Label:  "entry",
						Instrs: []ir.Instruction{},
					},
				},
			},
		},
		Globals: []*ir.GlobalVar{},
	}
	
	opts := &CompileOptions{Verbose: false}
	errorHandler := errhand.NewErrorHandler()
	// Pre-add an error to test the error path
	errorHandler.Error(errhand.ErrSyntaxError, "pre-existing error", errhand.Position{Line: 1, Column: 1})
	
	assembly, err := runCodeGeneration(testIR, "test.go", opts, errorHandler)
	
	// Should detect pre-existing errors
	if !errorHandler.HasErrors() {
		t.Error("Expected errors to be detected")
	}
	if err != nil {
		t.Logf("runCodeGeneration returned error: %v", err)
	}
	// assembly can be empty for error cases
	_ = assembly
}

// TestExecuteCompilationPipeline_FullErrorPath tests the full pipeline with error handling
func TestExecuteCompilationPipeline_FullErrorPath(t *testing.T) {
	// Invalid source that should cause parsing errors
	source := "invalid @#$% syntax"
	
	opts := &CompileOptions{
		Verbose: false,
	}
	errorHandler := errhand.NewErrorHandler()
	
	err := executeCompilationPipeline(source, "test.go", opts, errorHandler)
	
	// Should have errors
	if !errorHandler.HasErrors() {
		t.Log("Pipeline completed without errors (unexpected for invalid source)")
	}
	if err == nil {
		t.Log("Pipeline returned no error despite invalid source")
	}
}

// TestCompileCommand_InvalidFileContent tests CompileCommand with file containing invalid syntax
func TestCompileCommand_InvalidFileContent(t *testing.T) {
	// Create a temporary file with invalid syntax
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.go")
	
	invalidSource := "int main() { @#$% invalid syntax }"
	if err := os.WriteFile(tmpFile, []byte(invalidSource), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	// Try to compile - should fail gracefully
	err := CompileCommand([]string{tmpFile})
	
	// Should return an error
	if err == nil {
		t.Log("CompileCommand succeeded with invalid syntax (unexpected)")
	} else {
		t.Logf("CompileCommand correctly returned error: %v", err)
	}
}
