// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package client

import (
	"testing"

	"k8s.io/client-go/rest"

	"github.com/envoyproxy/gateway/pkg/client/clientset/versioned"
	"github.com/envoyproxy/gateway/pkg/client/clientset/versioned/fake"
)

// TestNewForConfig ensures the clientset can be created with a valid config.
func TestNewForConfig(t *testing.T) {
	config := &rest.Config{
		Host: "https://localhost:6443",
	}

	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		t.Fatalf("Failed to create clientset: %v", err)
	}

	if clientset == nil {
		t.Fatal("Expected non-nil clientset")
	}

	// Verify all resource clients are accessible
	if clientset.GatewayV1alpha1() == nil {
		t.Error("Expected non-nil GatewayV1alpha1 client")
	}
}

// TestFakeClientset ensures the fake clientset can be created.
func TestFakeClientset(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	if clientset == nil {
		t.Fatal("Expected non-nil fake clientset")
	}

	// Verify all resource clients are accessible
	if clientset.GatewayV1alpha1() == nil {
		t.Error("Expected non-nil GatewayV1alpha1 client")
	}

	// Verify we can access individual resource clients
	v1alpha1 := clientset.GatewayV1alpha1()

	if v1alpha1.Backends("default") == nil {
		t.Error("Expected non-nil Backends client")
	}
	if v1alpha1.BackendTrafficPolicies("default") == nil {
		t.Error("Expected non-nil BackendTrafficPolicies client")
	}
	if v1alpha1.ClientTrafficPolicies("default") == nil {
		t.Error("Expected non-nil ClientTrafficPolicies client")
	}
	if v1alpha1.EnvoyExtensionPolicies("default") == nil {
		t.Error("Expected non-nil EnvoyExtensionPolicies client")
	}
	if v1alpha1.EnvoyPatchPolicies("default") == nil {
		t.Error("Expected non-nil EnvoyPatchPolicies client")
	}
	if v1alpha1.EnvoyProxies("default") == nil {
		t.Error("Expected non-nil EnvoyProxies client")
	}
	if v1alpha1.HTTPRouteFilters("default") == nil {
		t.Error("Expected non-nil HTTPRouteFilters client")
	}
	if v1alpha1.SecurityPolicies("default") == nil {
		t.Error("Expected non-nil SecurityPolicies client")
	}
}
