// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
)

func extensionBackendResources(t *testing.T) *resource.Resources {
	t.Helper()
	data, err := os.ReadFile("testdata/envoyextensionpolicy-with-dynamicmodule.in.yaml")
	require.NoError(t, err)
	resources := &resource.Resources{}
	mustUnmarshal(t, data, resources)
	base, err := os.ReadFile("testdata/base/base.yaml")
	require.NoError(t, err)
	baseResources := &resource.Resources{}
	mustUnmarshal(t, base, baseResources)
	resources.Secrets = append(resources.Secrets, baseResources.Secrets...)
	for _, namespace := range []string{"default", "envoy-gateway"} {
		resources.Namespaces = append(resources.Namespaces, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}})
		for _, name := range []string{"service-1", "service-2"} {
			resources.Services = append(resources.Services, &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
				Spec:       corev1.ServiceSpec{ClusterIP: "10.0.0.1", Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
			})
		}
	}
	for index, policy := range resources.EnvoyExtensionPolicies {
		policy.CreationTimestamp = metav1.NewTime(time.Unix(int64(index+1), 0))
		policy.Spec.Backends = []egv1a1.ExtensionBackend{{
			Name: "resolver", BackendRef: gwapiv1.BackendObjectReference{Name: "service-2", Port: new(gwapiv1.PortNumber(8080))},
		}}
	}
	return resources
}

func translateExtensionBackends(t *testing.T, resources *resource.Resources) (*TranslateResult, map[string]*ir.HTTPRoute) {
	t.Helper()
	return translateExtensionBackendsWithFeature(t, resources, true)
}

func translateExtensionBackendsWithFeature(t *testing.T, resources *resource.Resources, enabled bool) (*TranslateResult, map[string]*ir.HTTPRoute) {
	t.Helper()
	resources.Sort()
	translator := &Translator{
		GatewayControllerName: egv1a1.GatewayControllerName, GatewayClassName: "envoy-gateway-class",
		ControllerNamespace: "envoy-gateway-system", EndpointRoutingDisabled: true, BackendEnabled: true,
		ExtensionBackendsEnabled: enabled,
		MergeGateways:            IsMergeGatewaysEnabled(resources), Logger: logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo),
	}
	result, _ := translator.Translate(t.Context(), resources)
	routes := make(map[string]*ir.HTTPRoute)
	for _, deployment := range result.XdsIR {
		for _, listener := range deployment.HTTP {
			for _, route := range listener.Routes {
				routes[route.Metadata.Name] = route
			}
		}
	}
	require.Len(t, routes, 2)
	return result, routes
}

// Resolve references through the registry, just as serialized or copied IR consumers do.
func resolverBackendForRoute(t *testing.T, result *TranslateResult, route *ir.HTTPRoute) *ir.ExtensionBackend {
	t.Helper()
	name := route.EnvoyExtensions.Backends["resolver"]
	require.NotEmpty(t, name)
	for _, x := range result.XdsIR {
		for _, backend := range x.ExtensionBackends {
			if backend.Name == name {
				return backend
			}
		}
	}
	t.Fatalf("backend %s not found", name)
	return nil
}

func extensionPolicyCondition(t *testing.T, result *TranslateResult, policyName string) *metav1.Condition {
	t.Helper()
	for _, policy := range result.EnvoyExtensionPolicies {
		if policy.Name == policyName {
			require.NotEmpty(t, policy.Status.Ancestors)
			condition := meta.FindStatusCondition(policy.Status.Ancestors[0].Conditions, string(gwapiv1.PolicyConditionAccepted))
			require.NotNil(t, condition)
			return condition
		}
	}
	t.Fatalf("policy %s not found", policyName)
	return nil
}

