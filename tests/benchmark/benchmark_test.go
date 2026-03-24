// Package benchmark provides benchmark tests for the GOC compiler.
// This framework measures performance across different compiler stages:
// parsing, semantic analysis, IR generation, and code generation.
//
// # Running Benchmarks
//
// To run all benchmarks:
//
//	go test -bench=. -benchmem ./tests/benchmark/
//
// To run specific benchmarks:
//
//	# Parser benchmarks only
//	go test -bench=BenchmarkParser -benchmem ./tests/benchmark/
//
//	# Semantic analyzer benchmarks only
//	go test -bench=BenchmarkSemantic -benchmem ./tests/benchmark/
//
//	# IR generator benchmarks only
//	go test -bench=BenchmarkIR -benchmem ./tests/benchmark/
//
//	# All compiler stage benchmarks
//	go test -bench=BenchmarkCompile -benchmem ./tests/benchmark/
//
// # Benchmark Programs
//
// The benchmark suite includes programs of varying complexity:
//   - simple.c: Basic variable declarations and arithmetic (low complexity)
//   - moderate.c: Functions, loops, conditionals, arrays (medium complexity)
//   - complex.c: Structs, pointers, recursion, multiple functions (high complexity)
//
// # Output Interpretation
//
// Benchmark output includes:
//   - ns/op: Nanoseconds per operation (lower is better)
//   - B/op: Bytes allocated per operation (lower is better)
//   - allocs/op: Allocations per operation (lower is better)
//
// Example output:
//
//	BenchmarkParserSimple-8        10000    125000 ns/op    50000 B/op    500 allocs/op
//	BenchmarkParserModerate-8       5000    350000 ns/op   150000 B/op   1500 allocs/op
//	BenchmarkParserComplex-8        2000    850000 ns/op   350000 B/op   3500 allocs/op
package benchmark

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/codegen"
	"github.com/akzj/goc/pkg/ir"
	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/parser"
	"github.com/akzj/goc/pkg/semantic"
)

// BenchmarkProgram represents a program to benchmark with metadata.
type BenchmarkProgram struct {
	Name        string // Program name (e.g., "simple", "moderate", "complex")
	FileName    string // File name in programs directory (e.g., "simple.c")
	Description string // Description of program complexity
}

// benchmarkPrograms defines the set of programs to benchmark.
var benchmarkPrograms = []BenchmarkProgram{
	{
		Name:        "Simple",
		FileName:    "simple.c",
		Description: "Basic variable declarations and arithmetic",
	},
	{
		Name:        "Moderate",
		FileName:    "moderate.c",
		Description: "Functions, loops, conditionals, arrays",
	},
	{
		Name:        "Complex",
		FileName:    "complex.c",
		Description: "Structs, pointers, recursion, multiple functions",
	},
}

// findBenchmarkRoot finds the directory containing the benchmark tests.
func findBenchmarkRoot() (string, error) {
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
			return "", err
		}
		dir = parent
	}
}

// loadBenchmarkProgram loads a benchmark program source code.
func loadBenchmarkProgram(b *testing.B, fileName string) string {
	b.Helper()

	root, err := findBenchmarkRoot()
	if err != nil {
		b.Fatalf("Failed to find benchmark root: %v", err)
	}

	programPath := filepath.Join(root, "tests", "benchmark", "programs", fileName)
	data, err := os.ReadFile(programPath)
	if err != nil {
		b.Fatalf("Failed to load program %s: %v", fileName, err)
	}

	return string(data)
}

// STAGE BENCHMARK HELPERS

// benchmarkLexer benchmarks the lexical analysis stage.
func benchmarkLexer(b *testing.B, source string) {
	b.Helper()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Reset between iterations
		b.StartTimer()

		l := lexer.NewLexer(source, "benchmark.c")
		tokens := l.Tokenize()

		// Prevent optimization
		if len(tokens) == 0 {
			b.Fatal("No tokens produced")
		}
	}
}

// benchmarkParser benchmarks the parsing stage.
func benchmarkParser(b *testing.B, source string) {
	b.Helper()

	// Tokenize once (not part of benchmark)
	l := lexer.NewLexer(source, "benchmark.c")
	tokens := l.Tokenize()
	if len(tokens) == 0 {
		b.Fatal("No tokens produced")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		errorHandler := errhand.NewErrorHandler()
		b.StartTimer()

		p := parser.NewParser(tokens, errorHandler)
		ast, err := p.Parse()

		// Prevent optimization
		if err != nil && ast == nil {
			// Some errors are expected for certain programs
			continue
		}
	}
}

