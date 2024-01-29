package otel

import (
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type State struct {
	tracerProvider *sdktrace.TracerProvider
}

var state State
