# GOC Compiler Benchmarks

This directory contains benchmark tests for the GOC compiler. The benchmarks measure performance across different compiler stages and with programs of varying complexity.

## Directory Structure

```
tests/benchmark/
├── README.md                      # This file
├── benchmark_test.go              # Basic benchmark test suite (12 benchmarks)
├── benchmark_specialized_test.go  # Specialized benchmarks (17 benchmarks)
└── programs/                      # Benchmark programs
    ├── simple.c                   # Low complexity program
    ├── moderate.c                 # Medium complexity program
    ├── complex.c                  # High complexity program
    ├── memory_intensive.c         # Memory-intensive operations
    ├── deep_recursion.c           # Deep recursion patterns
    ├── codegen_small.c            # Small code generation output
    ├── codegen_medium.c           # Medium code generation output
    ├── codegen_large.c            # Large code generation output
    ├── optimization_unoptimized.c # Unoptimized code patterns
    └── optimization_optimized.c   # Optimized code patterns
```

## Running Benchmarks

### Run All Benchmarks

```bash
go test -bench=. -benchmem ./tests/benchmark/
```

### Run Specific Stage Benchmarks

```bash
# Lexer benchmarks
go test -bench=BenchmarkLexer -benchmem ./tests/benchmark/

# Parser benchmarks
go test -bench=BenchmarkParser -benchmem ./tests/benchmark/

# Semantic analyzer benchmarks
go test -bench=BenchmarkSemantic -benchmem ./tests/benchmark/

# IR generator benchmarks
go test -bench=BenchmarkIR -benchmem ./tests/benchmark/

# Code generator benchmarks
go test -bench=BenchmarkCodegen -benchmem ./tests/benchmark/

# Full pipeline benchmarks
go test -bench=BenchmarkCompile -benchmem ./tests/benchmark/
```

### Run Benchmarks for Specific Complexity

```bash
# Simple program benchmarks
go test -bench='Simple$' -benchmem ./tests/benchmark/

# Moderate program benchmarks
go test -bench='Moderate$' -benchmem ./tests/benchmark/

# Complex program benchmarks
go test -bench='Complex$' -benchmem ./tests/benchmark/
```

### Run with CPU Profile

```bash
go test -bench=. -benchmem -cpuprofile=cpu.prof ./tests/benchmark/
go tool pprof cpu.prof
```

### Run with Memory Profile

```bash
go test -bench=. -benchmem -memprofile=mem.prof ./tests/benchmark/
go tool pprof mem.prof
```

### Run with Specific Iterations

```bash
# Run each benchmark 10 times
go test -bench=. -benchmem -count=10 ./tests/benchmark/
```

## Benchmark Programs

### simple.c (Low Complexity)
- Basic variable declarations
- Simple arithmetic operations
- Minimal control flow
- ~10 lines of code

### moderate.c (Medium Complexity)
- Multiple functions
- For loops
- Conditional statements
- Array operations
- Recursion (factorial)
- ~40 lines of code

### complex.c (High Complexity)
- Struct definitions and usage
- Pointer operations
- Multiple functions with parameters
- Nested conditionals
- Recursion (fibonacci)
- Array iteration
- ~80 lines of code

## Specialized Benchmark Programs (Phase 5.2)

### memory_intensive.c (Memory-Intensive Operations)
- Struct definitions (Point, Rectangle)
- Multiple struct instances
- Arrays (30 elements)
- Array of structs (5 elements)
- Struct passing by value
- Array sum operations
- ~60 lines of code

### deep_recursion.c (Recursion Depth Analysis)
- Factorial recursion
- Fibonacci recursion
- Ackermann function (complex recursion)
- Mutual recursion (is_even/is_odd)
- Tree depth calculation
- Binary tree traversal
- ~130 lines of code

### codegen_small.c (Small Code Generation)
- Single function (add)
- Simple arithmetic operations
- ~15 lines of code
- Baseline for code generation performance

### codegen_medium.c (Medium Code Generation)
- Multiple functions (7 functions)
- Arithmetic operations (add, subtract, multiply, divide, modulo)
- Loop operations
- Recursion (factorial)
- ~65 lines of code

### codegen_large.c (Large Code Generation)
- Multiple functions (7 functions)
- Struct operations (Vector2)
- Array operations
- Recursion
- ~60 lines of code

