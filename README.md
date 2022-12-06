# GoWrap
[![License](https://img.shields.io/badge/license-mit-green.svg)](https://github.com/fitan/gowrap/blob/master/LICENSE)
[![Build](https://github.com/fitan/gowrap/actions/workflows/go.yml/badge.svg)](https://github.com/fitan/gowrap/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/hexdigest/gowrap/badge.svg?branch=master)](https://coveralls.io/github/hexdigest/gowrap?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/fitan/gowrap?dropcache)](https://goreportcard.com/report/github.com/fitan/gowrap)
[![GoDoc](https://godoc.org/github.com/fitan/gowrap?status.svg)](http://godoc.org/github.com/fitan/gowrap)
[![Awesome](https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg)](https://github.com/avelino/awesome-go#generation-and-generics)
[![Release](https://img.shields.io/github/release/hexdigest/gowrap.svg)](https://github.com/fitan/gowrap/releases/latest)

GoWrap is a command line tool that generates decorators for Go interface types using simple templates.
With GoWrap you can easily add metrics, tracing, fallbacks, pools, and many other features into your existing code in a few seconds.


## Demo

![demo](https://github.com/fitan/gowrap/blob/master/gowrap.gif)

## 安装

```
go get -u github.com/fitan/gowrap/cmd/gk
```

## 生成Kit Http模板

```shell
gk gen -g "myHttp:http.go" -g "myEndpoint:endpoint.go" -g "log:log.go" -g "trace:trace.go" -g "dto:dto.go"
```

## 初始化模板

### 创建crud service
```shell
gk init -t sc -n user -o ./user -p true
```

### 创建crud repo
```shell
gk init -t rc -n user -o ./user -p true
```

### 创建空 service
```shell
gk init -t s -n user -o ./user -p true
```

### 创建空 repo
```shell
gk init -t r -n user -o ./user -p true
```

### 注释语法
```go
// @tags hello
// base-path user
// @impl
type Service interface {
	// 设置 http path 和 method
	// @kit-http /list GET
	// kithttp request 对象名字
	// @kit-http-request ListRequest
	// 使用自定义的endpoint
	// @kit-http-endpoint xxx
	// 使用自定义的decode
	// @kit-http-decode xxx
	// 使用自定义的encode
	// @kit-http-encode xxx
	// 使用自定义的wrap
	// @kit-http-endpoint-wrap xxx
	// 是否生成swagger
	// @swag true
	List(ctx context.Context, page, limit int) (list []ListResponse, total int64, err error)
}

func (s *service) List(ctx context.Context, page, limit int) (list []ListResponse, total int64, err error) {
    res, total, err := s.repo.User.List(ctx, page, limit, "", nil, scope)
    if err != nil {
        err = errors.Wrap(err, "repo.Pm.List")
        return
    }
    // 自动生成对象转换
    // @call copy
    list = listDTO(res)
    return
}

type ListRequest struct {
    Page  int `json:"page" param:"query,page"`
    Limit int `json:"limit" param:"query,limit"` 
    // @kit-http-param contextKey.RoleID
    RoleID int `json:"roleID" param:"ctx,roleID"`
}
```

### kit-http-request和func的param的映射关系
```text
param:"type,page" query是http请求的类型，page是func的参数名
query的类型: path,query,body,header,ctx
```