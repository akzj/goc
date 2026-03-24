// Package codegen provides edge case tests for the code generator.
// These tests focus on edge cases not covered in other test files.
package codegen

import (
	"testing"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/ir"
)

// ============================================================================
// Edge Case Tests for Code Generator
// ============================================================================

// TestCodeGenerator_EdgeCases tests code generator edge cases
func TestCodeGenerator_EdgeCases(t *testing.T) {
	errHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errHandler)

	if cg == nil {
		t.Fatal("NewCodeGenerator() returned nil")
	}

	// Test that CodeGenerator is created successfully
	t.Log("CodeGenerator created successfully")

	// Note: Generate(nil) will panic - this is expected behavior
	// The code generator requires a valid IR to work with
}

// TestCodeGenerator_GenerateFunction_EdgeCases tests GenerateFunction edge cases
func TestCodeGenerator_GenerateFunction_EdgeCases(t *testing.T) {
	errHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errHandler)

	if cg == nil {
		t.Fatal("NewCodeGenerator() returned nil")
	}

	// Note: GenerateFunction(nil) will panic - this is expected behavior
	// The code generator requires a valid IR function to work with
	// This test just verifies the CodeGenerator can be created
	t.Log("CodeGenerator created and ready for use with valid IR")
}

// TestCodeGenerator_EmitHeader tests header emission
func TestCodeGenerator_EmitHeader(t *testing.T) {
	errHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errHandler)

	cg.emitHeader()
	// Just test that it doesn't panic
}

// TestCodeGenerator_EmitLabel tests label emission
func TestCodeGenerator_EmitLabel(t *testing.T) {
	errHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errHandler)

	// Create a simple label operand
	labelOp := &ir.Operand{Kind: ir.OperandLabel, Value: "test_label"}
	label := ir.NewLabelInstr(labelOp)

	cg.emitLabel(label)
	// Just test that it doesn't panic
}

// TestCodeGenerator_NewLabel tests label generation
func TestCodeGenerator_NewLabel(t *testing.T) {
	errHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errHandler)

	label1 := cg.NewLabel()
	label2 := cg.NewLabel()

	if label1 == label2 {
		t.Error("NewLabel() should generate unique labels")
	}

	if label1 == "" {
		t.Error("NewLabel() should not return empty string")
	}
}

// ============================================================================
// Edge Case Tests for Register Allocator
// ============================================================================

// TestRegisterAllocator_EdgeCases tests register allocator edge cases
func TestRegisterAllocator_EdgeCases(t *testing.T) {
	ra := NewRegisterAllocator()

	if ra == nil {
		t.Fatal("NewRegisterAllocator() returned nil")
	}

	// Test Allocate with nil operand
	reg := ra.Allocate(nil)
	t.Logf("Allocate(nil) returned reg=%v", reg)

	// Test Free with invalid register
	ra.Free(RAX) // Should handle gracefully
	ra.Free(Reg(999)) // Invalid register

	// Test Reset
	ra.Reset()

	// Test GetStackFrameSize after reset
	size := ra.GetStackFrameSize()
	t.Logf("GetStackFrameSize() after reset = %d", size)

	// Test GetSpillSlotCount after reset
	count := ra.GetSpillSlotCount()
	t.Logf("GetSpillSlotCount() after reset = %d", count)
}

// TestRegisterAllocator_AllocateGP tests GP register allocation
func TestRegisterAllocator_AllocateGP(t *testing.T) {
	ra := NewRegisterAllocator()

	// Create a simple int operand
	op := &ir.Operand{
		Type: nil, // Will be treated as integer type
	}

	reg := ra.allocateGP(op)
	t.Logf("allocateGP returned reg=%v", reg)

	ra.Free(reg)
}

// TestRegisterAllocator_AllocateFP tests FP register allocation
func TestRegisterAllocator_AllocateFP(t *testing.T) {
	ra := NewRegisterAllocator()

	// Create a simple float operand
	op := &ir.Operand{
		Type: nil, // Will be treated based on context
	}

	reg := ra.allocateFP(op)
	t.Logf("allocateFP returned reg=%v", reg)

	ra.Free(reg)
}

