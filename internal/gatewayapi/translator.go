// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/api/v1alpha1/validation"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/wasm"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
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
	// TranslatorContext holds pre-indexed resource maps for efficient lookup resources
	// during translation operations.
	*TranslatorContext

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

	// LuaEnvoyExtensionPolicyDisabled when the Lua EnvoyExtensionPolicy feature is disabled.
	LuaEnvoyExtensionPolicyDisabled bool

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

	// RunningOnHost indicates whether Envoy Gateway is running locally on the host machine.
	//
	// When running on the local host using the Host infrastructure provider, disable translating the
	// gateway listener port into a non-privileged port and reuse the specified value.
	// Also, allow loopback IP addresses in Backend endpoints, as the threat model is different from
	// the cluster environment and the related security risk is not applicable.
	RunningOnHost bool

	// oidcDiscoveryCache is the cache for OIDC configurations discovered from issuer's well-known URL.
	oidcDiscoveryCache *oidcDiscoveryCache

	// Logger is the logger used by the translator.
	Logger logging.Logger
}

type TranslateResult struct {
	resource.Resources
	XdsIR   resource.XdsIRMap   `json:"xdsIR" yaml:"xdsIR"`
	InfraIR resource.InfraIRMap `json:"infraIR" yaml:"infraIR"`
}

func newTranslateResult(
	gc *gwapiv1.GatewayClass,
	gateways []*GatewayContext,
	httpRoutes []*HTTPRouteContext,
	grpcRoutes []*GRPCRouteContext,
	tlsRoutes []*TLSRouteContext,
	tcpRoutes []*TCPRouteContext,
	udpRoutes []*UDPRouteContext,
	clientTrafficPolicies []*egv1a1.ClientTrafficPolicy,
	backendTrafficPolicies []*egv1a1.BackendTrafficPolicy,
	securityPolicies []*egv1a1.SecurityPolicy,
	backendTLSPolicies []*gwapiv1.BackendTLSPolicy,
	envoyExtensionPolicies []*egv1a1.EnvoyExtensionPolicy,
	extPolicies []unstructured.Unstructured,
	backends []*egv1a1.Backend,
	xListenerSets []*gwapiv1.ListenerSet,
	xdsIR resource.XdsIRMap, infraIR resource.InfraIRMap,
) *TranslateResult {
	translateResult := &TranslateResult{
		XdsIR:   xdsIR,
		InfraIR: infraIR,
	}

	translateResult.GatewayClass = gc

	if n := len(gateways); n > 0 {
		translateResult.Gateways = make([]*gwapiv1.Gateway, n)
		for i, gateway := range gateways {
			translateResult.Gateways[i] = gateway.Gateway
		}
	}

	if n := len(httpRoutes); n > 0 {
		translateResult.HTTPRoutes = make([]*gwapiv1.HTTPRoute, n)
		for i, httpRoute := range httpRoutes {
			translateResult.HTTPRoutes[i] = httpRoute.HTTPRoute
		}
	}

	if n := len(grpcRoutes); n > 0 {
		translateResult.GRPCRoutes = make([]*gwapiv1.GRPCRoute, n)
		for i, grpcRoute := range grpcRoutes {
			translateResult.GRPCRoutes[i] = grpcRoute.GRPCRoute
		}
	}

	if n := len(tlsRoutes); n > 0 {
		translateResult.TLSRoutes = make([]*gwapiv1.TLSRoute, n)
		for i, tlsRoute := range tlsRoutes {
			translateResult.TLSRoutes[i] = tlsRoute.TLSRoute
		}
	}

	if n := len(tcpRoutes); n > 0 {
		translateResult.TCPRoutes = make([]*gwapiv1a2.TCPRoute, n)
		for i, tcpRoute := range tcpRoutes {
			translateResult.TCPRoutes[i] = tcpRoute.TCPRoute
		}
	}

	if n := len(udpRoutes); n > 0 {
		translateResult.UDPRoutes = make([]*gwapiv1a2.UDPRoute, n)
		for i, udpRoute := range udpRoutes {
			translateResult.UDPRoutes[i] = udpRoute.UDPRoute
		}
	}

	if len(clientTrafficPolicies) > 0 {
		translateResult.ClientTrafficPolicies = clientTrafficPolicies
	}
	if len(backendTrafficPolicies) > 0 {
		translateResult.BackendTrafficPolicies = backendTrafficPolicies
	}
	if len(securityPolicies) > 0 {
		translateResult.SecurityPolicies = securityPolicies
	}
	if len(backendTLSPolicies) > 0 {
		translateResult.BackendTLSPolicies = backendTLSPolicies
	}
	if len(envoyExtensionPolicies) > 0 {
		translateResult.EnvoyExtensionPolicies = envoyExtensionPolicies
	}
	if len(extPolicies) > 0 {
		translateResult.ExtensionServerPolicies = extPolicies
	}
	if len(backends) > 0 {
		translateResult.Backends = backends
	}
	if len(xListenerSets) > 0 {
		translateResult.ListenerSets = xListenerSets
	}

	return translateResult
}

