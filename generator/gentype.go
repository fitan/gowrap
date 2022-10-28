package generator

import (
	"github.com/fitan/jennifer/jen"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
)

const GenTypeMark = "@type"

type GenType struct {
	Pkg      *packages.Package
	TypeList map[string]Type
	Plugs    map[string]GenPlug
}

func (g *GenType) AddPlug(plug GenPlug) {
	g.Plugs[plug.Name()] = plug
}

func (g *GenType) Run() error {
	g.parse()
	return nil
}

type TypePlug interface {
	Name() string
	Gen(pkg *packages.Package, name string, t Type) error
	JenF(name string) *jen.File
}

type Type struct {
	Doc *ast.CommentGroup
	T types.Type
}

func (g *GenType) GetFile(plugName, jenFName string) string {
	f := g.Plugs[plugName].JenF(jenFName)
	return f.GoString()
}



func (g *GenType) parse() {
	for _, v := range g.Pkg.Syntax {
		ast.Inspect(v, func(node ast.Node) bool {
			if genDecl, ok := node.(*ast.GenDecl);ok {
				var t Type

				format := AstDocFormat{doc: genDecl.Doc}
				if !format.ContainsMark(GenTypeMark) {
					return false
				}

				typeSpec, ok := genDecl.Specs[0].(*ast.TypeSpec)
				if ok {
					t.T = g.Pkg.TypesInfo.TypeOf(typeSpec.Type)
					g.TypeList[typeSpec.Name.Name] = t
					return true
				}
			}

			return true
		})
	}
}

