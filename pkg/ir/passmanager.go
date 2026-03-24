// Package ir provides intermediate representation for the GOC compiler.
// This file defines the PassManager for orchestrating optimization passes.
package ir

import (
	"fmt"
	"sort"
)

// PassManager orchestrates the execution of optimization passes.
type PassManager struct {
	// passes is the list of passes to run in order.
	passes []Pass
	// enabled indicates if the pass manager is enabled.
	enabled bool
	// results stores results from pass executions.
	results []PassResult
	// registry is the pass registry for creating passes.
	registry *PassRegistry
}

// PassManagerConfig holds configuration for the PassManager.
type PassManagerConfig struct {
	// Enabled indicates if optimization is enabled (default: false).
	Enabled bool
	// PassNames is the list of pass names to run in order.
	PassNames []string
	// Registry is the pass registry to use (optional, uses global if nil).
	Registry *PassRegistry
}

// NewPassManager creates a new PassManager with the given configuration.
func NewPassManager(config PassManagerConfig) (*PassManager, error) {
	pm := &PassManager{
		passes:   make([]Pass, 0),
		enabled:  config.Enabled,
		results:  make([]PassResult, 0),
		registry: config.Registry,
	}

	// Use global registry if not specified
	if pm.registry == nil {
		pm.registry = globalPassRegistry
	}

	// Create passes from names
	for _, name := range config.PassNames {
		pass, err := pm.registry.Create(name)
		if err != nil {
			return nil, fmt.Errorf("failed to create pass %s: %w", name, err)
		}
		pm.passes = append(pm.passes, pass)
	}

	return pm, nil
}

// NewPassManagerWithPasses creates a PassManager with explicit pass instances.
func NewPassManagerWithPasses(enabled bool, passes ...Pass) *PassManager {
	return &PassManager{
		passes:   passes,
		enabled:  enabled,
		results:  make([]PassResult, 0),
		registry: globalPassRegistry,
	}
}

// AddPass adds a pass to the manager.
func (pm *PassManager) AddPass(pass Pass) {
	pm.passes = append(pm.passes, pass)
}

// AddPassByName adds a pass by name from the registry.
func (pm *PassManager) AddPassByName(name string) error {
	pass, err := pm.registry.Create(name)
	if err != nil {
		return fmt.Errorf("failed to create pass %s: %w", name, err)
	}
	pm.passes = append(pm.passes, pass)
	return nil
}

// RemovePass removes a pass by name.
func (pm *PassManager) RemovePass(name string) bool {
	for i, pass := range pm.passes {
		if pass.Info().Name == name {
			pm.passes = append(pm.passes[:i], pm.passes[i+1:]...)
			return true
		}
	}
	return false
}

// ClearPasses removes all passes from the manager.
func (pm *PassManager) ClearPasses() {
	pm.passes = make([]Pass, 0)
	pm.results = make([]PassResult, 0)
}

// SetEnabled enables or disables the pass manager.
func (pm *PassManager) SetEnabled(enabled bool) {
	pm.enabled = enabled
}

// IsEnabled returns whether the pass manager is enabled.
func (pm *PassManager) IsEnabled() bool {
	return pm.enabled
}

// Run executes all registered passes on the IR.
// Returns true if any pass modified the IR.
func (pm *PassManager) Run(ir *IR) (bool, error) {
	if !pm.enabled {
		return false, nil
	}

	if ir == nil {
		return false, fmt.Errorf("nil IR")
	}

	pm.results = make([]PassResult, 0, len(pm.passes))
	anyModified := false

	for _, pass := range pm.passes {
		result, err := pm.runPass(pass, ir)
		if err != nil {
			return anyModified, fmt.Errorf("pass %s failed: %w", pass.Info().Name, err)
		}

		pm.results = append(pm.results, result)
		if result.Modified {
			anyModified = true
		}
	}

	return anyModified, nil
}

// runPass executes a single pass and returns the result.
func (pm *PassManager) runPass(pass Pass, ir *IR) (PassResult, error) {
	modified, err := pass.Run(ir)
	if err != nil {
		return PassResult{
			PassName: pass.Info().Name,
			Modified: false,
			Message:  fmt.Sprintf("error: %v", err),
		}, err
	}

	return PassResult{
		PassName: pass.Info().Name,
		Modified: modified,
		Message:  "completed successfully",
	}, nil
}

// RunPhase executes all passes in a specific phase.
func (pm *PassManager) RunPhase(ir *IR, phase PassPhase) (bool, error) {
	if !pm.enabled {
		return false, nil
	}

	if ir == nil {
		return false, fmt.Errorf("nil IR")
	}

	phasePasses := pm.getPassesByPhase(phase)
	anyModified := false

	for _, pass := range phasePasses {
		result, err := pm.runPass(pass, ir)
		if err != nil {
			return anyModified, fmt.Errorf("pass %s failed: %w", pass.Info().Name, err)
		}

		pm.results = append(pm.results, result)
		if result.Modified {
			anyModified = true
		}
	}

	return anyModified, nil
}

