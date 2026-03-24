/*
 * optimization_optimized.c - Optimized Code Patterns
 * Tests: Efficient patterns for comparison
 */

int sum_powers(int* arr, int size) {
    int sum;
    int i;
    int val;
    int sq;
    sum = 0;
    for (i = 0; i < size; i = i + 1) {
        val = arr[i];
        sq = val * val;
        sum = sum + sq + sq * val;
    }
    return sum;
}

int main(void) {
    int data[10];
    int i;
    int result;
    
    for (i = 0; i < 10; i = i + 1) {
        data[i] = i + 1;
    }
    
    result = sum_powers(data, 10);
    
    return result;
}