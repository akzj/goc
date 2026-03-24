# Lua Compilation Architecture Design

**Mission ID**: goc-trunk-lua-compilation-arch  
**Version**: 1.0  
**Date**: 2025-03-25  
**Author**: Trunk Node (Architecture & Design)  
**Status**: Design Complete - Ready for Branch Delegation

---

## Executive Summary

This document provides the architecture design for compiling Lua 5.5 with the GOC C11 compiler. The design addresses Phase 1 (Core Extensions) requirements: setjmp/longjmp support, dlopen/dlsym implementation, and standard library coverage.

**Key Decisions**:
1. **setjmp/longjmp**: x86-64 callee-saved registers (rbx, rbp, r12-r15, rsp, rip)
2. **dlopen**: Full ELF64 shared library loader (not stub)
3. **stdlib**: Prioritized coverage based on Lua requirements
4. **Build**: Separate Makefile.goc with GOC-specific toolchain

---

## 1. Lua Source Analysis

### 1.1 Source Structure

**Location**: `/home/ubuntu/workspace/goc/lua-master/`

**File Count**: 40 C files, ~15K lines

**Build Targets**:
- `liblua.a` - Static library (all core files)
- `lua` - Interpreter executable (lua.c + liblua.a)

### 1.2 C Features Used by Lua

| Feature | Usage Level | Files | GOC Status | Priority |
|---------|-------------|-------|------------|----------|
| Standard C library | Heavy | All files | Partial | CRITICAL |
| Structs/unions | Heavy | lobject.h, lstate.h | ✅ Supported | CRITICAL |
| Function pointers | Heavy | All files | ✅ Supported | CRITICAL |
| Variable args | Medium | lauxlib.c, lua.c | ⚠️ Verify | HIGH |
| Setjmp/longjmp | Medium | ldo.c, ltests.c | ❌ Missing | HIGH |
| Dynamic loading | Light | loadlib.c, lua.c | ❌ Missing | MEDIUM |
| Signal handling | Light | lua.c | ⚠️ Verify | LOW |
| Inline assembly | None | - | N/A | N/A |
| Threading/Atomics | None | - | N/A | N/A |

### 1.3 GOC Compatibility Assessment

**✅ Compatible (No Changes Needed)**:
- Basic C syntax (structs, unions, enums)
- Function pointers and callbacks
- Pointer arithmetic
- Type casting
- Static linking

**⚠️ Needs Verification**:
- Variable argument functions (va_list, va_start, va_end)
- Signal handling (signal.h, sigaction)
- Float/double precision in math.h

**❌ Missing (Must Implement)**:
- setjmp.h (setjmp, longjmp, jmp_buf)
- dlfcn.h (dlopen, dlsym, dlclose, dlerror)
- Complete stdio.h (FILE*, fopen, fclose, fprintf, etc.)
- Complete stdlib.h (malloc, free, atoi, atof, etc.)
- Complete string.h (memcpy, memset, strcmp, etc.)
- Complete math.h (sin, cos, sqrt, pow, etc.)

---

## 2. Architecture Decisions

### 2.1 setjmp/longjmp Implementation

**Decision**: Implement full setjmp/longjmp using x86-64 calling conventions

**Rationale**:
- Lua uses setjmp/longjmp for error handling (lua_pcall, lua_error)
- Critical for Lua VM stability
- Cannot stub - must work correctly

**Technical Specification**:

```c
// jmp_buf structure - stores CPU state
typedef struct {
    uint64_t rbx;      // Callee-saved
    uint64_t rbp;      // Frame pointer
    uint64_t r12;      // Callee-saved
    uint64_t r13;      // Callee-saved
    uint64_t r14;      // Callee-saved
    uint64_t r15;      // Callee-saved
    uint64_t rsp;      // Stack pointer
    uint64_t rip;      // Return address
    uint64_t rflags;   // Flags (optional, for safety)
} jmp_buf[1];

#define _JB_SIZE 9  // Number of 64-bit values
```

**Register Save/Restore**:
- **Save (setjmp)**: rbx, rbp, r12-r15, rsp, rip
- **Restore (longjmp)**: Same registers in reverse
- **Return value**: setjmp returns 0 on first call, value on longjmp

