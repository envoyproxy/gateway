// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	certificatesv1b1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes/test"
)

// TestGatewayClassHasMatchingController tests the hasMatchingController
// predicate function.
func TestGatewayClassHasMatchingController(t *testing.T) {
	testCases := []struct {
		name   string
		gc     *gwapiv1.GatewayClass
		client client.Client
		expect bool
	}{
		{
			name:   "matching controller name",
			gc:     test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
			expect: true,
		},
		{
			name:   "non-matching controller name",
			gc:     test.GetGatewayClass("test-gc", "not.configured/controller", nil),
			expect: false,
		},
	}

	// Create the reconciler.
	logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

	r := gatewayAPIReconciler{
		classController: egv1a1.GatewayControllerName,
		log:             logger,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := r.hasMatchingController(tc.gc)
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

	logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

	for _, tc := range testCases {
		r := gatewayAPIReconciler{
			classController: egv1a1.GatewayControllerName,
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
			sampleServiceBackendRef := test.GetServiceBackendRef(types.NamespacedName{Name: "service"}, 80)
			res := r.hasMatchingNamespaceLabels(
				test.GetHTTPRoute(types.NamespacedName{
					Namespace: ns,
					Name:      "httproute-test",
				}, "scheduled-status-test", sampleServiceBackendRef, ""))
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
			configs: []client.Object{test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil)},
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
	logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

	r := gatewayAPIReconciler{
		classController: egv1a1.GatewayControllerName,
		log:             logger,
	}

	for _, tc := range testCases {
		r.client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.configs...).Build()
		t.Run(tc.name, func(t *testing.T) {
			res := r.validateGatewayForReconcile(tc.gateway)
			require.Equal(t, tc.expect, res)
		})
	}
}

// TestValidateConfigMapForReconcile tests the validateConfigMapForReconcile
// predicate function.
func TestValidateConfigMapForReconcile(t *testing.T) {
	testCases := []struct {
		name      string
		configs   []client.Object
		configMap client.Object
		expect    bool
	}{
		{
			name: "references EnvoyExtensionPolicy Lua config map",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				test.GetGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc", 8080),
				&egv1a1.EnvoyExtensionPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "lua-cm",
						Namespace: "test",
					},
					Spec: egv1a1.EnvoyExtensionPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
								{
									LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
										Kind: "Gateway",
										Name: "scheduled-status-test",
									},
								},
							},
						},
						Lua: []egv1a1.Lua{
							{
								Type: egv1a1.LuaValueTypeValueRef,
								ValueRef: &gwapiv1.LocalObjectReference{
									Kind:  gwapiv1a2.Kind("ConfigMap"),
									Name:  gwapiv1a2.ObjectName("lua"),
									Group: gwapiv1a2.Group("v1"),
								},
							},
						},
					},
				},
			},
			configMap: test.GetConfigMap(types.NamespacedName{Name: "lua", Namespace: "test"}, make(map[string]string), make(map[string]string)),
			expect:    true,
		},
		{
			name: "does not reference EnvoyExtensionPolicy Lua config map",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				test.GetGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc", 8080),
				&egv1a1.EnvoyExtensionPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "lua-cm",
						Namespace: "test",
					},
					Spec: egv1a1.EnvoyExtensionPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
								{
									LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
										Kind: "Gateway",
										Name: "scheduled-status-test",
									},
								},
							},
						},
						Lua: []egv1a1.Lua{
							{
								Type: egv1a1.LuaValueTypeValueRef,
								ValueRef: &gwapiv1.LocalObjectReference{
									Kind:  gwapiv1a2.Kind("ConfigMap"),
									Name:  gwapiv1a2.ObjectName("lua"),
									Group: gwapiv1a2.Group("v1"),
								},
							},
						},
					},
				},
			},
			configMap: test.GetConfigMap(types.NamespacedName{Name: "not-lua", Namespace: "test"}, make(map[string]string), make(map[string]string)),
			expect:    false,
		},
	}

	// Create the reconciler.
	logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

	r := gatewayAPIReconciler{
		classController: egv1a1.GatewayControllerName,
		log:             logger,
		spCRDExists:     true,
		epCRDExists:     true,
		eepCRDExists:    true,
	}

	for _, tc := range testCases {
		r.client = fakeclient.NewClientBuilder().
			WithScheme(envoygateway.GetScheme()).
			WithObjects(tc.configs...).
			WithIndex(&egv1a1.EnvoyExtensionPolicy{}, configMapEepIndex, configMapEepIndexFunc).
			Build()
		t.Run(tc.name, func(t *testing.T) {
			res := r.validateConfigMapForReconcile(tc.configMap)
			require.Equal(t, tc.expect, res)
		})
	}
}

