package generator

import jen "github.com/dave/jennifer/jen"

// http

func genFuncMakeHTTPHandlerNewEndpoint(methodNameList []string) jen.Code {
	l := make([]jen.Code, 0, len(methodNameList))
	for _, methodName := range methodNameList {
		l = append(l, jen.Id(methodName).Op(":").Id("ems"))
	}
	return jen.Id("eps").Op(":=").Id("NewEndpoint").Call(jen.Id("s"), jen.Map(jen.Id("string")).Index().Qual("github.com/go-kit/kit/endpoint", "Middleware").Values(
		l...,
	))
}

func genFuncMakeHTTPHandlerHandler(name, path, method, annotation string) jen.Code {
	return jen.Id("r").Dot("Handle").Call(jen.Lit(path), jen.Qual("github.com/go-kit/kit/transport/http", "NewServer").Call(
		jen.Id("eps").Dot(name + "Endpoint"),
		jen.Id("decodeQueryRangeRequest"),
		jen.Id("encode").Dot("JsonResponse"),
		jen.Id("opts").Op("..."))).Dot("Methods").Call(jen.Lit(method)).Dot("Name").Call(jen.Lit(annotation))
}

func genFuncMakeHTTPHandler(newEndpoint jen.Code, HandlerList jen.Code) jen.Code {
	return jen.Func().Id("MakeHTTPHandler").Params(
		jen.Id("s").Id("Service"),
		jen.Id("dmw").Index().Qual("github.com/go-kit/kit/endpoint", "Middleware"),
		jen.Id("opts").Index().Qual("github.com/go-kit/kit/transport/http", "ServerOption")).
		Params(jen.Qual("net/http", "Handler")).Block(
			jen.Null().Var().Id("ems").Index().Qual("github.com/go-kit/kit/endpoint", "Middleware"),
			jen.Id("opts").Op("=").Id("append").Call(jen.Id("opts"), jen.Qual("github.com/go-kit/kit/transport/http", "ServerBefore").Call(
			jen.Func().Params(jen.Id("ctx").Qual("context", "Context"),
			jen.Id("request").Op("*").Qual("net/http", "Request")).Params(jen.Qual("context", "Context")).Block(jen.Return().Id("ctx")))),
			jen.Id("ems").Op("=").Id("append").Call(jen.Id("ems"), jen.Id("dmw").Op("...")),
			newEndpoint,
			jen.Id("r").Op(":=").Id("mux").Dot("NewRouter").Call(),
			HandlerList,
			jen.Return().Id("r"))
}


// endpoint

