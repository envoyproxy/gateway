// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/ir"
)

type FiltersTranslator interface {
	HTTPFiltersTranslator
}

var _ FiltersTranslator = (*Translator)(nil)

type HTTPFiltersTranslator interface {
	processURLRewriteFilter(filter v1beta1.HTTPRouteFilter, filterContext *HTTPFiltersContext)
	processRedirectFilter(filter v1beta1.HTTPRouteFilter, filterContext *HTTPFiltersContext)
	processRequestHeaderModifierFilter(filter v1beta1.HTTPRouteFilter, filterContext *HTTPFiltersContext)
	processResponseHeaderModifierFilter(filter v1beta1.HTTPRouteFilter, filterContext *HTTPFiltersContext)
	processRequestMirrorFilter(filter v1beta1.HTTPRouteFilter, filterContext *HTTPFiltersContext, resources *Resources)
	processExtensionRefHTTPFilter(filter v1beta1.HTTPRouteFilter, filterContext *HTTPFiltersContext, resources *Resources)
	processUnsupportedHTTPFilter(filter v1beta1.HTTPRouteFilter, filterContext *HTTPFiltersContext)
}

// HTTPFiltersContext is the context of http filters processing.
type HTTPFiltersContext struct {
	*HTTPFilterChains

	ParentRef *RouteParentContext
	HTTPRoute *HTTPRouteContext
}

// HTTPFilterChains contains the ir processing results.
type HTTPFilterChains struct {
	DirectResponse   *ir.DirectResponse
	RedirectResponse *ir.Redirect

	URLRewrite *ir.URLRewrite

	AddRequestHeaders    []ir.AddHeader
	RemoveRequestHeaders []string

	AddResponseHeaders    []ir.AddHeader
	RemoveResponseHeaders []string

	Mirrors []*ir.RouteDestination
}

// ProcessHTTPFilters translate gateway api http filters to IRs.
func (t *Translator) ProcessHTTPFilters(parentRef *RouteParentContext,
	httpRoute *HTTPRouteContext,
	filters []v1beta1.HTTPRouteFilter,
	resources *Resources) *HTTPFiltersContext {
	httpFiltersContext := &HTTPFiltersContext{
		ParentRef: parentRef,
		HTTPRoute: httpRoute,

		HTTPFilterChains: &HTTPFilterChains{},
	}

	for i := range filters {
		filter := filters[i]
		// If an invalid filter type has been configured then skip processing any more filters
		if httpFiltersContext.DirectResponse != nil {
			break
		}
		if err := ValidateHTTPRouteFilter(&filter); err != nil {
			t.processInvalidHTTPFilter(filter, httpFiltersContext, err)
			break
		}

		switch filter.Type {
		case v1beta1.HTTPRouteFilterURLRewrite:
			t.processURLRewriteFilter(filter, httpFiltersContext)
		case v1beta1.HTTPRouteFilterRequestRedirect:
			t.processRedirectFilter(filter, httpFiltersContext)
		case v1beta1.HTTPRouteFilterRequestHeaderModifier:
			t.processRequestHeaderModifierFilter(filter, httpFiltersContext)
		case v1beta1.HTTPRouteFilterResponseHeaderModifier:
			t.processResponseHeaderModifierFilter(filter, httpFiltersContext)
		case v1beta1.HTTPRouteFilterRequestMirror:
			t.processRequestMirrorFilter(filter, httpFiltersContext, resources)
		case v1beta1.HTTPRouteFilterExtensionRef:
			t.processExtensionRefHTTPFilter(filter, httpFiltersContext, resources)
		default:
			t.processUnsupportedHTTPFilter(filter, httpFiltersContext)
		}
	}

	return httpFiltersContext
}