func (t *Translator) Translate(resources *resource.Resources) (*TranslateResult, error) {
	var errs error

	// Preprocessing to improve get resources operations performance.
	translatorContext := &TranslatorContext{}
	translatorContext.SetNamespaces(resources.Namespaces)
	translatorContext.SetServices(resources.Services)
	translatorContext.SetServiceImports(resources.ServiceImports)
	translatorContext.SetBackends(resources.Backends)
	translatorContext.SetSecrets(resources.Secrets)
	translatorContext.SetConfigMaps(resources.ConfigMaps)
	translatorContext.SetClusterTrustBundles(resources.ClusterTrustBundles)
	translatorContext.SetEndpointSlicesForBackend(resources.EndpointSlices)

	t.TranslatorContext = translatorContext

	// Get Gateways belonging to our GatewayClass.
	acceptedGateways, failedGateways := t.GetRelevantGateways(resources)

	// Gateways are already sorted by the provider layer

	// Build IR maps.
	xdsIR, infraIR := t.InitIRs(acceptedGateways, failedGateways)

	// Build pre-computed BTP RoutingType index for O(1) lookups in processDestination.
	t.BTPRoutingTypeIndex = BuildBTPRoutingTypeIndex(
		resources.BackendTrafficPolicies,
		routesToObjects(resources),
		acceptedGateways,
	)

	// Process ListenerSets and attach them to the relevant Gateways
	t.ProcessListenerSets(resources.ListenerSets, acceptedGateways)

	// Process all Listeners for all relevant Gateways.
	t.ProcessListeners(acceptedGateways, xdsIR, infraIR, resources)

	// Compute ListenerSet status based on listener processing results
	// This should be done after ProcessListeners because ListenerSet status depends on listener processing results
	t.ProcessListenerSetStatus(resources.ListenerSets)

	// Process EnvoyPatchPolicies
	t.ProcessEnvoyPatchPolicies(resources.EnvoyPatchPolicies, xdsIR)

	// Process all Addresses for all relevant Gateways.
	t.ProcessAddresses(acceptedGateways, xdsIR, infraIR)

	// process all Backends
	backends := t.ProcessBackends(resources.Backends, resources.BackendTLSPolicies)

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

	routes := make([]RouteContext, len(httpRoutes)+len(grpcRoutes)+len(tlsRoutes)+len(tcpRoutes)+len(udpRoutes))
	offset := 0
	for i := range httpRoutes {
		routes[offset+i] = httpRoutes[i]
	}
	offset += len(httpRoutes)
	for i := range grpcRoutes {
		routes[offset+i] = grpcRoutes[i]
	}
	offset += len(grpcRoutes)
	for i := range tlsRoutes {
		routes[offset+i] = tlsRoutes[i]
	}
	offset += len(tlsRoutes)
	for i := range tcpRoutes {
		routes[offset+i] = tcpRoutes[i]
	}
	offset += len(tcpRoutes)
	for i := range udpRoutes {
		routes[offset+i] = udpRoutes[i]
	}

	// Process BackendTrafficPolicies
	backendTrafficPolicies := t.ProcessBackendTrafficPolicies(resources, acceptedGateways, routes, xdsIR)

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
	if err := t.ProcessGlobalResources(resources, xdsIR, acceptedGateways); err != nil {
		errs = errors.Join(errs, err)
	}

	// Update status of Backend TLS Policies after translating all resources
	t.ProcessBackendTLSPolicyStatus(resources.BackendTLSPolicies)

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

	return newTranslateResult(resources.GatewayClass,
		allGateways, httpRoutes, grpcRoutes, tlsRoutes,
		tcpRoutes, udpRoutes, clientTrafficPolicies, backendTrafficPolicies,
		securityPolicies, resources.BackendTLSPolicies, envoyExtensionPolicies,
		extServerPolicies, backends, resources.ListenerSets, xdsIR, infraIR), errs
}

