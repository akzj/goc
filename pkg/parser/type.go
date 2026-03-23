// Package parser provides parsing of C11 source code into an Abstract Syntax Tree ().
// This file defines type representations.
package parser

import (
	"fmt"
)

// Type is the interface implemented by all type nodes.
type Type interface {
	// TypeKind returns the kind of type.
	TypeKind() TypeKind
	// String returns a string representation.
	String() string
	// Size returns the size in bytes (-1 if incomplete).
	Size() int64
	// Align returns the alignment requirement.
	Align() int64
}

// TypeKind represents the kind of a type.
type TypeKind int

const (
	// TypeVoid represents void.
	TypeVoid TypeKind = iota
	// TypeBool represents _Bool.
	TypeBool
	// TypeChar represents char.
	TypeChar
	// TypeShort represents short.
	TypeShort
	// TypeInt represents int.
	TypeInt
	// TypeLong represents long.
	TypeLong
	// TypeFloat represents float.
	TypeFloat
	// TypeDouble represents double.
	TypeDouble
	// TypePointer represents T*.
	TypePointer
	// TypeArray represents T[N].
	TypeArray
	// TypeFunction represents function type.
	TypeFunction
	// TypeStruct represents struct.
	TypeStruct
	// TypeUnion represents union.
	TypeUnion
	// TypeEnum represents enum.
	TypeEnum
	// TypeTypedef represents typedef name.
	TypeTypedef
	// TypeQualified represents const/volatile/_Atomic qualified type.
	TypeQualified
)

// String returns the string representation of the type kind.
func (k TypeKind) String() string {
	switch k {
	case TypeVoid:
		return "void"
	case TypeBool:
		return "_Bool"
	case TypeChar:
		return "char"
	case TypeShort:
		return "short"
	case TypeInt:
		return "int"
	case TypeLong:
		return "long"
	case TypeFloat:
		return "float"
	case TypeDouble:
		return "double"
	case TypePointer:
		return "pointer"
	case TypeArray:
		return "array"
	case TypeFunction:
		return "function"
	case TypeStruct:
		return "struct"
	case TypeUnion:
		return "union"
	case TypeEnum:
		return "enum"
	case TypeTypedef:
		return "typedef"
	case TypeQualified:
		return "qualified"
	default:
		return "unknown"
	}
}

// BaseType represents a basic type (int, char, float, etc.).
type BaseType struct {
	// Kind is the type kind.
	Kind TypeKind
	// Signed is true for signed, false for unsigned.
	Signed bool
	// Long indicates the number of 'long' modifiers (0, 1, or 2).
	Long int
}

// TypeKind implements Type.
func (b *BaseType) TypeKind() TypeKind {
	return b.Kind
}

// String implements Type.
func (b *BaseType) String() string {
	prefix := ""
	if !b.Signed {
		prefix = "unsigned "
	}
	switch b.Kind {
	case TypeVoid:
		return "void"
	case TypeBool:
		return "_Bool"
	case TypeChar:
		return prefix + "char"
	case TypeShort:
		if b.Long > 0 {
			return prefix + "long short"
		}
		return prefix + "short"
	case TypeInt:
		if b.Long == 0 {
			return prefix + "int"
		} else if b.Long == 1 {
			return prefix + "long"
		}
		return prefix + "long long"
	case TypeLong:
		if b.Long == 0 {
			return prefix + "long"
		}
		return prefix + "long long"
	case TypeFloat:
		return "float"
	case TypeDouble:
		if b.Long > 0 {
			return "long double"
		}
		return "double"
	default:
		return b.Kind.String()
	}
}

// Size implements Type.
func (b *BaseType) Size() int64 {
	switch b.Kind {
	case TypeVoid:
		return -1 // incomplete type
	case TypeBool:
		return 1
	case TypeChar:
		return 1
	case TypeShort:
		return 2
	case TypeInt:
		return 4
	case TypeLong:
		if b.Long == 0 {
			return 8 // long on x86-64
		}
		return 8 // long long
	case TypeFloat:
		return 4
	case TypeDouble:
		if b.Long > 0 {
			return 16 // long double (typically 16 on x86-64)
		}
		return 8
	default:
		return -1
	}
}

// Align implements Type.
func (b *BaseType) Align() int64 {
	switch b.Kind {
	case TypeVoid:
		return -1 // incomplete type
	case TypeBool:
		return 1
	case TypeChar:
		return 1
	case TypeShort:
		return 2
	case TypeInt:
		return 4
	case TypeLong:
		if b.Long == 0 {
			return 8 // long on x86-64
		}
		return 8 // long long
	case TypeFloat:
		return 4
	case TypeDouble:
		if b.Long > 0 {
			return 16 // long double alignment
		}
		return 8
	default:
		return -1
	}
}

// PointerType represents a pointer type (T*).
type PointerType struct {
	// Elem is the element type.
	Elem Type
}

// TypeKind implements Type.
func (p *PointerType) TypeKind() TypeKind {
	return TypePointer
}

// String implements Type.
func (p *PointerType) String() string {
	if p.Elem == nil {
		return "void*"
	}
	return p.Elem.String() + "*"
}

// Size implements Type.
func (p *PointerType) Size() int64 {
	return 8 // All pointers are 8 bytes on x86-64
}

// Align implements Type.
func (p *PointerType) Align() int64 {
	return 8 // All pointers are 8-byte aligned on x86-64
}

// ArrayType represents an array type (T[N]).
type ArrayType struct {
	// Elem is the element type.
	Elem Type
	// ArraySize is the array size (-1 for incomplete arrays).
	ArraySize int64
}

