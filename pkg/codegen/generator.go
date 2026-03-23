// Package codegen generates x86-64 assembly code from IR.
// This file defines the code generator interface and implementation.
package codegen

import (
	"fmt"
	"strings"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/ir"
	"github.com/akzj/goc/pkg/parser"
)

// CodeGenerator generates x86-64 assembly from IR.
// It follows the System V AMD64 ABI calling convention.
type CodeGenerator struct {
	// ir is the IR to generate code for.
	ir *ir.IR
	// errors is the error handler.
	errors *errhand.ErrorHandler
	// output is the assembly output buffer.
	output *strings.Builder
	// regAlloc is the register allocator.
	regAlloc *RegisterAllocator
	// stackFrame is the current stack frame.
	stackFrame *StackFrame
	// labelCounter is used to generate unique labels.
	labelCounter int
	// currentFunc is the function currently being compiled.
	currentFunc *ir.Function
}

// StackFrame represents a function stack frame.
type StackFrame struct {
	// Size is the total frame size.
	Size int64
	// Locals is the list of local variables.
	Locals []*ir.LocalVar
	// SavedRegs is the list of saved registers.
	SavedRegs []Reg
}

// NewCodeGenerator creates a new code generator.
func NewCodeGenerator(errorHandler *errhand.ErrorHandler) *CodeGenerator {
	return &CodeGenerator{
		errors:       errorHandler,
		output:       &strings.Builder{},
		regAlloc:     NewRegisterAllocator(),
		labelCounter: 0,
	}
}

// Generate generates assembly code for the entire IR.
func (cg *CodeGenerator) Generate(ir *ir.IR) (string, error) {
	cg.ir = ir
	cg.output.Reset()

	// Emit file header
	cg.emitHeader()

	// Emit global variables
	for _, global := range ir.Globals {
		cg.emitGlobal(global)
	}

	// Emit constants
	for _, constant := range ir.Constants {
		cg.emitConstant(constant)
	}

	// Emit functions
	for _, fn := range ir.Functions {
		asm, err := cg.GenerateFunction(fn)
		if err != nil {
			return "", fmt.Errorf("generating function %s: %w", fn.Name, err)
		}
		cg.output.WriteString(asm)
	}

	return cg.output.String(), nil
}

// GenerateFunction generates assembly for a single function.
func (cg *CodeGenerator) GenerateFunction(fn *ir.Function) (string, error) {
	cg.currentFunc = fn
	cg.regAlloc.Reset()

	// Use a per-function buffer so Generate() can append each function without
	// wiping the file header / prior functions (Reset on the shared builder
	// used to drop everything except the last function).
	prev := cg.output
	cg.output = new(strings.Builder)
	defer func() { cg.output = prev }()

	cg.emitFunctionPrologue(fn)
	for _, block := range fn.Blocks {
		cg.emitBasicBlock(block)
	}
	cg.emitFunctionEpilogue(fn)

	return cg.output.String(), nil
}

// emitHeader emits the assembly file header.
func (cg *CodeGenerator) emitHeader() {
	cg.output.WriteString("\t.file\t\"source.c\"\n")
	cg.output.WriteString("\t.text\n")
}

// emitGlobal emits a global variable.
func (cg *CodeGenerator) emitGlobal(global *ir.GlobalVar) {
	cg.output.WriteString("\t.globl\t")
	cg.output.WriteString(global.Name)
	cg.output.WriteString("\n")
	cg.output.WriteString("\t.data\n")
	cg.output.WriteString(global.Name)
	cg.output.WriteString(":\n")

	// Emit initializer if present
	if global.Init != nil {
		// TODO: Handle complex initializers
		cg.output.WriteString("\t.quad\t0\n")
	} else {
		cg.output.WriteString("\t.zero\t8\n")
	}
}

// emitConstant emits a constant.
func (cg *CodeGenerator) emitConstant(constant *ir.Constant) {
	cg.output.WriteString("\t.section\t.rodata\n")
	cg.output.WriteString(constant.Name)
	cg.output.WriteString(":\n")

	switch v := constant.Value.(type) {
	case int64:
		cg.output.WriteString("\t.quad\t")
		cg.output.WriteString(fmt.Sprintf("%d", v))
		cg.output.WriteString("\n")
	case string:
		cg.output.WriteString("\t.string\t")
		cg.output.WriteString(fmt.Sprintf("\"%s\"", v))
		cg.output.WriteString("\n")
	default:
		cg.output.WriteString("\t.quad\t0\n")
	}
}

