package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/akzj/goc/pkg/cli"
	"github.com/akzj/goc/pkg/lexer"
)

// Version information
const (
	Version = "0.1.0"
	Name    = "goc"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "version", "--version", "-v":
		printVersion()
	case "help", "--help", "-h":
		printUsage()
	case "compile":
		if err := cli.CompileCommand(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "tokenize":
		handleTokenizeCommand(os.Args[2:])
	case "parse":
		// TODO: Implement parsing
		fmt.Println("Parsing not yet implemented")
		os.Exit(1)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf("%s - A C compiler written in Go\n\n", Name)
	fmt.Printf("Usage: %s <command> [options]\n\n", Name)
	fmt.Println("Commands:")
	fmt.Println("  compile    Compile a C source file to an executable")
	fmt.Println("  tokenize   Tokenize a C source file (debug)")
	fmt.Println("  parse      Parse a C source file and print AST (debug)")
	fmt.Println("  version    Show version information")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s compile hello.c -o hello\n", Name)
	fmt.Printf("  %s tokenize hello.c\n", Name)
	fmt.Printf("  %s parse hello.c\n", Name)
}

func printVersion() {
	fmt.Printf("%s version %s\n", Name, Version)
	fmt.Println("A C compiler written in Go")
	fmt.Println("Developed by Zero-FAS (Self-Aware Memory System)")
}

// Tokenize command flags and options
type tokenizeOptions struct {
	filePath string
	format   string // "default", "json", "compact"
}

// handleTokenizeCommand handles the tokenize subcommand
func handleTokenizeCommand(args []string) {
	opts := parseTokenizeArgs(args)
	if opts.filePath == "" {
		fmt.Fprintln(os.Stderr, "Error: no input file specified")
		fmt.Fprintln(os.Stderr, "Usage: goc tokenize <file.c> [--format=default|json|compact]")
		os.Exit(1)
	}

	// Read the source file
	source, err := os.ReadFile(opts.filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file '%s': %v\n", opts.filePath, err)
		os.Exit(1)
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
}

// parseTokenizeArgs parses command-line arguments for the tokenize command
func parseTokenizeArgs(args []string) tokenizeOptions {
	opts := tokenizeOptions{
		format: "default",
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--format=") {
			opts.format = strings.TrimPrefix(arg, "--format=")
			// Validate format
			if opts.format != "default" && opts.format != "json" && opts.format != "compact" {
				fmt.Fprintf(os.Stderr, "Error: invalid format '%s'. Must be 'default', 'json', or 'compact'\n", opts.format)
				os.Exit(1)
			}
		} else if arg == "--format" && i+1 < len(args) {
			i++
			opts.format = args[i]
			// Validate format
			if opts.format != "default" && opts.format != "json" && opts.format != "compact" {
				fmt.Fprintf(os.Stderr, "Error: invalid format '%s'. Must be 'default', 'json', or 'compact'\n", opts.format)
				os.Exit(1)
			}
		} else if !strings.HasPrefix(arg, "-") {
			opts.filePath = arg
		}
	}

	return opts
}

// outputTokensDefault outputs tokens in human-readable format
func outputTokensDefault(tokens []lexer.Token) {
	for _, tok := range tokens {
		// Skip EOF token in output
		if tok.Type == lexer.EOF {
			continue
		}
		fmt.Println(tok.String())
	}
}

// TokenJSON is a JSON-serializable representation of a token
type TokenJSON struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	HasSpace bool   `json:"hasSpace,omitempty"`
}

// outputTokensJSON outputs tokens as a JSON array
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
		os.Exit(1)
	}
}

// outputTokensCompact outputs tokens in a compact one-per-line format
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