**Implementation Location**:
- Header: `pkg/stdlib/setjmp.h`
- Runtime: `pkg/codegen/runtime/setjmp.S` (assembly for context switch)
- C wrapper: `pkg/codegen/runtime/setjmp.c`

**x86-64 Assembly Plan**:
```asm
# setjmp implementation (conceptual)
setjmp:
    mov %rbx, 0(%rdi)    # Save rbx to jmp_buf
    mov %rbp, 8(%rdi)    # Save rbp
    mov %r12, 16(%rdi)   # Save r12
    mov %r13, 24(%rdi)   # Save r13
    mov %r14, 32(%rdi)   # Save r14
    mov %r15, 40(%rdi)   # Save r15
    mov %rsp, 48(%rdi)   # Save rsp
    lea 8(%rsp), %rax    # Get return address
    mov %rax, 56(%rdi)   # Save rip
    xor %eax, %eax       # Return 0
    ret

# longjmp implementation (conceptual)
longjmp:
    mov 0(%rdi), %rbx    # Restore rbx from jmp_buf
    mov 8(%rdi), %rbp    # Restore rbp
    mov 16(%rdi), %r12   # Restore r12
    mov 24(%rdi), %r13   # Restore r13
    mov 32(%rdi), %r14   # Restore r14
    mov 40(%rdi), %r15   # Restore r15
    mov 48(%rdi), %rsp   # Restore rsp
    mov 56(%rdi), %rax   # Load return address
    push %rax            # Push return address
    mov %esi, %eax       # Load return value
    ret                  # Jump to saved rip
```

**Testing Strategy**:
1. Unit test: Basic setjmp/longjmp round-trip
2. Integration test: Lua error handling (lua_pcall)
3. Stress test: Nested setjmp/longjmp

### 2.2 dlopen/dlsym Implementation

**Decision**: Implement full ELF64 shared library loader

**Rationale**:
- Lua's require() depends on dynamic loading
- Stub would break module ecosystem
- GOC linker already has ELF parsing (can extend)

**Technical Specification**:

```c
// dlfcn.h interface
void *dlopen(const char *filename, int flag);
void *dlsym(void *handle, const char *symbol);
int dlclose(void *handle);
char *dlerror(void);

// Flags
#define RTLD_LAZY   1
#define RTLD_NOW    2
#define RTLD_LOCAL  0
#define RTLD_GLOBAL 4
```

**Implementation Architecture**:

```
┌─────────────────────────────────────────────────────────┐
│                    Application (Lua)                     │
│                      require("mod")                      │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                    loadlib.c                             │
│                  dlopen(path, flags)                     │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│              pkg/codegen/runtime/dlfcn.c                │
│  ┌───────────────────────────────────────────────────┐  │
│  │ dlopen()                                          │  │
│  │ 1. Open file                                      │  │
│  │ 2. Parse ELF64 header                             │  │
│  │ 3. Load segments into memory                      │  │
│  │ 4. Process relocations                            │  │
│  │ 5. Return handle                                  │  │
│  └───────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────┐  │
│  │ dlsym()                                           │  │
│  │ 1. Find symbol table                              │  │
│  │ 2. Lookup symbol name                             │  │
│  │ 3. Return address                                 │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│              pkg/linker/elf.go (extend)                 │
│  - Parse ELF64 shared libraries (.so)                   │
│  - Symbol resolution for dynamic libraries              │
│  - Relocation processing                                │
└─────────────────────────────────────────────────────────┘
```

**Implementation Location**:
- Header: `pkg/stdlib/dlfcn.h`
- Runtime: `pkg/codegen/runtime/dlfcn.c`
- Linker: Extend `pkg/linker/elf.go`

**Security Considerations**:
- Validate ELF headers before loading
- Check segment permissions
- Limit memory allocation
- Handle malformed libraries gracefully

**Testing Strategy**:
1. Unit test: Parse valid ELF64 .so file
2. Integration test: Load simple C module
3. Lua test: require() custom module

### 2.3 Standard Library Coverage

**Decision**: Prioritized implementation based on Lua requirements

**Coverage Matrix**:

