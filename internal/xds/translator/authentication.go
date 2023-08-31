// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	jwtauthnv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/jwt_authn/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	jwtAuthenFilter  = "envoy.filters.http.jwt_authn"
	envoyTrustBundle = "/etc/ssl/certs/ca-certificates.crt"
)

// patchHCMWithJwtAuthnFilter builds and appends the Jwt Filter to the HTTP
// Connection Manager if applicable, and it does not already exist.
func patchHCMWithJwtAuthnFilter(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	if !listenerContainsJwtAuthn(irListener) {
		return nil
	}

	// Return early if filter already exists.
	for _, httpFilter := range mgr.HttpFilters {
		if httpFilter.Name == jwtAuthenFilter {
			return nil
		}
	}

	jwtFilter, err := buildHCMJwtFilter(irListener)
	if err != nil {
		return err
	}

	// Ensure the authn filter is the first and the terminal filter is the last in the chain.
	mgr.HttpFilters = append([]*hcmv3.HttpFilter{jwtFilter}, mgr.HttpFilters...)

	return nil
}

// buildHCMJwtFilter returns a JWT authn HTTP filter from the provided IR listener.
func buildHCMJwtFilter(irListener *ir.HTTPListener) (*hcmv3.HttpFilter, error) {
	jwtAuthnProto, err := buildJwtAuthn(irListener)
	if err != nil {
		return nil, err
	}

	if err := jwtAuthnProto.ValidateAll(); err != nil {
		return nil, err
	}

	jwtAuthnAny, err := anypb.New(jwtAuthnProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: jwtAuthenFilter,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: jwtAuthnAny,
		},
	}, nil
}

// buildJwtAuthn returns a JwtAuthentication based on the provided IR HTTPListener.
func buildJwtAuthn(irListener *ir.HTTPListener) (*jwtauthnv3.JwtAuthentication, error) {
	jwtProviders := make(map[string]*jwtauthnv3.JwtProvider)
	reqMap := make(map[string]*jwtauthnv3.JwtRequirement)

	for _, route := range irListener.Routes {
		if route != nil && routeContainsJwtAuthn(route) {
			var reqs []*jwtauthnv3.JwtRequirement
			for i := range route.RequestAuthentication.JWT.Providers {
				irProvider := route.RequestAuthentication.JWT.Providers[i]
				// Create the cluster for the remote jwks, if it doesn't exist.
				jwksCluster, err := newJwksCluster(&irProvider)
				if err != nil {
					return nil, err
				}

				remote := &jwtauthnv3.JwtProvider_RemoteJwks{
					RemoteJwks: &jwtauthnv3.RemoteJwks{
						HttpUri: &corev3.HttpUri{
							Uri: irProvider.RemoteJWKS.URI,
							HttpUpstreamType: &corev3.HttpUri_Cluster{
								Cluster: jwksCluster.name,
							},
							Timeout: &durationpb.Duration{Seconds: 5},
						},
						CacheDuration: &durationpb.Duration{Seconds: 5 * 60},
						AsyncFetch:    &jwtauthnv3.JwksAsyncFetch{},
						RetryPolicy:   &corev3.RetryPolicy{},
					},
				}

				claimToHeaders := []*jwtauthnv3.JwtClaimToHeader{}
				for _, claimToHeader := range irProvider.ClaimToHeaders {
					claimToHeader := &jwtauthnv3.JwtClaimToHeader{HeaderName: claimToHeader.Header, ClaimName: claimToHeader.Claim}
					claimToHeaders = append(claimToHeaders, claimToHeader)
				}
				jwtProvider := &jwtauthnv3.JwtProvider{
					Issuer:              irProvider.Issuer,
					Audiences:           irProvider.Audiences,
					JwksSourceSpecifier: remote,
					PayloadInMetadata:   irProvider.Issuer,
					ClaimToHeaders:      claimToHeaders,
				}

				providerKey := fmt.Sprintf("%s/%s", route.Name, irProvider.Name)
				jwtProviders[providerKey] = jwtProvider
				reqs = append(reqs, &jwtauthnv3.JwtRequirement{
					RequiresType: &jwtauthnv3.JwtRequirement_ProviderName{
						ProviderName: providerKey,
					},
				})
			}
			if len(reqs) == 1 {
				reqMap[route.Name] = reqs[0]
			} else {
				orListReqs := &jwtauthnv3.JwtRequirement{
					RequiresType: &jwtauthnv3.JwtRequirement_RequiresAny{
						RequiresAny: &jwtauthnv3.JwtRequirementOrList{
							Requirements: reqs,
						},
					},
				}
				reqMap[route.Name] = orListReqs
			}
		}
	}

	return &jwtauthnv3.JwtAuthentication{
		RequirementMap: reqMap,
		Providers:      jwtProviders,
	}, nil
}

