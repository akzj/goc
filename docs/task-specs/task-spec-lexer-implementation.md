# Task Specification: Lexer Implementation

## Task ID
`task-lexer-implementation-001`

## Layer
Trunk → Branch

## Priority
High

## Status
Ready for Delegation

## Goal
Design and implement a complete C11 lexer that tokenizes C source code into a stream of tokens, handling all C11 lexical elements including keywords, identifiers, literals, operators, and comments.

## Context

### Design Documents
- Architecture Design: `/home/ubuntu/workspace/goc/docs/architecture-design.md`
- Implementation Plan: `/home/ubuntu/workspace/goc/docs/implementation-plan.md`

### Related Components
- **Token Definitions**: `/home/ubuntu/workspace/goc/pkg/lexer/token.go` (Already completed)
- **CLI**: Will use the lexer to tokenize files
- **Parser**: Will consume the token stream

### Current State
- ✅ Token definitions complete (94.7% test coverage)
- ✅ All C11 keywords defined (42 keywords)
- ✅ All operators defined (45 operators)
- ✅ Position tracking implemented
- ✅ Helper functions available
- ❌ Lexer implementation not started

## Requirements

### Functional Requirements

#### Core Functionality
1. **Tokenization**: Convert C source code into token stream
2. **Position Tracking**: Track line, column, and filename for each token
3. **Error Handling**: Report lexical errors with position information
4. **Comment Handling**: Support single-line (//) and multi-line (/* */) comments
5. **Whitespace Handling**: Skip whitespace while tracking position

#### Token Types to Handle
1. **Keywords**: All 42 C11 keywords (use token definitions)
2. **Identifiers**: Variable and function names
3. **Literals**:
   - Integer literals (decimal, octal, hexadecimal)
   - Floating-point literals (decimal, hexadecimal)
   - Character literals (with escape sequences)
   - String literals (with escape sequences)
4. **Operators**: All 45 operators (use token definitions)
5. **Punctuation**: All delimiters and separators
6. **Special**: EOF, ILLEGAL, COMMENT, NEWLINE

#### C11 Specific Features
1. **Digraphs and Trigraphs**: Support alternative representations
2. **Universal Character Names**: \uXXXX and \UXXXXXXXX
3. **Raw String Literals**: (C++ feature, optional)
4. **Binary Literals**: 0b prefix (C23, optional)
5. **Line Splicing**: Backslash-newline continuation

### Non-Functional Requirements
1. **Performance**: Efficient scanning (avoid unnecessary allocations)
2. **Memory**: Minimal memory footprint
3. **Error Recovery**: Continue after errors to find more issues
4. **Extensibility**: Easy to add new token types
5. **Testability**: High test coverage (> 80%)

## Technical Specifications

### Lexer Interface

```go
// Lexer tokenizes C source code
type Lexer interface {
    // NextToken returns the next token from the input
    NextToken() (Token, error)
    
    // Tokenize returns all tokens from the input
    Tokenize() ([]Token, error)
    
    // Position returns the current position in the source
    Position() Position
    
    // HasMore returns true if there are more tokens
    HasMore() bool
}
```

### Lexer Structure

```go
type Lexer struct {
    input   string    // Source code to tokenize
    pos     int       // Current position in input
    line    int       // Current line number (1-based)
    column  int       // Current column number (1-based)
    file    string    // Source file name
    tokens  []Token   // Token buffer
    err     error     // Last error encountered
}
```

### Main Methods

1. **NewLexer(input, filename)**: Create a new lexer instance
2. **NextToken()**: Return the next token
3. **Tokenize()**: Return all tokens
4. **skipWhitespace()**: Skip whitespace and comments
5. **scanIdentifier()**: Scan an identifier or keyword
6. **scanNumber()**: Scan a numeric literal
7. **scanString()**: Scan a string literal
8. **scanChar()**: Scan a character literal
9. **scanOperator()**: Scan an operator
10. **scanComment()**: Scan a comment

### Error Handling

```go
type LexerError struct {
    Position Position
    Message  string
}

func (e *LexerError) Error() string {
    return fmt.Sprintf("%s: %s", e.Position, e.Message)
}
```

### Token Categories

1. **Keywords**: Use `LookupKeyword()` from token.go
2. **Identifiers**: `[a-zA-Z_][a-zA-Z0-9_]*`
3. **Integer Literals**:
   - Decimal: `[0-9]+`
   - Octal: `0[0-7]+`
   - Hexadecimal: `0[xX][0-9a-fA-F]+`
   - Suffixes: `[uUlL]`
4. **Float Literals**:
   - Decimal: `[0-9]+\.[0-9]*([eE][+-]?[0-9]+)?`
   - Hexadecimal: `0[xX][0-9a-fA-F]+\.[0-9a-fA-F]*([pP][+-]?[0-9]+)?`
   - Suffixes: `[fFlL]`
5. **Character Literals**: `'([^'\\]|\\.)*'`
6. **String Literals**: `"([^"\\]|\\.)*"`
7. **Operators**: Use token definitions

## Constraints

### Must Do
- ✅ Implement all token types defined in token.go
- ✅ Handle all C11 keywords
- ✅ Handle all operators
- ✅ Handle all literal types
- ✅ Track position accurately
- ✅ Report errors with position
- ✅ Write comprehensive tests (> 80% coverage)
- ✅ Document all exported functions

### Must NOT Do
- ❌ Skip any token types
- ❌ Ignore error handling
- ❌ Use external dependencies
- ❌ Modify token.go (use it as-is)
- ❌ Implement parsing logic (that's for parser)

### Quality Standards
- Test coverage > 80%
- All exported types and functions documented
- Follow Go naming conventions
- No compiler warnings
- Efficient implementation (avoid unnecessary allocations)

## Deliverables

### Files to Create

1. **pkg/lexer/lexer.go**
   - Lexer interface definition
   - Lexer struct implementation
   - All scanning methods
   - Error handling
   - Helper functions

2. **pkg/lexer/lexer_test.go**
   - Unit tests for all token types
   - Integration tests for complete programs
   - Error handling tests
   - Edge case tests
   - Performance benchmarks (optional)

### Expected Content

#### pkg/lexer/lexer.go
```go
package lexer

import (
    "unicode"
    "unicode/utf8"
)

// Lexer tokenizes C source code
type Lexer struct {
    input  string
    pos    int
    line   int
    column int
    file   string
    // ... other fields
}

// NewLexer creates a new lexer for the given input
func NewLexer(input, filename string) *Lexer {
    return &Lexer{
        input:  input,
        pos:    0,
        line:   1,
        column: 1,
        file:   filename,
    }
}

// NextToken returns the next token
func (l *Lexer) NextToken() (Token, error) {
    // Implementation
}

// Tokenize returns all tokens
func (l *Lexer) Tokenize() ([]Token, error) {
    // Implementation
}

// Private methods
func (l *Lexer) skipWhitespace() { /* ... */ }
func (l *Lexer) scanIdentifier() Token { /* ... */ }
func (l *Lexer) scanNumber() Token { /* ... */ }
// ... other methods
```

#### pkg/lexer/lexer_test.go
```go
package lexer

import "testing"

func TestLexerKeywords(t *testing.T) {
    input := "int float if else while"
    lexer := NewLexer(input, "test.c")
    tokens, err := lexer.Tokenize()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // Verify tokens
}

func TestLexerIdentifiers(t *testing.T) { /* ... */ }
func TestLexerNumbers(t *testing.T) { /* ... */ }
func TestLexerStrings(t *testing.T) { /* ... */ }
func TestLexerOperators(t *testing.T) { /* ... */ }
func TestLexerComments(t *testing.T) { /* ... */ }
func TestLexerErrors(t *testing.T) { /* ... */ }
```

## Success Criteria

### Functional Criteria
- [ ] All C11 keywords tokenized correctly
- [ ] All operators tokenized correctly
- [ ] All literal types tokenized correctly
- [ ] Identifiers tokenized correctly
- [ ] Comments handled correctly
- [ ] Position tracking accurate
- [ ] Error reporting working

### Quality Criteria
- [ ] Test coverage > 80%
- [ ] All exported items documented
- [ ] No compiler warnings
- [ ] Code follows Go conventions
- [ ] Efficient implementation

### Integration Criteria
- [ ] Works with existing token definitions
- [ ] Can be used by CLI
- [ ] Ready for parser integration

## Example Test Cases

### Keywords
```c
int main() { return 0; }
// Should tokenize: INT, IDENT, LPAREN, RPAREN, LBRACE, RETURN, INT_LIT, SEMICOLON, RBRACE, EOF
```

### Operators
```c
a = b + c * d;
// Should tokenize: IDENT, ASSIGN, IDENT, ADD, IDENT, MUL, IDENT, SEMICOLON, EOF
```

### Literals
```c
int x = 42;
float y = 3.14;
char c = 'a';
char *s = "hello";
// Should tokenize all literals correctly
```

### Comments
```c
// Single line comment
/* Multi-line
   comment */
// Both should be skipped or returned as COMMENT tokens
```

### Errors
```c
int x = 42
// Missing semicolon - should report error with position
```

## Estimated Effort
- **Trunk Phase**: 2-3 hours (interface design)
- **Branch Phase**: 4-6 hours (implementation and testing)
- **Total**: 6-9 hours

## Dependencies
- ✅ Token definitions (completed)
- ❌ Error handling framework (can be basic for now)

## Next Steps After Completion
1. CLI integration (add tokenize command)
2. Error handling framework (enhance error reporting)
3. Parser interface design (Trunk)
4. Parser implementation (Branch)

## Notes
- This is a critical component that affects parser design
- Must handle all edge cases carefully
- Performance matters (lexer is called frequently)
- Consider using existing lexer generators as reference
- Test with real C programs for robustness

---

**Created**: 2025-06-17
**Author**: Zero-FAS (Root Node)
**Status**: Ready for Delegation