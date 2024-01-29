package otel

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Init initialize the OTEL system
func (C *Config) Init(serviceName string, serviceVersion string) error {

	// if not enabled -> return
	if !C.Enabled {
		return nil
	}

	// create exporter
	exporter, err := C.NewSpanExporter()
	if err != nil {
		return err
	}

	// create resource
	resource := C.NewResource(serviceName, serviceVersion)

	// create tracer provider
	C.NewTracerProvider(resource, exporter)

	return nil
}

func (C *Config) NewSpanExporter() (exp sdktrace.SpanExporter, err error) {

	switch C.Exporter {

	case stdout:
		exp, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		break
	case otlp:
		exp, err = C.NewOTLPExporter()
		break
	default:
		err = fmt.Errorf("unknown exporter: %v", C.Exporter)
		break
	}

	return exp, err
}
func (C *Config) NewOTLPExporter() (exp sdktrace.SpanExporter, err error) {

	ctx := context.Background()

	// not really sure why they do this
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, C.OtlpCollectorUrl,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	return otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
}

// NewResource creates a new resource
func (C *Config) NewResource(serviceName string, serviceVersion string) *sdkresource.Resource {

	r, err := sdkresource.Merge(
		sdkresource.Default(),
		sdkresource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)

	if err != nil {
		log.Println("ERROR: could not create resource: ", err)
	}

	return r
}

// NewTracerProvider creates the tracer provider and saves it to the state struct
func (C *Config) NewTracerProvider(resource *sdkresource.Resource, spanExporter sdktrace.SpanExporter) {

	// batch span processor to aggregate spans before export.
	batchSpanProcessor := sdktrace.NewBatchSpanProcessor(spanExporter)

	// create tracer provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource),
		sdktrace.WithSpanProcessor(batchSpanProcessor),
	)

	// save tracerProvider to state
	state.tracerProvider = tracerProvider

	// set tp as a global tracer provider
	otel.SetTracerProvider(tracerProvider)

	// setup propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

// ShutdownTracerProvider closes the tracer provider and flushes traces
// to be called when program exits
func (C *Config) ShutdownTracerProvider(ctx context.Context) error {

	// if not enabled -> return
	if !C.Enabled {
		return nil
	}

	// close tracer provider
	return state.tracerProvider.Shutdown(ctx)
}

// NewSpan creates a new span if otel is enabled
func (C *Config) NewSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {

	if !C.Enabled {
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
func (C *Config) RecordError(span trace.Span, err error) {

	if C.Enabled {
		span.RecordError(err)
	}
}
