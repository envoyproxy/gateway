// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package remote

import (
	"encoding/json"
	"fmt"

	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/proto/remoteinfra"
)

// infraToProto converts the internal ir.Infra into its wire representation.
//
// Most fields are mapped structurally. ProxyInfra.Config (an EnvoyProxy CRD) is
// the only field carried as JSON bytes, because the EnvoyProxy schema is large
// and evolves independently of this contract. ResolvedMetricSink.Destination is
// translated field-by-field into the subset of ir.RouteDestination that remote
// providers need to enable Metric Sinks in Envoy.
func infraToProto(infra *ir.Infra) (*remoteinfra.Infra, error) {
	if infra == nil {
		return nil, nil
	}

	out := &remoteinfra.Infra{}

	if infra.Proxy != nil {
		proxy, err := proxyInfraToProto(infra.Proxy)
		if err != nil {
			return nil, err
		}
		out.Proxy = proxy
	}

	return out, nil
}

func proxyInfraToProto(p *ir.ProxyInfra) (*remoteinfra.ProxyInfra, error) {
	out := &remoteinfra.ProxyInfra{
		Name:      p.Name,
		Namespace: p.Namespace,
		Addresses: p.Addresses,
	}

	if p.Metadata != nil {
		out.Metadata = infraMetadataToProto(p.Metadata)
	}

	if p.Config != nil {
		bs, err := json.Marshal(p.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal proxy config: %w", err)
		}
		out.Config = bs
	}

	for _, l := range p.Listeners {
		if l == nil {
			continue
		}
		out.Listeners = append(out.Listeners, proxyListenerToProto(l))
	}

	for i := range p.ResolvedMetricSinks {
		out.ResolvedMetricSinks = append(out.ResolvedMetricSinks, resolvedMetricSinkToProto(&p.ResolvedMetricSinks[i]))
	}

	return out, nil
}

func infraMetadataToProto(m *ir.InfraMetadata) *remoteinfra.InfraMetadata {
	out := &remoteinfra.InfraMetadata{
		Annotations: m.Annotations,
		Labels:      m.Labels,
	}

	if m.OwnerReference != nil {
		out.OwnerReference = resourceMetadataToProto(m.OwnerReference)
	}

	return out
}

func resourceMetadataToProto(m *ir.ResourceMetadata) *remoteinfra.ResourceMetadata {
	out := &remoteinfra.ResourceMetadata{
		Kind:        m.Kind,
		Name:        m.Name,
		Namespace:   m.Namespace,
		SectionName: m.SectionName,
	}

	for _, a := range m.Annotations {
		out.Annotations = append(out.Annotations, &remoteinfra.MapEntry{
			Key:   a.Key,
			Value: a.Value,
		})
	}

	for _, pol := range m.Policies {
		if pol == nil {
			continue
		}
		out.Policies = append(out.Policies, &remoteinfra.PolicyMetadata{
			Kind:      pol.Kind,
			Name:      pol.Name,
			Namespace: pol.Namespace,
		})
	}

	return out
}

func proxyListenerToProto(l *ir.ProxyListener) *remoteinfra.ProxyListener {
	out := &remoteinfra.ProxyListener{
		Name: l.Name,
	}

	for _, port := range l.Ports {
		out.Ports = append(out.Ports, &remoteinfra.ListenerPort{
			Name:          port.Name,
			Protocol:      string(port.Protocol),
			ServicePort:   port.ServicePort,
			ContainerPort: port.ContainerPort,
		})
	}

	return out
}

// resolvedMetricSinkToProto translates every field of ir.ResolvedMetricSink
// (Authority, Headers, ResourceAttributes, ReportCountersAsDeltas,
// ReportHistogramsAsDeltas). Only its Destination is narrowed to a subset; see
// routeDestinationToProto.
func resolvedMetricSinkToProto(s *ir.ResolvedMetricSink) *remoteinfra.ResolvedMetricSink {
	out := &remoteinfra.ResolvedMetricSink{
		Destination:              routeDestinationToProto(s.Destination),
		Authority:                s.Authority,
		ResourceAttributes:       s.ResourceAttributes,
		ReportCountersAsDeltas:   s.ReportCountersAsDeltas,
		ReportHistogramsAsDeltas: s.ReportHistogramsAsDeltas,
	}

	for _, h := range s.Headers {
		out.Headers = append(out.Headers, &remoteinfra.HTTPHeader{
			Name:  string(h.Name),
			Value: h.Value,
		})
	}

	return out
}

