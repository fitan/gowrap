package generator

import (
	"fmt"
	"github.com/fitan/gowrap/xtype"
	"github.com/fitan/jennifer/jen"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/ast/astutil"
)

const GenCallMark = "@call"

type GenCall struct {
	GenOption GenOption
	FuncList  map[string]Func
	Plugs     map[string]GenPlug
}

func NewGenCall(genOption GenOption) *GenCall {
	m := make(map[string]GenPlug)
	return &GenCall{GenOption: genOption, FuncList: map[string]Func{}, Plugs: m}
}

type Func struct {
	Args []*xtype.Type
	Lhs  []*xtype.Type
	Doc  *ast.CommentGroup
	Name string
}

type XType struct {
	T  types.Type
	Id string
}

func (g *GenCall) GetFile(plugName, jenFName string) string {
	f := g.Plugs[plugName].JenF(jenFName)
	return f.GoString()
}

func (g *GenCall) JenF(name string) *jen.File {
	return g.Plugs[name].JenF(name)
}

func (g *GenCall) AddPlug(plug GenPlug) {
	g.Plugs[plug.Name()] = plug
}

func (g *GenCall) Run() error {
	g.parse()
	for _, plug := range g.Plugs {
		err := plug.Gen()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GenCall) parse() {
	for _, v := range g.GenOption.Pkg.Syntax {
		astutil.Apply(v, func(c *astutil.Cursor) bool {
			if call, ok := c.Node().(*ast.CallExpr); ok {
				comment := GetCommentByTokenPos(g.GenOption.Pkg, call.Pos())
				var fn Func

				format := AstDocFormat{comment}
				if !format.ContainsMark(GenCallMark) {
					return true
				}

				callName := call.Fun.(*ast.Ident).Name
				fn.Name = callName
				fn.Doc = comment

				if as, ok := c.Parent().(*ast.AssignStmt); ok {
					if as.Tok.String() == "=" || as.Tok.String() == ":=" {
						for _, param := range as.Lhs {
							fn.Lhs = append(fn.Lhs, xtype.TypeOf(g.GenOption.Pkg.TypesInfo.TypeOf(param)))
						}
					} else {
						panic(fmt.Sprintf("fn %s tok must be = or :=", callName))
					}
				} else {
					panic(fmt.Sprintf("fn %s must be assignStmt", callName))
				}

				for _, param := range call.Args {
					fn.Args = append(fn.Args, xtype.TypeOf(g.GenOption.Pkg.TypesInfo.TypeOf(param)))
				}

				g.FuncList[callName] = fn
				return false
			}
			return true
		}, func(c *astutil.Cursor) bool {
			return true
		})
	}
}
