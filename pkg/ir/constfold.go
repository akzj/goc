// Package ir provides intermediate representation for the GOC compiler.
// This file implements the Constant Folding optimization pass.
package ir

import (
	"fmt"

	"github.com/akzj/goc/pkg/parser"
)

// ConstantFolding implements constant folding optimization.
type ConstantFolding struct {
	BasePass
	constValues map[*Temp]interface{}
}

// NewConstantFolding creates a new constant folding pass.
func NewConstantFolding() *ConstantFolding {
	return &ConstantFolding{
		BasePass: NewBasePass(PassInfo{
			Name: "constant-folding", Description: "Evaluates constant expressions at compile-time",
			Phase: PassPhaseMain, Dependencies: []PassDependency{{Name: "dead-code-elimination", Required: false}},
			Enabled: true,
		}),
		constValues: make(map[*Temp]interface{}),
	}
}

// Run executes the constant folding pass on the given IR.
func (cf *ConstantFolding) Run(ir *IR) (bool, error) {
	if ir == nil {
		return false, fmt.Errorf("nil IR")
	}
	modified := false
	for _, fn := range ir.Functions {
		fnModified, err := cf.processFunction(fn)
		if err != nil {
			return modified, err
		}
		if fnModified {
			modified = true
		}
	}
	return modified, nil
}

// Reset resets the constant folding pass state.
func (cf *ConstantFolding) Reset() {
	cf.constValues = make(map[*Temp]interface{})
}

// processFunction processes a single function for constant folding.
func (cf *ConstantFolding) processFunction(fn *Function) (bool, error) {
	if fn == nil || len(fn.Blocks) == 0 {
		return false, nil
	}
	modified := false
	cf.constValues = make(map[*Temp]interface{})
	changed := true
	for changed {
		changed = false
		for _, block := range fn.Blocks {
			if cf.foldBlock(block) {
				changed = true
				modified = true
			}
		}
	}
	return modified, nil
}

// foldBlock folds constants in a basic block.
func (cf *ConstantFolding) foldBlock(block *BasicBlock) bool {
	modified := false
	for i, instr := range block.Instrs {
		cf.propagateConstants(instr)
		if cf.foldInstructionInPlace(instr) {
			modified = true
			block.Instrs[i] = instr
			if dest := instr.Dest(); dest != nil {
				if temp, ok := dest.Value.(*Temp); ok {
					if binstr, ok := instr.(*BinaryInstr); ok && binstr.left != nil && binstr.left.Kind == OperandConst {
						cf.constValues[temp] = binstr.left.Value
					} else if uinstr, ok := instr.(*UnaryInstr); ok && uinstr.operand != nil && uinstr.operand.Kind == OperandConst {
						cf.constValues[temp] = uinstr.operand.Value
					} else if cinstr, ok := instr.(*CastInstr); ok && cinstr.operand != nil && cinstr.operand.Kind == OperandConst {
						cf.constValues[temp] = cinstr.operand.Value
					}
				}
			}
		}
	}
	return modified
}

// propagateConstants replaces temp operands with their constant values if known.
func (cf *ConstantFolding) propagateConstants(instr Instruction) {
	for _, op := range instr.Operands() {
		if op != nil && op.Kind == OperandTemp {
			if temp, ok := op.Value.(*Temp); ok {
				if constVal, exists := cf.constValues[temp]; exists {
					op.Kind = OperandConst
					op.Value = constVal
				}
			}
		}
	}
}

// foldInstructionInPlace folds a single instruction in-place.
func (cf *ConstantFolding) foldInstructionInPlace(instr Instruction) bool {
	switch i := instr.(type) {
	case *BinaryInstr:
		return cf.foldBinaryInstrInPlace(i)
	case *UnaryInstr:
		return cf.foldUnaryInstrInPlace(i)
	case *CastInstr:
		return cf.foldCastInstrInPlace(i)
	default:
		return false
	}
}

// foldBinaryInstrInPlace folds binary instructions with constant operands.
func (cf *ConstantFolding) foldBinaryInstrInPlace(instr *BinaryInstr) bool {
	if instr.left == nil || instr.right == nil || instr.left.Kind != OperandConst || instr.right.Kind != OperandConst {
		return false
	}
	if instr.opcode == OpNop {
		return false
	}
	result, ok := cf.foldBinaryOp(instr.opcode, instr.left.Value, instr.right.Value, instr.dest)
	if !ok {
		return false
	}
	instr.opcode = OpNop
	instr.left = &Operand{Kind: OperandConst, Value: result, Type: instr.dest.Type}
	instr.right = nil
	return true
}

