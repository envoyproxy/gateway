// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"strconv"

	expr "cel.dev/expr"
	cncfv3 "github.com/cncf/xds/go/xds/core/v3"
	matcherv3 "github.com/cncf/xds/go/xds/type/matcher/v3"
	typev3 "github.com/cncf/xds/go/xds/type/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	respv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/custom_response/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	policyv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/custom_response/local_response_policy/v3"
	redirectpolicyv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/custom_response/redirect_policy/v3"
	envoymatcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&customResponse{})
}

type customResponse struct{}

var _ httpFilter = &customResponse{}

// patchHCM builds and appends the customResponse Filters to the HTTP Connection Manager
// if applicable, and it does not already exist.
// Note: this method creates an customResponse filter for each route that contains an ResponseOverride config.
// the filter is disabled by default. It is enabled on the route level.
func (c *customResponse) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	var errs error

	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, route := range irListener.Routes {
		if !c.routeContainsResponseOverride(route) {
			continue
		}

		// Only generates one CustomResponse/CustomRedirect Envoy filter for each unique name.
		// For example, if there are two routes under the same gateway with the
		// same CustomResponse/CustomRedirect config, only one CustomResponse/CustomRedirect filter will be generated.
		if hcmContainsFilter(mgr, c.customResponseFilterName(route.Traffic.ResponseOverride)) {
			continue
		}

		filter, err := c.buildHCMCustomResponseFilter(route.Traffic.ResponseOverride)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return errs
}

// buildHCMCustomResponseFilter returns an Custom Response HTTP filter from the provided IR ResponseOverride.
func (c *customResponse) buildHCMCustomResponseFilter(ro *ir.ResponseOverride) (*hcmv3.HttpFilter, error) {
	config, err := c.customResponseConfig(ro)
	if err != nil {
		return nil, err
	}
	any, err := proto.ToAnyWithValidation(config)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name:     c.customResponseFilterName(ro),
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: any,
		},
	}, nil
}

func (c *customResponse) customResponseFilterName(ro *ir.ResponseOverride) string {
	return perRouteFilterName(egv1a1.EnvoyFilterCustomResponse, ro.Name)
}

func (c *customResponse) customResponseConfig(ro *ir.ResponseOverride) (*respv3.CustomResponse, error) {
	var matchers []*matcherv3.Matcher_MatcherList_FieldMatcher

	for _, r := range ro.Rules {
		var (
			action     *matcherv3.Matcher_OnMatch_Action
			predicates []*matcherv3.Matcher_MatcherList_Predicate
			err        error
		)

		if action, err = c.buildAction(&r); err != nil {
			return nil, err
		}

		if len(r.Match.StatusCodes) > 0 {
			statusCodePredicate, err := c.buildStatusCodePredicate(r.Match.StatusCodes)
			if err != nil {
				return nil, err
			}
			predicates = append(predicates, statusCodePredicate)
		}

		if len(r.Match.ResponseHeaders) > 0 {
			headerPredicate, err := c.buildResponseHeaderPredicate(r.Match.ResponseHeaders)
			if err != nil {
				return nil, err
			}
			predicates = append(predicates, headerPredicate)
		}

		if len(predicates) == 0 {
			// This is just a sanity check, as the CRD validation should have caught this.
			return nil, fmt.Errorf("missing match criteria in response override rule")
		}

		// For Local or Backend sources, AND a local_reply predicate so the rule
		// only fires for the correct response origin.
		if r.Source == egv1a1.ResponseOverrideSourceLocal || r.Source == egv1a1.ResponseOverrideSourceBackend {
			localReplyPredicate, err := c.buildLocalReplyPredicate(r.Source == egv1a1.ResponseOverrideSourceLocal)
			if err != nil {
				return nil, err
			}
			predicates = append(predicates, localReplyPredicate)
		}

		matchers = append(matchers, &matcherv3.Matcher_MatcherList_FieldMatcher{
			Predicate: andPredicate(predicates),
			OnMatch: &matcherv3.Matcher_OnMatch{
				OnMatch: action,
			},
		})
	}

	// Create a MatcherList.
	// The rules will be evaluated in order, and the first match wins.
	cr := &respv3.CustomResponse{
		CustomResponseMatcher: &matcherv3.Matcher{
			MatcherType: &matcherv3.Matcher_MatcherList_{
				MatcherList: &matcherv3.Matcher_MatcherList{
					Matchers: matchers,
				},
			},
		},
	}

	return cr, nil
}

