package test_data

import "github.com/fitan/gowrap/generator/test_data/nest"

type HelloRequest struct {
	ID int `param:"path,id"`
	Body struct{
		Name string `json:"name"`
		Age string `json:"age"`
	} `param:"body,user"`
	LastNames []string `param:"query,lastNames"`
	Paging
	Vm nest.Vm
}

type Paging struct {
	Page int64 `param:"query,page"`
	Size int64 `param:"query,size"`
}