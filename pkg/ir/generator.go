// Package ir provides intermediate representation for the GOC compiler.
// This file defines the IR generator interface and implementation.
package ir

import (
	"fmt"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/parser"
)

// IRGenerator generates IR from annotated AST.
type IRGenerator struct {
	// errors is the error handler.
	errors *errhand.ErrorHandler
	// ir is the generated IR.
	ir *IR
	// tempCounter is the counter for temporary variables.
	tempCounter int
	// labelCounter is the counter for labels.
	labelCounter int
	// currentFunc is the current function being generated.
	currentFunc *Function
	// currentBlock is the current basic block.
	currentBlock *BasicBlock
	// globalCounter is the counter for global variables.
	globalCounter int
	// constCounter is the counter for constants.
	constCounter int
}

// NewIRGenerator creates a new IR generator.
func NewIRGenerator(errorHandler *errhand.ErrorHandler) *IRGenerator {
	if errorHandler == nil {
		errorHandler = errhand.NewErrorHandler()
	}
	return &IRGenerator{
		errors:        errorHandler,
		ir:            &IR{},
		tempCounter:   0,
		labelCounter:  0,
		globalCounter: 0,
		constCounter:  0,
	}
}

// Generate generates IR from the AST.
func (g *IRGenerator) Generate(ast *parser.TranslationUnit) (*IR, error) {
	if ast == nil {
		return nil, fmt.Errorf("nil AST")
	}

	// Reset counters for new compilation unit
	g.tempCounter = 0
	g.labelCounter = 0
	g.globalCounter = 0
	g.constCounter = 0
	g.ir = &IR{
		Functions: make([]*Function, 0),
		Globals:   make([]*GlobalVar, 0),
		Constants: make([]*Constant, 0),
	}

	// Process all declarations
	for _, decl := range ast.Declarations {
		if err := g.generateDeclaration(decl); err != nil {
			g.errors.Error(errhand.ErrUndefinedSymbol, err.Error(), toErrhandPos(decl.Pos()))
		}
	}

	return g.ir, nil
}

// generateDeclaration generates IR for a declaration.
func (g *IRGenerator) generateDeclaration(decl parser.Declaration) error {
	switch d := decl.(type) {
	case *parser.FunctionDecl:
		return g.generateFunctionDecl(d)
	case *parser.VarDecl:
		return g.generateVarDecl(d)
	case *parser.StructDecl:
		// Struct declarations are handled in semantic analysis
		// No IR generation needed
		return nil
	case *parser.EnumDecl:
		// Enum declarations are handled in semantic analysis
		// No IR generation needed
		return nil
	default:
		return fmt.Errorf("unknown declaration type: %T", decl)
	}
}

// generateFunctionDecl generates IR for a function declaration/definition.
func (g *IRGenerator) generateFunctionDecl(decl *parser.FunctionDecl) error {
	// Create function
	fn := &Function{
		Name:       decl.Name,
		ReturnType: decl.Type,
		Params:     make([]*Param, 0, len(decl.Params)),
		Blocks:     make([]*BasicBlock, 0),
		LocalVars:  make([]*LocalVar, 0),
	}

	// Generate parameters
	for _, param := range decl.Params {
		fn.Params = append(fn.Params, &Param{
			Name: param.Name,
			Type: param.Type,
		})
	}

	// Set current function
	g.currentFunc = fn

	// If function has a body, generate IR for it
	if decl.Body != nil {
		// Create entry block
		entryBlock := &BasicBlock{
			Label:  "entry",
			Instrs: make([]Instruction, 0),
			Preds:  make([]*BasicBlock, 0),
			Succs:  make([]*BasicBlock, 0),
		}
		fn.Blocks = append(fn.Blocks, entryBlock)
		g.currentBlock = entryBlock

		// Generate IR for function body
		if err := g.generateCompoundStmt(decl.Body); err != nil {
			return err
		}

		// If no explicit return, add implicit return for void functions
		if len(fn.Blocks) > 0 {
			lastBlock := fn.Blocks[len(fn.Blocks)-1]
			if len(lastBlock.Instrs) == 0 || lastBlock.Instrs[len(lastBlock.Instrs)-1].Opcode() != OpRet {
				// Check if we need to add a return
				if ft, ok := decl.Type.(*parser.FuncType); ok {
					if ft.Return == nil || ft.Return.TypeKind() == parser.TypeVoid {
						g.Emit(NewRetInstr(nil))
					}
				}
			}
		}
	}

	// Add function to IR
	g.ir.Functions = append(g.ir.Functions, fn)
	g.currentFunc = nil
	g.currentBlock = nil

	return nil
}

