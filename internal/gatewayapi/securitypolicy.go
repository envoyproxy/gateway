// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/mail"
	"net/netip"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	perr "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
)

const (
	defaultRedirectURL           = "%REQ(x-forwarded-proto)%://%REQ(:authority)%/oauth2/callback"
	defaultRedirectPath          = "/oauth2/callback"
	defaultLogoutPath            = "/logout"
	defaultForwardAccessToken    = false
	defaultRefreshToken          = false
	defaultPassThroughAuthHeader = false

	// nolint: gosec
	oidcHMACSecretName = "envoy-oidc-hmac"
	oidcHMACSecretKey  = "hmac-secret"
)

func (t *Translator) ProcessSecurityPolicies(securityPolicies []*egv1a1.SecurityPolicy,
	gateways []*GatewayContext,
	routes []RouteContext,
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
) []*egv1a1.SecurityPolicy {
	var res []*egv1a1.SecurityPolicy

	// Sort based on timestamp
	sort.Slice(securityPolicies, func(i, j int) bool {
		return securityPolicies[i].CreationTimestamp.Before(&(securityPolicies[j].CreationTimestamp))
	})

	// First build a map out of the routes and gateways for faster lookup since users might have thousands of routes or more.
	// For gateways this probably isn't quite as necessary.
	routeMap := map[policyTargetRouteKey]*policyRouteTargetContext{}
	for _, route := range routes {
		key := policyTargetRouteKey{
			Kind:      string(GetRouteType(route)),
			Name:      route.GetName(),
			Namespace: route.GetNamespace(),
		}
		routeMap[key] = &policyRouteTargetContext{RouteContext: route}
	}
	gatewayMap := map[types.NamespacedName]*policyGatewayTargetContext{}
	for _, gw := range gateways {
		key := utils.NamespacedName(gw)
		gatewayMap[key] = &policyGatewayTargetContext{GatewayContext: gw, attachedToListeners: make(sets.Set[string])}
	}

	// Map of Gateway to the routes attached to it.
	// The routes are grouped by sectionNames of their targetRefs
	gatewayRouteMap := make(map[string]map[string]sets.Set[string])

	handledPolicies := make(map[types.NamespacedName]*egv1a1.SecurityPolicy)

	// Translate
	// 1. First translate Policies targeting xRoutes
	// 2. Then translate Policies targeting Listeners
	// 3. Finally, the policies targeting Gateways

	// Process the policies targeting xRoutes
	for _, currPolicy := range securityPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, routes)
		for _, currTarget := range targetRefs {
			if currTarget.Kind != resource.KindGateway {
				var (
					targetedRoute  RouteContext
					parentGateways []gwapiv1a2.ParentReference
					resolveErr     *status.PolicyResolveError
				)
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy.DeepCopy()
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				targetedRoute, resolveErr = resolveSecurityPolicyRouteTargetRef(policy, currTarget, routeMap)
				// Skip if the route is not found
				// It's not necessarily an error because the SecurityPolicy may be
				// reconciled by multiple controllers. And the other controller may
				// have the target route.
				if targetedRoute == nil {
					continue
				}

				// Find the parent Gateways for the route and add it to the
				// gatewayRouteMap, which will be used to check policy override.
				// The parent gateways are also used to set the status of the policy.
				parentRefs := GetParentReferences(targetedRoute)
				for _, p := range parentRefs {
					if p.Kind == nil || *p.Kind == resource.KindGateway {
						namespace := targetedRoute.GetNamespace()
						if p.Namespace != nil {
							namespace = string(*p.Namespace)
						}
						gwNN := types.NamespacedName{
							Namespace: namespace,
							Name:      string(p.Name),
						}

						key := gwNN.String()
						if _, ok := gatewayRouteMap[key]; !ok {
							gatewayRouteMap[key] = make(map[string]sets.Set[string])
						}
						listenerRouteMap := gatewayRouteMap[key]
						sectionName := ""
						if p.SectionName != nil {
							sectionName = string(*p.SectionName)
						}
						if _, ok := listenerRouteMap[sectionName]; !ok {
							listenerRouteMap[sectionName] = make(sets.Set[string])
						}
						listenerRouteMap[sectionName].Insert(utils.NamespacedName(targetedRoute).String())
						parentGateways = append(parentGateways, getAncestorRefForPolicy(gwNN, p.SectionName))
					}
				}

				// Set conditions for resolve error, then skip current xroute
				if resolveErr != nil {
					status.SetResolveErrorForPolicyAncestors(&policy.Status,
						parentGateways,
						t.GatewayControllerName,
						policy.Generation,
						resolveErr,
					)

					continue
				}

				if err := validateSecurityPolicy(policy); err != nil {
					status.SetTranslationErrorForPolicyAncestors(&policy.Status,
						parentGateways,
						t.GatewayControllerName,
						policy.Generation,
						status.Error2ConditionMsg(fmt.Errorf("invalid SecurityPolicy: %w", err)),
					)

					continue
				}

				if err := t.translateSecurityPolicyForRoute(policy, targetedRoute, resources, xdsIR); err != nil {
					status.SetTranslationErrorForPolicyAncestors(&policy.Status,
						parentGateways,
						t.GatewayControllerName,
						policy.Generation,
						status.Error2ConditionMsg(err),
					)
				}

				// Set Accepted condition if it is unset
				status.SetAcceptedForPolicyAncestors(&policy.Status, parentGateways, t.GatewayControllerName)
			}
		}
	}

	// Process the policies targeting Listeners
	for _, currPolicy := range securityPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, gateways)
		for _, currTarget := range targetRefs {
			if currTarget.Kind == resource.KindGateway && currTarget.SectionName != nil {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy.DeepCopy()
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				t.processSecurityPolicyForGateway(resources, xdsIR,
					gatewayMap, gatewayRouteMap, policy, currTarget)
			}
		}
	}
	// Process the policies targeting Gateways
	for _, currPolicy := range securityPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, gateways)
		for _, currTarget := range targetRefs {
			if currTarget.Kind == resource.KindGateway && currTarget.SectionName == nil {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy.DeepCopy()
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				t.processSecurityPolicyForGateway(resources, xdsIR,
					gatewayMap, gatewayRouteMap, policy, currTarget)
			}
		}
	}

	return res
}

