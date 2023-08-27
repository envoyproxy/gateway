// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes/test"
)

// TestGatewayClassHasMatchingController tests the hasMatchingController
// predicate function.
func TestGatewayClassHasMatchingController(t *testing.T) {
	testCases := []struct {
		name   string
		obj    client.Object
		client client.Client
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
	logger := logging.DefaultLogger(v1alpha1.LogLevelInfo)

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

// TestGatewayClassHasMatchingNamespaceLabels tests the hasMatchingNamespaceLabels
// predicate function.
func TestGatewayClassHasMatchingNamespaceLabels(t *testing.T) {
	ns := "namespace-1"
	testCases := []struct {
		name            string
		labels          []string
		namespaceLabels []string
		expect          bool
	}{
		{
			name:            "matching one label when namespace has one label",
			labels:          []string{"label-1"},
			namespaceLabels: []string{"label-1"},
			expect:          true,
		},
		{
			name:            "matching one label when namespace has two labels",
			labels:          []string{"label-1"},
			namespaceLabels: []string{"label-1", "label-2"},
			expect:          true,
		},
		{
			name:            "namespace has less labels than the specified labels",
			labels:          []string{"label-1", "label-2"},
			namespaceLabels: []string{"label-1"},
			expect:          false,
		},
	}

	logger := logging.DefaultLogger(v1alpha1.LogLevelInfo)

	for _, tc := range testCases {
		tc := tc

		namespaceLabelsToMap := make(map[string]string)
		for _, l := range tc.namespaceLabels {
			namespaceLabelsToMap[l] = ""
		}

		r := gatewayAPIReconciler{
			classController: v1alpha1.GatewayControllerName,
            namespaceLabels: tc.labels,
			log:             logger,
			client: fakeclient.NewClientBuilder().
				WithScheme(envoygateway.GetScheme()).
				WithObjects(&corev1.Namespace{
					TypeMeta: v1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: v1.ObjectMeta{Name: ns, Labels: namespaceLabelsToMap},
				}).
				Build(),
		}
		t.Run(tc.name, func(t *testing.T) {
			res := r.hasMatchingNamespaceLabels(
				test.GetHTTPRoute(
					types.NamespacedName{
						Namespace: ns,
						Name:      "httproute-test",
					},
					"scheduled-status-test",
					types.NamespacedName{Name: "service"},
				))
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
	logger := logging.DefaultLogger(v1alpha1.LogLevelInfo)

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
	logger := logging.DefaultLogger(v1alpha1.LogLevelInfo)

	r := gatewayAPIReconciler{
		classController: v1alpha1.GatewayControllerName,
		log:             logger,
	}

	for _, tc := range testCases {
		tc := tc
		r.client = fakeclient.NewClientBuilder().
			WithScheme(envoygateway.GetScheme()).
			WithObjects(tc.configs...).
			WithIndex(&gwapiv1b1.Gateway{}, secretGatewayIndex, secretGatewayIndexFunc).
			Build()
		t.Run(tc.name, func(t *testing.T) {
			res := r.validateSecretForReconcile(tc.secret)
			require.Equal(t, tc.expect, res)
		})
	}
}

// TestValidateEndpointSliceForReconcile tests the validateEndpointSliceForReconcile
// predicate function.
func TestValidateEndpointSliceForReconcile(t *testing.T) {
	sampleGateway := test.GetGateway(types.NamespacedName{Namespace: "default", Name: "scheduled-status-test"}, "test-gc")

	testCases := []struct {
		name          string
		configs       []client.Object
		endpointSlice client.Object
		expect        bool
	}{
		{
			name: "route service but no routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
				sampleGateway,
			},
			endpointSlice: test.GetEndpointSlice(types.NamespacedName{Name: "endpointslice"}, "service"),
			expect:        false,
		},
		{
			name: "http route service routes exist, but endpointslice is associated with another service",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
				sampleGateway,
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test", types.NamespacedName{Name: "service"}),
			},
			endpointSlice: test.GetEndpointSlice(types.NamespacedName{Name: "endpointslice"}, "other-service"),
			expect:        false,
		},
		{
			name: "http route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
				sampleGateway,
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test", types.NamespacedName{Name: "service"}),
			},
			endpointSlice: test.GetEndpointSlice(types.NamespacedName{Name: "endpointslice"}, "service"),
			expect:        true,
		},
	}

	// Create the reconciler.
	logger := logging.DefaultLogger(v1alpha1.LogLevelInfo)

	r := gatewayAPIReconciler{
		classController: v1alpha1.GatewayControllerName,
		log:             logger,
	}

	for _, tc := range testCases {
		tc := tc
		r.client = fakeclient.NewClientBuilder().
			WithScheme(envoygateway.GetScheme()).
			WithObjects(tc.configs...).
			WithIndex(&gwapiv1b1.HTTPRoute{}, serviceHTTPRouteIndex, serviceHTTPRouteIndexFunc).
			WithIndex(&gwapiv1a2.GRPCRoute{}, serviceGRPCRouteIndex, serviceGRPCRouteIndexFunc).
			WithIndex(&gwapiv1a2.TLSRoute{}, serviceTLSRouteIndex, serviceTLSRouteIndexFunc).
			WithIndex(&gwapiv1a2.TCPRoute{}, serviceTCPRouteIndex, serviceTCPRouteIndexFunc).
			WithIndex(&gwapiv1a2.UDPRoute{}, serviceUDPRouteIndex, serviceUDPRouteIndexFunc).
			Build()
		t.Run(tc.name, func(t *testing.T) {
			res := r.validateEndpointSliceForReconcile(tc.endpointSlice)
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
			name: "http route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
				sampleGateway,
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test", types.NamespacedName{Name: "service"}),
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
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test", types.NamespacedName{Name: "service"}),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  true,
		},
		{
			name: "grpc route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
				sampleGateway,
				test.GetGRPCRoute(types.NamespacedName{Name: "grpcroute-test"}, "scheduled-status-test", types.NamespacedName{Name: "service"}),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  true,
		},
		{
			name: "tls route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
				sampleGateway,
				test.GetTLSRoute(types.NamespacedName{Name: "tlsroute-test"}, "scheduled-status-test",
					types.NamespacedName{Name: "service"}),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  true,
		},
		{
			name: "udp route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
				sampleGateway,
				test.GetUDPRoute(types.NamespacedName{Name: "udproute-test"}, "scheduled-status-test",
					types.NamespacedName{Name: "service"}),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  true,
		},
		{
			name: "tcp route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName),
				sampleGateway,
				test.GetTCPRoute(types.NamespacedName{Name: "tcproute-test"}, "scheduled-status-test",
					types.NamespacedName{Name: "service"}),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  true,
		},
	}

	// Create the reconciler.
	logger := logging.DefaultLogger(v1alpha1.LogLevelInfo)

	r := gatewayAPIReconciler{
		classController: v1alpha1.GatewayControllerName,
		log:             logger,
	}

	for _, tc := range testCases {
		tc := tc
		r.client = fakeclient.NewClientBuilder().
			WithScheme(envoygateway.GetScheme()).
			WithObjects(tc.configs...).
			WithIndex(&gwapiv1b1.HTTPRoute{}, serviceHTTPRouteIndex, serviceHTTPRouteIndexFunc).
			WithIndex(&gwapiv1a2.GRPCRoute{}, serviceGRPCRouteIndex, serviceGRPCRouteIndexFunc).
			WithIndex(&gwapiv1a2.TLSRoute{}, serviceTLSRouteIndex, serviceTLSRouteIndexFunc).
			WithIndex(&gwapiv1a2.TCPRoute{}, serviceTCPRouteIndex, serviceTCPRouteIndexFunc).
			WithIndex(&gwapiv1a2.UDPRoute{}, serviceUDPRouteIndex, serviceUDPRouteIndexFunc).
			Build()
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
	logger := logging.DefaultLogger(v1alpha1.LogLevelInfo)

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
