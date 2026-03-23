// Package semantic tests the semantic analyzer implementation.
package semantic

import (
	"testing"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/parser"
)

// ============================================================================
// Helper Functions
// ============================================================================

func intType() parser.Type {
	return &parser.BaseType{Kind: parser.TypeInt, Signed: true, Long: 0}
}

func floatType() parser.Type {
	return &parser.BaseType{Kind: parser.TypeFloat, Signed: true}
}

func doubleType() parser.Type {
	return &parser.BaseType{Kind: parser.TypeDouble, Signed: true}
}

func charType() parser.Type {
	return &parser.BaseType{Kind: parser.TypeChar, Signed: true, Long: 0}
}

func voidType() parser.Type {
	return &parser.BaseType{Kind: parser.TypeVoid}
}

func longType() parser.Type {
	return &parser.BaseType{Kind: parser.TypeLong, Signed: true, Long: 0}
}

func uintType() parser.Type {
	return &parser.BaseType{Kind: parser.TypeInt, Signed: false}
}

func intPtrType() parser.Type {
	return &parser.PointerType{Elem: intType()}
}

func voidPtrType() parser.Type {
	return &parser.PointerType{Elem: voidType()}
}

func arrayType(elem parser.Type, size int64) parser.Type {
	return &parser.ArrayType{Elem: elem, ArraySize: size}
}

func funcType(ret parser.Type, params []parser.Type, variadic bool) parser.Type {
	return &parser.FuncType{Return: ret, Params: params, Variadic: variadic}
}

func testPos() lexer.Position {
	return lexer.Position{File: "test.c", Line: 1, Column: 1}
}

func newTestTypeChecker() *TypeChecker {
	analyzer := NewSemanticAnalyzer(nil)
	return NewTypeChecker(analyzer)
}

func newTestSemanticAnalyzer() *SemanticAnalyzer {
	return NewSemanticAnalyzer(nil)
}


// makeFuncType creates a function type with given return type, params, and variadic flag.
func makeFuncType(ret parser.Type, params []parser.Type, variadic bool) parser.Type {
	return &parser.FuncType{Return: ret, Params: params, Variadic: variadic}
}

// ============================================================================
// Tests for SymbolTable
// ============================================================================

func TestNewSymbolTable(t *testing.T) {
	st := NewSymbolTable()
	if st == nil {
		t.Fatal("NewSymbolTable() returned nil")
	}
	if st.globalScope == nil {
		t.Error("globalScope should not be nil")
	}
	if st.currentScope == nil {
		t.Error("currentScope should not be nil")
	}
	if len(st.scopes) != 1 {
		t.Errorf("scopes should have 1 element, got %d", len(st.scopes))
	}
}

func TestSymbolTable_PushPopScope(t *testing.T) {
	st := NewSymbolTable()

	// Push a new scope
	scope := st.PushScope("test")
	if scope == nil {
		t.Fatal("PushScope() returned nil")
	}
	if scope.level != 1 {
		t.Errorf("scope level should be 1, got %d", scope.level)
	}
	if len(st.scopes) != 2 {
		t.Errorf("scopes should have 2 elements, got %d", len(st.scopes))
	}

	// Pop the scope
	popped := st.PopScope()
	if popped != scope {
		t.Error("PopScope() should return the popped scope")
	}
	if len(st.scopes) != 1 {
		t.Errorf("scopes should have 1 element after pop, got %d", len(st.scopes))
	}
}

func TestSymbolTable_DeclareLookup(t *testing.T) {
	st := NewSymbolTable()

	// Declare a symbol
	symbol := &Symbol{
		Name: "test_var",
		Kind: SymbolVariable,
		Type: intType(),
	}
	err := st.Declare(symbol)
	if err != nil {
		t.Errorf("Declare() returned error: %v", err)
	}

	// Lookup the symbol
	found := st.Lookup("test_var")
	if found == nil {
		t.Fatal("Lookup() returned nil")
	}
	if found.Name != "test_var" {
		t.Errorf("Lookup() returned wrong name: %s", found.Name)
	}
	if found.Kind != SymbolVariable {
		t.Errorf("Lookup() returned wrong kind: %v", found.Kind)
	}
}

