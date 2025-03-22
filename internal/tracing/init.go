package tracing

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

type ExporterType string

const (
	ExporterTypeJaeger   ExporterType = "jaeger"
	ExporterTypeZipkin   ExporterType = "zipkin"
	ExporterTypeOTLPHTTP ExporterType = "otlp-http"
	ExporterTypeOTLPGRPC ExporterType = "otlp-grpc"
)

type Config struct {
	ServiceName        string
	ServiceVersion     string
	ExporterType       ExporterType
	Endpoint           string
	Username           string
	Password           string
	Headers            map[string]string
	SamplingRatio      float64
	PropagatorTypes    []string
	BatchTimeout       time.Duration
	ExportTimeout      time.Duration
	MaxExportBatchSize int
	MaxQueueSize       int
}

func InitTracer(cfg Config) (*sdktrace.TracerProvider, error) {
	var exporter sdktrace.SpanExporter
	var err error

	ctx := context.Background()

	switch cfg.ExporterType {
	case ExporterTypeJaeger:
		exporter, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.Endpoint)))
	case ExporterTypeZipkin:
		exporter, err = zipkin.New(cfg.Endpoint)
	case ExporterTypeOTLPHTTP:
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(cfg.Endpoint),
		}

		if cfg.Username != "" && cfg.Password != "" {
			opts = append(opts, otlptracehttp.WithBasicAuth(cfg.Username, cfg.Password))
		}

		for k, v := range cfg.Headers {
			opts = append(opts, otlptracehttp.WithHeader(k, v))
		}

		client := otlptracehttp.NewClient(opts...)
		exporter, err = otlptrace.New(ctx, client)

	case ExporterTypeOTLPGRPC:
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
		}

		if cfg.Username != "" && cfg.Password != "" {
			opts = append(opts, otlptracegrpc.WithBasicAuth(cfg.Username, cfg.Password))
		}

		for k, v := range cfg.Headers {
			opts = append(opts, otlptracegrpc.WithHeader(k, v))
		}

		client := otlptracegrpc.NewClient(opts...)
		exporter, err = otlptrace.New(ctx, client)
	default:
		exporter, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.Endpoint)))
	}

	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	batchOpts := []sdktrace.BatchSpanProcessorOption{
		sdktrace.WithMaxExportBatchSize(cfg.MaxExportBatchSize),
		sdktrace.WithBatchTimeout(cfg.BatchTimeout),
		sdktrace.WithExportTimeout(cfg.ExportTimeout),
		sdktrace.WithMaxQueueSize(cfg.MaxQueueSize),
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter, batchOpts...)

	var sampler sdktrace.Sampler
	if cfg.SamplingRatio >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.SamplingRatio <= 0.0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.SamplingRatio)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	propagator := createPropagator(cfg.PropagatorTypes)
	otel.SetTextMapPropagator(propagator)
	otel.SetTracerProvider(tp)

	return tp, nil
}

func createPropagator(types []string) propagation.TextMapPropagator {
	if len(types) == 0 {
		return propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
	}

	var propagators []propagation.TextMapPropagator

	for _, t := range types {
		switch t {
		case "tracecontext":
			propagators = append(propagators, propagation.TraceContext{})
		case "baggage":
			propagators = append(propagators, propagation.Baggage{})
		case "b3":
			// Add B3 propagator if available
		case "jaeger":
			// Add Jaeger propagator if available
		}
	}

	return propagation.NewCompositeTextMapPropagator(propagators...)
}
