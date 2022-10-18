package generator

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"reflect"
	"strings"
	"unicode"
)

const (
	RequestParamTagName string = "param"
	QueryTag            string = "query"
	HeaderTag           string = "header"
	PathTag             string = "path"
	BodyTag             string = "body"
	CtxTag              string = "ctx"

	DocKitHttpParamMark string = "@kit-http-param"
)

type KitRequest struct {
	pkg *packages.Package

	ServiceName   string
	RequestTypeOf *types.Struct
	RequestName   string
	RequestIsBody bool
	RequestIsNil  bool

	NamedMap map[string]string
	Query         map[string]RequestParam
	Path          map[string]RequestParam
	Body          map[string]RequestParam
	Header        map[string]RequestParam
	Ctx           map[string]RequestParam
	Empty         map[string]RequestParam
}

func NewKitRequest(pkg *packages.Package, serviceName, requestName string, requestIsBody bool) *KitRequest {
	return &KitRequest{
		pkg:           pkg,
		ServiceName:   serviceName,
		RequestName:   requestName,
		RequestIsBody: requestIsBody,
		NamedMap: 	make(map[string]string),
		Query:         make(map[string]RequestParam),
		Path:          make(map[string]RequestParam),
		Body:          make(map[string]RequestParam),
		Header:        make(map[string]RequestParam),
		Ctx:           make(map[string]RequestParam),
		Empty:         make(map[string]RequestParam),
	}
}

type RequestParam struct {
	ParamDoc  *ast.CommentGroup

	ParamPath string

	FieldName string

	ParamName string
	// path, query, header, body, empty
	ParamSource string
	// basic, map, slice,ptr
	ParamType string
	// time [time.Time]
	ParamTypeName string
	// []string map[string]string
	RawParamType string
	// int,string,bool,float
	BasicType string


	HasPtr bool
}

func (r RequestParam) Annotations() string {
	if r.ParamDoc == nil {
		return `" "`
	}
	for _, v := range r.ParamDoc.List {
		docFormat := DocFormat(v.Text)
		if strings.HasPrefix(docFormat, "// " + r.FieldName) {
			return fmt.Sprintf(`"%s"`,strings.TrimPrefix(docFormat, "// " + r.FieldName))
		}
	}
	return fmt.Sprintf(`"%s"`,strings.TrimPrefix(r.ParamDoc.List[0].Text, "// "))
}

func (r RequestParam) ToVal() jen.Code {
	return jen.Var().Id(r.ParamName).Id(r.ParamTypeName)
	//switch r.ParamType {
	//case "basic":
	//	return jen.Var().Id(r.ParamName).Id(r.BasicType)
	//case "map":
	//	return jen.Var().Id(r.ParamName).Map(jen.String()).Id(r.BasicType).Values()
	//case "slice":
	//	return jen.Var().Id(r.ParamName).Index().Id(r.BasicType)
	//case "struct":
	//	return jen.Var().Id(r.ParamName).Id(r.BasicType)
	//
	//}
	//return nil
}

func (k *KitRequest) ParamPath(paramName string) (res string) {
	defer func() {
		if res != "" {
			res = "." + res
		}
	}()

	if strings.ToUpper(paramName) == strings.ToUpper(k.RequestName) {
		return ""
	}
	param, ok := k.Query[paramName]
	if ok {
		return param.ParamPath
	}

	param, ok = k.Path[paramName]
	if ok {
		return param.ParamPath
	}

	param, ok = k.Header[paramName]
	if ok {
		return param.ParamPath
	}

	param, ok = k.Body[paramName]
	if ok {
		return param.ParamPath
	}

	param, ok = k.Ctx[paramName]
	if ok {
		return param.ParamPath
	}

	param, ok = k.Empty[paramName]
	if ok {
		return param.ParamPath
	}

	panic("param not found: " + paramName)
}

