package generator

import (
	"fmt"
	"go/ast"
	"go/types"
	"regexp"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/fitan/gowrap/xtype"
	"github.com/samber/lo"
	"golang.org/x/tools/go/packages"
)

type typePrinter interface {
	PrintType(ast.Node) (string, error)
}

// Method represents a method's signature
type Method struct {
	Doc     []string
	Comment []string
	Name    string
	Params  ParamsSlice
	Results ParamsSlice

	ReturnsError   bool
	AcceptsContext bool

	// my extra
	HasGin bool
	Gin    GinParams

	// my extra
	RawKit Kit
	Kit    KitParams

	KitRequest       *KitRequest
	KitRequestDecode string

	Type2Ast *type2ast
}

func DocFormat(doc string) string {
	return strings.Join(strings.Fields(doc), " ")
}

func (m Method) EnableSwag() bool {
	if m.Doc == nil {
		return true
	}

	for _, c := range m.Doc {
		docFormat := DocFormat(c)
		if strings.HasPrefix(docFormat, "// @swag false") {
			return false
		}
	}

	return true
}

func (m Method) ClientStruct() (code []jen.Code, err error) {
	code = make([]jen.Code, 0)
	code = append(code,
		jen.Type().Id(m.Name+"service").Struct(
			jen.Id("prePath").String(),
			jen.Id("opt").Index().Qual("github.com/go-kit/kit/transport/http", "ClientOption"),
			jen.Id("decode").Id("func(i interface{}) func(ctx context.Context, res *http.Response) (response interface{}, err error)"),
		),
	)

	return code, nil
}

func (m Method) ClientInterfaceFunc() jen.Code {

	resultParams := make([]jen.Code, 0)
	// resultStruct := jen.Line()
	if m.HasResultsExcludeErr() {
		// if len(m.ResultsExcludeErr()) > 0 {
		// 	resultParams = append(resultParams, jen.Id("res").Id(m.ResultsExcludeErr()[0].Type))
		// }

		if len(m.ResultsExcludeErr()) > 0 {
			resultParams = append(resultParams, jen.Id("res").Id(m.Name+"ClientRes"))
			// resultStruct.Type().Id(m.Name + "ClientResponse").StructFunc(func(g *jen.Group) {
			// 	for _, v := range m.ResultsExcludeErr() {
			// 		g.Id(strings.ToUpper(string(v.Name[0])) + v.Name[1:]).Id(v.Type).Tag(map[string]string{"json": v.Name})
			// 	}
			// })

		}
	}
	if m.ReturnsError {
		resultParams = append(resultParams, jen.Id("err").Error())
	}

	return jen.Id(m.Name).Params(jen.Id("ctx").Id("context.Context"), jen.Id("req").Id(m.Name+"ClientReq"), jen.Id("option").Id("*Option")).Params(resultParams...)
}

// func (m Method) ClientReq() string {
// 	return m.Type2Ast.Parse(xtype.TypeOf(m.KitRequest.RequestTypeOf), m.KitRequest.RequestName+"NewReq")
// }

