// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/ir"
)

var _ ListenersTranslator = (*Translator)(nil)

type ListenersTranslator interface {
	ProcessListeners(gateways []*GatewayContext, xdsIR XdsIRMap, infraIR InfraIRMap, resources *Resources)
}

func (t *Translator) ProcessListeners(gateways []*GatewayContext, xdsIR XdsIRMap, infraIR InfraIRMap, resources *Resources) {

	t.validateConflictedLayer7Listeners(gateways)
	t.validateConflictedLayer4Listeners(gateways, v1beta1.TCPProtocolType)
	t.validateConflictedLayer4Listeners(gateways, v1beta1.UDPProtocolType)

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
		if len(t.ProxyImage) > 0 {
			gwInfraIR.Proxy.Image = t.ProxyImage
		}

		// save the IR references in the map before the translation starts
		xdsIR[irKey] = gwXdsIR
		infraIR[irKey] = gwInfraIR

		// Infra IR proxy ports must be unique.
		var foundPorts []*protocolPort

		for _, listener := range gateway.listeners {
			// Process protocol & supported kinds
			switch listener.Protocol {
			case v1beta1.TLSProtocolType:
				t.validateAllowedRoutes(listener, KindTLSRoute)
			case v1beta1.HTTPProtocolType, v1beta1.HTTPSProtocolType:
				t.validateAllowedRoutes(listener, KindHTTPRoute, KindGRPCRoute)
			case v1beta1.TCPProtocolType:
				t.validateAllowedRoutes(listener, KindTCPRoute)
			case v1beta1.UDPProtocolType:
				t.validateAllowedRoutes(listener, KindUDPRoute)
			default:
				listener.SetCondition(
					v1beta1.ListenerConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.ListenerReasonUnsupportedProtocol,
					fmt.Sprintf("Protocol %s is unsupported, must be %s, %s, %s or %s.", listener.Protocol,
						v1beta1.HTTPProtocolType, v1beta1.HTTPSProtocolType, v1beta1.TCPProtocolType, v1beta1.UDPProtocolType),
				)
			}

			// Validate allowed namespaces
			t.validateAllowedNamespaces(listener)

			// Process TLS configuration
			t.validateTLSConfiguration(listener, resources)

			// Process Hostname configuration
			t.validateHostName(listener)

			// Process conditions and check if the listener is ready
			isReady := t.validateListenerConditions(listener)
			if !isReady {
				continue
			}

			// Add the listener to the Xds IR
			servicePort := &protocolPort{protocol: listener.Protocol, port: int32(listener.Port)}
			containerPort := servicePortToContainerPort(servicePort.port)
			switch listener.Protocol {
			case v1beta1.HTTPProtocolType, v1beta1.HTTPSProtocolType:
				irListener := &ir.HTTPListener{
					Name:    irHTTPListenerName(listener),
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
			}

			// Add the listener to the Infra IR. Infra IR ports must have a unique port number per layer-4 protocol
			// (TCP or UDP).
			if !containsPort(foundPorts, servicePort) {
				foundPorts = append(foundPorts, servicePort)
				var proto ir.ProtocolType
				switch listener.Protocol {
				case v1beta1.HTTPProtocolType:
					proto = ir.HTTPProtocolType
				case v1beta1.HTTPSProtocolType:
					proto = ir.HTTPSProtocolType
				case v1beta1.TLSProtocolType:
					proto = ir.TLSProtocolType
				case v1beta1.TCPProtocolType:
					proto = ir.TCPProtocolType
				case v1beta1.UDPProtocolType:
					proto = ir.UDPProtocolType
				}
				infraPort := ir.ListenerPort{
					Name:          string(listener.Name),
					Protocol:      proto,
					ServicePort:   servicePort.port,
					ContainerPort: containerPort,
				}
				// Only 1 listener is supported.
				gwInfraIR.Proxy.Listeners[0].Ports = append(gwInfraIR.Proxy.Listeners[0].Ports, infraPort)
			}
		}
	}
}
