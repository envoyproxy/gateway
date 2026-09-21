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

// ecdsEligibleFilters lists the HCM filter types served over ECDS rather than written
// into the listener: those whose config changes on its own and is costly to rebuild.
var ecdsEligibleFilters = []egv1a1.EnvoyFilter{
	egv1a1.EnvoyFilterLua,
}

// extractFiltersToECDS moves every ECDS-eligible HCM filter config out of the listener
// into its own xDS resource, leaving a config_discovery reference behind. Envoy then
// applies a config change in place instead of draining the listener. Walking the finished
// listeners visits each filter chain once, including chains shared by several Gateways.
func (t *Translator) extractFiltersToECDS(tCtx *types.ResourceVersionTable) error {
	var errs error
	for _, res := range tCtx.XdsResources[resourcev3.ListenerType] {
		xdsListener, ok := res.(*listenerv3.Listener)
		if !ok {
			continue
		}
		for _, filterChain := range allFilterChains(xdsListener) {
			if err := t.extractFilterChainToECDS(tCtx, filterChain); err != nil {
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

func (t *Translator) extractFilterChainToECDS(tCtx *types.ResourceVersionTable, filterChain *listenerv3.FilterChain) error {
	hcm, err := findHCMinFilterChain(filterChain)
	if errors.Is(err, errHCMNotFound) {
		// A filter chain without an HCM, a TCP proxy for example, has no HTTP filters.
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
		if !t.ecdsFilterNames.Has(httpFilter.Name) || httpFilter.GetTypedConfig() == nil {
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

// recordECDSFilterNames remembers the ECDS-eligible filters of a manager Envoy Gateway
// has just built, so the extraction pass can tell them from filters an EnvoyPatchPolicy or
// an extension server adds later. Matching on the name alone would not do: a filter called
// "envoy.filters.http.lua/my-filter/0" is indistinguishable from one of ours.
func (t *Translator) recordECDSFilterNames(mgr *hcmv3.HttpConnectionManager) {
	for _, httpFilter := range mgr.HttpFilters {
		for _, filterType := range ecdsEligibleFilters {
			if isFilterType(httpFilter, filterType) {
				t.ecdsFilterNames.Insert(httpFilter.Name)
				break
			}
		}
	}
}

// addECDSResource registers the filter's config as an ECDS resource and points the filter
// at it. The filter name doubles as the resource name and is scoped to a
// RouteConfiguration, so the resource carries the scripts of every IR listener whose
// routes land in that route config, and is referenced by every HCM serving it.
//
// TODO: with HTTP/3 that is two managers, the TCP one and the QUIC one, holding identical
// copies of the filter. An EnvoyPatchPolicy or an extension server would not normally
// treat them differently, but if one edited a single copy, last write wins here, and ECDS
// cannot express the difference anyway since both resolve the same name.
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

	// An Any with no bytes decodes to a default-constructed message of that type, so the
	// filter is a no-op, rather than a 503, until its ECDS resource arrives.
	defaultConfig := &anypb.Any{TypeUrl: typeURL}

	httpFilter.ConfigType = &hcmv3.HttpFilter_ConfigDiscovery{
		ConfigDiscovery: &corev3.ExtensionConfigSource{
			ConfigSource:                     makeConfigSource(),
			TypeUrls:                         []string{typeURL},
			DefaultConfig:                    defaultConfig,
			ApplyDefaultConfigWithoutWarming: true,
		},
	}
	return nil
}
