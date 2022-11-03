package generator

import (
	"fmt"
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
	Args []XType
	Lhs  []XType
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

							var id string
							var t types.Type
							t = g.GenOption.Pkg.TypesInfo.TypeOf(param)
							for k, vv := range g.GenOption.Pkg.TypesInfo.Types {
								if t.String() == vv.Type.String() {
									fmt.Println("string == string: ", Node2String(g.GenOption.Pkg.Fset, k))
								}
								if t.String() == vv.Type.String() && vv.IsType() {
									id = Node2String(g.GenOption.Pkg.Fset, k)
									break
								}
							}

							fn.Lhs = append(fn.Lhs, XType{
								T:  t,
								Id: id,
							})
						}
					} else {
						panic(fmt.Sprintf("fn %s tok must be = or :=", callName))
					}
				} else {
					panic(fmt.Sprintf("fn %s must be assignStmt", callName))
				}

				for _, param := range call.Args {
					var id string
					var t types.Type
					t = g.GenOption.Pkg.TypesInfo.TypeOf(param)

					id = type2RawTypeId(g.GenOption.Pkg, t, "")
					//for k, vv := range g.GenOption.Pkg.TypesInfo.Types {
					//	if t.String() == vv.Type.String() {
					//		fmt.Println("arg == arg: ", Node2String(g.GenOption.Pkg.Fset,k))
					//	}
					//fmt.Println("arg == arg: ", Node2String(g.GenOption.Pkg.Fset,k))
					//if t.String() == vv.Type.String() && vv.IsType() {
					//	id =  Node2String(g.GenOption.Pkg.Fset, k)
					//	break
					//}
					//}

					fn.Args = append(fn.Args, XType{
						T:  t,
						Id: id,
					})
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
