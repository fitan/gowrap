package test_data

import (
	"context"
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

type DTO struct{}

// parentPath: : // path: :
func (d *DTO) test_dataHelloRequestTotest_dataHelloRequest(src HelloRequest) (dest HelloRequest) {
	dest.Role = src.Role
	dest.Body.Name = src.Body.Name
	dest.Paging.Page = src.Paging.Page
	dest.Vm.Port = src.Vm.Port
	dest.Body.Age = src.Body.Age
	dest.Paging.Size = src.Paging.Size
	dest.Vm.Ip = src.Vm.Ip
	dest.HeaderName = src.HeaderName
	dest.ID = src.ID
	dest.UUID = src.UUID
	dest.Time = src.Time
	dest.Vm.VVMMSS = make([]nest.Vm, 0, len(src.Vm.VVMMSS))
	for i := 0; i < len(src.Vm.VVMMSS); i++ {
		dest.Vm.VVMMSS[i] = d.nestVmTonestVm(src.Vm.VVMMSS[i])
	}
	dest.VMS = make([]nest.Vm, 0, len(src.VMS))
	for i := 0; i < len(src.VMS); i++ {
		dest.VMS[i] = d.nestVmTonestVm(src.VMS[i])
	}
	dest.Namespace = make([]string, 0, len(src.Namespace))
	for i := 0; i < len(src.Namespace); i++ {
		dest.Namespace[i] = src.Namespace[i]
	}
	dest.ParentNames = make([]*string, 0, len(src.ParentNames))
	for i := 0; i < len(src.ParentNames); i++ {
		dest.ParentNames[i] = d.pStringTopString(src.ParentNames[i])
	}
	dest.LastNames = make([]string, 0, len(src.LastNames))
	for i := 0; i < len(src.LastNames); i++ {
		dest.LastNames[i] = src.LastNames[i]
	}
	dest.LastNamesInt = make([]int, 0, len(src.LastNamesInt))
	for i := 0; i < len(src.LastNamesInt); i++ {
		dest.LastNamesInt[i] = src.LastNamesInt[i]
	}
	dest.Vm.NetWorks = make([]nest.NetWork, 0, len(src.Vm.NetWorks))
	for i := 0; i < len(src.Vm.NetWorks); i++ {
		dest.Vm.NetWorks[i] = d.nestNetWorkTonestNetWork(src.Vm.NetWorks[i])
	}
	dest.VMMap = make(map[string]nest.Vm, len(src.VMMap))
	for key, value := range src.VMMap {
		dest.VMMap[key] = d.nestVmTonestVm(value)
	}
	dest.ParentName = src.ParentName
	if src.FatherNames != nil {
		v := d.stringListTostringList(*src.FatherNames)
		dest.FatherNames = &v
	} else {
		dest.FatherNames = src.FatherNames
	}
	return
}

// parentPath: Vm.VVMMSS:Vm.VVMMSS // path: :
func (d *DTO) nestVmTonestVm(src nest.Vm) (dest nest.Vm) {
	dest.Ip = src.Ip
	dest.Port = src.Port
	dest.NetWorks = make([]nest.NetWork, 0, len(src.NetWorks))
	for i := 0; i < len(src.NetWorks); i++ {
		dest.NetWorks[i] = d.nestNetWorkTonestNetWork(src.NetWorks[i])
	}
	dest.VVMMSS = make([]nest.Vm, 0, len(src.VVMMSS))
	for i := 0; i < len(src.VVMMSS); i++ {
		dest.VVMMSS[i] = d.nestVmTonestVm(src.VVMMSS[i])
	}
	return
}

// parentPath: Vm.VVMMSS.NetWorks:Vm.VVMMSS.NetWorks // path: :
func (d *DTO) nestNetWorkTonestNetWork(src nest.NetWork) (dest nest.NetWork) {
	dest.Mark = src.Mark
	dest.Ns = src.Ns
	return
}

// parentPath: ParentNames:ParentNames // path: :
func (d *DTO) pStringTopString(src *string) (dest *string) {
	dest = src
	return
}

// parentPath: FatherNames:FatherNames // path: FatherNames:FatherNames
func (d *DTO) stringListTostringList(src []string) (dest []string) {
	dest = make([]string, 0, len(src))
	for i := 0; i < len(src); i++ {
		dest[i] = src[i]
	}
	return
}

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
	// LastNames is the last names of the user.
	LastNames    []string `param:"query,lastNames"`
	LastNamesInt []int    `param:"query,lastNamesInt"`
	Paging
	Vm         nest.Vm
	VMS        []nest.Vm
	HeaderName string   `param:"header,headerName"`
	Namespace  []string `param:"query,namespace"`
	VMMap      map[string]nest.Vm
	// @kit-http-param ctx CtxKeyRole
	Role string `param:"ctx,role"`
}

type Paging struct {
	Page int64 `param:"query,page"`
	Size int64 `param:"query,size"`
}

// @tags paas-api
type Service interface {
	// Hello
	// @kit-http /hello/{id} GET
	// @kit-http-request HelloRequest
	// @kit-http-response HelloResponse
	Hello(ctx context.Context, id int, namespace []string, page int64, size int64, lastNames []string) (res string, err error)
	// SayHello
	// @kit-http /hello/say GET
	// @kit-http-request HelloRequest
	// @kit-http-response HelloResponse
	SayHello(ctx context.Context, uuid string, ip string, port int, headerName string) (res string, err error)
}
