/*
 * comments.c - Comment Styles
 * Tests: single-line, multi-line comments
 * Expected Output: "Comments test"
 */

#include <stdio.h>

/*
 * This is a multi-line comment
 * It can span multiple lines
 * and contain multiple paragraphs
 */

// This is a single-line comment
int main(void) {
    int x = 10;  // Inline comment
    
    /* Multi-line comment
       in the middle of code */
    int y = 20;
    
    // Comment with special characters: @#$%^&*()
    // Using variables to avoid unused warnings
    printf("Comments test: x=%d, y=%d\n", x, y);
    
    /* 
     * Nested-style comment block
     * (note: C doesn't support actual nesting)
     */
    
    return 0;  // Return success
}