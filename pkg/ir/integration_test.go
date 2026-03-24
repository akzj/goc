package ir

import (
	"testing"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/parser"
)

// TestIRGeneratorIntegration tests end-to-end IR generation from AST.
func TestIRGeneratorIntegration(t *testing.T) {
	tests := []struct {
		name        string
		ast         *parser.TranslationUnit
		wantFuncs   int
		wantGlobals int
		wantErr     bool
	}{
		{
			name: "simple function",
			ast: &parser.TranslationUnit{
				Declarations: []parser.Declaration{
					&parser.FunctionDecl{
						Name: "main",
						Type: &parser.FuncType{
							Return: &parser.BaseType{Kind: parser.TypeInt},
						},
						Params: []*parser.ParamDecl{},
						Body: &parser.CompoundStmt{
							Declarations: []parser.Declaration{},
							Statements: []parser.Statement{
								&parser.ReturnStmt{
									Value: &parser.IntLiteral{Value: 0},
								},
							},
						},
					},
				},
			},
			wantFuncs:   1,
			wantGlobals: 0,
			wantErr:     false,
		},
		{
			name: "function with parameters",
			ast: &parser.TranslationUnit{
				Declarations: []parser.Declaration{
					&parser.FunctionDecl{
						Name: "add",
						Type: &parser.FuncType{
							Return: &parser.BaseType{Kind: parser.TypeInt},
						},
						Params: []*parser.ParamDecl{
							{Name: "x", Type: &parser.BaseType{Kind: parser.TypeInt}},
							{Name: "y", Type: &parser.BaseType{Kind: parser.TypeInt}},
						},
						Body: &parser.CompoundStmt{
							Declarations: []parser.Declaration{},
							Statements: []parser.Statement{
								&parser.ReturnStmt{
									Value: &parser.BinaryExpr{
										Op:   lexer.ADD, // PLUS
										Left: &parser.IdentExpr{Name: "x"},
										Right: &parser.IdentExpr{Name: "y"},
									},
								},
							},
						},
					},
				},
			},
			wantFuncs:   1,
			wantGlobals: 0,
			wantErr:     false,
		},
		{
			name: "function with local variable",
			ast: &parser.TranslationUnit{
				Declarations: []parser.Declaration{
					&parser.FunctionDecl{
						Name: "test",
						Type: &parser.FuncType{
							Return: &parser.BaseType{Kind: parser.TypeInt},
						},
						Params: []*parser.ParamDecl{},
						Body: &parser.CompoundStmt{
							Declarations: []parser.Declaration{
								&parser.VarDecl{
									Name: "x",
									Type: &parser.BaseType{Kind: parser.TypeInt},
									Init: &parser.IntLiteral{Value: 10},
								},
							},
							Statements: []parser.Statement{
								&parser.ReturnStmt{
									Value: &parser.IdentExpr{Name: "x"},
								},
							},
						},
					},
				},
			},
			wantFuncs:   1,
			wantGlobals: 0,
			wantErr:     false,
		},
		{
			name: "function with if statement",
			ast: &parser.TranslationUnit{
				Declarations: []parser.Declaration{
					&parser.FunctionDecl{
						Name: "max",
						Type: &parser.FuncType{
							Return: &parser.BaseType{Kind: parser.TypeInt},
						},
						Params: []*parser.ParamDecl{
							{Name: "a", Type: &parser.BaseType{Kind: parser.TypeInt}},
							{Name: "b", Type: &parser.BaseType{Kind: parser.TypeInt}},
						},
						Body: &parser.CompoundStmt{
							Declarations: []parser.Declaration{},
							Statements: []parser.Statement{
								&parser.IfStmt{
									Cond: &parser.BinaryExpr{
										Op:   lexer.ADD, // LT
										Left: &parser.IdentExpr{Name: "a"},
										Right: &parser.IdentExpr{Name: "b"},
									},
									Then: &parser.ReturnStmt{
										Value: &parser.IdentExpr{Name: "b"},
									},
									Else: &parser.ReturnStmt{
										Value: &parser.IdentExpr{Name: "a"},
									},
								},
							},
						},
					},
				},
			},
			wantFuncs:   1,
			wantGlobals: 0,
			wantErr:     false,
		},
		{
			name: "function with while loop",
			ast: &parser.TranslationUnit{
				Declarations: []parser.Declaration{
					&parser.FunctionDecl{
						Name: "sum",
						Type: &parser.FuncType{
							Return: &parser.BaseType{Kind: parser.TypeInt},
						},
						Params: []*parser.ParamDecl{
							{Name: "n", Type: &parser.BaseType{Kind: parser.TypeInt}},
						},
						Body: &parser.CompoundStmt{
							Declarations: []parser.Declaration{
								&parser.VarDecl{
									Name: "result",
									Type: &parser.BaseType{Kind: parser.TypeInt},
									Init: &parser.IntLiteral{Value: 0},
								},
							},
							Statements: []parser.Statement{
								&parser.WhileStmt{
									Cond: &parser.BinaryExpr{
										Op:   lexer.ADD, // LT
										Left: &parser.IdentExpr{Name: "result"},
										Right: &parser.IdentExpr{Name: "n"},
									},
									Body: &parser.CompoundStmt{
										Declarations: []parser.Declaration{},
										Statements: []parser.Statement{
											&parser.ExprStmt{
												Expr: &parser.AssignExpr{
													Op:    lexer.ASSIGN, // ASSIGN
													Left:  &parser.IdentExpr{Name: "result"},
													Right: &parser.BinaryExpr{
														Op:   lexer.ADD, // PLUS
														Left: &parser.IdentExpr{Name: "result"},
														Right: &parser.IntLiteral{Value: 1},
													},
												},
											},
										},
									},
								},
								&parser.ReturnStmt{
									Value: &parser.IdentExpr{Name: "result"},
								},
							},
						},
					},
				},
			},
			wantFuncs:   1,
			wantGlobals: 0,
			wantErr:     false,
		},
		{
			name: "global variable",
			ast: &parser.TranslationUnit{
				Declarations: []parser.Declaration{
					&parser.VarDecl{
						Name: "global_count",
						Type: &parser.BaseType{Kind: parser.TypeInt},
						Init: &parser.IntLiteral{Value: 100},
					},
					&parser.FunctionDecl{
						Name: "get_count",
						Type: &parser.FuncType{
							Return: &parser.BaseType{Kind: parser.TypeInt},
						},
						Params: []*parser.ParamDecl{},
						Body: &parser.CompoundStmt{
							Declarations: []parser.Declaration{},
							Statements: []parser.Statement{
								&parser.ReturnStmt{
									Value: &parser.IdentExpr{Name: "global_count"},
								},
							},
						},
					},
				},
			},
			wantFuncs:   1,
			wantGlobals: 1,
			wantErr:     false,
		},
		{
			name: "multiple functions",
			ast: &parser.TranslationUnit{
				Declarations: []parser.Declaration{
					&parser.FunctionDecl{
						Name: "func1",
						Type: &parser.FuncType{
							Return: &parser.BaseType{Kind: parser.TypeInt},
						},
						Params: []*parser.ParamDecl{},
						Body: &parser.CompoundStmt{
							Declarations: []parser.Declaration{},
							Statements: []parser.Statement{
								&parser.ReturnStmt{
									Value: &parser.IntLiteral{Value: 1},
								},
							},
						},
					},
					&parser.FunctionDecl{
						Name: "func2",
						Type: &parser.FuncType{
							Return: &parser.BaseType{Kind: parser.TypeInt},
						},
						Params: []*parser.ParamDecl{},
						Body: &parser.CompoundStmt{
							Declarations: []parser.Declaration{},
							Statements: []parser.Statement{
								&parser.ReturnStmt{
									Value: &parser.IntLiteral{Value: 2},
								},
							},
						},
					},
				},
			},
			wantFuncs:   2,
			wantGlobals: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorHandler := errhand.NewErrorHandler()
			g := NewIRGenerator(errorHandler)

			ir, err := g.Generate(tt.ast)

			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if ir == nil {
				t.Fatal("Generate() returned nil IR")
			}

			if len(ir.Functions) != tt.wantFuncs {
				t.Errorf("Generate() created %d functions, want %d", len(ir.Functions), tt.wantFuncs)
			}

			if len(ir.Globals) != tt.wantGlobals {
				t.Errorf("Generate() created %d globals, want %d", len(ir.Globals), tt.wantGlobals)
			}

			// Verify function structure
			for _, fn := range ir.Functions {
				if fn.Name == "" {
					t.Error("Function has empty name")
				}
				if len(fn.Blocks) == 0 {
					t.Errorf("Function %s has no basic blocks", fn.Name)
				}
			}
		})
	}
}

