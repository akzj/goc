// Package ir provides intermediate representation for the GOC compiler.
// This file defines IR structures.
package ir

import (
	"github.com/akzj/goc/pkg/parser"
)

// TODO: Implement IR package
// Reference: docs/architecture-design-phases-2-7.md Section 5

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