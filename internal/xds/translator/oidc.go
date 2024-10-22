// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	oauth2v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/oauth2/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&oidc{})
}

type oidc struct{}

var _ httpFilter = &oidc{}

// patchHCM builds and appends the oauth2 Filters to the HTTP Connection Manager
// if applicable, and it does not already exist.
// Note: this method creates an oauth2 filter for each route that contains an OIDC config.
// the filter is disabled by default. It is enabled on the route level.
func (*oidc) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	var errs error

	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, route := range irListener.Routes {
		if !routeContainsOIDC(route) {
			continue
		}

		// Only generates one OAuth2 Envoy filter for each unique name.
		// For example, if there are two routes under the same gateway with the
		// same OAuth2 config, only one OAuth2 filter will be generated.
		if hcmContainsFilter(mgr, oauth2FilterName(route.Security.OIDC)) {
			continue
		}

		filter, err := buildHCMOAuth2Filter(route.Security.OIDC)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return errs
}

// buildHCMOAuth2Filter returns an OAuth2 HTTP filter from the provided IR HTTPRoute.
func buildHCMOAuth2Filter(oidc *ir.OIDC) (*hcmv3.HttpFilter, error) {
	oauth2Proto, err := oauth2Config(oidc)
	if err != nil {
		return nil, err
	}

	if err := oauth2Proto.ValidateAll(); err != nil {
		return nil, err
	}

	OAuth2Any, err := anypb.New(oauth2Proto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name:     oauth2FilterName(oidc),
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: OAuth2Any,
		},
	}, nil
}

func oauth2FilterName(oidc *ir.OIDC) string {
	return perRouteFilterName(egv1a1.EnvoyFilterOAuth2, oidc.Name)
}

func oauth2Config(oidc *ir.OIDC) (*oauth2v3.OAuth2, error) {
	var (
		tokenEndpointCluster string
		err                  error
	)

	if oidc.Provider.Destination != nil && len(oidc.Provider.Destination.Settings) > 0 {
		tokenEndpointCluster = oidc.Provider.Destination.Name
	} else {
		var cluster *urlCluster
		if cluster, err = url2Cluster(oidc.Provider.TokenEndpoint); err != nil {
			return nil, err
		}
		if cluster.endpointType == EndpointTypeStatic {
			return nil, fmt.Errorf(
				"static IP cluster is not allowed: %s",
				oidc.Provider.TokenEndpoint)
		}
		tokenEndpointCluster = cluster.name
	}

	// Envoy OAuth2 filter deletes the HTTP authorization header by default, which surprises users.
	preserveAuthorizationHeader := true

	// If the user wants to forward the oauth2 access token to the upstream service,
	// we should not preserve the original authorization header.
	if oidc.ForwardAccessToken {
		preserveAuthorizationHeader = false
	}

	oauth2 := &oauth2v3.OAuth2{
		Config: &oauth2v3.OAuth2Config{
			TokenEndpoint: &corev3.HttpUri{
				Uri: oidc.Provider.TokenEndpoint,
				HttpUpstreamType: &corev3.HttpUri_Cluster{
					Cluster: tokenEndpointCluster,
				},
				Timeout: &durationpb.Duration{
					Seconds: defaultExtServiceRequestTimeout,
				},
			},
			AuthorizationEndpoint: oidc.Provider.AuthorizationEndpoint,
			RedirectUri:           oidc.RedirectURL,
			RedirectPathMatcher: &matcherv3.PathMatcher{
				Rule: &matcherv3.PathMatcher_Path{
					Path: &matcherv3.StringMatcher{
						MatchPattern: &matcherv3.StringMatcher_Exact{
							Exact: oidc.RedirectPath,
						},
					},
				},
			},
			SignoutPath: &matcherv3.PathMatcher{
				Rule: &matcherv3.PathMatcher_Path{
					Path: &matcherv3.StringMatcher{
						MatchPattern: &matcherv3.StringMatcher_Exact{
							Exact: oidc.LogoutPath,
						},
					},
				},
			},
			UseRefreshToken:    &wrappers.BoolValue{Value: oidc.RefreshToken},
			ForwardBearerToken: oidc.ForwardAccessToken,
			Credentials: &oauth2v3.OAuth2Credentials{
				ClientId: oidc.ClientID,
				TokenSecret: &tlsv3.SdsSecretConfig{
					Name:      oauth2ClientSecretName(oidc),
					SdsConfig: makeConfigSource(),
				},
				TokenFormation: &oauth2v3.OAuth2Credentials_HmacSecret{
					HmacSecret: &tlsv3.SdsSecretConfig{
						Name:      oauth2HMACSecretName(oidc),
						SdsConfig: makeConfigSource(),
					},
				},
				CookieNames: &oauth2v3.OAuth2Credentials_CookieNames{
					BearerToken:  fmt.Sprintf("AccessToken-%s", oidc.CookieSuffix),
					OauthHmac:    fmt.Sprintf("OauthHMAC-%s", oidc.CookieSuffix),
					OauthExpires: fmt.Sprintf("OauthExpires-%s", oidc.CookieSuffix),
					IdToken:      fmt.Sprintf("IdToken-%s", oidc.CookieSuffix),
					RefreshToken: fmt.Sprintf("RefreshToken-%s", oidc.CookieSuffix),
					OauthNonce:   fmt.Sprintf("OauthNonce-%s", oidc.CookieSuffix),
				},
			},
			// every OIDC provider supports basic auth
			AuthType:   oauth2v3.OAuth2Config_BASIC_AUTH,
			AuthScopes: oidc.Scopes,
			Resources:  oidc.Resources,

			PreserveAuthorizationHeader: preserveAuthorizationHeader,
		},
	}

	if oidc.DefaultTokenTTL != nil {
		oauth2.Config.DefaultExpiresIn = &durationpb.Duration{
			Seconds: int64(oidc.DefaultTokenTTL.Seconds()),
		}
	}

	if oidc.DefaultRefreshTokenTTL != nil {
		oauth2.Config.DefaultRefreshTokenExpiresIn = &durationpb.Duration{
			Seconds: int64(oidc.DefaultRefreshTokenTTL.Seconds()),
		}
	}

	if oidc.CookieNameOverrides != nil &&
		oidc.CookieNameOverrides.AccessToken != nil {
		oauth2.Config.Credentials.CookieNames.BearerToken = *oidc.CookieNameOverrides.AccessToken
	}

	if oidc.CookieNameOverrides != nil &&
		oidc.CookieNameOverrides.IDToken != nil {
		oauth2.Config.Credentials.CookieNames.IdToken = *oidc.CookieNameOverrides.IDToken
	}

	if oidc.CookieDomain != nil {
		oauth2.Config.Credentials.CookieDomain = *oidc.CookieDomain
	}

	// Set the retry policy if it exists.
	if oidc.Provider.Traffic != nil && oidc.Provider.Traffic.Retry != nil {
		var rp *corev3.RetryPolicy
		if rp, err = buildNonRouteRetryPolicy(oidc.Provider.Traffic.Retry); err != nil {
			return nil, err
		}
		oauth2.Config.RetryPolicy = rp
	}
	return oauth2, nil
}