func TestSymbolTable_DeclareDuplicate(t *testing.T) {
	st := NewSymbolTable()

	// Declare a symbol
	symbol := &Symbol{
		Name: "test_var",
		Kind: SymbolVariable,
		Type: intType(),
	}
	err := st.Declare(symbol)
	if err != nil {
		t.Errorf("First Declare() returned error: %v", err)
	}

	// Try to declare again
	err = st.Declare(symbol)
	if err == nil {
		t.Error("Second Declare() should return error for duplicate")
	}
}

func TestSymbolTable_LookupNotFound(t *testing.T) {
	st := NewSymbolTable()

	found := st.Lookup("nonexistent")
	if found != nil {
		t.Errorf("Lookup() should return nil for nonexistent symbol, got %v", found)
	}
}

func TestSymbolTable_ScopeChain(t *testing.T) {
	st := NewSymbolTable()

	// Declare in global scope
	globalSymbol := &Symbol{
		Name: "global_var",
		Kind: SymbolVariable,
		Type: intType(),
	}
	st.Declare(globalSymbol)

	// Push a new scope
	st.PushScope("inner")

	// Lookup should find global symbol
	found := st.Lookup("global_var")
	if found == nil {
		t.Error("Lookup() should find symbol from parent scope")
	}

	// Declare in inner scope
	innerSymbol := &Symbol{
		Name: "inner_var",
		Kind: SymbolVariable,
		Type: intType(),
	}
	st.Declare(innerSymbol)

	// Pop scope
	st.PopScope()

	// Should not find inner symbol anymore
	found = st.Lookup("inner_var")
	if found != nil {
		t.Error("Lookup() should not find symbol from popped scope")
	}

	// Should still find global symbol
	found = st.Lookup("global_var")
	if found == nil {
		t.Error("Lookup() should still find global symbol")
	}
}

func TestSymbolTable_GetCurrentScope(t *testing.T) {
	st := NewSymbolTable()
	scope := st.GetCurrentScope()
	if scope == nil {
		t.Error("GetCurrentScope() returned nil")
	}
	if scope != st.globalScope {
		t.Error("GetCurrentScope() should return global scope initially")
	}
}

func TestSymbolTable_GetGlobalScope(t *testing.T) {
	st := NewSymbolTable()
	scope := st.GetGlobalScope()
	if scope == nil {
		t.Error("GetGlobalScope() returned nil")
	}
	if scope.level != 0 {
		t.Errorf("global scope level should be 0, got %d", scope.level)
	}
}

// ============================================================================
// Tests for Scope
// ============================================================================

func TestScope_GetName(t *testing.T) {
	scope := &Scope{name: "test"}
	if scope.GetName() != "test" {
		t.Errorf("GetName() should return 'test', got %s", scope.GetName())
	}
}

func TestScope_GetLevel(t *testing.T) {
	scope := &Scope{level: 5}
	if scope.GetLevel() != 5 {
		t.Errorf("GetLevel() should return 5, got %d", scope.GetLevel())
	}
}

func TestScope_GetParent(t *testing.T) {
	parent := &Scope{name: "parent"}
	child := &Scope{name: "child", parent: parent}
	if child.GetParent() != parent {
		t.Error("GetParent() should return parent scope")
	}
}

func TestScope_GetChildren(t *testing.T) {
	scope := &Scope{children: []*Scope{{name: "child1"}, {name: "child2"}}}
	children := scope.GetChildren()
	if len(children) != 2 {
		t.Errorf("GetChildren() should return 2 children, got %d", len(children))
	}
}

func TestScope_GetSymbols(t *testing.T) {
	scope := &Scope{symbols: map[string]*Symbol{"x": {Name: "x"}}}
	symbols := scope.GetSymbols()
	if len(symbols) != 1 {
		t.Errorf("GetSymbols() should return 1 symbol, got %d", len(symbols))
	}
}

func TestScope_Lookup(t *testing.T) {
	scope := &Scope{symbols: map[string]*Symbol{"x": {Name: "x"}}}
	found := scope.Lookup("x")
	if found == nil {
		t.Error("Lookup() should find symbol in scope")
	}
	found = scope.Lookup("y")
	if found != nil {
		t.Error("Lookup() should return nil for nonexistent symbol")
	}
}

// ============================================================================
// Tests for Symbol
// ============================================================================

func TestSymbol_String(t *testing.T) {
	symbol := &Symbol{
		Name: "test",
		Kind: SymbolVariable,
		Type: intType(),
	}
	str := symbol.String()
	if str == "" {
		t.Error("String() should return non-empty string")
	}
}

