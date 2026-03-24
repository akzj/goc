# Lua Compilation Test Strategy

**Mission ID**: goc-trunk-lua-compilation-arch  
**Date**: 2025-03-25  
**Purpose**: Define incremental validation strategy for Lua compilation

---

## Test Levels

### Level 1: Unit Tests (Component Level)

**Purpose**: Verify individual components work correctly

#### 1.1 setjmp/longjmp Tests

**File**: `test_setjmp.c`

```c
#include <stdio.h>
#include "../pkg/stdlib/setjmp.h"

static jmp_buf env;
static int test_count = 0;

void test_function(void) {
    printf("In test_function, about to longjmp\n");
    longjmp(env, 42);
    printf("ERROR: Should not reach here\n");
}

int main(void) {
    int val;
    
    printf("Test 1: Basic setjmp/longjmp\n");
    val = setjmp(env);
    
    if (val == 0) {
        printf("  Initial setjmp call, val=%d\n", val);
        test_function();
    } else {
        printf("  Returned from longjmp, val=%d\n", val);
        test_count++;
    }
    
    printf("Test 2: Nested setjmp\n");
    jmp_buf env2;
    int val2 = setjmp(env2);
    if (val2 == 0) {
        printf("  Inner setjmp, val=%d\n", val2);
        longjmp(env2, 100);
    } else {
        printf("  Inner returned, val=%d\n", val2);
        test_count++;
    }
    
    printf("\nResults: %d/2 tests passed\n", test_count);
    return (test_count == 2) ? 0 : 1;
}
```

**Build & Run**:
```bash
./goc -I./pkg/stdlib test_setjmp.c -o test_setjmp
./test_setjmp
```

**Expected Output**:
```
Test 1: Basic setjmp/longjmp
  Initial setjmp call, val=0
  In test_function, about to longjmp
  Returned from longjmp, val=42
Test 2: Nested setjmp
  Inner setjmp, val=0
  Inner returned, val=100

Results: 2/2 tests passed
```

**Acceptance Criteria**:
- [ ] Compiles without errors
- [ ] Runs without crashes
- [ ] setjmp returns 0 initially
- [ ] longjmp causes setjmp to return with correct value
- [ ] Nested setjmp/longjmp works

#### 1.2 stdlib Function Tests

**File**: `test_stdlib.c`

```c
#include <stddef.h>

// Include builtin.c directly for testing
#include "../pkg/codegen/runtime/builtin.c"

int main(void) {
    int passed = 0;
    int total = 0;
    
    // Test strlen
    total++;
    if (strlen("hello") == 5) passed++;
    else printf("FAIL: strlen\n");
    
    // Test strcmp
    total++;
    if (strcmp("abc", "abc") == 0 && strcmp("abc", "abd") < 0) passed++;
    else printf("FAIL: strcmp\n");
    
    // Test memcpy
    total++;
    char buf[10];
    memcpy(buf, "test", 5);
    if (strcmp(buf, "test") == 0) passed++;
    else printf("FAIL: memcpy\n");
    
    // Test memset
    total++;
    memset(buf, 0, 10);
    if (buf[0] == 0 && buf[4] == 0) passed++;
    else printf("FAIL: memset\n");
    
    // Test malloc/free (basic)
    total++;
    void *p = malloc(100);
    if (p != NULL) {
        free(p);
        passed++;
    } else {
        printf("FAIL: malloc returned NULL\n");
    }
    
    // Test atoi
    total++;
    if (atoi("42") == 42 && atoi("-10") == -10) passed++;
    else printf("FAIL: atoi\n");
    
    printf("\nstdlib tests: %d/%d passed\n", passed, total);
    return (passed == total) ? 0 : 1;
}
```

**Acceptance Criteria**:
- [ ] All string functions work correctly
- [ ] Memory allocation works
- [ ] Conversion functions work

#### 1.3 dlopen/dlsym Tests

**File**: `test_dlfcn.c`

```c
#include "../pkg/stdlib/dlfcn.h"

int main(void) {
    void *handle;
    void *sym;
    char *error;
    
    printf("Test: dlopen with NULL (main program)\n");
    handle = dlopen(NULL, RTLD_NOW);
    if (handle != NULL) {
        printf("  PASS: dlopen(NULL) returned handle\n");
        dlclose(handle);
    } else {
        error = dlerror();
        printf("  INFO: dlopen(NULL) failed: %s\n", error ? error : "unknown");
    }
    
    printf("Test: dlopen with non-existent file\n");
    handle = dlopen("/nonexistent.so", RTLD_NOW);
    if (handle == NULL) {
        error = dlerror();
        printf("  PASS: dlopen failed as expected: %s\n", error ? error : "unknown");
    } else {
        printf("  FAIL: dlopen should have failed\n");
        dlclose(handle);
    }
    
    return 0;
}
```