// generateVarDecl generates IR for a variable declaration.
func (g *IRGenerator) generateVarDecl(decl *parser.VarDecl) error {
	// Check if we're inside a function (local variable)
	if g.currentFunc != nil {
		// Create local variable
		lv := &LocalVar{
			Name:        decl.Name,
			Type:        decl.Type,
			StackOffset: int64(len(g.currentFunc.LocalVars) * 8), // Simplified stack allocation
		}
		g.currentFunc.LocalVars = append(g.currentFunc.LocalVars, lv)
	} else {
		// Create global variable
		gv := &GlobalVar{
			Name: decl.Name,
			Type: decl.Type,
			Init: decl.Init,
		}
		g.ir.Globals = append(g.ir.Globals, gv)
	}

	return nil
}

// generateStatement generates IR for a statement.
func (g *IRGenerator) generateStatement(stmt parser.Statement) error {
	if stmt == nil {
		return nil
	}

	switch s := stmt.(type) {
	case *parser.CompoundStmt:
		return g.generateCompoundStmt(s)
	case *parser.ExprStmt:
		return g.generateExprStmt(s)
	case *parser.ReturnStmt:
		return g.generateReturnStmt(s)
	case *parser.IfStmt:
		return g.generateIfStmt(s)
	case *parser.WhileStmt:
		return g.generateWhileStmt(s)
	case *parser.DoWhileStmt:
		return g.generateDoWhileStmt(s)
	case *parser.ForStmt:
		return g.generateForStmt(s)
	case *parser.BreakStmt:
		return g.generateBreakStmt(s)
	case *parser.ContinueStmt:
		return g.generateContinueStmt(s)
	case *parser.GotoStmt:
		return g.generateGotoStmt(s)
	case *parser.LabelStmt:
		return g.generateLabelStmt(s)
	case *parser.SwitchStmt:
		return g.generateSwitchStmt(s)
	case *parser.CaseStmt:
		return g.generateCaseStmt(s)
	default:
		return fmt.Errorf("unknown statement type: %T", stmt)
	}
}

// generateCompoundStmt generates IR for a compound statement.
func (g *IRGenerator) generateCompoundStmt(stmt *parser.CompoundStmt) error {
	if stmt == nil {
		return nil
	}

	// Generate declarations (for local variables)
	for _, decl := range stmt.Declarations {
		if err := g.generateDeclaration(decl); err != nil {
			return err
		}
	}

	// Generate statements
	for _, s := range stmt.Statements {
		if err := g.generateStatement(s); err != nil {
			return err
		}
	}

	return nil
}

// generateExprStmt generates IR for an expression statement.
func (g *IRGenerator) generateExprStmt(stmt *parser.ExprStmt) error {
	if stmt.Expr == nil {
		return nil
	}

	// Generate expression (result is discarded)
	_, err := g.generateExpr(stmt.Expr)
	return err
}

// generateReturnStmt generates IR for a return statement.
func (g *IRGenerator) generateReturnStmt(stmt *parser.ReturnStmt) error {
	var retValue *Operand
	if stmt.Value != nil {
		var err error
		retValue, err = g.generateExpr(stmt.Value)
		if err != nil {
			return err
		}
	}

	g.Emit(NewRetInstr(retValue))
	return nil
}

