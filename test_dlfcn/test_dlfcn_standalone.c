/* test_dlfcn_standalone.c - Standalone test for dlfcn implementation using system dlfcn */

#include <stdio.h>
#include <stdlib.h>
#include <dlfcn.h>

typedef int (*add_func_t)(int, int);
typedef int (*multiply_func_t)(int, int);

int main(void) {
    void *handle;
    add_func_t test_add;
    multiply_func_t test_multiply;
    const char *test_message;
    char *error;
    int result;
    
    printf("=== Testing dlfcn implementation ===\n");
    
    /* Test 1: dlopen */
    printf("Test 1: dlopen - Loading test_lib.so... ");
    handle = dlopen("./test_dlfcn/test_lib.so", RTLD_NOW);
    
    if (!handle) {
        printf("FAILED\n");
        error = dlerror();
        if (error) {
            printf("Error: %s\n", error);
        }
        return 1;
    }
    printf("OK (handle=%p)\n", handle);
    
    /* Test 2: dlsym - test_add */
    printf("Test 2: dlsym - Finding test_add... ");
    test_add = (add_func_t)dlsym(handle, "test_add");
    
    if (!test_add) {
        printf("FAILED\n");
        error = dlerror();
        if (error) {
            printf("Error: %s\n", error);
        }
        dlclose(handle);
        return 1;
    }
    printf("OK\n");
    
    /* Test 3: Call test_add */
    printf("Test 3: Call test_add(5, 3)... ");
    result = test_add(5, 3);
    printf("%d", result);
    if (result == 8) {
        printf(" OK\n");
    } else {
        printf(" FAILED (expected 8)\n");
        dlclose(handle);
        return 1;
    }
    
    /* Test 4: dlsym - test_multiply */
    printf("Test 4: dlsym - Finding test_multiply... ");
    test_multiply = (multiply_func_t)dlsym(handle, "test_multiply");
    
    if (!test_multiply) {
        printf("FAILED\n");
        error = dlerror();
        if (error) {
            printf("Error: %s\n", error);
        }
        dlclose(handle);
        return 1;
    }
    printf("OK\n");
    
    /* Test 5: Call test_multiply */
    printf("Test 5: Call test_multiply(4, 7)... ");
    result = test_multiply(4, 7);
    printf("%d", result);
    if (result == 28) {
        printf(" OK\n");
    } else {
        printf(" FAILED (expected 28)\n");
        dlclose(handle);
        return 1;
    }
    
    /* Test 6: dlsym - test_message (data symbol) */
    printf("Test 6: dlsym - Finding test_message... ");
    test_message = (const char *)dlsym(handle, "test_message");
    
    if (!test_message) {
        printf("FAILED\n");
        error = dlerror();
        if (error) {
            printf("Error: %s\n", error);
        }
        dlclose(handle);
        return 1;
    }
    printf("OK (message=%s)\n", test_message);
    
    /* Test 7: dlclose */
    printf("Test 7: dlclose - Unloading library... ");
    if (dlclose(handle) != 0) {
        printf("FAILED\n");
        error = dlerror();
        if (error) {
            printf("Error: %s\n", error);
        }
        return 1;
    }
    printf("OK\n");
    
    /* Test 8: dlopen invalid file */
    printf("Test 8: dlopen invalid file... ");
    handle = dlopen("./nonexistent.so", RTLD_NOW);
    if (handle) {
        printf("FAILED (should return NULL)\n");
        dlclose(handle);
        return 1;
    }
    error = dlerror();
    if (!error) {
        printf("FAILED (should have error message)\n");
        return 1;
    }
    printf("OK (error=%s)\n", error);
    
    /* Test 9: dlsym invalid symbol */
    printf("Test 9: dlsym invalid symbol... ");
    handle = dlopen("./test_dlfcn/test_lib.so", RTLD_NOW);
    if (!handle) {
        printf("FAILED (cannot open library)\n");
        return 1;
    }
    
    test_add = (add_func_t)dlsym(handle, "nonexistent_symbol");
    if (test_add) {
        printf("FAILED (should return NULL)\n");
        dlclose(handle);
        return 1;
    }
    error = dlerror();
    if (!error) {
        printf("FAILED (should have error message)\n");
        dlclose(handle);
        return 1;
    }
    printf("OK (error=%s)\n", error);
    dlclose(handle);
    
    /* Test 10: Reference counting */
    printf("Test 10: Reference counting... ");
    handle = dlopen("./test_dlfcn/test_lib.so", RTLD_NOW);
    if (!handle) {
        printf("FAILED (first dlopen)\n");
        return 1;
    }
    
    void *handle2 = dlopen("./test_dlfcn/test_lib.so", RTLD_NOW);
    if (!handle2) {
        printf("FAILED (second dlopen)\n");
        dlclose(handle);
        return 1;
    }
    
    if (handle != handle2) {
        printf("FAILED (should return same handle)\n");
        dlclose(handle);
        dlclose(handle2);
        return 1;
    }
    
    /* First dlclose should not unload */
    if (dlclose(handle) != 0) {
        printf("FAILED (first dlclose)\n");
        return 1;
    }
    
    /* Second dlclose should unload */
    if (dlclose(handle2) != 0) {
        printf("FAILED (second dlclose)\n");
        return 1;
    }
    printf("OK\n");
    
    printf("=== All tests passed! ===\n");
    
    return 0;
}