package otel

import (
	"context"

	"github.com/gozix/di"
	gzGlue "github.com/gozix/glue/v3"
	gzViper "github.com/gozix/viper/v3"
	gzZap "github.com/gozix/zap/v2"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// BundleName is default definition name.
const BundleName = "otel"

var _ gzGlue.Bundle = (*Bundle)(nil)

type (
	// Bundle implements the glue.Bundle interface.
	Bundle struct{}

	// Option interface.
	Option interface {
		apply(b *Bundle)
	}

	// optionFunc wraps a func, so it satisfies the Option interface.
	optionFunc func(t *Bundle)
)

// NewBundle create bundle instance.
func NewBundle(options ...Option) *Bundle {
	var t = &Bundle{}

	for _, option := range options {
		option.apply(t)
	}

	return t
}

// Name implements the glue.Bundle interface.
func (t *Bundle) Name() string {
	return BundleName
}

// Build implements the glue.Bundle interface.
func (t *Bundle) Build(builder di.Builder) error {
	return builder.Apply(
		di.Provide(t.provideTracerProvider),
	)
}

// DependsOn implements the glue.DependsOn interface.
func (t *Bundle) DependsOn() []string {
	return []string{
		gzViper.BundleName,
		gzZap.BundleName,
	}
}

func (t *Bundle) provideTracerProvider(
	ctx context.Context,
	cfg *viper.Viper,
	logger *zap.Logger,
) trace.TracerProvider {
	// Create the Jaeger exporter
	var endpointOpt jaeger.EndpointOption
	switch cfg.GetString("otel.connection_type") {
	case "collector":
		endpoint := cfg.GetString("otel.collector.endpoint")
		endpointOpt = jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(endpoint),
		)
	case "agent":
		var opts []jaeger.AgentEndpointOption
		host := cfg.GetString("otel.agent.host")
		if host != "" {
			opts = append(opts, jaeger.WithAgentHost(host))
		}
		port := cfg.GetString("otel.agent.port")
		if port != "" {
			opts = append(opts, jaeger.WithAgentPort(port))
		}
		endpointOpt = jaeger.WithAgentEndpoint(opts...)
	default:
		logger.Warn("unknown connection type",
			zap.String("type", cfg.GetString("otel.connection_type")))
		return trace.NewNoopTracerProvider()
	}

	exp, err := jaeger.New(endpointOpt)
	if err != nil {
		logger.Error("init jaeger fail", zap.Error(err))
		return trace.NewNoopTracerProvider()
	}

	version, ok := ctx.Value("app.version").(string)
	if !ok {
		version = "unknown"
	}

	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.GetString("service")),
			attribute.String("environment", cfg.GetString("env")),
			attribute.String("version", version),
		)),
	)

	// Set global
	otel.SetTracerProvider(tp)

	return tp
}

// apply implements Option.
func (f optionFunc) apply(bundle *Bundle) {
	f(bundle)
}
