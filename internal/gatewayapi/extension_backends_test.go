// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"io"
	"os"
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
		policy.Spec.DynamicModule[0].Backends = []egv1a1.ExtensionBackend{{
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

func TestExtensionBackendConflictAndRecovery(t *testing.T) {
	resources := extensionBackendResources(t)
	parent, child := resources.EnvoyExtensionPolicies[0], resources.EnvoyExtensionPolicies[1]
	// The newer policy returns 500 when an older policy owns the cluster.
	result, routes := translateExtensionBackends(t, resources)
	condition := extensionPolicyCondition(t, result, child.Name)
	require.Equal(t, string(gwapiv1.PolicyReasonConflicted), condition.Reason)
	for _, detail := range []string{"resolver", "envoy-gateway/" + parent.Name, "Service default/service-2:8080", "Service envoy-gateway/service-2:8080", "dynamicModule[0].backends[0]"} {
		require.Contains(t, condition.Message, detail)
	}
	require.EqualValues(t, 500, *routes["httproute-1"].DirectResponse.StatusCode)
	require.Nil(t, routes["httproute-1"].EnvoyExtensions)
	require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, parent.Name).Status)
	require.Nil(t, routes["httproute-2"].DirectResponse)

	// One policy cannot map one cluster name to different backends.
	parent.Spec.DynamicModule[1].Backends = []egv1a1.ExtensionBackend{{
		Name: "resolver", BackendRef: gwapiv1.BackendObjectReference{Name: "service-1", Port: new(gwapiv1.PortNumber(8080))},
	}}
	result, routes = translateExtensionBackends(t, resources)
	require.Equal(t, string(gwapiv1.PolicyReasonInvalid), extensionPolicyCondition(t, result, parent.Name).Reason)
	require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
	require.EqualValues(t, 500, *routes["httproute-2"].DirectResponse.StatusCode)
	parent.Spec.DynamicModule[1].Backends = nil

	// Equivalent references share a cluster after normalization.
	parent.Spec.DynamicModule[0].Backends[0].BackendRef.Namespace = new(gwapiv1.Namespace("default"))
	parent.Spec.DynamicModule[0].Backends[0].BackendRef.Kind = new(gwapiv1.Kind("Service"))
	resources.ReferenceGrants = []*gwapiv1b1.ReferenceGrant{{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "allow-module"},
		Spec: gwapiv1b1.ReferenceGrantSpec{
			From: []gwapiv1b1.ReferenceGrantFrom{{Group: egv1a1.GroupName, Kind: egv1a1.KindEnvoyExtensionPolicy, Namespace: "envoy-gateway"}},
			To:   []gwapiv1b1.ReferenceGrantTo{{Group: "", Kind: "Service"}},
		},
	}}
	result, routes = translateExtensionBackends(t, resources)
	for _, policy := range result.EnvoyExtensionPolicies {
		require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, policy.Name).Status)
	}
	first := routes["httproute-1"].EnvoyExtensions.DynamicModules[0].Backends[0]
	second := routes["httproute-2"].EnvoyExtensions.DynamicModules[0].Backends[0]
	require.Same(t, first, second)
	require.Equal(t, "resolver", first.Destination.Name)
	require.Equal(t, "10.0.0.1", first.Destination.Settings[0].Endpoints[0].Host)
	for _, route := range routes {
		require.Nil(t, route.DirectResponse)
	}
}

