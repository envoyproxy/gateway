// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapixv1a1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
)

func (t *Translator) ProcessXBackendTrafficPolicy(resources *resource.Resources) []*gwapixv1a1.XBackendTrafficPolicy {
	res := make([]*gwapixv1a1.XBackendTrafficPolicy, 0, len(resources.XBackendTrafficPolicies))

	btpByRefSets := sets.New[string]()
	for _, xbtp := range resources.XBackendTrafficPolicies {
		err := t.validateXBackendTrafficPolicy(xbtp)
		for _, ref := range xbtp.Spec.TargetRefs {
			ancestor := gwapiv1.ParentReference{
				Group: ptr.To(ref.Group),
				Kind:  ptr.To(ref.Kind),
				Name:  ref.Name,
			}

			if err != nil {
				status.SetConditionForPolicyAncestor(&xbtp.Status,
					ancestor,
					t.GatewayControllerName,
					gwapiv1a2.PolicyConditionAccepted,
					metav1.ConditionFalse,
					gwapiv1a2.PolicyReasonInvalid,
					status.Error2ConditionMsg(err),
					xbtp.Generation,
				)
				continue
			}

			kind := KindDerefOr(&ref.Kind, resource.KindService)
			if err := validateXBackendTrafficPolicyTargetRef(resources, xbtp, &ref); err != nil {
				status.SetConditionForPolicyAncestor(&xbtp.Status,
					ancestor,
					t.GatewayControllerName,
					gwapiv1a2.PolicyConditionAccepted,
					metav1.ConditionFalse,
					gwapiv1a2.PolicyReasonTargetNotFound,
					status.Error2ConditionMsg(err),
					xbtp.Generation,
				)
				continue
			}

			// Check if the backend reference is already in the set.
			// Key format: /<kind>/<namespace>/<name>
			refKey := fmt.Sprintf("/%s/%s/%s", kind, xbtp.Namespace, ref.Name)
			if btpByRefSets.Has(refKey) {
				// More than one policy reference to the same backend reference.
				status.SetConditionForPolicyAncestor(&xbtp.Status,
					ancestor,
					t.GatewayControllerName,
					gwapiv1a2.PolicyConditionAccepted,
					metav1.ConditionFalse,
					gwapiv1a2.PolicyReasonConflicted,
					fmt.Sprintf("%s %s backend reference conflict.", kind, ref.Name),
					xbtp.Generation,
				)
				continue
			}
			btpByRefSets.Insert(refKey)

			status.SetConditionForPolicyAncestor(&xbtp.Status,
				ancestor,
				t.GatewayControllerName,
				gwapiv1a2.PolicyConditionAccepted,
				metav1.ConditionTrue,
				gwapiv1a2.PolicyReasonAccepted,
				"Policy has been accepted.",
				xbtp.Generation,
			)
		}

		res = append(res, xbtp)
	}

	return res
}

func (t *Translator) applyXBackendTrafficPolicy(backendRef *gwapiv1a2.NamespacedPolicyTargetReference,
	resources *resource.Resources,
) *ir.RetryBudget {
	policy := t.getXBackendTrafficPolicy(resources, backendRef)
	if policy == nil {
		return nil
	}

	return translateXBackendTrafficPolicyRetryBudget(policy)
}

func translateXBackendTrafficPolicyRetryBudget(btp *gwapixv1a1.XBackendTrafficPolicy) *ir.RetryBudget {
	if btp == nil || btp.Spec.RetryConstraint == nil {
		return nil
	}

	retry := btp.Spec.RetryConstraint
	var (
		p        *gwapiv1.Fraction
		minRetry *uint32
	)
	if retry.Budget != nil {
		num := int32(20)
		if retry.Budget.Percent != nil {
			num = int32(*retry.Budget.Percent)
		}

		p = &gwapiv1.Fraction{
			Numerator:   num,
			Denominator: ptr.To[int32](100),
		}
	}
	if retry.MinRetryRate != nil {
		minRetry = ptr.To[uint32](10)
		if retry.MinRetryRate.Count != nil {
			minRetry = ptr.To[uint32](uint32(*retry.MinRetryRate.Count))
		}
	}

	return &ir.RetryBudget{
		Percent:             p,
		MinRetryConcurrency: minRetry,
	}
}

func (t *Translator) getXBackendTrafficPolicy(resources *resource.Resources,
	backendRef *gwapiv1a2.NamespacedPolicyTargetReference,
) *gwapixv1a1.XBackendTrafficPolicy {
	ns := backendRefNamespace(backendRef)
	for _, btp := range resources.XBackendTrafficPolicies {
		if btp.Namespace != ns {
			continue
		}
		if err := t.validateXBackendTrafficPolicy(btp); err != nil {
			continue
		}

		for _, ref := range btp.Spec.TargetRefs {
			if err := validateXBackendTrafficPolicyTargetRef(resources, btp, &ref); err != nil {
				continue
			}

			if policyTargetRefMatched(backendRef, ref) {
				return btp
			}
		}
	}
	return nil
}

// policyTargetRefMatched checks if the backend reference matches the target reference.
func policyTargetRefMatched(left *gwapiv1a2.NamespacedPolicyTargetReference, right gwapiv1a2.LocalPolicyTargetReference) bool {
	if left.Name != right.Name {
		return false
	}

	if KindDerefOr(&left.Kind, resource.KindService) != KindDerefOr(&right.Kind, resource.KindService) {
		return false
	}

	if left.Group != right.Group {
		return false
	}

	return true
}

// backendRefNamespace returns the namespace of the backend reference.
// Please make sure the namespace is set.
func backendRefNamespace(backendRef *gwapiv1a2.NamespacedPolicyTargetReference) string {
	if backendRef == nil ||
		backendRef.Namespace == nil {
		return ""
	}
	return string(*backendRef.Namespace)
}

func (t *Translator) validateXBackendTrafficPolicy(btp *gwapixv1a1.XBackendTrafficPolicy) error {
	if btp == nil {
		return nil
	}

	if t.XBackendTrafficPolicyEnabled {
		return errors.New("XBackendTrafficPolicy is not enabled in Envoy Gateway")
	}

	if btp.Spec.SessionPersistence != nil {
		return errors.New("session persistence is not supported")
	}

	if retry := btp.Spec.RetryConstraint; retry != nil {
		if retry.MinRetryRate != nil && retry.MinRetryRate.Interval != nil {
			return errors.New("min retry rate interval is not supported")
		}

		if retry.Budget != nil && retry.Budget.Interval != nil {
			return errors.New("budget interval is not supported")
		}
	}

	return nil
}

func validateXBackendTrafficPolicyTargetRef(resources *resource.Resources,
	xbtp *gwapixv1a1.XBackendTrafficPolicy, ref *gwapiv1a2.LocalPolicyTargetReference,
) error {
	if ref == nil {
		return nil
	}

	kind := KindDerefOr(&ref.Kind, resource.KindService)
	switch kind {
	case resource.KindService:
		svc := resources.GetService(xbtp.Namespace, string(ref.Name))
		if svc == nil {
			return fmt.Errorf("service %s not found", ref.Name)
		}
	case resource.KindBackend:
		backend := resources.GetBackend(xbtp.Namespace, string(ref.Name))
		if backend == nil {
			return fmt.Errorf("backend %s not found", ref.Name)
		}
	case resource.KindServiceImport:
		svcImport := resources.GetServiceImport(xbtp.Namespace, string(ref.Name))
		if svcImport == nil {
			return fmt.Errorf("serviceImport %s not found", ref.Name)
		}
	default:
		return fmt.Errorf("unsupported kind: %s", kind)
	}

	return nil
}
