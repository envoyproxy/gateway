// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, EnvoyProxyStatus)
}

var EnvoyProxyStatus = suite.ConformanceTest{
	ShortName:   "EnvoyProxyStatus",
	Description: "Make sure that the status of EnvoyProxy works as expected.",
	Manifests:   []string{"testdata/envoyproxy-status.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		epName := "test-envoyproxy"
		EnvoyProxyMustBeAccepted(t, suite.Client, types.NamespacedName{
			Name:      epName,
			Namespace: ns,
		}, gwapiv1.ParentReference{
			Group: gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:  gatewayapi.KindPtr(resource.KindGatewayClass),
			Name:  gwapiv1.ObjectName("test-gc1"),
		})
		EnvoyProxyMustBeAccepted(t, suite.Client, types.NamespacedName{
			Name:      epName,
			Namespace: ns,
		}, gwapiv1.ParentReference{
			Group: gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:  gatewayapi.KindPtr(resource.KindGatewayClass),
			Name:  gwapiv1.ObjectName("test-gc2"),
		})

		// Remove parametersRef from gc1, and make sure the status is removed from EnvoyProxy.
		http.AwaitConvergence(t, 1, suite.TimeoutConfig.MaxTimeToConsistency, func(_ time.Duration) bool {
			// Get gc1 and remove parametersRef.
			gc1 := &gwapiv1.GatewayClass{}
			if err := suite.Client.Get(t.Context(), types.NamespacedName{
				Name: "test-gc1",
			}, gc1); err != nil {
				tlog.Logf(t, "failed to get gc1: %v", err)
				return false
			}

			gc1.Spec.ParametersRef = nil
			if err := suite.Client.Update(t.Context(), gc1); err != nil {
				tlog.Logf(t, "failed to update gc1: %v", err)
				return false
			}

			return true
		})

		// ParametersRef is removed from gc1, so the status of EnvoyProxy should be removed.
		EnvoyProxyMustNotAccepted(t, suite.Client, types.NamespacedName{
			Name:      epName,
			Namespace: ns,
		}, gwapiv1.ParentReference{
			Group: gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:  gatewayapi.KindPtr(resource.KindGatewayClass),
			Name:  gwapiv1.ObjectName("test-gc1"),
		})
		// The status of EnvoyProxy should still be accepted by gc2.
		EnvoyProxyMustBeAccepted(t, suite.Client, types.NamespacedName{
			Name:      epName,
			Namespace: ns,
		}, gwapiv1.ParentReference{
			Group: gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:  gatewayapi.KindPtr(resource.KindGatewayClass),
			Name:  gwapiv1.ObjectName("test-gc2"),
		})
	},
}