func (t *Translator) processURLRewriteFilter(
	filter v1beta1.HTTPRouteFilter,
	filterContext *HTTPFiltersContext) {
	if filterContext.URLRewrite != nil {
		filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
			v1beta1.RouteConditionAccepted,
			metav1.ConditionFalse,
			v1beta1.RouteReasonUnsupportedValue,
			"Cannot configure multiple urlRewrite filters for a single HTTPRouteRule",
		)
		return
	}

	rewrite := filter.URLRewrite
	if rewrite == nil {
		return
	}

	newURLRewrite := &ir.URLRewrite{}

	if rewrite.Hostname != nil {
		if err := t.validateHostname(string(*rewrite.Hostname)); err != nil {
			filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonUnsupportedValue,
				err.Error(),
			)
			return
		}
		redirectHost := string(*rewrite.Hostname)
		newURLRewrite.Hostname = &redirectHost
	}

	if rewrite.Path != nil {
		switch rewrite.Path.Type {
		case v1beta1.FullPathHTTPPathModifier:
			if rewrite.Path.ReplacePrefixMatch != nil {
				errMsg := "ReplacePrefixMatch cannot be set when rewrite path type is \"ReplaceFullPath\""
				filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					errMsg,
				)
				return
			}
			if rewrite.Path.ReplaceFullPath == nil {
				errMsg := "ReplaceFullPath must be set when rewrite path type is \"ReplaceFullPath\""
				filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					errMsg,
				)
				return
			}
			if rewrite.Path.ReplaceFullPath != nil {
				newURLRewrite.Path = &ir.HTTPPathModifier{
					FullReplace: rewrite.Path.ReplaceFullPath,
				}
			}
		case v1beta1.PrefixMatchHTTPPathModifier:
			if rewrite.Path.ReplaceFullPath != nil {
				errMsg := "ReplaceFullPath cannot be set when rewrite path type is \"ReplacePrefixMatch\""
				filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					errMsg,
				)
				return
			}
			if rewrite.Path.ReplacePrefixMatch == nil {
				errMsg := "ReplacePrefixMatch must be set when rewrite path type is \"ReplacePrefixMatch\""
				filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					errMsg,
				)
				return
			}
			if rewrite.Path.ReplacePrefixMatch != nil {
				newURLRewrite.Path = &ir.HTTPPathModifier{
					PrefixMatchReplace: rewrite.Path.ReplacePrefixMatch,
				}
			}
		default:
			errMsg := fmt.Sprintf("Rewrite path type: %s is invalid, only \"ReplaceFullPath\" and \"ReplacePrefixMatch\" are supported", rewrite.Path.Type)
			filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonUnsupportedValue,
				errMsg,
			)
			return
		}
	}

	filterContext.URLRewrite = newURLRewrite
}

func (t *Translator) processRedirectFilter(
	filter v1beta1.HTTPRouteFilter,
	filterContext *HTTPFiltersContext) {
	// Can't have two redirects for the same route
	if filterContext.RedirectResponse != nil {
		filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
			v1beta1.RouteConditionAccepted,
			metav1.ConditionFalse,
			v1beta1.RouteReasonUnsupportedValue,
			"Cannot configure multiple requestRedirect filters for a single HTTPRouteRule",
		)
		return
	}

	redirect := filter.RequestRedirect
	if redirect == nil {
		return
	}

	redir := &ir.Redirect{}
	if redirect.Scheme != nil {
		// Note that gateway API may support additional schemes in the future, but unknown values
		// must result in an UnsupportedValue status
		if *redirect.Scheme == "http" || *redirect.Scheme == "https" {
			redir.Scheme = redirect.Scheme
		} else {
			errMsg := fmt.Sprintf("Scheme: %s is unsupported, only 'https' and 'http' are supported", *redirect.Scheme)
			filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonUnsupportedValue,
				errMsg,
			)
			return
		}
	}

	if redirect.Hostname != nil {
		if err := t.validateHostname(string(*redirect.Hostname)); err != nil {
			filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonUnsupportedValue,
				err.Error(),
			)
		} else {
			redirectHost := string(*redirect.Hostname)
			redir.Hostname = &redirectHost
		}
	}

	if redirect.Path != nil {
		switch redirect.Path.Type {
		case v1beta1.FullPathHTTPPathModifier:
			if redirect.Path.ReplaceFullPath != nil {
				redir.Path = &ir.HTTPPathModifier{
					FullReplace: redirect.Path.ReplaceFullPath,
				}
			}
		case v1beta1.PrefixMatchHTTPPathModifier:
			if redirect.Path.ReplacePrefixMatch != nil {
				redir.Path = &ir.HTTPPathModifier{
					PrefixMatchReplace: redirect.Path.ReplacePrefixMatch,
				}
			}
		default:
			errMsg := fmt.Sprintf("Redirect path type: %s is invalid, only \"ReplaceFullPath\" and \"ReplacePrefixMatch\" are supported", redirect.Path.Type)
			filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonUnsupportedValue,
				errMsg,
			)
			return
		}
	}

	if redirect.StatusCode != nil {
		redirectCode := int32(*redirect.StatusCode)
		// Envoy supports 302, 303, 307, and 308, but gateway API only includes 301 and 302
		if redirectCode == 301 || redirectCode == 302 {
			redir.StatusCode = &redirectCode
		} else {
			errMsg := fmt.Sprintf("Status code %d is invalid, only 302 and 301 are supported", redirectCode)
			filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonUnsupportedValue,
				errMsg,
			)
			return
		}
	}

	if redirect.Port != nil {
		redirectPort := uint32(*redirect.Port)
		redir.Port = &redirectPort
	}

	filterContext.RedirectResponse = redir
}

