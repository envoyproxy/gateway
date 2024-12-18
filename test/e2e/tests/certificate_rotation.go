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

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	provider "github.com/envoyproxy/gateway/internal/provider/kubernetes"
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

			// create a gRPC client with envoy's TLS credentials
			tlsConfig, err := tlsClientConfig(crt, key, ca)
			require.NoError(t, err)
			conn, err := grpc.NewClient(envoyGatewayAddr,
				grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
			require.NoError(t, err)

			// Connect to Envoy Gateway's XDS endpoint with Envoy TLS credentials
			streamClient, err := discovery.NewAggregatedDiscoveryServiceClient(conn).
				StreamAggregatedResources(ctx)
			require.NoError(t, err)
			require.NotNil(t, streamClient)
			err = conn.Close()
			require.NoError(t, err)

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

			// Wait for connection with new credentials to succeed
			http.AwaitConvergence(
				t,
				1,
				suite.TimeoutConfig.NamespacesMustBeReady,
				func(_ time.Duration) bool {
					// create gRPC client with envoy's TLS credentials
					tlsConfig, err = tlsClientConfig(certs.EnvoyCertificate, certs.EnvoyPrivateKey,
						certs.CACertificate)
					require.NoError(t, err)
					conn, err = grpc.NewClient(envoyGatewayAddr,
						grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
					require.NoError(t, err)

					// Connect to Envoy Gateway's XDS endpoint with Envoy TLS credentials
					streamClient, err = discovery.NewAggregatedDiscoveryServiceClient(conn).
						StreamAggregatedResources(ctx)
					if err != nil {
						tlog.Logf(t, "failed to connect to Envoy Gateway with new tls credentials: %v", err)
						err = conn.Close()
						require.NoError(t, err)
						time.Sleep(1 * time.Second)
						return false
					}

					tlog.Logf(t, "Connected to Envoy Gateway with new tls credentials")
					err = conn.Close()
					require.NoError(t, err)
					return true
				})
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
		ServerName:   "envoy-gateway.envoy-gateway-system.svc.cluster.local",
		MinVersion:   tls.VersionTLS13,
	}, nil
}
