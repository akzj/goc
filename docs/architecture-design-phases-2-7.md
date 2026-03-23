# GOC Compiler Architecture Design - Phases 2-7

**Document Version**: 2.0  
**Created**: 2025-06-17  
**Author**: Zero-FAS (Trunk Node)  
**Status**: Architecture Complete - Ready for Implementation  

---

## 1. Overview

This document provides the complete architecture design for GOC compiler Phases 2-7, covering:
- **Phase 2**: Parser (AST nodes, grammar)
- **Phase 3**: Semantic Analyzer (type system, symbol table)
- **Phase 4**: IR Generator (intermediate representation)
- **Phase 5**: Code Generator (x86-64 assembly)
- **Phase 6**: Linker (ELF format)
- **Phase 7**: CLI Enhancement & Integration

### 1.1 Compiler Pipeline

```
Source Code (.c)
      ↓
┌─────────────┐
│   Lexer     │ → Token Stream (COMPLETED)
└─────────────┘
      ↓
┌─────────────┐
│   Parser    │ → Abstract Syntax Tree (AST)
└─────────────┘
      ↓
┌─────────────┐
│  Semantic   │ → Annotated AST + Symbol Table
│  Analyzer   │
└─────────────┘
      ↓
┌─────────────┐
│   IR        │ → Intermediate Representation (Three-Address Code)
│  Generator  │
└─────────────┘
      ↓
┌─────────────┐
│   Code      │ → x86-64 Assembly
│  Generator  │
└─────────────┘
      ↓
┌─────────────┐
│   Linker    │ → ELF64 Executable
└─────────────┘
      ↓
Executable Binary
```

### 1.2 Data Flow Between Components

```
┌──────────────────────────────────────────────────────────────────┐
│                        DATA FLOW                                  │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Tokens (lexer.Token[])                                          │
│       ↓                                                           │
│  ┌─────────────────────────────────────────────────────────┐     │
│  │ Parser.Parse() → AST (parser.Node)                       │     │
│  └─────────────────────────────────────────────────────────┘     │
│       ↓                                                           │
│  AST (parser.Node)                                                │
│       ↓                                                           │
│  ┌─────────────────────────────────────────────────────────┐     │
│  │ SemanticAnalyzer.Analyze() → AnnotatedAST + SymbolTable │     │
│  └─────────────────────────────────────────────────────────┘     │
│       ↓                                                           │
│  AnnotatedAST + SymbolTable                                       │
│       ↓                                                           │
│  ┌─────────────────────────────────────────────────────────┐     │
│  │ IRGenerator.Generate() → IR (ir.Function[])              │     │
│  └─────────────────────────────────────────────────────────┘     │
│       ↓                                                           │
│  IR (ir.Function[])                                               │
│       ↓                                                           │
│  ┌─────────────────────────────────────────────────────────┐     │
│  │ CodeGenerator.Generate() → Assembly (string)             │     │
│  └─────────────────────────────────────────────────────────┘     │
│       ↓                                                           │
│  Assembly (string)                                                │
│       ↓                                                           │
│  ┌─────────────────────────────────────────────────────────┐     │
│  │ Linker.Link() → ELF64 Binary ([]byte)                    │     │
│  └─────────────────────────────────────────────────────────┘     │
│       ↓                                                           │
│  ELF64 Binary ([]byte)                                            │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

---

## 2. Error Handling Framework (internal/error/)

### 2.1 Design Principles

- **Consistent Error Format**: All errors follow the same structure
- **Source Location Tracking**: Every error includes file, line, column
- **Error Categories**: Distinguish between errors, warnings, notes
- **Error Recovery**: Support continuing after recoverable errors
- **Diagnostic Output**: Human-readable error messages with context

### 2.2 Interface Design

```go
// Error represents a compilation error or warning.
type Error struct {
    Level     ErrorLevel    // ERROR, WARNING, NOTE
    Code      ErrorCode     // Unique error code
    Message   string        // Human-readable message
    Position  Position      // Source location
    Hint      string        // Optional hint for fixing
    Related   []RelatedInfo // Related locations (for multi-point errors)
}

// ErrorLevel indicates the severity of a diagnostic.
type ErrorLevel int

