// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"fmt"
	"sync"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

var lock sync.RWMutex

func (r *Resources) AppendResource(obj client.Object) bool {
	lock.Lock()
	defer lock.Unlock()

	objIdentification := string(obj.GetUID())
	if objIdentification == "" {
		objIdentification = fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName())
	}

	resourceCache := r.resourceCache(obj.GetObjectKind().GroupVersionKind().Kind)

	switch res := obj.(type) {
	case *gwapiv1.Gateway:
		if !resourceCache.Has(objIdentification) {
			resourceCache.Insert(objIdentification)
			r.Gateways = append(r.Gateways, res)
			return true
		}
	case *gwapiv1b1.ReferenceGrant:
		if !resourceCache.Has(objIdentification) {
			resourceCache.Insert(objIdentification)
			r.ReferenceGrants = append(r.ReferenceGrants, res)
			return true
		}
	case *corev1.Namespace:
		if !resourceCache.Has(objIdentification) {
			resourceCache.Insert(objIdentification)
			r.Namespaces = append(r.Namespaces, res)
			return true
		}
	case *corev1.Service:
		if !resourceCache.Has(objIdentification) {
			resourceCache.Insert(objIdentification)
			r.Services = append(r.Services, res)
			return true
		}
	case *mcsapiv1a1.ServiceImport:
		if !resourceCache.Has(objIdentification) {
			resourceCache.Insert(objIdentification)
			r.ServiceImports = append(r.ServiceImports, res)
			return true
		}
	case *discoveryv1.EndpointSlice:
		if !resourceCache.Has(objIdentification) {
			resourceCache.Insert(objIdentification)
			r.EndpointSlices = append(r.EndpointSlices, res)
			return true
		}
	case *corev1.Secret:
		if !resourceCache.Has(objIdentification) {
			resourceCache.Insert(objIdentification)
			r.Secrets = append(r.Secrets, res)
			return true
		}
	case *corev1.ConfigMap:
		if !resourceCache.Has(objIdentification) {
			resourceCache.Insert(objIdentification)
			r.ConfigMaps = append(r.ConfigMaps, res)
			return true
		}
	case *egv1a1.EnvoyPatchPolicy:
		if !resourceCache.Has(objIdentification) {
			resourceCache.Insert(objIdentification)
			r.EnvoyPatchPolicies = append(r.EnvoyPatchPolicies, res)
			return true
		}
	case *egv1a1.ClientTrafficPolicy:
		if !resourceCache.Has(objIdentification) {
			resourceCache.Insert(objIdentification)
			r.ClientTrafficPolicies = append(r.ClientTrafficPolicies, res)
			return true
		}
	case *egv1a1.BackendTrafficPolicy:
		if !resourceCache.Has(objIdentification) {
			resourceCache.Insert(objIdentification)
			r.BackendTrafficPolicies = append(r.BackendTrafficPolicies, res)
			return true
		}
	case *egv1a1.SecurityPolicy:
		if !resourceCache.Has(objIdentification) {
			resourceCache.Insert(objIdentification)
			r.SecurityPolicies = append(r.SecurityPolicies, res)
			return true
		}
	case *gwapiv1a3.BackendTLSPolicy:
		if !resourceCache.Has(objIdentification) {
			resourceCache.Insert(objIdentification)
			r.BackendTLSPolicies = append(r.BackendTLSPolicies, res)
			return true
		}
	case *egv1a1.EnvoyExtensionPolicy:
		if !resourceCache.Has(objIdentification) {
			resourceCache.Insert(objIdentification)
			r.EnvoyExtensionPolicies = append(r.EnvoyExtensionPolicies, res)
			return true
		}
	case *egv1a1.HTTPRouteFilter:
		if !resourceCache.Has(objIdentification) {
			resourceCache.Insert(objIdentification)
			r.HTTPRouteFilters = append(r.HTTPRouteFilters, res)
			return true
		}
	}

	return false
}

func (r *Resources) AppendClientTrafficPolicies(clientTrafficPolicies ...*egv1a1.ClientTrafficPolicy) {
	for _, policy := range clientTrafficPolicies {
		r.AppendResource(policy)
	}
}

func (r *Resources) AppendBackendTrafficPolicies(backendTrafficPolicies ...*egv1a1.BackendTrafficPolicy) {
	for _, policy := range backendTrafficPolicies {
		r.AppendResource(policy)
	}
}

func (r *Resources) AppendSecurityPolicies(securityPolicies ...*egv1a1.SecurityPolicy) {
	for _, policy := range securityPolicies {
		r.AppendResource(policy)
	}
}

func (r *Resources) AppendBackendTLSPolicies(backendTLSPolicies ...*gwapiv1a3.BackendTLSPolicy) {
	for _, policy := range backendTLSPolicies {
		r.AppendResource(policy)
	}
}

func (r *Resources) AppendEnvoyExtensionPolicies(envoyExtensionPolicies ...*egv1a1.EnvoyExtensionPolicy) {
	for _, policy := range envoyExtensionPolicies {
		r.AppendResource(policy)
	}
}

func (r *Resources) GetNamespace(name string) *corev1.Namespace {
	for _, ns := range r.Namespaces {
		if ns.Name == name {
			return ns
		}
	}

	return nil
}

func (r *Resources) GetEnvoyProxy(namespace, name string) *egv1a1.EnvoyProxy {
	for _, ep := range r.EnvoyProxiesForGateways {
		if ep.Namespace == namespace && ep.Name == name {
			return ep
		}
	}

	return nil
}

func (r *Resources) GetService(namespace, name string) *corev1.Service {
	for _, svc := range r.Services {
		if svc.Namespace == namespace && svc.Name == name {
			return svc
		}
	}

	return nil
}

func (r *Resources) GetServiceImport(namespace, name string) *mcsapiv1a1.ServiceImport {
	for _, svcImp := range r.ServiceImports {
		if svcImp.Namespace == namespace && svcImp.Name == name {
			return svcImp
		}
	}

	return nil
}

func (r *Resources) GetBackend(namespace, name string) *egv1a1.Backend {
	for _, be := range r.Backends {
		if be.Namespace == namespace && be.Name == name {
			return be
		}
	}

	return nil
}

func (r *Resources) GetSecret(namespace, name string) *corev1.Secret {
	for _, secret := range r.Secrets {
		if secret.Namespace == namespace && secret.Name == name {
			return secret
		}
	}

	return nil
}

func (r *Resources) GetConfigMap(namespace, name string) *corev1.ConfigMap {
	for _, configMap := range r.ConfigMaps {
		if configMap.Namespace == namespace && configMap.Name == name {
			return configMap
		}
	}

	return nil
}

func (r *Resources) GetEndpointSlicesForBackend(svcNamespace, svcName string, backendKind string) []*discoveryv1.EndpointSlice {
	var endpointSlices []*discoveryv1.EndpointSlice
	for _, endpointSlice := range r.EndpointSlices {
		var backendSelectorLabel string
		switch backendKind {
		case KindService:
			backendSelectorLabel = discoveryv1.LabelServiceName
		case KindServiceImport:
			backendSelectorLabel = mcsapiv1a1.LabelServiceName
		}
		if svcNamespace == endpointSlice.Namespace &&
			endpointSlice.GetLabels()[backendSelectorLabel] == svcName {
			endpointSlices = append(endpointSlices, endpointSlice)
		}
	}
	return endpointSlices
}
