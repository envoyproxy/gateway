// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

const (
	KindEnvoyProxy    = "EnvoyProxy"
	KindGateway       = "Gateway"
	KindGatewayClass  = "GatewayClass"
	KindGRPCRoute     = "GRPCRoute"
	KindHTTPRoute     = "HTTPRoute"
	KindNamespace     = "Namespace"
	KindTLSRoute      = "TLSRoute"
	KindTCPRoute      = "TCPRoute"
	KindUDPRoute      = "UDPRoute"
	KindService       = "Service"
	KindServiceImport = "ServiceImport"
	KindSecret        = "Secret"

	GroupMultiClusterService = "multicluster.x-k8s.io"
	// OwningGatewayNamespaceLabel is the owner reference label used for managed infra.
	// The value should be the namespace of the accepted Envoy Gateway.
	OwningGatewayNamespaceLabel = "gateway.envoyproxy.io/owning-gateway-namespace"

	// OwningGatewayNameLabel is the owner reference label used for managed infra.
	// The value should be the name of the accepted Envoy Gateway.
	OwningGatewayNameLabel = "gateway.envoyproxy.io/owning-gateway-name"

	// minEphemeralPort is the first port in the ephemeral port range.
	minEphemeralPort = 1024
	// wellKnownPortShift is the constant added to the well known port (1-1023)
	// to convert it into an ephemeral port.
	wellKnownPortShift = 10000
)

var _ TranslatorManager = (*Translator)(nil)

type TranslatorManager interface {
	Translate(resources *Resources) *TranslateResult
	GetRelevantGateways(gateways []*v1beta1.Gateway) []*GatewayContext

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
	GatewayClassName v1beta1.ObjectName

	// GlobalRateLimitEnabled is true when global
	// ratelimiting has been configured by the admin.
	GlobalRateLimitEnabled bool

	// EndpointRoutingDisabled can be set to true to use
	// the Service Cluster IP for routing to the backend
	// instead.
	EndpointRoutingDisabled bool

	// ExtensionGroupKinds stores the group/kind for all resources
	// introduced by an Extension so that the translator can
	// store referenced resources in the IR for later use.
	ExtensionGroupKinds []schema.GroupKind
}

type TranslateResult struct {
	Resources
	XdsIR   XdsIRMap   `json:"xdsIR" yaml:"xdsIR"`
	InfraIR InfraIRMap `json:"infraIR" yaml:"infraIR"`
}

func newTranslateResult(gateways []*GatewayContext,
	httpRoutes []*HTTPRouteContext,
	grpcRoutes []*GRPCRouteContext,
	tlsRoutes []*TLSRouteContext,
	tcpRoutes []*TCPRouteContext,
	udpRoutes []*UDPRouteContext,
	clientTrafficPolicies []*egv1a1.ClientTrafficPolicy,
	backendTrafficPolicies []*egv1a1.BackendTrafficPolicy,
	xdsIR XdsIRMap, infraIR InfraIRMap) *TranslateResult {
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

	return translateResult
}

func (t *Translator) Translate(resources *Resources) *TranslateResult {
	xdsIR := make(XdsIRMap)
	infraIR := make(InfraIRMap)

	// Get Gateways belonging to our GatewayClass.
	gateways := t.GetRelevantGateways(resources.Gateways)

	// Get Routes belonging to our GatewayClass.
	routes := t.GetRelevantRoutes(gateways,
		resources.HTTPRoutes,
		resources.GRPCRoutes,
		resources.TLSRoutes,
		resources.TCPRoutes,
		resources.UDPRoutes,
	)

	// Process all Listeners for all relevant Gateways.
	t.ProcessListeners(gateways, xdsIR, infraIR, resources)

	// Process EnvoyPatchPolicies
	t.ProcessEnvoyPatchPolicies(resources.EnvoyPatchPolicies, xdsIR)

	// Process ClientTrafficPolicies
	clientTrafficPolicies := ProcessClientTrafficPolicies(resources.ClientTrafficPolicies, gateways, xdsIR)

	// Process BackendTrafficPolicies
	backendTrafficPolicies := ProcessBackendTrafficPolicies(resources.BackendTrafficPolicies, gateways, routes, xdsIR)

	// Process all Addresses for all relevant Gateways.
	t.ProcessAddresses(gateways, xdsIR, infraIR, resources)

	// Process all relevant HTTPRoutes.
	httpRoutes := t.ProcessHTTPRoutes(resources.HTTPRoutes, gateways, resources, xdsIR)

	// Process all relevant GRPCRoutes.
	grpcRoutes := t.ProcessGRPCRoutes(resources.GRPCRoutes, gateways, resources, xdsIR)

	// Process all relevant TLSRoutes.
	tlsRoutes := t.ProcessTLSRoutes(resources.TLSRoutes, gateways, resources, xdsIR)

	// Process all relevant TCPRoutes.
	tcpRoutes := t.ProcessTCPRoutes(resources.TCPRoutes, gateways, resources, xdsIR)

	// Process all relevant UDPRoutes.
	udpRoutes := t.ProcessUDPRoutes(resources.UDPRoutes, gateways, resources, xdsIR)

	// Sort xdsIR based on the Gateway API spec
	sortXdsIRMap(xdsIR)

	return newTranslateResult(gateways, httpRoutes, grpcRoutes, tlsRoutes, tcpRoutes, udpRoutes, clientTrafficPolicies, backendTrafficPolicies, xdsIR, infraIR)
}

