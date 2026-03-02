// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/netip"
	"strconv"
	"strings"

	"github.com/google/cel-go/cel"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/utils/naming"
	netutils "github.com/envoyproxy/gateway/internal/utils/net"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

var _ ListenersTranslator = (*Translator)(nil)

type ListenersTranslator interface {
	ProcessListeners(gateways []*GatewayContext, xdsIR resource.XdsIRMap, infraIR resource.InfraIRMap, resources *resource.Resources)
}

func (t *Translator) ProcessListeners(gateways []*GatewayContext, xdsIR resource.XdsIRMap, infraIR resource.InfraIRMap, resources *resource.Resources) {
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
		t.processProxyReadyListener(xdsIR[irKey], gateway.envoyProxy)
		t.processProxyObservability(gateway, xdsIR[irKey], infraIR[irKey].Proxy, resources)

		for _, listener := range gateway.listeners {
			// Process protocol & supported kinds
			switch listener.Protocol {
			case gwapiv1.TLSProtocolType:
				if listener.TLS != nil {
					switch *listener.TLS.Mode {
					case gwapiv1.TLSModePassthrough:
						t.validateAllowedRoutes(listener, resource.KindTLSRoute)
					case gwapiv1.TLSModeTerminate:
						t.validateAllowedRoutes(listener, resource.KindTCPRoute, resource.KindTLSRoute)
					default:
						t.validateAllowedRoutes(listener, resource.KindTCPRoute, resource.KindTLSRoute)
					}
				} else {
					t.validateAllowedRoutes(listener, resource.KindTCPRoute, resource.KindTLSRoute)
				}
			case gwapiv1.HTTPProtocolType, gwapiv1.HTTPSProtocolType:
				t.validateAllowedRoutes(listener, resource.KindHTTPRoute, resource.KindGRPCRoute)
			case gwapiv1.TCPProtocolType:
				t.validateAllowedRoutes(listener, resource.KindTCPRoute)
			case gwapiv1.UDPProtocolType:
				t.validateAllowedRoutes(listener, resource.KindUDPRoute)
			default:
				listener.SetSupportedKinds()
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

			address := netutils.IPv4ListenerAddress
			ipFamily := getEnvoyIPFamily(gateway.envoyProxy)
			if ipFamily != nil && (*ipFamily == egv1a1.IPv6 || *ipFamily == egv1a1.DualStack) {
				address = netutils.IPv6ListenerAddress
			}

			// Add the listener to the Xds IR
			servicePort := &protocolPort{protocol: listener.Protocol, port: listener.Port}
			containerPort := t.servicePortToContainerPort(listener.Port, gateway.envoyProxy)
			switch listener.Protocol {
			case gwapiv1.HTTPProtocolType, gwapiv1.HTTPSProtocolType:
				irListener := &ir.HTTPListener{
					CoreListenerDetails: ir.CoreListenerDetails{
						Name:         irListenerName(listener),
						Address:      address,
						Port:         uint32(containerPort),
						ExternalPort: uint32(listener.Port),
						Metadata:     buildListenerMetadata(listener, gateway),
						IPFamily:     ipFamily,
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
					// When unspecified, all hostnames are matched. This field is ignored for protocols that don't require hostname based matching.
					// see more https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/gwapiv1.Listener.
					irListener.Hostnames = append(irListener.Hostnames, "*")
				}
				irListener.PreserveRouteOrder = getPreserveRouteOrder(gateway.envoyProxy)
				irListener.RequestID = getRequestIDExtensionAction(gateway.envoyProxy)
				xdsIR[irKey].HTTP = append(xdsIR[irKey].HTTP, irListener)
				// Store the HTTPListener IR in the listener context for use in the overlapping TLS config check.
				listener.httpIR = irListener
			case gwapiv1.TCPProtocolType, gwapiv1.TLSProtocolType:
				irListener := &ir.TCPListener{
					CoreListenerDetails: ir.CoreListenerDetails{
						Name:         irListenerName(listener),
						Address:      address,
						Port:         uint32(containerPort),
						ExternalPort: uint32(listener.Port),
						Metadata:     buildListenerMetadata(listener, gateway),
						IPFamily:     ipFamily,
					},

					// Gateway is processed firstly, then ClientTrafficPolicy, then xRoute.
					// TLS field should be added to TCPListener as ClientTrafficPolicy will affect
					// Listener TLS. Then TCPRoute whose TLS should be configured as Terminate just
					// refers to the Listener TLS.
					TLS: irTLSConfigsForTCPListener(listener.tlsSecrets...),
				}
				xdsIR[irKey].TCP = append(xdsIR[irKey].TCP, irListener)
			case gwapiv1.UDPProtocolType:
				irListener := &ir.UDPListener{
					CoreListenerDetails: ir.CoreListenerDetails{
						Name:         irListenerName(listener),
						Address:      address,
						Port:         uint32(containerPort),
						ExternalPort: uint32(listener.Port),
						Metadata:     buildListenerMetadata(listener, gateway),
						IPFamily:     ipFamily,
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

	t.checkOverlappingTLSConfig(gateways)
}

// checkOverlappingTLSConfig checks for overlapping hostnames and certificates between listeners and sets
// the `OverlappingTLSConfig` condition if there are overlapping hostnames or certificates.
func (t *Translator) checkOverlappingTLSConfig(gateways []*GatewayContext) {
	// If merging gateways, check overlapping hostnames and certificates between listeners in all merged gateways.
	if t.MergeGateways {
		httpsListeners := []*ListenerContext{}
		for _, gateway := range gateways {
			for _, listener := range gateway.listeners {
				if listener.Protocol == gwapiv1.HTTPSProtocolType {
					httpsListeners = append(httpsListeners, listener)
				}
			}
		}
		// Note: order of processing matters here.
		// According to the Gateway API spec, If both hostname and certificate overlap,
		// the controller SHOULD set the "OverlappingCertificates" Reason.
		checkOverlappingHostnames(httpsListeners)
		checkOverlappingCertificates(httpsListeners)
	} else {
		// Check overlapping hostnames and certificates between listeners in each gateway.
		for _, gateway := range gateways {
			httpsListeners := []*ListenerContext{}
			for _, listener := range gateway.listeners {
				if listener.Protocol == gwapiv1.HTTPSProtocolType {
					httpsListeners = append(httpsListeners, listener)
				}
			}
			// Note: order of processing matters here.
			// According to the Gateway API spec, If both hostname and certificate overlap,
			// the controller SHOULD set the "OverlappingCertificates" Reason.
			checkOverlappingHostnames(httpsListeners)
			checkOverlappingCertificates(httpsListeners)
		}
	}
}

// checkOverlappingHostnames checks for overlapping hostnames between HTTPS listeners and sets
// the `OverlappingTLSConfig` condition if there are overlapping hostnames.
func checkOverlappingHostnames(httpsListeners []*ListenerContext) {
	type overlappingListener struct {
		gateway1  *GatewayContext
		gateway2  *GatewayContext
		listener1 string
		listener2 string
		hostname1 string
		hostname2 string
	}
	overlappingListeners := make([]*overlappingListener, len(httpsListeners))
	for i := range httpsListeners {
		if overlappingListeners[i] != nil {
			continue
		}
		for j := i + 1; j < len(httpsListeners); j++ {
			if overlappingListeners[j] != nil {
				continue
			}
			if httpsListeners[i].Port != httpsListeners[j].Port {
				continue
			}
			if areOverlappingHostnames(httpsListeners[i].Hostname, httpsListeners[j].Hostname) {
				// Overlapping listeners can be more than two, we only report the first two for simplicity.
				overlappingListeners[i] = &overlappingListener{
					gateway1:  httpsListeners[i].gateway,
					gateway2:  httpsListeners[j].gateway,
					listener1: string(httpsListeners[i].Name),
					listener2: string(httpsListeners[j].Name),
					hostname1: string(ptr.Deref(httpsListeners[i].Hostname, "")),
					hostname2: string(ptr.Deref(httpsListeners[j].Hostname, "")),
				}
				overlappingListeners[j] = &overlappingListener{
					gateway1:  httpsListeners[j].gateway,
					gateway2:  httpsListeners[i].gateway,
					listener1: string(httpsListeners[j].Name),
					listener2: string(httpsListeners[i].Name),
					hostname1: string(ptr.Deref(httpsListeners[j].Hostname, "")),
					hostname2: string(ptr.Deref(httpsListeners[i].Hostname, "")),
				}
			}
		}
	}

	for i, listener := range httpsListeners {
		if overlappingListeners[i] != nil {
			var message string
			gateway1 := overlappingListeners[i].gateway1
			gateway2 := overlappingListeners[i].gateway2
			if gateway1.Name == gateway2.Name &&
				gateway1.Namespace == gateway2.Namespace {
				message = fmt.Sprintf(
					"The hostname %s overlaps with the hostname %s in listener %s. ALPN will default to HTTP/1.1 to prevent HTTP/2 connection coalescing, unless explicitly configured via ClientTrafficPolicy",
					overlappingListeners[i].hostname1,
					overlappingListeners[i].hostname2,
					overlappingListeners[i].listener2,
				)
			} else {
				message = fmt.Sprintf(
					"The hostname %s overlaps with the hostname %s in listener %s of gateway %s. ALPN will default to HTTP/1.1 to prevent HTTP/2 connection coalescing, unless explicitly configured via ClientTrafficPolicy",
					overlappingListeners[i].hostname1,
					overlappingListeners[i].hostname2,
					overlappingListeners[i].listener2,
					gateway2.GetName(),
				)
			}

			listener.SetCondition(
				gwapiv1.ListenerConditionOverlappingTLSConfig,
				metav1.ConditionTrue,
				gwapiv1.ListenerReasonOverlappingHostnames,
				message,
			)
			if listener.httpIR != nil {
				listener.httpIR.TLSOverlaps = true
			}
		}
	}
}

// checkOverlappingCertificates checks for overlapping certificates SANs between HTTPSlisteners and sets
// the `OverlappingTLSConfig` condition if there are overlapping certificates.
func checkOverlappingCertificates(httpsListeners []*ListenerContext) {
	type overlappingListener struct {
		gateway1  *GatewayContext
		gateway2  *GatewayContext
		listener1 string
		listener2 string
		san1      string
		san2      string
	}

	overlappingListeners := make([]*overlappingListener, len(httpsListeners))
	for i := range httpsListeners {
		if overlappingListeners[i] != nil {
			continue
		}
		for j := i + 1; j < len(httpsListeners); j++ {
			if overlappingListeners[j] != nil {
				continue
			}
			if httpsListeners[i].Port != httpsListeners[j].Port {
				continue
			}

			overlappingCertificate := isOverlappingCertificate(httpsListeners[i].certDNSNames, httpsListeners[j].certDNSNames)
			if overlappingCertificate != nil {
				// Overlapping listeners can be more than two, we only report the first two for simplicity.
				overlappingListeners[i] = &overlappingListener{
					gateway1:  httpsListeners[i].gateway,
					gateway2:  httpsListeners[j].gateway,
					listener1: string(httpsListeners[i].Name),
					listener2: string(httpsListeners[j].Name),
					san1:      overlappingCertificate.san1,
					san2:      overlappingCertificate.san2,
				}
				overlappingListeners[j] = &overlappingListener{
					gateway1:  httpsListeners[j].gateway,
					gateway2:  httpsListeners[i].gateway,
					listener1: string(httpsListeners[j].Name),
					listener2: string(httpsListeners[i].Name),
					san1:      overlappingCertificate.san2,
					san2:      overlappingCertificate.san1,
				}
			}
		}
	}

	for i, listener := range httpsListeners {
		if overlappingListeners[i] != nil {
			var message string
			gateway1 := overlappingListeners[i].gateway1
			gateway2 := overlappingListeners[i].gateway2
			if gateway1.Name == gateway2.Name &&
				gateway1.Namespace == gateway2.Namespace {
				message = fmt.Sprintf(
					"The certificate SAN %s overlaps with the certificate SAN %s in listener %s. ALPN will default to HTTP/1.1 to prevent HTTP/2 connection coalescing, unless explicitly configured via ClientTrafficPolicy",
					overlappingListeners[i].san1,
					overlappingListeners[i].san2,
					overlappingListeners[i].listener2,
				)
			} else {
				message = fmt.Sprintf(
					"The certificate SAN %s overlaps with the certificate SAN %s in listener %s of gateway %s. ALPN will default to HTTP/1.1 to prevent HTTP/2 connection coalescing, unless explicitly configured via ClientTrafficPolicy",
					overlappingListeners[i].san1,
					overlappingListeners[i].san2,
					overlappingListeners[i].listener2,
					gateway2.GetName(),
				)
			}

			listener.SetCondition(
				gwapiv1.ListenerConditionOverlappingTLSConfig,
				metav1.ConditionTrue,
				gwapiv1.ListenerReasonOverlappingCertificates,
				message)
			if listener.httpIR != nil {
				listener.httpIR.TLSOverlaps = true
			}
		}
	}
}

type overlappingCertificate struct {
	san1 string
	san2 string
}

func isOverlappingCertificate(cert1DNSNames, cert2DNSNames []string) *overlappingCertificate {
	for _, dns1 := range cert1DNSNames {
		for _, dns2 := range cert2DNSNames {
			if areOverlappingHostnames(ptr.To(gwapiv1.Hostname(dns1)), ptr.To(gwapiv1.Hostname(dns2))) {
				return &overlappingCertificate{
					san1: dns1,
					san2: dns2,
				}
			}
		}
	}
	return nil
}

func areOverlappingHostnames(this, other *gwapiv1.Hostname) bool {
	if this == nil || other == nil {
		return true
	}
	return hostnameMatchesWithOther(this, other) || hostnameMatchesWithOther(other, this)
}

// hostnameMatchesWithOther returns true if this hostname matches other hostname.
// Assumes that hostnames will either be fully qualified or a wildcard hostname prefixed with a single wildcard.
// E.g. "*.*.example.com" is not valid.
func hostnameMatchesWithOther(this, other *gwapiv1.Hostname) bool {
	thisString := string(*this)
	otherString := string(*other)
	if hasWildcardPrefix(other) && !hasWildcardPrefix(this) {
		return strings.HasSuffix(thisString, otherString[1:]) &&
			!strings.Contains(strings.TrimSuffix(thisString, otherString[1:]), ".") // not a subdomain
	}
	return thisString == otherString
}

func hasWildcardPrefix(h *gwapiv1.Hostname) bool {
	if h == nil {
		return false
	}
	return len(string(*h)) > 1 && string(*h)[0] == '*'
}

func buildListenerMetadata(listener *ListenerContext, gateway *GatewayContext) *ir.ResourceMetadata {
	return &ir.ResourceMetadata{
		Kind:        gateway.GetObjectKind().GroupVersionKind().Kind,
		Name:        gateway.GetName(),
		Namespace:   gateway.GetNamespace(),
		Annotations: ir.MapToSlice(filterEGPrefix(gateway.GetAnnotations())),
		SectionName: string(listener.Name),
	}
}

func (t *Translator) processProxyReadyListener(xdsIR *ir.Xds, envoyProxy *egv1a1.EnvoyProxy) {
	var (
		ipFamily = egv1a1.IPv4
		address  = netutils.IPv4ListenerAddress
	)

	if envoyProxy != nil && envoyProxy.Spec.IPFamily != nil {
		ipFamily = *envoyProxy.Spec.IPFamily
	}
	if ipFamily == egv1a1.IPv6 || ipFamily == egv1a1.DualStack {
		address = netutils.IPv6ListenerAddress
	}

	xdsIR.ReadyListener = &ir.ReadyListener{
		Address:  address,
		Port:     uint32(bootstrap.EnvoyReadinessPort),
		Path:     bootstrap.EnvoyReadinessPath,
		IPFamily: ipFamily,
	}
}

func (t *Translator) processProxyObservability(gwCtx *GatewayContext, xdsIR *ir.Xds, proxyInfra *ir.ProxyInfra, resources *resource.Resources) {
	var err error
	envoyProxy := proxyInfra.Config

	xdsIR.AccessLog, err = t.processAccessLog(envoyProxy, resources)
	if err != nil {
		status.UpdateGatewayStatusNotAccepted(gwCtx.Gateway, gwapiv1.GatewayReasonInvalidParameters,
			fmt.Sprintf("Invalid access log backendRefs in the referenced EnvoyProxy: %v", err))
		return
	}

	xdsIR.Tracing, err = t.processTracing(gwCtx.Gateway, envoyProxy, t.MergeGateways, resources)
	if err != nil {
		status.UpdateGatewayStatusNotAccepted(gwCtx.Gateway, gwapiv1.GatewayReasonInvalidParameters,
			fmt.Sprintf("Invalid tracing backendRefs in the referenced EnvoyProxy: %v", err))
		return
	}

	var resolvedSinks []ir.ResolvedMetricSink
	xdsIR.Metrics, resolvedSinks, err = t.processMetrics(envoyProxy, resources)
	if err != nil {
		status.UpdateGatewayStatusNotAccepted(gwCtx.Gateway, gwapiv1.GatewayReasonInvalidParameters,
			fmt.Sprintf("Invalid metrics backendRefs in the referenced EnvoyProxy: %v", err))
		return
	}
	proxyInfra.ResolvedMetricSinks = resolvedSinks
}

func (t *Translator) processInfraIRListener(listener *ListenerContext, infraIR resource.InfraIRMap, irKey string, servicePort *protocolPort, containerPort int32) {
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

func (t *Translator) processAccessLog(envoyproxy *egv1a1.EnvoyProxy, resources *resource.Resources) (*ir.AccessLog, error) {
	if envoyproxy == nil ||
		envoyproxy.Spec.Telemetry == nil ||
		envoyproxy.Spec.Telemetry.AccessLog == nil ||
		(!ptr.Deref(envoyproxy.Spec.Telemetry.AccessLog.Disable, false) && len(envoyproxy.Spec.Telemetry.AccessLog.Settings) == 0) {
		// use the default access log
		return &ir.AccessLog{
			JSON: []*ir.JSONAccessLog{
				{
					Path: "/dev/stdout",
				},
			},
		}, nil
	}
	if ptr.Deref(envoyproxy.Spec.Telemetry.AccessLog.Disable, false) {
		return nil, nil
	}

	irAccessLog := &ir.AccessLog{}
	// translate the access log configuration to the IR
	for i, accessLog := range envoyproxy.Spec.Telemetry.AccessLog.Settings {
		var accessLogType *ir.ProxyAccessLogType
		if accessLog.Type != nil {
			switch *accessLog.Type {
			case egv1a1.ProxyAccessLogTypeRoute:
				accessLogType = ptr.To(ir.ProxyAccessLogTypeRoute)
			case egv1a1.ProxyAccessLogTypeListener:
				accessLogType = ptr.To(ir.ProxyAccessLogTypeListener)
			}
		}

		var format egv1a1.ProxyAccessLogFormat
		if accessLog.Format != nil {
			format = *accessLog.Format
		} else {
			defaultType := egv1a1.ProxyAccessLogFormatTypeJSON
			format = egv1a1.ProxyAccessLogFormat{
				Type: &defaultType,
				// Empty means default format
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

		if len(accessLog.Sinks) == 0 {
			al := &ir.JSONAccessLog{
				JSON:       ir.MapToSlice(format.JSON),
				CELMatches: validExprs,
				LogType:    accessLogType,
				Path:       "/dev/stdout",
			}
			irAccessLog.JSON = append(irAccessLog.JSON, al)
		}

		for j, sink := range accessLog.Sinks {
			switch sink.Type {
			case egv1a1.ProxyAccessLogSinkTypeFile:
				if sink.File == nil {
					continue
				}

				if format.Type != nil && *format.Type == egv1a1.ProxyAccessLogFormatTypeText {
					al := &ir.TextAccessLog{
						Format:     format.Text,
						Path:       sink.File.Path,
						CELMatches: validExprs,
						LogType:    accessLogType,
					}
					irAccessLog.Text = append(irAccessLog.Text, al)
				} else {
					// Default to JSON format if type is nil or JSON
					al := &ir.JSONAccessLog{
						JSON:       ir.MapToSlice(format.JSON),
						Path:       sink.File.Path,
						CELMatches: validExprs,
						LogType:    accessLogType,
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

				// TODO: rename this, so that we can share backend with tracing?
				destName := fmt.Sprintf("accesslog_als_%d_%d", i, j)
				settingName := irDestinationSettingName(destName, -1)
				// TODO: how to get authority from the backendRefs?
				ds, traffic, err := t.processBackendRefs(settingName, sink.ALS.BackendCluster, envoyproxy.Namespace, resources, envoyproxy)
				if err != nil {
					return nil, err
				}
				// ALS should always use GRPC protocol. Setting this adds http2 by default to the cluster.
				for _, setting := range ds {
					setting.Protocol = ir.GRPC
				}

				al := &ir.ALSAccessLog{
					LogName: logName,
					Destination: ir.RouteDestination{
						Name:     destName,
						Settings: ds,
						Metadata: buildResourceMetadata(envoyproxy, nil),
					},
					Traffic:    traffic,
					Type:       sink.ALS.Type,
					CELMatches: validExprs,
					LogType:    accessLogType,
				}

				if al.Type == egv1a1.ALSEnvoyProxyAccessLogTypeHTTP && sink.ALS.HTTP != nil {
					http := &ir.ALSAccessLogHTTP{
						RequestHeaders:   sink.ALS.HTTP.RequestHeaders,
						ResponseHeaders:  sink.ALS.HTTP.ResponseHeaders,
						ResponseTrailers: sink.ALS.HTTP.ResponseTrailers,
					}
					al.HTTP = http
				}
				if format.Type != nil && *format.Type == egv1a1.ProxyAccessLogFormatTypeText {
					al.Text = format.Text
				} else {
					// Default to JSON format if type is nil or JSON
					al.Attributes = ir.MapToSlice(format.JSON)
				}

				irAccessLog.ALS = append(irAccessLog.ALS, al)
			case egv1a1.ProxyAccessLogSinkTypeOpenTelemetry:
				if sink.OpenTelemetry == nil {
					continue
				}

				// TODO: rename this, so that we can share backend with tracing?
				destName := fmt.Sprintf("accesslog_otel_%d_%d", i, j)
				settingName := irDestinationSettingName(destName, -1)
				ds, traffic, err := t.processBackendRefs(settingName, sink.OpenTelemetry.BackendCluster, envoyproxy.Namespace, resources, envoyproxy)
				if err != nil {
					return nil, err
				}
				// TODO: update when OTLP/HTTP is completely supported (logs, traces, metrics)
				for _, d := range ds {
					d.Protocol = ir.GRPC
				}

				al := &ir.OpenTelemetryAccessLog{
					CELMatches:         validExprs,
					ResourceAttributes: ir.MapToSlice(sink.OpenTelemetry.GetResourceAttributes()),
					Headers:            sink.OpenTelemetry.Headers,
					Authority:          getAuthorityFromDestination(ds),
					Destination: ir.RouteDestination{
						Name:     destName,
						Settings: ds,
						Metadata: buildResourceMetadata(envoyproxy, nil),
					},
					Traffic: traffic,
					LogType: accessLogType,
				}

				if len(ds) == 0 {
					// fallback to host and port
					var host string
					var port uint32
					if sink.OpenTelemetry.Host != nil {
						host, port = *sink.OpenTelemetry.Host, uint32(sink.OpenTelemetry.Port)
					}
					al.Destination.Settings = destinationSettingFromHostAndPort(settingName, host, port)
					al.Authority = host
				}

				// For OpenTelemetry, text (body) and attributes can be used together.
				// When format.Type is nil, both text and json from format can be used.
				if format.Type == nil || *format.Type == egv1a1.ProxyAccessLogFormatTypeText {
					al.Text = format.Text
				}
				if format.Type == nil || *format.Type == egv1a1.ProxyAccessLogFormatTypeJSON {
					al.Attributes = ir.MapToSlice(format.JSON)
				}

				irAccessLog.OpenTelemetry = append(irAccessLog.OpenTelemetry, al)
			}
		}
	}
	return irAccessLog, nil
}

func (t *Translator) processTracing(gw *gwapiv1.Gateway, envoyproxy *egv1a1.EnvoyProxy,
	mergeGateways bool, resources *resource.Resources,
) (*ir.Tracing, error) {
	if envoyproxy == nil ||
		envoyproxy.Spec.Telemetry == nil ||
		envoyproxy.Spec.Telemetry.Tracing == nil {
		return nil, nil
	}
	tracing := envoyproxy.Spec.Telemetry.Tracing

	// TODO: rename this, so that we can share backend with accesslog?
	destName := "tracing"
	settingName := irDestinationSettingName(destName, -1)
	ds, traffic, err := t.processBackendRefs(settingName, tracing.Provider.BackendCluster, envoyproxy.Namespace, resources, envoyproxy)
	if err != nil {
		return nil, err
	}
	if tracing.Provider.Type == egv1a1.TracingProviderTypeOpenTelemetry {
		// TODO: update when OTLP/HTTP is completely supported (logs, traces, metrics)
		for _, d := range ds {
			d.Protocol = ir.GRPC
		}
	}

	authority := getAuthorityFromDestination(ds)

	// fallback to host and port
	// TODO: remove support for Host/Port in v1.2
	if len(ds) == 0 {
		var host string
		var port uint32
		if tracing.Provider.Host != nil {
			host, port = *tracing.Provider.Host, uint32(tracing.Provider.Port)
		}
		ds = destinationSettingFromHostAndPort(settingName, host, port)
		authority = host
	}

	serviceName := naming.ServiceName(utils.NamespacedName(gw))
	if mergeGateways {
		serviceName = string(gw.Spec.GatewayClassName)
	}

	// Use configured service name if provided
	if tracing.Provider.ServiceName != nil {
		serviceName = *tracing.Provider.ServiceName
	}

	return &ir.Tracing{
		Authority:          authority,
		ServiceName:        serviceName,
		SamplingRate:       proxySamplingRate(tracing),
		CustomTags:         ir.CustomTagMapToSlice(tracing.CustomTags),
		Tags:               ir.MapToSlice(tracing.Tags),
		ResourceAttributes: ir.MapToSlice(getOpenTelemetryTracingResourceAttributes(&tracing.Provider)),
		Destination: ir.RouteDestination{
			Name:     destName,
			Settings: ds,
			Metadata: buildResourceMetadata(envoyproxy, nil),
		},
		Provider: tracing.Provider,
		Traffic:  traffic,
		Headers:  getOpenTelemetryTracingHeaders(&tracing.Provider),
		SpanName: tracing.SpanName,
	}, nil
}

func proxySamplingRate(tracing *egv1a1.ProxyTracing) float64 {
	rate := 100.0
	if tracing.SamplingRate != nil {
		rate = float64(*tracing.SamplingRate)
	} else if tracing.SamplingFraction != nil {
		numerator := float64(tracing.SamplingFraction.Numerator)
		denominator := ptr.Deref(tracing.SamplingFraction.Denominator, 100)

		rate = numerator * 100 / float64(denominator)
		// Identifies a percentage, in the range [0.0, 100.0]
		rate = math.Max(0, rate)
		rate = math.Min(100, rate)
	}
	return rate
}

// getAuthorityFromDestination extracts the gRPC authority from a destination setting.
// Priority: SNI > hostname > Service/Backend metadata.
func getAuthorityFromDestination(ds []*ir.DestinationSetting) string {
	if len(ds) == 0 {
		return ""
	}
	dest := ds[0]

	// Priority 1: SNI from TLS config
	if dest.TLS != nil && dest.TLS.SNI != nil {
		return *dest.TLS.SNI
	}

	// Priority 2: Endpoint host if it's a hostname (not IP)
	if len(dest.Endpoints) > 0 {
		host := dest.Endpoints[0].Host
		if _, err := netip.ParseAddr(host); err != nil {
			// Not an IP - use as authority
			return host
		}

		// Priority 3: Derive from metadata when endpoint is an IP
		if dest.Metadata != nil && dest.Metadata.Name != "" {
			if dest.Metadata.Namespace != "" {
				if dest.Metadata.Kind == resource.KindService {
					return fmt.Sprintf("%s.%s.svc", dest.Metadata.Name, dest.Metadata.Namespace)
				}
				return fmt.Sprintf("%s.%s", dest.Metadata.Name, dest.Metadata.Namespace)
			}
			return dest.Metadata.Name
		}
	}
	// Don't set authority to an IP - let Envoy use defaults
	return ""
}

func getOpenTelemetryTracingHeaders(provider *egv1a1.TracingProvider) []gwapiv1.HTTPHeader {
	if provider != nil && provider.OpenTelemetry != nil {
		return provider.OpenTelemetry.Headers
	}
	return nil
}

func getOpenTelemetryTracingResourceAttributes(provider *egv1a1.TracingProvider) map[string]string {
	if provider != nil && provider.OpenTelemetry != nil {
		return provider.OpenTelemetry.ResourceAttributes
	}
	return nil
}

func (t *Translator) processMetrics(envoyproxy *egv1a1.EnvoyProxy, resources *resource.Resources) (*ir.Metrics, []ir.ResolvedMetricSink, error) {
	if envoyproxy == nil ||
		envoyproxy.Spec.Telemetry == nil ||
		envoyproxy.Spec.Telemetry.Metrics == nil {
		return nil, nil, nil
	}

	var resolvedSinks []ir.ResolvedMetricSink
	seen := sets.NewString()

	for i, sink := range envoyproxy.Spec.Telemetry.Metrics.Sinks {
		if sink.OpenTelemetry == nil {
			continue
		}

		destName := fmt.Sprintf("metrics_otel_%d", i)
		settingName := irDestinationSettingName(destName, -1)
		ds, _, err := t.processBackendRefs(settingName, sink.OpenTelemetry.BackendCluster, envoyproxy.Namespace, resources, envoyproxy)
		if err != nil {
			return nil, nil, err
		}
		// TODO: update when OTLP/HTTP is completely supported (logs, traces, metrics)
		for _, d := range ds {
			d.Protocol = ir.GRPC
		}

		authority := getAuthorityFromDestination(ds)

		// Fallback to deprecated host/port
		if len(ds) == 0 && sink.OpenTelemetry.Host != nil {
			ds = destinationSettingFromHostAndPort(settingName, *sink.OpenTelemetry.Host, uint32(sink.OpenTelemetry.Port))
			authority = *sink.OpenTelemetry.Host
		}

		if len(ds) > 0 && len(ds[0].Endpoints) > 0 {
			// Skip duplicate sinks (same address:port)
			ep := ds[0].Endpoints[0]
			addr := net.JoinHostPort(ep.Host, strconv.Itoa(int(ep.Port)))
			if seen.Has(addr) {
				continue
			}
			seen.Insert(addr)

			resolvedSinks = append(resolvedSinks, ir.ResolvedMetricSink{
				Destination: ir.RouteDestination{
					Name:     destName,
					Settings: ds,
					Metadata: buildResourceMetadata(envoyproxy, nil),
				},
				Authority:                authority,
				Headers:                  sink.OpenTelemetry.Headers,
				ResourceAttributes:       sink.OpenTelemetry.ResourceAttributes,
				ReportCountersAsDeltas:   ptr.Deref(sink.OpenTelemetry.ReportCountersAsDeltas, false),
				ReportHistogramsAsDeltas: ptr.Deref(sink.OpenTelemetry.ReportHistogramsAsDeltas, false),
			})
		}
	}

	return &ir.Metrics{
		EnableVirtualHostStats:          ptr.Deref(envoyproxy.Spec.Telemetry.Metrics.EnableVirtualHostStats, false),
		EnablePerEndpointStats:          ptr.Deref(envoyproxy.Spec.Telemetry.Metrics.EnablePerEndpointStats, false),
		EnableRequestResponseSizesStats: ptr.Deref(envoyproxy.Spec.Telemetry.Metrics.EnableRequestResponseSizesStats, false),
	}, resolvedSinks, nil
}

func (t *Translator) processBackendRefs(name string, backendCluster egv1a1.BackendCluster, namespace string,
	resources *resource.Resources, envoyProxy *egv1a1.EnvoyProxy,
) ([]*ir.DestinationSetting, *ir.TrafficFeatures, error) {
	traffic, err := translateTrafficFeatures(backendCluster.BackendSettings)
	if err != nil {
		return nil, nil, err
	}
	result := make([]*ir.DestinationSetting, 0, len(backendCluster.BackendRefs))
	for i := range backendCluster.BackendRefs {
		ref := &backendCluster.BackendRefs[i]
		ns := NamespaceDerefOr(ref.Namespace, namespace)
		kind := KindDerefOr(ref.Kind, resource.KindService)
		switch kind {
		case resource.KindService:
			if err := t.validateBackendRefService(ref.BackendObjectReference, ns, corev1.ProtocolTCP); err != nil {
				return nil, nil, err
			}
			ds, err := t.processServiceDestinationSetting(name, ref.BackendObjectReference, ns, ir.TCP, envoyProxy, nil)
			if err != nil {
				return nil, nil, err
			}
			result = append(result, ds)
		case resource.KindBackend:
			if err := t.validateBackendRefBackend(ref.BackendObjectReference, resources, ns, true); err != nil {
				return nil, nil, err
			}
			ds := t.processBackendDestinationSetting(name, ref.BackendObjectReference, ns, ir.TCP)
			// Dynamic resolver destinations are not supported for none-route destinations
			if ds.IsDynamicResolver {
				return nil, nil, errors.New("dynamic resolver destinations are not supported")
			}
			// Apply TLS config for backend (telemetry) clusters
			backend := t.GetBackend(ns, string(ref.Name))
			if backend.Spec.TLS != nil {
				tlsConfig, err := t.processServerValidationTLSSettings(backend)
				if err != nil {
					return nil, nil, err
				}
				ds.TLS = tlsConfig
				// Infer SNI from FQDN for telemetry backends (no Host header available)
				if ds.TLS.SNI == nil && len(backend.Spec.Endpoints) == 1 && backend.Spec.Endpoints[0].FQDN != nil {
					ds.TLS.SNI = &backend.Spec.Endpoints[0].FQDN.Hostname
				}
			}
			result = append(result, ds)
		default:
			return nil, nil, fmt.Errorf("unsupported kind for backendRefs: %s", kind)
		}
	}
	if len(result) == 0 {
		return nil, traffic, nil
	}
	return result, traffic, nil
}

func destinationSettingFromHostAndPort(name, host string, port uint32) []*ir.DestinationSetting {
	// check if host is an IP address or a hostname
	addressType := ir.FQDN
	if net.ParseIP(host) != nil {
		addressType = ir.IP
	}

	return []*ir.DestinationSetting{
		{
			Name:        name,
			Weight:      ptr.To[uint32](1),
			Protocol:    ir.GRPC,
			AddressType: ptr.To(addressType),
			Endpoints:   []*ir.DestinationEndpoint{ir.NewDestEndpoint(nil, host, port, false, nil)},
		},
	}
}

var celEnv, _ = cel.NewEnv()

func validCELExpression(expr string) bool {
	_, issue := celEnv.Parse(expr)
	return issue.Err() == nil
}

// servicePortToContainerPort translates a service port into an ephemeral
// container port.
func (t *Translator) servicePortToContainerPort(servicePort int32, envoyProxy *egv1a1.EnvoyProxy) int32 {
	// When running on the local host using the Host infrastructure provider, disable translating the
	// gateway listener port into a non-privileged port and reuse the specified value.
	if t.RunningOnHost {
		return servicePort
	}

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