// emitFunctionPrologue emits the function prologue.
// Follows System V AMD64 ABI:
//
//	function_name:
//	    pushq   %rbp
//	    movq    %rsp, %rbp
//	    subq    $N, %rsp    # N = stack frame size (16-byte aligned)
func (cg *CodeGenerator) emitFunctionPrologue(fn *ir.Function) {
	// Emit function label
	cg.output.WriteString("\t.globl\t")
	cg.output.WriteString(fn.Name)
	cg.output.WriteString("\n")
	cg.output.WriteString("\t.type\t")
	cg.output.WriteString(fn.Name)
	cg.output.WriteString(", @function\n")
	cg.output.WriteString(fn.Name)
	cg.output.WriteString(":\n")
	cg.output.WriteString("\t.cfi_startproc\n")

	// Save base pointer
	cg.output.WriteString("\tpushq\t%rbp\n")
	cg.output.WriteString("\t.cfi_def_cfa_offset 16\n")
	cg.output.WriteString("\t.cfi_offset 6, -16\n")

	// Set up frame pointer
	cg.output.WriteString("\tmovq\t%rsp, %rbp\n")
	cg.output.WriteString("\t.cfi_def_cfa_register 6\n")

	// Allocate stack space for local variables and spills
	// Reserve space for saved RBP (8 bytes) + align to 16 bytes
	stackSize := int64(8) // Space for saved RBP

	// Add space for local variables
	for _, local := range fn.LocalVars {
		local.StackOffset = -stackSize
		stackSize += 8 // 8 bytes per local variable
	}

	// Round up to 16-byte alignment (System V ABI requirement)
	if stackSize%16 != 0 {
		stackSize = (stackSize/16 + 1) * 16
	}

	cg.stackFrame = &StackFrame{
		Size:      stackSize,
		Locals:    fn.LocalVars,
		SavedRegs: make([]Reg, 0),
	}

	// Allocate stack space
	if stackSize > 8 {
		cg.output.WriteString("\tsubq\t$")
		cg.output.WriteString(fmt.Sprintf("%d", stackSize-8))
		cg.output.WriteString(", %rsp\n")
	}
}

// emitFunctionEpilogue emits the function epilogue.
//
//	leave
//	ret
func (cg *CodeGenerator) emitFunctionEpilogue(fn *ir.Function) {
	cg.output.WriteString("\tleave\n")
	cg.output.WriteString("\t.cfi_def_cfa 7, 8\n")
	cg.output.WriteString("\tret\n")
	cg.output.WriteString("\t.cfi_endproc\n")
	cg.output.WriteString("\t.size\t")
	cg.output.WriteString(fn.Name)
	cg.output.WriteString(", .-")
	cg.output.WriteString(fn.Name)
	cg.output.WriteString("\n")
}

// emitBasicBlock emits a basic block.
func (cg *CodeGenerator) emitBasicBlock(block *ir.BasicBlock) {
	// Emit block label
	if block.Label != "" {
		cg.output.WriteString(block.Label)
		cg.output.WriteString(":\n")
	}

	// Emit instructions
	for _, instr := range block.Instrs {
		cg.emitInstruction(instr)
	}
}

