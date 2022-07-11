package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
)

var (
	inPath = "./testdata/decoder/in/"
)

func TestDecode(t *testing.T) {
	testCases := []struct {
		in     string
		out    *v1alpha1.EnvoyGateway
		expect bool
	}{
		{
			in: inPath + "kube-provider.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Provider: v1alpha1.DefaultProvider(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "gateway-controller-name.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway: v1alpha1.DefaultGateway(),
				},
			},
			expect: true,
		},
		{
			in:     inPath + "no-api-version.yaml",
			expect: false,
		},
		{
			in:     inPath + "no-kind.yaml",
			expect: false,
		},
		{
			in:     "/non/existent/config.yaml",
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			eg, err := Decode(tc.in)
			if tc.expect {
				require.NoError(t, err)
				require.Equal(t, tc.out, eg)
			} else {
				require.Error(t, err, "An error was expected")
			}
		})
	}
}
