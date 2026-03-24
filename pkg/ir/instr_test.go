package ir

import (
	"testing"

	"github.com/akzj/goc/pkg/parser"
)

func TestOpcodeString(t *testing.T) {
	tests := []struct {
		opcode Opcode
		want   string
	}{
		{OpAdd, "add"},
		{OpSub, "sub"},
		{OpMul, "mul"},
		{OpDiv, "div"},
		{OpMod, "mod"},
		{OpNeg, "neg"},
		{OpBitNot, "bitnot"},
		{OpBitAnd, "and"},
		{OpBitOr, "or"},
		{OpBitXor, "xor"},
		{OpShl, "shl"},
		{OpShr, "shr"},
		{OpEq, "eq"},
		{OpNe, "ne"},
		{OpLt, "lt"},
		{OpLe, "le"},
		{OpGt, "gt"},
		{OpGe, "ge"},
		{OpAnd, "and"},
		{OpOr, "or"},
		{OpNot, "not"},
		{OpLoad, "load"},
		{OpStore, "store"},
		{OpLea, "lea"},
		{OpAlloc, "alloc"},
		{OpFree, "free"},
		{OpJmp, "jmp"},
		{OpJmpIf, "jmpif"},
		{OpJmpUnless, "jmpunless"},
		{OpCall, "call"},
		{OpRet, "ret"},
		{OpLabel, "label"},
		{OpCast, "cast"},
		{OpZeroExt, "zext"},
		{OpSignExt, "sext"},
		{OpTrunc, "trunc"},
		{OpPhi, "phi"},
		{OpNop, "nop"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.opcode.String()
			if got != tt.want {
				t.Errorf("Opcode.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOperandKindString(t *testing.T) {
	tests := []struct {
		kind OperandKind
		want string
	}{
		{OperandTemp, "temp"},
		{OperandParam, "param"},
		{OperandGlobal, "global"},
		{OperandConst, "const"},
		{OperandLabel, "label"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.kind.String()
			if got != tt.want {
				t.Errorf("OperandKind.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOperandString(t *testing.T) {
	tests := []struct {
		name string
		op   *Operand
		want string
	}{
		{
			name: "temp operand",
			op: &Operand{
				Kind: OperandTemp,
				Value: &Temp{
					ID:   1,
					Type: &parser.BaseType{Kind: parser.TypeInt},
				},
			},
			want: "t1",
		},
		{
			name: "param operand",
			op: &Operand{
				Kind: OperandParam,
				Value: &Temp{
					ID:   2,
					Type: &parser.BaseType{Kind: parser.TypeInt},
				},
			},
			want: "p2",
		},
		{
			name: "global operand",
			op: &Operand{
				Kind:  OperandGlobal,
				Value: "global_var",
			},
			want: "global(global_var)",
		},
		{
			name: "const operand",
			op: &Operand{
				Kind:  OperandConst,
				Value: int64(42),
			},
			want: "42",
		},
		{
			name: "label operand",
			op: &Operand{
				Kind:  OperandLabel,
				Value: "L1",
			},
			want: "label(L1)",
		},
		{
			name: "nil operand",
			op:   nil,
			want: "nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.op.String()
			if got != tt.want {
				t.Errorf("Operand.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBinaryInstr(t *testing.T) {
	dest := &Operand{Kind: OperandTemp, Value: &Temp{ID: 1}}
	left := &Operand{Kind: OperandTemp, Value: &Temp{ID: 2}}
	right := &Operand{Kind: OperandTemp, Value: &Temp{ID: 3}}

	instr := NewBinaryInstr(OpAdd, dest, left, right)

	if instr.Opcode() != OpAdd {
		t.Errorf("BinaryInstr.Opcode() = %v, want %v", instr.Opcode(), OpAdd)
	}

	if instr.Dest() != dest {
		t.Errorf("BinaryInstr.Dest() = %v, want %v", instr.Dest(), dest)
	}

	ops := instr.Operands()
	if len(ops) != 2 {
		t.Errorf("BinaryInstr.Operands() = %d, want 2", len(ops))
	}

	expected := "add t1, t2, t3"
	if instr.String() != expected {
		t.Errorf("BinaryInstr.String() = %q, want %q", instr.String(), expected)
	}
}

func TestUnaryInstr(t *testing.T) {
	dest := &Operand{Kind: OperandTemp, Value: &Temp{ID: 1}}
	operand := &Operand{Kind: OperandTemp, Value: &Temp{ID: 2}}

	instr := NewUnaryInstr(OpNeg, dest, operand)

	if instr.Opcode() != OpNeg {
		t.Errorf("UnaryInstr.Opcode() = %v, want %v", instr.Opcode(), OpNeg)
	}

	if instr.Dest() != dest {
		t.Errorf("UnaryInstr.Dest() = %v, want %v", instr.Dest(), dest)
	}

	ops := instr.Operands()
	if len(ops) != 1 {
		t.Errorf("UnaryInstr.Operands() = %d, want 1", len(ops))
	}

	expected := "neg t1, t2"
	if instr.String() != expected {
		t.Errorf("UnaryInstr.String() = %q, want %q", instr.String(), expected)
	}
}

func TestLoadInstr(t *testing.T) {
	dest := &Operand{Kind: OperandTemp, Value: &Temp{ID: 1}}
	addr := &Operand{Kind: OperandGlobal, Value: "x"}

	instr := NewLoadInstr(dest, addr)

	if instr.Opcode() != OpLoad {
		t.Errorf("LoadInstr.Opcode() = %v, want %v", instr.Opcode(), OpLoad)
	}

	if instr.Dest() != dest {
		t.Errorf("LoadInstr.Dest() = %v, want %v", instr.Dest(), dest)
	}

	ops := instr.Operands()
	if len(ops) != 1 {
		t.Errorf("LoadInstr.Operands() = %d, want 1", len(ops))
	}

	expected := "load t1, global(x)"
	if instr.String() != expected {
		t.Errorf("LoadInstr.String() = %q, want %q", instr.String(), expected)
	}
}

func TestStoreInstr(t *testing.T) {
	value := &Operand{Kind: OperandTemp, Value: &Temp{ID: 1}}
	addr := &Operand{Kind: OperandGlobal, Value: "x"}

	instr := NewStoreInstr(value, addr)

	if instr.Opcode() != OpStore {
		t.Errorf("StoreInstr.Opcode() = %v, want %v", instr.Opcode(), OpStore)
	}

	if instr.Dest() != nil {
		t.Errorf("StoreInstr.Dest() = %v, want nil", instr.Dest())
	}

	ops := instr.Operands()
	if len(ops) != 2 {
		t.Errorf("StoreInstr.Operands() = %d, want 2", len(ops))
	}

	expected := "store t1, global(x)"
	if instr.String() != expected {
		t.Errorf("StoreInstr.String() = %q, want %q", instr.String(), expected)
	}
}

func TestJmpInstr(t *testing.T) {
	target := &Operand{Kind: OperandLabel, Value: "L1"}

	instr := NewJmpInstr(target)

	if instr.Opcode() != OpJmp {
		t.Errorf("JmpInstr.Opcode() = %v, want %v", instr.Opcode(), OpJmp)
	}

	if instr.Dest() != nil {
		t.Errorf("JmpInstr.Dest() = %v, want nil", instr.Dest())
	}

	ops := instr.Operands()
	if len(ops) != 1 {
		t.Errorf("JmpInstr.Operands() = %d, want 1", len(ops))
	}

	expected := "jmp label(L1)"
	if instr.String() != expected {
		t.Errorf("JmpInstr.String() = %q, want %q", instr.String(), expected)
	}
}

func TestCondJmpInstr(t *testing.T) {
	cond := &Operand{Kind: OperandTemp, Value: &Temp{ID: 1}}
	target := &Operand{Kind: OperandLabel, Value: "L1"}

	instr := NewCondJmpInstr(OpJmpIf, cond, target)

	if instr.Opcode() != OpJmpIf {
		t.Errorf("CondJmpInstr.Opcode() = %v, want %v", instr.Opcode(), OpJmpIf)
	}

	if instr.Dest() != nil {
		t.Errorf("CondJmpInstr.Dest() = %v, want nil", instr.Dest())
	}

	ops := instr.Operands()
	if len(ops) != 2 {
		t.Errorf("CondJmpInstr.Operands() = %d, want 2", len(ops))
	}

	expected := "jmpif t1, label(L1)"
	if instr.String() != expected {
		t.Errorf("CondJmpInstr.String() = %q, want %q", instr.String(), expected)
	}
}

func TestCallInstr(t *testing.T) {
	dest := &Operand{Kind: OperandTemp, Value: &Temp{ID: 1}}
	funcOp := &Operand{Kind: OperandGlobal, Value: "foo"}
	args := []*Operand{
		{Kind: OperandTemp, Value: &Temp{ID: 2}},
		{Kind: OperandTemp, Value: &Temp{ID: 3}},
	}

	instr := NewCallInstr(dest, funcOp, args)

	if instr.Opcode() != OpCall {
		t.Errorf("CallInstr.Opcode() = %v, want %v", instr.Opcode(), OpCall)
	}

	if instr.Dest() != dest {
		t.Errorf("CallInstr.Dest() = %v, want %v", instr.Dest(), dest)
	}

	ops := instr.Operands()
	if len(ops) != 3 {
		t.Errorf("CallInstr.Operands() = %d, want 3", len(ops))
	}

	expected := "call t1 = global(foo)(t2, t3)"
	if instr.String() != expected {
		t.Errorf("CallInstr.String() = %q, want %q", instr.String(), expected)
	}
}

func TestCallInstrNoDest(t *testing.T) {
	funcOp := &Operand{Kind: OperandGlobal, Value: "foo"}
	args := []*Operand{
		{Kind: OperandTemp, Value: &Temp{ID: 1}},
	}

	instr := NewCallInstr(nil, funcOp, args)

	if instr.Opcode() != OpCall {
		t.Errorf("CallInstr.Opcode() = %v, want %v", instr.Opcode(), OpCall)
	}

	if instr.Dest() != nil {
		t.Errorf("CallInstr.Dest() = %v, want nil", instr.Dest())
	}

	expected := "call global(foo)(t1)"
	if instr.String() != expected {
		t.Errorf("CallInstr.String() = %q, want %q", instr.String(), expected)
	}
}

func TestRetInstr(t *testing.T) {
	value := &Operand{Kind: OperandTemp, Value: &Temp{ID: 1}}

	instr := NewRetInstr(value)

	if instr.Opcode() != OpRet {
		t.Errorf("RetInstr.Opcode() = %v, want %v", instr.Opcode(), OpRet)
	}

	if instr.Dest() != nil {
		t.Errorf("RetInstr.Dest() = %v, want nil", instr.Dest())
	}

	ops := instr.Operands()
	if len(ops) != 1 {
		t.Errorf("RetInstr.Operands() = %d, want 1", len(ops))
	}

	expected := "ret t1"
	if instr.String() != expected {
		t.Errorf("RetInstr.String() = %q, want %q", instr.String(), expected)
	}
}

func TestRetInstrNoValue(t *testing.T) {
	instr := NewRetInstr(nil)

	if instr.Opcode() != OpRet {
		t.Errorf("RetInstr.Opcode() = %v, want %v", instr.Opcode(), OpRet)
	}

	ops := instr.Operands()
	if len(ops) != 0 {
		t.Errorf("RetInstr.Operands() = %d, want 0", len(ops))
	}

	expected := "ret"
	if instr.String() != expected {
		t.Errorf("RetInstr.String() = %q, want %q", instr.String(), expected)
	}
}

func TestLabelInstr(t *testing.T) {
	label := &Operand{Kind: OperandLabel, Value: "L1"}

	instr := NewLabelInstr(label)

	if instr.Opcode() != OpLabel {
		t.Errorf("LabelInstr.Opcode() = %v, want %v", instr.Opcode(), OpLabel)
	}

	if instr.Dest() != nil {
		t.Errorf("LabelInstr.Dest() = %v, want nil", instr.Dest())
	}

	ops := instr.Operands()
	if len(ops) != 1 {
		t.Errorf("LabelInstr.Operands() = %d, want 1", len(ops))
	}

	expected := "label(L1):"
	if instr.String() != expected {
		t.Errorf("LabelInstr.String() = %q, want %q", instr.String(), expected)
	}
}

func TestCastInstr(t *testing.T) {
	dest := &Operand{Kind: OperandTemp, Value: &Temp{ID: 1}}
	operand := &Operand{Kind: OperandTemp, Value: &Temp{ID: 2}}

	instr := NewCastInstr(OpCast, dest, operand)

	if instr.Opcode() != OpCast {
		t.Errorf("CastInstr.Opcode() = %v, want %v", instr.Opcode(), OpCast)
	}

	if instr.Dest() != dest {
		t.Errorf("CastInstr.Dest() = %v, want %v", instr.Dest(), dest)
	}

	ops := instr.Operands()
	if len(ops) != 1 {
		t.Errorf("CastInstr.Operands() = %d, want 1", len(ops))
	}

	expected := "cast t1, t2"
	if instr.String() != expected {
		t.Errorf("CastInstr.String() = %q, want %q", instr.String(), expected)
	}
}

func TestLeaInstr(t *testing.T) {
	dest := &Operand{Kind: OperandTemp, Value: &Temp{ID: 1}}
	addr := &Operand{Kind: OperandGlobal, Value: "x"}

	instr := NewLeaInstr(dest, addr)

	if instr.Opcode() != OpLea {
		t.Errorf("LeaInstr.Opcode() = %v, want %v", instr.Opcode(), OpLea)
	}

	if instr.Dest() != dest {
		t.Errorf("LeaInstr.Dest() = %v, want %v", instr.Dest(), dest)
	}

	ops := instr.Operands()
	if len(ops) != 1 {
		t.Errorf("LeaInstr.Operands() = %d, want 1", len(ops))
	}

	expected := "lea t1, global(x)"
	if instr.String() != expected {
		t.Errorf("LeaInstr.String() = %q, want %q", instr.String(), expected)
	}
}

func TestAllocInstr(t *testing.T) {
	dest := &Operand{Kind: OperandTemp, Value: &Temp{ID: 1}}
	size := &Operand{Kind: OperandConst, Value: int64(8)}

	instr := NewAllocInstr(dest, size)

	if instr.Opcode() != OpAlloc {
		t.Errorf("AllocInstr.Opcode() = %v, want %v", instr.Opcode(), OpAlloc)
	}

	if instr.Dest() != dest {
		t.Errorf("AllocInstr.Dest() = %v, want %v", instr.Dest(), dest)
	}

	ops := instr.Operands()
	if len(ops) != 1 {
		t.Errorf("AllocInstr.Operands() = %d, want 1", len(ops))
	}

	expected := "alloc t1, 8"
	if instr.String() != expected {
		t.Errorf("AllocInstr.String() = %q, want %q", instr.String(), expected)
	}
}

func TestNopInstr(t *testing.T) {
	instr := NewNopInstr()

	if instr.Opcode() != OpNop {
		t.Errorf("NopInstr.Opcode() = %v, want %v", instr.Opcode(), OpNop)
	}

	if instr.Dest() != nil {
		t.Errorf("NopInstr.Dest() = %v, want nil", instr.Dest())
	}

	ops := instr.Operands()
	if len(ops) != 0 {
		t.Errorf("NopInstr.Operands() = %d, want 0", len(ops))
	}

	expected := "nop"
	if instr.String() != expected {
		t.Errorf("NopInstr.String() = %q, want %q", instr.String(), expected)
	}
}

func TestPhiInstr(t *testing.T) {
	dest := &Operand{Kind: OperandTemp, Value: &Temp{ID: 1}}
	values := []*Operand{
		{Kind: OperandTemp, Value: &Temp{ID: 2}},
		{Kind: OperandTemp, Value: &Temp{ID: 3}},
	}
	labels := []*Operand{
		{Kind: OperandLabel, Value: "L1"},
		{Kind: OperandLabel, Value: "L2"},
	}

	instr := NewPhiInstr(dest, values, labels)

	if instr.Opcode() != OpPhi {
		t.Errorf("PhiInstr.Opcode() = %v, want %v", instr.Opcode(), OpPhi)
	}

	if instr.Dest() != dest {
		t.Errorf("PhiInstr.Dest() = %v, want %v", instr.Dest(), dest)
	}

	ops := instr.Operands()
	if len(ops) != 4 {
		t.Errorf("PhiInstr.Operands() = %d, want 4", len(ops))
	}

	expected := "phi t1 = [t2, label(L1)], [t3, label(L2)]"
	if instr.String() != expected {
		t.Errorf("PhiInstr.String() = %q, want %q", instr.String(), expected)
	}
}