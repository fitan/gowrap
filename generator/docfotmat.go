package generator

import (
	"go/ast"
	"strings"
)



type AstDocFormat struct {
	doc *ast.CommentGroup
}

func NewAstDocFormat(doc *ast.CommentGroup) *AstDocFormat {
	return &AstDocFormat{doc: doc}
}

// 包含字符
func (a *AstDocFormat) Contains(s string) bool {
	if a.doc == nil {
		return false
	}
	return strings.Contains(a.doc.Text(), s)
}

func (a *AstDocFormat) ContainsMark(s string) bool {
	if a.doc == nil {
		return false
	}
	for _, v := range a.doc.List {
		fs := strings.Fields(v.Text)
		if len(fs) >=2 {
			if fs[1] == s {
				return true
			}
		}
	}

	return false
}

// @xxx xxx1 xxx2  -> []string{xxx1,xxx2}
func (a *AstDocFormat) MarkValues(mark string) []string {
	if !a.ContainsMark(mark) {
		return []string{}
	}
	for _, v := range a.doc.List {
		fs := strings.Fields(v.Text)
		if len(fs) >=2 {
			if fs[1] == mark {
				return fs[2:]
			}
		}
	}

	return []string{}
}

// mapping
func (a *AstDocFormat) MarkValuesMapping(mark string, mapping ...*string) {
	vs := a.MarkValues(mark)
	vsLen := len(vs)
	mappingLen := len(mapping)
	for i := 0; i < mappingLen; i++ {
		if i < vsLen {
			*mapping[i] = vs[i]
		} else {
			*mapping[i] = ""
		}
	}
}