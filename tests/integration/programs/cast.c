/*
 * cast.c - Type Casting
 * Tests: type casting, implicit/explicit conversions
 * Expected Output: Results of type conversions
 */

#include <stdio.h>

int main(void) {
    // Implicit conversion (promotion)
    int a = 5;
    double b = 2.5;
    double result = a + b;  // int promoted to double
    printf("Implicit conversion: %d + %.1f = %.1f\n", a, b, result);
    
    // Explicit casting
    double x = 9.99;
    int y = (int)x;  // Truncates to 9
    printf("\nExplicit cast: (int)%.2f = %d\n", x, y);
    
    // Integer division vs float division
    int m = 7, n = 2;
    printf("\nInteger division: %d / %d = %d\n", m, n, m / n);
    printf("Float division: (double)%d / %d = %.2f\n", m, n, (double)m / n);
    
    // Cast between pointer types
    int value = 65;
    int *int_ptr = &value;
    char *char_ptr = (char *)int_ptr;
    printf("\nPointer cast:\n");
    printf("  int value: %d\n", *int_ptr);
    printf("  as char: %c\n", *char_ptr);
    
    // Cast in arithmetic
    long big_num = 1000000L;
    int small_num = (int)(big_num / 1000);
    printf("\nCast in arithmetic: %ld / 1000 = %d\n", big_num, small_num);
    
    // Unsigned to signed cast
    unsigned int u = 300;
    signed int s = (signed int)u;
    printf("\nUnsigned to signed: %u -> %d\n", u, s);
    
    return 0;
}