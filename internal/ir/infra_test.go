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
			name:   "default",
			infra:  NewInfra(),
			expect: false,
		},
		{
			name: "no-proxy-infra-name",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:      "",
					Image:     "image",
					Listeners: NewProxyListeners(),
				},
			},
			expect: false,
		},
		{
			name: "no-listeners",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:  "test",
					Image: "image",
				},
			},
			expect: true,
		},
		{
			name: "no-listener-ports",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:  "test",
					Image: "image",
					Listeners: []ProxyListener{
						{
							Name:    "test",
							Ports:   []ListenerPort{},
							Address: "1.2.3.4",
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "no-listener-name",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:  "test",
					Image: "image",
					Listeners: []ProxyListener{
						{
							Ports: []ListenerPort{
								{
									ServicePort:   int32(80),
									ContainerPort: int32(8080),
								},
							},
							Address: "1.2.3.4",
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "one-listener-two-ports",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:  "test",
					Image: "image",
					Listeners: []ProxyListener{
						{
							Name: "gateway-1",
							Ports: []ListenerPort{
								{
									Name:          "http-1",
									ServicePort:   int32(80),
									ContainerPort: int32(8080),
								},
								{
									Name:          "http-2",
									ServicePort:   int32(81),
									ContainerPort: int32(8081),
								},
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "two-listeners",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:  "test",
					Image: "image",
					Listeners: []ProxyListener{
						{
							Name: "gateway-1",
							Ports: []ListenerPort{
								{
									Name:          "http-1",
									ServicePort:   int32(80),
									ContainerPort: int32(8080),
								},
							},
						},
						{
							Name: "gateway-2",
							Ports: []ListenerPort{
								{
									Name:          "http-1",
									ServicePort:   int32(81),
									ContainerPort: int32(8081),
								},
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "no-listener-port-name",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:  "test",
					Image: "image",
					Listeners: []ProxyListener{
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
					Name:  "test",
					Image: "image",
					Listeners: []ProxyListener{
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
					Name:  "test",
					Image: "image",
					Listeners: []ProxyListener{
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

		{
			name: "no-image",
			infra: &Infra{
				Proxy: &ProxyInfra{
					Name:  "test",
					Image: "",
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
				Metadata:  NewInfraMetadata(),
				Name:      DefaultProxyName,
				Image:     DefaultProxyImage,
				Listeners: NewProxyListeners(),
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
