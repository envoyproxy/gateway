// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestSetByObjectReplacesSameType(t *testing.T) {
	m := map[client.Object]cache.ByObject{
		&corev1.Secret{}:    {UnsafeDisableDeepCopy: new(true)},
		&corev1.ConfigMap{}: {},
	}
	ns := map[string]cache.Config{"a": {}}
	setByObject(m, &corev1.Secret{}, cache.ByObject{Namespaces: ns})

	require.Len(t, m, 2)
	for k, v := range m {
		if _, ok := k.(*corev1.Secret); ok {
			require.Equal(t, ns, v.Namespaces)
		}
	}
}
