package generator

import (
	"github.com/dave/jennifer/jen"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
)

const GenImplMark = "@impl"

type GenImpl struct {
	Pkg      *packages.Package
	ImplList map[string]Impl
	Plugs    []ImplPlug
	JenF     *jen.File
}

func NewGenImpl(pkg *packages.Package, plugs ...ImplPlug) *GenImpl {
	return &GenImpl{
		Pkg:      pkg,
		ImplList: make(map[string]Impl),
		Plugs:    plugs,
	}
}

type ImplPlug struct {
	ListMethod map[string]ImplMethod
}

type Impl struct {
	Doc     *ast.CommentGroup
	Methods []ImplMethod
}

type ImplMethod struct {
	Name           string
	Comment        []*ast.Comment
	Params         MethodParamSlice
	Results        MethodParamSlice
	ReturnsError   bool
	AcceptsContext bool
}

type MethodParamSlice []MethodParam

type MethodParam struct {
	Comment  []*ast.Comment
	Name     string
	Type     types.Type
	Variadic bool
}

func (g *GenImpl) parseImpl(name string, doc *ast.CommentGroup, ti *types.Interface) {

	impl := Impl{
		Doc:     doc,
		Methods: make([]ImplMethod, 0),
	}

	for i := 0; i < ti.NumMethods(); i++ {
		var returnsError bool
		var acceptsContext bool
		var params MethodParamSlice
		var results MethodParamSlice
		var comment []*ast.Comment
		m := ti.Method(i)
		comment = GetCommentByTokenPos(g.Pkg, m.Pos()).List
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

	g.ImplList[name] = impl
}

func (g *GenImpl) Parse() {
	for _, v := range g.Pkg.Syntax {
		ast.Inspect(v, func(node ast.Node) bool {
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

			g.parseImpl(spec.Name.String(), gd.Doc, t.(*types.Interface))
			return false
		})
	}
}
