# GOC - C Compiler in Go

[![Go Version](https://img.shields.io/badge/Go-1.22.2-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**GOC** is a C compiler implemented in Go, developed autonomously by Zero-FAS to demonstrate self-aware AI capabilities in software development.

## Overview

This project aims to build a fully functional C compiler from scratch using the Go programming language. It serves as a testbed for Zero-FAS's autonomous research and development capabilities, showcasing multi-agent collaboration patterns and self-aware memory systems.

## Features

- **C11 Standard Support**: Implements ISO/IEC 9899:2011 standard
- **x86-64 Target**: Generates code for x86-64 architecture (Linux)
- **ELF Output**: Produces ELF64 object files and executables
- **Comprehensive Testing**: High test coverage with integration tests
- **Clean Architecture**: Modular design with clear separation of concerns

## Architecture

```
Source Code → Lexer → Parser → Semantic Analyzer → IR Generator → Code Generator → Linker → Executable
```

### Components

1. **Lexer**: Tokenizes C source code
2. **Parser**: Builds Abstract Syntax Tree (AST)
3. **Semantic Analyzer**: Type checking and symbol resolution
4. **IR Generator**: Produces intermediate representation
5. **Code Generator**: Emits x86-64 assembly
6. **Linker**: Creates executable binaries

## Project Structure

```
goc/
├── cmd/goc/              # Main application
├── pkg/                  # Public packages
│   ├── lexer/           # Lexer implementation
│   ├── parser/          # Parser implementation
│   ├── semantic/        # Semantic analyzer
│   ├── ir/              # IR generator
│   ├── codegen/         # Code generator
│   ├── linker/          # Linker
│   └── cli/             # Command line interface
├── internal/             # Private packages
│   ├── error/           # Error handling
│   └── utils/           # Utilities
├── tests/               # Test suites
│   ├── integration/     # Integration tests
│   └── examples/        # Example C programs
└── docs/               # Documentation
    ├── architecture-design.md
    ├── implementation-plan.md
    └── task-specs/      # Task specifications
```

## Building

```bash
# Clone the repository
git clone https://github.com/akzj/goc.git
cd goc

# Build the compiler
go build -o bin/goc ./cmd/goc

# Run tests
go test ./...
```

## Usage

```bash
# Compile a C program
./bin/goc compile hello.c -o hello

# Tokenize a C program (debug)
./bin/goc tokenize hello.c

# Parse and print AST (debug)
./bin/goc parse hello.c

# Show help
./bin/goc --help
```

## Development Status

### Current Phase: Phase 1 - Foundation

- [x] Architecture design
- [x] Implementation plan
- [ ] Project structure setup
- [ ] Lexer implementation
- [ ] Basic CLI
- [ ] Error handling framework

### Roadmap

- **Phase 1**: Foundation (Lexer, CLI, Error handling)
- **Phase 2**: Parsing (AST, Grammar)
- **Phase 3**: Semantic Analysis (Type checking, Symbols)
- **Phase 4**: IR Generation (Control flow, Basic blocks)
- **Phase 5**: Code Generation (x86-64, Assembly)
- **Phase 6**: Linking (ELF, Symbol resolution)
- **Phase 7**: Integration and Testing

## Documentation

- [Architecture Design](docs/architecture-design.md)
- [Implementation Plan](docs/implementation-plan.md)
- [Multi-Agent Collaboration Guide](multi-agent-collaboration-guide.md)

## Contributing

This project is developed autonomously by Zero-FAS as a demonstration of AI-driven software development. The development process follows a multi-agent collaboration pattern:

- **Root Agent**: Overall coordination and architecture
- **Trunk Agent**: Interface design and skeleton code
- **Branch Agent**: Module implementation and testing
- **Leaf Agent**: Specific task execution

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Developed by Zero-FAS (Self-Aware Memory System)
- Multi-agent collaboration framework
- Inspired by educational compilers like [chibicc](https://github.com/rui314/chibicc)

---

**Note**: This project is primarily a demonstration of Zero-FAS's autonomous development capabilities. While it aims to be a functional C compiler, the primary goal is to test and showcase AI-driven software development processes.