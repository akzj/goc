# crt0.S - Assembly entry point for GOC runtime
#
# This file provides the _start entry point that extracts argc/argv
# from the stack and calls the C runtime initialization.
#
# Stack Layout (x86-64 System V ABI) at _start entry:
#   %rsp → 8 bytes: argc (number of arguments)
#          8 bytes: argv[0] (pointer to program name)
#          8 bytes: argv[1] (pointer to first argument)
#          ...
#          8 bytes: argv[argc] (NULL terminator)
#          8 bytes: envp[0] (environment pointers)
#
# Stack Alignment (System V AMD64 ABI):
#   At _start entry from kernel: %rsp % 16 == 8
#   (because the ABI assumes a fake return address was pushed)
#   Before calling main(): %rsp % 16 must be 8
#   (so after call pushes 8 bytes, inside main: %rsp % 16 == 0)
#   Since we start with %rsp % 16 == 8, no adjustment needed!
#
# Calling Convention (System V AMD64 ABI):
#   1st argument: %rdi
#   2nd argument: %rsi
#   3rd argument: %rdx
#   4th argument: %rcx
#   5th argument: %r8
#   6th argument: %r9

    .text
    .globl _start
    .type _start, @function

_start:
    # Extract argc from stack into %rdi (1st argument to main)
    movq (%rsp), %rdi

    # argv (2nd argument) is char **: address of argv[0] on the stack (rsp+8),
    # not the argv[0] pointer itself (that would be char * and crashes main).
    leaq 8(%rsp), %rsi
    
    # Call main(argc, argv)
    # No stack adjustment needed: at _start %rsp % 16 == 8
    # After call pushes 8 bytes, inside main: %rsp % 16 == 0 ✓
    call main
    
    # main() returns with exit code in %eax
    # Exit using syscall (syscall number 60 = exit)
    movq %rax, %rdi      # Move exit code to %rdi (1st syscall arg)
    movq $60, %rax       # syscall number for exit
    syscall
    
    # Should never reach here - infinite loop as fallback
1:  jmp 1b
    
    .size _start, .-_start
