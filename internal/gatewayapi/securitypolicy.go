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
	// SecurityPolicies are already sorted by the provider layer

	// First build a map out of the routes and gateways for faster lookup since users might have thousands of routes or more.
	// For gateways this probably isn't quite as necessary.
	routeMap := map[policyTargetRouteKey]*policyRouteTargetContext{}
	for _, route := range routes {
		key := policyTargetRouteKey{
			Kind:      string(GetRouteType(route)),
			Name:      route.GetName(),
			Namespace: route.GetNamespace(),
		}
		routeMap[key] = &policyRouteTargetContext{RouteContext: route, attachedToRouteRules: make(sets.Set[string])}
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
	// 1. First translate Policies targeting RouteRules
	// 2. Next translate Policies targeting xRoutes
	// 3. Then translate Policies targeting Listeners
	// 4. Finally, the policies targeting Gateways

	// Process the policies targeting RouteRules
	// Process the policies targeting RouteRules (HTTP + now TCP)
	for _, currPolicy := range securityPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, routes)
		for _, currTarget := range targetRefs {
			if currTarget.Kind != resource.KindGateway && currTarget.SectionName != nil {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy.DeepCopy()
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}
				t.processSecurityPolicyForRoute(resources, xdsIR,
					routeMap, gatewayRouteMap, policy, currTarget)
			}
		}
	}

	// Process the policies targeting whole xRoutes (HTTP + TCP)
	for _, currPolicy := range securityPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, routes)
		for _, currTarget := range targetRefs {
			if currTarget.Kind != resource.KindGateway && currTarget.SectionName == nil {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy.DeepCopy()
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}
				t.processSecurityPolicyForRoute(resources, xdsIR,
					routeMap, gatewayRouteMap, policy, currTarget)
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

	for _, policy := range res {
		// Truncate Ancestor list of longer than 16
		if len(policy.Status.Ancestors) > 16 {
			status.TruncatePolicyAncestors(&policy.Status, t.GatewayControllerName, policy.Generation)
		}
	}
	return res
}

