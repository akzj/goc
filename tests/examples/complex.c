// Complex C file for testing lexer
#include <stdlib.h>

#define MAX_SIZE 100

/* Multi-line comment
   spanning multiple lines */

typedef struct {
    int id;
    char name[64];
    float score;
} Student;

int calculate(int a, int b) {
    int result = a + b;
    result *= 2;
    
    if (result > MAX_SIZE) {
        return result - 1;
    } else if (result < 0) {
        return 0;
    }
    
    // Test various operators
    int x = 10;
    x++;
    x--;
    x += 5;
    x -= 3;
    x *= 2;
    x /= 4;
    x %= 3;
    x &= 0xFF;
    x |= 0x0F;
    x ^= 0xAA;
    x <<= 2;
    x >>= 1;
    
    // Test logical operators
    if (a > 0 && b > 0) {
        return a || b;
    }
    
    // Test literals
    int hex = 0xDEADBEEF;
    int octal = 0755;
    float pi = 3.14159f;
    double e = 2.71828;
    char ch = 'A';
    char *str = "Hello, World!\n\t\"";
    
    return result;
}

int main(int argc, char *argv[]) {
    Student s = {1, "Alice", 95.5f};
    int (*func_ptr)(int, int) = calculate;
    
    // Test switch
    switch (argc) {
        case 1:
            printf("No arguments\n");
            break;
        case 2:
            printf("One argument: %s\n", argv[1]);
            break;
        default:
            printf("Multiple arguments\n");
    }
    
    // Test loops
    for (int i = 0; i < 10; i++) {
        while (i < 5) {
            do {
                i++;
                if (i == 3) continue;
                if (i == 4) break;
            } while (i < 3);
        }
    }
    
    // Test ternary
    int max = (a > b) ? a : b;
    
    // Test sizeof
    size_t size = sizeof(Student);
    
    // Test goto
    goto end;
    printf("This is skipped\n");
    
end:
    printf("Done\n");
    
    return 0;
}