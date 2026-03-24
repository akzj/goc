// Package cli provides the command-line interface for the GOC compiler.
// This file implements the compile command.
package cli

import (
	"fmt"
	"os"
	"strconv"
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
	// Debug is true for debug mode (extra debugging information).
	Debug bool
	// Target is the target architecture (e.g., "x86_64", "arm64").
	Target string
	// Optimize is the optimization level (e.g., "0", "1", "2", "3", "s", "z").
	Optimize string
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

	// Execute compilation pipeline
	return executeCompilationPipeline(sourceStr, opts.InputFile, opts, errorHandler)
}

// executeCompilationPipeline executes the full compilation pipeline.
func executeCompilationPipeline(sourceStr, inputFile string, opts *CompileOptions, errorHandler *errhand.ErrorHandler) error {
	printVerbose(opts, "Processing file: %s\n", inputFile)

	// Stage 1: Lexical Analysis
	tokens, err := runLexicalAnalysis(sourceStr, inputFile, opts, errorHandler)
	if err != nil {
		return err
	}

	// Stage 2: Parsing
	ast, err := runParsing(tokens, inputFile, opts, errorHandler)
	if err != nil {
		return err
	}

	// Stage 3: Semantic Analysis
	if err := runSemanticAnalysis(ast, inputFile, opts, errorHandler); err != nil {
		return err
	}

	// Stage 4: IR Generation
	irResult, err := runIRGeneration(ast, inputFile, opts, errorHandler)
	if err != nil {
		return err
	}

	// Stage 5: Code Generation
	assembly, err := runCodeGeneration(irResult, inputFile, opts, errorHandler)
	if err != nil {
		return err
	}

	// Output the result
	return outputResult(assembly, opts, errorHandler)
}

// runLexicalAnalysis performs lexical analysis on the source code.
func runLexicalAnalysis(sourceStr, inputFile string, opts *CompileOptions, errorHandler *errhand.ErrorHandler) ([]lexer.Token, error) {
	printVerbose(opts, "Stage 1/5: Lexical analysis\n")
	l := lexer.NewLexer(sourceStr, inputFile)
	tokens := l.Tokenize()
	printVerbose(opts, "  Generated %d tokens\n", len(tokens))
	return tokens, nil
}

// runParsing performs parsing on the tokens.
func runParsing(tokens []lexer.Token, inputFile string, opts *CompileOptions, errorHandler *errhand.ErrorHandler) (*parser.TranslationUnit, error) {
	printVerbose(opts, "Stage 2/5: Parsing\n")
	p := parser.NewParser(tokens, errorHandler)
	ast, err := p.Parse()
	if err != nil {
		errorHandler.Error(errhand.ErrSyntaxError, err.Error(), errhand.Position{
			File: inputFile,
			Line: 1,
		})
	}

	if errorHandler.HasErrors() {
		fmt.Fprintf(os.Stderr, "Compilation failed during parsing:\n")
		errorHandler.Report()
		return nil, fmt.Errorf("parsing failed with %d error(s)", errorHandler.ErrorCount())
	}

	printVerbose(opts, "  AST generated successfully\n")
	return ast, nil
}

// runSemanticAnalysis performs semantic analysis on the AST.
func runSemanticAnalysis(ast *parser.TranslationUnit, inputFile string, opts *CompileOptions, errorHandler *errhand.ErrorHandler) error {
	printVerbose(opts, "Stage 3/5: Semantic analysis\n")
	sem := semantic.NewSemanticAnalyzer(errorHandler)
	if err := sem.Analyze(ast); err != nil {
		errorHandler.Error(errhand.ErrTypeMismatch, err.Error(), errhand.Position{
			File: inputFile,
			Line: 1,
		})
	}

	if errorHandler.HasErrors() {
		fmt.Fprintf(os.Stderr, "Compilation failed during semantic analysis:\n")
		errorHandler.Report()
		return fmt.Errorf("semantic analysis failed with %d error(s)", errorHandler.ErrorCount())
	}

	printVerbose(opts, "  Semantic analysis completed successfully\n")
	return nil
}

// runIRGeneration generates IR from the AST.
func runIRGeneration(ast *parser.TranslationUnit, inputFile string, opts *CompileOptions, errorHandler *errhand.ErrorHandler) (*ir.IR, error) {
	printVerbose(opts, "Stage 4/5: IR generation\n")
	irGen := ir.NewIRGenerator(errorHandler)
	irResult, err := irGen.Generate(ast)
	if err != nil {
		errorHandler.Error(errhand.ErrInvalidIR, err.Error(), errhand.Position{
			File: inputFile,
			Line: 1,
		})
	}

	if errorHandler.HasErrors() {
		fmt.Fprintf(os.Stderr, "Compilation failed during IR generation:\n")
		errorHandler.Report()
		return nil, fmt.Errorf("IR generation failed with %d error(s)", errorHandler.ErrorCount())
	}

	printVerbose(opts, "  IR generated: %d functions, %d globals\n",
		len(irResult.Functions), len(irResult.Globals))
	return irResult, nil
}

// runCodeGeneration generates assembly from IR.
func runCodeGeneration(irResult *ir.IR, inputFile string, opts *CompileOptions, errorHandler *errhand.ErrorHandler) (string, error) {
	printVerbose(opts, "Stage 5/5: Code generation\n")
	codeGen := codegen.NewCodeGenerator(errorHandler)
	assembly, err := codeGen.Generate(irResult)
	if err != nil {
		errorHandler.Error(errhand.ErrUnsupportedOp, err.Error(), errhand.Position{
			File: inputFile,
			Line: 1,
		})
	}

	if errorHandler.HasErrors() {
		fmt.Fprintf(os.Stderr, "Compilation failed during code generation:\n")
		errorHandler.Report()
		return "", fmt.Errorf("code generation failed with %d error(s)", errorHandler.ErrorCount())
	}

	printVerbose(opts, "  Assembly generated: %d bytes\n", len(assembly))
	return assembly, nil
}

