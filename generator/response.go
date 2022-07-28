package generator

import (
	"github.com/dave/jennifer/jen"
	"go/types"
	"golang.org/x/tools/go/packages"
	"strings"
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


type Field struct {
	Path []string
	Name string
	//Type *Type
	// slice map struct *
	TypeName string
}

type DataFieldMap struct {
	Name string
	TypeID string
	NamedMap map[string]Field
	PointerMap map[string]Field
	SliceMap map[string]Field
	MapMap map[string]Field
	StructMap map[string]Field
	BasicMap map[string]Field
}

func (d *DataFieldMap)Parse(prefix []string, name string,t types.Type) {
	f := Field{
		Name: name,
		Path: prefix,
		//Type: TypeOf(t),
	}

	switch v := t.(type) {
	case *types.Pointer:
		d.PointerMap[name] = f
	case *types.Basic:
		d.BasicMap[name] = f
		return
	case *types.Map:
		d.MapMap[name] = f
		return
	case *types.Slice:
		d.SliceMap[name] = f
	case *types.Array:
	case *types.Named:
		d.Parse(append(prefix), name, v.Underlying())
	case *types.Struct:
		for i := 0; i < v.NumFields(); i++ {
			field := v.Field(i)
			if !field.Exported() {
				continue
			}
			d.Parse(append(prefix,field.Name()), field.Name(), field.Type())
		}
	default:
		panic("unknown types.Type " + t.String())
	}
}

func GenMethod(statement jen.Statement, src DataFieldMap, dest DataFieldMap) *jen.Statement {
	method := jen.Func().Params(jen.Id("d").Id("*DTO")).Id(src.Name + "To" + dest.Name).Params(jen.Id("src").Id(src.TypeID)).Id(dest.TypeID)
	var bindVar jen.Statement
	for _, v := range dest.BasicMap {
		srcBasicMap := src.BasicMap[v.Name]
		bindVar.Add(jen.Id("dest." + strings.Join(v.Path,",")).Op("=").Id("src." + strings.Join(srcBasicMap.Path, ",")))
	}
	method.Block(bindVar...)
	return method
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

	namedMap map[string]Field
	pointerMap map[string]Field
	sliceMap map[string]Field
	mapMap map[string]Field
	structMap map[string]Field
	basicMap map[string]Field
}




func (d *DTO)Parse(prefix []string, name string,t types.Type) {
	f := Field{
		Name: name,
		Path: prefix,
		//Type: TypeOf(t),
	}

	switch v := t.(type) {
	case *types.Pointer:
		d.pointerMap[name] = f
	case *types.Basic:
		d.basicMap[name] = f
		return
	case *types.Map:
		d.mapMap[name] = f
		return
	case *types.Slice:
		d.sliceMap[name] = f
	case *types.Array:
	case *types.Named:
		d.Parse(append(prefix), name, v.Underlying())
	case *types.Struct:
		for i := 0; i < v.NumFields(); i++ {
			field := v.Field(i)
			if !field.Exported() {
				continue
			}
			d.Parse(append(prefix,field.Name()), field.Name(), field.Type())
		}
	default:
		panic("unknown types.Type " + t.String())
	}
}

