package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestMain(t *testing.T) {
	// Test version command
	cmd := exec.Command(os.Args[0], "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if len(output) == 0 {
		t.Error("version command produced no output")
	}
}

func TestVersionFlag(t *testing.T) {
	// Test --version flag
	cmd := exec.Command(os.Args[0], "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--version flag failed: %v", err)
	}
	if len(output) == 0 {
		t.Error("--version flag produced no output")
	}
}

func TestHelpCommand(t *testing.T) {
	// Test help command
	cmd := exec.Command(os.Args[0], "help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}
	if len(output) == 0 {
		t.Error("help command produced no output")
	}
}