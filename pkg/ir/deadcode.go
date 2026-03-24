// Package ir provides intermediate representation for the GOC compiler.
// This file implements the Dead Code Elimination (DCE) optimization pass.
package ir

import (
	"fmt"
)

// DeadCodeElimination implements dead code elimination optimization.
// It removes:
// - Unreachable basic blocks from CFG
// - Unused local variables
// - Dead instructions (no side effects, results unused)
type DeadCodeElimination struct {
	BasePass
}

// NewDeadCodeElimination creates a new DCE pass.
func NewDeadCodeElimination() *DeadCodeElimination {
	return &DeadCodeElimination{
		BasePass: NewBasePass(PassInfo{
			Name:        "dead-code-elimination",
			Description: "Removes unreachable code, unused variables, and dead instructions",
			Phase:       PassPhaseMain,
			Dependencies: []PassDependency{
				{Name: "cfg-construction", Required: false},
			},
			Enabled: true,
		}),
	}
}

// Run executes the DCE pass on the given IR.
// Returns true if the IR was modified, false otherwise.
func (dce *DeadCodeElimination) Run(ir *IR) (bool, error) {
	if ir == nil {
		return false, fmt.Errorf("nil IR")
	}

	modified := false

	// Process each function
	for _, fn := range ir.Functions {
		fnModified, err := dce.processFunction(fn)
		if err != nil {
			return modified, err
		}
		if fnModified {
			modified = true
		}
	}

	return modified, nil
}

// Reset resets the DCE pass state.
func (dce *DeadCodeElimination) Reset() {
	// No state to reset
}

// processFunction processes a single function for dead code elimination.
func (dce *DeadCodeElimination) processFunction(fn *Function) (bool, error) {
	if fn == nil || len(fn.Blocks) == 0 {
		return false, nil
	}

	modified := false

	// Step 1: Remove unreachable basic blocks
	unreachableRemoved := dce.removeUnreachableBlocks(fn)
	if unreachableRemoved {
		modified = true
	}

	// Step 2: Remove dead instructions and track used temps
	usedTemps, deadInstrsRemoved := dce.removeDeadInstructions(fn)
	if deadInstrsRemoved {
		modified = true
	}

	// Step 3: Remove unused local variables
	unusedVarsRemoved := dce.removeUnusedLocalVars(fn, usedTemps)
	if unusedVarsRemoved {
		modified = true
	}

	return modified, nil
}

// removeUnreachableBlocks removes basic blocks that are not reachable from the entry block.
// Returns true if any blocks were removed.
func (dce *DeadCodeElimination) removeUnreachableBlocks(fn *Function) bool {
	if len(fn.Blocks) == 0 {
		return false
	}

	// Find entry block (first block or block with label "entry" or "start")
	entryBlock := fn.Blocks[0]
	for _, block := range fn.Blocks {
		if block.Label == "entry" || block.Label == "start" {
			entryBlock = block
			break
		}
	}

	// BFS/DFS to find all reachable blocks
	reachable := make(map[*BasicBlock]bool)
	stack := []*BasicBlock{entryBlock}
	reachable[entryBlock] = true

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		for _, succ := range current.Succs {
			if !reachable[succ] {
				reachable[succ] = true
				stack = append(stack, succ)
			}
		}
	}

	// Remove unreachable blocks
	originalCount := len(fn.Blocks)
	newBlocks := make([]*BasicBlock, 0, len(fn.Blocks))
	for _, block := range fn.Blocks {
		if reachable[block] {
			newBlocks = append(newBlocks, block)
		}
	}

	fn.Blocks = newBlocks
	return len(fn.Blocks) < originalCount
}

// removeDeadInstructions removes instructions whose results are never used.
// Returns a set of temp IDs that are actually used and whether any instructions were removed.
func (dce *DeadCodeElimination) removeDeadInstructions(fn *Function) (map[int]bool, bool) {
	// First pass: collect all used temps
	usedTemps := make(map[int]bool)

	// Multiple iterations until no more changes
	changed := true
	anyModified := false
	for changed {
		changed = false
		usedTemps = dce.collectUsedTemps(fn)
		changed = dce.eliminateDeadInstructions(fn, usedTemps)
		if changed {
			anyModified = true
		}
	}

	return usedTemps, anyModified
}

// collectUsedTemps collects all temp variables that are used as operands.
func (dce *DeadCodeElimination) collectUsedTemps(fn *Function) map[int]bool {
	usedTemps := make(map[int]bool)

	for _, block := range fn.Blocks {
		for _, instr := range block.Instrs {
			// Mark all operands as used
			for _, operand := range instr.Operands() {
				if operand != nil && operand.Kind == OperandTemp {
					if temp, ok := operand.Value.(*Temp); ok {
						usedTemps[temp.ID] = true
					}
				}
			}
		}
	}

	return usedTemps
}

// eliminateDeadInstructions removes instructions that define unused temps.
// Returns true if any instructions were removed.
func (dce *DeadCodeElimination) eliminateDeadInstructions(fn *Function, usedTemps map[int]bool) bool {
	modified := false

	for _, block := range fn.Blocks {
		newInstrs := make([]Instruction, 0, len(block.Instrs))
		for _, instr := range block.Instrs {
			// Check if instruction has side effects
			if dce.hasSideEffects(instr) {
				newInstrs = append(newInstrs, instr)
				continue
			}

			// Check if instruction defines a temp that is used
			dest := instr.Dest()
			if dest == nil {
				// No destination, keep the instruction (might have other effects)
				newInstrs = append(newInstrs, instr)
				continue
			}

			if dest.Kind == OperandTemp {
				if temp, ok := dest.Value.(*Temp); ok {
					if usedTemps[temp.ID] {
						// Temp is used, keep the instruction
						newInstrs = append(newInstrs, instr)
					} else {
						// Temp is not used, eliminate the instruction
						modified = true
					}
					continue
				}
			}

			// Keep other instructions
			newInstrs = append(newInstrs, instr)
		}
		block.Instrs = newInstrs
	}

	return modified
}

// hasSideEffects returns true if an instruction has side effects.
// Instructions with side effects cannot be eliminated even if their result is unused.
func (dce *DeadCodeElimination) hasSideEffects(instr Instruction) bool {
	switch instr.Opcode() {
	// Memory operations with side effects
	case OpStore, OpAlloc, OpFree:
		return true

	// Control flow operations
	case OpJmp, OpJmpIf, OpJmpUnless, OpRet, OpCall:
		return true

	// Label is not really an instruction with side effects, but we keep it for CFG structure
	case OpLabel:
		return true

	default:
		return false
	}
}

// removeUnusedLocalVars removes local variables that are never used.
// Returns true if any variables were removed.
func (dce *DeadCodeElimination) removeUnusedLocalVars(fn *Function, usedTemps map[int]bool) bool {
	// For now, we track local variable usage through temp operands
	// In a more sophisticated implementation, we would track direct local var references

	// Conservative: don't remove local vars in this pass
	// They can be removed in a more sophisticated analysis pass
	return false
}

// RegisterDeadCodeElimination registers the DCE pass with the global registry.
func RegisterDeadCodeElimination() {
	RegisterPass("dead-code-elimination", func() Pass {
		return NewDeadCodeElimination()
	})
}

// init automatically registers the DCE pass when the package is imported.
func init() {
	RegisterDeadCodeElimination()
}