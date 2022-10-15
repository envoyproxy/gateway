package gatewayapi

import (
	"fmt"
	"net/netip"
	"strings"

	"golang.org/x/exp/slices"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	KindGateway   = "Gateway"
	KindHTTPRoute = "HTTPRoute"
	KindTLSRoute  = "TLSRoute"
	KindService   = "Service"
	KindSecret    = "Secret"

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

type XdsIRMap map[string]*ir.Xds
type InfraIRMap map[string]*ir.Infra

// Resources holds the Gateway API and related
// resources that the translators needs as inputs.
type Resources struct {
	Gateways        []*v1beta1.Gateway
	HTTPRoutes      []*v1beta1.HTTPRoute
	TLSRoutes       []*v1alpha2.TLSRoute
	ReferenceGrants []*v1alpha2.ReferenceGrant
	Namespaces      []*v1.Namespace
	Services        []*v1.Service
	Secrets         []*v1.Secret
}

func (r *Resources) GetNamespace(name string) *v1.Namespace {
	for _, ns := range r.Namespaces {
		if ns.Name == name {
			return ns
		}
	}

	return nil
}

func (r *Resources) GetService(namespace, name string) *v1.Service {
	for _, svc := range r.Services {
		if svc.Namespace == namespace && svc.Name == name {
			return svc
		}
	}

	return nil
}

func (r *Resources) GetSecret(namespace, name string) *v1.Secret {
	for _, secret := range r.Secrets {
		if secret.Namespace == namespace && secret.Name == name {
			return secret
		}
	}

	return nil
}

// Translator translates Gateway API resources to IRs and computes status
// for Gateway API resources.
type Translator struct {
	GatewayClassName v1beta1.ObjectName
}

type TranslateResult struct {
	Gateways   []*v1beta1.Gateway
	HTTPRoutes []*v1beta1.HTTPRoute
	TLSRoutes  []*v1alpha2.TLSRoute
	XdsIR      XdsIRMap
	InfraIR    InfraIRMap
}

func newTranslateResult(gateways []*GatewayContext,
	httpRoutes []*HTTPRouteContext, tlsRoutes []*TLSRouteContext,
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
	for _, tlsRoute := range tlsRoutes {
		translateResult.TLSRoutes = append(translateResult.TLSRoutes, tlsRoute.TLSRoute)
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

	// Process all relevant TLSRoutes.
	tlsRoutes := t.ProcessTLSRoutes(resources.TLSRoutes, gateways, resources, xdsIR)

	// Sort xdsIR based on the Gateway API spec
	sortXdsIRMap(xdsIR)

	return newTranslateResult(gateways, httpRoutes, tlsRoutes, xdsIR, infraIR)
}

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

			for _, listener := range gateway.Spec.Listeners {
				l := gc.GetListenerContext(listener.Name)
				// Reset conditions and attached route count
				// since it will be recomputed during translation.
				l.ResetConditions()
				l.ResetAttachedRoutes()
			}

			relevant = append(relevant, gc)
		}
	}

	return relevant
}

type portListeners struct {
	listeners []*ListenerContext
	protocols sets.String
	hostnames map[string]int
}