func (t *Translator) processSecurityPolicyForGateway(
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
	gatewayMap map[types.NamespacedName]*policyGatewayTargetContext,
	gatewayRouteMap map[string]map[string]sets.Set[string],
	policy *egv1a1.SecurityPolicy,
	currTarget gwapiv1a2.LocalPolicyTargetReferenceWithSectionName,
) {
	var (
		targetedGateway *GatewayContext
		resolveErr      *status.PolicyResolveError
	)

	targetedGateway, resolveErr = resolveSecurityPolicyGatewayTargetRef(policy, currTarget, gatewayMap)
	// Skip if the gateway is not found
	// It's not necessarily an error because the SecurityPolicy may be
	// reconciled by multiple controllers. And the other controller may
	// have the target gateway.
	if targetedGateway == nil {
		return
	}

	// Find its ancestor reference by resolved gateway, even with resolve error
	gatewayNN := utils.NamespacedName(targetedGateway)
	parentGateways := []gwapiv1a2.ParentReference{
		getAncestorRefForPolicy(gatewayNN, currTarget.SectionName),
	}

	// Set conditions for resolve error, then skip current gateway
	if resolveErr != nil {
		status.SetResolveErrorForPolicyAncestors(&policy.Status,
			parentGateways,
			t.GatewayControllerName,
			policy.Generation,
			resolveErr,
		)

		return
	}

	if err := t.translateSecurityPolicyForGateway(policy, targetedGateway, currTarget, resources, xdsIR); err != nil {
		status.SetTranslationErrorForPolicyAncestors(&policy.Status,
			parentGateways,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(err),
		)
	}

	// Set Accepted condition if it is unset
	status.SetAcceptedForPolicyAncestors(&policy.Status, parentGateways, t.GatewayControllerName)

	// Check if this policy is overridden by other policies targeting at route and listener levels
	overriddenTargetsMessage := getOverriddenTargetsMessage(
		gatewayMap[gatewayNN], gatewayRouteMap[gatewayNN.String()], currTarget.SectionName)
	if overriddenTargetsMessage != "" {
		status.SetConditionForPolicyAncestors(&policy.Status,
			parentGateways,
			t.GatewayControllerName,
			egv1a1.PolicyConditionOverridden,
			metav1.ConditionTrue,
			egv1a1.PolicyReasonOverridden,
			"This policy is being overridden by other securityPolicies for "+overriddenTargetsMessage,
			policy.Generation,
		)
	}
}

func getOverriddenTargetsMessage(
	targetContext *policyGatewayTargetContext,
	listenerRouteMap map[string]sets.Set[string],
	sectionName *gwapiv1.SectionName,
) string {
	var listeners, routes []string
	if sectionName == nil {
		if targetContext != nil {
			listeners = targetContext.attachedToListeners.UnsortedList()
		}
		for _, routeSet := range listenerRouteMap {
			routes = append(routes, routeSet.UnsortedList()...)
		}
	} else if listenerRouteMap != nil {
		if routeSet, ok := listenerRouteMap[string(*sectionName)]; ok {
			routes = routeSet.UnsortedList()
		}
		if routeSet, ok := listenerRouteMap[""]; ok {
			routes = append(routes, routeSet.UnsortedList()...)
		}
	}
	if len(listeners) > 0 {
		sort.Strings(listeners)
		if len(routes) > 0 {
			sort.Strings(routes)
			return fmt.Sprintf("these listeners: %v and these routes: %v", listeners, routes)
		} else {
			return fmt.Sprintf("these listeners: %v", listeners)
		}
	} else if len(routes) > 0 {
		sort.Strings(routes)
		return fmt.Sprintf("these routes: %v", routes)
	}
	return ""
}

// validateSecurityPolicy validates the SecurityPolicy.
// It checks some constraints that are not covered by the CRD schema validation.
func validateSecurityPolicy(p *egv1a1.SecurityPolicy) error {
	apiKeyAuth := p.Spec.APIKeyAuth
	if apiKeyAuth != nil {
		if err := validateAPIKeyAuth(apiKeyAuth); err != nil {
			return err
		}
	}

	oidc := p.Spec.OIDC
	jwt := p.Spec.JWT
	if oidc != nil && oidc.PassThroughAuthHeader != nil && *oidc.PassThroughAuthHeader {
		if jwt == nil {
			return errors.New("the OIDC.PassThroughAuthHeader setting must be used in conjunction with JWT settings")
		}

		hasValidJwtExtractor := false
		for _, provider := range jwt.Providers {
			// When ExtractFrom is not specified it falls back to looking at the "Authorization: Bearer ..." header
			if provider.ExtractFrom == nil || len(provider.ExtractFrom.Headers) > 0 {
				hasValidJwtExtractor = true
				break
			}
		}
		if !hasValidJwtExtractor {
			return errors.New("the OIDC.PassThroughAuthHeader setting must be used in conjunction with a JWT provider that is configured to read from a header")
		}
	}

	return nil
}

func validateAPIKeyAuth(apiKeyAuth *egv1a1.APIKeyAuth) error {
	for _, keySource := range apiKeyAuth.ExtractFrom {
		// only one of headers, params or cookies is supposed to be specified.
		if len(keySource.Headers) > 0 && len(keySource.Params) > 0 ||
			len(keySource.Headers) > 0 && len(keySource.Cookies) > 0 ||
			len(keySource.Params) > 0 && len(keySource.Cookies) > 0 {
			return errors.New("only one of headers, params or cookies must be specified")
		}
	}
	return nil
}

