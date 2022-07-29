package generator

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/fitan/gowrap/xtype"
	"go/types"
	"os"
	"strings"
)

type Field struct {
	Path []string
	Name string
	Type *xtype.Type
	// slice map struct *
	TypeName string
}

func (f Field) SrcIdPath() *jen.Statement {
	return jen.Id("src." + strings.Join(f.Path,"."))
}

func (f Field) DestIdPath() *jen.Statement {
	return jen.Id("dest." + strings.Join(f.Path, "."))
}

func (f Field) FieldName() string {
	if len(f.Path) == 0 {
		return ""
	}
	return f.Path[len(f.Path)-1]
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

func NewDataFieldMap(prefix []string,name string, typeID string, t types.Type) *DataFieldMap {
	m := &DataFieldMap{
		Name:       name,
		TypeID:     typeID,
		NamedMap: map[string]Field{},
		PointerMap: map[string]Field{},
		SliceMap:   map[string]Field{},
		MapMap:     map[string]Field{},
		StructMap:  map[string]Field{},
		BasicMap:   map[string]Field{},
	}
	m.Parse(prefix, name, t)
	return m
}

func (d *DataFieldMap) Parse(prefix []string, name string, t types.Type) {
	f := Field{
		Name: name,
		Path: prefix,
		Type: xtype.TypeOf(t),
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
	JenF           *jen.File
	ParentSrcPath  []string
	Src            *DataFieldMap
	ParentDestPath []string
	Dest           *DataFieldMap
	Nest           []*DTO
}

func (d *DTO) SumParentPath() string {
	return strings.Join(d.ParentSrcPath, ".") + ":" + strings.Join(d.ParentDestPath, ".")
}

func (d *DTO) Gen() *jen.File {
	fn := d.GenFuncName(d.Dest.Name, d.Src.TypeID, d.Dest.TypeID)
	//var bindVar jen.Statement
	//for _, v := range d.Dest.BasicMap {
	//	fmt.Println("xtype","ttype", "basic","name", v.Name, "id",v.Type.ID(), "unescapedid",v.Type.UnescapedID(), "jen", v.Type.TypeAsJen().GoString())
	//	srcBasicMap := d.Src.BasicMap[v.Name]
	//	bindVar.Add(v.DestIdPath().Op("=").Add(srcBasicMap.SrcIdPath()))
	//}
	//
	//for _, v := range d.Dest.SliceMap {
	//	fmt.Println("xtype","ttype","slice", "id",v.Type.ID(), "unescapedid",v.Type.UnescapedID(), "jen", v.Type.TypeAsJen().Render(os.Stdout))
	//	srcV := d.Src.SliceMap[v.Name]
	//	bindVar.Add(v.DestIdPath().Op("=").Make(v.Type.TypeAsJen(), jen.Id("0"), jen.Id("len").Call(srcV.SrcIdPath())))
	//	block := v.DestIdPath().Index(jen.Id("i")).Op("=").Id("d." + v.Type.ID()).Call(srcV.SrcIdPath().Index(jen.Id("i")))
	//	if v.Type.ListInner.Basic {
	//		block = v.DestIdPath().Index(jen.Id("i")).Op("=").Add(srcV.SrcIdPath()).Index(jen.Id("i"))
	//	}
	//	bindVar.Add(jen.For(
	//		jen.Id("i").Op(":=").Lit(0),
	//		jen.Id("i").Op("<").Id("len").Call(srcV.SrcIdPath()),
	//		jen.Id("i").Op("++")).
	//		Block(
	//			block,
	//		))
	//}
	bind := make(jen.Statement, 0)
	bind = append(bind,d.GenBasic()...)
	bind = append(bind, d.GenSlice()...)
	bind = append(bind, d.GenMap()...)
	bind = append(bind, jen.Return())

	fn.Block(bind...)
	d.JenF.Add(fn)
	for _, v := range d.Nest {
		v.Gen()
	}
	return d.JenF
}

func (d *DTO) GenBasic() jen.Statement {
	bind := make(jen.Statement, 0)
	for _, v := range d.Dest.BasicMap {
		fmt.Println("xtype","ttype", "basic","name", v.Name, "id",v.Type.ID(), "unescapedid",v.Type.UnescapedID(), "jen", v.Type.TypeAsJen().GoString())
		srcBasicMap := d.Src.BasicMap[v.Name]
		bind.Add(v.DestIdPath().Op("=").Add(srcBasicMap.SrcIdPath()))
	}
	return bind
}

func (d *DTO) GenMap() jen.Statement {
	bind := make(jen.Statement, 0)
	for _, v := range d.Dest.MapMap {
		fmt.Println("xtype","ttype","slice", "id",v.Type.ID(), "unescapedid",v.Type.UnescapedID(), "jen", v.Type.TypeAsJen().Render(os.Stdout))
		srcV := d.Src.MapMap[v.Name]
		bind.Add(v.DestIdPath().Op("=").Make(v.Type.TypeAsJen(), jen.Id("len").Call(srcV.SrcIdPath())))
		block := v.DestIdPath().Index(jen.Id("key")).Op("=").Add(srcV.SrcIdPath()).Index(jen.Id("value"))
		if !v.Type.MapValue.Basic {
			srcMapValue := srcV.Type.MapValue
			destMapValue := v.Type.MapValue
			fmt.Println("mapValue", srcMapValue.TypeAsJen().GoString())
			srcName := destMapValue.HashID(d.SumParentPath())
			destName := destMapValue.HashID(d.SumParentPath())
			nestDTO := &DTO{
				JenF:           d.JenF,
				ParentSrcPath:  append([]string{}, srcV.Path...),
				Src:            NewDataFieldMap([]string{}, srcName, srcMapValue.TypeAsJen().GoString(), srcMapValue.T),
				ParentDestPath: append([]string{}, v.Path...),
				Dest:           NewDataFieldMap([]string{}, destName, destMapValue.TypeAsJen().GoString(), destMapValue.T),
				Nest:           make([]*DTO, 0),
			}
			d.Nest = append(d.Nest, nestDTO)

			block = v.DestIdPath().Index(jen.Id("key")).Op("=").Id("d." + destName).Call(jen.Id("value"))
		}
		bind.Add(jen.For(
			jen.List(jen.Id("key"), jen.Id("value")).
				Op(":=").Range().Add(srcV.SrcIdPath()).Block(
			block,
		)))
	}
	return bind
}

func (d *DTO) GenSlice() jen.Statement {
	bind := make(jen.Statement, 0)
	for _, v := range d.Dest.SliceMap {
		fmt.Println("xtype","ttype","slice", "id",v.Type.ID(), "unescapedid",v.Type.UnescapedID(), "jen", v.Type.TypeAsJen().Render(os.Stdout))
		srcV := d.Src.SliceMap[v.Name]
		bind.Add(v.DestIdPath().Op("=").Make(v.Type.TypeAsJen(), jen.Id("0"), jen.Id("len").Call(srcV.SrcIdPath())))
		block := v.DestIdPath().Index(jen.Id("i")).Op("=").Add(srcV.SrcIdPath()).Index(jen.Id("i"))
		if !v.Type.ListInner.Basic {
			srcLiner := srcV.Type.ListInner
			destLiner := v.Type.ListInner
			fmt.Println("listInner",srcLiner.TypeAsJen().GoString())
			//srcName := srcLiner.HashID(d.SumParentPath())
			srcName := srcV.FieldName()
			destName := v.FieldName()
			//destName := srcLiner.HashID(d.SumParentPath())
			nestDTO := &DTO{
				JenF:           d.JenF,
				ParentSrcPath:  append([]string{}, srcV.Path...),
				Src:            NewDataFieldMap([]string{}, srcName, srcLiner.TypeAsJen().GoString(), srcLiner.T),
				ParentDestPath: append([]string{}, v.Path...),
				Dest:           NewDataFieldMap([]string{}, destName, destLiner.TypeAsJen().GoString(), destLiner.T),
				Nest:           make([]*DTO, 0),
			}
			d.Nest = append(d.Nest, nestDTO)

			block = v.DestIdPath().Index(jen.Id("i")).Op("=").Id("d." + destName).Call(srcV.SrcIdPath().Index(jen.Id("i")))
		}
		bind.Add(jen.For(
			jen.Id("i").Op(":=").Lit(0),
			jen.Id("i").Op("<").Id("len").Call(srcV.SrcIdPath()),
			jen.Id("i").Op("++")).
			Block(
				block,
			))
	}
	return bind
}

//func (d *DTO) GenSliceFunc() {
//	bind := make(jen.Statement, 0)
//	for _, destV := range d.Dest.SliceMap {
//		if destV.Type.ListInner.Basic {
//			continue
//		}
//		srcV := d.Src.SliceMap[destV.Name]
//		FnName := destV.Type.ID()
//		Fn := d.GenFuncName(FnName, srcV.Type.ListInner.TypeAsJen(), destV.Type.ListInner.TypeAsJen()).Block(
//			jen.Return(),
//		)
//		d.JenF.Add(Fn)
//	}
//}

func (d *DTO) GenFuncName(funcName string, srcId, destId string) *jen.Statement {
	return jen.Func().Params(jen.Id("d").Id("*DTO")).
		Id(funcName).Params(jen.Id("src").Id(srcId)).Params(jen.Id("dest").Id(destId))
}