// generateIfStmt generates IR for an if statement.
func (g *IRGenerator) generateIfStmt(stmt *parser.IfStmt) error {
	// Generate condition
	cond, err := g.generateExpr(stmt.Cond)
	if err != nil {
		return err
	}

	// Create labels
	thenLabel := g.NewLabel()
	endLabel := g.NewLabel()
	var elseLabel string

	if stmt.Else != nil {
		elseLabel = g.NewLabel()
	}

	// Generate conditional jump
	if stmt.Else != nil {
		g.Emit(NewCondJmpInstr(OpJmpUnless, cond, &Operand{
			Kind:  OperandLabel,
			Value: elseLabel,
		}))
	} else {
		g.Emit(NewCondJmpInstr(OpJmpUnless, cond, &Operand{
			Kind:  OperandLabel,
			Value: endLabel,
		}))
	}

	// Generate then block
	thenBlock := &BasicBlock{
		Label:  thenLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{g.currentBlock},
		Succs:  make([]*BasicBlock, 0),
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, thenBlock)
	g.currentBlock.Succs = append(g.currentBlock.Succs, thenBlock)
	g.currentBlock = thenBlock

	if err := g.generateStatement(stmt.Then); err != nil {
		return err
	}

	// Add jump to end if not already terminated
	if len(g.currentBlock.Instrs) == 0 || !isTerminator(g.currentBlock.Instrs[len(g.currentBlock.Instrs)-1]) {
		g.Emit(NewJmpInstr(&Operand{
			Kind:  OperandLabel,
			Value: endLabel,
		}))
	}

	// Generate else block if present
	if stmt.Else != nil {
		elseBlock := &BasicBlock{
			Label:  elseLabel,
			Instrs: make([]Instruction, 0),
			Preds:  []*BasicBlock{g.currentBlock},
			Succs:  []*BasicBlock{thenBlock},
		}
		g.currentFunc.Blocks = append(g.currentFunc.Blocks, elseBlock)
		g.currentBlock.Succs = append(g.currentBlock.Succs, elseBlock)
		g.currentBlock = elseBlock

		if err := g.generateStatement(stmt.Else); err != nil {
			return err
		}

		// Add jump to end if not already terminated
		if len(g.currentBlock.Instrs) == 0 || !isTerminator(g.currentBlock.Instrs[len(g.currentBlock.Instrs)-1]) {
			g.Emit(NewJmpInstr(&Operand{
				Kind:  OperandLabel,
				Value: endLabel,
			}))
		}
	}

	// Create end block
	endBlock := &BasicBlock{
		Label:  endLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{g.currentBlock},
		Succs:  make([]*BasicBlock, 0),
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, endBlock)
	g.currentBlock.Succs = append(g.currentBlock.Succs, endBlock)
	g.currentBlock = endBlock

	return nil
}

// generateWhileStmt generates IR for a while statement.
func (g *IRGenerator) generateWhileStmt(stmt *parser.WhileStmt) error {
	// Create labels
	condLabel := g.NewLabel()
	bodyLabel := g.NewLabel()
	endLabel := g.NewLabel()

	// Jump to condition
	g.Emit(NewJmpInstr(&Operand{
		Kind:  OperandLabel,
		Value: condLabel,
	}))

	// Create condition block
	condBlock := &BasicBlock{
		Label:  condLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{g.currentBlock},
		Succs:  make([]*BasicBlock, 0),
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, condBlock)
	g.currentBlock.Succs = append(g.currentBlock.Succs, condBlock)
	g.currentBlock = condBlock

	// Generate condition
	cond, err := g.generateExpr(stmt.Cond)
	if err != nil {
		return err
	}

	// Generate conditional jump to body
	g.Emit(NewCondJmpInstr(OpJmpIf, cond, &Operand{
		Kind:  OperandLabel,
		Value: bodyLabel,
	}))

	// Jump to end if condition is false
	g.Emit(NewJmpInstr(&Operand{
		Kind:  OperandLabel,
		Value: endLabel,
	}))

	// Create body block
	bodyBlock := &BasicBlock{
		Label:  bodyLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{condBlock},
		Succs:  []*BasicBlock{condBlock},
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, bodyBlock)
	condBlock.Succs = append(condBlock.Succs, bodyBlock)
	g.currentBlock = bodyBlock

	// Generate body
	if err := g.generateStatement(stmt.Body); err != nil {
		return err
	}

	// Jump back to condition
	if len(g.currentBlock.Instrs) == 0 || !isTerminator(g.currentBlock.Instrs[len(g.currentBlock.Instrs)-1]) {
		g.Emit(NewJmpInstr(&Operand{
			Kind:  OperandLabel,
			Value: condLabel,
		}))
	}

	// Create end block
	endBlock := &BasicBlock{
		Label:  endLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{condBlock},
		Succs:  make([]*BasicBlock, 0),
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, endBlock)
	condBlock.Succs = append(condBlock.Succs, endBlock)
	g.currentBlock = endBlock

	return nil
}

