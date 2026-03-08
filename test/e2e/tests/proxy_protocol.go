// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func init() {
	ConformanceTests = append(ConformanceTests, ProxyProtocolTest)
}

var ProxyProtocolTest = suite.ConformanceTest{
	ShortName:   "ProxyProtocol",
	Description: "Make sure ProxyProtocol is working",
	Manifests:   []string{"testdata/proxy-protocol-with-tls.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "http", Namespace: ns}
		gwNN := types.NamespacedName{Name: "proxy-protocol-gtw", Namespace: ns}

		// Update the backend FQDN to point to the service in the same namespace when using gateway namespace mode.
		if IsGatewayNamespaceMode() {
			require.NoError(t, wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true, func(_ context.Context) (bool, error) {
				backend := &egv1a1.Backend{}
				err := suite.Client.Get(t.Context(), types.NamespacedName{
					Name:      "proxy-protocol-backend",
					Namespace: ns,
				}, backend)
				if err != nil {
					return false, nil
				}
				backend.Spec.Endpoints = []egv1a1.BackendEndpoint{
					{
						FQDN: &egv1a1.FQDNEndpoint{
							Hostname: fmt.Sprintf("%s.%s.svc", gwNN.Name, gwNN.Namespace),
							Port:     443,
						},
					},
				}

				if err := suite.Client.Update(t.Context(), backend); err != nil {
					return false, nil
				}

				return true, nil
			}))
		}

		BackendMustBeAccepted(t, suite.Client, types.NamespacedName{
			Name:      "proxy-protocol-backend",
			Namespace: ns,
		})

		_ = kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client,
			suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, types.NamespacedName{
				Name: "proxy-protocol", Namespace: ns,
			})

		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client,
			suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		expectedResponse := httputils.ExpectedResponse{
			Request: httputils.Request{
				Path: "/",
			},
			Response: httputils.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}
		httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
	},
}
