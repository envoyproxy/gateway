// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

var (
	// HTTPListener
	happyHTTPListener = HTTPListener{
		CoreListenerDetails: CoreListenerDetails{
			Name:    "happy",
			Address: "0.0.0.0",
			Port:    80,
		},
		Hostnames: []string{"example.com"},
		Routes:    []*HTTPRoute{&happyHTTPRoute},
	}
	happyHTTPSListener = HTTPListener{
		CoreListenerDetails: CoreListenerDetails{
			Name:    "happy",
			Address: "0.0.0.0",
			Port:    80,
		},
		Hostnames: []string{"example.com"},
		TLS: &TLSConfig{
			Certificates: []TLSCertificate{{
				Name:        "happy",
				Certificate: []byte{1, 2, 3},
				PrivateKey:  []byte{1, 2, 3},
			}},
		},
		Routes: []*HTTPRoute{&happyHTTPRoute},
	}
	redactedHappyHTTPSListener = HTTPListener{
		CoreListenerDetails: CoreListenerDetails{
			Name:    "happy",
			Address: "0.0.0.0",
			Port:    80,
		},
		Hostnames: []string{"example.com"},
		TLS: &TLSConfig{
			Certificates: []TLSCertificate{{
				Name:        "happy",
				Certificate: []byte{1, 2, 3},
				PrivateKey:  redacted,
			}},
		},
		Routes: []*HTTPRoute{&happyHTTPRoute},
	}
	invalidAddrHTTPListener = HTTPListener{
		CoreListenerDetails: CoreListenerDetails{
			Name:    "invalid-addr",
			Address: "1.0.0",
			Port:    80,
		},
		Hostnames: []string{"example.com"},
		Routes:    []*HTTPRoute{&happyHTTPRoute},
	}
	invalidBackendHTTPListener = HTTPListener{
		CoreListenerDetails: CoreListenerDetails{
			Name:    "invalid-backend-match",
			Address: "0.0.0.0",
			Port:    80,
		},
		Hostnames: []string{"example.com"},
		Routes:    []*HTTPRoute{&invalidBackendHTTPRoute},
	}
	weightedInvalidBackendsHTTPListener = HTTPListener{
		CoreListenerDetails: CoreListenerDetails{
			Name:    "weighted-invalid-backends-match",
			Address: "0.0.0.0",
			Port:    80,
		},
		Hostnames: []string{"example.com"},
		Routes:    []*HTTPRoute{&weightedInvalidBackendsHTTPRoute},
	}

	// TCPListener
	happyTCPListenerTLSPassthrough = TCPListener{
		CoreListenerDetails: CoreListenerDetails{
			Name:    "happy",
			Address: "0.0.0.0",
			Port:    80,
		},
		Routes: []*TCPRoute{&happyTCPRouteTLSPassthrough},
	}

	happyTCPListenerTLSTerminate = TCPListener{
		CoreListenerDetails: CoreListenerDetails{
			Name:    "happy",
			Address: "0.0.0.0",
			Port:    80,
		},
		Routes: []*TCPRoute{&happyTCPRouteTLSTermination},
	}

	emptySNITCPListenerTLSPassthrough = TCPListener{
		CoreListenerDetails: CoreListenerDetails{
			Name:    "empty-sni",
			Address: "0.0.0.0",
			Port:    80,
		},
		Routes: []*TCPRoute{&emptySNITCPRoute},
	}
	invalidNameTCPListenerTLSPassthrough = TCPListener{
		CoreListenerDetails: CoreListenerDetails{
			Address: "0.0.0.0",
			Port:    80,
		},
		Routes: []*TCPRoute{&happyTCPRouteTLSPassthrough},
	}
	invalidAddrTCPListenerTLSPassthrough = TCPListener{
		CoreListenerDetails: CoreListenerDetails{
			Name:    "invalid-addr",
			Address: "1.0.0",
			Port:    80,
		},
		Routes: []*TCPRoute{&happyTCPRouteTLSPassthrough},
	}
	invalidSNITCPListenerTLSPassthrough = TCPListener{
		CoreListenerDetails: CoreListenerDetails{
			Address: "0.0.0.0",
			Port:    80,
		},
		Routes: []*TCPRoute{&invalidSNITCPRoute},
	}

	// TCPRoute
	happyTCPRouteTLSPassthrough = TCPRoute{
		Name:        "happy-tls-passthrough",
		TLS:         &TLS{TLSInspectorConfig: &TLSInspectorConfig{SNIs: []string{"example.com"}}},
		Destination: &happyRouteDestination,
	}
	happyTCPRouteTLSTermination = TCPRoute{
		Name: "happy-tls-termination",
		TLS: &TLS{Terminate: &TLSConfig{Certificates: []TLSCertificate{{
			Name:        "happy",
			Certificate: []byte("server-cert"),
			PrivateKey:  []byte("priv-key"),
		}}}},
		Destination: &happyRouteDestination,
	}
	emptySNITCPRoute = TCPRoute{
		Name:        "empty-sni",
		Destination: &happyRouteDestination,
	}

	invalidSNITCPRoute = TCPRoute{
		Name:        "invalid-sni",
		TLS:         &TLS{TLSInspectorConfig: &TLSInspectorConfig{SNIs: []string{}}},
		Destination: &happyRouteDestination,
	}

	// UDPListener
	happyUDPListener = UDPListener{
		CoreListenerDetails: CoreListenerDetails{
			Name:    "happy",
			Address: "0.0.0.0",
			Port:    80,
		},
		Route: &happyUDPRoute,
	}
	invalidNameUDPListener = UDPListener{
		CoreListenerDetails: CoreListenerDetails{
			Address: "0.0.0.0",
			Port:    80,
		},
		Route: &happyUDPRoute,
	}
	invalidAddrUDPListener = UDPListener{
		CoreListenerDetails: CoreListenerDetails{
			Name:    "invalid-addr",
			Address: "1.0.0",
			Port:    80,
		},
		Route: &happyUDPRoute,
	}
	invalidPortUDPListenerT = UDPListener{
		CoreListenerDetails: CoreListenerDetails{
			Name:    "invalid-port",
			Address: "0.0.0.0",
			Port:    0,
		},
		Route: &happyUDPRoute,
	}

	// UDPRoute
	happyUDPRoute = UDPRoute{
		Name:        "happy",
		Destination: &happyRouteDestination,
	}

	// HTTPRoute
	happyHTTPRoute = HTTPRoute{
		Name:     "happy",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("example"),
		},
		Destination: &happyRouteDestination,
	}
	invalidBackendHTTPRoute = HTTPRoute{
		Name:     "invalid-backend",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("invalid-backend"),
		},
	}
	weightedInvalidBackendsHTTPRoute = HTTPRoute{
		Name:     "weighted-invalid-backends",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("invalid-backends"),
		},
		Destination: &happyRouteDestination,
	}

	redirectHTTPRoute = HTTPRoute{
		Name:     "redirect",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("redirect"),
		},
		Redirect: &Redirect{
			Scheme:   ptr.To("https"),
			Hostname: ptr.To("redirect.example.com"),
			Path: &HTTPPathModifier{
				FullReplace: ptr.To("/redirect"),
			},
			Port:       ptr.To(uint32(8443)),
			StatusCode: ptr.To[int32](301),
		},
	}
	// A direct response error is used when an invalid filter type is supplied
	invalidFilterHTTPRoute = HTTPRoute{
		Name:     "filter-error",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("filter-error"),
		},
		DirectResponse: &CustomResponse{
			StatusCode: ptr.To(uint32(500)),
		},
	}

	redirectFilterInvalidStatus = HTTPRoute{
		Name:     "redirect-bad-status-scheme-nopat",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("redirect"),
		},
		Redirect: &Redirect{
			Scheme:     ptr.To("err"),
			Hostname:   ptr.To("redirect.example.com"),
			Path:       &HTTPPathModifier{},
			Port:       ptr.To(uint32(8443)),
			StatusCode: ptr.To[int32](305),
		},
	}
	redirectFilterBadPath = HTTPRoute{
		Name:     "redirect",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("redirect"),
		},
		Redirect: &Redirect{
			Scheme:   ptr.To("https"),
			Hostname: ptr.To("redirect.example.com"),
			Path: &HTTPPathModifier{
				FullReplace:        ptr.To("/redirect"),
				PrefixMatchReplace: ptr.To("/redirect"),
			},
			Port:       ptr.To(uint32(8443)),
			StatusCode: ptr.To[int32](301),
		},
	}
	directResponseBadStatus = HTTPRoute{
		Name:     "redirect",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("redirect"),
		},
		DirectResponse: &CustomResponse{
			StatusCode: ptr.To(uint32(799)),
		},
	}

	urlRewriteHTTPRoute = HTTPRoute{
		Name:     "rewrite",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("rewrite"),
		},
		URLRewrite: &URLRewrite{
			Host: &HTTPHostModifier{
				Name: ptr.To("rewrite.example.com"),
			},
			Path: &ExtendedHTTPPathModifier{
				HTTPPathModifier: HTTPPathModifier{
					FullReplace: ptr.To("/rewrite"),
				},
			},
		},
	}

	urlRewriteFilterBadPath = HTTPRoute{
		Name:     "rewrite",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("rewrite"),
		},
		URLRewrite: &URLRewrite{
			Host: &HTTPHostModifier{
				Name: ptr.To("rewrite.example.com"),
			},
			Path: &ExtendedHTTPPathModifier{
				HTTPPathModifier: HTTPPathModifier{
					FullReplace:        ptr.To("/rewrite"),
					PrefixMatchReplace: ptr.To("/rewrite"),
				},
			},
		},
	}

	addRequestHeaderHTTPRoute = HTTPRoute{
		Name:     "addheader",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("addheader"),
		},
		AddRequestHeaders: []AddHeader{
			{
				Name:   "example-header",
				Value:  []string{"example-value"},
				Append: true,
			},
			{
				Name:   "example-header-2",
				Value:  []string{"example-value-2"},
				Append: false,
			},
			{
				Name:   "empty-header",
				Append: false,
			},
		},
	}

	removeRequestHeaderHTTPRoute = HTTPRoute{
		Name:     "remheader",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("remheader"),
		},
		RemoveRequestHeaders: []string{
			"x-request-header",
			"example-header",
			"another-header",
		},
	}

	addAndRemoveRequestHeadersDupeHTTPRoute = HTTPRoute{
		Name:     "duplicateheader",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("duplicateheader"),
		},
		AddRequestHeaders: []AddHeader{
			{
				Name:   "example-header",
				Value:  []string{"example-value"},
				Append: true,
			},
			{
				Name:   "example-header",
				Value:  []string{"example-value-2"},
				Append: false,
			},
		},
		RemoveRequestHeaders: []string{
			"x-request-header",
			"example-header",
			"example-header",
		},
	}

	addRequestHeaderEmptyHTTPRoute = HTTPRoute{
		Name:     "addemptyheader",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("addemptyheader"),
		},
		AddRequestHeaders: []AddHeader{
			{
				Name:   "",
				Value:  []string{"example-value"},
				Append: true,
			},
		},
	}

	addResponseHeaderHTTPRoute = HTTPRoute{
		Name:     "addheader",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("addheader"),
		},
		AddResponseHeaders: []AddHeader{
			{
				Name:   "example-header",
				Value:  []string{"example-value"},
				Append: true,
			},
			{
				Name:   "example-header-2",
				Value:  []string{"example-value-2"},
				Append: false,
			},
			{
				Name:   "empty-header",
				Append: false,
			},
		},
	}

	removeResponseHeaderHTTPRoute = HTTPRoute{
		Name:     "remheader",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("remheader"),
		},
		RemoveResponseHeaders: []string{
			"x-request-header",
			"example-header",
			"another-header",
		},
	}

	addAndRemoveResponseHeadersDupeHTTPRoute = HTTPRoute{
		Name:     "duplicateheader",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("duplicateheader"),
		},
		AddResponseHeaders: []AddHeader{
			{
				Name:   "example-header",
				Value:  []string{"example-value"},
				Append: true,
			},
			{
				Name:   "example-header",
				Value:  []string{"example-value-2"},
				Append: false,
			},
		},
		RemoveResponseHeaders: []string{
			"x-request-header",
			"example-header",
			"example-header",
		},
	}

	addResponseHeaderEmptyHTTPRoute = HTTPRoute{
		Name:     "addemptyheader",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("addemptyheader"),
		},
		AddResponseHeaders: []AddHeader{
			{
				Name:   "",
				Value:  []string{"example-value"},
				Append: true,
			},
		},
	}

	jwtAuthenHTTPRoute = HTTPRoute{
		Name:     "jwtauthen",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("jwtauthen"),
		},
		Security: &SecurityFeatures{
			JWT: &JWT{
				Providers: []JWTProvider{
					{
						Name: "test1",
						RemoteJWKS: RemoteJWKS{
							URI: "https://test1.local",
						},
					},
				},
			},
		},
	}
	requestMirrorFilter = HTTPRoute{
		Name:     "mirrorfilter",
		Hostname: "*",
		PathMatch: &StringMatch{
			Exact: ptr.To("mirrorfilter"),
		},
		Mirrors: []*MirrorPolicy{
			{
				Destination: &happyRouteDestination,
			},
		},
	}

	// RouteDestination
	happyRouteDestination = RouteDestination{
		Name: "happy-dest",
		Settings: []*DestinationSetting{
			{
				Endpoints: []*DestinationEndpoint{
					{
						Host: "10.11.12.13",
						Port: 8080,
					},
				},
			},
		},
	}
)

