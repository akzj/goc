# Lua Compilation with GOC - Development Plan

## Executive Summary

This document outlines a development plan to compile Lua 5.5 with the GOC C11 compiler. Lua is an excellent target: **no threading, no C11 atomics, minimal dependencies**, and a comprehensive test suite.

**Current State**: GOC is a functional C11 compiler with full pipeline (lexer → parser → semantic analyzer → IR → codegen → linker → ELF64).

**Target State**: GOC can compile and run Lua 5.5 with all standard libraries and pass the test suite.

---

## Phase 0: Lua Source Analysis (COMPLETED)

### 0.1 Source Structure
- **Location**: `/home/ubuntu/workspace/goc/lua-master/`
- **Core files**: ~40 C files
- **Main targets**: `liblua.a` (static library), `lua` (interpreter)
- **Test suite**: `testes/` directory with 35+ test files

### 0.2 C Features Used by Lua

| Feature | Usage Level | Files | Priority |
|---------|-------------|-------|----------|
| Standard C library | Heavy | All files | CRITICAL |
| Structs/unions | Heavy | lobject.h, lstate.h, ltable.h | CRITICAL |
| Function pointers | Heavy | All files | CRITICAL |
| Variable args | Medium | lauxlib.c, lua.c | HIGH |
| Setjmp/longjmp | Medium | ldo.c, lvm.c | HIGH |
| Dynamic loading (dlopen) | Light | loadlib.c, lua.c | MEDIUM |
| Signal handling | Light | lua.c | LOW |
| Inline assembly | None | - | N/A |
| Threading/Atomics | None | - | N/A |

### 0.3 Key Advantages Over Redis
1. **No threading** - no pthreads, no atomics, no synchronization
2. **Single executable** - no complex build dependencies
3. **Minimal system calls** - mostly stdlib, optional dlopen
4. **Comprehensive tests** - 35+ test files covering all features
5. **Portable C code** - designed to compile everywhere
6. **Small codebase** - ~15K lines vs Redis ~100K+ lines

### 0.4 External Dependencies

| Dependency | Purpose | Complexity | GOC Action |
|------------|---------|------------|------------|
| libdl (dlopen) | Dynamic library loading | Low | Implement or stub |
| readline (optional) | REPL editing | Low | Optional, can skip |
| libm (math) | Math functions | Low | Already supported |
| libc | Standard library | Medium | GOC stdlib coverage |

---

## Phase 1: GOC Core Extensions (Estimated: 2-3 weeks)

### 1.1 Setjmp/Longjmp Support
**Goal**: Implement setjmp/longjmp for Lua error handling

**Tasks**:
- [ ] Add setjmp.h to GOC standard library
- [ ] Implement setjmp/longjmp using x86-64 calling conventions
- [ ] Save/restore: callee-saved registers, stack pointer, instruction pointer
- [ ] Handle return values correctly (setjmp returns 0 or value)
- [ ] Test with Lua error handling (lua_pcall, lua_error)

**Acceptance Criteria**:
- setjmp/longjmp compile and link
- Lua error handling works correctly
- No crashes on lua_pcall/lua_error
- Stack unwinding works correctly

**Files to modify**:
- `pkg/stdlib/setjmp.h` - new standard header
- `pkg/codegen/runtime.c` - setjmp/longjmp implementation
- `pkg/linker/symbols.go` - symbol resolution

### 1.2 Dynamic Loading (dlopen/dlsym/dlclose)
**Goal**: Implement dynamic library loading for Lua modules

**Tasks**:
- [ ] Add dlfcn.h to GOC standard library
- [ ] Implement dlopen, dlsym, dlclose, dlerror
- [ ] Load ELF64 shared libraries (.so files)
- [ ] Symbol resolution in loaded libraries
- [ ] Handle errors correctly

**Acceptance Criteria**:
- loadlib.c compiles and links
- Can load C modules dynamically
- require() works for C libraries
- Error messages are informative

