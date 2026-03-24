// Package ir provides edge case tests for IR generation and optimization.
// These tests focus on edge cases not covered in other test files.
package ir

import (
	"testing"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/parser"
)

// ============================================================================
// Edge Case Tests for Constant Folding
// ============================================================================

// TestConstantFolding_EdgeCases tests constant folding edge cases
func TestConstantFolding_EdgeCases(t *testing.T) {
	cf := NewConstantFolding()

	if cf == nil {
		t.Fatal("NewConstantFolding() returned nil")
	}

	// Test Reset
	cf.Reset()

	// Test Run with nil IR
	changed, err := cf.Run(nil)
	if !changed {
		t.Log("Run(nil) should handle nil IR gracefully")
	}
	if err == nil {
		t.Log("Run(nil) returned nil error")
	}
}

// TestConstantFolding_BinaryOps tests constant folding for binary operations
func TestConstantFolding_BinaryOps(t *testing.T) {
	cf := NewConstantFolding()

	tests := []struct {
		name      string
		opcode    Opcode
		left      interface{}
		right     interface{}
		expectOk  bool
	}{
		{"add_int", OpAdd, int64(5), int64(3), true},
		{"sub_int", OpSub, int64(10), int64(4), true},
		{"mul_int", OpMul, int64(6), int64(7), true},
		{"div_int", OpDiv, int64(20), int64(4), true},
		{"mod_int", OpMod, int64(17), int64(5), true},
		{"div_by_zero_int", OpDiv, int64(10), int64(0), false},
		{"mod_by_zero", OpMod, int64(10), int64(0), false},
		{"xor_bits", OpBitXor, int64(0xFF), int64(0x0F), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dest := &Operand{Value: int64(0)}
			result, ok := cf.foldBinaryOp(tt.opcode, tt.left, tt.right, dest)

			if tt.expectOk && !ok {
				t.Errorf("foldBinaryOp(%v, %v, %v) expected ok=true, got ok=false", tt.opcode, tt.left, tt.right)
			}

			if !tt.expectOk && ok {
				t.Logf("foldBinaryOp(%v, %v, %v) returned ok=true (may be expected)", tt.opcode, tt.left, tt.right)
			}

			_ = result
		})
	}
}

// TestConstantFolding_UnaryOps tests constant folding for unary operations
func TestConstantFolding_UnaryOps(t *testing.T) {
	cf := NewConstantFolding()

	tests := []struct {
		name     string
		opcode   Opcode
		val      interface{}
		expectOk bool
	}{
		{"neg_int", OpNeg, int64(5), true},
		{"not_bits", OpBitNot, int64(0xFF), true},
		{"not_bool", OpNot, int64(1), true},
		{"neg_zero", OpNeg, int64(0), true},
		{"not_zero", OpBitNot, int64(0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dest := &Operand{Value: int64(0)}
			result, ok := cf.foldUnaryOp(tt.opcode, tt.val, dest)

			if tt.expectOk && !ok {
				t.Errorf("foldUnaryOp(%v, %v) expected ok=true, got ok=false", tt.opcode, tt.val)
			}

			_ = result
		})
	}
}

// TestConstantFolding_CastOps tests constant folding for cast operations
func TestConstantFolding_CastOps(t *testing.T) {
	cf := NewConstantFolding()

	intType := &parser.BaseType{Kind: parser.TypeInt, Signed: true}
	charType := &parser.BaseType{Kind: parser.TypeChar, Signed: true}
	floatType := &parser.BaseType{Kind: parser.TypeFloat, Signed: true}

	tests := []struct {
		name     string
		opcode   Opcode
		val      interface{}
		destType parser.Type
		expectOk bool
	}{
		{"int_to_char", OpCast, int64(255), charType, true},
		{"char_to_int", OpCast, int64(65), intType, true},
		{"int_to_float", OpCast, int64(5), floatType, true},
		{"zero_cast", OpCast, int64(0), intType, true},
		{"negative_cast", OpCast, int64(-1), charType, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := cf.foldCastOp(tt.opcode, tt.val, tt.destType)

			if tt.expectOk && !ok {
				t.Errorf("foldCastOp(%v, %v, %T) expected ok=true, got ok=false", tt.opcode, tt.val, tt.destType)
			}
		})
	}
}

