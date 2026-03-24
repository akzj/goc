# C11 Sample Programs for Integration Testing

This directory contains 20+ sample C11 programs for integration testing. Each program tests specific C11 language constructs and is designed to be compiled through the full pipeline to verify end-to-end functionality.

## Program Categories

### 1. Basic Syntax (3 programs)

| Program | Tests | Expected Output |
|---------|-------|-----------------|
| `hello.c` | printf, main function, string literals | "Hello, World!" |
| `variables.c` | int/char/float declarations, initialization | Values of different variable types |
| `operators.c` | arithmetic, comparison, logical operators | Results of various operator operations |

### 2. Control Flow (4 programs)

| Program | Tests | Expected Output |
|---------|-------|-----------------|
| `if_else.c` | if, else if, else statements | Conditional messages based on value |
| `switch.c` | switch, case, default, break | Day name based on number |
| `while_loop.c` | while loops, loop conditions | Counting sequences |
| `for_loop.c` | for loops, increment/decrement | Various loop sequences |

### 3. Functions (3 programs)

| Program | Tests | Expected Output |
|---------|-------|-----------------|
| `functions.c` | function definitions, calls, parameters, return | Results from various function calls |
| `recursion.c` | recursive function calls (factorial, fibonacci) | Factorial and Fibonacci sequences |
| `pointers.c` | pointer declaration, dereference, address-of | Pointer addresses and values |

### 4. Data Structures (3 programs)

| Program | Tests | Expected Output |
|---------|-------|-----------------|
| `arrays.c` | array declaration, indexing, iteration | Array elements and operations |
| `structs.c` | struct definition, initialization, member access | Struct member values |
| `unions.c` | union definition, member access | Union member values (shared memory) |

### 5. Advanced Features (4 programs)

| Program | Tests | Expected Output |
|---------|-------|-----------------|
| `typedef.c` | type aliases, complex type definitions | Demonstration of typedef usage |
| `enums.c` | enum definition, usage in switch/if | Enum values and their usage |
| `sizeof.c` | sizeof operator with various types | Size of different types in bytes |
| `cast.c` | type casting, implicit/explicit conversions | Results of type conversions |

### 6. Edge Cases (3 programs)

| Program | Tests | Expected Output |
|---------|-------|-----------------|
| `empty.c` | minimal valid program (empty main) | No output, returns 0 |
| `comments.c` | single-line, multi-line comments | "Comments test" |
| `preprocessor.c` | #include, #define (basic preprocessor) | Macro expansion results |

## Compilation

All programs are written in valid C11 syntax and can be compiled with:

```bash
gcc -std=c11 -o <output> <program>.c
```

## Coverage

These programs cover all major C11 constructs from Waves 1-4:
- Basic types and operators
- Control flow statements
- Functions and recursion
- Pointers and memory
- Arrays and data structures
- Type system (typedef, enum, sizeof, cast)
- Preprocessor directives

## Constraints

- Each program is <100 lines
- All programs include comments explaining what is being tested
- Valid C11 syntax only
- No undefined behavior
- Each program focuses on specific features