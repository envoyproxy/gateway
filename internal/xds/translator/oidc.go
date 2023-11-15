// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	oauth2v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/oauth2/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/tetratelabs/multierror"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/ptr"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	oauth2Filter                = "envoy.filters.http.oauth2"
	defaultTokenEndpointPort    = 443
	defaultTokenEndpointTimeout = 10
	redirectURL                 = "%REQ(x-forwarded-proto)%://%REQ(:authority)%/oauth2/callback"
	redirectPathMatcher         = "/oauth2/callback"
	defaultSignoutPath          = "/signout"
)

// patchHCMWithOAuth2Filter builds and appends the oauth2 Filters to the HTTP
// Connection Manager if applicable, and it does not already exist.
func patchHCMWithOAuth2Filter(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	var errs error

	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, route := range irListener.Routes {
		if routeContainsOIDC(route) {
			filter, err := buildHCMOAuth2Filter(route)
			if err != nil {
				errs = multierror.Append(errs, err)
			}

			// skip if the filter already exists
			for _, existingFilter := range mgr.HttpFilters {
				if filter.Name == existingFilter.Name {
					continue
				}
			}

			mgr.HttpFilters = append(mgr.HttpFilters, filter)
		}
	}

	return nil
}

// buildHCMOAuth2Filter returns an OAuth2 HTTP filter from the provided IR HTTPRoute.
func buildHCMOAuth2Filter(route *ir.HTTPRoute) (*hcmv3.HttpFilter, error) {
	oauth2Proto := oauth2Config(route)

	if err := oauth2Proto.ValidateAll(); err != nil {
		return nil, err
	}

	OAuth2Any, err := anypb.New(oauth2Proto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: oauth2FilterName(route),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: OAuth2Any,
		},
	}, nil
}

func oauth2FilterName(route *ir.HTTPRoute) string {
	return fmt.Sprintf("%s_%s", oauth2Filter, route.Name)
}

func oauth2Config(route *ir.HTTPRoute) *oauth2v3.OAuth2 {
	// Ignore the errors because we already validate the token endpoint
	// URL in the gateway API translator.
	tokenEndpointURL, _ := url.Parse(route.OIDC.Provider.TokenEndpoint)

	oauth2 := &oauth2v3.OAuth2{
		Config: &oauth2v3.OAuth2Config{
			TokenEndpoint: &corev3.HttpUri{
				Uri: route.OIDC.Provider.TokenEndpoint,
				HttpUpstreamType: &corev3.HttpUri_Cluster{
					Cluster: oauth2TokenEndpointClusterName(tokenEndpointURL),
				},
				Timeout: &duration.Duration{
					Seconds: defaultTokenEndpointTimeout,
				},
			},
			AuthorizationEndpoint: route.OIDC.Provider.AuthorizationEndpoint,
			RedirectUri:           redirectURL,
			RedirectPathMatcher: &matcherv3.PathMatcher{
				Rule: &matcherv3.PathMatcher_Path{
					Path: &matcherv3.StringMatcher{
						MatchPattern: &matcherv3.StringMatcher_Exact{
							Exact: redirectPathMatcher,
						},
					},
				},
			},
			SignoutPath: &matcherv3.PathMatcher{
				Rule: &matcherv3.PathMatcher_Path{
					Path: &matcherv3.StringMatcher{
						MatchPattern: &matcherv3.StringMatcher_Exact{
							Exact: defaultSignoutPath,
						},
					},
				},
			},
			ForwardBearerToken: true,
			Credentials: &oauth2v3.OAuth2Credentials{
				ClientId: route.OIDC.ClientID,
				TokenSecret: &tlsv3.SdsSecretConfig{
					Name:      oauth2ClientSecretName(route),
					SdsConfig: makeConfigSource(),
				},
				TokenFormation: &oauth2v3.OAuth2Credentials_HmacSecret{
					HmacSecret: &tlsv3.SdsSecretConfig{
						Name:      oauth2HMACSecretName(route),
						SdsConfig: makeConfigSource(),
					},
				},
			},
			AuthType:   oauth2v3.OAuth2Config_BASIC_AUTH, // every OIDC provider supports basic auth
			AuthScopes: route.OIDC.Scopes,
		},
	}
	return oauth2
}

// routeContainsOIDC returns true if OIDC exists for the provided route.
func routeContainsOIDC(irRoute *ir.HTTPRoute) bool {
	if irRoute == nil {
		return false
	}

	if irRoute != nil &&
		irRoute.OIDC != nil {
		return true
	}

	return false
}

// createOAuth2TokenEndpointClusters creates token endpoint clusters from the provided routes, if needed.
func createOAuth2TokenEndpointClusters(tCtx *types.ResourceVersionTable, routes []*ir.HTTPRoute) error {
	if tCtx == nil ||
		tCtx.XdsResources == nil ||
		tCtx.XdsResources[resourcev3.ClusterType] == nil ||
		len(routes) == 0 {
		return nil
	}

	for _, route := range routes {
		if !routeContainsOIDC(route) {
			continue
		}

		// Ignore the errors because we already validate the token endpoint
		// URL in the gateway API translator.
		tokenEndpointURL, _ := url.Parse(route.OIDC.Provider.TokenEndpoint)
		port := defaultTokenEndpointPort
		if tokenEndpointURL.Port() != "" {
			port, _ = strconv.Atoi(tokenEndpointURL.Port())
		}

		tlsContext := &tlsv3.UpstreamTlsContext{
			Sni: tokenEndpointURL.Hostname(),
		}

		tlsContextAny, err := anypb.New(tlsContext)
		if err != nil {
			return err
		}
		tSocket := &corev3.TransportSocket{
			Name: "envoy.transport_sockets.tls",
			ConfigType: &corev3.TransportSocket_TypedConfig{
				TypedConfig: tlsContextAny,
			},
		}

		ds := &ir.DestinationSetting{
			Weight: ptr.To(uint32(1)),
			Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(
				tokenEndpointURL.Hostname(),
				uint32(port))},
		}

		if err := addXdsCluster(tCtx, &xdsClusterArgs{
			name:         oauth2TokenEndpointClusterName(tokenEndpointURL),
			settings:     []*ir.DestinationSetting{ds},
			tSocket:      tSocket,
			endpointType: DefaultEndpointType, // TODO support static endpoint
		}); err != nil && !errors.Is(err, ErrXdsClusterExists) {
			return err
		}
	}

	return nil
}

