package codegen

import (
	"testing"
)

// TestReg_String tests the String() method for all registers.
func TestReg_String(t *testing.T) {
	tests := []struct {
		reg  Reg
		want string
	}{
		// General purpose registers
		{RAX, "rax"},
		{RBX, "rbx"},
		{RCX, "rcx"},
		{RDX, "rdx"},
		{RSI, "rsi"},
		{RDI, "rdi"},
		{RBP, "rbp"},
		{RSP, "rsp"},
		{R8, "r8"},
		{R9, "r9"},
		{R10, "r10"},
		{R11, "r11"},
		{R12, "r12"},
		{R13, "r13"},
		{R14, "r14"},
		{R15, "r15"},
		// Floating point registers
		{XMM0, "xmm0"},
		{XMM1, "xmm1"},
		{XMM2, "xmm2"},
		{XMM3, "xmm3"},
		{XMM4, "xmm4"},
		{XMM5, "xmm5"},
		{XMM6, "xmm6"},
		{XMM7, "xmm7"},
		{XMM8, "xmm8"},
		{XMM9, "xmm9"},
		{XMM10, "xmm10"},
		{XMM11, "xmm11"},
		{XMM12, "xmm12"},
		{XMM13, "xmm13"},
		{XMM14, "xmm14"},
		{XMM15, "xmm15"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.reg.String(); got != tt.want {
				t.Errorf("Reg(%d).String() = %q, want %q", tt.reg, got, tt.want)
			}
		})
	}
}

// TestReg_Size tests the Size() method for all registers.
func TestReg_Size(t *testing.T) {
	tests := []struct {
		reg  Reg
		want int
	}{
		// General purpose registers: 8 bytes
		{RAX, 8},
		{RBX, 8},
		{RCX, 8},
		{RDX, 8},
		{RSI, 8},
		{RDI, 8},
		{RBP, 8},
		{RSP, 8},
		{R8, 8},
		{R9, 8},
		{R10, 8},
		{R11, 8},
		{R12, 8},
		{R13, 8},
		{R14, 8},
		{R15, 8},
		// Floating point registers: 16 bytes
		{XMM0, 16},
		{XMM1, 16},
		{XMM2, 16},
		{XMM3, 16},
		{XMM4, 16},
		{XMM5, 16},
		{XMM6, 16},
		{XMM7, 16},
		{XMM8, 16},
		{XMM9, 16},
		{XMM10, 16},
		{XMM11, 16},
		{XMM12, 16},
		{XMM13, 16},
		{XMM14, 16},
		{XMM15, 16},
	}

	for _, tt := range tests {
		t.Run(tt.reg.String(), func(t *testing.T) {
			if got := tt.reg.Size(); got != tt.want {
				t.Errorf("Reg(%s).Size() = %d, want %d", tt.reg.String(), got, tt.want)
			}
		})
	}
}

// TestReg_String_Unknown tests the String() method for invalid register values.
func TestReg_String_Unknown(t *testing.T) {
	// Test with an invalid register value
	invalidReg := Reg(999)
	if got := invalidReg.String(); got != "unknown" {
		t.Errorf("Reg(999).String() = %q, want %q", got, "unknown")
	}
}

// TestReg_Size_Unknown tests the Size() method for invalid register values.
func TestReg_Size_Unknown(t *testing.T) {
	// Test with an invalid register value
	invalidReg := Reg(999)
	if got := invalidReg.Size(); got != 0 {
		t.Errorf("Reg(999).Size() = %d, want %d", got, 0)
	}
}

// TestReg_GeneralPurposeRegisters tests that all general purpose registers have correct size.
func TestReg_GeneralPurposeRegisters(t *testing.T) {
	generalPurposeRegs := []Reg{RAX, RBX, RCX, RDX, RSI, RDI, RBP, RSP, R8, R9, R10, R11, R12, R13, R14, R15}

	for _, reg := range generalPurposeRegs {
		if size := reg.Size(); size != 8 {
			t.Errorf("General purpose register %s has size %d, want 8", reg.String(), size)
		}
	}
}

// TestReg_XMMRegisters tests that all XMM registers have correct size.
func TestReg_XMMRegisters(t *testing.T) {
	xmmRegs := []Reg{XMM0, XMM1, XMM2, XMM3, XMM4, XMM5, XMM6, XMM7, XMM8, XMM9, XMM10, XMM11, XMM12, XMM13, XMM14, XMM15}

	for _, reg := range xmmRegs {
		if size := reg.Size(); size != 16 {
			t.Errorf("XMM register %s has size %d, want 16", reg.String(), size)
		}
	}
}