func resolveSecurityPolicyGatewayTargetRef(
	policy *egv1a1.SecurityPolicy,
	target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName,
	gateways map[types.NamespacedName]*policyGatewayTargetContext,
) (*GatewayContext, *status.PolicyResolveError) {
	// Find the Gateway
	key := types.NamespacedName{
		Name:      string(target.Name),
		Namespace: policy.Namespace,
	}
	gateway, ok := gateways[key]

	// Gateway not found
	// It's not an error if the gateway is not found because the SecurityPolicy
	// may be reconciled by multiple controllers, and the gateway may not be managed
	// by this controller.
	if !ok {
		return nil, nil
	}

	if target.SectionName == nil {
		// Check if another policy targeting the same Gateway exists
		if gateway.attached {
			message := fmt.Sprintf("Unable to target Gateway %s, another SecurityPolicy has already attached to it",
				string(target.Name))

			return gateway.GatewayContext, &status.PolicyResolveError{
				Reason:  gwapiv1a2.PolicyReasonConflicted,
				Message: message,
			}
		}
		gateway.attached = true
	} else {
		listenerName := string(*target.SectionName)
		if gateway.attachedToListeners.Has(listenerName) {
			message := fmt.Sprintf("Unable to target Listener %s/%s, another SecurityPolicy has already attached to it",
				string(target.Name), listenerName)

			return gateway.GatewayContext, &status.PolicyResolveError{
				Reason:  gwapiv1a2.PolicyReasonConflicted,
				Message: message,
			}
		}
		gateway.attachedToListeners.Insert(listenerName)
	}

	gateways[key] = gateway

	return gateway.GatewayContext, nil
}

func resolveSecurityPolicyRouteTargetRef(
	policy *egv1a1.SecurityPolicy,
	target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName,
	routes map[policyTargetRouteKey]*policyRouteTargetContext,
) (RouteContext, *status.PolicyResolveError) {
	// Check if the route exists
	key := policyTargetRouteKey{
		Kind:      string(target.Kind),
		Name:      string(target.Name),
		Namespace: policy.Namespace,
	}
	route, ok := routes[key]

	// Route not found
	// It's not an error if the gateway is not found because the SecurityPolicy
	// may be reconciled by multiple controllers, and the gateway may not be managed
	// by this controller.
	if !ok {
		return nil, nil
	}

	// Check if another policy targeting the same xRoute exists
	if route.attached {
		message := fmt.Sprintf("Unable to target %s %s, another SecurityPolicy has already attached to it",
			string(target.Kind), string(target.Name))

		return route.RouteContext, &status.PolicyResolveError{
			Reason:  gwapiv1a2.PolicyReasonConflicted,
			Message: message,
		}
	}

	// Set context and save
	route.attached = true
	routes[key] = route

	return route.RouteContext, nil
}

