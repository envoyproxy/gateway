// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
)

type FiltersTranslator interface {
	HTTPFiltersTranslator
}

var _ FiltersTranslator = (*Translator)(nil)

type HTTPFiltersTranslator interface {
	processURLRewriteFilter(rewrite *gwapiv1.HTTPURLRewriteFilter, filterContext *HTTPFiltersContext)
	processRedirectFilter(redirect *gwapiv1.HTTPRequestRedirectFilter, filterContext *HTTPFiltersContext)
	processRequestHeaderModifierFilter(headerModifier *gwapiv1.HTTPHeaderFilter, filterContext *HTTPFiltersContext)
	processResponseHeaderModifierFilter(headerModifier *gwapiv1.HTTPHeaderFilter, filterContext *HTTPFiltersContext)
	processRequestMirrorFilter(filterIdx int, mirror *gwapiv1.HTTPRequestMirrorFilter, filterContext *HTTPFiltersContext, resources *resource.Resources) error
	processUnsupportedHTTPFilter(filterType string, filterContext *HTTPFiltersContext)
}

// HTTPFiltersContext is the context of http filters processing.
type HTTPFiltersContext struct {
	*HTTPFilterIR

	ParentRef *RouteParentContext
	Route     RouteContext
	RuleIdx   int
}

// HTTPFilterIR contains the ir processing results.
type HTTPFilterIR struct {
	DirectResponse   *ir.CustomResponse
	RedirectResponse *ir.Redirect

	URLRewrite *ir.URLRewrite

	AddRequestHeaders    []ir.AddHeader
	RemoveRequestHeaders []string

	AddResponseHeaders    []ir.AddHeader
	RemoveResponseHeaders []string

	Mirrors []*ir.MirrorPolicy

	ExtensionRefs []*ir.UnstructuredRef
}

// ProcessHTTPFilters translates gateway api http filters to IRs.
func (t *Translator) ProcessHTTPFilters(parentRef *RouteParentContext,
	route RouteContext,
	filters []gwapiv1.HTTPRouteFilter,
	ruleIdx int,
	resources *resource.Resources,
) (*HTTPFiltersContext, error) {
	httpFiltersContext := &HTTPFiltersContext{
		ParentRef:    parentRef,
		Route:        route,
		RuleIdx:      ruleIdx,
		HTTPFilterIR: &HTTPFilterIR{},
	}
	var err error
	for i := range filters {
		filter := filters[i]
		// If an invalid filter type has been configured then skip processing any more filters
		if httpFiltersContext.DirectResponse != nil {
			break
		}
		if err := ValidateHTTPRouteFilter(&filter, t.ExtensionGroupKinds...); err != nil {
			t.processInvalidHTTPFilter(string(filter.Type), httpFiltersContext, err)
			break
		}

		switch filter.Type {
		case gwapiv1.HTTPRouteFilterURLRewrite:
			t.processURLRewriteFilter(filter.URLRewrite, httpFiltersContext)
		case gwapiv1.HTTPRouteFilterRequestRedirect:
			t.processRedirectFilter(filter.RequestRedirect, httpFiltersContext)
		case gwapiv1.HTTPRouteFilterRequestHeaderModifier:
			t.processRequestHeaderModifierFilter(filter.RequestHeaderModifier, httpFiltersContext)
		case gwapiv1.HTTPRouteFilterResponseHeaderModifier:
			t.processResponseHeaderModifierFilter(filter.ResponseHeaderModifier, httpFiltersContext)
		case gwapiv1.HTTPRouteFilterRequestMirror:
			err = t.processRequestMirrorFilter(i, filter.RequestMirror, httpFiltersContext, resources)
		case gwapiv1.HTTPRouteFilterExtensionRef:
			t.processExtensionRefHTTPFilter(filter.ExtensionRef, httpFiltersContext, resources)
		default:
			t.processUnsupportedHTTPFilter(string(filter.Type), httpFiltersContext)
		}
	}

	return httpFiltersContext, err
}

