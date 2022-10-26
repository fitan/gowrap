package generator

import (
	"github.com/fitan/jennifer/jen"
	"golang.org/x/tools/go/packages"
)

type GenTypeGorm struct {
	recorder map[string]struct{}
	jenF *jen.File
}

func NewGenTypeGorm() *GenTypeGorm {
	return &GenTypeGorm{recorder: map[string]struct{}{}, jenF: jen.NewFile("gorm")}
}

func (g *GenTypeGorm) Name() string {
	return "gorm"
}

func (g *GenTypeGorm) Gen(pkg *packages.Package, name string, t Type) error {
	if _, ok := g.recorder[name]; ok {
		return nil
	}
	g.recorder[name] = struct{}{}

	g.jenF.Type().Id(name).StructFunc(func(group *jen.Group) {
		group.Id("gorm.Model")
	})
	return nil
}

func (g *GenTypeGorm) JenF() *jen.File {
	return g.jenF
}


