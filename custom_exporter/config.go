package custom_exporter

import (
	"errors"

	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

// Config defines configuration for the custom HTTP metrics exporter
type Config struct {
	TimeoutSettings exporterhelper.TimeoutConfig `mapstructure:",squash"`
	QueueSettings   exporterhelper.QueueBatchConfig   `mapstructure:"sending_queue"`
	// exporterhelper.RetrySettings   `mapstructure:"retry_on_failure"`
	BackoffConfig   configretry.BackOffConfig  
	

	confighttp.ClientConfig `mapstructure:",squash"`//confighttp.HTTPClientSettings `mapstructure:",squash"`

	// Endpoint is the URL to send metrics to
	Endpoint string `mapstructure:"endpoint"`
	
	// APIKey for authentication (optional)
	APIKey configopaque.String `mapstructure:"api_key"`
	
	// CustomHeaders to add to each request
	CustomHeaders map[string]string `mapstructure:"custom_headers"`
	
	// MetricFormat specifies the format to send metrics in (json, prometheus, etc.)
	MetricFormat string `mapstructure:"metric_format"`
}

func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	
	if c.MetricFormat == "" {
		c.MetricFormat = "json" // default format
	}
	
	return nil
}