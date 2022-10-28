package generator

import (
	"github.com/fitan/jennifer/jen"
	"go/ast"
)

const GenFnMark = "@fn"



type GenFn struct {
	GenOption GenOption
	FuncList  map[string]Func
	Plugs     map[string]GenPlug
}

func NewGenFn(genOption GenOption) *GenFn {
	m := make(map[string]GenPlug)
	return &GenFn{GenOption: genOption, FuncList: map[string]Func{}, Plugs: m}
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
		ast.Inspect(v, func(node ast.Node) bool {
			if fnDecl,ok := node.(*ast.FuncDecl);ok {
				var fn Func

				format := AstDocFormat{fnDecl.Doc}
				if !format.ContainsMark(GenFnMark) {
					return true
				}

				for _, param := range fnDecl.Type.Params.List {
					fn.Args = append(fn.Lhs,g.GenOption.Pkg.TypesInfo.TypeOf(param.Type))
				}
				for _, param := range fnDecl.Type.Results.List {
					fn.Lhs = append(fn.Args,g.GenOption.Pkg.TypesInfo.TypeOf(param.Type))
				}
				fn.Name = fnDecl.Name.Name
				fn.Doc = fnDecl.Doc

				g.FuncList[fn.Name] = fn
				return false
			}
			return true
		})

	}
}