// routeDestinationToProto translates the subset of ir.RouteDestination that
// remote providers need.
//
// Taken:
//   - Settings
//
// Intentionally dropped (not needed by remote providers):
//   - Name, StatName: xDS cluster naming/stat details internal to Envoy Gateway.
//   - Metadata: provider/user resource metadata used only for Envoy route
//     metadata enrichment.
func routeDestinationToProto(d ir.RouteDestination) *remoteinfra.RouteDestination {
	out := &remoteinfra.RouteDestination{}

	for _, s := range d.Settings {
		if s == nil {
			continue
		}
		out.Settings = append(out.Settings, destinationSettingToProto(s))
	}

	return out
}

// destinationSettingToProto translates the subset of ir.DestinationSetting that
// remote providers need.
//
// Taken:
//   - Endpoints (see the DestinationEndpoint field notes below)
//   - TLS
//
// Intentionally dropped (routing/load-balancing details Envoy Gateway resolves
// itself and that a remote infra provider does not act on):
//   - Name, Weight, Priority: cluster/endpoint weighting and naming.
//   - Protocol, ForceHTTP1Upstream: upstream protocol selection.
//   - IsDynamicResolver, IsCustomBackend, Invalid: destination classification.
//   - AddressType, IPFamily: endpoint address-family resolution.
//   - Filters: per-destination request/response filters.
//   - PreferLocal: zone-aware routing.
//   - Metadata: provider/user resource metadata.
func destinationSettingToProto(s *ir.DestinationSetting) *remoteinfra.DestinationSetting {
	out := &remoteinfra.DestinationSetting{}

	// DestinationEndpoint taken: Host, Port.
	// Intentionally dropped: Hostname (SNI/health-check hostname), Path (UDS
	// path), Draining (endpoint drain state), Zone (topology zone) — all used
	// for xDS endpoint programming, not by remote infra providers.
	for _, e := range s.Endpoints {
		if e == nil {
			continue
		}
		out.Endpoints = append(out.Endpoints, &remoteinfra.DestinationEndpoint{
			Host: e.Host,
			Port: e.Port,
		})
	}

	if s.TLS != nil {
		out.Tls = tlsUpstreamConfigToProto(s.TLS)
	}

	return out
}

// tlsUpstreamConfigToProto translates the subset of ir.TLSUpstreamConfig that
// remote providers need.
//
// Taken:
//   - SNI
//   - UseSystemTrustStore
//   - CACertificate.Certificate (via the TLSConfig message; see note below)
//
// Intentionally dropped:
//   - AutoSNIFromEndpointHostname, InsecureSkipVerify: connection-time TLS
//     behavior applied by Envoy, not the provider.
//   - SubjectAltNames: SAN verification matchers.
//   - The embedded ir.TLSConfig fields other than the CA certificate
//     (client/server certificates, CRL, cipher suites, TLS versions, session
//     resumption, etc.): full downstream/upstream TLS material Envoy Gateway
//     programs into xDS directly.
//   - CACertificate.Name and CACertificate.SDS: the Secret name and SDS server
//     config; only the resolved certificate bytes are forwarded.
func tlsUpstreamConfigToProto(t *ir.TLSUpstreamConfig) *remoteinfra.TLSUpstreamConfig {
	out := &remoteinfra.TLSUpstreamConfig{
		Sni:                 t.SNI,
		UseSystemTrustStore: t.UseSystemTrustStore,
	}

	if t.CACertificate != nil {
		out.TlsConfig = &remoteinfra.TLSConfig{
			CaCertificate: &remoteinfra.TLSCACertificate{
				Certificate: t.CACertificate.Certificate,
			},
		}
	}

	return out
}

// protoToInfra converts the wire representation back into an ir.Infra. It is
// the inverse of infraToProto and is primarily used to validate the initial conversion.
func protoToInfra(in *remoteinfra.Infra) (*ir.Infra, error) {
	if in == nil {
		return nil, nil
	}

	out := &ir.Infra{}

	if in.GetProxy() != nil {
		proxy, err := protoToProxyInfra(in.GetProxy())
		if err != nil {
			return nil, err
		}
		out.Proxy = proxy
	}

	return out, nil
}

