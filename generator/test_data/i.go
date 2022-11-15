package test_data

import (
	"context"
	// @extra "gitlab.creditease.corp/paas/paas-assets/src/middleware"
	// @extra "github.com/fitan/gowrap/generator/test_data/encode"
	"fmt"
	"github.com/fitan/gowrap/generator/test_data/nest"
)


type NestAccount struct {
	Slice []nest.Account
}

type TestAccount struct {
	Slice []Account
}

// Account represents a user account.
type Account struct {
	ID    int
	Name  string
	Email string
	Ips   []DomainUser
}

type CopyAccount struct {
	ID    int
	Name  string
	Email string
	Ips   []Password
}

// DomainUser represents a user in relation to business logic.
type DomainUser struct {
	UserID   int
	Username string
	Password Password // The fields of Password operate at depth level 3.
}

// Password represents a password in relation to business logic.
type Password struct {
	Password string
	Hash     string
	Salt     string
}

type CtxKeyRole string

// Code generated by gowrap. DO NOT EDIT.
type HelloRequest struct {
	// ID is the ID of the user.
	ID          int    `param:"path,id"`
	UUID        string `param:"path,uuid"`
	Time        int64  `param:"path,time"`
	ParentName  *string
	ParentNames []*string
	FatherNames *[]string
	Body        struct {
		Name string `json:"name"`
		Age  string `json:"age"`
	} `param:"body,user"`
	SayHi []struct {
		Say string "json:\"say\""
		Hi  string "json:\"hi\""
	}
	// LastNames is the last names of the user.
	LastNames    []string `param:"query,lastNames"`
	LastNamesInt []int    `param:"query,lastNamesInt"`
	Paging
	Vm  nest.Vm
	VMS []nest.Vm
	// @dto-method fmt Sprintf
	HeaderName string   `param:"header,headerName"`
	Namespace  []string `param:"query,namespace"`
	VMMap      map[string]nest.Vm
	// @dto-method RoleToRole
	//Role string `param:"ctx,role"`
}

type HelloRequestCopy struct {
	// ID is the ID of the user.
	ID          int    `param:"path,id"`
	UUID        string `param:"path,uuid"`
	Time        int64  `param:"path,time"`
	ParentName  *string
	ParentNames []*string
	FatherNames *[]string
	Body        struct {
		Name string `json:"name"`
		Age  string `json:"age"`
	} `param:"body,user"`
	SayHi []struct {
		Say string "json:\"say\""
		Hi  string "json:\"hi\""
	}
	// LastNames is the last names of the user.
	LastNames    []string `param:"query,lastNames"`
	LastNamesInt []int    `param:"query,lastNamesInt"`
	Paging
	Vm  nest.Vm
	VMS []nest.Vm
	// @dto-method fmt Sprintf
	HeaderName string   `param:"header,headerName"`
	Namespace  []string `param:"query,namespace"`
	VMMap      map[string]nest.Vm
	// @dto-method RoleToRole
	//Role string `param:"ctx,role"`
}

type Query struct {
	Name string `json:"name" query:"op:eq" gorm:"column:name"`
	Age  *int   `json:"age" query:"op:gt" gorm:"column:age"`
	Ids  []int  `json:"ids" query:"op:in" gorm:"column:id"`
	VM   struct {
		Ip string `json:"ip" query:"op:eq"`
	} `json:"vm" query:"op:eq;type:sub;table:vm;foreignKey:vm_uuid;references:uuid"`
	Email struct{
		Name string `json:"name" query:"op:eq"`

		Project struct{
			Id int `json:"id" query:"op:eq"`
			Device struct{
				Port int `json:"port" query:"op:eq"`
			} `json:"device" query:"op:eq;type:sub;table:device;foreignKey:device_id;references:id"`
		} `json:"project" query:"op:eq;type:sub;table:project;foreignKey:project_id;references:id"`

	} `json:"email" query:"op:eq;type:or;table:email;foreignKey:email_uuid;references:uuid"`
	*PM
	Brand
}

type PM struct {
	NetMask *string `json:"netMask" query:"op:eq"`
	Limit int `json:"limit" query:"op:eq"`
}

type Brand struct {
	Model string `json:"model" query:"op:eq"`
}

func RoleToRole(s string) string {
	return s
}

func HeaderNameToHeaderName(s string) string {
	return fmt.Sprintf(s)
}

type Paging struct {
	Page int64 `param:"query,page"`
	Size int64 `param:"query,size"`
}

type Middleware func(service Service) Service

// @tags paas-api
// @impl
type Service interface {
	// Hello
	// @kit-http /hello/{id} GET
	// @kit-http-request HelloRequest
	// @kit-http-response HelloRequest
	Hello(ctx context.Context, id int, namespace []string, page int64, size int64, lastNames []string) (res HelloRequest, err error)
	// SayHello
	// @kit-http /hello/say GET
	// @kit-http-request HelloRequest
	// @kit-http-endpoint-wrap false
	// @kit-http-response HelloRequest
	SayHello(ctx context.Context, uuid string, ip string, port int, headerName string) (m map[string][]nest.NetWork, err error)

	// HelloBody
	// @kit-http /hello/body GET
	// @kit-http-request HelloRequest body
	// @kit-http-endpoint-wrap NopEndpointWrap
	// @kit-http-response HelloRequest
	// @kit-cache get helloRequest method GetRedisKey 1s
	// @kit-cache delete helloRequest key 1s
	// @kit-cache put helloRequest Interface 1s
	HelloBody(ctx context.Context, helloRequest HelloRequest) (list HelloRequest, total int64, err error)
	//nest.Nest
}

type RedisService struct {
}

func (r RedisService) Hello(
	ctx context.Context, id int, namespace []string, page int64, size int64, lastNames []string,
) (res HelloRequest, err error) {
	panic("implement me")
}

func (r RedisService) SayHello(
	ctx context.Context, uuid string, ip string, port int, headerName string,
) (res HelloRequest, err error) {
	src := HelloRequestCopy{}
	// @call copy
	res = copyDTO(src)

	q := Query{}

	// @call query
	queryM = queryDTO(q)
	panic("implement me")
}

func (r RedisService) HelloBody(ctx context.Context, helloRequest HelloRequest) (
	list HelloRequest, total int64, err error,
) {
	panic("implement me")
}