func (k *KitRequest) SetParam(param RequestParam) {
	switch param.ParamSource {
	case QueryTag:
		k.Query[param.ParamName] = param
	case PathTag:
		k.Path[param.ParamName] = param
	case HeaderTag:
		k.Header[param.ParamName] = param
	case BodyTag:
		k.Body[param.ParamName] = param
	case CtxTag:
		k.Ctx[param.ParamName] = param
	case "":
		k.Empty[param.ParamName] = param

	default:
		panic("param source error: " + param.ParamSource + "," + param.ParamName)
	}
}

func (k *KitRequest) ParseParamTag(fieldName, tag string) (paramSource string, paramName string) {

	split := strings.Split(tag, ",")
	if len(split) == 1 {
		return split[0], downFirst(fieldName)
	}

	if len(split) == 2 {
		return split[0], split[1]
	}

	return "", ""

}

func (k *KitRequest) DecodeRequest() (s string) {
	listCode := make([]jen.Code, 0, 0)
	// req := Request{}
	listCode = append(listCode, jen.Id("req").Op(":=").Id(k.RequestName).Block())
	listCode = append(listCode, k.DefineVal()...)
	listCode = append(listCode, k.BindPathParam()...)
	listCode = append(listCode, k.BindQueryParam()...)
	listCode = append(listCode, k.BindHeaderParam()...)
	listCode = append(listCode, k.BindBodyParam()...)
	listCode = append(listCode, k.BindCtxParam()...)
	listCode = append(listCode, k.BindRequest()...)
	listCode = append(listCode, k.Validate()...)
	listCode = append(listCode, jen.Return(jen.Id("req"), jen.Id("err")))
	var LineListCode []jen.Code
	for _, v := range listCode {
		LineListCode = append(LineListCode, jen.Line(),v)
	}

	fn := jen.Func().Id("decode"+upFirst(k.ServiceName)+"Request").Params(
		jen.Id("ctx").Id("context.Context"),
		jen.Id("r").Id("*http").Dot("Request"),
	).Call(
		jen.Id("res").Interface(),
		jen.Id("err").Id("error"),
	).Block(
		LineListCode...,
	)
	return fn.GoString()
}

func (k *KitRequest) DefineVal() []jen.Code {
	listCode := make([]jen.Code, 0, 0)
	for _, v := range k.Query {
		listCode = append(listCode, v.ToVal())
	}
	for _, v := range k.Path {
		listCode = append(listCode, v.ToVal())
	}
	for _, v := range k.Header {
		listCode = append(listCode, v.ToVal())
	}

	for _, v := range k.Ctx {
		listCode = append(listCode, v.ToVal())
	}
	return listCode
}

func (k *KitRequest) Validate() []jen.Code {
	list := make([]jen.Code, 0, 0)
	list = append(
		list,
		jen.List(jen.Id("validReq"), jen.Id("err")).Op(":=").Id("valid").Dot("ValidateStruct").Call(jen.Id("req")),
		jen.If(jen.Err().Op("!=").Nil()).Block(
			jen.Err().Op("=").Id("errors.Wrap").Call(jen.Id("err"), jen.Lit("valid.ValidateStruct")),
			jen.Return(),
		),
		jen.If(jen.Id("!validReq")).Block(
			jen.Err().Op("=").Id("errors.Wrap").Call(jen.Id("err"), jen.Lit("valid false")),
			jen.Return(),
		),
	)
	return list
}

