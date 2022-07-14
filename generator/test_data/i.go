package test_data

import (
	"context"
	"github.com/fitan/gowrap/generator/test_data/nest"
)

type HelloRequest struct {
	ID   int    `param:"path,id"`
	UUID string `param:"path,uuid"`
	Time int64  `param:"path,time"`
	Body struct {
		Name string `json:"name"`
		Age  string `json:"age"`
	} `param:"body,user"`
	LastNames    []string `param:"query,lastNames"`
	LastNamesInt []int    `param:"query,lastNamesInt"`
	Paging
	Vm         nest.Vm
	HeaderName string `param:"header,name"`
	// @kit-request ctx middleware.ContextKeyNamespaceList
	Namespace []string `param:"query,namespace"`
}

type Paging struct {
	Page int64 `param:"query,page"`
	Size int64 `param:"query,size"`
}

type Service interface {
	// @kit-http /hello/{id} GET
	// @kit-http-request HelloRequest
	Hello(ctx context.Context,id int, namespace []string, page int, size int) (res string, err error)
}