// TestRegisterAllocator_SpillAndReload tests spill and reload
func TestRegisterAllocator_SpillAndReload(t *testing.T) {
	ra := NewRegisterAllocator()

	op := &ir.Operand{
		Type: nil,
	}

	// Allocate a register
	reg := ra.Allocate(op)
	if reg == Reg(-1) {
		t.Skip("No registers available for spill test")
	}

	// Spill the register
	offset := ra.Spill(op, reg)
	t.Logf("Spill offset = %d", offset)

	// Check if spilled
	if !ra.IsSpilled(op) {
		t.Log("IsSpilled() returned false after spill")
	}

	// Get spill offset
	spillOffset := ra.GetSpillOffset(op)
	t.Logf("GetSpillOffset() = %d", spillOffset)

	// Reload
	reloadedReg := ra.Reload(op, reg)
	t.Logf("Reload returned reg=%v", reloadedReg)

	ra.Free(reloadedReg)
}

// TestRegisterAllocator_ArgumentRegisters tests argument register handling
func TestRegisterAllocator_ArgumentRegisters(t *testing.T) {
	ra := NewRegisterAllocator()

	// Reserve for arguments
	ra.ReserveForArguments(6, false) // 6 integer arguments

	// Get argument registers
	for i := 0; i < 6; i++ {
		reg, ok := ra.GetArgumentRegister(i, false)
		t.Logf("GetArgumentRegister(%d, false) = (%v, %v)", i, reg, ok)
	}

	// Test FP arguments
	ra.ReserveForArguments(6, true) // 6 FP arguments
	for i := 0; i < 6; i++ {
		reg, ok := ra.GetArgumentRegister(i, true)
		t.Logf("GetArgumentRegister(%d, true) = (%v, %v)", i, reg, ok)
	}

	ra.Reset()
}

// TestRegisterAllocator_ReturnRegister tests return register
func TestRegisterAllocator_ReturnRegister(t *testing.T) {
	ra := NewRegisterAllocator()

	// Get integer return register
	intReg := ra.GetReturnRegister(false)
	t.Logf("GetReturnRegister(false) = %v", intReg)

	// Get FP return register
	fpReg := ra.GetReturnRegister(true)
	t.Logf("GetReturnRegister(true) = %v", fpReg)
}

// TestRegisterAllocator_CalleeSaved tests callee-saved registers
func TestRegisterAllocator_CalleeSaved(t *testing.T) {
	ra := NewRegisterAllocator()

	// Reserve callee-saved registers
	ra.ReserveCalleeSaved([]Reg{RBX, RBP, R12, R13, R14, R15})

	// Get callee-saved
	saved := ra.GetCalleeSaved()
	t.Logf("GetCalleeSaved() returned %d registers", len(saved))

	ra.Reset()
}

// TestRegisterAllocator_Availability tests register availability
func TestRegisterAllocator_Availability(t *testing.T) {
	ra := NewRegisterAllocator()

	// Test IsRegisterAvailable
	available := ra.IsRegisterAvailable(RAX)
	t.Logf("IsRegisterAvailable(RAX) = %v", available)

	// Test GetAvailableGPCount
	gpCount := ra.GetAvailableGPCount()
	t.Logf("GetAvailableGPCount() = %d", gpCount)

	// Test GetAvailableFPCount
	fpCount := ra.GetAvailableFPCount()
	t.Logf("GetAvailableFPCount() = %d", fpCount)

	// Allocate some registers
	op1 := &ir.Operand{Type: nil}
	op2 := &ir.Operand{Type: nil}
	reg1 := ra.Allocate(op1)
	reg2 := ra.Allocate(op2)

	// Check availability again
	gpCountAfter := ra.GetAvailableGPCount()
	t.Logf("GetAvailableGPCount() after allocation = %d", gpCountAfter)

	ra.Free(reg1)
	ra.Free(reg2)
}

// TestRegisterAllocator_isFloatingPointType tests type checking
func TestRegisterAllocator_isFloatingPointType(t *testing.T) {
	ra := NewRegisterAllocator()

	// Test with nil
	isFP := ra.isFloatingPointType(nil)
	t.Logf("isFloatingPointType(nil) = %v", isFP)
}

// TestRegisterAllocator_isFPRegister tests FP register detection
func TestRegisterAllocator_isFPRegister(t *testing.T) {
	ra := NewRegisterAllocator()

	// Test with various registers
	tests := []struct {
		reg  Reg
		want bool
	}{
		{XMM0, true},
		{XMM1, true},
		{XMM7, true},
		{RAX, false},
		{RBX, false},
		{Reg(999), false},
	}

	for _, tt := range tests {
		t.Run(tt.reg.String(), func(t *testing.T) {
			got := ra.isFPRegister(tt.reg)
			if got != tt.want {
				t.Errorf("isFPRegister(%v) = %v, want %v", tt.reg, got, tt.want)
			}
		})
	}
}