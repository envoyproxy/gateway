// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"sort"

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/wasm"
)

const (
	GroupMultiClusterService = "multicluster.x-k8s.io"
	// OwningGatewayNamespaceLabel is the owner reference label used for managed infra.
	// The value should be the namespace of the accepted Envoy Gateway.
	OwningGatewayNamespaceLabel = "gateway.envoyproxy.io/owning-gateway-namespace"

	OwningGatewayClassLabel = "gateway.envoyproxy.io/owning-gatewayclass"
	// OwningGatewayNameLabel is the owner reference label used for managed infra.
	// The value should be the name of the accepted Envoy Gateway.
	OwningGatewayNameLabel = "gateway.envoyproxy.io/owning-gateway-name"

	GatewayNameLabel = "gateway.networking.k8s.io/gateway-name"

	// minEphemeralPort is the first port in the ephemeral port range.
	minEphemeralPort = 1024
	// wellKnownPortShift is the constant added to the well known port (1-1023)
	// to convert it into an ephemeral port.
	wellKnownPortShift = 10000
)

var _ TranslatorManager = (*Translator)(nil)

type TranslatorManager interface {
	Translate(resources *resource.Resources) (*TranslateResult, error)
	GetRelevantGateways(resources *resource.Resources) (acceptedGateways, failedGateways []*GatewayContext)

	RoutesTranslator
	ListenersTranslator
	AddressesTranslator
	FiltersTranslator
}

// Translator translates Gateway API resources to IRs and computes status
// for Gateway API resources.
type Translator struct {
	// GatewayControllerName is the name of the Gateway API controller
	GatewayControllerName string

	// GatewayClassName is the name of the GatewayClass
	// to process Gateways for.
	GatewayClassName gwapiv1.ObjectName

	// GlobalRateLimitEnabled is true when global
	// ratelimiting has been configured by the admin.
	GlobalRateLimitEnabled bool

	// EndpointRoutingDisabled can be set to true to use
	// the Service Cluster IP for routing to the backend
	// instead.
	EndpointRoutingDisabled bool

	// MergeGateways is true when all Gateway Listeners
	// should be merged under the parent GatewayClass.
	MergeGateways bool

	// GatewayNamespaceMode is true if controller uses gateway namespace mode for infra deployments.
	GatewayNamespaceMode bool

	// EnvoyPatchPolicyEnabled when the EnvoyPatchPolicy
	// feature is enabled.
	EnvoyPatchPolicyEnabled bool

	// BackendEnabled when the Backend feature is enabled.
	BackendEnabled bool

	// ExtensionGroupKinds stores the group/kind for all resources
	// introduced by an Extension so that the translator can
	// store referenced resources in the IR for later use.
	ExtensionGroupKinds []schema.GroupKind

	// ControllerNamespace is the namespace that Envoy Gateway controller runs in.
	ControllerNamespace string

	// WasmCache is the cache for Wasm modules.
	WasmCache wasm.Cache

	// ListenerPortShiftDisabled disables translating the
	// gateway listener port into a non privileged port
	// and reuses the specified value.
	ListenerPortShiftDisabled bool
}

type TranslateResult struct {
	resource.Resources
	XdsIR   resource.XdsIRMap   `json:"xdsIR" yaml:"xdsIR"`
	InfraIR resource.InfraIRMap `json:"infraIR" yaml:"infraIR"`
}