const (
    ERROR   ErrorLevel = iota // Compilation stops
    WARNING                   // Compilation continues
    NOTE                      // Informational
)

// ErrorCode uniquely identifies error types.
type ErrorCode string

// Position tracks source code location.
type Position struct {
    File   string // File path
    Line   int    // 1-based line number
    Column int    // 1-based column (in runes)
}

// ErrorHandler collects and reports errors.
type ErrorHandler struct {
    errors   []*Error
    maxErrors int    // Stop after this many errors
    source   map[string]string // Source cache for context
}
```

### 2.3 Error Codes (Planned)

| Code Range | Category | Examples |
|------------|----------|----------|
| E0001-E0999 | Lexer Errors | E0001: invalid character, E0002: unterminated string |
| E1001-E1999 | Parser Errors | E1001: syntax error, E1002: unexpected token |
| E2001-E2999 | Semantic Errors | E2001: undefined symbol, E2002: type mismatch |
| E3001-E3999 | IR Errors | E3001: invalid IR, E3002: control flow error |
| E4001-E4999 | CodeGen Errors | E4001: unsupported operation |
| E5001-E5999 | Linker Errors | E5001: undefined symbol, E5002: duplicate symbol |

---

## 3. Parser (pkg/parser/)

### 3.1 Responsibility

Convert token stream into Abstract Syntax Tree (AST) following C11 grammar.

### 3.2 Interface Design

```go
// Parser parses C11 source code into an AST.
type Parser struct {
    tokens   []lexer.Token
    pos      int              // Current token position
    errors   *error.ErrorHandler
    ast      *TranslationUnit // Root of AST
}

// Parse parses the token stream and returns the AST.
func (p *Parser) Parse() (*TranslationUnit, error)

// ParseExpression parses an expression.
func (p *Parser) ParseExpression() (Expr, error)

// ParseStatement parses a statement.
func (p *Parser) ParseStatement() (Stmt, error)

// ParseDeclaration parses a declaration.
func (p *Parser) ParseDeclaration() (Decl, error)
```

### 3.3 AST Node Hierarchy

```
Node (interface)
├── TranslationUnit
│   └── []Declaration
│
├── Declaration (interface)
│   ├── FunctionDecl
│   │   ├── Type (FuncType)
│   │   ├── Name (string)
│   │   ├── Params ([]ParamDecl)
│   │   └── Body (CompoundStmt)
│   │
│   ├── VarDecl
│   │   ├── Type (Type)
│   │   ├── Name (string)
│   │   └── Init (Expr, optional)
│   │
│   ├── TypeDecl (typedef)
│   │   ├── Name (string)
│   │   └── Type (Type)
│   │
│   └── StructDecl / UnionDecl / EnumDecl
│
├── Statement (interface)
│   ├── CompoundStmt
│   │   └── []Statement
│   │
│   ├── ExprStmt
│   │   └── Expr
│   │
│   ├── ReturnStmt
│   │   └── Expr (optional)
│   │
│   ├── IfStmt
│   │   ├── Cond (Expr)
│   │   ├── Then (Statement)
│   │   └── Else (Statement, optional)
│   │
│   ├── WhileStmt
│   │   ├── Cond (Expr)
│   │   └── Body (Statement)
│   │
│   ├── DoWhileStmt
│   │   ├── Body (Statement)
│   │   └── Cond (Expr)
│   │
│   ├── ForStmt
│   │   ├── Init (Expr, optional)
│   │   ├── Cond (Expr, optional)
│   │   ├── Update (Expr, optional)
│   │   └── Body (Statement)
│   │
│   ├── BreakStmt
│   ├── ContinueStmt
│   ├── GotoStmt
│   │   └── Label (string)
│   │
│   └── LabelStmt
│       └── Label (string)
│
└── Expression (interface)
    ├── BinaryExpr
    │   ├── Op (TokenType)
    │   ├── Left (Expr)
    │   └── Right (Expr)
    │
    ├── UnaryExpr
    │   ├── Op (TokenType)
    │   └── Operand (Expr)
    │
    ├── CallExpr
    │   ├── Func (Expr)
    │   └── Args ([]Expr)
    │
    ├── MemberExpr
    │   ├── Object (Expr)
    │   ├── Field (string)
    │   └── IsPointer (bool)
    │
    ├── IndexExpr
    │   ├── Array (Expr)
    │   └── Index (Expr)
    │
    ├── CondExpr (ternary)
    │   ├── Cond (Expr)
    │   ├── True (Expr)
    │   └── False (Expr)
    │
    ├── CastExpr
    │   ├── Type (Type)
    │   └── Expr (Expr)
    │
    ├── SizeofExpr
    │   └── Type or Expr
    │
    ├── AssignExpr
    │   ├── Op (TokenType)
    │   ├── Left (Expr)
    │   └── Right (Expr)
    │
    ├── IdentExpr
    │   └── Name (string)
    │
    ├── IntLiteral
    │   └── Value (int64)
    │
    ├── FloatLiteral
    │   └── Value (float64)
    │
    ├── CharLiteral
    │   └── Value (rune)
    │
    └── StringLiteral
        └── Value (string)
