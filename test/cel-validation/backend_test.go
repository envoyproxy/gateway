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
					BackendAddresses: []egv1a1.BackendAddress{
						{
							Unix: &egv1a1.UnixSocket{
								Path: "/path/to/service.sock",
							},
						},
						{
							IP: &egv1a1.IPAddress{
								Host: "1.1.1.1",
								Port: 443,
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
					BackendAddresses: []egv1a1.BackendAddress{
						{
							FQDN: &egv1a1.FQDNAddress{
								Host: "example.com",
								Port: 443,
							},
						},
						{
							FQDN: &egv1a1.FQDNAddress{
								Host: "example2.com",
								Port: 443,
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
					AppProtocols: []egv1a1.AppProtocolType{egv1a1.AppProtocolType("HTTP7")},
					BackendAddresses: []egv1a1.BackendAddress{
						{
							FQDN: &egv1a1.FQDNAddress{
								Host: "example.com",
								Port: 443,
							},
						},
					},
				}
			},
			wantErrors: []string{"spec.applicationProtocol[0]: Unsupported value: \"HTTP7\": supported values: \"gateway.envoyproxy.io/h2c\", \"gateway.envoyproxy.io/ws\", \"gateway.envoyproxy.io/wss\""},
		},
		{
			desc: "No address",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					AppProtocols:     []egv1a1.AppProtocolType{egv1a1.AppProtocolTypeH2C},
					BackendAddresses: []egv1a1.BackendAddress{{}},
				}
			},
			wantErrors: []string{"spec.addresses[0]: Invalid value: \"object\": one of fqdn, ip or unix must be specified"},
		},
		{
			desc: "Multiple addresses",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					AppProtocols: []egv1a1.AppProtocolType{egv1a1.AppProtocolTypeH2C},
					BackendAddresses: []egv1a1.BackendAddress{
						{
							FQDN: &egv1a1.FQDNAddress{
								Host: "example.com",
								Port: 443,
							},
							Unix: &egv1a1.UnixSocket{
								Path: "/path/to/service.sock",
							},
						},
					},
				}
			},
			wantErrors: []string{"spec.addresses[0]: Invalid value: \"object\": only one of fqdn, ip or unix can be specified"},
		},
		{
			desc: "Mixed types",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					AppProtocols: []egv1a1.AppProtocolType{egv1a1.AppProtocolTypeH2C},
					BackendAddresses: []egv1a1.BackendAddress{
						{
							FQDN: &egv1a1.FQDNAddress{
								Host: "example.com",
								Port: 443,
							},
						},
						{
							IP: &egv1a1.IPAddress{
								Host: "1.1.1.1",
								Port: 443,
							},
						},
					},
				}
			},
			wantErrors: []string{"spec.addresses: Invalid value: \"array\": FQDN addresses cannot be mixed with other address types"},
		},
		{
			desc: "Invalid hostname",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					AppProtocols: []egv1a1.AppProtocolType{egv1a1.AppProtocolTypeH2C},
					BackendAddresses: []egv1a1.BackendAddress{
						{
							FQDN: &egv1a1.FQDNAddress{
								Host: "host name",
								Port: 443,
							},
						},
						{
							FQDN: &egv1a1.FQDNAddress{
								Host: "host_name",
								Port: 443,
							},
						},
						{
							FQDN: &egv1a1.FQDNAddress{
								Host: "hostname:443",
								Port: 443,
							},
						},
						{
							FQDN: &egv1a1.FQDNAddress{
								Host: "host.*.name",
								Port: 443,
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.addresses[0].fqdn.host: Invalid value: \"host name\": spec.addresses[0].fqdn.host in body should match '^(\\*\\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'",
				"spec.addresses[1].fqdn.host: Invalid value: \"host_name\": spec.addresses[1].fqdn.host in body should match '^(\\*\\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'",
				"spec.addresses[2].fqdn.host: Invalid value: \"hostname:443\": spec.addresses[2].fqdn.host in body should match '^(\\*\\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'",
				"spec.addresses[3].fqdn.host: Invalid value: \"host.*.name\": spec.addresses[3].fqdn.host in body should match '^(\\*\\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'",
			},
		},
		{
			desc: "Invalid IP",
			mutate: func(backend *egv1a1.Backend) {
				backend.Spec = egv1a1.BackendSpec{
					AppProtocols: []egv1a1.AppProtocolType{egv1a1.AppProtocolTypeH2C},
					BackendAddresses: []egv1a1.BackendAddress{
						{
							IP: &egv1a1.IPAddress{
								Host: "300.0.0.0",
								Port: 443,
							},
						},
						{
							IP: &egv1a1.IPAddress{
								Host: "0.0.0.0:443",
								Port: 443,
							},
						},
						{
							IP: &egv1a1.IPAddress{
								Host: "0.0.0.0/12",
								Port: 443,
							},
						},
						{
							IP: &egv1a1.IPAddress{
								Host: "a.b.c.e",
								Port: 443,
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.addresses[0].ip.host: Invalid value: \"300.0.0.0\": spec.addresses[0].ip.host in body should match '^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$'",
				"spec.addresses[1].ip.host: Invalid value: \"0.0.0.0:443\": spec.addresses[1].ip.host in body should match '^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$'",
				"spec.addresses[2].ip.host: Invalid value: \"0.0.0.0/12\": spec.addresses[2].ip.host in body should match '^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$'",
				"spec.addresses[3].ip.host: Invalid value: \"a.b.c.e\": spec.addresses[3].ip.host in body should match '^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$'",
			},
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
