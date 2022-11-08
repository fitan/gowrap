
package test_data

import (
	"context"
	"encoding/json"
	"fmt"
	govalidator "github.com/asaskevich/govalidator"
	encode "github.com/fitan/gowrap/generator/test_data/encode"
	"github.com/fitan/gowrap/generator/test_data/nest"
	endpoint "github.com/go-kit/kit/endpoint"
	http "github.com/go-kit/kit/transport/http"
	mux "github.com/gorilla/mux"
	errors "github.com/pkg/errors"
	cast "github.com/spf13/cast"
	otel "go.opentelemetry.io/otel"
	attribute "go.opentelemetry.io/otel/attribute"
	codes "go.opentelemetry.io/otel/codes"
	trace "go.opentelemetry.io/otel/trace"
	zap "go.uber.org/zap"
	http1 "net/http"
	"strings"
	"time"
)

const HelloMethodName = "Hello"
const HelloBodyMethodName = "HelloBody"
const SayHelloMethodName = "SayHello"

var MethodNameList = []string{HelloMethodName, HelloBodyMethodName, SayHelloMethodName}

type Endpoints struct {
	HelloEndpoint     endpoint.Endpoint
	HelloBodyEndpoint endpoint.Endpoint
	SayHelloEndpoint  endpoint.Endpoint
}

func NewEndpoint(s Service, dmw map[string][]endpoint.Middleware) Endpoints {
	eps := Endpoints{HelloEndpoint: makeHelloEndpoint(s), HelloBodyEndpoint: makeHelloBodyEndpoint(s), SayHelloEndpoint: makeSayHelloEndpoint(s)}
	for _, m := range dmw[HelloMethodName] {
		eps.HelloEndpoint = m(eps.HelloEndpoint)
	}
	for _, m := range dmw[HelloBodyMethodName] {
		eps.HelloBodyEndpoint = m(eps.HelloBodyEndpoint)
	}
	for _, m := range dmw[SayHelloMethodName] {
		eps.SayHelloEndpoint = m(eps.SayHelloEndpoint)
	}

	return eps
}
func makeHelloEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(HelloRequest)
		res, err := s.Hello(ctx, req.ID, req.Namespace, req.Paging.Page, req.Paging.Size, req.LastNames)
		return encode.WrapResponse(res, err)

	}
}
func makeHelloBodyEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(HelloRequest)
		list, total, err := s.HelloBody(ctx, req)
		return NopEndpointWrap(map[string]interface{}{
			"list":  list,
			"total": total}, err)

	}
}
func makeSayHelloEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(HelloRequest)
		m, err := s.SayHello(ctx, req.UUID, req.Vm.Ip, req.Vm.Port, req.HeaderName)
		return m, err

	}
}

type Mws map[string][]endpoint.Middleware

func MethodAddMws(mw Mws, m endpoint.Middleware, methods []string) {
	for _, v := range methods {
		mw[v] = append(mw[v], m)
	}
}

