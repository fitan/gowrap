package generator

import "github.com/fitan/jennifer/jen"

// myEndpoint

func myExtraEndpoint(methodList []ImplMethod) jen.Code {
	return jen.Type().Id("Mws").Map(jen.String()).Index().Id("endpoint.Middleware").Line().
		Func().Id("AllMethodAddEmw").Params(jen.Id("mw").Id("map[string][]endpoint.Middleware"), jen.Id("m").Id("endpoint.Middleware")).Block(
		jen.Id("methods").Op(":=").Index().String().ValuesFunc(
			func(group *jen.Group) {
				for _, m := range methodList {
					group.Id(m.Name + "MethodName")
				}
			}),
		jen.For(jen.List(jen.Id("_"), jen.Id("v")).Op(":=").Range().Id("methods")).Block(
			jen.Id("mw").Index(jen.Id("v")).Op("=").Append(jen.Id("mw").Index(jen.Id("v")), jen.Id("m")),
		),
	)
}

// myHttp
func myExtraHttp(methodList []ImplMethod) jen.Code {
	return jen.Type().Id("Ops").Map(jen.String()).Index().Id("http.ServerOption").Line().
		Func().Id("AllMethodAddOps").Params(jen.Id("options").Id("map[string][]http.ServerOption"), jen.Id("option").Id("http.ServerOption")).Block(
		jen.Id("methods").Op(":=").Index().String().ValuesFunc(
			func(group *jen.Group) {
				for _, m := range methodList {
					group.Id(m.Name + "MethodName")
				}
			}),
		jen.For(jen.List(jen.Id("_"), jen.Id("v")).Op(":=").Range().Id("methods")).Block(
			jen.Id("options").Index(jen.Id("v")).Op("=").Append(jen.Id("options").Index(jen.Id("v")), jen.Id("option")),
		),
	)
}

// myTrace
func genMyKitTrace(tracingPrefix string, methodList []ImplMethod) jen.Code {
	code := jen.Null()
	code.Type().Id("tracing").Struct(jen.Id("next").Id("Service")).Line()

	for _, method := range methodList {

		methodParamCode := make([]jen.Code, 0)
		methodParamCode = append(methodParamCode, jen.Id("ctx").Qual("context", "Context"))
		methodResultCode := make([]jen.Code, 0)

		tracingParamCode := make([]jen.Code, 0)
		nextMethodParamCode := make([]jen.Code, 0)
		for _, param := range method.Params {
			nextMethodParamCode = append(nextMethodParamCode, jen.Id(param.Name))
		}

		for _, param := range method.ParamsExcludeCtx() {
			methodParamCode = append(methodParamCode, jen.Id(param.Name).Id(param.ID))
			tracingParamCode = append(tracingParamCode, jen.Lit(param.Name), jen.Id(param.Name))
		}

		for _, param := range method.Results {
			methodResultCode = append(methodResultCode, jen.Id(param.Name).Id(param.ID))
		}

		code.Line()

		code.Func().Params(jen.Id("s").Op("*").Id("tracing")).Id(method.Name).Params(
			methodParamCode...,
		).Params(
			methodResultCode...,
		).Block(
			jen.List(jen.Id("_"), jen.Id("span")).Op(":=").Qual("go.opentelemetry.io/otel", "Tracer").
				Call(jen.Lit(tracingPrefix)).Dot("Start").Call(jen.Id("ctx"), jen.Lit(method.Name)).Line(),
			jen.Defer().Func().Params().Block(
				jen.Id("paramIn").Op(":=").Map(jen.String()).Interface().Values(
					jen.DictFunc(
						func(dict jen.Dict) {
							for _, param := range method.ParamsExcludeCtx() {
								dict[jen.Lit(param.Name)] = jen.Id(param.Name)
							}
						}),
				),
				jen.List(jen.Id("paramInJsonB"), jen.Id("_")).Op(":=").Qual("encoding/json", "Marshal").Call(jen.Id("paramIn")),
				jen.Id("span").Dot("AddEvent").Call(
					jen.Lit("paramIn"),
					jen.Qual("go.opentelemetry.io/otel/trace", "WithAttributes").Call(
						jen.Qual("go.opentelemetry.io/otel/attribute", "String").Call(
							jen.Lit("param list"),
							jen.Id("string").Call(jen.Id("paramInJsonB"))),
					),
				),
				func() jen.Code {
					c := jen.Null()
					c.If(jen.Err().Op("!=").Nil()).Block(
						jen.Id("span").Dot("SetStatus").Call(jen.Qual("go.opentelemetry.io/otel/codes", "Error"), jen.Lit(method.Name+" error")),
						jen.Id("span").Dot("RecordError").Call(jen.Err()),
					)
					return c
				}(),
				jen.Id("span").Dot("End").Call(),
			).Call().Line(),
			jen.Return().Id("s").Dot("next").Dot(method.Name).Call(nextMethodParamCode...),
		)

	}

	code.Line()

	code.Func().Id("NewTracing").Params().Params(jen.Id("Middleware")).Block(
		jen.Return().Func().Params(jen.Id("next").Id("Service")).Params(
			jen.Id("Service")).Block(jen.Return().Op("&").Id("tracing").Values(jen.Id("next").Op(":").Id("next")))).Line()

	code.Line().Line()

	return code
}