func buildNonRouteRetryPolicy(rr *ir.Retry) (*corev3.RetryPolicy, error) {
	rp := &corev3.RetryPolicy{
		RetryOn: retryDefaultRetryOn,
	}

	// These two fields in the RetryPolicy are just for route-level retries, they are not used for non-route retries.
	// retry.PerRetry.Timeout
	// retry.RetryOn.HTTPStatusCodes

	if rr.PerRetry != nil && rr.PerRetry.BackOff != nil {
		rp.RetryBackOff = &corev3.BackoffStrategy{
			BaseInterval: &durationpb.Duration{
				Seconds: int64(rr.PerRetry.BackOff.BaseInterval.Seconds()),
			},
			MaxInterval: &durationpb.Duration{
				Seconds: int64(rr.PerRetry.BackOff.MaxInterval.Seconds()),
			},
		}
	}

	if rr.NumRetries != nil {
		rp.NumRetries = &wrappers.UInt32Value{
			Value: *rr.NumRetries,
		}
	}

	if rr.RetryOn != nil {
		if len(rr.RetryOn.Triggers) > 0 {
			if ro, err := buildRetryOn(rr.RetryOn.Triggers); err == nil {
				rp.RetryOn = ro
			} else {
				return nil, err
			}
		}
	}
	return rp, nil
}

// routeContainsOIDC returns true if OIDC exists for the provided route.
func routeContainsOIDC(irRoute *ir.HTTPRoute) bool {
	if irRoute != nil &&
		irRoute.Security != nil &&
		irRoute.Security.OIDC != nil {
		return true
	}
	return false
}

func (*oidc) patchResources(tCtx *types.ResourceVersionTable,
	routes []*ir.HTTPRoute,
) error {
	if err := createOAuthServerClusters(tCtx, routes); err != nil {
		return err
	}
	if err := createOAuth2Secrets(tCtx, routes); err != nil {
		return err
	}
	return nil
}

