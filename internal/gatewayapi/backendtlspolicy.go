// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/status"
)

func (t *Translator) processBackendTLSPolicy(
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	parent gwapiv1a2.ParentReference,
	resources *Resources) *ir.TLSUpstreamConfig {
	tlsBundle, err := getBackendTLSBundle(resources.BackendTLSPolicies, resources.ConfigMaps, backendRef, backendNamespace)
	if err == nil && tlsBundle == nil {
		return nil
	}

	policy := getBackendTLSPolicy(resources.BackendTLSPolicies, backendRef, backendNamespace)

	ancestor := gwapiv1a2.PolicyAncestorStatus{
		AncestorRef:    parent,
		ControllerName: gwapiv1.GatewayController(t.GatewayControllerName),
	}

	if err != nil {
		status.SetBackendTLSPolicyCondition(
			policy,
			ancestor,
			gwapiv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwapiv1a2.PolicyReasonInvalid,
			status.Error2ConditionMsg(err))
		return nil
	}

	// Check if the reference from BackendTLSPolicy to BackendRef is permitted by
	// any ReferenceGrant
	backendRefKind := KindService
	if backendRef.Kind != nil {
		backendRefKind = string(*backendRef.Kind)
	}
	if policy.Namespace != backendNamespace {
		if !t.validateCrossNamespaceRef(
			crossNamespaceFrom{
				group:     gwapiv1.GroupName,
				kind:      KindBackendTLSPolicy,
				namespace: policy.Namespace,
			},
			crossNamespaceTo{
				group:     "",
				kind:      backendRefKind,
				namespace: backendNamespace,
				name:      string(backendRef.Name),
			},
			resources.ReferenceGrants,
		) {
			status.SetBackendTLSPolicyCondition(
				policy,
				ancestor,
				gwapiv1a2.PolicyConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1a2.PolicyReasonInvalid,
				fmt.Sprintf("target ref to %s %s/%s not permitted by any ReferenceGrant",
					backendRefKind, backendNamespace, backendRef.Name))
			return nil
		}
	}

	status.SetBackendTLSPolicyCondition(
		policy,
		ancestor,
		gwapiv1a2.PolicyConditionAccepted,
		metav1.ConditionTrue,
		gwapiv1a2.PolicyReasonAccepted,
		"BackendTLSPolicy is Accepted")
	return tlsBundle
}

func backendTLSTargetMatched(policy gwapiv1a2.BackendTLSPolicy, target gwapiv1a2.PolicyTargetReferenceWithSectionName) bool {

	policyTarget := policy.Spec.TargetRef

	if target.Group == policyTarget.Group &&
		target.Kind == policyTarget.Kind &&
		target.Name == policyTarget.Name &&
		NamespaceDerefOr(policyTarget.Namespace, policy.Namespace) == string(*target.Namespace) {
		if policyTarget.SectionName != nil && *policyTarget.SectionName != *target.SectionName {
			return false
		}
		return true
	}
	return false
}

func getBackendTLSPolicy(policies []*gwapiv1a2.BackendTLSPolicy, backendRef gwapiv1a2.BackendObjectReference, backendNamespace string) *gwapiv1a2.BackendTLSPolicy {
	target := GetTargetBackendReference(backendRef, backendNamespace)
	for _, policy := range policies {
		if backendTLSTargetMatched(*policy, target) {
			return policy
		}
	}
	return nil
}

func getBackendTLSBundle(policies []*gwapiv1a2.BackendTLSPolicy, configmaps []*corev1.ConfigMap, backendRef gwapiv1a2.BackendObjectReference, backendNamespace string) (*ir.TLSUpstreamConfig, error) {

	backendTLSPolicy := getBackendTLSPolicy(policies, backendRef, backendNamespace)

	if backendTLSPolicy == nil {
		return nil, nil
	}

	tlsBundle := &ir.TLSUpstreamConfig{
		SNI:                 string(backendTLSPolicy.Spec.TLS.Hostname),
		UseSystemTrustStore: ptr.Deref(backendTLSPolicy.Spec.TLS.WellKnownCACerts, "") == gwapiv1a2.WellKnownCACertSystem,
	}
	if tlsBundle.UseSystemTrustStore {
		return tlsBundle, nil
	}

	caRefMap := make(map[string]string)

	for _, caRef := range backendTLSPolicy.Spec.TLS.CACertRefs {
		caRefMap[string(caRef.Name)] = string(caRef.Kind)
	}

	ca := ""

	for _, cmap := range configmaps {
		if kind, ok := caRefMap[cmap.Name]; ok && kind == cmap.Kind {
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

	if ca == "" {
		return nil, fmt.Errorf("no ca found in referred configmaps")
	}
	tlsBundle.CACertificate = &ir.TLSCACertificate{
		Certificate: []byte(ca),
		Name:        fmt.Sprintf("%s/%s-ca", backendTLSPolicy.Name, backendTLSPolicy.Namespace),
	}

	return tlsBundle, nil
}

func (t *Translator) ProcessBackendTLSPoliciesAncestorRef(backendTLSPolicies []*gwapiv1a2.BackendTLSPolicy, gateways []*GatewayContext) []*gwapiv1a2.BackendTLSPolicy {

	var res []*gwapiv1a2.BackendTLSPolicy

	for _, btlsPolicy := range backendTLSPolicies {

		policy := btlsPolicy.DeepCopy()
		res = append(res, policy)

		if policy.Status.Ancestors != nil {
			for k, status := range policy.Status.Ancestors {
				if status.AncestorRef.Kind != nil && *status.AncestorRef.Kind != KindGateway {
					continue
				}
				exist := false
				for _, gwContext := range gateways {
					gw := gwContext.Gateway
					if gw.Name == string(status.AncestorRef.Name) && gw.Namespace == NamespaceDerefOrAlpha(status.AncestorRef.Namespace, "default") {
						for _, lis := range gw.Spec.Listeners {
							if lis.Name == ptr.Deref(status.AncestorRef.SectionName, "") {
								exist = true
							}
						}
					}
				}

				if !exist {
					policy.Status.Ancestors = append(policy.Status.Ancestors[:k], policy.Status.Ancestors[k+1:]...)
				}
			}
		} else {
			policy.Status.Ancestors = []gwapiv1a2.PolicyAncestorStatus{}
		}
	}

	return res
}
