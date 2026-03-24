// Package cli provides the command-line interface for the GOC compiler.
// This file implements the compile command.
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/codegen"
	"github.com/akzj/goc/pkg/ir"
	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/linker"
	"github.com/akzj/goc/pkg/parser"
	"github.com/akzj/goc/pkg/semantic"
)

// CompileOptions represents options for the compile command.
type CompileOptions struct {
	// InputFile is the input C source file.
	InputFile string
	// Output is the output file path.
	Output string
	// EmitAssembly is true to output assembly only.
	EmitAssembly bool
	// EmitObject is true to output object file only.
	EmitObject bool
	// Verbose is true for verbose output.
	Verbose bool
}

// CompileCommand handles the compile command.
// It executes the full compilation pipeline: Lexer → Parser → Semantic → IR → CodeGen.
func CompileCommand(args []string) error {
	// Parse command-line flags
	opts, err := ParseCompileFlags(args)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	// Validate options
	if opts.InputFile == "" {
		return fmt.Errorf("no input file specified")
	}

	// Read the source file
	source, err := os.ReadFile(opts.InputFile)
	if err != nil {
		return fmt.Errorf("reading file '%s': %w", opts.InputFile, err)
	}

	sourceStr := string(source)

	// Create error handler
	errorHandler := errhand.NewErrorHandler()
	errorHandler.CacheSource(opts.InputFile, sourceStr)

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[compile] Processing file: %s\n", opts.InputFile)
	}

	// Stage 1: Lexical Analysis
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[compile] Stage 1/5: Lexical analysis\n")
	}
	l := lexer.NewLexer(sourceStr, opts.InputFile)
	tokens := l.Tokenize()

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[compile]   Generated %d tokens\n", len(tokens))
	}

	// Stage 2: Parsing
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[compile] Stage 2/5: Parsing\n")
	}
	p := parser.NewParser(tokens, errorHandler)
	ast, err := p.Parse()
	if err != nil {
		errorHandler.Error(errhand.ErrSyntaxError, err.Error(), errhand.Position{
			File: opts.InputFile,
			Line: 1,
		})
	}

	if errorHandler.HasErrors() {
		fmt.Fprintf(os.Stderr, "Compilation failed during parsing:\n")
		errorHandler.Report()
		return fmt.Errorf("parsing failed with %d error(s)", errorHandler.ErrorCount())
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[compile]   AST generated successfully\n")
	}

	// Stage 3: Semantic Analysis
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[compile] Stage 3/5: Semantic analysis\n")
	}
	sem := semantic.NewSemanticAnalyzer(errorHandler)
	if err := sem.Analyze(ast); err != nil {
		errorHandler.Error(errhand.ErrTypeMismatch, err.Error(), errhand.Position{
			File: opts.InputFile,
			Line: 1,
		})
	}

	if errorHandler.HasErrors() {
		fmt.Fprintf(os.Stderr, "Compilation failed during semantic analysis:\n")
		errorHandler.Report()
		return fmt.Errorf("semantic analysis failed with %d error(s)", errorHandler.ErrorCount())
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[compile]   Semantic analysis completed successfully\n")
	}

	// Stage 4: IR Generation
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[compile] Stage 4/5: IR generation\n")
	}
	irGen := ir.NewIRGenerator(errorHandler)
	irResult, err := irGen.Generate(ast)
	if err != nil {
		errorHandler.Error(errhand.ErrInvalidIR, err.Error(), errhand.Position{
			File: opts.InputFile,
			Line: 1,
		})
	}

	if errorHandler.HasErrors() {
		fmt.Fprintf(os.Stderr, "Compilation failed during IR generation:\n")
		errorHandler.Report()
		return fmt.Errorf("IR generation failed with %d error(s)", errorHandler.ErrorCount())
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[compile]   IR generated: %d functions, %d globals\n",
			len(irResult.Functions), len(irResult.Globals))
	}

	// Stage 5: Code Generation
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[compile] Stage 5/5: Code generation\n")
	}
	codeGen := codegen.NewCodeGenerator(errorHandler)
	assembly, err := codeGen.Generate(irResult)
	if err != nil {
		errorHandler.Error(errhand.ErrUnsupportedOp, err.Error(), errhand.Position{
			File: opts.InputFile,
			Line: 1,
		})
	}

	if errorHandler.HasErrors() {
		fmt.Fprintf(os.Stderr, "Compilation failed during code generation:\n")
		errorHandler.Report()
		return fmt.Errorf("code generation failed with %d error(s)", errorHandler.ErrorCount())
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[compile]   Assembly generated: %d bytes\n", len(assembly))
	}

	// Output the result
	if opts.EmitAssembly {
		// Output assembly to file or stdout
		if opts.Output != "" {
			if err := os.WriteFile(opts.Output, []byte(assembly), 0644); err != nil {
				return fmt.Errorf("writing assembly to '%s': %w", opts.Output, err)
			}
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "[compile] Assembly written to: %s\n", opts.Output)
			}
		} else {
			fmt.Print(assembly)
		}
	} else if opts.EmitObject {
		// Output object file
		link := linker.NewLinker(errorHandler)
		outputPath := opts.Output
		if outputPath == "" {
			baseName := strings.TrimSuffix(opts.InputFile, ".c")
			outputPath = baseName + ".o"
		}
		if err := link.CompileToObject(assembly, outputPath); err != nil {
			return fmt.Errorf("compiling to object: %w", err)
		}
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "[compile] Object file written to: %s\n", outputPath)
		}
	} else {
		// Default: link to executable
		link := linker.NewLinker(errorHandler)
		outputPath := opts.Output
		if outputPath == "" {
			baseName := strings.TrimSuffix(opts.InputFile, ".c")
			outputPath = baseName + ".exe"
		}
		if err := link.LinkAssembly(assembly, outputPath); err != nil {
			return fmt.Errorf("linking: %w", err)
		}
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "[compile] Executable written to: %s\n", outputPath)
		}
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[compile] Compilation completed successfully\n")
	}

	return nil
}

// ParseCompileFlags parses compile command flags.
// Supported flags:
//   -o <file>  : Output file path
//   -S         : Output assembly only (don't assemble)
//   -c         : Output object file only (don't link)
//   -v         : Verbose output
func ParseCompileFlags(args []string) (*CompileOptions, error) {
	opts := &CompileOptions{
		EmitAssembly: false,
		EmitObject:   false,
		Verbose:      false,
	}

	i := 0
	for i < len(args) {
		arg := args[i]

		if arg == "-o" {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("-o requires an argument")
			}
			i++
			opts.Output = args[i]
		} else if arg == "-S" {
			opts.EmitAssembly = true
		} else if arg == "-c" {
			opts.EmitObject = true
		} else if arg == "-v" || arg == "--verbose" {
			opts.Verbose = true
		} else if strings.HasPrefix(arg, "-") {
			return nil, fmt.Errorf("unknown flag: %s", arg)
		} else {
			// Positional argument (input file)
			if opts.InputFile == "" {
				opts.InputFile = arg
			} else {
				return nil, fmt.Errorf("unexpected argument: %s", arg)
			}
		}
		i++
	}

	return opts, nil
}