# GOC Implementation Plan - Branch Agent Tasks

**Document Version**: 1.0  
**Created**: 2025-06-17  
**Author**: Zero-FAS (Trunk Node)  
**Status**: Ready for Delegation  

---

## Overview

This document provides detailed implementation tasks for Branch agents to implement the GOC compiler Phases 2-7. All skeleton code has been created by the Trunk node with TODO markers.

---

## Phase 2: Parser Implementation

### Task P2.1: AST Node Definitions (ast.go)
**Files**: `pkg/parser/ast.go`, `pkg/parser/stmt.go`, `pkg/parser/expr.go`
**Complexity**: M
**Dependencies**: None

**Implementation Requirements**:
1. Implement all `Pos()` and `End()` methods to return correct positions
2. Implement all `String()` methods for debugging output
3. Ensure all node types properly implement their interfaces
4. Add helper methods for node traversal

**Acceptance Criteria**:
- [ ] All AST node methods implemented
- [ ] Position tracking works correctly
- [ ] String representations are readable
- [ ] All tests pass

---

### Task P2.2: Type System (type.go)
**Files**: `pkg/parser/type.go`
**Complexity**: M
**Dependencies**: P2.1

**Implementation Requirements**:
1. Implement `Size()` and `Align()` for all type kinds
2. Implement `String()` methods for type printing
3. Add type comparison utilities
4. Implement type equality checking

**Acceptance Criteria**:
- [ ] All type methods implemented
- [ ] Size/alignment calculations correct for x86-64
- [ ] Type comparison works correctly
- [ ] All tests pass

---

### Task P2.3: Parser Core (parser.go)
**Files**: `pkg/parser/parser.go`
**Complexity**: L
**Dependencies**: P2.1, P2.2

**Implementation Requirements**:
1. Implement token stream navigation (current, peek, advance)
2. Implement match/expect helpers
3. Implement error reporting with synchronization
4. Implement main Parse() method

**Acceptance Criteria**:
- [ ] Token navigation works correctly
- [ ] Error recovery works
- [ ] Can parse simple programs
- [ ] All tests pass

---

### Task P2.4: Expression Parsing (expr.go - new file)
**Files**: `pkg/parser/expr_parse.go`
**Complexity**: L
**Dependencies**: P2.3

**Implementation Requirements**:
1. Implement expression parsing with correct precedence
2. Handle all binary and unary operators
3. Implement function call parsing
4. Implement member access and array indexing
5. Implement ternary operator

**Acceptance Criteria**:
- [ ] All expressions parse correctly
- [ ] Operator precedence is correct
- [ ] Error messages are helpful
- [ ] All tests pass

---

### Task P2.5: Statement Parsing (stmt_parse.go - new file)
**Files**: `pkg/parser/stmt_parse.go`
**Complexity**: L
**Dependencies**: P2.4

**Implementation Requirements**:
1. Implement all statement type parsing
2. Handle compound statements with declarations
3. Implement control flow (if, while, for, switch)
4. Implement break/continue/goto

**Acceptance Criteria**:
- [ ] All statements parse correctly
- [ ] Nested blocks work correctly
- [ ] Error recovery works
- [ ] All tests pass

---

### Task P2.6: Declaration Parsing (decl_parse.go - new file)
**Files**: `pkg/parser/decl_parse.go`
**Complexity**: L
**Dependencies**: P2.5

**Implementation Requirements**:
1. Implement function declaration/definition parsing
2. Implement variable declaration parsing
3. Implement struct/union/enum parsing
4. Implement typedef parsing
5. Handle declarators and abstract declarators

**Acceptance Criteria**:
- [ ] All declarations parse correctly
- [ ] Complex declarators work (function pointers, arrays)
- [ ] Error messages are helpful
- [ ] All tests pass

---

### Task P2.7: Parser Tests
**Files**: `pkg/parser/parser_test.go`
**Complexity**: M
**Dependencies**: P2.4, P2.5, P2.6

**Implementation Requirements**:
1. Write unit tests for each parsing function
2. Write integration tests for complete programs
3. Create test fixtures for common C patterns
4. Achieve > 80% code coverage

