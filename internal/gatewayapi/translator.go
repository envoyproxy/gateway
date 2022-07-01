package gatewayapi

import (
	"fmt"

	"github.com/envoyproxy/gateway/pkg/ir"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

const (
	KindGateway   = "Gateway"
	KindHTTPRoute = "HTTPRoute"
	KindService   = "Service"
)

// Cache is an interface that the Translator requires
// for getting Gateway API and related resources.
type Cache interface {
	ListGateways() []*v1beta1.Gateway
	ListHTTPRoutes() []*v1beta1.HTTPRoute
	GetNamespace(name string) *v1.Namespace
	GetService(namespace, name string) *v1.Service
}

// Translator translates Gateway API resources to the IR,
// and computes status for Gateway API resources.
type Translator struct {
	gatewayClassName v1beta1.ObjectName
}

type TranslateResult struct {
	Gateways   []*v1beta1.Gateway
	HTTPRoutes []*v1beta1.HTTPRoute
	IR         *ir.Xds
}

func newTranslateResult(gateways []*GatewayContext, httpRoutes []*HTTPRouteContext, xdsIR *ir.Xds) *TranslateResult {
	translateResult := &TranslateResult{
		IR: xdsIR,
	}

	for _, gateway := range gateways {
		translateResult.Gateways = append(translateResult.Gateways, gateway.Gateway)
	}
	for _, httpRoute := range httpRoutes {
		translateResult.HTTPRoutes = append(translateResult.HTTPRoutes, httpRoute.HTTPRoute)
	}

	return translateResult
}

func (t *Translator) Translate(cache Cache) *TranslateResult {
	xdsIR := &ir.Xds{}

	// Get Gateways belonging to our GatewayClass.
	gateways := t.GetRelevantGateways(cache.ListGateways())

	// Process all Listeners for all relevant Gateways.
	t.ProcessListeners(gateways, xdsIR)

	// Process all relevant HTTPRoutes.
	httpRoutes := t.ProcessHTTPRoutes(cache.ListHTTPRoutes(), gateways, cache, xdsIR)

	return newTranslateResult(gateways, httpRoutes, xdsIR)
}

func (t *Translator) GetRelevantGateways(gateways []*v1beta1.Gateway) []*GatewayContext {
	var relevant []*GatewayContext

	for _, gateway := range gateways {
		if gateway.Spec.GatewayClassName == t.gatewayClassName {
			gc := &GatewayContext{
				Gateway: gateway.DeepCopy(),
			}

			for _, listener := range gateway.Spec.Listeners {
				gc.GetListenerContext(listener.Name)
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

func (t *Translator) ProcessListeners(gateways []*GatewayContext, xdsIR *ir.Xds) {
	portListenerInfo := map[v1beta1.PortNumber]*portListeners{}

	// Iterate through all listeners and collect info about protocols
	// and hostnames per port.
	for _, gateway := range gateways {
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

	// Iterate through all listeners to validate spec
	// and compute status for each, and add valid ones
	// to the IR.
	for _, gateway := range gateways {
		for _, listener := range gateway.listeners {
			// Process protocol & supported kinds
			switch listener.Protocol {
			case v1beta1.HTTPProtocolType:
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
						}

						if kind.Kind != KindHTTPRoute {
							listener.SetCondition(
								v1beta1.ListenerConditionResolvedRefs,
								metav1.ConditionFalse,
								v1beta1.ListenerReasonInvalidRouteKinds,
								fmt.Sprintf("Kind is not supported, kind must be %s", KindHTTPRoute),
							)
						}
					}
				}
			default:
				listener.SetCondition(
					v1beta1.ListenerConditionDetached,
					metav1.ConditionTrue,
					v1beta1.ListenerReasonUnsupportedProtocol,
					"Protocol must be HTTP",
				)
			}

			// Validate allowed namespaces
			if listener.AllowedRoutes != nil && listener.AllowedRoutes.Namespaces != nil && listener.AllowedRoutes.Namespaces.From != nil && *listener.AllowedRoutes.Namespaces.From == v1beta1.NamespacesFromSelector {
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

			// Any condition on the listener indicates an error,
			// so set "Ready: false" if it's not set already.
			if len(listener.GetConditions()) > 0 {
				var hasReadyCond bool
				for _, existing := range listener.GetConditions() {
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

				continue
			}

			listener.SetCondition(v1beta1.ListenerConditionReady, metav1.ConditionTrue, v1beta1.ListenerReasonReady, "Listener is ready")

			irListener := &ir.HTTPListener{
				Name:    irListenerName(listener),
				Address: "0.0.0.0",
				Port:    uint32(listener.Port),
			}
			if listener.Hostname != nil {
				irListener.Hostnames = append(irListener.Hostnames, string(*listener.Hostname))
			}
			xdsIR.HTTP = append(xdsIR.HTTP, irListener)
		}
	}
}

func (t *Translator) ProcessHTTPRoutes(httpRoutes []*v1beta1.HTTPRoute, gateways []*GatewayContext, cache Cache, xdsIR *ir.Xds) []*HTTPRouteContext {
	var relevantHTTPRoutes []*HTTPRouteContext

	for _, h := range httpRoutes {
		httpRoute := &HTTPRouteContext{HTTPRoute: h}

		// Find out if this route attaches to one of our Gateway's listeners,
		// and if so, get the list of listeners that allow it to attach for each
		// parentRef.
		var relevantRoute bool
		for _, parentRef := range httpRoute.Spec.ParentRefs {
			isRelevantParentRef, selectedListeners := GetReferencedListeners(parentRef, gateways)

			// Parent ref is not to a Gateway that we control: skip it
			if !isRelevantParentRef {
				continue
			}
			relevantRoute = true

			parentRefCtx := httpRoute.GetRouteParentContext(parentRef)

			if !HasReadyListener(selectedListeners) {
				parentRefCtx.SetCondition(v1beta1.RouteConditionAccepted, metav1.ConditionFalse, "NoReadyListeners", "There are no ready listeners for this parent ref")
				continue
			}

			var allowedListeners []*ListenerContext
			for _, listener := range selectedListeners {
				if listener.AllowsKind(v1beta1.RouteGroupKind{Group: GroupPtr(v1beta1.GroupName), Kind: KindHTTPRoute}) && listener.AllowsNamespace(cache.GetNamespace(httpRoute.Namespace)) {
					allowedListeners = append(allowedListeners, listener)
				}
			}

			if len(allowedListeners) == 0 {
				parentRefCtx.SetCondition(v1beta1.RouteConditionAccepted, metav1.ConditionFalse, v1beta1.RouteReasonNotAllowedByListeners, "No listeners included by this parent ref allowed this attachment.")
				continue
			}

			parentRefCtx.SetListeners(allowedListeners...)

			parentRefCtx.SetCondition(v1beta1.RouteConditionAccepted, metav1.ConditionTrue, v1beta1.RouteReasonAccepted, "Route is accepted")
		}

		if !relevantRoute {
			continue
		}

		relevantHTTPRoutes = append(relevantHTTPRoutes, httpRoute)

		for _, parentRef := range httpRoute.parentRefs {
			// Skip parent refs that did not accept the route
			if !parentRef.IsAccepted() {
				continue
			}

			// Need to compute Route rules within the parentRef loop because
			// any conditions that come out of it have to go on each RouteParentStatus,
			// not on the Route as a whole.
			var routeRoutes []*ir.HTTPRoute

			// compute matches, filters, backends
			for _, rule := range httpRoute.Spec.Rules {
				var ruleRoutes []*ir.HTTPRoute

				// A rule is matched if any one of its matches
				// is satisfied (i.e. a logical "OR"), so generate
				// a unique IR HTTPRoute per match.
				for _, match := range rule.Matches {
					irRoute := &ir.HTTPRoute{}

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

					ruleRoutes = append(ruleRoutes, irRoute)
				}

				// TODO implement core filters (header modifier, redirect)

				for _, backendRef := range rule.BackendRefs {
					if backendRef.Group != nil && *backendRef.Group != "" {
						parentRef.SetCondition(
							v1beta1.RouteConditionResolvedRefs,
							metav1.ConditionFalse,
							v1beta1.RouteReasonInvalidKind,
							"Group is invalid, only the core API group (specified by omitting the group field or setting it to an empty string) is supported",
						)
						continue
					}

					if backendRef.Kind != nil && *backendRef.Kind != KindService {
						parentRef.SetCondition(
							v1beta1.RouteConditionResolvedRefs,
							metav1.ConditionFalse,
							v1beta1.RouteReasonInvalidKind,
							"Kind is invalid, only Service is supported",
						)
						continue
					}

					if backendRef.Namespace != nil && string(*backendRef.Namespace) != httpRoute.Namespace {
						// TODO implement ReferenceGrant
						parentRef.SetCondition(
							v1beta1.RouteConditionResolvedRefs,
							metav1.ConditionFalse,
							v1beta1.RouteReasonRefNotPermitted,
							"Backend must be in the same namespace as the HTTPRoute",
						)
						continue
					}

					if backendRef.Port == nil {
						parentRef.SetCondition(
							v1beta1.RouteConditionResolvedRefs,
							metav1.ConditionFalse,
							"PortNotSpecified",
							"A valid port number corresponding to a port on the Service must be specified",
						)
						continue
					}

					service := cache.GetService(NamespaceDerefOr(backendRef.Namespace, httpRoute.Namespace), string(backendRef.Name))
					if service == nil {
						parentRef.SetCondition(
							v1beta1.RouteConditionResolvedRefs,
							metav1.ConditionFalse,
							v1beta1.RouteReasonBackendNotFound,
							fmt.Sprintf("Service %s/%s not found", NamespaceDerefOr(backendRef.Namespace, httpRoute.Namespace), string(backendRef.Name)),
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
						parentRef.SetCondition(
							v1beta1.RouteConditionResolvedRefs,
							metav1.ConditionFalse,
							"PortNotFound",
							fmt.Sprintf("Port %d not found on service %s/%s", *backendRef.Port, NamespaceDerefOr(backendRef.Namespace, httpRoute.Namespace), string(backendRef.Name)),
						)
						continue
					}

					weight := uint32(1)
					if backendRef.Weight != nil {
						weight = uint32(*backendRef.Weight)
					}

					for _, route := range ruleRoutes {
						route.Destinations = append(route.Destinations, &ir.RouteDestination{
							Host:   service.Spec.ClusterIP,
							Port:   uint32(*backendRef.Port),
							Weight: weight,
						})
					}
				}

				// TODO handle:
				//	- no valid backend refs
				//	- sum of weights for valid backend refs is 0
				//	- returning 500's for invalid backend refs
				//	- etc.

				routeRoutes = append(routeRoutes, ruleRoutes...)
			}

			var hasHostnameIntersection bool
			for _, listener := range parentRef.listeners {
				hosts := ComputeHosts(httpRoute.Spec.Hostnames, listener.Hostname)
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
						perHostRoutes = append(perHostRoutes, &ir.HTTPRoute{
							PathMatch:         routeRoute.PathMatch,
							HeaderMatches:     append(headerMatches, routeRoute.HeaderMatches...),
							QueryParamMatches: routeRoute.QueryParamMatches,
							Destinations:      routeRoute.Destinations,
						})
					}
				}

				irListener := xdsIR.GetListener(irListenerName(listener))
				irListener.Routes = append(irListener.Routes, perHostRoutes...)

				// Theoretically there should only be one parent ref per
				// Route that attaches to a given Listener, so fine to just increment here, but we
				// might want to check to ensure we're not double-counting.
				if len(routeRoutes) > 0 {
					listener.IncrementAttachedRoutes()
				}
			}

			if !hasHostnameIntersection {
				parentRef.SetCondition(
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonNoMatchingListenerHostname,
					"There were no hostname intersections between the HTTPRoute and this parent ref's Listener(s).",
				)
			} else {
				parentRef.SetCondition(
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

func irListenerName(listener *ListenerContext) string {
	return fmt.Sprintf("%s-%s-%s", listener.gateway.Namespace, listener.gateway.Name, listener.Name)
}
