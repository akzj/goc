package ir

import (
	"testing"

	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/parser"
)

func setupGenerator(t *testing.T) *IRGenerator {
	g := NewIRGenerator(nil)
	fn := &Function{
		Name:       "test",
		ReturnType: &parser.BaseType{Kind: parser.TypeInt},
		Params:     []*Param{},
		Blocks:     []*BasicBlock{},
		LocalVars:  []*LocalVar{},
	}
	g.currentFunc = fn
	entryBlock := &BasicBlock{
		Label:  "entry",
		Instrs: make([]Instruction, 0),
		Preds:  make([]*BasicBlock, 0),
		Succs:  make([]*BasicBlock, 0),
	}
	fn.Blocks = append(fn.Blocks, entryBlock)
	g.currentBlock = entryBlock
	return g
}

func TestGenerateDoWhileStmt(t *testing.T) {
	g := setupGenerator(t)
	stmt := &parser.DoWhileStmt{
		Body: &parser.ExprStmt{Expr: &parser.IntLiteral{Value: 1}},
		Cond: &parser.IntLiteral{Value: 1},
	}
	if err := g.generateDoWhileStmt(stmt); err != nil {
		t.Fatalf("generateDoWhileStmt() error = %v", err)
	}
	if len(g.currentFunc.Blocks) < 3 {
		t.Errorf("Expected at least 3 blocks, got %d", len(g.currentFunc.Blocks))
	}
}

func TestGenerateForStmt(t *testing.T) {
	g := setupGenerator(t)
	stmt := &parser.ForStmt{
		Init: &parser.ExprStmt{Expr: &parser.IntLiteral{Value: 0}},
		Cond: &parser.IntLiteral{Value: 1},
		Body: &parser.ExprStmt{Expr: &parser.IntLiteral{Value: 1}},
	}
	if err := g.generateForStmt(stmt); err != nil {
		t.Fatalf("generateForStmt() error = %v", err)
	}
	if len(g.currentFunc.Blocks) < 4 {
		t.Errorf("Expected at least 4 blocks, got %d", len(g.currentFunc.Blocks))
	}
}

func TestGenerateBreakStmt(t *testing.T) {
	g := setupGenerator(t)
	stmt := &parser.BreakStmt{}
	if err := g.generateBreakStmt(stmt); err != nil {
		t.Fatalf("generateBreakStmt() error = %v", err)
	}
	if len(g.currentBlock.Instrs) != 1 {
		t.Errorf("Expected 1 instruction, got %d", len(g.currentBlock.Instrs))
	}
}

func TestGenerateContinueStmt(t *testing.T) {
	g := setupGenerator(t)
	stmt := &parser.ContinueStmt{}
	if err := g.generateContinueStmt(stmt); err != nil {
		t.Fatalf("generateContinueStmt() error = %v", err)
	}
	if len(g.currentBlock.Instrs) != 1 {
		t.Errorf("Expected 1 instruction, got %d", len(g.currentBlock.Instrs))
	}
}

func TestGenerateGotoStmt(t *testing.T) {
	g := setupGenerator(t)
	stmt := &parser.GotoStmt{Label: "target"}
	if err := g.generateGotoStmt(stmt); err != nil {
		t.Fatalf("generateGotoStmt() error = %v", err)
	}
	if len(g.currentBlock.Instrs) != 1 {
		t.Errorf("Expected 1 instruction, got %d", len(g.currentBlock.Instrs))
	}
}

func TestGenerateLabelStmt(t *testing.T) {
	g := setupGenerator(t)
	stmt := &parser.LabelStmt{Label: "mylabel"}
	if err := g.generateLabelStmt(stmt); err != nil {
		t.Fatalf("generateLabelStmt() error = %v", err)
	}
	if len(g.currentBlock.Instrs) != 1 {
		t.Errorf("Expected 1 instruction, got %d", len(g.currentBlock.Instrs))
	}
}

