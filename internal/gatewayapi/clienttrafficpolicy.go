// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	gwv1b1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/status"
	"github.com/envoyproxy/gateway/internal/utils"
)

const (
	// Use an invalid string to represent all sections (listeners) within a Gateway
	AllSections = "/"
)

func hasSectionName(policy *egv1a1.ClientTrafficPolicy) bool {
	return policy.Spec.TargetRef.SectionName != nil
}

func (t *Translator) ProcessClientTrafficPolicies(resources *Resources,
	gateways []*GatewayContext,
	xdsIR XdsIRMap, infraIR InfraIRMap) []*egv1a1.ClientTrafficPolicy {
	var res []*egv1a1.ClientTrafficPolicy

	clientTrafficPolicies := resources.ClientTrafficPolicies
	// Sort based on timestamp
	sort.Slice(clientTrafficPolicies, func(i, j int) bool {
		return clientTrafficPolicies[i].CreationTimestamp.Before(&(clientTrafficPolicies[j].CreationTimestamp))
	})

	policyMap := make(map[types.NamespacedName]sets.Set[string])

	// Build a map out of gateways for faster lookup since users might have hundreds of gateway or more.
	gatewayMap := map[types.NamespacedName]*policyGatewayTargetContext{}
	for _, gw := range gateways {
		key := utils.NamespacedName(gw)
		gatewayMap[key] = &policyGatewayTargetContext{GatewayContext: gw}
	}

	// Translate
	// 1. First translate Policies with a sectionName set
	// 2. Then loop again and translate the policies without a sectionName
	// TODO: Import sort order to ensure policy with same section always appear
	// before policy with no section so below loops can be flattened into 1.

	for _, policy := range clientTrafficPolicies {
		if hasSectionName(policy) {
			policy := policy.DeepCopy()
			res = append(res, policy)

			gateway, resolveErr := resolveCTPolicyTargetRef(policy, gatewayMap)

			// Negative statuses have already been assigned so its safe to skip
			if gateway == nil {
				continue
			}

			key := utils.NamespacedName(gateway)
			ancestorRefs := []gwv1a2.ParentReference{
				getAncestorRefForPolicy(key, policy.Spec.TargetRef.SectionName),
			}

			// Set conditions for resolve error, then skip current gateway
			if resolveErr != nil {
				status.SetResolveErrorForPolicyAncestors(&policy.Status,
					ancestorRefs,
					t.GatewayControllerName,
					policy.Generation,
					resolveErr,
				)

				continue
			}

			// Check if another policy targeting the same section exists
			section := string(*(policy.Spec.TargetRef.SectionName))
			s, ok := policyMap[key]
			if ok && s.Has(section) {
				message := "Unable to target section, another ClientTrafficPolicy has already attached to it"

				resolveErr = &status.PolicyResolveError{
					Reason:  gwv1a2.PolicyReasonConflicted,
					Message: message,
				}

				status.SetResolveErrorForPolicyAncestors(&policy.Status,
					ancestorRefs,
					t.GatewayControllerName,
					policy.Generation,
					resolveErr,
				)

				continue
			}

			// Add section to policy map
			if s == nil {
				policyMap[key] = sets.New[string]()
			}
			policyMap[key].Insert(section)

			// Translate for listener matching section name
			var err error
			for _, l := range gateway.listeners {
				// Find IR
				irKey := t.getIRKey(l.gateway)
				// It must exist since we've already finished processing the gateways
				gwXdsIR := xdsIR[irKey]
				if string(l.Name) == section {
					err = validatePortOverlapForClientTrafficPolicy(l, gwXdsIR, false)
					if err == nil {
						err = t.translateClientTrafficPolicyForListener(policy, l, xdsIR, infraIR, resources)
					}
					break
				}
			}

			// Set conditions for translation error if it got any
			if err != nil {
				status.SetTranslationErrorForPolicyAncestors(&policy.Status,
					ancestorRefs,
					t.GatewayControllerName,
					policy.Generation,
					status.Error2ConditionMsg(err),
				)
			}

			// Set Accepted condition if it is unset
			status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName)
		}
	}

	// Policy with no section set (targeting all sections)
	for _, policy := range clientTrafficPolicies {
		if !hasSectionName(policy) {

			policy := policy.DeepCopy()
			res = append(res, policy)

			gateway, resolveErr := resolveCTPolicyTargetRef(policy, gatewayMap)

			// Negative statuses have already been assigned so its safe to skip
			if gateway == nil {
				continue
			}

			key := utils.NamespacedName(gateway)
			ancestorRefs := []gwv1a2.ParentReference{
				getAncestorRefForPolicy(key, nil),
			}

			// Set conditions for resolve error, then skip current gateway
			if resolveErr != nil {
				status.SetResolveErrorForPolicyAncestors(&policy.Status,
					ancestorRefs,
					t.GatewayControllerName,
					policy.Generation,
					resolveErr,
				)

				continue
			}

			// Check if another policy targeting the same Gateway exists
			s, ok := policyMap[key]
			if ok && s.Has(AllSections) {
				message := "Unable to target Gateway, another ClientTrafficPolicy has already attached to it"

				resolveErr = &status.PolicyResolveError{
					Reason:  gwv1a2.PolicyReasonConflicted,
					Message: message,
				}

				status.SetResolveErrorForPolicyAncestors(&policy.Status,
					ancestorRefs,
					t.GatewayControllerName,
					policy.Generation,
					resolveErr,
				)

				continue
			}

			// Check if another policy targeting the same Gateway exists
			if ok && (s.Len() > 0) {
				// Maintain order here to ensure status/string does not change with same data
				sections := s.UnsortedList()
				sort.Strings(sections)
				message := fmt.Sprintf("There are existing ClientTrafficPolicies that are overriding these sections %v", sections)

				status.SetConditionForPolicyAncestors(&policy.Status,
					ancestorRefs,
					t.GatewayControllerName,
					egv1a1.PolicyConditionOverridden,
					metav1.ConditionTrue,
					egv1a1.PolicyReasonOverridden,
					message,
					policy.Generation,
				)
			}

			// Add section to policy map
			if s == nil {
				policyMap[key] = sets.New[string]()
			}
			policyMap[key].Insert(AllSections)

			// Translate sections that have not yet been targeted
			var errs error
			for _, l := range gateway.listeners {
				// Skip if section has already been targeted
				if s != nil && s.Has(string(l.Name)) {
					continue
				}

				// Find IR
				irKey := t.getIRKey(l.gateway)
				// It must exist since we've already finished processing the gateways
				gwXdsIR := xdsIR[irKey]
				if err := validatePortOverlapForClientTrafficPolicy(l, gwXdsIR, true); err != nil {
					errs = errors.Join(errs, err)
				} else if err := t.translateClientTrafficPolicyForListener(policy, l, xdsIR, infraIR, resources); err != nil {
					errs = errors.Join(errs, err)
				}
			}

			// Set conditions for translation error if it got any
			if errs != nil {
				status.SetTranslationErrorForPolicyAncestors(&policy.Status,
					ancestorRefs,
					t.GatewayControllerName,
					policy.Generation,
					status.Error2ConditionMsg(errs),
				)
			}

			// Set Accepted condition if it is unset
			status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName)
		}
	}

	return res
}

