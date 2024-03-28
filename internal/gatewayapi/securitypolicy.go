// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"sort"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/status"
	"github.com/envoyproxy/gateway/internal/utils"
)

const (
	defaultRedirectURL  = "%REQ(x-forwarded-proto)%://%REQ(:authority)%/oauth2/callback"
	defaultRedirectPath = "/oauth2/callback"
	defaultLogoutPath   = "/logout"

	// nolint: gosec
	oidcHMACSecretName = "envoy-oidc-hmac"
	oidcHMACSecretKey  = "hmac-secret"
)

func (t *Translator) ProcessSecurityPolicies(securityPolicies []*egv1a1.SecurityPolicy,
	gateways []*GatewayContext,
	routes []RouteContext,
	resources *Resources,
	xdsIR XdsIRMap) []*egv1a1.SecurityPolicy {
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
		gatewayMap[key] = &policyGatewayTargetContext{GatewayContext: gw}
	}

	// Map of Gateway to the routes attached to it
	gatewayRouteMap := make(map[string]sets.Set[string])

	// Translate
	// 1. First translate Policies targeting xRoutes
	// 2. Finally, the policies targeting Gateways

	// Process the policies targeting xRoutes
	for _, policy := range securityPolicies {
		if policy.Spec.TargetRef.Kind != KindGateway {
			var (
				policy         = policy.DeepCopy()
				targetedRoute  RouteContext
				parentGateways []gwv1a2.ParentReference
				resolveErr     *status.PolicyResolveError
			)

			res = append(res, policy)

			targetedRoute, resolveErr = resolveSecurityPolicyRouteTargetRef(policy, routeMap)
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
				if p.Kind == nil || *p.Kind == KindGateway {
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
						gatewayRouteMap[key] = make(sets.Set[string])
					}
					gatewayRouteMap[key].Insert(utils.NamespacedName(targetedRoute).String())
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

	// Process the policies targeting Gateways
	for _, policy := range securityPolicies {
		if policy.Spec.TargetRef.Kind == KindGateway {
			var (
				policy          = policy.DeepCopy()
				targetedGateway *GatewayContext
				resolveErr      *status.PolicyResolveError
			)

			res = append(res, policy)

			targetedGateway, resolveErr = resolveSecurityPolicyGatewayTargetRef(policy, gatewayMap)
			// Skip if the gateway is not found
			// It's not necessarily an error because the SecurityPolicy may be
			// reconciled by multiple controllers. And the other controller may
			// have the target gateway.
			if targetedGateway == nil {
				continue
			}

			// Find its ancestor reference by resolved gateway, even with resolve error
			gatewayNN := utils.NamespacedName(targetedGateway)
			parentGateways := []gwv1a2.ParentReference{
				getAncestorRefForPolicy(gatewayNN, nil),
			}

			// Set conditions for resolve error, then skip current gateway
			if resolveErr != nil {
				status.SetResolveErrorForPolicyAncestors(&policy.Status,
					parentGateways,
					t.GatewayControllerName,
					policy.Generation,
					resolveErr,
				)

				continue
			}

			if err := t.translateSecurityPolicyForGateway(policy, targetedGateway, resources, xdsIR); err != nil {
				status.SetTranslationErrorForPolicyAncestors(&policy.Status,
					parentGateways,
					t.GatewayControllerName,
					policy.Generation,
					status.Error2ConditionMsg(err),
				)
			}

			// Set Accepted condition if it is unset
			status.SetAcceptedForPolicyAncestors(&policy.Status, parentGateways, t.GatewayControllerName)

			// Check if this policy is overridden by other policies targeting
			// at route level
			if r, ok := gatewayRouteMap[gatewayNN.String()]; ok {
				// Maintain order here to ensure status/string does not change with the same data
				routes := r.UnsortedList()
				sort.Strings(routes)
				message := fmt.Sprintf(
					"This policy is being overridden by other securityPolicies for these routes: %v",
					routes)
				status.SetConditionForPolicyAncestors(&policy.Status,
					parentGateways,
					t.GatewayControllerName,
					egv1a1.PolicyConditionOverridden,
					metav1.ConditionTrue,
					egv1a1.PolicyReasonOverridden,
					message,
					policy.Generation,
				)
			}
		}
	}

	return res
}

func resolveSecurityPolicyGatewayTargetRef(
	policy *egv1a1.SecurityPolicy,
	gateways map[types.NamespacedName]*policyGatewayTargetContext) (*GatewayContext, *status.PolicyResolveError) {
	targetNs := policy.Spec.TargetRef.Namespace
	// If empty, default to namespace of policy
	if targetNs == nil {
		targetNs = ptr.To(gwv1b1.Namespace(policy.Namespace))
	}

	// Find the Gateway
	key := types.NamespacedName{
		Name:      string(policy.Spec.TargetRef.Name),
		Namespace: string(*targetNs),
	}
	gateway, ok := gateways[key]

	// Gateway not found
	// It's not an error if the gateway is not found because the SecurityPolicy
	// may be reconciled by multiple controllers, and the gateway may not be managed
	// by this controller.
	if !ok {
		return nil, nil
	}

	// Ensure Policy and target are in the same namespace
	if policy.Namespace != string(*targetNs) {
		// TODO zhaohuabing use CEL to validate cross-namespace reference
		message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, SecurityPolicy can only target a resource in the same namespace.",
			policy.Namespace, *targetNs)

		return gateway.GatewayContext, &status.PolicyResolveError{
			Reason:  gwv1a2.PolicyReasonInvalid,
			Message: message,
		}
	}

	// Check if another policy targeting the same Gateway exists
	if gateway.attached {
		message := "Unable to target Gateway, another SecurityPolicy has already attached to it"

		return gateway.GatewayContext, &status.PolicyResolveError{
			Reason:  gwv1a2.PolicyReasonConflicted,
			Message: message,
		}
	}

	// Set context and save
	gateway.attached = true
	gateways[key] = gateway

	return gateway.GatewayContext, nil
}

