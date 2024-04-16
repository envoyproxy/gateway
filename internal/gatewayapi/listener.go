// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/utils/naming"
)

var _ ListenersTranslator = (*Translator)(nil)

type ListenersTranslator interface {
	ProcessListeners(gateways []*GatewayContext, xdsIR XdsIRMap, infraIR InfraIRMap, resources *Resources)
}

func (t *Translator) ProcessListeners(gateways []*GatewayContext, xdsIR XdsIRMap, infraIR InfraIRMap, resources *Resources) {
	// Infra IR proxy ports must be unique.
	foundPorts := make(map[string][]*protocolPort)
	t.validateConflictedLayer7Listeners(gateways)
	t.validateConflictedLayer4Listeners(gateways, gwapiv1.TCPProtocolType, gwapiv1.TLSProtocolType)
	t.validateConflictedLayer4Listeners(gateways, gwapiv1.UDPProtocolType)
	if t.MergeGateways {
		t.validateConflictedMergedListeners(gateways)
	}

	// Iterate through all listeners to validate spec
	// and compute status for each, and add valid ones
	// to the Xds IR.
	for _, gateway := range gateways {
		irKey := t.getIRKey(gateway.Gateway)

		if resources.EnvoyProxy != nil {
			infraIR[irKey].Proxy.Config = resources.EnvoyProxy
		}

		xdsIR[irKey].AccessLog = t.processAccessLog(infraIR[irKey].Proxy.Config, resources)
		xdsIR[irKey].Tracing = t.processTracing(gateway.Gateway, infraIR[irKey].Proxy.Config, resources)
		xdsIR[irKey].Metrics = t.processMetrics(infraIR[irKey].Proxy.Config)

		for _, listener := range gateway.listeners {
			// Process protocol & supported kinds
			switch listener.Protocol {
			case gwapiv1.TLSProtocolType:
				if listener.TLS != nil {
					switch *listener.TLS.Mode {
					case gwapiv1.TLSModePassthrough:
						t.validateAllowedRoutes(listener, KindTLSRoute)
					case gwapiv1.TLSModeTerminate:
						t.validateAllowedRoutes(listener, KindTCPRoute)
					default:
						t.validateAllowedRoutes(listener, KindTCPRoute, KindTLSRoute)
					}
				} else {
					t.validateAllowedRoutes(listener, KindTCPRoute, KindTLSRoute)
				}
			case gwapiv1.HTTPProtocolType, gwapiv1.HTTPSProtocolType:
				t.validateAllowedRoutes(listener, KindHTTPRoute, KindGRPCRoute)
			case gwapiv1.TCPProtocolType:
				t.validateAllowedRoutes(listener, KindTCPRoute)
			case gwapiv1.UDPProtocolType:
				t.validateAllowedRoutes(listener, KindUDPRoute)
			default:
				listener.SetCondition(
					gwapiv1.ListenerConditionAccepted,
					metav1.ConditionFalse,
					gwapiv1.ListenerReasonUnsupportedProtocol,
					fmt.Sprintf("Protocol %s is unsupported, must be %s, %s, %s or %s.", listener.Protocol,
						gwapiv1.HTTPProtocolType, gwapiv1.HTTPSProtocolType, gwapiv1.TCPProtocolType, gwapiv1.UDPProtocolType),
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
			case gwapiv1.HTTPProtocolType, gwapiv1.HTTPSProtocolType:
				irListener := &ir.HTTPListener{
					Name:    irHTTPListenerName(listener),
					Address: "0.0.0.0",
					Port:    uint32(containerPort),
					TLS:     irTLSConfigs(listener.tlsSecrets),
					Path: ir.PathSettings{
						MergeSlashes:         true,
						EscapedSlashesAction: ir.UnescapeAndRedirect,
					},
				}
				if listener.Hostname != nil {
					irListener.Hostnames = append(irListener.Hostnames, string(*listener.Hostname))
				} else {
					// Hostname specifies the virtual hostname to match for protocol types that define this concept.
					// When unspecified, all hostnames are matched. This field is ignored for protocols that don’t require hostname based matching.
					// see more https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/gwapiv1.Listener.
					irListener.Hostnames = append(irListener.Hostnames, "*")
				}
				xdsIR[irKey].HTTP = append(xdsIR[irKey].HTTP, irListener)
			}

			// Add the listener to the Infra IR. Infra IR ports must have a unique port number per layer-4 protocol
			// (TCP or UDP).
			if !containsPort(foundPorts[irKey], servicePort) {
				t.processInfraIRListener(listener, infraIR, irKey, servicePort)
				foundPorts[irKey] = append(foundPorts[irKey], servicePort)
			}
		}
	}
}

func (t *Translator) processInfraIRListener(listener *ListenerContext, infraIR InfraIRMap, irKey string, servicePort *protocolPort) {
	var proto ir.ProtocolType
	switch listener.Protocol {
	case gwapiv1.HTTPProtocolType:
		proto = ir.HTTPProtocolType
	case gwapiv1.HTTPSProtocolType:
		proto = ir.HTTPSProtocolType
	case gwapiv1.TLSProtocolType:
		proto = ir.TLSProtocolType
	case gwapiv1.TCPProtocolType:
		proto = ir.TCPProtocolType
	case gwapiv1.UDPProtocolType:
		proto = ir.UDPProtocolType
	}

	infraPortName := string(listener.Name)
	if t.MergeGateways {
		infraPortName = irHTTPListenerName(listener)
	}
	infraPort := ir.ListenerPort{
		Name:          infraPortName,
		Protocol:      proto,
		ServicePort:   servicePort.port,
		ContainerPort: servicePortToContainerPort(servicePort.port),
	}

	proxyListener := &ir.ProxyListener{
		Name:  irHTTPListenerName(listener),
		Ports: []ir.ListenerPort{infraPort},
	}

	infraIR[irKey].Proxy.Listeners = append(infraIR[irKey].Proxy.Listeners, proxyListener)
}

func (t *Translator) processAccessLog(envoyproxy *egv1a1.EnvoyProxy, resources *Resources) *ir.AccessLog {
	if envoyproxy == nil ||
		envoyproxy.Spec.Telemetry == nil ||
		envoyproxy.Spec.Telemetry.AccessLog == nil ||
		(!envoyproxy.Spec.Telemetry.AccessLog.Disable && len(envoyproxy.Spec.Telemetry.AccessLog.Settings) == 0) {
		// use the default access log
		return &ir.AccessLog{
			Text: []*ir.TextAccessLog{
				{
					Path: "/dev/stdout",
				},
			},
		}
	}

	if envoyproxy.Spec.Telemetry.AccessLog.Disable {
		return nil
	}

	irAccessLog := &ir.AccessLog{}
	// translate the access log configuration to the IR
	for _, accessLog := range envoyproxy.Spec.Telemetry.AccessLog.Settings {
		for index, sink := range accessLog.Sinks {
			switch sink.Type {
			case egv1a1.ProxyAccessLogSinkTypeFile:
				if sink.File == nil {
					continue
				}

				switch accessLog.Format.Type {
				case egv1a1.ProxyAccessLogFormatTypeText:
					al := &ir.TextAccessLog{
						Format: accessLog.Format.Text,
						Path:   sink.File.Path,
					}
					irAccessLog.Text = append(irAccessLog.Text, al)
				case egv1a1.ProxyAccessLogFormatTypeJSON:
					if len(accessLog.Format.JSON) == 0 {
						// TODO: use a default JSON format if not specified?
						continue
					}

					al := &ir.JSONAccessLog{
						JSON: accessLog.Format.JSON,
						Path: sink.File.Path,
					}
					irAccessLog.JSON = append(irAccessLog.JSON, al)
				}
			case egv1a1.ProxyAccessLogSinkTypeOpenTelemetry:
				if sink.OpenTelemetry == nil {
					continue
				}

				al := &ir.OpenTelemetryAccessLog{
					Resources: sink.OpenTelemetry.Resources,
					Destination: ir.RouteDestination{
						Name:     fmt.Sprintf("accesslog/%s/%s/sink/%d", envoyproxy.Namespace, envoyproxy.Name, index),
						Settings: make([]*ir.DestinationSetting, 0, len(sink.OpenTelemetry.BackendRefs)),
					},
				}

				for _, backendRef := range sink.OpenTelemetry.BackendRefs {
					al.Destination.Settings = append(al.Destination.Settings, t.processServiceDestination(backendRef, ir.GRPC, envoyproxy, resources))
				}

				// TODO: remove support for Host/Port in v1.2
				if sink.OpenTelemetry.Host != nil {
					al.Destination.Settings = append(al.Destination.Settings, &ir.DestinationSetting{
						Weight:   ptr.To(uint32(1)),
						Protocol: ir.GRPC,
						Endpoints: []*ir.DestinationEndpoint{
							{
								Port: uint32(sink.OpenTelemetry.Port),
								Host: *sink.OpenTelemetry.Host,
							},
						},
						AddressType: ptr.To(ir.FQDN),
					})
				}

				switch accessLog.Format.Type {
				case egv1a1.ProxyAccessLogFormatTypeJSON:
					al.Attributes = accessLog.Format.JSON
				case egv1a1.ProxyAccessLogFormatTypeText:
					al.Text = accessLog.Format.Text
				}

				irAccessLog.OpenTelemetry = append(irAccessLog.OpenTelemetry, al)
			}
		}
	}

	return irAccessLog
}

func (t *Translator) processTracing(gw *gwapiv1.Gateway, envoyproxy *egv1a1.EnvoyProxy, resources *Resources) *ir.Tracing {
	if envoyproxy == nil ||
		envoyproxy.Spec.Telemetry == nil ||
		envoyproxy.Spec.Telemetry.Tracing == nil {
		return nil
	}

	tracing := envoyproxy.Spec.Telemetry.Tracing
	tr := &ir.Tracing{
		ServiceName:  naming.ServiceName(utils.NamespacedName(gw)),
		SamplingRate: 100.0,
		CustomTags:   tracing.CustomTags,
		Destination: ir.RouteDestination{
			Name:     fmt.Sprintf("tracing/%s/%s", envoyproxy.Namespace, envoyproxy.Name),
			Settings: make([]*ir.DestinationSetting, 0, len(tracing.Provider.BackendRefs)),
		},
	}

	for _, backendRef := range tracing.Provider.BackendRefs {
		tr.Destination.Settings = append(tr.Destination.Settings, t.processServiceDestination(backendRef, ir.GRPC, envoyproxy, resources))
	}

	// TODO: remove support for Host/Port in v1.2
	if tracing.Provider.Host != nil {
		tr.Destination.Settings = append(tr.Destination.Settings, &ir.DestinationSetting{
			Weight:   ptr.To(uint32(1)),
			Protocol: ir.GRPC,
			Endpoints: []*ir.DestinationEndpoint{
				{
					Port: uint32(tracing.Provider.Port),
					Host: *tracing.Provider.Host,
				},
			},
			AddressType: ptr.To(ir.FQDN),
		})
	}

	if tracing.SamplingRate != nil {
		tr.SamplingRate = float64(*tracing.SamplingRate)
	}

	return tr
}

func (t *Translator) processMetrics(envoyproxy *egv1a1.EnvoyProxy) *ir.Metrics {
	if envoyproxy == nil ||
		envoyproxy.Spec.Telemetry == nil ||
		envoyproxy.Spec.Telemetry.Metrics == nil {
		return nil
	}
	return &ir.Metrics{
		EnableVirtualHostStats: envoyproxy.Spec.Telemetry.Metrics.EnableVirtualHostStats,
	}
}

func (t *Translator) processServiceDestination(backendRef egv1a1.BackendRef, protocol ir.AppProtocol, envoyproxy *egv1a1.EnvoyProxy, resources *Resources) *ir.DestinationSetting {
	var (
		endpoints   []*ir.DestinationEndpoint
		addrType    *ir.DestinationAddressType
		servicePort v1.ServicePort
		backendTLS  *ir.TLSUpstreamConfig
	)

	// TODO (davidalger) Handle case where Service referenced by backendRef doesn't exist
	serviceNamespace := NamespaceDerefOr(backendRef.Namespace, envoyproxy.Namespace)
	service := resources.GetService(serviceNamespace, string(backendRef.Name))
	for _, port := range service.Spec.Ports {
		if port.Port == int32(*backendRef.Port) {
			servicePort = port
			break
		}
	}

	if servicePort.AppProtocol != nil &&
		*servicePort.AppProtocol == "kubernetes.io/h2c" {
		protocol = ir.HTTP2
	}

	// Route to endpoints by default
	if !t.EndpointRoutingDisabled {
		endpointSlices := resources.GetEndpointSlicesForBackend(serviceNamespace, string(backendRef.Name), KindDerefOr(backendRef.Kind, KindService))
		endpoints, addrType = getIREndpointsFromEndpointSlices(endpointSlices, servicePort.Name, servicePort.Protocol)
	} else {
		// Fall back to Service ClusterIP routing
		ep := ir.NewDestEndpoint(service.Spec.ClusterIP, uint32(*backendRef.Port))
		endpoints = append(endpoints, ep)
	}

	return &ir.DestinationSetting{
		Weight:      ptr.To(uint32(1)),
		Protocol:    protocol,
		Endpoints:   endpoints,
		AddressType: addrType,
		TLS:         backendTLS,
	}
}
