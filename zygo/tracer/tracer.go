package tracer

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
	"net/http"
)

func CreateTracer(serviceName string, samplerConfig *config.SamplerConfig, reporter *config.ReporterConfig, options ...config.Option) (opentracing.Tracer, io.Closer, error) {
	var cfg = config.Configuration{
		ServiceName: serviceName,
		Sampler:     samplerConfig, //采样器
		Reporter:    reporter,      //如何上报
	}
	tracer, closer, err := cfg.NewTracer(options...)
	return tracer, closer, err
}

func CreateTracerHeader(serviceName string, header http.Header, samplerConfig *config.SamplerConfig, reporter *config.ReporterConfig, options ...config.Option) (opentracing.Tracer, io.Closer, opentracing.SpanContext, error) {
	var cfg = config.Configuration{
		ServiceName: serviceName,
		Sampler:     samplerConfig, //采样器
		Reporter:    reporter,      //如何上报
	}
	tracer, closer, err := cfg.NewTracer(options...)
	//携带别的进程传递的上文信息的需要解析
	spanContext, _ := tracer.Extract(opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(header))

	return tracer, closer, spanContext, err
}
