// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/test/e2e/utils"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTP3Test)
}

var HTTP3Test = suite.ConformanceTest{
	ShortName:   "HTTP3",
	Description: "HTTP3 tests ensure that Envoy Gateway supports HTTP/3 features.",
	Manifests:   []string{"testdata/http3.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		routeNN := types.NamespacedName{Name: "http3-route-foo", Namespace: ConformanceInfraNamespace}
		gwNN := types.NamespacedName{Name: "http3-gtw", Namespace: ConformanceInfraNamespace}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), routeNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		ClientTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "http3-ctp", Namespace: ConformanceInfraNamespace},
			suite.ControllerName, ancestorRef)

		testHTTP3 := func(host, secretName string) {
			quicRoundTripper := &utils.QuicRoundTripper{
				Debug:         suite.Debug,
				TimeoutConfig: suite.TimeoutConfig,
			}

			expected := http.ExpectedResponse{
				Request: http.Request{
					Host: host,
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ConformanceInfraNamespace,
			}

			cPem, keyPem, _, err := GetTLSSecret(suite.Client, types.NamespacedName{Name: secretName, Namespace: ConformanceInfraNamespace})
			if err != nil {
				t.Fatalf("unexpected error finding TLS secret: %v", err)
			}

			req := http.MakeRequest(t, &expected, gwAddr, "HTTPS", "https")
			WaitForConsistentMTLSResponse(t, quicRoundTripper, &req, &expected, suite.TimeoutConfig.RequiredConsecutiveSuccesses, suite.TimeoutConfig.MaxTimeToConsistency,
				cPem, keyPem, host)
		}

		testHTTP3("foo.example.com", "foo-com-tls")
		testHTTP3("bar.example.com", "bar-com-tls")
		testHTTP3("www.awesome.org", "www-awesome-org-tls") // test with a domain overlapped with a wildcard domain
		testHTTP3("api.awesome.org", "awesome-org-tls")     // test with a wildcard domain
	},
}
