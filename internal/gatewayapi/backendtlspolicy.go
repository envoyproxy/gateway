// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
)

// ProcessBackendTLSPolicyStatus is called to post-process Backend TLS Policy status
// after they were applied in all relevant translations.
func (t *Translator) ProcessBackendTLSPolicyStatus(btlsp []*gwapiv1a3.BackendTLSPolicy) {
	for _, policy := range btlsp {
		// Truncate Ancestor list of longer than 16
		if len(policy.Status.Ancestors) > 16 {
			status.TruncatePolicyAncestors(&policy.Status, t.GatewayControllerName, policy.Generation)
		}
	}
}

func (t *Translator) applyBackendTLSSetting(
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	parent gwapiv1a2.ParentReference,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.TLSUpstreamConfig, error) {
	var (
		backendServerValidationTLSConfig *ir.TLSUpstreamConfig // the TLS config to validate the server cert from Backend TLS settings
		backendTLSPolicyConfig           *ir.TLSUpstreamConfig // the TLS config to validate the server cert from BackendTLSPolicy
		backendClientTLSConfig           *ir.TLSConfig         // the TLS config for client cert and common TLS settings from Backend TLS settings
		envoyProxyClientTLSConfig        *ir.TLSConfig         // the TLS config for client cert and common TLS settings from EnvoyProxy BackendTLS
		mergedClientTLSConfig            *ir.TLSConfig         // the final merged client TLS config to return
		upstreamConfig                   *ir.TLSUpstreamConfig // the final merged TLS config to return
		err                              error
	)

	// If the backendRef is a Backend resource, we need to check if it has TLS settings.
	if KindDerefOr(backendRef.Kind, resource.KindService) == egv1a1.KindBackend {
		backend := resources.GetBackend(backendNamespace, string(backendRef.Name))
		if backend == nil {
			return nil, fmt.Errorf("backend %s not found", backendRef.Name)
		}
		if backend.Spec.TLS != nil {
			// Get the backend server certificate validation settings from Backend resource.
			if backendServerValidationTLSConfig, err = t.processServerValidationTLSSettings(backend, resources); err != nil {
				return nil, err
			}

			// Get the client certificate and common TLS settings from Backend resource.
			if backend.Spec.TLS.ClientTLS != nil {
				if backendClientTLSConfig, err = t.processClientTLSSettings(resources, backend.Spec.TLS.ClientTLS, backend.Namespace, backend.Name, false); err != nil {
					return nil, err
				}
			}
		}
	}

	// Get the backend certificate validation settings from BackendTLSPolicy.
	if backendTLSPolicyConfig, err = t.processBackendTLSPolicy(backendRef, backendNamespace, parent, resources); err != nil {
		return nil, err
	}

	// Get the client certificate and common TLS settings from EnvoyProxy resource.
	backendTLSEnabled := backendServerValidationTLSConfig != nil || backendTLSPolicyConfig != nil
	if backendTLSEnabled && envoyProxy != nil && envoyProxy.Spec.BackendTLS != nil {
		if envoyProxyClientTLSConfig, err = t.processClientTLSSettings(resources, envoyProxy.Spec.BackendTLS, envoyProxy.Namespace, envoyProxy.Name, true); err != nil {
			return nil, err
		}
	}

	// Merge server validation TLS settings from Backend resource and BackendTLSPolicy.
	// BackendTLSPolicy takes precedence over Backend resource for identical attributes that are set in both.
	upstreamConfig = mergeServerValidationTLSConfigs(backendServerValidationTLSConfig, backendTLSPolicyConfig)

	if upstreamConfig != nil && !upstreamConfig.InsecureSkipVerify && upstreamConfig.CACertificate == nil {
		return nil, fmt.Errorf("CACertificate must be specified when InsecureSkipVerify is false")
	}

	// Merge client TLS settings from Backend resource and EnvoyProxy resource.
	// Backend resource client TLS settings take precedence over EnvoyProxy client TLS settings.
	mergedClientTLSConfig = mergeClientTLSConfigs(backendClientTLSConfig, envoyProxyClientTLSConfig)
	if mergedClientTLSConfig != nil {
		upstreamConfig.TLSConfig = *mergedClientTLSConfig
	}

	return upstreamConfig, nil
}

