package generator

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/fitan/gowrap/xtype"
	"golang.org/x/tools/go/packages"
)

const ResponseTag = "copy"
const CopyMethodName = "@copy-method"

type Field struct {
	Path []string
	Name string
	Type *xtype.Type
	// slice map struct *
	TypeName string
	Doc      *ast.CommentGroup
}

func (f Field) CopyMethod() (pkgName, methodName string) {
	if f.Doc == nil {
		return
	}
	for _, v := range f.Doc.List {
		s := DocFormat(v.Text)
		if strings.HasPrefix(s, "// "+CopyMethodName) {
			params := strings.TrimPrefix(s, "// "+CopyMethodName)
			paramList := strings.Fields(params)
			if len(paramList) < 1 {
				panic("dto method format error: " + s)
			}
			if len(paramList) == 1 {
				methodName = paramList[0]
				//pkgName = paramList[0]
			} else {
				pkgName = paramList[0]
				methodName = paramList[1]
			}
			return

		}
	}
	return
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
	Pkg        *packages.Package
	Name       string
	Type       *xtype.Type
	NamedMap   map[string]Field
	PointerMap map[string]Field
	SliceMap   map[string]Field
	MapMap     map[string]Field
	BasicMap   map[string]Field
}

func NewDataFieldMap(pkg *packages.Package, prefix []string, name string, xType *xtype.Type, t types.Type) *DataFieldMap {
	m := &DataFieldMap{
		Pkg:        pkg,
		Name:       name,
		Type:       xType,
		NamedMap:   map[string]Field{},
		PointerMap: map[string]Field{},
		SliceMap:   map[string]Field{},
		MapMap:     map[string]Field{},
		BasicMap:   map[string]Field{},
	}
	m.Parse(prefix, name, t, nil)
	return m
}

func (d *DataFieldMap) saveField(m map[string]Field, name string, field Field) {
	var oldField Field
	var ok bool
	if oldField, ok = m[name]; !ok {
		m[name] = field
		return
	}

	fmt.Printf("作用域内重复定义: %s. src.DestIdPath: %s, src.SrcIdPath: %s, dest.DestIdPath: %s, dest.SrcIdPath: %s \n", name, oldField.DestIdPath().GoString(), oldField.SrcIdPath().GoString(), field.DestIdPath().GoString(), field.SrcIdPath().GoString())
	if len(oldField.Path) > len(field.Path) {
		m[name] = field
	}

	return
}

