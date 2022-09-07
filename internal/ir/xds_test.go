package ir

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// HTTPListener
	happyHTTPListener = HTTPListener{
		Name:      "happy",
		Address:   "0.0.0.0",
		Port:      80,
		Hostnames: []string{"example.com"},
		Routes:    []*HTTPRoute{&happyHTTPRoute},
	}
	invalidAddrHTTPListener = HTTPListener{
		Name:      "invalid-addr",
		Address:   "1.0.0",
		Port:      80,
		Hostnames: []string{"example.com"},
		Routes:    []*HTTPRoute{&happyHTTPRoute},
	}
	invalidRouteMatchHTTPListener = HTTPListener{
		Name:      "invalid-route-match",
		Address:   "0.0.0.0",
		Port:      80,
		Hostnames: []string{"example.com"},
		Routes:    []*HTTPRoute{&emptyMatchHTTPRoute},
	}

	// HTTPRoute
	happyHTTPRoute = HTTPRoute{
		Name: "happy",
		PathMatch: &StringMatch{
			Exact: ptrTo("example"),
		},
		Destinations: []*RouteDestination{&happyRouteDestination},
	}
	emptyMatchHTTPRoute = HTTPRoute{
		Name:         "empty-match",
		Destinations: []*RouteDestination{&happyRouteDestination},
	}

	redirectHTTPRoute = HTTPRoute{
		Name: "redirect",
		PathMatch: &StringMatch{
			Exact: ptrTo("redirect"),
		},
		Redirect: &Redirect{
			Scheme:   ptrTo("https"),
			Hostname: ptrTo("redirect.example.com"),
			Path: &HTTPPathModifier{
				FullReplace: ptrTo("/redirect"),
			},
			Port:       ptrTo(uint32(8443)),
			StatusCode: ptrTo(int32(301)),
		},
	}
	// A direct response error is used when an invalid filter type is supplied
	invalidFilterHTTPRoute = HTTPRoute{
		Name: "filter-error",
		PathMatch: &StringMatch{
			Exact: ptrTo("filter-error"),
		},
		DirectResponse: &DirectResponse{
			Body:       ptrTo("invalid filter type"),
			StatusCode: uint32(500),
		},
	}

	redirectFilterInvalidStatus = HTTPRoute{
		Name: "redirect-bad-status-scheme-nopat",
		PathMatch: &StringMatch{
			Exact: ptrTo("redirect"),
		},
		Redirect: &Redirect{
			Scheme:     ptrTo("err"),
			Hostname:   ptrTo("redirect.example.com"),
			Path:       &HTTPPathModifier{},
			Port:       ptrTo(uint32(8443)),
			StatusCode: ptrTo(int32(305)),
		},
	}
	redirectFilterBadPath = HTTPRoute{
		Name: "redirect",
		PathMatch: &StringMatch{
			Exact: ptrTo("redirect"),
		},
		Redirect: &Redirect{
			Scheme:   ptrTo("https"),
			Hostname: ptrTo("redirect.example.com"),
			Path: &HTTPPathModifier{
				FullReplace:        ptrTo("/redirect"),
				PrefixMatchReplace: ptrTo("/redirect"),
			},
			Port:       ptrTo(uint32(8443)),
			StatusCode: ptrTo(int32(301)),
		},
	}
	directResponseBadStatus = HTTPRoute{
		Name: "redirect",
		PathMatch: &StringMatch{
			Exact: ptrTo("redirect"),
		},
		DirectResponse: &DirectResponse{
			Body:       ptrTo("invalid filter type"),
			StatusCode: uint32(799),
		},
	}

	// RouteDestination
	happyRouteDestination = RouteDestination{
		Host: "10.11.12.13",
		Port: 8080,
	}
)

// Creates a pointer to any type
func ptrTo[T any](x T) *T {
	return &x
}

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
			name: "invalid listener",
			input: Xds{
				HTTP: []*HTTPListener{&happyHTTPListener, &invalidAddrHTTPListener, &invalidRouteMatchHTTPListener},
			},
			want: []error{ErrHTTPListenerAddressInvalid, ErrHTTPRouteMatchEmpty},
		},
	}
	for _, test := range tests {
		test := test
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
				Address:   "0.0.0.0",
				Port:      80,
				Hostnames: []string{"example.com"},
				Routes:    []*HTTPRoute{&happyHTTPRoute},
			},
			want: []error{ErrHTTPListenerNameEmpty},
		},
		{
			name:  "invalid addr",
			input: invalidAddrHTTPListener,
			want:  []error{ErrHTTPListenerAddressInvalid},
		},
		{
			name: "invalid port and hostnames",
			input: HTTPListener{
				Name:    "invalid-port-and-hostnames",
				Address: "1.0.0",
				Routes:  []*HTTPRoute{&happyHTTPRoute},
			},
			want: []error{ErrHTTPListenerPortInvalid, ErrHTTPListenerHostnamesEmpty},
		},
		{
			name:  "invalid route match",
			input: invalidRouteMatchHTTPListener,
			want:  []error{ErrHTTPRouteMatchEmpty},
		},
	}
	for _, test := range tests {
		test := test
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

func TestValidateTLSListenerConfig(t *testing.T) {
	tests := []struct {
		name  string
		input TLSListenerConfig
		want  error
	}{
		{
			name: "happy",
			input: TLSListenerConfig{
				ServerCertificate: []byte("server-cert"),
				PrivateKey:        []byte("priv-key"),
			},
			want: nil,
		},
		{
			name: "invalid server cert",
			input: TLSListenerConfig{
				PrivateKey: []byte("priv-key"),
			},
			want: ErrTLSServerCertEmpty,
		},
		{
			name: "invalid private key",
			input: TLSListenerConfig{
				ServerCertificate: []byte("server-cert"),
			},
			want: ErrTLSPrivateKey,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if test.want == nil {
				require.NoError(t, test.input.Validate())
			} else {
				require.EqualError(t, test.input.Validate(), test.want.Error())
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
				PathMatch: &StringMatch{
					Exact: ptrTo("example"),
				},
				Destinations: []*RouteDestination{&happyRouteDestination},
			},
			want: []error{ErrHTTPRouteNameEmpty},
		},
		{
			name:  "empty match",
			input: emptyMatchHTTPRoute,
			want:  []error{ErrHTTPRouteMatchEmpty},
		},
		{
			name: "empty name and invalid match",
			input: HTTPRoute{
				HeaderMatches: []*StringMatch{ptrTo(StringMatch{})},
				Destinations:  []*RouteDestination{&happyRouteDestination},
			},
			want: []error{ErrHTTPRouteNameEmpty, ErrStringMatchConditionInvalid},
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
	}
	for _, test := range tests {
		test := test
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
			name: "invalid ip",
			input: RouteDestination{
				Host: "example.com",
				Port: 8080,
			},
			want: ErrRouteDestinationHostInvalid,
		},
		{
			name: "invalid port",
			input: RouteDestination{
				Host: "10.11.12.13",
			},
			want: ErrRouteDestinationPortInvalid,
		},
	}
	for _, test := range tests {
		test := test
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
			name: "happy",
			input: StringMatch{
				Exact: ptrTo("example"),
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
				Exact:  ptrTo("example"),
				Name:   "example",
				Prefix: ptrTo("example"),
			},
			want: ErrStringMatchConditionInvalid,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if test.want == nil {
				require.NoError(t, test.input.Validate())
			} else {
				require.EqualError(t, test.input.Validate(), test.want.Error())
			}
		})
	}
}
