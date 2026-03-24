/*
 * operators.c - Operators in C11
 * Tests: arithmetic, comparison, logical operators
 * Expected Output: Results of various operator operations
 */

#include <stdio.h>

int main(void) {
    int a = 10, b = 3;
    
    // Arithmetic operators
    printf("Arithmetic Operators:\n");
    printf("  %d + %d = %d\n", a, b, a + b);
    printf("  %d - %d = %d\n", a, b, a - b);
    printf("  %d * %d = %d\n", a, b, a * b);
    printf("  %d / %d = %d\n", a, b, a / b);
    printf("  %d %% %d = %d\n", a, b, a % b);
    
    // Comparison operators
    printf("\nComparison Operators:\n");
    printf("  %d == %d: %d\n", a, b, a == b);
    printf("  %d != %d: %d\n", a, b, a != b);
    printf("  %d > %d: %d\n", a, b, a > b);
    printf("  %d < %d: %d\n", a, b, a < b);
    
    // Logical operators
    printf("\nLogical Operators:\n");
    printf("  (1 && 0): %d\n", 1 && 0);
    printf("  (1 || 0): %d\n", 1 || 0);
    printf("  !1: %d\n", !1);
    
    return 0;
}