// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	infraCreateOrUpdateFailed          = metrics.NewCounter("infra_create_or_update_failed", "Number of created or updated infrastructures that failed.")
	infraCreateOrUpdateSuccess         = metrics.NewCounter("infra_create_or_update_success", "Number of created or updated infrastructures that succeed.")
	infraCreateOrUpdateDurationSeconds = metrics.NewHistogram("infra_create_or_update_duration_seconds", "How long in seconds a infrastructure is created or updated successfully.", []float64{0.001, 0.01, 0.1, 1, 5, 10})

	infraDeleteFailed          = metrics.NewCounter("infra_delete_failed", "Number of deleted infrastructures that failed.")
	infraDeleteSuccess         = metrics.NewCounter("infra_delete_success", "Number of deleted infrastructures that succeed.")
	infraDeleteDurationSeconds = metrics.NewHistogram("infra_delete_duration_seconds", "How long in seconds a infrastructure is deleted successfully.", []float64{0.001, 0.01, 0.1, 1, 5, 10})

	kindLabel  = metrics.NewLabel("kind")
	infraLabel = metrics.NewLabel("infra")
)
