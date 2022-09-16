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
	Plugs    []TypePlug
	JenF *jen.File
}

type TypePlug interface {
	Gen(jenF *jen.File,name string, t Type)
}

type Type struct {
	MarkParam []string
	T types.Type
}

func (g *GenType) Parse() {
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

