package main

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// gocCmd runs the real goc entrypoint via "go run ." in this package directory.
// Do not use os.Args[0] here: during "go test" it points at the test binary, so
// exec'ing it with "version" re-runs the full test suite in each child and can
// exhaust memory (recursive test + subprocess explosion).
func gocCmd(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	c := exec.Command("go", append([]string{"run", "."}, args...)...)
	c.Dir = dir
	return c
}

func TestCLIVersionCommand(t *testing.T) {
	cmd := gocCmd(t, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version command failed: %v\n%s", err, output)
	}
	if len(output) == 0 {
		t.Error("version command produced no output")
	}
}

func TestCLIVersionFlag(t *testing.T) {
	cmd := gocCmd(t, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--version flag failed: %v\n%s", err, output)
	}
	if len(output) == 0 {
		t.Error("--version flag produced no output")
	}
}

func TestCLIHelpCommand(t *testing.T) {
	cmd := gocCmd(t, "help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help command failed: %v\n%s", err, output)
	}
	if len(output) == 0 {
		t.Error("help command produced no output")
	}
}
