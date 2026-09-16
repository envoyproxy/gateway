// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	tlsutils "sigs.k8s.io/gateway-api/conformance/utils/tls"
)

func init() {
	ConformanceTests = append(ConformanceTests, SDS)
}

var SDS = suite.ConformanceTest{
	ShortName:   "SDS",
	Description: "Test SDS server providing secrets to Envoy, for both backend TLS policy client certificates and HTTPS listener certificates",
	Manifests: []string{
		"testdata/sds.yaml",
		"testdata/sds-listener-certificate.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("backend TLS policy client certificate", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			acceptedCond := metav1.Condition{
				Type:   string(gwapiv1.PolicyConditionAccepted),
				Status: metav1.ConditionTrue,
				Reason: string(gwapiv1.PolicyReasonAccepted),
			}
			resolvedRefsCond := metav1.Condition{
				Type:   string(gwapiv1.BackendTLSPolicyConditionResolvedRefs),
				Status: metav1.ConditionTrue,
				Reason: string(gwapiv1.BackendTLSPolicyReasonResolvedRefs),
			}

			routeNN := types.NamespacedName{Name: "sds-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "sds-gateway", Namespace: ns}
			validPolicyNN := types.NamespacedName{Name: "tls-backend-policy", Namespace: ns}
			kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, validPolicyNN, gwNN, acceptedCond)
			kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, validPolicyNN, gwNN, resolvedRefsCond)
			gwAddr := GatewayAndTCPRoutesMustBeAccepted(t, suite.Client, &suite.TimeoutConfig, suite.ControllerName, NewGatewayRef(gwNN), routeNN)

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			})
		})

		t.Run("HTTPS listener certificate", func(t *testing.T) {
			const (
				namespace  = "gateway-conformance-infra"
				serverName = "sds.example.com"
			)
			gatewayName := types.NamespacedName{Name: "sds-listener-certificate", Namespace: namespace}
			routeName := types.NamespacedName{Name: "sds-listener-certificate", Namespace: namespace}

			// Given an accepted HTTPS Gateway whose certificate reference is backed by SDS.
			gatewayHost := kubernetes.GatewayAndRoutesMustBeAccepted(
				t,
				suite.Client,
				suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(gatewayName),
				&gwapiv1.HTTPRoute{},
				false,
				routeName,
			)
			certificatePEM, _, _, err := GetTLSSecret(
				suite.Client,
				types.NamespacedName{Name: "sds-listener-certificate-source", Namespace: namespace},
			)
			require.NoError(t, err)
			gwPodNamespace := GetGatewayResourceNamespace()
			WaitForPods(
				t,
				suite.Client,
				gwPodNamespace,
				map[string]string{
					"gateway.envoyproxy.io/owning-gateway-name":      gatewayName.Name,
					"gateway.envoyproxy.io/owning-gateway-namespace": gatewayName.Namespace,
				},
				corev1.PodRunning,
				&PodReady,
			)

			// When traffic is sent through the SDS-backed HTTPS listener.
			expected := http.ExpectedResponse{
				Request: http.Request{Host: serverName, Path: "/sds-listener"},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: namespace,
			}
			tlsutils.MakeTLSRequestAndExpectEventuallyConsistentResponse(
				t,
				suite.RoundTripper,
				suite.TimeoutConfig,
				gatewayHost,
				certificatePEM,
				nil,
				nil,
				serverName,
				expected,
			)

			// Then Envoy presents the exact certificate delivered by the SDS server.
			expectedCertificate := parseCertificate(t, certificatePEM)
			rootCAs := x509.NewCertPool()
			require.True(t, rootCAs.AppendCertsFromPEM(certificatePEM))
			tlsConfig := &tls.Config{
				MinVersion: tls.VersionTLS12,
				RootCAs:    rootCAs,
				ServerName: serverName,
			}
			connection, err := tls.DialWithDialer(
				&net.Dialer{Timeout: 10 * time.Second},
				"tcp",
				net.JoinHostPort(gatewayHost, "443"),
				tlsConfig,
			)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, connection.Close()) })
			require.NotEmpty(t, connection.ConnectionState().PeerCertificates)
			require.True(t, bytes.Equal(expectedCertificate.Raw, connection.ConnectionState().PeerCertificates[0].Raw))
		})
	},
}

func parseCertificate(t *testing.T, certificatePEM []byte) *x509.Certificate {
	t.Helper()
	block, _ := pem.Decode(certificatePEM)
	require.NotNil(t, block)
	certificate, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)
	return certificate
}
