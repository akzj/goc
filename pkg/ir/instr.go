// Package ir provides intermediate representation for the GOC compiler.
// This file defines IR instructions and operands.
package ir

import (
	"github.com/akzj/goc/pkg/parser"
)

// TODO: Implement IR instructions
// Reference: docs/architecture-design-phases-2-7.md Section 5.3

// Instruction represents a three-address instruction.
type Instruction interface {
	// Opcode returns the instruction opcode.
	Opcode() Opcode
	// Dest returns the destination operand (nil if no destination).
	Dest() *Operand
	// Operands returns the list of operands.
	Operands() []*Operand
	// String returns a string representation.
	String() string
}

// Opcode represents an instruction opcode.
type Opcode int

const (
	// Arithmetic opcodes
	OpAdd Opcode = iota
	OpSub
	OpMul
	OpDiv
	OpMod
	OpNeg
	OpBitNot
	OpBitAnd
	OpBitOr
	OpBitXor
	OpShl
	OpShr

	// Comparison opcodes
	OpEq
	OpNe
	OpLt
	OpLe
	OpGt
	OpGe

	// Logical opcodes
	OpAnd
	OpOr
	OpNot

	// Memory opcodes
	OpLoad
	OpStore
	OpLea
	OpAlloc
	OpFree

	// Control flow opcodes
	OpJmp
	OpJmpIf
	OpJmpUnless
	OpCall
	OpRet
	OpLabel

	// Conversion opcodes
	OpCast
	OpZeroExt
	OpSignExt
	OpTrunc

	// Special opcodes
	OpPhi
	OpNop
)

// Operand represents an instruction operand.
type Operand struct {
	// Kind is the operand kind.
	Kind OperandKind
	// Type is the operand type.
	Type parser.Type
	// Value is the operand value (depends on kind).
	Value interface{}
}

// OperandKind represents the kind of operand.
type OperandKind int

const (
	// OperandTemp represents a temporary variable.
	OperandTemp OperandKind = iota
	// OperandParam represents a function parameter.
	OperandParam
	// OperandGlobal represents a global variable.
	OperandGlobal
	// OperandConst represents a constant value.
	OperandConst
	// OperandLabel represents a label.
	OperandLabel
)

// Temp represents a temporary variable.
type Temp struct {
	// ID is the temporary variable ID.
	ID int
	// Type is the variable type.
	Type parser.Type
}