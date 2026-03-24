// Package parser provides parsing of C11 source code into an Abstract Syntax Tree (AST).
// This file defines the Parser struct and main parsing logic.
package parser

import (
	"fmt"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/lexer"
)

// Parser parses C11 source code tokens into an AST using recursive descent parsing.
// It implements error recovery to continue parsing after encountering errors.
type Parser struct {
	// tokens is the input token stream.
	tokens []lexer.Token
	// pos is the current token position (index into tokens).
	pos int
	// errors is the error handler for reporting parse errors.
	errs *errhand.ErrorHandler
	// ast is the resulting AST (set after successful parse).
	ast *TranslationUnit
}

// NewParser creates a new parser for the given tokens.
// The error handler is used to report parse errors.
func NewParser(tokens []lexer.Token, errorHandler *errhand.ErrorHandler) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
		errs:   errorHandler,
		ast:    nil,
	}
}

// Parse parses the token stream and returns the AST.
// This is the main entry point for parsing.
func (p *Parser) Parse() (*TranslationUnit, error) {
	p.ast = p.ParseTranslationUnit()
	
	// Check if there were any errors during parsing
	if p.errs.HasErrors() {
		return p.ast, fmt.Errorf("parsing failed with %d error(s)", p.errs.ErrorCount())
	}
	
	return p.ast, nil
}

// ParseTranslationUnit parses a translation unit (sequence of declarations).
// TranslationUnit = { Declaration } .
func (p *Parser) ParseTranslationUnit() *TranslationUnit {
	tu := &TranslationUnit{
		Declarations: make([]Declaration, 0),
	}
	
	// Record start position
	if !p.isEOF() {
		tu.pos = p.current().Pos
	}
	
	// Parse declarations until EOF
	for !p.isEOF() {
		// Skip preprocessing directives (basic support)
		if p.current().Type == lexer.PREPROCESSOR {
			p.advance()
			// Skip until newline or EOF
			for !p.isEOF() && p.current().Type != lexer.NEWLINE {
				p.advance()
			}
			if !p.isEOF() {
				p.advance() // consume NEWLINE
			}
			continue
		}

		prev := p.pos
		decl := p.ParseDeclaration()
		if decl != nil {
			tu.Declarations = append(tu.Declarations, decl)
		} else if p.pos == prev && !p.isEOF() {
			// Error recovery failed to advance; avoid infinite loop / OOM.
			p.advance()
		}
	}
	
	// Record end position
	if len(tu.Declarations) > 0 {
		lastDecl := tu.Declarations[len(tu.Declarations)-1]
		tu.end = lastDecl.End()
	} else {
		// Empty translation unit
		if len(p.tokens) > 0 {
			tu.end = p.tokens[len(p.tokens)-1].Pos
		}
	}
	
	return tu
}

// ParseDeclaration parses a single declaration.
// Declaration = FunctionDecl | VarDecl | TypeDecl | StructDecl | UnionDecl | EnumDecl .
func (p *Parser) ParseDeclaration() Declaration {
	// Parse declaration specifiers (type, storage class, etc.)
	specifiers := p.parseDeclarationSpecifiers()
	if specifiers == nil {
		p.synchronize()
		return nil
	}
	
	// Check what kind of declaration this is
	if p.current().Type == lexer.IDENT {
		name := p.current().Value
		namePos := p.current().Pos
		p.advance()

		// Check if this is a function (followed by '(')
		if p.match(lexer.LPAREN) {
			// Function declaration/definition
			params := p.parseParameterList()
			p.expect(lexer.RPAREN)

			// Check for function body or semicolon (parseCompoundStatement consumes '{')
			if p.current().Type == lexer.LBRACE {
				body := p.parseCompoundStatement()
				return p.createFunctionDecl(specifiers, name, namePos, params, body)
			} else {
				// Function declaration
				p.expect(lexer.SEMICOLON)
				return p.createFunctionDecl(specifiers, name, namePos, params, nil)
			}
		} else if p.current().Type == lexer.ASSIGN || p.current().Type == lexer.SEMICOLON || p.current().Type == lexer.COMMA {
			// Variable declaration
			var init Expr
			if p.match(lexer.ASSIGN) {
				init = p.ParseExpression()
			}
			p.expect(lexer.SEMICOLON)
			return p.createVarDecl(specifiers, name, namePos, init)
		} else {
			p.errs.Error("E1001", fmt.Sprintf("unexpected token after identifier: %s", p.current().Type), toErrhandPos(p.current().Pos))
			p.synchronize()
			return nil
		}
	} else if p.match(lexer.LBRACE) {
		// Struct/union/enum with no name (anonymous)
		// This is handled in parseDeclarationSpecifiers
		p.synchronize()
		return nil
	} else {
		p.errs.Error("E1002", "expected identifier after type specifiers", toErrhandPos(p.current().Pos))
		p.synchronize()
		return nil
	}
}

