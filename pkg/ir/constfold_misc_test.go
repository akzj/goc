// Package ir provides intermediate representation for the GOC compiler.
// This file contains unit tests for the Constant Folding pass - misc operations.
package ir

import (
	"testing"

	"github.com/akzj/goc/pkg/parser"
)

// TestConstantFoldingCreation tests creating a constant folding pass.
func TestConstantFoldingCreation(t *testing.T) {
	cf := NewConstantFolding()
	if cf == nil {
		t.Fatal("Expected non-nil ConstantFolding pass")
	}
	info := cf.Info()
	if info.Name != "constant-folding" {
		t.Errorf("Expected name 'constant-folding', got '%s'", info.Name)
	}
	if !info.Enabled {
		t.Error("Expected constant folding pass to be enabled")
	}
	if info.Phase != PassPhaseMain {
		t.Errorf("Expected phase PassPhaseMain, got %d", info.Phase)
	}
}

// TestConstantFoldingReset tests resetting the constant folding pass.
func TestConstantFoldingReset(t *testing.T) {
	cf := NewConstantFolding()
	cf.Reset()
}

// TestConstantFoldingNilIR tests handling of nil IR.
func TestConstantFoldingNilIR(t *testing.T) {
	cf := NewConstantFolding()
	modified, err := cf.Run(nil)
	if err == nil {
		t.Error("Expected error for nil IR")
	}
	if modified {
		t.Error("Expected no modification for nil IR")
	}
}

// TestConstantFoldingTypeConversion tests folding of type conversions.
func TestConstantFoldingTypeConversion(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeBool}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewCastInstr(OpCast, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeBool}},
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

// TestConstantFoldingIdempotent tests that constant folding is idempotent.
func TestConstantFoldingIdempotent(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpAdd, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandConst, Value: int64(2)},
				&Operand{Kind: OperandConst, Value: int64(3)}),
			NewRetInstr(&Operand{Kind: OperandTemp, Value: t0}),
		},
		Preds: []*BasicBlock{}, Succs: []*BasicBlock{},
	}
	fn := &Function{Name: "test", ReturnType: &parser.BaseType{Kind: parser.TypeInt}, Params: []*Param{}, Blocks: []*BasicBlock{entry}, LocalVars: []*LocalVar{}}
	ir := &IR{Functions: []*Function{fn}}
	modified1, err := cf.Run(ir)
	if err != nil {
		t.Fatalf("Unexpected error on first run: %v", err)
	}
	if !modified1 {
		t.Error("Expected first run to modify IR")
	}
	modified2, err := cf.Run(ir)
	if err != nil {
		t.Fatalf("Unexpected error on second run: %v", err)
	}
	if modified2 {
		t.Error("Expected second run to not modify IR (idempotent)")
	}
}

// TestConstantFoldingPreservesSemantics tests that folding preserves program semantics.
func TestConstantFoldingPreservesSemantics(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpSub, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandConst, Value: int64(10)},
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
	if operands[0].Value != int64(7) {
		t.Errorf("Expected result 7, got %v", operands[0].Value)
	}
}

// TestConstantFoldingNonConstantOperands tests that non-constant operands are not folded.
func TestConstantFoldingNonConstantOperands(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	t1 := &Temp{ID: 1, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpAdd, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandTemp, Value: t1},
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
	if modified {
		t.Error("Expected no modification when operands are not constant")
	}
}

// TestConstantFoldingRegistry tests registration with the global registry.
func TestConstantFoldingRegistry(t *testing.T) {
	registry := GetGlobalPassRegistry()
	constructor, ok := registry.Get("constant-folding")
	if !ok {
		t.Fatal("Expected constant-folding pass to be registered")
	}
	pass := constructor()
	if pass == nil {
		t.Fatal("Expected non-nil pass from constructor")
	}
	info := pass.Info()
	if info.Name != "constant-folding" {
		t.Errorf("Expected name 'constant-folding', got '%s'", info.Name)
	}
}

// TestConstantFoldingIntegrationConstantPropagation is an integration test showing constant propagation.
func TestConstantFoldingIntegrationConstantPropagation(t *testing.T) {
	cf := NewConstantFolding()
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	t1 := &Temp{ID: 1, Type: &parser.BaseType{Kind: parser.TypeInt}}
	t2 := &Temp{ID: 2, Type: &parser.BaseType{Kind: parser.TypeInt}}
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpAdd, &Operand{Kind: OperandTemp, Value: t0, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandConst, Value: int64(2)},
				&Operand{Kind: OperandConst, Value: int64(3)}),
			NewBinaryInstr(OpMul, &Operand{Kind: OperandTemp, Value: t1, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandTemp, Value: t0},
				&Operand{Kind: OperandConst, Value: int64(4)}),
			NewBinaryInstr(OpSub, &Operand{Kind: OperandTemp, Value: t2, Type: &parser.BaseType{Kind: parser.TypeInt}},
				&Operand{Kind: OperandTemp, Value: t1},
				&Operand{Kind: OperandConst, Value: int64(10)}),
			NewRetInstr(&Operand{Kind: OperandTemp, Value: t2}),
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
	foldedCount := 0
	for _, instr := range entry.Instrs {
		if binstr, ok := instr.(*BinaryInstr); ok {
			ops := binstr.Operands()
			if len(ops) >= 1 && ops[0].Kind == OperandConst {
				foldedCount++
			}
		}
	}
	if foldedCount != 3 {
		t.Errorf("Expected 3 folded instructions, got %d", foldedCount)
	}
	lastInstr := entry.Instrs[2]
	if binstr, ok := lastInstr.(*BinaryInstr); ok {
		ops := binstr.Operands()
		if ops[0].Value != int64(10) {
			t.Errorf("Expected final result 10, got %v", ops[0].Value)
		}
	}
}