// emitInstruction emits a single instruction.
func (cg *CodeGenerator) emitInstruction(instr ir.Instruction) {
	switch instr.Opcode() {
	// Arithmetic
	case ir.OpAdd:
		cg.emitArithmetic(instr, "add")
	case ir.OpSub:
		cg.emitArithmetic(instr, "sub")
	case ir.OpMul:
		cg.emitArithmetic(instr, "imul")
	case ir.OpDiv:
		cg.emitDiv(instr)
	case ir.OpMod:
		cg.emitMod(instr)
	case ir.OpNeg:
		cg.emitUnary(instr, "neg")
	case ir.OpBitNot:
		cg.emitUnary(instr, "not")
	case ir.OpBitAnd:
		cg.emitArithmetic(instr, "and")
	case ir.OpBitOr:
		cg.emitArithmetic(instr, "or")
	case ir.OpBitXor:
		cg.emitArithmetic(instr, "xor")
	case ir.OpShl:
		cg.emitShift(instr, "shl")
	case ir.OpShr:
		cg.emitShift(instr, "shr")

	// Comparison
	case ir.OpEq:
		cg.emitCompare(instr, "sete")
	case ir.OpNe:
		cg.emitCompare(instr, "setne")
	case ir.OpLt:
		cg.emitCompare(instr, "setl")
	case ir.OpLe:
		cg.emitCompare(instr, "setle")
	case ir.OpGt:
		cg.emitCompare(instr, "setg")
	case ir.OpGe:
		cg.emitCompare(instr, "setge")

	// Logical
	case ir.OpAnd:
		cg.emitLogical(instr)
	case ir.OpOr:
		cg.emitLogical(instr)
	case ir.OpNot:
		cg.emitUnary(instr, "test")

	// Memory
	case ir.OpLoad:
		cg.emitLoad(instr)
	case ir.OpStore:
		cg.emitStore(instr)
	case ir.OpLea:
		cg.emitLea(instr)
	case ir.OpAlloc:
		cg.emitAlloc(instr)
	case ir.OpFree:
		// No-op for now

	// Control Flow
	case ir.OpJmp:
		cg.emitJmp(instr)
	case ir.OpJmpIf:
		cg.emitJmpIf(instr)
	case ir.OpJmpUnless:
		cg.emitJmpUnless(instr)
	case ir.OpCall:
		cg.emitCall(instr)
	case ir.OpRet:
		cg.emitRet(instr)
	case ir.OpLabel:
		cg.emitLabel(instr)

	// Conversion
	case ir.OpCast:
		cg.emitCast(instr)
	case ir.OpZeroExt:
		cg.emitZeroExt(instr)
	case ir.OpSignExt:
		cg.emitSignExt(instr)
	case ir.OpTrunc:
		cg.emitTrunc(instr)

	// Special
	case ir.OpPhi:
		// SSA phi nodes are eliminated before code generation
	case ir.OpNop:
		// No operation
	}
}

// emitArithmetic emits arithmetic instructions (add, sub, imul, and, or, xor).
func (cg *CodeGenerator) emitArithmetic(instr ir.Instruction, op string) {
	dest := instr.Dest()
	operands := instr.Operands()

	if len(operands) < 2 {
		return
	}

	src := operands[0]
	src2 := operands[1]

	// Allocate registers
	destReg := cg.regAlloc.Allocate(dest)
	srcReg := cg.regAlloc.Allocate(src)
	src2Reg := cg.regAlloc.Allocate(src2)

	// Move first operand to dest
	cg.emitMove(destReg, srcReg, dest.Type)

	// Perform operation with second operand
	cg.output.WriteString("\t")
	cg.output.WriteString(op)
	cg.output.WriteString("\t")
	cg.emitOperand(src2Reg, src2.Type)
	cg.output.WriteString(", ")
	cg.emitOperand(destReg, dest.Type)
	cg.output.WriteString("\n")

	// Free source registers
	cg.regAlloc.FreeOperand(src)
	cg.regAlloc.FreeOperand(src2)
}

// emitDiv emits division instruction.
func (cg *CodeGenerator) emitDiv(instr ir.Instruction) {
	dest := instr.Dest()
	operands := instr.Operands()

	if len(operands) < 2 {
		return
	}

	dividend := operands[0]
	divisor := operands[1]

	// Allocate registers
	dividendReg := cg.regAlloc.Allocate(dividend)
	divisorReg := cg.regAlloc.Allocate(divisor)
	destReg := cg.regAlloc.Allocate(dest)

	// Move dividend to RAX
	cg.output.WriteString("\tmovq\t")
	cg.emitOperand(dividendReg, dividend.Type)
	cg.output.WriteString(", %rax\n")

	// Clear RDX for unsigned division, or sign-extend for signed
	cg.output.WriteString("\tcqo\n")

	// Perform division
	cg.output.WriteString("\tidivq\t")
	cg.emitOperand(divisorReg, divisor.Type)
	cg.output.WriteString("\n")

	// Move quotient from RAX to dest
	cg.output.WriteString("\tmovq\t%rax, ")
	cg.emitOperand(destReg, dest.Type)
	cg.output.WriteString("\n")

	cg.regAlloc.FreeOperand(dividend)
	cg.regAlloc.FreeOperand(divisor)
}