func TestExtensionBackendsFeatureGate(t *testing.T) {
	for _, scope := range []string{"route", "gateway", "inherited", "no bindings"} {
		for _, enabled := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/enabled=%t", scope, enabled), func(t *testing.T) {
				resources := extensionBackendResources(t)
				parent, child := resources.EnvoyExtensionPolicies[0], resources.EnvoyExtensionPolicies[1]
				affected := make(map[string]bool)
				switch scope {
				case "route":
					parent.Spec.Backends = nil
					affected["httproute-1"] = true
				case "gateway":
					child.Spec.Backends = nil
					affected["httproute-2"] = true
				case "inherited":
					child.Spec.Backends = nil
					child.Spec.MergeType = new(egv1a1.JSONMerge)
					affected["httproute-1"] = true
					affected["httproute-2"] = true
				case "no bindings":
					parent.Spec.Backends = nil
					child.Spec.Backends = nil
				}

				// These bindings reference Services, and the Backend API is enabled
				// by the helper. Neither bypasses the independent extension gate.
				result, routes := translateExtensionBackendsWithFeature(t, resources, enabled)
				for name, route := range routes {
					if affected[name] && !enabled {
						require.NotNil(t, route.DirectResponse)
						require.EqualValues(t, 500, *route.DirectResponse.StatusCode)
						require.Empty(t, route.EnvoyExtensions.Backends)
						require.Empty(t, route.EnvoyExtensions.DynamicModules, "a parent must not restore extensions")
					} else {
						require.Nil(t, route.DirectResponse)
						require.NotEmpty(t, route.EnvoyExtensions.DynamicModules)
						if affected[name] {
							require.Contains(t, route.EnvoyExtensions.Backends, "resolver")
						} else {
							require.Empty(t, route.EnvoyExtensions.Backends)
						}
					}
				}
				for _, policy := range resources.EnvoyExtensionPolicies {
					condition := extensionPolicyCondition(t, result, policy.Name)
					hasBindings := len(policy.Spec.Backends) > 0 || scope == "inherited"
					if hasBindings && !enabled {
						require.Equal(t, metav1.ConditionFalse, condition.Status)
						require.Equal(t, string(gwapiv1.PolicyReasonInvalid), condition.Reason)
						require.Contains(t, condition.Message, "extensionApis.enableExtensionBackends")
					} else {
						require.Equal(t, metav1.ConditionTrue, condition.Status)
					}
				}
				for _, x := range result.XdsIR {
					if !enabled || len(affected) == 0 {
						require.Empty(t, x.ExtensionBackends, "disabled bindings must not register clusters")
					} else {
						require.Len(t, x.ExtensionBackends, 1)
					}
				}
			})
		}
	}
}

func TestExtensionBackendIsolation(t *testing.T) {
	for _, mergeGateways := range []bool{false, true} {
		t.Run(fmt.Sprint("mergeGateways=", mergeGateways), func(t *testing.T) {
			resources := extensionBackendResources(t)
			if mergeGateways {
				second := resources.Gateways[0].DeepCopy()
				second.Name = "gateway-2"
				second.Spec.Listeners[0].Port = 81
				resources.Gateways = append(resources.Gateways, second)
				resources.HTTPRoutes[0].Spec.ParentRefs[0].Name = "gateway-2"
				resources.EnvoyProxyForGatewayClass = resources.EnvoyProxiesForGateways[0].DeepCopy()
				resources.EnvoyProxyForGatewayClass.Spec.MergeGateways = new(true)
				for _, gateway := range resources.Gateways {
					gateway.Spec.Infrastructure = nil
				}
			}
			result, routes := translateExtensionBackends(t, resources)
			require.Len(t, result.XdsIR, 1)
			for _, policy := range result.EnvoyExtensionPolicies {
				require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, policy.Name).Status)
			}
			first := resolverBackendForRoute(t, result, routes["httproute-1"])
			second := resolverBackendForRoute(t, result, routes["httproute-2"])
			require.NotEqual(t, first.Name, second.Name)
			require.Equal(t, "default", first.Settings[0].Metadata.Namespace)
			require.Equal(t, "envoy-gateway", second.Settings[0].Metadata.Namespace)
			for _, route := range routes {
				require.Nil(t, route.DirectResponse)
			}
		})
	}
}

