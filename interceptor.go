package otel

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
)

func (c *Config) UnaryServerInterceptor() grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		if !c.Enabled {
			return handler(ctx, req)
		}

		newCtx, span := CreateSpan(ctx, info.FullMethod)
		defer span.End()

		resp, err = handler(newCtx, req)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return resp, err
	}
}
func (c *Config) StreamServerInterceptor() grpc.StreamServerInterceptor {

	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		if !c.Enabled {
			return handler(srv, stream)
		}

		newCtx, span := CreateSpan(stream.Context(), info.FullMethod)
		defer span.End()

		err := handler(srv, &wrappedStream{stream, newCtx})
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return err
	}
}
func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) RecvMsg(m interface{}) error {

	name := fmt.Sprintf("grpc-received: %T - %v\n", m, m)

	newCtx, span := otel.Tracer(name).Start(w.Context(), "RecvMsg")
	defer span.End()
	w.ctx = newCtx

	return w.ServerStream.RecvMsg(m)
}
func (w *wrappedStream) SendMsg(m interface{}) error {

	name := fmt.Sprintf("grpc-received: %T - %v\n", m, m)

	newCtx, span := otel.Tracer(name).Start(w.Context(), "SendMsg")
	defer span.End()
	w.ctx = newCtx

	return w.ServerStream.SendMsg(m)
}

func CreateSpan(ctx context.Context, FullMethodName string) (context.Context, trace.Span) {

	completeMethodName := ExtractFullMethodNameFromInfoFullMethod(FullMethodName)

	newCtx, span := otel.Tracer(completeMethodName).Start(ctx, completeMethodName)
	span.SetAttributes(attribute.String("rpc.system", "grpc"))

	serviceName, methodName := GetServiceAndMethodFromInfoFullMethod(FullMethodName)

	span.SetAttributes(attribute.String("rpc.service", serviceName))
	span.SetAttributes(attribute.String("rpc.method", methodName))

	// save trace name in context
	newCtx = context.WithValue(newCtx, "trace_name", completeMethodName)

	return newCtx, span
}

func ExtractFullMethodNameFromInfoFullMethod(FullMethod string) string {

	_, methodName, _ := strings.Cut(FullMethod, "/")

	return methodName
}

func GetServiceAndMethodFromInfoFullMethod(FullMethod string) (string, string) {

	fullMethodName := ExtractFullMethodNameFromInfoFullMethod(FullMethod)

	serviceName, methodName, _ := strings.Cut(fullMethodName, "/")

	return serviceName, methodName
}
