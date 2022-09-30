package jaeger

import (
	"fmt"
	"net/url"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

// InitGlobalTracer sets global jaeger tracer.
func InitGlobalTracer(serviceName string) error {
	u, err := url.JoinPath(defaultOption.CollectorHost, "api/traces")
	if err != nil {
		return errors.Wrapf(err, "join jaeger collector path failed, host(%s)",
			defaultOption.CollectorHost)
	}

	// jaeger tracer configuration
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		ServiceName: fmt.Sprintf("%s-gorm", serviceName),
		Reporter: &config.ReporterConfig{
			// LocalAgentHostPort:  "127.0.0.1:6381",
			LogSpans:            true,
			BufferFlushInterval: 100 * time.Millisecond,
			CollectorEndpoint:   u,
		},
	}

	// jaeger tracer client
	tracer, _, err := cfg.NewTracer(
		config.Logger(jaegerlog.StdLogger),
		config.ZipkinSharedRPCSpan(true),
	)
	if err != nil {
		return errors.Wrap(err, "failed to use jaeger tracer plugin")
	}

	// set into opentracing's global tracer, so the plugin would take it as default tracer.
	opentracing.SetGlobalTracer(tracer)

	return nil
}
