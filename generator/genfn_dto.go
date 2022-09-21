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

type GenFnDTO struct {
	recorder map[string]struct{}
	jenF *jen.File
}

func NewGenFnDTO() *GenFnDTO {
	return &GenFnDTO{recorder: map[string]struct{}{}, jenF: jen.NewFile("DTO")}
}

func (g *GenFnDTO) Name() string {
	return "dto"
}

func (g *GenFnDTO) JenF() *jen.File {
	return g.jenF
}

func (g *GenFnDTO) Gen(pkg *packages.Package , name string, fn Func) error {

	if !(len(fn.MarkParam) >0 && fn.MarkParam[0] == "dto") {
		return nil
	}

	if _,ok := g.recorder[name];ok {
		return nil
	}

	if len(fn.Args) != 2 {
		return fmt.Errorf("plug DTO: fn %s args count must be 2",name)
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
	fmt.Println("underlyingdesttype: ", DestTypeElem.String())

	srcTypeID := strings.TrimPrefix(strings.TrimPrefix(srcType.String(), pkg.PkgPath), ".")
	destTypeID := strings.TrimPrefix(strings.TrimPrefix(DestTypeElem.String(), pkg.PkgPath),".")



	g.jenF.Func().Id(name).Params(jen.Id("src").Id(srcTypeID), jen.Id("dest").Id("*"+destTypeID)).Block(
		jen.Id("v").Op(":=").Id(objName).Block().Dot("DTO").Call(jen.Id("src")),
		jen.Id("dest").Op("=").Id("&v"),
	)

	g.jenF.Type().Id(objName).Struct()


	dto := DTO{
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
			Id("DTO").Params(jen.Id("src").Id(srcTypeID)).Params(jen.Id("dest").Id(destTypeID)),
	}
	dto.Gen()
	return nil
}

