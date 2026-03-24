/*
 * moderate.c - Moderate Complexity Benchmark Program
 * Complexity: Medium
 * Tests: Functions, loops, conditionals, arrays
 * Used for: Parser/semantic/IR benchmarks with control flow
 */

int add(int x, int y) {
    return x + y;
}

int multiply(int x, int y) {
    return x * y;
}

int factorial(int n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

int main(void) {
    int numbers[5];
    int i;
    int sum = 0;
    int result;
    
    // Initialize array
    numbers[0] = 1;
    numbers[1] = 2;
    numbers[2] = 3;
    numbers[3] = 4;
    numbers[4] = 5;
    
    // Sum array elements
    for (i = 0; i < 5; i = i + 1) {
        sum = add(sum, numbers[i]);
    }
    
    // Calculate factorial
    result = factorial(5);
    
    // Multiply sum by result
    result = multiply(sum, result);
    
    return result;
}