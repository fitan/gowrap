package generator

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"strings"
)

const GenFnMark = "// @fn"

type GenFn struct {
	Pkg      *packages.Package
	FuncList map[string]Func
	Plugs    []FnPlug
	JenF *jen.File
}

func NewGenFn(pkg *packages.Package, jenF *jen.File, plugs ...FnPlug) *GenFn {
	return &GenFn{Pkg: pkg, FuncList: map[string]Func{}, Plugs: plugs, JenF: jenF}
}

type FnPlug interface {
	Name() string
	Gen(pkg *packages.Package, name string, fn Func) error
	JenF() *jen.File
}

type Func struct {
	MarkParam []string
	Args []types.Type
}

func (g *GenFn) JenFile() *jen.File {
	return g.JenF
}

func (g *GenFn) Run() {
	for name, fn := range g.FuncList {
		for _,plug := range g.Plugs {
			fmt.Println("plug name:",name, fn)
			err := plug.Gen(g.Pkg, name, fn)
			if err != nil {
				panic(err)
			}

		}
	}
}

func (g *GenFn) Parse() {
	for _, v  := range g.Pkg.Syntax {
		ast.Inspect(v, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				var fn Func
				comment := GetCommentByTokenPos(g.Pkg, call.Pos())
				if comment == nil {
					return false
				}
				for _, l := range comment.List {
					if strings.HasPrefix(DocFormat(l.Text), GenFnMark) {
						fn.MarkParam = strings.Fields(strings.TrimPrefix(DocFormat(l.Text),GenFnMark))
						break
					}

					return true
				}

				fnName := call.Fun.(*ast.Ident).Name
				for _,arg :=range call.Args {
					fn.Args = append(fn.Args,g.Pkg.TypesInfo.TypeOf(arg))
				}

				g.FuncList[fnName] = fn
				return false
			}

			return true
		})
	}
}