func (t *Translator) ProcessListeners(gateways []*GatewayContext, xdsIR XdsIRMap, infraIR InfraIRMap, resources *Resources) {

	// Iterate through all listeners and collect info about protocols
	// and hostnames per port.
	for _, gateway := range gateways {
		portListenerInfo := map[v1beta1.PortNumber]*portListeners{}
		for _, listener := range gateway.listeners {
			if portListenerInfo[listener.Port] == nil {
				portListenerInfo[listener.Port] = &portListeners{
					protocols: sets.NewString(),
					hostnames: map[string]int{},
				}
			}

			portListenerInfo[listener.Port].listeners = append(portListenerInfo[listener.Port].listeners, listener)

			var protocol string
			switch listener.Protocol {
			// HTTPS and TLS can co-exist on the same port
			case v1beta1.HTTPSProtocolType, v1beta1.TLSProtocolType:
				protocol = "https/tls"
			default:
				protocol = string(listener.Protocol)
			}
			portListenerInfo[listener.Port].protocols.Insert(protocol)

			var hostname string
			if listener.Hostname != nil {
				hostname = string(*listener.Hostname)
			}

			portListenerInfo[listener.Port].hostnames[hostname]++
		}

		// Set Conflicted conditions for any listeners with conflicting specs.
		for _, info := range portListenerInfo {
			for _, listener := range info.listeners {
				if len(info.protocols) > 1 {
					listener.SetCondition(
						v1beta1.ListenerConditionConflicted,
						metav1.ConditionTrue,
						v1beta1.ListenerReasonProtocolConflict,
						"All listeners for a given port must use a compatible protocol",
					)
				}

				var hostname string
				if listener.Hostname != nil {
					hostname = string(*listener.Hostname)
				}

				if info.hostnames[hostname] > 1 {
					listener.SetCondition(
						v1beta1.ListenerConditionConflicted,
						metav1.ConditionTrue,
						v1beta1.ListenerReasonHostnameConflict,
						"All listeners for a given port must use a unique hostname",
					)
				}
			}
		}
	}

	// Iterate through all listeners to validate spec
	// and compute status for each, and add valid ones
	// to the Xds IR.
	for _, gateway := range gateways {
		// init IR per gateway
		irKey := irStringKey(gateway.Gateway)
		gwXdsIR := &ir.Xds{}
		gwInfraIR := ir.NewInfra()
		gwInfraIR.Proxy.Name = irKey
		gwInfraIR.Proxy.GetProxyMetadata().Labels = GatewayOwnerLabels(gateway.Namespace, gateway.Name)
		// save the IR references in the map before the translation starts
		xdsIR[irKey] = gwXdsIR
		infraIR[irKey] = gwInfraIR

		// Infra IR proxy ports must be unique.
		var foundPorts []int32

		for _, listener := range gateway.listeners {
			// Process protocol & supported kinds
			switch listener.Protocol {
			case v1beta1.TLSProtocolType:
				if listener.AllowedRoutes == nil || len(listener.AllowedRoutes.Kinds) == 0 {
					listener.SetSupportedKinds(v1beta1.RouteGroupKind{Group: GroupPtr(v1beta1.GroupName), Kind: KindTLSRoute})
				} else {
					for _, kind := range listener.AllowedRoutes.Kinds {
						if kind.Group != nil && string(*kind.Group) != v1beta1.GroupName {
							listener.SetCondition(
								v1beta1.ListenerConditionResolvedRefs,
								metav1.ConditionFalse,
								v1beta1.ListenerReasonInvalidRouteKinds,
								fmt.Sprintf("Group is not supported, group must be %s", v1beta1.GroupName),
							)
							continue
						}

						if kind.Kind != KindTLSRoute {
							listener.SetCondition(
								v1beta1.ListenerConditionResolvedRefs,
								metav1.ConditionFalse,
								v1beta1.ListenerReasonInvalidRouteKinds,
								fmt.Sprintf("Kind is not supported, kind must be %s", KindTLSRoute),
							)
							continue
						}
						listener.SetSupportedKinds(kind)
					}
				}
			case v1beta1.HTTPProtocolType, v1beta1.HTTPSProtocolType:
				if listener.AllowedRoutes == nil || len(listener.AllowedRoutes.Kinds) == 0 {
					listener.SetSupportedKinds(v1beta1.RouteGroupKind{Group: GroupPtr(v1beta1.GroupName), Kind: KindHTTPRoute})
				} else {
					for _, kind := range listener.AllowedRoutes.Kinds {
						if kind.Group != nil && string(*kind.Group) != v1beta1.GroupName {
							listener.SetCondition(
								v1beta1.ListenerConditionResolvedRefs,
								metav1.ConditionFalse,
								v1beta1.ListenerReasonInvalidRouteKinds,
								fmt.Sprintf("Group is not supported, group must be %s", v1beta1.GroupName),
							)
							continue
						}

						if kind.Kind != KindHTTPRoute {
							listener.SetCondition(
								v1beta1.ListenerConditionResolvedRefs,
								metav1.ConditionFalse,
								v1beta1.ListenerReasonInvalidRouteKinds,
								fmt.Sprintf("Kind is not supported, kind must be %s", KindHTTPRoute),
							)
							continue
						}
						listener.SetSupportedKinds(kind)
					}
				}
			default:
				listener.SetCondition(
					v1beta1.ListenerConditionDetached,
					metav1.ConditionTrue,
					v1beta1.ListenerReasonUnsupportedProtocol,
					fmt.Sprintf("Protocol %s is unsupported, must be %s or %s.", listener.Protocol, v1beta1.HTTPProtocolType, v1beta1.HTTPSProtocolType),
				)
			}

			// Validate allowed namespaces
			if listener.AllowedRoutes != nil &&
				listener.AllowedRoutes.Namespaces != nil &&
				listener.AllowedRoutes.Namespaces.From != nil &&
				*listener.AllowedRoutes.Namespaces.From == v1beta1.NamespacesFromSelector {
				if listener.AllowedRoutes.Namespaces.Selector == nil {
					listener.SetCondition(
						v1beta1.ListenerConditionReady,
						metav1.ConditionFalse,
						v1beta1.ListenerReasonInvalid,
						"The allowedRoutes.namespaces.selector field must be specified when allowedRoutes.namespaces.from is set to \"Selector\".",
					)
				} else {
					selector, err := metav1.LabelSelectorAsSelector(listener.AllowedRoutes.Namespaces.Selector)
					if err != nil {
						listener.SetCondition(
							v1beta1.ListenerConditionReady,
							metav1.ConditionFalse,
							v1beta1.ListenerReasonInvalid,
							fmt.Sprintf("The allowedRoutes.namespaces.selector could not be parsed: %v.", err),
						)
					}

					listener.namespaceSelector = selector
				}
			}

			// Process TLS configuration
			switch listener.Protocol {
			case v1beta1.HTTPProtocolType:
				if listener.TLS != nil {
					listener.SetCondition(
						v1beta1.ListenerConditionReady,
						metav1.ConditionFalse,
						v1beta1.ListenerReasonInvalid,
						fmt.Sprintf("Listener must not have TLS set when protocol is %s.", listener.Protocol),
					)
				}
			case v1beta1.HTTPSProtocolType:
				if listener.TLS == nil {
					listener.SetCondition(
						v1beta1.ListenerConditionReady,
						metav1.ConditionFalse,
						v1beta1.ListenerReasonInvalid,
						fmt.Sprintf("Listener must have TLS set when protocol is %s.", listener.Protocol),
					)
					break
				}

				if listener.TLS.Mode != nil && *listener.TLS.Mode != v1beta1.TLSModeTerminate {
					listener.SetCondition(
						v1beta1.ListenerConditionReady,
						metav1.ConditionFalse,
						"UnsupportedTLSMode",
						fmt.Sprintf("TLS %s mode is not supported, TLS mode must be Terminate.", *listener.TLS.Mode),
					)
					break
				}

				if len(listener.TLS.CertificateRefs) != 1 {
					listener.SetCondition(
						v1beta1.ListenerConditionReady,
						metav1.ConditionFalse,
						v1beta1.ListenerReasonInvalid,
						"Listener must have exactly 1 TLS certificate ref",
					)
					break
				}

				certificateRef := listener.TLS.CertificateRefs[0]

				if certificateRef.Group != nil && string(*certificateRef.Group) != "" {
					listener.SetCondition(
						v1beta1.ListenerConditionResolvedRefs,
						metav1.ConditionFalse,
						v1beta1.ListenerReasonInvalidCertificateRef,
						"Listener's TLS certificate ref group must be unspecified/empty.",
					)
					break
				}

				if certificateRef.Kind != nil && string(*certificateRef.Kind) != KindSecret {
					listener.SetCondition(
						v1beta1.ListenerConditionResolvedRefs,
						metav1.ConditionFalse,
						v1beta1.ListenerReasonInvalidCertificateRef,
						fmt.Sprintf("Listener's TLS certificate ref kind must be %s.", KindSecret),
					)
					break
				}

				secretNamespace := listener.gateway.Namespace

				if certificateRef.Namespace != nil && string(*certificateRef.Namespace) != "" && string(*certificateRef.Namespace) != listener.gateway.Namespace {
					if !isValidCrossNamespaceRef(
						crossNamespaceFrom{
							group:     string(v1beta1.GroupName),
							kind:      KindGateway,
							namespace: listener.gateway.Namespace,
						},
						crossNamespaceTo{
							group:     "",
							kind:      KindSecret,
							namespace: string(*certificateRef.Namespace),
							name:      string(certificateRef.Name),
						},
						resources.ReferenceGrants,
					) {
						listener.SetCondition(
							v1beta1.ListenerConditionResolvedRefs,
							metav1.ConditionFalse,
							v1beta1.ListenerReasonInvalidCertificateRef,
							fmt.Sprintf("Certificate ref to secret %s/%s not permitted by any ReferenceGrant", *certificateRef.Namespace, certificateRef.Name),
						)
						break
					}

					secretNamespace = string(*certificateRef.Namespace)
				}

				secret := resources.GetSecret(secretNamespace, string(certificateRef.Name))

				if secret == nil {
					listener.SetCondition(
						v1beta1.ListenerConditionResolvedRefs,
						metav1.ConditionFalse,
						v1beta1.ListenerReasonInvalidCertificateRef,
						fmt.Sprintf("Secret %s/%s does not exist.", listener.gateway.Namespace, certificateRef.Name),
					)
					break
				}

				if secret.Type != v1.SecretTypeTLS {
					listener.SetCondition(
						v1beta1.ListenerConditionResolvedRefs,
						metav1.ConditionFalse,
						v1beta1.ListenerReasonInvalidCertificateRef,
						fmt.Sprintf("Secret %s/%s must be of type %s.", listener.gateway.Namespace, certificateRef.Name, v1.SecretTypeTLS),
					)
					break
				}

				if len(secret.Data[v1.TLSCertKey]) == 0 || len(secret.Data[v1.TLSPrivateKeyKey]) == 0 {
					listener.SetCondition(
						v1beta1.ListenerConditionResolvedRefs,
						metav1.ConditionFalse,
						v1beta1.ListenerReasonInvalidCertificateRef,
						fmt.Sprintf("Secret %s/%s must contain %s and %s.", listener.gateway.Namespace, certificateRef.Name, v1.TLSCertKey, v1.TLSPrivateKeyKey),
					)
					break
				}

				listener.SetTLSSecret(secret)
			case v1beta1.TLSProtocolType:
				if listener.TLS == nil {
					listener.SetCondition(
						v1beta1.ListenerConditionReady,
						metav1.ConditionFalse,
						v1beta1.ListenerReasonInvalid,
						fmt.Sprintf("Listener must have TLS set when protocol is %s.", listener.Protocol),
					)
					break
				}

				if listener.TLS.Mode != nil && *listener.TLS.Mode != v1beta1.TLSModePassthrough {
					listener.SetCondition(
						v1beta1.ListenerConditionReady,
						metav1.ConditionFalse,
						"UnsupportedTLSMode",
						fmt.Sprintf("TLS %s mode is not supported, TLS mode must be Passthrough.", *listener.TLS.Mode),
					)
					break
				}

				// With TLS Passthrough, partial wildcards are not allowed in xDS config, so "*", "*w.abc.com" are
				// invalid configurations.
				// TODO: add regex match to detect partial wildcards like *w.abc.com
				if listener.Hostname == nil || *listener.Hostname == "" {
					listener.SetCondition(
						v1beta1.ListenerConditionReady,
						metav1.ConditionFalse,
						v1beta1.ListenerReasonInvalid,
						"Hostname must not be empty with TLS mode Passthrough.",
					)
					break
				}

				if len(listener.TLS.CertificateRefs) > 0 {
					listener.SetCondition(
						v1beta1.ListenerConditionReady,
						metav1.ConditionFalse,
						v1beta1.ListenerReasonInvalid,
						"Listener must not have TLS certificate refs set for TLS mode Passthrough",
					)
					break
				}
			}

			lConditions := listener.GetConditions()
			if len(lConditions) == 0 {
				listener.SetCondition(v1beta1.ListenerConditionReady, metav1.ConditionTrue, v1beta1.ListenerReasonReady, "Listener is ready")
				// Any condition on the listener apart from Ready=true indicates an error.
			} else if !(lConditions[0].Type == string(v1beta1.ListenerConditionReady) && lConditions[0].Status == metav1.ConditionTrue) {
				// set "Ready: false" if it's not set already.
				var hasReadyCond bool
				for _, existing := range lConditions {
					if existing.Type == string(v1beta1.ListenerConditionReady) {
						hasReadyCond = true
						break
					}
				}
				if !hasReadyCond {
					listener.SetCondition(
						v1beta1.ListenerConditionReady,
						metav1.ConditionFalse,
						v1beta1.ListenerReasonInvalid,
						"Listener is invalid, see other Conditions for details.",
					)
				}
				// skip computing IR
				continue
			}

			// Add the listener to the Xds IR.
			servicePort := int32(listener.Port)
			containerPort := servicePortToContainerPort(servicePort)
			switch listener.Protocol {
			case v1beta1.HTTPProtocolType, v1beta1.HTTPSProtocolType:
				irListener := &ir.HTTPListener{
					Name:    irListenerName(listener),
					Address: "0.0.0.0",
					Port:    uint32(containerPort),
					TLS:     irTLSConfig(listener.tlsSecret),
				}
				if listener.Hostname != nil {
					irListener.Hostnames = append(irListener.Hostnames, string(*listener.Hostname))
				} else {
					// Hostname specifies the virtual hostname to match for protocol types that define this concept.
					// When unspecified, all hostnames are matched. This field is ignored for protocols that don’t require hostname based matching.
					// see more https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.Listener.
					irListener.Hostnames = append(irListener.Hostnames, "*")
				}
				gwXdsIR.HTTP = append(gwXdsIR.HTTP, irListener)
			case v1beta1.TLSProtocolType:
				irListener := &ir.TCPListener{
					Name:    irListenerName(listener),
					Address: "0.0.0.0",
					Port:    uint32(containerPort),
					TLS: &ir.TLSInspectorConfig{
						SNIs: []string{},
					},
				}
				if listener.Hostname == nil || *listener.Hostname == "" {
					listener.SetCondition(
						v1beta1.ListenerConditionReady,
						metav1.ConditionFalse,
						v1beta1.ListenerReasonInvalid,
						"Listener is invalid, see other Conditions for details.",
					)
				}
				if listener.Hostname != nil && *listener.Hostname != "" {
					irListener.TLS.SNIs = append(irListener.TLS.SNIs, string(*listener.Hostname))
				}
				gwXdsIR.TCP = append(gwXdsIR.TCP, irListener)
			}

			// Add the listener to the Infra IR. Infra IR ports must have a unique port number.
			if !slices.Contains(foundPorts, servicePort) {
				foundPorts = append(foundPorts, servicePort)
				var proto ir.ProtocolType
				switch listener.Protocol {
				case v1beta1.HTTPProtocolType:
					proto = ir.HTTPProtocolType
				case v1beta1.HTTPSProtocolType:
					proto = ir.HTTPSProtocolType
				case v1beta1.TLSProtocolType:
					proto = ir.TLSProtocolType
				}
				infraPort := ir.ListenerPort{
					Name:          string(listener.Name),
					Protocol:      proto,
					ServicePort:   servicePort,
					ContainerPort: containerPort,
				}
				// Only 1 listener is supported.
				gwInfraIR.Proxy.Listeners[0].Ports = append(gwInfraIR.Proxy.Listeners[0].Ports, infraPort)
			}
		}
	}
}