func (k *KitRequest) BindBodyParam() []jen.Code {
	listCode := make([]jen.Code, 0, 0)
	if k.RequestIsNil {
		return listCode
	}
	returnCode := jen.If(jen.Err().Op("!=").Nil()).Block(
		jen.Err().Op("=").Id("errors.Wrap").Call(jen.Id("err"), jen.Lit("json.Decode")),
		jen.Return(),
	)
	if k.RequestIsBody {
		// err = json.NewDecoder(r.Body).Decode(&req)
		decode := jen.Id("err").Op("=").Id("json.NewDecoder").Call(jen.Id("r.Body")).Dot("Decode").Parens(jen.Id("&req"))
		listCode = append(listCode, decode, returnCode)

		return listCode
	}

	if len(k.Body) == 0 {
		return listCode
	}

	if len(k.Body) != 1 {
		panic("body param count error " + fmt.Sprint(len(k.Body)))
	}

	for _, v := range k.Body {
		decode := jen.Id("err").Op("=").Id("json.NewDecoder").Call(jen.Id("r.Body")).Dot("Decode").Parens(jen.Id("&req." + v.ParamPath))
		listCode = append(listCode, decode, returnCode)
	}

	return listCode
}

func (k *KitRequest) BindHeaderParam() []jen.Code {
	list := make([]jen.Code, 0, 0)

	for _, v := range k.Header {
		//r.Header.Get("project")
		varBind := jen.Id("r.Header.Get").Call(jen.Lit(v.ParamName))
		if v.BasicType != "string" {
			// cast.ToInt(vars["id"])
			varBind = jen.Id("cast").Dot("To" + upFirst(v.BasicType) + "E").Call(varBind)
			// id, err := cast.ToIntE(vars["id"])
			varBind = jen.List(jen.Id(v.ParamName), jen.Err()).Op("=").Add(varBind)
			// if err != nil {
			// 	return err
			// }
			returnCode := jen.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(),
			)
			list = append(list, varBind, returnCode)
			continue
		}
		// id = vars["id"]
		val := jen.Id(v.ParamName).Op("=").Add(varBind)
		list = append(list, val)
	}

	return list

}

func (k *KitRequest) BindQueryParam() []jen.Code {
	list := make([]jen.Code, 0, 0)

	if len(k.Query) == 0 {
		return list
	}

	for _, v := range k.Query {
		//r.URL.Query().Get("project")
		varBind := jen.Id("r.URL.Query().Get").Call(jen.Lit(v.ParamName))
		if !(v.ParamType == "basic" && v.BasicType == "string") {
			castCode, err := CastMap(v.ParamName, v.ParamType, v.ParamTypeName,  varBind)
			if err != nil {
				panic(err)
			}
			list = append(list, castCode...)
			continue
		}
		// id = vars["id"]
		val := jen.Id(v.ParamName).Op("=").Add(varBind)
		list = append(list, val)
	}

	return list
}

func (k *KitRequest) BindPathParam() []jen.Code {
	list := make([]jen.Code, 0, 0)

	if len(k.Path) == 0 {
		return list
	}

	// vars := mux.Vars(r)
	vars := jen.Id("vars").Op(":=").Qual("github.com/gorilla/mux", "Vars").Call(jen.Id("r"))
	list = append(list, vars)
	for _, v := range k.Path {
		// vars["id"]
		varBind := jen.Id("vars").Index(jen.Lit(v.ParamName))
		if v.BasicType != "string" {
			// cast.ToInt(vars["id"])
			varBind = jen.Id("cast").Dot("To" + upFirst(v.BasicType) + "E").Call(varBind)
			// id, err := cast.ToIntE(vars["id"])
			varBind = jen.List(jen.Id(v.ParamName), jen.Err()).Op("=").Add(varBind)
			// if err != nil {
			// 	return err
			// }
			returnCode := jen.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(),
			)
			list = append(list, varBind, returnCode)
			continue
		}
		// id = vars["id"]
		val := jen.Id(v.ParamName).Op("=").Add(varBind)
		list = append(list, val)
	}

	return list
}