func andPredicate(predicates []*matcherv3.Matcher_MatcherList_Predicate) *matcherv3.Matcher_MatcherList_Predicate {
	if len(predicates) == 1 {
		return predicates[0]
	}
	return &matcherv3.Matcher_MatcherList_Predicate{
		MatchType: &matcherv3.Matcher_MatcherList_Predicate_AndMatcher{
			AndMatcher: &matcherv3.Matcher_MatcherList_Predicate_PredicateList{
				Predicate: predicates,
			},
		},
	}
}

func (c *customResponse) buildStatusCodePredicate(codes []ir.StatusCodeMatch) (*matcherv3.Matcher_MatcherList_Predicate, error) {
	predicates := make([]*matcherv3.Matcher_MatcherList_Predicate, 0, len(codes))
	for _, codeMatch := range codes {
		predicate, err := c.buildSingleStatusCodePredicate(codeMatch)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, predicate)
	}

	if len(predicates) == 1 {
		return predicates[0], nil
	}

	return &matcherv3.Matcher_MatcherList_Predicate{
		MatchType: &matcherv3.Matcher_MatcherList_Predicate_OrMatcher{
			OrMatcher: &matcherv3.Matcher_MatcherList_Predicate_PredicateList{
				Predicate: predicates,
			},
		},
	}, nil
}

func (c *customResponse) buildResponseHeaderPredicate(headers []ir.ResponseOverrideHeaderMatch) (*matcherv3.Matcher_MatcherList_Predicate, error) {
	predicates := make([]*matcherv3.Matcher_MatcherList_Predicate, 0, len(headers))
	for _, header := range headers {
		input, err := c.buildResponseHeaderInput(header.Name)
		if err != nil {
			return nil, err
		}
		valueMatcher, err := buildStringMatcher(header.Value)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, &matcherv3.Matcher_MatcherList_Predicate{
			MatchType: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_{
				SinglePredicate: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate{
					Input: input,
					Matcher: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_ValueMatch{
						ValueMatch: valueMatcher,
					},
				},
			},
		})
	}

	if len(predicates) == 1 {
		return predicates[0], nil
	}

	return &matcherv3.Matcher_MatcherList_Predicate{
		MatchType: &matcherv3.Matcher_MatcherList_Predicate_AndMatcher{
			AndMatcher: &matcherv3.Matcher_MatcherList_Predicate_PredicateList{
				Predicate: predicates,
			},
		},
	}, nil
}

func buildStringMatcher(irMatch ir.StringMatch) (*matcherv3.StringMatcher, error) {
	switch {
	case irMatch.Exact != nil:
		return &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Exact{Exact: *irMatch.Exact},
		}, nil
	case irMatch.Prefix != nil:
		return &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Prefix{Prefix: *irMatch.Prefix},
		}, nil
	case irMatch.Suffix != nil:
		return &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Suffix{Suffix: *irMatch.Suffix},
		}, nil
	case irMatch.SafeRegex != nil:
		return &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_SafeRegex{SafeRegex: &matcherv3.RegexMatcher{
				Regex:      *irMatch.SafeRegex,
				EngineType: &matcherv3.RegexMatcher_GoogleRe2{GoogleRe2: &matcherv3.RegexMatcher_GoogleRE2{}},
			}},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported string matcher")
	}
}

func (c *customResponse) buildSingleStatusCodePredicate(codeMatch ir.StatusCodeMatch) (*matcherv3.Matcher_MatcherList_Predicate, error) {
	var (
		httpAttributeCELInput *cncfv3.TypedExtensionConfig
		statusCodeInput       *cncfv3.TypedExtensionConfig
		statusCodeCELMatcher  *cncfv3.TypedExtensionConfig
		err                   error
	)

	// Use CEL to match a range of status codes.
	if codeMatch.Range != nil {
		if httpAttributeCELInput, err = c.buildHTTPAttributeCELInput(); err != nil {
			return nil, err
		}

		if statusCodeCELMatcher, err = c.buildStatusCodeCELMatcher(*codeMatch.Range); err != nil {
			return nil, err
		}

		return &matcherv3.Matcher_MatcherList_Predicate{
			MatchType: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_{
				SinglePredicate: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate{
					Input: httpAttributeCELInput,
					Matcher: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_CustomMatch{
						CustomMatch: statusCodeCELMatcher,
					},
				},
			},
		}, nil
	} else {
		// Use exact string match to match a single status code.
		if statusCodeInput, err = c.buildStatusCodeInput(); err != nil {
			return nil, err
		}

		return &matcherv3.Matcher_MatcherList_Predicate{
			MatchType: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_{
				SinglePredicate: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate{
					Input: statusCodeInput,
					Matcher: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_ValueMatch{
						ValueMatch: &matcherv3.StringMatcher{
							MatchPattern: &matcherv3.StringMatcher_Exact{
								Exact: strconv.Itoa(*codeMatch.Value),
							},
						},
					},
				},
			},
		}, nil
	}
}