// servicePortToContainerPort translates a service port into an ephemeral
// container port.
func servicePortToContainerPort(servicePort int32) int32 {
	// If the service port is a privileged port (1-1023)
	// add a constant to the value converting it into an ephemeral port.
	// This allows the container to bind to the port without needing a
	// CAP_NET_BIND_SERVICE capability.
	if servicePort < minEphemeralPort {
		return servicePort + wellKnownPortShift
	}
	return servicePort
}

// buildRuleRouteDest takes a backendRef and translates it into a destination or sets error statuses and
// returns the weight for the backend so that 500 error responses can be returned for invalid backends in
// the same proportion as the backend would have otherwise received
func buildRuleRouteDest(backendRef v1beta1.HTTPBackendRef,
	parentRef *RouteParentContext,
	httpRoute *HTTPRouteContext,
	resources *Resources) (destination *ir.RouteDestination, backendWeight uint32) {

	weight := uint32(1)
	if backendRef.Weight != nil {
		weight = uint32(*backendRef.Weight)
	}

	if backendRef.Group != nil && *backendRef.Group != "" {
		parentRef.SetCondition(httpRoute,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			v1beta1.RouteReasonInvalidKind,
			"Group is invalid, only the core API group (specified by omitting the group field or setting it to an empty string) is supported",
		)
		return nil, weight
	}

	if backendRef.Kind != nil && *backendRef.Kind != KindService {
		parentRef.SetCondition(httpRoute,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			v1beta1.RouteReasonInvalidKind,
			"Kind is invalid, only Service is supported",
		)
		return nil, weight
	}

	if backendRef.Namespace != nil && string(*backendRef.Namespace) != "" && string(*backendRef.Namespace) != httpRoute.Namespace {
		if !isValidCrossNamespaceRef(
			crossNamespaceFrom{
				group:     v1beta1.GroupName,
				kind:      KindHTTPRoute,
				namespace: httpRoute.Namespace,
			},
			crossNamespaceTo{
				group:     "",
				kind:      KindService,
				namespace: string(*backendRef.Namespace),
				name:      string(backendRef.Name),
			},
			resources.ReferenceGrants,
		) {
			parentRef.SetCondition(httpRoute,
				v1beta1.RouteConditionResolvedRefs,
				metav1.ConditionFalse,
				v1beta1.RouteReasonRefNotPermitted,
				fmt.Sprintf("Backend ref to service %s/%s not permitted by any ReferenceGrant", *backendRef.Namespace, backendRef.Name),
			)
			return nil, weight
		}
	}

	if backendRef.Port == nil {
		parentRef.SetCondition(httpRoute,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			"PortNotSpecified",
			"A valid port number corresponding to a port on the Service must be specified",
		)
		return nil, weight
	}

	service := resources.GetService(NamespaceDerefOr(backendRef.Namespace, httpRoute.Namespace), string(backendRef.Name))
	if service == nil {
		parentRef.SetCondition(httpRoute,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			v1beta1.RouteReasonBackendNotFound,
			fmt.Sprintf("Service %s/%s not found", NamespaceDerefOr(backendRef.Namespace, httpRoute.Namespace), string(backendRef.Name)),
		)
		return nil, weight
	}

	var portFound bool
	for _, port := range service.Spec.Ports {
		if port.Port == int32(*backendRef.Port) {
			portFound = true
			break
		}
	}

	if !portFound {
		parentRef.SetCondition(httpRoute,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			"PortNotFound",
			fmt.Sprintf("Port %d not found on service %s/%s", *backendRef.Port, NamespaceDerefOr(backendRef.Namespace, httpRoute.Namespace), string(backendRef.Name)),
		)
		return nil, weight
	}

	return &ir.RouteDestination{
		Host:   service.Spec.ClusterIP,
		Port:   uint32(*backendRef.Port),
		Weight: weight,
	}, weight

}