func (m Method) ClientFunc(basePath string) string {
	code := make([]jen.Code, 0)
	code = append(code,
		jen.Id(`_, err = valid.ValidateStruct(req)`),
		jen.If().Err().Op("!=").Nil().Block(
			jen.Err().Op("=").Id(`fmt.Errorf("validate request error: %v", err)`),
			jen.Id("return"),
		),
		jen.Id(`opt := s.mergeOpt(option)`),
	)
	queryCode := make([]jen.Code, 0)
	formCode := make([]jen.Code, 0)

	var fileCode *jen.Statement

	if len(m.KitRequest.Form) != 0 {
		fileCode = jen.Type().Id(m.KitRequest.RequestName + "NewBody").StructFunc(
			func(group *jen.Group) {
				for _, v := range m.KitRequest.Path {
					group.Comment(v.Comment()).Line().Id(v.FieldName).String().Tag(map[string]string{"json": v.ParamName})
				}

				for _, v := range m.KitRequest.Header {
					group.Comment(v.Comment()).Line().Id(v.FieldName).String().Tag(map[string]string{"json": v.ParamName})
				}

				for _, v := range m.KitRequest.Query {
					group.Comment(v.Comment()).Line().Id(v.FieldName).String().Tag(map[string]string{"json": v.ParamName})
				}

				for _, v := range m.KitRequest.Form {
					if v.XTypeID == "[]*multipart.FileHeader" {
						group.Comment(v.Comment()).Line().Id(v.FieldName).Struct(jen.Id("FileName").String().Tag(map[string]string{"json": "fileName"}), jen.Id("File").String().Tag(map[string]string{"json": "file"}))
					} else {
						group.Comment(v.Comment()).Line().Id(v.FieldName).Id(v.RawParamType).Tag(map[string]string{"json": v.ParamName})
					}
				}
			},
		)

		for k, v := range m.KitRequest.Query {
			q := m.KitRequest.Query[k]
			q.ParamPath, _ = lo.Last(strings.Split(v.ParamPath, "."))
			m.KitRequest.Query[k] = q
		}

		for k, v := range m.KitRequest.Header {
			h := m.KitRequest.Header[k]
			h.ParamPath, _ = lo.Last(strings.Split(v.ParamPath, "."))
			m.KitRequest.Header[k] = h
		}

		for k, v := range m.KitRequest.Path {
			p := m.KitRequest.Path[k]
			p.ParamPath, _ = lo.Last(strings.Split(v.ParamPath, "."))
			m.KitRequest.Path[k] = p
		}

		for k, v := range m.KitRequest.Form {
			f := m.KitRequest.Form[k]
			f.ParamPath, _ = lo.Last(strings.Split(v.ParamPath, "."))
			m.KitRequest.Form[k] = f
		}
	}

	if len(m.KitRequest.Form) != 0 {
		code = append(code,
			jen.Var().Id("formB").Qual("bytes", "Buffer"),
			jen.Id("form").Op(":=").Qual("mime/multipart", "NewWriter").Call(jen.Id("&formB")),
			jen.Var().Id("part").Qual("io", "Writer"),
		)

		for k, v := range m.KitRequest.Form {
			if v.XTypeID == "[]*multipart.FileHeader" {
				formCode = append(formCode,
					jen.List(jen.Id("part"), jen.Err()).Op("=").Id("form.CreateFormFile").Call(jen.Lit(k), jen.Id("req."+v.ParamPath+".FileName")),
					jen.If().Err().Op("!=").Nil().Block(
						jen.Err().Op("=").Qual("fmt", "Errorf").Call(jen.Lit("create form field %s error: %v"), jen.Lit(k), jen.Err()),
						jen.Return(),
					),
					jen.List(jen.Id("_"), jen.Err()).Op("=").Id("io.WriteString").Call(jen.Id("part"), jen.Id("req."+v.ParamPath+".File")),
					jen.If().Err().Op("!=").Nil().Block(
						jen.Err().Op("=").Qual("fmt", "Errorf").Call(jen.Lit("io.WriteString part %s error: %v"), jen.Lit(k), jen.Err()),
						jen.Return(),
					),
				)
				continue
			}

			if !v.XType.Basic {
				formCode = append(formCode,
					jen.List(jen.Id("part"), jen.Err()).Op("=").Id("form.CreateFormField").Call(jen.Lit(k)),
					jen.If().Err().Op("!=").Nil().Block(
						jen.Err().Op("=").Qual("fmt", "Errorf").Call(jen.Lit("create form field %s error: %v"), jen.Lit(k), jen.Err()),
						jen.Return(),
					),
					jen.List(jen.Id("_"), jen.Err()).Op("=").Id("io.WriteString").Call(jen.Id("part"), jen.Qual("github.com/spf13/cast", "ToString").Call(jen.Id("req."+v.ParamPath))),
					jen.If().Err().Op("!=").Nil().Block(
						jen.Err().Op("=").Qual("fmt", "Errorf").Call(jen.Lit("io.WriteString part %s error: %v"), jen.Lit(k), jen.Err()),
						jen.Return(),
					),
				)
				continue
			}

			if v.XType.BasicType.Kind() != types.String {
				formCode = append(formCode,
					jen.List(jen.Id("part"), jen.Err()).Op("=").Id("form.CreateFormField").Call(jen.Lit(k)),
					jen.If().Err().Op("!=").Nil().Block(
						jen.Err().Op("=").Qual("fmt", "Errorf").Call(jen.Lit("create form field %s error: %v"), jen.Lit(k), jen.Err()),
						jen.Return(),
					),
					jen.List(jen.Id("_"), jen.Err()).Op("=").Id("io.WriteString").Call(jen.Id("part"), jen.Qual("github.com/spf13/cast", "ToString").Call(jen.Id("req."+v.ParamPath))),
					jen.If().Err().Op("!=").Nil().Block(
						jen.Err().Op("=").Qual("fmt", "Errorf").Call(jen.Lit("io.WriteString part %s error: %v"), jen.Lit(k), jen.Err()),
						jen.Return(),
					),
				)
				continue
			}

			formCode = append(formCode,
				jen.List(jen.Id("part"), jen.Err()).Op("=").Id("form.CreateFormField").Call(jen.Lit(k)),
				jen.If().Err().Op("!=").Nil().Block(
					jen.Err().Op("=").Qual("fmt", "Errorf").Call(jen.Lit("create form field %s error: %v"), jen.Lit(k), jen.Err()),
					jen.Return(),
				),
				jen.List(jen.Id("_"), jen.Err()).Op("=").Id("io.WriteString").Call(jen.Id("part"), jen.Id("req."+v.ParamPath)),
				jen.If().Err().Op("!=").Nil().Block(
					jen.Err().Op("=").Qual("fmt", "Errorf").Call(jen.Lit("io.WriteString part %s error: %v"), jen.Lit(k), jen.Err()),
					jen.Return(),
				),
			)
		}
		code = append(code, formCode...)
		code = append(code, jen.If().Err().Op("=").Id("form.Close();").Id("err").Op("!=").Nil().Block(
			jen.Err().Op("=").Qual("fmt", "Errorf").Call(jen.Lit("close form error: %v"), jen.Err()),
			jen.Return(),
		))
	}

	if len(m.KitRequest.Query) != 0 {
		code = append(code,
			jen.Id("q").Op(":=").Qual("net/url", "Values").Values(),
		)

		for k, v := range m.KitRequest.Query {
			if !v.XType.Basic {
				queryCode = append(queryCode,
					jen.Id("q.Add").Call(jen.Lit(k), jen.Qual("github.com/spf13/cast", "ToString").Call(jen.Id("req."+v.ParamPath))),
				)
				continue
			}

			if v.XType.BasicType.Kind() != types.String {
				queryCode = append(queryCode,
					jen.Id("q.Add").Call(jen.Lit(k), jen.Qual("github.com/spf13/cast", "ToString").Call(jen.Id("req."+v.ParamPath))),
				)
				continue
			}

			queryCode = append(queryCode,
				jen.Id("q.Add").Call(jen.Lit(k), jen.Id("req."+v.ParamPath)),
			)
		}
		code = append(code, queryCode...)
	}

	re, _ := regexp.Compile(`\{(.*?)\}`)
	matches := re.FindAllStringSubmatch(m.RawKit.Conf.Url, -1)
	params := make([]string, 0)
	for _, match := range matches {
		if len(match) > 1 {
			params = append(params, match[1])
		}
	}

	fmtSprintParam := make([]jen.Code, 0)
	fmtSprintParam = append(fmtSprintParam, jen.Lit(re.ReplaceAllString(m.RawKit.Conf.Url, "%s")))
	for _, v := range params {
		fmtSprintParam = append(fmtSprintParam, jen.Id("req"+m.KitRequest.ParamPath(v)))
	}
	if len(params) != 0 {
		code = append(
			code,
			jen.List(jen.Id("urlStr"), jen.Err()).Op(":=").Qual("net/url", "JoinPath").Call(jen.Id("s.PrePath"), jen.Lit(basePath), jen.Id("fmt.Sprintf").Call(
				fmtSprintParam...,
			)),
			jen.If(jen.Err().Op("!=").Nil()).Block(
				jen.Err().Op("=").Qual("fmt", "Errorf").Call(jen.Lit("parse prePath %s bathPath %s error: %v"), jen.Id("s.PrePath"), jen.Lit(basePath), jen.Err()),
				jen.Return(),
			),
		)
	} else {
		code = append(
			code,
			jen.List(jen.Id("urlStr"), jen.Err()).Op(":=").Qual("net/url", "JoinPath").Call(jen.Id("s.PrePath"), jen.Lit(basePath), jen.Lit(re.ReplaceAllString(m.RawKit.Conf.Url, "%s"))),
			jen.If(jen.Err().Op("!=").Nil()).Block(
				jen.Err().Op("=").Qual("fmt", "Errorf").Call(jen.Lit("parse prePath %s bathPath %s error: %v"), jen.Id("s.PrePath"), jen.Lit(basePath), jen.Err()),
				jen.Return(),
			),
		)
	}

	code = append(code,
		jen.List(jen.Id("u"), jen.Err()).Op(":=").Qual("net/url", "Parse").CallFunc(func(group *jen.Group) {
			if len(queryCode) != 0 {
				group.Id(`fmt.Sprintf("%s?%s", strings.TrimRight(urlStr, "/"), q.Encode())`)
			} else {
				group.Id("urlStr")
			}
		}),
	)

	code = append(code,
		jen.If(jen.Err().Op("!=").Nil()).Block(
			jen.Err().Op("=").Qual("fmt", "Errorf").Call(jen.Lit("parse url %s error: %v"), jen.Id("urlStr"), jen.Err()),
			jen.Return(),
		),
	)

	var setHeaderCode []jen.Code
	for k, v := range m.KitRequest.Header {
		setHeaderCode = append(setHeaderCode,
			jen.Id("req.Header.Set").Call(jen.Lit(k), jen.Id("req."+v.ParamPath)),
		)
	}

	if len(setHeaderCode) != 0 {
		code = append(code,
			jen.Id(`headerOpt := kithttp.ClientBefore(func(ctx context.Context, r *http.Request) context.Context`).BlockFunc(func(g *jen.Group) {
				g.Add(setHeaderCode...)
				g.Return(jen.Id("ctx"))
			}),
			jen.Id("opt.ClientOpts").Op("=").Append(jen.Id("opt.ClientOpts"), jen.Id("headerOpt")),
		)
	}

	encode := "kithttp.EncodeJSONRequest"

	if len(formCode) != 0 {
		encode = "EncodeFormRequest(form.FormDataContentType())"
	}

	code = append(code,
		jen.Id("factory").Op(":=").Id(`func (instance string) (endpoint.Endpoint, io.Closer, error)`).BlockFunc(func(group *jen.Group) {
			group.Return(jen.Id("kithttp.NewClient").Call(jen.Lit(m.RawKit.Conf.UrlMethod), jen.Id("u"), jen.Id(encode), func() jen.Code {
				if m.HasResultsExcludeErr() {
					return jen.Id("opt.Decode(res)")
				}
				return jen.Id("opt.Decode(nil)")
			}(),
				jen.Id("opt.ClientOpts..."),
			).Dot("Endpoint").Call(), jen.Nil(), jen.Nil())
		}),
		jen.Id(`e := sd.NewEndpointer(s.Instancer, factory, s.Logger,s.EndpointOpts...)`),
		jen.Id(`balancer := lb.NewRoundRobin(e)`),
		jen.Id(`retry := lb.Retry(opt.RetryMax, opt.RetryTimeout, balancer)`),
	)

	code = append(code,
		jen.List(jen.Id("_"), jen.Err()).Op("=").Id("retry").Call(jen.Id("ctx"), func() jen.Code {
			if m.KitRequest.RequestIsBody {
				return jen.Id("req")
			}

			if len(m.KitRequest.Body) != 0 {
				return jen.Id("req." + m.KitRequest.Body.OrderSlice()[0].ParamPath)
			}

			if len(m.KitRequest.Form) != 0 {
				return jen.Id("&formB")
			}

			return jen.Nil()
		}()),
		jen.If(jen.Err().Op("!=").Nil()).Block(
			jen.Id("err").Op("=").Qual("fmt", "Errorf").Call(jen.Lit("endpoint error: %v"), jen.Err()),
			jen.Return(),
		),

		jen.Return(),
	)

	resultParams := make([]jen.Code, 0)
	resultStruct := jen.Line()

	if m.HasResultsExcludeErr() {

		fields := make([]*types.Var, 0)
		tags := make([]string, 0)
		for _, v := range m.ResultsExcludeErr() {

			fields = append(fields, types.NewVar(0, m.KitRequest.pkg.Types, upFirst(v.Name), v.XType.T))
			tags = append(tags, fmt.Sprintf(`json:"%s"`, v.Name))
		}

		newStruct := types.NewStruct(fields, tags)

		resStr := m.Type2Ast.Parse(xtype.TypeOf(newStruct), m.Name+"ClientRes")

		resultStruct.Id(resStr).Line()

		// if len(m.ResultsExcludeErr()) == 1 {
		// 	resultParams = append(resultParams, jen.Id("res").Id(m.ResultsExcludeErr()[0].Type))
		// }

		if len(m.ResultsExcludeErr()) > 0 {
			resultParams = append(resultParams, jen.Id("res").Id(m.Name+"ClientRes"))
			// resultStruct.Type().Id(m.Name + "ClientResponse").StructFunc(func(g *jen.Group) {
			// 	for _, v := range m.ResultsExcludeErr() {
			// 		g.Id(strings.ToUpper(string(v.Name[0])) + v.Name[1:]).Id(v.Type).Tag(map[string]string{"json": v.Name})
			// 	}
			// })

		}
	}
	if m.ReturnsError {
		resultParams = append(resultParams, jen.Id("err").Error())
	}

	reqName := m.Name + "ClientReq"
	if fileCode != nil {
		reqName = m.KitRequest.RequestName + "NewBody"
	}

	funcCode := jen.Add(fileCode).Line().Func().Params(jen.Id("s").Id("HttpClientService")).Id(m.Name).Params(jen.Id("ctx").Id("context.Context"), jen.Id("req").Id(reqName), jen.Id("option").Id("*Option")).Params(resultParams...).Block(
		code...,
	)

	return resultStruct.Line().Add(funcCode).GoString()

}

