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
	"github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
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

type tracing struct {
	next   Service
	tracer opentracing.Tracer
}

func (s *tracing) Hello(ctx context.Context, id int, namespace []string, page int64, size int64, lastNames []string) (res HelloRequest, err error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, s.tracer, "Hello", opentracing.Tag{Key: string(ext.Component), Value: "generator"})
	defer func() {
		span.LogKV("id", id, "namespace", namespace, "page", page, "size", size, "lastNames", lastNames, "err", err)
		span.SetTag(string(ext.Error), err != nil)
		span.Finish()
	}()
	return s.next.Hello(ctx, id, namespace, page, size, lastNames)
}
func (s *tracing) HelloBody(ctx context.Context, helloRequest HelloRequest) (list HelloRequest, total int64, err error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, s.tracer, "HelloBody", opentracing.Tag{Key: string(ext.Component), Value: "generator"})
	defer func() {
		span.LogKV("helloRequest", helloRequest, "err", err)
		span.SetTag(string(ext.Error), err != nil)
		span.Finish()
	}()
	return s.next.HelloBody(ctx, helloRequest)
}
func (s *tracing) SayHello(ctx context.Context, uuid string, ip string, port int, headerName string) (m map[string][]nest.NetWork, err error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, s.tracer, "SayHello", opentracing.Tag{Key: string(ext.Component), Value: "generator"})
	defer func() {
		span.LogKV("uuid", uuid, "ip", ip, "port", port, "headerName", headerName, "err", err)
		span.SetTag(string(ext.Error), err != nil)
		span.Finish()
	}()
	return s.next.SayHello(ctx, uuid, ip, port, headerName)
}

func NewTracing(otTracer opentracing.Tracer) Middleware {
	return func(next Service) Service {
		return &tracing{next: next, tracer: otTracer}
	}
}