func TestExtensionBackendDeploymentScope(t *testing.T) {
	resources := extensionBackendResources(t)
	second := resources.Gateways[0].DeepCopy()
	second.Name = "gateway-2"
	second.Spec.Listeners[0].Port = 81
	resources.Gateways = append(resources.Gateways, second)
	resources.HTTPRoutes[0].Spec.ParentRefs[0].Name = "gateway-2"
	result, _ := translateExtensionBackends(t, resources)
	for _, policy := range result.EnvoyExtensionPolicies {
		require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, policy.Name).Status)
	}

	resources.EnvoyProxyForGatewayClass = resources.EnvoyProxiesForGateways[0].DeepCopy()
	resources.EnvoyProxyForGatewayClass.Spec.MergeGateways = new(true)
	for _, gateway := range resources.Gateways {
		gateway.Spec.Infrastructure = nil
	}
	result, routes := translateExtensionBackends(t, resources)
	require.Len(t, result.XdsIR, 1)
	require.Equal(t, string(gwapiv1.PolicyReasonConflicted), extensionPolicyCondition(t, result, "policy-for-http-route").Reason)
	require.EqualValues(t, 500, *routes["httproute-1"].DirectResponse.StatusCode)

	for _, policy := range resources.EnvoyExtensionPolicies {
		if policy.Name == "policy-for-http-route" {
			policy.Spec.DynamicModule[0].Backends[0].Name = "envoy-gateway-class"
		}
	}
	result, _ = translateExtensionBackends(t, resources)
	condition := extensionPolicyCondition(t, result, "policy-for-http-route")
	require.Equal(t, string(gwapiv1.PolicyReasonInvalid), condition.Reason)
	require.Contains(t, condition.Message, "reserved")

	for _, policy := range resources.EnvoyExtensionPolicies {
		if policy.Name == "policy-for-http-route" {
			policy.Spec.DynamicModule[0].Backends[0].Name = "tracing"
		}
	}
	result, routes = translateExtensionBackends(t, resources)
	condition = extensionPolicyCondition(t, result, "policy-for-http-route")
	require.Equal(t, metav1.ConditionFalse, condition.Status)
	require.Equal(t, string(gwapiv1.PolicyReasonInvalid), condition.Reason)
	require.Contains(t, condition.Message, `cluster name "tracing" is reserved`)
	require.EqualValues(t, 500, *routes["httproute-1"].DirectResponse.StatusCode)
}

