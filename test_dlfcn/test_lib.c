/* test_lib.c - Simple test library for dlfcn testing */

int test_add(int a, int b) {
    return a + b;
}

int test_multiply(int a, int b) {
    return a * b;
}

const char *test_message = "Hello from test library!";

int test_value = 42;