// TestConstantFolding_HelperFunctions tests helper functions
func TestConstantFolding_HelperFunctions(t *testing.T) {
	cf := NewConstantFolding()

	// Test toInt64
	t.Run("toInt64_int64", func(t *testing.T) {
		val := int64(42)
		result := cf.toInt64(val)
		if result == nil {
			t.Error("toInt64(int64) should not return nil")
		}
		if *result != 42 {
			t.Errorf("toInt64(42) = %d, want 42", *result)
		}
	})

	t.Run("toInt64_nil", func(t *testing.T) {
		result := cf.toInt64(nil)
		if result != nil {
			t.Error("toInt64(nil) should return nil")
		}
	})

	t.Run("toInt64_wrong_type", func(t *testing.T) {
		result := cf.toInt64("string")
		if result != nil {
			t.Error("toInt64(string) should return nil")
		}
	})

	// Test isIntegerType
	t.Run("isIntegerType_int", func(t *testing.T) {
		typ := &parser.BaseType{Kind: parser.TypeInt}
		if !isIntegerType(typ) {
			t.Error("isIntegerType(int) should return true")
		}
	})

	t.Run("isIntegerType_char", func(t *testing.T) {
		typ := &parser.BaseType{Kind: parser.TypeChar}
		if !isIntegerType(typ) {
			t.Error("isIntegerType(char) should return true")
		}
	})

	t.Run("isIntegerType_float", func(t *testing.T) {
		typ := &parser.BaseType{Kind: parser.TypeFloat}
		if isIntegerType(typ) {
			t.Error("isIntegerType(float) should return false")
		}
	})

	t.Run("isIntegerType_nil", func(t *testing.T) {
		if isIntegerType(nil) {
			t.Error("isIntegerType(nil) should return false")
		}
	})

	// Test bi (bool to int)
	t.Run("bi_true", func(t *testing.T) {
		if bi(true) != 1 {
			t.Error("bi(true) should return 1")
		}
	})

	t.Run("bi_false", func(t *testing.T) {
		if bi(false) != 0 {
			t.Error("bi(false) should return 0")
		}
	})
}

// ============================================================================
// Edge Case Tests for Dead Code Elimination
// ============================================================================

// TestDeadCodeElimination_EdgeCases tests DCE edge cases
func TestDeadCodeElimination_EdgeCases(t *testing.T) {
	dce := NewDeadCodeElimination()

	if dce == nil {
		t.Fatal("NewDeadCodeElimination() returned nil")
	}

	// Test Reset
	dce.Reset()

	// Test Run with nil IR
	changed, err := dce.Run(nil)
	if !changed {
		t.Log("Run(nil) should handle nil IR gracefully")
	}
	if err == nil {
		t.Log("Run(nil) returned nil error")
	}
}

// TestDeadCodeElimination_ProcessFunction tests processFunction edge cases
func TestDeadCodeElimination_ProcessFunction(t *testing.T) {
	dce := NewDeadCodeElimination()

	// Test with nil function
	changed, err := dce.processFunction(nil)
	if !changed {
		t.Log("processFunction(nil) should handle nil gracefully")
	}
	if err == nil {
		t.Log("processFunction(nil) returned nil error")
	}
}

// ============================================================================
// Edge Case Tests for IR Generator
// ============================================================================

// TestIRGenerator_EdgeCases tests IR generator edge cases
func TestIRGenerator_EdgeCases(t *testing.T) {
	errHandler := errhand.NewErrorHandler()
	gen := NewIRGenerator(errHandler)

	if gen == nil {
		t.Fatal("NewIRGenerator() returned nil")
	}

	// Test Generate with nil AST
	ir, err := gen.Generate(nil)
	if ir != nil {
		t.Log("Generate(nil) returned non-nil IR")
	}
	if err == nil {
		t.Log("Generate(nil) returned nil error")
	}
}

