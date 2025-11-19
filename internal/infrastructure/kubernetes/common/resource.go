// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package common

import (
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// ResourceKind indicates the main resources of envoy-ratelimit,
// but also the key for the uid of their ownerReference.
const (
	ResourceKindDeployment = "Deployment"
)

func GetPodDisruptionBudget(pdb *egv1a1.KubernetesPodDisruptionBudgetSpec,
	selector *metav1.LabelSelector, nn *types.NamespacedName, ownerReferences []metav1.OwnerReference,
) (*policyv1.PodDisruptionBudget, error) {
	// If podDisruptionBudget config is nil, ignore PodDisruptionBudget.
	if pdb == nil {
		return nil, nil
	}

	pdbSpec := policyv1.PodDisruptionBudgetSpec{
		Selector: selector,
	}

	switch {
	case pdb.MinAvailable != nil:
		pdbSpec.MinAvailable = pdb.MinAvailable
	case pdb.MaxUnavailable != nil:
		pdbSpec.MaxUnavailable = pdb.MaxUnavailable
	default:
		pdbSpec.MinAvailable = &intstr.IntOrString{Type: intstr.Int, IntVal: 0}
	}

	podDisruptionBudget := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name,
			Namespace: nn.Namespace,
			Labels:    selector.MatchLabels,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy/v1",
			Kind:       "PodDisruptionBudget",
		},
		Spec: pdbSpec,
	}

	podDisruptionBudget.OwnerReferences = append(podDisruptionBudget.OwnerReferences, ownerReferences...)

	// apply merge patch to PodDisruptionBudget
	podDisruptionBudget, err := pdb.ApplyMergePatch(podDisruptionBudget)
	if err != nil {
		return nil, err
	}
	return podDisruptionBudget, nil
}
