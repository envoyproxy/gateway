// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
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

	// Translate
	// 1. First translate Policies with a sectionName set
	// 2. Then loop again and translate the policies without a sectionName
	// TODO: Import sort order to ensure policy with same section always appear
	// before policy with no section so below loops can be flattened into 1.

	for _, policy := range clientTrafficPolicies {
		if hasSectionName(policy) {
			policy := policy.DeepCopy()
			res = append(res, policy)

			gateway := resolveCTPolicyTargetRef(policy, gateways)

			// Negative statuses have already been assigned so its safe to skip
			if gateway == nil {
				continue
			}

			// Check for conflicts
			key := types.NamespacedName{
				Name:      gateway.Name,
				Namespace: gateway.Namespace,
			}

			// Check if another policy targeting the same section exists
			section := string(*(policy.Spec.TargetRef.SectionName))
			s, ok := policyMap[key]
			if ok && s.Has(section) {
				message := "Unable to target section, another ClientTrafficPolicy has already attached to it"

				status.SetClientTrafficPolicyCondition(policy,
					gwv1a2.PolicyConditionAccepted,
					metav1.ConditionFalse,
					gwv1a2.PolicyReasonConflicted,
					message,
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
				if string(l.Name) == section {
					err = t.translateClientTrafficPolicyForListener(policy, l, xdsIR, infraIR, resources)
					break
				}
			}
			if err != nil {
				status.SetClientTrafficPolicyCondition(policy,
					gwv1a2.PolicyConditionAccepted,
					metav1.ConditionFalse,
					gwv1a2.PolicyReasonInvalid,
					status.Error2ConditionMsg(err),
				)
			} else {
				// Set Accepted=True
				status.SetClientTrafficPolicyCondition(policy,
					gwv1a2.PolicyConditionAccepted,
					metav1.ConditionTrue,
					gwv1a2.PolicyReasonAccepted,
					"ClientTrafficPolicy has been accepted.",
				)
			}
		}
	}

	// Policy with no section set (targeting all sections)
	for _, policy := range clientTrafficPolicies {
		if !hasSectionName(policy) {

			policy := policy.DeepCopy()
			res = append(res, policy)

			gateway := resolveCTPolicyTargetRef(policy, gateways)

			// Negative statuses have already been assigned so its safe to skip
			if gateway == nil {
				continue
			}

			// Check for conflicts
			key := types.NamespacedName{
				Name:      gateway.Name,
				Namespace: gateway.Namespace,
			}
			s, ok := policyMap[key]
			// Check if another policy targeting the same Gateway exists
			if ok && s.Has(AllSections) {
				message := "Unable to target Gateway, another ClientTrafficPolicy has already attached to it"

				status.SetClientTrafficPolicyCondition(policy,
					gwv1a2.PolicyConditionAccepted,
					metav1.ConditionFalse,
					gwv1a2.PolicyReasonConflicted,
					message,
				)

				continue

			}

			// Check if another policy targeting the same Gateway exists
			if ok && (s.Len() > 0) {
				// Maintain order here to ensure status/string does not change with same data
				sections := s.UnsortedList()
				sort.Strings(sections)
				message := fmt.Sprintf("There are existing ClientTrafficPolicies that are overriding these sections %v", sections)

				status.SetClientTrafficPolicyCondition(policy,
					egv1a1.PolicyConditionOverridden,
					metav1.ConditionTrue,
					egv1a1.PolicyReasonOverridden,
					message,
				)
			}

			// Add section to policy map
			if s == nil {
				policyMap[key] = sets.New[string]()
			}
			policyMap[key].Insert(AllSections)

			// Translate sections that have not yet been targeted
			var err error
			for _, l := range gateway.listeners {
				// Skip if section has already been targeted
				if s != nil && s.Has(string(l.Name)) {
					continue
				}

				err = t.translateClientTrafficPolicyForListener(policy, l, xdsIR, infraIR, resources)
			}

			if err != nil {
				status.SetClientTrafficPolicyCondition(policy,
					gwv1a2.PolicyConditionAccepted,
					metav1.ConditionFalse,
					gwv1a2.PolicyReasonInvalid,
					status.Error2ConditionMsg(err),
				)
			} else {
				// Set Accepted=True
				status.SetClientTrafficPolicyCondition(policy,
					gwv1a2.PolicyConditionAccepted,
					metav1.ConditionTrue,
					gwv1a2.PolicyReasonAccepted,
					"ClientTrafficPolicy has been accepted.",
				)
			}
		}
	}

	return res
}

