
package test_data

import (
	"context"
	"encoding/json"
	"fmt"
	govalidator "github.com/asaskevich/govalidator"
	encode "github.com/fitan/gowrap/generator/test_data/encode"
	"github.com/fitan/gowrap/generator/test_data/nest"
	endpoint "github.com/go-kit/kit/endpoint"
	http "github.com/go-kit/kit/transport/http"
	mux "github.com/gorilla/mux"
	errors "github.com/pkg/errors"
	cast "github.com/spf13/cast"
	otel "go.opentelemetry.io/otel"
	attribute "go.opentelemetry.io/otel/attribute"
	codes "go.opentelemetry.io/otel/codes"
	trace "go.opentelemetry.io/otel/trace"
	zap "go.uber.org/zap"
	http1 "net/http"
	"strings"
	"time"
)

type logging struct {
	logger *zap.SugaredLogger
	next   Service
}

func (s *logging) Hello(ctx context.Context, id int, namespace []string, page int64, size int64, lastNames []string) (res HelloRequest, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Errorw("Hello error", "error", err, "id", id, "namespace", namespace, "page", page, "size", size, "lastNames", lastNames, "took", time.Since(begin), "traceId", trace.SpanContextFromContext(ctx).TraceID().String())
		} else {
			s.logger.Infow("Hello success", "id", id, "namespace", namespace, "page", page, "size", size, "lastNames", lastNames, "took", time.Since(begin), "traceId", trace.SpanContextFromContext(ctx).TraceID().String())
		}
	}(time.Now())
	return s.next.Hello(ctx, id, namespace, page, size, lastNames)
}
func (s *logging) HelloBody(ctx context.Context, helloRequest HelloRequest) (list HelloRequest, total int64, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Errorw("HelloBody error", "error", err, "helloRequest", helloRequest, "took", time.Since(begin), "traceId", trace.SpanContextFromContext(ctx).TraceID().String())
		} else {
			s.logger.Infow("HelloBody success", "helloRequest", helloRequest, "took", time.Since(begin), "traceId", trace.SpanContextFromContext(ctx).TraceID().String())
		}
	}(time.Now())
	return s.next.HelloBody(ctx, helloRequest)
}
func (s *logging) SayHello(ctx context.Context, uuid string, ip string, port int, headerName string) (m map[string][]nest.NetWork, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Errorw("SayHello error", "error", err, "uuid", uuid, "ip", ip, "port", port, "headerName", headerName, "took", time.Since(begin), "traceId", trace.SpanContextFromContext(ctx).TraceID().String())
		} else {
			s.logger.Infow("SayHello success", "uuid", uuid, "ip", ip, "port", port, "headerName", headerName, "took", time.Since(begin), "traceId", trace.SpanContextFromContext(ctx).TraceID().String())
		}
	}(time.Now())
	return s.next.SayHello(ctx, uuid, ip, port, headerName)
}
func NewLogging(logger *zap.SugaredLogger) Middleware {
	return func(next Service) Service {
		return &logging{logger: logger, next: next}
	}
}

