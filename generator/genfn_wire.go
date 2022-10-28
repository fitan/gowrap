package generator

import (
	"fmt"
	"github.com/fitan/jennifer/jen"
)

const wireGenFName = "wire"

type GenFnWire struct {
	recorder map[string]struct{}
	jenFM map[string]*jen.File
	genFn *GenFn
}

func (g *GenFnWire) Name() string {
	return "wire"
}

func (g *GenFnWire) Gen() error {
	jenF := jen.NewFile(g.genFn.GenOption.Pkg.Name)
	jenF.HeaderComment(" +build wireinject")
	mapSet := make(map[string][]string, 0)

	var wireMark, wireType, wireName string

	for fnName,fn := range g.genFn.FuncList {
		docFormat := AstDocFormat{fn.Doc}
		docFormat.MarkValuesMapping(GenFnMark, &wireMark, &wireType, &wireName)
		if wireMark != wireGenFName || wireType != "set" {
			continue
		}

		if wireName == "" {
			wireName = "Default"
		}

		if _, ok := mapSet[wireName];ok {
			mapSet[wireName] = append(mapSet[wireName], fnName)
		} else {
			mapSet[wireName] = []string{fnName}
		}
	}
	fmt.Println(mapSet)

	for setName, fnNames := range mapSet {
		jenF.Var().Id(setName + "Set").Op("=").Qual("github.com/google/wire", "NewSet").Call(jen.ListFunc(func(g *jen.Group) {
			for _, fnName := range fnNames {
				g.Id(fnName)
			}
		})).Line()
	}

	g.jenFM[wireGenFName] = jenF
	return nil
}

func (g *GenFnWire) JenF(name string) *jen.File {
	return g.jenFM[name]
}

func NewGenFnWire(fn *GenFn) *GenFnWire  {
	return &GenFnWire{
		recorder: map[string]struct{}{},
		jenFM: map[string]*jen.File{},
		genFn: fn,
	}
}

