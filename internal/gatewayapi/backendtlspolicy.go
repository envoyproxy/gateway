// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"

	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
)

func (t *Translator) processBackendTLSPolicy(
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	parent gwapiv1a2.ParentReference,
	resources *Resources,
) *ir.TLSUpstreamConfig {
	policy := getBackendTLSPolicy(resources.BackendTLSPolicies, backendRef, backendNamespace)
	if policy == nil {
		return nil
	}

	tlsBundle, err := getBackendTLSBundle(policy, resources)
	if err == nil && tlsBundle == nil {
		return nil
	}

	ancestorRefs := []gwapiv1a2.ParentReference{
		parent,
	}

	if err != nil {
		status.SetTranslationErrorForPolicyAncestors(&policy.Status,
			ancestorRefs,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(err),
		)
		return nil
	}

	status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName)
	// apply defaults as per envoyproxy
	if resources.EnvoyProxy != nil {
		if resources.EnvoyProxy.Spec.BackendTLS != nil {
			if len(resources.EnvoyProxy.Spec.BackendTLS.Ciphers) > 0 {
				tlsBundle.Ciphers = resources.EnvoyProxy.Spec.BackendTLS.Ciphers
			}
			if len(resources.EnvoyProxy.Spec.BackendTLS.ECDHCurves) > 0 {
				tlsBundle.ECDHCurves = resources.EnvoyProxy.Spec.BackendTLS.ECDHCurves
			}
			if len(resources.EnvoyProxy.Spec.BackendTLS.SignatureAlgorithms) > 0 {
				tlsBundle.SignatureAlgorithms = resources.EnvoyProxy.Spec.BackendTLS.SignatureAlgorithms
			}
			if resources.EnvoyProxy.Spec.BackendTLS.MinVersion != nil {
				tlsBundle.MinVersion = ptr.To(ir.TLSVersion(*resources.EnvoyProxy.Spec.BackendTLS.MinVersion))
			}
			if resources.EnvoyProxy.Spec.BackendTLS.MinVersion != nil {
				tlsBundle.MaxVersion = ptr.To(ir.TLSVersion(*resources.EnvoyProxy.Spec.BackendTLS.MaxVersion))
			}
			if len(resources.EnvoyProxy.Spec.BackendTLS.ALPNProtocols) > 0 {
				tlsBundle.ALPNProtocols = make([]string, len(resources.EnvoyProxy.Spec.BackendTLS.ALPNProtocols))
				for i := range resources.EnvoyProxy.Spec.BackendTLS.ALPNProtocols {
					tlsBundle.ALPNProtocols[i] = string(resources.EnvoyProxy.Spec.BackendTLS.ALPNProtocols[i])
				}
			}
		}
	}
	return tlsBundle
}

func backendTLSTargetMatched(policy gwapiv1a3.BackendTLSPolicy, target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName, backendNamespace string) bool {
	// TODO: support multiple targetRefs
	policyTarget := policy.Spec.TargetRefs[0]

	if target.Group == policyTarget.Group &&
		target.Kind == policyTarget.Kind &&
		backendNamespace == policy.Namespace &&
		target.Name == policyTarget.Name {
		if policyTarget.SectionName != nil && *policyTarget.SectionName != *target.SectionName {
			return false
		}
		return true
	}
	return false
}

func getBackendTLSPolicy(policies []*gwapiv1a3.BackendTLSPolicy, backendRef gwapiv1a2.BackendObjectReference, backendNamespace string) *gwapiv1a3.BackendTLSPolicy {
	target := getTargetBackendReference(backendRef)
	for _, policy := range policies {
		if backendTLSTargetMatched(*policy, target, backendNamespace) {
			return policy
		}
	}
	return nil
}

func getBackendTLSBundle(backendTLSPolicy *gwapiv1a3.BackendTLSPolicy, resources *Resources) (*ir.TLSUpstreamConfig, error) {
	tlsBundle := &ir.TLSUpstreamConfig{
		SNI:                 string(backendTLSPolicy.Spec.Validation.Hostname),
		UseSystemTrustStore: ptr.Deref(backendTLSPolicy.Spec.Validation.WellKnownCACertificates, "") == gwapiv1a3.WellKnownCACertificatesSystem,
	}
	if tlsBundle.UseSystemTrustStore {
		return tlsBundle, nil
	}

	ca := ""
	for _, caRef := range backendTLSPolicy.Spec.Validation.CACertificateRefs {
		kind := string(caRef.Kind)

		switch kind {
		case KindConfigMap:
			for _, cmap := range resources.ConfigMaps {
				if cmap.Name == string(caRef.Name) {
					if crt, dataOk := cmap.Data["ca.crt"]; dataOk {
						if ca != "" {
							ca += "\n"
						}
						ca += crt
					} else {
						return nil, fmt.Errorf("no ca found in configmap %s", cmap.Name)
					}
				}
			}
		case KindSecret:
			for _, secret := range resources.Secrets {
				if secret.Name == string(caRef.Name) {
					if crt, dataOk := secret.Data["ca.crt"]; dataOk {
						if ca != "" {
							ca += "\n"
						}
						ca += string(crt)
					} else {
						return nil, fmt.Errorf("no ca found in secret %s", secret.Name)
					}
				}
			}
		}
	}

	if ca == "" {
		return nil, fmt.Errorf("no ca found in referred configmaps")
	}
	tlsBundle.CACertificate = &ir.TLSCACertificate{
		Certificate: []byte(ca),
		Name:        fmt.Sprintf("%s/%s-ca", backendTLSPolicy.Name, backendTLSPolicy.Namespace),
	}

	return tlsBundle, nil
}
