// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

const (
	KindGateway      = "Gateway"
	KindGatewayClass = "GatewayClass"
	KindGRPCRoute    = "GRPCRoute"
	KindHTTPRoute    = "HTTPRoute"
	KindNamespace    = "Namespace"
	KindTLSRoute     = "TLSRoute"
	KindTCPRoute     = "TCPRoute"
	KindUDPRoute     = "UDPRoute"
	KindService      = "Service"
	KindSecret       = "Secret"

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

	// ProxyImage is the optional proxy image to use in
	// the Infra IR. If unspecified, the default proxy
	// image will be used.
	ProxyImage string

	// GlobalRateLimitEnabled is true when global
	// ratelimiting has been configured by the admin.
	GlobalRateLimitEnabled bool

	// ExtensionGroupKinds stores the group/kind for all resources
	// introduced by an Extension so that the translator can
	// store referenced resources in the IR for later use.
	ExtensionGroupKinds []schema.GroupKind
}

type TranslateResult struct {
	Resources
	XdsIR   XdsIRMap
	InfraIR InfraIRMap
}

func newTranslateResult(gateways []*GatewayContext,
	httpRoutes []*HTTPRouteContext,
	grpcRoutes []*GRPCRouteContext,
	tlsRoutes []*TLSRouteContext,
	tcpRoutes []*TCPRouteContext,
	udpRoutes []*UDPRouteContext,
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

	return translateResult
}

func (t *Translator) Translate(resources *Resources) *TranslateResult {
	xdsIR := make(XdsIRMap)
	infraIR := make(InfraIRMap)

	// Get Gateways belonging to our GatewayClass.
	gateways := t.GetRelevantGateways(resources.Gateways)

	// Process all Listeners for all relevant Gateways.
	t.ProcessListeners(gateways, xdsIR, infraIR, resources)

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

	return newTranslateResult(gateways, httpRoutes, grpcRoutes, tlsRoutes, tcpRoutes, udpRoutes, xdsIR, infraIR)
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