| Header | Functions Needed | GOC Status | Priority | Phase |
|--------|-----------------|------------|----------|-------|
| stdio.h | printf, fprintf, sprintf, snprintf | ⚠️ Partial | CRITICAL | 1 |
| stdio.h | fopen, fclose, fread, fwrite | ⚠️ Partial | CRITICAL | 1 |
| stdio.h | FILE*, stdin, stdout, stderr | ⚠️ Partial | CRITICAL | 1 |
| stdlib.h | malloc, calloc, realloc, free | ⚠️ Verify | CRITICAL | 1 |
| stdlib.h | atoi, atol, atof | ❌ Missing | HIGH | 1 |
| stdlib.h | abs, div, rand, srand | ❌ Missing | MEDIUM | 1 |
| stdlib.h | exit, atexit | ⚠️ Verify | HIGH | 1 |
| string.h | memcpy, memmove, memset | ⚠️ Partial | CRITICAL | 1 |
| string.h | strlen, strcmp, strncmp | ⚠️ Partial | CRITICAL | 1 |
| string.h | strcpy, strncpy, strcat | ❌ Missing | HIGH | 1 |
| string.h | strchr, strstr, strerror | ❌ Missing | MEDIUM | 1 |
| math.h | sin, cos, tan, asin, acos, atan | ❌ Missing | HIGH | 1 |
| math.h | sqrt, pow, exp, log, log10 | ❌ Missing | HIGH | 1 |
| math.h | floor, ceil, fabs | ❌ Missing | MEDIUM | 1 |
| signal.h | signal, raise | ❌ Missing | LOW | 2 |
| setjmp.h | setjmp, longjmp, jmp_buf | ❌ Missing | CRITICAL | 1 |
| dlfcn.h | dlopen, dlsym, dlclose, dlerror | ❌ Missing | HIGH | 1 |

**Implementation Strategy**:

**Phase 1a** (Critical - Blocker):
- setjmp.h (setjmp, longjmp)
- dlfcn.h (dlopen, dlsym)
- stdio.h (printf family, FILE* ops)
- stdlib.h (malloc family, atoi/atof)
- string.h (memory ops, string ops)

**Phase 1b** (High Priority):
- math.h (basic math functions)
- stdlib.h (exit, atexit)
- string.h (search functions)

**Phase 2** (Medium/Low):
- signal.h (signal handling)
- math.h (advanced functions)
- Additional stdlib functions

**Implementation Location**:
- Headers: `pkg/stdlib/*.h`
- Runtime: `pkg/codegen/runtime/builtin.c`
- Linker symbols: `pkg/linker/symbols.go`

---

## 3. Module Boundaries

### 3.1 Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         GOC Compiler                             │
├─────────────────────────────────────────────────────────────────┤
│  Frontend (Go)                                                  │
│  ├── Lexer (pkg/lexer)                                          │
│  ├── Parser (pkg/parser)                                        │
│  ├── Semantic Analysis (pkg/sema)                               │
│  └── IR Generation (pkg/ir)                                     │
├─────────────────────────────────────────────────────────────────┤
│  Backend (Go)                                                   │
│  ├── Codegen (pkg/codegen)                                      │
│  ├── Register Allocation (pkg/codegen/regalloc.go)              │
│  └── Assembly Output (pkg/codegen/generator.go)                 │
├─────────────────────────────────────────────────────────────────┤
│  Linker (Go)                                                    │
│  ├── ELF64 Generation (pkg/linker/elf.go)                       │
│  ├── Symbol Resolution (pkg/linker/symbol.go)                   │
│  └── Dynamic Loading (pkg/linker/elf.go - extend)               │
├─────────────────────────────────────────────────────────────────┤
│  Runtime Support (C + Assembly) ← NEW FOR LUA                   │
│  ├── setjmp/longjmp (pkg/codegen/runtime/setjmp.S)              │
│  ├── dlopen/dlsym (pkg/codegen/runtime/dlfcn.c)                 │
│  ├── Standard Library (pkg/codegen/runtime/builtin.c)           │
│  └── Startup Code (pkg/codegen/runtime/crt0.c)                  │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 Interface Contracts

#### 3.2.1 setjmp/longjmp Interface

