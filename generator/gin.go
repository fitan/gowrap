package generator

import (
	"bytes"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
	"path"
	"reflect"
	"strings"
)

const HttpGinMark = "@http-gin"

func HasGinHttpMark(fi *ast.Field) (has bool, url string, method string) {
	if fi.Doc == nil {
		return false, "", ""
	}
	for _, v := range fi.Doc.List {
		vl := strings.Split(v.Text, " ")
		fmt.Println(vl)
		if len(vl) == 4 && vl[1] == HttpGinMark {
			return true, vl[2], vl[3]
		}
	}
	return false, "", ""
}

type HttpMethod struct {
	Name      string
	fs        *token.FileSet
	SrcPkg    *packages.Package
	SrcFile   *ast.File
	SrcField  *ast.Field
	DstPkg    *packages.Package
	DstFile   *ast.File
	DstStruct *ast.StructType
	GinParams GinParams
}

func NewHttpMethod(name string, srcPkg *packages.Package, srcFile *ast.File, fi *ast.Field) *HttpMethod {
	return &HttpMethod{Name: name, SrcPkg: srcPkg, SrcFile: srcFile, SrcField: fi, GinParams: GinParams{}}
}

func (h *HttpMethod) Parse() (bool, error) {
	f, ok := h.SrcField.Type.(*ast.FuncType)
	if !ok {
		return false, fmt.Errorf("%q is not a method", h.Name)
	}
	has, url, method := HasGinHttpMark(h.SrcField)
	if !has {
		return has, nil
	}

	if has {
		h.GinParams.Url = url
		h.GinParams.Method = method
	}

	if len(f.Params.List) != 2 {
		return has, fmt.Errorf("%s method params len must 2", h.Name)
	}

	if len(f.Results.List) != 2 {
		return has, fmt.Errorf("%s method results len must 2", h.Name)
	}
	findPkg, findFile, findStruct, err := FindByExpr(h.SrcPkg, h.SrcFile, f.Params.List[1].Type)
	if err != nil {
		return has, err
	}

	h.DstPkg = findPkg
	h.DstFile = findFile
	h.DstStruct = findStruct
	h.gin()
	return has, nil

}

func (h *HttpMethod) gin() (gp GinParams) {
	for _, field := range h.DstStruct.Fields.List {
		for _, ident := range field.Names {
			switch ident.Name {
			case "Query":
				h.GinParams.HasQuery = true
			case "Body":
				h.GinParams.HasBody = true
			case "Uri":
				h.GinParams.HasUri = true
			case "Header":
				h.GinParams.HasHeader = true
			case "CtxKey":
				h.GinParams.HasKey = true
			}
		}
	}
	return
}

