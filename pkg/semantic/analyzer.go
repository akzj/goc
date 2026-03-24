// Package semantic performs semantic analysis on the AST.
// This file defines the semantic analyzer with complete AST traversal,
// type checking, scope management, and declaration validation.
package semantic

import (
	"fmt"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/parser"
)

// SemanticAnalyzer performs semantic analysis on AST.
// It traverses the AST using a visitor pattern, performs type checking,
// manages scopes, and validates declarations.
type SemanticAnalyzer struct {
	// symbolTable is the symbol table for scope management.
	symbolTable *SymbolTable
	// errors is the error handler for reporting semantic errors.
	errors *errhand.ErrorHandler
	// typeChecker is the type checker for type validation.
	typeChecker *TypeChecker
	// currentFunc tracks the current function being analyzed.
	currentFunc *parser.FunctionDecl
	// breakStack tracks loop nesting for break/continue validation.
	breakStack []bool
	// returnTypes tracks expected return types for functions.
	returnTypes []parser.Type
}

// NewSemanticAnalyzer creates a new semantic analyzer.
// The analyzer initializes a symbol table, error handler, and type checker.
func NewSemanticAnalyzer(errorHandler *errhand.ErrorHandler) *SemanticAnalyzer {
	if errorHandler == nil {
		errorHandler = errhand.NewErrorHandler()
	}
	analyzer := &SemanticAnalyzer{
		symbolTable: NewSymbolTable(),
		errors:      errorHandler,
		breakStack:  make([]bool, 0),
		returnTypes: make([]parser.Type, 0),
	}
	analyzer.typeChecker = NewTypeChecker(analyzer)
	return analyzer
}

// Analyze performs semantic analysis on the AST.
// It traverses all declarations, performs type checking, and validates semantics.
// Returns nil if analysis succeeds, or an error if fatal errors occur.
// Errors are collected in the error handler and can be retrieved via GetErrors().
func (a *SemanticAnalyzer) Analyze(ast *parser.TranslationUnit) error {
	if ast == nil {
		return fmt.Errorf("nil AST")
	}

	// Analyze all declarations in global scope
	for _, decl := range ast.Declarations {
		if err := a.analyzeDeclaration(decl); err != nil {
			// Continue analyzing other declarations despite errors
			a.errors.Error(errhand.ErrUndefinedSymbol, err.Error(), toErrhandPos(decl.Pos()))
		}
	}

	return nil
}

// analyzeDeclaration analyzes a single declaration.
// It dispatches to specific handlers based on declaration type.
func (a *SemanticAnalyzer) analyzeDeclaration(decl parser.Declaration) error {
	switch d := decl.(type) {
	case *parser.FunctionDecl:
		return a.analyzeFunctionDecl(d)
	case *parser.VarDecl:
		return a.analyzeVarDecl(d)
	case *parser.StructDecl:
		return a.analyzeStructDecl(d)
	case *parser.EnumDecl:
		return a.analyzeEnumDecl(d)
	default:
		// Unknown declaration type, skip
		return nil
	}
}

