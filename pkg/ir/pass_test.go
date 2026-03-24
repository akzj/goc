// Package ir provides intermediate representation for the GOC compiler.
// This file contains unit tests for the optimization pass framework.
package ir

import (
	"testing"
)

// TestPassInfo tests the PassInfo structure.
func TestPassInfo(t *testing.T) {
	info := PassInfo{
		Name:        "test-pass",
		Description: "A test pass",
		Phase:       PassPhaseMain,
		Enabled:     true,
	}

	if info.Name != "test-pass" {
		t.Errorf("Expected name 'test-pass', got '%s'", info.Name)
	}
	if info.Description != "A test pass" {
		t.Errorf("Expected description 'A test pass', got '%s'", info.Description)
	}
	if info.Phase != PassPhaseMain {
		t.Errorf("Expected phase PassPhaseMain, got %d", info.Phase)
	}
	if !info.Enabled {
		t.Error("Expected pass to be enabled")
	}
}

// TestBasePass tests the BasePass implementation.
func TestBasePass(t *testing.T) {
	info := PassInfo{
		Name:        "base-pass",
		Description: "Base pass test",
		Phase:       PassPhaseEarly,
		Enabled:     false,
	}

	basePass := NewBasePass(info)

	// Test Info
	retrievedInfo := basePass.Info()
	if retrievedInfo.Name != info.Name {
		t.Errorf("Expected name '%s', got '%s'", info.Name, retrievedInfo.Name)
	}

	// Test Reset (should not panic)
	basePass.Reset()
}

// MockPass is a mock pass for testing.
type MockPass struct {
	BasePass
	runCount int
	modified bool
}

// NewMockPass creates a new mock pass.
func NewMockPass(name string, modified bool) *MockPass {
	return &MockPass{
		BasePass: NewBasePass(PassInfo{
			Name:        name,
			Description: "Mock pass for testing",
			Phase:       PassPhaseMain,
			Enabled:     true,
		}),
		modified: modified,
	}
}

// Run executes the mock pass.
func (mp *MockPass) Run(ir *IR) (bool, error) {
	mp.runCount++
	return mp.modified, nil
}

// Reset resets the mock pass state.
func (mp *MockPass) Reset() {
	mp.runCount = 0
}

// TestPassRegistry tests the PassRegistry functionality.
func TestPassRegistry(t *testing.T) {
	registry := NewPassRegistry()

	// Test empty registry
	names := registry.List()
	if len(names) != 0 {
		t.Errorf("Expected empty registry, got %d passes", len(names))
	}

	// Test registration
	registry.Register("mock-pass", func() Pass {
		return NewMockPass("mock-pass", false)
	})

	// Test listing
	names = registry.List()
	if len(names) != 1 {
		t.Errorf("Expected 1 pass, got %d", len(names))
	}

	// Test retrieval
	_, ok := registry.Get("mock-pass")
	if !ok {
		t.Error("Expected to find mock-pass")
	}

	// Test creation
	pass, err := registry.Create("mock-pass")
	if err != nil {
		t.Errorf("Failed to create pass: %v", err)
	}
	if pass == nil {
		t.Error("Expected non-nil pass")
	}
	if pass.Info().Name != "mock-pass" {
		t.Errorf("Expected name 'mock-pass', got '%s'", pass.Info().Name)
	}

	// Test non-existent pass
	_, ok = registry.Get("non-existent")
	if ok {
		t.Error("Expected not to find non-existent pass")
	}

	_, err = registry.Create("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent pass")
	}
}

// TestPassManagerCreation tests PassManager creation.
func TestPassManagerCreation(t *testing.T) {
	// Test with config
	config := PassManagerConfig{
		Enabled:   true,
		PassNames: []string{},
	}
	pm, err := NewPassManager(config)
	if err != nil {
		t.Errorf("Failed to create PassManager: %v", err)
	}
	if pm == nil {
		t.Error("Expected non-nil PassManager")
	}
	if !pm.IsEnabled() {
		t.Error("Expected PassManager to be enabled")
	}

	// Test with explicit passes
	pass1 := NewMockPass("pass1", false)
	pass2 := NewMockPass("pass2", true)
	pm2 := NewPassManagerWithPasses(true, pass1, pass2)
	if pm2.GetPassCount() != 2 {
		t.Errorf("Expected 2 passes, got %d", pm2.GetPassCount())
	}
}

