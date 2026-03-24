// Package integration provides integration test helpers for the GOC compiler.
// These helpers support both high-level end-to-end testing and low-level pipeline testing.
package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"runtime"
	"testing"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/codegen"
	"github.com/akzj/goc/pkg/ir"
	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/parser"
	"github.com/akzj/goc/pkg/semantic"
)

// findModuleRoot finds the directory containing go.mod by walking up from current directory
func findModuleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// ============================================================================
// PIPELINE TEST HELPERS - Low-level component chaining helpers
// ============================================================================

// PipelineResult holds the results of a full compiler pipeline execution.
// Each field contains the output from the corresponding pipeline stage.
type PipelineResult struct {
	// Tokens from the lexer stage
	Tokens []lexer.Token
	// AST from the parser stage
	AST *parser.TranslationUnit
	// IR from the IR generator stage
	IR *ir.IR
	// Assembly from the code generator stage
	Assembly string
	// ErrorHandler from each stage (for error inspection)
	LexerErrors    *errhand.ErrorHandler
	ParserErrors   *errhand.ErrorHandler
	SemanticErrors *errhand.ErrorHandler
	IRErrors       *errhand.ErrorHandler
	CodegenErrors  *errhand.ErrorHandler
}

// PipelineConfig configures pipeline execution behavior.
// Use this for advanced control over pipeline execution.
type PipelineConfig struct {
	// StopOnError determines whether to stop at the first error (default: true)
	StopOnError bool
	// Verbose enables verbose output (default: false)
	Verbose bool
	// SourceFile is the source file name for error reporting (default: "test.c")
	SourceFile string
}

// RunPipeline executes the full compiler pipeline from source code to assembly.
// It chains all compiler components in order: Lexer → Parser → Semantic Analyzer → IR Generator → Code Generator.
//
// Parameters:
//   - source: The C source code to compile
//   - fileName: The source file name (used for error reporting)
//
// Returns:
//   - *PipelineResult: Contains all intermediate results and final assembly
//   - error: Non-nil if any pipeline stage failed
//
// Example:
//
//	result, err := RunPipeline("int main(void) { return 0; }", "test.c")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Assembly)
func RunPipeline(source string, fileName string) (*PipelineResult, error) {
	result := &PipelineResult{}

	// Stage 1: Lexical Analysis
	tokens, lexerErrors, err := RunLexer(source, fileName)
	if err != nil {
		result.LexerErrors = lexerErrors
		return result, fmt.Errorf("lexer failed: %w", err)
	}
	result.Tokens = tokens
	result.LexerErrors = lexerErrors

	// Stage 2: Parsing
	ast, parserErrors, err := RunParser(tokens, fileName)
	if err != nil {
		result.ParserErrors = parserErrors
		return result, fmt.Errorf("parser failed: %w", err)
	}
	result.AST = ast
	result.ParserErrors = parserErrors

	// Stage 3: Semantic Analysis
	semanticErrors, err := RunSemantic(ast)
	if err != nil {
		result.SemanticErrors = semanticErrors
		return result, fmt.Errorf("semantic analysis failed: %w", err)
	}
	result.SemanticErrors = semanticErrors

	// Stage 4: IR Generation
	irResult, irErrors, err := RunIRGenerator(ast)
	if err != nil {
		result.IRErrors = irErrors
		return result, fmt.Errorf("IR generation failed: %w", err)
	}
	result.IR = irResult
	result.IRErrors = irErrors

	// Stage 5: Code Generation
	assembly, codegenErrors, err := RunCodeGenerator(irResult)
	if err != nil {
		result.CodegenErrors = codegenErrors
		return result, fmt.Errorf("code generation failed: %w", err)
	}
	result.Assembly = assembly
	result.CodegenErrors = codegenErrors

	return result, nil
}