func (m Method) KitHttpServiceEndpointName() string {
	if m.RawKit.Conf.KitServiceParam.EndpointName != `""` && m.RawKit.Conf.KitServiceParam.EndpointName != "" {
		return m.RawKit.Conf.KitServiceParam.EndpointName
	}
	return "eps." + m.Name + "Endpoint"
}

func (m Method) KitHttpServiceDecodeName() string {

	if m.RawKit.Conf.KitServiceParam.DecodeName != `""` && m.RawKit.Conf.KitServiceParam.DecodeName != "" {
		return m.RawKit.Conf.KitServiceParam.DecodeName
	}
	return "decode" + m.Name + "Request"
}

func (m Method) KitHttpServiceEncodeName() string {
	if m.RawKit.Conf.KitServiceParam.EncodeName != `""` && m.RawKit.Conf.KitServiceParam.EncodeName != "" {
		return m.RawKit.Conf.KitServiceParam.EncodeName
	}
	return "encode.JsonResponse"
}

func (m Method) Annotation() string {
	if m.Doc == nil {
		return ""
	}
	for _, c := range m.Doc {
		docFormat := DocFormat(c)
		if strings.HasPrefix(docFormat, "// "+m.Name) {
			return strings.TrimPrefix(docFormat, "// "+m.Name)
		}
	}
	return strings.TrimPrefix(DocFormat(m.Doc[0]), "// ")
}

