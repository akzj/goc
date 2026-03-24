// Package ir provides intermediate representation for the GOC compiler.
// This file provides integration between IR generation and optimization.
package ir

import (
	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/parser"
)

// OptimizingIRGenerator wraps IRGenerator with optimization support.
type OptimizingIRGenerator struct {
	// generator is the underlying IR generator.
	generator *IRGenerator
	// passManager manages optimization passes.
	passManager *PassManager
	// config holds optimization configuration.
	config OptimizationConfig
}

// OptimizingIRGeneratorConfig holds configuration for the optimizing generator.
type OptimizingIRGeneratorConfig struct {
	// ErrorHandler is the error handler to use.
	ErrorHandler *errhand.ErrorHandler
	// Optimization holds optimization configuration.
	Optimization OptimizationConfig
}

// NewOptimizingIRGenerator creates a new optimizing IR generator.
func NewOptimizingIRGenerator(config OptimizingIRGeneratorConfig) (*OptimizingIRGenerator, error) {
	// Create base generator
	generator := NewIRGenerator(config.ErrorHandler)

	// Create pass manager
	passManager := CreateDefaultPassManager()

	oig := &OptimizingIRGenerator{
		generator:   generator,
		passManager: passManager,
		config:      config.Optimization,
	}

	// Configure optimization if enabled
	if config.Optimization.Enabled {
		if err := oig.configureOptimization(config.Optimization.Passes); err != nil {
			return nil, err
		}
	}

	return oig, nil
}

// configureOptimization sets up optimization passes.
func (oig *OptimizingIRGenerator) configureOptimization(passNames []string) error {
	if len(passNames) == 0 {
		return nil
	}

	pm, err := CreateOptimizedPassManager(passNames)
	if err != nil {
		return err
	}

	oig.passManager = pm
	return nil
}

// Generate generates IR from the AST with optional optimization.
func (oig *OptimizingIRGenerator) Generate(ast *parser.TranslationUnit) (*IR, error) {
	// Generate base IR
	ir, err := oig.generator.Generate(ast)
	if err != nil {
		return nil, err
	}

	// Run optimization if enabled
	if oig.config.Enabled && oig.passManager != nil {
		_, err = oig.passManager.Run(ir)
		if err != nil {
			return nil, err
		}
	}

	return ir, nil
}

// GetPassManager returns the pass manager.
func (oig *OptimizingIRGenerator) GetPassManager() *PassManager {
	return oig.passManager
}

// SetOptimizationEnabled enables or disables optimization.
func (oig *OptimizingIRGenerator) SetOptimizationEnabled(enabled bool) {
	if oig.passManager != nil {
		oig.passManager.SetEnabled(enabled)
	}
	oig.config.Enabled = enabled
}

// IsOptimizationEnabled returns whether optimization is enabled.
func (oig *OptimizingIRGenerator) IsOptimizationEnabled() bool {
	return oig.config.Enabled && oig.passManager != nil && oig.passManager.IsEnabled()
}

// AddOptimizationPass adds an optimization pass.
func (oig *OptimizingIRGenerator) AddOptimizationPass(pass Pass) {
	if oig.passManager != nil {
		oig.passManager.AddPass(pass)
	}
}

// GetOptimizationResults returns results from the last optimization run.
func (oig *OptimizingIRGenerator) GetOptimizationResults() []PassResult {
	if oig.passManager == nil {
		return nil
	}
	return oig.passManager.GetResults()
}

// IRGeneratorWithOptimization provides a simple interface for IR generation
// with optional optimization support.
type IRGeneratorWithOptimization struct {
	*IRGenerator
	passManager *PassManager
	enabled     bool
}

// NewIRGeneratorWithOptimization creates a new IR generator with optimization.
func NewIRGeneratorWithOptimization(
	errorHandler *errhand.ErrorHandler,
	enabled bool,
	passes []string,
) (*IRGeneratorWithOptimization, error) {
	gen := NewIRGenerator(errorHandler)

	var pm *PassManager
	if enabled && len(passes) > 0 {
		config := PassManagerConfig{
			Enabled:   true,
			PassNames: passes,
		}
		var err error
		pm, err = NewPassManager(config)
		if err != nil {
			return nil, err
		}
	} else if enabled {
		pm = CreateDefaultPassManager()
		pm.SetEnabled(true)
	}

	return &IRGeneratorWithOptimization{
		IRGenerator: gen,
		passManager: pm,
		enabled:     enabled,
	}, nil
}

// Generate generates IR with optional optimization.
func (g *IRGeneratorWithOptimization) Generate(ast *parser.TranslationUnit) (*IR, error) {
	ir, err := g.IRGenerator.Generate(ast)
	if err != nil {
		return nil, err
	}

	if g.enabled && g.passManager != nil {
		_, err = g.passManager.Run(ir)
		if err != nil {
			return nil, err
		}
	}

	return ir, nil
}

// EnableOptimization enables optimization.
func (g *IRGeneratorWithOptimization) EnableOptimization() {
	g.enabled = true
	if g.passManager != nil {
		g.passManager.SetEnabled(true)
	}
}

// DisableOptimization disables optimization.
func (g *IRGeneratorWithOptimization) DisableOptimization() {
	g.enabled = false
	if g.passManager != nil {
		g.passManager.SetEnabled(false)
	}
}

// ApplyOptimization applies optimization to an existing IR.
func ApplyOptimization(ir *IR, passes []Pass) (bool, error) {
	if ir == nil {
		return false, nil
	}

	pm := NewPassManagerWithPasses(true, passes...)
	return pm.Run(ir)
}

// ApplyOptimizationByName applies optimization using pass names.
func ApplyOptimizationByName(ir *IR, passNames []string) (bool, error) {
	if ir == nil {
		return false, nil
	}

	if len(passNames) == 0 {
		return false, nil
	}

	config := PassManagerConfig{
		Enabled:   true,
		PassNames: passNames,
	}
	pm, err := NewPassManager(config)
	if err != nil {
		return false, err
	}

	return pm.Run(ir)
}