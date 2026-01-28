// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"crypto/tls"
	"encoding/base64"
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
	defaultRefreshToken          = true
	defaultPassThroughAuthHeader = false
	defaultOIDCHTTPTimeout       = 5 * time.Second

	// nolint: gosec
	oidcHMACSecretName = "envoy-oidc-hmac"
	oidcHMACSecretKey  = "hmac-secret"
)

// deprecatedFieldsUsedInSecurityPolicy returns a map of deprecated field paths to their alternatives.
func deprecatedFieldsUsedInSecurityPolicy(policy *egv1a1.SecurityPolicy) map[string]string {
	deprecatedFields := make(map[string]string)
	if policy.Spec.TargetRef != nil {
		deprecatedFields["spec.targetRef"] = "spec.targetRefs"
	}
	if policy.Spec.ExtAuth != nil {
		if policy.Spec.ExtAuth.GRPC != nil && policy.Spec.ExtAuth.GRPC.BackendRef != nil {
			deprecatedFields["spec.extAuth.grpc.backendRef"] = "spec.extAuth.grpc.backendRefs"
		}
		if policy.Spec.ExtAuth.HTTP != nil && policy.Spec.ExtAuth.HTTP.BackendRef != nil {
			deprecatedFields["spec.extAuth.http.backendRef"] = "spec.extAuth.http.backendRefs"
		}
	}
	return deprecatedFields
}