// generateDoWhileStmt generates IR for a do-while statement.
func (g *IRGenerator) generateDoWhileStmt(stmt *parser.DoWhileStmt) error {
	// Create labels
	bodyLabel := g.NewLabel()
	condLabel := g.NewLabel()
	endLabel := g.NewLabel()

	// Create body block
	bodyBlock := &BasicBlock{
		Label:  bodyLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{g.currentBlock},
		Succs:  make([]*BasicBlock, 0),
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, bodyBlock)
	g.currentBlock.Succs = append(g.currentBlock.Succs, bodyBlock)
	g.currentBlock = bodyBlock

	// Generate body
	if err := g.generateStatement(stmt.Body); err != nil {
		return err
	}

	// Jump to condition
	if len(g.currentBlock.Instrs) == 0 || !isTerminator(g.currentBlock.Instrs[len(g.currentBlock.Instrs)-1]) {
		g.Emit(NewJmpInstr(&Operand{
			Kind:  OperandLabel,
			Value: condLabel,
		}))
	}

	// Create condition block
	condBlock := &BasicBlock{
		Label:  condLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{bodyBlock},
		Succs:  make([]*BasicBlock, 0),
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, condBlock)
	bodyBlock.Succs = append(bodyBlock.Succs, condBlock)
	g.currentBlock = condBlock

	// Generate condition
	cond, err := g.generateExpr(stmt.Cond)
	if err != nil {
		return err
	}

	// Generate conditional jump back to body
	g.Emit(NewCondJmpInstr(OpJmpIf, cond, &Operand{
		Kind:  OperandLabel,
		Value: bodyLabel,
	}))

	// Jump to end if condition is false
	g.Emit(NewJmpInstr(&Operand{
		Kind:  OperandLabel,
		Value: endLabel,
	}))

	// Create end block
	endBlock := &BasicBlock{
		Label:  endLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{condBlock},
		Succs:  make([]*BasicBlock, 0),
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, endBlock)
	condBlock.Succs = append(condBlock.Succs, endBlock)
	g.currentBlock = endBlock

	return nil
}

// generateForStmt generates IR for a for statement.
func (g *IRGenerator) generateForStmt(stmt *parser.ForStmt) error {
	// Create labels
	condLabel := g.NewLabel()
	bodyLabel := g.NewLabel()
	updateLabel := g.NewLabel()
	endLabel := g.NewLabel()

	// Generate initialization
	if stmt.Init != nil {
		switch init := stmt.Init.(type) {
		case parser.Declaration:
			if err := g.generateDeclaration(init); err != nil {
				return err
			}
		case parser.Statement:
			if err := g.generateStatement(init); err != nil {
				return err
			}
		case parser.Expr:
			_, err := g.generateExpr(init)
			if err != nil {
				return err
			}
		}
	}

	// Jump to condition
	g.Emit(NewJmpInstr(&Operand{
		Kind:  OperandLabel,
		Value: condLabel,
	}))

	// Create condition block
	condBlock := &BasicBlock{
		Label:  condLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{g.currentBlock},
		Succs:  make([]*BasicBlock, 0),
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, condBlock)
	g.currentBlock.Succs = append(g.currentBlock.Succs, condBlock)
	g.currentBlock = condBlock

	// Generate condition if present
	if stmt.Cond != nil {
		cond, err := g.generateExpr(stmt.Cond)
		if err != nil {
			return err
		}

		// Generate conditional jump to body
		g.Emit(NewCondJmpInstr(OpJmpIf, cond, &Operand{
			Kind:  OperandLabel,
			Value: bodyLabel,
		}))

		// Jump to end if condition is false
		g.Emit(NewJmpInstr(&Operand{
			Kind:  OperandLabel,
			Value: endLabel,
		}))
	} else {
		// No condition, always jump to body
		g.Emit(NewJmpInstr(&Operand{
			Kind:  OperandLabel,
			Value: bodyLabel,
		}))
	}

	// Create body block
	bodyBlock := &BasicBlock{
		Label:  bodyLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{condBlock},
		Succs:  make([]*BasicBlock, 0),
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, bodyBlock)
	condBlock.Succs = append(condBlock.Succs, bodyBlock)
	g.currentBlock = bodyBlock

	// Generate body
	if err := g.generateStatement(stmt.Body); err != nil {
		return err
	}

	// Jump to update
	if len(g.currentBlock.Instrs) == 0 || !isTerminator(g.currentBlock.Instrs[len(g.currentBlock.Instrs)-1]) {
		g.Emit(NewJmpInstr(&Operand{
			Kind:  OperandLabel,
			Value: updateLabel,
		}))
	}

	// Create update block
	updateBlock := &BasicBlock{
		Label:  updateLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{bodyBlock},
		Succs:  []*BasicBlock{condBlock},
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, updateBlock)
	bodyBlock.Succs = append(bodyBlock.Succs, updateBlock)
	g.currentBlock = updateBlock

	// Generate update expression
	if stmt.Update != nil {
		_, err := g.generateExpr(stmt.Update)
		if err != nil {
			return err
		}
	}

	// Jump back to condition
	g.Emit(NewJmpInstr(&Operand{
		Kind:  OperandLabel,
		Value: condLabel,
	}))

	// Create end block
	endBlock := &BasicBlock{
		Label:  endLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{condBlock},
		Succs:  make([]*BasicBlock, 0),
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, endBlock)
	condBlock.Succs = append(condBlock.Succs, endBlock)
	g.currentBlock = endBlock

	return nil
}