func TestExtensionBackendRegistry(t *testing.T) {
	translator := &Translator{
		GatewayControllerName: egv1a1.GatewayControllerName, GatewayClassName: "envoy-gateway-class",
		ControllerNamespace: "envoy-gateway-system", EndpointRoutingDisabled: true, BackendEnabled: true,
		ExtensionBackendsEnabled: true,
		Logger:                   logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo),
	}
	// Run the same translator with maxConnections 17, then 18. The second output
	// must use the new setting, showing the lookup is rebuilt for each translation.
	for _, limit := range []int64{17, 18} {
		resources := extensionBackendResources(t)
		parent, child := resources.EnvoyExtensionPolicies[0], resources.EnvoyExtensionPolicies[1]
		parent.Spec.Backends[0].BackendSettings = &egv1a1.ClusterSettings{
			CircuitBreaker: &egv1a1.CircuitBreaker{MaxConnections: new(limit)},
		}
		child.Spec.MergeType = new(egv1a1.JSONMerge)
		child.Spec.Backends = nil
		resources.Sort()
		result, err := translator.Translate(t.Context(), resources)
		require.NoError(t, err)
		require.Len(t, result.XdsIR, 1)
		for _, x := range result.XdsIR {
			require.Len(t, x.ExtensionBackends, 1, "inherited and direct uses of one binding share a definition")
			require.EqualValues(t, limit, *x.ExtensionBackends[0].Traffic.CircuitBreaker.MaxConnections)
			data, err := json.Marshal(x)
			require.NoError(t, err)
			var decoded ir.Xds
			require.NoError(t, json.Unmarshal(data, &decoded))
			for _, snapshot := range []*ir.Xds{x, x.DeepCopy(), &decoded} {
				require.Len(t, snapshot.ExtensionBackends, 1)
				require.NoError(t, snapshot.Validate())
				for _, listener := range snapshot.HTTP {
					for _, route := range listener.Routes {
						require.Equal(t, map[string]string{"resolver": snapshot.ExtensionBackends[0].Name}, route.EnvoyExtensions.Backends)
					}
				}
			}
		}
	}
}

func TestExtensionBackendRegistryOnlyIncludesAttachedPolicies(t *testing.T) {
	for _, failure := range []string{"overridden", "invalid backend", "invalid filter"} {
		t.Run(failure, func(t *testing.T) {
			resources := extensionBackendResources(t)
			parent, child := resources.EnvoyExtensionPolicies[0], resources.EnvoyExtensionPolicies[1]
			child.Spec.TargetRef = nil
			child.Spec.TargetRefs = nil
			for _, route := range resources.HTTPRoutes {
				child.Spec.TargetRefs = append(child.Spec.TargetRefs, gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{Group: gwapiv1.GroupName, Kind: "HTTPRoute", Name: gwapiv1.ObjectName(route.Name)},
				})
			}
			switch failure {
			case "invalid backend":
				// Fail after one backend has already translated successfully.
				child.Spec.Backends = append(child.Spec.Backends, egv1a1.ExtensionBackend{
					Name: "missing", BackendRef: gwapiv1.BackendObjectReference{Name: "missing", Port: new(gwapiv1.PortNumber(8080))},
				})
			case "invalid filter":
				child.Spec.DynamicModule[0].Name = "missing"
			}
			result, routes := translateExtensionBackends(t, resources)
			for _, x := range result.XdsIR {
				switch failure {
				case "overridden":
					require.Len(t, x.ExtensionBackends, 1)
					require.Contains(t, x.ExtensionBackends[0].Name, child.Name)
					require.NotContains(t, x.ExtensionBackends[0].Name, parent.Name)
				case "invalid backend", "invalid filter":
					require.Empty(t, x.ExtensionBackends)
				}
				if failure != "overridden" {
					for _, route := range routes {
						require.EqualValues(t, 500, *route.DirectResponse.StatusCode)
						require.Equal(t, &ir.EnvoyExtensionFeatures{}, route.EnvoyExtensions, "a failed child must block parent extensions")
					}
				}
			}
		})
	}
}