// analyzeFunctionDecl analyzes a function declaration or definition.
// It validates the function signature, declares the function symbol,
// and analyzes the function body if present.
func (a *SemanticAnalyzer) analyzeFunctionDecl(decl *parser.FunctionDecl) error {
	// Create symbol for the function
	flags := FlagNone
	if decl.IsStatic {
		flags |= FlagStatic
	}
	if decl.IsInline {
		flags |= FlagInline
	}
	if decl.IsExtern {
		flags |= FlagExtern
	}

	// decl.Type is already a *parser.FuncType from the parser with correct Return and Params
	funcType, ok := decl.Type.(*parser.FuncType)
	if !ok {
		a.errors.Error(errhand.ErrInvalidType,
			fmt.Sprintf("function %s has invalid type", decl.Name),
			toErrhandPos(decl.Pos()))
		return fmt.Errorf("function %s has invalid type", decl.Name)
	}

	symbol := &Symbol{
		Name:     decl.Name,
		Kind:     SymbolFunction,
		Type:     funcType,
		Position: decl.Pos(),
		Flags:    flags,
	}

	// Check for duplicate function declaration in the same scope
	existing := a.symbolTable.Lookup(decl.Name)
	if existing != nil && existing.Kind == SymbolFunction {
		// Check if it's a redeclaration (allowed for functions)
		// But if both have bodies, it's an error
		if existing.Position.File == decl.Pos().File &&
			existing.Position.Line == decl.Pos().Line {
			// Same location, skip
			return nil
		}
	}

	// Declare the function
	if err := a.symbolTable.Declare(symbol); err != nil {
		a.errors.Error(errhand.ErrDuplicateSymbol,
			fmt.Sprintf("duplicate function declaration '%s'", decl.Name),
			toErrhandPos(decl.Pos()))
		// Continue analysis despite duplicate
	}

	// Analyze function body if present
	if decl.Body != nil {
		a.currentFunc = decl

		// Push return type for validation
		var returnType parser.Type
		if ft, ok := decl.Type.(*parser.FuncType); ok {
			returnType = ft.Return
		}
		a.returnTypes = append(a.returnTypes, returnType)

		// Enter function scope
		a.EnterScope()

		// Declare parameters
		for _, param := range decl.Params {
			paramSymbol := &Symbol{
				Name:     param.Name,
				Kind:     SymbolParameter,
				Type:     param.Type,
				Position: param.Pos(),
				Flags:    FlagNone,
			}
			if err := a.symbolTable.Declare(paramSymbol); err != nil {
				a.errors.Error(errhand.ErrDuplicateSymbol,
					fmt.Sprintf("duplicate parameter '%s'", param.Name),
					toErrhandPos(param.Pos()))
			}
		}

		// Analyze function body statements
		if err := a.analyzeCompoundStmt(decl.Body); err != nil {
			a.ExitScope()
			a.returnTypes = a.returnTypes[:len(a.returnTypes)-1]
			a.currentFunc = nil
			return err
		}

		// Exit function scope
		a.ExitScope()
		a.returnTypes = a.returnTypes[:len(a.returnTypes)-1]
		a.currentFunc = nil
	}

	return nil
}

// analyzeVarDecl analyzes a variable declaration.
// It validates the variable type, checks for duplicates,
// and analyzes the initializer if present.
func (a *SemanticAnalyzer) analyzeVarDecl(decl *parser.VarDecl) error {
	flags := FlagNone
	if decl.IsStatic {
		flags |= FlagStatic
	}
	if decl.IsExtern {
		flags |= FlagExtern
	}
	if decl.IsConst {
		flags |= FlagConst
	}

	symbol := &Symbol{
		Name:     decl.Name,
		Kind:     SymbolVariable,
		Type:     decl.Type,
		Position: decl.Pos(),
		Flags:    flags,
	}

	// Declare the variable
	if err := a.symbolTable.Declare(symbol); err != nil {
		a.errors.Error(errhand.ErrDuplicateSymbol,
			fmt.Sprintf("duplicate declaration '%s'", decl.Name),
			toErrhandPos(decl.Pos()))
		// Continue analysis despite duplicate
	}

	// Analyze initializer if present
	if decl.Init != nil {
		// Type check the initializer
		initType := a.typeChecker.InferType(decl.Init)
		if initType != nil && decl.Type != nil {
			a.typeChecker.CheckAssignable(decl.Type, initType, decl.Pos())
		}
	}

	return nil
}

// analyzeStructDecl analyzes a struct/union declaration.
// It validates field types and checks for duplicate field names.
func (a *SemanticAnalyzer) analyzeStructDecl(decl *parser.StructDecl) error {
	if decl.Name == "" && len(decl.Fields) == 0 {
		// Anonymous struct without fields, skip
		return nil
	}

	// Check for duplicate field names
	fieldNames := make(map[string]bool)
	for _, field := range decl.Fields {
		if field.Name == "" {
			// Anonymous field, skip
			continue
		}
		if fieldNames[field.Name] {
			a.errors.Error(errhand.ErrDuplicateSymbol,
				fmt.Sprintf("duplicate field '%s' in struct/union", field.Name),
				toErrhandPos(field.Pos()))
		}
		fieldNames[field.Name] = true

		// Analyze field type
		if field.BitWidth != nil {
			// Bitfield: width must be integer constant
			bitType := a.typeChecker.InferType(field.BitWidth)
			if bitType != nil {
				if !a.typeChecker.isIntegerType(a.typeChecker.unwrapType(bitType)) {
					a.errors.Error(errhand.ErrInvalidType,
						"bitfield width must be integer",
						toErrhandPos(field.BitWidth.Pos()))
				}
			}
		}
	}

	return nil
}

