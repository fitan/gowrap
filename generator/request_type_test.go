package generator

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"golang.org/x/tools/go/packages"
	"testing"
)

const mode packages.LoadMode = packages.NeedName |
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
	//for _, v := range pkg.pkg.Syntax {
	//	fmt.Println("syntax file: ", v.Name)
	//	fmt.Println("syntax file: ", v.Name.Name)
	//	fmt.Println(pkg.pkg.Fset.Position(v.Pos()).Filename)
	//	for _, c :=range v.Comments {
	//		fieldLine := pkg.pkg.Fset.Position(c.End()).Line + 1
	//		fmt.Println("Comment:", c.Text(), "pos:", c.Pos(), "line:", pkg.pkg.Fset.Position(c.Pos()).Line, "end:", pkg.pkg.Fset.Position(c.End()).Line, "fieldLine:", fieldLine)
	//	}
	//
	//}

	type args struct {
		pkg *packages.Package
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "HelloRequest", args: args{
			pkg: pkg.pkg,
		}},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				k := NewKitRequest(tt.args.pkg, "Hello", "HelloRequest", false)
				k.ParseRequest()
				spew.Dump(k)
				//fmt.Println(k.BindPathParam())
				//fmt.Println(k.BindQueryParam())
				//fmt.Println(k.BindHeaderParam())
				//fmt.Println(k.BindBodyParam())
				fmt.Println(k.DecodeRequest())
			},
		)
	}
}