func TestExtensionBackendResources(t *testing.T) {
	for _, tc := range []struct {
		name        string
		spec        egv1a1.BackendSpec
		addressType ir.DestinationAddressType
	}{
		{"ip", egv1a1.BackendSpec{Endpoints: []egv1a1.BackendEndpoint{{IP: &egv1a1.IPEndpoint{Address: "10.0.0.5", Port: 8080}}}}, ir.IP},
		{"dns", egv1a1.BackendSpec{Endpoints: []egv1a1.BackendEndpoint{{FQDN: &egv1a1.FQDNEndpoint{Hostname: "resolver.example.com", Port: 8080}}}}, ir.FQDN},
		{"unix", egv1a1.BackendSpec{Endpoints: []egv1a1.BackendEndpoint{{Unix: &egv1a1.UnixSocket{Path: "/run/resolver.sock"}}}}, ir.UDS},
		{"dynamic resolver", egv1a1.BackendSpec{Type: new(egv1a1.BackendTypeDynamicResolver)}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resources := extensionBackendResources(t)
			child := resources.EnvoyExtensionPolicies[1]
			child.Spec.Backends[0].BackendRef = gwapiv1.BackendObjectReference{
				Group: new(gwapiv1.Group(egv1a1.GroupName)), Kind: new(gwapiv1.Kind(egv1a1.KindBackend)), Name: "resolver",
			}
			resources.Backends = []*egv1a1.Backend{{
				ObjectMeta: metav1.ObjectMeta{Namespace: child.Namespace, Name: "resolver"}, Spec: tc.spec,
			}}
			result, routes := translateExtensionBackends(t, resources)
			if tc.addressType == "" {
				require.Contains(t, extensionPolicyCondition(t, result, child.Name).Message, "dynamic resolver backend default/resolver is not supported")
				require.EqualValues(t, 500, *routes["httproute-1"].DirectResponse.StatusCode)
				return
			}
			require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
			backend := resolverBackendForRoute(t, result, routes["httproute-1"])
			require.Equal(t, tc.addressType, *backend.Settings[0].AddressType)
			require.Len(t, backend.Settings[0].Endpoints, 1)
			if tc.addressType == ir.UDS {
				require.Equal(t, new(tc.spec.Endpoints[0].Unix.Path), backend.Settings[0].Endpoints[0].Path)
			}
		})
	}
}

func TestExtensionBackendNames(t *testing.T) {
	resources := extensionBackendResources(t)
	child := resources.EnvoyExtensionPolicies[1]
	// A built in cluster name remains a valid alias.
	child.Spec.Backends[0].Name = "tracing"
	child.Spec.Backends = append(child.Spec.Backends, egv1a1.ExtensionBackend{
		Name: "auth", BackendRef: child.Spec.Backends[0].BackendRef,
	})
	result, routes := translateExtensionBackends(t, resources)
	require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
	before := routes["httproute-1"].EnvoyExtensions.Backends
	for alias, destination := range before {
		require.NotEqual(t, alias, destination)
	}

	slices.Reverse(child.Spec.Backends)
	result, routes = translateExtensionBackends(t, resources)
	require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
	for alias, destination := range routes["httproute-1"].EnvoyExtensions.Backends {
		require.Equal(t, before[alias], destination)
	}
}

func TestExtensionBackendInvalidAndRecovery(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mutate  func(*egv1a1.EnvoyExtensionPolicy)
		message string
	}{
		{"missing", func(p *egv1a1.EnvoyExtensionPolicy) { p.Spec.Backends[0].BackendRef.Name = "missing" }, "not found"},
		{"duplicate", func(p *egv1a1.EnvoyExtensionPolicy) { p.Spec.Backends = append(p.Spec.Backends, p.Spec.Backends[0]) }, "duplicated"},
		{"invalid alias", func(p *egv1a1.EnvoyExtensionPolicy) { p.Spec.Backends[0].Name = "bad/name" }, "invalid"},
		{"missing port", func(p *egv1a1.EnvoyExtensionPolicy) { p.Spec.Backends[0].BackendRef.Port = nil }, "port"},
		{"unsupported kind", func(p *egv1a1.EnvoyExtensionPolicy) {
			p.Spec.Backends[0].BackendRef.Kind = new(gwapiv1.Kind("ConfigMap"))
		}, "kind"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resources := extensionBackendResources(t)
			parent, child := resources.EnvoyExtensionPolicies[0], resources.EnvoyExtensionPolicies[1]
			valid := child.Spec.Backends
			child.Spec.Backends = slices.Clone(valid)
			tc.mutate(child)
			result, routes := translateExtensionBackends(t, resources)
			condition := extensionPolicyCondition(t, result, child.Name)
			require.Equal(t, metav1.ConditionFalse, condition.Status)
			require.Equal(t, string(gwapiv1.PolicyReasonInvalid), condition.Reason)
			require.Contains(t, condition.Message, tc.message)
			require.EqualValues(t, 500, *routes["httproute-1"].DirectResponse.StatusCode)
			// A failed child must keep the parent from installing its filters or bindings.
			require.Empty(t, routes["httproute-1"].EnvoyExtensions.Backends)
			require.Empty(t, routes["httproute-1"].EnvoyExtensions.DynamicModules)
			require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, parent.Name).Status)
			require.Nil(t, routes["httproute-2"].DirectResponse)

			child.Spec.Backends = valid
			result, routes = translateExtensionBackends(t, resources)
			require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
			require.Nil(t, routes["httproute-1"].DirectResponse)
			require.Contains(t, routes["httproute-1"].EnvoyExtensions.Backends, "resolver")
		})
	}
}

