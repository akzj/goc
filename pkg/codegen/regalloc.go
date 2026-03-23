// Package codegen generates x86-64 assembly code from IR.
// This file defines the register allocator.
package codegen

import (
	"fmt"

	"github.com/akzj/goc/pkg/ir"
)

// RegisterAllocator manages register allocation for code generation.
// It implements a simple linear-scan register allocator with spilling support.
//
// The allocator follows the System V AMD64 ABI calling convention:
//   - Caller-saved (volatile): RAX, RCX, RDX, RSI, RDI, R8-R11
//   - Callee-saved (non-volatile): RBX, RBP, R12-R15
//   - Stack pointer: RSP (reserved, not allocated)
//   - Floating point: XMM0-XMM15 (caller-saved)
//
// Register allocation strategy:
//   1. Allocate caller-saved registers for temporary values
//   2. Reserve callee-saved registers for special purposes (frame pointer, etc.)
//   3. Spill to stack when all registers are in use
//   4. Track liveness to optimize register reuse
type RegisterAllocator struct {
	// availableGP maps general-purpose registers to availability.
	// True means available, false means in use.
	availableGP map[Reg]bool

	// availableFP maps floating-point registers to availability.
	availableFP map[Reg]bool

	// currentGP maps operands to their current GP register assignment.
	currentGP map[*ir.Operand]Reg

	// currentFP maps operands to their current FP register assignment.
	currentFP map[*ir.Operand]Reg

	// spilled maps spilled operands to their stack offset.
	spilled map[*ir.Operand]int64

	// stackOffset is the current stack offset for spills (grows downward).
	stackOffset int64

	// calleeSaved tracks which callee-saved registers are used.
	calleeSaved []Reg

	// argRegs tracks registers used for function arguments.
	argRegs []Reg

	// spillBase is the base offset for spill slots (relative to RBP).
	spillBase int64
}

// General-purpose registers available for allocation (caller-saved).
var gpAllocatable = []Reg{RAX, RCX, RDX, RSI, RDI, R8, R9, R10, R11}

// Floating-point registers available for allocation.
var fpAllocatable = []Reg{XMM0, XMM1, XMM2, XMM3, XMM4, XMM5, XMM6, XMM7, XMM8, XMM9, XMM10, XMM11, XMM12, XMM13, XMM14, XMM15}

// Callee-saved registers (reserved for special use).
var calleeSavedRegs = []Reg{RBX, R12, R13, R14, R15}

// Argument registers for System V AMD64 ABI (integer).
var intArgRegs = []Reg{RDI, RSI, RDX, RCX, R8, R9}

// Argument registers for System V AMD64 ABI (floating point).
var fpArgRegs = []Reg{XMM0, XMM1, XMM2, XMM3, XMM4, XMM5, XMM6, XMM7}

// NewRegisterAllocator creates a new register allocator.
func NewRegisterAllocator() *RegisterAllocator {
	ra := &RegisterAllocator{
		availableGP: make(map[Reg]bool),
		availableFP: make(map[Reg]bool),
		currentGP:   make(map[*ir.Operand]Reg),
		currentFP:   make(map[*ir.Operand]Reg),
		spilled:     make(map[*ir.Operand]int64),
		stackOffset: 0,
		calleeSaved: make([]Reg, 0),
		argRegs:     make([]Reg, 0),
		spillBase:   -16, // Start below saved RBP (which is at -8)
	}

	// Initialize available GP registers
	for _, reg := range gpAllocatable {
		ra.availableGP[reg] = true
	}

	// Initialize available FP registers
	for _, reg := range fpAllocatable {
		ra.availableFP[reg] = true
	}

	return ra
}

// Allocate allocates a register for the given operand.
// Returns the allocated register.
// If the operand is a constant or global, returns a special marker.
// If no registers are available, spills an existing allocation.
func (ra *RegisterAllocator) Allocate(op *ir.Operand) Reg {
	if op == nil {
		return RAX
	}

	// Check if operand already has a register
	if reg, ok := ra.currentGP[op]; ok {
		return reg
	}
	if reg, ok := ra.currentFP[op]; ok {
		return reg
	}

	// Determine if we need a FP register based on type
	needsFP := ra.isFloatingPointType(op.Type)

	var reg Reg
	if needsFP {
		reg = ra.allocateFP(op)
	} else {
		reg = ra.allocateGP(op)
	}

	return reg
}

// allocateGP allocates a general-purpose register for the operand.
func (ra *RegisterAllocator) allocateGP(op *ir.Operand) Reg {
	// Try to find an available register
	for _, reg := range gpAllocatable {
		if ra.availableGP[reg] {
			ra.availableGP[reg] = false
			ra.currentGP[op] = reg
			return reg
		}
	}

	// No registers available, need to spill
	return ra.spillAndAllocateGP(op)
}

