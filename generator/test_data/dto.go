package test_data

import (
	"fmt"
)

type HelloDTO struct{}

func (d *HelloDTO) DTO(src HelloRequest) (dest HelloRequest) {
	dest.Body.Age = src.Body.Age
	dest.Paging.Size = src.Paging.Size
	dest.Vm.Port = src.Vm.Port
	dest.Time = src.Time
	dest.Body.Name = src.Body.Name
	dest.Paging.Page = src.Paging.Page
	dest.Vm.Ip = src.Vm.Ip
	/*
	   @dto-method fmt Sprintf
	*/
	dest.HeaderName = fmt.Sprintf(src.HeaderName)
	/*
	   @kit-http-param ctx CtxKeyRole
	   @dto-method RoleToRole
	*/
	dest.Role = RoleToRole(src.Role)
	/*
	   ID is the ID of the user.
	*/
	dest.ID = src.ID
	dest.UUID = src.UUID
	dest.LastNamesInt = src.LastNamesInt
	dest.Vm.NetWorks = src.Vm.NetWorks
	dest.Vm.VVMMSS = src.Vm.VVMMSS
	dest.VMS = src.VMS
	dest.Namespace = src.Namespace
	dest.ParentNames = src.ParentNames
	dest.SayHi = src.SayHi
	/*
	   LastNames is the last names of the user.
	*/
	dest.LastNames = src.LastNames
	dest.VMMap = src.VMMap
	dest.ParentName = src.ParentName
	dest.FatherNames = src.FatherNames
	return
}

type SayHelloDTO struct{}

func (d *SayHelloDTO) DTO(src HelloRequest) (dest HelloRequest) {
	dest.UUID = src.UUID
	dest.Vm.Ip = src.Vm.Ip
	/*
	   @dto-method fmt Sprintf
	*/
	dest.HeaderName = fmt.Sprintf(src.HeaderName)
	/*
	   ID is the ID of the user.
	*/
	dest.ID = src.ID
	dest.Body.Name = src.Body.Name
	dest.Body.Age = src.Body.Age
	dest.Paging.Page = src.Paging.Page
	dest.Paging.Size = src.Paging.Size
	dest.Vm.Port = src.Vm.Port
	/*
	   @kit-http-param ctx CtxKeyRole
	   @dto-method RoleToRole
	*/
	dest.Role = RoleToRole(src.Role)
	dest.Time = src.Time
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

type HelloBodyDTO struct{}

func (d *HelloBodyDTO) DTO(src HelloRequest) (dest HelloRequest) {
	dest.Body.Age = src.Body.Age
	dest.Paging.Size = src.Paging.Size
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
	/*
	   ID is the ID of the user.
	*/
	dest.ID = src.ID
	dest.Paging.Page = src.Paging.Page
	dest.Vm.Ip = src.Vm.Ip
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
	dest.ParentNames = src.ParentNames
	dest.VMMap = src.VMMap
	dest.ParentName = src.ParentName
	dest.FatherNames = src.FatherNames
	return
}
