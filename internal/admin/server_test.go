// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package admin

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

func TestInitAdminServer(t *testing.T) {
	svrConfig := &config.Server{
		EnvoyGateway: &v1alpha1.EnvoyGateway{
			EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{},
		},
	}
	err := Init(svrConfig)
	require.NoError(t, err)
}