// benchmarkSemantic benchmarks the semantic analysis stage.
func benchmarkSemantic(b *testing.B, source string) {
	b.Helper()

	// Parse once (not part of benchmark)
	l := lexer.NewLexer(source, "benchmark.c")
	tokens := l.Tokenize()
	errorHandler := errhand.NewErrorHandler()
	p := parser.NewParser(tokens, errorHandler)
	ast, err := p.Parse()
	if err != nil || ast == nil {
		b.Skip("Skipping semantic benchmark: parse failed")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		errorHandler := errhand.NewErrorHandler()
		b.StartTimer()

		analyzer := semantic.NewSemanticAnalyzer(errorHandler)
		_ = analyzer.Analyze(ast)
	}
}

// benchmarkIR benchmarks the IR generation stage.
func benchmarkIR(b *testing.B, source string) {
	b.Helper()

	// Parse and analyze once (not part of benchmark)
	l := lexer.NewLexer(source, "benchmark.c")
	tokens := l.Tokenize()
	errorHandler := errhand.NewErrorHandler()
	p := parser.NewParser(tokens, errorHandler)
	ast, err := p.Parse()
	if err != nil || ast == nil {
		b.Skip("Skipping IR benchmark: parse failed")
	}

	semErrorHandler := errhand.NewErrorHandler()
	semAnalyzer := semantic.NewSemanticAnalyzer(semErrorHandler)
	_ = semAnalyzer.Analyze(ast)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		errorHandler := errhand.NewErrorHandler()
		b.StartTimer()

		gen := ir.NewIRGenerator(errorHandler)
		irResult, err := gen.Generate(ast)

		// Prevent optimization
		if err != nil && irResult == nil {
			continue
		}
	}
}

// benchmarkCodegen benchmarks the code generation stage.
func benchmarkCodegen(b *testing.B, source string) {
	b.Helper()

	// Full pipeline up to IR (not part of benchmark)
	l := lexer.NewLexer(source, "benchmark.c")
	tokens := l.Tokenize()
	errorHandler := errhand.NewErrorHandler()
	p := parser.NewParser(tokens, errorHandler)
	ast, err := p.Parse()
	if err != nil || ast == nil {
		b.Skip("Skipping codegen benchmark: parse failed")
	}

	semErrorHandler := errhand.NewErrorHandler()
	semAnalyzer := semantic.NewSemanticAnalyzer(semErrorHandler)
	_ = semAnalyzer.Analyze(ast)

	irErrorHandler := errhand.NewErrorHandler()
	irGen := ir.NewIRGenerator(irErrorHandler)
	irResult, err := irGen.Generate(ast)
	if err != nil || irResult == nil {
		b.Skip("Skipping codegen benchmark: IR generation failed")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		errorHandler := errhand.NewErrorHandler()
		b.StartTimer()

		gen := codegen.NewCodeGenerator(errorHandler)
		assembly, err := gen.Generate(irResult)

		// Prevent optimization
		if err != nil && assembly == "" {
			continue
		}
	}
}

// PARSER BENCHMARKS

// BenchmarkParserSimple benchmarks the parser with a simple program.
func BenchmarkParserSimple(b *testing.B) {
	source := loadBenchmarkProgram(b, "simple.c")
	b.ReportAllocs()
	benchmarkParser(b, source)
}

// BenchmarkParserModerate benchmarks the parser with a moderate program.
func BenchmarkParserModerate(b *testing.B) {
	source := loadBenchmarkProgram(b, "moderate.c")
	b.ReportAllocs()
	benchmarkParser(b, source)
}

// BenchmarkParserComplex benchmarks the parser with a complex program.
func BenchmarkParserComplex(b *testing.B) {
	source := loadBenchmarkProgram(b, "complex.c")
	b.ReportAllocs()
	benchmarkParser(b, source)
}

// SEMANTIC ANALYZER BENCHMARKS

// BenchmarkSemanticSimple benchmarks semantic analysis with a simple program.
func BenchmarkSemanticSimple(b *testing.B) {
	source := loadBenchmarkProgram(b, "simple.c")
	b.ReportAllocs()
	benchmarkSemantic(b, source)
}

// BenchmarkSemanticModerate benchmarks semantic analysis with a moderate program.
func BenchmarkSemanticModerate(b *testing.B) {
	source := loadBenchmarkProgram(b, "moderate.c")
	b.ReportAllocs()
	benchmarkSemantic(b, source)
}

// BenchmarkSemanticComplex benchmarks semantic analysis with a complex program.
func BenchmarkSemanticComplex(b *testing.B) {
	source := loadBenchmarkProgram(b, "complex.c")
	b.ReportAllocs()
	benchmarkSemantic(b, source)
}

// IR GENERATOR BENCHMARKS

// BenchmarkIRSimple benchmarks IR generation with a simple program.
func BenchmarkIRSimple(b *testing.B) {
	source := loadBenchmarkProgram(b, "simple.c")
	b.ReportAllocs()
	benchmarkIR(b, source)
}

