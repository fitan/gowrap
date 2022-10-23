package generator

import (
	"go/ast"
	"go/token"
	"go/types"
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

// 寻找非named types
func FindNotNamedType(pkg *packages.Package, expr ast.Expr) types.Type {
	t := pkg.TypesInfo.TypeOf(expr)
	if _, ok := t.(*types.Named); ok {
		return t.Underlying()
	}

	return t

}

func IdByType(pkg packages.Package, t types.Type) string {
	s := strings.Replace(t.String(), pkg.PkgPath, "", -1)
	return s
}
