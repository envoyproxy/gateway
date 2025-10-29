// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package troubleshoot

import (
	"context"
	"errors"
	"fmt"

	troubleshootv1b2 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	tbcollect "github.com/replicatedhq/troubleshoot/pkg/collect"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/rest"

	"github.com/envoyproxy/gateway/internal/troubleshoot/collect"
)

type CollectorType string

const (
	CollectorTypeEnvoyGatewayResource CollectorType = "EnvoyGatewayResource"
	CollectorTypePrometheusMetrics    CollectorType = "PrometheusMetrics"
	CollectorTypePodLogs              CollectorType = "PodLogs"
	CollectorTypeConfigDump           CollectorType = "ConfigDump"
)

type CollectOptions struct {
	namespaces []string
	bundlePath string
	// enableSDS indicates whether to remove the SDS section from the config dump
	// to avoid collecting sensitive information.
	// Default to false
	enableSDS bool

	disabledCollectors sets.Set[CollectorType]
}

type CollectOption func(opts *CollectOptions)

func WithBundlePath(path string) CollectOption {
	return func(opts *CollectOptions) {
		opts.bundlePath = path
	}
}

func WithCollectedNamespaces(ns []string) CollectOption {
	return func(opts *CollectOptions) {
		opts.namespaces = ns
	}
}

func DisableCollector(collector CollectorType) CollectOption {
	return func(opts *CollectOptions) {
		if opts.disabledCollectors == nil {
			opts.disabledCollectors = sets.New[CollectorType]()
		}
		opts.disabledCollectors.Insert(collector)
	}
}

func WithSDS(enabled bool) CollectOption {
	return func(opts *CollectOptions) {
		opts.enableSDS = enabled
	}
}

func CollectResult(ctx context.Context, restConfig *rest.Config, opts ...CollectOption) (tbcollect.CollectorResult, error) {
	collectorOpts := &CollectOptions{
		enableSDS: false,
	}
	for _, o := range opts {
		o(collectorOpts)
	}

	var result tbcollect.CollectorResult
	progressChan := make(chan interface{})
	go func() {
		select {
		case <-ctx.Done():
			close(progressChan)
		case msg := <-progressChan:
			fmt.Printf("Collecting support bundle: %v\n", msg)
		}
	}()

	collectors := []tbcollect.Collector{}
	bundlePath := collectorOpts.bundlePath
	for _, ns := range collectorOpts.namespaces {
		collectorList := []struct {
			cType     CollectorType
			collector tbcollect.Collector
		}{
			{
				cType: CollectorTypeEnvoyGatewayResource,
				// Collect the custom resources from Gateway API and EG
				collector: &collect.CustomResource{
					ClientConfig: restConfig,
					BundlePath:   collectorOpts.bundlePath,
					Namespace:    ns,
					IncludeGroups: []string{
						"gateway.envoyproxy.io",
						"gateway.networking.k8s.io",
					},
				},
			},
			{
				cType: CollectorTypeEnvoyGatewayResource,
				collector: collect.EnvoyGatewayResource{
					ClientConfig: restConfig,
					BundlePath:   bundlePath,
					Namespace:    ns,
				},
			},
			{
				cType: CollectorTypePodLogs,
				collector: &tbcollect.CollectLogs{
					Collector: &troubleshootv1b2.Logs{
						Name:      "pod-logs",
						Namespace: ns,
					},
					ClientConfig: restConfig,
					BundlePath:   bundlePath,
					Context:      ctx,
				},
			},
			{
				cType: CollectorTypePrometheusMetrics,
				collector: collect.PrometheusMetric{
					BundlePath:   bundlePath,
					ClientConfig: restConfig,
					Namespace:    ns,
				},
			},
			{
				cType: CollectorTypeConfigDump,
				collector: collect.ConfigDump{
					BundlePath:   bundlePath,
					ClientConfig: restConfig,
					Namespace:    ns,
					EnableSDS:    collectorOpts.enableSDS,
				},
			},
		}

		for _, c := range collectorList {
			if !collectorOpts.disabledCollectors.Has(c.cType) {
				collectors = append(collectors, c.collector)
			}
		}
	}

	total := len(collectors)
	allCollectedData := make(map[string][]byte)
	errs := make([]error, 0, len(collectors))
	for i, collector := range collectors {
		res, err := collector.Collect(progressChan)
		if err != nil {
			errs = append(errs, err)
			progressChan <- fmt.Errorf("failed to run collector: %s: %w", collector.Title(), err)
			progressChan <- tbcollect.CollectProgress{
				CurrentName:    collector.Title(),
				CurrentStatus:  "failed",
				CompletedCount: i + 1,
				TotalCount:     total,
			}
			continue
		}
		for k, v := range res {
			allCollectedData[k] = v
		}
	}
	result = allCollectedData

	return result, errors.Join(errs...)
}
