// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func TestExtensionBackendReferencesAndIndex(t *testing.T) {
	policy := &egv1a1.EnvoyExtensionPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "extension", Namespace: "default"},
		Spec: egv1a1.EnvoyExtensionPolicySpec{Backends: []egv1a1.ExtensionBackend{
			{Name: "local", BackendRef: gwapiv1.BackendObjectReference{Name: "local-service", Port: new(gwapiv1.PortNumber(8080))}},
			{Name: "remote", BackendRef: gwapiv1.BackendObjectReference{
				Group: new(gwapiv1.Group(egv1a1.GroupName)), Kind: new(gwapiv1.Kind(egv1a1.KindBackend)),
				Namespace: new(gwapiv1.Namespace("other")), Name: "remote-backend",
			}},
		}},
	}
	backend := &egv1a1.Backend{ObjectMeta: metav1.ObjectMeta{Name: "remote-backend", Namespace: "other"}}
	grant := &gwapiv1b1.ReferenceGrant{
		ObjectMeta: metav1.ObjectMeta{Name: "allow-extension", Namespace: "other"},
		Spec: gwapiv1b1.ReferenceGrantSpec{
			From: []gwapiv1b1.ReferenceGrantFrom{{Group: egv1a1.GroupName, Kind: egv1a1.KindEnvoyExtensionPolicy, Namespace: "default"}},
			To:   []gwapiv1b1.ReferenceGrantTo{{Group: egv1a1.GroupName, Kind: egv1a1.KindBackend}},
		},
	}
	require.ElementsMatch(t, []string{"default/local-service", "other/remote-backend"}, backendEnvoyExtensionPolicyIndexFunc(policy))
	r := setupReferenceGrantReconciler([]client.Object{policy, backend, grant})
	resources := resource.NewResources()
	resources.EnvoyExtensionPolicies = []*egv1a1.EnvoyExtensionPolicy{policy}
	mapping := newResourceMapping()
	require.NoError(t, r.processEnvoyExtensionPolicyObjectRefs(t.Context(), resources, mapping))
	require.Len(t, mapping.allAssociatedBackendRefs, 2)
	for _, backendRef := range mapping.allAssociatedBackendRefs {
		if backendRef.Name == "remote-backend" {
			require.Equal(t, "other", string(*backendRef.Namespace))
		} else {
			require.Equal(t, "local-service", string(backendRef.Name))
			require.Equal(t, "default", string(*backendRef.Namespace))
		}
	}
	require.Contains(t, resources.ReferenceGrants, grant)
}
