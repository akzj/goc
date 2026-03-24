/*
 * complex.c - Complex Benchmark Program
 * Complexity: High
 * Tests: Structs, pointers, multiple functions, recursion, arrays
 * Used for: Full pipeline benchmarks with complex semantic analysis
 */

struct Point {
    int x;
    int y;
};

struct Rectangle {
    struct Point top_left;
    struct Point bottom_right;
};

int max(int a, int b) {
    if (a > b) {
        return a;
    }
    return b;
}

int min(int a, int b) {
    if (a < b) {
        return a;
    }
    return b;
}

int absolute(int value) {
    if (value < 0) {
        return 0 - value;
    }
    return value;
}

int fibonacci(int n) {
    if (n <= 0) {
        return 0;
    }
    if (n == 1) {
        return 1;
    }
    return fibonacci(n - 1) + fibonacci(n - 2);
}

int sum_array(int* arr, int size) {
    int sum = 0;
    int i;
    for (i = 0; i < size; i = i + 1) {
        sum = sum + arr[i];
    }
    return sum;
}

int calculate_area(struct Rectangle rect) {
    int width;
    int height;
    
    width = absolute(rect.bottom_right.x - rect.top_left.x);
    height = absolute(rect.bottom_right.y - rect.top_left.y);
    
    return width * height;
}

struct Point create_point(int x, int y) {
    struct Point p;
    p.x = x;
    p.y = y;
    return p;
}

struct Rectangle create_rectangle(struct Point tl, struct Point br) {
    struct Rectangle rect;
    rect.top_left = tl;
    rect.bottom_right = br;
    return rect;
}

int main(void) {
    int data[10];
    int i;
    int fib_result;
    int array_sum;
    struct Point p1;
    struct Point p2;
    struct Rectangle rect;
    int area;
    
    // Initialize array with fibonacci numbers
    for (i = 0; i < 10; i = i + 1) {
        data[i] = fibonacci(i);
    }
    
    // Calculate sum
    array_sum = sum_array(data, 10);
    
    // Create points and rectangle
    p1 = create_point(0, 0);
    p2 = create_point(10, 20);
    rect = create_rectangle(p1, p2);
    
    // Calculate area
    area = calculate_area(rect);
    
    // Final result
    return array_sum + area;
}