```c
// File: pkg/stdlib/setjmp.h
#ifndef _SETJMP_H
#define _SETJMP_H

#include <stdint.h>

typedef struct {
    uint64_t rbx;
    uint64_t rbp;
    uint64_t r12;
    uint64_t r13;
    uint64_t r14;
    uint64_t r15;
    uint64_t rsp;
    uint64_t rip;
} jmp_buf[1];

int setjmp(jmp_buf env);
void longjmp(jmp_buf env, int val);

#endif
```

**Contract**:
- `setjmp(env)` saves current context, returns 0
- `longjmp(env, val)` restores context, setjmp returns val
- env must remain valid between setjmp and longjmp
- Undefined behavior if env is invalid

#### 3.2.2 dlopen/dlsym Interface

```c
// File: pkg/stdlib/dlfcn.h
#ifndef _DLFCN_H
#define _DLFCN_H

#define RTLD_LAZY   1
#define RTLD_NOW    2
#define RTLD_LOCAL  0
#define RTLD_GLOBAL 4

void *dlopen(const char *filename, int flag);
void *dlsym(void *handle, const char *symbol);
int dlclose(void *handle);
char *dlerror(void);

#endif
```

**Contract**:
- `dlopen(path, flags)` loads .so file, returns handle or NULL
- `dlsym(handle, name)` finds symbol, returns address or NULL
- `dlclose(handle)` unloads library, returns 0 on success
- `dlerror()` returns last error message

#### 3.2.3 Standard Library Interface

```c
// File: pkg/stdlib/stdio.h (partial)
#ifndef _STDIO_H
#define _STDIO_H

typedef struct _FILE FILE;

extern FILE *stdin;
extern FILE *stdout;
extern FILE *stderr;

int printf(const char *format, ...);
int fprintf(FILE *stream, const char *format, ...);
int sprintf(char *str, const char *format, ...);
int snprintf(char *str, int size, const char *format, ...);

FILE *fopen(const char *filename, const char *mode);
int fclose(FILE *stream);
size_t fread(void *ptr, size_t size, size_t nmemb, FILE *stream);
size_t fwrite(const void *ptr, size_t size, size_t nmemb, FILE *stream);

#endif
```

**Contract**:
- Functions match POSIX C standard behavior
- Error handling via return values and errno
- Thread safety: Not required (Lua is single-threaded)

### 3.3 Data Flow

```
Lua Source (ldo.c)
       │
       │ #include <setjmp.h>
       ▼
GOC Compiler
       │
       │ Generates call to setjmp/longjmp
       ▼
pkg/codegen/runtime/setjmp.S
       │
       │ Saves/restores CPU state
       ▼
Lua Error Handling Works

Lua Source (loadlib.c)
       │
       │ #include <dlfcn.h>
       │ require("module")
       ▼
GOC Compiler
       │
       │ Generates call to dlopen/dlsym
       ▼
pkg/codegen/runtime/dlfcn.c
       │
       │ Loads ELF64 .so file
       │ Resolves symbols
       ▼
pkg/linker/elf.go
       │
       │ Parses ELF, processes relocations
       ▼
Module Loaded Successfully
```

---

## 4. Build System Design

### 4.1 Makefile.goc Structure

**Location**: `/home/ubuntu/workspace/goc/lua-master/Makefile.goc`

**Key Components**:
1. **Compiler Configuration**: GOC path, flags
2. **Source Files**: List of .c files
3. **Build Rules**: Compile, archive, link
4. **Targets**: all, clean, test, lua, liblua.a

