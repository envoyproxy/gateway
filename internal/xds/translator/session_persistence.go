// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	statefulsessionv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/stateful_session/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	cookiev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/stateful_session/cookie/v3"
	headerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/stateful_session/header/v3"
	httpv3 "github.com/envoyproxy/go-control-plane/envoy/type/http/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	cookieConfigName = "envoy.http.stateful_session.cookie"
	headerConfigName = "envoy.http.stateful_session.header"
)

type sessionPersistence struct{}

func init() {
	registerHTTPFilter(&sessionPersistence{})
}

var _ httpFilter = &sessionPersistence{}

// patchHCM patches the HttpConnectionManager with the filter.
// Note: this method may be called multiple times for the same filter, please
// make sure to avoid duplicate additions of the same filter.
func (s *sessionPersistence) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, route := range irListener.Routes {
		sp := route.SessionPersistence
		if sp == nil {
			continue
		}

		if hcmContainsFilter(mgr, perRouteFilterName(egv1a1.EnvoyFilterSessionPersistence, route.Name)) {
			continue
		}

		var sessionCfg proto.Message
		var configName string
		switch {
		case sp.Cookie != nil:
			configName = cookieConfigName
			sessionCfg = &cookiev3.CookieBasedSessionState{
				Cookie: &httpv3.Cookie{
					Name: sp.Cookie.Name,
					Path: routePathToCookiePath(route.PathMatch),
					Ttl:  durationpb.New(sp.Cookie.TTL.Duration),
				},
			}
		case sp.Header != nil:
			configName = headerConfigName
			sessionCfg = &headerv3.HeaderBasedSessionState{
				Name: sp.Header.Name,
			}
		}

		sessionCfgAny, err := anypb.New(sessionCfg)
		if err != nil {
			return fmt.Errorf("failed to marshal %s config: %w", egv1a1.EnvoyFilterSessionPersistence.String(), err)
		}

		cfg := &statefulsessionv3.StatefulSession{
			SessionState: &corev3.TypedExtensionConfig{
				Name:        configName,
				TypedConfig: sessionCfgAny,
			},
		}

		cfgAny, err := anypb.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal %s config: %w", egv1a1.EnvoyFilterSessionPersistence.String(), err)
		}

		mgr.HttpFilters = append(mgr.HttpFilters, &hcmv3.HttpFilter{
			Name:     perRouteFilterName(egv1a1.EnvoyFilterSessionPersistence, route.Name),
			Disabled: true,
			ConfigType: &hcmv3.HttpFilter_TypedConfig{
				TypedConfig: cfgAny,
			},
		})
	}

	return nil
}

func routePathToCookiePath(path *ir.StringMatch) string {
	if path == nil {
		return "/"
	}
	switch {
	case path.Exact != nil:
		return *path.Exact
	case path.Prefix != nil:
		return *path.Prefix
	case path.SafeRegex != nil:
		return getLongestNonRegexPrefix(*path.SafeRegex)
	}

	// Shouldn't reach here because the path should be either of the above three kinds.
	return "/"
}

// getLongestNonRegexPrefix takes a regex path and returns the longest non-regex prefix.
// > 3. For an xRoute using a path that is a regex, the Path should be set to the longest non-regex prefix
// (.e.g. if the path is /p1/p2/*/p3 and the request path was /p1/p2/foo/p3, then the cookie path would be /p1/p2).
// https://gateway-api.sigs.k8s.io/geps/gep-1619/#path
func getLongestNonRegexPrefix(path string) string {
	parts := strings.Split(path, "/")
	var longestNonRegexPrefix []string
	for _, part := range parts {
		if part == "*" || strings.Contains(part, "*") {
			break
		}
		longestNonRegexPrefix = append(longestNonRegexPrefix, part)
	}

	return strings.Join(longestNonRegexPrefix, "/")
}

// patchRoute patches the provide Route with a filter's Route level configuration.
func (s *sessionPersistence) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.SessionPersistence == nil {
		return nil
	}

	if err := enableFilterOnRoute(route, perRouteFilterName(egv1a1.EnvoyFilterSessionPersistence, route.Name)); err != nil {
		return err
	}

	return nil
}

// patchResources adds all the other needed resources referenced by this
// filter to the resource version table.
func (s *sessionPersistence) patchResources(tCtx *types.ResourceVersionTable, routes []*ir.HTTPRoute) error {
	return nil
}
