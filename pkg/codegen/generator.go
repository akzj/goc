// Package codegen generates x86-64 assembly code from IR.
// This file defines the code generator interface.
package codegen

import (
	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/ir"
)

// TODO: Implement code generator
// Reference: docs/architecture-design-phases-2-7.md Section 6.3

// CodeGenerator generates x86-64 assembly from IR.
type CodeGenerator struct {
	// ir is the IR to generate code for.
	ir *ir.IR
	// errors is the error handler.
	errors *errhand.ErrorHandler
	// output is the assembly output buffer.
	output *string
	// regAlloc is the register allocator.
	regAlloc *RegisterAllocator
	// stackFrame is the current stack frame.
	stackFrame *StackFrame
}

// StackFrame represents a function stack frame.
type StackFrame struct {
	// Size is the total frame size.
	Size int64
	// Locals is the list of local variables.
	Locals []*ir.LocalVar
	// SavedRegs is the list of saved registers.
	SavedRegs []Reg
}

// NewCodeGenerator creates a new code generator.
func NewCodeGenerator(errorHandler *errhand.ErrorHandler) *CodeGenerator {
	// TODO: Implement
	return nil
}

// Generate generates assembly code.
func (cg *CodeGenerator) Generate(ir *ir.IR) (string, error) {
	// TODO: Implement
	return "", nil
}

// GenerateFunction generates assembly for a single function.
func (cg *CodeGenerator) GenerateFunction(fn *ir.Function) string {
	// TODO: Implement
	return ""
}

// Emit emits an assembly instruction.
func (cg *CodeGenerator) Emit(instr *Instruction) {
	// TODO: Implement
}

// EmitLabel emits a label.
func (cg *CodeGenerator) EmitLabel(label string) {
	// TODO: Implement
}