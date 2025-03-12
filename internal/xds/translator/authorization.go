// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"strings"

	cncfv3 "github.com/cncf/xds/go/xds/core/v3"
	matcherv3 "github.com/cncf/xds/go/xds/type/matcher/v3"
	configv3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	rbacconfigv3 "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	rbacv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	networkinput "github.com/envoyproxy/go-control-plane/envoy/extensions/matching/common_inputs/network/v3"
	ipmatcherv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/matching/input_matchers/ip/v3"
	metadatav3 "github.com/envoyproxy/go-control-plane/envoy/extensions/matching/input_matchers/metadata/v3"
	envoymatcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&rbac{})
}

type rbac struct{}

var _ httpFilter = &rbac{}

// patchHCM builds and appends the RBAC Filter to the HTTP Connection Manager if
// applicable.
func (*rbac) patchHCM(
	mgr *hcmv3.HttpConnectionManager,
	irListener *ir.HTTPListener,
) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	if !listenerContainsRBAC(irListener) {
		return nil
	}

	// Return early if filter already exists.
	for _, f := range mgr.HttpFilters {
		if f.Name == egv1a1.EnvoyFilterRBAC.String() {
			return nil
		}
	}

	rbacFilter, err := buildHCMRBACFilter()
	if err != nil {
		return err
	}

	mgr.HttpFilters = append([]*hcmv3.HttpFilter{rbacFilter}, mgr.HttpFilters...)

	return nil
}

// buildHCMRBACFilter returns a RBAC filter from the provided IR listener.
func buildHCMRBACFilter() (*hcmv3.HttpFilter, error) {
	rbacProto := &rbacv3.RBAC{}
	rbacAny, err := proto.ToAnyWithValidation(rbacProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: egv1a1.EnvoyFilterRBAC.String(),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: rbacAny,
		},
	}, nil
}

// listenerContainsRBAC returns true if the provided listener has RBAC
// policies attached to its routes.
func listenerContainsRBAC(irListener *ir.HTTPListener) bool {
	if irListener == nil {
		return false
	}

	for _, route := range irListener.Routes {
		if route.Security != nil && route.Security.Authorization != nil {
			return true
		}
	}

	return false
}

// patchRoute patches the provided route with the RBAC config if applicable.
func (*rbac) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.Security == nil || irRoute.Security.Authorization == nil {
		return nil
	}

	filterCfg := route.GetTypedPerFilterConfig()
	if _, ok := filterCfg[egv1a1.EnvoyFilterRBAC.String()]; ok {
		// This should not happen since this is the only place where the RBAC
		// filter is added in a route.
		return fmt.Errorf("route already contains rbac config: %+v", route)
	}

	var (
		rbacPerRoute *rbacv3.RBACPerRoute
		cfgAny       *anypb.Any
		err          error
	)

	if rbacPerRoute, err = buildRBACPerRoute(irRoute.Security.Authorization); err != nil {
		return err
	}

	if cfgAny, err = proto.ToAnyWithValidation(rbacPerRoute); err != nil {
		return err
	}

	if filterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	route.TypedPerFilterConfig[egv1a1.EnvoyFilterRBAC.String()] = cfgAny

	return nil
}

