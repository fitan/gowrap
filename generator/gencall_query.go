package generator

import (
	"fmt"
	"github.com/fitan/jennifer/jen"
	"github.com/gobeam/stringy"
	"github.com/pkg/errors"
	"go/types"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
)

const queryGenFName = "query"

var queryCode map[string]func(fieldName string, path string) jen.Code = map[string]func(fieldName, path string) jen.Code{
	"eq": func(fieldName, path string) jen.Code {
		return jen.Func().Params(jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB")).Params(jen.Id("*gorm.DB")).Block(
			jen.Return(jen.Id("db").Dot("Where").Call(jen.Lit(fieldName+" = ?"), jen.Id(path))),
		)
	},
	"ne": func(fieldName, path string) jen.Code {
		return jen.Func().Params(jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB")).Params(jen.Id("*gorm.DB")).Block(
			jen.Return(jen.Id("db").Dot("Where").Call(jen.Lit(fieldName+" != ?"), jen.Id(path))),
		)
	},
	"gt": func(fieldName, path string) jen.Code {
		return jen.Func().Params(jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB")).Params(jen.Id("*gorm.DB")).Block(
			jen.Return(jen.Id("db").Dot("Where").Call(jen.Lit(fieldName+" > ?"), jen.Id(path))),
		)
	},
	"ge": func(fieldName, path string) jen.Code {
		return jen.Func().Params(jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB")).Params(jen.Id("*gorm.DB")).Block(
			jen.Return(jen.Id("db").Dot("Where").Call(jen.Lit(fieldName+" >= ?"), jen.Id(path))),
		)
	},
	"lt": func(fieldName, path string) jen.Code {
		return jen.Func().Params(jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB")).Params(jen.Id("*gorm.DB")).Block(
			jen.Return(jen.Id("db").Dot("Where").Call(jen.Lit(fieldName+" < ?"), jen.Id(path))),
		)
	},
	"le": func(fieldName, path string) jen.Code {
		return jen.Func().Params(jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB")).Params(jen.Id("*gorm.DB")).Block(
			jen.Return(jen.Id("db").Dot("Where").Call(jen.Lit(fieldName+" <= ?"), jen.Id(path))),
		)
	},
	"like": func(fieldName, path string) jen.Code {
		return jen.Func().Params(jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB")).Params(jen.Id("*gorm.DB")).Block(
			jen.Return(jen.Id("db").Dot("Where").Call(jen.Lit(fieldName+" LIKE ?"), jen.Id(path))),
		)
	},
	"nlike": func(fieldName, path string) jen.Code {
		return jen.Func().Params(jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB")).Params(jen.Id("*gorm.DB")).Block(
			jen.Return(jen.Id("db").Dot("Where").Call(jen.Lit(fieldName+" NOT LIKE ?"), jen.Id(path))),
		)
	},
	"ilike": func(fieldName, path string) jen.Code {
		return jen.Func().Params(jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB")).Params(jen.Id("*gorm.DB")).Block(
			jen.Return(jen.Id("db").Dot("Where").Call(jen.Lit(fieldName+" ILIKE ?"), jen.Id(path))),
		)
	},
	"nilike": func(fieldName, path string) jen.Code {
		return jen.Func().Params(jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB")).Params(jen.Id("*gorm.DB")).Block(
			jen.Return(jen.Id("db").Dot("Where").Call(jen.Lit(fieldName+" NOT ILIKE ?"), jen.Id(path))),
		)
	},
	"in": func(fieldName, path string) jen.Code {
		return jen.Func().Params(jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB")).Params(jen.Id("*gorm.DB")).Block(
			jen.Return(jen.Id("db").Dot("Where").Call(jen.Lit(fieldName+" IN ?"), jen.Id(path))),
		)
	},
	"nin": func(fieldName, path string) jen.Code {
		return jen.Func().Params(jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB")).Params(jen.Id("*gorm.DB")).Block(
			jen.Return(jen.Id("db").Dot("Where").Call(jen.Lit(fieldName+" NOT IN ?"), jen.Id(path))),
		)
	},
	"between": func(fieldName, path string) jen.Code {
		return jen.Func().Params(jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB")).Params(jen.Id("*gorm.DB")).Block(
			jen.Return(jen.Id("db").Dot("Where").Call(jen.Lit(fieldName+" BETWEEN ? AND ?"), jen.Id(path).Index(jen.Id("0")), jen.Id(path).Index(jen.Id("1")))),
		)
	},
	"nbetween": func(fieldName, path string) jen.Code {
		return jen.Func().Params(jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB")).Params(jen.Id("*gorm.DB")).Block(
			jen.Return(jen.Id("db").Dot("Where").Call(jen.Lit(fieldName+" NOT BETWEEN ? AND ?"), jen.Id(path).Index(jen.Id("0")), jen.Id(path).Index(jen.Id("1")))),
		)
	},
	//"regexp": "regexp ",
	//"nregexp": "not regexp",
	//"iregexp": "iregexp",
	//"niregexp": "not iregexp",
	//"overlap": "&&",
	//"contains": "@>",
	//"contained_by": "<@",
	//"any": "any",
	//"all": "all",
	//"null": "is null",
	//"not_null": "is not null",
	//"empty": "is empty",
	//"not_empty": "is not empty",
}

