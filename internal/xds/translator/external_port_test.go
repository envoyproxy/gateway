// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	listenersetfilterstatev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/set_filter_state/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestBuildExternalPortListenerFilters(t *testing.T) {
	testCases := []struct {
		name         string
		port         uint32
		externalPort uint32
		wantPort     string
	}{
		{name: "no external port", port: 10080},
		{name: "external port is the listener port", port: 8080, externalPort: 8080},
		{name: "shifted port", port: 10443, externalPort: 443, wantPort: "443"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters, err := buildExternalPortListenerFilters(&ir.CoreListenerDetails{
				Port:         tc.port,
				ExternalPort: tc.externalPort,
			})
			require.NoError(t, err)

			if tc.wantPort == "" {
				require.Empty(t, filters)
				return
			}
			require.Len(t, filters, 2)
			require.Equal(t, listenerSetFilterState, filters[0].Name)
			require.Equal(t, wellknown.OriginalDestination, filters[1].Name)
			for _, filter := range filters {
				require.True(t, isExternalPortListenerFilter(filter))
			}

			config := &listenersetfilterstatev3.Config{}
			require.NoError(t, filters[0].GetTypedConfig().UnmarshalTo(config))
			require.Len(t, config.OnAccept, 1)
			require.Equal(t, originalDstLocalPortKey, config.OnAccept[0].GetObjectKey())
			require.Equal(t, tc.wantPort, config.OnAccept[0].GetFormatString().GetTextFormatSource().GetInlineString())
		})
	}
}

func TestPatchProxyProtocolFilterAfterExternalPortFilters(t *testing.T) {
	filters, err := buildExternalPortListenerFilters(&ir.CoreListenerDetails{Port: 10443, ExternalPort: 443})
	require.NoError(t, err)
	xdsListener := &listenerv3.Listener{ListenerFilters: filters}
	require.NoError(t, addXdsTLSInspectorFilter(xdsListener, nil))

	patchProxyProtocolFilter(xdsListener, &ir.ProxyProtocolSettings{})

	names := make([]string, 0, len(xdsListener.ListenerFilters))
	for _, filter := range xdsListener.ListenerFilters {
		names = append(names, filter.Name)
	}
	require.Equal(t, []string{
		listenerSetFilterState,
		wellknown.OriginalDestination,
		wellknown.ProxyProtocol,
		wellknown.TlsInspector,
	}, names)
}
