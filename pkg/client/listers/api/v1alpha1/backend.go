// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	apiv1alpha1 "github.com/envoyproxy/gateway/api/v1alpha1"
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
)

// BackendLister helps list Backends.
// All objects returned here must be treated as read-only.
type BackendLister interface {
	// List lists all Backends in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*apiv1alpha1.Backend, err error)
	// Backends returns an object that can list and get Backends.
	Backends(namespace string) BackendNamespaceLister
	BackendListerExpansion
}

// backendLister implements the BackendLister interface.
type backendLister struct {
	listers.ResourceIndexer[*apiv1alpha1.Backend]
}

// NewBackendLister returns a new BackendLister.
func NewBackendLister(indexer cache.Indexer) BackendLister {
	return &backendLister{listers.New[*apiv1alpha1.Backend](indexer, apiv1alpha1.Resource("backend"))}
}

// Backends returns an object that can list and get Backends.
func (s *backendLister) Backends(namespace string) BackendNamespaceLister {
	return backendNamespaceLister{listers.NewNamespaced[*apiv1alpha1.Backend](s.ResourceIndexer, namespace)}
}

// BackendNamespaceLister helps list and get Backends.
// All objects returned here must be treated as read-only.
type BackendNamespaceLister interface {
	// List lists all Backends in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*apiv1alpha1.Backend, err error)
	// Get retrieves the Backend from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*apiv1alpha1.Backend, error)
	BackendNamespaceListerExpansion
}

// backendNamespaceLister implements the BackendNamespaceLister
// interface.
type backendNamespaceLister struct {
	listers.ResourceIndexer[*apiv1alpha1.Backend]
}
