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

type logging struct {
	logger  log.Logger
	next    Service
	traceId string
}

func (s *logging) Hello(ctx context.Context, id int, namespace []string, page int64, size int64, lastNames []string) (res HelloRequest, err error) {
	defer func(begin time.Time) {
		_ = s.logger.Log(s.traceId, ctx.Value(s.traceId), "method", "Hello", "id", id, "namespace", namespace, "page", page, "size", size, "lastNames", lastNames, "took", time.Since(begin), "err", err)
	}(time.Now())
	return s.next.Hello(ctx, id, namespace, page, size, lastNames)
}
func (s *logging) HelloBody(ctx context.Context, helloRequest HelloRequest) (list HelloRequest, total int64, err error) {
	defer func(begin time.Time) {
		_ = s.logger.Log(s.traceId, ctx.Value(s.traceId), "method", "HelloBody", "helloRequest", helloRequest, "took", time.Since(begin), "err", err)
	}(time.Now())
	return s.next.HelloBody(ctx, helloRequest)
}
func (s *logging) SayHello(ctx context.Context, uuid string, ip string, port int, headerName string) (m map[string][]nest.NetWork, err error) {
	defer func(begin time.Time) {
		_ = s.logger.Log(s.traceId, ctx.Value(s.traceId), "method", "SayHello", "uuid", uuid, "ip", ip, "port", port, "headerName", headerName, "took", time.Since(begin), "err", err)
	}(time.Now())
	return s.next.SayHello(ctx, uuid, ip, port, headerName)
}

func NewLogging(logger log.Logger, traceId string) Middleware {
	logger = log.With(logger, "generator", "logging")
	return func(next Service) Service {
		return &logging{logger: level.Info(logger), next: next, traceId: traceId}
	}
}
