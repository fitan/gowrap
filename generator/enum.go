package generator

import (
	"fmt"
	"github.com/fitan/gowrap/generator/parse"
	"github.com/fitan/jennifer/jen"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"strings"
)

func NewEnumGen(pkg *packages.Package) *Enum {
	return &Enum{
		Pkg: pkg,
	}
}

type Enum struct {
	Doc      *parse.Doc
	Args     []Arg
	TypeName string
	Pkg      *packages.Package
}

type Arg struct {
	Key   string
	Value string
}

func (e *Enum) ToJenFString() (s string) {
	jenF := jen.NewFile(e.Pkg.Name)
	state := jen.Statement{}
	for _, f := range e.Pkg.Syntax {
		tsList, gdDoc := typeSpecs(f)

		for tsIndex, ts := range tsList {
			if i, ok := ts.Type.(*ast.Ident); ok {
				if i.String() == "int" {
					text := gdDoc[tsIndex].Text()
					if strings.Contains(text, "@enum") {
						parseEnum, err := parse.Parse(text)
						if err != nil {
							panic(err)
						}
						enum := &Enum{
							TypeName: ts.Name.String(),
							Doc:      parseEnum,
						}
						codes, err := enum.Gen()
						if err != nil {
							panic(err)
						}
						jenF.Add(codes...)
						state.Add(codes...)
					}
				}
			}
		}
	}

	return state.GoString()
}

func (e *Enum) Init() (err error) {

	args := make([]Arg, 0, 0)
	for _, v := range e.Doc.Lines {
		if v.Call != nil {
			if v.Call.Name == "@enum" {
				for _, arg := range v.Call.Args {
					arg = strings.TrimPrefix(arg, `"`)
					arg = strings.TrimSuffix(arg, `"`)

					argSplit := strings.Split(arg, ":")
					if len(argSplit) < 2 {
						err = fmt.Errorf("enum arg error: %s", arg)
						return
					}
					args = append(args, Arg{
						Key:   argSplit[0],
						Value: argSplit[1],
					})
				}
			}
		}
	}
	e.Args = args
	return nil
}

func (e *Enum) Gen() (codes []jen.Code, err error) {
	codes = make([]jen.Code, 0, 100)
	err = e.Init()
	if err != nil {
		return
	}
	if len(e.Args) == 0 {
		return
	}
	codes = append(codes, e.Const()...)
	codes = append(codes, e.Value()...)
	codes = append(codes, e.Json()...)
	codes = append(codes, e.GormSerialize()...)
	codes = append(codes, e.String()...)
	return
}

func (e *Enum) Const() (codes []jen.Code) {
	codes = append(codes, jen.Id("_").Op("=").Id("iota"))
	for _, arg := range e.Args {
		codes = append(codes, jen.Id(strings.ToTitle(arg.Key)))
	}

	return []jen.Code{jen.Const().Defs(codes...).Line().Line()}
}

func (e *Enum) Value() (codes []jen.Code) {
	varCode := jen.Var().Id("_" + e.TypeName + "Value").Op("=").Map(jen.String()).Id(e.TypeName).Values(jen.DictFunc(func(d jen.Dict) {
		for _, arg := range e.Args {
			d[jen.Lit(arg.Value)] = jen.Id(strings.ToTitle(arg.Key))
		}
	})).Line().Line()

	parseCode := jen.Func().Id("Parse"+e.TypeName).Params(jen.Id("name").String()).Params(jen.Id(e.TypeName), jen.Error()).Block(
		jen.If(jen.Id("x").Op(",").Id("ok").Op(":=").Id("_"+e.TypeName+"Value").Index(jen.Id("name")), jen.Id("ok")).Block(
			jen.Return(jen.Id("x"), jen.Nil()),
		),
		jen.Return(jen.Lit(0), jen.Qual("fmt", "Errorf").Call(jen.Lit("unknown enum value: %s"), jen.Id("name"))),
	).Line().Line()

	codes = append(codes, varCode, parseCode)
	return
}