func (t *Translator) ProcessSecurityPolicies(
	securityPolicies []*egv1a1.SecurityPolicy,
	gateways []*GatewayContext,
	routes []RouteContext,
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
) []*egv1a1.SecurityPolicy {
	// Cache is only reused during one translation across multiple routes and gateways.
	// The failed fetches will be retried in the next translation when the provider resources are reconciled again.
	t.oidcDiscoveryCache = newOIDCDiscoveryCache()

	// SecurityPolicies are already sorted by the provider layer

	// First build a map out of the routes and gateways for faster lookup since users might have thousands of routes or more.
	// For gateways this probably isn't quite as necessary.
	routeMapSize := len(routes)
	gatewayMapSize := len(gateways)
	policyMapSize := len(securityPolicies)

	// Pre-allocate result slice and maps with estimated capacity to reduce memory allocations
	res := make([]*egv1a1.SecurityPolicy, 0, len(securityPolicies))
	routeMap := make(map[policyTargetRouteKey]*policyRouteTargetContext, routeMapSize)
	for _, route := range routes {
		key := policyTargetRouteKey{
			Kind:      string(route.GetRouteType()),
			Name:      route.GetName(),
			Namespace: route.GetNamespace(),
		}
		routeMap[key] = &policyRouteTargetContext{RouteContext: route}
	}

	gatewayMap := make(map[types.NamespacedName]*policyGatewayTargetContext, gatewayMapSize)
	for _, gw := range gateways {
		key := utils.NamespacedName(gw)
		gatewayMap[key] = &policyGatewayTargetContext{GatewayContext: gw}
	}

	// Map of Gateway to the routes attached to it.
	// The routes are grouped by sectionNames of their targetRefs
	gatewayRouteMap := make(map[string]map[string]sets.Set[string], gatewayMapSize)

	handledPolicies := make(map[types.NamespacedName]*egv1a1.SecurityPolicy, policyMapSize)

	// Translate
	// 1. First translate Policies targeting RouteRules
	// 2. Next translate Policies targeting xRoutes
	// 3. Then translate Policies targeting Listeners
	// 4. Finally, the policies targeting Gateways

	// Process the policies targeting RouteRules (HTTP + TCP)
	for _, currPolicy := range securityPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, routes, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			// If the target is not a gateway, then it's an xRoute. If the section name is defined, then it's a route rule.
			if currTarget.Kind != resource.KindGateway && currTarget.SectionName != nil {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				t.processSecurityPolicyForRoute(resources, xdsIR,
					routeMap, gatewayRouteMap, policy, currTarget)
			}
		}
	}
	// Process the policies targeting xRoutes (HTTP + TCP)
	for _, currPolicy := range securityPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, routes, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			// If the target is not a gateway, then it's an xRoute. If the section name is not defined, then it's a route.
			if currTarget.Kind != resource.KindGateway && currTarget.SectionName == nil {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy
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
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, gateways, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			// If the target is a gateway and the section name is defined, then it's a listener.
			if currTarget.Kind == resource.KindGateway && currTarget.SectionName != nil {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy
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
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, gateways, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			// If the target is a gateway and the section name is not defined, then it's a gateway.
			if currTarget.Kind == resource.KindGateway && currTarget.SectionName == nil {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy
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
	currTarget gwapiv1.LocalPolicyTargetReferenceWithSectionName,
) {
	var (
		targetedRoute  RouteContext
		parentGateways []*gwapiv1.ParentReference
		resolveErr     *status.PolicyResolveError
	)

	targetedRoute, resolveErr = resolveSecurityPolicyRouteTargetRef(policy, currTarget, routeMap)
	// Skip if the route is not found
	// It's not necessarily an error because the SecurityPolicy may be
	// reconciled by multiple controllers. And the other controller may
	// have the target route.
	if targetedRoute == nil {
		return
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
			ancestorRef := getAncestorRefForPolicy(gwNN, p.SectionName)
			parentGateways = append(parentGateways, &ancestorRef)
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

	// Protocol-specific validation: pick the appropriate validator and message,
	// then run it once to keep the flow linear and easier to read.
	validator := validateSecurityPolicy
	errMsg := "invalid SecurityPolicy"
	if currTarget.Kind == resource.KindTCPRoute {
		validator = validateSecurityPolicyForTCP
		errMsg = "invalid SecurityPolicy for TCP route"
	}
	if err := validator(policy); err != nil {
		status.SetTranslationErrorForPolicyAncestors(&policy.Status,
			parentGateways,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(fmt.Errorf("%s: %w", errMsg, err)),
		)

		return
	}

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

	// Check for deprecated fields and set warning if any are found
	if deprecatedFields := deprecatedFieldsUsedInSecurityPolicy(policy); len(deprecatedFields) > 0 {
		status.SetDeprecatedFieldsWarningForPolicyAncestors(&policy.Status, parentGateways, t.GatewayControllerName, policy.Generation, deprecatedFields)
	}

	// Check if this policy is overridden by other policies targeting at route rule levels
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

func (t *Translator) processSecurityPolicyForGateway(
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
	gatewayMap map[types.NamespacedName]*policyGatewayTargetContext,
	gatewayRouteMap map[string]map[string]sets.Set[string],
	policy *egv1a1.SecurityPolicy,
	currTarget gwapiv1.LocalPolicyTargetReferenceWithSectionName,
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
	parentGateway := getAncestorRefForPolicy(gatewayNN, currTarget.SectionName)

	// Set conditions for resolve error, then skip current gateway
	if resolveErr != nil {
		status.SetResolveErrorForPolicyAncestor(&policy.Status,
			&parentGateway,
			t.GatewayControllerName,
			policy.Generation,
			resolveErr,
		)

		return
	}

	if err := t.translateSecurityPolicyForGateway(policy, targetedGateway, currTarget, resources, xdsIR); err != nil {
		status.SetTranslationErrorForPolicyAncestor(&policy.Status,
			&parentGateway,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(err),
		)
	}

	// Set Accepted condition if it is unset
	status.SetAcceptedForPolicyAncestor(&policy.Status, &parentGateway, t.GatewayControllerName, policy.Generation)

	// Check if this policy is overridden by other policies targeting at route and listener levels
	overriddenTargetsMessage := getOverriddenTargetsMessageForGateway(
		gatewayMap[gatewayNN], gatewayRouteMap[gatewayNN.String()], currTarget.SectionName)
	if overriddenTargetsMessage != "" {
		status.SetConditionForPolicyAncestor(&policy.Status,
			&parentGateway,
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

// validateSecurityPolicyForTCP ensures SecurityPolicy usage on TCP is compatible.
//
// TCP supports Authorization with ClientCIDRs ONLY.
// - Principals.JWT      => invalid (HTTP-only)
// - Principals.Headers  => invalid (HTTP-only)
// - Empty/no Authorization is allowed and results in no-op on TCP.
// Returns an error when any HTTP-only field is present or CIDRs are invalid.
func validateSecurityPolicyForTCP(p *egv1a1.SecurityPolicy) error {
	if p.Spec.CORS != nil || p.Spec.JWT != nil || p.Spec.OIDC != nil || p.Spec.APIKeyAuth != nil || p.Spec.BasicAuth != nil || p.Spec.ExtAuth != nil {
		return fmt.Errorf("only authorization is supported for TCP (routes/listeners)")
	}
	if p.Spec.Authorization == nil || len(p.Spec.Authorization.Rules) == 0 {
		return nil
	}
	for i, rule := range p.Spec.Authorization.Rules {
		if rule.Principal.JWT != nil {
			return fmt.Errorf("rule %d: JWT not supported for TCP", i)
		}
		if len(rule.Principal.Headers) > 0 {
			return fmt.Errorf("rule %d: headers not supported for TCP", i)
		}
		if err := validateCIDRs(rule.Principal.ClientCIDRs); err != nil {
			return fmt.Errorf("rule %d: %w", i, err)
		}
	}
	return nil
}

// validateCIDRs validates CIDR strings for TCP authorization rules.
func validateCIDRs(cidrs []egv1a1.CIDR) error {
	for _, c := range cidrs {
		if _, _, err := net.ParseCIDR(string(c)); err != nil {
			return fmt.Errorf("invalid ClientCIDR %q: %w", c, err)
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
func validateBasicAuth(_ *egv1a1.BasicAuth) error {
	// The actual validation of the htpasswd format will happen when the secret is read
	// in the buildBasicAuth function.
	return nil
}

func resolveSecurityPolicyGatewayTargetRef(
	policy *egv1a1.SecurityPolicy,
	target gwapiv1.LocalPolicyTargetReferenceWithSectionName,
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
				Reason:  gwapiv1.PolicyReasonConflicted,
				Message: message,
			}
		}
		gateway.attached = true
	} else {
		listenerName := string(*target.SectionName)
		if gateway.attachedToListeners != nil && gateway.attachedToListeners.Has(listenerName) {
			message := fmt.Sprintf("Unable to target Listener %s/%s, another SecurityPolicy has already attached to it",
				string(target.Name), listenerName)

			return gateway.GatewayContext, &status.PolicyResolveError{
				Reason:  gwapiv1.PolicyReasonConflicted,
				Message: message,
			}
		}
		if gateway.attachedToListeners == nil {
			gateway.attachedToListeners = make(sets.Set[string])
		}
		gateway.attachedToListeners.Insert(listenerName)
	}

	gateways[key] = gateway

	return gateway.GatewayContext, nil
}

func resolveSecurityPolicyRouteTargetRef(
	policy *egv1a1.SecurityPolicy,
	target gwapiv1.LocalPolicyTargetReferenceWithSectionName,
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

	// If sectionName is set, make sure its valid
	if target.SectionName != nil {
		if err := validateRouteRuleSectionName(*target.SectionName, key, route); err != nil {
			return route.RouteContext, err
		}
	}

	if target.SectionName == nil {
		// Check if another policy targeting the same xRoute exists
		if route.attached {
			message := fmt.Sprintf("Unable to target %s %s, another SecurityPolicy has already attached to it",
				string(target.Kind), string(target.Name))

			return route.RouteContext, &status.PolicyResolveError{
				Reason:  gwapiv1.PolicyReasonConflicted,
				Message: message,
			}
		}
		route.attached = true
	} else {
		routeRuleName := string(*target.SectionName)
		if route.attachedToRouteRules != nil && route.attachedToRouteRules.Has(routeRuleName) {
			message := fmt.Sprintf("Unable to target RouteRule %s/%s, another SecurityPolicy has already attached to it",
				string(target.Name), routeRuleName)
			return route.RouteContext, &status.PolicyResolveError{
				Reason:  gwapiv1.PolicyReasonConflicted,
				Message: message,
			}
		}
		if route.attachedToRouteRules == nil {
			route.attachedToRouteRules = make(sets.Set[string])
		}
		route.attachedToRouteRules.Insert(routeRuleName)
	}

	routes[key] = route

	return route.RouteContext, nil
}

func (t *Translator) translateSecurityPolicyForRoute(
	policy *egv1a1.SecurityPolicy,
	route RouteContext,
	target gwapiv1.LocalPolicyTargetReferenceWithSectionName,
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
	routesWithDirectResponse := sets.New[string]()
	for _, p := range parentRefs {
		// Skip if this parentRef was not processed by this translator
		// (e.g., references a Gateway with a different GatewayClass)
		parentRefCtx := route.GetRouteParentContext(p)
		if parentRefCtx == nil {
			continue
		}
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

		// Pre-create security features to avoid repeated allocations
		securityFeatures := &ir.SecurityFeatures{
			CORS:          cors,
			JWT:           jwt,
			OIDC:          oidc,
			APIKeyAuth:    apiKeyAuth,
			BasicAuth:     basicAuth,
			ExtAuth:       extAuth,
			Authorization: authorization,
		}

		// Pre-create error response to avoid repeated allocations
		var errorResponse *ir.CustomResponse
		if errs != nil {
			shouldFailOpen := extAuthErr != nil && !hasNonExtAuthError && ptr.Deref(policy.Spec.ExtAuth.FailOpen, false)
			if !shouldFailOpen {
				errorResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(500)),
				}
			}
		}

		irKey := t.getIRKey(gtwCtx.Gateway)
		switch route.GetRouteType() {
		case resource.KindTCPRoute:
			for _, listener := range parentRefCtx.listeners {
				tl := xdsIR[irKey].GetTCPListener(irListenerName(listener))
				for _, r := range tl.Routes {
					// If target.SectionName is specified it must match the route-rule section name
					// in the IR. For HTTP/GRPC routes this is r.Metadata.SectionName; for TCP
					// routes the section name is currently stored on r.Destination.Metadata.SectionName.
					if target.SectionName != nil && string(*target.SectionName) != r.Destination.Metadata.SectionName {
						continue
					}

					if r.Authorization != nil {
						continue
					}
					// Only authorization for TCP
					authCopy := *authorization
					r.Authorization = &authCopy
				}
			}
		case resource.KindHTTPRoute, resource.KindGRPCRoute:
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

							r.Security = securityFeatures
							if errorResponse != nil {
								// Return a 500 direct response to avoid unauthorized access
								r.DirectResponse = errorResponse
								routesWithDirectResponse.Insert(r.Name)
							}
						}
					}
				}
			}
		}
	}
	if len(routesWithDirectResponse) > 0 {
		t.Logger.Info("setting 500 direct response in routes due to errors in SecurityPolicy",
			"policy", fmt.Sprintf("%s/%s", policy.Namespace, policy.Name),
			"routes", sets.List(routesWithDirectResponse),
			"error", errs,
		)
	}
	return nil
}

func (t *Translator) translateSecurityPolicyForGateway(
	policy *egv1a1.SecurityPolicy,
	gateway *GatewayContext,
	target gwapiv1.LocalPolicyTargetReferenceWithSectionName,
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

	var extAuth *ir.ExtAuth
	var extAuthErr error
	if policy.Spec.ExtAuth != nil {
		if extAuth, extAuthErr = t.buildExtAuth(
			policy,
			resources,
			gateway.envoyProxy); extAuthErr != nil {
			extAuthErr = perr.WithMessage(extAuthErr, "ExtAuth")
			errs = errors.Join(errs, extAuthErr)
		}
	}

	var oidc *ir.OIDC
	if policy.Spec.OIDC != nil {
		if oidc, err = t.buildOIDC(
			policy,
			resources,
			gateway.envoyProxy); err != nil {
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
			gateway.envoyProxy); err != nil {
			err = perr.WithMessage(err, "JWT")
			errs = errors.Join(errs, err)
			hasNonExtAuthError = true
		}
	}

	// Apply IR to all relevant Listeners
	// If sectionName is specified, only apply to that listener
	// If not, apply to all listeners
	irKey := t.getIRKey(gateway.Gateway)
	// Apply IR to HTTP Listeners
	for _, listener := range xdsIR[irKey].HTTP {
		if target.SectionName != nil && string(*target.SectionName) != listener.Name {
			continue
		}

		// A Policy targeting the most specific scope(listener) wins over a policy
		// targeting a lesser specific scope(gateway).
		if listener.IsHTTP2 {
			// for grpc listener, we need to check the security features on the routes
			for _, r := range listener.Routes {
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
					shouldFailOpen := extAuthErr != nil && !hasNonExtAuthError && ptr.Deref(policy.Spec.ExtAuth.FailOpen, false)
					if !shouldFailOpen {
						r.DirectResponse = &ir.CustomResponse{
							StatusCode: ptr.To(uint32(500)),
						}

						t.Logger.Info("setting 500 direct response in routes due to errors in SecurityPolicy",
							"policy", fmt.Sprintf("%s/%s", policy.Namespace, policy.Name),
							"routes", r.Name,
							"error", errs,
						)
					}
				}
			}
		} else {
			// If already set - there's a specific level policy, so skip
			if listener.Security != nil {
				continue
			}

			listener.Security = &ir.SecurityFeatures{
				CORS:          cors,
				JWT:           jwt,
				OIDC:          oidc,
				APIKeyAuth:    apiKeyAuth,
				BasicAuth:     basicAuth,
				ExtAuth:       extAuth,
				Authorization: authorization,
			}

			if errs != nil {
				shouldFailOpen := extAuthErr != nil && !hasNonExtAuthError && ptr.Deref(policy.Spec.ExtAuth.FailOpen, false)
				if !shouldFailOpen {
					// Add a dummy route to return 500
					// This should be the first route in the listener
					listener.Routes = append([]*ir.HTTPRoute{
						{
							Name: "security-policy-error",
							DirectResponse: &ir.CustomResponse{
								StatusCode: ptr.To(uint32(500)),
							},
							PathMatch: &ir.StringMatch{
								MatchType: ir.StringMatchPrefix,
								Name:      "/",
							},
						},
					}, listener.Routes...)

					t.Logger.Info("setting 500 direct response in listeners due to errors in SecurityPolicy",
						"policy", fmt.Sprintf("%s/%s", policy.Namespace, policy.Name),
						"listener", listener.Name,
						"error", errs,
					)
				}
			}
		}

	}

	// Apply IR to TCP Listeners
	// TCP listeners only support Authorization
	for _, listener := range xdsIR[irKey].TCP {
		if target.SectionName != nil && string(*target.SectionName) != listener.Name {
			continue
		}
		// If already set - there's a specific level policy, so skip
		if listener.Authorization != nil {
			continue
		}
		// Skip if authorization is not present/set
		if authorization == nil {
			continue
		}
		authCopy := *authorization
		listener.Authorization = &authCopy
	}

	return nil
}

