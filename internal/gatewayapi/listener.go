// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"

	"github.com/google/cel-go/cel"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
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
	t.validateConflictedLayer4Listeners(gateways, gwapiv1.TCPProtocolType)
	t.validateConflictedLayer4Listeners(gateways, gwapiv1.UDPProtocolType)
	if t.MergeGateways {
		t.validateConflictedMergedListeners(gateways)
	}

	// Iterate through all listeners to validate spec
	// and compute status for each, and add valid ones
	// to the Xds IR.
	for _, gateway := range gateways {
		irKey := t.getIRKey(gateway.Gateway)

		if gateway.envoyProxy != nil {
			infraIR[irKey].Proxy.Config = gateway.envoyProxy
		}
		t.processProxyObservability(gateway, xdsIR[irKey], infraIR[irKey].Proxy.Config, resources)

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
				status.SetGatewayListenerStatusCondition(listener.gateway.Gateway,
					listener.listenerStatusIdx,
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
			containerPort := servicePortToContainerPort(int32(listener.Port), gateway.envoyProxy)
			switch listener.Protocol {
			case gwapiv1.HTTPProtocolType, gwapiv1.HTTPSProtocolType:
				irListener := &ir.HTTPListener{
					CoreListenerDetails: ir.CoreListenerDetails{
						Name:     irListenerName(listener),
						Address:  "0.0.0.0",
						Port:     uint32(containerPort),
						Metadata: buildListenerMetadata(listener, gateway),
					},
					TLS: irTLSConfigs(listener.tlsSecrets...),
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
			case gwapiv1.TCPProtocolType, gwapiv1.TLSProtocolType:
				irListener := &ir.TCPListener{
					CoreListenerDetails: ir.CoreListenerDetails{
						Name:    irListenerName(listener),
						Address: "0.0.0.0",
						Port:    uint32(containerPort),
					},

					// Gateway is processed firstly, then ClientTrafficPolicy, then xRoute.
					// TLS field should be added to TCPListener as ClientTrafficPolicy will affect
					// Listener TLS. Then TCPRoute whose TLS should be configured as Terminate just
					// refers to the Listener TLS.
					TLS: irTLSConfigs(listener.tlsSecrets...),
				}
				xdsIR[irKey].TCP = append(xdsIR[irKey].TCP, irListener)
			case gwapiv1.UDPProtocolType:
				irListener := &ir.UDPListener{
					CoreListenerDetails: ir.CoreListenerDetails{
						Name:    irListenerName(listener),
						Address: "0.0.0.0",
						Port:    uint32(containerPort),
					},
				}
				xdsIR[irKey].UDP = append(xdsIR[irKey].UDP, irListener)
			}

			// Add the listener to the Infra IR. Infra IR ports must have a unique port number per layer-4 protocol
			// (TCP or UDP).
			if !containsPort(foundPorts[irKey], servicePort) {
				t.processInfraIRListener(listener, infraIR, irKey, servicePort, containerPort)
				foundPorts[irKey] = append(foundPorts[irKey], servicePort)
			}
		}
	}
}

func buildListenerMetadata(listener *ListenerContext, gateway *GatewayContext) *ir.ResourceMetadata {
	return &ir.ResourceMetadata{
		Kind:        gateway.GetObjectKind().GroupVersionKind().Kind,
		Name:        gateway.GetName(),
		Namespace:   gateway.GetNamespace(),
		Annotations: filterEGPrefix(gateway.GetAnnotations()),
		SectionName: string(listener.Name),
	}
}

func (t *Translator) processProxyObservability(gwCtx *GatewayContext, xdsIR *ir.Xds, envoyProxy *egv1a1.EnvoyProxy, resources *Resources) {
	var err error

	xdsIR.AccessLog, err = t.processAccessLog(envoyProxy, resources)
	if err != nil {
		status.UpdateGatewayListenersNotValidCondition(gwCtx.Gateway, gwapiv1.GatewayReasonInvalid, metav1.ConditionFalse,
			fmt.Sprintf("Invalid access log backendRefs: %v", err))
		return
	}

	xdsIR.Tracing, err = t.processTracing(gwCtx.Gateway, envoyProxy, t.MergeGateways, resources)
	if err != nil {
		status.UpdateGatewayListenersNotValidCondition(gwCtx.Gateway, gwapiv1.GatewayReasonInvalid, metav1.ConditionFalse,
			fmt.Sprintf("Invalid tracing backendRefs: %v", err))
		return
	}

	xdsIR.Metrics, err = t.processMetrics(envoyProxy, resources)
	if err != nil {
		status.UpdateGatewayListenersNotValidCondition(gwCtx.Gateway, gwapiv1.GatewayReasonInvalid, metav1.ConditionFalse,
			fmt.Sprintf("Invalid metrics backendRefs: %v", err))
		return
	}
}