// emitMod emits modulo instruction.
func (cg *CodeGenerator) emitMod(instr ir.Instruction) {
	dest := instr.Dest()
	operands := instr.Operands()

	if len(operands) < 2 {
		return
	}

	dividend := operands[0]
	divisor := operands[1]

	// Allocate registers
	dividendReg := cg.regAlloc.Allocate(dividend)
	divisorReg := cg.regAlloc.Allocate(divisor)
	destReg := cg.regAlloc.Allocate(dest)

	// Move dividend to RAX
	cg.output.WriteString("\tmovq\t")
	cg.emitOperand(dividendReg, dividend.Type)
	cg.output.WriteString(", %rax\n")

	// Clear RDX
	cg.output.WriteString("\tcqo\n")

	// Perform division (remainder goes to RDX)
	cg.output.WriteString("\tidivq\t")
	cg.emitOperand(divisorReg, divisor.Type)
	cg.output.WriteString("\n")

	// Move remainder from RDX to dest
	cg.output.WriteString("\tmovq\t%rdx, ")
	cg.emitOperand(destReg, dest.Type)
	cg.output.WriteString("\n")

	cg.regAlloc.FreeOperand(dividend)
	cg.regAlloc.FreeOperand(divisor)
}

// emitUnary emits unary instructions (neg, not).
func (cg *CodeGenerator) emitUnary(instr ir.Instruction, op string) {
	dest := instr.Dest()
	operands := instr.Operands()

	if len(operands) < 1 {
		return
	}

	src := operands[0]

	// Allocate registers
	destReg := cg.regAlloc.Allocate(dest)
	srcReg := cg.regAlloc.Allocate(src)

	// Move source to dest
	cg.emitMove(destReg, srcReg, dest.Type)

	// Perform operation
	cg.output.WriteString("\t")
	cg.output.WriteString(op)
	cg.output.WriteString("\t")
	cg.emitOperand(destReg, dest.Type)
	cg.output.WriteString("\n")

	cg.regAlloc.FreeOperand(src)
}

// emitCompare emits comparison instructions.
func (cg *CodeGenerator) emitCompare(instr ir.Instruction, setOp string) {
	dest := instr.Dest()
	operands := instr.Operands()

	if len(operands) < 2 {
		return
	}

	left := operands[0]
	right := operands[1]

	// Allocate registers
	destReg := cg.regAlloc.Allocate(dest)
	leftReg := cg.regAlloc.Allocate(left)
	rightReg := cg.regAlloc.Allocate(right)

	// Compare operands
	cg.output.WriteString("\tcmpq\t")
	cg.emitOperand(rightReg, right.Type)
	cg.output.WriteString(", ")
	cg.emitOperand(leftReg, left.Type)
	cg.output.WriteString("\n")

	// Clear dest register
	cg.output.WriteString("\txorl\t")
	cg.emitOperand(destReg, dest.Type)
	cg.output.WriteString(", ")
	cg.emitOperand(destReg, dest.Type)
	cg.output.WriteString("\n")

	// Set byte based on comparison
	cg.output.WriteString("\t")
	cg.output.WriteString(setOp)
	cg.output.WriteString("\t")
	cg.emitOperand(destReg, dest.Type)
	cg.output.WriteString("\n")

	cg.regAlloc.FreeOperand(left)
	cg.regAlloc.FreeOperand(right)
}

// emitLogical emits logical instructions (and, or).
func (cg *CodeGenerator) emitLogical(instr ir.Instruction) {
	dest := instr.Dest()
	operands := instr.Operands()

	if len(operands) < 2 {
		return
	}

	left := operands[0]
	right := operands[1]

	// Allocate registers
	destReg := cg.regAlloc.Allocate(dest)
	leftReg := cg.regAlloc.Allocate(left)
	rightReg := cg.regAlloc.Allocate(right)

	// Move first operand to dest
	cg.emitMove(destReg, leftReg, dest.Type)

	// Perform logical operation
	switch instr.Opcode() {
	case ir.OpAnd:
		cg.output.WriteString("\tandq\t")
	case ir.OpOr:
		cg.output.WriteString("\torq\t")
	}
	cg.emitOperand(rightReg, right.Type)
	cg.output.WriteString(", ")
	cg.emitOperand(destReg, dest.Type)
	cg.output.WriteString("\n")

	cg.regAlloc.FreeOperand(left)
	cg.regAlloc.FreeOperand(right)
}