func TestExtensionBackendSettings(t *testing.T) {
	settings := func(timeout string) *egv1a1.ClusterSettings {
		return &egv1a1.ClusterSettings{
			LoadBalancer: &egv1a1.LoadBalancer{Type: egv1a1.RandomLoadBalancerType},
			Timeout: &egv1a1.Timeout{TCP: &egv1a1.TCPTimeout{
				ConnectTimeout: new(gwapiv1.Duration(timeout)),
			}},
		}
	}
	requestTimeouts := &egv1a1.HTTPTimeout{
		RequestTimeout: new(gwapiv1.Duration("3s")), StreamIdleTimeout: new(gwapiv1.Duration("4s")),
	}
	settingsWithRequestTimeouts := settings("2s")
	settingsWithRequestTimeouts.Timeout.HTTP = requestTimeouts.DeepCopy()
	for _, tc := range []struct {
		name         string
		parent       *egv1a1.ClusterSettings
		child        *egv1a1.ClusterSettings
		withinPolicy bool
		wantReason   gwapiv1.PolicyConditionReason
	}{
		{name: "omitted and empty settings share", child: &egv1a1.ClusterSettings{}},
		{name: "equivalent settings share", parent: settings("2s"), child: settings("2000ms")},
		{name: "request timeouts do not affect cluster sharing", parent: settings("2s"), child: settingsWithRequestTimeouts},
		{name: "request timeouts alone share with omitted settings", child: &egv1a1.ClusterSettings{
			Timeout: &egv1a1.Timeout{HTTP: requestTimeouts.DeepCopy()},
		}},
		{name: "different settings conflict", parent: settings("2s"), child: settings("3s"), wantReason: gwapiv1.PolicyReasonConflicted},
		{name: "configured and omitted settings conflict", parent: settings("2s"), wantReason: gwapiv1.PolicyReasonConflicted},
		{name: "different settings within policy are invalid", parent: settings("2s"), child: settings("3s"), withinPolicy: true, wantReason: gwapiv1.PolicyReasonInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resources := extensionBackendResources(t)
			parent, child := resources.EnvoyExtensionPolicies[0], resources.EnvoyExtensionPolicies[1]
			parent.Spec.DynamicModule[0].Backends[0].BackendRef.Namespace = new(gwapiv1.Namespace("default"))
			resources.ReferenceGrants = []*gwapiv1b1.ReferenceGrant{{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "allow-module"},
				Spec: gwapiv1b1.ReferenceGrantSpec{
					From: []gwapiv1b1.ReferenceGrantFrom{{Group: egv1a1.GroupName, Kind: egv1a1.KindEnvoyExtensionPolicy, Namespace: "envoy-gateway"}},
					To:   []gwapiv1b1.ReferenceGrantTo{{Group: "", Kind: "Service"}},
				},
			}}
			parent.Spec.DynamicModule[0].Backends[0].BackendSettings = tc.parent
			child.Spec.DynamicModule[0].Backends[0].BackendSettings = tc.child
			rejectedPolicy, rejectedRoute := child.Name, "httproute-1"
			if tc.withinPolicy {
				backend := parent.Spec.DynamicModule[0].Backends[0].DeepCopy()
				backend.BackendSettings = tc.child
				parent.Spec.DynamicModule[1].Backends = []egv1a1.ExtensionBackend{*backend}
				child.Spec.DynamicModule[0].Backends[0].Name = "child-resolver"
				rejectedPolicy, rejectedRoute = parent.Name, "httproute-2"
			}
			result, routes := translateExtensionBackends(t, resources)
			if tc.wantReason != "" {
				condition := extensionPolicyCondition(t, result, rejectedPolicy)
				require.Equal(t, metav1.ConditionFalse, condition.Status)
				require.Equal(t, string(tc.wantReason), condition.Reason)
				require.Contains(t, condition.Message, "settings")
				require.EqualValues(t, 500, *routes[rejectedRoute].DirectResponse.StatusCode)
				return
			}
			for _, policy := range result.EnvoyExtensionPolicies {
				require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, policy.Name).Status)
			}
			first := routes["httproute-1"].EnvoyExtensions.DynamicModules[0].Backends[0]
			second := routes["httproute-2"].EnvoyExtensions.DynamicModules[0].Backends[0]
			require.Same(t, first, second)
			if tc.parent == nil {
				require.Nil(t, first.Traffic)
			} else {
				require.NotNil(t, first.Traffic.LoadBalancer.Random)
				require.Equal(t, 2*time.Second, first.Traffic.Timeout.TCP.ConnectTimeout.Duration)
				require.Nil(t, first.Traffic.Timeout.HTTP)
			}
		})
	}

	for _, tc := range []struct {
		name     string
		settings *egv1a1.ClusterSettings
		message  string
	}{
		{name: "invalid settings reject the policy", settings: settings("invalid"), message: "invalid ConnectTimeout"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resources := extensionBackendResources(t)
			child := resources.EnvoyExtensionPolicies[1]
			child.Spec.DynamicModule[0].Backends[0].BackendSettings = tc.settings
			result, routes := translateExtensionBackends(t, resources)
			condition := extensionPolicyCondition(t, result, child.Name)
			require.Equal(t, metav1.ConditionFalse, condition.Status)
			require.Equal(t, string(gwapiv1.PolicyReasonInvalid), condition.Reason)
			require.Contains(t, condition.Message, tc.message)
			require.EqualValues(t, 500, *routes["httproute-1"].DirectResponse.StatusCode)
		})
	}
}

func TestExtensionBackendInheritanceAndAuthorization(t *testing.T) {
	resources := extensionBackendResources(t)
	parent, child := resources.EnvoyExtensionPolicies[0], resources.EnvoyExtensionPolicies[1]
	parent.Spec.DynamicModule[0].Backends[0].BackendSettings = &egv1a1.ClusterSettings{
		LoadBalancer: &egv1a1.LoadBalancer{Type: egv1a1.RandomLoadBalancerType},
	}
	child.Spec.MergeType = new(egv1a1.JSONMerge)
	child.Spec.DynamicModule = nil
	result, routes := translateExtensionBackends(t, resources)
	require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
	backend := routes["httproute-1"].EnvoyExtensions.DynamicModules[0].Backends[0]
	require.Equal(t, "envoy-gateway", backend.Destination.Settings[0].Metadata.Namespace)
	require.NotNil(t, backend.Traffic.LoadBalancer.Random)

	parent.Spec.DynamicModule[0].Backends[0].BackendRef.Namespace = new(gwapiv1.Namespace("default"))
	result, routes = translateExtensionBackends(t, resources)
	require.Contains(t, extensionPolicyCondition(t, result, child.Name).Message, "not permitted by any ReferenceGrant")
	require.EqualValues(t, 500, *routes["httproute-1"].DirectResponse.StatusCode)

	resources.ReferenceGrants = []*gwapiv1b1.ReferenceGrant{{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "allow-module"},
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
	for _, policy := range result.EnvoyExtensionPolicies {
		require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, policy.Name).Status)
	}
	backend = routes["httproute-1"].EnvoyExtensions.DynamicModules[0].Backends[0]
	require.Equal(t, "default", backend.Destination.Settings[0].Metadata.Namespace)
	require.NotNil(t, backend.Destination.Settings[0].TLS)
	require.True(t, backend.Destination.Settings[0].TLS.UseSystemTrustStore)
	require.Equal(t, "resolver.example.com", *backend.Destination.Settings[0].TLS.SNI)
	require.Contains(t, routes["httproute-1"].EnvoyExtensions.DynamicModules[0].Name, parent.Name)
}

