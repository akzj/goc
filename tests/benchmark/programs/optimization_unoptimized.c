/*
 * optimization_unoptimized.c - Unoptimized Code Patterns
 * Tests: Inefficient patterns for comparison
 */

int square(int x) {
    return x * x;
}

int sum_squares(int* arr, int size) {
    int sum;
    int i;
    sum = 0;
    for (i = 0; i < size; i = i + 1) {
        sum = sum + square(arr[i]);
    }
    return sum;
}

int sum_cubes(int* arr, int size) {
    int sum;
    int i;
    sum = 0;
    for (i = 0; i < size; i = i + 1) {
        sum = sum + arr[i] * arr[i] * arr[i];
    }
    return sum;
}

int main(void) {
    int data[10];
    int i;
    int result;
    int sq_sum;
    int cb_sum;
    
    for (i = 0; i < 10; i = i + 1) {
        data[i] = i + 1;
    }
    
    sq_sum = sum_squares(data, 10);
    cb_sum = sum_cubes(data, 10);
    
    result = sq_sum + cb_sum;
    
    return result;
}