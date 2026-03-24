// Package cli provides the command-line interface for the GOC compiler.
// This file contains unit tests for the compile command flag parsing.
package cli

import (
	"testing"
)

func TestParseCompileFlags_Basic(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		verify  func(*CompileOptions) bool
	}{
		{
			name:    "empty args",
			args:    []string{},
			wantErr: false, // Empty args is valid, will error on InputFile == "" check in CompileCommand
			verify:  func(o *CompileOptions) bool { return o.InputFile == "" },
		},
		{
			name: "input file only",
			args: []string{"main.c"},
			wantErr: false,
			verify: func(o *CompileOptions) bool {
				return o.InputFile == "main.c" &&
					o.Output == "" &&
					!o.EmitAssembly &&
					!o.EmitObject &&
					!o.Verbose &&
					!o.Debug &&
					o.Target == "" &&
					o.Optimize == ""
			},
		},
		{
			name: "output flag -o",
			args: []string{"-o", "output.exe", "main.c"},
			wantErr: false,
			verify: func(o *CompileOptions) bool {
				return o.InputFile == "main.c" && o.Output == "output.exe"
			},
		},
		{
			name: "assembly flag -S",
			args: []string{"-S", "main.c"},
			wantErr: false,
			verify: func(o *CompileOptions) bool {
				return o.InputFile == "main.c" && o.EmitAssembly
			},
		},
		{
			name: "object flag -c",
			args: []string{"-c", "main.c"},
			wantErr: false,
			verify: func(o *CompileOptions) bool {
				return o.InputFile == "main.c" && o.EmitObject
			},
		},
		{
			name: "verbose flag -v",
			args: []string{"-v", "main.c"},
			wantErr: false,
			verify: func(o *CompileOptions) bool {
				return o.InputFile == "main.c" && o.Verbose
			},
		},
		{
			name: "verbose flag --verbose",
			args: []string{"--verbose", "main.c"},
			wantErr: false,
			verify: func(o *CompileOptions) bool {
				return o.InputFile == "main.c" && o.Verbose
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseCompileFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCompileFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.verify(opts) {
				t.Errorf("ParseCompileFlags() verification failed for %+v", opts)
			}
		})
	}
}

func TestParseCompileFlags_Debug(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		debug   bool
	}{
		{
			name:    "debug flag -d",
			args:    []string{"-d", "main.c"},
			wantErr: false,
			debug:   true,
		},
		{
			name:    "debug flag --debug",
			args:    []string{"--debug", "main.c"},
			wantErr: false,
			debug:   true,
		},
		{
			name:    "no debug flag",
			args:    []string{"main.c"},
			wantErr: false,
			debug:   false,
		},
		{
			name:    "debug with other flags",
			args:    []string{"-d", "-v", "-o", "out.exe", "main.c"},
			wantErr: false,
			debug:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseCompileFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCompileFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && opts.Debug != tt.debug {
				t.Errorf("ParseCompileFlags() Debug = %v, want %v", opts.Debug, tt.debug)
			}
		})
	}
}

func TestParseCompileFlags_Target(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		target  string
	}{
		{
			name:    "target flag -t x86_64",
			args:    []string{"-t", "x86_64", "main.c"},
			wantErr: false,
			target:  "x86_64",
		},
		{
			name:    "target flag --target x86_64",
			args:    []string{"--target", "x86_64", "main.c"},
			wantErr: false,
			target:  "x86_64",
		},
		{
			name:    "target flag --target=x86_64",
			args:    []string{"--target=x86_64", "main.c"},
			wantErr: false,
			target:  "x86_64",
		},
		{
			name:    "target flag -t arm64",
			args:    []string{"-t", "arm64", "main.c"},
			wantErr: false,
			target:  "arm64",
		},
		{
			name:    "target flag -t missing argument",
			args:    []string{"-t"},
			wantErr: true,
			target:  "",
		},
		{
			name:    "target flag --target missing argument",
			args:    []string{"--target"},
			wantErr: true,
			target:  "",
		},
		{
			name:    "no target flag",
			args:    []string{"main.c"},
			wantErr: false,
			target:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseCompileFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCompileFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && opts.Target != tt.target {
				t.Errorf("ParseCompileFlags() Target = %v, want %v", opts.Target, tt.target)
			}
		})
	}
}

