// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package troubleshoot

import (
	"context"
	"fmt"

	troubleshootv1b2 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	tbcollect "github.com/replicatedhq/troubleshoot/pkg/collect"
	"k8s.io/client-go/rest"

	"github.com/envoyproxy/gateway/internal/troubleshoot/collect"
)

type CollectOptions struct {
	CollectedNamespaces []string
	BundlePath          string
}

func CollectResult(ctx context.Context, restConfig *rest.Config, opts CollectOptions) tbcollect.CollectorResult {
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
	for _, ns := range opts.CollectedNamespaces {
		bundlePath := opts.BundlePath
		collectors = append(collectors,
			// Collect logs from EnvoyGateway system namespace
			&tbcollect.CollectLogs{
				Collector: &troubleshootv1b2.Logs{
					Name:      "pod-logs",
					Namespace: ns,
				},
				ClientConfig: restConfig,
				BundlePath:   bundlePath,
				Context:      ctx,
			},
			// Collect config dump from EnvoyGateway system namespace
			collect.ConfigDump{
				BundlePath:   bundlePath,
				ClientConfig: restConfig,
				Namespace:    ns,
			})
	}
	total := len(collectors)
	allCollectedData := make(map[string][]byte)
	for i, collector := range collectors {
		res, err := collector.Collect(progressChan)
		if err != nil {
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

	return result
}
