// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build celvalidation

package celvalidation

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestBackend(t *testing.T) {
	ctx := context.Background()
	baseBackend := egv1a1.Backend{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "backend",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: egv1a1.BackendSpec{},
	}

	cases := []struct {
		desc         string
		mutate       func(backend *egv1a1.Backend)
		mutateStatus func(backend *egv1a1.Backend)
		wantErrors   []string
	}{
		{
			desc: "Valid static",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					AppProtocols: []egv1a1.AppProtocolType{egv1a1.AppProtocolTypeH2C},
					Endpoints: []egv1a1.BackendEndpoint{
						{
							Unix: &egv1a1.UnixSocket{
								Path: "/path/to/service.sock",
							},
						},
						{
							IP: &egv1a1.IPEndpoint{
								Address: "1.1.1.1",
								Port:    443,
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "Valid DNS",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					AppProtocols: []egv1a1.AppProtocolType{egv1a1.AppProtocolTypeH2C},
					Endpoints: []egv1a1.BackendEndpoint{
						{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "example.com",
								Port:     443,
							},
						},
						{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "example2.com",
								Port:     443,
							},
						},
						{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "sub.example.com",
								Port:     443,
							},
						},
						{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "sub1.sub.sub.example.com",
								Port:     443,
							},
						},
						{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "sub.s.example.com",
								Port:     443,
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "unsupported application protocol type",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					AppProtocols: []egv1a1.AppProtocolType{"HTTP7"},
					Endpoints: []egv1a1.BackendEndpoint{
						{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "example.com",
								Port:     443,
							},
						},
					},
				}
			},
			wantErrors: []string{"spec.appProtocols[0]: Unsupported value: \"HTTP7\": supported values: \"gateway.envoyproxy.io/h2c\", \"gateway.envoyproxy.io/ws\", \"gateway.envoyproxy.io/wss\""},
		},
		{
			desc: "No address",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					AppProtocols: []egv1a1.AppProtocolType{egv1a1.AppProtocolTypeH2C},
					Endpoints:    []egv1a1.BackendEndpoint{{}},
				}
			},
			wantErrors: []string{"spec.endpoints[0]: Invalid value: \"object\": one of fqdn, ip or unix must be specified"},
		},
		{
			desc: "Multiple addresses",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					AppProtocols: []egv1a1.AppProtocolType{egv1a1.AppProtocolTypeH2C},
					Endpoints: []egv1a1.BackendEndpoint{
						{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "example.com",
								Port:     443,
							},
							Unix: &egv1a1.UnixSocket{
								Path: "/path/to/service.sock",
							},
						},
					},
				}
			},
			wantErrors: []string{"spec.endpoints[0]: Invalid value: \"object\": only one of fqdn, ip or unix can be specified"},
		},
		{
			desc: "Mixed types",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					AppProtocols: []egv1a1.AppProtocolType{egv1a1.AppProtocolTypeH2C},
					Endpoints: []egv1a1.BackendEndpoint{
						{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "example.com",
								Port:     443,
							},
						},
						{
							IP: &egv1a1.IPEndpoint{
								Address: "1.1.1.1",
								Port:    443,
							},
						},
					},
				}
			},
			wantErrors: []string{"spec.endpoints: Invalid value: \"array\": FQDN addresses cannot be mixed with other address types"},
		},
		{
			desc: "Invalid hostname",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					AppProtocols: []egv1a1.AppProtocolType{egv1a1.AppProtocolTypeH2C},
					Endpoints: []egv1a1.BackendEndpoint{
						{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "host name",
								Port:     443,
							},
						},
						{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "host_name",
								Port:     443,
							},
						},
						{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "hostname:443",
								Port:     443,
							},
						},
						{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "host.*.name",
								Port:     443,
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.endpoints[0].fqdn.hostname: Invalid value: \"host name\": spec.endpoints[0].fqdn.hostname in body should match '^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'",
				"spec.endpoints[1].fqdn.hostname: Invalid value: \"host_name\": spec.endpoints[1].fqdn.hostname in body should match '^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'",
				"spec.endpoints[2].fqdn.hostname: Invalid value: \"hostname:443\": spec.endpoints[2].fqdn.hostname in body should match '^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$",
				"spec.endpoints[3].fqdn.hostname: Invalid value: \"host.*.name\": spec.endpoints[3].fqdn.hostname in body should match '^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$",
			},
		},
		{
			desc: "Invalid IP",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					AppProtocols: []egv1a1.AppProtocolType{egv1a1.AppProtocolTypeH2C},
					Endpoints: []egv1a1.BackendEndpoint{
						{
							IP: &egv1a1.IPEndpoint{
								Address: "300.0.0.0",
								Port:    443,
							},
						},
						{
							IP: &egv1a1.IPEndpoint{
								Address: "0.0.0.0:443",
								Port:    443,
							},
						},
						{
							IP: &egv1a1.IPEndpoint{
								Address: "0.0.0.0/12",
								Port:    443,
							},
						},
						{
							IP: &egv1a1.IPEndpoint{
								Address: "a.b.c.e",
								Port:    443,
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.endpoints[0].ip.address: Invalid value: \"300.0.0.0\": spec.endpoints[0].ip.address in body should match '^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$|^(([0-9a-fA-F]{1,4}:){1,7}[0-9a-fA-F]{1,4}|::|(([0-9a-fA-F]{1,4}:){0,5})?(:[0-9a-fA-F]{1,4}){1,2})$'",
				"spec.endpoints[1].ip.address: Invalid value: \"0.0.0.0:443\": spec.endpoints[1].ip.address in body should match '^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$|^(([0-9a-fA-F]{1,4}:){1,7}[0-9a-fA-F]{1,4}|::|(([0-9a-fA-F]{1,4}:){0,5})?(:[0-9a-fA-F]{1,4}){1,2})$'",
				"spec.endpoints[2].ip.address: Invalid value: \"0.0.0.0/12\": spec.endpoints[2].ip.address in body should match '^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$|^(([0-9a-fA-F]{1,4}:){1,7}[0-9a-fA-F]{1,4}|::|(([0-9a-fA-F]{1,4}:){0,5})?(:[0-9a-fA-F]{1,4}){1,2})$'",
				"spec.endpoints[3].ip.address: Invalid value: \"a.b.c.e\": spec.endpoints[3].ip.address in body should match '^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$|^(([0-9a-fA-F]{1,4}:){1,7}[0-9a-fA-F]{1,4}|::|(([0-9a-fA-F]{1,4}:){0,5})?(:[0-9a-fA-F]{1,4}){1,2})$'",
			},
		},
		{
			desc: "invalid type",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{Type: ptr.To[egv1a1.BackendType]("FOO")}
			},
			wantErrors: []string{`spec.type: Unsupported value: "FOO": supported values: "Endpoints", "DynamicResolver"`},
		},
		{
			desc: "dynamic resolver ok",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					Type:         ptr.To(egv1a1.BackendTypeDynamicResolver),
					AppProtocols: []egv1a1.AppProtocolType{egv1a1.AppProtocolTypeH2C},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "dynamic resolver invalid",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					Type: ptr.To(egv1a1.BackendTypeDynamicResolver),
					Endpoints: []egv1a1.BackendEndpoint{
						{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "example.com",
								Port:     443,
							},
						},
					},
				}
			},
			wantErrors: []string{"DynamicResolver type cannot have endpoints specified"},
		},
		{
			desc: "Invalid Unix socket path length",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					Type:         ptr.To(egv1a1.BackendTypeEndpoints),
					AppProtocols: []egv1a1.AppProtocolType{egv1a1.AppProtocolTypeH2C},
					Endpoints: []egv1a1.BackendEndpoint{
						{
							Unix: &egv1a1.UnixSocket{
								Path: "/path/to/a/very/long/unix/socket/path/that/exceeds/the/maximum/allowed/length/of/108/characters/and/should/fail/validation.sock",
							},
						},
					},
				}
			},
			wantErrors: []string{`spec.endpoints[0].unix.path: Invalid value: "string": unix domain socket path must not exceed 108 characters`},
		},
		{
			desc: "dynamic resolver invalid WellKnownCACertificates and InsecureSkipVerify specified",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					Type: ptr.To(egv1a1.BackendTypeDynamicResolver),
					TLS: &egv1a1.BackendTLSSettings{
						InsecureSkipVerify:      ptr.To(true),
						WellKnownCACertificates: ptr.To(gwapiv1a3.WellKnownCACertificatesSystem),
					},
				}
			},
			wantErrors: []string{`must not contain either CACertificateRefs or WellKnownCACertificates when InsecureSkipVerify is enabled`},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			backend := baseBackend.DeepCopy()
			backend.Name = fmt.Sprintf("backend-%v", time.Now().UnixNano())

			if tc.mutate != nil {
				tc.mutate(backend)
			}
			err := c.Create(ctx, backend)

			if tc.mutateStatus != nil {
				tc.mutateStatus(backend)
				err = c.Status().Update(ctx, backend)
			}

			if (len(tc.wantErrors) != 0) != (err != nil) {
				t.Fatalf("Unexpected response while creating Backend; got err=\n%v\n;want error=%v", err, tc.wantErrors)
			}

			var missingErrorStrings []string
			for _, wantError := range tc.wantErrors {
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(wantError)) {
					missingErrorStrings = append(missingErrorStrings, wantError)
				}
			}
			if len(missingErrorStrings) != 0 {
				t.Errorf("Unexpected response while creating Backend; got err=\n%v\n;missing strings within error=%q", err, missingErrorStrings)
			}
		})
	}
}