func resolveCTPolicyTargetRef(policy *egv1a1.ClientTrafficPolicy, gateways []*GatewayContext) *GatewayContext {
	targetNs := policy.Spec.TargetRef.Namespace
	// If empty, default to namespace of policy
	if targetNs == nil {
		targetNs = ptr.To(gwv1b1.Namespace(policy.Namespace))
	}

	// Ensure policy can only target a Gateway
	if policy.Spec.TargetRef.Group != gwv1b1.GroupName || policy.Spec.TargetRef.Kind != KindGateway {
		message := fmt.Sprintf("TargetRef.Group:%s TargetRef.Kind:%s, only TargetRef.Group:%s and TargetRef.Kind:%s is supported.",
			policy.Spec.TargetRef.Group, policy.Spec.TargetRef.Kind, gwv1b1.GroupName, KindGateway)

		status.SetClientTrafficPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonInvalid,
			message,
		)
		return nil
	}

	// Ensure Policy and target Gateway are in the same namespace
	if policy.Namespace != string(*targetNs) {

		message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, ClientTrafficPolicy can only target a Gateway in the same namespace.",
			policy.Namespace, *targetNs)
		status.SetClientTrafficPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonInvalid,
			message,
		)
		return nil
	}

	// Find the Gateway
	var gateway *GatewayContext
	for _, g := range gateways {
		if g.Name == string(policy.Spec.TargetRef.Name) && g.Namespace == string(*targetNs) {
			gateway = g
			break
		}
	}

	// Gateway not found
	if gateway == nil {
		message := fmt.Sprintf("Gateway:%s not found.", policy.Spec.TargetRef.Name)

		status.SetClientTrafficPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonTargetNotFound,
			message,
		)
		return nil
	}

	// If sectionName is set, make sure its valid
	if policy.Spec.TargetRef.SectionName != nil {
		found := false
		for _, l := range gateway.listeners {
			if l.Name == *(policy.Spec.TargetRef.SectionName) {
				found = true
				break
			}
		}
		if !found {
			message := fmt.Sprintf("SectionName(Listener):%s not found.", *(policy.Spec.TargetRef.SectionName))
			status.SetClientTrafficPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionFalse,
				gwv1a2.PolicyReasonTargetNotFound,
				message,
			)
			return nil
		}
	}

	return gateway
}

func (t *Translator) translateClientTrafficPolicyForListener(policy *egv1a1.ClientTrafficPolicy, l *ListenerContext,
	xdsIR XdsIRMap, infraIR InfraIRMap, resources *Resources) error {
	// Find IR
	irKey := irStringKey(l.gateway.Namespace, l.gateway.Name)
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

		// Translate Proxy Protocol
		translateListenerProxyProtocol(policy.Spec.EnableProxyProtocol, httpIR)

		// Translate Client IP Detection
		translateClientIPDetection(policy.Spec.ClientIPDetection, httpIR)

		// Translate Header Settings
		translateListenerHeaderSettings(policy.Spec.Headers, httpIR)

		// Translate Path Settings
		translatePathSettings(policy.Spec.Path, httpIR)

		// Translate HTTP1 Settings
		if err := translateHTTP1Settings(policy.Spec.HTTP1, httpIR); err != nil {
			return err
		}

		// enable http3 if set and TLS is enabled
		if httpIR.TLS != nil && policy.Spec.HTTP3 != nil {
			httpIR.HTTP3 = &ir.HTTP3Settings{}
			var proxyListenerIR *ir.ProxyListener
			for _, proxyListener := range infraIR[irKey].Proxy.Listeners {
				if proxyListener.Name == irListenerName {
					proxyListenerIR = proxyListener
					break
				}
			}
			if proxyListenerIR != nil {
				proxyListenerIR.HTTP3 = &ir.HTTP3Settings{}
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
		EnableEnvoyHeaders: ptr.Deref(headerSettings.EnableEnvoyHeaders, false),
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
	// If HTTP3 is enabled, the ALPN protocols array should be hardcoded
	// for HTTP3
	if httpIR.HTTP3 != nil {
		httpIR.TLS.ALPNProtocols = []string{"h3"}
	} else if tlsParams != nil && len(tlsParams.ALPNProtocols) > 0 {
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