// emitShift emits shift instructions (shl, shr).
func (cg *CodeGenerator) emitShift(instr ir.Instruction, op string) {
	dest := instr.Dest()
	operands := instr.Operands()

	if len(operands) < 2 {
		return
	}

	value := operands[0]
	shift := operands[1]

	// Allocate registers
	destReg := cg.regAlloc.Allocate(dest)
	valueReg := cg.regAlloc.Allocate(value)
	shiftReg := cg.regAlloc.Allocate(shift)

	// Move value to dest
	cg.emitMove(destReg, valueReg, dest.Type)

	// Move shift amount to RCX (required by x86 shift instructions)
	cg.output.WriteString("\tmovq\t")
	cg.emitOperand(shiftReg, shift.Type)
	cg.output.WriteString(", %rcx\n")

	// Perform shift
	cg.output.WriteString("\t")
	cg.output.WriteString(op)
	cg.output.WriteString("\t%cl, ")
	cg.emitOperand(destReg, dest.Type)
	cg.output.WriteString("\n")

	cg.regAlloc.FreeOperand(value)
	cg.regAlloc.FreeOperand(shift)
}

// emitLoad emits load instruction.
func (cg *CodeGenerator) emitLoad(instr ir.Instruction) {
	dest := instr.Dest()
	operands := instr.Operands()

	if len(operands) < 1 {
		return
	}

	src := operands[0]

	// Allocate register for dest
	destReg := cg.regAlloc.Allocate(dest)

	// Load from memory
	cg.output.WriteString("\tmovq\t")
	cg.emitMemoryOperand(src)
	cg.output.WriteString(", ")
	cg.emitOperand(destReg, dest.Type)
	cg.output.WriteString("\n")
}

// emitStore emits store instruction.
func (cg *CodeGenerator) emitStore(instr ir.Instruction) {
	operands := instr.Operands()

	if len(operands) < 2 {
		return
	}

	value := operands[0]
	addr := operands[1]

	// Allocate register for value
	valueReg := cg.regAlloc.Allocate(value)

	// Store to memory
	cg.output.WriteString("\tmovq\t")
	cg.emitOperand(valueReg, value.Type)
	cg.output.WriteString(", ")
	cg.emitMemoryOperand(addr)
	cg.output.WriteString("\n")

	cg.regAlloc.FreeOperand(value)
}

// emitLea emits load effective address instruction.
func (cg *CodeGenerator) emitLea(instr ir.Instruction) {
	dest := instr.Dest()
	operands := instr.Operands()

	if len(operands) < 1 {
		return
	}

	src := operands[0]

	// Allocate register for dest
	destReg := cg.regAlloc.Allocate(dest)

	// Load effective address
	cg.output.WriteString("\tleaq\t")
	cg.emitMemoryOperand(src)
	cg.output.WriteString(", ")
	cg.emitOperand(destReg, dest.Type)
	cg.output.WriteString("\n")
}

// emitAlloc emits allocation instruction.
func (cg *CodeGenerator) emitAlloc(instr ir.Instruction) {
	// Stack allocation is handled in prologue
	// This is a no-op for now
}

// emitJmp emits unconditional jump.
func (cg *CodeGenerator) emitJmp(instr ir.Instruction) {
	operands := instr.Operands()

	if len(operands) < 1 {
		return
	}

	label := operands[0]

	cg.output.WriteString("\tjmp\t")
	if label.Kind == ir.OperandLabel {
		cg.output.WriteString(fmt.Sprintf("%v", label.Value))
	} else {
		cg.output.WriteString(fmt.Sprintf(".L%v", label.Value))
	}
	cg.output.WriteString("\n")
}

// emitJmpIf emits conditional jump (jump if not zero).
func (cg *CodeGenerator) emitJmpIf(instr ir.Instruction) {
	operands := instr.Operands()

	if len(operands) < 2 {
		return
	}

	cond := operands[0]
	label := operands[1]

	// Allocate register for condition
	condReg := cg.regAlloc.Allocate(cond)

	// Test condition
	cg.output.WriteString("\ttestq\t")
	cg.emitOperand(condReg, cond.Type)
	cg.output.WriteString(", ")
	cg.emitOperand(condReg, cond.Type)
	cg.output.WriteString("\n")

	// Jump if not zero
	cg.output.WriteString("\tjne\t")
	if label.Kind == ir.OperandLabel {
		cg.output.WriteString(fmt.Sprintf("%v", label.Value))
	} else {
		cg.output.WriteString(fmt.Sprintf(".L%v", label.Value))
	}
	cg.output.WriteString("\n")

	cg.regAlloc.FreeOperand(cond)
}