**Files to modify**:
- `pkg/stdlib/dlfcn.h` - new standard header
- `pkg/linker/elf.go` - extend to load .so files
- `pkg/codegen/runtime.c` - dlopen implementation

### 1.3 Enhanced Standard Library Coverage
**Goal**: Ensure all Lua-required stdlib functions are implemented

**Tasks**:
- [ ] Verify stdio.h coverage (printf, scanf, FILE*, etc.)
- [ ] Verify stdlib.h coverage (malloc, free, atoi, etc.)
- [ ] Verify string.h coverage (memcpy, memset, strcmp, etc.)
- [ ] Verify math.h coverage (sin, cos, sqrt, etc.)
- [ ] Verify signal.h coverage (signal, sigaction)
- [ ] Add any missing functions

**Acceptance Criteria**:
- All Lua source files compile without missing symbol errors
- Standard library functions work correctly
- No runtime crashes from stdlib issues

**Files to modify**:
- `pkg/stdlib/*.h` - various standard headers
- `pkg/codegen/builtin.c` - builtin function implementations

---

## Phase 2: Lua Core Compilation (Estimated: 2-3 weeks)

### 2.1 Core Library Files
**Goal**: Compile all Lua core library files

**Priority Order**:
1. **Foundation**: lua.h, luaconf.h, llimits.h, lprefix.h
2. **Core VM**: lobject.c, lvm.c, lopcodes.c, lstate.c
3. **Memory**: lmem.c, lgc.c (garbage collector)
4. **Parser**: llex.c, lparser.c, lcode.c
5. **API**: lapi.c, lauxlib.c
6. **Execution**: ldo.c, lfunc.c, lundump.c, ldump.c
7. **Data Structures**: lstring.c, ltable.c, lzio.c
8. **Libraries**: lbaselib.c, ltablib.c, lstrlib.c, lmathlib.c, etc.

**Tasks**:
- [ ] Fix compilation errors file by file
- [ ] Handle any GOC limitations discovered
- [ ] Verify type checking passes
- [ ] Test incremental compilation

**Acceptance Criteria**:
- All .c files compile to .o files
- No unresolved symbols (except optional dlopen)
- All type checks pass

### 2.2 Static Library Creation
**Goal**: Create liblua.a static library

**Tasks**:
- [ ] Archive all object files into liblua.a
- [ ] Verify symbol table is correct
- [ ] Test linking against liblua.a
- [ ] Verify library can be used by external code

**Acceptance Criteria**:
- liblua.a is valid archive
- Can link external programs against liblua.a
- All symbols are resolved correctly

### 2.3 Lua Interpreter
**Goal**: Compile and link lua interpreter

**Tasks**:
- [ ] Compile lua.c (main interpreter)
- [ ] Link with liblua.a
- [ ] Link with libm, libdl
- [ ] Generate valid ELF64 executable
- [ ] Verify executable runs

**Acceptance Criteria**:
- lua binary is valid ELF64
- Can execute `lua -v` (version)
- Can execute `lua -e "print('hello')"`
- Interactive REPL works

**Files to modify**:
- `pkg/linker/elf.go` - ensure proper executable generation
- Makefile.goc - build configuration

---

## Phase 3: Build System & Testing (Estimated: 2-3 weeks)

### 3.1 Makefile Integration
**Goal**: Create GOC-compatible Makefile

**Tasks**:
- [ ] Create `Makefile.goc` based on Lua makefile
- [ ] Replace GCC with GOC in build commands
- [ ] Handle compilation order
- [ ] Integrate with GOC's linker
- [ ] Support clean, all, test targets

**Acceptance Criteria**:
- `make -f Makefile.goc` builds lua
- Build completes without manual intervention
- Can rebuild individual files

### 3.2 Basic Functionality Testing
**Goal**: Verify Lua interpreter works

**Tasks**:
- [ ] Test basic arithmetic
- [ ] Test string operations
- [ ] Test table operations
- [ ] Test function definitions
- [ ] Test control structures
- [ ] Test coroutines
- [ ] Test modules/require