func extensionBackendTLSResources(t *testing.T) *resource.Resources {
	t.Helper()
	resources := extensionBackendResources(t)
	data, err := os.ReadFile("testdata/gateway-tls-frontend-backend.in.yaml")
	require.NoError(t, err)
	fixture := &resource.Resources{}
	mustUnmarshal(t, data, fixture)
	require.GreaterOrEqual(t, len(fixture.Secrets), 2)
	for index, name := range []string{"client-a", "client-b"} {
		secret := fixture.Secrets[index].DeepCopy()
		secret.Name, secret.Namespace = name, "envoy-gateway"
		resources.Secrets = append(resources.Secrets, secret)
	}
	resources.EnvoyProxyForGatewayClass = resources.EnvoyProxiesForGateways[0].DeepCopy()
	gateway := resources.Gateways[0]
	gateway.Spec.Infrastructure = nil
	gateway.Spec.TLS = &gwapiv1.GatewayTLSConfig{Backend: &gwapiv1.GatewayBackendTLS{
		ClientCertificateRef: &gwapiv1.SecretObjectReference{Name: "client-a"},
	}}
	second := gateway.DeepCopy()
	second.Name = "gateway-2"
	second.Spec.Listeners[0].Port = 81
	second.Spec.TLS.Backend.ClientCertificateRef.Name = "client-b"
	resources.Gateways = append(resources.Gateways, second)
	resources.HTTPRoutes[0].Spec.ParentRefs[0].Name = "gateway-2"
	resources.EnvoyExtensionPolicies[0].Spec.DynamicModule[0].Backends[0].BackendRef.Namespace = new(gwapiv1.Namespace("default"))
	resources.ReferenceGrants = []*gwapiv1b1.ReferenceGrant{{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "allow-module"},
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
	return resources
}

func TestExtensionBackendTLSConflict(t *testing.T) {
	for _, tc := range []struct {
		name        string
		merge       bool
		certificate string
		conflict    bool
	}{
		{name: "separate deployments preserve different certificates", certificate: "client-b"},
		{name: "merged deployments reject different certificates", merge: true, certificate: "client-b", conflict: true},
		{name: "merged deployments reject a missing certificate", merge: true, conflict: true},
		{name: "merged deployments share the same certificate", merge: true, certificate: "client-a"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resources := extensionBackendTLSResources(t)
			resources.EnvoyProxyForGatewayClass.Spec.MergeGateways = new(tc.merge)
			if tc.certificate == "" {
				resources.Gateways[1].Spec.TLS = nil
			} else {
				resources.Gateways[1].Spec.TLS.Backend.ClientCertificateRef.Name = gwapiv1.ObjectName(tc.certificate)
			}
			parent, child := resources.EnvoyExtensionPolicies[0], resources.EnvoyExtensionPolicies[1]
			result, routes := translateExtensionBackends(t, resources)
			require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, parent.Name).Status)
			parentBackend := routes["httproute-2"].EnvoyExtensions.DynamicModules[0].Backends[0]
			require.Equal(t, "envoy-gateway/client-a", parentBackend.Destination.Settings[0].TLS.ClientCertificates[0].Name)
			condition := extensionPolicyCondition(t, result, child.Name)
			if tc.conflict {
				require.Equal(t, metav1.ConditionFalse, condition.Status)
				require.Equal(t, string(gwapiv1.PolicyReasonConflicted), condition.Reason)
				for _, detail := range []string{"resolved transport configuration", "Gateway envoy-gateway/gateway-1", "Gateway envoy-gateway/gateway-2"} {
					require.Contains(t, condition.Message, detail)
				}
				require.Nil(t, routes["httproute-1"].EnvoyExtensions)
				require.EqualValues(t, 500, *routes["httproute-1"].DirectResponse.StatusCode)
				// Aligning the certificates clears the conflict on the next translation.
				resources.Gateways[1].Spec.TLS = resources.Gateways[0].Spec.TLS.DeepCopy()
				result, routes = translateExtensionBackends(t, resources)
				require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
				require.Nil(t, routes["httproute-1"].DirectResponse)
				require.Same(t, routes["httproute-2"].EnvoyExtensions.DynamicModules[0].Backends[0],
					routes["httproute-1"].EnvoyExtensions.DynamicModules[0].Backends[0])
				return
			}
			require.Equal(t, metav1.ConditionTrue, condition.Status)
			childBackend := routes["httproute-1"].EnvoyExtensions.DynamicModules[0].Backends[0]
			require.Equal(t, "envoy-gateway/"+tc.certificate, childBackend.Destination.Settings[0].TLS.ClientCertificates[0].Name)
			if tc.merge {
				require.Same(t, parentBackend, childBackend)
			}
			for _, route := range routes {
				require.Nil(t, route.DirectResponse)
			}
		})
	}
}

