// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func (r *gatewayAPIReconciler) getAuthenticationFilters(ctx context.Context) ([]egv1a1.AuthenticationFilter, error) {
	authenList := new(egv1a1.AuthenticationFilterList)
	if err := r.client.List(ctx, authenList); err != nil {
		return nil, fmt.Errorf("failed to list AuthenticationFilters: %v", err)
	}

	authens := authenList.Items
	if len(r.namespaceLabels) != 0 {
		var as []egv1a1.AuthenticationFilter
		for _, a := range authens {
			a := a
			ok, err := r.checkObjectNamespaceLabels(&a)
			if err != nil {
				// TODO: should return? or just proceed?
				return nil, fmt.Errorf("failed to check namespace labels for AuthenicationFilter %s: %s", a.GetName(), err)
			}

			if ok {
				as = append(as, a)
			}
		}

		authens = as
	}

	return authens, nil
}

func (r *gatewayAPIReconciler) getRateLimitFilters(ctx context.Context) ([]egv1a1.RateLimitFilter, error) {
	rateLimitList := new(egv1a1.RateLimitFilterList)
	if err := r.client.List(ctx, rateLimitList); err != nil {
		return nil, fmt.Errorf("failed to list RateLimitFilters: %v", err)
	}

	rateLimits := rateLimitList.Items
	if len(r.namespaceLabels) != 0 {
		var rls []egv1a1.RateLimitFilter
		for _, rl := range rateLimits {
			rl := rl
			ok, err := r.checkObjectNamespaceLabels(&rl)
			if err != nil {
				// TODO: should return? or just proceed?
				return nil, fmt.Errorf("failed to check namespace labels for RateLimitFilter %s: %s", rl.GetName(), err)
			}

			if ok {
				rls = append(rls, rl)
			}
		}

		rateLimits = rls
	}

	return rateLimits, nil
}

func (r *gatewayAPIReconciler) getExtensionRefFilters(ctx context.Context) ([]unstructured.Unstructured, error) {
	var resourceItems []unstructured.Unstructured
	for _, gvk := range r.extGVKs {
		uExtResourceList := &unstructured.UnstructuredList{}
		uExtResourceList.SetGroupVersionKind(gvk)
		if err := r.client.List(ctx, uExtResourceList); err != nil {
			r.log.Info("no associated resources found for %s", gvk.String())
			return nil, fmt.Errorf("failed to list %s: %v", gvk.String(), err)
		}

		uExtResources := uExtResourceList.Items
		if len(r.namespaceLabels) != 0 {
			var extRs []unstructured.Unstructured
			for _, extR := range uExtResources {
				extR := extR
				ok, err := r.checkObjectNamespaceLabels(&extR)
				if err != nil {
					// TODO: should return? or just proceed?
					return nil, fmt.Errorf("failed to check namespace labels for ExtensionRefFilter %s: %s", extR.GetName(), err)
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
