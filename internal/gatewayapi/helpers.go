package gatewayapi

import (
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

func GroupPtr(name string) *v1beta1.Group {
	group := v1beta1.Group(name)
	return &group
}

func KindPtr(name string) *v1beta1.Kind {
	kind := v1beta1.Kind(name)
	return &kind
}

func NamespacePtr(name string) *v1beta1.Namespace {
	namespace := v1beta1.Namespace(name)
	return &namespace
}

func FromNamespacesPtr(fromNamespaces v1beta1.FromNamespaces) *v1beta1.FromNamespaces {
	return &fromNamespaces
}

func SectionNamePtr(name string) *v1beta1.SectionName {
	sectionName := v1beta1.SectionName(name)
	return &sectionName
}

func StringPtr(val string) *string {
	return &val
}

func Int32Ptr(val int32) *int32 {
	return &val
}

func PortNumPtr(val int32) *v1beta1.PortNumber {
	portNum := v1beta1.PortNumber(val)
	return &portNum
}

func ObjectNamePtr(val string) *v1alpha2.ObjectName {
	objectName := v1alpha2.ObjectName(val)
	return &objectName
}

func PathMatchTypePtr(pType v1beta1.PathMatchType) *v1beta1.PathMatchType {
	return &pType
}

func GatewayAddressTypePtr(addr v1beta1.AddressType) *v1beta1.AddressType {
	return &addr
}

func PathMatchTypeDerefOr(matchType *v1beta1.PathMatchType, defaultType v1beta1.PathMatchType) v1beta1.PathMatchType {
	if matchType != nil {
		return *matchType
	}
	return defaultType
}

func HeaderMatchTypeDerefOr(matchType *v1beta1.HeaderMatchType, defaultType v1beta1.HeaderMatchType) v1beta1.HeaderMatchType {
	if matchType != nil {
		return *matchType
	}
	return defaultType
}

func NamespaceDerefOr(namespace *v1beta1.Namespace, defaultNamespace string) string {
	if namespace != nil && *namespace != "" {
		return string(*namespace)
	}
	return defaultNamespace
}

func GroupDerefOr(group *v1beta1.Group, defaultGroup string) string {
	if group != nil && *group != "" {
		return string(*group)
	}
	return defaultGroup
}

// IsRefToGateway returns whether the provided parent ref is a reference
// to a Gateway with the given namespace/name, irrespective of whether a
// section/listener name has been specified (i.e. a parent ref to a listener
// on the specified gateway will return "true").
func IsRefToGateway(parentRef v1beta1.ParentReference, gateway types.NamespacedName) bool {
	if parentRef.Group != nil && string(*parentRef.Group) != v1beta1.GroupName {
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
func GetReferencedListeners(parentRef v1beta1.ParentReference, gateways []*GatewayContext) (bool, []*ListenerContext) {
	var selectsGateway bool
	var referencedListeners []*ListenerContext

	for _, gateway := range gateways {
		if !IsRefToGateway(parentRef, types.NamespacedName{Namespace: gateway.Namespace, Name: gateway.Name}) {
			continue
		}

		selectsGateway = true

		// The parentRef may be to the entire Gateway, or to a specific listener.
		for listenerName, listener := range gateway.listeners {
			if parentRef.SectionName == nil || *parentRef.SectionName == listenerName {
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

// computeHosts returns a list of the intersecting hostnames between the route
// and the listener.
func computeHosts(routeHostnames []string, listenerHostname *v1beta1.Hostname) []string {
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
