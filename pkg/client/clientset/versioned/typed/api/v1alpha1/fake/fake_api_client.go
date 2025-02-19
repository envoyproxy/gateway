// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/envoyproxy/gateway/pkg/client/clientset/versioned/typed/api/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeEnvoyGatewayV1alpha1 struct {
	*testing.Fake
}

func (c *FakeEnvoyGatewayV1alpha1) Backends(namespace string) v1alpha1.BackendInterface {
	return newFakeBackends(c, namespace)
}

func (c *FakeEnvoyGatewayV1alpha1) BackendTrafficPolicies(namespace string) v1alpha1.BackendTrafficPolicyInterface {
	return newFakeBackendTrafficPolicies(c, namespace)
}

func (c *FakeEnvoyGatewayV1alpha1) ClientTrafficPolicies(namespace string) v1alpha1.ClientTrafficPolicyInterface {
	return newFakeClientTrafficPolicies(c, namespace)
}

func (c *FakeEnvoyGatewayV1alpha1) EnvoyExtensionPolicies(namespace string) v1alpha1.EnvoyExtensionPolicyInterface {
	return newFakeEnvoyExtensionPolicies(c, namespace)
}

func (c *FakeEnvoyGatewayV1alpha1) EnvoyPatchPolicies(namespace string) v1alpha1.EnvoyPatchPolicyInterface {
	return newFakeEnvoyPatchPolicies(c, namespace)
}

func (c *FakeEnvoyGatewayV1alpha1) EnvoyProxies(namespace string) v1alpha1.EnvoyProxyInterface {
	return newFakeEnvoyProxies(c, namespace)
}

func (c *FakeEnvoyGatewayV1alpha1) HTTPRouteFilters(namespace string) v1alpha1.HTTPRouteFilterInterface {
	return newFakeHTTPRouteFilters(c, namespace)
}

func (c *FakeEnvoyGatewayV1alpha1) SecurityPolicies(namespace string) v1alpha1.SecurityPolicyInterface {
	return newFakeSecurityPolicies(c, namespace)
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeEnvoyGatewayV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
