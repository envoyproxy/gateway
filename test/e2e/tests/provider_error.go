// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

func init() {
	ConformanceTests = append(ConformanceTests, ProviderErrorTest)
}

var ProviderErrorTest = suite.ConformanceTest{
	ShortName:   "ProviderError",
	Description: "Test error handling in the provider",
	Manifests:   []string{"testdata/provider-error.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Continue serving requests with last known good configuration when errors are present in the provider", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "simple-http-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			// send a request to the gateway
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// Update the Envoy Proxy client TLS settings to an invalid value
			// This causes the provider to return an error when processing the resources
			proxyNN := types.NamespacedName{Name: "proxy-config", Namespace: "envoy-gateway-system"}
			config := &egv1a1.BackendTLSConfig{
				ClientCertificateRef: &gwapiv1.SecretObjectReference{
					Kind:      gatewayapi.KindPtr("Secret"),
					Name:      "client-tls-certificate",
					Namespace: gatewayapi.NamespacePtr("envoy-gateway-system"),
				},
			}
			// We set replicas to 0, if EG doesn't skip the IRs with errors, the envoy proxy pods will be scaled down to 0
			// and the gateway will stop serving traffic.
			err := UpdateEnvoyProxy(suite.Client, proxyNN, config, 0)
			if err != nil {
				t.Error(err)
			}

			// Wait for a short period to allow the provider to process the resources and encounter the error
			time.Sleep(5 * time.Second)
			// We expect the gateway to continue serving requests with the last known good configuration
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// Set the Envoy Proxy client TLS settings back to the valid value
			err = UpdateEnvoyProxy(suite.Client, proxyNN, &egv1a1.BackendTLSConfig{
				ClientCertificateRef: nil,
				TLSSettings:          egv1a1.TLSSettings{},
			}, 1)
			if err != nil {
				t.Error(err)
			}
		})
	},
}

func UpdateEnvoyProxy(client client.Client, proxyNN types.NamespacedName, config *egv1a1.BackendTLSConfig, replica int32) error {
	proxyConfig := &egv1a1.EnvoyProxy{}
	err := client.Get(context.Background(), proxyNN, proxyConfig)
	if err != nil {
		return err
	}

	proxyConfig.Spec.BackendTLS = config
	proxyConfig.Spec.Provider.Kubernetes.EnvoyDeployment.Replicas = ptr.To(replica)
	err = client.Update(context.Background(), proxyConfig)
	if err != nil {
		return err
	}
	return nil
}
