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

var queryCode map[string]func(fieldName string, path *jen.Statement) jen.Code = map[string]func(fieldName string, path *jen.Statement) jen.Code{
	"eq": func(fieldName string, path *jen.Statement) jen.Code {
		return jen.Id("db").Op("=").Id("db").Dot("Where").Call(jen.Lit(fieldName+" = ?"), path)
	},
	"ne": func(fieldName string, path *jen.Statement) jen.Code {
		return jen.Id("db").Op("=").Id("db").Dot("Where").Call(jen.Lit(fieldName+" != ?"), path)
	},
	"gt": func(fieldName string, path *jen.Statement) jen.Code {
		return jen.Id("db").Op("=").Id("db").Dot("Where").Call(jen.Lit(fieldName+" > ?"), path)
	},
	"ge": func(fieldName string, path *jen.Statement) jen.Code {
		return jen.Id("db").Op("=").Id("db").Dot("Where").Call(jen.Lit(fieldName+" >= ?"), path)
	},
	"lt": func(fieldName string, path *jen.Statement) jen.Code {
		return jen.Id("db").Op("=").Id("db").Dot("Where").Call(jen.Lit(fieldName+" < ?"), path)
	},
	"le": func(fieldName string, path *jen.Statement) jen.Code {
		return jen.Id("db").Op("=").Id("db").Dot("Where").Call(jen.Lit(fieldName+" <= ?"), path)
	},
	"like": func(fieldName string, path *jen.Statement) jen.Code {
		return jen.Id("db").Op("=").Id("db").Dot("Where").Call(jen.Lit(fieldName+" LIKE ?"), path)
	},
	"nlike": func(fieldName string, path *jen.Statement) jen.Code {
		return jen.Id("db").Op("=").Id("db").Dot("Where").Call(jen.Lit(fieldName+" NOT LIKE ?"), path)
	},
	"ilike": func(fieldName string, path *jen.Statement) jen.Code {
		return jen.Id("db").Op("=").Id("db").Dot("Where").Call(jen.Lit(fieldName+" ILIKE ?"), path)
	},
	"nilike": func(fieldName string, path *jen.Statement) jen.Code {
		return jen.Id("db").Op("=").Id("db").Dot("Where").Call(jen.Lit(fieldName+" NOT ILIKE ?"), path)
	},
	"in": func(fieldName string, path *jen.Statement) jen.Code {
		return jen.Id("db").Op("=").Id("db").Dot("Where").Call(jen.Lit(fieldName+" IN ?"), path)
	},
	"nin": func(fieldName string, path *jen.Statement) jen.Code {
		return jen.Id("db").Op("=").Id("db").Dot("Where").Call(jen.Lit(fieldName+" NOT IN ?"), path)
	},
	"between": func(fieldName string, path *jen.Statement) jen.Code {
		return jen.Id("db").Op("=").Id("db").Dot("Where").Call(jen.Lit(fieldName+" BETWEEN ? AND ?"), path.Index(jen.Id("0")), path.Index(jen.Id("1")))
	},
	"nbetween": func(fieldName string, path *jen.Statement) jen.Code {
		return jen.Id("db").Op("=").Id("db").Dot("Where").Call(jen.Lit(fieldName+" NOT BETWEEN ? AND ?"), path.Index(jen.Id("0")), path.Index(jen.Id("1")))
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

func NewQuery(prePath []string) *Query {
	return &Query{
		PrePath: prePath,
		List:    make([]QueryMsg, 0),
	}
}

type Query struct {
	Point   bool
	Named   string
	PrePath []string
	List    []QueryMsg
	Or      []NestQuery
	Sub     []NestQuery
	// 嵌套的where
	Where []*Query
}

type NestQuery struct {
	ForeignKey string
	References string
	Table      string
	Op         string
	Query      *Query
}

type QueryMsg struct {
	Column string
	Op     string
	Point  bool
	PATH   string
}

type TagMsg struct {
	ForeignKey string
	References string
	T          string
	Op         string
	Table      string
	Column     string
}

func parseTag(field *types.Var, tag string) *TagMsg {
	tagQuery, ok := reflect.StructTag(tag).Lookup("query")
	if !ok {
		return nil
	}
	tagMsg := &TagMsg{}
	tagQueryList := strings.Split(tagQuery, ";")
	for _, v := range tagQueryList {
		switch strings.Split(v, ":")[0] {
		case "type":
			tagMsg.T = strings.TrimPrefix(v, "type:")
		case "table":
			tagMsg.Table = strings.TrimPrefix(v, "table:")
		case "op":
			tagMsg.Op = strings.TrimPrefix(v, "op:")
		case "foreignKey":
			tagMsg.ForeignKey = strings.TrimPrefix(v, "foreignKey:")
		case "references":
			tagMsg.References = strings.TrimPrefix(v, "references:")
		}
	}
	if tagMsg.Op == "" {
		tagMsg.Op = "eq"
	}

	column := schema.ParseTagSetting(reflect.StructTag(tag).Get("gorm"), ";")["COLUMN"]
	if column == "" {
		column = stringy.New(field.Name()).SnakeCase().ToLower()
	}

	tagMsg.Column = column
	return tagMsg
}

func (g *GenCallQuery) parseType(t types.Type, query *Query) {
	switch v := t.(type) {
	case *types.Struct:
		g.parseStruct(v, query)
	case *types.Named:
		if namedStruct, ok := v.Underlying().(*types.Struct); ok {
			g.parseStruct(namedStruct, query)
		}
	case *types.Pointer:
		if pointStruct, ok := v.Elem().(*types.Struct); ok {
			q := NewQuery(query.PrePath)
			q.Point = true
			query.Where = append(query.Where, q)
			g.parseStruct(pointStruct, q)
		}

		if pointNamedStruct, ok := v.Elem().(*types.Named).Underlying().(*types.Struct); ok {
			q := NewQuery(query.PrePath)
			q.Point = true
			q.Named = v.Elem().(*types.Named).Obj().Name()
			query.Where = append(query.Where, q)
			g.parseStruct(pointNamedStruct, q)
		}
	}
}

func (g *GenCallQuery) parseStruct(v *types.Struct, query *Query) error {
	for i := 0; i < v.NumFields(); i++ {
		field := v.Field(i)
		t := v.Tag(i)

		if !field.Exported() {
			continue
		}
		tagMsg := parseTag(field, t)

		// query tag 不存在 如果是结构体则递归
		if tagMsg == nil {
			g.parseType(field.Type(), query)
			continue
		}

		switch tagMsg.T {
		case "sub":
			subQuery := NewQuery(append(query.PrePath, field.Name()))
			query.Sub = append(query.Sub, NestQuery{
				ForeignKey: tagMsg.ForeignKey,
				References: tagMsg.References,
				Table:      tagMsg.Table,
				Op:         tagMsg.Op,
				Query:      subQuery,
			})
			g.parseStruct(field.Type().(*types.Struct), subQuery)
			continue
		case "or":
			orQuery := NewQuery(append(query.PrePath, field.Name()))
			query.Or = append(query.Or, NestQuery{
				ForeignKey: tagMsg.ForeignKey,
				References: tagMsg.References,
				Table:      tagMsg.Table,
				Op:         tagMsg.Op,
				Query:      orQuery,
			})
			g.parseStruct(field.Type().(*types.Struct), orQuery)
			continue
		}

		_, hasPoint := field.Type().(*types.Pointer)
		query.List = append(
			query.List, QueryMsg{
				Column: tagMsg.Column,
				Op:     tagMsg.Op,
				Point:  hasPoint,
				PATH:   strings.Join(append(query.PrePath, field.Name()), "."),
			})
	}
	return nil
}

func (g *GenCallQuery) gen(query *Query) *jen.Statement {
	block := jen.Null()
	valBind := jen.Null()
	code := jen.Func().Params(jen.Id("db").Id("*gorm.DB")).Params(jen.Id("*gorm.DB")).Block(
		block,
		jen.Return(jen.Id("db")),
	)

	if query.Point == true {
		p := query.PrePath
		if query.Named != "" {
			p = append(p, query.Named)
		}
		block.If(jen.Id("v." + strings.Join(p, ".")).Op("!=").Nil()).Block(
			valBind,
		).Line()
	} else {
		block.Add(valBind)
	}

	for _, v := range query.List {
		if v.Point {
			valBind.If(jen.Id("v." + v.PATH).Op("!=").Nil()).Block(
				queryCode[v.Op](v.Column, jen.Id("v."+v.PATH)),
			).Line()
		} else {
			valBind.Add(queryCode[v.Op](v.Column, jen.Id("v."+v.PATH))).Line()
		}
	}

	for _, v := range query.Or {
		valBind.Id("db").Op("=").Id("db").Dot("Or").Call(
			g.gen(v.Query).Call(jen.Id("db").Dot("Session").Call(jen.Id("&gorm.Session{NewDB: true}")))).Line()
	}

	for _, v := range query.Sub {
		valBind.Add(queryCode[v.Op](v.ForeignKey, g.gen(v.Query).Call(
			jen.Id("db").Dot("Session").Call(jen.Id("&gorm.Session{NewDB: true}")).Dot("Table").Call(jen.Lit(v.Table)).Dot("Select").Call(jen.Lit(v.References)),
		))).Line()
	}

	for _, v := range query.Where {
		whereCode := jen.Id("db").Op("=").Id("db").Dot("Where").Call(
			g.gen(v).Call(jen.Id("db").Dot("Session").Call(jen.Id("&gorm.Session{NewDB: true}"))),
		).Line()
		p := v.PrePath
		if v.Point {
			if v.Named != "" {
				p = append(p, v.Named)
				valBind.If(jen.Id("v." + strings.Join(p, ".")).Op("!=").Nil()).Block(
					whereCode,
				).Line()
			} else {
				valBind.If(jen.Id("v." + strings.Join(p, ".")).Op("!=").Nil()).Block(
					whereCode,
				).Line()
			}
		} else {
			valBind.Add(whereCode)
		}
	}
	return code
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

	_, ok := argT.Underlying().(*types.Struct)
	if !ok {
		return fmt.Errorf("plug query: %s fn args must be struct", name)
	}

	q := NewQuery([]string{})
	g.parseType(argT, q)

	code := g.gen(q)

	s := strings.Replace(argT.String(), g.genCall.GenOption.Pkg.PkgPath+".", "", -1)
	sSplit := strings.Split(s, "/")
	s = sSplit[len(sSplit)-1]

	jenF.Add(jen.Func().Id(name).Params(jen.Id("v").Id(s)).Parens(jen.Func().Params(jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB")).Params(jen.Id("*gorm.DB"))).Block(
		jen.Return(
			code,
		)))

	return nil
}

func (g *GenCallQuery) JenF(name string) *jen.File {
	return g.jenFM[name]
}

func NewGenCallQuery(fn *GenCall) *GenCallQuery {
	return &GenCallQuery{recorder: map[string]struct{}{}, genCall: fn, jenFM: map[string]*jen.File{}}
}
