package runner

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/log"
	"github.com/envoyproxy/gateway/internal/message"
)

func TestStart(t *testing.T) {
	logger, err := log.NewLogger()
	require.NoError(t, err)

	testCases := []struct {
		name   string
		cfg    *config.Server
		expect bool
	}{
		{
			name: "file provider",
			cfg: &config.Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					TypeMeta: metav1.TypeMeta{
						APIVersion: v1alpha1.GroupVersion.String(),
						Kind:       v1alpha1.KindEnvoyGateway,
					},
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Provider: &v1alpha1.Provider{
							Type: v1alpha1.ProviderTypeFile,
						},
					},
				},
				Logger: logger,
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runner := &Runner{
				Config: Config{
					Server:            *tc.cfg,
					ProviderResources: new(message.ProviderResources),
				},
			}
			ctx := ctrl.SetupSignalHandler()
			err := runner.Start(ctx)
			if tc.expect {
				require.NoError(t, err)
			} else {
				require.Error(t, err, "An error was expected")
			}
		})
	}
}
