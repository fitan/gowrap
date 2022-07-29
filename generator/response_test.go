package generator

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"go/types"
	"golang.org/x/tools/go/packages"
	"testing"
)

func TestDTO_Parse(t *testing.T) {
	pkg := LoadPkgs()
	obj := pkg.pkg.Types.Scope().Lookup("HelloRequest")

	type fields struct {
		pkg        *packages.Package
		namedMap   map[string]Field
		pointerMap map[string]Field
		sliceMap   map[string]Field
		mapMap     map[string]Field
		structMap  map[string]Field
		basicMap   map[string]Field
	}
	type args struct {
		prefix []string
		name   string
		t      types.Type
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{name: "HelloRequest", fields: fields{
			pkg:        pkg.pkg,
			namedMap:   map[string]Field{},
			pointerMap: map[string]Field{},
			sliceMap:   map[string]Field{},
			mapMap:     map[string]Field{},
			structMap:  map[string]Field{},
			basicMap:   map[string]Field{},
		}, args: args{
			prefix: []string{},
			name:   "HelloRequest",
			t:      obj.Type().Underlying().(*types.Struct),
		}},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				jenF := jen.NewFile("DTO")
				dto := DTO{
					JenF:           jenF,
					ParentSrcPath:  tt.args.prefix,
					Src:            NewDataFieldMap(tt.args.prefix,tt.name, tt.args.name, tt.args.t),
					ParentDestPath: tt.args.prefix,
					Dest:           NewDataFieldMap(tt.args.prefix,tt.name, tt.args.name, tt.args.t),
				}
				dto.Gen()
				fmt.Println(jenF.GoString())
			},
		)
	}
}
