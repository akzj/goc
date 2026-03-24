/*
 * preprocessor.c - Preprocessor Directives
 * Tests: #include, #define (basic preprocessor)
 * Expected Output: Demonstrates preprocessor usage
 */

#include <stdio.h>

// Macro definition
#define PI 3.14159
#define MAX(a, b) ((a) > (b) ? (a) : (b))
#define SQUARE(x) ((x) * (x))

// Conditional compilation
#define DEBUG 1

int main(void) {
    // Using defined macros
    printf("PI = %f\n", PI);
    printf("MAX(5, 10) = %d\n", MAX(5, 10));
    printf("SQUARE(7) = %d\n", SQUARE(7));
    
    // Conditional compilation
#ifdef DEBUG
    printf("\nDebug mode enabled\n");
#endif
    
#ifndef RELEASE
    printf("Not in release mode\n");
#endif
    
    // Undefining a macro
#undef DEBUG
    
#ifdef DEBUG
    printf("This won't print\n");
#else
    printf("DEBUG macro undefined\n");
#endif
    
    // Predefined macros
    printf("\nPredefined macros:\n");
    printf("  __FILE__: %s\n", __FILE__);
    printf("  __LINE__: %d\n", __LINE__);
    printf("  __DATE__: %s\n", __DATE__);
    printf("  __TIME__: %s\n", __TIME__);
    
    return 0;
}