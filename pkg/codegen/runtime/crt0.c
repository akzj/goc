/*
 * crt0.c - C Runtime Startup for GOC
 * 
 * This file works with crt0.S to provide the _start entry point.
 * The assembly stub (crt0.S) extracts argc/argv from the stack
 * and calls main(argc, argv).
 * 
 * Stack Layout (x86-64 System V ABI) at _start entry:
 *   %rsp → 8 bytes: argc (number of arguments)
 *          8 bytes: argv[0] (pointer to program name)
 *          8 bytes: argv[1] (pointer to first argument)
 *          ...
 *          8 bytes: argv[argc] (NULL terminator)
 *          8 bytes: envp[0] (environment pointers)
 */

/* External declaration of main function with arguments */
extern int main(int argc, char **argv);

/* External declaration of exit function */
extern void exit(int code);

/* 
 * Note: _start is defined in crt0.S assembly file.
 * The assembly stub extracts argc/argv from stack and calls main().
 */