func buildRBACPerRoute(authorization *ir.Authorization) (*rbacv3.RBACPerRoute, error) {
	var (
		rbac        *rbacv3.RBACPerRoute
		allowAction *anypb.Any
		denyAction  *anypb.Any
		matcherList []*matcherv3.Matcher_MatcherList_FieldMatcher
		err         error
	)

	allow := &rbacconfigv3.Action{
		Name:   "ALLOW",
		Action: rbacconfigv3.RBAC_ALLOW,
	}
	if allowAction, err = proto.ToAnyWithValidation(allow); err != nil {
		return nil, err
	}

	deny := &rbacconfigv3.Action{
		Name:   "DENY",
		Action: rbacconfigv3.RBAC_DENY,
	}
	if denyAction, err = proto.ToAnyWithValidation(deny); err != nil {
		return nil, err
	}

	// Build a list of matchers based on the rules.
	// The matchers will be evaluated in order, and the first one that matches
	// will be used to determine the action, the rest of the matchers will be
	// skipped.
	// If no matcher matches, the default action will be used.
	for _, rule := range authorization.Rules {
		var (
			// Predicates for HTTP methods.
			methodPredicate *matcherv3.Matcher_MatcherList_Predicate

			// Predicates for HTTP headers.
			headerPredicate []*matcherv3.Matcher_MatcherList_Predicate

			// Predicates for IP ranges.
			ipPredicate *matcherv3.Matcher_MatcherList_Predicate

			// Predicates for JWT claims and scopes.
			jwtPredicate []*matcherv3.Matcher_MatcherList_Predicate

			// The final predicate that will be used for the current rule.
			finalPredicate *matcherv3.Matcher_MatcherList_Predicate
		)

		// Determine the action for the current rule.
		ruleAction := allowAction
		if rule.Action == egv1a1.AuthorizationActionDeny {
			ruleAction = denyAction
		}

		if len(rule.Principal.ClientCIDRs) > 0 {
			if ipPredicate, err = buildIPPredicate(rule.Principal.ClientCIDRs); err != nil {
				return nil, err
			}
		}

		if rule.Principal.JWT != nil {
			if jwtPredicate, err = buildJWTPredicate(*rule.Principal.JWT); err != nil {
				return nil, err
			}
		}

		var methodPredicates []*matcherv3.Matcher_MatcherList_Predicate
		if rule.Operation != nil && len(rule.Operation.Methods) > 0 {
			if methodPredicates, err = buildMethodsPredicate(rule.Operation.Methods); err != nil {
				return nil, err
			}
		}

		// If there are multiple methods, OR them together.
		// Methods are matched if any of them match.
		switch {
		case len(methodPredicates) > 1:
			methodPredicate = &matcherv3.Matcher_MatcherList_Predicate{
				MatchType: &matcherv3.Matcher_MatcherList_Predicate_OrMatcher{
					OrMatcher: &matcherv3.Matcher_MatcherList_Predicate_PredicateList{
						Predicate: methodPredicates,
					},
				},
			}
		case len(methodPredicates) == 1:
			methodPredicate = &matcherv3.Matcher_MatcherList_Predicate{
				MatchType: methodPredicates[0].MatchType.(*matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_),
			}
		}

		if len(rule.Principal.Headers) > 0 {
			if headerPredicate, err = buildHeadersPredicate(rule.Principal.Headers); err != nil {
				return nil, err
			}
		}

		// AND all the predicates together.
		var allPredicates []*matcherv3.Matcher_MatcherList_Predicate
		if methodPredicate != nil {
			allPredicates = append(allPredicates, methodPredicate)
		}
		if ipPredicate != nil {
			allPredicates = append(allPredicates, ipPredicate)
		}
		allPredicates = append(allPredicates, jwtPredicate...)
		allPredicates = append(allPredicates, headerPredicate...)

		switch {
		case len(allPredicates) > 1:
			finalPredicate = &matcherv3.Matcher_MatcherList_Predicate{
				MatchType: &matcherv3.Matcher_MatcherList_Predicate_AndMatcher{
					AndMatcher: &matcherv3.Matcher_MatcherList_Predicate_PredicateList{
						Predicate: allPredicates,
					},
				},
			}
		case len(allPredicates) == 1:
			finalPredicate = allPredicates[0]
		}

		// Add the matcher generated with the current rule to the matcher list.
		// The first matcher that matches will be used to determine the action.
		matcherList = append(matcherList, &matcherv3.Matcher_MatcherList_FieldMatcher{
			Predicate: finalPredicate,
			OnMatch: &matcherv3.Matcher_OnMatch{
				OnMatch: &matcherv3.Matcher_OnMatch_Action{
					Action: &cncfv3.TypedExtensionConfig{
						Name:        rule.Name,
						TypedConfig: ruleAction,
					},
				},
			},
		})
	}

	// Set the default action.
	defaultAction := denyAction
	if authorization.DefaultAction == egv1a1.AuthorizationActionAllow {
		defaultAction = allowAction
	}

	rbac = &rbacv3.RBACPerRoute{
		Rbac: &rbacv3.RBAC{
			Matcher: &matcherv3.Matcher{
				MatcherType: &matcherv3.Matcher_MatcherList_{
					MatcherList: &matcherv3.Matcher_MatcherList{
						Matchers: matcherList,
					},
				},
				// If no matcher matches, the default action will be used.
				OnNoMatch: &matcherv3.Matcher_OnMatch{
					OnMatch: &matcherv3.Matcher_OnMatch_Action{
						Action: &cncfv3.TypedExtensionConfig{
							Name:        "default",
							TypedConfig: defaultAction,
						},
					},
				},
			},
		},
	}

	// If there are no matchers, the default action will be used for all requests.
	// Setting the matcher type to nil since Proto validation will fail if the list
	// is empty.
	if len(matcherList) == 0 {
		rbac.Rbac.Matcher.MatcherType = nil
	}

	return rbac, nil
}

