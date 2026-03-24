/*
 * setjmp.S - x86-64 Assembly Implementation of setjmp/longjmp
 * 
 * This file implements context save/restore for error handling.
 * Used by Lua for lua_pcall and lua_error.
 * 
 * Calling Convention (System V AMD64 ABI):
 *   1st argument (env): rdi - pointer to jmp_buf
 *   2nd argument (val): esi - value for longjmp
 *   Return value: eax
 * 
 * Registers Saved (callee-saved + critical):
 *   rbx, rbp, r12, r13, r14, r15, rsp, rip
 * 
 * jmp_buf Layout (8 × 8 bytes = 64 bytes):
 *   [0]   rbx
 *   [8]   rbp
 *   [16]  r12
 *   [24]  r13
 *   [32]  r14
 *   [40]  r15
 *   [48]  rsp
 *   [56]  rip
 */

.section .text
.globl setjmp_asm
.type setjmp_asm, @function

/*
 * int setjmp_asm(jmp_buf env)
 * 
 * Saves current CPU state to env buffer.
 * Returns 0 on initial call.
 * Returns val when returned to via longjmp.
 */
setjmp_asm:
    /* Save callee-saved registers to jmp_buf */
    movq %rbx, 0(%rdi)      /* Save rbx */
    movq %rbp, 8(%rdi)      /* Save rbp */
    movq %r12, 16(%rdi)     /* Save r12 */
    movq %r13, 24(%rdi)     /* Save r13 */
    movq %r14, 32(%rdi)     /* Save r14 */
    movq %r15, 40(%rdi)     /* Save r15 */
    
    /* Save stack pointer */
    movq %rsp, 48(%rdi)     /* Save rsp */
    
    /* Save return address (rip) */
    /* The return address is on top of the stack */
    movq 0(%rsp), %rax      /* Load return address from stack */
    movq %rax, 56(%rdi)     /* Save rip to jmp_buf */
    
    /* Return 0 (initial call) */
    xorl %eax, %eax         /* eax = 0 */
    ret

.size setjmp_asm, .-setjmp_asm


.globl longjmp_asm
.type longjmp_asm, @function

/*
 * void longjmp_asm(jmp_buf env, int val)
 * 
 * Restores CPU state from env buffer.
 * Causes setjmp to return with val.
 */
longjmp_asm:
    /* Restore callee-saved registers from jmp_buf */
    movq 0(%rdi), %rbx      /* Restore rbx */
    movq 8(%rdi), %rbp      /* Restore rbp */
    movq 16(%rdi), %r12     /* Restore r12 */
    movq 24(%rdi), %r13     /* Restore r13 */
    movq 32(%rdi), %r14     /* Restore r14 */
    movq 40(%rdi), %r15     /* Restore r15 */
    
    /* Restore stack pointer */
    movq 48(%rdi), %rsp     /* Restore rsp */
    
    /* Load return value */
    movl %esi, %eax         /* eax = val */
    
    /* Jump to saved instruction pointer */
    /* Push saved rip onto stack and return */
    movq 56(%rdi), %rax     /* Load saved rip */
    pushq %rax              /* Push rip onto stack */
    ret                     /* Pop rip and jump */

.size longjmp_asm, .-longjmp_asm


/*
 * Mark stack as non-executable (fixes linker warning)
 */
.section .note.GNU-stack,"",@progbits
.byte 0


/*
 * Test Program (for reference):
 * 
 * #include <stdio.h>
 * #include <setjmp.h>
 * 
 * static jmp_buf env;
 * 
 * void test_longjmp(void) {
 *     printf("In test_longjmp, about to longjmp\n");
 *     longjmp(env, 42);
 *     printf("This should not print\n");
 * }
 * 
 * int main(void) {
 *     int val = setjmp(env);
 *     
 *     if (val == 0) {
 *         printf("Initial setjmp call, val=%d\n", val);
 *         test_longjmp();
 *     } else {
 *         printf("Returned from longjmp, val=%d\n", val);
 *     }
 *     
 *     return 0;
 * }
 * 
 * Expected output:
 *   Initial setjmp call, val=0
 *   In test_longjmp, about to longjmp
 *   Returned from longjmp, val=42
 */