func TestGenerateSwitchStmt(t *testing.T) {
	g := setupGenerator(t)
	stmt := &parser.SwitchStmt{
		Cond: &parser.IntLiteral{Value: 1},
		Body: &parser.CompoundStmt{},
	}
	if err := g.generateSwitchStmt(stmt); err != nil {
		t.Fatalf("generateSwitchStmt() error = %v", err)
	}
}

func TestGenerateCaseStmt(t *testing.T) {
	g := setupGenerator(t)
	stmt := &parser.CaseStmt{
		Value: &parser.IntLiteral{Value: 1},
		Stmt:  &parser.ExprStmt{Expr: &parser.IntLiteral{Value: 1}},
	}
	if err := g.generateCaseStmt(stmt); err != nil {
		t.Fatalf("generateCaseStmt() error = %v", err)
	}
}

func TestGenerateCallExpr(t *testing.T) {
	g := setupGenerator(t)
	expr := &parser.CallExpr{
		Func: &parser.IdentExpr{Name: "foo"},
		Args: []parser.Expr{&parser.IntLiteral{Value: 1}},
	}
	result, err := g.generateCallExpr(expr)
	if err != nil {
		t.Fatalf("generateCallExpr() error = %v", err)
	}
	if result == nil {
		t.Error("generateCallExpr() returned nil")
	}
}

func TestGenerateMemberExpr(t *testing.T) {
	g := setupGenerator(t)
	expr := &parser.MemberExpr{
		Object:    &parser.IdentExpr{Name: "obj"},
		Field:     "field",
		IsPointer: false,
	}
	result, err := g.generateMemberExpr(expr)
	if err != nil {
		t.Fatalf("generateMemberExpr() error = %v", err)
	}
	if result == nil {
		t.Error("generateMemberExpr() returned nil")
	}
}

func TestGenerateIndexExpr(t *testing.T) {
	g := setupGenerator(t)
	expr := &parser.IndexExpr{
		Array: &parser.IdentExpr{Name: "arr"},
		Index: &parser.IntLiteral{Value: 0},
	}
	result, err := g.generateIndexExpr(expr)
	if err != nil {
		t.Fatalf("generateIndexExpr() error = %v", err)
	}
	if result == nil {
		t.Error("generateIndexExpr() returned nil")
	}
}

func TestGenerateFloatLiteral(t *testing.T) {
	g := NewIRGenerator(nil)
	expr := &parser.FloatLiteral{Value: 3.14}
	result, err := g.generateFloatLiteral(expr)
	if err != nil {
		t.Fatalf("generateFloatLiteral() error = %v", err)
	}
	if result.Kind != OperandConst {
		t.Errorf("Expected OperandConst, got %v", result.Kind)
	}
}

func TestGenerateCharLiteral(t *testing.T) {
	g := NewIRGenerator(nil)
	expr := &parser.CharLiteral{Value: 'a'}
	result, err := g.generateCharLiteral(expr)
	if err != nil {
		t.Fatalf("generateCharLiteral() error = %v", err)
	}
	if result.Kind != OperandConst {
		t.Errorf("Expected OperandConst, got %v", result.Kind)
	}
}

func TestGenerateStringLiteral(t *testing.T) {
	g := NewIRGenerator(nil)
	expr := &parser.StringLiteral{Value: "hello"}
	result, err := g.generateStringLiteral(expr)
	if err != nil {
		t.Fatalf("generateStringLiteral() error = %v", err)
	}
	if result.Kind != OperandGlobal {
		t.Errorf("Expected OperandGlobal, got %v", result.Kind)
	}
	if len(g.ir.Constants) != 1 {
		t.Errorf("Expected 1 constant, got %d", len(g.ir.Constants))
	}
}

func TestGenerateCastExpr(t *testing.T) {
	g := setupGenerator(t)
	expr := &parser.CastExpr{
		Type: &parser.BaseType{Kind: parser.TypeInt},
		Expr: &parser.IntLiteral{Value: 42},
	}
	result, err := g.generateCastExpr(expr)
	if err != nil {
		t.Fatalf("generateCastExpr() error = %v", err)
	}
	if result == nil {
		t.Error("generateCastExpr() returned nil")
	}
}

