package test_data

import (
	"context"
	"encoding/json"
	"fmt"
	govalidator "github.com/asaskevich/govalidator"
	encode "github.com/fitan/gowrap/generator/test_data/encode"
	"github.com/fitan/gowrap/generator/test_data/nest"
	endpoint "github.com/go-kit/kit/endpoint"
	log "github.com/go-kit/kit/log"
	level "github.com/go-kit/kit/log/level"
	http "github.com/go-kit/kit/transport/http"
	mux "github.com/gorilla/mux"
	errors "github.com/pkg/errors"
	cast "github.com/spf13/cast"
	otel "go.opentelemetry.io/otel"
	attribute "go.opentelemetry.io/otel/attribute"
	codes "go.opentelemetry.io/otel/codes"
	trace "go.opentelemetry.io/otel/trace"
	gorm "gorm.io/gorm"
	http1 "net/http"
	"strings"
	"time"
)

type Handler struct{}

func MakeHTTPHandler(r *mux.Router, s Service, mws Mws, ops Ops) Handler {
	eps := NewEndpoint(s, mws)
	r.Handle("/hello/{id}", http.NewServer(eps.HelloEndpoint, decodeHelloRequest, http.EncodeJSONResponse, ops[HelloMethodName]...)).Methods("GET").Name("Hello")
	r.Handle("/hello/body", http.NewServer(eps.HelloBodyEndpoint, decodeHelloBodyRequest, http.EncodeJSONResponse, ops[HelloBodyMethodName]...)).Methods("GET").Name("HelloBody")
	r.Handle("/hello/say", http.NewServer(eps.SayHelloEndpoint, decodeSayHelloRequest, http.EncodeJSONResponse, ops[SayHelloMethodName]...)).Methods("GET").Name("SayHello")

	return Handler{}
}

type Ops map[string][]http.ServerOption

func MethodAddOps(options map[string][]http.ServerOption, option http.ServerOption, methods []string) {
	for _, v := range methods {
		options[v] = append(options[v], option)
	}
}

/*


Hello
@Summary Hello
@Description Hello

@Accept json
@Produce json
@Param id path string true is
@Param ip path string true
@Param port path string true
@Param time path string true
@Param uuid path string true
@Param lastNames query string false is
@Param lastNamesInt query string false
@Param namespace query string false
@Param page query string false
@Param size query string false
@Param headerName header string false "@dto-method fmt Sprintf"
@Param user body struct{Name string "json:\"name\""; Age string "json:\"age\""} true
@Success 200 {object} encode.Response{data=HelloRequest}
@Router /hello/{id} [GET]
*/
func decodeHelloRequest(ctx context.Context, r *http1.Request) (res interface{}, err error) {

	req := HelloRequest{}

	var _lastNames []string

	var _lastNamesInt []int

	var _page int64

	var _size int64

	var _namespace []string

	var _port int

	var _id int

	var _uuid string

	var _time int64

	var _ip string

	var _headerName string

	vars := mux.Vars(r)

	_id, err = cast.ToIntE(vars["id"])

	if err != nil {
		return
	}

	_uuid = vars["uuid"]

	_time, err = cast.ToInt64E(vars["time"])

	if err != nil {
		return
	}

	_ip = vars["ip"]

	_port, err = cast.ToIntE(vars["port"])

	if err != nil {
		return
	}

	pageStr := r.URL.Query().Get("page")

	if pageStr != "" {
		_page, err = cast.ToInt64E(pageStr)
		if err != nil {
			return
		}
	}

	sizeStr := r.URL.Query().Get("size")

	if sizeStr != "" {
		_size, err = cast.ToInt64E(sizeStr)
		if err != nil {
			return
		}
	}

	_namespace = strings.Split(r.URL.Query().Get("namespace"), ",")

	_lastNames = strings.Split(r.URL.Query().Get("lastNames"), ",")

	lastNamesIntStr := r.URL.Query().Get("lastNamesInt")

	if lastNamesIntStr != "" {
		_lastNamesInt, err = cast.ToIntSliceE(strings.Split(lastNamesIntStr, ","))
		if err != nil {
			return
		}
	}

	_headerName = r.Header.Get("headerName")

	err = json.NewDecoder(r.Body).Decode(&req.Body)

	if err != nil {
		err = errors.Wrap(err, "json.Decode")
		return
	}

	req.Time = _time

	req.Vm.Ip = _ip

	req.Vm.Port = _port

	req.ID = _id

	req.UUID = _uuid

	req.LastNames = _lastNames

	req.LastNamesInt = _lastNamesInt

	req.Paging.Page = _page

	req.Paging.Size = _size

	req.Namespace = _namespace

	req.HeaderName = _headerName

	validReq, err := govalidator.ValidateStruct(req)

	if err != nil {
		err = errors.Wrap(err, "govalidator.ValidateStruct")
		return
	}

	if !validReq {
		err = errors.Wrap(err, "valid false")
		return
	}

	return req, err
}

