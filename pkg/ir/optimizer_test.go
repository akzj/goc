// Package ir provides intermediate representation for the GOC compiler.
// This file contains unit tests for the optimization integration.
package ir

import (
	"testing"

	"github.com/akzj/goc/internal/errhand"
)

// TestOptimizingIRGeneratorCreation tests creating the optimizing generator.
func TestOptimizingIRGeneratorCreation(t *testing.T) {
	// Test with default config (optimization disabled)
	config := OptimizingIRGeneratorConfig{
		ErrorHandler: errhand.NewErrorHandler(),
		Optimization: DefaultOptimizationConfig(),
	}

	gen, err := NewOptimizingIRGenerator(config)
	if err != nil {
		t.Errorf("Failed to create generator: %v", err)
	}
	if gen == nil {
		t.Error("Expected non-nil generator")
	}
	if gen.IsOptimizationEnabled() {
		t.Error("Expected optimization to be disabled")
	}
}

// TestOptimizingIRGeneratorWithOptimization tests creating with optimization enabled.
func TestOptimizingIRGeneratorWithOptimization(t *testing.T) {
	// Register a test pass
	RegisterPass("test-opt-pass", func() Pass {
		return NewMockPass("test-opt-pass", false)
	})

	config := OptimizingIRGeneratorConfig{
		ErrorHandler: errhand.NewErrorHandler(),
		Optimization: OptimizationConfig{
			Enabled: true,
			Passes:  []string{"test-opt-pass"},
		},
	}

	gen, err := NewOptimizingIRGenerator(config)
	if err != nil {
		t.Errorf("Failed to create generator: %v", err)
	}
	if gen == nil {
		t.Error("Expected non-nil generator")
	}
	if !gen.IsOptimizationEnabled() {
		t.Error("Expected optimization to be enabled")
	}
}

// TestOptimizingIRGeneratorSetEnabled tests enabling/disabling optimization.
func TestOptimizingIRGeneratorSetEnabled(t *testing.T) {
	config := OptimizingIRGeneratorConfig{
		ErrorHandler: errhand.NewErrorHandler(),
		Optimization: DefaultOptimizationConfig(),
	}

	gen, err := NewOptimizingIRGenerator(config)
	if err != nil {
		t.Errorf("Failed to create generator: %v", err)
	}

	if gen.IsOptimizationEnabled() {
		t.Error("Expected optimization to be disabled initially")
	}

	gen.SetOptimizationEnabled(true)
	if !gen.IsOptimizationEnabled() {
		t.Error("Expected optimization to be enabled")
	}

	gen.SetOptimizationEnabled(false)
	if gen.IsOptimizationEnabled() {
		t.Error("Expected optimization to be disabled")
	}
}

// TestOptimizingIRGeneratorAddPass tests adding optimization passes.
func TestOptimizingIRGeneratorAddPass(t *testing.T) {
	config := OptimizingIRGeneratorConfig{
		ErrorHandler: errhand.NewErrorHandler(),
		Optimization: DefaultOptimizationConfig(),
	}

	gen, err := NewOptimizingIRGenerator(config)
	if err != nil {
		t.Errorf("Failed to create generator: %v", err)
	}

	pass := NewMockPass("added-pass", false)
	gen.AddOptimizationPass(pass)

	pm := gen.GetPassManager()
	if pm == nil {
		t.Error("Expected non-nil pass manager")
	}
	if pm.GetPassCount() != 1 {
		t.Errorf("Expected 1 pass, got %d", pm.GetPassCount())
	}
}

// TestOptimizingIRGeneratorGetResults tests getting optimization results.
func TestOptimizingIRGeneratorGetResults(t *testing.T) {
	config := OptimizingIRGeneratorConfig{
		ErrorHandler: errhand.NewErrorHandler(),
		Optimization: DefaultOptimizationConfig(),
	}

	gen, err := NewOptimizingIRGenerator(config)
	if err != nil {
		t.Errorf("Failed to create generator: %v", err)
	}

	// No results before running
	results := gen.GetOptimizationResults()
	if results != nil && len(results) > 0 {
		t.Errorf("Expected no results before running, got %d", len(results))
	}
}

// TestIRGeneratorWithOptimization tests the simple optimization wrapper.
func TestIRGeneratorWithOptimization(t *testing.T) {
	// Test with optimization disabled
	gen, err := NewIRGeneratorWithOptimization(
		errhand.NewErrorHandler(),
		false,
		[]string{},
	)
	if err != nil {
		t.Errorf("Failed to create generator: %v", err)
	}
	if gen == nil {
		t.Error("Expected non-nil generator")
	}

	// Test enable/disable
	gen.EnableOptimization()
	// Note: enabled flag is set, but passManager might be nil if no passes

	gen.DisableOptimization()
}

