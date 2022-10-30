package generator

import (
	"github.com/fitan/jennifer/jen"
)

// myEndpoint

func myExtraEndpoint(methodList []ImplMethod) jen.Code {
	return jen.Type().Id("Mws").Map(jen.String()).Index().Id("endpoint.Middleware").Line().
		Func().Id("AllMethodAddMws").Params(jen.Id("mw").Id("Mws"), jen.Id("m").Id("endpoint.Middleware")).Block(
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
func genFuncMyMakeHTTPHandlerHandler(name, path, method, annotation string) jen.Code {
	return jen.Id("r").Dot("Handle").Call(
		jen.Lit(path),
		jen.Qual("github.com/go-kit/kit/transport/http", "NewServer").Call(
			jen.Id("eps").Dot(name+"Endpoint"),
			jen.Id("decode"+name+"Request"),
			jen.Qual("github.com/go-kit/kit/transport/http", "EncodeJSONResponse"),
			jen.Id("ops").Index(jen.Id(name+"MethodName")).Op("..."))).Dot("Methods").Call(jen.Lit(method)).Dot("Name").Call(jen.Lit(annotation)).Line()
}

func myExtraHttp(methodList []ImplMethod, handlerList jen.Code) jen.Code {
	code := jen.Null()
	code.Type().Id("Handler").Struct().Line()
	code.Func().Id("MakeHTTPHandler").Params(
		jen.Id("r").Op("*").Qual("github.com/gorilla/mux", "Router"),
		jen.Id("s").Id("Service"),
		jen.Id("mws").Id("Mws"),
		jen.Id("ops").Id("Ops"),
	).Params(
		jen.Id("Handler"),
	).Block(
		jen.Id("eps").Op(":=").Id("NewEndpoint").Call(jen.Id("s"), jen.Id("mws")),
		handlerList,
		jen.Return(jen.Id("Handler").Values()),
	).Line()

	code.Type().Id("Ops").Map(jen.String()).Index().Id("http.ServerOption").Line()
	code.Func().Id("AllMethodAddOps").Params(jen.Id("options").Id("map[string][]http.ServerOption"), jen.Id("option").Id("http.ServerOption")).Block(
		jen.Id("methods").Op(":=").Index().String().ValuesFunc(
			func(group *jen.Group) {
				for _, m := range methodList {
					group.Id(m.Name + "MethodName")
				}
			}),
		jen.For(jen.List(jen.Id("_"), jen.Id("v")).Op(":=").Range().Id("methods")).Block(
			jen.Id("options").Index(jen.Id("v")).Op("=").Append(jen.Id("options").Index(jen.Id("v")), jen.Id("option")),
		),
	).Line()

	//code.Type().Id("Mws").Id("map[string][]endpoint.Middleware").Line()
	//code.Func().Id("AllMethodAddMws").Params(jen.Id("mws").Id("map[string][]endpoint.Middleware"), jen.Id("mw").Id("endpoint.Middleware")).Block(
	//	jen.Id("methods").Op(":=").Index().String().ValuesFunc(
	//		func(group *jen.Group) {
	//			for _, m := range methodList {
	//				group.Id(m.Name + "MethodName")
	//			}
	//		}),
	//	jen.For(jen.List(jen.Id("_"), jen.Id("v")).Op(":=").Range().Id("methods")).Block(
	//		jen.Id("mws").Index(jen.Id("v")).Op("=").Append(jen.Id("mws").Index(jen.Id("v")), jen.Id("mw")),
	//	),
	//).Line()

	return code
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

// myLog

func genMyLogging(methodList []ImplMethod) jen.Code {
	code := jen.Null()
	code.Type().Id("logging").Struct(jen.Id("logger").Op("*").Qual("go.uber.org/zap", "SugaredLogger"), jen.Id("next").Id("Service")).Line()

	methodParamCode := make([]jen.Code, 0)
	methodResultCode := make([]jen.Code, 0)

	logParamCode := make([]jen.Code, 0)
	nextMethodParamCode := make([]jen.Code, 0)
	for _, method := range methodList {

		for _, param := range method.Params {
			if param.ID == "context.Context" {
				methodParamCode = append(methodParamCode, jen.Id(param.Name).Qual("context", "Context"))
			} else {
				methodParamCode = append(methodParamCode, jen.Id(param.Name).Id(param.ID))
			}
			nextMethodParamCode = append(nextMethodParamCode, jen.Id(param.Name))
		}

		for _, param := range method.ParamsExcludeCtx() {
			logParamCode = append(logParamCode, jen.Lit(param.Name), jen.Id(param.Name))
		}

		logParamSatement := jen.List(logParamCode...)

		for _, param := range method.Results {
			methodResultCode = append(methodResultCode, jen.Id(param.Name).Id(param.ID))
		}

		code.Func().Params(
			jen.Id("s").Op("*").Id("logging"),
		).Id(method.Name).Params(
			methodParamCode...,
		).Params(
			methodResultCode...,
		).Block(
			jen.Defer().Func().Params(
				jen.Id("begin").Qual("time", "Time"),
			).
				Block(
					func() jen.Code {
						v := jen.Null()

						if method.ReturnsError {
							v.If(jen.Err().Op("!=").Nil()).Block(
								jen.Id("s").Dot("logger").Dot("Errorw").Call(
									jen.Lit(method.Name+" error"),
									jen.Lit("error"),
									jen.Err(),
									logParamSatement,
									jen.Lit("took"),
									jen.Qual("time", "Since").Call(jen.Id("begin")),
									jen.Lit("traceId"),
									jen.Qual(
										"go.opentelemetry.io/otel/trace",
										"SpanContextFromContext",
									).Call(jen.Id("ctx")).Dot("TraceID().String()"),
								),
							).Else().Block(
								jen.Id("s").Dot("logger").Dot("Infow").Call(
									jen.Lit(method.Name+" success"),
									logParamSatement,
									jen.Lit("took"),
									jen.Qual("time", "Since").Call(jen.Id("begin")),
									jen.Lit("traceId"),
									jen.Qual(
										"go.opentelemetry.io/otel/trace",
										"SpanContextFromContext",
									).Call(jen.Id("ctx")).Dot("TraceID().String()"),
								),
							)
						} else {
							v.Id("s").Dot("logger").Dot("Infow").Call(
								jen.Lit(method.Name+" success"),
								logParamSatement,
								jen.Lit("took"),
								jen.Qual("time", "Since").Call(jen.Id("begin")),
								jen.Lit("traceId"),
								jen.Qual(
									"go.opentelemetry.io/otel/trace",
									"SpanContextFromContext",
								).Call(jen.Id("ctx")).Dot("TraceID().String()"),
							)
						}

						return v
					}(),
				).Call(jen.Qual("time", "Now").Call()),
			jen.Return(jen.Id("s").Dot("next").Dot(method.Name).Call(nextMethodParamCode...)),
		).Line()

	}
	code.Func().Id("NewLogging").Params(
		jen.Id("logger").Op("*").Qual("go.uber.org/zap", "SugaredLogger"),
	).Params(
		jen.Id("Middleware"),
	).Block(
		jen.Return().Func().Params(jen.Id("next").Id("Service")).Params(jen.Id("Service")).Block(
			jen.Return().Op("&").Id("logging").Values(jen.Id("logger").Op(":").Id("logger"),
				jen.Id("next").Op(":").Id("next"),
			),
		))

	return code
}
