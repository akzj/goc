// Test typedef and enums
typedef unsigned long size_t;

enum Color {
    RED,
    GREEN,
    BLUE
};

int main() {
    enum Color c = RED;
    size_t s = 100;
    return c + (int)s;
}