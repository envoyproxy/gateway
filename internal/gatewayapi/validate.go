// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func (t *Translator) validateBackendRef(backendRefContext BackendRefContext, parentRef *RouteParentContext, route RouteContext,
	resources *Resources, backendNamespace string, routeKind gwapiv1.Kind) bool {
	if !t.validateBackendRefFilters(backendRefContext, parentRef, route, routeKind) {
		return false
	}
	backendRef := GetBackendRef(backendRefContext)

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
	backendRefKind := KindDerefOr(backendRef.Kind, KindService)
	switch backendRefKind {
	case KindService:
		if !t.validateBackendService(backendRef, parentRef, resources, backendNamespace, route, protocol) {
			return false
		}
	case KindServiceImport:
		if !t.validateBackendServiceImport(backendRef, parentRef, resources, backendNamespace, route, protocol) {
			return false
		}
	}
	return true
}

func (t *Translator) validateBackendRefGroup(backendRef *gwapiv1a2.BackendRef, parentRef *RouteParentContext, route RouteContext) bool {
	if backendRef.Group != nil && *backendRef.Group != "" && *backendRef.Group != GroupMultiClusterService {
		parentRef.SetCondition(route,
			gwapiv1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			gwapiv1.RouteReasonInvalidKind,
			fmt.Sprintf("Group is invalid, only the core API group (specified by omitting the group field or setting it to an empty string) and %s are supported", GroupMultiClusterService),
		)
		return false
	}
	return true
}

func (t *Translator) validateBackendRefKind(backendRef *gwapiv1a2.BackendRef, parentRef *RouteParentContext, route RouteContext) bool {
	if backendRef.Kind != nil && *backendRef.Kind != KindService && *backendRef.Kind != KindServiceImport {
		parentRef.SetCondition(route,
			gwapiv1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			gwapiv1.RouteReasonInvalidKind,
			"Kind is invalid, only Service and MCS ServiceImport are supported",
		)
		return false
	}
	return true
}

func (t *Translator) validateBackendRefFilters(backendRef BackendRefContext, parentRef *RouteParentContext, route RouteContext, routeKind gwapiv1.Kind) bool {
	var filtersLen int
	switch routeKind {
	case KindHTTPRoute:
		filters := GetFilters(backendRef).([]gwapiv1.HTTPRouteFilter)
		filtersLen = len(filters)
	case KindGRPCRoute:
		filters := GetFilters(backendRef).([]gwapiv1a2.GRPCRouteFilter)
		filtersLen = len(filters)
	default:
		return true
	}

	if filtersLen > 0 {
		parentRef.SetCondition(route,
			gwapiv1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			"UnsupportedRefValue",
			"The filters field within BackendRef is not supported",
		)
		return false
	}

	return true
}

func (t *Translator) validateBackendNamespace(backendRef *gwapiv1a2.BackendRef, parentRef *RouteParentContext, route RouteContext,
	resources *Resources, routeKind gwapiv1.Kind) bool {
	if backendRef.Namespace != nil && string(*backendRef.Namespace) != "" && string(*backendRef.Namespace) != route.GetNamespace() {
		if !t.validateCrossNamespaceRef(
			crossNamespaceFrom{
				group:     gwapiv1.GroupName,
				kind:      string(routeKind),
				namespace: route.GetNamespace(),
			},
			crossNamespaceTo{
				group:     GroupDerefOr(backendRef.Group, ""),
				kind:      KindDerefOr(backendRef.Kind, KindService),
				namespace: string(*backendRef.Namespace),
				name:      string(backendRef.Name),
			},
			resources.ReferenceGrants,
		) {
			parentRef.SetCondition(route,
				gwapiv1.RouteConditionResolvedRefs,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonRefNotPermitted,
				fmt.Sprintf("Backend ref to %s %s/%s not permitted by any ReferenceGrant.", KindDerefOr(backendRef.Kind, KindService), *backendRef.Namespace, backendRef.Name),
			)
			return false
		}
	}
	return true
}

