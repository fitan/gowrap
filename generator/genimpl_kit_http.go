package generator

import (
	"github.com/dave/jennifer/jen"
	"golang.org/x/tools/go/packages"
	"log"
)

const name = "kitHttp"
const kitHttpRouterMark = "@kit-http"
const kitHttpRequestMark = "@kit-http-request"

type GenImplKitHttp struct {
	jenF *jen.File
	jenFList []*jen.File
}

func (g *GenImplKitHttp) Name() string {
	return name
}

func (g *GenImplKitHttp) Gen(pkg *packages.Package, name string, impl Impl) {

}

func (g *GenImplKitHttp) JenF() *jen.File {
	panic("implement me")
}


func (g *GenImplKitHttp) confirmMark(m ImplMethod) {
	f := NewAstDocFormat(m.Comment)
	var path string
	var method string
	var request string
	f.MarkValuesMapping(kitHttpRouterMark, &path, &method)
	f.MarkValuesMapping(kitHttpRequestMark, &request)

	if path == "" || method == "" || request == "" {
		log.Printf("method %s not found mark %s", m.Name, kitHttpRouterMark)
		return
	}


}

func (g *GenImplKitHttp) makeHttpHandler() {
	g.jenF.ImportAlias("github.com/go-kit/kit/transport/http","kithttp")
	g.jenF.ImportName("github.com/go-kit/kit/endpoint", "")

	jen.Func().Id("MakeHTTPHandler").
		Params(
			jen.Id("s").Id("Service"),
			jen.Id("dmw").Id("[]endpoint.Middleware"),
			jen.Id("opts").Id("[]kithttp.ServerOption"),
		).Params(jen.Id("http.Handler")).
		Block(
			jen.Var().Id("ems").Id("[]endpoint.Middleware"),
			jen.Id("r").Op(":=").Id("mux.NewRouter()"),
		)
}