func TestSymbolKind_String(t *testing.T) {
	tests := []struct {
		kind SymbolKind
		want string
	}{
		{SymbolFunction, "function"},
		{SymbolVariable, "variable"},
		{SymbolParameter, "parameter"},
		{SymbolTypedef, "typedef"},
		{SymbolStruct, "struct"},
		{SymbolUnion, "union"},
		{SymbolEnum, "enum"},
		{SymbolEnumConstant, "enum constant"},
		{SymbolLabel, "label"},
		{999, "unknown"},
	}

	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Errorf("SymbolKind(%d).String() = %s, want %s", tt.kind, got, tt.want)
		}
	}
}

// ============================================================================
// Tests for SymbolFlags
// ============================================================================

func TestSymbolFlags_HasFlag(t *testing.T) {
	flags := FlagConst | FlagStatic
	if !flags.HasFlag(FlagConst) {
		t.Error("HasFlag(FlagConst) should be true")
	}
	if !flags.HasFlag(FlagStatic) {
		t.Error("HasFlag(FlagStatic) should be true")
	}
	if flags.HasFlag(FlagExtern) {
		t.Error("HasFlag(FlagExtern) should be false")
	}
}

func TestSymbolFlags_AddFlag(t *testing.T) {
	flags := FlagNone
	flags.AddFlag(FlagConst)
	if !flags.HasFlag(FlagConst) {
		t.Error("AddFlag() should add the flag")
	}
}

func TestSymbolFlags_RemoveFlag(t *testing.T) {
	flags := FlagConst | FlagStatic
	flags.RemoveFlag(FlagConst)
	if flags.HasFlag(FlagConst) {
		t.Error("RemoveFlag() should remove the flag")
	}
	if !flags.HasFlag(FlagStatic) {
		t.Error("RemoveFlag() should not remove other flags")
	}
}

// ============================================================================
// Tests for SemanticAnalyzer
// ============================================================================

func TestNewSemanticAnalyzer(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	if analyzer == nil {
		t.Fatal("NewSemanticAnalyzer() returned nil")
	}
	if analyzer.symbolTable == nil {
		t.Error("symbolTable should not be nil")
	}
	if analyzer.errors == nil {
		t.Error("errors should not be nil")
	}
	if analyzer.typeChecker == nil {
		t.Error("typeChecker should not be nil")
	}
}

func TestNewSemanticAnalyzer_WithHandler(t *testing.T) {
	handler := errhand.NewErrorHandler()
	analyzer := NewSemanticAnalyzer(handler)
	if analyzer == nil {
		t.Fatal("NewSemanticAnalyzer() returned nil")
	}
	if analyzer.errors != handler {
		t.Error("errors should be the provided handler")
	}
}

func TestSemanticAnalyzer_Analyze_NilAST(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	err := analyzer.Analyze(nil)
	if err == nil {
		t.Error("Analyze(nil) should return error")
	}
}

func TestSemanticAnalyzer_Analyze_Empty(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{},
	}
	err := analyzer.Analyze(ast)
	if err != nil {
		t.Errorf("Analyze(empty) returned error: %v", err)
	}
}

func TestSemanticAnalyzer_EnterExitScope(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	scope := analyzer.symbolTable.GetCurrentScope()
	if scope == nil {
		t.Error("EnterScope() should set currentScope")
	}
	if scope.level != 1 {
		t.Errorf("EnterScope() should create scope at level 1, got %d", scope.level)
	}
	analyzer.ExitScope()
	// After exiting, currentScope should be back to global (level 0)
	scope = analyzer.symbolTable.GetCurrentScope()
	if scope == nil {
		t.Error("ExitScope() should not set currentScope to nil")
	}
	if scope.level != 0 {
		t.Errorf("ExitScope() should return to level 0, got %d", scope.level)
	}
}

func TestSemanticAnalyzer_LookupDeclare(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()

	symbol := &Symbol{
		Name: "test",
		Kind: SymbolVariable,
		Type: intType(),
	}
	err := analyzer.Declare(symbol)
	if err != nil {
		t.Errorf("Declare() returned error: %v", err)
	}

	found := analyzer.Lookup("test")
	if found == nil {
		t.Error("Lookup() should find declared symbol")
	}
}

// ============================================================================
// Tests for Function Declaration Analysis
// ============================================================================

