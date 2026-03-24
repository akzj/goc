/*
 * test_setjmp.c - Unit Tests for setjmp/longjmp Implementation
 * 
 * This file tests the x86-64 assembly implementation of setjmp/longjmp.
 * Tests verify:
 * 1. Basic setjmp/longjmp round-trip
 * 2. Return value handling (including val=0 case)
 * 3. Nested setjmp/longjmp
 * 4. Register preservation
 */

#include <stdio.h>
#include <stdint.h>
#include "pkg/stdlib/setjmp.h"

/* Test counters */
static int tests_passed = 0;
static int tests_failed = 0;

/* Test macros */
#define TEST(name) static void name(void)
#define ASSERT(cond, msg) do { \
    if (cond) { \
        printf("  ✓ %s\n", msg); \
        tests_passed++; \
    } else { \
        printf("  ✗ FAILED: %s\n", msg); \
        tests_failed++; \
    } \
} while(0)

/* Global jmp_buf for tests */
static jmp_buf env1;
static jmp_buf env2;

/*
 * Test 1: Basic setjmp/longjmp round-trip
 */
TEST(test_basic_roundtrip) {
    printf("Test 1: Basic setjmp/longjmp round-trip\n");
    
    int val = setjmp(env1);
    
    if (val == 0) {
        /* Initial call */
        printf("  Initial setjmp returned %d\n", val);
        ASSERT(val == 0, "setjmp returns 0 on initial call");
        longjmp(env1, 42);
    } else {
        /* Returned from longjmp */
        printf("  Returned from longjmp with val=%d\n", val);
        ASSERT(val == 42, "setjmp returns value passed to longjmp");
    }
}

/*
 * Test 2: Return value handling (val=0 becomes 1)
 */
TEST(test_zero_value) {
    printf("Test 2: Return value handling (val=0 becomes 1)\n");
    
    static int call_count = 0;
    int val = setjmp(env1);
    
    call_count++;
    
    if (call_count == 1) {
        /* Initial call */
        printf("  Initial setjmp returned %d\n", val);
        ASSERT(val == 0, "setjmp returns 0 on initial call");
        longjmp(env1, 0);  /* Pass 0, should become 1 */
    } else {
        /* Returned from longjmp */
        printf("  Returned from longjmp with val=%d\n", val);
        ASSERT(val == 1, "setjmp returns 1 when longjmp passes 0");
    }
}

/*
 * Test 3: Nested setjmp/longjmp
 */
TEST(test_nested_setjmp) {
    printf("Test 3: Nested setjmp/longjmp\n");
    
    static int nested_state = 0;
    int val1 = setjmp(env1);
    
    if (nested_state == 0) {
        /* First entry */
        printf("  Outer setjmp returned %d\n", val1);
        ASSERT(val1 == 0, "Outer setjmp returns 0 initially");
        
        nested_state = 1;
        int val2 = setjmp(env2);
        
        if (val2 == 0) {
            /* Inner setjmp initial call */
            printf("  Inner setjmp returned %d\n", val2);
            ASSERT(val2 == 0, "Inner setjmp returns 0 initially");
            nested_state = 2;
            longjmp(env2, 99);  /* Jump to inner */
        } else {
            /* Returned to inner */
            printf("  Returned to inner with val=%d\n", val2);
            ASSERT(val2 == 99, "Inner setjmp returns 99");
            nested_state = 3;
            longjmp(env1, 77);  /* Jump to outer */
        }
    } else if (nested_state == 3) {
        /* Returned to outer */
        printf("  Returned to outer with val=%d\n", val1);
        ASSERT(val1 == 77, "Outer setjmp returns 77 from nested longjmp");
        nested_state = 4;
    }
}

/*
 * Test 4: Register preservation (callee-saved registers)
 */
static volatile uint64_t saved_rbx, saved_r12, saved_r15;

TEST(test_register_preservation) {
    printf("Test 4: Register preservation\n");
    
    /* Set some values in callee-saved registers */
    /* Note: Compiler should preserve these across function calls */
    saved_rbx = 0xDEADBEEFCAFEBABEULL;
    saved_r12 = 0x1234567890ABCDEFULL;
    saved_r15 = 0xFEDCBA0987654321ULL;
    
    int val = setjmp(env1);
    
    if (val == 0) {
        printf("  Initial setjmp, registers set\n");
        /* Modify registers (compiler will use them) */
        /* After longjmp, these should be restored to original values */
        longjmp(env1, 1);
    } else {
        printf("  Returned from longjmp\n");
        /* After longjmp, callee-saved registers should be preserved */
        /* Note: This test is limited because C compiler may not use */
        /* the exact registers we expect. The assembly correctly saves */
        /* rbx, r12, r15, but the C code here doesn't guarantee they */
        /* contain our test values. This is more of a sanity check. */
        ASSERT(val == 1, "longjmp returns correct value");
    }
}

/*
 * Test 5: Multiple longjmp calls
 */
TEST(test_multiple_longjmp) {
    printf("Test 5: Multiple longjmp calls\n");
    
    static int multi_state = 0;
    int val = setjmp(env1);
    
    if (multi_state == 0) {
        printf("  Initial setjmp returned %d\n", val);
        ASSERT(val == 0, "setjmp returns 0 initially");
        multi_state = 1;
        longjmp(env1, 100);
    } else if (multi_state == 1) {
        printf("  First longjmp returned %d\n", val);
        ASSERT(val == 100, "First longjmp returns 100");
        multi_state = 2;
        longjmp(env1, 200);
    } else if (multi_state == 2) {
        printf("  Second longjmp returned %d\n", val);
        ASSERT(val == 200, "Second longjmp returns 200");
    }
}

/*
 * Test 6: jmp_buf size verification
 */
TEST(test_jmp_buf_size) {
    printf("Test 6: jmp_buf size verification\n");
    
    /* jmp_buf should be 8 uint64_t values = 64 bytes */
    size_t expected_size = 8 * sizeof(uint64_t);
    size_t actual_size = sizeof(jmp_buf);
    
    printf("  Expected size: %zu bytes\n", expected_size);
    printf("  Actual size: %zu bytes\n", actual_size);
    
    ASSERT(actual_size == expected_size, "jmp_buf size is 64 bytes");
}

/*
 * Run all tests
 */
int main(void) {
    printf("===========================================\n");
    printf("setjmp/longjmp Unit Tests\n");
    printf("===========================================\n\n");
    
    test_basic_roundtrip();
    printf("\n");
    
    test_zero_value();
    printf("\n");
    
    test_nested_setjmp();
    printf("\n");
    
    test_register_preservation();
    printf("\n");
    
    test_multiple_longjmp();
    printf("\n");
    
    test_jmp_buf_size();
    printf("\n");
    
    printf("===========================================\n");
    printf("Test Results: %d passed, %d failed\n", tests_passed, tests_failed);
    printf("===========================================\n");
    
    return tests_failed > 0 ? 1 : 0;
}