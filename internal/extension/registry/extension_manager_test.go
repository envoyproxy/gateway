// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math"
	"net"
	"os"
	"testing"

	v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/proto/extension"
)

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
			out := getExtensionServerAddress(tc.Service)
			require.Equal(t, tc.Expected, out)
		})
	}
}

func Test_setupGRPCOpts(t *testing.T) {
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
					MaxMessageSize: ptr.To(resource.MustParse(fmt.Sprintf("%dM", math.MaxInt))),
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
					MaxMessageSize: ptr.To(resource.MustParse(fmt.Sprintf("%dM", 0))),
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
					MaxMessageSize: ptr.To(resource.MustParse(fmt.Sprintf("%dM", 10))),
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
			_, err := setupGRPCOpts(context.TODO(), fc, tt.args.ext, "envoy-gateway-system")
			if (err != nil) != tt.wantErr {
				t.Errorf("setupGRPCOpts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

type testServer struct {
	extension.UnimplementedEnvoyGatewayExtensionServer
}

func (s *testServer) PostRouteModify(ctx context.Context, req *extension.PostRouteModifyRequest) (*extension.PostRouteModifyResponse, error) {
	return &extension.PostRouteModifyResponse{
		Route: req.Route,
	}, nil
}

func Test_TLS(t *testing.T) {
	testDir := "testdata"
	caFile := testDir + "/ca.pem"
	certFile := testDir + "/cert.pem"
	keyFile := testDir + "/key.pem"

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	require.NoError(t, err)

	caCert, err := os.ReadFile(caFile)
	require.NoError(t, err)
	caPool := x509.NewCertPool()
	ok := caPool.AppendCertsFromPEM(caCert)
	require.True(t, ok)

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer lis.Close()

	port := lis.Addr().(*net.TCPAddr).Port
	server := grpc.NewServer(grpc.Creds(credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.NoClientCert,
		MinVersion:   tls.VersionTLS12,
	})))
	extension.RegisterEnvoyGatewayExtensionServer(server, &testServer{})
	go func() {
		_ = server.Serve(lis)
		defer server.GracefulStop()
	}()

	extManager := &egv1a1.ExtensionManager{
		Service: &egv1a1.ExtensionService{
			BackendEndpoint: egv1a1.BackendEndpoint{
				IP: &egv1a1.IPEndpoint{
					Address: "localhost",
					Port:    int32(port),
				},
			},
			TLS: &egv1a1.ExtensionTLS{
				CertificateRef: gwapiv1.SecretObjectReference{
					Name:      "cert",
					Namespace: ptr.To(gwapiv1.Namespace("default")),
				},
			},
		},
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cert",
			Namespace: "default",
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey: caCert,
		},
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(secret).Build()

	opts, err := setupGRPCOpts(context.Background(), fakeClient, extManager, "test-ns")
	require.NoError(t, err)
	require.NotEmpty(t, opts)

	conn, err := grpc.DialContext(context.Background(), fmt.Sprintf("localhost:%d", port),
		opts...,
	)
	require.NoError(t, err)
	defer conn.Close()

	client := extension.NewEnvoyGatewayExtensionClient(conn)
	require.NotNil(t, client)

	response, err := client.PostRouteModify(context.Background(), &extension.PostRouteModifyRequest{
		Route: &v3.Route{
			Name: "test-route",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "test-route", response.Route.Name)
}