func (t *Translator) processSecurityPolicyForRoute(
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
	routeMap map[policyTargetRouteKey]*policyRouteTargetContext,
	gatewayRouteMap map[string]map[string]sets.Set[string],
	policy *egv1a1.SecurityPolicy,
	currTarget gwapiv1a2.LocalPolicyTargetReferenceWithSectionName,
) {
	var (
		targetedRoute  RouteContext
		parentGateways []gwapiv1a2.ParentReference
		resolveErr     *status.PolicyResolveError
	)
	// Skip if the route is not found
	// It's not necessarily an error because the SecurityPolicy may be
	// reconciled by multiple controllers. And the other controller may
	// have the target route.
	isTCP := currTarget.Kind == resource.KindTCPRoute

	// Unified resolution (TCP + HTTP/other routes use same helper now).
	targetedRoute, resolveErr = resolveSecurityPolicyRouteTargetRef(policy, currTarget, routeMap)
	if targetedRoute == nil {
		return
	}

	// Find the parent Gateways for the route and add it to the
	// gatewayRouteMap, which will be used to check policy override.
	// The parent gateways are also used to set the status of the policy.
	// Gateway scope (e.g. when a Gateway-level policy is overridden by a
	// route- or rule-level policy). Previously we skipped populating this
	// structure for TCP routes which meant Gateway-level override
	// detection could not “see” TCP route attachments. We now treat TCP
	// and HTTP uniformly here so getOverriddenTargetsMessageForGateway
	// has a complete view.
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

			// SectionName (listener) may be nil (whole gateway parentRef)
			sectionName := ""
			if p.SectionName != nil {
				sectionName = string(*p.SectionName)
			}
			if _, ok := listenerRouteMap[sectionName]; !ok {
				listenerRouteMap[sectionName] = make(sets.Set[string])
			}
			// Track this route (namespaced) under the listener (or whole gateway key "").
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
		return
	}

	// Validation per protocol
	if isTCP {
		if err := validateSecurityPolicyForTCP(policy); err != nil {
			status.SetTranslationErrorForPolicyAncestors(&policy.Status,
				parentGateways,
				t.GatewayControllerName,
				policy.Generation,
				status.Error2ConditionMsg(fmt.Errorf("invalid SecurityPolicy for TCP route: %w", err)),
			)
			return
		}
	} else {
		if err := validateSecurityPolicy(policy); err != nil {
			status.SetTranslationErrorForPolicyAncestors(&policy.Status,
				parentGateways,
				t.GatewayControllerName,
				policy.Generation,
				status.Error2ConditionMsg(fmt.Errorf("invalid SecurityPolicy: %w", err)),
			)
			return
		}
	}

	// Translate
	if err := t.translateSecurityPolicyForRoute(policy, targetedRoute, currTarget, resources, xdsIR); err != nil {
		status.SetTranslationErrorForPolicyAncestors(&policy.Status,
			parentGateways,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(err),
		)
	}

	// Set Accepted condition if it is unset
	status.SetAcceptedForPolicyAncestors(&policy.Status, parentGateways, t.GatewayControllerName, policy.Generation)

	// Check if this policy is overridden by other policies targeting at route rule levels
	// Override condition (HTTP only; TCP has single rule semantics)
	if !isTCP {
		key := policyTargetRouteKey{
			Kind:      string(currTarget.Kind),
			Name:      string(currTarget.Name),
			Namespace: policy.Namespace,
		}
		overriddenTargetsMessage := getOverriddenTargetsMessageForRoute(routeMap[key], currTarget.SectionName)
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

	// Added: protocol-aware validation (mirrors route handling) + mixed-protocol whole-gateway guard
	isTCPListener := false
	hasHTTP := false
	hasTCP := false
	for _, l := range targetedGateway.Spec.Listeners {
		switch l.Protocol {
		case gwapiv1.TCPProtocolType:
			hasTCP = true
		case gwapiv1.HTTPProtocolType, gwapiv1.HTTPSProtocolType:
			hasHTTP = true
		default:
			// ignore other protocols for now
		}
	}

	if currTarget.SectionName != nil {
		sec := string(*currTarget.SectionName)
		for _, l := range targetedGateway.Spec.Listeners {
			if string(l.Name) == sec && l.Protocol == gwapiv1.TCPProtocolType {
				isTCPListener = true
				break
			}
		}
	} else {
		// whole gateway target
		if hasTCP && !hasHTTP { // all TCP
			isTCPListener = true
		}
		// Mixed protocol whole-gateway target is disallowed (avoids silent partial application)
		if hasHTTP && hasTCP {
			status.SetTranslationErrorForPolicyAncestors(&policy.Status,
				parentGateways,
				t.GatewayControllerName,
				policy.Generation,
				status.Error2ConditionMsg(fmt.Errorf(
					"cannot attach SecurityPolicy to entire Gateway %s: mixed protocols (HTTP + TCP) detected; specify sectionName to target individual listeners",
					gatewayNN.String(),
				)),
			)
			return
		}
	}

	var vErr error
	if isTCPListener {
		vErr = validateSecurityPolicyForTCP(policy)
	} else {
		vErr = validateSecurityPolicy(policy)
	}
	if vErr != nil {
		status.SetTranslationErrorForPolicyAncestors(&policy.Status,
			parentGateways,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(fmt.Errorf("invalid SecurityPolicy: %w", vErr)),
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
	status.SetAcceptedForPolicyAncestors(&policy.Status, parentGateways, t.GatewayControllerName, policy.Generation)

	// Check if this policy is overridden by other policies targeting at route and listener levels
	overriddenTargetsMessage := getOverriddenTargetsMessageForGateway(
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

	basicAuth := p.Spec.BasicAuth
	if basicAuth != nil {
		if err := validateBasicAuth(basicAuth); err != nil {
			return err
		}
	}
	return nil
}

// validateSecurityPolicyForTCP validates that the SecurityPolicy is valid for TCP routes.
// Only authorization is allowed for TCP routes.
func validateSecurityPolicyForTCP(p *egv1a1.SecurityPolicy) error {
	// For TCP routes, only authorization is supported
	if p.Spec.CORS != nil ||
		p.Spec.JWT != nil ||
		p.Spec.OIDC != nil ||
		p.Spec.APIKeyAuth != nil ||
		p.Spec.BasicAuth != nil ||
		p.Spec.ExtAuth != nil {
		return fmt.Errorf("only authorization is supported for TCP (routes/listeners)")
	}

	// Placeholder is OK: no Authorization or no rules yet.
	if p.Spec.Authorization == nil || len(p.Spec.Authorization.Rules) == 0 {
		return nil
	}

	// Validate rules when present.
	for i, rule := range p.Spec.Authorization.Rules {
		// Unsupported selectors for TCP
		if rule.Principal.JWT != nil {
			return fmt.Errorf("rule %d: JWT is not supported for TCP routes", i)
		}
		if len(rule.Principal.Headers) > 0 {
			return fmt.Errorf("rule %d: Headers are not supported for TCP routes", i)
		}

		switch rule.Action {
		case egv1a1.AuthorizationActionAllow:
			// Allow must specify at least one CIDR
			if len(rule.Principal.ClientCIDRs) == 0 {
				return fmt.Errorf("rule %d with Allow action must specify at least one ClientCIDR for TCP routes", i)
			}
			// All CIDRs must be valid
			if err := validateCIDRs(rule.Principal.ClientCIDRs); err != nil {
				return fmt.Errorf("rule %d: %w", i, err)
			}

		case egv1a1.AuthorizationActionDeny:
			// CIDRs optional, but validate if present
			if len(rule.Principal.ClientCIDRs) > 0 {
				if err := validateCIDRs(rule.Principal.ClientCIDRs); err != nil {
					return fmt.Errorf("rule %d: %w", i, err)
				}
			}
		default:
			// If the enum is already constrained, this default is never hit.
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

// validateBasicAuth validates the BasicAuth configuration.
// Currently, we only validate that the secret exists, but we don't validate
// the content of the secret. This function will be called when the security policy
// is being processed, but before the secret is actually read.
func validateBasicAuth(basicAuth *egv1a1.BasicAuth) error {
	// The actual validation of the htpasswd format will happen when the secret is read
	// in the buildBasicAuth function.
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

	// If sectionName is set, make sure its valid
	if target.SectionName != nil {
		if err := validateGatewayListenerSectionName(
			*target.SectionName,
			key,
			gateway.listeners,
		); err != nil {
			return gateway.GatewayContext, err
		}
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

// validateCIDRs checks each provided CIDR for syntactic correctness.
func validateCIDRs(cidrs []egv1a1.CIDR) error {
	for _, c := range cidrs {
		if _, _, err := net.ParseCIDR(string(c)); err != nil {
			return fmt.Errorf("invalid ClientCIDR %q: %w", c, err)
		}
	}
	return nil
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

	isTCP := target.Kind == resource.KindTCPRoute

	// Disallow sectionName for TCPRoute (no rule granularity).
	if isTCP && target.SectionName != nil {
		return route.RouteContext, &status.PolicyResolveError{
			Reason: gwapiv1a2.PolicyReasonTargetNotFound,
			Message: fmt.Sprintf("sectionName %q not supported for TCPRoute %s/%s; omit sectionName to target the whole route",
				string(*target.SectionName), policy.Namespace, string(target.Name)),
		}
	}
	// HTTPRoute rule-level validation
	if !isTCP && target.SectionName != nil {
		section := string(*target.SectionName)
		if !httpRouteRuleExists(route, section) {
			return route.RouteContext, &status.PolicyResolveError{
				Reason:  gwapiv1a2.PolicyReasonTargetNotFound,
				Message: fmt.Sprintf("No section name %s found for HTTPRoute %s/%s", section, key.Namespace, key.Name),
			}
		}
	}
	if target.SectionName == nil {
		// Whole route attachment
		if route.attached {
			message := fmt.Sprintf("Unable to target %s %s, another SecurityPolicy has already attached to it",
				string(target.Kind), string(target.Name))
			return route.RouteContext, &status.PolicyResolveError{
				Reason:  gwapiv1a2.PolicyReasonConflicted,
				Message: message,
			}
		}
		route.attached = true
	} else {
		// Section (rule) attachment (for TCP this is a synthetic identifier)
		routeRuleName := string(*target.SectionName)
		if route.attachedToRouteRules.Has(routeRuleName) {
			message := fmt.Sprintf("Unable to target RouteRule %s/%s, another SecurityPolicy has already attached to it",
				string(target.Name), routeRuleName)
			return route.RouteContext, &status.PolicyResolveError{
				Reason:  gwapiv1a2.PolicyReasonConflicted,
				Message: message,
			}
		}
		route.attachedToRouteRules.Insert(routeRuleName)
	}

	routes[key] = route

	return route.RouteContext, nil
}

func (t *Translator) translateSecurityPolicyForRoute(
	policy *egv1a1.SecurityPolicy,
	route RouteContext,
	target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName,
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
) error {
	// Build IR
	var (
		cors               *ir.CORS
		apiKeyAuth         *ir.APIKeyAuth
		basicAuth          *ir.BasicAuth
		authorization      *ir.Authorization
		err, errs          error
		hasNonExtAuthError bool
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
			err = perr.WithMessage(err, "Authorization")
			errs = errors.Join(errs, err)
		}
	}

	hasNonExtAuthError = errs != nil

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
		var extAuthErr error
		if policy.Spec.ExtAuth != nil {
			if extAuth, extAuthErr = t.buildExtAuth(
				policy,
				resources,
				gtwCtx.envoyProxy); extAuthErr != nil {
				extAuthErr = perr.WithMessage(extAuthErr, "ExtAuth")
				errs = errors.Join(errs, extAuthErr)
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
				hasNonExtAuthError = true
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
				hasNonExtAuthError = true
			}
		}

		irKey := t.getIRKey(gtwCtx.Gateway)
		// Handle TCP routes differently from HTTP routes (determine by Kind)
		if GetRouteType(route) == resource.KindTCPRoute {
			for _, listener := range parentRefCtx.listeners {
				irListener := xdsIR[irKey].GetTCPListener(irListenerName(listener))
				if irListener != nil {
					// For TCP routes, we need exact route name matching (not prefix)
					expectedRouteName := strings.TrimSuffix(prefix, "/")
					for _, r := range irListener.Routes {
						// A Policy targeting the specific scope (TCPRoute) wins over a lesser scope (Gateway)
						if r.Name == expectedRouteName && r.Security == nil {
							r.Security = &ir.SecurityFeatures{
								Authorization: authorization,
							}
						}
					}
				}
			}
			// Nothing more to do for TCP for this parentRef
			continue
		}
		for _, listener := range parentRefCtx.listeners {
			irListener := xdsIR[irKey].GetHTTPListener(irListenerName(listener))
			if irListener != nil {
				for _, r := range irListener.Routes {
					// If specified the sectionName must match route rule from ir route metadata.
					if target.SectionName != nil && string(*target.SectionName) != r.Metadata.SectionName {
						continue
					}

					// A Policy targeting the most specific scope(xRoute rule) wins over a policy
					// targeting a lesser specific scope(xRoute).
					if strings.HasPrefix(r.Name, prefix) {
						// if already set - there's a specific level policy, so skip.
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
							// If there is only error for ext auth and ext auth is set to fail open, then skip the ext auth
							// and allow the request to go through.
							// Otherwise, return a 500 direct response to avoid unauthorized access.
							shouldFailOpen := extAuthErr != nil && !hasNonExtAuthError && ptr.Deref(policy.Spec.ExtAuth.FailOpen, false)
							if !shouldFailOpen {
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
		cors                  *ir.CORS
		jwt                   *ir.JWT
		oidc                  *ir.OIDC
		apiKeyAuth            *ir.APIKeyAuth
		basicAuth             *ir.BasicAuth
		extAuth               *ir.ExtAuth
		authorization         *ir.Authorization
		extAuthErr, err, errs error
		hasNonExtAuthError    bool
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

	if policy.Spec.Authorization != nil {
		if authorization, err = t.buildAuthorization(policy); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	hasNonExtAuthError = errs != nil

	if policy.Spec.ExtAuth != nil {
		if extAuth, extAuthErr = t.buildExtAuth(
			policy,
			resources,
			gateway.envoyProxy); extAuthErr != nil {
			extAuthErr = perr.WithMessage(extAuthErr, "ExtAuth")
			errs = errors.Join(errs, extAuthErr)
		}
	}

	// Apply IR to all the routes within the specific Gateway that originated
	// from the gateway to which this security policy was attached.
	// If the feature is already set, then skip it, since it must have be
	// set by a policy attaching to the route
	//
	// Note: there are multiple features in a security policy, even if some of them
	// are invalid, we still want to apply the valid ones.
	// Defensive checks: translateSecurityPolicyForGateway expects a valid GatewayContext
	// but we validate here to avoid panics when tests or upstream callers pass a nil/partial context.
	if gateway == nil {
		return fmt.Errorf("translateSecurityPolicyForGateway: gateway context is nil for policy %s/%s", policy.Namespace, policy.Name)
	}
	if gateway.Gateway == nil {
		return fmt.Errorf("translateSecurityPolicyForGateway: embedded Gateway is nil for gateway context %v (policy %s/%s)", gateway, policy.Namespace, policy.Name)
	}

	irKey := t.getIRKey(gateway.Gateway)
	// Should exist since we've validated this
	x, ok := xdsIR[irKey]
	if !ok || x == nil {
		return fmt.Errorf("no IR found for gateway %s (cannot apply SecurityPolicy)", irKey)
	}
	sectionName := ""
	if target.SectionName != nil {
		sectionName = string(*target.SectionName)
	}

	// Detect if target is TCP listener (or all listeners TCP when sectionName empty).
	isTCPListener := false
	if sectionName != "" {
		// defensive loop: ensure gateway.Spec is available
		for _, l := range gateway.Spec.Listeners {
			if string(l.Name) == sectionName && l.Protocol == gwapiv1.TCPProtocolType {
				isTCPListener = true
				break
			}
		}
	} else {
		allTCP := true
		for _, l := range gateway.Spec.Listeners {
			if l.Protocol != gwapiv1.TCPProtocolType {
				allTCP = false
				break
			}
		}
		isTCPListener = allTCP
	}

	policyTarget := irStringKey(policy.Namespace, string(target.Name))

	if isTCPListener {
		// Apply ONLY Authorization to all TCP routes under the targeted listener(s).
		for _, tl := range x.TCP {
			if tl == nil {
				continue
			}
			// defensive: ensure metadata present
			// Listener name format: namespace/gatewayName/listenerName
			// compute effective section name from metadata if present, else fallback to name suffix
			effectiveSection := ""
			if tl.Metadata != nil && tl.Metadata.SectionName != "" {
				effectiveSection = tl.Metadata.SectionName
			} else {
				// fallback: last path segment of tl.Name (`namespace/gateway/listener`)
				// TODO: remove fallback once IR builder guarantees tl.Metadata.SectionName is populated.
				if idx := strings.LastIndex(tl.Name, "/"); idx >= 0 && idx < len(tl.Name)-1 {
					effectiveSection = tl.Name[idx+1:]
				} else {
					effectiveSection = tl.Name
				}
			}

			if t.MergeGateways && !strings.HasPrefix(tl.Name, policyTarget) {
				continue
			}
			if sectionName != "" && effectiveSection != sectionName {
				continue
			}
			for _, r := range tl.Routes {
				if r == nil {
					continue
				}
				if r.Security != nil {
					continue
				}
				r.Security = &ir.SecurityFeatures{
					Authorization: authorization,
				}
			}
		}
		return errs
	}

	// HTTP branch: defensive nil checks for listeners/metadata
	for _, h := range x.HTTP {
		if h == nil {
			continue
		}
		// A HTTPListener name has the format namespace/gatewayName/listenerName
		gatewayNameEnd := strings.LastIndex(h.Name, "/")
		if gatewayNameEnd <= 0 || gatewayNameEnd >= len(h.Name) {
			continue
		}
		gatewayName := h.Name[0:gatewayNameEnd]
		if t.MergeGateways && gatewayName != policyTarget {
			continue
		}
		// If specified the sectionName must match listenerName from ir listener metadata.
		if h.Metadata == nil {
			continue
		}
		if target.SectionName != nil && string(*target.SectionName) != h.Metadata.SectionName {
			continue
		}
		// A Policy targeting the specific scope(xRoute rule, xRoute, Gateway listener) wins over a policy
		// targeting a lesser specific scope(Gateway).
		for _, r := range h.Routes {
			if r == nil {
				continue
			}
			// if already set - there's a specific level policy, so skip.
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
				// If there is only error for ext auth and ext auth is set to fail open, then skip the ext auth
				// and allow the request to go through.
				// Otherwise, return a 500 direct response to avoid unauthorized access.
				shouldFailOpen := extAuthErr != nil && !hasNonExtAuthError && ptr.Deref(policy.Spec.ExtAuth.FailOpen, false)
				if !shouldFailOpen {
					r.DirectResponse = &ir.CustomResponse{
						StatusCode: ptr.To(uint32(500)),
					}
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

	irCORS := &ir.CORS{
		AllowOrigins:     allowOrigins,
		AllowMethods:     cors.AllowMethods,
		AllowHeaders:     cors.AllowHeaders,
		ExposeHeaders:    cors.ExposeHeaders,
		AllowCredentials: cors.AllowCredentials != nil && *cors.AllowCredentials,
	}

	if cors.MaxAge != nil {
		if d, err := time.ParseDuration(string(*cors.MaxAge)); err == nil {
			irCORS.MaxAge = ir.MetaV1DurationPtr(d)
		}
	}

	return irCORS
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
		clientID              string
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

	// Client ID can be specified either as a string or as a reference to a secret.
	switch {
	case oidc.ClientID != nil:
		clientID = *oidc.ClientID
	case oidc.ClientIDRef != nil:
		var clientIDSecret *corev1.Secret
		if clientIDSecret, err = t.validateSecretRef(false, from, *oidc.ClientIDRef, resources); err != nil {
			return nil, err
		}
		clientIDBytes, ok := clientIDSecret.Data[egv1a1.OIDCClientIDKey]
		if !ok || len(clientIDBytes) == 0 {
			return nil, fmt.Errorf("client ID not found in secret %s/%s", clientIDSecret.Namespace, clientIDSecret.Name)
		}
		clientID = string(clientIDBytes)
	default:
		// This is just a sanity check - the CRD validation should have caught this.
		return nil, fmt.Errorf("client ID must be specified in OIDC policy %s/%s", policy.Namespace, policy.Name)
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

	irOIDC := &ir.OIDC{
		Name:                  irConfigName(policy),
		Provider:              *provider,
		ClientID:              clientID,
		ClientSecret:          clientSecretBytes,
		Scopes:                scopes,
		Resources:             oidc.Resources,
		RedirectURL:           redirectURL,
		RedirectPath:          redirectPath,
		LogoutPath:            logoutPath,
		ForwardAccessToken:    forwardAccessToken,
		RefreshToken:          refreshToken,
		CookieSuffix:          suffix,
		CookieNameOverrides:   policy.Spec.OIDC.CookieNames,
		CookieDomain:          policy.Spec.OIDC.CookieDomain,
		CookieConfig:          policy.Spec.OIDC.CookieConfig,
		HMACSecret:            hmacData,
		PassThroughAuthHeader: passThroughAuthHeader,
		DenyRedirect:          oidc.DenyRedirect,
	}

	if oidc.DefaultTokenTTL != nil {
		if d, err := time.ParseDuration(string(*oidc.DefaultTokenTTL)); err == nil {
			irOIDC.DefaultTokenTTL = ir.MetaV1DurationPtr(d)
		}
	}

	if oidc.DefaultRefreshTokenTTL != nil {
		if d, err := time.ParseDuration(string(*oidc.DefaultRefreshTokenTTL)); err == nil {
			irOIDC.DefaultRefreshTokenTTL = ir.MetaV1DurationPtr(d)
		}
	}

	return irOIDC, nil
}

func (t *Translator) buildOIDCProvider(policy *egv1a1.SecurityPolicy, resources *resource.Resources, envoyProxy *egv1a1.EnvoyProxy) (*ir.OIDCProvider, error) {
	var (
		provider              = policy.Spec.OIDC.Provider
		tokenEndpoint         string
		authorizationEndpoint string
		endSessionEndpoint    *string
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
		discoveredConfig, err := fetchEndpointsFromIssuer(provider.Issuer, providerTLS)
		if err != nil {
			return nil, fmt.Errorf("error fetching endpoints from issuer: %w", err)
		}
		tokenEndpoint = discoveredConfig.TokenEndpoint
		authorizationEndpoint = discoveredConfig.AuthorizationEndpoint
		// endSessionEndpoint is optional, and we prioritize using the one provided in the well-known configuration.
		if discoveredConfig.EndSessionEndpoint != nil && *discoveredConfig.EndSessionEndpoint != "" {
			endSessionEndpoint = discoveredConfig.EndSessionEndpoint
		}
	} else {
		tokenEndpoint = *provider.TokenEndpoint
		authorizationEndpoint = *provider.AuthorizationEndpoint
		endSessionEndpoint = provider.EndSessionEndpoint
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
		EndSessionEndpoint:    endSessionEndpoint,
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
	TokenEndpoint         string  `json:"token_endpoint"`
	AuthorizationEndpoint string  `json:"authorization_endpoint"`
	EndSessionEndpoint    *string `json:"end_session_endpoint,omitempty"`
}

func fetchEndpointsFromIssuer(issuerURL string, providerTLS *ir.TLSUpstreamConfig) (*OpenIDConfig, error) {
	var (
		tlsConfig *tls.Config
		err       error
	)

	if providerTLS != nil {
		if tlsConfig, err = providerTLS.ToTLSConfig(); err != nil {
			return nil, err
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
		return nil, err
	}

	return &config, nil
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
		Credentials:           credentials,
		ExtractFrom:           extractFrom,
		ForwardClientIDHeader: policy.Spec.APIKeyAuth.ForwardClientIDHeader,
		Sanitize:              policy.Spec.APIKeyAuth.Sanitize,
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

	// Validate the htpasswd format
	if err := validateHtpasswdFormat(usersSecretBytes); err != nil {
		return nil, err
	}

	return &ir.BasicAuth{
		Name:                  irConfigName(policy),
		Users:                 usersSecretBytes,
		ForwardUsernameHeader: basicAuth.ForwardUsernameHeader,
	}, nil
}

// validateHtpasswdFormat validates that the htpasswd data is in the correct format.
// Currently, only the SHA format is supported by Envoy.
func validateHtpasswdFormat(data []byte) error {
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid htpasswd format: each line must be in the format 'username:password'")
		}

		password := parts[1]
		if !strings.HasPrefix(password, "{SHA}") {
			return fmt.Errorf("unsupported htpasswd format: please use {SHA}")
		}
	}
	return nil
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
		Timeout:          parseExtAuthTimeout(policy.Spec.ExtAuth.Timeout),
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

// parseExtAuthTimeout parses the timeout from gwapiv1.Duration to metav1.Duration.
func parseExtAuthTimeout(timeout *gwapiv1.Duration) *metav1.Duration {
	if timeout == nil {
		return nil
	}
	d, err := time.ParseDuration(string(*timeout))
	if err != nil {
		return nil
	}
	return &metav1.Duration{
		Duration: d,
	}
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
	if backendKind == resource.KindService || backendKind == resource.KindServiceImport {
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

// httpRouteRuleExists checks if a rule (section name) exists on the HTTPRoute.
func httpRouteRuleExists(route *policyRouteTargetContext, section string) bool {
	if route == nil || route.RouteContext == nil {
		return false
	}
	for _, rn := range GetRuleNames(route.RouteContext) {
		if string(rn) == section {
			return true
		}
	}
	return false
}
