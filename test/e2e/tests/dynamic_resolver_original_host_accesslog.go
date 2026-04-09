// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, DynamicResolverOriginalHostAccessLogTest)
}

var DynamicResolverOriginalHostAccessLogTest = suite.ConformanceTest{
	ShortName:   "DynamicResolverOriginalHostAccessLog",
	Description: "Verify access logs preserve X-ENVOY-ORIGINAL-HOST when dynamic resolver hostname rewrite is configured and Envoy headers are enabled",
	Manifests:   []string{"testdata/dynamic-resolver-original-host-accesslog.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := ConformanceInfraNamespace
		gwNN := types.NamespacedName{Name: "dynamic-resolver-accesslog-gtw", Namespace: ns}
		routeNN := types.NamespacedName{Name: "dynamic-resolver-original-host-accesslog", Namespace: ns}
		backendNN := types.NamespacedName{Name: "backend-dynamic-resolver-accesslog", Namespace: ns}

		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
		BackendMustBeAccepted(t, suite.Client, backendNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		ClientTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "dynamic-resolver-accesslog-ctp", Namespace: ns}, suite.ControllerName, ancestorRef)

		requestID := "dynamic-resolver-original-host"
		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Host: "www.example.com",
				Path: "/original-host",
				Headers: map[string]string{
					"x-envoy-logged": "1",
					"x-request-id":   requestID,
				},
			},
			ExpectedRequest: &http.ExpectedRequest{
				Request: http.Request{
					Host: "test-service-foo.gateway-conformance-infra.svc.cluster.local",
					Path: "/original-host",
				},
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

		labels := getOTELLabels(ns)
		match := `\"x-request-id\":\"` + requestID + `\"`

		var matchedLog string
		err := wait.PollUntilContextTimeout(context.Background(), time.Second, time.Minute, true, func(_ context.Context) (bool, error) {
			lines, err := QueryLogLinesFromLoki(t, suite.Client, labels, match)
			if err != nil {
				tlog.Logf(t, "failed to query log lines from loki: %v", err)
				return false, nil
			}

			for _, line := range lines {
				tlog.Logf(t, "candidate access log line: %s", line)
				entry := map[string]any{}
				if err := json.Unmarshal([]byte(line), &entry); err != nil {
					tlog.Logf(t, "failed to unmarshal access log line: %v", err)
					continue
				}

				if gotRequestID, ok := entry["x-request-id"].(string); ok && gotRequestID == requestID {
					matchedLog = line
					return true, nil
				}
			}

			return false, nil
		})
		require.NoError(t, err, "timed out waiting for access log line for request id %s", requestID)

		entry := map[string]string{}
		require.NoError(t, json.Unmarshal([]byte(matchedLog), &entry))
		require.Equal(t, "test-service-foo.gateway-conformance-infra.svc.cluster.local", entry["http.host"])
		require.Equal(t, "www.example.com", entry["http.original_host"])
	},
}