func resolveCTPolicyTargetRef(policy *egv1a1.ClientTrafficPolicy, gateways map[types.NamespacedName]*policyGatewayTargetContext) (*GatewayContext, *status.PolicyResolveError) {
	targetNs := policy.Spec.TargetRef.Namespace
	// If empty, default to namespace of policy
	if targetNs == nil {
		targetNs = ptr.To(gwv1b1.Namespace(policy.Namespace))
	}

	// Check if the gateway exists
	key := types.NamespacedName{
		Name:      string(policy.Spec.TargetRef.Name),
		Namespace: string(*targetNs),
	}
	gateway, ok := gateways[key]

	// Gateway not found
	if !ok {
		return nil, nil
	}

	// Ensure Policy and target Gateway are in the same namespace
	if policy.Namespace != string(*targetNs) {
		message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, ClientTrafficPolicy can only target a Gateway in the same namespace.",
			policy.Namespace, *targetNs)

		return gateway.GatewayContext, &status.PolicyResolveError{
			Reason:  gwv1a2.PolicyReasonInvalid,
			Message: message,
		}
	}

	// If sectionName is set, make sure its valid
	sectionName := policy.Spec.TargetRef.SectionName
	if sectionName != nil {
		found := false
		for _, l := range gateway.listeners {
			if l.Name == *sectionName {
				found = true
				break
			}
		}
		if !found {
			message := fmt.Sprintf("No section name %s found for %s", *sectionName, key.String())

			return gateway.GatewayContext, &status.PolicyResolveError{
				Reason:  gwv1a2.PolicyReasonInvalid,
				Message: message,
			}
		}
	}

	return gateway.GatewayContext, nil
}