func (k *KitRequest) BindCtxParam() []jen.Code {
	list := make([]jen.Code, 0, 0)
	for _, v := range k.Ctx {
		var ctxKey string

		if v.ParamDoc == nil {
			panic("ctx param doc is nil")
		}
		for _, d := range v.ParamDoc.List {
			fields := strings.Fields(d.Text)
			if fields[1] == DocKitHttpParamMark {
				if len(fields) < 3 {
					panic("ctx param doc error: " + d.Text)
				}

				if fields[2] == "ctx" {
					ctxKey = fields[3]
				}

			}
		}
		if ctxKey == "" {
			panic("not find ctx param doc error: " + v.ParamDoc.Text())
		}
		ctxVal := jen.Var().Id(v.ParamName + "OK").Bool()
		varBind := jen.List(jen.Id(v.ParamName), jen.Id(v.ParamName+"OK")).Op("=").Id("ctx.Value").Call(jen.Id(ctxKey)).Assert(jen.Id(v.RawParamType))
		ifBind := jen.If(jen.Id(v.ParamName+"OK")).Op("==").False().Block(
			jen.Err().Op("=").Id("errors.New").Call(jen.Lit("ctx param "+v.ParamName+" is not found")),
			jen.Return(),
		)
		list = append(list, ctxVal, varBind, ifBind)
	}
	return list
}

func (k *KitRequest) BindRequest() []jen.Code {
	list := make([]jen.Code, 0, 0)
	for _, v := range k.Path {
		reqBindVal := jen.Id("req").Dot(v.ParamPath).Op("=").Id(v.ParamName)
		list = append(list, reqBindVal)
	}

	for _, v := range k.Query {
		reqBindVal := jen.Id("req").Dot(v.ParamPath).Op("=").Id(v.ParamName)
		list = append(list, reqBindVal)
	}

	for _, v := range k.Header {
		reqBindVal := jen.Id("req").Dot(v.ParamPath).Op("=").Id(v.ParamName)
		list = append(list, reqBindVal)
	}

	for _, v := range k.Ctx {
		reqBindVal := jen.Id("req").Dot(v.ParamPath).Op("=").Id(v.ParamName)
		list = append(list, reqBindVal)
	}
	return list
}

func (k *KitRequest) ParseRequest() {
	var hasFindRequest bool
	for _, s := range k.pkg.Syntax {
		ast.Inspect(s, func(node ast.Node) bool {
			switch nodeT := node.(type) {
			case *ast.GenDecl:
				for _, spec := range nodeT.Specs {
					if specT, ok := spec.(*ast.TypeSpec); ok {
						if specT.Name.Name == k.RequestName {
							hasFindRequest = true
							doc := nodeT.Doc
							k.Doc(doc)
							t := k.pkg.TypesInfo.TypeOf(specT.Type).(*types.Struct)
							k.RequestTypeOf = t
							k.RequestType([]string{}, k.RequestName, t, "", doc)
							k.CheckRequestIsNil()
							
							return false
						}
					}
				}

			}
			return true
		})
	}
	if !hasFindRequest {
		panic("not find request" + k.RequestName)
	}
}

func (k *KitRequest) Doc(doc *ast.CommentGroup) {
}

func (k *KitRequest) ParseFieldComment(pos token.Pos) (s *ast.CommentGroup) {
	fieldFileName := k.pkg.Fset.Position(pos).Filename
	fieldLine := k.pkg.Fset.Position(pos).Line
	var fieldComment *ast.CommentGroup
	for _, syntax := range k.pkg.Syntax {
		fileName := k.pkg.Fset.Position(syntax.Pos()).Filename
		if fieldFileName == fileName {
			for _, c := range syntax.Comments {
				if k.pkg.Fset.Position(c.End()).Line+1 == fieldLine {
					fieldComment = c
				}
			}
			break
		}
	}
	return fieldComment

	//if fieldComment == nil {
	//	return ""
	//}
	//
	//for _, c := range fieldComment {
	//	commentField := strings.Fields(c.Text)
	//	if len(commentField) < 3 {
	//		panic("comment error: " + c.Text)
	//	}
	//	fmt.Println("commentField", commentField)
	//	if commentField[0] == "@kit-request" && commentField[1] == "ctx" {
	//		return commentField[2]
	//	}
	//}
	//return ""
}

