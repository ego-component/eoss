package awos

import (
	"net/http"

	"github.com/gotomicro/ego/core/elog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type transport struct {
	rt     http.RoundTripper
	before func(r *http.Request)
	after  func(r *http.Request, res *http.Response, err error)
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.before != nil {
		t.before(r)
	}
	res, err := t.rt.RoundTrip(r)
	if t.after != nil {
		t.after(r, res, err)
	}
	return res, err
}

func traceLogReqIdInterceptor(name string, config *config, logger *elog.Component, base http.RoundTripper) *transport {
	t := &transport{rt: base}
	t.after = func(r *http.Request, res *http.Response, err error) {
		span := trace.SpanFromContext(r.Context())
		if !span.SpanContext().IsValid() {
			return
		}
		var reqId string
		switch config.StorageType {
		case StorageTypeS3:
			reqId = res.Header.Get("X-Amz-Request-Id")
		case StorageTypeOSS:
			reqId = res.Header.Get("X-Oss-Request-Id")
		}
		if reqId == "" {
			return
		}
		span.SetAttributes(attribute.String("request-id", reqId))
	}
	return t
}

func logAccessInterceptor(name string, config *config, logger *elog.Component, base http.RoundTripper) *transport {
	t := &transport{rt: base}
	t.after = func(r *http.Request, res *http.Response, err error) {
		span := trace.SpanFromContext(r.Context())
		if !span.SpanContext().IsValid() {
			return
		}
		var reqId string
		switch config.StorageType {
		case StorageTypeS3:
			reqId = res.Header.Get("X-Amz-Request-Id")
		case StorageTypeOSS:
			reqId = res.Header.Get("X-Oss-Request-Id")
		}
		if reqId == "" {
			return
		}
		span.SetAttributes(attribute.String("request-id", reqId))
	}
	return t
}