// parseDeclarationSpecifiers parses type specifiers, storage class, qualifiers, etc.
func (p *Parser) parseDeclarationSpecifiers() *DeclSpecifiers {
	specs := &DeclSpecifiers{}
	
	for !p.isEOF() {
		switch p.current().Type {
		// Storage class specifiers
		case lexer.AUTO, lexer.REGISTER, lexer.STATIC, lexer.EXTERN, lexer.TYPEDEF:
			specs.StorageClass = p.current().Type
			p.advance()
		
		// Type qualifiers
		case lexer.CONST, lexer.VOLATILE, lexer.RESTRICT, lexer.ATOMIC:
			specs.Qualifiers = append(specs.Qualifiers, p.current().Type)
			p.advance()
		
		// Type specifiers
		case lexer.VOID, lexer.CHAR, lexer.SHORT, lexer.INT, lexer.LONG,
		     lexer.FLOAT, lexer.DOUBLE, lexer.SIGNED, lexer.UNSIGNED,
		     lexer.BOOL, lexer.COMPLEX, lexer.IMAGINARY:
			specs.TypeSpecifiers = append(specs.TypeSpecifiers, p.current().Type)
			p.advance()
		
		// Complex types
		case lexer.STRUCT, lexer.UNION, lexer.ENUM:
			specs.ComplexType = p.parseComplexType()
		
		default:
			// End of specifiers
			return specs
		}
	}
	
	return specs
}

// parseComplexType parses struct, union, or enum types.
func (p *Parser) parseComplexType() Type {
	switch p.current().Type {
	case lexer.STRUCT:
		return p.parseStructType()
	case lexer.UNION:
		return p.parseUnionType()
	case lexer.ENUM:
		return p.parseEnumType()
	default:
		return nil
	}
}

// parseStructType parses a struct type.
func (p *Parser) parseStructType() Type {
	p.expect(lexer.STRUCT)
	
	var name string
	if p.current().Type == lexer.IDENT {
		name = p.current().Value
		p.advance()
	}

	st := &StructType{
		Name: name,
	}

	if p.current().Type == lexer.LBRACE {
		p.advance() // '{'
		st.Fields = p.parseFieldList()
		if p.current().Type == lexer.RBRACE {
			p.advance()
		} else {
			p.expect(lexer.RBRACE)
		}
	}

	return st
}

// parseUnionType parses a union type.
func (p *Parser) parseUnionType() Type {
	p.expect(lexer.UNION)

	var name string
	if p.current().Type == lexer.IDENT {
		name = p.current().Value
		p.advance()
	}

	st := &StructType{
		Name:    name,
		IsUnion: true,
	}

	if p.current().Type == lexer.LBRACE {
		p.advance()
		st.Fields = p.parseFieldList()
		if p.current().Type == lexer.RBRACE {
			p.advance()
		} else {
			p.expect(lexer.RBRACE)
		}
	}

	return st
}

// parseEnumType parses an enum type.
func (p *Parser) parseEnumType() Type {
	p.expect(lexer.ENUM)

	var name string
	if p.current().Type == lexer.IDENT {
		name = p.current().Value
		p.advance()
	}
	
	et := &EnumType{
		Name:   name,
		Values: make([]*EnumValue, 0),
	}
	
	if p.current().Type == lexer.LBRACE {
		p.advance()
		et.Values = p.parseEnumValues()
		if p.current().Type == lexer.RBRACE {
			p.advance()
		} else {
			p.expect(lexer.RBRACE)
		}
	}

	return et
}

