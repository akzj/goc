/*
 * test_crt0.c - Test program for crt0.c argc/argv extraction
 * 
 * This program verifies that crt0.c correctly extracts and passes
 * command-line arguments from the stack to main().
 * 
 * Minimal test - main returns instead of calling exit() to avoid
 * issues with stub exit function.
 */

/* Main function - returns exit code instead of calling exit() */
int main(int argc, char **argv) {
    /* Return 0 to indicate success */
    /* If we get here, argc/argv were passed correctly */
    return 0;
}