// createOAuthServerClusters creates clusters for the OAuth2 server.
func createOAuthServerClusters(tCtx *types.ResourceVersionTable,
	routes []*ir.HTTPRoute,
) error {
	if tCtx == nil || tCtx.XdsResources == nil {
		return errors.New("xds resource table is nil")
	}

	var errs error
	for _, route := range routes {
		if !routeContainsOIDC(route) {
			continue
		}

		oidc := route.Security.OIDC

		// If the OIDC provider has a destination, use it.
		if oidc.Provider.Destination != nil && len(oidc.Provider.Destination.Settings) > 0 {
			if err := createExtServiceXDSCluster(
				oidc.Provider.Destination, oidc.Provider.Traffic, tCtx); err != nil && !errors.Is(
				err, ErrXdsClusterExists) {
				errs = errors.Join(errs, err)
			}
		} else {
			// Create a cluster with the token endpoint url.
			if err := createOAuth2TokenEndpointCluster(tCtx, oidc.Provider.TokenEndpoint); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	return errs
}

// createOAuth2TokenEndpointClusters creates token endpoint clusters from the
// provided routes, if needed.
func createOAuth2TokenEndpointCluster(tCtx *types.ResourceVersionTable,
	tokenEndpoint string,
) error {
	var (
		cluster *urlCluster
		ds      *ir.DestinationSetting
		tSocket *corev3.TransportSocket
		err     error
	)

	if cluster, err = url2Cluster(tokenEndpoint); err != nil {
		return err
	}

	// EG does not support static IP clusters for token endpoint clusters.
	// This validation could be removed since it's already validated in the
	// Gateway API translator.
	if cluster.endpointType == EndpointTypeStatic {
		return fmt.Errorf(
			"static IP cluster is not allowed: %s",
			tokenEndpoint)
	}

	ds = &ir.DestinationSetting{
		Weight: ptr.To[uint32](1),
		Endpoints: []*ir.DestinationEndpoint{
			ir.NewDestEndpoint(
				cluster.hostname,
				cluster.port),
		},
	}

	clusterArgs := &xdsClusterArgs{
		name:         cluster.name,
		settings:     []*ir.DestinationSetting{ds},
		tSocket:      tSocket,
		endpointType: cluster.endpointType,
	}
	if cluster.tls {
		if tSocket, err = buildXdsUpstreamTLSSocket(cluster.hostname); err != nil {
			return err
		}
		clusterArgs.tSocket = tSocket
	}

	if err = addXdsCluster(tCtx, clusterArgs); err != nil && !errors.Is(err, ErrXdsClusterExists) {
		return err
	}

	return err
}

// createOAuth2Secrets creates OAuth2 client and HMAC secrets from the provided
// routes, if needed.
func createOAuth2Secrets(tCtx *types.ResourceVersionTable, routes []*ir.HTTPRoute) error {
	var errs error

	for _, route := range routes {
		if !routeContainsOIDC(route) {
			continue
		}

		// a separate secret is created for each route, even they share the same
		// oauth2 client ID and secret.
		clientSecret := buildOAuth2ClientSecret(route.Security.OIDC)
		if err := addXdsSecret(tCtx, clientSecret); err != nil {
			errs = errors.Join(errs, err)
		}

		if err := addXdsSecret(tCtx, buildOAuth2HMACSecret(route.Security.OIDC)); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func buildOAuth2ClientSecret(oidc *ir.OIDC) *tlsv3.Secret {
	clientSecret := &tlsv3.Secret{
		Name: oauth2ClientSecretName(oidc),
		Type: &tlsv3.Secret_GenericSecret{
			GenericSecret: &tlsv3.GenericSecret{
				Secret: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineBytes{
						InlineBytes: oidc.ClientSecret,
					},
				},
			},
		},
	}

	return clientSecret
}

func buildOAuth2HMACSecret(oidc *ir.OIDC) *tlsv3.Secret {
	hmacSecret := &tlsv3.Secret{
		Name: oauth2HMACSecretName(oidc),
		Type: &tlsv3.Secret_GenericSecret{
			GenericSecret: &tlsv3.GenericSecret{
				Secret: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineBytes{
						InlineBytes: oidc.HMACSecret,
					},
				},
			},
		},
	}

	return hmacSecret
}

func oauth2ClientSecretName(oidc *ir.OIDC) string {
	return fmt.Sprintf("oauth2/client_secret/%s", oidc.Name)
}

func oauth2HMACSecretName(oidc *ir.OIDC) string {
	return fmt.Sprintf("oauth2/hmac_secret/%s", oidc.Name)
}

// patchRoute patches the provided route with the oauth2 config if applicable.
// Note: this method enables the corresponding oauth2 filter for the provided route.
func (*oidc) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.Security == nil || irRoute.Security.OIDC == nil {
		return nil
	}
	filterName := oauth2FilterName(irRoute.Security.OIDC)
	if err := enableFilterOnRoute(route, filterName); err != nil {
		return err
	}
	return nil
}
