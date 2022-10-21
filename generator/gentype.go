package generator

import (
	"github.com/dave/jennifer/jen"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"strings"
)

const GenTypeMark = "// @type"

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
	MarkParam []string
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
				if genDecl.Doc == nil {
					return false
				}

				for _, l := range genDecl.Doc.List {
					if strings.HasPrefix(DocFormat(l.Text), GenTypeMark) {
						t.MarkParam = strings.Fields(strings.TrimPrefix(DocFormat(l.Text),GenTypeMark))

						typeSpec, ok := genDecl.Specs[0].(*ast.TypeSpec)
						if ok {
							t.T = g.Pkg.TypesInfo.TypeOf(typeSpec.Type)
							g.TypeList[typeSpec.Name.Name] = t
							return true
						}
						panic("not type: " + genDecl.Doc.Text())
					}
				}
			}

			return true
		})
	}
}