func (t *Translator) processRequestHeaderModifierFilter(
	filter v1beta1.HTTPRouteFilter,
	filterContext *HTTPFiltersContext) {
	// Make sure the header modifier config actually exists
	headerModifier := filter.RequestHeaderModifier
	if headerModifier == nil {
		return
	}
	emptyFilterConfig := true // keep track of whether the provided config is empty or not

	// Add request headers
	if headersToAdd := headerModifier.Add; headersToAdd != nil {
		if len(headersToAdd) > 0 {
			emptyFilterConfig = false
		}
		for _, addHeader := range headersToAdd {
			emptyFilterConfig = false
			if addHeader.Name == "" {
				filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					"RequestHeaderModifier Filter cannot add a header with an empty name",
				)
				// try to process the rest of the headers and produce a valid config.
				continue
			}
			// Per Gateway API specification on HTTPHeaderName, : and / are invalid characters in header names
			if strings.Contains(string(addHeader.Name), "/") || strings.Contains(string(addHeader.Name), ":") {
				filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					fmt.Sprintf("RequestHeaderModifier Filter cannot set headers with a '/' or ':' character in them. Header: %q", string(addHeader.Name)),
				)
				continue
			}
			// Check if the header is a duplicate
			headerKey := string(addHeader.Name)
			canAddHeader := true
			for _, h := range filterContext.AddRequestHeaders {
				if strings.EqualFold(h.Name, headerKey) {
					canAddHeader = false
					break
				}
			}

			if !canAddHeader {
				continue
			}

			newHeader := ir.AddHeader{
				Name:   headerKey,
				Append: true,
				Value:  addHeader.Value,
			}

			filterContext.AddRequestHeaders = append(filterContext.AddRequestHeaders, newHeader)
		}
	}

	// Set headers
	if headersToSet := headerModifier.Set; headersToSet != nil {
		if len(headersToSet) > 0 {
			emptyFilterConfig = false
		}
		for _, setHeader := range headersToSet {

			if setHeader.Name == "" {
				filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					"RequestHeaderModifier Filter cannot set a header with an empty name",
				)
				continue
			}
			// Per Gateway API specification on HTTPHeaderName, : and / are invalid characters in header names
			if strings.Contains(string(setHeader.Name), "/") || strings.Contains(string(setHeader.Name), ":") {
				filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					fmt.Sprintf("RequestHeaderModifier Filter cannot set headers with a '/' or ':' character in them. Header: '%s'", string(setHeader.Name)),
				)
				continue
			}

			// Check if the header to be set has already been configured
			headerKey := string(setHeader.Name)
			canAddHeader := true
			for _, h := range filterContext.AddRequestHeaders {
				if strings.EqualFold(h.Name, headerKey) {
					canAddHeader = false
					break
				}
			}
			if !canAddHeader {
				continue
			}
			newHeader := ir.AddHeader{
				Name:   string(setHeader.Name),
				Append: false,
				Value:  setHeader.Value,
			}

			filterContext.AddRequestHeaders = append(filterContext.AddRequestHeaders, newHeader)
		}
	}

	// Remove request headers
	// As far as Envoy is concerned, it is ok to configure a header to be added/set and also in the list of
	// headers to remove. It will remove the original header if present and then add/set the header after.
	if headersToRemove := headerModifier.Remove; headersToRemove != nil {
		if len(headersToRemove) > 0 {
			emptyFilterConfig = false
		}
		for _, removedHeader := range headersToRemove {
			if removedHeader == "" {
				filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					"RequestHeaderModifier Filter cannot remove a header with an empty name",
				)
				continue
			}

			canRemHeader := true
			for _, h := range filterContext.RemoveRequestHeaders {
				if strings.EqualFold(h, removedHeader) {
					canRemHeader = false
					break
				}
			}
			if !canRemHeader {
				continue
			}

			filterContext.RemoveRequestHeaders = append(filterContext.RemoveRequestHeaders, removedHeader)
		}
	}

	// Update the status if the filter failed to configure any valid headers to add/remove
	if len(filterContext.AddRequestHeaders) == 0 && len(filterContext.RemoveRequestHeaders) == 0 && !emptyFilterConfig {
		filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
			v1beta1.RouteConditionAccepted,
			metav1.ConditionFalse,
			v1beta1.RouteReasonUnsupportedValue,
			"RequestHeaderModifier Filter did not provide valid configuration to add/set/remove any headers",
		)
	}
}

