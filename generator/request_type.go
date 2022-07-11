package generator

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"go/types"
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
)

type KitRequest struct {
	Query  map[string]RequestParam
	Path   map[string]RequestParam
	Body   map[string]RequestParam
	Header map[string]RequestParam
}

func NewKitRequest() *KitRequest {
	return &KitRequest{
		Query:  make(map[string]RequestParam),
		Path:   make(map[string]RequestParam),
		Body:   make(map[string]RequestParam),
		Header: make(map[string]RequestParam),
	}
}

type RequestParam struct {
	ParamPath string

	ParamName string
	// path, query, header, body
	ParamSource string
	// basic, map, slice,ptr
	ParamType string
	// int,string,bool,float
	BasicType string

	HasPtr bool
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

func (k *KitRequest) BindQueryParam() (s string) {
	if len(k.Query) == 0 {
		return
	}

	list := make([]jen.Code, 0, 0)

	for _, v := range k.Query {
		//r.URL.Query().Get("project")
		varBind := jen.Id("r.URL.Query().Get").Call(jen.Lit(v.ParamName))
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

	return jen.Block(list...).GoString()
}

func (k *KitRequest) BindPathParam() (s string) {
	if len(k.Path) == 0 {
		return
	}

	list := make([]jen.Code, 0, 0)

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

	return jen.Block(list...).GoString()
}

func (k *KitRequest) RequestType(prefix []string, requestName string, requestType types.Type, requestParamTagTypeTag string) {
	paramSource, paramName := k.ParseParamTag(requestName, requestParamTagTypeTag)

	switch rt := requestType.(type) {
	case *types.Named:
		k.RequestType(prefix, requestName, requestType.(*types.Named).Underlying(), requestParamTagTypeTag)
	case *types.Struct:
		if requestParamTagTypeTag != "" {

			k.SetParam(RequestParam{
				ParamPath:   strings.Join(prefix, "."),
				ParamName:   paramName,
				ParamSource: paramSource,
				ParamType:   "body",
				BasicType:   "",
				HasPtr:      false,
			})
			return
		}
		for i := 0; i < rt.NumFields(); i++ {
			field := rt.Field(i)
			fieldName := field.Name()
			fieldType := field.Type()
			tag, _ := reflect.StructTag(rt.Tag(i)).Lookup(RequestParamTagName)
			k.RequestType(append(prefix, fieldName), fieldName, fieldType, tag)
		}
	case *types.Pointer:
		k.RequestType(prefix, requestName, rt.Elem().Underlying(), requestParamTagTypeTag)
	case *types.Slice:
		if requestParamTagTypeTag == "" {
			return
		}
		k.SetParam(RequestParam{
			ParamPath:   strings.Join(prefix, "."),
			ParamName:   paramName,
			ParamSource: paramSource,
			ParamType:   "slice",
			BasicType:   "",
			HasPtr:      false,
		})
	case *types.Map:
		if requestParamTagTypeTag == "" {
			return
		}
		k.SetParam(RequestParam{
			ParamPath:   strings.Join(prefix, "."),
			ParamName:   paramName,
			ParamSource: paramSource,
			ParamType:   "map",
			BasicType:   "",
			HasPtr:      false,
		})
	case *types.Basic:
		if requestParamTagTypeTag == "" {
			return
		}
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

func CastMap(t, basicT string) (fnStr string, err error) {
	var m = map[string]string{
		"basic.int":     "ToIntE",
		"basic.int8":    "ToInt8E",
		"basic.int16":   "ToInt16E",
		"basic.int32":   "ToInt32E",
		"basic.int64":   "ToInt64E",
		"basic.uint":    "ToUintE",
		"basic.uint8":   "ToUint8E",
		"basic.uint16":  "ToUint16E",
		"basic.uint32":  "ToUint32E",
		"basic.uint64":  "ToUint64E",
		"basic.float32": "ToFloat32E",
		"basic.float64": "ToFloat64E",
		"basic.string":  "ToStringE",
		"basic.bool":    "ToBoolE",

		"slice.int":  "ToIntSliceE",
		"slice.bool": "ToBoolSliceE",

		"map.int":   "ToStringMapIntE",
		"map.int64": "ToStringMapInt64E",
		"map.bool":  "ToStringMapBoolE",
	}
	var ok bool
	fnStr, ok = m[t+"."+basicT]
	if !ok {
		err = fmt.Errorf("CastMap not found %s %s", t, basicT)
	}
	return
}