func buildIPPredicate(clientCIDRs []*ir.CIDRMatch) (*matcherv3.Matcher_MatcherList_Predicate, error) {
	var (
		sourceIPInput *anypb.Any
		ipMatcher     *anypb.Any
		err           error
	)

	// Build the IPMatcher based on the client CIDRs.
	ipRangeMatcher := &ipmatcherv3.Ip{
		StatPrefix: "client_ip",
	}

	for _, cidr := range clientCIDRs {
		ipRangeMatcher.CidrRanges = append(ipRangeMatcher.CidrRanges, &configv3.CidrRange{
			AddressPrefix: cidr.IP,
			PrefixLen: &wrapperspb.UInt32Value{
				Value: cidr.MaskLen,
			},
		})
	}

	if ipMatcher, err = proto.ToAnyWithValidation(ipRangeMatcher); err != nil {
		return nil, err
	}

	if sourceIPInput, err = proto.ToAnyWithValidation(&networkinput.SourceIPInput{}); err != nil {
		return nil, err
	}

	return &matcherv3.Matcher_MatcherList_Predicate{
		MatchType: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_{
			SinglePredicate: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate{
				Input: &cncfv3.TypedExtensionConfig{
					Name:        "client_ip",
					TypedConfig: sourceIPInput,
				},
				Matcher: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_CustomMatch{
					CustomMatch: &cncfv3.TypedExtensionConfig{
						Name:        "ip_matcher",
						TypedConfig: ipMatcher,
					},
				},
			},
		},
	}, nil
}

