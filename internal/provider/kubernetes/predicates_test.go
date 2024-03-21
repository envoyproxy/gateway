// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/api/v1alpha1"
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
			obj:    test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
			expect: true,
		},
		{
			name:   "non-matching controller name",
			obj:    test.GetGatewayClass("test-gc", "not.configured/controller", nil),
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
	matchExpressions := func(key string, operator metav1.LabelSelectorOperator, values []string) []metav1.LabelSelectorRequirement {
		return []metav1.LabelSelectorRequirement{{
			Key:      key,
			Operator: operator,
			Values:   values,
		}}
	}
	ns := "namespace-1"
	testCases := []struct {
		name            string
		labels          map[string]string
		namespaceLabels string
		expect          bool
	}{
		{
			name:            "matching one label when namespace has one label",
			labels:          map[string]string{"label-1": ""},
			namespaceLabels: "label-1",
			expect:          true,
		},
		{
			name:            "matching one label when namespace has two labels",
			labels:          map[string]string{"label-1": ""},
			namespaceLabels: "label-2",
			expect:          false,
		},
		{
			name:            "namespace has less labels than the specified labels",
			labels:          map[string]string{"label-1": "", "label-2": ""},
			namespaceLabels: "label-1",
			expect:          true,
		},
	}

	logger := logging.DefaultLogger(v1alpha1.LogLevelInfo)

	for _, tc := range testCases {
		tc := tc

		r := gatewayAPIReconciler{
			classController: v1alpha1.GatewayControllerName,
			namespaceLabel:  &metav1.LabelSelector{MatchExpressions: matchExpressions(tc.namespaceLabels, metav1.LabelSelectorOpExists, []string{})},
			log:             logger,
			client: fakeclient.NewClientBuilder().
				WithScheme(envoygateway.GetScheme()).
				WithObjects(&corev1.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{Name: ns, Labels: tc.labels},
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
					80,
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
			configs: []client.Object{test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil)},
			gateway: test.GetGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc", 8080),
			expect:  true,
		},
		{
			name:    "references class with non-matching controller name",
			configs: []client.Object{test.GetGatewayClass("test-gc", "not.configured/controller", nil)},
			gateway: test.GetGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc", 8080),
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
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
				test.GetSecureGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc", test.GroupKindNamespacedName{
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
				test.GetGatewayClass("test-gc", "not.configured/controller", nil),
				test.GetSecureGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc", test.GroupKindNamespacedName{
					Kind: gatewayapi.KindSecret,
					Name: "secret",
				}),
			},
			secret: test.GetSecret(types.NamespacedName{Name: "secret"}),
			expect: false,
		},
		{
			name: "references SecurityPolicy OIDC",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
				test.GetGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc", 8080),
				&v1alpha1.SecurityPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "oidc",
					},
					Spec: v1alpha1.SecurityPolicySpec{
						TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
							PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
								Kind: "Gateway",
								Name: "scheduled-status-test",
							},
						},
						OIDC: &v1alpha1.OIDC{
							Provider: v1alpha1.OIDCProvider{
								Issuer:                "https://accounts.google.com",
								AuthorizationEndpoint: ptr.To("https://accounts.google.com/o/oauth2/v2/auth"),
								TokenEndpoint:         ptr.To("https://oauth2.googleapis.com/token"),
							},
							ClientID: "client-id",
							ClientSecret: gwapiv1b1.SecretObjectReference{
								Name: "secret",
							},
						},
					},
				},
			},
			secret: test.GetSecret(types.NamespacedName{Name: "secret"}),
			expect: true,
		},
		{
			name: "references SecurityPolicy Basic Auth",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
				test.GetGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc", 8080),
				&v1alpha1.SecurityPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "basic-auth",
					},
					Spec: v1alpha1.SecurityPolicySpec{
						TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
							PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
								Kind: "Gateway",
								Name: "scheduled-status-test",
							},
						},
						BasicAuth: &v1alpha1.BasicAuth{
							Users: gwapiv1b1.SecretObjectReference{
								Name: "secret",
							},
						},
					},
				},
			},
			secret: test.GetSecret(types.NamespacedName{Name: "secret"}),
			expect: true,
		},
		{
			name: "secret is not referenced by any EG CRs",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
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
			WithIndex(&gwapiv1.Gateway{}, secretGatewayIndex, secretGatewayIndexFunc).
			WithIndex(&v1alpha1.SecurityPolicy{}, secretSecurityPolicyIndex, secretSecurityPolicyIndexFunc).
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
	sampleGateway := test.GetGateway(types.NamespacedName{Namespace: "default", Name: "scheduled-status-test"}, "test-gc", 8080)

	testCases := []struct {
		name          string
		configs       []client.Object
		endpointSlice client.Object
		expect        bool
	}{
		{
			name: "route service but no routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
				sampleGateway,
			},
			endpointSlice: test.GetEndpointSlice(types.NamespacedName{Name: "endpointslice"}, "service"),
			expect:        false,
		},
		{
			name: "http route service routes exist, but endpointslice is associated with another service",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
				sampleGateway,
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test", types.NamespacedName{Name: "service"}, 80),
			},
			endpointSlice: test.GetEndpointSlice(types.NamespacedName{Name: "endpointslice"}, "other-service"),
			expect:        false,
		},
		{
			name: "http route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
				sampleGateway,
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test", types.NamespacedName{Name: "service"}, 80),
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
			WithIndex(&gwapiv1.HTTPRoute{}, backendHTTPRouteIndex, backendHTTPRouteIndexFunc).
			WithIndex(&gwapiv1a2.GRPCRoute{}, backendGRPCRouteIndex, backendGRPCRouteIndexFunc).
			WithIndex(&gwapiv1a2.TLSRoute{}, backendTLSRouteIndex, backendTLSRouteIndexFunc).
			WithIndex(&gwapiv1a2.TCPRoute{}, backendTCPRouteIndex, backendTCPRouteIndexFunc).
			WithIndex(&gwapiv1a2.UDPRoute{}, backendUDPRouteIndex, backendUDPRouteIndexFunc).
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
	sampleGateway := test.GetGateway(types.NamespacedName{Namespace: "default", Name: "scheduled-status-test"}, "test-gc", 8080)
	mergeGatewaysConfig := test.GetEnvoyProxy(types.NamespacedName{Namespace: "default", Name: "merge-gateways-config"}, true)

	testCases := []struct {
		name    string
		configs []client.Object
		service client.Object
		expect  bool
	}{
		{
			name: "gateway service but deployment does not exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
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
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
				sampleGateway,
				test.GetGatewayDeployment(types.NamespacedName{Name: infraName(sampleGateway, false)}, nil),
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
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
				sampleGateway,
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  false,
		},
		{
			name: "http route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
				sampleGateway,
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test", types.NamespacedName{Name: "service"}, 80),
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
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test", types.NamespacedName{Name: "service"}, 80),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  true,
		},
		{
			name: "grpc route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
				sampleGateway,
				test.GetGRPCRoute(types.NamespacedName{Name: "grpcroute-test"}, "scheduled-status-test", types.NamespacedName{Name: "service"}, 80),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  true,
		},
		{
			name: "tls route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
				sampleGateway,
				test.GetTLSRoute(types.NamespacedName{Name: "tlsroute-test"}, "scheduled-status-test",
					types.NamespacedName{Name: "service"}, 443),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  true,
		},
		{
			name: "udp route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
				sampleGateway,
				test.GetUDPRoute(types.NamespacedName{Name: "udproute-test"}, "scheduled-status-test",
					types.NamespacedName{Name: "service"}, 80),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  true,
		},
		{
			name: "tcp route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
				sampleGateway,
				test.GetTCPRoute(types.NamespacedName{Name: "tcproute-test"}, "scheduled-status-test",
					types.NamespacedName{Name: "service"}, 80),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  true,
		},
		{
			name: "service referenced by SecurityPolicy ExtAuth HTTP service",
			configs: []client.Object{
				&v1alpha1.SecurityPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ext-auth-http",
					},
					Spec: v1alpha1.SecurityPolicySpec{
						TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
							PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
								Kind: "Gateway",
								Name: "scheduled-status-test",
							},
						},
						ExtAuth: &v1alpha1.ExtAuth{
							HTTP: &v1alpha1.HTTPExtAuthService{
								BackendRef: gwapiv1.BackendObjectReference{
									Name: "ext-auth-http-service",
								},
							},
						},
					},
				},
			},
			service: test.GetService(types.NamespacedName{Name: "ext-auth-http-service"}, nil, nil),
			expect:  true,
		},
		{
			name: "update status of all gateways under gatewayclass when MergeGateways enabled",
			configs: []client.Object{
				test.GetGatewayClass("test-mg", v1alpha1.GatewayControllerName, &test.GroupKindNamespacedName{
					Group:     gwapiv1.Group(mergeGatewaysConfig.GroupVersionKind().Group),
					Kind:      gwapiv1.Kind(mergeGatewaysConfig.Kind),
					Namespace: gwapiv1.Namespace(mergeGatewaysConfig.Namespace),
					Name:      gwapiv1.ObjectName(mergeGatewaysConfig.Name),
				}),
				mergeGatewaysConfig,
				test.GetGateway(types.NamespacedName{Name: "merged-gateway-1", Namespace: "default"}, "test-mg", 8081),
				test.GetGateway(types.NamespacedName{Name: "merged-gateway-2", Namespace: "default"}, "test-mg", 8082),
				test.GetGateway(types.NamespacedName{Name: "merged-gateway-3", Namespace: "default"}, "test-mg", 8083),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, map[string]string{
				gatewayapi.OwningGatewayClassLabel: "test-mg",
			}, nil),
			expect: false,
		},
		{
			name: "no gateways found under gatewayclass when MergeGateways enabled",
			configs: []client.Object{
				test.GetGatewayClass("test-mg", v1alpha1.GatewayControllerName, &test.GroupKindNamespacedName{
					Group:     gwapiv1.Group(mergeGatewaysConfig.GroupVersionKind().Group),
					Kind:      gwapiv1.Kind(mergeGatewaysConfig.Kind),
					Namespace: gwapiv1.Namespace(mergeGatewaysConfig.Namespace),
					Name:      gwapiv1.ObjectName(mergeGatewaysConfig.Name),
				}),
				mergeGatewaysConfig,
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, map[string]string{
				gatewayapi.OwningGatewayClassLabel: "test-mg",
			}, nil),
			expect: false,
		},
	}

	// Create the reconciler.
	logger := logging.DefaultLogger(v1alpha1.LogLevelInfo)

	r := gatewayAPIReconciler{
		classController: v1alpha1.GatewayControllerName,
		log:             logger,
		mergeGateways:   sets.New[string]("test-mg"),
	}

	for _, tc := range testCases {
		tc := tc
		r.client = fakeclient.NewClientBuilder().
			WithScheme(envoygateway.GetScheme()).
			WithObjects(tc.configs...).
			WithIndex(&gwapiv1.HTTPRoute{}, backendHTTPRouteIndex, backendHTTPRouteIndexFunc).
			WithIndex(&gwapiv1a2.GRPCRoute{}, backendGRPCRouteIndex, backendGRPCRouteIndexFunc).
			WithIndex(&gwapiv1a2.TLSRoute{}, backendTLSRouteIndex, backendTLSRouteIndexFunc).
			WithIndex(&gwapiv1a2.TCPRoute{}, backendTCPRouteIndex, backendTCPRouteIndexFunc).
			WithIndex(&gwapiv1a2.UDPRoute{}, backendUDPRouteIndex, backendUDPRouteIndexFunc).
			WithIndex(&v1alpha1.SecurityPolicy{}, backendSecurityPolicyIndex, backendSecurityPolicyIndexFunc).
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
	sampleGateway := test.GetGateway(types.NamespacedName{Namespace: "default", Name: "scheduled-status-test"}, "test-gc", 8080)
	mergeGatewaysConfig := test.GetEnvoyProxy(types.NamespacedName{Namespace: "default", Name: "merge-gateways-config"}, true)

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
				test.GetGatewayClass("test-gc", v1alpha1.GatewayControllerName, nil),
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
		{
			name: "update status of all gateways under gatewayclass when MergeGateways enabled",
			configs: []client.Object{
				test.GetGatewayClass("test-mg", v1alpha1.GatewayControllerName, &test.GroupKindNamespacedName{
					Group:     gwapiv1.Group(mergeGatewaysConfig.GroupVersionKind().Group),
					Kind:      gwapiv1.Kind(mergeGatewaysConfig.Kind),
					Namespace: gwapiv1.Namespace(mergeGatewaysConfig.Namespace),
					Name:      gwapiv1.ObjectName(mergeGatewaysConfig.Name),
				}),
				mergeGatewaysConfig,
			},
			deployment: test.GetGatewayDeployment(types.NamespacedName{Name: "deployment"}, map[string]string{
				gatewayapi.OwningGatewayClassLabel: "test-mg",
			}),
			expect: false,
		},
		{
			name: "no gateways found under gatewayclass when MergeGateways enabled",
			configs: []client.Object{
				test.GetGatewayClass("test-mg", v1alpha1.GatewayControllerName, &test.GroupKindNamespacedName{
					Group:     gwapiv1.Group(mergeGatewaysConfig.GroupVersionKind().Group),
					Kind:      gwapiv1.Kind(mergeGatewaysConfig.Kind),
					Namespace: gwapiv1.Namespace(mergeGatewaysConfig.Namespace),
					Name:      gwapiv1.ObjectName(mergeGatewaysConfig.Name),
				}),
				mergeGatewaysConfig,
				test.GetGateway(types.NamespacedName{Name: "merged-gateway-1", Namespace: "default"}, "test-mg", 8081),
				test.GetGateway(types.NamespacedName{Name: "merged-gateway-2", Namespace: "default"}, "test-mg", 8082),
				test.GetGateway(types.NamespacedName{Name: "merged-gateway-3", Namespace: "default"}, "test-mg", 8083),
			},
			deployment: test.GetGatewayDeployment(types.NamespacedName{Name: "deployment"}, map[string]string{
				gatewayapi.OwningGatewayClassLabel: "test-mg",
			}),
			expect: false,
		},
	}

	// Create the reconciler.
	logger := logging.DefaultLogger(v1alpha1.LogLevelInfo)

	r := gatewayAPIReconciler{
		classController: v1alpha1.GatewayControllerName,
		log:             logger,
		mergeGateways:   sets.New[string]("test-mg"),
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

func TestCheckObjectNamespaceLabels(t *testing.T) {
	matchExpressions := func(key string, operator metav1.LabelSelectorOperator, values []string) []metav1.LabelSelectorRequirement {
		return []metav1.LabelSelectorRequirement{{
			Key:      key,
			Operator: operator,
			Values:   values,
		}}
	}
	testCases := []struct {
		name            string
		object          client.Object
		reconcileLabels string
		ns              *corev1.Namespace
		expect          bool
	}{
		{
			name: "matching labels of namespace of the object is a subset of namespaceLabels",
			object: test.GetHTTPRoute(
				types.NamespacedName{
					Name:      "foo-route",
					Namespace: "foo",
				},
				"eg",
				types.NamespacedName{
					Name:      "foo-svc",
					Namespace: "foo",
				},
				8080,
			),
			ns: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
					Labels: map[string]string{
						"label-1": "",
					},
				},
			},
			reconcileLabels: "label-1",
			expect:          true,
		},
		{
			name: "non-matching labels of namespace of the object is a subset of namespaceLabels",
			object: test.GetHTTPRoute(
				types.NamespacedName{
					Name:      "bar-route",
					Namespace: "bar",
				},
				"eg",
				types.NamespacedName{
					Name:      "bar-svc",
					Namespace: "bar",
				},
				8080,
			),
			ns: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "bar",
					Labels: map[string]string{
						"label-2": "",
					},
				},
			},
			reconcileLabels: "label-1",
			expect:          false,
		},
		{
			name: "non-matching labels of namespace of the cluster-level object is a subset of namespaceLabels",
			object: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo-1",
					Labels: map[string]string{
						"label-1": "",
					},
				},
			},
			ns: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "bar-1",
					Labels: map[string]string{
						"label-1": "",
					},
				},
			},
			reconcileLabels: "label-1",
			expect:          false,
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
		r.client = fakeclient.NewClientBuilder().WithObjects(tc.ns).Build()
		r.namespaceLabel = &metav1.LabelSelector{MatchExpressions: matchExpressions(tc.reconcileLabels, metav1.LabelSelectorOpExists, []string{})}
		ok, err := r.checkObjectNamespaceLabels(tc.object)
		require.NoError(t, err)
		require.Equal(t, tc.expect, ok)
	}
}

