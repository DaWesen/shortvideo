package tracing

import (
	"context"
	"log"
	"shortvideo/pkg/config"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otlptracehttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	trace "go.opentelemetry.io/otel/trace"
)

var (
	tracingInstance *TracingManager
	tracingOnce     sync.Once
)

// 管理分布式链路追踪
type TracingManager struct {
	config        *config.TracingConfig
	tracer        trace.Tracer
	traceProvider *sdktrace.TracerProvider
}

// 创建Tracing管理器
func NewTracingManager() (*TracingManager, error) {
	var err error
	tracingOnce.Do(func() {
		cfg := config.Get().Tracing

		if !cfg.Enable {
			//如果未启用Tracing，则使用NoopTracerProvider
			noopProvider := sdktrace.NewTracerProvider()
			tracingInstance = &TracingManager{
				config:        &cfg,
				tracer:        noopProvider.Tracer("shortvideo"),
				traceProvider: noopProvider,
			}
			return
		}

		//创建OTLP HTTP exporter
		exporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpoint(cfg.JaegerEndpoint))
		if err != nil {
			log.Printf("创建OTLP exporter失败: %v，将使用NoopTracerProvider", err)
			noopProvider := sdktrace.NewTracerProvider()
			tracingInstance = &TracingManager{
				config:        &cfg,
				tracer:        noopProvider.Tracer("shortvideo"),
				traceProvider: noopProvider,
			}
			return
		}

		//创建资源
		res := resource.NewWithAttributes(
			"", //使用默认schema
			attribute.String("service.name", "shortvideo"),
			attribute.String("service.version", config.Get().App.Version),
			attribute.String("environment", config.Get().App.Env),
		)

		//创建采样器
		sampler := sdktrace.TraceIDRatioBased(cfg.SampleRate)

		//创建TraceProvider
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sampler),
		)

		//设置全局TracerProvider
		otel.SetTracerProvider(tp)

		tracingInstance = &TracingManager{
			config:        &cfg,
			tracer:        tp.Tracer("shortvideo"),
			traceProvider: tp,
		}

		log.Println("分布式链路追踪已启动")
	})

	return tracingInstance, err
}
