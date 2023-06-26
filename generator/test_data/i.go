package test_data

import (
	"context"
	"gorm.io/gorm/schema"
	"reflect"

	// @extra "gitlab.creditease.corp/paas/paas-assets/src/middleware"
	// @extra "github.com/fitan/gowrap/generator/test_data/encode"
	"fmt"
	"github.com/fitan/gowrap/generator/test_data/nest"
)

// @enum(say:中文, hello:英文, hi:日文, say:"fdsfsdf")
type Enum int

func (e *Enum) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	switch value := dbValue.(type) {
	case string:
		*e = _EnumValue[value]
	default:
		return fmt.Errorf("unknown enum value: %v", value)
	}

	return nil
}

func (e Enum) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	return e.String(), nil
}

const (
	_ = iota
	Say
	Hello
	Hi
)

var _EnumValue = map[string]Enum{
	"say":   Say,
	"hello": Hello,
	"hi":    Hi,
}

func ParseEnum(name string) (Enum, error) {
	if x, ok := _EnumValue[name]; ok {
		return x, nil
	}
	return 0, fmt.Errorf("unknown enum value: %s", name)
}

func (e Enum) String() string {
	switch e {
	case 1:
		return "say"
	case 2:
		return "hello"
	case 3:
		return "hi"
	}

	return fmt.Sprintf("Enum(%d)", e)
}

func (e *Enum) UnmarshalJSON(data []byte) error {
	v, err := ParseEnum(string(data))
	if err != nil {
		return err
	}
	*e = v
	return nil
}

func (e *Enum) MarshalJSON() ([]byte, error) {
	switch *e {
	case Say:
		return []byte(`say`), nil
	case Hello:
		return []byte(`hello`), nil
	case Hi:
		return []byte(`hi`), nil
	default:
		return nil, fmt.Errorf("unknown enum value: %d", e)
	}
}

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
	// @kit-http-param ctx CtxKeyRole
	// @dto-method RoleToRole
	Role string `param:"ctx,role"`
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

// @tags paas-api
// @impl kit
type Service interface {
	// Hello
	// @kit-http /hello/{id} GET
	// @kit-http-request HelloRequest
	// @kit-http-response HelloRequest
	Hello(ctx context.Context, id int, namespace []string, page int64, size int64, lastNames []string) (res HelloRequest, err error)
	// SayHello
	// @kit-http /hello/say GET
	// @kit-http-request HelloRequest
	// @kit-http-response HelloRequest
	SayHello(ctx context.Context, uuid string, ip string, port int, headerName string) (res HelloRequest, err error)

	// HelloBody @kit-http /hello/body GET
	// @kit-http-request HelloRequest body
	// @kit-http-response HelloRequest
	// @kit-cache get helloRequest method GetRedisKey 1s
	// @kit-cache delete helloRequest key 1s
	// @kit-cache put helloRequest Interface 1s
	HelloBody(ctx context.Context, helloRequest HelloRequest) (list HelloRequest, total int64, err error)
	nest.Nest
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
	src := HelloRequest{}
	// @fn dto
	apiDTO(src, &res)
	panic("implement me")
}

func (r RedisService) HelloBody(ctx context.Context, helloRequest HelloRequest) (
	list HelloRequest, total int64, err error,
) {
	panic("implement me")
}