func TestExtensionBackendInheritanceAndAuthorization(t *testing.T) {
	for _, mergeType := range []egv1a1.MergeType{egv1a1.JSONMerge, egv1a1.StrategicMerge} {
		t.Run(string(mergeType), func(t *testing.T) {
			resources := extensionBackendResources(t)
			parent, child := resources.EnvoyExtensionPolicies[0], resources.EnvoyExtensionPolicies[1]
			child.Spec.MergeType = new(mergeType)
			child.Spec.Backends = nil
			result, routes := translateExtensionBackends(t, resources)
			require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
			backend := resolverBackendForRoute(t, result, routes["httproute-1"])
			require.Equal(t, "envoy-gateway", backend.Settings[0].Metadata.Namespace)
			require.Contains(t, backend.Name, parent.Name)

			parent.Spec.Backends[0].BackendRef.Namespace = new(gwapiv1.Namespace("default"))
			result, routes = translateExtensionBackends(t, resources)
			require.Contains(t, extensionPolicyCondition(t, result, child.Name).Message, "not permitted by any ReferenceGrant")
			require.EqualValues(t, 500, *routes["httproute-1"].DirectResponse.StatusCode)

			resources.ReferenceGrants = []*gwapiv1b1.ReferenceGrant{{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "allow-extension"},
				Spec: gwapiv1b1.ReferenceGrantSpec{
					From: []gwapiv1b1.ReferenceGrantFrom{{Group: egv1a1.GroupName, Kind: egv1a1.KindEnvoyExtensionPolicy, Namespace: "envoy-gateway"}},
					To:   []gwapiv1b1.ReferenceGrantTo{{Group: "", Kind: "Service"}},
				},
			}}
			resources.BackendTLSPolicies = []*gwapiv1.BackendTLSPolicy{{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "resolver-tls"},
				Spec: gwapiv1.BackendTLSPolicySpec{
					TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{{LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{Group: "", Kind: "Service", Name: "service-2"}}},
					Validation: gwapiv1.BackendTLSPolicyValidation{Hostname: "resolver.example.com", WellKnownCACertificates: new(gwapiv1.WellKnownCACertificatesType("System"))},
				},
			}}
			result, routes = translateExtensionBackends(t, resources)
			require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
			backend = resolverBackendForRoute(t, result, routes["httproute-1"])
			require.Equal(t, "default", backend.Settings[0].Metadata.Namespace)
			require.True(t, backend.Settings[0].TLS.UseSystemTrustStore)
			require.Equal(t, "resolver.example.com", *backend.Settings[0].TLS.SNI)

			// A child replaces the list as a whole and resolves its own references.
			parent.Spec.Backends = append(parent.Spec.Backends, egv1a1.ExtensionBackend{Name: "parent-only", BackendRef: parent.Spec.Backends[0].BackendRef})
			child.Spec.Backends = []egv1a1.ExtensionBackend{{Name: "resolver", BackendRef: gwapiv1.BackendObjectReference{Name: "service-1", Port: new(gwapiv1.PortNumber(8080))}}}
			result, routes = translateExtensionBackends(t, resources)
			require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
			require.Len(t, routes["httproute-1"].EnvoyExtensions.Backends, 1)
			require.Contains(t, routes["httproute-1"].EnvoyExtensions.Backends["resolver"], child.Name)
			require.Equal(t, "service-1", resolverBackendForRoute(t, result, routes["httproute-1"]).Settings[0].Metadata.Name)
		})
	}
}

