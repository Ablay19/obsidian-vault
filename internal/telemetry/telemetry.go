package telemetry

import (
	"log"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"
)

var (
	ZapLogger        *otelzap.Logger
	UnderlyingLogger *zap.Logger
)

// Init initializes the OpenTelemetry SDK and the otelzap logger.
func Init(serviceName string) (*trace.TracerProvider, error) {
	// Configure the console exporter.
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	// Create a new tracer provider with a batch span processor and the console exporter.
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	// Set the global tracer provider.
	otel.SetTracerProvider(tp)

	// Set the global propagator to W3C Trace Context.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// Initialize the otelzap logger.
	UnderlyingLogger = zap.Must(zap.NewDevelopment())
	ZapLogger = otelzap.New(UnderlyingLogger)

	log.Println("Telemetry initialized successfully.")
	return tp, nil
}