func newTranslateResult(gateways []*GatewayContext,
	httpRoutes []*HTTPRouteContext,
	grpcRoutes []*GRPCRouteContext,
	tlsRoutes []*TLSRouteContext,
	tcpRoutes []*TCPRouteContext,
	udpRoutes []*UDPRouteContext,
	clientTrafficPolicies []*egv1a1.ClientTrafficPolicy,
	backendTrafficPolicies []*egv1a1.BackendTrafficPolicy,
	securityPolicies []*egv1a1.SecurityPolicy,
	backendTLSPolicies []*gwapiv1a3.BackendTLSPolicy,
	envoyExtensionPolicies []*egv1a1.EnvoyExtensionPolicy,
	extPolicies []unstructured.Unstructured,
	backends []*egv1a1.Backend,
	xdsIR resource.XdsIRMap, infraIR resource.InfraIRMap,
) *TranslateResult {
	translateResult := &TranslateResult{
		XdsIR:   xdsIR,
		InfraIR: infraIR,
	}

	for _, gateway := range gateways {
		translateResult.Gateways = append(translateResult.Gateways, gateway.Gateway)
	}
	for _, httpRoute := range httpRoutes {
		translateResult.HTTPRoutes = append(translateResult.HTTPRoutes, httpRoute.HTTPRoute)
	}
	for _, grpcRoute := range grpcRoutes {
		translateResult.GRPCRoutes = append(translateResult.GRPCRoutes, grpcRoute.GRPCRoute)
	}
	for _, tlsRoute := range tlsRoutes {
		translateResult.TLSRoutes = append(translateResult.TLSRoutes, tlsRoute.TLSRoute)
	}
	for _, tcpRoute := range tcpRoutes {
		translateResult.TCPRoutes = append(translateResult.TCPRoutes, tcpRoute.TCPRoute)
	}
	for _, udpRoute := range udpRoutes {
		translateResult.UDPRoutes = append(translateResult.UDPRoutes, udpRoute.UDPRoute)
	}

	translateResult.ClientTrafficPolicies = append(translateResult.ClientTrafficPolicies, clientTrafficPolicies...)
	translateResult.BackendTrafficPolicies = append(translateResult.BackendTrafficPolicies, backendTrafficPolicies...)
	translateResult.SecurityPolicies = append(translateResult.SecurityPolicies, securityPolicies...)
	translateResult.BackendTLSPolicies = append(translateResult.BackendTLSPolicies, backendTLSPolicies...)
	translateResult.EnvoyExtensionPolicies = append(translateResult.EnvoyExtensionPolicies, envoyExtensionPolicies...)
	translateResult.ExtensionServerPolicies = append(translateResult.ExtensionServerPolicies, extPolicies...)

	translateResult.Backends = append(translateResult.Backends, backends...)
	return translateResult
}

