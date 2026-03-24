// Package ir provides intermediate representation for the GOC compiler.
// This file contains advanced unit tests for the Dead Code Elimination pass.
package ir

import (
	"testing"

	"github.com/akzj/goc/pkg/parser"
)

func TestDeadCodeEliminationIdempotent(t *testing.T) {
	dce := NewDeadCodeElimination()

	// Create a function with dead code
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}

	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpAdd, &Operand{Kind: OperandTemp, Value: t0}, &Operand{Kind: OperandConst, Value: 1}, &Operand{Kind: OperandConst, Value: 2}), // dead
			NewRetInstr(&Operand{Kind: OperandConst, Value: 0}),
		},
		Preds: []*BasicBlock{},
		Succs: []*BasicBlock{},
	}

	fn := &Function{
		Name:       "test",
		ReturnType: &parser.BaseType{Kind: parser.TypeInt},
		Params:     []*Param{},
		Blocks:     []*BasicBlock{entry},
		LocalVars:  []*LocalVar{},
	}

	ir := &IR{
		Functions: []*Function{fn},
	}

	// First run
	modified1, err := dce.Run(ir)
	if err != nil {
		t.Fatalf("Unexpected error on first run: %v", err)
	}
	if !modified1 {
		t.Error("Expected modification on first run")
	}
	firstRunInstrCount := len(entry.Instrs)

	// Second run (should be idempotent)
	dce.Reset()
	modified2, err := dce.Run(ir)
	if err != nil {
		t.Fatalf("Unexpected error on second run: %v", err)
	}
	if modified2 {
		t.Error("Expected no modification on second run (idempotent)")
	}
	if len(entry.Instrs) != firstRunInstrCount {
		t.Errorf("Expected same instruction count, got %d vs %d", len(entry.Instrs), firstRunInstrCount)
	}
}

// TestDeadCodeEliminationIntegration tests DCE integration with PassManager.
func TestDeadCodeEliminationIntegration(t *testing.T) {
	// Register DCE pass
	RegisterDeadCodeElimination()

	registry := GetGlobalPassRegistry()
	_, ok := registry.Get("dead-code-elimination")
	if !ok {
		t.Fatal("Expected DCE pass to be registered")
	}

	// Create pass manager with DCE
	pm, err := NewPassManager(PassManagerConfig{
		Enabled:   true,
		PassNames: []string{"dead-code-elimination"},
	})
	if err != nil {
		t.Fatalf("Failed to create PassManager: %v", err)
	}

	// Create IR with dead code
	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}

	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpAdd, &Operand{Kind: OperandTemp, Value: t0}, &Operand{Kind: OperandConst, Value: 1}, &Operand{Kind: OperandConst, Value: 2}), // dead
			NewRetInstr(&Operand{Kind: OperandConst, Value: 0}),
		},
		Preds: []*BasicBlock{},
		Succs: []*BasicBlock{},
	}

	fn := &Function{
		Name:       "test",
		ReturnType: &parser.BaseType{Kind: parser.TypeInt},
		Params:     []*Param{},
		Blocks:     []*BasicBlock{entry},
		LocalVars:  []*LocalVar{},
	}

	ir := &IR{
		Functions: []*Function{fn},
	}

	modified, err := pm.Run(ir)
	if err != nil {
		t.Fatalf("PassManager.Run failed: %v", err)
	}
	if !modified {
		t.Error("Expected IR to be modified")
	}
	if len(entry.Instrs) != 1 {
		t.Errorf("Expected 1 instruction after DCE, got %d", len(entry.Instrs))
	}
}

// TestDeadCodeEliminationComplexCFG tests DCE on a complex CFG.
func TestDeadCodeEliminationComplexCFG(t *testing.T) {
	dce := NewDeadCodeElimination()

	// Create a CFG with multiple paths and unreachable blocks
	// entry -> block1 -> block2 (reachable)
	// entry -> block3 (unreachable)

	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewJmpInstr(&Operand{Kind: OperandLabel, Value: "block1"}),
		},
		Preds: []*BasicBlock{},
		Succs: []*BasicBlock{},
	}

	block1 := &BasicBlock{
		Label: "block1",
		Instrs: []Instruction{
			NewRetInstr(&Operand{Kind: OperandConst, Value: 0}),
		},
		Preds: []*BasicBlock{entry},
		Succs: []*BasicBlock{},
	}

	block2 := &BasicBlock{
		Label: "block2",
		Instrs: []Instruction{
			NewRetInstr(&Operand{Kind: OperandConst, Value: 1}),
		},
		Preds: []*BasicBlock{},
		Succs: []*BasicBlock{},
	}

	block3 := &BasicBlock{
		Label: "block3",
		Instrs: []Instruction{
			NewRetInstr(&Operand{Kind: OperandConst, Value: 2}),
		},
		Preds: []*BasicBlock{},
		Succs: []*BasicBlock{},
	}

	entry.Succs = []*BasicBlock{block1}

	fn := &Function{
		Name:       "test",
		ReturnType: &parser.BaseType{Kind: parser.TypeInt},
		Params:     []*Param{},
		Blocks:     []*BasicBlock{entry, block1, block2, block3},
		LocalVars:  []*LocalVar{},
	}

	ir := &IR{
		Functions: []*Function{fn},
	}

	modified, err := dce.Run(ir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !modified {
		t.Error("Expected IR to be modified (unreachable blocks removed)")
	}
	// entry, block1 are reachable; block2, block3 are not
	if len(fn.Blocks) != 2 {
		t.Errorf("Expected 2 blocks after DCE, got %d", len(fn.Blocks))
	}
}

