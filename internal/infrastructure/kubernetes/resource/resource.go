// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
)

// GetSelector returns a label selector used to select resources
// based on the provided labels.
func GetSelector(labels map[string]string) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: labels,
	}
}

// ExpectedServiceSpec returns service spec.
func ExpectedServiceSpec(serviceType *egcfgv1a1.ServiceType) corev1.ServiceSpec {
	serviceSpec := corev1.ServiceSpec{}
	serviceSpec.Type = corev1.ServiceType(*serviceType)
	serviceSpec.SessionAffinity = corev1.ServiceAffinityNone
	if *serviceType == egcfgv1a1.ServiceTypeLoadBalancer {
		// Preserve the client source IP and avoid a second hop for LoadBalancer.
		serviceSpec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyTypeLocal
	}
	return serviceSpec
}

// CompareSvc Only compare the type, selector and ports(not include nodePort) in case user have modified for some scene.
func CompareSvc(currentSvc, originalSvc *corev1.Service) bool {
	return cmp.Equal(currentSvc.Spec, originalSvc.Spec,
		cmpopts.IgnoreFields(corev1.ServicePort{}, "NodePort"),
		cmpopts.IgnoreFields(corev1.ServiceSpec{}, "ClusterIP", "ClusterIPs"))
}