func (t *Translator) validateBackendPort(backendRef *gwapiv1a2.BackendRef, parentRef *RouteParentContext, route RouteContext) bool {
	if backendRef.Port == nil {
		parentRef.SetCondition(route,
			gwapiv1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			"PortNotSpecified",
			"A valid port number corresponding to a port on the Service must be specified",
		)
		return false
	}
	return true
}
func (t *Translator) validateBackendService(backendRef *gwapiv1a2.BackendRef, parentRef *RouteParentContext, resources *Resources,
	serviceNamespace string, route RouteContext, protocol v1.Protocol) bool {
	service := resources.GetService(serviceNamespace, string(backendRef.Name))
	if service == nil {
		parentRef.SetCondition(route,
			gwapiv1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			gwapiv1.RouteReasonBackendNotFound,
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
			gwapiv1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			"PortNotFound",
			fmt.Sprintf(string(protocol)+" Port %d not found on service %s/%s", *backendRef.Port, serviceNamespace,
				string(backendRef.Name)),
		)
		return false
	}
	return true
}

func (t *Translator) validateBackendServiceImport(backendRef *gwapiv1a2.BackendRef, parentRef *RouteParentContext, resources *Resources,
	serviceImportNamespace string, route RouteContext, protocol v1.Protocol) bool {
	serviceImport := resources.GetServiceImport(serviceImportNamespace, string(backendRef.Name))
	if serviceImport == nil {
		parentRef.SetCondition(route,
			gwapiv1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			gwapiv1.RouteReasonBackendNotFound,
			fmt.Sprintf("ServiceImport %s/%s not found", NamespaceDerefOr(backendRef.Namespace, route.GetNamespace()),
				string(backendRef.Name)),
		)
		return false
	}
	var portFound bool
	for _, port := range serviceImport.Spec.Ports {
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
			gwapiv1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			"PortNotFound",
			fmt.Sprintf(string(protocol)+" Port %d not found on ServiceImport %s/%s", *backendRef.Port, serviceImportNamespace,
				string(backendRef.Name)),
		)
		return false
	}
	return true
}