func (m Method) KitEndpointName() string {
	if m.Kit.Endpoint == `""` || m.Kit.Endpoint == "" {
		return "Make" + m.Name + "Endpoint"
	}
	return m.Kit.Endpoint
}

func (m Method) KitEncodeName() string {
	if m.Kit.Encode == `""` || m.Kit.Encode == "" {
		return "http.EncodeJSONFormatResponse"
	}
	return m.Kit.Encode
}

func (m Method) KitDecodeName() string {
	if m.Kit.Decode == `""` || m.Kit.Decode == "" {
		return "decode" + m.Name + "Request"
	}
	return m.Kit.Decode
}

type KitParams struct {
	Endpoint string
	Decode   string
	Encode   string
}

type GinParams struct {
	Url     string
	SwagUrl string
	Method  string
	Result  string

	HasUri     bool
	UriTagMsgs []TagMsg

	HasQuery           bool
	QueryRawStructName string
	QueryRawStruct     string

	HasBody           bool
	BodyRawStructName string
	BodyRawStruct     string

	HasHeader     bool
	HeaderTagMsgs []TagMsg

	HasKey bool
}

// Param represents fuction argument or result
type Param struct {
	Doc          []string
	Comment      []string
	Name         string
	Type         string
	Variadic     bool
	HasSerialize bool

	XType *xtype.Type
}