```

### 3.4 Type Hierarchy

```go
// Type represents a C type.
type Type interface {
    TypeKind() TypeKind
    String() string
    Size() int64  // In bytes, -1 if incomplete
    Align() int64 // Alignment requirement
}

type TypeKind int

const (
    TypeVoid TypeKind = iota
    TypeBool
    TypeChar
    TypeShort
    TypeInt
    TypeLong
    TypeFloat
    TypeDouble
    TypePointer
    TypeArray
    TypeFunction
    TypeStruct
    TypeUnion
    TypeEnum
    TypeTypedef
    TypeQualifed // const, volatile, restrict, _Atomic
)

// PointerType represents T*.
type PointerType struct {
    Elem Type
}

// ArrayType represents T[N].
type ArrayType struct {
    Elem  Type
    Size  int64 // -1 for incomplete arrays
}

// FuncType represents function type.
type FuncType struct {
    Return   Type
    Params   []Type
    Variadic bool
}

// StructType represents struct/union.
type StructType struct {
    Name   string
    Fields []*FieldDecl
    IsUnion bool
}
```

### 3.5 Grammar Reference (EBNF Subset)

```ebnf
TranslationUnit = { Declaration } .
Declaration     = FunctionDecl | VarDecl | TypeDecl | StructDecl .
FunctionDecl    = TypeSpec Identifier "(" ParamList? ")" CompoundStmt .
ParamList       = ParamDecl { "," ParamDecl } .
ParamDecl       = TypeSpec Identifier .
CompoundStmt    = "{" { Declaration | Statement } "}" .
Statement       = CompoundStmt
                | IfStmt | WhileStmt | DoWhileStmt | ForStmt
                | ReturnStmt | BreakStmt | ContinueStmt | GotoStmt
                | ExprStmt .
IfStmt          = "if" "(" Expression ")" Statement ( "else" Statement )? .
WhileStmt       = "while" "(" Expression ")" Statement .
ForStmt         = "for" "(" ExprStmt? Expression? ";" Expression? ")" Statement .
ReturnStmt      = "return" Expression? ";" .
ExprStmt        = Expression? ";" .
Expression      = AssignmentExpr .
AssignmentExpr  = ConditionalExpr ( AssignOp ConditionalExpr )* .
ConditionalExpr = LogicalOrExpr ( "?" Expression ":" ConditionalExpr )? .
LogicalOrExpr   = LogicalAndExpr { "||" LogicalAndExpr } .
... (operator precedence continues)
PrimaryExpr     = Identifier | Literal | "(" Expression ")" .
```

---

## 4. Semantic Analyzer (pkg/semantic/)

### 4.1 Responsibility

Perform semantic analysis on AST:
- Type checking
- Symbol resolution
- Scope management
- Declaration validation
- Type inference

### 4.2 Interface Design

```go
// SemanticAnalyzer performs semantic analysis on AST.
type SemanticAnalyzer struct {
    symbolTable *SymbolTable
    errors      *error.ErrorHandler
    currentScope *Scope
}

// Analyze performs semantic analysis on the AST.
func (a *SemanticAnalyzer) Analyze(ast *parser.TranslationUnit) (*AnnotatedAST, error)

