package pkg

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"time"

	"golang.org/x/tools/go/packages"
)

var errPackageNotFound = errors.New("package not found")

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

// Load loads package by its import path
func Load(path string, pkgNeedSyntax bool) (*packages.Package, error) {
	t1 := time.Now()
	log.Printf("open pkg need syntax %v", pkgNeedSyntax)
	defer func() {
		log.Printf("load pkg time: %v", time.Now().Sub(t1).String())
	}()
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedModule}
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

// AST returns package's abstract syntax tree
func AST(fs *token.FileSet, p *packages.Package) (*ast.Package, error) {
	dir := Dir(p)

	pkgs, err := parser.ParseDir(fs, dir, nil, parser.DeclarationErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}

	if ap, ok := pkgs[p.Name]; ok {
		return ap, nil
	}

	return &ast.Package{Name: p.Name}, nil
}

// Dir returns absolute path of the package in a filesystem
func Dir(p *packages.Package) string {
	files := append(p.GoFiles, p.OtherFiles...)
	if len(files) < 1 {
		return p.PkgPath
	}

	return filepath.Dir(files[0])
}