// TestPassManagerAddRemove tests adding and removing passes.
func TestPassManagerAddRemove(t *testing.T) {
	pm := NewPassManagerWithPasses(false)

	// Test AddPass
	pass := NewMockPass("test", false)
	pm.AddPass(pass)
	if pm.GetPassCount() != 1 {
		t.Errorf("Expected 1 pass, got %d", pm.GetPassCount())
	}

	// Test AddPassByName
	registry := NewPassRegistry()
	registry.Register("named-pass", func() Pass {
		return NewMockPass("named-pass", false)
	})
	pm2 := &PassManager{
		passes:   make([]Pass, 0),
		enabled:  false,
		results:  make([]PassResult, 0),
		registry: registry,
	}

	err := pm2.AddPassByName("named-pass")
	if err != nil {
		t.Errorf("Failed to add pass by name: %v", err)
	}
	if pm2.GetPassCount() != 1 {
		t.Errorf("Expected 1 pass, got %d", pm2.GetPassCount())
	}

	// Test RemovePass
	removed := pm.RemovePass("test")
	if !removed {
		t.Error("Expected to remove pass")
	}
	if pm.GetPassCount() != 0 {
		t.Errorf("Expected 0 passes, got %d", pm.GetPassCount())
	}

	// Test RemovePass non-existent
	removed = pm.RemovePass("non-existent")
	if removed {
		t.Error("Expected not to remove non-existent pass")
	}

	// Test ClearPasses
	pm.AddPass(NewMockPass("pass1", false))
	pm.AddPass(NewMockPass("pass2", false))
	pm.ClearPasses()
	if pm.GetPassCount() != 0 {
		t.Errorf("Expected 0 passes after clear, got %d", pm.GetPassCount())
	}
}