// ParamsSlice slice of parameters
type ParamsSlice []Param

// String implements fmt.Stringer
func (ps ParamsSlice) String() string {
	ss := []string{}
	for _, p := range ps {
		ss = append(ss, p.Name+" "+p.Type)
	}

	return strings.Join(ss, ", ")
}

// Pass returns comma separated params names to
// be passed to a function call with respect to
// variadic functions
func (ps ParamsSlice) Pass() string {
	params := []string{}
	for _, p := range ps {
		params = append(params, p.Pass())
	}

	return strings.Join(params, ", ")
}

// Pass returns a name of the parameter
// If parameter is variadic it returns a name followed by a ...
func (p Param) Pass() string {
	if p.Variadic {
		return p.Name + "..."
	}
	return p.Name
}

// NewMethod returns pointer to Signature struct or error
func NewMethod(pkg *packages.Package, file *ast.File, name string, fi *ast.Field, printer typePrinter, type2ast *type2ast, o Options) (*Method, error) {
	f, ok := fi.Type.(*ast.FuncType)
	if !ok {
		return nil, fmt.Errorf("%q is not a method", name)
	}

	m := Method{Name: name, Type2Ast: type2ast}
	if fi.Doc != nil && len(fi.Doc.List) > 0 {
		m.Doc = make([]string, 0, len(fi.Doc.List))
		for _, comment := range fi.Doc.List {
			m.Doc = append(m.Doc, comment.Text)
		}
	}

	if fi.Comment != nil && len(fi.Comment.List) > 0 {
		m.Comment = make([]string, 0, len(fi.Comment.List))
		for _, comment := range fi.Comment.List {
			m.Comment = append(m.Comment, comment.Text)
		}
	}

	usedNames := map[string]bool{}

	//Always name the last return parameter as an "err" if it's of type "error"
	if f.Results != nil {
		ident, ok := f.Results.List[len(f.Results.List)-1].Type.(*ast.Ident)
		m.ReturnsError = ok && ident.Name == "error"
		usedNames["err"] = true
	}

	if len(f.Params.List) > 0 {
		if se, ok := f.Params.List[0].Type.(*ast.SelectorExpr); ok {
			m.AcceptsContext = ok && se.Sel.Name == "Context"
			usedNames["ctx"] = true
		}
	}

	var err error

	m.Params, err = makeParams(pkg, file, f.Params, usedNames, printer)
	if err != nil {
		return nil, err
	}
	m.Results, err = makeParams(pkg, file, f.Results, usedNames, printer)
	if err != nil {
		return nil, err
	}

	if m.ReturnsError {
		m.Results[len(m.Results)-1].Name = "err"
	}

	if m.AcceptsContext {
		m.Params[0].Name = "ctx"
	}

	return &m, nil
}