// TestIRGeneratorBasicBlocks tests basic block generation.
func TestIRGeneratorBasicBlocks(t *testing.T) {
	g := NewIRGenerator(nil)

	// Create a function with if-else
	fnDecl := &parser.FunctionDecl{
		Name: "test",
		Type: &parser.FuncType{
			Return: &parser.BaseType{Kind: parser.TypeInt},
		},
		Params: []*parser.ParamDecl{},
		Body: &parser.CompoundStmt{
			Declarations: []parser.Declaration{},
			Statements: []parser.Statement{
				&parser.IfStmt{
					Cond: &parser.IntLiteral{Value: 1},
					Then: &parser.ReturnStmt{
						Value: &parser.IntLiteral{Value: 1},
					},
					Else: &parser.ReturnStmt{
						Value: &parser.IntLiteral{Value: 0},
					},
				},
			},
		},
	}

	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{fnDecl},
	}

	ir, err := g.Generate(ast)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(ir.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(ir.Functions))
	}

	fn := ir.Functions[0]

	// Should have at least 3 blocks: entry, then, else, end
	if len(fn.Blocks) < 3 {
		t.Errorf("Expected at least 3 basic blocks, got %d", len(fn.Blocks))
	}

	// Verify block labels are unique
	labels := make(map[string]bool)
	for _, block := range fn.Blocks {
		if labels[block.Label] {
			t.Errorf("Duplicate block label: %s", block.Label)
		}
		labels[block.Label] = true
	}
}

