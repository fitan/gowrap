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
	EmptyTag            string = "empty"
)

type KitRequest struct {
	pkg *packages.Package

	ServiceName   string
	RequestName   string
	RequestIsBody bool
	Query         map[string]RequestParam
	Path          map[string]RequestParam
	Body          map[string]RequestParam
	Header        map[string]RequestParam
	Ctx           map[string]RequestParam
	Empty         map[string]RequestParam
}

func NewKitRequest(pkg *packages.Package, serviceName, requestName string, requestIsBody bool) *KitRequest {
	return &KitRequest{
		pkg:    pkg,
		ServiceName: serviceName,
		RequestName: requestName,
		RequestIsBody: requestIsBody,
		Query:  make(map[string]RequestParam),
		Path:   make(map[string]RequestParam),
		Body:   make(map[string]RequestParam),
		Header: make(map[string]RequestParam),
		Ctx:    make(map[string]RequestParam),
		Empty:  make(map[string]RequestParam),
	}
}

type RequestParam struct {
	ParamPath string

	ParamName string
	// path, query, header, body, empty
	ParamSource string
	// basic, map, slice,ptr
	ParamType string
	// int,string,bool,float
	BasicType string

	HasPtr bool
}

func (r RequestParam) ToVal() jen.Code {
	switch r.ParamType {
	case "basic":
		return jen.Var().Id(r.ParamName).Id(r.BasicType)
	case "map":
		return jen.Var().Id(r.ParamName).Map(jen.String()).Id(r.BasicType).Values()
	case "slice":
		return jen.Var().Id(r.ParamName).Index().Id(r.BasicType)
	}
	return nil
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
	case EmptyTag:
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
	listCode = append(listCode, k.Validate()...)
	listCode = append(listCode, jen.Return())

	fn := jen.Func().Id("decode"+upFirst(k.ServiceName)+"Request").Params(
		jen.Id("ctx").Id("context.Context"),
		jen.Id("r").Id("*http").Dot("Request"),
	).Call(
		jen.Id("res").Interface(),
		jen.Id("err").Id("error"),
	).Block(
		listCode...,
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
	return listCode
}

func (k *KitRequest) Validate() []jen.Code {
	list := make([]jen.Code, 0, 0)
	list = append(
		list,
		jen.List(jen.Id("validReq"), jen.Id("err")).Op(":=").Id("valid").Dot("Validate").Call(jen.Id("req")),
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
	returnCode := jen.If(jen.Err().Op("!=").Nil()).Block(
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
	if len(k.Query) == 0 {
		return list
	}

	for _, v := range k.Header {
		//r.Header.Get("project")
		varBind := jen.Id("r.Header.Get").Call(jen.Lit(v.ParamName))
		// req.ID = id
		reqBindVal := jen.Id("req").Dot(v.ParamPath).Op("=").Id(v.ParamName)
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
			list = append(list, varBind, returnCode, reqBindVal)
			continue
		}
		// id = vars["id"]
		val := jen.Id(v.ParamName).Op("=").Add(varBind)
		list = append(list, val, reqBindVal)
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
		// req.ID = id
		reqBindVal := jen.Id("req").Dot(v.ParamPath).Op("=").Id(v.ParamName)
		fmt.Println(v.ParamName, v.ParamType, v.BasicType)
		if !(v.ParamType == "basic" && v.BasicType == "string") {
			castCode, err := CastMap(v.ParamName,v.ParamType, v.BasicType, varBind)
			if err != nil {
				panic(err)
			}
			list = append(list, varBind)
			list = append(list, castCode...)
			list = append(list, reqBindVal)
			continue
		}
		// id = vars["id"]
		val := jen.Id(v.ParamName).Op("=").Add(varBind)
		list = append(list, val, reqBindVal)
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
		// req.ID = id
		reqBindVal := jen.Id("req").Dot(v.ParamPath).Op("=").Id(v.ParamName)
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
			list = append(list, varBind, returnCode, reqBindVal)
			continue
		}
		// id = vars["id"]
		val := jen.Id(v.ParamName).Op("=").Add(varBind)
		list = append(list, val, reqBindVal)
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
						fmt.Println("specT.Name.Name", specT.Name.Name, "k.RequestName", k.RequestName)
						if specT.Name.Name == k.RequestName {
							hasFindRequest =  true
							doc := nodeT.Doc
							k.Doc(doc)
							t := k.pkg.TypesInfo.TypeOf(specT.Type)
							k.RequestType([]string{}, k.RequestName, t, "")
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
	k.RequestIsBody = strings.Contains(doc.Text(), "@kit-request body")
}

func (k *KitRequest) ParseFieldComment(pos token.Pos) (s string) {
	fmt.Println("parse field comment")
	fieldFileName := k.pkg.Fset.Position(pos).Filename
	fieldLine := k.pkg.Fset.Position(pos).Line
	var fieldComment []*ast.Comment
	for _, syntax := range k.pkg.Syntax {
		fileName := k.pkg.Fset.Position(syntax.Pos()).Filename
		if fieldFileName == fileName {
			fmt.Println("find line: ", fieldFileName)
			for _, c := range syntax.Comments {
				fmt.Println("comment line: ", k.pkg.Fset.Position(c.End()).Line, "field line: ", fieldLine)
				if k.pkg.Fset.Position(c.End()).Line+1 == fieldLine {
					fieldComment = c.List
				}
			}
			break
		}
	}

	if fieldComment == nil {
		return ""
	}

	for _, c := range fieldComment {
		commentField := strings.Fields(c.Text)
		if len(commentField) < 3 {
			panic("comment error: " + c.Text)
		}
		fmt.Println("commentField", commentField)
		if commentField[0] == "@kit-request" && commentField[1] == "ctx" {
			return commentField[2]
		}
	}
	return ""
}

func (k *KitRequest) RequestType(prefix []string, requestName string, requestType types.Type, requestParamTagTypeTag string) {
	paramSource, paramName := k.ParseParamTag(requestName, requestParamTagTypeTag)
	if paramSource == "" {
		paramSource = "empty"
	}

	switch rt := requestType.(type) {
	case *types.Named:
		k.RequestType(prefix, requestName, requestType.(*types.Named).Underlying(), requestParamTagTypeTag)
	case *types.Struct:
		k.SetParam(RequestParam{
			ParamPath:   strings.Join(prefix, "."),
			ParamName:   paramName,
			ParamSource: paramSource,
			ParamType:   "struct",
			BasicType:   "",
			HasPtr:      false,
		})
		for i := 0; i < rt.NumFields(); i++ {
			field := rt.Field(i)
			fieldName := field.Name()
			fieldType := field.Type()
			k.ParseFieldComment(field.Pos())
			tag, _ := reflect.StructTag(rt.Tag(i)).Lookup(RequestParamTagName)
			k.RequestType(append(prefix, fieldName), fieldName, fieldType, tag)
		}
	case *types.Pointer:
		k.RequestType(prefix, requestName, rt.Elem().Underlying(), requestParamTagTypeTag)
	case *types.Slice:
		k.SetParam(RequestParam{
			ParamPath:   strings.Join(prefix, "."),
			ParamName:   paramName,
			ParamSource: paramSource,
			ParamType:   "slice",
			BasicType:   rt.Elem().Underlying().String(),
			HasPtr:      false,
		})
	case *types.Map:
		k.SetParam(RequestParam{
			ParamPath:   strings.Join(prefix, "."),
			ParamName:   paramName,
			ParamSource: paramSource,
			ParamType:   "map",
			BasicType:   rt.Elem().Underlying().String(),
			HasPtr:      false,
		})
	case *types.Basic:
		k.SetParam(RequestParam{
			ParamPath:   strings.Join(prefix, "."),
			ParamName:   paramName,
			ParamSource: paramSource,
			ParamType:   "basic",
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

func CastMap(paramName ,t, basicT string, code jen.Code) (res []jen.Code, err error) {
	if t == "slice" && basicT == "string" {
		res = append(res,jen.Id(paramName).Op("=").Id("strings.Split").Call(code))
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
	}
	var ok bool
	fnStr, ok := m[t+"."+basicT]
	if !ok {
		err = fmt.Errorf("CastMap not found %s %s", t, basicT)
		return
	}

	if t == "slice" {
		code = jen.Id("strings.Split").Call(code)
	}


	code = jen.List(jen.Id(paramName), jen.Err()).Op("=").Id(fnStr).Call(code)
	// if err != nil {
	// 	return err
	// }
	returnCode := jen.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(),
	)

	res = append(res, code, returnCode)


	return
}
