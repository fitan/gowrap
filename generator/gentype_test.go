package generator

import (
	"fmt"
	"github.com/fitan/jennifer/jen"
	"golang.org/x/tools/go/packages"
	"testing"
)

func TestGenType_Parse(t *testing.T) {
	pkg := LoadPkgs()

	type fields struct {
		Pkg      *packages.Package
		TypeList map[string]Type
		Plugs    []TypePlug
		JenF     *jen.File
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "t1",
			fields: fields{
				Pkg:      pkg.pkg,
				TypeList: make(map[string]Type, 0),
				Plugs:    make([]TypePlug,0,0),
				JenF:     nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				g := &GenType{
					Pkg:      tt.fields.Pkg,
					TypeList: tt.fields.TypeList,
					Plugs:    tt.fields.Plugs,
					JenF:     tt.fields.JenF,
				}
				g.Parse()
				fmt.Println(g.TypeList)
			},
		)
	}
}
