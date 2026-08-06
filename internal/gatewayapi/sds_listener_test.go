// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestIRTLSConfigsForSDSListenerCertificate(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "listener-sds"},
		Type:       egv1a1.SDSSecretType,
		Data: map[string][]byte{
			"secretName": []byte("listener-certificate"),
			"url":        []byte("/var/run/sds.sock"),
		},
	}

	tls := irTLSConfigs(&ListenerTLSConfig{secrets: []*corev1.Secret{secret}})
	require.Len(t, tls.Certificates, 1)
	require.Equal(t, &ir.SDSConfig{SecretName: "listener-certificate", URL: "/var/run/sds.sock"}, tls.Certificates[0].SDS)
	require.Equal(t, "default/listener-sds", tls.Certificates[0].Name)
}
