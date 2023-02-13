// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// ProcessCustomGRPCFilters translates gateway api grpc filters to IRs.
func (t *Translator) ProcessCustomGRPCFilters(parentRef *RouteParentContext,
	route RouteContext,
	filters []v1alpha2.GRPCRouteFilter,
	resources *Resources) *HTTPFiltersContext {
	httpFiltersContext := &HTTPFiltersContext{
		ParentRef: parentRef,
		Route:     route,

		HTTPFilterIR: &HTTPFilterIR{},
	}

	for i := range filters {
		filter := filters[i]
		// If an invalid filter type has been configured then skip processing any more filters
		if httpFiltersContext.DirectResponse != nil {
			break
		}
		if err := ValidateCustomGRPCRouteFilter(&filter); err != nil {
			t.processInvalidHTTPFilter(string(filter.Type), httpFiltersContext, err)
			break
		}

		switch filter.Type {
		case v1alpha2.GRPCRouteFilterRequestHeaderModifier:
			t.processRequestHeaderModifierFilter(filter.RequestHeaderModifier, httpFiltersContext)
		case v1alpha2.GRPCRouteFilterResponseHeaderModifier:
			t.processResponseHeaderModifierFilter(filter.ResponseHeaderModifier, httpFiltersContext)
		case v1alpha2.GRPCRouteFilterRequestMirror:
			t.processRequestMirrorFilter(filter.RequestMirror, httpFiltersContext, resources)
		case v1alpha2.GRPCRouteFilterExtensionRef:
			t.processExtensionRefHTTPFilter(filter.ExtensionRef, httpFiltersContext, resources)
		default:
			t.processUnsupportedHTTPFilter(string(filter.Type), httpFiltersContext)
		}
	}

	return httpFiltersContext
}