func TestValidateXds(t *testing.T) {
	tests := []struct {
		name  string
		input Xds
		want  []error
	}{
		{
			name: "happy",
			input: Xds{
				HTTP: []*HTTPListener{&happyHTTPListener},
			},
			want: nil,
		},
		{
			name: "happy tls passthrough",
			input: Xds{
				TCP: []*TCPListener{&happyTCPListenerTLSPassthrough},
			},
			want: nil,
		},
		{
			name: "happy tls terminate",
			input: Xds{
				TCP: []*TCPListener{&happyTCPListenerTLSTerminate},
			},
			want: nil,
		},
		{
			name: "invalid listener",
			input: Xds{
				HTTP: []*HTTPListener{&happyHTTPListener, &invalidAddrHTTPListener},
			},
			want: []error{ErrListenerAddressInvalid},
		},
		{
			name: "invalid backend",
			input: Xds{
				HTTP: []*HTTPListener{&happyHTTPListener, &invalidBackendHTTPListener},
			},
			want: nil,
		},
		{
			name: "weighted invalid backend",
			input: Xds{
				HTTP: []*HTTPListener{&happyHTTPListener, &weightedInvalidBackendsHTTPListener},
			},
			want: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.want == nil {
				require.NoError(t, test.input.Validate())
			} else {
				got := test.input.Validate()
				for _, w := range test.want {
					require.ErrorContains(t, got, w.Error())
				}
			}
		})
	}
}

