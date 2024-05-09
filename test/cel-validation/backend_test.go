// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build celvalidation
// +build celvalidation

package celvalidation

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

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
					ApplicationProtocol: ptr.To(egv1a1.ApplicationProtocolTypeHTTP2),
					BackendAddresses: []egv1a1.BackendAddress{
						{
							Type: "UDS",
							UnixDomainSocketAddress: &egv1a1.UnixDomainSocketAddress{
								Path: "/path/to/service.sock",
							},
						},
						{
							Type: "IPv4",
							SocketAddress: &egv1a1.SocketAddress{
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
					ApplicationProtocol: ptr.To(egv1a1.ApplicationProtocolTypeHTTP2),
					BackendAddresses: []egv1a1.BackendAddress{
						{
							Type: "FQDN",
							SocketAddress: &egv1a1.SocketAddress{
								Address: "example.com",
								Port:    443,
							},
						},
						{
							Type: "FQDN",
							SocketAddress: &egv1a1.SocketAddress{
								Address: "example2.com",
								Port:    443,
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "unsupported address type",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					BackendAddresses: []egv1a1.BackendAddress{
						{
							Type:          "not-a-type",
							SocketAddress: &egv1a1.SocketAddress{},
						},
					},
				}
			},
			wantErrors: []string{"Unsupported value: \"not-a-type\": supported values: \"FQDN\", \"UDS\", \"IPv4\""},
		},
		{
			desc: "unsupported application protocol type",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					ApplicationProtocol: ptr.To(egv1a1.ApplicationProtocolType("HTTP7")),
					BackendAddresses: []egv1a1.BackendAddress{
						{
							Type: "FQDN",
							SocketAddress: &egv1a1.SocketAddress{
								Address: "example.com",
								Port:    443,
							},
						},
					},
				}
			},
			wantErrors: []string{"Unsupported value: \"HTTP7\": supported values: \"HTTP2\", \"WS\""},
		},
		{
			desc: "unsupported transport protocol type",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					ApplicationProtocol: ptr.To(egv1a1.ApplicationProtocolTypeHTTP2),
					BackendAddresses: []egv1a1.BackendAddress{
						{
							Type: "FQDN",
							SocketAddress: &egv1a1.SocketAddress{
								Address:  "example.com",
								Port:     443,
								Protocol: ptr.To(egv1a1.ProtocolType("TDP")),
							},
						},
					},
				}
			},
			wantErrors: []string{"Unsupported value: \"TDP\": supported values: \"TCP\", \"UDP\""},
		},
		{
			desc: "No address",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					ApplicationProtocol: ptr.To(egv1a1.ApplicationProtocolTypeHTTP2),
					BackendAddresses: []egv1a1.BackendAddress{
						{
							Type: "FQDN",
						},
					},
				}
			},
			wantErrors: []string{"[spec.addresses[0]: Invalid value: \"object\": one of socketAddress or unixDomainSocketAddress must be specified"},
		},
		{
			desc: "Both addresses",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					ApplicationProtocol: ptr.To(egv1a1.ApplicationProtocolTypeHTTP2),
					BackendAddresses: []egv1a1.BackendAddress{
						{
							Type: "FQDN",
							SocketAddress: &egv1a1.SocketAddress{
								Address: "example.com",
								Port:    443,
							},
							UnixDomainSocketAddress: &egv1a1.UnixDomainSocketAddress{
								Path: "/path/to/service.sock",
							},
						},
					},
				}
			},
			wantErrors: []string{"spec.addresses[0]: Invalid value: \"object\": only one of socketAddress or unixDomainSocketAddress can be specified"},
		},
		{
			desc: "Socket with wrong type",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					ApplicationProtocol: ptr.To(egv1a1.ApplicationProtocolTypeHTTP2),
					BackendAddresses: []egv1a1.BackendAddress{
						{
							Type: "UDS",
							SocketAddress: &egv1a1.SocketAddress{
								Address: "example.com",
								Port:    443,
							},
						},
					},
				}
			},
			wantErrors: []string{"spec.addresses[0]: Invalid value: \"object\": if type is FQDN or IPv4, socketAddress must be set; if type is UDS, unixDomainSocketAddress must be set"},
		},
		{
			desc: "Unix socket with wrong type",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					ApplicationProtocol: ptr.To(egv1a1.ApplicationProtocolTypeHTTP2),
					BackendAddresses: []egv1a1.BackendAddress{
						{
							Type: "FQDN",
							UnixDomainSocketAddress: &egv1a1.UnixDomainSocketAddress{
								Path: "/path/to/service.sock",
							},
						},
					},
				}
			},
			wantErrors: []string{"spec.addresses[0]: Invalid value: \"object\": if type is FQDN or IPv4, socketAddress must be set; if type is UDS, unixDomainSocketAddress must be set"},
		},
		{
			desc: "Mixed types",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					ApplicationProtocol: ptr.To(egv1a1.ApplicationProtocolTypeHTTP2),
					BackendAddresses: []egv1a1.BackendAddress{
						{
							Type: "FQDN",
							SocketAddress: &egv1a1.SocketAddress{
								Address: "example.com",
								Port:    443,
							},
						},
						{
							Type: "IPv4",
							SocketAddress: &egv1a1.SocketAddress{
								Address: "1.1.1.1",
								Port:    443,
							},
						},
					},
				}
			},
			wantErrors: []string{"spec.addresses: Invalid value: \"array\": FQDN addresses cannot be mixed with other address types"},
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
