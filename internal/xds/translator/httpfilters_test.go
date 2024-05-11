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
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/api/v1alpha1"
)

func Test_sortHTTPFilters(t *testing.T) {
	tests := []struct {
		name        string
		filters     []*hcmv3.HttpFilter
		filterOrder []v1alpha1.FilterPosition
		want        []*hcmv3.HttpFilter
	}{
		{
			name: "sort filters",
			filters: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Router),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
			},
			want: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(wellknown.Router),
			},
		},
		{
			name: "custom filter order-singleton filter",
			filters: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Router),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
			},
			filterOrder: []v1alpha1.FilterPosition{
				{
					Name:  v1alpha1.EnvoyFilterFault,
					After: ptr.To(v1alpha1.EnvoyFilterCORS),
				},
				{
					Name:   v1alpha1.EnvoyFilterRateLimit,
					Before: ptr.To(v1alpha1.EnvoyFilterJWTAuthn),
				},
			},
			want: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wellknown.Router),
			},
		},
		{
			name: "custom filter order-singleton-before-multipleton",
			filters: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Router),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
			},
			filterOrder: []v1alpha1.FilterPosition{
				{
					Name:   v1alpha1.EnvoyFilterRateLimit,
					Before: ptr.To(v1alpha1.EnvoyFilterWasm),
				},
			},
			want: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wellknown.Router),
			},
		},
		{
			name: "custom filter order-singleton-after-multipleton",
			filters: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Router),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
			},
			filterOrder: []v1alpha1.FilterPosition{
				{
					Name:  v1alpha1.EnvoyFilterJWTAuthn,
					After: ptr.To(v1alpha1.EnvoyFilterWasm),
				},
			},
			want: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(wellknown.Router),
			},
		},
		{
			name: "custom filter order-multipleton-before-singleton",
			filters: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Router),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
			},
			filterOrder: []v1alpha1.FilterPosition{
				{
					Name:   v1alpha1.EnvoyFilterWasm,
					Before: ptr.To(v1alpha1.EnvoyFilterJWTAuthn),
				},
			},
			want: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(wellknown.Router),
			},
		},
		{
			name: "custom filter order-multipleton-after-singleton",
			filters: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Router),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
			},
			filterOrder: []v1alpha1.FilterPosition{
				{
					Name:  v1alpha1.EnvoyFilterWasm,
					After: ptr.To(v1alpha1.EnvoyFilterRateLimit),
				},
			},
			want: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(wellknown.Router),
			},
		},
		{
			name: "custom filter order-multipleton-before-multipleton",
			filters: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Router),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
			},
			filterOrder: []v1alpha1.FilterPosition{
				{
					Name:   v1alpha1.EnvoyFilterWasm,
					Before: ptr.To(v1alpha1.EnvoyFilterExtProc),
				},
			},
			want: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(wellknown.Router),
			},
		},
		{
			name: "custom filter order-multipleton-after-multipleton",
			filters: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Router),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
			},
			filterOrder: []v1alpha1.FilterPosition{
				{
					Name:  v1alpha1.EnvoyFilterExtProc,
					After: ptr.To(v1alpha1.EnvoyFilterWasm),
				},
			},
			want: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(wellknown.Router),
			},
		},
		{
			name: "custom filter order-complex-ordering",
			filters: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Router),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
			},
			filterOrder: []v1alpha1.FilterPosition{
				{
					Name:   v1alpha1.EnvoyFilterLocalRateLimit,
					Before: ptr.To(v1alpha1.EnvoyFilterJWTAuthn),
				},
				{
					Name:  v1alpha1.EnvoyFilterLocalRateLimit,
					After: ptr.To(v1alpha1.EnvoyFilterCORS),
				},
				{
					Name:   v1alpha1.EnvoyFilterWasm,
					Before: ptr.To(v1alpha1.EnvoyFilterOAuth2),
				},
				{
					Name:   v1alpha1.EnvoyFilterExtProc,
					Before: ptr.To(v1alpha1.EnvoyFilterWasm),
				},
			},
			want: []*hcmv3.HttpFilter{
				httpFilterForTest(wellknown.Fault),
				httpFilterForTest(wellknown.CORS),
				httpFilterForTest(localRateLimitFilter),
				httpFilterForTest(extAuthFilter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(basicAuthFilter),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(extProcFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/0"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/1"),
				httpFilterForTest(wasmFilter + "/envoyextensionpolicy/default/policy-for-http-route-1/2"),
				httpFilterForTest(oauth2Filter + "/securitypolicy/default/policy-for-http-route-1"),
				httpFilterForTest(jwtAuthn),
				httpFilterForTest(wellknown.HTTPRateLimit),
				httpFilterForTest(wellknown.Router),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sortHTTPFilters(tt.filters, tt.filterOrder)
			assert.Equalf(t, tt.want, result, "sortHTTPFilters(%v)", tt.filters)
		})
	}
}

func httpFilterForTest(name string) *hcmv3.HttpFilter {
	return &hcmv3.HttpFilter{
		Name: name,
	}
}
