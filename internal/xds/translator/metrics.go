// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	// metrics definitions
	extensionManagerPostHookTimeSeconds = metrics.NewHistogram(
		"extension_manager_post_hook_time_seconds",
		"How long in seconds a post hook called in extension manager.",
		[]float64{0.001, 0.01, 0.1, 1, 5, 10},
	)

	extensionManagerPostHookCalls = metrics.NewCounter(
		"extension_manager_post_hook_calls_total",
		"Total number of the post hook calls in extension manager.",
	)

	extensionManagerPostHookCallErrors = metrics.NewCounter(
		"extension_manager_post_hook_call_errors_total",
		"Total number of the post hook call errors in extension manager.",
	)

	// metrics label definitions
	targetLabel = metrics.NewLabel("target")
)

const (
	routeTarget       = "route"
	virtualHostTarget = "virtualHost"
	listenerTarget    = "listener"
	clusterTarget     = "cluster"
)