// outputResult outputs the compilation result based on options.
func outputResult(assembly string, opts *CompileOptions, errorHandler *errhand.ErrorHandler) error {
	if opts.EmitAssembly {
		return outputAssembly(assembly, opts)
	} else if opts.EmitObject {
		return outputObjectFile(assembly, opts, errorHandler)
	}
	return outputExecutable(assembly, opts, errorHandler)
}

// outputAssembly outputs assembly to file or stdout.
func outputAssembly(assembly string, opts *CompileOptions) error {
	if opts.Output != "" {
		if err := os.WriteFile(opts.Output, []byte(assembly), 0644); err != nil {
			return fmt.Errorf("writing assembly to '%s': %w", opts.Output, err)
		}
		printVerbose(opts, "Assembly written to: %s\n", opts.Output)
	} else {
		fmt.Print(assembly)
	}
	return nil
}

// outputObjectFile outputs an object file.
func outputObjectFile(assembly string, opts *CompileOptions, errorHandler *errhand.ErrorHandler) error {
	link := linker.NewLinker(errorHandler)
	outputPath := opts.Output
	if outputPath == "" {
		baseName := strings.TrimSuffix(opts.InputFile, ".c")
		outputPath = baseName + ".o"
	}
	if err := link.CompileToObject(assembly, outputPath); err != nil {
		return fmt.Errorf("compiling to object: %w", err)
	}
	printVerbose(opts, "Object file written to: %s\n", outputPath)
	return nil
}

// outputExecutable outputs an executable file.
func outputExecutable(assembly string, opts *CompileOptions, errorHandler *errhand.ErrorHandler) error {
	link := linker.NewLinker(errorHandler)
	outputPath := opts.Output
	if outputPath == "" {
		baseName := strings.TrimSuffix(opts.InputFile, ".c")
		outputPath = baseName + ".exe"
	}
	if err := link.LinkAssembly(assembly, outputPath); err != nil {
		return fmt.Errorf("linking: %w", err)
	}
	printVerbose(opts, "Executable written to: %s\n", outputPath)
	return nil
}

// printVerbose prints a message if verbose mode is enabled.
func printVerbose(opts *CompileOptions, format string, args ...interface{}) {
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "[compile] "+format, args...)
	}
}

// ParseCompileFlags parses compile command flags.
// Supported flags:
//   -o <file>  : Output file path
//   -S         : Output assembly only (don't assemble)
//   -c         : Output object file only (don't link)
//   -v, --verbose     : Verbose output
//   -d, --debug       : Debug mode (extra debugging information)
//   -t, --target <arch> : Target architecture (e.g., x86_64, arm64)
//   -O, --optimize <level> : Optimization level (0, 1, 2, 3, s, z)
func ParseCompileFlags(args []string) (*CompileOptions, error) {
	opts := &CompileOptions{
		EmitAssembly: false,
		EmitObject:   false,
		Verbose:      false,
		Debug:        false,
		Target:       "",
		Optimize:     "",
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
		} else if arg == "-d" || arg == "--debug" {
			opts.Debug = true
		} else if arg == "-t" || strings.HasPrefix(arg, "--target") {
			if arg == "-t" {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("-t requires an argument")
				}
				i++
				opts.Target = args[i]
			} else {
				// Handle --target=value or --target value
				if strings.HasPrefix(arg, "--target=") {
					opts.Target = strings.TrimPrefix(arg, "--target=")
				} else if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
					opts.Target = args[i]
				} else {
					return nil, fmt.Errorf("--target requires an argument")
				}
			}
		} else if arg == "-O" || strings.HasPrefix(arg, "--optimize") {
			if arg == "-O" {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("-O requires an argument")
				}
				i++
				opts.Optimize = args[i]
			} else {
				// Handle --optimize=value or --optimize value
				if strings.HasPrefix(arg, "--optimize=") {
					opts.Optimize = strings.TrimPrefix(arg, "--optimize=")
				} else if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
					opts.Optimize = args[i]
				} else {
					return nil, fmt.Errorf("--optimize requires an argument")
				}
			}
			// Validate optimization level
			if !isValidOptimizeLevel(opts.Optimize) {
				return nil, fmt.Errorf("invalid optimization level '%s'. Must be 0, 1, 2, 3, s, or z", opts.Optimize)
			}
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

// isValidOptimizeLevel checks if the optimization level is valid.
func isValidOptimizeLevel(level string) bool {
	validLevels := map[string]bool{
		"0": true,
		"1": true,
		"2": true,
		"3": true,
		"s": true,
		"z": true,
	}
	return validLevels[level]
}

// validateOptimizeLevel validates the optimization level and returns an error if invalid.
// Valid levels: 0, 1, 2, 3, s, z
func validateOptimizeLevel(level string) error {
	if level == "" {
		return nil // No optimization specified is valid
	}
	validLevels := []string{"0", "1", "2", "3", "s", "z"}
	for _, valid := range validLevels {
		if level == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid optimization level '%s'. Must be 0, 1, 2, 3, s, or z", level)
}

// validateTarget validates the target architecture.
func validateTarget(target string) error {
	if target == "" {
		return nil // No target specified is valid (uses default)
	}
	validTargets := []string{"x86_64", "arm64", "x86", "arm", "riscv64"}
	for _, valid := range validTargets {
		if target == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid target architecture '%s'. Supported: x86_64, arm64, x86, arm, riscv64", target)
}

// ParseInt parses a string to an integer with validation.
func ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}