// ProcessGRPCFilters translates gateway api grpc filters to IRs.
func (t *Translator) ProcessGRPCFilters(parentRef *RouteParentContext,
	route RouteContext,
	filters []gwapiv1.GRPCRouteFilter,
	resources *resource.Resources,
) (*HTTPFiltersContext, error) {
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
		if err := ValidateGRPCRouteFilter(&filter); err != nil {
			t.processInvalidHTTPFilter(string(filter.Type), httpFiltersContext, err)
			break
		}

		switch filter.Type {
		case gwapiv1.GRPCRouteFilterRequestHeaderModifier:
			t.processRequestHeaderModifierFilter(filter.RequestHeaderModifier, httpFiltersContext)
		case gwapiv1.GRPCRouteFilterResponseHeaderModifier:
			t.processResponseHeaderModifierFilter(filter.ResponseHeaderModifier, httpFiltersContext)
		case gwapiv1.GRPCRouteFilterRequestMirror:
			err := t.processRequestMirrorFilter(i, filter.RequestMirror, httpFiltersContext, resources)
			if err != nil {
				return nil, err
			}
		case gwapiv1.GRPCRouteFilterExtensionRef:
			t.processExtensionRefHTTPFilter(filter.ExtensionRef, httpFiltersContext, resources)
		default:
			t.processUnsupportedHTTPFilter(string(filter.Type), httpFiltersContext)
		}
	}

	return httpFiltersContext, nil
}

// Checks if the context and the rewrite both contain a core gw-api HTTP URL rewrite
func hasMultipleCoreRewrites(rewrite *gwapiv1.HTTPURLRewriteFilter, contextRewrite *ir.URLRewrite) bool {
	contextHasCoreRewrites := contextRewrite.Path != nil && (contextRewrite.Path.FullReplace != nil ||
		contextRewrite.Path.PrefixMatchReplace != nil) || (contextRewrite.Host != nil && contextRewrite.Host.Name != nil)
	rewriteHasCoreRewrites := rewrite.Hostname != nil || rewrite.Path != nil
	return contextHasCoreRewrites && rewriteHasCoreRewrites
}

// Checks if the context and the rewrite both contain a envoy-gateway extended HTTP URL rewrite
func hasMultipleExtensionRewrites(rewrite *egv1a1.HTTPURLRewriteFilter, contextRewrite *ir.URLRewrite) bool {
	contextHasExtensionRewrites := (contextRewrite.Path != nil && contextRewrite.Path.RegexMatchReplace != nil) ||
		(contextRewrite.Host != nil && (contextRewrite.Host.Header != nil || contextRewrite.Host.Backend != nil))

	return contextHasExtensionRewrites && (rewrite.Hostname != nil || rewrite.Path != nil)
}

// Checks if the context and the gw-api core rewrite both contain an HTTP URL rewrite that creates a conflict (e.g. both rewrite path)
func hasConflictingCoreAndExtensionRewrites(rewrite *gwapiv1.HTTPURLRewriteFilter, contextRewrite *ir.URLRewrite) bool {
	contextHasExtensionPathRewrites := contextRewrite.Path != nil && contextRewrite.Path.RegexMatchReplace != nil
	contextHasExtensionHostRewrites := contextRewrite.Host != nil && (contextRewrite.Host.Header != nil ||
		contextRewrite.Host.Backend != nil)
	return (rewrite.Hostname != nil && contextHasExtensionHostRewrites) || (rewrite.Path != nil && contextHasExtensionPathRewrites)
}

// Checks if the context and the envoy-gateway extended rewrite both contain an HTTP URL rewrite that creates a conflict (e.g. both rewrite path)
func hasConflictingExtensionAndCoreRewrites(rewrite *egv1a1.HTTPURLRewriteFilter, contextRewrite *ir.URLRewrite) bool {
	contextHasCorePathRewrites := contextRewrite.Path != nil && (contextRewrite.Path.FullReplace != nil ||
		contextRewrite.Path.PrefixMatchReplace != nil)
	contextHasCoreHostnameRewrites := contextRewrite.Host != nil && contextRewrite.Host.Name != nil

	return (rewrite.Hostname != nil && contextHasCoreHostnameRewrites) || (rewrite.Path != nil && contextHasCorePathRewrites)
}

