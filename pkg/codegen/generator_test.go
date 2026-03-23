// Package codegen generates x86-64 assembly code from IR.
// This file contains tests for the code generator.
package codegen

import (
	"strings"
	"testing"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/ir"
	"github.com/akzj/goc/pkg/parser"
)

// Helper function to create an int type
func intType() *parser.BaseType {
	return &parser.BaseType{Kind: parser.TypeInt}
}

// Helper function to create a long type
func longType() *parser.BaseType {
	return &parser.BaseType{Kind: parser.TypeLong}
}

// Helper function to create a void type
func voidType() *parser.BaseType {
	return &parser.BaseType{Kind: parser.TypeVoid}
}

// Helper function to create a temp operand
func newTempOperand(id int, typ parser.Type) *ir.Operand {
	return &ir.Operand{
		Kind: ir.OperandTemp,
		Type: typ,
		Value: &ir.Temp{
			ID:   id,
			Type: typ,
		},
	}
}

func floatType() *parser.BaseType {
	return &parser.BaseType{Kind: parser.TypeFloat}
}

func doubleType() *parser.BaseType {
	return &parser.BaseType{Kind: parser.TypeDouble}
}

func newParamOperand(id int, typ parser.Type) *ir.Operand {
	return &ir.Operand{
		Kind: ir.OperandParam,
		Type: typ,
		Value: &ir.Temp{
			ID:   id,
			Type: typ,
		},
	}
}

// Helper function to create a label operand
func newLabelOperand(name string) *ir.Operand {
	return &ir.Operand{
		Kind:  ir.OperandLabel,
		Value: name,
	}
}

// Helper function to create a constant operand
func newConstOperand(value interface{}) *ir.Operand {
	return &ir.Operand{
		Kind:  ir.OperandConst,
		Value: value,
	}
}

// Helper function to create a global operand
func newGlobalOperand(name string) *ir.Operand {
	return &ir.Operand{
		Kind:  ir.OperandGlobal,
		Value: name,
	}
}

// Helper function to create a basic instruction
func newInstruction(opcode ir.Opcode, dest *ir.Operand, operands ...*ir.Operand) ir.Instruction {
	return &testInstruction{
		opcode:   opcode,
		dest:     dest,
		operands: operands,
	}
}

// testInstruction is a test implementation of ir.Instruction
type testInstruction struct {
	opcode   ir.Opcode
	dest     *ir.Operand
	operands []*ir.Operand
}

func (i *testInstruction) Opcode() ir.Opcode     { return i.opcode }
func (i *testInstruction) Dest() *ir.Operand    { return i.dest }
func (i *testInstruction) Operands() []*ir.Operand { return i.operands }
func (i *testInstruction) String() string {
	return "test instruction"
}

// TestNewCodeGenerator tests the creation of a new code generator.
func TestNewCodeGenerator(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	if cg == nil {
		t.Fatal("NewCodeGenerator returned nil")
	}

	if cg.errors != errorHandler {
		t.Error("CodeGenerator errors should be set to errorHandler")
	}

	if cg.output == nil {
		t.Error("CodeGenerator output should be initialized")
	}

	if cg.regAlloc == nil {
		t.Error("CodeGenerator regAlloc should be initialized")
	}
}

// TestGenerateEmptyIR tests generating code for an empty IR.
func TestGenerateEmptyIR(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	ir := &ir.IR{
		Functions: []*ir.Function{},
		Globals:   []*ir.GlobalVar{},
		Constants: []*ir.Constant{},
	}

	asm, err := cg.Generate(ir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if asm == "" {
		t.Error("Expected non-empty assembly output")
	}

	// Check for file header
	if !strings.Contains(asm, ".file") {
		t.Error("Expected file header in assembly")
	}

	// Check for .text section
	if !strings.Contains(asm, ".text") {
		t.Error("Expected .text section in assembly")
	}
}

// TestGenerateFunction tests generating code for a single function.
func TestGenerateFunction(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "test_func",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpRet, nil, newConstOperand(int64(42))),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for function label
	if !strings.Contains(asm, "test_func:") {
		t.Error("Expected function label in assembly")
	}

	// Check for prologue
	if !strings.Contains(asm, "pushq\t%rbp") {
		t.Error("Expected pushq %rbp in prologue")
	}
	if !strings.Contains(asm, "movq\t%rsp, %rbp") {
		t.Error("Expected movq %rsp, %rbp in prologue")
	}

	// Check for epilogue
	if !strings.Contains(asm, "leave") {
		t.Error("Expected leave in epilogue")
	}
	if !strings.Contains(asm, "ret") {
		t.Error("Expected ret in epilogue")
	}
}

