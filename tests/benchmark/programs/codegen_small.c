/*
 * codegen_small.c - Small Code Generation Benchmark
 * Tests: Minimal code generation
 */

int add(int a, int b) {
    return a + b;
}

int main(void) {
    int x;
    int y;
    int result;
    
    x = 10;
    y = 20;
    result = add(x, y);
    
    return result;
}