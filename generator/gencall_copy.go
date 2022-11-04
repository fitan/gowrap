package generator

import (
	"fmt"
	"github.com/fitan/gowrap/xtype"
	"github.com/fitan/jennifer/jen"
	"golang.org/x/tools/go/packages"
	"strings"
)

const copyGenFName = "copy"

type GenCallCopy struct {
	recorder map[string]struct{}
	jenFM    map[string]*jen.File
	genCall  *GenCall
}

func NewGenCallCopy(call *GenCall) *GenCallCopy {
	return &GenCallCopy{recorder: map[string]struct{}{}, genCall: call, jenFM: map[string]*jen.File{}}
}

func (g *GenCallCopy) Name() string {
	return "copy"
}

func (g *GenCallCopy) JenF(name string) *jen.File {
	return g.jenFM[name]
}

func (g *GenCallCopy) Gen() error {

	jenF := jen.NewFile(g.genCall.GenOption.Pkg.Name)

	for callName, call := range g.genCall.FuncList {
		err := g.gen(jenF, g.genCall.GenOption.Pkg, callName, call)
		if err != nil {
			return err
		}
	}

	g.jenFM[copyGenFName] = jenF
	return nil
}

func (g *GenCallCopy) gen(jenF *jen.File, pkg *packages.Package, name string, call Func) error {

	var copyMark string

	format := &AstDocFormat{call.Doc}
	format.MarkValuesMapping(GenCallMark, &copyMark)
	if copyMark != copyGenFName {
		return nil
	}

	if _, ok := g.recorder[name]; ok {
		return nil
	}

	if len(call.Args) != 1 {
		return fmt.Errorf("plug Copy: call %s args count must be 1", name)
	}

	if len(call.Lhs) != 1 {
		return fmt.Errorf("plug Copy: call %s lhs count must be 1", name)
	}

	g.recorder[name] = struct{}{}
	objName := name + "Obj"

	srcType := call.Args[0]
	destType := call.Lhs[0]

	//destTypePoinit, ok := destType.(*types.Pointer);
	//if !ok {
	//	return errors.New("dest type must be pointer")
	//}

	//DestTypeElem := destTypePoinit.Elem()

	jenF.Func().Id(name).Params(jen.Id("src").Id(srcType.TypeAsJenComparePkgName(pkg).GoString())).Params(jen.Id("dest").Id(destType.TypeAsJenComparePkgName(pkg).GoString())).Block(
		jen.Id("dest").Op("=").Id(objName).Block().Dot("Copy").Call(jen.Id("src")),
		jen.Return(),
	)

	jenF.Type().Id(objName).Struct()

	dto := Copy{
		Pkg:            pkg,
		StructName:     objName,
		JenF:           jenF,
		Recorder:       NewRecorder(),
		SrcParentPath:  []string{},
		SrcPath:        []string{},
		Src:            NewDataFieldMap(pkg, []string{}, objName, xtype.TypeOf(srcType.T), srcType.T),
		DestParentPath: []string{},
		DestPath:       []string{},
		Dest:           NewDataFieldMap(pkg, []string{}, objName, xtype.TypeOf(destType.T), destType.T),
		DefaultFn: jen.Func().Params(jen.Id("d").Id(objName)).
			Id("Copy").Params(jen.Id("src").Id(srcType.TypeAsJenComparePkgName(pkg).GoString())).Params(jen.Id("dest").Id(destType.TypeAsJenComparePkgName(pkg).GoString())),
	}
	dto.Gen()
	return nil
}

type TypeString string

func (t TypeString) PkgPath() string {
	split := strings.Split(string(t), "/")
	pathSplit := split[0 : len(split)-1]
	return strings.Join(pathSplit, "/")
}

// types.Name
func (t TypeString) ID() string {
	split := strings.Split(string(t), "/")
	return split[len(split)-1]
}
