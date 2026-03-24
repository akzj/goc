# Standard Library Audit for Lua Compilation

**Date**: 2025-03-25  
**Purpose**: Identify GOC stdlib gaps vs Lua requirements

---

## Audit Summary

| Category | Functions Needed | Implemented | Missing | Priority |
|----------|-----------------|-------------|---------|----------|
| stdio.h | 15+ | 0 | 15+ | CRITICAL |
| stdlib.h | 12+ | 0 | 12+ | CRITICAL |
| string.h | 12+ | 0 | 12+ | CRITICAL |
| math.h | 15+ | 0 | 15+ | HIGH |
| setjmp.h | 2 | 0 | 2 | CRITICAL |
| dlfcn.h | 4 | 0 | 4 | HIGH |
| signal.h | 3 | 0 | 3 | LOW |
| **Total** | **63+** | **0** | **63+** | - |

---

## Detailed Audit

### stdio.h (CRITICAL)

| Function | Used By | Status | Priority |
|----------|---------|--------|----------|
| printf | All files | ❌ Missing | CRITICAL |
| fprintf | liolib.c, ldebug.c | ❌ Missing | CRITICAL |
| sprintf | lstrlib.c | ❌ Missing | HIGH |
| snprintf | lauxlib.c | ❌ Missing | HIGH |
| fopen | liolib.c | ❌ Missing | CRITICAL |
| fclose | liolib.c | ❌ Missing | CRITICAL |
| fread | liolib.c, lundump.c | ❌ Missing | CRITICAL |
| fwrite | liolib.c, ldump.c | ❌ Missing | CRITICAL |
| fflush | liolib.c | ❌ Missing | MEDIUM |
| fgets | lua.c | ❌ Missing | HIGH |
| fputs | lua.c | ❌ Missing | MEDIUM |
| fseek | liolib.c | ❌ Missing | MEDIUM |
| ftell | liolib.c | ❌ Missing | MEDIUM |
| rewind | liolib.c | ❌ Missing | LOW |
| FILE*, stdin, stdout, stderr | All I/O | ❌ Missing | CRITICAL |

**Implementation Notes**:
- printf family requires format string parsing
- FILE* requires buffer management
- Can use system calls (read, write, open, close)

### stdlib.h (CRITICAL)

| Function | Used By | Status | Priority |
|----------|---------|--------|----------|
| malloc | All files | ❌ Missing | CRITICAL |
| calloc | lmem.c | ❌ Missing | CRITICAL |
| realloc | lmem.c | ❌ Missing | CRITICAL |
| free | All files | ❌ Missing | CRITICAL |
| atoi | lstrlib.c, lua.c | ❌ Missing | HIGH |
| atol | lstrlib.c | ❌ Missing | MEDIUM |
| atof | lstrlib.c, lmathlib.c | ❌ Missing | HIGH |
| abs | lmathlib.c | ❌ Missing | MEDIUM |
| div | - | ❌ Missing | LOW |
| rand | lmathlib.c | ❌ Missing | MEDIUM |
| srand | lmathlib.c | ❌ Missing | MEDIUM |
| exit | lua.c, lapi.c | ❌ Missing | HIGH |
| atexit | - | ❌ Missing | LOW |
| getenv | loslib.c | ❌ Missing | MEDIUM |
| system | loslib.c | ❌ Missing | LOW |

**Implementation Notes**:
- Memory management is CRITICAL for Lua
- Can start with simple bump allocator
- atoi/atof require string parsing

### string.h (CRITICAL)

| Function | Used By | Status | Priority |
|----------|---------|--------|----------|
| memcpy | All files | ❌ Missing | CRITICAL |
| memmove | lstring.c | ❌ Missing | CRITICAL |
| memset | All files | ❌ Missing | CRITICAL |
| memcmp | ltable.c | ❌ Missing | HIGH |
| strlen | All files | ❌ Missing | CRITICAL |
| strcmp | All files | ❌ Missing | CRITICAL |
| strncmp | lcode.c, lstrlib.c | ❌ Missing | HIGH |
| strcpy | linit.c | ❌ Missing | HIGH |
| strncpy | - | ❌ Missing | MEDIUM |
| strcat | lstrlib.c | ❌ Missing | MEDIUM |
| strchr | lstrlib.c | ❌ Missing | MEDIUM |
| strstr | lstrlib.c | ❌ Missing | MEDIUM |
| strerror | lapi.c | ❌ Missing | MEDIUM |
| memcpy | lvm.c | ❌ Missing | CRITICAL |

**Implementation Notes**:
- Most are simple loops
- memcpy/memmove can use assembly for speed
- Critical for Lua VM performance

### math.h (HIGH)

| Function | Used By | Status | Priority |
|----------|---------|--------|----------|
| sin | lmathlib.c | ❌ Missing | HIGH |
| cos | lmathlib.c | ❌ Missing | HIGH |
| tan | lmathlib.c | ❌ Missing | MEDIUM |
| asin | lmathlib.c | ❌ Missing | MEDIUM |
| acos | lmathlib.c | ❌ Missing | MEDIUM |
| atan | lmathlib.c | ❌ Missing | MEDIUM |
| atan2 | lmathlib.c | ❌ Missing | MEDIUM |
| sqrt | lmathlib.c | ❌ Missing | HIGH |
| pow | lmathlib.c | ❌ Missing | HIGH |
| exp | lmathlib.c | ❌ Missing | MEDIUM |
| log | lmathlib.c | ❌ Missing | HIGH |
| log10 | lmathlib.c | ❌ Missing | MEDIUM |
| floor | lmathlib.c | ❌ Missing | HIGH |
| ceil | lmathlib.c | ❌ Missing | HIGH |
| fabs | lmathlib.c | ❌ Missing | HIGH |
| fmod | lmathlib.c | ❌ Missing | MEDIUM |
| HUGE_VAL | lmathlib.c | ❌ Missing | HIGH |