// EnterScope creates a new scope.
func (a *SemanticAnalyzer) EnterScope()

// ExitScope closes the current scope.
func (a *SemanticAnalyzer) ExitScope()

// Lookup looks up a symbol in the current scope chain.
func (a *SemanticAnalyzer) Lookup(name string) *Symbol

// Declare declares a symbol in the current scope.
func (a *SemanticAnalyzer) Declare(symbol *Symbol) error
```

### 4.3 Symbol Table Design

```
SymbolTable
├── globalScope
│   ├── symbols: map[string]*Symbol
│   └── parent: nil
│
├── functionScope (for each function)
│   ├── symbols: map[string]*Symbol
│   ├── parent: globalScope
│   └── children: [blockScope1, blockScope2, ...]
│
└── blockScope (for each block)
    ├── symbols: map[string]*Symbol
    ├── parent: functionScope or outer blockScope
    └── children: [innerBlockScope, ...]
```

```go
// SymbolTable manages symbols across all scopes.
type SymbolTable struct {
    globalScope *Scope
    currentScope *Scope
    scopes      []*Scope // Stack for scope traversal
}

// Scope represents a lexical scope.
type Scope struct {
    name     string
    symbols  map[string]*Symbol
    parent   *Scope
    children []*Scope
    level    int // Nesting level
}

// Symbol represents a declared symbol.
type Symbol struct {
    Name     string
    Kind     SymbolKind
    Type     types.Type
    Position lexer.Position
    Flags    SymbolFlags
    
    // Kind-specific data
    FuncInfo    *FunctionInfo
    VarInfo     *VariableInfo
    TypeDefInfo *TypeDefInfo
    StructInfo  *StructInfo
}

type SymbolKind int

const (
    SymbolFunction SymbolKind = iota
    SymbolVariable
    SymbolParameter
    SymbolTypedef
    SymbolStruct
    SymbolUnion
    SymbolEnum
    SymbolEnumConstant
    SymbolLabel
)

type SymbolFlags int

const (
    FlagNone     SymbolFlags = 0
    FlagConst    SymbolFlags = 1 << iota
    FlagVolatile
    FlagStatic
    FlagExtern
    FlagInline
    FlagThreadLocal
)
```

### 4.4 Type System

```go
// TypeChecker performs type checking.
type TypeChecker struct {
    analyzer *SemanticAnalyzer
    errors   *error.ErrorHandler
}

// CheckAssignable checks if srcType can be assigned to dstType.
func (tc *TypeChecker) CheckAssignable(dstType, srcType types.Type, pos lexer.Position) error

// CheckBinaryOp checks if binary operation is valid.
func (tc *TypeChecker) CheckBinaryOp(op lexer.TokenType, left, right types.Type, pos lexer.Position) (types.Type, error)

// CheckUnaryOp checks if unary operation is valid.
func (tc *TypeChecker) CheckUnaryOp(op lexer.TokenType, operand types.Type, pos lexer.Position) (types.Type, error)

// CheckCall checks if function call is valid.
func (tc *TypeChecker) CheckCall(funcType *types.FuncType, args []types.Type, pos lexer.Position) (types.Type, error)

// ImplicitCast performs implicit type conversion.
func (tc *TypeChecker) ImplicitCast(expr parser.Expr, from, to types.Type) parser.Expr
```

### 4.5 Type Promotion Rules (C11)

```
Integer Promotion:
- char, short → int
- unsigned char, unsigned short → unsigned int (if int can hold all values, else unsigned int)

Usual Arithmetic Conversions:
1. If either is long double, convert other to long double
2. Else if either is double, convert other to double
3. Else if either is float, convert other to float
4. Else (integer types):
   - Apply integer promotions
   - If both same signedness, convert to wider type
   - If unsigned rank >= signed rank, convert signed to unsigned
   - If signed can hold all unsigned values, convert unsigned to signed
   - Else convert both to unsigned version of signed type
```

---

## 5. IR Generator (pkg/ir/)

### 5.1 Responsibility

Generate intermediate representation (three-address code) from annotated AST.

### 5.2 IR Design (Three-Address Code)

```go
// IR represents the intermediate representation.
type IR struct {
    Functions []*Function
    Globals   []*GlobalVar
    Constants []*Constant
}