**Acceptance Criteria**:
- [ ] All parser functions tested
- [ ] Edge cases covered
- [ ] Code coverage > 80%
- [ ] All tests pass

---

## Phase 3: Semantic Analyzer Implementation

### Task P3.1: Symbol Table (symbol.go)
**Files**: `pkg/semantic/symbol.go`
**Complexity**: M
**Dependencies**: Parser complete

**Implementation Requirements**:
1. Implement scope creation and nesting
2. Implement symbol insertion and lookup
3. Implement scope chain traversal
4. Handle shadowing correctly

**Acceptance Criteria**:
- [ ] Symbol table operations work correctly
- [ ] Scope nesting works
- [ ] Shadowing handled correctly
- [ ] All tests pass

---

### Task P3.2: Type Checking (type.go)
**Files**: `pkg/semantic/type.go`
**Complexity**: L
**Dependencies**: P3.1

**Implementation Requirements**:
1. Implement type compatibility checking
2. Implement binary/unary operator type checking
3. Implement function call type checking
4. Implement implicit conversions (usual arithmetic conversions)

**Acceptance Criteria**:
- [ ] Type checking is correct for C11
- [ ] Error messages include type information
- [ ] Implicit conversions work correctly
- [ ] All tests pass

---

### Task P3.3: Semantic Analyzer (analyzer.go)
**Files**: `pkg/semantic/analyzer.go`
**Complexity**: L
**Dependencies**: P3.1, P3.2

**Implementation Requirements**:
1. Implement AST traversal
2. Implement declaration processing
3. Implement statement analysis
4. Implement expression analysis
5. Build symbol table during traversal

**Acceptance Criteria**:
- [ ] Complete semantic analysis works
- [ ] All errors detected and reported
- [ ] Symbol table built correctly
- [ ] All tests pass

---

### Task P3.4: Semantic Tests
**Files**: `pkg/semantic/analyzer_test.go`
**Complexity**: M
**Dependencies**: P3.3

**Implementation Requirements**:
1. Write tests for symbol resolution
2. Write tests for type checking
3. Write tests for error detection
4. Achieve > 80% coverage

**Acceptance Criteria**:
- [ ] All semantic rules tested
- [ ] Error cases covered
- [ ] Code coverage > 80%
- [ ] All tests pass

---

## Phase 4: IR Generator Implementation

### Task P4.1: IR Instructions (instr.go)
**Files**: `pkg/ir/instr.go`
**Complexity**: M
**Dependencies**: None

**Implementation Requirements**:
1. Implement all instruction types
2. Implement instruction creation helpers
3. Implement String() methods for debugging
4. Implement operand utilities

**Acceptance Criteria**:
- [ ] All instruction types implemented
- [ ] Instructions print correctly
- [ ] All tests pass

---

### Task P4.2: IR Generator Core (generator.go)
**Files**: `pkg/ir/generator.go`
**Complexity**: L
**Dependencies**: P4.1, Semantic complete

**Implementation Requirements**:
1. Implement IR generation from AST
2. Implement temporary variable management
3. Implement label generation
4. Implement basic block management
5. Generate control flow graph

**Acceptance Criteria**:
- [ ] IR generation works for all constructs
- [ ] CFG is correct
- [ ] All tests pass

---

### Task P4.3: Expression Lowering (expr_ir.go - new file)
**Files**: `pkg/ir/expr_ir.go`
**Complexity**: L
**Dependencies**: P4.2

**Implementation Requirements**:
1. Lower expressions to three-address code
2. Handle operator semantics
3. Generate temporaries for subexpressions
4. Handle short-circuit evaluation

**Acceptance Criteria**:
- [ ] All expressions lower correctly
- [ ] Temporaries used correctly
- [ ] All tests pass

---

### Task P4.4: IR Tests
**Files**: `pkg/ir/generator_test.go`
**Complexity**: M
**Dependencies**: P4.3

**Implementation Requirements**:
1. Write tests for IR generation
2. Verify CFG structure
3. Test instruction sequences
4. Achieve > 80% coverage

**Acceptance Criteria**:
- [ ] IR generation tested
- [ ] Code coverage > 80%
- [ ] All tests pass

---

## Phase 5: Code Generator Implementation

