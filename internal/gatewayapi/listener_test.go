// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
)

// expectedListenerStatus defines the expected status for a listener in the test
type expectedListenerStatus struct {
	listenerName string
	condition    gwapiv1.ListenerConditionType
	status       metav1.ConditionStatus
	reason       gwapiv1.ListenerConditionReason
	message      string
}

func TestProxySamplingRate(t *testing.T) {
	cases := []struct {
		name     string
		tracing  *egv1a1.ProxyTracing
		expected float64
	}{
		{
			name:     "default",
			tracing:  &egv1a1.ProxyTracing{},
			expected: 100.0,
		},
		{
			name: "rate",
			tracing: &egv1a1.ProxyTracing{
				SamplingRate: ptr.To[uint32](10),
			},
			expected: 10.0,
		},
		{
			name: "fraction numerator only",
			tracing: &egv1a1.ProxyTracing{
				SamplingFraction: &gwapiv1.Fraction{
					Numerator: 100,
				},
			},
			expected: 1.0,
		},
		{
			name: "fraction",
			tracing: &egv1a1.ProxyTracing{
				SamplingFraction: &gwapiv1.Fraction{
					Numerator:   1,
					Denominator: ptr.To[int32](10),
				},
			},
			expected: 0.1,
		},
		{
			name: "less than zero",
			tracing: &egv1a1.ProxyTracing{
				SamplingFraction: &gwapiv1.Fraction{
					Numerator:   1,
					Denominator: ptr.To[int32](-1),
				},
			},
			expected: 0,
		},
		{
			name: "greater than 100",
			tracing: &egv1a1.ProxyTracing{
				SamplingFraction: &gwapiv1.Fraction{
					Numerator:   101,
					Denominator: ptr.To[int32](1),
				},
			},
			expected: 100,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := proxySamplingRate(tc.tracing)
			if actual != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

func TestIsOverlappingHostname(t *testing.T) {
	tests := []struct {
		name      string
		hostname1 *gwapiv1.Hostname
		hostname2 *gwapiv1.Hostname
		want      bool
	}{
		{
			name:      "exact match",
			hostname1: ptr.To(gwapiv1.Hostname("example.com")),
			hostname2: ptr.To(gwapiv1.Hostname("example.com")),
			want:      true,
		},
		{
			name:      "two wildcards with same suffix",
			hostname1: ptr.To(gwapiv1.Hostname("*.example.com")),
			hostname2: ptr.To(gwapiv1.Hostname("*.example.com")),
			want:      true,
		},
		{
			name:      "two wildcards matches subdomain",
			hostname1: ptr.To(gwapiv1.Hostname("*.example.com")),
			hostname2: ptr.To(gwapiv1.Hostname("*.test.example.com")),
			want:      true,
		},
		{
			name:      "nil hostname matches all",
			hostname1: nil,
			hostname2: ptr.To(gwapiv1.Hostname("www.example.com")),
			want:      true,
		},
		{
			name:      "nil hostname matches subdomain",
			hostname1: nil,
			hostname2: ptr.To(gwapiv1.Hostname("*.example.com")),
			want:      true,
		},
		{
			name:      "two nil hostnames",
			hostname1: nil,
			hostname2: nil,
			want:      true,
		},
		{
			name:      "wildcard matches exact",
			hostname1: ptr.To(gwapiv1.Hostname("*.example.com")),
			hostname2: ptr.To(gwapiv1.Hostname("test.example.com")),
			want:      true,
		},
		{
			name:      "wildcard matches subdomain",
			hostname1: ptr.To(gwapiv1.Hostname("*.example.com")),
			hostname2: ptr.To(gwapiv1.Hostname("sub.test.example.com")),
			want:      true,
		},
		{
			name:      "wildcard matches empty subdomain",
			hostname1: ptr.To(gwapiv1.Hostname("*.example.com")),
			hostname2: ptr.To(gwapiv1.Hostname("example.com")),
			want:      true,
		},
		{
			name:      "different domains",
			hostname1: ptr.To(gwapiv1.Hostname("example.com")),
			hostname2: ptr.To(gwapiv1.Hostname("test.com")),
			want:      false,
		},
		{
			name:      "wildcard doesn't match different domain",
			hostname1: ptr.To(gwapiv1.Hostname("*.example.com")),
			hostname2: ptr.To(gwapiv1.Hostname("test.com")),
			want:      false,
		},
		{
			name:      "different wildcard domains",
			hostname1: ptr.To(gwapiv1.Hostname("*.example.com")),
			hostname2: ptr.To(gwapiv1.Hostname("*.test.com")),
			want:      false,
		},
		{
			name:      "different sub domains of same domain",
			hostname1: ptr.To(gwapiv1.Hostname("api.foo.dev")),
			hostname2: ptr.To(gwapiv1.Hostname("testing-api.foo.dev")),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOverlappingHostname(tt.hostname1, tt.hostname2); got != tt.want {
				t.Errorf("isOverlappingHostname(%q, %q) = %v, want %v", ptr.Deref(tt.hostname1, ""), ptr.Deref(tt.hostname2, ""), got, tt.want)
			}
			// Test should be symmetric
			if got := isOverlappingHostname(tt.hostname2, tt.hostname1); got != tt.want {
				t.Errorf("isOverlappingHostname(%q, %q) = %v, want %v", ptr.Deref(tt.hostname2, ""), ptr.Deref(tt.hostname1, ""), got, tt.want)
			}
		})
	}
}

func TestCheckOverlappingHostnames(t *testing.T) {
	tests := []struct {
		name     string
		gateway  *GatewayContext
		expected map[int]string // map of listener index to overlapping hostname
	}{
		{
			name: "no overlapping listeners",
			gateway: &GatewayContext{
				listeners: []*ListenerContext{
					{
						Listener: &gwapiv1.Listener{
							Name:     "listener-1",
							Protocol: gwapiv1.HTTPSProtocolType,
							Port:     443,
							Hostname: ptr.To(gwapiv1.Hostname("example.com")),
						},
					},
					{
						Listener: &gwapiv1.Listener{
							Name:     "listener-2",
							Protocol: gwapiv1.HTTPSProtocolType,
							Port:     443,
							Hostname: ptr.To(gwapiv1.Hostname("test.com")),
						},
					},
				},
			},
			expected: map[int]string{},
		},
		{
			name: "overlapping hostnames with same port",
			gateway: &GatewayContext{
				listeners: []*ListenerContext{
					{
						Listener: &gwapiv1.Listener{
							Name:     "listener-1",
							Protocol: gwapiv1.HTTPSProtocolType,
							Port:     443,
							Hostname: ptr.To(gwapiv1.Hostname("*.example.com")),
						},
					},
					{
						Listener: &gwapiv1.Listener{
							Name:     "listener-2",
							Protocol: gwapiv1.HTTPSProtocolType,
							Port:     443,
							Hostname: ptr.To(gwapiv1.Hostname("test.example.com")),
						},
					},
				},
			},
			expected: map[int]string{
				0: "test.example.com",
				1: "*.example.com",
			},
		},
		{
			name: "overlapping hostnames with different ports",
			gateway: &GatewayContext{
				listeners: []*ListenerContext{
					{
						Listener: &gwapiv1.Listener{
							Name:     "listener-1",
							Protocol: gwapiv1.HTTPSProtocolType,
							Port:     443,
							Hostname: ptr.To(gwapiv1.Hostname("*.example.com")),
						},
					},
					{
						Listener: &gwapiv1.Listener{
							Name:     "listener-2",
							Protocol: gwapiv1.HTTPSProtocolType,
							Port:     8443,
							Hostname: ptr.To(gwapiv1.Hostname("test.example.com")),
						},
					},
				},
			},
			expected: map[int]string{},
		},
		{
			name: "multiple overlapping listeners",
			gateway: &GatewayContext{
				listeners: []*ListenerContext{
					{
						Listener: &gwapiv1.Listener{
							Name:     "listener-1",
							Protocol: gwapiv1.HTTPSProtocolType,
							Port:     443,
							Hostname: ptr.To(gwapiv1.Hostname("*.example.com")),
						},
					},
					{
						Listener: &gwapiv1.Listener{
							Name:     "listener-2",
							Protocol: gwapiv1.HTTPSProtocolType,
							Port:     443,
							Hostname: ptr.To(gwapiv1.Hostname("test.example.com")),
						},
					},
					{
						Listener: &gwapiv1.Listener{
							Name:     "listener-3",
							Protocol: gwapiv1.HTTPSProtocolType,
							Port:     443,
							Hostname: ptr.To(gwapiv1.Hostname("sub.test.example.com")),
						},
					},
				},
			},
			expected: map[int]string{
				0: "test.example.com",
				1: "*.example.com",
				2: "*.example.com",
			},
		},
		{
			name: "nil hostnames",
			gateway: &GatewayContext{
				listeners: []*ListenerContext{
					{
						Listener: &gwapiv1.Listener{
							Name:     "listener-1",
							Protocol: gwapiv1.HTTPSProtocolType,
							Port:     443,
							Hostname: nil,
						},
					},
					{
						Listener: &gwapiv1.Listener{
							Name:     "listener-2",
							Protocol: gwapiv1.HTTPSProtocolType,
							Port:     443,
							Hostname: ptr.To(gwapiv1.Hostname("example.com")),
						},
					},
				},
			},
			expected: map[int]string{
				0: "example.com",
				1: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize Gateway and listener indices
			tt.gateway.Gateway = &gwapiv1.Gateway{
				Status: gwapiv1.GatewayStatus{
					Listeners: make([]gwapiv1.ListenerStatus, len(tt.gateway.listeners)),
				},
			}
			for i := range tt.gateway.listeners {
				tt.gateway.listeners[i].listenerStatusIdx = i
				tt.gateway.listeners[i].gateway = tt.gateway
				tt.gateway.Status.Listeners[i] = gwapiv1.ListenerStatus{
					Name:       tt.gateway.listeners[i].Name,
					Conditions: []metav1.Condition{},
				}
			}

			checkOverlappingHostnames(tt.gateway.listeners)

			// Verify the status conditions
			for idx, expectedHostname := range tt.expected {
				conditions := tt.gateway.Gateway.Status.Listeners[idx].Conditions
				var condition *metav1.Condition
				for i := range conditions {
					if conditions[i].Type == string(gwapiv1.ListenerConditionOverlappingTLSConfig) {
						condition = &conditions[i]
						break
					}
				}

				if expectedHostname != "" {
					if condition == nil {
						t.Errorf("expected condition for listener %d, got nil", idx)
						continue
					}
					if condition.Status != metav1.ConditionTrue {
						t.Errorf("expected condition status to be True for listener %d, got %v", idx, condition.Status)
					}
					if !strings.Contains(condition.Message, expectedHostname) {
						t.Errorf("expected condition message to contain %q for listener %d, got %q", expectedHostname, idx, condition.Message)
					}
				} else if condition == nil || condition.Status == metav1.ConditionFalse {
					// expectedHostname == "" means matching all hostnames
					t.Errorf("expected condition for listener %d, got nil or False", idx)
				}
			}

			if len(tt.expected) == 0 {
				if len(tt.gateway.Status.Listeners) != 0 {
					for idx, listener := range tt.gateway.Status.Listeners {
						if len(listener.Conditions) != 0 {
							t.Errorf("expected 0 conditions for listener %d, got %d", idx, len(listener.Conditions))
						}
					}
				}
			}
		})
	}
}

func TestCheckOverlappingCertificates(t *testing.T) {
	tests := []struct {
		name           string
		listeners      []*ListenerContext
		expectedStatus []expectedListenerStatus
	}{
		{
			name: "No overlapping certificates",
			listeners: []*ListenerContext{
				{
					Listener: &gwapiv1.Listener{
						Name:     "listener-1",
						Protocol: gwapiv1.HTTPSProtocolType,
						Port:     443,
					},
					listenerStatusIdx: 0,
					certDNSNames:      []string{"foo.example.com"},
				},
				{
					Listener: &gwapiv1.Listener{
						Name:     "listener-2",
						Protocol: gwapiv1.HTTPSProtocolType,
						Port:     443,
					},
					listenerStatusIdx: 1,
					certDNSNames:      []string{"bar.example.com"},
				},
			},
			expectedStatus: []expectedListenerStatus{},
		},
		{
			name: "Overlapping certificates with same port",
			listeners: []*ListenerContext{
				{
					Listener: &gwapiv1.Listener{
						Name:     "listener-1",
						Protocol: gwapiv1.HTTPSProtocolType,
						Port:     443,
					},
					listenerStatusIdx: 0,
					certDNSNames:      []string{"foo.example.com"},
				},
				{
					Listener: &gwapiv1.Listener{
						Name:     "listener-2",
						Protocol: gwapiv1.HTTPSProtocolType,
						Port:     443,
					},
					listenerStatusIdx: 1,
					certDNSNames:      []string{"foo.example.com"},
				},
			},
			expectedStatus: []expectedListenerStatus{
				{
					listenerName: "listener-1",
					condition:    gwapiv1.ListenerConditionOverlappingTLSConfig,
					status:       metav1.ConditionTrue,
					reason:       gwapiv1.ListenerReasonOverlappingCertificates,
					message:      "The certificate SAN foo.example.com overlaps with the certificate SAN foo.example.com in listener listener-2. ALPN will default to HTTP/1.1 to prevent HTTP/2 connection coalescing, unless explicitly configured via ClientTrafficPolicy",
				},
				{
					listenerName: "listener-2",
					condition:    gwapiv1.ListenerConditionOverlappingTLSConfig,
					status:       metav1.ConditionTrue,
					reason:       gwapiv1.ListenerReasonOverlappingCertificates,
					message:      "The certificate SAN foo.example.com overlaps with the certificate SAN foo.example.com in listener listener-1. ALPN will default to HTTP/1.1 to prevent HTTP/2 connection coalescing, unless explicitly configured via ClientTrafficPolicy",
				},
			},
		},
		{
			name: "Overlapping certificates with different ports",
			listeners: []*ListenerContext{
				{
					Listener: &gwapiv1.Listener{
						Name:     "listener-1",
						Protocol: gwapiv1.HTTPSProtocolType,
						Port:     443,
					},
					listenerStatusIdx: 0,
					certDNSNames:      []string{"foo.example.com"},
				},
				{
					Listener: &gwapiv1.Listener{
						Name:     "listener-2",
						Protocol: gwapiv1.HTTPSProtocolType,
						Port:     8443,
					},
					listenerStatusIdx: 1,
					certDNSNames:      []string{"foo.example.com"},
				},
			},
			expectedStatus: []expectedListenerStatus{},
		},
		{
			name: "Overlapping certificates with wildcard domain",
			listeners: []*ListenerContext{
				{
					Listener: &gwapiv1.Listener{
						Name:     "listener-1",
						Protocol: gwapiv1.HTTPSProtocolType,
						Port:     443,
					},
					listenerStatusIdx: 0,
					certDNSNames:      []string{"*.example.com"},
				},
				{
					Listener: &gwapiv1.Listener{
						Name:     "listener-2",
						Protocol: gwapiv1.HTTPSProtocolType,
						Port:     443,
					},
					listenerStatusIdx: 1,
					certDNSNames:      []string{"foo.example.com"},
				},
			},
			expectedStatus: []expectedListenerStatus{
				{
					listenerName: "listener-1",
					condition:    gwapiv1.ListenerConditionOverlappingTLSConfig,
					status:       metav1.ConditionTrue,
					reason:       gwapiv1.ListenerReasonOverlappingCertificates,
					message:      "The certificate SAN *.example.com overlaps with the certificate SAN foo.example.com in listener listener-2. ALPN will default to HTTP/1.1 to prevent HTTP/2 connection coalescing, unless explicitly configured via ClientTrafficPolicy",
				},
				{
					listenerName: "listener-2",
					condition:    gwapiv1.ListenerConditionOverlappingTLSConfig,
					status:       metav1.ConditionTrue,
					reason:       gwapiv1.ListenerReasonOverlappingCertificates,
					message:      "The certificate SAN foo.example.com overlaps with the certificate SAN *.example.com in listener listener-1. ALPN will default to HTTP/1.1 to prevent HTTP/2 connection coalescing, unless explicitly configured via ClientTrafficPolicy",
				},
			},
		},
		{
			name: "Overlapping certificates with multiple dns names",
			listeners: []*ListenerContext{
				{
					Listener: &gwapiv1.Listener{
						Name:     "listener-1",
						Protocol: gwapiv1.HTTPSProtocolType,
						Port:     443,
					},
					listenerStatusIdx: 0,
					certDNSNames:      []string{"foo.example.com", "bar.example.org"},
				},
				{
					Listener: &gwapiv1.Listener{
						Name:     "listener-2",
						Protocol: gwapiv1.HTTPSProtocolType,
						Port:     443,
					},
					listenerStatusIdx: 1,
					certDNSNames:      []string{"bar.example.com", "*.example.org", "bar.example.com"},
				},
			},
			expectedStatus: []expectedListenerStatus{
				{
					listenerName: "listener-1",
					condition:    gwapiv1.ListenerConditionOverlappingTLSConfig,
					status:       metav1.ConditionTrue,
					reason:       gwapiv1.ListenerReasonOverlappingCertificates,
					message:      "The certificate SAN bar.example.org overlaps with the certificate SAN *.example.org in listener listener-2. ALPN will default to HTTP/1.1 to prevent HTTP/2 connection coalescing, unless explicitly configured via ClientTrafficPolicy",
				},
				{
					listenerName: "listener-2",
					condition:    gwapiv1.ListenerConditionOverlappingTLSConfig,
					status:       metav1.ConditionTrue,
					reason:       gwapiv1.ListenerReasonOverlappingCertificates,
					message:      "The certificate SAN *.example.org overlaps with the certificate SAN bar.example.org in listener listener-1. ALPN will default to HTTP/1.1 to prevent HTTP/2 connection coalescing, unless explicitly configured via ClientTrafficPolicy",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock GatewayContext
			gateway := &GatewayContext{
				Gateway: &gwapiv1.Gateway{
					Status: gwapiv1.GatewayStatus{
						Listeners: make([]gwapiv1.ListenerStatus, len(tt.listeners)),
					},
				},
				listeners: tt.listeners,
			}

			// Initialize listener
			for i := range gateway.Status.Listeners {
				gateway.Status.Listeners[i] = gwapiv1.ListenerStatus{
					Name:       tt.listeners[i].Name,
					Conditions: []metav1.Condition{},
				}
				gateway.listeners[i].listenerStatusIdx = i
				gateway.listeners[i].gateway = gateway
			}

			// Process overlapping certificates
			checkOverlappingCertificates(tt.listeners)

			// Verify the status conditions
			for _, expected := range tt.expectedStatus {
				found := false
				for _, listener := range gateway.listeners {
					if string(listener.Name) != expected.listenerName {
						continue
					}

					conditions := status.GetGatewayListenerStatusConditions(gateway.Gateway, listener.listenerStatusIdx)
					for _, condition := range conditions {
						if condition.Type == string(expected.condition) &&
							condition.Status == expected.status &&
							condition.Reason == string(expected.reason) &&
							condition.Message == expected.message {
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					t.Errorf("Expected status condition not found for listener %s: %+v", expected.listenerName, expected)
				}
			}

			// Verify no unexpected status conditions
			for _, listener := range gateway.listeners {
				conditions := status.GetGatewayListenerStatusConditions(gateway.Gateway, listener.listenerStatusIdx)
				for _, condition := range conditions {
					if condition.Type == string(gwapiv1.ListenerConditionOverlappingTLSConfig) {
						found := false
						for _, expected := range tt.expectedStatus {
							if string(listener.Name) == expected.listenerName &&
								condition.Status == expected.status &&
								condition.Reason == string(expected.reason) &&
								condition.Message == expected.message {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("Unexpected status condition found for listener %s: %+v", listener.Name, condition)
						}
					}
				}
			}
		})
	}
}

func TestProcessTracingServiceName(t *testing.T) {
	cases := []struct {
		name                string
		gateway             *gwapiv1.Gateway
		envoyProxy          *egv1a1.EnvoyProxy
		mergeGateways       bool
		expectedServiceName string
		expectError         bool
	}{
		{
			name: "no tracing configuration",
			gateway: &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gateway",
					Namespace: "test-namespace",
				},
			},
			envoyProxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxy",
					Namespace: "test-namespace",
				},
			},
			expectedServiceName: "",
		},
		{
			name: "tracing with default service name",
			gateway: &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gateway",
					Namespace: "test-namespace",
				},
			},
			envoyProxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxy",
					Namespace: "test-namespace",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Tracing: &egv1a1.ProxyTracing{
							Provider: egv1a1.TracingProvider{
								Type: egv1a1.TracingProviderTypeOpenTelemetry,
								BackendCluster: egv1a1.BackendCluster{
									BackendRefs: []egv1a1.BackendRef{
										{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Name:      "otel-collector",
												Port:      ptr.To(gwapiv1.PortNumber(4317)),
												Namespace: ptr.To(gwapiv1.Namespace("monitoring")),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedServiceName: "test-gateway.test-namespace",
		},
		{
			name: "tracing with custom service name",
			gateway: &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gateway",
					Namespace: "test-namespace",
				},
			},
			envoyProxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxy",
					Namespace: "test-namespace",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Tracing: &egv1a1.ProxyTracing{
							Provider: egv1a1.TracingProvider{
								Type: egv1a1.TracingProviderTypeOpenTelemetry,
								BackendCluster: egv1a1.BackendCluster{
									BackendRefs: []egv1a1.BackendRef{
										{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Name:      "otel-collector",
												Port:      ptr.To(gwapiv1.PortNumber(4317)),
												Namespace: ptr.To(gwapiv1.Namespace("monitoring")),
											},
										},
									},
								},
								ServiceName: ptr.To("my-custom-service"),
							},
						},
					},
				},
			},
			expectedServiceName: "my-custom-service",
		},
		{
			name: "tracing with merge gateways and custom service name",
			gateway: &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gateway",
					Namespace: "test-namespace",
				},
				Spec: gwapiv1.GatewaySpec{
					GatewayClassName: "test-gateway-class",
				},
			},
			envoyProxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxy",
					Namespace: "test-namespace",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Tracing: &egv1a1.ProxyTracing{
							Provider: egv1a1.TracingProvider{
								Type: egv1a1.TracingProviderTypeOpenTelemetry,
								BackendCluster: egv1a1.BackendCluster{
									BackendRefs: []egv1a1.BackendRef{
										{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Name:      "otel-collector",
												Port:      ptr.To(gwapiv1.PortNumber(4317)),
												Namespace: ptr.To(gwapiv1.Namespace("monitoring")),
											},
										},
									},
								},
								ServiceName: ptr.To("custom-service"),
							},
						},
					},
				},
			},
			mergeGateways:       true,
			expectedServiceName: "custom-service", // Custom service name should override merge logic
		},
		{
			name: "tracing with merge gateways without custom service name",
			gateway: &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gateway",
					Namespace: "test-namespace",
				},
				Spec: gwapiv1.GatewaySpec{
					GatewayClassName: "test-gateway-class",
				},
			},
			envoyProxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxy",
					Namespace: "test-namespace",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Tracing: &egv1a1.ProxyTracing{
							Provider: egv1a1.TracingProvider{
								Type: egv1a1.TracingProviderTypeOpenTelemetry,
								BackendCluster: egv1a1.BackendCluster{
									BackendRefs: []egv1a1.BackendRef{
										{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Name:      "otel-collector",
												Port:      ptr.To(gwapiv1.PortNumber(4317)),
												Namespace: ptr.To(gwapiv1.Namespace("monitoring")),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			mergeGateways:       true,
			expectedServiceName: "test-gateway-class", // Should use gateway class name when merging
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			translator := &Translator{}
			resources := &resource.Resources{}

			// Mock service to resolve BackendRefs
			resources.Services = append(resources.Services,
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "monitoring",
						Name:      "otel-collector",
					},
					Spec: corev1.ServiceSpec{
						ClusterIP: "3.3.3.3",
						Ports: []corev1.ServicePort{
							{
								Name:        "grpc",
								Port:        4317,
								TargetPort:  intstr.IntOrString{IntVal: 4317},
								Protocol:    corev1.ProtocolTCP,
								AppProtocol: ptr.To("grpc"),
							},
						},
					},
				},
			)

			// Mock endpointSlice to resolve Service
			resources.EndpointSlices = append(resources.EndpointSlices,
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "monitoring",
						Name:      "otel-collector",
						Labels: map[string]string{
							"kubernetes.io/service-name": "otel-collector",
						},
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
						},
					},
					Ports: []discoveryv1.EndpointPort{
						{
							Name:        ptr.To("grpc"),
							Protocol:    ptr.To(corev1.ProtocolTCP),
							Port:        ptr.To(int32(4317)),
							AppProtocol: ptr.To("grpc"),
						},
					},
				},
			)

			result, err := translator.processTracing(tc.gateway, tc.envoyProxy, tc.mergeGateways, resources)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tc.expectedServiceName == "" {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, tc.expectedServiceName, result.ServiceName)
		})
	}
}
