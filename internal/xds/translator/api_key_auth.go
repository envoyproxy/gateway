// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	apikeyauthv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/api_key_auth/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&apiKeyAuth{})
}

type apiKeyAuth struct{}

var _ httpFilter = &apiKeyAuth{}

// patchHCM builds and appends the api_key_auth Filter to the HTTP Connection Manager
// if applicable, and it does not already exist.
func (*apiKeyAuth) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}
	if hcmContainsFilter(mgr, egv1a1.EnvoyFilterAPIKeyAuth.String()) {
		return nil
	}

	var (
		irAPIKeyAuth *ir.APIKeyAuth
		filter       *hcmv3.HttpFilter
		err          error
	)

	for _, route := range irListener.Routes {
		if route.Security != nil && route.Security.APIKeyAuth != nil {
			irAPIKeyAuth = route.Security.APIKeyAuth
			break
		}
	}
	if irAPIKeyAuth == nil {
		return nil
	}

	// We use the first route that contains the api key auth config to build the filter.
	// The HCM-level filter config doesn't matter since it is overridden at the route level.
	if filter, err = buildHCMAPIKeyAuthFilter(irAPIKeyAuth); err != nil {
		return err
	}
	mgr.HttpFilters = append(mgr.HttpFilters, filter)
	return err
}

// buildHCMAPIKeyAuthFilter returns a api_key_auth HTTP filter from the provided IR HTTPRoute.
func buildHCMAPIKeyAuthFilter(apiKeyAuth *ir.APIKeyAuth) (*hcmv3.HttpFilter, error) {
	apiKeyAuthProto := buildAPIKeyAuthFilterConfig(apiKeyAuth)
	apiKeyAuthAny, err := proto.ToAnyWithValidation(apiKeyAuthProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: egv1a1.EnvoyFilterAPIKeyAuth.String(),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: apiKeyAuthAny,
		},
		Disabled: true,
	}, nil
}

func (*apiKeyAuth) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

// patchRoute patches the provided route with the apiKeyAuth config if applicable.
// Note: this method overwrites the HCM level filter config with the per route filter config.
func (*apiKeyAuth) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.Security == nil || irRoute.Security.APIKeyAuth == nil {
		return nil
	}

	perFilterCfg := route.GetTypedPerFilterConfig()
	if _, ok := perFilterCfg[egv1a1.EnvoyFilterAPIKeyAuth.String()]; ok {
		// This should not happen since this is the only place where the filter
		// config is added in a route.
		return fmt.Errorf("route already contains filter config: %s, %+v",
			egv1a1.EnvoyFilterAPIKeyAuth.String(), route)
	}

	// Overwrite the HCM level filter config with the per route filter config.
	apiKeyAuthProto := buildAPIKeyAuthFilterPerRouteConfig(irRoute.Security.APIKeyAuth)
	apiKeyAuthAny, err := proto.ToAnyWithValidation(apiKeyAuthProto)
	if err != nil {
		return err
	}

	if perFilterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}
	route.TypedPerFilterConfig[egv1a1.EnvoyFilterAPIKeyAuth.String()] = apiKeyAuthAny

	return nil
}

func buildAPIKeyAuthFilterConfig(apiKeyAuth *ir.APIKeyAuth) *apikeyauthv3.ApiKeyAuth {
	apiKeyAuthProto := &apikeyauthv3.ApiKeyAuth{
		Credentials: make([]*apikeyauthv3.Credential, 0, len(apiKeyAuth.Credentials)),
	}
	for clientid, key := range apiKeyAuth.Credentials {
		apiKeyAuthProto.Credentials = append(apiKeyAuthProto.Credentials, &apikeyauthv3.Credential{
			Client: clientid,
			Key:    string(key),
		})
	}

	for _, e := range apiKeyAuth.ExtractFrom {
		for _, header := range e.Headers {
			source := &apikeyauthv3.KeySource{
				Header: header,
			}
			apiKeyAuthProto.KeySources = append(apiKeyAuthProto.KeySources, source)
		}
		for _, param := range e.Params {
			source := &apikeyauthv3.KeySource{
				Query: param,
			}
			apiKeyAuthProto.KeySources = append(apiKeyAuthProto.KeySources, source)
		}
		for _, cookie := range e.Cookies {
			source := &apikeyauthv3.KeySource{
				Cookie: cookie,
			}
			apiKeyAuthProto.KeySources = append(apiKeyAuthProto.KeySources, source)
		}
	}

	clientIDHeader := ptr.Deref(apiKeyAuth.ForwardClientIDHeader, "")
	sanitize := ptr.Deref(apiKeyAuth.Sanitize, false)
	if clientIDHeader != "" || sanitize {
		apiKeyAuthProto.Forwarding = &apikeyauthv3.Forwarding{
			Header:          clientIDHeader,
			HideCredentials: sanitize,
		}
	}

	return apiKeyAuthProto
}

func buildAPIKeyAuthFilterPerRouteConfig(apiKeyAuth *ir.APIKeyAuth) *apikeyauthv3.ApiKeyAuthPerRoute {
	apiKeyAuthProto := buildAPIKeyAuthFilterConfig(apiKeyAuth)
	return &apikeyauthv3.ApiKeyAuthPerRoute{
		Credentials: apiKeyAuthProto.Credentials,
		KeySources:  apiKeyAuthProto.KeySources,
		Forwarding:  apiKeyAuthProto.Forwarding,
	}
}
