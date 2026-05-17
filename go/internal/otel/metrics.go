package otel

import (
	"context"
	"fmt"
	"time"

	"github.com/koblas/besops/lib/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type MetricsExporter struct {
	provider  *sdkmetric.MeterProvider
	status    otelmetric.Int64Gauge
	latency   otelmetric.Float64Histogram
	checkTime otelmetric.Int64Counter
}

func NewMetricsExporter(ctx context.Context, endpoint string) (*MetricsExporter, error) {
	opts := []otlpmetrichttp.Option{}
	if endpoint != "" {
		opts = append(opts, otlpmetrichttp.WithEndpoint(endpoint))
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP metric exporter: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(30*time.Second))),
	)

	meter := provider.Meter("besops.monitor")

	statusGauge, err := meter.Int64Gauge("monitor.status",
		otelmetric.WithDescription("Current monitor status (1=up, 0=down, 2=pending, 3=maintenance)"),
	)
	if err != nil {
		return nil, fmt.Errorf("creating status gauge: %w", err)
	}

	latencyHist, err := meter.Float64Histogram("monitor.latency_ms",
		otelmetric.WithDescription("Monitor check latency in milliseconds"),
		otelmetric.WithExplicitBucketBoundaries(5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000),
	)
	if err != nil {
		return nil, fmt.Errorf("creating latency histogram: %w", err)
	}

	checkCounter, err := meter.Int64Counter("monitor.checks_total",
		otelmetric.WithDescription("Total number of monitor checks performed"),
	)
	if err != nil {
		return nil, fmt.Errorf("creating check counter: %w", err)
	}

	return &MetricsExporter{
		provider:  provider,
		status:    statusGauge,
		latency:   latencyHist,
		checkTime: checkCounter,
	}, nil
}

func (m *MetricsExporter) Record(ctx context.Context, monitor telemetry.MonitorInfo, up bool, latencyMs int64) {
	kvs := []attribute.KeyValue{
		attribute.String("monitor.id", monitor.MonitorID()),
		attribute.String("monitor.name", monitor.MonitorName()),
		attribute.String("monitor.type", monitor.MonitorType()),
	}
	if gname := monitor.GroupName(); gname != "" {
		kvs = append(kvs, attribute.String("monitor.group", gname))
	}
	for _, tag := range monitor.Tags() {
		kvs = append(kvs, attribute.String("monitor.tag."+tag, "true"))
	}
	attrSet := otelmetric.WithAttributeSet(attribute.NewSet(kvs...))

	var statusVal int64
	if up {
		statusVal = 1
	}
	m.status.Record(ctx, statusVal, attrSet)
	m.checkTime.Add(ctx, 1, attrSet)

	if latencyMs > 0 {
		m.latency.Record(ctx, float64(latencyMs), attrSet)
	}
}

func (m *MetricsExporter) Shutdown(ctx context.Context) error {
	return m.provider.Shutdown(ctx)
}