func oauth2TokenEndpointClusterName(tokenEndpointURL *url.URL) string {
	return fmt.Sprintf("oauth2_token_endpoint_%s", tokenEndpointURL.Hostname())
}

// createOAuth2Secrets creates OAuth2 client and HMAC secrets from the provided
// routes, if needed.
func createOAuth2Secrets(tCtx *types.ResourceVersionTable, routes []*ir.HTTPRoute) error {
	var errs error

	for _, route := range routes {
		if !routeContainsOIDC(route) {
			continue
		}

		clientSecret := buildOAuth2ClientSecret(route)
		if err := tCtx.AddXdsResource(resourcev3.SecretType, clientSecret); err != nil {
			errs = multierror.Append(errs, err)
		}

		hmacSecret, err := buildOAuth2HMACSecret(route)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		if err := tCtx.AddXdsResource(resourcev3.SecretType, hmacSecret); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}

func buildOAuth2ClientSecret(route *ir.HTTPRoute) *tlsv3.Secret {
	clientSecret := &tlsv3.Secret{
		Name: oauth2ClientSecretName(route),
		Type: &tlsv3.Secret_GenericSecret{
			GenericSecret: &tlsv3.GenericSecret{
				Secret: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineBytes{
						InlineBytes: route.OIDC.ClientSecret,
					},
				},
			},
		},
	}

	return clientSecret
}

func buildOAuth2HMACSecret(route *ir.HTTPRoute) (*tlsv3.Secret, error) {
	hmac, err := generateHMACSecretKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate hmack secret key: %w", err)
	}
	hmacSecret := &tlsv3.Secret{
		Name: oauth2HMACSecretName(route),
		Type: &tlsv3.Secret_GenericSecret{
			GenericSecret: &tlsv3.GenericSecret{
				Secret: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineBytes{
						InlineBytes: hmac,
					},
				},
			},
		},
	}

	return hmacSecret, nil
}

func oauth2ClientSecretName(route *ir.HTTPRoute) string {
	return fmt.Sprintf("%s/oauth2/client_secret", route.Name)
}

func oauth2HMACSecretName(route *ir.HTTPRoute) string {
	return fmt.Sprintf("%s/oauth2/hmac_secret", route.Name)
}

func generateHMACSecretKey() ([]byte, error) {
	// Set the desired length of the secret key in bytes
	keyLength := 32 // Adjust this value as needed

	// Create a byte slice to hold the random bytes
	key := make([]byte, keyLength)

	// Read random bytes from the cryptographically secure random number generator
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// patchRouteCfgWithOAuth2Filter patches the provided route configuration with
// the oauth2 filter if applicable.
func patchRouteCfgWithOAuth2Filter(routeCfg *routev3.RouteConfiguration, irListener *ir.HTTPListener) error {
	if routeCfg == nil {
		return errors.New("route configuration is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, route := range irListener.Routes {
		if routeContainsOIDC(route) {
			perRouteFilterName := oauth2FilterName(route)
			filterCfg := routeCfg.TypedPerFilterConfig

			if _, ok := filterCfg[perRouteFilterName]; ok {
				// This should not happen since this is the only place where the oauth2
				// filter is added in a route.
				return fmt.Errorf("route config already contains oauth2 config: %+v", route)
			}

			// Disable all the filters by default. The filter will be enabled
			// per-route in the typePerFilterConfig of the route.
			routeCfgAny, err := anypb.New(&routev3.FilterConfig{
				Disabled: true,
			})
			if err != nil {
				return err
			}
			if filterCfg == nil {
				routeCfg.TypedPerFilterConfig = make(map[string]*anypb.Any)
			}

			routeCfg.TypedPerFilterConfig[perRouteFilterName] = routeCfgAny
		}
	}
	return nil
}

// patchRouteWithOAuth2 patches the provided route with the oauth2 config if
// applicable.
func patchRouteWithOAuth2(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.OIDC == nil {
		return nil
	}

	perRouteFilterName := oauth2FilterName(irRoute)
	filterCfg := route.GetTypedPerFilterConfig()
	if _, ok := filterCfg[perRouteFilterName]; ok {
		// This should not happen since this is the only place where the oauth2
		// filter is added in a route.
		return fmt.Errorf("route already contains oauth2 config: %+v", route)
	}

	// Enable the corresponding oauth2 filter for this route.
	routeCfgAny, err := anypb.New(&routev3.FilterConfig{
		Config: &anypb.Any{},
	})
	if err != nil {
		return err
	}

	if filterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	route.TypedPerFilterConfig[perRouteFilterName] = routeCfgAny

	return nil
}
