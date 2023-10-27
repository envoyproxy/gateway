// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	TCPProtocol = "TCP"
	UDPProtocol = "UDP"

	L4Protocol = "L4"
	L7Protocol = "L7"
)

type protocolPort struct {
	protocol gwapiv1.ProtocolType
	port     int32
}

func GroupPtr(name string) *gwapiv1.Group {
	group := gwapiv1.Group(name)
	return &group
}

func KindPtr(name string) *gwapiv1.Kind {
	kind := gwapiv1.Kind(name)
	return &kind
}

func NamespacePtr(name string) *gwapiv1.Namespace {
	namespace := gwapiv1.Namespace(name)
	return &namespace
}

func FromNamespacesPtr(fromNamespaces gwapiv1.FromNamespaces) *gwapiv1.FromNamespaces {
	return &fromNamespaces
}

func SectionNamePtr(name string) *gwapiv1.SectionName {
	sectionName := gwapiv1.SectionName(name)
	return &sectionName
}

func PortNumPtr(val int32) *gwapiv1.PortNumber {
	portNum := gwapiv1.PortNumber(val)
	return &portNum
}

func ObjectNamePtr(val string) *v1alpha2.ObjectName {
	objectName := v1alpha2.ObjectName(val)
	return &objectName
}

func PathMatchTypeDerefOr(matchType *gwapiv1.PathMatchType, defaultType gwapiv1.PathMatchType) gwapiv1.PathMatchType {
	if matchType != nil {
		return *matchType
	}
	return defaultType
}

func GRPCMethodMatchTypeDerefOr(matchType *v1alpha2.GRPCMethodMatchType, defaultType v1alpha2.GRPCMethodMatchType) v1alpha2.GRPCMethodMatchType {
	if matchType != nil {
		return *matchType
	}
	return defaultType
}

func HeaderMatchTypeDerefOr(matchType *gwapiv1.HeaderMatchType, defaultType gwapiv1.HeaderMatchType) gwapiv1.HeaderMatchType {
	if matchType != nil {
		return *matchType
	}
	return defaultType
}

func QueryParamMatchTypeDerefOr(matchType *gwapiv1.QueryParamMatchType,
	defaultType gwapiv1.QueryParamMatchType) gwapiv1.QueryParamMatchType {
	if matchType != nil {
		return *matchType
	}
	return defaultType
}

func NamespaceDerefOr(namespace *gwapiv1.Namespace, defaultNamespace string) string {
	if namespace != nil && *namespace != "" {
		return string(*namespace)
	}
	return defaultNamespace
}

func GroupDerefOr(group *gwapiv1.Group, defaultGroup string) string {
	if group != nil && *group != "" {
		return string(*group)
	}
	return defaultGroup
}

func KindDerefOr(kind *gwapiv1.Kind, defaultKind string) string {
	if kind != nil && *kind != "" {
		return string(*kind)
	}
	return defaultKind
}

// IsRefToGateway returns whether the provided parent ref is a reference
// to a Gateway with the given namespace/name, irrespective of whether a
// section/listener name has been specified (i.e. a parent ref to a listener
// on the specified gateway will return "true").
func IsRefToGateway(parentRef gwapiv1.ParentReference, gateway types.NamespacedName) bool {
	if parentRef.Group != nil && string(*parentRef.Group) != gwapiv1.GroupName {
		return false
	}

	if parentRef.Kind != nil && string(*parentRef.Kind) != KindGateway {
		return false
	}

	if parentRef.Namespace != nil && string(*parentRef.Namespace) != gateway.Namespace {
		return false
	}

	return string(parentRef.Name) == gateway.Name
}

