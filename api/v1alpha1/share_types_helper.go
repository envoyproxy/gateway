// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

func ToBackendObjectReference(ref BackendRef) *gwapiv1.BackendObjectReference {
	return &gwapiv1.BackendObjectReference{
		Group:     ref.Group,
		Kind:      ref.Kind,
		Namespace: ref.Namespace,
		Name:      ref.Name,
		Port:      ref.Port,
	}
}
