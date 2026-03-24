/*
 * codegen_medium.c - Medium Code Generation Benchmark
 * Tests: Multiple functions and operations
 */

int add(int a, int b) {
    return a + b;
}

int subtract(int a, int b) {
    return a - b;
}

int multiply(int a, int b) {
    return a * b;
}

int divide(int a, int b) {
    if (b == 0) {
        return 0;
    }
    return a / b;
}

int modulo(int a, int b) {
    if (b == 0) {
        return 0;
    }
    return a % b;
}

int sum_range(int start, int end) {
    int sum;
    int i;
    sum = 0;
    for (i = start; i <= end; i = i + 1) {
        sum = sum + i;
    }
    return sum;
}

int factorial(int n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

int main(void) {
    int a;
    int b;
    int result;
    
    a = 100;
    b = 23;
    
    result = add(a, b);
    result = result + subtract(a, b);
    result = result + multiply(a, b);
    result = result + divide(a, b);
    result = result + modulo(a, b);
    
    result = result + sum_range(1, 50);
    result = result + factorial(6);
    
    return result;
}