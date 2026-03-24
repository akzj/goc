/*
 * sizeof.c - Sizeof Operator
 * Tests: sizeof operator with various types
 * Expected Output: Size of different types in bytes
 */

#include <stdio.h>

struct MyStruct {
    int a;
    char b;
    double c;
};

union MyUnion {
    int i;
    double d;
    char c;
};

int main(void) {
    // Sizeof basic types
    printf("Size of basic types:\n");
    printf("  char: %zu bytes\n", sizeof(char));
    printf("  short: %zu bytes\n", sizeof(short));
    printf("  int: %zu bytes\n", sizeof(int));
    printf("  long: %zu bytes\n", sizeof(long));
    printf("  float: %zu bytes\n", sizeof(float));
    printf("  double: %zu bytes\n", sizeof(double));
    printf("  long double: %zu bytes\n", sizeof(long double));
    
    // Sizeof pointers
    printf("\nSize of pointers:\n");
    printf("  int*: %zu bytes\n", sizeof(int*));
    printf("  char*: %zu bytes\n", sizeof(char*));
    printf("  void*: %zu bytes\n", sizeof(void*));
    
    // Sizeof arrays
    int arr[10];
    printf("\nSize of array[10]: %zu bytes\n", sizeof(arr));
    printf("Size of array element: %zu bytes\n", sizeof(arr[0]));
    printf("Number of elements: %zu\n", sizeof(arr) / sizeof(arr[0]));
    
    // Sizeof struct
    printf("\nSize of struct: %zu bytes\n", sizeof(struct MyStruct));
    
    // Sizeof union (size of largest member)
    printf("Size of union: %zu bytes\n", sizeof(union MyUnion));
    
    // Sizeof with expressions (not evaluated)
    int x = 5;
    printf("\nSizeof with expression (not evaluated): %zu\n", sizeof(x + 10));
    
    return 0;
}