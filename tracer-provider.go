package otel

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	tra "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

func (c *Config) NewTracerProvider() (exp trace.SpanExporter, err error) {

	switch c.Exporter {

	case stdout:
		exp, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		break
	case otlp:
		exp, err = c.NewOTLPTracerProvider()
		break
	case file:
		err = fmt.Errorf("not Implemented yet")
		break
	default:
		err = fmt.Errorf("unknown exporter: %v", c.Exporter)
		break
	}

	return exp, err
}

func (c *Config) NewOTLPTracerProvider() (exp trace.SpanExporter, err error) {

	ctx := context.Background()

	// not really sure why they do this
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.OtlpCollectorUrl,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	return otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
}

func (c *Config) NewResource(serviceName string, serviceVersion string) *resource.Resource {

	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	return r
}

func (c *Config) InitTracerProvider(res *resource.Resource, exp trace.SpanExporter) {

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := trace.NewBatchSpanProcessor(exp)
	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
	)

	c.tp = tracerProvider

	// set tp as a global tracer provider
	otel.SetTracerProvider(tracerProvider)
	//otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func (c *Config) ShutdownTracerProvider(ctx context.Context) error {
	return c.tp.Shutdown(ctx)
}

func (c *Config) NewSpan(ctx context.Context, spanName string) (context.Context, tra.Span) {

	if !c.Enabled {
		return ctx, nil
	}

	// retrieve span name
	value := ctx.Value("trace_name")

	traceName, ok := value.(string)
	if !ok {
		traceName = spanName
	}

	return otel.Tracer(traceName).Start(ctx, spanName)
}

// RecordError records error to span only if otel is enabled, removing nullpointer exceptions
func (c *Config) RecordError(span tra.Span, err error) {

	if c.Enabled {
		span.RecordError(err)
	}
}
