// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build celvalidation

package celvalidation

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestExtensionBackendValidation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		aliases   []string
		wantError string
	}{
		{"valid", []string{"resolver", "tracing", "xds-cluster"}, ""},
		{"absent", nil, ""},
		{"duplicate", []string{"resolver", "resolver"}, "Duplicate value"},
		{"empty", []string{""}, "spec.backends[0].name"},
		{"slash", []string{"a/b"}, "spec.backends[0].name"},
		{"uppercase", []string{"Resolver"}, "spec.backends[0].name"},
		{"too long", []string{strings.Repeat("a", 64)}, "Too long"},
		{"too many", func() []string {
			aliases := make([]string, 17)
			for i := range aliases {
				aliases[i] = fmt.Sprintf("backend-%d", i)
			}
			return aliases
		}(), "Too many"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			policy := &egv1a1.EnvoyExtensionPolicy{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("extension-backends-%d", time.Now().UnixNano()), Namespace: "default"},
				Spec: egv1a1.EnvoyExtensionPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{{
						LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{Group: gwapiv1.GroupName, Kind: "HTTPRoute", Name: "route"},
					}}},
				},
			}
			for _, alias := range tc.aliases {
				policy.Spec.Backends = append(policy.Spec.Backends, egv1a1.ExtensionBackend{
					Name: alias, BackendRef: gwapiv1.BackendObjectReference{Name: "service", Port: new(gwapiv1.PortNumber(8080))},
				})
			}
			err := c.Create(t.Context(), policy)
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExtensionBackendSettingsValidation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		settings  egv1a1.ClusterSettings
		wantError string
	}{
		{"connection", egv1a1.ClusterSettings{Timeout: &egv1a1.Timeout{TCP: &egv1a1.TCPTimeout{ConnectTimeout: new(gwapiv1.Duration("3s"))}}}, ""},
		{"request", egv1a1.ClusterSettings{Timeout: &egv1a1.Timeout{HTTP: &egv1a1.HTTPTimeout{RequestTimeout: new(gwapiv1.Duration("1s"))}}}, "controlled by the extension"},
		{"stream", egv1a1.ClusterSettings{Timeout: &egv1a1.Timeout{HTTP: &egv1a1.HTTPTimeout{StreamIdleTimeout: new(gwapiv1.Duration("1s"))}}}, "controlled by the extension"},
		{"round robin", egv1a1.ClusterSettings{LoadBalancer: &egv1a1.LoadBalancer{Type: egv1a1.RoundRobinLoadBalancerType}}, ""},
		{"dynamic module", egv1a1.ClusterSettings{LoadBalancer: &egv1a1.LoadBalancer{
			Type:          egv1a1.DynamicModuleLoadBalancerType,
			DynamicModule: &egv1a1.DynamicModuleLBPolicy{Name: "my-auth-module", LBPolicyName: "test"},
		}}, "DynamicModule load balancing is not supported"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			policy := &egv1a1.EnvoyExtensionPolicy{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("backend-settings-%d", time.Now().UnixNano()), Namespace: "default"},
				Spec: egv1a1.EnvoyExtensionPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{{
						LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{Group: gwapiv1.GroupName, Kind: "HTTPRoute", Name: "route"},
					}}},
					Backends: []egv1a1.ExtensionBackend{{Name: "resolver", BackendRef: gwapiv1.BackendObjectReference{Name: "service", Port: new(gwapiv1.PortNumber(8080))}, BackendSettings: &tc.settings}},
				},
			}
			err := c.Create(t.Context(), policy)
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
