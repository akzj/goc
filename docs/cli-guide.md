# GOC CLI User Guide

Complete reference for the GOC compiler command-line interface.

## Quick Start

```bash
# Compile a C file
goc compile hello.c -o hello

# Compile with optimization
goc compile main.c -O2 -o main

# View help
goc help
goc help compile
```

## Commands Overview

| Command | Description | Usage |
|---------|-------------|-------|
| `compile` | Compile C source to executable | `goc compile <file.c> [options]` |
| `tokenize` | Tokenize source (debug) | `goc tokenize <file.c> [options]` |
| `parse` | Parse source and print AST (debug) | `goc parse <file.c> [options]` |
| `help` | Show help information | `goc help [command]` |
| `version` | Show version info | `goc version` |

---

## Compile Command

### Usage
```bash
goc compile <source.c> [options]
```

### All Flags

| Flag | Short | Description | Default | Example |
|------|-------|-------------|---------|---------|
| `--output` | `-o` | Output file path | - | `-o myprogram` |
| `--assembly` | `-S` | Output assembly only | - | `-S` |
| `--compile-only` | `-c` | Compile to object only | - | `-c` |
| `--preprocess` | `-E` | Preprocess only | - | `-E` |
| `--verbose` | `-v` | Verbose output | - | `-v` |
| `--debug` | `-d` | Debug mode | - | `--debug` |
| `--target` | `-t` | Target architecture | `x86-64` | `-t x86_64` |
| `--optimize` | `-O` | Optimization level | `0` | `-O2` |
| `--include` | `-I` | Add include directory | - | `-I ./inc` |
| `--define` | `-D` | Define macro | - | `-D DEBUG=1` |
| `--debug-info` | `-g` | Generate debug info | - | `-g` |
| `--help` | `-h` | Show help | - | `-h` |

### Flag Examples

#### `-o, --output <file>` - Output File
```bash
goc compile hello.c -o hello
goc compile src/main.c -o bin/myapp
```

#### `-S, --assembly` - Assembly Output
```bash
goc compile hello.c -S -o hello.s
goc compile hello.c -S  # stdout
```

#### `-c, --compile-only` - Object File Only
```bash
goc compile hello.c -c -o hello.o
goc compile main.c -c -o main.o
```

#### `-v, --verbose` - Verbose Output
```bash
goc compile hello.c -v
# Shows: Stage 1/5: Lexical analysis, Stage 2/5: Parsing, etc.
```

#### `--debug` - Debug Mode
```bash
goc compile hello.c --debug
goc compile hello.c -v --debug  # Maximum detail
```

#### `-t, --target <arch>` - Target Architecture
```bash
goc compile hello.c -t x86_64   # Default
goc compile hello.c -t arm64
goc compile hello.c -t riscv64
```
**Supported:** `x86_64`, `arm64`, `x86`, `arm`, `riscv64`

#### `-O, --optimize <level>` - Optimization
```bash
goc compile hello.c -O0  # No optimization (default)
goc compile hello.c -O1  # Basic
goc compile hello.c -O2  # Standard
goc compile hello.c -O3  # Maximum
goc compile hello.c -Os  # Size optimized
goc compile hello.c -Oz  # Very small size
```
**Valid levels:** `0`, `1`, `2`, `3`, `s`, `z`

#### `-I, --include <dir>` - Include Directory
```bash
goc compile main.c -I ./include
goc compile main.c -I ./include -I /usr/local/include
```

#### `-D, --define <macro>` - Define Macro
```bash
goc compile main.c -D DEBUG=1
goc compile main.c -D RELEASE
goc compile main.c -D DEBUG=1 -D VERSION=\"1.0\"
```

#### `-g, --debug-info` - Debug Information
```bash
goc compile hello.c -g -o hello
goc compile hello.c -g -O0 -o hello_debug
```

#### `-E, --preprocess` - Preprocess Only
```bash
goc compile hello.c -E       # stdout
goc compile hello.c -E -o hello.i
```

---

## Tokenize Command

### Usage
```bash
goc tokenize <file.c> [options]
```

### Flags
| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--format` | `-f` | Output format | `default` |
| `--help` | `-h` | Show help | - |

### Output Formats
```bash
goc tokenize hello.c -f default   # Human-readable
goc tokenize hello.c -f json      # JSON array
goc tokenize hello.c -f compact   # One per line
```

---

## Parse Command

### Usage
```bash
goc parse <file.c>
```
*Note: AST output not yet implemented.*

---

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--help` | `-h` | Show help information |
| `--version` | `-v` | Show version info |

