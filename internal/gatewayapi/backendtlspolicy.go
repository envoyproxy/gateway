// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"reflect"

	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
)

func (t *Translator) applyBackendTLSSetting(backendRef gwapiv1.BackendObjectReference, backendNamespace string, parent gwapiv1a2.ParentReference, resources *resource.Resources, envoyProxy *egv1a1.EnvoyProxy) *ir.TLSUpstreamConfig {
	upstreamConfig, policy := t.processBackendTLSPolicy(backendRef, backendNamespace, parent, resources, envoyProxy)
	return t.applyEnvoyProxyBackendTLSSetting(policy, upstreamConfig, resources, parent, envoyProxy)
}

func (t *Translator) processBackendTLSPolicy(
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	parent gwapiv1a2.ParentReference,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.TLSUpstreamConfig, *gwapiv1a3.BackendTLSPolicy) {
	policy := getBackendTLSPolicy(resources.BackendTLSPolicies, backendRef, backendNamespace, resources)
	if policy == nil {
		return nil, nil
	}

	tlsBundle, err := getBackendTLSBundle(policy, resources)
	if err == nil && tlsBundle == nil {
		return nil, nil
	}

	ancestorRefs := getAncestorRefs(policy)
	ancestorRefs = append(ancestorRefs, parent)

	if err != nil {
		status.SetTranslationErrorForPolicyAncestors(&policy.Status,
			ancestorRefs,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(err),
		)
		return nil, nil
	}

	status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName)
	// apply defaults as per envoyproxy
	if envoyProxy != nil {
		if envoyProxy.Spec.BackendTLS != nil {
			if len(envoyProxy.Spec.BackendTLS.Ciphers) > 0 {
				tlsBundle.Ciphers = envoyProxy.Spec.BackendTLS.Ciphers
			}
			if len(envoyProxy.Spec.BackendTLS.ECDHCurves) > 0 {
				tlsBundle.ECDHCurves = envoyProxy.Spec.BackendTLS.ECDHCurves
			}
			if len(envoyProxy.Spec.BackendTLS.SignatureAlgorithms) > 0 {
				tlsBundle.SignatureAlgorithms = envoyProxy.Spec.BackendTLS.SignatureAlgorithms
			}
			if envoyProxy.Spec.BackendTLS.MinVersion != nil {
				tlsBundle.MinVersion = ptr.To(ir.TLSVersion(*envoyProxy.Spec.BackendTLS.MinVersion))
			}
			if envoyProxy.Spec.BackendTLS.MaxVersion != nil {
				tlsBundle.MaxVersion = ptr.To(ir.TLSVersion(*envoyProxy.Spec.BackendTLS.MaxVersion))
			}
			if len(envoyProxy.Spec.BackendTLS.ALPNProtocols) > 0 {
				tlsBundle.ALPNProtocols = make([]string, len(envoyProxy.Spec.BackendTLS.ALPNProtocols))
				for i := range envoyProxy.Spec.BackendTLS.ALPNProtocols {
					tlsBundle.ALPNProtocols[i] = string(envoyProxy.Spec.BackendTLS.ALPNProtocols[i])
				}
			}
		}
	}
	return tlsBundle, policy
}

