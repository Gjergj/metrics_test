package custom_exporter

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const (
	typeStr   = "customexporter"
	stability = component.StabilityLevelBeta
)

// NewFactory creates a new exporter factory
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		exporter.WithMetrics(createMetricsExporter, stability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		TimeoutSettings: exporterhelper.TimeoutConfig{
			Timeout: 30 * time.Second,
		},
		BackoffConfig: configretry.NewDefaultBackOffConfig(),
		QueueSettings: exporterhelper.NewDefaultQueueConfig(),
		// HTTPClientSettings: confighttp.NewDefaultHTTPClientSettings(),
		MetricFormat: "json",
	}
}

func createMetricsExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Metrics, error) {
	c := cfg.(*Config)
	
	if err := c.Validate(); err != nil {
		return nil, err
	}

	exp := &metricsExporter{
		config: c,
		logger: set.Logger,
	}

	return exporterhelper.NewMetrics(
		ctx,
		set,
		cfg,
		exp.pushMetrics,
		exporterhelper.WithTimeout(c.TimeoutSettings),
		exporterhelper.WithRetry(c.BackoffConfig),
		exporterhelper.WithQueue(c.QueueSettings),
		exporterhelper.WithStart()
	)
}