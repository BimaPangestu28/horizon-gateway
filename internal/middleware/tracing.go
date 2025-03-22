package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type TracingConfig struct {
	ServiceName     string
	TracerProvider  trace.TracerProvider
	Propagator      propagation.TextMapPropagator
	SamplingRate    float64
	IgnoredPaths    []string
	IgnoredMethods  []string
	EnabledForAdmin bool
}

func TracingMiddleware(config TracingConfig, logger logging.Logger) fiber.Handler {
	tracer := config.TracerProvider.Tracer(config.ServiceName)
	propagator := config.Propagator
	if propagator == nil {
		propagator = otel.GetTextMapPropagator()
	}

	pathMap := make(map[string]bool)
	for _, path := range config.IgnoredPaths {
		pathMap[path] = true
	}

	methodMap := make(map[string]bool)
	for _, method := range config.IgnoredMethods {
		methodMap[method] = true
	}

	return func(c *fiber.Ctx) error {
		path := c.Path()
		method := c.Method()

		isAdminPath := path == "/admin" || len(path) >= 7 && path[:7] == "/admin/"
		if isAdminPath && !config.EnabledForAdmin {
			return c.Next()
		}

		if pathMap[path] || methodMap[method] {
			return c.Next()
		}

		carrier := propagation.HeaderCarrier{}
		c.Request().Header.VisitAll(func(key, value []byte) {
			carrier.Set(string(key), string(value))
		})

		ctx := propagator.Extract(context.Background(), carrier)
		spanCtx := trace.SpanContextFromContext(ctx)

		opts := []trace.SpanStartOption{
			trace.WithAttributes(
				attribute.String("http.method", method),
				attribute.String("http.url", c.OriginalURL()),
				attribute.String("http.host", c.Hostname()),
				attribute.String("http.user_agent", string(c.Request().Header.UserAgent())),
				attribute.String("http.request_id", c.Get("X-Request-ID")),
				attribute.String("http.remote_addr", c.IP()),
			),
		}

		routeName := "unknown"
		if routeData := c.Locals("route"); routeData != nil {
			if route, ok := routeData.(map[string]string); ok {
				if name, exists := route["name"]; exists {
					routeName = name
					opts = append(opts, trace.WithAttributes(attribute.String("route.name", name)))
				}
			}
		}

		if spanCtx.IsValid() {
			opts = append(opts, trace.WithLinks(trace.Link{SpanContext: spanCtx}))
		}

		spanName := method + " " + routeName
		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()

		c.Locals("tracing.context", ctx)
		c.Locals("tracing.span", span)

		err := c.Next()

		status := c.Response().StatusCode()
		span.SetAttributes(attribute.Int("http.status_code", status))

		if err != nil {
			span.RecordError(err)
		}

		return err
	}
}

func ExtractTraceContext(c *fiber.Ctx) (context.Context, trace.Span) {
	ctx, ok := c.Locals("tracing.context").(context.Context)
	if !ok {
		return context.Background(), nil
	}

	span, ok := c.Locals("tracing.span").(trace.Span)
	if !ok {
		return ctx, nil
	}

	return ctx, span
}

func CreateChildSpan(c *fiber.Ctx, name string) (context.Context, trace.Span) {
	ctx, parentSpan := ExtractTraceContext(c)
	if parentSpan == nil {
		return ctx, nil
	}

	tracer := parentSpan.TracerProvider().Tracer("horizon-gateway")
	ctx, span := tracer.Start(ctx, name)

	return ctx, span
}

func InjectTraceContext(ctx context.Context, headers map[string]string) {
	propagator := otel.GetTextMapPropagator()
	carrier := propagation.MapCarrier(headers)
	propagator.Inject(ctx, carrier)
}
