# GOC IR Optimization Guide

Documentation for the GOC compiler's IR optimization framework.

## Table of Contents

- [Quick Start](#quick-start)
- [Framework Architecture](#framework-architecture)
- [Optimization Passes](#optimization-passes)
- [Using --optimize Flag](#using-the---optimize-flag)
- [IR Transformation Examples](#ir-transformation-examples)
- [Adding New Passes](#adding-new-passes)
- [Performance Considerations](#performance-considerations)

---

## Quick Start

```bash
# No optimization (default)
goc compile program.c -o program

# Standard optimization (recommended)
goc compile program.c -O2 -o program

# Maximum optimization
goc compile program.c -O3 -o program

# Optimize for size
goc compile program.c -Os -o program
```

---

## Framework Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────┐
│                  OptimizingIRGenerator                   │
│  ┌────────────┐    ┌────────────┐    ┌────────────┐    │
│  │ IRGenerator│ →  │ PassManager│ →  │ OptimizedIR│    │
│  └────────────┘    └─────┬──────┘    └────────────┘    │
│                          │                              │
│                   ┌──────┴──────┐                      │
│                   │ DCE Pass    │                      │
│                   │ ConstFold   │                      │
│                   │ Custom...   │                      │
│                   └─────────────┘                      │
└─────────────────────────────────────────────────────────┘
```

### Core Components

#### OptimizingIRGenerator

```go
type OptimizingIRGenerator struct {
    generator   *IRGenerator
    passManager *PassManager
    config      OptimizationConfig
}
```

**Usage:**
```go
config := ir.OptimizingIRGeneratorConfig{
    ErrorHandler: errorHandler,
    Optimization: ir.OptimizationConfig{
        Enabled: true,
        Passes:  []string{"dead-code-elimination", "constant-folding"},
    },
}
gen, _ := ir.NewOptimizingIRGenerator(config)
ir, _ := gen.Generate(ast)
```

#### PassManager

Orchestrates pass execution:

```go
type PassManager struct {
    passes   []Pass
    enabled  bool
    results  []PassResult
    registry *PassRegistry
}
```

**Key Methods:**
- `Run(ir *IR) (bool, error)` - Execute all passes
- `AddPass(pass Pass)` - Add a pass
- `SetEnabled(bool)` - Enable/disable

#### Pass Interface

```go
type Pass interface {
    Info() PassInfo
    Run(ir *IR) (bool, error)
    Reset()
}
```

#### PassInfo

```go
type PassInfo struct {
    Name         string
    Description  string
    Phase        PassPhase  // Early, Main, or Late
    Dependencies []PassDependency
    Enabled      bool
}
```

### Pass Phases

| Phase | Value | Description |
|-------|-------|-------------|
| Early | 0 | Before most optimizations |
| Main | 1 | Main optimization phase |
| Late | 2 | After most optimizations |

### OptimizationConfig

```go
type OptimizationConfig struct {
    Enabled bool
    Passes  []string
    Verbose bool
}
```

---

## Optimization Passes

### Dead Code Elimination (DCE)

**Name:** `dead-code-elimination` | **Phase:** Main

**Removes:**
- Unreachable basic blocks from CFG
- Dead instructions (results unused, no side effects)
- Unused local variables

**Preserves:**
- Side-effect instructions: `OpStore`, `OpAlloc`, `OpFree`
- Instructions with used results

**Example:**
```c
// Before
int fn() {
    int a = 10;        // Dead
    int b = 20;        // Live
    return b;
}

// After
int fn() {
    int b = 20;
    return b;
}
```

**IR Transformation:**
```
// Before DCE
entry:
    t0 = const 10      // Dead
    t1 = const 20
    ret t1

// After DCE
entry:
    t1 = const 20
    ret t1
```

---

### Constant Folding

**Name:** `constant-folding` | **Phase:** Main

**Description:** Evaluates constant expressions at compile-time.

**Supported Operations:**

| Category | Operations |
|----------|------------|
| Arithmetic | `+`, `-`, `*`, `/`, `%` |
| Comparison | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| Bitwise | `&`, `|`, `^` |

**Example:**
```c
// Before
int fn() {
    int x = 10 * 5 + 3;  // 53
    return x;
}

// After
int fn() {
    int x = 53;
    return x;
}
```

**IR Transformation:**
```
// Before
entry:
    t0 = const 10
    t1 = const 5
    t2 = mul t0, t1      // 50
    t3 = const 3
    t4 = add t2, t3      // 53
    ret t4

// After
entry:
    t0 = const 53
    ret t0
```

**Constant Propagation:**
```c
// Chained constants are fully propagated
int a = 2 + 3;      // 5
int b = a * 4;      // 20
int c = b - 1;      // 19
// Result: c = 19 (fully folded)
```

---

## Using the --optimize Flag

### Syntax

```bash
goc compile <file.c> -O<level> -o <output>
goc compile <file.c> --optimize=<level> -o <output>
```

### Optimization Levels

| Level | Flag | Description | Passes |
|-------|------|-------------|--------|
| O0 | `-O0` | No optimization (default) | None |
| O1 | `-O1` | Basic | DCE |
| O2 | `-O2` | Standard (recommended) | DCE, ConstFold |
| O3 | `-O3` | Maximum | All passes |
| Os | `-Os` | Size optimized | DCE, ConstFold |
| Oz | `-Oz` | Aggressive size | All + size-focused |

### Examples

```bash
# Development (fastest compile)
goc compile prog.c -O0 -o prog

# Production (best performance)
goc compile prog.c -O2 -o prog

# Embedded (smallest binary)
goc compile prog.c -Os -o prog
```

### Validation

Valid levels: `0`, `1`, `2`, `3`, `s`, `z`

```bash
$ goc compile prog.c -O4 -o prog
Error: invalid optimization level '4'. Must be 0, 1, 2, 3, s, or z
```

---

## IR Transformation Examples

### Example 1: Combined DCE + Constant Folding

**Source:**
```c
int compute() {
    int unused = 100;       // Dead
    int x = 5 * 10;         // Fold: 50
    int y = x + 20;         // Fold: 70
    return y;
}
```

**Original IR:**
```
entry:
    t0 = const 100          // unused (dead)
    t1 = const 5
    t2 = const 10
    t3 = mul t1, t2         // 50
    t4 = const 20
    t5 = add t3, t4         // 70
    ret t5
```

**Optimized IR:**
```
entry:
    t0 = const 70
    ret t0
```

**Transformations:**
1. Constant fold: `5 * 10` → `50`
2. Constant fold: `50 + 20` → `70`
3. DCE: Remove dead `unused`
4. Propagate: Replace with `70`

### Example 2: Unreachable Code

**Source:**
```c
int fn() {
    return 42;
    int x = 100;    // Unreachable
    return x;
}
```

**IR Before:**
```
entry:
    ret 42
dead:
    t0 = const 100
    ret t0
```

**IR After:**
```
entry:
    ret 42
```

---

## Adding New Passes

### Step 1: Create Pass Structure

```go
// pkg/ir/mypass.go
package ir

type MyPass struct {
    BasePass
}

func NewMyPass() *MyPass {
    return &MyPass{
        BasePass: NewBasePass(PassInfo{
            Name:        "my-pass",
            Description: "What it does",
            Phase:       PassPhaseMain,
            Dependencies: []PassDependency{
                {Name: "dead-code-elimination", Required: false},
            },
            Enabled: true,
        }),
    }
}
```

### Step 2: Implement Interface

```go
func (mp *MyPass) Run(ir *IR) (bool, error) {
    if ir == nil {
        return false, fmt.Errorf("nil IR")
    }
    modified := false
    for _, fn := range ir.Functions {
        if m, err := mp.processFunction(fn); err != nil {
            return modified, err
        } else if m {
            modified = true
        }
    }
    return modified, nil
}

func (mp *MyPass) Reset() {
    // Reset state
}

func (mp *MyPass) processFunction(fn *Function) (bool, error) {
    // Implement optimization logic
    return modified, nil
}
```

### Step 3: Register Pass

```go
func init() {
    RegisterPass("my-pass", func() Pass {
        return NewMyPass()
    })
}
```

### Step 4: Configure

```go
config := OptimizationConfig{
    Enabled: true,
    Passes:  []string{"dead-code-elimination", "constant-folding", "my-pass"},
}
```

### Pass Checklist

- [ ] Implement `Pass` interface
- [ ] Set appropriate `PassPhase`
- [ ] Declare dependencies
- [ ] Handle nil IR
- [ ] Return accurate `modified` flag
- [ ] Write tests
- [ ] Register pass

---

## Performance Considerations

### Compilation vs. Code Quality

| Level | Compile Time | Code Size | Runtime |
|-------|--------------|-----------|---------|
| O0 | Fastest | Largest | Slowest |
| O1 | Fast | Large | Good |
| O2 | Moderate | Medium | Better |
| O3 | Slow | Medium | Best |
| Os | Moderate | Smallest | Good |
| Oz | Slowest | Very Small | Moderate |

### Pass Overhead

**DCE:**
- Time: O(V + E) per function
- Memory: O(V) for reachability
- Recommendation: Enable for all builds

**Constant Folding:**
- Time: O(I × N) iterations
- Memory: O(T) for constant tracking
- May need multiple passes

### When to Use

| Scenario | Level | Reason |
|----------|-------|--------|
| Development | O0 | Fast iteration, easy debug |
| Testing | O1-O2 | Balance speed/correctness |
| Production | O2-O3 | Best performance |
| Embedded | Os-Oz | Minimize size |

### Benchmarking

```bash
# Compile with different levels
goc compile prog.c -O0 -o prog_o0
goc compile prog.c -O2 -o prog_o2
goc compile prog.c -O3 -o prog_o3
# Compare: ls -lh prog_o*
# Runtime: time ./prog_o0; time ./prog_o2
```

### Debugging
```bash
goc compile prog.c -O2 -v --debug  # Shows pass execution
```

---

## References
- [CLI Guide](cli-guide.md)
- [Architecture](architecture-design.md)
- [Source](../pkg/ir/)

*Version: 1.0.0 | Phase 6.2, Wave 7*