// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	certificatesv1b1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/utils/naming"
)

func caRef(kind, namespace, name string) gwapiv1.ObjectReference {
	ref := gwapiv1.ObjectReference{Kind: gwapiv1.Kind(kind), Name: gwapiv1.ObjectName(name)}
	if namespace != "" {
		ns := gwapiv1.Namespace(namespace)
		ref.Namespace = &ns
	}
	return ref
}

func TestUpstreamCASecretName(t *testing.T) {
	cases := []struct {
		name             string
		refs             []gwapiv1.ObjectReference
		defaultNamespace string
		expected         string
	}{
		{
			name:             "configmap uses kind/namespace/name",
			refs:             []gwapiv1.ObjectReference{caRef(resource.KindConfigMap, "", "ca-cmap")},
			defaultNamespace: "policies",
			expected:         "configmap/policies/ca-cmap",
		},
		{
			name:             "secret uses kind/namespace/name",
			refs:             []gwapiv1.ObjectReference{caRef(resource.KindSecret, "", "ca-secret")},
			defaultNamespace: "policies",
			expected:         "secret/policies/ca-secret",
		},
		{
			name:             "cluster trust bundle is cluster scoped so has no namespace segment",
			refs:             []gwapiv1.ObjectReference{caRef(resource.KindClusterTrustBundle, "", "my-roots")},
			defaultNamespace: "policies",
			expected:         "clustertrustbundle/my-roots",
		},
		{
			name:             "explicit ref namespace wins over the policy namespace",
			refs:             []gwapiv1.ObjectReference{caRef(resource.KindConfigMap, "other", "ca-cmap")},
			defaultNamespace: "policies",
			expected:         "configmap/other/ca-cmap",
		},
		{
			name: "multiple refs join in declared order",
			refs: []gwapiv1.ObjectReference{
				caRef(resource.KindConfigMap, "", "ca1"),
				caRef(resource.KindSecret, "", "ca2"),
				caRef(resource.KindClusterTrustBundle, "", "roots"),
			},
			defaultNamespace: "policies",
			expected:         "configmap/policies/ca1,secret/policies/ca2,clustertrustbundle/roots",
		},
		{
			name:             "unsupported kinds contribute nothing",
			refs:             []gwapiv1.ObjectReference{caRef("Service", "", "svc")},
			defaultNamespace: "policies",
			expected:         "",
		},
		{
			name:             "no refs",
			refs:             nil,
			defaultNamespace: "policies",
			expected:         "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, upstreamCASecretName(c.refs, c.defaultNamespace))
		})
	}
}

// A Secret and a ConfigMap of the same name in one namespace are different objects holding
// different bytes. Without the kind segment they would share an SDS secret name and one would
// silently win.
func TestUpstreamCASecretNameDistinguishesKinds(t *testing.T) {
	cm := upstreamCASecretName([]gwapiv1.ObjectReference{caRef(resource.KindConfigMap, "", "ca")}, "ns")
	secret := upstreamCASecretName([]gwapiv1.ObjectReference{caRef(resource.KindSecret, "", "ca")}, "ns")
	assert.NotEqual(t, cm, secret)
}

// The bundle is concatenated in ref order, so a reordered list is genuinely different bytes and
// must not share a name.
func TestUpstreamCASecretNameIsOrderSensitive(t *testing.T) {
	a := caRef(resource.KindConfigMap, "", "ca1")
	b := caRef(resource.KindSecret, "", "ca2")
	assert.NotEqual(t,
		upstreamCASecretName([]gwapiv1.ObjectReference{a, b}, "ns"),
		upstreamCASecretName([]gwapiv1.ObjectReference{b, a}, "ns"))
}

func TestUpstreamCASecretNameIsBounded(t *testing.T) {
	refs := make([]gwapiv1.ObjectReference, 0, 64)
	for i := range 64 {
		refs = append(refs, caRef(resource.KindConfigMap, "a-long-namespace-name", fmt.Sprintf("ca-bundle-%d", i)))
	}
	name := upstreamCASecretName(refs, "ns")
	require.LessOrEqual(t, len(name), maxUpstreamCASecretNameBytes)
	assert.Contains(t, name, "-", "a truncated name carries a hash suffix")

	// Two long lists sharing a truncated head must stay distinct.
	other := append(append([]gwapiv1.ObjectReference{}, refs...), caRef(resource.KindConfigMap, "ns", "tail"))
	assert.NotEqual(t, name, upstreamCASecretName(other, "ns"))
}

// caRefNameSegment and the switch in getCaCertsFromCARefs must agree on which kinds contribute.
// If a kind is added to supportedCAKind without a matching case in the byte loop, it earns a
// name segment while contributing nothing, splitting one logical bundle across two SDS secrets.
func TestCANameSegmentMatchesSupportedKinds(t *testing.T) {
	const ns = "policies"
	kinds := []string{
		resource.KindConfigMap, resource.KindSecret, resource.KindClusterTrustBundle,
		resource.KindService, resource.KindGateway, "", "NotAKind",
	}

	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			tr := &Translator{}
			tr.TranslatorContext = &TranslatorContext{
				SecretMap: map[types.NamespacedName]*corev1.Secret{
					{Namespace: ns, Name: "ca"}: {
						ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "ca"},
						Data:       map[string][]byte{CACertKey: []byte("x")},
					},
				},
				ConfigMapMap: map[types.NamespacedName]*corev1.ConfigMap{
					{Namespace: ns, Name: "ca"}: {
						ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "ca"},
						Data:       map[string]string{CACertKey: "x"},
					},
				},
				ClusterTrustBundleMap: map[types.NamespacedName]*certificatesv1b1.ClusterTrustBundle{
					{Name: "ca"}: {
						ObjectMeta: metav1.ObjectMeta{Name: "ca"},
						Spec:       certificatesv1b1.ClusterTrustBundleSpec{TrustBundle: "x"},
					},
				},
			}

			named := caRefNameSegment(kind, ns, "ca") != ""
			caCert, _, err := tr.getCaCertsFromCARefs(nil,
				[]gwapiv1.ObjectReference{caRef(kind, ns, "ca")},
				resource.ResourceMetadata{Namespace: ns, Name: "p", Kind: resource.KindBackendTLSPolicy})
			produced := err == nil && len(caCert) > 0

			assert.Equal(t, named, produced,
				"kind %q: caRefNameSegment names it (%v) but getCaCertsFromCARefs produced bytes (%v), err=%v",
				kind, named, produced, err)
		})
	}
}

func TestBoundedNameUsedForCASecrets(t *testing.T) {
	// The builder must not shorten a name that already fits, or every real single-ref name
	// would lose its source object.
	short := "configmap/policies/ca-cmap"
	assert.Equal(t, short, naming.Bounded(short, maxUpstreamCASecretNameBytes))
}