func TestValidateHTTPListener(t *testing.T) {
	tests := []struct {
		name  string
		input HTTPListener
		want  []error
	}{
		{
			name:  "happy",
			input: happyHTTPListener,
			want:  nil,
		},
		{
			name: "invalid name",
			input: HTTPListener{
				CoreListenerDetails: CoreListenerDetails{
					Address: "0.0.0.0",
					Port:    80,
				},
				Hostnames: []string{"example.com"},
				Routes:    []*HTTPRoute{&happyHTTPRoute},
			},
			want: []error{ErrListenerNameEmpty},
		},
		{
			name:  "invalid addr",
			input: invalidAddrHTTPListener,
			want:  []error{ErrListenerAddressInvalid},
		},
		{
			name: "invalid port and hostnames",
			input: HTTPListener{
				CoreListenerDetails: CoreListenerDetails{
					Name:    "invalid-port-and-hostnames",
					Address: "1.0.0",
				},
				Routes: []*HTTPRoute{&happyHTTPRoute},
			},
			want: []error{ErrListenerPortInvalid, ErrHTTPListenerHostnamesEmpty},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.want == nil {
				require.NoError(t, test.input.Validate())
			} else {
				got := test.input.Validate()
				for _, w := range test.want {
					require.ErrorContains(t, got, w.Error())
				}
			}
		})
	}
}

func TestValidateTCPListener(t *testing.T) {
	tests := []struct {
		name  string
		input TCPListener
		want  []error
	}{
		{
			name:  "tls passthrough happy",
			input: happyTCPListenerTLSPassthrough,
			want:  nil,
		},
		{
			name:  "tls terminate happy",
			input: happyTCPListenerTLSTerminate,
			want:  nil,
		},
		{
			name:  "tcp empty SNIs",
			input: emptySNITCPListenerTLSPassthrough,
			want:  nil,
		},
		{
			name:  "tls passthrough invalid name",
			input: invalidNameTCPListenerTLSPassthrough,
			want:  []error{ErrListenerNameEmpty},
		},
		{
			name:  "tls passthrough invalid addr",
			input: invalidAddrTCPListenerTLSPassthrough,
			want:  []error{ErrListenerAddressInvalid},
		},
		{
			name:  "tls passthrough empty SNIs",
			input: invalidSNITCPListenerTLSPassthrough,
			want:  []error{ErrTCPRouteSNIsEmpty},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.want == nil {
				require.NoError(t, test.input.Validate())
			} else {
				got := test.input.Validate()
				for _, w := range test.want {
					require.ErrorContains(t, got, w.Error())
				}
			}
		})
	}
}

func TestValidateTLSListenerConfig(t *testing.T) {
	tests := []struct {
		name  string
		input TLSConfig
		want  error
	}{
		{
			name: "happy",
			input: TLSConfig{
				Certificates: []TLSCertificate{{
					Certificate: []byte("server-cert"),
					PrivateKey:  []byte("priv-key"),
				}},
			},
			want: nil,
		},
		{
			name: "invalid server cert",
			input: TLSConfig{
				Certificates: []TLSCertificate{{
					PrivateKey: []byte("priv-key"),
				}},
			},
			want: ErrTLSServerCertEmpty,
		},
		{
			name: "invalid private key",
			input: TLSConfig{
				Certificates: []TLSCertificate{{
					Certificate: []byte("server-cert"),
				}},
			},
			want: ErrTLSPrivateKey,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.want == nil {
				require.NoError(t, test.input.Validate())
			} else {
				require.EqualError(t, test.input.Validate(), test.want.Error())
			}
		})
	}
}

