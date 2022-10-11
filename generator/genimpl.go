package generator

import (
	"github.com/dave/jennifer/jen"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
	"strings"
)

const GenImplMark = "@impl"

type GenImpl struct {
	Pkg      *packages.Package
	ImplList map[string]Impl
	Plugs    []ImplPlug
}

func NewGenImpl(pkg *packages.Package, plugs ...ImplPlug) *GenImpl {
	return &GenImpl{
		Pkg:      pkg,
		ImplList: make(map[string]Impl),
		Plugs:    plugs,
	}
}

type ImplPlug interface {
	Name() string
	Gen(pkg *packages.Package, name string, impl Impl) error
	JenF(name string) *jen.File
}


type Impl struct {
	Imports []*ast.ImportSpec
	Doc     *ast.CommentGroup
	Methods []ImplMethod
}

type ImplMethod struct {
	Name           string
	Comment        *ast.CommentGroup
	Params         MethodParamSlice
	Results        MethodParamSlice
	ReturnsError   bool
	AcceptsContext bool
}


func (m ImplMethod) ResultsExcludeErr() MethodParamSlice {
	tmp := make(MethodParamSlice, 0, 0)
	for _, p := range m.Results {
		if p.Type.String() == "error" {
			continue
		}

		tmp = append(tmp, p)
	}
	return tmp
}

func (m ImplMethod) ParamsExcludeCtx() MethodParamSlice {
	tmp := make(MethodParamSlice, 0, 0)
	for _, p := range m.Results {
		if p.Type.String() == "context.Context" {
			continue
		}

		tmp = append(tmp, p)
	}
	return tmp
}

func (m ImplMethod) SwagRespObjData() string {
	if len(m.ResultsExcludeErr()) == 0 {
		return "data=string"
	}

	if len(m.ResultsExcludeErr()) == 1 {
		return "data=" + m.ResultsExcludeErr()[0].Type.String()
	}

	var s []string
	for _,v:= range m.ResultsExcludeErr() {
		s = append(s, "data."+v.Name + "=" + v.Type.String())
	}
	return strings.Join(s, ",")
}

func (m ImplMethod) ResultsMapExcludeErr() string {
	ss := []string{}
	for _, r := range m.ResultsExcludeErr() {
		ss = append(ss, `"`+r.Name+`": `+r.Name)
	}
	return "map[string]interface{}{\n" + strings.Join(ss, ",\n ") + "}"
}

type MethodParamSlice []MethodParam

type MethodParam struct {
	Comment  []*ast.Comment
	Name     string
	Type     types.Type
	Variadic bool
}

func (g *GenImpl) parseImpl(ti *types.Interface) Impl {

	impl := Impl{
		Methods: make([]ImplMethod, 0),
	}

	for i := 0; i < ti.NumMethods(); i++ {
		var returnsError bool
		var acceptsContext bool
		var params MethodParamSlice
		var results MethodParamSlice
		var comment *ast.CommentGroup
		m := ti.Method(i)
		comment = GetCommentByTokenPos(g.Pkg, m.Pos())
		methodName := m.Name()
		ps := m.Type().(*types.Signature).Params()
		for i := 0; i < ps.Len(); i++ {
			mParam := MethodParam{}
			if i == 0 {
				if ps.At(i).Type().String() == "context.Context" {
					acceptsContext = true
				}
			}
			p := ps.At(i)
			t := p.Type()
			pName := p.Name()

			mParam.Name = pName
			mParam.Type = t

			params = append(params, mParam)
		}

		rs := m.Type().(*types.Signature).Results()
		for i := 0; i < rs.Len(); i++ {
			mParam := MethodParam{}
			if i == rs.Len()-1 {
				if rs.At(i).Type().String() == "error" {
					returnsError = true
				}
			}

			r := rs.At(i)
			t := r.Type()
			if _, ok := t.(*types.Named); ok {
				t = t.Underlying()
			}
			rName := r.Name()

			mParam.Name = rName
			mParam.Type = t

			results = append(results, mParam)
		}

		implMethod := ImplMethod{
			Name:           methodName,
			Comment:        comment,
			Params:         params,
			Results:        results,
			ReturnsError:   returnsError,
			AcceptsContext: acceptsContext,
		}
		impl.Methods = append(impl.Methods, implMethod)
	}

	return impl
}

func (g *GenImpl) Parse() {
	for _, v := range g.Pkg.Syntax {
		recordImportSpec := make([]*ast.ImportSpec,0)
		ast.Inspect(v, func(node ast.Node) bool {
			if importSpec, ok := node.(*ast.ImportSpec);ok {
				recordImportSpec = append(recordImportSpec,importSpec)
			}

			gd, ok := node.(*ast.GenDecl)
			if !ok {
				return true
			}

			doc := NewAstDocFormat(gd.Doc)

			if !doc.ContainsMark(GenImplMark) {
				return true
			}

			spec, ok := gd.Specs[0].(*ast.TypeSpec)
			if !ok {
				log.Printf("genimpl: %s is not a type", gd.Specs[0])
				return false
			}

			_, ok = spec.Type.(*ast.InterfaceType)
			if !ok {
				log.Printf("genimpl: %s is not a interface", spec.Name)
				return false
			}

			t := g.Pkg.TypesInfo.TypeOf(spec.Type)
			if t == nil {
				log.Printf("genimpl: %s is not a type", spec.Name)
				return false
			}

			impl := g.parseImpl(t.(*types.Interface))
			impl.Doc = gd.Doc
			impl.Imports = recordImportSpec
			g.ImplList[spec.Name.String()] = impl
			return false
		})
	}
}
