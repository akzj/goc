// Package ir provides intermediate representation for the GOC compiler.
// This file defines IR instructions and operands.
package ir

import (
	"fmt"
	"strings"

	"github.com/akzj/goc/pkg/parser"
)

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

// String returns the string representation of the opcode.
func (op Opcode) String() string {
	switch op {
	// Arithmetic
	case OpAdd:
		return "add"
	case OpSub:
		return "sub"
	case OpMul:
		return "mul"
	case OpDiv:
		return "div"
	case OpMod:
		return "mod"
	case OpNeg:
		return "neg"
	case OpBitNot:
		return "bitnot"
	case OpBitAnd:
		return "and"
	case OpBitOr:
		return "or"
	case OpBitXor:
		return "xor"
	case OpShl:
		return "shl"
	case OpShr:
		return "shr"

	// Comparison
	case OpEq:
		return "eq"
	case OpNe:
		return "ne"
	case OpLt:
		return "lt"
	case OpLe:
		return "le"
	case OpGt:
		return "gt"
	case OpGe:
		return "ge"

	// Logical
	case OpAnd:
		return "and"
	case OpOr:
		return "or"
	case OpNot:
		return "not"

	// Memory
	case OpLoad:
		return "load"
	case OpStore:
		return "store"
	case OpLea:
		return "lea"
	case OpAlloc:
		return "alloc"
	case OpFree:
		return "free"

	// Control flow
	case OpJmp:
		return "jmp"
	case OpJmpIf:
		return "jmpif"
	case OpJmpUnless:
		return "jmpunless"
	case OpCall:
		return "call"
	case OpRet:
		return "ret"
	case OpLabel:
		return "label"

	// Conversion
	case OpCast:
		return "cast"
	case OpZeroExt:
		return "zext"
	case OpSignExt:
		return "sext"
	case OpTrunc:
		return "trunc"

	// Special
	case OpPhi:
		return "phi"
	case OpNop:
		return "nop"

	default:
		return "unknown"
	}
}

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

// String returns the string representation of the operand kind.
func (k OperandKind) String() string {
	switch k {
	case OperandTemp:
		return "temp"
	case OperandParam:
		return "param"
	case OperandGlobal:
		return "global"
	case OperandConst:
		return "const"
	case OperandLabel:
		return "label"
	default:
		return "unknown"
	}
}

// String returns a string representation of the operand.
func (op *Operand) String() string {
	if op == nil {
		return "nil"
	}
	switch op.Kind {
	case OperandTemp:
		if t, ok := op.Value.(*Temp); ok {
			return fmt.Sprintf("t%d", t.ID)
		}
		return fmt.Sprintf("temp(%v)", op.Value)
	case OperandParam:
		if t, ok := op.Value.(*Temp); ok {
			return fmt.Sprintf("p%d", t.ID)
		}
		return fmt.Sprintf("param(%v)", op.Value)
	case OperandGlobal:
		return fmt.Sprintf("global(%v)", op.Value)
	case OperandConst:
		return fmt.Sprintf("%v", op.Value)
	case OperandLabel:
		return fmt.Sprintf("label(%v)", op.Value)
	default:
		return fmt.Sprintf("unknown(%v)", op.Value)
	}
}

// Temp represents a temporary variable.
type Temp struct {
	// ID is the temporary variable ID.
	ID int
	// Type is the variable type.
	Type parser.Type
}

// String returns a string representation of the temp.
func (t *Temp) String() string {
	return fmt.Sprintf("t%d", t.ID)
}

// ============================================================================
// Concrete Instruction Types
// ============================================================================

// BinaryInstr represents a binary operation instruction.
type BinaryInstr struct {
	opcode Opcode
	dest   *Operand
	left   *Operand
	right  *Operand
}

// NewBinaryInstr creates a new binary instruction.
func NewBinaryInstr(opcode Opcode, dest, left, right *Operand) *BinaryInstr {
	return &BinaryInstr{
		opcode: opcode,
		dest:   dest,
		left:   left,
		right:  right,
	}
}

// Opcode implements Instruction.
func (i *BinaryInstr) Opcode() Opcode { return i.opcode }

// Dest implements Instruction.
func (i *BinaryInstr) Dest() *Operand { return i.dest }

// Operands implements Instruction.
func (i *BinaryInstr) Operands() []*Operand { return []*Operand{i.left, i.right} }

// String implements Instruction.
func (i *BinaryInstr) String() string {
	return fmt.Sprintf("%s %s, %s, %s", i.opcode, i.dest, i.left, i.right)
}

// UnaryInstr represents a unary operation instruction.
type UnaryInstr struct {
	opcode Opcode
	dest   *Operand
	operand *Operand
}

// NewUnaryInstr creates a new unary instruction.
func NewUnaryInstr(opcode Opcode, dest, operand *Operand) *UnaryInstr {
	return &UnaryInstr{
		opcode: opcode,
		dest:   dest,
		operand: operand,
	}
}

// Opcode implements Instruction.
func (i *UnaryInstr) Opcode() Opcode { return i.opcode }

