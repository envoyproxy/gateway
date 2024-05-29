// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	resourceApplyTotal           = metrics.NewCounter("resource_apply_total", "Total number of applied resources.")
	resourceApplyFailed          = metrics.NewCounter("resource_apply_failed", "Number of applied resources that failed.")
	resourceApplySuccess         = metrics.NewCounter("resource_apply_success", "Number of applied resources that succeed.")
	resourceApplyDurationSeconds = metrics.NewHistogram("resource_apply_duration_seconds", "How long in seconds a resource be applied successfully.", []float64{0.001, 0.01, 0.1, 1, 5, 10})

	resourceDeleteTotal           = metrics.NewCounter("resource_delete_total", "Total number of deleted resources.")
	resourceDeleteFailed          = metrics.NewCounter("resource_delete_failed", "Number of deleted resources that failed.")
	resourceDeleteSuccess         = metrics.NewCounter("resource_delete_success", "Number of deleted resources that succeed.")
	resourceDeleteDurationSeconds = metrics.NewHistogram("resource_delete_duration_seconds", "How long in seconds a resource be deleted successfully.", []float64{0.001, 0.01, 0.1, 1, 5, 10})

	kindLabel      = metrics.NewLabel("kind")
	nameLabel      = metrics.NewLabel("name")
	namespaceLabel = metrics.NewLabel("namespace")
)
