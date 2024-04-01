// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/stretchr/testify/assert"
)

func Test_sortHTTPFilters(t *testing.T) {
	tests := []struct {
		name    string
		filters []*hcmv3.HttpFilter
		want    []*hcmv3.HttpFilter
	}{
		{
			name: "sort filters",
			filters: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Router),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(oauth2Filter + "-route1"),
				httpFilterForTest(basicAuthFilter + "-route1"),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(extAuthFilter + "-route1"),
			},
			want: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(extAuthFilter + "-route1"),
				httpFilterForTest(basicAuthFilter + "-route1"),
				httpFilterForTest(oauth2Filter + "-route1"),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(wellknown.Router),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, sortHTTPFilters(tt.filters), "sortHTTPFilters(%v)", tt.filters)
		})
	}
}

func httpFilterForTest(name string) *hcmv3.HttpFilter {
	return &hcmv3.HttpFilter{
		Name: name,
	}
}
