// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/log"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// TestGatewayClassHasMatchingController tests the hasMatchingController
// predicate function.
func TestGatewayClassHasMatchingController(t *testing.T) {
	testCases := []struct {
		name   string
		obj    client.Object
		expect bool
	}{
		{
			name: "matching controller name",
			obj: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: v1alpha1.GatewayControllerName,
				},
			},
			expect: true,
		},
		{
			name: "non-matching controller name",
			obj: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: "not.configured/controller",
				},
			},
			expect: false,
		},
	}

	// Create the reconciler.
	logger, err := log.NewLogger()
	require.NoError(t, err)
	r := gatewayAPIReconciler{
		classController: v1alpha1.GatewayControllerName,
		log:             logger,
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res := r.hasMatchingController(tc.obj)
			require.Equal(t, tc.expect, res)
		})
	}
}

// TestValidateGatewayForReconcile tests the validateGatewayForReconcile
// predicate function.
func TestValidateGatewayForReconcile(t *testing.T) {

}

// TestValidateSecretForReconcile tests the validateSecretForReconcile
// predicate function.
func TestValidateSecretForReconcile(t *testing.T) {

}

// TestValidateServiceForReconcile tests the validateServiceForReconcile
// predicate function.
func TestValidateServiceForReconcile(t *testing.T) {

}

// TestValidateDeploymentForReconcile tests the validateDeploymentForReconcile
// predicate function.
func TestValidateDeploymentForReconcile(t *testing.T) {

}