**Makefile.goc Skeleton**:
```makefile
# Makefile.goc - Build Lua with GOC compiler

# Compiler configuration
GOC = /home/ubuntu/workspace/goc/goc
GOCFLAGS = -O2 -I.
AR = ar
ARFLAGS = rcs

# Source files
CORE_OBJS = lapi.o lcode.o lctype.o ldebug.o ldo.o ldump.o \
            lfunc.o lgc.o llex.o lmem.o lobject.o lopcodes.o \
            lparser.o lstate.o lstring.o ltable.o ltm.o \
            lundump.o lvm.o lzio.o lauxlib.o lbaselib.o \
            lcorolib.o ldblib.o liolib.c lmathlib.o loslib.o \
            lstrlib.o ltablib.o lutf8lib.o loadlib.o linit.o

LUA_OBJS = lua.o

# Standard library objects (GOC runtime)
STD_OBJS = setjmp.o dlfcn.o builtin.o

# Targets
all: lua liblua.a

liblua.a: $(CORE_OBJS)
	$(AR) $(ARFLAGS) $@ $(CORE_OBJS)

lua: $(LUA_OBJS) liblua.a $(STD_OBJS)
	$(GOC) $(GOCFLAGS) -o $@ $(LUA_OBJS) -L. -llua -lm

%.o: %.c
	$(GOC) $(GOCFLAGS) -c $< -o $@

clean:
	rm -f *.o lua liblua.a

.PHONY: all clean
```

### 4.2 Build Order

```
1. Compile GOC runtime support
   ├── setjmp.o (from setjmp.S)
   ├── dlfcn.o (from dlfcn.c)
   └── builtin.o (from builtin.c)

2. Compile Lua core files
   ├── Foundation headers (lua.h, luaconf.h)
   ├── Core VM (lobject.c, lvm.c, lstate.c)
   ├── Parser (llex.c, lparser.c, lcode.c)
   ├── Libraries (lbaselib.c, lstrlib.c, etc.)
   └── Main (lua.c)

3. Create static library
   └── liblua.a (ar rcs)

4. Link final executable
   └── lua (GOC linker)
```

### 4.3 Integration with GOC

**GOC Invocation**:
```bash
# Compile single file
./goc -c src/lapi.c -o lapi.o

# Link executable
./goc lua.o -L. -llua -lm -o lua

# Or use Makefile
make -f Makefile.goc
```

**Linker Extensions Needed**:
- Support for -l flag (library search)
- Support for -L flag (library path)
- Static library (.a) parsing
- Symbol resolution across libraries

---

## 5. Test Strategy

### 5.1 Incremental Validation

**Level 1: Unit Tests** (Each Component)
```bash
# setjmp/longjmp
test_setjmp.c → compile → run → verify context save/restore

# dlopen/dlsym
test_dlfcn.c → compile → run → verify library load

# stdlib functions
test_stdlib.c → compile → run → verify function behavior
```

**Level 2: Integration Tests** (Lua Core)
```bash
# Compile individual Lua files
./goc -c ldo.c -o ldo.o
./goc -c loadlib.c -o loadlib.o

# Verify no unresolved symbols
```

**Level 3: System Tests** (Full Lua)
```bash
# Build liblua.a
make -f Makefile.goc liblua.a

# Build lua interpreter
make -f Makefile.goc lua

# Run basic tests
./lua -v
./lua -e "print('hello')"
./lua -e "print(1+1)"
```

**Level 4: Test Suite** (Lua testes/)
```bash
# Run Lua test suite
cd testes
../lua api.lua
../lua calls.lua
../lua gc.lua
# ... (35+ test files)

# Track pass rate
# Target: 80%+ initially, 90%+ after fixes
```

### 5.2 Test Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Compilation success rate | 100% | All .c files compile |
| Symbol resolution rate | 100% | No unresolved symbols |
| Basic functionality | 100% | print, arithmetic, tables work |
| Test suite pass rate | 80%+ (initial) | testes/ pass count |
| Test suite pass rate | 90%+ (final) | After bug fixes |
| Performance | <50% slowdown | vs GCC-built Lua |

---

## 6. Risk Mitigation

### 6.1 Top 3 Risks

#### Risk 1: setjmp/longjmp Implementation Complexity
**Probability**: Medium  
**Impact**: High (blocks Lua error handling)

**Mitigation**:
1. Study x86-64 ABI calling conventions thoroughly
2. Implement incrementally: save/restore one register at a time
3. Test with simple C program before Lua integration
4. Have fallback: use C longjmp if assembly fails
5. Document register layout clearly

**Contingency**:
- If assembly implementation fails, use compiler intrinsic
- If that fails, modify Lua to use C++ exceptions (last resort)

#### Risk 2: dlopen/dlsym ELF Loading
**Probability**: Medium  
**Impact**: Medium (blocks module loading)

