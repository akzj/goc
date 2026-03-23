// Package parser provides parsing of C11 source code into an Abstract Syntax Tree (AST).
// This file contains tests for type definitions in type.go.
package parser

import (
	"testing"
)

// TestTypeKindString tests the TypeKind.String() method.
func TestTypeKindString(t *testing.T) {
	tests := []struct {
		kind     TypeKind
		expected string
	}{
		{TypeVoid, "void"},
		{TypeBool, "_Bool"},
		{TypeChar, "char"},
		{TypeShort, "short"},
		{TypeInt, "int"},
		{TypeLong, "long"},
		{TypeFloat, "float"},
		{TypeDouble, "double"},
		{TypePointer, "pointer"},
		{TypeArray, "array"},
		{TypeFunction, "function"},
		{TypeStruct, "struct"},
		{TypeUnion, "union"},
		{TypeEnum, "enum"},
		{TypeTypedef, "typedef"},
		{TypeQualified, "qualified"},
		{TypeKind(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.kind.String()
			if got != tt.expected {
				t.Errorf("TypeKind(%d).String() = %q, want %q", tt.kind, got, tt.expected)
			}
		})
	}
}

// TestBaseType tests the BaseType implementation.
func TestBaseType(t *testing.T) {
	t.Run("TypeKind", func(t *testing.T) {
		bt := &BaseType{Kind: TypeInt}
		if bt.TypeKind() != TypeInt {
			t.Errorf("TypeKind() = %v, want %v", bt.TypeKind(), TypeInt)
		}
	})

	t.Run("String signed int", func(t *testing.T) {
		bt := &BaseType{Kind: TypeInt, Signed: true}
		s := bt.String()
		if s != "int" {
			t.Errorf("String() = %q, want 'int'", s)
		}
	})

	t.Run("String unsigned int", func(t *testing.T) {
		bt := &BaseType{Kind: TypeInt, Signed: false}
		s := bt.String()
		if s != "unsigned int" {
			t.Errorf("String() = %q, want 'unsigned int'", s)
		}
	})

	t.Run("String long", func(t *testing.T) {
		bt := &BaseType{Kind: TypeInt, Long: 1, Signed: true}
		s := bt.String()
		if s != "long" {
			t.Errorf("String() = %q, want 'long'", s)
		}
	})

	t.Run("String long long", func(t *testing.T) {
		bt := &BaseType{Kind: TypeInt, Long: 2, Signed: true}
		s := bt.String()
		if s != "long long" {
			t.Errorf("String() = %q, want 'long long'", s)
		}
	})

	t.Run("String long double", func(t *testing.T) {
		bt := &BaseType{Kind: TypeDouble, Long: 1}
		s := bt.String()
		if s != "long double" {
			t.Errorf("String() = %q, want 'long double'", s)
		}
	})

	t.Run("Size int", func(t *testing.T) {
		bt := &BaseType{Kind: TypeInt}
		if bt.Size() != 4 {
			t.Errorf("Size() = %d, want 4", bt.Size())
		}
	})

	t.Run("Size char", func(t *testing.T) {
		bt := &BaseType{Kind: TypeChar}
		if bt.Size() != 1 {
			t.Errorf("Size() = %d, want 1", bt.Size())
		}
	})

	t.Run("Size short", func(t *testing.T) {
		bt := &BaseType{Kind: TypeShort}
		if bt.Size() != 2 {
			t.Errorf("Size() = %d, want 2", bt.Size())
		}
	})

	t.Run("Size long", func(t *testing.T) {
		bt := &BaseType{Kind: TypeLong, Long: 0}
		if bt.Size() != 8 {
			t.Errorf("Size() = %d, want 8", bt.Size())
		}
	})

	t.Run("Size float", func(t *testing.T) {
		bt := &BaseType{Kind: TypeFloat}
		if bt.Size() != 4 {
			t.Errorf("Size() = %d, want 4", bt.Size())
		}
	})

	t.Run("Size double", func(t *testing.T) {
		bt := &BaseType{Kind: TypeDouble}
		if bt.Size() != 8 {
			t.Errorf("Size() = %d, want 8", bt.Size())
		}
	})

	t.Run("Size long double", func(t *testing.T) {
		bt := &BaseType{Kind: TypeDouble, Long: 1}
		if bt.Size() != 16 {
			t.Errorf("Size() = %d, want 16", bt.Size())
		}
	})

	t.Run("Size void", func(t *testing.T) {
		bt := &BaseType{Kind: TypeVoid}
		if bt.Size() != -1 {
			t.Errorf("Size() = %d, want -1 (incomplete)", bt.Size())
		}
	})

	t.Run("Size bool", func(t *testing.T) {
		bt := &BaseType{Kind: TypeBool}
		if bt.Size() != 1 {
			t.Errorf("Size() = %d, want 1", bt.Size())
		}
	})

	t.Run("Align int", func(t *testing.T) {
		bt := &BaseType{Kind: TypeInt}
		if bt.Align() != 4 {
			t.Errorf("Align() = %d, want 4", bt.Align())
		}
	})

	t.Run("Align char", func(t *testing.T) {
		bt := &BaseType{Kind: TypeChar}
		if bt.Align() != 1 {
			t.Errorf("Align() = %d, want 1", bt.Align())
		}
	})

	t.Run("Align double", func(t *testing.T) {
		bt := &BaseType{Kind: TypeDouble}
		if bt.Align() != 8 {
			t.Errorf("Align() = %d, want 8", bt.Align())
		}
	})

	t.Run("Align void", func(t *testing.T) {
		bt := &BaseType{Kind: TypeVoid}
		if bt.Align() != -1 {
			t.Errorf("Align() = %d, want -1 (incomplete)", bt.Align())
		}
	})

	t.Run("String void", func(t *testing.T) {
		bt := &BaseType{Kind: TypeVoid}
		s := bt.String()
		if s != "void" {
			t.Errorf("String() = %q, want 'void'", s)
		}
	})

	t.Run("String bool", func(t *testing.T) {
		bt := &BaseType{Kind: TypeBool}
		s := bt.String()
		if s != "_Bool" {
			t.Errorf("String() = %q, want '_Bool'", s)
		}
	})
}

