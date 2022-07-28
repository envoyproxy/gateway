package ir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateInfra(t *testing.T) {
	testCases := []struct {
		name   string
		infra  *Infra
		expect bool
	}{
		{
			name: "valid",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:      "test",
					Namespace: "test",
					Image:     "image",
				},
			},
			expect: true,
		},
		{
			name: "no-name",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:      "",
					Namespace: "test",
					Image:     "image",
				},
			},
			expect: false,
		},
		{
			name: "no-namespace",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:      "test",
					Namespace: "",
					Image:     "image",
				},
			},
			expect: false,
		},
		{
			name: "no-image",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:      "test",
					Namespace: "test",
					Image:     "",
				},
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateInfra(tc.infra)
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
