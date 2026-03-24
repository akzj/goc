/* test_dlfcn.c - Test program for dlfcn implementation */

/* Include the dlfcn header */
#include "../pkg/stdlib/dlfcn.h"

/* Include syscall wrapper for write */
#include "../pkg/stdlib/syscall_wrapper.h"

/* Simple printf using write */
static void print_str(const char *s) {
    int len = 0;
    const char *p = s;
    while (*p++) len++;
    write(1, (void *)s, len);
}

static void print_int(int n) {
    char buf[32];
    int i = sizeof(buf) - 1;
    int negative = 0;
    
    if (n < 0) {
        negative = 1;
        n = -n;
    }
    
    buf[i] = '\0';
    do {
        buf[--i] = '0' + (n % 10);
        n /= 10;
    } while (n > 0);
    
    if (negative) {
        buf[--i] = '-';
    }
    
    print_str(&buf[i]);
}

static void print_nl(void) {
    print_str("\n");
}

/* Function pointer types for test library */
typedef int (*add_func_t)(int, int);
typedef int (*multiply_func_t)(int, int);
typedef const char *(*message_func_t)(void);

int main(void) {
    void *handle;
    add_func_t test_add;
    multiply_func_t test_multiply;
    const char *test_message;
    char *error;
    int result;
    
    print_str("=== Testing dlfcn implementation ===");
    print_nl();
    
    /* Test 1: dlopen */
    print_str("Test 1: dlopen - Loading test_lib.so... ");
    handle = dlopen("./test_dlfcn/test_lib.so", 0);
    
    if (!handle) {
        print_str("FAILED");
        print_nl();
        error = dlerror();
        if (error) {
            print_str("Error: ");
            print_str(error);
            print_nl();
        }
        return 1;
    }
    print_str("OK (handle=");
    print_int((int)(long)handle);
    print_str(")");
    print_nl();
    
    /* Test 2: dlsym - test_add */
    print_str("Test 2: dlsym - Finding test_add... ");
    test_add = (add_func_t)dlsym(handle, "test_add");
    
    if (!test_add) {
        print_str("FAILED");
        print_nl();
        error = dlerror();
        if (error) {
            print_str("Error: ");
            print_str(error);
            print_nl();
        }
        dlclose(handle);
        return 1;
    }
    print_str("OK");
    print_nl();
    
    /* Test 3: Call test_add */
    print_str("Test 3: Call test_add(5, 3)... ");
    result = test_add(5, 3);
    print_int(result);
    if (result == 8) {
        print_str(" OK");
    } else {
        print_str(" FAILED (expected 8)");
    }
    print_nl();
    
    /* Test 4: dlsym - test_multiply */
    print_str("Test 4: dlsym - Finding test_multiply... ");
    test_multiply = (multiply_func_t)dlsym(handle, "test_multiply");
    
    if (!test_multiply) {
        print_str("FAILED");
        print_nl();
        error = dlerror();
        if (error) {
            print_str("Error: ");
            print_str(error);
            print_nl();
        }
        dlclose(handle);
        return 1;
    }
    print_str("OK");
    print_nl();
    
    /* Test 5: Call test_multiply */
    print_str("Test 5: Call test_multiply(4, 7)... ");
    result = test_multiply(4, 7);
    print_int(result);
    if (result == 28) {
        print_str(" OK");
    } else {
        print_str(" FAILED (expected 28)");
    }
    print_nl();
    
    /* Test 6: dlsym - test_message (data symbol) */
    print_str("Test 6: dlsym - Finding test_message... ");
    test_message = (const char *)dlsym(handle, "test_message");
    
    if (!test_message) {
        print_str("FAILED");
        print_nl();
        error = dlerror();
        if (error) {
            print_str("Error: ");
            print_str(error);
            print_nl();
        }
        dlclose(handle);
        return 1;
    }
    print_str("OK (message=");
    print_str(test_message);
    print_str(")");
    print_nl();
    
    /* Test 7: dlclose */
    print_str("Test 7: dlclose - Unloading library... ");
    if (dlclose(handle) != 0) {
        print_str("FAILED");
        print_nl();
        error = dlerror();
        if (error) {
            print_str("Error: ");
            print_str(error);
            print_nl();
        }
        return 1;
    }
    print_str("OK");
    print_nl();
    
    /* Test 8: dlopen invalid file */
    print_str("Test 8: dlopen invalid file... ");
    handle = dlopen("./nonexistent.so", 0);
    if (handle) {
        print_str("FAILED (should return NULL)");
        print_nl();
        dlclose(handle);
        return 1;
    }
    error = dlerror();
    if (!error) {
        print_str("FAILED (should have error message)");
        print_nl();
        return 1;
    }
    print_str("OK (error=");
    print_str(error);
    print_str(")");
    print_nl();
    
    /* Test 9: dlsym invalid symbol */
    print_str("Test 9: dlsym invalid symbol... ");
    handle = dlopen("./test_dlfcn/test_lib.so", 0);
    if (!handle) {
        print_str("FAILED (cannot open library)");
        print_nl();
        return 1;
    }
    
    test_add = (add_func_t)dlsym(handle, "nonexistent_symbol");
    if (test_add) {
        print_str("FAILED (should return NULL)");
        print_nl();
        dlclose(handle);
        return 1;
    }
    error = dlerror();
    if (!error) {
        print_str("FAILED (should have error message)");
        print_nl();
        dlclose(handle);
        return 1;
    }
    print_str("OK (error=");
    print_str(error);
    print_str(")");
    print_nl();
    dlclose(handle);
    
    /* Test 10: Reference counting */
    print_str("Test 10: Reference counting... ");
    handle = dlopen("./test_dlfcn/test_lib.so", 0);
    if (!handle) {
        print_str("FAILED (first dlopen)");
        print_nl();
        return 1;
    }
    
    void *handle2 = dlopen("./test_dlfcn/test_lib.so", 0);
    if (!handle2) {
        print_str("FAILED (second dlopen)");
        print_nl();
        dlclose(handle);
        return 1;
    }
    
    if (handle != handle2) {
        print_str("FAILED (should return same handle)");
        print_nl();
        dlclose(handle);
        dlclose(handle2);
        return 1;
    }
    
    /* First dlclose should not unload */
    if (dlclose(handle) != 0) {
        print_str("FAILED (first dlclose)");
        print_nl();
        return 1;
    }
    
    /* Second dlclose should unload */
    if (dlclose(handle2) != 0) {
        print_str("FAILED (second dlclose)");
        print_nl();
        return 1;
    }
    print_str("OK");
    print_nl();
    
    print_str("=== All tests passed! ===");
    print_nl();
    
    return 0;
}