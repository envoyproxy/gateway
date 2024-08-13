// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
)

const (
	TCPProtocol = "TCP"
	UDPProtocol = "UDP"

	L4Protocol = "L4"
	L7Protocol = "L7"

	caCertKey = "ca.crt"
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

func ObjectNamePtr(val string) *gwapiv1a2.ObjectName {
	objectName := gwapiv1a2.ObjectName(val)
	return &objectName
}

var (
	PathMatchTypeDerefOr       = ptr.Deref[gwapiv1.PathMatchType]
	GRPCMethodMatchTypeDerefOr = ptr.Deref[gwapiv1.GRPCMethodMatchType]
	HeaderMatchTypeDerefOr     = ptr.Deref[gwapiv1.HeaderMatchType]
	QueryParamMatchTypeDerefOr = ptr.Deref[gwapiv1.QueryParamMatchType]
)

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
func IsRefToGateway(routeNamespace gwapiv1.Namespace, parentRef gwapiv1.ParentReference, gateway types.NamespacedName) bool {
	if parentRef.Group != nil && string(*parentRef.Group) != gwapiv1.GroupName {
		return false
	}

	if parentRef.Kind != nil && string(*parentRef.Kind) != KindGateway {
		return false
	}

	ns := routeNamespace
	if parentRef.Namespace != nil {
		ns = *parentRef.Namespace
	}

	if string(ns) != gateway.Namespace {
		return false
	}

	return string(parentRef.Name) == gateway.Name
}

// GetReferencedListeners returns whether a given parent ref references a Gateway
// in the given list, and if so, a list of the Listeners within that Gateway that
// are included by the parent ref (either one specific Listener, or all Listeners
// in the Gateway, depending on whether section name is specified or not).
func GetReferencedListeners(routeNamespace gwapiv1.Namespace, parentRef gwapiv1.ParentReference, gateways []*GatewayContext) (bool, []*ListenerContext) {
	var referencedListeners []*ListenerContext

	for _, gateway := range gateways {
		if IsRefToGateway(routeNamespace, parentRef, utils.NamespacedName(gateway)) {
			// The parentRef may be to the entire Gateway, or to a specific listener.
			for _, listener := range gateway.listeners {
				if (parentRef.SectionName == nil || *parentRef.SectionName == listener.Name) && (parentRef.Port == nil || *parentRef.Port == listener.Port) {
					referencedListeners = append(referencedListeners, listener)
				}
			}
			return true, referencedListeners
		}
	}

	return false, referencedListeners
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

// ValidateGRPCRouteFilter validates the provided filter within GRPCRoute.
func ValidateGRPCRouteFilter(filter *gwapiv1.GRPCRouteFilter, extGKs ...schema.GroupKind) error {
	switch {
	case filter == nil:
		return errors.New("filter is nil")
	case filter.Type == gwapiv1.GRPCRouteFilterRequestMirror ||
		filter.Type == gwapiv1.GRPCRouteFilterRequestHeaderModifier ||
		filter.Type == gwapiv1.GRPCRouteFilterResponseHeaderModifier:
		return nil
	case filter.Type == gwapiv1.GRPCRouteFilterExtensionRef:
		switch {
		case filter.ExtensionRef == nil:
			return errors.New("extensionRef field must be specified for an extended filter")
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

// OwnerLabels returns the owner labels based on the mergeGateways setting
func OwnerLabels(gateway *gwapiv1.Gateway, mergeGateways bool) map[string]string {
	if mergeGateways {
		return GatewayClassOwnerLabel(string(gateway.Spec.GatewayClassName))
	}

	return GatewayOwnerLabels(gateway.Namespace, gateway.Name)
}

// servicePortToContainerPort translates a service port into an ephemeral
// container port.
func servicePortToContainerPort(servicePort int32, envoyProxy *egv1a1.EnvoyProxy) int32 {
	if envoyProxy != nil {
		if !envoyProxy.NeedToSwitchPorts() {
			return servicePort
		}
	}

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

func irListenerName(listener *ListenerContext) string {
	return fmt.Sprintf("%s/%s/%s", listener.gateway.Namespace, listener.gateway.Name, listener.Name)
}

func irListenerPortName(proto ir.ProtocolType, port int32) string {
	return strings.ToLower(fmt.Sprintf("%s-%d", proto, port))
}

func irRoutePrefix(route RouteContext) string {
	// add a "/" at the end of the prefix to prevent mismatching routes with the
	// same prefix. For example, route prefix "/foo/" should not match a route "/foobar".
	return fmt.Sprintf("%s/%s/%s/", strings.ToLower(string(GetRouteType(route))), route.GetNamespace(), route.GetName())
}

func irRouteName(route RouteContext, ruleIdx, matchIdx int) string {
	return fmt.Sprintf("%srule/%d/match/%d", irRoutePrefix(route), ruleIdx, matchIdx)
}

func irTCPRouteName(route RouteContext) string {
	return fmt.Sprintf("%s/%s/%s", strings.ToLower(string(GetRouteType(route))), route.GetNamespace(), route.GetName())
}

func irUDPRouteName(route RouteContext) string {
	return irTCPRouteName(route)
}

func irRouteDestinationName(route RouteContext, ruleIdx int) string {
	return fmt.Sprintf("%srule/%d", irRoutePrefix(route), ruleIdx)
}

func irTLSConfigs(tlsSecrets ...*corev1.Secret) *ir.TLSConfig {
	if len(tlsSecrets) == 0 {
		return nil
	}

	tlsListenerConfigs := &ir.TLSConfig{
		Certificates: make([]ir.TLSCertificate, len(tlsSecrets)),
	}
	for i, tlsSecret := range tlsSecrets {
		tlsListenerConfigs.Certificates[i] = ir.TLSCertificate{
			Name:        irTLSListenerConfigName(tlsSecret),
			Certificate: tlsSecret.Data[corev1.TLSCertKey],
			PrivateKey:  tlsSecret.Data[corev1.TLSPrivateKeyKey],
		}
	}
	return tlsListenerConfigs
}

func irTLSListenerConfigName(secret *corev1.Secret) string {
	return fmt.Sprintf("%s/%s", secret.Namespace, secret.Name)
}

func irTLSCACertName(namespace, name string) string {
	return fmt.Sprintf("%s/%s/%s", namespace, name, caCertKey)
}

func IsMergeGatewaysEnabled(resources *Resources) bool {
	return resources.EnvoyProxyForGatewayClass != nil && resources.EnvoyProxyForGatewayClass.Spec.MergeGateways != nil && *resources.EnvoyProxyForGatewayClass.Spec.MergeGateways
}

func protocolSliceToStringSlice(protocols []gwapiv1.ProtocolType) []string {
	var protocolStrings []string
	for _, protocol := range protocols {
		protocolStrings = append(protocolStrings, string(protocol))
	}
	return protocolStrings
}

// getAncestorRefForPolicy returns Gateway as an ancestor reference for policy.
func getAncestorRefForPolicy(gatewayNN types.NamespacedName, sectionName *gwapiv1a2.SectionName) gwapiv1a2.ParentReference {
	return gwapiv1a2.ParentReference{
		Group:       GroupPtr(gwapiv1.GroupName),
		Kind:        KindPtr(KindGateway),
		Namespace:   NamespacePtr(gatewayNN.Namespace),
		Name:        gwapiv1.ObjectName(gatewayNN.Name),
		SectionName: sectionName,
	}
}

type policyTargetRouteKey struct {
	Kind      string
	Namespace string
	Name      string
}

type policyRouteTargetContext struct {
	RouteContext
	attached bool
}

type policyGatewayTargetContext struct {
	*GatewayContext
	attached bool
}

// listenersWithSameHTTPPort returns a list of the names of all other HTTP listeners
// that would share the same filter chain as the provided listener when translated
// to XDS
func listenersWithSameHTTPPort(xdsIR *ir.Xds, listener *ir.HTTPListener) []string {
	// if TLS is enabled, the listener would have its own filterChain in Envoy, so
	// no conflicts are possible
	if listener.TLS != nil {
		return nil
	}
	res := []string{}
	for _, http := range xdsIR.HTTP {
		if http == listener {
			continue
		}
		// Non-TLS listeners can be distinguished by their ports
		if http.Port == listener.Port {
			res = append(res, http.Name)
		}
	}
	return res
}

func parseCIDR(cidr string) (*ir.CIDRMatch, error) {
	ip, ipn, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	mask, _ := ipn.Mask.Size()
	return &ir.CIDRMatch{
		CIDR:    ipn.String(),
		IP:      ip.String(),
		MaskLen: uint32(mask),
		IsIPv6:  ip.To4() == nil,
	}, nil
}

func irConfigName(policy client.Object) string {
	return fmt.Sprintf(
		"%s/%s",
		strings.ToLower(policy.GetObjectKind().GroupVersionKind().Kind),
		utils.NamespacedName(policy).String())
}

type targetRefWithTimestamp struct {
	gwapiv1a2.LocalPolicyTargetReferenceWithSectionName
	CreationTimestamp metav1.Time
}

func getPolicyTargetRefs[T client.Object](policy egv1a1.PolicyTargetReferences, potentialTargets []T) []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName {
	dedup := sets.New[targetRefWithTimestamp]()
	for _, currSelector := range policy.TargetSelectors {
		labelSelector := labels.SelectorFromSet(currSelector.MatchLabels)
		for _, obj := range potentialTargets {
			gvk := obj.GetObjectKind().GroupVersionKind()
			if gvk.Kind != string(currSelector.Kind) ||
				gvk.Group != string(ptr.Deref(currSelector.Group, gwapiv1a2.GroupName)) {
				continue
			}

			if labelSelector.Matches(labels.Set(obj.GetLabels())) {
				dedup.Insert(targetRefWithTimestamp{
					CreationTimestamp: obj.GetCreationTimestamp(),
					LocalPolicyTargetReferenceWithSectionName: gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
						LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
							Group: gwapiv1a2.Group(gvk.Group),
							Kind:  gwapiv1a2.Kind(gvk.Kind),
							Name:  gwapiv1a2.ObjectName(obj.GetName()),
						},
					},
				})
			}
		}
	}
	selectorsList := dedup.UnsortedList()
	slices.SortFunc(selectorsList, func(i, j targetRefWithTimestamp) int {
		return i.CreationTimestamp.Compare(j.CreationTimestamp.Time)
	})
	ret := []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{}
	for _, v := range selectorsList {
		ret = append(ret, v.LocalPolicyTargetReferenceWithSectionName)
	}
	// Plain targetRefs in the policy don't have an associated creation timestamp, but can still refer
	// to targets that were already found via the selectors. Only add them to the returned list if
	// they are not yet there. Always add them at the end.
	fastLookup := sets.New(ret...)
	var emptyTargetRef gwapiv1a2.LocalPolicyTargetReferenceWithSectionName
	for _, v := range policy.GetTargetRefs() {
		if v == emptyTargetRef {
			// This can happen when the targetRef structure is read from extension server policies
			continue
		}
		if !fastLookup.Has(v) {
			ret = append(ret, v)
		}
	}

	return ret
}

// Sets *target to value if and only if *target is nil
func setIfNil[T any](target **T, value *T) {
	if *target == nil {
		*target = value
	}
}