// TestGenerateArithmetic tests arithmetic instruction generation.
func TestGenerateArithmetic(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "add_func",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpAdd, newTempOperand(0, intType()),
						newConstOperand(int64(10)), newConstOperand(int64(20))),
					newInstruction(ir.OpRet, nil, newTempOperand(0, intType())),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for add instruction
	if !strings.Contains(asm, "add") {
		t.Error("Expected add instruction in assembly")
	}
}

// TestGenerateSubtract tests subtract instruction generation.
func TestGenerateSubtract(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "sub_func",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpSub, newTempOperand(0, intType()),
						newConstOperand(int64(50)), newConstOperand(int64(30))),
					newInstruction(ir.OpRet, nil, newTempOperand(0, intType())),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for sub instruction
	if !strings.Contains(asm, "sub") {
		t.Error("Expected sub instruction in assembly")
	}
}

// TestGenerateMultiply tests multiply instruction generation.
func TestGenerateMultiply(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "mul_func",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpMul, newTempOperand(0, intType()),
						newConstOperand(int64(5)), newConstOperand(int64(6))),
					newInstruction(ir.OpRet, nil, newTempOperand(0, intType())),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for imul instruction
	if !strings.Contains(asm, "imul") {
		t.Error("Expected imul instruction in assembly")
	}
}

// TestGenerateDivision tests division instruction generation.
func TestGenerateDivision(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "div_func",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpDiv, newTempOperand(0, intType()),
						newConstOperand(int64(100)), newConstOperand(int64(10))),
					newInstruction(ir.OpRet, nil, newTempOperand(0, intType())),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for idiv instruction
	if !strings.Contains(asm, "idiv") {
		t.Error("Expected idiv instruction in assembly")
	}
}

// TestGenerateModulo tests modulo instruction generation.
func TestGenerateModulo(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "mod_func",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpMod, newTempOperand(0, intType()),
						newConstOperand(int64(17)), newConstOperand(int64(5))),
					newInstruction(ir.OpRet, nil, newTempOperand(0, intType())),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for idiv instruction (modulo uses division)
	if !strings.Contains(asm, "idiv") {
		t.Error("Expected idiv instruction for modulo in assembly")
	}
}

// TestGenerateBitwise tests bitwise instruction generation.
func TestGenerateBitwise(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "bitwise_func",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpBitAnd, newTempOperand(0, intType()),
						newConstOperand(int64(0xFF)), newConstOperand(int64(0x0F))),
					newInstruction(ir.OpBitOr, newTempOperand(1, intType()),
						newConstOperand(int64(0xF0)), newConstOperand(int64(0x0F))),
					newInstruction(ir.OpBitXor, newTempOperand(2, intType()),
						newConstOperand(int64(0xFF)), newConstOperand(int64(0x0F))),
					newInstruction(ir.OpRet, nil, newTempOperand(2, intType())),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for bitwise instructions
	if !strings.Contains(asm, "and") {
		t.Error("Expected and instruction in assembly")
	}
	if !strings.Contains(asm, "or") {
		t.Error("Expected or instruction in assembly")
	}
	if !strings.Contains(asm, "xor") {
		t.Error("Expected xor instruction in assembly")
	}
}

// TestGenerateUnary tests unary instruction generation.
func TestGenerateUnary(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "unary_func",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpNeg, newTempOperand(0, intType()),
						newConstOperand(int64(42))),
					newInstruction(ir.OpBitNot, newTempOperand(1, intType()),
						newConstOperand(int64(0xFF))),
					newInstruction(ir.OpRet, nil, newTempOperand(1, intType())),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for unary instructions
	if !strings.Contains(asm, "neg") {
		t.Error("Expected neg instruction in assembly")
	}
	if !strings.Contains(asm, "not") {
		t.Error("Expected not instruction in assembly")
	}
}

// TestGenerateComparison tests comparison instruction generation.
func TestGenerateComparison(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "cmp_func",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpEq, newTempOperand(0, intType()),
						newConstOperand(int64(10)), newConstOperand(int64(10))),
					newInstruction(ir.OpNe, newTempOperand(1, intType()),
						newConstOperand(int64(10)), newConstOperand(int64(20))),
					newInstruction(ir.OpLt, newTempOperand(2, intType()),
						newConstOperand(int64(5)), newConstOperand(int64(10))),
					newInstruction(ir.OpRet, nil, newTempOperand(2, intType())),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for comparison instructions
	if !strings.Contains(asm, "cmp") {
		t.Error("Expected cmp instruction in assembly")
	}
	if !strings.Contains(asm, "sete") {
		t.Error("Expected sete instruction in assembly")
	}
	if !strings.Contains(asm, "setne") {
		t.Error("Expected setne instruction in assembly")
	}
	if !strings.Contains(asm, "setl") {
		t.Error("Expected setl instruction in assembly")
	}
}

