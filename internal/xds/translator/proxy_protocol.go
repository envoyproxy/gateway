// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	proxyprotocolv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/proxy_protocol/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/internal/ir"
)

// patchProxyProtocolFilter builds and appends the Proxy Protocol Filter to the
// HTTP Listener's Listener Filters if applicable.
func patchProxyProtocolFilter(xdsListener *listenerv3.Listener, enableProxyProtocol bool, proxyProtocolSettings *ir.ProxyProtocolSettings) {
	// Early return if listener is nil
	if xdsListener == nil {
		return
	}

	// Determine if proxy protocol is enabled
	isEnabled := enableProxyProtocol
	if proxyProtocolSettings != nil {
		// ProxyProtocolSettings takes precedence when provided
		isEnabled = proxyProtocolSettings.Enabled
	}

	// Early return if proxy protocol is not enabled
	if !isEnabled {
		return
	}

	// Build and patch the Proxy Protocol Filter.
	filter := buildProxyProtocolFilter(proxyProtocolSettings)
	if filter != nil {
		xdsListener.ListenerFilters = append(xdsListener.ListenerFilters, filter)
	}
}

// buildProxyProtocolFilter returns a Proxy Protocol listener filter from the provided IR listener.
func buildProxyProtocolFilter(proxyProtocolSettings *ir.ProxyProtocolSettings) *listenerv3.ListenerFilter {
	pp := &proxyprotocolv3.ProxyProtocol{}

	// Configure allow_requests_without_proxy_protocol if ProxyProtocolSettings are provided
	if proxyProtocolSettings != nil {
		pp.AllowRequestsWithoutProxyProtocol = proxyProtocolSettings.AllowRequestsWithoutProxyProtocol
	}

	ppAny, err := anypb.New(pp)
	if err != nil {
		return nil
	}

	return &listenerv3.ListenerFilter{
		Name: wellknown.ProxyProtocol,
		ConfigType: &listenerv3.ListenerFilter_TypedConfig{
			TypedConfig: ppAny,
		},
	}
}