// GetRelevantGateways returns GatewayContexts, containing a copy of the original
// Gateway with the Listener statuses reset.
func (t *Translator) GetRelevantGateways(resources *resource.Resources) (
	acceptedGateways, failedGateways []*GatewayContext,
) {
	envoyproxyMap := make(map[types.NamespacedName]*egv1a1.EnvoyProxy, len(resources.EnvoyProxiesForGateways)+1)
	envoyproxyValidationErrorMap := make(map[types.NamespacedName]error, len(resources.EnvoyProxiesForGateways))

	for _, ep := range resources.EnvoyProxiesForGateways {
		key := utils.NamespacedName(ep)
		envoyproxyMap[key] = ep
		if err := validateEnvoyProxy(ep); err != nil {
			envoyproxyValidationErrorMap[key] = err
		}
	}

	// if EnvoyProxy not found, provider layer set GC status to not accepted.
	// if EnvoyProxy found but invalid, set GC status to not accepted,
	// otherwise set GC status to accepted.
	if ep := resources.EnvoyProxyForGatewayClass; ep != nil {
		err := validateEnvoyProxy(ep)
		if err != nil {
			envoyproxyValidationErrorMap[utils.NamespacedName(ep)] = err
			t.Logger.Error(err, "Skipping GatewayClass because EnvoyProxy is invalid",
				"gatewayclass", t.GatewayClassName,
				"envoyproxy", ep.Name, "namespace", ep.Namespace)
			status.SetGatewayClassAccepted(resources.GatewayClass,
				false, string(gwapiv1.GatewayClassReasonInvalidParameters),
				fmt.Sprintf("%s: %v", status.MsgGatewayClassInvalidParams, err))
		} else {
			// TODO: remove this nil check after we update all the testdata.
			if resources.GatewayClass != nil {
				status.SetGatewayClassAccepted(
					resources.GatewayClass,
					true,
					string(gwapiv1.GatewayClassReasonAccepted),
					status.MsgValidGatewayClass)
			}

			key := utils.NamespacedName(ep)
			envoyproxyMap[key] = ep
			// we didn't append to envoyproxyValidatioErrorMap because it's valid.
		}
	}

	for _, gateway := range resources.Gateways {
		if gateway == nil {
			// Should not happen
			panic("received nil gateway")
		}

		logKeysAndValues := []any{
			"namespace", gateway.Namespace, "name", gateway.Name,
		}
		if gateway.Spec.GatewayClassName != t.GatewayClassName {
			t.Logger.Info("Skipping Gateway because GatewayClassName doesn't match", logKeysAndValues...)
			continue
		}

		gCtx := &GatewayContext{
			Gateway: gateway,
		}
		gCtx.attachEnvoyProxy(resources, envoyproxyMap)

		// Gateways that are not accepted by the controller because they reference an invalid EnvoyProxy.
		if status.GatewayNotAccepted(gCtx.Gateway) {
			failedGateways = append(failedGateways, gCtx)
			t.Logger.Info("EnvoyProxy for Gateway not found ", logKeysAndValues...)
			continue
		}

		if ep := gCtx.envoyProxy; ep != nil {
			key := utils.NamespacedName(ep)
			if err, exits := envoyproxyValidationErrorMap[key]; exits {
				failedGateways = append(failedGateways, gCtx)
				t.Logger.Info("EnvoyProxy for Gateway invalid", logKeysAndValues...)
				status.UpdateGatewayStatusNotAccepted(gCtx.Gateway, gwapiv1.GatewayReasonInvalidParameters,
					fmt.Sprintf("%s: %v", "Invalid parametersRef:", err.Error()))
				continue
			}
		}

		// we cannot do this early, otherwise there's an error when updating status.
		gCtx.ResetListeners()
		acceptedGateways = append(acceptedGateways, gCtx)
	}
	return acceptedGateways, failedGateways
}