func (t *Translator) buildOIDC(
	policy *egv1a1.SecurityPolicy,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.OIDC, error) {
	var (
		clientID               string
		clientSecret           string
		err                    error
		provider               *ir.OIDCProvider
		scopes                 = policy.Spec.OIDC.Scopes
		resourcesList          = policy.Spec.OIDC.Resources
		redirectURI            = defaultRedirectURL
		logoutPath             = defaultLogoutPath
		cookieNames            *ir.OIDCCookieNames
		cookieConfig           *ir.OIDCCookieConfig
		cookieDomain           *string
		forwardAccessToken     = defaultForwardAccessToken
		defaultTokenTTL        = time.Duration(0)
		refreshToken           = defaultRefreshToken
		defaultRefreshTokenTTL = 604800 * time.Second
		csrfTokenTTL           = 600 * time.Second
		passThroughAuthHeader  = defaultPassThroughAuthHeader
		disableTokenEncryption = false
	)

	if provider, err = t.buildOIDCProvider(policy, resources, envoyProxy); err != nil {
		return nil, err
	}

	from := crossNamespaceFrom{
		group:     egv1a1.GroupName,
		kind:      egv1a1.KindSecurityPolicy,
		namespace: policy.Namespace,
	}

	// Resolve the client ID
	if policy.Spec.OIDC.ClientID != nil {
		clientID = *policy.Spec.OIDC.ClientID
	} else if policy.Spec.OIDC.ClientIDRef != nil {
		ref := policy.Spec.OIDC.ClientIDRef
		if clientID, err = t.getOpaqueSecretValue(from, ref.Name, string(ptr.Deref(ref.Namespace, gwapiv1.Namespace(policy.Namespace))), egv1a1.OIDCClientIDKey); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("one of clientID or clientIDRef must be specified")
	}

	// Resolve the client Secret
	ref := policy.Spec.OIDC.ClientSecret
	if clientSecret, err = t.getOpaqueSecretValue(from, ref.Name, string(ptr.Deref(ref.Namespace, gwapiv1.Namespace(policy.Namespace))), egv1a1.OIDCClientSecretKey); err != nil {
		return nil, err
	}

	if policy.Spec.OIDC.RedirectURL != nil {
		redirectURI = *policy.Spec.OIDC.RedirectURL
	}

	redirectPath, err := extractRedirectPath(redirectURI)
	if err != nil {
		return nil, err
	}

	if policy.Spec.OIDC.LogoutPath != nil {
		logoutPath = *policy.Spec.OIDC.LogoutPath
	}

	if policy.Spec.OIDC.ForwardAccessToken != nil {
		forwardAccessToken = *policy.Spec.OIDC.ForwardAccessToken
	}

	if policy.Spec.OIDC.DefaultTokenTTL != nil {
		defaultTokenTTL = policy.Spec.OIDC.DefaultTokenTTL.Duration
	}

	if policy.Spec.OIDC.RefreshToken != nil {
		refreshToken = *policy.Spec.OIDC.RefreshToken
	}

	if policy.Spec.OIDC.DefaultRefreshTokenTTL != nil {
		defaultRefreshTokenTTL = policy.Spec.OIDC.DefaultRefreshTokenTTL.Duration
	}

	if policy.Spec.OIDC.CSRFTokenTTL != nil {
		csrfTokenTTL = policy.Spec.OIDC.CSRFTokenTTL.Duration
	}

	if policy.Spec.OIDC.DisableTokenEncryption != nil {
		disableTokenEncryption = *policy.Spec.OIDC.DisableTokenEncryption
	}

	if policy.Spec.OIDC.PassThroughAuthHeader != nil {
		passThroughAuthHeader = *policy.Spec.OIDC.PassThroughAuthHeader
	}

	if policy.Spec.OIDC.CookieNames != nil {
		cookieNames = &ir.OIDCCookieNames{}
		if policy.Spec.OIDC.CookieNames.AccessToken != nil {
			cookieNames.AccessToken = *policy.Spec.OIDC.CookieNames.AccessToken
		}
		if policy.Spec.OIDC.CookieNames.IDToken != nil {
			cookieNames.IDToken = *policy.Spec.OIDC.CookieNames.IDToken
		}
	}

	if policy.Spec.OIDC.CookieConfig != nil {
		cookieConfig = &ir.OIDCCookieConfig{}
		if policy.Spec.OIDC.CookieConfig.SameSite != nil {
			cookieConfig.SameSite = *policy.Spec.OIDC.CookieConfig.SameSite
		}
	}

	if policy.Spec.OIDC.CookieDomain != nil {
		cookieDomain = policy.Spec.OIDC.CookieDomain
	}

	// Append the openid scope if it is not already present
	scopes = appendOpenidScopeIfNotExist(scopes)

	irOIDC := &ir.OIDC{
		Provider:               provider,
		ClientID:               clientID,
		ClientSecret:           clientSecret,
		RedirectURI:            redirectURI,
		RedirectPath:           redirectPath,
		LogoutPath:             logoutPath,
		ForwardAccessToken:     forwardAccessToken,
		DefaultTokenTTL:        defaultTokenTTL,
		RefreshToken:           refreshToken,
		DefaultRefreshTokenTTL: defaultRefreshTokenTTL,
		CSRFTokenTTL:           csrfTokenTTL,
		Scopes:                 scopes,
		Resources:              resourcesList,
		CookieNames:            cookieNames,
		CookieConfig:           cookieConfig,
		CookieDomain:           cookieDomain,
		DisableTokenEncryption: disableTokenEncryption,
		PassThroughAuthHeader:  passThroughAuthHeader,
	}

	// OIDC filter configuration: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/oauth2/v3/oauth.proto
	if policy.Spec.OIDC.DenyRedirect != nil {
		denyRedirect := &ir.OIDCDenyRedirect{}
		if len(policy.Spec.OIDC.DenyRedirect.Headers) > 0 {
			denyRedirect.Headers = make([]ir.CommonHeaderMatch, 0, len(policy.Spec.OIDC.DenyRedirect.Headers))
			for _, denyRedirectHeader := range policy.Spec.OIDC.DenyRedirect.Headers {
				m := t.getIRStringMatch(denyRedirectHeader.StringMatch)
				h := ir.CommonHeaderMatch{
					Name:        denyRedirectHeader.Name,
					StringMatch: &m,
				}
				denyRedirect.Headers = append(denyRedirect.Headers, h)
			}
		}
		irOIDC.DenyRedirect = denyRedirect
	}

	return irOIDC, nil
}

