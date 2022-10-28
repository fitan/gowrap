package generator

import (
	"fmt"
	"golang.org/x/tools/go/packages"
	"testing"
)

func TestGenFn_Parse(t *testing.T) {
	pkg := LoadPkgs()

	type fields struct {
		Pkg   *packages.Package
		Funcs map[string]Func
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "f1",
			fields: fields{
				Pkg:   pkg.pkg,
				Funcs: make(map[string]Func,0),
			},
		},

	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				g := &GenFn{
					Pkg:      tt.fields.Pkg,
					FuncList: tt.fields.Funcs,
				}

				g.Parse()

				fmt.Println(g.FuncList)
			},
		)
	}
}