func (t *Translator) processURLRewriteFilter(
	rewrite *gwapiv1.HTTPURLRewriteFilter,
	filterContext *HTTPFiltersContext,
) {
	if filterContext.URLRewrite != nil {
		if hasMultipleCoreRewrites(rewrite, filterContext.URLRewrite) ||
			hasConflictingCoreAndExtensionRewrites(rewrite, filterContext.URLRewrite) {
			routeStatus := GetRouteStatus(filterContext.Route)
			status.SetRouteStatusCondition(routeStatus,
				filterContext.ParentRef.routeParentStatusIdx,
				filterContext.Route.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonUnsupportedValue,
				"Cannot configure multiple urlRewrite filters for a single HTTPRouteRule",
			)
			return
		}
	}

	if rewrite == nil {
		return
	}

	newURLRewrite := &ir.URLRewrite{}

	if rewrite.Hostname != nil {
		if err := t.validateHostname(string(*rewrite.Hostname)); err != nil {
			routeStatus := GetRouteStatus(filterContext.Route)
			status.SetRouteStatusCondition(routeStatus,
				filterContext.ParentRef.routeParentStatusIdx,
				filterContext.Route.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonUnsupportedValue,
				err.Error(),
			)
			return
		}
		redirectHost := string(*rewrite.Hostname)
		newURLRewrite.Host = &ir.HTTPHostModifier{
			Name: &redirectHost,
		}
	}

	if rewrite.Path != nil {
		switch rewrite.Path.Type {
		case gwapiv1.FullPathHTTPPathModifier:
			if rewrite.Path.ReplacePrefixMatch != nil {
				errMsg := "ReplacePrefixMatch cannot be set when rewrite path type is \"ReplaceFullPath\""
				routeStatus := GetRouteStatus(filterContext.Route)
				status.SetRouteStatusCondition(routeStatus,
					filterContext.ParentRef.routeParentStatusIdx,
					filterContext.Route.GetGeneration(),
					gwapiv1.RouteConditionAccepted,
					metav1.ConditionFalse,
					gwapiv1.RouteReasonUnsupportedValue,
					errMsg,
				)
				return
			}
			if rewrite.Path.ReplaceFullPath == nil {
				errMsg := "ReplaceFullPath must be set when rewrite path type is \"ReplaceFullPath\""
				routeStatus := GetRouteStatus(filterContext.Route)
				status.SetRouteStatusCondition(routeStatus,
					filterContext.ParentRef.routeParentStatusIdx,
					filterContext.Route.GetGeneration(),
					gwapiv1.RouteConditionAccepted,
					metav1.ConditionFalse,
					gwapiv1.RouteReasonUnsupportedValue,
					errMsg,
				)
				return
			}
			if rewrite.Path.ReplaceFullPath != nil {
				newURLRewrite.Path = &ir.ExtendedHTTPPathModifier{
					HTTPPathModifier: ir.HTTPPathModifier{
						FullReplace: rewrite.Path.ReplaceFullPath,
					},
				}
			}
		case gwapiv1.PrefixMatchHTTPPathModifier:
			if rewrite.Path.ReplaceFullPath != nil {
				errMsg := "ReplaceFullPath cannot be set when rewrite path type is \"ReplacePrefixMatch\""
				routeStatus := GetRouteStatus(filterContext.Route)
				status.SetRouteStatusCondition(routeStatus,
					filterContext.ParentRef.routeParentStatusIdx,
					filterContext.Route.GetGeneration(),
					gwapiv1.RouteConditionAccepted,
					metav1.ConditionFalse,
					gwapiv1.RouteReasonUnsupportedValue,
					errMsg,
				)
				return
			}
			if rewrite.Path.ReplacePrefixMatch == nil {
				errMsg := "ReplacePrefixMatch must be set when rewrite path type is \"ReplacePrefixMatch\""
				routeStatus := GetRouteStatus(filterContext.Route)
				status.SetRouteStatusCondition(routeStatus,
					filterContext.ParentRef.routeParentStatusIdx,
					filterContext.Route.GetGeneration(),
					gwapiv1.RouteConditionAccepted,
					metav1.ConditionFalse,
					gwapiv1.RouteReasonUnsupportedValue,
					errMsg,
				)
				return
			}
			if rewrite.Path.ReplacePrefixMatch != nil {
				newURLRewrite.Path = &ir.ExtendedHTTPPathModifier{
					HTTPPathModifier: ir.HTTPPathModifier{
						PrefixMatchReplace: rewrite.Path.ReplacePrefixMatch,
					},
				}
			}
		default:
			errMsg := fmt.Sprintf("Rewrite path type: %s is invalid, only \"ReplaceFullPath\" and \"ReplacePrefixMatch\" are supported", rewrite.Path.Type)
			routeStatus := GetRouteStatus(filterContext.Route)
			status.SetRouteStatusCondition(routeStatus,
				filterContext.ParentRef.routeParentStatusIdx,
				filterContext.Route.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonUnsupportedValue,
				errMsg,
			)
			return
		}
	}

	filterContext.URLRewrite = newURLRewrite
}