// TestPointerType tests the PointerType implementation.
func TestPointerType(t *testing.T) {
	t.Run("TypeKind", func(t *testing.T) {
		pt := &PointerType{Elem: &BaseType{Kind: TypeInt}}
		if pt.TypeKind() != TypePointer {
			t.Errorf("TypeKind() = %v, want %v", pt.TypeKind(), TypePointer)
		}
	})

	t.Run("String", func(t *testing.T) {
		pt := &PointerType{Elem: &BaseType{Kind: TypeInt, Signed: true}}
		s := pt.String()
		if s != "int*" {
			t.Errorf("String() = %q, want 'int*'", s)
		}
	})

	t.Run("String void pointer", func(t *testing.T) {
		pt := &PointerType{Elem: nil}
		s := pt.String()
		if s != "void*" {
			t.Errorf("String() = %q, want 'void*'", s)
		}
	})

	t.Run("String nested pointer", func(t *testing.T) {
		pt := &PointerType{Elem: &PointerType{Elem: &BaseType{Kind: TypeChar, Signed: true}}}
		s := pt.String()
		if s != "char**" {
			t.Errorf("String() = %q, want 'char**'", s)
		}
	})

	t.Run("Size", func(t *testing.T) {
		pt := &PointerType{Elem: &BaseType{Kind: TypeInt}}
		if pt.Size() != 8 {
			t.Errorf("Size() = %d, want 8 (x86-64 pointer size)", pt.Size())
		}
	})

	t.Run("Size void pointer", func(t *testing.T) {
		pt := &PointerType{Elem: nil}
		if pt.Size() != 8 {
			t.Errorf("Size() = %d, want 8", pt.Size())
		}
	})

	t.Run("Align", func(t *testing.T) {
		pt := &PointerType{Elem: &BaseType{Kind: TypeInt}}
		if pt.Align() != 8 {
			t.Errorf("Align() = %d, want 8 (x86-64 pointer alignment)", pt.Align())
		}
	})
}

