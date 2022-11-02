package hello

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport/http"
	"github.com/google/wire"
	"go.uber.org/zap"
)

func NewMws(ms []endpoint.Middleware) (res Mws) {
	res = Mws{}
	for _, v := range ms {
		AllMethodAddMws(res, v)
	}
	return
}

func NewService(log *zap.SugaredLogger) Service {
	var ms []Middleware
	ms = append(ms, NewLogging(log), NewTracing())
	svc := New()

	for _, m := range ms {
		svc = m(svc)
	}
	return svc
}

func NewOps(ops []http.ServerOption) (res Ops) {
	res = Ops{}
	for _, v := range ops {
		AllMethodAddOps(res, v)
	}
	return
}

var HelloSet = wire.NewSet(
	MakeHTTPHandler,
	NewService,
	NewOps,
	NewMws,
)
