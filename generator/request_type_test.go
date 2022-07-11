package generator

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"go/types"
	"golang.org/x/tools/go/packages"
	"testing"
)

const mode packages.LoadMode =packages.NeedName |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo |
	packages.NeedImports |
	packages.NeedModule |
	packages.NeedExportsFile |
	packages.NeedTypesSizes |
	packages.NeedDeps |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles

type Pkgs struct {
	pkg *packages.Package
}

func (p *Pkgs) GetRequestType(name string) types.Type {
	fmt.Println(p.pkg.Types.Scope().Names())
	lookup := p.pkg.Types.Scope().Lookup(name)
	return lookup.Type()
}

func LoadPkgs() *Pkgs {
	cfg := &packages.Config{Mode: mode}
	pkgs, err := packages.Load(cfg, "./test_data")
	if err != nil {
		panic(err.Error())
	}
	return &Pkgs{pkg: pkgs[0]}
}

func TestKitRequest_RequestType(t *testing.T) {

	pkg := LoadPkgs()

	type args struct {
		prefix              []string
		requestName         string
		requestType         types.Type
		requestParamTagType string
	}
	tests := []struct {
		name   string
		args   args
	}{
		{name: "HelloRequest", args: args{
			prefix:              []string{},
			requestName:         "",
			requestType:         pkg.GetRequestType("HelloRequest"),
			requestParamTagType: "",
		}},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				k := NewKitRequest()
				k.RequestType(tt.args.prefix, tt.args.requestName, tt.args.requestType, tt.args.requestParamTagType)
				spew.Dump(k)
				fmt.Println(k.BindPathParam())
			},
		)
	}
}
