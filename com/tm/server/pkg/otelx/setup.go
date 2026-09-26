package otelx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Config cấu hình OpenTelemetry của một service.
type Config struct {
	ServiceName    string
	ServiceVersion string
	// Endpoint OTLP HTTP của collector, vd http://localhost:4318.
	Endpoint string
	// Disabled tắt export (tracer/meter no-op); propagator vẫn bật.
	Disabled bool
	// MetricInterval: chu kỳ đẩy metric.
	MetricInterval time.Duration
}

const (
	defaultEndpoint       = "http://localhost:4318"
	defaultMetricInterval = 10 * time.Second
)

// ConfigFromEnv đọc biến môi trường chuẩn của OpenTelemetry:
// OTEL_SDK_DISABLED, OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_METRIC_EXPORT_INTERVAL (ms).
func ConfigFromEnv(service, version string, getenv func(string) string) (Config, error) {
	cfg := Config{
		ServiceName:    service,
		ServiceVersion: version,
		Endpoint:       getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		Disabled:       getenv("OTEL_SDK_DISABLED") == "true",
		MetricInterval: defaultMetricInterval,
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultEndpoint
	}
	if v := getenv("OTEL_METRIC_EXPORT_INTERVAL"); v != "" {
		ms, err := strconv.Atoi(v)
		if err != nil || ms <= 0 {
			return cfg, fmt.Errorf("OTEL_METRIC_EXPORT_INTERVAL: %q không phải số ms dương", v)
		}
		cfg.MetricInterval = time.Duration(ms) * time.Millisecond
	}
	return cfg, nil
}

// Setup cài TracerProvider, MeterProvider (OTLP HTTP) và propagator W3C làm global.
// Exporter kết nối lười: collector chưa chạy thì Setup vẫn thành công, lỗi export được log.
// Hàm trả về phải được gọi khi dừng service để đẩy nốt dữ liệu còn trong bộ đệm.
func Setup(ctx context.Context, cfg Config, log *slog.Logger) (shutdown func(context.Context) error, err error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	if cfg.Disabled {
		return func(context.Context) error { return nil }, nil
	}
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Warn("opentelemetry", slog.Any("error", err))
	}))

	u, err := url.Parse(cfg.Endpoint)
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("OTLP endpoint không hợp lệ %q", cfg.Endpoint)
	}
	traceOpts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(u.Host)}
	metricOpts := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpoint(u.Host)}
	if u.Scheme == "http" {
		traceOpts = append(traceOpts, otlptracehttp.WithInsecure())
		metricOpts = append(metricOpts, otlpmetrichttp.WithInsecure())
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
			attribute.String("service.version", cfg.ServiceVersion),
		),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, fmt.Errorf("otel resource: %w", err)
	}

	traceExp, err := otlptracehttp.New(ctx, traceOpts...)
	if err != nil {
		return nil, fmt.Errorf("otlp trace exporter: %w", err)
	}
	metricExp, err := otlpmetrichttp.New(ctx, metricOpts...)
	if err != nil {
		return nil, fmt.Errorf("otlp metric exporter: %w", err)
	}

	interval := cfg.MetricInterval
	if interval <= 0 {
		interval = defaultMetricInterval
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExp), sdktrace.WithResource(res))
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp, sdkmetric.WithInterval(interval))),
		sdkmetric.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)

	return func(ctx context.Context) error {
		return errors.Join(tp.Shutdown(ctx), mp.Shutdown(ctx))
	}, nil
}