// Merges TLS settings from Gateway API BackendTLSPolicy and Envoy Gateway Backend TL.
// BackendTLSPolicy takes precedence for identical attributes that are set in both.
func mergeServerValidationTLSConfigs(
	backendServerValidationTLSConfig *ir.TLSUpstreamConfig,
	backendTLSPolicyConfig *ir.TLSUpstreamConfig,
) *ir.TLSUpstreamConfig {
	if backendServerValidationTLSConfig == nil && backendTLSPolicyConfig == nil {
		return nil
	}

	if backendServerValidationTLSConfig == nil {
		return backendTLSPolicyConfig
	}
	if backendTLSPolicyConfig == nil {
		return backendServerValidationTLSConfig
	}

	// We don't use DeepCopy here to avoid unnecessary memory allocation.
	mergedConfig := backendServerValidationTLSConfig

	if backendTLSPolicyConfig.CACertificate != nil {
		mergedConfig.CACertificate = backendTLSPolicyConfig.CACertificate
	}
	if backendTLSPolicyConfig.SNI != nil {
		mergedConfig.SNI = backendTLSPolicyConfig.SNI
	}
	if backendTLSPolicyConfig.UseSystemTrustStore {
		mergedConfig.UseSystemTrustStore = backendTLSPolicyConfig.UseSystemTrustStore
	}
	if backendTLSPolicyConfig.SubjectAltNames != nil {
		mergedConfig.SubjectAltNames = backendTLSPolicyConfig.SubjectAltNames
	}

	return mergedConfig
}

// Merges client TLS settings from backend TLS settings and EnvoyProxy BackendTLS settings.
// Backend TLS settings take precedence for identical attributes that are set in both.
func mergeClientTLSConfigs(
	backendClientTLSConfig *ir.TLSConfig,
	envoyProxyClientTLSConfig *ir.TLSConfig,
) *ir.TLSConfig {
	if backendClientTLSConfig == nil && envoyProxyClientTLSConfig == nil {
		return nil
	}

	if backendClientTLSConfig == nil {
		return envoyProxyClientTLSConfig
	}

	if envoyProxyClientTLSConfig == nil {
		return backendClientTLSConfig
	}

	// We don't use DeepCopy here to avoid unnecessary memory allocation.
	mergedConfig := envoyProxyClientTLSConfig

	if len(backendClientTLSConfig.ClientCertificates) > 0 {
		mergedConfig.ClientCertificates = backendClientTLSConfig.ClientCertificates
	}

	if backendClientTLSConfig.MinVersion != nil {
		minVersion := *backendClientTLSConfig.MinVersion
		mergedConfig.MinVersion = &minVersion
	}

	if backendClientTLSConfig.MaxVersion != nil {
		maxVersion := *backendClientTLSConfig.MaxVersion
		mergedConfig.MaxVersion = &maxVersion
	}

	if len(backendClientTLSConfig.Ciphers) > 0 {
		mergedConfig.Ciphers = backendClientTLSConfig.Ciphers
	}

	if len(backendClientTLSConfig.ECDHCurves) > 0 {
		mergedConfig.ECDHCurves = backendClientTLSConfig.ECDHCurves
	}

	if len(backendClientTLSConfig.SignatureAlgorithms) > 0 {
		mergedConfig.SignatureAlgorithms = backendClientTLSConfig.SignatureAlgorithms
	}

	if len(backendClientTLSConfig.ALPNProtocols) > 0 {
		mergedConfig.ALPNProtocols = backendClientTLSConfig.ALPNProtocols
	}

	return mergedConfig
}

func (t *Translator) processServerValidationTLSSettings(
	backend *egv1a1.Backend,
	resources *resource.Resources,
) (*ir.TLSUpstreamConfig, error) {
	tlsConfig := &ir.TLSUpstreamConfig{
		InsecureSkipVerify: ptr.Deref(backend.Spec.TLS.InsecureSkipVerify, false),
	}

	if backend.Spec.TLS.SNI != nil {
		tlsConfig.SNI = ptr.To(string(*backend.Spec.TLS.SNI))
	}

	if !tlsConfig.InsecureSkipVerify {
		tlsConfig.UseSystemTrustStore = ptr.Deref(backend.Spec.TLS.WellKnownCACertificates, "") == gwapiv1a3.WellKnownCACertificatesSystem

		if tlsConfig.UseSystemTrustStore {
			tlsConfig.CACertificate = &ir.TLSCACertificate{
				Name: fmt.Sprintf("%s/%s-ca", backend.Name, backend.Namespace),
			}
		} else if len(backend.Spec.TLS.CACertificateRefs) > 0 {
			caCert, err := getCaCertsFromCARefs(backend.Namespace, backend.Spec.TLS.CACertificateRefs, resources)
			if err != nil {
				return nil, err
			}
			tlsConfig.CACertificate = &ir.TLSCACertificate{
				Certificate: caCert,
				Name:        fmt.Sprintf("%s/%s-ca", backend.Name, backend.Namespace),
			}
		}
	}
	return tlsConfig, nil
}

