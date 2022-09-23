package generator

import (
	"errors"
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/fitan/gowrap/xtype"
	"go/types"
	"golang.org/x/tools/go/packages"
	"strings"
)

type GenFnCopy struct {
	recorder map[string]struct{}
	jenF *jen.File
}

func NewGenFnCopy() *GenFnCopy {
	return &GenFnCopy{recorder: map[string]struct{}{}, jenF: jen.NewFile("Copy")}
}

func (g *GenFnCopy) Name() string {
	return "copy"
}

func (g *GenFnCopy) JenF() *jen.File {
	return g.jenF
}

func (g *GenFnCopy) Gen(pkg *packages.Package , name string, fn Func) error {

	if !(len(fn.MarkParam) >0 && fn.MarkParam[0] == "copy") {
		return nil
	}

	if _,ok := g.recorder[name];ok {
		return nil
	}

	if len(fn.Args) != 2 {
		return fmt.Errorf("plug Copy: fn %s args count must be 2",name)
	}



	g.recorder[name] = struct{}{}
	objName := name + "Obj"

	srcType := fn.Args[0]
	destType := fn.Args[1]

	destTypePoinit, ok := destType.(*types.Pointer);
	if !ok {
		return errors.New("dest type must be pointer")
	}

	DestTypeElem := destTypePoinit.Elem()


	srcTypeString := TypeString(srcType.String())
	destTypeElemString := TypeString(destTypePoinit.String())

	fmt.Println("srcTypeString:",srcTypeString.PkgPath(), srcTypeString.ID())
	fmt.Println("destTypeElemString:",destTypeElemString.PkgPath(), destTypeElemString.ID())
	fmt.Println("pkg.Name",pkg.Name, "pkg.id", pkg.ID, "pkg.PkgPath", pkg.PkgPath)


	srcTypeID := strings.TrimPrefix(srcTypeString.ID(), TypeString(pkg.ID).ID()+".")
	destTypeID := strings.TrimPrefix(destTypeElemString.ID(), TypeString(pkg.Name).ID()+".")



	g.jenF.Func().Id(name).Params(jen.Id("src").Id(srcTypeID), jen.Id("dest").Id("*"+destTypeID)).Block(
		jen.Id("v").Op(":=").Id(objName).Block().Dot("Copy").Call(jen.Id("src")),
		jen.Id("dest").Op("=").Id("&v"),
	)

	g.jenF.Type().Id(objName).Struct()


	dto := Copy{
		Pkg: pkg,
		StructName: objName,
		JenF: g.jenF,
		Recorder: NewRecorder(),
		SrcParentPath: []string{},
		SrcPath: []string{},
		Src: NewDataFieldMap(pkg,[]string{}, objName, xtype.TypeOf(srcType), srcType),
		DestParentPath: []string{},
		DestPath: []string{},
		Dest: NewDataFieldMap(pkg,[]string{}, objName, xtype.TypeOf(DestTypeElem), DestTypeElem),
		DefaultFn: jen.Func().Params(jen.Id("d").Id(objName)).
			Id("Copy").Params(jen.Id("src").Id(srcTypeID)).Params(jen.Id("dest").Id(destTypeID)),
	}
	dto.Gen()
	return nil
}


type TypeString string

func (t TypeString) PkgPath() string {
	split := strings.Split(string(t), "/")
	pathSplit := split[0: len(split)-1]
	return strings.Join(pathSplit, "/")
}

// types.Name
func (t TypeString) ID() string {
	split := strings.Split(string(t), "/")
	return split[len(split)-1]
}