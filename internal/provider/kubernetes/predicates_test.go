// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/log"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes/test"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
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
			name:   "matching controller name",
			obj:    test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
			expect: true,
		},
		{
			name:   "non-matching controller name",
			obj:    test.GetGatewayClass("test-gc", "not.configured/controller"),
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
	testCases := []struct {
		name    string
		configs []client.Object
		gateway client.Object
		expect  bool
	}{
		{
			name:    "references class with matching controller name",
			configs: []client.Object{test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName)},
			gateway: test.GetGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc"),
			expect:  true,
		},
		{
			name:    "references class with non-matching controller name",
			configs: []client.Object{test.GetGatewayClass("test-gc", "not.configured/controller")},
			gateway: test.GetGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc"),
			expect:  false,
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
		r.client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.configs...).Build()
		t.Run(tc.name, func(t *testing.T) {
			res := r.validateGatewayForReconcile(tc.gateway)
			require.Equal(t, tc.expect, res)
		})
	}
}

// TestValidateSecretForReconcile tests the validateSecretForReconcile
// predicate function.
func TestValidateSecretForReconcile(t *testing.T) {
	testCases := []struct {
		name    string
		configs []client.Object
		secret  client.Object
		expect  bool
	}{
		{
			name: "references valid gateway",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
				test.GetSecureGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc", test.ObjectKindNamespacedName{
					Kind: gatewayapi.KindSecret,
					Name: "secret",
				}),
			},
			secret: test.GetSecret(types.NamespacedName{Name: "secret"}),
			expect: true,
		},
		{
			name: "references invalid gateway",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", "not.configured/controller"),
				test.GetSecureGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc", test.ObjectKindNamespacedName{
					Kind: gatewayapi.KindSecret,
					Name: "secret",
				}),
			},
			secret: test.GetSecret(types.NamespacedName{Name: "secret"}),
			expect: false,
		},
		{
			name: "gateway does not exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
			},
			secret: test.GetSecret(types.NamespacedName{Name: "secret"}),
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
		r.client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.configs...).Build()
		t.Run(tc.name, func(t *testing.T) {
			res := r.validateSecretForReconcile(tc.secret)
			require.Equal(t, tc.expect, res)
		})
	}
}

// TestValidateServiceForReconcile tests the validateServiceForReconcile
// predicate function.
func TestValidateServiceForReconcile(t *testing.T) {
	sampleGateway := test.GetGateway(types.NamespacedName{Namespace: "default", Name: "scheduled-status-test"}, "test-gc")

	testCases := []struct {
		name    string
		configs []client.Object
		service client.Object
		expect  bool
	}{
		{
			name: "gateway service but deployment does not exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
				sampleGateway,
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, map[string]string{
				gatewayapi.OwningGatewayNameLabel:      "scheduled-status-test",
				gatewayapi.OwningGatewayNamespaceLabel: "default",
			}, nil),
			expect: false,
		},
		{
			name: "gateway service deployment also exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
				sampleGateway,
				test.GetGatewayDeployment(types.NamespacedName{Name: infraDeploymentName(sampleGateway)}, nil),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, map[string]string{
				gatewayapi.OwningGatewayNameLabel:      "scheduled-status-test",
				gatewayapi.OwningGatewayNamespaceLabel: "default",
			}, nil),
			// Note that in case when a deployment exists, the Service is just processed for Gateway status
			// updates and not reconciled further.
			expect: false,
		},
		{
			name: "route service but no routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
				sampleGateway,
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  false,
		},
		{
			name: "route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
				sampleGateway,
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test"),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  true,
		},
		{
			// The service should still be reconciled if the Route object references an invalid parent.
			// This takes care of a case where the Route objects' parent reference is updated  - from valid to invalid.
			// in which case we'll have to reconcile the bad config, and remove listeners accordingly.
			name: "route service routes exist but with non-existing gateway reference",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test"),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  true,
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
		r.client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.configs...).Build()
		t.Run(tc.name, func(t *testing.T) {
			res := r.validateServiceForReconcile(tc.service)
			require.Equal(t, tc.expect, res)
		})
	}
}

// TestValidateDeploymentForReconcile tests the validateDeploymentForReconcile
// predicate function.
func TestValidateDeploymentForReconcile(t *testing.T) {
	sampleGateway := test.GetGateway(types.NamespacedName{Namespace: "default", Name: "scheduled-status-test"}, "test-gc")

	testCases := []struct {
		name       string
		configs    []client.Object
		deployment client.Object
		expect     bool
	}{
		{
			// No config should lead to a reconciliation of a Deployment object. The main
			// purpose of the Deployment watcher is just for update Gateway object statuses.
			name: "gateway deployment deployment also exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
				sampleGateway,
				test.GetService(types.NamespacedName{Name: "deployment"}, map[string]string{
					gatewayapi.OwningGatewayNameLabel:      "scheduled-status-test",
					gatewayapi.OwningGatewayNamespaceLabel: "default",
				}, nil),
			},
			deployment: test.GetGatewayDeployment(types.NamespacedName{Name: "deployment"}, map[string]string{
				gatewayapi.OwningGatewayNameLabel:      "scheduled-status-test",
				gatewayapi.OwningGatewayNamespaceLabel: "default",
			}),
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
		r.client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.configs...).Build()
		t.Run(tc.name, func(t *testing.T) {
			res := r.validateDeploymentForReconcile(tc.deployment)
			require.Equal(t, tc.expect, res)
		})
	}
}