func FindByExpr(pkg *packages.Package, file *ast.File, expr ast.Expr) (*packages.Package, *ast.File, *ast.StructType, error) {
	switch t := expr.(type) {
	// struct 在同一个pkg里面
	case *ast.Ident:
		findFile, findStruct, err := FindStructTypeByName(pkg, t.Name)
		if err != nil {
			return nil, nil, nil, err
		}
		return pkg, findFile, findStruct, nil
	// struct 是selector类型， 在另外的pkg里面
	case *ast.SelectorExpr:
		findPkg, err := FindSelectPkg(pkg, file, t)
		if err != nil {
			return nil, nil, nil, err
		}

		findFile, findStruct, err := FindStructTypeByName(findPkg, t.Sel.Name)
		if err != nil {
			return nil, nil, nil, err
		}
		return findPkg, findFile, findStruct, nil

	// 本身就是struct类型
	case *ast.StructType:
		return pkg, file, t, nil
	}
	// 未知的状态
	return nil, nil, nil, fmt.Errorf("unknown type %s", Node2String(pkg.Fset, expr))
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

// FindSelectPkg a.x  找到 a的pkg
func FindSelectPkg(pkg *packages.Package, file *ast.File, selector *ast.SelectorExpr) (*packages.Package, error) {
	selectName := selector.X.(*ast.Ident).Name
	for _, importSpec := range file.Imports {
		if importSpec.Name != nil && selectName == selectName {
			return pkg.Imports[TrimImport(importSpec.Path.Value)], nil
		} else {
			if selectName == path.Base(TrimImport(importSpec.Path.Value)) {
				return pkg.Imports[TrimImport(importSpec.Path.Value)], nil
			}
		}
	}
	return nil, fmt.Errorf("not find select pkg. pkgName: %s, pkgPath: %s, selectName: %s", pkg.Name, pkg.PkgPath, selectName)
}

// TrimImport 去掉 import "a"  还原为 “a”  否则为 “”a“”
func TrimImport(s string) string {
	s = strings.TrimSuffix(s, `"`)
	s = strings.TrimPrefix(s, `"`)
	return s
}

// FindStructTypeByName 在pkg中找到名字为typeName的struct 类型
func FindStructTypeByName(pkg *packages.Package, typeName string) (*ast.File, *ast.StructType, error) {
	findFile, findTypeSpec, err := FindTypeSpecByName(pkg, typeName)
	if err != nil {
		return nil, nil, err
	}
	if st, ok := findTypeSpec.Type.(*ast.StructType); ok {
		return findFile, st, nil
	}
	return nil, nil, fmt.Errorf("typeName: %s not structType", typeName)
}

// FindTypeSpecByName 在pkg中找到名字为typeName的type
func FindTypeSpecByName(pkg *packages.Package, typeName string) (*ast.File, *ast.TypeSpec, error) {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if typeSpec.Name.Name == typeName {
							return file, typeSpec, nil
						}
					}
				}
			}
		}
	}
	return nil, nil, fmt.Errorf("pkgName: %s, pkgPath: %s, not find typeName: %s", typeName, pkg.Name, pkg.PkgPath)
}

type TagMsg struct {
	TagValue string
	Comment  string
}

func FindTagAndCommentByStruct(pkg *packages.Package, file *ast.File, structType *ast.StructType, tagName string) []TagMsg {
	tagMsgs := make([]TagMsg, 0, 0)
	ast.Inspect(structType.Fields, func(node ast.Node) bool {
		fd, ok := node.(*ast.Field)
		if ok {
			if fd.Tag != nil {
				tagTool := reflect.StructTag(fd.Tag.Value[1 : len(fd.Tag.Value)-1])
				value, ok := tagTool.Lookup(tagName)
				if ok {
					msg := TagMsg{
						TagValue: value,
						Comment:  strings.ReplaceAll(fd.Doc.Text(), "\n", "\\n"),
					}
					tagMsgs = append(tagMsgs, msg)
				}
			}
		}

		if _, ok := node.(*ast.BasicLit); ok {
			return false
		}
		switch nodeType := node.(type) {
		case *ast.Field:
			tagMsgs = append(tagMsgs, FindTagByType(pkg, file, nodeType.Type, tagName)...)
		}
		return true
	})
	return tagMsgs
}

func FindTagByType(pkg *packages.Package, file *ast.File, ty ast.Node, tagName string) []TagMsg {
	tagMsgs := make([]TagMsg, 0, 0)
	ast.Inspect(ty, func(node ast.Node) bool {
		switch t := node.(type) {
		case *ast.StructType:
			return false
		default:
			e, ok := node.(ast.Expr)
			if ok {
				_, ok := pkg.TypesInfo.TypeOf(e).Underlying().(*types.Struct)
				if ok {
					switch structType := t.(type) {
					// remote pkg
					case *ast.SelectorExpr:
						FindPkg, err := FindSelectPkg(pkg, file, structType)
						findFile, findStruct, err := FindStructTypeByName(FindPkg, structType.X.(*ast.Ident).Name)
						if err != nil {
							return false
						}
						tagMsgs = append(tagMsgs, FindTagAndCommentByStruct(FindPkg, findFile, findStruct, tagName)...)
						return false
					// local pkg
					case *ast.Ident:
						localFile, findStruct, err := FindStructTypeByName(pkg, structType.Name)
						if err != nil {
							return false
						}
						tagMsgs = append(tagMsgs, FindTagAndCommentByStruct(pkg, localFile, findStruct, tagName)...)
					}
				}
			}
		}
		return true
	})
	return tagMsgs
}
