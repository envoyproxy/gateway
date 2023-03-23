// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"net/netip"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

func (t *Translator) validateBackendRef(backendRef *v1alpha2.BackendRef, parentRef *RouteParentContext, route RouteContext,
	resources *Resources, serviceNamespace string, routeKind v1beta1.Kind) bool {
	if !t.validateBackendRefGroup(backendRef, parentRef, route) {
		return false
	}
	if !t.validateBackendRefKind(backendRef, parentRef, route) {
		return false
	}
	if !t.validateBackendNamespace(backendRef, parentRef, route, resources, routeKind) {
		return false
	}
	if !t.validateBackendPort(backendRef, parentRef, route) {
		return false
	}
	protocol := v1.ProtocolTCP
	if routeKind == KindUDPRoute {
		protocol = v1.ProtocolUDP
	}
	if !t.validateBackendService(backendRef, parentRef, resources, serviceNamespace, route, protocol) {
		return false
	}
	return true
}

func (t *Translator) validateBackendRefGroup(backendRef *v1alpha2.BackendRef, parentRef *RouteParentContext, route RouteContext) bool {
	if backendRef.Group != nil && *backendRef.Group != "" {
		parentRef.SetCondition(route,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			v1beta1.RouteReasonInvalidKind,
			"Group is invalid, only the core API group (specified by omitting the group field or setting it to an empty string) is supported",
		)
		return false
	}
	return true
}

func (t *Translator) validateBackendRefKind(backendRef *v1alpha2.BackendRef, parentRef *RouteParentContext, route RouteContext) bool {
	if backendRef.Kind != nil && *backendRef.Kind != KindService {
		parentRef.SetCondition(route,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			v1beta1.RouteReasonInvalidKind,
			"Kind is invalid, only Service is supported",
		)
		return false
	}
	return true
}

func (t *Translator) validateBackendNamespace(backendRef *v1alpha2.BackendRef, parentRef *RouteParentContext, route RouteContext,
	resources *Resources, routeKind v1beta1.Kind) bool {
	if backendRef.Namespace != nil && string(*backendRef.Namespace) != "" && string(*backendRef.Namespace) != route.GetNamespace() {
		if !t.validateCrossNamespaceRef(
			crossNamespaceFrom{
				group:     v1beta1.GroupName,
				kind:      string(routeKind),
				namespace: route.GetNamespace(),
			},
			crossNamespaceTo{
				group:     "",
				kind:      KindService,
				namespace: string(*backendRef.Namespace),
				name:      string(backendRef.Name),
			},
			resources.ReferenceGrants,
		) {
			parentRef.SetCondition(route,
				v1beta1.RouteConditionResolvedRefs,
				metav1.ConditionFalse,
				v1beta1.RouteReasonRefNotPermitted,
				fmt.Sprintf("Backend ref to service %s/%s not permitted by any ReferenceGrant", *backendRef.Namespace, backendRef.Name),
			)
			return false
		}
	}
	return true
}

func (t *Translator) validateBackendPort(backendRef *v1alpha2.BackendRef, parentRef *RouteParentContext, route RouteContext) bool {
	if backendRef.Port == nil {
		parentRef.SetCondition(route,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			"PortNotSpecified",
			"A valid port number corresponding to a port on the Service must be specified",
		)
		return false
	}
	return true
}
func (t *Translator) validateBackendService(backendRef *v1alpha2.BackendRef, parentRef *RouteParentContext, resources *Resources,
	serviceNamespace string, route RouteContext, protocol v1.Protocol) bool {
	service := resources.GetService(serviceNamespace, string(backendRef.Name))
	if service == nil {
		parentRef.SetCondition(route,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			v1beta1.RouteReasonBackendNotFound,
			fmt.Sprintf("Service %s/%s not found", NamespaceDerefOr(backendRef.Namespace, route.GetNamespace()),
				string(backendRef.Name)),
		)
		return false
	}
	var portFound bool
	for _, port := range service.Spec.Ports {
		portProtocol := port.Protocol
		if port.Protocol == "" { // Default protocol is TCP
			portProtocol = v1.ProtocolTCP
		}
		if port.Port == int32(*backendRef.Port) && portProtocol == protocol {
			portFound = true
			break
		}
	}

	if !portFound {
		parentRef.SetCondition(route,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			"PortNotFound",
			fmt.Sprintf(string(protocol)+" Port %d not found on service %s/%s", *backendRef.Port, serviceNamespace,
				string(backendRef.Name)),
		)
		return false
	}
	return true
}

