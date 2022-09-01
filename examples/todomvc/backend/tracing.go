package main

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// NewTracer creates a new TracerProvider. It must be closed on service exit
// using the returned io.Closer.
func NewTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	var bsp sdktrace.SpanProcessor
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		panic(fmt.Sprintf("failed to initialise stdouttrace exporter %v\n", err))
	}
	bsp = sdktrace.NewBatchSpanProcessor(exp)

	resources, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithContainer(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource")
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(resources),
	), nil
}
