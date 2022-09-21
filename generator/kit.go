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
	KitHttp         = "@kit-http"
	KitService      = "@kit-http-service"
	kitParam        = "@kit-http-param"
	KitHttpRequest  = "@kit-http-request"
	KitHttpResponse = "@kit-http-response"
)

type KitCommentConf struct {
	Url       string
	UrlMethod string

	HttpRequestName string
	HttpRequestBody bool

	//HttpResponseName string

	HttpParams      map[string]HttpParam
	KitServiceParam KitServiceParam
}

type HttpParam struct {
	MethodParamName string
	SourceType      string
	Validate        string
	Annotation      string
}

type KitServiceParam struct {
	HasEndpointName bool
	EndpointName    string
	HasDecodeName   bool
	DecodeName      string
	HasEncodeName   bool
	EncodeName      string
}

func KitComment(comments []*ast.Comment) (kitConf KitCommentConf, err error) {
	kitConf.HttpParams = make(map[string]HttpParam, 0)
	for _, comment := range comments {
		fields := strings.Fields(strings.TrimSpace(comment.Text))
		switch fields[1] {
		case KitHttp:
			err = (&kitConf).ParamKitHttp(fields)
			if err != nil {
				err = errors.Wrap(err, comment.Text)
				return
			}
		case kitParam:
			err = (&kitConf).ParamKitParam(fields)
			if err != nil {
				err = errors.Wrap(err, comment.Text)
				return
			}

		case KitService:
			err = (&kitConf).ParamKitService(fields)
			if err != nil {
				err = errors.Wrap(err, comment.Text)
				return
			}
		case KitHttpRequest:
			err = (&kitConf).ParamKitHttpRequest(fields)
			if err != nil {
				err = errors.Wrap(err, comment.Text)
				return
			}
		//case KitHttpResponse:
		//	err = (&kitConf).ParamKitHttpResponse(fields)
		//	if err != nil {
		//		err = errors.Wrap(err, comment.Text)
		//		return
		//	}

		}
	}
	return
}

//func (m *KitCommentConf) ParamKitHttpResponse(s []string) (err error) {
//	if len(s) < 3 {
//		err = errors.New("must format: @kit-http-response responseName")
//		return
//	}
//	m.HttpResponseName = s[2]
//	return
//}

func (m *KitCommentConf) ParamKitHttpRequest(s []string) (err error) {
	if len(s) < 3 {
		err = errors.New("must format: @kit-http-request requestName ?body")
		return
	}
	m.HttpRequestName = s[2]

	if len(s) > 3 {
		isBody := s[3]
		if isBody != `""` && isBody != "false" {
			m.HttpRequestBody = true
		}
	}
	return
}

func (m *KitCommentConf) ParamKitHttp(s []string) (err error) {
	if len(s) < 3 {
		err = errors.New("must format: @kit-http url method")
		return
	}
	m.Url = s[2]
	m.UrlMethod = s[3]
	return
}

func (m *KitCommentConf) ParamKitService(s []string) (err error) {
	if len(s) < 4 {
		err = errors.New("must format: @kit-service endpoint decode encode")
		return
	}
	m.KitServiceParam.EndpointName = s[2]
	if m.KitServiceParam.EndpointName != "" {
		m.KitServiceParam.HasEndpointName = true
	}
	m.KitServiceParam.DecodeName = s[3]
	if m.KitServiceParam.DecodeName != "" {
		m.KitServiceParam.HasDecodeName = true
	}
	m.KitServiceParam.EncodeName = s[4]
	if m.KitServiceParam.EncodeName != "" {
		m.KitServiceParam.HasEncodeName = true
	}
	return
}

func (m *KitCommentConf) ParamKitParam(s []string) (err error) {
	if len(s) < 5 {
		err = errors.New("must format: @kit-param methodParamName sourceType sourceParamName validate annotation")
		return
	}

	httpParam := HttpParam{
		MethodParamName: s[2],
		SourceType:      s[3],
		Validate:        s[4],
		Annotation:      s[5],
	}

	m.HttpParams[s[2]] = httpParam
	return
}

type Kit struct {
	//Comment KitConf
	HasParamSourcePath    bool
	InterfaceMethodParams map[string]InterfaceMethodParam
	Conf                  KitCommentConf
	HasCtx                bool
}

type InterfaceMethodParam struct {
	ParamName       string
	ParamType       string
	MethodParamName string
	SourceType      string
	Validate        string
	Annotation      string
}

func NewKit(interfaceName string, methodName string, srcPkg *packages.Package, fi *ast.Field) (res Kit, err error) {
	kit := Kit{
		InterfaceMethodParams: make(map[string]InterfaceMethodParam, 0),
	}
	commentConf, err := KitComment(fi.Doc.List)
	if err != nil {
		return
	}

	kit.Conf = commentConf

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

				if _, ok := commentConf.HttpParams[param.Name()]; !ok {
					continue
					//return nil,errors.New(
					//	fmt.Sprintf(
					//		"%s.%s param %s not found in @kit-http comment", interfaceName, methodName, param.Name(),
					//	),
					//)
				}

				kit.InterfaceMethodParams[param.Name()] = InterfaceMethodParam{
					ParamName:       param.Name(),
					ParamType:       param.Type().String(),
					MethodParamName: commentConf.HttpParams[param.Name()].MethodParamName,
					SourceType:      commentConf.HttpParams[param.Name()].SourceType,
					Validate:        commentConf.HttpParams[param.Name()].Validate,
					Annotation:      commentConf.HttpParams[param.Name()].Annotation,
				}

				if commentConf.HttpParams[param.Name()].SourceType == "path" {
					kit.HasParamSourcePath = true
				}

				paramStruct, ok := param.Type().Underlying().(*types.Struct)
				if ok {
					fmt.Println("string: ", paramStruct.String())
					fmt.Println("field1: ", paramStruct.Field(0).Name())
					fmt.Println("tag: ", paramStruct.Tag(0))
				}
			}
		}
	}
	return kit, nil
}

func (k *Kit) Gen() {

}
