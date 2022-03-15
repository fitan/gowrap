package pkg

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

var errPackageNotFound = errors.New("package not found")

const mode packages.LoadMode = packages.NeedName |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo |
	packages.NeedImports |
	packages.NeedModule |
	packages.NeedExportsFile |
	packages.NeedTypesSizes |
	packages.NeedFiles |
	packages.NeedDeps |
	packages.NeedCompiledGoFiles

// Load loads package by its import path
func Load(path string, pkgNeedSyntax bool) (*packages.Package, error) {
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedDeps}
	if pkgNeedSyntax {
		cfg = &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedDeps | packages.NeedSyntax}
	}
	//cfg := &packages.Config{Mode: mode}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, err
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
