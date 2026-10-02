// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

const (
	testCertGroup = "cert.example.io"
	testCertKind  = "ExampleCertificate"
	testCertID    = "example-cert-id"
)

func extensionCertificate(namespace, name string, ready *bool, reason string) unstructured.Unstructured {
	obj := map[string]interface{}{
		"apiVersion": testCertGroup + "/v1alpha1",
		"kind":       testCertKind,
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"certificateID": testCertID,
		},
	}

	if ready != nil {
		condStatus := string(metav1.ConditionFalse)
		if *ready {
			condStatus = string(metav1.ConditionTrue)
		}
		obj["status"] = map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":    "Ready",
					"status":  condStatus,
					"reason":  reason,
					"message": "from the provider",
				},
			},
		}
	}

	return unstructured.Unstructured{Object: obj}
}

// listenerWithExtensionCertRef builds a single HTTPS listener whose only certificate ref
// points at an extension-registered kind.
func listenerWithExtensionCertRef(t *testing.T, namespace string) *ListenerContext {
	t.Helper()

	group := gwapiv1.Group(testCertGroup)
	kind := gwapiv1.Kind(testCertKind)
	gw := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: namespace},
		Spec: gwapiv1.GatewaySpec{
			Listeners: []gwapiv1.Listener{{
				Name:     "https",
				Protocol: gwapiv1.HTTPSProtocolType,
				Port:     443,
				TLS: &gwapiv1.ListenerTLSConfig{
					Mode: ptr.To(gwapiv1.TLSModeTerminate),
					CertificateRefs: []gwapiv1.SecretObjectReference{{
						Group: &group,
						Kind:  &kind,
						Name:  "app-cert",
					}},
				},
			}},
		},
	}

	gwCtx := &GatewayContext{Gateway: gw}
	gwCtx.ResetListeners()
	require.Len(t, gwCtx.listeners, 1)
	return gwCtx.listeners[0]
}

// A certificate ref to a registered extension kind is admitted when the object exists,
// whatever its status, and the resolved certificate carries no key material.
func TestExtensionCertificateRefAdmission(t *testing.T) {
	translator := &Translator{
		ExtensionCertificateGroupKinds: []schema.GroupKind{
			{Group: testCertGroup, Kind: testCertKind},
		},
	}

	tests := []struct {
		name       string
		cert       *unstructured.Unstructured
		wantOK     bool
		wantReason gwapiv1.ListenerConditionReason
	}{
		{
			name:   "ready certificate is admitted",
			cert:   ptr.To(extensionCertificate("default", "app-cert", ptr.To(true), "Associated")),
			wantOK: true,
		},
		{
			// Readiness is the extension's decision at xDS time, so status is not read here.
			name:   "not ready certificate is still admitted",
			cert:   ptr.To(extensionCertificate("default", "app-cert", ptr.To(false), "AssociationInProgress")),
			wantOK: true,
		},
		{
			name:   "certificate with no status is still admitted",
			cert:   ptr.To(extensionCertificate("default", "app-cert", nil, "")),
			wantOK: true,
		},
		{
			name:       "absent certificate is an invalid ref",
			cert:       nil,
			wantOK:     false,
			wantReason: gwapiv1.ListenerReasonInvalidCertificateRef,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := &resource.Resources{}
			if tc.cert != nil {
				res.ExtensionCertificates = []unstructured.Unstructured{*tc.cert}
			}

			listener := listenerWithExtensionCertRef(t, "default")
			secrets, extCerts, _, ok := translator.validateTerminateModeAndGetTLSSecrets(listener, res)

			require.Equal(t, tc.wantOK, ok)
			require.Empty(t, secrets, "an extension certificate must not appear as a Secret")

			if !tc.wantOK {
				require.Empty(t, extCerts)
				var reasons []string
				for _, c := range listener.GetConditions() {
					reasons = append(reasons, c.Reason)
				}
				require.Contains(t, reasons, string(tc.wantReason))
				return
			}

			require.Len(t, extCerts, 1)
			listener.SetTLSExtensionCertificates(extCerts)

			// The IR must carry a reference, never key material.
			irTLS := irTLSConfigs(&listener.tls)
			require.NotNil(t, irTLS)
			require.Len(t, irTLS.Certificates, 1)
			cert := irTLS.Certificates[0]
			require.NotNil(t, cert.ExtensionRef)
			require.Nil(t, cert.SDS)
			require.Empty(t, cert.Certificate)
			require.Empty(t, cert.PrivateKey)
			require.NoError(t, cert.Validate())

			id, found, err := unstructured.NestedString(cert.ExtensionRef.Object.Object, "spec", "certificateID")
			require.NoError(t, err)
			require.True(t, found)
			require.Equal(t, testCertID, id)
		})
	}
}

// A ref to an unregistered group and kind must still be rejected, so the relaxed gate does
// not become a hole for arbitrary kinds.
func TestUnregisteredCertificateKindStillRejected(t *testing.T) {
	translator := &Translator{}

	listener := listenerWithExtensionCertRef(t, "default")
	res := &resource.Resources{
		ExtensionCertificates: []unstructured.Unstructured{
			extensionCertificate("default", "app-cert", ptr.To(true), "Associated"),
		},
	}

	_, extCerts, _, ok := translator.validateTerminateModeAndGetTLSSecrets(listener, res)
	require.False(t, ok)
	require.Empty(t, extCerts)
}
