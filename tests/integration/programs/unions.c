// Test unions
union Data {
    int i;
    float f;
    char c;
};

int main() {
    union Data d;
    d.i = 42;
    return d.i;
}