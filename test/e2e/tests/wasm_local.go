// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, LocalWasmTest)
}

var LocalWasmTest = suite.ConformanceTest{
	ShortName:   "WasmLocalCodeSource",
	Description: "Test Wasm extension loaded from a local file mounted via a ConfigMap volume",
	Manifests:   []string{"testdata/wasm-local.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("http route with local wasm source", func(t *testing.T) {
			ns := "gateway-conformance-infra"

			// Create a ConfigMap containing the wasm binary so it can be
			// mounted into the Envoy proxy pod as a volume.
			wasmBytes, err := os.ReadFile(filepath.Join("testdata", "wasm", "plugin.wasm"))
			if err != nil {
				t.Fatalf("failed to read wasm binary: %v", err)
			}
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wasm-module-local",
					Namespace: ns,
				},
				BinaryData: map[string][]byte{
					"plugin.wasm": wasmBytes,
				},
			}
			if err := suite.Client.Create(context.Background(), cm); err != nil {
				t.Fatalf("failed to create ConfigMap: %v", err)
			}
			t.Cleanup(func() {
				_ = suite.Client.Delete(context.Background(), cm)
			})

			routeNN := types.NamespacedName{Name: "http-with-local-wasm-source", Namespace: ns}
			gwNN := types.NamespacedName{Name: "local-wasm-gateway", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "local-wasm-source-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/wasm-local",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host:    "",
						Method:  "",
						Path:    "",
						Headers: nil,
					},
				},
				Namespace: "",
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"x-wasm-custom": "FOO",
					},
				},
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
