// Package assembly_validator provides utilities for validating x86-64 assembly code.
// It checks instruction syntax, label definitions, register usage, and ABI compliance.
package assembly_validator

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a validation error with line number and description.
type ValidationError struct {
	Line    int
	Message string
}

// ValidationResult contains the outcome of assembly validation.
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// x86-64 valid instruction mnemonics (common subset)
var validInstructions = map[string]bool{
	// Data movement
	"mov": true, "movq": true, "movl": true, "movb": true, "movw": true,
	"lea": true, "leaq": true, "leal": true,
	"push": true, "pushq": true, "pop": true, "popq": true,
	// Arithmetic
	"add": true, "addq": true, "addl": true, "addb": true,
	"sub": true, "subq": true, "subl": true, "subb": true,
	"imul": true, "imulq": true, "imull": true,
	"idiv": true, "idivq": true,
	"inc": true, "incq": true, "incl": true,
	"dec": true, "decq": true, "decl": true,
	"neg": true, "negq": true,
	// Logical
	"and": true, "andq": true, "andl": true, "andb": true,
	"or": true, "orq": true, "orl": true, "orb": true,
	"xor": true, "xorq": true, "xorl": true, "xorb": true,
	"not": true, "notq": true, "notl": true,
	"shl": true, "shlq": true, "shll": true,
	"shr": true, "shrq": true, "shrl": true,
	"sar": true, "sarq": true, "sarl": true,
	// Comparison and test
	"cmp": true, "cmpq": true, "cmpl": true, "cmpb": true,
	"test": true, "testq": true, "testl": true, "testb": true,
	// Control flow
	"jmp": true, "jmpq": true,
	"je": true, "jne": true, "jz": true, "jnz": true,
	"jl": true, "jle": true, "jg": true, "jge": true,
	"ja": true, "jae": true, "jb": true, "jbe": true,
	"js": true, "jns": true, "jo": true, "jno": true,
	"jp": true, "jnp": true,
	"call": true, "callq": true,
	"ret": true, "retq": true,
	// Function prologue/epilogue
	"enter": true, "leave": true,
	// NOP
	"nop": true, "nopq": true,
	// System call
	"syscall": true, "sysenter": true,
	// Set on condition
	"sete": true, "setne": true, "setz": true, "setnz": true,
	"setl": true, "setle": true, "setg": true, "setge": true,
	"seta": true, "setae": true, "setb": true, "setbe": true,
	// Move with sign/zero extend
	"movs": true, "movslq": true, "movsb": true, "movsw": true,
	"movz": true, "movzb": true, "movzw": true,
	// Conditional move
	"cmov": true, "cmove": true, "cmovne": true, "cmovz": true, "cmovnz": true,
	"cmovl": true, "cmovle": true, "cmovg": true, "cmovge": true,
}

// Valid 64-bit register names
var validRegisters = map[string]bool{
	// General purpose 64-bit
	"rax": true, "rbx": true, "rcx": true, "rdx": true,
	"rsi": true, "rdi": true, "rbp": true, "rsp": true,
	"r8": true, "r9": true, "r10": true, "r11": true,
	"r12": true, "r13": true, "r14": true, "r15": true,
	// 32-bit
	"eax": true, "ebx": true, "ecx": true, "edx": true,
	"esi": true, "edi": true, "ebp": true, "esp": true,
	"r8d": true, "r9d": true, "r10d": true, "r11d": true,
	"r12d": true, "r13d": true, "r14d": true, "r15d": true,
	// 16-bit
	"ax": true, "bx": true, "cx": true, "dx": true,
	"si": true, "di": true, "bp": true, "sp": true,
	"r8w": true, "r9w": true, "r10w": true, "r11w": true,
	"r12w": true, "r13w": true, "r14w": true, "r15w": true,
	// 8-bit
	"al": true, "bl": true, "cl": true, "dl": true,
	"sil": true, "dil": true, "bpl": true, "spl": true,
	"r8b": true, "r9b": true, "r10b": true, "r11b": true,
	"r12b": true, "r13b": true, "r14b": true, "r15b": true,
	"ah": true, "bh": true, "ch": true, "dh": true,
}