**Mitigation**:
1. Start with simplest .so file (single function)
2. Leverage existing ELF parsing in pkg/linker/elf.go
3. Implement RTLD_NOW first (eager loading)
4. Test symbol resolution separately
5. Have stub implementation as fallback

**Contingency**:
- If full implementation too complex, start with stub
- Stub returns error but allows compilation
- Implement fully in Phase 4

#### Risk 3: Standard Library Gaps
**Probability**: High  
**Impact**: Medium (compilation failures)

**Mitigation**:
1. Audit stdlib before Phase 2 (Lua compilation)
2. Create compatibility matrix (GOC vs Lua needs)
3. Implement missing functions in priority order
4. Use compiler warnings to identify gaps
5. Test each function individually

**Contingency**:
- If function too complex, implement simplified version
- Document limitations clearly
- Prioritize based on Lua test failures

### 6.2 Risk Monitoring

**Weekly Checkpoints**:
- Week 1: setjmp/longjmp prototype working
- Week 2: dlopen basic implementation
- Week 3: stdlib audit complete
- Week 4: First Lua file compiles
- Week 5: liblua.a builds
- Week 6: lua interpreter runs

**Escalation Triggers**:
- 2 weeks behind schedule → Notify Root
- Critical blocker unresolved > 1 week → Escalate
- Test pass rate < 50% after Phase 3 → Review architecture

---

## 7. Phase Dependencies

```
Phase 0: Analysis ✅ COMPLETE
    │
    ▼
Phase 1: Core Extensions (2-3 weeks)
├── 1.1 setjmp/longjmp ← BLOCKER
├── 1.2 dlopen/dlsym
└── 1.3 stdlib coverage
    │
    ▼
Phase 2: Lua Core Compilation (2-3 weeks)
├── 2.1 Compile all .c files ← Needs Phase 1
├── 2.2 Create liblua.a
└── 2.3 Link lua interpreter
    │
    ▼
Phase 3: Build System & Testing (2-3 weeks)
├── 3.1 Makefile.goc
├── 3.2 Basic functionality tests
└── 3.3 Test suite execution ← Needs Phase 2
    │
    ▼
Phase 4: Advanced Features (2-3 weeks)
├── 4.1 Full dlopen support
├── 4.2 Performance optimization
└── 4.3 Stress testing
    │
    ▼
Phase 5: Documentation (1 week)
└── 5.1 Compilation guide
```

**Critical Path**: Phase 1 → Phase 2 → Phase 3  
**Parallel Work**: Phase 1.1, 1.2, 1.3 can be parallelized

---

## 8. Success Criteria

### 8.1 Phase 1 (Core Extensions)

| Criterion | Metric | Status |
|-----------|--------|--------|
| setjmp/longjmp works | Context save/restore verified | ⬜ |
| dlopen/dlsym works | Can load .so file | ⬜ |
| stdlib coverage | All Lua-required functions present | ⬜ |
| Test coverage | Unit tests for all components | ⬜ |

### 8.2 Phase 2 (Lua Core)

| Criterion | Metric | Status |
|-----------|--------|--------|
| All files compile | 40/40 .c files → .o | ⬜ |
| liblua.a builds | Valid static library | ⬜ |
| lua interpreter runs | ./lua -v works | ⬜ |
| Basic Lua works | print, arithmetic, tables | ⬜ |

### 8.3 Phase 3 (Testing)

| Criterion | Metric | Status |
|-----------|--------|--------|
| Test suite runs | All 35+ tests execute | ⬜ |
| Pass rate | ≥80% initially | ⬜ |
| Pass rate | ≥90% after fixes | ⬜ |
| No crashes | Stable execution | ⬜ |

### 8.4 Overall Success

**MVP** (Minimum Viable Product):
- [ ] lua interpreter compiles and runs
- [ ] Can execute `print("hello")`
- [ ] Basic arithmetic works
- [ ] No immediate crashes

**Full Success**:
- [ ] 90%+ of Lua test suite passes
- [ ] Dynamic module loading works
- [ ] Performance within 50% of GCC
- [ ] All standard libraries work
- [ ] GC works correctly
- [ ] Coroutines work
- [ ] Error handling works (pcall)

