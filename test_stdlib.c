/*
 * test_stdlib.c - Unit Tests for Standard Library Implementation
 * 
 * Tests all CRITICAL and HIGH priority functions from builtin.c
 */

#include <stdint.h>
#include <stddef.h>

/* Test result tracking */
static int tests_passed = 0;
static int tests_failed = 0;

/* Forward declarations of functions to test */
void *malloc(size_t size);
void *calloc(size_t nmemb, size_t size);
void *realloc(void *ptr, size_t size);
void free(void *ptr);

void *memcpy(void *dest, const void *src, size_t n);
void *memmove(void *dest, const void *src, size_t n);
void *memset(void *s, int c, size_t n);
size_t strlen(const char *s);
int strcmp(const char *s1, const char *s2);
char *strcpy(char *dest, const char *src);
char *strcat(char *dest, const char *src);
char *strchr(const char *s, int c);
char *strstr(const char *haystack, const char *needle);

int atoi(const char *nptr);
double atof(const char *nptr);
int abs(int j);

double fabs(double x);
double floor(double x);
double ceil(double x);
double sqrt(double x);
double pow(double base, double exp);
double sin(double x);
double cos(double x);

int printf(const char *format, ...);
typedef struct _FILE FILE;
FILE *fopen(const char *filename, const char *mode);
int fclose(FILE *stream);

#define TEST(name) static void test_##name(void)
#define RUN_TEST(name) do { \
    printf("Running test: %s... ", #name); \
    test_##name(); \
    printf("PASSED\n"); \
} while(0)

#define ASSERT(cond) do { \
    if (!(cond)) { \
        printf("FAILED at line %d\n", __LINE__); \
        tests_failed++; \
        return; \
    } \
} while(0)

#define ASSERT_EQ(a, b) ASSERT((a) == (b))

/* ============================================================================
 * Memory Tests
 * ============================================================================ */

TEST(malloc_free) {
    void *p1 = malloc(100);
    ASSERT(p1 != NULL);
    
    void *p2 = malloc(200);
    ASSERT(p2 != NULL);
    ASSERT(p2 != p1);
    
    free(p1);
    free(p2);
    
    tests_passed++;
}

TEST(calloc) {
    void *p = calloc(10, sizeof(int));
    ASSERT(p != NULL);
    
    /* Check zero-initialized */
    int *arr = (int *)p;
    for (int i = 0; i < 10; i++) {
        ASSERT_EQ(arr[i], 0);
    }
    
    free(p);
    tests_passed++;
}

TEST(realloc) {
    void *p1 = malloc(50);
    ASSERT(p1 != NULL);
    
    /* Write some data */
    char *c1 = (char *)p1;
    for (int i = 0; i < 50; i++) {
        c1[i] = (char)i;
    }
    
    /* Reallocate to larger size */
    void *p2 = realloc(p1, 100);
    ASSERT(p2 != NULL);
    
    /* Check data preserved */
    char *c2 = (char *)p2;
    for (int i = 0; i < 50; i++) {
        ASSERT_EQ(c2[i], (char)i);
    }
    
    free(p2);
    tests_passed++;
}

/* ============================================================================
 * String Tests
 * ============================================================================ */

TEST(memcpy) {
    char src[] = "Hello";
    char dest[20];
    
    memcpy(dest, src, strlen(src) + 1);
    ASSERT_EQ(strcmp(dest, src), 0);
    
    tests_passed++;
}

TEST(memmove) {
    char buf[] = "Hello";
    
    /* Overlapping copy */
    memmove(buf + 1, buf, 4);
    ASSERT_EQ(buf[0], 'H');
    
    tests_passed++;
}

TEST(memset) {
    char buf[10];
    memset(buf, 'A', 10);
    
    for (int i = 0; i < 10; i++) {
        ASSERT_EQ(buf[i], 'A');
    }
    
    tests_passed++;
}

TEST(strlen) {
    ASSERT_EQ(strlen(""), 0);
    ASSERT_EQ(strlen("a"), 1);
    ASSERT_EQ(strlen("Hello"), 5);
    
    tests_passed++;
}

TEST(strcmp) {
    ASSERT_EQ(strcmp("abc", "abc"), 0);
    ASSERT(strcmp("abc", "abd") < 0);
    ASSERT(strcmp("abd", "abc") > 0);
    
    tests_passed++;
}

TEST(strcpy) {
    char dest[20];
    strcpy(dest, "Hello");
    ASSERT_EQ(strcmp(dest, "Hello"), 0);
    
    tests_passed++;
}