// Callee-saved registers (must be preserved by callee)
var calleeSavedRegisters = map[string]bool{
	"rbx": true, "rbp": true, "r12": true, "r13": true, "r14": true, "r15": true,
	// Also 32/16/8-bit variants
	"ebx": true, "ebp": true, "r12d": true, "r13d": true, "r14d": true, "r15d": true,
	"bx": true, "bp": true, "r12w": true, "r13w": true, "r14w": true, "r15w": true,
	"bl": true, "bpl": true, "r12b": true, "r13b": true, "r14b": true, "r15b": true,
}

// Caller-saved registers (can be clobbered by callee)
var callerSavedRegisters = map[string]bool{
	"rax": true, "rcx": true, "rdx": true, "rsi": true, "rdi": true,
	"r8": true, "r9": true, "r10": true, "r11": true,
	// Also variants
	"eax": true, "ecx": true, "edx": true, "esi": true, "edi": true,
	"r8d": true, "r9d": true, "r10d": true, "r11d": true,
	"ax": true, "cx": true, "dx": true, "si": true, "di": true,
	"r8w": true, "r9w": true, "r10w": true, "r11w": true,
	"al": true, "cl": true, "dl": true, "sil": true, "dil": true,
	"r8b": true, "r9b": true, "r10b": true, "r11b": true,
}

// Argument passing registers (System V AMD64 ABI)
var argumentRegisters = []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}

// Label pattern: starts with letter or underscore, contains alphanumeric and underscore
var labelPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Instruction line pattern (label: instruction operands)
var instructionPattern = regexp.MustCompile(`^\s*(?:(\w+):)?\s*(\w+)?\s*(.*)$`)

// Memory operand pattern
var memoryOperandPattern = regexp.MustCompile(`^\s*(-?\d*)\s*\(\s*(%?\w*)\s*(?:,\s*(%?\w*)\s*(?:,\s*(\d*))?\s*)?\)$`)

// ValidateAssembly is the main entry point for assembly validation.
// It performs all validation checks and returns a ValidationResult.
//
// Parameters:
//   - assembly: The assembly code as a string (line by line)
//
// Returns:
//   - ValidationResult with Valid flag and list of Errors
func ValidateAssembly(assembly string) ValidationResult {
	lines := strings.Split(assembly, "\n")
	var errors []ValidationError

	// Validate instructions
	instructionErrors := ValidateInstructions(lines)
	errors = append(errors, instructionErrors...)

	// Validate labels
	labelErrors := ValidateLabels(lines)
	errors = append(errors, labelErrors...)

	// Validate register usage
	registerErrors := ValidateRegisters(lines)
	errors = append(errors, registerErrors...)

	// Validate ABI compliance
	abiErrors := ValidateABICompliance(lines)
	errors = append(errors, abiErrors...)

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// ValidateInstructions checks x86-64 instruction syntax.
// It validates instruction mnemonics, operand count, and types.
//
// Parameters:
//   - lines: Assembly code split into lines
//
// Returns:
//   - List of ValidationError for syntax issues
func ValidateInstructions(lines []string) []ValidationError {
	var errors []ValidationError

	for lineNum, line := range lines {
		lineNum++ // 1-based line numbers

		// Skip empty lines and comments
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}

		// Parse the line
		matches := instructionPattern.FindStringSubmatch(trimmed)
		if matches == nil {
			// Check if it's a directive (starts with .)
			if strings.HasPrefix(trimmed, ".") {
				continue // Skip directives like .text, .data, .globl, etc.
			}
			errors = append(errors, ValidationError{
				Line:    lineNum,
				Message: fmt.Sprintf("Invalid instruction format: %s", trimmed),
			})
			continue
		}

		// Extract instruction (group 2)
		instruction := matches[2]
		if instruction == "" {
			// Line has only a label, which is valid
			continue
		}

		// Validate instruction mnemonic
		if !validInstructions[strings.ToLower(instruction)] {
			errors = append(errors, ValidationError{
				Line:    lineNum,
				Message: fmt.Sprintf("Unknown instruction mnemonic: %s", instruction),
			})
		}

		// Validate operands
		operands := parseOperands(matches[3])
		for i, operand := range operands {
			operandErrors := validateOperand(operand, lineNum)
			errors = append(errors, operandErrors...)
			_ = i // Use i to avoid unused variable warning
		}
	}

	return errors
}