// emitJmpUnless emits conditional jump (jump if zero).
func (cg *CodeGenerator) emitJmpUnless(instr ir.Instruction) {
	operands := instr.Operands()

	if len(operands) < 2 {
		return
	}

	cond := operands[0]
	label := operands[1]

	// Allocate register for condition
	condReg := cg.regAlloc.Allocate(cond)

	// Test condition
	cg.output.WriteString("\ttestq\t")
	cg.emitOperand(condReg, cond.Type)
	cg.output.WriteString(", ")
	cg.emitOperand(condReg, cond.Type)
	cg.output.WriteString("\n")

	// Jump if zero
	cg.output.WriteString("\tje\t")
	if label.Kind == ir.OperandLabel {
		cg.output.WriteString(fmt.Sprintf("%v", label.Value))
	} else {
		cg.output.WriteString(fmt.Sprintf(".L%v", label.Value))
	}
	cg.output.WriteString("\n")

	cg.regAlloc.FreeOperand(cond)
}

// emitCall emits function call.
func (cg *CodeGenerator) emitCall(instr ir.Instruction) {
	dest := instr.Dest()
	operands := instr.Operands()

	if len(operands) < 1 {
		return
	}

	funcOp := operands[0]
	args := operands[1:]

	// Allocate argument registers
	intArgCount := 0

	// Move arguments to argument registers
	for _, arg := range args {
		argReg := cg.regAlloc.Allocate(arg)
		isFP := cg.isFloatingPointType(arg.Type)

		if isFP {
			// TODO: Handle FP arguments properly
			intArgCount++
		} else {
			if intArgCount < len(intArgRegs) {
				targetReg := intArgRegs[intArgCount]
				cg.output.WriteString("\tmovq\t")
				cg.emitOperand(argReg, arg.Type)
				cg.output.WriteString(", ")
				cg.emitOperand(targetReg, arg.Type)
				cg.output.WriteString("\n")
				intArgCount++
			} else {
				// TODO: Handle stack arguments
			}
		}
		cg.regAlloc.FreeOperand(arg)
	}

	// Emit call
	cg.output.WriteString("\tcall\t")
	if funcOp.Kind == ir.OperandGlobal {
		cg.output.WriteString(fmt.Sprintf("%v", funcOp.Value))
	} else {
		cg.output.WriteString("*")
		cg.emitOperand(cg.regAlloc.Allocate(funcOp), funcOp.Type)
	}
	cg.output.WriteString("\n")

	// Move return value to dest if present
	if dest != nil {
		destReg := cg.regAlloc.Allocate(dest)
		cg.output.WriteString("\tmovq\t%rax, ")
		cg.emitOperand(destReg, dest.Type)
		cg.output.WriteString("\n")
	}
}

// emitRet emits return instruction.
func (cg *CodeGenerator) emitRet(instr ir.Instruction) {
	operands := instr.Operands()

	// Move return value to RAX if present
	if len(operands) >= 1 && operands[0] != nil {
		retval := operands[0]
		retvalReg := cg.regAlloc.Allocate(retval)
		cg.output.WriteString("\tmovq\t")
		cg.emitOperand(retvalReg, retval.Type)
		cg.output.WriteString(", %rax\n")
		cg.regAlloc.FreeOperand(retval)
	}

	// Jump to epilogue
	cg.output.WriteString("\tjmp\t.Lepilogue\n")
}

// emitLabel emits a label.
func (cg *CodeGenerator) emitLabel(instr ir.Instruction) {
	operands := instr.Operands()

	if len(operands) < 1 {
		return
	}

	label := operands[0]
	cg.output.WriteString(fmt.Sprintf("%v", label.Value))
	cg.output.WriteString(":\n")
}

// emitCast emits type cast instruction.
func (cg *CodeGenerator) emitCast(instr ir.Instruction) {
	dest := instr.Dest()
	operands := instr.Operands()

	if len(operands) < 1 {
		return
	}

	src := operands[0]

	// Allocate registers
	destReg := cg.regAlloc.Allocate(dest)
	srcReg := cg.regAlloc.Allocate(src)

	// Move with potential size change
	cg.emitMove(destReg, srcReg, dest.Type)
}

