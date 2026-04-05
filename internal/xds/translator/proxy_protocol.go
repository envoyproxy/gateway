// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"slices"

	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	proxyprotocolv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/proxy_protocol/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/internal/ir"
)

// patchProxyProtocolFilter builds and appends the Proxy Protocol Filter to the
// HTTP Listener's Listener Filters if applicable.
func patchProxyProtocolFilter(xdsListener *listenerv3.Listener, proxyProtocolSettings *ir.ProxyProtocolSettings) {
	// Early return if listener is nil
	if xdsListener == nil {
		return
	}

	// Return early if filter already exists.
	for _, filter := range xdsListener.ListenerFilters {
		if filter.Name == wellknown.ProxyProtocol {
			return
		}
	}

	// Early return if proxy protocol is not enabled
	if proxyProtocolSettings == nil {
		return
	}

	// Build and patch the Proxy Protocol Filter.
	filter := buildProxyProtocolFilter(proxyProtocolSettings)
	if filter != nil {
		xdsListener.ListenerFilters = slices.Insert(xdsListener.ListenerFilters, 0, filter)
	}
}

// buildProxyProtocolFilter returns a Proxy Protocol listener filter from the provided IR listener.
func buildProxyProtocolFilter(proxyProtocolSettings *ir.ProxyProtocolSettings) *listenerv3.ListenerFilter {
	pp := &proxyprotocolv3.ProxyProtocol{}

	// Configure allow_requests_without_proxy_protocol if ProxyProtocolSettings are provided
	if proxyProtocolSettings.Optional {
		pp.AllowRequestsWithoutProxyProtocol = proxyProtocolSettings.Optional
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
