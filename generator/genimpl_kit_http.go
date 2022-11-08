package generator

import (
	"bytes"
	"fmt"
	"github.com/fitan/jennifer/jen"
	"github.com/pkg/errors"
	"log"
	"path"
	"strings"
	"text/template"
)

const name = "kitHttp"
const implTags = "@tags"
const implBasePath = "@basePath"
const kitHttpRouterMark = "@kit-http"
const kitHttpRequestMark = "@kit-http-request"
const kitHttpEndpointMark = "@kit-http-endpoint"
const kitHttpEndpointWrapMark = "@kit-http-endpoint-wrap"
const kitHttpEncodeMark = "@kit-http-encode"
const kitHttpDecodeMark = "@kit-http-decode"

const kitHttpSwagMark = "@swag"
const httpJenFName = "http"
const myHttpJenFName = "myHttp"
const endpointJenFName = "endpoint"
const logJenFName = "log"
const myLogJenFName = "myLog"
const tracingJenFName = "tracing"
const myTracingJenFName = "myTrace"
const myEndpointJenFName = "myEndpoint"

type GenImplKitHttp struct {
	genImpl *GenImpl
	//jenF *jen.File
	jenFM map[string]*jen.File
	//impl Impl
	implConf map[string]*kitHttpConf
}

func NewGenImplKitHttp(impl *GenImpl) *GenImplKitHttp {
	return &GenImplKitHttp{
		genImpl:  impl,
		jenFM:    make(map[string]*jen.File, 0),
		implConf: make(map[string]*kitHttpConf, 0),
	}
}

func (g *GenImplKitHttp) Name() string {
	return name
}

func (g *GenImplKitHttp) Gen() error {
	return g.genJenF()
}

func (g *GenImplKitHttp) JenF(name string) *jen.File {
	return g.jenFM[name]
}

