// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

var cacheKinds = []string{
	gwapiv1.Gateway{}.Kind, gwapiv1b1.ReferenceGrant{}.Kind, corev1.Namespace{}.Kind,
	corev1.Service{}.Kind, mcsapiv1a1.ServiceImport{}.Kind, discoveryv1.EndpointSlice{}.Kind,
	corev1.Secret{}.Kind, corev1.ConfigMap{}.Kind, egv1a1.EnvoyPatchPolicy{}.Kind,
	egv1a1.ClientTrafficPolicy{}.Kind, egv1a1.BackendTrafficPolicy{}.Kind, egv1a1.SecurityPolicy{}.Kind,
	gwapiv1a3.BackendTLSPolicy{}.Kind, egv1a1.EnvoyExtensionPolicy{}.Kind, egv1a1.HTTPRouteFilter{}.Kind,
}

// Set holds the resources with the kind.
// +k8s:deepcopy-gen=true
type Set struct {
	Values map[string]string
}

func newSet() *Set {
	return &Set{
		Values: map[string]string{},
	}
}

func (s *Set) Has(item string) bool {
	_, contained := s.Values[item]
	return contained
}

func (s *Set) Insert(item string) {
	s.Values[item] = ""
}

// Cache holds some duplicate resources in memory.
// +k8s:deepcopy-gen=true
type Cache struct {
	ResourceSet map[string]*Set
}

func (r *Resources) InitCache() {
	r.Cache = newCache()
}

func newCache() *Cache {
	resourceSets := make(map[string]*Set, len(cacheKinds))
	for _, kind := range cacheKinds {
		resourceSets[kind] = newSet()
	}

	return &Cache{
		ResourceSet: resourceSets,
	}
}

func (r *Resources) resourceCache(kind string) *Set {
	if rs, ok := r.Cache.ResourceSet[kind]; ok {
		return rs
	}

	return newSet()
}