func (k *KitRequest) CheckRequestIsNil() {
	if k.RequestIsBody {
		if k.RequestTypeOf.NumFields() == 0 {
			k.RequestIsNil = true
		}
	}
}

func (k *KitRequest) RequestType(prefix []string, requestName string, requestType types.Type, requestParamTagTypeTag string, doc *ast.CommentGroup) {
	rawParamType := requestType.String()
	paramSource, paramName := k.ParseParamTag(requestName, requestParamTagTypeTag)

	switch rt := requestType.(type) {
	case *types.Named:
		//fmt.Println("paramName", paramName)
		//fmt.Println("obj.name",rt.Obj().Name())
		//fmt.Println("obj.pkg",rt.Obj().Pkg().Path())
		//fmt.Println("obj.pos",rt.Obj().Pos())
		//fmt.Println("obj.type",rt.Obj().Type())
		//fmt.Println("obj.type.string",rt.Obj().Type().String())
		//fmt.Println("obj.id",rt.Obj().Id())
		//fmt.Println("local.pkg.pkgPath", k.pkg.PkgPath)
		split := strings.Split(strings.TrimPrefix(rt.Obj().Type().String(), k.pkg.PkgPath+"."),"/")
		named := split[len(split)-1]

		k.NamedMap[paramName] = named
		//k.SetParam(RequestParam{
		//	ParamDoc:     doc,
		//	ParamPath:    strings.Join(prefix, "."),
		//	FieldName:    requestName,
		//	ParamName:    paramName,
		//	ParamSource:  paramSource,
		//	ParamType:    "named",
		//	RawParamType: rawParamType,
		//	BasicType:    rt.Underlying().String(),
		//	HasPtr:       false,
		//})
		k.RequestType(prefix, requestName, rt.Underlying(), requestParamTagTypeTag, doc)
	case *types.Struct:
		k.SetParam(RequestParam{
			FieldName:   requestName,
			ParamDoc:    doc,
			ParamPath:   strings.Join(prefix, "."),
			ParamName:   paramName,
			ParamSource: paramSource,
			ParamType:   "struct",
			ParamTypeName: k.NamedMap[paramName],
			RawParamType: rawParamType,
			BasicType:   k.NamedMap[paramName],
			HasPtr:      false,
		})
		for i := 0; i < rt.NumFields(); i++ {
			field := rt.Field(i)
			fieldName := field.Name()
			fieldType := field.Type()
			tag, _ := reflect.StructTag(rt.Tag(i)).Lookup(RequestParamTagName)
			k.RequestType(append(prefix, fieldName), fieldName, fieldType, tag, k.ParseFieldComment(field.Pos()))
		}
	case *types.Pointer:
		k.RequestType(prefix, requestName, rt.Elem().Underlying(), requestParamTagTypeTag, doc)
	case *types.Slice:
		var paramTypeName string
		var ok bool
		if paramTypeName, ok = k.NamedMap[paramName]; !ok {
			split := strings.Split(strings.TrimPrefix(rt.Elem().String(), k.pkg.PkgPath+"."),"/")
			paramTypeName = "[]" + split[len(split)-1]
		}
		k.SetParam(RequestParam{
			FieldName:   requestName,
			ParamDoc:    doc,
			ParamPath:   strings.Join(prefix, "."),
			ParamName:   paramName,
			ParamSource: paramSource,
			ParamType:   "slice",
			ParamTypeName: paramTypeName,
			RawParamType: rawParamType,
			BasicType:   rt.Elem().Underlying().String(),
			HasPtr:      false,
		})
	case *types.Map:
		var paramTypeName string
		var ok bool
		if paramTypeName, ok = k.NamedMap[paramName]; !ok {
			split := strings.Split(strings.TrimPrefix(rt.Elem().String(), k.pkg.PkgPath+"."),"/")
			paramTypeName = split[len(split)-1]
		}
		k.SetParam(RequestParam{
			FieldName:   requestName,
			ParamDoc:    doc,
			ParamPath:   strings.Join(prefix, "."),
			ParamName:   paramName,
			ParamSource: paramSource,
			ParamType:   "map",
			ParamTypeName: paramTypeName,
			RawParamType: rawParamType,
			BasicType:   rt.Elem().Underlying().String(),
			HasPtr:      false,
		})
	case *types.Basic:
		var paramTypeName string
		var ok bool
		if paramTypeName, ok = k.NamedMap[paramName]; !ok {
			paramTypeName = rt.Name()
		}
		k.SetParam(RequestParam{
			FieldName:   requestName,
			ParamDoc:    doc,
			ParamPath:   strings.Join(prefix, "."),
			ParamName:   paramName,
			ParamSource: paramSource,
			ParamType:   "basic",
			ParamTypeName: paramTypeName,
			RawParamType: rawParamType,
			BasicType:   rt.Name(),
			HasPtr:      false,
		})
	default:
		return
	}

	return
}

