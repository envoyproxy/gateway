// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, EndpointFastPathTest)
}

var EndpointFastPathTest = suite.ConformanceTest{
	ShortName:   "EndpointFastPath",
	Description: "Endpoint updates and empty-backend fallback converge with fast publication enabled",
	Manifests:   []string{"testdata/endpoint-fast-path.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		if DeployProfile != "endpoint-fast-path" {
			t.Skip("requires the endpoint-fast-path deployment profile")
		}
		kubernetes.NamespacesMustBeReady(t, s.Client, s.TimeoutConfig, []string{ConformanceInfraNamespace})
		// Use a selectorless Service so only this test controls its endpoint inputs.
		endpoints := make([]discoveryv1.Endpoint, 0, 2)
		for _, app := range []string{"infra-backend-v1", "infra-backend-v2"} {
			var pods corev1.PodList
			require.NoError(t, s.Client.List(t.Context(), &pods, client.InNamespace(ConformanceInfraNamespace), client.MatchingLabels{"app": app}))
			require.NotEmpty(t, pods.Items)
			require.NotEmpty(t, pods.Items[0].Status.PodIP)
			endpoints = append(endpoints, discoveryv1.Endpoint{
				Addresses:  []string{pods.Items[0].Status.PodIP},
				Conditions: discoveryv1.EndpointConditions{Ready: new(true)},
			})
		}
		slice := &discoveryv1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{
				Name: "endpoint-fast-path", Namespace: ConformanceInfraNamespace,
				Labels: map[string]string{discoveryv1.LabelServiceName: "endpoint-fast-path", discoveryv1.LabelManagedBy: "endpoint-fast-path-test"},
			},
			AddressType: discoveryv1.AddressTypeIPv4,
			Ports:       []discoveryv1.EndpointPort{{Name: new("http"), Protocol: new(corev1.ProtocolTCP), Port: new(int32(3000))}},
			Endpoints:   []discoveryv1.Endpoint{endpoints[0]},
		}
		require.NoError(t, s.Client.Create(t.Context(), slice))
		t.Cleanup(func() { require.NoError(t, client.IgnoreNotFound(s.Client.Delete(context.Background(), slice))) })
		route := types.NamespacedName{Namespace: ConformanceInfraNamespace, Name: "endpoint-fast-path"}
		address := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, s.Client, s.TimeoutConfig, s.ControllerName,
			kubernetes.NewGatewayRef(SameNamespaceGateway), route)
		expect := func(backend string, code int) {
			t.Helper()
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, address, http.ExpectedResponse{
				Request: http.Request{Path: "/endpoint-fast-path"}, Response: http.Response{StatusCode: code},
				Backend: backend, Namespace: ConformanceInfraNamespace,
			})
		}
		expect("infra-backend-v1", 200)
		for _, step := range []struct {
			index, code int
			backend     string
		}{
			{1, 200, "infra-backend-v2"},
			{-1, 503, ""},
			{0, 200, "infra-backend-v1"},
		} {
			require.NoError(t, s.Client.Get(t.Context(), client.ObjectKeyFromObject(slice), slice))
			slice.Endpoints = nil
			if step.index >= 0 {
				slice.Endpoints = []discoveryv1.Endpoint{endpoints[step.index]}
			}
			require.NoError(t, s.Client.Update(t.Context(), slice))
			expect(step.backend, step.code)
		}
	},
}