func (t *Translator) processResponseHeaderModifierFilter(
	filter v1beta1.HTTPRouteFilter,
	filterContext *HTTPFiltersContext) {
	// Make sure the header modifier config actually exists
	headerModifier := filter.ResponseHeaderModifier
	if headerModifier == nil {
		return
	}
	emptyFilterConfig := true // keep track of whether the provided config is empty or not

	// Add response headers
	if headersToAdd := headerModifier.Add; headersToAdd != nil {
		if len(headersToAdd) > 0 {
			emptyFilterConfig = false
		}
		for _, addHeader := range headersToAdd {
			emptyFilterConfig = false
			if addHeader.Name == "" {
				filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					"ResponseHeaderModifier Filter cannot add a header with an empty name",
				)
				// try to process the rest of the headers and produce a valid config.
				continue
			}
			// Per Gateway API specification on HTTPHeaderName, : and / are invalid characters in header names
			if strings.Contains(string(addHeader.Name), "/") || strings.Contains(string(addHeader.Name), ":") {
				filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					fmt.Sprintf("ResponseHeaderModifier Filter cannot set headers with a '/' or ':' character in them. Header: %q", string(addHeader.Name)),
				)
				continue
			}
			// Check if the header is a duplicate
			headerKey := string(addHeader.Name)
			canAddHeader := true
			for _, h := range filterContext.AddResponseHeaders {
				if strings.EqualFold(h.Name, headerKey) {
					canAddHeader = false
					break
				}
			}

			if !canAddHeader {
				continue
			}

			newHeader := ir.AddHeader{
				Name:   headerKey,
				Append: true,
				Value:  addHeader.Value,
			}

			filterContext.AddResponseHeaders = append(filterContext.AddResponseHeaders, newHeader)
		}
	}

	// Set headers
	if headersToSet := headerModifier.Set; headersToSet != nil {
		if len(headersToSet) > 0 {
			emptyFilterConfig = false
		}
		for _, setHeader := range headersToSet {

			if setHeader.Name == "" {
				filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					"ResponseHeaderModifier Filter cannot set a header with an empty name",
				)
				continue
			}
			// Per Gateway API specification on HTTPHeaderName, : and / are invalid characters in header names
			if strings.Contains(string(setHeader.Name), "/") || strings.Contains(string(setHeader.Name), ":") {
				filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					fmt.Sprintf("ResponseHeaderModifier Filter cannot set headers with a '/' or ':' character in them. Header: '%s'", string(setHeader.Name)),
				)
				continue
			}

			// Check if the header to be set has already been configured
			headerKey := string(setHeader.Name)
			canAddHeader := true
			for _, h := range filterContext.AddResponseHeaders {
				if strings.EqualFold(h.Name, headerKey) {
					canAddHeader = false
					break
				}
			}
			if !canAddHeader {
				continue
			}
			newHeader := ir.AddHeader{
				Name:   string(setHeader.Name),
				Append: false,
				Value:  setHeader.Value,
			}

			filterContext.AddResponseHeaders = append(filterContext.AddResponseHeaders, newHeader)
		}
	}

	// Remove response headers
	// As far as Envoy is concerned, it is ok to configure a header to be added/set and also in the list of
	// headers to remove. It will remove the original header if present and then add/set the header after.
	if headersToRemove := headerModifier.Remove; headersToRemove != nil {
		if len(headersToRemove) > 0 {
			emptyFilterConfig = false
		}
		for _, removedHeader := range headersToRemove {
			if removedHeader == "" {
				filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					"ResponseHeaderModifier Filter cannot remove a header with an empty name",
				)
				continue
			}

			canRemHeader := true
			for _, h := range filterContext.RemoveResponseHeaders {
				if strings.EqualFold(h, removedHeader) {
					canRemHeader = false
					break
				}
			}
			if !canRemHeader {
				continue
			}

			filterContext.RemoveResponseHeaders = append(filterContext.RemoveResponseHeaders, removedHeader)

		}
	}

	// Update the status if the filter failed to configure any valid headers to add/remove
	if len(filterContext.AddResponseHeaders) == 0 && len(filterContext.RemoveResponseHeaders) == 0 && !emptyFilterConfig {
		filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
			v1beta1.RouteConditionAccepted,
			metav1.ConditionFalse,
			v1beta1.RouteReasonUnsupportedValue,
			"ResponseHeaderModifier Filter did not provide valid configuration to add/set/remove any headers",
		)
	}
}

