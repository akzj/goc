// Package benchmark provides specialized benchmark tests for the GOC compiler.
// Phase 5.2: Expands the benchmark suite with specialized benchmarks covering:
// - Optimization passes (with/without optimization)
// - Code generation scenarios (different output sizes)
// - Memory-intensive operations (large structs, arrays)
// - Recursion depth analysis
package benchmark

import (
	"testing"
)

// =============================================================================
// SPECIALIZED BENCHMARKS (Phase 5.2)
// =============================================================================
// These benchmarks provide specialized performance profiling for:
// - Optimization passes (with/without optimization patterns)
// - Code generation scenarios (different output sizes)
// - Memory-intensive operations (large structs, arrays)
// - Recursion depth analysis
// =============================================================================

// MEMORY-INTENSIVE BENCHMARKS

// BenchmarkMemoryIntensiveLexer benchmarks lexical analysis with memory-intensive code.
func BenchmarkMemoryIntensiveLexer(b *testing.B) {
	source := loadBenchmarkProgram(b, "memory_intensive.c")
	b.ReportAllocs()
	benchmarkLexer(b, source)
}

// BenchmarkMemoryIntensiveParser benchmarks parsing with memory-intensive code.
func BenchmarkMemoryIntensiveParser(b *testing.B) {
	source := loadBenchmarkProgram(b, "memory_intensive.c")
	b.ReportAllocs()
	benchmarkParser(b, source)
}

// BenchmarkMemoryIntensiveSemantic benchmarks semantic analysis with memory-intensive code.
func BenchmarkMemoryIntensiveSemantic(b *testing.B) {
	source := loadBenchmarkProgram(b, "memory_intensive.c")
	b.ReportAllocs()
	benchmarkSemantic(b, source)
}

// BenchmarkMemoryIntensiveIR benchmarks IR generation with memory-intensive code.
func BenchmarkMemoryIntensiveIR(b *testing.B) {
	source := loadBenchmarkProgram(b, "memory_intensive.c")
	b.ReportAllocs()
	benchmarkIR(b, source)
}

// BenchmarkMemoryIntensiveCodegen benchmarks code generation with memory-intensive code.
func BenchmarkMemoryIntensiveCodegen(b *testing.B) {
	source := loadBenchmarkProgram(b, "memory_intensive.c")
	b.ReportAllocs()
	benchmarkCodegen(b, source)
}

// BenchmarkMemoryIntensiveCompile benchmarks full compilation with memory-intensive code.
func BenchmarkMemoryIntensiveCompile(b *testing.B) {
	source := loadBenchmarkProgram(b, "memory_intensive.c")
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

// RECURSION DEPTH BENCHMARKS

// BenchmarkRecursionLexer benchmarks lexical analysis with deep recursion code.
func BenchmarkRecursionLexer(b *testing.B) {
	source := loadBenchmarkProgram(b, "deep_recursion.c")
	b.ReportAllocs()
	benchmarkLexer(b, source)
}

// BenchmarkRecursionParser benchmarks parsing with deep recursion code.
func BenchmarkRecursionParser(b *testing.B) {
	source := loadBenchmarkProgram(b, "deep_recursion.c")
	b.ReportAllocs()
	benchmarkParser(b, source)
}

// BenchmarkRecursionSemantic benchmarks semantic analysis with deep recursion code.
func BenchmarkRecursionSemantic(b *testing.B) {
	source := loadBenchmarkProgram(b, "deep_recursion.c")
	b.ReportAllocs()
	benchmarkSemantic(b, source)
}

// BenchmarkRecursionIR benchmarks IR generation with deep recursion code.
func BenchmarkRecursionIR(b *testing.B) {
	source := loadBenchmarkProgram(b, "deep_recursion.c")
	b.ReportAllocs()
	benchmarkIR(b, source)
}

// BenchmarkRecursionCodegen benchmarks code generation with deep recursion code.
func BenchmarkRecursionCodegen(b *testing.B) {
	source := loadBenchmarkProgram(b, "deep_recursion.c")
	b.ReportAllocs()
	benchmarkCodegen(b, source)
}

// BenchmarkRecursionCompile benchmarks full compilation with deep recursion code.
func BenchmarkRecursionCompile(b *testing.B) {
	source := loadBenchmarkProgram(b, "deep_recursion.c")
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

// CODE GENERATION SIZE BENCHMARKS

// BenchmarkCodegenSmall benchmarks code generation with small output.
func BenchmarkCodegenSmall(b *testing.B) {
	source := loadBenchmarkProgram(b, "codegen_small.c")
	b.ReportAllocs()
	benchmarkCodegen(b, source)
}

// BenchmarkCodegenMedium benchmarks code generation with medium output.
func BenchmarkCodegenMedium(b *testing.B) {
	source := loadBenchmarkProgram(b, "codegen_medium.c")
	b.ReportAllocs()
	benchmarkCodegen(b, source)
}

// BenchmarkCodegenLarge benchmarks code generation with large output.
func BenchmarkCodegenLarge(b *testing.B) {
	source := loadBenchmarkProgram(b, "codegen_large.c")
	b.ReportAllocs()
	benchmarkCodegen(b, source)
}

// BenchmarkCompileSmall benchmarks full compilation of small program.
func BenchmarkCompileSmall(b *testing.B) {
	source := loadBenchmarkProgram(b, "codegen_small.c")
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

// BenchmarkCompileMedium benchmarks full compilation of medium program.
func BenchmarkCompileMedium(b *testing.B) {
	source := loadBenchmarkProgram(b, "codegen_medium.c")
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

// BenchmarkCompileLarge benchmarks full compilation of large program.
func BenchmarkCompileLarge(b *testing.B) {
	source := loadBenchmarkProgram(b, "codegen_large.c")
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

// OPTIMIZATION PATTERN BENCHMARKS

// BenchmarkOptimizationUnoptimized benchmarks unoptimized code patterns.
func BenchmarkOptimizationUnoptimized(b *testing.B) {
	source := loadBenchmarkProgram(b, "optimization_unoptimized.c")
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

// BenchmarkOptimizationOptimized benchmarks optimized code patterns.
func BenchmarkOptimizationOptimized(b *testing.B) {
	source := loadBenchmarkProgram(b, "optimization_optimized.c")
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

// BenchmarkOptimizationCompare benchmarks both patterns for comparison.
func BenchmarkOptimizationCompare(b *testing.B) {
	unoptSource := loadBenchmarkProgram(b, "optimization_unoptimized.c")
	optSource := loadBenchmarkProgram(b, "optimization_optimized.c")

	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		b.StartTimer()

		// Compile unoptimized version
		assembly1, err1 := compilePipeline(unoptSource)
		if err1 != nil && assembly1 == "" {
			continue
		}

		// Compile optimized version
		assembly2, err2 := compilePipeline(optSource)
		if err2 != nil && assembly2 == "" {
			continue
		}
	}
}