// Function represents a function in IR.
type Function struct {
    Name       string
    ReturnType types.Type
    Params     []*Param
    Blocks     []*BasicBlock
    LocalVars  []*LocalVar
}

// BasicBlock represents a basic block in CFG.
type BasicBlock struct {
    Label      string
    Instrs     []Instruction
    Preds      []*BasicBlock  // Predecessors
    Succs      []*BasicBlock  // Successors
}

// Instruction represents a three-address instruction.
type Instruction interface {
    Opcode() Opcode
    Dest() *Operand
    Operands() []*Operand
    String() string
}

// Operand represents an instruction operand.
type Operand struct {
    Kind  OperandKind
    Type  types.Type
    Value interface{} // *Temp, *Param, *Global, constant value
}

type OperandKind int

const (
    OperandTemp OperandKind = iota
    OperandParam
    OperandGlobal
    OperandConst
    OperandLabel
)
```

### 5.3 Instruction Set

```go
type Opcode int

const (
    // Arithmetic
    OpAdd Opcode = iota
    OpSub
    OpMul
    OpDiv
    OpMod
    OpNeg
    OpBitNot
    OpBitAnd
    OpBitOr
    OpBitXor
    OpShl
    OpShr
    
    // Comparison
    OpEq
    OpNe
    OpLt
    OpLe
    OpGt
    OpGe
    
    // Logical
    OpAnd
    OpOr
    OpNot
    
    // Memory
    OpLoad
    OpStore
    OpLea  // Load Effective Address
    OpAlloc
    OpFree
    
    // Control Flow
    OpJmp
    OpJmpIf
    OpJmpUnless
    OpCall
    OpRet
    OpLabel
    
    // Conversion
    OpCast
    OpZeroExt
    OpSignExt
    OpTrunc
    
    // Special
    OpPhi  // For SSA (optional)
    OpNop
)
```

### 5.4 Interface Design

```go
// IRGenerator generates IR from annotated AST.
type IRGenerator struct {
    errors     *error.ErrorHandler
    ir         *IR
    tempCounter int
    labelCounter int
    currentFunc *Function
    currentBlock *BasicBlock
}

// Generate generates IR from the AST.
func (g *IRGenerator) Generate(ast *parser.TranslationUnit) (*IR, error)

// NewTemp creates a new temporary variable.
func (g *IRGenerator) NewTemp(t types.Type) *Operand

// NewLabel creates a new label.
func (g *IRGenerator) NewLabel() string

// Emit emits an instruction to the current block.
func (g *IRGenerator) Emit(instr Instruction)
```

### 5.5 Control Flow Graph

```
For "if (cond) { A } else { B }":

    ┌─────────────┐
    │ cond_block  │
    │ t1 = cond   │
    │ jmp_unless t1 → else_block
    └──────┬──────┘
           │
           ↓
    ┌─────────────┐
    │ then_block  │
    │ A           │
    │ jmp → end_block
    └──────┬──────┘
           │
           ↓
    ┌─────────────┐     ┌─────────────┐
    │ else_block  │────→│ end_block   │
    │ B           │     │             │
    │ jmp → end_block   │             │
    └─────────────┘     └─────────────┘
```

---

## 6. Code Generator (pkg/codegen/)

### 6.1 Responsibility

Generate x86-64 assembly code from IR.

### 6.2 Target Architecture (x86-64 System V AMD64 ABI)

**Calling Convention:**
- Arguments: RDI, RSI, RDX, RCX, R8, R9 (first 6 integer args)
- Return: RAX (integer), XMM0 (float)
- Callee-saved: RBX, RBP, R12-R15
- Caller-saved: RAX, RCX, RDX, RSI, RDI, R8-R11

**Stack Alignment:**
- Stack must be 16-byte aligned before CALL
- RBP points to previous RBP (frame pointer)
- RSP points to top of stack

### 6.3 Interface Design

```go
// CodeGenerator generates x86-64 assembly from IR.
type CodeGenerator struct {
    ir         *ir.IR
    errors     *error.ErrorHandler
    output     *strings.Builder
    regAlloc   *RegisterAllocator
    stackFrame *StackFrame
}