func (t *Translator) processInfraIRListener(listener *ListenerContext, infraIR InfraIRMap, irKey string, servicePort *protocolPort, containerPort int32) {
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
		ContainerPort: containerPort,
	}

	proxyListener := &ir.ProxyListener{
		Name:  irListenerName(listener),
		Ports: []ir.ListenerPort{infraPort},
	}

	infraIR[irKey].Proxy.Listeners = append(infraIR[irKey].Proxy.Listeners, proxyListener)
}

func (t *Translator) processAccessLog(envoyproxy *egv1a1.EnvoyProxy, resources *Resources) (*ir.AccessLog, error) {
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
		}, nil
	}

	if envoyproxy.Spec.Telemetry.AccessLog.Disable {
		return nil, nil
	}

	irAccessLog := &ir.AccessLog{}
	// translate the access log configuration to the IR
	for i, accessLog := range envoyproxy.Spec.Telemetry.AccessLog.Settings {
		var format egv1a1.ProxyAccessLogFormat
		if accessLog.Format != nil {
			format = *accessLog.Format
		} else {
			format = egv1a1.ProxyAccessLogFormat{
				Type: egv1a1.ProxyAccessLogFormatTypeText,
				// Empty text format means default format
			}
		}

		var (
			validExprs []string
			errs       []error
		)
		for _, expr := range accessLog.Matches {
			if !validCELExpression(expr) {
				errs = append(errs, fmt.Errorf("invalid CEL expression: %s", expr))
				continue
			}
			validExprs = append(validExprs, expr)
		}
		if len(errs) > 0 {
			return nil, utilerrors.NewAggregate(errs)
		}

		for j, sink := range accessLog.Sinks {
			switch sink.Type {
			case egv1a1.ProxyAccessLogSinkTypeFile:
				if sink.File == nil {
					continue
				}

				switch format.Type {
				case egv1a1.ProxyAccessLogFormatTypeText:
					al := &ir.TextAccessLog{
						Format:     format.Text,
						Path:       sink.File.Path,
						CELMatches: validExprs,
					}
					irAccessLog.Text = append(irAccessLog.Text, al)
				case egv1a1.ProxyAccessLogFormatTypeJSON:
					if len(format.JSON) == 0 {
						// TODO: use a default JSON format if not specified?
						continue
					}

					al := &ir.JSONAccessLog{
						JSON:       format.JSON,
						Path:       sink.File.Path,
						CELMatches: validExprs,
					}
					irAccessLog.JSON = append(irAccessLog.JSON, al)
				}
			case egv1a1.ProxyAccessLogSinkTypeALS:
				if sink.ALS == nil {
					continue
				}

				var logName string
				if sink.ALS.LogName != nil {
					logName = *sink.ALS.LogName
				} else {
					logName = fmt.Sprintf("%s/%s", envoyproxy.Namespace, envoyproxy.Name)
				}

				// TODO: how to get authority from the backendRefs?
				ds, traffic, err := t.processBackendRefs(sink.ALS.BackendCluster, envoyproxy.Namespace, resources, envoyproxy)
				if err != nil {
					return nil, err
				}

				al := &ir.ALSAccessLog{
					LogName: logName,
					Destination: ir.RouteDestination{
						Name:     fmt.Sprintf("accesslog_als_%d_%d", i, j), // TODO: rename this, so that we can share backend with tracing?
						Settings: ds,
					},
					Traffic:    traffic,
					Type:       sink.ALS.Type,
					CELMatches: validExprs,
				}

				if al.Type == egv1a1.ALSEnvoyProxyAccessLogTypeHTTP && sink.ALS.HTTP != nil {
					http := &ir.ALSAccessLogHTTP{
						RequestHeaders:   sink.ALS.HTTP.RequestHeaders,
						ResponseHeaders:  sink.ALS.HTTP.ResponseHeaders,
						ResponseTrailers: sink.ALS.HTTP.ResponseTrailers,
					}
					al.HTTP = http
				}

				switch format.Type {
				case egv1a1.ProxyAccessLogFormatTypeJSON:
					al.Attributes = format.JSON
				case egv1a1.ProxyAccessLogFormatTypeText:
					al.Text = format.Text
				}

				irAccessLog.ALS = append(irAccessLog.ALS, al)
			case egv1a1.ProxyAccessLogSinkTypeOpenTelemetry:
				if sink.OpenTelemetry == nil {
					continue
				}

				// TODO: how to get authority from the backendRefs?
				ds, traffic, err := t.processBackendRefs(sink.OpenTelemetry.BackendCluster, envoyproxy.Namespace, resources, envoyproxy)
				if err != nil {
					return nil, err
				}
				// TODO: remove support for Host/Port in v1.2
				al := &ir.OpenTelemetryAccessLog{
					CELMatches: validExprs,
					Resources:  sink.OpenTelemetry.Resources,
					Destination: ir.RouteDestination{
						Name:     fmt.Sprintf("accesslog_otel_%d_%d", i, j), // TODO: rename this, so that we can share backend with tracing?
						Settings: ds,
					},
					Traffic: traffic,
				}

				if len(ds) == 0 {
					// fallback to host and port
					var host string
					var port uint32
					if sink.OpenTelemetry.Host != nil {
						host, port = *sink.OpenTelemetry.Host, uint32(sink.OpenTelemetry.Port)
					}
					al.Destination.Settings = destinationSettingFromHostAndPort(host, port)
					al.Authority = host
				}

				switch format.Type {
				case egv1a1.ProxyAccessLogFormatTypeJSON:
					al.Attributes = format.JSON
				case egv1a1.ProxyAccessLogFormatTypeText:
					al.Text = format.Text
				}

				irAccessLog.OpenTelemetry = append(irAccessLog.OpenTelemetry, al)
			}
		}
	}

	return irAccessLog, nil
}

