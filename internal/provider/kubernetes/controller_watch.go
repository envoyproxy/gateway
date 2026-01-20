// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	certificatesv1b1 "k8s.io/api/certificates/v1beta1"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func (r *gatewayAPIReconciler) watchClusterTrustBundle(ctx context.Context, c controller.Controller, mgr manager.Manager, discoveryClient discovery.DiscoveryInterface) error {
	groupVersion := certificatesv1b1.SchemeGroupVersion.String()
	exists, err := r.crdExists(ctx, discoveryClient, resource.KindClusterTrustBundle, groupVersion)
	if err != nil {
		return fmt.Errorf("failed to discover %s CRD: %w", resource.KindClusterTrustBundle, err)
	}
	r.clusterTrustBundleExits = exists
	if !r.clusterTrustBundleExits {
		r.log.Info("Skipping watch", "kind", resource.KindClusterTrustBundle, "groupVersion", groupVersion)
	} else {
		predicates := commonPredicates[*certificatesv1b1.ClusterTrustBundle]()
		predicates = append(predicates,
			predicate.NewTypedPredicateFuncs(func(ctb *certificatesv1b1.ClusterTrustBundle) bool {
				return r.validateClusterTrustBundleForReconcile(ctb)
			}),
		)
		if err := c.Watch(
			source.Kind(mgr.GetCache(), &certificatesv1b1.ClusterTrustBundle{},
				handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, obj *certificatesv1b1.ClusterTrustBundle) []reconcile.Request {
					return r.enqueueClass(ctx, obj)
				}),
				predicates...)); err != nil {
			return err
		}
	}

	return nil
}
