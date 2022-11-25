package awos

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/emetric"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type transport struct {
	rt          http.RoundTripper
	onReqBefore func(r *http.Request)
	onReqAfter  func(r *http.Request, res *http.Response, err error)
	onEnd       func(r *http.Request, res *http.Response, err error)
}

type wrappedBody struct {
	body  io.ReadCloser
	onEnd func(r *http.Request, res *http.Response, err error)
	req   *http.Request
	res   *http.Response
}

func (wb *wrappedBody) Read(b []byte) (int, error) {
	n, err := wb.body.Read(b)

	switch err {
	case nil:
		// nothing to do here but fall through to the return
	case io.EOF:
		if wb.onEnd != nil {
			wb.onEnd(wb.req, wb.res, nil)
		}
	default:
		if wb.onEnd != nil {
			wb.onEnd(wb.req, wb.res, err)
		}
	}
	return n, err
}

func (wb *wrappedBody) Close() error {
	if wb.onEnd != nil {
		wb.onEnd(wb.req, wb.res, nil)
	}
	if wb.body != nil {
		return wb.body.Close()
	}
	return nil
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.onReqBefore != nil {
		t.onReqBefore(r)
	}
	res, err := t.rt.RoundTrip(r)
	if t.onReqAfter != nil {
		t.onReqAfter(r, res, err)
	}
	if err != nil {
		if t.onEnd != nil {
			t.onEnd(r, res, err)
		}
		return res, err
	}
	res.Body = &wrappedBody{body: res.Body, onEnd: t.onEnd, req: r, res: res}
	return res, err
}

type begKey struct{}

func beg(ctx context.Context) time.Time {
	begTime, _ := ctx.Value(begKey{}).(time.Time)
	return begTime
}

func fixedInterceptor(name string, config *config, logger *elog.Component, base http.RoundTripper) *transport {
	t := &transport{rt: base}
	t.onReqBefore = func(r *http.Request) {
		*r = *(r.WithContext(context.WithValue(r.Context(), begKey{}, time.Now())))
	}
	return t
}

func traceLogReqIdInterceptor(name string, config *config, logger *elog.Component, base http.RoundTripper) *transport {
	t := &transport{rt: base}
	t.onReqAfter = func(r *http.Request, res *http.Response, err error) {
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

func metricInterceptor(name string, config *config, logger *elog.Component, base http.RoundTripper) *transport {
	t := &transport{rt: base}
	t.onReqAfter = func(r *http.Request, res *http.Response, err error) {
		code := ""
		if err != nil {
			code = "request error"
		} else {
			code = http.StatusText(res.StatusCode)
		}
		emetric.ClientHandleCounter.Inc("oss", name, r.Method, config.Bucket, code)
	}
	t.onEnd = func(r *http.Request, res *http.Response, err error) {
		emetric.ClientHandleHistogram.Observe(time.Since(beg(r.Context())).Seconds(), "oss", name, r.Method, config.Bucket)
	}
	return t
}
