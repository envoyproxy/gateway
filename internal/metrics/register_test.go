// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
)

func TestMetricServer(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: &egv1a1.EnvoyGateway{
			EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{},
		},
		Logger: logging.NewLogger(os.Stdout, egv1a1.DefaultEnvoyGatewayLogging()),
	}

	runner := New(cfg)
	err := runner.Start(context.Background())
	require.NoError(t, err)

	// Clean up
	err = runner.Close()
	require.NoError(t, err)
}
