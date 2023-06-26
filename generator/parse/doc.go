package parse

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	enumLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"whitespace", `\s+`},
		{`String`, `"(?:\\.|[^"])*"|'(?:\\.|[^'])*'`},
		{"Punct", `[)(,]`},
		{"Name", `^@[a-zA-Z][a-zA-Z_\d]*`},
		{"Comment", `^[^@].+`},
	})
	parser = participle.MustBuild[Doc](
		participle.Lexer(enumLexer),
	)
)

type Doc struct {
	Lines []Line `@@*`
}

type Line struct {
	Comment *string `@Comment`
	Call    *Func   `| @@`
}

type Func struct {
	Name string   `@Name`
	Args []string `( "(" @String ("," @String)* ")" )`
}

func Parse(s string) (*Doc, error) {
	return parser.ParseString("", s)
}
