// Package ir provides intermediate representation for the GOC compiler.
// This file defines the optimization pass interface and framework.
package ir

import (
	"fmt"
)

// PassPhase defines the phase in which a pass should run.
type PassPhase int

const (
	// PassPhaseEarly runs before most optimizations.
	PassPhaseEarly PassPhase = iota
	// PassPhaseMain runs during main optimization phase.
	PassPhaseMain
	// PassPhaseLate runs after most optimizations.
	PassPhaseLate
)

// PassDependency defines a dependency relationship between passes.
type PassDependency struct {
	// Name is the name of the required pass.
	Name string
	// Required indicates if this dependency is mandatory.
	Required bool
}

// PassInfo contains metadata about an optimization pass.
type PassInfo struct {
	// Name is the unique identifier for the pass.
	Name string
	// Description provides a brief description of what the pass does.
	Description string
	// Phase indicates when this pass should run.
	Phase PassPhase
	// Dependencies lists passes that must run before this one.
	Dependencies []PassDependency
	// Enabled indicates if the pass is enabled by default.
	Enabled bool
}

// Pass defines the interface for IR optimization passes.
type Pass interface {
	// Info returns metadata about the pass.
	Info() PassInfo

	// Run executes the pass on the given IR.
	// Returns true if the IR was modified, false otherwise.
	Run(ir *IR) (bool, error)

	// Reset resets the pass state for reuse.
	Reset()
}

// BasePass provides common functionality for passes.
type BasePass struct {
	info PassInfo
}

// NewBasePass creates a new base pass with the given info.
func NewBasePass(info PassInfo) BasePass {
	return BasePass{info: info}
}

// Info returns the pass information.
func (bp *BasePass) Info() PassInfo {
	return bp.info
}

// Reset resets the pass state (no-op for base pass).
func (bp *BasePass) Reset() {
	// No state to reset in base pass
}

// PassResult contains the result of running a pass.
type PassResult struct {
	// Modified indicates if the IR was changed.
	Modified bool
	// PassName is the name of the pass that was run.
	PassName string
	// Message contains any informational message.
	Message string
}

// String returns a string representation of the pass result.
func (pr *PassResult) String() string {
	status := "unchanged"
	if pr.Modified {
		status = "modified"
	}
	return fmt.Sprintf("Pass %s: %s - %s", pr.PassName, status, pr.Message)
}

// PassRegistry maintains a registry of available passes.
type PassRegistry struct {
	passes map[string]func() Pass
}

// NewPassRegistry creates a new pass registry.
func NewPassRegistry() *PassRegistry {
	return &PassRegistry{
		passes: make(map[string]func() Pass),
	}
}

// Register registers a pass constructor with the registry.
func (pr *PassRegistry) Register(name string, constructor func() Pass) {
	pr.passes[name] = constructor
}

// Get retrieves a pass constructor by name.
func (pr *PassRegistry) Get(name string) (func() Pass, bool) {
	constructor, ok := pr.passes[name]
	return constructor, ok
}

// List returns all registered pass names.
func (pr *PassRegistry) List() []string {
	names := make([]string, 0, len(pr.passes))
	for name := range pr.passes {
		names = append(names, name)
	}
	return names
}

// Create creates a new instance of a registered pass.
func (pr *PassRegistry) Create(name string) (Pass, error) {
	constructor, ok := pr.passes[name]
	if !ok {
		return nil, fmt.Errorf("pass not found: %s", name)
	}
	return constructor(), nil
}

// Global registry for all passes.
var globalPassRegistry = NewPassRegistry()

// GetGlobalPassRegistry returns the global pass registry.
func GetGlobalPassRegistry() *PassRegistry {
	return globalPassRegistry
}

// RegisterPass registers a pass with the global registry.
func RegisterPass(name string, constructor func() Pass) {
	globalPassRegistry.Register(name, constructor)
}