// analyzeEnumDecl analyzes an enum declaration.
// It validates enum constant names and values.
func (a *SemanticAnalyzer) analyzeEnumDecl(decl *parser.EnumDecl) error {
	if decl.Name == "" && len(decl.Values) == 0 {
		// Anonymous enum without values, skip
		return nil
	}

	// Check for duplicate enum constant names
	constantNames := make(map[string]bool)
	for _, val := range decl.Values {
		if constantNames[val.Name] {
			a.errors.Error(errhand.ErrDuplicateSymbol,
				fmt.Sprintf("duplicate enum constant '%s'", val.Name),
				toErrhandPos(val.Pos()))
		}
		constantNames[val.Name] = true

		// Analyze enum value if present
		if val.Value != nil {
			valType := a.typeChecker.InferType(val.Value)
			if valType != nil {
				if !a.typeChecker.isIntegerType(a.typeChecker.unwrapType(valType)) {
					a.errors.Error(errhand.ErrInvalidType,
						"enum value must be integer",
						toErrhandPos(val.Value.Pos()))
				}
			}
		}
	}

	return nil
}

// analyzeCompoundStmt analyzes a compound statement (block).
// It analyzes all statements within the block.
func (a *SemanticAnalyzer) analyzeCompoundStmt(stmt *parser.CompoundStmt) error {
	if stmt == nil {
		return nil
	}

	// Process declarations first
	for _, decl := range stmt.Declarations {
		if err := a.analyzeDeclaration(decl); err != nil {
			return err
		}
	}

	for _, s := range stmt.Statements {
		if err := a.analyzeStatement(s); err != nil {
			return err
		}
	}

	return nil
}

// analyzeStatement analyzes a single statement.
// It dispatches to specific handlers based on statement type.
func (a *SemanticAnalyzer) analyzeStatement(stmt parser.Statement) error {
	if stmt == nil {
		return nil
	}

	switch s := stmt.(type) {
	case *parser.CompoundStmt:
		return a.analyzeCompoundStmt(s)

	case *parser.ExprStmt:
		return a.analyzeExprStmt(s)

	case *parser.ReturnStmt:
		return a.analyzeReturnStmt(s)

	case *parser.IfStmt:
		return a.analyzeIfStmt(s)

	case *parser.WhileStmt:
		return a.analyzeWhileStmt(s)

	case *parser.DoWhileStmt:
		return a.analyzeDoWhileStmt(s)

	case *parser.ForStmt:
		return a.analyzeForStmt(s)

	case *parser.BreakStmt:
		return a.analyzeBreakStmt(s)

	case *parser.ContinueStmt:
		return a.analyzeContinueStmt(s)

	case *parser.GotoStmt:
		return a.analyzeGotoStmt(s)

	case *parser.LabelStmt:
		return a.analyzeLabelStmt(s)

	case *parser.SwitchStmt:
		return a.analyzeSwitchStmt(s)

	case *parser.CaseStmt:
		return a.analyzeCaseStmt(s)

	default:
		// Unknown statement type, skip
		return nil
	}
}

// analyzeExprStmt analyzes an expression statement.
func (a *SemanticAnalyzer) analyzeExprStmt(stmt *parser.ExprStmt) error {
	if stmt.Expr != nil {
		// Check if it's an assignment (lvalue check)
		if assignExpr, ok := stmt.Expr.(*parser.AssignExpr); ok {
			if !a.isLValue(assignExpr.Left) {
				a.errors.Error(errhand.ErrInvalidType,
					"expression is not assignable",
					toErrhandPos(assignExpr.Left.Pos()))
			}
		}
		// Type check the expression
		_ = a.typeChecker.InferType(stmt.Expr)
	}
	return nil
}