```bash
goc --version   # Output: goc version 0.1.0
goc -h          # Show main help
```

---

## Common Workflows

### Basic Compilation
```bash
goc compile hello.c -o hello
./hello
```

### Debugging
```bash
goc compile main.c -v -o main      # Verbose
goc compile main.c -g -O0 -o main  # Debug info
goc tokenize main.c -f json        # View tokens
```

### Optimized Build
```bash
goc compile main.c -O2 -o main      # Standard
goc compile main.c -O3 -o main_fast # Maximum
goc compile main.c -Os -o main_small # Size
```

### Multi-File Project
```bash
goc compile main.c -c -o main.o
goc compile utils.c -c -o utils.o
```

### Cross-Compilation
```bash
goc compile hello.c -t arm64 -o hello_arm
goc compile hello.c -t riscv64 -o hello_riscv
```

### Custom Headers & Macros
```bash
goc compile main.c -I ./include -D DEBUG=1 -o main
```

---

## Error Messages & Troubleshooting

### Common Errors

**"no input file specified"**
```bash
# Wrong: goc compile
# Correct: goc compile hello.c
```

**"Error reading file 'xxx': no such file or directory"**
```bash
# Check: ls hello.c
# Use correct path: goc compile ./src/hello.c
```

**"unknown flag: xxx"**
```bash
# Wrong: goc compile hello.c --optimise 2
# Correct: goc compile hello.c --optimize 2
```

**"flag -o requires an argument"**
```bash
# Wrong: goc compile hello.c -o
# Correct: goc compile hello.c -o hello
```

**"invalid optimization level 'xxx'"**
```bash
# Wrong: goc compile hello.c -O 4
# Correct: goc compile hello.c -O 3
# Valid: 0, 1, 2, 3, s, z
```

**"invalid target architecture 'xxx'"**
```bash
# Wrong: goc compile hello.c -t x86_32
# Correct: goc compile hello.c -t x86_64
# Valid: x86_64, arm64, x86, arm, riscv64
```

**"Compilation failed during parsing"**
```bash
# Check syntax, use verbose mode
goc compile hello.c -v
goc tokenize hello.c  # Check tokens
```

**"Compilation failed during semantic analysis"**
```bash
# Check for type errors, undefined symbols
goc compile hello.c -v
```

### Troubleshooting Tips
1. Use `-v` to see which stage fails
2. Use `tokenize` to check lexer output
3. Start with minimal code: `int main() { return 0; }`
4. Ensure UTF-8 encoding without BOM
5. Verify C11 syntax compliance

---

## Help Examples

### Main Help
```bash
$ goc help
goc - A C compiler written in Go

Usage: goc <command> [options]

Commands:
  compile      Compile a C source file to an executable
  tokenize     Tokenize a C source file (debug)
  parse        Parse a C source file and print AST (debug)

Flags:
  -h, --help     Show help information
  -v, --version  Show version information

Run 'goc help <command>' for more information.
```

### Compile Help
```bash
$ goc help compile
Usage: goc compile <source.c> [options]

Description:
  Compile a C source file to an executable

Options:
  -o, --output     <value>   Output file
  -S, --assembly             Output assembly only
  -c, --compile-only         Compile to object file only
  -E, --preprocess           Preprocess only
  -I, --include    <value>   Add include directory
  -D, --define     <value>   Define macro
  -O, --optimize   <value>   Optimization level (0-3) (default: 0)
  -g, --debug                Generate debug info
  -v, --verbose              Verbose output
  -t, --target     <value>   Target architecture (default: x86-64)
  -h, --help                 Show help for this command

Examples:
  goc compile hello.c -o hello
  goc compile main.c -O2 -g
```

### Version
```bash
$ goc version
goc version 0.1.0
A C compiler written in Go
```

---

## Quick Reference

```
COMPILE: goc compile <file.c> [options]
  Output:  -o <file>  -S (asm)  -c (obj)  -E (preprocess)
  Debug:   -v (verbose)  --debug  -g (debug info)
  Build:   -O <0-3|s|z>  -t <arch>  -I <dir>  -D <macro>
  Help:    -h

TOKENIZE: goc tokenize <file.c> [-f default|json|compact]

GLOBAL: goc -h|--help    goc -v|--version
```

---

**GOC v0.1.0** | **C11 Standard** | **Target: x86-64 Linux ELF64**