func validatePortOverlapForClientTrafficPolicy(l *ListenerContext, xds *ir.Xds, attachedToGateway bool) error {
	// Find Listener IR
	// TODO: Support TLSRoute and TCPRoute once
	// https://github.com/envoyproxy/gateway/issues/1635 is completed

	irListenerName := irHTTPListenerName(l)
	var httpIR *ir.HTTPListener
	for _, http := range xds.HTTP {
		if http.Name == irListenerName {
			httpIR = http
			break
		}
	}

	// IR must exist since we're past validation
	if httpIR != nil {
		// Get a list of all other non-TLS listeners on this Gateway that share a port with
		// the listener in question.
		if sameListeners := listenersWithSameHTTPPort(xds, httpIR); len(sameListeners) != 0 {
			if attachedToGateway {
				// If this policy is attached to an entire gateway and the mergeGateways feature
				// is turned on, validate that all the listeners affected by this policy originated
				// from the same Gateway resource. The name of the Gateway from which this listener
				// originated is part of the listener's name by construction.
				gatewayName := irListenerName[0:strings.LastIndex(irListenerName, "/")]
				conflictingListeners := []string{}
				for _, currName := range sameListeners {
					if strings.Index(currName, gatewayName) != 0 {
						conflictingListeners = append(conflictingListeners, currName)
					}
				}
				if len(conflictingListeners) != 0 {
					return fmt.Errorf("ClientTrafficPolicy is being applied to multiple http (non https) listeners (%s) on the same port, which is not allowed", strings.Join(conflictingListeners, ", "))
				}
			} else {
				// If this policy is attached to a specific listener, any other listeners in the list
				// would be affected by this policy but should not be, so this policy can't be accepted.
				return fmt.Errorf("ClientTrafficPolicy is being applied to multiple http (non https) listeners (%s) on the same port, which is not allowed", strings.Join(sameListeners, ", "))
			}
		}
	}
	return nil
}