// GetRelevantGateways returns GatewayContexts, containing a copy of the original
// Gateway with the Listener statuses reset.
func (t *Translator) GetRelevantGateways(gateways []*v1beta1.Gateway) []*GatewayContext {
	var relevant []*GatewayContext

	for _, gateway := range gateways {
		if gateway == nil {
			panic("received nil gateway")
		}

		if gateway.Spec.GatewayClassName == t.GatewayClassName {
			gc := &GatewayContext{
				Gateway: gateway.DeepCopy(),
			}
			gc.ResetListeners()

			relevant = append(relevant, gc)
		}
	}

	return relevant
}

func (t *Translator) GetRelevantRoutes(relevantGateways []*GatewayContext,
	httpRoutes []*v1beta1.HTTPRoute,
	grpcRoutes []*v1alpha2.GRPCRoute,
	tlsRoutes []*v1alpha2.TLSRoute,
	tcpRoutes []*v1alpha2.TCPRoute,
	udpRoutes []*v1alpha2.UDPRoute) []RouteContext {

	relevantRoutes := []RouteContext{}

	// build a map of the relevant Gateways for faster lookup
	gateways := map[types.NamespacedName]bool{}
	for _, gw := range relevantGateways {
		gw := gw
		gateways[types.NamespacedName{Name: gw.Name, Namespace: gw.Namespace}] = true
	}

	// Check all the different types of routes to see if any belong to one of our relevant gateways
	for _, httpRoute := range httpRoutes {
		for _, parentGW := range httpRoute.Spec.ParentRefs {
			key := types.NamespacedName{Name: string(parentGW.Name)}
			if parentGW.Namespace != nil {
				key.Namespace = string(*parentGW.Namespace)
			}
			if _, ok := gateways[key]; ok {
				relevantRoutes = append(relevantRoutes, &HTTPRouteContext{
					GatewayControllerName: t.GatewayControllerName,
					HTTPRoute:             httpRoute.DeepCopy(),
				})
				break
			}
		}
	}

	for _, grpcRoute := range grpcRoutes {
		for _, parentGW := range grpcRoute.Spec.ParentRefs {
			key := types.NamespacedName{Name: string(parentGW.Name)}
			if parentGW.Namespace != nil {
				key.Namespace = string(*parentGW.Namespace)
			}
			if _, ok := gateways[key]; ok {
				relevantRoutes = append(relevantRoutes, &GRPCRouteContext{
					GatewayControllerName: t.GatewayControllerName,
					GRPCRoute:             grpcRoute.DeepCopy(),
				})
				break
			}
		}
	}

	for _, tlsRoute := range tlsRoutes {
		for _, parentGW := range tlsRoute.Spec.ParentRefs {
			key := types.NamespacedName{Name: string(parentGW.Name)}
			if parentGW.Namespace != nil {
				key.Namespace = string(*parentGW.Namespace)
			}
			if _, ok := gateways[key]; ok {
				relevantRoutes = append(relevantRoutes, &TLSRouteContext{
					GatewayControllerName: t.GatewayControllerName,
					TLSRoute:              tlsRoute.DeepCopy(),
				})
				break
			}
		}
	}

	for _, tcpRoute := range tcpRoutes {
		for _, parentGW := range tcpRoute.Spec.ParentRefs {
			key := types.NamespacedName{Name: string(parentGW.Name)}
			if parentGW.Namespace != nil {
				key.Namespace = string(*parentGW.Namespace)
			}
			if _, ok := gateways[key]; ok {
				relevantRoutes = append(relevantRoutes, &TCPRouteContext{
					GatewayControllerName: t.GatewayControllerName,
					TCPRoute:              tcpRoute.DeepCopy(),
				})
				break
			}
		}
	}

	for _, udpRoute := range udpRoutes {
		for _, parentGW := range udpRoute.Spec.ParentRefs {
			key := types.NamespacedName{Name: string(parentGW.Name)}
			if parentGW.Namespace != nil {
				key.Namespace = string(*parentGW.Namespace)
			}
			if _, ok := gateways[key]; ok {
				relevantRoutes = append(relevantRoutes, &UDPRouteContext{
					GatewayControllerName: t.GatewayControllerName,
					UDPRoute:              udpRoute.DeepCopy(),
				})
				break
			}
		}
	}

	return relevantRoutes
}