func (t *Translator) ProcessHTTPRoutes(httpRoutes []*v1beta1.HTTPRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*HTTPRouteContext {
	var relevantHTTPRoutes []*HTTPRouteContext

	for _, h := range httpRoutes {
		if h == nil {
			panic("received nil httproute")
		}
		httpRoute := &HTTPRouteContext{HTTPRoute: h}

		// Find out if this route attaches to one of our Gateway's listeners,
		// and if so, get the list of listeners that allow it to attach for each
		// parentRef.
		relevantRoute := processAllowedListenersForParentRefs(httpRoute, gateways, resources)
		if !relevantRoute {
			continue
		}

		relevantHTTPRoutes = append(relevantHTTPRoutes, httpRoute)

		for _, parentRef := range httpRoute.parentRefs {
			// Skip parent refs that did not accept the route
			if !parentRef.IsAccepted(httpRoute) {
				continue
			}

			// Need to compute Route rules within the parentRef loop because
			// any conditions that come out of it have to go on each RouteParentStatus,
			// not on the Route as a whole.
			var routeRoutes []*ir.HTTPRoute

			// compute matches, filters, backends
			for ruleIdx, rule := range httpRoute.Spec.Rules {
				var ruleRoutes []*ir.HTTPRoute

				// First see if there are any filters in the rules. Then apply those filters to any irRoutes.
				var directResponse *ir.DirectResponse
				var redirectResponse *ir.Redirect
				addRequestHeaders := []ir.AddHeader{}
				removeRequestHeaders := []string{}

				// Process the filters for this route rule
				for _, filter := range rule.Filters {
					if directResponse != nil {
						break // If an invalid filter type has been configured then skip processing any more filters
					}
					switch filter.Type {
					case v1beta1.HTTPRouteFilterRequestRedirect:
						// Can't have two redirects for the same route
						if redirectResponse != nil {
							parentRef.SetCondition(httpRoute,
								v1beta1.RouteConditionAccepted,
								metav1.ConditionFalse,
								v1beta1.RouteReasonUnsupportedValue,
								"Cannot configure multiple requestRedirect filters for a single HTTPRouteRule",
							)
							continue
						}

						redirect := filter.RequestRedirect
						if redirect == nil {
							break
						}

						redir := &ir.Redirect{}
						if redirect.Scheme != nil {
							// Note that gateway API may support additional schemes in the future, but unknown values
							// must result in an UnsupportedValue status
							if *redirect.Scheme == "http" || *redirect.Scheme == "https" {
								redir.Scheme = redirect.Scheme
							} else {
								errMsg := fmt.Sprintf("Scheme: %s is unsupported, only 'https' and 'http' are supported", *redirect.Scheme)
								parentRef.SetCondition(httpRoute,
									v1beta1.RouteConditionAccepted,
									metav1.ConditionFalse,
									v1beta1.RouteReasonUnsupportedValue,
									errMsg,
								)
								continue
							}
						}

						if redirect.Hostname != nil {
							if err := isValidHostname(string(*redirect.Hostname)); err != nil {
								parentRef.SetCondition(httpRoute,
									v1beta1.RouteConditionAccepted,
									metav1.ConditionFalse,
									v1beta1.RouteReasonUnsupportedValue,
									err.Error(),
								)
								continue
							} else {
								redirectHost := string(*redirect.Hostname)
								redir.Hostname = &redirectHost
							}
						}

						if redirect.Path != nil {
							switch redirect.Path.Type {
							case v1beta1.FullPathHTTPPathModifier:
								if redirect.Path.ReplaceFullPath != nil {
									redir.Path = &ir.HTTPPathModifier{
										FullReplace: redirect.Path.ReplaceFullPath,
									}
								}
							case v1beta1.PrefixMatchHTTPPathModifier:
								if redirect.Path.ReplacePrefixMatch != nil {
									redir.Path = &ir.HTTPPathModifier{
										PrefixMatchReplace: redirect.Path.ReplacePrefixMatch,
									}
								}
							default:
								errMsg := fmt.Sprintf("Redirect path type: %s is invalid, only \"ReplaceFullPath\" and \"ReplacePrefixMatch\" are supported", redirect.Path.Type)
								parentRef.SetCondition(httpRoute,
									v1beta1.RouteConditionAccepted,
									metav1.ConditionFalse,
									v1beta1.RouteReasonUnsupportedValue,
									errMsg,
								)
								continue
							}
						}

						if redirect.StatusCode != nil {
							redirectCode := int32(*redirect.StatusCode)
							// Envoy supports 302, 303, 307, and 308, but gateway API only includes 301 and 302
							if redirectCode == 301 || redirectCode == 302 {
								redir.StatusCode = &redirectCode
							} else {
								errMsg := fmt.Sprintf("Status code %d is invalid, only 302 and 301 are supported", redirectCode)
								parentRef.SetCondition(httpRoute,
									v1beta1.RouteConditionAccepted,
									metav1.ConditionFalse,
									v1beta1.RouteReasonUnsupportedValue,
									errMsg,
								)
								continue
							}
						}

						if redirect.Port != nil {
							redirectPort := uint32(*redirect.Port)
							redir.Port = &redirectPort
						}

						redirectResponse = redir
					case v1beta1.HTTPRouteFilterRequestHeaderModifier:
						// Make sure the header modifier config actually exists
						headerModifier := filter.RequestHeaderModifier
						if headerModifier == nil {
							break
						}
						emptyFilterConfig := true // keep track of whether the provided config is empty or not

						// Add request headers
						if headersToAdd := headerModifier.Add; headersToAdd != nil {
							if len(headersToAdd) > 0 {
								emptyFilterConfig = false
							}
							for _, addHeader := range headersToAdd {
								emptyFilterConfig = false
								if addHeader.Name == "" {
									parentRef.SetCondition(httpRoute,
										v1beta1.RouteConditionAccepted,
										metav1.ConditionFalse,
										v1beta1.RouteReasonUnsupportedValue,
										"RequestHeaderModifier Filter cannot add a header with an empty name",
									)
									// try to process the rest of the headers and produce a valid config.
									continue
								}
								// Per Gateway API specification on HTTPHeaderName, : and / are invalid characters in header names
								if strings.Contains(string(addHeader.Name), "/") || strings.Contains(string(addHeader.Name), ":") {
									parentRef.SetCondition(httpRoute,
										v1beta1.RouteConditionAccepted,
										metav1.ConditionFalse,
										v1beta1.RouteReasonUnsupportedValue,
										fmt.Sprintf("RequestHeaderModifier Filter cannot set headers with a '/' or ':' character in them. Header: %q", string(addHeader.Name)),
									)
									continue
								}
								// Check if the header is a duplicate
								headerKey := string(addHeader.Name)
								canAddHeader := true
								for _, h := range addRequestHeaders {
									if strings.EqualFold(h.Name, headerKey) {
										canAddHeader = false
										break
									}
								}

								if !canAddHeader {
									continue
								}

								newHeader := ir.AddHeader{
									Name:   headerKey,
									Append: true,
									Value:  addHeader.Value,
								}

								addRequestHeaders = append(addRequestHeaders, newHeader)
							}
						}

						// Set headers
						if headersToSet := headerModifier.Set; headersToSet != nil {
							if len(headersToSet) > 0 {
								emptyFilterConfig = false
							}
							for _, setHeader := range headersToSet {

								if setHeader.Name == "" {
									parentRef.SetCondition(httpRoute,
										v1beta1.RouteConditionAccepted,
										metav1.ConditionFalse,
										v1beta1.RouteReasonUnsupportedValue,
										"RequestHeaderModifier Filter cannot set a header with an empty name",
									)
									continue
								}
								// Per Gateway API specification on HTTPHeaderName, : and / are invalid characters in header names
								if strings.Contains(string(setHeader.Name), "/") || strings.Contains(string(setHeader.Name), ":") {
									parentRef.SetCondition(httpRoute,
										v1beta1.RouteConditionAccepted,
										metav1.ConditionFalse,
										v1beta1.RouteReasonUnsupportedValue,
										fmt.Sprintf("RequestHeaderModifier Filter cannot set headers with a '/' or ':' character in them. Header: '%s'", string(setHeader.Name)),
									)
									continue
								}

								// Check if the header to be set has already been configured
								headerKey := string(setHeader.Name)
								canAddHeader := true
								for _, h := range addRequestHeaders {
									if strings.EqualFold(h.Name, headerKey) {
										canAddHeader = false
										break
									}
								}
								if !canAddHeader {
									continue
								}
								newHeader := ir.AddHeader{
									Name:   string(setHeader.Name),
									Append: false,
									Value:  setHeader.Value,
								}

								addRequestHeaders = append(addRequestHeaders, newHeader)
							}
						}

						// Remove request headers
						// As far as Envoy is concerned, it is ok to configure a header to be added/set and also in the list of
						// headers to remove. It will remove the original header if present and then add/set the header after.
						if headersToRemove := headerModifier.Remove; headersToRemove != nil {
							if len(headersToRemove) > 0 {
								emptyFilterConfig = false
							}
							for _, removedHeader := range headersToRemove {
								if removedHeader == "" {
									parentRef.SetCondition(httpRoute,
										v1beta1.RouteConditionAccepted,
										metav1.ConditionFalse,
										v1beta1.RouteReasonUnsupportedValue,
										"RequestHeaderModifier Filter cannot remove a header with an empty name",
									)
									continue
								}

								canRemHeader := true
								for _, h := range removeRequestHeaders {
									if strings.EqualFold(h, removedHeader) {
										canRemHeader = false
										break
									}
								}
								if !canRemHeader {
									continue
								}

								removeRequestHeaders = append(removeRequestHeaders, removedHeader)

							}
						}

						// Update the status if the filter failed to configure any valid headers to add/remove
						if len(addRequestHeaders) == 0 && len(removeRequestHeaders) == 0 && !emptyFilterConfig {
							parentRef.SetCondition(httpRoute,
								v1beta1.RouteConditionAccepted,
								metav1.ConditionFalse,
								v1beta1.RouteReasonUnsupportedValue,
								"RequestHeaderModifier Filter did not provide valid configuration to add/set/remove any headers",
							)
						}
					case v1beta1.HTTPRouteFilterExtensionRef:
						// "If a reference to a custom filter type cannot be resolved, the filter MUST NOT be skipped.
						// Instead, requests that would have been processed by that filter MUST receive a HTTP error response."
						errMsg := fmt.Sprintf("Unknown custom filter type: %s", filter.Type)
						parentRef.SetCondition(httpRoute,
							v1beta1.RouteConditionAccepted,
							metav1.ConditionFalse,
							v1beta1.RouteReasonUnsupportedValue,
							errMsg,
						)
						directResponse = &ir.DirectResponse{
							Body:       &errMsg,
							StatusCode: 500,
						}
					default:
						// Unsupported filters.
						errMsg := fmt.Sprintf("Unsupported filter type: %s", filter.Type)
						parentRef.SetCondition(httpRoute,
							v1beta1.RouteConditionAccepted,
							metav1.ConditionFalse,
							v1beta1.RouteReasonUnsupportedValue,
							errMsg,
						)
						directResponse = &ir.DirectResponse{
							Body:       &errMsg,
							StatusCode: 500,
						}
					}
				}

				// A rule is matched if any one of its matches
				// is satisfied (i.e. a logical "OR"), so generate
				// a unique Xds IR HTTPRoute per match.
				for matchIdx, match := range rule.Matches {
					irRoute := &ir.HTTPRoute{
						Name: routeName(httpRoute, ruleIdx, matchIdx),
					}

					if match.Path != nil {
						switch PathMatchTypeDerefOr(match.Path.Type, v1beta1.PathMatchPathPrefix) {
						case v1beta1.PathMatchPathPrefix:
							irRoute.PathMatch = &ir.StringMatch{
								Prefix: match.Path.Value,
							}
						case v1beta1.PathMatchExact:
							irRoute.PathMatch = &ir.StringMatch{
								Exact: match.Path.Value,
							}
						}
					}
					for _, headerMatch := range match.Headers {
						if HeaderMatchTypeDerefOr(headerMatch.Type, v1beta1.HeaderMatchExact) == v1beta1.HeaderMatchExact {
							irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
								Name:  string(headerMatch.Name),
								Exact: StringPtr(headerMatch.Value),
							})
						}
					}

					// Add the redirect filter or direct response that were created earlier to all the irRoutes
					if redirectResponse != nil {
						irRoute.Redirect = redirectResponse
					}
					if directResponse != nil {
						irRoute.DirectResponse = directResponse
					}
					if len(addRequestHeaders) > 0 {
						irRoute.AddRequestHeaders = addRequestHeaders
					}
					if len(removeRequestHeaders) > 0 {
						irRoute.RemoveRequestHeaders = removeRequestHeaders
					}
					ruleRoutes = append(ruleRoutes, irRoute)
				}

				for _, backendRef := range rule.BackendRefs {
					destination, backendWeight := buildRuleRouteDest(backendRef, parentRef, httpRoute, resources)
					for _, route := range ruleRoutes {
						// If the route already has a direct response or redirect configured, then it was from a filter so skip
						// processing any destinations for this route.
						if route.DirectResponse == nil && route.Redirect == nil {
							if destination != nil {
								route.Destinations = append(route.Destinations, destination)
								route.BackendWeights.Valid += backendWeight

							} else {
								route.BackendWeights.Invalid += backendWeight
							}
						}
					}
				}

				// If the route has no valid backends then just use a direct response and don't fuss with weighted responses
				for _, ruleRoute := range ruleRoutes {
					if ruleRoute.BackendWeights.Invalid > 0 && len(ruleRoute.Destinations) == 0 {
						ruleRoute.DirectResponse = &ir.DirectResponse{
							StatusCode: 500,
						}
					}
				}

				// TODO handle:
				//	- sum of weights for valid backend refs is 0
				//	- etc.

				routeRoutes = append(routeRoutes, ruleRoutes...)
			}

			var hasHostnameIntersection bool
			for _, listener := range parentRef.listeners {
				hosts := computeHosts(httpRoute.GetHostnames(), listener.Hostname)
				if len(hosts) == 0 {
					continue
				}
				hasHostnameIntersection = true

				var perHostRoutes []*ir.HTTPRoute
				for _, host := range hosts {
					var headerMatches []*ir.StringMatch

					// If the intersecting host is more specific than the Listener's hostname,
					// add an additional header match to all of the routes for it
					if host != "*" && (listener.Hostname == nil || string(*listener.Hostname) != host) {
						headerMatches = append(headerMatches, &ir.StringMatch{
							Name:  ":authority",
							Exact: StringPtr(host),
						})
					}

					for _, routeRoute := range routeRoutes {
						hostRoute := &ir.HTTPRoute{
							Name:                 fmt.Sprintf("%s-%s", routeRoute.Name, host),
							PathMatch:            routeRoute.PathMatch,
							HeaderMatches:        append(headerMatches, routeRoute.HeaderMatches...),
							QueryParamMatches:    routeRoute.QueryParamMatches,
							AddRequestHeaders:    routeRoute.AddRequestHeaders,
							RemoveRequestHeaders: routeRoute.RemoveRequestHeaders,
							Destinations:         routeRoute.Destinations,
							Redirect:             routeRoute.Redirect,
							DirectResponse:       routeRoute.DirectResponse,
						}
						// Don't bother copying over the weights unless the route has invalid backends.
						if routeRoute.BackendWeights.Invalid > 0 {
							hostRoute.BackendWeights = routeRoute.BackendWeights
						}
						perHostRoutes = append(perHostRoutes, hostRoute)
					}
				}

				irKey := irStringKey(listener.gateway)
				irListener := xdsIR[irKey].GetHTTPListener(irListenerName(listener))
				if irListener != nil {
					irListener.Routes = append(irListener.Routes, perHostRoutes...)
				}
				// Theoretically there should only be one parent ref per
				// Route that attaches to a given Listener, so fine to just increment here, but we
				// might want to check to ensure we're not double-counting.
				if len(routeRoutes) > 0 {
					listener.IncrementAttachedRoutes()
				}
			}

			if !hasHostnameIntersection {
				parentRef.SetCondition(httpRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonNoMatchingListenerHostname,
					"There were no hostname intersections between the HTTPRoute and this parent ref's Listener(s).",
				)
			}

			// If no negative conditions have been set, the route is considered "Accepted=True".
			if parentRef.httpRoute != nil &&
				len(parentRef.httpRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
				parentRef.SetCondition(httpRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionTrue,
					v1beta1.RouteReasonAccepted,
					"Route is accepted",
				)
			}
		}
	}

	return relevantHTTPRoutes
}