func (t *Translator) processBackendTLSPolicy(
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	parent gwapiv1a2.ParentReference,
	resources *resource.Resources,
) (*ir.TLSUpstreamConfig, error) {
	policy := getBackendTLSPolicy(resources.BackendTLSPolicies, backendRef, backendNamespace, resources)
	if policy == nil {
		return nil, nil
	}

	tlsBundle, err := getBackendTLSBundle(policy, resources)
	ancestorRefs := getAncestorRefs(policy)
	ancestorRefs = append(ancestorRefs, &parent)

	if err != nil {
		status.SetTranslationErrorForPolicyAncestors(&policy.Status,
			ancestorRefs,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(err),
		)
		return nil, err
	}

	status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName, policy.Generation)
	return tlsBundle, nil
}

func (t *Translator) processClientTLSSettings(resources *resource.Resources, clientTLS *egv1a1.BackendTLSConfig, ownerNs, ownerName string, fromEnvoyProxy bool) (*ir.TLSConfig, error) {
	tlsConfig := &ir.TLSConfig{}

	if len(clientTLS.Ciphers) > 0 {
		tlsConfig.Ciphers = clientTLS.Ciphers
	}
	if len(clientTLS.ECDHCurves) > 0 {
		tlsConfig.ECDHCurves = clientTLS.ECDHCurves
	}
	if len(clientTLS.SignatureAlgorithms) > 0 {
		tlsConfig.SignatureAlgorithms = clientTLS.SignatureAlgorithms
	}
	if clientTLS.MinVersion != nil {
		tlsConfig.MinVersion = ptr.To(ir.TLSVersion(*clientTLS.MinVersion))
	}
	if clientTLS.MaxVersion != nil {
		tlsConfig.MaxVersion = ptr.To(ir.TLSVersion(*clientTLS.MaxVersion))
	}
	if len(clientTLS.ALPNProtocols) > 0 {
		tlsConfig.ALPNProtocols = make([]string, len(clientTLS.ALPNProtocols))
		for i := range clientTLS.ALPNProtocols {
			tlsConfig.ALPNProtocols[i] = string(clientTLS.ALPNProtocols[i])
		}
	}
	if clientTLS.ClientCertificateRef != nil {
		var err error
		var ownerResource string

		if fromEnvoyProxy {
			ownerResource = "EnvoyProxy"
		} else {
			ownerResource = "Backend"
		}

		ns := string(ptr.Deref(clientTLS.ClientCertificateRef.Namespace, ""))
		if ns != ownerNs {
			err = fmt.Errorf("ClientCertificateRef Secret is not located in the same namespace as %s. Secret namespace: %s does not match %s namespace: %s", ownerResource, ns, ownerResource, ownerNs)
			return tlsConfig, err
		}
		secret := resources.GetSecret(ns, string(clientTLS.ClientCertificateRef.Name))
		if secret == nil {
			err = fmt.Errorf(
				"failed to locate TLS secret for client auth: %s specified in %s %s",
				types.NamespacedName{
					Namespace: ownerNs,
					Name:      string(clientTLS.ClientCertificateRef.Name),
				}.String(),
				ownerResource,
				types.NamespacedName{
					Namespace: ownerNs,
					Name:      ownerName,
				}.String(),
			)
			return tlsConfig, err
		}
		tlsConf := irTLSConfigs(secret)
		tlsConfig.ClientCertificates = tlsConf.Certificates
	}
	return tlsConfig, nil
}

