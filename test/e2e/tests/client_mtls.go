// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	tlsutils "sigs.k8s.io/gateway-api/conformance/utils/tls"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, ClientMTLSTest, ClientMTLSClusterTrustBundleTest)
}

var ClientMTLSTest = suite.ConformanceTest{
	ShortName:   "ClientMTLS",
	Description: "Use Gateway with Client MTLS policy",
	Manifests:   []string{"testdata/client-mtls.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Use Client MTLS", func(t *testing.T) {
			depNS := "envoy-gateway-system"
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-client-mtls", Namespace: ns}
			gwNN := types.NamespacedName{Name: "client-mtls-gateway", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
			kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{depNS})

			expected := http.ExpectedResponse{
				Request: http.Request{
					Host: "mtls.example.com",
					Path: "/client-mtls",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host: "mtls.example.com",
						Path: "/client-mtls",
						Headers: map[string]string{
							"X-Forwarded-Client-Cert": "Hash=ac77d86dd638969a0a39b4e0743370e860d1b70da58b1b08ce950417b6386a8b;Subject=\"CN=mtls.example.com,OU=Gateway,O=EnvoyProxy,L=SomeCity,ST=VA,C=US\"",
						},
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			// This test uses the same key/cert pair as both a client cert and server cert
			// Both backend and client treat the self-signed cert as a trusted CA
			clientCertificate, clientCertificateKey, _, err := GetTLSSecret(suite.Client, types.NamespacedName{Name: "client-mtls-certificate", Namespace: ns})
			if err != nil {
				t.Fatalf("unexpected error finding TLS secret: %v", err)
			}

			tlsutils.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig,
				gwAddr, clientCertificate, clientCertificate, clientCertificateKey, "mtls.example.com", expected)
		})

		t.Run("Client TLS Settings Enforced", func(t *testing.T) {
			depNS := "envoy-gateway-system"
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-client-tls-settings", Namespace: ns}
			gwNN := types.NamespacedName{Name: "client-mtls-gateway", Namespace: ns}
			gwHost := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
			certNN := types.NamespacedName{Name: "client-tls-settings-certificate", Namespace: ns}
			kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{depNS})

			gwAddr := net.JoinHostPort(gwHost, "443")

			const serverName = "tls-settings.example.com"

			expected := http.ExpectedResponse{
				Request: http.Request{
					Host: serverName,
					Path: "/client-tls-settings",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host: serverName,
						Path: "/client-tls-settings",
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			// added but not used, as these are required by test utils when for SNI to be added
			clientCertificate, clientCertificateKey, _, err := GetTLSSecret(suite.Client, certNN)
			if err != nil {
				t.Fatalf("unexpected error finding TLS secret: %v", err)
			}

			tlsutils.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig,
				gwAddr, clientCertificate, clientCertificate, clientCertificateKey, serverName, expected)

			certPool := x509.NewCertPool()
			if !certPool.AppendCertsFromPEM(clientCertificate) {
				t.Errorf("Error setting Root CAs: %v", err)
			}

			// nolint: gosec
			baseTLSConfig := &tls.Config{
				ServerName: serverName,
				RootCAs:    certPool,
			}

			// Check positive and negative TLS versions
			dialWithTLSVersion(t, gwAddr, baseTLSConfig, tls.VersionTLS10, true)
			dialWithTLSVersion(t, gwAddr, baseTLSConfig, tls.VersionTLS11, false)
			dialWithTLSVersion(t, gwAddr, baseTLSConfig, tls.VersionTLS12, false)
			dialWithTLSVersion(t, gwAddr, baseTLSConfig, tls.VersionTLS13, true)

			// check positive and negative ciphers
			dialWithCipher(t, gwAddr, baseTLSConfig, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, false)
			dialWithCipher(t, gwAddr, baseTLSConfig, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, true)

			// check positive and negative curves
			dialWithCurve(t, gwAddr, baseTLSConfig, tls.CurveP256, false)
			dialWithCurve(t, gwAddr, baseTLSConfig, tls.X25519, false)
			dialWithCurve(t, gwAddr, baseTLSConfig, tls.CurveP521, true)

			// Check ALPN
			dialAndExpectALPN(t, gwAddr, baseTLSConfig, "http/1.1")

			// Check that tickets are not assigned as per EG defaults
			dialAndCheckSessionTicketAssignment(t, gwAddr, baseTLSConfig, 0)
		})
	},
}