func TestExtensionBackendMergedGatewayTLS(t *testing.T) {
	resources := extensionBackendResources(t)
	data, err := os.ReadFile("testdata/backend-tls-settings.in.yaml")
	require.NoError(t, err)
	tlsResources := &resource.Resources{}
	mustUnmarshal(t, data, tlsResources)
	second := resources.Gateways[0].DeepCopy()
	second.Name = "gateway-2"
	second.Spec.Listeners[0].Port = 81
	resources.Gateways = append(resources.Gateways, second)
	resources.HTTPRoutes[1].Spec.ParentRefs[0].Name = gwapiv1.ObjectName(second.Name)
	resources.EnvoyProxyForGatewayClass = resources.EnvoyProxiesForGateways[0].DeepCopy()
	resources.EnvoyProxyForGatewayClass.Spec.MergeGateways = new(true)
	for i, gateway := range resources.Gateways {
		gateway.Spec.Infrastructure = nil
		gateway.Spec.TLS = tlsResources.Gateways[1].Spec.TLS.DeepCopy()
		secret := tlsResources.Secrets[1].DeepCopy()
		secret.Name = fmt.Sprintf("client-auth-%d", i)
		secret.Namespace = gateway.Namespace
		resources.Secrets = append(resources.Secrets, secret)
		gateway.Spec.TLS.Backend.ClientCertificateRef.Name = gwapiv1.ObjectName(secret.Name)
	}
	policy := resources.EnvoyExtensionPolicies[1]
	policy.Spec.TargetRef = nil
	policy.Spec.TargetRefs = []gwapiv1.LocalPolicyTargetReferenceWithSectionName{}
	for _, route := range resources.HTTPRoutes {
		policy.Spec.TargetRefs = append(policy.Spec.TargetRefs, gwapiv1.LocalPolicyTargetReferenceWithSectionName{
			LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{Group: gwapiv1.GroupName, Kind: "HTTPRoute", Name: gwapiv1.ObjectName(route.Name)},
		})
	}
	resources.EnvoyExtensionPolicies = []*egv1a1.EnvoyExtensionPolicy{policy}
	resources.BackendTLSPolicies = []*gwapiv1.BackendTLSPolicy{{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "resolver-tls"},
		Spec: gwapiv1.BackendTLSPolicySpec{
			TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{{LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{Group: "", Kind: "Service", Name: "service-2"}}},
			Validation: gwapiv1.BackendTLSPolicyValidation{Hostname: "resolver.example.com", WellKnownCACertificates: new(gwapiv1.WellKnownCACertificatesType("System"))},
		},
	}}
	result, routes := translateExtensionBackends(t, resources)
	require.Len(t, result.XdsIR, 1)
	require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, policy.Name).Status)
	first := resolverBackendForRoute(t, result, routes["httproute-1"])
	secondBackend := resolverBackendForRoute(t, result, routes["httproute-2"])
	require.NotEqual(t, first.Settings[0].TLS.ClientCertificates, secondBackend.Settings[0].TLS.ClientCertificates)
	require.NotEqual(t, first.Name, secondBackend.Name, "different Gateway TLS configurations must not share a cluster")
}

