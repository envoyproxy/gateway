// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package debug

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

func TestInitDebugServer(t *testing.T) {
	svrConfig := &config.Server{
		EnvoyGateway: &v1alpha1.EnvoyGateway{
			EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
				Debug: &v1alpha1.EnvoyGatewayDebug{
					EnableDumpConfig: true,
					EnablePprof:      true,
					Address: &v1alpha1.EnvoyGatewayDebugAddress{
						Host: "127.0.0.1",
						Port: 19010,
					},
				},
			},
		},
	}
	err := Init(svrConfig)
	assert.NoError(t, err)
}
