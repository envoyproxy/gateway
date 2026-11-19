// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

const (
	endpointSliceControllerName          = "endpointslice-controller.k8s.io"
	endpointSliceMirroringControllerName = "endpointslicemirroring-controller.k8s.io"
)

// applyAWSZoneIDs resolves local endpoints without changing cached EndpointSlices.
func (r *gatewayAPIReconciler) applyAWSZoneIDs(resources resource.ControllerResources) {
	if !r.store.platform.aws.Load() {
		return
	}

	zoneIDs := make(map[string]string)
	for _, res := range resources {
		for i, slice := range res.EndpointSlices {
			// Slices from other controllers can refer to nodes in another cluster.
			managedBy := slice.Labels[discoveryv1.LabelManagedBy]
			if managedBy != endpointSliceControllerName &&
				managedBy != endpointSliceMirroringControllerName {
				continue
			}
			var updated *discoveryv1.EndpointSlice
			for j, endpoint := range slice.Endpoints {
				if endpoint.NodeName == nil || *endpoint.NodeName == "" {
					continue
				}
				zoneID, found := zoneIDs[*endpoint.NodeName]
				if !found {
					zoneID = r.store.nodeZoneID(*endpoint.NodeName)
					zoneIDs[*endpoint.NodeName] = zoneID
				}
				if zoneID == "" || (endpoint.Zone != nil && *endpoint.Zone == zoneID) {
					continue
				}
				if updated == nil {
					updated = slice.DeepCopy()
				}
				if updated.Endpoints[j].Zone == nil {
					updated.Endpoints[j].Zone = new(zoneID)
				} else {
					*updated.Endpoints[j].Zone = zoneID
				}
			}
			if updated != nil {
				res.EndpointSlices[i] = updated
			}
		}
	}
}

func (r *gatewayAPIReconciler) nodePredicate() predicate.TypedPredicate[*corev1.Node] {
	return predicate.Or(
		predicate.TypedGenerationChangedPredicate[*corev1.Node]{},
		predicate.TypedFuncs[*corev1.Node]{
			UpdateFunc: func(e event.TypedUpdateEvent[*corev1.Node]) bool {
				return e.ObjectOld != nil && e.ObjectNew != nil &&
					(e.ObjectOld.Spec.ProviderID != e.ObjectNew.Spec.ProviderID ||
						e.ObjectOld.Labels[awsZoneIDLabel] != e.ObjectNew.Labels[awsZoneIDLabel])
			},
		},
	)
}
