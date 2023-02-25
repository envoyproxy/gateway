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

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	jwtext "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/jwt_authn/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
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
func patchHCMWithJwtAuthnFilter(mgr *hcm.HttpConnectionManager, irListener *ir.HTTPListener) error {
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
	mgr.HttpFilters = append([]*hcm.HttpFilter{jwtFilter}, mgr.HttpFilters...)

	return nil
}

// buildHCMJwtFilter returns a JWT authn HTTP filter from the provided IR listener.
func buildHCMJwtFilter(irListener *ir.HTTPListener) (*hcm.HttpFilter, error) {
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

	return &hcm.HttpFilter{
		Name: jwtAuthenFilter,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: jwtAuthnAny,
		},
	}, nil
}

// buildJwtAuthn returns a JwtAuthentication based on the provided IR HTTPListener.
func buildJwtAuthn(irListener *ir.HTTPListener) (*jwtext.JwtAuthentication, error) {
	jwtProviders := make(map[string]*jwtext.JwtProvider)
	reqMap := make(map[string]*jwtext.JwtRequirement)

	for _, route := range irListener.Routes {
		if route != nil && routeContainsJwtAuthn(route) {
			var reqs []*jwtext.JwtRequirement
			for i := range route.RequestAuthentication.JWT.Providers {
				irProvider := route.RequestAuthentication.JWT.Providers[i]
				// Create the cluster for the remote jwks, if it doesn't exist.
				jwksCluster, err := newJwksCluster(&irProvider)
				if err != nil {
					return nil, err
				}

				remote := &jwtext.JwtProvider_RemoteJwks{
					RemoteJwks: &jwtext.RemoteJwks{
						HttpUri: &core.HttpUri{
							Uri: irProvider.RemoteJWKS.URI,
							HttpUpstreamType: &core.HttpUri_Cluster{
								Cluster: jwksCluster.name,
							},
							Timeout: &durationpb.Duration{Seconds: 5},
						},
						CacheDuration: &durationpb.Duration{Seconds: 5 * 60},
					},
				}

				jwtProvider := &jwtext.JwtProvider{
					Issuer:              irProvider.Issuer,
					Audiences:           irProvider.Audiences,
					JwksSourceSpecifier: remote,
					PayloadInMetadata:   irProvider.Issuer,
				}

				providerKey := fmt.Sprintf("%s-%s", route.Name, irProvider.Name)
				jwtProviders[providerKey] = jwtProvider
				reqs = append(reqs, &jwtext.JwtRequirement{
					RequiresType: &jwtext.JwtRequirement_ProviderName{
						ProviderName: providerKey,
					},
				})
			}
			if len(reqs) == 1 {
				reqMap[route.Name] = reqs[0]
			} else {
				orListReqs := &jwtext.JwtRequirement{
					RequiresType: &jwtext.JwtRequirement_RequiresAny{
						RequiresAny: &jwtext.JwtRequirementOrList{
							Requirements: reqs,
						},
					},
				}
				reqMap[route.Name] = orListReqs
			}
		}
	}

	return &jwtext.JwtAuthentication{
		RequirementMap: reqMap,
		Providers:      jwtProviders,
	}, nil
}