// generateBreakStmt generates IR for a break statement.
func (g *IRGenerator) generateBreakStmt(stmt *parser.BreakStmt) error {
	// TODO: Need to track loop end labels in a stack
	// For now, emit a jump to a placeholder label
	g.Emit(NewJmpInstr(&Operand{
		Kind:  OperandLabel,
		Value: "break_target",
	}))
	return nil
}

// generateContinueStmt generates IR for a continue statement.
func (g *IRGenerator) generateContinueStmt(stmt *parser.ContinueStmt) error {
	// TODO: Need to track loop continue labels in a stack
	// For now, emit a jump to a placeholder label
	g.Emit(NewJmpInstr(&Operand{
		Kind:  OperandLabel,
		Value: "continue_target",
	}))
	return nil
}

// generateGotoStmt generates IR for a goto statement.
func (g *IRGenerator) generateGotoStmt(stmt *parser.GotoStmt) error {
	g.Emit(NewJmpInstr(&Operand{
		Kind:  OperandLabel,
		Value: stmt.Label,
	}))
	return nil
}

// generateLabelStmt generates IR for a label statement.
func (g *IRGenerator) generateLabelStmt(stmt *parser.LabelStmt) error {
	label := &Operand{
		Kind:  OperandLabel,
		Value: stmt.Label,
	}
	g.Emit(NewLabelInstr(label))
	return nil
}

// generateSwitchStmt generates IR for a switch statement.
func (g *IRGenerator) generateSwitchStmt(stmt *parser.SwitchStmt) error {
	// Generate condition
	_, err := g.generateExpr(stmt.Cond)
	if err != nil {
		return err
	}

	// Create end label
	endLabel := g.NewLabel()

	// TODO: Generate case dispatch
	// For now, just generate the body
	if err := g.generateStatement(stmt.Body); err != nil {
		return err
	}

	// Create end block
	endBlock := &BasicBlock{
		Label:  endLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{g.currentBlock},
		Succs:  make([]*BasicBlock, 0),
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, endBlock)
	g.currentBlock.Succs = append(g.currentBlock.Succs, endBlock)
	g.currentBlock = endBlock

	return nil
}

// generateCaseStmt generates IR for a case statement.
func (g *IRGenerator) generateCaseStmt(stmt *parser.CaseStmt) error {
	// Generate case label
	if stmt.Value != nil {
		// Named case
		label := g.NewLabel()
		labelOp := &Operand{
			Kind:  OperandLabel,
			Value: label,
		}
		g.Emit(NewLabelInstr(labelOp))

		// TODO: Generate comparison and jump
	} else {
		// Default case
		label := g.NewLabel()
		labelOp := &Operand{
			Kind:  OperandLabel,
			Value: label,
		}
		g.Emit(NewLabelInstr(labelOp))
	}

	// Generate statement
	if stmt.Stmt != nil {
		if err := g.generateStatement(stmt.Stmt); err != nil {
			return err
		}
	}

	return nil
}

