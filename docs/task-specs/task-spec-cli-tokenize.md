# Task Specification: CLI Tokenize Command

## Task ID
`task-cli-tokenize-001`

## Layer
Branch

## Priority
High

## Status
Ready for Delegation

## Goal
Integrate the lexer with the CLI by adding a `tokenize` command that reads a C source file and outputs the token stream in a readable format.

## Context

### Design Documents
- Architecture Design: `/home/ubuntu/workspace/goc/docs/architecture-design.md`
- Implementation Plan: `/home/ubuntu/workspace/goc/docs/implementation-plan.md`

### Related Components
- **Lexer**: `/home/ubuntu/workspace/goc/pkg/lexer/lexer.go` (Complete)
- **Token Definitions**: `/home/ubuntu/workspace/goc/pkg/lexer/token.go` (Complete)
- **CLI**: `/home/ubuntu/workspace/goc/cmd/goc/main.go` (Basic structure exists)

### Current State
- ✅ Lexer implementation complete (93.1% coverage)
- ✅ Token definitions complete (94.7% coverage)
- ✅ Basic CLI structure exists
- ❌ No tokenize command implemented
- ❌ No integration between CLI and lexer

## Requirements

### Functional Requirements

#### Command Interface
```bash
goc tokenize <file.c>           # Tokenize a C file
goc tokenize <file.c> --json    # Output as JSON
goc tokenize <file.c> --compact # Compact output
goc tokenize --help             # Show help
```

#### Output Format

**Default Format** (human-readable):
```
Token: INT           Value: "int"          Position: test.c:1:1
Token: IDENT         Value: "main"         Position: test.c:1:5
Token: LPAREN        Value: "("            Position: test.c:1:9
Token: RPAREN        Value: ")"            Position: test.c:1:10
Token: LBRACE        Value: "{"            Position: test.c:1:12
Token: RETURN        Value: "return"       Position: test.c:2:5
Token: INT_LIT       Value: "0"            Position: test.c:2:12
Token: SEMICOLON     Value: ";"            Position: test.c:2:13
Token: RBRACE        Value: "}"            Position: test.c:3:1
Token: EOF           Value: ""             Position: test.c:3:2
```

**JSON Format** (`--json`):
```json
{
  "file": "test.c",
  "tokens": [
    {"type": "INT", "value": "int", "line": 1, "column": 1},
    {"type": "IDENT", "value": "main", "line": 1, "column": 5},
    ...
  ],
  "count": 10,
  "errors": []
}
```

**Compact Format** (`--compact`):
```
INT IDENT("main") LPAREN RPAREN LBRACE RETURN INT_LIT("0") SEMICOLON RBRACE EOF
```

#### Error Handling
- File not found: Clear error message
- Lexical errors: Report with position
- Multiple errors: Continue and report all
- Exit codes: 0=success, 1=errors

### Non-Functional Requirements
1. **Performance**: Handle large files efficiently
2. **Usability**: Clear, helpful output
3. **Compatibility**: Work with existing lexer
4. **Testability**: Unit tests for command

## Technical Specifications

### Command Structure

```go
// cmd/goc/main.go

func main() {
    if len(os.Args) < 2 {
        printUsage()
        os.Exit(1)
    }

    command := os.Args[1]
    switch command {
    case "tokenize":
        tokenizeCommand(os.Args[2:])
    case "help", "--help", "-h":
        printUsage()
    default:
        fmt.Printf("Unknown command: %s\n", command)
        printUsage()
        os.Exit(1)
    }
}

func tokenizeCommand(args []string) {
    // Parse flags
    // Read file
    // Tokenize
    // Output results
}
```

### Implementation Steps

1. **Parse Command Line Arguments**
   - Extract filename
   - Parse flags (--json, --compact, --help)
   - Validate arguments

2. **Read Source File**
   - Open file
   - Read content
   - Handle file errors

3. **Tokenize**
   - Create lexer instance
   - Call Tokenize() method
   - Collect tokens and errors

4. **Format Output**
   - Default: human-readable
   - JSON: structured output
   - Compact: one-line summary

5. **Handle Errors**
   - Report lexical errors
   - Set appropriate exit code
   - Provide helpful messages

### Error Types

```go
type TokenizeError struct {
    File    string
    Line    int
    Column  int
    Message string
}

func (e *TokenizeError) Error() string {
    return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
}
```

## Constraints