func (e *Enum) String() (code []jen.Code) {
	stringCode := jen.Func().Params(jen.Id("e").Id(e.TypeName)).Id("String").Params().String().Block(
		jen.Switch(jen.Id("e")).BlockFunc(func(g *jen.Group) {
			for argIndex, arg := range e.Args {
				g.Case(jen.Lit(argIndex + 1)).Block(
					jen.Return(jen.Lit(arg.Value)),
				)
			}
		}),
		jen.Return(jen.Qual("fmt", "Sprintf").Call(jen.Lit("unknown %d"), jen.Id("e"))),
	).Line().Line()
	code = append(code, stringCode)
	return
}

func (e *Enum) Json() (codes []jen.Code) {
	marshalCode := jen.Func().Params(jen.Id("e").Id("*"+e.TypeName)).Id("MarshalJSON").Params().Params(jen.Id("[]byte"), jen.Id("error")).Block(
		jen.Switch(jen.Id("*e")).BlockFunc(func(g *jen.Group) {
			for _, arg := range e.Args {
				g.Case(jen.Id(strings.ToTitle(arg.Key))).Block(
					jen.Return(jen.Index().Byte().Call(jen.Id("`\""+arg.Value+"\"`")), jen.Nil()),
				)
			}
		}),
		jen.Return(jen.Nil(), jen.Qual("fmt", "Errorf").Call(jen.Lit("unknown enum value: %v"), jen.Id("e"))),
	).Line().Line()

	unmarshalCode := jen.Func().Params(jen.Id("e").Id("*"+e.TypeName)).Id("UnmarshalJSON").Params(jen.Id("data").Index().Byte()).Params(jen.Id("error")).Block(
		jen.Id("s").Op(":=").Id("string").Call(jen.Id("data")),
		jen.If(
			jen.Qual("strings", "HasPrefix").Call(jen.Id("s"), jen.Id("`\"`")).
				Op("&&").Qual("strings", "HasSuffix").Call(jen.Id("s"), jen.Id("`\"`")),
		).Block(
			jen.Id("v").Op(",").Id("err").Op(":=").Id("Parse"+e.TypeName).Call(jen.Qual("strings", "Trim").Call(jen.Id("s"), jen.Id("`\"`"))),
			jen.If(jen.Id("err").Op("!=").Nil()).Block(
				jen.Return(jen.Id("err")),
			),
			jen.Id("*e").Op("=").Id("v"),
			jen.Return(jen.Nil()),
		),
		jen.Return(jen.Qual("fmt", "Errorf").Call(jen.Lit("unknown enum value: %v"), jen.Id("s"))),
	).Line().Line()

	codes = append(codes, marshalCode, unmarshalCode)
	return
}

func (e *Enum) GormSerialize() (codes []jen.Code) {
	scanCode := jen.Func().Params(jen.Id("e").Id("*"+e.TypeName)).Id("Scan").
		Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("field").Op("*").Qual("gorm.io/gorm/schema", "Field"),
			jen.Id("dst").Qual("reflect", "Value"),
			jen.Id("dbValue").Interface(),
		).Params(jen.Id("error")).Block(
		jen.Switch(jen.Id("value").Op(":=").Id("dbValue").Assert(jen.Id("type")).Block(
			jen.Case(jen.String()).Block(
				jen.Id("*e").Op("=").Id("_"+e.TypeName+"Value").Index(jen.Id("value")),
			),
			jen.Default().Block(
				jen.Return(jen.Qual("fmt", "Errorf").Call(jen.Lit("unknown enum value: %v"), jen.Id("value"))),
			),
		),

			jen.Return(jen.Nil()),
		)).Line().Line()
	valueCode := jen.Func().Params(jen.Id("e").Id(e.TypeName)).Id("Value").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("field").Op("*").Qual("gorm.io/gorm/schema", "Field"),
		jen.Id("dst").Qual("reflect", "Value"),
		jen.Id("fieldValue").Interface(),
	).Params(jen.Id("driver.Value"), jen.Id("error")).Block(
		jen.Return(jen.Id("e").Dot("String").Call(), jen.Nil()),
	).Line().Line()

	codes = append(codes, scanCode, valueCode)
	return
}