**Acceptance Criteria**:
- [ ] dlopen handles NULL filename
- [ ] dlopen reports errors for missing files
- [ ] dlerror returns meaningful messages

---

### Level 2: Integration Tests (Lua Core)

**Purpose**: Verify Lua core files compile correctly

#### 2.1 Individual File Compilation

**Script**: `test_compile_files.sh`

```bash
#!/bin/bash

GOC="/home/ubuntu/workspace/goc/goc"
FLAGS="-I. -O2"

FILES=(
    "lapi.c"
    "lcode.c"
    "ldo.c"
    "lfunc.c"
    "lgc.c"
    "llex.c"
    "lmem.c"
    "lobject.c"
    "lopcodes.c"
    "lparser.c"
    "lstate.c"
    "lstring.c"
    "ltable.c"
    "ltm.c"
    "lundump.c"
    "lvm.c"
    "lzio.c"
)

passed=0
failed=0

for file in "${FILES[@]}"; do
    echo -n "Compiling $file... "
    if $GOC $FLAGS -c "$file" -o "${file%.c}.o" 2>/dev/null; then
        echo "OK"
        ((passed++))
    else
        echo "FAILED"
        ((failed++))
    fi
done

echo ""
echo "Results: $passed/${#FILES[@]} files compiled successfully"
if [ $failed -gt 0 ]; then
    echo "Failed: $failed files"
    exit 1
fi
exit 0
```

**Acceptance Criteria**:
- [ ] All core files compile without errors
- [ ] No unresolved symbol errors (except optional dlopen)
- [ ] Type checking passes

#### 2.2 Symbol Resolution Test

**Purpose**: Verify all required symbols are available

**Script**: `check_symbols.sh`

```bash
#!/bin/bash

# Check for missing symbols in compiled object files
echo "Checking for undefined symbols..."

for obj in *.o; do
    echo "Checking $obj:"
    nm -u "$obj" 2>/dev/null | grep -v "U __" | head -20
done

echo ""
echo "Common missing symbols to watch for:"
echo "  - setjmp, longjmp (setjmp.h)"
echo "  - malloc, free (stdlib.h)"
echo "  - printf, fprintf (stdio.h)"
echo "  - memcpy, strlen (string.h)"
echo "  - sin, cos, sqrt (math.h)"
echo "  - dlopen, dlsym (dlfcn.h)"
```

---

### Level 3: System Tests (Full Lua)

**Purpose**: Verify complete Lua build works

#### 3.1 Build Test

**Script**: `test_build.sh`

```bash
#!/bin/bash

echo "Building Lua with GOC..."
make -f Makefile.goc clean
make -f Makefile.goc all

if [ $? -eq 0 ]; then
    echo "BUILD: SUCCESS"
else
    echo "BUILD: FAILED"
    exit 1
fi

# Check outputs exist
if [ -f "liblua.a" ]; then
    echo "liblua.a: EXISTS"
else
    echo "liblua.a: MISSING"
    exit 1
fi

if [ -f "lua" ]; then
    echo "lua: EXISTS"
else
    echo "lua: MISSING"
    exit 1
fi
```

#### 3.2 Basic Functionality Tests

**Script**: `test_basic.sh`

```bash
#!/bin/bash

LUA="./lua"
passed=0
failed=0

test_lua() {
    local name="$1"
    local code="$2"
    local expected="$3"
    
    echo -n "Test: $name... "
    result=$($LUA -e "$code" 2>&1)
    if [ "$result" = "$expected" ]; then
        echo "PASS"
        ((passed++))
    else
        echo "FAIL (expected '$expected', got '$result')"
        ((failed++))
    fi
}

echo "Running basic Lua tests..."
echo ""

test_lua "Version" "-v" "Lua 5.5"
test_lua "Hello" "print('hello')" "hello"
test_lua "Arithmetic" "print(2+2)" "4"
test_lua "String concat" "print('a'..'b')" "ab"
test_lua "Table length" "t={1,2,3}; print(#t)" "3"
test_lua "Function" "function f(x) return x*2 end; print(f(5))" "10"
test_lua "If statement" "if true then print('yes') else print('no') end" "yes"
test_lua "While loop" "i=1; while i<=3 do print(i); i=i+1 end" "1
2
3"

echo ""
echo "Results: $passed/$((passed+failed)) tests passed"

if [ $failed -gt 0 ]; then
    exit 1
fi
exit 0
```

**Acceptance Criteria**:
- [ ] lua -v shows version
- [ ] print() works
- [ ] Arithmetic works
- [ ] String operations work
- [ ] Tables work
- [ ] Functions work
- [ ] Control structures work

---

### Level 4: Lua Test Suite

**Purpose**: Run official Lua test suite

#### 4.1 Test Runner

**Script**: `run_testes.sh`

