package generator

import (
	"bytes"
	"github.com/davecgh/go-spew/spew"
	"github.com/fitan/jennifer/jen"
	"github.com/pkg/errors"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 获取struct字段里的注释
func GetCommentByTokenPos(pkg *packages.Package, pos token.Pos) *ast.CommentGroup {
	fieldFileName := pkg.Fset.Position(pos).Filename
	fieldLine := pkg.Fset.Position(pos).Line
	var fieldComment *ast.CommentGroup
	for _, syntax := range pkg.Syntax {
		fileName := pkg.Fset.Position(syntax.Pos()).Filename
		if fieldFileName == fileName {
			for _, c := range syntax.Comments {
				if pkg.Fset.Position(c.End()).Line+1 == fieldLine {
					fieldComment = c
				}
			}
			break
		}
	}
	return fieldComment
}

func JenFAddImports(p *packages.Package, f *jen.File) {
	for _, s := range p.Syntax {
		for _, v := range s.Imports {
			var path, pathName string
			if v.Path != nil {
				path = strings.Trim(v.Path.Value, `"`)
			}
			if v.Name != nil {
				pathName = strings.Trim(v.Name.Name, `"`)
			}
			f.AddImport(path, pathName)
		}
	}
}

func Node2String(fset *token.FileSet, node interface{}) string {
	var buf bytes.Buffer
	err := printer.Fprint(&buf, fset, node)
	if err != nil {
		spew.Dump(node)
		log.Panicln(err.Error())
	}
	return buf.String()
}

func LoadMainImports() (res []*ast.ImportSpec, err error) {
	mainName := "main.go"
	readMain := func(path string) (res []*ast.ImportSpec, err error) {
		fset := token.NewFileSet()
		var f *ast.File
		f, err = parser.ParseFile(fset, path, nil, parser.ParseComments|parser.ImportsOnly)
		if err != nil {
			return
		}

		res = f.Imports
		return
	}

	workPath, err := os.Getwd()
	if err != nil {
		return
	}

	for {
		_, err = os.Stat(filepath.Join(workPath, mainName))
		if err == nil {
			return readMain(filepath.Join(workPath, mainName))
		}

		if err != nil {
			if !os.IsNotExist(err) {
				return
			}
		}

		_, err = os.Stat(filepath.Join(workPath, "cmd", mainName))
		if err == nil {
			return readMain(filepath.Join(workPath, "cmd", mainName))
		}

		if err != nil {
			if !os.IsNotExist(err) {
				return
			}
		}

		preWorkPath := filepath.Dir(workPath)
		if preWorkPath == workPath {
			err = errors.New("not found main.go")
			return
		}

		workPath = preWorkPath

	}
}

var errPackageNotFound = errors.New("package not found")

const mode packages.LoadMode = packages.NeedName |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo |
	packages.NeedImports |
	//packages.NeedModule |
	//packages.NeedTypesSizes |
	//packages.NeedDeps |
	packages.NeedFiles

//packages.NeedCompiledGoFiles |
//packages.NeedExportFile

// Load loads package by its import path
func Load(path string, pkgNeedSyntax bool) (*packages.Package, error) {
	t1 := time.Now()
	log.Printf("open pkg need syntax %v", pkgNeedSyntax)
	defer func() {
		log.Printf("load pkg time: %v", time.Now().Sub(t1).String())
	}()
	//cfg := &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedModule | }
	cfg := &packages.Config{Mode: mode}
	if pkgNeedSyntax {
		//cfg = &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedSyntax | packages.NeedTypesInfo}
		cfg = &packages.Config{Mode: mode}
	}
	//cfg := &packages.Config{Mode: mode}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		log.Printf("packages.Load err: %v", err.Error())
		return pkgs[0], nil
	}

	if len(pkgs) < 1 {
		return nil, errPackageNotFound
	}

	//if len(pkgs[0].Errors) > 0 {
	//	return nil, pkgs[0].Errors[0]
	//}

	return pkgs[0], nil
}

func type2RawTypeId(pkg *packages.Package, t types.Type, s string) (res string) {
	switch tName := t.(type) {
	case *types.Named:
		if pkg.Name == tName.Obj().Pkg().Name() {
			return s + tName.Obj().Id()
		}

		return s + tName.Obj().Pkg().Name() + "." + tName.Obj().Id()

	case *types.Map:
		return "map[" + type2RawTypeId(pkg, tName.Key(), "") + "]" + type2RawTypeId(pkg, tName.Elem(), s)
	case *types.Slice:
		return "[]" + type2RawTypeId(pkg, tName.Elem(), s)
	}
	return s
}
