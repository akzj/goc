package assembly_validator

import (
	"fmt"
	"testing"
)

func TestValidateAssembly(t *testing.T) {
	// Test valid assembly
	validAssembly := `
.text
.globl main
main:
	pushq %rbp
	movq %rsp, %rbp
	movq %rdi, %rax
	popq %rbp
	ret
`
	result := ValidateAssembly(validAssembly)
	if !result.Valid {
		t.Errorf("Expected valid assembly, got errors: %v", result.Errors)
	}
	fmt.Println("✅ Valid assembly test passed")

	// Test invalid instruction
	invalidAssembly := `
.text
main:
	invalid_instruction %rax, %rbx
	ret
`
	result2 := ValidateAssembly(invalidAssembly)
	if result2.Valid {
		t.Errorf("Expected invalid assembly to fail validation")
	}
	fmt.Println("✅ Invalid instruction test passed")

	// Test duplicate label
	duplicateLabel := `
.text
main:
	movq %rax, %rbx
main:
	ret
`
	result3 := ValidateAssembly(duplicateLabel)
	if result3.Valid {
		t.Errorf("Expected duplicate label to fail validation")
	}
	fmt.Println("✅ Duplicate label test passed")
}

func TestValidateInstructions(t *testing.T) {
	lines := []string{
		"main:",
		"  movq %rax, %rbx",
		"  ret",
	}
	errors := ValidateInstructions(lines)
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got: %v", errors)
	}
	fmt.Println("✅ ValidateInstructions test passed")
}

func TestValidateLabels(t *testing.T) {
	lines := []string{
		"main:",
		"  jmp end",
		"end:",
		"  ret",
	}
	errors := ValidateLabels(lines)
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got: %v", errors)
	}
	fmt.Println("✅ ValidateLabels test passed")
}

func TestValidateRegisters(t *testing.T) {
	lines := []string{
		"main:",
		"  pushq %rbx",
		"  movq %rax, %rbx",
		"  popq %rbx",
		"  ret",
	}
	errors := ValidateRegisters(lines)
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got: %v", errors)
	}
	fmt.Println("✅ ValidateRegisters test passed")
}

func TestHelperFunctions(t *testing.T) {
	if !IsRegister("rax") {
		t.Error("IsRegister should return true for rax")
	}
	if !IsCalleeSaved("rbx") {
		t.Error("IsCalleeSaved should return true for rbx")
	}
	if !IsCallerSaved("rax") {
		t.Error("IsCallerSaved should return true for rax")
	}
	args := GetArgumentRegisters()
	if len(args) != 6 {
		t.Errorf("Expected 6 argument registers, got %d", len(args))
	}
	fmt.Println("✅ Helper functions test passed")
}