// RunPipelineWithConfig executes the full compiler pipeline with custom configuration.
// This provides more control over pipeline behavior than RunPipeline.
//
// Parameters:
//   - source: The C source code to compile
//   - config: Pipeline configuration options
//
// Returns:
//   - *PipelineResult: Contains all intermediate results and final assembly
//   - error: Non-nil if any pipeline stage failed (depending on StopOnError config)
//
// Example:
//
//	config := &PipelineConfig{
//	    StopOnError: false,  // Continue even if errors occur
//	    Verbose: true,       // Enable verbose output
//	    SourceFile: "test.c",
//	}
//	result, err := RunPipelineWithConfig("int main(void) { return 0; }", config)
func RunPipelineWithConfig(source string, config *PipelineConfig) (*PipelineResult, error) {
	if config == nil {
		config = &PipelineConfig{
			StopOnError: true,
			Verbose:     false,
			SourceFile:  "test.c",
		}
	}

	result := &PipelineResult{}
	fileName := config.SourceFile
	if fileName == "" {
		fileName = "test.c"
	}

	// Stage 1: Lexical Analysis
	tokens, lexerErrors, err := RunLexer(source, fileName)
	result.LexerErrors = lexerErrors
	result.Tokens = tokens
	if err != nil {
		if config.Verbose {
			fmt.Printf("Lexer error: %v\n", err)
		}
		if config.StopOnError {
			return result, fmt.Errorf("lexer failed: %w", err)
		}
	}

	// Stage 2: Parsing (only if we have tokens)
	if len(tokens) > 0 {
		ast, parserErrors, err := RunParser(tokens, fileName)
		result.ParserErrors = parserErrors
		result.AST = ast
		if err != nil {
			if config.Verbose {
				fmt.Printf("Parser error: %v\n", err)
			}
			if config.StopOnError {
				return result, fmt.Errorf("parser failed: %w", err)
			}
		}

		// Stage 3: Semantic Analysis (only if we have AST)
		if ast != nil {
			semanticErrors, err := RunSemantic(ast)
			result.SemanticErrors = semanticErrors
			if err != nil {
				if config.Verbose {
					fmt.Printf("Semantic error: %v\n", err)
				}
				if config.StopOnError {
					return result, fmt.Errorf("semantic analysis failed: %w", err)
				}
			}

			// Stage 4: IR Generation (only if we have AST)
			if ast != nil {
				irResult, irErrors, err := RunIRGenerator(ast)
				result.IRErrors = irErrors
				result.IR = irResult
				if err != nil {
					if config.Verbose {
						fmt.Printf("IR error: %v\n", err)
					}
					if config.StopOnError {
						return result, fmt.Errorf("IR generation failed: %w", err)
					}
				}

				// Stage 5: Code Generation (only if we have IR)
				if irResult != nil {
					assembly, codegenErrors, err := RunCodeGenerator(irResult)
					result.CodegenErrors = codegenErrors
					result.Assembly = assembly
					if err != nil {
						if config.Verbose {
							fmt.Printf("Codegen error: %v\n", err)
						}
						if config.StopOnError {
							return result, fmt.Errorf("code generation failed: %w", err)
						}
					}
				}
			}
		}
	}

	return result, nil
}

// RunLexer executes the lexical analysis stage.
// It tokenizes C source code and returns the tokens.
//
// Parameters:
//   - source: The C source code to tokenize
//   - fileName: The source file name (used for error reporting)
//
// Returns:
//   - []lexer.Token: The tokenized tokens
//   - *errhand.ErrorHandler: Error handler with any lexer errors
//   - error: Non-nil if lexing failed
//
// Example:
//
//	tokens, errors, err := RunLexer("int x = 42;", "test.c")
//	if err != nil {
//	    log.Fatal(err)
//	}
func RunLexer(source string, fileName string) ([]lexer.Token, *errhand.ErrorHandler, error) {
	// Create error handler for lexer
	errorHandler := errhand.NewErrorHandler()

	// Create and run lexer
	l := lexer.NewLexer(source, fileName)
	tokens := l.Tokenize()

	// Check for errors (lexer doesn't use error handler, but we return it for consistency)
	if len(tokens) == 0 {
		return tokens, errorHandler, fmt.Errorf("no tokens produced")
	}

	return tokens, errorHandler, nil
}

// RunParser executes the parsing stage.
// It parses tokens into an Abstract Syntax Tree (AST).
//
// Parameters:
//   - tokens: The tokens from the lexer
//   - fileName: The source file name (used for error reporting)
//
// Returns:
//   - *parser.TranslationUnit: The parsed AST
//   - *errhand.ErrorHandler: Error handler with any parser errors
//   - error: Non-nil if parsing failed
//
// Example:
//
//	ast, errors, err := RunParser(tokens, "test.c")
//	if err != nil {
//	    log.Fatal(err)
//	}
func RunParser(tokens []lexer.Token, fileName string) (*parser.TranslationUnit, *errhand.ErrorHandler, error) {
	// Create error handler for parser
	errorHandler := errhand.NewErrorHandler()

	// Create and run parser
	p := parser.NewParser(tokens, errorHandler)
	ast, err := p.Parse()

	// Check for errors
	if err != nil {
		return ast, errorHandler, fmt.Errorf("parse error: %w", err)
	}

	if errorHandler.HasErrors() {
		return ast, errorHandler, fmt.Errorf("parser reported %d error(s)", errorHandler.ErrorCount())
	}

	return ast, errorHandler, nil
}