// emitZeroExt emits zero extension instruction.
func (cg *CodeGenerator) emitZeroExt(instr ir.Instruction) {
	dest := instr.Dest()
	operands := instr.Operands()

	if len(operands) < 1 {
		return
	}

	src := operands[0]

	// Allocate registers
	destReg := cg.regAlloc.Allocate(dest)
	srcReg := cg.regAlloc.Allocate(src)

	// Zero extend (movzlq for 32-bit to 64-bit)
	cg.output.WriteString("\tmovl\t")
	cg.emitOperand(srcReg, src.Type)
	cg.output.WriteString(", ")
	cg.emitOperand(destReg, dest.Type)
	cg.output.WriteString("\n")
}

// emitSignExt emits sign extension instruction.
func (cg *CodeGenerator) emitSignExt(instr ir.Instruction) {
	dest := instr.Dest()
	operands := instr.Operands()

	if len(operands) < 1 {
		return
	}

	src := operands[0]

	// Allocate registers
	destReg := cg.regAlloc.Allocate(dest)
	cg.regAlloc.Allocate(src)

	// Sign extend based on source type
	cg.output.WriteString("\tcltq\n") // Sign extend EAX to RAX
	cg.output.WriteString("\tmovq\t%rax, ")
	cg.emitOperand(destReg, dest.Type)
	cg.output.WriteString("\n")
}

// emitTrunc emits truncation instruction.
func (cg *CodeGenerator) emitTrunc(instr ir.Instruction) {
	dest := instr.Dest()
	operands := instr.Operands()

	if len(operands) < 1 {
		return
	}

	src := operands[0]

	// Allocate registers
	destReg := cg.regAlloc.Allocate(dest)
	srcReg := cg.regAlloc.Allocate(src)

	// Truncate (just use lower bits)
	cg.output.WriteString("\tmovl\t")
	cg.emitOperand(srcReg, src.Type)
	cg.output.WriteString(", ")
	cg.emitOperand(destReg, dest.Type)
	cg.output.WriteString("\n")
}

// emitMove emits a move instruction.
func (cg *CodeGenerator) emitMove(destReg Reg, srcReg Reg, typ interface{}) {
	cg.output.WriteString("\tmovq\t")
	cg.emitOperand(srcReg, typ)
	cg.output.WriteString(", ")
	cg.emitOperand(destReg, typ)
	cg.output.WriteString("\n")
}

// emitOperand emits an operand.
func (cg *CodeGenerator) emitOperand(reg Reg, typ interface{}) {
	if cg.isFloatingPointType(typ) {
		cg.output.WriteString("%")
		cg.output.WriteString(reg.String())
	} else {
		cg.output.WriteString("%")
		cg.output.WriteString(reg.String())
	}
}

// emitMemoryOperand emits a memory operand.
func (cg *CodeGenerator) emitMemoryOperand(op *ir.Operand) {
	switch op.Kind {
	case ir.OperandTemp:
		if temp, ok := op.Value.(*ir.Temp); ok {
			// Stack-local temporary
			cg.output.WriteString(fmt.Sprintf("%d(%%rbp)", -8-temp.ID*8))
		}
	case ir.OperandParam:
		if temp, ok := op.Value.(*ir.Temp); ok {
			// Parameter (positive offset from RBP)
			cg.output.WriteString(fmt.Sprintf("%d(%%rbp)", 16+temp.ID*8))
		}
	case ir.OperandGlobal:
		cg.output.WriteString(fmt.Sprintf("%v(%%rip)", op.Value))
	case ir.OperandConst:
		cg.output.WriteString(fmt.Sprintf("$%v", op.Value))
	}
}

// isFloatingPointType returns true if the type is floating-point.
func (cg *CodeGenerator) isFloatingPointType(t interface{}) bool {
	if t == nil {
		return false
	}
	// Check if it's a BaseType with float/double kind
	if bt, ok := t.(*parser.BaseType); ok {
		return bt.Kind == parser.TypeFloat || bt.Kind == parser.TypeDouble
	}
	return false
}

// NewLabel generates a unique label.
func (cg *CodeGenerator) NewLabel() string {
	label := fmt.Sprintf(".L%d", cg.labelCounter)
	cg.labelCounter++
	return label
}