var queryMap map[string]string = map[string]string{
	"eq":     "= ?",
	"ne":     "!= ?",
	"gt":     "> ?",
	"ge":     ">= ?",
	"lt":     "< ?",
	"le":     "<= ?",
	"like":   "like ?",
	"nlike":  "not like ?",
	"ilike":  "ilike ?",
	"nilike": "not ilike ?",
	"in":     "in ?",
	"nin":    "not in ?",
	//"regexp": "regexp ",
	//"nregexp": "not regexp",
	//"iregexp": "iregexp",
	//"niregexp": "not iregexp",
	//"overlap": "&&",
	//"contains": "@>",
	//"contained_by": "<@",
	//"any": "any",
	//"all": "all",
	"between":  "between ? and ?",
	"nbetween": "not between ? and ?",
	//"null": "is null",
	//"not_null": "is not null",
	//"empty": "is empty",
	//"not_empty": "is not empty",
}

type GenCallQuery struct {
	recorder map[string]struct{}
	jenFM    map[string]*jen.File
	genCall  *GenCall
}

func (g *GenCallQuery) Name() string {
	return "query"
}

func (g *GenCallQuery) Gen() error {
	jenF := jen.NewFile(g.genCall.GenOption.Pkg.Name)
	for name, fn := range g.genCall.FuncList {
		var queryMark string
		format := &AstDocFormat{fn.Doc}
		format.MarkValuesMapping(GenCallMark, &queryMark)
		if queryMark != queryGenFName {
			continue
		}

		err := g.parse(jenF, name, fn)
		if err != nil {
			err = errors.Wrap(err, "g.parse")
			return err
		}
	}

	g.jenFM["query"] = jenF

	return nil
}

type QueryMsg struct {
	Column string
	Op     string
	Point  bool
	PATH   string
}

func (g *GenCallQuery) parseField(path []string, v *types.Var, tag string, list *[]QueryMsg) {
	tagQuery, ok := reflect.StructTag(tag).Lookup("query")
	if !ok {
		if structType, ok := v.Type().(*types.Struct); ok {
			for i := 0; i < structType.NumFields(); i++ {
				field := structType.Field(i)
				if !field.Exported() {
					continue
				}
				g.parseField(append(path, v.Name()), field, structType.Tag(i), list)
			}
		}
		return
	}

	_, hasPoint := v.Type().(*types.Pointer)
	FiledName := v.Name()
	column := schema.ParseTagSetting(reflect.StructTag(tag).Get("gorm"), ";")["COLUMN"]
	if column == "" {
		column = stringy.New(FiledName).SnakeCase().ToLower()
	}

	*list = append(*list, QueryMsg{
		Column: column,
		Op:     tagQuery,
		Point:  hasPoint,
		PATH:   strings.Join(append(path, FiledName), "."),
	})
}

func (g *GenCallQuery) parse(jenF *jen.File, name string, fn Func) error {

	if _, ok := g.recorder[name]; ok {
		return nil
	}

	g.recorder[name] = struct{}{}

	if len(fn.Args) != 1 {
		return fmt.Errorf("plug query: %s fn args count must be 1", name)
	}

	argT := fn.Args[0].T

	if v, ok := argT.(*types.Pointer); ok {
		argT = v.Elem()
	}

	argStruct, ok := argT.Underlying().(*types.Struct)
	if !ok {
		return fmt.Errorf("plug query: %s fn args must be struct", name)
	}
	queryList := []QueryMsg{}

	for i := 0; i < argStruct.NumFields(); i++ {
		g.parseField([]string{"v"}, argStruct.Field(i), argStruct.Tag(i), &queryList)
	}

	s := strings.Replace(argT.String(), g.genCall.GenOption.Pkg.PkgPath+".", "", -1)
	sSplit := strings.Split(s, "/")
	s = sSplit[len(sSplit)-1]

	setM := jen.Null().Line()

	for _, v := range queryList {
		fmt.Println(v)
		if v.Point {
			setM.If(jen.Id(v.PATH).Op("!=").Nil()).Block(
				jen.Id("res").Op("=").Append(jen.Id("res"), queryCode[v.Op](v.Column, v.PATH)),
			).Line()
		} else {
			setM.Id("res").Op("=").Append(jen.Id("res"), queryCode[v.Op](v.Column, v.PATH)).Line()
		}
	}

	jenF.Add(jen.Func().Id(name).Params(jen.Id("v").Id(s)).Parens(jen.Id("res").Index().Func().Params(jen.Id("db").Id("*gorm.DB")).Params(jen.Id("*gorm.DB"))).Block(
		setM,
		jen.Return(),
	))

	return nil
}

func (g *GenCallQuery) JenF(name string) *jen.File {
	return g.jenFM[name]
}

func NewGenCallQuery(fn *GenCall) *GenCallQuery {
	return &GenCallQuery{recorder: map[string]struct{}{}, genCall: fn, jenFM: map[string]*jen.File{}}
}