**Acceptance Criteria**:
- All basic Lua features work
- No crashes on simple scripts
- Output matches expected results

### 3.3 Test Suite Execution
**Goal**: Run Lua test suite

**Tasks**:
- [ ] Run `testes/api.lua`
- [ ] Run `testes/calls.lua`
- [ ] Run `testes/coroutine.lua`
- [ ] Run `testes/gc.lua`
- [ ] Run `testes/strings.lua`
- [ ] Run `testes/tables.lua`
- [ ] Run `testes/math.lua`
- [ ] Run all 35+ test files
- [ ] Track pass/fail rate

**Acceptance Criteria**:
- 80%+ of tests pass initially
- Failures are analyzed and documented
- Critical failures are fixed

### 3.4 Debugging & Fixing
**Goal**: Fix issues discovered during testing

**Tasks**:
- [ ] Create test cases for failures
- [ ] Debug GOC compiler issues
- [ ] Debug code generation issues
- [ ] Debug runtime issues
- [ ] Fix Lua source incompatibilities (if any)
- [ ] Re-run tests after fixes

**Acceptance Criteria**:
- 90%+ of tests pass after fixes
- No crashes or memory corruption
- Known issues are documented

---

## Phase 4: Advanced Features & Optimization (Estimated: 2-3 weeks)

### 4.1 Dynamic Library Loading
**Goal**: Full support for Lua C modules

**Tasks**:
- [ ] Test require() with C modules
- [ ] Create test C module
- [ ] Verify dlopen/dlsym work correctly
- [ ] Test module unloading (if supported)
- [ ] Handle errors gracefully

**Acceptance Criteria**:
- Can load external C modules
- Module functions are callable
- No memory leaks on unload

### 4.2 Performance Optimization
**Goal**: Ensure acceptable performance

**Tasks**:
- [ ] Profile Lua execution
- [ ] Identify hot paths in VM
- [ ] Optimize code generation for VM loop
- [ ] Optimize garbage collector performance
- [ ] Compare with GCC-built Lua
- [ ] Benchmark common operations

**Acceptance Criteria**:
- Performance within 50% of GCC-built Lua (initial target)
- No major regressions
- Hot paths are optimized

### 4.3 Stress Testing
**Goal**: Verify stability under load

**Tasks**:
- [ ] Run long-running scripts
- [ ] Test with large tables
- [ ] Test heavy GC pressure
- [ ] Test deep recursion
- [ ] Test memory allocation patterns

**Acceptance Criteria**:
- No crashes after extended testing
- GC works correctly under pressure
- Memory usage is reasonable

---

## Phase 5: Documentation & Polish (Estimated: 1 week)

### 5.1 Compilation Guide
**Goal**: Document how to compile Lua with GOC

**Tasks**:
- [ ] Write step-by-step compilation guide
- [ ] Document prerequisites
- [ ] Document known issues
- [ ] Document build options

**Acceptance Criteria**:
- Users can follow guide to compile Lua
- Common issues are documented

### 5.2 Test Results Documentation
**Goal**: Document test results and compatibility

**Tasks**:
- [ ] Document test pass rate
- [ ] List passing tests
- [ ] List failing tests with reasons
- [ ] Document known limitations
- [ ] Create compatibility matrix

**Acceptance Criteria**:
- Test results are publicly available
- Users know what works and what doesn't

### 5.3 Compiler Improvements
**Goal**: Improve GOC based on Lua compilation experience

**Tasks**:
- [ ] Fix compiler bugs discovered
- [ ] Improve error messages
- [ ] Add Lua-specific optimizations
- [ ] Document lessons learned

**Acceptance Criteria**:
- Compiler is more robust
- Error messages are clearer
- Future C projects are easier to compile

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| setjmp/longjmp complex | Medium | High | Study x86-64 ABI, test incrementally |
| dlopen implementation difficult | Medium | Medium | Start with stub, implement fully later |
| GC issues in Lua | Low | High | Run GC tests early, fix immediately |
| Performance unacceptable | Medium | Medium | Profile early, optimize hot paths |
| Standard library gaps | High | Medium | Audit stdlib before starting Phase 2 |