func TestSemanticAnalyzer_AnalyzeFunctionDecl(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{
			&parser.FunctionDecl{
				Name: "test_fn",
				Type: makeFuncType(intType(), nil, false),
				Body: &parser.CompoundStmt{
					Statements: []parser.Statement{
						&parser.ReturnStmt{
							Value: &parser.IntLiteral{Value: 0},
						},
					},
				},
			},
		},
	}
	err := analyzer.Analyze(ast)
	if err != nil {
		t.Errorf("Analyze(FunctionDecl) should not return error: %v", err)
	}

	// Check function was declared
	fn := analyzer.Lookup("test_fn")
	if fn == nil {
		t.Error("Function should be declared in symbol table")
	}
	if fn.Kind != SymbolFunction {
		t.Errorf("Function kind should be SymbolFunction, got %v", fn.Kind)
	}
}

func TestSemanticAnalyzer_AnalyzeFunctionWithParams(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{
			&parser.FunctionDecl{
				Name: "test_fn",
				Type: makeFuncType(intType(), []parser.Type{intType()}, false),
				Params: []*parser.ParamDecl{
					{Name: "x", Type: intType()},
				},
				Body: &parser.CompoundStmt{
					Statements: []parser.Statement{},
				},
			},
		},
	}
	err := analyzer.Analyze(ast)
	if err != nil {
		t.Errorf("Analyze(FunctionDecl with params) should not return error: %v", err)
	}
}

func TestSemanticAnalyzer_AnalyzeFunctionWithModifiers(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{
			&parser.FunctionDecl{
				Name:     "static_fn",
				Type:     makeFuncType(intType(), nil, false),
				IsStatic: true,
				Body:     &parser.CompoundStmt{},
			},
			&parser.FunctionDecl{
				Name:      "inline_fn",
				Type:      makeFuncType(intType(), nil, false),
				IsInline:  true,
				Body:      &parser.CompoundStmt{},
			},
			&parser.FunctionDecl{
				Name:     "extern_fn",
				Type:     makeFuncType(intType(), nil, false),
				IsExtern: true,
				Body:     nil,
			},
		},
	}
	err := analyzer.Analyze(ast)
	if err != nil {
		t.Errorf("Analyze(FunctionDecl with modifiers) should not return error: %v", err)
	}
}

// ============================================================================
// Tests for Variable Declaration Analysis
// ============================================================================

func TestSemanticAnalyzer_AnalyzeVarDecl(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{
			&parser.VarDecl{
				Name: "x",
				Type: intType(),
			},
		},
	}
	err := analyzer.Analyze(ast)
	if err != nil {
		t.Errorf("Analyze(VarDecl) should not return error: %v", err)
	}

	// Check variable was declared
	v := analyzer.Lookup("x")
	if v == nil {
		t.Error("Variable should be declared in symbol table")
	}
	if v.Kind != SymbolVariable {
		t.Errorf("Variable kind should be SymbolVariable, got %v", v.Kind)
	}
}

func TestSemanticAnalyzer_AnalyzeVarDeclWithInit(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{
			&parser.VarDecl{
				Name: "x",
				Type: intType(),
				Init: &parser.IntLiteral{Value: 42},
			},
		},
	}
	err := analyzer.Analyze(ast)
	if err != nil {
		t.Errorf("Analyze(VarDecl with init) should not return error: %v", err)
	}
}

func TestSemanticAnalyzer_AnalyzeVarDeclWithModifiers(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{
			&parser.VarDecl{
				Name:     "static_var",
				Type:     intType(),
				IsStatic: true,
			},
			&parser.VarDecl{
				Name:     "extern_var",
				Type:     intType(),
				IsExtern: true,
			},
			&parser.VarDecl{
				Name:    "const_var",
				Type:    intType(),
				IsConst: true,
			},
		},
	}
	err := analyzer.Analyze(ast)
	if err != nil {
		t.Errorf("Analyze(VarDecl with modifiers) should not return error: %v", err)
	}
}

// ============================================================================
// Tests for Struct/Enum Declaration Analysis
// ============================================================================

func TestSemanticAnalyzer_AnalyzeStructDecl(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{
			&parser.StructDecl{
				Name: "Point",
				Fields: []*parser.FieldDecl{
					{Name: "x", Type: intType()},
					{Name: "y", Type: intType()},
				},
			},
		},
	}
	err := analyzer.Analyze(ast)
	if err != nil {
		t.Errorf("Analyze(StructDecl) should not return error: %v", err)
	}
}

