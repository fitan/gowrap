package generator

import (
	"github.com/fitan/jennifer/jen"
	"github.com/pkg/errors"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"log"
	"strings"
)

type GenPlug interface {
	Name() string
	Gen() error
	JenF(name string) *jen.File
}

type GenObjType interface {
	GetFile(plugName, jenFName string) string
	AddPlug(plug GenPlug)
	Run() error
}

type GenOption struct {
	// 当前目录
	Pkg *packages.Package
	// main.go 文件中默认引用的import
	Imports         []*ast.ImportSpec
	MainExtraImport [][]string
}

func (g GenOption) ImportByName(name string) (path string) {
	for _, i := range g.MainExtraImport {
		importName, imposrtPath := i[0], i[1]
		if importName == name {
			return imposrtPath
		}
	}
	return ""
}

func (g *GenOption) ExtraImports() {
	res := make([][]string, 0)

	for _, i := range g.Imports {
		docF := NewAstDocFormat(i.Doc)
		log.Printf("doc: %s", i.Doc.Text())
		var v1, v2 string
		docF.MarkValuesMapping("@extra", &v1, &v2)
		if v1 == "" && v2 == "" {
			continue
		}

		var pathName, path string
		if v1 != "" && v2 != "" {
			pathName = v1
			path = v2
			res = append(res, []string{pathName, path})
			continue
		}

		if v1 != "" {
			pathName = ""
			path = v1
			res = append(res, []string{pathName, path})
			continue
		}
	}
	g.MainExtraImport = res

	return
}

// 最后n目录转换为 dirName.dirName
func (g GenOption) CutLast2DirName() string {
	pathList := strings.Split(g.Pkg.PkgPath, "/")
	return strings.Join(pathList[len(pathList)-2:len(pathList)-1], ".")
}

type Gen struct {
	GenOption
	*GenFn
	*GenImpl
	*GenType
}

func NewGen(option GenOption) (Gen, error) {
	fn := NewGenFn(option)
	fn.AddPlug(NewGenFnWire(fn))
	err := fn.Run()
	if err != nil {
		return Gen{}, errors.Wrap(err, "gen fn run")
	}

	call := NewGenCall(option)
	call.AddPlug(NewGenCallCopy(call))
	call.AddPlug(NewGenCallQuery(call))
	err = call.Run()
	if err != nil {
		return Gen{}, errors.Wrap(err, "gen call run")
	}

	impl := NewGenImpl(option)
	impl.AddPlug(NewGenImplKitHttp(impl))
	err = impl.Run()
	if err != nil {
		return Gen{}, errors.Wrap(err, "gen impl run")
	}

	g := Gen{
		GenOption: option,
		GenFn:     fn,
		GenImpl:   impl,
		GenType:   nil,
	}
	return g, nil
}