// allocateFP allocates a floating-point register for the operand.
func (ra *RegisterAllocator) allocateFP(op *ir.Operand) Reg {
	// Try to find an available register
	for _, reg := range fpAllocatable {
		if ra.availableFP[reg] {
			ra.availableFP[reg] = false
			ra.currentFP[op] = reg
			return reg
		}
	}

	// No registers available, need to spill
	return ra.spillAndAllocateFP(op)
}

// spillAndAllocateGP spills an existing allocation and allocates a GP register.
func (ra *RegisterAllocator) spillAndAllocateGP(op *ir.Operand) Reg {
	// Find a register to spill (use first one for simplicity)
	// In a more sophisticated allocator, we'd choose based on liveness
	var victimReg Reg
	var victimOp *ir.Operand

	for o, r := range ra.currentGP {
		victimOp = o
		victimReg = r
		break
	}

	if victimOp != nil {
		// Spill the victim
		offset := ra.Spill(victimOp, victimReg)

		// Remove victim from current allocations
		delete(ra.currentGP, victimOp)
		ra.availableGP[victimReg] = true

		// Mark victim as spilled
		ra.spilled[victimOp] = int64(offset)
	}

	// Allocate the freed register
	ra.availableGP[victimReg] = false
	ra.currentGP[op] = victimReg

	return victimReg
}

// spillAndAllocateFP spills an existing allocation and allocates an FP register.
func (ra *RegisterAllocator) spillAndAllocateFP(op *ir.Operand) Reg {
	// Find a register to spill
	var victimReg Reg
	var victimOp *ir.Operand

	for o, r := range ra.currentFP {
		victimOp = o
		victimReg = r
		break
	}

	if victimOp != nil {
		// Spill the victim
		offset := ra.Spill(victimOp, victimReg)

		// Remove victim from current allocations
		delete(ra.currentFP, victimOp)
		ra.availableFP[victimReg] = true

		// Mark victim as spilled
		ra.spilled[victimOp] = int64(offset)
	}

	// Allocate the freed register
	ra.availableFP[victimReg] = false
	ra.currentFP[op] = victimReg

	return victimReg
}

// Free releases a register, making it available for reuse.
func (ra *RegisterAllocator) Free(reg Reg) {
	// Find and remove any operand mapped to this register
	for op, r := range ra.currentGP {
		if r == reg {
			delete(ra.currentGP, op)
			ra.availableGP[reg] = true
			return
		}
	}

	for op, r := range ra.currentFP {
		if r == reg {
			delete(ra.currentFP, op)
			ra.availableFP[reg] = true
			return
		}
	}

	// If no operand was mapped, just mark as available
	if ra.isFPRegister(reg) {
		ra.availableFP[reg] = true
	} else {
		ra.availableGP[reg] = true
	}
}

// FreeOperand releases the register assigned to an operand.
func (ra *RegisterAllocator) FreeOperand(op *ir.Operand) {
	if reg, ok := ra.currentGP[op]; ok {
		delete(ra.currentGP, op)
		ra.availableGP[reg] = true
	}
	if reg, ok := ra.currentFP[op]; ok {
		delete(ra.currentFP, op)
		ra.availableFP[reg] = true
	}
}

// Spill spills a register's value to the stack.
// Returns the stack offset (relative to RBP) where the value is stored.
func (ra *RegisterAllocator) Spill(op *ir.Operand, reg Reg) int {
	// Allocate stack space for the spill
	ra.stackOffset += 8 // 8 bytes for 64-bit values
	offset := ra.stackOffset

	// Store the spill offset
	ra.spilled[op] = int64(offset)

	return int(offset)
}

// Reload reloads a spilled value from the stack into a register.
// Returns the register containing the reloaded value.
func (ra *RegisterAllocator) Reload(op *ir.Operand, reg Reg) Reg {
	// Check if operand is spilled
	_, ok := ra.spilled[op]
	if !ok {
		// Not spilled, just return the register
		return reg
	}

	// Allocate a register for the reload
	allocatedReg := ra.Allocate(op)

	// The caller should emit a load instruction from [rbp - offset] to allocatedReg
	// This is handled by the code generator

	return allocatedReg
}

// IsSpilled returns true if the operand is currently spilled to the stack.
func (ra *RegisterAllocator) IsSpilled(op *ir.Operand) bool {
	_, ok := ra.spilled[op]
	return ok
}

// GetSpillOffset returns the stack offset for a spilled operand.
// Returns -1 if the operand is not spilled.
func (ra *RegisterAllocator) GetSpillOffset(op *ir.Operand) int64 {
	if offset, ok := ra.spilled[op]; ok {
		return offset
	}
	return -1
}