// NewParam returns Param struct
func NewParam(pkg *packages.Package, file *ast.File, name string, fi *ast.Field, usedNames map[string]bool, printer typePrinter) (*Param, error) {

	//var hasStruct bool

	//fmt.Println("paramName", Node2String(pkg.Fset, fi.Type))
	//if !JudgeBuiltInType(Node2String(pkg.Fset, fi.Type)) {
	//	_, _, ts, err := FindTypeSpecByExpr(pkg, file, fi.Type)
	//	if err != nil {
	//		return nil, err
	//	}
	//	_, hasStruct = ts.Type.(*ast.StructType)
	//}

	//typeOf := pkg.TypesInfo.TypeOf(fi.Type)
	//xtype.TypeOf(typeOf).

	typ := fi.Type
	if name == "" || usedNames[name] {
		name = genName(typePrefix(typ), 1, usedNames)
	}

	usedNames[name] = true

	typeStr, err := printer.PrintType(typ)
	if err != nil {
		return nil, err
	}

	_, variadic := typ.(*ast.Ellipsis)
	p := &Param{
		Name:         name,
		Variadic:     variadic,
		Type:         typeStr,
		HasSerialize: !JudgeBuiltInType(Node2String(pkg.Fset, fi.Type)),

		XType: xtype.TypeOf(pkg.TypesInfo.TypeOf(fi.Type)),
	}
	if fi.Doc != nil && len(fi.Doc.List) > 0 {
		p.Doc = make([]string, 0, len(fi.Doc.List))
		for _, comment := range fi.Doc.List {
			p.Doc = append(p.Doc, comment.Text)
		}
	}

	if fi.Comment != nil && len(fi.Comment.List) > 0 {
		p.Comment = make([]string, 0, len(fi.Comment.List))
		for _, comment := range fi.Comment.List {
			p.Comment = append(p.Comment, comment.Text)
		}
	}

	return p, nil
}

func makeParams(pkg *packages.Package, file *ast.File, params *ast.FieldList, usedNames map[string]bool, printer typePrinter) (ParamsSlice, error) {
	if params == nil {
		return nil, nil
	}

	result := []Param{}
	for _, p := range params.List {
		//for anonymous parameters we generate params and results names
		//based on their type
		if p.Names == nil {
			param, err := NewParam(pkg, file, "", p, usedNames, printer)
			if err != nil {
				return nil, err
			}
			result = append(result, *param)
		} else {
			for _, ident := range p.Names {
				param, err := NewParam(pkg, file, ident.Name, p, usedNames, printer)
				if err != nil {
					return nil, err
				}
				result = append(result, *param)
			}
		}
	}

	return result, nil
}

func typePrefix(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.SelectorExpr:
		return typePrefix(t.Sel)
	case *ast.StarExpr:
		return typePrefix(t.X) + "p" //*string -> sp (string pointer)
	case *ast.SliceExpr:
		return typePrefix(t.X) + "s" //[]string -> ss (string slice)
	case *ast.ArrayType:
		return typePrefix(t.Elt) + "a" //[2]string -> sa (string array)
	case *ast.MapType:
		return "m"
	case *ast.ChanType:
		return "ch"
	case *ast.StructType:
		return "st"
	case *ast.FuncType:
		return "f"
	case *ast.Ident:
		return strings.ToLower(t.Name[0:1])
	}

	return "p"
}

func genName(prefix string, n int, usedNames map[string]bool) string {
	name := fmt.Sprintf("%s%d", prefix, n)
	if usedNames[name] {
		return genName(prefix, n+1, usedNames)
	}

	return name
}

// Call returns a string with the method call
func (m Method) Call() string {
	params := []string{}
	for _, p := range m.Params {
		params = append(params, p.Pass())
	}

	return m.Name + "(" + strings.Join(params, ", ") + ")"
}

// Pass returns a return statement followed by the method call
// If method does not have any results it returns a method call followed by return statement
func (m Method) Pass(prefix string) string {
	if len(m.Results) > 0 {
		return "return " + prefix + m.Call()
	}

	return prefix + m.Call() + "\nreturn"
}

