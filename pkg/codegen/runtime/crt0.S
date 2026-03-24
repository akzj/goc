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
    
    # Extract argv from stack into %rsi (2nd argument to main)
    # argv is the VALUE at offset 8 from rsp (pointer to argv[0])
    movq 8(%rsp), %rsi
    
    # Call main(argc, argv)
    call main
    
    # main() returns with exit code in %eax
    # Exit using syscall (syscall number 60 = exit)
    movq %rax, %rdi      # Move exit code to %rdi (1st syscall arg)
    movq $60, %rax       # syscall number for exit
    syscall
    
    # Should never reach here - infinite loop as fallback
1:  jmp 1b
    
    .size _start, .-_start