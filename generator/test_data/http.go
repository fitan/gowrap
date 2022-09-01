package test_data

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	valid "github.com/asaskevich/govalidator"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

func MakeHTTPHandler(s Service, dmw []endpoint.Middleware, opts []kithttp.ServerOption) http.Handler {
	var ems []endpoint.Middleware

	opts = append(opts, kithttp.ServerBefore(func(ctx context.Context, request *http.Request) context.Context {
		return ctx
	}))

	ems = append(ems, dmw...)

	eps := NewEndpoint(s, map[string][]endpoint.Middleware{

		"Hello": ems,

		"HelloBody": ems,

		"SayHello": ems,
	})

	r := mux.NewRouter()

	r.Handle("/hello/{id}", kithttp.NewServer(
		eps.HelloEndpoint,
		decodeHelloRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)).Methods("GET")

	r.Handle("", kithttp.NewServer(
		eps.HelloBodyEndpoint,
		decodeHelloBodyRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)).Methods("")

	r.Handle("/hello/say", kithttp.NewServer(
		eps.SayHelloEndpoint,
		decodeSayHelloRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)).Methods("GET")

	return r
}

// Hello
// @Summary
// @Description
// @tags paas-api
// @Accept json
// @Produce json
// @Param id path int true " is the ID of the user."
// @Param ip path string true " "
// @Param port path int true " "
// @Param time path int64 true " "
// @Param uuid path string true " "
// @Param lastNames query []string false " is the last names of the user."
// @Param lastNamesInt query []int false " "
// @Param namespace query []string false " "
// @Param page query int64 false " "
// @Param size query int64 false " "
// @Param headerName header string false "@dto-method fmt Sprintf"
// @Param user body  true " "
// @Success 200 {object} encode.Response{data=HelloRequest}
// @Router /hello/{id} [GET]
func decodeHelloRequest(ctx context.Context, r *http.Request) (res interface{}, err error) {

	req := HelloRequest{}

	var namespace []string

	var lastNames []string

	var lastNamesInt []int

	var page int64

	var size int64

	var id int

	var uuid string

	var time int64

	var ip string

	var port int

	var headerName string

	var role string

	vars := mux.Vars(r)

	ip = vars["ip"]

	port, err = cast.ToIntE(vars["port"])

	if err != nil {
		return
	}

	id, err = cast.ToIntE(vars["id"])

	if err != nil {
		return
	}

	uuid = vars["uuid"]

	time, err = cast.ToInt64E(vars["time"])

	if err != nil {
		return
	}

	pageStr := r.URL.Query().Get("page")

	if pageStr != "" {
		page, err = cast.ToInt64E(pageStr)
		if err != nil {
			return
		}
	}

	sizeStr := r.URL.Query().Get("size")

	if sizeStr != "" {
		size, err = cast.ToInt64E(sizeStr)
		if err != nil {
			return
		}
	}

	namespace = strings.Split(r.URL.Query().Get("namespace"), ",")

	lastNames = strings.Split(r.URL.Query().Get("lastNames"), ",")

	lastNamesIntStr := r.URL.Query().Get("lastNamesInt")

	if lastNamesIntStr != "" {
		lastNamesInt, err = cast.ToIntSliceE(strings.Split(lastNamesIntStr, ","))
		if err != nil {
			return
		}
	}

	headerName = r.Header.Get("headerName")

	err = json.NewDecoder(r.Body).Decode(&req.Body)

	if err != nil {
		return
	}

	var roleOK bool

	role, roleOK = ctx.Value(CtxKeyRole).(string)

	if roleOK == false {
		err = errors.New("ctx param role is not found")
		return
	}

	req.ID = id

	req.UUID = uuid

	req.Time = time

	req.Vm.Ip = ip

	req.Vm.Port = port

	req.Paging.Size = size

	req.Namespace = namespace

	req.LastNames = lastNames

	req.LastNamesInt = lastNamesInt

	req.Paging.Page = page

	req.HeaderName = headerName

	req.Role = role

	validReq, err := valid.ValidateStruct(res)

	if err != nil {
		err = errors.Wrap(err, "valid.ValidateStruct")
		return
	}

	if !validReq {
		err = errors.Wrap(err, "valid false")
		return
	}

	return req, err
}

