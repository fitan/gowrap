package generator

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/fitan/gowrap/xtype"
	"go/types"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
)

var queryMap map[string]string = map[string]string{
	"eq": "= ?",
	"ne": "!= ?",
	"gt": "> ?",
	"ge": ">= ?",
	"lt": "< ?",
	"le": "<= ?",
	"like": "like ?",
	"nlike": "not like ?",
	"ilike": "ilike ?",
	"nilike": "not ilike ?",
	//"regexp": "regexp ",
	//"nregexp": "not regexp",
	//"iregexp": "iregexp",
	//"niregexp": "not iregexp",
	//"overlap": "&&",
	//"contains": "@>",
	//"contained_by": "<@",
	//"any": "any",
	//"all": "all",
	"between": "between ? and ?",
	"nbetween": "not between ? and ?",
	//"null": "is null",
	//"not_null": "is not null",
	//"empty": "is empty",
	//"not_empty": "is not empty",
}

type GenFnQuery struct {
	recorder map[string]struct{}
	jenFM map[string]*jen.File
	genFn *GenFn
}

func (g *GenFnQuery) Name() string {
	return "query"
}

func (g *GenFnQuery) Gen() error {
	jen.NewFile("query")
	for _, fn := range g.genFn.FuncList {
		if !(len(fn.MarkParam) >0 && fn.MarkParam[0] == "query") {
			continue
		}
	}

	return nil
}


func (g *GenFnQuery) parseField(path []string, v *types.Var, tag string, m map[string]string) {
	tagQuery,ok := reflect.StructTag(tag).Lookup("query")
	if !ok {
		if structType, ok := v.Type().(*types.Struct); ok {
			for i := 0; i < structType.NumFields(); i++ {
				field := structType.Field(i)
				if !field.Exported() {
					continue
				}
				g.parseField(path,field, structType.Tag(i), m)
			}
		}
		return
	}

	FiledName := v.Name()
	op := queryMap[tagQuery]
	column := schema.ParseTagSetting(reflect.StructTag(tag).Get("gorm"), ";")["COLUMN"]

	m[column + " " + op] = strings.Join(append(path, FiledName), ".")
}

func (g *GenFnQuery) parse(jenF *jen.File,name string, fn Func) error {

	if _, ok := g.recorder[name];ok {
		return nil
	}

	g.recorder[name] = struct{}{}

	if len(fn.Args) != 1 {
		return fmt.Errorf("plug query: %s fn args count must be 1", name)
	}

	arg := fn.Args[0]

	if v, ok := arg.(*types.Pointer); ok {
		arg = v.Elem()
	}

	argStruct, ok := arg.(*types.Struct)
	if !ok {
		return fmt.Errorf("plug query: %s fn args must be struct", name)
	}

	queryM := map[string]string{}

	for i := 0; i < argStruct.NumFields(); i++ {
		g.parseField([]string{"v"}, argStruct.Field(i), argStruct.Tag(i), queryM)
	}



	jenF.Add(jen.Func().Id(name).Params(jen.Id(arg.String())).Params(jen.Map(jen.String()).Interface()).Block(
	//	dict := jen.Dict{
	//		jen.Lit("v"): jen.Id("v"),
	//}
	//	jen.Return(jen.Map(jen.String()).Interface().ValuesFunc(func(g *jen.Group) {
	//		for k, v := range queryM {
	//			.Dict(jen.Lit(k), jen.Id(v.(string)))
	//		}
	//	}
	))


}



func (g *GenFnQuery) JenF(name string) *jen.File {
	return g.jenFM[name]
}

func NewGenFnQuery(fn *GenFn) *GenFnCopy {
	return &GenFnCopy{recorder: map[string]struct{}{}, genFn: fn, jenFM: map[string]*jen.File{}}
}

