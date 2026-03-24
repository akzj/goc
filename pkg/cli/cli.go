// Package cli provides the command-line interface for the GOC compiler.
// This file defines the CLI framework with help system support.
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/akzj/goc/internal/errhand"
)

// CLI represents the command-line interface.
type CLI struct {
	// name is the program name.
	name string
	// version is the version string.
	version string
	// description is the program description.
	description string
	// commands maps command names to handlers.
	commands map[string]*Command
	// errors is the error handler.
	errs *errhand.ErrorHandler
}

// Command represents a CLI command.
type Command struct {
	// Name is the command name.
	Name string
	// Description is the command description.
	Description string
	// Usage is the usage string for the command.
	Usage string
	// Examples is a list of usage examples.
	Examples []string
	// Handler is the command handler function.
	Handler CommandHandler
	// Flags is the list of command flags.
	Flags []Flag
	// Hidden indicates if the command should be hidden from help.
	Hidden bool
}

// CommandHandler handles a command.
type CommandHandler func(args []string, flags map[string]interface{}) error

// Flag represents a command-line flag.
type Flag struct {
	// Name is the flag name (long form, e.g., "output").
	Name string
	// Short is the short flag name (single character, e.g., "o").
	Short string
	// Description is the flag description.
	Description string
	// Default is the default value.
	Default interface{}
	// HasValue indicates if the flag requires a value.
	HasValue bool
}

// NewCLI creates a new CLI with the given name and version.
func NewCLI(name, version, description string) *CLI {
	cli := &CLI{
		name:        name,
		version:     version,
		description: description,
		commands:    make(map[string]*Command),
		errs:        errhand.NewErrorHandler(),
	}

	// Register built-in help command
	cli.RegisterCommand(&Command{
		Name:        "help",
		Description: "Show help information",
		Usage:       fmt.Sprintf("%s help [command]", name),
		Examples: []string{
			fmt.Sprintf("%s help", name),
			fmt.Sprintf("%s help compile", name),
		},
		Handler: cli.helpHandler,
		Hidden:  true,
	})

	// Register built-in version command
	cli.RegisterCommand(&Command{
		Name:        "version",
		Description: "Show version information",
		Usage:       fmt.Sprintf("%s version", name),
		Handler:     cli.versionHandler,
		Hidden:      true,
	})

	return cli
}

// RegisterCommand registers a command with the CLI.
func (cli *CLI) RegisterCommand(cmd *Command) {
	if cmd.Name == "" {
		return
	}
	cli.commands[cmd.Name] = cmd
}

// Run runs the CLI with the given arguments.
func (cli *CLI) Run(args []string) error {
	if len(args) < 1 {
		cli.PrintUsage()
		return nil
	}

	// Check for global help flags
	if args[0] == "--help" || args[0] == "-h" {
		cli.PrintUsage()
		return nil
	}

	// Check for global version flags
	if args[0] == "--version" || args[0] == "-v" {
		cli.PrintVersion()
		return nil
	}

	// Find and execute command
	cmdName := args[0]
	cmd, exists := cli.commands[cmdName]
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", cmdName)
		cli.PrintUsage()
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	// Parse flags and arguments for the command
	cmdArgs, flags, err := cli.parseFlags(cmd, args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		cli.printCommandHelp(cmd)
		return err
	}

	// Check for command-specific help
	if flags["help"] == true {
		cli.printCommandHelp(cmd)
		return nil
	}

	// Execute command handler
	if cmd.Handler != nil {
		return cmd.Handler(cmdArgs, flags)
	}

	return nil
}

// parseFlags parses command-line flags for a command.
func (cli *CLI) parseFlags(cmd *Command, args []string) ([]string, map[string]interface{}, error) {
	flags := make(map[string]interface{})
	var cmdArgs []string

	// Initialize flags with defaults
	for _, flag := range cmd.Flags {
		if flag.Default != nil {
			flags[flag.Name] = flag.Default
		}
	}

	i := 0
	for i < len(args) {
		arg := args[i]

		if strings.HasPrefix(arg, "--") {
			// Long flag
			flagName := strings.TrimPrefix(arg, "--")
			if flagName == "help" {
				flags["help"] = true
				i++
				continue
			}

			// Find flag definition
			flagDef := cli.findFlag(cmd, flagName, "")
			if flagDef == nil {
				return nil, nil, fmt.Errorf("unknown flag: --%s", flagName)
			}

			if flagDef.HasValue {
				if i+1 >= len(args) {
					return nil, nil, fmt.Errorf("flag --%s requires a value", flagName)
				}
				i++
				flags[flagName] = args[i]
			} else {
				flags[flagName] = true
			}
			i++
		} else if strings.HasPrefix(arg, "-") && len(arg) == 2 {
			// Short flag
			flagName := string(arg[1])
			if flagName == "h" {
				flags["help"] = true
				i++
				continue
			}
			if flagName == "v" && cmd.Name == "" {
				// Global version flag already handled
				i++
				continue
			}

			// Find flag definition
			flagDef := cli.findFlag(cmd, "", flagName)
			if flagDef == nil {
				return nil, nil, fmt.Errorf("unknown flag: -%s", flagName)
			}

			if flagDef.HasValue {
				if i+1 >= len(args) {
					return nil, nil, fmt.Errorf("flag -%s requires a value", flagName)
				}
				i++
				flags[flagDef.Name] = args[i]
			} else {
				flags[flagDef.Name] = true
			}
			i++
		} else {
			// Regular argument
			cmdArgs = append(cmdArgs, arg)
			i++
		}
	}

	return cmdArgs, flags, nil
}

