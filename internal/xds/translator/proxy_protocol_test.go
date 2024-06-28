// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/stretchr/testify/require"
)

func TestPatchProxyProtocolFilter(t *testing.T) {
	type testCase struct {
		name     string
		listener *listenerv3.Listener
	}

	enableProxyProtocol := true

	testCases := []testCase{
		{
			name: "listener with proxy proto available already",
			listener: &listenerv3.Listener{
				ListenerFilters: []*listenerv3.ListenerFilter{
					{
						Name: wellknown.ProxyProtocol,
					},
				},
			},
		},
		{
			name: "listener with tls, append proxy proto",
			listener: &listenerv3.Listener{
				ListenerFilters: []*listenerv3.ListenerFilter{
					{
						Name: wellknown.TLSInspector,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patchProxyProtocolFilter(tc.listener, enableProxyProtocol)
			// proxy proto filter should be added always as first
			require.Equal(t, wellknown.ProxyProtocol, tc.listener.ListenerFilters[0].Name)
		})
	}
}