func TestParseCompileFlags_Optimize(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		optimize string
	}{
		{
			name:     "optimize flag -O 0",
			args:     []string{"-O", "0", "main.c"},
			wantErr:  false,
			optimize: "0",
		},
		{
			name:     "optimize flag -O 1",
			args:     []string{"-O", "1", "main.c"},
			wantErr:  false,
			optimize: "1",
		},
		{
			name:     "optimize flag -O 2",
			args:     []string{"-O", "2", "main.c"},
			wantErr:  false,
			optimize: "2",
		},
		{
			name:     "optimize flag -O 3",
			args:     []string{"-O", "3", "main.c"},
			wantErr:  false,
			optimize: "3",
		},
		{
			name:     "optimize flag -O s",
			args:     []string{"-O", "s", "main.c"},
			wantErr:  false,
			optimize: "s",
		},
		{
			name:     "optimize flag -O z",
			args:     []string{"-O", "z", "main.c"},
			wantErr:  false,
			optimize: "z",
		},
		{
			name:     "optimize flag --optimize 2",
			args:     []string{"--optimize", "2", "main.c"},
			wantErr:  false,
			optimize: "2",
		},
		{
			name:     "optimize flag --optimize=3",
			args:     []string{"--optimize=3", "main.c"},
			wantErr:  false,
			optimize: "3",
		},
		{
			name:    "optimize flag -O missing argument",
			args:    []string{"-O"},
			wantErr: true,
		},
		{
			name:    "optimize flag --optimize missing argument",
			args:    []string{"--optimize"},
			wantErr: true,
		},
		{
			name:    "optimize flag invalid level",
			args:    []string{"-O", "invalid", "main.c"},
			wantErr: true,
		},
		{
			name:    "optimize flag invalid level 4",
			args:    []string{"-O", "4", "main.c"},
			wantErr: true,
		},
		{
			name:     "no optimize flag",
			args:     []string{"main.c"},
			wantErr:  false,
			optimize: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseCompileFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCompileFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && opts.Optimize != tt.optimize {
				t.Errorf("ParseCompileFlags() Optimize = %v, want %v", opts.Optimize, tt.optimize)
			}
		})
	}
}

func TestParseCompileFlags_Combined(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		verify  func(*CompileOptions) bool
	}{
		{
			name: "all flags combined",
			args: []string{"-d", "-v", "-S", "-o", "out.s", "-t", "x86_64", "-O", "2", "main.c"},
			wantErr: false,
			verify: func(o *CompileOptions) bool {
				return o.InputFile == "main.c" &&
					o.Output == "out.s" &&
					o.EmitAssembly &&
					!o.EmitObject &&
					o.Verbose &&
					o.Debug &&
					o.Target == "x86_64" &&
					o.Optimize == "2"
			},
		},
		{
			name: "long flags combined",
			args: []string{"--debug", "--verbose", "-S", "-o", "out.s", "--target=arm64", "--optimize=3", "main.c"},
			wantErr: false,
			verify: func(o *CompileOptions) bool {
				return o.InputFile == "main.c" &&
					o.Output == "out.s" &&
					o.EmitAssembly &&
					o.Verbose &&
					o.Debug &&
					o.Target == "arm64" &&
					o.Optimize == "3"
			},
		},
		{
			name: "mixed short and long flags",
			args: []string{"-d", "--verbose", "-S", "-o", "out.s", "--target=x86_64", "-O", "s", "main.c"},
			wantErr: false,
			verify: func(o *CompileOptions) bool {
				return o.InputFile == "main.c" &&
					o.Output == "out.s" &&
					o.EmitAssembly &&
					o.Verbose &&
					o.Debug &&
					o.Target == "x86_64" &&
					o.Optimize == "s"
			},
		},
		{
			name:    "backward compatibility - old flags only",
			args:    []string{"-v", "-S", "-c", "-o", "out.s", "main.c"},
			wantErr: false,
			verify: func(o *CompileOptions) bool {
				return o.InputFile == "main.c" &&
					o.Output == "out.s" &&
					o.EmitAssembly &&
					o.EmitObject &&
					o.Verbose &&
					!o.Debug &&
					o.Target == "" &&
					o.Optimize == ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseCompileFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCompileFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.verify(opts) {
				t.Errorf("ParseCompileFlags() verification failed for %+v", opts)
			}
		})
	}
}

func TestParseCompileFlags_Errors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "unknown flag",
			args: []string{"--unknown", "main.c"},
		},
		{
			name: "-o missing argument",
			args: []string{"-o"},
		},
		{
			name: "multiple input files",
			args: []string{"main.c", "other.c"},
		},
		{
			name: "invalid optimize level",
			args: []string{"-O", "5", "main.c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCompileFlags(tt.args)
			if err == nil {
				t.Errorf("ParseCompileFlags() expected error for args %v, got nil", tt.args)
			}
		})
	}
}

func TestIsValidOptimizeLevel(t *testing.T) {
	tests := []struct {
		level string
		want  bool
	}{
		{"0", true},
		{"1", true},
		{"2", true},
		{"3", true},
		{"s", true},
		{"z", true},
		{"4", false},
		{"invalid", false},
		{"", false},
		{"O", false},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			got := isValidOptimizeLevel(tt.level)
			if got != tt.want {
				t.Errorf("isValidOptimizeLevel(%q) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

func TestValidateOptimizeLevel(t *testing.T) {
	tests := []struct {
		level   string
		wantErr bool
	}{
		{"", false},    // Empty is valid (no optimization)
		{"0", false},
		{"1", false},
		{"2", false},
		{"3", false},
		{"s", false},
		{"z", false},
		{"4", true},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			err := validateOptimizeLevel(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateOptimizeLevel(%q) error = %v, wantErr %v", tt.level, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTarget(t *testing.T) {
	tests := []struct {
		target  string
		wantErr bool
	}{
		{"", false},         // Empty is valid (default target)
		{"x86_64", false},
		{"arm64", false},
		{"x86", false},
		{"arm", false},
		{"riscv64", false},
		{"invalid", true},
		{"x86-64", true},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			err := validateTarget(tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTarget(%q) error = %v, wantErr %v", tt.target, err, tt.wantErr)
			}
		})
	}
}