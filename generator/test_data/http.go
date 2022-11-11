package test_data

import (
	"context"
	"encoding/json"
	http1 "net/http"
	"strings"

	govalidator "github.com/asaskevich/govalidator"
	endpoint "github.com/go-kit/kit/endpoint"
	http "github.com/go-kit/kit/transport/http"
	mux "github.com/gorilla/mux"
	errors "github.com/pkg/errors"
	cast "github.com/spf13/cast"
)

func MakeHTTPHandler(s Service, dmw []endpoint.Middleware, opts []http.ServerOption) http1.Handler {
	var ems []endpoint.Middleware
	opts = append(opts, http.ServerBefore(func(ctx context.Context, request *http1.Request) context.Context {
		return ctx
	}))
	ems = append(ems, dmw...)
	eps := NewEndpoint(s, map[string][]endpoint.Middleware{HelloMethodName: ems, HelloBodyMethodName: ems, SayHelloMethodName: ems})
	r := mux.NewRouter()
	r.Handle("/hello/{id}", http.NewServer(eps.HelloEndpoint, decodeHelloRequest, http.EncodeJSONResponse, opts...)).Methods("GET").Name("Hello")
	r.Handle("/hello/body", http.NewServer(eps.HelloBodyEndpoint, decodeHelloBodyRequest, http.EncodeJSONResponse, opts...)).Methods("GET").Name("HelloBody")
	r.Handle("/hello/say", http.NewServer(eps.SayHelloEndpoint, decodeSayHelloRequest, http.EncodeJSONResponse, opts...)).Methods("GET").Name("SayHello")

	return r
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

	var _page int64

	var _size int64

	var _namespace []string

	var _lastNames []string

	var _lastNamesInt []int

	var _id int

	var _uuid string

	var _time int64

	var _ip string

	var _port int

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

	req.UUID = _uuid

	req.Time = _time

	req.Vm.Ip = _ip

	req.Vm.Port = _port

	req.ID = _id

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

	var _lastNames []string

	var _lastNamesInt []int

	var _page int64

	var _size int64

	var _namespace []string

	var _id int

	var _uuid string

	var _time int64

	var _ip string

	var _port int

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

	sizeStr := r.URL.Query().Get("size")

	if sizeStr != "" {
		_size, err = cast.ToInt64E(sizeStr)
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

	var _namespace []string

	var _lastNames []string

	var _lastNamesInt []int

	var _page int64

	var _size int64

	var _ip string

	var _port int

	var _id int

	var _uuid string

	var _time int64

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

	req.UUID = _uuid

	req.Time = _time

	req.Vm.Ip = _ip

	req.Vm.Port = _port

	req.ID = _id

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
