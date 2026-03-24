// Package cli provides the command-line interface for the GOC compiler.
// This file contains comprehensive unit tests for the CLI framework.
package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureOutput captures stdout output from a function
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestNewCLI(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	if cli == nil {
		t.Fatal("NewCLI returned nil")
	}
	
	if cli.name != "goc" {
		t.Errorf("Expected name 'goc', got '%s'", cli.name)
	}
	
	if cli.version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", cli.version)
	}
	
	if cli.description != "GOC Compiler" {
		t.Errorf("Expected description 'GOC Compiler', got '%s'", cli.description)
	}
	
	if cli.commands == nil {
		t.Fatal("Commands map is nil")
	}
	
	// Check built-in commands are registered
	if _, exists := cli.commands["help"]; !exists {
		t.Error("Help command not registered")
	}
	
	if _, exists := cli.commands["version"]; !exists {
		t.Error("Version command not registered")
	}
}

func TestRegisterCommand(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	// Test registering a new command
	handler := func(args []string, flags map[string]interface{}) error {
		return nil
	}
	
	cmd := &Command{
		Name:        "test",
		Description: "Test command",
		Usage:       "goc test [options]",
		Handler:     handler,
		Flags: []Flag{
			{Name: "verbose", Short: "v", Description: "Verbose output", HasValue: false},
		},
	}
	
	cli.RegisterCommand(cmd)
	
	if _, exists := cli.commands["test"]; !exists {
		t.Error("Test command not registered")
	}
	
	// Test registering command with empty name (should be ignored)
	emptyCmd := &Command{Name: ""}
	cli.RegisterCommand(emptyCmd)
	
	if _, exists := cli.commands[""]; exists {
		t.Error("Empty command name should not be registered")
	}
}

func TestCLI_Run_NoArgs(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	output := captureOutput(func() {
		err := cli.Run([]string{})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
	
	if !strings.Contains(output, "Usage:") {
		t.Error("Expected usage output")
	}
}

func TestCLI_Run_HelpFlag(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	tests := []struct {
		name string
		args []string
	}{
		{"short help", []string{"-h"}},
		{"long help", []string{"--help"}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				err := cli.Run(tt.args)
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			})
			
			if !strings.Contains(output, "Usage:") {
				t.Error("Expected usage output")
			}
		})
	}
}

func TestCLI_Run_VersionFlag(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	tests := []struct {
		name string
		args []string
	}{
		{"short version", []string{"-v"}},
		{"long version", []string{"--version"}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				err := cli.Run(tt.args)
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			})
			
			if !strings.Contains(output, "1.0.0") {
				t.Error("Expected version output")
			}
		})
	}
}

func TestCLI_Run_UnknownCommand(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	err := cli.Run([]string{"unknown"})
	if err == nil {
		t.Error("Expected error for unknown command")
	}
}