// HelloBody
// @Summary  @kit-http /hello/body GET
// @Description  @kit-http /hello/body GET
// @tags paas-api
// @Accept json
// @Produce json
// @Param id path int true " is the ID of the user."
// @Param ip path string true " "
// @Param port path int true " "
// @Param time path int64 true " "
// @Param uuid path string true " "
// @Param lastNames query []string false " is the last names of the user."
// @Param lastNamesInt query []int false " "
// @Param namespace query []string false " "
// @Param page query int64 false " "
// @Param size query int64 false " "
// @Param headerName header string false "@dto-method fmt Sprintf"
// @Param HelloRequest body HelloRequest 	true "http request body"
// @Success 200 {object} encode.Response{data=HelloRequest}
// @Router  []
func decodeHelloBodyRequest(ctx context.Context, r *http.Request) (res interface{}, err error) {

	req := HelloRequest{}

	var size int64

	var namespace []string

	var lastNames []string

	var lastNamesInt []int

	var page int64

	var id int

	var uuid string

	var time int64

	var ip string

	var port int

	var headerName string

	var role string

	vars := mux.Vars(r)

	id, err = cast.ToIntE(vars["id"])

	if err != nil {
		return
	}

	uuid = vars["uuid"]

	time, err = cast.ToInt64E(vars["time"])

	if err != nil {
		return
	}

	ip = vars["ip"]

	port, err = cast.ToIntE(vars["port"])

	if err != nil {
		return
	}

	lastNamesIntStr := r.URL.Query().Get("lastNamesInt")

	if lastNamesIntStr != "" {
		lastNamesInt, err = cast.ToIntSliceE(strings.Split(lastNamesIntStr, ","))
		if err != nil {
			return
		}
	}

	pageStr := r.URL.Query().Get("page")

	if pageStr != "" {
		page, err = cast.ToInt64E(pageStr)
		if err != nil {
			return
		}
	}

	sizeStr := r.URL.Query().Get("size")

	if sizeStr != "" {
		size, err = cast.ToInt64E(sizeStr)
		if err != nil {
			return
		}
	}

	namespace = strings.Split(r.URL.Query().Get("namespace"), ",")

	lastNames = strings.Split(r.URL.Query().Get("lastNames"), ",")

	headerName = r.Header.Get("headerName")

	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		return
	}

	var roleOK bool

	role, roleOK = ctx.Value(CtxKeyRole).(string)

	if roleOK == false {
		err = errors.New("ctx param role is not found")
		return
	}

	req.ID = id

	req.UUID = uuid

	req.Time = time

	req.Vm.Ip = ip

	req.Vm.Port = port

	req.LastNames = lastNames

	req.LastNamesInt = lastNamesInt

	req.Paging.Page = page

	req.Paging.Size = size

	req.Namespace = namespace

	req.HeaderName = headerName

	req.Role = role

	validReq, err := valid.ValidateStruct(res)

	if err != nil {
		err = errors.Wrap(err, "valid.ValidateStruct")
		return
	}

	if !validReq {
		err = errors.Wrap(err, "valid false")
		return
	}

	return req, err
}

// SayHello
// @Summary
// @Description
// @tags paas-api
// @Accept json
// @Produce json
// @Param id path int true " is the ID of the user."
// @Param ip path string true " "
// @Param port path int true " "
// @Param time path int64 true " "
// @Param uuid path string true " "
// @Param lastNames query []string false " is the last names of the user."
// @Param lastNamesInt query []int false " "
// @Param namespace query []string false " "
// @Param page query int64 false " "
// @Param size query int64 false " "
// @Param headerName header string false "@dto-method fmt Sprintf"
// @Param user body  true " "
// @Success 200 {object} encode.Response{data=HelloRequest}
// @Router /hello/say [GET]
func decodeSayHelloRequest(ctx context.Context, r *http.Request) (res interface{}, err error) {

	req := HelloRequest{}

	var lastNames []string

	var lastNamesInt []int

	var page int64

	var size int64

	var namespace []string

	var id int

	var uuid string

	var time int64

	var ip string

	var port int

	var headerName string

	var role string

	vars := mux.Vars(r)

	time, err = cast.ToInt64E(vars["time"])

	if err != nil {
		return
	}

	ip = vars["ip"]

	port, err = cast.ToIntE(vars["port"])

	if err != nil {
		return
	}

	id, err = cast.ToIntE(vars["id"])

	if err != nil {
		return
	}

	uuid = vars["uuid"]

	namespace = strings.Split(r.URL.Query().Get("namespace"), ",")

	lastNames = strings.Split(r.URL.Query().Get("lastNames"), ",")

	lastNamesIntStr := r.URL.Query().Get("lastNamesInt")

	if lastNamesIntStr != "" {
		lastNamesInt, err = cast.ToIntSliceE(strings.Split(lastNamesIntStr, ","))
		if err != nil {
			return
		}
	}

	pageStr := r.URL.Query().Get("page")

	if pageStr != "" {
		page, err = cast.ToInt64E(pageStr)
		if err != nil {
			return
		}
	}

	sizeStr := r.URL.Query().Get("size")

	if sizeStr != "" {
		size, err = cast.ToInt64E(sizeStr)
		if err != nil {
			return
		}
	}

	headerName = r.Header.Get("headerName")

	err = json.NewDecoder(r.Body).Decode(&req.Body)

	if err != nil {
		return
	}

	var roleOK bool

	role, roleOK = ctx.Value(CtxKeyRole).(string)

	if roleOK == false {
		err = errors.New("ctx param role is not found")
		return
	}

	req.ID = id

	req.UUID = uuid

	req.Time = time

	req.Vm.Ip = ip

	req.Vm.Port = port

	req.Paging.Size = size

	req.Namespace = namespace

	req.LastNames = lastNames

	req.LastNamesInt = lastNamesInt

	req.Paging.Page = page

	req.HeaderName = headerName

	req.Role = role

	validReq, err := valid.ValidateStruct(res)

	if err != nil {
		err = errors.Wrap(err, "valid.ValidateStruct")
		return
	}

	if !validReq {
		err = errors.Wrap(err, "valid false")
		return
	}

	return req, err
}
