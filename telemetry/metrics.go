package telemetry

import (
	"context"
	"fmt"
	"math/rand/v2"

	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
)

func InitMetrics(ctx context.Context, enableOtel bool, enableStdout bool) *metric.MeterProvider {
	var metricsOptions []metric.Option

	otelMeterTracerTags, err := resource.New(ctx, resource.WithAttributes(otelMetricsTracesAttributes...))
	ifErrorLogAndExit("create Otel GRPC metrics resource failed: %v", err)
	metricsOptions = append(metricsOptions, metric.WithResource(otelMeterTracerTags))

	if enableOtel {
		otelGrpcMetrics, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(OtelGrpcPush), otlpmetricgrpc.WithInsecure())
		ifErrorLogAndExit("create Otel GRPC metrics failed: %v", err)
		metricsOptions = append(metricsOptions, metric.WithReader(metric.NewPeriodicReader(otelGrpcMetrics, metric.WithInterval(MetricsTimeout))))
	}

	if enableStdout {
		stdoutMetrics, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
		ifErrorLogAndExit("create STDOUT metrics failed: %v", err)
		metric.NewPeriodicReader(stdoutMetrics)
		metricsOptions = append(metricsOptions, metric.WithReader(metric.NewPeriodicReader(stdoutMetrics, metric.WithInterval(MetricsTimeout))))
	}

	return metric.NewMeterProvider(metricsOptions...)
}

func DoMetricExample(ctx context.Context, telemetryService *Service) {
	exampleMetricsScope := telemetryService.MetricsProvider.Meter("example-scope")

	exampleMetricCounter, err := exampleMetricsScope.Float64UpDownCounter("example-counter")
	if err == nil {
		exampleMetricCounter.Add(ctx, 1)
		exampleMetricCounter.Add(ctx, -2)
		exampleMetricCounter.Add(ctx, 4)
	} else {
		telemetryService.Slogger.Error("metric failed", "error", fmt.Errorf("metric error: %w", err))
	}

	exampleMetricHistogram, err := exampleMetricsScope.Int64Histogram("example-histogram")
	if err == nil {
		exampleMetricHistogram.Record(ctx, rand.Int64N(100))
		exampleMetricHistogram.Record(ctx, rand.Int64N(100))
		exampleMetricHistogram.Record(ctx, rand.Int64N(100))
	} else {
		telemetryService.Slogger.Error("metric failed", "error", fmt.Errorf("metric error: %w", err))
	}
}