// generateExpr generates IR for an expression and returns the result operand.
func (g *IRGenerator) generateExpr(expr parser.Expr) (*Operand, error) {
	if expr == nil {
		return nil, nil
	}

	switch e := expr.(type) {
	case *parser.BinaryExpr:
		return g.generateBinaryExpr(e)
	case *parser.UnaryExpr:
		return g.generateUnaryExpr(e)
	case *parser.CallExpr:
		return g.generateCallExpr(e)
	case *parser.MemberExpr:
		return g.generateMemberExpr(e)
	case *parser.IndexExpr:
		return g.generateIndexExpr(e)
	case *parser.CondExpr:
		return g.generateCondExpr(e)
	case *parser.AssignExpr:
		return g.generateAssignExpr(e)
	case *parser.IdentExpr:
		return g.generateIdentExpr(e)
	case *parser.IntLiteral:
		return g.generateIntLiteral(e)
	case *parser.FloatLiteral:
		return g.generateFloatLiteral(e)
	case *parser.CharLiteral:
		return g.generateCharLiteral(e)
	case *parser.StringLiteral:
		return g.generateStringLiteral(e)
	case *parser.CastExpr:
		return g.generateCastExpr(e)
	case *parser.SizeofExpr:
		return g.generateSizeofExpr(e)
	default:
		return nil, fmt.Errorf("unknown expression type: %T", expr)
	}
}

// generateBinaryExpr generates IR for a binary expression.
func (g *IRGenerator) generateBinaryExpr(expr *parser.BinaryExpr) (*Operand, error) {
	// Generate left operand
	left, err := g.generateExpr(expr.Left)
	if err != nil {
		return nil, err
	}

	// Generate right operand
	right, err := g.generateExpr(expr.Right)
	if err != nil {
		return nil, err
	}

	// Map operator to opcode
	opcode := g.mapBinaryOp(expr.Op)

	// Create result temp
	result := g.NewTemp(nil)

	// Emit instruction
	g.Emit(NewBinaryInstr(opcode, result, left, right))

	return result, nil
}

// generateUnaryExpr generates IR for a unary expression.
func (g *IRGenerator) generateUnaryExpr(expr *parser.UnaryExpr) (*Operand, error) {
	// Generate operand
	operand, err := g.generateExpr(expr.Operand)
	if err != nil {
		return nil, err
	}

	// Map operator to opcode
	opcode := g.mapUnaryOp(expr.Op)

	// Create result temp
	result := g.NewTemp(nil)

	// Emit instruction
	g.Emit(NewUnaryInstr(opcode, result, operand))

	return result, nil
}

// generateCallExpr generates IR for a call expression.
func (g *IRGenerator) generateCallExpr(expr *parser.CallExpr) (*Operand, error) {
	// Generate function operand
	funcOp, err := g.generateExpr(expr.Func)
	if err != nil {
		return nil, err
	}

	// Generate arguments
	args := make([]*Operand, 0, len(expr.Args))
	for _, arg := range expr.Args {
		argOp, err := g.generateExpr(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, argOp)
	}

	// Create result temp (if function returns a value)
	var result *Operand
	// TODO: Determine if function returns a value
	result = g.NewTemp(nil)

	// Emit call instruction
	g.Emit(NewCallInstr(result, funcOp, args))

	return result, nil
}

// generateMemberExpr generates IR for a member access expression.
func (g *IRGenerator) generateMemberExpr(expr *parser.MemberExpr) (*Operand, error) {
	// Generate object operand
	obj, err := g.generateExpr(expr.Object)
	if err != nil {
		return nil, err
	}

	// Calculate field offset (simplified - would need type info)
	// TODO: Implement proper field offset calculation

	// Create result temp
	result := g.NewTemp(nil)

	// Emit load instruction
	if expr.IsPointer {
		// Pointer access: load from address
		g.Emit(NewLoadInstr(result, obj))
	} else {
		// Direct access: load from struct
		// TODO: Implement proper struct field access
		g.Emit(NewLoadInstr(result, obj))
	}

	return result, nil
}

// generateIndexExpr generates IR for an array indexing expression.
func (g *IRGenerator) generateIndexExpr(expr *parser.IndexExpr) (*Operand, error) {
	// Generate array operand
	array, err := g.generateExpr(expr.Array)
	if err != nil {
		return nil, err
	}

	// Generate index operand
	_, err = g.generateExpr(expr.Index)
	if err != nil {
		return nil, err
	}

	// Calculate address: array + index * sizeof(element)
	// TODO: Implement proper address calculation

	// Create result temp
	result := g.NewTemp(nil)

	// Emit load instruction
	g.Emit(NewLoadInstr(result, array))

	return result, nil
}