func TestSemanticAnalyzer_AnalyzeStructDeclDuplicateFields(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{
			&parser.StructDecl{
				Name: "BadStruct",
				Fields: []*parser.FieldDecl{
					{Name: "x", Type: intType()},
					{Name: "x", Type: intType()}, // Duplicate
				},
			},
		},
	}
	err := analyzer.Analyze(ast)
	// Should report error but continue
	if err != nil {
		t.Errorf("Analyze should continue despite duplicate fields: %v", err)
	}
	// Check error was reported
	errors := analyzer.GetErrors().Errors()
	if len(errors) == 0 {
		t.Error("Should report duplicate field error")
	}
}

func TestSemanticAnalyzer_AnalyzeEnumDecl(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{
			&parser.EnumDecl{
				Name: "Color",
				Values: []*parser.EnumValue{
					{Name: "RED"},
					{Name: "GREEN"},
					{Name: "BLUE"},
				},
			},
		},
	}
	err := analyzer.Analyze(ast)
	if err != nil {
		t.Errorf("Analyze(EnumDecl) should not return error: %v", err)
	}
}

func TestSemanticAnalyzer_AnalyzeEnumDeclWithValue(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{
			&parser.EnumDecl{
				Name: "Color",
				Values: []*parser.EnumValue{
					{Name: "RED", Value: &parser.IntLiteral{Value: 1}},
					{Name: "GREEN", Value: &parser.IntLiteral{Value: 2}},
				},
			},
		},
	}
	err := analyzer.Analyze(ast)
	if err != nil {
		t.Errorf("Analyze(EnumDecl with values) should not return error: %v", err)
	}
}

// ============================================================================
// Tests for Statement Analysis
// ============================================================================

func TestSemanticAnalyzer_StatementCoverage(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	// Test various statement types
	stmts := []parser.Statement{
		&parser.ExprStmt{},
		&parser.ReturnStmt{Value: &parser.IntLiteral{Value: 0}},
		&parser.IfStmt{
			Cond: &parser.IntLiteral{Value: 1},
			Then: &parser.ExprStmt{},
		},
		&parser.WhileStmt{
			Cond: &parser.IntLiteral{Value: 1},
			Body: &parser.ExprStmt{},
		},
		&parser.DoWhileStmt{
			Body: &parser.ExprStmt{},
			Cond: &parser.IntLiteral{Value: 1},
		},
		&parser.ForStmt{
			Cond: &parser.IntLiteral{Value: 1},
			Body: &parser.ExprStmt{},
		},
		&parser.BreakStmt{},
		&parser.ContinueStmt{},
		&parser.GotoStmt{Label: "label1"},
		&parser.LabelStmt{Label: "label1"},
	}

	for _, stmt := range stmts {
		err := analyzer.analyzeStatement(stmt)
		if err != nil {
			t.Errorf("analyzeStatement(%T) should not return error: %v", stmt, err)
		}
	}
}

func TestSemanticAnalyzer_IfStmtWithElse(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	stmt := &parser.IfStmt{
		Cond: &parser.IntLiteral{Value: 1},
		Then: &parser.ExprStmt{},
		Else: &parser.ExprStmt{},
	}
	err := analyzer.analyzeStatement(stmt)
	if err != nil {
		t.Errorf("analyzeStatement(IfStmt with else) should not return error: %v", err)
	}
}

func TestSemanticAnalyzer_ForStmtWithDeclInit(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	stmt := &parser.ForStmt{
		Init: &parser.VarDecl{
			Name: "i",
			Type: intType(),
		},
		Cond: &parser.IntLiteral{Value: 1},
		Body: &parser.ExprStmt{},
	}
	err := analyzer.analyzeStatement(stmt)
	if err != nil {
		t.Errorf("analyzeStatement(ForStmt with decl init) should not return error: %v", err)
	}
}

func TestSemanticAnalyzer_NestedScopes(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)

	// Enter multiple scopes
	analyzer.EnterScope()
	analyzer.EnterScope()
	analyzer.EnterScope()

	scope := analyzer.GetSymbolTable().GetCurrentScope()
	if scope.level != 3 {
		t.Errorf("Expected scope level 3, got %d", scope.level)
	}

	// Exit all scopes
	analyzer.ExitScope()
	analyzer.ExitScope()
	analyzer.ExitScope()

	scope = analyzer.GetSymbolTable().GetCurrentScope()
	if scope.level != 0 {
		t.Errorf("Expected scope level 0 after exiting all, got %d", scope.level)
	}
}

