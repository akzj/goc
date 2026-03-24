# Lua Phase 2 - Core Compilation Design Document

## Overview

**Mission ID**: goc-trunk-lua-phase2-core  
**Phase**: Phase 2 - Core Compilation & Runtime  
**Status**: Design Complete - Ready for Implementation  
**Date**: 2024-03-24

## Executive Summary

Phase 1 successfully completed Lua compilation infrastructure with 56 tests passing. Phase 2 focuses on fixing GOC runtime limitations to enable full Lua interpreter execution. The primary blockers are:

1. **char** support** - GOC parser doesn't handle pointer-to-pointer types
2. **crt0.c argc/argv** - Runtime doesn't pass command-line arguments to main()
3. **Inline assembly** - May be needed for stack extraction (alternative: C-only approach)

## Architecture

### Current State (Phase 1 ✅)

```
┌─────────────────────────────────────────────────────────────┐
│                    GOC Compiler                              │
├─────────────────────────────────────────────────────────────┤
│  Lexer → Parser → Semantic → IR → CodeGen → ELF64           │
├─────────────────────────────────────────────────────────────┤
│  ✅ 35 Lua source files compile                              │
│  ✅ liblua.a static library creation                         │
│  ✅ setjmp/longjmp (x86-64 assembly)                         │
│  ✅ dlopen/dlsym (ELF64 loader)                              │
│  ✅ 45+ stdlib functions (builtin.c)                         │
│  ✅ 56 compiler tests pass                                   │
├─────────────────────────────────────────────────────────────┤
│  ❌ char** parsing fails                                     │
│  ❌ crt0.c doesn't pass argc/argv                            │
│  ❌ lua executable segfaults at runtime                      │
└─────────────────────────────────────────────────────────────┘
```

### Target State (Phase 2 🎯)

```
┌─────────────────────────────────────────────────────────────┐
│                    GOC Compiler (+ Phase 2)                  │
├─────────────────────────────────────────────────────────────┤
│  ✅ All Phase 1 features                                     │
│  ✅ char** pointer-to-pointer support                        │
│  ✅ crt0.c with argc/argv extraction                         │
│  ✅ lua interpreter runs successfully                        │
│  ✅ Lua test suite executes (>50% pass rate)                 │
└─────────────────────────────────────────────────────────────┘
```

## Problem Analysis

### Issue 1: char** Pointer-to-Pointer Support

**Root Cause**: GOC parser handles `T*` but not `T**` (pointer to pointer).

**Evidence**:
- `lua.c` uses: `int main(int argc, char **argv)`
- Parser fails on `char **argv` declaration
- Type system has `PointerType` but declarator parsing may not handle `**`

**Impact**: 
- Cannot compile any code with pointer-to-pointer parameters
- Blocks Lua interpreter main() function
- Affects ~15-20 Lua API functions using `char**`

**Solution Approach**:
1. Modify parser declarator parsing to handle multiple `*` tokens
2. Build nested `PointerType` structures: `char**` = `PointerType{Elem: PointerType{Elem: CharType}}`
3. Test with minimal reproduction case

### Issue 2: crt0.c argc/argv Handling

**Root Cause**: Current crt0.c calls `main(void)` instead of `main(argc, argv)`.

**Current Code**:
```c
extern int main(void);
void _start(void) {
    int ret = main();  // ❌ No arguments
}
```

**Required Code**:
```c
extern int main(int argc, char **argv);
void _start(void) {
    // Extract argc/argv from stack
    // Call main(argc, argv)
}
```

**Stack Layout (x86-64 System V ABI)**:
```
rsp → [argc]          (8 bytes)
      [argv[0]]       (8 bytes) - pointer to program name
      [argv[1]]       (8 bytes) - pointer to first argument
      ...
      [argv[argc]]    (8 bytes) - NULL terminator
      [envp[0]]       (8 bytes) - environment pointers
      ...
```

**Solution Approaches**:

**Option A: Inline Assembly** (Preferred if supported)
```c
void _start(void) {
    long argc, argv;
    __asm__ volatile ("movq (%%rsp), %0" : "=r"(argc));
    __asm__ volatile ("movq 8(%%rsp), %1" : "=r"(argv));
    main((int)argc, (char**)argv);
}
```

**Option B: C-only with Assembly Stub** (Fallback)
- Create small assembly stub to extract stack values
- Call C function with extracted values

**Option C: Pure C with Compiler Builtin** (If available)
- Use `__builtin_frame_address()` or similar
- Less portable but may work

### Issue 3: Inline Assembly Support

**Status**: Unknown - needs investigation

**Required Instructions**:
- `movq (%%rsp), %reg` - Load from stack
- Basic register operations

**Fallback**: If inline assembly not supported, use Option B above (separate .S file).

## Module Boundaries

### Parser Module (pkg/parser)

**Responsibility**: Parse C source including pointer declarators

**Interface**:
- `ParseDeclarator()`: Handle `*`, `**`, `***` etc.
- `BuildPointerType()`: Create nested PointerType structures
- `ParseType()`: Integrate pointer parsing with type parsing

**Files to Modify**:
- `pkg/parser/expr.go` - Declarator parsing logic
- `pkg/parser/type.go` - PointerType already exists, may need updates
- `pkg/parser/parser.go` - Main parsing integration

**Constraints**:
- Must maintain backward compatibility with `T*`
- Must handle arbitrary pointer depth (`T***...`)
- Must integrate with existing type system

### Runtime Module (pkg/codegen/runtime)

**Responsibility**: Provide C runtime entry point with argc/argv

**Interface**:
- `_start()`: Entry point called by kernel
- `main(argc, argv)`: Call user's main with arguments

**Files to Modify**:
- `pkg/codegen/runtime/crt0.c` - Main runtime startup
- `pkg/codegen/runtime/crt0.S` - Optional assembly stub (if needed)

**Constraints**:
- Must follow x86-64 System V ABI
- Must work with GOC-generated ELF64 executables
- Should minimize assembly usage (prefer C where possible)

### Test Module (test/)

**Responsibility**: Verify Phase 2 fixes

**Test Cases**:
1. `char**` parsing test
2. `crt0.c` argc/argv test
3. Lua interpreter basic execution
4. Lua test suite execution

## Implementation Plan

### Task 1: char** Parser Support (Branch Node)

**Goal**: Enable GOC parser to handle pointer-to-pointer types

**Subtasks**:
1. Create minimal test case: `char **argv`
2. Identify parser location for declarator handling
3. Implement multi-pointer parsing
4. Build nested PointerType structures
5. Test with Lua source files

**Acceptance Criteria**:
- `char **argv` parses without errors
- `char ***envp` parses without errors
- Nested PointerType created correctly
- All 35 Lua files compile

**Estimated Complexity**: M (5-7)

### Task 2: crt0.c argc/argv Support (Branch Node)

**Goal**: Update crt0.c to extract and pass argc/argv to main()

**Subtasks**:
1. Analyze stack layout at _start
2. Implement argc/argv extraction (assembly or C)
3. Update main() signature
4. Test with simple C program
5. Test with lua.c

**Acceptance Criteria**:
- `./lua -v` prints version
- `./lua -e "print('hello')"` works
- argc/argv correctly passed

**Estimated Complexity**: M (6-8)

### Task 3: Lua Runtime Integration (Branch Node)

**Goal**: Integrate fixes and verify Lua interpreter runs

**Subtasks**:
1. Rebuild GOC compiler with fixes
2. Rebuild Lua with new compiler
3. Run basic Lua tests
4. Execute Lua test suite
5. Document results

**Acceptance Criteria**:
- lua executable runs without segfault
- Basic Lua commands work
- Test suite >50% pass rate
- Issues documented

**Estimated Complexity**: M (5-7)