var ClientMTLSClusterTrustBundleTest = suite.ConformanceTest{
	ShortName:   "ClientMTLSClusterTrustBundle",
	Description: "Use Gateway with Client MTLS policy",
	Manifests:   []string{"testdata/client-mtls-trustbundle.yaml"},
	Features: []features.FeatureName{
		ClusterTrustBundleFeature,
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Client MTLS with ClusterTrustBundle", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "client-mtls-clustertrustbundle", Namespace: ns}
			gwNN := types.NamespacedName{Name: "client-mtls-clustertrustbundle", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			expected := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/cluster-trust-bundle",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host: "www.example.com",
						Path: "/cluster-trust-bundle",
						Headers: map[string]string{
							"X-Forwarded-Client-Cert": "Hash=42a13e4b02c8a6d2ae5bf2fdaa032e24fdbabbaa79b6017fd0db6c077e6999e0;Subject=\"O=example organization,CN=client.example.com\"",
						},
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			serverCertificate, _, _, err := GetTLSSecret(suite.Client, types.NamespacedName{
				Name:      "example-com-tls",
				Namespace: ns,
			})
			if err != nil {
				t.Fatalf("unexpected error finding TLS secret: %v", err)
			}
			// This test uses the same key/cert pair as both a client cert and server cert
			// Both backend and client treat the self-signed cert as a trusted CA
			clientCertificate, clientCertificateKey, _, err := GetTLSSecret(suite.Client, types.NamespacedName{
				Name:      "client-example-com",
				Namespace: ns,
			})
			if err != nil {
				t.Fatalf("unexpected error finding TLS secret: %v", err)
			}

			tlsutils.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig,
				gwAddr, serverCertificate, clientCertificate, clientCertificateKey, "www.example.com", expected)
		})
	},
}

// GetTLSSecret fetches the named Secret and converts both cert and key to []byte
func GetTLSSecret(client client.Client, secretName types.NamespacedName) ([]byte, []byte, []byte, error) {
	var cert, key, ca []byte

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	secret := &corev1.Secret{}
	err := client.Get(ctx, secretName, secret)
	if err != nil {
		return cert, key, ca, fmt.Errorf("error fetching TLS Secret: %w", err)
	}
	cert = secret.Data["tls.crt"]
	key = secret.Data["tls.key"]
	ca = secret.Data["ca.crt"]

	return cert, key, ca, nil
}

func dialWithTLSVersion(t *testing.T, gwAddr string, baseTLSConfig *tls.Config, version uint16, expectedError bool) {
	tlsConfig := baseTLSConfig.Clone()
	tlsConfig.MinVersion = version
	tlsConfig.MaxVersion = version
	tlsConfig.CipherSuites = []uint16{tls.TLS_RSA_WITH_AES_128_CBC_SHA, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256}

	conn, err := tls.Dial("tcp", gwAddr, tlsConfig)

	if expectedError {
		require.Error(t, err, "protocol version not supported")
	} else { // not error
		require.NoError(t, err)
		require.NotNil(t, conn)
		require.Equal(t, conn.ConnectionState().Version, version)
		_ = conn.Close()
	}
}

func dialAndExpectALPN(t *testing.T, gwAddr string, baseTLSConfig *tls.Config, expectedALPN string) {
	tlsConfig := baseTLSConfig.Clone()
	tlsConfig.NextProtos = []string{"h2", "http/1.1"}

	conn, err := tls.Dial("tcp", gwAddr, tlsConfig)

	require.NoError(t, err)
	require.NotNil(t, conn)
	require.Equal(t, expectedALPN, conn.ConnectionState().NegotiatedProtocol)
	_ = conn.Close()
}

func dialWithCipher(t *testing.T, gwAddr string, baseTLSConfig *tls.Config, cipher uint16, expectedError bool) {
	tlsConfig := baseTLSConfig.Clone()
	tlsConfig.CipherSuites = []uint16{cipher}
	conn, err := tls.Dial("tcp", gwAddr, tlsConfig)

	if expectedError {
		require.Error(t, err, "remote error: tls: handshake failure")
	} else { // not error
		require.NoError(t, err)
		require.NotNil(t, conn)
		require.Equal(t, cipher, conn.ConnectionState().CipherSuite)
		_ = conn.Close()
	}
}

func dialWithCurve(t *testing.T, gwAddr string, baseTLSConfig *tls.Config, curve tls.CurveID, expectedError bool) {
	tlsConfig := baseTLSConfig.Clone()
	tlsConfig.CurvePreferences = []tls.CurveID{curve}
	conn, err := tls.Dial("tcp", gwAddr, tlsConfig)

	if expectedError {
		require.Error(t, err, "remote error: tls: handshake failure")
	} else { // not error
		require.NoError(t, err)
		require.NotNil(t, conn)
		_ = conn.Close()
	}
}

func dialAndCheckSessionTicketAssignment(t *testing.T, gwAddr string, baseTLSConfig *tls.Config, expectedSessionTickets int) {
	tlsConfig := baseTLSConfig.Clone()
	sessionCache := newClientSessionCache(tls.NewLRUClientSessionCache(100))
	tlsConfig.ClientSessionCache = sessionCache
	conn, err := tls.Dial("tcp", gwAddr, tlsConfig)

	require.NoError(t, err)
	require.NotNil(t, conn)
	require.Equal(t, expectedSessionTickets, sessionCache.writes)
	_ = conn.Close()
}

type clientSessionCache struct {
	cache  tls.ClientSessionCache
	writes int
}

func newClientSessionCache(cache tls.ClientSessionCache) *clientSessionCache {
	return &clientSessionCache{
		cache:  cache,
		writes: 0,
	}
}

func (c *clientSessionCache) Get(sessionKey string) (*tls.ClientSessionState, bool) {
	return c.cache.Get(sessionKey)
}

func (c *clientSessionCache) Put(sessionKey string, cs *tls.ClientSessionState) {
	c.cache.Put(sessionKey, cs)
	c.writes++
}

func (c *clientSessionCache) Writes() int {
	return c.writes
}
