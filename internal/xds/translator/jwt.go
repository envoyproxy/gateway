// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	jwtauthnv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/jwt_authn/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	envoyTrustBundle = "/etc/ssl/certs/ca-certificates.crt"
)

func init() {
	registerHTTPFilter(&jwt{})
}

type jwt struct{}

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
		if httpFilter.Name == egv1a1.EnvoyFilterJWTAuthn.String() {
			return nil
		}
	}

	jwtFilter, err := buildHCMJWTFilter(irListener)
	if err != nil {
		return err
	}

	mgr.HttpFilters = append([]*hcmv3.HttpFilter{jwtFilter}, mgr.HttpFilters...)

	return nil
}

// buildHCMJWTFilter returns a JWT authn HTTP filter from the provided IR listener.
func buildHCMJWTFilter(irListener *ir.HTTPListener) (*hcmv3.HttpFilter, error) {
	jwtAuthnProto, err := buildJWTAuthn(irListener)
	if err != nil {
		return nil, err
	}

	jwtAuthnAny, err := proto.ToAnyWithValidation(jwtAuthnProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: egv1a1.EnvoyFilterJWTAuthn.String(),
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
		for i := range route.Security.JWT.Providers {
			var (
				irProvider = route.Security.JWT.Providers[i]
				err        error
			)

			claimToHeaders := []*jwtauthnv3.JwtClaimToHeader{}
			for _, claimToHeader := range irProvider.ClaimToHeaders {
				claimToHeader := &jwtauthnv3.JwtClaimToHeader{
					HeaderName: claimToHeader.Header,
					ClaimName:  claimToHeader.Claim,
				}
				claimToHeaders = append(claimToHeaders, claimToHeader)
			}
			jwtProvider := &jwtauthnv3.JwtProvider{
				Issuer:            irProvider.Issuer,
				Audiences:         irProvider.Audiences,
				PayloadInMetadata: irProvider.Name,
				ClaimToHeaders:    claimToHeaders,
				Forward:           true,
				NormalizePayloadInMetadata: &jwtauthnv3.JwtProvider_NormalizePayload{
					// Normalize the scopes to facilitate matching in Authorization.
					SpaceDelimitedClaims: []string{"scope"},
				},
			}
			if irProvider.LocalJWKS != nil {
				local := &jwtauthnv3.JwtProvider_LocalJwks{
					LocalJwks: &corev3.DataSource{
						Specifier: &corev3.DataSource_InlineString{
							InlineString: *irProvider.LocalJWKS,
						},
					},
				}
				jwtProvider.JwksSourceSpecifier = local
			} else {
				var jwksCluster string

				jwks := irProvider.RemoteJWKS
				if jwks.Destination != nil && len(jwks.Destination.Settings) > 0 {
					jwksCluster = jwks.Destination.Name
				} else {
					var cluster *urlCluster
					if cluster, err = url2Cluster(jwks.URI); err != nil {
						return nil, err
					}
					jwksCluster = cluster.name
				}

				var duration *metav1.Duration
				if jwks.CacheDuration != nil {
					duration = jwks.CacheDuration
				}

				var asyncFetch jwtauthnv3.JwksAsyncFetch

				if jwks.AsyncFetch != nil {
					asyncFetch = jwtauthnv3.JwksAsyncFetch{
						FastListener:          jwks.AsyncFetch.FastListener,
						FailedRefetchDuration: durationpb.New(jwks.AsyncFetch.FailedRefetchDuration.Duration),
					}
				}

				remote := &jwtauthnv3.JwtProvider_RemoteJwks{
					RemoteJwks: &jwtauthnv3.RemoteJwks{
						HttpUri: &corev3.HttpUri{
							Uri: jwks.URI,
							HttpUpstreamType: &corev3.HttpUri_Cluster{
								Cluster: jwksCluster,
							},
							Timeout: durationpb.New(defaultExtServiceRequestTimeout),
						},
						CacheDuration: &durationpb.Duration{Seconds: 5 * 60},
						AsyncFetch:    &jwtauthnv3.JwksAsyncFetch{},
					},
				}

				// Set the retry policy if it exists.
				if jwks.Traffic != nil && jwks.Traffic.Retry != nil {
					var rp *corev3.RetryPolicy
					if rp, err = buildNonRouteRetryPolicy(jwks.Traffic.Retry); err != nil {
						return nil, err
					}
					remote.RemoteJwks.RetryPolicy = rp
				}
				jwtProvider.JwksSourceSpecifier = remote
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

		if route.Security.JWT.AllowMissing {
			reqs = append(reqs, &jwtauthnv3.JwtRequirement{
				RequiresType: &jwtauthnv3.JwtRequirement_AllowMissing{
					AllowMissing: &emptypb.Empty{},
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

	tlsCtxAny, err := proto.ToAnyWithValidation(tlsCtxProto)
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
func (*jwt) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}

	filterCfg := route.GetTypedPerFilterConfig()
	if _, ok := filterCfg[egv1a1.EnvoyFilterJWTAuthn.String()]; !ok {
		if !routeContainsJWTAuthn(irRoute) {
			return nil
		}

		routeCfgProto := &jwtauthnv3.PerRouteConfig{
			RequirementSpecifier: &jwtauthnv3.PerRouteConfig_RequirementName{RequirementName: irRoute.Name},
		}

		routeCfgAny, err := proto.ToAnyWithValidation(routeCfgProto)
		if err != nil {
			return err
		}

		if filterCfg == nil {
			route.TypedPerFilterConfig = make(map[string]*anypb.Any)
		}

		route.TypedPerFilterConfig[egv1a1.EnvoyFilterJWTAuthn.String()] = routeCfgAny
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

		for i := range route.Security.JWT.Providers {
			jwks := route.Security.JWT.Providers[i].RemoteJWKS
			if jwks == nil {
				continue
			}

			// If the rmote JWKS has a destination, use it.
			if jwks.Destination != nil && len(jwks.Destination.Settings) > 0 {
				if err := createExtServiceXDSCluster(
					jwks.Destination, jwks.Traffic, tCtx); err != nil {
					errs = errors.Join(errs, err)
				}
			} else {
				// Create a cluster with the token endpoint url.
				if err := addClusterFromURL(jwks.URI, tCtx); err != nil {
					errs = errors.Join(errs, err)
				}
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
		irRoute.Security != nil &&
		irRoute.Security.JWT != nil &&
		irRoute.Security.JWT.Providers != nil &&
		len(irRoute.Security.JWT.Providers) > 0 {
		return true
	}
	return false
}

// buildJwtFromHeaders returns a list of JwtHeader transformed from JWTFromHeader struct
func buildJwtFromHeaders(headers []egv1a1.JWTHeaderExtractor) []*jwtauthnv3.JwtHeader {
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
