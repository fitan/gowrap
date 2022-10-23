package generator

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/ast/astutil"
	"strings"
)

const GenFnMark = "// @fn"

type GenFn struct {
	GenOption GenOption
	FuncList  map[string]Func
	Plugs     map[string]GenPlug
}

func NewGenFn(genOption GenOption) *GenFn {
	m := make(map[string]GenPlug)
	return &GenFn{GenOption: genOption, FuncList: map[string]Func{}, Plugs: m}
}

type Func struct {
	MarkParam []string
	Args      []types.Type
	Lhs       []types.Type
}

func (g *GenFn) GetFile(plugName, jenFName string) string {
	f := g.Plugs[plugName].JenF(jenFName)
	return f.GoString()
}

func (g *GenFn) JenF(name string) *jen.File {
	return g.Plugs[name].JenF(name)
}

func (g *GenFn) AddPlug(plug GenPlug) {
	g.Plugs[plug.Name()] = plug
}

func (g *GenFn) Run() error {
	g.parse()
	for _, plug := range g.Plugs {
		err := plug.Gen()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GenFn) parse() {
	for _, v := range g.GenOption.Pkg.Syntax {
		astutil.Apply(v, func(c *astutil.Cursor) bool {
			if call, ok := c.Node().(*ast.CallExpr); ok {
				comment := GetCommentByTokenPos(g.GenOption.Pkg, call.Pos())
				var fn Func
				if comment == nil {
					return false
				}
				for _, l := range comment.List {
					if strings.HasPrefix(DocFormat(l.Text), GenFnMark) {
						fn.MarkParam = strings.Fields(strings.TrimPrefix(DocFormat(l.Text), GenFnMark))
						break
					}

					return true
				}

				fnName := call.Fun.(*ast.Ident).Name

				if as, ok := c.Parent().(*ast.AssignStmt); ok {
					if as.Tok.String() == "=" {
						for _, l := range as.Lhs {
							fn.Lhs = append(fn.Lhs, g.GenOption.Pkg.TypesInfo.TypeOf(l))
						}
					} else {
						panic(fmt.Sprintf("fn %s tok must be =", fnName))
					}
				} else {
					panic(fmt.Sprintf("fn %s must be assignStmt", fnName))
				}

				for _, arg := range call.Args {
					fn.Args = append(fn.Args, g.GenOption.Pkg.TypesInfo.TypeOf(arg))
				}

				g.FuncList[fnName] = fn
				return false
			}
			return true
		}, func(c *astutil.Cursor) bool {
			return true
		})
	}
}