func (c *customResponse) buildHTTPAttributeCELInput() (*cncfv3.TypedExtensionConfig, error) {
	var (
		pb  *anypb.Any
		err error
	)

	if pb, err = proto.ToAnyWithValidation(&matcherv3.HttpAttributesCelMatchInput{}); err != nil {
		return nil, err
	}

	return &cncfv3.TypedExtensionConfig{
		Name:        "http-attributes-cel-match-input",
		TypedConfig: pb,
	}, nil
}

func (c *customResponse) buildStatusCodeInput() (*cncfv3.TypedExtensionConfig, error) {
	var (
		pb  *anypb.Any
		err error
	)

	if pb, err = proto.ToAnyWithValidation(&envoymatcherv3.HttpResponseStatusCodeMatchInput{}); err != nil {
		return nil, err
	}

	return &cncfv3.TypedExtensionConfig{
		Name:        "http-response-status-code-match-input",
		TypedConfig: pb,
	}, nil
}

func (c *customResponse) buildResponseHeaderInput(headerName string) (*cncfv3.TypedExtensionConfig, error) {
	var (
		pb  *anypb.Any
		err error
	)

	if pb, err = proto.ToAnyWithValidation(&envoymatcherv3.HttpResponseHeaderMatchInput{HeaderName: headerName}); err != nil {
		return nil, err
	}

	return &cncfv3.TypedExtensionConfig{
		Name:        "http-response-header-match-input",
		TypedConfig: pb,
	}, nil
}

func (c *customResponse) buildLocalReplyInput() (*cncfv3.TypedExtensionConfig, error) {
	pb, err := proto.ToAnyWithValidation(&envoymatcherv3.HttpResponseLocalReplyMatchInput{})
	if err != nil {
		return nil, err
	}
	return &cncfv3.TypedExtensionConfig{
		Name:        "http-response-local-reply-match-input",
		TypedConfig: pb,
	}, nil
}

func (c *customResponse) buildLocalReplyPredicate(isLocal bool) (*matcherv3.Matcher_MatcherList_Predicate, error) {
	input, err := c.buildLocalReplyInput()
	if err != nil {
		return nil, err
	}
	value := "false"
	if isLocal {
		value = "true"
	}
	return &matcherv3.Matcher_MatcherList_Predicate{
		MatchType: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_{
			SinglePredicate: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate{
				Input: input,
				Matcher: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_ValueMatch{
					ValueMatch: &matcherv3.StringMatcher{
						MatchPattern: &matcherv3.StringMatcher_Exact{
							Exact: value,
						},
					},
				},
			},
		},
	}, nil
}

