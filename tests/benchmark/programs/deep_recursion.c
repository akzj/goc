/*
 * deep_recursion.c - Deep Recursion Benchmark Program
 * Complexity: High (Recursion Depth)
 * Tests: Deep recursive calls, stack usage, tail recursion patterns
 * Used for: Recursion depth and call stack benchmarks
 */

int factorial(int n) {
    if (n <= 0) {
        return 1;
    }
    return n * factorial(n - 1);
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

int ackermann(int m, int n) {
    if (m == 0) {
        return n + 1;
    }
    if (n == 0) {
        return ackermann(m - 1, 1);
    }
    return ackermann(m - 1, ackermann(m, n - 1));
}

int sum_range(int start, int end) {
    if (start > end) {
        return 0;
    }
    if (start == end) {
        return start;
    }
    int mid;
    mid = (start + end) / 2;
    return sum_range(start, mid) + sum_range(mid + 1, end);
}

int tree_depth(int depth) {
    if (depth <= 0) {
        return 1;
    }
    return 1 + tree_depth(depth - 1) + tree_depth(depth - 1);
}

int nested_recursion(int n) {
    if (n <= 0) {
        return 1;
    }
    return nested_recursion(n - 1) + nested_recursion(nested_recursion(n - 1));
}

int mutual_a(int n);
int mutual_b(int n);

int mutual_a(int n) {
    if (n <= 0) {
        return 1;
    }
    return mutual_b(n - 1) + 1;
}

int mutual_b(int n) {
    if (n <= 0) {
        return 1;
    }
    return mutual_a(n - 1) + 1;
}

int deep_call_chain(int n) {
    if (n <= 0) {
        return 0;
    }
    return 1 + deep_call_chain(n - 1);
}

int binary_tree_sum(int depth, int value) {
    if (depth <= 0) {
        return value;
    }
    return value + binary_tree_sum(depth - 1, value * 2) + binary_tree_sum(depth - 1, value * 2 + 1);
}

int main(void) {
    int fact_result;
    int fib_result;
    int ack_result;
    int sum_result;
    int tree_result;
    int mutual_result;
    int chain_result;
    int binary_result;
    
    // Factorial (moderate depth)
    fact_result = factorial(10);
    
    // Fibonacci (moderate depth)
    fib_result = fibonacci(15);
    
    // Ackermann (shallow but complex)
    ack_result = ackermann(3, 4);
    
    // Sum range (divide and conquer)
    sum_result = sum_range(1, 50);
    
    // Tree depth calculation
    tree_result = tree_depth(8);
    
    // Mutual recursion
    mutual_result = mutual_a(20) + mutual_b(20);
    
    // Deep call chain
    chain_result = deep_call_chain(30);
    
    // Binary tree sum
    binary_result = binary_tree_sum(6, 1);
    
    return fact_result + fib_result + ack_result + sum_result + 
           tree_result + mutual_result + chain_result + binary_result;
}