// GetReferencedListeners returns whether a given parent ref references a Gateway
// in the given list, and if so, a list of the Listeners within that Gateway that
// are included by the parent ref (either one specific Listener, or all Listeners
// in the Gateway, depending on whether section name is specified or not).
func GetReferencedListeners(parentRef gwapiv1.ParentReference, gateways []*GatewayContext) (bool, []*ListenerContext) {
	var selectsGateway bool
	var referencedListeners []*ListenerContext

	for _, gateway := range gateways {
		if !IsRefToGateway(parentRef, types.NamespacedName{Namespace: gateway.Namespace, Name: gateway.Name}) {
			continue
		}

		selectsGateway = true

		// The parentRef may be to the entire Gateway, or to a specific listener.
		for _, listener := range gateway.listeners {
			if (parentRef.SectionName == nil || *parentRef.SectionName == listener.Name) && (parentRef.Port == nil || *parentRef.Port == listener.Port) {
				referencedListeners = append(referencedListeners, listener)
			}
		}
	}

	return selectsGateway, referencedListeners
}

// HasReadyListener returns true if at least one Listener in the
// provided list has a condition of "Ready: true", and false otherwise.
func HasReadyListener(listeners []*ListenerContext) bool {
	for _, listener := range listeners {
		if listener.IsReady() {
			return true
		}
	}
	return false
}

// ValidateHTTPRouteFilter validates the provided filter within HTTPRoute.
func ValidateHTTPRouteFilter(filter *gwapiv1.HTTPRouteFilter, extGKs ...schema.GroupKind) error {
	switch {
	case filter == nil:
		return errors.New("filter is nil")
	case filter.Type == gwapiv1.HTTPRouteFilterRequestMirror ||
		filter.Type == gwapiv1.HTTPRouteFilterURLRewrite ||
		filter.Type == gwapiv1.HTTPRouteFilterRequestRedirect ||
		filter.Type == gwapiv1.HTTPRouteFilterRequestHeaderModifier ||
		filter.Type == gwapiv1.HTTPRouteFilterResponseHeaderModifier:
		return nil
	case filter.Type == gwapiv1.HTTPRouteFilterExtensionRef:
		switch {
		case filter.ExtensionRef == nil:
			return errors.New("extensionRef field must be specified for an extended filter")
		case string(filter.ExtensionRef.Group) == egv1a1.GroupVersion.Group &&
			string(filter.ExtensionRef.Kind) == egv1a1.KindAuthenticationFilter:
			return nil
		default:
			for _, gk := range extGKs {
				if filter.ExtensionRef.Group == gwapiv1.Group(gk.Group) &&
					filter.ExtensionRef.Kind == gwapiv1.Kind(gk.Kind) {
					return nil
				}
			}
			return fmt.Errorf("unknown kind %s/%s", string(filter.ExtensionRef.Group), string(filter.ExtensionRef.Kind))
		}
	default:
		return fmt.Errorf("unsupported filter type %v", filter.Type)
	}
}

// IsAuthnHTTPFilter returns true if the provided filter is an AuthenticationFilter.
func IsAuthnHTTPFilter(filter *gwapiv1.HTTPRouteFilter) bool {
	return filter.Type == gwapiv1.HTTPRouteFilterExtensionRef &&
		filter.ExtensionRef != nil &&
		string(filter.ExtensionRef.Group) == egv1a1.GroupVersion.Group &&
		string(filter.ExtensionRef.Kind) == egv1a1.KindAuthenticationFilter
}

// ValidateGRPCRouteFilter validates the provided filter within GRPCRoute.
func ValidateGRPCRouteFilter(filter *v1alpha2.GRPCRouteFilter, extGKs ...schema.GroupKind) error {
	switch {
	case filter == nil:
		return errors.New("filter is nil")
	case filter.Type == v1alpha2.GRPCRouteFilterRequestMirror ||
		filter.Type == v1alpha2.GRPCRouteFilterRequestHeaderModifier ||
		filter.Type == v1alpha2.GRPCRouteFilterResponseHeaderModifier:
		return nil
	case filter.Type == v1alpha2.GRPCRouteFilterExtensionRef:
		switch {
		case filter.ExtensionRef == nil:
			return errors.New("extensionRef field must be specified for an extended filter")
		case string(filter.ExtensionRef.Group) == egv1a1.GroupVersion.Group &&
			string(filter.ExtensionRef.Kind) == egv1a1.KindAuthenticationFilter:
			return nil
		default:
			for _, gk := range extGKs {
				if filter.ExtensionRef.Group == gwapiv1.Group(gk.Group) &&
					filter.ExtensionRef.Kind == gwapiv1.Kind(gk.Kind) {
					return nil
				}
			}
			return fmt.Errorf("unknown kind %s/%s", string(filter.ExtensionRef.Group), string(filter.ExtensionRef.Kind))
		}
	default:
		return fmt.Errorf("unsupported filter type %v", filter.Type)
	}
}

