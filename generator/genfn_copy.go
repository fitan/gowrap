package generator

import (
	"fmt"
	"github.com/fitan/jennifer/jen"
	"github.com/fitan/gowrap/xtype"
	"golang.org/x/tools/go/packages"
	"strings"
)

const copyGenFName = "copy"

type GenFnCopy struct {
	recorder map[string]struct{}
	jenFM map[string]*jen.File
	genFn *GenFn
}

func NewGenFnCopy(fn *GenFn) *GenFnCopy {
	return &GenFnCopy{recorder: map[string]struct{}{}, genFn: fn, jenFM: map[string]*jen.File{}}
}

func (g *GenFnCopy) Name() string {
	return "copy"
}

func (g *GenFnCopy) JenF(name string) *jen.File {
	return g.jenFM[name]
}

func (g *GenFnCopy) Gen() error {

	jenF := jen.NewFile("copy")

	for fnName,fn := range g.genFn.FuncList {
		err := g.gen(jenF,g.genFn.GenOption.Pkg, fnName, fn)
		if err != nil {
			return err
		}
	}

	g.jenFM[copyGenFName] = jenF
	return nil
}

func (g *GenFnCopy) gen(jenF *jen.File, pkg *packages.Package , name string, fn Func) error {

	if !(len(fn.MarkParam) >0 && fn.MarkParam[0] == "copy") {
		return nil
	}

	if _,ok := g.recorder[name];ok {
		return nil
	}

	if len(fn.Args) != 1 {
		return fmt.Errorf("plug Copy: fn %s args count must be 1",name)
	}

	if len(fn.Lhs) != 1 {
		return fmt.Errorf("plug Copy: fn %s lhs count must be 1",name)
	}



	g.recorder[name] = struct{}{}
	objName := name + "Obj"

	srcType := fn.Args[0]
	destType := fn.Lhs[0]

	//destTypePoinit, ok := destType.(*types.Pointer);
	//if !ok {
	//	return errors.New("dest type must be pointer")
	//}

	//DestTypeElem := destTypePoinit.Elem()


	srcTypeString := TypeString(srcType.String())
	destTypeElemString := TypeString(destType.String())

	srcTypeID := strings.TrimPrefix(srcTypeString.ID(), TypeString(pkg.ID).ID()+".")
	destTypeID := strings.TrimPrefix(destTypeElemString.ID(), TypeString(pkg.Name).ID()+".")


	jenF.Func().Id(name).Params(jen.Id("src").Id(srcTypeID)).Params(jen.Id("dest").Id(destTypeID)).Block(
		jen.Id("dest").Op(":=").Id(objName).Block().Dot("Copy").Call(jen.Id("src")),
		jen.Return(),
	)

	jenF.Type().Id(objName).Struct()


	dto := Copy{
		Pkg: pkg,
		StructName: objName,
		JenF: jenF,
		Recorder: NewRecorder(),
		SrcParentPath: []string{},
		SrcPath: []string{},
		Src: NewDataFieldMap(pkg,[]string{}, objName, xtype.TypeOf(srcType), srcType),
		DestParentPath: []string{},
		DestPath: []string{},
		Dest: NewDataFieldMap(pkg,[]string{}, objName, xtype.TypeOf(destType), destType),
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