/*
 * variables.c - Variable Declarations
 * Tests: int/char/float declarations, initialization
 * Expected Output: Values of different variable types
 */

#include <stdio.h>

int main(void) {
    // Integer declaration and initialization
    int integer_var = 42;
    
    // Character declaration and initialization
    char char_var = 'A';
    
    // Float declaration and initialization
    float float_var = 3.14f;
    
    // Double declaration
    double double_var = 2.71828;
    
    // Print all variables
    printf("Integer: %d\n", integer_var);
    printf("Character: %c\n", char_var);
    printf("Float: %f\n", float_var);
    printf("Double: %lf\n", double_var);
    
    return 0;
}