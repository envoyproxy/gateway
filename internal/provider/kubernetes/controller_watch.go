// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"

	certificatesv1a1 "k8s.io/api/certificates/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func (r *gatewayAPIReconciler) watchClusterTrustBundle(c controller.Controller, mgr manager.Manager) error {
	r.clusterTrustBundleExits = r.crdExists(mgr, resource.KindClusterTrustBundle, certificatesv1a1.SchemeGroupVersion.String())
	if !r.clusterTrustBundleExits {
		r.log.Info("Skipping watch", "kind", resource.KindClusterTrustBundle, "groupVersion", certificatesv1a1.SchemeGroupVersion.String())
	} else {
		predicates := commonPredicates[*certificatesv1a1.ClusterTrustBundle]()
		predicates = append(predicates,
			predicate.NewTypedPredicateFuncs(func(ctb *certificatesv1a1.ClusterTrustBundle) bool {
				return r.validateClusterTrustBundleForReconcile(ctb)
			}),
		)
		if err := c.Watch(
			source.Kind(mgr.GetCache(), &certificatesv1a1.ClusterTrustBundle{},
				handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, obj *certificatesv1a1.ClusterTrustBundle) []reconcile.Request {
					return r.enqueueClass(ctx, obj)
				}),
				predicates...)); err != nil {
			return err
		}
	}

	return nil
}