// parseFieldList parses struct/union fields (inside braces; closing '}' not yet consumed).
func (p *Parser) parseFieldList() []*FieldDecl {
	fields := make([]*FieldDecl, 0)

	for !p.isEOF() && p.current().Type != lexer.RBRACE {
		field := p.parseFieldDeclaration()
		if field != nil {
			fields = append(fields, field)
		} else {
			// If we couldn't parse a field, advance to avoid infinite loop
			p.advance()
		}
	}

	return fields
}

// parseFieldDeclaration parses a single field declaration.
func (p *Parser) parseFieldDeclaration() *FieldDecl {
	// Parse type specifiers for field
	specifiers := p.parseDeclarationSpecifiers()
	if specifiers == nil {
		return nil
	}
	
	if p.current().Type != lexer.IDENT {
		p.errs.Error("E1003", "expected field name", toErrhandPos(p.current().Pos))
		return nil
	}

	name := p.current().Value
	namePos := p.current().Pos
	p.advance()
	
	// Check for bitfield
	var bitWidth Expr
	if p.match(lexer.COLON) {
		bitWidth = p.ParseExpression()
	}
	
	p.expect(lexer.SEMICOLON)
	
	// Create field type from specifiers
	fieldType := p.specifiersToType(specifiers)
	
	return &FieldDecl{
		pos:      namePos,
		end:      namePos,
		Type:     fieldType,
		Name:     name,
		BitWidth: bitWidth,
	}
}

// parseEnumValues parses enum constant list (inside braces; closing '}' not yet consumed).
func (p *Parser) parseEnumValues() []*EnumValue {
	values := make([]*EnumValue, 0)

	for !p.isEOF() && p.current().Type != lexer.RBRACE {
		if p.current().Type != lexer.IDENT {
			break
		}

		name := p.current().Value
		namePos := p.current().Pos
		p.advance()
		
		var value Expr
		if p.match(lexer.ASSIGN) {
			value = p.ParseExpression()
		}
		
		values = append(values, &EnumValue{
			pos:   namePos,
			end:   namePos,
			Name:  name,
			Value: value,
		})
		
		if !p.match(lexer.COMMA) {
			break
		}
	}
	
	return values
}

// createFunctionDecl creates a FunctionDecl from parsed components.
func (p *Parser) createFunctionDecl(specs *DeclSpecifiers, name string, namePos lexer.Position, params []*ParamDecl, body *CompoundStmt) *FunctionDecl {
	retType := p.specifiersToType(specs)
	paramTypes := make([]Type, 0, len(params))
	for _, param := range params {
		if param != nil && param.Type != nil {
			paramTypes = append(paramTypes, param.Type)
		} else {
			paramTypes = append(paramTypes, nil)
		}
	}
	fnType := &FuncType{
		Return: retType,
		Params: paramTypes,
	}
	return &FunctionDecl{
		pos:      namePos,
		end:      namePos,
		Type:     fnType,
		Name:     name,
		Params:   params,
		Body:     body,
		IsInline: specs != nil && specs.IsInline,
		IsStatic: specs != nil && specs.StorageClass == lexer.STATIC,
		IsExtern: specs != nil && specs.StorageClass == lexer.EXTERN,
	}
}

// createVarDecl creates a VarDecl from parsed components.
func (p *Parser) createVarDecl(specs *DeclSpecifiers, name string, namePos lexer.Position, init Expr) *VarDecl {
	return &VarDecl{
		pos:      namePos,
		end:      namePos,
		Type:     p.specifiersToType(specs),
		Name:     name,
		Init:     init,
		IsStatic: specs != nil && specs.StorageClass == lexer.STATIC,
		IsExtern: specs != nil && specs.StorageClass == lexer.EXTERN,
		IsConst:  specs != nil && specs.hasQualifier(lexer.CONST),
	}
}