func resolveSecurityPolicyRouteTargetRef(
	policy *egv1a1.SecurityPolicy,
	routes map[policyTargetRouteKey]*policyRouteTargetContext) (RouteContext, *status.PolicyResolveError) {
	targetNs := policy.Spec.TargetRef.Namespace
	// If empty, default to namespace of policy
	if targetNs == nil {
		targetNs = ptr.To(gwv1b1.Namespace(policy.Namespace))
	}

	// Check if the route exists
	key := policyTargetRouteKey{
		Kind:      string(policy.Spec.TargetRef.Kind),
		Name:      string(policy.Spec.TargetRef.Name),
		Namespace: string(*targetNs),
	}
	route, ok := routes[key]

	// Route not found
	// It's not an error if the gateway is not found because the SecurityPolicy
	// may be reconciled by multiple controllers, and the gateway may not be managed
	// by this controller.
	if !ok {
		return nil, nil
	}

	// Ensure Policy and target are in the same namespace
	// TODO zhaohuabing use CEL to validate cross-namespace reference
	if policy.Namespace != string(*targetNs) {
		message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, SecurityPolicy can only target a resource in the same namespace.",
			policy.Namespace, *targetNs)

		return route.RouteContext, &status.PolicyResolveError{
			Reason:  gwv1a2.PolicyReasonInvalid,
			Message: message,
		}
	}

	// Check if another policy targeting the same xRoute exists
	if route.attached {
		message := fmt.Sprintf("Unable to target %s, another SecurityPolicy has already attached to it",
			string(policy.Spec.TargetRef.Kind))

		return route.RouteContext, &status.PolicyResolveError{
			Reason:  gwv1a2.PolicyReasonConflicted,
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
	resources *Resources, xdsIR XdsIRMap) error {
	// Build IR
	var (
		cors      *ir.CORS
		jwt       *ir.JWT
		oidc      *ir.OIDC
		basicAuth *ir.BasicAuth
		extAuth   *ir.ExtAuth
		err, errs error
	)

	if policy.Spec.CORS != nil {
		cors = t.buildCORS(policy.Spec.CORS)
	}

	if policy.Spec.JWT != nil {
		jwt = t.buildJWT(policy.Spec.JWT)
	}

	if policy.Spec.OIDC != nil {
		if oidc, err = t.buildOIDC(
			policy,
			resources); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.BasicAuth != nil {
		if basicAuth, err = t.buildBasicAuth(
			policy,
			resources); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.ExtAuth != nil {
		if extAuth, err = t.buildExtAuth(
			policy,
			resources); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	// Apply IR to all relevant routes
	// Note: there are multiple features in a security policy, even if some of them
	// are invalid, we still want to apply the valid ones.
	prefix := irRoutePrefix(route)
	for _, ir := range xdsIR {
		for _, http := range ir.HTTP {
			for _, r := range http.Routes {
				// Apply if there is a match
				// TODO zhaohuabing: extract a utils function to check if an HTTP
				// route is associated with a Gateway API xRoute
				if strings.HasPrefix(r.Name, prefix) {
					// This security policy matches the current route. It should only be accepted if it doesn't match any other route
					r.CORS = cors
					r.JWT = jwt
					r.OIDC = oidc
					r.BasicAuth = basicAuth
					r.ExtAuth = extAuth
				}
			}
		}
	}
	return errs
}

func (t *Translator) translateSecurityPolicyForGateway(
	policy *egv1a1.SecurityPolicy, gateway *GatewayContext,
	resources *Resources, xdsIR XdsIRMap) error {
	// Build IR
	var (
		cors      *ir.CORS
		jwt       *ir.JWT
		oidc      *ir.OIDC
		basicAuth *ir.BasicAuth
		extAuth   *ir.ExtAuth
		err, errs error
	)

	if policy.Spec.CORS != nil {
		cors = t.buildCORS(policy.Spec.CORS)
	}

	if policy.Spec.JWT != nil {
		jwt = t.buildJWT(policy.Spec.JWT)
	}

	if policy.Spec.OIDC != nil {
		if oidc, err = t.buildOIDC(
			policy,
			resources); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.BasicAuth != nil {
		if basicAuth, err = t.buildBasicAuth(
			policy,
			resources); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.ExtAuth != nil {
		if extAuth, err = t.buildExtAuth(
			policy,
			resources); err != nil {
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
	ir := xdsIR[irKey]

	policyTarget := irStringKey(
		string(ptr.Deref(policy.Spec.TargetRef.Namespace, gwv1a2.Namespace(policy.Namespace))),
		string(policy.Spec.TargetRef.Name),
	)
	for _, http := range ir.HTTP {
		gatewayName := http.Name[0:strings.LastIndex(http.Name, "/")]
		if t.MergeGateways && gatewayName != policyTarget {
			continue
		}
		// A Policy targeting the most specific scope(xRoute) wins over a policy
		// targeting a lesser specific scope(Gateway).
		for _, r := range http.Routes {
			// If any of the features are already set, it means that a more specific
			// policy(targeting xRoute) has already set it, so we skip it.
			// TODO: zhaohuabing group the features into a struct and check if all of them are set
			if r.CORS != nil ||
				r.JWT != nil ||
				r.OIDC != nil ||
				r.BasicAuth != nil ||
				r.ExtAuth != nil {
				continue
			}
			if r.CORS == nil {
				r.CORS = cors
			}
			if r.JWT == nil {
				r.JWT = jwt
			}
			if r.OIDC == nil {
				r.OIDC = oidc
			}
			if r.BasicAuth == nil {
				r.BasicAuth = basicAuth
			}
			if r.ExtAuth == nil {
				r.ExtAuth = extAuth
			}
		}
	}
	return errs
}

func (t *Translator) buildCORS(cors *egv1a1.CORS) *ir.CORS {
	var allowOrigins []*ir.StringMatch

	for _, origin := range cors.AllowOrigins {
		origin := origin
		if isWildcard(string(origin)) {
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

func isWildcard(s string) bool {
	return strings.ContainsAny(s, "*")
}

func wildcard2regex(wildcard string) string {
	regexStr := strings.ReplaceAll(wildcard, ".", "\\.")
	regexStr = strings.ReplaceAll(regexStr, "*", ".*")
	return regexStr
}

func (t *Translator) buildJWT(jwt *egv1a1.JWT) *ir.JWT {
	return &ir.JWT{
		Providers: jwt.Providers,
	}
}

func (t *Translator) buildOIDC(
	policy *egv1a1.SecurityPolicy,
	resources *Resources) (*ir.OIDC, error) {
	var (
		oidc         = policy.Spec.OIDC
		clientSecret *v1.Secret
		provider     *ir.OIDCProvider
		err          error
	)

	from := crossNamespaceFrom{
		group:     egv1a1.GroupName,
		kind:      KindSecurityPolicy,
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

	// Discover the token and authorization endpoints from the issuer's
	// well-known url if not explicitly specified
	if provider, err = discoverEndpointsFromIssuer(&oidc.Provider); err != nil {
		return nil, err
	}

	if err = validateTokenEndpoint(provider.TokenEndpoint); err != nil {
		return nil, err
	}
	scopes := appendOpenidScopeIfNotExist(oidc.Scopes)

	var (
		redirectURL  = defaultRedirectURL
		redirectPath = defaultRedirectPath
		logoutPath   = defaultLogoutPath
	)

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

	// Generate a unique cookie suffix for oauth filters
	suffix := utils.Digest32(string(policy.UID))

	// Get the HMAC secret
	// HMAC secret is generated by the CertGen job and stored in a secret
	// We need to rotate the HMAC secret in the future, probably the same
	// way we rotate the certs generated by the CertGen job.
	hmacSecret := resources.GetSecret(t.Namespace, oidcHMACSecretName)
	if hmacSecret == nil {
		return nil, fmt.Errorf("HMAC secret %s/%s not found", t.Namespace, oidcHMACSecretName)
	}
	hmacData, ok := hmacSecret.Data[oidcHMACSecretKey]
	if !ok || len(hmacData) == 0 {
		return nil, fmt.Errorf(
			"HMAC secret not found in secret %s/%s", t.Namespace, oidcHMACSecretName)
	}

	return &ir.OIDC{
		Name:         irConfigName(policy),
		Provider:     *provider,
		ClientID:     oidc.ClientID,
		ClientSecret: clientSecretBytes,
		Scopes:       scopes,
		Resources:    oidc.Resources,
		RedirectURL:  redirectURL,
		RedirectPath: redirectPath,
		LogoutPath:   logoutPath,
		CookieSuffix: suffix,
		HMACSecret:   hmacData,
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

// discoverEndpointsFromIssuer discovers the token and authorization endpoints from the issuer's well-known url
// return error if failed to fetch the well-known configuration
func discoverEndpointsFromIssuer(provider *egv1a1.OIDCProvider) (*ir.OIDCProvider, error) {
	if provider.TokenEndpoint == nil || provider.AuthorizationEndpoint == nil {
		tokenEndpoint, authorizationEndpoint, err := fetchEndpointsFromIssuer(provider.Issuer)
		if err != nil {
			return nil, fmt.Errorf("error fetching endpoints from issuer: %w", err)
		}
		return &ir.OIDCProvider{
			TokenEndpoint:         tokenEndpoint,
			AuthorizationEndpoint: authorizationEndpoint,
		}, nil
	}

	return &ir.OIDCProvider{
		TokenEndpoint:         *provider.TokenEndpoint,
		AuthorizationEndpoint: *provider.AuthorizationEndpoint,
	}, nil
}

func fetchEndpointsFromIssuer(issuerURL string) (string, string, error) {
	// Fetch the OpenID configuration from the issuer URL
	resp, err := http.Get(fmt.Sprintf("%s/.well-known/openid-configuration", issuerURL))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	// Parse the OpenID configuration response
	var config OpenIDConfig
	err = json.NewDecoder(resp.Body).Decode(&config)
	if err != nil {
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

func (t *Translator) buildBasicAuth(
	policy *egv1a1.SecurityPolicy,
	resources *Resources) (*ir.BasicAuth, error) {
	var (
		basicAuth   = policy.Spec.BasicAuth
		usersSecret *v1.Secret
		err         error
	)

	from := crossNamespaceFrom{
		group:     egv1a1.GroupName,
		kind:      KindSecurityPolicy,
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
		Name:  irConfigName(policy),
		Users: usersSecretBytes,
	}, nil
}

func (t *Translator) buildExtAuth(
	policy *egv1a1.SecurityPolicy,
	resources *Resources) (*ir.ExtAuth, error) {
	var (
		http       = policy.Spec.ExtAuth.HTTP
		grpc       = policy.Spec.ExtAuth.GRPC
		backendRef *gwapiv1.BackendObjectReference
		protocol   ir.AppProtocol
		ds         *ir.DestinationSetting
		authority  string
		err        error
	)

	switch {
	// These are sanity checks, they should never happen because the API server
	// should have caught them
	case http == nil && grpc == nil:
		return nil, errors.New("one of grpc or http must be specified")
	case http != nil && grpc != nil:
		return nil, errors.New("only one of grpc or http can be specified")
	case http != nil:
		backendRef = &http.BackendRef
		protocol = ir.HTTP
	case grpc != nil:
		backendRef = &grpc.BackendRef
		protocol = ir.GRPC
	}

	if err = t.validateExtServiceBackendReference(
		backendRef,
		policy.Namespace,
		resources); err != nil {
		return nil, err
	}
	authority = fmt.Sprintf(
		"%s.%s:%d",
		backendRef.Name,
		NamespaceDerefOr(backendRef.Namespace, policy.Namespace),
		*backendRef.Port)

	if ds, err = t.processExtServiceDestination(
		backendRef,
		policy,
		protocol,
		resources); err != nil {
		return nil, err
	}
	rd := ir.RouteDestination{
		Name:     irExtServiceDestinationName(policy, backendRef),
		Settings: []*ir.DestinationSetting{ds},
	}

	extAuth := &ir.ExtAuth{
		Name:             irConfigName(policy),
		HeadersToExtAuth: policy.Spec.ExtAuth.HeadersToExtAuth,
		FailOpen:         policy.Spec.ExtAuth.FailOpen,
	}

	if http != nil {
		extAuth.HTTP = &ir.HTTPExtAuthService{
			Destination:      rd,
			Authority:        authority,
			Path:             ptr.Deref(http.Path, ""),
			HeadersToBackend: http.HeadersToBackend,
		}
	} else {
		extAuth.GRPC = &ir.GRPCExtAuthService{
			Destination: rd,
			Authority:   authority,
		}
	}
	return extAuth, nil
}

// TODO: zhaohuabing combine this function with the one in the route translator
func (t *Translator) processExtServiceDestination(
	backendRef *gwapiv1.BackendObjectReference,
	policy *egv1a1.SecurityPolicy,
	protocol ir.AppProtocol,
	resources *Resources) (*ir.DestinationSetting, error) {
	var (
		endpoints   []*ir.DestinationEndpoint
		addrType    *ir.DestinationAddressType
		servicePort v1.ServicePort
		backendTLS  *ir.TLSUpstreamConfig
	)

	serviceNamespace := NamespaceDerefOr(backendRef.Namespace, policy.Namespace)
	service := resources.GetService(serviceNamespace, string(backendRef.Name))
	for _, port := range service.Spec.Ports {
		if port.Port == int32(*backendRef.Port) {
			servicePort = port
			break
		}
	}

	if servicePort.AppProtocol != nil &&
		*servicePort.AppProtocol == "kubernetes.io/h2c" {
		protocol = ir.HTTP2
	}

	// Route to endpoints by default
	if !t.EndpointRoutingDisabled {
		endpointSlices := resources.GetEndpointSlicesForBackend(
			serviceNamespace, string(backendRef.Name), KindService)
		endpoints, addrType = getIREndpointsFromEndpointSlices(
			endpointSlices, servicePort.Name, servicePort.Protocol)
	} else {
		// Fall back to Service ClusterIP routing
		ep := ir.NewDestEndpoint(
			service.Spec.ClusterIP,
			uint32(*backendRef.Port))
		endpoints = append(endpoints, ep)
	}

	// TODO: support mixed endpointslice address type for the same backendRef
	if !t.EndpointRoutingDisabled && addrType != nil && *addrType == ir.MIXED {
		return nil, errors.New(
			"mixed endpointslice address type for the same backendRef is not supported")
	}

	backendTLS = t.processBackendTLSPolicy(
		*backendRef,
		serviceNamespace,
		// Gateway is not the appropriate parent reference here because the owner
		// of the BackendRef is the security policy, and there is no hierarchy
		// relationship between the security policy and a gateway.
		// The owner security policy of the BackendRef is used as the parent reference here.
		gwv1a2.ParentReference{
			Group:     ptr.To(gwapiv1.Group(egv1a1.GroupName)),
			Kind:      ptr.To(gwapiv1.Kind(egv1a1.KindSecurityPolicy)),
			Namespace: ptr.To(gwapiv1.Namespace(policy.Namespace)),
			Name:      gwapiv1.ObjectName(policy.Name),
		},
		resources)

	return &ir.DestinationSetting{
		Weight:      ptr.To(uint32(1)),
		Protocol:    protocol,
		Endpoints:   endpoints,
		AddressType: addrType,
		TLS:         backendTLS,
	}, nil
}

func irExtServiceDestinationName(policy *egv1a1.SecurityPolicy, backendRef *gwapiv1.BackendObjectReference) string {
	nn := types.NamespacedName{
		Name:      string(backendRef.Name),
		Namespace: NamespaceDerefOr(backendRef.Namespace, policy.Namespace),
	}

	return strings.ToLower(fmt.Sprintf(
		"%s/%s",
		irConfigName(policy),
		nn.String()))
}

func irConfigName(policy *egv1a1.SecurityPolicy) string {
	return fmt.Sprintf(
		"%s/%s",
		strings.ToLower(KindSecurityPolicy),
		utils.NamespacedName(policy).String())
}
