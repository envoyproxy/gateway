// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// namespaceSelectorClient wraps a client.Client and automatically filters
// List results based on namespace label selector. This centralizes
// namespace filtering logic and ensures it's applied consistently to all
// List operations.
type namespaceSelectorClient struct {
	client.Client
	namespaceSelector *metav1.LabelSelector
}

// newNamespaceSelectorClient creates a new namespace-filtered client wrapper.
// If namespaceSelector is nil, the wrapper passes through all operations unchanged.
func newNamespaceSelectorClient(c client.Client, namespaceSelector *metav1.LabelSelector) client.Client {
	if namespaceSelector == nil {
		return c
	}
	return &namespaceSelectorClient{
		Client:            c,
		namespaceSelector: namespaceSelector,
	}
}

// List retrieves a list of objects and filters them based on namespace labels.
// Cluster-scoped resources are not filtered.
func (c *namespaceSelectorClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	// Call the underlying List first
	if err := c.Client.List(ctx, list, opts...); err != nil {
		return err
	}

	// If no namespace selector, return as-is
	if c.namespaceSelector == nil {
		return nil
	}

	// Filter the results by namespace labels
	return c.filterByNamespaceLabels(list)
}

// filterByNamespaceLabels filters the list items based on namespace label matching.
func (c *namespaceSelectorClient) filterByNamespaceLabels(list client.ObjectList) error {
	listKind := list.GetObjectKind().GroupVersionKind().Kind

	// Extract items from the list using k8s meta utilities
	items, err := meta.ExtractList(list)
	if err != nil {
		return fmt.Errorf("failed to extract items from list (listKind=%s): %w", listKind, err)
	}

	// Empty list - nothing to filter
	if len(items) == 0 {
		return nil
	}

	// Check if this is a cluster-scoped resource by examining the first item.
	// All items in a list are the same type, so if the first has no namespace,
	// they're all cluster-scoped and should not be filtered.
	if firstObj, ok := items[0].(metav1.Object); ok && firstObj.GetNamespace() == "" {
		return nil
	}

	// Filter items based on namespace labels
	var filtered []runtime.Object
	for _, item := range items {
		obj, ok := item.(metav1.Object)
		if !ok {
			return fmt.Errorf("item in list is not a metav1.Object (listKind=%s)", listKind)
		}

		matches, err := checkObjectNamespaceLabels(c.Client, c.namespaceSelector, obj)
		if err != nil {
			return fmt.Errorf("failed to check namespace labels for object %s/%s: %w",
				obj.GetNamespace(), obj.GetName(), err)
		}

		if matches {
			filtered = append(filtered, item)
		}
	}

	// Set the filtered items back to the list
	if err := meta.SetList(list, filtered); err != nil {
		return fmt.Errorf("failed to set filtered items to list (listKind=%s): %w", listKind, err)
	}
	return nil
}
