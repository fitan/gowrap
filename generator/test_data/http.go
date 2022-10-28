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

// @fn wire set
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
@Param id path string true " is the ID of the user."
@Param ip path string true " "
@Param port path string true " "
@Param time path string true " "
@Param uuid path string true " "
@Param lastNames query string false " is the last names of the user."
@Param lastNamesInt query string false " "
@Param namespace query string false " "
@Param page query string false " "
@Param size query string false " "
@Param headerName header string false "@dto-method fmt Sprintf"
@Param user body  true " "
@Success 200 {object} encode.Response{data=HelloRequest}
@Router /hello/{id} [GET]
*/
func decodeHelloRequest(ctx context.Context, r *http1.Request) (res interface{}, err error) {

	req := HelloRequest{}

	var page int64

	var size int64

	var namespace []string

	var lastNames []string

	var lastNamesInt []int

	var id int

	var uuid string

	var time int64

	var ip string

	var port int

	var headerName string

	vars := mux.Vars(r)

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

	id, err = cast.ToIntE(vars["id"])

	if err != nil {
		return
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

	pageStr := r.URL.Query().Get("page")

	if pageStr != "" {
		page, err = cast.ToInt64E(pageStr)
		if err != nil {
			return
		}
	}

	headerName = r.Header.Get("headerName")

	err = json.NewDecoder(r.Body).Decode(&req.Body)

	if err != nil {
		err = errors.Wrap(err, "json.Decode")
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
@Param id path string true " is the ID of the user."
@Param ip path string true " "
@Param port path string true " "
@Param time path string true " "
@Param uuid path string true " "
@Param lastNames query string false " is the last names of the user."
@Param lastNamesInt query string false " "
@Param namespace query string false " "
@Param page query string false " "
@Param size query string false " "
@Param headerName header string false "@dto-method fmt Sprintf"
@Param HelloRequest body HelloRequest true "http request body"
@Success 200 {object} encode.Response{data.list=HelloRequest,data.total=int64}
@Router /hello/body [GET]
*/
func decodeHelloBodyRequest(ctx context.Context, r *http1.Request) (res interface{}, err error) {

	req := HelloRequest{}

	var lastNames []string

	var lastNamesInt []int

	var page int64

	var size int64

	var namespace []string

	var ip string

	var port int

	var id int

	var uuid string

	var time int64

	var headerName string

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

	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		err = errors.Wrap(err, "json.Decode")
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
@Param id path string true " is the ID of the user."
@Param ip path string true " "
@Param port path string true " "
@Param time path string true " "
@Param uuid path string true " "
@Param lastNames query string false " is the last names of the user."
@Param lastNamesInt query string false " "
@Param namespace query string false " "
@Param page query string false " "
@Param size query string false " "
@Param headerName header string false "@dto-method fmt Sprintf"
@Param user body  true " "
@Success 200 {object} encode.Response{data=map[string][]nest.NetWork}
@Router /hello/say [GET]
*/
func decodeSayHelloRequest(ctx context.Context, r *http1.Request) (res interface{}, err error) {

	req := HelloRequest{}

	var lastNames []string

	var lastNamesInt []int

	var page int64

	var size int64

	var namespace []string

	var port int

	var id int

	var uuid string

	var time int64

	var ip string

	var headerName string

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

	err = json.NewDecoder(r.Body).Decode(&req.Body)

	if err != nil {
		err = errors.Wrap(err, "json.Decode")
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