func TestGenerateSizeofExpr(t *testing.T) {
	g := NewIRGenerator(nil)
	expr := &parser.SizeofExpr{Type: &parser.BaseType{Kind: parser.TypeInt}}
	result, err := g.generateSizeofExpr(expr)
	if err != nil {
		t.Fatalf("generateSizeofExpr() error = %v", err)
	}
	if result.Kind != OperandConst {
		t.Errorf("Expected OperandConst, got %v", result.Kind)
	}
}

func TestMapBinaryOpAll(t *testing.T) {
	g := NewIRGenerator(nil)
	ops := []lexer.TokenType{
		lexer.ADD, lexer.SUB, lexer.MUL, lexer.QUO, lexer.REM,
		lexer.AND, lexer.OR, lexer.XOR, lexer.SHL, lexer.SHR,
		lexer.EQL, lexer.NEQ, lexer.LSS, lexer.LEQ, lexer.GTR, lexer.GEQ,
		lexer.LAND, lexer.LOR,
	}
	for _, op := range ops {
		opcode := g.mapBinaryOp(op)
		if opcode < 0 {
			t.Errorf("mapBinaryOp(%v) returned invalid opcode %d", op, opcode)
		}
	}
}

func TestMapUnaryOpAll(t *testing.T) {
	g := NewIRGenerator(nil)
	ops := []lexer.TokenType{
		lexer.SUB, lexer.NOT, lexer.BITNOT, lexer.MUL, lexer.AND,
	}
	for _, op := range ops {
		opcode := g.mapUnaryOp(op)
		if opcode < 0 {
			t.Errorf("mapUnaryOp(%v) returned invalid opcode %d", op, opcode)
		}
	}
}

func TestEmitWithNilBlock(t *testing.T) {
	g := NewIRGenerator(nil)
	g.currentFunc = &Function{
		Name:       "test",
		ReturnType: &parser.BaseType{Kind: parser.TypeInt},
		Params:     []*Param{},
		Blocks:     []*BasicBlock{},
		LocalVars:  []*LocalVar{},
	}
	g.currentBlock = nil
	g.Emit(NewNopInstr())
	if g.currentBlock == nil {
		t.Error("Emit() did not create block when nil")
	}
}

func TestToErrhandPos(t *testing.T) {
	pos := lexer.Position{File: "test.go", Line: 10, Column: 5}
	errPos := toErrhandPos(pos)
	if errPos.File != "test.go" || errPos.Line != 10 || errPos.Column != 5 {
		t.Errorf("toErrhandPos() = %+v, want File=test.go, Line=10, Column=5", errPos)
	}
}

func TestGenerateNilAST(t *testing.T) {
	g := NewIRGenerator(nil)
	ir, err := g.Generate(nil)
	if err == nil {
		t.Error("Generate(nil) should return error")
	}
	if ir != nil {
		t.Error("Generate(nil) should return nil IR")
	}
}

func TestGenerateDeclaration(t *testing.T) {
	g := NewIRGenerator(nil)
	
	// Struct decl
	structDecl := &parser.StructDecl{Name: "S"}
	if err := g.generateDeclaration(structDecl); err != nil {
		t.Fatalf("generateDeclaration(struct) error = %v", err)
	}
	
	// Enum decl
	enumDecl := &parser.EnumDecl{Name: "E"}
	if err := g.generateDeclaration(enumDecl); err != nil {
		t.Fatalf("generateDeclaration(enum) error = %v", err)
	}
}

func TestGenerateNilStatement(t *testing.T) {
	g := NewIRGenerator(nil)
	if err := g.generateStatement(nil); err != nil {
		t.Fatalf("generateStatement(nil) error = %v", err)
	}
}