// TypeKind implements Type.
func (a *ArrayType) TypeKind() TypeKind {
	return TypeArray
}

// String implements Type.
func (a *ArrayType) String() string {
	if a.Elem == nil {
		if a.ArraySize < 0 {
			return "[]"
		}
		return "[" + fmt.Sprintf("%d", a.ArraySize) + "]"
	}
	if a.ArraySize < 0 {
		return a.Elem.String() + "[]"
	}
	return a.Elem.String() + "[" + fmt.Sprintf("%d", a.ArraySize) + "]"
}

// Size implements Type.
func (a *ArrayType) Size() int64 {
	if a.ArraySize < 0 {
		return -1 // incomplete array
	}
	if a.Elem == nil {
		return -1
	}
	elemSize := a.Elem.Size()
	if elemSize < 0 {
		return -1 // incomplete element type
	}
	return elemSize * a.ArraySize
}

// Align implements Type.
func (a *ArrayType) Align() int64 {
	if a.Elem == nil {
		return -1
	}
	return a.Elem.Align()
}

// FuncType represents a function type.
type FuncType struct {
	// Return is the return type.
	Return Type
	// Params is the list of parameter types.
	Params []Type
	// Variadic is true if the function is variadic (...).
	Variadic bool
}

// TypeKind implements Type.
func (f *FuncType) TypeKind() TypeKind {
	return TypeFunction
}

// String implements Type.
func (f *FuncType) String() string {
	params := ""
	for i, p := range f.Params {
		if i > 0 {
			params += ", "
		}
		if p == nil {
			params += "void"
		} else {
			params += p.String()
		}
	}
	if f.Variadic {
		if len(f.Params) > 0 {
			params += ", ..."
		} else {
			params = "..."
		}
	}
	if f.Return == nil {
		return "func(" + params + ")"
	}
	return "func(" + params + ") " + f.Return.String()
}

// Size implements Type.
func (f *FuncType) Size() int64 {
	return -1 // Functions are incomplete types (not objects)
}

// Align implements Type.
func (f *FuncType) Align() int64 {
	return -1 // Functions are incomplete types
}

// StructType represents a struct or union type.
type StructType struct {
	// Name is the struct/union tag name.
	Name string
	// Fields is the list of fields.
	Fields []*FieldDecl
	// IsUnion is true if this is a union.
	IsUnion bool
	// TotalSize is the total size (computed after field layout).
	TotalSize int64
	// AlignReq is the alignment requirement.
	AlignReq int64
}

// TypeKind implements Type.
func (s *StructType) TypeKind() TypeKind {
	if s.IsUnion {
		return TypeUnion
	}
	return TypeStruct
}

// String implements Type.
func (s *StructType) String() string {
	if s.IsUnion {
		if s.Name == "" {
			return "union"
		}
		return "union " + s.Name
	}
	if s.Name == "" {
		return "struct"
	}
	return "struct " + s.Name
}

// Size implements Type.
func (s *StructType) Size() int64 {
	return s.TotalSize
}

// Align implements Type.
func (s *StructType) Align() int64 {
	return s.AlignReq
}

// EnumType represents an enum type.
type EnumType struct {
	// Name is the enum tag name.
	Name string
	// Values is the list of enum values.
	Values []*EnumValue
}

// TypeKind implements Type.
func (e *EnumType) TypeKind() TypeKind {
	return TypeEnum
}

// String implements Type.
func (e *EnumType) String() string {
	if e.Name == "" {
		return "enum"
	}
	return "enum " + e.Name
}

// Size implements Type.
func (e *EnumType) Size() int64 {
	return 4 // enum is int-sized on x86-64
}

// Align implements Type.
func (e *EnumType) Align() int64 {
	return 4 // enum has int alignment on x86-64
}

// TypedefType represents a typedef name.
type TypedefType struct {
	// Name is the typedef name.
	Name string
	// Underlying is the underlying type.
	Underlying Type
}

// TypeKind implements Type.
func (t *TypedefType) TypeKind() TypeKind {
	return TypeTypedef
}

// String implements Type.
func (t *TypedefType) String() string {
	return t.Name
}

// Size implements Type.
func (t *TypedefType) Size() int64 {
	if t.Underlying == nil {
		return -1
	}
	return t.Underlying.Size()
}

// Align implements Type.
func (t *TypedefType) Align() int64 {
	if t.Underlying == nil {
		return -1
	}
	return t.Underlying.Align()
}

// QualifiedType represents a type with qualifiers (const, volatile, _Atomic).
type QualifiedType struct {
	// Type is the underlying type.
	Type Type
	// IsConst is true if const-qualified.
	IsConst bool
	// IsVolatile is true if volatile-qualified.
	IsVolatile bool
	// IsAtomic is true if _Atomic-qualified.
	IsAtomic bool
}

// TypeKind implements Type.
func (q *QualifiedType) TypeKind() TypeKind {
	return TypeQualified
}

// String implements Type.
func (q *QualifiedType) String() string {
	qualifiers := ""
	if q.IsConst {
		qualifiers += "const "
	}
	if q.IsVolatile {
		qualifiers += "volatile "
	}
	if q.IsAtomic {
		qualifiers += "_Atomic "
	}
	if q.Type == nil {
		return qualifiers
	}
	return qualifiers + q.Type.String()
}

// Size implements Type.
func (q *QualifiedType) Size() int64 {
	if q.Type == nil {
		return -1
	}
	return q.Type.Size()
}

// Align implements Type.
func (q *QualifiedType) Align() int64 {
	if q.Type == nil {
		return -1
	}
	return q.Type.Align()
}