func (t *Translator) buildOIDCProvider(
	policy *egv1a1.SecurityPolicy,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.OIDCProvider, error) {
	var (
		provider              = policy.Spec.OIDC.Provider
		tokenEndpoint         string
		authorizationEndpoint string
		endSessionEndpoint    *string
		issuer                string
		protocol              ir.AppProtocol
		rd                    *ir.RouteDestination
		traffic               *ir.TrafficFeatures
		providerTLS           *ir.TLSUpstreamConfig
		err                   error
	)

	// Resolve issuer
	switch {
	case provider.Issuer != nil:
		issuer = *provider.Issuer
	case provider.IssuerRef != nil:
		if issuer, err = t.getOpaqueSecretValue(crossNamespaceFrom{
			group:     egv1a1.GroupName,
			kind:      egv1a1.KindSecurityPolicy,
			namespace: policy.Namespace,
		}, provider.IssuerRef.Name, string(ptr.Deref(provider.IssuerRef.Namespace, gwapiv1.Namespace(policy.Namespace))), egv1a1.OIDCIssuerKey); err != nil {
			return nil, perr.WithMessage(err, "issuerRef")
		}
	}

	var u *url.URL
	// Use resolved issuer if token endpoint is not provided
	if provider.TokenEndpoint != nil || provider.TokenEndpointRef != nil {
		var tokenEndpointStr string
		switch {
		case provider.TokenEndpoint != nil:
			tokenEndpointStr = *provider.TokenEndpoint
		case provider.TokenEndpointRef != nil:
			if tokenEndpointStr, err = t.getOpaqueSecretValue(crossNamespaceFrom{
				group:     egv1a1.GroupName,
				kind:      egv1a1.KindSecurityPolicy,
				namespace: policy.Namespace,
			}, provider.TokenEndpointRef.Name, string(ptr.Deref(provider.TokenEndpointRef.Namespace, gwapiv1.Namespace(policy.Namespace))), egv1a1.OIDCTokenEndpointKey); err != nil {
				return nil, perr.WithMessage(err, "tokenEndpointRef")
			}
		}
		u, err = url.Parse(tokenEndpointStr)
	} else {
		u, err = url.Parse(issuer)
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
		if rd, err = t.translateExtServiceBackendRefs(
			policy, provider.BackendRefs, protocol, resources, envoyProxy, "oidc", 0); err != nil {
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
	var (
		userProvidedAuthorizationEndpoint string
		userProvidedTokenEndpoint         string
		userProvidedEndSessionEndpoint    string
	)

	switch {
	case provider.AuthorizationEndpoint != nil:
		userProvidedAuthorizationEndpoint = *provider.AuthorizationEndpoint
	case provider.AuthorizationEndpointRef != nil:
		if userProvidedAuthorizationEndpoint, err = t.getOpaqueSecretValue(crossNamespaceFrom{
			group:     egv1a1.GroupName,
			kind:      egv1a1.KindSecurityPolicy,
			namespace: policy.Namespace,
		}, provider.AuthorizationEndpointRef.Name, string(ptr.Deref(provider.AuthorizationEndpointRef.Namespace, gwapiv1.Namespace(policy.Namespace))), egv1a1.OIDCAuthorizationEndpointKey); err != nil {
			return nil, perr.WithMessage(err, "authorizationEndpointRef")
		}
	}

	switch {
	case provider.TokenEndpoint != nil:
		userProvidedTokenEndpoint = *provider.TokenEndpoint
	case provider.TokenEndpointRef != nil:
		if userProvidedTokenEndpoint, err = t.getOpaqueSecretValue(crossNamespaceFrom{
			group:     egv1a1.GroupName,
			kind:      egv1a1.KindSecurityPolicy,
			namespace: policy.Namespace,
		}, provider.TokenEndpointRef.Name, string(ptr.Deref(provider.TokenEndpointRef.Namespace, gwapiv1.Namespace(policy.Namespace))), egv1a1.OIDCTokenEndpointKey); err != nil {
			return nil, perr.WithMessage(err, "tokenEndpointRef")
		}
	}

	switch {
	case provider.EndSessionEndpoint != nil:
		userProvidedEndSessionEndpoint = *provider.EndSessionEndpoint
	case provider.EndSessionEndpointRef != nil:
		if userProvidedEndSessionEndpoint, err = t.getOpaqueSecretValue(crossNamespaceFrom{
			group:     egv1a1.GroupName,
			kind:      egv1a1.KindSecurityPolicy,
			namespace: policy.Namespace,
		}, provider.EndSessionEndpointRef.Name, string(ptr.Deref(provider.EndSessionEndpointRef.Namespace, gwapiv1.Namespace(policy.Namespace))), egv1a1.OIDCEndSessionEndpointKey); err != nil {
			return nil, perr.WithMessage(err, "endSessionEndpointRef")
		}
	}

	// Authorization endpoint and token endpoint are required fields.
	// If either of them is not provided, we need to fetch them from the issuer's well-known url.
	if userProvidedAuthorizationEndpoint == "" || userProvidedTokenEndpoint == "" {
		// Fetch the endpoints from the issuer's well-known url.
		discoveredConfig, err := t.fetchEndpointsFromIssuer(issuer, providerTLS)
		if err != nil {
			return nil, err
		}

		// Prioritize using the explicitly provided authorization endpoints if available.
		// This allows users to add extra parameters to the authorization endpoint if needed.
		if userProvidedAuthorizationEndpoint != "" {
			authorizationEndpoint = userProvidedAuthorizationEndpoint
		} else {
			authorizationEndpoint = discoveredConfig.AuthorizationEndpoint
		}

		// Prioritize using the explicitly provided token endpoints if available.
		// This may not be necessary, but we do it for consistency with authorization endpoint.
		if userProvidedTokenEndpoint != "" {
			tokenEndpoint = userProvidedTokenEndpoint
		} else {
			tokenEndpoint = discoveredConfig.TokenEndpoint
		}

		// Prioritize using the explicitly provided end session endpoints if available.
		// This may not be necessary, but we do it for consistency with other endpoints.
		if userProvidedEndSessionEndpoint != "" {
			endSessionEndpoint = &userProvidedEndSessionEndpoint
		} else {
			endSessionEndpoint = discoveredConfig.EndSessionEndpoint
		}
	} else {
		tokenEndpoint = userProvidedTokenEndpoint
		authorizationEndpoint = userProvidedAuthorizationEndpoint
		if userProvidedEndSessionEndpoint != "" {
			endSessionEndpoint = &userProvidedEndSessionEndpoint
		}
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

// getObjectValueFromRef assumes the local object reference points to
// a Kubernetes ConfigMap or Secret.
func (t *Translator) getObjectValueFromRef(
	valueRef *egv1a1.LocalObjectKeyReference,
	policyNs string,
) (string, error) {
	if valueRef == nil {
		return "", errors.New("unexpected nil reference")
	}

	switch valueRef.Kind {
	case resource.KindConfigMap:
		cm := t.GetConfigMap(policyNs, string(valueRef.Name))
		if cm != nil {
			s, dataOk := cm.Data[valueRef.Key]
			if !dataOk {
				return "", fmt.Errorf("can't find the key %q in the referenced configmap %q", valueRef.Key, valueRef.Name)
			}
			return s, nil
		}
		return "", fmt.Errorf("can't find the referenced configmap %q in namespace %q", valueRef.Name, policyNs)
	case resource.KindSecret:
		sec := t.GetSecret(policyNs, string(valueRef.Name))
		if sec != nil {
			b, dataOk := sec.Data[valueRef.Key]
			if !dataOk {
				return "", fmt.Errorf("can't find the key %q in the referenced secret %q", valueRef.Key, valueRef.Name)
			}
			return string(b), nil
		}
		return "", fmt.Errorf("can't find the referenced secret %q in namespace %q", valueRef.Name, policyNs)
	}
	return "", fmt.Errorf("unexpected reference to kind %q", valueRef.Kind)
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
//  is a required scope for OIDC.
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
	TokenEndpoint         string  
	AuthorizationEndpoint string  
	EndSessionEndpoint    *string 
}

func (o *OpenIDConfig) validate() error {
	if o.TokenEndpoint == "" {
		return errors.New("token_endpoint not found in OpenID configuration")
	}
	if o.AuthorizationEndpoint == "" {
		return errors.New("authorization_endpoint not found in OpenID configuration")
	}
	return nil
}

func (t *Translator) fetchEndpointsFromIssuer(issuerURL string, providerTLS *ir.TLSUpstreamConfig) (*OpenIDConfig, error) {
	if config, cachedErr, ok := t.oidcDiscoveryCache.Get(issuerURL); ok {
		if cachedErr != nil {
			return nil, cachedErr
		}
		return config, nil
	}

	config, err := discoverEndpointsFromIssuer(issuerURL, providerTLS)
	if err != nil {
		t.oidcDiscoveryCache.Set(issuerURL, nil, err)
		return nil, err
	}

	// Validate the discovered configuration
	if err = config.validate(); err != nil {
		t.oidcDiscoveryCache.Set(issuerURL, nil, err)
		return nil, err
	}

	t.oidcDiscoveryCache.Set(issuerURL, config, nil)
	return config, nil
}

// discoverEndpointsFromIssuer discovers the token and authorization endpoints from the issuer's well-known url.
func discoverEndpointsFromIssuer(issuerURL string, providerTLS *ir.TLSUpstreamConfig) (*OpenIDConfig, error) {
	// The well-known configuration endpoint is at {issuer}/.well-known/openid-configuration
	// https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationRequest
	wellKnownURL, err := url.JoinPath(issuerURL, "/.well-known/openid-configuration")
	if err != nil {
		return nil, err
	}

	// Create a new client for fetching the well-known configuration.
	// We need to create a new client because we may need to configure custom CA for the issuer.
	client := http.Client{
		Timeout: defaultOIDCHTTPTimeout,
	}

	// Configure TLS if needed
	if providerTLS != nil {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		if len(providerTLS.CACertificate) > 0 {
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(providerTLS.CACertificate); !ok {
				return nil, errors.New("failed to append CA certificate")
			}
			tlsConfig.RootCAs = caCertPool
		}
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
		}
	}

	req, err := http.NewRequest("GET", wellKnownURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch openid configuration from %s: %s", wellKnownURL, resp.Status)
	}

	var config OpenIDConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateTokenEndpoint(tokenEndpoint string) error {
	u, err := url.Parse(tokenEndpoint)
	if err != nil {
		return err
	}

	if u.Scheme != "https" && u.Scheme != "http" {
		return fmt.Errorf("invalid token endpoint %s: scheme must be http or https", tokenEndpoint)
	}

	if u.Host == "" {
		return fmt.Errorf("invalid token endpoint %s: host must be specified", tokenEndpoint)
	}

	return nil
}
