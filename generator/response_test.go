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
			name:   "",
			t:      obj.Type().Underlying().(*types.Struct),
		}},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				src := &DataFieldMap{
					Name:       "Src",
					TypeID:     obj.Id(),
					NamedMap:   map[string]Field{},
					PointerMap: map[string]Field{},
					SliceMap:   map[string]Field{},
					MapMap:     map[string]Field{},
					StructMap:  map[string]Field{},
					BasicMap:   map[string]Field{},
				}
				src.Parse(tt.args.prefix, tt.args.name, tt.args.t)

				dest := &DataFieldMap{
					Name:       "Dest",
					TypeID:     obj.Id(),
					NamedMap:   map[string]Field{},
					PointerMap: map[string]Field{},
					SliceMap:   map[string]Field{},
					MapMap:     map[string]Field{},
					StructMap:  map[string]Field{},
					BasicMap:   map[string]Field{},
				}
				dest.Parse(tt.args.prefix, tt.args.name, tt.args.t)

				jenF := jen.NewFile("DTO")
				dto := DTO{
					JenF: jenF,
					Src:  src,
					Dest: dest,
				}
				dto.GenBasic()
				dto.GenSlice()
				fmt.Println(jenF.GoString())
			},
		)
	}
}
