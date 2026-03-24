/*
 * setjmp.c - Setjmp/Longjmp Implementation for GOC
 * 
 * This file provides C wrappers around the assembly implementation.
 * The actual context save/restore is done in setjmp.S for efficiency.
 * 
 * See: pkg/stdlib/setjmp.h for interface documentation
 */

#include "../../stdlib/setjmp.h"

/*
 * Assembly implementation is in pkg/codegen/runtime/setjmp.S
 * 
 * External declarations for assembly functions:
 *   int setjmp_asm(jmp_buf env) - Save context
 *   void longjmp_asm(jmp_buf env, int val) - Restore context
 */

extern int setjmp_asm(jmp_buf env);
extern void longjmp_asm(jmp_buf env, int val);

/*
 * setjmp - Save current execution context
 * 
 * Wrapper around assembly implementation.
 */
int setjmp(jmp_buf env) {
    return setjmp_asm(env);
}

/*
 * longjmp - Restore previously saved execution context
 * 
 * Wrapper around assembly implementation.
 */
void longjmp(jmp_buf env, int val) {
    /* longjmp cannot make setjmp return 0 */
    if (val == 0) {
        val = 1;
    }
    longjmp_asm(env, val);
}

/*
 * Implementation Notes:
 * 
 * The assembly implementation (setjmp.S) handles:
 * 1. Saving callee-saved registers to jmp_buf
 * 2. Saving stack pointer and return address
 * 3. Restoring registers on longjmp
 * 4. Jumping to saved instruction pointer
 * 
 * Why Assembly?
 * - Need precise control over register save/restore
 * - Must save return address (rip) explicitly
 * - Compiler-generated code would not work correctly
 * 
 * Alternative (if assembly fails):
 * - Use compiler intrinsic: __builtin_setjmp, __builtin_longjmp
 * - Less portable but may work with some compilers
 */