func backendTLSTargetMatched(policy *gwapiv1a3.BackendTLSPolicy, target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName, backendNamespace string) bool {
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

func getBackendTLSPolicy(
	policies []*gwapiv1a3.BackendTLSPolicy,
	backendRef gwapiv1a2.BackendObjectReference,
	backendNamespace string,
	resources *resource.Resources,
) *gwapiv1a3.BackendTLSPolicy {
	// SectionName is port number for EG Backend object
	target := getTargetBackendReference(backendRef, backendNamespace, resources)
	for _, policy := range policies {
		if backendTLSTargetMatched(policy, target, backendNamespace) {
			return policy
		}
	}
	return nil
}

func getBackendTLSBundle(backendTLSPolicy *gwapiv1a3.BackendTLSPolicy, resources *resource.Resources) (*ir.TLSUpstreamConfig, error) {
	// Translate SubjectAltNames from gwapiv1a3 to ir
	subjectAltNames := make([]ir.SubjectAltName, 0, len(backendTLSPolicy.Spec.Validation.SubjectAltNames))
	for _, san := range backendTLSPolicy.Spec.Validation.SubjectAltNames {
		var subjectAltName ir.SubjectAltName
		switch san.Type {
		case gwapiv1a3.HostnameSubjectAltNameType:
			subjectAltName.Hostname = ptr.To(string(san.Hostname))
		case gwapiv1a3.URISubjectAltNameType:
			subjectAltName.URI = ptr.To(string(san.URI))
		default:
			continue // skip unknown types
		}
		subjectAltNames = append(subjectAltNames, subjectAltName)
	}

	tlsBundle := &ir.TLSUpstreamConfig{
		SNI:                 ptr.To(string(backendTLSPolicy.Spec.Validation.Hostname)),
		UseSystemTrustStore: ptr.Deref(backendTLSPolicy.Spec.Validation.WellKnownCACertificates, "") == gwapiv1a3.WellKnownCACertificatesSystem,
		SubjectAltNames:     subjectAltNames,
	}
	if tlsBundle.UseSystemTrustStore {
		tlsBundle.CACertificate = &ir.TLSCACertificate{
			Name: fmt.Sprintf("%s/%s-ca", backendTLSPolicy.Name, backendTLSPolicy.Namespace),
		}
		return tlsBundle, nil
	}

	caCert, err := getCaCertsFromCARefs(backendTLSPolicy.Namespace, backendTLSPolicy.Spec.Validation.CACertificateRefs, resources)
	if err != nil {
		return nil, err
	}
	tlsBundle.CACertificate = &ir.TLSCACertificate{
		Certificate: caCert,
		Name:        fmt.Sprintf("%s/%s-ca", backendTLSPolicy.Name, backendTLSPolicy.Namespace),
	}
	return tlsBundle, nil
}

func getCaCertsFromCARefs(namespace string, caCertificates []gwapiv1.LocalObjectReference, resources *resource.Resources) ([]byte, error) {
	ca := ""
	for _, caRef := range caCertificates {
		kind := string(caRef.Kind)

		switch kind {
		case resource.KindConfigMap:
			cm := resources.GetConfigMap(namespace, string(caRef.Name))
			if cm != nil {
				if crt, dataOk := getCaCertFromConfigMap(cm); dataOk {
					if ca != "" {
						ca += "\n"
					}
					ca += crt
				} else {
					return nil, fmt.Errorf("no ca found in configmap %s", cm.Name)
				}
			} else {
				return nil, fmt.Errorf("configmap %s not found in namespace %s", caRef.Name, namespace)
			}
		case resource.KindSecret:
			secret := resources.GetSecret(namespace, string(caRef.Name))
			if secret != nil {
				if crt, dataOk := getCaCertFromSecret(secret); dataOk {
					if ca != "" {
						ca += "\n"
					}
					ca += string(crt)
				} else {
					return nil, fmt.Errorf("no ca found in secret %s", secret.Name)
				}
			} else {
				return nil, fmt.Errorf("secret %s not found in namespace %s", caRef.Name, namespace)
			}
		case resource.KindClusterTrustBundle:
			ctb := resources.GetClusterTrustBundle(string(caRef.Name))
			if ctb != nil {
				if ca != "" {
					ca += "\n"
				}
				ca += ctb.Spec.TrustBundle
			} else {
				return nil, fmt.Errorf("cluster trust bundle %s not found", caRef.Name)
			}
		}
	}

	if ca == "" {
		return nil, fmt.Errorf("no ca found in referred ConfigMap or Secret")
	}
	return []byte(ca), nil
}

func getAncestorRefs(policy *gwapiv1a3.BackendTLSPolicy) []*gwapiv1a2.ParentReference {
	ret := make([]*gwapiv1a2.ParentReference, len(policy.Status.Ancestors))
	for i, ancestor := range policy.Status.Ancestors {
		ret[i] = &ancestor.AncestorRef
	}
	return ret
}