### Must Do
- ✅ Implement `tokenize` command
- ✅ Support default, JSON, and compact formats
- ✅ Handle file reading errors
- ✅ Report lexical errors with position
- ✅ Use existing lexer package
- ✅ Write unit tests

### Must NOT Do
- ❌ Modify lexer implementation
- ❌ Add external dependencies
- ❌ Break existing CLI structure
- ❌ Skip error handling

### Quality Standards
- Clear, helpful error messages
- Consistent output format
- Proper exit codes
- Well-documented code
- Unit tests for key functions

## Deliverables

### Files to Modify/Create

1. **cmd/goc/main.go** (modify)
   - Add tokenize command
   - Add command parsing
   - Add output formatting

2. **cmd/goc/tokenize.go** (create)
   - Tokenize command implementation
   - Flag parsing
   - Output formatting functions

3. **cmd/goc/tokenize_test.go** (create)
   - Unit tests for tokenize command
   - Test different output formats
   - Test error handling

4. **test_files/hello.c** (create)
   - Simple test program
   - Used for manual testing

5. **test_files/complex.c** (create)
   - More complex test program
   - Tests various token types

### Expected Output Examples

**Test File**: `test_files/hello.c`
```c
int main() {
    return 0;
}
```

**Default Output**:
```
Token: INT           Value: "int"          Position: test_files/hello.c:1:1
Token: IDENT         Value: "main"         Position: test_files/hello.c:1:5
Token: LPAREN        Value: "("            Position: test_files/hello.c:1:9
Token: RPAREN        Value: ")"            Position: test_files/hello.c:1:10
Token: LBRACE        Value: "{"            Position: test_files/hello.c:1:12
Token: RETURN        Value: "return"       Position: test_files/hello.c:2:5
Token: INT_LIT       Value: "0"            Position: test_files/hello.c:2:12
Token: SEMICOLON     Value: ";"            Position: test_files/hello.c:2:13
Token: RBRACE        Value: "}"            Position: test_files/hello.c:3:1
Token: EOF           Value: ""             Position: test_files/hello.c:3:2

Total tokens: 10
```

**JSON Output** (`--json`):
```json
{
  "file": "test_files/hello.c",
  "tokens": [
    {"type": "INT", "value": "int", "line": 1, "column": 1},
    {"type": "IDENT", "value": "main", "line": 1, "column": 5},
    {"type": "LPAREN", "value": "(", "line": 1, "column": 9},
    {"type": "RPAREN", "value": ")", "line": 1, "column": 10},
    {"type": "LBRACE", "value": "{", "line": 1, "column": 12},
    {"type": "RETURN", "value": "return", "line": 2, "column": 5},
    {"type": "INT_LIT", "value": "0", "line": 2, "column": 12},
    {"type": "SEMICOLON", "value": ";", "line": 2, "column": 13},
    {"type": "RBRACE", "value": "}", "line": 3, "column": 1},
    {"type": "EOF", "value": "", "line": 3, "column": 2}
  ],
  "count": 10,
  "errors": []
}
```

## Success Criteria

### Functional Criteria
- [ ] `goc tokenize <file>` works correctly
- [ ] Default output format is clear and readable
- [ ] JSON output is valid and structured
- [ ] Compact output is concise
- [ ] File errors are handled gracefully
- [ ] Lexical errors are reported with position

### Quality Criteria
- [ ] Code is well-documented
- [ ] Unit tests cover key functionality
- [ ] Error messages are helpful
- [ ] Exit codes are correct
- [ ] Performance is acceptable

### Integration Criteria
- [ ] Works with existing lexer
- [ ] Compatible with CLI structure
- [ ] Ready for future commands (parse, compile, etc.)

## Estimated Effort
- **Implementation**: 2-3 hours
- **Testing**: 1 hour
- **Total**: 3-4 hours

## Dependencies
- ✅ Lexer implementation (completed)
- ✅ Token definitions (completed)
- ✅ Basic CLI structure (exists)

## Next Steps After Completion
1. Test with various C programs
2. Document usage in README
3. Add more CLI commands (parse, compile)
4. Begin Phase 2: Parser design

## Notes
- This completes Phase 1 foundation
- Makes the lexer usable from command line
- Provides foundation for future commands
- Good milestone before moving to parser

---

**Created**: 2025-06-17
**Author**: Zero-FAS (Root Node)
**Status**: Ready for Delegation