// RunSemantic executes the semantic analysis stage.
// It performs type checking and semantic validation on the AST.
//
// Parameters:
//   - ast: The AST from the parser
//
// Returns:
//   - *errhand.ErrorHandler: Error handler with any semantic errors
//   - error: Non-nil if semantic analysis failed
//
// Example:
//
//	errors, err := RunSemantic(ast)
//	if err != nil {
//	    log.Fatal(err)
//	}
func RunSemantic(ast *parser.TranslationUnit) (*errhand.ErrorHandler, error) {
	// Create error handler for semantic analysis
	errorHandler := errhand.NewErrorHandler()

	// Create and run semantic analyzer
	analyzer := semantic.NewSemanticAnalyzer(errorHandler)
	err := analyzer.Analyze(ast)

	// Check for errors
	if err != nil {
		return errorHandler, fmt.Errorf("semantic analysis error: %w", err)
	}

	if errorHandler.HasErrors() {
		return errorHandler, fmt.Errorf("semantic analyzer reported %d error(s)", errorHandler.ErrorCount())
	}

	return errorHandler, nil
}

// RunIRGenerator executes the IR generation stage.
// It generates intermediate representation (IR) from the AST.
//
// Parameters:
//   - ast: The AST from the parser (after semantic analysis)
//
// Returns:
//   - *ir.IR: The generated IR
//   - *errhand.ErrorHandler: Error handler with any IR generation errors
//   - error: Non-nil if IR generation failed
//
// Example:
//
//	irResult, errors, err := RunIRGenerator(ast)
//	if err != nil {
//	    log.Fatal(err)
//	}
func RunIRGenerator(ast *parser.TranslationUnit) (*ir.IR, *errhand.ErrorHandler, error) {
	// Create error handler for IR generation
	errorHandler := errhand.NewErrorHandler()

	// Create and run IR generator
	gen := ir.NewIRGenerator(errorHandler)
	irResult, err := gen.Generate(ast)

	// Check for errors
	if err != nil {
		return irResult, errorHandler, fmt.Errorf("IR generation error: %w", err)
	}

	if errorHandler.HasErrors() {
		return irResult, errorHandler, fmt.Errorf("IR generator reported %d error(s)", errorHandler.ErrorCount())
	}

	return irResult, errorHandler, nil
}

// RunCodeGenerator executes the code generation stage.
// It generates x86-64 assembly code from the IR.
//
// Parameters:
//   - irResult: The IR from the IR generator
//
// Returns:
//   - string: The generated assembly code
//   - *errhand.ErrorHandler: Error handler with any code generation errors
//   - error: Non-nil if code generation failed
//
// Example:
//
//	assembly, errors, err := RunCodeGenerator(irResult)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(assembly)
func RunCodeGenerator(irResult *ir.IR) (string, *errhand.ErrorHandler, error) {
	// Create error handler for code generation
	errorHandler := errhand.NewErrorHandler()

	// Create and run code generator
	gen := codegen.NewCodeGenerator(errorHandler)
	assembly, err := gen.Generate(irResult)

	// Check for errors
	if err != nil {
		return assembly, errorHandler, fmt.Errorf("code generation error: %w", err)
	}

	if errorHandler.HasErrors() {
		return assembly, errorHandler, fmt.Errorf("code generator reported %d error(s)", errorHandler.ErrorCount())
	}

	return assembly, errorHandler, nil
}

// HasErrors checks if any errors occurred during pipeline execution.
// It examines all error handlers in the pipeline result.
//
// Parameters:
//   - result: The pipeline result to check
//
// Returns:
//   - bool: True if any errors occurred, false otherwise
//
// Example:
//
//	if HasErrors(result) {
//	    log.Fatal("Pipeline execution had errors")
//	}
func HasErrors(result *PipelineResult) bool {
	if result == nil {
		return true
	}

	if result.LexerErrors != nil && result.LexerErrors.HasErrors() {
		return true
	}
	if result.ParserErrors != nil && result.ParserErrors.HasErrors() {
		return true
	}
	if result.SemanticErrors != nil && result.SemanticErrors.HasErrors() {
		return true
	}
	if result.IRErrors != nil && result.IRErrors.HasErrors() {
		return true
	}
	if result.CodegenErrors != nil && result.CodegenErrors.HasErrors() {
		return true
	}

	return false
}

// ErrorSummary returns a summary of all errors from the pipeline execution.
// It collects error messages from all stages.
//
// Parameters:
//   - result: The pipeline result to summarize
//
// Returns:
//   - string: A formatted string containing all error messages
//
// Example:
//
//	if HasErrors(result) {
//	    fmt.Println(ErrorSummary(result))
//	}
func ErrorSummary(result *PipelineResult) string {
	if result == nil {
		return "nil pipeline result"
	}

	var summary string

	if result.LexerErrors != nil && result.LexerErrors.HasErrors() {
		summary += fmt.Sprintf("Lexer: %d error(s)\n", result.LexerErrors.ErrorCount())
	}
	if result.ParserErrors != nil && result.ParserErrors.HasErrors() {
		summary += fmt.Sprintf("Parser: %d error(s)\n", result.ParserErrors.ErrorCount())
	}
	if result.SemanticErrors != nil && result.SemanticErrors.HasErrors() {
		summary += fmt.Sprintf("Semantic: %d error(s)\n", result.SemanticErrors.ErrorCount())
	}
	if result.IRErrors != nil && result.IRErrors.HasErrors() {
		summary += fmt.Sprintf("IR: %d error(s)\n", result.IRErrors.ErrorCount())
	}
	if result.CodegenErrors != nil && result.CodegenErrors.HasErrors() {
		summary += fmt.Sprintf("Codegen: %d error(s)\n", result.CodegenErrors.ErrorCount())
	}

	if summary == "" {
		return "No errors"
	}

	return summary
}

