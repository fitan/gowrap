package generator

import (
	"fmt"
	"github.com/pkg/errors"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"strings"
)

const (
	KitHttp = "@kit-http" 
	KitService = "@kit-http-service"
	kitParam = "@kit-http-param"
)

type KitCommentConf struct {
	Url string
	UrlMethod string

	HttpParams map[string]HttpParam
	KitServiceParam KitServiceParam
}

type HttpParam struct {
	MethodParamName string
	SourceType string
	SourceParamName string
	Validate string
	Annotation string
}

type KitServiceParam struct {
	Has bool
	EndpointName string
	DecodeName string
	EncodeName string
}

func KitComment(comments []*ast.Comment) (kitCommentConf KitCommentConf,err error) {
	kitCommentConf.HttpParams = make(map[string]HttpParam, 0)
	for _, comment := range comments {
		fields := strings.Fields(strings.TrimSpace(comment.Text))
		switch fields[0] {
		case KitHttp:
			err = (&kitCommentConf).ParamKitHttp(fields)
			if err != nil {
				err = errors.Wrap(err, comment.Text)
				return
			}
		case kitParam:
			err = (&kitCommentConf).ParamKitParam(fields)
			if err != nil {
				err = errors.Wrap(err, comment.Text)
				return
			}

		case KitService:
			err = (&kitCommentConf).ParamKitService(fields)
			if err != nil {
				err = errors.Wrap(err, comment.Text)
				return
			}
		}
	}
	return
}

func (m *KitCommentConf) ParamKitHttp(s []string) (err error) {
	if len(s) < 2 {
		err = errors.New("must format: @kit-http url method")
		return
	}
	m.Url = s[1]
	m.UrlMethod = s[2]
	return
}

func (m *KitCommentConf) ParamKitService(s []string) (err error) {
	if len(s) < 4 {
		err = errors.New("must format: @kit-service endpoint decode encode")
		return
	}
	m.KitServiceParam.Has = true
	m.KitServiceParam.EndpointName = s[1]
	m.KitServiceParam.DecodeName = s[2]
	m.KitServiceParam.EncodeName = s[3]
	return
}

func (m *KitCommentConf) ParamKitParam(s []string) (err error) {
	if len(s) < 6 {
		err = errors.New("must format: @kit-param methodParamName sourceType sourceParamName validate annotation")
		return
	}

	httpParam := HttpParam{
		MethodParamName: s[1],
		SourceType:      s[2],
		SourceParamName: s[3],
		Validate:        s[4],
		Annotation:      s[5],
	}

	m.HttpParams[s[1]] = httpParam
	return
}

type Kit struct {
	Comment KitCommentConf
	InterfaceMethodParams map[string]InterfaceMethodParam
	HasCtx bool
}

type InterfaceMethodParam struct {
	ParamName string
	ParamType string
}

func NewKit(interfaceName string,methodName string, srcPkg *packages.Package,  fi *ast.Field) (*Kit,error) {
	kit := Kit{
		InterfaceMethodParams: make(map[string]InterfaceMethodParam, 0),
	}
	commentConf, err := KitComment(fi.Doc.List)
	if err != nil {
		return nil,err
	}

	kit.Comment = commentConf

	obj := srcPkg.Types.Scope().Lookup(interfaceName)
	objInterface := obj.Type().Underlying().(*types.Interface)

	for i := 0; i < objInterface.NumMethods(); i++ {
		method := objInterface.Method(i)
		signature := method.Type().(*types.Signature)
		if method.Name() == methodName {
			params := signature.Params()
			for i := 0; i < params.Len(); i++ {
				param := params.At(i)
				if param.Type().String() == "context.Context" {
					kit.HasCtx = true
					continue
				}

				if _, ok := kit.Comment.HttpParams[param.Name()]; !ok {
					return nil,errors.New(
						fmt.Sprintf(
							"%s.%s param %s not found in @kit-http comment", interfaceName, methodName, param.Name(),
						),
					)
				}

				kit.InterfaceMethodParams[param.Name()] = InterfaceMethodParam{
					ParamName: param.Name(),
					ParamType: param.Type().String(),
				}
			}
		}
	}
	return &kit,nil
}
