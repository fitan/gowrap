package generator

import (
	"github.com/dave/jennifer/jen"
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
	Dir string
	Pkg *packages.Package
}

// 最后n目录转换为 dirName.dirName
func (g GenOption) CutLast2DirName() string {
	return strings.Join(filepath.SplitList(g.Dir)[len(filepath.SplitList(g.Dir))-2:len(filepath.SplitList(g.Dir))-1],".")
}

type Gen struct {
	GenFn
	GenImpl
	GenType
}

func NewGen(option GenOption)  {
	fn := NewGenFn(option)
	fn.AddPlug(NewGenFnCopy(fn))


}