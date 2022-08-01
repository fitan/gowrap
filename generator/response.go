package generator

import (
	"crypto/sha1"
	"encoding/hex"
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
	path := append([]string{"src"}, f.Path...)
	return jen.Id(strings.Join(path, "."))
}

func (f Field) DestIdPath() *jen.Statement {
	path := append([]string{"dest"}, f.Path...)
	return jen.Id(strings.Join(path, "."))
}

func (f Field) FieldName(s string) string {
	if len(f.Path) == 0 {
		return ""
	}
	hash := sha1.New()
	hash.Write([]byte(s))
	b := hash.Sum(nil)
	fmt.Println("s:", s, "b:", hex.EncodeToString(b)[0:4])
	return strings.ToLower(f.Path[len(f.Path)-1][0:1]) + f.Path[len(f.Path)-1][1:] + hex.EncodeToString(b)[0:4]
}

type DataFieldMap struct {
	Name       string
	TypeID     jen.Code
	NamedMap   map[string]Field
	PointerMap map[string]Field
	SliceMap   map[string]Field
	MapMap     map[string]Field
	BasicMap   map[string]Field
}

func NewDataFieldMap(prefix []string, name string, typeID jen.Code, t types.Type) *DataFieldMap {
	m := &DataFieldMap{
		Name:       name,
		TypeID:     typeID,
		NamedMap:   map[string]Field{},
		PointerMap: map[string]Field{},
		SliceMap:   map[string]Field{},
		MapMap:     map[string]Field{},
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
		d.Parse(prefix, name, v.Underlying())
	case *types.Struct:
		for i := 0; i < v.NumFields(); i++ {
			field := v.Field(i)
			if !field.Exported() {
				continue
			}
			d.Parse(append(prefix[0:], field.Name()), field.Name(), field.Type())
		}
	default:
		panic("unknown types.Type " + t.String())
	}
}

type Recorder struct {
	m map[string]struct{}
}

func NewRecorder() *Recorder {
	return &Recorder{m: map[string]struct{}{}}
}

func (r *Recorder) Reg(name string) {
	r.m[name] = struct{}{}
}

func (r *Recorder) Lookup(name string) bool {
	_, ok := r.m[name]
	return ok
}

type DTO struct {
	JenF     *jen.File
	Recorder *Recorder
	SrcParentPath []string
	SrcPath  []string
	Src      *DataFieldMap
	DestParentPath []string
	DestPath []string
	Dest     *DataFieldMap
	Nest     []*DTO
}

func (d *DTO) SumPath() string {
	return strings.Join(d.SrcPath, ".") + ":" + strings.Join(d.DestPath, ".")
}

func (d *DTO) Doc() *jen.Statement {
	code := make(jen.Statement,0)
	code = append(code,jen.Comment("parentPath: "+strings.Join(d.SrcParentPath, ".") + ":" + strings.Join(d.DestParentPath, ".")))
	code = append(code, jen.Comment("path: "+strings.Join(d.SrcPath, ".") + ":" + strings.Join(d.DestPath, ".")))
	return &code
}

func (d *DTO) SumParentPath() string {
	return strings.Join(d.SrcParentPath, ".") + ":" + strings.Join(d.DestParentPath, ".")
}

func (d *DTO) Gen() {
	has, fn := d.GenFuncName(d.Dest.Name, d.Src.TypeID, d.Dest.TypeID)
	if has {
		return
	}
	bind := make(jen.Statement, 0)
	bind = append(bind, d.GenBasic()...)
	bind = append(bind, d.GenSlice()...)
	bind = append(bind, d.GenMap()...)
	bind = append(bind, d.GenPointer()...)
	bind = append(bind, jen.Return())

	fn.Block(bind...)
	d.JenF.Add(d.Doc())
	d.JenF.Add(fn)
	for _, v := range d.Nest {
		v.Gen()
	}
}

func (d *DTO) GenBasic() jen.Statement {
	bind := make(jen.Statement, 0)
	for _, v := range d.Dest.BasicMap {
		fmt.Println("xtype", "ttype", "basic", "name", v.Name, "id", v.Type.ID(), "unescapedid", v.Type.UnescapedID(), "jen", v.Type.TypeAsJen().GoString())
		srcBasicMap := d.Src.BasicMap[v.Name]
		bind.Add(v.DestIdPath().Op("=").Add(srcBasicMap.SrcIdPath()))
	}
	return bind
}

