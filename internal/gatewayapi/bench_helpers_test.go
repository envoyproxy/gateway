// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

// benchBaseSecrets returns the base test secrets, which include the Envoy client TLS secret every
// translation needs.
func benchBaseSecrets(b *testing.B) []*corev1.Secret {
	b.Helper()
	data, err := os.ReadFile("testdata/base/base.yaml")
	if err != nil {
		b.Fatal(err)
	}
	base := &resource.Resources{}
	if err := yaml.Unmarshal(data, base); err != nil {
		b.Fatal(err)
	}
	return base.Secrets
}