func protoToProxyInfra(p *remoteinfra.ProxyInfra) (*ir.ProxyInfra, error) {
	out := &ir.ProxyInfra{
		Name:      p.GetName(),
		Namespace: p.GetNamespace(),
		Addresses: p.GetAddresses(),
	}

	if p.GetMetadata() != nil {
		out.Metadata = protoToInfraMetadata(p.GetMetadata())
	}

	if len(p.GetConfig()) > 0 {
		cfg := &egv1a1.EnvoyProxy{}
		if err := json.Unmarshal(p.GetConfig(), cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal proxy config: %w", err)
		}
		out.Config = cfg
	}

	for _, l := range p.GetListeners() {
		out.Listeners = append(out.Listeners, protoToProxyListener(l))
	}

	for _, s := range p.GetResolvedMetricSinks() {
		out.ResolvedMetricSinks = append(out.ResolvedMetricSinks, protoToResolvedMetricSink(s))
	}

	return out, nil
}

func protoToInfraMetadata(m *remoteinfra.InfraMetadata) *ir.InfraMetadata {
	out := &ir.InfraMetadata{
		Annotations: m.GetAnnotations(),
		Labels:      m.GetLabels(),
	}

	if m.GetOwnerReference() != nil {
		out.OwnerReference = protoToResourceMetadata(m.GetOwnerReference())
	}

	return out
}

func protoToResourceMetadata(m *remoteinfra.ResourceMetadata) *ir.ResourceMetadata {
	out := &ir.ResourceMetadata{
		Kind:        m.GetKind(),
		Name:        m.GetName(),
		Namespace:   m.GetNamespace(),
		SectionName: m.GetSectionName(),
	}

	for _, a := range m.GetAnnotations() {
		out.Annotations = append(out.Annotations, ir.MapEntry{
			Key:   a.GetKey(),
			Value: a.GetValue(),
		})
	}

	for _, pol := range m.GetPolicies() {
		out.Policies = append(out.Policies, &ir.PolicyMetadata{
			Kind:      pol.GetKind(),
			Name:      pol.GetName(),
			Namespace: pol.GetNamespace(),
		})
	}

	return out
}

func protoToProxyListener(l *remoteinfra.ProxyListener) *ir.ProxyListener {
	out := &ir.ProxyListener{
		Name: l.GetName(),
	}

	for _, port := range l.GetPorts() {
		out.Ports = append(out.Ports, ir.ListenerPort{
			Name:          port.GetName(),
			Protocol:      ir.ProtocolType(port.GetProtocol()),
			ServicePort:   port.GetServicePort(),
			ContainerPort: port.GetContainerPort(),
		})
	}

	return out
}

func protoToResolvedMetricSink(s *remoteinfra.ResolvedMetricSink) ir.ResolvedMetricSink {
	out := ir.ResolvedMetricSink{
		Authority:                s.GetAuthority(),
		ResourceAttributes:       s.GetResourceAttributes(),
		ReportCountersAsDeltas:   s.GetReportCountersAsDeltas(),
		ReportHistogramsAsDeltas: s.GetReportHistogramsAsDeltas(),
	}

	if s.GetDestination() != nil {
		out.Destination = protoToRouteDestination(s.GetDestination())
	}

	for _, h := range s.GetHeaders() {
		out.Headers = append(out.Headers, gwapiv1.HTTPHeader{
			Name:  gwapiv1.HTTPHeaderName(h.GetName()),
			Value: h.GetValue(),
		})
	}

	return out
}

func protoToRouteDestination(d *remoteinfra.RouteDestination) ir.RouteDestination {
	out := ir.RouteDestination{}

	for _, s := range d.GetSettings() {
		out.Settings = append(out.Settings, protoToDestinationSetting(s))
	}

	return out
}

func protoToDestinationSetting(s *remoteinfra.DestinationSetting) *ir.DestinationSetting {
	out := &ir.DestinationSetting{}

	for _, e := range s.GetEndpoints() {
		out.Endpoints = append(out.Endpoints, &ir.DestinationEndpoint{
			Host: e.GetHost(),
			Port: e.GetPort(),
		})
	}

	if s.GetTls() != nil {
		out.TLS = protoToTLSUpstreamConfig(s.GetTls())
	}

	return out
}

func protoToTLSUpstreamConfig(t *remoteinfra.TLSUpstreamConfig) *ir.TLSUpstreamConfig {
	out := &ir.TLSUpstreamConfig{
		SNI:                 t.Sni,
		UseSystemTrustStore: t.GetUseSystemTrustStore(),
	}

	if tc := t.GetTlsConfig(); tc != nil {
		if ca := tc.GetCaCertificate(); ca != nil {
			out.CACertificate = &ir.TLSCACertificate{
				Certificate: ca.GetCertificate(),
			}
		}
	}

	return out
}
