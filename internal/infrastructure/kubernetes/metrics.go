// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"github.com/envoyproxy/gateway/internal/metrics"
)

var (
	infraManagerResourcesCreated = metrics.NewCounter(
		"infra_manager_resources_created_total",
		"Total number of the resources created by infra manager.",
	)

	infraManagerResourcesUpdated = metrics.NewCounter(
		"infra_manager_resources_updated_total",
		"Total number of the resources updated by infra manager.",
	)

	infraManagerResourcesDeleted = metrics.NewCounter(
		"infra_manager_resources_deleted_total",
		"Total number of the resources deleted by infra manager.",
	)

	infraManagerResourcesErrors = metrics.NewCounter(
		"infra_manager_resources_errors_total",
		"Total number of the resources errors encountered by infra manager.",
	)

	// metrics label definitions
	operationLabel            = metrics.NewLabel("operation")
	k8sResourceTypeLabel      = metrics.NewLabel("k8s_resource_type")
	k8sResourceNamespaceLabel = metrics.NewLabel("k8s_resource_namespace")
	k8sResourceNameLabel      = metrics.NewLabel("k8s_resource_name")
)