func genEndpointConst(methodNameList []string) jen.Code {
	j := jen.Null()

	for _, methodName := range methodNameList {
		j.Var().Id(methodName + "MethodName").Op("=").Lit(methodName)
	}
	return j
}
func genEndpoints(methodNameList []string) jen.Code {
	listCode := make([]jen.Code, 0, len(methodNameList))
	for _, methodName := range methodNameList {
		listCode = append(listCode, jen.Id(methodName+"Endpoint").Qual("github.com/go-kit/kit/endpoint", "Endpoint"))
	}
	return jen.Null().Type().Id("Endpoints").Struct(
		listCode...
	)
}
func genNewEndpoint(methodNameList []string) jen.Code {
	endpointVarList := make([]jen.Code, 0, len(methodNameList))
	endpointForList := make([]jen.Code, 0, len(methodNameList))

	for _, methodName := range methodNameList {
		endpointVarList = append(endpointVarList, jen.Id(methodName + "Endpoint").Op(":").Id("make"+ methodName +"Endpoint").Call(jen.Id("s")))

		endpointForList = append(endpointForList, jen.For(jen.List(jen.Id("_"), jen.Id("m")).Op(":=").Range().Id("dmw").Index(jen.Id(methodName + "MethodName"))).Block(jen.Id("eps").Dot(methodName + "Endpoint").Op("=").Id("m").Call(jen.Id("eps").Dot(methodName + "Endpoint"))))
	}

	endpointForListStatement := jen.Statement(endpointForList)

	return jen.Func().Id("NewEndpoint").Params(jen.Id("s").Id("Service"), jen.Id("dmw").Map(jen.Id("string")).Index().Qual("github.com/go-kit/kit/endpoint", "Middleware")).Params(jen.Id("Endpoints")).Block(
		jen.Id("eps").Op(":=").Id("Endpoints").Values(
			endpointVarList...,
		),
		&endpointForListStatement,
		jen.Return().Id("eps"),
	)
}
func genMakeEndpoint(method ImplMethod, request *KitRequest) jen.Code {
	paramList := make([]jen.Code, 0, len(method.ParamsExcludeCtx()))
	paramList = append(paramList, jen.Id("ctx").Qual("context", "Context"))
	resultNameList := make([]jen.Code, 0, len(method.Results))
	endpointVarList := jen.Null()
	for _, param := range method.ParamsExcludeCtx() {
		endpointVarList.Var().Id(param.Name).Id(param.Type.String())
		paramList = append(paramList, jen.Id("req" + request.ParamPath(param.Name)))
		resultNameList = append(resultNameList, jen.Id(param.Name))
	}

	var responseDataID string

	if len(method.ResultsExcludeErr()) == 1 {
		responseDataID = method.ResultsExcludeErr()[0].Name
	} else {
		responseDataID = method.ResultsMapExcludeErr()
	}

	return jen.Func().Id("make"+method.Name + "Endpoint").Params(jen.Id("s").Id("Service")).Params(jen.Qual("github.com/go-kit/kit/endpoint", "Endpoint")).Block(jen.Return().Func().Params(jen.Id("ctx").Qual("context", "Context"), jen.Id("request").Interface()).Params(jen.Id("response").Interface(), jen.Id("err").Id("error")).Block(
		jen.Id("req").Op(":=").Id("request").Assert(jen.Id(method.Name + "Request")),
		endpointVarList,
		jen.List(resultNameList...).Op("=").Id("s").Dot(method.Name).Call(
			paramList...,
		),
		jen.Return().List(jen.Id("encode").Dot("Response").Values(
			jen.Id("Data").Op(":").Id(responseDataID),
			jen.Id("Error").Op(":").Id("err"),
		),
		jen.Id("err"))))
}




// logging

func genLoggingStruct() jen.Code {
	return jen.Null().Type().Id("logging").Struct(jen.Id("logger").Id("log").Dot("Logger"), jen.Id("next").Id("Service"), jen.Id("traceId").Id("string"))
}
func genLoggingFunc(method ImplMethod) jen.Code {

	methodParamCode := make([]jen.Code, 0)
	methodResultCode := make([]jen.Code, 0)

	logParamCode := make([]jen.Code, 0)
	nextMethodParamCode := make([]jen.Code, 0)

	for _, param := range method.Params {
		methodParamCode = append(methodParamCode, jen.Id(param.Name).Id(param.Type.String()))
		nextMethodParamCode = append(nextMethodParamCode, jen.Id(param.Name))
	}

	for _, param := range method.ParamsExcludeCtx() {
		logParamCode = append(logParamCode, jen.Lit(param.Name), jen.Id(param.Name))
	}

	logParamSatement := jen.Statement(logParamCode)

	for _, param := range method.Results {
		methodResultCode = append(methodResultCode, jen.Id(param.Name).Id(param.Type.String()))
	}



	return jen.Func().Params(
		jen.Id("s").Op("*").Id("logging")).Id(method.Name).Params(
			methodParamCode...,
		).Params(
			methodResultCode...,
		).Block(
			jen.Defer().Func().Params(
				jen.Id("begin").Qual("time", "Time")).
			Block(
				jen.Id("_").Op("=").Id("s").Dot("logger").Dot("Log").Call(
					jen.Id("s").Dot("traceId"),
					jen.Id("ctx").Dot("Value").Call(jen.Id("s").Dot("traceId")),
					jen.Lit("method"), jen.Lit(method.Name),
					&logParamSatement,
					jen.Lit("took"), jen.Qual("time", "Since").Call(jen.Id("begin")), jen.Lit("err"),
					jen.Id("err"))).Call(jen.Qual("time", "Now").Call(),
			),
			jen.Return().Id("s").Dot("next").Dot(method.Name).Call(
				nextMethodParamCode...,
			))
}