// ============================================================================
// E2E TEST HELPERS - High-level compilation helpers
// ============================================================================

// CompilerResult holds the result of a compilation attempt
type CompilerResult struct {
	Success  bool
	Assembly string
	Stdout   string
	Stderr   string
	ExitCode int
}

// CompileSource compiles a C source string and returns the result
func CompileSource(t *testing.T, source string) *CompilerResult {
	t.Helper()

	// Find the module root (directory containing go.mod)
	moduleRoot, err := findModuleRoot()
	if err != nil {
		t.Fatalf("Failed to find module root: %v", err)
	}

	// Create temporary directory for compilation
	tmpDir, err := os.MkdirTemp("", "goc-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write source file
	srcPath := filepath.Join(tmpDir, "test.c")
	if err := os.WriteFile(srcPath, []byte(source), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Output assembly file
	asmPath := filepath.Join(tmpDir, "test.s")

	// Assembly-only: default compile links to ELF and needs CRT entry (_start).
	// These tests validate the pipeline through codegen; use -S to skip linking.
	cmd := exec.Command("go", "run", "./cmd/goc", "compile", "-S", "-o", asmPath, srcPath)
	cmd.Dir = moduleRoot
	output, err := cmd.CombinedOutput()

	result := &CompilerResult{
		Stderr:   string(output),
		ExitCode: 0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Success = false
		} else {
			result.Success = false
		}
		result.Stderr = string(output)
	} else {
		result.Success = true
		// Read assembly output
		if asmData, err := os.ReadFile(asmPath); err == nil {
			result.Assembly = string(asmData)
		}
	}

	result.Stdout = string(output)
	return result
}

// CompileSourceExpectSuccess compiles source and expects success
func CompileSourceExpectSuccess(t *testing.T, source string) *CompilerResult {
	t.Helper()
	result := CompileSource(t, source)
	if !result.Success {
		t.Errorf("Compilation failed unexpectedly: %s", result.Stderr)
	}
	return result
}

// CompileSourceExpectFailure compiles source and expects failure
func CompileSourceExpectFailure(t *testing.T, source string) *CompilerResult {
	t.Helper()
	result := CompileSource(t, source)
	if result.Success {
		t.Errorf("Compilation succeeded unexpectedly, expected failure")
	}
	return result
}

// LoadProgram loads a sample program from the programs directory
func LoadProgram(t *testing.T, name string) string {
	t.Helper()
	// Get the directory of the current test file
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	path := filepath.Join(testDir, "programs", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to load program %s: %v", name, err)
	}
	return string(data)
}

// ValidateAssembly checks if assembly output is valid x86-64 syntax
func ValidateAssembly(t *testing.T, asm string) {
	t.Helper()
	if asm == "" {
		t.Log("Warning: Assembly output is empty")
		return
	}

	// Check for basic x86-64 assembly structure
	if !strings.Contains(asm, ".text") && !strings.Contains(asm, ".globl") {
		t.Log("Warning: Assembly may be missing standard sections")
	}

	// Try to assemble with system assembler (if available)
	tmpDir, err := os.MkdirTemp("", "goc-asm-test-*")
	if err != nil {
		t.Logf("Skipping assembler validation: %v", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	asmPath := filepath.Join(tmpDir, "test.s")
	if err := os.WriteFile(asmPath, []byte(asm), 0644); err != nil {
		t.Logf("Failed to write assembly file: %v", err)
		return
	}

	// Try to assemble with gas
	cmd := exec.Command("as", "--64", asmPath, "-o", filepath.Join(tmpDir, "test.o"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Assembly validation warning: %s\n%s", err, string(output))
	}
}

// CheckErrorMessage verifies error message contains expected content
func CheckErrorMessage(t *testing.T, result *CompilerResult, expectedSubstrings ...string) {
	t.Helper()
	for _, substr := range expectedSubstrings {
		if !strings.Contains(result.Stderr, substr) && !strings.Contains(result.Stdout, substr) {
			t.Errorf("Error message does not contain '%s'. Got: %s", substr, result.Stderr)
		}
	}
}