func (d *DataFieldMap) Parse(prefix []string, name string, t types.Type, doc *ast.CommentGroup) {
	f := Field{
		Name: name,
		Path: prefix,
		Type: xtype.TypeOf(t),
		Doc:  doc,
	}
	if name == "Cabinets" {
		fmt.Println("print field Cabinets", f)
	}
	if name == "Status" {
		fmt.Println("print field Status", f)
	}
	switch v := t.(type) {
	case *types.Pointer:
		//if b,ok := d.PointerMap[name]; ok {
		//	fmt.Printf("作用域内重复定义: %s. src.DestIdPath: %s, src.SrcIdPath: %s, dest.DestIdPath: %s, dest.SrcIdPath: %s \n", name, b.DestIdPath().GoString(), b.SrcIdPath().GoString(), f.DestIdPath().GoString(), f.SrcIdPath().GoString())
		//	return
		//}
		//d.PointerMap[name] = f
		d.saveField(d.PointerMap, name, f)
	case *types.Basic:
		//if b,ok := d.BasicMap[name]; ok {
		//	fmt.Printf("作用域内重复定义: %s. src.DestIdPath: %s, src.SrcIdPath: %s, dest.DestIdPath: %s, dest.SrcIdPath: %s \n", name, b.DestIdPath().GoString(), b.SrcIdPath().GoString(), f.DestIdPath().GoString(), f.SrcIdPath().GoString())
		//	return
		//}
		//d.BasicMap[name] = f
		d.saveField(d.PointerMap, name, f)
		return
	case *types.Map:
		//if b,ok := d.MapMap[name]; ok {
		//	fmt.Printf("作用域内重复定义: %s. src.DestIdPath: %s, src.SrcIdPath: %s, dest.DestIdPath: %s, dest.SrcIdPath: %s \n", name, b.DestIdPath().GoString(), b.SrcIdPath().GoString(), f.DestIdPath().GoString(), f.SrcIdPath().GoString())
		//	return
		//}
		//d.MapMap[name] = f
		d.saveField(d.MapMap, name, f)
		return
	case *types.Slice:
		//if b,ok := d.SliceMap[name]; ok {
		//	fmt.Printf("作用域内重复定义: %s. src.DestIdPath: %s, src.SrcIdPath: %s, dest.DestIdPath: %s, dest.SrcIdPath: %s \n", name, b.DestIdPath().GoString(), b.SrcIdPath().GoString(), f.DestIdPath().GoString(), f.SrcIdPath().GoString())
		//	return
		//}
		//d.SliceMap[name] = f
		d.saveField(d.SliceMap, name, f)
		return
	case *types.Array:
	case *types.Named:
		d.Parse(prefix, name, v.Underlying(), doc)
		return
	case *types.Struct:
		for i := 0; i < v.NumFields(); i++ {
			field := v.Field(i)
			if !field.Exported() {
				continue
			}
			indexName := field.Name()
			if field.Name() == "Cabinets" {
				fmt.Println("find struct Cabinets: ", field.Type().String())
			}
			tagName, ok := reflect.StructTag(v.Tag(i)).Lookup(ResponseTag)
			if ok {
				indexName = tagName
			}
			doc = GetCommentByTokenPos(d.Pkg, field.Pos())
			d.Parse(append(prefix[0:], field.Name()), indexName, field.Type(), doc)
		}
		return
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

func NewResponse(pkg *packages.Package, f *types.Func, responseName string) *Copy {
	fnName := f.Id()
	src := f.Type().(*types.Signature).Results().At(0)
	srcType := src.Type()
	_, typeFile := path.Split(src.Type().String())
	srcTypeName := strings.TrimPrefix(strings.TrimPrefix(typeFile, src.Pkg().Name()), ".")
	fmt.Println("name: ", src.Name(), "id: ", src.Id(), "typestring", src.Type(), "pkg: ", src.Pkg().Name(), "srctypename: ", srcTypeName)
	//srcName := fnType.Results.List[0].Names[0].Name
	//spew.Dump(pkg.Types.Scope())
	//fmt.Println("srcName: ", srcName)
	//srcType := pkg.TypesInfo.TypeOf(fnType.Results.List[0].Type)
	//srcType := pkg.TypesInfo.Types[fnType]
	//fmt.Println("names: ", pkg.Types.Scope().Names(), "path: ", pkg.Types.Path())
	destType := pkg.Types.Scope().Lookup(responseName)

	jenF := jen.NewFile("Copy")
	jenF.Add(jen.Type().Id(fnName + "Copy").Struct())

	dto := Copy{
		Pkg:            pkg,
		JenF:           jenF,
		Recorder:       NewRecorder(),
		SrcParentPath:  []string{},
		SrcPath:        []string{},
		Src:            NewDataFieldMap(pkg, []string{}, "src", xtype.TypeOf(srcType), srcType),
		DestParentPath: []string{},
		DestPath:       []string{},
		Dest:           NewDataFieldMap(pkg, []string{}, "dest", xtype.TypeOf(destType.Type()), destType.Type()),
		DefaultFn: jen.Func().Params(jen.Id("d").Id("*" + fnName + "Copy")).
			Id("Copy").Params(jen.Id("src").Id(srcTypeName)).Params(jen.Id("dest").Id(responseName)),
		StructName: fnName,
	}
	dto.Gen()
	return &dto
}

type Copy struct {
	Pkg            *packages.Package
	JenF           *jen.File
	Recorder       *Recorder
	SrcParentPath  []string
	SrcPath        []string
	Src            *DataFieldMap
	DestParentPath []string
	DestPath       []string
	Dest           *DataFieldMap
	Nest           []*Copy
	DefaultFn      *jen.Statement
	StructName     string
}

func (d *Copy) FnName() string {
	return d.Src.Type.ID() + "To" + upFirst(d.Dest.Type.ID())
}

func (d *Copy) SumPath() string {
	return strings.Join(d.SrcPath, ".") + ":" + strings.Join(d.DestPath, ".")
}

func (d *Copy) Doc() *jen.Statement {
	code := make(jen.Statement, 0)
	code = append(code, jen.Comment("parentPath: "+strings.Join(d.SrcParentPath, ".")+":"+strings.Join(d.DestParentPath, ".")))
	code = append(code, jen.Comment("path: "+strings.Join(d.SrcPath, ".")+":"+strings.Join(d.DestPath, ".")))
	return &code
}

func (d *Copy) SumParentPath() string {
	return strings.Join(d.SrcParentPath, ".") + ":" + strings.Join(d.DestParentPath, ".")
}

func (d *Copy) Gen() {
	has, fn := d.GenFn(d.FnName(), d.Src.Type.TypeAsJen(), d.Dest.Type.TypeAsJen())
	if has {
		return
	}
	bind := make(jen.Statement, 0)
	bind = append(bind, jen.Comment("basic ="))
	bind = append(bind, d.GenBasic()...)
	bind = append(bind,jen.Comment("slice = "))
	bind = append(bind, d.GenSlice()...)
	bind = append(bind, jen.Comment("map = "))
	bind = append(bind, d.GenMap()...)
	bind = append(bind, jen.Comment("pointer = "))
	bind = append(bind, d.GenPointer()...)
	bind = append(bind, jen.Return())

	fn.Block(bind...)
	//d.JenF.Add(d.Doc())
	d.JenF.Add(fn)
	for _, v := range d.Nest {
		v.Gen()
	}
}

func (d *Copy) GenExtraCopyMethod(bind *jen.Statement, destV, srcV Field) (has bool) {
	pkgName, methodName := destV.CopyMethod()
	if pkgName == "" && methodName == "" {
		return false
	}

	bind.Add(destV.DestIdPath().Op("=").Add(jen.Qual(pkgName, methodName).Call(srcV.SrcIdPath())))
	return true

}

func (d *Copy) GenBasic() jen.Statement {
	bind := make(jen.Statement, 0)
	for _, v := range d.Dest.BasicMap {
		if v.Name == "Cabinets" {
			fmt.Println("find basic cabinets: ", v.DestIdPath().String(),v.Type.T.String())
		}
		srcV,ok := d.Src.BasicMap[v.Name]
		if !ok {
			fmt.Printf("not found %s in %s\n", v.Name, d.SumPath())
			continue
		}

		if v.Doc != nil {
			bind.Add(jen.Comment(v.Doc.Text()))
		}

		if d.GenExtraCopyMethod(&bind, v, srcV) {
			continue
		}
		//dtoMethod := v.CopyMethod()
		//if dtoMethod != nil {
		//	bind.Add(v.DestIdPath().Op("=").Add(dtoMethod.Call(srcV.SrcIdPath())))
		//	continue
		//}
		fmt.Println("xtype", "ttype", "basic", "name", v.Name, "id", v.Type.ID(), "unescapedid", v.Type.UnescapedID(), "jen", v.Type.TypeAsJen().GoString())
		bind = append(bind, jen.Comment("basic = "+v.Name))
		bind = append(bind ,jen.Comment(strings.Join(v.Path, ".")))
		bind = append(bind, jen.Comment(v.SrcIdPath().GoString()))
		bind = append(bind, jen.Comment(v.DestIdPath().GoString()))

		bind.Add(v.DestIdPath().Op("=").Add(srcV.SrcIdPath()))
	}
	return bind
}

func (d *Copy) GenMap() jen.Statement {
	bind := make(jen.Statement, 0)
	for _, v := range d.Dest.MapMap {
		srcV,ok := d.Src.MapMap[v.Name]
		if !ok {
			fmt.Printf("not found %s in %s\n", v.Name, d.SumPath())
			continue
		}
		if v.Doc != nil {
			bind.Add(jen.Comment(v.Doc.Text()))
		}

		if d.GenExtraCopyMethod(&bind, v, srcV) {
			continue
		}

		fmt.Println("xtype", "ttype", "slice", "id", v.Type.ID(), v.Type.T.String(), "unescapedid", v.Type.UnescapedID(), "jen", v.Type.TypeAsJen().Render(os.Stdout))
		if v.Type.T.String() == srcV.Type.T.String() {
			bind.Add(v.DestIdPath().Op("=").Add(srcV.SrcIdPath()))
			continue
		}

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
			nestCopy := &Copy{
				JenF:           d.JenF,
				Recorder:       d.Recorder,
				SrcParentPath:  append(d.SrcParentPath, srcV.Path...),
				SrcPath:        append([]string{}, srcV.Path...),
				Src:            NewDataFieldMap(d.Pkg, []string{}, srcName, srcMapValue, srcMapValue.T),
				DestParentPath: append(d.DestParentPath, v.Path...),
				DestPath:       append([]string{}, v.Path...),
				Dest:           NewDataFieldMap(d.Pkg, []string{}, destName, destMapValue, destMapValue.T),
				Nest:           make([]*Copy, 0),
				StructName:     d.StructName,
			}
			d.Nest = append(d.Nest, nestCopy)

			block = v.DestIdPath().Index(jen.Id("key")).Op("=").Id("d." + nestCopy.FnName()).Call(jen.Id("value"))
		}
		bind.Add(jen.For(
			jen.List(jen.Id("key"), jen.Id("value")).
				Op(":=").Range().Add(srcV.SrcIdPath()).Block(
				block,
			)))
	}
	return bind
}

func (d *Copy) GenPointer() jen.Statement {
	bind := make(jen.Statement, 0)
	for _, v := range d.Dest.PointerMap {
		srcV,ok := d.Src.PointerMap[v.Name]
		if !ok {
			fmt.Printf("not found %s in %s\n", v.Name, d.SumPath())
			continue
		}

		if v.Doc != nil {
			bind.Add(jen.Comment(v.Doc.Text()))
		}

		if d.GenExtraCopyMethod(&bind, v, srcV) {
			continue
		}

		if v.Type.T.String() == srcV.Type.T.String() {
			bind.Add(v.DestIdPath().Op("=").Add(srcV.SrcIdPath()))
			continue
		}
		if v.Type.PointerInner.Basic {
			bind.Add(v.DestIdPath().Op("=").Add(srcV.SrcIdPath()))
		} else {
			srcLiner := srcV.Type.PointerInner
			destLiner := v.Type.PointerInner
			srcName := srcV.FieldName(d.SumPath())
			destName := v.FieldName(d.SumPath())
			//destName := srcLiner.HashID(d.SumPath())
			nestCopy := &Copy{
				JenF:           d.JenF,
				Recorder:       d.Recorder,
				SrcParentPath:  append(d.SrcParentPath, srcV.Path...),
				SrcPath:        append([]string{}, srcV.Path...),
				Src:            NewDataFieldMap(d.Pkg, []string{}, srcName, srcLiner, srcLiner.T),
				DestParentPath: append(d.DestParentPath, v.Path...),
				DestPath:       append([]string{}, v.Path...),
				Dest:           NewDataFieldMap(d.Pkg, []string{}, destName, srcLiner, destLiner.T),
				Nest:           make([]*Copy, 0),
				StructName:     d.StructName,
			}
			d.Nest = append(d.Nest, nestCopy)

			bind.Add(
				jen.If(srcV.SrcIdPath().Op("!=").Nil()).Block(
					jen.Id("v").Op(":=").Id("d."+nestCopy.FnName()).Call(jen.Id("*").Add(srcV.SrcIdPath())),
					v.DestIdPath().Op("=").Id("&v"),
				).Else().Block(
					v.DestIdPath().Op("=").Add(srcV.SrcIdPath()),
				),
			)
		}
	}
	return bind
}

func (d *Copy) GenSlice() jen.Statement {
	bind := make(jen.Statement, 0)
	for _, v := range d.Dest.SliceMap {
		srcV,ok := d.Src.SliceMap[v.Name]
		if v.Name == "Cabinets" {
			fmt.Println("find genCabinets: ", v.DestIdPath().GoString(),v.Type.T.String(),  "find srcv: ", ok)
		}
		if !ok {
			fmt.Printf("not found %s in %s\n", v.Name, d.SumPath())
			continue
		}
		if v.Doc != nil {
			bind.Add(jen.Comment(v.Doc.Text()))
		}
		//fmt.Println("xtype", "ttype", "slice", "id", v.Type.ID(), "unescapedid", v.Type.UnescapedID(), "jen", v.Type.TypeAsJen().Render(os.Stdout))

		if d.GenExtraCopyMethod(&bind, v, srcV) {
			continue
		}

		if v.Type.T.String() == srcV.Type.T.String() {
			bind.Add(v.DestIdPath().Op("=").Add(srcV.SrcIdPath()))
			continue
		}
		bind.Add(v.DestIdPath().Op("=").Make(v.Type.TypeAsJen(), jen.Id("0"), jen.Id("len").Call(srcV.SrcIdPath())))
		block := v.DestIdPath().Index(jen.Id("i")).Op("=").Add(srcV.SrcIdPath()).Index(jen.Id("i"))
		if !v.Type.ListInner.Basic {
			srcLiner := srcV.Type.ListInner
			destLiner := v.Type.ListInner
			//fmt.Println("listInner", srcLiner.TypeAsJen().GoString())
			//srcName := srcLiner.HashID(d.SumPath())
			srcName := srcV.FieldName(d.SumPath())
			destName := v.FieldName(d.SumPath())
			//destName := srcLiner.HashID(d.SumPath())
			nestCopy := &Copy{
				JenF:          d.JenF,
				Recorder:      d.Recorder,
				SrcParentPath: append(d.SrcParentPath, srcV.Path...),
				//SrcPath:  append([]string{}, srcV.Path...),
				SrcPath:        d.SrcPath[0:],
				Src:            NewDataFieldMap(d.Pkg, []string{}, srcName, srcLiner, srcLiner.T),
				DestParentPath: append(d.DestParentPath, v.Path...),
				//DestPath: append([]string{}, v.Path...),
				DestPath:   d.DestPath[0:],
				Dest:       NewDataFieldMap(d.Pkg, []string{}, destName, destLiner, destLiner.T),
				Nest:       make([]*Copy, 0),
				StructName: d.StructName,
			}
			d.Nest = append(d.Nest, nestCopy)

			block = v.DestIdPath().Index(jen.Id("i")).Op("=").Id("d." + nestCopy.FnName()).Call(srcV.SrcIdPath().Index(jen.Id("i")))
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

func (d *Copy) GenFn(funcName string, srcId, destId jen.Code) (has bool, fn *jen.Statement) {
	if d.DefaultFn != nil {
		return false, d.DefaultFn
	}
	srcType := jen.Type().Id("src").Add(srcId)
	destType := jen.Type().Id("dest").Add(destId)

	funcKey := fmt.Sprintf("%s_%s_%s", funcName, srcType.GoString(), destType.GoString())
	//fmt.Printf("funcName: %s, srcpath: %#v, destpath %#v \n", funcName,srcType, destType)

	has = d.Recorder.Lookup(funcKey)
	if has {
		return has, nil
	}
	d.Recorder.Reg(funcKey)

	return false, jen.Func().Params(jen.Id("d").Id(d.StructName)).
		Id(funcName).Params(jen.Id("src").Add(srcId)).Params(jen.Id("dest").Add(destId))
}