func TestEqualXds(t *testing.T) {
	tests := []struct {
		desc  string
		a     *Xds
		b     *Xds
		equal bool
	}{
		{
			desc: "out of order tcp listeners are equal",
			a: &Xds{
				TCP: []*TCPListener{
					{CoreListenerDetails: CoreListenerDetails{Name: "listener-1"}},
					{CoreListenerDetails: CoreListenerDetails{Name: "listener-2"}},
				},
			},
			b: &Xds{
				TCP: []*TCPListener{
					{CoreListenerDetails: CoreListenerDetails{Name: "listener-2"}},
					{CoreListenerDetails: CoreListenerDetails{Name: "listener-1"}},
				},
			},
			equal: true,
		},
		{
			desc: "out of order http routes are equal",
			a: &Xds{
				HTTP: []*HTTPListener{
					{
						CoreListenerDetails: CoreListenerDetails{Name: "listener-1"},
						Routes: []*HTTPRoute{
							{Name: "route-1"},
							{Name: "route-2"},
						},
					},
				},
			},
			b: &Xds{
				HTTP: []*HTTPListener{
					{
						CoreListenerDetails: CoreListenerDetails{Name: "listener-1"},
						Routes: []*HTTPRoute{
							{Name: "route-2"},
							{Name: "route-1"},
						},
					},
				},
			},
			equal: true,
		},
		{
			desc: "out of order udp listeners are equal",
			a: &Xds{
				UDP: []*UDPListener{
					{CoreListenerDetails: CoreListenerDetails{Name: "listener-1"}},
					{CoreListenerDetails: CoreListenerDetails{Name: "listener-2"}},
				},
			},
			b: &Xds{
				UDP: []*UDPListener{
					{CoreListenerDetails: CoreListenerDetails{Name: "listener-2"}},
					{CoreListenerDetails: CoreListenerDetails{Name: "listener-1"}},
				},
			},
			equal: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, tc.equal, cmp.Equal(tc.a, tc.b))
		})
	}
}

func TestValidateUDPListener(t *testing.T) {
	tests := []struct {
		name  string
		input UDPListener
		want  []error
	}{
		{
			name:  "udp happy",
			input: happyUDPListener,
			want:  nil,
		},
		{
			name:  "udp invalid name",
			input: invalidNameUDPListener,
			want:  []error{ErrListenerNameEmpty},
		},
		{
			name:  "udp invalid addr",
			input: invalidAddrUDPListener,
			want:  []error{ErrListenerAddressInvalid},
		},
		{
			name:  "udp invalid port",
			input: invalidPortUDPListenerT,
			want:  []error{ErrListenerPortInvalid},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.want == nil {
				require.NoError(t, test.input.Validate())
			} else {
				got := test.input.Validate()
				for _, w := range test.want {
					require.ErrorContains(t, got, w.Error())
				}
			}
		})
	}
}

func TestValidateHTTPRoute(t *testing.T) {
	tests := []struct {
		name  string
		input HTTPRoute
		want  []error
	}{
		{
			name:  "happy",
			input: happyHTTPRoute,
			want:  nil,
		},
		{
			name: "invalid name",
			input: HTTPRoute{
				Hostname: "*",
				PathMatch: &StringMatch{
					Exact: ptr.To("example"),
				},
				Destination: &happyRouteDestination,
			},
			want: []error{ErrRouteNameEmpty},
		},
		{
			name: "invalid hostname",
			input: HTTPRoute{
				Name: "invalid hostname",
				PathMatch: &StringMatch{
					Exact: ptr.To("example"),
				},
				Destination: &happyRouteDestination,
			},
			want: []error{ErrHTTPRouteHostnameEmpty},
		},
		{
			name:  "invalid backend",
			input: invalidBackendHTTPRoute,
			want:  nil,
		},
		{
			name:  "weighted invalid backends",
			input: weightedInvalidBackendsHTTPRoute,
			want:  nil,
		},
		{
			name: "empty name and invalid match",
			input: HTTPRoute{
				Hostname:      "*",
				HeaderMatches: []*StringMatch{ptr.To(StringMatch{})},
				Destination:   &happyRouteDestination,
			},
			want: []error{ErrRouteNameEmpty, ErrStringMatchConditionInvalid},
		},
		{
			name:  "redirect-httproute",
			input: redirectHTTPRoute,
			want:  nil,
		},
		{
			name:  "filter-error-httproute",
			input: invalidFilterHTTPRoute,
			want:  nil,
		},
		{
			name:  "redirect-bad-status-scheme-nopath",
			input: redirectFilterInvalidStatus,
			want:  []error{ErrRedirectUnsupportedStatus, ErrRedirectUnsupportedScheme, ErrHTTPPathModifierNoReplace},
		},
		{
			name:  "redirect-bad-path",
			input: redirectFilterBadPath,
			want:  []error{ErrHTTPPathModifierDoubleReplace},
		},
		{
			name:  "direct-response-bad-status",
			input: directResponseBadStatus,
			want:  []error{ErrDirectResponseStatusInvalid},
		},
		{
			name:  "rewrite-httproute",
			input: urlRewriteHTTPRoute,
			want:  nil,
		},
		{
			name:  "rewrite-bad-path",
			input: urlRewriteFilterBadPath,
			want:  []error{ErrHTTPPathModifierDoubleReplace},
		},
		{
			name:  "add-request-headers-httproute",
			input: addRequestHeaderHTTPRoute,
			want:  nil,
		},
		{
			name:  "remove-request-headers-httproute",
			input: removeRequestHeaderHTTPRoute,
			want:  nil,
		},
		{
			name:  "add-remove-request-headers-duplicate",
			input: addAndRemoveRequestHeadersDupeHTTPRoute,
			want:  []error{ErrAddHeaderDuplicate, ErrRemoveHeaderDuplicate},
		},
		{
			name:  "add-request-header-empty",
			input: addRequestHeaderEmptyHTTPRoute,
			want:  []error{ErrAddHeaderEmptyName},
		},
		{
			name:  "add-response-headers-httproute",
			input: addResponseHeaderHTTPRoute,
			want:  nil,
		},
		{
			name:  "remove-response-headers-httproute",
			input: removeResponseHeaderHTTPRoute,
			want:  nil,
		},
		{
			name:  "add-remove-response-headers-duplicate",
			input: addAndRemoveResponseHeadersDupeHTTPRoute,
			want:  []error{ErrAddHeaderDuplicate, ErrRemoveHeaderDuplicate},
		},
		{
			name:  "add-response-header-empty",
			input: addResponseHeaderEmptyHTTPRoute,
			want:  []error{ErrAddHeaderEmptyName},
		},
		{
			name:  "jwt-authen-httproute",
			input: jwtAuthenHTTPRoute,
		},
		{
			name:  "mirror-filter",
			input: requestMirrorFilter,
			want:  nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.want == nil {
				require.NoError(t, test.input.Validate())
			} else {
				got := test.input.Validate()
				for _, w := range test.want {
					require.ErrorContains(t, got, w.Error())
				}
			}
		})
	}
}

