package generator

import (
	"go/types"
	"reflect"
	"strings"
)

const (
	RequestParamTagName string =  "param"
	QueryTag string = "query"
	HeaderTag string = "header"
	PathTag string = "path"
	BodyTag string = "body"
)



type KitRequest struct {
	Query map[string]RequestParam
	Path map[string]RequestParam
	Body map[string]RequestParam
	Header map[string]RequestParam
}

func NewKitRequest() *KitRequest {
	return &KitRequest{
		Query: make(map[string]RequestParam),
		Path: make(map[string]RequestParam),
		Body: make(map[string]RequestParam),
		Header: make(map[string]RequestParam),
	}
}

type RequestParam struct {
	ParamPath string
	ParamType string
	HasPtr bool
}

func (k *KitRequest) ParseRequestType(prefix []string, request *types.Struct) {
	for i := 0; i < request.NumFields(); i++ {
		if !request.Field(i).Exported() || request.Field(i).Anonymous() {
			continue
		}
		name := request.Field(i).Name()
		fieldType := request.Field(i).Type()
		tag,ok := reflect.StructTag(request.Tag(i)).Lookup(RequestParamTagName)
		if ok {
			switch tag {
			case QueryTag:
				switch ft := fieldType.(type) {
				//case *types.Map:
				//	paramType = "map"
				//case *types.Slice:
				//	paramType = "slice"
				case *types.Basic:
					k.Query[name] = RequestParam{
						ParamPath: strings.Join(append(prefix,name), "."),
						ParamType: ft.Name(),
						HasPtr:    false,
					}
				default:
					panic("query field type must be map or slice or basic " + name)
				}
			case PathTag:
				basic, ok := fieldType.(*types.Basic)
				if !ok {
					panic("path field must be basic type " + name)
				}

				k.Path[name] = RequestParam{
					ParamPath: strings.Join(append(prefix,name), ".") ,
					ParamType:  basic.Name(),
					HasPtr: false,
				}
			case HeaderTag:
				basic, ok := fieldType.(*types.Basic)
				if !ok {
					panic("header field must be basic type " + name)
				}

				k.Path[name] = RequestParam{
					ParamPath: strings.Join(append(prefix, name), "."),
					ParamType: basic.Name(),
					HasPtr:    false,
				}
			case BodyTag:
				structType, ok := fieldType.(*types.Struct)
				if !ok {
					panic("body field must be struct type " + name)
				}
				k.Body[name] = RequestParam{
					ParamPath: ,
					ParamType: "",
					HasPtr:    false,
				}


			default:
				panic("unknown tag: " + tag)
			}
		}

		switch ft := fieldType.(type) {
		case *types.Interface:
			continue
		case *types.Chan:
			continue
		case *types.Struct:
		case *types.Pointer:
		case *types.Slice:
		case *types.Map:
		case *types.Basic:
		case *types.Named:
		case *types.Array:
		}
	}

}

func (k *KitRequest)GenerateRequestType(prefix *[]string,requestType types.Type) {
	switch rt := requestType.(type) {
	case *types.Named:
		k.GenerateRequestType(prefix,requestType.(*types.Named).Underlying())
	case *types.Struct:
		for i := 0; i < rt.NumFields(); i++ {
			field := rt.Field(i)
			tag,ok := reflect.StructTag(rt.Tag(i)).Lookup(RequestParamTagName)
			if ok {
				k.ParseRequestType(prefix,field)
			}
		}
	case *types.Pointer:
		k.GenerateRequestType(prefix,requestType.(*types.Pointer).Elem())
	case *types.Interface:
		return
	case *types.Slice:
		k.GenerateRequestType(prefix,requestType.(*types.Slice).Elem())
	case *types.Map:
		k.GenerateRequestType(prefix,requestType.(*types.Map).Elem())
	case *types.Chan:
		return
	case *types.Array:
		k.GenerateRequestType(prefix,requestType.(*types.Array).Elem())
	case *types.Basic:
		rt.
		k.GenerateRequestTypeBasic(prefix,requestType.(*types.Basic))
	}

}