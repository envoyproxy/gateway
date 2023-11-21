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
func patchProxyProtocolFilter(xdsListener *listenerv3.Listener, irListener *ir.HTTPListener) {
	// Return early if unset
	if xdsListener == nil || irListener == nil || !irListener.EnableProxyProtocol {
		return
	}

	// Return early if filter already exists.
	for _, filter := range xdsListener.ListenerFilters {
		if filter.Name == wellknown.ProxyProtocol {
			return
		}
	}

	proxyProtocolFilter := buildProxyProtocolFilter()

	if proxyProtocolFilter != nil {
		xdsListener.ListenerFilters = append(xdsListener.ListenerFilters, proxyProtocolFilter)
	}
}

// buildProxypProtocolFilter returns a Proxy Protocol listener filter from the provided IR listener.
func buildProxyProtocolFilter() *listenerv3.ListenerFilter {
	pp := &proxyprotocolv3.ProxyProtocol{}

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