func TestValidateTCPRoute(t *testing.T) {
	tests := []struct {
		name  string
		input TCPRoute
		want  []error
	}{
		{
			name:  "tls passthrough happy",
			input: happyTCPRouteTLSPassthrough,
			want:  nil,
		},
		{
			name:  "tls terminatation happy",
			input: happyTCPRouteTLSTermination,
			want:  nil,
		},
		{
			name:  "empty sni",
			input: emptySNITCPRoute,
			want:  nil,
		},
		{
			name:  "invalid sni",
			input: invalidSNITCPRoute,
			want:  []error{ErrTCPRouteSNIsEmpty},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.want == nil {
				require.NoError(t, test.input.Validate())
			} else {
				got := test.input.Validate()
				for _, w := range test.want {
					assert.ErrorContains(t, got, w.Error())
				}
			}
		})
	}
}

func TestValidateRouteDestination(t *testing.T) {
	tests := []struct {
		name  string
		input RouteDestination
		want  error
	}{
		{
			name:  "happy",
			input: happyRouteDestination,
			want:  nil,
		},
		{
			name: "valid hostname",
			input: RouteDestination{
				Name: "valid hostname",
				Settings: []*DestinationSetting{
					{
						Endpoints: []*DestinationEndpoint{
							{
								Host: "example.com",
								Port: 8080,
							},
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "valid ip",
			input: RouteDestination{
				Name: "valid ip",
				Settings: []*DestinationSetting{
					{
						Endpoints: []*DestinationEndpoint{
							{
								Host: "1.2.3.4",
								Port: 8080,
							},
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "invalid address",
			input: RouteDestination{
				Name: "invalid address",
				Settings: []*DestinationSetting{
					{
						Endpoints: []*DestinationEndpoint{
							{
								Host: "example.com::foo.bar",
								Port: 8080,
							},
						},
					},
				},
			},
			want: ErrDestEndpointHostInvalid,
		},
		{
			name: "missing ip",
			input: RouteDestination{
				Name: "missing ip",
				Settings: []*DestinationSetting{
					{
						Endpoints: []*DestinationEndpoint{
							{
								Port: 8080,
							},
						},
					},
				},
			},
			want: ErrDestEndpointHostInvalid,
		},
		{
			name: "missing port",
			input: RouteDestination{
				Name: "missing port",
				Settings: []*DestinationSetting{
					{
						Endpoints: []*DestinationEndpoint{
							{
								Host: "10.11.12.13",
							},
						},
					},
				},
			},
			want: ErrDestEndpointPortInvalid,
		},
		{
			name: "missing name",
			input: RouteDestination{
				Settings: []*DestinationSetting{
					{
						Endpoints: []*DestinationEndpoint{
							{
								Host: "10.11.12.13",
								Port: 8080,
							},
						},
					},
				},
			},
			want: ErrDestinationNameEmpty,
		},
		{
			name: "mixed address types with MIXED in destinations",
			input: RouteDestination{
				Settings: []*DestinationSetting{
					{
						Endpoints: []*DestinationEndpoint{
							{
								Host: "10.11.12.13",
								Port: 8080,
							},
							{
								Host: "example.com",
								Port: 8080,
							},
						},
						AddressType: ptr.To(MIXED),
					},
				},
			},
			want: ErrRouteDestinationsFQDNMixed,
		},
		{
			name: "mixed address types with FQDN in destinations",
			input: RouteDestination{
				Settings: []*DestinationSetting{
					{
						Endpoints: []*DestinationEndpoint{
							{
								Host: "10.11.12.13",
								Port: 8080,
							},
						},
						AddressType: ptr.To(IP),
					},
					{
						Endpoints: []*DestinationEndpoint{
							{
								Host: "example.com",
								Port: 8080,
							},
						},
						AddressType: ptr.To(FQDN),
					},
				},
			},
			want: ErrRouteDestinationsFQDNMixed,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.want == nil {
				require.NoError(t, test.input.Validate())
			} else {
				require.EqualError(t, test.input.Validate(), test.want.Error())
			}
		})
	}
}

func TestValidateStringMatch(t *testing.T) {
	tests := []struct {
		name  string
		input StringMatch
		want  error
	}{
		{
			name: "happy exact",
			input: StringMatch{
				Exact: ptr.To("example"),
			},
			want: nil,
		},
		{
			name: "happy distinct",
			input: StringMatch{
				Distinct: true,
				Name:     "example",
			},
			want: nil,
		},
		{
			name:  "no fields set",
			input: StringMatch{},
			want:  ErrStringMatchConditionInvalid,
		},
		{
			name: "multiple fields set",
			input: StringMatch{
				Exact:  ptr.To("example"),
				Name:   "example",
				Prefix: ptr.To("example"),
			},
			want: ErrStringMatchConditionInvalid,
		},
		{
			name: "both invert and distinct fields are set",
			input: StringMatch{
				Distinct: true,
				Name:     "example",
				Invert:   ptr.To(true),
			},
			want: ErrStringMatchInvertDistinctInvalid,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.want == nil {
				require.NoError(t, test.input.Validate())
			} else {
				require.EqualError(t, test.input.Validate(), test.want.Error())
			}
		})
	}
}

func TestValidateLoadBalancer(t *testing.T) {
	tests := []struct {
		name  string
		input LoadBalancer
		want  error
	}{
		{
			name: "random",
			input: LoadBalancer{
				Random: &Random{},
			},
			want: nil,
		},
		{
			name: "consistent hash with source IP hash policy",
			input: LoadBalancer{
				ConsistentHash: &ConsistentHash{
					SourceIP: ptr.To(true),
				},
			},
			want: nil,
		},
		{
			name: "consistent hash with header hash policy",
			input: LoadBalancer{
				ConsistentHash: &ConsistentHash{
					Header: &Header{
						Name: "name",
					},
				},
			},
			want: nil,
		},
		{
			name: "least request and random set",
			input: LoadBalancer{
				Random:       &Random{},
				LeastRequest: &LeastRequest{},
			},
			want: ErrLoadBalancerInvalid,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			if test.want == nil {
				require.NoError(t, test.input.Validate())
			} else {
				require.EqualError(t, test.input.Validate(), test.want.Error())
			}
		})
	}
}

func TestRedaction(t *testing.T) {
	tests := []struct {
		name    string
		input   Xds
		want    *Xds
		wantStr string
	}{
		{
			name:  "empty",
			input: Xds{},
			want:  &Xds{},
		},
		{
			name: "http",
			input: Xds{
				HTTP: []*HTTPListener{&happyHTTPListener},
			},
			want: &Xds{
				HTTP: []*HTTPListener{&happyHTTPListener},
			},
		},
		{
			name: "https",
			input: Xds{
				HTTP: []*HTTPListener{&happyHTTPSListener},
			},
			want: &Xds{
				HTTP: []*HTTPListener{&redactedHappyHTTPSListener},
			},
		},
		{
			name: "explicit string check",
			input: Xds{
				HTTP: []*HTTPListener{{
					TLS: &TLSConfig{
						Certificates: []TLSCertificate{{
							Name:        "server",
							Certificate: []byte("---"),
							PrivateKey:  []byte("secret"),
						}},
						ClientCertificates: []TLSCertificate{{
							Name:        "client",
							Certificate: []byte("---"),
							PrivateKey:  []byte("secret"),
						}},
					},
					Routes: []*HTTPRoute{{
						Security: &SecurityFeatures{
							OIDC: &OIDC{
								ClientSecret: []byte("secret"),
								HMACSecret:   []byte("secret"),
							},
							APIKeyAuth: &APIKeyAuth{
								Credentials: map[string]PrivateBytes{"client-id": []byte("secret")},
							},
							BasicAuth: &BasicAuth{
								Users: []byte("secret"),
							},
						},
					}},
				}},
			},
			wantStr: `{"http":[{"name":"","address":"","port":0,"hostnames":null,` +
				`"tls":{` +
				`"certificates":[{"name":"server","serverCertificate":"LS0t","privateKey":"[redacted]"}],` +
				`"clientCertificates":[{"name":"client","serverCertificate":"LS0t","privateKey":"[redacted]"}],` +
				`"alpnProtocols":null},` +
				`"routes":[{` +
				`"name":"","hostname":"","isHTTP2":false,"security":{` +
				`"oidc":{"name":"","provider":{},"clientID":"","clientSecret":"[redacted]","hmacSecret":"[redacted]"},` +
				`"apiKeyAuth":{"credentials":{"client-id":"[redacted]"},"extractFrom":null},` +
				`"basicAuth":{"name":"","users":"[redacted]"}` +
				`}}],` +
				`"isHTTP2":false,"path":{"mergeSlashes":false,"escapedSlashesAction":""}}]}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.want != nil {
				if test.wantStr != "" {
					t.Fatalf("Don't set both want and wantStr")
				}
				wantJSON, err := json.Marshal(test.want)
				require.NoError(t, err)
				test.wantStr = string(wantJSON)
			}
			assert.Equal(t, test.wantStr, test.input.JSONString())
		})
	}
}

func TestValidateHealthCheck(t *testing.T) {
	tests := []struct {
		name  string
		input HealthCheck
		want  error
	}{
		{
			name: "invalid timeout",
			input: HealthCheck{
				&ActiveHealthCheck{
					Timeout:            &metav1.Duration{Duration: time.Duration(0)},
					Interval:           &metav1.Duration{Duration: time.Second},
					UnhealthyThreshold: ptr.To[uint32](3),
					HealthyThreshold:   ptr.To[uint32](3),
					HTTP: &HTTPHealthChecker{
						Host:             "*",
						Path:             "/healthz",
						ExpectedStatuses: []HTTPStatus{200, 400},
					},
				}, &OutlierDetection{},
				ptr.To[uint32](10),
			},
			want: ErrHealthCheckTimeoutInvalid,
		},
		{
			name: "invalid panic threshold",
			input: HealthCheck{
				&ActiveHealthCheck{
					Timeout:            &metav1.Duration{Duration: time.Duration(3)},
					Interval:           &metav1.Duration{Duration: time.Second},
					UnhealthyThreshold: ptr.To[uint32](3),
					HealthyThreshold:   ptr.To[uint32](3),
					HTTP: &HTTPHealthChecker{
						Host:             "*",
						Path:             "/healthz",
						ExpectedStatuses: []HTTPStatus{200, 400},
					},
				}, &OutlierDetection{},
				ptr.To[uint32](200),
			},
			want: ErrPanicThresholdInvalid,
		},
		{
			name: "invalid interval",
			input: HealthCheck{
				&ActiveHealthCheck{
					Timeout:            &metav1.Duration{Duration: time.Second},
					Interval:           &metav1.Duration{Duration: time.Duration(0)},
					UnhealthyThreshold: ptr.To[uint32](3),
					HealthyThreshold:   ptr.To[uint32](3),
					HTTP: &HTTPHealthChecker{
						Host:             "*",
						Path:             "/healthz",
						Method:           ptr.To(http.MethodGet),
						ExpectedStatuses: []HTTPStatus{200, 400},
					},
				},
				&OutlierDetection{},
				ptr.To[uint32](10),
			},
			want: ErrHealthCheckIntervalInvalid,
		},
		{
			name: "invalid unhealthy threshold",
			input: HealthCheck{
				&ActiveHealthCheck{
					Timeout:            &metav1.Duration{Duration: time.Second},
					Interval:           &metav1.Duration{Duration: time.Second},
					UnhealthyThreshold: ptr.To[uint32](0),
					HealthyThreshold:   ptr.To[uint32](3),
					HTTP: &HTTPHealthChecker{
						Host:             "*",
						Path:             "/healthz",
						Method:           ptr.To(http.MethodPatch),
						ExpectedStatuses: []HTTPStatus{200, 400},
					},
				},
				&OutlierDetection{},
				ptr.To[uint32](10),
			},
			want: ErrHealthCheckUnhealthyThresholdInvalid,
		},
		{
			name: "invalid healthy threshold",
			input: HealthCheck{
				&ActiveHealthCheck{
					Timeout:            &metav1.Duration{Duration: time.Second},
					Interval:           &metav1.Duration{Duration: time.Second},
					UnhealthyThreshold: ptr.To[uint32](3),
					HealthyThreshold:   ptr.To[uint32](0),
					HTTP: &HTTPHealthChecker{
						Host:             "*",
						Path:             "/healthz",
						Method:           ptr.To(http.MethodPost),
						ExpectedStatuses: []HTTPStatus{200, 400},
					},
				},
				&OutlierDetection{},
				ptr.To[uint32](10),
			},
			want: ErrHealthCheckHealthyThresholdInvalid,
		},
		{
			name: "http-health-check: invalid host",
			input: HealthCheck{
				&ActiveHealthCheck{
					Timeout:            &metav1.Duration{Duration: time.Second},
					Interval:           &metav1.Duration{Duration: time.Second},
					UnhealthyThreshold: ptr.To[uint32](3),
					HealthyThreshold:   ptr.To[uint32](3),
					HTTP: &HTTPHealthChecker{
						Path:             "/healthz",
						Method:           ptr.To(http.MethodPut),
						ExpectedStatuses: []HTTPStatus{200, 400},
					},
				},
				&OutlierDetection{},
				ptr.To[uint32](10),
			},
			want: ErrHCHTTPHostInvalid,
		},
		{
			name: "http-health-check: invalid path",
			input: HealthCheck{
				&ActiveHealthCheck{
					Timeout:            &metav1.Duration{Duration: time.Second},
					Interval:           &metav1.Duration{Duration: time.Second},
					UnhealthyThreshold: ptr.To[uint32](3),
					HealthyThreshold:   ptr.To[uint32](3),
					HTTP: &HTTPHealthChecker{
						Host:             "*",
						Path:             "",
						Method:           ptr.To(http.MethodPut),
						ExpectedStatuses: []HTTPStatus{200, 400},
					},
				},
				&OutlierDetection{},
				ptr.To[uint32](10),
			},
			want: ErrHCHTTPPathInvalid,
		},
		{
			name: "http-health-check: invalid method",
			input: HealthCheck{
				&ActiveHealthCheck{
					Timeout:            &metav1.Duration{Duration: time.Second},
					Interval:           &metav1.Duration{Duration: time.Second},
					UnhealthyThreshold: ptr.To(uint32(3)),
					HealthyThreshold:   ptr.To(uint32(3)),
					HTTP: &HTTPHealthChecker{
						Host:             "*",
						Path:             "/healthz",
						Method:           ptr.To(http.MethodConnect),
						ExpectedStatuses: []HTTPStatus{200, 400},
					},
				},
				&OutlierDetection{},
				ptr.To[uint32](10),
			},
			want: ErrHCHTTPMethodInvalid,
		},
		{
			name: "http-health-check: invalid expected-statuses",
			input: HealthCheck{
				&ActiveHealthCheck{
					Timeout:            &metav1.Duration{Duration: time.Second},
					Interval:           &metav1.Duration{Duration: time.Second},
					UnhealthyThreshold: ptr.To(uint32(3)),
					HealthyThreshold:   ptr.To(uint32(3)),
					HTTP: &HTTPHealthChecker{
						Host:             "*",
						Path:             "/healthz",
						Method:           ptr.To(http.MethodDelete),
						ExpectedStatuses: []HTTPStatus{},
					},
				},
				&OutlierDetection{},
				ptr.To[uint32](10),
			},
			want: ErrHCHTTPExpectedStatusesInvalid,
		},
		{
			name: "http-health-check: invalid range",
			input: HealthCheck{
				&ActiveHealthCheck{
					Timeout:            &metav1.Duration{Duration: time.Second},
					Interval:           &metav1.Duration{Duration: time.Second},
					UnhealthyThreshold: ptr.To(uint32(3)),
					HealthyThreshold:   ptr.To(uint32(3)),
					HTTP: &HTTPHealthChecker{
						Host:             "*",
						Path:             "/healthz",
						Method:           ptr.To(http.MethodHead),
						ExpectedStatuses: []HTTPStatus{100, 600},
					},
				},
				&OutlierDetection{},
				ptr.To[uint32](10),
			},
			want: ErrHTTPStatusInvalid,
		},
		{
			name: "http-health-check: invalid expected-responses",
			input: HealthCheck{
				&ActiveHealthCheck{
					Timeout:            &metav1.Duration{Duration: time.Second},
					Interval:           &metav1.Duration{Duration: time.Second},
					UnhealthyThreshold: ptr.To(uint32(3)),
					HealthyThreshold:   ptr.To(uint32(3)),
					HTTP: &HTTPHealthChecker{
						Host:             "*",
						Path:             "/healthz",
						Method:           ptr.To(http.MethodOptions),
						ExpectedStatuses: []HTTPStatus{200, 300},
						ExpectedResponse: &HealthCheckPayload{
							Text:   ptr.To("foo"),
							Binary: []byte{'f', 'o', 'o'},
						},
					},
				},
				&OutlierDetection{},
				ptr.To[uint32](10),
			},
			want: ErrHealthCheckPayloadInvalid,
		},
		{
			name: "tcp-health-check: invalid send payload",
			input: HealthCheck{
				&ActiveHealthCheck{
					Timeout:            &metav1.Duration{Duration: time.Second},
					Interval:           &metav1.Duration{Duration: time.Second},
					UnhealthyThreshold: ptr.To(uint32(3)),
					HealthyThreshold:   ptr.To(uint32(3)),
					TCP: &TCPHealthChecker{
						Send: &HealthCheckPayload{
							Text:   ptr.To("foo"),
							Binary: []byte{'f', 'o', 'o'},
						},
						Receive: &HealthCheckPayload{
							Text: ptr.To("foo"),
						},
					},
				},
				&OutlierDetection{},
				ptr.To[uint32](10),
			},
			want: ErrHealthCheckPayloadInvalid,
		},
		{
			name: "tcp-health-check: invalid receive payload",
			input: HealthCheck{
				&ActiveHealthCheck{
					Timeout:            &metav1.Duration{Duration: time.Second},
					Interval:           &metav1.Duration{Duration: time.Second},
					UnhealthyThreshold: ptr.To(uint32(3)),
					HealthyThreshold:   ptr.To(uint32(3)),
					TCP: &TCPHealthChecker{
						Send: &HealthCheckPayload{
							Text: ptr.To("foo"),
						},
						Receive: &HealthCheckPayload{
							Text:   ptr.To("foo"),
							Binary: []byte{'f', 'o', 'o'},
						},
					},
				},
				&OutlierDetection{},
				ptr.To[uint32](10),
			},
			want: ErrHealthCheckPayloadInvalid,
		},
		{
			name: "OutlierDetection invalid interval",
			input: HealthCheck{
				&ActiveHealthCheck{},
				&OutlierDetection{
					Interval:         &metav1.Duration{Duration: time.Duration(0)},
					BaseEjectionTime: &metav1.Duration{Duration: time.Second},
				},
				ptr.To[uint32](10),
			},
			want: ErrOutlierDetectionIntervalInvalid,
		},
		{
			name: "OutlierDetection invalid BaseEjectionTime",
			input: HealthCheck{
				&ActiveHealthCheck{},
				&OutlierDetection{
					Interval:         &metav1.Duration{Duration: time.Second},
					BaseEjectionTime: &metav1.Duration{Duration: time.Duration(0)},
				},
				ptr.To[uint32](10),
			},
			want: ErrOutlierDetectionBaseEjectionTimeInvalid,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			if test.want == nil {
				require.NoError(t, test.input.Validate())
			} else {
				require.EqualError(t, test.input.Validate(), test.want.Error())
			}
		})
	}
}

func TestJSONPatchOperationValidation(t *testing.T) {
	tests := []struct {
		name  string
		input JSONPatchOperation
		want  *string
	}{
		{
			name: "no path or jsonpath",
			input: JSONPatchOperation{
				Op: TranslateJSONPatchOp(egv1a1.JSONPatchOperationType("remove")),
			},
			want: ptr.To("a patch operation must specify a path or jsonPath"),
		},
		{
			name: "replace with from",
			input: JSONPatchOperation{
				Op:       TranslateJSONPatchOp(egv1a1.JSONPatchOperationType("replace")),
				JSONPath: ptr.To("$.some.json[@?name=='lala'].key"),
				Value: &apiextensionsv1.JSON{
					Raw: []byte{},
				},
				From: ptr.To("/some/from"),
			},
			want: ptr.To("the replace operation doesn't support a from attribute"),
		},
		{
			name: "add with no value",
			input: JSONPatchOperation{
				Op:       TranslateJSONPatchOp(egv1a1.JSONPatchOperationType("add")),
				JSONPath: ptr.To("$.some.json[@?name=='lala'].key"),
			},
			want: ptr.To("the add operation requires a value"),
		},
		{
			name: "remove with from",
			input: JSONPatchOperation{
				Op:       TranslateJSONPatchOp(egv1a1.JSONPatchOperationType("remove")),
				JSONPath: ptr.To("$.some.json[@?name=='lala'].key"),
				From:     ptr.To("/some/from"),
			},
			want: ptr.To("value and from can't be specified with the remove operation"),
		},
		{
			name: "remove with value",
			input: JSONPatchOperation{
				Op:       TranslateJSONPatchOp(egv1a1.JSONPatchOperationType("remove")),
				JSONPath: ptr.To("$.some.json[@?name=='lala'].key"),
				Value: &apiextensionsv1.JSON{
					Raw: []byte{},
				},
			},
			want: ptr.To("value and from can't be specified with the remove operation"),
		},
		{
			name: "move without from",
			input: JSONPatchOperation{
				Op:       TranslateJSONPatchOp(egv1a1.JSONPatchOperationType("move")),
				JSONPath: ptr.To("$.some.json[@?name=='lala'].key"),
			},
			want: ptr.To("the move operation requires a valid from attribute"),
		},
		{
			name: "copy with value",
			input: JSONPatchOperation{
				Op:       TranslateJSONPatchOp(egv1a1.JSONPatchOperationType("copy")),
				JSONPath: ptr.To("$.some.json[@?name=='lala'].key"),
				From:     ptr.To("/some/from"),
				Value: &apiextensionsv1.JSON{
					Raw: []byte{},
				},
			},
			want: ptr.To("the copy operation doesn't support a value attribute"),
		},
		{
			name: "invalid operation",
			input: JSONPatchOperation{
				Op:   TranslateJSONPatchOp(egv1a1.JSONPatchOperationType("invalid")),
				Path: ptr.To("/some/path"),
			},
			want: ptr.To("unsupported JSONPatch operation"),
		},
		{
			name: "valid test operation",
			input: JSONPatchOperation{
				Op:   TranslateJSONPatchOp(egv1a1.JSONPatchOperationType("test")),
				Path: ptr.To("/some/path"),
				Value: &apiextensionsv1.JSON{
					Raw: []byte{},
				},
			},
		},
		{
			name: "valid remove operation",
			input: JSONPatchOperation{
				Op:   TranslateJSONPatchOp(egv1a1.JSONPatchOperationType("remove")),
				Path: ptr.To("/some/path"),
			},
		},
		{
			name: "valid copy operation",
			input: JSONPatchOperation{
				Op:   TranslateJSONPatchOp(egv1a1.JSONPatchOperationType("copy")),
				Path: ptr.To("/some/path"),
				From: ptr.To("/some/other/path"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Validate()
			if tc.want != nil {
				require.EqualError(t, err, *tc.want)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