func downFirst(s string) string {
	for _, v := range s {
		return string(unicode.ToLower(v)) + s[len(string(v)):]
	}
	return ""
}

func upFirst(s string) string {
	for _, v := range s {
		return string(unicode.ToUpper(v)) + s[len(string(v)):]
	}
	return ""
}

func CastMap(paramName, t, paramTypeName string,  code jen.Code) (res []jen.Code, err error) {
	if t == "slice" && paramTypeName == "string" {
		res = append(res, jen.Id(paramName).Op("=").Id("strings.Split").Call(code, jen.Lit(",")))
		return
	}
	var m = map[string]string{
		"basic.int":     "cast.ToIntE",
		"basic.int8":    "cast.ToInt8E",
		"basic.int16":   "cast.ToInt16E",
		"basic.int32":   "cast.ToInt32E",
		"basic.int64":   "cast.ToInt64E",
		"basic.uint":    "cast.ToUintE",
		"basic.uint8":   "cast.ToUint8E",
		"basic.uint16":  "cast.ToUint16E",
		"basic.uint32":  "cast.ToUint32E",
		"basic.uint64":  "cast.ToUint64E",
		"basic.float32": "cast.ToFloat32E",
		"basic.float64": "cast.ToFloat64E",
		"basic.string":  "cast.ToStringE",
		"basic.bool":    "cast.ToBoolE",

		"slice.int":  "cast.ToIntSliceE",
		"slice.bool": "cast.ToBoolSliceE",

		"map.int":   "cast.ToStringMapIntE",
		"map.int64": "cast.ToStringMapInt64E",
		"map.bool":  "cast.ToStringMapBoolE",

		"struct.time.Time": "cast.ToTimeE",
		"basic.time.Duration": "cast.ToDurationE",
	}
	var ok bool
	mKey := t+"."+paramTypeName
	fnStr, ok := m[mKey]
	if !ok {
		err = fmt.Errorf("CastMap not found %s %s", t, paramTypeName)
		return
	}

	paramStr := paramName +"Str"
	varParamStr := jen.Id(paramStr).Op(":=").Add(code)
	paramStrCode := jen.Id(paramStr)
	if t == "slice" {
		paramStrCode = jen.Id("strings.Split").Call(paramStrCode, jen.Lit(","))
	}

	switch mKey {
	case "struct.time.Time":
		paramStrCode = jen.Id("cast.ToInt64").Call(paramStrCode)
	case "basic.time.Duration":
		paramStrCode = jen.Id("cast.ToInt64").Call(paramStrCode)
	}

	ifParamStr := jen.If(jen.Id(paramStr).Op("!=").Lit("")).Block(
		jen.List(jen.Id(paramName), jen.Err()).Op("=").Id(fnStr).Call(paramStrCode),
		// if err != nil {
		// 	return err
		// }
		jen.If(jen.Err().Op("!=").Nil()).Block(
			jen.Return(),
		),
	)

	res = append(res, varParamStr,ifParamStr)

	return
}
