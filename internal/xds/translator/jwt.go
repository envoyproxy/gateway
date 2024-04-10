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
	jwtauthnv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/jwt_authn/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	jwtAuthn         = "envoy.filters.http.jwt_authn"
	envoyTrustBundle = "/etc/ssl/certs/ca-certificates.crt"
)

func init() {
	registerHTTPFilter(&jwt{})
}

type jwt struct {
}

var _ httpFilter = &jwt{}

// patchHCM builds and appends the JWT Filter to the HTTP Connection Manager if
// applicable, and it does not already exist.
func (*jwt) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	if !listenerContainsJWTAuthn(irListener) {
		return nil
	}

	// Return early if filter already exists.
	for _, httpFilter := range mgr.HttpFilters {
		if httpFilter.Name == jwtAuthn {
			return nil
		}
	}

	jwtFilter, err := buildHCMJWTFilter(irListener)
	if err != nil {
		return err
	}

	// Ensure the authn filter is the first and the terminal filter is the last in the chain.
	mgr.HttpFilters = append([]*hcmv3.HttpFilter{jwtFilter}, mgr.HttpFilters...)

	return nil
}

// buildHCMJWTFilter returns a JWT authn HTTP filter from the provided IR listener.
func buildHCMJWTFilter(irListener *ir.HTTPListener) (*hcmv3.HttpFilter, error) {
	jwtAuthnProto, err := buildJWTAuthn(irListener)
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
		Name: jwtAuthn,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: jwtAuthnAny,
		},
	}, nil
}

// buildJWTAuthn returns a JwtAuthentication based on the provided IR HTTPListener.
func buildJWTAuthn(irListener *ir.HTTPListener) (*jwtauthnv3.JwtAuthentication, error) {
	jwtProviders := make(map[string]*jwtauthnv3.JwtProvider)
	reqMap := make(map[string]*jwtauthnv3.JwtRequirement)

	for _, route := range irListener.Routes {
		if route == nil || !routeContainsJWTAuthn(route) {
			continue
		}

		var reqs []*jwtauthnv3.JwtRequirement
		for i := range route.JWT.Providers {
			irProvider := route.JWT.Providers[i]
			// Create the cluster for the remote jwks, if it doesn't exist.
			jwksCluster, err := url2Cluster(irProvider.RemoteJWKS.URI)
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
						Timeout: &durationpb.Duration{Seconds: defaultExtServiceRequestTimeout},
					},
					CacheDuration: &durationpb.Duration{Seconds: 5 * 60},
					AsyncFetch:    &jwtauthnv3.JwksAsyncFetch{},
					RetryPolicy:   &corev3.RetryPolicy{},
				},
			}

			claimToHeaders := []*jwtauthnv3.JwtClaimToHeader{}
			for _, claimToHeader := range irProvider.ClaimToHeaders {
				claimToHeader := &jwtauthnv3.JwtClaimToHeader{
					HeaderName: claimToHeader.Header,
					ClaimName:  claimToHeader.Claim}
				claimToHeaders = append(claimToHeaders, claimToHeader)
			}
			jwtProvider := &jwtauthnv3.JwtProvider{
				Issuer:              irProvider.Issuer,
				Audiences:           irProvider.Audiences,
				JwksSourceSpecifier: remote,
				PayloadInMetadata:   irProvider.Issuer,
				ClaimToHeaders:      claimToHeaders,
				Forward:             true,
			}

			if irProvider.RecomputeRoute != nil {
				jwtProvider.ClearRouteCache = *irProvider.RecomputeRoute
			}

			if irProvider.ExtractFrom != nil {
				jwtProvider.FromHeaders = buildJwtFromHeaders(irProvider.ExtractFrom.Headers)
				jwtProvider.FromCookies = irProvider.ExtractFrom.Cookies
				jwtProvider.FromParams = irProvider.ExtractFrom.Params
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

	return &jwtauthnv3.JwtAuthentication{
		RequirementMap: reqMap,
		Providers:      jwtProviders,
	}, nil
}

// buildXdsUpstreamTLSSocket returns an xDS TransportSocket that uses envoyTrustBundle
// as the CA to authenticate server certificates.
// TODO huabing: add support for custom CA and client certificate.
func buildXdsUpstreamTLSSocket(sni string) (*corev3.TransportSocket, error) {
	tlsCtxProto := &tlsv3.UpstreamTlsContext{
		Sni: sni,
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

// patchRoute patches the provided route with a JWT PerRouteConfig, if the route
// doesn't contain it.
func (*jwt) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}

	filterCfg := route.GetTypedPerFilterConfig()
	if _, ok := filterCfg[jwtAuthn]; !ok {
		if !routeContainsJWTAuthn(irRoute) {
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

		route.TypedPerFilterConfig[jwtAuthn] = routeCfgAny
	}

	return nil
}

// patchResources creates JWKS clusters from the provided routes, if needed.
func (*jwt) patchResources(tCtx *types.ResourceVersionTable, routes []*ir.HTTPRoute) error {
	if tCtx == nil || tCtx.XdsResources == nil {
		return errors.New("xds resource table is nil")
	}

	var errs error
	for _, route := range routes {
		if !routeContainsJWTAuthn(route) {
			continue
		}

		for i := range route.JWT.Providers {
			var (
				jwks    *urlCluster
				ds      *ir.DestinationSetting
				tSocket *corev3.TransportSocket
				err     error
			)

			provider := route.JWT.Providers[i]
			jwks, err = url2Cluster(provider.RemoteJWKS.URI)
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}

			ds = &ir.DestinationSetting{
				Weight:    ptr.To[uint32](1),
				Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(jwks.hostname, jwks.port)},
			}

			clusterArgs := &xdsClusterArgs{
				name:         jwks.name,
				settings:     []*ir.DestinationSetting{ds},
				endpointType: jwks.endpointType,
			}
			if jwks.tls {
				tSocket, err = buildXdsUpstreamTLSSocket(jwks.hostname)
				if err != nil {
					errs = errors.Join(errs, err)
					continue
				}
				clusterArgs.tSocket = tSocket
			}

			if err = addXdsCluster(tCtx, clusterArgs); err != nil && !errors.Is(err, ErrXdsClusterExists) {
				errs = errors.Join(errs, err)
			}
		}
	}

	return errs
}

// listenerContainsJWTAuthn returns true if JWT authentication exists for the
// provided listener.
func listenerContainsJWTAuthn(irListener *ir.HTTPListener) bool {
	if irListener == nil {
		return false
	}

	for _, route := range irListener.Routes {
		if routeContainsJWTAuthn(route) {
			return true
		}
	}

	return false
}

// routeContainsJWTAuthn returns true if JWT authentication exists for the
// provided route.
func routeContainsJWTAuthn(irRoute *ir.HTTPRoute) bool {
	if irRoute != nil &&
		irRoute.JWT != nil &&
		irRoute.JWT.Providers != nil &&
		len(irRoute.JWT.Providers) > 0 {
		return true
	}
	return false
}

// buildJwtFromHeaders returns a list of JwtHeader transformed from JWTFromHeader struct
func buildJwtFromHeaders(headers []v1alpha1.JWTHeaderExtractor) []*jwtauthnv3.JwtHeader {
	jwtHeaders := make([]*jwtauthnv3.JwtHeader, 0, len(headers))

	for _, header := range headers {
		jwtHeader := &jwtauthnv3.JwtHeader{
			Name:        header.Name,
			ValuePrefix: ptr.Deref(header.ValuePrefix, ""),
		}

		jwtHeaders = append(jwtHeaders, jwtHeader)
	}

	return jwtHeaders
}