func (t *Translator) validateListenerConditions(listener *ListenerContext) (isReady bool) {
	lConditions := listener.GetConditions()
	if len(lConditions) == 0 {
		listener.SetCondition(v1beta1.ListenerConditionProgrammed, metav1.ConditionTrue, v1beta1.ListenerReasonProgrammed,
			"Listener has been successfully translated")
		return true

	}
	// Any condition on the listener apart from Programmed=true indicates an error.
	if !(lConditions[0].Type == string(v1beta1.ListenerConditionProgrammed) && lConditions[0].Status == metav1.ConditionTrue) {
		// set "Programmed: false" if it's not set already.
		var hasReadyCond bool
		for _, existing := range lConditions {
			if existing.Type == string(v1beta1.ListenerConditionProgrammed) {
				hasReadyCond = true
				break
			}
		}
		if !hasReadyCond {
			listener.SetCondition(
				v1beta1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				v1beta1.ListenerReasonInvalid,
				"Listener is invalid, see other Conditions for details.",
			)
		}
		// skip computing IR
		return false
	}
	return true
}

func (t *Translator) validateAllowedNamespaces(listener *ListenerContext) {
	if listener.AllowedRoutes != nil &&
		listener.AllowedRoutes.Namespaces != nil &&
		listener.AllowedRoutes.Namespaces.From != nil &&
		*listener.AllowedRoutes.Namespaces.From == v1beta1.NamespacesFromSelector {
		if listener.AllowedRoutes.Namespaces.Selector == nil {
			listener.SetCondition(
				v1beta1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				v1beta1.ListenerReasonInvalid,
				"The allowedRoutes.namespaces.selector field must be specified when allowedRoutes.namespaces.from is set to \"Selector\".",
			)
		} else {
			selector, err := metav1.LabelSelectorAsSelector(listener.AllowedRoutes.Namespaces.Selector)
			if err != nil {
				listener.SetCondition(
					v1beta1.ListenerConditionProgrammed,
					metav1.ConditionFalse,
					v1beta1.ListenerReasonInvalid,
					fmt.Sprintf("The allowedRoutes.namespaces.selector could not be parsed: %v.", err),
				)
			}

			listener.namespaceSelector = selector
		}
	}
}