func TestMatchLabelsAndExpressions(t *testing.T) {
	matchLabels := map[string]string{"foo": "bar"}
	matchExpressions := func(operator metav1.LabelSelectorOperator, values []string) []metav1.LabelSelectorRequirement {
		return []metav1.LabelSelectorRequirement{{
			Key:      "baz",
			Operator: operator,
			Values:   values,
		}}
	}

	tests := []struct {
		ls        *metav1.LabelSelector
		objLabels map[string]string
		want      bool
	}{
		{
			ls:        &metav1.LabelSelector{MatchLabels: matchLabels},
			objLabels: map[string]string{"foo": "bar"},
			want:      true,
		},
		{
			ls:        &metav1.LabelSelector{MatchLabels: matchLabels, MatchExpressions: matchExpressions(metav1.LabelSelectorOpIn, []string{"norf"})},
			objLabels: map[string]string{"foo": "bar", "baz": "norf"},
			want:      true,
		},
		{
			ls:        &metav1.LabelSelector{MatchExpressions: matchExpressions(metav1.LabelSelectorOpIn, []string{"norf"})},
			objLabels: map[string]string{"baz": "norf"},
			want:      true,
		},
		{
			ls:        &metav1.LabelSelector{MatchLabels: matchLabels, MatchExpressions: matchExpressions(metav1.LabelSelectorOpIn, []string{"norf", "qux"})},
			objLabels: map[string]string{"foo": "bar", "baz": "norf"},
			want:      true,
		},
		{
			ls:        &metav1.LabelSelector{MatchLabels: matchLabels, MatchExpressions: matchExpressions(metav1.LabelSelectorOpIn, []string{"norf", "qux"})},
			objLabels: map[string]string{"foo": "bar"},
			want:      false,
		},
		{
			ls:        &metav1.LabelSelector{MatchExpressions: matchExpressions(metav1.LabelSelectorOpNotIn, []string{"norf", "qux"})},
			objLabels: map[string]string{},
			want:      true,
		},
		{
			ls:        &metav1.LabelSelector{MatchExpressions: matchExpressions(metav1.LabelSelectorOpNotIn, []string{"norf", "qux"})},
			objLabels: map[string]string{"baz": "norf"},
			want:      false,
		},
		{
			ls:        &metav1.LabelSelector{MatchLabels: matchLabels, MatchExpressions: matchExpressions(metav1.LabelSelectorOpNotIn, []string{"norf", "qux"})},
			objLabels: map[string]string{"foo": "bar"},
			want:      true,
		},
		{
			ls:        &metav1.LabelSelector{MatchLabels: matchLabels, MatchExpressions: matchExpressions(metav1.LabelSelectorOpNotIn, []string{"norf", "qux"})},
			objLabels: map[string]string{"foo": "bar", "baz": "norf"},
			want:      false,
		},
		{
			ls:        &metav1.LabelSelector{MatchLabels: matchLabels, MatchExpressions: matchExpressions(metav1.LabelSelectorOpExists, []string{})},
			objLabels: map[string]string{"foo": "bar"},
			want:      false,
		},
		{
			ls:        &metav1.LabelSelector{MatchLabels: matchLabels, MatchExpressions: matchExpressions(metav1.LabelSelectorOpExists, []string{})},
			objLabels: map[string]string{"foo": "bar", "baz": "1111"},
			want:      true,
		},
		{
			ls:        &metav1.LabelSelector{MatchLabels: matchLabels, MatchExpressions: matchExpressions(metav1.LabelSelectorOpDoesNotExist, []string{})},
			objLabels: map[string]string{"foo": "bar", "baz": "1111"},
			want:      false,
		},
		{
			ls:        &metav1.LabelSelector{MatchExpressions: matchExpressions(metav1.LabelSelectorOpDoesNotExist, []string{})},
			objLabels: map[string]string{"baz": "1111"},
			want:      false,
		},
		{
			ls:        &metav1.LabelSelector{MatchExpressions: matchExpressions(metav1.LabelSelectorOpDoesNotExist, []string{})},
			objLabels: map[string]string{"bazz": "1111"},
			want:      true,
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			if got := matchLabelsAndExpressions(tc.ls, tc.objLabels); got != tc.want {
				t.Errorf("ExtractMatchedSelectorInfo() = %v, want %v", got, tc.want)
			}
		})
	}
}