**Implementation Notes**:
- Requires floating-point support
- Can link with libm if available
- Or implement software versions

### setjmp.h (CRITICAL)

| Function | Used By | Status | Priority |
|----------|---------|--------|----------|
| setjmp | ldo.c, ltests.c | ⚠️ Skeleton | CRITICAL |
| longjmp | ldo.c, ltests.c | ⚠️ Skeleton | CRITICAL |
| jmp_buf | ldo.c, ltests.c | ⚠️ Skeleton | CRITICAL |

**Status**: Headers created, assembly implementation skeleton in place

### dlfcn.h (HIGH)

| Function | Used By | Status | Priority |
|----------|---------|--------|----------|
| dlopen | loadlib.c, lua.c | ⚠️ Skeleton | HIGH |
| dlsym | loadlib.c, lua.c | ⚠️ Skeleton | HIGH |
| dlclose | loadlib.c | ⚠️ Skeleton | MEDIUM |
| dlerror | loadlib.c | ⚠️ Skeleton | MEDIUM |

**Status**: Headers created, stub implementation in place

### signal.h (LOW)

| Function | Used By | Status | Priority |
|----------|---------|--------|----------|
| signal | lua.c | ❌ Missing | LOW |
| raise | - | ❌ Missing | LOW |
| sigaction | - | ❌ Missing | LOW |

**Status**: Not required for basic Lua functionality

---

## Implementation Priority

### Phase 1a (Blockers - Week 1-2)

**setjmp.h**:
- [ ] Complete setjmp.S assembly implementation
- [ ] Test with simple C program
- [ ] Test with Lua error handling

**Memory (stdlib.h)**:
- [ ] malloc (simple bump allocator)
- [ ] free
- [ ] calloc
- [ ] realloc

**String (string.h)**:
- [ ] memcpy
- [ ] memmove
- [ ] memset
- [ ] memcmp
- [ ] strlen
- [ ] strcmp

**I/O (stdio.h)**:
- [ ] printf (basic %d, %s, %f)
- [ ] FILE* structure
- [ ] fopen, fclose
- [ ] fread, fwrite

### Phase 1b (High Priority - Week 2-3)

**Conversion (stdlib.h)**:
- [ ] atoi
- [ ] atof

**String (string.h)**:
- [ ] strcpy
- [ ] strcat
- [ ] strchr
- [ ] strstr

**Math (math.h)**:
- [ ] sqrt
- [ ] pow
- [ ] sin, cos
- [ ] floor, ceil
- [ ] fabs

**dlfcn.h**:
- [ ] Complete ELF64 loader
- [ ] Test with simple .so file

### Phase 2 (Medium Priority - Week 3-4)

**Additional stdio.h**:
- [ ] fprintf, sprintf, snprintf
- [ ] fflush, fgets, fputs
- [ ] fseek, ftell

**Additional stdlib.h**:
- [ ] abs, rand, srand
- [ ] exit, getenv

**Additional string.h**:
- [ ] strncmp, strncpy
- [ ] strerror

### Phase 3 (Low Priority - Week 4+)

**signal.h**:
- [ ] signal
- [ ] raise

**Additional math.h**:
- [ ] tan, asin, acos, atan
- [ ] exp, log, log10
- [ ] fmod

---

## Testing Strategy

### Unit Tests

Create test file for each category:

```c
// test_stdlib.c
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>

int main() {
    // Test malloc/free
    void *p = malloc(100);
    free(p);
    
    // Test string functions
    char buf[100];
    strcpy(buf, "hello");
    if (strlen(buf) != 5) return 1;
    
    // Test printf
    printf("Test: %d\n", 42);
    
    return 0;
}
```

### Integration Tests

Compile Lua files and check for unresolved symbols:

```bash
./goc -c ldo.c -o ldo.o
./goc -c loadlib.c -o loadlib.o
./goc -c liolib.c -o liolib.o
```

### Lua Test Suite

Run Lua testes/ and track pass rate:

```bash
cd testes
../lua api.lua
../lua calls.lua
../lua strings.lua
# ... track results
```

---

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| malloc implementation complex | HIGH | Start with bump allocator, improve later |
| printf format parsing | MEDIUM | Implement basic formats first (%d, %s) |
| Floating-point math | MEDIUM | Link with libm or use software implementation |
| ELF64 loader complexity | HIGH | Start with simple .so, extend gradually |

---

## Conclusion

**Total Functions to Implement**: 63+  
**Critical (Blockers)**: ~20  
**High Priority**: ~20  
**Medium Priority**: ~15  
**Low Priority**: ~8  

**Estimated Effort**: 2-3 weeks for Phase 1 (critical + high priority)

**Next Steps**:
1. Delegate stdlib implementation to Branch
2. Start with memory and string functions
3. Test incrementally with Lua compilation
4. Track progress against this audit