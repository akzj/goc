package main

import (
	"encoding/json"
	"fmt"
	"os"
	

	"github.com/akzj/goc/pkg/cli"
	"github.com/akzj/goc/pkg/lexer"
)

// exitFunc allows tests to override os.Exit for testing
var exitFunc = os.Exit

// Version information
const (
	Version     = "0.1.0"
	Name        = "goc"
	Description = "A C compiler written in Go"
)

func main() {
	// Create CLI instance
	cliApp := cli.NewCLI(Name, Version, Description)

	// Register compile command
	cliApp.RegisterCommand(&cli.Command{
		Name:        "compile",
		Description: "Compile a C source file to an executable",
		Usage:       "compile <source.c> [options]",
		Examples: []string{
			fmt.Sprintf("%s compile hello.c -o hello", Name),
			fmt.Sprintf("%s compile main.c -O2 -g", Name),
		},
		Flags: []cli.Flag{
			{Name: "output", Short: "o", Description: "Output file", Default: "", HasValue: true},
			{Name: "assembly", Short: "S", Description: "Output assembly only", HasValue: false},
			{Name: "compile-only", Short: "c", Description: "Compile to object file only", HasValue: false},
			{Name: "preprocess", Short: "E", Description: "Preprocess only", HasValue: false},
			{Name: "include", Short: "I", Description: "Add include directory", HasValue: true},
			{Name: "define", Short: "D", Description: "Define macro", HasValue: true},
			{Name: "optimize", Short: "O", Description: "Optimization level (0-3)", Default: "0", HasValue: true},
			{Name: "debug", Short: "g", Description: "Generate debug info", HasValue: false},
			{Name: "verbose", Short: "v", Description: "Verbose output", HasValue: false},
			{Name: "target", Short: "", Description: "Target architecture", Default: "x86-64", HasValue: true},
			{Name: "help", Short: "h", Description: "Show help for this command", HasValue: false},
		},
		Handler: handleCompileCommand,
	})

	// Register tokenize command
	cliApp.RegisterCommand(&cli.Command{
		Name:        "tokenize",
		Description: "Tokenize a C source file (debug)",
		Usage:       "tokenize <file.c> [options]",
		Examples: []string{
			fmt.Sprintf("%s tokenize hello.c", Name),
			fmt.Sprintf("%s tokenize main.c --format=json", Name),
		},
		Flags: []cli.Flag{
			{Name: "format", Short: "f", Description: "Output format (default, json, compact)", Default: "default", HasValue: true},
			{Name: "help", Short: "h", Description: "Show help for this command", HasValue: false},
		},
		Handler: handleTokenizeCommand,
	})

	// Register parse command
	cliApp.RegisterCommand(&cli.Command{
		Name:        "parse",
		Description: "Parse a C source file and print AST (debug)",
		Usage:       "parse <file.c> [options]",
		Examples: []string{
			fmt.Sprintf("%s parse hello.c", Name),
		},
		Flags: []cli.Flag{
			{Name: "help", Short: "h", Description: "Show help for this command", HasValue: false},
		},
		Handler: handleParseCommand,
	})

	// Run the CLI
	if err := cliApp.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		exitFunc(1)
	}
}

// handleCompileCommand handles the compile subcommand.
func handleCompileCommand(args []string, flags map[string]interface{}) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no input file specified")
		fmt.Fprintf(os.Stderr, "Usage: %s compile <source.c> [options]\n", Name)
		fmt.Fprintln(os.Stderr, "Run 'goc help compile' for more information.")
		exitFunc(1)
	}

	var compileArgs []string
	if v, ok := flags["assembly"].(bool); ok && v {
		compileArgs = append(compileArgs, "-S")
	}
	if v, ok := flags["compile-only"].(bool); ok && v {
		compileArgs = append(compileArgs, "-c")
	}
	if v, ok := flags["verbose"].(bool); ok && v {
		compileArgs = append(compileArgs, "-v")
	}
	if v, ok := flags["debug"].(bool); ok && v {
		compileArgs = append(compileArgs, "-d")
	}
	if out, ok := flags["output"].(string); ok && out != "" {
		compileArgs = append(compileArgs, "-o", out)
	}
	if t, ok := flags["target"].(string); ok && t != "" {
		compileArgs = append(compileArgs, "-t", t)
	}
	if o, ok := flags["optimize"].(string); ok && o != "" {
		compileArgs = append(compileArgs, "-O", o)
	}
	compileArgs = append(compileArgs, args[0])

	return cli.CompileCommand(compileArgs)
}

