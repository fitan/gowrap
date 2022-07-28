package generator

import (
	"github.com/dave/jennifer/jen"
	"go/types"
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
	Type *Type
	// slice map struct *
	TypeName string
}

type DataFieldMap struct {
	Name       string
	TypeID     string
	NamedMap   map[string]Field
	PointerMap map[string]Field
	SliceMap   map[string]Field
	MapMap     map[string]Field
	StructMap  map[string]Field
	BasicMap   map[string]Field
}

func (d *DataFieldMap) Parse(prefix []string, name string, t types.Type) {
	f := Field{
		Name: name,
		Path: prefix,
		Type: TypeOf(t),
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
			d.Parse(append(prefix, field.Name()), field.Name(), field.Type())
		}
	default:
		panic("unknown types.Type " + t.String())
	}
}

type DTO struct {
	JenF *jen.File
	Src  *DataFieldMap
	Dest *DataFieldMap
}

func (d *DTO) GenBasic() {
	fn := d.GenFuncName(d.Src.Name+"To"+d.Dest.Name, d.Src.TypeID, d.Dest.TypeID)
	var bindVar jen.Statement
	for _, v := range d.Dest.BasicMap {
		srcBasicMap := d.Src.BasicMap[v.Name]
		bindVar.Add(jen.Id("dest." + strings.Join(v.Path, ".")).Op("=").Id("src." + strings.Join(srcBasicMap.Path, ".")))
	}

	for _, v := range d.Dest.SliceMap {
		srcV := d.Src.SliceMap[v.Name]
		bindVar.Add(jen.Id("dest."+strings.Join(v.Path, ".")).Op("=").Make(jen.Id("[]"+v.Type.ListInner.T.String()), jen.Id("0"), jen.Id("len").Call(jen.Id("src."+strings.Join(v.Path, ".")))))
		block := jen.Id("dest." + strings.Join(v.Path, ".")).Index(jen.Id("i")).Op("=").Id("d." + v.Name).Call(jen.Id("src." + v.Name).Index(jen.Id("i")))
		if v.Type.ListInner.Basic {
			block = jen.Id("dest." + strings.Join(v.Path, ".")).Index(jen.Id("i")).Op("=").Id("src." + strings.Join(srcV.Path, ".")).Index(jen.Id("i"))
		}
		bindVar.Add(jen.For(
			jen.Id("i").Op(":=").Lit(0),
			jen.Id("i").Op("<").Id("len").Call(jen.Id("src."+strings.Join(v.Path, "."))),
			jen.Id("i").Op("++")).
			Block(
				block,
			))
		bindVar.Add(jen.Return())
	}
	fn.Block(bindVar...)
	d.JenF.Add(fn)
}

func (d *DTO) GenSlice() {
	for _, destV := range d.Dest.SliceMap {
		srcV := d.Src.SliceMap[destV.Name]
		FnName := destV.Name + "Slice"
		Fn := d.GenFuncName(FnName, srcV.Type.ListInner.T.String(), destV.Type.ListInner.T.String()).Block(
			jen.Return(),
		)
		d.JenF.Add(Fn)
	}
}

func (d *DTO) GenFuncName(funcName string, srcId string, destId string) *jen.Statement {
	return jen.Func().Params(jen.Id("d").Id("*DTO")).
		Id(funcName).Params(jen.Id("src").Id(srcId)).Params(jen.Id("dest").Id(destId))
}