// foldUnaryInstrInPlace folds unary instructions with constant operands.
func (cf *ConstantFolding) foldUnaryInstrInPlace(instr *UnaryInstr) bool {
	if instr.operand == nil || instr.operand.Kind != OperandConst || instr.opcode == OpNop {
		return false
	}
	result, ok := cf.foldUnaryOp(instr.opcode, instr.operand.Value, instr.dest)
	if !ok {
		return false
	}
	instr.opcode = OpNop
	instr.operand = &Operand{Kind: OperandConst, Value: result, Type: instr.dest.Type}
	return true
}

// foldCastInstrInPlace folds cast instructions with constant operands.
func (cf *ConstantFolding) foldCastInstrInPlace(instr *CastInstr) bool {
	if instr.operand == nil || instr.operand.Kind != OperandConst || instr.opcode == OpNop {
		return false
	}
	result, ok := cf.foldCastOp(instr.opcode, instr.operand.Value, instr.dest.Type)
	if !ok {
		return false
	}
	instr.opcode = OpNop
	instr.operand = &Operand{Kind: OperandConst, Value: result, Type: instr.dest.Type}
	return true
}

// foldBinaryOp folds a binary operation on constants.
func (cf *ConstantFolding) foldBinaryOp(opcode Opcode, left, right interface{}, dest *Operand) (interface{}, bool) {
	l, r := cf.toInt64(left), cf.toInt64(right)
	if l == nil || r == nil {
		return nil, false
	}
	switch opcode {
	case OpAdd:
		return *l + *r, true
	case OpSub:
		return *l - *r, true
	case OpMul:
		return *l * *r, true
	case OpDiv:
		if *r == 0 {
			return nil, false
		}
		return *l / *r, true
	case OpMod:
		if *r == 0 {
			return nil, false
		}
		return *l % *r, true
	case OpEq:
		return int64(bi(*l == *r)), true
	case OpNe:
		return int64(bi(*l != *r)), true
	case OpLt:
		return int64(bi(*l < *r)), true
	case OpLe:
		return int64(bi(*l <= *r)), true
	case OpGt:
		return int64(bi(*l > *r)), true
	case OpGe:
		return int64(bi(*l >= *r)), true
	case OpAnd:
		return *l & *r, true
	case OpOr:
		return *l | *r, true
	case OpBitXor:
		return *l ^ *r, true
	}
	return nil, false
}

// foldUnaryOp folds a unary operation on constants.
func (cf *ConstantFolding) foldUnaryOp(opcode Opcode, val interface{}, dest *Operand) (interface{}, bool) {
	v := cf.toInt64(val)
	if v == nil {
		return nil, false
	}
	switch opcode {
	case OpNeg:
		return -*v, true
	case OpNot:
		return int64(bi(*v == 0)), true
	case OpBitNot:
		return ^*v, true
	}
	return nil, false
}

// foldCastOp folds a cast operation on constants.
func (cf *ConstantFolding) foldCastOp(opcode Opcode, val interface{}, destType parser.Type) (interface{}, bool) {
	v := cf.toInt64(val)
	if v == nil {
		return nil, false
	}
	switch opcode {
	case OpCast:
		if destType != nil && destType.TypeKind() == parser.TypeBool {
			return int64(bi(*v != 0)), true
		}
		return *v, true
	case OpZeroExt, OpSignExt:
		return *v, true
	case OpTrunc:
		if destType == nil {
			return *v, true
		}
		switch destType.Size() {
		case 1:
			return int64(int8(*v)), true
		case 2:
			return int64(int16(*v)), true
		case 4:
			return int64(int32(*v)), true
		}
		return *v, true
	}
	return nil, false
}

// toInt64 converts a value to int64 for computation.
func (cf *ConstantFolding) toInt64(val interface{}) *int64 {
	var r int64
	switch v := val.(type) {
	case int64:
		r = v
	case int:
		r = int64(v)
	case int32, int16, int8:
		r = int64(v.(int))
	case uint64, uint, uint32, uint16, uint8:
		r = int64(v.(int))
	case float64:
		r = int64(v)
	case float32:
		r = int64(v)
	default:
		return nil
	}
	return &r
}

// bi converts bool to int (0 or 1).
func bi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// isIntegerType checks if a type is an integer type.
func isIntegerType(t parser.Type) bool {
	if t == nil {
		return false
	}
	switch t.TypeKind() {
	case parser.TypeBool, parser.TypeChar, parser.TypeShort, parser.TypeInt, parser.TypeLong:
		return true
	}
	return false
}

// RegisterConstantFolding registers the constant folding pass with the global registry.
func RegisterConstantFolding() {
	RegisterPass("constant-folding", func() Pass { return NewConstantFolding() })
}

func init() {
	RegisterConstantFolding()
}