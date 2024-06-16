// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
)

func TestStart(t *testing.T) {
	logger := logging.DefaultLogger(egv1a1.LogLevelInfo)

	testCases := []struct {
		name   string
		cfg    *config.Server
		expect bool
	}{
		{
			name: "file provider",
			cfg: &config.Server{
				EnvoyGateway: &egv1a1.EnvoyGateway{
					TypeMeta: metav1.TypeMeta{
						APIVersion: egv1a1.GroupVersion.String(),
						Kind:       egv1a1.KindEnvoyGateway,
					},
					EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
						Provider: &egv1a1.EnvoyGatewayProvider{
							Type: egv1a1.ProviderTypeFile,
						},
					},
				},
				Logger: logger,
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runner := &Runner{
				Config: Config{
					Server:            *tc.cfg,
					ProviderResources: new(message.ProviderResources),
				},
			}
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			err := runner.Start(ctx)
			if tc.expect {
				require.NoError(t, err)
			} else {
				require.Error(t, err, "An error was expected")
			}
		})
	}
}
