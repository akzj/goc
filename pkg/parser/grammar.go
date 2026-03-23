// Package parser provides parsing of C11 source code into an Abstract Syntax Tree (AST).
// This file documents the C11 grammar rules used by the parser.
package parser

// TODO: Implement grammar reference
// Reference: docs/architecture-design-phases-2-7.md Section 3.5
// Reference: ISO/IEC 9899:2011 specification

// C11 Grammar (EBNF subset)
//
// TranslationUnit = { Declaration } .
//
// Declaration = FunctionDecl | VarDecl | TypeDecl | StructDecl | UnionDecl | EnumDecl .
//
// FunctionDecl = TypeSpec Identifier "(" ParamList? ")" CompoundStmt .
//
// ParamList = ParamDecl { "," ParamDecl } .
//
// ParamDecl = TypeSpec Identifier .
//
// CompoundStmt = "{" { Declaration | Statement } "}" .
//
// Statement = CompoundStmt
//           | IfStmt | WhileStmt | DoWhileStmt | ForStmt | SwitchStmt
//           | ReturnStmt | BreakStmt | ContinueStmt | GotoStmt | LabelStmt
//           | ExprStmt .
//
// IfStmt = "if" "(" Expression ")" Statement ( "else" Statement )? .
//
// WhileStmt = "while" "(" Expression ")" Statement .
//
// DoWhileStmt = "do" Statement "while" "(" Expression ")" ";" .
//
// ForStmt = "for" "(" ExprStmt? Expression? ";" Expression? ")" Statement .
//
// SwitchStmt = "switch" "(" Expression ")" CompoundStmt .
//
// ReturnStmt = "return" Expression? ";" .
//
// ExprStmt = Expression? ";" .
//
// Expression = AssignmentExpr .
//
// AssignmentExpr = ConditionalExpr ( AssignOp ConditionalExpr )* .
//
// AssignOp = "=" | "+=" | "-=" | "*=" | "/=" | "%=" | "&=" | "|=" | "^=" | "<<=" | ">>=" .
//
// ConditionalExpr = LogicalOrExpr ( "?" Expression ":" ConditionalExpr )? .
//
// LogicalOrExpr = LogicalAndExpr { "||" LogicalAndExpr } .
//
// LogicalAndExpr = InclusiveOrExpr { "&&" InclusiveOrExpr } .
//
// InclusiveOrExpr = ExclusiveOrExpr { "|" ExclusiveOrExpr } .
//
// ExclusiveOrExpr = AndExpr { "^" AndExpr } .
//
// AndExpr = EqExpr { "&" EqExpr } .
//
// EqExpr = RelExpr { ( "==" | "!=" ) RelExpr } .
//
// RelExpr = ShiftExpr { ( "<" | ">" | "<=" | ">=" ) ShiftExpr } .
//
// ShiftExpr = AddExpr { ( "<<" | ">>" ) AddExpr } .
//
// AddExpr = MulExpr { ( "+" | "-" ) MulExpr } .
//
// MulExpr = CastExpr { ( "*" | "/" | "%" ) CastExpr } .
//
// CastExpr = "(" Type ")" CastExpr | UnaryExpr .
//
// UnaryExpr = UnaryOp UnaryExpr | PostfixExpr .
//
// UnaryOp = "&" | "*" | "+" | "-" | "~" | "!" | "++" | "--" .
//
// PostfixExpr = PrimaryExpr { ( "[" Expression "]" | "(" ArgList? ")" | "." Identifier | "->" Identifier | "++" | "--" ) } .
//
// ArgList = Expression { "," Expression } .
//
// PrimaryExpr = Identifier | Literal | "(" Expression ")" .
//
// Literal = IntLiteral | FloatLiteral | CharLiteral | StringLiteral .
//
// TypeSpec = "void" | "_Bool" | "char" | "short" | "int" | "long" | "float" | "double"
//          | "signed" | "unsigned" | "struct" Identifier | "union" Identifier | "enum" Identifier
//          | Identifier (typedef name) .
//
// Operator Precedence (highest to lowest):
// 1. Postfix: () [] -> . ++ --
// 2. Unary: & * + - ~ ! ++ -- sizeof
// 3. Cast: (type)
// 4. Multiplicative: * / %
// 5. Additive: + -
// 6. Shift: << >>
// 7. Relational: < > <= >=
// 8. Equality: == !=
// 9. Bitwise AND: &
// 10. Bitwise XOR: ^
// 11. Bitwise OR: |
// 12. Logical AND: &&
// 13. Logical OR: ||
// 14. Conditional: ? :
// 15. Assignment: = += -= *= /= %= &= ^= |= <<= >>=
// 16. Comma: ,