// Generate generates assembly code.
func (cg *CodeGenerator) Generate(ir *ir.IR) (string, error)

// GenerateFunction generates assembly for a single function.
func (cg *CodeGenerator) GenerateFunction(fn *ir.Function) string
```

### 6.4 Register Allocation

```go
// RegisterAllocator manages register allocation.
type RegisterAllocator struct {
    available map[Reg]bool
    spilled   map[*ir.Operand]int // Spilled temps → stack offset
    current   map[*ir.Operand]Reg // Current register assignments
}

type Reg int

const (
    RAX Reg = iota
    RBX
    RCX
    RDX
    RSI
    RDI
    RBP
    RSP
    R8
    R9
    R10
    R11
    R12
    R13
    R14
    R15
)
```

### 6.5 Assembly Output Format

```asm
    .file   "source.c"
    .text
    .globl  main
    .type   main, @function
main:
    .cfi_startproc
    pushq   %rbp
    .cfi_def_cfa_offset 16
    .cfi_offset 6, -16
    movq    %rsp, %rbp
    .cfi_def_cfa_register 6
    subq    $16, %rsp
    
    # Function body
    movl    $42, -4(%rbp)
    movl    -4(%rbp), %eax
    
    leave
    .cfi_def_cfa 7, 8
    ret
    .cfi_endproc
    .size   main, .-main
    
    .section    .rodata
.LC0:
    .string "Hello, World!"
```

---

## 7. Linker (pkg/linker/)

### 7.1 Responsibility

Link object files into ELF64 executable.

### 7.2 ELF64 Format

```
ELF Header
Program Header Table (segments)
    .text segment (CODE)
    .data segment (DATA)
    .rodata segment (DATA)
Section Headers
    .text
    .data
    .bss
    .rodata
    .symtab
    .strtab
    .shstrtab
```

### 7.3 Interface Design

```go
// Linker links object files into an executable.
type Linker struct {
    errors   *error.ErrorHandler
    symbols  map[string]*Symbol
    sections []*Section
}

// Link links the given object files and libraries.
func (l *Linker) Link(objects []ObjectFile, libs []string) ([]byte, error)

// ResolveSymbols resolves undefined symbols.
func (l *Linker) ResolveSymbols() error

// Relocate performs relocations.
func (l *Linker) Relocate() error

// Emit emits the final ELF binary.
func (l *Linker) Emit() ([]byte, error)
```

### 7.4 Symbol Resolution

```go
// Symbol represents a symbol in the linker.
type Symbol struct {
    Name     string
    Value    uint64
    Size     uint64
    Section  *Section
    Binding  SymbolBinding
    Type     SymbolType
    Defined  bool
}

type SymbolBinding int

const (
    STB_LOCAL  SymbolBinding = iota
    STB_GLOBAL
    STB_WEAK
)

type SymbolType int

const (
    STT_NOTYPE SymbolType = iota
    STT_OBJECT
    STT_FUNC
    STT_SECTION
    STT_FILE
)
```

---

## 8. CLI (pkg/cli/)

### 8.1 Responsibility

Enhanced command-line interface for the compiler.

### 8.2 Commands

```bash
goc compile <source.c> [options]   # Compile to executable
goc tokenize <source.c> [options]  # Tokenize (debug)
goc parse <source.c> [options]     # Parse and print AST (debug)
goc semantic <source.c> [options]  # Semantic analysis (debug)
goc ir <source.c> [options]        # Generate and print IR (debug)
goc asm <source.c> [options]       # Generate assembly (debug)
goc version                        # Show version
goc help                           # Show help
```

### 8.3 Compile Options

```bash
-o <output>        # Output file (default: a.out)
-S                 # Output assembly only
-c                 # Compile to object file only
-E                 # Preprocess only
-I <dir>           # Add include directory
-D <macro>         # Define macro
-O0|-O1|-O2|-O3    # Optimization level
-g                 # Generate debug info
-v                 # Verbose output
--target <arch>    # Target architecture (default: x86-64)
```

### 8.4 Interface Design

```go
// CLI represents the command-line interface.
type CLI struct {
    name    string
    version string
    commands map[string]*Command
}