func validateEnvoyProxy(ep *egv1a1.EnvoyProxy) error {
	if err := validation.ValidateEnvoyProxy(ep); err != nil {
		return err
	}

	if err := bootstrap.Validate(ep.Spec.Bootstrap); err != nil {
		return err
	}

	return nil
}

// InitIRs checks if mergeGateways is enabled in EnvoyProxy config and initializes XdsIR and InfraIR maps with adequate keys.
func (t *Translator) InitIRs(acceptedGateways, failedGateways []*GatewayContext) (map[string]*ir.Xds, map[string]*ir.Infra) {
	xdsIR := make(resource.XdsIRMap)
	infraIR := make(resource.InfraIRMap)

	for _, gateway := range acceptedGateways {
		irKey, gwXdsIR, gwInfraIR := t.buildIR(gateway)
		// save the IR references in the map before the translation starts
		xdsIR[irKey] = gwXdsIR
		infraIR[irKey] = gwInfraIR
	}

	// we need to build IRs for failed gateways to avoid delete event to Infra IR,
	// which will cause the deletion of infra resources managed by the failed gateways.
	for _, gtw := range failedGateways {
		irKey, gwXdsIR, gwInfraIR := t.buildIR(gtw)
		// save the IR references in the map before the translation starts
		xdsIR[irKey] = gwXdsIR
		infraIR[irKey] = gwInfraIR
	}

	return xdsIR, infraIR
}

func (t *Translator) buildIR(gateway *GatewayContext) (string, *ir.Xds, *ir.Infra) {
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

	return irKey, gwXdsIR, gwInfraIR
}

// routesToObjects collects all route types from Resources into a single []client.Object slice.
func routesToObjects(resources *resource.Resources) []client.Object {
	out := make([]client.Object, 0,
		len(resources.HTTPRoutes)+len(resources.GRPCRoutes)+
			len(resources.TLSRoutes)+len(resources.TCPRoutes)+len(resources.UDPRoutes))
	for _, r := range resources.HTTPRoutes {
		out = append(out, r)
	}
	for _, r := range resources.GRPCRoutes {
		out = append(out, r)
	}
	for _, r := range resources.TLSRoutes {
		out = append(out, r)
	}
	for _, r := range resources.TCPRoutes {
		out = append(out, r)
	}
	for _, r := range resources.UDPRoutes {
		out = append(out, r)
	}
	return out
}

// IsServiceRouting determines if Service ClusterIP routing should be used.
// It follows the priority hierarchy:
//  1. Translator.EndpointRoutingDisabled (for tests) - if true, always use Service routing
//  2. BTP RoutingType - per-route/gateway override
//  3. EnvoyProxy RoutingType - cluster-wide setting
//  4. Default: Endpoint routing
func (t *Translator) IsServiceRouting(envoyProxy *egv1a1.EnvoyProxy, btpRoutingType *egv1a1.RoutingType) bool {
	if t.EndpointRoutingDisabled {
		return true
	}

	// BTP RoutingType has priority over EnvoyProxy
	if btpRoutingType != nil {
		switch *btpRoutingType {
		case egv1a1.ServiceRoutingType:
			return true
		case egv1a1.EndpointRoutingType:
			return false
		}
	}

	// Fall back to EnvoyProxy RoutingType
	if envoyProxy != nil && envoyProxy.Spec.RoutingType != nil && *envoyProxy.Spec.RoutingType == egv1a1.ServiceRoutingType {
		return true
	}
	return false
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
	return t.IRKey(types.NamespacedName{
		Namespace: gateway.Namespace,
		Name:      gateway.Name,
	})
}

func (t *Translator) IRKey(gatewayNN types.NamespacedName) string {
	if t.MergeGateways {
		return string(t.GatewayClassName)
	}
	return irStringKey(gatewayNN.Namespace, gatewayNN.Name)
}