func (c *customResponse) buildStatusCodeCELMatcher(codeRange ir.StatusCodeRange) (*cncfv3.TypedExtensionConfig, error) {
	var (
		pb  *anypb.Any
		err error
	)

	// Build the CEL expression AST: response.code >= codeRange.Start && response.code <= codeRange.End
	matcher := &matcherv3.CelMatcher{
		ExprMatch: &typev3.CelExpression{
			CelExprParsed: &expr.ParsedExpr{
				Expr: &expr.Expr{
					Id: 9,
					ExprKind: &expr.Expr_CallExpr{
						CallExpr: &expr.Expr_Call{
							Function: "_&&_",
							Args: []*expr.Expr{
								{
									Id: 3,
									ExprKind: &expr.Expr_CallExpr{
										CallExpr: &expr.Expr_Call{
											Function: "_>=_",
											Args: []*expr.Expr{
												{
													Id: 2,
													ExprKind: &expr.Expr_SelectExpr{
														SelectExpr: &expr.Expr_Select{
															Operand: &expr.Expr{
																Id: 1,
																ExprKind: &expr.Expr_IdentExpr{
																	IdentExpr: &expr.Expr_Ident{
																		Name: "response",
																	},
																},
															},
															Field: "code",
														},
													},
												},
												{
													Id: 4,
													ExprKind: &expr.Expr_ConstExpr{
														ConstExpr: &expr.Constant{
															ConstantKind: &expr.Constant_Int64Value{
																Int64Value: int64(codeRange.Start),
															},
														},
													},
												},
											},
										},
									},
								},
								{
									Id: 7,
									ExprKind: &expr.Expr_CallExpr{
										CallExpr: &expr.Expr_Call{
											Function: "_<=_",
											Args: []*expr.Expr{
												{
													Id: 6,
													ExprKind: &expr.Expr_SelectExpr{
														SelectExpr: &expr.Expr_Select{
															Operand: &expr.Expr{
																Id: 5,
																ExprKind: &expr.Expr_IdentExpr{
																	IdentExpr: &expr.Expr_Ident{
																		Name: "response",
																	},
																},
															},
															Field: "code",
														},
													},
												},
												{
													Id: 8,
													ExprKind: &expr.Expr_ConstExpr{
														ConstExpr: &expr.Constant{
															ConstantKind: &expr.Constant_Int64Value{
																Int64Value: int64(codeRange.End),
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	if pb, err = proto.ToAnyWithValidation(matcher); err != nil {
		return nil, err
	}

	return &cncfv3.TypedExtensionConfig{
		Name:        "cel-matcher",
		TypedConfig: pb,
	}, nil
}

func (c *customResponse) buildAction(r *ir.ResponseOverrideRule) (*matcherv3.Matcher_OnMatch_Action, error) {
	var (
		pb  *anypb.Any
		err error
	)

	if r.Redirect != nil {
		pb, err = c.buildRedirectAction(r)
	} else {
		pb, err = c.buildResponseAction(r)
	}

	if err != nil {
		return nil, err
	}
	return &matcherv3.Matcher_OnMatch_Action{
		Action: &cncfv3.TypedExtensionConfig{
			Name:        r.Name,
			TypedConfig: pb,
		},
	}, nil
}

func (c *customResponse) buildRedirectAction(r *ir.ResponseOverrideRule) (*anypb.Any, error) {
	redirectAction := &routev3.RedirectAction{}
	if r.Redirect.Scheme != nil {
		redirectAction.SchemeRewriteSpecifier = &routev3.RedirectAction_SchemeRedirect{
			SchemeRedirect: *r.Redirect.Scheme,
		}
	}
	if r.Redirect.Hostname != nil {
		redirectAction.HostRedirect = *r.Redirect.Hostname
	}
	if r.Redirect.Port != nil {
		redirectAction.PortRedirect = *r.Redirect.Port
	}
	if r.Redirect.Path != nil && r.Redirect.Path.FullReplace != nil {
		redirectAction.PathRewriteSpecifier = &routev3.RedirectAction_PathRedirect{
			PathRedirect: *r.Redirect.Path.FullReplace,
		}
	}
	redirect := &redirectpolicyv3.RedirectPolicy{
		RedirectActionSpecifier: &redirectpolicyv3.RedirectPolicy_RedirectAction{
			RedirectAction: redirectAction,
		},
		StatusCode: wrapperspb.UInt32(uint32(*r.Redirect.StatusCode)),
	}

	return proto.ToAnyWithValidation(redirect)
}

func (c *customResponse) buildResponseAction(r *ir.ResponseOverrideRule) (*anypb.Any, error) {
	response := &policyv3.LocalResponsePolicy{}

	if len(r.Response.Body) > 0 {
		response.BodyFormat = &corev3.SubstitutionFormatString{
			Format: &corev3.SubstitutionFormatString_TextFormat{
				TextFormat: string(r.Response.Body),
			},
		}
	}

	if r.Response.ContentType != nil && *r.Response.ContentType != "" {
		response.ResponseHeadersToAdd = append(response.ResponseHeadersToAdd, &corev3.HeaderValueOption{
			Header: &corev3.HeaderValue{
				Key:   "Content-Type",
				Value: *r.Response.ContentType,
			},
			AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
		})
	}

	if r.Response.StatusCode != nil {
		response.StatusCode = &wrapperspb.UInt32Value{Value: *r.Response.StatusCode}
	}

	if r.Response.AddResponseHeaders != nil {
		response.ResponseHeadersToAdd = append(response.ResponseHeadersToAdd, buildXdsAddedHeaders(r.Response.AddResponseHeaders)...)
	}

	return proto.ToAnyWithValidation(response)
}

// routeContainsResponseOverride returns true if ResponseOverride exists for the provided route.
func (c *customResponse) routeContainsResponseOverride(irRoute *ir.HTTPRoute) bool {
	if irRoute != nil &&
		irRoute.Traffic != nil &&
		irRoute.Traffic.ResponseOverride != nil {
		return true
	}
	return false
}

func (c *customResponse) patchResources(_ *types.ResourceVersionTable, _ []*ir.HTTPRoute) error {
	return nil
}

// patchRoute patches the provided route with the customResponse config if applicable.
// Note: this method enables the corresponding customResponse filter for the provided route.
func (c *customResponse) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.Traffic == nil || irRoute.Traffic.ResponseOverride == nil {
		return nil
	}
	filterName := c.customResponseFilterName(irRoute.Traffic.ResponseOverride)
	if err := enableFilterOnRoute(route, filterName, &routev3.FilterConfig{
		Config: &anypb.Any{},
	}); err != nil {
		return err
	}
	return nil
}