// analyzeReturnStmt analyzes a return statement.
// It validates the return type matches the function signature.
func (a *SemanticAnalyzer) analyzeReturnStmt(stmt *parser.ReturnStmt) error {
	// Get expected return type
	if len(a.returnTypes) == 0 {
		// Return outside function
		a.errors.Error(errhand.ErrUndefinedSymbol,
			"return statement outside function",
			toErrhandPos(stmt.Pos()))
		return nil
	}

	expectedType := a.returnTypes[len(a.returnTypes)-1]

	if stmt.Value != nil {
		// Return with value
		returnType := a.typeChecker.InferType(stmt.Value)
		if returnType != nil && expectedType != nil {
			// Check if return type matches
			a.typeChecker.CheckAssignable(expectedType, returnType, stmt.Pos())
		}
	} else {
		// Return without value
		if expectedType != nil {
			expectedBase := a.typeChecker.unwrapType(expectedType)
			if expectedBase.TypeKind() != parser.TypeVoid {
				a.errors.Error(errhand.ErrTypeMismatch,
					fmt.Sprintf("function returning '%s' should not return void", expectedType.String()),
					toErrhandPos(stmt.Pos()))
			}
		}
	}

	return nil
}

// analyzeIfStmt analyzes an if statement.
// It validates the condition type and analyzes then/else branches.
func (a *SemanticAnalyzer) analyzeIfStmt(stmt *parser.IfStmt) error {
	// Check condition type (must be scalar)
	if stmt.Cond != nil {
		condType := a.typeChecker.InferType(stmt.Cond)
		if condType != nil {
			condBase := a.typeChecker.unwrapType(condType)
			if !a.typeChecker.isScalarType(condBase) {
				a.errors.Error(errhand.ErrInvalidType,
					fmt.Sprintf("if condition must be scalar, got '%s'", condType.String()),
					toErrhandPos(stmt.Cond.Pos()))
			}
		}
	}

	// Analyze then branch
	if stmt.Then != nil {
		if err := a.analyzeStatement(stmt.Then); err != nil {
			return err
		}
	}

	// Analyze else branch
	if stmt.Else != nil {
		if err := a.analyzeStatement(stmt.Else); err != nil {
			return err
		}
	}

	return nil
}

// analyzeWhileStmt analyzes a while statement.
// It validates the condition and analyzes the body.
func (a *SemanticAnalyzer) analyzeWhileStmt(stmt *parser.WhileStmt) error {
	// Check condition type
	if stmt.Cond != nil {
		condType := a.typeChecker.InferType(stmt.Cond)
		if condType != nil {
			condBase := a.typeChecker.unwrapType(condType)
			if !a.typeChecker.isScalarType(condBase) {
				a.errors.Error(errhand.ErrInvalidType,
					fmt.Sprintf("while condition must be scalar, got '%s'", condType.String()),
					toErrhandPos(stmt.Cond.Pos()))
			}
		}
	}

	// Enter loop context for break/continue
	a.breakStack = append(a.breakStack, true)
	defer func() {
		a.breakStack = a.breakStack[:len(a.breakStack)-1]
	}()

	// Analyze body
	if stmt.Body != nil {
		if err := a.analyzeStatement(stmt.Body); err != nil {
			return err
		}
	}

	return nil
}

// analyzeDoWhileStmt analyzes a do-while statement.
func (a *SemanticAnalyzer) analyzeDoWhileStmt(stmt *parser.DoWhileStmt) error {
	// Enter loop context for break/continue
	a.breakStack = append(a.breakStack, true)
	defer func() {
		a.breakStack = a.breakStack[:len(a.breakStack)-1]
	}()

	// Analyze body
	if stmt.Body != nil {
		if err := a.analyzeStatement(stmt.Body); err != nil {
			return err
		}
	}

	// Check condition type
	if stmt.Cond != nil {
		condType := a.typeChecker.InferType(stmt.Cond)
		if condType != nil {
			condBase := a.typeChecker.unwrapType(condType)
			if !a.typeChecker.isScalarType(condBase) {
				a.errors.Error(errhand.ErrInvalidType,
					fmt.Sprintf("do-while condition must be scalar, got '%s'", condType.String()),
					toErrhandPos(stmt.Cond.Pos()))
			}
		}
	}

	return nil
}

