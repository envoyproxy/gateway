// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

// GatewaysMustBeAccepted waits for the Gateways to be accepted and returns the address of the Gateways.
// This is used when a HTTPRoute referenced by multiple Gateways.
// Warning: we didn't check the status of HTTPRoute.
func GatewaysMustBeAccepted(t *testing.T, suite *suite.ConformanceTestSuite, gwRefs []kubernetes.GatewayRef) map[types.NamespacedName]string {
	t.Helper()

	gwAddress := make(map[types.NamespacedName]string)
	requiredListenerConditions := []metav1.Condition{
		{
			Type:   string(gwapiv1.ListenerConditionResolvedRefs),
			Status: metav1.ConditionTrue,
			Reason: "", // any reason
		},
		{
			Type:   string(gwapiv1.ListenerConditionAccepted),
			Status: metav1.ConditionTrue,
			Reason: "", // any reason
		},
		{
			Type:   string(gwapiv1.ListenerConditionProgrammed),
			Status: metav1.ConditionTrue,
			Reason: "", // any reason
		},
	}

	for _, gw := range gwRefs {
		gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, gw)
		require.NoErrorf(t, err, "timed out waiting for Gateway %s address to be assigned", gw.NamespacedName)
		gwAddress[gw.NamespacedName] = gwAddr

		kubernetes.GatewayListenersMustHaveConditions(t, suite.Client, suite.TimeoutConfig, gw.NamespacedName, requiredListenerConditions)
	}

	return gwAddress
}