func (d *DTO) GenMap() jen.Statement {
	bind := make(jen.Statement, 0)
	for _, v := range d.Dest.MapMap {
		fmt.Println("xtype", "ttype", "slice", "id", v.Type.ID(), "unescapedid", v.Type.UnescapedID(), "jen", v.Type.TypeAsJen().Render(os.Stdout))
		srcV := d.Src.MapMap[v.Name]
		bind.Add(v.DestIdPath().Op("=").Make(v.Type.TypeAsJen(), jen.Id("len").Call(srcV.SrcIdPath())))
		block := v.DestIdPath().Index(jen.Id("key")).Op("=").Add(srcV.SrcIdPath()).Index(jen.Id("value"))
		if !v.Type.MapValue.Basic {
			srcMapValue := srcV.Type.MapValue
			destMapValue := v.Type.MapValue
			fmt.Println("mapValue", srcMapValue.TypeAsJen().GoString())
			//srcName := destMapValue.HashID(d.SumPath())
			//destName := destMapValue.HashID(d.SumPath())
			srcName := srcV.FieldName(d.SumPath())
			destName := v.FieldName(d.SumPath())
			nestDTO := &DTO{
				JenF:     d.JenF,
				Recorder: d.Recorder,
				SrcParentPath: append(d.SrcParentPath, srcV.Path...),
				SrcPath:  append([]string{}, srcV.Path...),
				Src:      NewDataFieldMap([]string{}, srcName, srcMapValue.TypeAsJen(), srcMapValue.T),
				DestParentPath: append(d.DestParentPath, v.Path...),
				DestPath: append([]string{}, v.Path...),
				Dest:     NewDataFieldMap([]string{}, destName, destMapValue.TypeAsJen(), destMapValue.T),
				Nest:     make([]*DTO, 0),
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

func (d *DTO) GenPointer() jen.Statement {
	bind := make(jen.Statement, 0)
	for _, v := range d.Dest.PointerMap {
		srcV := d.Src.PointerMap[v.Name]
		if v.Type.PointerInner.Basic {
			bind.Add(v.DestIdPath().Op("=").Add(srcV.SrcIdPath()))
		} else {
			srcLiner := srcV.Type.PointerInner
			destLiner := v.Type.PointerInner
			srcName := srcV.FieldName(d.SumPath())
			destName := v.FieldName(d.SumPath())
			//destName := srcLiner.HashID(d.SumPath())
			nestDTO := &DTO{
				JenF:     d.JenF,
				Recorder: d.Recorder,
				SrcParentPath: append(d.SrcParentPath, srcV.Path...),
				SrcPath:  append([]string{}, srcV.Path...),
				Src:      NewDataFieldMap([]string{}, srcName, srcLiner.TypeAsJen(), srcLiner.T),
				DestParentPath: append(d.DestParentPath, v.Path...),
				DestPath: append([]string{}, v.Path...),
				Dest:     NewDataFieldMap([]string{}, destName, srcLiner.TypeAsJen(), destLiner.T),
				Nest:     make([]*DTO, 0),
			}
			d.Nest = append(d.Nest, nestDTO)

			bind.Add(
				jen.If(srcV.SrcIdPath().Op("!=").Nil()).Block(
					jen.Id("v").Op(":=").Id("d."+destName).Call(jen.Id("*").Add(srcV.SrcIdPath())),
					v.DestIdPath().Op("=").Id("&v"),
				).Else().Block(
					v.DestIdPath().Op("=").Add(srcV.SrcIdPath()),
				),
			)
		}
	}
	return bind
}

func (d *DTO) GenSlice() jen.Statement {
	bind := make(jen.Statement, 0)
	for _, v := range d.Dest.SliceMap {
		fmt.Println("xtype", "ttype", "slice", "id", v.Type.ID(), "unescapedid", v.Type.UnescapedID(), "jen", v.Type.TypeAsJen().Render(os.Stdout))
		srcV := d.Src.SliceMap[v.Name]
		bind.Add(v.DestIdPath().Op("=").Make(v.Type.TypeAsJen(), jen.Id("0"), jen.Id("len").Call(srcV.SrcIdPath())))
		block := v.DestIdPath().Index(jen.Id("i")).Op("=").Add(srcV.SrcIdPath()).Index(jen.Id("i"))
		if !v.Type.ListInner.Basic {
			srcLiner := srcV.Type.ListInner
			destLiner := v.Type.ListInner
			fmt.Println("listInner", srcLiner.TypeAsJen().GoString())
			//srcName := srcLiner.HashID(d.SumPath())
			srcName := srcV.FieldName(d.SumPath())
			destName := v.FieldName(d.SumPath())
			//destName := srcLiner.HashID(d.SumPath())
			nestDTO := &DTO{
				JenF:     d.JenF,
				Recorder: d.Recorder,
				SrcParentPath: append(d.SrcParentPath, srcV.Path...),
				//SrcPath:  append([]string{}, srcV.Path...),
				SrcPath: d.SrcPath[0:],
				Src:      NewDataFieldMap([]string{}, srcName, srcLiner.TypeAsJen(), srcLiner.T),
				DestParentPath: append(d.DestParentPath, v.Path...),
				//DestPath: append([]string{}, v.Path...),
				DestPath: d.DestPath[0:],
				Dest:     NewDataFieldMap([]string{}, destName, destLiner.TypeAsJen(), destLiner.T),
				Nest:     make([]*DTO, 0),
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

func (d *DTO) GenFuncName(funcName string, srcId, destId jen.Code) (has bool, fn *jen.Statement) {
	has = d.Recorder.Lookup(funcName)
	if has {
		return has, nil
	}
	d.Recorder.Reg(funcName)

	return false, jen.Func().Params(jen.Id("d").Id("*DTO")).
		Id(funcName).Params(jen.Id("src").Add(srcId)).Params(jen.Id("dest").Add(destId))
}