func (t *Translator) translateClientTrafficPolicyForListener(policy *egv1a1.ClientTrafficPolicy, l *ListenerContext,
	xdsIR XdsIRMap, infraIR InfraIRMap, resources *Resources) error {
	// Find IR
	irKey := t.getIRKey(l.gateway)
	// It must exist since we've already finished processing the gateways
	gwXdsIR := xdsIR[irKey]

	// Find Listener IR
	// TODO: Support TLSRoute and TCPRoute once
	// https://github.com/envoyproxy/gateway/issues/1635 is completed

	irListenerName := irHTTPListenerName(l)
	var httpIR *ir.HTTPListener
	for _, http := range gwXdsIR.HTTP {
		if http.Name == irListenerName {
			httpIR = http
			break
		}
	}

	// IR must exist since we're past validation
	if httpIR != nil {
		// Translate TCPKeepalive
		translateListenerTCPKeepalive(policy.Spec.TCPKeepalive, httpIR)

		// Translate Connection
		if err := translateListenerConnection(policy.Spec.Connection, httpIR); err != nil {
			return err
		}

		// Translate Proxy Protocol
		translateListenerProxyProtocol(policy.Spec.EnableProxyProtocol, httpIR)

		// Translate Client IP Detection
		translateClientIPDetection(policy.Spec.ClientIPDetection, httpIR)

		// Translate Header Settings
		translateListenerHeaderSettings(policy.Spec.Headers, httpIR)

		// Translate Path Settings
		translatePathSettings(policy.Spec.Path, httpIR)

		// Translate Client Timeout Settings
		if err := translateClientTimeout(policy.Spec.Timeout, httpIR); err != nil {
			return err
		}

		// Translate HTTP1 Settings
		if err := translateHTTP1Settings(policy.Spec.HTTP1, httpIR); err != nil {
			return err
		}

		// enable http3 if set and TLS is enabled
		if httpIR.TLS != nil && policy.Spec.HTTP3 != nil {
			http3 := &ir.HTTP3Settings{
				QUICPort: int32(l.Port),
			}
			httpIR.HTTP3 = http3
			var proxyListenerIR *ir.ProxyListener
			for _, proxyListener := range infraIR[irKey].Proxy.Listeners {
				if proxyListener.Name == irListenerName {
					proxyListenerIR = proxyListener
					break
				}
			}
			if proxyListenerIR != nil {
				proxyListenerIR.HTTP3 = http3
			}
		}

		// Translate TLS parameters
		if err := t.translateListenerTLSParameters(policy, httpIR, resources); err != nil {
			return err
		}
	}
	return nil
}

func translateListenerTCPKeepalive(tcpKeepAlive *egv1a1.TCPKeepalive, httpIR *ir.HTTPListener) {
	// Return early if not set
	if tcpKeepAlive == nil {
		return
	}

	irTCPKeepalive := &ir.TCPKeepalive{}

	if tcpKeepAlive.Probes != nil {
		irTCPKeepalive.Probes = tcpKeepAlive.Probes
	}

	if tcpKeepAlive.IdleTime != nil {
		d, err := time.ParseDuration(string(*tcpKeepAlive.IdleTime))
		if err != nil {
			return
		}
		irTCPKeepalive.IdleTime = ptr.To(uint32(d.Seconds()))
	}

	if tcpKeepAlive.Interval != nil {
		d, err := time.ParseDuration(string(*tcpKeepAlive.Interval))
		if err != nil {
			return
		}
		irTCPKeepalive.Interval = ptr.To(uint32(d.Seconds()))
	}

	httpIR.TCPKeepalive = irTCPKeepalive
}

func translatePathSettings(pathSettings *egv1a1.PathSettings, httpIR *ir.HTTPListener) {
	if pathSettings == nil {
		return
	}
	if pathSettings.DisableMergeSlashes != nil {
		httpIR.Path.MergeSlashes = !*pathSettings.DisableMergeSlashes
	}
	if pathSettings.EscapedSlashesAction != nil {
		httpIR.Path.EscapedSlashesAction = ir.PathEscapedSlashAction(*pathSettings.EscapedSlashesAction)
	}
}

func translateClientTimeout(clientTimeout *egv1a1.ClientTimeout, httpIR *ir.HTTPListener) error {
	if clientTimeout == nil {
		return nil
	}

	irClientTimeout := &ir.ClientTimeout{}

	if clientTimeout.HTTP != nil {
		irHTTPTimeout := &ir.HTTPClientTimeout{}
		if clientTimeout.HTTP.RequestReceivedTimeout != nil {
			d, err := time.ParseDuration(string(*clientTimeout.HTTP.RequestReceivedTimeout))
			if err != nil {
				return err
			}
			irHTTPTimeout.RequestReceivedTimeout = &metav1.Duration{
				Duration: d,
			}
		}

		if clientTimeout.HTTP.IdleTimeout != nil {
			d, err := time.ParseDuration(string(*clientTimeout.HTTP.IdleTimeout))
			if err != nil {
				return err
			}
			irHTTPTimeout.IdleTimeout = &metav1.Duration{
				Duration: d,
			}
		}
		irClientTimeout.HTTP = irHTTPTimeout
	}

	httpIR.Timeout = irClientTimeout

	return nil
}