// tokenizeOptions holds options for the tokenize command.
type tokenizeOptions struct {
	filePath string
	format   string // "default", "json", "compact"
}

// handleTokenizeCommand handles the tokenize subcommand.
func handleTokenizeCommand(args []string, flags map[string]interface{}) error {
	opts := parseTokenizeOptions(args, flags)

	if opts.filePath == "" {
		fmt.Fprintln(os.Stderr, "Error: no input file specified")
		fmt.Fprintln(os.Stderr, "Usage: goc tokenize <file.c> [--format=default|json|compact]")
		exitFunc(1)
	}

	// Read the source file
	source, err := os.ReadFile(opts.filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file '%s': %v\n", opts.filePath, err)
		exitFunc(1)
	}

	// Create lexer and tokenize
	l := lexer.NewLexer(string(source), opts.filePath)
	tokens := l.Tokenize()

	// Output in the specified format
	switch opts.format {
	case "json":
		outputTokensJSON(tokens)
	case "compact":
		outputTokensCompact(tokens)
	default:
		outputTokensDefault(tokens)
	}

	return nil
}

// parseTokenizeOptions parses options for the tokenize command.
func parseTokenizeOptions(args []string, flags map[string]interface{}) tokenizeOptions {
	opts := tokenizeOptions{
		format: "default",
	}

	if format, ok := flags["format"].(string); ok {
		opts.format = format
	}

	// Get file path from positional arguments
	if len(args) > 0 {
		opts.filePath = args[0]
	}

	return opts
}

// handleParseCommand handles the parse subcommand.
func handleParseCommand(args []string, flags map[string]interface{}) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no input file specified")
		fmt.Fprintln(os.Stderr, "Usage: goc parse <file.c>")
		exitFunc(1)
	}

	fmt.Println("Parsing not yet implemented")
	exitFunc(1)
	return nil
}

// outputTokensDefault outputs tokens in human-readable format.
func outputTokensDefault(tokens []lexer.Token) {
	for _, tok := range tokens {
		// Skip EOF token in output
		if tok.Type == lexer.EOF {
			continue
		}
		fmt.Println(tok.String())
	}
}

// TokenJSON is a JSON-serializable representation of a token.
type TokenJSON struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	HasSpace bool   `json:"hasSpace,omitempty"`
}

// outputTokensJSON outputs tokens as a JSON array.
func outputTokensJSON(tokens []lexer.Token) {
	// Convert tokens to JSON-friendly format
	jsonTokens := make([]TokenJSON, 0, len(tokens))
	for _, tok := range tokens {
		// Skip EOF token in output
		if tok.Type == lexer.EOF {
			continue
		}
		jsonTokens = append(jsonTokens, TokenJSON{
			Type:     string(tok.Type),
			Value:    tok.Value,
			File:     tok.Pos.File,
			Line:     tok.Pos.Line,
			Column:   tok.Pos.Column,
			HasSpace: tok.HasSpace,
		})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(jsonTokens); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		exitFunc(1)
	}
}

// outputTokensCompact outputs tokens in a compact one-per-line format.
func outputTokensCompact(tokens []lexer.Token) {
	for _, tok := range tokens {
		// Skip EOF token in output
		if tok.Type == lexer.EOF {
			continue
		}
		// Format: TYPE VALUE LINE:COLUMN
		if tok.Value != "" {
			fmt.Printf("%s\t%s\t%d:%d\n", tok.Type, tok.Value, tok.Pos.Line, tok.Pos.Column)
		} else {
			fmt.Printf("%s\t\t%d:%d\n", tok.Type, tok.Pos.Line, tok.Pos.Column)
		}
	}
}