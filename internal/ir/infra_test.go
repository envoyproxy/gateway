// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestValidateInfra(t *testing.T) {
	testCases := []struct {
		name   string
		infra  *Infra
		expect bool
	}{
		{
			name:   "default",
			infra:  NewInfra(),
			expect: true,
		},
		{
			name: "no-name",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:      "",
					Listeners: NewProxyListeners(),
				},
			},
			expect: false,
		},
		{
			name: "no-listeners",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name: "test",
				},
			},
			expect: true,
		},
		{
			name: "no-listener-ports",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name: "test",
					Listeners: []*ProxyListener{
						{
							Ports: []ListenerPort{},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "no-listener-port-name",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name: "test",
					Listeners: []*ProxyListener{
						{
							Ports: []ListenerPort{
								{
									ServicePort:   int32(80),
									ContainerPort: int32(8080),
								},
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "no-listener-service-port-number",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name: "test",
					Listeners: []*ProxyListener{
						{
							Ports: []ListenerPort{
								{
									Name:          "http",
									ContainerPort: int32(8080),
								},
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "no-listener-container-port-number",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name: "test",
					Listeners: []*ProxyListener{
						{
							Ports: []ListenerPort{
								{
									Name:        "http",
									ServicePort: int32(80),
								},
							},
						},
					},
				},
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.infra.Validate()
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewInfra(t *testing.T) {
	testCases := []struct {
		name     string
		expected *Infra
	}{
		{
			name: "default infra",
			expected: &Infra{
				Proxy: NewProxyInfra(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			actual := NewInfra()
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestNewProxyInfra(t *testing.T) {
	testCases := []struct {
		name     string
		expected *ProxyInfra
	}{
		{
			name: "default infra",
			expected: &ProxyInfra{
				Metadata: NewInfraMetadata(),
				Name:     DefaultProxyName,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			actual := NewProxyInfra()
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestObjectName(t *testing.T) {
	defaultInfra := NewInfra()

	testCases := []struct {
		name     string
		infra    *Infra
		expected string
	}{
		{
			name:     "default infra",
			infra:    defaultInfra,
			expected: "envoy-default",
		},
		{
			name: "defined infra",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name: "foo",
				},
			},
			expected: "envoy-foo",
		},
		{
			name: "unspecified infra name",
			infra: &Infra{
				Proxy: &ProxyInfra{},
			},
			expected: "envoy-default",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.infra.Proxy.ObjectName()
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestEqualInfra(t *testing.T) {
	tests := []struct {
		desc  string
		a     *ProxyInfra
		b     *ProxyInfra
		equal bool
	}{
		{
			desc: "out of order proxy listeners are equal",
			a: &ProxyInfra{
				Listeners: []*ProxyListener{
					{Name: "listener-1"},
					{Name: "listener-2"},
				},
			},
			b: &ProxyInfra{
				Listeners: []*ProxyListener{
					{Name: "listener-2"},
					{Name: "listener-1"},
				},
			},
			equal: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, tc.equal, cmp.Equal(tc.a, tc.b))
		})
	}
}