// analyzeForStmt analyzes a for statement.
// It handles initialization, condition, update, and body.
func (a *SemanticAnalyzer) analyzeForStmt(stmt *parser.ForStmt) error {
	// Enter loop context for break/continue
	a.breakStack = append(a.breakStack, true)
	defer func() {
		a.breakStack = a.breakStack[:len(a.breakStack)-1]
	}()

	// Analyze initialization
	if stmt.Init != nil {
		if initDecl, ok := stmt.Init.(parser.Declaration); ok {
			if err := a.analyzeDeclaration(initDecl); err != nil {
				return err
			}
		} else if initStmt, ok := stmt.Init.(parser.Statement); ok {
			if err := a.analyzeStatement(initStmt); err != nil {
				return err
			}
		} else if initExpr, ok := stmt.Init.(parser.Expr); ok {
			_ = a.typeChecker.InferType(initExpr)
		}
	}

	// Check condition type
	if stmt.Cond != nil {
		condType := a.typeChecker.InferType(stmt.Cond)
		if condType != nil {
			condBase := a.typeChecker.unwrapType(condType)
			if !a.typeChecker.isScalarType(condBase) {
				a.errors.Error(errhand.ErrInvalidType,
					fmt.Sprintf("for condition must be scalar, got '%s'", condType.String()),
					toErrhandPos(stmt.Cond.Pos()))
			}
		}
	}

	// Analyze update expression
	if stmt.Update != nil {
		_ = a.typeChecker.InferType(stmt.Update)
	}

	// Analyze body
	if stmt.Body != nil {
		if err := a.analyzeStatement(stmt.Body); err != nil {
			return err
		}
	}

	return nil
}

// analyzeBreakStmt analyzes a break statement.
// It validates that break is used within a loop or switch.
func (a *SemanticAnalyzer) analyzeBreakStmt(stmt *parser.BreakStmt) error {
	if len(a.breakStack) == 0 {
		a.errors.Error(errhand.ErrUndefinedSymbol,
			"break statement not in loop or switch",
			toErrhandPos(stmt.Pos()))
	}
	return nil
}

// analyzeContinueStmt analyzes a continue statement.
// It validates that continue is used within a loop.
func (a *SemanticAnalyzer) analyzeContinueStmt(stmt *parser.ContinueStmt) error {
	if len(a.breakStack) == 0 {
		a.errors.Error(errhand.ErrUndefinedSymbol,
			"continue statement not in loop",
			toErrhandPos(stmt.Pos()))
	}
	return nil
}

// analyzeGotoStmt analyzes a goto statement.
// It validates the label exists (forward references are allowed in C).
func (a *SemanticAnalyzer) analyzeGotoStmt(stmt *parser.GotoStmt) error {
	// In C, goto can reference labels defined later in the same function
	// We just record the usage; full validation would require two-pass analysis
	if stmt.Label == "" {
		a.errors.Error(errhand.ErrUndefinedSymbol,
			"goto with empty label",
			toErrhandPos(stmt.Pos()))
	}
	return nil
}

// analyzeLabelStmt analyzes a label statement.
// It declares the label in the current function scope.
func (a *SemanticAnalyzer) analyzeLabelStmt(stmt *parser.LabelStmt) error {
	if stmt.Label == "" {
		a.errors.Error(errhand.ErrUndefinedSymbol,
			"label with empty name",
			toErrhandPos(stmt.Pos()))
		return nil
	}

	// Labels have function scope, declare in current scope
	labelSymbol := &Symbol{
		Name:     stmt.Label,
		Kind:     SymbolLabel,
		Type:     nil,
		Position: stmt.Pos(),
		Flags:    FlagNone,
	}

	// Check for duplicate label
	existing := a.symbolTable.Lookup(stmt.Label)
	if existing != nil && existing.Kind == SymbolLabel {
		a.errors.Error(errhand.ErrDuplicateSymbol,
			fmt.Sprintf("duplicate label '%s'", stmt.Label),
			toErrhandPos(stmt.Pos()))
	} else {
		a.symbolTable.Declare(labelSymbol)
	}

	return nil
}