// getPassesByPhase returns passes for a specific phase in dependency order.
func (pm *PassManager) getPassesByPhase(phase PassPhase) []Pass {
	var phasePasses []Pass
	for _, pass := range pm.passes {
		if pass.Info().Phase == phase {
			phasePasses = append(phasePasses, pass)
		}
	}

	// Sort by dependencies
	return pm.sortPassesByDeps(phasePasses)
}

// sortPassesByDeps sorts passes based on their dependencies.
func (pm *PassManager) sortPassesByDeps(passes []Pass) []Pass {
	// Build dependency graph
	depCount := make(map[string]int)
	depMap := make(map[string][]string)

	for _, pass := range passes {
		info := pass.Info()
		depCount[info.Name] = 0
		for _, dep := range info.Dependencies {
			depMap[dep.Name] = append(depMap[dep.Name], info.Name)
		}
	}

	// Count dependencies for each pass
	for _, pass := range passes {
		info := pass.Info()
		for _, dep := range info.Dependencies {
			if _, exists := depCount[dep.Name]; exists {
				depCount[info.Name]++
			}
		}
	}

	// Topological sort
	sorted := make([]Pass, 0, len(passes))
	remaining := make(map[string]Pass)
	for _, pass := range passes {
		remaining[pass.Info().Name] = pass
	}

	for len(remaining) > 0 {
		// Find pass with no remaining dependencies
		var next Pass
		for name, pass := range remaining {
			if depCount[name] == 0 {
				next = pass
				break
			}
		}

		if next == nil {
			// Circular dependency or all remaining have deps
			// Just add remaining passes in original order
			for _, pass := range passes {
				if _, exists := remaining[pass.Info().Name]; exists {
					sorted = append(sorted, pass)
				}
			}
			break
		}

		sorted = append(sorted, next)
		delete(remaining, next.Info().Name)

		// Decrease dependency count for dependent passes
		for _, depName := range depMap[next.Info().Name] {
			depCount[depName]--
		}
	}

	return sorted
}

// GetResults returns the results from the last run.
func (pm *PassManager) GetResults() []PassResult {
	return pm.results
}

// GetPassCount returns the number of registered passes.
func (pm *PassManager) GetPassCount() int {
	return len(pm.passes)
}

// ListPasses returns information about all registered passes.
func (pm *PassManager) ListPasses() []PassInfo {
	infos := make([]PassInfo, 0, len(pm.passes))
	for _, pass := range pm.passes {
		infos = append(infos, pass.Info())
	}
	return infos
}

// Reset resets all passes in the manager.
func (pm *PassManager) Reset() {
	for _, pass := range pm.passes {
		pass.Reset()
	}
	pm.results = make([]PassResult, 0)
}

// OptimizationConfig holds configuration for IR optimization.
type OptimizationConfig struct {
	// Enabled indicates if optimization is enabled (default: false).
	Enabled bool
	// Passes is the list of pass names to run.
	Passes []string
	// Verbose enables verbose output.
	Verbose bool
}

// DefaultOptimizationConfig returns the default optimization configuration.
func DefaultOptimizationConfig() OptimizationConfig {
	return OptimizationConfig{
		Enabled: false, // Optimization is off by default
		Passes:  []string{},
		Verbose: false,
	}
}

// CreateDefaultPassManager creates a PassManager with default configuration.
func CreateDefaultPassManager() *PassManager {
	return NewPassManagerWithPasses(false)
}

// CreateOptimizedPassManager creates a PassManager with standard optimization passes.
func CreateOptimizedPassManager(passes []string) (*PassManager, error) {
	config := PassManagerConfig{
		Enabled:   true,
		PassNames: passes,
	}
	return NewPassManager(config)
}

// SortPassesByPhase sorts passes by their phase and dependencies.
func SortPassesByPhase(passes []Pass) []Pass {
	// Group by phase
	phaseGroups := make(map[PassPhase][]Pass)
	for _, pass := range passes {
		phase := pass.Info().Phase
		phaseGroups[phase] = append(phaseGroups[phase], pass)
	}

	// Sort each phase by dependencies
	sorted := make([]Pass, 0, len(passes))
	for phase := PassPhaseEarly; phase <= PassPhaseLate; phase++ {
		if group, exists := phaseGroups[phase]; exists {
			// Sort within phase by name for deterministic order
			sort.Slice(group, func(i, j int) bool {
				return group[i].Info().Name < group[j].Info().Name
			})
			sorted = append(sorted, group...)
		}
	}

	return sorted
}