func translateListenerProxyProtocol(enableProxyProtocol *bool, httpIR *ir.HTTPListener) {
	// Return early if not set
	if enableProxyProtocol == nil {
		return
	}

	if *enableProxyProtocol {
		httpIR.EnableProxyProtocol = true
	}
}

func translateClientIPDetection(clientIPDetection *egv1a1.ClientIPDetectionSettings, httpIR *ir.HTTPListener) {
	// Return early if not set
	if clientIPDetection == nil {
		return
	}

	httpIR.ClientIPDetection = (*ir.ClientIPDetectionSettings)(clientIPDetection)
}

func translateListenerHeaderSettings(headerSettings *egv1a1.HeaderSettings, httpIR *ir.HTTPListener) {
	if headerSettings == nil {
		return
	}
	httpIR.Headers = &ir.HeaderSettings{
		EnableEnvoyHeaders:    ptr.Deref(headerSettings.EnableEnvoyHeaders, false),
		WithUnderscoresAction: ir.WithUnderscoresAction(ptr.Deref(headerSettings.WithUnderscoresAction, egv1a1.WithUnderscoresActionRejectRequest)),
	}
}

func translateHTTP1Settings(http1Settings *egv1a1.HTTP1Settings, httpIR *ir.HTTPListener) error {
	if http1Settings == nil {
		return nil
	}
	httpIR.HTTP1 = &ir.HTTP1Settings{
		EnableTrailers:     ptr.Deref(http1Settings.EnableTrailers, false),
		PreserveHeaderCase: ptr.Deref(http1Settings.PreserveHeaderCase, false),
	}
	if http1Settings.HTTP10 != nil {
		var defaultHost *string
		if ptr.Deref(http1Settings.HTTP10.UseDefaultHost, false) {
			for _, hostname := range httpIR.Hostnames {
				if !strings.Contains(hostname, "*") {
					// make linter happy
					theHost := hostname
					defaultHost = &theHost
					break
				}
			}
			if defaultHost == nil {
				return fmt.Errorf("can't set http10 default host on listener with only wildcard hostnames")
			}
		}
		// If useDefaultHost was set, then defaultHost will have the hostname to use.
		// If no good hostname was found, an error would have been returned.
		httpIR.HTTP1.HTTP10 = &ir.HTTP10Settings{
			DefaultHost: defaultHost,
		}
	}
	return nil
}

