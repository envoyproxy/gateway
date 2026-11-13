// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	certificatesv1 "k8s.io/api/certificates/v1"
	certificatesv1b1 "k8s.io/api/certificates/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

// clusterTrustBundleGroupVersions lists the ClusterTrustBundle API versions Envoy Gateway can
// consume, in order of preference. certificates.k8s.io/v1 is GA (and the only version served by
// default) since Kubernetes 1.37; v1beta1 is kept as a fallback for older clusters that enabled it.
var clusterTrustBundleGroupVersions = []schema.GroupVersion{
	certificatesv1.SchemeGroupVersion,
	certificatesv1b1.SchemeGroupVersion,
}

func (r *gatewayAPIReconciler) watchClusterTrustBundle(ctx context.Context, c controller.Controller, mgr manager.Manager, discoveryClient discovery.DiscoveryInterface) error {
	r.clusterTrustBundleGV = schema.GroupVersion{}
	for _, gv := range clusterTrustBundleGroupVersions {
		exists, err := r.crdExists(ctx, discoveryClient, resource.KindClusterTrustBundle, gv.String())
		if err != nil {
			return fmt.Errorf("failed to discover %s CRD: %w", resource.KindClusterTrustBundle, err)
		}
		if exists {
			r.clusterTrustBundleGV = gv
			break
		}
	}

	switch r.clusterTrustBundleGV {
	case certificatesv1.SchemeGroupVersion:
		return watchClusterTrustBundleKind(r, c, mgr, &certificatesv1.ClusterTrustBundle{})
	case certificatesv1b1.SchemeGroupVersion:
		return watchClusterTrustBundleKind(r, c, mgr, &certificatesv1b1.ClusterTrustBundle{})
	default:
		r.log.Info("Skipping watch", "kind", resource.KindClusterTrustBundle, "groupVersions", clusterTrustBundleGroupVersions)
		return nil
	}
}

func watchClusterTrustBundleKind[T client.Object](r *gatewayAPIReconciler, c controller.Controller, mgr manager.Manager, obj T) error {
	r.log.Info("Watching", "kind", resource.KindClusterTrustBundle, "groupVersion", r.clusterTrustBundleGV.String())
	// Like Secrets and ConfigMaps, ClusterTrustBundles don't track metadata.generation, so
	// commonPredicates (GenerationChangedPredicate) would drop trustBundle updates such as CA rotation.
	return c.Watch(
		source.Kind(mgr.GetCache(), obj,
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, obj T) []reconcile.Request {
				return r.enqueueClass(ctx, obj)
			}),
			predicate.NewTypedPredicateFuncs(func(ctb T) bool {
				return r.validateClusterTrustBundleForReconcile(ctb)
			})))
}

// getClusterTrustBundle fetches the named ClusterTrustBundle using the API version discovered at
// startup and always returns it as a certificates.k8s.io/v1 object.
func (r *gatewayAPIReconciler) getClusterTrustBundle(ctx context.Context, name string) (*certificatesv1.ClusterTrustBundle, error) {
	key := types.NamespacedName{Name: name}
	switch r.clusterTrustBundleGV {
	case certificatesv1.SchemeGroupVersion:
		ctb := &certificatesv1.ClusterTrustBundle{}
		if err := r.client.Get(ctx, key, ctb); err != nil {
			return nil, err
		}
		return ctb, nil
	case certificatesv1b1.SchemeGroupVersion:
		ctb := &certificatesv1b1.ClusterTrustBundle{}
		if err := r.client.Get(ctx, key, ctb); err != nil {
			return nil, err
		}
		return &certificatesv1.ClusterTrustBundle{
			TypeMeta: metav1.TypeMeta{
				APIVersion: certificatesv1.SchemeGroupVersion.String(),
				Kind:       resource.KindClusterTrustBundle,
			},
			ObjectMeta: ctb.ObjectMeta,
			Spec: certificatesv1.ClusterTrustBundleSpec{
				SignerName:  ctb.Spec.SignerName,
				TrustBundle: ctb.Spec.TrustBundle,
			},
		}, nil
	default:
		return nil, fmt.Errorf("%s API is not served by the cluster, one of %v is required",
			resource.KindClusterTrustBundle, clusterTrustBundleGroupVersions)
	}
}