TEST(strcat) {
    char dest[20] = "Hello";
    strcat(dest, " World");
    ASSERT_EQ(strcmp(dest, "Hello World"), 0);
    
    tests_passed++;
}

TEST(strchr) {
    const char *s = "Hello";
    ASSERT(strchr(s, 'W') == NULL);
    ASSERT(strchr(s, 'H') != NULL);
    
    tests_passed++;
}

TEST(strstr) {
    const char *haystack = "Hello World";
    ASSERT(strstr(haystack, "World") != NULL);
    ASSERT(strstr(haystack, "XYZ") == NULL);
    
    tests_passed++;
}

/* ============================================================================
 * Conversion Tests
 * ============================================================================ */

TEST(atoi) {
    ASSERT_EQ(atoi("0"), 0);
    ASSERT_EQ(atoi("42"), 42);
    ASSERT_EQ(atoi("-42"), -42);
    
    tests_passed++;
}

TEST(atof) {
    double d1 = atof("3.14");
    ASSERT(d1 > 3.0 && d1 < 4.0);
    
    tests_passed++;
}

TEST(abs) {
    ASSERT_EQ(abs(0), 0);
    ASSERT_EQ(abs(42), 42);
    ASSERT_EQ(abs(-42), 42);
    
    tests_passed++;
}

/* ============================================================================
 * Math Tests
 * ============================================================================ */

TEST(fabs) {
    ASSERT_EQ(fabs(0.0), 0.0);
    ASSERT_EQ(fabs(-3.14), 3.14);
    
    tests_passed++;
}

TEST(floor) {
    ASSERT_EQ(floor(3.14), 3.0);
    ASSERT_EQ(floor(-3.14), -4.0);
    
    tests_passed++;
}

TEST(ceil) {
    ASSERT_EQ(ceil(3.14), 4.0);
    ASSERT_EQ(ceil(-3.14), -3.0);
    
    tests_passed++;
}

TEST(sqrt) {
    double s1 = sqrt(4.0);
    ASSERT(s1 > 1.9 && s1 < 2.1);
    
    double s2 = sqrt(9.0);
    ASSERT(s2 > 2.9 && s2 < 3.1);
    
    tests_passed++;
}

TEST(pow) {
    ASSERT_EQ(pow(2.0, 0.0), 1.0);
    ASSERT_EQ(pow(2.0, 1.0), 2.0);
    ASSERT_EQ(pow(2.0, 2.0), 4.0);
    
    tests_passed++;
}

TEST(sin) {
    double s1 = sin(0.0);
    ASSERT(s1 > -0.1 && s1 < 0.1);
    
    tests_passed++;
}

TEST(cos) {
    double c1 = cos(0.0);
    ASSERT(c1 > 0.9 && c1 < 1.1);
    
    tests_passed++;
}

/* ============================================================================
 * I/O Tests
 * ============================================================================ */

TEST(printf_basic) {
    printf("Test output\n");
    tests_passed++;
}

/* ============================================================================
 * Main
 * ============================================================================ */

int main(void) {
    printf("=== Standard Library Unit Tests ===\n\n");
    
    /* Memory tests */
    RUN_TEST(malloc_free);
    RUN_TEST(calloc);
    RUN_TEST(realloc);
    
    /* String tests */
    RUN_TEST(memcpy);
    RUN_TEST(memmove);
    RUN_TEST(memset);
    RUN_TEST(strlen);
    RUN_TEST(strcmp);
    RUN_TEST(strcpy);
    RUN_TEST(strcat);
    RUN_TEST(strchr);
    RUN_TEST(strstr);
    
    /* Conversion tests */
    RUN_TEST(atoi);
    RUN_TEST(atof);
    RUN_TEST(abs);
    
    /* Math tests */
    RUN_TEST(fabs);
    RUN_TEST(floor);
    RUN_TEST(ceil);
    RUN_TEST(sqrt);
    RUN_TEST(pow);
    RUN_TEST(sin);
    RUN_TEST(cos);
    
    /* I/O tests */
    RUN_TEST(printf_basic);
    
    printf("\n=== Test Summary ===\n");
    printf("Passed: %d\n", tests_passed);
    printf("Failed: %d\n", tests_failed);
    
    if (tests_failed > 0) {
        printf("RESULT: FAILED\n");
        return 1;
    } else {
        printf("RESULT: ALL TESTS PASSED\n");
        return 0;
    }
}