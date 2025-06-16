// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"strconv"

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
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
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

// buildHCMCustomResponseFilter returns an OAuth2 HTTP filter from the provided IR HTTPRoute.
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
			action    *matcherv3.Matcher_OnMatch_Action
			predicate *matcherv3.Matcher_MatcherList_Predicate
			err       error
		)

		if action, err = c.buildAction(r); err != nil {
			return nil, err
		}

		switch {
		case len(r.Match.StatusCodes) == 0:
			// This is just a sanity check, as the CRD validation should have caught this.
			return nil, fmt.Errorf("missing status code in response override rule")
		case len(r.Match.StatusCodes) == 1:
			if predicate, err = c.buildSinglePredicate(r.Match.StatusCodes[0]); err != nil {
				return nil, err
			}

			matcher := &matcherv3.Matcher_MatcherList_FieldMatcher{
				Predicate: predicate,
				OnMatch: &matcherv3.Matcher_OnMatch{
					OnMatch: action,
				},
			}

			matchers = append(matchers, matcher)
		case len(r.Match.StatusCodes) > 1:
			var predicates []*matcherv3.Matcher_MatcherList_Predicate

			for _, codeMatch := range r.Match.StatusCodes {
				if predicate, err = c.buildSinglePredicate(codeMatch); err != nil {
					return nil, err
				}

				predicates = append(predicates, predicate)
			}

			// Create a single matcher that ORs all the predicates together.
			// The rule will match if any of the codes match.
			matcher := &matcherv3.Matcher_MatcherList_FieldMatcher{
				Predicate: &matcherv3.Matcher_MatcherList_Predicate{
					MatchType: &matcherv3.Matcher_MatcherList_Predicate_OrMatcher{
						OrMatcher: &matcherv3.Matcher_MatcherList_Predicate_PredicateList{
							Predicate: predicates,
						},
					},
				},
				OnMatch: &matcherv3.Matcher_OnMatch{
					OnMatch: action,
				},
			}

			matchers = append(matchers, matcher)
		}

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

func (c *customResponse) buildSinglePredicate(codeMatch ir.StatusCodeMatch) (*matcherv3.Matcher_MatcherList_Predicate, error) {
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

func (c *customResponse) buildStatusCodeCELMatcher(codeRange ir.StatusCodeRange) (*cncfv3.TypedExtensionConfig, error) {
	var (
		pb  *anypb.Any
		err error
	)

	// Build the CEL expression AST: response.code >= codeRange.Start && response.code <= codeRange.End
	matcher := &matcherv3.CelMatcher{
		ExprMatch: &typev3.CelExpression{
			ExprSpecifier: &typev3.CelExpression_ParsedExpr{
				ParsedExpr: &expr.ParsedExpr{
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

func (c *customResponse) buildAction(r ir.ResponseOverrideRule) (*matcherv3.Matcher_OnMatch_Action, error) {
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

func (c *customResponse) buildRedirectAction(r ir.ResponseOverrideRule) (*anypb.Any, error) {
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

func (c *customResponse) buildResponseAction(r ir.ResponseOverrideRule) (*anypb.Any, error) {
	response := &policyv3.LocalResponsePolicy{}

	if r.Response.Body != nil && *r.Response.Body != "" {
		response.BodyFormat = &corev3.SubstitutionFormatString{
			Format: &corev3.SubstitutionFormatString_TextFormat{
				TextFormat: *r.Response.Body,
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

func (c *customResponse) patchResources(tCtx *types.ResourceVersionTable,
	routes []*ir.HTTPRoute,
) error {
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
	if err := enableFilterOnRoute(route, filterName); err != nil {
		return err
	}
	return nil
}
