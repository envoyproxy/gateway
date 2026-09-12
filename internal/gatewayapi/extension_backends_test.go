// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
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
	require.Equal(t, "resolver", first.Name)
	require.Equal(t, "10.0.0.1", first.Settings[0].Endpoints[0].Host)
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
}

func TestExtensionBackendInheritanceAndAuthorization(t *testing.T) {
	resources := extensionBackendResources(t)
	parent, child := resources.EnvoyExtensionPolicies[0], resources.EnvoyExtensionPolicies[1]
	child.Spec.MergeType = new(egv1a1.JSONMerge)
	child.Spec.DynamicModule = nil
	result, routes := translateExtensionBackends(t, resources)
	require.Equal(t, metav1.ConditionTrue, extensionPolicyCondition(t, result, child.Name).Status)
	backend := routes["httproute-1"].EnvoyExtensions.DynamicModules[0].Backends[0]
	require.Equal(t, "envoy-gateway", backend.Settings[0].Metadata.Namespace)

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
	require.Equal(t, "default", backend.Settings[0].Metadata.Namespace)
	require.NotNil(t, backend.Settings[0].TLS)
	require.True(t, backend.Settings[0].TLS.UseSystemTrustStore)
	require.Equal(t, "resolver.example.com", *backend.Settings[0].TLS.SNI)
	require.Contains(t, routes["httproute-1"].EnvoyExtensions.DynamicModules[0].Name, parent.Name)
}
