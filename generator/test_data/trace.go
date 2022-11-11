package test_data

import (
	"context"
	"encoding/json"

	"github.com/fitan/gowrap/generator/test_data/nest"
	otel "go.opentelemetry.io/otel"
	attribute "go.opentelemetry.io/otel/attribute"
	codes "go.opentelemetry.io/otel/codes"
	trace "go.opentelemetry.io/otel/trace"
)

type tracing struct {
	next Service
}

func (s *tracing) Hello(ctx context.Context, id int, namespace []string, page int64, size int64, lastNames []string) (res HelloRequest, err error) {
	_, span := otel.Tracer("generator").Start(ctx, "Hello")

	defer func() {
		paramIn := map[string]interface{}{
			"id":        id,
			"lastNames": lastNames,
			"namespace": namespace,
			"page":      page,
			"size":      size,
		}
		paramInJsonB, _ := json.Marshal(paramIn)
		span.AddEvent("paramIn", trace.WithAttributes(attribute.String("param list", string(paramInJsonB))))
		if err != nil {
			span.SetStatus(codes.Error, "Hello error")
			span.RecordError(err)
		}
		span.End()
	}()

	return s.next.Hello(ctx, id, namespace, page, size, lastNames)
}
func (s *tracing) HelloBody(ctx context.Context, helloRequest HelloRequest) (list HelloRequest, total int64, err error) {
	_, span := otel.Tracer("generator").Start(ctx, "HelloBody")

	defer func() {
		paramIn := map[string]interface{}{"helloRequest": helloRequest}
		paramInJsonB, _ := json.Marshal(paramIn)
		span.AddEvent("paramIn", trace.WithAttributes(attribute.String("param list", string(paramInJsonB))))
		if err != nil {
			span.SetStatus(codes.Error, "HelloBody error")
			span.RecordError(err)
		}
		span.End()
	}()

	return s.next.HelloBody(ctx, helloRequest)
}
func (s *tracing) SayHello(ctx context.Context, uuid string, ip string, port int, headerName string) (m map[string][]nest.NetWork, err error) {
	_, span := otel.Tracer("generator").Start(ctx, "SayHello")

	defer func() {
		paramIn := map[string]interface{}{
			"headerName": headerName,
			"ip":         ip,
			"port":       port,
			"uuid":       uuid,
		}
		paramInJsonB, _ := json.Marshal(paramIn)
		span.AddEvent("paramIn", trace.WithAttributes(attribute.String("param list", string(paramInJsonB))))
		if err != nil {
			span.SetStatus(codes.Error, "SayHello error")
			span.RecordError(err)
		}
		span.End()
	}()

	return s.next.SayHello(ctx, uuid, ip, port, headerName)
}
func NewTracing() Middleware {
	return func(next Service) Service {
		return &tracing{next: next}
	}
}
