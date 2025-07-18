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
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, EnvoyPatchPolicyTest)
}

var EnvoyPatchPolicyTest = suite.ConformanceTest{
	ShortName:   "EnvoyPatchPolicy",
	Description: "update xds using EnvoyPatchPolicy",
	Manifests:   []string{"testdata/envoy-patch-policy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("envoy patch policy", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-envoy-patch-policy", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			envoyPatchPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "custom-response-patch-policy", Namespace: ns}, suite.ControllerName, ancestorRef)

			OkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/foo",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			// Send a request to an valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)

			customResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/bar",
				},
				Response: http.Response{
					StatusCode: 406,
				},
				Namespace: ns,
			}

			// Send a request to an invalid path and expect a custom response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, customResp)
		})
	},
}

// envoyPatchPolicyMustBeAccepted waits for the specified EnvoyPatchPolicy to be accepted.
func envoyPatchPolicyMustBeAccepted(t *testing.T, client client.Client, policyName types.NamespacedName, controllerName string, ancestorRef gwapiv1a2.ParentReference) {
	t.Helper()

	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		policy := &egv1a1.EnvoyPatchPolicy{}
		err := client.Get(ctx, policyName, policy)
		if err != nil {
			return false, fmt.Errorf("error fetching EnvoyPatchPolicy: %w", err)
		}

		if policyAcceptedByAncestor(policy.Status.Ancestors, controllerName, ancestorRef) {
			tlog.Logf(t, "EnvoyPatchPolicy has been accepted: %v", policy)
			return true, nil
		}

		tlog.Logf(t, "EnvoyPatchPolicy not yet accepted: %v", policy)
		return false, nil
	})

	require.NoErrorf(t, waitErr, "error waiting for EnvoyPatchPolicy to be accepted")
}
