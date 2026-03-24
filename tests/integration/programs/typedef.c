/*
 * typedef.c - Type Aliases
 * Tests: type aliases, complex type definitions
 * Expected Output: Demonstration of typedef usage
 */

#include <stdio.h>

// Simple type alias
typedef unsigned long ulong;

// Typedef for struct
typedef struct {
    int x;
    int y;
} Point;

// Typedef for array
typedef int IntArray[10];

// Typedef for function pointer
typedef int (*MathOp)(int, int);

// Math functions for function pointer
int add(int a, int b) { return a + b; }
int multiply(int a, int b) { return a * b; }

int main(void) {
    // Using simple type alias
    ulong big_num = 1000000UL;
    printf("ulong value: %lu\n", big_num);
    
    // Using struct typedef
    Point p1 = {10, 20};
    Point p2 = {.x = 30, .y = 40};
    printf("\nPoint p1: (%d, %d)\n", p1.x, p1.y);
    printf("Point p2: (%d, %d)\n", p2.x, p2.y);
    
    // Using array typedef
    IntArray arr = {0, 1, 2, 3, 4, 5, 6, 7, 8, 9};
    printf("\nIntArray first element: %d\n", arr[0]);
    
    // Using function pointer typedef
    MathOp op = add;
    printf("\nFunction pointer result: %d + %d = %d\n", 5, 3, op(5, 3));
    
    op = multiply;
    printf("Function pointer result: %d * %d = %d\n", 5, 3, op(5, 3));
    
    return 0;
}