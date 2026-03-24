package ir

import (
	"fmt"
	"testing"

	"github.com/akzj/goc/internal/errhand"
	"github.com/akzj/goc/pkg/lexer"
	"github.com/akzj/goc/pkg/parser"
	"github.com/akzj/goc/pkg/semantic"
)

func TestIRGeneratorAfterSemantic(t *testing.T) {
	source := `
int add(int a, int b) {
    return a + b;
}

int main() {
    int x = 5;
    return 0;
}
`
	errorHandler := errhand.NewErrorHandler()
	l := lexer.NewLexer(source, "test.c")
	tokens := l.Tokenize()
	fmt.Printf("Tokens: %d\n", len(tokens))
	
	p := parser.NewParser(tokens, errorHandler)
	ast, err := p.Parse()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
	}
	
	fmt.Printf("After Parse - AST Declarations: %d\n", len(ast.Declarations))
	for i, decl := range ast.Declarations {
		fmt.Printf("  [%d] %T: %v\n", i, decl, decl)
	}
	
	// Run semantic analysis
	sem := semantic.NewSemanticAnalyzer(errorHandler)
	if err := sem.Analyze(ast); err != nil {
		fmt.Printf("Semantic error: %v\n", err)
	}
	
	fmt.Printf("After Semantic - AST Declarations: %d\n", len(ast.Declarations))
	for i, decl := range ast.Declarations {
		fmt.Printf("  [%d] %T: %v\n", i, decl, decl)
	}
	
	// Run IR generation
	irGen := NewIRGenerator(errorHandler)
	irResult, err := irGen.Generate(ast)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	fmt.Printf("IR: %d functions, %d globals\n", len(irResult.Functions), len(irResult.Globals))
	for i, fn := range irResult.Functions {
		fmt.Printf("  Function[%d]: %s with %d blocks\n", i, fn.Name, len(fn.Blocks))
	}
	
	if len(irResult.Functions) == 0 {
		t.Errorf("Expected at least 1 function, got 0")
	}
}