// TestArrayType tests the ArrayType implementation.
func TestArrayType(t *testing.T) {
	t.Run("TypeKind", func(t *testing.T) {
		at := &ArrayType{Elem: &BaseType{Kind: TypeInt}, ArraySize: 10}
		if at.TypeKind() != TypeArray {
			t.Errorf("TypeKind() = %v, want %v", at.TypeKind(), TypeArray)
		}
	})

	t.Run("String fixed size", func(t *testing.T) {
		at := &ArrayType{Elem: &BaseType{Kind: TypeInt, Signed: true}, ArraySize: 10}
		s := at.String()
		if s != "int[10]" {
			t.Errorf("String() = %q, want 'int[10]'", s)
		}
	})

	t.Run("String incomplete", func(t *testing.T) {
		at := &ArrayType{Elem: &BaseType{Kind: TypeInt, Signed: true}, ArraySize: -1}
		s := at.String()
		if s != "int[]" {
			t.Errorf("String() = %q, want 'int[]'", s)
		}
	})

	t.Run("String nil elem incomplete", func(t *testing.T) {
		at := &ArrayType{Elem: nil, ArraySize: -1}
		s := at.String()
		if s != "[]" {
			t.Errorf("String() = %q, want '[]'", s)
		}
	})

	t.Run("String nil elem fixed", func(t *testing.T) {
		at := &ArrayType{Elem: nil, ArraySize: 5}
		s := at.String()
		if s != "[5]" {
			t.Errorf("String() = %q, want '[5]'", s)
		}
	})

	t.Run("Size", func(t *testing.T) {
		at := &ArrayType{Elem: &BaseType{Kind: TypeInt}, ArraySize: 10}
		if at.Size() != 40 {
			t.Errorf("Size() = %d, want 40 (10 * 4)", at.Size())
		}
	})

	t.Run("Size incomplete", func(t *testing.T) {
		at := &ArrayType{Elem: &BaseType{Kind: TypeInt}, ArraySize: -1}
		if at.Size() != -1 {
			t.Errorf("Size() = %d, want -1 (incomplete)", at.Size())
		}
	})

	t.Run("Size nil elem", func(t *testing.T) {
		at := &ArrayType{Elem: nil, ArraySize: 10}
		if at.Size() != -1 {
			t.Errorf("Size() = %d, want -1", at.Size())
		}
	})

	t.Run("Size incomplete element type", func(t *testing.T) {
		at := &ArrayType{Elem: &BaseType{Kind: TypeVoid}, ArraySize: 10}
		if at.Size() != -1 {
			t.Errorf("Size() = %d, want -1 (incomplete element type)", at.Size())
		}
	})

	t.Run("Align", func(t *testing.T) {
		at := &ArrayType{Elem: &BaseType{Kind: TypeInt}, ArraySize: 10}
		if at.Align() != 4 {
			t.Errorf("Align() = %d, want 4 (element alignment)", at.Align())
		}
	})

	t.Run("Align nil elem", func(t *testing.T) {
		at := &ArrayType{Elem: nil, ArraySize: 10}
		if at.Align() != -1 {
			t.Errorf("Align() = %d, want -1", at.Align())
		}
	})
}

// TestFuncType tests the FuncType implementation.
func TestFuncType(t *testing.T) {
	t.Run("TypeKind", func(t *testing.T) {
		ft := &FuncType{Return: &BaseType{Kind: TypeInt}}
		if ft.TypeKind() != TypeFunction {
			t.Errorf("TypeKind() = %v, want %v", ft.TypeKind(), TypeFunction)
		}
	})

	t.Run("String no params", func(t *testing.T) {
		ft := &FuncType{Return: &BaseType{Kind: TypeInt, Signed: true}}
		s := ft.String()
		if s != "func() int" {
			t.Errorf("String() = %q, want 'func() int'", s)
		}
	})

	t.Run("String with params", func(t *testing.T) {
		ft := &FuncType{
			Return: &BaseType{Kind: TypeInt, Signed: true},
			Params: []Type{
				&BaseType{Kind: TypeInt, Signed: true},
				&BaseType{Kind: TypeChar, Signed: true},
			},
		}
		s := ft.String()
		if s != "func(int, char) int" {
			t.Errorf("String() = %q, want 'func(int, char) int'", s)
		}
	})

	t.Run("String variadic", func(t *testing.T) {
		ft := &FuncType{
			Return:   &BaseType{Kind: TypeInt, Signed: true},
			Params:   []Type{&BaseType{Kind: TypeInt, Signed: true}},
			Variadic: true,
		}
		s := ft.String()
		if s != "func(int, ...) int" {
			t.Errorf("String() = %q, want 'func(int, ...) int'", s)
		}
	})

	t.Run("String variadic no params", func(t *testing.T) {
		ft := &FuncType{
			Return:   &BaseType{Kind: TypeInt, Signed: true},
			Variadic: true,
		}
		s := ft.String()
		if s != "func(...) int" {
			t.Errorf("String() = %q, want 'func(...) int'", s)
		}
	})

	t.Run("String nil return", func(t *testing.T) {
		ft := &FuncType{Return: nil}
		s := ft.String()
		if s != "func()" {
			t.Errorf("String() = %q, want 'func()'", s)
		}
	})

	t.Run("String nil param", func(t *testing.T) {
		ft := &FuncType{
			Return: &BaseType{Kind: TypeInt, Signed: true},
			Params: []Type{nil},
		}
		s := ft.String()
		if s != "func(void) int" {
			t.Errorf("String() = %q, want 'func(void) int'", s)
		}
	})

	t.Run("Size", func(t *testing.T) {
		ft := &FuncType{Return: &BaseType{Kind: TypeInt}}
		if ft.Size() != -1 {
			t.Errorf("Size() = %d, want -1 (incomplete type)", ft.Size())
		}
	})

	t.Run("Align", func(t *testing.T) {
		ft := &FuncType{Return: &BaseType{Kind: TypeInt}}
		if ft.Align() != -1 {
			t.Errorf("Align() = %d, want -1 (incomplete type)", ft.Align())
		}
	})
}