### Task P5.1: x86-64 Definitions (x86_64.go)
**Files**: `pkg/codegen/x86_64.go`
**Complexity**: M
**Dependencies**: None

**Implementation Requirements**:
1. Implement register string names
2. Implement register sizes
3. Implement instruction emission helpers
4. Define calling convention constants

**Acceptance Criteria**:
- [ ] All registers defined correctly
- [ ] Instruction helpers work
- [ ] All tests pass

---

### Task P5.2: Register Allocator (regalloc.go)
**Files**: `pkg/codegen/regalloc.go`
**Complexity**: L
**Dependencies**: P5.1

**Implementation Requirements**:
1. Implement simple linear scan allocation
2. Handle register spilling
3. Track live ranges
4. Implement reload from spills

**Acceptance Criteria**:
- [ ] Register allocation works
- [ ] Spilling works correctly
- [ ] All tests pass

---

### Task P5.3: Code Generator (generator.go)
**Files**: `pkg/codegen/generator.go`
**Complexity**: L
**Dependencies**: P5.1, P5.2, IR complete

**Implementation Requirements**:
1. Implement function code generation
2. Implement stack frame management
3. Implement instruction selection
4. Implement assembly emission
5. Follow System V AMD64 ABI

**Acceptance Criteria**:
- [ ] Valid x86-64 assembly generated
- [ ] Calling convention followed
- [ ] All tests pass

---

### Task P5.4: Codegen Tests
**Files**: `pkg/codegen/generator_test.go`
**Complexity**: M
**Dependencies**: P5.3

**Implementation Requirements**:
1. Write tests for code generation
2. Verify assembly output
3. Test calling convention
4. Achieve > 80% coverage

**Acceptance Criteria**:
- [ ] Code generation tested
- [ ] Code coverage > 80%
- [ ] All tests pass

---

## Phase 6: Linker Implementation

### Task P6.1: ELF64 Handling (elf.go)
**Files**: `pkg/linker/elf.go`
**Complexity**: M
**Dependencies**: None

**Implementation Requirements**:
1. Implement ELF header creation
2. Implement section header creation
3. Implement program header creation
4. Implement ELF emission

**Acceptance Criteria**:
- [ ] Valid ELF64 headers created
- [ ] All tests pass

---

### Task P6.2: Symbol Resolution (symbol.go)
**Files**: `pkg/linker/symbol.go`
**Complexity**: M
**Dependencies**: None

**Implementation Requirements**:
1. Implement symbol table management
2. Implement symbol resolution
3. Handle undefined symbols
4. Handle duplicate symbols

**Acceptance Criteria**:
- [ ] Symbol resolution works
- [ ] Errors detected correctly
- [ ] All tests pass

---

### Task P6.3: Linker Core (linker.go)
**Files**: `pkg/linker/linker.go`
**Complexity**: L
**Dependencies**: P6.1, P6.2, CodeGen complete

**Implementation Requirements**:
1. Implement object file loading
2. Implement section merging
3. Implement relocation
4. Implement ELF binary emission

**Acceptance Criteria**:
- [ ] Valid ELF executable created
- [ ] All tests pass

---

### Task P6.4: Linker Tests
**Files**: `pkg/linker/linker_test.go`
**Complexity**: M
**Dependencies**: P6.3

**Implementation Requirements**:
1. Write tests for linking
2. Test symbol resolution
3. Test relocation
4. Achieve > 80% coverage

**Acceptance Criteria**:
- [ ] Linking tested
- [ ] Code coverage > 80%
- [ ] All tests pass

---

## Phase 7: CLI & Integration

### Task P7.1: CLI Framework (cli.go)
**Files**: `pkg/cli/cli.go`
**Complexity**: M
**Dependencies**: None

**Implementation Requirements**:
1. Implement command registration
2. Implement flag parsing
3. Implement help/version output
4. Implement command dispatch

**Acceptance Criteria**:
- [ ] CLI framework works
- [ ] All commands registered
- [ ] All tests pass

---

### Task P7.2: Compile Command (compile.go)
**Files**: `pkg/cli/compile.go`
**Complexity**: L
**Dependencies**: P7.1, All phases complete

**Implementation Requirements**:
1. Implement compile command
2. Wire up all compiler phases
3. Implement output file handling
4. Implement error reporting