func buildJWTPredicate(jwt egv1a1.JWTPrincipal) ([]*matcherv3.Matcher_MatcherList_Predicate, error) {
	jwtPredicate := []*matcherv3.Matcher_MatcherList_Predicate{}

	// Build the scope matchers.
	// Multiple scopes are ANDed together.
	for _, scope := range jwt.Scopes {
		var (
			inputPb   *anypb.Any
			matcherPb *anypb.Any
			err       error
		)

		input := &networkinput.DynamicMetadataInput{
			Filter: "envoy.filters.http.jwt_authn",
			Path: []*networkinput.DynamicMetadataInput_PathSegment{
				{
					Segment: &networkinput.DynamicMetadataInput_PathSegment_Key{
						Key: jwt.Provider, // The name of the jwt provider is used as the `payload_in_metadata` in the JWT Authn filter.
					},
				},
				{
					Segment: &networkinput.DynamicMetadataInput_PathSegment_Key{
						Key: "scope",
					},
				},
			},
		}

		// The scope has already been normalized to a string array in the JWT Authn filter.
		scopeMatcher := &metadatav3.Metadata{
			Value: &envoymatcherv3.ValueMatcher{
				MatchPattern: &envoymatcherv3.ValueMatcher_ListMatch{
					ListMatch: &envoymatcherv3.ListMatcher{
						MatchPattern: &envoymatcherv3.ListMatcher_OneOf{
							OneOf: &envoymatcherv3.ValueMatcher{
								MatchPattern: &envoymatcherv3.ValueMatcher_StringMatch{
									StringMatch: &envoymatcherv3.StringMatcher{
										MatchPattern: &envoymatcherv3.StringMatcher_Exact{
											Exact: string(scope),
										},
									},
								},
							},
						},
					},
				},
			},
		}

		if inputPb, err = proto.ToAnyWithValidation(input); err != nil {
			return nil, err
		}

		if matcherPb, err = proto.ToAnyWithValidation(scopeMatcher); err != nil {
			return nil, err
		}

		scopePredicate := matcherv3.Matcher_MatcherList_Predicate_SinglePredicate{
			Input: &cncfv3.TypedExtensionConfig{
				Name:        "scope",
				TypedConfig: inputPb,
			},
			Matcher: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_CustomMatch{
				CustomMatch: &cncfv3.TypedExtensionConfig{
					Name:        "scope_matcher",
					TypedConfig: matcherPb,
				},
			},
		}

		jwtPredicate = append(jwtPredicate,
			&matcherv3.Matcher_MatcherList_Predicate{
				MatchType: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_{
					SinglePredicate: &scopePredicate,
				},
			},
		)
	}

	// Build the claim matchers.
	// Multiple claims are ANDed together.
	// Multiple values for a claim are ORed together.
	// For example, if we have two claims: "claim1" with values ["value1", "value2"], and "claim2" with values ["value3", "value4"],
	// the resulting matcher will be: (claim1 == value1 OR claim1 == value2) AND (claim2 == value3 OR claim2 == value4).
	predicateForAllClaims := []*matcherv3.Matcher_MatcherList_Predicate{}
	for _, claim := range jwt.Claims {
		var (
			inputPb   *anypb.Any
			matcherPb *anypb.Any
			err       error
		)

		path := []*networkinput.DynamicMetadataInput_PathSegment{
			{
				Segment: &networkinput.DynamicMetadataInput_PathSegment_Key{
					Key: jwt.Provider, // The name of the jwt provider is used as the `payload_in_metadata` in the JWT Authn filter.
				},
			},
		}

		// A nested claim is represented as a dot-separated string, e.g., "user.email".
		for _, segment := range strings.Split(claim.Name, ".") {
			path = append(path, &networkinput.DynamicMetadataInput_PathSegment{
				Segment: &networkinput.DynamicMetadataInput_PathSegment_Key{
					Key: segment,
				},
			})
		}

		input := &networkinput.DynamicMetadataInput{
			Filter: "envoy.filters.http.jwt_authn",
			Path:   path,
		}

		if inputPb, err = proto.ToAnyWithValidation(input); err != nil {
			return nil, err
		}

		predicateForOneClaim := []*matcherv3.Matcher_MatcherList_Predicate{}
		for _, value := range claim.Values {
			var valueMatcher *envoymatcherv3.ValueMatcher

			if claim.ValueType != nil && *claim.ValueType == egv1a1.JWTClaimValueTypeStringArray {
				valueMatcher = &envoymatcherv3.ValueMatcher{
					MatchPattern: &envoymatcherv3.ValueMatcher_ListMatch{
						ListMatch: &envoymatcherv3.ListMatcher{
							MatchPattern: &envoymatcherv3.ListMatcher_OneOf{
								OneOf: &envoymatcherv3.ValueMatcher{
									MatchPattern: &envoymatcherv3.ValueMatcher_StringMatch{
										StringMatch: &envoymatcherv3.StringMatcher{
											MatchPattern: &envoymatcherv3.StringMatcher_Exact{
												Exact: value,
											},
										},
									},
								},
							},
						},
					},
				}
			} else {
				valueMatcher = &envoymatcherv3.ValueMatcher{
					MatchPattern: &envoymatcherv3.ValueMatcher_StringMatch{
						StringMatch: &envoymatcherv3.StringMatcher{
							MatchPattern: &envoymatcherv3.StringMatcher_Exact{
								Exact: value,
							},
						},
					},
				}
			}

			if matcherPb, err = proto.ToAnyWithValidation(&metadatav3.Metadata{
				Value: valueMatcher,
			}); err != nil {
				return nil, err
			}

			predicateForOneClaim = append(predicateForOneClaim, &matcherv3.Matcher_MatcherList_Predicate{
				MatchType: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_{
					SinglePredicate: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate{
						Input: &cncfv3.TypedExtensionConfig{
							Name:        "claim",
							TypedConfig: inputPb,
						},
						Matcher: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_CustomMatch{
							CustomMatch: &cncfv3.TypedExtensionConfig{
								Name:        "claim_matcher",
								TypedConfig: matcherPb,
							},
						},
					},
				},
			})
		}

		// For a claim to match, one of the values must match.
		// If there are multiple values for a claim, OR them together.
		if len(predicateForOneClaim) > 1 {
			predicateForAllClaims = append(predicateForAllClaims, &matcherv3.Matcher_MatcherList_Predicate{
				MatchType: &matcherv3.Matcher_MatcherList_Predicate_OrMatcher{
					OrMatcher: &matcherv3.Matcher_MatcherList_Predicate_PredicateList{
						Predicate: predicateForOneClaim,
					},
				},
			})
		} else if len(predicateForOneClaim) == 1 {
			predicateForAllClaims = append(predicateForAllClaims, &matcherv3.Matcher_MatcherList_Predicate{
				MatchType: predicateForOneClaim[0].MatchType.(*matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_),
			})
		}
	}

	// For a JWT principal to match, all the specified claims and scopes must match.
	// And all the claims and scopes together.
	jwtPredicate = append(jwtPredicate, predicateForAllClaims...)

	return jwtPredicate, nil
}