// TestStructType tests the StructType implementation.
func TestStructType(t *testing.T) {
	t.Run("TypeKind struct", func(t *testing.T) {
		st := &StructType{Name: "S", IsUnion: false}
		if st.TypeKind() != TypeStruct {
			t.Errorf("TypeKind() = %v, want %v", st.TypeKind(), TypeStruct)
		}
	})

	t.Run("TypeKind union", func(t *testing.T) {
		st := &StructType{Name: "U", IsUnion: true}
		if st.TypeKind() != TypeUnion {
			t.Errorf("TypeKind() = %v, want %v", st.TypeKind(), TypeUnion)
		}
	})

	t.Run("String struct with name", func(t *testing.T) {
		st := &StructType{Name: "MyStruct"}
		s := st.String()
		if s != "struct MyStruct" {
			t.Errorf("String() = %q, want 'struct MyStruct'", s)
		}
	})

	t.Run("String struct anonymous", func(t *testing.T) {
		st := &StructType{Name: ""}
		s := st.String()
		if s != "struct" {
			t.Errorf("String() = %q, want 'struct'", s)
		}
	})

	t.Run("String union with name", func(t *testing.T) {
		st := &StructType{Name: "MyUnion", IsUnion: true}
		s := st.String()
		if s != "union MyUnion" {
			t.Errorf("String() = %q, want 'union MyUnion'", s)
		}
	})

	t.Run("String union anonymous", func(t *testing.T) {
		st := &StructType{Name: "", IsUnion: true}
		s := st.String()
		if s != "union" {
			t.Errorf("String() = %q, want 'union'", s)
		}
	})

	t.Run("Size", func(t *testing.T) {
		st := &StructType{TotalSize: 16}
		if st.Size() != 16 {
			t.Errorf("Size() = %d, want 16", st.Size())
		}
	})

	t.Run("Size zero", func(t *testing.T) {
		st := &StructType{TotalSize: 0}
		if st.Size() != 0 {
			t.Errorf("Size() = %d, want 0", st.Size())
		}
	})

	t.Run("Align", func(t *testing.T) {
		st := &StructType{AlignReq: 8}
		if st.Align() != 8 {
			t.Errorf("Align() = %d, want 8", st.Align())
		}
	})
}

// TestEnumType tests the EnumType implementation.
func TestEnumType(t *testing.T) {
	t.Run("TypeKind", func(t *testing.T) {
		et := &EnumType{Name: "E"}
		if et.TypeKind() != TypeEnum {
			t.Errorf("TypeKind() = %v, want %v", et.TypeKind(), TypeEnum)
		}
	})

	t.Run("String with name", func(t *testing.T) {
		et := &EnumType{Name: "Color"}
		s := et.String()
		if s != "enum Color" {
			t.Errorf("String() = %q, want 'enum Color'", s)
		}
	})

	t.Run("String anonymous", func(t *testing.T) {
		et := &EnumType{Name: ""}
		s := et.String()
		if s != "enum" {
			t.Errorf("String() = %q, want 'enum'", s)
		}
	})

	t.Run("Size", func(t *testing.T) {
		et := &EnumType{}
		if et.Size() != 4 {
			t.Errorf("Size() = %d, want 4 (x86-64 enum size)", et.Size())
		}
	})

	t.Run("Align", func(t *testing.T) {
		et := &EnumType{}
		if et.Align() != 4 {
			t.Errorf("Align() = %d, want 4 (x86-64 enum alignment)", et.Align())
		}
	})
}

