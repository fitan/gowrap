package generator

import (
	"github.com/fitan/gowrap/generator/parse"
	"strings"
)

func EnumGen(doc *parse.Doc) {
	for _, v := range doc.Lines {
		if v.Call != nil {
			if v.Call.Name == "@enum" {
				for _, arg := range v.Call.Args {
					arg = strings.TrimPrefix(arg, `"`)
					arg = strings.TrimSuffix(arg, `"`)

					argSplit := strings.Split(arg, ":")
				}
			}
		}
	}
}