func TestSemanticAnalyzer_DeclareInDifferentScopes(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)

	// Declare in global scope
	sym1 := &Symbol{Name: "global_var", Kind: SymbolVariable, Type: intType()}
	err := analyzer.Declare(sym1)
	if err != nil {
		t.Errorf("Declare in global scope should not return error: %v", err)
	}

	// Enter local scope and declare same name
	analyzer.EnterScope()
	sym2 := &Symbol{Name: "local_var", Kind: SymbolVariable, Type: intType()}
	err = analyzer.Declare(sym2)
	if err != nil {
		t.Errorf("Declare in local scope should not return error: %v", err)
	}

	// Lookup should find local first
	local := analyzer.Lookup("local_var")
	if local == nil {
		t.Error("Lookup should find local_var")
	}

	// Exit scope
	analyzer.ExitScope()

	// Lookup should not find local_var anymore
	local = analyzer.Lookup("local_var")
	if local != nil {
		t.Error("Lookup should not find local_var after exiting scope")
	}

	// But should still find global_var
	global := analyzer.Lookup("global_var")
	if global == nil {
		t.Error("Lookup should still find global_var")
	}
}

func TestSemanticAnalyzer_LookupNil(t *testing.T) {
	analyzer := &SemanticAnalyzer{}
	result := analyzer.Lookup("test")
	if result != nil {
		t.Error("Lookup on nil symbolTable should return nil")
	}
}

func TestSemanticAnalyzer_DeclareNil(t *testing.T) {
	analyzer := &SemanticAnalyzer{}
	err := analyzer.Declare(&Symbol{Name: "test"})
	if err == nil {
		t.Error("Declare on nil symbolTable should return error")
	}
}

func TestSymbolTable_PopScopeEdgeCases(t *testing.T) {
	st := NewSymbolTable()

	// Pop when at global scope
	result := st.PopScope()
	if result == nil {
		t.Error("PopScope at global should not return nil")
	}

	// Pop with nil current scope
	st2 := &SymbolTable{}
	result = st2.PopScope()
	if result != nil {
		t.Error("PopScope with nil currentScope should return nil")
	}
}

func TestSymbolTable_DeclareNoScope(t *testing.T) {
	st := &SymbolTable{}
	err := st.Declare(&Symbol{Name: "test"})
	if err == nil {
		t.Error("Declare with no current scope should return error")
	}
}

// ============================================================================
// Tests for Break/Continue Validation
// ============================================================================

func TestSemanticAnalyzer_BreakOutsideLoop(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	stmt := &parser.BreakStmt{}
	err := analyzer.analyzeStatement(stmt)
	if err != nil {
		t.Errorf("analyzeStatement should not return error: %v", err)
	}

	// Check error was reported
	errors := analyzer.GetErrors().Errors()
	if len(errors) == 0 {
		t.Error("Should report break outside loop error")
	}
}

func TestSemanticAnalyzer_ContinueOutsideLoop(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	stmt := &parser.ContinueStmt{}
	err := analyzer.analyzeStatement(stmt)
	if err != nil {
		t.Errorf("analyzeStatement should not return error: %v", err)
	}

	// Check error was reported
	errors := analyzer.GetErrors().Errors()
	if len(errors) == 0 {
		t.Error("Should report continue outside loop error")
	}
}

func TestSemanticAnalyzer_BreakInsideLoop(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	// Enter loop context
	analyzer.breakStack = append(analyzer.breakStack, true)

	stmt := &parser.BreakStmt{}
	err := analyzer.analyzeStatement(stmt)
	if err != nil {
		t.Errorf("analyzeStatement(break in loop) should not return error: %v", err)
	}

	// Should not report error
	errors := analyzer.GetErrors().Errors()
	if len(errors) > 0 {
		t.Error("Should not report error for break inside loop")
	}
}

// ============================================================================
// Tests for Return Statement Validation
// ============================================================================

func TestSemanticAnalyzer_ReturnOutsideFunction(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	stmt := &parser.ReturnStmt{Value: &parser.IntLiteral{Value: 0}}
	err := analyzer.analyzeStatement(stmt)
	if err != nil {
		t.Errorf("analyzeStatement should not return error: %v", err)
	}

	// Check error was reported
	errors := analyzer.GetErrors().Errors()
	if len(errors) == 0 {
		t.Error("Should report return outside function error")
	}
}

