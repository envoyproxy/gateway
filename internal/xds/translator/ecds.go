// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	resourceTypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

// ecdsEligibleFilters lists the HCM filter types whose configuration is served over ECDS
// instead of being written into the listener.
//
// A filter belongs here when its configuration changes independently of the rest of the
// listener and rebuilding it is expensive. Lua qualifies on both counts: a script edit is
// a routine operation, and each distinct script costs one Lua VM per worker thread.
var ecdsEligibleFilters = []egv1a1.EnvoyFilter{
	egv1a1.EnvoyFilterLua,
}

// extractFiltersToECDS moves the configuration of every ECDS-eligible HCM filter out of
// the listener into its own xDS resource, leaving a config_discovery reference behind.
// Changing such a configuration is then an ECDS update, which Envoy applies in place,
// rather than a listener update, which drains the listener.
//
// It runs over the finished listeners so that it sees every filter chain exactly once,
// including chains shared by several Gateway listeners.
func extractFiltersToECDS(tCtx *types.ResourceVersionTable) error {
	var errs error
	for _, res := range tCtx.XdsResources[resourcev3.ListenerType] {
		xdsListener, ok := res.(*listenerv3.Listener)
		if !ok {
			continue
		}
		for _, filterChain := range allFilterChains(xdsListener) {
			if err := extractFilterChainToECDS(tCtx, filterChain); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}
	return errs
}

func allFilterChains(xdsListener *listenerv3.Listener) []*listenerv3.FilterChain {
	filterChains := make([]*listenerv3.FilterChain, 0, len(xdsListener.FilterChains)+1)
	filterChains = append(filterChains, xdsListener.FilterChains...)
	if xdsListener.DefaultFilterChain != nil {
		filterChains = append(filterChains, xdsListener.DefaultFilterChain)
	}
	return filterChains
}

func extractFilterChainToECDS(tCtx *types.ResourceVersionTable, filterChain *listenerv3.FilterChain) error {
	// Filter chains without an HCM, a TCP proxy for example, have no HTTP filters.
	hcm, err := findHCMinFilterChain(filterChain)
	if hcm == nil {
		return nil
	}
	if err != nil {
		return err
	}

	var (
		errs    error
		patched bool
	)
	for _, httpFilter := range hcm.HttpFilters {
		if !ecdsEligible(httpFilter) || httpFilter.GetTypedConfig() == nil {
			continue
		}
		if err := addECDSResource(tCtx, httpFilter); err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		patched = true
	}
	if !patched {
		return errs
	}

	return errors.Join(errs, replaceHCMInFilterChain(hcm, filterChain))
}

func ecdsEligible(httpFilter *hcmv3.HttpFilter) bool {
	for _, filterType := range ecdsEligibleFilters {
		if isFilterType(httpFilter, filterType) {
			return true
		}
	}
	return false
}

// addECDSResource registers the filter's current configuration as an ECDS resource and
// rewrites the filter to point at it. The filter name doubles as the resource name, which
// is why filter names have to be unique across the proxy and not just within one HCM.
func addECDSResource(tCtx *types.ResourceVersionTable, httpFilter *hcmv3.HttpFilter) error {
	typeURL := httpFilter.GetTypedConfig().GetTypeUrl()
	extensionConfig := &corev3.TypedExtensionConfig{
		Name:        httpFilter.Name,
		TypedConfig: httpFilter.GetTypedConfig(),
	}
	if err := tCtx.AddOrReplaceXdsResource(resourcev3.ExtensionConfigType, extensionConfig,
		func(existing, new resourceTypes.Resource) bool {
			return existing.(*corev3.TypedExtensionConfig).Name == new.(*corev3.TypedExtensionConfig).Name
		}); err != nil {
		return err
	}

	// An Any carrying no bytes decodes to a default-constructed message of that type, which
	// gives us a do-nothing config of the right type without knowing which type it is. The
	// filter therefore behaves as a no-op, rather than failing requests with a 503, in the
	// window before its ECDS resource arrives.
	defaultConfig := &anypb.Any{TypeUrl: typeURL}

	httpFilter.ConfigType = &hcmv3.HttpFilter_ConfigDiscovery{
		ConfigDiscovery: &corev3.ExtensionConfigSource{
			ConfigSource:                     makeConfigSource(),
			TypeUrls:                         []string{httpFilter.GetTypedConfig().GetTypeUrl()},
			DefaultConfig:                    defaultConfig,
			ApplyDefaultConfigWithoutWarming: true,
		},
	}
	return nil
}