// generateCondExpr generates IR for a conditional expression.
func (g *IRGenerator) generateCondExpr(expr *parser.CondExpr) (*Operand, error) {
	// Generate condition
	cond, err := g.generateExpr(expr.Cond)
	if err != nil {
		return nil, err
	}

	// Create labels
	thenLabel := g.NewLabel()
	elseLabel := g.NewLabel()
	endLabel := g.NewLabel()

	// Generate conditional jump
	g.Emit(NewCondJmpInstr(OpJmpUnless, cond, &Operand{
		Kind:  OperandLabel,
		Value: elseLabel,
	}))

	// Create then block
	thenBlock := &BasicBlock{
		Label:  thenLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{g.currentBlock},
		Succs:  make([]*BasicBlock, 0),
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, thenBlock)
	g.currentBlock.Succs = append(g.currentBlock.Succs, thenBlock)
	g.currentBlock = thenBlock

	// Generate true expression
	trueVal, err := g.generateExpr(expr.True)
	if err != nil {
		return nil, err
	}

	// Create result temp
	result := g.NewTemp(nil)

	// Move true value to result
	g.Emit(NewBinaryInstr(OpAdd, result, trueVal, &Operand{
		Kind:  OperandConst,
		Value: int64(0),
	}))

	// Jump to end
	g.Emit(NewJmpInstr(&Operand{
		Kind:  OperandLabel,
		Value: endLabel,
	}))

	// Create else block
	elseBlock := &BasicBlock{
		Label:  elseLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{thenBlock},
		Succs:  make([]*BasicBlock, 0),
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, elseBlock)
	thenBlock.Succs = append(thenBlock.Succs, elseBlock)
	g.currentBlock = elseBlock

	// Generate false expression
	falseVal, err := g.generateExpr(expr.False)
	if err != nil {
		return nil, err
	}

	// Move false value to result
	g.Emit(NewBinaryInstr(OpAdd, result, falseVal, &Operand{
		Kind:  OperandConst,
		Value: int64(0),
	}))

	// Jump to end
	g.Emit(NewJmpInstr(&Operand{
		Kind:  OperandLabel,
		Value: endLabel,
	}))

	// Create end block
	endBlock := &BasicBlock{
		Label:  endLabel,
		Instrs: make([]Instruction, 0),
		Preds:  []*BasicBlock{elseBlock},
		Succs:  make([]*BasicBlock, 0),
	}
	g.currentFunc.Blocks = append(g.currentFunc.Blocks, endBlock)
	elseBlock.Succs = append(elseBlock.Succs, endBlock)
	g.currentBlock = endBlock

	return result, nil
}

// generateAssignExpr generates IR for an assignment expression.
func (g *IRGenerator) generateAssignExpr(expr *parser.AssignExpr) (*Operand, error) {
	// Generate right-hand side
	value, err := g.generateExpr(expr.Right)
	if err != nil {
		return nil, err
	}

	// Generate left-hand side (address)
	addr, err := g.generateExpr(expr.Left)
	if err != nil {
		return nil, err
	}

	// Emit store instruction
	g.Emit(NewStoreInstr(value, addr))

	return value, nil
}

// generateIdentExpr generates IR for an identifier expression.
func (g *IRGenerator) generateIdentExpr(expr *parser.IdentExpr) (*Operand, error) {
	// Create operand for identifier
	// TODO: Look up symbol to determine if it's local, global, or parameter
	op := &Operand{
		Kind:  OperandGlobal,
		Value: expr.Name,
	}

	return op, nil
}

// generateIntLiteral generates IR for an integer literal.
func (g *IRGenerator) generateIntLiteral(expr *parser.IntLiteral) (*Operand, error) {
	return &Operand{
		Kind:  OperandConst,
		Value: expr.Value,
	}, nil
}

// generateFloatLiteral generates IR for a float literal.
func (g *IRGenerator) generateFloatLiteral(expr *parser.FloatLiteral) (*Operand, error) {
	return &Operand{
		Kind:  OperandConst,
		Value: expr.Value,
	}, nil
}

// generateCharLiteral generates IR for a character literal.
func (g *IRGenerator) generateCharLiteral(expr *parser.CharLiteral) (*Operand, error) {
	return &Operand{
		Kind:  OperandConst,
		Value: int64(expr.Value),
	}, nil
}