func TestSemanticAnalyzer_ReturnWithValue(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	// Set return type
	analyzer.returnTypes = append(analyzer.returnTypes, intType())

	stmt := &parser.ReturnStmt{Value: &parser.IntLiteral{Value: 42}}
	err := analyzer.analyzeStatement(stmt)
	if err != nil {
		t.Errorf("analyzeStatement(return with value) should not return error: %v", err)
	}
}

func TestSemanticAnalyzer_ReturnVoidInNonVoidFunction(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	// Set non-void return type
	analyzer.returnTypes = append(analyzer.returnTypes, intType())

	stmt := &parser.ReturnStmt{}
	err := analyzer.analyzeStatement(stmt)
	if err != nil {
		t.Errorf("analyzeStatement should not return error: %v", err)
	}

	// Check error was reported
	errors := analyzer.GetErrors().Errors()
	if len(errors) == 0 {
		t.Error("Should report return void in non-void function error")
	}
}

// ============================================================================
// Tests for LValue Checking
// ============================================================================

func TestSemanticAnalyzer_IsLValue_Ident(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	// Declare a variable
	analyzer.Declare(&Symbol{Name: "x", Kind: SymbolVariable, Type: intType()})

	ident := &parser.IdentExpr{Name: "x"}
	if !analyzer.isLValue(ident) {
		t.Error("Variable identifier should be lvalue")
	}
}

func TestSemanticAnalyzer_IsLValue_Deref(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)

	deref := &parser.UnaryExpr{Op: lexer.MUL, Operand: &parser.IdentExpr{Name: "p"}}
	if !analyzer.isLValue(deref) {
		t.Error("*ptr should be lvalue")
	}
}

func TestSemanticAnalyzer_IsLValue_Index(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)

	index := &parser.IndexExpr{
		Array: &parser.IdentExpr{Name: "arr"},
		Index: &parser.IntLiteral{Value: 0},
	}
	if !analyzer.isLValue(index) {
		t.Error("arr[i] should be lvalue")
	}
}

func TestSemanticAnalyzer_IsLValue_Member(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)

	member := &parser.MemberExpr{
		Object:   &parser.IdentExpr{Name: "s"},
		Field:    "field",
		IsPointer: false,
	}
	if !analyzer.isLValue(member) {
		t.Error("s.field should be lvalue")
	}
}

func TestSemanticAnalyzer_IsLValue_Literal(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)

	literal := &parser.IntLiteral{Value: 42}
	if analyzer.isLValue(literal) {
		t.Error("Literal should not be lvalue")
	}
}

func TestSemanticAnalyzer_AssignToNonLValue(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	// Assignment to literal (not lvalue)
	assign := &parser.AssignExpr{
		Op:    lexer.ASSIGN,
		Left:  &parser.IntLiteral{Value: 42},
		Right: &parser.IntLiteral{Value: 0},
	}
	exprStmt := &parser.ExprStmt{Expr: assign}
	err := analyzer.analyzeStatement(exprStmt)
	if err != nil {
		t.Errorf("analyzeStatement should not return error: %v", err)
	}

	// Check error was reported
	errors := analyzer.GetErrors().Errors()
	if len(errors) == 0 {
		t.Error("Should report lvalue error for assignment to literal")
	}
}

// ============================================================================
// Tests for Type Checking Integration
// ============================================================================

func TestSemanticAnalyzer_IfConditionType(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	// Valid: scalar condition
	stmt := &parser.IfStmt{
		Cond: &parser.IntLiteral{Value: 1},
		Then: &parser.ExprStmt{},
	}
	err := analyzer.analyzeStatement(stmt)
	if err != nil {
		t.Errorf("analyzeStatement(if with scalar cond) should not return error: %v", err)
	}
}

func TestSemanticAnalyzer_WhileConditionType(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	stmt := &parser.WhileStmt{
		Cond: &parser.IntLiteral{Value: 1},
		Body: &parser.ExprStmt{},
	}
	err := analyzer.analyzeStatement(stmt)
	if err != nil {
		t.Errorf("analyzeStatement(while) should not return error: %v", err)
	}
}