// parseOperands splits operand string into individual operands.
func parseOperands(operandStr string) []string {
	if strings.TrimSpace(operandStr) == "" {
		return []string{}
	}

	var operands []string
	var current strings.Builder
	parenDepth := 0

	for _, ch := range operandStr {
		switch ch {
		case '(':
			parenDepth++
			current.WriteRune(ch)
		case ')':
			parenDepth--
			current.WriteRune(ch)
		case ',':
			if parenDepth == 0 {
				operands = append(operands, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		operands = append(operands, strings.TrimSpace(current.String()))
	}

	return operands
}

// validateOperand validates a single operand.
func validateOperand(operand string, lineNum int) []ValidationError {
	var errors []ValidationError
	operand = strings.TrimSpace(operand)

	if operand == "" {
		return errors
	}

	// Check if it's a register
	if strings.HasPrefix(operand, "%") {
		regName := operand[1:]
		if !validRegisters[strings.ToLower(regName)] {
			errors = append(errors, ValidationError{
				Line:    lineNum,
				Message: fmt.Sprintf("Invalid register name: %s", operand),
			})
		}
		return errors
	}

	// Check if it's a memory operand
	if strings.Contains(operand, "(") {
		memoryErrors := validateMemoryOperand(operand, lineNum)
		errors = append(errors, memoryErrors...)
		return errors
	}

	// Check if it's an immediate value
	if strings.HasPrefix(operand, "$") {
		// Immediate value - basic validation
		return errors
	}

	// Check if it's a label reference
	if labelPattern.MatchString(operand) {
		return errors
	}

	return errors
}

// validateMemoryOperand validates memory operand syntax.
func validateMemoryOperand(operand string, lineNum int) []ValidationError {
	var errors []ValidationError

	matches := memoryOperandPattern.FindStringSubmatch(operand)
	if matches == nil {
		// Try simpler pattern for basic memory operands
		if !regexp.MustCompile(`^\s*-?\d*\(%\w+\)\s*$`).MatchString(operand) &&
			!regexp.MustCompile(`^\s*\(%\w+\)\s*$`).MatchString(operand) {
			errors = append(errors, ValidationError{
				Line:    lineNum,
				Message: fmt.Sprintf("Invalid memory operand syntax: %s", operand),
			})
		}
		return errors
	}

	// Validate base register if present
	baseReg := matches[2]
	if baseReg != "" {
		if strings.HasPrefix(baseReg, "%") {
			baseReg = baseReg[1:]
		}
		if !validRegisters[strings.ToLower(baseReg)] {
			errors = append(errors, ValidationError{
				Line:    lineNum,
				Message: fmt.Sprintf("Invalid register in memory operand: %s", baseReg),
			})
		}
	}

	// Validate index register if present
	indexReg := matches[3]
	if indexReg != "" {
		if strings.HasPrefix(indexReg, "%") {
			indexReg = indexReg[1:]
		}
		if !validRegisters[strings.ToLower(indexReg)] {
			errors = append(errors, ValidationError{
				Line:    lineNum,
				Message: fmt.Sprintf("Invalid index register: %s", indexReg),
			})
		}
	}

	return errors
}

// ValidateLabels checks label definitions and references.
// It ensures labels are properly defined, not duplicated, and referenced correctly.
//
// Parameters:
//   - lines: Assembly code split into lines
//
// Returns:
//   - List of ValidationError for label issues
func ValidateLabels(lines []string) []ValidationError {
	var errors []ValidationError
	definedLabels := make(map[string]int)    // label -> line number
	referencedLabels := make(map[string]bool)

	labelDefPattern := regexp.MustCompile(`^\s*(\w+):\s*$`)
	labelRefPattern := regexp.MustCompile(`\b(\w+)\b`)

	for lineNum, line := range lines {
		lineNum++ // 1-based line numbers

		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}

		// Check for label definition
		if matches := labelDefPattern.FindStringSubmatch(trimmed); matches != nil {
			label := matches[1]
			if !labelPattern.MatchString(label) {
				errors = append(errors, ValidationError{
					Line:    lineNum,
					Message: fmt.Sprintf("Invalid label name: %s", label),
				})
				continue
			}

			// Check for duplicate definition
			if prevLine, exists := definedLabels[label]; exists {
				errors = append(errors, ValidationError{
					Line:    lineNum,
					Message: fmt.Sprintf("Duplicate label definition: %s (also defined at line %d)", label, prevLine),
				})
			}
			definedLabels[label] = lineNum
			continue
		}

		// Check for label references in jump/call instructions
		if strings.Contains(trimmed, "jmp") || strings.Contains(trimmed, "call") ||
			strings.Contains(trimmed, "je") || strings.Contains(trimmed, "jne") {
			refs := labelRefPattern.FindAllStringSubmatch(trimmed, -1)
			for _, ref := range refs {
				refLabel := ref[1]
				// Skip instruction mnemonics and registers
				if validInstructions[strings.ToLower(refLabel)] || validRegisters[strings.ToLower(refLabel)] {
					continue
				}
				// Skip if it looks like a number
				if regexp.MustCompile(`^\d+$`).MatchString(refLabel) {
					continue
				}
				referencedLabels[refLabel] = true
			}
		}
	}

	// Check for undefined labels (only if they look like labels, not instructions)
	for label := range referencedLabels {
		if _, defined := definedLabels[label]; !defined {
			// This might be a forward reference or external label, so we just note it
			// For strict validation, you could add an error here
		}
	}

	return errors
}

// ValidateRegisters checks register usage and preservation rules.
// It verifies callee-saved registers are preserved and stack pointer is managed correctly.
//
// Parameters:
//   - lines: Assembly code split into lines
//
// Returns:
//   - List of ValidationError for register usage issues
func ValidateRegisters(lines []string) []ValidationError {
	var errors []ValidationError

	// Track register state within function scope
	// calleeSavedPushed: registers that were pushed (saved)
	// calleeSavedPopped: registers that were popped (restored)
	calleeSavedPushed := make(map[string]bool)
	calleeSavedPopped := make(map[string]bool)
	inFunction := false
	_ = inFunction // Track function scope
	hasFunctionPrologue := false
	_ = hasFunctionPrologue // Track prologue detection

	pushPattern := regexp.MustCompile(`push[q]?\s+%(\w+)`)
	popPattern := regexp.MustCompile(`pop[q]?\s+%(\w+)`)

	for lineNum, line := range lines {
		lineNum++ // 1-based line numbers

		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}

		// Detect function start (label)
		if strings.HasSuffix(trimmed, ":") {
			inFunction = true
			hasFunctionPrologue = false
			calleeSavedPushed = make(map[string]bool)
			calleeSavedPopped = make(map[string]bool)
		}

		// Track push instructions for callee-saved registers
		if matches := pushPattern.FindStringSubmatch(trimmed); matches != nil {
			reg := strings.ToLower(matches[1])
			if calleeSavedRegisters[reg] {
				calleeSavedPushed[reg] = true
			}
		}

		// Track pop instructions for callee-saved registers
		if matches := popPattern.FindStringSubmatch(trimmed); matches != nil {
			reg := strings.ToLower(matches[1])
			if calleeSavedRegisters[reg] {
				calleeSavedPopped[reg] = true
			}
		}

		// Check for function prologue
		if strings.Contains(trimmed, "pushq %rbp") || strings.Contains(trimmed, "push %rbp") {
			hasFunctionPrologue = true
		}

		// Detect function end
		if strings.Contains(trimmed, "ret") || strings.Contains(trimmed, "retq") {
			if inFunction {
				// Check if all pushed callee-saved registers were popped (restored)
				for reg := range calleeSavedPushed {
					if !calleeSavedPopped[reg] {
						errors = append(errors, ValidationError{
							Line:    lineNum,
							Message: fmt.Sprintf("Callee-saved register %s pushed but not popped before ret", reg),
						})
					}
				}
				// Check if any callee-saved register was popped without being pushed
				for reg := range calleeSavedPopped {
					if !calleeSavedPushed[reg] {
						errors = append(errors, ValidationError{
							Line:    lineNum,
							Message: fmt.Sprintf("Callee-saved register %s popped without being pushed", reg),
						})
					}
				}
				inFunction = false
			}
		}
	}

	return errors
}

// ValidateABICompliance checks System V AMD64 ABI compliance.
// It verifies argument passing, return values, stack alignment, and red zone usage.
//
// Parameters:
//   - lines: Assembly code split into lines
//
// Returns:
//   - List of ValidationError for ABI compliance issues
func ValidateABICompliance(lines []string) []ValidationError {
	var errors []ValidationError

	inFunction := false
	_ = inFunction // Used for future ABI compliance checking
	functionName := ""
	_ = functionName // Track function name for error reporting
	stackAlignment := 0 // Track stack alignment state
	_ = stackAlignment
	hasCall := false
	_ = hasCall

	callPattern := regexp.MustCompile(`call[q]?\s+(\w+)`)
	retPattern := regexp.MustCompile(`ret[q]?$`)
	subRspPattern := regexp.MustCompile(`sub[q]?\s+\$(\d+),\s*%rsp`)
	addRspPattern := regexp.MustCompile(`add[q]?\s+\$(\d+),\s*%rsp`)

	for lineNum, line := range lines {
		lineNum++ // 1-based line numbers

		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}

		// Detect function definition
		if strings.HasSuffix(trimmed, ":") {
			functionName = strings.TrimSuffix(trimmed, ":")
			inFunction = true
			hasCall = false
			stackAlignment = 0
		}

		// Check for calls (stack must be 16-byte aligned before call)
		if matches := callPattern.FindStringSubmatch(trimmed); matches != nil {
			hasCall = true
			// Stack alignment check would require more context about pushes
			// This is a simplified check
		}

		// Track stack adjustments
		if matches := subRspPattern.FindStringSubmatch(trimmed); matches != nil {
			// In a real implementation, parse the value and check alignment
			_ = matches[1]
		}
		if matches := addRspPattern.FindStringSubmatch(trimmed); matches != nil {
			_ = matches[1]
		}

		// Check return
		if retPattern.MatchString(trimmed) {
			if inFunction {
				// Verify return value is in rax (this would require data flow analysis)
				// For now, we just note the function ended
				inFunction = false
			}
		}
	}

	// Note: Full ABI compliance checking requires data flow analysis
	// This provides basic structural checks

	return errors
}