func TestGenerateNilCompoundStmt(t *testing.T) {
	g := NewIRGenerator(nil)
	if err := g.generateCompoundStmt(nil); err != nil {
		t.Fatalf("generateCompoundStmt(nil) error = %v", err)
	}
}

func TestGenerateNilExprStmt(t *testing.T) {
	g := setupGenerator(t)
	stmt := &parser.ExprStmt{Expr: nil}
	if err := g.generateExprStmt(stmt); err != nil {
		t.Fatalf("generateExprStmt(nil expr) error = %v", err)
	}
}

func TestGenerateWithLocalVars(t *testing.T) {
	g := NewIRGenerator(nil)
	fnDecl := &parser.FunctionDecl{
		Name:   "test",
		Type:   &parser.FuncType{Return: &parser.BaseType{Kind: parser.TypeInt}},
		Params: []*parser.ParamDecl{},
		Body: &parser.CompoundStmt{
			Declarations: []parser.Declaration{
				&parser.VarDecl{Name: "x", Type: &parser.BaseType{Kind: parser.TypeInt}, Init: &parser.IntLiteral{Value: 10}},
				&parser.VarDecl{Name: "y", Type: &parser.BaseType{Kind: parser.TypeInt}, Init: &parser.IntLiteral{Value: 20}},
			},
			Statements: []parser.Statement{
				&parser.ReturnStmt{Value: &parser.IntLiteral{Value: 0}},
			},
		},
	}
	ast := &parser.TranslationUnit{Declarations: []parser.Declaration{fnDecl}}
	_, err := g.Generate(ast)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(g.ir.Functions[0].LocalVars) != 2 {
		t.Errorf("Expected 2 local vars, got %d", len(g.ir.Functions[0].LocalVars))
	}
}

func TestGenerateStructDecl(t *testing.T) {
	g := NewIRGenerator(nil)
	structDecl := &parser.StructDecl{Name: "Point"}
	ast := &parser.TranslationUnit{Declarations: []parser.Declaration{structDecl}}
	ir, err := g.Generate(ast)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(ir.Functions) != 0 {
		t.Errorf("Expected 0 functions, got %d", len(ir.Functions))
	}
}

func TestGenerateEnumDecl(t *testing.T) {
	g := NewIRGenerator(nil)
	enumDecl := &parser.EnumDecl{Name: "Color"}
	ast := &parser.TranslationUnit{Declarations: []parser.Declaration{enumDecl}}
	ir, err := g.Generate(ast)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(ir.Functions) != 0 {
		t.Errorf("Expected 0 functions, got %d", len(ir.Functions))
	}
}

func TestGenerateVarDecl(t *testing.T) {
	g := NewIRGenerator(nil)
	varDecl := &parser.VarDecl{
		Name: "global_var",
		Type: &parser.BaseType{Kind: parser.TypeInt},
		Init: &parser.IntLiteral{Value: 42},
	}
	ast := &parser.TranslationUnit{Declarations: []parser.Declaration{varDecl}}
	_, err := g.Generate(ast)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(g.ir.Globals) != 1 {
		t.Errorf("Expected 1 global, got %d", len(g.ir.Globals))
	}
}

func TestGenerateFunctionDecl(t *testing.T) {
	g := NewIRGenerator(nil)
	fnDecl := &parser.FunctionDecl{
		Name:   "main",
		Type:   &parser.FuncType{Return: &parser.BaseType{Kind: parser.TypeInt}},
		Params: []*parser.ParamDecl{},
		Body: &parser.CompoundStmt{
			Declarations: []parser.Declaration{},
			Statements: []parser.Statement{
				&parser.ReturnStmt{Value: &parser.IntLiteral{Value: 0}},
			},
		},
	}
	ast := &parser.TranslationUnit{Declarations: []parser.Declaration{fnDecl}}
	_, err := g.Generate(ast)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(g.ir.Functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(g.ir.Functions))
	}
	if g.ir.Functions[0].Name != "main" {
		t.Errorf("Expected function name 'main', got %s", g.ir.Functions[0].Name)
	}
}