func TestSemanticAnalyzer_SwitchConditionType(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	stmt := &parser.SwitchStmt{
		Cond: &parser.IntLiteral{Value: 1},
		Body: &parser.CompoundStmt{},
	}
	err := analyzer.analyzeStatement(stmt)
	if err != nil {
		t.Errorf("analyzeStatement(switch) should not return error: %v", err)
	}
}

func TestSemanticAnalyzer_CaseValueType(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	stmt := &parser.CaseStmt{
		Value: &parser.IntLiteral{Value: 1},
		Stmt:  &parser.ExprStmt{},
	}
	err := analyzer.analyzeStatement(stmt)
	if err != nil {
		t.Errorf("analyzeStatement(case) should not return error: %v", err)
	}
}

func TestSemanticAnalyzer_GotoEmptyLabel(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	stmt := &parser.GotoStmt{Label: ""}
	err := analyzer.analyzeStatement(stmt)
	if err != nil {
		t.Errorf("analyzeStatement should not return error: %v", err)
	}

	// Check error was reported
	errors := analyzer.GetErrors().Errors()
	if len(errors) == 0 {
		t.Error("Should report error for goto with empty label")
	}
}

func TestSemanticAnalyzer_LabelEmptyName(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	stmt := &parser.LabelStmt{Label: ""}
	err := analyzer.analyzeStatement(stmt)
	if err != nil {
		t.Errorf("analyzeStatement should not return error: %v", err)
	}

	// Check error was reported
	errors := analyzer.GetErrors().Errors()
	if len(errors) == 0 {
		t.Error("Should report error for label with empty name")
	}
}

// ============================================================================
// Tests for Getters
// ============================================================================

func TestSemanticAnalyzer_GetSymbolTable(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	st := analyzer.GetSymbolTable()
	if st == nil {
		t.Error("GetSymbolTable() should not return nil")
	}
}

func TestSemanticAnalyzer_GetErrors(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	errs := analyzer.GetErrors()
	if errs == nil {
		t.Error("GetErrors() should not return nil")
	}
}

func TestSemanticAnalyzer_GetTypeChecker(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	tc := analyzer.GetTypeChecker()
	if tc == nil {
		t.Error("GetTypeChecker() should not return nil")
	}
}

// ============================================================================
// Tests for Compound Statement with Declarations
// ============================================================================

func TestSemanticAnalyzer_CompoundStmtWithDecls(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	analyzer.EnterScope()
	defer analyzer.ExitScope()

	stmt := &parser.CompoundStmt{
		Declarations: []parser.Declaration{
			&parser.VarDecl{Name: "x", Type: intType()},
		},
		Statements: []parser.Statement{
			&parser.ExprStmt{},
		},
	}
	err := analyzer.analyzeStatement(stmt)
	if err != nil {
		t.Errorf("analyzeStatement(compound with decls) should not return error: %v", err)
	}

	// Check variable was declared
	v := analyzer.Lookup("x")
	if v == nil {
		t.Error("Variable in compound stmt should be declared")
	}
}

// ============================================================================
// Tests for Nil Statement Handling
// ============================================================================

func TestSemanticAnalyzer_NilStatement(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	err := analyzer.analyzeStatement(nil)
	if err != nil {
		t.Errorf("analyzeStatement(nil) should not return error: %v", err)
	}
}

func TestSemanticAnalyzer_NilCompoundStmt(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	err := analyzer.analyzeCompoundStmt(nil)
	if err != nil {
		t.Errorf("analyzeCompoundStmt(nil) should not return error: %v", err)
	}
}

func TestSemanticAnalyzer_NilExprStmt(t *testing.T) {
	analyzer := NewSemanticAnalyzer(nil)
	stmt := &parser.ExprStmt{Expr: nil}
	err := analyzer.analyzeExprStmt(stmt)
	if err != nil {
		t.Errorf("analyzeExprStmt(nil expr) should not return error: %v", err)
	}
}

// ============================================================================
// Tests for toErrhandPos
// ============================================================================

func TestToErrhandPos(t *testing.T) {
	pos := lexer.Position{File: "test.c", Line: 10, Column: 5}
	errPos := toErrhandPos(pos)
	if errPos.File != "test.c" {
		t.Errorf("File should be 'test.c', got %s", errPos.File)
	}
	if errPos.Line != 10 {
		t.Errorf("Line should be 10, got %d", errPos.Line)
	}
	if errPos.Column != 5 {
		t.Errorf("Column should be 5, got %d", errPos.Column)
	}
}