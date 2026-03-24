// Package ir provides intermediate representation for the GOC compiler.
// This file defines IR structures.
package ir

import (
	"fmt"
	"strings"

	"github.com/akzj/goc/pkg/parser"
)

// IR represents the intermediate representation.
type IR struct {
	// Functions is the list of functions.
	Functions []*Function
	// Globals is the list of global variables.
	Globals []*GlobalVar
	// Constants is the list of constants.
	Constants []*Constant
}

// Function represents a function in IR.
type Function struct {
	// Name is the function name.
	Name string
	// ReturnType is the return type.
	ReturnType parser.Type
	// Params is the list of parameters.
	Params []*Param
	// Blocks is the list of basic blocks.
	Blocks []*BasicBlock
	// LocalVars is the list of local variables.
	LocalVars []*LocalVar
}

// BasicBlock represents a basic block in CFG.
type BasicBlock struct {
	// Label is the block label.
	Label string
	// Instrs is the list of instructions.
	Instrs []Instruction
	// Preds is the list of predecessor blocks.
	Preds []*BasicBlock
	// Succs is the list of successor blocks.
	Succs []*BasicBlock
}

// Param represents a function parameter.
type Param struct {
	// Name is the parameter name.
	Name string
	// Type is the parameter type.
	Type parser.Type
}

// LocalVar represents a local variable.
type LocalVar struct {
	// Name is the variable name.
	Name string
	// Type is the variable type.
	Type parser.Type
	// StackOffset is the stack offset.
	StackOffset int64
}

// GlobalVar represents a global variable.
type GlobalVar struct {
	// Name is the variable name.
	Name string
	// Type is the variable type.
	Type parser.Type
	// Init is the initializer (nil if uninitialized).
	Init parser.Expr
}

// Constant represents a constant.
type Constant struct {
	// Name is the constant name.
	Name string
	// Value is the constant value.
	Value interface{}
}

// String returns a string representation of the IR.
func (ir *IR) String() string {
	var sb strings.Builder
	sb.WriteString("IR {\n")

	// Emit globals
	if len(ir.Globals) > 0 {
		sb.WriteString("  Globals:\n")
		for _, g := range ir.Globals {
			sb.WriteString(fmt.Sprintf("    %s\n", g.String()))
		}
	}

	// Emit constants
	if len(ir.Constants) > 0 {
		sb.WriteString("  Constants:\n")
		for _, c := range ir.Constants {
			sb.WriteString(fmt.Sprintf("    %s\n", c.String()))
		}
	}

	// Emit functions
	if len(ir.Functions) > 0 {
		sb.WriteString("  Functions:\n")
		for _, fn := range ir.Functions {
			sb.WriteString(fmt.Sprintf("    %s\n", fn.String()))
		}
	}

	sb.WriteString("}")
	return sb.String()
}

// String returns a string representation of the function.
func (fn *Function) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Function %s(", fn.Name))
	for i, param := range fn.Params {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(param.String())
	}
	sb.WriteString(")")
	if fn.ReturnType != nil {
		sb.WriteString(fmt.Sprintf(" -> %s", fn.ReturnType.String()))
	}
	sb.WriteString(" {\n")
	for _, block := range fn.Blocks {
		sb.WriteString(fmt.Sprintf("    %s\n", block.String()))
	}
	sb.WriteString("  }")
	return sb.String()
}

// String returns a string representation of the basic block.
func (bb *BasicBlock) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Block %s:", bb.Label))
	for _, instr := range bb.Instrs {
		sb.WriteString(fmt.Sprintf("\n      %s", instr.String()))
	}
	return sb.String()
}

// String returns a string representation of the parameter.
func (p *Param) String() string {
	if p.Type != nil {
		return fmt.Sprintf("%s %s", p.Type.String(), p.Name)
	}
	return p.Name
}

// String returns a string representation of the local variable.
func (lv *LocalVar) String() string {
	if lv.Type != nil {
		return fmt.Sprintf("%s %s (offset: %d)", lv.Type.String(), lv.Name, lv.StackOffset)
	}
	return fmt.Sprintf("%s (offset: %d)", lv.Name, lv.StackOffset)
}

// String returns a string representation of the global variable.
func (gv *GlobalVar) String() string {
	if gv.Type != nil {
		if gv.Init != nil {
			return fmt.Sprintf("global %s %s = %s", gv.Type.String(), gv.Name, gv.Init.String())
		}
		return fmt.Sprintf("global %s %s", gv.Type.String(), gv.Name)
	}
	return fmt.Sprintf("global %s", gv.Name)
}

// String returns a string representation of the constant.
func (c *Constant) String() string {
	return fmt.Sprintf("const %s = %v", c.Name, c.Value)
}