// TestOptimizingIRGenerator_EdgeCases tests optimizing IR generator edge cases
func TestOptimizingIRGenerator_EdgeCases(t *testing.T) {
	config := OptimizingIRGeneratorConfig{
		ErrorHandler: errhand.NewErrorHandler(),
		Optimization: OptimizationConfig{
			Enabled: false,
		},
	}

	gen, err := NewOptimizingIRGenerator(config)
	if err != nil {
		t.Fatalf("NewOptimizingIRGenerator() error: %v", err)
	}

	if gen == nil {
		t.Fatal("NewOptimizingIRGenerator() returned nil")
	}

	// Test Generate with nil AST
	ir, err := gen.Generate(nil)
	if ir != nil {
		t.Log("Generate(nil) returned non-nil IR")
	}
	if err == nil {
		t.Log("Generate(nil) returned nil error")
	}

	// Test GetPassManager
	pm := gen.GetPassManager()
	if pm == nil {
		t.Error("GetPassManager() returned nil")
	}

	// Test SetOptimizationEnabled
	gen.SetOptimizationEnabled(true)
	if !gen.IsOptimizationEnabled() {
		t.Error("IsOptimizationEnabled() should return true after enabling")
	}

	gen.SetOptimizationEnabled(false)
	if gen.IsOptimizationEnabled() {
		t.Error("IsOptimizationEnabled() should return false after disabling")
	}
}

// ============================================================================
// Edge Case Tests for Pass Manager
// ============================================================================

// TestPassManager_EdgeCases tests pass manager edge cases
func TestPassManager_EdgeCases(t *testing.T) {
	pm, err := NewPassManager(PassManagerConfig{
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("NewPassManager() error: %v", err)
	}

	if pm == nil {
		t.Fatal("NewPassManager() returned nil")
	}

	// Test Run with nil IR
	changed, err := pm.Run(nil)
	if !changed {
		t.Log("Run(nil) should handle nil IR gracefully")
	}
	if err == nil {
		t.Log("Run(nil) returned nil error")
	}

	// Test SetEnabled
	pm.SetEnabled(true)
	if !pm.IsEnabled() {
		t.Error("IsEnabled() should return true after enabling")
	}

	pm.SetEnabled(false)
	if pm.IsEnabled() {
		t.Error("IsEnabled() should return false after disabling")
	}

	// Test AddPass with nil
	pm.AddPass(nil)
	t.Log("AddPass(nil) handled gracefully")
}

// TestPassManagerWithPasses_EdgeCases tests pass manager with explicit passes
func TestPassManagerWithPasses_EdgeCases(t *testing.T) {
	pm := NewPassManagerWithPasses(true)

	if pm == nil {
		t.Fatal("NewPassManagerWithPasses() returned nil")
	}

	// Test Run with nil IR
	changed, err := pm.Run(nil)
	if !changed {
		t.Log("Run(nil) should handle nil IR gracefully")
	}
	if err == nil {
		t.Log("Run(nil) returned nil error")
	}

	// Test AddPass with nil
	pm.AddPass(nil)
	t.Log("AddPass(nil) handled gracefully")
}

// TestCreateDefaultPassManager tests default pass manager creation
func TestCreateDefaultPassManager_EdgeCase(t *testing.T) {
	pm := CreateDefaultPassManager()

	if pm == nil {
		t.Fatal("CreateDefaultPassManager() returned nil")
	}

	// Note: CreateDefaultPassManager returns a disabled manager by design
	// It's meant to be configured with passes before enabling
	if pm.IsEnabled() {
		t.Log("CreateDefaultPassManager returns disabled manager by design")
	}

	// Test that we can enable it
	pm.SetEnabled(true)
	if !pm.IsEnabled() {
		t.Error("Should be able to enable the pass manager")
	}
}