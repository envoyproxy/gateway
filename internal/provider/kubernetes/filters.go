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

	return authenList.Items, nil
}

func (r *gatewayAPIReconciler) getExtensionRefFilters(ctx context.Context) ([]unstructured.Unstructured, error) {
	var resourceItems []unstructured.Unstructured
	for _, gvk := range r.extGVKs {
		uExtResources := &unstructured.UnstructuredList{}
		uExtResources.SetGroupVersionKind(gvk)
		if err := r.client.List(ctx, uExtResources); err != nil {
			r.log.Info("no associated resources found for %s", gvk.String())
			return nil, fmt.Errorf("failed to list %s: %v", gvk.String(), err)
		}

		resourceItems = append(resourceItems, uExtResources.Items...)
	}

	return resourceItems, nil
}