// analyzeSwitchStmt analyzes a switch statement.
func (a *SemanticAnalyzer) analyzeSwitchStmt(stmt *parser.SwitchStmt) error {
	// Check condition type (must be integer)
	if stmt.Cond != nil {
		condType := a.typeChecker.InferType(stmt.Cond)
		if condType != nil {
			condBase := a.typeChecker.unwrapType(condType)
			if !a.typeChecker.isIntegerType(condBase) {
				a.errors.Error(errhand.ErrInvalidType,
					fmt.Sprintf("switch condition must be integer, got '%s'", condType.String()),
					toErrhandPos(stmt.Cond.Pos()))
			}
		}
	}

	// Enter switch context for break
	a.breakStack = append(a.breakStack, true)
	defer func() {
		a.breakStack = a.breakStack[:len(a.breakStack)-1]
	}()

	// Analyze body
	if stmt.Body != nil {
		if err := a.analyzeStatement(stmt.Body); err != nil {
			return err
		}
	}

	return nil
}

// analyzeCaseStmt analyzes a case statement.
// CaseStmt with Value == nil represents a default case.
func (a *SemanticAnalyzer) analyzeCaseStmt(stmt *parser.CaseStmt) error {
	// Check case value type (must be integer constant)
	if stmt.Value != nil {
		valType := a.typeChecker.InferType(stmt.Value)
		if valType != nil {
			valBase := a.typeChecker.unwrapType(valType)
			if !a.typeChecker.isIntegerType(valBase) {
				a.errors.Error(errhand.ErrInvalidType,
					fmt.Sprintf("case value must be integer, got '%s'", valType.String()),
					toErrhandPos(stmt.Value.Pos()))
			}
		}
	}

	// Analyze the statement following the case
	if stmt.Stmt != nil {
		if err := a.analyzeStatement(stmt.Stmt); err != nil {
			return err
		}
	}

	return nil
}

// isLValue checks if an expression is a valid lvalue (assignable).
// Lvalues include: identifiers, dereferences, array indexing, member access.
func (a *SemanticAnalyzer) isLValue(expr parser.Expr) bool {
	if expr == nil {
		return false
	}

	switch e := expr.(type) {
	case *parser.IdentExpr:
		// Identifier is lvalue if it refers to a variable
		symbol := a.Lookup(e.Name)
		return symbol != nil && (symbol.Kind == SymbolVariable || symbol.Kind == SymbolParameter)

	case *parser.UnaryExpr:
		// *ptr is lvalue
		return e.Op == lexer.MUL

	case *parser.IndexExpr:
		// arr[i] is lvalue
		return true

	case *parser.MemberExpr:
		// struct.field or ptr->field is lvalue
		return true

	default:
		return false
	}
}

// EnterScope creates a new scope.
func (a *SemanticAnalyzer) EnterScope() {
	a.symbolTable.PushScope("block")
}

// ExitScope closes the current scope.
func (a *SemanticAnalyzer) ExitScope() {
	a.symbolTable.PopScope()
}

// Lookup looks up a symbol in the current scope chain.
func (a *SemanticAnalyzer) Lookup(name string) *Symbol {
	if a.symbolTable == nil {
		return nil
	}
	return a.symbolTable.Lookup(name)
}

// Declare declares a symbol in the current scope.
func (a *SemanticAnalyzer) Declare(symbol *Symbol) error {
	if a.symbolTable == nil {
		return fmt.Errorf("symbol table not initialized")
	}
	return a.symbolTable.Declare(symbol)
}

// GetSymbolTable returns the symbol table.
func (a *SemanticAnalyzer) GetSymbolTable() *SymbolTable {
	return a.symbolTable
}

// GetErrors returns the error handler.
func (a *SemanticAnalyzer) GetErrors() *errhand.ErrorHandler {
	return a.errors
}

// GetTypeChecker returns the type checker.
func (a *SemanticAnalyzer) GetTypeChecker() *TypeChecker {
	return a.typeChecker
}

