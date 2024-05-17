package generator

import (
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/fitan/gowrap/xtype"
	"github.com/fitan/jennifer/jen"
	"golang.org/x/tools/go/packages"
)

type KitResponse struct {
	Pkg *packages.Package

	ResponseNames []string

	ResponseMap map[string]*xtype.Type
}

func (k *KitResponse) xtypeParse(codes *[]*jen.Statement, names []string, xt *xtype.Type) *jen.Statement {
	if xt.Basic {
		return jen.Id(xt.BasicType.String())
	}

	if xt.Map {
		return jen.Map(k.xtypeParse(codes, names, xt.MapKey)).Add(k.xtypeParse(codes, names, xt.MapValue))
	}

	if xt.List {
		return jen.Index().Add(k.xtypeParse(codes, names, xt.ListInner))
	}

	if xt.Interface {
		return jen.Interface()
	}

	if xt.Pointer {
		return jen.Op("*").Add(k.xtypeParse(codes, names, xt.PointerInner))
	}

	if xt.Named {
		if !strings.Contains(xt.T.String(), "/") {
			// spew.Dump(xt.NamedType.Obj())
			return jen.Id(xt.NamedType.Obj().Pkg().Name()).Dot(xt.NamedType.Obj().Name())
		}
	}

	if xt.Struct {
		structNewName := strings.Join(names, "")
		*codes = append(*codes, jen.Type().Id(structNewName).StructFunc(func(g *jen.Group) {
			for i := 0; i < xt.StructType.NumFields(); i++ {
				field := xt.StructType.Field(i)
				xtField := xtype.TypeOf(field.Type())
				if xtField.Struct {
					names = append(names, field.Name())
				}

				tag := xt.StructType.Tag(i)
				g.Id(field.Name()).Add(k.xtypeParse(codes, names, xtField)).Tag(ParseTagIntoMap(tag))
			}
		}))

		return jen.Id(structNewName)
	}

	return nil
}

func ParseTagIntoMap(tag string) map[string]string {
	result := make(map[string]string)
	tags := strings.Split(tag, " ")
	for _, t := range tags {
		kv := strings.Split(t, ":")
		if len(kv) == 2 {
			// 去除引号
			value := strings.Trim(kv[1], "`\"")
			result[kv[0]] = value
		}
	}
	return result
}

func (k *KitResponse) Parse(code *jen.Statement, xt *xtype.Type, name string) string {
	codes := make([]*jen.Statement, 0)
	code = code.Type().Id(name).StructFunc(func(g *jen.Group) {
		if xt.Struct {
			for i := 0; i < xt.StructType.NumFields(); i++ {
				field := xt.StructType.Field(i)
				fmt.Println(field.Name())
				xtField := xtype.TypeOf(field.Type())
				spew.Dump(xtField)
				tag := xt.StructType.Tag(i)
				names := []string{name, field.Name()}
				g.Id(field.Name()).Add(k.xtypeParse(&codes, names, xtField)).Tag(ParseTagIntoMap(tag))
			}
		}
	})
	// code = code.Type().Id(name).Add(k.xtypeParse(code, []string{name}, xt))
	for _, c := range codes {
		code.Line().Add(c)
	}
	return code.GoString()

}
