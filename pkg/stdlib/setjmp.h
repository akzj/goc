/*
 * setjmp.h - Setjmp/Longjmp Support for GOC
 * 
 * This header defines the setjmp/longjmp interface for error handling.
 * Lua uses this for lua_pcall and lua_error.
 * 
 * Implementation: x86-64 callee-saved registers
 * Saved: rbx, rbp, r12-r15, rsp, rip
 */

#ifndef _SETJMP_H
#define _SETJMP_H

#include <stdint.h>

/*
 * jmp_buf - Storage for saved register state
 * 
 * Layout (x86-64, 64 bytes total):
 *   Offset 0:  rbx  (callee-saved)
 *   Offset 8:  rbp  (frame pointer)
 *   Offset 16: r12  (callee-saved)
 *   Offset 24: r13  (callee-saved)
 *   Offset 32: r14  (callee-saved)
 *   Offset 40: r15  (callee-saved)
 *   Offset 48: rsp  (stack pointer)
 *   Offset 56: rip  (return address)
 */
typedef struct {
    uint64_t rbx;   /* Callee-saved general purpose */
    uint64_t rbp;   /* Frame pointer */
    uint64_t r12;   /* Callee-saved general purpose */
    uint64_t r13;   /* Callee-saved general purpose */
    uint64_t r14;   /* Callee-saved general purpose */
    uint64_t r15;   /* Callee-saved general purpose */
    uint64_t rsp;   /* Stack pointer */
    uint64_t rip;   /* Return address (instruction pointer) */
} jmp_buf[1];

/* Size of jmp_buf in 64-bit values */
#define _JB_SIZE 8

/*
 * setjmp - Save current execution context
 * 
 * @env: Pointer to jmp_buf to save context into
 * @returns: 0 on initial call, non-zero value on longjmp return
 * 
 * Saves the current CPU state (registers, stack pointer, return address)
 * into the provided jmp_buf. Returns 0 on initial call.
 * 
 * When longjmp is called with this env, setjmp "returns" again with
 * the value passed to longjmp.
 */
int setjmp(jmp_buf env);

/*
 * longjmp - Restore previously saved execution context
 * 
 * @env: Pointer to jmp_buf with saved context
 * @val: Value for setjmp to return (must be non-zero)
 * 
 * Restores the CPU state saved by setjmp. Execution resumes at the
 * point where setjmp was called, and setjmp returns val.
 * 
 * If val is 0, setjmp returns 1 instead (longjmp cannot make setjmp return 0).
 * 
 * Undefined behavior if:
 *   - env was not initialized by setjmp
 *   - the function containing setjmp has returned
 *   - env has been invalidated
 */
void longjmp(jmp_buf env, int val);

/*
 * Implementation Notes:
 * 
 * 1. Assembly Implementation (pkg/codegen/runtime/setjmp.S):
 *    - setjmp saves callee-saved registers to jmp_buf
 *    - longjmp restores registers and jumps to saved rip
 * 
 * 2. Calling Convention (System V AMD64 ABI):
 *    - 1st argument (env): rdi
 *    - 2nd argument (val): esi
 *    - Return value: eax
 * 
 * 3. Callee-Saved Registers (must be preserved):
 *    - rbx, rbp, r12, r13, r14, r15
 *    - rsp (implicitly via call/ret)
 * 
 * 4. Caller-Saved Registers (not preserved):
 *    - rax, rcx, rdx, rsi, rdi, r8-r11
 *    - These don't need to be saved
 */

#endif /* _SETJMP_H */