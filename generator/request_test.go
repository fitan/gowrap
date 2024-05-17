package generator

import (
	"fmt"
	"go/ast"
	"go/types"
	"testing"

	"github.com/fitan/gowrap/xtype"
	"github.com/fitan/jennifer/jen"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

const mode packages.LoadMode = packages.NeedName |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo |
	packages.NeedImports |
	packages.NeedModule |
	packages.NeedExportsFile |
	packages.NeedTypesSizes |
	packages.NeedDeps |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles

type Pkgs struct {
	pkg *packages.Package
}

func LoadPkgs() *Pkgs {
	cfg := &packages.Config{Mode: mode}
	pkgs, err := packages.Load(cfg, "./test_data")
	if err != nil {
		panic(err.Error())
	}
	return &Pkgs{pkg: pkgs[0]}
}

func TestKitRequest_RequestType(t *testing.T) {
	pkgs := LoadPkgs()
	for _, v := range pkgs.pkg.Syntax {
		astutil.Apply(v, func(c *astutil.Cursor) bool {
			switch t := c.Node().(type) {
			case *ast.GenDecl:
				for _, typeSpec := range t.Specs {
					switch typeSpec := typeSpec.(type) {
					case *ast.TypeSpec:
						switch typeSpec.Type.(type) {
						case *ast.StructType:
							fmt.Println("struct: ", typeSpec.Name.String())
							if typeSpec.Name.String() == "GenStruct" {
								defStruct := pkgs.pkg.TypesInfo.Defs[typeSpec.Name].Type().(*types.Named).Underlying().(*types.Struct)

								// fmt.Println(defT.Underlying().String())
								// fmt.Println("defT: ", defT.String())

								k := &KitResponse{Pkg: pkgs.pkg}
								file := jen.NewFile("test").Line()
								out := k.Parse(file, xtype.TypeOf(defStruct), "GenStruct")
								fmt.Println(out)
								break

							}

						}
					}
				}
			}
			return true
		}, func(c *astutil.Cursor) bool {
			return true
		})

	}
	//for _, v := range pkg.pkg.Syntax {
	//	fmt.Println("syntax file: ", v.Name)
	//	fmt.Println("syntax file: ", v.Name.Name)
	//	fmt.Println(pkg.pkg.Fset.Position(v.Pos()).Filename)
	//	for _, c :=range v.Comments {
	//		fieldLine := pkg.pkg.Fset.Position(c.End()).Line + 1
	//		fmt.Println("Comment:", c.Text(), "pos:", c.Pos(), "line:", pkg.pkg.Fset.Position(c.Pos()).Line, "end:", pkg.pkg.Fset.Position(c.End()).Line, "fieldLine:", fieldLine)
	//	}
	//
	//}

}
