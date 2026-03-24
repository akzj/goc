package ir

import (
	"testing"

	"github.com/akzj/goc/pkg/parser"
)

func TestIRString(t *testing.T) {
	tests := []struct {
		name string
		ir   *IR
		want string
	}{
		{
			name: "empty IR",
			ir: &IR{
				Functions: []*Function{},
				Globals:   []*GlobalVar{},
				Constants: []*Constant{},
			},
			want: "IR {\n}",
		},
		{
			name: "IR with function",
			ir: &IR{
				Functions: []*Function{
					{
						Name:       "main",
						ReturnType: &parser.BaseType{Kind: parser.TypeInt},
						Params:     []*Param{},
						Blocks:     []*BasicBlock{},
						LocalVars:  []*LocalVar{},
					},
				},
				Globals:   []*GlobalVar{},
				Constants: []*Constant{},
			},
			want: "IR {\n  Functions:\n    Function main() -> unsigned int {\n  }\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ir.String()
			if got != tt.want {
				t.Errorf("IR.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFunctionString(t *testing.T) {
	tests := []struct {
		name string
		fn   *Function
		want string
	}{
		{
			name: "function without return type",
			fn: &Function{
				Name:       "foo",
				ReturnType: nil,
				Params:     []*Param{},
				Blocks:     []*BasicBlock{},
				LocalVars:  []*LocalVar{},
			},
			want: "Function foo() {\n  }",
		},
		{
			name: "function with params",
			fn: &Function{
				Name:       "bar",
				ReturnType: &parser.BaseType{Kind: parser.TypeInt},
				Params: []*Param{
					{Name: "x", Type: &parser.BaseType{Kind: parser.TypeInt}},
					{Name: "y", Type: &parser.BaseType{Kind: parser.TypeInt}},
				},
				Blocks:    []*BasicBlock{},
				LocalVars: []*LocalVar{},
			},
			want: "Function bar(unsigned int x, unsigned int y) -> unsigned int {\n  }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn.String()
			if got != tt.want {
				t.Errorf("Function.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBasicBlockString(t *testing.T) {
	tests := []struct {
		name string
		bb   *BasicBlock
		want string
	}{
		{
			name: "empty block",
			bb: &BasicBlock{
				Label:  "entry",
				Instrs: []Instruction{},
				Preds:  []*BasicBlock{},
				Succs:  []*BasicBlock{},
			},
			want: "Block entry:",
		},
		{
			name: "block with instructions",
			bb: &BasicBlock{
				Label: "L1",
				Instrs: []Instruction{
					NewNopInstr(),
				},
				Preds: []*BasicBlock{},
				Succs: []*BasicBlock{},
			},
			want: "Block L1:\n      nop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bb.String()
			if got != tt.want {
				t.Errorf("BasicBlock.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParamString(t *testing.T) {
	tests := []struct {
		name string
		p    *Param
		want string
	}{
		{
			name: "param with type",
			p: &Param{
				Name: "x",
				Type: &parser.BaseType{Kind: parser.TypeInt},
			},
			want: "unsigned int x",
		},
		{
			name: "param without type",
			p: &Param{
				Name: "y",
				Type: nil,
			},
			want: "y",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.String()
			if got != tt.want {
				t.Errorf("Param.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLocalVarString(t *testing.T) {
	tests := []struct {
		name string
		lv   *LocalVar
		want string
	}{
		{
			name: "local var with type",
			lv: &LocalVar{
				Name:        "x",
				Type:        &parser.BaseType{Kind: parser.TypeInt},
				StackOffset: 8,
			},
			want: "unsigned int x (offset: 8)",
		},
		{
			name: "local var without type",
			lv: &LocalVar{
				Name:        "y",
				Type:        nil,
				StackOffset: 16,
			},
			want: "y (offset: 16)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.lv.String()
			if got != tt.want {
				t.Errorf("LocalVar.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGlobalVarString(t *testing.T) {
	tests := []struct {
		name string
		gv   *GlobalVar
		want string
	}{
		{
			name: "global var with init",
			gv: &GlobalVar{
				Name: "x",
				Type: &parser.BaseType{Kind: parser.TypeInt},
				Init: &parser.IntLiteral{Value: 42},
			},
			want: "global unsigned int x = IntLiteral{value=42, raw=, suffix=}",
		},
		{
			name: "global var without init",
			gv: &GlobalVar{
				Name: "y",
				Type: &parser.BaseType{Kind: parser.TypeInt},
				Init: nil,
			},
			want: "global unsigned int y",
		},
		{
			name: "global var without type",
			gv: &GlobalVar{
				Name: "z",
				Type: nil,
				Init: nil,
			},
			want: "global z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.gv.String()
			if got != tt.want {
				t.Errorf("GlobalVar.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConstantString(t *testing.T) {
	tests := []struct {
		name string
		c    *Constant
		want string
	}{
		{
			name: "constant",
			c: &Constant{
				Name:  "PI",
				Value: 3.14,
			},
			want: "const PI = 3.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.String()
			if got != tt.want {
				t.Errorf("Constant.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTempString(t *testing.T) {
	tests := []struct {
		name string
		t    *Temp
		want string
	}{
		{
			name: "temp",
			t: &Temp{
				ID:   42,
				Type: &parser.BaseType{Kind: parser.TypeInt},
			},
			want: "t42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.t.String()
			if got != tt.want {
				t.Errorf("Temp.String() = %q, want %q", got, tt.want)
			}
		})
	}
}