// specifiersToType converts declaration specifiers to a Type.
func (p *Parser) specifiersToType(specs *DeclSpecifiers) Type {
	if specs == nil {
		return &BaseType{Kind: TypeInt, Signed: true}
	}
	
	// Handle complex types
	if specs.ComplexType != nil {
		return specs.ComplexType
	}
	
	// Determine base type from type specifiers
	baseType := &BaseType{Kind: TypeInt, Signed: true}
	
	for _, ts := range specs.TypeSpecifiers {
		switch ts {
		case lexer.VOID:
			baseType.Kind = TypeVoid
		case lexer.CHAR:
			baseType.Kind = TypeChar
		case lexer.SHORT:
			baseType.Kind = TypeShort
		case lexer.INT:
			baseType.Kind = TypeInt
		case lexer.LONG:
			if baseType.Kind == TypeLong {
				baseType.Long = 2 // long long
			} else {
				baseType.Kind = TypeLong
				baseType.Long = 1
			}
		case lexer.FLOAT:
			baseType.Kind = TypeFloat
		case lexer.DOUBLE:
			baseType.Kind = TypeDouble
		case lexer.SIGNED:
			baseType.Signed = true
		case lexer.UNSIGNED:
			baseType.Signed = false
		case lexer.BOOL:
			baseType.Kind = TypeBool
		}
	}
	
	// Apply qualifiers
	if len(specs.Qualifiers) > 0 {
		return &QualifiedType{
			Type:       baseType,
			IsConst:    specs.hasQualifier(lexer.CONST),
			IsVolatile: specs.hasQualifier(lexer.VOLATILE),
			IsAtomic:   specs.hasQualifier(lexer.ATOMIC),
		}
	}
	
	return baseType
}

// parseParameterList parses function parameter list.
func (p *Parser) parseParameterList() []*ParamDecl {
	params := make([]*ParamDecl, 0)
	
	// Check for void parameter list
	if p.match(lexer.VOID) {
		return params
	}
	
	// Check for empty parameter list (K&R style)
	// Use peek (current) instead of match to avoid consuming ')'
	if p.current().Type == lexer.RPAREN {
		return params  // Don't consume, let caller consume it
	}
	
	// Parse parameters
	// Use current().Type != RPAREN instead of !match(RPAREN) to avoid consuming ')'
	for !p.isEOF() && p.current().Type != lexer.RPAREN {
		param := p.parseParameter()
		if param != nil {
			params = append(params, param)
		}
		
		if p.current().Type != lexer.COMMA {
			break
		}
		p.advance()  // Consume comma
	}
	
	return params
}

// parseParameter parses a single function parameter.
func (p *Parser) parseParameter() *ParamDecl {
	specifiers := p.parseDeclarationSpecifiers()
	if specifiers == nil {
		return nil
	}
	
	var name string
	var namePos lexer.Position
	
	if p.current().Type == lexer.IDENT {
		name = p.current().Value
		namePos = p.current().Pos
		p.advance()
	}

	return &ParamDecl{
		pos:  namePos,
		end:  namePos,
		Type: p.specifiersToType(specifiers),
		Name: name,
	}
}

// parseCompoundStatement parses a compound statement (block).
func (p *Parser) parseCompoundStatement() *CompoundStmt {
	p.expect(lexer.LBRACE)
	
	stmt := &CompoundStmt{
		Statements:   make([]Statement, 0),
		Declarations: make([]Declaration, 0),
	}
	
	stmt.pos = p.current().Pos
	
	// Do not use match(RBRACE) here: it would consume '}' before the loop exits,
	// then expect(RBRACE) would run past the closing brace and desynchronize
	// the rest of the translation unit (leading to tight error-recovery loops).
	for !p.isEOF() && p.current().Type != lexer.RBRACE {
		if p.isDeclaration() {
			decl := p.ParseDeclaration()
			if decl != nil {
				stmt.Declarations = append(stmt.Declarations, decl)
			}
		} else {
			s := p.ParseStatement()
			if s != nil {
				stmt.Statements = append(stmt.Statements, s)
			}
		}
	}

	if p.current().Type == lexer.RBRACE {
		rbrace := p.advance()
		stmt.end = rbrace.Pos
	} else {
		p.expect(lexer.RBRACE)
		stmt.end = p.current().Pos
	}

	return stmt
}

