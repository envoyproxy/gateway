// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/utils/naming"
	"github.com/envoyproxy/gateway/internal/utils/net"
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
		t.processProxyObservability(gateway.Gateway, xdsIR[irKey], infraIR[irKey].Proxy.Config)

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

func (t *Translator) processProxyObservability(gw *gwapiv1.Gateway, xdsIR *ir.Xds, envoyProxy *egv1a1.EnvoyProxy) {
	xdsIR.AccessLog = processAccessLog(envoyProxy)
	xdsIR.Tracing = processTracing(gw, envoyProxy, t.MergeGateways)
	xdsIR.Metrics = processMetrics(envoyProxy)
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

	infraPort := ir.ListenerPort{
		Name:          irListenerPortName(proto, servicePort.port),
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

func processAccessLog(envoyproxy *egv1a1.EnvoyProxy) *ir.AccessLog {
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
		for _, sink := range accessLog.Sinks {
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

				// TODO: remove support for Host/Port in v1.2
				al := &ir.OpenTelemetryAccessLog{
					Port:      uint32(sink.OpenTelemetry.Port),
					Resources: sink.OpenTelemetry.Resources,
				}

				if sink.OpenTelemetry.Host != nil {
					al.Host = *sink.OpenTelemetry.Host
				}

				if len(sink.OpenTelemetry.BackendRefs) > 0 {
					al.Host, al.Port = net.BackendHostAndPort(sink.OpenTelemetry.BackendRefs[0].BackendObjectReference, envoyproxy.Namespace)
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

func processTracing(gw *gwapiv1.Gateway, envoyproxy *egv1a1.EnvoyProxy, mergeGateways bool) *ir.Tracing {
	if envoyproxy == nil ||
		envoyproxy.Spec.Telemetry == nil ||
		envoyproxy.Spec.Telemetry.Tracing == nil {
		return nil
	}
	tracing := envoyproxy.Spec.Telemetry.Tracing

	// TODO: remove support for Host/Port in v1.2
	var host string
	var port uint32
	if tracing.Provider.Host != nil {
		host, port = *tracing.Provider.Host, uint32(tracing.Provider.Port)
	}
	if len(tracing.Provider.BackendRefs) > 0 {
		host, port = net.BackendHostAndPort(tracing.Provider.BackendRefs[0].BackendObjectReference, gw.Namespace)
	}

	samplingRate := 100.0
	if tracing.SamplingRate != nil {
		samplingRate = float64(*tracing.SamplingRate)
	}

	serviceName := naming.ServiceName(utils.NamespacedName(gw))
	if mergeGateways {
		serviceName = string(gw.Spec.GatewayClassName)
	}

	return &ir.Tracing{
		ServiceName:  serviceName,
		Host:         host,
		Port:         port,
		SamplingRate: samplingRate,
		CustomTags:   tracing.CustomTags,
	}
}

func processMetrics(envoyproxy *egv1a1.EnvoyProxy) *ir.Metrics {
	if envoyproxy == nil ||
		envoyproxy.Spec.Telemetry == nil ||
		envoyproxy.Spec.Telemetry.Metrics == nil {
		return nil
	}
	return &ir.Metrics{
		EnableVirtualHostStats: envoyproxy.Spec.Telemetry.Metrics.EnableVirtualHostStats,
		EnablePerEndpointStats: envoyproxy.Spec.Telemetry.Metrics.EnablePerEndpointStats,
	}
}