// IsAuthnGRPCFilter returns true if the provided filter is an AuthenticationFilter.
func IsAuthnGRPCFilter(filter *v1alpha2.GRPCRouteFilter) bool {
	return filter.Type == v1alpha2.GRPCRouteFilterExtensionRef &&
		filter.ExtensionRef != nil &&
		string(filter.ExtensionRef.Group) == egv1a1.GroupVersion.Group &&
		string(filter.ExtensionRef.Kind) == egv1a1.KindAuthenticationFilter
}

// GatewayOwnerLabels returns the Gateway Owner labels using
// the provided namespace and name as the values.
func GatewayOwnerLabels(namespace, name string) map[string]string {
	return map[string]string{
		OwningGatewayNamespaceLabel: namespace,
		OwningGatewayNameLabel:      name,
	}
}

// GatewayClassOwnerLabel returns the GatewayCLass Owner label using
// the provided name as the value.
func GatewayClassOwnerLabel(name string) map[string]string {
	return map[string]string{OwningGatewayClassLabel: name}
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

// computeHosts returns a list of the intersecting hostnames between the route
// and the listener.
func computeHosts(routeHostnames []string, listenerHostname *gwapiv1.Hostname) []string {
	var listenerHostnameVal string
	if listenerHostname != nil {
		listenerHostnameVal = string(*listenerHostname)
	}

	// No route hostnames specified: use the listener hostname if specified,
	// or else match all hostnames.
	if len(routeHostnames) == 0 {
		if len(listenerHostnameVal) > 0 {
			return []string{listenerHostnameVal}
		}

		return []string{"*"}
	}

	var hostnames []string

	for i := range routeHostnames {
		routeHostname := routeHostnames[i]

		// TODO ensure routeHostname is a valid hostname

		switch {
		// No listener hostname: use the route hostname.
		case len(listenerHostnameVal) == 0:
			hostnames = append(hostnames, routeHostname)

		// Listener hostname matches the route hostname: use it.
		case listenerHostnameVal == routeHostname:
			hostnames = append(hostnames, routeHostname)

		// Listener has a wildcard hostname: check if the route hostname matches.
		case strings.HasPrefix(listenerHostnameVal, "*"):
			if hostnameMatchesWildcardHostname(routeHostname, listenerHostnameVal) {
				hostnames = append(hostnames, routeHostname)
			}

		// Route has a wildcard hostname: check if the listener hostname matches.
		case strings.HasPrefix(routeHostname, "*"):
			if hostnameMatchesWildcardHostname(listenerHostnameVal, routeHostname) {
				hostnames = append(hostnames, listenerHostnameVal)
			}

		}
	}

	return hostnames
}

// hostnameMatchesWildcardHostname returns true if hostname has the non-wildcard
// portion of wildcardHostname as a suffix, plus at least one DNS label matching the
// wildcard.
func hostnameMatchesWildcardHostname(hostname, wildcardHostname string) bool {
	if !strings.HasSuffix(hostname, strings.TrimPrefix(wildcardHostname, "*")) {
		return false
	}

	wildcardMatch := strings.TrimSuffix(hostname, strings.TrimPrefix(wildcardHostname, "*"))
	return len(wildcardMatch) > 0
}

func containsPort(ports []*protocolPort, port *protocolPort) bool {
	for _, protocolPort := range ports {
		curProtocol, curLevel := layer4Protocol(protocolPort)
		myProtocol, myLevel := layer4Protocol(port)
		if protocolPort.port == port.port && (curProtocol == myProtocol && curLevel == myLevel) {
			return true
		}
	}
	return false
}

// layer4Protocol returns listener L4 protocol and listen protocol level
func layer4Protocol(protocolPort *protocolPort) (string, string) {
	switch protocolPort.protocol {
	case gwapiv1.HTTPProtocolType, gwapiv1.HTTPSProtocolType, gwapiv1.TLSProtocolType:
		return TCPProtocol, L7Protocol
	case gwapiv1.TCPProtocolType:
		return TCPProtocol, L4Protocol
	default:
		return UDPProtocol, L4Protocol
	}
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

func irStringKey(gatewayNs, gatewayName string) string {
	return fmt.Sprintf("%s/%s", gatewayNs, gatewayName)
}

func irHTTPListenerName(listener *ListenerContext) string {
	return fmt.Sprintf("%s/%s/%s", listener.gateway.Namespace, listener.gateway.Name, listener.Name)
}

func irTLSListenerName(listener *ListenerContext, tlsRoute *TLSRouteContext) string {
	return fmt.Sprintf("%s/%s/%s/%s", listener.gateway.Namespace, listener.gateway.Name, listener.Name, tlsRoute.Name)
}

func irTCPListenerName(listener *ListenerContext, tcpRoute *TCPRouteContext) string {
	return fmt.Sprintf("%s/%s/%s/%s", listener.gateway.Namespace, listener.gateway.Name, listener.Name, tcpRoute.Name)
}

func irUDPListenerName(listener *ListenerContext, udpRoute *UDPRouteContext) string {
	return fmt.Sprintf("%s/%s/%s/%s", listener.gateway.Namespace, listener.gateway.Name, listener.Name, udpRoute.Name)
}

func irRoutePrefix(route RouteContext) string {
	return fmt.Sprintf("%s/%s/%s", strings.ToLower(string(GetRouteType(route))), route.GetNamespace(), route.GetName())
}

func irRouteName(route RouteContext, ruleIdx, matchIdx int) string {
	return fmt.Sprintf("%s/rule/%d/match/%d", irRoutePrefix(route), ruleIdx, matchIdx)
}

func irRouteDestinationName(route RouteContext, ruleIdx int) string {
	return fmt.Sprintf("%s/rule/%d", irRoutePrefix(route), ruleIdx)
}

func irTLSConfigs(tlsSecrets []*v1.Secret) []*ir.TLSListenerConfig {
	if len(tlsSecrets) == 0 {
		return nil
	}

	tlsListenerConfigs := make([]*ir.TLSListenerConfig, len(tlsSecrets))
	for i, tlsSecret := range tlsSecrets {
		tlsListenerConfigs[i] = &ir.TLSListenerConfig{
			Name:              irTLSListenerConfigName(tlsSecret),
			ServerCertificate: tlsSecret.Data[v1.TLSCertKey],
			PrivateKey:        tlsSecret.Data[v1.TLSPrivateKeyKey],
		}
	}
	return tlsListenerConfigs
}

func irTLSListenerConfigName(secret *v1.Secret) string {
	return fmt.Sprintf("%s-%s", secret.Namespace, secret.Name)
}

func isMergeGatewaysEnabled(resources *Resources) bool {
	return resources.EnvoyProxy != nil && resources.EnvoyProxy.Spec.MergeGateways != nil && *resources.EnvoyProxy.Spec.MergeGateways
}

func protocolSliceToStringSlice(protocols []gwapiv1.ProtocolType) []string {
	var protocolStrings []string
	for _, protocol := range protocols {
		protocolStrings = append(protocolStrings, string(protocol))
	}
	return protocolStrings
}