### optimization_unoptimized.c (Unoptimized Code Patterns)
- Separate square and cube calculations
- Two passes over data
- Function call overhead
- ~45 lines of code

### optimization_optimized.c (Optimized Code Patterns)
- Combined square and cube calculation
- Single pass over data
- Reduced function calls
- ~30 lines of code

## Understanding Benchmark Output

Example output:
```
BenchmarkLexerSimple-8       50000     25000 ns/op    10000 B/op    100 allocs/op
BenchmarkParserSimple-8      20000     65000 ns/op    25000 B/op    250 allocs/op
```

### Metrics Explained

- **ns/op**: Nanoseconds per operation (lower is better)
  - Measures execution time for one benchmark iteration
  
- **B/op**: Bytes allocated per operation (lower is better)
  - Measures memory allocation per iteration
  
- **allocs/op**: Allocations per operation (lower is better)
  - Measures number of memory allocations per iteration
  
- **-8**: GOMAXPROCS value (number of CPU cores used)

## Benchmark Framework

The benchmark framework provides:

1. **Stage-specific benchmarks**: Each compiler stage (lexer, parser, semantic, IR, codegen) can be benchmarked independently

2. **Complexity-based benchmarks**: Programs of varying complexity test scalability

3. **Full pipeline benchmarks**: Measure end-to-end compilation performance

4. **Memory tracking**: All benchmarks use `b.ReportAllocs()` for memory profiling

5. **Reusable helpers**: Internal helper functions for consistent benchmarking

## Specialized Benchmarks (Phase 5.2)

The following specialized benchmark categories were added to provide comprehensive performance profiling:

### Memory-Intensive Benchmarks
- `BenchmarkMemoryIntensiveLexer` - Lexical analysis with struct-heavy code
- `BenchmarkMemoryIntensiveParser` - Parsing with struct-heavy code
- `BenchmarkMemoryIntensiveCompile` - Full compilation with struct-heavy code

### Recursion Depth Benchmarks
- `BenchmarkRecursionLexer` - Lexical analysis with deep recursion code
- `BenchmarkRecursionParser` - Parsing with deep recursion code
- `BenchmarkRecursionSemantic` - Semantic analysis with deep recursion code
- `BenchmarkRecursionIR` - IR generation with deep recursion code
- `BenchmarkRecursionCodegen` - Code generation with deep recursion code
- `BenchmarkRecursionCompile` - Full compilation with deep recursion code

### Code Generation Size Benchmarks
- `BenchmarkCodegenSmall` - Code generation with small program
- `BenchmarkCodegenMedium` - Code generation with medium program
- `BenchmarkCodegenLarge` - Code generation with large program
- `BenchmarkCompileSmall` - Full compilation of small program
- `BenchmarkCompileMedium` - Full compilation of medium program
- `BenchmarkCompileLarge` - Full compilation of large program

### Optimization Pattern Benchmarks
- `BenchmarkOptimizationUnoptimized` - Compilation of unoptimized code patterns
- `BenchmarkOptimizationOptimized` - Compilation of optimized code patterns
- `BenchmarkOptimizationCompare` - Comparison of both patterns

## Adding New Benchmarks

To add a new benchmark:

1. Create a new program in `programs/` directory
2. Add the program to `benchmarkPrograms` slice in `benchmark_test.go`
3. Create benchmark functions following the naming convention:
   - `Benchmark{Stage}{Complexity}` (e.g., `BenchmarkParserComplex`)
4. Use `b.ReportAllocs()` for memory tracking
5. Keep benchmark functions under 100 lines

## Best Practices

1. **Warm-up**: The first iteration may be slower due to initialization
2. **Consistency**: Run benchmarks multiple times (`-count=10`) for reliable results
3. **Isolation**: Close other applications to reduce noise
4. **Comparison**: Use `-benchmem` to track memory allocations
5. **Profiling**: Use CPU/memory profiles to identify bottlenecks

## CI/CD Integration

Add benchmark tracking to CI pipeline:

```yaml
- name: Run Benchmarks
  run: go test -bench=. -benchmem ./tests/benchmark/ | tee benchmark_results.txt
```

Compare results across commits to detect performance regressions.