// TestDeadCodeEliminationChainOfDeadCode tests elimination of chained dead code.
func TestDeadCodeEliminationChainOfDeadCode(t *testing.T) {
	dce := NewDeadCodeElimination()

	// Create a chain of dead instructions
	// t0 = add 1, 2 (dead)
	// t1 = add t0, 3 (dead because t0 is dead)
	// t2 = add t1, 4 (dead because t1 is dead)
	// ret 0

	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	t1 := &Temp{ID: 1, Type: &parser.BaseType{Kind: parser.TypeInt}}
	t2 := &Temp{ID: 2, Type: &parser.BaseType{Kind: parser.TypeInt}}

	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpAdd, &Operand{Kind: OperandTemp, Value: t0}, &Operand{Kind: OperandConst, Value: 1}, &Operand{Kind: OperandConst, Value: 2}),
			NewBinaryInstr(OpAdd, &Operand{Kind: OperandTemp, Value: t1}, &Operand{Kind: OperandTemp, Value: t0}, &Operand{Kind: OperandConst, Value: 3}),
			NewBinaryInstr(OpAdd, &Operand{Kind: OperandTemp, Value: t2}, &Operand{Kind: OperandTemp, Value: t1}, &Operand{Kind: OperandConst, Value: 4}),
			NewRetInstr(&Operand{Kind: OperandConst, Value: 0}),
		},
		Preds: []*BasicBlock{},
		Succs: []*BasicBlock{},
	}

	fn := &Function{
		Name:       "test",
		ReturnType: &parser.BaseType{Kind: parser.TypeInt},
		Params:     []*Param{},
		Blocks:     []*BasicBlock{entry},
		LocalVars:  []*LocalVar{},
	}

	ir := &IR{
		Functions: []*Function{fn},
	}

	modified, err := dce.Run(ir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !modified {
		t.Error("Expected IR to be modified")
	}
	// All three add instructions should be eliminated
	if len(entry.Instrs) != 1 {
		t.Errorf("Expected 1 instruction (ret only), got %d", len(entry.Instrs))
	}
	if entry.Instrs[0].Opcode() != OpRet {
		t.Errorf("Expected ret instruction, got %v", entry.Instrs[0].Opcode())
	}
}

// TestDeadCodeEliminationPreserveUsedTemps tests that used temps are preserved.
func TestDeadCodeEliminationPreserveUsedTemps(t *testing.T) {
	dce := NewDeadCodeElimination()

	// t0 = add 1, 2 (used)
	// t1 = add t0, 3 (used)
	// ret t1

	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	t1 := &Temp{ID: 1, Type: &parser.BaseType{Kind: parser.TypeInt}}

	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpAdd, &Operand{Kind: OperandTemp, Value: t0}, &Operand{Kind: OperandConst, Value: 1}, &Operand{Kind: OperandConst, Value: 2}),
			NewBinaryInstr(OpAdd, &Operand{Kind: OperandTemp, Value: t1}, &Operand{Kind: OperandTemp, Value: t0}, &Operand{Kind: OperandConst, Value: 3}),
			NewRetInstr(&Operand{Kind: OperandTemp, Value: t1}),
		},
		Preds: []*BasicBlock{},
		Succs: []*BasicBlock{},
	}

	fn := &Function{
		Name:       "test",
		ReturnType: &parser.BaseType{Kind: parser.TypeInt},
		Params:     []*Param{},
		Blocks:     []*BasicBlock{entry},
		LocalVars:  []*LocalVar{},
	}

	ir := &IR{
		Functions: []*Function{fn},
	}

	modified, err := dce.Run(ir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Nothing should be eliminated
	if modified {
		t.Error("Expected no modification (all instructions used)")
	}
	if len(entry.Instrs) != 3 {
		t.Errorf("Expected 3 instructions, got %d", len(entry.Instrs))
	}
}