func (t *Translator) translateSecurityPolicyForRoute(
	policy *egv1a1.SecurityPolicy, route RouteContext,
	resources *resource.Resources, xdsIR resource.XdsIRMap,
) error {
	// Build IR
	var (
		cors          *ir.CORS
		apiKeyAuth    *ir.APIKeyAuth
		basicAuth     *ir.BasicAuth
		authorization *ir.Authorization
		err, errs     error
	)

	if policy.Spec.CORS != nil {
		cors = t.buildCORS(policy.Spec.CORS)
	}

	if policy.Spec.BasicAuth != nil {
		if basicAuth, err = t.buildBasicAuth(
			policy,
			resources); err != nil {
			err = perr.WithMessage(err, "BasicAuth")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.APIKeyAuth != nil {
		if apiKeyAuth, err = t.buildAPIKeyAuth(
			policy,
			resources); err != nil {
			err = perr.WithMessage(err, "APIKeyAuth")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.Authorization != nil {
		if authorization, err = t.buildAuthorization(policy); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	// Apply IR to all relevant routes
	prefix := irRoutePrefix(route)
	parentRefs := GetParentReferences(route)
	for _, p := range parentRefs {
		parentRefCtx := GetRouteParentContext(route, p)
		gtwCtx := parentRefCtx.GetGateway()
		if gtwCtx == nil {
			continue
		}

		var extAuth *ir.ExtAuth
		if policy.Spec.ExtAuth != nil {
			if extAuth, err = t.buildExtAuth(
				policy,
				resources,
				gtwCtx.envoyProxy); err != nil {
				err = perr.WithMessage(err, "ExtAuth")
				errs = errors.Join(errs, err)
			}
		}

		var oidc *ir.OIDC
		if policy.Spec.OIDC != nil {
			if oidc, err = t.buildOIDC(
				policy,
				resources,
				gtwCtx.envoyProxy); err != nil {
				err = perr.WithMessage(err, "OIDC")
				errs = errors.Join(errs, err)
			}
		}

		var jwt *ir.JWT
		if policy.Spec.JWT != nil {
			if jwt, err = t.buildJWT(
				policy,
				resources,
				gtwCtx.envoyProxy); err != nil {
				err = perr.WithMessage(err, "JWT")
				errs = errors.Join(errs, err)
			}
		}

		irKey := t.getIRKey(gtwCtx.Gateway)
		for _, listener := range parentRefCtx.listeners {
			irListener := xdsIR[irKey].GetHTTPListener(irListenerName(listener))
			if irListener != nil {
				for _, r := range irListener.Routes {
					if strings.HasPrefix(r.Name, prefix) {
						r.Security = &ir.SecurityFeatures{
							CORS:          cors,
							JWT:           jwt,
							OIDC:          oidc,
							APIKeyAuth:    apiKeyAuth,
							BasicAuth:     basicAuth,
							ExtAuth:       extAuth,
							Authorization: authorization,
						}
						if errs != nil {
							// Return a 500 direct response to avoid unauthorized access
							r.DirectResponse = &ir.CustomResponse{
								StatusCode: ptr.To(uint32(500)),
							}
						}
					}
				}
			}
		}
	}
	return errs
}

func (t *Translator) translateSecurityPolicyForGateway(
	policy *egv1a1.SecurityPolicy,
	gateway *GatewayContext,
	target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName,
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
) error {
	// Build IR
	var (
		cors          *ir.CORS
		jwt           *ir.JWT
		oidc          *ir.OIDC
		apiKeyAuth    *ir.APIKeyAuth
		basicAuth     *ir.BasicAuth
		extAuth       *ir.ExtAuth
		authorization *ir.Authorization
		err, errs     error
	)

	if policy.Spec.CORS != nil {
		cors = t.buildCORS(policy.Spec.CORS)
	}

	if policy.Spec.JWT != nil {
		if jwt, err = t.buildJWT(
			policy,
			resources,
			gateway.envoyProxy); err != nil {
			err = perr.WithMessage(err, "JWT")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.OIDC != nil {
		if oidc, err = t.buildOIDC(
			policy,
			resources,
			gateway.envoyProxy); err != nil {
			err = perr.WithMessage(err, "OIDC")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.BasicAuth != nil {
		if basicAuth, err = t.buildBasicAuth(
			policy,
			resources); err != nil {
			err = perr.WithMessage(err, "BasicAuth")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.APIKeyAuth != nil {
		if apiKeyAuth, err = t.buildAPIKeyAuth(
			policy,
			resources); err != nil {
			err = perr.WithMessage(err, "APIKeyAuth")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.ExtAuth != nil {
		if extAuth, err = t.buildExtAuth(
			policy,
			resources,
			gateway.envoyProxy); err != nil {
			err = perr.WithMessage(err, "ExtAuth")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.Authorization != nil {
		if authorization, err = t.buildAuthorization(policy); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	// Apply IR to all the routes within the specific Gateway that originated
	// from the gateway to which this security policy was attached.
	// If the feature is already set, then skip it, since it must have be
	// set by a policy attaching to the route
	//
	// Note: there are multiple features in a security policy, even if some of them
	// are invalid, we still want to apply the valid ones.
	irKey := t.getIRKey(gateway.Gateway)
	// Should exist since we've validated this
	x := xdsIR[irKey]

	policyTarget := irStringKey(policy.Namespace, string(target.Name))
	for _, h := range x.HTTP {
		// A HTTPListener name has the format namespace/gatewayName/listenerName
		gatewayNameEnd := strings.LastIndex(h.Name, "/")
		gatewayName := h.Name[0:gatewayNameEnd]
		if t.MergeGateways && gatewayName != policyTarget {
			continue
		}
		// If specified the sectionName must match the listenerName part of the HTTPListener name
		if target.SectionName != nil && string(*target.SectionName) != h.Name[gatewayNameEnd+1:] {
			continue
		}
		// A Policy targeting the most specific scope(xRoute) wins over a policy
		// targeting a lesser specific scope(Gateway).
		for _, r := range h.Routes {
			// If any of the features are already set, it means that a more specific
			// policy(targeting xRoute) has already set it, so we skip it.
			if r.Security != nil {
				continue
			}
			r.Security = &ir.SecurityFeatures{
				CORS:          cors,
				JWT:           jwt,
				OIDC:          oidc,
				APIKeyAuth:    apiKeyAuth,
				BasicAuth:     basicAuth,
				ExtAuth:       extAuth,
				Authorization: authorization,
			}
			if errs != nil {
				// Return a 500 direct response to avoid unauthorized access
				r.DirectResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(500)),
				}
			}
		}
	}
	return errs
}

func (t *Translator) buildCORS(cors *egv1a1.CORS) *ir.CORS {
	var allowOrigins []*ir.StringMatch

	for _, origin := range cors.AllowOrigins {
		if containsWildcard(string(origin)) {
			regexStr := wildcard2regex(string(origin))
			allowOrigins = append(allowOrigins, &ir.StringMatch{
				SafeRegex: &regexStr,
			})
		} else {
			allowOrigins = append(allowOrigins, &ir.StringMatch{
				Exact: (*string)(&origin),
			})
		}
	}

	return &ir.CORS{
		AllowOrigins:     allowOrigins,
		AllowMethods:     cors.AllowMethods,
		AllowHeaders:     cors.AllowHeaders,
		ExposeHeaders:    cors.ExposeHeaders,
		MaxAge:           cors.MaxAge,
		AllowCredentials: cors.AllowCredentials != nil && *cors.AllowCredentials,
	}
}

func containsWildcard(s string) bool {
	return strings.ContainsAny(s, "*")
}

func wildcard2regex(wildcard string) string {
	regexStr := strings.ReplaceAll(wildcard, ".", "\\.")
	regexStr = strings.ReplaceAll(regexStr, "*", ".*")
	return regexStr
}

func (t *Translator) buildJWT(
	policy *egv1a1.SecurityPolicy,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.JWT, error) {
	if err := validateJWTProvider(policy.Spec.JWT.Providers); err != nil {
		return nil, err
	}

	var providers []ir.JWTProvider
	for i, p := range policy.Spec.JWT.Providers {
		provider := ir.JWTProvider{
			Name:           p.Name,
			Issuer:         p.Issuer,
			Audiences:      p.Audiences,
			ClaimToHeaders: p.ClaimToHeaders,
			RecomputeRoute: p.RecomputeRoute,
			ExtractFrom:    p.ExtractFrom,
		}
		if p.RemoteJWKS != nil {
			remoteJWKS, err := t.buildRemoteJWKS(policy, p.RemoteJWKS, i, resources, envoyProxy)
			if err != nil {
				return nil, err
			}
			provider.RemoteJWKS = remoteJWKS
		} else {
			localJWKS, err := t.buildLocalJWKS(policy, p.LocalJWKS, resources)
			if err != nil {
				return nil, err
			}
			provider.LocalJWKS = &localJWKS
		}
		providers = append(providers, provider)
	}

	return &ir.JWT{
		AllowMissing: ptr.Deref(policy.Spec.JWT.Optional, false),
		Providers:    providers,
	}, nil
}

func validateJWTProvider(providers []egv1a1.JWTProvider) error {
	var errs []error

	var names []string
	for _, provider := range providers {
		if len(provider.Name) == 0 {
			errs = append(errs, errors.New("jwt provider cannot be an empty string"))
		}

		if len(provider.Issuer) != 0 {
			switch {
			// Issuer follows StringOrURI format based on https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.1.
			// Hence, when it contains ':', it MUST be a valid URI.
			case strings.Contains(provider.Issuer, ":"):
				if _, err := url.ParseRequestURI(provider.Issuer); err != nil {
					errs = append(errs, fmt.Errorf("invalid issuer; when issuer contains ':' character, it MUST be a valid URI"))
				}
			// Adding reserved character for '@', to represent an email address.
			// Hence, when it contains '@', it MUST be a valid Email Address.
			case strings.Contains(provider.Issuer, "@"):
				if _, err := mail.ParseAddress(provider.Issuer); err != nil {
					errs = append(errs, fmt.Errorf("invalid issuer; when issuer contains '@' character, it MUST be a valid Email Address format: %w", err))
				}
			}
		}

		if (provider.RemoteJWKS == nil && provider.LocalJWKS == nil) ||
			(provider.RemoteJWKS != nil && provider.LocalJWKS != nil) {
			errs = append(errs, fmt.Errorf(
				"either remoteJWKS or localJWKS must be specified for jwt provider: %s", provider.Name))
		}

		if provider.RemoteJWKS != nil {
			if len(provider.RemoteJWKS.URI) == 0 {
				errs = append(errs, fmt.Errorf("uri must be set for remote JWKS provider: %s", provider.Name))
			} else if _, err := url.ParseRequestURI(provider.RemoteJWKS.URI); err != nil {
				errs = append(errs, fmt.Errorf("invalid remote JWKS URI: %w", err))
			}
		}

		if provider.LocalJWKS != nil {
			localJWKS := provider.LocalJWKS
			if localJWKS.Type == nil || *localJWKS.Type == egv1a1.LocalJWKSTypeInline {
				if localJWKS.Inline == nil {
					errs = append(errs, fmt.Errorf("inline JWKS must be set for local JWKS provider: %s if type is Inline", provider.Name))
				}
			} else if localJWKS.ValueRef == nil {
				errs = append(errs, fmt.Errorf("valueRef must be set for local JWKS provider: %s if type is ValueRef", provider.Name))
			}
		}

		if len(errs) == 0 {
			if strErrs := validation.IsQualifiedName(provider.Name); len(strErrs) != 0 {
				for _, strErr := range strErrs {
					errs = append(errs, errors.New(strErr))
				}
			}
			// Ensure uniqueness among provider names.
			if names == nil {
				names = append(names, provider.Name)
			} else {
				for _, name := range names {
					if name == provider.Name {
						errs = append(errs, fmt.Errorf("provider name %s must be unique", provider.Name))
					} else {
						names = append(names, provider.Name)
					}
				}
			}
		}

		for _, claimToHeader := range provider.ClaimToHeaders {
			switch {
			case len(claimToHeader.Header) == 0:
				errs = append(errs, fmt.Errorf("header must be set for claimToHeader provider: %s", claimToHeader.Header))
			case len(claimToHeader.Claim) == 0:
				errs = append(errs, fmt.Errorf("claim must be set for claimToHeader provider: %s", claimToHeader.Claim))
			}
		}
	}

	return errors.Join(errs...)
}

func (t *Translator) buildRemoteJWKS(
	policy *egv1a1.SecurityPolicy,
	remoteJWKS *egv1a1.RemoteJWKS,
	index int,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.RemoteJWKS, error) {
	var (
		protocol ir.AppProtocol
		rd       *ir.RouteDestination
		traffic  *ir.TrafficFeatures
		err      error
	)

	u, err := url.Parse(remoteJWKS.URI)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "https" {
		protocol = ir.HTTPS
	} else {
		protocol = ir.HTTP
	}

	if len(remoteJWKS.BackendRefs) > 0 {
		if rd, err = t.translateExtServiceBackendRefs(
			policy, remoteJWKS.BackendRefs, protocol, resources, envoyProxy, "jwt", index); err != nil {
			return nil, err
		}
	}

	if remoteJWKS.BackendSettings != nil {
		if traffic, err = translateTrafficFeatures(remoteJWKS.BackendSettings); err != nil {
			return nil, err
		}
	}

	return &ir.RemoteJWKS{
		Destination: rd,
		Traffic:     traffic,
		URI:         remoteJWKS.URI,
	}, nil
}

func (t *Translator) buildLocalJWKS(
	policy *egv1a1.SecurityPolicy,
	localJWKS *egv1a1.LocalJWKS,
	resources *resource.Resources,
) (string, error) {
	jwksType := egv1a1.LocalJWKSTypeInline
	if localJWKS.Type != nil {
		jwksType = *localJWKS.Type
	}

	if jwksType == egv1a1.LocalJWKSTypeValueRef {
		cm := resources.GetConfigMap(policy.Namespace, string(localJWKS.ValueRef.Name))
		if cm == nil {
			return "", fmt.Errorf("local JWKS ConfigMap %s/%s not found", policy.Namespace, localJWKS.ValueRef.Name)
		}

		jwksBytes, ok := cm.Data["jwks"]
		if ok {
			return jwksBytes, nil
		}
		if len(cm.Data) > 0 {
			// Fallback to the first entry in the ConfigMap data if "jwks" key is not present
			for _, v := range cm.Data {
				return v, nil
			}
		}

		return "", fmt.Errorf(
			"JWKS data not found in ConfigMap %s/%s, no 'jwks' key and no other data found",
			cm.Namespace, cm.Name)
	}

	return *localJWKS.Inline, nil
}

func (t *Translator) buildOIDC(
	policy *egv1a1.SecurityPolicy,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.OIDC, error) {
	var (
		oidc                  = policy.Spec.OIDC
		provider              *ir.OIDCProvider
		clientSecret          *corev1.Secret
		redirectURL           = defaultRedirectURL
		redirectPath          = defaultRedirectPath
		logoutPath            = defaultLogoutPath
		forwardAccessToken    = defaultForwardAccessToken
		refreshToken          = defaultRefreshToken
		passThroughAuthHeader = defaultPassThroughAuthHeader
		err                   error
	)

	if provider, err = t.buildOIDCProvider(policy, resources, envoyProxy); err != nil {
		return nil, err
	}

	from := crossNamespaceFrom{
		group:     egv1a1.GroupName,
		kind:      resource.KindSecurityPolicy,
		namespace: policy.Namespace,
	}

	if clientSecret, err = t.validateSecretRef(
		false, from, oidc.ClientSecret, resources); err != nil {
		return nil, err
	}

	clientSecretBytes, ok := clientSecret.Data[egv1a1.OIDCClientSecretKey]
	if !ok || len(clientSecretBytes) == 0 {
		return nil, fmt.Errorf(
			"client secret not found in secret %s/%s",
			clientSecret.Namespace, clientSecret.Name)
	}

	scopes := appendOpenidScopeIfNotExist(oidc.Scopes)

	if oidc.RedirectURL != nil {
		path, err := extractRedirectPath(*oidc.RedirectURL)
		if err != nil {
			return nil, err
		}
		redirectURL = *oidc.RedirectURL
		redirectPath = path
	}
	if oidc.LogoutPath != nil {
		logoutPath = *oidc.LogoutPath
	}
	if oidc.ForwardAccessToken != nil {
		forwardAccessToken = *oidc.ForwardAccessToken
	}
	if oidc.RefreshToken != nil {
		refreshToken = *oidc.RefreshToken
	}

	if oidc.PassThroughAuthHeader != nil {
		passThroughAuthHeader = *oidc.PassThroughAuthHeader
	}

	// Generate a unique cookie suffix for oauth filters.
	// This is to avoid cookie name collision when multiple security policies are applied
	// to the same route.
	suffix := utils.Digest32(string(policy.UID))

	// Get the HMAC secret.
	// HMAC secret is generated by the CertGen job and stored in a secret
	// We need to rotate the HMAC secret in the future, probably the same
	// way we rotate the certs generated by the CertGen job.
	hmacSecret := resources.GetSecret(t.ControllerNamespace, oidcHMACSecretName)
	if hmacSecret == nil {
		return nil, fmt.Errorf("HMAC secret %s/%s not found", t.ControllerNamespace, oidcHMACSecretName)
	}
	hmacData, ok := hmacSecret.Data[oidcHMACSecretKey]
	if !ok || len(hmacData) == 0 {
		return nil, fmt.Errorf(
			"HMAC secret not found in secret %s/%s", t.ControllerNamespace, oidcHMACSecretName)
	}

	return &ir.OIDC{
		Name:                   irConfigName(policy),
		Provider:               *provider,
		ClientID:               oidc.ClientID,
		ClientSecret:           clientSecretBytes,
		Scopes:                 scopes,
		Resources:              oidc.Resources,
		RedirectURL:            redirectURL,
		RedirectPath:           redirectPath,
		LogoutPath:             logoutPath,
		ForwardAccessToken:     forwardAccessToken,
		DefaultTokenTTL:        oidc.DefaultTokenTTL,
		RefreshToken:           refreshToken,
		DefaultRefreshTokenTTL: oidc.DefaultRefreshTokenTTL,
		CookieSuffix:           suffix,
		CookieNameOverrides:    policy.Spec.OIDC.CookieNames,
		CookieDomain:           policy.Spec.OIDC.CookieDomain,
		CookieConfig:           policy.Spec.OIDC.CookieConfig,
		HMACSecret:             hmacData,
		PassThroughAuthHeader:  passThroughAuthHeader,
		DenyRedirect:           oidc.DenyRedirect,
	}, nil
}

func (t *Translator) buildOIDCProvider(policy *egv1a1.SecurityPolicy, resources *resource.Resources, envoyProxy *egv1a1.EnvoyProxy) (*ir.OIDCProvider, error) {
	var (
		provider              = policy.Spec.OIDC.Provider
		tokenEndpoint         string
		authorizationEndpoint string
		protocol              ir.AppProtocol
		rd                    *ir.RouteDestination
		traffic               *ir.TrafficFeatures
		providerTLS           *ir.TLSUpstreamConfig
		err                   error
	)

	var u *url.URL
	if provider.TokenEndpoint != nil {
		u, err = url.Parse(*provider.TokenEndpoint)
	} else {
		u, err = url.Parse(provider.Issuer)
	}

	if err != nil {
		return nil, err
	}

	if u.Scheme == "https" {
		protocol = ir.HTTPS
	} else {
		protocol = ir.HTTP
	}

	if len(provider.BackendRefs) > 0 {
		if rd, err = t.translateExtServiceBackendRefs(policy, provider.BackendRefs, protocol, resources, envoyProxy, "oidc", 0); err != nil {
			return nil, err
		}
	}

	if rd != nil {
		for _, st := range rd.Settings {
			if st.TLS != nil {
				providerTLS = st.TLS
				break
			}
		}
	}

	// Discover the token and authorization endpoints from the issuer's well-known url if not explicitly specified.
	// EG assumes that the issuer url uses the same protocol and CA as the token endpoint.
	// If we need to support different protocols or CAs, we need to add more fields to the OIDCProvider CRD.
	if provider.TokenEndpoint == nil || provider.AuthorizationEndpoint == nil {
		tokenEndpoint, authorizationEndpoint, err = fetchEndpointsFromIssuer(provider.Issuer, providerTLS)
		if err != nil {
			return nil, fmt.Errorf("error fetching endpoints from issuer: %w", err)
		}
	} else {
		tokenEndpoint = *provider.TokenEndpoint
		authorizationEndpoint = *provider.AuthorizationEndpoint
	}

	if err = validateTokenEndpoint(tokenEndpoint); err != nil {
		return nil, err
	}

	if traffic, err = translateTrafficFeatures(provider.BackendSettings); err != nil {
		return nil, err
	}

	return &ir.OIDCProvider{
		Destination:           rd,
		Traffic:               traffic,
		AuthorizationEndpoint: authorizationEndpoint,
		TokenEndpoint:         tokenEndpoint,
	}, nil
}

func extractRedirectPath(redirectURL string) (string, error) {
	schemeDelimiter := strings.Index(redirectURL, "://")
	if schemeDelimiter <= 0 {
		return "", fmt.Errorf("invalid redirect URL %s", redirectURL)
	}
	scheme := redirectURL[:schemeDelimiter]
	if scheme != "http" && scheme != "https" && scheme != "%REQ(x-forwarded-proto)%" {
		return "", fmt.Errorf("invalid redirect URL %s", redirectURL)
	}
	hostDelimiter := strings.Index(redirectURL[schemeDelimiter+3:], "/")
	if hostDelimiter <= 0 {
		return "", fmt.Errorf("invalid redirect URL %s", redirectURL)
	}
	path := redirectURL[schemeDelimiter+3+hostDelimiter:]
	if path == "/" {
		return "", fmt.Errorf("invalid redirect URL %s", redirectURL)
	}
	return path, nil
}

// appendOpenidScopeIfNotExist appends the openid scope to the provided scopes
// if it is not already present.
// `openid` is a required scope for OIDC.
// see https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
func appendOpenidScopeIfNotExist(scopes []string) []string {
	const authScopeOpenID = "openid"

	hasOpenIDScope := false
	for _, scope := range scopes {
		if scope == authScopeOpenID {
			hasOpenIDScope = true
		}
	}
	if !hasOpenIDScope {
		scopes = append(scopes, authScopeOpenID)
	}
	return scopes
}

type OpenIDConfig struct {
	TokenEndpoint         string `json:"token_endpoint"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
}

func fetchEndpointsFromIssuer(issuerURL string, providerTLS *ir.TLSUpstreamConfig) (string, string, error) {
	var (
		tlsConfig *tls.Config
		err       error
	)

	if providerTLS != nil {
		if tlsConfig, err = providerTLS.ToTLSConfig(); err != nil {
			return "", "", err
		}
	}

	client := &http.Client{}
	if tlsConfig != nil {
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	// Parse the OpenID configuration response
	var config OpenIDConfig
	if err = backoff.Retry(func() error {
		resp, err := client.Get(fmt.Sprintf("%s/.well-known/openid-configuration", issuerURL))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if err = json.NewDecoder(resp.Body).Decode(&config); err != nil {
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff(backoff.WithMaxElapsedTime(5*time.Second))); err != nil {
		return "", "", err
	}

	return config.TokenEndpoint, config.AuthorizationEndpoint, nil
}

// validateTokenEndpoint validates the token endpoint URL
func validateTokenEndpoint(tokenEndpoint string) error {
	parsedURL, err := url.Parse(tokenEndpoint)
	if err != nil {
		return fmt.Errorf("error parsing token endpoint URL: %w", err)
	}

	if ip, err := netip.ParseAddr(parsedURL.Hostname()); err == nil {
		if ip.Unmap().Is4() {
			return fmt.Errorf("token endpoint URL must be a domain name: %s", tokenEndpoint)
		}
	}

	if parsedURL.Port() != "" {
		_, err = strconv.Atoi(parsedURL.Port())
		if err != nil {
			return fmt.Errorf("error parsing token endpoint URL port: %w", err)
		}
	}
	return nil
}

func (t *Translator) buildAPIKeyAuth(
	policy *egv1a1.SecurityPolicy,
	resources *resource.Resources,
) (*ir.APIKeyAuth, error) {
	from := crossNamespaceFrom{
		group:     egv1a1.GroupName,
		kind:      resource.KindSecurityPolicy,
		namespace: policy.Namespace,
	}

	credentials := make(map[string]ir.PrivateBytes)
	seenKeys := make(sets.Set[string])

	for _, ref := range policy.Spec.APIKeyAuth.CredentialRefs {
		credentialsSecret, err := t.validateSecretRef(
			false, from, ref, resources)
		if err != nil {
			return nil, err
		}
		for clientid, key := range credentialsSecret.Data {
			if _, ok := credentials[clientid]; ok {
				continue
			}

			keyString := string(key)
			if seenKeys.Has(keyString) {
				return nil, errors.New("duplicated API key")
			}

			seenKeys.Insert(keyString)
			credentials[clientid] = key
		}
	}

	extractFrom := make([]*ir.ExtractFrom, 0, len(policy.Spec.APIKeyAuth.ExtractFrom))
	for _, e := range policy.Spec.APIKeyAuth.ExtractFrom {
		extractFrom = append(extractFrom, &ir.ExtractFrom{
			Headers: e.Headers,
			Cookies: e.Cookies,
			Params:  e.Params,
		})
	}

	return &ir.APIKeyAuth{
		Credentials: credentials,
		ExtractFrom: extractFrom,
	}, nil
}

func (t *Translator) buildBasicAuth(
	policy *egv1a1.SecurityPolicy,
	resources *resource.Resources,
) (*ir.BasicAuth, error) {
	var (
		basicAuth   = policy.Spec.BasicAuth
		usersSecret *corev1.Secret
		err         error
	)

	from := crossNamespaceFrom{
		group:     egv1a1.GroupName,
		kind:      resource.KindSecurityPolicy,
		namespace: policy.Namespace,
	}
	if usersSecret, err = t.validateSecretRef(
		false, from, basicAuth.Users, resources); err != nil {
		return nil, err
	}

	usersSecretBytes, ok := usersSecret.Data[egv1a1.BasicAuthUsersSecretKey]
	if !ok || len(usersSecretBytes) == 0 {
		return nil, fmt.Errorf(
			"users secret not found in secret %s/%s",
			usersSecret.Namespace, usersSecret.Name)
	}

	return &ir.BasicAuth{
		Name:                  irConfigName(policy),
		Users:                 usersSecretBytes,
		ForwardUsernameHeader: basicAuth.ForwardUsernameHeader,
	}, nil
}

func (t *Translator) buildExtAuth(
	policy *egv1a1.SecurityPolicy,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.ExtAuth, error) {
	var (
		http            = policy.Spec.ExtAuth.HTTP
		grpc            = policy.Spec.ExtAuth.GRPC
		backendRefs     []egv1a1.BackendRef
		backendSettings *egv1a1.ClusterSettings
		protocol        ir.AppProtocol
		rd              *ir.RouteDestination
		authority       string
		err             error
		traffic         *ir.TrafficFeatures
	)

	// These are sanity checks, they should never happen because the API server
	// should have caught them
	if http == nil && grpc == nil {
		return nil, errors.New("one of grpc or http must be specified")
	} else if http != nil && grpc != nil {
		return nil, errors.New("only one of grpc or http can be specified")
	}

	switch {
	case http != nil:
		protocol = ir.HTTP
		backendSettings = http.BackendSettings
		switch {
		case len(http.BackendRefs) > 0:
			backendRefs = http.BackendRefs
		case http.BackendRef != nil:
			backendRefs = []egv1a1.BackendRef{
				{
					BackendObjectReference: *http.BackendRef,
				},
			}
		default:
			// This is a sanity check, it should never happen because the API server should have caught it
			return nil, errors.New("http backend refs must be specified")
		}
	case grpc != nil:
		protocol = ir.GRPC
		backendSettings = grpc.BackendSettings
		switch {
		case len(grpc.BackendRefs) > 0:
			backendRefs = grpc.BackendRefs
		case grpc.BackendRef != nil:
			backendRefs = []egv1a1.BackendRef{
				{
					BackendObjectReference: *grpc.BackendRef,
				},
			}
		default:
			// This is a sanity check, it should never happen because the API server should have caught it
			return nil, errors.New("grpc backend refs must be specified")
		}
	}

	if rd, err = t.translateExtServiceBackendRefs(policy, backendRefs, protocol, resources, envoyProxy, "extauth", 0); err != nil {
		return nil, err
	}

	for _, backendRef := range backendRefs {
		// Authority is the calculated hostname that will be used as the Authority header.
		// If there are multiple backend referenced, simply use the first one - there are no good answers here.
		// When translated to XDS, the authority is used on the filter level not on the cluster level.
		// There's no way to translate to XDS and use a different authority for each backendref
		if authority == "" {
			authority = backendRefAuthority(resources, &backendRef.BackendObjectReference, policy)
		}
	}

	if traffic, err = translateTrafficFeatures(backendSettings); err != nil {
		return nil, err
	}
	extAuth := &ir.ExtAuth{
		Name:             irConfigName(policy),
		HeadersToExtAuth: policy.Spec.ExtAuth.HeadersToExtAuth,
		FailOpen:         policy.Spec.ExtAuth.FailOpen,
		Traffic:          traffic,
		RecomputeRoute:   policy.Spec.ExtAuth.RecomputeRoute,
	}

	if http != nil {
		extAuth.HTTP = &ir.HTTPExtAuthService{
			Destination:      *rd,
			Authority:        authority,
			Path:             ptr.Deref(http.Path, ""),
			HeadersToBackend: http.HeadersToBackend,
		}
	} else {
		extAuth.GRPC = &ir.GRPCExtAuthService{
			Destination: *rd,
			Authority:   authority,
		}
	}

	if policy.Spec.ExtAuth.BodyToExtAuth != nil {
		extAuth.BodyToExtAuth = &ir.BodyToExtAuth{
			MaxRequestBytes: policy.Spec.ExtAuth.BodyToExtAuth.MaxRequestBytes,
		}
	}

	return extAuth, nil
}

func backendRefAuthority(resources *resource.Resources, backendRef *gwapiv1.BackendObjectReference, policy *egv1a1.SecurityPolicy) string {
	if backendRef == nil {
		return ""
	}

	backendNamespace := NamespaceDerefOr(backendRef.Namespace, policy.Namespace)
	backendKind := KindDerefOr(backendRef.Kind, resource.KindService)
	if backendKind == resource.KindBackend {
		backend := resources.GetBackend(backendNamespace, string(backendRef.Name))
		if backend != nil {
			// TODO: exists multi FQDN endpoints?
			for _, ep := range backend.Spec.Endpoints {
				if ep.FQDN != nil {
					return net.JoinHostPort(ep.FQDN.Hostname, strconv.Itoa(int(ep.FQDN.Port)))
				}
			}
		}
	}

	// Port is mandatory for Kubernetes services
	if backendKind == resource.KindService {
		return net.JoinHostPort(
			fmt.Sprintf("%s.%s", backendRef.Name, backendNamespace),
			strconv.Itoa(int(*backendRef.Port)),
		)
	}

	// Fallback to the backendRef name, normally it's a unix domain socket in this case
	return fmt.Sprintf("%s.%s", backendRef.Name, backendNamespace)
}

func (t *Translator) buildAuthorization(policy *egv1a1.SecurityPolicy) (*ir.Authorization, error) {
	var (
		authorization = policy.Spec.Authorization
		irAuth        = &ir.Authorization{}
		// The default action is Deny if not specified
		defaultAction = egv1a1.AuthorizationActionDeny
	)

	if authorization.DefaultAction != nil {
		defaultAction = *authorization.DefaultAction
	}
	irAuth.DefaultAction = defaultAction

	for i, rule := range authorization.Rules {
		irPrincipal := ir.Principal{}

		for _, cidr := range rule.Principal.ClientCIDRs {
			cidrMatch, err := parseCIDR(string(cidr))
			if err != nil {
				return nil, fmt.Errorf("unable to translate authorization rule: %w", err)
			}

			irPrincipal.ClientCIDRs = append(irPrincipal.ClientCIDRs, cidrMatch)
		}

		irPrincipal.JWT = rule.Principal.JWT
		irPrincipal.Headers = rule.Principal.Headers

		var name string
		if rule.Name != nil && *rule.Name != "" {
			name = *rule.Name
		} else {
			name = defaultAuthorizationRuleName(policy, i)
		}
		irAuth.Rules = append(irAuth.Rules, &ir.AuthorizationRule{
			Name:      name,
			Action:    rule.Action,
			Operation: rule.Operation,
			Principal: irPrincipal,
		})
	}

	return irAuth, nil
}

func defaultAuthorizationRuleName(policy *egv1a1.SecurityPolicy, index int) string {
	return fmt.Sprintf(
		"%s/authorization/rule/%s",
		irConfigName(policy),
		strconv.Itoa(index))
}
