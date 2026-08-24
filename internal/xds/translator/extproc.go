// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"slices"

	cncfv3 "github.com/cncf/xds/go/xds/core/v3"
	matcherv3 "github.com/cncf/xds/go/xds/type/matcher/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matchingv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/matching/v3"
	actionv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/matcher/action/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_proc/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoymatcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&extProc{})
}

type extProc struct{}

var _ httpFilter = &extProc{}

// patchHCM builds and appends the ext_proc Filters to the HTTP Connection Manager
// if applicable, and it does not already exist.
// Note: this method creates an ext_proc filter for each route that contains an ExtAuthz config.
// The filter is disabled by default. It is enabled on the route level.
func (*extProc) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	var errs error

	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, route := range irListener.Routes {
		if !routeContainsExtProc(route) {
			continue
		}

		for i := range route.EnvoyExtensions.ExtProcs {
			ep := &route.EnvoyExtensions.ExtProcs[i]
			if hcmContainsFilter(mgr, extProcFilterName(ep)) {
				continue
			}

			filter, err := buildHCMExtProcFilter(ep)
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}

			mgr.HttpFilters = append(mgr.HttpFilters, filter)
		}
	}

	return errs
}

// buildHCMExtProcFilter returns an ext_proc HTTP filter from the provided IR HTTPRoute.
func buildHCMExtProcFilter(extProc *ir.ExtProc) (*hcmv3.HttpFilter, error) {
	extProcProto, err := extProcConfig(extProc)
	if err != nil {
		return nil, err
	}
	extProcAny, err := anypb.New(extProcProto)
	if err != nil {
		return nil, err
	}

	// For matches-based ext_proc, keep the wrapper at HCM so route-level
	// ExtensionWithMatcherPerRoute matcher overrides are honored, but keep
	// matcher content only at route scope to avoid duplication.
	if len(extProc.Matches) > 0 {
		extProcAny, err = buildExtProcWrapper(extProc, extProcAny, nil)
		if err != nil {
			return nil, err
		}
	}

	// All extproc filters for all Routes are aggregated on HCM and disabled by default
	// Per-route config is used to enable the relevant filters on appropriate routes
	return &hcmv3.HttpFilter{
		Name:     extProcFilterName(extProc),
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: extProcAny,
		},
	}, nil
}

func buildExtProcWrapper(extProc *ir.ExtProc, extProcAny *anypb.Any, matcher *matcherv3.Matcher) (*anypb.Any, error) {
	return anypb.New(&matchingv3.ExtensionWithMatcher{
		ExtensionConfig: &corev3.TypedExtensionConfig{
			Name:        extProcFilterName(extProc),
			TypedConfig: extProcAny,
		},
		XdsMatcher: matcher,
	})
}

