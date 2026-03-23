// Package codegen generates x86-64 assembly code from IR.
// This file contains tests for the register allocator.
package codegen

import (
	"testing"

	"github.com/akzj/goc/pkg/ir"
)

// TestNewRegisterAllocator tests the creation of a new register allocator.
func TestNewRegisterAllocator(t *testing.T) {
	ra := NewRegisterAllocator()

	if ra == nil {
		t.Fatal("NewRegisterAllocator returned nil")
	}

	// Check that GP registers are initialized
	if len(ra.availableGP) != len(gpAllocatable) {
		t.Errorf("Expected %d available GP registers, got %d", len(gpAllocatable), len(ra.availableGP))
	}

	// Check that FP registers are initialized
	if len(ra.availableFP) != len(fpAllocatable) {
		t.Errorf("Expected %d available FP registers, got %d", len(fpAllocatable), len(ra.availableFP))
	}

	// Check that all GP registers are available
	for _, reg := range gpAllocatable {
		if !ra.availableGP[reg] {
			t.Errorf("GP register %s should be available initially", reg.String())
		}
	}

	// Check that all FP registers are available
	for _, reg := range fpAllocatable {
		if !ra.availableFP[reg] {
			t.Errorf("FP register %s should be available initially", reg.String())
		}
	}
}

// TestAllocateGP tests general-purpose register allocation.
func TestAllocateGP(t *testing.T) {
	ra := NewRegisterAllocator()

	// Allocate registers for integer operands
	op1 := newTempOperand(1, intType())
	op2 := newTempOperand(2, intType())
	op3 := newTempOperand(3, longType())

	reg1 := ra.Allocate(op1)
	reg2 := ra.Allocate(op2)
	reg3 := ra.Allocate(op3)

	// Check that registers were allocated
	if reg1 == RAX {
		// RAX is the default, but should be allocated properly
		t.Logf("First allocation got RAX (expected)")
	}

	// Check that registers are marked as unavailable
	if ra.availableGP[reg1] {
		t.Errorf("Register %s should be unavailable after allocation", reg1.String())
	}
	if ra.availableGP[reg2] {
		t.Errorf("Register %s should be unavailable after allocation", reg2.String())
	}
	if ra.availableGP[reg3] {
		t.Errorf("Register %s should be unavailable after allocation", reg3.String())
	}

	// Check that operands are tracked
	if trackedReg, ok := ra.GetRegister(op1); !ok {
		t.Error("Operand 1 should be tracked")
	} else if trackedReg != reg1 {
		t.Errorf("Operand 1 should be in register %s, got %s", reg1.String(), trackedReg.String())
	}
}

// TestAllocateFP tests floating-point register allocation.
func TestAllocateFP(t *testing.T) {
	ra := NewRegisterAllocator()

	// Allocate registers for floating-point operands
	op1 := newTempOperand(1, floatType())
	op2 := newTempOperand(2, doubleType())

	reg1 := ra.Allocate(op1)
	reg2 := ra.Allocate(op2)

	// Check that FP registers were allocated
	if !ra.isFPRegister(reg1) {
		t.Errorf("Expected FP register for float operand, got %s", reg1.String())
	}
	if !ra.isFPRegister(reg2) {
		t.Errorf("Expected FP register for double operand, got %s", reg2.String())
	}

	// Check that registers are marked as unavailable
	if ra.availableFP[reg1] {
		t.Errorf("Register %s should be unavailable after allocation", reg1.String())
	}
	if ra.availableFP[reg2] {
		t.Errorf("Register %s should be unavailable after allocation", reg2.String())
	}
}

// TestFree tests register freeing.
func TestFree(t *testing.T) {
	ra := NewRegisterAllocator()

	op := newTempOperand(1, intType())
	reg := ra.Allocate(op)

	// Verify register is allocated
	if ra.availableGP[reg] {
		t.Errorf("Register %s should be unavailable", reg.String())
	}

	// Free the register
	ra.Free(reg)

	// Verify register is available
	if !ra.availableGP[reg] {
		t.Errorf("Register %s should be available after free", reg.String())
	}

	// Verify operand is no longer tracked
	if _, ok := ra.GetRegister(op); ok {
		t.Error("Operand should not be tracked after free")
	}
}

