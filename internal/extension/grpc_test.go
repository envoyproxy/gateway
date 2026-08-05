// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package extension

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/api/resource"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
)

func Test_GenerateGRPCOptions(t *testing.T) {
	type args struct {
		ext *egv1a1.ExtensionManager
	}
	tests := []struct {
		name    string
		args    args
		want    []grpc.DialOption
		wantErr bool
	}{
		{
			args: args{
				ext: &egv1a1.ExtensionManager{
					MaxMessageSize: new(resource.MustParse(fmt.Sprintf("%dM", math.MaxInt))),
					Service: &egv1a1.ExtensionService{
						BackendEndpoint: egv1a1.BackendEndpoint{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "foo.bar",
								Port:     44344,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			args: args{
				ext: &egv1a1.ExtensionManager{
					MaxMessageSize: new(resource.MustParse(fmt.Sprintf("%dM", 0))),
					Service: &egv1a1.ExtensionService{
						BackendEndpoint: egv1a1.BackendEndpoint{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "foo.bar",
								Port:     44344,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			args: args{
				ext: &egv1a1.ExtensionManager{
					MaxMessageSize: new(resource.MustParse(fmt.Sprintf("%dM", 10))),
					Service: &egv1a1.ExtensionService{
						BackendEndpoint: egv1a1.BackendEndpoint{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "foo.bar",
								Port:     44344,
							},
						},
						Retry: &egv1a1.ExtensionServiceRetry{
							MaxAttempts:    new(20),
							InitialBackoff: new(gwapiv1.Duration("500ms")),
							MaxBackoff:     new(gwapiv1.Duration("5s")),
							BackoffMultiplier: &gwapiv1.Fraction{
								Numerator: 50,
							},
							RetryableStatusCodes: []egv1a1.RetryableGRPCStatusCode{
								"CANCELLED",
								"UNKNOWN",
								"INVALID_ARGUMENT",
								"DEADLINE_EXCEEDED",
								"NOT_FOUND",
								"ALREADY_EXISTS",
								"PERMISSION_DENIED",
								"RESOURCE_EXHAUSTED",
								"FAILED_PRECONDITION",
								"ABORTED",
								"OUT_OF_RANGE",
								"UNIMPLEMENTED",
								"INTERNAL",
								"UNAVAILABLE",
								"DATA_LOSS",
								"UNAUTHENTICATED",
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
			_, err := GenerateGRPCOptions(context.TODO(), fc, tt.args.ext.Service, tt.args.ext.MaxMessageSize, "test-svc", "envoy-gateway-system")
			if (err != nil) != tt.wantErr {
				t.Errorf("setupGRPCOpts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_buildServiceConfig(t *testing.T) {
	type args struct {
		extSvc *egv1a1.ExtensionService
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				extSvc: &egv1a1.ExtensionService{
					BackendEndpoint: egv1a1.BackendEndpoint{
						FQDN: &egv1a1.FQDNEndpoint{
							Hostname: "foo.bar",
							Port:     44344,
						},
					},
				},
			},
			want: `{
"methodConfig": [{
	"name": [{"service": "envoygateway.extension.EnvoyGatewayExtension"}],
	"waitForReady": true,
	"retryPolicy": {
		"MaxAttempts": 4,
		"InitialBackoff": "0.100000s",
		"MaxBackoff": "1.000000s",
		"BackoffMultiplier": 2.000000,
		"RetryableStatusCodes": [ "UNAVAILABLE" ]
	}
}]}`,
		},
		{
			name: "valid",
			args: args{
				extSvc: &egv1a1.ExtensionService{
					BackendEndpoint: egv1a1.BackendEndpoint{
						FQDN: &egv1a1.FQDNEndpoint{
							Hostname: "foo.bar",
							Port:     44344,
						},
					},
					Retry: &egv1a1.ExtensionServiceRetry{
						MaxAttempts:    new(20),
						InitialBackoff: new(gwapiv1.Duration("500ms")),
						MaxBackoff:     new(gwapiv1.Duration("5s")),
						BackoffMultiplier: &gwapiv1.Fraction{
							Numerator: 50,
						},
						RetryableStatusCodes: []egv1a1.RetryableGRPCStatusCode{
							"CANCELLED",
							"UNKNOWN",
							"INVALID_ARGUMENT",
							"DEADLINE_EXCEEDED",
							"NOT_FOUND",
							"ALREADY_EXISTS",
							"PERMISSION_DENIED",
							"RESOURCE_EXHAUSTED",
							"FAILED_PRECONDITION",
							"ABORTED",
							"OUT_OF_RANGE",
							"UNIMPLEMENTED",
							"INTERNAL",
							"UNAVAILABLE",
							"DATA_LOSS",
							"UNAUTHENTICATED",
						},
					},
				},
			},
			want: `{
"methodConfig": [{
	"name": [{"service": "envoygateway.extension.EnvoyGatewayExtension"}],
	"waitForReady": true,
	"retryPolicy": {
		"MaxAttempts": 20,
		"InitialBackoff": "0.500000s",
		"MaxBackoff": "5.000000s",
		"BackoffMultiplier": 0.500000,
		"RetryableStatusCodes": [ "CANCELLED","UNKNOWN","INVALID_ARGUMENT","DEADLINE_EXCEEDED","NOT_FOUND","ALREADY_EXISTS","PERMISSION_DENIED","RESOURCE_EXHAUSTED","FAILED_PRECONDITION","ABORTED","OUT_OF_RANGE","UNIMPLEMENTED","INTERNAL","UNAVAILABLE","DATA_LOSS","UNAUTHENTICATED" ]
	}
}]}`,
		},
		{
			name: "defaults",
			args: args{
				extSvc: &egv1a1.ExtensionService{
					BackendEndpoint: egv1a1.BackendEndpoint{
						FQDN: &egv1a1.FQDNEndpoint{
							Hostname: "foo.bar",
							Port:     44344,
						},
					},
				},
			},
			want: `{
"methodConfig": [{
	"name": [{"service": "envoygateway.extension.EnvoyGatewayExtension"}],
	"waitForReady": true,
	"retryPolicy": {
		"MaxAttempts": 4,
		"InitialBackoff": "0.100000s",
		"MaxBackoff": "1.000000s",
		"BackoffMultiplier": 2.000000,
		"RetryableStatusCodes": [ "UNAVAILABLE" ]
	}
}]}`,
		},
		{
			name: "invalid-code",
			args: args{
				extSvc: &egv1a1.ExtensionService{
					BackendEndpoint: egv1a1.BackendEndpoint{
						FQDN: &egv1a1.FQDNEndpoint{
							Hostname: "foo.bar",
							Port:     44344,
						},
					},
					Retry: &egv1a1.ExtensionServiceRetry{
						RetryableStatusCodes: []egv1a1.RetryableGRPCStatusCode{
							"CANCELLED",
							"NOTACODE",
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildServiceConfig("envoygateway.extension.EnvoyGatewayExtension", tt.args.extSvc)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildServiceConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildServiceConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetExtensionServerAddress(t *testing.T) {
	tests := []struct {
		Name     string
		Service  *egv1a1.ExtensionService
		Expected string
	}{
		{
			Name: "has an FQDN",
			Service: &egv1a1.ExtensionService{
				BackendEndpoint: egv1a1.BackendEndpoint{
					FQDN: &egv1a1.FQDNEndpoint{
						Hostname: "extserver.svc.cluster.local",
						Port:     5050,
					},
				},
			},
			Expected: "extserver.svc.cluster.local:5050",
		},
		{
			Name: "has an IP",
			Service: &egv1a1.ExtensionService{
				BackendEndpoint: egv1a1.BackendEndpoint{
					IP: &egv1a1.IPEndpoint{
						Address: "10.10.10.10",
						Port:    5050,
					},
				},
			},
			Expected: "10.10.10.10:5050",
		},
		{
			Name: "has a Unix path",
			Service: &egv1a1.ExtensionService{
				BackendEndpoint: egv1a1.BackendEndpoint{
					Unix: &egv1a1.UnixSocket{
						Path: "/some/path",
					},
				},
			},
			Expected: "unix:///some/path",
		},
		{
			Name: "has a Unix path",
			Service: &egv1a1.ExtensionService{
				Host: "foo.bar",
				Port: 5050,
			},
			Expected: "foo.bar:5050",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			out := GetExtensionServerAddress(tc.Service)
			require.Equal(t, tc.Expected, out)
		})
	}
}