// buildXdsUpstreamTLSSocket returns an xDS TransportSocket that uses envoyTrustBundle
// as the CA to authenticate server certificates.
func buildXdsUpstreamTLSSocket() (*corev3.TransportSocket, error) {
	tlsCtxProto := &tlsv3.UpstreamTlsContext{
		CommonTlsContext: &tlsv3.CommonTlsContext{
			ValidationContextType: &tlsv3.CommonTlsContext_ValidationContext{
				ValidationContext: &tlsv3.CertificateValidationContext{
					TrustedCa: &corev3.DataSource{
						Specifier: &corev3.DataSource_Filename{
							Filename: envoyTrustBundle,
						},
					},
				},
			},
		},
	}

	tlsCtxAny, err := anypb.New(tlsCtxProto)
	if err != nil {
		return nil, err
	}

	return &corev3.TransportSocket{
		Name: wellknown.TransportSocketTls,
		ConfigType: &corev3.TransportSocket_TypedConfig{
			TypedConfig: tlsCtxAny,
		},
	}, nil
}

// patchRouteWithJwtConfig patches the provided route with a JWT PerRouteConfig, if the
// route doesn't contain it.
func patchRouteWithJwtConfig(route *routev3.Route, irRoute *ir.HTTPRoute) error { //nolint:unparam
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}

	filterCfg := route.GetTypedPerFilterConfig()
	if _, ok := filterCfg[jwtAuthenFilter]; !ok {
		if !routeContainsJwtAuthn(irRoute) {
			return nil
		}

		routeCfgProto := &jwtauthnv3.PerRouteConfig{
			RequirementSpecifier: &jwtauthnv3.PerRouteConfig_RequirementName{RequirementName: irRoute.Name}}

		routeCfgAny, err := anypb.New(routeCfgProto)
		if err != nil {
			return err
		}

		if filterCfg == nil {
			route.TypedPerFilterConfig = make(map[string]*anypb.Any)
		}

		route.TypedPerFilterConfig[jwtAuthenFilter] = routeCfgAny
	}

	return nil
}

type jwksCluster struct {
	name     string
	hostname string
	port     uint32
	isStatic bool
}

// createJwksClusters creates JWKS clusters from the provided routes, if needed.
func createJwksClusters(tCtx *types.ResourceVersionTable, routes []*ir.HTTPRoute) error {
	if tCtx == nil ||
		tCtx.XdsResources == nil ||
		tCtx.XdsResources[resource.ClusterType] == nil ||
		len(routes) == 0 {
		return nil
	}

	for _, route := range routes {
		if routeContainsJwtAuthn(route) {
			for i := range route.RequestAuthentication.JWT.Providers {
				provider := route.RequestAuthentication.JWT.Providers[i]
				jwks, err := newJwksCluster(&provider)
				epType := DefaultEndpointType
				if jwks.isStatic {
					epType = Static
				}
				if err != nil {
					return err
				}
				endpoints := []*ir.DestinationEndpoint{ir.NewDestEndpoint(jwks.hostname, jwks.port)}
				tSocket, err := buildXdsUpstreamTLSSocket()
				if err != nil {
					return err
				}
				if err := addXdsCluster(tCtx, addXdsClusterArgs{
					name:         jwks.name,
					endpoints:    endpoints,
					tSocket:      tSocket,
					protocol:     DefaultProtocol,
					endpointType: epType,
				}); err != nil && !errors.Is(err, ErrXdsClusterExists) {
					return err
				}
			}
		}
	}

	return nil
}

// newJwksCluster returns a jwksCluster from the provided provider.
func newJwksCluster(provider *v1alpha1.JwtAuthenticationFilterProvider) (*jwksCluster, error) {
	static := false
	if provider == nil {
		return nil, errors.New("nil provider")
	}

	u, err := url.Parse(provider.RemoteJWKS.URI)
	if err != nil {
		return nil, err
	}

	var strPort string
	switch u.Scheme {
	case "https":
		strPort = "443"
	default:
		return nil, fmt.Errorf("unsupported JWKS URI scheme %s", u.Scheme)
	}

	if u.Port() != "" {
		strPort = u.Port()
	}

	name := fmt.Sprintf("%s_%s", strings.ReplaceAll(u.Hostname(), ".", "_"), strPort)

	port, err := strconv.Atoi(strPort)
	if err != nil {
		return nil, err
	}

	if ip := net.ParseIP(u.Hostname()); ip != nil {
		if v4 := ip.To4(); v4 != nil {
			static = true
		}
	}

	return &jwksCluster{
		name:     name,
		hostname: u.Hostname(),
		port:     uint32(port),
		isStatic: static,
	}, nil
}

// listenerContainsJwtAuthn returns true if JWT authentication exists for the
// provided listener.
func listenerContainsJwtAuthn(irListener *ir.HTTPListener) bool {
	if irListener == nil {
		return false
	}

	for _, route := range irListener.Routes {
		if routeContainsJwtAuthn(route) {
			return true
		}
	}

	return false
}

// routeContainsJwtAuthn returns true if JWT authentication exists for the
// provided route.
func routeContainsJwtAuthn(irRoute *ir.HTTPRoute) bool {
	if irRoute == nil {
		return false
	}

	if irRoute != nil &&
		irRoute.RequestAuthentication != nil &&
		irRoute.RequestAuthentication.JWT != nil &&
		irRoute.RequestAuthentication.JWT.Providers != nil &&
		len(irRoute.RequestAuthentication.JWT.Providers) > 0 {
		return true
	}

	return false
}