func buildExtProcMatcher(matches []ir.ExtProcMatch) (*matcherv3.Matcher, error) {
	if len(matches) == 0 {
		return nil, errors.New("extproc matches cannot be empty")
	}
	matchPredicate, err := buildExtProcMatchesPredicate(matches)
	if err != nil {
		return nil, err
	}
	skipAction, err := anypb.New(&actionv3.SkipFilter{})
	if err != nil {
		return nil, err
	}
	return &matcherv3.Matcher{
		MatcherType: &matcherv3.Matcher_MatcherList_{
			MatcherList: &matcherv3.Matcher_MatcherList{
				Matchers: []*matcherv3.Matcher_MatcherList_FieldMatcher{
					{
						Predicate: wrapPredicateWithNot(matchPredicate, true),
						OnMatch: &matcherv3.Matcher_OnMatch{
							OnMatch: &matcherv3.Matcher_OnMatch_Action{
								Action: &cncfv3.TypedExtensionConfig{
									Name:        "skip-ext-proc",
									TypedConfig: skipAction,
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

// buildExtProcMatchesPredicate converts ExtProc matches into an xDS predicate.
// Multiple match entries are ORed, while all headers in each entry are ANDed.
func buildExtProcMatchesPredicate(matches []ir.ExtProcMatch) (*matcherv3.Matcher_MatcherList_Predicate, error) {
	matchPredicates := make([]*matcherv3.Matcher_MatcherList_Predicate, 0, len(matches))
	for _, match := range matches {
		if len(match.Headers) == 0 {
			return nil, errors.New("extproc match must contain at least one header")
		}

		headerPredicates := make([]*matcherv3.Matcher_MatcherList_Predicate, 0, len(match.Headers))
		for _, header := range match.Headers {
			headerPredicate, err := buildExtProcHeaderMatchPredicate(header)
			if err != nil {
				return nil, err
			}
			headerPredicates = append(headerPredicates, headerPredicate)
		}

		if len(headerPredicates) == 1 {
			matchPredicates = append(matchPredicates, headerPredicates[0])
			continue
		}
		matchPredicates = append(matchPredicates, &matcherv3.Matcher_MatcherList_Predicate{
			MatchType: &matcherv3.Matcher_MatcherList_Predicate_AndMatcher{
				AndMatcher: &matcherv3.Matcher_MatcherList_Predicate_PredicateList{
					Predicate: headerPredicates,
				},
			},
		})
	}

	if len(matchPredicates) == 1 {
		return matchPredicates[0], nil
	}
	return &matcherv3.Matcher_MatcherList_Predicate{
		MatchType: &matcherv3.Matcher_MatcherList_Predicate_OrMatcher{
			OrMatcher: &matcherv3.Matcher_MatcherList_Predicate_PredicateList{
				Predicate: matchPredicates,
			},
		},
	}, nil
}

func buildExtProcHeaderMatchPredicate(header ir.ExtProcHeaderMatch) (*matcherv3.Matcher_MatcherList_Predicate, error) {
	headerInput, err := proto.ToAnyWithValidation(&envoymatcherv3.HttpRequestHeaderMatchInput{
		HeaderName: header.Name,
	})
	if err != nil {
		return nil, err
	}

	return wrapPredicateWithNot(buildHTTPHeaderSinglePredicate(headerInput, &matcherv3.StringMatcher{
		MatchPattern: &matcherv3.StringMatcher_Exact{Exact: header.Value},
	}), header.Invert), nil
}

func extProcFilterName(extProc *ir.ExtProc) string {
	return perRouteFilterName(egv1a1.EnvoyFilterExtProc, extProc.Name)
}

func extProcConfig(extProc *ir.ExtProc) (*extprocv3.ExternalProcessor, error) {
	config := &extprocv3.ExternalProcessor{
		GrpcService: &corev3.GrpcService{
			TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
				EnvoyGrpc: grpcExtProcService(extProc),
			},
			Timeout: durationpb.New(defaultExtServiceRequestTimeout),
		},
	}

	config.ProcessingMode = buildProcessingMode(extProc)

	if extProc.ShadowMode != nil {
		config.ObservabilityMode = *extProc.ShadowMode
	}

	if extProc.FailOpen != nil {
		config.FailureModeAllow = *extProc.FailOpen
	}

	if extProc.MessageTimeout != nil {
		config.MessageTimeout = durationpb.New(extProc.MessageTimeout.Duration)
	}

	if extProc.RequestAttributes != nil {
		config.RequestAttributes = slices.Clone(extProc.RequestAttributes)
	}

	if extProc.ResponseAttributes != nil {
		config.ResponseAttributes = slices.Clone(extProc.ResponseAttributes)
	}

	if extProc.Traffic != nil && extProc.Traffic.Retry != nil {
		rp, err := buildNonRouteRetryPolicy(extProc.Traffic.Retry)
		if err != nil {
			return nil, fmt.Errorf("failed to build retry policy for extproc: %w", err)
		}
		config.GrpcService.RetryPolicy = rp
	}

	if extProc.ForwardingMetadataNamespaces != nil || extProc.ReceivingMetadataNamespaces != nil {
		config.MetadataOptions = &extprocv3.MetadataOptions{}

		if extProc.ForwardingMetadataNamespaces != nil {
			config.MetadataOptions.ForwardingNamespaces = &extprocv3.MetadataOptions_MetadataNamespaces{
				Untyped: slices.Clone(extProc.ForwardingMetadataNamespaces),
			}
		}

		if extProc.ReceivingMetadataNamespaces != nil {
			config.MetadataOptions.ReceivingNamespaces = &extprocv3.MetadataOptions_MetadataNamespaces{
				Untyped: slices.Clone(extProc.ReceivingMetadataNamespaces),
			}
		}
	}
	config.AllowModeOverride = extProc.AllowModeOverride

	if extProc.StatusOnError != nil {
		config.StatusOnError = &typev3.HttpStatus{
			Code: typev3.StatusCode(*extProc.StatusOnError),
		}
	}
	return config, nil
}

func grpcExtProcService(extProc *ir.ExtProc) *corev3.GrpcService_EnvoyGrpc {
	return &corev3.GrpcService_EnvoyGrpc{
		ClusterName: extProc.Destination.Name,
		Authority:   extProc.Authority,
	}
}

// routeContainsExtProc returns true if ExtProcs exists for the provided route.
func routeContainsExtProc(irRoute *ir.HTTPRoute) bool {
	if irRoute == nil {
		return false
	}

	return irRoute.EnvoyExtensions != nil && len(irRoute.EnvoyExtensions.ExtProcs) > 0
}

// patchResources patches the cluster resources for the external services.
func (*extProc) patchResources(tCtx *types.ResourceVersionTable,
	routes []*ir.HTTPRoute,
) error {
	if tCtx == nil || tCtx.XdsResources == nil {
		return errors.New("xds resource table is nil")
	}

	var errs error
	for _, route := range routes {
		if !routeContainsExtProc(route) {
			continue
		}

		for i := range route.EnvoyExtensions.ExtProcs {
			ep := route.EnvoyExtensions.ExtProcs[i]
			if err := createExtServiceXDSCluster(
				&ep.Destination, ep.Traffic, tCtx); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	return errs
}

// patchRoute patches the provided route with the extProc config if applicable.
// Note: this method enables the corresponding extProc filter for the provided route.
func (*extProc) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.EnvoyExtensions == nil {
		return nil
	}

	for i := range irRoute.EnvoyExtensions.ExtProcs {
		ep := &irRoute.EnvoyExtensions.ExtProcs[i]

		filterName := extProcFilterName(ep)

		// For matches-based ext_proc, route-level matcher controls effective gating.
		if len(ep.Matches) > 0 {
			routeCfg, err := buildRouteExtProcMatchOverride(ep)
			if err != nil {
				return err
			}
			if err := enableFilterOnRoute(route, filterName, &routev3.FilterConfig{
				Config: routeCfg,
			}); err != nil {
				return err
			}
			continue
		}

		if err := enableFilterOnRoute(route, filterName, &routev3.FilterConfig{
			Config: &anypb.Any{},
		}); err != nil {
			return err
		}
	}
	return nil
}

func buildRouteExtProcMatchOverride(extProc *ir.ExtProc) (*anypb.Any, error) {
	matcher, err := buildExtProcMatcher(extProc.Matches)
	if err != nil {
		return nil, err
	}
	return anypb.New(&matchingv3.ExtensionWithMatcherPerRoute{
		XdsMatcher: matcher,
	})
}

func buildProcessingMode(extProc *ir.ExtProc) *extprocv3.ProcessingMode {
	processingMode := &extprocv3.ProcessingMode{
		RequestHeaderMode:   extprocv3.ProcessingMode_SKIP,
		ResponseHeaderMode:  extprocv3.ProcessingMode_SKIP,
		RequestBodyMode:     extprocv3.ProcessingMode_NONE,
		ResponseBodyMode:    extprocv3.ProcessingMode_NONE,
		RequestTrailerMode:  extprocv3.ProcessingMode_SKIP,
		ResponseTrailerMode: extprocv3.ProcessingMode_SKIP,
	}

	if extProc.RequestBodyProcessingMode != nil {
		processingMode.RequestBodyMode = translateExtProcBodyProcessingMode(extProc.RequestBodyProcessingMode)
		//
		if processingMode.RequestBodyMode == extprocv3.ProcessingMode_FULL_DUPLEX_STREAMED {
			processingMode.RequestTrailerMode = extprocv3.ProcessingMode_SEND
		}
	}

	if extProc.RequestHeaderProcessing {
		processingMode.RequestHeaderMode = extprocv3.ProcessingMode_SEND
	}

	if extProc.ResponseBodyProcessingMode != nil {
		processingMode.ResponseBodyMode = translateExtProcBodyProcessingMode(extProc.ResponseBodyProcessingMode)
		if processingMode.ResponseBodyMode == extprocv3.ProcessingMode_FULL_DUPLEX_STREAMED {
			processingMode.ResponseTrailerMode = extprocv3.ProcessingMode_SEND
		}
	}

	if extProc.ResponseHeaderProcessing {
		processingMode.ResponseHeaderMode = extprocv3.ProcessingMode_SEND
	}

	return processingMode
}

func translateExtProcBodyProcessingMode(mode *ir.ExtProcBodyProcessingMode) extprocv3.ProcessingMode_BodySendMode {
	lookup := map[ir.ExtProcBodyProcessingMode]extprocv3.ProcessingMode_BodySendMode{
		ir.ExtProcBodyBuffered:           extprocv3.ProcessingMode_BUFFERED,
		ir.ExtProcBodyBufferedPartial:    extprocv3.ProcessingMode_BUFFERED_PARTIAL,
		ir.ExtProcBodyStreamed:           extprocv3.ProcessingMode_STREAMED,
		ir.ExtProcBodyFullDuplexStreamed: extprocv3.ProcessingMode_FULL_DUPLEX_STREAMED,
	}
	if r, found := lookup[*mode]; found {
		return r
	}
	return extprocv3.ProcessingMode_NONE
}