```bash
#!/bin/bash

LUA="./lua"
TEST_DIR="./testes"
passed=0
failed=0
total=0

echo "Running Lua test suite..."
echo ""

for test_file in "$TEST_DIR"/*.lua; do
    test_name=$(basename "$test_file")
    ((total++))
    
    echo -n "[$total] $test_name... "
    
    if timeout 30 $LUA "$test_file" > /dev/null 2>&1; then
        echo "PASS"
        ((passed++))
    else
        echo "FAIL"
        ((failed++))
        echo "  Command: $LUA $test_file"
    fi
done

echo ""
echo "================================"
echo "Test Suite Results"
echo "================================"
echo "Total:  $total"
echo "Passed: $passed"
echo "Failed: $failed"
echo "Rate:   $(echo "scale=1; $passed * 100 / $total" | bc)%"
echo ""

if [ $failed -gt 0 ]; then
    echo "Failed tests:"
    # Re-run failed tests with output
fi

exit 0
```

#### 4.2 Priority Test Files

| Test File | Purpose | Priority | Blocks |
|-----------|---------|----------|--------|
| api.lua | C API tests | HIGH | Phase 3 |
| calls.lua | Function calls | HIGH | Phase 3 |
| strings.lua | String operations | HIGH | Phase 3 |
| tables.lua | Table operations | HIGH | Phase 3 |
| gc.lua | Garbage collection | HIGH | Phase 3 |
| errors.lua | Error handling | HIGH | Phase 3 (needs setjmp) |
| libs.lua | Library loading | HIGH | Phase 4 (needs dlopen) |
| math.lua | Math functions | MEDIUM | Phase 3 |
| coroutines.lua | Coroutines | HIGH | Phase 3 |
| locals.lua | Local variables | MEDIUM | Phase 3 |
| constructs.lua | Syntax | MEDIUM | Phase 3 |
| files.lua | File I/O | MEDIUM | Phase 3 |

#### 4.3 Success Criteria

| Phase | Pass Rate | Criteria |
|-------|-----------|----------|
| Initial | 50%+ | Basic functionality works |
| Phase 3 | 80%+ | Most features work |
| Final | 90%+ | Production-ready |

---

## Test Metrics Dashboard

### Compilation Metrics

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Files compiled | 40/40 | 0/40 | ⬜ |
| Unresolved symbols | 0 | - | ⬜ |
| Build time | <5 min | - | ⬜ |

### Functionality Metrics

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| print() works | YES | NO | ⬜ |
| Arithmetic works | YES | NO | ⬜ |
| Tables work | YES | NO | ⬜ |
| Functions work | YES | NO | ⬜ |
| Error handling | YES | NO | ⬜ |
| Module loading | YES | NO | ⬜ |

### Test Suite Metrics

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Tests run | 35+ | 0 | ⬜ |
| Pass rate (initial) | 80%+ | 0% | ⬜ |
| Pass rate (final) | 90%+ | 0% | ⬜ |
| Critical failures | 0 | - | ⬜ |

---

## Debugging Strategy

### Common Issues

| Issue | Symptoms | Debug Approach |
|-------|----------|----------------|
| Missing symbol | Linker error | Check stdlib coverage |
| Segfault | Crash at runtime | Use gdb, check setjmp |
| Wrong output | Incorrect result | Add debug printf |
| Infinite loop | Hang | Check GC, coroutines |
| Memory leak | Growing memory | Check malloc/free |

### Debug Tools

```bash
# Compile with debug symbols
./goc -g source.c -o output

# Run with gdb
gdb ./lua
(gdb) run
(gdb) bt  # Backtrace on crash

# Check memory (if implemented)
./lua -e "collectgarbage('count')"
```

---

## Continuous Integration

### Automated Testing

**Goal**: Run tests on every commit

**Pipeline**:
1. Compile GOC (if changed)
2. Compile Lua with GOC
3. Run unit tests
4. Run basic functionality tests
5. Run test suite (subset)
6. Report results

**Pass Criteria**:
- All unit tests pass
- Basic functionality tests pass
- No new test failures

---

## Test Schedule

| Week | Focus | Tests |
|------|-------|-------|
| 1 | setjmp/longjmp | Unit tests |
| 2 | stdlib functions | Unit + integration |
| 3 | Lua compilation | File compilation |
| 4 | Basic functionality | System tests |
| 5 | Test suite (initial) | Level 4 tests |
| 6 | Bug fixes | Re-run failing tests |
| 7 | Test suite (final) | Full suite |

---

## Conclusion

**Test Strategy Summary**:
1. Start small (unit tests)
2. Build up (integration tests)
3. Full system (Lua build)
4. Comprehensive (test suite)

**Key Metrics**:
- Compilation success rate
- Basic functionality pass rate
- Test suite pass rate

**Success Criteria**:
- 90%+ test suite pass rate
- No critical bugs
- Acceptable performance