func (t *Translator) Translate(resources *resource.Resources) (*TranslateResult, error) {
	var errs error

	// Get Gateways belonging to our GatewayClass.
	acceptedGateways, failedGateways := t.GetRelevantGateways(resources)

	// Sort gateways based on timestamp.
	sort.Slice(acceptedGateways, func(i, j int) bool {
		return acceptedGateways[i].CreationTimestamp.Before(&(acceptedGateways[j].CreationTimestamp))
	})

	// Build IR maps.
	xdsIR, infraIR := t.InitIRs(acceptedGateways)

	// Process all Listeners for all relevant Gateways.
	t.ProcessListeners(acceptedGateways, xdsIR, infraIR, resources)

	// Process EnvoyPatchPolicies
	t.ProcessEnvoyPatchPolicies(resources.EnvoyPatchPolicies, xdsIR)

	// Process all Addresses for all relevant Gateways.
	t.ProcessAddresses(acceptedGateways, xdsIR, infraIR)

	// process all Backends
	backends := t.ProcessBackends(resources.Backends)

	// Process all relevant HTTPRoutes.
	httpRoutes := t.ProcessHTTPRoutes(resources.HTTPRoutes, acceptedGateways, resources, xdsIR)

	// Process all relevant GRPCRoutes.
	grpcRoutes := t.ProcessGRPCRoutes(resources.GRPCRoutes, acceptedGateways, resources, xdsIR)

	// Process all relevant TLSRoutes.
	tlsRoutes := t.ProcessTLSRoutes(resources.TLSRoutes, acceptedGateways, resources, xdsIR)

	// Process all relevant TCPRoutes.
	tcpRoutes := t.ProcessTCPRoutes(resources.TCPRoutes, acceptedGateways, resources, xdsIR)

	// Process all relevant UDPRoutes.
	udpRoutes := t.ProcessUDPRoutes(resources.UDPRoutes, acceptedGateways, resources, xdsIR)

	// Process ClientTrafficPolicies
	clientTrafficPolicies := t.ProcessClientTrafficPolicies(resources, acceptedGateways, xdsIR, infraIR)

	// Process BackendTrafficPolicies
	routes := []RouteContext{}
	for _, h := range httpRoutes {
		routes = append(routes, h)
	}
	for _, g := range grpcRoutes {
		routes = append(routes, g)
	}
	for _, t := range tlsRoutes {
		routes = append(routes, t)
	}
	for _, t := range tcpRoutes {
		routes = append(routes, t)
	}
	for _, u := range udpRoutes {
		routes = append(routes, u)
	}

	// Process BackendTrafficPolicies
	backendTrafficPolicies := t.ProcessBackendTrafficPolicies(
		resources, acceptedGateways, routes, xdsIR)

	// Process SecurityPolicies
	securityPolicies := t.ProcessSecurityPolicies(
		resources.SecurityPolicies, acceptedGateways, routes, resources, xdsIR)

	// Process EnvoyExtensionPolicies
	envoyExtensionPolicies := t.ProcessEnvoyExtensionPolicies(
		resources.EnvoyExtensionPolicies, acceptedGateways, routes, resources, xdsIR)

	extServerPolicies, err := t.ProcessExtensionServerPolicies(
		resources.ExtensionServerPolicies, acceptedGateways, xdsIR)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	// Process global resources that are not tied to a specific listener or route
	if err := t.ProcessGlobalResources(resources, xdsIR); err != nil {
		errs = errors.Join(errs, err)
	}

	// Sort xdsIR based on the Gateway API spec
	sortXdsIRMap(xdsIR)

	// Set custom filter order if EnvoyProxy is set
	// The custom filter order will be applied when generating the HTTP filter chain.
	for _, gateway := range acceptedGateways {
		if gateway.envoyProxy != nil {
			irKey := t.getIRKey(gateway.Gateway)
			xdsIR[irKey].FilterOrder = gateway.envoyProxy.Spec.FilterOrder
		}
	}

	// Add both accepted and failed gateways to the result because we need to update the status of all gateways.
	allGateways := make([]*GatewayContext, 0, len(acceptedGateways)+len(failedGateways))
	allGateways = append(allGateways, acceptedGateways...)
	allGateways = append(allGateways, failedGateways...)
	return newTranslateResult(allGateways, httpRoutes, grpcRoutes, tlsRoutes,
		tcpRoutes, udpRoutes, clientTrafficPolicies, backendTrafficPolicies,
		securityPolicies, resources.BackendTLSPolicies, envoyExtensionPolicies,
		extServerPolicies, backends, xdsIR, infraIR), errs
}

// GetRelevantGateways returns GatewayContexts, containing a copy of the original
// Gateway with the Listener statuses reset.
func (t *Translator) GetRelevantGateways(resources *resource.Resources) (
	acceptedGateways, failedGateways []*GatewayContext,
) {
	for _, gateway := range resources.Gateways {
		if gateway == nil {
			panic("received nil gateway")
		}

		if gateway.Spec.GatewayClassName == t.GatewayClassName {
			gc := &GatewayContext{
				Gateway: gateway.DeepCopy(),
			}

			// Gateways that are not accepted by the controller because they reference an invalid EnvoyProxy.
			if status.GatewayNotAccepted(gc.Gateway) {
				failedGateways = append(failedGateways, gc)
			} else {
				gc.ResetListeners(resources)
				acceptedGateways = append(acceptedGateways, gc)
			}
		}
	}
	return
}

