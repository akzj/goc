/*
 * enums.c - Enumeration Types
 * Tests: enum definition, usage in switch/if
 * Expected Output: Enum values and their usage
 */

#include <stdio.h>

// Basic enum
enum Day {
    SUNDAY,
    MONDAY,
    TUESDAY,
    WEDNESDAY,
    THURSDAY,
    FRIDAY,
    SATURDAY
};

// Enum with explicit values
enum Status {
    PENDING = 0,
    ACTIVE = 1,
    COMPLETED = 2,
    CANCELLED = 3
};

// Typedef for enum
typedef enum {
    RED,
    GREEN,
    BLUE
} Color;

int main(void) {
    // Using basic enum
    enum Day today = WEDNESDAY;
    printf("Today is day %d\n", today);
    
    // Enum in switch statement
    printf("Day name: ");
    switch (today) {
        case SUNDAY: printf("Sunday\n"); break;
        case MONDAY: printf("Monday\n"); break;
        case TUESDAY: printf("Tuesday\n"); break;
        case WEDNESDAY: printf("Wednesday\n"); break;
        case THURSDAY: printf("Thursday\n"); break;
        case FRIDAY: printf("Friday\n"); break;
        case SATURDAY: printf("Saturday\n"); break;
    }
    
    // Enum with explicit values
    enum Status status = ACTIVE;
    printf("\nStatus: %d\n", status);
    
    // Enum in if statement
    if (status == ACTIVE) {
        printf("Item is active\n");
    }
    
    // Using typedef enum
    Color my_color = BLUE;
    printf("\nColor value: %d\n", my_color);
    
    // Print color name
    const char *color_name;
    switch (my_color) {
        case RED: color_name = "Red"; break;
        case GREEN: color_name = "Green"; break;
        case BLUE: color_name = "Blue"; break;
        default: color_name = "Unknown";
    }
    printf("Color name: %s\n", color_name);
    
    return 0;
}