func (t *Translator) validateListenerConditions(listener *ListenerContext) (isReady bool) {
	lConditions := listener.GetConditions()
	if len(lConditions) == 0 {
		listener.SetCondition(gwapiv1.ListenerConditionProgrammed, metav1.ConditionTrue, gwapiv1.ListenerReasonProgrammed,
			"Sending translated listener configuration to the data plane")
		listener.SetCondition(gwapiv1.ListenerConditionAccepted, metav1.ConditionTrue, gwapiv1.ListenerReasonAccepted,
			"Listener has been successfully translated")
		listener.SetCondition(gwapiv1.ListenerConditionResolvedRefs, metav1.ConditionTrue, gwapiv1.ListenerReasonResolvedRefs,
			"Listener references have been resolved")
		return true
	}

	// Any condition on the listener apart from Programmed=true indicates an error.
	if !(lConditions[0].Type == string(gwapiv1.ListenerConditionProgrammed) && lConditions[0].Status == metav1.ConditionTrue) {
		hasProgrammedCond := false
		hasRefsCond := false
		for _, existing := range lConditions {
			if existing.Type == string(gwapiv1.ListenerConditionProgrammed) {
				hasProgrammedCond = true
			}
			if existing.Type == string(gwapiv1.ListenerConditionResolvedRefs) {
				hasRefsCond = true
			}
		}
		// set "Programmed: false" if it's not set already.
		if !hasProgrammedCond {
			listener.SetCondition(
				gwapiv1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				gwapiv1.ListenerReasonInvalid,
				"Listener is invalid, see other Conditions for details.",
			)
		}
		// set "ResolvedRefs: true" if it's not set already.
		if !hasRefsCond {
			listener.SetCondition(
				gwapiv1.ListenerConditionResolvedRefs,
				metav1.ConditionTrue,
				gwapiv1.ListenerReasonResolvedRefs,
				"Listener references have been resolved",
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
		*listener.AllowedRoutes.Namespaces.From == gwapiv1.NamespacesFromSelector {
		if listener.AllowedRoutes.Namespaces.Selector == nil {
			listener.SetCondition(
				gwapiv1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				gwapiv1.ListenerReasonInvalid,
				"The allowedRoutes.namespaces.selector field must be specified when allowedRoutes.namespaces.from is set to \"Selector\".",
			)
		} else {
			selector, err := metav1.LabelSelectorAsSelector(listener.AllowedRoutes.Namespaces.Selector)
			if err != nil {
				listener.SetCondition(
					gwapiv1.ListenerConditionProgrammed,
					metav1.ConditionFalse,
					gwapiv1.ListenerReasonInvalid,
					fmt.Sprintf("The allowedRoutes.namespaces.selector could not be parsed: %v.", err),
				)
			}

			listener.namespaceSelector = selector
		}
	}
}

func (t *Translator) validateTerminateModeAndGetTLSSecrets(listener *ListenerContext, resources *Resources) []*v1.Secret {
	if len(listener.TLS.CertificateRefs) == 0 {
		listener.SetCondition(
			gwapiv1.ListenerConditionProgrammed,
			metav1.ConditionFalse,
			gwapiv1.ListenerReasonInvalid,
			"Listener must have at least 1 TLS certificate ref",
		)
		return nil
	}

	secrets := make([]*v1.Secret, 0)
	for _, certificateRef := range listener.TLS.CertificateRefs {
		// TODO zhaohuabing: reuse validateSecretRef
		if certificateRef.Group != nil && string(*certificateRef.Group) != "" {
			listener.SetCondition(
				gwapiv1.ListenerConditionResolvedRefs,
				metav1.ConditionFalse,
				gwapiv1.ListenerReasonInvalidCertificateRef,
				"Listener's TLS certificate ref group must be unspecified/empty.",
			)
			break
		}

		if certificateRef.Kind != nil && string(*certificateRef.Kind) != KindSecret {
			listener.SetCondition(
				gwapiv1.ListenerConditionResolvedRefs,
				metav1.ConditionFalse,
				gwapiv1.ListenerReasonInvalidCertificateRef,
				fmt.Sprintf("Listener's TLS certificate ref kind must be %s.", KindSecret),
			)
			break
		}

		secretNamespace := listener.gateway.Namespace

		if certificateRef.Namespace != nil && string(*certificateRef.Namespace) != "" && string(*certificateRef.Namespace) != listener.gateway.Namespace {
			if !t.validateCrossNamespaceRef(
				crossNamespaceFrom{
					group:     gwapiv1.GroupName,
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
					gwapiv1.ListenerConditionResolvedRefs,
					metav1.ConditionFalse,
					gwapiv1.ListenerReasonRefNotPermitted,
					fmt.Sprintf("Certificate ref to secret %s/%s not permitted by any ReferenceGrant.", *certificateRef.Namespace, certificateRef.Name),
				)
				break
			}

			secretNamespace = string(*certificateRef.Namespace)
		}

		secret := resources.GetSecret(secretNamespace, string(certificateRef.Name))

		if secret == nil {
			listener.SetCondition(
				gwapiv1.ListenerConditionResolvedRefs,
				metav1.ConditionFalse,
				gwapiv1.ListenerReasonInvalidCertificateRef,
				fmt.Sprintf("Secret %s/%s does not exist.", listener.gateway.Namespace, certificateRef.Name),
			)
			break
		}

		if secret.Type != v1.SecretTypeTLS {
			listener.SetCondition(
				gwapiv1.ListenerConditionResolvedRefs,
				metav1.ConditionFalse,
				gwapiv1.ListenerReasonInvalidCertificateRef,
				fmt.Sprintf("Secret %s/%s must be of type %s.", listener.gateway.Namespace, certificateRef.Name, v1.SecretTypeTLS),
			)
			break
		}

		if len(secret.Data[v1.TLSCertKey]) == 0 || len(secret.Data[v1.TLSPrivateKeyKey]) == 0 {
			listener.SetCondition(
				gwapiv1.ListenerConditionResolvedRefs,
				metav1.ConditionFalse,
				gwapiv1.ListenerReasonInvalidCertificateRef,
				fmt.Sprintf("Secret %s/%s must contain %s and %s.", listener.gateway.Namespace, certificateRef.Name, v1.TLSCertKey, v1.TLSPrivateKeyKey),
			)
			break
		}

		secrets = append(secrets, secret)
	}

	err := validateTLSSecretsData(secrets, listener.Hostname)
	if err != nil {
		listener.SetCondition(
			gwapiv1.ListenerConditionResolvedRefs,
			metav1.ConditionFalse,
			gwapiv1.ListenerReasonInvalidCertificateRef,
			fmt.Sprintf("Secret %s.", err.Error()),
		)
	}

	return secrets
}

func (t *Translator) validateTLSConfiguration(listener *ListenerContext, resources *Resources) {
	switch listener.Protocol {
	case gwapiv1.HTTPProtocolType, gwapiv1.UDPProtocolType, gwapiv1.TCPProtocolType:
		if listener.TLS != nil {
			listener.SetCondition(
				gwapiv1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				gwapiv1.ListenerReasonInvalid,
				fmt.Sprintf("Listener must not have TLS set when protocol is %s.", listener.Protocol),
			)
		}
	case gwapiv1.HTTPSProtocolType:
		if listener.TLS == nil {
			listener.SetCondition(
				gwapiv1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				gwapiv1.ListenerReasonInvalid,
				fmt.Sprintf("Listener must have TLS set when protocol is %s.", listener.Protocol),
			)
			break
		}

		if listener.TLS.Mode != nil && *listener.TLS.Mode != gwapiv1.TLSModeTerminate {
			listener.SetCondition(
				gwapiv1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				"UnsupportedTLSMode",
				fmt.Sprintf("TLS %s mode is not supported, TLS mode must be Terminate.", *listener.TLS.Mode),
			)
			break
		}

		secrets := t.validateTerminateModeAndGetTLSSecrets(listener, resources)
		listener.SetTLSSecrets(secrets)

	case gwapiv1.TLSProtocolType:
		if listener.TLS == nil {
			listener.SetCondition(
				gwapiv1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				gwapiv1.ListenerReasonInvalid,
				fmt.Sprintf("Listener must have TLS set when protocol is %s.", listener.Protocol),
			)
			break
		}

		if listener.TLS.Mode != nil && *listener.TLS.Mode == gwapiv1.TLSModePassthrough {
			if len(listener.TLS.CertificateRefs) > 0 {
				listener.SetCondition(
					gwapiv1.ListenerConditionProgrammed,
					metav1.ConditionFalse,
					gwapiv1.ListenerReasonInvalid,
					"Listener must not have TLS certificate refs set for TLS mode Passthrough.",
				)
				break
			}
		}

		if listener.TLS.Mode != nil && *listener.TLS.Mode == gwapiv1.TLSModeTerminate {
			if len(listener.TLS.CertificateRefs) == 0 {
				listener.SetCondition(
					gwapiv1.ListenerConditionProgrammed,
					metav1.ConditionFalse,
					gwapiv1.ListenerReasonInvalid,
					"Listener must have TLS certificate refs set for TLS mode Terminate.",
				)
				break
			}
			secrets := t.validateTerminateModeAndGetTLSSecrets(listener, resources)
			listener.SetTLSSecrets(secrets)
		}
	}
}

func (t *Translator) validateHostName(listener *ListenerContext) {
	if listener.Protocol == gwapiv1.UDPProtocolType || listener.Protocol == gwapiv1.TCPProtocolType {
		if listener.Hostname != nil {
			listener.SetCondition(
				gwapiv1.ListenerConditionProgrammed,
				metav1.ConditionFalse,
				gwapiv1.ListenerReasonInvalid,
				fmt.Sprintf("Listener must not have hostname set when protocol is %s.", listener.Protocol),
			)
		}
	}
}

func (t *Translator) validateAllowedRoutes(listener *ListenerContext, routeKinds ...gwapiv1.Kind) {
	canSupportKinds := make([]gwapiv1.RouteGroupKind, len(routeKinds))
	for i, routeKind := range routeKinds {
		canSupportKinds[i] = gwapiv1.RouteGroupKind{Group: GroupPtr(gwapiv1.GroupName), Kind: routeKind}
	}
	if listener.AllowedRoutes == nil || len(listener.AllowedRoutes.Kinds) == 0 {
		listener.SetSupportedKinds(canSupportKinds...)
		return
	}

	supportedRouteKinds := make([]gwapiv1.Kind, 0)
	supportedKinds := make([]gwapiv1.RouteGroupKind, 0)
	unSupportedKinds := make([]gwapiv1.RouteGroupKind, 0)

	for _, kind := range listener.AllowedRoutes.Kinds {

		// if there is a group it must match `gateway.networking.k8s.io`
		if kind.Group != nil && string(*kind.Group) != gwapiv1.GroupName {
			listener.SetCondition(
				gwapiv1.ListenerConditionResolvedRefs,
				metav1.ConditionFalse,
				gwapiv1.ListenerReasonInvalidRouteKinds,
				fmt.Sprintf("Group is not supported, group must be %s", gwapiv1.GroupName),
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
		var printRouteKinds []gwapiv1.Kind
		if len(supportedKinds) == 0 {
			printRouteKinds = routeKinds
		} else {
			printRouteKinds = supportedRouteKinds
		}
		listener.SetCondition(
			gwapiv1.ListenerConditionResolvedRefs,
			metav1.ConditionFalse,
			gwapiv1.ListenerReasonInvalidRouteKinds,
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

// Port, protocol and hostname tuple should be unique across all listeners on merged Gateways.
func (t *Translator) validateConflictedMergedListeners(gateways []*GatewayContext) {
	listenerSets := sets.Set[string]{}
	for _, gateway := range gateways {
		for _, listener := range gateway.listeners {
			hostname := new(gwapiv1.Hostname)
			if listener.Hostname != nil {
				hostname = listener.Hostname
			}
			portProtocolHostname := fmt.Sprintf("%s:%s:%d", listener.Protocol, *hostname, listener.Port)
			if listenerSets.Has(portProtocolHostname) {
				listener.SetCondition(
					gwapiv1.ListenerConditionConflicted,
					metav1.ConditionTrue,
					gwapiv1.ListenerReasonHostnameConflict,
					"Port, protocol and hostname tuple must be unique for every listener",
				)
			}
			listenerSets.Insert(portProtocolHostname)
		}
	}
}

func (t *Translator) validateConflictedLayer7Listeners(gateways []*GatewayContext) {
	// Iterate through all layer-7 (HTTP, HTTPS, TLS) listeners and collect info about protocols
	// and hostnames per port.
	for _, gateway := range gateways {
		portListenerInfo := map[gwapiv1.PortNumber]*portListeners{}
		for _, listener := range gateway.listeners {
			if listener.Protocol == gwapiv1.UDPProtocolType || listener.Protocol == gwapiv1.TCPProtocolType {
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
			case gwapiv1.HTTPSProtocolType, gwapiv1.TLSProtocolType:
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
						gwapiv1.ListenerConditionConflicted,
						metav1.ConditionTrue,
						gwapiv1.ListenerReasonProtocolConflict,
						"All listeners for a given port must use a compatible protocol",
					)
				}

				var hostname string
				if listener.Hostname != nil {
					hostname = string(*listener.Hostname)
				}

				if info.hostnames[hostname] > 1 {
					listener.SetCondition(
						gwapiv1.ListenerConditionConflicted,
						metav1.ConditionTrue,
						gwapiv1.ListenerReasonHostnameConflict,
						"All listeners for a given port must use a unique hostname",
					)
				}
			}
		}
	}
}

func (t *Translator) validateConflictedLayer4Listeners(gateways []*GatewayContext, protocols ...gwapiv1.ProtocolType) {
	// Iterate through all layer-4(TCP UDP) listeners and check if there are more than one listener on the same port
	for _, gateway := range gateways {
		portListenerInfo := map[gwapiv1.PortNumber]*portListeners{}
		for _, listener := range gateway.listeners {
			for _, protocol := range protocols {
				if listener.Protocol == protocol {
					if portListenerInfo[listener.Port] == nil {
						portListenerInfo[listener.Port] = &portListeners{}
					}
					portListenerInfo[listener.Port].listeners = append(portListenerInfo[listener.Port].listeners, listener)
				}
			}
		}

		// Leave the first one and set Conflicted conditions for all other listeners with conflicting specs.
		for _, info := range portListenerInfo {
			if len(info.listeners) > 1 {
				for i := 1; i < len(info.listeners); i++ {
					info.listeners[i].SetCondition(
						gwapiv1.ListenerConditionConflicted,
						metav1.ConditionTrue,
						gwapiv1.ListenerReasonProtocolConflict,
						fmt.Sprintf("Only one %s listener is allowed in a given port", strings.Join(protocolSliceToStringSlice(protocols), "/")),
					)
				}
			}
		}
	}
}

func (t *Translator) validateCrossNamespaceRef(from crossNamespaceFrom, to crossNamespaceTo, referenceGrants []*gwapiv1b1.ReferenceGrant) bool {
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

// validateSecretRef checks three things:
//  1. Does the secret reference have a valid Group and kind
//  2. If the secret reference is a cross-namespace reference,
//     is it permitted by any ReferenceGrant
//  3. Does the secret exist
func (t *Translator) validateSecretRef(
	allowCrossNamespace bool,
	from crossNamespaceFrom,
	secretObjRef gwapiv1b1.SecretObjectReference,
	resources *Resources) (*v1.Secret, error) {

	if err := t.validateSecretObjectRef(allowCrossNamespace, from, secretObjRef, resources); err != nil {
		return nil, err
	}

	secretNamespace := from.namespace
	if secretObjRef.Namespace != nil {
		secretNamespace = string(*secretObjRef.Namespace)
	}
	secret := resources.GetSecret(secretNamespace, string(secretObjRef.Name))

	if secret == nil {
		return nil, fmt.Errorf(
			"secret %s/%s does not exist", secretNamespace, secretObjRef.Name)
	}

	return secret, nil
}

func (t *Translator) validateConfigMapRef(
	allowCrossNamespace bool,
	from crossNamespaceFrom,
	secretObjRef gwapiv1b1.SecretObjectReference,
	resources *Resources) (*v1.ConfigMap, error) {

	if err := t.validateSecretObjectRef(allowCrossNamespace, from, secretObjRef, resources); err != nil {
		return nil, err
	}

	configMapNamespace := from.namespace
	if secretObjRef.Namespace != nil {
		configMapNamespace = string(*secretObjRef.Namespace)
	}
	configMap := resources.GetConfigMap(configMapNamespace, string(secretObjRef.Name))

	if configMap == nil {
		return nil, fmt.Errorf(
			"configmap %s/%s does not exist", configMapNamespace, secretObjRef.Name)
	}

	return configMap, nil
}

func (t *Translator) validateSecretObjectRef(
	allowCrossNamespace bool,
	from crossNamespaceFrom,
	secretRef gwapiv1b1.SecretObjectReference,
	resources *Resources) error {
	var kind string
	if secretRef.Group != nil && string(*secretRef.Group) != "" {
		return errors.New("secret ref group must be unspecified/empty")
	}

	if secretRef.Kind == nil { // nolint
		kind = KindSecret
	} else if string(*secretRef.Kind) == KindSecret {
		kind = KindSecret
	} else if string(*secretRef.Kind) == KindConfigMap {
		kind = KindConfigMap
	} else {
		return fmt.Errorf("secret ref kind must be %s", KindSecret)
	}

	if secretRef.Namespace != nil &&
		string(*secretRef.Namespace) != "" &&
		string(*secretRef.Namespace) != from.namespace {
		if !allowCrossNamespace {
			return fmt.Errorf(
				"secret ref namespace must be unspecified/empty or %s",
				from.namespace)
		}

		if !t.validateCrossNamespaceRef(
			from,
			crossNamespaceTo{
				group:     "",
				kind:      kind,
				namespace: string(*secretRef.Namespace),
				name:      string(secretRef.Name),
			},
			resources.ReferenceGrants,
		) {
			return fmt.Errorf(
				"certificate ref to secret %s/%s not permitted by any ReferenceGrant",
				*secretRef.Namespace, secretRef.Name)
		}

	}

	return nil
}

// TODO: zhaohuabing combine this function with the one in the route translator
// validateExtServiceBackendReference validates the backend reference for an
// external service referenced by an EG policy.
// This can also be used for the other external services deployed in the cluster,
// such as the external processing filter, gRPC Access Log Service, etc.
// It checks:
//  1. The group is nil or empty, indicating the core API group.
//  2. The kind is Service.
//  3. The port is specified.
//  4. The service exists and the specified port is found.
//  5. The cross-namespace reference is permitted by the ReferenceGrants if the
//     namespace is different from the policy's namespace.
func (t *Translator) validateExtServiceBackendReference(
	backendRef *gwapiv1.BackendObjectReference,
	ownerNamespace string,
	resources *Resources) error {

	// These are sanity checks, they should never happen because the API server
	// should have caught them
	if backendRef.Group != nil && *backendRef.Group != "" {
		return errors.New(
			"group is invalid, only the core API group (specified by omitting" +
				" the group field or setting it to an empty string) is supported")
	}
	if backendRef.Kind != nil && *backendRef.Kind != KindService {
		return errors.New("kind is invalid, only Service (specified by omitting " +
			"the kind field or setting it to 'Service') is supported")
	}
	if backendRef.Port == nil {
		return errors.New("a valid port number corresponding to a port on the Service must be specified")
	}

	// check if the service is valid
	serviceNamespace := NamespaceDerefOr(backendRef.Namespace, ownerNamespace)
	service := resources.GetService(serviceNamespace, string(backendRef.Name))
	if service == nil {
		return fmt.Errorf("service %s/%s not found", serviceNamespace, backendRef.Name)
	}
	var portFound bool
	for _, port := range service.Spec.Ports {
		portProtocol := port.Protocol
		if port.Protocol == "" { // Default protocol is TCP
			portProtocol = v1.ProtocolTCP
		}
		// currently only HTTP and GRPC are supported, both of which are TCP
		if port.Port == int32(*backendRef.Port) && portProtocol == v1.ProtocolTCP {
			portFound = true
			break
		}
	}

	if !portFound {
		return fmt.Errorf(
			"TCP Port %d not found on service %s/%s",
			*backendRef.Port, serviceNamespace, string(backendRef.Name),
		)
	}

	// check if the cross-namespace reference is permitted
	if backendRef.Namespace != nil && string(*backendRef.Namespace) != "" &&
		string(*backendRef.Namespace) != ownerNamespace {
		if !t.validateCrossNamespaceRef(
			crossNamespaceFrom{
				group:     egv1a1.GroupName,
				kind:      KindSecurityPolicy,
				namespace: ownerNamespace,
			},
			crossNamespaceTo{
				group:     GroupDerefOr(backendRef.Group, ""),
				kind:      KindDerefOr(backendRef.Kind, KindService),
				namespace: string(*backendRef.Namespace),
				name:      string(backendRef.Name),
			},
			resources.ReferenceGrants,
		) {
			return fmt.Errorf(
				"backend ref to %s %s/%s not permitted by any ReferenceGrant",
				KindService, *backendRef.Namespace, backendRef.Name)
		}
	}
	return nil
}