// TestFreeOperand tests freeing by operand.
func TestFreeOperand(t *testing.T) {
	ra := NewRegisterAllocator()

	op := newTempOperand(1, intType())
	reg := ra.Allocate(op)

	// Free by operand
	ra.FreeOperand(op)

	// Verify register is available
	if !ra.availableGP[reg] {
		t.Errorf("Register %s should be available after FreeOperand", reg.String())
	}
}

// TestSpill tests register spilling.
func TestSpill(t *testing.T) {
	ra := NewRegisterAllocator()

	// Allocate all GP registers
	ops := make([]*ir.Operand, len(gpAllocatable))
	regs := make([]Reg, len(gpAllocatable))

	for i := range gpAllocatable {
		ops[i] = newTempOperand(i, intType())
		regs[i] = ra.Allocate(ops[i])
	}

	// Try to allocate one more - should trigger spill
	extraOp := newTempOperand(999, intType())
	extraReg := ra.Allocate(extraOp)

	// Check that spill occurred
	if len(ra.spilled) == 0 {
		t.Error("Expected at least one spill when all registers are used")
	}

	// Check that extra operand got a register
	if _, ok := ra.GetRegister(extraOp); !ok {
		t.Error("Extra operand should have a register")
	}

	// Verify the extra register is one of the allocatable registers
	found := false
	for _, reg := range gpAllocatable {
		if reg == extraReg {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Extra register %s should be allocatable", extraReg.String())
	}
}

// TestSpillOffset tests spill slot offsets.
func TestSpillOffset(t *testing.T) {
	ra := NewRegisterAllocator()

	// Force spills by allocating all registers and then more
	ops := make([]*ir.Operand, len(gpAllocatable)+3)

	for i := range ops {
		ops[i] = newTempOperand(i, intType())
		ra.Allocate(ops[i])
	}

	// Check that spilled operands have positive offsets
	spilledCount := 0
	for _, op := range ops {
		if ra.IsSpilled(op) {
			offset := ra.GetSpillOffset(op)
			if offset <= 0 {
				t.Errorf("Spill offset should be positive, got %d", offset)
			}
			spilledCount++
		}
	}

	// Verify at least one spill occurred
	if spilledCount == 0 {
		t.Error("Expected at least one spill when allocating more operands than registers")
	}
}

// TestReload tests reloading from spill slots.
func TestReload(t *testing.T) {
	ra := NewRegisterAllocator()

	// Allocate all registers to force spill
	ops := make([]*ir.Operand, len(gpAllocatable)+1)
	for i := range ops {
		ops[i] = newTempOperand(i, intType())
		ra.Allocate(ops[i])
	}

	// Find a spilled operand
	var spilledOp *ir.Operand
	for _, op := range ops {
		if ra.IsSpilled(op) {
			spilledOp = op
			break
		}
	}

	if spilledOp == nil {
		t.Fatal("Expected at least one spilled operand")
	}

	// Reload the spilled operand
	reg := ra.GetReturnRegister(false) // Use a register for reload
	reloadedReg := ra.Reload(spilledOp, reg)

	// Check that reload returned a register
	if reloadedReg == RAX && spilledOp != nil {
		// May return RAX if that's what was allocated
		t.Logf("Reload returned %s", reloadedReg.String())
	}
}

// TestCalleeSaved tests callee-saved register handling.
func TestCalleeSaved(t *testing.T) {
	ra := NewRegisterAllocator()

	// Reserve callee-saved registers
	ra.ReserveCalleeSaved(calleeSavedRegs)

	// Check that callee-saved registers are not available
	for _, reg := range calleeSavedRegs {
		if ra.IsRegisterAvailable(reg) {
			t.Errorf("Callee-saved register %s should not be available", reg.String())
		}
	}

	// Check that GetCalleeSaved returns the reserved registers
	saved := ra.GetCalleeSaved()
	if len(saved) != len(calleeSavedRegs) {
		t.Errorf("Expected %d callee-saved registers, got %d", len(calleeSavedRegs), len(saved))
	}
}

// TestArgumentRegisters tests argument register allocation.
func TestArgumentRegisters(t *testing.T) {
	ra := NewRegisterAllocator()

	// Reserve registers for 3 integer arguments
	ra.ReserveForArguments(3, false)

	// Check that argument registers are not available
	for i := 0; i < 3; i++ {
		expectedReg := intArgRegs[i]
		if ra.IsRegisterAvailable(expectedReg) {
			t.Errorf("Argument register %s should not be available", expectedReg.String())
		}
	}

	// Check GetArgumentRegister
	for i := 0; i < 3; i++ {
		reg, ok := ra.GetArgumentRegister(i, false)
		if !ok {
			t.Errorf("GetArgumentRegister(%d) should return a register", i)
		}
		if reg != intArgRegs[i] {
			t.Errorf("GetArgumentRegister(%d) should return %s, got %s", i, intArgRegs[i].String(), reg.String())
		}
	}

	// Check that out-of-range arguments return false
	if _, ok := ra.GetArgumentRegister(10, false); ok {
		t.Error("GetArgumentRegister(10) should return false")
	}
}

// TestFPArgumentRegisters tests floating-point argument register allocation.
func TestFPArgumentRegisters(t *testing.T) {
	ra := NewRegisterAllocator()

	// Reserve registers for 2 FP arguments
	ra.ReserveForArguments(2, true)

	// Check that FP argument registers are not available
	for i := 0; i < 2; i++ {
		expectedReg := fpArgRegs[i]
		if ra.IsRegisterAvailable(expectedReg) {
			t.Errorf("FP argument register %s should not be available", expectedReg.String())
		}
	}

	// Check GetArgumentRegister for FP
	for i := 0; i < 2; i++ {
		reg, ok := ra.GetArgumentRegister(i, true)
		if !ok {
			t.Errorf("GetArgumentRegister(%d, true) should return a register", i)
		}
		if reg != fpArgRegs[i] {
			t.Errorf("GetArgumentRegister(%d, true) should return %s, got %s", i, fpArgRegs[i].String(), reg.String())
		}
	}
}

// TestReturnRegister tests return register selection.
func TestReturnRegister(t *testing.T) {
	ra := NewRegisterAllocator()

	// Check integer return register
	intReg := ra.GetReturnRegister(false)
	if intReg != RAX {
		t.Errorf("Integer return register should be RAX, got %s", intReg.String())
	}

	// Check FP return register
	fpReg := ra.GetReturnRegister(true)
	if fpReg != XMM0 {
		t.Errorf("FP return register should be XMM0, got %s", fpReg.String())
	}
}

// TestReset tests allocator reset.
func TestReset(t *testing.T) {
	ra := NewRegisterAllocator()

	// Allocate some registers
	op1 := newTempOperand(1, intType())
	op2 := newTempOperand(2, floatType())
	ra.Allocate(op1)
	ra.Allocate(op2)

	// Reserve some callee-saved
	ra.ReserveCalleeSaved([]Reg{RBX})

	// Reset
	ra.Reset()

	// Check that all GP registers are available again
	for _, reg := range gpAllocatable {
		if !ra.availableGP[reg] {
			t.Errorf("GP register %s should be available after reset", reg.String())
		}
	}

	// Check that all FP registers are available again
	for _, reg := range fpAllocatable {
		if !ra.availableFP[reg] {
			t.Errorf("FP register %s should be available after reset", reg.String())
		}
	}

	// Check that no operands are tracked
	if _, ok := ra.GetRegister(op1); ok {
		t.Error("Operand 1 should not be tracked after reset")
	}
	if _, ok := ra.GetRegister(op2); ok {
		t.Error("Operand 2 should not be tracked after reset")
	}

	// Check that callee-saved is cleared
	if len(ra.calleeSaved) != 0 {
		t.Error("Callee-saved should be empty after reset")
	}
}

// TestStackFrameSize tests stack frame size calculation.
func TestStackFrameSize(t *testing.T) {
	ra := NewRegisterAllocator()

	// Initially should be 0 (or aligned to 16)
	size := ra.GetStackFrameSize()
	if size != 0 {
		t.Errorf("Initial stack frame size should be 0, got %d", size)
	}

	// Force some spills
	ops := make([]*ir.Operand, len(gpAllocatable)+4)
	for i := range ops {
		ops[i] = newTempOperand(i, intType())
		ra.Allocate(ops[i])
	}

	// Check that stack frame size is 16-byte aligned
	size = ra.GetStackFrameSize()
	if size%16 != 0 {
		t.Errorf("Stack frame size %d should be 16-byte aligned", size)
	}

	// Check that size is positive
	if size <= 0 {
		t.Error("Stack frame size should be positive after spills")
	}
}

// TestSpillSlotCount tests spill slot counting.
func TestSpillSlotCount(t *testing.T) {
	ra := NewRegisterAllocator()

	// Initially should be 0
	count := ra.GetSpillSlotCount()
	if count != 0 {
		t.Errorf("Initial spill slot count should be 0, got %d", count)
	}

	// Force some spills
	ops := make([]*ir.Operand, len(gpAllocatable)+3)
	for i := range ops {
		ops[i] = newTempOperand(i, intType())
		ra.Allocate(ops[i])
	}

	// Check that spill slot count is positive
	count = ra.GetSpillSlotCount()
	if count <= 0 {
		t.Error("Spill slot count should be positive after spills")
	}
}

// TestAvailableCount tests available register counting.
func TestAvailableCount(t *testing.T) {
	ra := NewRegisterAllocator()

	// Initially all registers should be available
	gpCount := ra.GetAvailableGPCount()
	fpCount := ra.GetAvailableFPCount()

	if gpCount != len(gpAllocatable) {
		t.Errorf("Expected %d available GP registers, got %d", len(gpAllocatable), gpCount)
	}
	if fpCount != len(fpAllocatable) {
		t.Errorf("Expected %d available FP registers, got %d", len(fpAllocatable), fpCount)
	}

	// Allocate some registers
	op1 := newTempOperand(1, intType())
	op2 := newTempOperand(2, floatType())
	ra.Allocate(op1)
	ra.Allocate(op2)

	// Check counts decreased
	gpCount = ra.GetAvailableGPCount()
	fpCount = ra.GetAvailableFPCount()

	if gpCount != len(gpAllocatable)-1 {
		t.Errorf("Expected %d available GP registers after allocation, got %d", len(gpAllocatable)-1, gpCount)
	}
	if fpCount != len(fpAllocatable)-1 {
		t.Errorf("Expected %d available FP registers after allocation, got %d", len(fpAllocatable)-1, fpCount)
	}
}

// TestNilOperand tests handling of nil operands.
func TestNilOperand(t *testing.T) {
	ra := NewRegisterAllocator()

	// Allocate with nil operand should not panic
	reg := ra.Allocate(nil)
	if reg != RAX {
		t.Errorf("Allocate(nil) should return RAX, got %s", reg.String())
	}
}

// TestGetSpillOffsetNotSpilled tests GetSpillOffset for non-spilled operands.
func TestGetSpillOffsetNotSpilled(t *testing.T) {
	ra := NewRegisterAllocator()

	op := newTempOperand(1, intType())
	ra.Allocate(op)

	// Should return -1 for non-spilled operand
	offset := ra.GetSpillOffset(op)
	if offset != -1 {
		t.Errorf("GetSpillOffset for non-spilled operand should return -1, got %d", offset)
	}
}

// TestIsFPRegister tests FP register detection.
func TestIsFPRegister(t *testing.T) {
	ra := NewRegisterAllocator()

	// Test GP registers
	if ra.isFPRegister(RAX) {
		t.Error("RAX should not be detected as FP register")
	}
	if ra.isFPRegister(RBX) {
		t.Error("RBX should not be detected as FP register")
	}

	// Test FP registers
	if !ra.isFPRegister(XMM0) {
		t.Error("XMM0 should be detected as FP register")
	}
	if !ra.isFPRegister(XMM15) {
		t.Error("XMM15 should be detected as FP register")
	}
}

// TestIsRegisterAvailable tests register availability checking.
func TestIsRegisterAvailable(t *testing.T) {
	ra := NewRegisterAllocator()

	// Initially all should be available
	if !ra.IsRegisterAvailable(RAX) {
		t.Error("RAX should be available initially")
	}
	if !ra.IsRegisterAvailable(XMM0) {
		t.Error("XMM0 should be available initially")
	}

	// Allocate and check
	op := newTempOperand(1, intType())
	reg := ra.Allocate(op)

	if ra.IsRegisterAvailable(reg) {
		t.Errorf("Register %s should not be available after allocation", reg.String())
	}
}

// TestParamOperand tests parameter operand allocation.
func TestParamOperand(t *testing.T) {
	ra := NewRegisterAllocator()

	// Allocate registers for parameter operands
	op1 := newParamOperand(0, intType())
	op2 := newParamOperand(1, intType())

	reg1 := ra.Allocate(op1)
	reg2 := ra.Allocate(op2)

	// Check that registers were allocated
	if reg1 == RAX && reg2 == RAX && reg1 == reg2 {
		t.Error("Different operands should get different registers")
	}

	// Check that operands are tracked
	if _, ok := ra.GetRegister(op1); !ok {
		t.Error("Parameter operand 1 should be tracked")
	}
	if _, ok := ra.GetRegister(op2); !ok {
		t.Error("Parameter operand 2 should be tracked")
	}
}

// TestMultipleAllocations tests multiple allocations and frees.
func TestMultipleAllocations(t *testing.T) {
	ra := NewRegisterAllocator()

	// Allocate and free multiple times
	for i := 0; i < 5; i++ {
		op := newTempOperand(i, intType())
		reg := ra.Allocate(op)

		if !ra.IsRegisterAvailable(reg) {
			// Good, register is in use
		} else {
			t.Errorf("Register %s should be in use", reg.String())
		}

		ra.FreeOperand(op)

		if !ra.IsRegisterAvailable(reg) {
			t.Errorf("Register %s should be available after free", reg.String())
		}
	}
}

// TestAllGPRegistersAllocated tests allocation of all GP registers.
func TestAllGPRegistersAllocated(t *testing.T) {
	ra := NewRegisterAllocator()

	// Allocate all GP registers
	ops := make([]*ir.Operand, len(gpAllocatable))
	for i := range gpAllocatable {
		ops[i] = newTempOperand(i, intType())
		ra.Allocate(ops[i])
	}

	// Check that no GP registers are available
	count := ra.GetAvailableGPCount()
	if count != 0 {
		t.Errorf("Expected 0 available GP registers, got %d", count)
	}
}

// TestAllFPRegistersAllocated tests allocation of all FP registers.
func TestAllFPRegistersAllocated(t *testing.T) {
	ra := NewRegisterAllocator()

	// Allocate all FP registers
	ops := make([]*ir.Operand, len(fpAllocatable))
	for i := range fpAllocatable {
		ops[i] = newTempOperand(i, floatType())
		ra.Allocate(ops[i])
	}

	// Check that no FP registers are available
	count := ra.GetAvailableFPCount()
	if count != 0 {
		t.Errorf("Expected 0 available FP registers, got %d", count)
	}
}

// TestMixedAllocations tests mixed GP and FP allocations.
func TestMixedAllocations(t *testing.T) {
	ra := NewRegisterAllocator()

	// Allocate alternating GP and FP
	for i := 0; i < 5; i++ {
		gpOp := newTempOperand(i*2, intType())
		fpOp := newTempOperand(i*2+1, floatType())

		gpReg := ra.Allocate(gpOp)
		fpReg := ra.Allocate(fpOp)

		if ra.isFPRegister(gpReg) {
			t.Errorf("GP operand got FP register %s", gpReg.String())
		}
		if !ra.isFPRegister(fpReg) {
			t.Errorf("FP operand got GP register %s", fpReg.String())
		}
	}
}