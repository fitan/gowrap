package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/davecgh/go-spew/spew"
	"golang.org/x/tools/go/packages"
	"testing"
)

func TestGenImpl_Parse(t *testing.T) {
	pkg := LoadPkgs()

	type fields struct {
		Pkg      *packages.Package
		ImplList map[string]Impl
		Plugs    []ImplPlug
		JenF     *jen.File
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "f1",
			fields: fields{
				Pkg:      pkg.pkg,
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				g := NewGenImpl(tt.fields.Pkg)
				g.Parse()
				spew.Dump(g.ImplList)
			},
		)
	}
}