// BenchmarkIRModerate benchmarks IR generation with a moderate program.
func BenchmarkIRModerate(b *testing.B) {
	source := loadBenchmarkProgram(b, "moderate.c")
	b.ReportAllocs()
	benchmarkIR(b, source)
}

// BenchmarkIRComplex benchmarks IR generation with a complex program.
func BenchmarkIRComplex(b *testing.B) {
	source := loadBenchmarkProgram(b, "complex.c")
	b.ReportAllocs()
	benchmarkIR(b, source)
}

// CODE GENERATOR BENCHMARKS

// BenchmarkCodegenSimple benchmarks code generation with a simple program.
func BenchmarkCodegenSimple(b *testing.B) {
	source := loadBenchmarkProgram(b, "simple.c")
	b.ReportAllocs()
	benchmarkCodegen(b, source)
}

// BenchmarkCodegenModerate benchmarks code generation with a moderate program.
func BenchmarkCodegenModerate(b *testing.B) {
	source := loadBenchmarkProgram(b, "moderate.c")
	b.ReportAllocs()
	benchmarkCodegen(b, source)
}

// BenchmarkCodegenComplex benchmarks code generation with a complex program.
func BenchmarkCodegenComplex(b *testing.B) {
	source := loadBenchmarkProgram(b, "complex.c")
	b.ReportAllocs()
	benchmarkCodegen(b, source)
}

// FULL PIPELINE BENCHMARKS

// compilePipeline executes the full compilation pipeline.
func compilePipeline(source string) (string, error) {
	// Stage 1: Lexical Analysis
	l := lexer.NewLexer(source, "benchmark.c")
	tokens := l.Tokenize()
	if len(tokens) == 0 {
		return "", fmt.Errorf("lexer failed: no tokens produced")
	}

	// Stage 2: Parsing
	errorHandler := errhand.NewErrorHandler()
	p := parser.NewParser(tokens, errorHandler)
	ast, err := p.Parse()
	if err != nil {
		return "", err
	}

	// Stage 3: Semantic Analysis
	semErrorHandler := errhand.NewErrorHandler()
	semAnalyzer := semantic.NewSemanticAnalyzer(semErrorHandler)
	if err := semAnalyzer.Analyze(ast); err != nil {
		return "", err
	}

	// Stage 4: IR Generation
	irErrorHandler := errhand.NewErrorHandler()
	irGen := ir.NewIRGenerator(irErrorHandler)
	irResult, err := irGen.Generate(ast)
	if err != nil {
		return "", err
	}

	// Stage 5: Code Generation
	codegenErrorHandler := errhand.NewErrorHandler()
	gen := codegen.NewCodeGenerator(codegenErrorHandler)
	assembly, err := gen.Generate(irResult)
	if err != nil {
		return "", err
	}

	return assembly, nil
}

// BenchmarkCompileSimple benchmarks full compilation of a simple program.
func BenchmarkCompileSimple(b *testing.B) {
	source := loadBenchmarkProgram(b, "simple.c")
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		b.StartTimer()

		assembly, err := compilePipeline(source)
		if err != nil && assembly == "" {
			// Some errors may occur, continue benchmarking
			continue
		}
	}
}

// BenchmarkCompileModerate benchmarks full compilation of a moderate program.
func BenchmarkCompileModerate(b *testing.B) {
	source := loadBenchmarkProgram(b, "moderate.c")
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		b.StartTimer()

		assembly, err := compilePipeline(source)
		if err != nil && assembly == "" {
			continue
		}
	}
}

// BenchmarkCompileComplex benchmarks full compilation of a complex program.
func BenchmarkCompileComplex(b *testing.B) {
	source := loadBenchmarkProgram(b, "complex.c")
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		b.StartTimer()

		assembly, err := compilePipeline(source)
		if err != nil && assembly == "" {
			continue
		}
	}
}

// LEXER BENCHMARKS

// BenchmarkLexerSimple benchmarks the lexer with a simple program.
func BenchmarkLexerSimple(b *testing.B) {
	source := loadBenchmarkProgram(b, "simple.c")
	b.ReportAllocs()
	benchmarkLexer(b, source)
}

// BenchmarkLexerModerate benchmarks the lexer with a moderate program.
func BenchmarkLexerModerate(b *testing.B) {
	source := loadBenchmarkProgram(b, "moderate.c")
	b.ReportAllocs()
	benchmarkLexer(b, source)
}

// BenchmarkLexerComplex benchmarks the lexer with a complex program.
func BenchmarkLexerComplex(b *testing.B) {
	source := loadBenchmarkProgram(b, "complex.c")
	b.ReportAllocs()
	benchmarkLexer(b, source)
}