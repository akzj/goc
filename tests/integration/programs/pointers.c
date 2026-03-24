/*
 * pointers.c - Pointer Operations
 * Tests: pointer declaration, dereference, address-of
 * Expected Output: Pointer addresses and values
 */

#include <stdio.h>

int main(void) {
    int value = 42;
    int *ptr;
    
    // Pointer declaration and assignment
    ptr = &value;
    
    // Print value and address
    printf("Value: %d\n", value);
    printf("Address of value: %p\n", (void*)&value);
    printf("Pointer value (address): %p\n", (void*)ptr);
    printf("Dereferenced pointer: %d\n", *ptr);
    
    // Modify value through pointer
    *ptr = 100;
    printf("\nAfter *ptr = 100:\n");
    printf("Value: %d\n", value);
    
    // Pointer arithmetic
    int arr[3] = {10, 20, 30};
    int *arr_ptr = arr;
    
    printf("\nPointer arithmetic:\n");
    printf("  arr_ptr[0] = %d\n", *arr_ptr);
    printf("  arr_ptr[1] = %d\n", *(arr_ptr + 1));
    printf("  arr_ptr[2] = %d\n", *(arr_ptr + 2));
    
    // NULL pointer
    int *null_ptr = NULL;
    printf("\nNULL pointer: %p\n", (void*)null_ptr);
    
    return 0;
}