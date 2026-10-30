// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package version

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestGetEnvoyProxyVersion(t *testing.T) {
	got := envoyProxyVersion("docker.io/envoyproxy/envoy:distroless-v1.38.8@sha256:7ddf5ed97b84056026bbe2cf6f74e717d690bb7062e5a8ddc1bdd0e2015872e2")
	assert.Equal(t, "distroless-v1.38.8@sha256:7ddf5ed97b84056026bbe2cf6f74e717d690bb7062e5a8ddc1bdd0e2015872e2", got)

	got = envoyProxyVersion(egv1a1.DefaultEnvoyProxyImage)
	assert.False(t, strings.HasSuffix(got, "@sha256"))
}