// findFlag finds a flag definition by name or short name.
func (cli *CLI) findFlag(cmd *Command, name, short string) *Flag {
	for i := range cmd.Flags {
		if (name != "" && cmd.Flags[i].Name == name) ||
			(short != "" && cmd.Flags[i].Short == short) {
			return &cmd.Flags[i]
		}
	}
	return nil
}

// helpHandler handles the help command.
func (cli *CLI) helpHandler(args []string, flags map[string]interface{}) error {
	if len(args) > 0 {
		// Show help for specific command
		cmdName := args[0]
		cmd, exists := cli.commands[cmdName]
		if !exists {
			fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", cmdName)
			cli.PrintUsage()
			return fmt.Errorf("unknown command: %s", cmdName)
		}
		cli.printCommandHelp(cmd)
	} else {
		// Show general help
		cli.PrintUsage()
	}
	return nil
}

// versionHandler handles the version command.
func (cli *CLI) versionHandler(args []string, flags map[string]interface{}) error {
	cli.PrintVersion()
	return nil
}

// PrintUsage prints the main usage message.
func (cli *CLI) PrintUsage() {
	fmt.Printf("%s - %s\n\n", cli.name, cli.description)
	fmt.Printf("Usage: %s <command> [options]\n\n", cli.name)

	fmt.Println("Commands:")
	for _, cmd := range cli.getSortedCommands() {
		if cmd.Hidden {
			continue
		}
		fmt.Printf("  %-12s %s\n", cmd.Name, cmd.Description)
	}
	fmt.Println()

	fmt.Println("Flags:")
	fmt.Println("  -h, --help     Show help information")
	fmt.Println("  -v, --version  Show version information")
	fmt.Println()

	fmt.Println("Examples:")
	if len(cli.commands) > 0 {
		for _, cmd := range cli.getSortedCommands() {
			if cmd.Hidden || len(cmd.Examples) == 0 {
				continue
			}
			fmt.Printf("  %s\n", cmd.Examples[0])
		}
	}
	fmt.Println()

	fmt.Printf("Run '%s help <command>' for more information on a command.\n", cli.name)
}

// printCommandHelp prints help for a specific command.
func (cli *CLI) printCommandHelp(cmd *Command) {
	fmt.Printf("Usage: %s %s\n\n", cli.name, cmd.Usage)

	if cmd.Description != "" {
		fmt.Printf("Description:\n  %s\n\n", cmd.Description)
	}

	if len(cmd.Flags) > 0 {
		fmt.Println("Options:")
		for _, flag := range cmd.Flags {
			flagStr := ""
			if flag.Short != "" {
				flagStr = fmt.Sprintf("  -%s, --%-10s", flag.Short, flag.Name)
			} else {
				flagStr = fmt.Sprintf("      --%-14s", flag.Name)
			}
			if flag.HasValue {
				if flag.Default != nil {
					fmt.Printf("%s <value>   %s (default: %v)\n", flagStr, flag.Description, flag.Default)
				} else {
					fmt.Printf("%s <value>   %s\n", flagStr, flag.Description)
				}
			} else {
				fmt.Printf("%s %s\n", flagStr, flag.Description)
			}
		}
		fmt.Println()
	}

	if len(cmd.Examples) > 0 {
		fmt.Println("Examples:")
		for _, example := range cmd.Examples {
			fmt.Printf("  %s\n", example)
		}
		fmt.Println()
	}
}

// PrintVersion prints the version information.
func (cli *CLI) PrintVersion() {
	fmt.Printf("%s version %s\n", cli.name, cli.version)
	if cli.description != "" {
		fmt.Println(cli.description)
	}
}

// getSortedCommands returns commands sorted by name.
func (cli *CLI) getSortedCommands() []*Command {
	var commands []*Command
	for _, cmd := range cli.commands {
		commands = append(commands, cmd)
	}

	// Simple bubble sort for small number of commands
	for i := 0; i < len(commands)-1; i++ {
		for j := 0; j < len(commands)-i-1; j++ {
			if commands[j].Name > commands[j+1].Name {
				commands[j], commands[j+1] = commands[j+1], commands[j]
			}
		}
	}

	return commands
}