func (t *Translator) ProcessTLSRoutes(tlsRoutes []*v1alpha2.TLSRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*TLSRouteContext {
	var relevantTLSRoutes []*TLSRouteContext

	for _, t := range tlsRoutes {
		if t == nil {
			panic("received nil tlsroute")
		}
		tlsRoute := &TLSRouteContext{TLSRoute: t}

		// Find out if this route attaches to one of our Gateway's listeners,
		// and if so, get the list of listeners that allow it to attach for each
		// parentRef.
		relevantRoute := processAllowedListenersForParentRefs(tlsRoute, gateways, resources)
		if !relevantRoute {
			continue
		}

		relevantTLSRoutes = append(relevantTLSRoutes, tlsRoute)

		for _, parentRef := range tlsRoute.parentRefs {
			// Skip parent refs that did not accept the route
			if !parentRef.IsAccepted(tlsRoute) {
				continue
			}

			// Need to compute Route rules within the parentRef loop because
			// any conditions that come out of it have to go on each RouteParentStatus,
			// not on the Route as a whole.
			var routeDestinations []*ir.RouteDestination

			// compute backends
			for _, rule := range tlsRoute.Spec.Rules {
				for _, backendRef := range rule.BackendRefs {
					if backendRef.Group != nil && *backendRef.Group != "" {
						parentRef.SetCondition(tlsRoute,
							v1beta1.RouteConditionResolvedRefs,
							metav1.ConditionFalse,
							v1beta1.RouteReasonInvalidKind,
							"Group is invalid, only the core API group (specified by omitting the group field or setting it to an empty string) is supported",
						)
						continue
					}

					if backendRef.Kind != nil && *backendRef.Kind != KindService {
						parentRef.SetCondition(tlsRoute,
							v1beta1.RouteConditionResolvedRefs,
							metav1.ConditionFalse,
							v1beta1.RouteReasonInvalidKind,
							"Kind is invalid, only Service is supported",
						)
						continue
					}

					if backendRef.Namespace != nil && string(*backendRef.Namespace) != "" && string(*backendRef.Namespace) != tlsRoute.Namespace {
						if !isValidCrossNamespaceRef(
							crossNamespaceFrom{
								group:     v1beta1.GroupName,
								kind:      KindTLSRoute,
								namespace: tlsRoute.Namespace,
							},
							crossNamespaceTo{
								group:     "",
								kind:      KindService,
								namespace: string(*backendRef.Namespace),
								name:      string(backendRef.Name),
							},
							resources.ReferenceGrants,
						) {
							parentRef.SetCondition(tlsRoute,
								v1beta1.RouteConditionResolvedRefs,
								metav1.ConditionFalse,
								v1beta1.RouteReasonRefNotPermitted,
								fmt.Sprintf("Backend ref to service %s/%s not permitted by any ReferenceGrant", *backendRef.Namespace, backendRef.Name),
							)
							continue
						}
					}

					if backendRef.Port == nil {
						parentRef.SetCondition(tlsRoute,
							v1beta1.RouteConditionResolvedRefs,
							metav1.ConditionFalse,
							"PortNotSpecified",
							"A valid port number corresponding to a port on the Service must be specified",
						)
						continue
					}

					// TODO: [v1alpha2-v1beta1] Replace with NamespaceDerefOr when TLSRoute graduates to v1beta1.
					serviceNamespace := NamespaceDerefOrAlpha(backendRef.Namespace, tlsRoute.Namespace)
					service := resources.GetService(serviceNamespace, string(backendRef.Name))
					if service == nil {
						parentRef.SetCondition(tlsRoute,
							v1beta1.RouteConditionResolvedRefs,
							metav1.ConditionFalse,
							v1beta1.RouteReasonBackendNotFound,
							fmt.Sprintf("Service %s/%s not found", serviceNamespace, string(backendRef.Name)),
						)
						continue
					}

					var portFound bool
					for _, port := range service.Spec.Ports {
						if port.Port == int32(*backendRef.Port) {
							portFound = true
							break
						}
					}

					if !portFound {
						parentRef.SetCondition(tlsRoute,
							v1beta1.RouteConditionResolvedRefs,
							metav1.ConditionFalse,
							"PortNotFound",
							fmt.Sprintf("Port %d not found on service %s/%s", *backendRef.Port, serviceNamespace, string(backendRef.Name)),
						)
						continue
					}

					weight := uint32(1)
					if backendRef.Weight != nil {
						weight = uint32(*backendRef.Weight)
					}

					routeDestinations = append(routeDestinations, &ir.RouteDestination{
						Host:   service.Spec.ClusterIP,
						Port:   uint32(*backendRef.Port),
						Weight: weight,
					})
				}

				// TODO handle:
				//	- no valid backend refs
				//	- sum of weights for valid backend refs is 0
				//	- returning 500's for invalid backend refs
				//	- etc.
			}

			var hasHostnameIntersection bool
			for _, listener := range parentRef.listeners {
				hosts := computeHosts(tlsRoute.GetHostnames(), listener.Hostname)
				if len(hosts) == 0 {
					continue
				}
				hasHostnameIntersection = true

				irKey := irStringKey(listener.gateway)
				irListener := xdsIR[irKey].GetTCPListener(irListenerName(listener))
				if irListener != nil {
					irListener.Destinations = routeDestinations
				}
				// Theoretically there should only be one parent ref per
				// Route that attaches to a given Listener, so fine to just increment here, but we
				// might want to check to ensure we're not double-counting.
				if len(routeDestinations) > 0 {
					listener.IncrementAttachedRoutes()
				}
			}

			if !hasHostnameIntersection {
				parentRef.SetCondition(tlsRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonNoMatchingListenerHostname,
					"There were no hostname intersections between the HTTPRoute and this parent ref's Listener(s).",
				)
			}

			// If no negative conditions have been set, the route is considered "Accepted=True".
			if parentRef.tlsRoute != nil &&
				len(parentRef.tlsRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
				parentRef.SetCondition(tlsRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionTrue,
					v1beta1.RouteReasonAccepted,
					"Route is accepted",
				)
			}
		}
	}

	return relevantTLSRoutes
}

// processAllowedListenersForParentRefs finds out if the route attaches to one of our
// Gateways' listeners, and if so, gets the list of listeners that allow it to
// attach for each parentRef.
func processAllowedListenersForParentRefs(routeContext RouteContext, gateways []*GatewayContext, resources *Resources) bool {
	var relevantRoute bool

	for _, parentRef := range routeContext.GetParentReferences() {
		isRelevantParentRef, selectedListeners := GetReferencedListeners(parentRef, gateways)

		// Parent ref is not to a Gateway that we control: skip it
		if !isRelevantParentRef {
			continue
		}
		relevantRoute = true

		parentRefCtx := routeContext.GetRouteParentContext(parentRef)
		// Reset conditions since they will be recomputed during translation
		parentRefCtx.ResetConditions(routeContext)

		if !HasReadyListener(selectedListeners) {
			parentRefCtx.SetCondition(routeContext,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				"NoReadyListeners",
				"There are no ready listeners for this parent ref",
			)
			continue
		}

		var allowedListeners []*ListenerContext
		for _, listener := range selectedListeners {
			acceptedKind := routeContext.GetRouteType()
			if listener.AllowsKind(v1beta1.RouteGroupKind{Group: GroupPtr(v1beta1.GroupName), Kind: v1beta1.Kind(acceptedKind)}) &&
				listener.AllowsNamespace(resources.GetNamespace(routeContext.GetNamespace())) {
				allowedListeners = append(allowedListeners, listener)
			}
		}

		if len(allowedListeners) == 0 {
			parentRefCtx.SetCondition(routeContext,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonNotAllowedByListeners,
				"No listeners included by this parent ref allowed this attachment.",
			)
			continue
		}

		parentRefCtx.SetListeners(allowedListeners...)

		parentRefCtx.SetCondition(routeContext,
			v1beta1.RouteConditionAccepted,
			metav1.ConditionTrue,
			v1beta1.RouteReasonAccepted,
			"Route is accepted",
		)
	}
	return relevantRoute
}

type crossNamespaceFrom struct {
	group     string
	kind      string
	namespace string
}

type crossNamespaceTo struct {
	group     string
	kind      string
	namespace string
	name      string
}

func isValidCrossNamespaceRef(from crossNamespaceFrom, to crossNamespaceTo, referenceGrants []*v1alpha2.ReferenceGrant) bool {
	for _, referenceGrant := range referenceGrants {
		// The ReferenceGrant must be defined in the namespace of
		// the "to" (the referent).
		if referenceGrant.Namespace != to.namespace {
			continue
		}

		// Check if the ReferenceGrant has a matching "from".
		var fromAllowed bool
		for _, refGrantFrom := range referenceGrant.Spec.From {
			if string(refGrantFrom.Namespace) == from.namespace && string(refGrantFrom.Group) == from.group && string(refGrantFrom.Kind) == from.kind {
				fromAllowed = true
				break
			}
		}
		if !fromAllowed {
			continue
		}

		// Check if the ReferenceGrant has a matching "to".
		var toAllowed bool
		for _, refGrantTo := range referenceGrant.Spec.To {
			if string(refGrantTo.Group) == to.group && string(refGrantTo.Kind) == to.kind && (refGrantTo.Name == nil || *refGrantTo.Name == "" || string(*refGrantTo.Name) == to.name) {
				toAllowed = true
				break
			}
		}
		if !toAllowed {
			continue
		}

		// If we got here, both the "from" and the "to" were allowed by this
		// reference grant.
		return true
	}

	// If we got here, no reference policy or reference grant allowed both the "from" and "to".
	return false
}

// Checks if a hostname is valid according to RFC 1123 and gateway API's requirement that it not be an IP address
func isValidHostname(hostname string) error {
	if errs := validation.IsDNS1123Subdomain(hostname); errs != nil {
		return fmt.Errorf("hostname %q is invalid for a redirect filter: %v", hostname, errs)
	}

	// IP addresses are not allowed so parsing the hostname as an address needs to fail
	if _, err := netip.ParseAddr(hostname); err == nil {
		return fmt.Errorf("hostname: %q cannot be an ip address", hostname)
	}

	labelIdx := 0
	for i := range hostname {
		if hostname[i] == '.' {

			if i-labelIdx > 63 {
				return fmt.Errorf("label: %q in hostname %q cannot exceed 63 characters", hostname[labelIdx:i], hostname)
			}
			labelIdx = i + 1
		}
	}
	// Check the last label
	if len(hostname)-labelIdx > 63 {
		return fmt.Errorf("label: %q in hostname %q cannot exceed 63 characters", hostname[labelIdx:], hostname)
	}

	return nil
}

func irStringKey(gateway *v1beta1.Gateway) string {
	return fmt.Sprintf("%s-%s", gateway.Namespace, gateway.Name)
}

func irListenerName(listener *ListenerContext) string {
	return fmt.Sprintf("%s-%s-%s", listener.gateway.Namespace, listener.gateway.Name, listener.Name)
}

func routeName(route RouteContext, ruleIdx, matchIdx int) string {
	return fmt.Sprintf("%s-%s-rule-%d-match-%d", route.GetNamespace(), route.GetName(), ruleIdx, matchIdx)
}

func irTLSConfig(tlsSecret *v1.Secret) *ir.TLSListenerConfig {
	if tlsSecret == nil {
		return nil
	}

	return &ir.TLSListenerConfig{
		ServerCertificate: tlsSecret.Data[v1.TLSCertKey],
		PrivateKey:        tlsSecret.Data[v1.TLSPrivateKeyKey],
	}
}

// GatewayOwnerLabels returns the Gateway Owner labels using
// the provided namespace and name as the values.
func GatewayOwnerLabels(namespace, name string) map[string]string {
	return map[string]string{
		OwningGatewayNamespaceLabel: namespace,
		OwningGatewayNameLabel:      name,
	}
}