// Command represents a CLI command.
type Command struct {
    Name        string
    Description string
    Handler     CommandHandler
    Flags       []Flag
}

// CommandHandler handles a command.
type CommandHandler func(args []string, flags map[string]interface{}) error

// Run runs the CLI.
func (cli *CLI) Run(args []string) error
```

---

## 9. Module Boundaries

### 9.1 Import Dependencies

```
pkg/lexer      → (no internal dependencies)
pkg/parser     → pkg/lexer, internal/error
pkg/semantic   → pkg/parser, pkg/lexer, internal/error
pkg/ir         → pkg/parser, pkg/semantic, internal/error
pkg/codegen    → pkg/ir, internal/error
pkg/linker     → pkg/codegen, internal/error
pkg/cli        → all packages
internal/error → (no internal dependencies)
```

### 9.2 Data Ownership

| Component | Owns | Consumes | Produces |
|-----------|------|----------|----------|
| Lexer | Source code | - | Tokens |
| Parser | Tokens | Tokens | AST |
| Semantic | AST | AST | Annotated AST + Symbol Table |
| IR Gen | Annotated AST | Annotated AST | IR |
| CodeGen | IR | IR | Assembly |
| Linker | Assembly + Objects | Assembly + Objects | ELF Binary |

### 9.3 Error Propagation

```
All components → internal/error.ErrorHandler
                    ↓
            Collects errors
                    ↓
            Reports to user
                    ↓
            Returns error count
                    ↓
            If errors > 0: stop compilation
```

---

## 10. Implementation Plan for Branch Agents

### Phase 2: Parser (Week 3-4)

| Task ID | Description | Files | Complexity |
|---------|-------------|-------|------------|
| P2.1 | AST node definitions | pkg/parser/ast.go | M |
| P2.2 | Type system definitions | pkg/parser/type.go | M |
| P2.3 | Parser interface and skeleton | pkg/parser/parser.go | M |
| P2.4 | Expression parsing | pkg/parser/expr.go | L |
| P2.5 | Statement parsing | pkg/parser/stmt.go | L |
| P2.6 | Declaration parsing | pkg/parser/decl.go | L |
| P2.7 | Grammar reference | pkg/parser/grammar.go | S |
| P2.8 | Parser tests | pkg/parser/parser_test.go | M |

### Phase 3: Semantic Analyzer (Week 5-6)

| Task ID | Description | Files | Complexity |
|---------|-------------|-------|------------|
| P3.1 | Symbol table implementation | pkg/semantic/symbol.go | M |
| P3.2 | Type system implementation | pkg/semantic/type.go | M |
| P3.3 | Semantic analyzer interface | pkg/semantic/analyzer.go | M |
| P3.4 | Type checking | pkg/semantic/check.go | L |
| P3.5 | Scope management | pkg/semantic/scope.go | M |
| P3.6 | Semantic tests | pkg/semantic/analyzer_test.go | M |

### Phase 4: IR Generator (Week 7-8)

| Task ID | Description | Files | Complexity |
|---------|-------------|-------|------------|
| P4.1 | IR definitions | pkg/ir/ir.go | M |
| P4.2 | Instruction set | pkg/ir/instr.go | M |
| P4.3 | IR generator interface | pkg/ir/generator.go | M |
| P4.4 | Expression lowering | pkg/ir/expr.go | L |
| P4.5 | Control flow generation | pkg/ir/cfg.go | L |
| P4.6 | IR tests | pkg/ir/generator_test.go | M |

### Phase 5: Code Generator (Week 9-10)

| Task ID | Description | Files | Complexity |
|---------|-------------|-------|------------|
| P5.1 | x86-64 definitions | pkg/codegen/x86_64.go | M |
| P5.2 | Register allocator | pkg/codegen/regalloc.go | L |
| P5.3 | Code generator interface | pkg/codegen/generator.go | M |
| P5.4 | Instruction selection | pkg/codegen/select.go | L |
| P5.5 | Assembly emission | pkg/codegen/emit.go | M |
| P5.6 | Codegen tests | pkg/codegen/generator_test.go | M |

### Phase 6: Linker (Week 11-12)

| Task ID | Description | Files | Complexity |
|---------|-------------|-------|------------|
| P6.1 | ELF64 definitions | pkg/linker/elf.go | M |
| P6.2 | Symbol table | pkg/linker/symbol.go | M |
| P6.3 | Linker interface | pkg/linker/linker.go | M |
| P6.4 | Symbol resolution | pkg/linker/resolve.go | L |
| P6.5 | Relocation | pkg/linker/reloc.go | L |
| P6.6 | Linker tests | pkg/linker/linker_test.go | M |

### Phase 7: CLI & Integration (Week 13-14)

| Task ID | Description | Files | Complexity |
|---------|-------------|-------|------------|
| P7.1 | CLI framework | pkg/cli/cli.go | M |
| P7.2 | Compile command | pkg/cli/compile.go | L |
| P7.3 | Debug commands | pkg/cli/debug.go | M |
| P7.4 | Main entry point | cmd/goc/main.go | S |
| P7.5 | Integration tests | tests/integration/ | L |
| P7.6 | Example programs | tests/examples/ | S |

### Error Handling Framework

| Task ID | Description | Files | Complexity |
|---------|-------------|-------|------------|
| E1 | Error types | internal/error/error.go | S |
| E2 | Position tracking | internal/error/position.go | S |
| E3 | Error handler | internal/error/handler.go | M |
| E4 | Error codes | internal/error/codes.go | S |
| E5 | Error tests | internal/error/error_test.go | S |

---

## 11. Testing Strategy

### 11.1 Unit Tests

Each package must have comprehensive unit tests:
- **Lexer**: Token recognition, edge cases
- **Parser**: Grammar rules, error recovery
- **Semantic**: Type checking, symbol resolution
- **IR**: Instruction generation, CFG
- **CodeGen**: Assembly output, register allocation
- **Linker**: Symbol resolution, relocation

### 11.2 Integration Tests

```
tests/integration/
├── simple/          # Simple programs
│   ├── hello.c
│   ├── arithmetic.c
│   └── control.c
├── functions/       # Function tests
│   ├── call.c
│   ├── recursive.c
│   └── variadic.c
├── types/           # Type tests
│   ├── pointers.c
│   ├── arrays.c
│   └── structs.c
└── advanced/        # Advanced features
    ├── preprocessor.c
    └── bitfields.c