func (g *GenImplKitHttp) genJenF() error {
	methodNameList := make([]string, 0)
	handlerCodeList := make([]jen.Code, 0)
	myHandlerCodeList := make([]jen.Code, 0)
	decodeRequestCodeList := make([]jen.Code, 0)

	var EndpointsConstCode jen.Code
	var EndpointsCode jen.Code
	var NewEndpointsCode jen.Code
	var MakeEndpointCodeList []jen.Code
	var MakeMyEndpointCodeList []jen.Code

	var LoggingFuncCodeList []jen.Code

	var TracingFuncCodeList []jen.Code

	for implName, impl := range g.genImpl.ImplList {
		g.implConf[implName] = NewKitHttpConf(impl)

		for _, m := range impl.Methods {
			conform, err := g.implConf[implName].MethodConform(m.Name)
			if err != nil {
				return err
			}
			if !conform {
				log.Printf("methodHttpMethod %s not conform", m.Name)
				continue
			}

			methodNameList = append(methodNameList, m.Name)

			methodHttpPath, _ := g.implConf[implName].MethodHttpPath(m.Name)
			methodHttpMethod, _ := g.implConf[implName].MethodHttpMethod(m.Name)
			annotation, _ := g.implConf[implName].MethodAnnotation(m.Name)
			requestName, requestBody, _ := g.implConf[implName].MethodHttpRequest(m.Name)
			enableSwag, _ := g.implConf[implName].EnableSwag(m.Name)
			tags := g.implConf[implName].implTags
			methodConf,_ := g.implConf[implName].MethodConf(m.Name)

			handlerCodeList = append(
				handlerCodeList, genFuncMakeHTTPHandlerHandler(m.Name, methodConf),
			)
			myHandlerCodeList = append(
				myHandlerCodeList, genFuncMyMakeHTTPHandlerHandler(m.Name, methodConf),
			)

			r := NewKitRequest(g.genImpl.GenOption.Pkg, m.Name, requestName, requestBody)
			r.ParseRequest()

			vars := swagVars{
				MethodName:       m.Name,
				MethodHttpPath:   methodHttpPath,
				MethodHttpMethod: methodHttpMethod,
				EnableSwag:       enableSwag,
				Annotation:       annotation,
				Tags:             tags,
				KitRequest:       r,
				ImplMethod:       m,
			}

			swagStr, err := g.swag(vars)
			if err != nil {
				return errors.Wrap(err, "swag")
			}

			decodeRequestCodeList = append(decodeRequestCodeList, jen.Comment(swagStr).Add(jen.Line()).Add(r.Statement().Line()))

			MakeEndpointCodeList = append(MakeEndpointCodeList, genMakeEndpoint(requestName, m, r, g.genImpl.GenOption))

			MakeMyEndpointCodeList = append(MakeMyEndpointCodeList, genMyMakeEndpoint(m, methodConf,r, g.genImpl.GenOption))

			LoggingFuncCodeList = append(LoggingFuncCodeList, genLoggingFunc(m))

			TracingFuncCodeList = append(TracingFuncCodeList, genTracingFunc(g.genImpl.GenOption.CutLast2DirName(), m))
		}
	}

	var methodList []ImplMethod
	for _, impl := range g.genImpl.ImplList {
		for _, m := range impl.Methods {
			methodList = append(methodList, m)
		}
	}

	h := jen.Statement(handlerCodeList)
	makeHttpCode := genFuncMakeHTTPHandler(genFuncMakeHTTPHandlerNewEndpoint(methodNameList), &h)

	httpJenF := jen.NewFile(g.genImpl.GenOption.Pkg.Name)
	JenFAddImports(g.genImpl.GenOption.Pkg, httpJenF)
	httpJenF.Add(makeHttpCode)
	httpJenF.Add(decodeRequestCodeList...)

	myHttpJenF := jen.NewFile(g.genImpl.GenOption.Pkg.Name)
	JenFAddImports(g.genImpl.GenOption.Pkg, myHttpJenF)
	myH := jen.Statement(myHandlerCodeList)
	myHttpJenF.Add(myExtraHttp(methodList, &myH))
	myHttpJenF.Add(decodeRequestCodeList...)

	EndpointsConstCode = genEndpointConst(methodNameList)
	EndpointsCode = genEndpoints(methodNameList)
	NewEndpointsCode = genNewEndpoint(methodNameList)

	endpointJenF := jen.NewFile(g.genImpl.GenOption.Pkg.Name)
	JenFAddImports(g.genImpl.GenOption.Pkg, endpointJenF)
	endpointJenF.Add(EndpointsConstCode)
	endpointJenF.Add(EndpointsCode)
	endpointJenF.Add(NewEndpointsCode)
	endpointJenF.Add(MakeEndpointCodeList...)

	myEndpointJenF := jen.NewFile(g.genImpl.GenOption.Pkg.Name)
	JenFAddImports(g.genImpl.GenOption.Pkg, myEndpointJenF)
	myEndpointJenF.Add(EndpointsConstCode)
	myEndpointJenF.Add(EndpointsCode)
	myEndpointJenF.Add(NewEndpointsCode)
	myEndpointJenF.Add(MakeMyEndpointCodeList...)
	myEndpointJenF.Add(myExtraEndpoint())

	logJenF := jen.NewFile(g.genImpl.GenOption.Pkg.Name)
	JenFAddImports(g.genImpl.GenOption.Pkg, logJenF)
	logJenF.Add(genLoggingStruct())
	logJenF.Add(LoggingFuncCodeList...)
	logJenF.Add(genNewLogging(g.genImpl.GenOption.CutLast2DirName()))

	myLogJenF := jen.NewFile(g.genImpl.GenOption.Pkg.Name)
	JenFAddImports(g.genImpl.GenOption.Pkg, myLogJenF)
	myLogJenF.Add(genMyLogging(methodList))

	tracingJenF := jen.NewFile(g.genImpl.GenOption.Pkg.Name)
	tracingJenF.ImportName("github.com/opentracing/opentracing-go", "opentracing")
	JenFAddImports(g.genImpl.GenOption.Pkg, tracingJenF)
	tracingJenF.Add(genTracingStruct())
	tracingJenF.Add(TracingFuncCodeList...)
	tracingJenF.Add(genNewTracing())

	myTracingJenF := jen.NewFile(g.genImpl.GenOption.Pkg.Name)
	JenFAddImports(g.genImpl.GenOption.Pkg, myTracingJenF)
	myTracingJenF.Add(genMyKitTrace(g.genImpl.GenOption.CutLast2DirName(), methodList))

	g.jenFM[httpJenFName] = httpJenF
	g.jenFM[endpointJenFName] = endpointJenF
	g.jenFM[logJenFName] = logJenF
	g.jenFM[tracingJenFName] = tracingJenF
	g.jenFM[myTracingJenFName] = myTracingJenF
	g.jenFM[myEndpointJenFName] = myEndpointJenF
	g.jenFM[myHttpJenFName] = myHttpJenF
	g.jenFM[myLogJenFName] = myLogJenF

	return nil
}

type swagVars struct {
	MethodName       string
	MethodHttpPath   string
	MethodHttpMethod string
	EnableSwag       bool
	Annotation       string
	Tags             string
	KitRequest       *KitRequest
	ImplMethod       ImplMethod
}

func (g *GenImplKitHttp) swag(vars swagVars) (string, error) {
	doc := `
{{if $.EnableSwag}}
{{$.KitRequest.ServiceName}}
@Summary {{$.Annotation}}
@Description {{$.Annotation}}
{{$.Tags}}
@Accept json
@Produce json
{{- range $k,$v := $.KitRequest.Path}}
@Param {{$v.ParamName}} path string true {{$v.Annotations}}
{{- end}}
{{- range $k, $v := $.KitRequest.Query}}
@Param {{$v.ParamName}} query string false {{$v.Annotations}}
{{- end}}
{{- range $k, $v := $.KitRequest.Header}}
@Param {{$v.ParamName}} header string false {{$v.Annotations}}
{{- end}}
{{- if $.KitRequest.RequestIsBody}}
@Param {{$.KitRequest.RequestName}} body {{$.KitRequest.RequestName}} true "http request body"
{{- else}}
{{- range $k, $v := $.KitRequest.Body}}
@Param {{$v.ParamName}} body {{$v.ParamTypeName}} true {{$v.Annotations}}
{{- end}}
{{- end}}
@Success 200 {object} encode.Response{ {{- $.ImplMethod.SwagRespObjData}}}
@Router {{$.MethodHttpPath}} [{{$.MethodHttpMethod}}]{{end}}`
	t, err := template.New("doc").Parse(doc)
	if err != nil {
		return "", err
	}
	w := bytes.NewBuffer([]byte{})
	err = t.Execute(w, vars)
	if err != nil {
		return "", err
	}

	return w.String(), nil
}