func genNewLogging(logPrefix string) jen.Code {
	return jen.Func().Id("NewLogging").Params(jen.Id("logger").Id("log").Dot("Logger"), jen.Id("traceId").Id("string")).Params(jen.Id("Middleware")).Block(
		jen.Id("logger").Op("=").Id("log").Dot("With").Call(
			jen.Id("logger"),
			jen.Lit(logPrefix),
			jen.Lit("logging"),
		), jen.Return().Func().Params(jen.Id("next").Id("Service")).Params(jen.Id("Service")).Block(
			jen.Return().Op("&").Id("logging").Values(jen.Id("logger").Op(":").Id("level").Dot("Info").Call(
				jen.Id("logger")),
				jen.Id("next").Op(":").Id("next"),
				jen.Id("traceId").Op(":").Id("traceId")),
		),
	)
}

// tracing

func genTracingStruct() jen.Code {
	return jen.Null().Type().Id("tracing").Struct(jen.Id("next").Id("Service"), jen.Id("tracer").Id("opentracing").Dot("Tracer"))
}
func genTracingFunc(tracingPrefix string,method ImplMethod) jen.Code {
	methodParamCode := make([]jen.Code, 0)
	methodResultCode := make([]jen.Code, 0)

	tracingParamCode := make([]jen.Code, 0)
	nextMethodParamCode := make([]jen.Code, 0)
	for _, param := range method.Params {
		methodParamCode = append(methodParamCode, jen.Id(param.Name).Id(param.Type.String()))
		nextMethodParamCode = append(nextMethodParamCode, jen.Id(param.Name))
	}

	for _, param := range method.ParamsExcludeCtx() {
		tracingParamCode = append(tracingParamCode, jen.Lit(param.Name), jen.Id(param.Name))
	}
	tracingParamSatement := jen.Statement(tracingParamCode)

	for _, param := range method.Results {
		methodResultCode = append(methodResultCode, jen.Id(param.Name).Id(param.Type.String()))
	}

	return jen.Func().Params(jen.Id("s").Op("*").Id("tracing")).Id(method.Name).Params(
		methodParamCode...,
	).Params(
		methodResultCode...,
	).Block(
		jen.List(jen.Id("span"), jen.Id("ctx")).Op(":=").Id("opentracing").Dot("StartSpanFromContextWithTracer").Call(
			jen.Id("ctx"),
			jen.Id("s").Dot("tracer"),
			jen.Lit(method.Name),
			jen.Id("opentracing").Dot("Tag").Values(jen.Id("Key").Op(":").Id("string").Call(
				jen.Id("ext").Dot("Component")),
				jen.Id("Value").Op(":").Lit(tracingPrefix),
			),
		),
		jen.Defer().Func().Params().Block(jen.Id("span").Dot("LogKV").Call(
			&tracingParamSatement,
			jen.Lit("err"), jen.Id("err")),
			jen.Id("span").Dot("SetTag").Call(jen.Id("string").Call(jen.Id("ext").Dot("Error")), jen.Id("err").Op("!=").Id("nil")), jen.Id("span").Dot("Finish").Call()).Call(),
		jen.Return().Id("s").Dot("next").Dot(method.Name).Call(
			nextMethodParamCode...,
		))
}
func genNewTracing() jen.Code {
	return jen.Func().Id("NewTracing").Params(jen.Id("otTracer").Id("opentracing").Dot("Tracer")).Params(jen.Id("Middleware")).Block(jen.Return().Func().Params(jen.Id("next").Id("Service")).Params(jen.Id("Service")).Block(jen.Return().Op("&").Id("tracing").Values(jen.Id("next").Op(":").Id("next"), jen.Id("tracer").Op(":").Id("otTracer"))))
}