/*


HelloBody
@Summary HelloBody
@Description HelloBody

@Accept json
@Produce json
@Param id path string true is
@Param ip path string true
@Param port path string true
@Param time path string true
@Param uuid path string true
@Param lastNames query string false is
@Param lastNamesInt query string false
@Param namespace query string false
@Param page query string false
@Param size query string false
@Param headerName header string false "@dto-method fmt Sprintf"
@Param HelloRequest body HelloRequest true "http request body"
@Success 200 {object} encode.Response{data.list=HelloRequest,data.total=int64}
@Router /hello/body [GET]
*/
func decodeHelloBodyRequest(ctx context.Context, r *http1.Request) (res interface{}, err error) {

	req := HelloRequest{}

	var _size int64

	var _namespace []string

	var _lastNames []string

	var _lastNamesInt []int

	var _page int64

	var _id int

	var _uuid string

	var _time int64

	var _ip string

	var _port int

	var _headerName string

	vars := mux.Vars(r)

	_ip = vars["ip"]

	_port, err = cast.ToIntE(vars["port"])

	if err != nil {
		return
	}

	_id, err = cast.ToIntE(vars["id"])

	if err != nil {
		return
	}

	_uuid = vars["uuid"]

	_time, err = cast.ToInt64E(vars["time"])

	if err != nil {
		return
	}

	sizeStr := r.URL.Query().Get("size")

	if sizeStr != "" {
		_size, err = cast.ToInt64E(sizeStr)
		if err != nil {
			return
		}
	}

	_namespace = strings.Split(r.URL.Query().Get("namespace"), ",")

	_lastNames = strings.Split(r.URL.Query().Get("lastNames"), ",")

	lastNamesIntStr := r.URL.Query().Get("lastNamesInt")

	if lastNamesIntStr != "" {
		_lastNamesInt, err = cast.ToIntSliceE(strings.Split(lastNamesIntStr, ","))
		if err != nil {
			return
		}
	}

	pageStr := r.URL.Query().Get("page")

	if pageStr != "" {
		_page, err = cast.ToInt64E(pageStr)
		if err != nil {
			return
		}
	}

	_headerName = r.Header.Get("headerName")

	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		err = errors.Wrap(err, "json.Decode")
		return
	}

	req.ID = _id

	req.UUID = _uuid

	req.Time = _time

	req.Vm.Ip = _ip

	req.Vm.Port = _port

	req.Paging.Page = _page

	req.Paging.Size = _size

	req.Namespace = _namespace

	req.LastNames = _lastNames

	req.LastNamesInt = _lastNamesInt

	req.HeaderName = _headerName

	validReq, err := govalidator.ValidateStruct(req)

	if err != nil {
		err = errors.Wrap(err, "govalidator.ValidateStruct")
		return
	}

	if !validReq {
		err = errors.Wrap(err, "valid false")
		return
	}

	return req, err
}

/*


SayHello
@Summary SayHello
@Description SayHello

@Accept json
@Produce json
@Param id path string true is
@Param ip path string true
@Param port path string true
@Param time path string true
@Param uuid path string true
@Param lastNames query string false is
@Param lastNamesInt query string false
@Param namespace query string false
@Param page query string false
@Param size query string false
@Param headerName header string false "@dto-method fmt Sprintf"
@Param user body struct{Name string "json:\"name\""; Age string "json:\"age\""} true
@Success 200 {object} encode.Response{data=map[string][]nest.NetWork}
@Router /hello/say [GET]
*/
func decodeSayHelloRequest(ctx context.Context, r *http1.Request) (res interface{}, err error) {

	req := HelloRequest{}

	var _lastNamesInt []int

	var _page int64

	var _size int64

	var _namespace []string

	var _lastNames []string

	var _time int64

	var _ip string

	var _port int

	var _id int

	var _uuid string

	var _headerName string

	vars := mux.Vars(r)

	_uuid = vars["uuid"]

	_time, err = cast.ToInt64E(vars["time"])

	if err != nil {
		return
	}

	_ip = vars["ip"]

	_port, err = cast.ToIntE(vars["port"])

	if err != nil {
		return
	}

	_id, err = cast.ToIntE(vars["id"])

	if err != nil {
		return
	}

	_lastNames = strings.Split(r.URL.Query().Get("lastNames"), ",")

	lastNamesIntStr := r.URL.Query().Get("lastNamesInt")

	if lastNamesIntStr != "" {
		_lastNamesInt, err = cast.ToIntSliceE(strings.Split(lastNamesIntStr, ","))
		if err != nil {
			return
		}
	}

	pageStr := r.URL.Query().Get("page")

	if pageStr != "" {
		_page, err = cast.ToInt64E(pageStr)
		if err != nil {
			return
		}
	}

	sizeStr := r.URL.Query().Get("size")

	if sizeStr != "" {
		_size, err = cast.ToInt64E(sizeStr)
		if err != nil {
			return
		}
	}

	_namespace = strings.Split(r.URL.Query().Get("namespace"), ",")

	_headerName = r.Header.Get("headerName")

	err = json.NewDecoder(r.Body).Decode(&req.Body)

	if err != nil {
		err = errors.Wrap(err, "json.Decode")
		return
	}

	req.ID = _id

	req.UUID = _uuid

	req.Time = _time

	req.Vm.Ip = _ip

	req.Vm.Port = _port

	req.Paging.Size = _size

	req.Namespace = _namespace

	req.LastNames = _lastNames

	req.LastNamesInt = _lastNamesInt

	req.Paging.Page = _page

	req.HeaderName = _headerName

	validReq, err := govalidator.ValidateStruct(req)

	if err != nil {
		err = errors.Wrap(err, "govalidator.ValidateStruct")
		return
	}

	if !validReq {
		err = errors.Wrap(err, "valid false")
		return
	}

	return req, err
}