type kitHttpConf struct {
	impl         Impl
	implDoc      *AstDocFormat
	methodDocM   map[string]*AstDocFormat
	implBasePath string
	implTags     string
}

type MethodConf struct {
	Method      string
	Path        string
	Request     string
	Endpoint string
	EndpointWrap string
	Decode string
	Encode string
	RequestBody bool
	EnableSwag  bool
	Annotation  string
}

func NewKitHttpConf(impl Impl) *kitHttpConf {
	return &kitHttpConf{
		impl: impl,
	}
}

func (k *kitHttpConf) parse() {
	docF := NewAstDocFormat(k.impl.Doc)
	docF.MarkValuesMapping(implTags, &k.implTags)
	docF.MarkValuesMapping(implBasePath, &k.implBasePath)
}

func (k *kitHttpConf) getMethod(name string) (ImplMethod, error) {
	for _, m := range k.impl.Methods {
		if m.Name == name {
			return m, nil
		}
	}
	return ImplMethod{}, fmt.Errorf("method %s not found", name)
}

// 方法的注释 默认取方法名后面的注释 如果没有则取注释的第一行
func (k *kitHttpConf) MethodAnnotation(name string) (string, error) {
	m, err := k.getMethod(name)
	if err != nil {
		return "", err
	}

	docF := NewAstDocFormat(m.Comment)
	var annotation string
	docF.MarkValuesMapping(name, &annotation)
	if annotation == "" {
		annotation = strings.TrimPrefix(docF.doc.List[0].Text, "// ")
	}
	return annotation, nil
}

func (k *kitHttpConf) MethodConform(name string) (bool, error) {
	conf, err := k.MethodConf(name)
	if err != nil {
		return false, err
	}

	if conf.Path == "" || conf.Method == "" {
		return false, fmt.Errorf("method %s not found param path: %s, method: %s, request %s", name, conf.Path, conf.Method, conf.Request)
	}

	return true, nil
}

func (k *kitHttpConf) MethodConf(name string) (res MethodConf, err error) {
	m, err := k.getMethod(name)
	if err != nil {
		return
	}

	var path string
	var method string
	var request string
	var requestBody string
	var endpoint string
	var endpointWrap string
	var decode string
	var encode string
	var enableSwag string
	var annotation string
	docF := NewAstDocFormat(m.Comment)
	docF.MarkValuesMapping(kitHttpRouterMark, &path, &method)
	docF.MarkValuesMapping(kitHttpRequestMark, &request, &requestBody)
	docF.MarkValuesMapping(kitHttpSwagMark, &enableSwag)
	docF.MarkValuesMapping(kitHttpEndpointMark, &endpoint)
	docF.MarkValuesMapping(kitHttpDecodeMark, &decode)
	docF.MarkValuesMapping(kitHttpEncodeMark, &encode)
	docF.MarkValuesMapping(kitHttpEndpointWrapMark, &endpointWrap)
	annotation,_ = k.MethodAnnotation(name)

	return MethodConf{
		Method:      strings.ToUpper(method),
		Path:        path,
		Request:     request,
		Endpoint:    encode,
		EndpointWrap: endpointWrap,
		Decode:      decode,
		Encode:      encode,
		RequestBody: requestBody != "",
		EnableSwag:  enableSwag != "false",
		Annotation: annotation,
	}, nil
}

func (k *kitHttpConf) MethodHttpPath(name string) (string, error) {
	conf, err := k.MethodConf(name)
	if err != nil {
		return "", err
	}

	return path.Join(k.implBasePath, conf.Path), nil
}

func (k *kitHttpConf) MethodHttpMethod(name string) (string, error) {
	conf, err := k.MethodConf(name)
	if err != nil {
		return "", err
	}

	return conf.Method, nil

}

func (k *kitHttpConf) MethodHttpRequest(name string) (string, bool, error) {
	conf, err := k.MethodConf(name)
	if err != nil {
		return "", false, err
	}

	return conf.Request, conf.RequestBody, nil
}

func (k *kitHttpConf) EnableSwag(name string) (bool, error) {
	conf, err := k.MethodConf(name)
	if err != nil {
		return false, err
	}

	return conf.EnableSwag, nil
}

func (k *kitHttpConf) Tags() string {
	return k.implTags
}
