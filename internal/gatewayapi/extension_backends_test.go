// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
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
	resources.Sort()
	translator := &Translator{
		GatewayControllerName: egv1a1.GatewayControllerName, GatewayClassName: "envoy-gateway-class",
		ControllerNamespace: "envoy-gateway-system", EndpointRoutingDisabled: true, BackendEnabled: true,
		MergeGateways: IsMergeGatewaysEnabled(resources), Logger: logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo),
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
			first := routes["httproute-1"].EnvoyExtensions.Backends["resolver"]
			second := routes["httproute-2"].EnvoyExtensions.Backends["resolver"]
			require.NotEqual(t, first.Name, second.Name)
			require.Equal(t, "default", first.Settings[0].Metadata.Namespace)
			require.Equal(t, "envoy-gateway", second.Settings[0].Metadata.Namespace)
			for _, route := range routes {
				require.Nil(t, route.DirectResponse)
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
			backend := routes["httproute-1"].EnvoyExtensions.Backends["resolver"]
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
		require.NotEqual(t, alias, destination.Name)
	}

	slices.Reverse(child.Spec.Backends)
	result, routes = translateExtensionBackends(t, resources)
	require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
	for alias, destination := range routes["httproute-1"].EnvoyExtensions.Backends {
		require.Equal(t, before[alias].Name, destination.Name)
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
			backend := routes["httproute-1"].EnvoyExtensions.Backends["resolver"]
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
			backend = routes["httproute-1"].EnvoyExtensions.Backends["resolver"]
			require.Equal(t, "default", backend.Settings[0].Metadata.Namespace)
			require.True(t, backend.Settings[0].TLS.UseSystemTrustStore)
			require.Equal(t, "resolver.example.com", *backend.Settings[0].TLS.SNI)

			// A child replaces the list as a whole and resolves its own references.
			parent.Spec.Backends = append(parent.Spec.Backends, egv1a1.ExtensionBackend{Name: "parent-only", BackendRef: parent.Spec.Backends[0].BackendRef})
			child.Spec.Backends = []egv1a1.ExtensionBackend{{Name: "resolver", BackendRef: gwapiv1.BackendObjectReference{Name: "service-1", Port: new(gwapiv1.PortNumber(8080))}}}
			result, routes = translateExtensionBackends(t, resources)
			require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
			require.Len(t, routes["httproute-1"].EnvoyExtensions.Backends, 1)
			require.Contains(t, routes["httproute-1"].EnvoyExtensions.Backends["resolver"].Name, child.Name)
			require.Equal(t, "service-1", routes["httproute-1"].EnvoyExtensions.Backends["resolver"].Settings[0].Metadata.Name)
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
	first := routes["httproute-1"].EnvoyExtensions.Backends["resolver"]
	secondBackend := routes["httproute-2"].EnvoyExtensions.Backends["resolver"]
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
	backend := routes["httproute-1"].EnvoyExtensions.Backends["resolver"]
	require.NotNil(t, backend.Traffic.LoadBalancer.RoundRobin)
	require.Equal(t, 3*time.Second, backend.Traffic.Timeout.TCP.ConnectTimeout.Duration)
	require.EqualValues(t, 17, *backend.Traffic.CircuitBreaker.MaxConnections)
	require.Equal(t, backend.Traffic, routes["httproute-2"].EnvoyExtensions.Backends["resolver"].Traffic)
	child.Spec.Backends = []egv1a1.ExtensionBackend{{Name: "resolver", BackendRef: parent.Spec.Backends[0].BackendRef}}
	result, routes = translateExtensionBackends(t, resources)
	require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
	require.Nil(t, routes["httproute-1"].EnvoyExtensions.Backends["resolver"].Traffic)
	require.NotNil(t, routes["httproute-2"].EnvoyExtensions.Backends["resolver"].Traffic)
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
			require.Contains(t, extensionPolicyCondition(t, result, child.Name).Message, tc.message)
			require.EqualValues(t, 500, *routes["httproute-1"].DirectResponse.StatusCode)
			require.Empty(t, routes["httproute-1"].EnvoyExtensions.Backends)
			require.Nil(t, routes["httproute-2"].DirectResponse)
		})
	}
}