```

### 11.3 Test Coverage Target

- **Line Coverage**: ≥ 80%
- **Branch Coverage**: ≥ 70%
- **Critical Paths**: 100%

---

## 12. Performance Considerations

### 12.1 Memory Management

- **Lexer**: Stream tokens, don't buffer entire token list
- **Parser**: Build AST incrementally
- **Semantic**: Reuse symbol table structures
- **IR**: Use arena allocation for instructions
- **CodeGen**: Buffer assembly output

### 12.2 Compilation Speed Targets

| Program Size | Target Time |
|--------------|-------------|
| < 100 lines | < 100ms |
| < 1000 lines | < 500ms |
| < 10000 lines | < 5s |

---

## 13. Security Considerations

### 13.1 Input Validation

- Validate all source input
- Limit maximum file size
- Limit maximum identifier length
- Limit maximum nesting depth

### 13.2 Memory Safety

- No buffer overflows in string handling
- Validate array bounds in internal data structures
- Use safe integer arithmetic

---

## 14. Future Extensions

### 14.1 Planned Features

- **Preprocessor**: #define, #include, #ifdef
- **Optimization**: Constant folding, dead code elimination
- **Debug Info**: DWARF format for gdb
- **Multiple Targets**: ARM64, RISC-V

### 14.2 C Standard Extensions

- **GCC Extensions**: Statement expressions, typeof
- **C23 Features**: When C23 is more widely adopted

---

## 15. Appendix

### 15.1 C11 Grammar Reference

See: ISO/IEC 9899:2011 specification

### 15.2 x86-64 Resources

- System V AMD64 ABI Specification
- Intel 64 and IA-32 Architectures Software Developer's Manual

### 15.3 ELF64 Resources

- ELF Specification: https://refspecs.linuxfoundation.org/elf/elf.pdf

---

**End of Architecture Design Document**