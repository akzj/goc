# Task Specification: Token Definitions

## Task ID
`task-lexer-token-001`

## Layer
Trunk → Branch

## Priority
High

## Status
Ready for Delegation

## Goal
Design and implement all token types required for the C11 lexer, including keywords, operators, literals, identifiers, and special tokens.

## Context

### Design Documents
- Architecture Design: `/home/ubuntu/workspace/goc/docs/architecture-design.md`
- Implementation Plan: `/home/ubuntu/workspace/goc/docs/implementation-plan.md`

### Related Components
- Lexer: Will use these tokens
- Parser: Will consume these tokens
- Error Handler: Will report token-related errors

### Current State
- Project structure is set up
- Basic CLI is working
- No token definitions exist yet

## Requirements

### Functional Requirements
1. Define all C11 keywords as token types
2. Define all C operators as token types
3. Define literal token types (numbers, strings, characters)
4. Define identifier token type
5. Define special tokens (EOF, ERROR, etc.)
6. Provide token position information (line, column, filename)
7. Provide token value storage (for literals, identifiers)

### Non-Functional Requirements
1. Code must be well-documented
2. Token types must be easily extensible
3. Token representation must be efficient
4. Must follow Go idioms and best practices

### C11 Keywords to Support
```
auto, break, case, char, const, continue, default, do, double,
else, enum, extern, float, for, goto, if, inline, int, long,
register, restrict, return, short, signed, sizeof, static,
struct, switch, typedef, union, unsigned, void, volatile,
while, _Alignas, _Alignof, _Atomic, _Bool, _Complex,
_Generic, _Imaginary, _Noreturn, _Static_assert, _Thread_local
```

### C Operators to Support
```
Arithmetic: + - * / % ++ --
Relational: == != < > <= >=
Logical: && || !
Bitwise: & | ^ ~ << >>
Assignment: = += -= *= /= %= &= |= ^= <<= >>=
Other: ? : ; , . -> [ ] ( ) { } ...
```

### Literal Types to Support
1. Integer literals (decimal, octal, hexadecimal)
2. Floating-point literals
3. Character literals
4. String literals
5. Escape sequences

## Technical Specifications

### Token Structure
```go
type Token struct {
    Type     TokenType
    Value    string
    Position Position
}

type Position struct {
    Filename string
    Line     int
    Column   int
}

type TokenType int
```

### Token Categories
1. **Keywords**: All C11 keywords
2. **Operators**: All C operators
3. **Literals**: Numbers, strings, characters
4. **Identifiers**: Variable and function names
5. **Special**: EOF, ERROR, COMMENT

### Token Type Naming Convention
- Keywords: `TOKEN_KEYWORD_<NAME>` (e.g., `TOKEN_KEYWORD_INT`)
- Operators: `TOKEN_<SYMBOL>` (e.g., `TOKEN_PLUS`, `TOKEN_MINUS`)
- Literals: `TOKEN_LITERAL_<TYPE>` (e.g., `TOKEN_LITERAL_INTEGER`)
- Identifiers: `TOKEN_IDENTIFIER`
- Special: `TOKEN_EOF`, `TOKEN_ERROR`, `TOKEN_COMMENT`

## Constraints

### Must Do
- ✅ Define all token types as specified
- ✅ Provide complete token structure
- ✅ Include position information
- ✅ Follow naming conventions
- ✅ Write comprehensive tests

### Must NOT Do
- ❌ Implement tokenization logic (that's for lexer.go)
- ❌ Make token types mutable
- ❌ Use external dependencies
- ❌ Skip any C11 keywords or operators

### Quality Standards
- Test coverage > 80%
- All exported types and functions documented
- Follow Go naming conventions
- No compiler warnings

## Deliverables

### Files to Create
1. `pkg/lexer/token.go` - Token definitions
2. `pkg/lexer/token_test.go` - Token tests

### Expected Content

#### pkg/lexer/token.go
```go
package lexer

// TokenType represents the type of a token
type TokenType int

// Position represents a position in source code
type Position struct {
    Filename string
    Line     int
    Column   int
}

// Token represents a lexical token
type Token struct {
    Type     TokenType
    Value    string
    Position Position
}

// Token type definitions
const (
    // Special tokens
    TOKEN_EOF TokenType = iota
    TOKEN_ERROR
    
    // Literals
    TOKEN_LITERAL_INTEGER
    TOKEN_LITERAL_FLOAT
    TOKEN_LITERAL_CHAR
    TOKEN_LITERAL_STRING
    
    // Identifiers
    TOKEN_IDENTIFIER
    
    // Keywords
    TOKEN_KEYWORD_AUTO
    TOKEN_KEYWORD_BREAK
    // ... (all C11 keywords)
    
    // Operators
    TOKEN_PLUS
    TOKEN_MINUS
    // ... (all C operators)
    
    // Delimiters
    TOKEN_SEMICOLON
    TOKEN_COMMA
    // ... (all delimiters)
)

// Token type utilities
func (t TokenType) String() string { /* ... */ }
func (t TokenType) IsKeyword() bool { /* ... */ }
func (t TokenType) IsOperator() bool { /* ... */ }
func (t TokenType) IsLiteral() bool { /* ... */ }

// Token utilities
func (t Token) String() string { /* ... */ }
```

#### pkg/lexer/token_test.go
```go
package lexer

import "testing"

func TestTokenTypeString(t *testing.T) {
    // Test all token types
}

func TestTokenIsKeyword(t *testing.T) {
    // Test keyword detection
}

func TestTokenIsOperator(t *testing.T) {
    // Test operator detection
}

func TestTokenString(t *testing.T) {
    // Test token string representation
}

func TestPositionString(t *testing.T) {
    // Test position string representation
}
```

## Success Criteria

### Functional Criteria
- [ ] All C11 keywords defined
- [ ] All C operators defined
- [ ] All literal types defined
- [ ] Token structure complete
- [ ] Position tracking implemented
- [ ] Token utilities implemented

### Quality Criteria
- [ ] Test coverage > 80%
- [ ] All exported items documented
- [ ] No compiler warnings
- [ ] Code follows Go conventions
- [ ] Code is well-organized

### Integration Criteria
- [ ] Can be imported by lexer.go
- [ ] Can be used by parser
- [ ] Compatible with error handler

## Estimated Effort
- **Trunk Phase**: 1-2 hours (interface design)
- **Branch Phase**: 2-3 hours (implementation and testing)
- **Total**: 3-5 hours

## Dependencies
- None (this is a foundational component)

## Next Steps After Completion
1. Lexer interface design (Trunk)
2. Lexer implementation (Branch)
3. Error handling framework (Trunk → Branch)

## Notes
- This is a critical foundation component
- Must be done correctly as it affects all later components
- Take time to ensure completeness and correctness
- Consider future extensions (C23 keywords, GCC extensions)

---

**Created**: 2025-06-17
**Author**: Zero-FAS (Root Node)
**Status**: Ready for Delegation