// ParamsNames returns a list of method params names
func (m Method) ParamsNames() string {
	ss := []string{}
	for _, p := range m.Params {
		ss = append(ss, p.Name)
	}
	return strings.Join(ss, ", ")
}

func (m Method) ParamsNamesExcludeCtx() string {
	ss := []string{}
	tmp := make(ParamsSlice, 0, 0)
	for _, p := range m.Params {
		if p.Type == "context.Context" {
			continue
		}

		tmp = append(tmp, p)
	}
	for _, p := range tmp {
		ss = append(ss, p.Name)
	}
	return strings.Join(ss, ", ")
}

func (m Method) ParamsExcludeCtx() ParamsSlice {
	tmp := make(ParamsSlice, 0, 0)
	for _, p := range m.Params {
		if p.Type == "context.Context" {
			continue
		}

		tmp = append(tmp, p)
	}
	return tmp
}

// ResultsNames returns a list of method results names
func (m Method) ResultsNames() string {
	ss := []string{}
	for _, r := range m.Results {
		ss = append(ss, r.Name)
	}
	return strings.Join(ss, ", ")
}

func (m Method) ResultsExcludeErr() ParamsSlice {
	tmp := make(ParamsSlice, 0, 0)
	for _, p := range m.Results {
		if p.Type == "error" {
			continue
		}

		tmp = append(tmp, p)
	}
	return tmp
}

// ParamsStruct returns a struct type with fields corresponding
// to the method params
func (m Method) ParamsStruct() string {
	ss := []string{}
	for _, p := range m.Params {
		if p.Variadic {
			ss = append(ss, p.Name+" "+strings.Replace(p.Type, "...", "[]", 1))
		} else {
			ss = append(ss, p.Name+" "+p.Type)
		}
	}
	return "struct{\n" + strings.Join(ss, "\n ") + "}"
}

func (m Method) ParamsStructExcludeCtx() string {
	ss := []string{}
	tmp := make(ParamsSlice, 0, 0)
	for _, p := range m.Params {
		if p.Type == "context.Context" {
			continue
		}

		tmp = append(tmp, p)
	}
	for _, p := range tmp {
		if p.Variadic {
			ss = append(ss, p.Name+" "+strings.Replace(p.Type, "...", "[]", 1))
		} else {
			ss = append(ss, p.Name+" "+p.Type)
		}
	}
	return "struct{\n" + strings.Join(ss, "\n ") + "}"
}

// ResultsStruct returns a struct type with fields corresponding
// to the method results
func (m Method) ResultsStruct() string {
	ss := []string{}
	for _, r := range m.Results {
		ss = append(ss, r.Name+" "+r.Type)
	}
	return "struct{\n" + strings.Join(ss, "\n ") + "}"
}

// ParamsMap returns a string representation of the map[string]interface{}
// filled with method's params
func (m Method) ParamsMap() string {
	ss := []string{}
	for _, p := range m.Params {
		ss = append(ss, `"`+p.Name+`": `+p.Name)
	}
	return "map[string]interface{}{\n" + strings.Join(ss, ",\n ") + "}"
}

func (m Method) ParamsMapExcludeCtx() string {
	ss := []string{}
	tmp := make(ParamsSlice, 0, 0)
	for _, p := range m.Params {
		if p.Type == "context.Context" {
			continue
		}

		tmp = append(tmp, p)
	}
	for _, p := range tmp {
		ss = append(ss, `"`+p.Name+`": `+p.Name)
	}
	return "map[string]interface{}{\n" + strings.Join(ss, ",\n ") + "}"
}

// ResultsMap returns a string representation of the map[string]interface{}
// filled with method's results
func (m Method) ResultsMap() string {
	ss := []string{}
	for _, r := range m.Results {
		ss = append(ss, `"`+r.Name+`": `+r.Name)
	}
	return "map[string]interface{}{\n" + strings.Join(ss, ",\n ") + "}"
}

func (m Method) ResultsMapErr2Str() string {
	ss := []string{}
	for _, r := range m.Results {
		if r.Type == "error" {
			ss = append(ss, `"`+r.Name+`": `+fmt.Sprintf(`fmt.Sprintf("%%v", %v)`, r.Name))
			continue
		}
		ss = append(ss, `"`+r.Name+`": `+r.Name)
	}
	return "map[string]interface{}{\n" + strings.Join(ss, ",\n ") + "}"
}

func (m Method) ResultsMapExcludeErr() string {
	ss := []string{}
	for _, r := range m.Results {
		if r.Type == "error" {
			continue
		}
		ss = append(ss, `"`+r.Name+`": `+r.Name)
	}
	return "map[string]interface{}{\n" + strings.Join(ss, ",\n ") + "}"
}

// HasParams returns true if method has params
func (m Method) HasParams() bool {
	return len(m.Params) > 0
}