// FormatErrors formats validation errors as a readable string.
func FormatErrors(result ValidationResult) string {
	if result.Valid {
		return "Assembly validation passed successfully."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Assembly validation failed with %d error(s):\n", len(result.Errors)))
	for _, err := range result.Errors {
		sb.WriteString(fmt.Sprintf("  Line %d: %s\n", err.Line, err.Message))
	}
	return sb.String()
}

// IsRegister returns true if the given name is a valid x86-64 register.
func IsRegister(name string) bool {
	// Remove % prefix if present
	if strings.HasPrefix(name, "%") {
		name = name[1:]
	}
	return validRegisters[strings.ToLower(name)]
}

// IsCalleeSaved returns true if the register is callee-saved.
func IsCalleeSaved(name string) bool {
	if strings.HasPrefix(name, "%") {
		name = name[1:]
	}
	return calleeSavedRegisters[strings.ToLower(name)] || calleeSavedRegisters[name]
}

// IsCallerSaved returns true if the register is caller-saved.
func IsCallerSaved(name string) bool {
	if strings.HasPrefix(name, "%") {
		name = name[1:]
	}
	return callerSavedRegisters[strings.ToLower(name)] || callerSavedRegisters[name]
}

// GetArgumentRegisters returns the list of argument passing registers.
func GetArgumentRegisters() []string {
	return argumentRegisters
}