// TestIRGeneratorControlFlow tests control flow generation.
func TestIRGeneratorControlFlow(t *testing.T) {
	tests := []struct {
		name     string
		stmt     parser.Statement
		wantJmps int
	}{
		{
			name: "if without else",
			stmt: &parser.IfStmt{
				Cond: &parser.IntLiteral{Value: 1},
				Then: &parser.ReturnStmt{Value: &parser.IntLiteral{Value: 1}},
				Else: nil,
			},
			wantJmps: 1,
		},
		{
			name: "if with else",
			stmt: &parser.IfStmt{
				Cond: &parser.IntLiteral{Value: 1},
				Then: &parser.ReturnStmt{Value: &parser.IntLiteral{Value: 1}},
				Else: &parser.ReturnStmt{Value: &parser.IntLiteral{Value: 0}},
			},
			// Both branches end with return, so only one conditional jump needed
			wantJmps: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewIRGenerator(nil)

			fnDecl := &parser.FunctionDecl{
				Name:   "test",
				Type:   &parser.FuncType{Return: &parser.BaseType{Kind: parser.TypeInt}},
				Params: []*parser.ParamDecl{},
				Body: &parser.CompoundStmt{
					Declarations: []parser.Declaration{},
					Statements:   []parser.Statement{tt.stmt},
				},
			}

			ast := &parser.TranslationUnit{
				Declarations: []parser.Declaration{fnDecl},
			}

			ir, err := g.Generate(ast)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if len(ir.Functions) != 1 {
				t.Fatalf("Expected 1 function, got %d", len(ir.Functions))
			}

			fn := ir.Functions[0]

			// Count jump instructions
			jmpCount := 0
			for _, block := range fn.Blocks {
				for _, instr := range block.Instrs {
					switch instr.Opcode() {
					case OpJmp, OpJmpIf, OpJmpUnless:
						jmpCount++
					}
				}
			}

			if jmpCount < tt.wantJmps {
				t.Errorf("Expected at least %d jumps, got %d", tt.wantJmps, jmpCount)
			}
		})
	}
}

// TestIRGeneratorExpressions tests expression lowering.
func TestIRGeneratorExpressions(t *testing.T) {
	tests := []struct {
		name string
		expr parser.Expr
	}{
		{
			name: "binary expression",
			expr: &parser.BinaryExpr{
				Op:    lexer.ASSIGN,
				Left:  &parser.IntLiteral{Value: 1},
				Right: &parser.IntLiteral{Value: 2},
			},
		},
		{
			name: "unary expression",
			expr: &parser.UnaryExpr{
				Op:      lexer.SUB,
				Operand: &parser.IntLiteral{Value: 5},
			},
		},
		{
			name: "conditional expression",
			expr: &parser.CondExpr{
				Cond:  &parser.IntLiteral{Value: 1},
				True:  &parser.IntLiteral{Value: 10},
				False: &parser.IntLiteral{Value: 20},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewIRGenerator(nil)

			fnDecl := &parser.FunctionDecl{
				Name:   "test",
				Type:   &parser.FuncType{Return: &parser.BaseType{Kind: parser.TypeInt}},
				Params: []*parser.ParamDecl{},
				Body: &parser.CompoundStmt{
					Declarations: []parser.Declaration{},
					Statements: []parser.Statement{
						&parser.ReturnStmt{Value: tt.expr},
					},
				},
			}

			ast := &parser.TranslationUnit{
				Declarations: []parser.Declaration{fnDecl},
			}

			_, err := g.Generate(ast)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// If we get here without error, the expression was lowered successfully
		})
	}
}

// TestIRGeneratorStringOutput tests IR string output.
func TestIRGeneratorStringOutput(t *testing.T) {
	g := NewIRGenerator(nil)

	fnDecl := &parser.FunctionDecl{
		Name: "main",
		Type: &parser.FuncType{
			Return: &parser.BaseType{Kind: parser.TypeInt},
		},
		Params: []*parser.ParamDecl{},
		Body: &parser.CompoundStmt{
			Declarations: []parser.Declaration{},
			Statements: []parser.Statement{
				&parser.ReturnStmt{
					Value: &parser.IntLiteral{Value: 0},
				},
			},
		},
	}

	ast := &parser.TranslationUnit{
		Declarations: []parser.Declaration{fnDecl},
	}

	ir, err := g.Generate(ast)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	str := ir.String()
	if str == "" {
		t.Error("IR.String() returned empty string")
	}

	// Check that string contains expected elements
	expectedSubstrings := []string{"IR {", "Functions:", "Function main", "ret"}
	for _, substr := range expectedSubstrings {
		if !contains(str, substr) {
			t.Errorf("IR.String() does not contain %q", substr)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}