// GetRegister returns the register assigned to an operand.
// Returns false if the operand is not currently in a register.
func (ra *RegisterAllocator) GetRegister(op *ir.Operand) (Reg, bool) {
	if reg, ok := ra.currentGP[op]; ok {
		return reg, true
	}
	if reg, ok := ra.currentFP[op]; ok {
		return reg, true
	}
	return RAX, false
}

// ReserveCalleeSaved marks callee-saved registers as used.
// These registers must be saved and restored by the function.
func (ra *RegisterAllocator) ReserveCalleeSaved(regs []Reg) {
	ra.calleeSaved = append(ra.calleeSaved, regs...)
	for _, reg := range regs {
		ra.availableGP[reg] = false
	}
}

// GetCalleeSaved returns the list of callee-saved registers used.
func (ra *RegisterAllocator) GetCalleeSaved() []Reg {
	return ra.calleeSaved
}

// ReserveForArguments marks registers used for function arguments.
func (ra *RegisterAllocator) ReserveForArguments(count int, isFP bool) {
	if isFP {
		for i := 0; i < count && i < len(fpArgRegs); i++ {
			ra.argRegs = append(ra.argRegs, fpArgRegs[i])
			ra.availableFP[fpArgRegs[i]] = false
		}
	} else {
		for i := 0; i < count && i < len(intArgRegs); i++ {
			ra.argRegs = append(ra.argRegs, intArgRegs[i])
			ra.availableGP[intArgRegs[i]] = false
		}
	}
}

// GetArgumentRegister returns the register for the nth argument.
func (ra *RegisterAllocator) GetArgumentRegister(n int, isFP bool) (Reg, bool) {
	if isFP {
		if n < len(fpArgRegs) {
			return fpArgRegs[n], true
		}
	} else {
		if n < len(intArgRegs) {
			return intArgRegs[n], true
		}
	}
	return RAX, false
}

// GetReturnRegister returns the register used for function return values.
func (ra *RegisterAllocator) GetReturnRegister(isFP bool) Reg {
	if isFP {
		return XMM0
	}
	return RAX
}

// Reset resets the allocator state for a new function.
func (ra *RegisterAllocator) Reset() {
	// Clear current allocations
	ra.currentGP = make(map[*ir.Operand]Reg)
	ra.currentFP = make(map[*ir.Operand]Reg)
	ra.spilled = make(map[*ir.Operand]int64)
	ra.stackOffset = 0
	ra.calleeSaved = make([]Reg, 0)
	ra.argRegs = make([]Reg, 0)
	ra.spillBase = -16

	// Re-initialize available registers
	for _, reg := range gpAllocatable {
		ra.availableGP[reg] = true
	}
	for _, reg := range fpAllocatable {
		ra.availableFP[reg] = true
	}
}

// GetStackFrameSize returns the total stack frame size needed.
// This includes space for spilled values and local variables.
func (ra *RegisterAllocator) GetStackFrameSize() int64 {
	// Round up to 16-byte alignment (System V ABI requirement)
	size := ra.stackOffset
	if size%16 != 0 {
		size = (size/16 + 1) * 16
	}
	return size
}

// GetSpillSlotCount returns the number of spill slots allocated.
func (ra *RegisterAllocator) GetSpillSlotCount() int {
	return len(ra.spilled)
}

// isFloatingPointType returns true if the type is a floating-point type.
func (ra *RegisterAllocator) isFloatingPointType(t interface{}) bool {
	if t == nil {
		return false
	}
	// Check type string for float/double
	// This is a simplified check; a more robust implementation would use type enums
	typeString := fmt.Sprintf("%v", t)
	return typeString == "float" || typeString == "double"
}

// isFPRegister returns true if the register is a floating-point register.
func (ra *RegisterAllocator) isFPRegister(reg Reg) bool {
	return reg >= XMM0 && reg <= XMM15
}

// IsRegisterAvailable returns true if the register is currently available.
func (ra *RegisterAllocator) IsRegisterAvailable(reg Reg) bool {
	if ra.isFPRegister(reg) {
		return ra.availableFP[reg]
	}
	return ra.availableGP[reg]
}

// GetAvailableGPCount returns the number of available general-purpose registers.
func (ra *RegisterAllocator) GetAvailableGPCount() int {
	count := 0
	for _, available := range ra.availableGP {
		if available {
			count++
		}
	}
	return count
}

// GetAvailableFPCount returns the number of available floating-point registers.
func (ra *RegisterAllocator) GetAvailableFPCount() int {
	count := 0
	for _, available := range ra.availableFP {
		if available {
			count++
		}
	}
	return count
}