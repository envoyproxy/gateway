// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
				extR := extR
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
