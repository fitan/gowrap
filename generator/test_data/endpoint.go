package test_data

import (
	"context"

	endpoint "github.com/go-kit/kit/endpoint"
)

var HelloMethodName = "Hello"
var HelloBodyMethodName = "HelloBody"
var SayHelloMethodName = "SayHello"

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
		return Response{Data: map[string]interface{}{
			"res": res,
			"err": err}, Error: err}, err
	}
}
func makeHelloBodyEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(HelloRequest)
		list, total, err := s.HelloBody(ctx, req)
		return Response{Data: map[string]interface{}{
			"list":  list,
			"total": total,
			"err":   err}, Error: err}, err
	}
}
func makeSayHelloEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(HelloRequest)
		res, err := s.SayHello(ctx, req.UUID, req.Vm.Ip, req.Vm.Port, req.HeaderName)
		return Response{Data: map[string]interface{}{
			"res": res,
			"err": err}, Error: err}, err
	}
}
