/*
 * codegen_large.c - Large Code Generation Benchmark
 * Tests: Multiple functions and struct operations
 */

struct Vector2 {
    int x;
    int y;
};

int add(int a, int b) {
    return a + b;
}

int multiply(int a, int b) {
    return a * b;
}

int vector_sum(struct Vector2 a, struct Vector2 b) {
    struct Vector2 result;
    result.x = a.x + b.x;
    result.y = a.y + b.y;
    return result.x + result.y;
}

int sum_array(int* arr, int size) {
    int sum;
    int i;
    sum = 0;
    for (i = 0; i < size; i = i + 1) {
        sum = sum + arr[i];
    }
    return sum;
}

int main(void) {
    struct Vector2 v1;
    struct Vector2 v2;
    int arr[10];
    int i;
    int result;
    
    v1.x = 5;
    v1.y = 10;
    v2.x = 3;
    v2.y = 7;
    
    result = add(10, 20);
    result = result + multiply(5, 6);
    result = result + vector_sum(v1, v2);
    
    for (i = 0; i < 10; i = i + 1) {
        arr[i] = i + 1;
    }
    
    result = result + sum_array(arr, 10);
    
    return result;
}