func TestCLI_Run_HelpCommand(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	// Register a test command
	cli.RegisterCommand(&Command{
		Name:        "compile",
		Description: "Compile a source file",
		Usage:       "goc compile [options] <file>",
		Handler:     func(args []string, flags map[string]interface{}) error { return nil },
	})
	
	output := captureOutput(func() {
		err := cli.Run([]string{"help"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
	
	if !strings.Contains(output, "compile") {
		t.Error("Expected compile command in help output")
	}
}

func TestCLI_Run_VersionCommand(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	output := captureOutput(func() {
		err := cli.Run([]string{"version"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
	
	if !strings.Contains(output, "1.0.0") {
		t.Error("Expected version output")
	}
}

func TestCLI_Run_CommandWithFlags(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	var capturedArgs []string
	var capturedFlags map[string]interface{}
	
	cli.RegisterCommand(&Command{
		Name: "test",
		Flags: []Flag{
			{Name: "output", Short: "o", Description: "Output file", HasValue: true},
			{Name: "verbose", Short: "v", Description: "Verbose", HasValue: false},
		},
		Handler: func(args []string, flags map[string]interface{}) error {
			capturedArgs = args
			capturedFlags = flags
			return nil
		},
	})
	
	err := cli.Run([]string{"test", "-v", "-o", "output.s", "input.c"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if capturedFlags["verbose"] != true {
		t.Errorf("Expected verbose=true, got %v", capturedFlags["verbose"])
	}
	
	if capturedFlags["output"] != "output.s" {
		t.Errorf("Expected output='output.s', got %v", capturedFlags["output"])
	}
	
	if len(capturedArgs) != 1 || capturedArgs[0] != "input.c" {
		t.Errorf("Expected args=['input.c'], got %v", capturedArgs)
	}
}

func TestCLI_Run_CommandHelpFlag(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	cli.RegisterCommand(&Command{
		Name:        "test",
		Description: "Test command",
		Usage:       "goc test [options]",
		Handler:     func(args []string, flags map[string]interface{}) error { return nil },
	})
	
	output := captureOutput(func() {
		err := cli.Run([]string{"test", "--help"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
	
	if !strings.Contains(output, "Test command") {
		t.Error("Expected command description in help output")
	}
}

func TestCLI_Run_UnknownFlag(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	cli.RegisterCommand(&Command{
		Name: "test",
		Flags: []Flag{
			{Name: "verbose", Short: "v", HasValue: false},
		},
		Handler: func(args []string, flags map[string]interface{}) error { return nil },
	})
	
	err := cli.Run([]string{"test", "--unknown"})
	if err == nil {
		t.Error("Expected error for unknown flag")
	}
}

func TestCLI_Run_FlagMissingValue(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	cli.RegisterCommand(&Command{
		Name: "test",
		Flags: []Flag{
			{Name: "output", Short: "o", HasValue: true},
		},
		Handler: func(args []string, flags map[string]interface{}) error { return nil },
	})
	
	err := cli.Run([]string{"test", "--output"})
	if err == nil {
		t.Error("Expected error for flag missing value")
	}
}

func TestCLI_PrintUsage(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	cli.RegisterCommand(&Command{
		Name:        "compile",
		Description: "Compile source file",
		Usage:       "goc compile [options] <file>",
		Hidden:      false,
	})
	
	cli.RegisterCommand(&Command{
		Name:   "hidden",
		Hidden: true,
	})
	
	output := captureOutput(func() {
		cli.PrintUsage()
	})
	
	if !strings.Contains(output, "goc") {
		t.Error("Expected program name in usage")
	}
	
	if !strings.Contains(output, "compile") {
		t.Error("Expected compile command in usage")
	}
	
	if strings.Contains(output, "hidden") {
		t.Error("Hidden command should not appear in usage")
	}
}

func TestCLI_PrintVersion(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	output := captureOutput(func() {
		cli.PrintVersion()
	})
	
	if !strings.Contains(output, "goc") {
		t.Error("Expected program name in version")
	}
	
	if !strings.Contains(output, "1.0.0") {
		t.Error("Expected version number in version output")
	}
}

func TestCLI_printCommandHelp(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	cmd := &Command{
		Name:        "compile",
		Description: "Compile a source file",
		Usage:       "goc compile [options] <file>",
		Examples: []string{
			"goc compile main.c",
			"goc compile -o main.s main.c",
		},
		Flags: []Flag{
			{Name: "output", Short: "o", Description: "Output file", HasValue: true},
			{Name: "verbose", Short: "v", Description: "Verbose output", HasValue: false},
		},
	}
	
	output := captureOutput(func() {
		cli.printCommandHelp(cmd)
	})
	
	if !strings.Contains(output, "Compile a source file") {
		t.Error("Expected command description")
	}
	
	if !strings.Contains(output, "Usage:") {
		t.Error("Expected usage section")
	}
	
	if !strings.Contains(output, "Flags:") && !strings.Contains(output, "Options:") {
		t.Error("Expected flags/options section")
	}
	
	if !strings.Contains(output, "Examples:") {
		t.Error("Expected examples section")
	}
}

func TestCLI_getSortedCommands(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	// Register commands in random order (in addition to built-in help/version)
	cli.RegisterCommand(&Command{Name: "zebra", Hidden: false})
	cli.RegisterCommand(&Command{Name: "alpha", Hidden: false})
	cli.RegisterCommand(&Command{Name: "middle", Hidden: false})
	cli.RegisterCommand(&Command{Name: "hidden", Hidden: true})
	
	sorted := cli.getSortedCommands()
	
	// Note: getSortedCommands returns ALL commands including hidden ones
	// Should have built-in commands (help, version) + our 4 commands
	if len(sorted) < 4 {
		t.Errorf("Expected at least 4 commands, got %d", len(sorted))
	}
	
	// Verify our specific commands are present
	found := make(map[string]bool)
	for _, cmd := range sorted {
		if cmd.Name == "alpha" || cmd.Name == "middle" || cmd.Name == "zebra" || cmd.Name == "hidden" {
			found[cmd.Name] = true
		}
	}
	
	if !found["alpha"] || !found["middle"] || !found["zebra"] || !found["hidden"] {
		t.Errorf("Expected to find alpha, middle, zebra, hidden; got all commands")
	}
	
	// Verify commands are sorted alphabetically
	for i := 0; i < len(sorted)-1; i++ {
		if sorted[i].Name > sorted[i+1].Name {
			t.Errorf("Commands not sorted: %s > %s", sorted[i].Name, sorted[i+1].Name)
		}
	}
}

func TestCLI_parseFlags_LongFlagWithValue(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	cmd := &Command{
		Name: "test",
		Flags: []Flag{
			{Name: "output", Short: "o", HasValue: true},
		},
	}
	
	_, flags, err := cli.parseFlags(cmd, []string{"--output", "file.s", "input.c"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if flags["output"] != "file.s" {
		t.Errorf("Expected output='file.s', got %v", flags["output"])
	}
}

func TestCLI_parseFlags_ShortFlagWithValue(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	cmd := &Command{
		Name: "test",
		Flags: []Flag{
			{Name: "output", Short: "o", HasValue: true},
		},
	}
	
	_, flags, err := cli.parseFlags(cmd, []string{"-o", "file.s", "input.c"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if flags["output"] != "file.s" {
		t.Errorf("Expected output='file.s', got %v", flags["output"])
	}
}

func TestCLI_parseFlags_FlagWithDefaultValue(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	cmd := &Command{
		Name: "test",
		Flags: []Flag{
			{Name: "optimize", Short: "O", HasValue: true, Default: "0"},
		},
	}
	
	_, flags, err := cli.parseFlags(cmd, []string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if flags["optimize"] != "0" {
		t.Errorf("Expected optimize='0' (default), got %v", flags["optimize"])
	}
}

func TestCLI_parseFlags_MultipleFlags(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	cmd := &Command{
		Name: "test",
		Flags: []Flag{
			{Name: "verbose", Short: "v", HasValue: false},
			{Name: "output", Short: "o", HasValue: true},
			{Name: "debug", Short: "d", HasValue: false},
		},
	}
	
	_, flags, err := cli.parseFlags(cmd, []string{"-v", "-o", "out.s", "-d"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if flags["verbose"] != true {
		t.Error("Expected verbose=true")
	}
	
	if flags["output"] != "out.s" {
		t.Errorf("Expected output='out.s', got %v", flags["output"])
	}
	
	if flags["debug"] != true {
		t.Error("Expected debug=true")
	}
}

func TestCLI_findFlag(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	cmd := &Command{
		Name: "test",
		Flags: []Flag{
			{Name: "output", Short: "o", HasValue: true},
			{Name: "verbose", Short: "v", HasValue: false},
		},
	}
	
	// Find by long name
	flag := cli.findFlag(cmd, "output", "")
	if flag == nil || flag.Name != "output" {
		t.Error("Expected to find flag by long name")
	}
	
	// Find by short name
	flag = cli.findFlag(cmd, "", "o")
	if flag == nil || flag.Short != "o" {
		t.Error("Expected to find flag by short name")
	}
	
	// Find non-existent flag
	flag = cli.findFlag(cmd, "nonexistent", "x")
	if flag != nil {
		t.Error("Expected nil for non-existent flag")
	}
}

func TestCommandHandler(t *testing.T) {
	handlerCalled := false
	
	handler := func(args []string, flags map[string]interface{}) error {
		handlerCalled = true
		return nil
	}
	
	cmd := &Command{
		Name:    "test",
		Handler: handler,
	}
	
	err := cmd.Handler([]string{"arg1"}, map[string]interface{}{"flag": true})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

func TestCLI_Run_CommandReturnsError(t *testing.T) {
	cli := NewCLI("goc", "1.0.0", "GOC Compiler")
	
	cli.RegisterCommand(&Command{
		Name: "test",
		Handler: func(args []string, flags map[string]interface{}) error {
			return os.ErrNotExist
		},
	})
	
	err := cli.Run([]string{"test"})
	if err == nil {
		t.Error("Expected error from command handler")
	}
}
// TestCLI_helpHandler tests the helpHandler method
func TestCLI_helpHandler(t *testing.T) {
	t.Run("NoArgs_GeneralHelp", func(t *testing.T) {
		cli := NewCLI("goc", "1.0.0", "GOC Compiler")
		
		cli.RegisterCommand(&Command{
			Name:        "compile",
			Description: "Compile a source file",
			Usage:       "goc compile [options] <file>",
			Handler:     func(args []string, flags map[string]interface{}) error { return nil },
		})
		
		output := captureOutput(func() {
			err := cli.helpHandler([]string{}, map[string]interface{}{})
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
		
		if !strings.Contains(output, "Usage:") {
			t.Error("Expected general help output to contain 'Usage:'")
		}
		if !strings.Contains(output, "compile") {
			t.Error("Expected general help output to contain 'compile' command")
		}
	})
	
	t.Run("KnownCommand_CommandHelp", func(t *testing.T) {
		cli := NewCLI("goc", "1.0.0", "GOC Compiler")
		
		cli.RegisterCommand(&Command{
			Name:        "compile",
			Description: "Compile a source file",
			Usage:       "goc compile [options] <file>",
			Handler:     func(args []string, flags map[string]interface{}) error { return nil },
		})
		
		output := captureOutput(func() {
			err := cli.helpHandler([]string{"compile"}, map[string]interface{}{})
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
		
		if !strings.Contains(output, "compile") {
			t.Error("Expected command help to contain command name")
		}
		if !strings.Contains(output, "Compile a source file") {
			t.Error("Expected command help to contain description")
		}
	})
	
	t.Run("UnknownCommand_Error", func(t *testing.T) {
		cli := NewCLI("goc", "1.0.0", "GOC Compiler")
		
		err := cli.helpHandler([]string{"unknown"}, map[string]interface{}{})
		if err == nil {
			t.Error("Expected error for unknown command")
		}
		if !strings.Contains(err.Error(), "unknown command") {
			t.Errorf("Expected error to contain 'unknown command', got: %v", err)
		}
	})
}

// captureStderr captures stderr output from a function
func captureStderr(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