// Dest implements Instruction.
func (i *UnaryInstr) Dest() *Operand { return i.dest }

// Operands implements Instruction.
func (i *UnaryInstr) Operands() []*Operand { return []*Operand{i.operand} }

// String implements Instruction.
func (i *UnaryInstr) String() string {
	return fmt.Sprintf("%s %s, %s", i.opcode, i.dest, i.operand)
}

// LoadInstr represents a load instruction.
type LoadInstr struct {
	dest *Operand
	addr *Operand
}

// NewLoadInstr creates a new load instruction.
func NewLoadInstr(dest, addr *Operand) *LoadInstr {
	return &LoadInstr{
		dest: dest,
		addr: addr,
	}
}

// Opcode implements Instruction.
func (i *LoadInstr) Opcode() Opcode { return OpLoad }

// Dest implements Instruction.
func (i *LoadInstr) Dest() *Operand { return i.dest }

// Operands implements Instruction.
func (i *LoadInstr) Operands() []*Operand { return []*Operand{i.addr} }

// String implements Instruction.
func (i *LoadInstr) String() string {
	return fmt.Sprintf("load %s, %s", i.dest, i.addr)
}

// StoreInstr represents a store instruction.
type StoreInstr struct {
	value *Operand
	addr  *Operand
}

// NewStoreInstr creates a new store instruction.
func NewStoreInstr(value, addr *Operand) *StoreInstr {
	return &StoreInstr{
		value: value,
		addr:  addr,
	}
}

// Opcode implements Instruction.
func (i *StoreInstr) Opcode() Opcode { return OpStore }

// Dest implements Instruction.
func (i *StoreInstr) Dest() *Operand { return nil }

// Operands implements Instruction.
func (i *StoreInstr) Operands() []*Operand { return []*Operand{i.value, i.addr} }

// String implements Instruction.
func (i *StoreInstr) String() string {
	return fmt.Sprintf("store %s, %s", i.value, i.addr)
}

// JmpInstr represents an unconditional jump instruction.
type JmpInstr struct {
	target *Operand
}

// NewJmpInstr creates a new jump instruction.
func NewJmpInstr(target *Operand) *JmpInstr {
	return &JmpInstr{
		target: target,
	}
}

// Opcode implements Instruction.
func (i *JmpInstr) Opcode() Opcode { return OpJmp }

// Dest implements Instruction.
func (i *JmpInstr) Dest() *Operand { return nil }

// Operands implements Instruction.
func (i *JmpInstr) Operands() []*Operand { return []*Operand{i.target} }

// String implements Instruction.
func (i *JmpInstr) String() string {
	return fmt.Sprintf("jmp %s", i.target)
}

// CondJmpInstr represents a conditional jump instruction.
type CondJmpInstr struct {
	opcode Opcode
	cond   *Operand
	target *Operand
}

// NewCondJmpInstr creates a new conditional jump instruction.
func NewCondJmpInstr(opcode Opcode, cond, target *Operand) *CondJmpInstr {
	return &CondJmpInstr{
		opcode: opcode,
		cond:   cond,
		target: target,
	}
}

// Opcode implements Instruction.
func (i *CondJmpInstr) Opcode() Opcode { return i.opcode }

// Dest implements Instruction.
func (i *CondJmpInstr) Dest() *Operand { return nil }

// Operands implements Instruction.
func (i *CondJmpInstr) Operands() []*Operand { return []*Operand{i.cond, i.target} }

// String implements Instruction.
func (i *CondJmpInstr) String() string {
	return fmt.Sprintf("%s %s, %s", i.opcode, i.cond, i.target)
}

// CallInstr represents a function call instruction.
type CallInstr struct {
	dest *Operand
	funcOp *Operand
	args []*Operand
}

// NewCallInstr creates a new call instruction.
func NewCallInstr(dest, funcOp *Operand, args []*Operand) *CallInstr {
	return &CallInstr{
		dest:  dest,
		funcOp: funcOp,
		args:  args,
	}
}

// Opcode implements Instruction.
func (i *CallInstr) Opcode() Opcode { return OpCall }

// Dest implements Instruction.
func (i *CallInstr) Dest() *Operand { return i.dest }

// Operands implements Instruction.
func (i *CallInstr) Operands() []*Operand {
	ops := []*Operand{i.funcOp}
	ops = append(ops, i.args...)
	return ops
}

// String implements Instruction.
func (i *CallInstr) String() string {
	args := make([]string, len(i.args))
	for j, arg := range i.args {
		args[j] = arg.String()
	}
	if i.dest != nil {
		return fmt.Sprintf("call %s = %s(%s)", i.dest, i.funcOp, strings.Join(args, ", "))
	}
	return fmt.Sprintf("call %s(%s)", i.funcOp, strings.Join(args, ", "))
}

// RetInstr represents a return instruction.
type RetInstr struct {
	value *Operand
}

// NewRetInstr creates a new return instruction.
func NewRetInstr(value *Operand) *RetInstr {
	return &RetInstr{
		value: value,
	}
}

// Opcode implements Instruction.
func (i *RetInstr) Opcode() Opcode { return OpRet }

