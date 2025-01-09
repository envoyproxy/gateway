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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	provider "github.com/envoyproxy/gateway/internal/provider/kubernetes"
	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests, CertificateRotationTest)
}

var CertificateRotationTest = suite.ConformanceTest{
	ShortName:   "CertificateRotation",
	Description: "Rotate Control Plane Certificates",
	Manifests:   []string{},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Envoy Gateway uses new TLS credentials after rotation", func(t *testing.T) {
			envoyGatewayNS := "envoy-gateway-system"
			EnvoyGatewayLBSVC := "envoy-gateway-ext-lb"
			EnvoyCertificateSecret := "envoy"
			EnvoyGatewayXDSPort := 18000
			var envoyGatewayAddr string

			ctx := context.Background()
			envoyGatewaySvc := &corev1.Service{}
			err := suite.Client.Get(ctx, types.NamespacedName{Namespace: envoyGatewayNS, Name: EnvoyGatewayLBSVC}, envoyGatewaySvc)
			require.NoError(t, err)
			require.Len(t, envoyGatewaySvc.Status.LoadBalancer.Ingress, 1)
			require.NotEmpty(t, envoyGatewaySvc.Status.LoadBalancer.Ingress[0].IP)

			if IPFamily == "ipv6" {
				envoyGatewayAddr = fmt.Sprintf("[%s]:%d", envoyGatewaySvc.Status.LoadBalancer.Ingress[0].IP, EnvoyGatewayXDSPort)
			} else {
				envoyGatewayAddr = fmt.Sprintf("%s:%d", envoyGatewaySvc.Status.LoadBalancer.Ingress[0].IP, EnvoyGatewayXDSPort)
			}

			// get the current envoy TLS credentials
			certNN := types.NamespacedName{Namespace: envoyGatewayNS, Name: EnvoyCertificateSecret}
			crt, key, ca, err := GetTLSSecret(suite.Client, certNN)
			require.NoError(t, err)

			// create a TLS connection with envoy's TLS credentials towards EG
			tlsConfig, err := tlsClientConfig(crt, key, ca)
			require.NoError(t, err)
			conn, err := tls.Dial("tcp", envoyGatewayAddr, tlsConfig)
			require.NoError(t, err)
			err = conn.Close()
			require.NoError(t, err)

			// fetch the current value of envoy's ssl context updates by SDS
			promClient, err := prometheus.NewClient(suite.Client, types.NamespacedName{Name: "prometheus", Namespace: "monitoring"})
			require.NoError(t, err)
			clusterName := "xds_cluster"
			gtwName := "same-namespace"
			promQL := fmt.Sprintf(`envoy_cluster_client_ssl_socket_factory_ssl_context_update_by_sds{envoy_cluster_name="%s",gateway_envoyproxy_io_owning_gateway_name="%s"}`, clusterName, gtwName)
			sslContextUpdateCounter, err := promClient.QuerySum(ctx, promQL)
			require.NoError(t, err)
			tlog.Logf(t, "Envoy SSL context updates before rotation: %f", sslContextUpdateCounter)

			// rotate certs and apply them similar to how EG certgen works
			certs, err := crypto.GenerateCerts(&config.Server{
				EnvoyGateway: &egv1a1.EnvoyGateway{
					EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
						Provider: &egv1a1.EnvoyGatewayProvider{
							Type: egv1a1.ProviderTypeKubernetes,
						},
						Gateway: egv1a1.DefaultGateway(),
					},
				},
				Namespace: envoyGatewayNS,
				DNSDomain: "cluster.local",
			})
			require.NoError(t, err)
			secrets := provider.CertsToSecret(envoyGatewayNS, certs)
			_, err = provider.CreateOrUpdateSecrets(ctx, suite.Client, secrets, true)
			require.NoError(t, err)

			// Wait for connection with new credentials to Envoy Gateway to succeed
			http.AwaitConvergence(
				t,
				1,
				suite.TimeoutConfig.NamespacesMustBeReady,
				func(_ time.Duration) bool {
					// create gRPC client with envoy's TLS credentials
					tlsConfig, err = tlsClientConfig(certs.EnvoyCertificate, certs.EnvoyPrivateKey,
						certs.CACertificate)
					require.NoError(t, err)
					conn, err = tls.Dial("tcp", envoyGatewayAddr, tlsConfig)
					if err != nil {
						tlog.Logf(t, "failed to connect to Envoy Gateway with new tls credentials: %v", err)
						if conn != nil {
							err = conn.Close()
							require.NoError(t, err)
						}
						time.Sleep(1 * time.Second)
						return false
					}

					tlog.Logf(t, "Connected to Envoy Gateway with new tls credentials")
					err = conn.Close()
					require.NoError(t, err)
					return true
				})

			// Wait for Envoy's ssl context to be updated by SDS
			http.AwaitConvergence(
				t,
				1,
				suite.TimeoutConfig.NamespacesMustBeReady,
				func(_ time.Duration) bool {
					// check ssl context was updated in envoy stats from Prometheus
					v, err := promClient.QuerySum(ctx, promQL)
					if err != nil {
						// wait until Prometheus sync stats
						return false
					}

					tlog.Logf(t, "Envoy SSL context updates after rotation: %f", v)
					return v > sslContextUpdateCounter
				},
			)

			// Apply a new config and confirm that it's programmed successfully on proxies
			suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/certificate-rotation.yaml", false)
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-for-cert-rotation", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

			expected := http.ExpectedResponse{
				Request: http.Request{
					Path: "/cert-rotation",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)
		})
	},
}

func tlsClientConfig(certPem []byte, keyPem []byte, caPem []byte) (*tls.Config, error) {
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return nil, fmt.Errorf("unexpected error creating cert: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPem) {
		return nil, fmt.Errorf("unexpected error adding trusted CA: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
		ServerName:   "envoy-gateway",
		MinVersion:   tls.VersionTLS13,
		NextProtos:   []string{"h2", "http/1.1"},
	}, nil
}
