// Package ir provides intermediate representation for the GOC compiler.
// This file contains unit tests for the Constant Folding pass - arithmetic operations.
package ir

import (
	"testing"

	"github.com/akzj/goc/pkg/parser"
)

// TestConstantFoldingArithmeticAdd tests folding of constant addition.
func TestConstantFoldingArithmeticAdd(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpAdd, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
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
		t.Error("Expected IR to be modified (constant folded)")
	}
	if len(entry.Instrs) != 2 {
		t.Errorf("Expected 2 instructions, got %d", len(entry.Instrs))
	}
	firstInstr := entry.Instrs[0]
	if firstInstr.Opcode() != OpNop {
		t.Errorf("Expected opcode OpNop after folding, got %v", firstInstr.Opcode())
	}
	operands := firstInstr.Operands()
	if len(operands) < 1 || operands[0].Kind != OperandConst || operands[0].Value != int64(8) {
		t.Errorf("Expected result 8, got %v", operands[0].Value)
	}
}

// TestConstantFoldingArithmeticSub tests folding of constant subtraction.
func TestConstantFoldingArithmeticSub(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpSub, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandConst, Value: int64(10)},
				&Operand{Kind: OperandConst, Value: int64(4)}),
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
	if operands[0].Value != int64(6) {
		t.Errorf("Expected result 6, got %v", operands[0].Value)
	}
}

// TestConstantFoldingArithmeticMul tests folding of constant multiplication.
func TestConstantFoldingArithmeticMul(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpMul, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandConst, Value: int64(7)},
				&Operand{Kind: OperandConst, Value: int64(6)}),
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
	if operands[0].Value != int64(42) {
		t.Errorf("Expected result 42, got %v", operands[0].Value)
	}
}

// TestConstantFoldingArithmeticDiv tests folding of constant division.
func TestConstantFoldingArithmeticDiv(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpDiv, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandConst, Value: int64(20)},
				&Operand{Kind: OperandConst, Value: int64(4)}),
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
	if operands[0].Value != int64(5) {
		t.Errorf("Expected result 5, got %v", operands[0].Value)
	}
}

// TestConstantFoldingArithmeticMod tests folding of constant modulo.
func TestConstantFoldingArithmeticMod(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpMod, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandConst, Value: int64(17)},
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
	if operands[0].Value != int64(2) {
		t.Errorf("Expected result 2, got %v", operands[0].Value)
	}
}

// TestConstantFoldingOverflowWrapAround tests overflow handling with wrap-around.
func TestConstantFoldingOverflowWrapAround(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpAdd, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandConst, Value: int64(9223372036854775807)},
				&Operand{Kind: OperandConst, Value: int64(1)}),
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
	expected := int64(-9223372036854775808)
	if operands[0].Value != expected {
		t.Errorf("Expected wrap-around to %d, got %v", expected, operands[0].Value)
	}
}

// TestConstantFoldingDivisionByZero tests that division by zero is not folded.
func TestConstantFoldingDivisionByZero(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpDiv, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandConst, Value: int64(10)},
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
	if modified {
		t.Error("Expected no modification for division by zero")
	}
}