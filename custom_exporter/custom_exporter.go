package custom_exporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type metricsExporter struct {
	config     *Config
	logger     *zap.Logger
	httpClient *http.Client
}

// MetricData represents the structure sent to your custom endpoint
type MetricData struct {
	Timestamp   int64             `json:"timestamp"`
	MetricName  string            `json:"metric_name"`
	Value       float64           `json:"value"`
	Unit        string            `json:"unit"`
	MetricType  string            `json:"metric_type"`
	Labels      map[string]string `json:"labels"`
	ResourceLabels map[string]string `json:"resource_labels"`
}

type MetricsPayload struct {
	Metrics []MetricData `json:"metrics"`
}

func (e *metricsExporter) Start(ctx context.Context, host component.Host) error {
	e.logger.Info("Starting Custom Metrics Exporter")

	httpClient, err := e.config.ClientConfig.ToClient(ctx, host, component.TelemetrySettings{})
	if err != nil {
		return err
	}
	e.httpClient = httpClient

	e.logger.Info("Initialized Custom Exporter")
	return nil
}

func (e *metricsExporter) pushMetrics(ctx context.Context, md pmetric.Metrics) error {
	payload, err := e.convertMetrics(md)
	if err != nil {
		return fmt.Errorf("failed to convert metrics: %w", err)
	}

	return e.sendMetrics(ctx, payload)
}

func (e *metricsExporter) convertMetrics(md pmetric.Metrics) (*MetricsPayload, error) {
	var metrics []MetricData

	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		resourceLabels := make(map[string]string)
		
		// Extract resource attributes
		// rm.Resource().Attributes().Range(func(k string, v pmetric.AttributeValue) bool {
		// 	resourceLabels[k] = v.AsString()
		// 	return true
		// })

		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			
			for k := 0; k < sm.Metrics().Len(); k++ {
				metric := sm.Metrics().At(k)
				metricType := ""
				
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					metricType = "gauge"
					e.processGauge(metric, resourceLabels, &metrics)
				case pmetric.MetricTypeSum:
					metricType = "counter"
					e.processSum(metric, resourceLabels, &metrics)
				case pmetric.MetricTypeHistogram:
					metricType = "histogram"
					e.processHistogram(metric, resourceLabels, &metrics)
				default:
					e.logger.Warn("Unsupported metric type", zap.String("type", metric.Type().String()))
					continue
				}
				
				_ = metricType // Use metricType as needed
			}
		}
	}

	return &MetricsPayload{Metrics: metrics}, nil
}

func (e *metricsExporter) processGauge(metric pmetric.Metric, resourceLabels map[string]string, metrics *[]MetricData) {
	gauge := metric.Gauge()
	for i := 0; i < gauge.DataPoints().Len(); i++ {
		dp := gauge.DataPoints().At(i)
		labels := make(map[string]string)
		
		// dp.Attributes().Range(func(k string, v pmetric.AttributeValue) bool {
		// 	labels[k] = v.AsString()
		// 	return true
		// })

		*metrics = append(*metrics, MetricData{
			Timestamp:      dp.Timestamp().AsTime().Unix(),
			MetricName:     metric.Name(),
			Value:          dp.DoubleValue(),
			Unit:          metric.Unit(),
			MetricType:    "gauge",
			Labels:        labels,
			ResourceLabels: resourceLabels,
		})
	}
}

func (e *metricsExporter) processSum(metric pmetric.Metric, resourceLabels map[string]string, metrics *[]MetricData) {
	sum := metric.Sum()
	for i := 0; i < sum.DataPoints().Len(); i++ {
		dp := sum.DataPoints().At(i)
		labels := make(map[string]string)
		
		// dp.Attributes().Range(func(k string, v pmetric.AttributeValue) bool {
		// 	labels[k] = v.AsString()
		// 	return true
		// })

		*metrics = append(*metrics, MetricData{
			Timestamp:      dp.Timestamp().AsTime().Unix(),
			MetricName:     metric.Name(),
			Value:          dp.DoubleValue(),
			Unit:          metric.Unit(),
			MetricType:    "counter",
			Labels:        labels,
			ResourceLabels: resourceLabels,
		})
	}
}

func (e *metricsExporter) processHistogram(metric pmetric.Metric, resourceLabels map[string]string, metrics *[]MetricData) {
	histogram := metric.Histogram()
	for i := 0; i < histogram.DataPoints().Len(); i++ {
		dp := histogram.DataPoints().At(i)
		labels := make(map[string]string)
		
		// dp.Attributes().Range(func(k string, v pmetric.AttributeValue) bool {
		// 	labels[k] = v.AsString()
		// 	return true
		// })

		// Send count and sum as separate metrics
		*metrics = append(*metrics, MetricData{
			Timestamp:      dp.Timestamp().AsTime().Unix(),
			MetricName:     metric.Name() + "_count",
			Value:          float64(dp.Count()),
			Unit:          "",
			MetricType:    "counter",
			Labels:        labels,
			ResourceLabels: resourceLabels,
		})

		if dp.HasSum() {
			*metrics = append(*metrics, MetricData{
				Timestamp:      dp.Timestamp().AsTime().Unix(),
				MetricName:     metric.Name() + "_sum",
				Value:          dp.Sum(),
				Unit:          metric.Unit(),
				MetricType:    "counter",
				Labels:        labels,
				ResourceLabels: resourceLabels,
			})
		}
	}
}

func (e *metricsExporter) sendMetrics(ctx context.Context, payload *MetricsPayload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.config.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "opentelemetry-collector/custom-http-exporter")

	// Add API key if configured
	if e.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+string(e.config.APIKey))
	}

	// Add custom headers
	for k, v := range e.config.CustomHeaders {
		req.Header.Set(k, v)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	e.logger.Debug("Successfully sent metrics", 
		zap.Int("status_code", resp.StatusCode),
		zap.Int("metric_count", len(payload.Metrics)))

	return nil
}