func (t *Translator) translateListenerTLSParameters(policy *egv1a1.ClientTrafficPolicy,
	httpIR *ir.HTTPListener, resources *Resources) error {
	// Return if this listener isn't a TLS listener. There has to be
	// at least one certificate defined, which would cause httpIR to
	// have a TLS structure.
	if httpIR.TLS == nil {
		return nil
	}

	tlsParams := policy.Spec.TLS

	// Make sure that the negotiated TLS protocol version is as expected if TLS is used,
	// regardless of if TLS parameters were used in the ClientTrafficPolicy or not
	httpIR.TLS.MinVersion = ptr.To(ir.TLSv12)
	httpIR.TLS.MaxVersion = ptr.To(ir.TLSv13)

	if tlsParams != nil && len(tlsParams.ALPNProtocols) > 0 {
		httpIR.TLS.ALPNProtocols = make([]string, len(tlsParams.ALPNProtocols))
		for i := range tlsParams.ALPNProtocols {
			httpIR.TLS.ALPNProtocols[i] = string(tlsParams.ALPNProtocols[i])
		}
	}

	// Return early if not set
	if tlsParams == nil {
		return nil
	}

	if tlsParams.MinVersion != nil {
		httpIR.TLS.MinVersion = ptr.To(ir.TLSVersion(*tlsParams.MinVersion))
	}
	if tlsParams.MaxVersion != nil {
		httpIR.TLS.MaxVersion = ptr.To(ir.TLSVersion(*tlsParams.MaxVersion))
	}
	if len(tlsParams.Ciphers) > 0 {
		httpIR.TLS.Ciphers = tlsParams.Ciphers
	}
	if len(tlsParams.ECDHCurves) > 0 {
		httpIR.TLS.ECDHCurves = tlsParams.ECDHCurves
	}
	if len(tlsParams.SignatureAlgorithms) > 0 {
		httpIR.TLS.SignatureAlgorithms = tlsParams.SignatureAlgorithms
	}

	if tlsParams.ClientValidation != nil {
		from := crossNamespaceFrom{
			group:     egv1a1.GroupName,
			kind:      KindClientTrafficPolicy,
			namespace: policy.Namespace,
		}

		irCACert := &ir.TLSCACertificate{
			Name: irTLSCACertName(policy.Namespace, policy.Name),
		}

		for _, caCertRef := range tlsParams.ClientValidation.CACertificateRefs {
			if caCertRef.Kind == nil || string(*caCertRef.Kind) == KindSecret { // nolint
				secret, err := t.validateSecretRef(false, from, caCertRef, resources)
				if err != nil {
					return err
				}

				secretBytes, ok := secret.Data[caCertKey]
				if !ok || len(secretBytes) == 0 {
					return fmt.Errorf(
						"caCertificateRef not found in secret %s", caCertRef.Name)
				}

				if err := validateCertificate(secretBytes); err != nil {
					return fmt.Errorf("invalid certificate in secret %s: %w", caCertRef.Name, err)
				}

				irCACert.Certificate = append(irCACert.Certificate, secretBytes...)

			} else if string(*caCertRef.Kind) == KindConfigMap {
				configMap, err := t.validateConfigMapRef(false, from, caCertRef, resources)
				if err != nil {
					return err
				}

				configMapBytes, ok := configMap.Data[caCertKey]
				if !ok || len(configMapBytes) == 0 {
					return fmt.Errorf(
						"caCertificateRef not found in configMap %s", caCertRef.Name)
				}

				if err := validateCertificate([]byte(configMapBytes)); err != nil {
					return fmt.Errorf("invalid certificate in configmap %s: %w", caCertRef.Name, err)
				}

				irCACert.Certificate = append(irCACert.Certificate, configMapBytes...)
			} else {
				return fmt.Errorf("unsupported caCertificateRef kind:%s", string(*caCertRef.Kind))
			}
		}

		if len(irCACert.Certificate) > 0 {
			httpIR.TLS.CACertificate = irCACert
		}
	}

	return nil
}

func translateListenerConnection(connection *egv1a1.Connection, httpIR *ir.HTTPListener) error {
	// Return early if not set
	if connection == nil {
		return nil
	}

	irConnection := &ir.Connection{}

	if connection.ConnectionLimit != nil {
		irConnectionLimit := &ir.ConnectionLimit{}

		irConnectionLimit.Value = ptr.To(uint64(connection.ConnectionLimit.Value))

		if connection.ConnectionLimit.CloseDelay != nil {
			d, err := time.ParseDuration(string(*connection.ConnectionLimit.CloseDelay))
			if err != nil {
				return fmt.Errorf("invalid CloseDelay value %s", *connection.ConnectionLimit.CloseDelay)
			}
			irConnectionLimit.CloseDelay = ptr.To(metav1.Duration{Duration: d})
		}

		irConnection.Limit = irConnectionLimit
	}

	httpIR.Connection = irConnection

	return nil
}