---

## Success Criteria

### Minimum Viable Product (MVP)
- [ ] lua interpreter compiles and runs
- [ ] Can execute `print("hello")`
- [ ] Basic arithmetic works
- [ ] Tables and functions work
- [ ] No immediate crashes

### Full Success
- [ ] 90%+ of Lua test suite passes
- [ ] Dynamic module loading works
- [ ] Performance within 50% of GCC
- [ ] All standard libraries work
- [ ] GC works correctly
- [ ] Coroutines work
- [ ] Error handling works (pcall)

---

## Timeline Summary

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Phase 0: Analysis | ✅ Complete | None |
| Phase 1: Core Extensions | 2-3 weeks | None |
| Phase 2: Lua Core | 2-3 weeks | Phase 1 |
| Phase 3: Testing | 2-3 weeks | Phase 2 |
| Phase 4: Advanced | 2-3 weeks | Phase 3 |
| Phase 5: Documentation | 1 week | Phase 4 |
| **Total** | **9-13 weeks** | Sequential |

---

## Immediate Next Steps

1. **Create trunk mission** with this development plan
2. **Prioritize Phase 1** (setjmp/longjmp, stdlib audit)
3. **Audit GOC standard library** against Lua requirements
4. **Create branch missions** for each Phase 1 subtask
5. **Begin implementation** of setjmp/longjmp

---

## Appendix A: Lua Source Files

### Core VM (Critical)
- lobject.c - Lua objects (values, types)
- lvm.c - Virtual machine execution
- lopcodes.c - Opcode definitions
- lstate.c - Lua state (thread state)
- lgc.c - Garbage collector

### Parser & Compiler
- llex.c - Lexer
- lparser.c - Parser
- lcode.c - Code generation

### API & Libraries
- lapi.c - C API implementation
- lauxlib.c - Auxiliary library
- lbaselib.c - Basic library
- ltablib.c - Table library
- lstrlib.c - String library
- lmathlib.c - Math library
- liolib.c - I/O library
- loslib.c - OS library
- lcorolib.c - Coroutine library
- lutf8lib.c - UTF-8 library
- loadlib.c - Dynamic library loading

### Utilities
- lmem.c - Memory management
- lstring.c - String internals
- ltable.c - Table internals
- lfunc.c - Function handling
- lundump.c - Binary loader
- ldump.c - Binary dumper
- lzio.c - Input buffering
- ldebug.c - Debug interface
- ltm.c - Tag methods (metamethods)
- linit.c - Library initialization

### Main Program
- lua.c - Stand-alone interpreter

---

## Appendix B: GOC Standard Library Extensions Needed

### New Headers
- `setjmp.h` - setjmp/longjmp
- `dlfcn.h` - dlopen/dlsym/dlclose
- `signal.h` - signal handling (may exist)

### Runtime Support
- setjmp/longjmp implementation (register save/restore)
- dlopen/dlsym implementation (ELF loading)
- Enhanced stdlib coverage

---

## Appendix C: Lua Test Files

| Test File | Purpose | Priority |
|-----------|---------|----------|
| api.lua | C API tests | HIGH |
| calls.lua | Function calls | HIGH |
| coroutine.lua | Coroutines | HIGH |
| gc.lua | Garbage collection | HIGH |
| strings.lua | String operations | HIGH |
| tables.lua | Table operations | HIGH |
| math.lua | Math functions | MEDIUM |
| locals.lua | Local variables | MEDIUM |
| constructs.lua | Syntax constructs | MEDIUM |
| errors.lua | Error handling | HIGH |
| files.lua | File I/O | MEDIUM |
| libs.lua | Library loading | HIGH |
| ... | (35+ total) | |

---

*Document Version: 1.0*
*Created: Based on Lua 5.5 source analysis*
*Target: GOC Compiler v2.4.0+*