func (c *rbac) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

func buildMethodsPredicate(methods []gwapiv1.HTTPMethod) ([]*matcherv3.Matcher_MatcherList_Predicate, error) {
	methodStrings := make([]string, len(methods))
	for i, method := range methods {
		methodStrings[i] = string(method)
	}

	// Match the HTTP method as a pesudo-header.
	return buildHeaderPredicate(":method", methodStrings, true)
}

func buildHeadersPredicate(headers []egv1a1.AuthorizationHeaderMatch) ([]*matcherv3.Matcher_MatcherList_Predicate, error) {
	var (
		headersPredicates []*matcherv3.Matcher_MatcherList_Predicate // Predicates for all headers.
		headerPredicates  []*matcherv3.Matcher_MatcherList_Predicate // Predicates for a single header.
		err               error
	)

	for _, header := range headers {
		if headerPredicates, err = buildHeaderPredicate(header.Name, header.Values, false); err != nil {
			return nil, err
		}

		// For a header to match, one of the values must match.
		// If there are multiple values for a header, OR them together.
		if len(headerPredicates) > 1 {
			headersPredicates = append(headersPredicates, &matcherv3.Matcher_MatcherList_Predicate{
				MatchType: &matcherv3.Matcher_MatcherList_Predicate_OrMatcher{
					OrMatcher: &matcherv3.Matcher_MatcherList_Predicate_PredicateList{
						Predicate: headerPredicates,
					},
				},
			})
		} else if len(headerPredicates) == 1 {
			headersPredicates = append(headersPredicates, &matcherv3.Matcher_MatcherList_Predicate{
				MatchType: headerPredicates[0].MatchType.(*matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_),
			})
		}
	}

	return headersPredicates, nil
}

func buildHeaderPredicate(name string, values []string, ignoreCase bool) ([]*matcherv3.Matcher_MatcherList_Predicate, error) {
	var (
		headerMatchInput *anypb.Any
		err              error
	)

	if headerMatchInput, err = proto.ToAnyWithValidation(&envoymatcherv3.HttpRequestHeaderMatchInput{
		HeaderName: name,
	}); err != nil {
		return nil, err
	}

	var predicates []*matcherv3.Matcher_MatcherList_Predicate
	for _, value := range values {
		predicates = append(predicates, &matcherv3.Matcher_MatcherList_Predicate{
			MatchType: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_{
				SinglePredicate: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate{
					Input: &cncfv3.TypedExtensionConfig{
						Name:        "http_header",
						TypedConfig: headerMatchInput,
					},
					Matcher: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_ValueMatch{
						ValueMatch: &matcherv3.StringMatcher{
							MatchPattern: &matcherv3.StringMatcher_Exact{
								Exact: value,
							},
							IgnoreCase: ignoreCase,
						},
					},
				},
			},
		})
	}
	return predicates, nil
}
