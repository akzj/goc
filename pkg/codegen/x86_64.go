// Package codegen generates x86-64 assembly code from IR.
// This file defines x86-64 registers and instructions.
package codegen

// TODO: Implement codegen package
// Reference: docs/architecture-design-phases-2-7.md Section 6

// Reg represents an x86-64 register.
type Reg int

const (
	// General purpose registers
	RAX Reg = iota
	RBX
	RCX
	RDX
	RSI
	RDI
	RBP
	RSP
	R8
	R9
	R10
	R11
	R12
	R13
	R14
	R15

	// Floating point registers
	XMM0
	XMM1
	XMM2
	XMM3
	XMM4
	XMM5
	XMM6
	XMM7
	XMM8
	XMM9
	XMM10
	XMM11
	XMM12
	XMM13
	XMM14
	XMM15
)

// String returns the assembly name of the register.
func (r Reg) String() string {
	switch r {
	// General purpose registers
	case RAX:
		return "rax"
	case RBX:
		return "rbx"
	case RCX:
		return "rcx"
	case RDX:
		return "rdx"
	case RSI:
		return "rsi"
	case RDI:
		return "rdi"
	case RBP:
		return "rbp"
	case RSP:
		return "rsp"
	case R8:
		return "r8"
	case R9:
		return "r9"
	case R10:
		return "r10"
	case R11:
		return "r11"
	case R12:
		return "r12"
	case R13:
		return "r13"
	case R14:
		return "r14"
	case R15:
		return "r15"
	// Floating point registers
	case XMM0:
		return "xmm0"
	case XMM1:
		return "xmm1"
	case XMM2:
		return "xmm2"
	case XMM3:
		return "xmm3"
	case XMM4:
		return "xmm4"
	case XMM5:
		return "xmm5"
	case XMM6:
		return "xmm6"
	case XMM7:
		return "xmm7"
	case XMM8:
		return "xmm8"
	case XMM9:
		return "xmm9"
	case XMM10:
		return "xmm10"
	case XMM11:
		return "xmm11"
	case XMM12:
		return "xmm12"
	case XMM13:
		return "xmm13"
	case XMM14:
		return "xmm14"
	case XMM15:
		return "xmm15"
	default:
		return "unknown"
	}
}

// Size returns the register size in bytes.
func (r Reg) Size() int {
	switch r {
	// General purpose registers: 8 bytes (64-bit)
	case RAX, RBX, RCX, RDX, RSI, RDI, RBP, RSP, R8, R9, R10, R11, R12, R13, R14, R15:
		return 8
	// Floating point registers: 16 bytes (128-bit)
	case XMM0, XMM1, XMM2, XMM3, XMM4, XMM5, XMM6, XMM7, XMM8, XMM9, XMM10, XMM11, XMM12, XMM13, XMM14, XMM15:
		return 16
	default:
		return 0
	}
}

// Instruction represents an x86-64 instruction.
type Instruction struct {
	// Mnemonic is the instruction mnemonic.
	Mnemonic string
	// Operands is the list of operands.
	Operands []Operand
	// Comment is an optional comment.
	Comment string
}

// Operand represents an instruction operand.
type Operand struct {
	// Kind is the operand kind.
	Kind OperandKind
	// Value is the operand value.
	Value interface{}
}

// OperandKind represents the kind of operand.
type OperandKind int

const (
	// OperandReg represents a register.
	OperandReg OperandKind = iota
	// OperandImm represents an immediate value.
	OperandImm
	// OperandMem represents a memory operand.
	OperandMem
	// OperandLabel represents a label.
	OperandLabel
)

// MemOperand represents a memory operand.
type MemOperand struct {
	// Base is the base register.
	Base Reg
	// Index is the index register (optional).
	Index Reg
	// Scale is the scale factor (1, 2, 4, or 8).
	Scale int
	// Disp is the displacement.
	Disp int64
}