// TestPassManagerRun tests running passes.
func TestPassManagerRun(t *testing.T) {
	// Test with disabled manager
	pm := NewPassManagerWithPasses(false, NewMockPass("test", true))
	modified, err := pm.Run(&IR{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if modified {
		t.Error("Expected no modification when disabled")
	}

	// Test with enabled manager
	pass1 := NewMockPass("pass1", false)
	pass2 := NewMockPass("pass2", true)
	pm2 := NewPassManagerWithPasses(true, pass1, pass2)

	modified, err = pm2.Run(&IR{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !modified {
		t.Error("Expected modification")
	}

	// Check results
	results := pm2.GetResults()
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if results[0].PassName != "pass1" {
		t.Errorf("Expected first pass name 'pass1', got '%s'", results[0].PassName)
	}
	if results[1].PassName != "pass2" {
		t.Errorf("Expected second pass name 'pass2', got '%s'", results[1].PassName)
	}

	// Test with nil IR
	modified, err = pm2.Run(nil)
	if err == nil {
		t.Error("Expected error for nil IR")
	}
	if modified {
		t.Error("Expected no modification for nil IR")
	}
}

// TestPassManagerSetEnabled tests enabling/disabling the manager.
func TestPassManagerSetEnabled(t *testing.T) {
	pm := NewPassManagerWithPasses(false)

	if pm.IsEnabled() {
		t.Error("Expected manager to be disabled initially")
	}

	pm.SetEnabled(true)
	if !pm.IsEnabled() {
		t.Error("Expected manager to be enabled")
	}

	pm.SetEnabled(false)
	if pm.IsEnabled() {
		t.Error("Expected manager to be disabled")
	}
}

// TestPassResult tests PassResult functionality.
func TestPassResult(t *testing.T) {
	result := PassResult{
		Modified: true,
		PassName: "test-pass",
		Message:  "success",
	}

	str := result.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}

	result2 := PassResult{
		Modified: false,
		PassName: "test-pass",
		Message:  "no changes",
	}
	str2 := result2.String()
	if str2 == "" {
		t.Error("Expected non-empty string representation")
	}
}

// TestPassManagerReset tests resetting the manager.
func TestPassManagerReset(t *testing.T) {
	pass := NewMockPass("test", true)
	pass.runCount = 5 // Simulate some runs

	pm := NewPassManagerWithPasses(true, pass)
	pm.results = append(pm.results, PassResult{PassName: "test", Modified: true})

	pm.Reset()

	if pass.runCount != 0 {
		t.Errorf("Expected pass runCount to be 0, got %d", pass.runCount)
	}
	if len(pm.results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(pm.results))
	}
}

// TestPassManagerListPasses tests listing passes.
func TestPassManagerListPasses(t *testing.T) {
	pm := NewPassManagerWithPasses(false)
	pm.AddPass(NewMockPass("pass1", false))
	pm.AddPass(NewMockPass("pass2", true))

	infos := pm.ListPasses()
	if len(infos) != 2 {
		t.Errorf("Expected 2 pass infos, got %d", len(infos))
	}

	found := make(map[string]bool)
	for _, info := range infos {
		found[info.Name] = true
	}
	if !found["pass1"] || !found["pass2"] {
		t.Error("Expected to find both passes")
	}
}

// TestSortPassesByPhase tests sorting passes by phase.
func TestSortPassesByPhase(t *testing.T) {
	passes := []Pass{
		NewMockPass("late-pass", false),
		NewMockPass("early-pass", false),
		NewMockPass("main-pass", false),
	}

	// Set phases
	passes[0].(*MockPass).BasePass = NewBasePass(PassInfo{
		Name:  "late-pass",
		Phase: PassPhaseLate,
	})
	passes[1].(*MockPass).BasePass = NewBasePass(PassInfo{
		Name:  "early-pass",
		Phase: PassPhaseEarly,
	})
	passes[2].(*MockPass).BasePass = NewBasePass(PassInfo{
		Name:  "main-pass",
		Phase: PassPhaseMain,
	})

	sorted := SortPassesByPhase(passes)
	if len(sorted) != 3 {
		t.Errorf("Expected 3 sorted passes, got %d", len(sorted))
	}

	// Check order: early, main, late
	if sorted[0].Info().Name != "early-pass" {
		t.Errorf("Expected first pass 'early-pass', got '%s'", sorted[0].Info().Name)
	}
	if sorted[1].Info().Name != "main-pass" {
		t.Errorf("Expected second pass 'main-pass', got '%s'", sorted[1].Info().Name)
	}
	if sorted[2].Info().Name != "late-pass" {
		t.Errorf("Expected third pass 'late-pass', got '%s'", sorted[2].Info().Name)
	}
}

// TestDefaultOptimizationConfig tests default configuration.
func TestDefaultOptimizationConfig(t *testing.T) {
	config := DefaultOptimizationConfig()

	if config.Enabled {
		t.Error("Expected optimization to be disabled by default")
	}
	if len(config.Passes) != 0 {
		t.Errorf("Expected no default passes, got %d", len(config.Passes))
	}
	if config.Verbose {
		t.Error("Expected verbose to be false by default")
	}
}

// TestCreateDefaultPassManager tests creating default manager.
func TestCreateDefaultPassManager(t *testing.T) {
	pm := CreateDefaultPassManager()

	if pm == nil {
		t.Error("Expected non-nil PassManager")
	}
	if pm.IsEnabled() {
		t.Error("Expected default manager to be disabled")
	}
	if pm.GetPassCount() != 0 {
		t.Errorf("Expected 0 passes, got %d", pm.GetPassCount())
	}
}

// TestGlobalPassRegistry tests global registry functions.
func TestGlobalPassRegistry(t *testing.T) {
	// Get global registry
	registry := GetGlobalPassRegistry()
	if registry == nil {
		t.Error("Expected non-nil global registry")
	}

	// Test RegisterPass
	initialCount := len(registry.List())
	RegisterPass("global-test-pass", func() Pass {
		return NewMockPass("global-test-pass", false)
	})

	newCount := len(registry.List())
	if newCount != initialCount+1 {
		t.Errorf("Expected %d passes, got %d", initialCount+1, newCount)
	}
}

// TestPassManagerRunPhase tests running passes by phase.
func TestPassManagerRunPhase(t *testing.T) {
	pass1 := NewMockPass("early", false)
	pass1.BasePass = NewBasePass(PassInfo{
		Name:  "early",
		Phase: PassPhaseEarly,
	})

	pass2 := NewMockPass("main", true)
	pass2.BasePass = NewBasePass(PassInfo{
		Name:  "main",
		Phase: PassPhaseMain,
	})

	pass3 := NewMockPass("late", false)
	pass3.BasePass = NewBasePass(PassInfo{
		Name:  "late",
		Phase: PassPhaseLate,
	})

	pm := NewPassManagerWithPasses(true, pass1, pass2, pass3)

	// Run only Main phase
	modified, err := pm.RunPhase(&IR{}, PassPhaseMain)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !modified {
		t.Error("Expected modification from main pass")
	}

	// Run Early phase (should not modify)
	modified, err = pm.RunPhase(&IR{}, PassPhaseEarly)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if modified {
		t.Error("Expected no modification from early pass")
	}

	// Test with nil IR
	_, err = pm.RunPhase(nil, PassPhaseMain)
	if err == nil {
		t.Error("Expected error for nil IR")
	}

	// Test with disabled manager
	pm.SetEnabled(false)
	modified, err = pm.RunPhase(&IR{}, PassPhaseMain)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if modified {
		t.Error("Expected no modification when disabled")
	}
}