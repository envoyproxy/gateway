// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"strconv"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	setfilterstatev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/set_filter_state/v3"
	originaldstv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/original_dst/v3"
	listenersetfilterstatev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/set_filter_state/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
)

const (
	listenerSetFilterState = "envoy.filters.listener.set_filter_state"
	// originalDstLocalPortKey is the filter state object the original_dst listener filter reads
	// to replace the port of the connection's local address.
	originalDstLocalPortKey = "envoy.filters.listener.original_dst.local_port"
)

// buildExternalPortListenerFilters returns the listener filters that make Envoy treat connections
// as received on the listener's external port, when it binds a shifted container port instead,
// such as 10443 for 443. Otherwise the container port leaks to backends, for example as the
// destination port of upstream PROXY protocol headers.
func buildExternalPortListenerFilters(listenerDetails *ir.CoreListenerDetails) ([]*listenerv3.ListenerFilter, error) {
	if listenerDetails.ExternalPort == 0 || listenerDetails.ExternalPort == listenerDetails.Port {
		return nil, nil
	}

	setFilterState := &listenersetfilterstatev3.Config{
		OnAccept: []*setfilterstatev3.FilterStateValue{{
			Key: &setfilterstatev3.FilterStateValue_ObjectKey{ObjectKey: originalDstLocalPortKey},
			Value: &setfilterstatev3.FilterStateValue_FormatString{
				FormatString: &corev3.SubstitutionFormatString{
					Format: &corev3.SubstitutionFormatString_TextFormatSource{
						TextFormatSource: &corev3.DataSource{
							Specifier: &corev3.DataSource_InlineString{
								InlineString: strconv.FormatUint(uint64(listenerDetails.ExternalPort), 10),
							},
						},
					},
				},
			},
		}},
	}
	setFilterStateAny, err := proto.ToAnyWithValidation(setFilterState)
	if err != nil {
		return nil, err
	}
	originalDstAny, err := proto.ToAnyWithValidation(&originaldstv3.OriginalDst{})
	if err != nil {
		return nil, err
	}

	return []*listenerv3.ListenerFilter{
		{
			Name:       listenerSetFilterState,
			ConfigType: &listenerv3.ListenerFilter_TypedConfig{TypedConfig: setFilterStateAny},
		},
		{
			Name:       wellknown.OriginalDestination,
			ConfigType: &listenerv3.ListenerFilter_TypedConfig{TypedConfig: originalDstAny},
		},
	}, nil
}

// isExternalPortListenerFilter reports whether filter is one of the filters added by
// buildExternalPortListenerFilters.
func isExternalPortListenerFilter(filter *listenerv3.ListenerFilter) bool {
	return filter.Name == listenerSetFilterState || filter.Name == wellknown.OriginalDestination
}