// HasResults returns true if method has results
func (m Method) HasResults() bool {
	return len(m.Results) > 0
}

func (m Method) HasResultsExcludeErr() bool {
	return len(m.ResultsExcludeErr()) > 0
}

// ReturnStruct returns return statement with the return params
// taken from the structName
func (m Method) ReturnStruct(structName string) string {
	if len(m.Results) == 0 {
		return "return"
	}

	ss := []string{}
	for _, r := range m.Results {
		ss = append(ss, structName+"."+r.Name)
	}
	return "return " + strings.Join(ss, ", ")
}

func (m Method) DocContains(s ...string) bool {
	for _, v := range m.Doc {
		vl := strings.Split(v, " ")
		sLen := len(s)
		if len(vl) >= sLen+1 && strings.Join(vl[1:sLen+1], " ") == strings.Join(s, " ") {
			return true
		}
		return false
	}
	return false
}

// Signature returns comma separated method's params followed by the comma separated
// method's results
func (m Method) Signature() string {
	params := []string{}
	for _, p := range m.Params {
		params = append(params, p.Name+" "+p.Type)
	}

	results := []string{}
	for _, r := range m.Results {
		results = append(results, r.Name+" "+r.Type)
	}

	return "(" + strings.Join(params, ", ") + ") (" + strings.Join(results, ", ") + ")"
}

// Declaration returns a method name followed by it's signature
func (m Method) Declaration() string {
	return m.Name + m.Signature()
}

func (m Method) SwagFieldData() string {
	if len(m.ResultsMapExcludeErr()) == 0 {
		return "data=string"
	}
	if len(m.ResultsExcludeErr()) == 1 {
		return "data=" + m.Results[0].Type
	}
	var s []string
	for _, v := range m.ResultsExcludeErr() {
		s = append(s, "data."+v.Name+"="+v.Type)
	}
	return strings.Join(s, ",")
}

func (m Method) ResultsExcludeErrCode() *jen.Statement {
	var code []jen.Code

	for _, v := range m.ResultsExcludeErr() {
		code = append(code, jen.Id(v.Name).Id(v.Type))
	}

	return jen.Var().Defs(code...)
}

func (m Method) ResultsNamesCode() *jen.Statement {
	var l jen.Statement
	for _, v := range m.Results {
		l = append(l, jen.Id(v.Name))
	}

	return jen.List(l...)
}

func (m Method) ParamsNamesCode() *jen.Statement {
	var l jen.Statement

	l = append(l, jen.Id("ctx"))
	if m.KitRequest.RequestIsBody {
		l = append(l, jen.Id("req"))
		return jen.List(l...)
	}

	for _, v := range m.ParamsExcludeCtx() {
		l = append(l, jen.Id("req."+m.KitRequest.ParamPath(v.Name)))
	}

	return jen.List(l...)
}

func (m Method) ReturnCode() *jen.Statement {
	var code *jen.Statement
	index1ResultName := m.ResultsExcludeErr()[0].Name
	result1DataCode := jen.Id(index1ResultName)

	if len(m.ResultsExcludeErr()) == 1 {
		code = jen.Return(jen.Id("encode.Response").Values(
			jen.Dict{
				jen.Id("Data"):  result1DataCode,
				jen.Id("Error"): jen.Id("err"),
			},
		), jen.Id("err"))
	} else {
		valueDick := make(jen.Dict)
		valueDick[jen.Lit(m.ResultsExcludeErr()[0].Name)] = result1DataCode
		for _, v := range m.ResultsExcludeErr()[1:] {
			valueDick[jen.Lit(v.Name)] = jen.Id(v.Name)
		}
		dataMap := jen.Map(jen.String()).Interface().Values(valueDick)

		code = jen.Return(jen.Id("encode.Response").Values(
			jen.Dict{
				jen.Id("Data"):  dataMap,
				jen.Id("Error"): jen.Id("err"),
			},
		), jen.Id("err"))
	}

	return code

}

func (m Method) MakeEndpoint() string {
	return jen.Func().Id("make" + upFirst(m.Name) + "Endpoint").Params(jen.Id("s").Id("Service")).Id("endpoint.Endpoint").Block(
		jen.Return(jen.Func().Params(jen.Id("ctx").Id("context.Context"), jen.Id("request").Interface()).Params(jen.Id("response").Interface(), jen.Id("err").Error()).Block(
			jen.Id("req").Op(":=").Id("request.").Params(jen.Id(m.KitRequest.RequestName)),
			m.ResultsExcludeErrCode(),
			m.ResultsNamesCode().Op("=").Id("s").Dot(m.Name).Call(m.ParamsNamesCode()),
			m.ReturnCode(),
		))).GoString()
}
