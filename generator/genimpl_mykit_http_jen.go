package generator

import "github.com/fitan/jennifer/jen"

// Mytrace
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