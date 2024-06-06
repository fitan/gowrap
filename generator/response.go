package generator

import (
	"fmt"
	"strings"

	"github.com/fitan/gowrap/xtype"
	"github.com/fitan/jennifer/jen"
	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"golang.org/x/tools/go/packages"
)

func NewType2Ast(pkg *packages.Package) *type2ast {
	return &type2ast{
		current: make(map[string]struct{}),
		pkg:     pkg,
	}
}

type type2ast struct {
	pkg     *packages.Package
	current map[string]struct{}
}

func (t *type2ast) xtypeParse(codes *[]*jen.Statement, names []string, xt *xtype.Type) *jen.Statement {
	if xt.Basic {
		return jen.Id(xt.BasicType.String())
	}

	if xt.Map {
		return jen.Map(t.xtypeParse(codes, names, xt.MapKey)).Add(t.xtypeParse(codes, names, xt.MapValue))
	}

	if xt.List {
		return jen.Index().Add(t.xtypeParse(codes, names, xt.ListInner))
	}

	if xt.Interface {
		return jen.Interface()
	}

	if xt.Pointer {
		return jen.Op("*").Add(t.xtypeParse(codes, names, xt.PointerInner))
	}

	var typeName string

	if xt.Named {
		if !strings.Contains(xt.T.String(), "/") {
			return jen.Id(xt.T.String())
		} else {
			xtSplit := strings.Split(xt.T.String(), "/")
			lastName, _ := lo.Last(xtSplit)
			typeName = strcase.ToCamel(lastName)
			// typeName = strings.ReplaceAll(lastName, ".", "")
			// typeName = strings.ReplaceAll(typeName, "_", "")
			// typeName = strings.ReplaceAll(strings.TrimPrefix(xt.T.String(), t.pkg.PkgPath), ",", "")
			fmt.Println(xt.T.String())
			// if !strings.Contains(xt.T.String(), "/") {
			// 	// spew.Dump(xt.NamedType.Obj())
			// 	return jen.Id(xt.NamedType.Obj().Pkg().Name()).Dot(xt.NamedType.Obj().Name())
			// }
		}
	}

	if xt.Struct {
		// structNewName := strings.Join(names, "")
		if _, ok := t.current[typeName]; !ok {
			*codes = append(*codes, jen.Type().Id(typeName).StructFunc(func(g *jen.Group) {
				for i := 0; i < xt.StructType.NumFields(); i++ {
					var newNames []string
					field := xt.StructType.Field(i)
					xtField := xtype.TypeOf(field.Type())
					if xtField.Struct || xtField.Pointer || xtField.List || xtField.Map {
						newNames = append(names, field.Name())
					} else {
						newNames = names
					}

					tagMap := ParseTagIntoMap(xt.StructType.Tag(i))

					if !CheckTag(tagMap) {
						continue
					}

					g.Id(field.Name()).Add(t.xtypeParse(codes, newNames, xtField)).Tag(tagMap)
				}
			}))
			t.current[typeName] = struct{}{}
		}

		return jen.Id(typeName)
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

func CheckTag(tags map[string]string) bool {
	if _, ok := tags["param"]; !ok {
		return true
	}

	if strings.HasPrefix(tags["param"], `ctx,`) {
		return false
	}

	return true
}

func (t *type2ast) Parse(xt *xtype.Type, name string) string {
	fmt.Println(xt.T.String())
	if _, ok := t.current[name]; ok {
		return ""
	}

	t.current[xt.T.String()] = struct{}{}

	codes := make([]*jen.Statement, 0)
	code := jen.Type().Id(name).StructFunc(func(g *jen.Group) {
		if xt.Struct {
			for i := 0; i < xt.StructType.NumFields(); i++ {
				field := xt.StructType.Field(i)
				xtField := xtype.TypeOf(field.Type())
				tag := xt.StructType.Tag(i)
				tagMap := ParseTagIntoMap(tag)
				if !CheckTag(tagMap) {
					continue
				}
				names := []string{name, field.Name()}
				g.Id(field.Name()).Add(t.xtypeParse(&codes, names, xtField)).Tag(tagMap)
			}
		}
	})
	// code = code.Type().Id(name).Add(k.xtypeParse(code, []string{name}, xt))
	for _, c := range codes {
		code.Line().Add(c)
	}
	return code.GoString()

}