// InitIRs checks if mergeGateways is enabled in EnvoyProxy config and initializes XdsIR and InfraIR maps with adequate keys.
func (t *Translator) InitIRs(gateways []*GatewayContext) (map[string]*ir.Xds, map[string]*ir.Infra) {
	xdsIR := make(resource.XdsIRMap)
	infraIR := make(resource.InfraIRMap)

	for _, gateway := range gateways {
		gwXdsIR := &ir.Xds{}
		gwInfraIR := ir.NewInfra()
		labels := infrastructureLabels(gateway.Gateway)
		annotations := infrastructureAnnotations(gateway.Gateway)
		gwInfraIR.Proxy.GetProxyMetadata().Annotations = annotations

		irKey := t.IRKey(types.NamespacedName{Namespace: gateway.Namespace, Name: gateway.Name})
		if t.MergeGateways {
			maps.Copy(labels, GatewayClassOwnerLabel(string(t.GatewayClassName)))
			gwInfraIR.Proxy.GetProxyMetadata().Labels = labels
		} else {
			maps.Copy(labels, GatewayOwnerLabels(gateway.Namespace, gateway.Name))
			gwInfraIR.Proxy.GetProxyMetadata().Labels = labels
		}

		gwInfraIR.Proxy.Name = irKey
		gwInfraIR.Proxy.Namespace = t.ControllerNamespace
		gwInfraIR.Proxy.GetProxyMetadata().OwnerReference = &ir.ResourceMetadata{
			Kind: resource.KindGatewayClass,
			Name: string(t.GatewayClassName),
		}
		if t.GatewayNamespaceMode {
			gwInfraIR.Proxy.Name = gateway.Name
			gwInfraIR.Proxy.Namespace = gateway.Namespace
			gwInfraIR.Proxy.GetProxyMetadata().OwnerReference = &ir.ResourceMetadata{
				Kind: resource.KindGateway,
				Name: gateway.Name,
			}
		}
		// save the IR references in the map before the translation starts
		xdsIR[irKey] = gwXdsIR
		infraIR[irKey] = gwInfraIR
	}

	return xdsIR, infraIR
}

func (t *Translator) IRKey(gatewayNN types.NamespacedName) string {
	if t.MergeGateways {
		return string(t.GatewayClassName)
	}
	return irStringKey(gatewayNN.Namespace, gatewayNN.Name)
}

// IsEnvoyServiceRouting returns true if EnvoyProxy.Spec.RoutingType == ServiceRoutingType
// or, alternatively, if Translator.EndpointRoutingDisabled has been explicitly set to true;
// otherwise, it returns false.
func (t *Translator) IsEnvoyServiceRouting(r *egv1a1.EnvoyProxy) bool {
	if t.EndpointRoutingDisabled {
		return true
	}
	if r == nil {
		return false
	}
	switch ptr.Deref(r.Spec.RoutingType, egv1a1.EndpointRoutingType) {
	case egv1a1.ServiceRoutingType:
		return true
	case egv1a1.EndpointRoutingType:
		return false
	default:
		return false
	}
}

func infrastructureAnnotations(gtw *gwapiv1.Gateway) map[string]string {
	if gtw.Spec.Infrastructure != nil && len(gtw.Spec.Infrastructure.Annotations) > 0 {
		res := make(map[string]string)
		for k, v := range gtw.Spec.Infrastructure.Annotations {
			res[string(k)] = string(v)
		}
		return res
	}
	return nil
}

func infrastructureLabels(gtw *gwapiv1.Gateway) map[string]string {
	res := make(map[string]string)
	if gtw.Spec.Infrastructure != nil {
		for k, v := range gtw.Spec.Infrastructure.Labels {
			res[string(k)] = string(v)
		}
	}
	return res
}

// XdsIR and InfraIR map keys by default are {GatewayNamespace}/{GatewayName}, but if mergeGateways is set, they are merged under {GatewayClassName} key.
func (t *Translator) getIRKey(gateway *gwapiv1.Gateway) string {
	irKey := irStringKey(gateway.Namespace, gateway.Name)
	if t.MergeGateways {
		return string(t.GatewayClassName)
	}

	return irKey
}
