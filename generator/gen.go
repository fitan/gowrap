package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
	"path/filepath"
	"strings"
)

type GenPlug interface {
	Name() string
	Gen() error
	JenF(name string) *jen.File
}

type GenOption struct {
	// 当前目录
	Pkg *packages.Package
}

// 最后n目录转换为 dirName.dirName
func (g GenOption) CutLast2DirName() string {
	return strings.Join(filepath.SplitList(g.Pkg.PkgPath)[len(filepath.SplitList(g.Pkg.PkgPath))-2:len(filepath.SplitList(g.Pkg.PkgPath))-1],".")
}

type Gen struct {
	GenFn
	GenImpl
	GenType
}

func NewGen(option GenOption) error {
	fn := NewGenFn(option)
	fn.AddPlug(NewGenFnCopy(fn))
	err := fn.Run()
	if err != nil {
		return errors.Wrap(err, "gen fn run")
	}

	impl := NewGenImpl(option)
	impl.AddPlug(NewGenImplKitHttp(impl))
	err = impl.Run()
	if err != nil {
		return errors.Wrap(err, "gen impl run")
	}

	return nil
}