// TestGenerateJump tests jump instruction generation.
func TestGenerateJump(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "jmp_func",
		ReturnType: voidType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpJmp, nil, newLabelOperand("end")),
					newInstruction(ir.OpLabel, nil, newLabelOperand("end")),
					newInstruction(ir.OpRet, nil),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for jump instruction
	if !strings.Contains(asm, "jmp") {
		t.Error("Expected jmp instruction in assembly")
	}
}

// TestGenerateConditionalJump tests conditional jump instruction generation.
func TestGenerateConditionalJump(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "cond_jmp_func",
		ReturnType: voidType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpJmpIf, nil,
						newConstOperand(int64(1)), newLabelOperand("then")),
					newInstruction(ir.OpJmpUnless, nil,
						newConstOperand(int64(0)), newLabelOperand("then")),
					newInstruction(ir.OpLabel, nil, newLabelOperand("then")),
					newInstruction(ir.OpRet, nil),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for conditional jump instructions
	if !strings.Contains(asm, "jne") {
		t.Error("Expected jne instruction in assembly")
	}
	if !strings.Contains(asm, "je") {
		t.Error("Expected je instruction in assembly")
	}
}

// TestGenerateCall tests function call instruction generation.
func TestGenerateCall(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "call_func",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpCall, newTempOperand(0, intType()),
						newGlobalOperand("printf"), newConstOperand(int64(42))),
					newInstruction(ir.OpRet, nil, newTempOperand(0, intType())),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for call instruction
	if !strings.Contains(asm, "call") {
		t.Error("Expected call instruction in assembly")
	}
}

// TestGenerateReturn tests return instruction generation.
func TestGenerateReturn(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "ret_func",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpRet, nil, newConstOperand(int64(0))),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for return instruction
	if !strings.Contains(asm, "ret") {
		t.Error("Expected ret instruction in assembly")
	}
}

// TestGenerateShift tests shift instruction generation.
func TestGenerateShift(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "shift_func",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpShl, newTempOperand(0, intType()),
						newConstOperand(int64(1)), newConstOperand(int64(2))),
					newInstruction(ir.OpShr, newTempOperand(1, intType()),
						newConstOperand(int64(8)), newConstOperand(int64(2))),
					newInstruction(ir.OpRet, nil, newTempOperand(1, intType())),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for shift instructions
	if !strings.Contains(asm, "shl") {
		t.Error("Expected shl instruction in assembly")
	}
	if !strings.Contains(asm, "shr") {
		t.Error("Expected shr instruction in assembly")
	}
}

// TestGenerateWithLocals tests function with local variables.
func TestGenerateWithLocals(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "local_func",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpRet, nil, newConstOperand(int64(42))),
				},
			},
		},
		LocalVars: []*ir.LocalVar{
			{Name: "x", Type: intType()},
			{Name: "y", Type: intType()},
		},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for stack allocation
	if !strings.Contains(asm, "subq") {
		t.Error("Expected subq instruction for stack allocation")
	}
}