// generateStringLiteral generates IR for a string literal.
func (g *IRGenerator) generateStringLiteral(expr *parser.StringLiteral) (*Operand, error) {
	// Create constant
	constName := fmt.Sprintf(".str.%d", g.constCounter)
	g.constCounter++

	constant := &Constant{
		Name:  constName,
		Value: expr.Value,
	}
	g.ir.Constants = append(g.ir.Constants, constant)

	return &Operand{
		Kind:  OperandGlobal,
		Value: constName,
	}, nil
}

// generateCastExpr generates IR for a cast expression.
func (g *IRGenerator) generateCastExpr(expr *parser.CastExpr) (*Operand, error) {
	// Generate operand
	operand, err := g.generateExpr(expr.Expr)
	if err != nil {
		return nil, err
	}

	// Create result temp
	result := g.NewTemp(expr.Type)

	// Determine cast type
	// TODO: Implement proper cast type determination
	g.Emit(NewCastInstr(OpCast, result, operand))

	return result, nil
}

// generateSizeofExpr generates IR for a sizeof expression.
func (g *IRGenerator) generateSizeofExpr(expr *parser.SizeofExpr) (*Operand, error) {
	// sizeof is evaluated at compile time
	// TODO: Calculate actual size based on type
	size := int64(8) // Default size

	return &Operand{
		Kind:  OperandConst,
		Value: size,
	}, nil
}

// mapBinaryOp maps a parser operator to an IR opcode.
func (g *IRGenerator) mapBinaryOp(op lexer.TokenType) Opcode {
	switch op {
	case lexer.ADD:
		return OpAdd
	case lexer.SUB:
		return OpSub
	case lexer.MUL:
		return OpMul
	case lexer.QUO:
		return OpDiv
	case lexer.REM:
		return OpMod
	case lexer.AND:
		return OpBitAnd
	case lexer.OR:
		return OpBitOr
	case lexer.XOR:
		return OpBitXor
	case lexer.SHL:
		return OpShl
	case lexer.SHR:
		return OpShr
	case lexer.EQL:
		return OpEq
	case lexer.NEQ:
		return OpNe
	case lexer.LSS:
		return OpLt
	case lexer.LEQ:
		return OpLe
	case lexer.GTR:
		return OpGt
	case lexer.GEQ:
		return OpGe
	case lexer.LAND:
		return OpAnd
	case lexer.LOR:
		return OpOr
	default:
		return OpAdd // Default
	}
}

// mapUnaryOp maps a parser unary operator to an IR opcode.
func (g *IRGenerator) mapUnaryOp(op lexer.TokenType) Opcode {
	switch op {
	case lexer.SUB:
		return OpNeg
	case lexer.NOT:
		return OpNot
	case lexer.BITNOT:
		return OpBitNot
	case lexer.MUL:
		return OpLoad // Dereference
	case lexer.AND:
		return OpLea // Address-of
	default:
		return OpNeg // Default
	}
}

// NewTemp creates a new temporary variable.
func (g *IRGenerator) NewTemp(t parser.Type) *Operand {
	temp := &Temp{
		ID:   g.tempCounter,
		Type: t,
	}
	g.tempCounter++

	return &Operand{
		Kind:  OperandTemp,
		Type:  t,
		Value: temp,
	}
}

// NewLabel creates a new label.
func (g *IRGenerator) NewLabel() string {
	label := fmt.Sprintf("L%d", g.labelCounter)
	g.labelCounter++
	return label
}

// Emit emits an instruction to the current block.
func (g *IRGenerator) Emit(instr Instruction) {
	if g.currentBlock == nil {
		// Create a default block if none exists
		g.currentBlock = &BasicBlock{
			Label:  "entry",
			Instrs: make([]Instruction, 0),
			Preds:  make([]*BasicBlock, 0),
			Succs:  make([]*BasicBlock, 0),
		}
		if g.currentFunc != nil {
			g.currentFunc.Blocks = append(g.currentFunc.Blocks, g.currentBlock)
		}
	}
	g.currentBlock.Instrs = append(g.currentBlock.Instrs, instr)
}

// isTerminator checks if an instruction is a control flow terminator.
func isTerminator(instr Instruction) bool {
	switch instr.Opcode() {
	case OpJmp, OpJmpIf, OpJmpUnless, OpRet:
		return true
	default:
		return false
	}
}

// toErrhandPos converts a lexer.Position to errhand.Position.
func toErrhandPos(pos lexer.Position) errhand.Position {
	return errhand.Position{
		File:   pos.File,
		Line:   pos.Line,
		Column: pos.Column,
	}
}