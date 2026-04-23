// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package extensionserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	bav3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/basic_auth/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/exampleorg/envoygateway-extension/api/v1alpha1"
	"google.golang.org/protobuf/types/known/anypb"

	pb "github.com/envoyproxy/gateway/proto/extension"
)

// PostHTTPListenerModify is called after Envoy Gateway is done generating a
// Listener xDS configuration and before that configuration is passed on to
// Envoy Proxy.
// This example adds Basic Authentication on the Listener level as an example.
// Note: This implementation is not secure, and should not be used to protect
// anything important.
func (s *Server) PostHTTPListenerModify(ctx context.Context, req *pb.PostHTTPListenerModifyRequest) (*pb.PostHTTPListenerModifyResponse, error) {
	s.log.Info("postHTTPListenerModify callback was invoked")
	// Collect all of the required username/password combinations from the
	// provided contexts that were attached to the gateway.
	passwords := NewHtpasswd()
	for _, ext := range req.PostListenerContext.ExtensionResources {
		var listenerContext v1alpha1.ListenerContextExample
		if err := json.Unmarshal(ext.GetUnstructuredBytes(), &listenerContext); err != nil {
			s.log.Error("failed to unmarshal the extension", slog.String("error", err.Error()))
			continue
		}
		s.log.Info("processing an extension context", slog.String("username", listenerContext.Spec.Username))
		passwords.AddUser(listenerContext.Spec.Username, listenerContext.Spec.Password)
	}

	// First, get the filter chains from the listener
	filterChains := req.Listener.GetFilterChains()
	defaultFC := req.Listener.DefaultFilterChain
	if defaultFC != nil {
		filterChains = append(filterChains, defaultFC)
	}
	// Go over all of the chains, and add the basic authentication http filter
	for _, currChain := range filterChains {
		httpConManager, hcmIndex, err := findHCM(currChain)
		if err != nil {
			s.log.Error("failed to find an HCM in the current chain", slog.Any("error", err))
			continue
		}
		// If a basic authentication filter already exists, update it. Otherwise, create it.
		basicAuth, baIndex, err := findBasicAuthFilter(httpConManager.HttpFilters)
		if err != nil {
			s.log.Error("failed to unmarshal the existing basicAuth filter", slog.Any("error", err))
			continue
		}
		if baIndex == -1 {
			// Create a new basic auth filter
			basicAuth = &bav3.BasicAuth{
				Users: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineString{
						InlineString: passwords.String(),
					},
				},
				ForwardUsernameHeader: "X-Example-Ext",
			}
		} else {
			// Update the basic auth filter
			basicAuth.Users.Specifier = &corev3.DataSource_InlineString{
				InlineString: passwords.String(),
			}
		}
		// Add or update the Basic Authentication filter in the HCM
		anyBAFilter, _ := anypb.New(basicAuth)
		if baIndex > -1 {
			httpConManager.HttpFilters[baIndex].ConfigType = &hcm.HttpFilter_TypedConfig{
				TypedConfig: anyBAFilter,
			}
		} else {
			filters := []*hcm.HttpFilter{
				{
					Name: "envoy.filters.http.basic_auth",
					ConfigType: &hcm.HttpFilter_TypedConfig{
						TypedConfig: anyBAFilter,
					},
				},
			}
			filters = append(filters, httpConManager.HttpFilters...)
			httpConManager.HttpFilters = filters
		}

		// Write the updated HCM back to the filter chain
		anyConnectionMgr, _ := anypb.New(httpConManager)
		currChain.Filters[hcmIndex].ConfigType = &listenerv3.Filter_TypedConfig{
			TypedConfig: anyConnectionMgr,
		}
	}

	return &pb.PostHTTPListenerModifyResponse{
		Listener: req.Listener,
	}, nil
}

// Tries to find an HTTP connection manager in the provided filter chain.
func findHCM(filterChain *listenerv3.FilterChain) (*hcm.HttpConnectionManager, int, error) {
	for filterIndex, filter := range filterChain.Filters {
		if filter.Name == wellknown.HTTPConnectionManager {
			hcm := new(hcm.HttpConnectionManager)
			if err := filter.GetTypedConfig().UnmarshalTo(hcm); err != nil {
				return nil, -1, err
			}
			return hcm, filterIndex, nil
		}
	}
	return nil, -1, fmt.Errorf("unable to find HTTPConnectionManager in FilterChain: %s", filterChain.Name)
}

// Tries to find the Basic Authentication HTTP filter in the provided chain
func findBasicAuthFilter(chain []*hcm.HttpFilter) (*bav3.BasicAuth, int, error) {
	for i, filter := range chain {
		if filter.Name == "envoy.filters.http.basic_auth" {
			ba := new(bav3.BasicAuth)
			if err := filter.GetTypedConfig().UnmarshalTo(ba); err != nil {
				return nil, -1, err
			}
			return ba, i, nil
		}
	}
	return nil, -1, nil
}
