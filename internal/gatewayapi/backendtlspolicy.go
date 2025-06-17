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

func (t *Translator) applyBackendTLSSetting(
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	parent gwapiv1a2.ParentReference,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
	isDynamicResolver bool,
) (*ir.TLSUpstreamConfig, error) {
	var (
		err       error
		tlsBundle *ir.TLSUpstreamConfig
	)

	// If the destination is a dynamic resolver, we need to use the CACertificateRefs from the backend object
	// and not from the BackendTLSPolicy. This is because the BackendTLSPolicy requires a valid hostname, and
	// dynamic resolvers's hostname is not fixed.
	if isDynamicResolver {
		upstreamConfig, err := t.processDynamicResolverBackendTLSConfig(backendRef, backendNamespace, resources)
		if err != nil {
			return nil, err
		}

		if tlsBundle, err = t.applyCommonBackendTLSConfig(upstreamConfig, backendRef, backendNamespace, resources); err != nil {
			return nil, err
		}

		return t.applyEnvoyProxyBackendTLSSetting(tlsBundle, resources, envoyProxy)
	}

	upstreamConfig, err := t.processBackendTLSPolicy(backendRef, backendNamespace, parent, resources)
	if err != nil {
		return nil, err
	}

	if tlsBundle, err = t.applyCommonBackendTLSConfig(upstreamConfig, backendRef, backendNamespace, resources); err != nil {
		return nil, err
	}

	if tlsBundle, err = t.applyEnvoyProxyBackendTLSSetting(tlsBundle, resources, envoyProxy); err != nil {
		return nil, err
	}
	return tlsBundle, nil
}

func (t *Translator) applyCommonBackendTLSConfig(
	tlsConfig *ir.TLSUpstreamConfig,
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	resources *resource.Resources,
) (*ir.TLSUpstreamConfig, error) {
	if tlsConfig == nil || KindDerefOr(backendRef.Kind, resource.KindService) != resource.KindBackend {
		return tlsConfig, nil
	}

	backend := resources.GetBackend(backendNamespace, string(backendRef.Name))
	if backend == nil {
		return nil, fmt.Errorf("backend %s not found", backendRef.Name)
	}
	if backend.Spec.TLS == nil {
		return tlsConfig, nil
	}

	tlsConfig.InsecureSkipVerify = ptr.Deref(backend.Spec.TLS.InsecureSkipVerify, false)
	return tlsConfig, nil
}

func (t *Translator) processDynamicResolverBackendTLSConfig(
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	resources *resource.Resources,
) (*ir.TLSUpstreamConfig, error) {
	backend := resources.GetBackend(backendNamespace, string(backendRef.Name))
	if backend == nil {
		return nil, fmt.Errorf("backend %s not found", backendRef.Name)
	}
	if backend.Spec.TLS == nil || (len(backend.Spec.TLS.CACertificateRefs) == 0 && backend.Spec.TLS.WellKnownCACertificates == nil) {
		return nil, nil
	}

	tlsBundle := &ir.TLSUpstreamConfig{
		InsecureSkipVerify: ptr.Deref(backend.Spec.TLS.InsecureSkipVerify, false),
	}

	if !tlsBundle.InsecureSkipVerify {
		tlsBundle.UseSystemTrustStore = ptr.Deref(backend.Spec.TLS.WellKnownCACertificates, "") == gwapiv1a3.WellKnownCACertificatesSystem

		if tlsBundle.UseSystemTrustStore {
			tlsBundle.CACertificate = &ir.TLSCACertificate{
				Name: fmt.Sprintf("%s/%s-ca", backend.Name, backend.Namespace),
			}
		} else {
			caCert, err := getCaCertsFromCARefs(backend.Namespace, backend.Spec.TLS.CACertificateRefs, resources)
			if err != nil {
				return nil, err
			}
			tlsBundle.CACertificate = &ir.TLSCACertificate{
				Certificate: caCert,
				Name:        fmt.Sprintf("%s/%s-ca", backend.Name, backend.Namespace),
			}
		}
	}

	return tlsBundle, nil
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
	ancestorRefs = append(ancestorRefs, parent)

	if err != nil {
		status.SetTranslationErrorForPolicyAncestors(&policy.Status,
			ancestorRefs,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(err),
		)
		return nil, err
	}

	status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName)
	return tlsBundle, nil
}

func (t *Translator) applyEnvoyProxyBackendTLSSetting(tlsConfig *ir.TLSUpstreamConfig, resources *resource.Resources, ep *egv1a1.EnvoyProxy) (*ir.TLSUpstreamConfig, error) {
	if ep == nil || ep.Spec.BackendTLS == nil || tlsConfig == nil {
		return tlsConfig, nil
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

		var err error
		if ns != ep.Namespace {
			err = fmt.Errorf("ClientCertificateRef Secret is not located in the same namespace as Envoyproxy. Secret namespace: %s does not match Envoyproxy namespace: %s", ns, ep.Namespace)
			return tlsConfig, err
		}
		secret := resources.GetSecret(ns, string(ep.Spec.BackendTLS.ClientCertificateRef.Name))
		if secret == nil {
			err = fmt.Errorf(
				"failed to locate TLS secret for client auth: %s specified in EnvoyProxy %s",
				types.NamespacedName{
					Namespace: ep.Namespace,
					Name:      string(ep.Spec.BackendTLS.ClientCertificateRef.Name),
				}.String(),
				types.NamespacedName{
					Namespace: ep.Namespace,
					Name:      ep.Name,
				}.String(),
			)
			return tlsConfig, err
		}
		tlsConf := irTLSConfigs(secret)
		tlsConfig.ClientCertificates = tlsConf.Certificates
	}
	return tlsConfig, nil
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

func getBackendTLSPolicy(
	policies []*gwapiv1a3.BackendTLSPolicy,
	backendRef gwapiv1a2.BackendObjectReference,
	backendNamespace string,
	resources *resource.Resources,
) *gwapiv1a3.BackendTLSPolicy {
	// SectionName is port number for EG Backend object
	target := getTargetBackendReference(backendRef, backendNamespace, resources)
	for _, policy := range policies {
		if backendTLSTargetMatched(*policy, target, backendNamespace) {
			return policy
		}
	}
	return nil
}

func getBackendTLSBundle(backendTLSPolicy *gwapiv1a3.BackendTLSPolicy, resources *resource.Resources) (*ir.TLSUpstreamConfig, error) {
	// Translate SubjectAltNames from gwapiv1a3 to ir
	var subjectAltNames []ir.SubjectAltName
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
				if crt, dataOk := cm.Data[caCertKey]; dataOk {
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
				if crt, dataOk := secret.Data[caCertKey]; dataOk {
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
		}
	}

	if ca == "" {
		return nil, fmt.Errorf("no ca found in referred ConfigMap or Secret")
	}
	return []byte(ca), nil
}

func getAncestorRefs(policy *gwapiv1a3.BackendTLSPolicy) []gwapiv1a2.ParentReference {
	ret := make([]gwapiv1a2.ParentReference, len(policy.Status.Ancestors))
	for i, ancestor := range policy.Status.Ancestors {
		ret[i] = ancestor.AncestorRef
	}
	return ret
}
