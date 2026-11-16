// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	testCertGroup = "cert.example.io"
	testCertKind  = "ExampleCertificate"
	testCertID    = "example-cert-id"
)

func extensionCertificate(ready *bool, reason string) *unstructured.Unstructured {
	obj := map[string]interface{}{
		"apiVersion": testCertGroup + "/v1alpha1",
		"kind":       testCertKind,
		"metadata": map[string]interface{}{
			"name":      "app-cert",
			"namespace": "default",
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

	return &unstructured.Unstructured{Object: obj}
}

// listenerWithExtensionCertRef builds a single HTTPS listener whose only certificate ref
// points at an extension-registered kind.
func listenerWithExtensionCertRef(t *testing.T, namespace string) *ListenerContext {
	t.Helper()

	group := gwapiv1.Group(testCertGroup)
	kind := gwapiv1.Kind(testCertKind)
	mode := gwapiv1.TLSModeTerminate
	gw := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: namespace},
		Spec: gwapiv1.GatewaySpec{
			Listeners: []gwapiv1.Listener{{
				Name:     "https",
				Protocol: gwapiv1.HTTPSProtocolType,
				Port:     443,
				TLS: &gwapiv1.ListenerTLSConfig{
					Mode: &mode,
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
		ExtensionSecretGroupKinds: []schema.GroupKind{
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
			cert:   extensionCertificate(new(true), "Associated"),
			wantOK: true,
		},
		{
			// Readiness is the extension's decision at xDS time, so status is not read here.
			name:   "not ready certificate is still admitted",
			cert:   extensionCertificate(new(false), "AssociationInProgress"),
			wantOK: true,
		},
		{
			name:   "certificate with no status is still admitted",
			cert:   extensionCertificate(nil, ""),
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
				res.ExtensionSecrets = []unstructured.Unstructured{*tc.cert}
			}

			listener := listenerWithExtensionCertRef(t, "default")
			secrets, _, ok := translator.validateTerminateModeAndGetTLSSecrets(listener, res)

			require.Equal(t, tc.wantOK, ok)

			if !tc.wantOK {
				require.Empty(t, secrets)
				reasons := make([]string, 0, len(listener.GetConditions()))
				for _, c := range listener.GetConditions() {
					reasons = append(reasons, c.Reason)
				}
				require.Contains(t, reasons, string(tc.wantReason))
				return
			}

			require.Len(t, secrets, 1)
			require.NotNil(t, secrets[0].extensionSecret)
			require.Nil(t, secrets[0].kubernetesSecret, "an extension certificate must not appear as a Secret")
			listener.SetTLSSecrets(secrets)

			// The IR must carry a reference, never key material.
			irTLS := translator.irTLSConfigs(&listener.tls, nil, nil)
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
		ExtensionSecrets: []unstructured.Unstructured{
			*extensionCertificate(new(true), "Associated"),
		},
	}

	secrets, _, ok := translator.validateTerminateModeAndGetTLSSecrets(listener, res)
	require.False(t, ok)
	require.Empty(t, secrets)
}

// Extension certificates keep their configured position among secrets in the IR certificate list.
func TestExtensionCertificateOrder(t *testing.T) {
	sdsSecret := func(name string) *corev1.Secret {
		return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
			Type:       egv1a1.SDSSecretType,
			Data: map[string][]byte{
				"secretName": []byte(name),
				"url":        []byte("unix:///var/run/sds/" + name + ".sock"),
			},
		}
	}
	tlsSecret := func(name string, cert, key []byte) *corev1.Secret {
		return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
			Type:       corev1.SecretTypeTLS,
			Data:       map[string][]byte{corev1.TLSCertKey: cert, corev1.TLSPrivateKeyKey: key},
		}
	}
	chain, key := buildServingChain(t, time.Now().Add(24*time.Hour), time.Now().Add(24*time.Hour))
	otherCert := extensionCertificate(nil, "")
	otherCert.SetName("other-cert")

	group := gwapiv1.Group(testCertGroup)
	kind := gwapiv1.Kind(testCertKind)
	extensionRef := func(name gwapiv1.ObjectName) gwapiv1.SecretObjectReference {
		return gwapiv1.SecretObjectReference{Group: &group, Kind: &kind, Name: name}
	}

	tests := []struct {
		name string
		refs []gwapiv1.SecretObjectReference
		want []string
	}{
		{
			name: "extension certificate first",
			refs: []gwapiv1.SecretObjectReference{extensionRef("app-cert"), {Name: "sds-one"}},
			want: []string{"app-cert", "sds-one"},
		},
		{
			name: "extension certificate between secrets",
			refs: []gwapiv1.SecretObjectReference{{Name: "sds-one"}, extensionRef("app-cert"), {Name: "sds-two"}},
			want: []string{"sds-one", "app-cert", "sds-two"},
		},
		{
			name: "extension certificate last",
			refs: []gwapiv1.SecretObjectReference{{Name: "sds-one"}, {Name: "sds-two"}, extensionRef("app-cert")},
			want: []string{"sds-one", "sds-two", "app-cert"},
		},
		{
			name: "extension certificate before a TLS secret",
			refs: []gwapiv1.SecretObjectReference{extensionRef("app-cert"), {Name: "tls-one"}},
			want: []string{"app-cert", "tls-one"},
		},
		{
			name: "several extension certificates interleaved with secrets",
			refs: []gwapiv1.SecretObjectReference{
				extensionRef("app-cert"), {Name: "tls-one"}, extensionRef("other-cert"), {Name: "sds-one"},
			},
			want: []string{"app-cert", "tls-one", "other-cert", "sds-one"},
		},
		{
			name: "missing secret before an extension certificate",
			refs: []gwapiv1.SecretObjectReference{{Name: "missing"}, extensionRef("app-cert"), {Name: "sds-one"}},
			want: []string{"app-cert", "sds-one"},
		},
		{
			name: "invalid TLS secret before an extension certificate",
			refs: []gwapiv1.SecretObjectReference{{Name: "tls-invalid"}, extensionRef("app-cert"), {Name: "tls-one"}},
			want: []string{"app-cert", "tls-one"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mode := gwapiv1.TLSModeTerminate
			gwCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "default"},
				Spec: gwapiv1.GatewaySpec{Listeners: []gwapiv1.Listener{{
					Name:     "https",
					Protocol: gwapiv1.HTTPSProtocolType,
					Port:     443,
					TLS:      &gwapiv1.ListenerTLSConfig{Mode: &mode, CertificateRefs: tc.refs},
				}}},
			}}
			gwCtx.ResetListeners()
			listener := gwCtx.listeners[0]

			translator := &Translator{
				TranslatorContext:   &TranslatorContext{},
				SDSSecretRefEnabled: true,
				ExtensionSecretGroupKinds: []schema.GroupKind{
					{Group: testCertGroup, Kind: testCertKind},
				},
			}
			translator.SetSecrets([]*corev1.Secret{
				sdsSecret("sds-one"),
				sdsSecret("sds-two"),
				tlsSecret("tls-one", chain, key),
				tlsSecret("tls-invalid", []byte("not a certificate"), []byte("not a key")),
			})
			res := &resource.Resources{
				ExtensionSecrets: []unstructured.Unstructured{*extensionCertificate(nil, ""), *otherCert},
			}

			secrets, _, ok := translator.validateTerminateModeAndGetTLSSecrets(listener, res)
			require.True(t, ok)
			listener.SetTLSSecrets(secrets)

			irTLS := translator.irTLSConfigs(&listener.tls, &ir.Xds{}, gwCtx)
			require.NotNil(t, irTLS)
			got := make([]string, 0, len(irTLS.Certificates))
			for _, cert := range irTLS.Certificates {
				got = append(got, path.Base(cert.Name))
			}
			require.Equal(t, tc.want, got)
		})
	}
}

// Extension-resolved certificates are stored once in the gateway's IR registry and referenced by
// name, so a certificate shared by several listeners is not copied into each of them.
func TestExtensionCertificatesInRegistry(t *testing.T) {
	translator := &Translator{TranslatorContext: &TranslatorContext{}}
	gwIR := &ir.Xds{}
	gwCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "gw"}}}
	cert := extensionCertificate(nil, "")

	listenerTLS := func() *ir.TLSConfig {
		return translator.irTLSConfigs(&ListenerTLSConfig{secrets: []listenerSecret{{extensionSecret: cert}}}, gwIR, gwCtx)
	}
	first, second := listenerTLS(), listenerTLS()

	require.Len(t, gwIR.ExtensionResources, 1)
	require.Equal(t, cert.GetName(), gwIR.ExtensionResources[0].Object.GetName())
	for _, tlsConfig := range []*ir.TLSConfig{first, second} {
		ref := tlsConfig.Certificates[0].ExtensionRef
		require.Nil(t, ref.Object)
		require.Equal(t, gwIR.ExtensionResources[0].Name, ref.Name)
		require.NoError(t, tlsConfig.Certificates[0].Validate())
	}
}
