package generator

import (
	"github.com/fitan/jennifer/jen"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
	"strings"
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
