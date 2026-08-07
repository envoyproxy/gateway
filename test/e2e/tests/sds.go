// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, SDS)
}

var SDS = suite.ConformanceTest{
	ShortName:   "SDS",
	Description: "Test SDS server providing secrets to Envoy",
	Manifests:   []string{"testdata/sds.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		acceptedCond := metav1.Condition{
			Type:   string(gwapiv1.PolicyConditionAccepted),
			Status: metav1.ConditionTrue,
			Reason: string(gwapiv1.PolicyReasonAccepted),
		}
		resolvedRefsCond := metav1.Condition{
			Type:   string(gwapiv1.BackendTLSPolicyConditionResolvedRefs),
			Status: metav1.ConditionTrue,
			Reason: string(gwapiv1.BackendTLSPolicyReasonResolvedRefs),
		}

		routeNN := types.NamespacedName{Name: "sds-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: "sds-gateway", Namespace: ns}
		validPolicyNN := types.NamespacedName{Name: "tls-backend-policy", Namespace: ns}
		kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, validPolicyNN, gwNN, acceptedCond)
		kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, validPolicyNN, gwNN, resolvedRefsCond)
		gwAddr := GatewayAndTCPRoutesMustBeAccepted(t, suite.Client, &suite.TimeoutConfig, suite.ControllerName, NewGatewayRef(gwNN), routeNN)

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
			Request: http.Request{
				Path: "/",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		})
	},
}