func (t *Translator) processTracing(gw *gwapiv1.Gateway, envoyproxy *egv1a1.EnvoyProxy, mergeGateways bool, resources *Resources) (*ir.Tracing, error) {
	if envoyproxy == nil ||
		envoyproxy.Spec.Telemetry == nil ||
		envoyproxy.Spec.Telemetry.Tracing == nil {
		return nil, nil
	}
	tracing := envoyproxy.Spec.Telemetry.Tracing

	// TODO: how to get authority from the backendRefs?
	ds, traffic, err := t.processBackendRefs(tracing.Provider.BackendCluster, envoyproxy.Namespace, resources, envoyproxy)
	if err != nil {
		return nil, err
	}

	var authority string

	// fallback to host and port
	// TODO: remove support for Host/Port in v1.2
	if len(ds) == 0 {
		var host string
		var port uint32
		if tracing.Provider.Host != nil {
			host, port = *tracing.Provider.Host, uint32(tracing.Provider.Port)
		}
		ds = destinationSettingFromHostAndPort(host, port)
		authority = host
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
		Authority:    authority,
		ServiceName:  serviceName,
		SamplingRate: samplingRate,
		CustomTags:   tracing.CustomTags,
		Destination: ir.RouteDestination{
			Name:     "tracing", // TODO: rename this, so that we can share backend with accesslog?
			Settings: ds,
		},
		Provider: tracing.Provider,
		Traffic:  traffic,
	}, nil
}

func (t *Translator) processMetrics(envoyproxy *egv1a1.EnvoyProxy, resources *Resources) (*ir.Metrics, error) {
	if envoyproxy == nil ||
		envoyproxy.Spec.Telemetry == nil ||
		envoyproxy.Spec.Telemetry.Metrics == nil {
		return nil, nil
	}

	for _, sink := range envoyproxy.Spec.Telemetry.Metrics.Sinks {
		if sink.OpenTelemetry == nil {
			continue
		}

		_, _, err := t.processBackendRefs(sink.OpenTelemetry.BackendCluster, envoyproxy.Namespace, resources, envoyproxy)
		if err != nil {
			return nil, err
		}
	}

	return &ir.Metrics{
		EnableVirtualHostStats: envoyproxy.Spec.Telemetry.Metrics.EnableVirtualHostStats,
		EnablePerEndpointStats: envoyproxy.Spec.Telemetry.Metrics.EnablePerEndpointStats,
	}, nil
}

func (t *Translator) processBackendRefs(backendCluster egv1a1.BackendCluster, namespace string, resources *Resources, envoyProxy *egv1a1.EnvoyProxy) ([]*ir.DestinationSetting, *ir.TrafficFeatures, error) {
	traffic, err := translateTrafficFeatures(backendCluster.BackendSettings)
	if err != nil {
		return nil, nil, err
	}
	result := make([]*ir.DestinationSetting, 0, len(backendCluster.BackendRefs))
	for _, ref := range backendCluster.BackendRefs {
		ns := NamespaceDerefOr(ref.Namespace, namespace)
		kind := KindDerefOr(ref.Kind, KindService)
		if kind != KindService {
			return nil, nil, errors.New("only service kind is supported for backendRefs")
		}
		if err := validateBackendService(ref.BackendObjectReference, resources, ns, corev1.ProtocolTCP); err != nil {
			return nil, nil, err
		}

		ds := t.processServiceDestinationSetting(ref.BackendObjectReference, ns, ir.TCP, resources, envoyProxy)
		result = append(result, ds)
	}
	if len(result) == 0 {
		return nil, traffic, nil
	}
	return result, traffic, nil
}

func destinationSettingFromHostAndPort(host string, port uint32) []*ir.DestinationSetting {
	return []*ir.DestinationSetting{
		{
			Weight:    ptr.To[uint32](1),
			Protocol:  ir.GRPC,
			Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(host, port)},
		},
	}
}

var celEnv, _ = cel.NewEnv()

func validCELExpression(expr string) bool {
	_, issue := celEnv.Parse(expr)
	return issue.Err() == nil
}
