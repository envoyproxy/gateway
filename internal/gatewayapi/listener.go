// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"encoding/base64"
	"fmt"

	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	configv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"

	descpb "google.golang.org/protobuf/types/descriptorpb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

var _ ListenersTranslator = (*Translator)(nil)

type ListenersTranslator interface {
	ProcessListeners(gateways []*GatewayContext, xdsIR XdsIRMap, infraIR InfraIRMap, resources *Resources)
}

func (t *Translator) ProcessListeners(gateways []*GatewayContext, xdsIR XdsIRMap, infraIR InfraIRMap, resources *Resources) {
	t.validateConflictedLayer7Listeners(gateways)
	t.validateConflictedLayer4Listeners(gateways, v1beta1.TCPProtocolType, v1beta1.TLSProtocolType)
	t.validateConflictedLayer4Listeners(gateways, v1beta1.UDPProtocolType)

	// Iterate through all listeners to validate spec
	// and compute status for each, and add valid ones
	// to the Xds IR.

	var corsGlobal *egv1a1.CorsPolicy
	if t.GlobalCorsEnabled {
		for _, corsFilter := range resources.CorsFilters {
			if corsFilter.Spec.Type == egv1a1.GlobalCorsType {
				corsGlobal = &corsFilter.Spec.CorsPolicy
				break
			}
		}
	}

	for _, gateway := range gateways {
		// init IR per gateway
		irKey := irStringKey(gateway.Gateway.Namespace, gateway.Gateway.Name)
		gwXdsIR := &ir.Xds{}
		gwInfraIR := ir.NewInfra()
		gwInfraIR.Proxy.Name = irKey
		gwInfraIR.Proxy.GetProxyMetadata().Labels = GatewayOwnerLabels(gateway.Namespace, gateway.Name)
		if resources.EnvoyProxy != nil {
			gwInfraIR.Proxy.Config = resources.EnvoyProxy
		}

		// save the IR references in the map before the translation starts
		xdsIR[irKey] = gwXdsIR
		infraIR[irKey] = gwInfraIR

		// Infra IR proxy ports must be unique.
		var foundPorts []*protocolPort

		gwXdsIR.AccessLog = processAccessLog(gwInfraIR.Proxy.Config)
		gwXdsIR.Tracing = processTracing(gwInfraIR.Proxy.Config)

		for _, listener := range gateway.listeners {
			// Process protocol & supported kinds
			switch listener.Protocol {
			case v1beta1.TLSProtocolType:
				if listener.TLS != nil {
					switch *listener.TLS.Mode {
					case v1beta1.TLSModePassthrough:
						t.validateAllowedRoutes(listener, KindTLSRoute)
					case v1beta1.TLSModeTerminate:
						t.validateAllowedRoutes(listener, KindTCPRoute)
					default:
						t.validateAllowedRoutes(listener, KindTCPRoute, KindTLSRoute)
					}
				} else {
					t.validateAllowedRoutes(listener, KindTCPRoute, KindTLSRoute)
				}
			case v1beta1.HTTPProtocolType, v1beta1.HTTPSProtocolType:
				t.validateAllowedRoutes(listener, KindHTTPRoute, KindGRPCRoute, KindCustomGRPCRoute)
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
					TLS:     irTLSConfigs(listener.tlsSecrets),
				}
				if listener.Hostname != nil {
					irListener.Hostnames = append(irListener.Hostnames, string(*listener.Hostname))
				} else {
					// Hostname specifies the virtual hostname to match for protocol types that define this concept.
					// When unspecified, all hostnames are matched. This field is ignored for protocols that don’t require hostname based matching.
					// see more https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.Listener.
					irListener.Hostnames = append(irListener.Hostnames, "*")
				}
				if t.GlobalCorsEnabled {
					allowOrigins := make([]*ir.StringMatch, 0)
					for _, allowOrigin := range corsGlobal.AllowOrigins {
						switch {
						case allowOrigin.Exact != nil:
							m := &ir.StringMatch{Exact: allowOrigin.Exact}
							allowOrigins = append(allowOrigins, m)
						case allowOrigin.Prefix != nil:
							m := &ir.StringMatch{Prefix: allowOrigin.Prefix}
							allowOrigins = append(allowOrigins, m)
						default:
							return
						}

					}

					irListener.CorsPolicy = &ir.CorsPolicy{
						AllowOrigins:     allowOrigins,
						AllowCredentials: corsGlobal.AllowCredentials,
						AllowHeaders:     corsGlobal.AllowHeaders,
						AllowMethods:     corsGlobal.AllowMethods,
						ExposeHeaders:    corsGlobal.ExposeHeaders,
						MaxAge:           corsGlobal.MaxAge,
					}
				}

				if resources.GrpcJSONTranscoderFilters != nil {
					for _, grpcJSONTranscoderFilter := range resources.GrpcJSONTranscoderFilters {
						// loop grpcJSONTranscoderFilter.Spec.Services for check
						servicesValid := true
						for _, service := range grpcJSONTranscoderFilter.Spec.Services {
							protoDescBytes, err := base64.StdEncoding.DecodeString(grpcJSONTranscoderFilter.Spec.ProtoDescriptorBin)
							if err != nil {
								continue
							}

							serviceName := service

							// Parse the ProtoDescriptorBin bytes
							fdSet := &descpb.FileDescriptorSet{}
							if err := proto.Unmarshal(protoDescBytes, fdSet); err != nil {
								continue
							}

							// Loop through the parsed services and compare their names to the expected service name
							found := false
							for _, fd := range fdSet.GetFile() {
								for _, svc := range fd.GetService() {
									if serviceName == fmt.Sprintf("%s.%s", fd.GetPackage(), svc.GetName()) {
										found = true
										break
									}
								}
							}

							if !found {
								servicesValid = false
								break
							}
						}

						if servicesValid {
							grpcJSONTranscoderFilterAdd := &ir.GrpcJSONTranscoderFilter{
								ProtoDescriptorBin: grpcJSONTranscoderFilter.Spec.ProtoDescriptorBin,
								Services:           grpcJSONTranscoderFilter.Spec.Services,
								AutoMapping:        grpcJSONTranscoderFilter.Spec.AutoMapping,
								PrintOptions: &ir.PrintOptions{
									AddWhitespace:              grpcJSONTranscoderFilter.Spec.PrintOptions.AddWhitespace,
									AlwaysPrintPrimitiveFields: grpcJSONTranscoderFilter.Spec.PrintOptions.AlwaysPrintPrimitiveFields,
									AlwaysPrintEnumsAsInts:     grpcJSONTranscoderFilter.Spec.PrintOptions.AlwaysPrintEnumsAsInts,
									PreserveProtoFieldNames:    grpcJSONTranscoderFilter.Spec.PrintOptions.PreserveProtoFieldNames,
								},
							}
							irListener.GrpcJSONTranscoderFilters = append(irListener.GrpcJSONTranscoderFilters, grpcJSONTranscoderFilterAdd)
						}
					}
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

func processAccessLog(envoyproxy *configv1a1.EnvoyProxy) *ir.AccessLog {
	if envoyproxy == nil ||
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
			case configv1a1.ProxyAccessLogSinkTypeFile:
				switch accessLog.Format.Type {
				case configv1a1.ProxyAccessLogFormatTypeText:
					al := &ir.TextAccessLog{
						Format: accessLog.Format.Text,
						Path:   sink.File.Path,
					}
					irAccessLog.Text = append(irAccessLog.Text, al)
				case configv1a1.ProxyAccessLogFormatTypeJSON:
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
			case configv1a1.ProxyAccessLogSinkTypeOpenTelemetry:
				if sink.OpenTelemetry == nil {
					continue
				}

				al := &ir.OpenTelemetryAccessLog{
					Port:      uint32(sink.OpenTelemetry.Port),
					Host:      sink.OpenTelemetry.Host,
					Resources: sink.OpenTelemetry.Resources,
				}

				switch accessLog.Format.Type {
				case configv1a1.ProxyAccessLogFormatTypeJSON:
					al.Attributes = accessLog.Format.JSON
				case configv1a1.ProxyAccessLogFormatTypeText:
					al.Text = accessLog.Format.Text
				}

				irAccessLog.OpenTelemetry = append(irAccessLog.OpenTelemetry, al)
			}
		}
	}

	return irAccessLog
}

func processTracing(envoyproxy *configv1a1.EnvoyProxy) *configv1a1.ProxyTracing {
	if envoyproxy == nil || envoyproxy.Spec.Telemetry.Tracing == nil {
		return nil
	}

	return envoyproxy.Spec.Telemetry.Tracing
}