// isDeclaration checks if the current position starts a declaration.
func (p *Parser) isDeclaration() bool {
	if p.isEOF() {
		return false
	}
	
	t := p.current().Type
	
	// Storage class specifiers
	if t == lexer.AUTO || t == lexer.REGISTER || t == lexer.STATIC ||
		t == lexer.EXTERN || t == lexer.TYPEDEF {
		return true
	}
	
	// Type qualifiers
	if t == lexer.CONST || t == lexer.VOLATILE || t == lexer.RESTRICT || t == lexer.ATOMIC {
		return true
	}
	
	// Type specifiers
	if t == lexer.VOID || t == lexer.CHAR || t == lexer.SHORT || t == lexer.INT ||
		t == lexer.LONG || t == lexer.FLOAT || t == lexer.DOUBLE || t == lexer.SIGNED ||
		t == lexer.UNSIGNED || t == lexer.BOOL || t == lexer.COMPLEX || t == lexer.IMAGINARY {
		return true
	}
	
	// Complex types
	if t == lexer.STRUCT || t == lexer.UNION || t == lexer.ENUM {
		return true
	}
	
	// Identifier (could be typedef name or K&R function declaration)
	if t == lexer.IDENT {
		// Look ahead to determine if it's a declaration
		// If followed by DOT, ARROW, or LPAREN, it's an expression, not a declaration
		next := p.peek(1)
		if next.Type == lexer.DOT || next.Type == lexer.ARROW || next.Type == lexer.LPAREN {
			return false
		}
		return true
	}
	
	return false
}

// DeclSpecifiers holds declaration specifiers.
type DeclSpecifiers struct {
	StorageClass   lexer.TokenType
	TypeSpecifiers []lexer.TokenType
	Qualifiers     []lexer.TokenType
	ComplexType    Type
	IsInline       bool
}

// hasQualifier checks if a qualifier is present.
func (d *DeclSpecifiers) hasQualifier(q lexer.TokenType) bool {
	for _, qual := range d.Qualifiers {
		if qual == q {
			return true
		}
	}
	return false
}

// Helper methods for token navigation

// current returns the current token.
func (p *Parser) current() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.EOF}
	}
	return p.tokens[p.pos]
}

// peek returns the token at the given offset from current.
func (p *Parser) peek(offset int) lexer.Token {
	idx := p.pos + offset
	if idx >= len(p.tokens) {
		return lexer.Token{Type: lexer.EOF}
	}
	return p.tokens[idx]
}

// advance consumes and returns the current token.
func (p *Parser) advance() lexer.Token {
	tok := p.current()
	p.pos++
	return tok
}

// match advances if the current token matches any of the given types.
func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.current().Type == t {
			p.advance()
			return true
		}
	}
	return false
}

// expect advances and returns the current token, or reports an error if it doesn't match.
func (p *Parser) expect(t lexer.TokenType) lexer.Token {
	if p.current().Type == t {
		return p.advance()
	}
	
	p.errs.Error("E1004", fmt.Sprintf("expected %s, got %s", t, p.current().Type), toErrhandPos(p.current().Pos))
	return p.advance()
}

// isEOF returns true if the parser is at the end of the token stream.
func (p *Parser) isEOF() bool {
	return p.current().Type == lexer.EOF
}

// synchronize skips tokens until a synchronization point is reached.
// This is used for error recovery.
func (p *Parser) synchronize() {
	// Skip until we find a synchronization point
	for !p.isEOF() {
		switch p.current().Type {
		case lexer.SEMICOLON:
			p.advance()
			return
		case lexer.RBRACE:
			return
		case lexer.IF, lexer.WHILE, lexer.DO, lexer.FOR, lexer.RETURN,
			lexer.SWITCH, lexer.BREAK, lexer.CONTINUE, lexer.GOTO:
			return
		}
		p.advance()
	}
}

// toErrhandPos converts lexer.Position to errhand.Position.
func toErrhandPos(pos lexer.Position) errhand.Position {
	return errhand.Position{
		File:   pos.File,
		Line:   pos.Line,
		Column: pos.Column,
	}
}