// TestIRGeneratorWithOptimizationWithPasses tests with specific passes.
func TestIRGeneratorWithOptimizationWithPasses(t *testing.T) {
	RegisterPass("opt-pass-1", func() Pass {
		return NewMockPass("opt-pass-1", false)
	})

	gen, err := NewIRGeneratorWithOptimization(
		errhand.NewErrorHandler(),
		true,
		[]string{"opt-pass-1"},
	)
	if err != nil {
		t.Errorf("Failed to create generator: %v", err)
	}
	if gen == nil {
		t.Error("Expected non-nil generator")
	}
}

// TestApplyOptimization tests the ApplyOptimization function.
func TestApplyOptimization(t *testing.T) {
	ir := &IR{
		Functions: make([]*Function, 0),
		Globals:   make([]*GlobalVar, 0),
		Constants: make([]*Constant, 0),
	}

	passes := []Pass{
		NewMockPass("apply-pass-1", false),
		NewMockPass("apply-pass-2", true),
	}

	modified, err := ApplyOptimization(ir, passes)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !modified {
		t.Error("Expected IR to be modified")
	}

	// Test with nil IR
	modified, err = ApplyOptimization(nil, passes)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if modified {
		t.Error("Expected no modification for nil IR")
	}

	// Test with empty passes
	modified, err = ApplyOptimization(ir, []Pass{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if modified {
		t.Error("Expected no modification with empty passes")
	}
}

// TestApplyOptimizationByName tests ApplyOptimizationByName.
func TestApplyOptimizationByName(t *testing.T) {
	ir := &IR{
		Functions: make([]*Function, 0),
		Globals:   make([]*GlobalVar, 0),
		Constants: make([]*Constant, 0),
	}

	RegisterPass("named-opt-1", func() Pass {
		return NewMockPass("named-opt-1", true)
	})

	modified, err := ApplyOptimizationByName(ir, []string{"named-opt-1"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !modified {
		t.Error("Expected IR to be modified")
	}

	// Test with nil IR
	modified, err = ApplyOptimizationByName(nil, []string{"named-opt-1"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if modified {
		t.Error("Expected no modification for nil IR")
	}

	// Test with empty pass names
	modified, err = ApplyOptimizationByName(ir, []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if modified {
		t.Error("Expected no modification with empty pass names")
	}

	// Test with non-existent pass
	_, err = ApplyOptimizationByName(ir, []string{"non-existent-pass"})
	if err == nil {
		t.Error("Expected error for non-existent pass")
	}
}

// TestOptimizingIRGeneratorConfigureOptimization tests configuration.
func TestOptimizingIRGeneratorConfigureOptimization(t *testing.T) {
	config := OptimizingIRGeneratorConfig{
		ErrorHandler: errhand.NewErrorHandler(),
		Optimization: DefaultOptimizationConfig(),
	}

	gen, err := NewOptimizingIRGenerator(config)
	if err != nil {
		t.Errorf("Failed to create generator: %v", err)
	}

	// Test configuring with empty passes
	err = gen.configureOptimization([]string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Register a pass and test configuring
	RegisterPass("config-test-pass", func() Pass {
		return NewMockPass("config-test-pass", false)
	})

	err = gen.configureOptimization([]string{"config-test-pass"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	pm := gen.GetPassManager()
	if pm == nil {
		t.Error("Expected non-nil pass manager")
	}
	if pm.GetPassCount() != 1 {
		t.Errorf("Expected 1 pass, got %d", pm.GetPassCount())
	}

	// Test configuring with non-existent pass
	err = gen.configureOptimization([]string{"non-existent"})
	if err == nil {
		t.Error("Expected error for non-existent pass")
	}
}

// TestOptimizationConfig tests optimization configuration.
func TestOptimizationConfig(t *testing.T) {
	// Test default config
	config := DefaultOptimizationConfig()
	if config.Enabled {
		t.Error("Expected default config to have optimization disabled")
	}

	// Test custom config
	customConfig := OptimizationConfig{
		Enabled: true,
		Passes:  []string{"pass1", "pass2"},
		Verbose: true,
	}

	if !customConfig.Enabled {
		t.Error("Expected custom config to have optimization enabled")
	}
	if len(customConfig.Passes) != 2 {
		t.Errorf("Expected 2 passes, got %d", len(customConfig.Passes))
	}
	if !customConfig.Verbose {
		t.Error("Expected custom config to have verbose enabled")
	}
}