// Package error provides error handling and diagnostic reporting for the GOC compiler.
// This file defines predefined error codes.
package errhand

// Lexer error codes (E0001-E0999)
const (
	ErrInvalidChar     ErrorCode = "E0001"
	ErrUnterminatedStr ErrorCode = "E0002"
	ErrUnterminatedChar ErrorCode = "E0003"
	ErrInvalidNumber   ErrorCode = "E0004"
	ErrInvalidEscape   ErrorCode = "E0005"
)

// Parser error codes (E1001-E1999)
const (
	ErrSyntaxError     ErrorCode = "E1001"
	ErrUnexpectedToken ErrorCode = "E1002"
	ErrExpectedToken   ErrorCode = "E1003"
	ErrIncompleteExpr  ErrorCode = "E1004"
	ErrIncompleteStmt  ErrorCode = "E1005"
)

// Semantic error codes (E2001-E2999)
const (
	ErrUndefinedSymbol ErrorCode = "E2001"
	ErrDuplicateSymbol ErrorCode = "E2002"
	ErrTypeMismatch    ErrorCode = "E2003"
	ErrInvalidType     ErrorCode = "E2004"
	ErrConstViolation  ErrorCode = "E2005"
)

// IR error codes (E3001-E3999)
const (
	ErrInvalidIR       ErrorCode = "E3001"
	ErrControlFlow     ErrorCode = "E3002"
	ErrUndefinedLabel  ErrorCode = "E3003"
)

// CodeGen error codes (E4001-E4999)
const (
	ErrUnsupportedOp   ErrorCode = "E4001"
	ErrRegAlloc        ErrorCode = "E4002"
)

// Linker error codes (E5001-E5999)
const (
	ErrUndefinedRef    ErrorCode = "E5001"
	ErrLinkerDuplicateSymbol ErrorCode = "E5002"
	ErrInvalidElf      ErrorCode = "E5003"
)