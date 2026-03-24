/*
 * memory_intensive.c - Memory-Intensive Benchmark
 * Tests: Structs and arrays
 */

struct Point {
    int x;
    int y;
};

struct Rectangle {
    int width;
    int height;
    struct Point origin;
};

int sum_array(int* arr, int size) {
    int sum;
    int i;
    sum = 0;
    for (i = 0; i < size; i = i + 1) {
        sum = sum + arr[i];
    }
    return sum;
}

int area(struct Rectangle r) {
    return r.width * r.height;
}

struct Point midpoint(struct Point p1, struct Point p2) {
    struct Point result;
    result.x = (p1.x + p2.x) / 2;
    result.y = (p1.y + p2.y) / 2;
    return result;
}

int main(void) {
    int data[30];
    struct Rectangle rects[5];
    struct Point p1;
    struct Point p2;
    struct Point mid;
    int i;
    int total;
    
    for (i = 0; i < 30; i = i + 1) {
        data[i] = i;
    }
    
    for (i = 0; i < 5; i = i + 1) {
        rects[i].width = i + 1;
        rects[i].height = i + 2;
        rects[i].origin.x = i;
        rects[i].origin.y = i * 2;
    }
    
    total = sum_array(data, 30);
    
    p1.x = 0;
    p1.y = 0;
    p2.x = 10;
    p2.y = 20;
    
    mid = midpoint(p1, p2);
    
    total = total + area(rects[0]);
    total = total + area(rects[1]);
    total = total + mid.x + mid.y;
    
    return total;
}