func (t *Translator) applyEnvoyProxyBackendTLSSetting(policy *gwapiv1a3.BackendTLSPolicy, tlsConfig *ir.TLSUpstreamConfig, resources *resource.Resources, parent gwapiv1a2.ParentReference, ep *egv1a1.EnvoyProxy) *ir.TLSUpstreamConfig {
	if ep == nil || ep.Spec.BackendTLS == nil || tlsConfig == nil {
		return tlsConfig
	}

	if len(ep.Spec.BackendTLS.Ciphers) > 0 {
		tlsConfig.Ciphers = ep.Spec.BackendTLS.Ciphers
	}
	if len(ep.Spec.BackendTLS.ECDHCurves) > 0 {
		tlsConfig.ECDHCurves = ep.Spec.BackendTLS.ECDHCurves
	}
	if len(ep.Spec.BackendTLS.SignatureAlgorithms) > 0 {
		tlsConfig.SignatureAlgorithms = ep.Spec.BackendTLS.SignatureAlgorithms
	}
	if ep.Spec.BackendTLS.MinVersion != nil {
		tlsConfig.MinVersion = ptr.To(ir.TLSVersion(*ep.Spec.BackendTLS.MinVersion))
	}
	if ep.Spec.BackendTLS.MaxVersion != nil {
		tlsConfig.MaxVersion = ptr.To(ir.TLSVersion(*ep.Spec.BackendTLS.MaxVersion))
	}
	if len(ep.Spec.BackendTLS.ALPNProtocols) > 0 {
		tlsConfig.ALPNProtocols = make([]string, len(ep.Spec.BackendTLS.ALPNProtocols))
		for i := range ep.Spec.BackendTLS.ALPNProtocols {
			tlsConfig.ALPNProtocols[i] = string(ep.Spec.BackendTLS.ALPNProtocols[i])
		}
	}
	if ep.Spec.BackendTLS != nil && ep.Spec.BackendTLS.ClientCertificateRef != nil {
		ns := string(ptr.Deref(ep.Spec.BackendTLS.ClientCertificateRef.Namespace, ""))
		ancestorRefs := []gwapiv1a2.ParentReference{
			parent,
		}
		if ns != ep.Namespace {
			status.SetTranslationErrorForPolicyAncestors(&policy.Status,
				ancestorRefs,
				t.GatewayControllerName,
				policy.Generation,
				status.Error2ConditionMsg(fmt.Errorf("client authentication TLS secret is not located in the same namespace as Envoyproxy. Secret namespace: %s does not match Envoyproxy namespace: %s", ns, ep.Namespace)))
			return tlsConfig
		}
		secret := resources.GetSecret(ns, string(ep.Spec.BackendTLS.ClientCertificateRef.Name))
		if secret == nil {
			status.SetTranslationErrorForPolicyAncestors(&policy.Status,
				ancestorRefs,
				t.GatewayControllerName,
				policy.Generation,
				status.Error2ConditionMsg(fmt.Errorf("failed to locate TLS secret for client auth: %s in namespace: %s", ep.Spec.BackendTLS.ClientCertificateRef.Name, ns)),
			)
			return tlsConfig
		}
		tlsConf := irTLSConfigs(secret)
		tlsConfig.ClientCertificates = tlsConf.Certificates
	}
	return tlsConfig
}

func backendTLSTargetMatched(policy gwapiv1a3.BackendTLSPolicy, target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName, backendNamespace string) bool {
	for _, currTarget := range policy.Spec.TargetRefs {
		if target.Group == currTarget.Group &&
			target.Kind == currTarget.Kind &&
			backendNamespace == policy.Namespace &&
			target.Name == currTarget.Name {
			// if section name is not set, then it targets the entire backend
			if currTarget.SectionName == nil {
				return true
			} else if reflect.DeepEqual(currTarget.SectionName, target.SectionName) {
				return true
			}
		}
	}
	return false
}

// getTargetBackendReferenceWithPortName returns the LocalPolicyTargetReference for the given BackendObjectReference,
// and sets the sectionName to the port name if the BackendObjectReference is a Kubernetes Service.
func getTargetBackendReferenceWithPortName(
	backendRef gwapiv1a2.BackendObjectReference,
	backendNamespace string,
	resources *resource.Resources,
) gwapiv1a2.LocalPolicyTargetReferenceWithSectionName {
	ref := getTargetBackendReference(backendRef)
	if backendRef.Port == nil {
		return ref
	}

	if backendRef.Kind != nil && *backendRef.Kind != resource.KindService {
		return ref
	}

	if service := resources.GetService(backendNamespace, string(backendRef.Name)); service != nil {
		for _, port := range service.Spec.Ports {
			if port.Port == int32(*backendRef.Port) {
				if port.Name != "" {
					ref.SectionName = SectionNamePtr(port.Name)
				}
			}
		}
	}

	return ref
}

func getBackendTLSPolicy(
	policies []*gwapiv1a3.BackendTLSPolicy,
	backendRef gwapiv1a2.BackendObjectReference,
	backendNamespace string,
	resources *resource.Resources,
) *gwapiv1a3.BackendTLSPolicy {
	target := getTargetBackendReference(backendRef)
	for _, policy := range policies {
		if backendTLSTargetMatched(*policy, target, backendNamespace) {
			return policy
		}
	}

	// SectionName can be port name for Kubernetes Service
	if backendRef.Port != nil &&
		(backendRef.Kind == nil || *backendRef.Kind == resource.KindService) {
		target = getTargetBackendReferenceWithPortName(backendRef, backendNamespace, resources)
		for _, policy := range policies {
			if backendTLSTargetMatched(*policy, target, backendNamespace) {
				return policy
			}
		}
	}
	return nil
}

func getBackendTLSBundle(backendTLSPolicy *gwapiv1a3.BackendTLSPolicy, resources *resource.Resources) (*ir.TLSUpstreamConfig, error) {
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
		case resource.KindConfigMap:
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
		case resource.KindSecret:
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

func getAncestorRefs(policy *gwapiv1a3.BackendTLSPolicy) []gwapiv1a2.ParentReference {
	ret := make([]gwapiv1a2.ParentReference, len(policy.Status.Ancestors))
	for i, ancestor := range policy.Status.Ancestors {
		ret[i] = ancestor.AncestorRef
	}
	return ret
}