func TestExtensionBackendTLSConflictAcrossPolicyGateways(t *testing.T) {
	for _, inherited := range []bool{false, true} {
		for _, certificate := range []string{"client-a", "client-b"} {
			t.Run(fmt.Sprintf("inherited=%t/certificate=%s", inherited, certificate), func(t *testing.T) {
				resources := extensionBackendTLSResources(t)
				resources.EnvoyProxyForGatewayClass.Spec.MergeGateways = new(true)
				resources.Gateways[1].Spec.TLS.Backend.ClientCertificateRef.Name = gwapiv1.ObjectName(certificate)
				parent, child := resources.EnvoyExtensionPolicies[0], resources.EnvoyExtensionPolicies[1]
				parent.Spec.TargetRef = nil
				parent.Spec.TargetRefs = []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					{LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{Group: gwapiv1.GroupName, Kind: "Gateway", Name: "gateway-1"}},
					{LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{Group: gwapiv1.GroupName, Kind: "Gateway", Name: "gateway-2"}},
				}
				if inherited {
					child.Spec.MergeType = new(egv1a1.JSONMerge)
					child.Spec.DynamicModule = nil
				} else {
					resources.EnvoyExtensionPolicies = resources.EnvoyExtensionPolicies[:1]
				}
				result, routes := translateExtensionBackends(t, resources)
				for _, policy := range result.EnvoyExtensionPolicies {
					for _, ancestor := range policy.Status.Ancestors {
						condition := meta.FindStatusCondition(ancestor.Conditions, string(gwapiv1.PolicyConditionAccepted))
						require.NotNil(t, condition)
						if certificate == "client-b" {
							require.Equal(t, metav1.ConditionFalse, condition.Status)
							require.Equal(t, string(gwapiv1.PolicyReasonInvalid), condition.Reason)
							require.Contains(t, condition.Message, "resolved transport configuration")
						} else {
							require.Equal(t, metav1.ConditionTrue, condition.Status)
						}
					}
				}
				if certificate == "client-b" {
					for _, route := range routes {
						require.Nil(t, route.EnvoyExtensions)
						require.EqualValues(t, 500, *route.DirectResponse.StatusCode)
					}
				} else {
					first := routes["httproute-1"].EnvoyExtensions.DynamicModules[0].Backends[0]
					second := routes["httproute-2"].EnvoyExtensions.DynamicModules[0].Backends[0]
					require.Same(t, first, second)
					require.Len(t, first.Destination.Settings[0].TLS.ClientCertificates, 1)
					require.Equal(t, "envoy-gateway/client-a", first.Destination.Settings[0].TLS.ClientCertificates[0].Name)
				}
			})
		}
	}
}
