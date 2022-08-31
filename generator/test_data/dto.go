package test_data

import (
	"fmt"
)

type HelloDTO struct{}

func (d *HelloDTO) DTO(src HelloRequest) (dest HelloRequest) {
	dest.Paging.Size = src.Paging.Size
	dest.Vm.Ip = src.Vm.Ip
	dest.Vm.Port = src.Vm.Port
	/*
	   @kit-http-param ctx CtxKeyRole
	   @dto-method RoleToRole
	*/
	dest.Role = RoleToRole(src.Role)
	dest.Paging.Page = src.Paging.Page
	dest.UUID = src.UUID
	dest.Time = src.Time
	dest.Body.Name = src.Body.Name
	dest.Body.Age = src.Body.Age
	/*
	   @dto-method fmt Sprintf
	*/
	dest.HeaderName = fmt.Sprintf(src.HeaderName)
	/*
	   ID is the ID of the user.
	*/
	dest.ID = src.ID
	/*
	   LastNames is the last names of the user.
	*/
	dest.LastNames = src.LastNames
	dest.LastNamesInt = src.LastNamesInt
	dest.Vm.NetWorks = src.Vm.NetWorks
	dest.Vm.VVMMSS = src.Vm.VVMMSS
	dest.VMS = src.VMS
	dest.Namespace = src.Namespace
	dest.ParentNames = src.ParentNames
	dest.SayHi = src.SayHi
	dest.VMMap = src.VMMap
	dest.ParentName = src.ParentName
	dest.FatherNames = src.FatherNames
	return
}

type SayHelloDTO struct{}

func (d *SayHelloDTO) DTO(src HelloRequest) (dest HelloRequest) {
	/*
	   @kit-http-param ctx CtxKeyRole
	   @dto-method RoleToRole
	*/
	dest.Role = RoleToRole(src.Role)
	dest.Body.Name = src.Body.Name
	dest.Body.Age = src.Body.Age
	dest.Paging.Page = src.Paging.Page
	dest.Paging.Size = src.Paging.Size
	dest.Vm.Port = src.Vm.Port
	/*
	   ID is the ID of the user.
	*/
	dest.ID = src.ID
	dest.UUID = src.UUID
	dest.Time = src.Time
	dest.Vm.Ip = src.Vm.Ip
	/*
	   @dto-method fmt Sprintf
	*/
	dest.HeaderName = fmt.Sprintf(src.HeaderName)
	dest.VMS = src.VMS
	dest.Namespace = src.Namespace
	dest.ParentNames = src.ParentNames
	dest.SayHi = src.SayHi
	/*
	   LastNames is the last names of the user.
	*/
	dest.LastNames = src.LastNames
	dest.LastNamesInt = src.LastNamesInt
	dest.Vm.NetWorks = src.Vm.NetWorks
	dest.Vm.VVMMSS = src.Vm.VVMMSS
	dest.VMMap = src.VMMap
	dest.FatherNames = src.FatherNames
	dest.ParentName = src.ParentName
	return
}

type HelloBodyDTO struct{}

func (d *HelloBodyDTO) DTO(src HelloRequest) (dest HelloRequest) {
	dest.Body.Age = src.Body.Age
	dest.Paging.Page = src.Paging.Page
	dest.Paging.Size = src.Paging.Size
	dest.Vm.Ip = src.Vm.Ip
	/*
	   ID is the ID of the user.
	*/
	dest.ID = src.ID
	dest.UUID = src.UUID
	dest.Time = src.Time
	dest.Body.Name = src.Body.Name
	dest.Vm.Port = src.Vm.Port
	/*
	   @dto-method fmt Sprintf
	*/
	dest.HeaderName = fmt.Sprintf(src.HeaderName)
	/*
	   @kit-http-param ctx CtxKeyRole
	   @dto-method RoleToRole
	*/
	dest.Role = RoleToRole(src.Role)
	dest.ParentNames = src.ParentNames
	dest.SayHi = src.SayHi
	/*
	   LastNames is the last names of the user.
	*/
	dest.LastNames = src.LastNames
	dest.LastNamesInt = src.LastNamesInt
	dest.Vm.NetWorks = src.Vm.NetWorks
	dest.Vm.VVMMSS = src.Vm.VVMMSS
	dest.VMS = src.VMS
	dest.Namespace = src.Namespace
	dest.VMMap = src.VMMap
	dest.ParentName = src.ParentName
	dest.FatherNames = src.FatherNames
	return
}