### Task 4: Test Suite & Documentation (Branch Node)

**Goal**: Create comprehensive tests and update documentation

**Subtasks**:
1. Add char** parsing tests to GOC test suite
2. Add crt0.c tests
3. Update Makefile.goc with proper targets
4. Update LUA_COMPILATION_ARCH.md
5. Create Phase 2 completion report

**Acceptance Criteria**:
- All new tests pass
- Documentation updated
- Build process documented
- Phase 2 report complete

**Estimated Complexity**: S (3-4)

## Dependencies

```
Task 1 (char** parser) → Task 3 (Lua integration)
Task 2 (crt0.c)        → Task 3 (Lua integration)
Task 3 (integration)   → Task 4 (tests & docs)
```

**Execution Order**:
1. Task 1 and Task 2 can run in parallel (different files)
2. Task 3 waits for Task 1 and Task 2 completion
3. Task 4 runs after Task 3

## Risk Analysis

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Parser changes break existing code | Medium | High | Extensive regression testing |
| Inline assembly not supported | Medium | Medium | Use assembly stub fallback |
| Stack extraction fails | Low | High | Debug with GDB, verify ABI |
| Lua test suite reveals more issues | High | Medium | Iterative fixing approach |

## Success Metrics

- [ ] 100% Lua source files compile (35/35)
- [ ] lua executable runs without segfault
- [ ] `lua -v` prints version string
- [ ] `lua -e "print('hello')"` outputs "hello"
- [ ] Lua test suite >50% pass rate
- [ ] All new GOC tests pass
- [ ] Documentation complete

## Delegation Plan

| Task | Delegate To | Mission ID | Dependencies |
|------|-------------|------------|--------------|
| Task 1: char** Parser | Branch | goc-branch-phase2-parser | none |
| Task 2: crt0.c Runtime | Branch | goc-branch-phase2-crt0 | none |
| Task 3: Lua Integration | Branch | goc-branch-phase2-integration | parser, crt0 |
| Task 4: Tests & Docs | Branch | goc-branch-phase2-tests | integration |

## Validation Criteria

### Design Validation
- [x] Problem analysis complete
- [x] Solution approaches defined
- [x] Module boundaries clear
- [x] Dependencies mapped
- [x] Risks identified

### Implementation Validation (Post-Delegation)
- [ ] All tasks delegated to Branch nodes
- [ ] No implementation code written by Trunk
- [ ] Delegation tracking in working memory
- [ ] Progress monitoring in place

---

## Appendix A: Minimal Test Cases

### Test 1: char** Parsing
```c
// test_charptr.c
int main(int argc, char **argv) {
    char ***envp;
    return 0;
}
```

### Test 2: crt0.c argc/argv
```c
// test_crt0.c
extern int printf(const char *fmt, ...);
extern void exit(int code);

int main(int argc, char **argv) {
    printf("argc=%d\n", argc);
    printf("argv[0]=%s\n", argv[0]);
    exit(0);
}
```

### Test 3: Lua Basic Execution
```bash
./lua -v
./lua -e "print('hello from lua')"
./lua -e "print(2+2)"
```

## Appendix B: x86-64 Stack Layout

```
At _start entry:
%rsp → 8 bytes: argc (number of arguments)
       8 bytes: argv[0] (pointer to program name)
       8 bytes: argv[1] (pointer to first argument)
       ...
       8 bytes: argv[argc] (NULL pointer)
       8 bytes: envp[0] (pointer to first env var)
       ...
       8 bytes: envp[n] (NULL pointer)
       8 bytes: AT_EXECFN (auxiliary vector)
       ...
```

**Extraction Code**:
```asm
# Get argc
movq (%rsp), %rdi    # First argument to main

# Get argv
leaq 8(%rsp), %rsi   # Second argument to main

# Call main
call main
```

---

*Document Version: 1.0*  
*Created: 2024-03-24*  
*Author: Trunk Node (Architecture & Design)*