func (t *Translator) processExtensionRefHTTPFilter(filter v1beta1.HTTPRouteFilter, filterContext *HTTPFiltersContext, resources *Resources) {
	// If a reference to a custom filter type cannot be resolved, the filter MUST NOT be skipped.
	// Instead, requests that would have been processed by that filter MUST receive an HTTP error
	// response.
	found := false

	for _, f := range resources.AuthenFilters {
		if f.Namespace == filterContext.HTTPRoute.Namespace && f.Name == string(filter.ExtensionRef.Name) {
			found = true
			break
		}
	}

	if !found {
		errMsg := fmt.Sprintf("Reference not found for filter type: %s", filter.Type)
		// Reset the conditions in case the Accepted=True condition exists.
		filterContext.ParentRef.ResetConditions(filterContext.HTTPRoute)
		filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			v1beta1.RouteReasonBackendNotFound,
			errMsg,
		)
		filterContext.DirectResponse = &ir.DirectResponse{
			Body:       &errMsg,
			StatusCode: 500,
		}
	}
}

func (t *Translator) processRequestMirrorFilter(
	filter v1beta1.HTTPRouteFilter,
	filterContext *HTTPFiltersContext,
	resources *Resources) {

	// Make sure the config actually exists
	mirrorFilter := filter.RequestMirror
	if mirrorFilter == nil {
		return
	}

	mirrorBackend := mirrorFilter.BackendRef

	// Wrap the filter's BackendObjectReference into a BackendRef so we can use existing tooling to check it
	weight := int32(1)
	mirrorBackendRef := v1beta1.BackendRef{
		BackendObjectReference: mirrorBackend,
		Weight:                 &weight,
	}

	// This sets the status on the HTTPRoute, should the usage be changed so that the status message reflects that the backendRef is from the filter?
	serviceNamespace := NamespaceDerefOr(mirrorBackend.Namespace, filterContext.HTTPRoute.Namespace)
	if !t.validateBackendRef(&mirrorBackendRef, filterContext.ParentRef, filterContext.HTTPRoute,
		resources, serviceNamespace, KindHTTPRoute) {
		return
	}

	mirrorDest, _ := t.processRouteDestination(mirrorBackendRef, filterContext.ParentRef, filterContext.HTTPRoute, resources)

	// If we're already mirroring requests to this backend then there is no need to configure it twice
	for _, mirror := range filterContext.Mirrors {
		if mirror != nil {
			if mirror.Host == mirrorDest.Host &&
				mirror.Port == mirrorDest.Port {
				return
			}
		}
	}

	filterContext.Mirrors = append(filterContext.Mirrors, mirrorDest)

}

func (t *Translator) processUnsupportedHTTPFilter(filter v1beta1.HTTPRouteFilter, filterContext *HTTPFiltersContext) {
	errMsg := fmt.Sprintf("Unsupported filter type: %s", filter.Type)
	filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
		v1beta1.RouteConditionAccepted,
		metav1.ConditionFalse,
		v1beta1.RouteReasonUnsupportedValue,
		errMsg,
	)
	filterContext.DirectResponse = &ir.DirectResponse{
		Body:       &errMsg,
		StatusCode: 500,
	}
}

func (t *Translator) processInvalidHTTPFilter(filter v1beta1.HTTPRouteFilter, filterContext *HTTPFiltersContext, err error) {
	errMsg := fmt.Sprintf("Invalid filter %s: %v", filter.Type, err)
	filterContext.ParentRef.SetCondition(filterContext.HTTPRoute,
		v1beta1.RouteConditionAccepted,
		metav1.ConditionFalse,
		v1beta1.RouteReasonUnsupportedValue,
		errMsg,
	)
	filterContext.DirectResponse = &ir.DirectResponse{
		Body:       &errMsg,
		StatusCode: 500,
	}
}
