package generator

import (
	"github.com/dave/jennifer/jen"
	"go/types"
	"golang.org/x/tools/go/packages"
)
type Type struct {
	T             types.Type
	Interface     bool
	InterfaceType *types.Interface
	Struct        bool
	StructType    *types.Struct
	Named         bool
	NamedType     *types.Named
	Pointer       bool
	PointerType   *types.Pointer
	PointerInner  *Type
	List          bool
	ListFixed     bool
	ListInner     *Type
	Map           bool
	MapType       *types.Map
	MapKey        *Type
	MapValue      *Type
	Basic         bool
	BasicType     *types.Basic
}


func TypeOf(t types.Type) *Type {
	rt := &Type{}
	rt.T = t
	switch value := t.(type) {
	case *types.Pointer:
		rt.Pointer = true
		rt.PointerType = value
		rt.PointerInner = TypeOf(value.Elem())
	case *types.Basic:
		rt.Basic = true
		rt.BasicType = value
	case *types.Map:
		rt.Map = true
		rt.MapType = value
		rt.MapKey = TypeOf(value.Key())
		rt.MapValue = TypeOf(value.Elem())
	case *types.Slice:
		rt.List = true
		rt.ListInner = TypeOf(value.Elem())
	case *types.Array:
		rt.List = true
		rt.ListFixed = true
		rt.ListInner = TypeOf(value.Elem())
	case *types.Named:
		underlying := TypeOf(value.Underlying())
		underlying.T = value
		underlying.Named = true
		underlying.NamedType = value
		return underlying
	case *types.Struct:
		rt.Struct = true
		rt.StructType = value
	case *types.Interface:
		rt.Interface = true
		rt.InterfaceType = value
	default:
		panic("unknown types.Type " + t.String())
	}
	return rt
}



type Method struct {
	Src types.Type
	Dest types.Type
}

type Field struct {
	Name string
	Type *Type
	RawType types.Type
	// slice map struct *
	TypeName string
}



type DTO struct {
	jen *jen.File
	pkg *packages.Package

	srcName string
	src types.Type

	destName string
	dest types.Type

	srcMap map[string]Field
	destMap map[string]Field

	sliceMap map[string]Field
	mapMap map[string]Field
	structMap map[string]Field
}


func (d *DTO)Parse(name string,t types.Type) map[string]Field  {
	f := Field{
		Name: name,
		Type: TypeOf(t),
	}

	d.srcMap[name] = f
	switch v := t.(type) {
	case *types.Pointer:

		break
	case *types.Basic:
		break
	case *types.Map:
	case *types.Slice:
	case *types.Array:
	case *types.Named:
	case *types.Struct:
	case *types.Interface:
	default:
		panic("unknown types.Type " + t.String())
	}
}