func (t *Translator) validateTLSConfiguration(listener *ListenerContext, resources *Resources) {
	switch listener.Protocol {
	case v1beta1.HTTPProtocolType, v1beta1.UDPProtocolType, v1beta1.TCPProtocolType:
		if listener.TLS != nil {
			listener.SetCondition(
				v1beta1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				v1beta1.ListenerReasonInvalid,
				fmt.Sprintf("Listener must not have TLS set when protocol is %s.", listener.Protocol),
			)
		}
	case v1beta1.HTTPSProtocolType:
		if listener.TLS == nil {
			listener.SetCondition(
				v1beta1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				v1beta1.ListenerReasonInvalid,
				fmt.Sprintf("Listener must have TLS set when protocol is %s.", listener.Protocol),
			)
			break
		}

		if listener.TLS.Mode != nil && *listener.TLS.Mode != v1beta1.TLSModeTerminate {
			listener.SetCondition(
				v1beta1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				"UnsupportedTLSMode",
				fmt.Sprintf("TLS %s mode is not supported, TLS mode must be Terminate.", *listener.TLS.Mode),
			)
			break
		}

		if len(listener.TLS.CertificateRefs) != 1 {
			listener.SetCondition(
				v1beta1.ListenerConditionProgrammed,
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
			if !t.validateCrossNamespaceRef(
				crossNamespaceFrom{
					group:     v1beta1.GroupName,
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
					v1beta1.ListenerReasonRefNotPermitted,
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

		err := validateTLSSecretData(secret)
		if err != nil {
			listener.SetCondition(
				v1beta1.ListenerConditionResolvedRefs,
				metav1.ConditionFalse,
				v1beta1.ListenerReasonInvalidCertificateRef,
				fmt.Sprintf("Secret %s/%s must contain valid %s and %s, %s.", listener.gateway.Namespace, certificateRef.Name, v1.TLSCertKey, v1.TLSPrivateKeyKey, err.Error()),
			)
			break
		}

		listener.SetTLSSecret(secret)
	case v1beta1.TLSProtocolType:
		if listener.TLS == nil {
			listener.SetCondition(
				v1beta1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				v1beta1.ListenerReasonInvalid,
				fmt.Sprintf("Listener must have TLS set when protocol is %s.", listener.Protocol),
			)
			break
		}

		if listener.TLS.Mode != nil && *listener.TLS.Mode != v1beta1.TLSModePassthrough {
			listener.SetCondition(
				v1beta1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				"UnsupportedTLSMode",
				fmt.Sprintf("TLS %s mode is not supported, TLS mode must be Passthrough.", *listener.TLS.Mode),
			)
			break
		}

		if len(listener.TLS.CertificateRefs) > 0 {
			listener.SetCondition(
				v1beta1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				v1beta1.ListenerReasonInvalid,
				"Listener must not have TLS certificate refs set for TLS mode Passthrough",
			)
			break
		}
	}
}

func (t *Translator) validateHostName(listener *ListenerContext) {
	if listener.Protocol == v1beta1.UDPProtocolType || listener.Protocol == v1beta1.TCPProtocolType {
		if listener.Hostname != nil {
			listener.SetCondition(
				v1beta1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				v1beta1.ListenerReasonInvalid,
				fmt.Sprintf("Listener must not have hostname set when protocol is %s.", listener.Protocol),
			)
		}
	}
}

func (t *Translator) validateAllowedRoutes(listener *ListenerContext, routeKinds ...v1beta1.Kind) {
	canSupportKinds := make([]v1beta1.RouteGroupKind, len(routeKinds))
	for i, routeKind := range routeKinds {
		canSupportKinds[i] = v1beta1.RouteGroupKind{Group: GroupPtr(v1beta1.GroupName), Kind: routeKind}
	}
	if listener.AllowedRoutes == nil || len(listener.AllowedRoutes.Kinds) == 0 {
		listener.SetSupportedKinds(canSupportKinds...)
		return
	}

	supportedRouteKinds := make([]v1beta1.Kind, 0)
	supportedKinds := make([]v1beta1.RouteGroupKind, 0)
	unSupportedKinds := make([]v1beta1.RouteGroupKind, 0)

	for _, kind := range listener.AllowedRoutes.Kinds {

		// if there is a group it must match `gateway.networking.k8s.io`
		if kind.Group != nil && string(*kind.Group) != v1beta1.GroupName {
			listener.SetCondition(
				v1beta1.ListenerConditionResolvedRefs,
				metav1.ConditionFalse,
				v1beta1.ListenerReasonInvalidRouteKinds,
				fmt.Sprintf("Group is not supported, group must be %s", v1beta1.GroupName),
			)
			continue
		}

		found := false
		for _, routeKind := range routeKinds {
			if kind.Kind == routeKind {
				supportedKinds = append(supportedKinds, kind)
				supportedRouteKinds = append(supportedRouteKinds, kind.Kind)
				found = true
				break
			}
		}

		if !found {
			unSupportedKinds = append(unSupportedKinds, kind)
		}
	}

	for _, kind := range unSupportedKinds {
		var printRouteKinds []v1beta1.Kind
		if len(supportedKinds) == 0 {
			printRouteKinds = routeKinds
		} else {
			printRouteKinds = supportedRouteKinds
		}
		listener.SetCondition(
			v1beta1.ListenerConditionResolvedRefs,
			metav1.ConditionFalse,
			v1beta1.ListenerReasonInvalidRouteKinds,
			fmt.Sprintf("%s is not supported, kind must be one of %v", string(kind.Kind), printRouteKinds),
		)
	}

	listener.SetSupportedKinds(supportedKinds...)
}

type portListeners struct {
	listeners []*ListenerContext
	protocols sets.Set[string]
	hostnames map[string]int
}

func (t *Translator) validateConflictedLayer7Listeners(gateways []*GatewayContext) {
	// Iterate through all layer-7 (HTTP, HTTPS, TLS) listeners and collect info about protocols
	// and hostnames per port.
	for _, gateway := range gateways {
		portListenerInfo := map[v1beta1.PortNumber]*portListeners{}
		for _, listener := range gateway.listeners {
			if listener.Protocol == v1beta1.UDPProtocolType || listener.Protocol == v1beta1.TCPProtocolType {
				continue
			}
			if portListenerInfo[listener.Port] == nil {
				portListenerInfo[listener.Port] = &portListeners{
					protocols: sets.Set[string]{},
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
}

func (t *Translator) validateConflictedLayer4Listeners(gateways []*GatewayContext, protocol v1beta1.ProtocolType) {
	// Iterate through all layer-4(TCP UDP) listeners and check if there are more than one listener on the same port
	for _, gateway := range gateways {
		portListenerInfo := map[v1beta1.PortNumber]*portListeners{}
		for _, listener := range gateway.listeners {
			if listener.Protocol == protocol {
				if portListenerInfo[listener.Port] == nil {
					portListenerInfo[listener.Port] = &portListeners{}
				}
				portListenerInfo[listener.Port].listeners = append(portListenerInfo[listener.Port].listeners, listener)
			}
		}

		// Leave the first one and set Conflicted conditions for all other listeners with conflicting specs.
		for _, info := range portListenerInfo {
			if len(info.listeners) > 1 {
				for i := 1; i < len(info.listeners); i++ {
					info.listeners[i].SetCondition(
						v1beta1.ListenerConditionConflicted,
						metav1.ConditionTrue,
						v1beta1.ListenerReasonProtocolConflict,
						fmt.Sprintf("Only one %s listener is allowed in a given port", protocol),
					)
				}
			}
		}
	}
}

func (t *Translator) validateCrossNamespaceRef(from crossNamespaceFrom, to crossNamespaceTo, referenceGrants []*v1alpha2.ReferenceGrant) bool {
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
func (t *Translator) validateHostname(hostname string) error {
	if errs := validation.IsDNS1123Subdomain(hostname); errs != nil {
		return fmt.Errorf("hostname %q is invalid: %v", hostname, errs)
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