// TestValidateSecretForReconcile tests the validateSecretForReconcile
// predicate function.
func TestValidateSecretForReconcile(t *testing.T) {
	mtlsEnabledEnvoyProxyConfig := &egv1a1.EnvoyProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "mtls-settings",
		},
		Spec: egv1a1.EnvoyProxySpec{
			BackendTLS: &egv1a1.BackendTLSConfig{
				ClientCertificateRef: &gwapiv1.SecretObjectReference{
					Kind: gatewayapi.KindPtr("Secret"),
					Name: "client-tls-certificate",
				},
				TLSSettings: egv1a1.TLSSettings{},
			},
		},
	}
	testCases := []struct {
		name    string
		configs []client.Object
		secret  client.Object
		expect  bool
	}{
		{
			name: "envoy proxy references a secret",
			configs: []client.Object{
				test.GetGatewayClass("test-secret-ref", egv1a1.GatewayControllerName, &test.GroupKindNamespacedName{
					Group:     gwapiv1.Group(mtlsEnabledEnvoyProxyConfig.GroupVersionKind().Group),
					Kind:      gwapiv1.Kind(mtlsEnabledEnvoyProxyConfig.Kind),
					Namespace: gwapiv1.Namespace(mtlsEnabledEnvoyProxyConfig.Namespace),
					Name:      gwapiv1.ObjectName(mtlsEnabledEnvoyProxyConfig.Name),
				}),
				test.GetSecret(types.NamespacedName{Namespace: mtlsEnabledEnvoyProxyConfig.Namespace, Name: "client-tls-certificate"}),
				mtlsEnabledEnvoyProxyConfig,
			},
			secret: test.GetSecret(types.NamespacedName{Namespace: mtlsEnabledEnvoyProxyConfig.Namespace, Name: "client-tls-certificate"}),
			expect: true,
		},
		{
			name: "references valid gateway",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				test.GetSecureGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc", test.GroupKindNamespacedName{
					Kind: resource.KindSecret,
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
					Kind: resource.KindSecret,
					Name: "secret",
				}),
			},
			secret: test.GetSecret(types.NamespacedName{Name: "secret"}),
			expect: false,
		},
		{
			name: "references SecurityPolicy OIDC",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				test.GetGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc", 8080),
				&egv1a1.SecurityPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "oidc",
					},
					Spec: egv1a1.SecurityPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Kind: "Gateway",
									Name: "scheduled-status-test",
								},
							},
						},
						OIDC: &egv1a1.OIDC{
							Provider: egv1a1.OIDCProvider{
								Issuer:                "https://accounts.google.com",
								AuthorizationEndpoint: ptr.To("https://accounts.google.com/o/oauth2/v2/auth"),
								TokenEndpoint:         ptr.To("https://oauth2.googleapis.com/token"),
							},
							ClientID: ptr.To("client-id"),
							ClientSecret: gwapiv1.SecretObjectReference{
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
			name: "references SecurityPolicy APIKey Auth",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				test.GetGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc", 8080),
				&egv1a1.SecurityPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "apikey-auth",
					},
					Spec: egv1a1.SecurityPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Kind: "Gateway",
									Name: "scheduled-status-test",
								},
							},
						},
						APIKeyAuth: &egv1a1.APIKeyAuth{
							CredentialRefs: []gwapiv1.SecretObjectReference{
								{
									Name: "secret",
								},
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
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				test.GetGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc", 8080),
				&egv1a1.SecurityPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "basic-auth",
					},
					Spec: egv1a1.SecurityPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Kind: "Gateway",
									Name: "scheduled-status-test",
								},
							},
						},
						BasicAuth: &egv1a1.BasicAuth{
							Users: gwapiv1.SecretObjectReference{
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
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
			},
			secret: test.GetSecret(types.NamespacedName{Name: "secret"}),
			expect: false,
		},
		{
			name: "references EnvoyExtensionPolicy Wasm OCI Image",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				test.GetGateway(types.NamespacedName{Name: "scheduled-status-test"}, "test-gc", 8080),
				&egv1a1.EnvoyExtensionPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "wasm-oci",
					},
					Spec: egv1a1.EnvoyExtensionPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
								{
									LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
										Kind: "Gateway",
										Name: "scheduled-status-test",
									},
								},
							},
						},
						Wasm: []egv1a1.Wasm{
							{
								Name:   ptr.To("wasm-filter"),
								RootID: ptr.To("my_root_id"),
								Code: egv1a1.WasmCodeSource{
									Type: egv1a1.ImageWasmCodeSourceType,
									Image: &egv1a1.ImageWasmCodeSource{
										URL: "https://example.com/testwasm:v1.0.0",
										PullSecretRef: &gwapiv1.SecretObjectReference{
											Name: "secret",
										},
									},
								},
							},
						},
					},
				},
			},
			secret: test.GetSecret(types.NamespacedName{Name: "secret"}),
			expect: true,
		},
	}

	// Create the reconciler.
	logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

	r := gatewayAPIReconciler{
		classController: egv1a1.GatewayControllerName,
		log:             logger,
		spCRDExists:     true,
		epCRDExists:     true,
		eepCRDExists:    true,
	}

	for _, tc := range testCases {
		r.client = fakeclient.NewClientBuilder().
			WithScheme(envoygateway.GetScheme()).
			WithObjects(tc.configs...).
			WithIndex(&gwapiv1.Gateway{}, secretGatewayIndex, secretGatewayIndexFunc).
			WithIndex(&egv1a1.SecurityPolicy{}, secretSecurityPolicyIndex, secretSecurityPolicyIndexFunc).
			WithIndex(&egv1a1.EnvoyProxy{}, secretEnvoyProxyIndex, secretEnvoyProxyIndexFunc).
			WithIndex(&egv1a1.EnvoyExtensionPolicy{}, secretEnvoyExtensionPolicyIndex, secretEnvoyExtensionPolicyIndexFunc).
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
	sampleGatewayClass := test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil)
	sampleGateway := test.GetGateway(types.NamespacedName{Namespace: "default", Name: "scheduled-status-test"}, "test-gc", 8080)
	sampleServiceBackendRef := test.GetServiceBackendRef(types.NamespacedName{Name: "service"}, 80)
	sampleServiceImportBackendRef := test.GetServiceImportBackendRef(types.NamespacedName{Name: "imported-service"}, 80)

	testCases := []struct {
		name          string
		configs       []client.Object
		endpointSlice client.Object
		expect        bool
	}{
		{
			name: "route service but no routes exist",
			configs: []client.Object{
				sampleGatewayClass,
				sampleGateway,
			},
			endpointSlice: test.GetEndpointSlice(types.NamespacedName{Name: "endpointslice"}, "service", false),
			expect:        false,
		},
		{
			name: "http route service routes exist, but endpointslice is associated with another service",
			configs: []client.Object{
				sampleGatewayClass,
				sampleGateway,
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test", sampleServiceBackendRef, ""),
			},
			endpointSlice: test.GetEndpointSlice(types.NamespacedName{Name: "endpointslice"}, "other-service", false),
			expect:        false,
		},
		{
			name: "http route service routes exist",
			configs: []client.Object{
				sampleGatewayClass,
				sampleGateway,
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test", sampleServiceBackendRef, ""),
			},
			endpointSlice: test.GetEndpointSlice(types.NamespacedName{Name: "endpointslice"}, "service", false),
			expect:        true,
		},
		{
			name: "http route serviceimport routes exist",
			configs: []client.Object{
				sampleGatewayClass,
				sampleGateway,
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test", sampleServiceImportBackendRef, ""),
			},
			endpointSlice: test.GetEndpointSlice(types.NamespacedName{Name: "endpointslice"}, "imported-service", true),
			expect:        true,
		},
		{
			name: "mirrored backend route exists",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				sampleGateway,
				&gwapiv1.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name: "httproute-test",
					},
					Spec: gwapiv1.HTTPRouteSpec{
						CommonRouteSpec: gwapiv1.CommonRouteSpec{
							ParentRefs: []gwapiv1.ParentReference{
								{Name: gwapiv1.ObjectName("scheduled-status-test")},
							},
						},
						Rules: []gwapiv1.HTTPRouteRule{
							{
								BackendRefs: []gwapiv1.HTTPBackendRef{
									{
										BackendRef: gwapiv1.BackendRef{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Name: gwapiv1.ObjectName("service"),
												Port: ptr.To(gwapiv1.PortNumber(80)),
											},
										},
									},
								},
								Filters: []gwapiv1.HTTPRouteFilter{
									{
										Type: gwapiv1.HTTPRouteFilterRequestMirror,
										RequestMirror: &gwapiv1.HTTPRequestMirrorFilter{
											BackendRef: gwapiv1.BackendObjectReference{
												Name: "mirror-service",
												Port: ptr.To(gwapiv1.PortNumber(80)),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			endpointSlice: test.GetEndpointSlice(types.NamespacedName{Name: "endpointslice"}, "mirror-service", false),
			expect:        true,
		},
	}

	// Create the reconciler.
	logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

	r := gatewayAPIReconciler{
		classController: egv1a1.GatewayControllerName,
		log:             logger,
	}

	for _, tc := range testCases {
		r.client = fakeclient.NewClientBuilder().
			WithScheme(envoygateway.GetScheme()).
			WithObjects(tc.configs...).
			WithIndex(&gwapiv1.HTTPRoute{}, backendHTTPRouteIndex, backendHTTPRouteIndexFunc).
			WithIndex(&gwapiv1.GRPCRoute{}, backendGRPCRouteIndex, backendGRPCRouteIndexFunc).
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
	telemetryEnabledGatewaysConfig := &egv1a1.EnvoyProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "telemetry",
		},
		Spec: egv1a1.EnvoyProxySpec{
			Telemetry: &egv1a1.ProxyTelemetry{
				AccessLog: &egv1a1.ProxyAccessLog{
					Settings: []egv1a1.ProxyAccessLogSetting{
						{
							Sinks: []egv1a1.ProxyAccessLogSink{
								{
									Type: egv1a1.ProxyAccessLogSinkTypeOpenTelemetry,
									OpenTelemetry: &egv1a1.OpenTelemetryEnvoyProxyAccessLog{
										BackendCluster: egv1a1.BackendCluster{
											BackendRefs: []egv1a1.BackendRef{
												{
													BackendObjectReference: gwapiv1.BackendObjectReference{
														Name:      "otel-collector",
														Namespace: ptr.To(gwapiv1.Namespace("default")),
														Port:      ptr.To(gwapiv1.PortNumber(4317)),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Metrics: &egv1a1.ProxyMetrics{
					Sinks: []egv1a1.ProxyMetricSink{
						{
							Type: egv1a1.MetricSinkTypeOpenTelemetry,
							OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{
								BackendCluster: egv1a1.BackendCluster{
									BackendRefs: []egv1a1.BackendRef{
										{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Name:      "otel-collector",
												Namespace: ptr.To(gwapiv1.Namespace("default")),
												Port:      ptr.To(gwapiv1.PortNumber(4317)),
											},
										},
									},
								},
							},
						},
					},
				},
				Tracing: &egv1a1.ProxyTracing{
					Provider: egv1a1.TracingProvider{
						Type: egv1a1.TracingProviderTypeOpenTelemetry,
						BackendCluster: egv1a1.BackendCluster{
							BackendRefs: []egv1a1.BackendRef{
								{
									BackendObjectReference: gwapiv1.BackendObjectReference{
										Name:      "otel-collector",
										Namespace: ptr.To(gwapiv1.Namespace("default")),
										Port:      ptr.To(gwapiv1.PortNumber(4317)),
									},
								},
							},
						},
					},
				},
			},
		},
	}
	sampleServiceBackendRef := test.GetServiceBackendRef(types.NamespacedName{Name: "service"}, 80)

	testCases := []struct {
		name    string
		configs []client.Object
		service client.Object
		expect  bool
	}{
		{
			name: "gateway service but deployment or daemonset does not exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
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
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				sampleGateway,
				test.GetGatewayDeployment(types.NamespacedName{Name: proxy.ExpectedResourceHashedName("default/scheduled-status-test")}, nil),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, map[string]string{
				gatewayapi.OwningGatewayNameLabel:      "scheduled-status-test",
				gatewayapi.OwningGatewayNamespaceLabel: "default",
			}, nil),
			// Note that in case when a envoyObjects exists, the Service is just processed for Gateway status
			// updates and not reconciled further.
			expect: false,
		},
		{
			name: "gateway service daemonset also exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				sampleGateway,
				test.GetGatewayDaemonSet(types.NamespacedName{Name: proxy.ExpectedResourceHashedName("default/scheduled-status-test")}, nil),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, map[string]string{
				gatewayapi.OwningGatewayNameLabel:      "scheduled-status-test",
				gatewayapi.OwningGatewayNamespaceLabel: "default",
			}, nil),
			// Note that in case when a envoyObjects exists, the Service is just processed for Gateway status
			// updates and not reconciled further.
			expect: false,
		},
		{
			name: "route service but no routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				sampleGateway,
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  false,
		},
		{
			name: "http route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				sampleGateway,
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test", sampleServiceBackendRef, ""),
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
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test", sampleServiceBackendRef, ""),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  true,
		},
		{
			name: "grpc route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				sampleGateway,
				test.GetGRPCRoute(types.NamespacedName{Name: "grpcroute-test"}, "scheduled-status-test", types.NamespacedName{Name: "service"}, 80),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  true,
		},
		{
			name: "tls route service routes exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
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
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
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
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				sampleGateway,
				test.GetTCPRoute(types.NamespacedName{Name: "tcproute-test"}, "scheduled-status-test",
					types.NamespacedName{Name: "service"}, 80),
			},
			service: test.GetService(types.NamespacedName{Name: "service"}, nil, nil),
			expect:  true,
		},
		{
			name: "service referenced by EnvoyProxy updated",
			configs: []client.Object{
				test.GetGatewayClass("test-mg", egv1a1.GatewayControllerName, &test.GroupKindNamespacedName{
					Group:     gwapiv1.Group(telemetryEnabledGatewaysConfig.GroupVersionKind().Group),
					Kind:      gwapiv1.Kind(telemetryEnabledGatewaysConfig.Kind),
					Namespace: gwapiv1.Namespace(telemetryEnabledGatewaysConfig.Namespace),
					Name:      gwapiv1.ObjectName(telemetryEnabledGatewaysConfig.Name),
				}),
				telemetryEnabledGatewaysConfig,
			},
			service: test.GetService(types.NamespacedName{Name: "otel-collector", Namespace: "default"}, nil, nil),
			expect:  true,
		},
		{
			name: "service referenced by EnvoyProxy unrelated",
			configs: []client.Object{
				test.GetGatewayClass("test-mg", egv1a1.GatewayControllerName, &test.GroupKindNamespacedName{
					Group:     gwapiv1.Group(telemetryEnabledGatewaysConfig.GroupVersionKind().Group),
					Kind:      gwapiv1.Kind(telemetryEnabledGatewaysConfig.Kind),
					Namespace: gwapiv1.Namespace(telemetryEnabledGatewaysConfig.Namespace),
					Name:      gwapiv1.ObjectName(telemetryEnabledGatewaysConfig.Name),
				}),
				telemetryEnabledGatewaysConfig,
			},
			service: test.GetService(types.NamespacedName{Name: "otel-collector-unrelated", Namespace: "default"}, nil, nil),
			expect:  false,
		},
		{
			name: "service referenced by SecurityPolicy ExtAuth HTTP service",
			configs: []client.Object{
				&egv1a1.SecurityPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ext-auth-http",
					},
					Spec: egv1a1.SecurityPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Kind: "Gateway",
									Name: "scheduled-status-test",
								},
							},
						},
						ExtAuth: &egv1a1.ExtAuth{
							HTTP: &egv1a1.HTTPExtAuthService{
								BackendCluster: egv1a1.BackendCluster{
									BackendRefs: []egv1a1.BackendRef{
										{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Name: "ext-auth-http-service",
											},
										},
									},
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
			name: "service referenced by SecurityPolicy ExtAuth GRPC service",
			configs: []client.Object{
				&egv1a1.SecurityPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ext-auth-http",
					},
					Spec: egv1a1.SecurityPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Kind: "Gateway",
									Name: "scheduled-status-test",
								},
							},
						},
						ExtAuth: &egv1a1.ExtAuth{
							GRPC: &egv1a1.GRPCExtAuthService{
								BackendCluster: egv1a1.BackendCluster{
									BackendRefs: []egv1a1.BackendRef{
										{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Name: "ext-auth-grpc-service",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			service: test.GetService(types.NamespacedName{Name: "ext-auth-grpc-service"}, nil, nil),
			expect:  true,
		},
		{
			name: "service referenced by EnvoyExtensionPolicy ExtPrc GRPC service",
			configs: []client.Object{
				&egv1a1.EnvoyExtensionPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ext-proc",
					},
					Spec: egv1a1.EnvoyExtensionPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Kind: "Gateway",
									Name: "scheduled-status-test",
								},
							},
						},
						ExtProc: []egv1a1.ExtProc{
							{
								BackendCluster: egv1a1.BackendCluster{
									BackendRefs: []egv1a1.BackendRef{
										{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Name: "ext-proc-service",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			service: test.GetService(types.NamespacedName{Name: "ext-proc-service"}, nil, nil),
			expect:  true,
		},
		{
			name: "service referenced by EnvoyExtensionPolicy ExtPrc GRPC service unrelated",
			configs: []client.Object{
				&egv1a1.EnvoyExtensionPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ext-proc",
					},
					Spec: egv1a1.EnvoyExtensionPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Kind: "Gateway",
									Name: "scheduled-status-test",
								},
							},
						},
						ExtProc: []egv1a1.ExtProc{
							{
								BackendCluster: egv1a1.BackendCluster{
									BackendRefs: []egv1a1.BackendRef{
										{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Name: "ext-proc-service",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			service: test.GetService(types.NamespacedName{Name: "ext-proc-service-unrelated"}, nil, nil),
			expect:  false,
		},
		{
			name: "update status of all gateways under gatewayclass when MergeGateways enabled",
			configs: []client.Object{
				test.GetGatewayClass("test-mg", egv1a1.GatewayControllerName, &test.GroupKindNamespacedName{
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
				test.GetGatewayClass("test-mg", egv1a1.GatewayControllerName, &test.GroupKindNamespacedName{
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
	logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

	r := gatewayAPIReconciler{
		classController:    egv1a1.GatewayControllerName,
		log:                logger,
		mergeGateways:      sets.New[string]("test-mg"),
		resources:          &message.ProviderResources{},
		grpcRouteCRDExists: true,
		tcpRouteCRDExists:  true,
		udpRouteCRDExists:  true,
		tlsRouteCRDExists:  true,
		spCRDExists:        true,
		eepCRDExists:       true,
		epCRDExists:        true,
	}

	for _, tc := range testCases {
		r.client = fakeclient.NewClientBuilder().
			WithScheme(envoygateway.GetScheme()).
			WithObjects(tc.configs...).
			WithIndex(&gwapiv1.HTTPRoute{}, backendHTTPRouteIndex, backendHTTPRouteIndexFunc).
			WithIndex(&gwapiv1.GRPCRoute{}, backendGRPCRouteIndex, backendGRPCRouteIndexFunc).
			WithIndex(&gwapiv1a2.TLSRoute{}, backendTLSRouteIndex, backendTLSRouteIndexFunc).
			WithIndex(&gwapiv1a2.TCPRoute{}, backendTCPRouteIndex, backendTCPRouteIndexFunc).
			WithIndex(&gwapiv1a2.UDPRoute{}, backendUDPRouteIndex, backendUDPRouteIndexFunc).
			WithIndex(&egv1a1.SecurityPolicy{}, backendSecurityPolicyIndex, backendSecurityPolicyIndexFunc).
			WithIndex(&egv1a1.EnvoyExtensionPolicy{}, backendEnvoyExtensionPolicyIndex, backendEnvoyExtensionPolicyIndexFunc).
			WithIndex(&egv1a1.EnvoyProxy{}, backendEnvoyProxyTelemetryIndex, backendEnvoyProxyTelemetryIndexFunc).
			Build()
		t.Run(tc.name, func(t *testing.T) {
			res := r.validateServiceForReconcile(tc.service)
			require.Equal(t, tc.expect, res)
		})
	}
}

// TestValidateObjectForReconcile tests the validateObjectForReconcile
// predicate function.
func TestValidateObjectForReconcile(t *testing.T) {
	sampleGateway := test.GetGateway(types.NamespacedName{Namespace: "default", Name: "scheduled-status-test"}, "test-gc", 8080)
	mergeGatewaysConfig := test.GetEnvoyProxy(types.NamespacedName{Namespace: "default", Name: "merge-gateways-config"}, true)

	testCases := []struct {
		name         string
		configs      []client.Object
		envoyObjects []client.Object
		expect       bool
	}{
		{
			// No config should lead to a reconciliation of a Deployment or DaemonSet object. The main
			// purpose of the watcher is just for updating Gateway object statuses.
			name: "gateway deployment or daemonset also exist",
			configs: []client.Object{
				test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil),
				sampleGateway,
				test.GetService(types.NamespacedName{Name: "envoyObjects"}, map[string]string{
					gatewayapi.OwningGatewayNameLabel:      "scheduled-status-test",
					gatewayapi.OwningGatewayNamespaceLabel: "default",
				}, nil),
			},
			envoyObjects: []client.Object{
				test.GetGatewayDeployment(types.NamespacedName{Name: "deployment"}, map[string]string{
					gatewayapi.OwningGatewayNameLabel:      "scheduled-status-test",
					gatewayapi.OwningGatewayNamespaceLabel: "default",
				}), test.GetGatewayDaemonSet(types.NamespacedName{Name: "daemonset"}, map[string]string{
					gatewayapi.OwningGatewayNameLabel:      "scheduled-status-test",
					gatewayapi.OwningGatewayNamespaceLabel: "default",
				}),
			},
			expect: false,
		},
		{
			name: "update status of all gateways under gatewayclass when MergeGateways enabled",
			configs: []client.Object{
				test.GetGatewayClass("test-mg", egv1a1.GatewayControllerName, &test.GroupKindNamespacedName{
					Group:     gwapiv1.Group(mergeGatewaysConfig.GroupVersionKind().Group),
					Kind:      gwapiv1.Kind(mergeGatewaysConfig.Kind),
					Namespace: gwapiv1.Namespace(mergeGatewaysConfig.Namespace),
					Name:      gwapiv1.ObjectName(mergeGatewaysConfig.Name),
				}),
				mergeGatewaysConfig,
			},
			envoyObjects: []client.Object{
				test.GetGatewayDeployment(types.NamespacedName{Name: "deployment"}, map[string]string{
					gatewayapi.OwningGatewayClassLabel: "test-mg",
				}),
				test.GetGatewayDaemonSet(types.NamespacedName{Name: "daemonset"}, map[string]string{
					gatewayapi.OwningGatewayClassLabel: "test-mg",
				}),
			},
			expect: false,
		},
		{
			name: "no gateways found under gatewayclass when MergeGateways enabled",
			configs: []client.Object{
				test.GetGatewayClass("test-mg", egv1a1.GatewayControllerName, &test.GroupKindNamespacedName{
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
			envoyObjects: []client.Object{
				test.GetGatewayDeployment(types.NamespacedName{Name: "deployment"}, map[string]string{
					gatewayapi.OwningGatewayClassLabel: "test-mg",
				}),
				test.GetGatewayDaemonSet(types.NamespacedName{Name: "daemonset"}, map[string]string{
					gatewayapi.OwningGatewayClassLabel: "test-mg",
				}),
			},
			expect: false,
		},
	}

	// Create the reconciler.
	logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

	r := gatewayAPIReconciler{
		classController: egv1a1.GatewayControllerName,
		log:             logger,
		mergeGateways:   sets.New[string]("test-mg"),
		resources:       &message.ProviderResources{},
	}

	for _, tc := range testCases {
		r.client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.configs...).Build()
		t.Run(tc.name, func(t *testing.T) {
			for _, obj := range tc.envoyObjects {
				res := r.validateObjectForReconcile(obj)
				require.Equal(t, tc.expect, res)
			}
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
				test.GetServiceBackendRef(types.NamespacedName{
					Name:      "foo-svc",
					Namespace: "foo",
				}, 8080),
				""),
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
				test.GetServiceBackendRef(types.NamespacedName{
					Name:      "bar-svc",
					Namespace: "bar",
				}, 8080),
				""),
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
	logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

	r := gatewayAPIReconciler{
		classController: egv1a1.GatewayControllerName,
		log:             logger,
	}

	for _, tc := range testCases {
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

// TestValidateHTTPRouteFilerForReconcile tests the vlidateHTTPRouteFilerForReconcile
// predicate function.
func TestValidateHTTPRouteFilerForReconcile(t *testing.T) {
	sampleGWC := test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil)
	sampleGateway := test.GetGateway(types.NamespacedName{Namespace: "default", Name: "scheduled-status-test"}, "test-gc", 8080)
	sampleService := test.GetService(types.NamespacedName{Name: "service"}, nil, nil)
	sampleServiceBackendRef := test.GetServiceBackendRef(types.NamespacedName{Name: "service"}, 80)
	sampleHTTPRouteFilter := test.GetHTTPRouteFilter(types.NamespacedName{Name: "httproutefilter"})

	testCases := []struct {
		name            string
		configs         []client.Object
		httpRouteFilter client.Object
		expect          bool
	}{
		{
			name: "httproutefilter but not referenced by route",
			configs: []client.Object{
				sampleGWC,
				sampleGateway,
				sampleService,
				sampleHTTPRouteFilter,
			},
			httpRouteFilter: sampleHTTPRouteFilter,
			expect:          false,
		},
		{
			name: "httproutefitler referenced by route",
			configs: []client.Object{
				sampleGWC,
				sampleGateway,
				sampleService,
				sampleHTTPRouteFilter,
				test.GetHTTPRoute(types.NamespacedName{Name: "httproute-test"}, "scheduled-status-test", sampleServiceBackendRef, "httproutefilter"),
			},
			httpRouteFilter: sampleHTTPRouteFilter,
			expect:          true,
		},
	}

	// Create the reconciler.
	logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

	r := gatewayAPIReconciler{
		classController: egv1a1.GatewayControllerName,
		log:             logger,
	}

	for _, tc := range testCases {
		r.client = fakeclient.NewClientBuilder().
			WithScheme(envoygateway.GetScheme()).
			WithObjects(tc.configs...).
			WithIndex(&gwapiv1.HTTPRoute{}, backendHTTPRouteIndex, backendHTTPRouteIndexFunc).
			WithIndex(&gwapiv1.HTTPRoute{}, httpRouteFilterHTTPRouteIndex, httpRouteFilterHTTPRouteIndexFunc).
			Build()
		t.Run(tc.name, func(t *testing.T) {
			res := r.validateHTTPRouteFilterForReconcile(tc.httpRouteFilter)
			require.Equal(t, tc.expect, res)
		})
	}
}

func TestValidateClusterTrustBundleForReconcile(t *testing.T) {
	gc := test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil)
	gtw := test.GetGateway(types.NamespacedName{Namespace: "default", Name: "scheduled-status-test"}, "test-gc", 8080)
	ctb := test.GetClusterTrustBundle("fake-ctb")
	backend := &egv1a1.Backend{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "backend-dynamic-resolver-clustertrustbundle",
			Namespace: "default",
		},
		Spec: egv1a1.BackendSpec{
			Type: ptr.To(egv1a1.BackendTypeDynamicResolver),
			TLS: &egv1a1.BackendTLSSettings{
				CACertificateRefs: []gwapiv1.LocalObjectReference{
					{
						Kind: gwapiv1.Kind("ClusterTrustBundle"),
						Name: gwapiv1.ObjectName(ctb.Name),
					},
				},
			},
		},
	}
	btp := &gwapiv1a3.BackendTLSPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "backend-tls-policy-dynamic-resolver-clustertrustbundle",
			Namespace: "default",
		},
		Spec: gwapiv1a3.BackendTLSPolicySpec{
			Validation: gwapiv1a3.BackendTLSPolicyValidation{
				CACertificateRefs: []gwapiv1.LocalObjectReference{
					{
						Kind: gwapiv1.Kind("ClusterTrustBundle"),
						Name: gwapiv1.ObjectName(ctb.Name),
					},
				},
			},
		},
	}
	ctp := test.GetClientTrafficPolicy(
		types.NamespacedName{Name: "fake-ctp", Namespace: "default"},
		&egv1a1.ClientTLSSettings{
			ClientValidation: &egv1a1.ClientValidationContext{
				CACertificateRefs: []gwapiv1.SecretObjectReference{
					{
						Kind: ptr.To[gwapiv1.Kind]("ClusterTrustBundle"),
						Name: gwapiv1.ObjectName(ctb.Name),
					},
				},
			},
		})

	testCases := []struct {
		name    string
		configs []client.Object
		ctb     *certificatesv1b1.ClusterTrustBundle
		expect  bool
	}{
		{
			name: "referenced by Backend",
			configs: []client.Object{
				gc,
				gtw,
				backend,
			},
			ctb:    ctb,
			expect: true,
		},
		{
			name: "referenced by BackendTLSPolicy",
			configs: []client.Object{
				gc,
				gtw,
				btp,
			},
			ctb:    ctb,
			expect: true,
		},
		{
			name: "referenced by ClientTrafficPolicy",
			configs: []client.Object{
				gc,
				gtw,
				ctp,
			},
			ctb:    ctb,
			expect: true,
		},
		{
			name: "ClusterTrustBundle not referenced",
			configs: []client.Object{
				gc,
				gtw,
			},
			ctb:    ctb,
			expect: false,
		},
	}

	// Create the reconciler.
	logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

	r := gatewayAPIReconciler{
		classController:     egv1a1.GatewayControllerName,
		log:                 logger,
		backendCRDExists:    true,
		bTLSPolicyCRDExists: true,
		ctpCRDExists:        true,
	}

	for _, tc := range testCases {
		r.client = fakeclient.NewClientBuilder().
			WithScheme(envoygateway.GetScheme()).
			WithObjects(tc.configs...).
			WithIndex(&egv1a1.Backend{}, clusterTrustBundleBackendIndex, clusterTrustBundleBackendIndexFunc).
			WithIndex(&gwapiv1a3.BackendTLSPolicy{}, clusterTrustBundleBtlsIndex, clusterTrustBundleBtlsIndexFunc).
			WithIndex(&egv1a1.ClientTrafficPolicy{}, clusterTrustBundleCtpIndex, clusterTrustBundleCtpIndexFunc).
			Build()
		t.Run(tc.name, func(t *testing.T) {
			res := r.validateClusterTrustBundleForReconcile(tc.ctb)
			require.Equal(t, tc.expect, res)
		})
	}
}
