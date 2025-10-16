// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/stretchr/testify/assert"
)

type ListenerFilterSortTestCase struct {
	name    string
	filters []*listenerv3.ListenerFilter
	want    []*listenerv3.ListenerFilter
}

func Test_sortListenerFilters(t *testing.T) {
	tests := []*ListenerFilterSortTestCase{
		{
			name: "ensure proxy protocol before tls inspector",
			filters: []*listenerv3.ListenerFilter{
				listenerFilterForTest(wellknown.TLSInspector),
				listenerFilterForTest(wellknown.ProxyProtocol),
			},
			want: []*listenerv3.ListenerFilter{
				listenerFilterForTest(wellknown.ProxyProtocol),
				listenerFilterForTest(wellknown.TLSInspector),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sortListenerFilters(tt.filters)
			assert.Equalf(t, tt.want, result, "sortListenerFilters(%v)", tt.filters)
		})
	}
}

func listenerFilterForTest(name string) *listenerv3.ListenerFilter {
	return &listenerv3.ListenerFilter{
		Name: string(name),
	}
}