// TestGenerateGlobal tests global variable generation.
func TestGenerateGlobal(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	ir := &ir.IR{
		Functions: []*ir.Function{},
		Globals: []*ir.GlobalVar{
			{Name: "global_var", Type: intType(), Init: nil},
		},
		Constants: []*ir.Constant{},
	}

	asm, err := cg.Generate(ir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check for global variable
	if !strings.Contains(asm, "global_var") {
		t.Error("Expected global_var in assembly")
	}
	if !strings.Contains(asm, ".data") {
		t.Error("Expected .data section in assembly")
	}
}

// TestGenerateConstant tests constant generation.
func TestGenerateConstant(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	ir := &ir.IR{
		Functions: []*ir.Function{},
		Globals:   []*ir.GlobalVar{},
		Constants: []*ir.Constant{
			{Name: ".LC0", Value: "Hello, World!"},
		},
	}

	asm, err := cg.Generate(ir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check for constant
	if !strings.Contains(asm, ".LC0") {
		t.Error("Expected .LC0 in assembly")
	}
	if !strings.Contains(asm, ".rodata") {
		t.Error("Expected .rodata section in assembly")
	}
}

// TestRegisterAllocatorIntegration tests integration with register allocator.
func TestRegisterAllocatorIntegration(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	// Verify register allocator is initialized
	if cg.regAlloc == nil {
		t.Fatal("Register allocator should be initialized")
	}

	// Test that register allocation works during code generation
	fn := &ir.Function{
		Name:       "regalloc_test",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpAdd, newTempOperand(0, intType()),
						newConstOperand(int64(1)), newConstOperand(int64(2))),
					newInstruction(ir.OpAdd, newTempOperand(1, intType()),
						newTempOperand(0, intType()), newConstOperand(int64(3))),
					newInstruction(ir.OpRet, nil, newTempOperand(1, intType())),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Verify assembly was generated
	if asm == "" {
		t.Error("Expected non-empty assembly output")
	}
}

// TestMultipleFunctions tests generating code for multiple functions.
func TestMultipleFunctions(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	ir := &ir.IR{
		Functions: []*ir.Function{
			{
				Name:       "func1",
				ReturnType: intType(),
				Params:     []*ir.Param{},
				Blocks: []*ir.BasicBlock{
					{
						Label: "entry",
						Instrs: []ir.Instruction{
							newInstruction(ir.OpRet, nil, newConstOperand(int64(1))),
						},
					},
				},
				LocalVars: []*ir.LocalVar{},
			},
			{
				Name:       "func2",
				ReturnType: intType(),
				Params:     []*ir.Param{},
				Blocks: []*ir.BasicBlock{
					{
						Label: "entry",
						Instrs: []ir.Instruction{
							newInstruction(ir.OpRet, nil, newConstOperand(int64(2))),
						},
					},
				},
				LocalVars: []*ir.LocalVar{},
			},
		},
		Globals:   []*ir.GlobalVar{},
		Constants: []*ir.Constant{},
	}

	asm, err := cg.Generate(ir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check for both functions
	if !strings.Contains(asm, "func1:") {
		t.Error("Expected func1 in assembly")
	}
	if !strings.Contains(asm, "func2:") {
		t.Error("Expected func2 in assembly")
	}
}

// TestNewLabel tests label generation.
func TestNewLabel(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	label1 := cg.NewLabel()
	label2 := cg.NewLabel()

	if label1 == label2 {
		t.Error("Expected unique labels")
	}

	if !strings.HasPrefix(label1, ".L") {
		t.Error("Expected label to start with .L")
	}
}

// TestCFIDirectives tests CFI directive generation.
func TestCFIDirectives(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "cfi_func",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpRet, nil, newConstOperand(int64(0))),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for CFI directives
	if !strings.Contains(asm, ".cfi_startproc") {
		t.Error("Expected .cfi_startproc directive")
	}
	if !strings.Contains(asm, ".cfi_endproc") {
		t.Error("Expected .cfi_endproc directive")
	}
	if !strings.Contains(asm, ".cfi_def_cfa_offset") {
		t.Error("Expected .cfi_def_cfa_offset directive")
	}
	if !strings.Contains(asm, ".cfi_offset") {
		t.Error("Expected .cfi_offset directive")
	}
	if !strings.Contains(asm, ".cfi_def_cfa_register") {
		t.Error("Expected .cfi_def_cfa_register directive")
	}
}

// TestStackAlignment tests 16-byte stack alignment.
func TestStackAlignment(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "align_func",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpRet, nil, newConstOperand(int64(0))),
				},
			},
		},
		LocalVars: []*ir.LocalVar{
			{Name: "x", Type: intType()},
		},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Stack should be 16-byte aligned
	// This is a basic check - proper validation would parse the assembly
	if !strings.Contains(asm, "subq") {
		t.Error("Expected stack allocation")
	}
}

// TestSystemVABI tests System V AMD64 ABI compliance.
func TestSystemVABI(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	fn := &ir.Function{
		Name:       "abi_func",
		ReturnType: intType(),
		Params: []*ir.Param{
			{Name: "a", Type: intType()},
			{Name: "b", Type: intType()},
		},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpRet, nil, newConstOperand(int64(0))),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Check for standard prologue
	if !strings.Contains(asm, "pushq\t%rbp") {
		t.Error("Expected standard prologue with pushq %rbp")
	}
	if !strings.Contains(asm, "movq\t%rsp, %rbp") {
		t.Error("Expected standard prologue with movq %rsp, %rbp")
	}
}

// TestCodeGeneratorReset tests that code generator resets state between functions.
func TestCodeGeneratorReset(t *testing.T) {
	errorHandler := errhand.NewErrorHandler()
	cg := NewCodeGenerator(errorHandler)

	// Generate first function
	fn1 := &ir.Function{
		Name:       "first",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpRet, nil, newConstOperand(int64(1))),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	_, err := cg.GenerateFunction(fn1)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Generate second function
	fn2 := &ir.Function{
		Name:       "second",
		ReturnType: intType(),
		Params:     []*ir.Param{},
		Blocks: []*ir.BasicBlock{
			{
				Label: "entry",
				Instrs: []ir.Instruction{
					newInstruction(ir.OpRet, nil, newConstOperand(int64(2))),
				},
			},
		},
		LocalVars: []*ir.LocalVar{},
	}

	asm, err := cg.GenerateFunction(fn2)
	if err != nil {
		t.Fatalf("GenerateFunction failed: %v", err)
	}

	// Second function should not contain first function's name
	if strings.Contains(asm, "first:") {
		t.Error("Second function should not contain first function's label")
	}
}