// TestTypedefType tests the TypedefType implementation.
func TestTypedefType(t *testing.T) {
	t.Run("TypeKind", func(t *testing.T) {
		tt := &TypedefType{Name: "size_t"}
		if tt.TypeKind() != TypeTypedef {
			t.Errorf("TypeKind() = %v, want %v", tt.TypeKind(), TypeTypedef)
		}
	})

	t.Run("String", func(t *testing.T) {
		tt := &TypedefType{Name: "size_t"}
		s := tt.String()
		if s != "size_t" {
			t.Errorf("String() = %q, want 'size_t'", s)
		}
	})

	t.Run("Size with underlying", func(t *testing.T) {
		tt := &TypedefType{
			Name:       "size_t",
			Underlying: &BaseType{Kind: TypeLong},
		}
		if tt.Size() != 8 {
			t.Errorf("Size() = %d, want 8", tt.Size())
		}
	})

	t.Run("Size nil underlying", func(t *testing.T) {
		tt := &TypedefType{Name: "unknown_t", Underlying: nil}
		if tt.Size() != -1 {
			t.Errorf("Size() = %d, want -1", tt.Size())
		}
	})

	t.Run("Align with underlying", func(t *testing.T) {
		tt := &TypedefType{
			Name:       "size_t",
			Underlying: &BaseType{Kind: TypeLong},
		}
		if tt.Align() != 8 {
			t.Errorf("Align() = %d, want 8", tt.Align())
		}
	})

	t.Run("Align nil underlying", func(t *testing.T) {
		tt := &TypedefType{Name: "unknown_t", Underlying: nil}
		if tt.Align() != -1 {
			t.Errorf("Align() = %d, want -1", tt.Align())
		}
	})
}

// TestQualifiedType tests the QualifiedType implementation.
func TestQualifiedType(t *testing.T) {
	t.Run("TypeKind", func(t *testing.T) {
		qt := &QualifiedType{Type: &BaseType{Kind: TypeInt}}
		if qt.TypeKind() != TypeQualified {
			t.Errorf("TypeKind() = %v, want %v", qt.TypeKind(), TypeQualified)
		}
	})

	t.Run("String const", func(t *testing.T) {
		qt := &QualifiedType{Type: &BaseType{Kind: TypeInt, Signed: true}, IsConst: true}
		s := qt.String()
		if s != "const int" {
			t.Errorf("String() = %q, want 'const int'", s)
		}
	})

	t.Run("String volatile", func(t *testing.T) {
		qt := &QualifiedType{Type: &BaseType{Kind: TypeInt, Signed: true}, IsVolatile: true}
		s := qt.String()
		if s != "volatile int" {
			t.Errorf("String() = %q, want 'volatile int'", s)
		}
	})

	t.Run("String atomic", func(t *testing.T) {
		qt := &QualifiedType{Type: &BaseType{Kind: TypeInt, Signed: true}, IsAtomic: true}
		s := qt.String()
		if s != "_Atomic int" {
			t.Errorf("String() = %q, want '_Atomic int'", s)
		}
	})

	t.Run("String multiple qualifiers", func(t *testing.T) {
		qt := &QualifiedType{
			Type:       &BaseType{Kind: TypeInt, Signed: true},
			IsConst:    true,
			IsVolatile: true,
			IsAtomic:   true,
		}
		s := qt.String()
		if s != "const volatile _Atomic int" {
			t.Errorf("String() = %q, want 'const volatile _Atomic int'", s)
		}
	})

	t.Run("String nil type", func(t *testing.T) {
		qt := &QualifiedType{IsConst: true}
		s := qt.String()
		if s != "const " {
			t.Errorf("String() = %q, want 'const '", s)
		}
	})

	t.Run("Size", func(t *testing.T) {
		qt := &QualifiedType{
			Type:    &BaseType{Kind: TypeInt},
			IsConst: true,
		}
		if qt.Size() != 4 {
			t.Errorf("Size() = %d, want 4", qt.Size())
		}
	})

	t.Run("Size nil type", func(t *testing.T) {
		qt := &QualifiedType{Type: nil, IsConst: true}
		if qt.Size() != -1 {
			t.Errorf("Size() = %d, want -1", qt.Size())
		}
	})

	t.Run("Align", func(t *testing.T) {
		qt := &QualifiedType{
			Type:    &BaseType{Kind: TypeInt},
			IsConst: true,
		}
		if qt.Align() != 4 {
			t.Errorf("Align() = %d, want 4", qt.Align())
		}
	})

	t.Run("Align nil type", func(t *testing.T) {
		qt := &QualifiedType{Type: nil, IsConst: true}
		if qt.Align() != -1 {
			t.Errorf("Align() = %d, want -1", qt.Align())
		}
	})
}