// buildXdsUpstreamTLSSocket returns an xDS TransportSocket that uses envoyTrustBundle
// as the CA to authenticate server certificates.
func buildXdsUpstreamTLSSocket() (*core.TransportSocket, error) {
	tlsCtxProto := &tls.UpstreamTlsContext{
		CommonTlsContext: &tls.CommonTlsContext{
			ValidationContextType: &tls.CommonTlsContext_ValidationContext{
				ValidationContext: &tls.CertificateValidationContext{
					TrustedCa: &core.DataSource{
						Specifier: &core.DataSource_Filename{
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

	return &core.TransportSocket{
		Name: wellknown.TransportSocketTls,
		ConfigType: &core.TransportSocket_TypedConfig{
			TypedConfig: tlsCtxAny,
		},
	}, nil
}

// patchRouteWithJwtConfig patches the provided route with a JWT PerRouteConfig, if the
// route doesn't contain it. The listener is used to create the PerRouteConfig JWT
// requirement.
func patchRouteWithJwtConfig(route *routev3.Route, irRoute *ir.HTTPRoute, listener *listener.Listener) error { //nolint:unparam
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if listener == nil {
		return errors.New("listener is nil")
	}

	filterCfg := route.GetTypedPerFilterConfig()
	if _, ok := filterCfg[jwtAuthenFilter]; !ok {
		if !routeContainsJwtAuthn(irRoute) {
			return nil
		}

		routeCfgProto, err := buildJwtPerRouteConfig(irRoute, listener)
		if err != nil {
			return fmt.Errorf("failed to build per route config for ir route %s", irRoute.Name)
		}

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

// buildJwtPerRouteConfig returns a JWT PerRouteConfig based on the provided IR route and HCM.
func buildJwtPerRouteConfig(irRoute *ir.HTTPRoute, listener *listener.Listener) (*jwtext.PerRouteConfig, error) {
	if irRoute == nil {
		return nil, errors.New("ir route is nil")
	}
	if irRoute == nil {
		return nil, errors.New("ir route does not contain jwt authn")
	}
	if listener == nil {
		return nil, errors.New("listener is nil")
	}

	filterCh := listener.GetDefaultFilterChain()
	if filterCh == nil {
		return nil, fmt.Errorf("listener %s does not contain the default filterchain", listener.Name)
	}

	for _, filter := range filterCh.Filters {
		if filter.Name == wellknown.HTTPConnectionManager {
			// Unmarshal the filter to a jwt authn config and validate it.
			hcmProto := new(hcm.HttpConnectionManager)
			hcmAny := filter.GetTypedConfig()
			if err := hcmAny.UnmarshalTo(hcmProto); err != nil {
				return nil, err
			}
			if err := hcmProto.ValidateAll(); err != nil {
				return nil, err
			}
			//
			req, err := getJwtRequirement(hcmProto.GetHttpFilters(), irRoute.Name)
			if err != nil {
				return nil, err
			}

			return &jwtext.PerRouteConfig{
				RequirementSpecifier: req,
			}, nil
		}
	}

	return nil, errors.New("failed to find HTTP connection manager filter")
}

// getJwtRequirement iterates through the provided filters, returning a JWT requirement
// name if one exists.
func getJwtRequirement(filters []*hcm.HttpFilter, name string) (*jwtext.PerRouteConfig_RequirementName, error) {
	if len(filters) == 0 {
		return nil, errors.New("no hcm http filters")
	}

	for _, filter := range filters {
		if filter.Name == jwtAuthenFilter {
			// Unmarshal the filter to a jwt authn config and validate it.
			jwtAuthnProto := new(jwtext.JwtAuthentication)
			jwtAuthnAny := filter.GetTypedConfig()
			if err := jwtAuthnAny.UnmarshalTo(jwtAuthnProto); err != nil {
				return nil, err
			}
			if err := jwtAuthnProto.ValidateAll(); err != nil {
				return nil, err
			}
			// Return the requirement name if it's found.
			if _, found := jwtAuthnProto.RequirementMap[name]; found {
				return &jwtext.PerRouteConfig_RequirementName{RequirementName: name}, nil
			}
			return nil, fmt.Errorf("failed to find jwt requirement %s", name)
		}
	}

	return nil, errors.New("failed to find jwt authn filter")
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
				if err != nil {
					return err
				}
				if existingCluster := findXdsCluster(tCtx, jwks.name); existingCluster == nil {
					routeDestinations := []*ir.RouteDestination{ir.NewRouteDest(jwks.hostname, jwks.port, 0)}
					jwksServerCluster := buildXdsCluster(jwks.name, routeDestinations, false /*isHTTP2 */, jwks.isStatic)
					tSocket, err := buildXdsUpstreamTLSSocket()
					if err != nil {
						return err
					}
					jwksServerCluster.TransportSocket = tSocket

					tCtx.AddXdsResource(resource.ClusterType, jwksServerCluster)
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
