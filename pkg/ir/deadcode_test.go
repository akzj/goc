// Package ir provides intermediate representation for the GOC compiler.
// This file contains unit tests for the Dead Code Elimination pass.
package ir

import (
	"testing"

	"github.com/akzj/goc/pkg/parser"
)

// TestDeadCodeEliminationCreation tests creating a DCE pass.
func TestDeadCodeEliminationCreation(t *testing.T) {
	dce := NewDeadCodeElimination()

	if dce == nil {
		t.Fatal("Expected non-nil DCE pass")
	}

	info := dce.Info()
	if info.Name != "dead-code-elimination" {
		t.Errorf("Expected name 'dead-code-elimination', got '%s'", info.Name)
	}
	if !info.Enabled {
		t.Error("Expected DCE pass to be enabled")
	}
	if info.Phase != PassPhaseMain {
		t.Errorf("Expected phase PassPhaseMain, got %d", info.Phase)
	}
}

// TestDeadCodeEliminationReset tests resetting the DCE pass.
func TestDeadCodeEliminationReset(t *testing.T) {
	dce := NewDeadCodeElimination()

	// Should not panic
	dce.Reset()
}

// TestDeadCodeEliminationNilIR tests handling of nil IR.
func TestDeadCodeEliminationNilIR(t *testing.T) {
	dce := NewDeadCodeElimination()

	modified, err := dce.Run(nil)
	if err == nil {
		t.Error("Expected error for nil IR")
	}
	if modified {
		t.Error("Expected no modification for nil IR")
	}
}

// TestDeadCodeEliminationRemoveUnreachableBlocks tests removing unreachable blocks.
func TestDeadCodeEliminationRemoveUnreachableBlocks(t *testing.T) {
	dce := NewDeadCodeElimination()

	// Create a function with reachable and unreachable blocks
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewRetInstr(&Operand{Kind: OperandConst, Value: 0}),
		},
		Preds: []*BasicBlock{},
		Succs: []*BasicBlock{},
	}

	unreachable := &BasicBlock{
		Label: "unreachable",
		Instrs: []Instruction{
			NewRetInstr(&Operand{Kind: OperandConst, Value: 1}),
		},
		Preds: []*BasicBlock{},
		Succs: []*BasicBlock{},
	}

	fn := &Function{
		Name:       "test",
		ReturnType: &parser.BaseType{Kind: parser.TypeInt},
		Params:     []*Param{},
		Blocks:     []*BasicBlock{entry, unreachable},
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
		t.Error("Expected IR to be modified (unreachable block removed)")
	}
	if len(fn.Blocks) != 1 {
		t.Errorf("Expected 1 block after DCE, got %d", len(fn.Blocks))
	}
	if fn.Blocks[0].Label != "entry" {
		t.Errorf("Expected entry block, got '%s'", fn.Blocks[0].Label)
	}
}

// TestDeadCodeEliminationRemoveDeadInstructions tests removing dead instructions.
func TestDeadCodeEliminationRemoveDeadInstructions(t *testing.T) {
	dce := NewDeadCodeElimination()

	// Create a function with dead and live instructions
	// t0 = add 1, 2 (dead - result unused)
	// t1 = add 3, 4 (live - used in return)
	// ret t1

	t0 := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}
	t1 := &Temp{ID: 1, Type: &parser.BaseType{Kind: parser.TypeInt}}

	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewBinaryInstr(OpAdd, &Operand{Kind: OperandTemp, Value: t0}, &Operand{Kind: OperandConst, Value: 1}, &Operand{Kind: OperandConst, Value: 2}), // dead
			NewBinaryInstr(OpAdd, &Operand{Kind: OperandTemp, Value: t1}, &Operand{Kind: OperandConst, Value: 3}, &Operand{Kind: OperandConst, Value: 4}), // live
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
	if !modified {
		t.Error("Expected IR to be modified (dead instruction removed)")
	}
	if len(entry.Instrs) != 2 {
		t.Errorf("Expected 2 instructions after DCE, got %d", len(entry.Instrs))
	}
	// First instruction should be the live add, second should be ret
	if entry.Instrs[0].Opcode() != OpAdd {
		t.Errorf("Expected first instruction to be add, got %v", entry.Instrs[0].Opcode())
	}
	if entry.Instrs[1].Opcode() != OpRet {
		t.Errorf("Expected second instruction to be ret, got %v", entry.Instrs[1].Opcode())
	}
}

// TestDeadCodeEliminationPreserveSideEffects tests that instructions with side effects are preserved.
func TestDeadCodeEliminationPreserveSideEffects(t *testing.T) {
	dce := NewDeadCodeElimination()

	// Create a function with side-effect instructions that should be preserved
	// store const, addr (side effect - should be preserved)
	// ret 0

	addr := &Temp{ID: 0, Type: &parser.BaseType{Kind: parser.TypeInt}}

	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewStoreInstr(&Operand{Kind: OperandConst, Value: 42}, &Operand{Kind: OperandTemp, Value: addr}), // side effect - preserve
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
	// Store should be preserved, so no modification expected
	if modified {
		t.Error("Expected no modification (store has side effects)")
	}
	if len(entry.Instrs) != 2 {
		t.Errorf("Expected 2 instructions (store preserved), got %d", len(entry.Instrs))
	}
	if entry.Instrs[0].Opcode() != OpStore {
		t.Errorf("Expected first instruction to be store, got %v", entry.Instrs[0].Opcode())
	}
}

// TestDeadCodeEliminationPreserveControlFlow tests that control flow instructions are preserved.
func TestDeadCodeEliminationPreserveControlFlow(t *testing.T) {
	dce := NewDeadCodeElimination()

	// Create blocks with control flow
	entry := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			NewJmpInstr(&Operand{Kind: OperandLabel, Value: "block2"}),
		},
		Preds: []*BasicBlock{},
		Succs: []*BasicBlock{},
	}

	block2 := &BasicBlock{
		Label: "block2",
		Instrs: []Instruction{
			NewRetInstr(&Operand{Kind: OperandConst, Value: 0}),
		},
		Preds: []*BasicBlock{entry},
		Succs: []*BasicBlock{},
	}

	entry.Succs = []*BasicBlock{block2}

	fn := &Function{
		Name:       "test",
		ReturnType: &parser.BaseType{Kind: parser.TypeInt},
		Params:     []*Param{},
		Blocks:     []*BasicBlock{entry, block2},
		LocalVars:  []*LocalVar{},
	}

	ir := &IR{
		Functions: []*Function{fn},
	}

	modified, err := dce.Run(ir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// No dead code to eliminate
	if modified {
		t.Error("Expected no modification")
	}
	if len(fn.Blocks) != 2 {
		t.Errorf("Expected 2 blocks, got %d", len(fn.Blocks))
	}
}

// TestDeadCodeEliminationIdempotent tests that DCE is idempotent.
