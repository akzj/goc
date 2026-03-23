// Package codegen generates x86-64 assembly code from IR.
// This file defines the register allocator.
package codegen

import (
	"github.com/akzj/goc/pkg/ir"
)

// TODO: Implement register allocator
// Reference: docs/architecture-design-phases-2-7.md Section 6.4

// RegisterAllocator manages register allocation.
type RegisterAllocator struct {
	// available maps registers to availability.
	available map[Reg]bool
	// spilled maps spilled temps to stack offsets.
	spilled map[*ir.Operand]int
	// current maps operands to current register assignments.
	current map[*ir.Operand]Reg
	// stackOffset is the current stack offset for spills.
	stackOffset int64
}

// NewRegisterAllocator creates a new register allocator.
func NewRegisterAllocator() *RegisterAllocator {
	// TODO: Implement
	return nil
}

// Allocate allocates a register for the given operand.
func (ra *RegisterAllocator) Allocate(op *ir.Operand) Reg {
	// TODO: Implement
	return RAX
}

// Free releases a register.
func (ra *RegisterAllocator) Free(reg Reg) {
	// TODO: Implement
}

// Spill spills a register to the stack.
func (ra *RegisterAllocator) Spill(op *ir.Operand, reg Reg) int {
	// TODO: Implement
	return 0
}

// Reload reloads a spilled value from the stack.
func (ra *RegisterAllocator) Reload(op *ir.Operand, reg Reg) {
	// TODO: Implement
}