func (t *Translator) processRedirectFilter(
	redirect *gwapiv1.HTTPRequestRedirectFilter,
	filterContext *HTTPFiltersContext,
) {
	// Can't have two redirects for the same route
	if filterContext.RedirectResponse != nil {
		routeStatus := GetRouteStatus(filterContext.Route)
		status.SetRouteStatusCondition(routeStatus,
			filterContext.ParentRef.routeParentStatusIdx,
			filterContext.Route.GetGeneration(),
			gwapiv1.RouteConditionAccepted,
			metav1.ConditionFalse,
			gwapiv1.RouteReasonUnsupportedValue,
			"Cannot configure multiple requestRedirect filters for a single HTTPRouteRule",
		)
		return
	}

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
			routeStatus := GetRouteStatus(filterContext.Route)
			status.SetRouteStatusCondition(routeStatus,
				filterContext.ParentRef.routeParentStatusIdx,
				filterContext.Route.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonUnsupportedValue,
				errMsg,
			)
			return
		}
	}

	if redirect.Hostname != nil {
		if err := t.validateHostname(string(*redirect.Hostname)); err != nil {
			routeStatus := GetRouteStatus(filterContext.Route)
			status.SetRouteStatusCondition(routeStatus,
				filterContext.ParentRef.routeParentStatusIdx,
				filterContext.Route.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonUnsupportedValue,
				err.Error(),
			)
		} else {
			redirectHost := string(*redirect.Hostname)
			redir.Hostname = &redirectHost
		}
	}

	if redirect.Path != nil {
		switch redirect.Path.Type {
		case gwapiv1.FullPathHTTPPathModifier:
			if redirect.Path.ReplaceFullPath != nil {
				redir.Path = &ir.HTTPPathModifier{
					FullReplace: redirect.Path.ReplaceFullPath,
				}
			}
		case gwapiv1.PrefixMatchHTTPPathModifier:
			if redirect.Path.ReplacePrefixMatch != nil {
				redir.Path = &ir.HTTPPathModifier{
					PrefixMatchReplace: redirect.Path.ReplacePrefixMatch,
				}
			}
		default:
			errMsg := fmt.Sprintf("Redirect path type: %s is invalid, only \"ReplaceFullPath\" and \"ReplacePrefixMatch\" are supported", redirect.Path.Type)
			routeStatus := GetRouteStatus(filterContext.Route)
			status.SetRouteStatusCondition(routeStatus,
				filterContext.ParentRef.routeParentStatusIdx,
				filterContext.Route.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonUnsupportedValue,
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
			routeStatus := GetRouteStatus(filterContext.Route)
			status.SetRouteStatusCondition(routeStatus,
				filterContext.ParentRef.routeParentStatusIdx,
				filterContext.Route.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonUnsupportedValue,
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
	headerModifier *gwapiv1.HTTPHeaderFilter,
	filterContext *HTTPFiltersContext,
) {
	// Make sure the header modifier config actually exists
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
				updateRouteStatusForRHMFilter(
					filterContext,
					"RequestHeaderModifier Filter cannot add a header with an empty name",
				)
				// try to process the rest of the headers and produce a valid config.
				continue
			}
			if !isModifiableHeader(string(addHeader.Name)) {
				updateRouteStatusForRHMFilter(
					filterContext,
					fmt.Sprintf(
						"RequestHeaderModifier Filter cannot set headers with a '/' or ':' character in them, or the host header. Header: %q",
						string(addHeader.Name)),
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
				Value:  strings.Split(addHeader.Value, ","),
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
				updateRouteStatusForRHMFilter(
					filterContext,
					"RequestHeaderModifier Filter cannot set a header with an empty name",
				)
				continue
			}

			if !isModifiableHeader(string(setHeader.Name)) {
				updateRouteStatusForRHMFilter(
					filterContext,
					fmt.Sprintf(
						"RequestHeaderModifier Filter cannot set headers with a '/' or ':' character in them, or the host header. Header: %q",
						string(setHeader.Name)),
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
				Value:  strings.Split(setHeader.Value, ","),
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
				updateRouteStatusForRHMFilter(
					filterContext,
					"RequestHeaderModifier Filter cannot remove a header with an empty name",
				)
				continue
			}

			if !isModifiableHeader(removedHeader) {
				updateRouteStatusForRHMFilter(
					filterContext,
					fmt.Sprintf(
						"RequestHeaderModifier Filter cannot remove headers with a '/' or ':' character in them, or the host header. Header: %q",
						removedHeader),
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
		updateRouteStatusForRHMFilter(
			filterContext,
			"RequestHeaderModifier Filter did not provide valid configuration to add/set/remove any headers",
		)
	}
}

func updateRouteStatusForRHMFilter(filterContext *HTTPFiltersContext, message string) {
	routeStatus := GetRouteStatus(filterContext.Route)
	status.SetRouteStatusCondition(routeStatus,
		filterContext.ParentRef.routeParentStatusIdx,
		filterContext.Route.GetGeneration(),
		gwapiv1.RouteConditionAccepted,
		metav1.ConditionFalse,
		gwapiv1.RouteReasonUnsupportedValue,
		message,
	)
}

func isModifiableHeader(headerName string) bool {
	// Per Gateway API specification on HTTPHeaderName, : and / are invalid characters in header names
	// And Envoy does not allow modification the pseudo headers and the host header
	return !strings.Contains(headerName, "/") && !strings.Contains(headerName, ":") && !strings.EqualFold(headerName, "host")
}

func (t *Translator) processResponseHeaderModifierFilter(
	headerModifier *gwapiv1.HTTPHeaderFilter,
	filterContext *HTTPFiltersContext,
) {
	// Make sure the header modifier config actually exists
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
				updateRouteStatusForRHMFilter(
					filterContext,
					"ResponseHeaderModifier Filter cannot add a header with an empty name",
				)
				// try to process the rest of the headers and produce a valid config.
				continue
			}
			if !isModifiableHeader(string(addHeader.Name)) {
				updateRouteStatusForRHMFilter(
					filterContext,
					fmt.Sprintf("ResponseHeaderModifier Filter cannot set headers with a '/' or ':' character in them, or the host header. Header: %q",
						string(addHeader.Name)),
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
				Value:  strings.Split(addHeader.Value, ","),
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
				updateRouteStatusForRHMFilter(
					filterContext,
					"ResponseHeaderModifier Filter cannot set a header with an empty name",
				)
				continue
			}

			if !isModifiableHeader(string(setHeader.Name)) {
				updateRouteStatusForRHMFilter(
					filterContext,
					fmt.Sprintf("ResponseHeaderModifier Filter cannot set headers with a '/' or ':' character in them, or the host header. Header: %q",
						string(setHeader.Name)),
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
				Value:  strings.Split(setHeader.Value, ","),
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
				updateRouteStatusForRHMFilter(
					filterContext,
					"ResponseHeaderModifier Filter cannot remove a header with an empty name",
				)
				continue
			}
			if !isModifiableHeader(removedHeader) {
				updateRouteStatusForRHMFilter(
					filterContext,
					fmt.Sprintf(
						"ResponseHeaderModifier Filter cannot remove headers with a '/' or ':' character in them, or the host header. Header: %q",
						removedHeader),
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
		updateRouteStatusForRHMFilter(
			filterContext,
			"ResponseHeaderModifier Filter did not provide valid configuration to add/set/remove any headers",
		)
	}
}

func (t *Translator) processExtensionRefHTTPFilter(extFilter *gwapiv1.LocalObjectReference, filterContext *HTTPFiltersContext, resources *resource.Resources) {
	// Make sure the config actually exists.
	if extFilter == nil {
		return
	}

	filterNs := filterContext.Route.GetNamespace()

	if string(extFilter.Kind) == egv1a1.KindHTTPRouteFilter {
		found := false
		for _, hrf := range resources.HTTPRouteFilters {
			if hrf.Namespace == filterNs && hrf.Name == string(extFilter.Name) {
				found = true
				if hrf.Spec.URLRewrite != nil {

					if filterContext.URLRewrite != nil {
						if hasMultipleExtensionRewrites(hrf.Spec.URLRewrite, filterContext.URLRewrite) ||
							hasConflictingExtensionAndCoreRewrites(hrf.Spec.URLRewrite, filterContext.URLRewrite) {
							routeStatus := GetRouteStatus(filterContext.Route)
							status.SetRouteStatusCondition(routeStatus,
								filterContext.ParentRef.routeParentStatusIdx,
								filterContext.Route.GetGeneration(),
								gwapiv1.RouteConditionAccepted,
								metav1.ConditionFalse,
								gwapiv1.RouteReasonUnsupportedValue,
								"Cannot configure multiple urlRewrite filters for a single HTTPRouteRule",
							)
							return
						}
					}

					if hrf.Spec.URLRewrite.Path != nil {
						if hrf.Spec.URLRewrite.Path.Type == egv1a1.RegexHTTPPathModifier {
							if hrf.Spec.URLRewrite.Path.ReplaceRegexMatch == nil ||
								hrf.Spec.URLRewrite.Path.ReplaceRegexMatch.Pattern == "" {
								errMsg := "ReplaceRegexMatch Pattern must be set when rewrite path type is \"ReplaceRegexMatch\""
								routeStatus := GetRouteStatus(filterContext.Route)
								status.SetRouteStatusCondition(routeStatus,
									filterContext.ParentRef.routeParentStatusIdx,
									filterContext.Route.GetGeneration(),
									gwapiv1.RouteConditionAccepted,
									metav1.ConditionFalse,
									gwapiv1.RouteReasonUnsupportedValue,
									errMsg,
								)
								return
							} else if _, err := regexp.Compile(hrf.Spec.URLRewrite.Path.ReplaceRegexMatch.Pattern); err != nil {
								// Avoid envoy NACKs due to invalid regex.
								// Golang's regexp is almost identical to RE2: https://pkg.go.dev/regexp/syntax
								errMsg := "ReplaceRegexMatch must be a valid RE2 regular expression"
								routeStatus := GetRouteStatus(filterContext.Route)
								status.SetRouteStatusCondition(routeStatus,
									filterContext.ParentRef.routeParentStatusIdx,
									filterContext.Route.GetGeneration(),
									gwapiv1.RouteConditionAccepted,
									metav1.ConditionFalse,
									gwapiv1.RouteReasonUnsupportedValue,
									errMsg,
								)
								return
							}

							rmr := &ir.RegexMatchReplace{
								Pattern:      hrf.Spec.URLRewrite.Path.ReplaceRegexMatch.Pattern,
								Substitution: hrf.Spec.URLRewrite.Path.ReplaceRegexMatch.Substitution,
							}

							if filterContext.HTTPFilterIR.URLRewrite != nil {
								if filterContext.HTTPFilterIR.URLRewrite.Path == nil {
									filterContext.HTTPFilterIR.URLRewrite.Path = &ir.ExtendedHTTPPathModifier{
										RegexMatchReplace: rmr,
									}
								}
							} else { // no url rewrite
								filterContext.HTTPFilterIR.URLRewrite = &ir.URLRewrite{
									Path: &ir.ExtendedHTTPPathModifier{
										RegexMatchReplace: rmr,
									},
								}
							}
						}
					}

					if hrf.Spec.URLRewrite.Hostname != nil {
						var hm *ir.HTTPHostModifier
						if hrf.Spec.URLRewrite.Hostname.Type == egv1a1.HeaderHTTPHostnameModifier {
							if hrf.Spec.URLRewrite.Hostname.Header == nil {
								errMsg := "Header must be set when rewrite path type is \"Header\""
								routeStatus := GetRouteStatus(filterContext.Route)
								status.SetRouteStatusCondition(routeStatus,
									filterContext.ParentRef.routeParentStatusIdx,
									filterContext.Route.GetGeneration(),
									gwapiv1.RouteConditionAccepted,
									metav1.ConditionFalse,
									gwapiv1.RouteReasonUnsupportedValue,
									errMsg,
								)
								return
							}
							hm = &ir.HTTPHostModifier{
								Header: hrf.Spec.URLRewrite.Hostname.Header,
							}
						} else if hrf.Spec.URLRewrite.Hostname.Type == egv1a1.BackendHTTPHostnameModifier {
							hm = &ir.HTTPHostModifier{
								Backend: ptr.To(true),
							}
						}

						if filterContext.HTTPFilterIR.URLRewrite != nil {
							if filterContext.HTTPFilterIR.URLRewrite.Host == nil {
								filterContext.HTTPFilterIR.URLRewrite.Host = hm
							}
						} else { // no url rewrite
							filterContext.HTTPFilterIR.URLRewrite = &ir.URLRewrite{
								Host: hm,
							}
						}
					}

				}

				if hrf.Spec.DirectResponse != nil {
					dr := &ir.CustomResponse{}
					if hrf.Spec.DirectResponse.Body != nil {
						var err error
						if dr.Body, err = getCustomResponseBody(hrf.Spec.DirectResponse.Body, resources, filterNs); err != nil {
							t.processInvalidHTTPFilter(string(extFilter.Kind), filterContext, err)
							return
						}
					}

					if hrf.Spec.DirectResponse.StatusCode != nil {
						dr.StatusCode = ptr.To(uint32(*hrf.Spec.DirectResponse.StatusCode))
					} else {
						dr.StatusCode = ptr.To(uint32(200))
					}

					if hrf.Spec.DirectResponse.ContentType != nil {
						newHeader := ir.AddHeader{
							Name:  "Content-Type",
							Value: []string{*hrf.Spec.DirectResponse.ContentType},
						}
						filterContext.AddResponseHeaders = append(filterContext.AddResponseHeaders, newHeader)
					}

					filterContext.HTTPFilterIR.DirectResponse = dr
				}
			}
		}
		if !found {
			errMsg := fmt.Sprintf("Unable to translate HTTPRouteFilter: %s/%s", filterNs,
				extFilter.Name)
			t.processUnresolvedHTTPFilter(errMsg, filterContext)
		}
		return
	}

	// This list of resources will be empty unless an extension is loaded (and introduces resources)
	for _, res := range resources.ExtensionRefFilters {
		if res.GetKind() == string(extFilter.Kind) && res.GetName() == string(extFilter.Name) && res.GetNamespace() == filterNs {
			apiVers := res.GetAPIVersion()

			// To get only the group we cut off the version.
			// This could be a one liner but just to be safe we check that the APIVersion is properly formatted
			idx := strings.IndexByte(apiVers, '/')
			if idx == -1 {
				errMsg := fmt.Sprintf("Unable to translate APIVersion for Extension Filter: kind: %s, %s/%s", res.GetKind(), filterNs, extFilter.Name)
				t.processUnresolvedHTTPFilter(errMsg, filterContext)
				return
			}
			group := apiVers[:idx]
			if group == string(extFilter.Group) {
				res := res // Capture loop variable
				filterContext.ExtensionRefs = append(filterContext.ExtensionRefs, &ir.UnstructuredRef{
					Object: &res,
				})
				return
			}
		}
	}

	// Matching filter not found, so set negative status condition.
	errMsg := fmt.Sprintf("Reference %s/%s not found for filter type: %v", filterNs,
		extFilter.Name, extFilter.Kind)
	t.processUnresolvedHTTPFilter(errMsg, filterContext)
}

func (t *Translator) processRequestMirrorFilter(
	filterIdx int,
	mirrorFilter *gwapiv1.HTTPRequestMirrorFilter,
	filterContext *HTTPFiltersContext,
	resources *resource.Resources,
) error {
	// Make sure the config actually exists
	if mirrorFilter == nil {
		return nil
	}

	mirrorBackend := mirrorFilter.BackendRef

	// Wrap the filter's BackendObjectReference into a BackendRef so we can use existing tooling to check it
	weight := int32(1)
	mirrorBackendRef := gwapiv1.HTTPBackendRef{
		BackendRef: gwapiv1.BackendRef{
			BackendObjectReference: mirrorBackend,
			Weight:                 &weight,
		},
	}

	// This sets the status on the HTTPRoute, should the usage be changed so that the status message reflects that the backendRef is from the filter?
	filterNs := filterContext.Route.GetNamespace()
	serviceNamespace := NamespaceDerefOr(mirrorBackend.Namespace, filterNs)
	err := t.validateBackendRef(mirrorBackendRef, filterContext.ParentRef, filterContext.Route,
		resources, serviceNamespace, resource.KindHTTPRoute)
	if err != nil {
		return err
	}

	destName := fmt.Sprintf("%s-mirror-%d", irRouteDestinationName(filterContext.Route, filterContext.RuleIdx), filterIdx)
	settingName := irDestinationSettingName(destName, -1 /*unused*/)
	ds, err := t.processDestination(settingName, mirrorBackendRef, filterContext.ParentRef, filterContext.Route, resources)
	if err != nil {
		return err
	}

	routeDst := &ir.RouteDestination{
		Name:     destName,
		Settings: []*ir.DestinationSetting{ds},
	}

	var percent *float32
	if f := mirrorFilter.Fraction; f != nil {
		percent = ptr.To(100 * float32(f.Numerator) / float32(ptr.Deref(f.Denominator, int32(100))))
	} else if p := mirrorFilter.Percent; p != nil {
		percent = ptr.To(float32(*p))
	}

	filterContext.Mirrors = append(filterContext.Mirrors, &ir.MirrorPolicy{Destination: routeDst, Percentage: percent})
	return nil
}

func (t *Translator) processUnresolvedHTTPFilter(errMsg string, filterContext *HTTPFiltersContext) {
	routeStatus := GetRouteStatus(filterContext.Route)
	status.SetRouteStatusCondition(routeStatus,
		filterContext.ParentRef.routeParentStatusIdx,
		filterContext.Route.GetGeneration(),
		gwapiv1.RouteConditionResolvedRefs,
		metav1.ConditionFalse,
		gwapiv1.RouteReasonBackendNotFound,
		errMsg,
	)
	status.SetRouteStatusCondition(routeStatus,
		filterContext.ParentRef.routeParentStatusIdx,
		filterContext.Route.GetGeneration(),
		gwapiv1.RouteConditionAccepted,
		metav1.ConditionFalse,
		gwapiv1.RouteReasonUnsupportedValue,
		errMsg,
	)
	filterContext.DirectResponse = &ir.CustomResponse{
		StatusCode: ptr.To(uint32(500)),
	}
}

func (t *Translator) processUnsupportedHTTPFilter(filterType string, filterContext *HTTPFiltersContext) {
	errMsg := fmt.Sprintf("Unsupported filter type: %s", filterType)
	routeStatus := GetRouteStatus(filterContext.Route)
	status.SetRouteStatusCondition(routeStatus,
		filterContext.ParentRef.routeParentStatusIdx,
		filterContext.Route.GetGeneration(),
		gwapiv1.RouteConditionAccepted,
		metav1.ConditionFalse,
		gwapiv1.RouteReasonUnsupportedValue,
		errMsg,
	)
	filterContext.DirectResponse = &ir.CustomResponse{
		StatusCode: ptr.To(uint32(500)),
	}
}

func (t *Translator) processInvalidHTTPFilter(filterType string, filterContext *HTTPFiltersContext, err error) {
	errMsg := fmt.Sprintf("Invalid filter %s: %v", filterType, err)
	routeStatus := GetRouteStatus(filterContext.Route)
	status.SetRouteStatusCondition(routeStatus,
		filterContext.ParentRef.routeParentStatusIdx,
		filterContext.Route.GetGeneration(),
		gwapiv1.RouteConditionAccepted,
		metav1.ConditionFalse,
		gwapiv1.RouteReasonUnsupportedValue,
		errMsg,
	)
	filterContext.DirectResponse = &ir.CustomResponse{
		StatusCode: ptr.To(uint32(500)),
	}
}
