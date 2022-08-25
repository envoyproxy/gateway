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
		Name:         "happy",
		PathMatch:    &happyStringMatch,
		Destinations: []*RouteDestination{&happyRouteDestination},
	}
	emptyMatchHTTPRoute = HTTPRoute{
		Name:         "empty-match",
		Destinations: []*RouteDestination{&happyRouteDestination},
	}

	redirectScheme    = "https"
	redirectHostname  = "redirect.example.com"
	redirectPath      = "/redirect"
	redirectPort      = uint32(8443)
	redirectStatus    = int32(301)
	redirectHTTPRoute = HTTPRoute{
		Name:      "redirect",
		PathMatch: &redirectStringMatch,
		Redirect: &Redirect{
			Scheme:   &redirectScheme,
			Hostname: &redirectHostname,
			Path: &HTTPPathModifier{
				FullReplace: &redirectPath,
			},
			Port:       &redirectPort,
			StatusCode: &redirectStatus,
		},
	}
	// A direct response error is used when an invalid filter type is supplied
	errorBody              = "invalid filter type"
	invalidFilterHTTPRoute = HTTPRoute{
		Name:      "filter-error",
		PathMatch: &filterErrorStringMatch,
		DirectResponse: &DirectResponse{
			Body:       &errorBody,
			StatusCode: uint32(500),
		},
	}

	// RouteDestination
	happyRouteDestination = RouteDestination{
		Host: "10.11.12.13",
		Port: 8080,
	}
	invalidHostRouteDestination = RouteDestination{
		Host: "example.com",
		Port: 8080,
	}

	// StringMatch
	matchStr         = "example"
	happyStringMatch = StringMatch{
		Exact: &matchStr,
	}
	emptyStringMatch = StringMatch{}

	redirectStr         = "redirect"
	redirectStringMatch = StringMatch{
		Exact: &redirectStr,
	}

	filterErrorStr         = "filter-error"
	filterErrorStringMatch = StringMatch{
		Exact: &filterErrorStr,
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
			name: "invalid listener",
			input: Xds{
				HTTP: []*HTTPListener{&happyHTTPListener, &invalidAddrHTTPListener, &invalidRouteMatchHTTPListener},
			},
			want: []error{ErrHTTPListenerAddressInvalid, ErrHTTPRouteMatchEmpty},
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
				PathMatch:    &happyStringMatch,
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
				HeaderMatches: []*StringMatch{&emptyStringMatch},
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
			name:  "invalid ip",
			input: invalidHostRouteDestination,
			want:  ErrRouteDestinationHostInvalid,
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
				Exact: &matchStr,
			},
			want: nil,
		},
		{
			name:  "no fields set",
			input: emptyStringMatch,
			want:  ErrStringMatchConditionInvalid,
		},
		{
			name: "multiple fields set",
			input: StringMatch{
				Exact:  &matchStr,
				Name:   matchStr,
				Prefix: &matchStr,
			},
			want: ErrStringMatchConditionInvalid,
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