---

## 9. Delegation Plan

### 9.1 Branch Missions (Phase 1)

| Mission ID | Goal | Complexity | Files |
|------------|------|------------|-------|
| goc-branch-setjmp-impl | Implement setjmp/longjmp | M | pkg/stdlib/setjmp.h, pkg/codegen/runtime/setjmp.S |
| goc-branch-dlfcn-impl | Implement dlopen/dlsym | L | pkg/stdlib/dlfcn.h, pkg/codegen/runtime/dlfcn.c |
| goc-branch-stdlib-audit | Audit and implement stdlib | L | pkg/stdlib/*.h, pkg/codegen/runtime/builtin.c |
| goc-branch-makefile | Create Makefile.goc | S | lua-master/Makefile.goc |
| goc-branch-linker-ext | Extend linker for .a/.so | M | pkg/linker/elf.go, pkg/linker/symbol.go |

### 9.2 Delegation Order

**Wave 1** (Parallel - No Overlap):
- goc-branch-setjmp-impl (pkg/codegen/runtime/setjmp.S)
- goc-branch-stdlib-audit (pkg/stdlib/*.h)
- goc-branch-makefile (lua-master/Makefile.goc)

**Wave 2** (After Wave 1):
- goc-branch-dlfcn-impl (depends on stdlib audit)
- goc-branch-linker-ext (depends on stdlib audit)

### 9.3 Acceptance Criteria per Mission

Each branch mission must:
1. Implement according to interface contracts in this document
2. Pass unit tests
3. Document any deviations from design
4. Submit working code with tests

---

## 10. Working Memory Update

```markdown
## Architecture: Lua Compilation

### Design Document
- Location: `/home/ubuntu/workspace/goc/docs/LUA_COMPILATION_ARCH.md`
- Version: 1.0
- Status: Complete

### Key Decisions
1. setjmp/longjmp: x86-64 callee-saved registers (rbx, rbp, r12-r15, rsp, rip)
2. dlopen: Full ELF64 loader (not stub)
3. stdlib: Prioritized coverage based on Lua requirements
4. Build: Separate Makefile.goc

### Delegation Plan
- Wave 1: setjmp, stdlib-audit, makefile (parallel)
- Wave 2: dlfcn, linker-ext (after Wave 1)

### Next Steps
1. Delegate Phase 1 missions to Branch nodes
2. Monitor progress
3. Validate implementations against design
```

---

## Appendix A: x86-64 Register Layout for jmp_buf

```
Offset  Register  Size  Description
0       rbx       8     Callee-saved general purpose
8       rbp       8     Frame pointer
16      r12       8     Callee-saved general purpose
24      r13       8     Callee-saved general purpose
32      r14       8     Callee-saved general purpose
40      r15       8     Callee-saved general purpose
48      rsp       8     Stack pointer
56      rip       8     Return address (instruction pointer)
------
Total: 64 bytes (8 × 8-byte registers)
```

**Notes**:
- rax, rcx, rdx, rsi, rdi, r8-r11 are caller-saved (not preserved)
- rflags is optional (can be omitted for simplicity)
- Alignment: 8-byte (natural for x86-64)

---

## Appendix B: Lua Files Requiring setjmp/longjmp

```c
// ldo.c - Main error handling
#include <setjmp.h>
typedef struct lua_longjmp {
    struct lua_longjmp *previous;
    jmp_buf b;
} lua_longjmp;

// ltests.c - Test suite
#include <setjmp.h>
struct Aux { jmp_buf jb; const char *paniccode; lua_State *L; };
```

---

## Appendix C: Lua Files Requiring dlopen

```c
// loadlib.c - Dynamic library loading
#include <dlfcn.h>
void *lib = dlopen(path, RTLD_NOW | (seeglb ? RTLD_GLOBAL : RTLD_LOCAL));
lua_CFunction f = cast_Lfunc(dlsym(lib, sym));

// lua.c - Optional readline support
#include <dlfcn.h>
lib = dlopen(rllib, RTLD_NOW | RTLD_LOCAL);
l_readline = cast(l_readlineT, cast_func(dlsym(lib, "readline")));
```

---

*End of Architecture Design Document*