func TestExtensionBackendSettings(t *testing.T) {
	resources := extensionBackendResources(t)
	parent, child := resources.EnvoyExtensionPolicies[0], resources.EnvoyExtensionPolicies[1]
	parent.Spec.Backends[0].BackendSettings = &egv1a1.ClusterSettings{
		LoadBalancer:   &egv1a1.LoadBalancer{Type: egv1a1.RoundRobinLoadBalancerType},
		Timeout:        &egv1a1.Timeout{TCP: &egv1a1.TCPTimeout{ConnectTimeout: new(gwapiv1.Duration("3s"))}},
		CircuitBreaker: &egv1a1.CircuitBreaker{MaxConnections: new(int64(17))},
	}
	child.Spec.MergeType = new(egv1a1.JSONMerge)
	child.Spec.Backends = nil
	result, routes := translateExtensionBackends(t, resources)
	require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
	backend := resolverBackendForRoute(t, result, routes["httproute-1"])
	require.NotNil(t, backend.Traffic.LoadBalancer.RoundRobin)
	require.Equal(t, 3*time.Second, backend.Traffic.Timeout.TCP.ConnectTimeout.Duration)
	require.EqualValues(t, 17, *backend.Traffic.CircuitBreaker.MaxConnections)
	require.Equal(t, backend.Traffic, resolverBackendForRoute(t, result, routes["httproute-2"]).Traffic)
	child.Spec.Backends = []egv1a1.ExtensionBackend{{Name: "resolver", BackendRef: parent.Spec.Backends[0].BackendRef}}
	result, routes = translateExtensionBackends(t, resources)
	require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
	require.Nil(t, resolverBackendForRoute(t, result, routes["httproute-1"]).Traffic)
	require.NotNil(t, resolverBackendForRoute(t, result, routes["httproute-2"]).Traffic)
}

func TestExtensionBackendInvalidSettings(t *testing.T) {
	for _, tc := range []struct {
		name     string
		settings egv1a1.ClusterSettings
		message  string
	}{
		{"connection timeout", egv1a1.ClusterSettings{Timeout: &egv1a1.Timeout{TCP: &egv1a1.TCPTimeout{ConnectTimeout: new(gwapiv1.Duration("invalid"))}}}, "ConnectTimeout"},
		{"request timeout", egv1a1.ClusterSettings{Timeout: &egv1a1.Timeout{HTTP: &egv1a1.HTTPTimeout{RequestTimeout: new(gwapiv1.Duration("1s"))}}}, "controlled by the extension"},
		{"stream idle timeout", egv1a1.ClusterSettings{Timeout: &egv1a1.Timeout{HTTP: &egv1a1.HTTPTimeout{StreamIdleTimeout: new(gwapiv1.Duration("1s"))}}}, "controlled by the extension"},
		{"circuit breaker", egv1a1.ClusterSettings{CircuitBreaker: &egv1a1.CircuitBreaker{MaxConnections: new(int64(-1))}}, "MaxConnections"},
		{"http3 default", egv1a1.ClusterSettings{HTTP3: &egv1a1.BackendHTTP3Settings{}}, "http3 is not supported for extension backends"},
		{"http3 auto", egv1a1.ClusterSettings{HTTP3: &egv1a1.BackendHTTP3Settings{Mode: new(egv1a1.BackendHTTP3ModeAuto)}}, "http3 is not supported for extension backends"},
		{"http3 always", egv1a1.ClusterSettings{HTTP3: &egv1a1.BackendHTTP3Settings{Mode: new(egv1a1.BackendHTTP3ModeAlways)}}, "http3 is not supported for extension backends"},
		{"dynamic module load balancer", egv1a1.ClusterSettings{LoadBalancer: &egv1a1.LoadBalancer{
			Type:          egv1a1.DynamicModuleLoadBalancerType,
			DynamicModule: &egv1a1.DynamicModuleLBPolicy{Name: "my-auth-module", LBPolicyName: "test"},
		}}, "DynamicModule load balancing is not supported"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resources := extensionBackendResources(t)
			child := resources.EnvoyExtensionPolicies[1]
			child.Spec.Backends[0].BackendSettings = &tc.settings
			result, routes := translateExtensionBackends(t, resources)
			require.Equal(t, metav1.ConditionFalse, extensionPolicyCondition(t, result, child.Name).Status)
			require.Contains(t, extensionPolicyCondition(t, result, child.Name).Message, tc.message)
			require.EqualValues(t, 500, *routes["httproute-1"].DirectResponse.StatusCode)
			require.Empty(t, routes["httproute-1"].EnvoyExtensions.Backends)
			require.Nil(t, routes["httproute-2"].DirectResponse)
		})
	}
}
