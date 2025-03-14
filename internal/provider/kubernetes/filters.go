// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/utils"
)

func (r *gatewayAPIReconciler) getExtensionRefFilters(ctx context.Context) ([]unstructured.Unstructured, error) {
	var resourceItems []unstructured.Unstructured
	for _, gvk := range r.extGVKs {
		uExtResourceList := &unstructured.UnstructuredList{}
		uExtResourceList.SetGroupVersionKind(gvk)
		if err := r.client.List(ctx, uExtResourceList); err != nil {
			r.log.Info("no associated resources found for %s", gvk.String())
			return nil, fmt.Errorf("failed to list %s: %w", gvk.String(), err)
		}

		uExtResources := uExtResourceList.Items
		if r.namespaceLabel != nil {
			var extRs []unstructured.Unstructured
			for _, extR := range uExtResources {
				ok, err := r.checkObjectNamespaceLabels(&extR)
				if err != nil {
					r.log.Error(err, "failed to check namespace labels for ExtensionRefFilter %s in namespace %s: %w", extR.GetName(), extR.GetNamespace())
					continue
				}
				if ok {
					extRs = append(extRs, extR)
				}
			}
			uExtResources = extRs
		}

		resourceItems = append(resourceItems, uExtResources...)
	}

	return resourceItems, nil
}

func (r *gatewayAPIReconciler) getHTTPRouteFilter(ctx context.Context, name, namespace string) (*egv1a1.HTTPRouteFilter, error) {
	hrf := new(egv1a1.HTTPRouteFilter)
	if err := r.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, hrf); err != nil {
		return nil, fmt.Errorf("failed to get HTTPRouteFilter: %w", err)
	}
	return hrf, nil
}

// processRouteFilterConfigMapRef adds the referenced ConfigMap in a HTTPRouteFilter
// to the resourceTree
func (r *gatewayAPIReconciler) processRouteFilterConfigMapRef(
	ctx context.Context, filter *egv1a1.HTTPRouteFilter,
	resourceMap *resourceMappings, resourceTree *resource.Resources,
) {
	if filter.Spec.DirectResponse != nil &&
		filter.Spec.DirectResponse.Body != nil &&
		filter.Spec.DirectResponse.Body.ValueRef != nil &&
		string(filter.Spec.DirectResponse.Body.ValueRef.Kind) == resource.KindConfigMap {
		configMap := new(corev1.ConfigMap)
		err := r.client.Get(ctx,
			types.NamespacedName{Namespace: filter.Namespace, Name: string(filter.Spec.DirectResponse.Body.ValueRef.Name)},
			configMap)
		// we don't return an error here, because we want to continue
		// reconciling the rest of the HTTPRouteFilter despite that this
		// reference is invalid.
		// This HTTPRouteFilter will be marked as invalid in its status
		// when translating to IR because the referenced configmap can't be
		// found.
		if err != nil {
			r.log.Error(err,
				"failed to process DirectResponse ValueRef for HTTPRouteFilter",
				"filter", filter, "ValueRef", filter.Spec.DirectResponse.Body.ValueRef.Name)
			return
		}
		resourceMap.allAssociatedNamespaces.Insert(filter.Namespace)
		if !resourceMap.allAssociatedConfigMaps.Has(utils.NamespacedName(configMap).String()) {
			resourceMap.allAssociatedConfigMaps.Insert(utils.NamespacedName(configMap).String())
			resourceTree.ConfigMaps = append(resourceTree.ConfigMaps, configMap)
			r.log.Info("processing ConfigMap", "namespace", filter.Namespace, "name", string(filter.Spec.DirectResponse.Body.ValueRef.Name))

		}
	}
}

// processRouteFilterSecretRef adds the referenced Secret in a HTTPRouteFilter
// to the resourceTree
func (r *gatewayAPIReconciler) processRouteFilterSecretRef(
	ctx context.Context, filter *egv1a1.HTTPRouteFilter,
	resourceMap *resourceMappings, resourceTree *resource.Resources,
) {
	if filter.Spec.CredentialInjection != nil {
		name := string(filter.Spec.CredentialInjection.Credential.ValueRef.Name)
		secret := new(corev1.Secret)
		err := r.client.Get(ctx, types.NamespacedName{Namespace: filter.Namespace, Name: name}, secret)
		// we don't return an error here, because we want to continue
		// reconciling the rest of the HTTPRouteFilter despite that this
		// reference is invalid.
		// This HTTPRouteFilter will be marked as invalid in its status
		// when translating to IR because the referenced secret can't be
		// found.
		if err != nil {
			r.log.Error(err,
				"failed to process CredentialInjection ValueRef for HTTPRouteFilter",
				"filter", filter, "ValueRef", filter.Spec.CredentialInjection.Credential.ValueRef.Name)
			return
		}
		resourceMap.allAssociatedNamespaces.Insert(filter.Namespace)
		if !resourceMap.allAssociatedSecrets.Has(utils.NamespacedName(secret).String()) {
			resourceMap.allAssociatedSecrets.Insert(utils.NamespacedName(secret).String())
			resourceTree.Secrets = append(resourceTree.Secrets, secret)
			r.log.Info("processing Secret", "namespace", filter.Namespace, "name", name)
		}
	}
}
