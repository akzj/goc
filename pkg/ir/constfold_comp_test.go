// Package ir provides intermediate representation for the GOC compiler.
// This file contains unit tests for the Constant Folding pass - comparison and logical operations.
package ir

import (
	"testing"

	"github.com/akzj/goc/pkg/parser"
)

// TestConstantFoldingComparisonEq tests folding of equality comparison.
func TestConstantFoldingComparisonEq(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpEq, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandConst, Value: int64(5)},
				&Operand{Kind: OperandConst, Value: int64(5)}),
			NewRetInstr(&Operand{Kind: OperandTemp, Value: t0}),
		},
		Preds: []*BasicBlock{}, Succs: []*BasicBlock{},
	}
	fn := &Function{Name: "test", ReturnType: &parser.BaseType{Kind: parser.TypeInt}, Params: []*Param{}, Blocks: []*BasicBlock{entry}, LocalVars: []*LocalVar{}}
	ir := &IR{Functions: []*Function{fn}}
	modified, err := cf.Run(ir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !modified {
		t.Error("Expected IR to be modified")
	}
	operands := entry.Instrs[0].Operands()
	if operands[0].Value != int64(1) {
		t.Errorf("Expected result 1 (true), got %v", operands[0].Value)
	}
}

// TestConstantFoldingComparisonNe tests folding of not-equal comparison.
func TestConstantFoldingComparisonNe(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpNe, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandConst, Value: int64(5)},
				&Operand{Kind: OperandConst, Value: int64(3)}),
			NewRetInstr(&Operand{Kind: OperandTemp, Value: t0}),
		},
		Preds: []*BasicBlock{}, Succs: []*BasicBlock{},
	}
	fn := &Function{Name: "test", ReturnType: &parser.BaseType{Kind: parser.TypeInt}, Params: []*Param{}, Blocks: []*BasicBlock{entry}, LocalVars: []*LocalVar{}}
	ir := &IR{Functions: []*Function{fn}}
	modified, err := cf.Run(ir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !modified {
		t.Error("Expected IR to be modified")
	}
	operands := entry.Instrs[0].Operands()
	if operands[0].Value != int64(1) {
		t.Errorf("Expected result 1 (true), got %v", operands[0].Value)
	}
}

// TestConstantFoldingComparisonLt tests folding of less-than comparison.
func TestConstantFoldingComparisonLt(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpLt, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandConst, Value: int64(3)},
				&Operand{Kind: OperandConst, Value: int64(5)}),
			NewRetInstr(&Operand{Kind: OperandTemp, Value: t0}),
		},
		Preds: []*BasicBlock{}, Succs: []*BasicBlock{},
	}
	fn := &Function{Name: "test", ReturnType: &parser.BaseType{Kind: parser.TypeInt}, Params: []*Param{}, Blocks: []*BasicBlock{entry}, LocalVars: []*LocalVar{}}
	ir := &IR{Functions: []*Function{fn}}
	modified, err := cf.Run(ir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !modified {
		t.Error("Expected IR to be modified")
	}
	operands := entry.Instrs[0].Operands()
	if operands[0].Value != int64(1) {
		t.Errorf("Expected result 1 (true), got %v", operands[0].Value)
	}
}

// TestConstantFoldingLogicalAnd tests folding of bitwise AND.
func TestConstantFoldingLogicalAnd(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpAnd, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandConst, Value: int64(0b1100)},
				&Operand{Kind: OperandConst, Value: int64(0b1010)}),
			NewRetInstr(&Operand{Kind: OperandTemp, Value: t0}),
		},
		Preds: []*BasicBlock{}, Succs: []*BasicBlock{},
	}
	fn := &Function{Name: "test", ReturnType: &parser.BaseType{Kind: parser.TypeInt}, Params: []*Param{}, Blocks: []*BasicBlock{entry}, LocalVars: []*LocalVar{}}
	ir := &IR{Functions: []*Function{fn}}
	modified, err := cf.Run(ir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !modified {
		t.Error("Expected IR to be modified")
	}
	operands := entry.Instrs[0].Operands()
	if operands[0].Value != int64(0b1000) {
		t.Errorf("Expected result 8 (0b1000), got %v", operands[0].Value)
	}
}

// TestConstantFoldingLogicalOr tests folding of bitwise OR.
func TestConstantFoldingLogicalOr(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpOr, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandConst, Value: int64(0b1100)},
				&Operand{Kind: OperandConst, Value: int64(0b0011)}),
			NewRetInstr(&Operand{Kind: OperandTemp, Value: t0}),
		},
		Preds: []*BasicBlock{}, Succs: []*BasicBlock{},
	}
	fn := &Function{Name: "test", ReturnType: &parser.BaseType{Kind: parser.TypeInt}, Params: []*Param{}, Blocks: []*BasicBlock{entry}, LocalVars: []*LocalVar{}}
	ir := &IR{Functions: []*Function{fn}}
	modified, err := cf.Run(ir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !modified {
		t.Error("Expected IR to be modified")
	}
	operands := entry.Instrs[0].Operands()
	if operands[0].Value != int64(0b1111) {
		t.Errorf("Expected result 15 (0b1111), got %v", operands[0].Value)
	}
}

// TestConstantFoldingLogicalNot tests folding of logical NOT.
func TestConstantFoldingLogicalNot(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewUnaryInstr(OpNot, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandConst, Value: int64(0)}),
			NewRetInstr(&Operand{Kind: OperandTemp, Value: t0}),
		},
		Preds: []*BasicBlock{}, Succs: []*BasicBlock{},
	}
	fn := &Function{Name: "test", ReturnType: &parser.BaseType{Kind: parser.TypeInt}, Params: []*Param{}, Blocks: []*BasicBlock{entry}, LocalVars: []*LocalVar{}}
	ir := &IR{Functions: []*Function{fn}}
	modified, err := cf.Run(ir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !modified {
		t.Error("Expected IR to be modified")
	}
	operands := entry.Instrs[0].Operands()
	if operands[0].Value != int64(1) {
		t.Errorf("Expected result 1 (NOT 0), got %v", operands[0].Value)
	}
}