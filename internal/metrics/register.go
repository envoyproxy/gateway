// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

const (
	defaultEndpoint = "/metrics"
)

// Init initializes and registers the global metrics server.
func Init(cfg *config.Server) error {
	options := newOptions(cfg)
	handler, err := registerForHandler(options)
	if err != nil {
		return err
	}

	if !options.pullOptions.disable {
		return start(options.address, handler)
	}

	return nil
}

func start(address string, handler http.Handler) error {
	handlers := http.NewServeMux()

	metricsLogger.Info("starting metrics server", "address", address)
	if handler != nil {
		handlers.Handle(defaultEndpoint, handler)
	}

	metricsServer := &http.Server{
		Handler:           handlers,
		Addr:              address,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
	}

	// Listen And Serve Metrics Server.
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil {
			metricsLogger.Error(err, "start metrics server failed")
		}
	}()

	return nil
}

func newOptions(svr *config.Server) registerOptions {
	newOpts := registerOptions{}
	newOpts.address = net.JoinHostPort(v1alpha1.GatewayMetricsHost, fmt.Sprint(v1alpha1.GatewayMetricsPort))

	if svr.EnvoyGateway.DisablePrometheus() {
		newOpts.pullOptions.disable = true
	} else {
		newOpts.pullOptions.disable = false
		newOpts.pullOptions.registry = metricsserver.Registry
		newOpts.pullOptions.gatherer = metricsserver.Registry
	}

	for _, config := range svr.EnvoyGateway.GetEnvoyGatewayTelemetry().Metrics.Sinks {
		newOpts.pushOptions.sinks = append(newOpts.pushOptions.sinks, metricsSink{
			host:     config.OpenTelemetry.Host,
			port:     config.OpenTelemetry.Port,
			protocol: config.OpenTelemetry.Protocol,
		})
	}

	return newOpts
}

// registerForHandler sets the global metrics registry to the provided Prometheus registerer.
// if enables prometheus, it will return a prom http handler.
func registerForHandler(opts registerOptions) (http.Handler, error) {
	otelOpts := []metric.Option{}

	if err := registerOTELPromExporter(&otelOpts, opts); err != nil {
		return nil, err
	}
	if err := registerOTELHTTPexporter(&otelOpts, opts); err != nil {
		return nil, err
	}
	if err := registerOTELgRPCexporter(&otelOpts, opts); err != nil {
		return nil, err
	}
	otelOpts = append(otelOpts, stores.preAddOptions()...)

	mp := metric.NewMeterProvider(otelOpts...)
	otel.SetMeterProvider(mp)

	if !opts.pullOptions.disable {
		return promhttp.HandlerFor(opts.pullOptions.gatherer, promhttp.HandlerOpts{}), nil
	}
	return nil, nil
}

// registerOTELPromExporter registers OTEL prometheus exporter (PULL mode).
func registerOTELPromExporter(otelOpts *[]metric.Option, opts registerOptions) error {
	if !opts.pullOptions.disable {
		promOpts := []otelprom.Option{
			otelprom.WithoutScopeInfo(),
			otelprom.WithoutTargetInfo(),
			otelprom.WithoutUnits(),
			otelprom.WithRegisterer(opts.pullOptions.registry),
			otelprom.WithoutCounterSuffixes(),
		}
		promreader, err := otelprom.New(promOpts...)
		if err != nil {
			return err
		}

		*otelOpts = append(*otelOpts, metric.WithReader(promreader))
		metricsLogger.Info("initialized metrics pull endpoint", "address", opts.address, "endpoint", defaultEndpoint)
	}

	return nil
}

// registerOTELHTTPexporter registers OTEL HTTP metrics exporter (PUSH mode).
func registerOTELHTTPexporter(otelOpts *[]metric.Option, opts registerOptions) error {
	for _, sink := range opts.pushOptions.sinks {
		if sink.protocol == v1alpha1.HTTPProtocol {
			address := net.JoinHostPort(sink.host, fmt.Sprint(sink.port))
			httpexporter, err := otlpmetrichttp.New(
				context.Background(),
				otlpmetrichttp.WithEndpoint(address),
				otlpmetrichttp.WithInsecure(),
			)
			if err != nil {
				return err
			}

			otelreader := metric.NewPeriodicReader(httpexporter)
			*otelOpts = append(*otelOpts, metric.WithReader(otelreader))
			metricsLogger.Info("initialized otel http metrics push endpoint", "address", address)
		}
	}

	return nil
}

// registerOTELgRPCexporter registers OTEL gRPC metrics exporter (PUSH mode).
func registerOTELgRPCexporter(otelOpts *[]metric.Option, opts registerOptions) error {
	for _, sink := range opts.pushOptions.sinks {
		if sink.protocol == v1alpha1.GRPCProtocol {
			address := net.JoinHostPort(sink.host, fmt.Sprint(sink.port))
			httpexporter, err := otlpmetricgrpc.New(
				context.Background(),
				otlpmetricgrpc.WithEndpoint(address),
				otlpmetricgrpc.WithInsecure(),
			)
			if err != nil {
				return err
			}

			otelreader := metric.NewPeriodicReader(httpexporter)
			*otelOpts = append(*otelOpts, metric.WithReader(otelreader))
			metricsLogger.Info("initialized otel grpc metrics push endpoint", "address", address)
		}
	}

	return nil
}

type registerOptions struct {
	address     string
	pullOptions struct {
		registry prometheus.Registerer
		gatherer prometheus.Gatherer
		disable  bool
	}
	pushOptions struct {
		sinks []metricsSink
	}
}

type metricsSink struct {
	protocol string
	host     string
	port     int32
}