**Acceptance Criteria**:
- [ ] Can compile C programs
- [ ] Output executable runs correctly
- [ ] All tests pass

---

### Task P7.3: Integration Tests
**Files**: `tests/integration/*.c`
**Complexity**: L
**Dependencies**: P7.2

**Implementation Requirements**:
1. Create test C programs
2. Create expected output files
3. Implement test runner
4. Test end-to-end compilation

**Acceptance Criteria**:
- [ ] All integration tests pass
- [ ] Example programs compile and run
- [ ] Test coverage comprehensive

---

## Error Handling Framework

### Task E1: Error Types (error.go)
**Files**: `internal/errhand/error.go`
**Complexity**: S
**Dependencies**: None

**Implementation Requirements**:
1. Implement ErrorLevel.String()
2. Implement Error.String()
3. Implement Error.Error()

**Acceptance Criteria**:
- [ ] All methods implemented
- [ ] All tests pass

---

### Task E2: Position Tracking (position.go)
**Files**: `internal/errhand/position.go`
**Complexity**: S
**Dependencies**: None

**Implementation Requirements**:
1. Implement Position.String()
2. Implement Position.IsValid()
3. Implement SourceContext.String()

**Acceptance Criteria**:
- [ ] All methods implemented
- [ ] All tests pass

---

### Task E3: Error Handler (handler.go)
**Files**: `internal/errhand/handler.go`
**Complexity**: M
**Dependencies**: E1, E2

**Implementation Requirements**:
1. Implement all error reporting methods
2. Implement error collection
3. Implement source caching
4. Implement error reporting with context

**Acceptance Criteria**:
- [ ] All methods implemented
- [ ] Error context works
- [ ] All tests pass

---

### Task E4: Error Tests
**Files**: `internal/errhand/error_test.go`
**Complexity**: S
**Dependencies**: E3

**Implementation Requirements**:
1. Write tests for error handling
2. Test error formatting
3. Test position tracking
4. Achieve > 80% coverage

**Acceptance Criteria**:
- [ ] Error handling tested
- [ ] Code coverage > 80%
- [ ] All tests pass

---

## Delegation Order

**Wave 1** (Parallel - No dependencies):
- P2.1: AST Node Definitions
- E1: Error Types
- E2: Position Tracking
- P5.1: x86-64 Definitions
- P6.1: ELF64 Handling
- P6.2: Symbol Resolution

**Wave 2** (Parallel - Depends on Wave 1):
- P2.2: Type System
- E3: Error Handler
- P5.2: Register Allocator

**Wave 3** (Parallel - Depends on Wave 2):
- P2.3: Parser Core
- P3.1: Symbol Table
- P4.1: IR Instructions

**Wave 4** (Parallel - Depends on Wave 3):
- P2.4: Expression Parsing
- P3.2: Type Checking
- P4.2: IR Generator Core

**Wave 5** (Parallel - Depends on Wave 4):
- P2.5: Statement Parsing
- P3.3: Semantic Analyzer
- P4.3: Expression Lowering

**Wave 6** (Parallel - Depends on Wave 5):
- P2.6: Declaration Parsing
- P3.4: Semantic Tests
- P4.4: IR Tests

**Wave 7** (Parallel - Depends on Wave 6):
- P2.7: Parser Tests
- P5.3: Code Generator

**Wave 8** (Parallel - Depends on Wave 7):
- P5.4: Codegen Tests
- P6.3: Linker Core

**Wave 9** (Parallel - Depends on Wave 8):
- P6.4: Linker Tests
- P7.1: CLI Framework

**Wave 10** (Parallel - Depends on Wave 9):
- P7.2: Compile Command
- P7.3: Integration Tests
- E4: Error Tests

---

## Testing Strategy

Each task must include:
1. **Unit tests** for individual functions
2. **Integration tests** for complete workflows
3. **Edge case tests** for error conditions
4. **Coverage target**: > 80%

---

## Code Quality Requirements

1. **Follow existing style** in pkg/lexer/
2. **Document all public types and functions**
3. **Keep interfaces clean and testable**
4. **Consistent error handling** across components
5. **C11 standard compliance**

---

**End of Implementation Plan**