// Dest implements Instruction.
func (i *RetInstr) Dest() *Operand { return nil }

// Operands implements Instruction.
func (i *RetInstr) Operands() []*Operand {
	if i.value != nil {
		return []*Operand{i.value}
	}
	return []*Operand{}
}

// String implements Instruction.
func (i *RetInstr) String() string {
	if i.value != nil {
		return fmt.Sprintf("ret %s", i.value)
	}
	return "ret"
}

// LabelInstr represents a label instruction.
type LabelInstr struct {
	label *Operand
}

//NewLabelInstr creates a new label instruction.
func NewLabelInstr(label *Operand) *LabelInstr {
	return &LabelInstr{
		label: label,
	}
}

// Opcode implements Instruction.
func (i *LabelInstr) Opcode() Opcode { return OpLabel }

// Dest implements Instruction.
func (i *LabelInstr) Dest() *Operand { return nil }

// Operands implements Instruction.
func (i *LabelInstr) Operands() []*Operand { return []*Operand{i.label} }

// String implements Instruction.
func (i *LabelInstr) String() string {
	return fmt.Sprintf("%s:", i.label)
}

// CastInstr represents a type cast instruction.
type CastInstr struct {
	opcode Opcode
	dest   *Operand
	operand *Operand
}

// NewCastInstr creates a new cast instruction.
func NewCastInstr(opcode Opcode, dest, operand *Operand) *CastInstr {
	return &CastInstr{
		opcode: opcode,
		dest:   dest,
		operand: operand,
	}
}

// Opcode implements Instruction.
func (i *CastInstr) Opcode() Opcode { return i.opcode }

// Dest implements Instruction.
func (i *CastInstr) Dest() *Operand { return i.dest }

// Operands implements Instruction.
func (i *CastInstr) Operands() []*Operand { return []*Operand{i.operand} }

// String implements Instruction.
func (i *CastInstr) String() string {
	return fmt.Sprintf("%s %s, %s", i.opcode, i.dest, i.operand)
}

// LeaInstr represents a load effective address instruction.
type LeaInstr struct {
	dest *Operand
	addr *Operand
}

// NewLeaInstr creates a new LEA instruction.
func NewLeaInstr(dest, addr *Operand) *LeaInstr {
	return &LeaInstr{
		dest: dest,
		addr: addr,
	}
}

// Opcode implements Instruction.
func (i *LeaInstr) Opcode() Opcode { return OpLea }

// Dest implements Instruction.
func (i *LeaInstr) Dest() *Operand { return i.dest }

// Operands implements Instruction.
func (i *LeaInstr) Operands() []*Operand { return []*Operand{i.addr} }

// String implements Instruction.
func (i *LeaInstr) String() string {
	return fmt.Sprintf("lea %s, %s", i.dest, i.addr)
}

// AllocInstr represents a stack allocation instruction.
type AllocInstr struct {
	dest *Operand
	size *Operand
}

// NewAllocInstr creates a new alloc instruction.
func NewAllocInstr(dest, size *Operand) *AllocInstr {
	return &AllocInstr{
		dest: dest,
		size: size,
	}
}

// Opcode implements Instruction.
func (i *AllocInstr) Opcode() Opcode { return OpAlloc }

// Dest implements Instruction.
func (i *AllocInstr) Dest() *Operand { return i.dest }

// Operands implements Instruction.
func (i *AllocInstr) Operands() []*Operand { return []*Operand{i.size} }

// String implements Instruction.
func (i *AllocInstr) String() string {
	return fmt.Sprintf("alloc %s, %s", i.dest, i.size)
}

// NopInstr represents a no-operation instruction.
type NopInstr struct{}

// NewNopInstr creates a new nop instruction.
func NewNopInstr() *NopInstr {
	return &NopInstr{}
}

// Opcode implements Instruction.
func (i *NopInstr) Opcode() Opcode { return OpNop }

// Dest implements Instruction.
func (i *NopInstr) Dest() *Operand { return nil }

// Operands implements Instruction.
func (i *NopInstr) Operands() []*Operand { return []*Operand{} }

// String implements Instruction.
func (i *NopInstr) String() string {
	return "nop"
}

// PhiInstr represents a SSA phi instruction.
type PhiInstr struct {
	dest   *Operand
	values []*Operand
	labels []*Operand
}

// NewPhiInstr creates a new phi instruction.
func NewPhiInstr(dest *Operand, values, labels []*Operand) *PhiInstr {
	return &PhiInstr{
		dest:   dest,
		values: values,
		labels: labels,
	}
}

// Opcode implements Instruction.
func (i *PhiInstr) Opcode() Opcode { return OpPhi }

// Dest implements Instruction.
func (i *PhiInstr) Dest() *Operand { return i.dest }

// Operands implements Instruction.
func (i *PhiInstr) Operands() []*Operand {
	ops := append(i.values, i.labels...)
	return ops
}

// String implements Instruction.
func (i *PhiInstr) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("phi %s = ", i.dest))
	for j, val := range i.values {
		if j > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("[%s, %s]", val, i.labels[j]))
	}
	return sb.String()
}